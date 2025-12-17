// Package chat provides player-to-player chat functionality with E2E encryption.
package chat

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
)

// ChatSystem manages chat messages and channels
type ChatSystem struct {
	world *engine.World
}

// NewChatSystem creates a new chat system
func NewChatSystem(world *engine.World) *ChatSystem {
	return &ChatSystem{world: world}
}

// Update processes chat message delivery
func (s *ChatSystem) Update(deltaTime float64) {
	// Chat is event-driven, not time-based
}

// SendMessage sends a chat message
func (s *ChatSystem) SendMessage(senderID uint64, channel engine.ChatChannel, content string) error {
	// Generate message ID
	msgID, err := generateMessageID()
	if err != nil {
		return fmt.Errorf("failed to generate message ID: %w", err)
	}

	// Create message
	message := engine.ChatMessage{
		ID:        msgID,
		SenderID:  senderID,
		Channel:   channel,
		Content:   content,
		Timestamp: time.Now(), // Use current timestamp for message creation
		Encrypted: nil,        // E2E encryption available in pkg/network/chat.go
	}

	// Add to sender's chat component
	sender, ok := s.world.GetEntity(senderID)
	if !ok || sender == nil {
		return fmt.Errorf("sender entity not found")
	}

	chatCompRaw, ok := sender.GetComponent("chat")
	if !ok {
		// Create chat component
		chatComp := &engine.ChatComponent{
			Messages:       []engine.ChatMessage{},
			UnreadCount:    0,
			ActiveChannels: []engine.ChatChannel{channel},
		}
		sender.AddComponent(chatComp)
		chatCompRaw, _ = sender.GetComponent("chat")
	}

	chatComp := chatCompRaw.(*engine.ChatComponent)
	chatComp.Messages = append(chatComp.Messages, message)

	// Message broadcasting and encryption are handled by the main chat system
	// in pkg/network/chat.go which provides full E2E encryption support

	return nil
}

func generateMessageID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
