package adapters

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"udp-hole-punch/pkg/models"
)

// skipIfNoRedis skips the test if Redis is not available
func skipIfNoRedis(t *testing.T) *RedisRepository {
	redisAddr := os.Getenv("REDIS_TEST_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	config := RedisConfig{
		Addr:     redisAddr,
		Password: "",
		DB:       15, // Use DB 15 for testing
	}

	repo, err := NewRedisRepository(config)
	if err != nil {
		t.Skipf("Skipping Redis test: %v", err)
		return nil
	}

	// Clean up before test
	repo.client.FlushDB(repo.ctx)

	return repo
}

func TestNewRedisRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	config := RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       15,
	}

	repo, err := NewRedisRepository(config)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}
	defer repo.Close()

	if repo == nil {
		t.Fatal("NewRedisRepository() returned nil")
	}

	// Test ping
	if err := repo.Ping(); err != nil {
		t.Errorf("Ping() failed: %v", err)
	}
}

func TestNewRedisRepository_InvalidConnection(t *testing.T) {
	config := RedisConfig{
		Addr:     "invalid:9999",
		Password: "",
		DB:       0,
	}

	_, err := NewRedisRepository(config)
	if err == nil {
		t.Error("NewRedisRepository() should fail with invalid address")
	}
}

func TestRedisRepository_AddClient(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	repo := skipIfNoRedis(t)
	if repo == nil {
		return
	}
	defer repo.Close()

	client := models.NewClient()
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	client.SetRemoteAddr(addr).SetCreateAt()

	err := repo.AddClient("test-room", client, 60)
	if err != nil {
		t.Fatalf("AddClient() error = %v", err)
	}

	// Verify client was added
	clients, err := repo.GetClientsByKey("test-room")
	if err != nil {
		t.Fatalf("GetClientsByKey() error = %v", err)
	}

	if len(clients) != 1 {
		t.Errorf("Expected 1 client, got %d", len(clients))
	}
}

func TestRedisRepository_AddMultipleClients(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	repo := skipIfNoRedis(t)
	if repo == nil {
		return
	}
	defer repo.Close()

	addr1, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	client1 := models.NewClient().SetRemoteAddr(addr1).SetCreateAt()

	addr2, _ := net.ResolveUDPAddr("udp", "127.0.0.1:5678")
	client2 := models.NewClient().SetRemoteAddr(addr2).SetCreateAt()

	repo.AddClient("game-room", client1, 60)
	repo.AddClient("game-room", client2, 60)

	clients, _ := repo.GetClientsByKey("game-room")
	if len(clients) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(clients))
	}
}

func TestRedisRepository_RemoveClient(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	repo := skipIfNoRedis(t)
	if repo == nil {
		return
	}
	defer repo.Close()

	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	client := models.NewClient().SetRemoteAddr(addr).SetCreateAt()

	repo.AddClient("test-room", client, 60)

	err := repo.RemoveClient("test-room", client)
	if err != nil {
		t.Fatalf("RemoveClient() error = %v", err)
	}

	clients, _ := repo.GetClientsByKey("test-room")
	if len(clients) != 0 {
		t.Errorf("Expected 0 clients after removal, got %d", len(clients))
	}
}

func TestRedisRepository_TTL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	repo := skipIfNoRedis(t)
	if repo == nil {
		return
	}
	defer repo.Close()

	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	client := models.NewClient().SetRemoteAddr(addr).SetCreateAt()

	// Add client with 2 second TTL
	repo.AddClient("ttl-test", client, 2)

	// Verify client exists
	clients, _ := repo.GetClientsByKey("ttl-test")
	if len(clients) != 1 {
		t.Errorf("Expected 1 client before TTL, got %d", len(clients))
	}

	// Wait for TTL to expire
	time.Sleep(3 * time.Second)

	// Check if key expired
	roomKey := "room:ttl-test:clients"
	exists, err := repo.client.Exists(repo.ctx, roomKey).Result()
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}

	if exists != 0 {
		t.Error("Room key should have expired after TTL")
	}
}

func TestRedisRepository_GetClients(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	repo := skipIfNoRedis(t)
	if repo == nil {
		return
	}
	defer repo.Close()

	// Add clients to multiple rooms
	addr1, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	client1 := models.NewClient().SetRemoteAddr(addr1).SetCreateAt()

	addr2, _ := net.ResolveUDPAddr("udp", "127.0.0.1:5678")
	client2 := models.NewClient().SetRemoteAddr(addr2).SetCreateAt()

	repo.AddClient("room1", client1, 60)
	repo.AddClient("room2", client2, 60)

	clients, err := repo.GetClients()
	if err != nil {
		t.Fatalf("GetClients() error = %v", err)
	}

	if len(clients) < 2 {
		t.Errorf("Expected at least 2 clients across all rooms, got %d", len(clients))
	}
}

func TestRedisRepository_Ping(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	repo := skipIfNoRedis(t)
	if repo == nil {
		return
	}
	defer repo.Close()

	err := repo.Ping()
	if err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}

func TestRedisRepository_Close(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	config := RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       15,
	}

	repo, err := NewRedisRepository(config)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}

	err = repo.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// After close, Ping should fail
	err = repo.Ping()
	if err == nil || err != redis.ErrClosed {
		t.Error("Ping() should fail after Close()")
	}
}
