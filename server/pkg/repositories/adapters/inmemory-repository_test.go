package adapters

import (
	"net"
	"testing"
	"time"

	"udp-hole-punch/pkg/models"
)

func TestNewInMemoryRepository(t *testing.T) {
	repo := NewInMemoryRepository()
	defer repo.Close()

	if repo == nil {
		t.Fatal("NewInMemoryRepository() returned nil")
	}

	if repo.GetRoomCount() != 0 {
		t.Error("NewInMemoryRepository() should start with no rooms")
	}

	if repo.GetClientCount() != 0 {
		t.Error("NewInMemoryRepository() should start with no clients")
	}
}

func TestInMemoryRepository_AddClient(t *testing.T) {
	repo := NewInMemoryRepository()
	client := models.NewClient()

	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	client.SetRemoteAddr(addr)

	err := repo.AddClient("test-key", client, 60)
	if err != nil {
		t.Fatalf("AddClient() error = %v", err)
	}

	clients, _ := repo.GetClientsByKey("test-key")
	if len(clients) != 1 {
		t.Errorf("AddClient() clients count = %d, want 1", len(clients))
	}

	if !clients[0].CompareAddr(client.GetRemoteAddr()) {
		t.Error("AddClient() did not add the correct client")
	}
}

func TestInMemoryRepository_AddMultipleClients(t *testing.T) {
	repo := NewInMemoryRepository()

	addr1, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	client1 := models.NewClient().SetRemoteAddr(addr1)

	addr2, _ := net.ResolveUDPAddr("udp", "127.0.0.1:5678")
	client2 := models.NewClient().SetRemoteAddr(addr2)

	repo.AddClient("room1", client1, 60)
	repo.AddClient("room1", client2, 60)

	clients, _ := repo.GetClientsByKey("room1")
	if len(clients) != 2 {
		t.Errorf("AddClient() clients count = %d, want 2", len(clients))
	}
}

func TestInMemoryRepository_GetClientsByKey(t *testing.T) {
	repo := NewInMemoryRepository()

	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	client := models.NewClient().SetRemoteAddr(addr)

	repo.AddClient("test-key", client, 60)

	clients, err := repo.GetClientsByKey("test-key")
	if err != nil {
		t.Fatalf("GetClientsByKey() error = %v", err)
	}

	if len(clients) != 1 {
		t.Errorf("GetClientsByKey() returned %d clients, want 1", len(clients))
	}
}

func TestInMemoryRepository_GetClientsByKey_EmptyKey(t *testing.T) {
	repo := NewInMemoryRepository()

	clients, err := repo.GetClientsByKey("non-existent-key")
	if err != nil {
		t.Fatalf("GetClientsByKey() error = %v", err)
	}

	if clients != nil && len(clients) != 0 {
		t.Errorf("GetClientsByKey() for non-existent key returned %d clients, want 0", len(clients))
	}
}

func TestInMemoryRepository_RemoveClient(t *testing.T) {
	repo := NewInMemoryRepository()

	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	client := models.NewClient().SetRemoteAddr(addr)

	repo.AddClient("test-key", client, 60)

	err := repo.RemoveClient("test-key", client)
	if err != nil {
		t.Fatalf("RemoveClient() error = %v", err)
	}

	clients, _ := repo.GetClientsByKey("test-key")
	if len(clients) != 0 {
		t.Errorf("RemoveClient() clients count = %d, want 0", len(clients))
	}
}

func TestInMemoryRepository_RemoveClient_MultipleClients(t *testing.T) {
	repo := NewInMemoryRepository()

	addr1, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	client1 := models.NewClient().SetRemoteAddr(addr1)

	addr2, _ := net.ResolveUDPAddr("udp", "127.0.0.1:5678")
	client2 := models.NewClient().SetRemoteAddr(addr2)

	repo.AddClient("room1", client1, 60)
	repo.AddClient("room1", client2, 60)

	// Remove only client1
	repo.RemoveClient("room1", client1)

	clients, _ := repo.GetClientsByKey("room1")
	if len(clients) != 1 {
		t.Errorf("RemoveClient() clients count = %d, want 1", len(clients))
	}

	if !clients[0].CompareAddr(client2.GetRemoteAddr()) {
		t.Error("RemoveClient() removed wrong client")
	}
}

func TestInMemoryRepository_GetClients(t *testing.T) {
	repo := NewInMemoryRepository()
	defer repo.Close()

	clients, err := repo.GetClients()
	if err != nil {
		t.Errorf("GetClients() error = %v, want nil", err)
	}

	if len(clients) != 0 {
		t.Errorf("GetClients() should return empty slice, got %d", len(clients))
	}
}

func TestInMemoryRepository_TTLExpiry(t *testing.T) {
	repo := NewInMemoryRepository()
	defer repo.Close()

	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	client := models.NewClient().SetRemoteAddr(addr)

	// Add client with 1 second TTL
	repo.AddClient("short-lived", client, 1)

	// Should exist immediately
	clients, _ := repo.GetClientsByKey("short-lived")
	if len(clients) != 1 {
		t.Fatalf("Client should exist immediately after adding, got %d", len(clients))
	}

	// Wait for expiry
	time.Sleep(2 * time.Second)

	// Should be filtered out (even without cleanup tick)
	clients, _ = repo.GetClientsByKey("short-lived")
	if len(clients) != 0 {
		t.Errorf("Client should be expired after TTL, got %d", len(clients))
	}
}
