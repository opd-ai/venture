// Package engine provides the EntityThreatIndicatorSystem which assigns
// genre-aware colored ring parameters to AI entities based on their relative
// threat level compared to the nearest player. Players see at a glance whether
// a creature is trivial (grey), easy (green), fair (yellow), challenging
// (orange), or dangerous (red). Genre palettes shift the hue: horror uses
// desaturated blood tones, cyberpunk uses neon, sci-fi uses holographic teal.
// Ring opacity pulses slowly for challenging+ threats to draw attention.
package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// ThreatIndicatorComponent stores the computed visual ring parameters for an
// entity. The render pipeline reads these to draw a colored ellipse beneath
// the entity sprite at ZIndex 1 (just above ground shadows).
type ThreatIndicatorComponent struct {
	// Ring color (RGB, 0.0-1.0)
	RingR float64
	RingG float64
	RingB float64

	// Base opacity before pulse modulation (0.0-1.0)
	BaseOpacity float64

	// Current opacity after pulse (0.0-1.0)
	CurrentOpacity float64

	// Pulse speed in Hz (0 = no pulse)
	PulseSpeed float64

	// Accumulated phase in radians
	PulsePhase float64

	// Ring radius in pixels (half-width of ellipse)
	RingRadius float64

	// ThreatTier cached for external queries (0=trivial .. 4=dangerous)
	ThreatTier int

	// Whether the indicator is active
	Enabled bool
}

// Type returns the component type identifier.
func (t *ThreatIndicatorComponent) Type() string {
	return "threat_indicator"
}

// NewThreatIndicatorComponent creates a disabled default component.
func NewThreatIndicatorComponent() *ThreatIndicatorComponent {
	return &ThreatIndicatorComponent{
		BaseOpacity: 0.0,
		RingRadius:  8.0,
		Enabled:     false,
	}
}

// threatTierPreset defines color + pulse for one threat tier.
type threatTierPreset struct {
	R, G, B    float64
	Opacity    float64
	PulseSpeed float64
}

// genreThreatPalette holds per-tier presets for one genre.
type genreThreatPalette struct {
	Tiers [5]threatTierPreset // 0=trivial,1=easy,2=fair,3=challenging,4=dangerous
}

// EntityThreatIndicatorSystem computes relative threat and assigns ring
// visuals. Threat is derived from ExperienceComponent.Level when available,
// falling back to HealthComponent.Max as a proxy. Updates are throttled to
// twice per second; pulse animation runs every frame.
type EntityThreatIndicatorSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	genreID     string
	lastGenreID string
	initialized bool

	updateInterval float64
	timeSinceCheck float64

	palettes map[string]genreThreatPalette
}

// NewEntityThreatIndicatorSystem creates a new threat indicator system.
func NewEntityThreatIndicatorSystem(world *World, seed int64) *EntityThreatIndicatorSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "entity_threat_indicator")
	}

	sys := &EntityThreatIndicatorSystem{
		world:          world,
		logger:         logEntry,
		rng:            rand.New(rand.NewSource(seed)),
		genreID:        "fantasy",
		updateInterval: 0.5,
		palettes:       buildThreatPalettes(),
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("entity threat indicator system created")
	}
	return sys
}

// SetGenre configures the active genre for threat ring colors.
func (s *EntityThreatIndicatorSystem) SetGenre(genreID string) {
	s.genreID = genreID
}

// Update drives pulse animation every frame and re-evaluates threat tiers
// on a throttled interval.
func (s *EntityThreatIndicatorSystem) Update(entities []*Entity, deltaTime float64) {
	s.updatePulse(entities, deltaTime)

	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	genreChanged := s.genreID != s.lastGenreID || !s.initialized
	s.lastGenreID = s.genreID
	s.initialized = true

	playerLevel := s.findPlayerLevel(entities)
	if playerLevel < 0 && !genreChanged {
		return
	}
	if playerLevel < 0 {
		playerLevel = 1
	}

	s.assignThreatIndicators(entities, playerLevel)
}

// updatePulse modulates CurrentOpacity using a sine wave.
func (s *EntityThreatIndicatorSystem) updatePulse(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		comp, ok := entity.GetComponent("threat_indicator")
		if !ok {
			continue
		}
		ti, ok := comp.(*ThreatIndicatorComponent)
		if !ok || !ti.Enabled || ti.PulseSpeed <= 0 {
			continue
		}

		ti.PulsePhase += deltaTime * ti.PulseSpeed * 2 * math.Pi
		if ti.PulsePhase > 2*math.Pi {
			ti.PulsePhase -= 2 * math.Pi
		}

		pulse := math.Sin(ti.PulsePhase) * 0.15
		ti.CurrentOpacity = clampThreat(ti.BaseOpacity + pulse)
	}
}

// findPlayerLevel returns the level of the first player entity found, or -1.
func (s *EntityThreatIndicatorSystem) findPlayerLevel(entities []*Entity) int {
	for _, entity := range entities {
		if !entity.HasComponent("input") {
			continue
		}
		if comp, ok := entity.GetComponent("experience"); ok {
			if xp, ok := comp.(*ExperienceComponent); ok {
				return xp.Level
			}
		}
		// Fallback: estimate level from max health (100 HP ≈ level 1)
		if comp, ok := entity.GetComponent("health"); ok {
			if hc, ok := comp.(*HealthComponent); ok {
				return int(math.Max(1, hc.Max/100.0))
			}
		}
	}
	return -1
}

// assignThreatIndicators evaluates each AI entity's threat relative to the
// player and sets ring color/pulse accordingly.
func (s *EntityThreatIndicatorSystem) assignThreatIndicators(entities []*Entity, playerLevel int) {
	palette, ok := s.palettes[s.genreID]
	if !ok {
		palette = s.palettes["fantasy"]
	}

	count := 0
	for _, entity := range entities {
		// Only AI entities with a visible sprite
		if !entity.HasComponent("ai") || !entity.HasComponent("sprite") {
			continue
		}
		if entity.HasComponent("input") {
			continue
		}

		entityLevel := s.estimateEntityLevel(entity)
		tier := computeThreatTier(playerLevel, entityLevel)
		preset := palette.Tiers[tier]

		ti := s.getOrCreateComponent(entity)
		ti.Enabled = true
		ti.ThreatTier = tier

		variation := s.rng.Float64()*0.06 - 0.03
		ti.RingR = clampThreat(preset.R + variation)
		ti.RingG = clampThreat(preset.G + variation)
		ti.RingB = clampThreat(preset.B + variation)
		ti.BaseOpacity = preset.Opacity
		ti.CurrentOpacity = preset.Opacity
		ti.PulseSpeed = preset.PulseSpeed
		ti.RingRadius = 8.0 + float64(tier)*2.0

		// Randomize initial phase
		ti.PulsePhase = s.rng.Float64() * 2 * math.Pi

		count++
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"genre":            s.genreID,
			"player_level":     playerLevel,
			"entities_labeled": count,
		}).Debug("threat indicators assigned")
	}
}

// estimateEntityLevel returns the entity's level from ExperienceComponent or
// a health-based estimate.
func (s *EntityThreatIndicatorSystem) estimateEntityLevel(entity *Entity) int {
	if comp, ok := entity.GetComponent("experience"); ok {
		if xp, ok := comp.(*ExperienceComponent); ok {
			return xp.Level
		}
	}
	if comp, ok := entity.GetComponent("health"); ok {
		if hc, ok := comp.(*HealthComponent); ok {
			return int(math.Max(1, hc.Max/100.0))
		}
	}
	return 1
}

// computeThreatTier maps a level difference to a 0-4 tier index.
//
//	diff <= -5 → 0 (trivial)
//	diff -3..-4 → 1 (easy)
//	diff -2..+2 → 2 (fair)
//	diff +3..+4 → 3 (challenging)
//	diff >= +5 → 4 (dangerous)
func computeThreatTier(playerLevel, entityLevel int) int {
	diff := entityLevel - playerLevel
	switch {
	case diff <= -5:
		return 0
	case diff <= -3:
		return 1
	case diff <= 2:
		return 2
	case diff <= 4:
		return 3
	default:
		return 4
	}
}

// getOrCreateComponent lazily creates the ThreatIndicatorComponent.
func (s *EntityThreatIndicatorSystem) getOrCreateComponent(entity *Entity) *ThreatIndicatorComponent {
	comp, ok := entity.GetComponent("threat_indicator")
	if ok {
		if ti, ok := comp.(*ThreatIndicatorComponent); ok {
			return ti
		}
	}
	ti := NewThreatIndicatorComponent()
	entity.AddComponent(ti)
	return ti
}

// buildThreatPalettes returns genre-specific threat tier color presets.
func buildThreatPalettes() map[string]genreThreatPalette {
	return map[string]genreThreatPalette{
		"fantasy": {Tiers: [5]threatTierPreset{
			{R: 0.50, G: 0.50, B: 0.50, Opacity: 0.25, PulseSpeed: 0.0},  // trivial: grey
			{R: 0.30, G: 0.80, B: 0.30, Opacity: 0.35, PulseSpeed: 0.0},  // easy: green
			{R: 0.90, G: 0.85, B: 0.20, Opacity: 0.45, PulseSpeed: 0.0},  // fair: yellow
			{R: 0.95, G: 0.55, B: 0.15, Opacity: 0.55, PulseSpeed: 0.6},  // challenging: orange
			{R: 0.95, G: 0.15, B: 0.15, Opacity: 0.65, PulseSpeed: 0.9},  // dangerous: red
		}},
		"horror": {Tiers: [5]threatTierPreset{
			{R: 0.40, G: 0.38, B: 0.35, Opacity: 0.20, PulseSpeed: 0.0},
			{R: 0.45, G: 0.55, B: 0.35, Opacity: 0.30, PulseSpeed: 0.0},
			{R: 0.70, G: 0.55, B: 0.30, Opacity: 0.40, PulseSpeed: 0.0},
			{R: 0.80, G: 0.30, B: 0.20, Opacity: 0.55, PulseSpeed: 0.5},
			{R: 0.85, G: 0.10, B: 0.10, Opacity: 0.70, PulseSpeed: 0.8},
		}},
		"scifi": {Tiers: [5]threatTierPreset{
			{R: 0.50, G: 0.55, B: 0.60, Opacity: 0.25, PulseSpeed: 0.0},
			{R: 0.20, G: 0.75, B: 0.70, Opacity: 0.35, PulseSpeed: 0.0},
			{R: 0.25, G: 0.80, B: 0.90, Opacity: 0.45, PulseSpeed: 0.0},
			{R: 0.90, G: 0.60, B: 0.20, Opacity: 0.55, PulseSpeed: 0.7},
			{R: 0.95, G: 0.20, B: 0.20, Opacity: 0.65, PulseSpeed: 1.0},
		}},
		"cyberpunk": {Tiers: [5]threatTierPreset{
			{R: 0.45, G: 0.45, B: 0.50, Opacity: 0.25, PulseSpeed: 0.0},
			{R: 0.20, G: 0.90, B: 0.40, Opacity: 0.40, PulseSpeed: 0.0},
			{R: 0.90, G: 0.90, B: 0.20, Opacity: 0.50, PulseSpeed: 0.0},
			{R: 0.95, G: 0.40, B: 0.80, Opacity: 0.60, PulseSpeed: 0.8},
			{R: 0.95, G: 0.15, B: 0.50, Opacity: 0.70, PulseSpeed: 1.1},
		}},
		"postapoc": {Tiers: [5]threatTierPreset{
			{R: 0.45, G: 0.42, B: 0.38, Opacity: 0.20, PulseSpeed: 0.0},
			{R: 0.40, G: 0.65, B: 0.30, Opacity: 0.30, PulseSpeed: 0.0},
			{R: 0.75, G: 0.70, B: 0.30, Opacity: 0.40, PulseSpeed: 0.0},
			{R: 0.85, G: 0.50, B: 0.20, Opacity: 0.55, PulseSpeed: 0.6},
			{R: 0.90, G: 0.20, B: 0.15, Opacity: 0.65, PulseSpeed: 0.85},
		}},
	}
}

// clampThreat ensures a value stays in [0.0, 1.0].
func clampThreat(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
