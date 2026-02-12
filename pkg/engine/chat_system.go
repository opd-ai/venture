package engine

import (
	"fmt"
	"time"

	"github.com/opd-ai/venture/pkg/social"
)

// ChatSystem processes chat messages and enforces chat rules.
type ChatSystem struct {
	world *World
}

// NewChatSystem creates a new chat system.
func NewChatSystem(world *World) *ChatSystem {
	return &ChatSystem{
		world: world,
	}
}

// Update processes pending chat messages (called each frame).
func (cs *ChatSystem) Update(deltaTime float64) {
	// Chat processing happens on-demand via SendMessage/ReceiveMessage
	// This update can be used for cleanup or periodic tasks
}

// SendMessage sends a chat message from an entity to a channel.
// Returns error if entity cannot send (muted, rate limited, invalid channel).
func (cs *ChatSystem) SendMessage(senderID uint64, channel ChatChannel, content string, recipientID uint64) error {
	sender, chat, err := cs.getSenderAndChat(senderID)
	if err != nil {
		return err
	}

	if err := cs.validateMessagePermissions(chat, channel, recipientID); err != nil {
		return err
	}

	chat.RecordMessageSent(channel)
	msg := cs.createMessage(senderID, sender, channel, content, recipientID)
	chat.AddMessage(msg)
	cs.deliverMessage(msg, sender, chat)

	return nil
}

// getSenderAndChat retrieves the sender entity and their chat component.
func (cs *ChatSystem) getSenderAndChat(senderID uint64) (*Entity, *ChatComponent, error) {
	sender, exists := cs.world.GetEntity(senderID)
	if !exists {
		return nil, nil, fmt.Errorf("sender entity not found")
	}

	chatComp, exists := sender.GetComponent("chat")
	if !exists {
		chatComp = NewChatComponent()
		sender.AddComponent(chatComp)
	}

	chat, ok := chatComp.(*ChatComponent)
	if !ok {
		return nil, nil, fmt.Errorf("invalid chat component type")
	}

	return sender, chat, nil
}

// validateMessagePermissions checks if the entity can send a message to the channel.
func (cs *ChatSystem) validateMessagePermissions(chat *ChatComponent, channel ChatChannel, recipientID uint64) error {
	if !chat.CanSendMessage(channel) {
		if chat.IsMuted() {
			return social.ErrMuted(chat.MuteExpiry.Format(time.RFC1123))
		}
		return social.ErrRateLimit(channel.String())
	}

	if !chat.IsChannelActive(channel) {
		return social.ErrNotSubscribed(channel.String())
	}

	if channel == ChatWhisper {
		return cs.validateWhisperRecipient(recipientID)
	}

	return nil
}

// validateWhisperRecipient ensures a valid recipient exists for whisper messages.
func (cs *ChatSystem) validateWhisperRecipient(recipientID uint64) error {
	if recipientID == 0 {
		return fmt.Errorf("whisper requires recipient ID")
	}
	_, exists := cs.world.GetEntity(recipientID)
	if !exists {
		return fmt.Errorf("whisper recipient not found")
	}
	return nil
}

// createMessage constructs a new chat message.
func (cs *ChatSystem) createMessage(senderID uint64, sender *Entity, channel ChatChannel, content string, recipientID uint64) ChatMessage {
	return ChatMessage{
		ID:          cs.generateMessageID(),
		SenderID:    senderID,
		SenderName:  cs.getSenderName(sender),
		RecipientID: recipientID,
		Channel:     channel,
		Content:     content,
		Timestamp:   time.Now(),
		Delivered:   false,
		Failed:      false,
	}
}

// deliverMessage delivers a message to appropriate recipients based on channel.
func (cs *ChatSystem) deliverMessage(msg ChatMessage, sender *Entity, senderChat *ChatComponent) {
	switch msg.Channel {
	case ChatGlobal:
		cs.deliverToAll(msg, sender.ID)
	case ChatLocal:
		cs.deliverToLocal(msg, sender, senderChat)
	case ChatParty:
		cs.deliverToParty(msg, sender)
	case ChatWhisper:
		cs.deliverToRecipient(msg, msg.RecipientID)
	}
}

// deliverToAll delivers a message to all entities with chat components.
func (cs *ChatSystem) deliverToAll(msg ChatMessage, senderID uint64) {
	for _, entity := range cs.world.GetEntities() {
		if entity.ID == senderID {
			continue // Don't deliver to sender (already in history)
		}

		chatComp, exists := entity.GetComponent("chat")
		if !exists {
			continue
		}

		chat, ok := chatComp.(*ChatComponent)
		if !ok || !chat.IsChannelActive(ChatGlobal) {
			continue
		}

		// Mark as delivered and add to recipient's history
		deliveredMsg := msg
		deliveredMsg.Delivered = true
		chat.AddMessage(deliveredMsg)
	}
}

// deliverToLocal delivers a message to entities within range.
func (cs *ChatSystem) deliverToLocal(msg ChatMessage, sender *Entity, senderChat *ChatComponent) {
	senderPosComp, ok := cs.getSenderPosition(sender)
	if !ok {
		return
	}

	radius := senderChat.GetEffectiveRadius()
	radiusSquared := radius * radius

	for _, entity := range cs.world.GetEntities() {
		if entity.ID == sender.ID {
			continue
		}

		chat := cs.getLocalChatRecipient(entity)
		if chat == nil {
			continue
		}

		recipientPosComp := cs.getRecipientPosition(entity)
		if recipientPosComp == nil {
			continue
		}

		if cs.isWithinChatRange(senderPosComp, recipientPosComp, radius, radiusSquared) {
			cs.deliverMessageToChat(msg, chat)
		}
	}
}

// getSenderPosition retrieves and validates the sender's position component.
func (cs *ChatSystem) getSenderPosition(sender *Entity) (*PositionComponent, bool) {
	senderPos, exists := sender.GetComponent("position")
	if !exists {
		return nil, false
	}

	senderPosComp, ok := senderPos.(*PositionComponent)
	if !ok {
		return nil, false
	}

	return senderPosComp, true
}

// getLocalChatRecipient retrieves a valid chat component for local delivery.
func (cs *ChatSystem) getLocalChatRecipient(entity *Entity) *ChatComponent {
	chatComp, exists := entity.GetComponent("chat")
	if !exists {
		return nil
	}

	chat, ok := chatComp.(*ChatComponent)
	if !ok || !chat.IsChannelActive(ChatLocal) {
		return nil
	}

	return chat
}

// getRecipientPosition retrieves and validates the recipient's position component.
func (cs *ChatSystem) getRecipientPosition(entity *Entity) *PositionComponent {
	recipientPos, exists := entity.GetComponent("position")
	if !exists {
		return nil
	}

	recipientPosComp, ok := recipientPos.(*PositionComponent)
	if !ok {
		return nil
	}

	return recipientPosComp
}

// isWithinChatRange checks if recipient is within chat range of sender.
func (cs *ChatSystem) isWithinChatRange(senderPos, recipientPos *PositionComponent, radius, radiusSquared float64) bool {
	// Unlimited range for walkie-talkie
	if radius < 0 {
		return true
	}

	dx := senderPos.X - recipientPos.X
	dy := senderPos.Y - recipientPos.Y
	distSquared := dx*dx + dy*dy

	return distSquared <= radiusSquared
}

// deliverMessageToChat delivers a message to a chat component.
func (cs *ChatSystem) deliverMessageToChat(msg ChatMessage, chat *ChatComponent) {
	deliveredMsg := msg
	deliveredMsg.Delivered = true
	chat.AddMessage(deliveredMsg)
}

// deliverToParty delivers a message to party members.
func (cs *ChatSystem) deliverToParty(msg ChatMessage, sender *Entity) {
	// Get sender's party component
	partyComp, ok := sender.GetComponent("party")
	if !ok || partyComp == nil {
		// Sender is not in a party, message cannot be delivered
		return
	}

	party, ok := partyComp.(*PartyComponent)
	if !ok {
		return
	}

	// Deliver to all party members
	for _, memberID := range party.MemberIDs {
		if memberID == sender.ID {
			continue // Don't deliver to sender (already in history)
		}

		member, exists := cs.world.GetEntity(memberID)
		if !exists {
			continue
		}

		chatComp, exists := member.GetComponent("chat")
		if !exists {
			continue
		}

		chat, ok := chatComp.(*ChatComponent)
		if !ok || !chat.IsChannelActive(ChatParty) {
			continue
		}

		// Mark as delivered and add to recipient's history
		deliveredMsg := msg
		deliveredMsg.Delivered = true
		chat.AddMessage(deliveredMsg)
	}
}

// deliverToRecipient delivers a whisper to a specific recipient.
func (cs *ChatSystem) deliverToRecipient(msg ChatMessage, recipientID uint64) {
	recipient, exists := cs.world.GetEntity(recipientID)
	if !exists {
		return
	}

	chatComp, exists := recipient.GetComponent("chat")
	if !exists {
		// Create chat component for recipient
		chatComp = NewChatComponent()
		recipient.AddComponent(chatComp)
	}

	chat, ok := chatComp.(*ChatComponent)
	if !ok {
		return
	}

	// Deliver whisper
	deliveredMsg := msg
	deliveredMsg.Delivered = true
	chat.AddMessage(deliveredMsg)
}

// generateMessageID generates a unique message ID.
func (cs *ChatSystem) generateMessageID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// getSenderName retrieves the sender's display name from the entity.
func (cs *ChatSystem) getSenderName(entity *Entity) string {
	// Fallback to entity ID as name
	return fmt.Sprintf("Entity_%d", entity.ID)
}

// ApplyMute mutes an entity for the specified duration.
func (cs *ChatSystem) ApplyMute(entityID uint64, duration time.Duration) error {
	entity, exists := cs.world.GetEntity(entityID)
	if !exists {
		return fmt.Errorf("entity not found")
	}

	chatComp, exists := entity.GetComponent("chat")
	if !exists {
		chatComp = NewChatComponent()
		entity.AddComponent(chatComp)
	}

	chat, ok := chatComp.(*ChatComponent)
	if !ok {
		return fmt.Errorf("invalid chat component type")
	}

	chat.ApplyMute(duration)
	return nil
}

// IsMuted checks if an entity is currently muted.
func (cs *ChatSystem) IsMuted(entityID uint64) bool {
	entity, exists := cs.world.GetEntity(entityID)
	if !exists {
		return false
	}

	chatComp, exists := entity.GetComponent("chat")
	if !exists {
		return false
	}

	chat, ok := chatComp.(*ChatComponent)
	if !ok {
		return false
	}

	return chat.IsMuted()
}

// GetMessageHistory retrieves the message history for an entity.
func (cs *ChatSystem) GetMessageHistory(entityID uint64) ([]ChatMessage, error) {
	entity, exists := cs.world.GetEntity(entityID)
	if !exists {
		return nil, fmt.Errorf("entity not found")
	}

	chatComp, exists := entity.GetComponent("chat")
	if !exists {
		return []ChatMessage{}, nil
	}

	chat, ok := chatComp.(*ChatComponent)
	if !ok {
		return nil, fmt.Errorf("invalid chat component type")
	}

	return chat.Messages, nil
}

// MarkMessagesRead marks all messages as read for an entity.
func (cs *ChatSystem) MarkMessagesRead(entityID uint64) error {
	entity, exists := cs.world.GetEntity(entityID)
	if !exists {
		return fmt.Errorf("entity not found")
	}

	chatComp, exists := entity.GetComponent("chat")
	if !exists {
		return nil // No messages to mark
	}

	chat, ok := chatComp.(*ChatComponent)
	if !ok {
		return fmt.Errorf("invalid chat component type")
	}

	chat.MarkAllRead()
	return nil
}

// SubscribeChannel subscribes an entity to a chat channel.
func (cs *ChatSystem) SubscribeChannel(entityID uint64, channel ChatChannel) error {
	entity, exists := cs.world.GetEntity(entityID)
	if !exists {
		return fmt.Errorf("entity not found")
	}

	chatComp, exists := entity.GetComponent("chat")
	if !exists {
		chatComp = NewChatComponent()
		entity.AddComponent(chatComp)
	}

	chat, ok := chatComp.(*ChatComponent)
	if !ok {
		return fmt.Errorf("invalid chat component type")
	}

	chat.SubscribeChannel(channel)
	return nil
}

// UnsubscribeChannel unsubscribes an entity from a chat channel.
func (cs *ChatSystem) UnsubscribeChannel(entityID uint64, channel ChatChannel) error {
	entity, exists := cs.world.GetEntity(entityID)
	if !exists {
		return fmt.Errorf("entity not found")
	}

	chatComp, exists := entity.GetComponent("chat")
	if !exists {
		return nil // Not subscribed to anything
	}

	chat, ok := chatComp.(*ChatComponent)
	if !ok {
		return fmt.Errorf("invalid chat component type")
	}

	chat.UnsubscribeChannel(channel)
	return nil
}
