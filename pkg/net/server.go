package net

import (
	"bufio"
	"crypto/tls"
	"goflashdb/pkg/core"
	"goflashdb/pkg/persist"
	"goflashdb/pkg/resp"
	"goflashdb/pkg/security"
	"net"
	"strings"
	"sync"
	"time"
)

type Server struct {
	mu          sync.Mutex
	addr        string
	listener    net.Listener
	running     bool
	wg          sync.WaitGroup
	closeCh     chan struct{}
	readyCh     chan struct{}
	db          *core.DB
	auth        *security.Authenticator
	rateLimiter *security.RateLimiter
	filter      *security.CommandFilter
	persistMgr  *persist.PersistManager
	tlsConfig   *tls.Config
}

type ServerOption func(*Server)

func WithAuth(password string) ServerOption {
	return func(s *Server) {
		s.auth = security.NewAuthenticator(password)
	}
}

func WithRateLimit(limit int, window time.Duration) ServerOption {
	return func(s *Server) {
		s.rateLimiter = security.NewRateLimiter(limit, window)
	}
}

func WithFilter(renamedCommands map[string]string) ServerOption {
	return func(s *Server) {
		s.filter = security.NewCommandFilter(renamedCommands)
	}
}

func WithPersist(aofFile, rdbFile string, aofEnabled bool, saveInterval time.Duration) ServerOption {
	return func(s *Server) {
		pm, err := persist.NewPersistManager(aofFile, rdbFile, aofEnabled, saveInterval)
		if err == nil {
			s.persistMgr = pm
		}
	}
}

func WithTLS(certFile, keyFile string) ServerOption {
	return func(s *Server) {
		if certFile != "" && keyFile != "" {
			s.tlsConfig = &tls.Config{}
			if cert, err := tls.LoadX509KeyPair(certFile, keyFile); err == nil {
				s.tlsConfig.Certificates = []tls.Certificate{cert}
			}
		}
	}
}

func NewServer(addr string, opts ...ServerOption) (*Server, error) {
	s := &Server{
		addr:    addr,
		closeCh: make(chan struct{}),
		readyCh: make(chan struct{}),
		db:      core.NewDB(0),
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.persistMgr != nil {
		s.db = core.NewDBWithPersist(0, s.persistMgr)
		if err := s.db.LoadFromPersist(); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *Server) Start() error {
	var listener net.Listener
	var err error

	if s.tlsConfig != nil {
		listener, err = tls.Listen("tcp", s.addr, s.tlsConfig)
	} else {
		listener, err = net.Listen("tcp", s.addr)
	}
	if err != nil {
		return err
	}
	s.listener = listener
	s.mu.Lock()
	s.running = true
	s.mu.Unlock()
	close(s.readyCh)

	for {
		s.mu.Lock()
		running := s.running
		s.mu.Unlock()
		if !running {
			break
		}
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.closeCh:
				return nil
			default:
				return err
			}
		}
		s.wg.Add(1)
		go s.handleConn(conn)
	}
	return nil
}

func (s *Server) handleConn(conn net.Conn) {
	defer func() {
		conn.Close()
		s.wg.Done()
	}()

	clientAddr := conn.RemoteAddr().String()
	clientID := core.AddClient(clientAddr, 0)
	defer core.RemoveClient(clientID)

	reader := bufio.NewReader(conn)
	parser := resp.NewParser(reader)
	authenticated := s.auth == nil || !s.auth.IsEnabled()

	for {
		s.mu.Lock()
		running := s.running
		s.mu.Unlock()
		if !running {
			break
		}
		startTime := time.Now()
		reply, err := parser.Parse()
		if err != nil {
			return
		}
		arrayReply, ok := reply.(*resp.ArrayReply)
		if !ok || len(arrayReply.Replies) == 0 {
			conn.Write(resp.NewErrorReply("ERR invalid command").ToBytes())
			continue
		}

		cmdName := ""
		args := make([][]byte, 0, len(arrayReply.Replies)-1)
		for i, r := range arrayReply.Replies {
			bulkReply, ok := r.(*resp.BulkReply)
			if !ok {
				conn.Write(resp.NewErrorReply("ERR invalid command format").ToBytes())
				break
			}
			if i == 0 {
				cmdName = string(bulkReply.Arg)
			} else {
				args = append(args, bulkReply.Arg)
			}
		}
		if cmdName == "" {
			continue
		}

		cmdLower := strings.ToLower(cmdName)
		core.UpdateClient(clientID, cmdLower)

		if s.rateLimiter != nil && !s.rateLimiter.Allow(clientAddr) {
			conn.Write(resp.NewErrorReply("ERR rate limit exceeded").ToBytes())
			continue
		}

		if !authenticated {
			if cmdLower == "auth" {
				if len(args) != 1 {
					conn.Write(resp.NewErrorReply("ERR wrong number of arguments for 'auth' command").ToBytes())
					continue
				}
				if s.auth.Authenticate(string(args[0])) {
					authenticated = true
					conn.Write(resp.OkReply.ToBytes())
				} else {
					conn.Write(resp.NewErrorReply("ERR invalid password").ToBytes())
				}
				continue
			}
			conn.Write(resp.NewErrorReply("NOAUTH Authentication required").ToBytes())
			continue
		}

		if s.filter != nil {
			if s.filter.IsBlocked(cmdLower) {
				conn.Write(resp.NewErrorReply("ERR command '" + cmdLower + "' is blocked").ToBytes())
				continue
			}
			cmdLower = s.filter.Rename(cmdLower)
		}

		result := s.db.Exec(cmdLower, args)

		duration := time.Since(startTime)
		argsStr := make([]string, len(args))
		for i, arg := range args {
			argsStr[i] = string(arg)
		}
		core.AddSlowLog(cmdName, argsStr, duration)

		if result == nil {
			go s.handleShutdown()
			return
		}

		conn.Write(result.ToBytes())
	}
}

func (s *Server) handleShutdown() {
	s.Close()
}

func (s *Server) Close() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()
	close(s.closeCh)
	if s.listener != nil {
		s.listener.Close()
	}
	if s.persistMgr != nil {
		s.persistMgr.Close()
	}
	s.wg.Wait()
}

func (s *Server) GetDB() *core.DB {
	return s.db
}
