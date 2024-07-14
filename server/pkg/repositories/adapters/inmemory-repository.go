package adapters

// import redis
import (
	. "udp-hole-punch/pkg/models"
)

type InMemoryRepository struct {
	clients map[string][]*Client
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		clients: map[string][]*Client{},
	}
}

func (r *InMemoryRepository) AddClient(key string, client *Client, ttl int32) error {
	r.clients[key] = append(r.clients[key], client)
	return nil
}

func (r *InMemoryRepository) GetClients() ([]*Client, error) {
	return nil, nil
}

func (r *InMemoryRepository) GetClientsByKey(key string) ([]*Client, error) {
	return r.clients[key], nil
}
