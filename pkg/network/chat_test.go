package network

import (
	"math"
	"math/big"
	"testing"
	"time"
)

// TestGenerateMessageID tests message ID generation
func TestGenerateMessageID(t *testing.T) {
	// Generate multiple IDs and verify uniqueness
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateMessageID()
		if id == "" {
			t.Error("Generated empty message ID")
		}
		if ids[id] {
			t.Errorf("Duplicate message ID generated: %s", id)
		}
		ids[id] = true
	}

	// Verify all IDs are unique
	if len(ids) != 100 {
		t.Errorf("Expected 100 unique IDs, got %d", len(ids))
	}
}

// TestNewChatManager tests chat manager creation
func TestNewChatManager(t *testing.T) {
	cm := NewChatManager()
	if cm == nil {
		t.Fatal("NewChatManager returned nil")
	}

	if cm.players == nil {
		t.Error("players map not initialized")
	}
	if cm.encryptionKeys == nil {
		t.Error("encryptionKeys map not initialized")
	}
	if cm.pendingACKs == nil {
		t.Error("pendingACKs map not initialized")
	}
	if cm.messageQueue == nil {
		t.Error("messageQueue channel not initialized")
	}
	if cm.ackQueue == nil {
		t.Error("ackQueue channel not initialized")
	}
	if cm.rateLimiter == nil {
		t.Error("rateLimiter not initialized")
	}

	// Verify configuration
	if cm.maxRetries != 3 {
		t.Errorf("Expected maxRetries = 3, got %d", cm.maxRetries)
	}
	if cm.ackTimeout != 10*time.Second {
		t.Errorf("Expected ackTimeout = 10s, got %v", cm.ackTimeout)
	}
	if cm.maxPendingACKs != 100 {
		t.Errorf("Expected maxPendingACKs = 100, got %d", cm.maxPendingACKs)
	}
}

// TestAddRemovePlayer tests player registration
func TestAddRemovePlayer(t *testing.T) {
	cm := NewChatManager()
	playerID := uint64(123)
	position := Vector2{X: 10.0, Y: 20.0}
	encKey := DeriveAESKey(big.NewInt(456))

	// Add player
	cm.AddPlayer(playerID, position, encKey)

	cm.mu.RLock()
	state, exists := cm.players[playerID]
	key, hasKey := cm.encryptionKeys[playerID]
	cm.mu.RUnlock()

	if !exists {
		t.Error("Player not added to players map")
	}
	if !hasKey {
		t.Error("Encryption key not added")
	}

	// Verify state
	if state.PlayerID != playerID {
		t.Errorf("Expected PlayerID %d, got %d", playerID, state.PlayerID)
	}
	if state.Position != position {
		t.Errorf("Expected position %v, got %v", position, state.Position)
	}
	if key != encKey {
		t.Error("Encryption key mismatch")
	}

	// Remove player
	cm.RemovePlayer(playerID)

	cm.mu.RLock()
	_, existsAfter := cm.players[playerID]
	_, hasKeyAfter := cm.encryptionKeys[playerID]
	cm.mu.RUnlock()

	if existsAfter {
		t.Error("Player not removed from players map")
	}
	if hasKeyAfter {
		t.Error("Encryption key not removed")
	}
}

// TestUpdatePlayerPosition tests position updates
func TestUpdatePlayerPosition(t *testing.T) {
	cm := NewChatManager()
	playerID := uint64(789)
	initialPos := Vector2{X: 5.0, Y: 10.0}
	newPos := Vector2{X: 15.0, Y: 25.0}
	encKey := DeriveAESKey(big.NewInt(100))

	// Add player
	cm.AddPlayer(playerID, initialPos, encKey)

	// Update position
	cm.UpdatePlayerPosition(playerID, newPos)

	cm.mu.RLock()
	state := cm.players[playerID]
	cm.mu.RUnlock()

	if state.Position != newPos {
		t.Errorf("Expected position %v, got %v", newPos, state.Position)
	}

	// Update position for non-existent player (should not panic)
	cm.UpdatePlayerPosition(uint64(999), newPos)
}

// TestSendMessageEncryption tests message encryption
func TestSendMessageEncryption(t *testing.T) {
	cm := NewChatManager()
	senderID := uint64(111)
	position := Vector2{X: 0.0, Y: 0.0}

	// Generate encryption key
	params := DefaultDHParams()
	keyPair, _ := GenerateKeyPair(params)
	secret, _ := ComputeSharedSecret(keyPair.PrivateKey, keyPair.PublicKey, params)
	encKey := DeriveAESKey(secret)

	cm.AddPlayer(senderID, position, encKey)

	// Send message
	plaintext := "Hello, encrypted world!"
	packet, err := cm.SendMessage(senderID, 0, plaintext, 0, -1) // ChatGlobal, unlimited range
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	// Verify encrypted payload is different from plaintext
	if string(packet.EncryptedPayload) == plaintext {
		t.Error("Payload not encrypted (matches plaintext)")
	}

	// Decrypt and verify
	decrypted, err := DecryptMessage(encKey, packet.EncryptedPayload)
	if err != nil {
		t.Fatalf("Failed to decrypt message: %v", err)
	}
	if string(decrypted) != plaintext {
		t.Errorf("Decrypted text mismatch. Expected %q, got %q", plaintext, string(decrypted))
	}
}

// TestSendMessageRateLimit tests rate limiting
func TestSendMessageRateLimit(t *testing.T) {
	cm := NewChatManager()
	senderID := uint64(222)
	position := Vector2{X: 0.0, Y: 0.0}
	encKey := DeriveAESKey(big.NewInt(333))

	cm.AddPlayer(senderID, position, encKey)

	// First message should succeed
	_, err1 := cm.SendMessage(senderID, 0, "Message 1", 0, -1)
	if err1 != nil {
		t.Errorf("First message failed: %v", err1)
	}

	// Immediately send second message (should violate rate limit)
	_, err2 := cm.SendMessage(senderID, 0, "Message 2", 0, -1)
	if err2 == nil {
		t.Error("Expected rate limit error, got nil")
	}
}

// TestSendMessageUnregisteredSender tests sending from unregistered player
func TestSendMessageUnregisteredSender(t *testing.T) {
	cm := NewChatManager()
	_, err := cm.SendMessage(uint64(999), 0, "Test", 0, -1)

	if err == nil {
		t.Error("Expected error for unregistered sender, got nil")
	}
}

// TestValidateLocalRange tests local chat range validation
func TestValidateLocalRange(t *testing.T) {
	cm := NewChatManager()
	encKey := DeriveAESKey(big.NewInt(444))

	sender := uint64(1)
	recipient1 := uint64(2) // Within range
	recipient2 := uint64(3) // Out of range

	cm.AddPlayer(sender, Vector2{X: 0, Y: 0}, encKey)
	cm.AddPlayer(recipient1, Vector2{X: 5, Y: 0}, encKey)  // Distance = 5
	cm.AddPlayer(recipient2, Vector2{X: 20, Y: 0}, encKey) // Distance = 20

	tests := []struct {
		name          string
		senderID      uint64
		recipientID   uint64
		localRadius   float64
		expectInRange bool
	}{
		{"within range", sender, recipient1, 10.0, true},
		{"out of range", sender, recipient2, 10.0, false},
		{"exact range", sender, recipient1, 5.0, true},
		{"unlimited range", sender, recipient2, -1.0, true},
		{"zero range", sender, recipient1, 0.0, false},
		{"non-existent sender", uint64(999), recipient1, 10.0, false},
		{"non-existent recipient", sender, uint64(999), 10.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inRange := cm.validateLocalRange(tt.senderID, tt.recipientID, tt.localRadius)
			if inRange != tt.expectInRange {
				t.Errorf("Expected inRange=%v, got %v", tt.expectInRange, inRange)
			}
		})
	}
}

// TestProcessACK tests ACK processing
func TestProcessACK(t *testing.T) {
	cm := NewChatManager()
	senderID := uint64(555)
	encKey := DeriveAESKey(big.NewInt(666))
	cm.AddPlayer(senderID, Vector2{}, encKey)

	// Send a message
	packet, _ := cm.SendMessage(senderID, 0, "Test message", 0, -1)

	// Verify pending ACK
	initialPending := cm.GetPendingCount()
	if initialPending == 0 {
		t.Error("Expected pending ACK, got 0")
	}

	// Process successful ACK
	ack := &MessageACK{
		MessageID: packet.MessageID,
		SenderID:  senderID,
		Success:   true,
	}
	cm.ProcessACK(ack)

	// Verify message removed from pending
	afterPending := cm.GetPendingCount()
	if afterPending >= initialPending {
		t.Errorf("Expected pending count to decrease, got %d (was %d)", afterPending, initialPending)
	}
}

// TestProcessNACK tests NACK retries
func TestProcessNACK(t *testing.T) {
	cm := NewChatManager()
	senderID := uint64(777)
	encKey := DeriveAESKey(big.NewInt(888))
	cm.AddPlayer(senderID, Vector2{}, encKey)

	// Send a message
	packet, _ := cm.SendMessage(senderID, 0, "Test message", 0, -1)

	// Process NACK
	nack := &MessageACK{
		MessageID: packet.MessageID,
		SenderID:  senderID,
		Success:   false,
		Reason:    "Test failure",
	}
	cm.ProcessACK(nack)

	// Verify message still pending (retry)
	cm.mu.RLock()
	pending, exists := cm.pendingACKs[packet.MessageID]
	cm.mu.RUnlock()

	if !exists {
		t.Error("Message removed after first NACK (should retry)")
	}
	if pending.Attempts != 1 {
		t.Errorf("Expected 1 attempt after NACK, got %d", pending.Attempts)
	}
}

// TestProcessTimeouts tests timeout handling and retries
func TestProcessTimeouts(t *testing.T) {
	cm := NewChatManager()
	cm.ackTimeout = 100 * time.Millisecond // Short timeout for testing

	senderID := uint64(999)
	encKey := DeriveAESKey(big.NewInt(1000))
	cm.AddPlayer(senderID, Vector2{}, encKey)

	// Send a message
	packet, _ := cm.SendMessage(senderID, 0, "Timeout test", 0, -1)

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Process timeouts
	cm.ProcessTimeouts()

	// Verify retry attempt incremented
	cm.mu.RLock()
	pending, exists := cm.pendingACKs[packet.MessageID]
	cm.mu.RUnlock()

	if !exists {
		t.Error("Message removed after timeout (should retry)")
	}
	if pending.Attempts != 1 {
		t.Errorf("Expected 1 attempt after timeout, got %d", pending.Attempts)
	}
}

// TestMaxRetries tests maximum retry limit
func TestMaxRetries(t *testing.T) {
	cm := NewChatManager()
	cm.maxRetries = 2 // Low limit for testing

	senderID := uint64(1111)
	encKey := DeriveAESKey(big.NewInt(1112))
	cm.AddPlayer(senderID, Vector2{}, encKey)

	// Send a message
	packet, _ := cm.SendMessage(senderID, 0, "Max retries test", 0, -1)

	// Process max retries worth of NACKs
	nack := &MessageACK{
		MessageID: packet.MessageID,
		SenderID:  senderID,
		Success:   false,
		Reason:    "Persistent failure",
	}

	for i := 0; i < cm.maxRetries; i++ {
		cm.ProcessACK(nack)
	}

	// Process one more NACK (exceeds max retries)
	cm.ProcessACK(nack)

	// Verify message removed after max retries
	cm.mu.RLock()
	_, exists := cm.pendingACKs[packet.MessageID]
	cm.mu.RUnlock()

	if exists {
		t.Error("Message still pending after max retries exceeded")
	}
}

// TestRateLimiterCheckRateLimit tests rate limit checking
func TestRateLimiterCheckRateLimit(t *testing.T) {
	rl := NewRateLimiter(30*time.Second, 10*time.Minute)
	playerID := uint64(123)
	channel := 0

	// Create dummy player state
	state := &PlayerChatState{
		PlayerID:       playerID,
		RateLimitState: make(map[int]time.Time),
	}

	// First check should pass
	if !rl.CheckRateLimit(playerID, channel, state) {
		t.Error("First rate limit check failed")
	}

	// Simulate violation
	rl.RecordViolation(playerID)

	// Next check should fail (muted)
	if rl.CheckRateLimit(playerID, channel, state) {
		t.Error("Rate limit check passed when player should be muted")
	}
}

// TestRateLimiterViolationDuration tests mute duration doubling
func TestRateLimiterViolationDuration(t *testing.T) {
	baseDuration := 100 * time.Millisecond
	maxDuration := 1 * time.Second
	rl := NewRateLimiter(baseDuration, maxDuration)
	playerID := uint64(456)

	// First violation: 100ms mute
	rl.RecordViolation(playerID)
	rl.mu.Lock()
	expiry1 := rl.muteExpiry[playerID]
	rl.mu.Unlock()

	// Second violation: 200ms mute
	time.Sleep(150 * time.Millisecond) // Wait for first mute to expire
	rl.RecordViolation(playerID)
	rl.mu.Lock()
	expiry2 := rl.muteExpiry[playerID]
	count := rl.violations[playerID]
	rl.mu.Unlock()

	if count != 2 {
		t.Errorf("Expected violation count 2, got %d", count)
	}

	// Verify second mute is longer
	if !expiry2.After(expiry1) {
		t.Error("Second mute expiry not later than first")
	}
}

// TestRateLimiterClearViolations tests violation clearing
func TestRateLimiterClearViolations(t *testing.T) {
	rl := NewRateLimiter(30*time.Second, 10*time.Minute)
	playerID := uint64(789)

	// Record violations
	rl.RecordViolation(playerID)
	rl.RecordViolation(playerID)

	// Clear violations
	rl.ClearViolations(playerID)

	rl.mu.Lock()
	violations := rl.violations[playerID]
	_, muted := rl.muteExpiry[playerID]
	rl.mu.Unlock()

	if violations != 0 {
		t.Errorf("Expected 0 violations after clear, got %d", violations)
	}
	if muted {
		t.Error("Player still muted after clear")
	}
}

// TestChatMessagePacket tests packet structure
func TestChatMessagePacket(t *testing.T) {
	packet := &ChatMessagePacket{
		MessageID:        "test-uuid-1234",
		SenderID:         uint64(111),
		RecipientID:      uint64(222),
		Channel:          0,
		EncryptedPayload: []byte("encrypted data"),
		Timestamp:        time.Now(),
		SequenceNum:      1,
	}

	if packet.MessageID != "test-uuid-1234" {
		t.Errorf("MessageID mismatch")
	}
	if packet.SenderID != 111 {
		t.Errorf("SenderID mismatch")
	}
	if packet.RecipientID != 222 {
		t.Errorf("RecipientID mismatch")
	}
	if packet.Channel != 0 {
		t.Errorf("Channel mismatch")
	}
	if packet.SequenceNum != 1 {
		t.Errorf("SequenceNum mismatch")
	}
}

// TestPendingACKLimit tests max pending ACK limit enforcement
func TestPendingACKLimit(t *testing.T) {
	cm := NewChatManager()
	cm.maxPendingACKs = 5 // Low limit for testing

	senderID := uint64(1234)
	encKey := DeriveAESKey(big.NewInt(1235))
	cm.AddPlayer(senderID, Vector2{}, encKey)

	// Send messages up to limit
	for i := 0; i < 10; i++ {
		cm.SendMessage(senderID, 0, "Test message", 0, -1)

		// Temporarily allow by clearing rate limit
		cm.mu.Lock()
		if state, exists := cm.players[senderID]; exists {
			delete(state.RateLimitState, 0)
		}
		cm.mu.Unlock()
	}

	// Verify pending count capped at maxPendingACKs
	count := cm.GetPendingCount()
	if count > cm.maxPendingACKs {
		t.Errorf("Pending count %d exceeds max %d", count, cm.maxPendingACKs)
	}
}

// BenchmarkSendMessage benchmarks message sending
func BenchmarkSendMessage(b *testing.B) {
	cm := NewChatManager()
	senderID := uint64(1)
	encKey := DeriveAESKey(big.NewInt(2))
	cm.AddPlayer(senderID, Vector2{}, encKey)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clear rate limit for benchmark
		cm.mu.Lock()
		if state, exists := cm.players[senderID]; exists {
			delete(state.RateLimitState, 0)
		}
		cm.mu.Unlock()

		cm.SendMessage(senderID, 0, "Benchmark message", 0, -1)
	}
}

// BenchmarkEncryptDecrypt benchmarks E2E encryption cycle
func BenchmarkEncryptDecrypt(b *testing.B) {
	params := DefaultDHParams()
	alice, _ := GenerateKeyPair(params)
	bob, _ := GenerateKeyPair(params)
	secret, _ := ComputeSharedSecret(alice.PrivateKey, bob.PublicKey, params)
	key := DeriveAESKey(secret)
	plaintext := []byte("Hello, World! This is a test message for benchmarking.")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ciphertext, _ := EncryptMessage(key, plaintext)
		DecryptMessage(key, ciphertext)
	}
}

// BenchmarkRateLimitCheck benchmarks rate limit checking
func BenchmarkRateLimitCheck(b *testing.B) {
	rl := NewRateLimiter(30*time.Second, 10*time.Minute)
	playerID := uint64(123)
	channel := 0

	// Create dummy player state
	state := &PlayerChatState{
		PlayerID:       playerID,
		RateLimitState: make(map[int]time.Time),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.CheckRateLimit(playerID, channel, state)
	}
}

// TestSetMessageCallback tests message callback registration
func TestSetMessageCallback(t *testing.T) {
	cm := NewChatManager()
	callbackInvoked := false

	callback := func(packet *ChatMessagePacket) {
		callbackInvoked = true
	}

	cm.SetMessageCallback(callback)

	// Verify callback set
	if cm.onMessageCallback == nil {
		t.Error("Callback not set")
	}

	// Invoke callback
	if cm.onMessageCallback != nil {
		cm.onMessageCallback(&ChatMessagePacket{})
	}

	if !callbackInvoked {
		t.Error("Callback not invoked")
	}
}

// TestValidateLocalRangeEdgeCases tests edge cases for range validation
func TestValidateLocalRangeEdgeCases(t *testing.T) {
	cm := NewChatManager()
	encKey := DeriveAESKey(big.NewInt(555))

	// Add players at various positions
	cm.AddPlayer(1, Vector2{X: 0, Y: 0}, encKey)
	cm.AddPlayer(2, Vector2{X: math.MaxFloat64 / 2, Y: 0}, encKey)

	tests := []struct {
		name        string
		sender      uint64
		recipient   uint64
		radius      float64
		expectValid bool
	}{
		{"same position", 1, 1, 10.0, true},
		{"very large distance", 1, 2, 100.0, false},
		{"negative radius (unlimited)", 1, 2, -1.0, true},
		{"float precision test", 1, 1, 0.000001, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cm.validateLocalRange(tt.sender, tt.recipient, tt.radius)
			if result != tt.expectValid {
				t.Errorf("Expected %v, got %v", tt.expectValid, result)
			}
		})
	}
}
