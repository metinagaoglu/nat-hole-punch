package handlers

import (
	"encoding/json"
	"log"

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

	log.Printf("Logout client %s with key [%s]", client.GetRemoteAddr(), logoutRequest.Key)
	err = ctx.repository.RemoveClient(logoutRequest.Key, client)
	if err != nil {
		return err
	}

	return broadcastPeers(ctx, logoutRequest.Key)
}
