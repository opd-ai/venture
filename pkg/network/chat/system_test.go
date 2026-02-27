package chat

import (
	"strings"
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/errors"
)

// TestNewChatSystem tests creation of chat system
func TestNewChatSystem(t *testing.T) {
	tests := []struct {
		name  string
		world *engine.World
	}{
		{"valid world", engine.NewWorld()},
		{"nil world", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := NewChatSystem(tt.world)
			if cs == nil {
				t.Fatal("NewChatSystem returned nil")
			}
			if cs.world != tt.world {
				t.Error("world not set correctly")
			}
		})
	}
}

// TestChatSystemUpdate tests chat system update
func TestChatSystemUpdate(t *testing.T) {
	world := engine.NewWorld()
	cs := NewChatSystem(world)

	tests := []struct {
		name      string
		deltaTime float64
	}{
		{"zero delta", 0.0},
		{"normal delta", 0.016},
		{"large delta", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			cs.Update(tt.deltaTime)
		})
	}
}

// TestSendMessage tests sending chat messages
func TestSendMessage(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*engine.World) uint64
		channel engine.ChatChannel
		content string
		wantErr bool
		errMsg  string
	}{
		{
			name: "send to global channel",
			setup: func(w *engine.World) uint64 {
				sender := w.CreateEntity()
				return sender.ID
			},
			channel: engine.ChatGlobal,
			content: "Hello world!",
			wantErr: false,
		},
		{
			name: "send to local channel",
			setup: func(w *engine.World) uint64 {
				sender := w.CreateEntity()
				return sender.ID
			},
			channel: engine.ChatLocal,
			content: "Hello nearby!",
			wantErr: false,
		},
		{
			name: "send to party channel",
			setup: func(w *engine.World) uint64 {
				sender := w.CreateEntity()
				return sender.ID
			},
			channel: engine.ChatParty,
			content: "Hello party!",
			wantErr: false,
		},
		{
			name: "send whisper",
			setup: func(w *engine.World) uint64 {
				sender := w.CreateEntity()
				return sender.ID
			},
			channel: engine.ChatWhisper,
			content: "Secret message",
			wantErr: false,
		},
		{
			name: "empty message",
			setup: func(w *engine.World) uint64 {
				sender := w.CreateEntity()
				return sender.ID
			},
			channel: engine.ChatGlobal,
			content: "",
			wantErr: true, // Empty messages are now rejected by validation
			errMsg:  "message validation failed",
		},
		{
			name: "long message",
			setup: func(w *engine.World) uint64 {
				sender := w.CreateEntity()
				return sender.ID
			},
			channel: engine.ChatGlobal,
			content: "This is a very long message that contains a lot of text to test if the chat system can handle long messages without any issues",
			wantErr: false,
		},
		{
			name: "sender not found",
			setup: func(w *engine.World) uint64 {
				return 9999 // non-existent entity
			},
			channel: engine.ChatGlobal,
			content: "Hello",
			wantErr: true,
			errMsg:  "sender entity not found",
		},
		{
			name: "multiple messages from same sender",
			setup: func(w *engine.World) uint64 {
				sender := w.CreateEntity()
				return sender.ID
			},
			channel: engine.ChatGlobal,
			content: "First message",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := engine.NewWorld()
			cs := NewChatSystem(world)

			senderID := tt.setup(world)
			// Process pending entity additions
			world.Update(0)

			err := cs.SendMessage(senderID, tt.channel, tt.content)

			if (err != nil) != tt.wantErr {
				t.Errorf("SendMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error message = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}

			// If successful, verify message was added
			if !tt.wantErr {
				sender, ok := world.GetEntity(senderID)
				if !ok {
					t.Fatal("sender not found after sending message")
				}

				chatCompRaw, ok := sender.GetComponent("chat")
				if !ok {
					t.Fatal("chat component not found")
				}

				chatComp := chatCompRaw.(*engine.ChatComponent)
				if len(chatComp.Messages) == 0 {
					t.Error("no messages in chat component")
				}

				// Verify last message
				lastMsg := chatComp.Messages[len(chatComp.Messages)-1]
				if lastMsg.SenderID != senderID {
					t.Errorf("message sender ID = %v, want %v", lastMsg.SenderID, senderID)
				}
				if lastMsg.Channel != tt.channel {
					t.Errorf("message channel = %v, want %v", lastMsg.Channel, tt.channel)
				}
				if lastMsg.Content != tt.content {
					t.Errorf("message content = %q, want %q", lastMsg.Content, tt.content)
				}
				if lastMsg.ID == "" {
					t.Error("message ID is empty")
				}
			}
		})
	}
}

// TestSendMultipleMessages tests sending multiple messages
func TestSendMultipleMessages(t *testing.T) {
	world := engine.NewWorld()
	cs := NewChatSystem(world)

	sender := world.CreateEntity()
	world.Update(0)

	messages := []string{"First", "Second", "Third"}
	for _, content := range messages {
		err := cs.SendMessage(sender.ID, engine.ChatGlobal, content)
		if err != nil {
			t.Fatalf("SendMessage() failed: %v", err)
		}
	}

	chatCompRaw, ok := sender.GetComponent("chat")
	if !ok {
		t.Fatal("chat component not found")
	}

	chatComp := chatCompRaw.(*engine.ChatComponent)
	if len(chatComp.Messages) != len(messages) {
		t.Errorf("message count = %d, want %d", len(chatComp.Messages), len(messages))
	}

	// Verify message order
	for i, msg := range chatComp.Messages {
		if msg.Content != messages[i] {
			t.Errorf("message[%d] content = %q, want %q", i, msg.Content, messages[i])
		}
	}
}

// TestGenerateMessageID tests message ID generation
func TestGenerateMessageID(t *testing.T) {
	// Generate multiple IDs and verify they're unique
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id, err := generateMessageID()
		if err != nil {
			t.Fatalf("generateMessageID() failed: %v", err)
		}
		if id == "" {
			t.Error("generated empty message ID")
		}
		if ids[id] {
			t.Errorf("duplicate message ID generated: %s", id)
		}
		ids[id] = true
	}

	if len(ids) != 100 {
		t.Errorf("unique IDs = %d, want 100", len(ids))
	}
}

// TestChatChannelTypes tests different chat channel types
func TestChatChannelTypes(t *testing.T) {
	world := engine.NewWorld()
	cs := NewChatSystem(world)

	sender := world.CreateEntity()
	world.Update(0)

	channels := []engine.ChatChannel{
		engine.ChatGlobal,
		engine.ChatLocal,
		engine.ChatParty,
		engine.ChatWhisper,
	}

	for _, channel := range channels {
		err := cs.SendMessage(sender.ID, channel, "test")
		if err != nil {
			t.Errorf("SendMessage() failed for channel %d: %v", channel, err)
		}
	}

	chatCompRaw, ok := sender.GetComponent("chat")
	if !ok {
		t.Fatal("chat component not found")
	}

	chatComp := chatCompRaw.(*engine.ChatComponent)
	if len(chatComp.Messages) != len(channels) {
		t.Errorf("message count = %d, want %d", len(chatComp.Messages), len(channels))
	}
}

// TestChatComponentCreation tests automatic chat component creation
func TestChatComponentCreation(t *testing.T) {
	world := engine.NewWorld()
	cs := NewChatSystem(world)

	sender := world.CreateEntity()
	world.Update(0)

	// Initially, no chat component
	_, ok := sender.GetComponent("chat")
	if ok {
		t.Error("chat component should not exist initially")
	}

	// Send message should create component
	err := cs.SendMessage(sender.ID, engine.ChatGlobal, "test")
	if err != nil {
		t.Fatalf("SendMessage() failed: %v", err)
	}

	// Now should have chat component
	chatCompRaw, ok := sender.GetComponent("chat")
	if !ok {
		t.Fatal("chat component not created")
	}

	chatComp := chatCompRaw.(*engine.ChatComponent)
	if len(chatComp.Messages) != 1 {
		t.Errorf("message count = %d, want 1", len(chatComp.Messages))
	}
	if len(chatComp.ActiveChannels) == 0 {
		t.Error("no active channels")
	}
}

// BenchmarkNewChatSystem benchmarks chat system creation
func BenchmarkNewChatSystem(b *testing.B) {
	world := engine.NewWorld()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewChatSystem(world)
	}
}

// BenchmarkSendMessage benchmarks sending a message
func BenchmarkSendMessage(b *testing.B) {
	world := engine.NewWorld()
	cs := NewChatSystem(world)
	sender := world.CreateEntity()
	world.Update(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cs.SendMessage(sender.ID, engine.ChatGlobal, "test message")
	}
}

// BenchmarkGenerateMessageID benchmarks message ID generation
func BenchmarkGenerateMessageID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generateMessageID()
	}
}

// BenchmarkSendMessageWithExistingComponent benchmarks sending when component exists
func BenchmarkSendMessageWithExistingComponent(b *testing.B) {
	world := engine.NewWorld()
	cs := NewChatSystem(world)
	sender := world.CreateEntity()
	world.Update(0)

	// Create component first
	cs.SendMessage(sender.ID, engine.ChatGlobal, "initial")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cs.SendMessage(sender.ID, engine.ChatGlobal, "test")
	}
}

// TestStructuredErrors tests that errors include correlation IDs and context
func TestStructuredErrors(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*engine.World) uint64
		content        string
		expectedType   string
		expectedFields []string
	}{
		{
			name: "rate limit error includes correlation ID",
			setup: func(w *engine.World) uint64 {
				sender := w.CreateEntity()
				return sender.ID
			},
			content:        "test",
			expectedType:   "RateLimit",
			expectedFields: []string{"playerID", "channel", "limit"},
		},
		{
			name: "validation error includes correlation ID",
			setup: func(w *engine.World) uint64 {
				sender := w.CreateEntity()
				return sender.ID
			},
			content:        "", // empty message triggers validation error
			expectedType:   "Validation",
			expectedFields: []string{"playerID", "channel", "contentLength"},
		},
		{
			name: "entity not found error includes correlation ID",
			setup: func(w *engine.World) uint64 {
				return 99999 // non-existent entity
			},
			content:        "test",
			expectedType:   "Network",
			expectedFields: []string{"playerID"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := engine.NewWorld()
			cs := NewChatSystem(world)
			senderID := tt.setup(world)
			world.Update(0)

			// For rate limit test, send many messages to trigger rate limit
			if tt.expectedType == "RateLimit" {
				for i := 0; i < ChatRateLimit+1; i++ {
					cs.SendMessage(senderID, engine.ChatGlobal, tt.content)
				}
			}

			err := cs.SendMessage(senderID, engine.ChatGlobal, tt.content)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			// Verify error message contains type
			if !strings.Contains(err.Error(), tt.expectedType) {
				t.Errorf("error message %q does not contain type %q", err.Error(), tt.expectedType)
			}

			// Verify error message contains correlation ID pattern
			if !strings.Contains(err.Error(), "[") || !strings.Contains(err.Error(), "]") {
				t.Errorf("error message %q does not contain correlation ID brackets", err.Error())
			}
		})
	}
}

// TestErrorCorrelationIDUniqueness tests that each error gets a unique correlation ID
func TestErrorCorrelationIDUniqueness(t *testing.T) {
	world := engine.NewWorld()
	cs := NewChatSystem(world)

	// Create sender that doesn't exist to trigger errors
	senderID := uint64(99999)

	errors := make([]error, 10)
	for i := 0; i < 10; i++ {
		errors[i] = cs.SendMessage(senderID, engine.ChatGlobal, "test")
	}

	// Extract correlation IDs from error messages
	correlationIDs := make(map[string]bool)
	for _, err := range errors {
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		// Extract correlation ID from error message (format: [Type][CorrelationID] message)
		msg := err.Error()
		start := strings.Index(msg, "][")
		end := strings.Index(msg[start+2:], "]")
		if start == -1 || end == -1 {
			t.Fatalf("error message %q does not contain correlation ID", msg)
		}
		correlationID := msg[start+2 : start+2+end]
		correlationIDs[correlationID] = true
	}

	// Verify all correlation IDs are unique
	if len(correlationIDs) != 10 {
		t.Errorf("expected 10 unique correlation IDs, got %d", len(correlationIDs))
	}
}

// TestErrorContextPreservation tests that error context is preserved
func TestErrorContextPreservation(t *testing.T) {
	world := engine.NewWorld()
	cs := NewChatSystem(world)

	// Test validation error with context
	sender := world.CreateEntity()
	world.Update(0)

	// Send empty message to trigger validation error
	err := cs.SendMessage(sender.ID, engine.ChatGlobal, "")
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	// Verify error is a VentureError with context
	ventureErr, ok := err.(*errors.VentureError)
	if !ok {
		t.Fatalf("expected *errors.VentureError, got %T", err)
	}

	// Verify context fields
	if ventureErr.Context["playerID"] != sender.ID {
		t.Errorf("context playerID = %v, want %v", ventureErr.Context["playerID"], sender.ID)
	}

	if ventureErr.Context["channel"] == "" {
		t.Error("context channel is empty")
	}

	if _, ok := ventureErr.Context["contentLength"]; !ok {
		t.Error("context contentLength is missing")
	}

	// Verify correlation ID
	if ventureErr.CorrelationID == "" {
		t.Error("correlation ID is empty")
	}
}
