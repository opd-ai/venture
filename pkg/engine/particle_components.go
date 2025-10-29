// Package engine provides particle system components for the ECS.
// This file implements components for managing particle effects on entities.
package engine

import (
	"github.com/opd-ai/venture/pkg/rendering/particles"
)

// ParticleEmitterComponent allows entities to emit particle effects.
// This component stores active particle systems and their emission parameters.
type ParticleEmitterComponent struct {
	// Active particle systems attached to this entity
	Systems []*particles.ParticleSystem

	// Emission configuration (for continuous emitters)
	EmitRate     float64 // Particles per second (0 = one-shot)
	EmitTimer    float64 // Time until next emission
	EmitConfig   particles.Config
	MaxSystems   int     // Maximum number of concurrent systems
	AutoCleanup  bool    // Automatically remove dead systems
	EmissionTime float64 // Total time to emit (0 = infinite)
	ElapsedTime  float64 // Time since emitter started
}

// Type returns the component type identifier.
func (p *ParticleEmitterComponent) Type() string {
	return "particle_emitter"
}

// NewParticleEmitterComponent creates a new particle emitter component.
// Parameters:
//   - emitRate: Particles per second (0 for one-shot effects)
//   - config: Configuration for particle generation
//   - maxSystems: Maximum concurrent particle systems (default 10)
//
// Returns: Initialized ParticleEmitterComponent
func NewParticleEmitterComponent(emitRate float64, config particles.Config, maxSystems int) *ParticleEmitterComponent {
	if maxSystems <= 0 {
		maxSystems = 10
	}

	return &ParticleEmitterComponent{
		Systems:      make([]*particles.ParticleSystem, 0, maxSystems),
		EmitRate:     emitRate,
		EmitTimer:    0,
		EmitConfig:   config,
		MaxSystems:   maxSystems,
		AutoCleanup:  true,
		EmissionTime: 0, // Infinite by default
		ElapsedTime:  0,
	}
}

// AddSystem adds a particle system to this emitter.
// Returns: true if added, false if at capacity
func (p *ParticleEmitterComponent) AddSystem(system *particles.ParticleSystem) bool {
	if len(p.Systems) >= p.MaxSystems {
		// Try to clean up dead systems first
		if p.AutoCleanup {
			p.CleanupDeadSystems()
		}

		// Still at capacity?
		if len(p.Systems) >= p.MaxSystems {
			return false
		}
	}

	p.Systems = append(p.Systems, system)
	return true
}

// CleanupDeadSystems removes particle systems with no alive particles.
// Dead systems are returned to the pool to reduce GC pressure.
func (p *ParticleEmitterComponent) CleanupDeadSystems() {
	alive := make([]*particles.ParticleSystem, 0, len(p.Systems))
	for _, system := range p.Systems {
		if system.IsAlive() {
			alive = append(alive, system)
		} else {
			// Return dead system to pool for reuse
			particles.ReleaseParticleSystem(system)
		}
	}
	p.Systems = alive
}

// IsActive returns true if emitter should continue emitting.
func (p *ParticleEmitterComponent) IsActive() bool {
	// Infinite emission
	if p.EmissionTime <= 0 {
		return p.EmitRate > 0
	}

	// Time-limited emission
	return p.ElapsedTime < p.EmissionTime
}

// HasActiveSystems returns true if any particle systems are alive.
func (p *ParticleEmitterComponent) HasActiveSystems() bool {
	for _, system := range p.Systems {
		if system.IsAlive() {
			return true
		}
	}
	return false
}
