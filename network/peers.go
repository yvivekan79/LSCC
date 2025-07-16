package network

import "sync"

type PeerManager struct {
    peers map[string]bool
    mu    sync.RWMutex
}

func NewPeerManager() *PeerManager {
    return &PeerManager{
        peers: make(map[string]bool),
    }
}

func (pm *PeerManager) AddPeer(address string) {
    pm.mu.Lock()
    defer pm.mu.Unlock()
    pm.peers[address] = true
}

func (pm *PeerManager) RemovePeer(address string) {
    pm.mu.Lock()
    defer pm.mu.Unlock()
    delete(pm.peers, address)
}

func (pm *PeerManager) ListPeers() []string {
    pm.mu.RLock()
    defer pm.mu.RUnlock()
    peers := make([]string, 0, len(pm.peers))
    for addr := range pm.peers {
        peers = append(peers, addr)
    }
    return peers
}

