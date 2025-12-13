package prestige

import (
	"testing"
)

// mockEntity is a test entity implementation.
type mockEntity struct {
	id         string
	components map[string]interface{}
}

func newMockEntity(id string) *mockEntity {
	return &mockEntity{
		id:         id,
		components: make(map[string]interface{}),
	}
}

func (e *mockEntity) GetID() string {
	return e.id
}

func (e *mockEntity) HasComponent(componentType string) bool {
	_, exists := e.components[componentType]
	return exists
}

func (e *mockEntity) GetComponent(componentType string) interface{} {
	return e.components[componentType]
}

func (e *mockEntity) AddComponent(component interface{ Type() string }) {
	e.components[component.Type()] = component
}

func (e *mockEntity) RemoveComponent(componentType string) {
	delete(e.components, componentType)
}

func TestNewSystem(t *testing.T) {
	sys := NewSystem()
	if sys == nil {
		t.Fatal("NewSystem() returned nil")
	}
	if sys.manager == nil {
		t.Fatal("system manager is nil")
	}
	if sys.logger == nil {
		t.Fatal("system logger is nil")
	}
}

func TestSystem_InitializePlayer(t *testing.T) {
	sys := NewSystem()
	entity := newMockEntity("entity1")

	playerID := "player1"
	className := "Warrior"
	accountID := "account1"

	sys.InitializePlayer(entity, playerID, className, accountID)

	// Verify component added
	if !entity.HasComponent("prestige") {
		t.Fatal("prestige component not added to entity")
	}

	comp := entity.GetComponent("prestige")
	prestigeComp, ok := comp.(*PrestigeComponent)
	if !ok {
		t.Fatal("component is not a PrestigeComponent")
	}

	if prestigeComp.PlayerID != playerID {
		t.Errorf("expected PlayerID %s, got %s", playerID, prestigeComp.PlayerID)
	}

	if prestigeComp.PrestigeLevel != 0 {
		t.Errorf("expected PrestigeLevel 0, got %d", prestigeComp.PrestigeLevel)
	}

	if prestigeComp.VisualTier != VisualNone {
		t.Errorf("expected VisualTier None, got %v", prestigeComp.VisualTier)
	}

	// Verify manager has player
	level := sys.manager.GetPrestigeLevel(playerID)
	if level != 0 {
		t.Errorf("expected manager prestige level 0, got %d", level)
	}
}

func TestSystem_AwardPrestigeXP(t *testing.T) {
	tests := []struct {
		name          string
		xpToAward     int
		expectedGain  int
		expectedLevel int
	}{
		{
			name:          "no level gain",
			xpToAward:     50000,
			expectedGain:  0,
			expectedLevel: 0,
		},
		{
			name:          "one level gain",
			xpToAward:     BasePrestigeXP,
			expectedGain:  1,
			expectedLevel: 1,
		},
		{
			name:          "multiple levels",
			xpToAward:     BasePrestigeXP * 3,
			expectedGain:  2,
			expectedLevel: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewSystem()
			entity := newMockEntity("entity1")
			playerID := "player1"
			className := "Warrior"

			sys.InitializePlayer(entity, playerID, className, "account1")

			levelsGained := sys.AwardPrestigeXP(playerID, className, tt.xpToAward)

			if levelsGained != tt.expectedGain {
				t.Errorf("expected %d levels gained, got %d", tt.expectedGain, levelsGained)
			}

			currentLevel := sys.manager.GetPrestigeLevel(playerID)
			if currentLevel != tt.expectedLevel {
				t.Errorf("expected prestige level %d, got %d", tt.expectedLevel, currentLevel)
			}
		})
	}
}

func TestSystem_AllocateParagonPoint(t *testing.T) {
	sys := NewSystem()
	entity := newMockEntity("entity1")
	playerID := "player1"

	sys.InitializePlayer(entity, playerID, "Warrior", "account1")

	// Award XP to gain paragon point
	sys.AwardPrestigeXP(playerID, "Warrior", BasePrestigeXP)

	// Allocate point
	err := sys.AllocateParagonPoint(playerID, StatHealth)
	if err != nil {
		t.Fatalf("failed to allocate paragon point: %v", err)
	}

	// Verify allocation
	bonus := sys.GetStatBonus(playerID, StatHealth)
	expectedBonus := ParagonPointBonus
	if bonus != expectedBonus {
		t.Errorf("expected bonus %f, got %f", expectedBonus, bonus)
	}

	// Try allocating without points
	err = sys.AllocateParagonPoint(playerID, StatDamage)
	if err == nil {
		t.Fatal("expected error when allocating without points")
	}
}

func TestSystem_Update_AbilityUnlock(t *testing.T) {
	sys := NewSystem()
	entity := newMockEntity("entity1")
	playerID := "player1"
	className := "Warrior"

	sys.InitializePlayer(entity, playerID, className, "account1")

	// Award XP to reach prestige level 10 (first ability unlock)
	xpNeeded := 0
	for i := 1; i <= 10; i++ {
		xpNeeded += sys.manager.calculateXPRequired(i)
	}
	sys.AwardPrestigeXP(playerID, className, xpNeeded)

	// Run update to trigger ability unlock
	entities := []Entity{entity}
	sys.Update(entities, 0.016)

	// Verify ability unlocked
	comp := entity.GetComponent("prestige").(*PrestigeComponent)
	if len(comp.ActiveAbilities) == 0 {
		t.Fatal("expected ability to be unlocked")
	}

	expectedAbility := className + "'s Resolve"
	if comp.ActiveAbilities[0] != expectedAbility {
		t.Errorf("expected ability %s, got %s", expectedAbility, comp.ActiveAbilities[0])
	}

	// Verify visual tier updated
	if comp.VisualTier != VisualSubtle {
		t.Errorf("expected VisualSubtle tier at level 10, got %v", comp.VisualTier)
	}

	if comp.PrestigeLevel != 10 {
		t.Errorf("expected prestige level 10, got %d", comp.PrestigeLevel)
	}
}

func TestSystem_Update_VisualTier(t *testing.T) {
	tests := []struct {
		name         string
		level        int
		expectedTier VisualTier
	}{
		{"level 5", 5, VisualNone},
		{"level 10", 10, VisualSubtle},
		{"level 25", 25, VisualModerate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewSystem()
			entity := newMockEntity("entity1")
			playerID := "player1"

			sys.InitializePlayer(entity, playerID, "Warrior", "account1")

			// Directly set the prestige level to avoid overflow
			sys.manager.mu.Lock()
			if sys.manager.players[playerID] != nil {
				sys.manager.players[playerID].PrestigeLevel = tt.level
			}
			sys.manager.mu.Unlock()

			// Run update to update visual tier
			entities := []Entity{entity}
			sys.Update(entities, 0.016)

			comp := entity.GetComponent("prestige").(*PrestigeComponent)
			if comp.VisualTier != tt.expectedTier {
				t.Errorf("expected visual tier %v, got %v", tt.expectedTier, comp.VisualTier)
			}

			if comp.PrestigeLevel != tt.level {
				t.Errorf("expected prestige level %d, got %d", tt.level, comp.PrestigeLevel)
			}
		})
	}
}

func TestSystem_GetAccountXPBonus(t *testing.T) {
	sys := NewSystem()
	accountID := "account1"

	// Create first character and directly set to level 100
	entity1 := newMockEntity("entity1")
	sys.InitializePlayer(entity1, "player1", "Warrior", accountID)

	// Directly set to prestige 100 to avoid overflow
	sys.manager.mu.Lock()
	if sys.manager.players["player1"] != nil {
		sys.manager.players["player1"].PrestigeLevel = 100
	}
	sys.manager.mu.Unlock()

	// Trigger account bonus update
	sys.manager.updateAccountBonus("player1", 1)

	// Check account bonus
	bonus := sys.GetAccountXPBonus(accountID)
	expectedBonus := AccountXPBonus // 5% = 0.05
	delta := 0.0001
	if bonus < expectedBonus-delta || bonus > expectedBonus+delta {
		t.Errorf("expected account bonus %f, got %f", expectedBonus, bonus)
	}

	// Create second character and level to 100
	entity2 := newMockEntity("entity2")
	sys.InitializePlayer(entity2, "player2", "Mage", accountID)

	sys.manager.mu.Lock()
	if sys.manager.players["player2"] != nil {
		sys.manager.players["player2"].PrestigeLevel = 100
	}
	sys.manager.mu.Unlock()

	sys.manager.updateAccountBonus("player2", 1)

	// Check bonus stacking (should be multiplicative)
	bonus = sys.GetAccountXPBonus(accountID)
	// 2 chars at prestige 100: (1.05^2 - 1) = 0.1025
	expectedBonus = 0.1025
	if bonus < expectedBonus-delta || bonus > expectedBonus+delta {
		t.Errorf("expected account bonus ~%f, got %f", expectedBonus, bonus)
	}
}

func TestSystem_RespecParagonPoints(t *testing.T) {
	sys := NewSystem()
	entity := newMockEntity("entity1")
	playerID := "player1"

	sys.InitializePlayer(entity, playerID, "Warrior", "account1")

	// Award XP to gain 3 paragon points
	sys.AwardPrestigeXP(playerID, "Warrior", BasePrestigeXP*3)

	// Allocate points
	sys.AllocateParagonPoint(playerID, StatHealth)
	sys.AllocateParagonPoint(playerID, StatDamage)
	sys.AllocateParagonPoint(playerID, StatDefense)

	// Respec
	cost, err := sys.RespecParagonPoints(playerID)
	if err != nil {
		t.Fatalf("respec failed: %v", err)
	}

	expectedCost := 2 * RespecCostPerPoint // 2000 gold (3 points allocated from 3 levels)
	if cost != expectedCost {
		t.Errorf("expected respec cost %d, got %d", expectedCost, cost)
	}

	// Verify points returned
	bonus := sys.GetStatBonus(playerID, StatHealth)
	if bonus != 0.0 {
		t.Errorf("expected bonus 0 after respec, got %f", bonus)
	}
}

func TestSystem_SaveLoad(t *testing.T) {
	sys := NewSystem()
	playerID := "player1"
	className := "Warrior"
	accountID := "account1"

	entity := newMockEntity("entity1")
	sys.InitializePlayer(entity, playerID, className, accountID)

	// Award enough XP for 2 levels (level 1: 100k, level 2: 200k = 300k total)
	sys.AwardPrestigeXP(playerID, className, BasePrestigeXP+BasePrestigeXP*2)
	sys.AllocateParagonPoint(playerID, StatHealth)

	// Save
	data, err := sys.Save()
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// Create new system and load
	sys2 := NewSystem()
	err = sys2.Load(data)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	// Verify data restored
	level := sys2.manager.GetPrestigeLevel(playerID)
	if level != 2 {
		t.Errorf("expected prestige level 2 after load, got %d", level)
	}

	bonus := sys2.GetStatBonus(playerID, StatHealth)
	expectedBonus := ParagonPointBonus
	if bonus != expectedBonus {
		t.Errorf("expected stat bonus %f after load, got %f", expectedBonus, bonus)
	}
}
