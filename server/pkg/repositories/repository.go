package repositories

import (
	"log/slog"

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
			slog.Warn("Failed to initialize Redis, falling back to in-memory", "error", err)
			return adapters.NewInMemoryRepository()
		}

		slog.Info("Using Redis repository", "addr", redisConfig.Addr)
		return redisRepo

	default:
		slog.Info("Using in-memory repository")
		return adapters.NewInMemoryRepository()
	}
}
