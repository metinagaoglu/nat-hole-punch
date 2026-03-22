package repositories

import (
	"log"

	"udp-hole-punch/pkg/config"
	"udp-hole-punch/pkg/repositories/adapters"
)

// CreateRepository creates a repository instance based on configuration
func CreateRepository(cfg *config.Config) IRepository {
	switch cfg.RepositoryType {
	case "redis":
		redisConfig := adapters.RedisConfig{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		}

		redisRepo, err := adapters.NewRedisRepository(redisConfig)
		if err != nil {
			log.Printf("Failed to initialize Redis repository: %v, falling back to in-memory", err)
			return adapters.NewInMemoryRepository()
		}

		log.Printf("Using Redis repository at %s", redisConfig.Addr)
		return redisRepo

	default:
		log.Printf("Using in-memory repository")
		return adapters.NewInMemoryRepository()
	}
}
