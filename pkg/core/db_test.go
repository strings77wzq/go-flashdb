package core

import (
	"testing"

	"goflashdb/pkg/resp"
)

func newTestDB() *DB {
	return NewDB(0)
}

func getIntReply(r resp.Reply) int64 {
	if ir, ok := r.(*resp.IntegerReply); ok {
		return ir.Num
	}
	return -1
}

func getBulkReply(r resp.Reply) string {
	if br, ok := r.(*resp.BulkReply); ok {
		return string(br.Arg)
	}
	return ""
}

func getSimpleStringReply(r resp.Reply) string {
	if sr, ok := r.(*resp.SimpleStringReply); ok {
		return sr.Str
	}
	return ""
}

func getArrayReply(r resp.Reply) []resp.Reply {
	if ar, ok := r.(*resp.ArrayReply); ok {
		return ar.Replies
	}
	return nil
}

func TestPing(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("PING", [][]byte{})
	if getSimpleStringReply(reply) != "PONG" {
		t.Errorf("expected PONG, got %s", getSimpleStringReply(reply))
	}

	reply = db.Exec("PING", [][]byte{[]byte("hello")})
	if getBulkReply(reply) != "hello" {
		t.Errorf("expected hello, got %s", getBulkReply(reply))
	}
}

func TestEcho(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("ECHO", [][]byte{[]byte("hello world")})
	if getBulkReply(reply) != "hello world" {
		t.Errorf("expected hello world, got %s", getBulkReply(reply))
	}
}

func TestDBSize(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("DBSIZE", [][]byte{})
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0, got %d", getIntReply(reply))
	}

	db.Exec("SET", [][]byte{[]byte("k1"), []byte("v1")})
	db.Exec("SET", [][]byte{[]byte("k2"), []byte("v2")})

	reply = db.Exec("DBSIZE", [][]byte{})
	if getIntReply(reply) != 2 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}
}

func TestFlushDB(t *testing.T) {
	db := newTestDB()

	db.Exec("SET", [][]byte{[]byte("k1"), []byte("v1")})
	db.Exec("SET", [][]byte{[]byte("k2"), []byte("v2")})

	reply := db.Exec("FLUSHDB", [][]byte{})
	if reply.ToBytes() == nil {
		t.Errorf("expected OK reply")
	}

	reply = db.Exec("DBSIZE", [][]byte{})
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0 after flushdb, got %d", getIntReply(reply))
	}
}

func TestSetGet(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("SET", [][]byte{[]byte("key"), []byte("value")})
	if reply.ToBytes() == nil {
		t.Errorf("expected OK reply")
	}

	reply = db.Exec("GET", [][]byte{[]byte("key")})
	if getBulkReply(reply) != "value" {
		t.Errorf("expected value, got %s", getBulkReply(reply))
	}

	reply = db.Exec("GET", [][]byte{[]byte("nonexistent")})
	if len(getBulkReply(reply)) != 0 {
		t.Errorf("expected empty, got %s", getBulkReply(reply))
	}
}

func TestIncr(t *testing.T) {
	db := newTestDB()

	db.Exec("SET", [][]byte{[]byte("counter"), []byte("0")})

	reply := db.Exec("INCR", [][]byte{[]byte("counter")})
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}

	reply = db.Exec("INCR", [][]byte{[]byte("counter")})
	if getIntReply(reply) != 2 {
		t.Errorf("expected 2, got %d", getIntReply(reply))
	}

	reply = db.Exec("INCRBY", [][]byte{[]byte("counter"), []byte("5")})
	if getIntReply(reply) != 7 {
		t.Errorf("expected 7, got %d", getIntReply(reply))
	}

	reply = db.Exec("DECR", [][]byte{[]byte("counter")})
	if getIntReply(reply) != 6 {
		t.Errorf("expected 6, got %d", getIntReply(reply))
	}
}

func TestMSetMGet(t *testing.T) {
	db := newTestDB()

	db.Exec("MSET", [][]byte{[]byte("k1"), []byte("v1"), []byte("k2"), []byte("v2")})

	reply := db.Exec("MGET", [][]byte{[]byte("k1"), []byte("k2"), []byte("k3")})
	replies := getArrayReply(reply)
	if len(replies) != 3 {
		t.Errorf("expected 3 replies, got %d", len(replies))
	}
	if getBulkReply(replies[0]) != "v1" {
		t.Errorf("expected v1, got %s", getBulkReply(replies[0]))
	}
	if getBulkReply(replies[1]) != "v2" {
		t.Errorf("expected v2, got %s", getBulkReply(replies[1]))
	}
}

func TestExists(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("EXISTS", [][]byte{[]byte("key")})
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0, got %d", getIntReply(reply))
	}

	db.Exec("SET", [][]byte{[]byte("key"), []byte("value")})

	reply = db.Exec("EXISTS", [][]byte{[]byte("key")})
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}
}

func TestDel(t *testing.T) {
	db := newTestDB()

	db.Exec("SET", [][]byte{[]byte("k1"), []byte("v1")})
	db.Exec("SET", [][]byte{[]byte("k2"), []byte("v2")})

	reply := db.Exec("DEL", [][]byte{[]byte("k1")})
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}

	reply = db.Exec("EXISTS", [][]byte{[]byte("k1")})
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0, got %d", getIntReply(reply))
	}
}

func TestExpire(t *testing.T) {
	db := newTestDB()

	db.Exec("SET", [][]byte{[]byte("key"), []byte("value")})

	reply := db.Exec("EXPIRE", [][]byte{[]byte("key"), []byte("10")})
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}

	reply = db.Exec("TTL", [][]byte{[]byte("key")})
	if getIntReply(reply) <= 0 {
		t.Errorf("expected positive TTL, got %d", getIntReply(reply))
	}

	reply = db.Exec("PERSIST", [][]byte{[]byte("key")})
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}

	reply = db.Exec("TTL", [][]byte{[]byte("key")})
	if getIntReply(reply) != -1 {
		t.Errorf("expected -1, got %d", getIntReply(reply))
	}
}

func TestHash(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("HSET", [][]byte{[]byte("hash"), []byte("field1"), []byte("value1")})
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}

	reply = db.Exec("HGET", [][]byte{[]byte("hash"), []byte("field1")})
	if getBulkReply(reply) != "value1" {
		t.Errorf("expected value1, got %s", getBulkReply(reply))
	}

	reply = db.Exec("HLEN", [][]byte{[]byte("hash")})
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}

	reply = db.Exec("HEXISTS", [][]byte{[]byte("hash"), []byte("field1")})
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}
}

func TestList(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("LPUSH", [][]byte{[]byte("list"), []byte("a"), []byte("b"), []byte("c")})
	if getIntReply(reply) != 3 {
		t.Errorf("expected 3, got %d", getIntReply(reply))
	}

	reply = db.Exec("LLEN", [][]byte{[]byte("list")})
	if getIntReply(reply) != 3 {
		t.Errorf("expected 3, got %d", getIntReply(reply))
	}

	reply = db.Exec("LPOP", [][]byte{[]byte("list")})
	if getBulkReply(reply) != "c" {
		t.Errorf("expected c, got %s", getBulkReply(reply))
	}

	reply = db.Exec("RPOP", [][]byte{[]byte("list")})
	if getBulkReply(reply) != "a" {
		t.Errorf("expected a, got %s", getBulkReply(reply))
	}
}

func TestSet(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("SADD", [][]byte{[]byte("set"), []byte("a"), []byte("b"), []byte("c")})
	if getIntReply(reply) != 3 {
		t.Errorf("expected 3, got %d", getIntReply(reply))
	}

	reply = db.Exec("SCARD", [][]byte{[]byte("set")})
	if getIntReply(reply) != 3 {
		t.Errorf("expected 3, got %d", getIntReply(reply))
	}

	reply = db.Exec("SISMEMBER", [][]byte{[]byte("set"), []byte("a")})
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}

	reply = db.Exec("SREM", [][]byte{[]byte("set"), []byte("a")})
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}
}

func TestZSet(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("ZADD", [][]byte{[]byte("zset"), []byte("1"), []byte("one"), []byte("2"), []byte("two")})
	if getIntReply(reply) != 2 {
		t.Errorf("expected 2, got %d", getIntReply(reply))
	}

	reply = db.Exec("ZCARD", [][]byte{[]byte("zset")})
	if getIntReply(reply) != 2 {
		t.Errorf("expected 2, got %d", getIntReply(reply))
	}

	reply = db.Exec("ZSCORE", [][]byte{[]byte("zset"), []byte("one")})
	if getBulkReply(reply) != "1" {
		t.Errorf("expected 1, got %s", getBulkReply(reply))
	}

	reply = db.Exec("ZRANK", [][]byte{[]byte("zset"), []byte("two")})
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}
}

func TestBitmap(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("SETBIT", [][]byte{[]byte("bits"), []byte("7"), []byte("1")})
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0, got %d", getIntReply(reply))
	}

	reply = db.Exec("GETBIT", [][]byte{[]byte("bits"), []byte("7")})
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}
}

func TestHyperLogLog(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("PFADD", [][]byte{[]byte("hll"), []byte("a")})
	if getIntReply(reply) < 0 {
		t.Errorf("expected >= 0, got %d", getIntReply(reply))
	}
}

func TestMulti(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("MULTI", [][]byte{})
	if getSimpleStringReply(reply) != "OK" {
		t.Errorf("expected OK, got %s", getSimpleStringReply(reply))
	}

	db.Exec("SET", [][]byte{[]byte("k1"), []byte("v1")})

	reply = db.Exec("EXEC", [][]byte{})
	if _, ok := reply.(*resp.ArrayReply); !ok {
		t.Errorf("expected ArrayReply, got %v", reply)
	}
}

func TestNewDBWithPersist(t *testing.T) {
	db := NewDBWithPersist(0, nil)
	if db == nil {
		t.Error("NewDBWithPersist should not return nil")
	}
}

func TestDBWithExpire(t *testing.T) {
	db := newTestDB()

	db.Exec("SET", [][]byte{[]byte("key"), []byte("value")})
	db.Exec("EXPIRE", [][]byte{[]byte("key"), []byte("100")})

	reply := db.Exec("TTL", [][]byte{[]byte("key")})
	if getIntReply(reply) <= 0 {
		t.Errorf("expected positive TTL, got %d", getIntReply(reply))
	}

	reply = db.Exec("PTTL", [][]byte{[]byte("key")})
	if getIntReply(reply) <= 0 {
		t.Errorf("expected positive PTTL, got %d", getIntReply(reply))
	}

	reply = db.Exec("EXPIREAT", [][]byte{[]byte("key"), []byte("9999999999")})
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}

	reply = db.Exec("PEXPIRE", [][]byte{[]byte("key"), []byte("100000")})
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}

	reply = db.Exec("PEXPIREAT", [][]byte{[]byte("key"), []byte("9999999999000")})
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}
}

func TestDBGetAllData(t *testing.T) {
	db := newTestDB()

	db.Exec("SET", [][]byte{[]byte("k1"), []byte("v1")})
	db.Exec("SET", [][]byte{[]byte("k2"), []byte("v2")})

	data := db.GetAllData()
	if len(data) != 2 {
		t.Errorf("expected 2 keys, got %d", len(data))
	}
}

func TestDBGetAllExpireTimes(t *testing.T) {
	db := newTestDB()

	db.Exec("SET", [][]byte{[]byte("k1"), []byte("v1")})
	db.Exec("EXPIRE", [][]byte{[]byte("k1"), []byte("100")})

	expireTimes := db.GetAllExpireTimes()
	if len(expireTimes) != 1 {
		t.Errorf("expected 1 expire time, got %d", len(expireTimes))
	}
}

func TestDBLoadFromPersist(t *testing.T) {
	db := newTestDB()

	db.LoadFromPersist()
}

func TestDBAppend(t *testing.T) {
	db := newTestDB()

	db.Exec("SET", [][]byte{[]byte("str"), []byte("Hello")})

	reply := db.Exec("APPEND", [][]byte{[]byte("str"), []byte(" World")})
	if getIntReply(reply) != 11 {
		t.Errorf("expected 11, got %d", getIntReply(reply))
	}

	reply = db.Exec("GET", [][]byte{[]byte("str")})
	if getBulkReply(reply) != "Hello World" {
		t.Errorf("expected 'Hello World', got %s", getBulkReply(reply))
	}
}

func TestDBHGetAll(t *testing.T) {
	db := newTestDB()

	db.Exec("HSET", [][]byte{[]byte("hash"), []byte("f1"), []byte("v1"), []byte("f2"), []byte("v2")})

	reply := db.Exec("HGETALL", [][]byte{[]byte("hash")})
	replies := getArrayReply(reply)
	if len(replies) != 4 {
		t.Errorf("expected 4, got %d", len(replies))
	}
}

func TestDBHMGet(t *testing.T) {
	db := newTestDB()

	db.Exec("HSET", [][]byte{[]byte("hash"), []byte("f1"), []byte("v1"), []byte("f2"), []byte("v2")})

	reply := db.Exec("HMGET", [][]byte{[]byte("hash"), []byte("f1"), []byte("f3")})
	replies := getArrayReply(reply)
	if len(replies) != 2 {
		t.Errorf("expected 2, got %d", len(replies))
	}
	if getBulkReply(replies[0]) != "v1" {
		t.Errorf("expected v1, got %s", getBulkReply(replies[0]))
	}
}

func TestDBHKeys(t *testing.T) {
	db := newTestDB()

	db.Exec("HSET", [][]byte{[]byte("hash"), []byte("f1"), []byte("v1"), []byte("f2"), []byte("v2")})

	reply := db.Exec("HKEYS", [][]byte{[]byte("hash")})
	replies := getArrayReply(reply)
	if len(replies) != 2 {
		t.Errorf("expected 2, got %d", len(replies))
	}
}

func TestDBHVals(t *testing.T) {
	db := newTestDB()

	db.Exec("HSET", [][]byte{[]byte("hash"), []byte("f1"), []byte("v1"), []byte("f2"), []byte("v2")})

	reply := db.Exec("HVALS", [][]byte{[]byte("hash")})
	replies := getArrayReply(reply)
	if len(replies) != 2 {
		t.Errorf("expected 2, got %d", len(replies))
	}
}

func TestDBLRange(t *testing.T) {
	db := newTestDB()

	db.Exec("RPUSH", [][]byte{[]byte("list"), []byte("a"), []byte("b"), []byte("c")})

	reply := db.Exec("LRANGE", [][]byte{[]byte("list"), []byte("0"), []byte("-1")})
	replies := getArrayReply(reply)
	if len(replies) != 3 {
		t.Errorf("expected 3, got %d", len(replies))
	}
}

func TestDBLSet(t *testing.T) {
	db := newTestDB()

	db.Exec("RPUSH", [][]byte{[]byte("list"), []byte("a"), []byte("b")})

	reply := db.Exec("LSET", [][]byte{[]byte("list"), []byte("1"), []byte("c")})
	if getSimpleStringReply(reply) != "OK" {
		t.Errorf("expected OK, got %s", getSimpleStringReply(reply))
	}
}

func TestDBLTrim(t *testing.T) {
	db := newTestDB()

	db.Exec("RPUSH", [][]byte{[]byte("list"), []byte("a"), []byte("b"), []byte("c")})

	reply := db.Exec("LTRIM", [][]byte{[]byte("list"), []byte("0"), []byte("1")})
	if getSimpleStringReply(reply) != "OK" {
		t.Errorf("expected OK, got %s", getSimpleStringReply(reply))
	}

	reply = db.Exec("LLEN", [][]byte{[]byte("list")})
	if getIntReply(reply) != 2 {
		t.Errorf("expected 2, got %d", getIntReply(reply))
	}
}

func TestDBSAddMembers(t *testing.T) {
	db := newTestDB()

	db.Exec("SADD", [][]byte{[]byte("set"), []byte("a"), []byte("b")})

	reply := db.Exec("SMEMBERS", [][]byte{[]byte("set")})
	replies := getArrayReply(reply)
	if len(replies) != 2 {
		t.Errorf("expected 2, got %d", len(replies))
	}
}

func TestDBZAddMore(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("ZADD", [][]byte{[]byte("zset"), []byte("1"), []byte("one"), []byte("2"), []byte("two"), []byte("3"), []byte("three")})
	if getIntReply(reply) != 3 {
		t.Errorf("expected 3, got %d", getIntReply(reply))
	}
}

func TestDBZRange(t *testing.T) {
	db := newTestDB()

	db.Exec("ZADD", [][]byte{[]byte("zset"), []byte("1"), []byte("a"), []byte("2"), []byte("b"), []byte("3"), []byte("c")})

	reply := db.Exec("ZRANGE", [][]byte{[]byte("zset"), []byte("0"), []byte("-1")})
	replies := getArrayReply(reply)
	if len(replies) != 3 {
		t.Errorf("expected 3, got %d", len(replies))
	}
}

func TestDBZRevRange(t *testing.T) {
	db := newTestDB()

	db.Exec("ZADD", [][]byte{[]byte("zset"), []byte("1"), []byte("a"), []byte("2"), []byte("b"), []byte("3"), []byte("c")})

	reply := db.Exec("ZREVRANGE", [][]byte{[]byte("zset"), []byte("0"), []byte("-1")})
	replies := getArrayReply(reply)
	if len(replies) != 3 {
		t.Errorf("expected 3, got %d", len(replies))
	}
}

func TestDBZCard(t *testing.T) {
	db := newTestDB()

	db.Exec("ZADD", [][]byte{[]byte("zset"), []byte("1"), []byte("a"), []byte("2"), []byte("b")})

	reply := db.Exec("ZCARD", [][]byte{[]byte("zset")})
	if getIntReply(reply) != 2 {
		t.Errorf("expected 2, got %d", getIntReply(reply))
	}
}

func TestDBZCount(t *testing.T) {
	db := newTestDB()

	db.Exec("ZADD", [][]byte{[]byte("zset"), []byte("1"), []byte("a"), []byte("2"), []byte("b"), []byte("3"), []byte("c")})

	reply := db.Exec("ZCOUNT", [][]byte{[]byte("zset"), []byte("1"), []byte("2")})
	if getIntReply(reply) != 2 {
		t.Errorf("expected 2, got %d", getIntReply(reply))
	}
}

func TestDBZRem(t *testing.T) {
	db := newTestDB()

	db.Exec("ZADD", [][]byte{[]byte("zset"), []byte("1"), []byte("a"), []byte("2"), []byte("b")})

	reply := db.Exec("ZREM", [][]byte{[]byte("zset"), []byte("a")})
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}
}

func TestDBZIncrBy(t *testing.T) {
	db := newTestDB()

	db.Exec("ZADD", [][]byte{[]byte("zset"), []byte("1"), []byte("a")})

	reply := db.Exec("ZINCRBY", [][]byte{[]byte("zset"), []byte("2"), []byte("a")})
	if getBulkReply(reply) != "3" {
		t.Errorf("expected 3, got %s", getBulkReply(reply))
	}
}

func TestDBZRank(t *testing.T) {
	db := newTestDB()

	db.Exec("ZADD", [][]byte{[]byte("zset"), []byte("1"), []byte("a"), []byte("2"), []byte("b")})

	reply := db.Exec("ZRANK", [][]byte{[]byte("zset"), []byte("b")})
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}
}

func TestDBZRevRank(t *testing.T) {
	db := newTestDB()

	db.Exec("ZADD", [][]byte{[]byte("zset"), []byte("1"), []byte("a"), []byte("2"), []byte("b")})

	reply := db.Exec("ZREVRANK", [][]byte{[]byte("zset"), []byte("a")})
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}
}

func TestDBZScore(t *testing.T) {
	db := newTestDB()

	db.Exec("ZADD", [][]byte{[]byte("zset"), []byte("1.5"), []byte("a")})

	reply := db.Exec("ZSCORE", [][]byte{[]byte("zset"), []byte("a")})
	if getBulkReply(reply) != "1.5" {
		t.Errorf("expected 1.5, got %s", getBulkReply(reply))
	}
}

func TestDBPFMerge(t *testing.T) {
	db := newTestDB()

	db.Exec("PFADD", [][]byte{[]byte("hll1"), []byte("a"), []byte("b")})
	db.Exec("PFADD", [][]byte{[]byte("hll2"), []byte("b"), []byte("c")})

	reply := db.Exec("PFMERGE", [][]byte{[]byte("hll3"), []byte("hll1"), []byte("hll2")})
	if reply.ToBytes() == nil {
		t.Error("expected non-nil reply")
	}
}

func TestDBBitCountWithRange(t *testing.T) {
	db := newTestDB()

	db.Exec("SET", [][]byte{[]byte("bits"), []byte{0xFF, 0x00}})

	reply := db.Exec("BITCOUNT", [][]byte{[]byte("bits"), []byte("0"), []byte("0")})
	_ = reply
}

func TestDBBitCountNonExistent(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("BITCOUNT", [][]byte{[]byte("nonexistent")})
	_ = reply
}

func TestDBBitOpAND(t *testing.T) {
	db := newTestDB()

	db.Exec("SET", [][]byte{[]byte("key1"), []byte{0xFF}})
	db.Exec("SET", [][]byte{[]byte("key2"), []byte{0x0F}})

	reply := db.Exec("BITOP", [][]byte{[]byte("AND"), []byte("result"), []byte("key1"), []byte("key2")})
	_ = reply
}

func TestDBBitOpOR(t *testing.T) {
	db := newTestDB()

	db.Exec("SET", [][]byte{[]byte("key1"), []byte{0xF0}})
	db.Exec("SET", [][]byte{[]byte("key2"), []byte{0x0F}})

	reply := db.Exec("BITOP", [][]byte{[]byte("OR"), []byte("result"), []byte("key1"), []byte("key2")})
	_ = reply
}

func TestDBBitOpXOR(t *testing.T) {
	db := newTestDB()

	db.Exec("SET", [][]byte{[]byte("key1"), []byte{0xFF}})
	db.Exec("SET", [][]byte{[]byte("key2"), []byte{0xFF}})

	reply := db.Exec("BITOP", [][]byte{[]byte("XOR"), []byte("result"), []byte("key1"), []byte("key2")})
	_ = reply
}

func TestDBBitOpNOT(t *testing.T) {
	db := newTestDB()

	db.Exec("SET", [][]byte{[]byte("key1"), []byte{0x00}})

	reply := db.Exec("BITOP", [][]byte{[]byte("NOT"), []byte("result"), []byte("key1")})
	_ = reply
}

func TestDBBitPos(t *testing.T) {
	db := newTestDB()

	db.Exec("SETBIT", [][]byte{[]byte("bits"), []byte("7"), []byte("1")})

	reply := db.Exec("BITPOS", [][]byte{[]byte("bits"), []byte("1")})
	if getIntReply(reply) != 7 {
		t.Errorf("expected 7, got %d", getIntReply(reply))
	}
}

func TestDBLIndex(t *testing.T) {
	db := newTestDB()

	db.Exec("RPUSH", [][]byte{[]byte("list"), []byte("a"), []byte("b"), []byte("c")})

	reply := db.Exec("LINDEX", [][]byte{[]byte("list"), []byte("1")})
	if getBulkReply(reply) != "b" {
		t.Errorf("expected 'b', got '%s'", getBulkReply(reply))
	}
}

func TestDBLIndexOutOfRange(t *testing.T) {
	db := newTestDB()

	db.Exec("RPUSH", [][]byte{[]byte("list"), []byte("a")})

	reply := db.Exec("LINDEX", [][]byte{[]byte("list"), []byte("100")})
	if getBulkReply(reply) != "" {
		t.Errorf("expected empty, got '%s'", getBulkReply(reply))
	}
}

func TestDBSPop(t *testing.T) {
	db := newTestDB()

	db.Exec("SADD", [][]byte{[]byte("set"), []byte("a"), []byte("b"), []byte("c")})

	reply := db.Exec("SPOP", [][]byte{[]byte("set")})
	if reply.ToBytes() == nil {
		t.Error("expected non-nil reply")
	}
}

func TestDBSRandMember(t *testing.T) {
	db := newTestDB()

	db.Exec("SADD", [][]byte{[]byte("set"), []byte("a"), []byte("b"), []byte("c")})

	reply := db.Exec("SRANDMEMBER", [][]byte{[]byte("set")})
	if reply.ToBytes() == nil {
		t.Error("expected non-nil reply")
	}
}

func TestDBZRangeByScore(t *testing.T) {
	db := newTestDB()

	db.Exec("ZADD", [][]byte{[]byte("zset"), []byte("1"), []byte("a"), []byte("2"), []byte("b"), []byte("3"), []byte("c")})

	reply := db.Exec("ZRANGEBYSCORE", [][]byte{[]byte("zset"), []byte("1"), []byte("2")})
	_ = reply
}

func TestDBPFCountMultiple(t *testing.T) {
	db := newTestDB()

	db.Exec("PFADD", [][]byte{[]byte("hll1"), []byte("a"), []byte("b")})
	db.Exec("PFADD", [][]byte{[]byte("hll2"), []byte("b"), []byte("c")})

	reply := db.Exec("PFCOUNT", [][]byte{[]byte("hll1"), []byte("hll2")})
	_ = reply
}

func TestDBBitPosZero(t *testing.T) {
	db := newTestDB()

	db.Exec("SET", [][]byte{[]byte("bits"), []byte{0xFF}})

	reply := db.Exec("BITPOS", [][]byte{[]byte("bits"), []byte("0")})
	_ = reply
}

func TestDBBitCountEmpty(t *testing.T) {
	db := newTestDB()

	db.Exec("SET", [][]byte{[]byte("bits"), []byte{}})

	reply := db.Exec("BITCOUNT", [][]byte{[]byte("bits")})
	_ = reply
}

func TestDBStrlen(t *testing.T) {
	db := newTestDB()

	db.Exec("SET", [][]byte{[]byte("str"), []byte("hello")})

	reply := db.Exec("STRLEN", [][]byte{[]byte("str")})
	if getIntReply(reply) != 5 {
		t.Errorf("expected 5, got %d", getIntReply(reply))
	}
}

func TestDBStrlenNonExistent(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("STRLEN", [][]byte{[]byte("nonexistent")})
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0, got %d", getIntReply(reply))
	}
}

func TestDBDecr(t *testing.T) {
	db := newTestDB()

	db.Exec("SET", [][]byte{[]byte("num"), []byte("10")})

	reply := db.Exec("DECR", [][]byte{[]byte("num")})
	if getIntReply(reply) != 9 {
		t.Errorf("expected 9, got %d", getIntReply(reply))
	}
}

func TestDBDecrBy(t *testing.T) {
	db := newTestDB()

	db.Exec("SET", [][]byte{[]byte("num"), []byte("10")})

	reply := db.Exec("DECRBY", [][]byte{[]byte("num"), []byte("3")})
	if getIntReply(reply) != 7 {
		t.Errorf("expected 7, got %d", getIntReply(reply))
	}
}

func TestDBPersistNonExistent(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("PERSIST", [][]byte{[]byte("nonexistent")})
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0, got %d", getIntReply(reply))
	}
}

func TestDBExpireNonExistent(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("EXPIRE", [][]byte{[]byte("nonexistent"), []byte("10")})
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0, got %d", getIntReply(reply))
	}
}

func TestDBInfo(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("INFO", [][]byte{})
	if reply.ToBytes() == nil {
		t.Error("expected non-nil reply")
	}
}

func TestDBInfoSections(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("INFO", [][]byte{[]byte("server")})
	if reply.ToBytes() == nil {
		t.Error("expected non-nil reply")
	}
}

func TestDBConfigGet(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("CONFIG", [][]byte{[]byte("GET"), []byte("maxclients")})
	if reply.ToBytes() == nil {
		t.Error("expected non-nil reply")
	}
}

func TestDBConfigSet(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("CONFIG", [][]byte{[]byte("SET"), []byte("maxclients"), []byte("1000")})
	if reply.ToBytes() == nil {
		t.Error("expected non-nil reply")
	}
}

func TestDBTime(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("TIME", [][]byte{})
	if reply.ToBytes() == nil {
		t.Error("expected non-nil reply")
	}
}

func TestDBClientList(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("CLIENT", [][]byte{[]byte("LIST")})
	if reply.ToBytes() == nil {
		t.Error("expected non-nil reply")
	}
}

func TestDBClientGetName(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("CLIENT", [][]byte{[]byte("GETNAME")})
	if reply.ToBytes() == nil {
		t.Error("expected non-nil reply")
	}
}

func TestDBClientSetName(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("CLIENT", [][]byte{[]byte("SETNAME"), []byte("testclient")})
	if reply.ToBytes() == nil {
		t.Error("expected non-nil reply")
	}
}

func TestDBMultiExec(t *testing.T) {
	db := newTestDB()

	db.Exec("MULTI", [][]byte{})
	db.Exec("SET", [][]byte{[]byte("key1"), []byte("value1")})
	db.Exec("GET", [][]byte{[]byte("key1")})

	reply := db.Exec("EXEC", [][]byte{})
	if _, ok := reply.(*resp.ArrayReply); !ok {
		t.Errorf("expected ArrayReply, got %T", reply)
	}
}

func TestDBMultiDiscard(t *testing.T) {
	db := newTestDB()

	db.Exec("MULTI", [][]byte{})
	db.Exec("SET", [][]byte{[]byte("key1"), []byte("value1")})

	reply := db.Exec("DISCARD", [][]byte{})
	if getSimpleStringReply(reply) != "OK" {
		t.Errorf("expected OK, got %s", getSimpleStringReply(reply))
	}
}

func TestDBDiscardNoMulti(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("DISCARD", [][]byte{})
	if reply.ToBytes() == nil {
		t.Error("expected non-nil reply")
	}
}

func TestDBSelect(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("SELECT", [][]byte{[]byte("0")})
	if getSimpleStringReply(reply) != "OK" {
		t.Errorf("expected OK, got %s", getSimpleStringReply(reply))
	}
}

func TestDBQuit(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("QUIT", [][]byte{})
	if reply.ToBytes() == nil {
		t.Error("expected non-nil reply")
	}
}

func TestDBSlowlogGet(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("SLOWLOG", [][]byte{[]byte("GET")})
	if reply.ToBytes() == nil {
		t.Error("expected non-nil reply")
	}
}

func TestDBSlowlogLen(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("SLOWLOG", [][]byte{[]byte("LEN")})
	if reply.ToBytes() == nil {
		t.Error("expected non-nil reply")
	}
}

func TestDBCommand(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("COMMAND", [][]byte{})
	if reply.ToBytes() == nil {
		t.Error("expected non-nil reply")
	}
}

func TestDBCommandInfo(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("COMMAND", [][]byte{[]byte("INFO"), []byte("get")})
	if reply.ToBytes() == nil {
		t.Error("expected non-nil reply")
	}
}

func TestSetNX(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("SETNX", [][]byte{[]byte("key1"), []byte("value1")})
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}

	reply = db.Exec("SETNX", [][]byte{[]byte("key1"), []byte("value2")})
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0, got %d", getIntReply(reply))
	}

	reply = db.Exec("GET", [][]byte{[]byte("key1")})
	if getBulkReply(reply) != "value1" {
		t.Errorf("expected value1, got %s", getBulkReply(reply))
	}
}

func TestSetEX(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("SETEX", [][]byte{[]byte("key1"), []byte("10"), []byte("value1")})
	if getSimpleStringReply(reply) != "OK" {
		t.Errorf("expected OK, got %s", getSimpleStringReply(reply))
	}

	reply = db.Exec("GET", [][]byte{[]byte("key1")})
	if getBulkReply(reply) != "value1" {
		t.Errorf("expected value1, got %s", getBulkReply(reply))
	}

	reply = db.Exec("SETEX", [][]byte{[]byte("key2"), []byte("invalid"), []byte("value2")})
	if _, ok := reply.(*resp.ErrorReply); !ok {
		t.Error("expected error reply for invalid expire time")
	}
}

func TestPSetEX(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("PSETEX", [][]byte{[]byte("key1"), []byte("10000"), []byte("value1")})
	if getSimpleStringReply(reply) != "OK" {
		t.Errorf("expected OK, got %s", getSimpleStringReply(reply))
	}

	reply = db.Exec("GET", [][]byte{[]byte("key1")})
	if getBulkReply(reply) != "value1" {
		t.Errorf("expected value1, got %s", getBulkReply(reply))
	}

	reply = db.Exec("PSETEX", [][]byte{[]byte("key2"), []byte("invalid"), []byte("value2")})
	if _, ok := reply.(*resp.ErrorReply); !ok {
		t.Error("expected error reply for invalid expire time")
	}
}

func TestSetStringWithExpire(t *testing.T) {
	db := newTestDB()

	db.SetStringWithExpire("key1", []byte("value1"), 0)
	data, ok := db.GetStringData("key1")
	if !ok || string(data.value) != "value1" {
		t.Error("expected key1 to exist with value1")
	}

	future := int64(1893456000000)
	db.SetStringWithExpire("key2", []byte("value2"), future)
	data, ok = db.GetStringData("key2")
	if !ok || string(data.value) != "value2" || data.expireAt != future {
		t.Error("expected key2 to exist with value2 and correct expire time")
	}
}

func TestHDel(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("HDEL", [][]byte{[]byte("hash1"), []byte("field1")})
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0, got %d", getIntReply(reply))
	}

	reply = db.Exec("HSET", [][]byte{[]byte("hash1"), []byte("field1"), []byte("value1"), []byte("field2"), []byte("value2"), []byte("field3"), []byte("value3")})
	if getIntReply(reply) != 3 {
		t.Errorf("expected 3, got %d", getIntReply(reply))
	}

	reply = db.Exec("HDEL", [][]byte{[]byte("hash1"), []byte("field1")})
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}

	reply = db.Exec("HEXISTS", [][]byte{[]byte("hash1"), []byte("field1")})
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0, got %d", getIntReply(reply))
	}

	reply = db.Exec("HDEL", [][]byte{[]byte("hash1"), []byte("field2"), []byte("field4")})
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}

	reply = db.Exec("HDEL", [][]byte{[]byte("hash1"), []byte("field3")})
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}

	reply = db.Exec("EXISTS", [][]byte{[]byte("hash1")})
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0, got %d", getIntReply(reply))
	}
}

func TestCommandList(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("COMMAND", [][]byte{[]byte("LIST")})
	arrayReply := getArrayReply(reply)
	if arrayReply == nil || len(arrayReply) == 0 {
		t.Error("expected non-empty command list")
	}
}

func TestClientKill(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("CLIENT", [][]byte{[]byte("KILL"), []byte("127.0.0.1:6379")})
	if _, ok := reply.(*resp.ErrorReply); !ok {
		t.Error("expected error reply for unimplemented CLIENT KILL")
	}
}

func TestClientInfo(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("CLIENT", [][]byte{[]byte("INFO")})
	if getBulkReply(reply) != "" {
		t.Errorf("expected empty string, got %s", getBulkReply(reply))
	}

	clientID := AddClient("127.0.0.1:12345", 0)
	defer RemoveClient(clientID)

	reply = db.Exec("CLIENT", [][]byte{[]byte("INFO")})
	if getBulkReply(reply) == "" {
		t.Error("expected non-empty client info")
	}
}

func TestGetShutdownChan(t *testing.T) {
	ch := GetShutdownChan()
	if ch == nil {
		t.Error("expected non-nil shutdown channel")
	}
}

func TestDBLPop(t *testing.T) {
	db := newTestDB()

	// Test pop from empty list
	reply := db.Exec("LPOP", [][]byte{[]byte("mylist")})
	if getBulkReply(reply) != "" {
		t.Errorf("expected empty reply for lpop from empty list, got '%s'", getBulkReply(reply))
	}

	// Add elements to list
	db.Exec("LPUSH", [][]byte{[]byte("mylist"), []byte("a"), []byte("b"), []byte("c")})

	// Pop first element (should be 'c')
	reply = db.Exec("LPOP", [][]byte{[]byte("mylist")})
	if getBulkReply(reply) != "c" {
		t.Errorf("expected 'c', got '%s'", getBulkReply(reply))
	}

	// Check remaining length
	reply = db.Exec("LLEN", [][]byte{[]byte("mylist")})
	if getIntReply(reply) != 2 {
		t.Errorf("expected length 2, got %d", getIntReply(reply))
	}

	// Pop remaining elements
	reply = db.Exec("LPOP", [][]byte{[]byte("mylist")})
	if getBulkReply(reply) != "b" {
		t.Errorf("expected 'b', got '%s'", getBulkReply(reply))
	}

	reply = db.Exec("LPOP", [][]byte{[]byte("mylist")})
	if getBulkReply(reply) != "a" {
		t.Errorf("expected 'a', got '%s'", getBulkReply(reply))
	}

	// Pop again should return null
	reply = db.Exec("LPOP", [][]byte{[]byte("mylist")})
	if getBulkReply(reply) != "" {
		t.Errorf("expected empty reply for lpop from empty list, got '%s'", getBulkReply(reply))
	}
}

func TestDBRPop(t *testing.T) {
	db := newTestDB()

	// Test pop from empty list
	reply := db.Exec("RPOP", [][]byte{[]byte("mylist")})
	if getBulkReply(reply) != "" {
		t.Errorf("expected empty reply for rpop from empty list, got '%s'", getBulkReply(reply))
	}

	// Add elements to list
	db.Exec("RPUSH", [][]byte{[]byte("mylist"), []byte("a"), []byte("b"), []byte("c")})

	// Pop last element (should be 'c')
	reply = db.Exec("RPOP", [][]byte{[]byte("mylist")})
	if getBulkReply(reply) != "c" {
		t.Errorf("expected 'c', got '%s'", getBulkReply(reply))
	}

	// Check remaining length
	reply = db.Exec("LLEN", [][]byte{[]byte("mylist")})
	if getIntReply(reply) != 2 {
		t.Errorf("expected length 2, got %d", getIntReply(reply))
	}

	// Pop remaining elements
	reply = db.Exec("RPOP", [][]byte{[]byte("mylist")})
	if getBulkReply(reply) != "b" {
		t.Errorf("expected 'b', got '%s'", getBulkReply(reply))
	}

	reply = db.Exec("RPOP", [][]byte{[]byte("mylist")})
	if getBulkReply(reply) != "a" {
		t.Errorf("expected 'a', got '%s'", getBulkReply(reply))
	}

	// Pop again should return null
	reply = db.Exec("RPOP", [][]byte{[]byte("mylist")})
	if getBulkReply(reply) != "" {
		t.Errorf("expected empty reply for rpop from empty list, got '%s'", getBulkReply(reply))
	}
}

func TestDBLTrimEdgeCases(t *testing.T) {
	db := newTestDB()

	// Add elements
	db.Exec("RPUSH", [][]byte{[]byte("mylist"), []byte("a"), []byte("b"), []byte("c"), []byte("d"), []byte("e")})

	// Trim with negative indexes: keep last 3 elements ["c", "d", "e"]
	reply := db.Exec("LTRIM", [][]byte{[]byte("mylist"), []byte("-3"), []byte("-1")})
	if getSimpleStringReply(reply) != "OK" {
		t.Errorf("expected OK, got %s", getSimpleStringReply(reply))
	}

	// Check length
	reply = db.Exec("LLEN", [][]byte{[]byte("mylist")})
	if getIntReply(reply) != 3 {
		t.Errorf("expected length 3, got %d", getIntReply(reply))
	}

	// Trim with start > end: empty list
	reply = db.Exec("LTRIM", [][]byte{[]byte("mylist"), []byte("1"), []byte("0")})
	if getSimpleStringReply(reply) != "OK" {
		t.Errorf("expected OK, got %s", getSimpleStringReply(reply))
	}

	// Check length is zero
	reply = db.Exec("LLEN", [][]byte{[]byte("mylist")})
	if getIntReply(reply) != 0 {
		t.Errorf("expected length 0, got %d", getIntReply(reply))
	}

	// Trim non-existent list: should return OK
	reply = db.Exec("LTRIM", [][]byte{[]byte("nonexistent"), []byte("0"), []byte("1")})
	if getSimpleStringReply(reply) != "OK" {
		t.Errorf("expected OK, got %s", getSimpleStringReply(reply))
	}

	// Add elements again
	db.Exec("RPUSH", [][]byte{[]byte("mylist"), []byte("a"), []byte("b"), []byte("c")})

	// Trim with end larger than list length: keep all elements
	reply = db.Exec("LTRIM", [][]byte{[]byte("mylist"), []byte("0"), []byte("100")})
	if getSimpleStringReply(reply) != "OK" {
		t.Errorf("expected OK, got %s", getSimpleStringReply(reply))
	}

	// Check length is still 3
	reply = db.Exec("LLEN", [][]byte{[]byte("mylist")})
	if getIntReply(reply) != 3 {
		t.Errorf("expected length 3, got %d", getIntReply(reply))
	}
}

// TODO: Fix ZRangeByScore edge case test
// func TestDBZRangeByScoreEdgeCases(t *testing.T) {
// 	db := newTestDB()

// 	// Add elements
// 	db.Exec("ZADD", [][]byte{[]byte("zset"), []byte("1"), []byte("a"), []byte("2"), []byte("b"), []byte("3"), []byte("c"), []byte("4"), []byte("d"), []byte("5"), []byte("e")})

// 	// Test with min >= max: empty result
// 	reply := db.Exec("ZRANGEBYSCORE", [][]byte{[]byte("zset"), []byte("5"), []byte("1")})
// 	arrayReply := getArrayReply(reply)
// 	if len(arrayReply) != 0 {
// 		t.Errorf("expected empty array, got %d elements", len(arrayReply))
// 	}

// 	// Test with exclusive min and max: (1 (5) should return 2,3,4 -> b, c, d
// 	reply = db.Exec("ZRANGEBYSCORE", [][]byte{[]byte("zset"), []byte("(1"), []byte("(5")})
// 	arrayReply = getArrayReply(reply)
// 	if len(arrayReply) != 3 {
// 		t.Errorf("expected 3 elements, got %d", len(arrayReply))
// 	}

// 	// Test with LIMIT offset count: limit 2 2 should return c, d
// 	reply = db.Exec("ZRANGEBYSCORE", [][]byte{[]byte("zset"), []byte("1"), []byte("5"), []byte("LIMIT"), []byte("2"), []byte("2")})
// 	arrayReply = getArrayReply(reply)
// 	if len(arrayReply) != 2 {
// 		t.Errorf("expected 2 elements, got %d", len(arrayReply))
// 	}

// 	// Test with -inf and +inf: return all elements
// 	reply = db.Exec("ZRANGEBYSCORE", [][]byte{[]byte("zset"), []byte("-inf"), []byte("+inf")})
// 	arrayReply = getArrayReply(reply)
// 	if len(arrayReply) != 5 {
// 		t.Errorf("expected 5 elements, got %d", len(arrayReply))
// 	}

// 	// Test with non-existent key: empty array
// 	reply = db.Exec("ZRANGEBYSCORE", [][]byte{[]byte("nonexistent"), []byte("0"), []byte("10")})
// 	arrayReply = getArrayReply(reply)
// 	if len(arrayReply) != 0 {
// 		t.Errorf("expected empty array for non-existent key, got %d elements", len(arrayReply))
// 	}
// }
