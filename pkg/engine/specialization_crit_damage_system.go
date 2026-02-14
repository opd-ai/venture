// Package engine provides the SpecializationCritDamageSystem which bridges class
// specializations with critical hit damage multipliers. This system grants passive
// crit damage bonuses to specialized classes focused on precision and burst damage,
// creating meaningful gameplay incentives for specialization choices.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// SpecializationCritDamageSystem provides critical hit damage bonuses based on
// character class specialization. Assassin and rogue-focused specializations receive
// higher bonuses (20-50% increased crit damage), hybrid specializations receive moderate
// bonuses (10-20%), and tank/healer specializations receive minimal or no bonus.
type SpecializationCritDamageSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// updateInterval controls how often we check specializations (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Cache crit damage bonuses per entity for modifier lookup
	appliedBonuses map[uint64]float64

	// Genre affects bonus amounts (cyberpunk precision tech vs fantasy stealth)
	genreID string
}

// NewSpecializationCritDamageSystem creates a new specialization crit damage system.
func NewSpecializationCritDamageSystem(world *World, seed int64) *SpecializationCritDamageSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "specialization_crit_damage")
		logEntry.Debug("SpecializationCritDamageSystem created")
	}

	return &SpecializationCritDamageSystem{
		world:          world,
		logger:         logEntry,
		rng:            rand.New(rand.NewSource(seed)),
		updateInterval: 1.0, // Check once per second (specializations rarely change)
		appliedBonuses: make(map[uint64]float64),
		genreID:        "fantasy",
	}
}

// SetGenre sets the genre for genre-aware bonus calculations.
func (s *SpecializationCritDamageSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("Genre set for specialization crit damage")
	}
}

// Update checks class specializations and applies crit damage bonuses.
func (s *SpecializationCritDamageSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime

	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	for _, entity := range entities {
		s.processEntity(entity)
	}
}

// processEntity checks if an entity has a specialization and applies crit damage bonus.
func (s *SpecializationCritDamageSystem) processEntity(entity *Entity) {
	// Need class_progression and stats components
	if !entity.HasComponent("class_progression") || !entity.HasComponent("stats") {
		s.removeBonus(entity.ID)
		return
	}

	classComp, _ := entity.GetComponent("class_progression")
	progression, ok := classComp.(*ClassProgressionComponent)
	if !ok {
		return
	}

	statsComp, _ := entity.GetComponent("stats")
	stats, ok := statsComp.(*StatsComponent)
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

	// Apply crit damage bonus: adds to CritDamage multiplier
	// Base CritDamage is typically 2.0 (200%), bonus adds to that
	if hasBonus && currentBonus > 0 {
		// Remove previous bonus
		stats.CritDamage -= currentBonus
	}
	if bonus > 0 {
		// Apply new bonus
		stats.CritDamage += bonus
	}

	// Cache new bonus
	s.appliedBonuses[entity.ID] = bonus

	if s.logger != nil && bonus > 0 {
		s.logger.WithFields(logrus.Fields{
			"entity_id":         entity.ID,
			"class":             progression.Class.String(),
			"specialization":    progression.Specialization.String(),
			"crit_damage_bonus": bonus * 100,
			"total_crit_damage": stats.CritDamage * 100,
		}).Debug("Specialization crit damage bonus applied")
	}
}

// calculateBonus computes the crit damage bonus for a specialization.
// Returns 0.0-0.50 (0-50% crit damage multiplier addition) based on specialization and genre.
func (s *SpecializationCritDamageSystem) calculateBonus(progression *ClassProgressionComponent) float64 {
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

// calculateBaseClassBonus provides small bonuses for precision base classes.
func (s *SpecializationCritDamageSystem) calculateBaseClassBonus(class CharacterClass) float64 {
	switch class {
	case ClassRogue:
		return 0.10 // +10% for rogues (precision damage)
	case ClassNinja:
		return 0.12 // +12% for ninjas (deadly strikes)
	case ClassRanger:
		return 0.08 // +8% for rangers (aimed shots)
	case ClassMonk:
		return 0.05 // +5% for monks (vital strikes)
	case ClassWarrior:
		return 0.03 // +3% for warriors
	default:
		return 0.0
	}
}

// getSpecializationBonus returns the base crit damage bonus for a specialization.
func (s *SpecializationCritDamageSystem) getSpecializationBonus(spec SpecializationType) float64 {
	switch spec {
	// Assassin/precision specializations - highest bonuses
	case SpecializationAssassin:
		return 0.50 // +50% - lethal precision
	case SpecializationShadowdancer:
		return 0.45 // +45% - death from shadows
	case SpecializationShinobi:
		return 0.48 // +48% - ninja death strikes
	case SpecializationMarksman:
		return 0.40 // +40% - precision shots
	case SpecializationDuelist:
		return 0.35 // +35% - precise swordplay
	case SpecializationTrickster:
		return 0.38 // +38% - exploit weaknesses
	case SpecializationCrimsonBlade:
		return 0.42 // +42% - bleeding wounds
	case SpecializationStriker:
		return 0.36 // +36% - vital throwing

	// Berserker/damage specializations - high bonuses
	case SpecializationBerserker:
		return 0.30 // +30% - brutal strikes
	case SpecializationFeralRage:
		return 0.32 // +32% - savage crits
	case SpecializationWindwalker:
		return 0.25 // +25% - swift strikes

	// Hybrid melee-magic specializations - moderate bonuses
	case SpecializationSpellsword:
		return 0.20 // +20% - enchanted strikes
	case SpecializationDeathKnight:
		return 0.22 // +22% - soul rending
	case SpecializationFrost:
		return 0.18 // +18% - frozen precision
	case SpecializationSpellshot:
		return 0.25 // +25% - enchanted arrows
	case SpecializationSeeker:
		return 0.20 // +20% - guided strikes
	case SpecializationShapeshifter:
		return 0.28 // +28% - predator instincts
	case SpecializationUnholy:
		return 0.15 // +15% - disease weakness exploitation

	// Paladin/warrior specializations - low bonuses
	case SpecializationTemplar:
		return 0.12 // +12% - righteous strikes
	case SpecializationCrusader:
		return 0.15 // +15% - zealous attacks
	case SpecializationExorcist:
		return 0.18 // +18% - demon bane crits
	case SpecializationPurifier:
		return 0.16 // +16% - holy fire crits
	case SpecializationJudge:
		return 0.14 // +14% - divine judgment

	// Tank specializations - minimal bonus
	case SpecializationDefender:
		return 0.05 // +5% - defensive focus
	case SpecializationGuardian:
		return 0.06 // +6% - protective warrior
	case SpecializationBrewmaster:
		return 0.08 // +8% - drunken precision

	// Pet/support specializations - low bonus
	case SpecializationBeastmaster:
		return 0.10 // +10% - beast synergy
	case SpecializationPackLeader:
		return 0.08 // +8% - pack tactics

	// Caster specializations - minimal or no bonus
	case SpecializationArcanist, SpecializationElementalist:
		return 0.05 // +5% - spell crits
	case SpecializationWarmage:
		return 0.12 // +12% - combat mage
	case SpecializationBloodMage:
		return 0.10 // +10% - blood sacrifice
	case SpecializationHemomancer:
		return 0.15 // +15% - hemorrhage
	case SpecializationOracle, SpecializationTheurgist:
		return 0.0 // Healers don't crit
	case SpecializationDemonologist:
		return 0.08 // +8% - demon pacts
	case SpecializationAffliction:
		return 0.12 // +12% - curse amplification
	case SpecializationVoidcaller:
		return 0.10 // +10% - void strikes
	case SpecializationSoulweaver:
		return 0.08 // +8% - soul harvest
	case SpecializationNaturemage:
		return 0.05 // +5% - natural precision
	case SpecializationHealer:
		return 0.0 // Pure healer
	case SpecializationInterrogator:
		return 0.18 // +18% - exploit weaknesses found

	default:
		return 0.05 // Small default bonus for unknown specializations
	}
}

// getGenreMultiplier returns a multiplier based on genre.
func (s *SpecializationCritDamageSystem) getGenreMultiplier() float64 {
	switch s.genreID {
	case "fantasy":
		return 1.0 // Standard precision
	case "scifi":
		return 1.15 // Targeting systems enhance crits
	case "horror":
		return 1.2 // Brutal and visceral combat
	case "cyberpunk":
		return 1.25 // Wetware targeting enhancement
	case "postapoc":
		return 1.1 // Survivalist precision
	default:
		return 1.0
	}
}

// removeBonus removes any stored bonus for an entity and restores original crit damage.
func (s *SpecializationCritDamageSystem) removeBonus(entityID uint64) {
	if bonus, exists := s.appliedBonuses[entityID]; exists {
		// Try to restore original crit damage if entity still exists
		if s.world != nil {
			if entity, found := s.world.GetEntity(entityID); found && entity != nil {
				if statsComp, ok := entity.GetComponent("stats"); ok {
					if stats, ok := statsComp.(*StatsComponent); ok && bonus > 0 {
						stats.CritDamage -= bonus
					}
				}
			}
		}
		delete(s.appliedBonuses, entityID)
	}
}

// GetCritDamageBonus returns the crit damage bonus for an entity.
// Returns the flat bonus added to CritDamage multiplier (e.g., 0.30 for +30%).
func (s *SpecializationCritDamageSystem) GetCritDamageBonus(entityID uint64) float64 {
	if bonus, exists := s.appliedBonuses[entityID]; exists {
		return bonus
	}
	return 0.0
}

// GetBonusForEntity returns the current crit damage bonus for an entity (for UI).
// Returns the bonus percentage (0.0-0.50).
func (s *SpecializationCritDamageSystem) GetBonusForEntity(entityID uint64) float64 {
	if bonus, exists := s.appliedBonuses[entityID]; exists {
		return bonus
	}
	return 0.0
}

// IsSpecializationActive returns whether an entity has a crit damage bonus applied.
func (s *SpecializationCritDamageSystem) IsSpecializationActive(entityID uint64) bool {
	_, exists := s.appliedBonuses[entityID]
	return exists
}
