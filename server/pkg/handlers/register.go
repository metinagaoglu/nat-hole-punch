package handlers

import (
	"encoding/json"
	"log"
	"strings"

	. "udp-hole-punch/pkg/models"
)

type RegisterRequest struct {
	LocalIp string `json:"local_ip"`
	Key     string `json:"key"`
}

func register(ctx *HandlerContext, client *Client, payload string) error {
	var registerRequest RegisterRequest
	err := json.Unmarshal([]byte(payload), &registerRequest)
	if err != nil {
		return err
	}

	log.Printf("Registering client %s with key [%s]", client.GetRemoteAddr(), registerRequest.Key)
	ctx.repository.AddClient(registerRequest.Key, client, 60)

	return broadcastPeers(ctx, registerRequest.Key)
}

func broadcastPeers(ctx *HandlerContext, key string) error {
	if ctx.conn == nil {
		log.Printf("Warning: server connection not set, cannot send to clients")
		return nil
	}

	clients, err := ctx.repository.GetClientsByKey(key)
	if err != nil {
		log.Printf("Error getting clients by key [%s]", key)
		return err
	}

	log.Printf("Sending ip addresses to clients with key [%s] and %d clients", key, len(clients))

	var ipAddresses strings.Builder
	for _, client := range clients {
		if client.GetRemoteAddr() != nil {
			ipAddresses.WriteString(client.GetRemoteAddr().String() + ",")
		}
	}

	message := []byte(strings.TrimRight(ipAddresses.String(), ","))
	log.Printf("Broadcasting ip addresses to clients with key [%s], message: [%s]", key, string(message))

	for _, client := range clients {
		if client.GetRemoteAddr() == nil {
			continue
		}

		_, err := ctx.conn.WriteToUDP(message, client.GetRemoteAddr())
		if err != nil {
			log.Printf("Failed to send to %s: %v", client.GetRemoteAddr(), err)
			continue
		}
	}
	return nil
}
