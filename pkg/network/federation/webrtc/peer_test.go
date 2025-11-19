package webrtc

import (
	"testing"
	"time"
)

func TestConnectionState_String(t *testing.T) {
	tests := []struct {
		state    ConnectionState
		expected string
	}{
		{StateNew, "New"},
		{StateConnecting, "Connecting"},
		{StateConnected, "Connected"},
		{StateDisconnected, "Disconnected"},
		{StateFailed, "Failed"},
		{StateClosed, "Closed"},
		{ConnectionState(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.state.String()
			if got != tt.expected {
				t.Errorf("got %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestICECandidateType_String(t *testing.T) {
	tests := []struct {
		candType ICECandidateType
		expected string
	}{
		{CandidateHost, "Host"},
		{CandidateServerReflexive, "ServerReflexive"},
		{CandidateRelay, "Relay"},
		{ICECandidateType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.candType.String()
			if got != tt.expected {
				t.Errorf("got %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	if len(cfg.STUNServers) == 0 {
		t.Error("expected at least one STUN server")
	}

	if cfg.DataChannelLabel != "venture-federation" {
		t.Errorf("got label %s, want venture-federation", cfg.DataChannelLabel)
	}

	if cfg.ICETimeout != 10*time.Second {
		t.Errorf("got ICE timeout %v, want 10s", cfg.ICETimeout)
	}

	if cfg.MaxMessageSize != 1024*1024 {
		t.Errorf("got max message size %d, want 1MB", cfg.MaxMessageSize)
	}

	if !cfg.EnableFallback {
		t.Error("expected fallback to be enabled by default")
	}
}

func TestNewPeer(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid peer with default config",
			id:      "peer-1",
			config:  nil,
			wantErr: false,
		},
		{
			name:    "valid peer with custom config",
			id:      "peer-2",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name:    "empty peer ID",
			id:      "",
			config:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			peer, err := NewPeer(tt.id, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if peer == nil {
					t.Fatal("expected non-nil peer")
				}
				if peer.ID != tt.id {
					t.Errorf("got ID %s, want %s", peer.ID, tt.id)
				}
				if peer.GetState() != StateNew {
					t.Errorf("got state %s, want New", peer.GetState())
				}
				peer.Close()
			}
		})
	}
}

func TestPeer_Connect(t *testing.T) {
	config := DefaultConfig()
	config.ConnectionTimeout = 3 * time.Second

	peer, err := NewPeer("peer-alice", config)
	if err != nil {
		t.Fatalf("NewPeer failed: %v", err)
	}
	defer peer.Close()

	err = peer.Connect("peer-bob")
	if err != nil {
		t.Errorf("Connect failed: %v", err)
	}

	state := peer.GetState()
	if state != StateConnected {
		t.Errorf("got state %s, want Connected", state)
	}

	stats := peer.GetStats()
	if stats.State != StateConnected {
		t.Errorf("stats state %s, want Connected", stats.State)
	}

	if !stats.ConnectedAt.IsZero() && time.Since(stats.ConnectedAt) > 5*time.Second {
		t.Error("ConnectedAt timestamp seems incorrect")
	}

	remotePeerID := peer.GetRemotePeerID()
	if remotePeerID != "peer-bob" {
		t.Errorf("got remote peer %s, want peer-bob", remotePeerID)
	}
}

func TestPeer_Connect_AlreadyConnected(t *testing.T) {
	peer, err := NewPeer("peer-alice", DefaultConfig())
	if err != nil {
		t.Fatalf("NewPeer failed: %v", err)
	}
	defer peer.Close()

	err = peer.Connect("peer-bob")
	if err != nil {
		t.Fatalf("first Connect failed: %v", err)
	}

	// Try to connect again
	err = peer.Connect("peer-charlie")
	if err == nil {
		t.Error("expected error when connecting already connected peer")
	}
}

func TestPeer_SendReceive(t *testing.T) {
	peer, err := NewPeer("peer-alice", DefaultConfig())
	if err != nil {
		t.Fatalf("NewPeer failed: %v", err)
	}
	defer peer.Close()

	err = peer.Connect("peer-bob")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	testData := []byte("test message")
	err = peer.Send(testData)
	if err != nil {
		t.Errorf("Send failed: %v", err)
	}

	stats := peer.GetStats()
	if stats.MessagesSent != 1 {
		t.Errorf("got %d messages sent, want 1", stats.MessagesSent)
	}
	if stats.BytesSent != uint64(len(testData)) {
		t.Errorf("got %d bytes sent, want %d", stats.BytesSent, len(testData))
	}
}

func TestPeer_Send_NotConnected(t *testing.T) {
	peer, err := NewPeer("peer-alice", DefaultConfig())
	if err != nil {
		t.Fatalf("NewPeer failed: %v", err)
	}
	defer peer.Close()

	err = peer.Send([]byte("test"))
	if err != ErrNotConnected {
		t.Errorf("got error %v, want ErrNotConnected", err)
	}
}

func TestPeer_Send_MessageTooLarge(t *testing.T) {
	config := DefaultConfig()
	config.MaxMessageSize = 100

	peer, err := NewPeer("peer-alice", config)
	if err != nil {
		t.Fatalf("NewPeer failed: %v", err)
	}
	defer peer.Close()

	err = peer.Connect("peer-bob")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	largeData := make([]byte, 200)
	err = peer.Send(largeData)
	if err == nil || err != ErrMessageTooLarge {
		t.Errorf("got error %v, want ErrMessageTooLarge", err)
	}
}

func TestPeer_Close(t *testing.T) {
	peer, err := NewPeer("peer-alice", DefaultConfig())
	if err != nil {
		t.Fatalf("NewPeer failed: %v", err)
	}

	err = peer.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	state := peer.GetState()
	if state != StateClosed {
		t.Errorf("got state %s, want Closed", state)
	}

	// Close again should be idempotent
	err = peer.Close()
	if err != nil {
		t.Errorf("second Close failed: %v", err)
	}
}

func TestManager_CreatePeer(t *testing.T) {
	manager := NewManager(DefaultConfig())

	peer, err := manager.CreatePeer("peer-1")
	if err != nil {
		t.Fatalf("CreatePeer failed: %v", err)
	}
	if peer == nil {
		t.Fatal("expected non-nil peer")
	}

	// Try to create duplicate
	_, err = manager.CreatePeer("peer-1")
	if err == nil {
		t.Error("expected error when creating duplicate peer")
	}
}

func TestManager_GetPeer(t *testing.T) {
	manager := NewManager(DefaultConfig())

	_, err := manager.CreatePeer("peer-1")
	if err != nil {
		t.Fatalf("CreatePeer failed: %v", err)
	}

	peer, ok := manager.GetPeer("peer-1")
	if !ok {
		t.Error("expected to find peer-1")
	}
	if peer.ID != "peer-1" {
		t.Errorf("got peer ID %s, want peer-1", peer.ID)
	}

	_, ok = manager.GetPeer("nonexistent")
	if ok {
		t.Error("expected not to find nonexistent peer")
	}
}

func TestManager_RemovePeer(t *testing.T) {
	manager := NewManager(DefaultConfig())

	_, err := manager.CreatePeer("peer-1")
	if err != nil {
		t.Fatalf("CreatePeer failed: %v", err)
	}

	err = manager.RemovePeer("peer-1")
	if err != nil {
		t.Errorf("RemovePeer failed: %v", err)
	}

	_, ok := manager.GetPeer("peer-1")
	if ok {
		t.Error("peer should be removed")
	}

	// Remove again should fail
	err = manager.RemovePeer("peer-1")
	if err == nil {
		t.Error("expected error when removing nonexistent peer")
	}
}

func TestManager_GetMetrics(t *testing.T) {
	manager := NewManager(DefaultConfig())

	// Create several peers
	peer1, _ := manager.CreatePeer("peer-1")
	peer2, _ := manager.CreatePeer("peer-2")
	_, _ = manager.CreatePeer("peer-3")

	// Connect some peers
	peer1.Connect("peer-bob")
	peer2.Connect("peer-alice")
	// peer-3 not connected

	metrics := manager.GetMetrics()

	if metrics.TotalConnections != 3 {
		t.Errorf("got %d total connections, want 3", metrics.TotalConnections)
	}
	if metrics.ActiveConnections != 2 {
		t.Errorf("got %d active connections, want 2", metrics.ActiveConnections)
	}
}

func TestManager_CloseAll(t *testing.T) {
	manager := NewManager(DefaultConfig())

	manager.CreatePeer("peer-1")
	manager.CreatePeer("peer-2")
	manager.CreatePeer("peer-3")

	err := manager.CloseAll()
	if err != nil {
		t.Errorf("CloseAll failed: %v", err)
	}

	metrics := manager.GetMetrics()
	if metrics.TotalConnections != 0 {
		t.Errorf("expected 0 connections after CloseAll, got %d", metrics.TotalConnections)
	}
}

func TestPeerStats_Tracking(t *testing.T) {
	peer, err := NewPeer("peer-alice", DefaultConfig())
	if err != nil {
		t.Fatalf("NewPeer failed: %v", err)
	}
	defer peer.Close()

	err = peer.Connect("peer-bob")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Send multiple messages
	for i := 0; i < 5; i++ {
		data := []byte("test message")
		peer.Send(data)
	}

	stats := peer.GetStats()
	if stats.MessagesSent != 5 {
		t.Errorf("got %d messages sent, want 5", stats.MessagesSent)
	}

	expectedBytes := uint64(5 * len([]byte("test message")))
	if stats.BytesSent != expectedBytes {
		t.Errorf("got %d bytes sent, want %d", stats.BytesSent, expectedBytes)
	}

	if stats.LastActivity.IsZero() {
		t.Error("expected LastActivity to be set")
	}
}

// Benchmark tests
func BenchmarkPeer_Connect(b *testing.B) {
	config := DefaultConfig()
	for i := 0; i < b.N; i++ {
		peer, _ := NewPeer("peer-test", config)
		peer.Connect("peer-remote")
		peer.Close()
	}
}

func BenchmarkPeer_Send(b *testing.B) {
	peer, _ := NewPeer("peer-alice", DefaultConfig())
	peer.Connect("peer-bob")
	defer peer.Close()

	data := []byte("test message")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		peer.Send(data)
	}
}

func BenchmarkManager_CreatePeer(b *testing.B) {
	manager := NewManager(DefaultConfig())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		peerID := "peer-" + string(rune(i))
		manager.CreatePeer(peerID)
	}
}
