// Package engine provides the CreatureEyeGlowSystem which assigns genre-aware
// glowing eye parameters to hostile creatures and boss entities. Genre palettes
// drive glow color: fantasy uses golden amber, horror uses blood red, sci-fi
// uses electric cyan, cyberpunk uses neon magenta, and post-apocalyptic uses
// sickly green. Threat level (based on max health and faction) controls glow
// intensity and pulse speed, making stronger enemies more visually menacing.
package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// CreatureEyeGlowComponent stores procedurally assigned eye glow parameters
// for hostile creature entities. These values feed into the sprite rendering
// pipeline for per-entity glowing eye effects.
type CreatureEyeGlowComponent struct {
	// Glow color (RGB, 0.0-1.0) derived from genre palette
	GlowR float64
	GlowG float64
	GlowB float64

	// Base intensity before pulse modulation (0.0-1.0)
	BaseIntensity float64

	// Current intensity after pulse modulation (0.0-1.0)
	CurrentIntensity float64

	// Pulse speed in cycles per second (0.0 = no pulse)
	PulseSpeed float64

	// Pulse amplitude as fraction of BaseIntensity (0.0-0.5)
	PulseAmplitude float64

	// Accumulated pulse phase in radians
	PulsePhase float64

	// Glow radius in pixels around each eye (1.0-4.0)
	GlowRadius float64

	// Whether this entity should have eye glow at all
	Enabled bool
}

// Type returns the component type identifier.
func (c *CreatureEyeGlowComponent) Type() string {
	return "creature_eye_glow"
}

// NewCreatureEyeGlowComponent creates a component with disabled defaults.
func NewCreatureEyeGlowComponent() *CreatureEyeGlowComponent {
	return &CreatureEyeGlowComponent{
		GlowR:            0.0,
		GlowG:            0.0,
		GlowB:            0.0,
		BaseIntensity:    0.0,
		CurrentIntensity: 0.0,
		PulseSpeed:       0.0,
		PulseAmplitude:   0.0,
		PulsePhase:       0.0,
		GlowRadius:       1.0,
		Enabled:          false,
	}
}

// genreEyeGlowPalette holds genre-specific eye glow colors and base parameters.
type genreEyeGlowPalette struct {
	R, G, B        float64 // Primary glow color
	BaseIntensity  float64 // Genre default intensity
	BasePulseSpeed float64 // Genre default pulse speed
}

// CreatureEyeGlowSystem assigns genre-aware glowing eye effects to hostile
// creature entities. Updates pulse animation each frame for active glow
// components, and recomputes glow parameters when genre changes.
type CreatureEyeGlowSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	genreID     string
	lastGenreID string
	initialized bool

	// Throttle for genre-change checks (not for pulse updates)
	genreCheckInterval float64
	timeSinceCheck     float64

	palettes map[string]genreEyeGlowPalette
}

// NewCreatureEyeGlowSystem creates a new creature eye glow system.
func NewCreatureEyeGlowSystem(world *World, seed int64) *CreatureEyeGlowSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "creature_eye_glow")
	}

	sys := &CreatureEyeGlowSystem{
		world:              world,
		logger:             logEntry,
		rng:                rand.New(rand.NewSource(seed)),
		genreID:            "fantasy",
		lastGenreID:        "",
		genreCheckInterval: 1.0,
		palettes:           buildEyeGlowPalettes(),
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("creature eye glow system created")
	}
	return sys
}

// SetGenre configures the active genre for eye glow computation.
func (s *CreatureEyeGlowSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for creature eye glow")
	}
}

// Update animates pulse on existing glow components every frame and checks
// for genre changes on a throttled interval to assign new glow parameters.
func (s *CreatureEyeGlowSystem) Update(entities []*Entity, deltaTime float64) {
	// Pulse animation runs every frame for smooth visuals
	s.updatePulse(entities, deltaTime)

	// Genre-change detection is throttled
	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck < s.genreCheckInterval {
		return
	}
	s.timeSinceCheck = 0

	genreChanged := s.genreID != s.lastGenreID || !s.initialized
	if !genreChanged {
		return
	}

	s.lastGenreID = s.genreID
	s.initialized = true

	s.assignGlowParameters(entities)
}

// updatePulse modulates CurrentIntensity using a sine wave for enabled glow components.
func (s *CreatureEyeGlowSystem) updatePulse(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		comp, ok := entity.GetComponent("creature_eye_glow")
		if !ok {
			continue
		}
		gc, ok := comp.(*CreatureEyeGlowComponent)
		if !ok || !gc.Enabled || gc.PulseSpeed <= 0 {
			continue
		}

		gc.PulsePhase += deltaTime * gc.PulseSpeed * 2 * math.Pi
		// Wrap phase to avoid floating-point growth
		if gc.PulsePhase > 2*math.Pi {
			gc.PulsePhase -= 2 * math.Pi
		}

		pulse := math.Sin(gc.PulsePhase) * gc.PulseAmplitude
		gc.CurrentIntensity = clampEyeGlow(gc.BaseIntensity + pulse)
	}
}

// assignGlowParameters sets genre-based glow on hostile creature entities.
func (s *CreatureEyeGlowSystem) assignGlowParameters(entities []*Entity) {
	palette, ok := s.palettes[s.genreID]
	if !ok {
		palette = s.palettes["fantasy"]
	}

	count := 0
	for _, entity := range entities {
		if !entity.HasComponent("ai") || !entity.HasComponent("sprite") {
			continue
		}
		// Players don't get creature eye glow
		if entity.HasComponent("input") {
			continue
		}

		gc := s.getOrCreateComponent(entity)
		threatLevel := s.computeThreatLevel(entity)

		// Only hostile or high-threat creatures get eye glow
		if threatLevel < 0.1 {
			gc.Enabled = false
			continue
		}

		gc.Enabled = true

		// Color from genre palette with per-entity variation
		variation := s.rng.Float64()*0.08 - 0.04
		gc.GlowR = clampEyeGlow(palette.R + variation)
		gc.GlowG = clampEyeGlow(palette.G + variation)
		gc.GlowB = clampEyeGlow(palette.B + variation)

		// Intensity scales with threat: 0.2 (weak) to 0.9 (boss-tier)
		gc.BaseIntensity = clampEyeGlow(palette.BaseIntensity*0.5 + threatLevel*0.5)
		gc.CurrentIntensity = gc.BaseIntensity

		// Pulse speed: stronger creatures pulse faster (0.3-1.2 Hz)
		gc.PulseSpeed = clampEyeGlow(palette.BasePulseSpeed + threatLevel*0.6)
		gc.PulseAmplitude = 0.05 + threatLevel*0.15

		// Glow radius: 1.0 (weak) to 3.5 (boss)
		gc.GlowRadius = 1.0 + threatLevel*2.5

		// Randomize initial phase for visual variety
		gc.PulsePhase = s.rng.Float64() * 2 * math.Pi

		count++
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"genre":           s.genreID,
			"entities_glowed": count,
		}).Debug("creature eye glow parameters assigned")
	}
}

// computeThreatLevel derives a 0.0-1.0 threat value from entity health and faction.
func (s *CreatureEyeGlowSystem) computeThreatLevel(entity *Entity) float64 {
	threat := 0.0

	// Boss faction entities are always high threat
	if comp, ok := entity.GetComponent("faction"); ok {
		if fc, ok := comp.(*FactionComponent); ok {
			switch fc.FactionID {
			case "boss_faction":
				threat = 0.9
			case "neutral_faction":
				return 0.0 // Neutral creatures don't glow
			}
			if fc.Reputation < -50 {
				threat = math.Max(threat, 0.4) // Hostile factions glow moderately
			}
		}
	}

	// Max health as a proxy for power (scale 50-500 HP range to 0.0-1.0)
	if comp, ok := entity.GetComponent("health"); ok {
		if hc, ok := comp.(*HealthComponent); ok {
			healthThreat := (hc.Max - 50.0) / 450.0
			if healthThreat < 0 {
				healthThreat = 0
			}
			if healthThreat > 1 {
				healthThreat = 1
			}
			threat = math.Max(threat, healthThreat*0.7)
		}
	}

	// Entities with high detection range are more dangerous
	if comp, ok := entity.GetComponent("ai"); ok {
		if ai, ok := comp.(*AIComponent); ok {
			if ai.DetectionRange > 200 {
				rangeThreat := (ai.DetectionRange - 200) / 300.0
				if rangeThreat > 0.3 {
					rangeThreat = 0.3
				}
				threat = math.Min(1.0, threat+rangeThreat)
			}
		}
	}

	return threat
}

// getOrCreateComponent retrieves or lazily creates the eye glow component.
func (s *CreatureEyeGlowSystem) getOrCreateComponent(entity *Entity) *CreatureEyeGlowComponent {
	comp, ok := entity.GetComponent("creature_eye_glow")
	if ok {
		if gc, ok := comp.(*CreatureEyeGlowComponent); ok {
			return gc
		}
	}
	gc := NewCreatureEyeGlowComponent()
	entity.AddComponent(gc)
	return gc
}

// buildEyeGlowPalettes returns genre-specific eye glow color presets.
func buildEyeGlowPalettes() map[string]genreEyeGlowPalette {
	return map[string]genreEyeGlowPalette{
		"fantasy":   {R: 0.95, G: 0.75, B: 0.20, BaseIntensity: 0.5, BasePulseSpeed: 0.4},
		"horror":    {R: 0.95, G: 0.15, B: 0.10, BaseIntensity: 0.7, BasePulseSpeed: 0.3},
		"scifi":     {R: 0.20, G: 0.85, B: 0.95, BaseIntensity: 0.6, BasePulseSpeed: 0.5},
		"cyberpunk": {R: 0.90, G: 0.20, B: 0.85, BaseIntensity: 0.65, BasePulseSpeed: 0.6},
		"postapoc":  {R: 0.40, G: 0.85, B: 0.20, BaseIntensity: 0.55, BasePulseSpeed: 0.35},
	}
}

// clampEyeGlow ensures a value stays in [0.0, 1.0].
func clampEyeGlow(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
