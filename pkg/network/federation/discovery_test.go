package federation

import (
	"encoding/json"
	"fmt"
	"net"
	"testing"
	"time"
)

func TestNewDiscoverySystem(t *testing.T) {
	identity, err := NewServerIdentity("TestServer")
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	tests := []struct {
		name       string
		identity   *ServerIdentity
		listenAddr string
		wantErr    bool
	}{
		{
			name:       "valid configuration",
			identity:   identity,
			listenAddr: ":0", // Random port
			wantErr:    false,
		},
		{
			name:       "default listen address",
			identity:   identity,
			listenAddr: "",
			wantErr:    false,
		},
		{
			name:       "nil identity",
			identity:   nil,
			listenAddr: ":8090",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds, err := NewDiscoverySystem(tt.identity, tt.listenAddr)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDiscoverySystem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && ds == nil {
				t.Error("Expected non-nil discovery system")
			}
		})
	}
}

func TestDiscoverySystem_StartStop(t *testing.T) {
	identity, err := NewServerIdentity("TestServer")
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	ds, err := NewDiscoverySystem(identity, ":0") // Random port
	if err != nil {
		t.Fatalf("Failed to create discovery system: %v", err)
	}

	// Test start
	err = ds.Start()
	if err != nil {
		t.Fatalf("Failed to start discovery system: %v", err)
	}

	// Verify running
	ds.mu.RLock()
	running := ds.running
	ds.mu.RUnlock()
	if !running {
		t.Error("Discovery system should be running after Start()")
	}

	// Test double start
	err = ds.Start()
	if err == nil {
		t.Error("Expected error when starting already running system")
	}

	// Test stop
	err = ds.Stop()
	if err != nil {
		t.Fatalf("Failed to stop discovery system: %v", err)
	}

	// Verify stopped
	ds.mu.RLock()
	running = ds.running
	ds.mu.RUnlock()
	if running {
		t.Error("Discovery system should not be running after Stop()")
	}

	// Test double stop
	err = ds.Stop()
	if err == nil {
		t.Error("Expected error when stopping already stopped system")
	}
}

func TestDiscoverySystem_ProcessPacket(t *testing.T) {
	identity, err := NewServerIdentity("TestServer")
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	ds, err := NewDiscoverySystem(identity, ":0")
	if err != nil {
		t.Fatalf("Failed to create discovery system: %v", err)
	}

	// Create a test packet from different server
	otherIdentity, _ := NewServerIdentity("OtherServer")
	packet := DiscoveryPacket{
		ServerID:   otherIdentity.ServerID,
		ServerName: "OtherServer",
		Address:    "192.168.1.100:8080",
		Version:    "6.0.0",
		Features:   []string{"travel", "trade"},
		Timestamp:  time.Now().UnixMilli(),
		Hops:       0,
	}

	data, err := json.Marshal(packet)
	if err != nil {
		t.Fatalf("Failed to marshal packet: %v", err)
	}

	// Process packet
	addr := &net.UDPAddr{IP: net.ParseIP("192.168.1.100"), Port: 8090}
	ds.processPacket(data, addr)

	// Verify peer was added
	peers := ds.GetPeers()
	if len(peers) != 1 {
		t.Fatalf("Expected 1 peer, got %d", len(peers))
	}

	peer := peers[0]
	if peer.ServerID != otherIdentity.ServerID {
		t.Errorf("Expected ServerID %s, got %s", otherIdentity.ServerID, peer.ServerID)
	}
	if peer.ServerName != "OtherServer" {
		t.Errorf("Expected ServerName 'OtherServer', got %s", peer.ServerName)
	}
	if peer.Address != "192.168.1.100:8080" {
		t.Errorf("Expected Address '192.168.1.100:8080', got %s", peer.Address)
	}
}

func TestDiscoverySystem_IgnoreOwnPackets(t *testing.T) {
	identity, err := NewServerIdentity("TestServer")
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	ds, err := NewDiscoverySystem(identity, ":0")
	if err != nil {
		t.Fatalf("Failed to create discovery system: %v", err)
	}

	// Create packet from same server
	packet := DiscoveryPacket{
		ServerID:   identity.ServerID,
		ServerName: "TestServer",
		Address:    "localhost:8080",
		Version:    "6.0.0",
		Features:   []string{"travel"},
		Timestamp:  time.Now().UnixMilli(),
		Hops:       0,
	}

	data, err := json.Marshal(packet)
	if err != nil {
		t.Fatalf("Failed to marshal packet: %v", err)
	}

	// Process packet
	addr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8090}
	ds.processPacket(data, addr)

	// Verify no peers were added
	peers := ds.GetPeers()
	if len(peers) != 0 {
		t.Errorf("Expected 0 peers (should ignore own packets), got %d", len(peers))
	}
}

func TestDiscoverySystem_TimestampValidation(t *testing.T) {
	identity, err := NewServerIdentity("TestServer")
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	ds, err := NewDiscoverySystem(identity, ":0")
	if err != nil {
		t.Fatalf("Failed to create discovery system: %v", err)
	}

	tests := []struct {
		name          string
		timestampDiff int64 // Milliseconds from now
		shouldAdd     bool
	}{
		{
			name:          "current timestamp",
			timestampDiff: 0,
			shouldAdd:     true,
		},
		{
			name:          "30 seconds in past",
			timestampDiff: -30000,
			shouldAdd:     true,
		},
		{
			name:          "30 seconds in future",
			timestampDiff: 30000,
			shouldAdd:     true,
		},
		{
			name:          "2 minutes in past (too old)",
			timestampDiff: -120000,
			shouldAdd:     false,
		},
		{
			name:          "2 minutes in future (too far)",
			timestampDiff: 120000,
			shouldAdd:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otherIdentity, _ := NewServerIdentity("OtherServer")
			packet := DiscoveryPacket{
				ServerID:   otherIdentity.ServerID,
				ServerName: "OtherServer",
				Address:    "192.168.1.100:8080",
				Version:    "6.0.0",
				Features:   []string{"travel"},
				Timestamp:  time.Now().UnixMilli() + tt.timestampDiff,
				Hops:       0,
			}

			data, _ := json.Marshal(packet)
			addr := &net.UDPAddr{IP: net.ParseIP("192.168.1.100"), Port: 8090}

			// Clear existing peers
			ds.mu.Lock()
			ds.knownPeers = make(map[string]*DiscoveredPeer)
			ds.mu.Unlock()

			ds.processPacket(data, addr)

			peers := ds.GetPeers()
			added := len(peers) > 0

			if added != tt.shouldAdd {
				t.Errorf("Expected shouldAdd=%v, got added=%v", tt.shouldAdd, added)
			}
		})
	}
}

func TestDiscoverySystem_GetPeers(t *testing.T) {
	identity, err := NewServerIdentity("TestServer")
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	ds, err := NewDiscoverySystem(identity, ":0")
	if err != nil {
		t.Fatalf("Failed to create discovery system: %v", err)
	}

	// Add multiple peers
	for i := 0; i < 5; i++ {
		peer := &DiscoveredPeer{
			ServerID:   fmt.Sprintf("server-%d", i),
			ServerName: fmt.Sprintf("Server %d", i),
			Address:    fmt.Sprintf("192.168.1.%d:8080", 100+i),
			Version:    "6.0.0",
			Features:   []string{"travel", "trade"},
			LastSeen:   time.Now(),
			Hops:       0,
		}
		ds.mu.Lock()
		ds.knownPeers[peer.ServerID] = peer
		ds.mu.Unlock()
	}

	// Get all peers
	peers := ds.GetPeers()
	if len(peers) != 5 {
		t.Errorf("Expected 5 peers, got %d", len(peers))
	}

	// Verify deep copy (modifying returned slice shouldn't affect internal state)
	peers[0].ServerName = "Modified"

	ds.mu.RLock()
	original := ds.knownPeers["server-0"].ServerName
	ds.mu.RUnlock()

	if original == "Modified" {
		t.Error("GetPeers() should return deep copy, not reference to internal data")
	}
}

func TestDiscoverySystem_GetPeer(t *testing.T) {
	identity, err := NewServerIdentity("TestServer")
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	ds, err := NewDiscoverySystem(identity, ":0")
	if err != nil {
		t.Fatalf("Failed to create discovery system: %v", err)
	}

	// Add a test peer
	testPeer := &DiscoveredPeer{
		ServerID:   "test-server-123",
		ServerName: "Test Server",
		Address:    "192.168.1.100:8080",
		Version:    "6.0.0",
		Features:   []string{"travel", "trade"},
		LastSeen:   time.Now(),
		Hops:       0,
	}
	ds.mu.Lock()
	ds.knownPeers[testPeer.ServerID] = testPeer
	ds.mu.Unlock()

	// Test getting existing peer
	peer, exists := ds.GetPeer("test-server-123")
	if !exists {
		t.Fatal("Expected peer to exist")
	}
	if peer.ServerID != "test-server-123" {
		t.Errorf("Expected ServerID 'test-server-123', got %s", peer.ServerID)
	}

	// Test getting non-existent peer
	_, exists = ds.GetPeer("non-existent")
	if exists {
		t.Error("Expected peer not to exist")
	}
}

func TestDiscoverySystem_AddManualPeer(t *testing.T) {
	identity, err := NewServerIdentity("TestServer")
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	ds, err := NewDiscoverySystem(identity, ":0")
	if err != nil {
		t.Fatalf("Failed to create discovery system: %v", err)
	}

	tests := []struct {
		name       string
		serverID   string
		serverName string
		address    string
		version    string
		features   []string
		wantErr    bool
	}{
		{
			name:       "valid peer",
			serverID:   "manual-server-1",
			serverName: "Manual Server",
			address:    "192.168.1.200:8080",
			version:    "6.0.0",
			features:   []string{"travel", "trade", "post"},
			wantErr:    false,
		},
		{
			name:       "empty serverID",
			serverID:   "",
			serverName: "Manual Server",
			address:    "192.168.1.200:8080",
			version:    "6.0.0",
			features:   []string{"travel"},
			wantErr:    true,
		},
		{
			name:       "empty address",
			serverID:   "manual-server-2",
			serverName: "Manual Server",
			address:    "",
			version:    "6.0.0",
			features:   []string{"travel"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ds.AddManualPeer(tt.serverID, tt.serverName, tt.address, tt.version, tt.features)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddManualPeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				peer, exists := ds.GetPeer(tt.serverID)
				if !exists {
					t.Error("Expected manually added peer to exist")
				}
				if peer.ServerID != tt.serverID {
					t.Errorf("Expected ServerID %s, got %s", tt.serverID, peer.ServerID)
				}
				if peer.Hops != 0 {
					t.Errorf("Manual peers should have Hops=0, got %d", peer.Hops)
				}
			}
		})
	}
}

func TestDiscoverySystem_RemovePeer(t *testing.T) {
	identity, err := NewServerIdentity("TestServer")
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	ds, err := NewDiscoverySystem(identity, ":0")
	if err != nil {
		t.Fatalf("Failed to create discovery system: %v", err)
	}

	// Add a peer
	ds.AddManualPeer("test-peer", "Test Peer", "192.168.1.100:8080", "6.0.0", []string{"travel"})

	// Verify peer exists
	_, exists := ds.GetPeer("test-peer")
	if !exists {
		t.Fatal("Peer should exist before removal")
	}

	// Remove peer
	ds.RemovePeer("test-peer")

	// Verify peer removed
	_, exists = ds.GetPeer("test-peer")
	if exists {
		t.Error("Peer should not exist after removal")
	}
}

func TestDiscoverySystem_CleanupStalePeers(t *testing.T) {
	identity, err := NewServerIdentity("TestServer")
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	ds, err := NewDiscoverySystem(identity, ":0")
	if err != nil {
		t.Fatalf("Failed to create discovery system: %v", err)
	}

	// Add a fresh peer
	freshPeer := &DiscoveredPeer{
		ServerID:   "fresh-peer",
		ServerName: "Fresh Peer",
		Address:    "192.168.1.100:8080",
		Version:    "6.0.0",
		Features:   []string{"travel"},
		LastSeen:   time.Now(),
		Hops:       0,
	}

	// Add a stale peer (old LastSeen)
	stalePeer := &DiscoveredPeer{
		ServerID:   "stale-peer",
		ServerName: "Stale Peer",
		Address:    "192.168.1.101:8080",
		Version:    "6.0.0",
		Features:   []string{"travel"},
		LastSeen:   time.Now().Add(-2 * DiscoveryTimeout), // Very old
		Hops:       0,
	}

	ds.mu.Lock()
	ds.knownPeers[freshPeer.ServerID] = freshPeer
	ds.knownPeers[stalePeer.ServerID] = stalePeer
	ds.mu.Unlock()

	// Run cleanup
	ds.cleanupStalePeers()

	// Verify fresh peer still exists
	_, exists := ds.GetPeer("fresh-peer")
	if !exists {
		t.Error("Fresh peer should still exist after cleanup")
	}

	// Verify stale peer removed
	_, exists = ds.GetPeer("stale-peer")
	if exists {
		t.Error("Stale peer should be removed after cleanup")
	}
}

func TestDiscoverySystem_OnPeerDiscovered(t *testing.T) {
	identity, err := NewServerIdentity("TestServer")
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	ds, err := NewDiscoverySystem(identity, ":0")
	if err != nil {
		t.Fatalf("Failed to create discovery system: %v", err)
	}

	// Set callback
	callbackCalled := make(chan *DiscoveredPeer, 1)
	ds.OnPeerDiscovered(func(peer *DiscoveredPeer) {
		callbackCalled <- peer
	})

	// Add a peer
	ds.AddManualPeer("callback-test", "Callback Test", "192.168.1.100:8080", "6.0.0", []string{"travel"})

	// Wait for callback
	select {
	case peer := <-callbackCalled:
		if peer.ServerID != "callback-test" {
			t.Errorf("Expected callback with ServerID 'callback-test', got %s", peer.ServerID)
		}
	case <-time.After(1 * time.Second):
		t.Error("Callback was not called within timeout")
	}
}

// Benchmark tests
func BenchmarkDiscoverySystem_ProcessPacket(b *testing.B) {
	identity, _ := NewServerIdentity("TestServer")
	ds, _ := NewDiscoverySystem(identity, ":0")

	otherIdentity, _ := NewServerIdentity("OtherServer")
	packet := DiscoveryPacket{
		ServerID:   otherIdentity.ServerID,
		ServerName: "OtherServer",
		Address:    "192.168.1.100:8080",
		Version:    "6.0.0",
		Features:   []string{"travel", "trade"},
		Timestamp:  time.Now().UnixMilli(),
		Hops:       0,
	}

	data, _ := json.Marshal(packet)
	addr := &net.UDPAddr{IP: net.ParseIP("192.168.1.100"), Port: 8090}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ds.processPacket(data, addr)
	}
}

func BenchmarkDiscoverySystem_GetPeers(b *testing.B) {
	identity, _ := NewServerIdentity("TestServer")
	ds, _ := NewDiscoverySystem(identity, ":0")

	// Add 100 peers
	for i := 0; i < 100; i++ {
		peer := &DiscoveredPeer{
			ServerID:   fmt.Sprintf("server-%d", i),
			ServerName: fmt.Sprintf("Server %d", i),
			Address:    fmt.Sprintf("192.168.1.%d:8080", i),
			Version:    "6.0.0",
			Features:   []string{"travel", "trade"},
			LastSeen:   time.Now(),
			Hops:       0,
		}
		ds.mu.Lock()
		ds.knownPeers[peer.ServerID] = peer
		ds.mu.Unlock()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ds.GetPeers()
	}
}

func BenchmarkDiscoverySystem_AddManualPeer(b *testing.B) {
	identity, _ := NewServerIdentity("TestServer")
	ds, _ := NewDiscoverySystem(identity, ":0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ds.AddManualPeer(
			fmt.Sprintf("server-%d", i),
			"Test Server",
			"192.168.1.100:8080",
			"6.0.0",
			[]string{"travel", "trade"},
		)
	}
}
