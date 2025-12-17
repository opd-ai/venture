// Package engine provides tests for the NG+ Difficulty system.
// Phase 113: Difficulty Scaling System
package engine

import (
	"testing"
)

func TestNewNGPlusDifficultySystem(t *testing.T) {
	world := NewWorld()
	system := NewNGPlusDifficultySystem(world)

	if system == nil {
		t.Fatal("NewNGPlusDifficultySystem returned nil")
	}
	if !system.IsScalingEnabled() {
		t.Error("Scaling should be enabled by default")
	}
	if system.GetCurrentNGPlusCycle() != 0 {
		t.Error("Initial NG+ cycle should be 0")
	}
}

func TestNGPlusDifficultySystem_Update_NoScaling(t *testing.T) {
	world := NewWorld()
	system := NewNGPlusDifficultySystem(world)

	// Create a player without NG+ component (first playthrough)
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	player.AddComponent(&TeamComponent{TeamID: 1})

	// Create an enemy
	enemy := world.CreateEntity()
	enemy.AddComponent(&PositionComponent{X: 100, Y: 100})
	enemy.AddComponent(&TeamComponent{TeamID: 2})
	enemy.AddComponent(&HealthComponent{Current: 100, Max: 100})
	enemy.AddComponent(NewStatsComponent())

	entities := []*Entity{player, enemy}
	system.Update(entities, 0.016)

	// Enemy should not have difficulty component (no NG+ active)
	if _, hasDiff := enemy.GetComponent("ngplus_difficulty"); hasDiff {
		t.Error("Enemy should not have difficulty component in first playthrough")
	}
}

func TestNGPlusDifficultySystem_Update_WithNGPlus(t *testing.T) {
	world := NewWorld()
	system := NewNGPlusDifficultySystem(world)

	// Create a player with NG+ component at cycle 5
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	player.AddComponent(&TeamComponent{TeamID: 1})
	ngpComp := NewNewGamePlusComponent()
	ngpComp.Cycle = 5
	player.AddComponent(ngpComp)

	// Create an enemy with base stats
	enemy := world.CreateEntity()
	enemy.AddComponent(&PositionComponent{X: 100, Y: 100})
	enemy.AddComponent(&TeamComponent{TeamID: 2})
	enemy.AddComponent(&HealthComponent{Current: 100, Max: 100})
	stats := NewStatsComponent()
	stats.Attack = 10
	stats.Defense = 5
	enemy.AddComponent(stats)

	entities := []*Entity{player, enemy}
	system.Update(entities, 0.016)

	// Enemy should now have difficulty component
	diffComp, hasDiff := enemy.GetComponent("ngplus_difficulty")
	if !hasDiff {
		t.Fatal("Enemy should have difficulty component after NG+ scaling")
	}

	diff := diffComp.(*NGPlusDifficultyComponent)
	if diff.GetNGPlusCycle() != 5 {
		t.Errorf("Difficulty NGPlusCycle = %d, want 5", diff.GetNGPlusCycle())
	}

	// Check that health was scaled
	healthComp, _ := enemy.GetComponent("health")
	health := healthComp.(*HealthComponent)
	if health.Max <= 100 {
		t.Errorf("Health.Max = %v, should be > 100 after NG+5 scaling", health.Max)
	}

	// Check that attack was scaled
	statsComp, _ := enemy.GetComponent("stats")
	scaledStats := statsComp.(*StatsComponent)
	if scaledStats.Attack <= 10 {
		t.Errorf("Attack = %v, should be > 10 after NG+5 scaling", scaledStats.Attack)
	}
}

func TestNGPlusDifficultySystem_DisableScaling(t *testing.T) {
	world := NewWorld()
	system := NewNGPlusDifficultySystem(world)
	system.SetScalingEnabled(false)

	// Create player with NG+
	player := world.CreateEntity()
	player.AddComponent(&TeamComponent{TeamID: 1})
	ngpComp := NewNewGamePlusComponent()
	ngpComp.Cycle = 10
	player.AddComponent(ngpComp)

	// Create enemy
	enemy := world.CreateEntity()
	enemy.AddComponent(&TeamComponent{TeamID: 2})
	enemy.AddComponent(&HealthComponent{Current: 100, Max: 100})

	entities := []*Entity{player, enemy}
	system.Update(entities, 0.016)

	// Enemy should not be scaled when scaling is disabled
	if _, hasDiff := enemy.GetComponent("ngplus_difficulty"); hasDiff {
		t.Error("Enemy should not be scaled when scaling is disabled")
	}
}

func TestNGPlusDifficultySystem_BossEnhancement(t *testing.T) {
	world := NewWorld()
	system := NewNGPlusDifficultySystem(world)
	system.SetCurrentNGPlusCycle(10)

	// Create a boss entity (high health)
	boss := world.CreateEntity()
	boss.AddComponent(&TeamComponent{TeamID: 2})
	boss.AddComponent(&HealthComponent{Current: 1000, Max: 1000})
	boss.AddComponent(NewStatsComponent())

	// Apply scaling
	system.ScaleEnemyForNGPlus(boss)

	// Check boss enhancements
	diffComp, _ := boss.GetComponent("ngplus_difficulty")
	diff := diffComp.(*NGPlusDifficultyComponent)

	if diff.GetBossEnhancementLevel() != 3 {
		t.Errorf("BossEnhancementLevel = %d, want 3 for NG+10", diff.GetBossEnhancementLevel())
	}
	if !diff.HasEnragedPhase {
		t.Error("Boss should have enraged phase at NG+10")
	}
}

func TestNGPlusDifficultySystem_GetLootQualityBonus(t *testing.T) {
	world := NewWorld()
	system := NewNGPlusDifficultySystem(world)

	// First playthrough - no bonus
	if bonus := system.GetLootQualityBonus(); bonus != 0.0 {
		t.Errorf("Loot bonus = %v, want 0 for first playthrough", bonus)
	}

	// NG+5 - should have bonus
	system.SetCurrentNGPlusCycle(5)
	if bonus := system.GetLootQualityBonus(); bonus <= 0.0 {
		t.Errorf("Loot bonus = %v, should be > 0 for NG+5", bonus)
	}
}

func TestNGPlusDifficultySystem_GetXPMultiplier(t *testing.T) {
	world := NewWorld()
	system := NewNGPlusDifficultySystem(world)

	// First playthrough - full XP
	if mult := system.GetXPMultiplier(); mult != 1.0 {
		t.Errorf("XP multiplier = %v, want 1.0 for first playthrough", mult)
	}

	// NG+5 - reduced XP
	system.SetCurrentNGPlusCycle(5)
	if mult := system.GetXPMultiplier(); mult >= 1.0 {
		t.Errorf("XP multiplier = %v, should be < 1.0 for NG+5", mult)
	}

	// Very high NG+ - minimum 50%
	system.SetCurrentNGPlusCycle(99)
	if mult := system.GetXPMultiplier(); mult < 0.5 {
		t.Errorf("XP multiplier = %v, should not be < 0.5", mult)
	}
}

func TestNGPlusDifficultySystem_ApplyXPScaling(t *testing.T) {
	world := NewWorld()
	system := NewNGPlusDifficultySystem(world)
	system.SetCurrentNGPlusCycle(5)

	baseXP := 100.0
	scaledXP := system.ApplyXPScaling(baseXP)

	if scaledXP >= baseXP {
		t.Errorf("Scaled XP = %v, should be < %v for NG+5", scaledXP, baseXP)
	}
}

func TestNGPlusDifficultySystem_ApplyLootQualityScaling(t *testing.T) {
	world := NewWorld()
	system := NewNGPlusDifficultySystem(world)
	system.SetCurrentNGPlusCycle(5)

	baseChance := 0.10 // 10% rare chance
	scaledChance := system.ApplyLootQualityScaling(baseChance)

	if scaledChance <= baseChance {
		t.Errorf("Scaled chance = %v, should be > %v for NG+5", scaledChance, baseChance)
	}
}

func TestNGPlusDifficultySystem_GetDifficultyInfo(t *testing.T) {
	world := NewWorld()
	system := NewNGPlusDifficultySystem(world)

	// First playthrough
	info := system.GetDifficultyInfo()
	if info.Label != "Normal" {
		t.Errorf("Label = %s, want Normal", info.Label)
	}
	if info.HealthMultiplier != 1.0 {
		t.Errorf("HealthMultiplier = %v, want 1.0", info.HealthMultiplier)
	}

	// NG+10
	system.SetCurrentNGPlusCycle(10)
	info = system.GetDifficultyInfo()
	if info.Label != "Legendary" {
		t.Errorf("Label = %s, want Legendary for NG+10", info.Label)
	}
	if info.HealthMultiplier <= 1.0 {
		t.Errorf("HealthMultiplier = %v, should be > 1.0 for NG+10", info.HealthMultiplier)
	}
}

func TestNGPlusDifficultySystem_ShouldUnlockMechanic(t *testing.T) {
	world := NewWorld()
	system := NewNGPlusDifficultySystem(world)

	tests := []struct {
		cycle    int
		level    int
		expected bool
	}{
		{0, 0, true},
		{0, 1, false},
		{3, 0, true},
		{3, 1, true},
		{3, 2, false},
		{5, 2, true},
		{10, 4, true},
	}

	for _, tt := range tests {
		system.SetCurrentNGPlusCycle(tt.cycle)
		result := system.ShouldUnlockMechanic(tt.level)
		if result != tt.expected {
			t.Errorf("ShouldUnlockMechanic(%d) at cycle %d = %v, want %v",
				tt.level, tt.cycle, result, tt.expected)
		}
	}
}

func TestNGPlusDifficultySystem_GetUnlockedAbilities(t *testing.T) {
	world := NewWorld()
	system := NewNGPlusDifficultySystem(world)

	// First playthrough - no abilities
	abilities := system.GetUnlockedAbilities()
	if len(abilities) != 0 {
		t.Errorf("Abilities count = %d, want 0 for first playthrough", len(abilities))
	}

	// NG+5 - should have level 1 and 2 abilities
	system.SetCurrentNGPlusCycle(5)
	abilities = system.GetUnlockedAbilities()
	if len(abilities) < 4 {
		t.Errorf("Abilities count = %d, want >= 4 for NG+5", len(abilities))
	}
}

func TestNGPlusDifficultySystem_ResetScaling(t *testing.T) {
	world := NewWorld()
	system := NewNGPlusDifficultySystem(world)
	system.SetCurrentNGPlusCycle(5)

	// Create enemy with scaling
	enemy := world.CreateEntity()
	enemy.AddComponent(&TeamComponent{TeamID: 2})
	enemy.AddComponent(&HealthComponent{Current: 100, Max: 100})
	enemy.AddComponent(NewNGPlusDifficultyComponentForCycle(5))

	entities := []*Entity{enemy}
	system.ResetScaling(entities)

	// Scaling should be removed
	if _, hasDiff := enemy.GetComponent("ngplus_difficulty"); hasDiff {
		t.Error("Difficulty component should be removed after reset")
	}
	if system.GetCurrentNGPlusCycle() != 0 {
		t.Error("NG+ cycle should be reset to 0")
	}
}

func TestNGPlusDifficultySystem_Callback(t *testing.T) {
	world := NewWorld()
	system := NewNGPlusDifficultySystem(world)
	system.SetCurrentNGPlusCycle(5)

	var callbackEntityID uint64
	var callbackCycle int
	system.SetOnEnemyScaled(func(entityID uint64, cycle int) {
		callbackEntityID = entityID
		callbackCycle = cycle
	})

	// Create and scale enemy
	enemy := world.CreateEntity()
	enemy.AddComponent(&TeamComponent{TeamID: 2})
	enemy.AddComponent(&HealthComponent{Current: 100, Max: 100})
	enemy.AddComponent(NewStatsComponent())

	system.ScaleEnemyForNGPlus(enemy)

	if callbackEntityID != enemy.ID {
		t.Errorf("Callback entityID = %d, want %d", callbackEntityID, enemy.ID)
	}
	if callbackCycle != 5 {
		t.Errorf("Callback cycle = %d, want 5", callbackCycle)
	}
}

func TestNGPlusDifficultySystem_SkipsPlayerTeam(t *testing.T) {
	world := NewWorld()
	system := NewNGPlusDifficultySystem(world)
	system.SetCurrentNGPlusCycle(5)

	// Create player entity
	player := world.CreateEntity()
	player.AddComponent(&TeamComponent{TeamID: 1}) // Player team
	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	player.AddComponent(NewStatsComponent())
	player.AddComponent(NewNewGamePlusComponent())

	entities := []*Entity{player}
	system.Update(entities, 0.016)

	// Player should not get difficulty scaling
	if _, hasDiff := player.GetComponent("ngplus_difficulty"); hasDiff {
		t.Error("Player should not have difficulty component")
	}
}

func TestNGPlusDifficultySystem_ScaleEntityStats(t *testing.T) {
	world := NewWorld()
	system := NewNGPlusDifficultySystem(world)

	// Create enemy with attack component
	enemy := world.CreateEntity()
	enemy.AddComponent(&TeamComponent{TeamID: 2})
	enemy.AddComponent(&HealthComponent{Current: 100, Max: 100})
	enemy.AddComponent(&AttackComponent{Damage: 20, Range: 50})
	enemy.AddComponent(NewStatsComponent())

	system.SetCurrentNGPlusCycle(5)
	system.ScaleEnemyForNGPlus(enemy)

	// Check attack component was scaled
	attackComp, _ := enemy.GetComponent("attack")
	attack := attackComp.(*AttackComponent)
	if attack.Damage <= 20 {
		t.Errorf("Attack.Damage = %v, should be > 20 after NG+5 scaling", attack.Damage)
	}
}

func TestNGPlusDifficultySystem_DoesNotDoubleScale(t *testing.T) {
	world := NewWorld()
	system := NewNGPlusDifficultySystem(world)
	system.SetCurrentNGPlusCycle(5)

	// Create enemy
	enemy := world.CreateEntity()
	enemy.AddComponent(&TeamComponent{TeamID: 2})
	enemy.AddComponent(&HealthComponent{Current: 100, Max: 100})
	enemy.AddComponent(NewStatsComponent())

	// Scale once
	system.ScaleEnemyForNGPlus(enemy)
	healthComp1, _ := enemy.GetComponent("health")
	health1 := healthComp1.(*HealthComponent)
	firstMax := health1.Max

	// Scale again
	system.ScaleEnemyForNGPlus(enemy)
	healthComp2, _ := enemy.GetComponent("health")
	health2 := healthComp2.(*HealthComponent)

	if health2.Max != firstMax {
		t.Errorf("Health.Max changed from %v to %v on second scaling", firstMax, health2.Max)
	}
}

func TestNGPlusDifficultySystem_NilWorld(t *testing.T) {
	// Should not panic with nil world
	system := NewNGPlusDifficultySystem(nil)
	if system == nil {
		t.Fatal("System should be created even with nil world")
	}

	// Update should not panic
	system.Update([]*Entity{}, 0.016)
}

func BenchmarkNGPlusDifficultySystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewNGPlusDifficultySystem(world)

	// Create player with NG+
	player := world.CreateEntity()
	player.AddComponent(&TeamComponent{TeamID: 1})
	ngpComp := NewNewGamePlusComponent()
	ngpComp.Cycle = 5
	player.AddComponent(ngpComp)

	// Create 100 enemies
	entities := []*Entity{player}
	for i := 0; i < 100; i++ {
		enemy := world.CreateEntity()
		enemy.AddComponent(&TeamComponent{TeamID: 2})
		enemy.AddComponent(&HealthComponent{Current: 100, Max: 100})
		enemy.AddComponent(NewStatsComponent())
		entities = append(entities, enemy)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset scaling state for re-run
		for _, e := range entities[1:] {
			e.RemoveComponent("ngplus_difficulty")
		}
		system.Update(entities, 0.016)
	}
}

func BenchmarkNGPlusDifficultySystem_ScaleEnemy(b *testing.B) {
	world := NewWorld()
	system := NewNGPlusDifficultySystem(world)
	system.SetCurrentNGPlusCycle(10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enemy := world.CreateEntity()
		enemy.AddComponent(&TeamComponent{TeamID: 2})
		enemy.AddComponent(&HealthComponent{Current: 100, Max: 100})
		enemy.AddComponent(NewStatsComponent())
		system.ScaleEnemyForNGPlus(enemy)
	}
}
