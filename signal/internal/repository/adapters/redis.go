package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/metnagaoglu/holepunch/signal/internal/repository"
)

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// RedisRepository implements Repository using Redis
type RedisRepository struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisRepository creates a new Redis repository
func NewRedisRepository(config RedisConfig) (*RedisRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisRepository{client: client, ctx: ctx}, nil
}

func (r *RedisRepository) AddClient(key string, client *repository.Client, ttl int32) error {
	clientData := map[string]interface{}{
		"remote_addr": client.Addr.String(),
		"create_at":   client.CreatedAt,
	}

	data, err := json.Marshal(clientData)
	if err != nil {
		return fmt.Errorf("failed to marshal client data: %w", err)
	}

	roomKey := fmt.Sprintf("room:%s:clients", key)

	if err := r.client.SAdd(r.ctx, roomKey, string(data)).Err(); err != nil {
		return fmt.Errorf("failed to add client to Redis: %w", err)
	}

	if err := r.client.Expire(r.ctx, roomKey, time.Duration(ttl)*time.Second).Err(); err != nil {
		return fmt.Errorf("failed to set TTL: %w", err)
	}

	return nil
}

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

func (r *RedisRepository) RemoveClient(key string, addr *net.UDPAddr) error {
	roomKey := fmt.Sprintf("room:%s:clients", key)

	members, err := r.client.SMembers(r.ctx, roomKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get room members: %w", err)
	}

	targetAddr := addr.String()
	for _, member := range members {
		var clientData map[string]interface{}
		if err := json.Unmarshal([]byte(member), &clientData); err != nil {
			continue
		}
		if a, ok := clientData["remote_addr"].(string); ok && a == targetAddr {
			if err := r.client.SRem(r.ctx, roomKey, member).Err(); err != nil {
				return fmt.Errorf("failed to remove client: %w", err)
			}
			break
		}
	}
	return nil
}

func (r *RedisRepository) GetClientsByKey(key string) ([]*repository.Client, error) {
	roomKey := fmt.Sprintf("room:%s:clients", key)
	return r.getClientsByRoomKey(roomKey)
}

func (r *RedisRepository) GetClients() ([]*repository.Client, error) {
	keys, err := r.client.Keys(r.ctx, "room:*:clients").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get room keys: %w", err)
	}

	var all []*repository.Client
	for _, key := range keys {
		clients, err := r.getClientsByRoomKey(key)
		if err != nil {
			continue
		}
		all = append(all, clients...)
	}
	return all, nil
}

func (r *RedisRepository) getClientsByRoomKey(roomKey string) ([]*repository.Client, error) {
	members, err := r.client.SMembers(r.ctx, roomKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get clients: %w", err)
	}

	clients := make([]*repository.Client, 0, len(members))
	for _, member := range members {
		var clientData map[string]interface{}
		if err := json.Unmarshal([]byte(member), &clientData); err != nil {
			continue
		}

		addrStr, ok := clientData["remote_addr"].(string)
		if !ok {
			continue
		}

		addr, err := net.ResolveUDPAddr("udp", addrStr)
		if err != nil {
			continue
		}

		clients = append(clients, &repository.Client{Addr: addr})
	}
	return clients, nil
}

func (r *RedisRepository) Close() error {
	return r.client.Close()
}

// Ping checks if Redis connection is alive
func (r *RedisRepository) Ping() error {
	return r.client.Ping(r.ctx).Err()
}
