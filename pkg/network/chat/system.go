// Package chat provides player-to-player chat functionality with validation
// and rate limiting. This package acts as a network wrapper around the engine
// chat system, adding message validation and DoS protection.
//
// Note: This package uses time.Now() for message timestamps, which is intentional
// and exempt from the deterministic-procgen rule. Network chat messages inherently
// require real timestamps for multiplayer synchronization and message ordering.
package chat

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/errors"
	"github.com/opd-ai/venture/pkg/validation"
	log "github.com/sirupsen/logrus"
)

const (
	// ChatRateLimit is the maximum number of messages per player per second
	// This prevents spam and DoS attacks while allowing normal chat flow
	ChatRateLimit = 10

	// ChatRateLimitWindow is the time window for rate limiting
	ChatRateLimitWindow = time.Second
)

// ChatSystem manages chat messages and channels with validation and rate limiting
type ChatSystem struct {
	// world is the ECS world used to look up sender entities and attach chat components
	world *engine.World
	// validator sanitizes and validates chat message content before delivery
	validator *validation.ChatValidator
	// limiter enforces per-player message rate limits to prevent spam and DoS
	limiter *validation.RateLimiter
}

// NewChatSystem creates a new chat system with validation and rate limiting
func NewChatSystem(world *engine.World) *ChatSystem {
	return &ChatSystem{
		world:     world,
		validator: validation.NewChatValidator(),
		limiter:   validation.NewRateLimiter(ChatRateLimit, ChatRateLimitWindow),
	}
}

// Update processes chat message delivery
func (s *ChatSystem) Update(deltaTime float64) {
	// Chat is event-driven, not time-based
}

// SendMessage sends a chat message with validation and rate limiting
func (s *ChatSystem) SendMessage(senderID uint64, channel engine.ChatChannel, content string) error {
	// Generate correlation ID for tracing this message send operation
	correlationID := errors.NewCorrelationID()

	// Check rate limit
	if !s.limiter.Allow(senderID) {
		log.WithFields(log.Fields{
			"playerID":      senderID,
			"channel":       channel,
			"limit":         ChatRateLimit,
			"correlationID": correlationID,
		}).Warn("chat rate limit exceeded")
		return errors.RateLimit(fmt.Sprintf("rate limit exceeded (maximum %d messages per second)", ChatRateLimit)).
			WithCorrelationID(correlationID).
			WithContext("playerID", senderID).
			WithContext("channel", channel.String()).
			WithContext("limit", ChatRateLimit)
	}

	// Validate and sanitize message
	sanitized, err := s.validator.ValidateAndSanitize(content)
	if err != nil {
		log.WithFields(log.Fields{
			"playerID":      senderID,
			"channel":       channel,
			"error":         err.Error(),
			"correlationID": correlationID,
		}).Warn("chat message validation failed")
		return errors.ValidationWrap(err, "message validation failed").
			WithCorrelationID(correlationID).
			WithContext("playerID", senderID).
			WithContext("channel", channel.String()).
			WithContext("contentLength", len(content))
	}

	// Generate message ID
	msgID, err := generateMessageID()
	if err != nil {
		log.WithFields(log.Fields{
			"playerID":      senderID,
			"error":         err.Error(),
			"correlationID": correlationID,
		}).Error("failed to generate message ID")
		return errors.NetworkWrap(err, "failed to generate message ID").
			WithCorrelationID(correlationID).
			WithContext("playerID", senderID)
	}

	// Create message with sanitized content
	// Note: time.Now() is intentional here - chat messages require real timestamps
	message := engine.ChatMessage{
		ID:        msgID,
		SenderID:  senderID,
		Channel:   channel,
		Content:   sanitized,
		Timestamp: time.Now(),
		Encrypted: nil,
	}

	// Add to sender's chat component
	sender, ok := s.world.GetEntity(senderID)
	if !ok || sender == nil {
		log.WithFields(log.Fields{
			"playerID":      senderID,
			"correlationID": correlationID,
		}).Error("chat sender entity not found")
		return errors.Network("sender entity not found").
			WithCorrelationID(correlationID).
			WithContext("playerID", senderID)
	}

	chatCompRaw, ok := sender.GetComponent("chat")
	var chatComp *engine.ChatComponent
	if !ok {
		// Create chat component using helper for proper defaults
		chatComp = engine.NewChatComponent()
		chatComp.ActiveChannels = append(chatComp.ActiveChannels, channel)
		sender.AddComponent(chatComp)
	} else {
		chatComp, ok = chatCompRaw.(*engine.ChatComponent)
		if !ok {
			log.WithFields(log.Fields{
				"playerID":      senderID,
				"correlationID": correlationID,
			}).Error("chat component type assertion failed")
			return errors.Network("invalid chat component type").
				WithCorrelationID(correlationID).
				WithContext("playerID", senderID)
		}
	}

	chatComp.Messages = append(chatComp.Messages, message)

	// Message broadcasting is handled by the engine chat system

	return nil
}

// generateMessageID creates a collision-resistant message ID using 128 bits
// of cryptographic randomness encoded as URL-safe base64.
func generateMessageID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
