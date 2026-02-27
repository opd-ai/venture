package prestige

import (
	"testing"
	"time"
)

func TestManager_CreatePlayer(t *testing.T) {
	mgr := NewManager()

	playerID := "player1"
	className := "Warrior"
	accountID := "account1"

	mgr.CreatePlayer(playerID, className, accountID)

	// Verify player created
	mgr.mu.RLock()
	player, exists := mgr.players[playerID]
	mgr.mu.RUnlock()

	if !exists {
		t.Fatal("player was not created")
	}

	if player.PlayerID != playerID {
		t.Errorf("expected PlayerID %s, got %s", playerID, player.PlayerID)
	}

	if player.ClassName != className {
		t.Errorf("expected ClassName %s, got %s", className, player.ClassName)
	}

	if player.PrestigeLevel != 0 {
		t.Errorf("expected PrestigeLevel 0, got %d", player.PrestigeLevel)
	}

	// Verify account created
	mgr.mu.RLock()
	account, exists := mgr.accounts[accountID]
	mgr.mu.RUnlock()

	if !exists {
		t.Fatal("account was not created")
	}

	if len(account.CharacterIDs) != 1 || account.CharacterIDs[0] != playerID {
		t.Errorf("expected CharacterIDs [%s], got %v", playerID, account.CharacterIDs)
	}
}

func TestManager_AddPrestigeXP(t *testing.T) {
	tests := []struct {
		name          string
		initialLevel  int
		xpToAdd       int
		expectedLevel int
		expectedGain  int
	}{
		{
			name:          "first prestige level",
			initialLevel:  0,
			xpToAdd:       BasePrestigeXP, // 100,000
			expectedLevel: 1,
			expectedGain:  1,
		},
		{
			name:          "not enough for level",
			initialLevel:  0,
			xpToAdd:       50000,
			expectedLevel: 0,
			expectedGain:  0,
		},
		{
			name:          "multiple levels",
			initialLevel:  0,
			xpToAdd:       BasePrestigeXP * 3, // Level 1 (100k) + Level 2 (200k) = 300k total
			expectedLevel: 2,
			expectedGain:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewManager()
			playerID := "player1"
			className := "Warrior"
			accountID := "account1"

			mgr.CreatePlayer(playerID, className, accountID)

			// Set initial level if needed
			if tt.initialLevel > 0 {
				mgr.mu.Lock()
				mgr.players[playerID].PrestigeLevel = tt.initialLevel
				mgr.mu.Unlock()
			}

			levelsGained := mgr.AddPrestigeXP(playerID, className, tt.xpToAdd)

			if levelsGained != tt.expectedGain {
				t.Errorf("expected %d levels gained, got %d", tt.expectedGain, levelsGained)
			}

			finalLevel := mgr.GetPrestigeLevel(playerID)
			if finalLevel != tt.expectedLevel {
				t.Errorf("expected final level %d, got %d", tt.expectedLevel, finalLevel)
			}
		})
	}
}

func TestManager_ParagonPoints(t *testing.T) {
	mgr := NewManager()
	playerID := "player1"
	className := "Warrior"
	accountID := "account1"

	mgr.CreatePlayer(playerID, className, accountID)

	// Add points
	mgr.AddParagonPoints(playerID, 5)

	mgr.mu.RLock()
	points := mgr.players[playerID].ParagonPoints
	mgr.mu.RUnlock()

	if points != 5 {
		t.Errorf("expected 5 paragon points, got %d", points)
	}

	// Allocate to health
	err := mgr.AllocateParagonPoint(playerID, StatHealth)
	if err != nil {
		t.Fatalf("failed to allocate point: %v", err)
	}

	mgr.mu.RLock()
	pointsAfter := mgr.players[playerID].ParagonPoints
	healthPoints := mgr.players[playerID].ParagonAllocations[StatHealth]
	mgr.mu.RUnlock()

	if pointsAfter != 4 {
		t.Errorf("expected 4 remaining points, got %d", pointsAfter)
	}

	if healthPoints != 1 {
		t.Errorf("expected 1 health point, got %d", healthPoints)
	}

	// Check stat bonus
	bonus := mgr.GetStatBonus(playerID, StatHealth)
	expectedBonus := ParagonPointBonus // 0.001
	if bonus != expectedBonus {
		t.Errorf("expected bonus %f, got %f", expectedBonus, bonus)
	}
}

func TestManager_AllocateParagonPoint_NoPoints(t *testing.T) {
	mgr := NewManager()
	playerID := "player1"
	className := "Warrior"
	accountID := "account1"

	mgr.CreatePlayer(playerID, className, accountID)

	// Try to allocate without points
	err := mgr.AllocateParagonPoint(playerID, StatHealth)
	if err == nil {
		t.Error("expected error when allocating without points, got nil")
	}
}

func TestManager_AllocateParagonPoint_InvalidStat(t *testing.T) {
	mgr := NewManager()
	playerID := "player1"
	className := "Warrior"
	accountID := "account1"

	mgr.CreatePlayer(playerID, className, accountID)
	mgr.AddParagonPoints(playerID, 1)

	// Try to allocate to invalid stat
	err := mgr.AllocateParagonPoint(playerID, ParagonStat(999))
	if err == nil {
		t.Error("expected error for invalid stat, got nil")
	}
}

func TestManager_RespecParagonPoints(t *testing.T) {
	mgr := NewManager()
	playerID := "player1"
	className := "Warrior"
	accountID := "account1"

	mgr.CreatePlayer(playerID, className, accountID)
	mgr.AddParagonPoints(playerID, 10)

	// Allocate some points
	mgr.AllocateParagonPoint(playerID, StatHealth)
	mgr.AllocateParagonPoint(playerID, StatHealth)
	mgr.AllocateParagonPoint(playerID, StatDamage)

	mgr.mu.RLock()
	pointsBefore := mgr.players[playerID].ParagonPoints
	mgr.mu.RUnlock()

	if pointsBefore != 7 { // 10 - 3 allocated
		t.Errorf("expected 7 points before respec, got %d", pointsBefore)
	}

	// Respec
	cost, err := mgr.RespecParagonPoints(playerID)
	if err != nil {
		t.Fatalf("respec failed: %v", err)
	}

	expectedCost := 3 * RespecCostPerPoint // 3 points × 1000g
	if cost != expectedCost {
		t.Errorf("expected respec cost %d, got %d", expectedCost, cost)
	}

	mgr.mu.RLock()
	pointsAfter := mgr.players[playerID].ParagonPoints
	healthAlloc := mgr.players[playerID].ParagonAllocations[StatHealth]
	damageAlloc := mgr.players[playerID].ParagonAllocations[StatDamage]
	mgr.mu.RUnlock()

	if pointsAfter != 10 { // All points returned
		t.Errorf("expected 10 points after respec, got %d", pointsAfter)
	}

	if healthAlloc != 0 {
		t.Errorf("expected 0 health allocation, got %d", healthAlloc)
	}

	if damageAlloc != 0 {
		t.Errorf("expected 0 damage allocation, got %d", damageAlloc)
	}
}

func TestManager_GetVisualTier(t *testing.T) {
	mgr := NewManager()

	tests := []struct {
		level        int
		expectedTier VisualTier
		expectedName string
	}{
		{5, VisualNone, "None"},
		{10, VisualSubtle, "Subtle"},
		{24, VisualSubtle, "Subtle"},
		{25, VisualModerate, "Moderate"},
		{49, VisualModerate, "Moderate"},
		{50, VisualIntense, "Intense"},
		{99, VisualIntense, "Intense"},
		{100, VisualRadiant, "Radiant"},
		{200, VisualRadiant, "Radiant"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedName, func(t *testing.T) {
			tier := mgr.GetVisualTier(tt.level)
			if tier != tt.expectedTier {
				t.Errorf("level %d: expected tier %v, got %v", tt.level, tt.expectedTier, tier)
			}
			if tier.String() != tt.expectedName {
				t.Errorf("level %d: expected name %s, got %s", tt.level, tt.expectedName, tier.String())
			}
		})
	}
}

func TestManager_GetPrestigeAbility(t *testing.T) {
	mgr := NewManager()
	className := "Warrior"

	milestones := []int{10, 25, 50, 100}

	for _, milestone := range milestones {
		t.Run("milestone_"+string(rune(milestone+'0')), func(t *testing.T) {
			ability := mgr.GetPrestigeAbility(className, milestone)
			if ability == nil {
				t.Fatalf("expected ability at milestone %d, got nil", milestone)
			}

			if ability.ClassName != className {
				t.Errorf("expected ClassName %s, got %s", className, ability.ClassName)
			}

			if ability.UnlockLevel != milestone {
				t.Errorf("expected UnlockLevel %d, got %d", milestone, ability.UnlockLevel)
			}

			if ability.Name == "" {
				t.Error("ability Name should not be empty")
			}

			if ability.Description == "" {
				t.Error("ability Description should not be empty")
			}
		})
	}

	// Test invalid milestone
	ability := mgr.GetPrestigeAbility(className, 15)
	if ability != nil {
		t.Error("expected nil for invalid milestone 15")
	}
}

func TestManager_CheckAbilityUnlock(t *testing.T) {
	mgr := NewManager()
	playerID := "player1"
	className := "Warrior"
	accountID := "account1"

	mgr.CreatePlayer(playerID, className, accountID)

	// Set player to prestige 10
	mgr.mu.Lock()
	mgr.players[playerID].PrestigeLevel = 10
	mgr.mu.Unlock()

	// Check unlock
	ability := mgr.CheckAbilityUnlock(playerID)
	if ability == nil {
		t.Fatal("expected ability unlock at prestige 10, got nil")
	}

	if ability.UnlockLevel != 10 {
		t.Errorf("expected UnlockLevel 10, got %d", ability.UnlockLevel)
	}

	// Check that it doesn't unlock again
	ability2 := mgr.CheckAbilityUnlock(playerID)
	if ability2 != nil {
		t.Error("ability should not unlock twice")
	}

	// Verify unlocked abilities list
	mgr.mu.RLock()
	unlockedCount := len(mgr.players[playerID].UnlockedAbilities)
	mgr.mu.RUnlock()

	if unlockedCount != 1 {
		t.Errorf("expected 1 unlocked ability, got %d", unlockedCount)
	}
}

func TestManager_AccountXPBonus(t *testing.T) {
	mgr := NewManager()
	accountID := "account1"

	// Create first character
	mgr.CreatePlayer("player1", "Warrior", accountID)

	// Set to prestige 100
	mgr.AddPrestigeXP("player1", "Warrior", BasePrestigeXP*1000) // Large XP to hit 100
	mgr.mu.Lock()
	mgr.players["player1"].PrestigeLevel = 100
	mgr.mu.Unlock()
	mgr.updateAccountBonus("player1", 1)

	// Check bonus
	bonus := mgr.GetAccountXPBonus(accountID)
	expectedBonus := AccountXPBonus // 0.05 for 1 character

	if bonus < expectedBonus-0.0001 || bonus > expectedBonus+0.0001 {
		t.Errorf("expected bonus %f, got %f", expectedBonus, bonus)
	}

	// Create second character at prestige 100
	mgr.CreatePlayer("player2", "Mage", accountID)
	mgr.mu.Lock()
	mgr.players["player2"].PrestigeLevel = 100
	mgr.mu.Unlock()
	mgr.updateAccountBonus("player2", 1)

	// Check stacking bonus
	bonus2 := mgr.GetAccountXPBonus(accountID)
	// (1 + 0.05)^2 - 1 = 1.1025 - 1 = 0.1025
	expectedBonus2 := 0.1025

	if bonus2 < expectedBonus2-0.0001 || bonus2 > expectedBonus2+0.0001 {
		t.Errorf("expected bonus ~%f, got %f", expectedBonus2, bonus2)
	}
}

func TestManager_SaveLoad(t *testing.T) {
	mgr := NewManager()

	// Create some data
	mgr.CreatePlayer("player1", "Warrior", "account1")
	mgr.AddParagonPoints("player1", 10)
	mgr.AllocateParagonPoint("player1", StatHealth)
	mgr.AllocateParagonPoint("player1", StatDamage)

	mgr.CreatePlayer("player2", "Mage", "account1")
	mgr.AddPrestigeXP("player2", "Mage", BasePrestigeXP*2)

	// Save
	data, err := mgr.Save()
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Save returned empty data")
	}

	// Create new manager and load
	mgr2 := NewManager()
	if err := mgr2.Load(data); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify data
	mgr2.mu.RLock()
	player1, exists := mgr2.players["player1"]
	mgr2.mu.RUnlock()

	if !exists {
		t.Fatal("player1 not found after load")
	}

	if player1.ParagonPoints != 8 { // 10 - 2 allocated
		t.Errorf("expected 8 paragon points, got %d", player1.ParagonPoints)
	}

	if player1.ParagonAllocations[StatHealth] != 1 {
		t.Errorf("expected 1 health allocation, got %d", player1.ParagonAllocations[StatHealth])
	}

	mgr2.mu.RLock()
	account, exists := mgr2.accounts["account1"]
	mgr2.mu.RUnlock()

	if !exists {
		t.Fatal("account1 not found after load")
	}

	if len(account.CharacterIDs) != 2 {
		t.Errorf("expected 2 characters, got %d", len(account.CharacterIDs))
	}
}

func TestParagonStat_String(t *testing.T) {
	tests := []struct {
		stat     ParagonStat
		expected string
	}{
		{StatHealth, "Health"},
		{StatDamage, "Damage"},
		{StatDefense, "Defense"},
		{StatSpeed, "Speed"},
		{StatCritical, "Critical"},
		{ParagonStat(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.stat.String()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestPrestigeComponent_Type(t *testing.T) {
	comp := PrestigeComponent{}
	if comp.Type() != "prestige" {
		t.Errorf("expected type 'prestige', got %s", comp.Type())
	}
}

func TestPrestigeComponent_SerializeDeserialize(t *testing.T) {
	tests := []struct {
		name string
		comp PrestigeComponent
	}{
		{
			name: "full component",
			comp: PrestigeComponent{
				PlayerID:        "player1",
				PrestigeLevel:   42,
				VisualTier:      VisualIntense,
				ActiveAbilities: []string{"power_strike", "divine_shield"},
			},
		},
		{
			name: "empty component",
			comp: PrestigeComponent{},
		},
		{
			name: "max prestige",
			comp: PrestigeComponent{
				PlayerID:        "player_max",
				PrestigeLevel:   100,
				VisualTier:      VisualRadiant,
				ActiveAbilities: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.comp.Serialize()
			if err != nil {
				t.Fatalf("Serialize failed: %v", err)
			}
			if len(data) == 0 {
				t.Fatal("Serialize returned empty data")
			}

			var restored PrestigeComponent
			if err := restored.Deserialize(data); err != nil {
				t.Fatalf("Deserialize failed: %v", err)
			}

			if restored.PlayerID != tt.comp.PlayerID {
				t.Errorf("PlayerID: got %s, want %s", restored.PlayerID, tt.comp.PlayerID)
			}
			if restored.PrestigeLevel != tt.comp.PrestigeLevel {
				t.Errorf("PrestigeLevel: got %d, want %d", restored.PrestigeLevel, tt.comp.PrestigeLevel)
			}
			if restored.VisualTier != tt.comp.VisualTier {
				t.Errorf("VisualTier: got %d, want %d", restored.VisualTier, tt.comp.VisualTier)
			}
		})
	}
}

func TestPrestigeComponent_DeserializeInvalid(t *testing.T) {
	var comp PrestigeComponent
	if err := comp.Deserialize([]byte("invalid json")); err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestPlayerPrestige_MarshalJSON(t *testing.T) {
	player := &PlayerPrestige{
		PlayerID:           "player1",
		ClassName:          "Warrior",
		PrestigeLevel:      10,
		CurrentXP:          5000,
		TotalXP:            105000,
		ParagonPoints:      5,
		ParagonAllocations: map[ParagonStat]int{StatHealth: 3},
		UnlockedAbilities:  []int{10},
		LastUpdated:        time.Now(),
	}

	data, err := player.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("MarshalJSON returned empty data")
	}

	// Unmarshal and verify
	var player2 PlayerPrestige
	if err := player2.UnmarshalJSON(data); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if player2.PlayerID != player.PlayerID {
		t.Errorf("expected PlayerID %s, got %s", player.PlayerID, player2.PlayerID)
	}

	if player2.PrestigeLevel != player.PrestigeLevel {
		t.Errorf("expected PrestigeLevel %d, got %d", player.PrestigeLevel, player2.PrestigeLevel)
	}
}

// Benchmarks

func BenchmarkManager_AddPrestigeXP(b *testing.B) {
	mgr := NewManager()
	mgr.CreatePlayer("player1", "Warrior", "account1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.AddPrestigeXP("player1", "Warrior", 1000)
	}
}

func BenchmarkManager_AllocateParagonPoint(b *testing.B) {
	mgr := NewManager()
	mgr.CreatePlayer("player1", "Warrior", "account1")
	mgr.AddParagonPoints("player1", 1000000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.AllocateParagonPoint("player1", StatHealth)
	}
}

func BenchmarkManager_GetStatBonus(b *testing.B) {
	mgr := NewManager()
	mgr.CreatePlayer("player1", "Warrior", "account1")
	mgr.AddParagonPoints("player1", 10)
	mgr.AllocateParagonPoint("player1", StatHealth)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.GetStatBonus("player1", StatHealth)
	}
}

func BenchmarkManager_CheckAbilityUnlock(b *testing.B) {
	mgr := NewManager()
	mgr.CreatePlayer("player1", "Warrior", "account1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.mu.Lock()
		mgr.players["player1"].PrestigeLevel = 10 + (i % 4)
		mgr.players["player1"].UnlockedAbilities = []int{}
		mgr.mu.Unlock()
		mgr.CheckAbilityUnlock("player1")
	}
}

func BenchmarkManager_GetAccountXPBonus(b *testing.B) {
	mgr := NewManager()
	mgr.CreatePlayer("player1", "Warrior", "account1")
	mgr.CreatePlayer("player2", "Mage", "account1")
	
	// Simulate 3 prestige 100 characters for account bonus
	mgr.mu.Lock()
	mgr.accounts["account1"].Prestige100Count = 3
	mgr.accounts["account1"].XPBonus = 0.157625 // (1.05^3 - 1)
	mgr.mu.Unlock()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.GetAccountXPBonus("account1")
	}
}
