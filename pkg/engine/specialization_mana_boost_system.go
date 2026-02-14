// Package engine provides the SpecializationManaBoostSystem which bridges class
// specializations with mana regeneration. This system grants passive mana regen
// bonuses to specialized mages and caster classes, creating meaningful gameplay
// incentives for specialization choices.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// SpecializationManaBoostSystem provides mana regeneration bonuses based on
// character class specialization. Magic-focused specializations receive higher
// bonuses (20-50%), hybrid specializations receive moderate bonuses (10-25%),
// and melee specializations receive no bonus.
type SpecializationManaBoostSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// updateInterval controls how often we check specializations (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Cache original regen rates per entity to apply bonus correctly
	originalRegen map[uint64]float64

	// Track which entities have bonuses applied
	appliedBonuses map[uint64]float64

	// Genre affects bonus amounts (scifi/cyberpunk tech-casters vs fantasy mages)
	genreID string
}

// NewSpecializationManaBoostSystem creates a new specialization mana boost system.
func NewSpecializationManaBoostSystem(world *World, seed int64) *SpecializationManaBoostSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "specialization_mana_boost")
		logEntry.Debug("SpecializationManaBoostSystem created")
	}

	return &SpecializationManaBoostSystem{
		world:          world,
		logger:         logEntry,
		rng:            rand.New(rand.NewSource(seed)),
		updateInterval: 1.0, // Check once per second (specializations rarely change)
		originalRegen:  make(map[uint64]float64),
		appliedBonuses: make(map[uint64]float64),
		genreID:        "fantasy",
	}
}

// SetGenre sets the genre for genre-aware bonus calculations.
func (s *SpecializationManaBoostSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("Genre set for specialization mana boost")
	}
}

// Update checks class specializations and applies mana regeneration bonuses.
func (s *SpecializationManaBoostSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime

	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	for _, entity := range entities {
		s.processEntity(entity)
	}
}

// processEntity checks if an entity has a specialization and applies mana bonus.
func (s *SpecializationManaBoostSystem) processEntity(entity *Entity) {
	// Need both mana and class_progression components
	if !entity.HasComponent("mana") || !entity.HasComponent("class_progression") {
		// If entity lost components, remove any stored bonus
		s.removeBonus(entity.ID)
		return
	}

	manaComp, _ := entity.GetComponent("mana")
	mana, ok := manaComp.(*ManaComponent)
	if !ok {
		return
	}

	classComp, _ := entity.GetComponent("class_progression")
	progression, ok := classComp.(*ClassProgressionComponent)
	if !ok {
		return
	}

	// Calculate bonus based on class and specialization
	bonus := s.calculateBonus(progression)

	// Check if bonus changed
	currentBonus, hasBonus := s.appliedBonuses[entity.ID]
	if hasBonus && currentBonus == bonus {
		return // No change needed
	}

	// Store original regen if not already stored
	if _, exists := s.originalRegen[entity.ID]; !exists {
		s.originalRegen[entity.ID] = mana.Regen
	}

	// Apply new bonus
	originalRegen := s.originalRegen[entity.ID]
	newRegen := originalRegen * (1.0 + bonus)
	mana.Regen = newRegen
	s.appliedBonuses[entity.ID] = bonus

	if s.logger != nil && bonus > 0 {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entity.ID,
			"class":          progression.Class.String(),
			"specialization": progression.Specialization.String(),
			"bonus_percent":  bonus * 100,
			"original_regen": originalRegen,
			"new_regen":      newRegen,
		}).Debug("Specialization mana bonus applied")
	}
}

// calculateBonus computes the mana regen bonus multiplier for a specialization.
// Returns 0.0-0.5 (0-50% bonus) based on specialization and genre.
func (s *SpecializationManaBoostSystem) calculateBonus(progression *ClassProgressionComponent) float64 {
	// No bonus without specialization
	if progression.Specialization == SpecializationNone {
		return s.calculateBaseClassBonus(progression.Class)
	}

	// Specialization-based bonuses
	specBonus := s.getSpecializationBonus(progression.Specialization)

	// Genre modifier
	genreMultiplier := s.getGenreMultiplier()

	return specBonus * genreMultiplier
}

// calculateBaseClassBonus provides small bonuses for pure caster classes.
func (s *SpecializationManaBoostSystem) calculateBaseClassBonus(class CharacterClass) float64 {
	switch class {
	case ClassMage:
		return 0.10 // +10% for mages
	case ClassNecromancer:
		return 0.08 // +8% for necromancers
	case ClassCleric:
		return 0.05 // +5% for clerics
	case ClassMystic, ClassWarlock:
		return 0.06 // +6% for hybrid casters
	default:
		return 0.0
	}
}

// getSpecializationBonus returns the base mana regen bonus for a specialization.
func (s *SpecializationManaBoostSystem) getSpecializationBonus(spec SpecializationType) float64 {
	switch spec {
	// Pure magic specializations - highest bonuses
	case SpecializationArcanist:
		return 0.50 // +50% - pure arcane mastery
	case SpecializationElementalist:
		return 0.40 // +40% - elemental focus
	case SpecializationBloodMage:
		return 0.35 // +35% - life-mana exchange
	case SpecializationHealer:
		return 0.30 // +30% - holy mana flow
	case SpecializationDeathKnight:
		return 0.25 // +25% - dark energy
	case SpecializationOracle:
		return 0.45 // +45% - divine insight
	case SpecializationTheurgist:
		return 0.40 // +40% - holy-arcane fusion
	case SpecializationDemonologist:
		return 0.35 // +35% - demonic pacts
	case SpecializationAffliction:
		return 0.30 // +30% - curse mastery
	case SpecializationVoidcaller:
		return 0.38 // +38% - shadow magic
	case SpecializationSoulweaver:
		return 0.32 // +32% - spirit magic
	case SpecializationNaturemage:
		return 0.35 // +35% - nature magic
	case SpecializationHemomancer:
		return 0.33 // +33% - blood magic

	// Hybrid magic specializations - moderate bonuses
	case SpecializationWarmage:
		return 0.25 // +25% - magic-focused battlemage
	case SpecializationSpellsword:
		return 0.15 // +15% - melee-focused battlemage
	case SpecializationTrickster:
		return 0.20 // +20% - illusion focus
	case SpecializationDuelist:
		return 0.10 // +10% - combat focus
	case SpecializationTemplar:
		return 0.18 // +18% - holy warrior
	case SpecializationSpellshot:
		return 0.22 // +22% - magic arrows
	case SpecializationSeeker:
		return 0.15 // +15% - homing magic
	case SpecializationCrusader:
		return 0.12 // +12% - offensive paladin
	case SpecializationGuardian:
		return 0.10 // +10% - defensive paladin
	case SpecializationExorcist:
		return 0.20 // +20% - demon hunting magic
	case SpecializationPurifier:
		return 0.18 // +18% - undead hunting magic
	case SpecializationJudge:
		return 0.15 // +15% - divine judgment
	case SpecializationInterrogator:
		return 0.12 // +12% - mind magic
	case SpecializationCrimsonBlade:
		return 0.15 // +15% - blood weapons
	case SpecializationUnholy:
		return 0.20 // +20% - disease magic
	case SpecializationFrost:
		return 0.18 // +18% - frost magic

	// Physical specializations - no mana bonus
	case SpecializationBerserker, SpecializationDefender:
		return 0.0 // Warriors don't use mana
	case SpecializationAssassin, SpecializationShadowdancer:
		return 0.0 // Rogues use energy/combo points
	case SpecializationBeastmaster, SpecializationMarksman:
		return 0.0 // Rangers use focus
	case SpecializationWindwalker, SpecializationBrewmaster:
		return 0.0 // Monks use chi
	case SpecializationFeralRage, SpecializationPackLeader:
		return 0.0 // Beastlords use rage
	case SpecializationShapeshifter:
		return 0.0 // Druids in form don't regen mana
	case SpecializationShinobi, SpecializationStriker:
		return 0.0 // Ninjas use energy

	default:
		return 0.05 // Small default bonus for unknown specializations
	}
}

// getGenreMultiplier returns a multiplier based on genre.
func (s *SpecializationManaBoostSystem) getGenreMultiplier() float64 {
	switch s.genreID {
	case "fantasy":
		return 1.0 // Standard magic world
	case "scifi":
		return 0.8 // Tech replaces some magic
	case "horror":
		return 1.2 // Dark magic is stronger
	case "cyberpunk":
		return 0.7 // Tech-dominant world
	case "postapoc":
		return 0.9 // Magic is fading
	default:
		return 1.0
	}
}

// removeBonus removes any stored bonus for an entity.
func (s *SpecializationManaBoostSystem) removeBonus(entityID uint64) {
	if _, exists := s.appliedBonuses[entityID]; exists {
		delete(s.appliedBonuses, entityID)
		delete(s.originalRegen, entityID)
	}
}

// GetBonusForEntity returns the current mana regen bonus for an entity (for UI).
func (s *SpecializationManaBoostSystem) GetBonusForEntity(entityID uint64) float64 {
	if bonus, exists := s.appliedBonuses[entityID]; exists {
		return bonus
	}
	return 0.0
}
