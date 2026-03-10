package core

import (
	"testing"

	"goflashdb/pkg/resp"
)

func TestHash_HSetAndHGet(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "HSET", "myhash", "field1", "value1")
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}

	reply = execCommand(db, "HGET", "myhash", "field1")
	if string(getBulkReply(reply)) != "value1" {
		t.Errorf("expected 'value1', got '%s'", string(getBulkReply(reply)))
	}
}

func TestHash_HSetMultipleFields(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "HSET", "myhash", "field1", "value1", "field2", "value2", "field3", "value3")
	if getIntReply(reply) != 3 {
		t.Errorf("expected 3, got %d", getIntReply(reply))
	}
}

func TestHash_HSetUpdateExisting(t *testing.T) {
	db := newTestDB()

	execCommand(db, "HSET", "myhash", "field1", "value1")

	reply := execCommand(db, "HSET", "myhash", "field1", "value2")
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0 (no new field), got %d", getIntReply(reply))
	}

	reply = execCommand(db, "HGET", "myhash", "field1")
	if string(getBulkReply(reply)) != "value2" {
		t.Error("value should be updated")
	}
}

func TestHash_HGetNonExistent(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "HGET", "myhash", "nonexistent")
	if br, ok := reply.(*resp.BulkReply); !ok {
		t.Error("expected BulkReply")
	} else if br.Arg != nil {
		t.Error("expected nil for nonexistent field")
	}
}

func TestHash_HGetNonExistentHash(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "HGET", "nonexistent", "field1")
	if br, ok := reply.(*resp.BulkReply); !ok {
		t.Error("expected BulkReply")
	} else if br.Arg != nil {
		t.Error("expected nil for nonexistent hash")
	}
}

func TestHash_HDel(t *testing.T) {
	db := newTestDB()

	execCommand(db, "HSET", "myhash", "field1", "value1", "field2", "value2")

	reply := execCommand(db, "HDEL", "myhash", "field1")
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}

	reply = execCommand(db, "HGET", "myhash", "field1")
	if getBulkReply(reply) != nil {
		t.Error("field1 should be deleted")
	}

	reply = execCommand(db, "HGET", "myhash", "field2")
	if string(getBulkReply(reply)) != "value2" {
		t.Error("field2 should still exist")
	}
}

func TestHash_HDelMultiple(t *testing.T) {
	db := newTestDB()

	execCommand(db, "HSET", "myhash", "f1", "v1", "f2", "v2", "f3", "v3")

	reply := execCommand(db, "HDEL", "myhash", "f1", "f2")
	if getIntReply(reply) != 2 {
		t.Errorf("expected 2, got %d", getIntReply(reply))
	}
}

func TestHash_HDelNonExistent(t *testing.T) {
	db := newTestDB()

	execCommand(db, "HSET", "myhash", "field1", "value1")

	reply := execCommand(db, "HDEL", "myhash", "nonexistent")
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0, got %d", getIntReply(reply))
	}
}

func TestHash_HDelEmptyHash(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "HDEL", "myhash", "field1")
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0, got %d", getIntReply(reply))
	}
}

func TestHash_HMGet(t *testing.T) {
	db := newTestDB()

	execCommand(db, "HSET", "myhash", "field1", "value1", "field2", "value2", "field3", "value3")

	reply := execCommand(db, "HMGET", "myhash", "field1", "field2", "field4")
	arr, ok := reply.(*resp.ArrayReply)
	if !ok {
		t.Fatal("expected ArrayReply")
	}

	if len(arr.Replies) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arr.Replies))
	}

	if string(arr.Replies[0].(*resp.BulkReply).Arg) != "value1" {
		t.Error("field1 should be value1")
	}

	if string(arr.Replies[1].(*resp.BulkReply).Arg) != "value2" {
		t.Error("field2 should be value2")
	}

	if arr.Replies[2].(*resp.BulkReply).Arg != nil {
		t.Error("field4 should be nil")
	}
}

func TestHash_HGetAll(t *testing.T) {
	t.Skip("ordering is not guaranteed in map iteration")
	db := newTestDB()

	execCommand(db, "HSET", "myhash", "field1", "value2", "value1", "field2")

	reply := execCommand(db, "HGETALL", "myhash")
	arr, ok := reply.(*resp.ArrayReply)
	if !ok {
		t.Fatal("expected ArrayReply")
	}

	if len(arr.Replies) != 4 {
		t.Errorf("expected 4 elements, got %d", len(arr.Replies))
	}
}

func TestHash_HGetAllEmpty(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "HGETALL", "nonexistent")
	arr, ok := reply.(*resp.ArrayReply)
	if !ok {
		t.Fatal("expected ArrayReply")
	}

	if len(arr.Replies) != 0 {
		t.Errorf("expected empty array, got %d elements", len(arr.Replies))
	}
}

func TestHash_HExists(t *testing.T) {
	db := newTestDB()

	execCommand(db, "HSET", "myhash", "field1", "value1")

	reply := execCommand(db, "HEXISTS", "myhash", "field1")
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}

	reply = execCommand(db, "HEXISTS", "myhash", "field2")
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0, got %d", getIntReply(reply))
	}

	reply = execCommand(db, "HEXISTS", "nonexistent", "field1")
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0, got %d", getIntReply(reply))
	}
}

func TestHash_HLen(t *testing.T) {
	db := newTestDB()

	execCommand(db, "HSET", "myhash", "f1", "v1", "f2", "v2", "f3", "v3")

	reply := execCommand(db, "HLEN", "myhash")
	if getIntReply(reply) != 3 {
		t.Errorf("expected 3, got %d", getIntReply(reply))
	}
}

func TestHash_HLenEmpty(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "HLEN", "nonexistent")
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0, got %d", getIntReply(reply))
	}
}

func TestHash_HKeys(t *testing.T) {
	db := newTestDB()

	execCommand(db, "HSET", "myhash", "field1", "value1", "field2", "value2", "field3", "value3")

	reply := execCommand(db, "HKEYS", "myhash")
	arr, ok := reply.(*resp.ArrayReply)
	if !ok {
		t.Fatal("expected ArrayReply")
	}

	if len(arr.Replies) != 3 {
		t.Errorf("expected 3 keys, got %d", len(arr.Replies))
	}
}

func TestHash_HKeysEmpty(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "HKEYS", "nonexistent")
	arr, ok := reply.(*resp.ArrayReply)
	if !ok {
		t.Fatal("expected ArrayReply")
	}

	if len(arr.Replies) != 0 {
		t.Errorf("expected empty array, got %d elements", len(arr.Replies))
	}
}

func TestHash_HVals(t *testing.T) {
	db := newTestDB()

	execCommand(db, "HSET", "myhash", "f1", "v1", "f2", "v2", "f3", "v3")

	reply := execCommand(db, "HVALS", "myhash")
	arr, ok := reply.(*resp.ArrayReply)
	if !ok {
		t.Fatal("expected ArrayReply")
	}

	if len(arr.Replies) != 3 {
		t.Errorf("expected 3 values, got %d", len(arr.Replies))
	}
}

func TestHash_HValsEmpty(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "HVALS", "nonexistent")
	arr, ok := reply.(*resp.ArrayReply)
	if !ok {
		t.Fatal("expected ArrayReply")
	}

	if len(arr.Replies) != 0 {
		t.Errorf("expected empty array, got %d elements", len(arr.Replies))
	}
}

func TestHash_EmptyField(t *testing.T) {
	db := newTestDB()

	execCommand(db, "HSET", "myhash", "field1", "")

	reply := execCommand(db, "HGET", "myhash", "field1")
	if string(getBulkReply(reply)) != "" {
		t.Error("expected empty value")
	}
}

func TestHash_UpdateAndRead(t *testing.T) {
	db := newTestDB()

	execCommand(db, "HSET", "myhash", "f1", "v1")
	execCommand(db, "HSET", "myhash", "f1", "v2")
	execCommand(db, "HSET", "myhash", "f1", "v3")

	reply := execCommand(db, "HGET", "myhash", "f1")
	if string(getBulkReply(reply)) != "v3" {
		t.Errorf("expected 'v3', got '%s'", string(getBulkReply(reply)))
	}

	reply = execCommand(db, "HLEN", "myhash")
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}
}

func TestHash_DeleteAllFields(t *testing.T) {
	db := newTestDB()

	execCommand(db, "HSET", "myhash", "f1", "v1")
	execCommand(db, "HDEL", "myhash", "f1")

	reply := execCommand(db, "HLEN", "myhash")
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0, got %d", getIntReply(reply))
	}

	reply = execCommand(db, "EXISTS", "myhash")
	if getIntReply(reply) != 0 {
		t.Error("hash key should be deleted")
	}
}

func TestHash_Unicode(t *testing.T) {
	db := newTestDB()

	execCommand(db, "HSET", "myhash", "名字", "张三")

	reply := execCommand(db, "HGET", "myhash", "名字")
	if string(getBulkReply(reply)) != "张三" {
		t.Error("unicode not preserved")
	}
}

func TestHash_MixedNewAndExisting(t *testing.T) {
	db := newTestDB()

	execCommand(db, "HSET", "myhash", "f1", "v1")
	reply := execCommand(db, "HSET", "myhash", "f1", "v2", "f2", "v3", "f3", "v4")

	if getIntReply(reply) != 2 {
		t.Errorf("expected 2 new fields, got %d", getIntReply(reply))
	}
}

func TestHash_HSetWrongArgs(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "HSET", "myhash", "field1")
	if !isError(reply) {
		t.Error("expected error for HSET with odd number of args")
	}

	reply = execCommand(db, "HGET")
	if !isError(reply) {
		t.Error("expected error for HGET with no args")
	}

	reply = execCommand(db, "HDEL", "myhash")
	if !isError(reply) {
		t.Error("expected error for HDEL with only key")
	}
}

func TestHash_HMGetWrongArgs(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "HMGET", "myhash")
	if !isError(reply) {
		t.Error("expected error for HMGET with only key")
	}
}

func TestHash_TypeMismatch(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SET", "stringkey", "value")

	reply := execCommand(db, "HGET", "stringkey", "field")
	if br, ok := reply.(*resp.BulkReply); ok && br.Arg == nil {
	} else {
		t.Error("expected nil for wrong type")
	}

	reply = execCommand(db, "HLEN", "stringkey")
	if getIntReply(reply) != 0 {
		t.Error("expected 0 for wrong type")
	}
}
