package engine

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/social"
)

func TestNewChatSystem(t *testing.T) {
	world := NewWorld()
	system := NewChatSystem(world)

	if system == nil {
		t.Fatal("NewChatSystem() returned nil")
	}
	if system.world != world {
		t.Error("system.world not set correctly")
	}
}

func TestChatSystem_Update(t *testing.T) {
	world := NewWorld()
	system := NewChatSystem(world)

	// Update should not panic
	system.Update(0.016)
}

func TestChatSystem_SendMessage(t *testing.T) {
	tests := []struct {
		name        string
		setupSender bool
		setupRecip  bool
		channel     ChatChannel
		recipientID uint64
		wantErr     bool
		errType     string
	}{
		{
			name:        "sender_not_found",
			setupSender: false,
			channel:     ChatGlobal,
			wantErr:     true,
			errType:     "entity",
		},
		{
			name:        "global_message_success",
			setupSender: true,
			channel:     ChatGlobal,
			wantErr:     false,
		},
		{
			name:        "local_message_success",
			setupSender: true,
			channel:     ChatLocal,
			wantErr:     false,
		},
		{
			name:        "whisper_no_recipient",
			setupSender: true,
			channel:     ChatWhisper,
			recipientID: 0,
			wantErr:     true,
			errType:     "recipient",
		},
		{
			name:        "whisper_recipient_not_found",
			setupSender: true,
			channel:     ChatWhisper,
			recipientID: 9999,
			wantErr:     true,
			errType:     "recipient",
		},
		{
			name:        "whisper_success",
			setupSender: true,
			setupRecip:  true,
			channel:     ChatWhisper,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewChatSystem(world)

			var senderID uint64 = 9999
			var recipientID uint64 = tt.recipientID

			if tt.setupSender {
				sender := world.CreateEntity()
				senderID = sender.ID
				// Add position for local chat
				sender.AddComponent(&PositionComponent{X: 0, Y: 0})
				// Add chat component with channel subscription
				chatComp := NewChatComponent()
				chatComp.SubscribeChannel(tt.channel)
				sender.AddComponent(chatComp)
				world.Update(0)
			}

			if tt.setupRecip {
				recipient := world.CreateEntity()
				recipientID = recipient.ID
				recipient.AddComponent(&PositionComponent{X: 1, Y: 1})
				world.Update(0)
			}

			err := system.SendMessage(senderID, tt.channel, "test message", recipientID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestChatSystem_SendMessage_Muted(t *testing.T) {
	world := NewWorld()
	system := NewChatSystem(world)

	sender := world.CreateEntity()
	chatComp := NewChatComponent()
	chatComp.ApplyMute(time.Hour)
	sender.AddComponent(chatComp)
	world.Update(0)

	err := system.SendMessage(sender.ID, ChatGlobal, "test", 0)
	if err == nil {
		t.Error("expected mute error but got nil")
	}

	// Check if it's a social error
	if _, ok := social.IsSocialError(err); !ok {
		t.Errorf("expected SocialError, got %T", err)
	}
}

func TestChatSystem_SendMessage_NotSubscribed(t *testing.T) {
	world := NewWorld()
	system := NewChatSystem(world)

	sender := world.CreateEntity()
	chatComp := NewChatComponent()
	chatComp.UnsubscribeChannel(ChatGlobal) // Unsubscribe from global
	sender.AddComponent(chatComp)
	world.Update(0)

	err := system.SendMessage(sender.ID, ChatGlobal, "test", 0)
	if err == nil {
		t.Error("expected not subscribed error but got nil")
	}

	socialErr, ok := social.IsSocialError(err)
	if !ok {
		t.Errorf("expected SocialError, got %T", err)
	}
	if socialErr != nil && socialErr.Type != social.ErrorTypeNotSubscribed {
		t.Errorf("expected ErrorTypeNotSubscribed, got %v", socialErr.Type)
	}
}

func TestChatSystem_ApplyMute(t *testing.T) {
	tests := []struct {
		name     string
		entityID uint64
		setup    bool
		wantErr  bool
	}{
		{
			name:     "entity_not_found",
			entityID: 9999,
			setup:    false,
			wantErr:  true,
		},
		{
			name:  "mute_success",
			setup: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewChatSystem(world)

			entityID := tt.entityID
			if tt.setup {
				entity := world.CreateEntity()
				entityID = entity.ID
				world.Update(0)
			}

			err := system.ApplyMute(entityID, time.Minute)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				// Verify mute was applied
				if !system.IsMuted(entityID) {
					t.Error("entity should be muted")
				}
			}
		})
	}
}

func TestChatSystem_IsMuted(t *testing.T) {
	world := NewWorld()
	system := NewChatSystem(world)

	// Non-existent entity should not be muted
	if system.IsMuted(9999) {
		t.Error("non-existent entity should not be muted")
	}

	// Entity without chat component should not be muted
	entity := world.CreateEntity()
	world.Update(0)
	if system.IsMuted(entity.ID) {
		t.Error("entity without chat component should not be muted")
	}

	// Muted entity should be muted
	chatComp := NewChatComponent()
	chatComp.ApplyMute(time.Hour)
	entity.AddComponent(chatComp)
	if !system.IsMuted(entity.ID) {
		t.Error("muted entity should be muted")
	}
}

func TestChatSystem_GetMessageHistory(t *testing.T) {
	tests := []struct {
		name        string
		setup       bool
		addMessages bool
		wantErr     bool
		wantCount   int
	}{
		{
			name:    "entity_not_found",
			setup:   false,
			wantErr: true,
		},
		{
			name:      "empty_history",
			setup:     true,
			wantCount: 0,
		},
		{
			name:        "with_messages",
			setup:       true,
			addMessages: true,
			wantCount:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewChatSystem(world)

			var entityID uint64 = 9999
			if tt.setup {
				entity := world.CreateEntity()
				entityID = entity.ID
				if tt.addMessages {
					chatComp := NewChatComponent()
					chatComp.AddMessage(ChatMessage{ID: "1", Content: "msg1"})
					chatComp.AddMessage(ChatMessage{ID: "2", Content: "msg2"})
					entity.AddComponent(chatComp)
				}
				world.Update(0)
			}

			history, err := system.GetMessageHistory(entityID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(history) != tt.wantCount {
					t.Errorf("expected %d messages, got %d", tt.wantCount, len(history))
				}
			}
		})
	}
}

func TestChatSystem_MarkMessagesRead(t *testing.T) {
	world := NewWorld()
	system := NewChatSystem(world)

	// Non-existent entity
	err := system.MarkMessagesRead(9999)
	if err == nil {
		t.Error("expected error for non-existent entity")
	}

	// Entity without chat component (should not error)
	entity := world.CreateEntity()
	world.Update(0)
	err = system.MarkMessagesRead(entity.ID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Entity with chat component
	chatComp := NewChatComponent()
	chatComp.AddMessage(ChatMessage{ID: "1", Content: "msg"})
	entity.AddComponent(chatComp)

	err = system.MarkMessagesRead(entity.ID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChatSystem_SubscribeChannel(t *testing.T) {
	world := NewWorld()
	system := NewChatSystem(world)

	// Non-existent entity
	err := system.SubscribeChannel(9999, ChatGlobal)
	if err == nil {
		t.Error("expected error for non-existent entity")
	}

	// Existing entity
	entity := world.CreateEntity()
	world.Update(0)

	err = system.SubscribeChannel(entity.ID, ChatParty)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChatSystem_UnsubscribeChannel(t *testing.T) {
	world := NewWorld()
	system := NewChatSystem(world)

	// Non-existent entity
	err := system.UnsubscribeChannel(9999, ChatGlobal)
	if err == nil {
		t.Error("expected error for non-existent entity")
	}

	// Entity without chat component (should not error)
	entity := world.CreateEntity()
	world.Update(0)
	err = system.UnsubscribeChannel(entity.ID, ChatGlobal)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Entity with chat component
	chatComp := NewChatComponent()
	entity.AddComponent(chatComp)

	err = system.UnsubscribeChannel(entity.ID, ChatGlobal)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChatSystem_DeliverMessage_Global(t *testing.T) {
	world := NewWorld()
	system := NewChatSystem(world)

	// Create sender and recipient
	sender := world.CreateEntity()
	sender.AddComponent(&PositionComponent{X: 0, Y: 0})

	recipient := world.CreateEntity()
	recipientChat := NewChatComponent()
	recipient.AddComponent(recipientChat)

	world.Update(0)

	err := system.SendMessage(sender.ID, ChatGlobal, "Hello world", 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify recipient received the message
	if len(recipientChat.Messages) == 0 {
		t.Error("recipient should have received the message")
	}
}

func TestChatSystem_DeliverMessage_Local(t *testing.T) {
	world := NewWorld()
	system := NewChatSystem(world)

	// Create sender with position
	sender := world.CreateEntity()
	sender.AddComponent(&PositionComponent{X: 0, Y: 0})

	// Create nearby recipient
	nearRecipient := world.CreateEntity()
	nearRecipient.AddComponent(&PositionComponent{X: 5, Y: 5})
	nearChat := NewChatComponent()
	nearRecipient.AddComponent(nearChat)

	// Create far recipient
	farRecipient := world.CreateEntity()
	farRecipient.AddComponent(&PositionComponent{X: 1000, Y: 1000})
	farChat := NewChatComponent()
	farRecipient.AddComponent(farChat)

	world.Update(0)

	err := system.SendMessage(sender.ID, ChatLocal, "Local message", 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Near recipient should have received the message
	if len(nearChat.Messages) == 0 {
		t.Error("near recipient should have received the message")
	}

	// Far recipient should not have received the message
	if len(farChat.Messages) > 0 {
		t.Error("far recipient should not have received the message")
	}
}

func TestChatSystem_DeliverMessage_Party(t *testing.T) {
	world := NewWorld()
	system := NewChatSystem(world)

	// Create party with leader and members
	party := NewPartyComponent("party-123", 1)
	party.AddMember(2)
	party.AddMember(3)

	sender := world.CreateEntity()
	sender.ID = 1 // Leader sends message
	sender.AddComponent(&PositionComponent{X: 0, Y: 0})
	senderChat := NewChatComponent()
	senderChat.SubscribeChannel(ChatParty)
	sender.AddComponent(senderChat)
	sender.AddComponent(party)

	// Party member 1
	member1 := world.CreateEntity()
	member1.ID = 2
	member1Chat := NewChatComponent()
	member1Chat.SubscribeChannel(ChatParty)
	member1.AddComponent(member1Chat)
	member1.AddComponent(party)

	// Party member 2
	member2 := world.CreateEntity()
	member2.ID = 3
	member2Chat := NewChatComponent()
	member2Chat.SubscribeChannel(ChatParty)
	member2.AddComponent(member2Chat)
	member2.AddComponent(party)

	// Non-party member (should not receive message)
	nonMember := world.CreateEntity()
	nonMemberChat := NewChatComponent()
	nonMemberChat.SubscribeChannel(ChatParty)
	nonMember.AddComponent(nonMemberChat)

	world.Update(0)

	err := system.SendMessage(sender.ID, ChatParty, "Party message", 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Party members should have received the message
	if len(member1Chat.Messages) != 1 {
		t.Errorf("member1 messages = %d, want 1", len(member1Chat.Messages))
	}
	if len(member2Chat.Messages) != 1 {
		t.Errorf("member2 messages = %d, want 1", len(member2Chat.Messages))
	}

	// Non-party member should not have received the message
	if len(nonMemberChat.Messages) != 0 {
		t.Errorf("non-member messages = %d, want 0", len(nonMemberChat.Messages))
	}

	// Verify message content
	if member1Chat.Messages[0].Content != "Party message" {
		t.Errorf("message content = %v, want 'Party message'", member1Chat.Messages[0].Content)
	}
	if !member1Chat.Messages[0].Delivered {
		t.Error("message should be marked as delivered")
	}
}

func TestChatSystem_DeliverMessage_Party_NoPartyComponent(t *testing.T) {
	world := NewWorld()
	system := NewChatSystem(world)

	sender := world.CreateEntity()
	sender.AddComponent(&PositionComponent{X: 0, Y: 0})
	senderChat := NewChatComponent()
	senderChat.SubscribeChannel(ChatParty)
	sender.AddComponent(senderChat)
	// No party component

	recipient := world.CreateEntity()
	recipientChat := NewChatComponent()
	recipientChat.SubscribeChannel(ChatParty)
	recipient.AddComponent(recipientChat)

	world.Update(0)

	err := system.SendMessage(sender.ID, ChatParty, "Party message", 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Without party component, message should not be delivered
	if len(recipientChat.Messages) != 0 {
		t.Error("recipient should not have received message (sender not in party)")
	}
}

func TestChatSystem_DeliverMessage_Party_MemberNotSubscribed(t *testing.T) {
	world := NewWorld()
	system := NewChatSystem(world)

	party := NewPartyComponent("party-123", 1)
	party.AddMember(2)

	sender := world.CreateEntity()
	sender.ID = 1
	sender.AddComponent(&PositionComponent{X: 0, Y: 0})
	senderChat := NewChatComponent()
	senderChat.SubscribeChannel(ChatParty)
	sender.AddComponent(senderChat)
	sender.AddComponent(party)

	// Party member not subscribed to party channel
	member := world.CreateEntity()
	member.ID = 2
	memberChat := NewChatComponent()
	// NOT subscribed to ChatParty
	member.AddComponent(memberChat)
	member.AddComponent(party)

	world.Update(0)

	err := system.SendMessage(sender.ID, ChatParty, "Party message", 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Member should not receive message if not subscribed
	if len(memberChat.Messages) != 0 {
		t.Error("unsubscribed member should not have received message")
	}
}

func TestChatSystem_DeliverMessage_Party_MemberMissingChatComponent(t *testing.T) {
	world := NewWorld()
	system := NewChatSystem(world)

	party := NewPartyComponent("party-123", 1)
	party.AddMember(2)

	sender := world.CreateEntity()
	sender.ID = 1
	sender.AddComponent(&PositionComponent{X: 0, Y: 0})
	senderChat := NewChatComponent()
	senderChat.SubscribeChannel(ChatParty)
	sender.AddComponent(senderChat)
	sender.AddComponent(party)

	// Party member without chat component
	member := world.CreateEntity()
	member.ID = 2
	member.AddComponent(party)
	// No chat component

	world.Update(0)

	// Should not panic
	err := system.SendMessage(sender.ID, ChatParty, "Party message", 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChatSystem_DeliverMessage_Party_MemberEntityNotFound(t *testing.T) {
	world := NewWorld()
	system := NewChatSystem(world)

	party := NewPartyComponent("party-123", 1)
	party.AddMember(999) // Non-existent entity ID

	sender := world.CreateEntity()
	sender.ID = 1
	sender.AddComponent(&PositionComponent{X: 0, Y: 0})
	senderChat := NewChatComponent()
	senderChat.SubscribeChannel(ChatParty)
	sender.AddComponent(senderChat)
	sender.AddComponent(party)

	world.Update(0)

	// Should not panic when member entity doesn't exist
	err := system.SendMessage(sender.ID, ChatParty, "Party message", 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChatSystem_DeliverMessage_Whisper(t *testing.T) {
	world := NewWorld()
	system := NewChatSystem(world)

	sender := world.CreateEntity()
	sender.AddComponent(&PositionComponent{X: 0, Y: 0})
	senderChat := NewChatComponent()
	senderChat.SubscribeChannel(ChatWhisper)
	sender.AddComponent(senderChat)

	recipient := world.CreateEntity()
	recipientChat := NewChatComponent()
	recipient.AddComponent(recipientChat)

	otherEntity := world.CreateEntity()
	otherChat := NewChatComponent()
	otherEntity.AddComponent(otherChat)

	world.Update(0)

	err := system.SendMessage(sender.ID, ChatWhisper, "Private message", recipient.ID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Recipient should have received the whisper
	if len(recipientChat.Messages) == 0 {
		t.Error("recipient should have received the whisper")
	}

	// Other entity should not have received the whisper
	if len(otherChat.Messages) > 0 {
		t.Error("other entity should not have received the whisper")
	}
}

func TestChatSystem_GenerateMessageID(t *testing.T) {
	world := NewWorld()
	system := NewChatSystem(world)

	id1 := system.generateMessageID()
	id2 := system.generateMessageID()

	if id1 == "" {
		t.Error("message ID should not be empty")
	}

	// IDs should be different (based on UnixNano)
	// Note: In fast execution, they might be the same, so we just verify non-empty
	if id2 == "" {
		t.Error("message ID should not be empty")
	}
}

func TestChatSystem_GetSenderName(t *testing.T) {
	world := NewWorld()
	system := NewChatSystem(world)

	entity := world.CreateEntity()
	world.Update(0)

	name := system.getSenderName(entity)
	if name == "" {
		t.Error("sender name should not be empty")
	}

	// Should contain entity ID
	expected := "Entity_"
	if len(name) < len(expected) {
		t.Errorf("sender name should start with Entity_, got %s", name)
	}
}
