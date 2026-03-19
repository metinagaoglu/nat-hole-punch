package handlers

import (
	"encoding/json"
	"log"

	. "udp-hole-punch/pkg/models"
	. "udp-hole-punch/pkg/repositories"
)

// Note: serverConnection is defined in register.go and shared across handlers

type LogoutRequest struct {
	LocalIp string `json:"local_ip"`
	Key     string `json:"key"`
}

func Logout(client *Client, payload string) error {
	var logoutRequest LogoutRequest
	err := json.Unmarshal([]byte(payload), &logoutRequest)
	if err != nil {
		return err
	}

	log.Printf("Logout client %s with key [%s]", client.GetRemoteAddr(), logoutRequest.Key)
	repository := GetRepository()
	err = repository.RemoveClient(logoutRequest.Key, client)
	if err != nil {
		return err
	}

	err = SendToClient(logoutRequest.Key)
	if err != nil {
		return err
	}
	return nil
}
