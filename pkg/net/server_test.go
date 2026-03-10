package net

import (
	"bufio"
	"net"
	"testing"
	"time"

	"goflashdb/pkg/resp"
)

func TestWithAuth(t *testing.T) {
	opt := WithAuth("password")
	s := &Server{}
	opt(s)

	if s.auth == nil {
		t.Error("auth should not be nil")
	}
	if !s.auth.IsEnabled() {
		t.Error("auth should be enabled")
	}
}

func TestWithRateLimit(t *testing.T) {
	opt := WithRateLimit(100, time.Minute)
	s := &Server{}
	opt(s)

	if s.rateLimiter == nil {
		t.Error("rateLimiter should not be nil")
	}
}

func TestWithFilter(t *testing.T) {
	renamed := map[string]string{
		"flushall": "hidden",
	}
	opt := WithFilter(renamed)
	s := &Server{}
	opt(s)

	if s.filter == nil {
		t.Error("filter should not be nil")
	}
}

func TestWithPersist(t *testing.T) {
	opt := WithPersist("/tmp/test.aof", "/tmp/test.rdb", false, time.Minute)
	s := &Server{}
	opt(s)

	if s.persistMgr == nil {
		t.Error("persistMgr should not be nil")
	}
}

func TestNewServer(t *testing.T) {
	server, err := NewServer("localhost:0")
	if err != nil {
		t.Errorf("NewServer returned error: %v", err)
	}
	if server == nil {
		t.Error("server should not be nil")
	}
	if server.db == nil {
		t.Error("db should not be nil")
	}
	server.Close()
}

func TestNewServerWithOptions(t *testing.T) {
	server, err := NewServer(
		"localhost:0",
		WithAuth("password"),
		WithRateLimit(100, time.Minute),
	)
	if err != nil {
		t.Errorf("NewServer returned error: %v", err)
	}
	if server.auth == nil {
		t.Error("auth should not be nil")
	}
	if server.rateLimiter == nil {
		t.Error("rateLimiter should not be nil")
	}
	server.Close()
}

func TestServerGetDB(t *testing.T) {
	server, _ := NewServer("localhost:0")
	defer server.Close()

	db := server.GetDB()
	if db == nil {
		t.Error("db should not be nil")
	}

	_ = db
}

func TestWithTLS(t *testing.T) {
	opt := WithTLS("", "")
	s := &Server{}
	opt(s)

	if s.tlsConfig != nil {
		t.Error("tlsConfig should be nil for empty cert/key files")
	}
}

func TestServerStartAndPing(t *testing.T) {
	server, err := NewServer("localhost:0")
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	defer server.Close()

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start()
	}()

	// Wait for server to be ready with timeout
	select {
	case <-server.readyCh:
	case <-time.After(1 * time.Second):
		t.Fatal("server failed to start within timeout")
	}

	addr := server.listener.Addr().String()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	cmd := resp.NewArrayReply([]resp.Reply{
		resp.NewBulkReply([]byte("PING")),
	})
	_, err = conn.Write(cmd.ToBytes())
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	reader := bufio.NewReader(conn)
	parser := resp.NewParser(reader)
	reply, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse reply failed: %v", err)
	}

	if sr, ok := reply.(*resp.SimpleStringReply); !ok || sr.Str != "PONG" {
		t.Errorf("expected PONG, got %v", reply)
	}
}

func TestServerAuthFlow(t *testing.T) {
	server, err := NewServer("localhost:0", WithAuth("mypassword"))
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	defer server.Close()

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start()
	}()

	// Wait for server to be ready with timeout
	select {
	case <-server.readyCh:
	case <-time.After(1 * time.Second):
		t.Fatal("server failed to start within timeout")
	}

	addr := server.listener.Addr().String()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	parser := resp.NewParser(reader)

	cmd := resp.NewArrayReply([]resp.Reply{
		resp.NewBulkReply([]byte("GET")),
		resp.NewBulkReply([]byte("key1")),
	})
	_, err = conn.Write(cmd.ToBytes())
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	reply, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse reply failed: %v", err)
	}
	if er, ok := reply.(*resp.ErrorReply); !ok || er.Err != "NOAUTH Authentication required" {
		t.Errorf("expected NOAUTH error, got %v", reply)
	}

	cmd = resp.NewArrayReply([]resp.Reply{
		resp.NewBulkReply([]byte("AUTH")),
		resp.NewBulkReply([]byte("wrongpassword")),
	})
	_, err = conn.Write(cmd.ToBytes())
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	reply, err = parser.Parse()
	if err != nil {
		t.Fatalf("Parse reply failed: %v", err)
	}
	if er, ok := reply.(*resp.ErrorReply); !ok || er.Err != "ERR invalid password" {
		t.Errorf("expected invalid password error, got %v", reply)
	}

	cmd = resp.NewArrayReply([]resp.Reply{
		resp.NewBulkReply([]byte("AUTH")),
		resp.NewBulkReply([]byte("mypassword")),
	})
	_, err = conn.Write(cmd.ToBytes())
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	reply, err = parser.Parse()
	if err != nil {
		t.Fatalf("Parse reply failed: %v", err)
	}
	if sr, ok := reply.(*resp.SimpleStringReply); !ok || sr.Str != "OK" {
		t.Errorf("expected OK, got %v", reply)
	}

	cmd = resp.NewArrayReply([]resp.Reply{
		resp.NewBulkReply([]byte("SET")),
		resp.NewBulkReply([]byte("key1")),
		resp.NewBulkReply([]byte("value1")),
	})
	_, err = conn.Write(cmd.ToBytes())
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	reply, err = parser.Parse()
	if err != nil {
		t.Fatalf("Parse reply failed: %v", err)
	}
	if sr, ok := reply.(*resp.SimpleStringReply); !ok || sr.Str != "OK" {
		t.Errorf("expected OK, got %v", reply)
	}

	cmd = resp.NewArrayReply([]resp.Reply{
		resp.NewBulkReply([]byte("GET")),
		resp.NewBulkReply([]byte("key1")),
	})
	_, err = conn.Write(cmd.ToBytes())
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	reply, err = parser.Parse()
	if err != nil {
		t.Fatalf("Parse reply failed: %v", err)
	}
	if br, ok := reply.(*resp.BulkReply); !ok || string(br.Arg) != "value1" {
		t.Errorf("expected value1, got %v", reply)
	}
}

func TestServerInvalidCommand(t *testing.T) {
	server, err := NewServer("localhost:0")
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	defer server.Close()

	go func() {
		server.Start()
	}()

	// Wait for server to be ready with timeout
	select {
	case <-server.readyCh:
	case <-time.After(1 * time.Second):
		t.Fatal("server failed to start within timeout")
	}

	addr := server.listener.Addr().String()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	parser := resp.NewParser(reader)

	cmd := resp.NewArrayReply([]resp.Reply{
		resp.NewBulkReply([]byte("NOTEXISTSCMD")),
	})
	_, err = conn.Write(cmd.ToBytes())
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	reply, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse reply failed: %v", err)
	}
	if _, ok := reply.(*resp.ErrorReply); !ok {
		t.Errorf("expected error reply for invalid command, got %v", reply)
	}
}
