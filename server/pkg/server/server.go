package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"

	"udp-hole-punch/pkg/config"
	"udp-hole-punch/pkg/handlers"
	. "udp-hole-punch/pkg/models"
	. "udp-hole-punch/pkg/router"
)

type UDPServer struct {
	mu         sync.RWMutex
	conn       *net.UDPConn
	clients    map[string]*Client
	router     *Router
	handlerCtx *handlers.HandlerContext
	config     *config.Config
	bufferSize int
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewUDPServer(cfg *config.Config, handlerCtx *handlers.HandlerContext) *UDPServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &UDPServer{
		conn:       nil,
		clients:    make(map[string]*Client),
		handlerCtx: handlerCtx,
		config:     cfg,
		bufferSize: cfg.BufferSize,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (u *UDPServer) SetRoutes(r *Router) *UDPServer {
	u.router = r
	return u
}

func (u *UDPServer) Bind() (*UDPServer, error) {
	addr := net.UDPAddr{
		Port: u.config.ServerPort,
		IP:   net.ParseIP(u.config.ServerHost),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return nil, err
	}

	u.conn = conn
	u.handlerCtx.SetConnection(conn)

	return u, nil
}

func (u *UDPServer) Listen() error {
	if u.conn == nil {
		return fmt.Errorf("server not bound to any address")
	}

	slog.Info("Listening", "addr", u.conn.LocalAddr().String())
	for {
		select {
		case <-u.ctx.Done():
			return nil
		default:
		}

		buffer := make([]byte, u.bufferSize)
		bytesRead, remoteAddr, err := u.conn.ReadFromUDP(buffer)
		if err != nil {
			select {
			case <-u.ctx.Done():
				return nil
			default:
				slog.Error("Error reading from UDP", "error", err)
				continue
			}
		}

		slog.Debug("Received message", "from", remoteAddr, "data", string(buffer[0:bytesRead]))
		client := NewClient().SetRemoteAddr(remoteAddr).SetCreateAt().SetConn(u.conn)

		u.mu.Lock()
		u.clients[remoteAddr.String()] = client
		u.mu.Unlock()

		// Refresh TTL on every incoming message - keeps active clients alive
		u.handlerCtx.RefreshClient(remoteAddr.String())

		u.router.HandleEvent(client, buffer[0:bytesRead])
	}
}

// Shutdown gracefully stops the server
func (u *UDPServer) Shutdown() {
	slog.Info("Shutting down server")
	u.cancel()

	if u.conn != nil {
		u.conn.Close()
	}

	if err := u.handlerCtx.Close(); err != nil {
		slog.Error("Failed to close handler context", "error", err)
	}

	slog.Info("Server stopped")
}
