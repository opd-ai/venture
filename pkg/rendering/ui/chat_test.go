package ui

import (
	"testing"
	"time"
)

// TestNewChatUI tests chat UI creation
func TestNewChatUI(t *testing.T) {
	ui := NewChatUI(10, 20, 400, 300)

	if ui == nil {
		t.Fatal("NewChatUI returned nil")
	}

	if ui.X != 10 || ui.Y != 20 {
		t.Errorf("Position incorrect: got (%d, %d), want (10, 20)", ui.X, ui.Y)
	}

	if ui.Width != 400 || ui.Height != 300 {
		t.Errorf("Size incorrect: got (%d, %d), want (400, 300)", ui.Width, ui.Height)
	}

	if len(ui.Channels) != 4 {
		t.Errorf("Expected 4 default channels, got %d", len(ui.Channels))
	}

	if ui.ActiveChannel != 0 {
		t.Errorf("Expected active channel 0, got %d", ui.ActiveChannel)
	}

	if ui.MaxMessages != 100 {
		t.Errorf("Expected MaxMessages 100, got %d", ui.MaxMessages)
	}
}

// TestAddMessage tests adding messages to the chat
func TestAddMessage(t *testing.T) {
	ui := NewChatUI(0, 0, 400, 300)

	msg := ChatMessage{
		SenderName: "Player1",
		Content:    "Hello world",
		Channel:    0,
		Timestamp:  time.Now(),
		IsSystem:   false,
	}

	ui.AddMessage(msg)

	if len(ui.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(ui.Messages))
	}

	if ui.Messages[0].Content != "Hello world" {
		t.Errorf("Message content incorrect: %q", ui.Messages[0].Content)
	}
}

// TestAddSystemMessage tests adding system messages
func TestAddSystemMessage(t *testing.T) {
	ui := NewChatUI(0, 0, 400, 300)

	ui.AddSystemMessage("Server restarting")

	if len(ui.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(ui.Messages))
	}

	if !ui.Messages[0].IsSystem {
		t.Error("Message should be marked as system message")
	}

	if ui.Messages[0].Content != "Server restarting" {
		t.Errorf("Message content incorrect: %q", ui.Messages[0].Content)
	}
}

// TestUnreadCountIncrement tests unread count tracking
func TestUnreadCountIncrement(t *testing.T) {
	ui := NewChatUI(0, 0, 400, 300)
	ui.SetActiveChannel(0) // Global channel active

	// Add message to different channel
	msg := ChatMessage{
		SenderName: "Player2",
		Content:    "Local chat test",
		Channel:    1, // Local channel
		Timestamp:  time.Now(),
		IsSystem:   false,
	}

	ui.AddMessage(msg)

	// Check unread count for channel 1
	if ui.Channels[1].UnreadCount != 1 {
		t.Errorf("Expected unread count 1 for channel 1, got %d", ui.Channels[1].UnreadCount)
	}

	// Active channel should not increment
	if ui.Channels[0].UnreadCount != 0 {
		t.Errorf("Active channel should have 0 unread, got %d", ui.Channels[0].UnreadCount)
	}
}

// TestSetActiveChannel tests channel switching
func TestSetActiveChannel(t *testing.T) {
	ui := NewChatUI(0, 0, 400, 300)

	// Add some unread messages
	ui.Channels[1].UnreadCount = 5

	// Switch to channel 1
	ui.SetActiveChannel(1)

	if ui.ActiveChannel != 1 {
		t.Errorf("Expected active channel 1, got %d", ui.ActiveChannel)
	}

	// Unread count should be cleared
	if ui.Channels[1].UnreadCount != 0 {
		t.Errorf("Unread count should be cleared, got %d", ui.Channels[1].UnreadCount)
	}
}

// TestSetActiveChannelInvalid tests invalid channel IDs
func TestSetActiveChannelInvalid(t *testing.T) {
	ui := NewChatUI(0, 0, 400, 300)
	initialChannel := ui.ActiveChannel

	// Try invalid channel IDs
	ui.SetActiveChannel(-1)
	if ui.ActiveChannel != initialChannel {
		t.Error("Negative channel ID should be rejected")
	}

	ui.SetActiveChannel(999)
	if ui.ActiveChannel != initialChannel {
		t.Error("Out-of-range channel ID should be rejected")
	}
}

// TestInputText tests input field operations
func TestInputText(t *testing.T) {
	ui := NewChatUI(0, 0, 400, 300)

	// Append characters
	ui.AppendInputChar('H')
	ui.AppendInputChar('i')

	if ui.GetInputText() != "Hi" {
		t.Errorf("Expected input 'Hi', got %q", ui.GetInputText())
	}

	// Backspace
	ui.BackspaceInput()
	if ui.GetInputText() != "H" {
		t.Errorf("Expected input 'H' after backspace, got %q", ui.GetInputText())
	}

	// Clear
	ui.ClearInput()
	if ui.GetInputText() != "" {
		t.Errorf("Expected empty input after clear, got %q", ui.GetInputText())
	}
}

// TestInputTextMaxLength tests maximum input length
func TestInputTextMaxLength(t *testing.T) {
	ui := NewChatUI(0, 0, 400, 300)

	// Append 250 characters (exceeds 200 limit)
	for i := 0; i < 250; i++ {
		ui.AppendInputChar('A')
	}

	if len(ui.GetInputText()) > 200 {
		t.Errorf("Input length %d exceeds maximum 200", len(ui.GetInputText()))
	}
}

// TestInputActive tests input field activation
func TestInputActive(t *testing.T) {
	ui := NewChatUI(0, 0, 400, 300)

	if ui.IsInputActive() {
		t.Error("Input should not be active initially")
	}

	ui.SetInputActive(true)
	if !ui.IsInputActive() {
		t.Error("Input should be active after SetInputActive(true)")
	}

	ui.SetInputActive(false)
	if ui.IsInputActive() {
		t.Error("Input should not be active after SetInputActive(false)")
	}
}

// TestScrolling tests message history scrolling
func TestScrolling(t *testing.T) {
	ui := NewChatUI(0, 0, 400, 300)

	// Add many messages
	for i := 0; i < 50; i++ {
		ui.AddMessage(ChatMessage{
			SenderName: "Player",
			Content:    "Test message",
			Channel:    0,
			Timestamp:  time.Now(),
		})
	}

	initialOffset := ui.ScrollOffset

	// Scroll up
	ui.ScrollUp()
	if ui.ScrollOffset <= initialOffset {
		t.Error("ScrollUp should increase offset")
	}

	// Scroll down
	ui.ScrollDown()
	if ui.ScrollOffset >= initialOffset+1 {
		t.Error("ScrollDown should decrease offset")
	}
}

// TestMaxMessages tests message history limit
func TestMaxMessages(t *testing.T) {
	ui := NewChatUI(0, 0, 400, 300)
	ui.MaxMessages = 10

	// Add more than MaxMessages
	for i := 0; i < 20; i++ {
		ui.AddMessage(ChatMessage{
			SenderName: "Player",
			Content:    "Test",
			Channel:    0,
			Timestamp:  time.Now(),
		})
	}

	// Force update to trim messages
	ui.Update(0.016)

	if len(ui.Messages) > ui.MaxMessages {
		t.Errorf("Message count %d exceeds MaxMessages %d", len(ui.Messages), ui.MaxMessages)
	}
}

// TestSetPosition tests UI positioning
func TestSetPosition(t *testing.T) {
	ui := NewChatUI(0, 0, 400, 300)

	ui.SetPosition(100, 200)

	if ui.X != 100 || ui.Y != 200 {
		t.Errorf("Position incorrect: got (%d, %d), want (100, 200)", ui.X, ui.Y)
	}
}

// TestSetSize tests UI resizing
func TestSetSize(t *testing.T) {
	ui := NewChatUI(0, 0, 400, 300)

	ui.SetSize(800, 600)

	if ui.Width != 800 || ui.Height != 600 {
		t.Errorf("Size incorrect: got (%d, %d), want (800, 600)", ui.Width, ui.Height)
	}
}

// TestGetBounds tests bounding box calculation
func TestGetBounds(t *testing.T) {
	ui := NewChatUI(10, 20, 400, 300)

	bounds := ui.GetBounds()

	if bounds.Min.X != 10 || bounds.Min.Y != 20 {
		t.Errorf("Bounds min incorrect: got (%d, %d), want (10, 20)", bounds.Min.X, bounds.Min.Y)
	}

	if bounds.Max.X != 410 || bounds.Max.Y != 320 {
		t.Errorf("Bounds max incorrect: got (%d, %d), want (410, 320)", bounds.Max.X, bounds.Max.Y)
	}
}

// TestContainsPoint tests point-in-bounds checking
func TestContainsPoint(t *testing.T) {
	ui := NewChatUI(10, 20, 400, 300)

	tests := []struct {
		name     string
		x, y     int
		expected bool
	}{
		{"inside", 200, 150, true},
		{"outside left", 5, 150, false},
		{"outside right", 420, 150, false},
		{"outside top", 200, 10, false},
		{"outside bottom", 200, 330, false},
		{"top-left corner", 10, 20, true},
		{"bottom-right corner", 409, 319, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ui.ContainsPoint(tt.x, tt.y)
			if result != tt.expected {
				t.Errorf("ContainsPoint(%d, %d) = %v, want %v", tt.x, tt.y, result, tt.expected)
			}
		})
	}
}

// TestHandleClick tests mouse click handling
func TestHandleClick(t *testing.T) {
	ui := NewChatUI(0, 0, 400, 300)
	ui.SetActiveChannel(0)

	// Click on channel tab 1 (Local)
	tabWidth := ui.Width / len(ui.Channels)
	clickX := tabWidth + 10
	clickY := ui.Y + ui.Padding + 10

	ui.HandleClick(clickX, clickY)

	if ui.ActiveChannel != 1 {
		t.Errorf("Expected active channel 1 after click, got %d", ui.ActiveChannel)
	}
}

// TestUpdate tests UI state updates
func TestUpdate(t *testing.T) {
	ui := NewChatUI(0, 0, 400, 300)
	ui.SetInputActive(true)

	initialCursor := ui.CursorVisible

	// Update should toggle cursor after enough time
	ui.lastBlink = time.Now().Add(-600 * time.Millisecond)
	ui.Update(0.016)

	if ui.CursorVisible == initialCursor {
		t.Error("Cursor should toggle after blink interval")
	}
}

// TestFormatMessage tests message formatting
func TestFormatMessage(t *testing.T) {
	ui := NewChatUI(0, 0, 400, 300)

	tests := []struct {
		name     string
		msg      ChatMessage
		contains []string
	}{
		{
			name: "normal message",
			msg: ChatMessage{
				SenderName: "Alice",
				Content:    "Hello",
				Timestamp:  time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC),
				IsSystem:   false,
			},
			contains: []string{"14:30", "Alice", "Hello"},
		},
		{
			name: "system message",
			msg: ChatMessage{
				Content:  "Server restart",
				IsSystem: true,
			},
			contains: []string{"SYSTEM", "Server restart"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted := ui.formatMessage(tt.msg)

			for _, substr := range tt.contains {
				if !contains(formatted, substr) {
					t.Errorf("Formatted message %q does not contain %q", formatted, substr)
				}
			}
		})
	}
}

// TestDefaultChannels tests default channel configuration
func TestDefaultChannels(t *testing.T) {
	channels := defaultChannels()

	if len(channels) != 4 {
		t.Errorf("Expected 4 default channels, got %d", len(channels))
	}

	expectedNames := []string{"Global", "Local", "Party", "Whisper"}
	for i, name := range expectedNames {
		if channels[i].Name != name {
			t.Errorf("Channel %d: expected name %q, got %q", i, name, channels[i].Name)
		}
		if channels[i].ID != i {
			t.Errorf("Channel %d: expected ID %d, got %d", i, i, channels[i].ID)
		}
	}
}

// Helper function for string containment check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && (s[:len(substr)] == substr || contains(s[1:], substr))))
}

// BenchmarkAddMessage benchmarks message addition
func BenchmarkAddMessage(b *testing.B) {
	ui := NewChatUI(0, 0, 400, 300)
	msg := ChatMessage{
		SenderName: "Player",
		Content:    "Test message",
		Channel:    0,
		Timestamp:  time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ui.AddMessage(msg)
	}
}

// BenchmarkUpdate benchmarks UI updates
func BenchmarkUpdate(b *testing.B) {
	ui := NewChatUI(0, 0, 400, 300)

	// Add some messages
	for i := 0; i < 50; i++ {
		ui.AddMessage(ChatMessage{
			SenderName: "Player",
			Content:    "Message",
			Channel:    0,
			Timestamp:  time.Now(),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ui.Update(0.016)
	}
}

// BenchmarkFormatMessage benchmarks message formatting
func BenchmarkFormatMessage(b *testing.B) {
	ui := NewChatUI(0, 0, 400, 300)
	msg := ChatMessage{
		SenderName: "Player",
		Content:    "Test message content",
		Timestamp:  time.Now(),
		IsSystem:   false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ui.formatMessage(msg)
	}
}
