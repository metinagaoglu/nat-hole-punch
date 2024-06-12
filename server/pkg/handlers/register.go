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

// Hold clients by key map and clietn array
var clients = map[string][]*Client{}

func Register(client *Client, payload string) error {

	var registerRequest RegisterRequest
	err := json.Unmarshal([]byte(payload), &registerRequest)
	if err != nil {
		return err
	}

	log.Printf("Registering client %s with key [%s]", client.GetRemoteAddr(), registerRequest.Key)
	clients[registerRequest.Key] = append(clients[registerRequest.Key], client)
	err = SendToClient(registerRequest.Key)
	if err != nil {
		return err
	}
	return nil
}

// TODO: refactor: seperate the concerns(send logic)
func SendToClient(key string) error {
	// Get Ip addresses split of string from key
	var ipAddresses strings.Builder
	for _, client := range clients[key] {
		ipAddresses.WriteString(client.GetRemoteAddr().String() + ",")
	}

	log.Printf("Broadcasting ip addresses to clients with key [%s] , and with clients [%s]", key, ipAddresses.String())

	// Broadcast ip addresses to clients
	for _, client := range clients[key] {
		_, err := client.GetConn().WriteToUDP([]byte(strings.TrimRight(ipAddresses.String(), ",")), client.GetRemoteAddr())
		if err != nil {
			return err
		}
	}
	return nil
}
