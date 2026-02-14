package engine

import (
	"testing"
)

func TestNewManaRegenParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewManaRegenParticleSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewManaRegenParticleSystem returned nil")
	}
	if sys.world != world {
		t.Error("world not set")
	}
	if sys.seed != 12345 {
		t.Error("seed not set")
	}
	if sys.rng == nil {
		t.Error("rng not initialized")
	}
	if sys.prevMana == nil {
		t.Error("prevMana map not initialized")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got %q", sys.genreID)
	}
}

func TestManaRegenParticleSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewManaRegenParticleSystem(world, 12345)

	tests := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range tests {
		sys.SetGenre(genre)
		if sys.genreID != genre {
			t.Errorf("SetGenre(%q) failed, got %q", genre, sys.genreID)
		}
	}
}

func TestManaRegenParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewManaRegenParticleSystem(world, 12345)
	ps := NewParticleSystem()

	sys.SetParticleSystem(ps)
	if sys.particleSystem != ps {
		t.Error("particle system not set")
	}
}

func TestManaRegenParticleSystem_Update_NoParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewManaRegenParticleSystem(world, 12345)
	// No particle system set

	entity := world.CreateEntity()
	entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Should not panic
	sys.Update([]*Entity{entity}, 1.0/60.0)
}

func TestManaRegenParticleSystem_Update_NoWorld(t *testing.T) {
	sys := NewManaRegenParticleSystem(nil, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	// Should not panic
	sys.Update([]*Entity{}, 1.0/60.0)
}

func TestManaRegenParticleSystem_Update_NoManaComponent(t *testing.T) {
	world := NewWorld()
	sys := NewManaRegenParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Update multiple times to exceed pulse interval
	for i := 0; i < 100; i++ {
		sys.Update([]*Entity{entity}, 0.02)
	}
	// Should not panic, no particles spawned
}

func TestManaRegenParticleSystem_Update_LowRegenRate(t *testing.T) {
	world := NewWorld()
	sys := NewManaRegenParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	// Regen rate below threshold (minRegenRate = 1.0)
	entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 0.5})
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Update multiple times
	for i := 0; i < 100; i++ {
		sys.Update([]*Entity{entity}, 0.02)
	}
	// Should not spawn particles due to low regen rate
}

func TestManaRegenParticleSystem_Update_ManaFull(t *testing.T) {
	world := NewWorld()
	sys := NewManaRegenParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	// Mana is full - no need to regen
	entity.AddComponent(&ManaComponent{Current: 100, Max: 100, Regen: 5.0})
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Update multiple times
	for i := 0; i < 100; i++ {
		sys.Update([]*Entity{entity}, 0.02)
	}
	// Should not spawn particles when mana is full
}

func TestManaRegenParticleSystem_Update_ActiveRegen(t *testing.T) {
	world := NewWorld()
	sys := NewManaRegenParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	mana := &ManaComponent{Current: 50, Max: 100, Regen: 5.0}
	entity.AddComponent(mana)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// First update - store initial mana
	sys.Update([]*Entity{entity}, 0.02)

	// Simulate mana increasing
	mana.Current = 55

	// Update enough times to exceed pulse interval (1.2s)
	for i := 0; i < 70; i++ {
		sys.Update([]*Entity{entity}, 0.02)
	}
	// Should have spawned particles (no crash)
}

func TestManaRegenParticleSystem_Update_ManaDecreased(t *testing.T) {
	world := NewWorld()
	sys := NewManaRegenParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	mana := &ManaComponent{Current: 80, Max: 100, Regen: 5.0}
	entity.AddComponent(mana)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// First update - store initial mana
	sys.Update([]*Entity{entity}, 0.02)

	// Simulate mana decreasing (spell cast)
	mana.Current = 50

	// Update to exceed pulse interval
	for i := 0; i < 70; i++ {
		sys.Update([]*Entity{entity}, 0.02)
	}
	// Should not spawn particles when mana decreased
}

func TestManaRegenParticleSystem_Update_NoPosition(t *testing.T) {
	world := NewWorld()
	sys := NewManaRegenParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	mana := &ManaComponent{Current: 50, Max: 100, Regen: 5.0}
	entity.AddComponent(mana)
	// No position component

	// First update
	sys.Update([]*Entity{entity}, 0.02)
	mana.Current = 55

	// Update to exceed pulse interval
	for i := 0; i < 70; i++ {
		sys.Update([]*Entity{entity}, 0.02)
	}
	// Should not panic without position
}

func TestManaRegenParticleSystem_getManaParticleType(t *testing.T) {
	world := NewWorld()
	sys := NewManaRegenParticleSystem(world, 12345)

	tests := []struct {
		genre    string
		expected string
	}{
		{"fantasy", "magic"},
		{"scifi", "spark"},
		{"horror", "smoke"},
		{"cyberpunk", "spark"},
		{"postapoc", "dust"},
		{"unknown", "sparkle"},
	}

	for _, tt := range tests {
		sys.SetGenre(tt.genre)
		ptype := sys.getManaParticleType()
		if ptype.String() != tt.expected {
			t.Errorf("genre %q: expected particle type %q, got %q", tt.genre, tt.expected, ptype.String())
		}
	}
}

func TestManaRegenParticleSystem_EntityLostManaComponent(t *testing.T) {
	world := NewWorld()
	sys := NewManaRegenParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	mana := &ManaComponent{Current: 50, Max: 100, Regen: 5.0}
	entity.AddComponent(mana)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// First update - entity has mana
	sys.Update([]*Entity{entity}, 1.3) // Exceed pulse interval
	if _, exists := sys.prevMana[entity.ID]; !exists {
		t.Error("expected prevMana to be tracked after first update")
	}

	// Remove mana component
	entity.RemoveComponent("mana")

	// Second update - entity lost mana
	sys.Update([]*Entity{entity}, 1.3)
	if _, exists := sys.prevMana[entity.ID]; exists {
		t.Error("expected prevMana to be cleaned up after entity lost mana component")
	}
}

func BenchmarkManaRegenParticleSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewManaRegenParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})
		entity.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
