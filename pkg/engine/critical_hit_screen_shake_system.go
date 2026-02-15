// Package engine provides the CriticalHitScreenShakeSystem for visceral combat feedback.
// This system connects CombatSystem critical hit events with CameraSystem screen shake,
// providing genre-aware camera shake intensity and duration when critical hits land.
package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// CriticalHitScreenShakeSystem triggers camera screen shake on critical hits.
// It bridges CombatSystem critical hit callbacks with CameraSystem.ShakeAdvanced
// to deliver genre-aware tactile feedback proportional to damage dealt.
type CriticalHitScreenShakeSystem struct {
	world        *World
	cameraSystem *CameraSystem
	genreID      string
	seed         int64
	rng          *rand.Rand
	logger       *logrus.Entry

	// Genre-tuned parameters
	baseIntensity float64 // Base shake intensity (pixels)
	baseDuration  float64 // Base shake duration (seconds)
	damageScale   float64 // How much damage scales intensity (0-1)

	// Pending shakes from callbacks (processed in Update)
	pendingShakes []critShakeEvent
}

// critShakeEvent records a critical hit that needs a screen shake.
type critShakeEvent struct {
	damage    float64
	attackerX float64
	attackerY float64
	targetX   float64
	targetY   float64
}

// NewCriticalHitScreenShakeSystem creates a new critical hit screen shake system.
func NewCriticalHitScreenShakeSystem(world *World, seed int64) *CriticalHitScreenShakeSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "critical_hit_screen_shake")
		logEntry.Debug("critical hit screen shake system created")
	}

	return &CriticalHitScreenShakeSystem{
		world:         world,
		seed:          seed,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logEntry,
		baseIntensity: 4.0,
		baseDuration:  0.15,
		damageScale:   0.02,
		pendingShakes: make([]critShakeEvent, 0, 4),
	}
}

// SetCameraSystem sets the camera system for triggering shakes.
func (s *CriticalHitScreenShakeSystem) SetCameraSystem(cs *CameraSystem) {
	s.cameraSystem = cs
	if s.logger != nil {
		s.logger.Debug("camera system linked")
	}
}

// SetGenre sets the genre ID for genre-aware shake tuning.
func (s *CriticalHitScreenShakeSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.applyGenrePresets()
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// applyGenrePresets tunes shake parameters based on genre.
func (s *CriticalHitScreenShakeSystem) applyGenrePresets() {
	switch s.genreID {
	case "horror":
		// Horror: heavy, lingering shakes for dread
		s.baseIntensity = 6.0
		s.baseDuration = 0.25
		s.damageScale = 0.03
	case "cyberpunk":
		// Cyberpunk: sharp, snappy shakes
		s.baseIntensity = 5.0
		s.baseDuration = 0.10
		s.damageScale = 0.025
	case "fantasy":
		// Fantasy: moderate, heroic feel
		s.baseIntensity = 4.0
		s.baseDuration = 0.15
		s.damageScale = 0.02
	case "scifi":
		// Sci-fi: crisp, precise shakes
		s.baseIntensity = 3.5
		s.baseDuration = 0.12
		s.damageScale = 0.02
	case "postapoc":
		// Post-apocalyptic: heavy, raw shakes
		s.baseIntensity = 5.5
		s.baseDuration = 0.20
		s.damageScale = 0.025
	default:
		s.baseIntensity = 4.0
		s.baseDuration = 0.15
		s.damageScale = 0.02
	}
}

// OnCriticalHit is the callback invoked by CombatSystem on critical hits.
// It queues a shake event to be processed in the next Update cycle.
func (s *CriticalHitScreenShakeSystem) OnCriticalHit(attacker, target *Entity, damage float64) {
	var ax, ay, tx, ty float64
	if pos, ok := attacker.GetComponent("position"); ok {
		if p, ok := pos.(*PositionComponent); ok {
			ax, ay = p.X, p.Y
		}
	}
	if pos, ok := target.GetComponent("position"); ok {
		if p, ok := pos.(*PositionComponent); ok {
			tx, ty = p.X, p.Y
		}
	}
	s.pendingShakes = append(s.pendingShakes, critShakeEvent{
		damage:    damage,
		attackerX: ax, attackerY: ay,
		targetX: tx, targetY: ty,
	})
}

// Update processes pending critical hit shake events.
func (s *CriticalHitScreenShakeSystem) Update(entities []*Entity, deltaTime float64) {
	if s.cameraSystem == nil || len(s.pendingShakes) == 0 {
		return
	}

	// Process strongest shake this frame to avoid stacking
	var strongest critShakeEvent
	maxDamage := 0.0
	for _, evt := range s.pendingShakes {
		if evt.damage > maxDamage {
			maxDamage = evt.damage
			strongest = evt
		}
	}
	s.pendingShakes = s.pendingShakes[:0]

	// Calculate intensity: base + scaled by damage, clamped to max
	intensity := s.baseIntensity + strongest.damage*s.damageScale
	maxIntensity := s.baseIntensity * 3.0
	if intensity > maxIntensity {
		intensity = maxIntensity
	}

	// Calculate duration: base + small bonus for high damage
	duration := s.baseDuration
	if strongest.damage > 50.0 {
		duration += 0.05
	}
	if strongest.damage > 100.0 {
		duration += 0.05
	}
	maxDuration := s.baseDuration * 2.5
	if duration > maxDuration {
		duration = maxDuration
	}

	// Distance attenuation: reduce shake for far-away crits
	if s.cameraSystem.activeCamera != nil {
		if camPos, ok := s.cameraSystem.activeCamera.GetComponent("position"); ok {
			if cp, ok := camPos.(*PositionComponent); ok {
				dx := cp.X - strongest.targetX
				dy := cp.Y - strongest.targetY
				dist := math.Sqrt(dx*dx + dy*dy)
				// Full intensity within 200px, linear falloff to 600px, zero beyond
				if dist > 600.0 {
					return // Too far, skip shake
				}
				if dist > 200.0 {
					attenuation := 1.0 - (dist-200.0)/400.0
					intensity *= attenuation
				}
			}
		}
	}

	s.cameraSystem.ShakeAdvanced(intensity, duration)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"damage":    strongest.damage,
			"intensity": intensity,
			"duration":  duration,
			"genre":     s.genreID,
		}).Debug("critical hit screen shake triggered")
	}
}
