package router

import (
	"encoding/json"
	"log/slog"
	"net"

	"github.com/metnagaoglu/holepunch/signal/internal/handler"
)

type request struct {
	Event   string `json:"event"`
	Payload string `json:"payload"`
}

// Router dispatches incoming UDP events to registered handlers
type Router struct {
	methods map[string]handler.HandlerFunc
}

// New creates a new Router
func New() *Router {
	return &Router{
		methods: make(map[string]handler.HandlerFunc),
	}
}

// AddRoute registers a handler for an event name
func (r *Router) AddRoute(event string, h handler.HandlerFunc) *Router {
	slog.Debug("Route registered", "event", event)
	r.methods[event] = h
	return r
}

// HandleEvent parses and routes an incoming UDP message
func (r *Router) HandleEvent(addr *net.UDPAddr, conn *net.UDPConn, data []byte) {
	slog.Debug("Handling event", "from", addr)

	var req request
	if err := json.Unmarshal(data, &req); err != nil {
		conn.WriteToUDP([]byte("Invalid request"), addr)
		return
	}

	slog.Debug("Processing event", "event", req.Event)

	h, ok := r.methods[req.Event]
	if !ok {
		slog.Warn("Unknown event", "event", req.Event)
		return
	}

	if err := h(addr, conn, req.Payload); err != nil {
		slog.Error("Handler error", "event", req.Event, "error", err)
	}
}
