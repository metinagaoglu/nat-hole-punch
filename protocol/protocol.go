// Package protocol defines shared message types for UDP hole punch
// communication between signal servers and peers.
package protocol

import (
	"encoding/json"
	"fmt"
)

// Signal server event types
const (
	EventRegister = "register"
	EventLogout   = "logout"
)

// Peer-to-peer message types
const (
	MsgHeartbeat = "heartbeat"
	MsgText      = "text"
	MsgFile      = "file"
	MsgFileChunk = "file_chunk"
	MsgFileAck   = "file_ack"
)

// SignalEvent is the envelope for signal server communication
type SignalEvent struct {
	Event   string `json:"event"`
	Payload string `json:"payload"`
}

// SignalRequest carries registration/logout data
type SignalRequest struct {
	LocalIP string `json:"local_ip"`
	Key     string `json:"key"`
}

// PeerMessage is the envelope for all peer-to-peer messages
type PeerMessage struct {
	Type    string `json:"type"`
	From    string `json:"from"`
	Payload string `json:"payload,omitempty"`
}

// FileHeader is sent before file transfer begins
type FileHeader struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	TotalChunks int    `json:"total_chunks"`
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

// EncodePeerMessage creates a JSON-encoded peer message
func EncodePeerMessage(msgType, from, payload string) ([]byte, error) {
	return json.Marshal(PeerMessage{
		Type:    msgType,
		From:    from,
		Payload: payload,
	})
}

// DecodePeerMessage parses a JSON-encoded peer message
func DecodePeerMessage(data []byte) (*PeerMessage, error) {
	var msg PeerMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("invalid peer message: %w", err)
	}
	return &msg, nil
}
