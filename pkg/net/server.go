package net

import (
	"bufio"
	"goflashdb/pkg/core"
	"goflashdb/pkg/resp"
	"net"
	"sync"
)

type Server struct {
	addr     string
	listener net.Listener
	running  bool
	wg       sync.WaitGroup
	closeCh  chan struct{}
	db       *core.DB
}

func NewServer(addr string) (*Server, error) {
	return &Server{
		addr:    addr,
		closeCh: make(chan struct{}),
		db:      core.NewDB(0),
	}, nil
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = listener
	s.running = true

	for s.running {
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

	reader := bufio.NewReader(conn)
	parser := resp.NewParser(reader)

	for s.running {
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

		result := s.db.Exec(cmdName, args)
		conn.Write(result.ToBytes())
	}
}

func (s *Server) Close() {
	s.running = false
	close(s.closeCh)
	if s.listener != nil {
		s.listener.Close()
	}
	s.wg.Wait()
}
