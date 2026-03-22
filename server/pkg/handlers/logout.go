package handlers

import (
	"encoding/json"
	"log/slog"

	. "udp-hole-punch/pkg/models"
)

type LogoutRequest struct {
	LocalIp string `json:"local_ip"`
	Key     string `json:"key"`
}

func logout(ctx *HandlerContext, client *Client, payload string) error {
	if err := validatePayload(payload); err != nil {
		return err
	}

	var logoutRequest LogoutRequest
	err := json.Unmarshal([]byte(payload), &logoutRequest)
	if err != nil {
		return err
	}

	if err := validateKey(logoutRequest.Key); err != nil {
		return err
	}

	slog.Info("Client logout", "addr", client.GetRemoteAddr(), "room", logoutRequest.Key)
	err = ctx.repository.RemoveClient(logoutRequest.Key, client)
	if err != nil {
		return err
	}

	return broadcastPeers(ctx, logoutRequest.Key)
}
