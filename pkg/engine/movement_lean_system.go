// Package engine provides the MovementLeanSystem for directional sprite lean.
// When entities move, their sprites shift horizontally in the movement direction,
// creating a subtle momentum-lean effect. Genre presets control amplitude and
// smoothing: fantasy uses a gentle lean, horror a lurching stagger, sci-fi
// a tight precise offset, cyberpunk a sharp snap, and post-apocalyptic a heavy drag.
package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// MovementLeanComponent stores per-entity lean state and render offset.
// Pure data — all logic lives in MovementLeanSystem.
type MovementLeanComponent struct {
	// OffsetX is the horizontal pixel offset applied during rendering.
	OffsetX float64
	// TargetOffsetX is the desired offset based on current velocity.
	TargetOffsetX float64
	// Active indicates the entity currently has a non-zero lean.
	Active bool
}

// Type returns the component type identifier.
func (m *MovementLeanComponent) Type() string { return "movement_lean" }

// movementLeanPreset holds genre-specific lean configuration.
type movementLeanPreset struct {
	MaxLean  float64 // Max horizontal displacement in pixels (1–4)
	Smoothin float64 // Interpolation speed toward target (higher = snappier)
	Dampout  float64 // Interpolation speed toward zero when stopping
}

// MovementLeanSystem applies a horizontal offset to moving entities based on
// their velocity direction, creating a momentum-lean visual effect. It reads
// VelocityComponent speed/direction and writes OffsetX into MovementLeanComponent
// for the render system to consume.
type MovementLeanSystem struct {
	world   *World
	seed    int64
	rng     *rand.Rand
	logger  *logrus.Entry
	genreID string
	preset  movementLeanPreset

	// Squared speed threshold below which lean dampens toward zero.
	minSpeedSq float64
}

// NewMovementLeanSystem creates a movement lean system with default fantasy preset.
func NewMovementLeanSystem(world *World, seed int64) *MovementLeanSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "movement_lean",
	})

	sys := &MovementLeanSystem{
		world:      world,
		seed:       seed,
		rng:        rand.New(rand.NewSource(seed)),
		logger:     logger,
		genreID:    "fantasy",
		minSpeedSq: 100.0, // 10 px/s threshold
	}
	sys.preset = sys.getGenrePreset("fantasy")
	logger.Debug("movement lean system created")
	return sys
}

// SetGenre configures genre-specific lean parameters.
func (s *MovementLeanSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getGenrePreset(genreID)
	s.logger.WithField("genre", genreID).Debug("genre preset applied")
}

// getGenrePreset returns lean configuration for the given genre.
func (s *MovementLeanSystem) getGenrePreset(genreID string) movementLeanPreset {
	switch genreID {
	case "fantasy":
		return movementLeanPreset{MaxLean: 2.0, Smoothin: 8.0, Dampout: 6.0}
	case "horror":
		return movementLeanPreset{MaxLean: 3.0, Smoothin: 5.0, Dampout: 3.0}
	case "scifi":
		return movementLeanPreset{MaxLean: 1.5, Smoothin: 12.0, Dampout: 10.0}
	case "cyberpunk":
		return movementLeanPreset{MaxLean: 2.5, Smoothin: 14.0, Dampout: 8.0}
	case "postapoc":
		return movementLeanPreset{MaxLean: 3.5, Smoothin: 4.0, Dampout: 3.0}
	default:
		return movementLeanPreset{MaxLean: 2.0, Smoothin: 8.0, Dampout: 6.0}
	}
}

// Update processes all entities with velocity, computing horizontal lean offset.
func (s *MovementLeanSystem) Update(entities []*Entity, deltaTime float64) {
	if deltaTime <= 0 {
		return
	}
	// Clamp large delta to prevent overshooting
	if deltaTime > 0.1 {
		deltaTime = 0.1
	}

	for _, entity := range entities {
		vel := entity.GetVelocity()
		if vel == nil {
			continue
		}

		lean := s.getOrCreateLean(entity)
		if lean == nil {
			continue
		}

		speedSq := vel.VX*vel.VX + vel.VY*vel.VY

		if speedSq > s.minSpeedSq {
			// Compute normalized horizontal direction component
			speed := math.Sqrt(speedSq)
			normalizedX := vel.VX / speed

			// Scale lean by speed factor (0–1 range mapped to 0–MaxLean)
			speedFactor := math.Min(speed/200.0, 1.0)
			lean.TargetOffsetX = normalizedX * s.preset.MaxLean * speedFactor
			lean.Active = true

			// Smooth interpolation toward target
			lerpFactor := 1.0 - math.Exp(-s.preset.Smoothin*deltaTime)
			lean.OffsetX += (lean.TargetOffsetX - lean.OffsetX) * lerpFactor
		} else {
			// Damp toward zero when stationary
			lean.TargetOffsetX = 0
			lerpFactor := 1.0 - math.Exp(-s.preset.Dampout*deltaTime)
			lean.OffsetX += (0 - lean.OffsetX) * lerpFactor

			// Snap to zero when close enough
			if math.Abs(lean.OffsetX) < 0.01 {
				lean.OffsetX = 0
				lean.Active = false
			}
		}
	}
}

// getOrCreateLean retrieves or attaches a MovementLeanComponent to the entity.
func (s *MovementLeanSystem) getOrCreateLean(entity *Entity) *MovementLeanComponent {
	if comp, ok := entity.GetComponent("movement_lean"); ok {
		if lean, ok := comp.(*MovementLeanComponent); ok {
			return lean
		}
	}
	lean := &MovementLeanComponent{}
	entity.AddComponent(lean)
	return lean
}
