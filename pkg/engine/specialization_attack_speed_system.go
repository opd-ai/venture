// Package engine provides the SpecializationAttackSpeedSystem which bridges class
// specializations with attack speed (cooldown reduction). This system grants passive
// attack speed bonuses to specialized melee and hybrid classes, creating meaningful
// gameplay incentives for specialization choices.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// SpecializationAttackSpeedSystem provides attack speed bonuses based on
// character class specialization. Physical-focused specializations receive higher
// bonuses (15-30% cooldown reduction), hybrid specializations receive moderate
// bonuses (5-15%), and pure caster specializations receive minimal or no bonus.
type SpecializationAttackSpeedSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// updateInterval controls how often we check specializations (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Cache cooldown multipliers per entity for modifier lookup
	appliedBonuses map[uint64]float64

	// Genre affects bonus amounts (cyberpunk tech-enhanced vs fantasy warriors)
	genreID string
}

// NewSpecializationAttackSpeedSystem creates a new specialization attack speed system.
func NewSpecializationAttackSpeedSystem(world *World, seed int64) *SpecializationAttackSpeedSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "specialization_attack_speed")
		logEntry.Debug("SpecializationAttackSpeedSystem created")
	}

	return &SpecializationAttackSpeedSystem{
		world:          world,
		logger:         logEntry,
		rng:            rand.New(rand.NewSource(seed)),
		updateInterval: 1.0, // Check once per second (specializations rarely change)
		appliedBonuses: make(map[uint64]float64),
		genreID:        "fantasy",
	}
}

// SetGenre sets the genre for genre-aware bonus calculations.
func (s *SpecializationAttackSpeedSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("Genre set for specialization attack speed")
	}
}

// Update checks class specializations and applies attack speed bonuses.
func (s *SpecializationAttackSpeedSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime

	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	for _, entity := range entities {
		s.processEntity(entity)
	}
}

// processEntity checks if an entity has a specialization and applies attack speed bonus.
func (s *SpecializationAttackSpeedSystem) processEntity(entity *Entity) {
	// Need class_progression and attack components
	if !entity.HasComponent("class_progression") || !entity.HasComponent("attack") {
		s.removeBonus(entity.ID)
		return
	}

	classComp, _ := entity.GetComponent("class_progression")
	progression, ok := classComp.(*ClassProgressionComponent)
	if !ok {
		return
	}

	attackComp, _ := entity.GetComponent("attack")
	attack, ok := attackComp.(*AttackComponent)
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

	// Apply cooldown reduction: lower Cooldown = faster attacks
	// Bonus of 0.2 (20%) reduces 1.0s cooldown to 0.8s
	if attack.Cooldown > 0 {
		// Restore original cooldown if we had a previous bonus
		if hasBonus && currentBonus > 0 {
			attack.Cooldown = attack.Cooldown / (1.0 - currentBonus)
		}
		// Apply new bonus
		if bonus > 0 {
			attack.Cooldown = attack.Cooldown * (1.0 - bonus)
		}
	}

	// Cache new bonus
	s.appliedBonuses[entity.ID] = bonus

	if s.logger != nil && bonus > 0 {
		s.logger.WithFields(logrus.Fields{
			"entity_id":          entity.ID,
			"class":              progression.Class.String(),
			"specialization":     progression.Specialization.String(),
			"cooldown_reduction": bonus * 100,
		}).Debug("Specialization attack speed bonus applied")
	}
}

// calculateBonus computes the attack speed bonus (cooldown reduction) for a specialization.
// Returns 0.0-0.30 (0-30% cooldown reduction) based on specialization and genre.
func (s *SpecializationAttackSpeedSystem) calculateBonus(progression *ClassProgressionComponent) float64 {
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

// calculateBaseClassBonus provides small bonuses for physical base classes.
func (s *SpecializationAttackSpeedSystem) calculateBaseClassBonus(class CharacterClass) float64 {
	switch class {
	case ClassWarrior:
		return 0.05 // +5% for warriors
	case ClassRogue:
		return 0.08 // +8% for rogues (fast attackers)
	case ClassRanger:
		return 0.04 // +4% for rangers
	case ClassMonk:
		return 0.06 // +6% for monks
	case ClassNinja:
		return 0.07 // +7% for ninjas
	default:
		return 0.0
	}
}

// getSpecializationBonus returns the base attack speed bonus for a specialization.
func (s *SpecializationAttackSpeedSystem) getSpecializationBonus(spec SpecializationType) float64 {
	switch spec {
	// Physical damage specializations - highest bonuses
	case SpecializationBerserker:
		return 0.30 // +30% - frenzied attacks
	case SpecializationAssassin:
		return 0.28 // +28% - quick precise strikes
	case SpecializationShadowdancer:
		return 0.25 // +25% - evasive quick attacks
	case SpecializationWindwalker:
		return 0.30 // +30% - speed focused monk
	case SpecializationDuelist:
		return 0.26 // +26% - combat-focused spellblade
	case SpecializationMarksman:
		return 0.22 // +22% - rapid fire
	case SpecializationFeralRage:
		return 0.28 // +28% - savage melee
	case SpecializationShinobi:
		return 0.27 // +27% - silent quick kills
	case SpecializationStriker:
		return 0.25 // +25% - ranged thrown weapons
	case SpecializationCrimsonBlade:
		return 0.24 // +24% - blood weapon master

	// Tank specializations - moderate attack speed
	case SpecializationDefender:
		return 0.10 // +10% - defensive focus
	case SpecializationBrewmaster:
		return 0.12 // +12% - tank monk
	case SpecializationGuardian:
		return 0.08 // +8% - defensive paladin

	// Hybrid melee-magic specializations - moderate bonuses
	case SpecializationSpellsword:
		return 0.18 // +18% - melee battlemage
	case SpecializationTemplar:
		return 0.15 // +15% - holy warrior
	case SpecializationCrusader:
		return 0.16 // +16% - offensive paladin
	case SpecializationDeathKnight:
		return 0.14 // +14% - dark warrior
	case SpecializationFrost:
		return 0.16 // +16% - frost dual-wield
	case SpecializationUnholy:
		return 0.12 // +12% - disease focus
	case SpecializationTrickster:
		return 0.15 // +15% - illusion-focused
	case SpecializationPackLeader:
		return 0.10 // +10% - command focus
	case SpecializationExorcist:
		return 0.12 // +12% - demon hunter
	case SpecializationPurifier:
		return 0.12 // +12% - undead hunter
	case SpecializationJudge:
		return 0.14 // +14% - divine judgment
	case SpecializationInterrogator:
		return 0.10 // +10% - investigation focus
	case SpecializationShapeshifter:
		return 0.18 // +18% - animal forms attack fast
	case SpecializationBeastmaster:
		return 0.08 // +8% - pet focus

	// Ranged/spell hybrid specializations - low attack speed bonus
	case SpecializationSpellshot:
		return 0.15 // +15% - magic arrows
	case SpecializationSeeker:
		return 0.12 // +12% - homing projectiles

	// Pure caster specializations - minimal or no bonus
	case SpecializationArcanist, SpecializationElementalist:
		return 0.0 // Casters don't benefit from attack speed
	case SpecializationWarmage:
		return 0.05 // +5% - magic-focused battlemage
	case SpecializationBloodMage, SpecializationHemomancer:
		return 0.0 // Blood casters
	case SpecializationOracle, SpecializationTheurgist:
		return 0.0 // Divine casters
	case SpecializationDemonologist, SpecializationAffliction:
		return 0.0 // Warlocks
	case SpecializationVoidcaller, SpecializationSoulweaver:
		return 0.0 // Shadow casters
	case SpecializationNaturemage:
		return 0.0 // Nature caster
	case SpecializationHealer:
		return 0.0 // Healing focus

	default:
		return 0.02 // Small default bonus for unknown specializations
	}
}

// getGenreMultiplier returns a multiplier based on genre.
func (s *SpecializationAttackSpeedSystem) getGenreMultiplier() float64 {
	switch s.genreID {
	case "fantasy":
		return 1.0 // Standard martial prowess
	case "scifi":
		return 1.1 // Enhanced reflexes tech
	case "horror":
		return 0.9 // Survival horror pace
	case "cyberpunk":
		return 1.15 // Cybernetic enhancements
	case "postapoc":
		return 0.95 // Scrappy survivors
	default:
		return 1.0
	}
}

// removeBonus removes any stored bonus for an entity and restores original cooldown.
func (s *SpecializationAttackSpeedSystem) removeBonus(entityID uint64) {
	if bonus, exists := s.appliedBonuses[entityID]; exists {
		// Try to restore original cooldown if entity still exists
		if s.world != nil {
			if entity, found := s.world.GetEntity(entityID); found && entity != nil {
				if attackComp, ok := entity.GetComponent("attack"); ok {
					if attack, ok := attackComp.(*AttackComponent); ok && bonus > 0 {
						attack.Cooldown = attack.Cooldown / (1.0 - bonus)
					}
				}
			}
		}
		delete(s.appliedBonuses, entityID)
	}
}

// GetCooldownMultiplier returns the cooldown multiplier for an entity.
// Returns 1.0 - bonus (e.g., 0.75 for a 25% cooldown reduction).
func (s *SpecializationAttackSpeedSystem) GetCooldownMultiplier(entityID uint64) float64 {
	if bonus, exists := s.appliedBonuses[entityID]; exists {
		return 1.0 - bonus
	}
	return 1.0
}

// GetBonusForEntity returns the current attack speed bonus for an entity (for UI).
// Returns the bonus percentage (0.0-0.30) not the multiplier.
func (s *SpecializationAttackSpeedSystem) GetBonusForEntity(entityID uint64) float64 {
	if bonus, exists := s.appliedBonuses[entityID]; exists {
		return bonus
	}
	return 0.0
}

// IsSpecializationActive returns whether an entity has an attack speed bonus applied.
func (s *SpecializationAttackSpeedSystem) IsSpecializationActive(entityID uint64) bool {
	_, exists := s.appliedBonuses[entityID]
	return exists
}
