package engine

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/social/persistence"
)

func TestNewEnhancedChatSystem(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"creates_system_with_world", true},
		{"initializes_empty_history_map", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewEnhancedChatSystem(world)

			if system == nil {
				t.Errorf("NewEnhancedChatSystem() returned nil")
			}
			if system.world != world {
				t.Errorf("system.world not set correctly")
			}
			if system.history == nil {
				t.Errorf("history map is nil")
			}
		})
	}
}

func TestEnhancedChatSystem_RegisterPlayer(t *testing.T) {
	tests := []struct {
		name       string
		playerID   uint64
		createEnt  bool
		wantErr    bool
		errMessage string
	}{
		{
			name:      "valid_registration",
			playerID:  1,
			createEnt: true,
			wantErr:   false,
		},
		{
			name:       "missing_entity",
			playerID:   999,
			createEnt:  false,
			wantErr:    true,
			errMessage: "player entity not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewEnhancedChatSystem(world)

			var entity *Entity
			var playerID uint64

			if tt.createEnt {
				entity = world.CreateEntity()
				playerID = entity.ID
				world.Update(0) // Process pending entity additions
			} else {
				playerID = 999 // Non-existent ID
			}

			world.Update(0) // Process pending entity additions
			err := system.RegisterPlayer(playerID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("RegisterPlayer() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("RegisterPlayer() unexpected error: %v", err)
				return
			}

			// Verify history initialized
			if _, exists := system.history[playerID]; !exists {
				t.Errorf("history not initialized for player %d", playerID)
			}

			// Verify chat component added
			if _, exists := entity.GetComponent("chat"); !exists {
				t.Errorf("chat component not added to entity")
			}
		})
	}
}

func TestEnhancedChatSystem_UnregisterPlayer(t *testing.T) {
	world := NewWorld()
	system := NewEnhancedChatSystem(world)

	// Register a player
	entity := world.CreateEntity()
	playerID := entity.ID
	world.Update(0) // Process pending entity additions
	system.RegisterPlayer(playerID)
	system.RegisterPlayer(playerID)

	// Verify registration
	if _, exists := system.history[playerID]; !exists {
		t.Fatalf("player not registered")
	}

	// Unregister
	system.UnregisterPlayer(playerID)

	// Verify cleanup
	if _, exists := system.history[playerID]; exists {
		t.Errorf("player history not removed")
	}
}

func TestEnhancedChatSystem_SendMessage_GlobalChannel(t *testing.T) {
	world := NewWorld()
	system := NewEnhancedChatSystem(world)

	// Create sender
	sender := world.CreateEntity()
	senderID := sender.ID
	sender.AddComponent(&PositionComponent{X: 100, Y: 100})
	world.Update(0) // Process pending entity additions
	system.RegisterPlayer(senderID)
	system.RegisterPlayer(senderID)

	// Create recipient
	recipient := world.CreateEntity()
	recipientID := recipient.ID
	recipient.AddComponent(&PositionComponent{X: 200, Y: 200})
	world.Update(0) // Process pending entity additions
	system.RegisterPlayer(recipientID)
	system.RegisterPlayer(recipientID)

	// Send message
	err := system.SendMessage(senderID, ChatGlobal, "Hello world!", 0)
	if err != nil {
		t.Fatalf("SendMessage() failed: %v", err)
	}

	// Verify message in sender's history
	senderChatRaw, _ := sender.GetComponent("chat")
	senderChat := senderChatRaw.(*ChatComponent)
	if len(senderChat.Messages) != 1 {
		t.Errorf("sender should have 1 message, got %d", len(senderChat.Messages))
	}

	// Verify message delivered to recipient
	recipientChatRaw, _ := recipient.GetComponent("chat")
	recipientChat := recipientChatRaw.(*ChatComponent)
	if len(recipientChat.Messages) != 1 {
		t.Errorf("recipient should have 1 message, got %d", len(recipientChat.Messages))
	}
	if recipientChat.Messages[0].Content != "Hello world!" {
		t.Errorf("wrong message content: got %q", recipientChat.Messages[0].Content)
	}
	if !recipientChat.Messages[0].Delivered {
		t.Errorf("message not marked as delivered")
	}

	// Verify persistent history
	hist, err := system.GetPlayerHistory(senderID, nil)
	if err != nil {
		t.Fatalf("GetPlayerHistory() failed: %v", err)
	}
	if len(hist) != 1 {
		t.Errorf("sender persistent history should have 1 message, got %d", len(hist))
	}
}

func TestEnhancedChatSystem_SendMessage_LocalChannel(t *testing.T) {
	world := NewWorld()
	system := NewEnhancedChatSystem(world)

	// Create sender
	sender := world.CreateEntity()
	senderID := sender.ID
	sender.AddComponent(&PositionComponent{X: 100, Y: 100})
	world.Update(0) // Process pending entity additions
	system.RegisterPlayer(senderID)

	// Create nearby recipient (within 10 tile radius)
	nearRecipient := world.CreateEntity()
	nearID := nearRecipient.ID
	nearRecipient.AddComponent(&PositionComponent{X: 105, Y: 105})
	world.Update(0) // Process pending entity additions
	system.RegisterPlayer(nearID)

	// Create far recipient (beyond 10 tile radius)
	farRecipient := world.CreateEntity()
	farID := farRecipient.ID
	farRecipient.AddComponent(&PositionComponent{X: 200, Y: 200})
	world.Update(0) // Process pending entity additions
	system.RegisterPlayer(farID)

	// Subscribe both to local channel
	nearChatRaw, _ := nearRecipient.GetComponent("chat")
	nearChat := nearChatRaw.(*ChatComponent)
	nearChat.SubscribeChannel(ChatLocal)
	farChatRaw, _ := farRecipient.GetComponent("chat")
	farChat := farChatRaw.(*ChatComponent)
	farChat.SubscribeChannel(ChatLocal)

	// Send local message
	err := system.SendMessage(senderID, ChatLocal, "Anyone nearby?", 0)
	if err != nil {
		t.Fatalf("SendMessage() failed: %v", err)
	}

	// Verify near recipient received message
	if len(nearChat.Messages) != 1 {
		t.Errorf("near recipient should have 1 message, got %d", len(nearChat.Messages))
	}

	// Verify far recipient did NOT receive message
	if len(farChat.Messages) != 0 {
		t.Errorf("far recipient should have 0 messages, got %d", len(farChat.Messages))
	}
}

func TestEnhancedChatSystem_SendMessage_Whisper(t *testing.T) {
	world := NewWorld()
	system := NewEnhancedChatSystem(world)

	// Create sender
	sender := world.CreateEntity()
	senderID := sender.ID
	sender.AddComponent(&PositionComponent{X: 100, Y: 100})
	world.Update(0) // Process pending entity additions
	system.RegisterPlayer(senderID)

	// Subscribe sender to whisper channel
	senderChatRaw, _ := sender.GetComponent("chat")
	senderChat := senderChatRaw.(*ChatComponent)
	senderChat.SubscribeChannel(ChatWhisper)

	// Create recipient
	recipient := world.CreateEntity()
	recipientID := recipient.ID
	recipient.AddComponent(&PositionComponent{X: 200, Y: 200})
	world.Update(0) // Process pending entity additions
	system.RegisterPlayer(recipientID)

	// Create third player (should not receive whisper)
	other := world.CreateEntity()
	otherID := other.ID
	other.AddComponent(&PositionComponent{X: 300, Y: 300})
	world.Update(0) // Process pending entity additions
	system.RegisterPlayer(otherID)

	// Send whisper
	err := system.SendMessage(senderID, ChatWhisper, "Secret message", recipientID)
	if err != nil {
		t.Fatalf("SendMessage() failed: %v", err)
	}

	// Verify only recipient received message
	recipientChatRaw, _ := recipient.GetComponent("chat")
	recipientChat := recipientChatRaw.(*ChatComponent)
	if len(recipientChat.Messages) != 1 {
		t.Errorf("recipient should have 1 message, got %d", len(recipientChat.Messages))
	}

	otherChatRaw, _ := other.GetComponent("chat")
	otherChat := otherChatRaw.(*ChatComponent)
	if len(otherChat.Messages) != 0 {
		t.Errorf("other player should have 0 messages, got %d", len(otherChat.Messages))
	}
}

func TestEnhancedChatSystem_SendMessage_RateLimit(t *testing.T) {
	world := NewWorld()
	system := NewEnhancedChatSystem(world)

	// Create sender
	sender := world.CreateEntity()
	senderID := sender.ID
	sender.AddComponent(&PositionComponent{X: 100, Y: 100})
	world.Update(0) // Process pending entity additions
	system.RegisterPlayer(senderID)

	// Send first message (should succeed)
	err := system.SendMessage(senderID, ChatGlobal, "First message", 0)
	if err != nil {
		t.Fatalf("first SendMessage() failed: %v", err)
	}

	// Send second message immediately (should fail due to rate limit)
	err = system.SendMessage(senderID, ChatGlobal, "Second message", 0)
	if err == nil {
		t.Errorf("second SendMessage() should fail due to rate limit")
	}

	// Wait for cooldown and retry
	time.Sleep(3100 * time.Millisecond)

	err = system.SendMessage(senderID, ChatGlobal, "Third message", 0)
	if err != nil {
		t.Errorf("third SendMessage() after cooldown failed: %v", err)
	}
}

func TestEnhancedChatSystem_SendMessage_Muted(t *testing.T) {
	world := NewWorld()
	system := NewEnhancedChatSystem(world)

	// Create sender
	sender := world.CreateEntity()
	senderID := sender.ID
	sender.AddComponent(&PositionComponent{X: 100, Y: 100})
	world.Update(0) // Process pending entity additions
	system.RegisterPlayer(senderID)

	// Mute sender
	system.ApplyMute(senderID, 5*time.Second)

	// Try to send message (should fail)
	err := system.SendMessage(senderID, ChatGlobal, "I am muted", 0)
	if err == nil {
		t.Errorf("SendMessage() should fail for muted player")
	}
}

func TestEnhancedChatSystem_GetPlayerHistory(t *testing.T) {
	world := NewWorld()
	system := NewEnhancedChatSystem(world)

	// Create sender
	sender := world.CreateEntity()
	senderID := sender.ID
	sender.AddComponent(&PositionComponent{X: 100, Y: 100})
	world.Update(0) // Process pending entity additions
	system.RegisterPlayer(senderID)

	// Send multiple messages
	messages := []string{"First", "Second", "Third"}
	for i, msg := range messages {
		if i > 0 {
			time.Sleep(3100 * time.Millisecond) // Wait for GlobalChannel cooldown (3s)
		}
		system.SendMessage(senderID, ChatGlobal, msg, 0)
	}

	// Get history
	hist, err := system.GetPlayerHistory(senderID, nil)
	if err != nil {
		t.Fatalf("GetPlayerHistory() failed: %v", err)
	}

	if len(hist) != len(messages) {
		t.Errorf("history should have %d messages, got %d", len(messages), len(hist))
	}

	// Verify message order and content
	for i, msg := range hist {
		if msg.Content != messages[i] {
			t.Errorf("message %d: expected %q, got %q", i, messages[i], msg.Content)
		}
	}
}

func TestEnhancedChatSystem_GetPlayerHistory_WithFilter(t *testing.T) {
	world := NewWorld()
	system := NewEnhancedChatSystem(world)

	// Create sender
	sender := world.CreateEntity()
	senderID := sender.ID
	sender.AddComponent(&PositionComponent{X: 100, Y: 100})
	world.Update(0) // Process pending entity additions
	system.RegisterPlayer(senderID)

	// Create recipient
	recipient := world.CreateEntity()
	recipientID := recipient.ID
	recipient.AddComponent(&PositionComponent{X: 200, Y: 200})
	world.Update(0) // Process pending entity additions
	system.RegisterPlayer(recipientID)

	// Subscribe recipient to local channel
	recipientChatRaw, _ := recipient.GetComponent("chat")
	recipientChat := recipientChatRaw.(*ChatComponent)
	recipientChat.SubscribeChannel(ChatLocal)

	// Send messages to different channels
	time.Sleep(100 * time.Millisecond)
	system.SendMessage(senderID, ChatGlobal, "Global message", 0)
	time.Sleep(100 * time.Millisecond)
	system.SendMessage(senderID, ChatLocal, "Local message", 0)

	// Get history filtered by channel
	filter := &persistence.MessageFilter{
		Channel: "Global",
	}
	hist, err := system.GetPlayerHistory(senderID, filter)
	if err != nil {
		t.Fatalf("GetPlayerHistory() failed: %v", err)
	}

	if len(hist) != 1 {
		t.Errorf("filtered history should have 1 message, got %d", len(hist))
	}
	if hist[0].Content != "Global message" {
		t.Errorf("wrong message content: got %q", hist[0].Content)
	}
}

func TestEnhancedChatSystem_SaveLoadHistory(t *testing.T) {
	world := NewWorld()
	system := NewEnhancedChatSystem(world)

	// Create sender
	sender := world.CreateEntity()
	senderID := sender.ID
	sender.AddComponent(&PositionComponent{X: 100, Y: 100})
	world.Update(0) // Process pending entity additions
	system.RegisterPlayer(senderID)

	// Send message
	time.Sleep(100 * time.Millisecond)
	system.SendMessage(senderID, ChatGlobal, "Test message", 0)

	// Save history
	data, err := system.SaveHistory(senderID)
	if err != nil {
		t.Fatalf("SaveHistory() failed: %v", err)
	}
	if len(data) == 0 {
		t.Errorf("SaveHistory() returned empty data")
	}

	// Create new system and load history
	newWorld := NewWorld()
	newSystem := NewEnhancedChatSystem(newWorld)
	newEntity := newWorld.CreateEntity()
	newEntity.ID = senderID
	newEntity.AddComponent(&PositionComponent{X: 100, Y: 100})
	world.Update(0) // Process pending entity additions
	newSystem.RegisterPlayer(senderID)

	err = newSystem.LoadHistory(senderID, data)
	if err != nil {
		t.Fatalf("LoadHistory() failed: %v", err)
	}

	// Verify loaded history
	hist, err := newSystem.GetPlayerHistory(senderID, nil)
	if err != nil {
		t.Fatalf("GetPlayerHistory() failed: %v", err)
	}

	if len(hist) != 1 {
		t.Errorf("loaded history should have 1 message, got %d", len(hist))
	}
	if hist[0].Content != "Test message" {
		t.Errorf("wrong message content: got %q", hist[0].Content)
	}
}

func TestEnhancedChatSystem_ProcessACK(t *testing.T) {
	world := NewWorld()
	system := NewEnhancedChatSystem(world)

	// Create sender
	sender := world.CreateEntity()
	senderID := sender.ID
	sender.AddComponent(&PositionComponent{X: 100, Y: 100})
	world.Update(0) // Process pending entity additions
	system.RegisterPlayer(senderID)

	// Send message
	time.Sleep(100 * time.Millisecond)
	err := system.SendMessage(senderID, ChatGlobal, "Test message", 0)
	if err != nil {
		t.Fatalf("SendMessage() failed: %v", err)
	}

	// Get message ID
	senderChatRaw, _ := sender.GetComponent("chat")
	senderChat := senderChatRaw.(*ChatComponent)
	msgID := senderChat.Messages[0].ID

	// Process ACK
	system.ProcessACK(msgID, senderID, true, "")

	// Verify message marked as delivered
	msg := senderChat.GetMessageByID(msgID)
	if msg == nil {
		t.Fatalf("message not found")
	}
	if !msg.Delivered {
		t.Errorf("message not marked as delivered after ACK")
	}
}

func TestEnhancedChatSystem_ProcessNACK(t *testing.T) {
	world := NewWorld()
	system := NewEnhancedChatSystem(world)

	// Create sender
	sender := world.CreateEntity()
	senderID := sender.ID
	sender.AddComponent(&PositionComponent{X: 100, Y: 100})
	world.Update(0) // Process pending entity additions
	system.RegisterPlayer(senderID)

	// Send message
	time.Sleep(100 * time.Millisecond)
	err := system.SendMessage(senderID, ChatGlobal, "Test message", 0)
	if err != nil {
		t.Fatalf("SendMessage() failed: %v", err)
	}

	// Get message ID
	senderChatRaw, _ := sender.GetComponent("chat")
	senderChat := senderChatRaw.(*ChatComponent)
	msgID := senderChat.Messages[0].ID

	// Process NACK
	system.ProcessACK(msgID, senderID, false, "delivery failed")

	// Verify message marked as failed
	msg := senderChat.GetMessageByID(msgID)
	if msg == nil {
		t.Fatalf("message not found")
	}
	if !msg.Failed {
		t.Errorf("message not marked as failed after NACK")
	}
}

func TestEnhancedChatSystem_IsMuted(t *testing.T) {
	world := NewWorld()
	system := NewEnhancedChatSystem(world)

	// Create player
	player := world.CreateEntity()
	playerID := player.ID
	world.Update(0) // Process pending entity additions
	system.RegisterPlayer(playerID)

	// Initially not muted
	if system.IsMuted(playerID) {
		t.Errorf("player should not be muted initially")
	}

	// Apply mute
	system.ApplyMute(playerID, 2*time.Second)

	// Should be muted
	if !system.IsMuted(playerID) {
		t.Errorf("player should be muted")
	}

	// Wait for mute to expire
	time.Sleep(2100 * time.Millisecond)

	// Should not be muted anymore
	if system.IsMuted(playerID) {
		t.Errorf("player should not be muted after expiry")
	}
}
