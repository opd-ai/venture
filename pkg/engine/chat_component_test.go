package engine

import (
	"testing"
	"time"
)

func TestChatChannelString(t *testing.T) {
	tests := []struct {
		channel ChatChannel
		want    string
	}{
		{ChatGlobal, "Global"},
		{ChatLocal, "Local"},
		{ChatParty, "Party"},
		{ChatWhisper, "Whisper"},
		{ChatChannel(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.channel.String()
			if got != tt.want {
				t.Errorf("ChatChannel.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewChatComponent(t *testing.T) {
	comp := NewChatComponent()
	if comp == nil {
		t.Fatal("NewChatComponent returned nil")
	}
	if comp.Type() != "chat" {
		t.Errorf("Type() = %v, want chat", comp.Type())
	}
	if len(comp.Messages) != 0 {
		t.Errorf("Messages length = %d, want 0", len(comp.Messages))
	}
	if comp.UnreadCount != 0 {
		t.Errorf("UnreadCount = %d, want 0", comp.UnreadCount)
	}
	if comp.MaxHistorySize != 100 {
		t.Errorf("MaxHistorySize = %d, want 100", comp.MaxHistorySize)
	}
	if comp.LocalRadius != 10.0 {
		t.Errorf("LocalRadius = %f, want 10.0", comp.LocalRadius)
	}
	if !comp.IsChannelActive(ChatGlobal) {
		t.Error("Global channel not active by default")
	}
	if !comp.IsChannelActive(ChatLocal) {
		t.Error("Local channel not active by default")
	}
}

func TestChatComponentAddMessage(t *testing.T) {
	comp := NewChatComponent()
	msg := ChatMessage{
		ID:        "test-123",
		SenderID:  1,
		Channel:   ChatGlobal,
		Content:   "Hello",
		Timestamp: time.Now(),
	}

	comp.AddMessage(msg)

	if len(comp.Messages) != 1 {
		t.Errorf("Messages length = %d, want 1", len(comp.Messages))
	}
	if comp.UnreadCount != 1 {
		t.Errorf("UnreadCount = %d, want 1", comp.UnreadCount)
	}
	if comp.Messages[0].Content != "Hello" {
		t.Errorf("Message content = %v, want Hello", comp.Messages[0].Content)
	}
}

func TestChatComponentAddMessageTrimHistory(t *testing.T) {
	comp := NewChatComponent()
	comp.MaxHistorySize = 10

	// Add 15 messages
	for i := 0; i < 15; i++ {
		msg := ChatMessage{
			ID:        string(rune('a' + i)),
			SenderID:  uint64(i),
			Channel:   ChatGlobal,
			Content:   string(rune('a' + i)),
			Timestamp: time.Now(),
		}
		comp.AddMessage(msg)
	}

	// Should trim to MaxHistorySize
	if len(comp.Messages) != 10 {
		t.Errorf("Messages length = %d, want 10 (trimmed)", len(comp.Messages))
	}
	if comp.UnreadCount != 15 {
		t.Errorf("UnreadCount = %d, want 15 (not trimmed)", comp.UnreadCount)
	}
	// First message should be "f" (6th message, messages 0-4 trimmed, so f is first)
	if comp.Messages[0].Content != "f" {
		t.Errorf("First message after trim = %v, want f", comp.Messages[0].Content)
	}
}

func TestChatComponentMarkAllRead(t *testing.T) {
	comp := NewChatComponent()
	comp.AddMessage(ChatMessage{ID: "1", Content: "msg1", Timestamp: time.Now()})
	comp.AddMessage(ChatMessage{ID: "2", Content: "msg2", Timestamp: time.Now()})

	if comp.UnreadCount != 2 {
		t.Errorf("UnreadCount before MarkAllRead = %d, want 2", comp.UnreadCount)
	}

	comp.MarkAllRead()

	if comp.UnreadCount != 0 {
		t.Errorf("UnreadCount after MarkAllRead = %d, want 0", comp.UnreadCount)
	}
}

func TestChatComponentGetMessageByID(t *testing.T) {
	comp := NewChatComponent()
	msg1 := ChatMessage{ID: "test-123", Content: "Hello", Timestamp: time.Now()}
	msg2 := ChatMessage{ID: "test-456", Content: "World", Timestamp: time.Now()}
	comp.AddMessage(msg1)
	comp.AddMessage(msg2)

	// Find existing message
	found := comp.GetMessageByID("test-123")
	if found == nil {
		t.Fatal("GetMessageByID returned nil for existing message")
	}
	if found.Content != "Hello" {
		t.Errorf("Found message content = %v, want Hello", found.Content)
	}

	// Find non-existent message
	notFound := comp.GetMessageByID("test-999")
	if notFound != nil {
		t.Error("GetMessageByID returned non-nil for non-existent message")
	}
}

func TestChatComponentGetMessagesForChannel(t *testing.T) {
	comp := NewChatComponent()
	comp.AddMessage(ChatMessage{ID: "1", Channel: ChatGlobal, Content: "Global 1", Timestamp: time.Now()})
	comp.AddMessage(ChatMessage{ID: "2", Channel: ChatLocal, Content: "Local 1", Timestamp: time.Now()})
	comp.AddMessage(ChatMessage{ID: "3", Channel: ChatGlobal, Content: "Global 2", Timestamp: time.Now()})

	globalMsgs := comp.GetMessagesForChannel(ChatGlobal)
	if len(globalMsgs) != 2 {
		t.Errorf("Global messages count = %d, want 2", len(globalMsgs))
	}

	localMsgs := comp.GetMessagesForChannel(ChatLocal)
	if len(localMsgs) != 1 {
		t.Errorf("Local messages count = %d, want 1", len(localMsgs))
	}

	partyMsgs := comp.GetMessagesForChannel(ChatParty)
	if len(partyMsgs) != 0 {
		t.Errorf("Party messages count = %d, want 0", len(partyMsgs))
	}
}

func TestChatComponentChannelSubscription(t *testing.T) {
	comp := NewChatComponent()

	// Default subscriptions
	if !comp.IsChannelActive(ChatGlobal) {
		t.Error("ChatGlobal not active by default")
	}

	// Unsubscribe
	comp.UnsubscribeChannel(ChatGlobal)
	if comp.IsChannelActive(ChatGlobal) {
		t.Error("ChatGlobal still active after unsubscribe")
	}

	// Subscribe
	comp.SubscribeChannel(ChatParty)
	if !comp.IsChannelActive(ChatParty) {
		t.Error("ChatParty not active after subscribe")
	}

	// Duplicate subscription should be idempotent
	comp.SubscribeChannel(ChatParty)
	count := 0
	for _, ch := range comp.ActiveChannels {
		if ch == ChatParty {
			count++
		}
	}
	if count != 1 {
		t.Errorf("ChatParty appears %d times, want 1", count)
	}
}

func TestChatComponentMuting(t *testing.T) {
	comp := NewChatComponent()

	// Not muted initially
	if comp.IsMuted() {
		t.Error("Component muted initially")
	}

	// Apply 1-second mute
	comp.ApplyMute(1 * time.Second)
	if !comp.IsMuted() {
		t.Error("Component not muted after ApplyMute")
	}

	// Wait for mute to expire
	time.Sleep(1100 * time.Millisecond)
	if comp.IsMuted() {
		t.Error("Component still muted after expiry")
	}
}

func TestChatComponentMuteDoubling(t *testing.T) {
	comp := NewChatComponent()

	// First violation: 1 second
	comp.ApplyMute(1 * time.Second)
	if comp.ViolationCount != 1 {
		t.Errorf("ViolationCount = %d, want 1", comp.ViolationCount)
	}

	// Wait for expiry
	time.Sleep(1100 * time.Millisecond)

	// Second violation: should be 2 seconds
	startTime := time.Now()
	comp.ApplyMute(1 * time.Second)
	expectedExpiry := startTime.Add(2 * time.Second)

	// Allow 100ms tolerance for timing
	if comp.MuteExpiry.Before(expectedExpiry.Add(-100*time.Millisecond)) ||
		comp.MuteExpiry.After(expectedExpiry.Add(100*time.Millisecond)) {
		t.Errorf("MuteExpiry = %v, want ~%v (2x duration)", comp.MuteExpiry, expectedExpiry)
	}
}

func TestChatComponentMuteMaxDuration(t *testing.T) {
	comp := NewChatComponent()

	// Apply many violations to trigger max duration
	for i := 0; i < 10; i++ {
		comp.ApplyMute(1 * time.Minute)
		time.Sleep(1 * time.Millisecond) // Small delay to avoid timing issues
	}

	// Mute duration should be capped at 10 minutes
	maxExpiry := time.Now().Add(10 * time.Minute)
	if comp.MuteExpiry.After(maxExpiry.Add(1 * time.Second)) {
		t.Errorf("MuteExpiry exceeds 10-minute cap: %v", comp.MuteExpiry)
	}
}

func TestChatComponentCanSendMessage(t *testing.T) {
	comp := NewChatComponent()

	// Can send initially
	if !comp.CanSendMessage(ChatGlobal) {
		t.Error("Cannot send message initially")
	}

	// Record message sent
	comp.RecordMessageSent(ChatGlobal)

	// Cannot send immediately (3-second cooldown for global)
	if comp.CanSendMessage(ChatGlobal) {
		t.Error("Can send message during cooldown")
	}

	// Wait for cooldown
	time.Sleep(3100 * time.Millisecond)

	// Can send after cooldown
	if !comp.CanSendMessage(ChatGlobal) {
		t.Error("Cannot send message after cooldown")
	}
}

func TestChatComponentCanSendMessageWhileMuted(t *testing.T) {
	comp := NewChatComponent()

	// Apply mute
	comp.ApplyMute(10 * time.Second)

	// Cannot send while muted (regardless of cooldown)
	if comp.CanSendMessage(ChatGlobal) {
		t.Error("Can send message while muted")
	}
}

func TestChatComponentGetChannelCooldown(t *testing.T) {
	comp := NewChatComponent()

	tests := []struct {
		channel  ChatChannel
		expected time.Duration
	}{
		{ChatGlobal, 3 * time.Second},
		{ChatLocal, 1 * time.Second},
		{ChatParty, 500 * time.Millisecond},
		{ChatWhisper, 500 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.channel.String(), func(t *testing.T) {
			got := comp.GetChannelCooldown(tt.channel)
			if got != tt.expected {
				t.Errorf("GetChannelCooldown(%v) = %v, want %v", tt.channel, got, tt.expected)
			}
		})
	}
}

func TestChatComponentMegaphone(t *testing.T) {
	comp := NewChatComponent()
	comp.MegaphoneUses = 3

	// Activate megaphone
	if !comp.ActivateMegaphone() {
		t.Fatal("ActivateMegaphone failed")
	}
	if !comp.MegaphoneActive {
		t.Error("MegaphoneActive not set")
	}
	if comp.LocalRadius != 30.0 {
		t.Errorf("LocalRadius = %f, want 30.0", comp.LocalRadius)
	}
	if comp.MegaphoneUses != 2 {
		t.Errorf("MegaphoneUses = %d, want 2", comp.MegaphoneUses)
	}

	// Deactivate megaphone
	comp.DeactivateMegaphone()
	if comp.MegaphoneActive {
		t.Error("MegaphoneActive still set")
	}
	if comp.LocalRadius != 10.0 {
		t.Errorf("LocalRadius = %f, want 10.0 (default)", comp.LocalRadius)
	}
}

func TestChatComponentMegaphoneNoUses(t *testing.T) {
	comp := NewChatComponent()
	comp.MegaphoneUses = 0

	if comp.ActivateMegaphone() {
		t.Error("ActivateMegaphone succeeded with 0 uses")
	}
	if comp.MegaphoneActive {
		t.Error("MegaphoneActive set with 0 uses")
	}
}

func TestChatComponentWalkieTalkie(t *testing.T) {
	comp := NewChatComponent()

	// Activate walkie-talkie
	comp.ActivateWalkieTalkie()
	if !comp.WalkieTalkieActive {
		t.Error("WalkieTalkieActive not set")
	}
	if comp.LocalRadius != -1.0 {
		t.Errorf("LocalRadius = %f, want -1.0 (unlimited)", comp.LocalRadius)
	}

	// Deactivate walkie-talkie
	comp.DeactivateWalkieTalkie()
	if comp.WalkieTalkieActive {
		t.Error("WalkieTalkieActive still set")
	}
	if comp.LocalRadius != 10.0 {
		t.Errorf("LocalRadius = %f, want 10.0 (default)", comp.LocalRadius)
	}
}

func TestChatComponentMegaphoneAndWalkieTalkie(t *testing.T) {
	comp := NewChatComponent()
	comp.MegaphoneUses = 1

	// Activate megaphone
	comp.ActivateMegaphone()
	if comp.LocalRadius != 30.0 {
		t.Errorf("LocalRadius = %f, want 30.0", comp.LocalRadius)
	}

	// Activate walkie-talkie (overrides megaphone)
	comp.ActivateWalkieTalkie()
	if comp.LocalRadius != -1.0 {
		t.Errorf("LocalRadius = %f, want -1.0 (walkie-talkie)", comp.LocalRadius)
	}

	// Deactivate walkie-talkie (returns to megaphone)
	comp.DeactivateWalkieTalkie()
	if comp.LocalRadius != 30.0 {
		t.Errorf("LocalRadius = %f, want 30.0 (megaphone still active)", comp.LocalRadius)
	}

	// Deactivate megaphone (returns to default)
	comp.DeactivateMegaphone()
	if comp.LocalRadius != 10.0 {
		t.Errorf("LocalRadius = %f, want 10.0 (default)", comp.LocalRadius)
	}
}

func TestChatComponentGetEffectiveRadius(t *testing.T) {
	comp := NewChatComponent()

	// Default
	if comp.GetEffectiveRadius() != 10.0 {
		t.Errorf("EffectiveRadius = %f, want 10.0", comp.GetEffectiveRadius())
	}

	// Megaphone
	comp.MegaphoneUses = 1
	comp.ActivateMegaphone()
	if comp.GetEffectiveRadius() != 30.0 {
		t.Errorf("EffectiveRadius = %f, want 30.0", comp.GetEffectiveRadius())
	}

	// Walkie-talkie
	comp.ActivateWalkieTalkie()
	if comp.GetEffectiveRadius() != -1.0 {
		t.Errorf("EffectiveRadius = %f, want -1.0", comp.GetEffectiveRadius())
	}
}

// Benchmarks

func BenchmarkAddMessage(b *testing.B) {
	comp := NewChatComponent()
	msg := ChatMessage{
		ID:        "test",
		SenderID:  1,
		Channel:   ChatGlobal,
		Content:   "Benchmark message",
		Timestamp: time.Now(),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp.AddMessage(msg)
	}
}

func BenchmarkGetMessageByID(b *testing.B) {
	comp := NewChatComponent()
	for i := 0; i < 100; i++ {
		comp.AddMessage(ChatMessage{
			ID:        string(rune('a' + i)),
			Channel:   ChatGlobal,
			Timestamp: time.Now(),
		})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = comp.GetMessageByID("e")
	}
}

func BenchmarkCanSendMessage(b *testing.B) {
	comp := NewChatComponent()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = comp.CanSendMessage(ChatGlobal)
	}
}

// PartyComponent tests

func TestNewPartyComponent(t *testing.T) {
	partyID := "party-123"
	leaderID := uint64(100)

	party := NewPartyComponent(partyID, leaderID)
	if party == nil {
		t.Fatal("NewPartyComponent returned nil")
	}
	if party.Type() != "party" {
		t.Errorf("Type() = %v, want party", party.Type())
	}
	if party.PartyID != partyID {
		t.Errorf("PartyID = %v, want %v", party.PartyID, partyID)
	}
	if party.LeaderID != leaderID {
		t.Errorf("LeaderID = %v, want %v", party.LeaderID, leaderID)
	}
	if len(party.MemberIDs) != 1 {
		t.Errorf("MemberIDs length = %d, want 1", len(party.MemberIDs))
	}
	if party.MemberIDs[0] != leaderID {
		t.Errorf("MemberIDs[0] = %v, want %v", party.MemberIDs[0], leaderID)
	}
}

func TestPartyComponentAddMember(t *testing.T) {
	party := NewPartyComponent("party-123", 100)

	// Add a new member
	party.AddMember(101)
	if len(party.MemberIDs) != 2 {
		t.Errorf("MemberIDs length = %d, want 2", len(party.MemberIDs))
	}
	if !party.IsMember(101) {
		t.Error("Member 101 not found after AddMember")
	}

	// Add duplicate member (should be idempotent)
	party.AddMember(101)
	if len(party.MemberIDs) != 2 {
		t.Errorf("MemberIDs length = %d, want 2 (duplicate should not add)", len(party.MemberIDs))
	}
}

func TestPartyComponentRemoveMember(t *testing.T) {
	party := NewPartyComponent("party-123", 100)
	party.AddMember(101)
	party.AddMember(102)

	if len(party.MemberIDs) != 3 {
		t.Fatalf("Setup failed: MemberIDs length = %d, want 3", len(party.MemberIDs))
	}

	// Remove a member
	party.RemoveMember(101)
	if len(party.MemberIDs) != 2 {
		t.Errorf("MemberIDs length = %d, want 2", len(party.MemberIDs))
	}
	if party.IsMember(101) {
		t.Error("Member 101 still present after RemoveMember")
	}

	// Remove non-existent member (should be safe)
	party.RemoveMember(999)
	if len(party.MemberIDs) != 2 {
		t.Errorf("MemberIDs length = %d, want 2 (removing non-existent should not change)", len(party.MemberIDs))
	}
}

func TestPartyComponentIsMember(t *testing.T) {
	party := NewPartyComponent("party-123", 100)
	party.AddMember(101)

	tests := []struct {
		name     string
		entityID uint64
		want     bool
	}{
		{"leader_is_member", 100, true},
		{"added_member", 101, true},
		{"not_member", 999, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := party.IsMember(tt.entityID)
			if got != tt.want {
				t.Errorf("IsMember(%v) = %v, want %v", tt.entityID, got, tt.want)
			}
		})
	}
}

func TestPartyComponentIsLeader(t *testing.T) {
	party := NewPartyComponent("party-123", 100)
	party.AddMember(101)

	tests := []struct {
		name     string
		entityID uint64
		want     bool
	}{
		{"leader", 100, true},
		{"member_not_leader", 101, false},
		{"not_member", 999, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := party.IsLeader(tt.entityID)
			if got != tt.want {
				t.Errorf("IsLeader(%v) = %v, want %v", tt.entityID, got, tt.want)
			}
		})
	}
}

func TestPartyComponentMultipleMemberOperations(t *testing.T) {
	party := NewPartyComponent("party-123", 100)

	// Add multiple members
	memberIDs := []uint64{101, 102, 103, 104, 105}
	for _, id := range memberIDs {
		party.AddMember(id)
	}

	// Verify all members present
	expectedCount := len(memberIDs) + 1 // +1 for leader
	if len(party.MemberIDs) != expectedCount {
		t.Errorf("MemberIDs length = %d, want %d", len(party.MemberIDs), expectedCount)
	}

	for _, id := range memberIDs {
		if !party.IsMember(id) {
			t.Errorf("Member %d not found", id)
		}
	}

	// Remove some members
	party.RemoveMember(102)
	party.RemoveMember(104)

	if len(party.MemberIDs) != 4 {
		t.Errorf("MemberIDs length = %d, want 4", len(party.MemberIDs))
	}
	if party.IsMember(102) {
		t.Error("Member 102 still present")
	}
	if party.IsMember(104) {
		t.Error("Member 104 still present")
	}
	if !party.IsMember(101) {
		t.Error("Member 101 should still be present")
	}
}
