// Package engine provides the NPCInteractionProximityGlowSystem for visual
// interactability cues. When a player entity is near an NPC with a "dialog" or
// "merchant" component, a genre-aware glow tint is applied to that NPC's
// VisualFeedbackComponent, intensity scaling with proximity. This gives players
// an immediate visual cue that an NPC can be interacted with.
package engine

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// NPCProximityGlowComponent stores per-entity glow state driven by player
// proximity. Pure data — all logic lives in NPCInteractionProximityGlowSystem.
type NPCProximityGlowComponent struct {
	// GlowIntensity is the current glow strength (0.0–1.0).
	GlowIntensity float64
	// PulsePhase tracks the subtle pulse animation (radians).
	PulsePhase float64
	// Active indicates the glow is currently visible.
	Active bool
}

// Type returns the component type identifier.
func (c *NPCProximityGlowComponent) Type() string { return "npc_proximity_glow" }

// npcGlowPreset holds genre-specific glow configuration.
type npcGlowPreset struct {
	Color      color.RGBA // Base glow color
	PulseSpeed float64    // Pulse cycles per second
	PulseAmp   float64    // Pulse amplitude (fraction of intensity, 0.0–0.3)
	MaxTintMix float64    // Maximum tint blend strength (0.0–1.0)
}

// NPCInteractionProximityGlowSystem applies a genre-aware glow tint to NPCs
// that have "dialog" or "merchant" components when a player entity is nearby.
// Intensity scales inversely with distance: full at <=32px, fading to zero at
// the outer glow radius.
type NPCInteractionProximityGlowSystem struct {
	world   *World
	seed    int64
	rng     *rand.Rand
	logger  *logrus.Entry
	genreID string
	preset  npcGlowPreset

	// GlowRadius is the max distance (px) at which the glow appears.
	GlowRadius float64
	// InnerRadius is the distance at which glow is at full intensity.
	InnerRadius float64
	// DecayRate controls how fast glow fades when player leaves (units/s).
	DecayRate float64
}

// NewNPCInteractionProximityGlowSystem creates a new NPC proximity glow system.
func NewNPCInteractionProximityGlowSystem(world *World, seed int64) *NPCInteractionProximityGlowSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "npc_interaction_proximity_glow",
	})

	sys := &NPCInteractionProximityGlowSystem{
		world:       world,
		seed:        seed,
		rng:         rand.New(rand.NewSource(seed)),
		logger:      logger,
		genreID:     "fantasy",
		GlowRadius:  80.0,
		InnerRadius: 32.0,
		DecayRate:   4.0,
	}
	sys.preset = sys.getGenrePreset("fantasy")
	logger.Debug("npc interaction proximity glow system created")
	return sys
}

// SetGenre configures genre-specific glow color and pulse parameters.
func (s *NPCInteractionProximityGlowSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getGenrePreset(genreID)
	s.logger.WithField("genre", genreID).Debug("genre preset applied")
}

// getGenrePreset returns glow configuration for the given genre.
func (s *NPCInteractionProximityGlowSystem) getGenrePreset(genreID string) npcGlowPreset {
	switch genreID {
	case "horror":
		return npcGlowPreset{
			Color:      color.RGBA{R: 80, G: 200, B: 80, A: 255},
			PulseSpeed: 0.6,
			PulseAmp:   0.15,
			MaxTintMix: 0.25,
		}
	case "scifi":
		return npcGlowPreset{
			Color:      color.RGBA{R: 60, G: 200, B: 240, A: 255},
			PulseSpeed: 1.2,
			PulseAmp:   0.1,
			MaxTintMix: 0.3,
		}
	case "cyberpunk":
		return npcGlowPreset{
			Color:      color.RGBA{R: 220, G: 60, B: 240, A: 255},
			PulseSpeed: 1.5,
			PulseAmp:   0.2,
			MaxTintMix: 0.35,
		}
	case "postapoc":
		return npcGlowPreset{
			Color:      color.RGBA{R: 220, G: 170, B: 60, A: 255},
			PulseSpeed: 0.7,
			PulseAmp:   0.12,
			MaxTintMix: 0.25,
		}
	default: // fantasy
		return npcGlowPreset{
			Color:      color.RGBA{R: 255, G: 215, B: 80, A: 255},
			PulseSpeed: 0.8,
			PulseAmp:   0.1,
			MaxTintMix: 0.3,
		}
	}
}

// Update processes all entities, applying proximity glow to interactable NPCs
// near player entities and decaying glow for NPCs no longer in range.
func (s *NPCInteractionProximityGlowSystem) Update(entities []*Entity, deltaTime float64) {
	// Collect player positions
	type playerPos struct {
		x, y float64
	}
	var players []playerPos
	// Collect interactable NPCs
	type npcRef struct {
		entity *Entity
		pos    *PositionComponent
	}
	var npcs []npcRef

	for _, entity := range entities {
		pos := entity.GetPosition()
		if pos == nil {
			continue
		}
		if entity.HasComponent("input") {
			players = append(players, playerPos{pos.X, pos.Y})
		}
		if entity.HasComponent("dialog") || entity.HasComponent("merchant") {
			npcs = append(npcs, npcRef{entity: entity, pos: pos})
		}
	}

	if len(players) == 0 || len(npcs) == 0 {
		// Decay any active glows when no players present
		for _, npc := range npcs {
			s.decayGlow(npc.entity, deltaTime)
		}
		return
	}

	glowRadiusSq := s.GlowRadius * s.GlowRadius

	for _, npc := range npcs {
		// Find closest player distance
		closestDistSq := math.MaxFloat64
		for _, p := range players {
			dx := npc.pos.X - p.x
			dy := npc.pos.Y - p.y
			distSq := dx*dx + dy*dy
			if distSq < closestDistSq {
				closestDistSq = distSq
			}
		}

		glow := s.getOrCreateGlow(npc.entity)

		if closestDistSq <= glowRadiusSq {
			dist := math.Sqrt(closestDistSq)
			// Intensity: 1.0 at InnerRadius, linear falloff to 0.0 at GlowRadius
			var targetIntensity float64
			if dist <= s.InnerRadius {
				targetIntensity = 1.0
			} else {
				targetIntensity = 1.0 - (dist-s.InnerRadius)/(s.GlowRadius-s.InnerRadius)
			}
			if targetIntensity < 0 {
				targetIntensity = 0
			}

			// Smooth approach toward target
			glow.GlowIntensity += (targetIntensity - glow.GlowIntensity) * math.Min(deltaTime*6.0, 1.0)
			glow.Active = true

			// Advance pulse
			glow.PulsePhase += deltaTime * s.preset.PulseSpeed * 2.0 * math.Pi
			if glow.PulsePhase > 2.0*math.Pi {
				glow.PulsePhase -= 2.0 * math.Pi
			}
		} else {
			// Out of range — decay
			glow.GlowIntensity -= deltaTime * s.DecayRate
			if glow.GlowIntensity <= 0.01 {
				glow.GlowIntensity = 0
				glow.Active = false
			}
		}

		// Apply tint to VisualFeedbackComponent
		s.applyGlowTint(npc.entity, glow)
	}
}

// decayGlow reduces glow when no players are around.
func (s *NPCInteractionProximityGlowSystem) decayGlow(entity *Entity, deltaTime float64) {
	comp, ok := entity.GetComponent("npc_proximity_glow")
	if !ok {
		return
	}
	glow, ok := comp.(*NPCProximityGlowComponent)
	if !ok || !glow.Active {
		return
	}
	glow.GlowIntensity -= deltaTime * s.DecayRate
	if glow.GlowIntensity <= 0.01 {
		glow.GlowIntensity = 0
		glow.Active = false
	}
	s.applyGlowTint(entity, glow)
}

// applyGlowTint writes glow color into the entity's VisualFeedbackComponent tint.
func (s *NPCInteractionProximityGlowSystem) applyGlowTint(entity *Entity, glow *NPCProximityGlowComponent) {
	feedback := entity.GetVisualFeedback()
	if feedback == nil {
		feedback = NewVisualFeedbackComponent()
		entity.AddComponent(feedback)
	}

	if !glow.Active || glow.GlowIntensity <= 0 {
		// Reset tint when glow is off
		feedback.TintR = 1.0
		feedback.TintG = 1.0
		feedback.TintB = 1.0
		return
	}

	// Pulse modulates intensity
	pulse := 1.0 + s.preset.PulseAmp*math.Sin(glow.PulsePhase)
	mix := glow.GlowIntensity * s.preset.MaxTintMix * pulse

	// Lerp from white (1,1,1) toward glow color
	glowR := float64(s.preset.Color.R) / 255.0
	glowG := float64(s.preset.Color.G) / 255.0
	glowB := float64(s.preset.Color.B) / 255.0

	feedback.TintR = 1.0 + (glowR-1.0)*mix
	feedback.TintG = 1.0 + (glowG-1.0)*mix
	feedback.TintB = 1.0 + (glowB-1.0)*mix
}

// getOrCreateGlow retrieves or attaches an NPCProximityGlowComponent.
func (s *NPCInteractionProximityGlowSystem) getOrCreateGlow(entity *Entity) *NPCProximityGlowComponent {
	if comp, ok := entity.GetComponent("npc_proximity_glow"); ok {
		if glow, ok := comp.(*NPCProximityGlowComponent); ok {
			return glow
		}
	}
	glow := &NPCProximityGlowComponent{}
	entity.AddComponent(glow)
	return glow
}
