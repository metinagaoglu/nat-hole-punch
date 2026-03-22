package handlers

import (
	"net"
	"testing"

	"udp-hole-punch/pkg/models"
	"udp-hole-punch/pkg/repositories/adapters"
)

func newTestContext() *HandlerContext {
	repo := adapters.NewInMemoryRepository()
	return NewHandlerContext(repo, 60)
}

func TestRegister(t *testing.T) {
	ctx := newTestContext()

	client := models.NewClient()
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	client.SetRemoteAddr(addr)

	payload := `{"local_ip":"127.0.0.1:1234","key":"test-room"}`

	err := register(ctx, client, payload)
	if err != nil {
		t.Errorf("register() error = %v", err)
	}

	clients, _ := ctx.repository.GetClientsByKey("test-room")
	if len(clients) != 1 {
		t.Errorf("register() should add client to repository, got %d clients", len(clients))
	}
}

func TestRegister_InvalidJSON(t *testing.T) {
	ctx := newTestContext()

	client := models.NewClient()
	invalidPayload := `{invalid json}`

	err := register(ctx, client, invalidPayload)
	if err == nil {
		t.Error("register() should return error for invalid JSON")
	}
}

func TestRegister_MultipleClients(t *testing.T) {
	ctx := newTestContext()

	addr1, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	client1 := models.NewClient().SetRemoteAddr(addr1)

	addr2, _ := net.ResolveUDPAddr("udp", "127.0.0.1:5678")
	client2 := models.NewClient().SetRemoteAddr(addr2)

	payload1 := `{"local_ip":"127.0.0.1:1234","key":"game-room"}`
	payload2 := `{"local_ip":"127.0.0.1:5678","key":"game-room"}`

	register(ctx, client1, payload1)
	register(ctx, client2, payload2)

	clients, _ := ctx.repository.GetClientsByKey("game-room")
	if len(clients) != 2 {
		t.Errorf("register() should add multiple clients, got %d", len(clients))
	}
}

func TestLogout(t *testing.T) {
	ctx := newTestContext()

	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	client := models.NewClient().SetRemoteAddr(addr)

	registerPayload := `{"local_ip":"127.0.0.1:1234","key":"test-room"}`
	register(ctx, client, registerPayload)

	logoutPayload := `{"local_ip":"127.0.0.1:1234","key":"test-room"}`
	err := logout(ctx, client, logoutPayload)
	if err != nil {
		t.Errorf("logout() error = %v", err)
	}

	clients, _ := ctx.repository.GetClientsByKey("test-room")
	if len(clients) != 0 {
		t.Errorf("logout() should remove client, got %d clients remaining", len(clients))
	}
}

func TestLogout_InvalidJSON(t *testing.T) {
	ctx := newTestContext()

	client := models.NewClient()
	invalidPayload := `{invalid json}`

	err := logout(ctx, client, invalidPayload)
	if err == nil {
		t.Error("logout() should return error for invalid JSON")
	}
}

func TestBroadcastPeers_NoConnection(t *testing.T) {
	ctx := newTestContext()

	// ctx.conn is nil, should handle gracefully
	err := broadcastPeers(ctx, "non-existent-key")
	if err != nil {
		t.Logf("broadcastPeers() returned error: %v", err)
	}
}

func TestRegister_EmptyPayload(t *testing.T) {
	ctx := newTestContext()
	client := models.NewClient()

	err := register(ctx, client, "")
	if err == nil {
		t.Error("register() should reject empty payload")
	}
}

func TestRegister_EmptyKey(t *testing.T) {
	ctx := newTestContext()
	client := models.NewClient()

	payload := `{"local_ip":"127.0.0.1:1234","key":""}`
	err := register(ctx, client, payload)
	if err == nil {
		t.Error("register() should reject empty key")
	}
}

func TestRegister_InvalidKeyChars(t *testing.T) {
	ctx := newTestContext()
	client := models.NewClient()

	payload := `{"local_ip":"127.0.0.1:1234","key":"room with spaces!"}`
	err := register(ctx, client, payload)
	if err == nil {
		t.Error("register() should reject key with invalid characters")
	}
}

func TestRegister_KeyTooLong(t *testing.T) {
	ctx := newTestContext()
	client := models.NewClient()

	longKey := ""
	for i := 0; i < 65; i++ {
		longKey += "a"
	}
	payload := `{"local_ip":"127.0.0.1:1234","key":"` + longKey + `"}`
	err := register(ctx, client, payload)
	if err == nil {
		t.Error("register() should reject key longer than 64 chars")
	}
}

func TestValidateKey_ValidKeys(t *testing.T) {
	validKeys := []string{"room-1", "game_room", "TestRoom", "abc123", "my-game-room_v2"}
	for _, key := range validKeys {
		if err := validateKey(key); err != nil {
			t.Errorf("validateKey(%q) should be valid, got error: %v", key, err)
		}
	}
}
