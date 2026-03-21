// Package engine provides the SpecializationStatusResistSystem which bridges class
// specializations with status effect resistance. Tank specializations resist debuffs
// (shorter duration), while damage-focused specializations may have longer durations.
// This creates meaningful trade-offs between defensive and offensive specializations.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// SpecializationStatusResistSystem modifies status effect durations based on
// character class specialization. Tank specializations receive duration reduction
// (20-50%), while glass cannon specializations may have increased duration (10-30%).
type SpecializationStatusResistSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// updateInterval controls how often we process effects (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Track which effects we've already modified to avoid double-processing
	processedEffects map[uint64]map[*StatusEffectComponent]bool

	// Genre affects resistance values
	genreID string
}

// NewSpecializationStatusResistSystem creates a new specialization status resist system.
func NewSpecializationStatusResistSystem(world *World, seed int64) *SpecializationStatusResistSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "specialization_status_resist")
		logEntry.Debug("SpecializationStatusResistSystem created")
	}

	return &SpecializationStatusResistSystem{
		world:            world,
		logger:           logEntry,
		rng:              rand.New(rand.NewSource(seed)),
		updateInterval:   0.1, // Check frequently (status effects are time-sensitive)
		processedEffects: make(map[uint64]map[*StatusEffectComponent]bool),
		genreID:          "fantasy",
	}
}

// SetGenre sets the genre for genre-aware resistance calculations.
func (s *SpecializationStatusResistSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("Genre set for specialization status resist")
	}
}

// Update processes entities and modifies new status effect durations.
func (s *SpecializationStatusResistSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime

	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	for _, entity := range entities {
		s.processEntity(entity)
	}

	// Cleanup processed effects for removed entities
	s.cleanupProcessedEffects(entities)
}

// processEntity checks for new status effects and applies resistance modifiers.
func (s *SpecializationStatusResistSystem) processEntity(entity *Entity) {
	// Need class_progression to calculate resistance
	if !entity.HasComponent("class_progression") {
		return
	}

	classComp, _ := entity.GetComponent("class_progression")
	progression, ok := classComp.(*ClassProgressionComponent)
	if !ok {
		return
	}

	// Calculate resistance modifier for this entity
	resistMod := s.calculateResistanceModifier(progression)

	// Initialize processed effects map for this entity if needed
	if _, exists := s.processedEffects[entity.ID]; !exists {
		s.processedEffects[entity.ID] = make(map[*StatusEffectComponent]bool)
	}

	// Process all status effect components
	for _, comp := range entity.Components {
		effect, ok := comp.(*StatusEffectComponent)
		if ok {
			s.applyResistToEffect(entity, effect, resistMod, progression)
			continue
		}

		// Also check StatusEffectSetComponent which holds multiple effects
		effectSet, ok := comp.(*StatusEffectSetComponent)
		if ok {
			for _, effect := range effectSet.Effects {
				s.applyResistToEffect(entity, effect, resistMod, progression)
			}
		}
	}
}

// applyResistToEffect applies resistance modifier to a single status effect.
func (s *SpecializationStatusResistSystem) applyResistToEffect(entity *Entity, effect *StatusEffectComponent, resistMod float64, progression *ClassProgressionComponent) {
	if effect == nil {
		return
	}

	// Skip already processed effects
	if s.processedEffects[entity.ID][effect] {
		return
	}

	// Mark as processed
	s.processedEffects[entity.ID][effect] = true

	// Only modify debuffs (negative effects)
	if !s.isDebuff(effect.EffectType) {
		return
	}

	// Apply resistance modifier to duration
	originalDuration := effect.Duration
	effect.Duration *= resistMod

	if s.logger != nil && resistMod != 1.0 {
		s.logger.WithFields(logrus.Fields{
			"entity_id":         entity.ID,
			"effect_type":       effect.EffectType,
			"specialization":    progression.Specialization.String(),
			"resist_modifier":   resistMod,
			"original_duration": originalDuration,
			"new_duration":      effect.Duration,
		}).Debug("Status effect duration modified by specialization")
	}
}

// calculateResistanceModifier computes the duration multiplier based on specialization.
// Returns <1.0 for resistance (shorter duration), >1.0 for vulnerability (longer duration).
func (s *SpecializationStatusResistSystem) calculateResistanceModifier(progression *ClassProgressionComponent) float64 {
	// No specialization = no modifier
	if progression.Specialization == SpecializationNone {
		return s.calculateBaseClassModifier(progression.Class)
	}

	// Get specialization-specific resistance
	specResist := s.getSpecializationResistance(progression.Specialization)

	// Genre modifier affects resistance effectiveness
	genreMultiplier := s.getGenreMultiplier()

	// Calculate final modifier: higher specResist = lower duration
	baseModifier := 1.0 - (specResist * genreMultiplier)

	// Clamp to reasonable range [0.5, 1.5]
	if baseModifier < 0.5 {
		baseModifier = 0.5
	}
	if baseModifier > 1.5 {
		baseModifier = 1.5
	}

	return baseModifier
}

// calculateBaseClassModifier provides small resistance for defensive classes.
func (s *SpecializationStatusResistSystem) calculateBaseClassModifier(class CharacterClass) float64 {
	switch class {
	case ClassWarrior:
		return 0.92 // 8% reduction
	case ClassPaladin:
		return 0.90 // 10% reduction
	case ClassDeathKnight, ClassBloodKnight:
		return 0.88 // 12% reduction
	case ClassMonk:
		return 0.94 // 6% reduction (agility-based avoidance)
	case ClassMage:
		return 1.05 // 5% increase (squishy)
	default:
		return 1.0
	}
}

// getSpecializationResistance returns the base resistance value for a specialization.
// Higher values = more resistance = shorter debuff duration.
//
// COMPLEXITY JUSTIFICATION: High cyclomatic complexity (43) is intentional—maps each
// of 43 specialization types to a balance-tuned resistance value. The switch provides
// exhaustiveness checking and clear per-specialization documentation.
func (s *SpecializationStatusResistSystem) getSpecializationResistance(spec SpecializationType) float64 {
	switch spec {
	// Pure tank specializations - highest resistance
	case SpecializationDefender:
		return 0.50 // 50% duration reduction
	case SpecializationGuardian:
		return 0.45 // 45% reduction - holy purification
	case SpecializationBrewmaster:
		return 0.40 // 40% reduction - stagger cleanse
	case SpecializationPackLeader:
		return 0.35 // 35% reduction - pack resilience

	// Hybrid tank/damage specializations - high resistance
	case SpecializationTemplar:
		return 0.35 // 35% reduction - holy fortitude
	case SpecializationCrusader:
		return 0.32 // 32% reduction
	case SpecializationDeathKnight:
		return 0.40 // 40% reduction - undead immunity
	case SpecializationFrost:
		return 0.30 // 30% reduction - cold resistance
	case SpecializationUnholy:
		return 0.25 // 25% reduction - disease immunity

	// Melee damage specializations - moderate resistance
	case SpecializationBerserker:
		return 0.15 // 15% reduction - rage through pain
	case SpecializationDuelist:
		return 0.20 // 20% reduction - combat awareness
	case SpecializationSpellsword:
		return 0.18 // 18% reduction - arcane ward
	case SpecializationCrimsonBlade:
		return 0.22 // 22% reduction - blood resistance
	case SpecializationWindwalker:
		return 0.12 // 12% reduction - mobility
	case SpecializationStriker:
		return 0.10 // 10% reduction

	// Evasion-focused specializations - low resistance (rely on avoidance)
	case SpecializationAssassin:
		return 0.05 // 5% reduction
	case SpecializationShadowdancer:
		return 0.08 // 8% reduction - shadow protection
	case SpecializationTrickster:
		return 0.10 // 10% reduction - illusory cleanse
	case SpecializationShinobi:
		return 0.05 // 5% reduction

	// Pet/summon specializations - moderate resistance
	case SpecializationBeastmaster:
		return 0.15 // 15% reduction - animal bond
	case SpecializationDemonologist:
		return 0.20 // 20% reduction - demonic pact
	case SpecializationFeralRage:
		return 0.25 // 25% reduction - primal constitution
	case SpecializationShapeshifter:
		return 0.28 // 28% reduction - form adaptation

	// Healer specializations - moderate resistance (self-cleanse)
	case SpecializationHealer:
		return 0.30 // 30% reduction - divine cleansing
	case SpecializationSoulweaver:
		return 0.25 // 25% reduction - spirit purification
	case SpecializationOracle:
		return 0.22 // 22% reduction - foresight
	case SpecializationTheurgist:
		return 0.28 // 28% reduction - holy resistance
	case SpecializationNaturemage:
		return 0.20 // 20% reduction - natural immunity

	// Pure caster specializations - low resistance (glass cannon)
	case SpecializationArcanist:
		return -0.10 // 10% INCREASE - pure glass cannon
	case SpecializationElementalist:
		return 0.05 // 5% reduction - elemental attunement
	case SpecializationBloodMage:
		return 0.12 // 12% reduction - blood resistance
	case SpecializationWarmage:
		return 0.15 // 15% reduction - battle training
	case SpecializationVoidcaller:
		return -0.05 // 5% INCREASE - void corruption
	case SpecializationAffliction:
		return 0.08 // 8% reduction - curse familiarity
	case SpecializationHemomancer:
		return 0.10 // 10% reduction - blood magic

	// Ranged specializations - low resistance
	case SpecializationMarksman:
		return 0.08 // 8% reduction
	case SpecializationSpellshot:
		return 0.10 // 10% reduction
	case SpecializationSeeker:
		return 0.06 // 6% reduction

	// Paladin specializations - moderate to high resistance
	case SpecializationExorcist:
		return 0.35 // 35% reduction - holy purification
	case SpecializationPurifier:
		return 0.40 // 40% reduction - cleansing mastery
	case SpecializationJudge:
		return 0.30 // 30% reduction - righteous shield
	case SpecializationInterrogator:
		return 0.25 // 25% reduction - mental fortitude

	default:
		return 0.10 // 10% default resistance
	}
}

// getGenreMultiplier returns a multiplier affecting resistance effectiveness.
func (s *SpecializationStatusResistSystem) getGenreMultiplier() float64 {
	switch s.genreID {
	case "fantasy":
		return 1.0 // Standard fantasy resistances
	case "scifi":
		return 1.1 // Tech-enhanced resistance
	case "horror":
		return 0.7 // Debuffs more dangerous in horror
	case "cyberpunk":
		return 1.05 // Cyber-implant resistance
	case "postapoc":
		return 0.85 // Weakened immune systems
	default:
		return 1.0
	}
}

// isDebuff returns true if the effect type is a negative status effect.
func (s *SpecializationStatusResistSystem) isDebuff(effectType string) bool {
	debuffTypes := map[string]bool{
		"poison":        true,
		"burn":          true,
		"burning":       true,
		"bleed":         true,
		"bleeding":      true,
		"stun":          true,
		"stunned":       true,
		"slow":          true,
		"slowed":        true,
		"frozen":        true,
		"freeze":        true,
		"fear":          true,
		"feared":        true,
		"blind":         true,
		"blinded":       true,
		"silence":       true,
		"silenced":      true,
		"curse":         true,
		"cursed":        true,
		"weakness":      true,
		"weakened":      true,
		"vulnerable":    true,
		"vulnerability": true,
		"chill":         true,
		"chilled":       true,
		"wet":           true,
		"shock":         true,
		"shocked":       true,
		"paralysis":     true,
		"paralyzed":     true,
		"root":          true,
		"rooted":        true,
		"confuse":       true,
		"confused":      true,
		"sleep":         true,
		"sleeping":      true,
		"disease":       true,
		"diseased":      true,
		"decay":         true,
		"corruption":    true,
		"corrupted":     true,
	}
	return debuffTypes[effectType]
}

// cleanupProcessedEffects removes tracking for entities that no longer exist.
func (s *SpecializationStatusResistSystem) cleanupProcessedEffects(entities []*Entity) {
	// Build set of current entity IDs
	currentIDs := make(map[uint64]bool)
	for _, e := range entities {
		currentIDs[e.ID] = true

		// Also cleanup expired effects for existing entities
		if effects, exists := s.processedEffects[e.ID]; exists {
			for effect := range effects {
				if effect.IsExpired() {
					delete(effects, effect)
				}
			}
		}
	}

	// Remove entries for entities that no longer exist
	for id := range s.processedEffects {
		if !currentIDs[id] {
			delete(s.processedEffects, id)
		}
	}
}

// GetResistanceForEntity returns the current resistance modifier for an entity (for UI).
func (s *SpecializationStatusResistSystem) GetResistanceForEntity(entity *Entity) float64 {
	if !entity.HasComponent("class_progression") {
		return 1.0
	}

	classComp, _ := entity.GetComponent("class_progression")
	progression, ok := classComp.(*ClassProgressionComponent)
	if !ok {
		return 1.0
	}

	return s.calculateResistanceModifier(progression)
}
