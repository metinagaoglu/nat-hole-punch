package handlers

import (
	"encoding/json"
	"fmt"
	. "udp-hole-punch/pkg/models"
)

type RegisterRequest struct {
	LocalIp string `json:"local_ip"`
}

func Register(client *Client, payload string) error {

	var registerRequest RegisterRequest
	json.Unmarshal([]byte(payload), &registerRequest)
	fmt.Println(registerRequest.LocalIp)
	fmt.Println(client.GetRemoteAddr())
	return nil
}
