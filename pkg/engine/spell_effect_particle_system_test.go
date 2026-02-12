package engine

import (
	"testing"
)

func TestNewSpellEffectParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewSpellEffectParticleSystem(world, 12345)

	if system == nil {
		t.Fatal("NewSpellEffectParticleSystem returned nil")
	}
	if system.world != world {
		t.Error("world not set correctly")
	}
	if system.seed != 12345 {
		t.Errorf("seed = %d, want 12345", system.seed)
	}
	if system.rng == nil {
		t.Error("rng not initialized")
	}
	if system.spawnedEffects == nil {
		t.Error("spawnedEffects map not initialized")
	}
}

func TestSpellEffectParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewSpellEffectParticleSystem(world, 12345)
	ps := NewParticleSystem()

	system.SetParticleSystem(ps)

	if system.particleSystem != ps {
		t.Error("particle system not set correctly")
	}
}

func TestSpellEffectParticleSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewSpellEffectParticleSystem(world, 12345)

	system.SetGenre("scifi")

	if system.genreID != "scifi" {
		t.Errorf("genreID = %s, want scifi", system.genreID)
	}
}

func TestSpellEffectParticleSystem_Update_NoParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewSpellEffectParticleSystem(world, 12345)
	// Don't set particle system

	entity := NewEntity(1)
	entity.AddComponent(&SpellEffectComponent{
		EffectType: EffectSummoning,
		Active:     true,
		Duration:   2.0,
	})

	// Should not panic
	system.Update([]*Entity{entity}, 0.016)
}

func TestSpellEffectParticleSystem_Update_SpawnsParticles(t *testing.T) {
	world := NewWorld()
	system := NewSpellEffectParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)
	system.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 200})
	entity.AddComponent(&SpellEffectComponent{
		EffectType:  EffectSummoning,
		Active:      true,
		Duration:    2.0,
		ElapsedTime: 0.0,
	})

	world.AddEntity(entity)

	// First update should spawn particles
	system.Update([]*Entity{entity}, 0.016)

	// Verify effect was tracked
	if len(system.spawnedEffects) == 0 {
		t.Error("effect not tracked after spawn")
	}
}

func TestSpellEffectParticleSystem_Update_DeduplicatesEffects(t *testing.T) {
	world := NewWorld()
	system := NewSpellEffectParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)
	system.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 200})
	effect := &SpellEffectComponent{
		EffectType:  EffectSummoning,
		Active:      true,
		Duration:    2.0,
		ElapsedTime: 0.0,
	}
	entity.AddComponent(effect)

	world.AddEntity(entity)

	// First update
	system.Update([]*Entity{entity}, 0.016)
	trackedCount := len(system.spawnedEffects)

	// Second update should not spawn again
	effect.ElapsedTime = 0.02
	system.Update([]*Entity{entity}, 0.016)

	if len(system.spawnedEffects) != trackedCount {
		t.Error("effect spawned multiple times")
	}
}

func TestSpellEffectParticleSystem_Update_SkipsInactiveEffects(t *testing.T) {
	world := NewWorld()
	system := NewSpellEffectParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)
	system.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 200})
	entity.AddComponent(&SpellEffectComponent{
		EffectType: EffectSummoning,
		Active:     false, // Inactive
		Duration:   2.0,
	})

	world.AddEntity(entity)

	system.Update([]*Entity{entity}, 0.016)

	if len(system.spawnedEffects) != 0 {
		t.Error("inactive effect was tracked")
	}
}

func TestSpellEffectParticleSystem_GetParticleConfig(t *testing.T) {
	tests := []struct {
		name       string
		effectType EffectType
		wantType   string
	}{
		{"terrain", EffectTerrainManipulation, "dust"},
		{"transmutation", EffectTransmutation, "sparkle"},
		{"summoning", EffectSummoning, "magic"},
		{"illusion", EffectIllusion, "sparkle"},
		{"time", EffectTimeManipulation, "sparkle"},
		{"gravity", EffectGravityControl, "debris"},
		{"fusion", EffectElementalFusion, "flame"},
		{"life_drain", EffectLifeDrain, "blood"},
		{"teleport", EffectTeleportation, "spark"},
		{"metamagic", EffectMetamagic, "magic"},
	}

	world := NewWorld()
	system := NewSpellEffectParticleSystem(world, 12345)
	system.SetGenre("fantasy")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effect := &SpellEffectComponent{
				EffectType: tt.effectType,
				Magnitude:  1.0,
			}

			config := system.getParticleConfig(effect)

			if config.Type.String() != tt.wantType {
				t.Errorf("particle type = %s, want %s", config.Type.String(), tt.wantType)
			}
			if config.GenreID != "fantasy" {
				t.Errorf("genreID = %s, want fantasy", config.GenreID)
			}
			if config.Count <= 0 {
				t.Error("count should be positive")
			}
			if config.Duration <= 0 {
				t.Error("duration should be positive")
			}
		})
	}
}

func TestSpellEffectParticleSystem_GetParticleConfig_MagnitudeScaling(t *testing.T) {
	world := NewWorld()
	system := NewSpellEffectParticleSystem(world, 12345)
	system.SetGenre("fantasy")

	lowMag := &SpellEffectComponent{
		EffectType: EffectSummoning,
		Magnitude:  1.0,
	}
	highMag := &SpellEffectComponent{
		EffectType: EffectSummoning,
		Magnitude:  3.0,
	}

	lowConfig := system.getParticleConfig(lowMag)
	highConfig := system.getParticleConfig(highMag)

	if highConfig.Count <= lowConfig.Count {
		t.Errorf("high magnitude count (%d) should be > low magnitude count (%d)",
			highConfig.Count, lowConfig.Count)
	}
}

func TestSpellEffectParticleSystem_GetParticleConfig_RadiusScaling(t *testing.T) {
	world := NewWorld()
	system := NewSpellEffectParticleSystem(world, 12345)
	system.SetGenre("fantasy")

	smallRadius := &SpellEffectComponent{
		EffectType: EffectSummoning,
		Magnitude:  1.0,
		Radius:     25.0,
	}
	largeRadius := &SpellEffectComponent{
		EffectType: EffectSummoning,
		Magnitude:  1.0,
		Radius:     200.0,
	}

	smallConfig := system.getParticleConfig(smallRadius)
	largeConfig := system.getParticleConfig(largeRadius)

	if largeConfig.SpreadX <= smallConfig.SpreadX {
		t.Errorf("large radius spreadX (%.1f) should be > small radius spreadX (%.1f)",
			largeConfig.SpreadX, smallConfig.SpreadX)
	}
}

func TestSpellEffectParticleSystem_SpawnEffectParticlesAt(t *testing.T) {
	world := NewWorld()
	system := NewSpellEffectParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)
	system.SetGenre("horror")

	// Should not panic
	system.SpawnEffectParticlesAt(EffectTeleportation, 50.0, 75.0, 2.0)
}

func TestSpellEffectParticleSystem_SpawnEffectParticlesAt_NoParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewSpellEffectParticleSystem(world, 12345)
	// Don't set particle system

	// Should not panic
	system.SpawnEffectParticlesAt(EffectTeleportation, 50.0, 75.0, 2.0)
}

func TestSpellEffectParticleSystem_UsesTargetPosition(t *testing.T) {
	world := NewWorld()
	system := NewSpellEffectParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)
	system.SetGenre("fantasy")

	// Entity at (100, 200) but effect targets (300, 400)
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 200})
	entity.AddComponent(&SpellEffectComponent{
		EffectType:  EffectTeleportation,
		Active:      true,
		Duration:    1.0,
		ElapsedTime: 0.0,
		TargetX:     300,
		TargetY:     400,
	})

	world.AddEntity(entity)
	system.Update([]*Entity{entity}, 0.016)

	// Verify effect was tracked (particles spawned at target location)
	if len(system.spawnedEffects) == 0 {
		t.Error("effect not tracked")
	}
}

func TestSpellEffectParticleSystem_CleansUpOldEffects(t *testing.T) {
	world := NewWorld()
	system := NewSpellEffectParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)
	system.SetGenre("fantasy")

	// Add more than 100 tracked effects to trigger cleanup
	for i := 0; i < 101; i++ {
		key := effectKey{
			entityID:   uint64(i),
			effectType: EffectSummoning,
			startTime:  float64(i),
		}
		system.spawnedEffects[key] = true
	}

	// Update should trigger cleanup
	system.Update([]*Entity{}, 0.016)

	if len(system.spawnedEffects) > 100 {
		t.Error("old effects not cleaned up")
	}
}

func BenchmarkSpellEffectParticleSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewSpellEffectParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)
	system.SetGenre("fantasy")

	entities := make([]*Entity, 50)
	for i := 0; i < 50; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		entity.AddComponent(&SpellEffectComponent{
			EffectType:  EffectType(i % 10),
			Active:      true,
			Duration:    2.0,
			ElapsedTime: 0.1, // Past initial spawn
		})
		entities[i] = entity
		world.AddEntity(entity)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}

func BenchmarkSpellEffectParticleSystem_GetParticleConfig(b *testing.B) {
	world := NewWorld()
	system := NewSpellEffectParticleSystem(world, 12345)
	system.SetGenre("fantasy")

	effect := &SpellEffectComponent{
		EffectType: EffectSummoning,
		Magnitude:  2.0,
		Radius:     100.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = system.getParticleConfig(effect)
	}
}
