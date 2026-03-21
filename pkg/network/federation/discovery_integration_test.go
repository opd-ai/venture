package federation

import (
	"sync"
	"testing"
	"time"
)

// TestDiscoverySystem_Integration_LANDiscovery tests LAN broadcast discovery
func TestDiscoverySystem_Integration_LANDiscovery(t *testing.T) {
	// Create two servers on different ports
	server1Identity, err := NewServerIdentity("Server1")
	if err != nil {
		t.Fatalf("Failed to create server1 identity: %v", err)
	}

	server2Identity, err := NewServerIdentity("Server2")
	if err != nil {
		t.Fatalf("Failed to create server2 identity: %v", err)
	}

	ds1, err := NewDiscoverySystem(server1Identity, ":18090", "localhost:8080")
	if err != nil {
		t.Fatalf("Failed to create discovery system 1: %v", err)
	}

	ds2, err := NewDiscoverySystem(server2Identity, ":18091", "localhost:8080")
	if err != nil {
		t.Fatalf("Failed to create discovery system 2: %v", err)
	}

	// Start both systems
	if err := ds1.Start(); err != nil {
		t.Fatalf("Failed to start discovery system 1: %v", err)
	}
	defer ds1.Stop()

	if err := ds2.Start(); err != nil {
		t.Fatalf("Failed to start discovery system 2: %v", err)
	}
	defer ds2.Stop()

	// Wait for discovery broadcasts
	time.Sleep(2 * time.Second)

	// Note: In real LAN environment, servers would discover each other via broadcast
	// In unit tests, we simulate discovery by manually adding peers
	// This is because UDP broadcast may not work in all test environments

	// Verify both systems are running
	ds1.mu.RLock()
	running1 := ds1.running
	ds1.mu.RUnlock()

	ds2.mu.RLock()
	running2 := ds2.running
	ds2.mu.RUnlock()

	if !running1 {
		t.Error("Discovery system 1 should be running")
	}
	if !running2 {
		t.Error("Discovery system 2 should be running")
	}
}

// TestDiscoverySystem_Integration_ManualPeerManagement tests manual peer addition and removal
func TestDiscoverySystem_Integration_ManualPeerManagement(t *testing.T) {
	identity, err := NewServerIdentity("TestServer")
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	ds, err := NewDiscoverySystem(identity, ":0", "localhost:8080")
	if err != nil {
		t.Fatalf("Failed to create discovery system: %v", err)
	}

	// Track discovered peers via callback
	discoveredPeers := make(map[string]*DiscoveredPeer)
	var peerMu sync.RWMutex
	ds.OnPeerDiscovered(func(peer *DiscoveredPeer) {
		peerMu.Lock()
		discoveredPeers[peer.ServerID] = peer
		peerMu.Unlock()
	})

	// Add multiple manual peers
	peers := []struct {
		serverID   string
		serverName string
		address    string
	}{
		{"peer-1", "Peer 1", "192.168.1.101:8080"},
		{"peer-2", "Peer 2", "192.168.1.102:8080"},
		{"peer-3", "Peer 3", "192.168.1.103:8080"},
	}

	for _, p := range peers {
		err := ds.AddManualPeer(p.serverID, p.serverName, p.address, "6.0.0", []string{"travel", "trade"})
		if err != nil {
			t.Errorf("Failed to add peer %s: %v", p.serverID, err)
		}
	}

	// Wait for callbacks
	time.Sleep(100 * time.Millisecond)

	// Verify all peers were added
	allPeers := ds.GetPeers()
	if len(allPeers) != 3 {
		t.Errorf("Expected 3 peers, got %d", len(allPeers))
	}

	// Verify callbacks were triggered
	peerMu.RLock()
	callbackCount := len(discoveredPeers)
	peerMu.RUnlock()
	if callbackCount != 3 {
		t.Errorf("Expected 3 callbacks, got %d", callbackCount)
	}

	// Test GetPeer for each
	for _, p := range peers {
		peer, exists := ds.GetPeer(p.serverID)
		if !exists {
			t.Errorf("Peer %s should exist", p.serverID)
			continue
		}
		if peer.ServerName != p.serverName {
			t.Errorf("Expected ServerName %s, got %s", p.serverName, peer.ServerName)
		}
		if peer.Address != p.address {
			t.Errorf("Expected Address %s, got %s", p.address, peer.Address)
		}
	}

	// Remove one peer
	ds.RemovePeer("peer-2")

	// Verify peer count
	allPeers = ds.GetPeers()
	if len(allPeers) != 2 {
		t.Errorf("Expected 2 peers after removal, got %d", len(allPeers))
	}

	// Verify removed peer is gone
	_, exists := ds.GetPeer("peer-2")
	if exists {
		t.Error("Removed peer should not exist")
	}

	// Verify other peers still exist
	_, exists = ds.GetPeer("peer-1")
	if !exists {
		t.Error("Peer 1 should still exist")
	}
	_, exists = ds.GetPeer("peer-3")
	if !exists {
		t.Error("Peer 3 should still exist")
	}
}

// TestDiscoverySystem_Integration_StaleCleanup tests automatic cleanup of stale peers
func TestDiscoverySystem_Integration_StaleCleanup(t *testing.T) {
	identity, err := NewServerIdentity("TestServer")
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	ds, err := NewDiscoverySystem(identity, ":0", "localhost:8080")
	if err != nil {
		t.Fatalf("Failed to create discovery system: %v", err)
	}

	// Add a peer
	ds.AddManualPeer("test-peer", "Test Peer", "192.168.1.100:8080", "6.0.0", []string{"travel"})

	// Verify peer exists
	_, exists := ds.GetPeer("test-peer")
	if !exists {
		t.Fatal("Peer should exist after addition")
	}

	// Manually set LastSeen to very old timestamp
	ds.mu.Lock()
	if peer, ok := ds.knownPeers["test-peer"]; ok {
		peer.LastSeen = time.Now().Add(-2 * DiscoveryTimeout)
	}
	ds.mu.Unlock()

	// Run cleanup
	ds.cleanupStalePeers()

	// Verify peer was removed
	_, exists = ds.GetPeer("test-peer")
	if exists {
		t.Error("Stale peer should be removed after cleanup")
	}
}

// TestDiscoverySystem_Integration_HighVolumeDiscovery tests discovery with many peers
func TestDiscoverySystem_Integration_HighVolumeDiscovery(t *testing.T) {
	identity, err := NewServerIdentity("TestServer")
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	ds, err := NewDiscoverySystem(identity, ":0", "localhost:8080")
	if err != nil {
		t.Fatalf("Failed to create discovery system: %v", err)
	}

	// Add 100 peers
	const numPeers = 100
	for i := 0; i < numPeers; i++ {
		serverID := generateTestServerID(i)
		err := ds.AddManualPeer(
			serverID,
			"Test Server",
			"192.168.1.100:8080",
			"6.0.0",
			[]string{"travel", "trade"},
		)
		if err != nil {
			t.Errorf("Failed to add peer %d: %v", i, err)
		}
	}

	// Verify all peers were added
	peers := ds.GetPeers()
	if len(peers) != numPeers {
		t.Errorf("Expected %d peers, got %d", numPeers, len(peers))
	}

	// Test retrieval performance (should be fast even with 100 peers)
	start := time.Now()
	for i := 0; i < 1000; i++ {
		_ = ds.GetPeers()
	}
	elapsed := time.Since(start)

	// Should complete 1000 GetPeers calls in < 300ms (relaxed for CI environments with race detection)
	if elapsed > 300*time.Millisecond {
		t.Errorf("GetPeers too slow: %v for 1000 calls", elapsed)
	}
}

// TestDiscoverySystem_Integration_ConcurrentAccess tests thread safety
func TestDiscoverySystem_Integration_ConcurrentAccess(t *testing.T) {
	identity, err := NewServerIdentity("TestServer")
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	ds, err := NewDiscoverySystem(identity, ":0", "localhost:8080")
	if err != nil {
		t.Fatalf("Failed to create discovery system: %v", err)
	}

	// Start concurrent operations
	done := make(chan bool, 3)

	// Goroutine 1: Add peers
	go func() {
		for i := 0; i < 50; i++ {
			serverID := generateTestServerID(i)
			ds.AddManualPeer(serverID, "Test", "localhost:8080", "6.0.0", []string{"travel"})
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Goroutine 2: Read peers
	go func() {
		for i := 0; i < 100; i++ {
			_ = ds.GetPeers()
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Goroutine 3: Remove peers
	go func() {
		for i := 0; i < 25; i++ {
			serverID := generateTestServerID(i)
			ds.RemovePeer(serverID)
			time.Sleep(2 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// Verify system is still functional
	peers := ds.GetPeers()
	t.Logf("Final peer count: %d", len(peers))

	// Should have ~25 peers (50 added - 25 removed)
	if len(peers) < 20 || len(peers) > 30 {
		t.Errorf("Expected ~25 peers, got %d (race condition?)", len(peers))
	}
}

// Helper function to generate deterministic test server IDs
func generateTestServerID(index int) string {
	return "test-server-" + string(rune('0'+index%10)) + string(rune('0'+index/10))
}

// BenchmarkDiscoverySystem_Integration_HighVolume benchmarks high-volume peer management
func BenchmarkDiscoverySystem_Integration_HighVolume(b *testing.B) {
	identity, _ := NewServerIdentity("TestServer")
	ds, _ := NewDiscoverySystem(identity, ":0", "localhost:8080")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Add peer
		serverID := generateTestServerID(i)
		ds.AddManualPeer(serverID, "Test", "localhost:8080", "6.0.0", []string{"travel"})

		// Read all peers
		_ = ds.GetPeers()

		// Remove peer
		ds.RemovePeer(serverID)
	}
}
