package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	// Default configuration values
	defaultSignalAddress = "127.0.0.1:3986"
	defaultLocalAddress  = "127.0.0.1:4000"
	defaultRoomKey       = "default"

	// Protocol constants
	messageHello         = "Hello!"
	eventRegister        = "register"
	eventLogout          = "logout"

	// Timing constants
	heartbeatInterval    = 5 * time.Second
	shutdownTimeout      = 3 * time.Second
)

type RegisterRequest struct {
	LocalIp string `json:"local_ip"`
	Key     string `json:"key"`
}

type LogoutRequest struct {
	LocalIp string `json:"local_ip"`
	Key     string `json:"key"`
}

type Event struct {
	Event   string `json:"event"`
	Payload string `json:"payload"`
}

// Client represents a UDP hole punching client
type Client struct {
	conn          *net.UDPConn
	remote        *net.UDPAddr
	local         *net.UDPAddr
	roomKey       string
	ctx           context.Context
	cancel        context.CancelFunc
}

func main() {
	// Parse command-line flags
	signalAddressPtr := flag.String("signal-address", defaultSignalAddress, "Signal server address")
	localAddressPtr := flag.String("local-address", defaultLocalAddress, "Local address")
	roomKeyAddressPtr := flag.String("room-key", defaultRoomKey, "Room key")
	flag.Parse()

	// Create client
	client, err := NewClient(*signalAddressPtr, *localAddressPtr, *roomKeyAddressPtr)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Register with server
	if err := client.Register(); err != nil {
		log.Fatalf("Failed to register: %v", err)
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start listening in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- client.Listen()
	}()

	// Wait for shutdown signal or error
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v, shutting down gracefully...", sig)
		client.Shutdown()
	case err := <-errChan:
		if err != nil {
			log.Printf("Listen error: %v", err)
		}
	}
}

// NewClient creates a new UDP client
func NewClient(signalAddress, localAddress, roomKey string) (*Client, error) {
	remote, err := net.ResolveUDPAddr("udp", signalAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve signal address: %w", err)
	}

	local, err := net.ResolveUDPAddr("udp", localAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve local address: %w", err)
	}

	conn, err := net.ListenUDP("udp", local)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on UDP: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	log.Printf("Client initialized - Local: %s, Signal: %s, Room: %s",
		local.String(), remote.String(), roomKey)

	return &Client{
		conn:    conn,
		remote:  remote,
		local:   local,
		roomKey: roomKey,
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}

// Register sends registration request to server
func (c *Client) Register() error {
	register, err := json.Marshal(RegisterRequest{
		LocalIp: c.local.String(),
		Key:     c.roomKey,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal register request: %w", err)
	}

	jsonRegister, err := json.Marshal(Event{
		Event:   eventRegister,
		Payload: string(register),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal register event: %w", err)
	}

	log.Printf("Registering to room '%s'...", c.roomKey)

	n, err := c.conn.WriteTo(jsonRegister, c.remote)
	if err != nil {
		return fmt.Errorf("failed to send register: %w", err)
	}

	log.Printf("Registration sent (%d bytes)", n)
	return nil
}

// Listen handles incoming UDP messages
func (c *Client) Listen() error {
	log.Println("Listening for incoming messages...")

	buffer := make([]byte, 1024)
	for {
		select {
		case <-c.ctx.Done():
			return nil
		default:
			// Set read deadline to allow context cancellation
			c.conn.SetReadDeadline(time.Now().Add(1 * time.Second))

			n, addr, err := c.conn.ReadFromUDP(buffer)
			if err != nil {
				// Check if it's a timeout (expected for graceful shutdown)
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				return fmt.Errorf("read error: %w", err)
			}

			message := string(buffer[:n])
			c.handleMessage(message, addr)
		}
	}
}

// handleMessage processes incoming messages
func (c *Client) handleMessage(message string, addr *net.UDPAddr) {
	log.Printf("[INCOMING] From %s: %s", addr.String(), message)

	// Handle heartbeat message
	if message == messageHello {
		log.Printf("Received heartbeat from %s", addr.String())
		return
	}

	// Handle peer list from server
	peers := strings.Split(message, ",")
	for _, peerAddr := range peers {
		peerAddr = strings.TrimSpace(peerAddr)
		if peerAddr == "" || peerAddr == c.local.String() {
			continue
		}

		log.Printf("Discovered peer: %s, starting communication...", peerAddr)
		go c.startPeerCommunication(peerAddr)
	}
}

// startPeerCommunication initiates communication with a peer
func (c *Client) startPeerCommunication(peerAddress string) {
	addr, err := net.ResolveUDPAddr("udp", peerAddress)
	if err != nil {
		log.Printf("Failed to resolve peer address %s: %v", peerAddress, err)
		return
	}

	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			n, err := c.conn.WriteTo([]byte(messageHello), addr)
			if err != nil {
				log.Printf("Failed to send to peer %s: %v", peerAddress, err)
				return
			}
			log.Printf("Sent heartbeat to %s (%d bytes)", peerAddress, n)
		}
	}
}

// Shutdown performs graceful shutdown
func (c *Client) Shutdown() {
	log.Println("Initiating shutdown...")

	// Send logout request
	logout, err := json.Marshal(LogoutRequest{
		LocalIp: c.local.String(),
		Key:     c.roomKey,
	})
	if err != nil {
		log.Printf("Failed to marshal logout request: %v", err)
	} else {
		jsonLogout, err := json.Marshal(Event{
			Event:   eventLogout,
			Payload: string(logout),
		})
		if err != nil {
			log.Printf("Failed to marshal logout event: %v", err)
		} else {
			n, err := c.conn.WriteTo(jsonLogout, c.remote)
			if err != nil {
				log.Printf("Failed to send logout: %v", err)
			} else {
				log.Printf("Logout sent (%d bytes)", n)
			}
		}
	}

	// Cancel context and wait for goroutines
	c.cancel()
	time.Sleep(shutdownTimeout)

	log.Println("Shutdown complete")
}

// Close closes the UDP connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
