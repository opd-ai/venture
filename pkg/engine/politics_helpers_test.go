package engine

import (
	"testing"
	"time"
)

// Tests for ServerFaction helpers

func TestServerFaction_Helpers(t *testing.T) {
	alignment := Alignment{LawAxis: 0.5, GoodAxis: 0.5}
	sf := NewServerFaction("server1", "TestFaction", alignment)

	// Test NewServerFaction
	if sf.ServerID != "server1" {
		t.Errorf("ServerID = %v, want server1", sf.ServerID)
	}
	if sf.FactionName != "TestFaction" {
		t.Errorf("FactionName = %v, want TestFaction", sf.FactionName)
	}
	if sf.Alignment != alignment {
		t.Errorf("Alignment = %v, want %v", sf.Alignment, alignment)
	}
	if len(sf.AllyServers) != 0 {
		t.Errorf("AllyServers should be empty, got %d", len(sf.AllyServers))
	}
	if len(sf.EnemyServers) != 0 {
		t.Errorf("EnemyServers should be empty, got %d", len(sf.EnemyServers))
	}
	if sf.Reputation == nil {
		t.Error("Reputation map should be initialized")
	}
}

func TestServerFaction_IsAlly(t *testing.T) {
	sf := NewServerFaction("server1", "TestFaction", Alignment{})
	sf.AllyServers = []string{"server2", "server3"}

	tests := []struct {
		name     string
		serverID string
		want     bool
	}{
		{"Ally present", "server2", true},
		{"Another ally", "server3", true},
		{"Not an ally", "server4", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sf.IsAlly(tt.serverID); got != tt.want {
				t.Errorf("IsAlly(%v) = %v, want %v", tt.serverID, got, tt.want)
			}
		})
	}
}

func TestServerFaction_IsEnemy(t *testing.T) {
	sf := NewServerFaction("server1", "TestFaction", Alignment{})
	sf.EnemyServers = []string{"server4", "server5"}

	tests := []struct {
		name     string
		serverID string
		want     bool
	}{
		{"Enemy present", "server4", true},
		{"Another enemy", "server5", true},
		{"Not an enemy", "server2", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sf.IsEnemy(tt.serverID); got != tt.want {
				t.Errorf("IsEnemy(%v) = %v, want %v", tt.serverID, got, tt.want)
			}
		})
	}
}

func TestServerFaction_AddAlly(t *testing.T) {
	sf := NewServerFaction("server1", "TestFaction", Alignment{})

	// Add new ally
	sf.AddAlly("server2")
	if !sf.IsAlly("server2") {
		t.Error("server2 should be an ally after AddAlly")
	}

	// Add duplicate ally
	sf.AddAlly("server2")
	count := 0
	for _, ally := range sf.AllyServers {
		if ally == "server2" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("server2 should appear once, appears %d times", count)
	}

	// Add self as ally (should be ignored)
	sf.AddAlly("server1")
	if sf.IsAlly("server1") {
		t.Error("server cannot be its own ally")
	}

	// Add enemy as ally (should move from enemy to ally)
	sf.EnemyServers = []string{"server3"}
	sf.AddAlly("server3")
	if sf.IsEnemy("server3") {
		t.Error("server3 should no longer be an enemy")
	}
	if !sf.IsAlly("server3") {
		t.Error("server3 should be an ally")
	}
}

func TestServerFaction_AddEnemy(t *testing.T) {
	sf := NewServerFaction("server1", "TestFaction", Alignment{})

	// Add new enemy
	sf.AddEnemy("server4")
	if !sf.IsEnemy("server4") {
		t.Error("server4 should be an enemy after AddEnemy")
	}

	// Add duplicate enemy
	sf.AddEnemy("server4")
	count := 0
	for _, enemy := range sf.EnemyServers {
		if enemy == "server4" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("server4 should appear once, appears %d times", count)
	}

	// Add self as enemy (should be ignored)
	sf.AddEnemy("server1")
	if sf.IsEnemy("server1") {
		t.Error("server cannot be its own enemy")
	}

	// Add ally as enemy (should move from ally to enemy)
	sf.AllyServers = []string{"server5"}
	sf.AddEnemy("server5")
	if sf.IsAlly("server5") {
		t.Error("server5 should no longer be an ally")
	}
	if !sf.IsEnemy("server5") {
		t.Error("server5 should be an enemy")
	}
}

func TestServerFaction_RemoveAlly(t *testing.T) {
	sf := NewServerFaction("server1", "TestFaction", Alignment{})
	sf.AllyServers = []string{"server2", "server3", "server4"}

	sf.RemoveAlly("server3")
	if sf.IsAlly("server3") {
		t.Error("server3 should not be an ally after removal")
	}
	if len(sf.AllyServers) != 2 {
		t.Errorf("AllyServers length = %d, want 2", len(sf.AllyServers))
	}

	// Remove non-existent ally
	sf.RemoveAlly("server99")
	if len(sf.AllyServers) != 2 {
		t.Errorf("Removing non-existent ally changed length to %d", len(sf.AllyServers))
	}
}

func TestServerFaction_RemoveEnemy(t *testing.T) {
	sf := NewServerFaction("server1", "TestFaction", Alignment{})
	sf.EnemyServers = []string{"server5", "server6", "server7"}

	sf.RemoveEnemy("server6")
	if sf.IsEnemy("server6") {
		t.Error("server6 should not be an enemy after removal")
	}
	if len(sf.EnemyServers) != 2 {
		t.Errorf("EnemyServers length = %d, want 2", len(sf.EnemyServers))
	}

	// Remove non-existent enemy
	sf.RemoveEnemy("server99")
	if len(sf.EnemyServers) != 2 {
		t.Errorf("Removing non-existent enemy changed length to %d", len(sf.EnemyServers))
	}
}

func TestServerFaction_GetReputation(t *testing.T) {
	sf := NewServerFaction("server1", "TestFaction", Alignment{})
	sf.Reputation["player1"] = 50.0
	sf.Reputation["player2"] = -30.0

	tests := []struct {
		name     string
		playerID string
		want     float64
	}{
		{"Positive reputation", "player1", 50.0},
		{"Negative reputation", "player2", -30.0},
		{"No reputation", "player3", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sf.GetReputation(tt.playerID); got != tt.want {
				t.Errorf("GetReputation(%v) = %v, want %v", tt.playerID, got, tt.want)
			}
		})
	}
}

func TestServerFaction_ModifyReputation(t *testing.T) {
	tests := []struct {
		name    string
		initial float64
		delta   float64
		want    float64
	}{
		{"Increase from zero", 0.0, 25.0, 25.0},
		{"Decrease from zero", 0.0, -40.0, -40.0},
		{"Clamp at max", 90.0, 20.0, 100.0},
		{"Clamp at min", -90.0, -20.0, -100.0},
		{"Already at max", 100.0, 10.0, 100.0},
		{"Already at min", -100.0, -10.0, -100.0},
		{"Normal increase", 30.0, 15.0, 45.0},
		{"Normal decrease", 30.0, -45.0, -15.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sf := NewServerFaction("server1", "TestFaction", Alignment{})
			sf.Reputation["player1"] = tt.initial
			sf.ModifyReputation("player1", tt.delta)
			if got := sf.GetReputation("player1"); got != tt.want {
				t.Errorf("ModifyReputation(%v, %v) resulted in %v, want %v", tt.initial, tt.delta, got, tt.want)
			}
		})
	}
}

// Tests for PoliticalEvent helpers

func TestNewPoliticalEvent(t *testing.T) {
	partyServers := []string{"server1", "server2"}
	duration := int64(86400) // 1 day
	event := NewPoliticalEvent(EventTypeAlliance, partyServers, duration)

	if event.Type != EventTypeAlliance {
		t.Errorf("Type = %v, want %v", event.Type, EventTypeAlliance)
	}
	if len(event.PartyServers) != 2 {
		t.Errorf("PartyServers length = %d, want 2", len(event.PartyServers))
	}
	if event.Duration != duration {
		t.Errorf("Duration = %v, want %v", event.Duration, duration)
	}
	if event.Effects == nil {
		t.Error("Effects map should be initialized")
	}
}

func TestPoliticalEvent_IsActive(t *testing.T) {
	// Active event (just started, 1 day duration)
	activeEvent := NewPoliticalEvent(EventTypeWar, []string{"s1", "s2"}, 86400)
	if !activeEvent.IsActive() {
		t.Error("Newly created event should be active")
	}

	// Expired event (started in the past, very short duration)
	expiredEvent := &PoliticalEvent{
		StartTime: time.Now().Unix() - 100,
		Duration:  1, // 1 second
	}
	if expiredEvent.IsActive() {
		t.Error("Expired event should not be active")
	}
}

func TestPoliticalEvent_Effects(t *testing.T) {
	event := NewPoliticalEvent(EventTypeTradePact, []string{"s1", "s2"}, 86400)

	// Set effect
	event.SetEffect("trade_bonus", 0.2)
	event.SetEffect("travel_cost", -0.1)

	// Get existing effect
	if val, exists := event.GetEffect("trade_bonus"); !exists || val.(float64) != 0.2 {
		t.Errorf("GetEffect(trade_bonus) = %v, %v, want 0.2, true", val, exists)
	}

	// Get non-existent effect
	if val, exists := event.GetEffect("nonexistent"); exists {
		t.Errorf("GetEffect(nonexistent) should not exist, got %v", val)
	}
}

// Test event type constants
func TestEventTypeConstants(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		want      string
	}{
		{"Alliance", EventTypeAlliance, "alliance"},
		{"War", EventTypeWar, "war"},
		{"Treaty", EventTypeTreaty, "treaty"},
		{"Embargo", EventTypeEmbargo, "embargo"},
		{"TradePact", EventTypeTradePact, "trade_pact"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.eventType != tt.want {
				t.Errorf("EventType constant = %v, want %v", tt.eventType, tt.want)
			}
		})
	}
}
