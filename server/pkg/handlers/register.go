package handlers

import (
	"encoding/json"
	"fmt"
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
	for _, client := range clients[key] {
		fmt.Println(client)
		for _, subClient := range clients[key] {
			if client.GetRemoteAddr().String() == subClient.GetRemoteAddr().String() {
				continue
			}
			//client.Send([]byte(subClient.GetRemoteAddr().String()))
			client.GetConn().WriteToUDP([]byte(subClient.GetRemoteAddr().String()), client.GetRemoteAddr())
		}
	}
	return nil
}
