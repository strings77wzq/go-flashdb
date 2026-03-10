package core

import (
	"goflashdb/pkg/resp"
	"hash/fnv"
	"math"
)

// HLLData stores HyperLogLog data
// Uses 14 bits for bucket index (16384 buckets) and 6 bits per bucket (0-63)
type HLLData struct {
	buckets [16384]uint8
}

const (
	hllBucketBits = 14
	hllNumBuckets = 1 << hllBucketBits // 16384
	hllMaxCount   = 63
	hllPrecision  = 14
)

// NewHLLData creates a new HyperLogLog data structure
func NewHLLData() *HLLData {
	return &HLLData{}
}

// hllHash computes the hash of the input element for HLL
func hllHash(element []byte) uint64 {
	h := fnv.New64a()
	h.Write(element)
	return h.Sum64()
}

// getBucketAndCount extracts the bucket index and leading zero count from hash
func getBucketAndCount(h uint64) (bucket uint64, count uint8) {
	// Get the lower hllBucketBits for bucket index
	bucket = h & ((1 << hllBucketBits) - 1)

	// Get the remaining bits for counting leading zeros
	// We have 64 - hllBucketBits = 50 bits remaining
	// We only need 6 bits for the count (0-63), so we shift by 64 - hllBucketBits - 6 = 44
	remainder := h >> hllBucketBits

	// Count leading zeros in the remaining 50 bits, but cap at hllMaxCount
	count = 0
	if remainder != 0 {
		// Count leading zeros
		leadingZeros := 0
		for i := 49; i >= 0; i-- {
			if (remainder>>uint(i))&1 == 0 {
				leadingZeros++
			} else {
				break
			}
		}
		// Add 1 because we need to count the position of the first 1 bit
		count = uint8(leadingZeros + 1)
		if count > hllMaxCount {
			count = hllMaxCount
		}
	} else {
		count = hllMaxCount
	}

	return bucket, count
}

// add adds an element to the HyperLogLog and returns true if the register was updated
func (h *HLLData) add(element []byte) bool {
	hVal := hllHash(element)
	bucket, count := getBucketAndCount(hVal)

	oldCount := h.buckets[bucket]
	if count > oldCount {
		h.buckets[bucket] = count
		return true
	}
	return false
}

// count returns the cardinality estimate
func (h *HLLData) count() uint64 {
	// Count number of registers with zero
	zeroCount := 0
	for _, v := range h.buckets {
		if v == 0 {
			zeroCount++
		}
	}

	// HyperLogLog estimate
	// E = m * m / sum(1 / 2^M[i])
	m := float64(hllNumBuckets)

	var sumInv float64
	for _, v := range h.buckets {
		if v > 0 {
			sumInv += 1.0 / float64(uint64(1)<<v)
		}
	}

	if sumInv == 0 {
		return 0
	}

	estimate := m * m / sumInv

	// Apply bias correction for small cardinalities
	// Use linear counting when many registers are zero
	if float64(zeroCount) > m*0.05 { // More than 5% zeros
		// Linear counting
		if zeroCount == 0 {
			zeroCount = 1 // Avoid division by zero
		}
		estimate = m * math.Log(m/float64(zeroCount))
	}

	// Apply small range bias correction
	if estimate < float64(hllNumBuckets)*1.1 {
		// Count distinct registers
		var numZero uint64
		for _, v := range h.buckets {
			if v == 0 {
				numZero++
			}
		}
		if numZero > 0 {
			estimate = m * math.Log(m/float64(numZero))
		}
	}

	if estimate > 1<<36 {
		return 1 << 36
	}

	return uint64(estimate + 0.5)
}

// merge merges another HLL into this one
func (h *HLLData) merge(other *HLLData) {
	for i := 0; i < hllNumBuckets; i++ {
		if other.buckets[i] > h.buckets[i] {
			h.buckets[i] = other.buckets[i]
		}
	}
}

// serialize serializes the HLL to bytes for storage
func (h *HLLData) serialize() []byte {
	// Store as raw bytes (16384 bytes)
	data := make([]byte, hllNumBuckets)
	for i, v := range h.buckets {
		data[i] = byte(v)
	}
	return data
}

// deserialize deserializes HLL from bytes
func deserializeHLL(data []byte) *HLLData {
	h := NewHLLData()
	if len(data) >= hllNumBuckets {
		for i := 0; i < hllNumBuckets; i++ {
			h.buckets[i] = data[i]
		}
	}
	return h
}

// GetHLLData retrieves HLL data from the database
func (db *DB) GetHLLData(key string) (*HLLData, bool) {
	val, ok := db.data.Get(key)
	if !ok {
		return nil, false
	}
	hll, ok := val.(*HLLData)
	if !ok {
		// Try to deserialize from string data
		if sd, ok := val.(*StringData); ok {
			hll = deserializeHLL(sd.value)
			return hll, true
		}
		return nil, false
	}
	if db.IsExpired(key) {
		db.data.Delete(key)
		db.ttlDict.Delete(key)
		return nil, false
	}
	return hll, true
}

// SetHLL stores HLL data in the database
func (db *DB) SetHLL(key string, hllData *HLLData) {
	db.data.Set(key, hllData)
	db.RemoveExpire(key)
}

// execPFADD adds elements to a HyperLogLog
func execPFADD(db *DB, args [][]byte) resp.Reply {
	if len(args) < 2 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'pfadd' command")
	}

	key := string(args[1])
	elements := args[2:]

	if len(elements) == 0 {
		// Just return the cardinality
		hll, exists := db.GetHLLData(key)
		if !exists {
			return &resp.IntegerReply{Num: 0}
		}
		return &resp.IntegerReply{Num: int64(hll.count())}
	}

	// Get or create HLL
	hll, exists := db.GetHLLData(key)
	if !exists {
		hll = NewHLLData()
	}

	// Add all elements
	updated := false
	for _, elem := range elements {
		if hll.add(elem) {
			updated = true
		}
	}

	// Store the HLL
	db.SetHLL(key, hll)

	if updated {
		return &resp.IntegerReply{Num: 1}
	}
	return &resp.IntegerReply{Num: 0}
}

// execPFCOUNT returns the cardinality of the HyperLogLog
func execPFCOUNT(db *DB, args [][]byte) resp.Reply {
	if len(args) < 2 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'pfcount' command")
	}

	// Multiple keys can be provided - merge and count
	keys := make([]string, len(args)-1)
	for i := 1; i < len(args); i++ {
		keys[i-1] = string(args[i])
	}

	if len(keys) == 1 {
		// Single key
		hll, exists := db.GetHLLData(keys[0])
		if !exists {
			return &resp.IntegerReply{Num: 0}
		}
		return &resp.IntegerReply{Num: int64(hll.count())}
	}

	// Multiple keys - merge them
	result := NewHLLData()
	for _, key := range keys {
		hll, exists := db.GetHLLData(key)
		if exists {
			result.merge(hll)
		}
	}

	return &resp.IntegerReply{Num: int64(result.count())}
}

// execPFMERGE merges multiple HyperLogLogs into one
func execPFMERGE(db *DB, args [][]byte) resp.Reply {
	if len(args) < 3 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'pfmerge' command")
	}

	destKey := string(args[1])
	sourceKeys := make([]string, len(args)-2)
	for i := 2; i < len(args); i++ {
		sourceKeys[i-2] = string(args[i])
	}

	// Create destination HLL
	result := NewHLLData()

	// Merge all source HLLs
	for _, key := range sourceKeys {
		hll, exists := db.GetHLLData(key)
		if exists {
			result.merge(hll)
		}
	}

	// Store the result
	db.SetHLL(destKey, result)

	return resp.OkReply
}

func init() {
	RegisterCommand("pfadd", execPFADD, func(args [][]byte) ([]string, []string) {
		if len(args) > 1 {
			return []string{string(args[1])}, nil
		}
		return nil, nil
	}, -2)

	RegisterCommand("pfcount", execPFCOUNT, func(args [][]byte) ([]string, []string) {
		if len(args) > 1 {
			readKeys := make([]string, len(args)-1)
			for i := 1; i < len(args); i++ {
				readKeys[i-1] = string(args[i])
			}
			return nil, readKeys
		}
		return nil, nil
	}, -2)

	RegisterCommand("pfmerge", execPFMERGE, func(args [][]byte) ([]string, []string) {
		if len(args) > 2 {
			writeKeys := []string{string(args[1])}
			readKeys := make([]string, len(args)-2)
			for i := 2; i < len(args); i++ {
				readKeys[i-2] = string(args[i])
			}
			return writeKeys, readKeys
		}
		return nil, nil
	}, -3)
}
