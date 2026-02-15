// Package engine provides the SpellCastGlowSystem which applies genre-aware
// visual glow effects to entities actively casting spells. The glow color is
// derived from the spell's elemental affinity (fire=orange, ice=cyan,
// lightning=yellow, etc.), with genre-specific color shifts (horror darkens,
// cyberpunk saturates neon). Glow intensity ramps up with the CastingBar
// progress and includes a subtle pulse for visual feedback.
package engine

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/magic"
	"github.com/sirupsen/logrus"
)

// SpellCastGlowComponent stores the current visual glow state for an entity
// that is casting a spell. Values feed into the sprite rendering pipeline.
type SpellCastGlowComponent struct {
	// Glow color (RGB, 0.0-1.0) derived from element + genre
	GlowR float64
	GlowG float64
	GlowB float64

	// Current glow intensity (0.0-1.0), ramps with casting progress
	Intensity float64

	// Pulse phase in radians for subtle oscillation during cast
	PulsePhase float64

	// Pulse speed in cycles per second (0.8-1.5)
	PulseSpeed float64

	// Pulse amplitude as fraction of intensity (0.0-0.15)
	PulseAmplitude float64

	// Glow radius in pixels around entity center (2.0-6.0)
	GlowRadius float64

	// Whether the glow is currently active
	Active bool

	// Fade-out timer when casting finishes (seconds remaining)
	FadeTimer float64

	// Max fade duration for normalization
	FadeDuration float64
}

// Type returns the component type identifier.
func (c *SpellCastGlowComponent) Type() string {
	return "spell_cast_glow"
}

// NewSpellCastGlowComponent creates a component with inactive defaults.
func NewSpellCastGlowComponent() *SpellCastGlowComponent {
	return &SpellCastGlowComponent{
		PulseSpeed:   1.0,
		PulseAmplitude: 0.1,
		GlowRadius:   3.0,
		FadeDuration: 0.3,
	}
}

// elementGlowColor holds base RGB for a spell element.
type elementGlowColor struct {
	R, G, B float64
}

// genreGlowShift holds per-genre color modifications.
type genreGlowShift struct {
	Brightness float64 // Multiplier on intensity (0.5-1.2)
	Saturation float64 // Multiplier on color distance from grey (0.8-1.3)
	PulseSpeed float64 // Genre-specific pulse speed modifier
}

// SpellCastGlowSystem monitors entities with spell_slots and applies genre-aware
// glow effects while they are actively casting spells.
type SpellCastGlowSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	genreID string

	elementColors map[magic.ElementType]elementGlowColor
	genreShifts   map[string]genreGlowShift
}

// NewSpellCastGlowSystem creates a new spell cast glow system.
func NewSpellCastGlowSystem(world *World, seed int64) *SpellCastGlowSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "spell_cast_glow")
	}

	sys := &SpellCastGlowSystem{
		world:         world,
		logger:        logEntry,
		rng:           rand.New(rand.NewSource(seed)),
		genreID:       "fantasy",
		elementColors: buildElementGlowColors(),
		genreShifts:   buildGenreGlowShifts(),
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("spell cast glow system created")
	}
	return sys
}

// SetGenre configures the active genre for glow color shifts.
func (s *SpellCastGlowSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for spell cast glow")
	}
}

// Update processes all entities each frame: activates glow for casting entities,
// ramps intensity with CastingBar progress, animates pulse, and fades glow
// when casting completes.
func (s *SpellCastGlowSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		slotComp, ok := entity.GetComponent("spell_slots")
		if !ok {
			continue
		}
		slots, ok := slotComp.(*SpellSlotComponent)
		if !ok {
			continue
		}

		gc := s.getOrCreateComponent(entity)

		if slots.IsCasting() {
			s.updateCasting(gc, slots, deltaTime)
		} else if gc.Active {
			s.updateFadeOut(gc, deltaTime)
		}
	}
}

// updateCasting sets glow color from the spell element and ramps intensity.
func (s *SpellCastGlowSystem) updateCasting(gc *SpellCastGlowComponent, slots *SpellSlotComponent, deltaTime float64) {
	spell := slots.GetSlot(slots.Casting)

	if !gc.Active {
		// Starting a new cast: set color from element + genre
		gc.Active = true
		gc.FadeTimer = 0

		elem := magic.ElementArcane
		if spell != nil {
			elem = spell.Element
		}
		s.applyElementColor(gc, elem)
	}

	// Intensity ramps with casting bar (0.0 -> 1.0)
	baseIntensity := 0.2 + slots.CastingBar*0.6
	shift := s.getGenreShift()
	baseIntensity *= shift.Brightness

	// Pulse animation
	gc.PulsePhase += deltaTime * gc.PulseSpeed * shift.PulseSpeed * 2 * math.Pi
	if gc.PulsePhase > 2*math.Pi {
		gc.PulsePhase -= 2 * math.Pi
	}
	pulse := math.Sin(gc.PulsePhase) * gc.PulseAmplitude
	gc.Intensity = clampSpellGlow(baseIntensity + pulse)

	// Radius grows slightly during cast
	gc.GlowRadius = 2.0 + slots.CastingBar*4.0
}

// updateFadeOut decreases intensity after casting completes.
func (s *SpellCastGlowSystem) updateFadeOut(gc *SpellCastGlowComponent, deltaTime float64) {
	gc.FadeTimer += deltaTime
	if gc.FadeTimer >= gc.FadeDuration {
		gc.Active = false
		gc.Intensity = 0
		gc.FadeTimer = 0
		return
	}
	// Linear fade
	progress := gc.FadeTimer / gc.FadeDuration
	gc.Intensity *= (1.0 - progress)
	if gc.Intensity < 0.01 {
		gc.Active = false
		gc.Intensity = 0
		gc.FadeTimer = 0
	}
}

// applyElementColor sets the glow RGB from the spell element with genre shift.
func (s *SpellCastGlowSystem) applyElementColor(gc *SpellCastGlowComponent, elem magic.ElementType) {
	color, ok := s.elementColors[elem]
	if !ok {
		color = s.elementColors[magic.ElementArcane]
	}

	shift := s.getGenreShift()

	// Apply saturation shift: move color toward/away from grey
	grey := (color.R + color.G + color.B) / 3.0
	gc.GlowR = clampSpellGlow(grey + (color.R-grey)*shift.Saturation)
	gc.GlowG = clampSpellGlow(grey + (color.G-grey)*shift.Saturation)
	gc.GlowB = clampSpellGlow(grey + (color.B-grey)*shift.Saturation)

	// Per-entity variation for visual variety
	variation := s.rng.Float64()*0.06 - 0.03
	gc.GlowR = clampSpellGlow(gc.GlowR + variation)
	gc.GlowG = clampSpellGlow(gc.GlowG + variation)
	gc.GlowB = clampSpellGlow(gc.GlowB + variation)
}

// getGenreShift returns the genre color modifier, defaulting to fantasy.
func (s *SpellCastGlowSystem) getGenreShift() genreGlowShift {
	shift, ok := s.genreShifts[s.genreID]
	if !ok {
		shift = s.genreShifts["fantasy"]
	}
	return shift
}

// getOrCreateComponent retrieves or lazily creates the glow component.
func (s *SpellCastGlowSystem) getOrCreateComponent(entity *Entity) *SpellCastGlowComponent {
	comp, ok := entity.GetComponent("spell_cast_glow")
	if ok {
		if gc, ok := comp.(*SpellCastGlowComponent); ok {
			return gc
		}
	}
	gc := NewSpellCastGlowComponent()
	entity.AddComponent(gc)
	return gc
}

// buildElementGlowColors returns base glow colors per spell element.
func buildElementGlowColors() map[magic.ElementType]elementGlowColor {
	return map[magic.ElementType]elementGlowColor{
		magic.ElementNone:      {R: 0.70, G: 0.70, B: 0.70},
		magic.ElementFire:      {R: 1.00, G: 0.45, B: 0.10},
		magic.ElementIce:       {R: 0.20, G: 0.75, B: 0.95},
		magic.ElementLightning: {R: 0.95, G: 0.90, B: 0.30},
		magic.ElementEarth:     {R: 0.55, G: 0.40, B: 0.20},
		magic.ElementWind:      {R: 0.70, G: 0.90, B: 0.70},
		magic.ElementLight:     {R: 1.00, G: 0.95, B: 0.75},
		magic.ElementDark:      {R: 0.40, G: 0.15, B: 0.55},
		magic.ElementArcane:    {R: 0.65, G: 0.30, B: 0.90},
	}
}

// buildGenreGlowShifts returns genre-specific visual modifiers.
func buildGenreGlowShifts() map[string]genreGlowShift {
	return map[string]genreGlowShift{
		"fantasy":   {Brightness: 1.0, Saturation: 1.0, PulseSpeed: 1.0},
		"horror":    {Brightness: 0.7, Saturation: 0.85, PulseSpeed: 0.7},
		"scifi":     {Brightness: 1.1, Saturation: 1.15, PulseSpeed: 1.2},
		"cyberpunk": {Brightness: 1.15, Saturation: 1.3, PulseSpeed: 1.3},
		"postapoc":  {Brightness: 0.8, Saturation: 0.9, PulseSpeed: 0.9},
	}
}

// clampSpellGlow ensures a value stays in [0.0, 1.0].
func clampSpellGlow(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
