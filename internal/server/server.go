package server

import (
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/tectix/hpcs/internal/config"
)

type Server struct {
	cfg      *config.Config
	logger   *zap.Logger
	listener net.Listener
	shutdown chan struct{}
	wg       sync.WaitGroup
}

func New(cfg *config.Config, logger *zap.Logger) *Server {
	return &Server{
		cfg:      cfg,
		logger:   logger,
		shutdown: make(chan struct{}),
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	
	s.listener = listener
	s.logger.Info("Server starting", zap.String("address", addr))
	
	s.wg.Add(1)
	go s.acceptConnections()
	
	<-s.shutdown
	
	s.logger.Info("Server shutting down")
	s.listener.Close()
	s.wg.Wait()
	
	return nil
}

func (s *Server) Stop() {
	close(s.shutdown)
}

func (s *Server) acceptConnections() {
	defer s.wg.Done()
	
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.shutdown:
				return
			default:
				s.logger.Error("Failed to accept connection", zap.Error(err))
				continue
			}
		}
		
		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()
	
	if s.cfg.Server.ReadTimeout > 0 {
		conn.SetReadDeadline(time.Now().Add(s.cfg.Server.ReadTimeout))
	}
	if s.cfg.Server.WriteTimeout > 0 {
		conn.SetWriteDeadline(time.Now().Add(s.cfg.Server.WriteTimeout))
	}
	
	s.logger.Debug("New connection", zap.String("remote", conn.RemoteAddr().String()))
	
	conn.Write([]byte("+OK HPCS Server Ready\r\n"))
}