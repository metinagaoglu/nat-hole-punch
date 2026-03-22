package repositories

import (
	"testing"

	"udp-hole-punch/pkg/config"
)

func TestCreateRepository_DefaultMemory(t *testing.T) {
	cfg := config.DefaultConfig()

	repo := CreateRepository(cfg)
	if repo == nil {
		t.Fatal("CreateRepository() returned nil")
	}

	_, ok := repo.(IRepository)
	if !ok {
		t.Error("CreateRepository() should return an object implementing IRepository")
	}
}

func TestCreateRepository_InvalidRedis(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RepositoryType = "redis"
	cfg.RedisAddr = "localhost:19999" // Non-existent Redis

	// Should fall back to in-memory
	repo := CreateRepository(cfg)
	if repo == nil {
		t.Fatal("CreateRepository() should fall back to in-memory when Redis fails")
	}
}
