// Package engine provides the EntityFactionOutlineSystem for genre-aware
// colored outlines around entities based on team allegiance and faction
// reputation. Allies show green outlines, hostiles red, neutrals amber.
// Genre presets adjust saturation and intensity: fantasy uses warm muted
// tones, horror desaturated red emphasis, sci-fi electric neon, cyberpunk
// high-contrast neon, post-apocalyptic dusty amber.
package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// EntityFactionOutlineComponent stores per-entity outline visual parameters.
// Pure data — all logic lives in EntityFactionOutlineSystem.
type EntityFactionOutlineComponent struct {
	// Outline color (RGB 0.0-1.0)
	OutlineR float64
	OutlineG float64
	OutlineB float64

	// Outline intensity (0.0-1.0), controls glow strength
	Intensity float64

	// Pulse phase for hostile entities (radians)
	PulsePhase float64

	// Pulse speed for hostile entities (radians/sec)
	PulseSpeed float64

	// Outline thickness in pixels (1.0-3.0)
	Thickness float64

	// Allegiance category for rendering: "ally", "hostile", "neutral", "none"
	Allegiance string

	// Whether the outline is currently visible
	Visible bool
}

// Type returns the component type identifier.
func (c *EntityFactionOutlineComponent) Type() string { return "entity_faction_outline" }

// factionOutlinePreset holds genre-specific outline parameters.
type factionOutlinePreset struct {
	// Per-allegiance colors
	AllyR, AllyG, AllyB       float64
	HostileR, HostileG, HostileB float64
	NeutralR, NeutralG, NeutralB float64

	// Base intensity per allegiance
	AllyIntensity    float64
	HostileIntensity float64
	NeutralIntensity float64

	// Hostile pulse speed
	HostilePulseSpeed float64

	// Outline thickness
	Thickness float64
}

// EntityFactionOutlineSystem assigns genre-aware colored outlines to entities
// based on their TeamComponent allegiance. It reads TeamComponent.TeamID
// (1=player team=ally, 2+=enemy=hostile, 0=neutral) and optionally
// FactionComponent.Reputation for finer-grained coloring.
type EntityFactionOutlineSystem struct {
	world  *World
	seed   int64
	rng    *rand.Rand
	logger *logrus.Entry

	genreID string
	preset  factionOutlinePreset

	// Throttle state checks to avoid per-frame overhead
	checkInterval  float64
	timeSinceCheck float64
}

// NewEntityFactionOutlineSystem creates a new faction outline system.
func NewEntityFactionOutlineSystem(world *World, seed int64) *EntityFactionOutlineSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "entity_faction_outline",
	})

	sys := &EntityFactionOutlineSystem{
		world:          world,
		seed:           seed,
		rng:            rand.New(rand.NewSource(seed)),
		logger:         logger,
		genreID:        "fantasy",
		checkInterval:  0.25,
		timeSinceCheck: 0,
	}
	sys.preset = sys.getGenrePreset("fantasy")
	logger.Debug("entity faction outline system created")
	return sys
}

// SetGenre configures genre-specific outline parameters.
func (s *EntityFactionOutlineSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getGenrePreset(genreID)
	s.logger.WithField("genre", genreID).Debug("genre preset applied")
}

// getGenrePreset returns outline configuration for the given genre.
func (s *EntityFactionOutlineSystem) getGenrePreset(genreID string) factionOutlinePreset {
	switch genreID {
	case "horror":
		return factionOutlinePreset{
			AllyR: 0.3, AllyG: 0.6, AllyB: 0.3,
			HostileR: 0.8, HostileG: 0.15, HostileB: 0.1,
			NeutralR: 0.6, NeutralG: 0.5, NeutralB: 0.2,
			AllyIntensity: 0.35, HostileIntensity: 0.6, NeutralIntensity: 0.25,
			HostilePulseSpeed: 3.5, Thickness: 1.5,
		}
	case "scifi":
		return factionOutlinePreset{
			AllyR: 0.2, AllyG: 0.9, AllyB: 0.4,
			HostileR: 0.9, HostileG: 0.2, HostileB: 0.2,
			NeutralR: 0.8, NeutralG: 0.8, NeutralB: 0.3,
			AllyIntensity: 0.5, HostileIntensity: 0.65, NeutralIntensity: 0.35,
			HostilePulseSpeed: 4.0, Thickness: 1.0,
		}
	case "cyberpunk":
		return factionOutlinePreset{
			AllyR: 0.1, AllyG: 0.95, AllyB: 0.5,
			HostileR: 0.95, HostileG: 0.1, HostileB: 0.3,
			NeutralR: 0.9, NeutralG: 0.85, NeutralB: 0.1,
			AllyIntensity: 0.55, HostileIntensity: 0.7, NeutralIntensity: 0.4,
			HostilePulseSpeed: 4.5, Thickness: 1.0,
		}
	case "postapoc":
		return factionOutlinePreset{
			AllyR: 0.4, AllyG: 0.65, AllyB: 0.3,
			HostileR: 0.75, HostileG: 0.25, HostileB: 0.15,
			NeutralR: 0.7, NeutralG: 0.6, NeutralB: 0.25,
			AllyIntensity: 0.35, HostileIntensity: 0.5, NeutralIntensity: 0.3,
			HostilePulseSpeed: 2.5, Thickness: 2.0,
		}
	default: // fantasy
		return factionOutlinePreset{
			AllyR: 0.25, AllyG: 0.75, AllyB: 0.3,
			HostileR: 0.85, HostileG: 0.2, HostileB: 0.15,
			NeutralR: 0.8, NeutralG: 0.7, NeutralB: 0.2,
			AllyIntensity: 0.4, HostileIntensity: 0.55, NeutralIntensity: 0.3,
			HostilePulseSpeed: 3.0, Thickness: 1.5,
		}
	}
}

// Update processes all entities, updating allegiance on a throttled interval
// and animating hostile pulse every frame.
func (s *EntityFactionOutlineSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime
	checkAllegiance := s.timeSinceCheck >= s.checkInterval
	if checkAllegiance {
		s.timeSinceCheck = 0
	}

	for _, entity := range entities {
		if entity == nil {
			continue
		}
		if !entity.HasComponent("position") || !entity.HasComponent("sprite") {
			continue
		}
		// Skip entities without team affiliation
		if !entity.HasComponent("team") {
			continue
		}

		oc := s.getOrCreateOutline(entity)

		if checkAllegiance {
			s.updateAllegiance(entity, oc)
		}

		// Animate pulse for hostile entities every frame
		if oc.Visible && oc.Allegiance == "hostile" {
			oc.PulsePhase += oc.PulseSpeed * deltaTime
			if oc.PulsePhase > 2*math.Pi {
				oc.PulsePhase -= 2 * math.Pi
			}
			// Modulate intensity with pulse (base ± 20%)
			baseMod := 0.2 * math.Sin(oc.PulsePhase)
			oc.Intensity = clampOutline(s.preset.HostileIntensity + baseMod)
		}
	}
}

// updateAllegiance reads TeamComponent (and optionally FactionComponent)
// to determine the outline color and allegiance category.
func (s *EntityFactionOutlineSystem) updateAllegiance(entity *Entity, oc *EntityFactionOutlineComponent) {
	teamComp, ok := entity.GetComponent("team")
	if !ok {
		oc.Visible = false
		return
	}
	team, ok := teamComp.(*TeamComponent)
	if !ok {
		oc.Visible = false
		return
	}

	p := s.preset

	switch {
	case team.TeamID == 1:
		// Player team / ally — skip outlines on the player's own entity
		if entity.HasComponent("input") {
			oc.Visible = false
			return
		}
		oc.Allegiance = "ally"
		oc.OutlineR = p.AllyR
		oc.OutlineG = p.AllyG
		oc.OutlineB = p.AllyB
		oc.Intensity = p.AllyIntensity
		oc.PulseSpeed = 0
		oc.Thickness = p.Thickness

	case team.TeamID == 0:
		oc.Allegiance = "neutral"
		oc.OutlineR = p.NeutralR
		oc.OutlineG = p.NeutralG
		oc.OutlineB = p.NeutralB
		oc.Intensity = p.NeutralIntensity
		oc.PulseSpeed = 0
		oc.Thickness = p.Thickness

	default:
		oc.Allegiance = "hostile"
		oc.OutlineR = p.HostileR
		oc.OutlineG = p.HostileG
		oc.OutlineB = p.HostileB
		oc.Intensity = p.HostileIntensity
		oc.PulseSpeed = p.HostilePulseSpeed
		oc.Thickness = p.Thickness
	}

	// Refine with faction reputation if available
	if fComp, ok := entity.GetComponent("faction"); ok {
		if fc, ok := fComp.(*FactionComponent); ok {
			s.applyFactionReputation(oc, fc)
		}
	}

	// Per-entity subtle color variation
	variation := s.rng.Float64()*0.06 - 0.03
	oc.OutlineR = clampOutline(oc.OutlineR + variation)
	oc.OutlineG = clampOutline(oc.OutlineG + variation)
	oc.OutlineB = clampOutline(oc.OutlineB + variation)

	oc.Visible = true
}

// applyFactionReputation adjusts outline based on faction reputation level.
func (s *EntityFactionOutlineSystem) applyFactionReputation(oc *EntityFactionOutlineComponent, fc *FactionComponent) {
	if !fc.IsPlayerFaction {
		return
	}
	p := s.preset

	switch {
	case fc.IsFriendly():
		oc.Allegiance = "ally"
		oc.OutlineR = p.AllyR
		oc.OutlineG = p.AllyG
		oc.OutlineB = p.AllyB
		oc.Intensity = p.AllyIntensity
		oc.PulseSpeed = 0
	case fc.IsHostile():
		oc.Allegiance = "hostile"
		oc.OutlineR = p.HostileR
		oc.OutlineG = p.HostileG
		oc.OutlineB = p.HostileB
		oc.Intensity = p.HostileIntensity
		oc.PulseSpeed = p.HostilePulseSpeed
	case fc.IsSuspicious():
		oc.Allegiance = "neutral"
		oc.OutlineR = p.NeutralR * 0.9
		oc.OutlineG = p.NeutralG * 0.7
		oc.OutlineB = p.NeutralB
		oc.Intensity = p.NeutralIntensity * 0.8
		oc.PulseSpeed = 0
	}
}

// getOrCreateOutline retrieves or lazily attaches an outline component.
func (s *EntityFactionOutlineSystem) getOrCreateOutline(entity *Entity) *EntityFactionOutlineComponent {
	if comp, ok := entity.GetComponent("entity_faction_outline"); ok {
		if oc, ok := comp.(*EntityFactionOutlineComponent); ok {
			return oc
		}
	}
	oc := &EntityFactionOutlineComponent{
		Allegiance: "none",
		PulsePhase: s.rng.Float64() * 2 * math.Pi,
	}
	entity.AddComponent(oc)
	return oc
}

// clampOutline ensures a value stays in [0.0, 1.0].
func clampOutline(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
