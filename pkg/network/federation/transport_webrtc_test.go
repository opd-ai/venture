package federation

import (
	"fmt"
	"testing"

	"github.com/opd-ai/venture/pkg/network/federation/webrtc"
)

func TestNewWebRTCTransport(t *testing.T) {
	tests := []struct {
		name   string
		config *webrtc.Config
	}{
		{"nil config uses defaults", nil},
		{"custom config", webrtc.DefaultConfig()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := NewWebRTCTransport(tt.config)
			if transport == nil {
				t.Fatal("expected non-nil transport")
			}
			if transport.manager == nil {
				t.Fatal("expected non-nil manager")
			}
			if transport.ConnectedPeerCount() != 0 {
				t.Errorf("expected 0 connected peers, got %d", transport.ConnectedPeerCount())
			}
		})
	}
}

func TestWebRTCTransport_AddPeer(t *testing.T) {
	transport := NewWebRTCTransport(nil)
	defer transport.Close()

	if err := transport.AddPeer("peer1"); err != nil {
		t.Fatalf("AddPeer failed: %v", err)
	}

	// Verify peer exists
	peer, ok := transport.GetPeer("peer1")
	if !ok {
		t.Fatal("expected peer1 to exist")
	}
	if peer.ID != "peer1" {
		t.Errorf("expected peer ID peer1, got %s", peer.ID)
	}

	// Verify tracked
	transport.mu.RLock()
	if len(transport.peerIDs) != 1 || transport.peerIDs[0] != "peer1" {
		t.Errorf("expected peerIDs=[peer1], got %v", transport.peerIDs)
	}
	transport.mu.RUnlock()
}

func TestWebRTCTransport_AddPeer_Duplicate(t *testing.T) {
	transport := NewWebRTCTransport(nil)
	defer transport.Close()

	if err := transport.AddPeer("peer1"); err != nil {
		t.Fatalf("first AddPeer failed: %v", err)
	}
	if err := transport.AddPeer("peer1"); err == nil {
		t.Error("expected error for duplicate peer")
	}
}

func TestWebRTCTransport_RemovePeer(t *testing.T) {
	transport := NewWebRTCTransport(nil)
	defer transport.Close()

	if err := transport.AddPeer("peer1"); err != nil {
		t.Fatalf("AddPeer failed: %v", err)
	}

	if err := transport.RemovePeer("peer1"); err != nil {
		t.Fatalf("RemovePeer failed: %v", err)
	}

	// Verify removed
	if _, ok := transport.GetPeer("peer1"); ok {
		t.Error("expected peer1 to be removed")
	}

	// Verify untracked
	transport.mu.RLock()
	if len(transport.peerIDs) != 0 {
		t.Errorf("expected empty peerIDs, got %v", transport.peerIDs)
	}
	transport.mu.RUnlock()
}

func TestWebRTCTransport_RemovePeer_NotFound(t *testing.T) {
	transport := NewWebRTCTransport(nil)
	defer transport.Close()

	if err := transport.RemovePeer("nonexistent"); err == nil {
		t.Error("expected error for nonexistent peer")
	}
}

func TestWebRTCTransport_ConnectPeer(t *testing.T) {
	transport := NewWebRTCTransport(nil)
	defer transport.Close()

	if err := transport.AddPeer("local"); err != nil {
		t.Fatalf("AddPeer failed: %v", err)
	}

	if err := transport.ConnectPeer("local", "remote"); err != nil {
		t.Fatalf("ConnectPeer failed: %v", err)
	}

	// Verify connected
	peer, ok := transport.GetPeer("local")
	if !ok {
		t.Fatal("expected peer to exist")
	}
	if peer.GetState() != webrtc.StateConnected {
		t.Errorf("expected StateConnected, got %s", peer.GetState())
	}
}

func TestWebRTCTransport_ConnectPeer_NotFound(t *testing.T) {
	transport := NewWebRTCTransport(nil)
	defer transport.Close()

	if err := transport.ConnectPeer("nonexistent", "remote"); err == nil {
		t.Error("expected error for nonexistent peer")
	}
}

func TestWebRTCTransport_SendGossip(t *testing.T) {
	transport := NewWebRTCTransport(nil)
	defer transport.Close()

	if err := transport.AddPeer("peer1"); err != nil {
		t.Fatalf("AddPeer failed: %v", err)
	}
	if err := transport.ConnectPeer("peer1", "remote1"); err != nil {
		t.Fatalf("ConnectPeer failed: %v", err)
	}

	msg := &GossipMessage{
		MessageID: "test-msg-1",
		Servers:   []DiscoveryPacket{},
		Timestamp: 1234567890,
		MaxHops:   3,
		OriginID:  "origin-server",
	}

	if err := transport.SendGossip("peer1", msg); err != nil {
		t.Fatalf("SendGossip failed: %v", err)
	}

	// Verify stats updated
	peer, _ := transport.GetPeer("peer1")
	stats := peer.GetStats()
	if stats.MessagesSent == 0 {
		t.Error("expected MessagesSent > 0")
	}
}

func TestWebRTCTransport_SendGossip_UnknownPeer(t *testing.T) {
	transport := NewWebRTCTransport(nil)
	defer transport.Close()

	msg := &GossipMessage{MessageID: "test"}
	if err := transport.SendGossip("nonexistent", msg); err == nil {
		t.Error("expected error for unknown peer")
	}
}

func TestWebRTCTransport_SendGossip_NotConnected(t *testing.T) {
	transport := NewWebRTCTransport(nil)
	defer transport.Close()

	if err := transport.AddPeer("peer1"); err != nil {
		t.Fatalf("AddPeer failed: %v", err)
	}

	msg := &GossipMessage{MessageID: "test"}
	if err := transport.SendGossip("peer1", msg); err == nil {
		t.Error("expected error for unconnected peer")
	}
}

func TestWebRTCTransport_BroadcastGuildUpdate(t *testing.T) {
	transport := NewWebRTCTransport(nil)
	defer transport.Close()

	// Add and connect two peers
	for _, id := range []string{"peer1", "peer2"} {
		if err := transport.AddPeer(id); err != nil {
			t.Fatalf("AddPeer %s failed: %v", id, err)
		}
		if err := transport.ConnectPeer(id, "remote-"+id); err != nil {
			t.Fatalf("ConnectPeer %s failed: %v", id, err)
		}
	}

	data := []byte(`{"guild_id":"g1","type":"guild_sync"}`)
	if err := transport.BroadcastGuildUpdate("g1", data); err != nil {
		t.Fatalf("BroadcastGuildUpdate failed: %v", err)
	}

	// Verify both peers got the message
	for _, id := range []string{"peer1", "peer2"} {
		peer, _ := transport.GetPeer(id)
		stats := peer.GetStats()
		if stats.MessagesSent == 0 {
			t.Errorf("peer %s: expected MessagesSent > 0", id)
		}
	}
}

func TestWebRTCTransport_BroadcastGuildUpdate_NoPeers(t *testing.T) {
	transport := NewWebRTCTransport(nil)
	defer transport.Close()

	// No connected peers should return nil
	if err := transport.BroadcastGuildUpdate("g1", []byte("data")); err != nil {
		t.Errorf("expected nil error with no peers, got %v", err)
	}
}

func TestWebRTCTransport_BroadcastGuildUpdate_SkipsDisconnected(t *testing.T) {
	transport := NewWebRTCTransport(nil)
	defer transport.Close()

	// Add peer but don't connect
	if err := transport.AddPeer("peer1"); err != nil {
		t.Fatalf("AddPeer failed: %v", err)
	}
	// Add and connect another
	if err := transport.AddPeer("peer2"); err != nil {
		t.Fatalf("AddPeer failed: %v", err)
	}
	if err := transport.ConnectPeer("peer2", "remote2"); err != nil {
		t.Fatalf("ConnectPeer failed: %v", err)
	}

	data := []byte(`{"guild_id":"g1"}`)
	if err := transport.BroadcastGuildUpdate("g1", data); err != nil {
		t.Fatalf("BroadcastGuildUpdate failed: %v", err)
	}

	// peer1 (not connected) should have 0 messages
	peer1, _ := transport.GetPeer("peer1")
	if peer1.GetStats().MessagesSent != 0 {
		t.Error("disconnected peer1 should not have received broadcast")
	}

	// peer2 (connected) should have 1+ messages
	peer2, _ := transport.GetPeer("peer2")
	if peer2.GetStats().MessagesSent == 0 {
		t.Error("connected peer2 should have received broadcast")
	}
}

func TestWebRTCTransport_GetMetrics(t *testing.T) {
	transport := NewWebRTCTransport(nil)
	defer transport.Close()

	if err := transport.AddPeer("peer1"); err != nil {
		t.Fatalf("AddPeer failed: %v", err)
	}
	if err := transport.ConnectPeer("peer1", "remote1"); err != nil {
		t.Fatalf("ConnectPeer failed: %v", err)
	}

	metrics := transport.GetMetrics()
	if metrics.ActiveConnections != 1 {
		t.Errorf("expected 1 active connection, got %d", metrics.ActiveConnections)
	}
	if metrics.TotalConnections != 1 {
		t.Errorf("expected 1 total connection, got %d", metrics.TotalConnections)
	}
}

func TestWebRTCTransport_ConnectedPeerCount(t *testing.T) {
	transport := NewWebRTCTransport(nil)
	defer transport.Close()

	if transport.ConnectedPeerCount() != 0 {
		t.Error("expected 0 initially")
	}

	if err := transport.AddPeer("peer1"); err != nil {
		t.Fatalf("AddPeer failed: %v", err)
	}
	if transport.ConnectedPeerCount() != 0 {
		t.Error("expected 0 before connect")
	}

	if err := transport.ConnectPeer("peer1", "remote1"); err != nil {
		t.Fatalf("ConnectPeer failed: %v", err)
	}
	if transport.ConnectedPeerCount() != 1 {
		t.Errorf("expected 1 after connect, got %d", transport.ConnectedPeerCount())
	}
}

func TestWebRTCTransport_Close(t *testing.T) {
	transport := NewWebRTCTransport(nil)

	if err := transport.AddPeer("peer1"); err != nil {
		t.Fatalf("AddPeer failed: %v", err)
	}
	if err := transport.ConnectPeer("peer1", "remote1"); err != nil {
		t.Fatalf("ConnectPeer failed: %v", err)
	}

	if err := transport.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify peer IDs cleared
	transport.mu.RLock()
	if len(transport.peerIDs) != 0 {
		t.Errorf("expected empty peerIDs after close, got %v", transport.peerIDs)
	}
	transport.mu.RUnlock()

	// Verify metrics show no connections
	if transport.ConnectedPeerCount() != 0 {
		t.Error("expected 0 connections after close")
	}
}

func TestWebRTCTransport_MultiplePeers(t *testing.T) {
	transport := NewWebRTCTransport(nil)
	defer transport.Close()

	peerCount := 5
	for i := 0; i < peerCount; i++ {
		id := fmt.Sprintf("peer%d", i)
		if err := transport.AddPeer(id); err != nil {
			t.Fatalf("AddPeer %s failed: %v", id, err)
		}
		if err := transport.ConnectPeer(id, "remote-"+id); err != nil {
			t.Fatalf("ConnectPeer %s failed: %v", id, err)
		}
	}

	if transport.ConnectedPeerCount() != peerCount {
		t.Errorf("expected %d connected peers, got %d", peerCount, transport.ConnectedPeerCount())
	}

	// Remove one
	if err := transport.RemovePeer("peer2"); err != nil {
		t.Fatalf("RemovePeer failed: %v", err)
	}

	transport.mu.RLock()
	if len(transport.peerIDs) != peerCount-1 {
		t.Errorf("expected %d tracked peers, got %d", peerCount-1, len(transport.peerIDs))
	}
	transport.mu.RUnlock()
}

// TestWebRTCTransport_GossipTransportInterface verifies interface satisfaction.
func TestWebRTCTransport_GossipTransportInterface(t *testing.T) {
	var _ GossipTransport = (*WebRTCTransport)(nil)
}
