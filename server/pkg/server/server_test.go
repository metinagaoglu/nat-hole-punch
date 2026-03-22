package server

import (
	"fmt"
	"testing"

	"udp-hole-punch/pkg/config"
	"udp-hole-punch/pkg/handlers"
	m "udp-hole-punch/pkg/models"
	"udp-hole-punch/pkg/repositories/adapters"
	r "udp-hole-punch/pkg/router"
)

func newTestServer() *UDPServer {
	cfg := config.DefaultConfig()
	cfg.ServerPort = 0 // Let OS assign a free port
	repo := adapters.NewInMemoryRepository()
	ctx := handlers.NewHandlerContext(repo, 60)
	return NewUDPServer(cfg, ctx)
}

func TestNewUDPServer(t *testing.T) {
	server := newTestServer()

	if server == nil {
		t.Error("NewUDPServer() returned nil")
	}

	if server.conn != nil {
		t.Error("NewUDPServer() returned server with non-nil conn")
	}

	if server.clients == nil {
		t.Error("NewUDPServer() returned server with nil clients map")
	}

	if server.router != nil {
		t.Error("NewUDPServer() returned server with non-nil router")
	}

	if server.config == nil {
		t.Error("NewUDPServer() returned server with nil config")
	}

	cfg := config.DefaultConfig()
	if server.bufferSize != cfg.BufferSize {
		t.Errorf("NewUDPServer() bufferSize = %d, want %d", server.bufferSize, cfg.BufferSize)
	}
}

func TestUDPServer_SetRoutes(t *testing.T) {
	server := newTestServer()
	router := r.NewRouter()

	router.AddRoute("test", func(client *m.Client, payload string) error {
		fmt.Println("test")
		return nil
	})

	server.SetRoutes(router)

	if server.router == nil {
		t.Error("SetRoutes() did not set router")
	}
}

func TestUDPServer_Bind(t *testing.T) {
	server := newTestServer()

	boundServer, err := server.Bind()
	if err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	if boundServer.conn == nil {
		t.Error("Bind() did not set conn")
	}

	// Cleanup
	boundServer.conn.Close()
}
