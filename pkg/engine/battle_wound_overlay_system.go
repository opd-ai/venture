// Package engine provides the BattleWoundOverlaySystem which assigns genre-aware
// wound visual overlays to entities based on cumulative damage. It monitors
// HealthComponent decreases and progressively increases wound severity through
// four tiers: Scratched, Wounded, Bloodied, Critical. Genre palettes drive
// wound color and style: fantasy uses crimson slashes, horror uses dark blood
// spatters, sci-fi uses cyan burn marks, cyberpunk uses magenta circuit sparks,
// and post-apocalyptic uses rust-brown gashes.
package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// WoundSeverity represents the visual severity of battle wounds.
type WoundSeverity int

const (
	// WoundNone indicates no visible wounds.
	WoundNone WoundSeverity = iota
	// WoundScratched shows minor surface damage (health > 75%).
	WoundScratched
	// WoundWounded shows moderate damage (health 50-75%).
	WoundWounded
	// WoundBloodied shows severe damage (health 25-50%).
	WoundBloodied
	// WoundCritical shows near-fatal damage (health < 25%).
	WoundCritical
)

// BattleWoundOverlayComponent stores per-entity wound overlay state.
// Pure data component — all logic resides in BattleWoundOverlaySystem.
type BattleWoundOverlayComponent struct {
	// Current wound severity tier
	Severity WoundSeverity

	// Wound color (RGB, 0.0-1.0) derived from genre palette
	WoundR float64
	WoundG float64
	WoundB float64

	// Overlay opacity (0.0-1.0), scales with severity
	Opacity float64

	// Number of wound marks to render (1-8)
	MarkCount int

	// Wound mark positions as normalized offsets from entity center (0.0-1.0)
	MarkOffsetsX [8]float64
	MarkOffsetsY [8]float64

	// Mark sizes in pixels (1.0-4.0)
	MarkSizes [8]float64

	// Pulse phase for critical wound glow animation (radians)
	PulsePhase float64

	// Whether the overlay needs re-rendering
	Dirty bool
}

// Type returns the component type identifier.
func (c *BattleWoundOverlayComponent) Type() string {
	return "battle_wound_overlay"
}

// genreWoundPalette holds genre-specific wound visual parameters.
type genreWoundPalette struct {
	R, G, B       float64 // Primary wound color
	CritR, CritG, CritB float64 // Critical wound glow color
	PulseSpeed    float64 // Critical pulse speed (cycles/sec)
}

// BattleWoundOverlaySystem assigns and updates genre-aware wound overlays
// on entities that take damage. It tracks health changes per entity and
// adjusts wound severity, mark placement, and genre-driven color.
type BattleWoundOverlaySystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	genreID string

	// Track previous health to detect damage
	prevHealth map[uint64]float64

	// Cleanup timer for removed entities
	cleanupTimer    float64
	cleanupInterval float64

	palettes map[string]genreWoundPalette
}

// NewBattleWoundOverlaySystem creates a new battle wound overlay system.
func NewBattleWoundOverlaySystem(world *World, seed int64) *BattleWoundOverlaySystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "battle_wound_overlay")
	}

	sys := &BattleWoundOverlaySystem{
		world:           world,
		logger:          logEntry,
		rng:             rand.New(rand.NewSource(seed)),
		genreID:         "fantasy",
		prevHealth:      make(map[uint64]float64),
		cleanupTimer:    0.0,
		cleanupInterval: 5.0,
		palettes:        buildWoundPalettes(),
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("battle wound overlay system created")
	}
	return sys
}

// SetGenre configures the active genre for wound visuals.
func (s *BattleWoundOverlaySystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for battle wound overlay")
	}
}

// Update processes all entities, detecting health changes and updating wound overlays.
func (s *BattleWoundOverlaySystem) Update(entities []*Entity, deltaTime float64) {
	if s.world == nil {
		return
	}

	// Periodic cleanup of tracked entities that no longer exist
	s.cleanupTimer += deltaTime
	if s.cleanupTimer >= s.cleanupInterval {
		s.cleanupTimer = 0
		s.cleanupTrackedEntities(entities)
	}

	palette := s.getActivePalette()

	for _, entity := range entities {
		health := entity.GetHealth()
		if health == nil || health.Max <= 0 {
			continue
		}

		healthRatio := health.Current / health.Max
		if healthRatio < 0 {
			healthRatio = 0
		}
		if healthRatio > 1.0 {
			healthRatio = 1.0
		}

		prev, tracked := s.prevHealth[entity.ID]
		s.prevHealth[entity.ID] = health.Current

		newSeverity := severityFromHealthRatio(healthRatio)

		comp := s.getOrCreateComponent(entity)

		// Detect damage event — add new wound marks
		if tracked && prev > health.Current {
			s.onDamageTaken(comp, entity, healthRatio, palette)
		}

		// Update severity and visual parameters
		if comp.Severity != newSeverity {
			comp.Severity = newSeverity
			if newSeverity == WoundNone {
				comp.Opacity = 0
				comp.MarkCount = 0
			} else {
				s.updateSeverityVisuals(comp, palette, healthRatio)
			}
			comp.Dirty = true
		}

		// Animate critical wound pulse
		if comp.Severity == WoundCritical {
			comp.PulsePhase += deltaTime * palette.PulseSpeed * 2 * math.Pi
			if comp.PulsePhase > 2*math.Pi {
				comp.PulsePhase -= 2 * math.Pi
			}
			// Modulate opacity with pulse
			baseOpacity := 0.7 + 0.3*(1.0-healthRatio)
			comp.Opacity = baseOpacity + 0.15*math.Sin(comp.PulsePhase)
			if comp.Opacity > 1.0 {
				comp.Opacity = 1.0
			}
		}

		// Clear wounds when fully healed
		if healthRatio >= 1.0 && comp.Severity != WoundNone {
			comp.Severity = WoundNone
			comp.Opacity = 0
			comp.MarkCount = 0
			comp.Dirty = true
		}
	}
}

// severityFromHealthRatio maps health percentage to wound severity.
func severityFromHealthRatio(ratio float64) WoundSeverity {
	switch {
	case ratio >= 1.0:
		return WoundNone
	case ratio > 0.75:
		return WoundScratched
	case ratio > 0.50:
		return WoundWounded
	case ratio > 0.25:
		return WoundBloodied
	default:
		return WoundCritical
	}
}

// getOrCreateComponent retrieves or creates the wound overlay component on an entity.
func (s *BattleWoundOverlaySystem) getOrCreateComponent(entity *Entity) *BattleWoundOverlayComponent {
	if comp, ok := entity.GetComponent("battle_wound_overlay"); ok {
		if wc, ok := comp.(*BattleWoundOverlayComponent); ok {
			return wc
		}
	}
	comp := &BattleWoundOverlayComponent{}
	entity.AddComponent(comp)
	return comp
}

// onDamageTaken adds new wound marks when damage is detected.
func (s *BattleWoundOverlaySystem) onDamageTaken(comp *BattleWoundOverlayComponent, entity *Entity, healthRatio float64, palette genreWoundPalette) {
	// Add 1-2 marks per damage event, up to max 8
	newMarks := 1
	if healthRatio < 0.5 {
		newMarks = 2
	}

	for i := 0; i < newMarks && comp.MarkCount < 8; i++ {
		idx := comp.MarkCount
		// Deterministic placement using rng
		comp.MarkOffsetsX[idx] = s.rng.Float64()*0.6 + 0.2 // 0.2–0.8 range
		comp.MarkOffsetsY[idx] = s.rng.Float64()*0.6 + 0.2
		comp.MarkSizes[idx] = 1.0 + s.rng.Float64()*2.0 // 1–3 pixels
		if healthRatio < 0.25 {
			comp.MarkSizes[idx] += 1.0 // Larger marks at critical health
		}
		comp.MarkCount++
	}
	comp.Dirty = true
}

// updateSeverityVisuals sets color and opacity based on severity and genre.
func (s *BattleWoundOverlaySystem) updateSeverityVisuals(comp *BattleWoundOverlayComponent, palette genreWoundPalette, healthRatio float64) {
	switch comp.Severity {
	case WoundScratched:
		comp.WoundR = palette.R * 0.6
		comp.WoundG = palette.G * 0.6
		comp.WoundB = palette.B * 0.6
		comp.Opacity = 0.25
	case WoundWounded:
		comp.WoundR = palette.R * 0.8
		comp.WoundG = palette.G * 0.8
		comp.WoundB = palette.B * 0.8
		comp.Opacity = 0.45
	case WoundBloodied:
		comp.WoundR = palette.R
		comp.WoundG = palette.G
		comp.WoundB = palette.B
		comp.Opacity = 0.65
	case WoundCritical:
		// Blend between wound color and critical glow
		t := 0.5 + 0.5*(1.0-healthRatio/0.25)
		comp.WoundR = palette.R*(1-t) + palette.CritR*t
		comp.WoundG = palette.G*(1-t) + palette.CritG*t
		comp.WoundB = palette.B*(1-t) + palette.CritB*t
		comp.Opacity = 0.8
	default:
		comp.Opacity = 0
	}
}

// getActivePalette returns the wound palette for the current genre.
func (s *BattleWoundOverlaySystem) getActivePalette() genreWoundPalette {
	if p, ok := s.palettes[s.genreID]; ok {
		return p
	}
	return s.palettes["fantasy"]
}

// cleanupTrackedEntities removes stale entries from the prevHealth map.
func (s *BattleWoundOverlaySystem) cleanupTrackedEntities(entities []*Entity) {
	active := make(map[uint64]bool, len(entities))
	for _, e := range entities {
		active[e.ID] = true
	}
	for id := range s.prevHealth {
		if !active[id] {
			delete(s.prevHealth, id)
		}
	}
}

// buildWoundPalettes creates genre-specific wound color palettes.
func buildWoundPalettes() map[string]genreWoundPalette {
	return map[string]genreWoundPalette{
		"fantasy": {
			R: 0.75, G: 0.12, B: 0.12, // Crimson slashes
			CritR: 1.0, CritG: 0.2, CritB: 0.1, // Bright red glow
			PulseSpeed: 1.0,
		},
		"horror": {
			R: 0.45, G: 0.05, B: 0.08, // Dark blood spatters
			CritR: 0.6, CritG: 0.0, CritB: 0.0, // Deep crimson pulse
			PulseSpeed: 0.7,
		},
		"scifi": {
			R: 0.2, G: 0.7, B: 0.85, // Cyan burn marks
			CritR: 0.1, CritG: 0.9, CritB: 1.0, // Electric cyan glow
			PulseSpeed: 1.4,
		},
		"cyberpunk": {
			R: 0.85, G: 0.15, B: 0.75, // Magenta circuit sparks
			CritR: 1.0, CritG: 0.1, CritB: 0.9, // Neon magenta pulse
			PulseSpeed: 1.6,
		},
		"postapoc": {
			R: 0.65, G: 0.35, B: 0.15, // Rust-brown gashes
			CritR: 0.8, CritG: 0.3, CritB: 0.1, // Infected amber glow
			PulseSpeed: 0.8,
		},
	}
}
