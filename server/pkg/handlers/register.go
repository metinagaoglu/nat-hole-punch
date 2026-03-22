package handlers

import (
	"encoding/json"
	"log/slog"
	"strings"

	. "udp-hole-punch/pkg/models"
)

type RegisterRequest struct {
	LocalIp string `json:"local_ip"`
	Key     string `json:"key"`
}

func register(ctx *HandlerContext, client *Client, payload string) error {
	if err := validatePayload(payload); err != nil {
		return err
	}

	var registerRequest RegisterRequest
	err := json.Unmarshal([]byte(payload), &registerRequest)
	if err != nil {
		return err
	}

	if err := validateKey(registerRequest.Key); err != nil {
		return err
	}

	slog.Info("Registering client", "addr", client.GetRemoteAddr(), "room", registerRequest.Key)
	ctx.repository.AddClient(registerRequest.Key, client, ctx.clientTTL)

	return broadcastPeers(ctx, registerRequest.Key)
}

func broadcastPeers(ctx *HandlerContext, key string) error {
	if ctx.conn == nil {
		slog.Warn("Server connection not set, cannot broadcast")
		return nil
	}

	clients, err := ctx.repository.GetClientsByKey(key)
	if err != nil {
		slog.Error("Failed to get clients", "room", key, "error", err)
		return err
	}

	slog.Debug("Broadcasting peer list", "room", key, "client_count", len(clients))

	var ipAddresses strings.Builder
	for _, client := range clients {
		if client.GetRemoteAddr() != nil {
			ipAddresses.WriteString(client.GetRemoteAddr().String() + ",")
		}
	}

	message := []byte(strings.TrimRight(ipAddresses.String(), ","))
	slog.Debug("Sending peer addresses", "room", key, "message", string(message))

	for _, client := range clients {
		if client.GetRemoteAddr() == nil {
			continue
		}

		_, err := ctx.conn.WriteToUDP(message, client.GetRemoteAddr())
		if err != nil {
			slog.Error("Failed to send to client", "addr", client.GetRemoteAddr(), "error", err)
			continue
		}
	}
	return nil
}
