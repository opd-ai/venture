package network

import (
	"math/big"
	"math/rand"
	"testing"
	"time"
)

// startMessageConsumer starts a goroutine to consume messages from the chat manager's queue.
// This simulates the network layer processing and sending messages.
// Returns a done channel that should be closed when the test completes.
func startMessageConsumer(cm *ChatManager) chan struct{} {
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-cm.messageQueue:
				// Consume message (simulating network layer sending it)
			case <-done:
				return
			}
		}
	}()
	return done
}

// TestChatIntegrationE2E tests end-to-end chat message flow with encryption
func TestChatIntegrationE2E(t *testing.T) {
	cm := NewChatManager()
	done := startMessageConsumer(cm)
	defer close(done)

	// Setup two players with encryption
	params := DefaultDHParams()
	alice, _ := GenerateKeyPair(params)
	bob, _ := GenerateKeyPair(params)

	aliceSecret, _ := ComputeSharedSecret(alice.PrivateKey, bob.PublicKey, params)
	bobSecret, _ := ComputeSharedSecret(bob.PrivateKey, alice.PublicKey, params)

	aliceKey := DeriveAESKey(aliceSecret)
	bobKey := DeriveAESKey(bobSecret)

	// Register players
	cm.AddPlayer(1, Vector2{X: 0, Y: 0}, aliceKey)
	cm.AddPlayer(2, Vector2{X: 5, Y: 0}, bobKey)

	// Send message from Alice
	packet, err := cm.SendMessage(1, 0, "Hello Bob!", 2, -1)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Verify encryption
	if string(packet.EncryptedPayload) == "Hello Bob!" {
		t.Error("Message not encrypted")
	}

	// Decrypt message (simulating Bob's client)
	decrypted, err := DecryptMessage(bobKey, packet.EncryptedPayload)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	if string(decrypted) != "Hello Bob!" {
		t.Errorf("Decrypted message incorrect: %q", string(decrypted))
	}

	// Verify message in pending ACKs
	if cm.GetPendingCount() == 0 {
		t.Error("Message not added to pending ACKs")
	}

	// Simulate ACK from server
	ack := &MessageACK{
		MessageID: packet.MessageID,
		SenderID:  1,
		Success:   true,
	}
	cm.ProcessACK(ack)

	// Verify message removed from pending
	if cm.GetPendingCount() != 0 {
		t.Error("Message not removed from pending ACKs after ACK")
	}
}

// TestChatIntegrationLatencySimulation tests chat with simulated network latency
func TestChatIntegrationLatencySimulation(t *testing.T) {
	latencies := []time.Duration{
		200 * time.Millisecond,
		500 * time.Millisecond,
		2000 * time.Millisecond,
		5000 * time.Millisecond,
	}

	for _, latency := range latencies {
		t.Run(latency.String(), func(t *testing.T) {
			cm := NewChatManager()
			done := startMessageConsumer(cm)
			defer close(done)
			cm.ackTimeout = latency + 1*time.Second // Timeout longer than latency

			encKey := DeriveAESKey(big.NewInt(123))
			cm.AddPlayer(1, Vector2{}, encKey)

			// Send message
			packet, err := cm.SendMessage(1, 0, "Test message", 0, -1)
			if err != nil {
				// Clear rate limit and retry
				cm.mu.Lock()
				delete(cm.players[1].RateLimitState, 0)
				cm.mu.Unlock()
				packet, err = cm.SendMessage(1, 0, "Test message", 0, -1)
			}
			if err != nil {
				t.Fatalf("Failed to send message: %v", err)
			}

			// Simulate network latency
			time.Sleep(latency)

			// Message should still be pending (no ACK yet)
			if cm.GetPendingCount() == 0 {
				t.Error("Message prematurely removed from pending")
			}

			// Send delayed ACK
			ack := &MessageACK{
				MessageID: packet.MessageID,
				SenderID:  1,
				Success:   true,
			}
			cm.ProcessACK(ack)

			// Message should be removed
			if cm.GetPendingCount() != 0 {
				t.Error("Message not removed after delayed ACK")
			}
		})
	}
}

// TestChatIntegrationPacketLoss tests ACK/NACK with simulated packet loss
func TestChatIntegrationPacketLoss(t *testing.T) {
	tests := []struct {
		name          string
		lossRate      float64 // 0.0 to 1.0
		messageCount  int
		expectRetries bool
	}{
		{"5% loss", 0.05, 100, true},
		{"10% loss", 0.10, 100, true},
		{"20% loss", 0.20, 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewChatManager()
			cm.ackTimeout = 200 * time.Millisecond

			// Start goroutine to consume messages from the queue
			done := make(chan struct{})
			go func() {
				for {
					select {
					case <-cm.messageQueue:
						// Consume message (simulating network layer sending it)
					case <-done:
						return
					}
				}
			}()
			defer close(done)

			encKey := DeriveAESKey(big.NewInt(456))
			cm.AddPlayer(1, Vector2{}, encKey)

			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			successfulDeliveries := 0
			retriedMessages := 0

			for i := 0; i < tt.messageCount; i++ {
				// Clear rate limit
				cm.mu.Lock()
				delete(cm.players[1].RateLimitState, 0)
				cm.mu.Unlock()

				packet, err := cm.SendMessage(1, 0, "Test", 0, -1)
				if err != nil {
					continue
				}

				// Simulate packet loss
				if rng.Float64() < tt.lossRate {
					// Packet lost - wait for timeout and retry
					time.Sleep(cm.ackTimeout + 50*time.Millisecond)
					cm.ProcessTimeouts()

					cm.mu.RLock()
					pending, exists := cm.pendingACKs[packet.MessageID]
					cm.mu.RUnlock()

					if exists && pending.Attempts > 0 {
						retriedMessages++
					}

					// Eventually send ACK
					cm.ProcessACK(&MessageACK{
						MessageID: packet.MessageID,
						SenderID:  1,
						Success:   true,
					})
				} else {
					// Packet delivered successfully
					cm.ProcessACK(&MessageACK{
						MessageID: packet.MessageID,
						SenderID:  1,
						Success:   true,
					})
					successfulDeliveries++
				}
			}

			// Verify retries occurred due to packet loss
			if tt.expectRetries && retriedMessages == 0 {
				t.Error("Expected retries due to packet loss, got none")
			}

			t.Logf("Loss rate %.0f%%: %d/%d successful deliveries, %d retries",
				tt.lossRate*100, successfulDeliveries, tt.messageCount, retriedMessages)
		})
	}
}

// TestChatIntegrationReorderinghandling tests message reordering tolerance
func TestChatIntegrationMessageReordering(t *testing.T) {
	cm := NewChatManager()
	done := startMessageConsumer(cm)
	defer close(done)
	encKey := DeriveAESKey(big.NewInt(789))
	cm.AddPlayer(1, Vector2{}, encKey)

	// Send multiple messages
	var packets []*ChatMessagePacket
	for i := 0; i < 5; i++ {
		// Clear rate limit
		cm.mu.Lock()
		delete(cm.players[1].RateLimitState, 0)
		cm.mu.Unlock()

		packet, err := cm.SendMessage(1, 0, "Message", 0, -1)
		if err != nil {
			t.Fatalf("Failed to send message %d: %v", i, err)
		}
		packet.SequenceNum = uint32(i)
		packets = append(packets, packet)
	}

	// Process ACKs out of order (5, 3, 1, 4, 2)
	order := []int{4, 2, 0, 3, 1}
	for _, idx := range order {
		ack := &MessageACK{
			MessageID: packets[idx].MessageID,
			SenderID:  1,
			Success:   true,
		}
		cm.ProcessACK(ack)
	}

	// All messages should be acknowledged regardless of order
	if cm.GetPendingCount() != 0 {
		t.Errorf("Expected 0 pending after all ACKs, got %d", cm.GetPendingCount())
	}
}

// TestChatIntegrationDuplicateDetection tests duplicate message handling
func TestChatIntegrationDuplicateDetection(t *testing.T) {
	cm := NewChatManager()
	done := startMessageConsumer(cm)
	defer close(done)
	encKey := DeriveAESKey(big.NewInt(101112))
	cm.AddPlayer(1, Vector2{}, encKey)

	// Send message
	packet, _ := cm.SendMessage(1, 0, "Duplicate test", 0, -1)

	// Send ACK twice (simulating duplicate ACK)
	ack := &MessageACK{
		MessageID: packet.MessageID,
		SenderID:  1,
		Success:   true,
	}

	cm.ProcessACK(ack)
	initialPending := cm.GetPendingCount()

	// Send duplicate ACK
	cm.ProcessACK(ack)

	// Pending count should not change (duplicate ignored)
	if cm.GetPendingCount() != initialPending {
		t.Error("Duplicate ACK affected pending count")
	}
}

// TestChatIntegrationRangeValidation tests local chat range enforcement
func TestChatIntegrationRangeValidation(t *testing.T) {
	cm := NewChatManager()
	done := startMessageConsumer(cm)
	defer close(done)
	encKey := DeriveAESKey(big.NewInt(131415))

	// Setup players at different distances
	cm.AddPlayer(1, Vector2{X: 0, Y: 0}, encKey)
	cm.AddPlayer(2, Vector2{X: 5, Y: 0}, encKey)  // 5 units away
	cm.AddPlayer(3, Vector2{X: 15, Y: 0}, encKey) // 15 units away

	tests := []struct {
		name        string
		sender      uint64
		recipient   uint64
		localRadius float64
		expectError bool
	}{
		{"within range", 1, 2, 10.0, false},
		{"out of range", 1, 3, 10.0, true},
		{"exact range", 1, 2, 5.0, false},
		{"unlimited range", 1, 3, -1.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear rate limit
			cm.mu.Lock()
			delete(cm.players[tt.sender].RateLimitState, 1) // Local channel
			cm.mu.Unlock()

			_, err := cm.SendMessage(tt.sender, 1, "Local test", tt.recipient, tt.localRadius)

			if tt.expectError && err == nil {
				t.Error("Expected range error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestChatIntegrationMultiplePlayers tests chat with multiple concurrent players
func TestChatIntegrationMultiplePlayers(t *testing.T) {
	cm := NewChatManager()
	done := startMessageConsumer(cm)
	defer close(done)
	playerCount := 50

	// Register 50 players
	for i := 1; i <= playerCount; i++ {
		encKey := DeriveAESKey(big.NewInt(int64(i)))
		cm.AddPlayer(uint64(i), Vector2{X: float64(i * 10), Y: 0}, encKey)
	}

	// Each player sends a message
	messagesSent := 0
	for i := 1; i <= playerCount; i++ {
		// Clear rate limit
		cm.mu.Lock()
		delete(cm.players[uint64(i)].RateLimitState, 0)
		cm.mu.Unlock()

		_, err := cm.SendMessage(uint64(i), 0, "Hello from all!", 0, -1)
		if err == nil {
			messagesSent++
		}
	}

	if messagesSent != playerCount {
		t.Errorf("Expected %d messages sent, got %d", playerCount, messagesSent)
	}

	// Verify pending ACKs
	pending := cm.GetPendingCount()
	if pending != messagesSent {
		t.Errorf("Expected %d pending ACKs, got %d", messagesSent, pending)
	}
}

// TestChatIntegrationThroughput tests message throughput
func TestChatIntegrationThroughput(t *testing.T) {
	cm := NewChatManager()

	// Start goroutine to consume messages from the queue
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-cm.messageQueue:
				// Consume message (simulating network layer sending it)
			case <-done:
				return
			}
		}
	}()
	defer close(done)

	encKey := DeriveAESKey(big.NewInt(161718))
	cm.AddPlayer(1, Vector2{}, encKey)

	messageCount := 500
	messagesPerSecond := 10
	interval := time.Second / time.Duration(messagesPerSecond)

	start := time.Now()
	sent := 0

	for i := 0; i < messageCount; i++ {
		// Rate-limit messages
		time.Sleep(interval)

		// Clear rate limit (simulating different channels/timing)
		cm.mu.Lock()
		delete(cm.players[1].RateLimitState, 0)
		cm.mu.Unlock()

		_, err := cm.SendMessage(1, 0, "Throughput test", 0, -1)
		if err == nil {
			sent++
		}
	}

	elapsed := time.Since(start)
	actualThroughput := float64(sent) / elapsed.Seconds()

	t.Logf("Sent %d messages in %v (%.1f msg/s)", sent, elapsed, actualThroughput)

	// Verify reasonable throughput (accounting for rate limiting)
	if actualThroughput < 5.0 {
		t.Errorf("Throughput too low: %.1f msg/s", actualThroughput)
	}
}

// TestChatIntegrationProfanityFiltering tests profanity filter integration
func TestChatIntegrationProfanityFiltering(t *testing.T) {
	cm := NewChatManager()
	done := startMessageConsumer(cm)
	defer close(done)
	pf := NewProfanityFilter()
	pf.Enable()

	encKey := DeriveAESKey(big.NewInt(192021))
	cm.AddPlayer(1, Vector2{}, encKey)

	// Send message with profanity
	packet, _ := cm.SendMessage(1, 0, "This is damn annoying", 0, -1)

	// Decrypt message
	decrypted, _ := DecryptMessage(encKey, packet.EncryptedPayload)

	// Apply profanity filter
	filtered := pf.Filter(string(decrypted))

	// Verify filtering occurred
	if string(decrypted) == filtered {
		t.Error("Profanity filter did not modify message")
	}

	if containsString(filtered, "damn") {
		t.Error("Profanity still present after filtering")
	}
}

// TestChatIntegrationMaxRetryFailure tests message failure after max retries
func TestChatIntegrationMaxRetryFailure(t *testing.T) {
	cm := NewChatManager()
	done := startMessageConsumer(cm)
	defer close(done)
	cm.maxRetries = 2

	encKey := DeriveAESKey(big.NewInt(222324))
	cm.AddPlayer(1, Vector2{}, encKey)

	packet, _ := cm.SendMessage(1, 0, "Retry test", 0, -1)

	// Send multiple NACKs
	nack := &MessageACK{
		MessageID: packet.MessageID,
		SenderID:  1,
		Success:   false,
		Reason:    "Test failure",
	}

	for i := 0; i < cm.maxRetries+1; i++ {
		cm.ProcessACK(nack)
	}

	// Message should be removed after exceeding max retries
	if cm.GetPendingCount() != 0 {
		t.Error("Message still pending after max retries exceeded")
	}
}

// BenchmarkChatIntegrationE2E benchmarks full E2E message flow
func BenchmarkChatIntegrationE2E(b *testing.B) {
	cm := NewChatManager()

	params := DefaultDHParams()
	keyPair, _ := GenerateKeyPair(params)
	secret, _ := ComputeSharedSecret(keyPair.PrivateKey, keyPair.PublicKey, params)
	encKey := DeriveAESKey(secret)

	cm.AddPlayer(1, Vector2{}, encKey)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clear rate limit
		cm.mu.Lock()
		delete(cm.players[1].RateLimitState, 0)
		cm.mu.Unlock()

		packet, _ := cm.SendMessage(1, 0, "Benchmark message", 0, -1)

		ack := &MessageACK{
			MessageID: packet.MessageID,
			SenderID:  1,
			Success:   true,
		}
		cm.ProcessACK(ack)
	}
}

// BenchmarkChatIntegrationMultiPlayer benchmarks multi-player scenarios
func BenchmarkChatIntegrationMultiPlayer(b *testing.B) {
	cm := NewChatManager()
	playerCount := 50

	// Setup players
	for i := 1; i <= playerCount; i++ {
		encKey := DeriveAESKey(big.NewInt(int64(i)))
		cm.AddPlayer(uint64(i), Vector2{X: float64(i), Y: 0}, encKey)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		playerID := uint64((i % playerCount) + 1)

		// Clear rate limit
		cm.mu.Lock()
		delete(cm.players[playerID].RateLimitState, 0)
		cm.mu.Unlock()

		cm.SendMessage(playerID, 0, "Test", 0, -1)
	}
}

// Helper function for contains check (from chat_test.go)
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && (s[:len(substr)] == substr || containsString(s[1:], substr))))
}
