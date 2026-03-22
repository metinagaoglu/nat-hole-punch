package repositories

import (
	. "udp-hole-punch/pkg/models"
)

type IRepository interface {
	AddClient(key string, client *Client, ttl int32) error
	RemoveClient(key string, client *Client) error
	RefreshClient(addr string, ttl int32)
	GetClients() ([]*Client, error)
	GetClientsByKey(key string) ([]*Client, error)
	Close() error
}
