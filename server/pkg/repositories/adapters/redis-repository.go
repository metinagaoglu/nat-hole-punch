package adapters

import (
	"github.com/go-redis/redis"

	. "udp-hole-punch/pkg/models"
)

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{
		client: client,
	}
}

func (r *RedisRepository) AddClient(key string, client *Client, ttl int32) error {
	return nil
}

func (r *RedisRepository) GetClients() ([]*Client, error) {
	return nil, nil
}

func (r *RedisRepository) GetClientsByKey(key string) ([]*Client, error) {
	return nil, nil
}
