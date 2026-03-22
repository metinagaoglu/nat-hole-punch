package adapters

import (
	"log/slog"
	"sync"
	"time"

	. "udp-hole-punch/pkg/models"
)

// clientEntry wraps a Client with expiration tracking
type clientEntry struct {
	client    *Client
	expiresAt time.Time
}

// InMemoryRepository provides thread-safe in-memory storage for clients
type InMemoryRepository struct {
	mu      sync.RWMutex
	rooms   map[string][]clientEntry
	stopCh  chan struct{}
	stopped bool
}

// NewInMemoryRepository creates a new thread-safe in-memory repository
// with a background cleanup goroutine
func NewInMemoryRepository() *InMemoryRepository {
	r := &InMemoryRepository{
		rooms:  make(map[string][]clientEntry),
		stopCh: make(chan struct{}),
	}
	go r.cleanupLoop()
	return r
}

// cleanupLoop periodically removes expired clients
func (r *InMemoryRepository) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.stopCh:
			return
		case <-ticker.C:
			r.evictExpired()
		}
	}
}

// evictExpired removes all expired client entries
func (r *InMemoryRepository) evictExpired() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	totalEvicted := 0

	for key, entries := range r.rooms {
		alive := entries[:0]
		for _, e := range entries {
			if now.Before(e.expiresAt) {
				alive = append(alive, e)
			} else {
				totalEvicted++
			}
		}

		if len(alive) == 0 {
			delete(r.rooms, key)
		} else {
			r.rooms[key] = alive
		}
	}

	if totalEvicted > 0 {
		slog.Info("Evicted expired clients", "count", totalEvicted, "rooms_remaining", len(r.rooms))
	}
}

// AddClient adds a client to a room with the specified TTL (in seconds)
func (r *InMemoryRepository) AddClient(key string, client *Client, ttl int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry := clientEntry{
		client:    client,
		expiresAt: time.Now().Add(time.Duration(ttl) * time.Second),
	}

	r.rooms[key] = append(r.rooms[key], entry)
	slog.Debug("Client added", "room", key, "total", len(r.rooms[key]), "ttl", ttl)
	return nil
}

// RefreshClient extends the TTL for an active client across all rooms
func (r *InMemoryRepository) RefreshClient(addr string, ttl int32) {
	r.mu.Lock()
	defer r.mu.Unlock()

	newExpiry := time.Now().Add(time.Duration(ttl) * time.Second)
	for _, entries := range r.rooms {
		for i := range entries {
			if entries[i].client.GetRemoteAddr() != nil &&
				entries[i].client.GetRemoteAddr().String() == addr {
				entries[i].expiresAt = newExpiry
				return
			}
		}
	}
}

// RemoveClient removes a client from a room
func (r *InMemoryRepository) RemoveClient(key string, client *Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entries := r.rooms[key]
	for i, e := range entries {
		if e.client.GetRemoteAddr() != nil && client.CompareAddr(e.client.GetRemoteAddr()) {
			r.rooms[key] = append(entries[:i], entries[i+1:]...)
			if len(r.rooms[key]) == 0 {
				delete(r.rooms, key)
			}
			slog.Debug("Client removed", "room", key)
			return nil
		}
	}

	slog.Debug("Client not found in room", "room", key)
	return nil
}

// GetClients returns all non-expired clients across all rooms
func (r *InMemoryRepository) GetClients() ([]*Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	var allClients []*Client
	for _, entries := range r.rooms {
		for _, e := range entries {
			if now.Before(e.expiresAt) {
				allClients = append(allClients, e.client)
			}
		}
	}
	return allClients, nil
}

// GetClientsByKey returns all non-expired clients in a specific room
func (r *InMemoryRepository) GetClientsByKey(key string) ([]*Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	entries := r.rooms[key]
	if entries == nil {
		return []*Client{}, nil
	}

	result := make([]*Client, 0, len(entries))
	for _, e := range entries {
		if now.Before(e.expiresAt) {
			result = append(result, e.client)
		}
	}
	return result, nil
}

// Close stops the background cleanup goroutine
func (r *InMemoryRepository) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.stopped {
		close(r.stopCh)
		r.stopped = true
		slog.Debug("InMemoryRepository cleanup stopped")
	}
	return nil
}

// GetRoomCount returns the number of active rooms
func (r *InMemoryRepository) GetRoomCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.rooms)
}

// GetClientCount returns the total number of clients across all rooms
func (r *InMemoryRepository) GetClientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, entries := range r.rooms {
		count += len(entries)
	}
	return count
}
