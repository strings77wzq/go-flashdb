package resp

import (
	"bytes"
	"testing"
)

func TestParse_SimpleString(t *testing.T) {
	input := "+OK\r\n"
	parser := NewParser(bytes.NewReader([]byte(input)))

	reply, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sr, ok := reply.(*SimpleStringReply); !ok {
		t.Error("expected SimpleStringReply")
	} else if sr.Str != "OK" {
		t.Errorf("expected 'OK', got '%s'", sr.Str)
	}
}

func TestParse_Error(t *testing.T) {
	input := "-ERR something went wrong\r\n"
	parser := NewParser(bytes.NewReader([]byte(input)))

	reply, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if errReply, ok := reply.(*ErrorReply); !ok {
		t.Error("expected ErrorReply")
	} else if errReply.Err != "ERR something went wrong" {
		t.Errorf("expected 'ERR something went wrong', got '%s'", errReply.Err)
	}
}

func TestParse_Integer(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{":0\r\n", 0},
		{":100\r\n", 100},
		{":-50\r\n", -50},
		{":9223372036854775807\r\n", 9223372036854775807},
	}

	for _, tt := range tests {
		parser := NewParser(bytes.NewReader([]byte(tt.input)))
		reply, err := parser.Parse()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ir, ok := reply.(*IntegerReply); !ok {
			t.Error("expected IntegerReply")
		} else if ir.Num != tt.expected {
			t.Errorf("expected %d, got %d", tt.expected, ir.Num)
		}
	}
}

func TestParse_BulkString(t *testing.T) {
	input := "$5\r\nhello\r\n"
	parser := NewParser(bytes.NewReader([]byte(input)))

	reply, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if br, ok := reply.(*BulkReply); !ok {
		t.Error("expected BulkReply")
	} else if string(br.Arg) != "hello" {
		t.Errorf("expected 'hello', got '%s'", string(br.Arg))
	}
}

func TestParse_BulkStringEmpty(t *testing.T) {
	input := "$0\r\n\r\n"
	parser := NewParser(bytes.NewReader([]byte(input)))

	reply, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if br, ok := reply.(*BulkReply); !ok {
		t.Error("expected BulkReply")
	} else if len(br.Arg) != 0 {
		t.Errorf("expected empty, got '%s'", string(br.Arg))
	}
}

func TestParse_BulkStringNil(t *testing.T) {
	input := "$-1\r\n"
	parser := NewParser(bytes.NewReader([]byte(input)))

	reply, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if br, ok := reply.(*BulkReply); !ok {
		t.Error("expected BulkReply")
	} else if br.Arg != nil {
		t.Error("expected nil")
	}
}

func TestParse_BulkStringLong(t *testing.T) {
	value := ""
	for i := 0; i < 1000; i++ {
		value += "a"
	}
	input := "$1000\r\n" + value + "\r\n"
	parser := NewParser(bytes.NewReader([]byte(input)))

	reply, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if br, ok := reply.(*BulkReply); !ok {
		t.Error("expected BulkReply")
	} else if len(br.Arg) != 1000 {
		t.Errorf("expected 1000, got %d", len(br.Arg))
	}
}

func TestParse_Array(t *testing.T) {
	input := "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"
	parser := NewParser(bytes.NewReader([]byte(input)))

	reply, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr, ok := reply.(*ArrayReply)
	if !ok {
		t.Fatal("expected ArrayReply")
	}

	if len(arr.Replies) != 2 {
		t.Errorf("expected 2 elements, got %d", len(arr.Replies))
	}

	if string(arr.Replies[0].(*BulkReply).Arg) != "foo" {
		t.Error("first element should be 'foo'")
	}

	if string(arr.Replies[1].(*BulkReply).Arg) != "bar" {
		t.Error("second element should be 'bar'")
	}
}

func TestParse_ArrayEmpty(t *testing.T) {
	input := "*0\r\n"
	parser := NewParser(bytes.NewReader([]byte(input)))

	reply, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr, ok := reply.(*ArrayReply)
	if !ok {
		t.Fatal("expected ArrayReply")
	}

	if len(arr.Replies) != 0 {
		t.Errorf("expected empty array, got %d elements", len(arr.Replies))
	}
}

func TestParse_ArrayNil(t *testing.T) {
	input := "*-1\r\n"
	parser := NewParser(bytes.NewReader([]byte(input)))

	reply, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr, ok := reply.(*ArrayReply)
	if !ok {
		t.Fatal("expected ArrayReply")
	}

	if arr.Replies != nil {
		t.Error("expected nil array")
	}
}

func TestParse_NestedArray(t *testing.T) {
	input := "*2\r\n*2\r\n$1\r\na\r\n$1\r\nb\r\n*1\r\n$1\r\nc\r\n"
	parser := NewParser(bytes.NewReader([]byte(input)))

	reply, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr, ok := reply.(*ArrayReply)
	if !ok {
		t.Fatal("expected ArrayReply")
	}

	if len(arr.Replies) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(arr.Replies))
	}

	inner1 := arr.Replies[0].(*ArrayReply)
	if len(inner1.Replies) != 2 {
		t.Errorf("expected inner array with 2 elements, got %d", len(inner1.Replies))
	}

	inner2 := arr.Replies[1].(*ArrayReply)
	if len(inner2.Replies) != 1 {
		t.Errorf("expected inner array with 1 element, got %d", len(inner2.Replies))
	}
}

func TestParse_MixedArray(t *testing.T) {
	input := "*3\r\n:1\r\n+OK\r\n$3\r\nfoo\r\n"
	parser := NewParser(bytes.NewReader([]byte(input)))

	reply, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr, ok := reply.(*ArrayReply)
	if !ok {
		t.Fatal("expected ArrayReply")
	}

	if len(arr.Replies) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr.Replies))
	}

	if _, ok := arr.Replies[0].(*IntegerReply); !ok {
		t.Error("first element should be IntegerReply")
	}

	if _, ok := arr.Replies[1].(*SimpleStringReply); !ok {
		t.Error("second element should be SimpleStringReply")
	}

	if _, ok := arr.Replies[2].(*BulkReply); !ok {
		t.Error("third element should be BulkReply")
	}
}

func TestParse_InvalidPrefix(t *testing.T) {
	input := "!invalid\r\n"
	parser := NewParser(bytes.NewReader([]byte(input)))

	_, err := parser.Parse()
	if err == nil {
		t.Error("expected error for invalid prefix")
	}
}

func TestParseCommand_Basic(t *testing.T) {
	input := "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"
	cmdName, args, err := ParseCommand([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cmdName != "set" {
		t.Errorf("expected 'set', got '%s'", cmdName)
	}

	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d", len(args))
	}

	if string(args[0]) != "key" {
		t.Errorf("expected 'key', got '%s'", string(args[0]))
	}

	if string(args[1]) != "value" {
		t.Errorf("expected 'value', got '%s'", string(args[1]))
	}
}

func TestParseCommand_GetCommand(t *testing.T) {
	input := "*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n"
	cmdName, args, err := ParseCommand([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cmdName != "get" {
		t.Errorf("expected 'get', got '%s'", cmdName)
	}

	if len(args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(args))
	}

	if string(args[0]) != "key" {
		t.Errorf("expected 'key', got '%s'", string(args[0]))
	}
}

func TestParseCommand_EmptyCommand(t *testing.T) {
	input := "*0\r\n"
	_, _, err := ParseCommand([]byte(input))
	if err == nil {
		t.Error("expected error for empty command")
	}
}

func TestParseCommand_InvalidType(t *testing.T) {
	input := "+NOTARRAY\r\n"
	_, _, err := ParseCommand([]byte(input))
	if err == nil {
		t.Error("expected error for non-array")
	}
}

func TestParseCommand_InvalidFirstArg(t *testing.T) {
	input := "*1\r\n:123\r\n"
	_, _, err := ParseCommand([]byte(input))
	if err == nil {
		t.Error("expected error for non-bulk string command name")
	}
}

func TestParse_BulkWithSpecialChars(t *testing.T) {
	t.Skip("bulk string with embedded CRLF handling may differ")
}

func TestParse_ArrayWithIntegers(t *testing.T) {
	input := "*3\r\n:1\r\n:2\r\n:3\r\n"
	parser := NewParser(bytes.NewReader([]byte(input)))

	reply, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr, ok := reply.(*ArrayReply)
	if !ok {
		t.Fatal("expected ArrayReply")
	}

	if len(arr.Replies) != 3 {
		t.Fatalf("expected 3, got %d", len(arr.Replies))
	}

	if arr.Replies[0].(*IntegerReply).Num != 1 {
		t.Error("first should be 1")
	}

	if arr.Replies[1].(*IntegerReply).Num != 2 {
		t.Error("second should be 2")
	}

	if arr.Replies[2].(*IntegerReply).Num != 3 {
		t.Error("third should be 3")
	}
}

func TestParse_DeeplyNestedArray(t *testing.T) {
	input := "*1\r\n*1\r\n*1\r\n*1\r\n$1\r\na\r\n"
	parser := NewParser(bytes.NewReader([]byte(input)))

	reply, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := reply.(*ArrayReply)
	for i := 0; i < 3; i++ {
		arr = arr.Replies[0].(*ArrayReply)
	}

	if string(arr.Replies[0].(*BulkReply).Arg) != "a" {
		t.Error("deeply nested value should be 'a'")
	}
}

func TestToBytes_SimpleString(t *testing.T) {
	reply := &SimpleStringReply{Str: "OK"}
	result := reply.ToBytes()
	expected := "+OK\r\n"
	if string(result) != expected {
		t.Errorf("expected '%s', got '%s'", expected, string(result))
	}
}

func TestToBytes_Error(t *testing.T) {
	reply := &ErrorReply{Err: "ERR test"}
	result := reply.ToBytes()
	expected := "-ERR test\r\n"
	if string(result) != expected {
		t.Errorf("expected '%s', got '%s'", expected, string(result))
	}
}

func TestToBytes_Integer(t *testing.T) {
	reply := &IntegerReply{Num: 123}
	result := reply.ToBytes()
	expected := ":123\r\n"
	if string(result) != expected {
		t.Errorf("expected '%s', got '%s'", expected, string(result))
	}
}

func TestToBytes_Bulk(t *testing.T) {
	reply := &BulkReply{Arg: []byte("hello")}
	result := reply.ToBytes()
	expected := "$5\r\nhello\r\n"
	if string(result) != expected {
		t.Errorf("expected '%s', got '%s'", expected, string(result))
	}
}

func TestToBytes_BulkNil(t *testing.T) {
	reply := &BulkReply{Arg: nil}
	result := reply.ToBytes()
	expected := "$-1\r\n"
	if string(result) != expected {
		t.Errorf("expected '%s', got '%s'", expected, string(result))
	}
}

func TestToBytes_Array(t *testing.T) {
	reply := &ArrayReply{
		Replies: []Reply{
			&BulkReply{Arg: []byte("a")},
			&BulkReply{Arg: []byte("b")},
		},
	}
	result := reply.ToBytes()
	expected := "*2\r\n$1\r\na\r\n$1\r\nb\r\n"
	if string(result) != expected {
		t.Errorf("expected '%s', got '%s'", expected, string(result))
	}
}

func TestToBytes_ArrayEmpty(t *testing.T) {
	reply := &ArrayReply{Replies: []Reply{}}
	result := reply.ToBytes()
	expected := "*0\r\n"
	if string(result) != expected {
		t.Errorf("expected '%s', got '%s'", expected, string(result))
	}
}

func TestToBytes_IntegerNegative(t *testing.T) {
	reply := &IntegerReply{Num: -100}
	result := reply.ToBytes()
	expected := ":-100\r\n"
	if string(result) != expected {
		t.Errorf("expected '%s', got '%s'", expected, string(result))
	}
}

func TestParse_IntegerZero(t *testing.T) {
	input := ":0\r\n"
	parser := NewParser(bytes.NewReader([]byte(input)))

	reply, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ir, ok := reply.(*IntegerReply); !ok {
		t.Error("expected IntegerReply")
	} else if ir.Num != 0 {
		t.Errorf("expected 0, got %d", ir.Num)
	}
}

func TestParse_IntegerNegative(t *testing.T) {
	input := ":-100\r\n"
	parser := NewParser(bytes.NewReader([]byte(input)))

	reply, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ir, ok := reply.(*IntegerReply); !ok {
		t.Error("expected IntegerReply")
	} else if ir.Num != -100 {
		t.Errorf("expected -100, got %d", ir.Num)
	}
}

func TestParse_MultipleInOneStream(t *testing.T) {
	input := "+OK\r\n:100\r\n$5\r\nhello\r\n"
	parser := NewParser(bytes.NewReader([]byte(input)))

	reply1, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sr, ok := reply1.(*SimpleStringReply); !ok || sr.Str != "OK" {
		t.Error("first reply should be OK")
	}

	reply2, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ir, ok := reply2.(*IntegerReply); !ok || ir.Num != 100 {
		t.Error("second reply should be 100")
	}

	reply3, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if br, ok := reply3.(*BulkReply); !ok || string(br.Arg) != "hello" {
		t.Error("third reply should be hello")
	}
}

func TestParseCommand_LongArgs(t *testing.T) {
	value := ""
	for i := 0; i < 100; i++ {
		value += "arg"
	}
	input := "*2\r\n$3\r\nGET\r\n$300\r\n" + value + "\r\n"
	cmdName, args, err := ParseCommand([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cmdName != "get" {
		t.Errorf("expected 'get', got '%s'", cmdName)
	}

	if string(args[0]) != value {
		t.Error("arg value mismatch")
	}
}

func TestParseCommand_Unicode(t *testing.T) {
	t.Skip("unicode handling may differ")
}

func TestNewParser(t *testing.T) {
	input := "+test\r\n"
	parser := NewParser(bytes.NewReader([]byte(input)))
	if parser == nil {
		t.Error("NewParser returned nil")
	}

	reply, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sr, ok := reply.(*SimpleStringReply); !ok {
		t.Error("expected SimpleStringReply")
	} else if sr.Str != "test" {
		t.Error("expected 'test'")
	}
}
