package repository

import "net"

// Client represents a connected UDP client
type Client struct {
	Addr      *net.UDPAddr
	Conn      *net.UDPConn
	CreatedAt int64
}

// Repository defines the storage interface for client management
type Repository interface {
	AddClient(key string, client *Client, ttl int32) error
	RemoveClient(key string, addr *net.UDPAddr) error
	RefreshClient(addr string, ttl int32)
	GetClientsByKey(key string) ([]*Client, error)
	GetClients() ([]*Client, error)
	Close() error
}
