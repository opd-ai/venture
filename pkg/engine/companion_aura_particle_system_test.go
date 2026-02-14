package engine

import (
	"testing"
)

func TestNewCompanionAuraParticleSystem(t *testing.T) {
	world := NewWorld()
	seed := int64(12345)

	system := NewCompanionAuraParticleSystem(world, seed)

	if system == nil {
		t.Fatal("NewCompanionAuraParticleSystem returned nil")
	}

	if system.world != world {
		t.Error("system.world not set correctly")
	}

	if system.seed != seed {
		t.Errorf("system.seed = %d, want %d", system.seed, seed)
	}

	if system.rng == nil {
		t.Error("system.rng is nil")
	}

	if system.pulseInterval <= 0 {
		t.Error("pulseInterval should be positive")
	}

	if system.baseParticleCount <= 0 {
		t.Error("baseParticleCount should be positive")
	}

	if system.auraDistanceMax <= 0 {
		t.Error("auraDistanceMax should be positive")
	}
}

func TestCompanionAuraParticleSystem_SetGenre(t *testing.T) {
	system := NewCompanionAuraParticleSystem(NewWorld(), 12345)

	tests := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range tests {
		system.SetGenre(genre)
		if system.genreID != genre {
			t.Errorf("SetGenre(%s) failed, got %s", genre, system.genreID)
		}
	}
}

func TestCompanionAuraParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewCompanionAuraParticleSystem(world, 12345)
	ps := NewParticleSystem()

	system.SetParticleSystem(ps)

	if system.particleSystem != ps {
		t.Error("SetParticleSystem did not set particle system")
	}
}

func TestCompanionAuraParticleSystem_UpdateNoParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewCompanionAuraParticleSystem(world, 12345)
	// Don't set particle system

	entities := []*Entity{world.CreateEntity()}

	// Should not panic
	system.Update(entities, 0.016)
}

func TestCompanionAuraParticleSystem_UpdateNoCompanions(t *testing.T) {
	world := NewWorld()
	system := NewCompanionAuraParticleSystem(world, 12345)
	system.SetParticleSystem(NewParticleSystem())

	// Create entity without companion component
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	entities := []*Entity{entity}

	// Should not panic or spawn particles
	system.timeSinceEmit = system.pulseInterval // Force pulse
	system.Update(entities, 0.016)
}

func TestCompanionAuraParticleSystem_UpdateCompanionNoPerks(t *testing.T) {
	world := NewWorld()
	system := NewCompanionAuraParticleSystem(world, 12345)
	system.SetParticleSystem(NewParticleSystem())

	// Create owner
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})
	owner.AddComponent(NewStubInput())

	// Create companion without perks
	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 110, Y: 100})
	companion.AddComponent(&CompanionComponent{
		OwnerID:      owner.ID,
		BondingPerks: []BondingPerk{}, // No perks
	})

	entities := []*Entity{companion}
	system.timeSinceEmit = system.pulseInterval
	system.Update(entities, 0.016)

	// No error = success (particles not spawned without perks)
}

func TestCompanionAuraParticleSystem_UpdateCompanionWithPerks(t *testing.T) {
	world := NewWorld()
	system := NewCompanionAuraParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)
	system.SetGenre("fantasy")

	// Create owner
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})
	owner.AddComponent(NewStubInput())

	// Create companion with perks near owner
	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 120, Y: 100}) // Within range
	companion.AddComponent(&CompanionComponent{
		OwnerID:      owner.ID,
		BondingPerks: []BondingPerk{PerkExtraHealth, PerkExtraDamage},
	})

	entities := []*Entity{companion}
	system.timeSinceEmit = system.pulseInterval
	system.Update(entities, 0.016)

	// Verify time reset
	if system.timeSinceEmit != 0 {
		t.Error("timeSinceEmit should reset after pulse")
	}
}

func TestCompanionAuraParticleSystem_UpdateCompanionTooFar(t *testing.T) {
	world := NewWorld()
	system := NewCompanionAuraParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)

	// Create owner
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})
	owner.AddComponent(NewStubInput())

	// Create companion far from owner
	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 1000, Y: 1000}) // Far away
	companion.AddComponent(&CompanionComponent{
		OwnerID:      owner.ID,
		BondingPerks: []BondingPerk{PerkExtraHealth},
	})

	entities := []*Entity{companion}
	system.timeSinceEmit = system.pulseInterval
	system.Update(entities, 0.016)

	// Should not spawn particles when too far
}

func TestCompanionAuraParticleSystem_GetDominantPerkType(t *testing.T) {
	system := NewCompanionAuraParticleSystem(NewWorld(), 12345)

	tests := []struct {
		name     string
		perks    []BondingPerk
		expected BondingPerk
	}{
		{"empty", []BondingPerk{}, PerkNone},
		{"single health", []BondingPerk{PerkExtraHealth}, PerkExtraHealth},
		{"single damage", []BondingPerk{PerkExtraDamage}, PerkExtraDamage},
		{"multiple low tier", []BondingPerk{PerkExtraHealth, PerkExtraDamage}, PerkExtraDamage},
		{"high tier wins", []BondingPerk{PerkExtraHealth, PerkAutoRevive}, PerkAutoRevive},
		{"all perks", []BondingPerk{
			PerkExtraHealth, PerkExtraDamage, PerkFasterLearning,
			PerkLoyalGuard, PerkSharedExperience, PerkAutoRevive,
		}, PerkAutoRevive},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := system.getDominantPerkType(tt.perks)
			if result != tt.expected {
				t.Errorf("getDominantPerkType(%v) = %v, want %v", tt.perks, result, tt.expected)
			}
		})
	}
}

func TestCompanionAuraParticleSystem_GetParticleTypeForPerk_AllGenres(t *testing.T) {
	system := NewCompanionAuraParticleSystem(NewWorld(), 12345)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc", "unknown"}
	perks := []BondingPerk{
		PerkExtraHealth, PerkExtraDamage, PerkFasterLearning,
		PerkLoyalGuard, PerkSharedExperience, PerkAutoRevive, PerkNone,
	}

	for _, genre := range genres {
		system.SetGenre(genre)
		for _, perk := range perks {
			// Should not panic and return valid particle type
			pType := system.getParticleTypeForPerk(perk)
			if pType < 0 {
				t.Errorf("genre=%s, perk=%v returned invalid particle type", genre, perk)
			}
		}
	}
}

func TestCompanionAuraParticleSystem_GetCompanionComponent(t *testing.T) {
	system := NewCompanionAuraParticleSystem(NewWorld(), 12345)

	tests := []struct {
		name      string
		setup     func() *Entity
		expectNil bool
	}{
		{
			name: "entity with companion",
			setup: func() *Entity {
				e := &Entity{ID: 1, Components: make(map[string]Component)}
				e.AddComponent(&CompanionComponent{OwnerID: 2})
				return e
			},
			expectNil: false,
		},
		{
			name: "entity without companion",
			setup: func() *Entity {
				e := &Entity{ID: 1, Components: make(map[string]Component)}
				e.AddComponent(&PositionComponent{})
				return e
			},
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := tt.setup()
			result := system.getCompanionComponent(entity)
			if (result == nil) != tt.expectNil {
				t.Errorf("getCompanionComponent() nil=%v, want nil=%v", result == nil, tt.expectNil)
			}
		})
	}
}

func TestCompanionAuraParticleSystem_PulseInterval(t *testing.T) {
	world := NewWorld()
	system := NewCompanionAuraParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)

	// Create owner
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Create companion with perk
	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 110, Y: 100})
	companion.AddComponent(&CompanionComponent{
		OwnerID:      owner.ID,
		BondingPerks: []BondingPerk{PerkExtraHealth},
	})

	entities := []*Entity{companion}

	// First update - not enough time passed
	system.timeSinceEmit = 0.5 // Less than pulseInterval
	system.Update(entities, 0.016)

	// timeSinceEmit should increase
	if system.timeSinceEmit <= 0.5 {
		t.Error("timeSinceEmit should have increased")
	}

	// Force time past interval
	system.timeSinceEmit = system.pulseInterval
	system.Update(entities, 0.016)

	// Should have reset
	if system.timeSinceEmit != 0 {
		t.Errorf("timeSinceEmit should reset after pulse, got %f", system.timeSinceEmit)
	}
}

func TestCompanionAuraParticleSystem_NoOwnerEntity(t *testing.T) {
	world := NewWorld()
	system := NewCompanionAuraParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)

	// Create companion with invalid owner ID
	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 100, Y: 100})
	companion.AddComponent(&CompanionComponent{
		OwnerID:      999999, // Non-existent owner
		BondingPerks: []BondingPerk{PerkExtraHealth},
	})

	entities := []*Entity{companion}
	system.timeSinceEmit = system.pulseInterval

	// Should not panic when owner doesn't exist
	system.Update(entities, 0.016)
}

func TestCompanionAuraParticleSystem_OwnerNoPosition(t *testing.T) {
	world := NewWorld()
	system := NewCompanionAuraParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)

	// Create owner without position
	owner := world.CreateEntity()
	// No position component

	// Create companion
	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 100, Y: 100})
	companion.AddComponent(&CompanionComponent{
		OwnerID:      owner.ID,
		BondingPerks: []BondingPerk{PerkExtraHealth},
	})

	entities := []*Entity{companion}
	system.timeSinceEmit = system.pulseInterval

	// Should not panic when owner has no position
	system.Update(entities, 0.016)
}

func TestCompanionAuraParticleSystem_CompanionNoPosition(t *testing.T) {
	world := NewWorld()
	system := NewCompanionAuraParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)

	// Create owner
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Create companion without position
	companion := world.CreateEntity()
	// No position component
	companion.AddComponent(&CompanionComponent{
		OwnerID:      owner.ID,
		BondingPerks: []BondingPerk{PerkExtraHealth},
	})

	entities := []*Entity{companion}
	system.timeSinceEmit = system.pulseInterval

	// Should not panic when companion has no position
	system.Update(entities, 0.016)
}

func TestCompanionAuraParticleSystem_MultipleCompanions(t *testing.T) {
	world := NewWorld()
	system := NewCompanionAuraParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)
	system.SetGenre("fantasy")

	// Create owner
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Create multiple companions
	companions := make([]*Entity, 3)
	for i := 0; i < 3; i++ {
		c := world.CreateEntity()
		c.AddComponent(&PositionComponent{X: float64(110 + i*10), Y: 100})
		c.AddComponent(&CompanionComponent{
			OwnerID:      owner.ID,
			BondingPerks: []BondingPerk{PerkExtraHealth},
		})
		companions[i] = c
	}

	system.timeSinceEmit = system.pulseInterval
	system.Update(companions, 0.016)

	// Should process all companions without error
	if system.timeSinceEmit != 0 {
		t.Error("timeSinceEmit should reset after processing")
	}
}

func BenchmarkCompanionAuraParticleSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewCompanionAuraParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)
	system.SetGenre("fantasy")

	// Create owner
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Create 10 companions with perks
	entities := make([]*Entity, 10)
	for i := 0; i < 10; i++ {
		c := world.CreateEntity()
		c.AddComponent(&PositionComponent{X: float64(110 + i*5), Y: 100})
		c.AddComponent(&CompanionComponent{
			OwnerID:      owner.ID,
			BondingPerks: []BondingPerk{PerkExtraHealth, PerkExtraDamage},
		})
		entities[i] = c
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.timeSinceEmit = system.pulseInterval
		system.Update(entities, 0.016)
	}
}
