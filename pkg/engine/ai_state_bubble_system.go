// Package engine provides the AIStateBubbleSystem which renders genre-aware
// floating indicator symbols above NPC entities based on their current AI state.
// Idle shows a sleep symbol (Zzz), Detect shows alert (!), Chase shows urgency (!!),
// Attack shows combat (X), Flee shows panic (...), Patrol shows awareness (~).
// Genre palettes control symbol color: fantasy warm gold, horror blood red,
// sci-fi electric cyan, cyberpunk neon magenta, post-apocalyptic amber.
package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// AIStateBubbleComponent stores procedurally assigned bubble indicator
// parameters for an NPC entity. Values feed into the sprite rendering
// pipeline for per-entity state indicator overlays.
type AIStateBubbleComponent struct {
	// Symbol identifier for current state (e.g. "zzz", "!", "!!", "X", "...", "~")
	Symbol string

	// Symbol color (RGB 0.0-1.0) from genre palette
	SymbolR float64
	SymbolG float64
	SymbolB float64

	// Vertical offset above entity sprite in pixels (positive = up)
	OffsetY float64

	// Current bob phase for floating animation (radians)
	BobPhase float64

	// Bob amplitude in pixels (1.0-3.0)
	BobAmplitude float64

	// Bob speed in radians per second
	BobSpeed float64

	// Current opacity (0.0-1.0) for fade transitions
	Opacity float64

	// Target opacity for smooth transitions
	TargetOpacity float64

	// Scale factor for symbol size (0.5-1.5)
	Scale float64

	// The AI state this bubble represents
	DisplayState AIState

	// Previous state for detecting transitions
	PreviousState AIState

	// Whether the bubble is active
	Enabled bool
}

// Type returns the component type identifier.
func (c *AIStateBubbleComponent) Type() string {
	return "ai_state_bubble"
}

// NewAIStateBubbleComponent creates a component with disabled defaults.
func NewAIStateBubbleComponent() *AIStateBubbleComponent {
	return &AIStateBubbleComponent{
		Symbol:        "",
		SymbolR:       0.0,
		SymbolG:       0.0,
		SymbolB:       0.0,
		OffsetY:       -20.0,
		BobPhase:      0.0,
		BobAmplitude:  2.0,
		BobSpeed:      2.5,
		Opacity:       0.0,
		TargetOpacity: 0.0,
		Scale:         1.0,
		DisplayState:  AIStateIdle,
		PreviousState: AIStateIdle,
		Enabled:       false,
	}
}

// genreBubblePalette holds genre-specific symbol colors.
type genreBubblePalette struct {
	R, G, B float64
}

// aiStateBubbleSymbol maps AI state to display symbol and visual weight.
type aiStateBubbleSymbol struct {
	Symbol    string
	Opacity   float64 // Target opacity for this state
	Scale     float64 // Scale multiplier
	BobSpeed  float64 // Bob speed override
	BobAmp    float64 // Bob amplitude override
}

// AIStateBubbleSystem assigns genre-aware floating indicators to NPC
// entities based on their AI behavior state. Bob animation runs every
// frame; state detection is throttled.
type AIStateBubbleSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	genreID string

	// Throttle for state checks
	stateCheckInterval float64
	timeSinceCheck     float64

	// Fade speed for opacity transitions (units per second)
	fadeSpeed float64

	palettes map[string]genreBubblePalette
	symbols  map[AIState]aiStateBubbleSymbol
}

// NewAIStateBubbleSystem creates a new AI state bubble system.
func NewAIStateBubbleSystem(world *World, seed int64) *AIStateBubbleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "ai_state_bubble")
	}

	sys := &AIStateBubbleSystem{
		world:              world,
		logger:             logEntry,
		rng:                rand.New(rand.NewSource(seed)),
		genreID:            "fantasy",
		stateCheckInterval: 0.2,
		fadeSpeed:          4.0,
		palettes:           buildBubblePalettes(),
		symbols:            buildBubbleSymbols(),
	}

	if logEntry != nil {
		logEntry.WithFields(logrus.Fields{
			"system_name": "ai_state_bubble",
			"seed":        seed,
		}).Debug("AI state bubble system created")
	}
	return sys
}

// SetGenre sets the active genre for palette selection.
func (s *AIStateBubbleSystem) SetGenre(genreID string) {
	if _, ok := s.palettes[genreID]; ok {
		s.genreID = genreID
	}
}

// Update processes all entities, updating bob animation every frame and
// checking AI state transitions on a throttled interval.
func (s *AIStateBubbleSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime
	checkState := s.timeSinceCheck >= s.stateCheckInterval

	if checkState {
		s.timeSinceCheck = 0
	}

	palette := s.palettes[s.genreID]

	for _, entity := range entities {
		if entity == nil {
			continue
		}
		if !entity.HasComponent("position") || !entity.HasComponent("sprite") {
			continue
		}
		// Skip player entities
		if entity.HasComponent("input") {
			continue
		}
		if !entity.HasComponent("ai") {
			continue
		}

		bc := s.getOrCreateComponent(entity)

		if checkState {
			s.updateStateIndicator(entity, bc, palette)
		}

		// Animate bob and fade every frame
		s.animateBubble(bc, deltaTime)
	}
}

// updateStateIndicator reads AI state and updates bubble symbol + colors.
func (s *AIStateBubbleSystem) updateStateIndicator(entity *Entity, bc *AIStateBubbleComponent, palette genreBubblePalette) {
	aiComp, ok := entity.GetComponent("ai")
	if !ok {
		return
	}
	ai, ok := aiComp.(*AIComponent)
	if !ok {
		return
	}

	sym, known := s.symbols[ai.State]
	if !known {
		bc.TargetOpacity = 0
		return
	}

	// Detect state transition
	if ai.State != bc.DisplayState {
		bc.PreviousState = bc.DisplayState
		bc.DisplayState = ai.State

		// Per-entity color variation for visual diversity
		variation := s.rng.Float64()*0.08 - 0.04
		bc.SymbolR = clampBubble(palette.R + variation)
		bc.SymbolG = clampBubble(palette.G + variation)
		bc.SymbolB = clampBubble(palette.B + variation)

		// Flash scale on transition
		bc.Scale = sym.Scale * 1.3
	}

	bc.Symbol = sym.Symbol
	bc.TargetOpacity = sym.Opacity
	bc.BobSpeed = sym.BobSpeed
	bc.BobAmplitude = sym.BobAmp
	bc.Enabled = true
}

// animateBubble updates bob motion and fade transitions per frame.
func (s *AIStateBubbleSystem) animateBubble(bc *AIStateBubbleComponent, dt float64) {
	// Bob animation
	bc.BobPhase += bc.BobSpeed * dt
	if bc.BobPhase > 2*math.Pi {
		bc.BobPhase -= 2 * math.Pi
	}
	bc.OffsetY = -20.0 + bc.BobAmplitude*math.Sin(bc.BobPhase)

	// Smooth opacity fade
	if bc.Opacity < bc.TargetOpacity {
		bc.Opacity += s.fadeSpeed * dt
		if bc.Opacity > bc.TargetOpacity {
			bc.Opacity = bc.TargetOpacity
		}
	} else if bc.Opacity > bc.TargetOpacity {
		bc.Opacity -= s.fadeSpeed * dt
		if bc.Opacity < bc.TargetOpacity {
			bc.Opacity = bc.TargetOpacity
		}
	}

	// Decay scale back toward symbol default
	sym, ok := s.symbols[bc.DisplayState]
	if ok && bc.Scale > sym.Scale {
		bc.Scale -= 0.8 * dt
		if bc.Scale < sym.Scale {
			bc.Scale = sym.Scale
		}
	}

	// Disable when fully faded out
	if bc.Opacity <= 0 && bc.TargetOpacity <= 0 {
		bc.Enabled = false
	}
}

// getOrCreateComponent retrieves or lazily creates the bubble component.
func (s *AIStateBubbleSystem) getOrCreateComponent(entity *Entity) *AIStateBubbleComponent {
	comp, ok := entity.GetComponent("ai_state_bubble")
	if ok {
		if bc, ok := comp.(*AIStateBubbleComponent); ok {
			return bc
		}
	}
	bc := NewAIStateBubbleComponent()
	// Randomize initial phase to avoid synchronized bobbing
	bc.BobPhase = s.rng.Float64() * 2 * math.Pi
	entity.AddComponent(bc)
	return bc
}

// buildBubblePalettes returns genre-specific symbol colors.
func buildBubblePalettes() map[string]genreBubblePalette {
	return map[string]genreBubblePalette{
		"fantasy":   {R: 0.95, G: 0.85, B: 0.30}, // Warm gold
		"horror":    {R: 0.90, G: 0.15, B: 0.10}, // Blood red
		"scifi":     {R: 0.20, G: 0.85, B: 0.95}, // Electric cyan
		"cyberpunk": {R: 0.90, G: 0.20, B: 0.85}, // Neon magenta
		"postapoc":  {R: 0.90, G: 0.75, B: 0.20}, // Toxic amber
	}
}

// buildBubbleSymbols returns state-to-symbol mappings with visual parameters.
func buildBubbleSymbols() map[AIState]aiStateBubbleSymbol {
	return map[AIState]aiStateBubbleSymbol{
		AIStateIdle:    {Symbol: "zzz", Opacity: 0.45, Scale: 0.7, BobSpeed: 1.5, BobAmp: 1.5},
		AIStatePatrol:  {Symbol: "~", Opacity: 0.50, Scale: 0.8, BobSpeed: 2.0, BobAmp: 1.8},
		AIStateDetect:  {Symbol: "!", Opacity: 0.80, Scale: 1.1, BobSpeed: 3.5, BobAmp: 2.5},
		AIStateChase:   {Symbol: "!!", Opacity: 0.90, Scale: 1.2, BobSpeed: 4.0, BobAmp: 3.0},
		AIStateAttack:  {Symbol: "X", Opacity: 0.95, Scale: 1.3, BobSpeed: 5.0, BobAmp: 2.0},
		AIStateFlee:    {Symbol: "...", Opacity: 0.70, Scale: 0.9, BobSpeed: 4.5, BobAmp: 2.8},
		AIStateReturn:  {Symbol: "<", Opacity: 0.50, Scale: 0.8, BobSpeed: 2.0, BobAmp: 1.8},
	}
}

// clampBubble ensures a value stays in [0.0, 1.0].
func clampBubble(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
