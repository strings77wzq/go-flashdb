package pool

import (
	"sync"
	"testing"
	"time"
)

type mockConn struct {
	closed   bool
	lastUsed time.Time
	mu       sync.Mutex
}

func (c *mockConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	return nil
}

func (c *mockConn) IsClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}

func (c *mockConn) LastUsed() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lastUsed
}

func (c *mockConn) SetLastUsed(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastUsed = t
}

func TestNewConnPool(t *testing.T) {
	config := PoolConfig{
		MaxOpen:     10,
		MaxIdle:     5,
		MaxLifetime: time.Minute,
		MaxIdleTime: 30 * time.Second,
	}

	connFunc := func() (Conn, error) {
		return &mockConn{lastUsed: time.Now()}, nil
	}

	pool := NewConnPool(config, connFunc)

	if pool == nil {
		t.Fatal("expected non-nil pool")
	}

	stats := pool.Stats()
	if stats.MaxOpen != 10 {
		t.Errorf("expected MaxOpen 10, got %d", stats.MaxOpen)
	}
}

func TestConnPool_Get(t *testing.T) {
	config := PoolConfig{
		MaxOpen:     10,
		MaxIdle:     5,
		MaxLifetime: time.Minute,
		MaxIdleTime: time.Minute,
	}

	connFunc := func() (Conn, error) {
		return &mockConn{lastUsed: time.Now()}, nil
	}

	pool := NewConnPool(config, connFunc)

	conn, err := pool.Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conn == nil {
		t.Fatal("expected non-nil connection")
	}

	if conn.IsClosed() {
		t.Error("expected connection to be open")
	}
}

func TestConnPool_Put(t *testing.T) {
	config := PoolConfig{
		MaxOpen:     10,
		MaxIdle:     5,
		MaxLifetime: time.Minute,
		MaxIdleTime: time.Minute,
	}

	connFunc := func() (Conn, error) {
		return &mockConn{lastUsed: time.Now()}, nil
	}

	pool := NewConnPool(config, connFunc)

	conn, _ := pool.Get()
	err := pool.Put(conn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stats := pool.Stats()
	if stats.Idle != 1 {
		t.Errorf("expected 1 idle connection, got %d", stats.Idle)
	}
}

func TestConnPool_Reuse(t *testing.T) {
	config := PoolConfig{
		MaxOpen:     10,
		MaxIdle:     5,
		MaxLifetime: time.Minute,
		MaxIdleTime: time.Minute,
	}

	connFunc := func() (Conn, error) {
		return &mockConn{lastUsed: time.Now()}, nil
	}

	pool := NewConnPool(config, connFunc)

	conn1, _ := pool.Get()
	pool.Put(conn1)

	conn2, _ := pool.Get()
	if conn1 != conn2 {
		t.Error("expected connection reuse")
	}
}

func TestConnPool_MaxOpen(t *testing.T) {
	config := PoolConfig{
		MaxOpen:     2,
		MaxIdle:     2,
		MaxLifetime: time.Minute,
		MaxIdleTime: time.Minute,
	}

	connFunc := func() (Conn, error) {
		return &mockConn{lastUsed: time.Now()}, nil
	}

	pool := NewConnPool(config, connFunc)

	conn1, _ := pool.Get()
	conn2, _ := pool.Get()

	_, err := pool.Get()
	if err != ErrPoolExhausted {
		t.Errorf("expected ErrPoolExhausted, got %v", err)
	}

	pool.Put(conn1)
	pool.Put(conn2)
}

func TestConnPool_Close(t *testing.T) {
	config := PoolConfig{
		MaxOpen:     10,
		MaxIdle:     5,
		MaxLifetime: time.Minute,
		MaxIdleTime: time.Minute,
	}

	connFunc := func() (Conn, error) {
		return &mockConn{lastUsed: time.Now()}, nil
	}

	pool := NewConnPool(config, connFunc)

	conn, _ := pool.Get()
	pool.Put(conn)

	pool.Close()

	_, err := pool.Get()
	if err != ErrPoolClosed {
		t.Errorf("expected ErrPoolClosed, got %v", err)
	}
}

func TestConnPool_MaxIdleTime(t *testing.T) {
	config := PoolConfig{
		MaxOpen:     10,
		MaxIdle:     5,
		MaxLifetime: time.Minute,
		MaxIdleTime: 100 * time.Millisecond,
	}

	connFunc := func() (Conn, error) {
		return &mockConn{lastUsed: time.Now()}, nil
	}

	pool := NewConnPool(config, connFunc)

	conn, _ := pool.Get()
	pool.Put(conn)

	time.Sleep(150 * time.Millisecond)

	conn2, err := pool.Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conn == conn2 {
		t.Error("expected new connection after idle timeout")
	}
}

func TestConnPool_MaxIdle(t *testing.T) {
	config := PoolConfig{
		MaxOpen:     10,
		MaxIdle:     2,
		MaxLifetime: time.Minute,
		MaxIdleTime: time.Minute,
	}

	connFunc := func() (Conn, error) {
		return &mockConn{lastUsed: time.Now()}, nil
	}

	pool := NewConnPool(config, connFunc)

	conn1, _ := pool.Get()
	conn2, _ := pool.Get()
	conn3, _ := pool.Get()

	pool.Put(conn1)
	pool.Put(conn2)
	pool.Put(conn3)

	stats := pool.Stats()
	if stats.Idle > config.MaxIdle {
		t.Errorf("expected at most %d idle connections, got %d", config.MaxIdle, stats.Idle)
	}
}

func TestConnPool_PruneIdle(t *testing.T) {
	config := PoolConfig{
		MaxOpen:     10,
		MaxIdle:     5,
		MaxLifetime: time.Minute,
		MaxIdleTime: 100 * time.Millisecond,
	}

	connFunc := func() (Conn, error) {
		return &mockConn{lastUsed: time.Now()}, nil
	}

	pool := NewConnPool(config, connFunc)

	conn1, _ := pool.Get()
	conn2, _ := pool.Get()
	pool.Put(conn1)
	pool.Put(conn2)

	time.Sleep(150 * time.Millisecond)

	pruned := pool.PruneIdle()
	if pruned != 2 {
		t.Errorf("expected 2 pruned connections, got %d", pruned)
	}
}

func TestConnPool_Stats(t *testing.T) {
	config := PoolConfig{
		MaxOpen:     10,
		MaxIdle:     5,
		MaxLifetime: time.Minute,
		MaxIdleTime: time.Minute,
	}

	connFunc := func() (Conn, error) {
		return &mockConn{lastUsed: time.Now()}, nil
	}

	pool := NewConnPool(config, connFunc)

	conn1, _ := pool.Get()
	_, _ = pool.Get()
	pool.Put(conn1)

	stats := pool.Stats()
	if stats.Active != 2 {
		t.Errorf("expected 2 active connections, got %d", stats.Active)
	}
	if stats.Idle != 1 {
		t.Errorf("expected 1 idle connection, got %d", stats.Idle)
	}
}
