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
		comp, ok := entity.GetComponent("particle_emitter")
		if !ok {
			continue
		}

		emitter := comp.(*ParticleEmitterComponent)

		// Update elapsed time for time-limited emitters
		if emitter.EmissionTime > 0 {
			emitter.ElapsedTime += deltaTime
		}

		// Update all particle systems
		for _, system := range emitter.Systems {
			system.Update(deltaTime)
		}

		// Emit new particles for continuous emitters
		if emitter.EmitRate > 0 && emitter.IsActive() {
			emitter.EmitTimer += deltaTime

			// Time to emit?
			emitInterval := 1.0 / emitter.EmitRate
			for emitter.EmitTimer >= emitInterval {
				emitter.EmitTimer -= emitInterval

				// GAP-001 FIX: Cleanup dead systems BEFORE attempting to add new ones
				// This ensures capacity is available for continuous emission
				if emitter.AutoCleanup {
					emitter.CleanupDeadSystems()
				}

				// Generate new particle system
				system, err := ps.generator.Generate(emitter.EmitConfig)
				if err != nil {
					// Failed to generate - skip this emission
					continue
				}

				// Position particles at entity's position
				if posComp, ok := entity.GetComponent("position"); ok {
					pos := posComp.(*PositionComponent)
					ps.offsetParticles(system, pos.X, pos.Y)
				}

				// Add to emitter (with capacity check)
				if !emitter.AddSystem(system) {
					// Still at capacity after cleanup - this indicates a problem
					// Log in verbose mode but continue (fail gracefully)
					continue
				}
			}
		}

		// GAP-001 FIX: Also cleanup at end of Update() for one-shot emitters
		// Continuous emitters cleanup before emission, but one-shot emitters
		// need cleanup here to remove dead systems
		if emitter.AutoCleanup {
			emitter.CleanupDeadSystems()
		}
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
