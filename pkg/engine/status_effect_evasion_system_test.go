package engine

import (
	"math/rand"
	"testing"
)

func TestNewStatusEffectEvasionSystem(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectEvasionSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewStatusEffectEvasionSystem returned nil")
	}
	if sys.world != world {
		t.Error("world reference not set correctly")
	}
	if sys.rng == nil {
		t.Error("RNG not initialized")
	}
	if sys.evasionCache == nil {
		t.Error("evasion cache not initialized")
	}
}

func TestStatusEffectEvasionSystem_SetGenre(t *testing.T) {
	sys := NewStatusEffectEvasionSystem(nil, 12345)

	tests := []struct {
		genre string
	}{
		{"fantasy"},
		{"scifi"},
		{"horror"},
		{"cyberpunk"},
		{"postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys.SetGenre(tt.genre)
			if sys.genre != tt.genre {
				t.Errorf("genre = %s, want %s", sys.genre, tt.genre)
			}
		})
	}
}

func TestStatusEffectEvasionSystem_ChilledEffect(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectEvasionSystem(world, 12345)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Evasion = 0.20 // 20% base evasion
	entity.AddComponent(stats)
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "chilled",
		Magnitude:  0.15,
		Duration:   5.0,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// Chilled: -5% evasion
	// Expected: 0.20 - 0.05 = 0.15
	expectedEvasion := 0.15
	if stats.Evasion < expectedEvasion-0.001 || stats.Evasion > expectedEvasion+0.001 {
		t.Errorf("Chilled effect: expected evasion ~%f, got %f", expectedEvasion, stats.Evasion)
	}
}

func TestStatusEffectEvasionSystem_FrozenEffect(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectEvasionSystem(world, 12345)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Evasion = 0.25 // 25% base evasion
	entity.AddComponent(stats)
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "frozen",
		Magnitude:  0.5,
		Duration:   2.0,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// Frozen: -30% evasion
	// Expected: 0.25 - 0.30 = -0.05, clamped to 0.0
	if stats.Evasion != 0.0 {
		t.Errorf("Frozen effect: expected evasion 0.0, got %f", stats.Evasion)
	}
}

func TestStatusEffectEvasionSystem_HasteEffect(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectEvasionSystem(world, 12345)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Evasion = 0.20 // 20% base evasion
	entity.AddComponent(stats)
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "haste",
		Magnitude:  0.5,
		Duration:   5.0,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// Haste: +15% evasion
	// Expected: 0.20 + 0.15 = 0.35
	expectedEvasion := 0.35
	if stats.Evasion < expectedEvasion-0.001 || stats.Evasion > expectedEvasion+0.001 {
		t.Errorf("Haste effect: expected evasion ~%f, got %f", expectedEvasion, stats.Evasion)
	}
}

func TestStatusEffectEvasionSystem_BlindedEffect(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectEvasionSystem(world, 12345)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Evasion = 0.30 // 30% base evasion
	entity.AddComponent(stats)
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "blinded",
		Magnitude:  1.0,
		Duration:   3.0,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// Blinded: -20% evasion
	// Expected: 0.30 - 0.20 = 0.10
	expectedEvasion := 0.10
	if stats.Evasion < expectedEvasion-0.001 || stats.Evasion > expectedEvasion+0.001 {
		t.Errorf("Blinded effect: expected evasion ~%f, got %f", expectedEvasion, stats.Evasion)
	}
}

func TestStatusEffectEvasionSystem_StackedEffects(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectEvasionSystem(world, 12345)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Evasion = 0.30 // 30% base evasion
	entity.AddComponent(stats)
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})

	// Note: ECS uses component Type() as key, so only ONE StatusEffectComponent
	// can exist per entity. The second AddComponent overwrites the first.
	// Add just haste effect (later AddComponent wins)
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "haste",
		Magnitude:  0.5,
		Duration:   5.0,
	})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "blessed",
		Magnitude:  1.0,
		Duration:   5.0,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// Only blessed effect applies (it overwrote haste): +10%
	// Expected: 0.30 + 0.10 = 0.40
	expectedEvasion := 0.40
	if stats.Evasion < expectedEvasion-0.001 || stats.Evasion > expectedEvasion+0.001 {
		t.Errorf("Single effect (blessed overwrites haste): expected evasion ~%f, got %f", expectedEvasion, stats.Evasion)
	}
}

func TestStatusEffectEvasionSystem_EvasionCap(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectEvasionSystem(world, 12345)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Evasion = 0.90 // 90% base evasion (very high)
	entity.AddComponent(stats)
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "haste",
		Magnitude:  0.5,
		Duration:   5.0,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// Should be capped at 0.95
	if stats.Evasion > 0.95 {
		t.Errorf("Evasion should be capped at 0.95, got %f", stats.Evasion)
	}
}

func TestStatusEffectEvasionSystem_GenreScaling_Horror(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectEvasionSystem(world, 12345)
	sys.SetGenre("horror")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Evasion = 0.30
	entity.AddComponent(stats)
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "poisoned",
		Magnitude:  1.0,
		Duration:   5.0,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// Poisoned: -10% * 1.2 (horror scaling for debuffs) = -12%
	// Expected: 0.30 - 0.12 = 0.18
	expectedEvasion := 0.18
	if stats.Evasion < expectedEvasion-0.001 || stats.Evasion > expectedEvasion+0.001 {
		t.Errorf("Horror genre scaling: expected evasion ~%f, got %f", expectedEvasion, stats.Evasion)
	}
}

func TestStatusEffectEvasionSystem_GenreScaling_SciFi(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectEvasionSystem(world, 12345)
	sys.SetGenre("scifi")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Evasion = 0.30
	entity.AddComponent(stats)
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "poisoned",
		Magnitude:  1.0,
		Duration:   5.0,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// Poisoned: -10% * 0.8 (scifi tech resistance) = -8%
	// Expected: 0.30 - 0.08 = 0.22
	expectedEvasion := 0.22
	if stats.Evasion < expectedEvasion-0.001 || stats.Evasion > expectedEvasion+0.001 {
		t.Errorf("SciFi genre scaling: expected evasion ~%f, got %f", expectedEvasion, stats.Evasion)
	}
}

func TestStatusEffectEvasionSystem_HasEvasionEffect(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectEvasionSystem(world, 12345)

	tests := []struct {
		effectType string
		expected   bool
	}{
		{"chilled", true},
		{"frozen", true},
		{"haste", true},
		{"blinded", true},
		{"poisoned", true},
		{"stunned", true},
		{"blessed", true},
		{"cursed", true},
		{"wet", true},
		{"burning", false}, // burning doesn't affect evasion
		{"strength", false},
		{"fortify", false},
	}

	for _, tt := range tests {
		t.Run(tt.effectType, func(t *testing.T) {
			entity := world.CreateEntity()
			entity.AddComponent(&StatusEffectComponent{
				EffectType: tt.effectType,
				Duration:   5.0,
			})

			result := sys.HasEvasionEffect(entity)
			if result != tt.expected {
				t.Errorf("HasEvasionEffect(%s) = %v, want %v", tt.effectType, result, tt.expected)
			}
		})
	}
}

func TestStatusEffectEvasionSystem_GetEvasionModifier(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectEvasionSystem(world, 12345)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Evasion = 0.20
	entity.AddComponent(stats)
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "chilled",
		Magnitude:  0.15,
		Duration:   5.0,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	modifier := sys.GetEvasionModifier(entity.ID)
	expectedModifier := -0.05
	if modifier < expectedModifier-0.001 || modifier > expectedModifier+0.001 {
		t.Errorf("GetEvasionModifier = %f, want ~%f", modifier, expectedModifier)
	}
}

func TestStatusEffectEvasionSystem_ExpiredEffectIgnored(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectEvasionSystem(world, 12345)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Evasion = 0.20
	entity.AddComponent(stats)
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "chilled",
		Magnitude:  0.15,
		Duration:   0.0, // Already expired
	})

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// Expired effect should be ignored, evasion unchanged
	if stats.Evasion != 0.20 {
		t.Errorf("Expired effect should be ignored: expected evasion 0.20, got %f", stats.Evasion)
	}
}

func TestStatusEffectEvasionSystem_NoStatsComponent(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectEvasionSystem(world, 12345)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "chilled",
		Magnitude:  0.15,
		Duration:   5.0,
	})

	entities := []*Entity{entity}

	// Should not panic when entity has no stats component
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Update panicked with no stats component: %v", r)
		}
	}()

	sys.Update(entities, 0.016)
}

func TestStatusEffectEvasionSystem_Determinism(t *testing.T) {
	seed := int64(12345)

	// Run twice with same seed
	results := make([]float64, 2)
	for i := 0; i < 2; i++ {
		world := NewWorld()
		sys := NewStatusEffectEvasionSystem(world, seed)
		sys.SetGenre("fantasy")

		entity := world.CreateEntity()
		stats := NewStatsComponent()
		stats.Evasion = 0.20
		entity.AddComponent(stats)
		entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})
		entity.AddComponent(&StatusEffectComponent{
			EffectType: "chilled",
			Magnitude:  0.15,
			Duration:   5.0,
		})

		entities := []*Entity{entity}
		sys.Update(entities, 0.016)
		results[i] = stats.Evasion
	}

	if results[0] != results[1] {
		t.Errorf("Results not deterministic: run1=%f, run2=%f", results[0], results[1])
	}
}

func BenchmarkStatusEffectEvasionSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewStatusEffectEvasionSystem(world, 12345)
	sys.SetGenre("fantasy")

	// Create 100 entities with various status effects
	entities := make([]*Entity, 100)
	rng := rand.New(rand.NewSource(12345))
	effects := []string{"chilled", "frozen", "haste", "poisoned", "blessed"}

	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		stats := NewStatsComponent()
		stats.Evasion = rng.Float64() * 0.3
		entity.AddComponent(stats)
		entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})
		entity.AddComponent(&StatusEffectComponent{
			EffectType: effects[rng.Intn(len(effects))],
			Magnitude:  rng.Float64(),
			Duration:   5.0,
		})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}

func TestStatusEffectEvasionSystem_AllEffectTypes(t *testing.T) {
	sys := NewStatusEffectEvasionSystem(nil, 12345)

	effectTests := []struct {
		effectType  string
		expectedMod float64
		description string
	}{
		{"chilled", -0.05, "cold slows reactions"},
		{"frozen", -0.30, "near-immobile easy target"},
		{"wet", 0.05, "slippery harder to target"},
		{"haste", 0.15, "faster reflexes"},
		{"blinded", -0.20, "can't see attacks"},
		{"blindness", -0.20, "alternate blind effect"},
		{"poisoned", -0.10, "weakened state"},
		{"stunned", -0.25, "disoriented"},
		{"blessed", 0.10, "divine protection"},
		{"cursed", -0.10, "misfortune"},
		{"regeneration", 0.05, "vitality improves reflexes"},
		{"weakness", -0.05, "weakened state"},
		{"vulnerability", -0.05, "exposed state"},
		{"strength", 0.0, "no evasion effect"},
		{"fortify", 0.0, "defense focused"},
		{"burning", 0.0, "no evasion effect"},
	}

	for _, tt := range effectTests {
		t.Run(tt.effectType, func(t *testing.T) {
			effect := &StatusEffectComponent{
				EffectType: tt.effectType,
				Duration:   5.0,
			}
			mod := sys.effectToEvasionModifier(effect)
			if mod < tt.expectedMod-0.001 || mod > tt.expectedMod+0.001 {
				t.Errorf("%s: modifier = %f, want %f (%s)", tt.effectType, mod, tt.expectedMod, tt.description)
			}
		})
	}
}

func TestStatusEffectEvasionSystem_GetEffectiveEvasion(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectEvasionSystem(world, 12345)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Evasion = 0.20
	entity.AddComponent(stats)
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "haste",
		Duration:   5.0,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	effectiveEvasion := sys.GetEffectiveEvasion(entity)
	// After haste (+15%), should be 0.35
	expected := 0.35
	if effectiveEvasion < expected-0.001 || effectiveEvasion > expected+0.001 {
		t.Errorf("GetEffectiveEvasion = %f, want %f", effectiveEvasion, expected)
	}
}

func TestStatusEffectEvasionSystem_GetEffectiveEvasion_NoStats(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectEvasionSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})

	effectiveEvasion := sys.GetEffectiveEvasion(entity)
	if effectiveEvasion != 0.0 {
		t.Errorf("GetEffectiveEvasion with no stats = %f, want 0.0", effectiveEvasion)
	}
}

func TestStatusEffectEvasionSystem_GenreScaling_AllGenres(t *testing.T) {
	tests := []struct {
		genre          string
		debuffModifier float64
		buffModifier   float64
		expectedDebuff float64
		expectedBuff   float64
	}{
		{"fantasy", -0.10, 0.15, -0.10, 0.15},    // No scaling
		{"scifi", -0.10, 0.15, -0.08, 0.12},      // 0.8x scaling (tech resistance)
		{"horror", -0.10, 0.15, -0.12, 0.15},     // 1.2x debuff scaling only
		{"cyberpunk", -0.10, 0.15, -0.09, 0.135}, // 0.9x scaling (augments)
		{"postapoc", -0.10, 0.15, -0.11, 0.15},   // 1.1x debuff scaling only
		{"unknown", -0.10, 0.15, -0.10, 0.15},    // Default (no scaling)
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys := NewStatusEffectEvasionSystem(nil, 12345)
			sys.SetGenre(tt.genre)

			// Test debuff scaling
			scaledDebuff := sys.applyGenreScaling(tt.debuffModifier)
			if scaledDebuff < tt.expectedDebuff-0.001 || scaledDebuff > tt.expectedDebuff+0.001 {
				t.Errorf("%s debuff: got %f, want %f", tt.genre, scaledDebuff, tt.expectedDebuff)
			}

			// Test buff scaling
			scaledBuff := sys.applyGenreScaling(tt.buffModifier)
			if scaledBuff < tt.expectedBuff-0.001 || scaledBuff > tt.expectedBuff+0.001 {
				t.Errorf("%s buff: got %f, want %f", tt.genre, scaledBuff, tt.expectedBuff)
			}
		})
	}
}
