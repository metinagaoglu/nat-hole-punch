package router

import (
	"encoding/json"
	"log/slog"

	. "udp-hole-punch/pkg/models"
)

type Router struct {
	methods map[string]HandlerFunc
}

func NewRouter() *Router {
	return &Router{
		methods: make(map[string]HandlerFunc),
	}
}

func (r *Router) AddRoute(event string, handler HandlerFunc) *Router {
	slog.Debug("Route registered", "event", event)
	r.methods[event] = handler
	return r
}

func (r *Router) HandleEvent(client *Client, bytesRead []byte) {
	slog.Debug("Handling event", "from", client.GetRemoteAddr())
	var request Request

	err := json.Unmarshal(bytesRead, &request)
	if err != nil {
		if client.GetConn() != nil && client.GetRemoteAddr() != nil {
			_, writeErr := client.GetConn().WriteToUDP([]byte("Invalid request"), client.GetRemoteAddr())
			if writeErr != nil {
				slog.Error("Failed to send error response", "error", writeErr)
			}
		}
		return
	}

	slog.Debug("Processing event", "event", request.Event)
	if _, ok := r.methods[request.Event]; !ok {
		slog.Warn("Unknown event", "event", request.Event)
		return
	}
	err = r.methods[request.Event](client, request.Payload)
	if err != nil {
		slog.Error("Handler error", "event", request.Event, "error", err)
	}
}
