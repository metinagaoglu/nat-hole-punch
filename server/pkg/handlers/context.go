package handlers

import (
	"net"

	"udp-hole-punch/pkg/models"
	"udp-hole-punch/pkg/repositories"
)

// HandlerContext provides dependencies for request handlers
// This replaces global state with explicit dependency injection
type HandlerContext struct {
	repository repositories.IRepository
	conn       *net.UDPConn
	clientTTL  int32
}

// NewHandlerContext creates a new handler context with required dependencies
func NewHandlerContext(repo repositories.IRepository, clientTTL int) *HandlerContext {
	return &HandlerContext{
		repository: repo,
		clientTTL:  int32(clientTTL),
	}
}

// SetConnection updates the UDP connection (called after server binds)
func (ctx *HandlerContext) SetConnection(conn *net.UDPConn) {
	ctx.conn = conn
}

// RegisterHandler returns a HandlerFunc that handles client registration
func (ctx *HandlerContext) RegisterHandler() models.HandlerFunc {
	return func(client *models.Client, payload string) error {
		return register(ctx, client, payload)
	}
}

// LogoutHandler returns a HandlerFunc that handles client logout
func (ctx *HandlerContext) LogoutHandler() models.HandlerFunc {
	return func(client *models.Client, payload string) error {
		return logout(ctx, client, payload)
	}
}
