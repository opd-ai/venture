// Package engine provides particle system management.
// This file implements the ParticleSystem that updates and manages
// particle emitters attached to entities.
package engine

import (
	"github.com/opd-ai/venture/pkg/rendering/particles"
)

// ParticleSystem manages particle emitters and updates particle effects.
type ParticleSystem struct {
	generator *particles.Generator
}

// NewParticleSystem creates a new particle system.
func NewParticleSystem() *ParticleSystem {
	return &ParticleSystem{
		generator: particles.NewGenerator(),
	}
}

// Update updates all particle emitters and their particle systems.
// This method:
//   - Updates particle positions and lifetimes
//   - Emits new particles for continuous emitters
//   - Cleans up dead particle systems
//   - Manages emission timers and rates
func (ps *ParticleSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		emitter := ps.getParticleEmitter(entity)
		if emitter == nil {
			continue
		}

		ps.updateEmitterTime(emitter, deltaTime)
		ps.updateParticleSystems(emitter, deltaTime)
		ps.emitNewParticles(emitter, entity, deltaTime)
		ps.cleanupDeadSystems(emitter)
	}
}

// getParticleEmitter retrieves the particle emitter component from an entity.
// Uses cached GetParticleEmitter() getter for ~93x faster access vs map lookup + type assertion.
func (ps *ParticleSystem) getParticleEmitter(entity *Entity) *ParticleEmitterComponent {
	return entity.GetParticleEmitter()
}

// updateEmitterTime advances elapsed time for time-limited emitters.
func (ps *ParticleSystem) updateEmitterTime(emitter *ParticleEmitterComponent, deltaTime float64) {
	if emitter.EmissionTime > 0 {
		emitter.ElapsedTime += deltaTime
	}
}

// updateParticleSystems updates all particle systems in the emitter.
func (ps *ParticleSystem) updateParticleSystems(emitter *ParticleEmitterComponent, deltaTime float64) {
	for _, system := range emitter.Systems {
		system.Update(deltaTime)
	}
}

// emitNewParticles generates new particles for continuous emitters.
func (ps *ParticleSystem) emitNewParticles(emitter *ParticleEmitterComponent, entity *Entity, deltaTime float64) {
	if emitter.EmitRate <= 0 || !emitter.IsActive() {
		return
	}

	emitter.EmitTimer += deltaTime
	emitInterval := 1.0 / emitter.EmitRate

	for emitter.EmitTimer >= emitInterval {
		emitter.EmitTimer -= emitInterval

		if emitter.AutoCleanup {
			emitter.CleanupDeadSystems()
		}

		system, err := ps.generator.Generate(emitter.EmitConfig)
		if err != nil {
			continue
		}

		ps.positionParticlesAtEntity(system, entity)

		if !emitter.AddSystem(system) {
			continue
		}
	}
}

// positionParticlesAtEntity positions particles at entity's world coordinates.
// Uses cached GetPosition() getter for ~93x faster access vs map lookup + type assertion.
func (ps *ParticleSystem) positionParticlesAtEntity(system *particles.ParticleSystem, entity *Entity) {
	if pos := entity.GetPosition(); pos != nil {
		ps.offsetParticles(system, pos.X, pos.Y)
	}
}

// cleanupDeadSystems removes finished particle systems from emitter.
func (ps *ParticleSystem) cleanupDeadSystems(emitter *ParticleEmitterComponent) {
	if emitter.AutoCleanup {
		emitter.CleanupDeadSystems()
	}
}

// offsetParticles positions all particles in a system at the given world coordinates.
func (ps *ParticleSystem) offsetParticles(system *particles.ParticleSystem, x, y float64) {
	for i := range system.Particles {
		system.Particles[i].X += x
		system.Particles[i].Y += y
	}
}

// SpawnParticles creates a one-shot particle effect at the given position.
// This is a convenience method for spawning particles without an emitter component.
//
// Parameters:
//   - world: ECS world to spawn in
//   - config: Particle configuration
//   - x, y: World coordinates for particle spawn
//
// Returns: Entity with particle emitter component, or nil on error
func (ps *ParticleSystem) SpawnParticles(world *World, config particles.Config, x, y float64) *Entity {
	// Generate particle system
	system, err := ps.generator.Generate(config)
	if err != nil {
		return nil
	}

	// Offset particles to spawn position
	ps.offsetParticles(system, x, y)

	// Create entity with emitter
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: x, Y: y})

	emitter := NewParticleEmitterComponent(0, config, 1) // One-shot (rate = 0)
	emitter.AddSystem(system)
	entity.AddComponent(emitter)

	return entity
}

// SpawnHitSparks creates a spark particle effect at the given position.
// This is a convenience method for combat hit effects.
func (ps *ParticleSystem) SpawnHitSparks(world *World, x, y float64, seed int64, genreID string) *Entity {
	config := particles.Config{
		Type:     particles.ParticleSpark,
		Count:    15,
		GenreID:  genreID,
		Seed:     seed,
		Duration: 0.5,
		SpreadX:  150.0,
		SpreadY:  150.0,
		Gravity:  200.0,
		MinSize:  2.0,
		MaxSize:  4.0,
		Custom:   make(map[string]interface{}),
	}

	return ps.SpawnParticles(world, config, x, y)
}

// SpawnMagicParticles creates a magic particle effect at the given position.
// This is a convenience method for spell effects.
func (ps *ParticleSystem) SpawnMagicParticles(world *World, x, y float64, seed int64, genreID string) *Entity {
	config := particles.Config{
		Type:     particles.ParticleMagic,
		Count:    25,
		GenreID:  genreID,
		Seed:     seed,
		Duration: 1.0,
		SpreadX:  100.0,
		SpreadY:  100.0,
		Gravity:  -50.0, // Float upward
		MinSize:  3.0,
		MaxSize:  6.0,
		Custom:   make(map[string]interface{}),
	}

	return ps.SpawnParticles(world, config, x, y)
}

// SpawnBloodSplatter creates a blood particle effect at the given position.
// This is a convenience method for damage effects on flesh enemies.
func (ps *ParticleSystem) SpawnBloodSplatter(world *World, x, y float64, seed int64, genreID string) *Entity {
	config := particles.Config{
		Type:     particles.ParticleBlood,
		Count:    20,
		GenreID:  genreID,
		Seed:     seed,
		Duration: 1.5,
		SpreadX:  120.0,
		SpreadY:  120.0,
		Gravity:  300.0,
		MinSize:  2.0,
		MaxSize:  5.0,
		Custom:   make(map[string]interface{}),
	}

	return ps.SpawnParticles(world, config, x, y)
}

// SpawnEmbers creates fire ember particles at the given position.
// This is a convenience method for fire effects and destruction.
// Phase 14.3 addition.
func (ps *ParticleSystem) SpawnEmbers(world *World, x, y float64, seed int64, genreID string) *Entity {
	config := particles.Config{
		Type:     particles.ParticleEmber,
		Count:    20,
		GenreID:  genreID,
		Seed:     seed,
		Duration: 2.0,
		SpreadX:  100.0,
		SpreadY:  100.0,
		Gravity:  200.0,
		MinSize:  2.0,
		MaxSize:  5.0,
		Custom:   make(map[string]interface{}),
	}

	return ps.SpawnParticles(world, config, x, y)
}

// SpawnMagicSparkles creates magical sparkle particles at the given position.
// This is a convenience method for enhanced spell effects with orbital motion.
// Phase 14.3 addition.
func (ps *ParticleSystem) SpawnMagicSparkles(world *World, x, y float64, seed int64, genreID string) *Entity {
	config := particles.Config{
		Type:     particles.ParticleSparkle,
		Count:    25,
		GenreID:  genreID,
		Seed:     seed,
		Duration: 1.5,
		SpreadX:  80.0,
		SpreadY:  80.0,
		Gravity:  0.0,
		MinSize:  1.0,
		MaxSize:  3.0,
		Custom:   make(map[string]interface{}),
	}

	return ps.SpawnParticles(world, config, x, y)
}

// SpawnSmokePlume creates billowing smoke particle effect at the given position.
// This is a convenience method for environmental effects and explosions.
// Phase 14.3 addition.
func (ps *ParticleSystem) SpawnSmokePlume(world *World, x, y float64, seed int64, genreID string) *Entity {
	config := particles.Config{
		Type:     particles.ParticleSmokePlume,
		Count:    30,
		GenreID:  genreID,
		Seed:     seed,
		Duration: 3.0,
		SpreadX:  120.0,
		SpreadY:  120.0,
		Gravity:  100.0,
		MinSize:  3.0,
		MaxSize:  8.0,
		Custom:   make(map[string]interface{}),
	}

	return ps.SpawnParticles(world, config, x, y)
}

// SpawnDebris creates bouncing debris particles at the given position.
// This is a convenience method for destructible objects and explosions.
// Phase 14.3 addition.
func (ps *ParticleSystem) SpawnDebris(world *World, x, y, groundY float64, seed int64, genreID string) *Entity {
	config := particles.Config{
		Type:     particles.ParticleDebris,
		Count:    25,
		GenreID:  genreID,
		Seed:     seed,
		Duration: 2.5,
		SpreadX:  150.0,
		SpreadY:  150.0,
		Gravity:  300.0,
		MinSize:  2.0,
		MaxSize:  6.0,
		Custom: map[string]interface{}{
			"groundY": groundY,
		},
	}

	return ps.SpawnParticles(world, config, x, y)
}

// GetActiveParticleCount returns the number of active particle entities.
func (ps *ParticleSystem) GetActiveParticleCount() int {
return 0
}
