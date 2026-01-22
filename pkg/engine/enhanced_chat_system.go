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
	world         *World
	history       map[uint64]*persistence.ChatHistory // Player ID → chat history
	nextMsgID     uint64                              // For generating unique message IDs
	encryptedMsgs map[string][]byte                   // Message ID → encrypted payload (simulated encryption)
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
	sender, chatComp, err := ecs.validateSender(senderID, channel, recipientID)
	if err != nil {
		return err
	}

	msgID := ecs.generateMessageID()
	encryptedPayload := ecs.encryptMessage(msgID, content)
	chatComp.RecordMessageSent(channel)

	msg := ecs.createChatMessage(msgID, senderID, sender, recipientID, channel, content, encryptedPayload)
	chatComp.AddMessage(msg)

	ecs.persistMessage(senderID, recipientID, channel, content, msg)
	ecs.deliverMessage(msg, sender, chatComp)

	return nil
}

// validateSender validates the sender entity and chat component for message sending.
func (ecs *EnhancedChatSystem) validateSender(senderID uint64, channel ChatChannel, recipientID uint64) (*Entity, *ChatComponent, error) {
	sender, exists := ecs.world.GetEntity(senderID)
	if !exists {
		return nil, nil, fmt.Errorf("sender entity not found")
	}

	chatCompRaw, exists := sender.GetComponent("chat")
	if !exists {
		return nil, nil, fmt.Errorf("sender has no chat component")
	}

	chatComp, ok := chatCompRaw.(*ChatComponent)
	if !ok {
		return nil, nil, fmt.Errorf("invalid chat component type")
	}

	if err := ecs.validateChannelPermissions(chatComp, channel, recipientID); err != nil {
		return nil, nil, err
	}

	return sender, chatComp, nil
}

// validateChannelPermissions checks if the sender can send to the channel.
func (ecs *EnhancedChatSystem) validateChannelPermissions(chatComp *ChatComponent, channel ChatChannel, recipientID uint64) error {
	if !chatComp.CanSendMessage(channel) {
		if chatComp.IsMuted() {
			return fmt.Errorf("sender is muted")
		}
		return fmt.Errorf("rate limit exceeded for channel %s", channel.String())
	}

	if !chatComp.IsChannelActive(channel) {
		return fmt.Errorf("not subscribed to channel %s", channel.String())
	}

	if channel == ChatWhisper && recipientID == 0 {
		return fmt.Errorf("whisper requires recipient ID")
	}

	return nil
}

// encryptMessage encrypts the message content and stores it.
func (ecs *EnhancedChatSystem) encryptMessage(msgID, content string) []byte {
	encryptedPayload := ecs.simpleEncrypt(content)
	ecs.encryptedMsgs[msgID] = encryptedPayload
	return encryptedPayload
}

// createChatMessage constructs a ChatMessage struct with all required fields.
func (ecs *EnhancedChatSystem) createChatMessage(msgID string, senderID uint64, sender *Entity, recipientID uint64, channel ChatChannel, content string, encrypted []byte) ChatMessage {
	return ChatMessage{
		ID:          msgID,
		SenderID:    senderID,
		SenderName:  ecs.getSenderName(sender),
		RecipientID: recipientID,
		Channel:     channel,
		Content:     content,
		Encrypted:   encrypted,
		Timestamp:   time.Now(),
		Delivered:   false,
		Failed:      false,
	}
}

// persistMessage adds the message to persistent storage if available.
func (ecs *EnhancedChatSystem) persistMessage(senderID, recipientID uint64, channel ChatChannel, content string, msg ChatMessage) {
	hist, exists := ecs.history[senderID]
	if !exists {
		return
	}

	persistMsg := &persistence.Message{
		ID:        msg.ID,
		Sender:    msg.SenderName,
		Recipient: fmt.Sprintf("%d", recipientID),
		Channel:   channel.String(),
		Content:   content,
		Timestamp: msg.Timestamp,
	}

	if err := hist.AddMessage(persistMsg); err != nil {
		fmt.Printf("failed to persist message: %v\n", err)
	}
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
	senderPos := ecs.extractSenderPosition(sender)
	if senderPos == nil {
		return
	}

	radius := senderChat.GetEffectiveRadius()
	radiusSquared := radius * radius

	for _, entity := range ecs.world.GetEntities() {
		if entity.ID == sender.ID {
			continue
		}

		chatComp := ecs.extractRecipientChat(entity)
		if chatComp == nil {
			continue
		}

		recipientPos := ecs.extractRecipientPosition(entity)
		if recipientPos == nil {
			continue
		}

		if ecs.shouldDeliverMessage(senderPos, recipientPos, radius, radiusSquared) {
			ecs.deliverMessageToRecipient(msg, entity, chatComp)
		}
	}
}

// extractSenderPosition extracts and validates sender position component.
func (ecs *EnhancedChatSystem) extractSenderPosition(sender *Entity) *PositionComponent {
	senderPosRaw, exists := sender.GetComponent("position")
	if !exists {
		return nil
	}

	senderPos, ok := senderPosRaw.(*PositionComponent)
	if !ok {
		return nil
	}
	return senderPos
}

// extractRecipientChat extracts and validates recipient chat component.
func (ecs *EnhancedChatSystem) extractRecipientChat(entity *Entity) *ChatComponent {
	chatCompRaw, exists := entity.GetComponent("chat")
	if !exists {
		return nil
	}

	chatComp, ok := chatCompRaw.(*ChatComponent)
	if !ok || !chatComp.IsChannelActive(ChatLocal) {
		return nil
	}
	return chatComp
}

// extractRecipientPosition extracts and validates recipient position component.
func (ecs *EnhancedChatSystem) extractRecipientPosition(entity *Entity) *PositionComponent {
	recipientPosRaw, exists := entity.GetComponent("position")
	if !exists {
		return nil
	}

	recipientPos, ok := recipientPosRaw.(*PositionComponent)
	if !ok {
		return nil
	}
	return recipientPos
}

// shouldDeliverMessage checks if message should be delivered based on range.
func (ecs *EnhancedChatSystem) shouldDeliverMessage(senderPos, recipientPos *PositionComponent, radius, radiusSquared float64) bool {
	if radius < 0 {
		return true
	}

	dx := senderPos.X - recipientPos.X
	dy := senderPos.Y - recipientPos.Y
	distSquared := dx*dx + dy*dy
	return distSquared <= radiusSquared
}

// deliverMessageToRecipient delivers message to recipient and persists to history.
func (ecs *EnhancedChatSystem) deliverMessageToRecipient(msg ChatMessage, entity *Entity, chatComp *ChatComponent) {
	deliveredMsg := msg
	deliveredMsg.Delivered = true
	chatComp.AddMessage(deliveredMsg)

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
