package engine

import (
	"math"
	"testing"
)

func TestNewStatusEffectCriticalChanceSystem(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectCriticalChanceSystem(world, 12345)

	if system == nil {
		t.Fatal("NewStatusEffectCriticalChanceSystem returned nil")
	}

	if system.world != world {
		t.Error("World reference not set correctly")
	}

	if system.critCache == nil {
		t.Error("Crit cache not initialized")
	}
}

func TestStatusEffectCriticalChanceSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectCriticalChanceSystem(world, 12345)

	tests := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range tests {
		system.SetGenre(genre)
		if system.genre != genre {
			t.Errorf("Genre not set correctly: got %s, want %s", system.genre, genre)
		}
	}
}

func TestStatusEffectCriticalChanceSystem_BasicUpdate(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectCriticalChanceSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create entity with stats
	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.CritChance = 0.10 // 10% base crit
	entity.AddComponent(stats)

	// Add blessed status effect (+10% crit)
	blessedEffect := &StatusEffectComponent{
		EffectType: "blessed",
		Duration:   10.0,
		Magnitude:  1.0,
	}
	entity.AddComponent(blessedEffect)

	// Update system
	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// Check crit was modified
	expectedCrit := 0.20 // 10% base + 10% blessed
	if math.Abs(stats.CritChance-expectedCrit) > 0.001 {
		t.Errorf("CritChance = %f, want %f", stats.CritChance, expectedCrit)
	}
}

func TestStatusEffectCriticalChanceSystem_NegativeModifier(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectCriticalChanceSystem(world, 12345)
	system.SetGenre("fantasy")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.CritChance = 0.15 // 15% base crit
	entity.AddComponent(stats)

	// Add cursed status effect (-10% crit)
	cursedEffect := &StatusEffectComponent{
		EffectType: "cursed",
		Duration:   10.0,
		Magnitude:  1.0,
	}
	entity.AddComponent(cursedEffect)

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	expectedCrit := 0.05 // 15% base - 10% cursed
	if math.Abs(stats.CritChance-expectedCrit) > 0.001 {
		t.Errorf("CritChance = %f, want %f", stats.CritChance, expectedCrit)
	}
}

func TestStatusEffectCriticalChanceSystem_MultipleEffects(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectCriticalChanceSystem(world, 12345)
	system.SetGenre("fantasy")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.CritChance = 0.10
	entity.AddComponent(stats)

	// Note: In the current ECS, only one status effect component per entity is stored
	// (they share the same Type() "status_effect"). This tests the single effect case.
	// Add "focused" effect which has the highest crit bonus (+15%)
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "focused",
		Duration:   10.0,
		Magnitude:  1.0,
	})

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// 10% base + 15% focused = 25%
	expectedCrit := 0.25
	if math.Abs(stats.CritChance-expectedCrit) > 0.001 {
		t.Errorf("CritChance = %f, want %f", stats.CritChance, expectedCrit)
	}
}

func TestStatusEffectCriticalChanceSystem_ClampMin(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectCriticalChanceSystem(world, 12345)
	system.SetGenre("fantasy")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.CritChance = 0.05 // Low base crit
	entity.AddComponent(stats)

	// Add multiple negative effects to exceed base
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "frozen",
		Duration:   10.0,
		Magnitude:  1.0,
	})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "blinded",
		Duration:   10.0,
		Magnitude:  1.0,
	})

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// Should be clamped to 0
	if stats.CritChance < 0 {
		t.Errorf("CritChance should be clamped to 0, got %f", stats.CritChance)
	}
}

func TestStatusEffectCriticalChanceSystem_ClampMax(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectCriticalChanceSystem(world, 12345)
	system.SetGenre("fantasy")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.CritChance = 0.90 // Very high base crit
	entity.AddComponent(stats)

	// Add multiple positive effects to exceed 100%
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "blessed",
		Duration:   10.0,
		Magnitude:  1.0,
	})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "focused",
		Duration:   10.0,
		Magnitude:  1.0,
	})

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// Should be clamped to 1.0
	if stats.CritChance > 1.0 {
		t.Errorf("CritChance should be clamped to 1.0, got %f", stats.CritChance)
	}
}

func TestStatusEffectCriticalChanceSystem_ExpiredEffects(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectCriticalChanceSystem(world, 12345)
	system.SetGenre("fantasy")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.CritChance = 0.10
	entity.AddComponent(stats)

	// Add expired effect
	expiredEffect := &StatusEffectComponent{
		EffectType: "blessed",
		Duration:   0.0, // Expired
		Magnitude:  1.0,
	}
	entity.AddComponent(expiredEffect)

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// Crit should remain unchanged (expired effect ignored)
	if math.Abs(stats.CritChance-0.10) > 0.001 {
		t.Errorf("CritChance = %f, want 0.10 (expired effect should be ignored)", stats.CritChance)
	}
}

func TestStatusEffectCriticalChanceSystem_NoStats(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectCriticalChanceSystem(world, 12345)

	// Create entity without stats
	entity := world.CreateEntity()
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "blessed",
		Duration:   10.0,
		Magnitude:  1.0,
	})

	entities := []*Entity{entity}
	// Should not panic
	system.Update(entities, 0.016)
}

func TestStatusEffectCriticalChanceSystem_GenreScaling(t *testing.T) {
	tests := []struct {
		name        string
		genre       string
		effectType  string
		baseCrit    float64
		expectedMin float64
		expectedMax float64
	}{
		{"fantasy_blessed", "fantasy", "blessed", 0.10, 0.19, 0.21},
		{"scifi_blessed", "scifi", "blessed", 0.10, 0.20, 0.22},         // 10% boost
		{"horror_blessed", "horror", "blessed", 0.10, 0.17, 0.19},       // 20% weaker
		{"horror_cursed", "horror", "cursed", 0.20, 0.05, 0.09},         // 30% stronger debuff
		{"cyberpunk_blessed", "cyberpunk", "blessed", 0.10, 0.21, 0.23}, // 15% boost
		{"postapoc_blessed", "postapoc", "blessed", 0.10, 0.21, 0.23},   // 20% stronger
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewStatusEffectCriticalChanceSystem(world, 12345)
			system.SetGenre(tt.genre)

			entity := world.CreateEntity()
			stats := NewStatsComponent()
			stats.CritChance = tt.baseCrit
			entity.AddComponent(stats)

			entity.AddComponent(&StatusEffectComponent{
				EffectType: tt.effectType,
				Duration:   10.0,
				Magnitude:  1.0,
			})

			entities := []*Entity{entity}
			system.Update(entities, 0.016)

			if stats.CritChance < tt.expectedMin || stats.CritChance > tt.expectedMax {
				t.Errorf("%s: CritChance = %f, want between %f and %f",
					tt.name, stats.CritChance, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestStatusEffectCriticalChanceSystem_GetCritModifier(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectCriticalChanceSystem(world, 12345)
	system.SetGenre("fantasy")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.CritChance = 0.10
	entity.AddComponent(stats)

	entity.AddComponent(&StatusEffectComponent{
		EffectType: "blessed",
		Duration:   10.0,
		Magnitude:  1.0,
	})

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	modifier := system.GetCritModifier(entity.ID)
	if math.Abs(modifier-0.10) > 0.001 {
		t.Errorf("GetCritModifier() = %f, want 0.10", modifier)
	}

	// Non-existent entity should return 0
	unknownMod := system.GetCritModifier(99999)
	if unknownMod != 0.0 {
		t.Errorf("GetCritModifier(unknown) = %f, want 0.0", unknownMod)
	}
}

func TestStatusEffectCriticalChanceSystem_AllEffectTypes(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectCriticalChanceSystem(world, 12345)

	tests := []struct {
		effectType string
		expected   float64
	}{
		{"blessed", 0.10},
		{"cursed", -0.10},
		{"haste", 0.05},
		{"strength", 0.08},
		{"weakness", -0.08},
		{"focused", 0.15},
		{"blinded", -0.15},
		{"blindness", -0.15},
		{"enraged", 0.12},
		{"frozen", -0.20},
		{"chilled", -0.03},
		{"burning", 0.05},
		{"poisoned", -0.05},
		{"fortify", 0.0},
		{"regeneration", 0.03},
		{"vulnerability", 0.0},
		{"unknown", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.effectType, func(t *testing.T) {
			effect := &StatusEffectComponent{
				EffectType: tt.effectType,
				Duration:   10.0,
				Magnitude:  1.0,
			}

			modifier := system.effectToCritModifier(effect)
			if math.Abs(modifier-tt.expected) > 0.001 {
				t.Errorf("effectToCritModifier(%s) = %f, want %f",
					tt.effectType, modifier, tt.expected)
			}
		})
	}
}

func TestStatusEffectCriticalChanceSystem_CacheClear(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectCriticalChanceSystem(world, 12345)
	system.SetGenre("fantasy")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.CritChance = 0.10
	entity.AddComponent(stats)

	effect := &StatusEffectComponent{
		EffectType: "blessed",
		Duration:   10.0,
		Magnitude:  1.0,
	}
	entity.AddComponent(effect)

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// Verify cache has entry
	if _, ok := system.critCache[entity.ID]; !ok {
		t.Error("Cache should have entry after update")
	}

	// Remove effect and update
	entity.RemoveComponent("status_effect")
	stats.CritChance = 0.10 // Reset for clarity

	system.Update(entities, 0.016)

	// Cache should be cleared/updated
	modifier := system.GetCritModifier(entity.ID)
	if modifier != 0.0 {
		t.Errorf("Cache modifier should be 0 after effect removed, got %f", modifier)
	}
}

func BenchmarkStatusEffectCriticalChanceSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewStatusEffectCriticalChanceSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create 1000 entities with varying status effects
	entities := make([]*Entity, 1000)
	for i := 0; i < 1000; i++ {
		entity := world.CreateEntity()
		stats := NewStatsComponent()
		stats.CritChance = 0.10
		entity.AddComponent(stats)

		// Add random effects
		if i%3 == 0 {
			entity.AddComponent(&StatusEffectComponent{
				EffectType: "blessed",
				Duration:   10.0,
				Magnitude:  1.0,
			})
		}
		if i%5 == 0 {
			entity.AddComponent(&StatusEffectComponent{
				EffectType: "cursed",
				Duration:   10.0,
				Magnitude:  1.0,
			})
		}
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}

func BenchmarkStatusEffectCriticalChanceSystem_SingleEntity(b *testing.B) {
	world := NewWorld()
	system := NewStatusEffectCriticalChanceSystem(world, 12345)
	system.SetGenre("fantasy")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.CritChance = 0.10
	entity.AddComponent(stats)

	entity.AddComponent(&StatusEffectComponent{
		EffectType: "blessed",
		Duration:   10.0,
		Magnitude:  1.0,
	})

	entities := []*Entity{entity}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}

// TestG33_StatusEffectCritChance_NoCritAccumulation verifies that Update called
// N frames with the same active effect produces a stable CritChance, and that
// when the effect expires the CritChance returns to the baseline.
// This is the regression test for G33.
func TestG33_StatusEffectCritChance_NoCritAccumulation(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectCriticalChanceSystem(world, 1)
	system.SetGenre("fantasy")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	baseCrit := 0.05
	stats.CritChance = baseCrit
	entity.AddComponent(stats)

	blessedEffect := &StatusEffectComponent{
		EffectType: "blessed",
		Duration:   100.0, // long-lived so it does not expire during the 100 frames
		Magnitude:  1.0,
	}
	entity.AddComponent(blessedEffect)

	entities := []*Entity{entity}

	// Run 100 frames with the effect active.
	for i := 0; i < 100; i++ {
		system.Update(entities, 0.016)
	}

	expectedCrit := baseCrit + 0.10 // blessed = +10%
	if math.Abs(stats.CritChance-expectedCrit) > 0.001 {
		t.Errorf("G33 accumulation: CritChance after 100 frames = %.4f, want %.4f", stats.CritChance, expectedCrit)
	}

	// Expire the effect: set Duration to 0 (IsExpired returns true).
	blessedEffect.Duration = 0.0

	// Run a few more frames — CritChance must return to baseline.
	for i := 0; i < 5; i++ {
		system.Update(entities, 0.016)
	}

	if math.Abs(stats.CritChance-baseCrit) > 0.001 {
		t.Errorf("G33 restore: CritChance after effect expiry = %.4f, want %.4f (baseline)", stats.CritChance, baseCrit)
	}
}
