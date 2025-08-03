package server

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/tectix/hpcs/internal/cache"
	"github.com/tectix/hpcs/internal/config"
	"github.com/tectix/hpcs/internal/protocol"
)

type Server struct {
	cfg      *config.Config
	logger   *zap.Logger
	cache    *cache.Cache
	handler  *protocol.CommandHandler
	listener net.Listener
	shutdown chan struct{}
	wg       sync.WaitGroup
}

func New(cfg *config.Config, logger *zap.Logger) *Server {
	maxSize := parseMemorySize(cfg.Cache.MaxMemory)
	cacheInstance := cache.New(maxSize)
	handler := protocol.NewCommandHandler(cacheInstance)
	
	return &Server{
		cfg:      cfg,
		logger:   logger,
		cache:    cacheInstance,
		handler:  handler,
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
	
	s.logger.Debug("New connection", zap.String("remote", conn.RemoteAddr().String()))
	
	parser := protocol.NewParser(conn)
	
	for {
		if s.cfg.Server.ReadTimeout > 0 {
			conn.SetReadDeadline(time.Now().Add(s.cfg.Server.ReadTimeout))
		}
		
		value, err := parser.Parse()
		if err != nil {
			if err == io.EOF {
				break
			}
			s.logger.Debug("Parse error", zap.Error(err))
			break
		}
		
		response := s.handler.Execute(value)
		
		if s.cfg.Server.WriteTimeout > 0 {
			conn.SetWriteDeadline(time.Now().Add(s.cfg.Server.WriteTimeout))
		}
		
		_, err = conn.Write(response.Marshal())
		if err != nil {
			s.logger.Debug("Write error", zap.Error(err))
			break
		}
	}
}

func parseMemorySize(sizeStr string) int64 {
	if sizeStr == "" {
		return 1024 * 1024 * 1024
	}
	
	sizeStr = strings.ToUpper(sizeStr)
	
	if strings.HasSuffix(sizeStr, "GB") {
		if val, err := strconv.ParseInt(sizeStr[:len(sizeStr)-2], 10, 64); err == nil {
			return val * 1024 * 1024 * 1024
		}
	}
	
	if strings.HasSuffix(sizeStr, "MB") {
		if val, err := strconv.ParseInt(sizeStr[:len(sizeStr)-2], 10, 64); err == nil {
			return val * 1024 * 1024
		}
	}
	
	if strings.HasSuffix(sizeStr, "KB") {
		if val, err := strconv.ParseInt(sizeStr[:len(sizeStr)-2], 10, 64); err == nil {
			return val * 1024
		}
	}
	
	if val, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
		return val
	}
	
	return 1024 * 1024 * 1024
}