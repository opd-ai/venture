// Package engine provides the SpecializationSpellDamageSystem which bridges class
// specializations with spell damage output. This system grants passive spell damage
// bonuses to specialized mages and caster classes, creating meaningful gameplay
// incentives for specialization choices.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// SpecializationSpellDamageSystem provides spell damage bonuses based on
// character class specialization. Magic-focused specializations receive higher
// bonuses (15-40%), hybrid specializations receive moderate bonuses (5-15%),
// and melee specializations receive no bonus.
type SpecializationSpellDamageSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// updateInterval controls how often we check specializations (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Cache damage bonuses per entity for UI display and modifier lookup
	appliedBonuses map[uint64]float64

	// Genre affects bonus amounts (scifi/cyberpunk tech-casters vs fantasy mages)
	genreID string
}

// NewSpecializationSpellDamageSystem creates a new specialization spell damage system.
func NewSpecializationSpellDamageSystem(world *World, seed int64) *SpecializationSpellDamageSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "specialization_spell_damage")
		logEntry.Debug("SpecializationSpellDamageSystem created")
	}

	return &SpecializationSpellDamageSystem{
		world:          world,
		logger:         logEntry,
		rng:            rand.New(rand.NewSource(seed)),
		updateInterval: 1.0, // Check once per second (specializations rarely change)
		appliedBonuses: make(map[uint64]float64),
		genreID:        "fantasy",
	}
}

// SetGenre sets the genre for genre-aware bonus calculations.
func (s *SpecializationSpellDamageSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("Genre set for specialization spell damage")
	}
}

// Update checks class specializations and caches spell damage bonuses.
func (s *SpecializationSpellDamageSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime

	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	for _, entity := range entities {
		s.processEntity(entity)
	}
}

// processEntity checks if an entity has a specialization and caches spell damage bonus.
func (s *SpecializationSpellDamageSystem) processEntity(entity *Entity) {
	// Need class_progression component for specialization
	if !entity.HasComponent("class_progression") {
		// If entity lost component, remove any stored bonus
		s.removeBonus(entity.ID)
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

	// Cache new bonus
	s.appliedBonuses[entity.ID] = bonus

	if s.logger != nil && bonus > 0 {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entity.ID,
			"class":          progression.Class.String(),
			"specialization": progression.Specialization.String(),
			"bonus_percent":  bonus * 100,
		}).Debug("Specialization spell damage bonus cached")
	}
}

// calculateBonus computes the spell damage bonus multiplier for a specialization.
// Returns 0.0-0.4 (0-40% bonus) based on specialization and genre.
func (s *SpecializationSpellDamageSystem) calculateBonus(progression *ClassProgressionComponent) float64 {
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
func (s *SpecializationSpellDamageSystem) calculateBaseClassBonus(class CharacterClass) float64 {
	switch class {
	case ClassMage:
		return 0.08 // +8% for mages
	case ClassNecromancer:
		return 0.07 // +7% for necromancers
	case ClassCleric:
		return 0.04 // +4% for clerics (focus on healing)
	case ClassMystic, ClassWarlock:
		return 0.05 // +5% for hybrid casters
	default:
		return 0.0
	}
}

// getSpecializationBonus returns the base spell damage bonus for a specialization.
func (s *SpecializationSpellDamageSystem) getSpecializationBonus(spec SpecializationType) float64 {
	switch spec {
	// Pure magic specializations - highest bonuses
	case SpecializationArcanist:
		return 0.40 // +40% - pure arcane mastery
	case SpecializationElementalist:
		return 0.35 // +35% - elemental focus
	case SpecializationBloodMage:
		return 0.30 // +30% - life-mana exchange
	case SpecializationOracle:
		return 0.25 // +25% - divine insight
	case SpecializationTheurgist:
		return 0.32 // +32% - holy-arcane fusion
	case SpecializationDemonologist:
		return 0.35 // +35% - demonic pacts
	case SpecializationAffliction:
		return 0.28 // +28% - curse mastery
	case SpecializationVoidcaller:
		return 0.33 // +33% - shadow magic
	case SpecializationSoulweaver:
		return 0.26 // +26% - spirit magic
	case SpecializationNaturemage:
		return 0.28 // +28% - nature magic
	case SpecializationHemomancer:
		return 0.30 // +30% - blood magic

	// Hybrid magic specializations - moderate bonuses
	case SpecializationWarmage:
		return 0.22 // +22% - magic-focused battlemage
	case SpecializationSpellsword:
		return 0.12 // +12% - melee-focused battlemage
	case SpecializationTrickster:
		return 0.18 // +18% - illusion focus
	case SpecializationDuelist:
		return 0.08 // +8% - combat focus
	case SpecializationTemplar:
		return 0.15 // +15% - holy warrior
	case SpecializationSpellshot:
		return 0.20 // +20% - magic arrows
	case SpecializationSeeker:
		return 0.16 // +16% - homing magic
	case SpecializationCrusader:
		return 0.10 // +10% - offensive paladin
	case SpecializationGuardian:
		return 0.06 // +6% - defensive paladin
	case SpecializationExorcist:
		return 0.18 // +18% - demon hunting magic
	case SpecializationPurifier:
		return 0.16 // +16% - undead hunting magic
	case SpecializationUnholy:
		return 0.18 // +18% - disease magic
	case SpecializationFrost:
		return 0.16 // +16% - frost magic
	case SpecializationDeathKnight:
		return 0.14 // +14% - dark energy

	// Healing specializations - low damage bonus
	case SpecializationHealer:
		return 0.05 // +5% - focus is on healing

	// Physical specializations - no spell damage bonus
	case SpecializationBerserker, SpecializationDefender:
		return 0.0 // Warriors don't cast spells
	case SpecializationAssassin, SpecializationShadowdancer:
		return 0.0 // Rogues don't cast spells
	case SpecializationBeastmaster, SpecializationMarksman:
		return 0.0 // Rangers don't cast spells
	case SpecializationWindwalker, SpecializationBrewmaster:
		return 0.0 // Monks don't cast spells
	case SpecializationFeralRage, SpecializationPackLeader:
		return 0.0 // Beastlords don't cast spells
	case SpecializationShapeshifter:
		return 0.0 // Druids in form don't cast

	default:
		return 0.03 // Small default bonus for unknown specializations
	}
}

// getGenreMultiplier returns a multiplier based on genre.
func (s *SpecializationSpellDamageSystem) getGenreMultiplier() float64 {
	switch s.genreID {
	case "fantasy":
		return 1.0 // Standard magic world
	case "scifi":
		return 0.85 // Tech replaces some magic
	case "horror":
		return 1.15 // Dark magic is stronger
	case "cyberpunk":
		return 0.75 // Tech-dominant world
	case "postapoc":
		return 0.95 // Magic is fading
	default:
		return 1.0
	}
}

// removeBonus removes any stored bonus for an entity.
func (s *SpecializationSpellDamageSystem) removeBonus(entityID uint64) {
	delete(s.appliedBonuses, entityID)
}

// GetDamageMultiplier returns the spell damage multiplier for an entity.
// Returns 1.0 + bonus (e.g., 1.35 for a 35% bonus).
// This is called by SpellCastingSystem when calculating spell damage.
func (s *SpecializationSpellDamageSystem) GetDamageMultiplier(entityID uint64) float64 {
	if bonus, exists := s.appliedBonuses[entityID]; exists {
		return 1.0 + bonus
	}
	return 1.0
}

// GetBonusForEntity returns the current spell damage bonus for an entity (for UI).
// Returns the bonus percentage (0.0-0.4) not the multiplier.
func (s *SpecializationSpellDamageSystem) GetBonusForEntity(entityID uint64) float64 {
	if bonus, exists := s.appliedBonuses[entityID]; exists {
		return bonus
	}
	return 0.0
}

// IsSpecializationActive returns whether an entity has a bonus applied.
func (s *SpecializationSpellDamageSystem) IsSpecializationActive(entityID uint64) bool {
	_, exists := s.appliedBonuses[entityID]
	return exists
}
