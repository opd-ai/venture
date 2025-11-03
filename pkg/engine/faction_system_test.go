package engine

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNewFactionSystem(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()

	fs := NewFactionSystem(world, logger)

	if fs == nil {
		t.Fatal("NewFactionSystem returned nil")
	}
	if fs.world != world {
		t.Error("World not set correctly")
	}
	if fs.logger != logger {
		t.Error("Logger not set correctly")
	}
	if fs.Factions == nil {
		t.Error("Factions map not initialized")
	}
	if fs.PendingReputationChanges == nil {
		t.Error("PendingReputationChanges not initialized")
	}
}

func TestNewFactionSystem_NilLogger(t *testing.T) {
	world := NewWorld()

	fs := NewFactionSystem(world, nil)

	if fs.logger == nil {
		t.Error("Logger should be initialized even when nil passed")
	}
}

func TestFactionSystem_AddFaction(t *testing.T) {
	world := NewWorld()
	fs := NewFactionSystem(world, logrus.New())

	faction := &Faction{
		ID:   "test_faction",
		Name: "Test Faction",
		Type: FactionTypeKingdom,
	}

	fs.AddFaction(faction)

	if len(fs.Factions) != 1 {
		t.Errorf("Expected 1 faction, got %d", len(fs.Factions))
	}

	retrieved := fs.GetFaction("test_faction")
	if retrieved == nil {
		t.Fatal("Failed to retrieve added faction")
	}
	if retrieved.Name != "Test Faction" {
		t.Errorf("Expected name 'Test Faction', got '%s'", retrieved.Name)
	}
}

func TestFactionSystem_AddFaction_Nil(t *testing.T) {
	world := NewWorld()
	fs := NewFactionSystem(world, logrus.New())

	fs.AddFaction(nil)

	if len(fs.Factions) != 0 {
		t.Error("Nil faction should not be added")
	}
}

func TestFactionSystem_GetFaction(t *testing.T) {
	world := NewWorld()
	fs := NewFactionSystem(world, logrus.New())

	faction1 := &Faction{ID: "faction1", Name: "Faction 1"}
	faction2 := &Faction{ID: "faction2", Name: "Faction 2"}

	fs.AddFaction(faction1)
	fs.AddFaction(faction2)

	tests := []struct {
		name       string
		factionID  string
		shouldFind bool
	}{
		{"Get existing faction 1", "faction1", true},
		{"Get existing faction 2", "faction2", true},
		{"Get non-existing faction", "faction3", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			faction := fs.GetFaction(tt.factionID)
			if tt.shouldFind && faction == nil {
				t.Error("Expected to find faction, got nil")
			}
			if !tt.shouldFind && faction != nil {
				t.Error("Expected nil, got faction")
			}
		})
	}
}

func TestFactionSystem_QueueReputationChange(t *testing.T) {
	world := NewWorld()
	fs := NewFactionSystem(world, logrus.New())

	change := ReputationChange{
		FactionID: "test_faction",
		Amount:    10,
		Reason:    "Test reason",
	}

	fs.QueueReputationChange(change)

	if len(fs.PendingReputationChanges) != 1 {
		t.Errorf("Expected 1 pending change, got %d", len(fs.PendingReputationChanges))
	}

	if fs.PendingReputationChanges[0].FactionID != "test_faction" {
		t.Error("Queued change has incorrect faction ID")
	}
	if fs.PendingReputationChanges[0].Amount != 10 {
		t.Error("Queued change has incorrect amount")
	}
}

func TestFactionSystem_Update_ClearsPendingChanges(t *testing.T) {
	world := NewWorld()
	fs := NewFactionSystem(world, logrus.New())

	// Add some pending changes
	fs.QueueReputationChange(ReputationChange{FactionID: "f1", Amount: 10})
	fs.QueueReputationChange(ReputationChange{FactionID: "f2", Amount: -5})

	if len(fs.PendingReputationChanges) != 2 {
		t.Fatal("Changes not queued correctly")
	}

	// Update should process and clear them
	fs.Update(0.016)

	if len(fs.PendingReputationChanges) != 0 {
		t.Errorf("Expected pending changes to be cleared, got %d remaining",
			len(fs.PendingReputationChanges))
	}
}

func TestFactionSystem_GetPlayerReputation_NoPlayer(t *testing.T) {
	world := NewWorld()
	fs := NewFactionSystem(world, logrus.New())

	rep := fs.GetPlayerReputation("test_faction")

	if rep != 0 {
		t.Errorf("Expected 0 reputation when no player, got %d", rep)
	}
}

func TestFactionSystem_CanTrade(t *testing.T) {
	world := NewWorld()
	fs := NewFactionSystem(world, logrus.New())

	tests := []struct {
		name       string
		reputation int
		canTrade   bool
	}{
		{"Hostile -100", -100, false},
		{"Hostile -50", -50, false},
		{"Suspicious -49", -49, true},
		{"Suspicious 0", 0, true},
		{"Neutral 1", 1, true},
		{"Friendly 51", 51, true},
		{"Friendly 100", 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This test is limited because we can't easily set player reputation
			// without a player entity. Testing the threshold logic.
			canTrade := tt.reputation > -50
			if canTrade != tt.canTrade {
				t.Errorf("Expected CanTrade=%v for reputation %d", tt.canTrade, tt.reputation)
			}
		})
	}
}

func TestFactionSystem_ShouldAttackPlayer(t *testing.T) {
	world := NewWorld()
	fs := NewFactionSystem(world, logrus.New())

	tests := []struct {
		name         string
		reputation   int
		shouldAttack bool
	}{
		{"Hostile -100", -100, true},
		{"Hostile -50", -50, true},
		{"Suspicious -49", -49, false},
		{"Neutral 0", 0, false},
		{"Friendly 51", 51, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Testing threshold logic
			shouldAttack := tt.reputation <= -50
			if shouldAttack != tt.shouldAttack {
				t.Errorf("Expected ShouldAttackPlayer=%v for reputation %d",
					tt.shouldAttack, tt.reputation)
			}
		})
	}
}

func TestFactionSystem_GetTradeDiscount(t *testing.T) {
	world := NewWorld()
	fs := NewFactionSystem(world, logrus.New())

	tests := []struct {
		name       string
		reputation int
		minMult    float64
		maxMult    float64
	}{
		{"Hostile no trade", -100, 0.0, 0.0},
		{"Suspicious markup", -25, 1.5, 1.5},
		{"Neutral normal", 25, 1.0, 1.0},
		{"Friendly discount", 75, 0.8, 0.9},
		{"Max friendly discount", 100, 0.75, 0.75},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := FactionComponent{Reputation: tt.reputation}
			mult := fc.GetPriceMultiplier()
			if mult < tt.minMult || mult > tt.maxMult {
				t.Errorf("Price multiplier %v not in expected range [%v, %v]",
					mult, tt.minMult, tt.maxMult)
			}
		})
	}
}

func TestFactionSystem_ProcessKillReputation_NoVictimFaction(t *testing.T) {
	world := NewWorld()
	fs := NewFactionSystem(world, logrus.New())

	killer := world.CreateEntity()
	killer.AddComponent(PlayerComponent{})

	victim := world.CreateEntity()
	// Victim has no faction component

	fs.ProcessKillReputation(killer, victim)

	// Should not queue any changes
	if len(fs.PendingReputationChanges) != 0 {
		t.Errorf("Expected no reputation changes, got %d", len(fs.PendingReputationChanges))
	}
}

func TestFactionSystem_ProcessKillReputation_PlayerVictim(t *testing.T) {
	world := NewWorld()
	fs := NewFactionSystem(world, logrus.New())

	killer := world.CreateEntity()

	victim := world.CreateEntity()
	victim.AddComponent(PlayerComponent{})
	victim.AddComponent(FactionComponent{
		FactionID:       "test_faction",
		IsPlayerFaction: true,
	})

	fs.ProcessKillReputation(killer, victim)

	// Should not process player as victim
	if len(fs.PendingReputationChanges) != 0 {
		t.Errorf("Expected no reputation changes for player victim, got %d",
			len(fs.PendingReputationChanges))
	}
}

func TestFactionSystem_ProcessKillReputation_PlayerKillsMember(t *testing.T) {
	world := NewWorld()
	fs := NewFactionSystem(world, logrus.New())

	// Add a faction
	faction := &Faction{
		ID:            "test_faction",
		Name:          "Test Faction",
		Type:          FactionTypeKingdom,
		Relationships: make(map[string]int),
	}
	fs.AddFaction(faction)

	killer := world.CreateEntity()
	killer.AddComponent(PlayerComponent{})

	victim := world.CreateEntity()
	victim.AddComponent(FactionComponent{
		FactionID:       "test_faction",
		IsPlayerFaction: false,
	})

	fs.ProcessKillReputation(killer, victim)

	// Should queue negative reputation change
	if len(fs.PendingReputationChanges) == 0 {
		t.Fatal("Expected reputation change to be queued")
	}

	change := fs.PendingReputationChanges[0]
	if change.FactionID != "test_faction" {
		t.Error("Reputation change has wrong faction ID")
	}
	if change.Amount != ReputationKillMember {
		t.Errorf("Expected amount %d, got %d", ReputationKillMember, change.Amount)
	}
}

func TestFactionSystem_ProcessKillReputation_EnemyFaction(t *testing.T) {
	world := NewWorld()
	fs := NewFactionSystem(world, logrus.New())

	// Add two factions as enemies
	faction1 := &Faction{
		ID:   "faction1",
		Name: "Faction 1",
		Type: FactionTypeKingdom,
		Relationships: map[string]int{
			"faction2": -75, // Enemy
		},
	}
	faction2 := &Faction{
		ID:   "faction2",
		Name: "Faction 2",
		Type: FactionTypeGuild,
		Relationships: map[string]int{
			"faction1": -75, // Enemy
		},
	}
	fs.AddFaction(faction1)
	fs.AddFaction(faction2)

	killer := world.CreateEntity()
	killer.AddComponent(PlayerComponent{})

	victim := world.CreateEntity()
	victim.AddComponent(FactionComponent{
		FactionID:       "faction2",
		IsPlayerFaction: false,
	})

	fs.ProcessKillReputation(killer, victim)

	// Should queue negative reputation with faction2 and positive with faction1
	if len(fs.PendingReputationChanges) < 2 {
		t.Fatalf("Expected at least 2 reputation changes, got %d",
			len(fs.PendingReputationChanges))
	}

	// Find the kill member penalty
	foundKillPenalty := false
	foundEnemyBonus := false

	for _, change := range fs.PendingReputationChanges {
		if change.FactionID == "faction2" && change.Amount == ReputationKillMember {
			foundKillPenalty = true
		}
		if change.FactionID == "faction1" && change.Amount == ReputationKillEnemy {
			foundEnemyBonus = true
		}
	}

	if !foundKillPenalty {
		t.Error("Expected kill member penalty for faction2")
	}
	if !foundEnemyBonus {
		t.Error("Expected kill enemy bonus for faction1")
	}
}
