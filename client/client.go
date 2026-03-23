package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

const (
	heartbeatInterval = 10 * time.Second
	shutdownTimeout   = 3 * time.Second
	maxBufferSize     = 65535
)

// Client represents a UDP hole punching client
type Client struct {
	conn     *net.UDPConn
	remote   *net.UDPAddr
	local    *net.UDPAddr
	roomKey  string
	peers    *PeerManager
	transfer *FileTransfer
	ctx      context.Context
	cancel   context.CancelFunc
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

	// Get actual bound address (important when using port 0)
	boundAddr := conn.LocalAddr().(*net.UDPAddr)

	log.Printf("Client initialized - Bound: %s, Signal: %s, Room: %s",
		boundAddr.String(), remote.String(), roomKey)

	return &Client{
		conn:     conn,
		remote:   remote,
		local:    boundAddr,
		roomKey:  roomKey,
		peers:    NewPeerManager(),
		transfer: NewFileTransfer(),
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// Register sends registration request to server
func (c *Client) Register() error {
	payload, _ := json.Marshal(SignalRequest{
		LocalIp: c.local.String(),
		Key:     c.roomKey,
	})

	event, _ := json.Marshal(SignalEvent{
		Event:   EventRegister,
		Payload: string(payload),
	})

	log.Printf("Registering to room '%s'...", c.roomKey)

	_, err := c.conn.WriteTo(event, c.remote)
	if err != nil {
		return fmt.Errorf("failed to send register: %w", err)
	}
	return nil
}

// Listen handles incoming UDP messages
func (c *Client) Listen() error {
	log.Println("Listening for incoming messages...")

	buffer := make([]byte, maxBufferSize)
	for {
		select {
		case <-c.ctx.Done():
			return nil
		default:
		}

		c.conn.SetReadDeadline(time.Now().Add(1 * time.Second))

		n, addr, err := c.conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return fmt.Errorf("read error: %w", err)
		}

		c.handleMessage(buffer[:n], addr)
	}
}

// handleMessage routes incoming messages
func (c *Client) handleMessage(data []byte, addr *net.UDPAddr) {
	// Try to parse as peer message first
	msg, err := decodePeerMessage(data)
	if err == nil && msg.Type != "" {
		c.handlePeerMessage(msg, addr)
		return
	}

	// Otherwise treat as peer list from signal server
	c.handlePeerList(string(data), addr)
}

// handlePeerList processes comma-separated peer addresses from the signal server
func (c *Client) handlePeerList(message string, _ *net.UDPAddr) {
	peers := strings.Split(message, ",")
	for _, peerAddr := range peers {
		peerAddr = strings.TrimSpace(peerAddr)
		if peerAddr == "" {
			continue
		}

		// Skip our own address
		if peerAddr == c.local.String() {
			continue
		}

		resolved, err := net.ResolveUDPAddr("udp", peerAddr)
		if err != nil {
			log.Printf("Failed to resolve peer address %s: %v", peerAddr, err)
			continue
		}

		// Only start heartbeat if this is a new peer
		isNew := c.peers.AddOrUpdate(resolved)
		if isNew {
			log.Printf("Discovered new peer: %s", peerAddr)
			go c.heartbeatLoop(resolved)
		}
	}
}

// handlePeerMessage processes structured messages from peers
func (c *Client) handlePeerMessage(msg *PeerMessage, addr *net.UDPAddr) {
	// Update peer tracking
	c.peers.AddOrUpdate(addr)

	switch msg.Type {
	case MsgHeartbeat:
		// silently accept

	case MsgText:
		log.Printf("[MSG] %s: %s", addr, msg.Payload)

	case MsgFile:
		var header FileHeader
		if err := json.Unmarshal([]byte(msg.Payload), &header); err != nil {
			log.Printf("Invalid file header from %s: %v", addr, err)
			return
		}
		c.transfer.HandleFileHeader(&header)

	case MsgFileChunk:
		var chunk FileChunk
		if err := json.Unmarshal([]byte(msg.Payload), &chunk); err != nil {
			log.Printf("Invalid file chunk from %s: %v", addr, err)
			return
		}
		complete, err := c.transfer.HandleFileChunk(&chunk)
		if err != nil {
			log.Printf("File chunk error: %v", err)
			return
		}
		if complete {
			// Send ack
			ack, _ := json.Marshal(FileAck{Name: chunk.Name, Success: true})
			reply, _ := encodePeerMessage(MsgFileAck, c.local.String(), string(ack))
			c.conn.WriteTo(reply, addr)
		}

	case MsgFileAck:
		var ack FileAck
		if err := json.Unmarshal([]byte(msg.Payload), &ack); err == nil {
			log.Printf("File '%s' acknowledged by %s (success: %v)", ack.Name, addr, ack.Success)
		}

	default:
		log.Printf("Unknown message type '%s' from %s", msg.Type, addr)
	}
}

// heartbeatLoop sends periodic heartbeats to a peer
func (c *Client) heartbeatLoop(addr *net.UDPAddr) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	// Send initial heartbeat immediately
	c.sendHeartbeat(addr)

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.sendHeartbeat(addr)
		}
	}
}

func (c *Client) sendHeartbeat(addr *net.UDPAddr) {
	msg, _ := encodePeerMessage(MsgHeartbeat, c.local.String(), "")
	if _, err := c.conn.WriteTo(msg, addr); err != nil {
		log.Printf("Heartbeat failed to %s: %v", addr, err)
	}
}

// SendText sends a text message to all connected peers
func (c *Client) SendText(text string) {
	msg, _ := encodePeerMessage(MsgText, c.local.String(), text)
	peers := c.peers.GetAll()

	if len(peers) == 0 {
		log.Println("No peers connected")
		return
	}

	for _, p := range peers {
		if _, err := c.conn.WriteTo(msg, p.Addr); err != nil {
			log.Printf("Failed to send to %s: %v", p.Addr, err)
		}
	}
	log.Printf("Sent to %d peer(s)", len(peers))
}

// SendFile sends a file to all connected peers
func (c *Client) SendFile(filePath string) {
	peers := c.peers.GetAll()

	if len(peers) == 0 {
		log.Println("No peers connected")
		return
	}

	for _, p := range peers {
		if err := c.transfer.SendFile(c.conn, p.Addr, c.local.String(), filePath); err != nil {
			log.Printf("Failed to send file to %s: %v", p.Addr, err)
		}
	}
}

// Shutdown performs graceful shutdown
func (c *Client) Shutdown() {
	log.Println("Shutting down...")

	// Send logout to signal server
	payload, _ := json.Marshal(SignalRequest{
		LocalIp: c.local.String(),
		Key:     c.roomKey,
	})
	event, _ := json.Marshal(SignalEvent{
		Event:   EventLogout,
		Payload: string(payload),
	})

	if _, err := c.conn.WriteTo(event, c.remote); err != nil {
		log.Printf("Failed to send logout: %v", err)
	} else {
		log.Println("Logout sent")
	}

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
