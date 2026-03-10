package pool

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrPoolClosed    = errors.New("pool is closed")
	ErrPoolExhausted = errors.New("pool exhausted")
)

type Conn interface {
	Close() error
	IsClosed() bool
	LastUsed() time.Time
	SetLastUsed(t time.Time)
}

type PoolConfig struct {
	MaxOpen     int
	MaxIdle     int
	MaxLifetime time.Duration
	MaxIdleTime time.Duration
}

type ConnPool struct {
	config   PoolConfig
	mu       sync.RWMutex
	conns    []Conn
	active   int
	closed   bool
	connFunc func() (Conn, error)
}

func NewConnPool(config PoolConfig, connFunc func() (Conn, error)) *ConnPool {
	if config.MaxOpen <= 0 {
		config.MaxOpen = 100
	}
	if config.MaxIdle <= 0 {
		config.MaxIdle = 10
	}
	if config.MaxLifetime <= 0 {
		config.MaxLifetime = 30 * time.Minute
	}
	if config.MaxIdleTime <= 0 {
		config.MaxIdleTime = 5 * time.Minute
	}

	return &ConnPool{
		config:   config,
		conns:    make([]Conn, 0),
		connFunc: connFunc,
	}
}

func (p *ConnPool) Get() (Conn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, ErrPoolClosed
	}

	now := time.Now()

	for len(p.conns) > 0 {
		conn := p.conns[len(p.conns)-1]
		p.conns = p.conns[:len(p.conns)-1]

		if conn.IsClosed() {
			p.active--
			continue
		}

		if now.Sub(conn.LastUsed()) > p.config.MaxIdleTime {
			conn.Close()
			p.active--
			continue
		}

		conn.SetLastUsed(now)
		return conn, nil
	}

	if p.config.MaxOpen > 0 && p.active >= p.config.MaxOpen {
		return nil, ErrPoolExhausted
	}

	conn, err := p.connFunc()
	if err != nil {
		return nil, err
	}

	p.active++
	conn.SetLastUsed(now)
	return conn, nil
}

func (p *ConnPool) Put(conn Conn) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		conn.Close()
		p.active--
		return ErrPoolClosed
	}

	if conn.IsClosed() {
		p.active--
		return nil
	}

	if len(p.conns) >= p.config.MaxIdle {
		conn.Close()
		p.active--
		return nil
	}

	conn.SetLastUsed(time.Now())
	p.conns = append(p.conns, conn)
	return nil
}

func (p *ConnPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	for _, conn := range p.conns {
		conn.Close()
	}
	p.conns = nil
	p.active = 0

	return nil
}

func (p *ConnPool) Stats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return PoolStats{
		Active:  p.active,
		Idle:    len(p.conns),
		MaxOpen: p.config.MaxOpen,
	}
}

type PoolStats struct {
	Active  int
	Idle    int
	MaxOpen int
}

func (p *ConnPool) PruneIdle() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	pruned := 0

	var valid []Conn
	for _, conn := range p.conns {
		if now.Sub(conn.LastUsed()) > p.config.MaxIdleTime || conn.IsClosed() {
			conn.Close()
			p.active--
			pruned++
		} else {
			valid = append(valid, conn)
		}
	}
	p.conns = valid

	return pruned
}
