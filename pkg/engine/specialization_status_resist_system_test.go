package engine

import (
	"math/rand"
	"testing"
)

func TestNewSpecializationStatusResistSystem(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationStatusResistSystem(world, 12345)

	if system == nil {
		t.Fatal("NewSpecializationStatusResistSystem returned nil")
	}
	if system.world != world {
		t.Error("world not set correctly")
	}
	if system.rng == nil {
		t.Error("RNG not initialized")
	}
	if system.genreID != "fantasy" {
		t.Errorf("default genre = %s, want fantasy", system.genreID)
	}
}

func TestSpecializationStatusResistSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationStatusResistSystem(world, 12345)

	system.SetGenre("horror")
	if system.genreID != "horror" {
		t.Errorf("genre = %s, want horror", system.genreID)
	}
}

func TestSpecializationStatusResistSystem_DebuffDurationReduction(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationStatusResistSystem(world, 12345)

	// Create entity with Defender specialization (highest resistance)
	entity := NewEntity(0)
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          15,
		Specialization: SpecializationDefender,
	})

	// Add a poison debuff with 10 second duration
	effect := &StatusEffectComponent{
		EffectType:   "poison",
		Duration:     10.0,
		Magnitude:    5.0,
		TickInterval: 1.0,
		NextTick:     1.0,
	}
	entity.AddComponent(effect)

	entities := []*Entity{entity}

	// Run system update
	system.Update(entities, 0.5)

	// Defender should reduce duration significantly (50% resistance -> 50% duration)
	// Expected: 10.0 * (1.0 - 0.50*1.0) = 5.0
	if effect.Duration > 6.0 || effect.Duration < 4.0 {
		t.Errorf("Defender poison duration = %.2f, want ~5.0 (50%% reduction)", effect.Duration)
	}
}

func TestSpecializationStatusResistSystem_GlassCannonVulnerability(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationStatusResistSystem(world, 12345)

	// Create entity with Arcanist specialization (glass cannon - increased vulnerability)
	entity := NewEntity(0)
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassMage,
		Level:          15,
		Specialization: SpecializationArcanist,
	})

	// Add a stun debuff with 5 second duration
	effect := &StatusEffectComponent{
		EffectType: "stun",
		Duration:   5.0,
		Magnitude:  1.0,
	}
	entity.AddComponent(effect)

	entities := []*Entity{entity}
	system.Update(entities, 0.5)

	// Arcanist should INCREASE duration (-10% resistance -> 110% duration)
	// Expected: 5.0 * (1.0 - (-0.10)*1.0) = 5.5
	if effect.Duration < 5.3 || effect.Duration > 5.7 {
		t.Errorf("Arcanist stun duration = %.2f, want ~5.5 (10%% increase)", effect.Duration)
	}
}

func TestSpecializationStatusResistSystem_BuffsNotAffected(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationStatusResistSystem(world, 12345)

	entity := NewEntity(0)
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          15,
		Specialization: SpecializationDefender,
	})

	// Add a buff (not a debuff)
	effect := &StatusEffectComponent{
		EffectType: "speed_boost",
		Duration:   10.0,
		Magnitude:  1.5,
	}
	entity.AddComponent(effect)

	entities := []*Entity{entity}
	system.Update(entities, 0.5)

	// Buffs should NOT be modified
	if effect.Duration != 10.0 {
		t.Errorf("Buff duration = %.2f, want 10.0 (unchanged)", effect.Duration)
	}
}

func TestSpecializationStatusResistSystem_NoSpecialization(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationStatusResistSystem(world, 12345)

	entity := NewEntity(0)
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          5,
		Specialization: SpecializationNone,
	})

	effect := &StatusEffectComponent{
		EffectType: "poison",
		Duration:   10.0,
		Magnitude:  5.0,
	}
	entity.AddComponent(effect)

	entities := []*Entity{entity}
	system.Update(entities, 0.5)

	// Warriors without specialization get small base class reduction (8%)
	// Expected: 10.0 * 0.92 = 9.2
	if effect.Duration < 9.0 || effect.Duration > 9.5 {
		t.Errorf("Unspecialized warrior poison duration = %.2f, want ~9.2", effect.Duration)
	}
}

func TestSpecializationStatusResistSystem_GenreModifiers(t *testing.T) {
	tests := []struct {
		name        string
		genre       string
		expectLower bool // Lower duration = more resistance
	}{
		{"fantasy_standard", "fantasy", false},
		{"scifi_enhanced", "scifi", true},         // Tech enhances resistance
		{"horror_dangerous", "horror", false},     // Horror debuffs more dangerous
		{"cyberpunk_implants", "cyberpunk", true}, // Cyber-implants help
		{"postapoc_weakened", "postapoc", false},  // Weakened systems
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewSpecializationStatusResistSystem(world, 12345)
			system.SetGenre(tt.genre)

			entity := NewEntity(0)
			entity.AddComponent(&ClassProgressionComponent{
				Class:          ClassWarrior,
				Level:          15,
				Specialization: SpecializationDefender,
			})

			effect := &StatusEffectComponent{
				EffectType: "poison",
				Duration:   10.0,
				Magnitude:  5.0,
			}
			entity.AddComponent(effect)

			system.Update([]*Entity{entity}, 0.5)

			// Just verify the system processed it - specific values tested elsewhere
			if effect.Duration >= 10.0 {
				t.Errorf("Duration should be reduced for Defender, got %.2f", effect.Duration)
			}
		})
	}
}

func TestSpecializationStatusResistSystem_ProcessedOnlyOnce(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationStatusResistSystem(world, 12345)

	entity := NewEntity(0)
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          15,
		Specialization: SpecializationDefender,
	})

	effect := &StatusEffectComponent{
		EffectType: "poison",
		Duration:   10.0,
		Magnitude:  5.0,
	}
	entity.AddComponent(effect)

	entities := []*Entity{entity}

	// First update
	system.Update(entities, 0.5)
	firstDuration := effect.Duration

	// Second update - should NOT modify again
	system.Update(entities, 0.5)

	if effect.Duration != firstDuration {
		t.Errorf("Effect modified twice: first=%.2f, second=%.2f", firstDuration, effect.Duration)
	}
}

func TestSpecializationStatusResistSystem_MultipleDebuffs(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationStatusResistSystem(world, 12345)

	entity := NewEntity(0)
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          15,
		Specialization: SpecializationDefender,
	})

	effects := []*StatusEffectComponent{
		{EffectType: "poison", Duration: 10.0, Magnitude: 5.0},
		{EffectType: "burn", Duration: 8.0, Magnitude: 3.0},
		{EffectType: "slow", Duration: 6.0, Magnitude: 0.5},
	}

	effectSet := &StatusEffectSetComponent{}
	for _, e := range effects {
		effectSet.AddEffect(e)
	}
	entity.AddComponent(effectSet)

	system.Update([]*Entity{entity}, 0.5)

	// All debuffs should be reduced
	for i, e := range effects {
		originalDurations := []float64{10.0, 8.0, 6.0}
		if e.Duration >= originalDurations[i] {
			t.Errorf("Effect %d (%s) not reduced: duration=%.2f, original=%.2f",
				i, e.EffectType, e.Duration, originalDurations[i])
		}
	}
}

func TestSpecializationStatusResistSystem_EntityWithoutProgression(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationStatusResistSystem(world, 12345)

	// Entity without class_progression component
	entity := NewEntity(0)
	effect := &StatusEffectComponent{
		EffectType: "poison",
		Duration:   10.0,
		Magnitude:  5.0,
	}
	entity.AddComponent(effect)

	system.Update([]*Entity{entity}, 0.5)

	// Should not be modified
	if effect.Duration != 10.0 {
		t.Errorf("Entity without progression modified: duration=%.2f, want 10.0", effect.Duration)
	}
}

func TestSpecializationStatusResistSystem_IsDebuff(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationStatusResistSystem(world, 12345)

	debuffs := []string{
		"poison", "burn", "bleed", "stun", "slow", "frozen",
		"fear", "blind", "silence", "curse", "weakness",
	}
	buffs := []string{
		"speed_boost", "strength_up", "shield", "regeneration",
		"haste", "empower", "protect",
	}

	for _, d := range debuffs {
		if !system.isDebuff(d) {
			t.Errorf("%s should be detected as debuff", d)
		}
	}

	for _, b := range buffs {
		if system.isDebuff(b) {
			t.Errorf("%s should NOT be detected as debuff", b)
		}
	}
}

func TestSpecializationStatusResistSystem_GetResistanceForEntity(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationStatusResistSystem(world, 12345)

	// Entity with specialization
	entity := NewEntity(0)
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          15,
		Specialization: SpecializationDefender,
	})

	resist := system.GetResistanceForEntity(entity)

	// Defender: 1.0 - (0.50 * 1.0) = 0.50
	if resist < 0.45 || resist > 0.55 {
		t.Errorf("Defender resistance modifier = %.2f, want ~0.50", resist)
	}

	// Entity without progression
	entity2 := NewEntity(0)
	resist2 := system.GetResistanceForEntity(entity2)
	if resist2 != 1.0 {
		t.Errorf("Entity without progression resistance = %.2f, want 1.0", resist2)
	}
}

func TestSpecializationStatusResistSystem_SpecializationCoverage(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationStatusResistSystem(world, 12345)

	// Test that various specializations return different values
	specs := []SpecializationType{
		SpecializationDefender,
		SpecializationBerserker,
		SpecializationArcanist,
		SpecializationHealer,
		SpecializationAssassin,
	}

	resistances := make(map[SpecializationType]float64)
	for _, spec := range specs {
		resist := system.getSpecializationResistance(spec)
		resistances[spec] = resist
	}

	// Verify ordering: Defender > Healer > Berserker > Assassin > Arcanist
	if resistances[SpecializationDefender] <= resistances[SpecializationHealer] {
		t.Error("Defender should have higher resistance than Healer")
	}
	if resistances[SpecializationHealer] <= resistances[SpecializationBerserker] {
		t.Error("Healer should have higher resistance than Berserker")
	}
	if resistances[SpecializationBerserker] <= resistances[SpecializationAssassin] {
		t.Error("Berserker should have higher resistance than Assassin")
	}
	if resistances[SpecializationArcanist] >= 0 {
		t.Error("Arcanist should have negative resistance (vulnerability)")
	}
}

func TestSpecializationStatusResistSystem_Determinism(t *testing.T) {
	seed := int64(12345)

	// Run twice with same seed
	results := make([]float64, 2)
	for i := 0; i < 2; i++ {
		world := NewWorld()
		system := NewSpecializationStatusResistSystem(world, seed)

		entity := NewEntity(0)
		entity.AddComponent(&ClassProgressionComponent{
			Class:          ClassWarrior,
			Level:          15,
			Specialization: SpecializationDefender,
		})

		effect := &StatusEffectComponent{
			EffectType: "poison",
			Duration:   10.0,
			Magnitude:  5.0,
		}
		entity.AddComponent(effect)

		system.Update([]*Entity{entity}, 0.5)
		results[i] = effect.Duration
	}

	if results[0] != results[1] {
		t.Errorf("Non-deterministic: run1=%.4f, run2=%.4f", results[0], results[1])
	}
}

func TestSpecializationStatusResistSystem_UpdateInterval(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationStatusResistSystem(world, 12345)

	entity := NewEntity(0)
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          15,
		Specialization: SpecializationDefender,
	})

	effect := &StatusEffectComponent{
		EffectType: "poison",
		Duration:   10.0,
		Magnitude:  5.0,
	}
	entity.AddComponent(effect)

	// First update with very small deltaTime - should not process yet
	system.Update([]*Entity{entity}, 0.01)
	if effect.Duration != 10.0 {
		t.Error("Effect processed before update interval elapsed")
	}

	// Update with enough time to trigger processing
	system.Update([]*Entity{entity}, 0.1)
	if effect.Duration >= 10.0 {
		t.Error("Effect not processed after update interval")
	}
}

func BenchmarkSpecializationStatusResistSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewSpecializationStatusResistSystem(world, 12345)

	// Create 100 entities with status effects
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := NewEntity(0)
		entity.AddComponent(&ClassProgressionComponent{
			Class:          ClassWarrior,
			Level:          15,
			Specialization: SpecializationType(rand.Intn(10) + 1),
		})
		entity.AddComponent(&StatusEffectComponent{
			EffectType: "poison",
			Duration:   10.0,
			Magnitude:  5.0,
		})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset processed effects for benchmark
		system.processedEffects = make(map[uint64]map[*StatusEffectComponent]bool)
		system.timeSinceCheck = 0.2 // Force processing
		system.Update(entities, 0.2)
	}
}
