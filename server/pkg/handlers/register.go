package handlers

import (
	"encoding/json"
	"fmt"
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
	json.Unmarshal([]byte(payload), &registerRequest)
	fmt.Println(registerRequest.LocalIp)
	fmt.Println(client.GetRemoteAddr())
	fmt.Println("==========")
	clients[registerRequest.Key] = append(clients[registerRequest.Key], client)
	SendToClient(registerRequest.Key)
	return nil
}

// TODO: refactor: seperate the concerns(send logic)
func SendToClient(key string) error {

	// Get Ip addreses split of string from key
	var ipAdresses strings.Builder
	for _, client := range clients[key] {
		ipAdresses.WriteString(client.GetRemoteAddr().String() + ",")
	}

	// Broadcast ip addresses to clients
	for _, client := range clients[key] {
		client.GetConn().WriteToUDP([]byte(strings.TrimRight(ipAdresses.String(), ",")), client.GetRemoteAddr())
	}
	return nil
}
