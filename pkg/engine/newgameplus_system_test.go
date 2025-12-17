// Package engine provides tests for the New Game Plus system.
package engine

import (
	"testing"
)

func TestNewNewGamePlusSystem(t *testing.T) {
	world := NewWorld()
	system := NewNewGamePlusSystem(world)

	if system == nil {
		t.Fatal("NewNewGamePlusSystem returned nil")
	}
	if system.world != world {
		t.Error("System should store world reference")
	}
}

func TestNewNewGamePlusSystem_NilWorld(t *testing.T) {
	system := NewNewGamePlusSystem(nil)
	if system == nil {
		t.Fatal("NewNewGamePlusSystem should not return nil even with nil world")
	}
}

func TestNewGamePlusSystem_Update(t *testing.T) {
	world := NewWorld()
	system := NewNewGamePlusSystem(world)

	// Create entity with NG+ component
	entity := NewEntity(1)
	ngp := NewNewGamePlusComponent()
	entity.AddComponent(ngp)

	// Update should accumulate playtime
	entities := []*Entity{entity}
	system.Update(entities, 1.0) // 1 second

	if ngp.GetCurrentCyclePlaytime() != 1 {
		t.Errorf("CurrentCyclePlaytime = %d, want 1", ngp.GetCurrentCyclePlaytime())
	}
}

func TestNewGamePlusSystem_Update_NoComponent(t *testing.T) {
	world := NewWorld()
	system := NewNewGamePlusSystem(world)

	// Entity without NG+ component
	entity := NewEntity(1)
	entities := []*Entity{entity}

	// Should not panic
	system.Update(entities, 1.0)
}

func TestNewGamePlusSystem_InitiateNewGamePlus(t *testing.T) {
	world := NewWorld()
	system := NewNewGamePlusSystem(world)

	entity := NewEntity(1)
	ngp := NewNewGamePlusComponent()
	entity.AddComponent(ngp)

	stats := map[string]int64{
		"enemies_killed":   100,
		"quests_completed": 5,
	}

	err := system.InitiateNewGamePlus(entity, stats)
	if err != nil {
		t.Fatalf("InitiateNewGamePlus error: %v", err)
	}

	if ngp.GetCycle() != 1 {
		t.Errorf("Cycle = %d, want 1", ngp.GetCycle())
	}
}

func TestNewGamePlusSystem_InitiateNewGamePlus_CreatesComponent(t *testing.T) {
	world := NewWorld()
	system := NewNewGamePlusSystem(world)

	entity := NewEntity(1)
	// No NG+ component initially

	err := system.InitiateNewGamePlus(entity, map[string]int64{})
	if err != nil {
		t.Fatalf("InitiateNewGamePlus error: %v", err)
	}

	// Should have created component
	if !entity.HasComponent("newgameplus") {
		t.Error("Should have created newgameplus component")
	}
}

func TestNewGamePlusSystem_OnCycleStartCallback(t *testing.T) {
	world := NewWorld()
	system := NewNewGamePlusSystem(world)

	callbackCalled := false
	var callbackCycle int

	system.SetOnCycleStart(func(cycle int) {
		callbackCalled = true
		callbackCycle = cycle
	})

	entity := NewEntity(1)
	ngp := NewNewGamePlusComponent()
	entity.AddComponent(ngp)

	system.InitiateNewGamePlus(entity, map[string]int64{})

	if !callbackCalled {
		t.Error("OnCycleStart callback should have been called")
	}
	if callbackCycle != 1 {
		t.Errorf("Callback cycle = %d, want 1", callbackCycle)
	}
}

func TestNewGamePlusSystem_OnBonusUnlockCallback(t *testing.T) {
	world := NewWorld()
	system := NewNewGamePlusSystem(world)

	var unlockedBonus string
	system.SetOnBonusUnlock(func(bonusID string) {
		unlockedBonus = bonusID
	})

	entity := NewEntity(1)
	ngp := NewNewGamePlusComponent()
	ngp.Cycle = 1 // NG+1
	entity.AddComponent(ngp)

	// Update should trigger ng_veteran unlock
	system.Update([]*Entity{entity}, 0.0)

	if unlockedBonus != "ng_veteran" {
		t.Errorf("Expected ng_veteran bonus unlock, got %q", unlockedBonus)
	}
}

func TestNewGamePlusSystem_MilestoneUnlocks(t *testing.T) {
	tests := []struct {
		name       string
		setupNGP   func(*NewGamePlusComponent)
		wantBonus  string
		shouldHave bool
	}{
		{
			name: "NG+1 unlocks ng_veteran",
			setupNGP: func(ngp *NewGamePlusComponent) {
				ngp.Cycle = 1
			},
			wantBonus:  "ng_veteran",
			shouldHave: true,
		},
		{
			name: "NG+5 unlocks seasoned_adventurer",
			setupNGP: func(ngp *NewGamePlusComponent) {
				ngp.Cycle = 5
			},
			wantBonus:  "seasoned_adventurer",
			shouldHave: true,
		},
		{
			name: "NG+10 unlocks legend_reborn",
			setupNGP: func(ngp *NewGamePlusComponent) {
				ngp.Cycle = 10
			},
			wantBonus:  "legend_reborn",
			shouldHave: true,
		},
		{
			name: "100 hours unlocks dedicated_player",
			setupNGP: func(ngp *NewGamePlusComponent) {
				ngp.TotalPlaytime = 360001 // Just over 100 hours
			},
			wantBonus:  "dedicated_player",
			shouldHave: true,
		},
		{
			name: "10000 kills unlocks master_slayer",
			setupNGP: func(ngp *NewGamePlusComponent) {
				ngp.AddToLegacyStat("enemies_killed", 10001)
			},
			wantBonus:  "master_slayer",
			shouldHave: true,
		},
		{
			name: "First playthrough has no unlocks",
			setupNGP: func(ngp *NewGamePlusComponent) {
				// Default state
			},
			wantBonus:  "ng_veteran",
			shouldHave: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewNewGamePlusSystem(world)

			entity := NewEntity(1)
			ngp := NewNewGamePlusComponent()
			tt.setupNGP(ngp)
			entity.AddComponent(ngp)

			system.Update([]*Entity{entity}, 0.0)

			if ngp.HasBonus(tt.wantBonus) != tt.shouldHave {
				t.Errorf("HasBonus(%q) = %v, want %v", tt.wantBonus, ngp.HasBonus(tt.wantBonus), tt.shouldHave)
			}
		})
	}
}

func TestCalculateNGPlusMultiplier(t *testing.T) {
	tests := []struct {
		name       string
		cycle      int
		base       float64
		scaling    float64
		wantApprox float64
		tolerance  float64
	}{
		{"cycle 0", 0, 1.0, 0.2, 1.0, 0.01},
		{"cycle 1", 1, 1.0, 0.2, 1.139, 0.01},            // 1.0 + 0.2 * ln(2) ≈ 1.139
		{"cycle 5", 5, 1.0, 0.2, 1.358, 0.01},            // 1.0 + 0.2 * ln(6) ≈ 1.358
		{"cycle 10", 10, 1.0, 0.2, 1.480, 0.02},          // 1.0 + 0.2 * ln(11) ≈ 1.480
		{"negative scaling", 1, 1.0, -0.05, 0.965, 0.01}, // XP reduction
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateNGPlusMultiplier(tt.cycle, tt.base, tt.scaling)
			if got < tt.wantApprox-tt.tolerance || got > tt.wantApprox+tt.tolerance {
				t.Errorf("CalculateNGPlusMultiplier(%d, %f, %f) = %f, want ~%f",
					tt.cycle, tt.base, tt.scaling, got, tt.wantApprox)
			}
		})
	}
}

func TestNewGamePlusSystem_GetEnemyHealthMultiplier(t *testing.T) {
	world := NewWorld()
	system := NewNewGamePlusSystem(world)

	entity := NewEntity(1)
	ngp := NewNewGamePlusComponent()
	ngp.Cycle = 5
	entity.AddComponent(ngp)

	mult := system.GetEnemyHealthMultiplier(entity)
	// Should be > 1.0 for NG+5
	if mult <= 1.0 {
		t.Errorf("Health multiplier should be > 1.0 for NG+5, got %f", mult)
	}
}

func TestNewGamePlusSystem_GetEnemyDamageMultiplier(t *testing.T) {
	world := NewWorld()
	system := NewNewGamePlusSystem(world)

	entity := NewEntity(1)
	ngp := NewNewGamePlusComponent()
	ngp.Cycle = 5
	entity.AddComponent(ngp)

	mult := system.GetEnemyDamageMultiplier(entity)
	// Should be > 1.0 for NG+5
	if mult <= 1.0 {
		t.Errorf("Damage multiplier should be > 1.0 for NG+5, got %f", mult)
	}
}

func TestNewGamePlusSystem_GetLootQualityBonus(t *testing.T) {
	world := NewWorld()
	system := NewNewGamePlusSystem(world)

	entity := NewEntity(1)
	ngp := NewNewGamePlusComponent()
	ngp.Cycle = 5
	entity.AddComponent(ngp)

	bonus := system.GetLootQualityBonus(entity)
	// Should be > 0 for NG+5
	if bonus <= 0 {
		t.Errorf("Loot quality bonus should be > 0 for NG+5, got %f", bonus)
	}
}

func TestNewGamePlusSystem_GetXPMultiplier(t *testing.T) {
	world := NewWorld()
	system := NewNewGamePlusSystem(world)

	entity := NewEntity(1)
	ngp := NewNewGamePlusComponent()
	ngp.Cycle = 5
	entity.AddComponent(ngp)

	mult := system.GetXPMultiplier(entity)
	// Should be slightly less than 1.0 for NG+5, but not below 0.5
	if mult >= 1.0 || mult < 0.5 {
		t.Errorf("XP multiplier should be between 0.5 and 1.0 for NG+5, got %f", mult)
	}
}

func TestNewGamePlusSystem_GetXPMultiplier_Minimum(t *testing.T) {
	world := NewWorld()
	system := NewNewGamePlusSystem(world)

	entity := NewEntity(1)
	ngp := NewNewGamePlusComponent()
	ngp.Cycle = 99 // Very high NG+
	entity.AddComponent(ngp)

	mult := system.GetXPMultiplier(entity)
	// Should not go below 0.5
	if mult < 0.5 {
		t.Errorf("XP multiplier should not go below 0.5, got %f", mult)
	}
}

func TestNewGamePlusSystem_GetNGPlusMultiplier_NoComponent(t *testing.T) {
	world := NewWorld()
	system := NewNewGamePlusSystem(world)

	entity := NewEntity(1)
	// No NG+ component

	mult := system.GetNGPlusMultiplier(entity, 1.0, 0.2)
	if mult != 1.0 {
		t.Errorf("Multiplier should be base (1.0) without component, got %f", mult)
	}
}

func TestGetBonusDescription(t *testing.T) {
	tests := []struct {
		bonusID string
		wantLen int // Just check non-empty
	}{
		{"ng_veteran", 10},
		{"seasoned_adventurer", 10},
		{"legend_reborn", 10},
		{"dedicated_player", 10},
		{"master_slayer", 10},
		{"unknown_bonus", 5}, // Should return "Unknown bonus"
	}

	for _, tt := range tests {
		t.Run(tt.bonusID, func(t *testing.T) {
			desc := GetBonusDescription(tt.bonusID)
			if len(desc) < tt.wantLen {
				t.Errorf("GetBonusDescription(%q) = %q, expected longer description", tt.bonusID, desc)
			}
		})
	}
}

func TestGetAllBonuses(t *testing.T) {
	bonuses := GetAllBonuses()
	if len(bonuses) != 5 {
		t.Errorf("GetAllBonuses() returned %d bonuses, want 5", len(bonuses))
	}

	expectedBonuses := []string{
		"ng_veteran",
		"seasoned_adventurer",
		"legend_reborn",
		"dedicated_player",
		"master_slayer",
	}

	for _, expected := range expectedBonuses {
		if _, ok := bonuses[expected]; !ok {
			t.Errorf("GetAllBonuses() missing bonus %q", expected)
		}
	}
}

func TestNewGamePlusSystem_IsEligibleForNewGamePlus(t *testing.T) {
	world := NewWorld()
	system := NewNewGamePlusSystem(world)

	// Test with no relevant components
	entity := NewEntity(1)
	if system.IsEligibleForNewGamePlus(entity) {
		t.Error("Should not be eligible without completion markers")
	}

	// Test with player stats showing completion
	entity2 := NewEntity(2)
	playerStats := NewPlayerStatisticsComponent()
	playerStats.IncrementStat("main_story_completed", 1)
	entity2.AddComponent(playerStats)
	if !system.IsEligibleForNewGamePlus(entity2) {
		t.Error("Should be eligible with main_story_completed stat")
	}
}

func BenchmarkNewGamePlusSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewNewGamePlusSystem(world)

	entity := NewEntity(1)
	ngp := NewNewGamePlusComponent()
	entity.AddComponent(ngp)
	entities := []*Entity{entity}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016) // 60 FPS frame time
	}
}

func BenchmarkCalculateNGPlusMultiplier(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateNGPlusMultiplier(5, 1.0, 0.2)
	}
}
