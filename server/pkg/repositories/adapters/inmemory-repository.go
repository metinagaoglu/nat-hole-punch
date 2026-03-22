package adapters

import (
	"log"
	"sync"

	. "udp-hole-punch/pkg/models"
)

// InMemoryRepository provides thread-safe in-memory storage for clients
type InMemoryRepository struct {
	mu      sync.RWMutex
	clients map[string][]*Client
}

// NewInMemoryRepository creates a new thread-safe in-memory repository
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		clients: make(map[string][]*Client),
	}
}

// AddClient adds a client to a room with the specified TTL
func (r *InMemoryRepository) AddClient(key string, client *Client, ttl int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.clients[key] = append(r.clients[key], client)
	log.Printf("Client added to room [%s], total clients: %d", key, len(r.clients[key]))
	return nil
}

// RemoveClient removes a client from a room
func (r *InMemoryRepository) RemoveClient(key string, client *Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	clients := r.clients[key]
	for i, c := range clients {
		if c.GetRemoteAddr() != nil && client.CompareAddr(c.GetRemoteAddr()) {
			// Remove client by swapping with last element and truncating
			r.clients[key] = append(clients[:i], clients[i+1:]...)
			log.Printf("Client removed from room [%s], remaining clients: %d", key, len(r.clients[key]))
			return nil
		}
	}

	log.Printf("Client not found in room [%s]", key)
	return nil
}

// GetClients returns all clients across all rooms
func (r *InMemoryRepository) GetClients() ([]*Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var allClients []*Client
	for _, clients := range r.clients {
		allClients = append(allClients, clients...)
	}
	return allClients, nil
}

// GetClientsByKey returns all clients in a specific room
func (r *InMemoryRepository) GetClientsByKey(key string) ([]*Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification
	clients := r.clients[key]
	if clients == nil {
		return []*Client{}, nil
	}

	result := make([]*Client, len(clients))
	copy(result, clients)
	return result, nil
}

// GetRoomCount returns the number of active rooms
func (r *InMemoryRepository) GetRoomCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}

// GetClientCount returns the total number of clients across all rooms
func (r *InMemoryRepository) GetClientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, clients := range r.clients {
		count += len(clients)
	}
	return count
}
