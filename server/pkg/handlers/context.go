package handlers

import (
	"net"

	"udp-hole-punch/pkg/repositories"
)

// HandlerContext provides dependencies for request handlers
// This replaces global state with explicit dependency injection
type HandlerContext struct {
	repository repositories.IRepository
	conn       *net.UDPConn
}

// NewHandlerContext creates a new handler context with required dependencies
func NewHandlerContext(repo repositories.IRepository, conn *net.UDPConn) *HandlerContext {
	return &HandlerContext{
		repository: repo,
		conn:       conn,
	}
}

// GetRepository returns the repository instance
func (h *HandlerContext) GetRepository() repositories.IRepository {
	return h.repository
}

// GetConnection returns the UDP connection
func (h *HandlerContext) GetConnection() *net.UDPConn {
	return h.conn
}

// SetConnection updates the UDP connection (called after server binds)
func (h *HandlerContext) SetConnection(conn *net.UDPConn) {
	h.conn = conn
}
