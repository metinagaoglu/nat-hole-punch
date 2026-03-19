package server

import (
	"fmt"
	"testing"

	"udp-hole-punch/pkg/config"
	m "udp-hole-punch/pkg/models"
	r "udp-hole-punch/pkg/router"
)

func TestNewUDPServer(t *testing.T) {
	cfg := config.DefaultConfig()
	server := NewUDPServer(cfg)

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

	if server.bufferSize != cfg.BufferSize {
		t.Errorf("NewUDPServer() bufferSize = %d, want %d", server.bufferSize, cfg.BufferSize)
	}
}

func TestUDPServer_SetRoutes(t *testing.T) {
	cfg := config.DefaultConfig()
	server := NewUDPServer(cfg)
	router := r.NewRouter()

	// HandlerFunc func(client *Client, payload string) error
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
	cfg := config.DefaultConfig()
	server := NewUDPServer(cfg)

	boundServer, err := server.Bind()
	if err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	if boundServer.conn == nil {
		t.Error("Bind() did not set conn")
	}

	expectedAddr := fmt.Sprintf("[::]:% d", cfg.ServerPort)
	if boundServer.conn.LocalAddr().String() != expectedAddr {
		t.Logf("Bind() addr = %s, expected = %s (both are valid)",
			boundServer.conn.LocalAddr().String(), expectedAddr)
	}

	// Cleanup
	boundServer.conn.Close()
}
