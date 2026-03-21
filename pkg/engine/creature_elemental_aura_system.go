// Package engine provides the CreatureElementalAuraSystem which assigns
// genre-aware elemental aura visual effects to creatures based on their
// names, tags, or faction associations. Fire creatures (fire wolf, flame
// spider) get warm orange-red auras, ice creatures get cyan-white shimmer,
// poison creatures get sickly green glow, etc. The system infers elemental
// affinity from entity data and provides persistent visual feedback that
// makes elemental creature types immediately distinguishable.
package engine

import (
	"math"
	"math/rand"
	"strings"

	"github.com/opd-ai/venture/pkg/procgen/magic"
	"github.com/sirupsen/logrus"
)

// elementAuraPreset holds color and intensity values for one element type.
type elementAuraPreset struct {
	PrimaryR, PrimaryG, PrimaryB       float64
	SecondaryR, SecondaryG, SecondaryB float64
	BaseIntensity                      float64
	PulseSpeed                         float64
	PulseAmplitude                     float64
	AuraRadius                         float64
	ParticleEmission                   bool
	ParticleRate                       float64
}

// genreElementModifier adjusts element aura appearance per genre.
type genreElementModifier struct {
	IntensityMult float64
	SaturationAdj float64 // -1.0 to 1.0, shifts color saturation
	PulseMult     float64
}

// CreatureElementalAuraSystem scans creatures without elemental aura components
// and infers their elemental affinity from name/tag keywords, then assigns
// appropriate visual aura parameters. Runs once per entity (idempotent).
// Also updates pulse animation each frame for active aura components.
type CreatureElementalAuraSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	genreID     string
	lastGenreID string

	presets   map[magic.ElementType]elementAuraPreset
	genreMods map[string]genreElementModifier

	// updateInterval throttles new-entity scanning
	updateInterval float64
	timeSinceCheck float64
}

// NewCreatureElementalAuraSystem creates a new creature elemental aura system.
func NewCreatureElementalAuraSystem(world *World, seed int64) *CreatureElementalAuraSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "creature_elemental_aura")
	}

	sys := &CreatureElementalAuraSystem{
		world:          world,
		logger:         logEntry,
		rng:            rand.New(rand.NewSource(seed)),
		genreID:        "fantasy",
		lastGenreID:    "",
		presets:        buildElementAuraPresets(),
		genreMods:      buildGenreElementModifiers(),
		updateInterval: 1.0,
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("creature elemental aura system created")
	}
	return sys
}

// SetGenre configures genre-specific aura appearance modifiers.
func (s *CreatureElementalAuraSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithField("genre", genreID).Debug("genre set for creature elemental aura")
	}
}

// Update processes entities: assigns aura to new elemental creatures and
// updates pulse animation for existing aura components.
func (s *CreatureElementalAuraSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime

	fullScan := false
	if s.timeSinceCheck >= s.updateInterval {
		s.timeSinceCheck = 0
		fullScan = true
	}

	for _, entity := range entities {
		comp, hasComp := entity.GetComponent("creature_elemental_aura")

		if hasComp {
			// Update existing aura animation
			aura, ok := comp.(*CreatureElementalAuraComponent)
			if ok && aura.Enabled {
				s.updateAuraPulse(aura, deltaTime)
			}
			continue
		}

		// Only scan for new entities periodically
		if !fullScan {
			continue
		}

		// Only assign auras to enemy creatures (team 2)
		teamComp, hasTeam := entity.GetComponent("team")
		if !hasTeam {
			continue
		}
		team, ok := teamComp.(*TeamComponent)
		if !ok || team.TeamID != 2 {
			continue
		}

		// Infer elemental affinity from entity data
		element := s.inferElement(entity)
		if element == magic.ElementNone {
			continue
		}

		// Create and configure the aura component
		auraComp := s.createAuraComponent(entity, element)
		entity.AddComponent(auraComp)

		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"element":   element.String(),
			}).Debug("assigned elemental aura to creature")
		}
	}
}

// updateAuraPulse advances the pulse animation for an active aura.
func (s *CreatureElementalAuraSystem) updateAuraPulse(aura *CreatureElementalAuraComponent, dt float64) {
	if aura.PulseSpeed <= 0 {
		aura.CurrentIntensity = aura.BaseIntensity
		return
	}

	aura.PulsePhase += aura.PulseSpeed * 2 * math.Pi * dt
	if aura.PulsePhase > 2*math.Pi {
		aura.PulsePhase -= 2 * math.Pi
	}

	pulse := math.Sin(aura.PulsePhase) * aura.PulseAmplitude
	aura.CurrentIntensity = clampFloatAura(aura.BaseIntensity+pulse, 0, 1)
}

// inferElement determines elemental affinity from entity name, tags, and faction.
func (s *CreatureElementalAuraSystem) inferElement(entity *Entity) magic.ElementType {
	// Check entity name for elemental keywords
	if nameComp, ok := entity.GetComponent("name"); ok {
		if nc, ok := nameComp.(*NameComponent); ok {
			if elem := s.elementFromKeywords(nc.Name); elem != magic.ElementNone {
				return elem
			}
		}
	}

	// Check creature visual tags
	if visualComp, ok := entity.GetComponent("creature_visual"); ok {
		if cv, ok := visualComp.(*CreatureVisualComponent); ok {
			for _, tag := range cv.VisualTags {
				if elem := s.elementFromKeywords(tag); elem != magic.ElementNone {
					return elem
				}
			}
		}
	}

	// Check faction for elemental hints
	if factionComp, ok := entity.GetComponent("faction"); ok {
		if fc, ok := factionComp.(*FactionComponent); ok {
			if elem := s.elementFromKeywords(fc.FactionID); elem != magic.ElementNone {
				return elem
			}
		}
	}

	// Check attack component for elemental damage type
	if attackComp, ok := entity.GetComponent("attack"); ok {
		if ac, ok := attackComp.(*AttackComponent); ok {
			if elem := s.elementFromDamageType(ac.DamageType.String()); elem != magic.ElementNone {
				return elem
			}
		}
	}

	return magic.ElementNone
}

// elementFromKeywords maps name/tag keywords to element types.
//
// COMPLEXITY JUSTIFICATION: High cyclomatic complexity (60) is intentional—this is
// a keyword matching table that must check many genre-appropriate keywords for each
// element type. The if-chain structure is clearer than alternatives (regex, maps)
// and the complexity is purely from the number of keywords, not control flow depth.
func (s *CreatureElementalAuraSystem) elementFromKeywords(text string) magic.ElementType {
	lower := strings.ToLower(text)

	// Fire keywords
	if strings.Contains(lower, "fire") || strings.Contains(lower, "flame") ||
		strings.Contains(lower, "inferno") || strings.Contains(lower, "burn") ||
		strings.Contains(lower, "magma") || strings.Contains(lower, "lava") ||
		strings.Contains(lower, "ember") || strings.Contains(lower, "blaze") ||
		strings.Contains(lower, "pyro") || strings.Contains(lower, "scorch") {
		return magic.ElementFire
	}

	// Ice keywords
	if strings.Contains(lower, "ice") || strings.Contains(lower, "frost") ||
		strings.Contains(lower, "frozen") || strings.Contains(lower, "cold") ||
		strings.Contains(lower, "snow") || strings.Contains(lower, "blizzard") ||
		strings.Contains(lower, "glacier") || strings.Contains(lower, "cryo") ||
		strings.Contains(lower, "chill") || strings.Contains(lower, "winter") {
		return magic.ElementIce
	}

	// Lightning keywords
	if strings.Contains(lower, "lightning") || strings.Contains(lower, "thunder") ||
		strings.Contains(lower, "storm") || strings.Contains(lower, "electric") ||
		strings.Contains(lower, "shock") || strings.Contains(lower, "volt") ||
		strings.Contains(lower, "spark") || strings.Contains(lower, "static") {
		return magic.ElementLightning
	}

	// Poison/toxic keywords (mapped to Earth for visual purposes)
	if strings.Contains(lower, "poison") || strings.Contains(lower, "toxic") ||
		strings.Contains(lower, "venom") || strings.Contains(lower, "acid") ||
		strings.Contains(lower, "plague") || strings.Contains(lower, "corrupt") ||
		strings.Contains(lower, "blight") || strings.Contains(lower, "rot") {
		return magic.ElementEarth // Use Earth for poison green
	}

	// Wind keywords
	if strings.Contains(lower, "wind") || strings.Contains(lower, "air") ||
		strings.Contains(lower, "gale") || strings.Contains(lower, "zephyr") ||
		strings.Contains(lower, "tempest") || strings.Contains(lower, "breeze") {
		return magic.ElementWind
	}

	// Light keywords
	if strings.Contains(lower, "holy") || strings.Contains(lower, "light") ||
		strings.Contains(lower, "radiant") || strings.Contains(lower, "celestial") ||
		strings.Contains(lower, "divine") || strings.Contains(lower, "angel") {
		return magic.ElementLight
	}

	// Dark keywords
	if strings.Contains(lower, "dark") || strings.Contains(lower, "shadow") ||
		strings.Contains(lower, "void") || strings.Contains(lower, "abyss") ||
		strings.Contains(lower, "nightmare") || strings.Contains(lower, "demon") {
		return magic.ElementDark
	}

	// Arcane keywords
	if strings.Contains(lower, "arcane") || strings.Contains(lower, "magic") ||
		strings.Contains(lower, "mystic") || strings.Contains(lower, "ethereal") ||
		strings.Contains(lower, "mana") {
		return magic.ElementArcane
	}

	return magic.ElementNone
}

// elementFromDamageType maps combat damage types to elements.
func (s *CreatureElementalAuraSystem) elementFromDamageType(dmgType string) magic.ElementType {
	lower := strings.ToLower(dmgType)
	switch {
	case strings.Contains(lower, "fire"):
		return magic.ElementFire
	case strings.Contains(lower, "ice"), strings.Contains(lower, "cold"), strings.Contains(lower, "frost"):
		return magic.ElementIce
	case strings.Contains(lower, "lightning"), strings.Contains(lower, "electric"):
		return magic.ElementLightning
	case strings.Contains(lower, "poison"), strings.Contains(lower, "toxic"):
		return magic.ElementEarth
	default:
		return magic.ElementNone
	}
}

// createAuraComponent builds an aura component with element-specific visuals.
func (s *CreatureElementalAuraSystem) createAuraComponent(entity *Entity, element magic.ElementType) *CreatureElementalAuraComponent {
	preset, ok := s.presets[element]
	if !ok {
		preset = s.presets[magic.ElementNone]
	}

	// Apply genre modifiers
	genreMod, hasGenre := s.genreMods[s.genreID]
	if !hasGenre {
		genreMod = genreElementModifier{IntensityMult: 1.0, SaturationAdj: 0.0, PulseMult: 1.0}
	}

	// Seed-based variation for each entity
	seedVar := s.rng.Float64()*0.2 - 0.1 // ±10% variation

	comp := &CreatureElementalAuraComponent{
		Element:          element,
		AuraR:            clampFloatAura(preset.PrimaryR*(1+seedVar), 0, 1),
		AuraG:            clampFloatAura(preset.PrimaryG*(1+seedVar), 0, 1),
		AuraB:            clampFloatAura(preset.PrimaryB*(1+seedVar), 0, 1),
		BaseIntensity:    clampFloatAura(preset.BaseIntensity*genreMod.IntensityMult, 0, 1),
		CurrentIntensity: preset.BaseIntensity * genreMod.IntensityMult,
		PulseSpeed:       preset.PulseSpeed * genreMod.PulseMult,
		PulseAmplitude:   preset.PulseAmplitude,
		PulsePhase:       s.rng.Float64() * 2 * math.Pi, // Random start phase
		AuraRadius:       preset.AuraRadius + seedVar*2,
		SecondaryR:       preset.SecondaryR,
		SecondaryG:       preset.SecondaryG,
		SecondaryB:       preset.SecondaryB,
		ParticleEmission: preset.ParticleEmission,
		ParticleRate:     preset.ParticleRate,
		Enabled:          true,
	}

	return comp
}

// buildElementAuraPresets creates the color/intensity presets for each element.
func buildElementAuraPresets() map[magic.ElementType]elementAuraPreset {
	return map[magic.ElementType]elementAuraPreset{
		magic.ElementNone: {
			PrimaryR: 0.5, PrimaryG: 0.5, PrimaryB: 0.5,
			SecondaryR: 0.6, SecondaryG: 0.6, SecondaryB: 0.6,
			BaseIntensity:    0.3,
			PulseSpeed:       0.5,
			PulseAmplitude:   0.1,
			AuraRadius:       6.0,
			ParticleEmission: false,
			ParticleRate:     0,
		},
		magic.ElementFire: {
			PrimaryR: 1.0, PrimaryG: 0.45, PrimaryB: 0.1,
			SecondaryR: 1.0, SecondaryG: 0.8, SecondaryB: 0.2,
			BaseIntensity:    0.65,
			PulseSpeed:       2.5, // Fast flicker
			PulseAmplitude:   0.25,
			AuraRadius:       8.0,
			ParticleEmission: true,
			ParticleRate:     4.0, // Fire sparks
		},
		magic.ElementIce: {
			PrimaryR: 0.6, PrimaryG: 0.85, PrimaryB: 1.0,
			SecondaryR: 0.9, SecondaryG: 0.95, SecondaryB: 1.0,
			BaseIntensity:    0.55,
			PulseSpeed:       0.8, // Slow shimmer
			PulseAmplitude:   0.15,
			AuraRadius:       7.0,
			ParticleEmission: true,
			ParticleRate:     2.0, // Ice crystals
		},
		magic.ElementLightning: {
			PrimaryR: 0.7, PrimaryG: 0.8, PrimaryB: 1.0,
			SecondaryR: 1.0, SecondaryG: 1.0, SecondaryB: 0.9,
			BaseIntensity:    0.7,
			PulseSpeed:       8.0, // Rapid crackle
			PulseAmplitude:   0.4,
			AuraRadius:       9.0,
			ParticleEmission: true,
			ParticleRate:     6.0, // Sparks
		},
		magic.ElementEarth: { // Used for poison/toxic visuals
			PrimaryR: 0.35, PrimaryG: 0.8, PrimaryB: 0.2,
			SecondaryR: 0.5, SecondaryG: 0.9, SecondaryB: 0.3,
			BaseIntensity:    0.5,
			PulseSpeed:       1.2,
			PulseAmplitude:   0.2,
			AuraRadius:       6.5,
			ParticleEmission: true,
			ParticleRate:     3.0, // Toxic bubbles
		},
		magic.ElementWind: {
			PrimaryR: 0.8, PrimaryG: 0.9, PrimaryB: 0.95,
			SecondaryR: 0.95, SecondaryG: 1.0, SecondaryB: 1.0,
			BaseIntensity:    0.4,
			PulseSpeed:       1.5,
			PulseAmplitude:   0.2,
			AuraRadius:       10.0,
			ParticleEmission: true,
			ParticleRate:     5.0, // Wind wisps
		},
		magic.ElementLight: {
			PrimaryR: 1.0, PrimaryG: 0.95, PrimaryB: 0.7,
			SecondaryR: 1.0, SecondaryG: 1.0, SecondaryB: 0.9,
			BaseIntensity:    0.75,
			PulseSpeed:       0.6,
			PulseAmplitude:   0.1,
			AuraRadius:       10.0,
			ParticleEmission: true,
			ParticleRate:     2.0, // Light motes
		},
		magic.ElementDark: {
			PrimaryR: 0.3, PrimaryG: 0.1, PrimaryB: 0.4,
			SecondaryR: 0.15, SecondaryG: 0.0, SecondaryB: 0.25,
			BaseIntensity:    0.6,
			PulseSpeed:       1.0,
			PulseAmplitude:   0.25,
			AuraRadius:       8.0,
			ParticleEmission: true,
			ParticleRate:     3.0, // Shadow wisps
		},
		magic.ElementArcane: {
			PrimaryR: 0.7, PrimaryG: 0.4, PrimaryB: 0.9,
			SecondaryR: 0.9, SecondaryG: 0.6, SecondaryB: 1.0,
			BaseIntensity:    0.6,
			PulseSpeed:       1.8,
			PulseAmplitude:   0.2,
			AuraRadius:       9.0,
			ParticleEmission: true,
			ParticleRate:     4.0, // Magic sparkles
		},
	}
}

// buildGenreElementModifiers creates genre-specific adjustments.
func buildGenreElementModifiers() map[string]genreElementModifier {
	return map[string]genreElementModifier{
		"fantasy": {
			IntensityMult: 1.0,
			SaturationAdj: 0.0,
			PulseMult:     1.0,
		},
		"horror": {
			IntensityMult: 0.8,  // Dimmer, more menacing
			SaturationAdj: -0.2, // Desaturated
			PulseMult:     0.7,  // Slower, more ominous
		},
		"sci-fi": {
			IntensityMult: 1.2, // Brighter, more vibrant
			SaturationAdj: 0.1, // Slightly saturated
			PulseMult:     1.3, // Faster, more energetic
		},
		"cyberpunk": {
			IntensityMult: 1.1,
			SaturationAdj: 0.2, // Neon-saturated
			PulseMult:     1.5, // Flickering neon effect
		},
		"post-apocalyptic": {
			IntensityMult: 0.9,
			SaturationAdj: -0.1, // Slightly muted
			PulseMult:     0.9,
		},
	}
}

// clampFloatAura clamps a float64 to [min, max].
func clampFloatAura(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
