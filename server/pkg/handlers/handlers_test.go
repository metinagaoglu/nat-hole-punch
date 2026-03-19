package handlers

import (
	"net"
	"testing"

	"udp-hole-punch/pkg/models"
	"udp-hole-punch/pkg/repositories"
	"udp-hole-punch/pkg/repositories/adapters"
)

func setupTestRepository() {
	// Reset repository to use in-memory for testing
	repositories.SetRepository(adapters.NewInMemoryRepository())
}

func TestRegister(t *testing.T) {
	setupTestRepository()

	client := models.NewClient()
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	localAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		t.Fatalf("ListenUDP error = %v", err)
	}
	defer conn.Close()

	client.SetRemoteAddr(addr).SetConn(conn)

	payload := `{"local_ip":"127.0.0.1:1234","key":"test-room"}`

	err = Register(client, payload)
	if err != nil {
		t.Errorf("Register() error = %v", err)
	}

	// Verify client was added to repository
	repo := repositories.GetRepository()
	clients, _ := repo.GetClientsByKey("test-room")

	if len(clients) != 1 {
		t.Errorf("Register() should add client to repository, got %d clients", len(clients))
	}
}

func TestRegister_InvalidJSON(t *testing.T) {
	setupTestRepository()

	client := models.NewClient()
	invalidPayload := `{invalid json}`

	err := Register(client, invalidPayload)
	if err == nil {
		t.Error("Register() should return error for invalid JSON")
	}
}

func TestRegister_MultipleClients(t *testing.T) {
	setupTestRepository()

	localAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		t.Fatalf("ListenUDP error = %v", err)
	}
	defer conn.Close()

	addr1, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	client1 := models.NewClient().SetRemoteAddr(addr1).SetConn(conn)

	addr2, _ := net.ResolveUDPAddr("udp", "127.0.0.1:5678")
	client2 := models.NewClient().SetRemoteAddr(addr2).SetConn(conn)

	payload1 := `{"local_ip":"127.0.0.1:1234","key":"game-room"}`
	payload2 := `{"local_ip":"127.0.0.1:5678","key":"game-room"}`

	Register(client1, payload1)
	Register(client2, payload2)

	repo := repositories.GetRepository()
	clients, _ := repo.GetClientsByKey("game-room")

	if len(clients) != 2 {
		t.Errorf("Register() should add multiple clients, got %d", len(clients))
	}
}

func TestLogout(t *testing.T) {
	setupTestRepository()

	localAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		t.Fatalf("ListenUDP error = %v", err)
	}
	defer conn.Close()

	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	client := models.NewClient().SetRemoteAddr(addr).SetConn(conn)

	// First register the client
	registerPayload := `{"local_ip":"127.0.0.1:1234","key":"test-room"}`
	Register(client, registerPayload)

	// Now logout
	logoutPayload := `{"local_ip":"127.0.0.1:1234","key":"test-room"}`
	err = Logout(client, logoutPayload)
	if err != nil {
		t.Errorf("Logout() error = %v", err)
	}

	// Verify client was removed
	repo := repositories.GetRepository()
	clients, _ := repo.GetClientsByKey("test-room")

	if len(clients) != 0 {
		t.Errorf("Logout() should remove client, got %d clients remaining", len(clients))
	}
}

func TestLogout_InvalidJSON(t *testing.T) {
	setupTestRepository()

	client := models.NewClient()
	invalidPayload := `{invalid json}`

	err := Logout(client, invalidPayload)
	if err == nil {
		t.Error("Logout() should return error for invalid JSON")
	}
}

func TestSendToClient_EmptyKey(t *testing.T) {
	setupTestRepository()

	err := SendToClient("non-existent-key")

	// Should handle empty key gracefully
	if err != nil {
		t.Logf("SendToClient() returned error for empty key: %v", err)
	}
}
