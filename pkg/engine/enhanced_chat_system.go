package engine

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/opd-ai/venture/pkg/social/persistence"
)

// EnhancedChatSystem provides advanced chat with encryption and history persistence.
// Integrates with persistence.ChatHistory for message storage and retrieval.
type EnhancedChatSystem struct {
	world       *World
	history     map[uint64]*persistence.ChatHistory // Player ID → chat history
	nextMsgID   uint64                              // For generating unique message IDs
	encryptedMsgs map[string][]byte                // Message ID → encrypted payload (simulated encryption)
}

// NewEnhancedChatSystem creates a new enhanced chat system with history persistence.
func NewEnhancedChatSystem(world *World) *EnhancedChatSystem {
	return &EnhancedChatSystem{
		world:         world,
		history:       make(map[uint64]*persistence.ChatHistory),
		nextMsgID:     1,
		encryptedMsgs: make(map[string][]byte),
	}
}

// Update processes pending chat messages (event-driven, minimal per-frame work).
func (ecs *EnhancedChatSystem) Update(deltaTime float64) {
	// Chat is event-driven, no per-frame processing needed
}

// RegisterPlayer registers a player with the chat system and initializes history.
func (ecs *EnhancedChatSystem) RegisterPlayer(playerID uint64) error {
	// Initialize chat history
	ecs.history[playerID] = persistence.NewChatHistory(fmt.Sprintf("%d", playerID))

	// Add ChatComponent to entity if missing
	entity, exists := ecs.world.GetEntity(playerID)
	if !exists {
		return fmt.Errorf("player entity not found")
	}

	if _, hasChat := entity.GetComponent("chat"); !hasChat {
		entity.AddComponent(NewChatComponent())
	}

	return nil
}

// UnregisterPlayer removes a player from the chat system.
func (ecs *EnhancedChatSystem) UnregisterPlayer(playerID uint64) {
	delete(ecs.history, playerID)
}

// generateMessageID generates a unique message ID.
func (ecs *EnhancedChatSystem) generateMessageID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to simple counter
		id := fmt.Sprintf("msg_%d", ecs.nextMsgID)
		ecs.nextMsgID++
		return id
	}
	return base64.URLEncoding.EncodeToString(b)
}

// simpleEncrypt provides basic encryption simulation (XOR with key).
// In production, this would use AES-256-GCM with proper key exchange.
func (ecs *EnhancedChatSystem) simpleEncrypt(plaintext string) []byte {
	// Simple XOR encryption for demonstration
	key := []byte("chat_encryption_key_placeholder_32b")
	encrypted := make([]byte, len(plaintext))
	for i, b := range []byte(plaintext) {
		encrypted[i] = b ^ key[i%len(key)]
	}
	return encrypted
}

// SendMessage sends an encrypted chat message from a player.
// Validates permissions, applies rate limiting, encrypts content, and stores in history.
func (ecs *EnhancedChatSystem) SendMessage(senderID uint64, channel ChatChannel, content string, recipientID uint64) error {
	// Get sender's chat component for validation
	sender, exists := ecs.world.GetEntity(senderID)
	if !exists {
		return fmt.Errorf("sender entity not found")
	}

	chatCompRaw, exists := sender.GetComponent("chat")
	if !exists {
		return fmt.Errorf("sender has no chat component")
	}

	chatComp, ok := chatCompRaw.(*ChatComponent)
	if !ok {
		return fmt.Errorf("invalid chat component type")
	}

	// Validate sender can send to channel
	if !chatComp.CanSendMessage(channel) {
		if chatComp.IsMuted() {
			return fmt.Errorf("sender is muted")
		}
		return fmt.Errorf("rate limit exceeded for channel %s", channel.String())
	}

	// Validate channel subscription
	if !chatComp.IsChannelActive(channel) {
		return fmt.Errorf("not subscribed to channel %s", channel.String())
	}

	// Validate recipient for whispers
	if channel == ChatWhisper && recipientID == 0 {
		return fmt.Errorf("whisper requires recipient ID")
	}

	// Generate message ID
	msgID := ecs.generateMessageID()

	// Encrypt message (simulated)
	encryptedPayload := ecs.simpleEncrypt(content)
	ecs.encryptedMsgs[msgID] = encryptedPayload

	// Update sender's rate limit state
	chatComp.RecordMessageSent(channel)

	// Create message for local history
	msg := ChatMessage{
		ID:          msgID,
		SenderID:    senderID,
		SenderName:  ecs.getSenderName(sender),
		RecipientID: recipientID,
		Channel:     channel,
		Content:     content,
		Encrypted:   encryptedPayload,
		Timestamp:   time.Now(),
		Delivered:   false,
		Failed:      false,
	}

	// Add to sender's component history
	chatComp.AddMessage(msg)

	// Add to persistent history
	if hist, exists := ecs.history[senderID]; exists {
		persistMsg := &persistence.Message{
			ID:        msg.ID,
			Sender:    msg.SenderName,
			Recipient: fmt.Sprintf("%d", recipientID),
			Channel:   channel.String(),
			Content:   content,
			Timestamp: msg.Timestamp,
		}
		if err := hist.AddMessage(persistMsg); err != nil {
			// Log error but don't fail send
			fmt.Printf("failed to persist message: %v\n", err)
		}
	}

	// Deliver to recipients based on channel
	ecs.deliverMessage(msg, sender, chatComp)

	return nil
}

// deliverMessage delivers a message to appropriate recipients based on channel.
func (ecs *EnhancedChatSystem) deliverMessage(msg ChatMessage, sender *Entity, senderChat *ChatComponent) {
	switch msg.Channel {
	case ChatGlobal:
		ecs.deliverToAll(msg, sender.ID)
	case ChatLocal:
		ecs.deliverToLocal(msg, sender, senderChat)
	case ChatParty:
		ecs.deliverToParty(msg, sender)
	case ChatWhisper:
		ecs.deliverToRecipient(msg, msg.RecipientID)
	}
}

// deliverToAll delivers a message to all entities with chat components.
func (ecs *EnhancedChatSystem) deliverToAll(msg ChatMessage, senderID uint64) {
	for _, entity := range ecs.world.GetEntities() {
		if entity.ID == senderID {
			continue
		}

		chatCompRaw, exists := entity.GetComponent("chat")
		if !exists {
			continue
		}

		chatComp, ok := chatCompRaw.(*ChatComponent)
		if !ok || !chatComp.IsChannelActive(ChatGlobal) {
			continue
		}

		deliveredMsg := msg
		deliveredMsg.Delivered = true
		chatComp.AddMessage(deliveredMsg)

		// Add to persistent history
		if hist, exists := ecs.history[entity.ID]; exists {
			persistMsg := &persistence.Message{
				ID:        msg.ID,
				Sender:    msg.SenderName,
				Recipient: fmt.Sprintf("%d", entity.ID),
				Channel:   msg.Channel.String(),
				Content:   msg.Content,
				Timestamp: msg.Timestamp,
			}
			hist.AddMessage(persistMsg)
		}
	}
}

// deliverToLocal delivers a message to entities within range.
func (ecs *EnhancedChatSystem) deliverToLocal(msg ChatMessage, sender *Entity, senderChat *ChatComponent) {
	senderPosRaw, exists := sender.GetComponent("position")
	if !exists {
		return
	}

	senderPos, ok := senderPosRaw.(*PositionComponent)
	if !ok {
		return
	}

	radius := senderChat.GetEffectiveRadius()
	radiusSquared := radius * radius

	for _, entity := range ecs.world.GetEntities() {
		if entity.ID == sender.ID {
			continue
		}

		chatCompRaw, exists := entity.GetComponent("chat")
		if !exists {
			continue
		}

		chatComp, ok := chatCompRaw.(*ChatComponent)
		if !ok || !chatComp.IsChannelActive(ChatLocal) {
			continue
		}

		// Check range
		recipientPosRaw, exists := entity.GetComponent("position")
		if !exists {
			continue
		}

		recipientPos, ok := recipientPosRaw.(*PositionComponent)
		if !ok {
			continue
		}

		// Unlimited range for walkie-talkie
		if radius < 0 {
			deliveredMsg := msg
			deliveredMsg.Delivered = true
			chatComp.AddMessage(deliveredMsg)
			continue
		}

		// Calculate distance
		dx := senderPos.X - recipientPos.X
		dy := senderPos.Y - recipientPos.Y
		distSquared := dx*dx + dy*dy

		if distSquared <= radiusSquared {
			deliveredMsg := msg
			deliveredMsg.Delivered = true
			chatComp.AddMessage(deliveredMsg)

			// Add to persistent history
			if hist, exists := ecs.history[entity.ID]; exists {
				persistMsg := &persistence.Message{
					ID:        msg.ID,
					Sender:    msg.SenderName,
					Recipient: fmt.Sprintf("%d", entity.ID),
					Channel:   msg.Channel.String(),
					Content:   msg.Content,
					Timestamp: msg.Timestamp,
				}
				hist.AddMessage(persistMsg)
			}
		}
	}
}

// deliverToParty delivers a message to party members (placeholder until party system added).
func (ecs *EnhancedChatSystem) deliverToParty(msg ChatMessage, sender *Entity) {
	// Placeholder: deliver to all subscribed entities until party system is implemented
	for _, entity := range ecs.world.GetEntities() {
		if entity.ID == sender.ID {
			continue
		}

		chatCompRaw, exists := entity.GetComponent("chat")
		if !exists {
			continue
		}

		chatComp, ok := chatCompRaw.(*ChatComponent)
		if !ok || !chatComp.IsChannelActive(ChatParty) {
			continue
		}

		deliveredMsg := msg
		deliveredMsg.Delivered = true
		chatComp.AddMessage(deliveredMsg)
	}
}

// deliverToRecipient delivers a whisper to a specific recipient.
func (ecs *EnhancedChatSystem) deliverToRecipient(msg ChatMessage, recipientID uint64) {
	recipient, exists := ecs.world.GetEntity(recipientID)
	if !exists {
		return
	}

	chatCompRaw, exists := recipient.GetComponent("chat")
	if !exists {
		// Create chat component for recipient
		chatCompRaw = NewChatComponent()
		recipient.AddComponent(chatCompRaw)
	}

	chatComp, ok := chatCompRaw.(*ChatComponent)
	if !ok {
		return
	}

	deliveredMsg := msg
	deliveredMsg.Delivered = true
	chatComp.AddMessage(deliveredMsg)

	// Add to persistent history
	if hist, exists := ecs.history[recipientID]; exists {
		persistMsg := &persistence.Message{
			ID:        msg.ID,
			Sender:    msg.SenderName,
			Recipient: fmt.Sprintf("%d", recipientID),
			Channel:   msg.Channel.String(),
			Content:   msg.Content,
			Timestamp: msg.Timestamp,
		}
		hist.AddMessage(persistMsg)
	}
}

// getSenderName retrieves the sender's display name from the entity.
func (ecs *EnhancedChatSystem) getSenderName(entity *Entity) string {
	// Try to get name from NameComponent (if exists)
	if nameComp, exists := entity.GetComponent("name"); exists {
		if nc, ok := nameComp.(interface{ GetName() string }); ok {
			return nc.GetName()
		}
	}
	return fmt.Sprintf("Entity_%d", entity.ID)
}

// GetPlayerHistory retrieves the persistent chat history for a player.
func (ecs *EnhancedChatSystem) GetPlayerHistory(playerID uint64, filter *persistence.MessageFilter) ([]*persistence.Message, error) {
	hist, exists := ecs.history[playerID]
	if !exists {
		return []*persistence.Message{}, nil
	}
	return hist.GetMessages(filter), nil
}

// SaveHistory saves chat history to persistent storage (JSON).
func (ecs *EnhancedChatSystem) SaveHistory(playerID uint64) ([]byte, error) {
	hist, exists := ecs.history[playerID]
	if !exists {
		return nil, fmt.Errorf("no history for player %d", playerID)
	}
	return hist.Save()
}

// LoadHistory loads chat history from persistent storage.
func (ecs *EnhancedChatSystem) LoadHistory(playerID uint64, data []byte) error {
	hist, exists := ecs.history[playerID]
	if !exists {
		hist = persistence.NewChatHistory(fmt.Sprintf("%d", playerID))
		ecs.history[playerID] = hist
	}
	return hist.Load(data)
}

// ProcessACK processes an acknowledgment for a sent message.
func (ecs *EnhancedChatSystem) ProcessACK(msgID string, senderID uint64, success bool, reason string) {
	// Update message delivery status in sender's chat component
	sender, exists := ecs.world.GetEntity(senderID)
	if !exists {
		return
	}

	chatCompRaw, exists := sender.GetComponent("chat")
	if !exists {
		return
	}

	chatComp, ok := chatCompRaw.(*ChatComponent)
	if !ok {
		return
	}

	msg := chatComp.GetMessageByID(msgID)
	if msg != nil {
		if success {
			msg.Delivered = true
		} else {
			msg.Failed = true
		}
	}
}

// ApplyMute mutes a player for the specified duration.
func (ecs *EnhancedChatSystem) ApplyMute(playerID uint64, duration time.Duration) error {
	entity, exists := ecs.world.GetEntity(playerID)
	if !exists {
		return fmt.Errorf("entity not found")
	}

	chatCompRaw, exists := entity.GetComponent("chat")
	if !exists {
		chatCompRaw = NewChatComponent()
		entity.AddComponent(chatCompRaw)
	}

	chatComp, ok := chatCompRaw.(*ChatComponent)
	if !ok {
		return fmt.Errorf("invalid chat component type")
	}

	chatComp.ApplyMute(duration)
	return nil
}

// IsMuted checks if a player is currently muted.
func (ecs *EnhancedChatSystem) IsMuted(playerID uint64) bool {
	entity, exists := ecs.world.GetEntity(playerID)
	if !exists {
		return false
	}

	chatCompRaw, exists := entity.GetComponent("chat")
	if !exists {
		return false
	}

	chatComp, ok := chatCompRaw.(*ChatComponent)
	if !ok {
		return false
	}

	return chatComp.IsMuted()
}
