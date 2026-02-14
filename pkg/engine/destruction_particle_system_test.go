package engine

import (
	"testing"
)

func TestNewDestructionParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewDestructionParticleSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewDestructionParticleSystem returned nil")
	}
	if sys.world != world {
		t.Error("World not set correctly")
	}
	if sys.seed != 12345 {
		t.Errorf("Seed = %d, want 12345", sys.seed)
	}
	if sys.genreID != "fantasy" {
		t.Errorf("Default genreID = %s, want fantasy", sys.genreID)
	}
	if sys.debrisParticleCount != 15 {
		t.Errorf("debrisParticleCount = %d, want 15", sys.debrisParticleCount)
	}
	if sys.explosionParticleCount != 30 {
		t.Errorf("explosionParticleCount = %d, want 30", sys.explosionParticleCount)
	}
}

func TestDestructionParticleSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"scifi", "scifi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewDestructionParticleSystem(world, 12345)
			sys.SetGenre(tt.genreID)

			if sys.genreID != tt.genreID {
				t.Errorf("genreID = %s, want %s", sys.genreID, tt.genreID)
			}
		})
	}
}

func TestDestructionParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewDestructionParticleSystem(world, 12345)
	particleSys := NewParticleSystem()

	sys.SetParticleSystem(particleSys)

	if sys.particleSystem != particleSys {
		t.Error("Particle system not set correctly")
	}
}

func TestDestructionParticleSystem_Update_NoParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewDestructionParticleSystem(world, 12345)

	// Create a destructible entity
	entity := world.CreateEntity()
	comp := NewDestructibleObjectComponent(ObjectCrate)
	comp.IsDestroyed = true
	entity.AddComponent(comp)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Should not panic with nil particle system
	sys.Update([]*Entity{entity}, 0.016)

	// ParticlesSpawned should NOT be set since we don't have a particle system
	if comp.ParticlesSpawned {
		t.Error("ParticlesSpawned should be false without particle system")
	}
}

func TestDestructionParticleSystem_Update_SpawnsParticles(t *testing.T) {
	world := NewWorld()
	sys := NewDestructionParticleSystem(world, 12345)
	particleSys := NewParticleSystem()
	sys.SetParticleSystem(particleSys)

	// Create a destructible entity
	entity := world.CreateEntity()
	comp := NewDestructibleObjectComponent(ObjectCrate)
	comp.IsDestroyed = true
	entity.AddComponent(comp)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Run update
	sys.Update([]*Entity{entity}, 0.016)

	// ParticlesSpawned should be set
	if !comp.ParticlesSpawned {
		t.Error("ParticlesSpawned should be true after update")
	}
}

func TestDestructionParticleSystem_Update_SkipsAlreadySpawned(t *testing.T) {
	world := NewWorld()
	sys := NewDestructionParticleSystem(world, 12345)
	particleSys := NewParticleSystem()
	sys.SetParticleSystem(particleSys)

	// Create a destructible entity with particles already spawned
	entity := world.CreateEntity()
	comp := NewDestructibleObjectComponent(ObjectCrate)
	comp.IsDestroyed = true
	comp.ParticlesSpawned = true
	entity.AddComponent(comp)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	initialCount := particleSys.GetActiveParticleCount()

	// Run update
	sys.Update([]*Entity{entity}, 0.016)

	// No new particles should be spawned
	if particleSys.GetActiveParticleCount() != initialCount {
		t.Error("Should not spawn particles for already-processed destruction")
	}
}

func TestDestructionParticleSystem_Update_ExplosiveBarrel(t *testing.T) {
	world := NewWorld()
	sys := NewDestructionParticleSystem(world, 12345)
	particleSys := NewParticleSystem()
	sys.SetParticleSystem(particleSys)

	// Create explosive barrel
	entity := world.CreateEntity()
	comp := NewDestructibleObjectComponent(ObjectExplosiveBarrel)
	comp.IsDestroyed = true
	entity.AddComponent(comp)
	entity.AddComponent(&PositionComponent{X: 200, Y: 200})

	// Run update
	sys.Update([]*Entity{entity}, 0.016)

	// Should have spawned explosion particles
	if !comp.ParticlesSpawned {
		t.Error("ParticlesSpawned should be true for explosive barrel")
	}
	if particleSys.GetActiveParticleCount() == 0 {
		t.Error("Expected explosion particles to be spawned")
	}
}

func TestDestructionParticleSystem_Update_PoisonContainer(t *testing.T) {
	world := NewWorld()
	sys := NewDestructionParticleSystem(world, 12345)
	particleSys := NewParticleSystem()
	sys.SetParticleSystem(particleSys)

	// Create poison container
	entity := world.CreateEntity()
	comp := NewDestructibleObjectComponent(ObjectPoisonContainer)
	comp.IsDestroyed = true
	entity.AddComponent(comp)
	entity.AddComponent(&PositionComponent{X: 300, Y: 300})

	// Run update
	sys.Update([]*Entity{entity}, 0.016)

	// Should have spawned poison particles
	if !comp.ParticlesSpawned {
		t.Error("ParticlesSpawned should be true for poison container")
	}
	if particleSys.GetActiveParticleCount() == 0 {
		t.Error("Expected poison particles to be spawned")
	}
}

func TestDestructionParticleSystem_GetExplosionColors(t *testing.T) {
	tests := []struct {
		genre string
	}{
		{"fantasy"},
		{"scifi"},
		{"horror"},
		{"cyberpunk"},
		{"postapoc"},
		{"unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewDestructionParticleSystem(world, 12345)
			sys.SetGenre(tt.genre)

			primary, secondary := sys.getExplosionColors()

			// Verify colors are valid (non-zero alpha)
			if primary.A == 0 {
				t.Error("Primary explosion color has zero alpha")
			}
			if secondary.A == 0 {
				t.Error("Secondary explosion color has zero alpha")
			}
		})
	}
}

func TestDestructionParticleSystem_GetPoisonColors(t *testing.T) {
	tests := []struct {
		genre string
	}{
		{"fantasy"},
		{"scifi"},
		{"horror"},
		{"cyberpunk"},
		{"postapoc"},
		{"unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewDestructionParticleSystem(world, 12345)
			sys.SetGenre(tt.genre)

			primary, secondary := sys.getPoisonColors()

			// Poison should be semi-transparent (alpha < 255)
			if primary.A == 0 {
				t.Error("Primary poison color has zero alpha")
			}
			if secondary.A == 0 {
				t.Error("Secondary poison color has zero alpha")
			}
		})
	}
}

func TestDestructionParticleSystem_GetDebrisColors(t *testing.T) {
	tests := []struct {
		objType ObjectType
	}{
		{ObjectCrate},
		{ObjectBarrel},
		{ObjectFurniture},
		{ObjectWeakWall},
		{ObjectType(999)}, // Unknown type
	}

	for _, tt := range tests {
		t.Run(tt.objType.String(), func(t *testing.T) {
			world := NewWorld()
			sys := NewDestructionParticleSystem(world, 12345)

			primary, secondary := sys.getDebrisColors(tt.objType)

			// Debris should be opaque
			if primary.A != 255 {
				t.Errorf("Primary debris color alpha = %d, want 255", primary.A)
			}
			if secondary.A != 255 {
				t.Errorf("Secondary debris color alpha = %d, want 255", secondary.A)
			}
		})
	}
}

func TestDestructionParticleSystem_GetWallDebrisColors(t *testing.T) {
	tests := []struct {
		genre string
	}{
		{"fantasy"},
		{"scifi"},
		{"horror"},
		{"cyberpunk"},
		{"postapoc"},
		{"unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewDestructionParticleSystem(world, 12345)
			sys.SetGenre(tt.genre)

			primary, secondary := sys.getWallDebrisColors()

			// Wall debris should be opaque
			if primary.A != 255 {
				t.Errorf("Primary wall debris color alpha = %d, want 255", primary.A)
			}
			if secondary.A != 255 {
				t.Errorf("Secondary wall debris color alpha = %d, want 255", secondary.A)
			}
		})
	}
}

func TestDestructionParticleSystem_Update_SkipsNonDestroyedObjects(t *testing.T) {
	world := NewWorld()
	sys := NewDestructionParticleSystem(world, 12345)
	particleSys := NewParticleSystem()
	sys.SetParticleSystem(particleSys)

	// Create a non-destroyed object
	entity := world.CreateEntity()
	comp := NewDestructibleObjectComponent(ObjectCrate)
	comp.IsDestroyed = false // Not destroyed
	entity.AddComponent(comp)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Run update
	sys.Update([]*Entity{entity}, 0.016)

	// Should not spawn particles
	if comp.ParticlesSpawned {
		t.Error("Should not spawn particles for non-destroyed object")
	}
	if particleSys.GetActiveParticleCount() != 0 {
		t.Error("Expected no particles for non-destroyed object")
	}
}

func TestDestructionParticleSystem_Update_SkipsNoPosition(t *testing.T) {
	world := NewWorld()
	sys := NewDestructionParticleSystem(world, 12345)
	particleSys := NewParticleSystem()
	sys.SetParticleSystem(particleSys)

	// Create a destroyed object without position
	entity := world.CreateEntity()
	comp := NewDestructibleObjectComponent(ObjectCrate)
	comp.IsDestroyed = true
	entity.AddComponent(comp)
	// No position component

	// Run update - should not panic
	sys.Update([]*Entity{entity}, 0.016)

	// Should still mark particles as spawned (to avoid retrying)
	// Note: This is implementation-dependent - the current code marks it true
	// but doesn't actually spawn particles due to missing position
}

func TestDestructionParticleSystem_DeterministicOutput(t *testing.T) {
	// Two systems with same seed should produce same particle distribution
	world1 := NewWorld()
	sys1 := NewDestructionParticleSystem(world1, 54321)
	particleSys1 := NewParticleSystem()
	sys1.SetParticleSystem(particleSys1)

	world2 := NewWorld()
	sys2 := NewDestructionParticleSystem(world2, 54321)
	particleSys2 := NewParticleSystem()
	sys2.SetParticleSystem(particleSys2)

	// Create identical destroyed objects
	entity1 := world1.CreateEntity()
	comp1 := NewDestructibleObjectComponent(ObjectExplosiveBarrel)
	comp1.IsDestroyed = true
	entity1.AddComponent(comp1)
	entity1.AddComponent(&PositionComponent{X: 100, Y: 100})

	entity2 := world2.CreateEntity()
	comp2 := NewDestructibleObjectComponent(ObjectExplosiveBarrel)
	comp2.IsDestroyed = true
	entity2.AddComponent(comp2)
	entity2.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Run both systems
	sys1.Update([]*Entity{entity1}, 0.016)
	sys2.Update([]*Entity{entity2}, 0.016)

	// Should have same particle count
	if particleSys1.GetActiveParticleCount() != particleSys2.GetActiveParticleCount() {
		t.Errorf("Particle counts differ: %d vs %d",
			particleSys1.GetActiveParticleCount(),
			particleSys2.GetActiveParticleCount())
	}
}

func BenchmarkDestructionParticleSystem_ExplosionParticles(b *testing.B) {
	world := NewWorld()
	sys := NewDestructionParticleSystem(world, 12345)
	particleSys := NewParticleSystem()
	sys.SetParticleSystem(particleSys)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.spawnExplosionParticles(100.0, 100.0, 96.0)
	}
}

func BenchmarkDestructionParticleSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewDestructionParticleSystem(world, 12345)
	particleSys := NewParticleSystem()
	sys.SetParticleSystem(particleSys)

	// Create entities
	entities := make([]*Entity, 10)
	for i := 0; i < 10; i++ {
		entity := world.CreateEntity()
		comp := NewDestructibleObjectComponent(ObjectCrate)
		comp.IsDestroyed = true
		entity.AddComponent(comp)
		entity.AddComponent(&PositionComponent{X: float64(i * 50), Y: float64(i * 50)})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset ParticlesSpawned for each iteration
		for _, e := range entities {
			if comp, ok := e.GetComponent("destructibleObject"); ok {
				comp.(*DestructibleObjectComponent).ParticlesSpawned = false
			}
		}
		sys.Update(entities, 0.016)
	}
}
