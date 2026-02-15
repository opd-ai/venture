// Package engine provides the TimeOfDayShadowDirectionSystem which bridges
// TimeOfDayLightingSystem with DropShadowComponent to create directional
// shadows that shift based on the simulated sun position during day/night cycles.
// Dawn casts long shadows to the west, noon has short centered shadows,
// dusk casts long shadows to the east, and night produces faint ambient shadows.
package engine

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/sirupsen/logrus"
)

// shadowDirectionPreset holds genre-specific shadow direction parameters.
type shadowDirectionPreset struct {
	StretchScale  float64 // Multiplier for dawn/dusk shadow stretching (1.0 = normal)
	NightOpacity  float64 // Shadow opacity during night (0.0–1.0)
	DawnDuskBoost float64 // Extra stretch at dawn/dusk (added to base)
}

// TimeOfDayShadowDirectionSystem modifies DropShadowComponent offsets and
// dimensions based on the current time of day, simulating a sun arc that
// moves east→overhead→west across the sky.
//
// Shadow behavior by time of day:
//   - Dawn: Shadows stretch to the west (negative OffsetX), long and soft
//   - Day: Shadows are short, directly below the entity, strongest opacity
//   - Dusk: Shadows stretch to the east (positive OffsetX), long and soft
//   - Night: Shadows are faint, nearly centered (ambient moonlight)
//
// Genre-specific modifiers:
//   - Horror: 40% longer shadows, higher night opacity (eerie)
//   - Cyberpunk: 10% longer, neon glow keeps night shadows visible
//   - Sci-fi: Normal stretch, slightly reduced night opacity
//   - Postapoc: 20% longer, darker night shadows (no artificial light)
//   - Fantasy: Baseline behavior
type TimeOfDayShadowDirectionSystem struct {
	world  *World
	rng    *rand.Rand
	logger *logrus.Entry

	genreID string
	preset  shadowDirectionPreset

	lightingSystem *TimeOfDayLightingSystem

	updateInterval float64
	timeSinceCheck float64
	lastTimeOfDay  palette.TimeOfDay
}

// NewTimeOfDayShadowDirectionSystem creates a new time-of-day shadow direction system.
func NewTimeOfDayShadowDirectionSystem(world *World, seed int64) *TimeOfDayShadowDirectionSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "timeofday_shadow_direction")
	}

	sys := &TimeOfDayShadowDirectionSystem{
		world:          world,
		rng:            rand.New(rand.NewSource(seed)),
		logger:         logEntry,
		genreID:        "fantasy",
		updateInterval: 0.5, // Check twice per second (shadows shift smoothly)
		lastTimeOfDay:  palette.TimeOfDayDay,
	}
	sys.preset = sys.getPreset(sys.genreID)

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("TimeOfDayShadowDirectionSystem created")
	}
	return sys
}

// SetLightingSystem sets the time-of-day lighting system reference.
func (s *TimeOfDayShadowDirectionSystem) SetLightingSystem(ls *TimeOfDayLightingSystem) {
	s.lightingSystem = ls
	if s.logger != nil {
		s.logger.Debug("lighting system linked")
	}
}

// SetGenre configures genre-aware shadow direction parameters.
func (s *TimeOfDayShadowDirectionSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getPreset(genreID)
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for shadow direction")
	}
}

// getPreset returns genre-specific shadow direction parameters.
func (s *TimeOfDayShadowDirectionSystem) getPreset(genreID string) shadowDirectionPreset {
	switch genreID {
	case "horror":
		return shadowDirectionPreset{StretchScale: 1.4, NightOpacity: 0.30, DawnDuskBoost: 0.3}
	case "cyberpunk":
		return shadowDirectionPreset{StretchScale: 1.1, NightOpacity: 0.22, DawnDuskBoost: 0.1}
	case "sci-fi", "scifi":
		return shadowDirectionPreset{StretchScale: 1.0, NightOpacity: 0.12, DawnDuskBoost: 0.0}
	case "post-apocalyptic", "postapoc":
		return shadowDirectionPreset{StretchScale: 1.2, NightOpacity: 0.10, DawnDuskBoost: 0.15}
	case "fantasy":
		return shadowDirectionPreset{StretchScale: 1.0, NightOpacity: 0.15, DawnDuskBoost: 0.0}
	default:
		return shadowDirectionPreset{StretchScale: 1.0, NightOpacity: 0.15, DawnDuskBoost: 0.0}
	}
}

// Update processes entities with DropShadowComponent and adjusts offset/size
// based on the current time of day.
func (s *TimeOfDayShadowDirectionSystem) Update(entities []*Entity, deltaTime float64) {
	if s.lightingSystem == nil {
		return
	}

	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	currentTime := s.lightingSystem.GetCurrentTimeOfDay()
	timeChanged := currentTime != s.lastTimeOfDay

	if timeChanged && s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"from": s.lastTimeOfDay.String(),
			"to":   currentTime.String(),
		}).Debug("time of day changed, updating shadow directions")
	}
	s.lastTimeOfDay = currentTime

	offsetX, offsetYScale, opacityMult, widthScale, heightScale := s.computeShadowParams(currentTime)

	for _, entity := range entities {
		comp, _ := entity.GetComponent("drop_shadow")
		shadow, ok := comp.(*DropShadowComponent)
		if !ok || shadow == nil || !shadow.Enabled {
			continue
		}

		// Determine entity height for Y offset scaling
		entityHeight := 32.0
		if col := entity.GetCollider(); col != nil {
			entityHeight = col.Height
		} else if spr := entity.GetSprite(); spr != nil {
			entityHeight = spr.Height
		}

		// Apply directional offset
		shadow.OffsetX = offsetX * (entityHeight / 32.0) * s.preset.StretchScale
		shadow.OffsetY = entityHeight * 0.4 * offsetYScale

		// Scale shadow dimensions for stretch effect
		baseWidth := entityHeight * 0.70
		baseHeight := entityHeight * 0.25
		if baseWidth < 6 {
			baseWidth = 6
		}
		if baseHeight < 3 {
			baseHeight = 3
		}
		shadow.ShadowWidth = baseWidth * widthScale
		shadow.ShadowHeight = baseHeight * heightScale

		// Modulate opacity
		shadow.Opacity = clampFloat(0.35*opacityMult, 0.05, 0.55)
	}
}

// computeShadowParams returns (offsetX, offsetYScale, opacityMultiplier,
// widthScale, heightScale) based on the current time of day.
//
// Sun arc: East (dawn) → overhead (day) → West (dusk) → absent (night).
// OffsetX: negative = shadow cast west, positive = shadow cast east.
// offsetYScale: >1.0 = shadow further below (stretched), 1.0 = normal.
func (s *TimeOfDayShadowDirectionSystem) computeShadowParams(tod palette.TimeOfDay) (offsetX, offsetYScale, opacityMult, widthScale, heightScale float64) {
	stretch := s.preset.StretchScale
	dawnDuskBoost := s.preset.DawnDuskBoost

	switch tod {
	case palette.TimeOfDayDawn:
		// Sun in the east → shadows stretch west (negative X)
		shadowLen := (6.0 + dawnDuskBoost*4.0) * stretch
		return -shadowLen, 1.3 + dawnDuskBoost, 0.75, 1.0 + 0.2*stretch, 1.5 + 0.3*stretch

	case palette.TimeOfDayDay:
		// Sun overhead → shadows short, directly below
		return 0.0, 1.0, 1.0, 1.0, 1.0

	case palette.TimeOfDayDusk:
		// Sun in the west → shadows stretch east (positive X)
		shadowLen := (6.0 + dawnDuskBoost*4.0) * stretch
		return shadowLen, 1.3 + dawnDuskBoost, 0.70, 1.0 + 0.2*stretch, 1.5 + 0.3*stretch

	case palette.TimeOfDayNight:
		// No directional sun → faint ambient shadow below
		nightOp := s.preset.NightOpacity / 0.35 // Scale relative to base 0.35
		return 0.0, 0.9, math.Max(nightOp, 0.15), 0.85, 0.85

	default:
		return 0.0, 1.0, 1.0, 1.0, 1.0
	}
}
