package core

import (
	"goflashdb/pkg/resp"
)

// execSETBIT sets a bit value at the specified offset
func execSETBIT(db *DB, args [][]byte) resp.Reply {
	if len(args) != 3 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'setbit' command")
	}

	key := string(args[0])
	offset, err := parseInt64(string(args[1]))
	if err != nil {
		return resp.NewErrorReply("ERR bit offset is not an integer or out of range")
	}

	if offset < 0 {
		return resp.NewErrorReply("ERR bit offset is not an integer or out of range")
	}

	value, err := parseInt64(string(args[2]))
	if err != nil {
		return resp.NewErrorReply("ERR bit value is not an integer or out of range")
	}

	if value != 0 && value != 1 {
		return resp.NewErrorReply("ERR bit value must be 0 or 1")
	}

	// Get existing string data or create new
	var data []byte
	stringData, exists := db.GetStringData(key)
	if exists {
		data = make([]byte, len(stringData.value))
		copy(data, stringData.value)
	} else {
		data = []byte{}
	}

	// Calculate the byte index and bit position
	byteIndex := offset / 8
	bitPos := 7 - (offset % 8) // MSB first

	// Expand the data if necessary
	if int(byteIndex) >= len(data) {
		newData := make([]byte, byteIndex+1)
		copy(newData, data)
		data = newData
	}

	// Get the old bit value
	oldBit := (data[byteIndex] >> bitPos) & 1

	// Set the new bit value
	if value == 1 {
		data[byteIndex] |= (1 << bitPos)
	} else {
		data[byteIndex] &= ^(1 << bitPos)
	}

	// Store back
	db.SetString(key, data)

	return &resp.IntegerReply{Num: int64(oldBit)}
}

// execGETBIT gets a bit value at the specified offset
func execGETBIT(db *DB, args [][]byte) resp.Reply {
	if len(args) != 2 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'getbit' command")
	}

	key := string(args[0])
	offset, err := parseInt64(string(args[1]))
	if err != nil {
		return resp.NewErrorReply("ERR bit offset is not an integer or out of range")
	}

	if offset < 0 {
		return resp.NewErrorReply("ERR bit offset is not an integer or out of range")
	}

	stringData, exists := db.GetStringData(key)
	if !exists {
		return &resp.IntegerReply{Num: 0}
	}

	data := stringData.value
	byteIndex := offset / 8

	// If offset is beyond the data, return 0
	if int(byteIndex) >= len(data) {
		return &resp.IntegerReply{Num: 0}
	}

	bitPos := 7 - (offset % 8) // MSB first
	bit := (data[byteIndex] >> bitPos) & 1

	return &resp.IntegerReply{Num: int64(bit)}
}

// execBITCOUNT counts the number of set bits in a string
func execBITCOUNT(db *DB, args [][]byte) resp.Reply {
	if len(args) < 1 || len(args) > 3 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'bitcount' command")
	}

	key := string(args[0])

	stringData, exists := db.GetStringData(key)
	if !exists {
		return &resp.IntegerReply{Num: 0}
	}

	data := stringData.value
	if len(data) == 0 {
		return &resp.IntegerReply{Num: 0}
	}

	start := 0
	end := len(data) - 1

	if len(args) >= 2 {
		var err error
		start, err = parseInt(string(args[1]))
		if err != nil {
			return resp.NewErrorReply("ERR value is not an integer")
		}
	}

	if len(args) == 3 {
		var err error
		end, err = parseInt(string(args[2]))
		if err != nil {
			return resp.NewErrorReply("ERR value is not an integer")
		}
	}

	// Handle negative indices
	length := len(data)
	if start < 0 {
		start = length + start
	}
	if end < 0 {
		end = length + end
	}

	// Clamp to valid range
	if start < 0 {
		start = 0
	}
	if end >= length {
		end = length - 1
	}

	if start > end || start >= length {
		return &resp.IntegerReply{Num: 0}
	}

	// Count bits in the range
	var count int64
	for i := start; i <= end; i++ {
		count += int64(bitCountTable[data[i]])
	}

	return &resp.IntegerReply{Num: count}
}

// bitCountTable is a lookup table for counting bits in a byte
var bitCountTable = [256]int{
	0, 1, 1, 2, 1, 2, 2, 3, 1, 2, 2, 3, 2, 3, 3, 4,
	1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5,
	1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5,
	2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
	1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5,
	2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
	2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
	3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7,
	1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5,
	2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
	2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
	3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7,
	2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
	3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7,
	3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7,
	4, 5, 5, 6, 5, 6, 6, 7, 5, 6, 6, 7, 6, 7, 7, 8,
}

// execBITPOS finds the first bit set to 0 or 1
func execBITPOS(db *DB, args [][]byte) resp.Reply {
	if len(args) < 2 || len(args) > 4 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'bitpos' command")
	}

	key := string(args[0])

	bit, err := parseInt64(string(args[1]))
	if err != nil {
		return resp.NewErrorReply("ERR bit must be 0 or 1")
	}

	if bit != 0 && bit != 1 {
		return resp.NewErrorReply("ERR bit must be 0 or 1")
	}

	stringData, exists := db.GetStringData(key)
	if !exists || len(stringData.value) == 0 {
		// Redis returns -1 when looking for 0, but returns length when looking for 1 in empty string
		if bit == 0 {
			return &resp.IntegerReply{Num: -1}
		}
		return &resp.IntegerReply{Num: 0}
	}

	data := stringData.value

	start := 0
	end := len(data)*8 - 1

	if len(args) >= 3 {
		start, err = parseInt(string(args[2]))
		if err != nil {
			return resp.NewErrorReply("ERR value is not an integer")
		}
	}

	if len(args) == 4 {
		end, err = parseInt(string(args[3]))
		if err != nil {
			return resp.NewErrorReply("ERR value is not an integer")
		}
	}

	// Handle negative indices
	length := len(data) * 8
	if start < 0 {
		start = length + start
	}
	if end < 0 {
		end = length + end
	}

	// Clamp to valid range
	if start < 0 {
		start = 0
	}
	if end >= length {
		end = length - 1
	}

	if start > end || start >= length {
		if bit == 0 {
			return &resp.IntegerReply{Num: -1}
		}
		return &resp.IntegerReply{Num: 0}
	}

	// Find the first bit
	for i := start; i <= end; i++ {
		byteIndex := i / 8
		bitPos := 7 - (i % 8)
		currentBit := (data[byteIndex] >> bitPos) & 1

		if int(currentBit) == int(bit) {
			return &resp.IntegerReply{Num: int64(i)}
		}
	}

	// Not found
	if bit == 0 {
		return &resp.IntegerReply{Num: -1}
	}
	return &resp.IntegerReply{Num: 0}
}

// execBITOP performs bitwise operations on multiple strings
func execBITOP(db *DB, args [][]byte) resp.Reply {
	if len(args) < 3 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'bitop' command")
	}

	operation := string(args[0])
	destKey := string(args[1])
	sourceKeys := make([]string, len(args)-2)
	for i := 2; i < len(args); i++ {
		sourceKeys[i-2] = string(args[i])
	}

	// Validate operation
	switch operation {
	case "AND", "OR", "XOR", "NOT":
	default:
		return resp.NewErrorReply("ERR syntax error")
	}

	// Special case for NOT
	if operation == "NOT" {
		if len(sourceKeys) != 1 {
			return resp.NewErrorReply("ERR wrong number of arguments for 'bitop' command")
		}

		stringData, exists := db.GetStringData(sourceKeys[0])
		if !exists {
			db.SetString(destKey, []byte{})
			return &resp.IntegerReply{Num: 0}
		}

		result := make([]byte, len(stringData.value))
		for i, b := range stringData.value {
			result[i] = ^b
		}

		db.SetString(destKey, result)
		return &resp.IntegerReply{Num: int64(len(result) * 8)}
	}

	// Get the maximum length among all source strings
	var maxLen int
	var sourceData [][]byte

	for _, key := range sourceKeys {
		stringData, exists := db.GetStringData(key)
		if exists {
			sourceData = append(sourceData, stringData.value)
			if len(stringData.value) > maxLen {
				maxLen = len(stringData.value)
			}
		} else {
			sourceData = append(sourceData, []byte{})
		}
	}

	if maxLen == 0 {
		db.SetString(destKey, []byte{})
		return &resp.IntegerReply{Num: 0}
	}

	// Perform the operation
	result := make([]byte, maxLen)

	for i := 0; i < maxLen; i++ {
		var b byte
		switch operation {
		case "AND":
			b = 0xFF
			for j := 0; j < len(sourceData); j++ {
				if i < len(sourceData[j]) {
					b &= sourceData[j][i]
				}
			}
		case "OR":
			b = 0x00
			for j := 0; j < len(sourceData); j++ {
				if i < len(sourceData[j]) {
					b |= sourceData[j][i]
				}
			}
		case "XOR":
			b = 0x00
			for j := 0; j < len(sourceData); j++ {
				if i < len(sourceData[j]) {
					b ^= sourceData[j][i]
				}
			}
		}
		result[i] = b
	}

	// Trim trailing zeros for AND operation (Redis behavior)
	if operation == "AND" {
		for maxLen > 0 && result[maxLen-1] == 0x00 {
			maxLen--
		}
		result = result[:maxLen]
	}

	db.SetString(destKey, result)
	return &resp.IntegerReply{Num: int64(len(result) * 8)}
}

func init() {
	RegisterCommand("setbit", execSETBIT, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return []string{string(args[0])}, nil
		}
		return nil, nil
	}, 4)

	RegisterCommand("getbit", execGETBIT, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, 3)

	RegisterCommand("bitcount", execBITCOUNT, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, -3)

	RegisterCommand("bitpos", execBITPOS, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, -3)

	RegisterCommand("bitop", execBITOP, func(args [][]byte) ([]string, []string) {
		if len(args) >= 2 {
			writeKeys := []string{string(args[1])}
			readKeys := []string{}
			for i := 2; i < len(args); i++ {
				readKeys = append(readKeys, string(args[i]))
			}
			return writeKeys, readKeys
		}
		return nil, nil
	}, -4)
}
