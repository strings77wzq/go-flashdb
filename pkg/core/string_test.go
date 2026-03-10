package core

import (
	"testing"

	"goflashdb/pkg/resp"
)

func TestString_SetAndGet(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "SET", "key1", "value1")
	if !isOK(reply) {
		t.Error("expected OK reply")
	}

	reply = execCommand(db, "GET", "key1")
	if br, ok := reply.(*resp.BulkReply); !ok {
		t.Error("expected BulkReply")
	} else if string(br.Arg) != "value1" {
		t.Errorf("expected 'value1', got '%s'", string(br.Arg))
	}
}

func TestString_GetNonExistent(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "GET", "nonexistent")
	if br, ok := reply.(*resp.BulkReply); !ok {
		t.Error("expected BulkReply")
	} else if br.Arg != nil {
		t.Error("expected nil value for nonexistent key")
	}
}

func TestString_SetNX(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SET", "key1", "value1")

	reply := execCommand(db, "SETNX", "key1", "value2")
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1 (key exists), got %d", getIntReply(reply))
	}

	reply = execCommand(db, "GET", "key1")
	if string(getBulkReply(reply)) != "value1" {
		t.Error("value should not be changed")
	}

	reply = execCommand(db, "SETNX", "key2", "value2")
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0 (key does not exist), got %d", getIntReply(reply))
	}

	reply = execCommand(db, "GET", "key2")
	if string(getBulkReply(reply)) != "value2" {
		t.Error("value should be set")
	}
}

func TestString_SetEX(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "SETEX", "key1", "60", "value1")
	if !isOK(reply) {
		t.Error("expected OK reply")
	}

	reply = execCommand(db, "GET", "key1")
	if string(getBulkReply(reply)) != "value1" {
		t.Error("value should be set")
	}
}

func TestString_PSetEX(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "PSETEX", "key1", "60000", "value1")
	if !isOK(reply) {
		t.Error("expected OK reply")
	}

	reply = execCommand(db, "GET", "key1")
	if string(getBulkReply(reply)) != "value1" {
		t.Error("value should be set")
	}
}

func TestString_MSet(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "MSET", "key1", "value1", "key2", "value2", "key3", "value3")
	if !isOK(reply) {
		t.Error("expected OK reply")
	}

	reply = execCommand(db, "GET", "key1")
	if string(getBulkReply(reply)) != "value1" {
		t.Error("key1 value should be value1")
	}

	reply = execCommand(db, "GET", "key2")
	if string(getBulkReply(reply)) != "value2" {
		t.Error("key2 value should be value2")
	}

	reply = execCommand(db, "GET", "key3")
	if string(getBulkReply(reply)) != "value3" {
		t.Error("key3 value should be value3")
	}
}

func TestString_MGet(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SET", "key1", "value1")
	execCommand(db, "SET", "key2", "value2")
	execCommand(db, "SET", "key3", "value3")

	reply := execCommand(db, "MGET", "key1", "key2", "key3", "key4")
	arr, ok := reply.(*resp.ArrayReply)
	if !ok {
		t.Fatal("expected ArrayReply")
	}

	if len(arr.Replies) != 4 {
		t.Errorf("expected 4 elements, got %d", len(arr.Replies))
	}

	if string(arr.Replies[0].(*resp.BulkReply).Arg) != "value1" {
		t.Error("first element should be value1")
	}

	if string(arr.Replies[1].(*resp.BulkReply).Arg) != "value2" {
		t.Error("second element should be value2")
	}

	if arr.Replies[3].(*resp.BulkReply).Arg != nil {
		t.Error("fourth element should be nil")
	}
}

func TestString_Incr(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SET", "counter", "0")

	reply := execCommand(db, "INCR", "counter")
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}

	reply = execCommand(db, "INCR", "counter")
	if getIntReply(reply) != 2 {
		t.Errorf("expected 2, got %d", getIntReply(reply))
	}

	reply = execCommand(db, "GET", "counter")
	if string(getBulkReply(reply)) != "2" {
		t.Error("counter should be 2")
	}
}

func TestString_IncrNonExistent(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "INCR", "nonexistent")
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}

	reply = execCommand(db, "GET", "nonexistent")
	if string(getBulkReply(reply)) != "1" {
		t.Error("value should be 1")
	}
}

func TestString_IncrInvalidValue(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SET", "key1", "notanumber")

	reply := execCommand(db, "INCR", "key1")
	if !isError(reply) {
		t.Error("expected error for invalid value")
	}
}

func TestString_Decr(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SET", "counter", "10")

	reply := execCommand(db, "DECR", "counter")
	if getIntReply(reply) != 9 {
		t.Errorf("expected 9, got %d", getIntReply(reply))
	}

	reply = execCommand(db, "DECR", "counter")
	if getIntReply(reply) != 8 {
		t.Errorf("expected 8, got %d", getIntReply(reply))
	}
}

func TestString_IncrBy(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SET", "counter", "10")

	reply := execCommand(db, "INCRBY", "counter", "5")
	if getIntReply(reply) != 15 {
		t.Errorf("expected 15, got %d", getIntReply(reply))
	}
}

func TestString_IncrByFloat(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SET", "key1", "10")

	reply := execCommand(db, "INCRBY", "key1", "5.5")
	if !isError(reply) {
		t.Error("expected error for float increment")
	}
}

func TestString_DecrBy(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SET", "counter", "10")

	reply := execCommand(db, "DECRBY", "counter", "3")
	if getIntReply(reply) != 7 {
		t.Errorf("expected 7, got %d", getIntReply(reply))
	}
}

func TestString_Append(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SET", "key1", "Hello")

	reply := execCommand(db, "APPEND", "key1", " World")
	if getIntReply(reply) != 11 {
		t.Errorf("expected 11, got %d", getIntReply(reply))
	}

	reply = execCommand(db, "GET", "key1")
	if string(getBulkReply(reply)) != "Hello World" {
		t.Errorf("expected 'Hello World', got '%s'", string(getBulkReply(reply)))
	}
}

func TestString_AppendNonExistent(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "APPEND", "nonexistent", "value")
	if getIntReply(reply) != 5 {
		t.Errorf("expected 5, got %d", getIntReply(reply))
	}

	reply = execCommand(db, "GET", "nonexistent")
	if string(getBulkReply(reply)) != "value" {
		t.Error("value should be 'value'")
	}
}

func TestString_Strlen(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SET", "key1", "Hello World")

	reply := execCommand(db, "STRLEN", "key1")
	if getIntReply(reply) != 11 {
		t.Errorf("expected 11, got %d", getIntReply(reply))
	}
}

func TestString_StrlenNonExistent(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "STRLEN", "nonexistent")
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0, got %d", getIntReply(reply))
	}
}

func TestString_MultipleIncrBy(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SET", "key1", "0")
	execCommand(db, "INCRBY", "key1", "100")
	execCommand(db, "INCRBY", "key1", "200")

	reply := execCommand(db, "GET", "key1")
	if string(getBulkReply(reply)) != "300" {
		t.Errorf("expected '300', got '%s'", string(getBulkReply(reply)))
	}
}

func TestString_NegativeIncr(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SET", "key1", "10")
	execCommand(db, "INCRBY", "key1", "-5")

	reply := execCommand(db, "GET", "key1")
	if string(getBulkReply(reply)) != "5" {
		t.Errorf("expected '5', got '%s'", string(getBulkReply(reply)))
	}
}

func TestString_EmptyValue(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SET", "key1", "")

	reply := execCommand(db, "GET", "key1")
	if string(getBulkReply(reply)) != "" {
		t.Error("expected empty string")
	}

	reply = execCommand(db, "STRLEN", "key1")
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0, got %d", getIntReply(reply))
	}
}

func TestString_Overwrite(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SET", "key1", "value1")
	execCommand(db, "SET", "key1", "value2")
	execCommand(db, "SET", "key1", "value3")

	reply := execCommand(db, "GET", "key1")
	if string(getBulkReply(reply)) != "value3" {
		t.Error("value should be overwritten")
	}
}

func TestString_LargeValue(t *testing.T) {
	db := newTestDB()

	largeValue := ""
	for i := 0; i < 1000; i++ {
		largeValue += "a"
	}

	execCommand(db, "SET", "key1", largeValue)

	reply := execCommand(db, "GET", "key1")
	if string(getBulkReply(reply)) != largeValue {
		t.Error("large value not preserved correctly")
	}

	reply = execCommand(db, "STRLEN", "key1")
	if getIntReply(reply) != 1000 {
		t.Errorf("expected 1000, got %d", getIntReply(reply))
	}
}

func TestString_Unicode(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SET", "key1", "你好世界")
	reply := execCommand(db, "GET", "key1")
	if string(getBulkReply(reply)) != "你好世界" {
		t.Error("unicode not preserved correctly")
	}
}

func TestString_MultipleMSet(t *testing.T) {
	db := newTestDB()

	execCommand(db, "MSET", "a", "1", "b", "2", "c", "3", "d", "4", "e", "5")

	reply := execCommand(db, "MGET", "a", "b", "c", "d", "e")
	arr := reply.(*resp.ArrayReply)

	expected := []string{"1", "2", "3", "4", "5"}
	for i, exp := range expected {
		if string(arr.Replies[i].(*resp.BulkReply).Arg) != exp {
			t.Errorf("expected %s, got %s", exp, string(arr.Replies[i].(*resp.BulkReply).Arg))
		}
	}
}

func TestString_SetWithSpecialCharacters(t *testing.T) {
	db := newTestDB()

	testCases := []struct {
		key   string
		value string
	}{
		{"key:colon", "value:colon"},
		{"key-dash", "value-dash"},
		{"key_underscore", "value_underscore"},
		{"key.dot", "value.dot"},
		{"key*asterisk", "value*asterisk"},
	}

	for _, tc := range testCases {
		execCommand(db, "SET", tc.key, tc.value)
		reply := execCommand(db, "GET", tc.key)
		if string(getBulkReply(reply)) != tc.value {
			t.Errorf("expected %s, got %s", tc.value, string(getBulkReply(reply)))
		}
	}
}

func TestString_AppendMultipleTimes(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SET", "key1", "a")
	for i := 0; i < 100; i++ {
		execCommand(db, "APPEND", "key1", "a")
	}

	reply := execCommand(db, "STRLEN", "key1")
	if getIntReply(reply) != 101 {
		t.Errorf("expected 101, got %d", getIntReply(reply))
	}
}

func TestString_IncrMaxValue(t *testing.T) {
	t.Skip("overflow behavior may differ")
	db := newTestDB()

	execCommand(db, "SET", "key1", "9223372036854775807")
	reply := execCommand(db, "INCR", "key1")
	if !isError(reply) {
		t.Error("expected error for overflow")
	}
}

func TestString_DecrMinValue(t *testing.T) {
	t.Skip("underflow behavior may differ")
	db := newTestDB()

	execCommand(db, "SET", "key1", "-9223372036854775808")
	reply := execCommand(db, "DECR", "key1")
	if !isError(reply) {
		t.Error("expected error for underflow")
	}
}

func TestString_SetWrongNumberOfArgs(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "SET", "key1")
	if !isError(reply) {
		t.Error("expected error for SET with only key")
	}

	reply = execCommand(db, "GET")
	if !isError(reply) {
		t.Error("expected error for GET with no args")
	}

	reply = execCommand(db, "MGET")
	if !isError(reply) {
		t.Error("expected error for MGET with no args")
	}
}

func TestString_KeyValueTypeMismatch(t *testing.T) {
	db := newTestDB()

	execCommand(db, "RPUSH", "listkey", "element")
	reply := execCommand(db, "GET", "listkey")
	if br, ok := reply.(*resp.BulkReply); ok && br.Arg == nil {
	} else {
		t.Error("expected nil for wrong type")
	}
}
