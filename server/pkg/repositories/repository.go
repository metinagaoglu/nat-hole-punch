package repositories

import (
	"log"
	"os"

	"udp-hole-punch/pkg/repositories/adapters"
)

var repository IRepository

func GetRepository() IRepository {
	if repository == nil {
		repository = initializeRepository()
	}
	return repository
}

// initializeRepository creates repository based on REPOSITORY_TYPE env var
func initializeRepository() IRepository {
	repoType := os.Getenv("REPOSITORY_TYPE")
	if repoType == "" {
		repoType = "memory"
	}

	switch repoType {
	case "redis":
		config := adapters.RedisConfig{
			Addr:     getEnvOrDefault("REDIS_ADDR", "localhost:6379"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       0, // You can make this configurable too
		}

		redisRepo, err := adapters.NewRedisRepository(config)
		if err != nil {
			log.Printf("Failed to initialize Redis repository: %v, falling back to in-memory", err)
			return adapters.NewInMemoryRepository()
		}

		log.Printf("Using Redis repository at %s", config.Addr)
		return redisRepo

	case "memory":
		fallthrough
	default:
		log.Printf("Using in-memory repository")
		return adapters.NewInMemoryRepository()
	}
}

// SetRepository allows setting a custom repository (useful for testing)
func SetRepository(repo IRepository) {
	repository = repo
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
