package engine

import (
	"testing"
)

func TestNewCompanionLevelUpParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionLevelUpParticleSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewCompanionLevelUpParticleSystem returned nil")
	}
	if sys.world != world {
		t.Error("world reference not set")
	}
	if sys.seed != 12345 {
		t.Errorf("seed = %d, want 12345", sys.seed)
	}
	if sys.rng == nil {
		t.Error("rng not initialized")
	}
	if sys.baseParticleCount != 15 {
		t.Errorf("baseParticleCount = %d, want 15", sys.baseParticleCount)
	}
}

func TestCompanionLevelUpParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionLevelUpParticleSystem(world, 12345)
	particleSys := NewParticleSystem()

	sys.SetParticleSystem(particleSys)

	if sys.particleSystem != particleSys {
		t.Error("particle system not set correctly")
	}
}

func TestCompanionLevelUpParticleSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionLevelUpParticleSystem(world, 12345)

	sys.SetGenre("fantasy")

	if sys.genreID != "fantasy" {
		t.Errorf("genreID = %q, want %q", sys.genreID, "fantasy")
	}
}

func TestCompanionLevelUpParticleSystem_Update(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionLevelUpParticleSystem(world, 12345)

	// Update should be no-op (callback-driven)
	sys.Update([]*Entity{}, 0.016)
	// No panic = success
}

func TestCompanionLevelUpParticleSystem_OnCompanionLevelUp_NilInputs(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionLevelUpParticleSystem(world, 12345)

	// Should not panic with nil inputs
	sys.OnCompanionLevelUp(nil, 1)

	// Should not panic with no particle system
	entity := world.CreateEntity()
	sys.OnCompanionLevelUp(entity, 1)
}

func TestCompanionLevelUpParticleSystem_OnCompanionLevelUp_NotCompanion(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionLevelUpParticleSystem(world, 12345)
	sys.SetParticleSystem(NewParticleSystem())

	// Entity without companion component should be ignored
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	sys.OnCompanionLevelUp(entity, 5)
	// Should not spawn particles (no companion component)
}

func TestCompanionLevelUpParticleSystem_OnCompanionLevelUp_NoPosition(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionLevelUpParticleSystem(world, 12345)
	sys.SetParticleSystem(NewParticleSystem())

	// Companion without position should be ignored
	entity := world.CreateEntity()
	entity.AddComponent(&CompanionComponent{OwnerID: 1, Level: 1})

	sys.OnCompanionLevelUp(entity, 2)
	// Should not panic
}

func TestCompanionLevelUpParticleSystem_OnCompanionLevelUp_FullSetup(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionLevelUpParticleSystem(world, 12345)
	particleSys := NewParticleSystem()
	sys.SetParticleSystem(particleSys)
	sys.SetGenre("fantasy")

	// Create a proper companion entity
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 200})
	entity.AddComponent(&CompanionComponent{OwnerID: 1, Level: 1})

	// Should spawn particles without error
	sys.OnCompanionLevelUp(entity, 2)
	// No panic = success
}

func TestCompanionLevelUpParticleSystem_MilestoneScaling(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionLevelUpParticleSystem(world, 12345)
	particleSys := NewParticleSystem()
	sys.SetParticleSystem(particleSys)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 200})
	entity.AddComponent(&CompanionComponent{OwnerID: 1, Level: 4})

	// Level 5 should trigger milestone scaling
	sys.OnCompanionLevelUp(entity, 5)

	// Level 10 should also trigger
	sys.OnCompanionLevelUp(entity, 10)
	// No panic = success
}

func TestCompanionProgressionSystem_AddLevelUpCallback(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionProgressionSystem(world)

	callback := func(_ *Entity, _ int) {
		// callback registered
	}

	sys.AddLevelUpCallback(callback)

	if len(sys.levelUpCallbacks) != 1 {
		t.Errorf("callback count = %d, want 1", len(sys.levelUpCallbacks))
	}

	// Nil callback should not be added
	sys.AddLevelUpCallback(nil)
	if len(sys.levelUpCallbacks) != 1 {
		t.Errorf("nil callback was added, count = %d", len(sys.levelUpCallbacks))
	}
}

func TestCompanionProgressionSystem_CallbackTriggered(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionProgressionSystem(world)

	var callbackEntity *Entity
	callbackLevel := 0
	callback := func(entity *Entity, level int) {
		callbackEntity = entity
		callbackLevel = level
	}
	sys.AddLevelUpCallback(callback)

	// Create companion with enough XP to level up
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:    1,
		Level:      1,
		Experience: 150, // More than needed for level 2 (100 * 1.5^0 = 100)
	})
	companion.AddComponent(&CompanionStatsComponent{
		HP:      100,
		MaxHP:   100,
		Attack:  10,
		Defense: 5,
		Speed:   50,
	})

	sys.Update(0.016)

	if callbackEntity != companion {
		t.Error("callback not called with correct entity")
	}
	if callbackLevel != 2 {
		t.Errorf("callbackLevel = %d, want 2", callbackLevel)
	}
}

func TestCompanionProgressionSystem_NoCallbackWhenNoLevelUp(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionProgressionSystem(world)

	called := false
	callback := func(entity *Entity, level int) {
		called = true
	}
	sys.AddLevelUpCallback(callback)

	// Create companion without enough XP
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:    1,
		Level:      1,
		Experience: 50, // Not enough for level 2
	})
	companion.AddComponent(&CompanionStatsComponent{
		HP:      100,
		MaxHP:   100,
		Attack:  10,
		Defense: 5,
		Speed:   50,
	})

	sys.Update(0.016)

	if called {
		t.Error("callback should not be called when companion doesn't level up")
	}
}

func TestCompanionLevelUpParticleSystem_Integration(t *testing.T) {
	// Full integration test
	world := NewWorld()
	progressionSys := NewCompanionProgressionSystem(world)
	particleSys := NewParticleSystem()
	levelUpParticleSys := NewCompanionLevelUpParticleSystem(world, 12345)
	levelUpParticleSys.SetParticleSystem(particleSys)
	levelUpParticleSys.SetGenre("fantasy")

	// Register callback
	progressionSys.AddLevelUpCallback(levelUpParticleSys.OnCompanionLevelUp)

	// Create owner
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 0, Y: 0})

	// Create companion
	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 100, Y: 100})
	companion.AddComponent(&CompanionComponent{
		OwnerID:    owner.ID,
		Level:      1,
		Experience: 150,
	})
	companion.AddComponent(&CompanionStatsComponent{
		HP:      100,
		MaxHP:   100,
		Attack:  10,
		Defense: 5,
		Speed:   50,
	})

	// Process update - should trigger level up and particle spawn
	progressionSys.Update(0.016)

	// Verify companion leveled up
	compRaw, _ := companion.GetComponent("companion")
	comp := compRaw.(*CompanionComponent)
	if comp.Level != 2 {
		t.Errorf("companion level = %d, want 2", comp.Level)
	}
}

func BenchmarkCompanionLevelUpParticleSystem_OnLevelUp(b *testing.B) {
	world := NewWorld()
	sys := NewCompanionLevelUpParticleSystem(world, 12345)
	particleSys := NewParticleSystem()
	sys.SetParticleSystem(particleSys)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 200})
	entity.AddComponent(&CompanionComponent{OwnerID: 1, Level: 1})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.OnCompanionLevelUp(entity, i%50+1)
	}
}
