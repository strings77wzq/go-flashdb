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

	reply := db.Exec("SETBIT", [][]byte{[]byte("bits"), []byte("0"), []byte("1")})
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0, got %d", getIntReply(reply))
	}

	reply = db.Exec("GETBIT", [][]byte{[]byte("bits"), []byte("0")})
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}

	reply = db.Exec("BITCOUNT", [][]byte{[]byte("bits")})
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}
}

func TestHyperLogLog(t *testing.T) {
	db := newTestDB()

	reply := db.Exec("PFADD", [][]byte{[]byte("hll"), []byte("a"), []byte("b"), []byte("c")})
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}

	reply = db.Exec("PFCOUNT", [][]byte{[]byte("hll")})
	if getIntReply(reply) < 1 {
		t.Errorf("expected at least 1, got %d", getIntReply(reply))
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
	arr := getArrayReply(reply)
	if len(arr) != 1 {
		t.Errorf("expected 1 reply, got %d", len(arr))
	}
}
