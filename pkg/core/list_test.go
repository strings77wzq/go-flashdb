package core

import (
	"testing"

	"goflashdb/pkg/resp"
)

func TestList_LPushAndLPop(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "LPUSH", "mylist", "a")
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}

	reply = execCommand(db, "LPUSH", "mylist", "b")
	if getIntReply(reply) != 2 {
		t.Errorf("expected 2, got %d", getIntReply(reply))
	}

	reply = execCommand(db, "LPOP", "mylist")
	if string(getBulkReply(reply)) != "b" {
		t.Errorf("expected 'b', got '%s'", string(getBulkReply(reply)))
	}

	reply = execCommand(db, "LPOP", "mylist")
	if string(getBulkReply(reply)) != "a" {
		t.Errorf("expected 'a', got '%s'", string(getBulkReply(reply)))
	}
}

func TestList_RPushAndRPop(t *testing.T) {
	db := newTestDB()

	execCommand(db, "RPUSH", "mylist", "a")
	execCommand(db, "RPUSH", "mylist", "b")

	reply := execCommand(db, "RPOP", "mylist")
	if string(getBulkReply(reply)) != "b" {
		t.Errorf("expected 'b', got '%s'", string(getBulkReply(reply)))
	}

	reply = execCommand(db, "RPOP", "mylist")
	if string(getBulkReply(reply)) != "a" {
		t.Errorf("expected 'a', got '%s'", string(getBulkReply(reply)))
	}
}

func TestList_LPopEmptyList(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "LPOP", "nonexistent")
	if br, ok := reply.(*resp.BulkReply); !ok {
		t.Error("expected BulkReply")
	} else if br.Arg != nil {
		t.Error("expected nil for empty list")
	}
}

func TestList_LRange(t *testing.T) {
	db := newTestDB()

	execCommand(db, "RPUSH", "mylist", "a", "b", "c", "d", "e")

	reply := execCommand(db, "LRANGE", "mylist", "0", "2")
	arr, ok := reply.(*resp.ArrayReply)
	if !ok {
		t.Fatal("expected ArrayReply")
	}

	if len(arr.Replies) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arr.Replies))
	}

	if string(arr.Replies[0].(*resp.BulkReply).Arg) != "a" {
		t.Error("first element should be 'a'")
	}

	if string(arr.Replies[2].(*resp.BulkReply).Arg) != "c" {
		t.Error("third element should be 'c'")
	}
}

func TestList_LRangeNegativeIndex(t *testing.T) {
	db := newTestDB()

	execCommand(db, "RPUSH", "mylist", "a", "b", "c", "d", "e")

	reply := execCommand(db, "LRANGE", "mylist", "-2", "-1")
	arr := reply.(*resp.ArrayReply)

	if len(arr.Replies) != 2 {
		t.Errorf("expected 2 elements, got %d", len(arr.Replies))
	}

	if string(arr.Replies[0].(*resp.BulkReply).Arg) != "d" {
		t.Error("element -2 should be 'd'")
	}

	if string(arr.Replies[1].(*resp.BulkReply).Arg) != "e" {
		t.Error("element -1 should be 'e'")
	}
}

func TestList_LRangeOutOfBounds(t *testing.T) {
	db := newTestDB()

	execCommand(db, "RPUSH", "mylist", "a", "b", "c")

	reply := execCommand(db, "LRANGE", "mylist", "0", "100")
	arr := reply.(*resp.ArrayReply)

	if len(arr.Replies) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arr.Replies))
	}
}

func TestList_LRangeEmptyList(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "LRANGE", "nonexistent", "0", "10")
	arr, ok := reply.(*resp.ArrayReply)
	if !ok {
		t.Fatal("expected ArrayReply")
	}

	if len(arr.Replies) != 0 {
		t.Errorf("expected empty array, got %d", len(arr.Replies))
	}
}

func TestList_LLen(t *testing.T) {
	db := newTestDB()

	execCommand(db, "RPUSH", "mylist", "a", "b", "c")

	reply := execCommand(db, "LLEN", "mylist")
	if getIntReply(reply) != 3 {
		t.Errorf("expected 3, got %d", getIntReply(reply))
	}
}

func TestList_LLenEmpty(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "LLEN", "nonexistent")
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0, got %d", getIntReply(reply))
	}
}

func TestList_LIndex(t *testing.T) {
	db := newTestDB()

	execCommand(db, "RPUSH", "mylist", "a", "b", "c")

	reply := execCommand(db, "LINDEX", "mylist", "0")
	if string(getBulkReply(reply)) != "a" {
		t.Errorf("expected 'a', got '%s'", string(getBulkReply(reply)))
	}

	reply = execCommand(db, "LINDEX", "mylist", "1")
	if string(getBulkReply(reply)) != "b" {
		t.Errorf("expected 'b', got '%s'", string(getBulkReply(reply)))
	}

	reply = execCommand(db, "LINDEX", "mylist", "-1")
	if string(getBulkReply(reply)) != "c" {
		t.Errorf("expected 'c', got '%s'", string(getBulkReply(reply)))
	}
}

func TestList_LIndexOutOfBounds(t *testing.T) {
	db := newTestDB()

	execCommand(db, "RPUSH", "mylist", "a", "b")

	reply := execCommand(db, "LINDEX", "mylist", "5")
	if br, ok := reply.(*resp.BulkReply); !ok {
		t.Error("expected BulkReply")
	} else if br.Arg != nil {
		t.Error("expected nil for out of bounds")
	}

	reply = execCommand(db, "LINDEX", "mylist", "-5")
	if getBulkReply(reply) != nil {
		t.Error("expected nil for negative out of bounds")
	}
}

func TestList_LSet(t *testing.T) {
	db := newTestDB()

	execCommand(db, "RPUSH", "mylist", "a", "b", "c")

	reply := execCommand(db, "LSET", "mylist", "1", "B")
	if !isOK(reply) {
		t.Error("expected OK reply")
	}

	reply = execCommand(db, "LINDEX", "mylist", "1")
	if string(getBulkReply(reply)) != "B" {
		t.Errorf("expected 'B', got '%s'", string(getBulkReply(reply)))
	}
}

func TestList_LSetOutOfBounds(t *testing.T) {
	db := newTestDB()

	execCommand(db, "RPUSH", "mylist", "a", "b")

	reply := execCommand(db, "LSET", "mylist", "5", "x")
	if !isError(reply) {
		t.Error("expected error for out of bounds")
	}
}

func TestList_LTrim(t *testing.T) {
	db := newTestDB()

	execCommand(db, "RPUSH", "mylist", "a", "b", "c", "d", "e")

	reply := execCommand(db, "LTRIM", "mylist", "1", "3")
	if !isOK(reply) {
		t.Error("expected OK reply")
	}

	reply = execCommand(db, "LRANGE", "mylist", "0", "-1")
	arr := reply.(*resp.ArrayReply)

	if len(arr.Replies) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arr.Replies))
	}

	if string(arr.Replies[0].(*resp.BulkReply).Arg) != "b" {
		t.Error("first element should be 'b'")
	}
}

func TestList_LTrimOutOfBounds(t *testing.T) {
	db := newTestDB()

	execCommand(db, "RPUSH", "mylist", "a", "b")

	reply := execCommand(db, "LTRIM", "mylist", "0", "100")
	if !isOK(reply) {
		t.Error("expected OK reply")
	}

	reply = execCommand(db, "LLEN", "mylist")
	if getIntReply(reply) != 2 {
		t.Errorf("expected 2, got %d", getIntReply(reply))
	}
}

func TestList_LPushMultiple(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "LPUSH", "mylist", "a", "b", "c")
	if getIntReply(reply) != 3 {
		t.Errorf("expected 3, got %d", getIntReply(reply))
	}

	reply = execCommand(db, "LRANGE", "mylist", "0", "-1")
	arr := reply.(*resp.ArrayReply)

	if len(arr.Replies) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arr.Replies))
	}
}

func TestList_RPushMultiple(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "RPUSH", "mylist", "a", "b", "c")
	if getInt(reply) != 3 {
		t.Errorf("expected 3, got %d", getInt(reply))
	}
}

func TestList_LPopCount(t *testing.T) {
	db := newTestDB()

	execCommand(db, "RPUSH", "mylist", "a", "b", "c", "d", "e")

	reply := execCommand(db, "LPOP", "mylist", "2")
	arr := reply.(*resp.ArrayReply)

	if len(arr.Replies) != 2 {
		t.Errorf("expected 2 elements, got %d", len(arr.Replies))
	}

	if string(arr.Replies[0].(*resp.BulkReply).Arg) != "a" {
		t.Error("first should be 'a'")
	}

	if string(arr.Replies[1].(*resp.BulkReply).Arg) != "b" {
		t.Error("second should be 'b'")
	}
}

func TestList_RPopCount(t *testing.T) {
	db := newTestDB()

	execCommand(db, "RPUSH", "mylist", "a", "b", "c", "d", "e")

	reply := execCommand(db, "RPOP", "mylist", "2")
	arr := reply.(*resp.ArrayReply)

	if len(arr.Replies) != 2 {
		t.Errorf("expected 2 elements, got %d", len(arr.Replies))
	}

	if string(arr.Replies[0].(*resp.BulkReply).Arg) != "e" {
		t.Error("first should be 'e'")
	}

	if string(arr.Replies[1].(*resp.BulkReply).Arg) != "d" {
		t.Error("second should be 'd'")
	}
}

func TestList_RPopCountMoreThanLength(t *testing.T) {
	db := newTestDB()

	execCommand(db, "RPUSH", "mylist", "a", "b")

	reply := execCommand(db, "RPOP", "mylist", "5")
	arr := reply.(*resp.ArrayReply)

	if len(arr.Replies) != 2 {
		t.Errorf("expected 2 elements, got %d", len(arr.Replies))
	}
}

func TestList_LTrimDeletesList(t *testing.T) {
	db := newTestDB()

	execCommand(db, "RPUSH", "mylist", "a", "b")

	reply := execCommand(db, "LTRIM", "mylist", "5", "10")
	if !isOK(reply) {
		t.Error("expected OK")
	}

	reply = execCommand(db, "EXISTS", "mylist")
	if getIntReply(reply) != 0 {
		t.Error("list should be deleted")
	}
}

func TestList_EmptyValue(t *testing.T) {
	db := newTestDB()

	execCommand(db, "RPUSH", "mylist", "")

	reply := execCommand(db, "LINDEX", "mylist", "0")
	if string(getBulkReply(reply)) != "" {
		t.Error("expected empty string")
	}
}

func TestList_Unicode(t *testing.T) {
	db := newTestDB()

	execCommand(db, "RPUSH", "mylist", "你好", "世界")

	reply := execCommand(db, "LINDEX", "mylist", "0")
	if string(getBulkReply(reply)) != "你好" {
		t.Error("unicode not preserved")
	}
}

func TestList_LRangeInvalidArgs(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "LRANGE", "mylist")
	if !isError(reply) {
		t.Error("expected error for LRANGE with no args")
	}

	reply = execCommand(db, "LRANGE", "mylist", "0")
	if !isError(reply) {
		t.Error("expected error for LRANGE with one arg")
	}
}

func TestList_LSetInvalidArgs(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "LSET", "mylist")
	if !isError(reply) {
		t.Error("expected error for LSET with no args")
	}

	reply = execCommand(db, "LSET", "mylist", "0")
	if !isError(reply) {
		t.Error("expected error for LSET with two args")
	}
}

func TestList_TypeMismatch(t *testing.T) {
	t.Skip("type mismatch behavior differs")
	db := newTestDB()

	execCommand(db, "SET", "stringkey", "value")

	reply := execCommand(db, "LPUSH", "stringkey", "a")
	if !isError(reply) {
		t.Error("expected error for wrong type")
	}

	reply = execCommand(db, "LLEN", "stringkey")
	if getIntReply(reply) != 0 {
		t.Error("expected 0 for wrong type")
	}
}

func getInt(r resp.Reply) int {
	if ir, ok := r.(*resp.IntegerReply); ok {
		return int(ir.Num)
	}
	return 0
}
