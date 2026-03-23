// Package peer provides a P2P client for UDP hole punching.
//
// Usage:
//
//	client, _ := peer.Connect("signal.example.com:3986", "my-room")
//	defer client.Close()
//
//	client.OnMessage(func(from string, text string) {
//	    fmt.Printf("[%s]: %s\n", from, text)
//	})
//
//	client.Send("hello!")
//	client.SendFile("document.pdf")
package peer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/metnagaoglu/holepunch/protocol"
)

const (
	heartbeatInterval = 10 * time.Second
	shutdownTimeout   = 3 * time.Second
	maxBufferSize     = 65535
)

// MessageHandler is called when a text message is received from a peer
type MessageHandler func(from string, text string)

// Client represents a UDP hole punching client
type Client struct {
	conn       *net.UDPConn
	remote     *net.UDPAddr
	local      *net.UDPAddr
	roomKey    string
	peers      *PeerManager
	transfer   *FileTransfer
	onMessage  MessageHandler
	ctx        context.Context
	cancel     context.CancelFunc
}

// Connect creates a client, binds to a local address, and registers with the signal server.
// localAddr can be "0.0.0.0:0" to let the OS pick a free port.
func Connect(signalAddress, localAddress, roomKey string) (*Client, error) {
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
	boundAddr := conn.LocalAddr().(*net.UDPAddr)

	log.Printf("Client initialized - Bound: %s, Signal: %s, Room: %s",
		boundAddr.String(), remote.String(), roomKey)

	c := &Client{
		conn:     conn,
		remote:   remote,
		local:    boundAddr,
		roomKey:  roomKey,
		peers:    newPeerManager(),
		transfer: newFileTransfer(),
		ctx:      ctx,
		cancel:   cancel,
	}

	// Register with signal server
	if err := c.register(); err != nil {
		conn.Close()
		cancel()
		return nil, err
	}

	return c, nil
}

// OnMessage sets a handler for incoming text messages
func (c *Client) OnMessage(handler MessageHandler) {
	c.onMessage = handler
}

// Peers returns the peer manager for inspecting connected peers
func (c *Client) Peers() *PeerManager {
	return c.peers
}

// LocalAddr returns the client's bound local address
func (c *Client) LocalAddr() *net.UDPAddr {
	return c.local
}

// Listen starts the receive loop. Blocks until Shutdown is called.
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

// Send sends a text message to all connected peers
func (c *Client) Send(text string) {
	msg, _ := protocol.EncodePeerMessage(protocol.MsgText, c.local.String(), text)
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

// Shutdown gracefully disconnects from the signal server and stops
func (c *Client) Shutdown() {
	log.Println("Shutting down...")

	payload, _ := json.Marshal(protocol.SignalRequest{
		LocalIP: c.local.String(),
		Key:     c.roomKey,
	})
	event, _ := json.Marshal(protocol.SignalEvent{
		Event:   protocol.EventLogout,
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

func (c *Client) register() error {
	payload, _ := json.Marshal(protocol.SignalRequest{
		LocalIP: c.local.String(),
		Key:     c.roomKey,
	})
	event, _ := json.Marshal(protocol.SignalEvent{
		Event:   protocol.EventRegister,
		Payload: string(payload),
	})

	log.Printf("Registering to room '%s'...", c.roomKey)

	_, err := c.conn.WriteTo(event, c.remote)
	if err != nil {
		return fmt.Errorf("failed to send register: %w", err)
	}
	return nil
}

func (c *Client) handleMessage(data []byte, addr *net.UDPAddr) {
	msg, err := protocol.DecodePeerMessage(data)
	if err == nil && msg.Type != "" {
		c.handlePeerMessage(msg, addr)
		return
	}

	// Peer list from signal server
	c.handlePeerList(string(data))
}

func (c *Client) handlePeerList(message string) {
	peers := strings.Split(message, ",")
	for _, peerAddr := range peers {
		peerAddr = strings.TrimSpace(peerAddr)
		if peerAddr == "" || peerAddr == c.local.String() {
			continue
		}

		resolved, err := net.ResolveUDPAddr("udp", peerAddr)
		if err != nil {
			log.Printf("Failed to resolve peer address %s: %v", peerAddr, err)
			continue
		}

		if isNew := c.peers.AddOrUpdate(resolved); isNew {
			log.Printf("Discovered new peer: %s", peerAddr)
			go c.heartbeatLoop(resolved)
		}
	}
}

func (c *Client) handlePeerMessage(msg *protocol.PeerMessage, addr *net.UDPAddr) {
	c.peers.AddOrUpdate(addr)

	switch msg.Type {
	case protocol.MsgHeartbeat:
		// silently accept

	case protocol.MsgText:
		log.Printf("[MSG] %s: %s", addr, msg.Payload)
		if c.onMessage != nil {
			c.onMessage(msg.From, msg.Payload)
		}

	case protocol.MsgFile:
		var header protocol.FileHeader
		if err := json.Unmarshal([]byte(msg.Payload), &header); err != nil {
			log.Printf("Invalid file header from %s: %v", addr, err)
			return
		}
		c.transfer.handleFileHeader(&header)

	case protocol.MsgFileChunk:
		var chunk protocol.FileChunk
		if err := json.Unmarshal([]byte(msg.Payload), &chunk); err != nil {
			log.Printf("Invalid file chunk from %s: %v", addr, err)
			return
		}
		complete, err := c.transfer.handleFileChunk(&chunk)
		if err != nil {
			log.Printf("File chunk error: %v", err)
			return
		}
		if complete {
			ack, _ := json.Marshal(protocol.FileAck{Name: chunk.Name, Success: true})
			reply, _ := protocol.EncodePeerMessage(protocol.MsgFileAck, c.local.String(), string(ack))
			c.conn.WriteTo(reply, addr)
		}

	case protocol.MsgFileAck:
		var ack protocol.FileAck
		if err := json.Unmarshal([]byte(msg.Payload), &ack); err == nil {
			log.Printf("File '%s' acknowledged by %s (success: %v)", ack.Name, addr, ack.Success)
		}

	default:
		log.Printf("Unknown message type '%s' from %s", msg.Type, addr)
	}
}

func (c *Client) heartbeatLoop(addr *net.UDPAddr) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

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
	msg, _ := protocol.EncodePeerMessage(protocol.MsgHeartbeat, c.local.String(), "")
	if _, err := c.conn.WriteTo(msg, addr); err != nil {
		log.Printf("Heartbeat failed to %s: %v", addr, err)
	}
}
