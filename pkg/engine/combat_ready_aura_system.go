// Package engine provides the CombatReadyAuraSystem which assigns genre-aware
// visual aura parameters to AI entities entering hostile states (Detect, Chase,
// Attack). The aura color is genre-driven: fantasy uses warm orange, horror uses
// deep crimson, sci-fi uses electric blue, cyberpunk uses neon magenta, and
// post-apocalyptic uses toxic amber. AI state controls intensity and pulse speed,
// providing clear visual feedback about enemy aggression level.
package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// CombatReadyAuraComponent stores procedurally assigned aura parameters for
// AI entities in hostile states. These values feed into the sprite rendering
// pipeline for per-entity combat readiness indicator effects.
type CombatReadyAuraComponent struct {
	// Aura color (RGB, 0.0-1.0) derived from genre palette
	AuraR float64
	AuraG float64
	AuraB float64

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

	// Aura radius in pixels around the entity (2.0-8.0)
	AuraRadius float64

	// Opacity of the aura ring (0.0-1.0)
	Opacity float64

	// Target opacity for smooth fade transitions
	TargetOpacity float64

	// The AI state that triggered this aura
	TriggerState AIState

	// Whether this entity should have a combat aura
	Enabled bool
}

// Type returns the component type identifier.
func (c *CombatReadyAuraComponent) Type() string {
	return "combat_ready_aura"
}

// NewCombatReadyAuraComponent creates a component with disabled defaults.
func NewCombatReadyAuraComponent() *CombatReadyAuraComponent {
	return &CombatReadyAuraComponent{
		AuraR:            0.0,
		AuraG:            0.0,
		AuraB:            0.0,
		BaseIntensity:    0.0,
		CurrentIntensity: 0.0,
		PulseSpeed:       0.0,
		PulseAmplitude:   0.0,
		PulsePhase:       0.0,
		AuraRadius:       2.0,
		Opacity:          0.0,
		TargetOpacity:    0.0,
		TriggerState:     AIStateIdle,
		Enabled:          false,
	}
}

// genreCombatAuraPalette holds genre-specific aura colors and parameters.
type genreCombatAuraPalette struct {
	R, G, B        float64 // Primary aura color
	BaseIntensity  float64 // Genre default intensity
	BasePulseSpeed float64 // Genre default pulse speed
}

// CombatReadyAuraSystem assigns genre-aware aura effects to AI entities that
// enter hostile states. Pulse animation runs every frame; state detection is
// throttled to reduce overhead.
type CombatReadyAuraSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	genreID string

	// Throttle for state checks (pulse runs every frame)
	stateCheckInterval float64
	timeSinceCheck     float64

	// Fade speed for opacity transitions (units per second)
	fadeSpeed float64

	palettes map[string]genreCombatAuraPalette
}

// NewCombatReadyAuraSystem creates a new combat ready aura system.
func NewCombatReadyAuraSystem(world *World, seed int64) *CombatReadyAuraSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "combat_ready_aura")
	}

	sys := &CombatReadyAuraSystem{
		world:              world,
		logger:             logEntry,
		rng:                rand.New(rand.NewSource(seed)),
		genreID:            "fantasy",
		stateCheckInterval: 0.2, // Check AI states 5 times per second
		fadeSpeed:          3.0, // Fade in/out over ~0.33s
		palettes:           buildCombatAuraPalettes(),
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("combat ready aura system created")
	}
	return sys
}

// SetGenre configures the active genre for aura color computation.
func (s *CombatReadyAuraSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for combat ready aura")
	}
}

// Update animates pulse on existing aura components every frame and checks
// AI states on a throttled interval to assign or remove aura parameters.
func (s *CombatReadyAuraSystem) Update(entities []*Entity, deltaTime float64) {
	// Pulse and fade animation runs every frame
	s.updatePulseAndFade(entities, deltaTime)

	// AI state detection is throttled
	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck < s.stateCheckInterval {
		return
	}
	s.timeSinceCheck = 0

	s.updateAuraStates(entities)
}

// updatePulseAndFade modulates CurrentIntensity via sine wave and smoothly
// transitions Opacity toward TargetOpacity for enabled aura components.
func (s *CombatReadyAuraSystem) updatePulseAndFade(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		comp, ok := entity.GetComponent("combat_ready_aura")
		if !ok {
			continue
		}
		ac, ok := comp.(*CombatReadyAuraComponent)
		if !ok {
			continue
		}

		// Smooth opacity fade toward target
		if ac.Opacity < ac.TargetOpacity {
			ac.Opacity += s.fadeSpeed * deltaTime
			if ac.Opacity > ac.TargetOpacity {
				ac.Opacity = ac.TargetOpacity
			}
		} else if ac.Opacity > ac.TargetOpacity {
			ac.Opacity -= s.fadeSpeed * deltaTime
			if ac.Opacity < ac.TargetOpacity {
				ac.Opacity = ac.TargetOpacity
			}
		}

		// Disable when fully faded out
		if ac.TargetOpacity <= 0 && ac.Opacity <= 0.01 {
			ac.Enabled = false
			ac.Opacity = 0
			continue
		}

		if !ac.Enabled || ac.PulseSpeed <= 0 {
			continue
		}

		// Advance pulse phase
		ac.PulsePhase += deltaTime * ac.PulseSpeed * 2 * math.Pi
		if ac.PulsePhase > 2*math.Pi {
			ac.PulsePhase -= 2 * math.Pi
		}

		pulse := math.Sin(ac.PulsePhase) * ac.PulseAmplitude
		ac.CurrentIntensity = clampCombatAura(ac.BaseIntensity + pulse)
	}
}

// updateAuraStates checks AI state for each entity and assigns/removes aura.
func (s *CombatReadyAuraSystem) updateAuraStates(entities []*Entity) {
	palette, ok := s.palettes[s.genreID]
	if !ok {
		palette = s.palettes["fantasy"]
	}

	for _, entity := range entities {
		if !entity.HasComponent("ai") || !entity.HasComponent("sprite") {
			continue
		}
		// Players don't get combat ready auras
		if entity.HasComponent("input") {
			continue
		}

		aiComp, ok := entity.GetComponent("ai")
		if !ok {
			continue
		}
		ai, ok := aiComp.(*AIComponent)
		if !ok {
			continue
		}

		isHostile := ai.State == AIStateDetect || ai.State == AIStateChase || ai.State == AIStateAttack
		ac := s.getOrCreateComponent(entity)

		if isHostile {
			s.applyHostileAura(ac, ai.State, palette)
		} else {
			// Fade out when returning to passive states
			ac.TargetOpacity = 0
		}
	}
}

// applyHostileAura sets aura parameters based on AI state and genre palette.
func (s *CombatReadyAuraSystem) applyHostileAura(ac *CombatReadyAuraComponent, state AIState, palette genreCombatAuraPalette) {
	ac.Enabled = true
	ac.TriggerState = state

	// Per-entity color variation
	variation := s.rng.Float64()*0.06 - 0.03
	ac.AuraR = clampCombatAura(palette.R + variation)
	ac.AuraG = clampCombatAura(palette.G + variation)
	ac.AuraB = clampCombatAura(palette.B + variation)

	// State-based intensity, pulse, and radius
	switch state {
	case AIStateDetect:
		ac.BaseIntensity = clampCombatAura(palette.BaseIntensity * 0.5)
		ac.PulseSpeed = clampCombatAura(palette.BasePulseSpeed * 0.6)
		ac.PulseAmplitude = 0.08
		ac.AuraRadius = 3.0
		ac.TargetOpacity = 0.4
	case AIStateChase:
		ac.BaseIntensity = clampCombatAura(palette.BaseIntensity * 0.75)
		ac.PulseSpeed = clampCombatAura(palette.BasePulseSpeed * 1.0)
		ac.PulseAmplitude = 0.12
		ac.AuraRadius = 5.0
		ac.TargetOpacity = 0.65
	case AIStateAttack:
		ac.BaseIntensity = clampCombatAura(palette.BaseIntensity * 1.0)
		ac.PulseSpeed = clampCombatAura(palette.BasePulseSpeed * 1.4)
		ac.PulseAmplitude = 0.18
		ac.AuraRadius = 7.0
		ac.TargetOpacity = 0.85
	default:
		ac.TargetOpacity = 0
	}

	ac.CurrentIntensity = ac.BaseIntensity
}

// getOrCreateComponent retrieves or lazily creates the combat aura component.
func (s *CombatReadyAuraSystem) getOrCreateComponent(entity *Entity) *CombatReadyAuraComponent {
	comp, ok := entity.GetComponent("combat_ready_aura")
	if ok {
		if ac, ok := comp.(*CombatReadyAuraComponent); ok {
			return ac
		}
	}
	ac := NewCombatReadyAuraComponent()
	entity.AddComponent(ac)
	return ac
}

// buildCombatAuraPalettes returns genre-specific aura color presets.
func buildCombatAuraPalettes() map[string]genreCombatAuraPalette {
	return map[string]genreCombatAuraPalette{
		"fantasy":   {R: 0.95, G: 0.60, B: 0.15, BaseIntensity: 0.55, BasePulseSpeed: 0.5},
		"horror":    {R: 0.90, G: 0.10, B: 0.10, BaseIntensity: 0.70, BasePulseSpeed: 0.35},
		"scifi":     {R: 0.15, G: 0.70, B: 0.95, BaseIntensity: 0.60, BasePulseSpeed: 0.55},
		"cyberpunk": {R: 0.85, G: 0.15, B: 0.90, BaseIntensity: 0.65, BasePulseSpeed: 0.65},
		"postapoc":  {R: 0.90, G: 0.70, B: 0.15, BaseIntensity: 0.55, BasePulseSpeed: 0.40},
	}
}

// clampCombatAura ensures a value stays in [0.0, 1.0].
func clampCombatAura(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
