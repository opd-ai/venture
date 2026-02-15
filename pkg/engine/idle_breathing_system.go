// Package engine provides the EntityIdleBreathingSystem for subtle idle animation.
// When entities are stationary, their sprites oscillate vertically in a slow
// sinusoidal pattern simulating breathing. Genre presets control style: fantasy
// uses gentle breathing, horror uses labored irregular breaths, sci-fi a smooth
// mechanical hum, cyberpunk a glitchy micro-jitter, and post-apocalyptic a
// heavy tired rhythm.
package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// IdleBreathingComponent stores per-entity idle breathing state.
// Pure data — all logic lives in EntityIdleBreathingSystem.
type IdleBreathingComponent struct {
	// Phase tracks the current position in the breathing cycle (radians).
	Phase float64
	// OffsetY is the vertical pixel offset applied during rendering.
	OffsetY float64
	// Active indicates the entity is currently breathing (idle).
	Active bool
	// IdleTime tracks how long the entity has been stationary (seconds).
	IdleTime float64
}

// Type returns the component type identifier.
func (c *IdleBreathingComponent) Type() string { return "idle_breathing" }

// idleBreathingPreset holds genre-specific breathing configuration.
type idleBreathingPreset struct {
	Amplitude float64 // Max vertical displacement in pixels (0.3–1.0)
	Frequency float64 // Breathing speed (cycles per second)
	RampTime  float64 // Seconds of idle before full breathing activates
	Jitter    float64 // Random phase variation per cycle (0=smooth, 1=erratic)
}

// EntityIdleBreathingSystem applies a subtle sinusoidal vertical offset to
// stationary entities, simulating idle breathing. It reads VelocityComponent
// to detect idle state and writes OffsetY into IdleBreathingComponent for
// the render system to consume.
type EntityIdleBreathingSystem struct {
	world   *World
	seed    int64
	rng     *rand.Rand
	logger  *logrus.Entry
	genreID string
	preset  idleBreathingPreset

	// Squared speed threshold below which entity counts as idle.
	idleSpeedSq float64
}

// NewEntityIdleBreathingSystem creates a new idle breathing system.
func NewEntityIdleBreathingSystem(world *World, seed int64) *EntityIdleBreathingSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "idle_breathing",
	})

	sys := &EntityIdleBreathingSystem{
		world:       world,
		seed:        seed,
		rng:         rand.New(rand.NewSource(seed)),
		logger:      logger,
		genreID:     "fantasy",
		idleSpeedSq: 25.0, // 5 px/s threshold
	}
	sys.preset = sys.getGenrePreset("fantasy")
	logger.Debug("idle breathing system created")
	return sys
}

// SetGenre configures genre-specific breathing parameters.
func (s *EntityIdleBreathingSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getGenrePreset(genreID)
	s.logger.WithField("genre", genreID).Debug("genre preset applied")
}

// getGenrePreset returns breathing configuration for the given genre.
func (s *EntityIdleBreathingSystem) getGenrePreset(genreID string) idleBreathingPreset {
	switch genreID {
	case "horror":
		return idleBreathingPreset{Amplitude: 1.0, Frequency: 0.5, RampTime: 0.3, Jitter: 0.3}
	case "scifi":
		return idleBreathingPreset{Amplitude: 0.3, Frequency: 1.5, RampTime: 0.5, Jitter: 0.0}
	case "cyberpunk":
		return idleBreathingPreset{Amplitude: 0.4, Frequency: 2.0, RampTime: 0.4, Jitter: 0.15}
	case "postapoc":
		return idleBreathingPreset{Amplitude: 0.8, Frequency: 0.6, RampTime: 0.3, Jitter: 0.1}
	default: // fantasy
		return idleBreathingPreset{Amplitude: 0.5, Frequency: 0.8, RampTime: 0.5, Jitter: 0.0}
	}
}

// Update processes all entities, advancing breathing phase for idle entities
// and decaying the offset when they start moving.
func (s *EntityIdleBreathingSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		vel := entity.GetVelocity()
		if vel == nil {
			continue
		}

		breath := s.getOrCreateBreathing(entity)
		speedSq := vel.VX*vel.VX + vel.VY*vel.VY

		if speedSq <= s.idleSpeedSq {
			// Entity is idle — accumulate idle time and breathe
			breath.IdleTime += deltaTime
			breath.Active = true

			// Ramp factor: 0→1 over RampTime seconds for smooth onset
			ramp := 1.0
			if s.preset.RampTime > 0 {
				ramp = math.Min(breath.IdleTime/s.preset.RampTime, 1.0)
			}

			// Advance breathing phase
			breath.Phase += deltaTime * s.preset.Frequency * 2.0 * math.Pi
			if breath.Phase > 2.0*math.Pi {
				breath.Phase -= 2.0 * math.Pi
			}

			// Apply jitter for horror/cyberpunk irregular breathing
			jitterOffset := 0.0
			if s.preset.Jitter > 0 {
				jitterOffset = s.preset.Jitter * s.rng.Float64() * s.preset.Amplitude
			}

			breath.OffsetY = ramp * (s.preset.Amplitude*math.Sin(breath.Phase) + jitterOffset)
		} else {
			// Entity is moving — decay breathing offset and reset idle time
			breath.IdleTime = 0
			breath.OffsetY *= (1.0 - math.Min(deltaTime*8.0, 1.0))
			if math.Abs(breath.OffsetY) < 0.02 {
				breath.OffsetY = 0
				breath.Active = false
			}
		}
	}
}

// getOrCreateBreathing retrieves or attaches an IdleBreathingComponent.
func (s *EntityIdleBreathingSystem) getOrCreateBreathing(entity *Entity) *IdleBreathingComponent {
	if comp, ok := entity.GetComponent("idle_breathing"); ok {
		if breath, ok := comp.(*IdleBreathingComponent); ok {
			return breath
		}
	}
	breath := &IdleBreathingComponent{}
	entity.AddComponent(breath)
	return breath
}
