// Package engine provides the SpecializationEvasionSystem which bridges class
// specializations with evasion bonuses. This system grants passive evasion increases
// to agility and stealth-focused specializations, creating meaningful gameplay
// incentives for specialization choices that complement SpecializationDefenseSystem.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// SpecializationEvasionSystem provides evasion bonuses based on character class
// specialization. Agility specializations receive the highest bonuses (15-30%),
// hybrid melee specializations receive moderate bonuses (5-15%), and tank/caster
// specializations receive small or no bonus.
//
// Supported specializations and their evasion bonuses:
//   - Shadowdancer: +30% (pure evasion build)
//   - Windwalker: +25% (speed-focused monk)
//   - Assassin: +20% (stealth and agility)
//   - Trickster: +20% (illusion-based evasion)
//   - Marksman: +15% (mobile ranged combat)
//   - Duelist: +15% (parry and dodge)
//   - Beastmaster: +10% (companion-aided awareness)
//   - Exorcist: +10% (holy reflexes)
//
// Genre modifiers:
//   - Fantasy: Standard evasion rates
//   - Scifi: +10% bonus (enhanced reflexes augmentation)
//   - Horror: -10% penalty (panic reduces coordination)
//   - Cyberpunk: +15% bonus (reflex boosters common)
//   - Postapoc: -5% penalty (malnutrition affects agility)
type SpecializationEvasionSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// updateInterval controls how often we check specializations (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Cache original evasion values per entity to apply bonus correctly
	originalEvasion map[uint64]float64

	// Track which entities have bonuses applied
	appliedBonuses map[uint64]float64

	// Genre affects bonus amounts
	genreID string
}

// NewSpecializationEvasionSystem creates a new specialization evasion system.
func NewSpecializationEvasionSystem(world *World, seed int64) *SpecializationEvasionSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "specialization_evasion")
		logEntry.Debug("SpecializationEvasionSystem created")
	}

	return &SpecializationEvasionSystem{
		world:           world,
		logger:          logEntry,
		rng:             rand.New(rand.NewSource(seed)),
		updateInterval:  1.0, // Check once per second (specializations rarely change)
		originalEvasion: make(map[uint64]float64),
		appliedBonuses:  make(map[uint64]float64),
		genreID:         "fantasy",
	}
}

// SetGenre sets the genre for genre-aware bonus calculations.
func (s *SpecializationEvasionSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("Genre set for specialization evasion")
	}
}

// Update checks class specializations and applies evasion bonuses.
func (s *SpecializationEvasionSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime

	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	for _, entity := range entities {
		s.processEntity(entity)
	}
}

// processEntity checks if an entity has a specialization and applies evasion bonus.
func (s *SpecializationEvasionSystem) processEntity(entity *Entity) {
	// Need both stats and class_progression components
	if !entity.HasComponent("stats") || !entity.HasComponent("class_progression") {
		// If entity lost components, remove any stored bonus
		s.removeBonus(entity.ID)
		return
	}

	statsComp, _ := entity.GetComponent("stats")
	stats, ok := statsComp.(*StatsComponent)
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

	// Store original evasion if not already stored
	if _, exists := s.originalEvasion[entity.ID]; !exists {
		s.originalEvasion[entity.ID] = stats.Evasion
	}

	// Apply new bonus (additive for evasion)
	originalEvasion := s.originalEvasion[entity.ID]
	newEvasion := originalEvasion + bonus
	// Cap evasion at 0.85 (85%) to prevent invincibility
	if newEvasion > 0.85 {
		newEvasion = 0.85
	}
	stats.Evasion = newEvasion
	s.appliedBonuses[entity.ID] = bonus

	if s.logger != nil && bonus > 0 {
		s.logger.WithFields(logrus.Fields{
			"entity_id":        entity.ID,
			"class":            progression.Class.String(),
			"specialization":   progression.Specialization.String(),
			"bonus_percent":    bonus * 100,
			"original_evasion": originalEvasion,
			"new_evasion":      newEvasion,
		}).Debug("Specialization evasion bonus applied")
	}
}

// removeBonus removes any applied bonus for an entity.
func (s *SpecializationEvasionSystem) removeBonus(entityID uint64) {
	delete(s.appliedBonuses, entityID)
	delete(s.originalEvasion, entityID)
}

// calculateBonus computes the evasion bonus for a specialization.
// Returns 0.0-0.30 (0-30% additive bonus) based on specialization and genre.
func (s *SpecializationEvasionSystem) calculateBonus(progression *ClassProgressionComponent) float64 {
	// No bonus without specialization
	if progression.Specialization == SpecializationNone {
		return s.calculateBaseClassBonus(progression.Class)
	}

	// Specialization-based bonuses
	specBonus := s.getSpecializationBonus(progression.Specialization)

	// Genre modifier (multiplicative)
	genreMultiplier := s.getGenreMultiplier()

	return specBonus * genreMultiplier
}

// calculateBaseClassBonus provides small bonuses for agile base classes.
func (s *SpecializationEvasionSystem) calculateBaseClassBonus(class CharacterClass) float64 {
	switch class {
	case ClassRogue:
		return 0.08 // +8% for rogues
	case ClassRanger:
		return 0.05 // +5% for rangers
	case ClassMonk:
		return 0.06 // +6% for monks
	case ClassSpellblade:
		return 0.04 // +4% for spellblades
	default:
		return 0.0
	}
}

// getSpecializationBonus returns the base evasion bonus for a specialization.
func (s *SpecializationEvasionSystem) getSpecializationBonus(spec SpecializationType) float64 {
	switch spec {
	// Pure evasion specializations - highest bonuses
	case SpecializationShadowdancer:
		return 0.30 // +30% - master of evasion
	case SpecializationWindwalker:
		return 0.25 // +25% - speed-based evasion
	case SpecializationAssassin:
		return 0.20 // +20% - agility and instinct
	case SpecializationTrickster:
		return 0.20 // +20% - illusion-aided evasion

	// Hybrid agility specializations - moderate bonuses
	case SpecializationMarksman:
		return 0.15 // +15% - mobile ranged tactics
	case SpecializationDuelist:
		return 0.15 // +15% - parry and riposte
	case SpecializationBeastmaster:
		return 0.10 // +10% - companion warns of danger
	case SpecializationExorcist:
		return 0.10 // +10% - holy instincts
	case SpecializationPurifier:
		return 0.08 // +8% - trained reflexes

	// Light melee specializations - small bonuses
	case SpecializationFeralRage:
		return 0.08 // +8% - primal instincts
	case SpecializationCrimsonBlade:
		return 0.06 // +6% - blood-quickened reflexes
	case SpecializationSpellsword:
		return 0.05 // +5% - enchanted agility
	case SpecializationBloodMage:
		return 0.04 // +4% - life-stealing vitality

	// Tank/caster specializations - no evasion focus
	case SpecializationDefender, SpecializationGuardian, SpecializationBrewmaster:
		return 0.0 // Tanks don't rely on evasion
	case SpecializationElementalist, SpecializationArcanist, SpecializationHealer:
		return 0.0 // Casters don't rely on evasion

	default:
		return 0.0
	}
}

// getGenreMultiplier returns a multiplier based on genre.
func (s *SpecializationEvasionSystem) getGenreMultiplier() float64 {
	switch s.genreID {
	case "scifi":
		return 1.10 // +10% - reflex augmentations
	case "cyberpunk":
		return 1.15 // +15% - reflex boosters
	case "horror":
		return 0.90 // -10% - panic affects coordination
	case "postapoc":
		return 0.95 // -5% - malnutrition
	case "fantasy":
		return 1.0 // Standard
	default:
		return 1.0
	}
}

// GetEvasionBonus returns the current evasion bonus for an entity.
// This can be used by combat system for display or calculations.
func (s *SpecializationEvasionSystem) GetEvasionBonus(entityID uint64) float64 {
	if bonus, exists := s.appliedBonuses[entityID]; exists {
		return bonus
	}
	return 0.0
}
