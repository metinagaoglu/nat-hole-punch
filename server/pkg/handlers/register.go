package handlers

import (
	"encoding/json"
	"log"
	"net"
	"strings"

	. "udp-hole-punch/pkg/models"
	. "udp-hole-punch/pkg/repositories"
)

// serverConnection holds the server's UDP connection for sending messages
// This is set by the server during initialization
var serverConnection *net.UDPConn

// SetServerConnection sets the server connection for handlers to use
func SetServerConnection(conn *net.UDPConn) {
	serverConnection = conn
}

type RegisterRequest struct {
	LocalIp string `json:"local_ip"`
	Key     string `json:"key"`
}

func Register(client *Client, payload string) error {
	var registerRequest RegisterRequest
	err := json.Unmarshal([]byte(payload), &registerRequest)
	if err != nil {
		return err
	}

	log.Printf("Registering client %s with key [%s]", client.GetRemoteAddr(), registerRequest.Key)
	repository := GetRepository()
	repository.AddClient(registerRequest.Key, client, 60)

	err = SendToClient(registerRequest.Key)
	if err != nil {
		return err
	}
	return nil
}

func SendToClient(key string) error {
	if serverConnection == nil {
		log.Printf("Warning: server connection not set, cannot send to clients")
		return nil
	}

	repository := GetRepository()
	clients, err := repository.GetClientsByKey(key)

	if err != nil {
		log.Printf("Error getting clients by key [%s]", key)
		return err
	}

	log.Printf("Sending ip addresses to clients with key [%s] and %d clients", key, len(clients))

	// Get Ip addresses split of string from key
	var ipAddresses strings.Builder
	for _, client := range clients {
		if client.GetRemoteAddr() != nil {
			ipAddresses.WriteString(client.GetRemoteAddr().String() + ",")
		}
	}

	message := []byte(strings.TrimRight(ipAddresses.String(), ","))
	log.Printf("Broadcasting ip addresses to clients with key [%s], message: [%s]", key, string(message))

	// Broadcast ip addresses to clients using server connection
	for _, client := range clients {
		if client.GetRemoteAddr() == nil {
			continue
		}

		// Use server's connection to write to client's address
		_, err := serverConnection.WriteToUDP(message, client.GetRemoteAddr())
		if err != nil {
			log.Printf("Failed to send to %s: %v", client.GetRemoteAddr(), err)
			// Continue to other clients even if one fails
			continue
		}
	}
	return nil
}
