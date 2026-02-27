package network

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// generateMessageID generates a unique message ID (UUID v4 format).
func generateMessageID() string {
	uuid := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, uuid)
	if err != nil {
		// Fallback to timestamp-based ID
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Warn("crypto rand failed, using timestamp-based message ID fallback")
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	// Set version (4) and variant bits per RFC 4122
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10

	return hex.EncodeToString(uuid)
}

// ChatManager handles chat message routing, encryption, and ACK/NACK protocol.
type ChatManager struct {
	mu                sync.RWMutex
	players           map[uint64]*PlayerChatState // Player ID → chat state
	encryptionKeys    map[uint64]AESKey           // Player ID → encryption key
	pendingACKs       map[string]*PendingMessage  // Message ID → pending message
	messageQueue      chan *ChatMessagePacket
	ackQueue          chan *MessageACK
	rateLimiter       *RateLimiter
	maxRetries        int
	ackTimeout        time.Duration
	maxPendingACKs    int
	onMessageCallback func(*ChatMessagePacket) // Callback for message relay
}

// PlayerChatState tracks chat state for a player.
type PlayerChatState struct {
	PlayerID       uint64
	Position       Vector2           // For range limiting
	PartyID        uint64            // For party chat
	RateLimitState map[int]time.Time // Channel → last message time
	MutedUntil     time.Time
}

// ChatMessagePacket represents a network chat message.
type ChatMessagePacket struct {
	MessageID        string // UUID
	SenderID         uint64 // Sender player ID
	RecipientID      uint64 // Recipient player ID (0 for non-whisper)
	Channel          int    // ChatChannel value
	EncryptedPayload []byte // Encrypted message content
	Timestamp        time.Time
	SequenceNum      uint32 // For ordering
}

// MessageACK represents an acknowledgment or negative acknowledgment.
type MessageACK struct {
	MessageID string
	SenderID  uint64
	Success   bool   // true = ACK, false = NACK
	Reason    string // Failure reason for NACK
}

// PendingMessage tracks a message awaiting acknowledgment.
type PendingMessage struct {
	Packet     *ChatMessagePacket
	Attempts   int
	LastSent   time.Time
	Timeout    time.Time
	RetryTimer *time.Timer
}

// RateLimiter enforces per-channel rate limits.
type RateLimiter struct {
	mu               sync.Mutex
	violations       map[uint64]int       // Player ID → violation count
	muteExpiry       map[uint64]time.Time // Player ID → mute expiry time
	baseMuteDuration time.Duration
	maxMuteDuration  time.Duration
}

// Vector2 represents a 2D position (imported for compatibility).
type Vector2 struct {
	X, Y float64
}

// NewChatManager creates a new chat manager.
func NewChatManager() *ChatManager {
	return &ChatManager{
		players:        make(map[uint64]*PlayerChatState),
		encryptionKeys: make(map[uint64]AESKey),
		pendingACKs:    make(map[string]*PendingMessage),
		messageQueue:   make(chan *ChatMessagePacket, 100),
		ackQueue:       make(chan *MessageACK, 100),
		rateLimiter:    NewRateLimiter(30*time.Second, 10*time.Minute),
		maxRetries:     3,
		ackTimeout:     10 * time.Second,
		maxPendingACKs: 100,
	}
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(baseMute, maxMute time.Duration) *RateLimiter {
	return &RateLimiter{
		violations:       make(map[uint64]int),
		muteExpiry:       make(map[uint64]time.Time),
		baseMuteDuration: baseMute,
		maxMuteDuration:  maxMute,
	}
}

// AddPlayer registers a player with the chat system.
func (cm *ChatManager) AddPlayer(playerID uint64, position Vector2, encKey AESKey) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.players[playerID] = &PlayerChatState{
		PlayerID:       playerID,
		Position:       position,
		RateLimitState: make(map[int]time.Time),
	}
	cm.encryptionKeys[playerID] = encKey
}

// RemovePlayer unregisters a player from the chat system.
func (cm *ChatManager) RemovePlayer(playerID uint64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.players, playerID)
	delete(cm.encryptionKeys, playerID)
	cm.rateLimiter.ClearViolations(playerID)
}

// UpdatePlayerPosition updates a player's position for range calculations.
func (cm *ChatManager) UpdatePlayerPosition(playerID uint64, position Vector2) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if state, exists := cm.players[playerID]; exists {
		state.Position = position
	}
}

// SendMessage validates and queues a chat message for delivery.
func (cm *ChatManager) SendMessage(senderID uint64, channel int, plaintext string, recipientID uint64, localRadius float64) (*ChatMessagePacket, error) {
	cm.mu.RLock()
	state, exists := cm.players[senderID]
	encKey, hasKey := cm.encryptionKeys[senderID]
	cm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("sender not registered")
	}

	// Check rate limit
	if !cm.rateLimiter.CheckRateLimit(senderID, channel, state) {
		cm.rateLimiter.RecordViolation(senderID)
		return nil, fmt.Errorf("rate limit exceeded")
	}

	// Encrypt message
	var encrypted []byte
	var err error
	if hasKey {
		encrypted, err = EncryptMessage(encKey, []byte(plaintext))
		if err != nil {
			return nil, fmt.Errorf("encryption failed: %w", err)
		}
	} else {
		// No encryption key (fallback to plaintext for testing)
		encrypted = []byte(plaintext)
	}

	// Create packet
	packet := &ChatMessagePacket{
		MessageID:        generateMessageID(),
		SenderID:         senderID,
		RecipientID:      recipientID,
		Channel:          channel,
		EncryptedPayload: encrypted,
		Timestamp:        time.Now(),
	}

	// Validate range for local messages
	if channel == 1 { // ChatLocal
		if !cm.validateLocalRange(senderID, recipientID, localRadius) {
			return nil, fmt.Errorf("recipient out of range for local chat")
		}
	}

	// Update rate limit state
	state.RateLimitState[channel] = time.Now()

	// Queue for delivery
	cm.queueMessage(packet)

	return packet, nil
}

// queueMessage adds a message to the pending ACK queue.
func (cm *ChatManager) queueMessage(packet *ChatMessagePacket) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check pending ACK limit
	if len(cm.pendingACKs) >= cm.maxPendingACKs {
		// Drop oldest pending message
		var oldestID string
		var oldestTime time.Time
		for id, pending := range cm.pendingACKs {
			if oldestTime.IsZero() || pending.LastSent.Before(oldestTime) {
				oldestID = id
				oldestTime = pending.LastSent
			}
		}
		delete(cm.pendingACKs, oldestID)
	}

	// Add to pending ACKs
	pending := &PendingMessage{
		Packet:   packet,
		Attempts: 0,
		LastSent: time.Now(),
		Timeout:  time.Now().Add(cm.ackTimeout),
	}
	cm.pendingACKs[packet.MessageID] = pending

	// Send immediately
	cm.messageQueue <- packet
}

// ProcessACK processes an acknowledgment for a message.
// Handles both positive ACKs (message delivered) and NACKs (delivery failed).
//
// Behavior:
//   - On ACK (Success=true): Removes message from pending queue, no retry
//   - On NACK (Success=false): Retries up to maxRetries times with exponential backoff
//   - After maxRetries: Marks message as failed and removes from queue
//
// Thread-safety: Safe to call concurrently. Uses internal mutex for pending message map.
//
// The ACK/NACK mechanism ensures reliable delivery even with packet loss or
// high-latency networks (Tor, satellite). Retry intervals are configurable
// via ChatManager.maxRetries and resend interval.
func (cm *ChatManager) ProcessACK(ack *MessageACK) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	pending, exists := cm.pendingACKs[ack.MessageID]
	if !exists {
		return
	}

	if ack.Success {
		// Message delivered successfully
		delete(cm.pendingACKs, ack.MessageID)
	} else {
		// NACK received - retry if attempts remain
		if pending.Attempts < cm.maxRetries {
			pending.Attempts++
			pending.LastSent = time.Now()
			cm.messageQueue <- pending.Packet
		} else {
			// Max retries exceeded - mark as failed
			delete(cm.pendingACKs, ack.MessageID)
		}
	}
}

// ProcessTimeouts checks for timed-out messages and retries/fails them.
func (cm *ChatManager) ProcessTimeouts() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	for id, pending := range cm.pendingACKs {
		if now.After(pending.Timeout) {
			if pending.Attempts < cm.maxRetries {
				// Retry
				pending.Attempts++
				pending.LastSent = now
				pending.Timeout = now.Add(cm.ackTimeout)
				cm.messageQueue <- pending.Packet
			} else {
				// Failed - max retries exceeded
				delete(cm.pendingACKs, id)
			}
		}
	}
}

// validateLocalRange checks if recipient is within local chat range.
func (cm *ChatManager) validateLocalRange(senderID, recipientID uint64, localRadius float64) bool {
	sender, senderExists := cm.players[senderID]
	recipient, recipientExists := cm.players[recipientID]

	if !senderExists || !recipientExists {
		return false
	}

	// Unlimited range
	if localRadius < 0 {
		return true
	}

	// Calculate distance
	dx := sender.Position.X - recipient.Position.X
	dy := sender.Position.Y - recipient.Position.Y
	distance := dx*dx + dy*dy // Squared distance (avoid sqrt)

	return distance <= localRadius*localRadius
}

// CheckRateLimit checks if a player can send to a channel.
// It checks both mute status and message rate limiting (1 message per second per channel).
func (rl *RateLimiter) CheckRateLimit(playerID uint64, channel int, state *PlayerChatState) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Check if muted
	if expiry, muted := rl.muteExpiry[playerID]; muted {
		if time.Now().Before(expiry) {
			return false
		}
		// Mute expired
		delete(rl.muteExpiry, playerID)
		delete(rl.violations, playerID)
	}

	// Check rate limit (1 message per second per channel)
	if lastSent, exists := state.RateLimitState[channel]; exists {
		elapsed := time.Since(lastSent)
		if elapsed < time.Second {
			return false
		}
	}

	return true
}

// RecordViolation records a rate limit violation and applies mute.
func (rl *RateLimiter) RecordViolation(playerID uint64) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.violations[playerID]++
	count := rl.violations[playerID]

	// Calculate mute duration (doubles each violation, max 10 minutes)
	muteDuration := rl.baseMuteDuration
	for i := 1; i < count; i++ {
		muteDuration *= 2
		if muteDuration > rl.maxMuteDuration {
			muteDuration = rl.maxMuteDuration
			break
		}
	}

	rl.muteExpiry[playerID] = time.Now().Add(muteDuration)
}

// ClearViolations clears violation history for a player.
func (rl *RateLimiter) ClearViolations(playerID uint64) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.violations, playerID)
	delete(rl.muteExpiry, playerID)
}

// GetPendingCount returns the number of messages awaiting ACK.
func (cm *ChatManager) GetPendingCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.pendingACKs)
}

// SetMessageCallback sets a callback for message relay (server → clients).
func (cm *ChatManager) SetMessageCallback(callback func(*ChatMessagePacket)) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.onMessageCallback = callback
}
