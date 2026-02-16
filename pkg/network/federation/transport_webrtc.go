// WebRTC transport adapter for federation layer.
//
// This file implements the GossipTransport and guild.GuildTransport interfaces
// over WebRTC data channels, enabling browser-to-browser (WASM) federation
// without TCP/TLS connections. Wraps the webrtc.Manager for peer lifecycle
// management and provides structured logging for observability.
package federation

import (
	"encoding/json"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/opd-ai/venture/pkg/network/federation/webrtc"
)

// WebRTCTransport provides federation transport over WebRTC data channels.
// It implements GossipTransport for peer discovery and guild.GuildTransport
// for cross-server guild synchronization via browser-compatible P2P connections.
type WebRTCTransport struct {
	mu      sync.RWMutex
	manager *webrtc.Manager
	config  *webrtc.Config
	peerIDs []string // tracked peer IDs for broadcast iteration
}

// NewWebRTCTransport creates a new WebRTC-based federation transport.
// If config is nil, sensible defaults are used.
func NewWebRTCTransport(config *webrtc.Config) *WebRTCTransport {
	if config == nil {
		config = webrtc.DefaultConfig()
	}
	return &WebRTCTransport{
		manager: webrtc.NewManager(config),
		config:  config,
		peerIDs: make([]string, 0),
	}
}

// AddPeer creates a new WebRTC peer and tracks it for broadcast operations.
func (t *WebRTCTransport) AddPeer(peerID string) error {
	if _, err := t.manager.CreatePeer(peerID); err != nil {
		return fmt.Errorf("failed to add WebRTC peer %s: %w", peerID, err)
	}

	t.mu.Lock()
	t.peerIDs = append(t.peerIDs, peerID)
	t.mu.Unlock()

	log.WithFields(log.Fields{
		"peer_id": peerID,
	}).Debug("WebRTC peer added to transport")
	return nil
}

// ConnectPeer establishes a WebRTC connection to a remote peer.
func (t *WebRTCTransport) ConnectPeer(peerID, remotePeerID string) error {
	peer, ok := t.manager.GetPeer(peerID)
	if !ok {
		return fmt.Errorf("WebRTC peer %s not found", peerID)
	}
	if err := peer.Connect(remotePeerID); err != nil {
		return fmt.Errorf("failed to connect peer %s to %s: %w", peerID, remotePeerID, err)
	}
	log.WithFields(log.Fields{
		"peer_id":        peerID,
		"remote_peer_id": remotePeerID,
	}).Info("WebRTC peer connected")
	return nil
}

// RemovePeer disconnects and removes a peer from the transport.
func (t *WebRTCTransport) RemovePeer(peerID string) error {
	if err := t.manager.RemovePeer(peerID); err != nil {
		return err
	}

	t.mu.Lock()
	for i, id := range t.peerIDs {
		if id == peerID {
			t.peerIDs = append(t.peerIDs[:i], t.peerIDs[i+1:]...)
			break
		}
	}
	t.mu.Unlock()
	return nil
}

// SendGossip implements GossipTransport by sending a gossip message to a
// specific peer over its WebRTC data channel.
func (t *WebRTCTransport) SendGossip(peerID string, msg *GossipMessage) error {
	peer, ok := t.manager.GetPeer(peerID)
	if !ok {
		return fmt.Errorf("unknown WebRTC peer: %s", peerID)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal gossip for peer %s: %w", peerID, err)
	}

	if err := peer.Send(data); err != nil {
		log.WithFields(log.Fields{
			"peer_id": peerID,
		}).Warn("failed to send gossip via WebRTC")
		return fmt.Errorf("failed to send gossip to peer %s: %w", peerID, err)
	}
	return nil
}

// BroadcastGuildUpdate implements guild.GuildTransport by broadcasting guild
// state updates to all connected WebRTC peers.
func (t *WebRTCTransport) BroadcastGuildUpdate(guildID string, data []byte) error {
	t.mu.RLock()
	ids := make([]string, len(t.peerIDs))
	copy(ids, t.peerIDs)
	t.mu.RUnlock()

	var lastErr error
	sent := 0
	for _, id := range ids {
		peer, ok := t.manager.GetPeer(id)
		if !ok {
			continue
		}
		if peer.GetState() != webrtc.StateConnected {
			continue
		}
		if err := peer.Send(data); err != nil {
			lastErr = err
			log.WithFields(log.Fields{
				"peer_id":  id,
				"guild_id": guildID,
			}).Warn("failed to broadcast guild update via WebRTC")
			continue
		}
		sent++
	}

	if sent == 0 && lastErr != nil {
		return fmt.Errorf("failed to broadcast guild update for %s: %w", guildID, lastErr)
	}
	return nil
}

// GetPeer returns a peer by ID.
func (t *WebRTCTransport) GetPeer(peerID string) (*webrtc.Peer, bool) {
	return t.manager.GetPeer(peerID)
}

// GetMetrics returns aggregated connection metrics for all WebRTC peers.
func (t *WebRTCTransport) GetMetrics() webrtc.ConnectionMetrics {
	return t.manager.GetMetrics()
}

// ConnectedPeerCount returns the number of currently connected peers.
func (t *WebRTCTransport) ConnectedPeerCount() int {
	return t.manager.GetMetrics().ActiveConnections
}

// Close gracefully closes all WebRTC peer connections.
func (t *WebRTCTransport) Close() error {
	t.mu.Lock()
	t.peerIDs = make([]string, 0)
	t.mu.Unlock()

	return t.manager.CloseAll()
}
