package resp

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestNewParser(t *testing.T) {
	reader := strings.NewReader("+OK\r\n")
	parser := NewParser(reader)
	if parser == nil {
		t.Error("NewParser should not return nil")
	}
	if parser.reader == nil {
		t.Error("parser.reader should not be nil")
	}
}

func TestParserParseSimpleString(t *testing.T) {
	reader := strings.NewReader("+Hello World\r\n")
	parser := NewParser(reader)
	reply, err := parser.Parse()
	if err != nil {
		t.Errorf("Parse returned error: %v", err)
	}
	ss, ok := reply.(*SimpleStringReply)
	if !ok {
		t.Fatalf("Expected SimpleStringReply, got %T", reply)
	}
	if ss.Str != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", ss.Str)
	}
}

func TestParserParseError(t *testing.T) {
	reader := strings.NewReader("-ERR test error\r\n")
	parser := NewParser(reader)
	reply, err := parser.Parse()
	if err != nil {
		t.Errorf("Parse returned error: %v", err)
	}
	errReply, ok := reply.(*ErrorReply)
	if !ok {
		t.Fatalf("Expected ErrorReply, got %T", reply)
	}
	if errReply.Err != "ERR test error" {
		t.Errorf("Expected 'ERR test error', got '%s'", errReply.Err)
	}
}

func TestParserParseInteger(t *testing.T) {
	reader := strings.NewReader(":1000\r\n")
	parser := NewParser(reader)
	reply, err := parser.Parse()
	if err != nil {
		t.Errorf("Parse returned error: %v", err)
	}
	intReply, ok := reply.(*IntegerReply)
	if !ok {
		t.Fatalf("Expected IntegerReply, got %T", reply)
	}
	if intReply.Num != 1000 {
		t.Errorf("Expected 1000, got %d", intReply.Num)
	}
}

func TestParserParseNegativeInteger(t *testing.T) {
	reader := strings.NewReader(":-50\r\n")
	parser := NewParser(reader)
	reply, err := parser.Parse()
	if err != nil {
		t.Errorf("Parse returned error: %v", err)
	}
	intReply, ok := reply.(*IntegerReply)
	if !ok {
		t.Fatalf("Expected IntegerReply, got %T", reply)
	}
	if intReply.Num != -50 {
		t.Errorf("Expected -50, got %d", intReply.Num)
	}
}

func TestParserParseBulk(t *testing.T) {
	reader := strings.NewReader("$5\r\nhello\r\n")
	parser := NewParser(reader)
	reply, err := parser.Parse()
	if err != nil {
		t.Errorf("Parse returned error: %v", err)
	}
	bulk, ok := reply.(*BulkReply)
	if !ok {
		t.Fatalf("Expected BulkReply, got %T", reply)
	}
	if string(bulk.Arg) != "hello" {
		t.Errorf("Expected 'hello', got '%s'", string(bulk.Arg))
	}
}

func TestParserParseBulkNil(t *testing.T) {
	reader := strings.NewReader("$-1\r\n")
	parser := NewParser(reader)
	reply, err := parser.Parse()
	if err != nil {
		t.Errorf("Parse returned error: %v", err)
	}
	bulk, ok := reply.(*BulkReply)
	if !ok {
		t.Fatalf("Expected BulkReply, got %T", reply)
	}
	if bulk.Arg != nil {
		t.Error("Nil bulk should have nil Arg")
	}
}

func TestParserParseBulkEmpty(t *testing.T) {
	reader := strings.NewReader("$0\r\n\r\n")
	parser := NewParser(reader)
	reply, err := parser.Parse()
	if err != nil {
		t.Errorf("Parse returned error: %v", err)
	}
	bulk, ok := reply.(*BulkReply)
	if !ok {
		t.Fatalf("Expected BulkReply, got %T", reply)
	}
	if len(bulk.Arg) != 0 {
		t.Errorf("Expected empty bulk, got %d bytes", len(bulk.Arg))
	}
}

func TestParserParseArray(t *testing.T) {
	reader := strings.NewReader("*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n")
	parser := NewParser(reader)
	reply, err := parser.Parse()
	if err != nil {
		t.Errorf("Parse returned error: %v", err)
	}
	array, ok := reply.(*ArrayReply)
	if !ok {
		t.Fatalf("Expected ArrayReply, got %T", reply)
	}
	if len(array.Replies) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(array.Replies))
	}
}

func TestParserParseEmptyArray(t *testing.T) {
	reader := strings.NewReader("*0\r\n")
	parser := NewParser(reader)
	reply, err := parser.Parse()
	if err != nil {
		t.Errorf("Parse returned error: %v", err)
	}
	array, ok := reply.(*ArrayReply)
	if !ok {
		t.Fatalf("Expected ArrayReply, got %T", reply)
	}
	if len(array.Replies) != 0 {
		t.Errorf("Expected 0 elements, got %d", len(array.Replies))
	}
}

func TestParserParseNullArray(t *testing.T) {
	reader := strings.NewReader("*-1\r\n")
	parser := NewParser(reader)
	reply, err := parser.Parse()
	if err != nil {
		t.Errorf("Parse returned error: %v", err)
	}
	array, ok := reply.(*ArrayReply)
	if !ok {
		t.Fatalf("Expected ArrayReply, got %T", reply)
	}
	if array.Replies != nil {
		t.Error("Null array should have nil Replies")
	}
}

func TestParserParseInvalidPrefix(t *testing.T) {
	reader := strings.NewReader("!invalid\r\n")
	parser := NewParser(reader)
	_, err := parser.Parse()
	if err == nil {
		t.Error("Expected error for invalid prefix")
	}
}

func TestParserReadLine(t *testing.T) {
	reader := strings.NewReader("test line\r\n")
	parser := NewParser(reader)
	line, err := parser.readLine()
	if err != nil {
		t.Errorf("readLine returned error: %v", err)
	}
	if string(line) != "test line" {
		t.Errorf("Expected 'test line', got '%s'", string(line))
	}
}

func TestParseCommand(t *testing.T) {
	payload := []byte("*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n")
	cmdName, args, err := ParseCommand(payload)
	if err != nil {
		t.Errorf("ParseCommand returned error: %v", err)
	}
	if cmdName != "set" {
		t.Errorf("Expected cmdName 'set', got '%s'", cmdName)
	}
	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}
	if string(args[0]) != "key" {
		t.Errorf("Expected args[0] 'key', got '%s'", string(args[0]))
	}
	if string(args[1]) != "value" {
		t.Errorf("Expected args[1] 'value', got '%s'", string(args[1]))
	}
}

func TestParseCommandEmpty(t *testing.T) {
	payload := []byte("*0\r\n")
	_, _, err := ParseCommand(payload)
	if err == nil {
		t.Error("Expected error for empty command")
	}
}

func TestParseCommandNotArray(t *testing.T) {
	payload := []byte("+OK\r\n")
	_, _, err := ParseCommand(payload)
	if err == nil {
		t.Error("Expected error for non-array payload")
	}
}

func TestParseCommandEmptyArgs(t *testing.T) {
	payload := []byte("*1\r\n$4\r\nPING\r\n")
	cmdName, args, err := ParseCommand(payload)
	if err != nil {
		t.Errorf("ParseCommand returned error: %v", err)
	}
	if cmdName != "ping" {
		t.Errorf("Expected cmdName 'ping', got '%s'", cmdName)
	}
	if len(args) != 0 {
		t.Errorf("Expected 0 args, got %d", len(args))
	}
}

func TestParserReadLineMultipleChunks(t *testing.T) {
	buf := &bytes.Buffer{}
	buf.WriteString("part1")
	buf.WriteString("\r\n")
	parser := NewParser(buf)
	line, err := parser.readLine()
	if err != nil {
		t.Errorf("readLine returned error: %v", err)
	}
	if string(line) != "part1" {
		t.Errorf("Expected 'part1', got '%s'", string(line))
	}
}

func TestParserBulkLargeSize(t *testing.T) {
	data := strings.Repeat("a", 1000)
	payload := "$1000\r\n" + data + "\r\n"
	reader := strings.NewReader(payload)
	parser := NewParser(reader)
	reply, err := parser.Parse()
	if err != nil {
		t.Errorf("Parse returned error: %v", err)
	}
	bulk, ok := reply.(*BulkReply)
	if !ok {
		t.Fatalf("Expected BulkReply, got %T", reply)
	}
	if len(bulk.Arg) != 1000 {
		t.Errorf("Expected 1000 bytes, got %d", len(bulk.Arg))
	}
}

func TestSimpleStringReplyToBytes(t *testing.T) {
	reply := &SimpleStringReply{Str: "OK"}
	result := reply.ToBytes()
	expected := "+OK\r\n"
	if string(result) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(result))
	}
}

func TestErrorReplyToBytes(t *testing.T) {
	reply := &ErrorReply{Err: "ERR test"}
	result := reply.ToBytes()
	expected := "-ERR test\r\n"
	if string(result) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(result))
	}
}

func TestIntegerReplyToBytes(t *testing.T) {
	reply := &IntegerReply{Num: 12345}
	result := reply.ToBytes()
	expected := ":12345\r\n"
	if string(result) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(result))
	}
}

func TestBulkReplyToBytes(t *testing.T) {
	reply := &BulkReply{Arg: []byte("test")}
	result := reply.ToBytes()
	expected := "$4\r\ntest\r\n"
	if string(result) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(result))
	}
}

func TestBulkReplyNilToBytes(t *testing.T) {
	reply := &BulkReply{Arg: nil}
	result := reply.ToBytes()
	expected := "$-1\r\n"
	if string(result) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(result))
	}
}

func TestArrayReplyToBytes(t *testing.T) {
	replies := []Reply{
		&SimpleStringReply{Str: "OK"},
		&IntegerReply{Num: 1},
	}
	reply := &ArrayReply{Replies: replies}
	result := reply.ToBytes()
	expected := "*2\r\n+OK\r\n:1\r\n"
	if string(result) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(result))
	}
}

func TestArrayReplyEmptyToBytes(t *testing.T) {
	reply := &ArrayReply{Replies: []Reply{}}
	result := reply.ToBytes()
	expected := "*0\r\n"
	if string(result) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(result))
	}
}

func TestNewErrorReply(t *testing.T) {
	reply := NewErrorReply("test error")
	if reply.Err != "test error" {
		t.Errorf("Expected 'test error', got '%s'", reply.Err)
	}
}

func TestNewBulkReply(t *testing.T) {
	reply := NewBulkReply([]byte("test"))
	if string(reply.Arg) != "test" {
		t.Errorf("Expected 'test', got '%s'", string(reply.Arg))
	}
}

func TestNewArrayReply(t *testing.T) {
	replies := []Reply{
		&SimpleStringReply{Str: "OK"},
		&BulkReply{Arg: []byte("test")},
	}
	reply := NewArrayReply(replies)
	if len(reply.Replies) != 2 {
		t.Errorf("Expected 2 replies, got %d", len(reply.Replies))
	}
}

func TestOkReply(t *testing.T) {
	result := OkReply.ToBytes()
	expected := "+OK\r\n"
	if string(result) != expected {
		t.Error("OkReply.ToBytes() should return '+OK\\r\\n'")
	}
}

func TestPongReply(t *testing.T) {
	result := PongReply.ToBytes()
	expected := "+PONG\r\n"
	if string(result) != expected {
		t.Error("PongReply.ToBytes() should return '+PONG\\r\\n'")
	}
}

func TestNilBulkReply(t *testing.T) {
	result := NilBulkReply.ToBytes()
	expected := "$-1\r\n"
	if string(result) != expected {
		t.Error("NilBulkReply.ToBytes() should return '$-1\\r\\n'")
	}
}

func TestNullArrayReply(t *testing.T) {
	result := NullArrayReply.ToBytes()
	expected := "*0\r\n"
	if string(result) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(result))
	}
}

func TestParseCommandMixedCase(t *testing.T) {
	payload := []byte("*1\r\n$4\r\nPING\r\n")
	cmdName, _, err := ParseCommand(payload)
	if err != nil {
		t.Errorf("ParseCommand returned error: %v", err)
	}
	if cmdName != "ping" {
		t.Errorf("Expected 'ping', got '%s'", cmdName)
	}
}

func TestParserNestedArray(t *testing.T) {
	reader := strings.NewReader("*2\r\n*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n$3\r\nbaz\r\n")
	parser := NewParser(reader)
	reply, err := parser.Parse()
	if err != nil {
		t.Errorf("Parse returned error: %v", err)
	}
	array, ok := reply.(*ArrayReply)
	if !ok {
		t.Fatalf("Expected ArrayReply, got %T", reply)
	}
	if len(array.Replies) != 2 {
		t.Fatalf("Expected 2 elements, got %d", len(array.Replies))
	}
	nested, ok := array.Replies[0].(*ArrayReply)
	if !ok {
		t.Fatal("First element should be ArrayReply")
	}
	if len(nested.Replies) != 2 {
		t.Fatalf("Expected nested array to have 2 elements, got %d", len(nested.Replies))
	}
}

func TestParserReadLineError(t *testing.T) {
	reader := &errorReader{}
	parser := NewParser(reader)
	_, err := parser.readLine()
	if err == nil {
		t.Error("Expected error from errorReader")
	}
}

type errorReader struct{}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}
