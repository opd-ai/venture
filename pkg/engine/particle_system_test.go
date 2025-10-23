//go:build test
// +build test

// Package engine provides tests for particle system integration.
package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

// TestNewParticleSystem verifies particle system creation.
func TestNewParticleSystem(t *testing.T) {
	ps := NewParticleSystem()

	if ps == nil {
		t.Fatal("NewParticleSystem returned nil")
	}

	if ps.generator == nil {
		t.Error("ParticleSystem missing particle generator")
	}
}

// TestParticleSystem_Update_ContinuousEmitter tests continuous particle emission.
func TestParticleSystem_Update_ContinuousEmitter(t *testing.T) {
	world := NewWorld()
	ps := NewParticleSystem()

	// Create entity with continuous emitter
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	config := particles.Config{
		Type:     particles.ParticleSpark,
		Count:    10,
		GenreID:  "fantasy",
		Seed:     12345,
		Duration: 1.0,
		SpreadX:  50.0,
		SpreadY:  50.0,
		Gravity:  0.0,
		MinSize:  2.0,
		MaxSize:  4.0,
		Custom:   make(map[string]interface{}),
	}

	emitter := NewParticleEmitterComponent(5.0, config, 5) // 5 particles/second
	entity.AddComponent(emitter)

	// Update for 0.5 seconds (should emit 2-3 times at 5Hz)
	for i := 0; i < 5; i++ {
		ps.Update(world.GetEntities(), 0.1)
	}

	// Verify particle systems were emitted
	if len(emitter.Systems) == 0 {
		t.Error("No particle systems emitted by continuous emitter")
	}

	// Verify particles exist
	hasParticles := false
	for _, system := range emitter.Systems {
		if len(system.Particles) > 0 {
			hasParticles = true
			break
		}
	}

	if !hasParticles {
		t.Error("Emitted systems have no particles")
	}
}

// TestParticleSystem_Update_OneShotEmitter tests one-shot particle spawning.
func TestParticleSystem_Update_OneShotEmitter(t *testing.T) {
	world := NewWorld()
	ps := NewParticleSystem()

	// Spawn one-shot particles
	config := particles.Config{
		Type:     particles.ParticleSpark,
		Count:    15,
		GenreID:  "fantasy",
		Seed:     54321,
		Duration: 0.5,
		SpreadX:  100.0,
		SpreadY:  100.0,
		Gravity:  200.0,
		MinSize:  2.0,
		MaxSize:  4.0,
		Custom:   make(map[string]interface{}),
	}

	entity := ps.SpawnParticles(world, config, 200, 200)

	if entity == nil {
		t.Fatal("SpawnParticles returned nil entity")
	}

	// Verify entity has emitter component
	comp, ok := entity.GetComponent("particle_emitter")
	if !ok {
		t.Fatal("Spawned entity missing particle_emitter component")
	}

	emitter := comp.(*ParticleEmitterComponent)

	// Verify system was added
	if len(emitter.Systems) != 1 {
		t.Errorf("Expected 1 particle system, got %d", len(emitter.Systems))
	}

	// Verify particle count
	if len(emitter.Systems[0].Particles) != 15 {
		t.Errorf("Expected 15 particles, got %d", len(emitter.Systems[0].Particles))
	}

	// Verify particles are positioned at spawn location
	for _, particle := range emitter.Systems[0].Particles {
		// Particles should be near spawn point (within spread)
		if particle.X < 100 || particle.X > 300 {
			t.Errorf("Particle X position out of range: %f", particle.X)
		}
		if particle.Y < 100 || particle.Y > 300 {
			t.Errorf("Particle Y position out of range: %f", particle.Y)
		}
	}
}

// TestParticleSystem_Update_TimeLimitedEmitter tests time-limited emission.
func TestParticleSystem_Update_TimeLimitedEmitter(t *testing.T) {
	world := NewWorld()
	ps := NewParticleSystem()

	// Create entity with time-limited emitter (0.5 seconds total)
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	config := particles.Config{
		Type:     particles.ParticleSpark,
		Count:    10,
		GenreID:  "fantasy",
		Seed:     99999,
		Duration: 1.0,
		SpreadX:  50.0,
		SpreadY:  50.0,
		Gravity:  0.0,
		MinSize:  2.0,
		MaxSize:  4.0,
		Custom:   make(map[string]interface{}),
	}

	emitter := NewParticleEmitterComponent(10.0, config, 10) // 10 particles/second
	emitter.EmissionTime = 0.5                               // Only emit for 0.5 seconds
	entity.AddComponent(emitter)

	// Update for 0.3 seconds (within emission time)
	for i := 0; i < 3; i++ {
		ps.Update(world.GetEntities(), 0.1)
	}

	systemsBeforeTimeout := len(emitter.Systems)

	// Update for another 0.5 seconds (past emission time)
	for i := 0; i < 5; i++ {
		ps.Update(world.GetEntities(), 0.1)
	}

	systemsAfterTimeout := len(emitter.Systems)

	// After timeout, no new systems should be added (only existing ones updated)
	// Note: Systems may be cleaned up if they die, so count may decrease but shouldn't increase
	if systemsAfterTimeout > systemsBeforeTimeout {
		t.Errorf("Emitter continued emitting after timeout: before=%d, after=%d",
			systemsBeforeTimeout, systemsAfterTimeout)
	}
}

// TestParticleSystem_Update_ParticleLifetime tests particle aging and death.
func TestParticleSystem_Update_ParticleLifetime(t *testing.T) {
	world := NewWorld()
	ps := NewParticleSystem()

	// Spawn particles with very short lifetime
	config := particles.Config{
		Type:     particles.ParticleSpark,
		Count:    10,
		GenreID:  "fantasy",
		Seed:     11111,
		Duration: 0.2, // 0.2 second lifetime
		SpreadX:  10.0,
		SpreadY:  10.0,
		Gravity:  0.0,
		MinSize:  2.0,
		MaxSize:  4.0,
		Custom:   make(map[string]interface{}),
	}

	entity := ps.SpawnParticles(world, config, 100, 100)
	comp, ok := entity.GetComponent("particle_emitter")
	if !ok {
		t.Fatal("Entity missing particle_emitter component")
	}
	emitter := comp.(*ParticleEmitterComponent)
	emitter.AutoCleanup = true

	// Verify initial particle count
	initialAlive := 0
	for _, system := range emitter.Systems {
		initialAlive += len(system.GetAliveParticles())
	}

	if initialAlive != 10 {
		t.Errorf("Expected 10 alive particles initially, got %d", initialAlive)
	}

	// Update for 0.3 seconds (longer than particle lifetime)
	for i := 0; i < 6; i++ {
		ps.Update(world.GetEntities(), 0.05)
	}

	// All particles should be dead
	finalAlive := 0
	for _, system := range emitter.Systems {
		finalAlive += len(system.GetAliveParticles())
	}

	if finalAlive > 0 {
		t.Errorf("Expected 0 alive particles after lifetime, got %d", finalAlive)
	}

	// Systems should be cleaned up (auto-cleanup enabled)
	if len(emitter.Systems) > 0 {
		t.Errorf("Expected emitter to cleanup dead systems, still has %d", len(emitter.Systems))
	}
}

// TestParticleSystem_SpawnHitSparks tests combat hit effect convenience method.
func TestParticleSystem_SpawnHitSparks(t *testing.T) {
	world := NewWorld()
	ps := NewParticleSystem()

	entity := ps.SpawnHitSparks(world, 150, 200, 54321, "fantasy")

	if entity == nil {
		t.Fatal("SpawnHitSparks returned nil")
	}

	comp, ok := entity.GetComponent("particle_emitter")
	if !ok {
		t.Fatal("Entity missing particle_emitter component")
	}
	emitter := comp.(*ParticleEmitterComponent)

	// Verify configuration
	if len(emitter.Systems) != 1 {
		t.Errorf("Expected 1 system, got %d", len(emitter.Systems))
	}

	system := emitter.Systems[0]

	// Verify particle type
	if system.Type != particles.ParticleSpark {
		t.Errorf("Expected ParticleSpark type, got %v", system.Type)
	}

	// Verify particle count (15 in hit sparks config)
	if len(system.Particles) != 15 {
		t.Errorf("Expected 15 particles, got %d", len(system.Particles))
	}
}

// TestParticleSystem_SpawnMagicParticles tests magic spell effect convenience method.
func TestParticleSystem_SpawnMagicParticles(t *testing.T) {
	world := NewWorld()
	ps := NewParticleSystem()

	entity := ps.SpawnMagicParticles(world, 250, 300, 99999, "scifi")

	if entity == nil {
		t.Fatal("SpawnMagicParticles returned nil")
	}

	comp, ok := entity.GetComponent("particle_emitter")
	if !ok {
		t.Fatal("Entity missing particle_emitter component")
	}
	emitter := comp.(*ParticleEmitterComponent)

	// Verify configuration
	if len(emitter.Systems) != 1 {
		t.Errorf("Expected 1 system, got %d", len(emitter.Systems))
	}

	system := emitter.Systems[0]

	// Verify particle type
	if system.Type != particles.ParticleMagic {
		t.Errorf("Expected ParticleMagic type, got %v", system.Type)
	}

	// Verify particle count (25 in magic config)
	if len(system.Particles) != 25 {
		t.Errorf("Expected 25 particles, got %d", len(system.Particles))
	}

	// Verify upward gravity (magic floats)
	if system.Config.Gravity >= 0 {
		t.Errorf("Expected negative gravity for floating magic, got %f", system.Config.Gravity)
	}
}

// TestParticleSystem_SpawnBloodSplatter tests blood effect convenience method.
func TestParticleSystem_SpawnBloodSplatter(t *testing.T) {
	world := NewWorld()
	ps := NewParticleSystem()

	entity := ps.SpawnBloodSplatter(world, 175, 225, 44444, "horror")

	if entity == nil {
		t.Fatal("SpawnBloodSplatter returned nil")
	}

	comp, ok := entity.GetComponent("particle_emitter")
	if !ok {
		t.Fatal("Entity missing particle_emitter component")
	}
	emitter := comp.(*ParticleEmitterComponent)

	// Verify configuration
	if len(emitter.Systems) != 1 {
		t.Errorf("Expected 1 system, got %d", len(emitter.Systems))
	}

	system := emitter.Systems[0]

	// Verify particle type
	if system.Type != particles.ParticleBlood {
		t.Errorf("Expected ParticleBlood type, got %v", system.Type)
	}

	// Verify particle count (20 in blood config)
	if len(system.Particles) != 20 {
		t.Errorf("Expected 20 particles, got %d", len(system.Particles))
	}

	// Verify downward gravity (blood falls)
	if system.Config.Gravity <= 0 {
		t.Errorf("Expected positive gravity for falling blood, got %f", system.Config.Gravity)
	}
}

// TestParticleEmitterComponent_AddSystem tests system capacity management.
func TestParticleEmitterComponent_AddSystem(t *testing.T) {
	config := particles.DefaultConfig()
	emitter := NewParticleEmitterComponent(0, config, 3) // Max 3 systems

	// Create mock particle systems
	system1 := &particles.ParticleSystem{Particles: make([]particles.Particle, 10)}
	system2 := &particles.ParticleSystem{Particles: make([]particles.Particle, 10)}
	system3 := &particles.ParticleSystem{Particles: make([]particles.Particle, 10)}
	system4 := &particles.ParticleSystem{Particles: make([]particles.Particle, 10)}

	// Add up to capacity
	if !emitter.AddSystem(system1) {
		t.Error("Failed to add system 1")
	}
	if !emitter.AddSystem(system2) {
		t.Error("Failed to add system 2")
	}
	if !emitter.AddSystem(system3) {
		t.Error("Failed to add system 3")
	}

	// Should be at capacity
	if len(emitter.Systems) != 3 {
		t.Errorf("Expected 3 systems, got %d", len(emitter.Systems))
	}

	// Adding beyond capacity should fail
	if emitter.AddSystem(system4) {
		t.Error("Should not add system beyond capacity")
	}
}

// TestParticleEmitterComponent_CleanupDeadSystems tests memory management.
func TestParticleEmitterComponent_CleanupDeadSystems(t *testing.T) {
	config := particles.DefaultConfig()
	emitter := NewParticleEmitterComponent(0, config, 10)

	// Add systems with dead particles (Life <= 0)
	deadSystem1 := &particles.ParticleSystem{
		Particles: []particles.Particle{
			{Life: 0.0},  // Dead
			{Life: -0.1}, // Dead
		},
	}

	deadSystem2 := &particles.ParticleSystem{
		Particles: []particles.Particle{
			{Life: 0.0}, // Dead
		},
	}

	aliveSystem := &particles.ParticleSystem{
		Particles: []particles.Particle{
			{Life: 0.5}, // Alive
			{Life: 1.0}, // Alive
		},
	}

	emitter.AddSystem(deadSystem1)
	emitter.AddSystem(aliveSystem)
	emitter.AddSystem(deadSystem2)

	// Cleanup
	emitter.CleanupDeadSystems()

	// Only alive system should remain
	if len(emitter.Systems) != 1 {
		t.Errorf("Expected 1 system after cleanup, got %d", len(emitter.Systems))
	}

	if emitter.Systems[0] != aliveSystem {
		t.Error("Cleanup removed alive system")
	}
}

// BenchmarkParticleSystem_Update measures particle update performance.
func BenchmarkParticleSystem_Update(b *testing.B) {
	world := NewWorld()
	ps := NewParticleSystem()

	// Create 10 entities with 20 particles each
	for i := 0; i < 10; i++ {
		config := particles.Config{
			Type:     particles.ParticleSpark,
			Count:    20,
			GenreID:  "fantasy",
			Seed:     int64(i * 1000),
			Duration: 2.0,
			SpreadX:  100.0,
			SpreadY:  100.0,
			Gravity:  100.0,
			MinSize:  2.0,
			MaxSize:  4.0,
			Custom:   make(map[string]interface{}),
		}

		ps.SpawnParticles(world, config, float64(i*50), float64(i*50))
	}

	entities := world.GetEntities()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ps.Update(entities, 0.016) // ~60 FPS delta
	}
}
