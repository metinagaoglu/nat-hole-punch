package handler

import (
	"encoding/json"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/metnagaoglu/holepunch/signal/internal/repository"
)

// Context provides dependencies for request handlers via DI
type Context struct {
	repo      repository.Repository
	conn      *net.UDPConn
	clientTTL int32
}

// NewContext creates a handler context with the given dependencies
func NewContext(repo repository.Repository, clientTTL int) *Context {
	return &Context{
		repo:      repo,
		clientTTL: int32(clientTTL),
	}
}

// SetConnection sets the server UDP connection (called after bind)
func (ctx *Context) SetConnection(conn *net.UDPConn) {
	ctx.conn = conn
}

// RefreshClient extends a client's TTL
func (ctx *Context) RefreshClient(addr string) {
	ctx.repo.RefreshClient(addr, ctx.clientTTL)
}

// Close releases handler context resources
func (ctx *Context) Close() error {
	return ctx.repo.Close()
}

// HandlerFunc is the signature for event handlers
type HandlerFunc func(addr *net.UDPAddr, conn *net.UDPConn, payload string) error

// RegisterHandler returns a handler for client registration
func (ctx *Context) RegisterHandler() HandlerFunc {
	return func(addr *net.UDPAddr, conn *net.UDPConn, payload string) error {
		return ctx.handleRegister(addr, conn, payload)
	}
}

// LogoutHandler returns a handler for client logout
func (ctx *Context) LogoutHandler() HandlerFunc {
	return func(addr *net.UDPAddr, conn *net.UDPConn, payload string) error {
		return ctx.handleLogout(addr, conn, payload)
	}
}

type registerRequest struct {
	LocalIP string `json:"local_ip"`
	Key     string `json:"key"`
}

func (ctx *Context) handleRegister(addr *net.UDPAddr, conn *net.UDPConn, payload string) error {
	if err := validatePayload(payload); err != nil {
		return err
	}

	var req registerRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		return err
	}

	if err := validateKey(req.Key); err != nil {
		return err
	}

	slog.Info("Registering client", "addr", addr, "room", req.Key)

	client := &repository.Client{
		Addr:      addr,
		Conn:      conn,
		CreatedAt: time.Now().Unix(),
	}
	ctx.repo.AddClient(req.Key, client, ctx.clientTTL)

	return ctx.broadcastPeers(req.Key)
}

func (ctx *Context) handleLogout(addr *net.UDPAddr, _ *net.UDPConn, payload string) error {
	if err := validatePayload(payload); err != nil {
		return err
	}

	var req registerRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		return err
	}

	if err := validateKey(req.Key); err != nil {
		return err
	}

	slog.Info("Client logout", "addr", addr, "room", req.Key)

	if err := ctx.repo.RemoveClient(req.Key, addr); err != nil {
		return err
	}

	return ctx.broadcastPeers(req.Key)
}

func (ctx *Context) broadcastPeers(key string) error {
	if ctx.conn == nil {
		slog.Warn("Server connection not set, cannot broadcast")
		return nil
	}

	clients, err := ctx.repo.GetClientsByKey(key)
	if err != nil {
		slog.Error("Failed to get clients", "room", key, "error", err)
		return err
	}

	slog.Debug("Broadcasting peer list", "room", key, "client_count", len(clients))

	var addrs strings.Builder
	for _, c := range clients {
		if c.Addr != nil {
			addrs.WriteString(c.Addr.String() + ",")
		}
	}

	message := []byte(strings.TrimRight(addrs.String(), ","))

	for _, c := range clients {
		if c.Addr == nil {
			continue
		}
		if _, err := ctx.conn.WriteToUDP(message, c.Addr); err != nil {
			slog.Error("Failed to send to client", "addr", c.Addr, "error", err)
			continue
		}
	}
	return nil
}
