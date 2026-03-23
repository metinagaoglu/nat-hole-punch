package main

import (
	"encoding/json"
	"fmt"
)

// Message types for peer-to-peer communication
const (
	MsgHeartbeat = "heartbeat"
	MsgText      = "text"
	MsgFile      = "file"      // File metadata (name, size, chunks)
	MsgFileChunk = "file_chunk" // Individual file chunk
	MsgFileAck   = "file_ack"  // Acknowledge file receipt
)

// PeerMessage is the envelope for all peer-to-peer messages
type PeerMessage struct {
	Type    string `json:"type"`
	From    string `json:"from"`
	Payload string `json:"payload,omitempty"`
}

// FileHeader is sent before file transfer begins
type FileHeader struct {
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	TotalChunks int   `json:"total_chunks"`
}

// FileChunk represents a piece of a file being transferred
type FileChunk struct {
	Name  string `json:"name"`
	Index int    `json:"index"`
	Total int    `json:"total"`
	Data  string `json:"data"` // base64 encoded
}

// FileAck acknowledges receipt of a complete file
type FileAck struct {
	Name    string `json:"name"`
	Success bool   `json:"success"`
}

// Signal server event types (for register/logout)
const (
	EventRegister = "register"
	EventLogout   = "logout"
)

// SignalEvent is the envelope for signal server communication
type SignalEvent struct {
	Event   string `json:"event"`
	Payload string `json:"payload"`
}

// SignalRequest carries registration/logout data
type SignalRequest struct {
	LocalIp string `json:"local_ip"`
	Key     string `json:"key"`
}

func encodePeerMessage(msgType, from, payload string) ([]byte, error) {
	return json.Marshal(PeerMessage{
		Type:    msgType,
		From:    from,
		Payload: payload,
	})
}

func decodePeerMessage(data []byte) (*PeerMessage, error) {
	var msg PeerMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("invalid peer message: %w", err)
	}
	return &msg, nil
}
