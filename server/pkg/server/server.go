package server

import (
	"context"
	"fmt"
	"log"
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

	log.Printf("Listening on %s", u.conn.LocalAddr().String())
	for {
		select {
		case <-u.ctx.Done():
			return nil
		default:
		}

		buffer := make([]byte, u.bufferSize)
		bytesRead, remoteAddr, err := u.conn.ReadFromUDP(buffer)
		if err != nil {
			// Check if shutdown was requested
			select {
			case <-u.ctx.Done():
				return nil
			default:
				log.Printf("Error reading from UDP: %v", err)
				continue
			}
		}

		log.Printf("Received %s from %s", string(buffer[0:bytesRead]), remoteAddr)
		client := NewClient().SetRemoteAddr(remoteAddr).SetCreateAt().SetConn(u.conn)

		u.mu.Lock()
		u.clients[remoteAddr.String()] = client
		u.mu.Unlock()

		u.router.HandleEvent(client, buffer[0:bytesRead])
	}
}

// Shutdown gracefully stops the server
func (u *UDPServer) Shutdown() {
	log.Println("Shutting down server...")
	u.cancel()

	if u.conn != nil {
		u.conn.Close()
	}

	log.Println("Server stopped")
}
