package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/redis/go-redis/v9"

	"udp-hole-punch/pkg/models"
)

// RedisRepository implements IRepository using Redis as storage backend
type RedisRepository struct {
	client *redis.Client
	ctx    context.Context
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// NewRedisRepository creates a new Redis repository instance
func NewRedisRepository(config RedisConfig) (*RedisRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisRepository{
		client: client,
		ctx:    ctx,
	}, nil
}

// AddClient adds a client to a room with TTL
func (r *RedisRepository) AddClient(key string, client *models.Client, ttl int32) error {
	// Serialize client data
	clientData := map[string]interface{}{
		"remote_addr": client.GetRemoteAddr().String(),
		"create_at":   client.GetCreateAt(),
	}

	data, err := json.Marshal(clientData)
	if err != nil {
		return fmt.Errorf("failed to marshal client data: %w", err)
	}

	// Use Redis Set to store unique clients per room
	// Key format: room:{key}:clients
	roomKey := fmt.Sprintf("room:%s:clients", key)

	// Add client to set
	if err := r.client.SAdd(r.ctx, roomKey, string(data)).Err(); err != nil {
		return fmt.Errorf("failed to add client to Redis: %w", err)
	}

	// Set TTL for the room key
	if err := r.client.Expire(r.ctx, roomKey, time.Duration(ttl)*time.Second).Err(); err != nil {
		return fmt.Errorf("failed to set TTL: %w", err)
	}

	return nil
}

// RemoveClient removes a client from a room
func (r *RedisRepository) RemoveClient(key string, client *models.Client) error {
	roomKey := fmt.Sprintf("room:%s:clients", key)

	// Get all clients in the room
	members, err := r.client.SMembers(r.ctx, roomKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get room members: %w", err)
	}

	// Find and remove the matching client
	targetAddr := client.GetRemoteAddr().String()
	for _, member := range members {
		var clientData map[string]interface{}
		if err := json.Unmarshal([]byte(member), &clientData); err != nil {
			continue
		}

		if addr, ok := clientData["remote_addr"].(string); ok && addr == targetAddr {
			if err := r.client.SRem(r.ctx, roomKey, member).Err(); err != nil {
				return fmt.Errorf("failed to remove client: %w", err)
			}
			break
		}
	}

	return nil
}

// RefreshClient extends TTL for all rooms containing this client address
func (r *RedisRepository) RefreshClient(addr string, ttl int32) {
	keys, err := r.client.Keys(r.ctx, "room:*:clients").Result()
	if err != nil {
		return
	}

	duration := time.Duration(ttl) * time.Second
	for _, key := range keys {
		members, err := r.client.SMembers(r.ctx, key).Result()
		if err != nil {
			continue
		}
		for _, member := range members {
			var clientData map[string]interface{}
			if err := json.Unmarshal([]byte(member), &clientData); err != nil {
				continue
			}
			if a, ok := clientData["remote_addr"].(string); ok && a == addr {
				r.client.Expire(r.ctx, key, duration)
				break
			}
		}
	}
}

// GetClients returns all clients across all rooms
func (r *RedisRepository) GetClients() ([]*models.Client, error) {
	// Get all room keys
	keys, err := r.client.Keys(r.ctx, "room:*:clients").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get room keys: %w", err)
	}

	var allClients []*models.Client
	for _, key := range keys {
		clients, err := r.getClientsByRoomKey(key)
		if err != nil {
			continue // Skip failed rooms
		}
		allClients = append(allClients, clients...)
	}

	return allClients, nil
}

// GetClientsByKey returns all clients in a specific room
func (r *RedisRepository) GetClientsByKey(key string) ([]*models.Client, error) {
	roomKey := fmt.Sprintf("room:%s:clients", key)
	return r.getClientsByRoomKey(roomKey)
}

// getClientsByRoomKey is a helper to retrieve clients from a room key
func (r *RedisRepository) getClientsByRoomKey(roomKey string) ([]*models.Client, error) {
	members, err := r.client.SMembers(r.ctx, roomKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get clients: %w", err)
	}

	clients := make([]*models.Client, 0, len(members))
	for _, member := range members {
		var clientData map[string]interface{}
		if err := json.Unmarshal([]byte(member), &clientData); err != nil {
			continue // Skip malformed data
		}

		// Parse remote address from stored data
		addrStr, ok := clientData["remote_addr"].(string)
		if !ok {
			continue
		}

		// Reconstruct client with address (but no connection)
		// Connection will be set by the server when sending
		addr, err := net.ResolveUDPAddr("udp", addrStr)
		if err != nil {
			continue
		}

		client := models.NewClient().SetRemoteAddr(addr)

		// Set timestamp if available
		if _, ok := clientData["create_at"].(float64); ok {
			client.SetCreateAt()
		}

		clients = append(clients, client)
	}

	return clients, nil
}

// Close closes the Redis connection
func (r *RedisRepository) Close() error {
	return r.client.Close()
}

// Ping checks if Redis connection is alive
func (r *RedisRepository) Ping() error {
	return r.client.Ping(r.ctx).Err()
}
