package models

import (
	"net"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient()

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.remoteAddr != nil {
		t.Error("NewClient() should initialize remoteAddr as nil")
	}

	if client.createAt != 0 {
		t.Errorf("NewClient() createAt = %d, want 0", client.createAt)
	}

	if client.conn != nil {
		t.Error("NewClient() should initialize conn as nil")
	}
}

func TestClient_SetAndGetConn(t *testing.T) {
	client := NewClient()

	// Create a mock UDP connection
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("ResolveUDPAddr error = %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatalf("ListenUDP error = %v", err)
	}
	defer conn.Close()

	// Test SetConn returns self for chaining
	result := client.SetConn(conn)
	if result != client {
		t.Error("SetConn() should return self for method chaining")
	}

	// Test GetConn
	if client.GetConn() != conn {
		t.Error("GetConn() did not return the set connection")
	}
}

func TestClient_SetAndGetRemoteAddr(t *testing.T) {
	client := NewClient()

	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	if err != nil {
		t.Fatalf("ResolveUDPAddr error = %v", err)
	}

	// Test SetRemoteAddr returns self for chaining
	result := client.SetRemoteAddr(addr)
	if result != client {
		t.Error("SetRemoteAddr() should return self for method chaining")
	}

	// Test GetRemoteAddr
	if client.GetRemoteAddr() != addr {
		t.Error("GetRemoteAddr() did not return the set address")
	}
}

func TestClient_SetCreateAt(t *testing.T) {
	client := NewClient()

	before := time.Now().Unix()
	result := client.SetCreateAt()
	after := time.Now().Unix()

	// Test SetCreateAt returns self for chaining
	if result != client {
		t.Error("SetCreateAt() should return self for method chaining")
	}

	// Test GetCreateAt returns a timestamp within the expected range
	createAt := client.GetCreateAt()
	if createAt < before || createAt > after {
		t.Errorf("GetCreateAt() = %d, want between %d and %d", createAt, before, after)
	}
}

func TestClient_CompareAddr(t *testing.T) {
	client := NewClient()

	addr1, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	addr2, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	addr3, _ := net.ResolveUDPAddr("udp", "127.0.0.1:5678")
	addr4, _ := net.ResolveUDPAddr("udp", "192.168.1.1:1234")

	client.SetRemoteAddr(addr1)

	// Same IP and port
	if !client.CompareAddr(addr2) {
		t.Error("CompareAddr() should return true for same IP and port")
	}

	// Same IP, different port
	if client.CompareAddr(addr3) {
		t.Error("CompareAddr() should return false for different port")
	}

	// Different IP, same port
	if client.CompareAddr(addr4) {
		t.Error("CompareAddr() should return false for different IP")
	}
}

func TestClient_MethodChaining(t *testing.T) {
	client := NewClient()

	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	localAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		t.Fatalf("ListenUDP error = %v", err)
	}
	defer conn.Close()

	// Test full method chaining
	result := client.
		SetRemoteAddr(addr).
		SetConn(conn).
		SetCreateAt()

	if result != client {
		t.Error("Method chaining should return the same client instance")
	}

	if client.GetRemoteAddr() != addr {
		t.Error("Chained SetRemoteAddr() did not work")
	}

	if client.GetConn() != conn {
		t.Error("Chained SetConn() did not work")
	}

	if client.GetCreateAt() == 0 {
		t.Error("Chained SetCreateAt() did not work")
	}
}
