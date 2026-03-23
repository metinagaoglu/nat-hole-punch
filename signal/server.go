// Package signal provides an embeddable UDP signaling server for NAT hole punching.
//
// Usage:
//
//	cfg := signal.DefaultConfig()
//	cfg.Port = 3986
//	srv, _ := signal.NewServer(cfg)
//	go srv.ListenAndServe()
//	// ...
//	srv.Shutdown()
package signal

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"

	"github.com/metnagaoglu/holepunch/signal/internal/handler"
	"github.com/metnagaoglu/holepunch/signal/internal/repository"
	"github.com/metnagaoglu/holepunch/signal/internal/repository/adapters"
	"github.com/metnagaoglu/holepunch/signal/internal/router"
)

// Server is a UDP signaling server for NAT hole punching
type Server struct {
	mu         sync.RWMutex
	conn       *net.UDPConn
	router     *router.Router
	handlerCtx *handler.Context
	config     Config
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewServer creates a new signal server with the given configuration.
// Call ListenAndServe to start accepting connections.
func NewServer(cfg Config) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	repo := createRepository(cfg)
	handlerCtx := handler.NewContext(repo, cfg.ClientTTL)

	r := router.New()
	r.AddRoute("register", handlerCtx.RegisterHandler())
	r.AddRoute("logout", handlerCtx.LogoutHandler())

	ctx, cancel := context.WithCancel(context.Background())

	return &Server{
		router:     r,
		handlerCtx: handlerCtx,
		config:     cfg,
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// ListenAndServe binds to the configured address and starts serving.
// This call blocks until Shutdown is called or an error occurs.
func (s *Server) ListenAndServe() error {
	addr := net.UDPAddr{
		Port: s.config.Port,
		IP:   net.ParseIP(s.config.Host),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return fmt.Errorf("failed to bind: %w", err)
	}

	s.conn = conn
	s.handlerCtx.SetConnection(conn)

	slog.Info("Signal server listening", "addr", conn.LocalAddr().String())

	for {
		select {
		case <-s.ctx.Done():
			return nil
		default:
		}

		buffer := make([]byte, s.config.BufferSize)
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			select {
			case <-s.ctx.Done():
				return nil
			default:
				slog.Error("Error reading from UDP", "error", err)
				continue
			}
		}

		slog.Debug("Received message", "from", remoteAddr, "size", n)

		// Refresh TTL on every incoming message
		s.handlerCtx.RefreshClient(remoteAddr.String())

		s.router.HandleEvent(remoteAddr, conn, buffer[:n])
	}
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown() {
	slog.Info("Shutting down signal server")
	s.cancel()

	if s.conn != nil {
		s.conn.Close()
	}

	if err := s.handlerCtx.Close(); err != nil {
		slog.Error("Failed to close handler context", "error", err)
	}

	slog.Info("Signal server stopped")
}

// Addr returns the server's local address, or nil if not yet bound
func (s *Server) Addr() net.Addr {
	if s.conn == nil {
		return nil
	}
	return s.conn.LocalAddr()
}

func createRepository(cfg Config) repository.Repository {
	switch cfg.RepositoryType {
	case "redis":
		redisConfig := adapters.RedisConfig{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		}
		repo, err := adapters.NewRedisRepository(redisConfig)
		if err != nil {
			slog.Warn("Failed to initialize Redis, falling back to in-memory", "error", err)
			return adapters.NewMemoryRepository()
		}
		slog.Info("Using Redis repository", "addr", redisConfig.Addr)
		return repo

	default:
		slog.Info("Using in-memory repository")
		return adapters.NewMemoryRepository()
	}
}
