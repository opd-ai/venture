// Package engine provides the ShieldBubbleOverlaySystem which renders
// genre-aware translucent shield bubble visuals around entities with an
// active ShieldComponent. The bubble opacity and radius scale with remaining
// shield amount. Genre presets control color, pulse speed, and flicker
// behavior: fantasy uses warm golden glow, sci-fi uses cool energy fields,
// horror uses blood-red wards, cyberpunk uses neon holographic grids,
// post-apocalyptic uses dusty amber barriers.
package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// ShieldBubbleOverlayComponent stores per-entity shield bubble visual state.
// Pure data — all logic lives in ShieldBubbleOverlaySystem.
type ShieldBubbleOverlayComponent struct {
	// Bubble color (RGB 0.0-1.0)
	BubbleR float64
	BubbleG float64
	BubbleB float64

	// Current opacity (0.0-1.0), scaled by shield remaining fraction
	Opacity float64

	// Bubble radius in pixels around entity center
	Radius float64

	// Base radius before shield-amount scaling
	BaseRadius float64

	// Pulse phase accumulator (radians)
	PulsePhase float64

	// Pulse speed (radians/sec)
	PulseSpeed float64

	// Flicker intensity (0.0-1.0), increases as shield depletes
	FlickerIntensity float64

	// Whether the bubble should be drawn
	Visible bool

	// Fade-out timer (seconds remaining); >0 means fading out
	FadeTimer float64

	// Maximum fade duration for normalization
	FadeMax float64
}

// Type returns the component type identifier.
func (c *ShieldBubbleOverlayComponent) Type() string { return "shield_bubble_overlay" }

// shieldBubblePreset holds genre-specific bubble visual parameters.
type shieldBubblePreset struct {
	R, G, B    float64
	BaseAlpha  float64 // Maximum opacity when shield is full
	PulseSpeed float64 // Radians per second
	BaseRadius float64 // Default bubble radius in pixels
	FadeDur    float64 // Fade-out duration in seconds
}

// ShieldBubbleOverlaySystem reads ShieldComponent on entities and writes
// bubble visual parameters to ShieldBubbleOverlayComponent for the render
// pipeline. It handles activation, pulsing animation, low-shield flicker,
// and smooth fade-out when the shield expires.
type ShieldBubbleOverlaySystem struct {
	world  *World
	seed   int64
	rng    *rand.Rand
	logger *logrus.Entry

	genreID string
	preset  shieldBubblePreset

	// Throttle shield state reads to reduce per-frame overhead
	checkInterval  float64
	timeSinceCheck float64
}

// NewShieldBubbleOverlaySystem creates a new shield bubble overlay system.
func NewShieldBubbleOverlaySystem(world *World, seed int64) *ShieldBubbleOverlaySystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "shield_bubble_overlay",
	})

	sys := &ShieldBubbleOverlaySystem{
		world:          world,
		seed:           seed,
		rng:            rand.New(rand.NewSource(seed)),
		logger:         logger,
		genreID:        "fantasy",
		checkInterval:  0.15,
		timeSinceCheck: 0,
	}
	sys.preset = sys.getGenrePreset("fantasy")
	logger.Debug("shield bubble overlay system created")
	return sys
}

// SetGenre configures genre-specific bubble visual parameters.
func (s *ShieldBubbleOverlaySystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getGenrePreset(genreID)
	s.logger.WithField("genre", genreID).Debug("genre preset applied")
}

// getGenrePreset returns bubble configuration for the given genre.
func (s *ShieldBubbleOverlaySystem) getGenrePreset(genreID string) shieldBubblePreset {
	switch genreID {
	case "horror":
		return shieldBubblePreset{
			R: 0.8, G: 0.15, B: 0.1,
			BaseAlpha: 0.35, PulseSpeed: 2.0, BaseRadius: 18.0, FadeDur: 0.4,
		}
	case "scifi":
		return shieldBubblePreset{
			R: 0.2, G: 0.7, B: 0.9,
			BaseAlpha: 0.45, PulseSpeed: 3.5, BaseRadius: 20.0, FadeDur: 0.6,
		}
	case "cyberpunk":
		return shieldBubblePreset{
			R: 0.1, G: 0.95, B: 0.8,
			BaseAlpha: 0.5, PulseSpeed: 4.0, BaseRadius: 19.0, FadeDur: 0.5,
		}
	case "postapoc":
		return shieldBubblePreset{
			R: 0.75, G: 0.6, B: 0.2,
			BaseAlpha: 0.3, PulseSpeed: 1.5, BaseRadius: 17.0, FadeDur: 0.35,
		}
	default: // fantasy
		return shieldBubblePreset{
			R: 0.9, G: 0.75, B: 0.3,
			BaseAlpha: 0.4, PulseSpeed: 2.5, BaseRadius: 18.0, FadeDur: 0.5,
		}
	}
}

// Update processes all entities, updating shield bubble state on a throttled
// interval and animating pulse/flicker every frame.
func (s *ShieldBubbleOverlaySystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime
	checkShield := s.timeSinceCheck >= s.checkInterval
	if checkShield {
		s.timeSinceCheck = 0
	}

	for _, entity := range entities {
		if entity == nil {
			continue
		}
		if !entity.HasComponent("position") || !entity.HasComponent("sprite") {
			continue
		}

		comp, hasOverlay := entity.GetComponent("shield_bubble_overlay")
		shieldComp, hasShield := entity.GetComponent("shield")

		// No overlay and no shield — skip entirely
		if !hasOverlay && !hasShield {
			continue
		}

		var shield *ShieldComponent
		if hasShield {
			shield, _ = shieldComp.(*ShieldComponent)
		}

		shieldActive := shield != nil && shield.IsActive()

		if shieldActive {
			bc := s.getOrCreateOverlay(entity)
			if checkShield {
				s.updateBubbleFromShield(bc, shield)
			}
			s.animateBubble(bc, deltaTime)
		} else if hasOverlay {
			bc, ok := comp.(*ShieldBubbleOverlayComponent)
			if ok {
				s.handleFadeOut(bc, deltaTime)
			}
		}
	}
}

// updateBubbleFromShield sets bubble parameters based on current shield state.
func (s *ShieldBubbleOverlaySystem) updateBubbleFromShield(bc *ShieldBubbleOverlayComponent, shield *ShieldComponent) {
	p := s.preset
	fraction := shield.Amount / math.Max(shield.MaxAmount, 1.0)

	bc.BubbleR = p.R
	bc.BubbleG = p.G
	bc.BubbleB = p.B
	bc.Opacity = p.BaseAlpha * fraction
	bc.Radius = p.BaseRadius * (0.8 + 0.2*fraction)
	bc.BaseRadius = p.BaseRadius
	bc.PulseSpeed = p.PulseSpeed
	bc.Visible = true
	bc.FadeTimer = 0
	bc.FadeMax = p.FadeDur

	// Flicker increases as shield depletes below 30%
	if fraction < 0.3 {
		bc.FlickerIntensity = (0.3 - fraction) / 0.3
	} else {
		bc.FlickerIntensity = 0
	}
}

// animateBubble updates pulse phase and applies flicker modulation.
func (s *ShieldBubbleOverlaySystem) animateBubble(bc *ShieldBubbleOverlayComponent, deltaTime float64) {
	bc.PulsePhase += bc.PulseSpeed * deltaTime
	if bc.PulsePhase > 2*math.Pi {
		bc.PulsePhase -= 2 * math.Pi
	}

	// Pulse modulates opacity by ±10%
	pulseMod := 0.1 * math.Sin(bc.PulsePhase)
	bc.Opacity = clampShieldBubble(bc.Opacity + pulseMod)

	// Flicker creates random opacity dips
	if bc.FlickerIntensity > 0 {
		flicker := bc.FlickerIntensity * (s.rng.Float64()*0.3 - 0.15)
		bc.Opacity = clampShieldBubble(bc.Opacity + flicker)
	}
}

// handleFadeOut progressively reduces opacity until the bubble disappears.
func (s *ShieldBubbleOverlaySystem) handleFadeOut(bc *ShieldBubbleOverlayComponent, deltaTime float64) {
	if !bc.Visible {
		return
	}

	if bc.FadeMax <= 0 {
		bc.FadeMax = s.preset.FadeDur
	}
	if bc.FadeTimer == 0 {
		bc.FadeTimer = bc.FadeMax
	}

	bc.FadeTimer -= deltaTime
	if bc.FadeTimer <= 0 {
		bc.Visible = false
		bc.Opacity = 0
		bc.FadeTimer = 0
		return
	}

	bc.Opacity = s.preset.BaseAlpha * (bc.FadeTimer / bc.FadeMax) * 0.5
}

// getOrCreateOverlay retrieves or lazily attaches a bubble overlay component.
func (s *ShieldBubbleOverlaySystem) getOrCreateOverlay(entity *Entity) *ShieldBubbleOverlayComponent {
	if comp, ok := entity.GetComponent("shield_bubble_overlay"); ok {
		if bc, ok := comp.(*ShieldBubbleOverlayComponent); ok {
			return bc
		}
	}
	bc := &ShieldBubbleOverlayComponent{
		PulsePhase: s.rng.Float64() * 2 * math.Pi,
		FadeMax:    s.preset.FadeDur,
		BaseRadius: s.preset.BaseRadius,
	}
	entity.AddComponent(bc)
	return bc
}

// clampShieldBubble ensures a value stays in [0.0, 1.0].
func clampShieldBubble(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
