package persistence

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNewChatHistory(t *testing.T) {
	playerID := "player123"
	history := NewChatHistory(playerID)

	if history == nil {
		t.Fatal("NewChatHistory returned nil")
	}
	if history.PlayerID != playerID {
		t.Errorf("expected PlayerID %q, got %q", playerID, history.PlayerID)
	}
	if len(history.Messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(history.Messages))
	}
	if history.Version != 1 {
		t.Errorf("expected version 1, got %d", history.Version)
	}
}

func TestAddMessage(t *testing.T) {
	history := NewChatHistory("player1")

	tests := []struct {
		name    string
		msg     *Message
		wantErr bool
	}{
		{
			name: "valid message",
			msg: &Message{
				ID:        "msg1",
				Sender:    "player1",
				Recipient: "player2",
				Channel:   "whisper",
				Content:   "Hello",
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
		},
		{
			name: "empty ID",
			msg: &Message{
				ID:      "",
				Sender:  "player1",
				Content: "Hello",
			},
			wantErr: true,
		},
		{
			name: "empty sender",
			msg: &Message{
				ID:      "msg2",
				Sender:  "",
				Content: "Hello",
			},
			wantErr: true,
		},
		{
			name: "empty content (not deleted)",
			msg: &Message{
				ID:      "msg3",
				Sender:  "player1",
				Content: "",
				Deleted: false,
			},
			wantErr: true,
		},
		{
			name: "empty content (deleted)",
			msg: &Message{
				ID:      "msg4",
				Sender:  "player1",
				Content: "",
				Deleted: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := history.AddMessage(tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddMessage_Deduplication(t *testing.T) {
	history := NewChatHistory("player1")

	msg := &Message{
		ID:        "msg1",
		Sender:    "player1",
		Content:   "Hello",
		Timestamp: time.Now(),
	}

	// Add message twice
	if err := history.AddMessage(msg); err != nil {
		t.Fatalf("first AddMessage failed: %v", err)
	}
	if err := history.AddMessage(msg); err != nil {
		t.Fatalf("second AddMessage failed: %v", err)
	}

	messages := history.GetMessages(nil)
	if len(messages) != 1 {
		t.Errorf("expected 1 message after deduplication, got %d", len(messages))
	}
}

func TestAddMessage_MaxLimit(t *testing.T) {
	history := NewChatHistory("player1")

	// Add more than MaxMessagesPerPlayer
	for i := 0; i < MaxMessagesPerPlayer+100; i++ {
		msg := &Message{
			ID:        fmt.Sprintf("msg%d", i),
			Sender:    "player1",
			Content:   fmt.Sprintf("Message %d", i),
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
		}
		if err := history.AddMessage(msg); err != nil {
			t.Fatalf("AddMessage %d failed: %v", i, err)
		}
	}

	messages := history.GetMessages(nil)
	if len(messages) != MaxMessagesPerPlayer {
		t.Errorf("expected %d messages, got %d", MaxMessagesPerPlayer, len(messages))
	}

	// Verify oldest messages were removed (LRU)
	firstMsg := messages[0]
	expectedID := fmt.Sprintf("msg%d", 100) // Should start at msg100
	if firstMsg.ID != expectedID {
		t.Errorf("expected first message ID %q, got %q", expectedID, firstMsg.ID)
	}
}

func TestGetMessages_NoFilter(t *testing.T) {
	history := NewChatHistory("player1")

	// Add test messages
	for i := 0; i < 10; i++ {
		msg := &Message{
			ID:        fmt.Sprintf("msg%d", i),
			Sender:    fmt.Sprintf("player%d", i%3),
			Content:   fmt.Sprintf("Content %d", i),
			Timestamp: time.Now(),
		}
		history.AddMessage(msg)
	}

	messages := history.GetMessages(nil)
	if len(messages) != 10 {
		t.Errorf("expected 10 messages, got %d", len(messages))
	}
}

func TestGetMessages_WithFilter(t *testing.T) {
	history := NewChatHistory("player1")
	now := time.Now()

	messages := []*Message{
		{ID: "msg1", Sender: "alice", Recipient: "bob", Channel: "whisper", Content: "Hi", Timestamp: now.Add(-2 * time.Hour)},
		{ID: "msg2", Sender: "bob", Recipient: "alice", Channel: "whisper", Content: "Hello", Timestamp: now.Add(-1 * time.Hour)},
		{ID: "msg3", Sender: "alice", Recipient: "", Channel: "global", Content: "Anyone?", Timestamp: now},
		{ID: "msg4", Sender: "charlie", Recipient: "", Channel: "guild", Content: "Raid?", Timestamp: now.Add(1 * time.Hour)},
	}

	for _, msg := range messages {
		history.AddMessage(msg)
	}

	tests := []struct {
		name    string
		filter  *MessageFilter
		wantIDs []string
	}{
		{
			name:    "filter by sender",
			filter:  &MessageFilter{Sender: "alice"},
			wantIDs: []string{"msg1", "msg3"},
		},
		{
			name:    "filter by channel",
			filter:  &MessageFilter{Channel: "whisper"},
			wantIDs: []string{"msg1", "msg2"},
		},
		{
			name:    "filter by time (after)",
			filter:  &MessageFilter{After: now.Add(-90 * time.Minute)},
			wantIDs: []string{"msg2", "msg3", "msg4"},
		},
		{
			name:    "filter by time (before)",
			filter:  &MessageFilter{Before: now.Add(-30 * time.Minute)},
			wantIDs: []string{"msg1", "msg2"},
		},
		{
			name:    "combined filter",
			filter:  &MessageFilter{Sender: "alice", Channel: "whisper"},
			wantIDs: []string{"msg1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := history.GetMessages(tt.filter)
			if len(results) != len(tt.wantIDs) {
				t.Errorf("expected %d results, got %d", len(tt.wantIDs), len(results))
				return
			}
			for i, msg := range results {
				if msg.ID != tt.wantIDs[i] {
					t.Errorf("result[%d]: expected ID %q, got %q", i, tt.wantIDs[i], msg.ID)
				}
			}
		})
	}
}

func TestDeleteOldMessages(t *testing.T) {
	history := NewChatHistory("player1")
	now := time.Now()

	messages := []*Message{
		{ID: "msg1", Sender: "alice", Content: "Old", Timestamp: now.Add(-40 * 24 * time.Hour)},  // 40 days old
		{ID: "msg2", Sender: "bob", Content: "Recent", Timestamp: now.Add(-10 * 24 * time.Hour)}, // 10 days old
		{ID: "msg3", Sender: "charlie", Content: "New", Timestamp: now},                          // Now
	}

	for _, msg := range messages {
		history.AddMessage(msg)
	}

	deleted := history.DeleteOldMessages(now)
	if deleted != 1 {
		t.Errorf("expected 1 deleted message, got %d", deleted)
	}

	remaining := history.GetMessages(nil)
	if len(remaining) != 2 {
		t.Errorf("expected 2 remaining messages, got %d", len(remaining))
	}

	// Verify old message was removed
	for _, msg := range remaining {
		if msg.ID == "msg1" {
			t.Error("old message was not deleted")
		}
	}
}

func TestChatHistory_SaveLoad(t *testing.T) {
	history := NewChatHistory("player1")

	// Add test messages
	for i := 0; i < 50; i++ {
		msg := &Message{
			ID:        fmt.Sprintf("msg%d", i),
			Sender:    "player1",
			Recipient: "player2",
			Channel:   "whisper",
			Content:   strings.Repeat("x", 100), // 100 chars per message
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
		}
		history.AddMessage(msg)
	}

	// Save
	data, err := history.Save()
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify compression occurred
	if len(data) == 0 {
		t.Error("saved data is empty")
	}

	// Load into new history
	loaded := NewChatHistory("player1")
	if err := loaded.Load(data); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify data integrity
	if loaded.PlayerID != history.PlayerID {
		t.Errorf("PlayerID mismatch: expected %q, got %q", history.PlayerID, loaded.PlayerID)
	}
	if len(loaded.Messages) != len(history.Messages) {
		t.Errorf("message count mismatch: expected %d, got %d", len(history.Messages), len(loaded.Messages))
	}
	if loaded.Version != history.Version {
		t.Errorf("version mismatch: expected %d, got %d", history.Version, loaded.Version)
	}

	// Verify messages
	for i, msg := range loaded.Messages {
		original := history.Messages[i]
		if msg.ID != original.ID {
			t.Errorf("message[%d] ID mismatch: expected %q, got %q", i, original.ID, msg.ID)
		}
		if msg.Content != original.Content {
			t.Errorf("message[%d] content mismatch", i)
		}
	}

	t.Logf("Saved %d messages in %d bytes (%.1fx compression)", len(history.Messages), len(data),
		float64(len(history.Messages)*100)/float64(len(data)))
}

func TestGetDelta(t *testing.T) {
	history := NewChatHistory("player1")

	// Add messages to create versions
	for i := 0; i < 10; i++ {
		msg := &Message{
			ID:        fmt.Sprintf("msg%d", i),
			Sender:    "player1",
			Content:   fmt.Sprintf("Message %d", i),
			Timestamp: time.Now(),
		}
		history.AddMessage(msg)
	}

	currentVersion := history.GetVersion()

	tests := []struct {
		name        string
		fromVersion int
		wantDelta   bool
	}{
		{
			name:        "delta from version 0",
			fromVersion: 0,
			wantDelta:   true,
		},
		{
			name:        "delta from old version",
			fromVersion: 5,
			wantDelta:   true,
		},
		{
			name:        "no delta (current version)",
			fromVersion: currentVersion,
			wantDelta:   false,
		},
		{
			name:        "no delta (future version)",
			fromVersion: currentVersion + 1,
			wantDelta:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delta := history.GetDelta(tt.fromVersion)
			hasDelta := delta != nil && len(delta) > 0
			if hasDelta != tt.wantDelta {
				t.Errorf("GetDelta(%d) returned delta=%v, want delta=%v", tt.fromVersion, hasDelta, tt.wantDelta)
			}
		})
	}
}

func TestApplyDelta(t *testing.T) {
	history := NewChatHistory("player1")

	// Add initial messages
	for i := 0; i < 5; i++ {
		msg := &Message{
			ID:        fmt.Sprintf("msg%d", i),
			Sender:    "player1",
			Content:   fmt.Sprintf("Message %d", i),
			Timestamp: time.Now(),
		}
		history.AddMessage(msg)
	}

	// Create delta with new messages
	delta := []*Message{
		{ID: "msg5", Sender: "player2", Content: "New 1", Timestamp: time.Now()},
		{ID: "msg6", Sender: "player2", Content: "New 2", Timestamp: time.Now()},
		{ID: "msg1", Sender: "player1", Content: "Duplicate", Timestamp: time.Now()}, // Duplicate
	}

	initialVersion := history.GetVersion()
	if err := history.ApplyDelta(delta); err != nil {
		t.Fatalf("ApplyDelta failed: %v", err)
	}

	messages := history.GetMessages(nil)
	expectedCount := 7 // 5 original + 2 new (1 duplicate ignored)
	if len(messages) != expectedCount {
		t.Errorf("expected %d messages after delta, got %d", expectedCount, len(messages))
	}

	// Verify version increased
	if history.GetVersion() <= initialVersion {
		t.Error("version did not increase after applying delta")
	}

	// Verify new messages exist
	hasMsg5 := false
	hasMsg6 := false
	for _, msg := range messages {
		if msg.ID == "msg5" {
			hasMsg5 = true
		}
		if msg.ID == "msg6" {
			hasMsg6 = true
		}
	}
	if !hasMsg5 || !hasMsg6 {
		t.Error("new messages from delta not found")
	}
}

func TestApplyDelta_Empty(t *testing.T) {
	history := NewChatHistory("player1")
	initialVersion := history.GetVersion()

	if err := history.ApplyDelta(nil); err != nil {
		t.Errorf("ApplyDelta(nil) failed: %v", err)
	}

	if err := history.ApplyDelta([]*Message{}); err != nil {
		t.Errorf("ApplyDelta([]) failed: %v", err)
	}

	// Verify version didn't change
	if history.GetVersion() != initialVersion {
		t.Error("version changed after applying empty delta")
	}
}

func TestMessageFilter_Matches(t *testing.T) {
	now := time.Now()
	msg := &Message{
		ID:        "msg1",
		Sender:    "alice",
		Recipient: "bob",
		Channel:   "whisper",
		Content:   "Hello",
		Timestamp: now,
	}

	tests := []struct {
		name   string
		filter *MessageFilter
		want   bool
	}{
		{
			name:   "no filter",
			filter: &MessageFilter{},
			want:   true,
		},
		{
			name:   "matching sender",
			filter: &MessageFilter{Sender: "alice"},
			want:   true,
		},
		{
			name:   "non-matching sender",
			filter: &MessageFilter{Sender: "charlie"},
			want:   false,
		},
		{
			name:   "matching recipient",
			filter: &MessageFilter{Recipient: "bob"},
			want:   true,
		},
		{
			name:   "matching channel",
			filter: &MessageFilter{Channel: "whisper"},
			want:   true,
		},
		{
			name:   "after timestamp",
			filter: &MessageFilter{After: now.Add(-1 * time.Hour)},
			want:   true,
		},
		{
			name:   "before timestamp",
			filter: &MessageFilter{Before: now.Add(1 * time.Hour)},
			want:   true,
		},
		{
			name:   "combined match",
			filter: &MessageFilter{Sender: "alice", Channel: "whisper"},
			want:   true,
		},
		{
			name:   "combined no match",
			filter: &MessageFilter{Sender: "alice", Channel: "global"},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.Matches(msg); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConcurrency(t *testing.T) {
	history := NewChatHistory("player1")

	// Concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				msg := &Message{
					ID:        fmt.Sprintf("msg-%d-%d", id, j),
					Sender:    fmt.Sprintf("player%d", id),
					Content:   fmt.Sprintf("Content %d", j),
					Timestamp: time.Now(),
				}
				history.AddMessage(msg)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify no data corruption
	messages := history.GetMessages(nil)
	if len(messages) == 0 {
		t.Error("no messages after concurrent writes")
	}
	if len(messages) > MaxMessagesPerPlayer {
		t.Errorf("exceeded max messages: %d > %d", len(messages), MaxMessagesPerPlayer)
	}
}

// Benchmarks

func BenchmarkAddMessage(b *testing.B) {
	history := NewChatHistory("player1")
	msg := &Message{
		ID:        "msg1",
		Sender:    "player1",
		Content:   "Hello world",
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg.ID = fmt.Sprintf("msg%d", i)
		history.AddMessage(msg)
	}
}

func BenchmarkGetMessages_NoFilter(b *testing.B) {
	history := NewChatHistory("player1")
	for i := 0; i < 1000; i++ {
		msg := &Message{
			ID:        fmt.Sprintf("msg%d", i),
			Sender:    "player1",
			Content:   "Hello",
			Timestamp: time.Now(),
		}
		history.AddMessage(msg)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		history.GetMessages(nil)
	}
}

func BenchmarkGetMessages_WithFilter(b *testing.B) {
	history := NewChatHistory("player1")
	for i := 0; i < 1000; i++ {
		msg := &Message{
			ID:        fmt.Sprintf("msg%d", i),
			Sender:    fmt.Sprintf("player%d", i%10),
			Channel:   "global",
			Content:   "Hello",
			Timestamp: time.Now(),
		}
		history.AddMessage(msg)
	}

	filter := &MessageFilter{Sender: "player1"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		history.GetMessages(filter)
	}
}

func BenchmarkChatHistory_Save(b *testing.B) {
	history := NewChatHistory("player1")
	for i := 0; i < 1000; i++ {
		msg := &Message{
			ID:        fmt.Sprintf("msg%d", i),
			Sender:    "player1",
			Content:   strings.Repeat("x", 100),
			Timestamp: time.Now(),
		}
		history.AddMessage(msg)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		history.Save()
	}
}

func BenchmarkChatHistory_Load(b *testing.B) {
	history := NewChatHistory("player1")
	for i := 0; i < 1000; i++ {
		msg := &Message{
			ID:        fmt.Sprintf("msg%d", i),
			Sender:    "player1",
			Content:   strings.Repeat("x", 100),
			Timestamp: time.Now(),
		}
		history.AddMessage(msg)
	}

	data, _ := history.Save()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loaded := NewChatHistory("player1")
		loaded.Load(data)
	}
}

func BenchmarkGetDelta(b *testing.B) {
	history := NewChatHistory("player1")
	for i := 0; i < 1000; i++ {
		msg := &Message{
			ID:        fmt.Sprintf("msg%d", i),
			Sender:    "player1",
			Content:   "Hello",
			Timestamp: time.Now(),
		}
		history.AddMessage(msg)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		history.GetDelta(500)
	}
}

func BenchmarkApplyDelta(b *testing.B) {
	history := NewChatHistory("player1")
	for i := 0; i < 500; i++ {
		msg := &Message{
			ID:        fmt.Sprintf("msg%d", i),
			Sender:    "player1",
			Content:   "Hello",
			Timestamp: time.Now(),
		}
		history.AddMessage(msg)
	}

	delta := []*Message{
		{ID: "new1", Sender: "player2", Content: "New", Timestamp: time.Now()},
		{ID: "new2", Sender: "player2", Content: "New", Timestamp: time.Now()},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		history.ApplyDelta(delta)
	}
}

// TimeProvider Tests

// chatMockTimeProvider implements TimeProvider for deterministic testing
type chatMockTimeProvider struct {
	fixedTime time.Time
}

func (m *chatMockTimeProvider) Now() time.Time {
	return m.fixedTime
}

func TestNewChatHistoryWithTimeProvider(t *testing.T) {
	fixedTime := time.Date(2026, 2, 17, 12, 0, 0, 0, time.UTC)
	mockTP := &chatMockTimeProvider{fixedTime: fixedTime}

	history := NewChatHistoryWithTimeProvider("player1", mockTP)

	if history == nil {
		t.Fatal("NewChatHistoryWithTimeProvider returned nil")
	}
	if history.PlayerID != "player1" {
		t.Errorf("expected PlayerID \"player1\", got %q", history.PlayerID)
	}
	if history.timeProvider != mockTP {
		t.Error("TimeProvider not set correctly")
	}
}

func TestChatHistorySetTimeProvider(t *testing.T) {
	history := NewChatHistory("player1")

	// Verify default time provider is set
	if history.timeProvider == nil {
		t.Fatal("Default timeProvider not set")
	}

	// Set a mock time provider
	fixedTime := time.Date(2026, 2, 17, 15, 0, 0, 0, time.UTC)
	mockTP := &chatMockTimeProvider{fixedTime: fixedTime}
	history.SetTimeProvider(mockTP)

	// Verify the time provider was set
	if history.timeProvider != mockTP {
		t.Error("TimeProvider not set correctly after SetTimeProvider")
	}
}

func TestChatHistoryTimeProviderAfterLoad(t *testing.T) {
	// Create history with mock time provider
	fixedTime := time.Date(2026, 2, 17, 12, 0, 0, 0, time.UTC)
	mockTP := &chatMockTimeProvider{fixedTime: fixedTime}

	history := NewChatHistoryWithTimeProvider("player1", mockTP)
	history.AddMessage(&Message{
		ID:        "msg1",
		Sender:    "player1",
		Content:   "Hello",
		Timestamp: fixedTime,
	})

	// Save history
	data, err := history.Save()
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load into new history with different time provider
	newTime := time.Date(2026, 2, 18, 12, 0, 0, 0, time.UTC)
	newMockTP := &chatMockTimeProvider{fixedTime: newTime}
	newHistory := NewChatHistoryWithTimeProvider("player2", newMockTP)

	err = newHistory.Load(data)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify loaded data
	if newHistory.PlayerID != "player1" {
		t.Errorf("expected PlayerID \"player1\" after load, got %q", newHistory.PlayerID)
	}

	// Verify time provider was preserved (not overwritten by JSON unmarshal)
	if newHistory.timeProvider != newMockTP {
		t.Error("TimeProvider was not preserved after Load")
	}
}

func TestChatHistoryDeleteOldMessagesDeterministic(t *testing.T) {
	// Create two histories with same mock time
	fixedTime := time.Date(2026, 2, 17, 12, 0, 0, 0, time.UTC)
	mockTP1 := &chatMockTimeProvider{fixedTime: fixedTime}
	mockTP2 := &chatMockTimeProvider{fixedTime: fixedTime}

	history1 := NewChatHistoryWithTimeProvider("player1", mockTP1)
	history2 := NewChatHistoryWithTimeProvider("player2", mockTP2)

	// Add identical messages with old timestamps
	oldTime := fixedTime.Add(-40 * 24 * time.Hour) // 40 days ago (beyond MaxMessageAge)
	for i := 0; i < 5; i++ {
		msg := &Message{
			ID:        fmt.Sprintf("msg%d", i),
			Sender:    "sender",
			Content:   "Old message",
			Timestamp: oldTime,
		}
		history1.AddMessage(msg)
		history2.AddMessage(msg)
	}

	// Delete old messages with same time
	deleted1 := history1.DeleteOldMessages(fixedTime)
	deleted2 := history2.DeleteOldMessages(fixedTime)

	// Deleted counts should be identical
	if deleted1 != deleted2 {
		t.Errorf("Expected deterministic deletion: h1=%d, h2=%d", deleted1, deleted2)
	}

	// Message counts should be identical
	msgs1 := history1.GetMessages(nil)
	msgs2 := history2.GetMessages(nil)
	if len(msgs1) != len(msgs2) {
		t.Errorf("Expected same message count: h1=%d, h2=%d", len(msgs1), len(msgs2))
	}
}

// TestGetDelta_WithDeletions tests changelog-based delta sync with message deletions
func TestGetDelta_WithDeletions(t *testing.T) {
	history := NewChatHistory("player1")

	// Add 10 messages
	for i := 0; i < 10; i++ {
		msg := &Message{
			ID:        fmt.Sprintf("msg%d", i),
			Sender:    "player1",
			Content:   fmt.Sprintf("Message %d", i),
			Timestamp: time.Now().Add(-time.Hour * 24 * time.Duration(40-i*2)),
		}
		history.AddMessage(msg)
	}

	versionAfterAdds := history.GetVersion()

	// Delete old messages (should delete messages older than 30 days)
	deletedCount := history.DeleteOldMessages(time.Now())
	if deletedCount == 0 {
		t.Fatal("expected some messages to be deleted")
	}

	versionAfterDeletes := history.GetVersion()

	// Get delta from version after adds (should only include messages still present)
	delta := history.GetDelta(versionAfterAdds)

	// Delta should not include deleted messages
	// It should be empty because no new messages were added after versionAfterAdds
	if len(delta) != 0 {
		t.Errorf("expected 0 messages in delta (deletions don't add messages), got %d", len(delta))
	}

	// Now add a new message
	newMsg := &Message{
		ID:        "msg_new",
		Sender:    "player1",
		Content:   "New message after deletions",
		Timestamp: time.Now(),
	}
	history.AddMessage(newMsg)

	// Get delta from version after deletions (should include the new message)
	delta2 := history.GetDelta(versionAfterDeletes)
	if len(delta2) != 1 {
		t.Errorf("expected 1 message in delta, got %d", len(delta2))
	}
	if len(delta2) > 0 && delta2[0].ID != "msg_new" {
		t.Errorf("expected delta to contain msg_new, got %s", delta2[0].ID)
	}
}

// TestGetDelta_ChangelogFallback tests that GetDelta falls back to full sync when
// fromVersion is older than the changelog
func TestGetDelta_ChangelogFallback(t *testing.T) {
	history := NewChatHistory("player1")

	// Add a message to get version 2
	msg1 := &Message{
		ID:        "msg1",
		Sender:    "player1",
		Content:   "First message",
		Timestamp: time.Now(),
	}
	history.AddMessage(msg1)
	version2 := history.GetVersion()

	// Add MaxChangelogSize + 100 more messages to overflow the changelog
	for i := 0; i < MaxChangelogSize+100; i++ {
		msg := &Message{
			ID:        fmt.Sprintf("msg_overflow_%d", i),
			Sender:    "player1",
			Content:   fmt.Sprintf("Overflow message %d", i),
			Timestamp: time.Now(),
		}
		history.AddMessage(msg)
	}

	// Request delta from version2 (which is now older than changelog)
	delta := history.GetDelta(version2)

	// Should fall back to returning all messages
	allMessages := history.GetMessages(nil)
	if len(delta) != len(allMessages) {
		t.Errorf("expected fallback to full sync: delta=%d, all=%d", len(delta), len(allMessages))
	}
}

// TestGetDelta_MultipleOperations tests delta sync across multiple add/delete operations
func TestGetDelta_MultipleOperations(t *testing.T) {
	history := NewChatHistory("player1")

	// Add 5 messages
	for i := 0; i < 5; i++ {
		msg := &Message{
			ID:        fmt.Sprintf("msg%d", i),
			Sender:    "player1",
			Content:   fmt.Sprintf("Message %d", i),
			Timestamp: time.Now().Add(-time.Hour * 24 * time.Duration(10-i)),
		}
		history.AddMessage(msg)
	}

	checkpointVersion := history.GetVersion()

	// Add 3 more messages
	for i := 5; i < 8; i++ {
		msg := &Message{
			ID:        fmt.Sprintf("msg%d", i),
			Sender:    "player1",
			Content:   fmt.Sprintf("Message %d", i),
			Timestamp: time.Now(),
		}
		history.AddMessage(msg)
	}

	// Delete old messages (should delete msg0, msg1, msg2)
	cutoffTime := time.Now().Add(-time.Hour * 24 * 7)
	history.DeleteOldMessages(cutoffTime)

	// Add 2 more messages
	for i := 8; i < 10; i++ {
		msg := &Message{
			ID:        fmt.Sprintf("msg%d", i),
			Sender:    "player1",
			Content:   fmt.Sprintf("Message %d", i),
			Timestamp: time.Now(),
		}
		history.AddMessage(msg)
	}

	// Get delta from checkpoint (should include msg5, msg6, msg7, msg8, msg9)
	// but NOT msg0, msg1, msg2 (deleted) or msg3, msg4 (added before checkpoint)
	delta := history.GetDelta(checkpointVersion)

	expectedIDs := map[string]bool{
		"msg5": true,
		"msg6": true,
		"msg7": true,
		"msg8": true,
		"msg9": true,
	}

	if len(delta) != len(expectedIDs) {
		t.Errorf("expected %d messages in delta, got %d", len(expectedIDs), len(delta))
	}

	for _, msg := range delta {
		if !expectedIDs[msg.ID] {
			t.Errorf("unexpected message in delta: %s", msg.ID)
		}
	}
}

// TestChangelog_CircularBuffer tests that changelog maintains size limit
func TestChangelog_CircularBuffer(t *testing.T) {
	history := NewChatHistory("player1")

	// Add MaxChangelogSize + 50 messages
	totalMessages := MaxChangelogSize + 50
	for i := 0; i < totalMessages; i++ {
		msg := &Message{
			ID:        fmt.Sprintf("msg%d", i),
			Sender:    "player1",
			Content:   fmt.Sprintf("Message %d", i),
			Timestamp: time.Now(),
		}
		history.AddMessage(msg)
	}

	// Changelog should be capped at MaxChangelogSize
	if len(history.Changelog) > MaxChangelogSize {
		t.Errorf("expected changelog size <= %d, got %d", MaxChangelogSize, len(history.Changelog))
	}

	// Verify oldest entries were discarded (first entry should not be for msg0)
	if len(history.Changelog) > 0 {
		firstEntry := history.Changelog[0]
		if firstEntry.MessageID == "msg0" {
			t.Error("expected oldest entry to be discarded")
		}
	}
}

// TestGetDelta_AddThenDelete tests a message added and then deleted in the same delta range
func TestGetDelta_AddThenDelete(t *testing.T) {
	history := NewChatHistory("player1")

	// Add initial messages
	for i := 0; i < 5; i++ {
		msg := &Message{
			ID:        fmt.Sprintf("msg%d", i),
			Sender:    "player1",
			Content:   fmt.Sprintf("Message %d", i),
			Timestamp: time.Now().Add(-time.Hour * 24 * time.Duration(i)),
		}
		history.AddMessage(msg)
	}

	checkpointVersion := history.GetVersion()

	// Add a message with old timestamp
	tempMsg := &Message{
		ID:        "temp_msg",
		Sender:    "player1",
		Content:   "Temporary message",
		Timestamp: time.Now().Add(-time.Hour * 24 * 40), // Old timestamp
	}
	history.AddMessage(tempMsg)

	// Delete old messages (should delete temp_msg and some others)
	history.DeleteOldMessages(time.Now())

	// Get delta from checkpoint
	delta := history.GetDelta(checkpointVersion)

	// temp_msg should NOT be in delta (added then deleted in this range)
	for _, msg := range delta {
		if msg.ID == "temp_msg" {
			t.Error("expected temp_msg to not be in delta (added then deleted)")
		}
	}
}

// BenchmarkGetDelta_Changelog benchmarks the new changelog-based GetDelta
func BenchmarkGetDelta_Changelog(b *testing.B) {
	history := NewChatHistory("player1")

	// Add 1000 messages
	for i := 0; i < 1000; i++ {
		msg := &Message{
			ID:        fmt.Sprintf("msg%d", i),
			Sender:    "player1",
			Content:   "Hello",
			Timestamp: time.Now(),
		}
		history.AddMessage(msg)
	}

	fromVersion := history.GetVersion() - 100

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		history.GetDelta(fromVersion)
	}
}

// BenchmarkChangelog_Append benchmarks changelog append with circular buffer
func BenchmarkChangelog_Append(b *testing.B) {
	history := NewChatHistory("player1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg := &Message{
			ID:        fmt.Sprintf("msg%d", i),
			Sender:    "player1",
			Content:   "Benchmark message",
			Timestamp: time.Now(),
		}
		history.AddMessage(msg)
	}
}
