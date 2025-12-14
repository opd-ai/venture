package engine

import (
	"testing"
	"time"
)

func TestMatchmakingComponent_Type(t *testing.T) {
	c := NewMatchmakingComponent("server-1")
	if c.Type() != "matchmaking" {
		t.Errorf("Type() = %q, want %q", c.Type(), "matchmaking")
	}
}

func TestNewMatchmakingComponent(t *testing.T) {
	serverID := "server-1"
	c := NewMatchmakingComponent(serverID)

	if c.State != MatchmakingStateIdle {
		t.Errorf("State = %q, want %q", c.State, MatchmakingStateIdle)
	}
	if c.CurrentMode != MatchmakingMode1v1 {
		t.Errorf("CurrentMode = %q, want %q", c.CurrentMode, MatchmakingMode1v1)
	}
	if c.ServerID != serverID {
		t.Errorf("ServerID = %q, want %q", c.ServerID, serverID)
	}
	if len(c.Preferences.PreferredModes) != 1 {
		t.Errorf("PreferredModes length = %d, want 1", len(c.Preferences.PreferredModes))
	}
	if !c.Preferences.AcceptCrossServer {
		t.Error("AcceptCrossServer should be true by default")
	}
	if c.MaxHistorySize != 50 {
		t.Errorf("MaxHistorySize = %d, want 50", c.MaxHistorySize)
	}
}

func TestMatchmakingComponent_EnterQueue(t *testing.T) {
	c := NewMatchmakingComponent("server-1")

	// Should succeed when idle
	if !c.EnterQueue(MatchmakingMode1v1) {
		t.Error("EnterQueue() should succeed when idle")
	}

	if c.State != MatchmakingStateQueued {
		t.Errorf("State = %q, want %q", c.State, MatchmakingStateQueued)
	}
	if c.CurrentMode != MatchmakingMode1v1 {
		t.Errorf("CurrentMode = %q, want %q", c.CurrentMode, MatchmakingMode1v1)
	}
	if c.QueuedAt.IsZero() {
		t.Error("QueuedAt should be set")
	}

	// Should fail when already queued
	if c.EnterQueue(MatchmakingMode2v2) {
		t.Error("EnterQueue() should fail when already queued")
	}
}

func TestMatchmakingComponent_LeaveQueue(t *testing.T) {
	c := NewMatchmakingComponent("server-1")

	// Should fail when not queued
	if c.LeaveQueue() {
		t.Error("LeaveQueue() should fail when not queued")
	}

	// Enter queue first
	c.EnterQueue(MatchmakingMode1v1)
	time.Sleep(10 * time.Millisecond) // Small delay for queue time tracking

	// Should succeed when queued
	if !c.LeaveQueue() {
		t.Error("LeaveQueue() should succeed when queued")
	}

	if c.State != MatchmakingStateIdle {
		t.Errorf("State = %q, want %q", c.State, MatchmakingStateIdle)
	}
	if c.TotalQueueTime == 0 {
		t.Error("TotalQueueTime should be tracked")
	}
}

func TestMatchmakingComponent_AcceptMatch(t *testing.T) {
	c := NewMatchmakingComponent("server-1")

	// Should fail when not matched
	if c.AcceptMatch("match-1") {
		t.Error("AcceptMatch() should fail when not matched")
	}

	// Setup matched state
	c.EnterQueue(MatchmakingMode1v1)
	c.MarkMatched("match-1")

	// Should succeed when matched
	if !c.AcceptMatch("match-1") {
		t.Error("AcceptMatch() should succeed when matched")
	}

	if c.State != MatchmakingStateInMatch {
		t.Errorf("State = %q, want %q", c.State, MatchmakingStateInMatch)
	}
	if c.CurrentMatchID != "match-1" {
		t.Errorf("CurrentMatchID = %q, want %q", c.CurrentMatchID, "match-1")
	}
}

func TestMatchmakingComponent_DeclineMatch(t *testing.T) {
	c := NewMatchmakingComponent("server-1")

	// Should fail when not matched
	if c.DeclineMatch() {
		t.Error("DeclineMatch() should fail when not matched")
	}

	// Setup matched state
	c.EnterQueue(MatchmakingMode1v1)
	c.MarkMatched("match-1")

	// Should succeed when matched
	if !c.DeclineMatch() {
		t.Error("DeclineMatch() should succeed when matched")
	}

	if c.State != MatchmakingStateIdle {
		t.Errorf("State = %q, want %q", c.State, MatchmakingStateIdle)
	}
	if c.CurrentMatchID != "" {
		t.Error("CurrentMatchID should be cleared")
	}
}

func TestMatchmakingComponent_CompleteMatch(t *testing.T) {
	c := NewMatchmakingComponent("server-1")
	c.MaxHistorySize = 3 // Small size for testing

	// Setup in-match state
	c.EnterQueue(MatchmakingMode1v1)
	c.MarkMatched("match-1")
	c.AcceptMatch("match-1")

	result := MatchResult{
		MatchID:      "match-1",
		Mode:         MatchmakingMode1v1,
		Participants: []string{"player-1", "player-2"},
		WinnerIDs:    []string{"player-1"},
		LoserIDs:     []string{"player-2"},
		Duration:     5 * time.Minute,
		CompletedAt:  time.Now(),
		RatingChanges: map[string]int{
			"player-1": 16,
			"player-2": -16,
		},
	}

	c.CompleteMatch(result)

	if c.State != MatchmakingStateIdle {
		t.Errorf("State = %q, want %q", c.State, MatchmakingStateIdle)
	}
	if c.TotalMatches != 1 {
		t.Errorf("TotalMatches = %d, want 1", c.TotalMatches)
	}
	if len(c.MatchHistory) != 1 {
		t.Errorf("MatchHistory length = %d, want 1", len(c.MatchHistory))
	}

	// Test history size limit
	for i := 0; i < 5; i++ {
		result.MatchID = "match-" + string(rune('2'+i))
		c.State = MatchmakingStateInMatch
		c.CompleteMatch(result)
	}

	if len(c.MatchHistory) != c.MaxHistorySize {
		t.Errorf("MatchHistory length = %d, want %d", len(c.MatchHistory), c.MaxHistorySize)
	}
}

func TestMatchmakingComponent_MarkMatched(t *testing.T) {
	c := NewMatchmakingComponent("server-1")

	// Should fail when not queued
	if c.MarkMatched("match-1") {
		t.Error("MarkMatched() should fail when not queued")
	}

	// Enter queue first
	c.EnterQueue(MatchmakingMode1v1)

	// Should succeed when queued
	if !c.MarkMatched("match-1") {
		t.Error("MarkMatched() should succeed when queued")
	}

	if c.State != MatchmakingStateMatched {
		t.Errorf("State = %q, want %q", c.State, MatchmakingStateMatched)
	}
	if c.CurrentMatchID != "match-1" {
		t.Errorf("CurrentMatchID = %q, want %q", c.CurrentMatchID, "match-1")
	}
}

func TestMatchmakingComponent_IsQueued(t *testing.T) {
	c := NewMatchmakingComponent("server-1")

	if c.IsQueued() {
		t.Error("IsQueued() should be false when idle")
	}

	c.EnterQueue(MatchmakingMode1v1)

	if !c.IsQueued() {
		t.Error("IsQueued() should be true when queued")
	}
}

func TestMatchmakingComponent_IsInMatch(t *testing.T) {
	c := NewMatchmakingComponent("server-1")

	if c.IsInMatch() {
		t.Error("IsInMatch() should be false when idle")
	}

	c.EnterQueue(MatchmakingMode1v1)
	c.MarkMatched("match-1")
	c.AcceptMatch("match-1")

	if !c.IsInMatch() {
		t.Error("IsInMatch() should be true when in match")
	}
}

func TestMatchmakingComponent_GetQueueDuration(t *testing.T) {
	c := NewMatchmakingComponent("server-1")

	// Should be 0 when not queued
	if c.GetQueueDuration() != 0 {
		t.Error("GetQueueDuration() should be 0 when not queued")
	}

	c.EnterQueue(MatchmakingMode1v1)
	time.Sleep(20 * time.Millisecond)

	duration := c.GetQueueDuration()
	if duration < 20*time.Millisecond {
		t.Errorf("GetQueueDuration() = %v, want >= 20ms", duration)
	}
}

func TestMatchmakingComponent_GetAverageQueueTime(t *testing.T) {
	c := NewMatchmakingComponent("server-1")

	// Should be 0 with no matches
	if c.GetAverageQueueTime() != 0 {
		t.Error("GetAverageQueueTime() should be 0 with no matches")
	}

	// Simulate some queue time and matches
	c.TotalQueueTime = 30 * time.Second
	c.TotalMatches = 3

	avg := c.GetAverageQueueTime()
	if avg != 10*time.Second {
		t.Errorf("GetAverageQueueTime() = %v, want 10s", avg)
	}
}

func TestMatchmakingComponent_GetRecentMatches(t *testing.T) {
	c := NewMatchmakingComponent("server-1")

	// Empty history
	if c.GetRecentMatches(5) != nil {
		t.Error("GetRecentMatches() should return nil for empty history")
	}
	if c.GetRecentMatches(0) != nil {
		t.Error("GetRecentMatches(0) should return nil")
	}
	if c.GetRecentMatches(-1) != nil {
		t.Error("GetRecentMatches(-1) should return nil")
	}

	// Add some matches
	for i := 0; i < 10; i++ {
		c.MatchHistory = append(c.MatchHistory, MatchResult{
			MatchID: "match-" + string(rune('0'+i)),
		})
	}

	// Get recent 3
	recent := c.GetRecentMatches(3)
	if len(recent) != 3 {
		t.Errorf("GetRecentMatches(3) length = %d, want 3", len(recent))
	}
	if recent[0].MatchID != "match-7" {
		t.Errorf("First recent match = %q, want %q", recent[0].MatchID, "match-7")
	}

	// Get more than available
	recent = c.GetRecentMatches(20)
	if len(recent) != 10 {
		t.Errorf("GetRecentMatches(20) length = %d, want 10", len(recent))
	}
}

func TestMatchmakingComponent_SetPreferences(t *testing.T) {
	c := NewMatchmakingComponent("server-1")

	prefs := MatchmakingPreferences{
		PreferredModes:    []MatchmakingMode{MatchmakingMode1v1, MatchmakingMode2v2},
		AcceptCrossServer: false,
		MaxPingMs:         100,
	}

	c.SetPreferences(prefs)

	if len(c.Preferences.PreferredModes) != 2 {
		t.Errorf("PreferredModes length = %d, want 2", len(c.Preferences.PreferredModes))
	}
	if c.Preferences.AcceptCrossServer {
		t.Error("AcceptCrossServer should be false")
	}
	if c.Preferences.MaxPingMs != 100 {
		t.Errorf("MaxPingMs = %d, want 100", c.Preferences.MaxPingMs)
	}
}

func TestMatchmakingComponent_Team(t *testing.T) {
	c := NewMatchmakingComponent("server-1")

	if c.TeamID != "" {
		t.Error("TeamID should be empty initially")
	}

	c.SetTeam("team-1")
	if c.TeamID != "team-1" {
		t.Errorf("TeamID = %q, want %q", c.TeamID, "team-1")
	}

	c.ClearTeam()
	if c.TeamID != "" {
		t.Error("TeamID should be empty after ClearTeam()")
	}
}

func TestMatchmakingComponent_Serialize(t *testing.T) {
	c := NewMatchmakingComponent("server-1")
	c.TotalMatches = 25
	c.TotalQueueTime = 5 * time.Minute
	c.MatchHistory = append(c.MatchHistory, MatchResult{
		MatchID: "match-1",
		Mode:    MatchmakingMode1v1,
	})

	data, err := c.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	c2 := &MatchmakingComponent{}
	if err := c2.Deserialize(data); err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	if c2.ServerID != c.ServerID {
		t.Errorf("ServerID = %q, want %q", c2.ServerID, c.ServerID)
	}
	if c2.TotalMatches != c.TotalMatches {
		t.Errorf("TotalMatches = %d, want %d", c2.TotalMatches, c.TotalMatches)
	}
	if len(c2.MatchHistory) != len(c.MatchHistory) {
		t.Errorf("MatchHistory length = %d, want %d", len(c2.MatchHistory), len(c.MatchHistory))
	}
}

func TestMatchmakingComponent_Deserialize_Invalid(t *testing.T) {
	c := &MatchmakingComponent{}
	err := c.Deserialize([]byte("invalid json"))
	if err == nil {
		t.Error("Deserialize() expected error for invalid JSON")
	}
}

func TestMatchmakingComponent_GetWinCount(t *testing.T) {
	c := NewMatchmakingComponent("server-1")
	c.MatchHistory = []MatchResult{
		{MatchID: "m1", WinnerIDs: []string{"player-1"}, LoserIDs: []string{"player-2"}},
		{MatchID: "m2", WinnerIDs: []string{"player-2"}, LoserIDs: []string{"player-1"}},
		{MatchID: "m3", WinnerIDs: []string{"player-1"}, LoserIDs: []string{"player-3"}},
	}

	wins := c.GetWinCount("player-1")
	if wins != 2 {
		t.Errorf("GetWinCount(player-1) = %d, want 2", wins)
	}

	wins = c.GetWinCount("player-2")
	if wins != 1 {
		t.Errorf("GetWinCount(player-2) = %d, want 1", wins)
	}

	wins = c.GetWinCount("player-unknown")
	if wins != 0 {
		t.Errorf("GetWinCount(unknown) = %d, want 0", wins)
	}
}

func TestMatchmakingComponent_GetLossCount(t *testing.T) {
	c := NewMatchmakingComponent("server-1")
	c.MatchHistory = []MatchResult{
		{MatchID: "m1", WinnerIDs: []string{"player-1"}, LoserIDs: []string{"player-2"}},
		{MatchID: "m2", WinnerIDs: []string{"player-2"}, LoserIDs: []string{"player-1"}},
		{MatchID: "m3", WinnerIDs: []string{"player-1"}, LoserIDs: []string{"player-3"}},
	}

	losses := c.GetLossCount("player-1")
	if losses != 1 {
		t.Errorf("GetLossCount(player-1) = %d, want 1", losses)
	}

	losses = c.GetLossCount("player-2")
	if losses != 1 {
		t.Errorf("GetLossCount(player-2) = %d, want 1", losses)
	}

	losses = c.GetLossCount("player-3")
	if losses != 1 {
		t.Errorf("GetLossCount(player-3) = %d, want 1", losses)
	}
}

func TestMatchmakingModes(t *testing.T) {
	modes := []MatchmakingMode{
		MatchmakingMode1v1,
		MatchmakingMode2v2,
		MatchmakingModeFFA,
	}

	for _, mode := range modes {
		c := NewMatchmakingComponent("server-1")
		if !c.EnterQueue(mode) {
			t.Errorf("EnterQueue(%s) should succeed", mode)
		}
		if c.CurrentMode != mode {
			t.Errorf("CurrentMode = %q, want %q", c.CurrentMode, mode)
		}
	}
}

func TestMatchmakingStates(t *testing.T) {
	tests := []struct {
		name  string
		state MatchmakingState
	}{
		{"idle", MatchmakingStateIdle},
		{"queued", MatchmakingStateQueued},
		{"matched", MatchmakingStateMatched},
		{"in_match", MatchmakingStateInMatch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewMatchmakingComponent("server-1")
			c.State = tt.state
			if c.State != tt.state {
				t.Errorf("State = %q, want %q", c.State, tt.state)
			}
		})
	}
}
