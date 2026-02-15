// Package engine provides the MovementBobSystem for walk-cycle sprite bobbing.
// When entities move, their sprites oscillate vertically in a sinusoidal pattern,
// giving visual weight and rhythm to movement. Genre presets control amplitude
// and frequency: fantasy uses a gentle bob, horror a lurching stagger, sci-fi
// a smooth low-amplitude glide, and cyberpunk a sharp rhythmic bounce.
package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// MovementBobComponent stores per-entity bob state and render offset.
// Pure data — all logic lives in MovementBobSystem.
type MovementBobComponent struct {
	// Phase tracks the current position in the sinusoidal cycle (radians).
	Phase float64
	// OffsetY is the vertical pixel offset applied during rendering.
	OffsetY float64
	// Active indicates the entity is currently bobbing.
	Active bool
}

// Type returns the component type identifier.
func (m *MovementBobComponent) Type() string { return "movement_bob" }

// movementBobPreset holds genre-specific bob configuration.
type movementBobPreset struct {
	Amplitude float64 // Max vertical displacement in pixels (1–3)
	Frequency float64 // Oscillation speed multiplier
	Damping   float64 // How quickly bob decays when stopping (0–1, higher = faster)
}

// MovementBobSystem applies sinusoidal vertical offset to moving entities,
// creating a walk-cycle bobbing effect. It reads VelocityComponent to determine
// movement speed and writes OffsetY into MovementBobComponent for the render
// system to consume.
type MovementBobSystem struct {
	world   *World
	seed    int64
	rng     *rand.Rand
	logger  *logrus.Entry
	genreID string
	preset  movementBobPreset

	// Squared speed threshold below which bob dampens toward zero.
	minSpeedSq float64
}

// NewMovementBobSystem creates a movement bob system with default fantasy preset.
func NewMovementBobSystem(world *World, seed int64) *MovementBobSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "movement_bob",
	})

	sys := &MovementBobSystem{
		world:      world,
		seed:       seed,
		rng:        rand.New(rand.NewSource(seed)),
		logger:     logger,
		genreID:    "fantasy",
		minSpeedSq: 100.0, // 10 px/s threshold
	}
	sys.preset = sys.getGenrePreset("fantasy")
	logger.Debug("movement bob system created")
	return sys
}

// SetGenre configures genre-specific bob parameters.
func (s *MovementBobSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getGenrePreset(genreID)
	s.logger.WithField("genre", genreID).Debug("genre preset applied")
}

// getGenrePreset returns bob configuration for the given genre.
func (s *MovementBobSystem) getGenrePreset(genreID string) movementBobPreset {
	switch genreID {
	case "horror":
		return movementBobPreset{Amplitude: 2.5, Frequency: 0.7, Damping: 0.6}
	case "scifi":
		return movementBobPreset{Amplitude: 0.8, Frequency: 1.4, Damping: 0.9}
	case "cyberpunk":
		return movementBobPreset{Amplitude: 1.8, Frequency: 1.6, Damping: 0.85}
	case "postapoc":
		return movementBobPreset{Amplitude: 2.2, Frequency: 0.9, Damping: 0.7}
	default: // fantasy
		return movementBobPreset{Amplitude: 1.5, Frequency: 1.2, Damping: 0.8}
	}
}

// Update processes all entities, advancing bob phase for moving entities and
// decaying offset for stationary ones.
func (s *MovementBobSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		vel := entity.GetVelocity()
		if vel == nil {
			continue
		}

		bob := s.getOrCreateBob(entity)
		speedSq := vel.VX*vel.VX + vel.VY*vel.VY

		if speedSq > s.minSpeedSq {
			// Scale frequency with speed (capped to avoid jitter)
			speedFactor := math.Min(math.Sqrt(speedSq)/200.0, 1.5)
			bob.Phase += deltaTime * s.preset.Frequency * speedFactor * 2.0 * math.Pi
			// Keep phase in [0, 2π) to avoid float drift
			if bob.Phase > 2.0*math.Pi {
				bob.Phase -= 2.0 * math.Pi
			}
			bob.OffsetY = s.preset.Amplitude * math.Sin(bob.Phase)
			bob.Active = true
		} else {
			// Dampen toward zero when stationary
			bob.OffsetY *= (1.0 - s.preset.Damping*deltaTime*10.0)
			if math.Abs(bob.OffsetY) < 0.05 {
				bob.OffsetY = 0
				bob.Active = false
			}
		}
	}
}

// getOrCreateBob retrieves or attaches a MovementBobComponent to the entity.
func (s *MovementBobSystem) getOrCreateBob(entity *Entity) *MovementBobComponent {
	if comp, ok := entity.GetComponent("movement_bob"); ok {
		if bob, ok := comp.(*MovementBobComponent); ok {
			return bob
		}
	}
	bob := &MovementBobComponent{}
	entity.AddComponent(bob)
	return bob
}
