//go:build ignore

package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine/physics/fluids"
)

func TestNewDrowningParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewDrowningParticleSystem(world, 12345)

	if system == nil {
		t.Fatal("NewDrowningParticleSystem returned nil")
	}

	if system.world != world {
		t.Error("World not set correctly")
	}

	if system.seed != 12345 {
		t.Errorf("Seed = %d, want 12345", system.seed)
	}

	if system.rng == nil {
		t.Error("RNG not initialized")
	}

	if system.emitInterval != 0.5 {
		t.Errorf("emitInterval = %f, want 0.5", system.emitInterval)
	}

	if system.baseCount != 6 {
		t.Errorf("baseCount = %d, want 6", system.baseCount)
	}
}

func TestDrowningParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewDrowningParticleSystem(world, 12345)
	ps := NewParticleSystem()

	system.SetParticleSystem(ps)

	if system.particleSystem != ps {
		t.Error("ParticleSystem not set correctly")
	}
}

func TestDrowningParticleSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewDrowningParticleSystem(world, 12345)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		system.SetGenre(genre)
		if system.genreID != genre {
			t.Errorf("Genre = %s, want %s", system.genreID, genre)
		}
	}
}

func TestDrowningParticleSystem_IsDrowning(t *testing.T) {
	world := NewWorld()
	system := NewDrowningParticleSystem(world, 12345)

	tests := []struct {
		name     string
		setup    func(*Entity)
		expected bool
	}{
		{
			name:     "no swimming component",
			setup:    func(e *Entity) {},
			expected: false,
		},
		{
			name: "swimming but not drowning",
			setup: func(e *Entity) {
				e.AddComponent(&fluids.SwimmingComponent{
					IsSwimming:     true,
					Drowning:       false,
					Stamina:        50,
					MaxStamina:     100,
					StaminaDrain:   10,
					StaminaRegen:   20,
					SwimSpeed:      0.8,
					DrowningDamage: 5,
				})
			},
			expected: false,
		},
		{
			name: "actively drowning",
			setup: func(e *Entity) {
				e.AddComponent(&fluids.SwimmingComponent{
					IsSwimming:     true,
					Drowning:       true,
					Stamina:        0,
					MaxStamina:     100,
					StaminaDrain:   10,
					StaminaRegen:   20,
					SwimSpeed:      0.8,
					DrowningDamage: 5,
				})
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: 100, Y: 100})
			tt.setup(entity)

			result := system.isDrowning(entity)
			if result != tt.expected {
				t.Errorf("isDrowning() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDrowningParticleSystem_Update_NoDrowning(t *testing.T) {
	world := NewWorld()
	system := NewDrowningParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)
	system.SetGenre("fantasy")

	// Create non-drowning entity
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&fluids.SwimmingComponent{
		IsSwimming: true,
		Drowning:   false,
		Stamina:    50,
		MaxStamina: 100,
	})

	entities := []*Entity{entity}

	// Run update past emit interval
	system.Update(entities, 1.0)

	// Should not crash and should process normally
}

func TestDrowningParticleSystem_Update_WithDrowning(t *testing.T) {
	world := NewWorld()
	system := NewDrowningParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)
	system.SetGenre("fantasy")

	// Create drowning entity
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&fluids.SwimmingComponent{
		IsSwimming:     true,
		Drowning:       true,
		Stamina:        0,
		MaxStamina:     100,
		DrowningDamage: 5,
	})

	entities := []*Entity{entity}

	// Run update past emit interval
	system.Update(entities, 1.0)

	// Should spawn particles (verify no crash, actual spawn tested via particle system)
}

func TestDrowningParticleSystem_Update_NilChecks(t *testing.T) {
	world := NewWorld()
	system := NewDrowningParticleSystem(world, 12345)

	// No particle system set
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&fluids.SwimmingComponent{Drowning: true})

	entities := []*Entity{entity}

	// Should not panic without particle system
	system.Update(entities, 1.0)

	// Set particle system but nil world
	system.particleSystem = NewParticleSystem()
	system.world = nil

	// Should not panic with nil world
	system.Update(entities, 1.0)
}

func TestDrowningParticleSystem_Update_NoPosition(t *testing.T) {
	world := NewWorld()
	system := NewDrowningParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)
	system.SetGenre("fantasy")

	// Create drowning entity without position
	entity := world.CreateEntity()
	entity.AddComponent(&fluids.SwimmingComponent{Drowning: true})

	entities := []*Entity{entity}

	// Should not panic without position
	system.Update(entities, 1.0)
}

func TestDrowningParticleSystem_Update_EmitInterval(t *testing.T) {
	world := NewWorld()
	system := NewDrowningParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)
	system.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&fluids.SwimmingComponent{Drowning: true})

	entities := []*Entity{entity}

	// First update below interval - should not spawn
	system.timeSinceEmit = 0
	system.Update(entities, 0.3)

	if system.timeSinceEmit != 0.3 {
		t.Errorf("timeSinceEmit = %f after 0.3s, want 0.3", system.timeSinceEmit)
	}

	// Second update to exceed interval
	system.Update(entities, 0.3)

	// Timer should be reset after emit
	if system.timeSinceEmit != 0 {
		t.Errorf("timeSinceEmit = %f after reset, want 0", system.timeSinceEmit)
	}
}

func TestDrowningParticleSystem_GetParticleType(t *testing.T) {
	world := NewWorld()
	system := NewDrowningParticleSystem(world, 12345)

	tests := []struct {
		genre    string
		expected string
	}{
		{"fantasy", "magic"},
		{"scifi", "spark"},
		{"horror", "smoke"},
		{"cyberpunk", "spark"},
		{"postapoc", "debris"},
		{"unknown", "magic"}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			system.SetGenre(tt.genre)
			pType := system.getParticleType()
			if string(pType) != tt.expected {
				t.Errorf("getParticleType() for %s = %s, want %s", tt.genre, pType, tt.expected)
			}
		})
	}
}

func TestDrowningParticleSystem_MultipleDrowningEntities(t *testing.T) {
	world := NewWorld()
	system := NewDrowningParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)
	system.SetGenre("fantasy")

	// Create multiple drowning entities
	for i := 0; i < 5; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 100), Y: 100})
		entity.AddComponent(&fluids.SwimmingComponent{
			IsSwimming:     true,
			Drowning:       true,
			Stamina:        0,
			MaxStamina:     100,
			DrowningDamage: 5,
		})
	}

	entities := world.GetEntities()

	// Should process all entities without issue
	system.Update(entities, 1.0)
}

func BenchmarkDrowningParticleSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewDrowningParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)
	system.SetGenre("fantasy")

	// Create 100 entities, 50% drowning
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 10), Y: 100})
		entity.AddComponent(&fluids.SwimmingComponent{
			IsSwimming: true,
			Drowning:   i%2 == 0,
			Stamina:    float64(i % 100),
			MaxStamina: 100,
		})
	}

	entities := world.GetEntities()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016) // ~60 FPS
	}
}
