package main

import (
	"log"
	"net"
	"sync"
	"time"
)

// PeerManager tracks known peers and prevents duplicate connections
type PeerManager struct {
	mu    sync.RWMutex
	peers map[string]*Peer
}

// Peer represents a known peer in the network
type Peer struct {
	Addr     *net.UDPAddr
	LastSeen time.Time
}

func NewPeerManager() *PeerManager {
	return &PeerManager{
		peers: make(map[string]*Peer),
	}
}

// AddOrUpdate registers a peer or updates its last seen time.
// Returns true if this is a newly discovered peer.
func (pm *PeerManager) AddOrUpdate(addr *net.UDPAddr) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	key := addr.String()
	if _, exists := pm.peers[key]; exists {
		pm.peers[key].LastSeen = time.Now()
		return false
	}

	pm.peers[key] = &Peer{
		Addr:     addr,
		LastSeen: time.Now(),
	}
	return true
}

// Remove deletes a peer from tracking
func (pm *PeerManager) Remove(addr string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.peers, addr)
}

// GetAll returns all tracked peers
func (pm *PeerManager) GetAll() []*Peer {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make([]*Peer, 0, len(pm.peers))
	for _, p := range pm.peers {
		result = append(result, p)
	}
	return result
}

// Count returns the number of tracked peers
func (pm *PeerManager) Count() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.peers)
}

// PrintPeers logs all known peers
func (pm *PeerManager) PrintPeers() {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if len(pm.peers) == 0 {
		log.Println("No peers connected")
		return
	}

	log.Printf("Connected peers (%d):", len(pm.peers))
	for addr, p := range pm.peers {
		log.Printf("  - %s (last seen: %s ago)", addr, time.Since(p.LastSeen).Round(time.Second))
	}
}
