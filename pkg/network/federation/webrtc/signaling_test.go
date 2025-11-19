package webrtc

import (
	"testing"
	"time"
)

func TestNewSignalingClient(t *testing.T) {
	client := NewSignalingClient("ws://localhost:8080", "peer-1")
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.peerID != "peer-1" {
		t.Errorf("got peer ID %s, want peer-1", client.peerID)
	}
	if client.url != "ws://localhost:8080" {
		t.Errorf("got URL %s, want ws://localhost:8080", client.url)
	}
}

func TestSignalingClient_Connect(t *testing.T) {
	client := NewSignalingClient("ws://localhost:8080", "peer-1")
	defer client.Close()

	err := client.Connect()
	if err != nil {
		t.Errorf("Connect failed: %v", err)
	}

	client.mu.RLock()
	connected := client.connected
	client.mu.RUnlock()

	if !connected {
		t.Error("expected client to be connected")
	}

	// Connect again should be idempotent
	err = client.Connect()
	if err != nil {
		t.Errorf("second Connect failed: %v", err)
	}
}

func TestSignalingClient_SendOffer(t *testing.T) {
	client := NewSignalingClient("ws://localhost:8080", "peer-alice")
	defer client.Close()

	err := client.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Register remote peer
	client.RegisterPeer("peer-bob")

	offer := &SDPOffer{
		Type: "offer",
		SDP:  "v=0\r\no=- 123456 2 IN IP4 127.0.0.1\r\n",
	}

	err = client.SendOffer("peer-bob", offer)
	if err != nil {
		t.Errorf("SendOffer failed: %v", err)
	}
}

func TestSignalingClient_SendOffer_NotConnected(t *testing.T) {
	client := NewSignalingClient("ws://localhost:8080", "peer-alice")
	defer client.Close()

	offer := &SDPOffer{Type: "offer", SDP: "test"}
	err := client.SendOffer("peer-bob", offer)
	if err != ErrSignalingNotConnected {
		t.Errorf("got error %v, want ErrSignalingNotConnected", err)
	}
}

func TestSignalingClient_SendAnswer(t *testing.T) {
	client := NewSignalingClient("ws://localhost:8080", "peer-alice")
	defer client.Close()

	err := client.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	client.RegisterPeer("peer-bob")

	answer := &SDPAnswer{
		Type: "answer",
		SDP:  "v=0\r\no=- 654321 2 IN IP4 127.0.0.1\r\n",
	}

	err = client.SendAnswer("peer-bob", answer)
	if err != nil {
		t.Errorf("SendAnswer failed: %v", err)
	}
}

func TestSignalingClient_SendICECandidate(t *testing.T) {
	client := NewSignalingClient("ws://localhost:8080", "peer-alice")
	defer client.Close()

	err := client.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	client.RegisterPeer("peer-bob")

	candidate := &ICECandidate{
		Candidate:     "candidate:1 1 UDP 2130706431 192.168.1.100 54321 typ host",
		SDPMid:        "0",
		SDPMLineIndex: 0,
	}

	err = client.SendICECandidate("peer-bob", candidate)
	if err != nil {
		t.Errorf("SendICECandidate failed: %v", err)
	}
}

func TestSignalingClient_SendBye(t *testing.T) {
	client := NewSignalingClient("ws://localhost:8080", "peer-alice")
	defer client.Close()

	err := client.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	err = client.SendBye("peer-bob")
	if err != nil {
		t.Errorf("SendBye failed: %v", err)
	}
}

func TestSignalingClient_RegisterUnregisterPeer(t *testing.T) {
	client := NewSignalingClient("ws://localhost:8080", "peer-alice")

	client.RegisterPeer("peer-1")
	client.RegisterPeer("peer-2")

	peers := client.GetActivePeers()
	if len(peers) != 2 {
		t.Errorf("got %d peers, want 2", len(peers))
	}

	client.UnregisterPeer("peer-1")
	peers = client.GetActivePeers()
	if len(peers) != 1 {
		t.Errorf("got %d peers after unregister, want 1", len(peers))
	}
}

func TestSignalingClient_CleanupPeers(t *testing.T) {
	client := NewSignalingClient("ws://localhost:8080", "peer-alice")

	// Register peer with old timestamp
	client.mu.Lock()
	client.peers["old-peer"] = time.Now().Add(-10 * time.Minute)
	client.peers["recent-peer"] = time.Now()
	client.mu.Unlock()

	client.cleanupPeers()

	client.mu.RLock()
	_, hasOld := client.peers["old-peer"]
	_, hasRecent := client.peers["recent-peer"]
	client.mu.RUnlock()

	if hasOld {
		t.Error("old peer should be cleaned up")
	}
	if !hasRecent {
		t.Error("recent peer should remain")
	}
}

func TestSignalingClient_Close(t *testing.T) {
	client := NewSignalingClient("ws://localhost:8080", "peer-alice")

	err := client.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	err = client.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	client.mu.RLock()
	connected := client.connected
	client.mu.RUnlock()

	if connected {
		t.Error("client should not be connected after Close")
	}

	// Close again should be idempotent
	err = client.Close()
	if err != nil {
		t.Errorf("second Close failed: %v", err)
	}
}

func TestNewSignalingServer(t *testing.T) {
	server := NewSignalingServer("ws://localhost:8080")
	if server == nil {
		t.Fatal("expected non-nil server")
	}
	if server.address != "ws://localhost:8080" {
		t.Errorf("got address %s, want ws://localhost:8080", server.address)
	}
}

func TestSignalingServer_Start(t *testing.T) {
	server := NewSignalingServer("ws://localhost:8080")
	defer server.Stop()

	err := server.Start()
	if err != nil {
		t.Errorf("Start failed: %v", err)
	}

	// Give server time to start
	time.Sleep(100 * time.Millisecond)
}

func TestSignalingServer_RegisterClient(t *testing.T) {
	server := NewSignalingServer("ws://localhost:8080")
	defer server.Stop()

	client, err := server.RegisterClient("client-1")
	if err != nil {
		t.Fatalf("RegisterClient failed: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}

	// Try to register duplicate
	_, err = server.RegisterClient("client-1")
	if err == nil {
		t.Error("expected error when registering duplicate client")
	}
}

func TestSignalingServer_RelayMessage(t *testing.T) {
	server := NewSignalingServer("ws://localhost:8080")
	defer server.Stop()

	clientA, _ := server.RegisterClient("client-a")
	clientB, _ := server.RegisterClient("client-b")

	clientA.Connect()
	clientB.Connect()

	// Register peers
	clientA.RegisterPeer("client-b")
	clientB.RegisterPeer("client-a")

	msg := &SignalingMessage{
		Type:  "offer",
		From:  "client-a",
		To:    "client-b",
		Offer: &SDPOffer{Type: "offer", SDP: "test"},
	}

	err := server.RelayMessage(msg)
	if err != nil {
		t.Errorf("RelayMessage failed: %v", err)
	}

	// Check if message was received
	select {
	case receivedMsg := <-clientB.Receive():
		if receivedMsg.From != "client-a" {
			t.Errorf("got message from %s, want client-a", receivedMsg.From)
		}
	case <-time.After(1 * time.Second):
		t.Error("timeout waiting for message")
	}
}

func TestSignalingServer_RelayMessage_PeerNotFound(t *testing.T) {
	server := NewSignalingServer("ws://localhost:8080")
	defer server.Stop()

	msg := &SignalingMessage{
		Type: "offer",
		From: "client-a",
		To:   "nonexistent",
	}

	err := server.RelayMessage(msg)
	if err != ErrPeerNotFound {
		t.Errorf("got error %v, want ErrPeerNotFound", err)
	}
}

func TestSignalingServer_GetStats(t *testing.T) {
	server := NewSignalingServer("ws://localhost:8080")
	defer server.Stop()

	server.RegisterClient("client-1")
	server.RegisterClient("client-2")

	stats := server.GetStats()
	if stats["clients"].(int) != 2 {
		t.Errorf("got %d clients, want 2", stats["clients"])
	}
	if stats["address"].(string) != "ws://localhost:8080" {
		t.Errorf("got address %s, want ws://localhost:8080", stats["address"])
	}
}

func TestSignalingServer_Stop(t *testing.T) {
	server := NewSignalingServer("ws://localhost:8080")

	server.RegisterClient("client-1")
	server.RegisterClient("client-2")

	err := server.Stop()
	if err != nil {
		t.Errorf("Stop failed: %v", err)
	}

	server.mu.RLock()
	clientCount := len(server.clients)
	server.mu.RUnlock()

	if clientCount != 0 {
		t.Errorf("expected 0 clients after Stop, got %d", clientCount)
	}
}

func TestSignalingMessage_JSON(t *testing.T) {
	msg := &SignalingMessage{
		Type:      "offer",
		From:      "peer-a",
		To:        "peer-b",
		Offer:     &SDPOffer{Type: "offer", SDP: "test sdp"},
		Timestamp: time.Now(),
	}

	// Marshal
	data, err := msg.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Unmarshal
	var decoded SignalingMessage
	err = decoded.UnmarshalJSON(data)
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if decoded.Type != msg.Type {
		t.Errorf("got type %s, want %s", decoded.Type, msg.Type)
	}
	if decoded.From != msg.From {
		t.Errorf("got from %s, want %s", decoded.From, msg.From)
	}
	if decoded.To != msg.To {
		t.Errorf("got to %s, want %s", decoded.To, msg.To)
	}
	if decoded.Offer == nil {
		t.Error("expected non-nil offer")
	}
	if decoded.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestSignalingMessage_JSON_InvalidTimestamp(t *testing.T) {
	invalidJSON := []byte(`{"type":"offer","from":"a","to":"b","timestamp":"invalid"}`)

	var msg SignalingMessage
	err := msg.UnmarshalJSON(invalidJSON)
	if err == nil {
		t.Error("expected error for invalid timestamp")
	}
}

// Benchmark tests
func BenchmarkSignalingClient_SendOffer(b *testing.B) {
	client := NewSignalingClient("ws://localhost:8080", "peer-alice")
	client.Connect()
	defer client.Close()

	client.RegisterPeer("peer-bob")
	offer := &SDPOffer{Type: "offer", SDP: "test sdp"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.SendOffer("peer-bob", offer)
	}
}

func BenchmarkSignalingServer_RelayMessage(b *testing.B) {
	server := NewSignalingServer("ws://localhost:8080")
	defer server.Stop()

	clientA, _ := server.RegisterClient("client-a")
	clientB, _ := server.RegisterClient("client-b")
	clientA.Connect()
	clientB.Connect()
	clientA.RegisterPeer("client-b")
	clientB.RegisterPeer("client-a")

	msg := &SignalingMessage{
		Type:  "offer",
		From:  "client-a",
		To:    "client-b",
		Offer: &SDPOffer{Type: "offer", SDP: "test"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.RelayMessage(msg)
		// Drain receive channel
		select {
		case <-clientB.Receive():
		default:
		}
	}
}
