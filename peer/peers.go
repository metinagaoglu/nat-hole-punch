package peer

import (
	"net"
	"sync"
	"time"
)

// PeerInfo represents a discovered peer
type PeerInfo struct {
	Addr     *net.UDPAddr
	LastSeen time.Time
}

// PeerManager tracks known peers with thread safety
type PeerManager struct {
	mu    sync.RWMutex
	peers map[string]*PeerInfo
}

func newPeerManager() *PeerManager {
	return &PeerManager{
		peers: make(map[string]*PeerInfo),
	}
}

// AddOrUpdate registers a peer or updates last seen time.
// Returns true if this is a newly discovered peer.
func (pm *PeerManager) AddOrUpdate(addr *net.UDPAddr) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	key := addr.String()
	if _, exists := pm.peers[key]; exists {
		pm.peers[key].LastSeen = time.Now()
		return false
	}

	pm.peers[key] = &PeerInfo{
		Addr:     addr,
		LastSeen: time.Now(),
	}
	return true
}

// Remove deletes a peer
func (pm *PeerManager) Remove(addr string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.peers, addr)
}

// GetAll returns all tracked peers
func (pm *PeerManager) GetAll() []*PeerInfo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make([]*PeerInfo, 0, len(pm.peers))
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
