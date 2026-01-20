// Package webrtc peer connection manager.
// This file manages multiple WebRTC peer connections, providing centralized
// connection lifecycle management, metrics aggregation, and resource cleanup.
// Extracted from peer.go for better separation of single-peer vs multi-peer concerns.
package webrtc

import (
	"fmt"
	"sync"
	"time"
)

// Manager manages multiple WebRTC peer connections.
// Code relocated from: peer.go
type Manager struct {
	mu     sync.RWMutex
	peers  map[string]*Peer
	config *Config
}

// NewManager creates a new peer connection manager.
func NewManager(config *Config) *Manager {
	if config == nil {
		config = DefaultConfig()
	}
	return &Manager{
		peers:  make(map[string]*Peer),
		config: config,
	}
}

// CreatePeer creates a new peer with the given ID.
func (m *Manager) CreatePeer(id string) (*Peer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.peers[id]; exists {
		return nil, fmt.Errorf("peer %s already exists", id)
	}

	peer, err := NewPeer(id, m.config)
	if err != nil {
		return nil, err
	}

	m.peers[id] = peer
	return peer, nil
}

// GetPeer retrieves a peer by ID.
func (m *Manager) GetPeer(id string) (*Peer, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	peer, ok := m.peers[id]
	return peer, ok
}

// RemovePeer removes and closes a peer connection.
func (m *Manager) RemovePeer(id string) error {
	m.mu.Lock()
	peer, exists := m.peers[id]
	if exists {
		delete(m.peers, id)
	}
	m.mu.Unlock()

	if exists {
		return peer.Close()
	}
	return fmt.Errorf("peer %s not found", id)
}

// GetMetrics returns aggregated metrics for all peer connections.
func (m *Manager) GetMetrics() ConnectionMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := ConnectionMetrics{}
	var totalRTT time.Duration
	var activeCount int
	var turnCount int

	for _, peer := range m.peers {
		stats := peer.GetStats()
		metrics.TotalConnections++
		if stats.State == StateConnected {
			metrics.ActiveConnections++
			activeCount++
			totalRTT += stats.RTT
			if stats.UsingTURN {
				turnCount++
			}
		} else if stats.State == StateFailed {
			metrics.FailedConnections++
		}
		metrics.TotalBytesSent += stats.BytesSent
		metrics.TotalBytesReceived += stats.BytesReceived
	}

	if activeCount > 0 {
		metrics.AverageRTT = totalRTT / time.Duration(activeCount)
		metrics.TURNUsageRate = float64(turnCount) / float64(activeCount) * 100.0
	}

	return metrics
}

// CloseAll closes all peer connections.
func (m *Manager) CloseAll() error {
	m.mu.Lock()
	peers := make([]*Peer, 0, len(m.peers))
	for _, peer := range m.peers {
		peers = append(peers, peer)
	}
	m.peers = make(map[string]*Peer)
	m.mu.Unlock()

	for _, peer := range peers {
		if err := peer.Close(); err != nil {
			return err
		}
	}
	return nil
}
