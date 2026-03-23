package adapters

import (
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/metnagaoglu/holepunch/signal/internal/repository"
)

type clientEntry struct {
	client    *repository.Client
	expiresAt time.Time
}

// MemoryRepository provides thread-safe in-memory storage
type MemoryRepository struct {
	mu      sync.RWMutex
	rooms   map[string][]clientEntry
	stopCh  chan struct{}
	stopped bool
}

// NewMemoryRepository creates a new in-memory repository with background cleanup
func NewMemoryRepository() *MemoryRepository {
	r := &MemoryRepository{
		rooms:  make(map[string][]clientEntry),
		stopCh: make(chan struct{}),
	}
	go r.cleanupLoop()
	return r
}

func (r *MemoryRepository) cleanupLoop() {
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

func (r *MemoryRepository) evictExpired() {
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

func (r *MemoryRepository) AddClient(key string, client *repository.Client, ttl int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry := clientEntry{
		client:    client,
		expiresAt: time.Now().Add(time.Duration(ttl) * time.Second),
	}
	r.rooms[key] = append(r.rooms[key], entry)
	slog.Debug("Client added", "room", key, "total", len(r.rooms[key]))
	return nil
}

func (r *MemoryRepository) RefreshClient(addr string, ttl int32) {
	r.mu.Lock()
	defer r.mu.Unlock()

	newExpiry := time.Now().Add(time.Duration(ttl) * time.Second)
	for _, entries := range r.rooms {
		for i := range entries {
			if entries[i].client.Addr != nil && entries[i].client.Addr.String() == addr {
				entries[i].expiresAt = newExpiry
				return
			}
		}
	}
}

func (r *MemoryRepository) RemoveClient(key string, addr *net.UDPAddr) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entries := r.rooms[key]
	for i, e := range entries {
		if e.client.Addr != nil && e.client.Addr.String() == addr.String() {
			r.rooms[key] = append(entries[:i], entries[i+1:]...)
			if len(r.rooms[key]) == 0 {
				delete(r.rooms, key)
			}
			slog.Debug("Client removed", "room", key)
			return nil
		}
	}
	return nil
}

func (r *MemoryRepository) GetClientsByKey(key string) ([]*repository.Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	entries := r.rooms[key]
	if entries == nil {
		return []*repository.Client{}, nil
	}

	result := make([]*repository.Client, 0, len(entries))
	for _, e := range entries {
		if now.Before(e.expiresAt) {
			result = append(result, e.client)
		}
	}
	return result, nil
}

func (r *MemoryRepository) GetClients() ([]*repository.Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	var all []*repository.Client
	for _, entries := range r.rooms {
		for _, e := range entries {
			if now.Before(e.expiresAt) {
				all = append(all, e.client)
			}
		}
	}
	return all, nil
}

func (r *MemoryRepository) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.stopped {
		close(r.stopCh)
		r.stopped = true
	}
	return nil
}

// GetRoomCount returns the number of active rooms (for testing)
func (r *MemoryRepository) GetRoomCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.rooms)
}

// GetClientCount returns the total number of clients (for testing)
func (r *MemoryRepository) GetClientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, entries := range r.rooms {
		count += len(entries)
	}
	return count
}
