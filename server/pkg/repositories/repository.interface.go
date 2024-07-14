package repositories

import (
	. "udp-hole-punch/pkg/models"
)

type IRepository interface {
	AddClient(key string, client *Client, ttl int32) error
	GetClients() ([]*Client, error)
	GetClientsByKey(key string) ([]*Client, error)
}
