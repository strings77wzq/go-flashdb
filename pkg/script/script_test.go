package script

import (
	"testing"

	"goflashdb/pkg/resp"
)

func TestLuaEngine_Eval(t *testing.T) {
	engine := NewLuaEngine()

	script := "return KEYS[1] .. ' ' .. ARGV[1]"
	reply, err := engine.Eval(script, []string{"hello"}, [][]byte{[]byte("world")})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bulkReply, ok := reply.(*resp.BulkReply)
	if !ok {
		t.Fatalf("expected BulkReply, got %T", reply)
	}

	if string(bulkReply.Arg) != "hello world" {
		t.Errorf("expected 'hello world', got %s", string(bulkReply.Arg))
	}
}

func TestLuaEngine_EvalSHA(t *testing.T) {
	engine := NewLuaEngine()

	script := "return 42"
	sha1, err := engine.LoadScript(script)
	if err != nil {
		t.Fatalf("unexpected error loading script: %v", err)
	}

	reply, err := engine.EvalSHA(sha1, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	intReply, ok := reply.(*resp.IntegerReply)
	if !ok {
		t.Fatalf("expected IntegerReply, got %T", reply)
	}

	if intReply.Num != 42 {
		t.Errorf("expected 42, got %d", intReply.Num)
	}
}

func TestLuaEngine_EvalSHA_NotFound(t *testing.T) {
	engine := NewLuaEngine()

	reply, err := engine.EvalSHA("nonexistent", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !isErrorReply(reply) {
		t.Error("expected error reply for nonexistent script")
	}
}

func TestLuaEngine_LoadScript(t *testing.T) {
	engine := NewLuaEngine()

	script := "return 1 + 1"
	sha1, err := engine.LoadScript(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sha1) != 40 {
		t.Errorf("expected sha1 hash of length 40, got %d", len(sha1))
	}
}

func TestLuaEngine_LoadScript_SyntaxError(t *testing.T) {
	engine := NewLuaEngine()

	script := "invalid lua syntax here !!!"
	_, err := engine.LoadScript(script)
	if err == nil {
		t.Error("expected error for invalid syntax")
	}
}

func TestLuaEngine_Exists(t *testing.T) {
	engine := NewLuaEngine()

	script := "return 1"
	sha1, _ := engine.LoadScript(script)

	results := engine.Exists([]string{sha1, "nonexistent"})

	if results[0] != 1 {
		t.Error("expected existing script to return 1")
	}
	if results[1] != 0 {
		t.Error("expected nonexistent script to return 0")
	}
}

func TestLuaEngine_Flush(t *testing.T) {
	engine := NewLuaEngine()

	script := "return 1"
	sha1, _ := engine.LoadScript(script)

	engine.Flush()

	results := engine.Exists([]string{sha1})
	if results[0] != 0 {
		t.Error("expected script to be flushed")
	}
}

func TestLuaEngine_TableReturn(t *testing.T) {
	engine := NewLuaEngine()

	script := "return {1, 2, 3}"
	reply, err := engine.Eval(script, nil, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arrReply, ok := reply.(*resp.ArrayReply)
	if !ok {
		t.Fatalf("expected ArrayReply, got %T", reply)
	}

	if len(arrReply.Replies) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arrReply.Replies))
	}
}

func TestLuaEngine_NilReturn(t *testing.T) {
	engine := NewLuaEngine()

	script := "return nil"
	reply, err := engine.Eval(script, nil, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bulkReply, ok := reply.(*resp.BulkReply)
	if !ok {
		t.Fatalf("expected BulkReply, got %T", reply)
	}

	if bulkReply.Arg != nil {
		t.Errorf("expected nil bulk reply, got %s", string(bulkReply.Arg))
	}
}

func TestLuaEngine_BooleanReturn(t *testing.T) {
	engine := NewLuaEngine()

	script := "return true"
	reply, err := engine.Eval(script, nil, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	intReply, ok := reply.(*resp.IntegerReply)
	if !ok {
		t.Fatalf("expected IntegerReply, got %T", reply)
	}

	if intReply.Num != 1 {
		t.Errorf("expected 1 for true, got %d", intReply.Num)
	}
}

func TestLuaEngine_Sandbox(t *testing.T) {
	engine := NewLuaEngine()

	script := `return io.open("/etc/passwd")`
	reply, err := engine.Eval(script, nil, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !isErrorReply(reply) {
		t.Error("expected error reply for forbidden io access")
	}
}

func TestSha1Hash(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"return 1", "e0e057d3576a7d5b7c8d0e1f2a3b4c5d6e7f8a9b"},
	}

	for _, tt := range tests {
		result := sha1Hash(tt.input)
		if len(result) != 40 {
			t.Errorf("sha1Hash(%q) returned length %d, expected 40", tt.input, len(result))
		}
	}
}

func isErrorReply(reply resp.Reply) bool {
	_, ok := reply.(*resp.ErrorReply)
	return ok
}
