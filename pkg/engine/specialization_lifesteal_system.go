// Package engine provides the SpecializationLifestealSystem which bridges class
// specializations with lifesteal percentage bonuses. This system grants passive
// lifesteal bonuses to specialized classes focused on blood magic and survival,
// creating meaningful gameplay incentives for sustain-focused specialization choices.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// SpecializationLifestealSystem provides lifesteal bonuses based on character
// class specialization. Blood knight and vampire-themed specializations receive
// higher bonuses (15-30% lifesteal), hybrid melee specializations receive moderate
// bonuses (5-10%), and pure caster/tank specializations receive minimal or no bonus.
type SpecializationLifestealSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// updateInterval controls how often we check specializations (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Cache lifesteal bonuses per entity for modifier lookup
	appliedBonuses map[uint64]float64

	// Genre affects bonus amounts (horror blood magic vs scifi nanobots)
	genreID string
}

// NewSpecializationLifestealSystem creates a new specialization lifesteal system.
func NewSpecializationLifestealSystem(world *World, seed int64) *SpecializationLifestealSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "specialization_lifesteal")
		logEntry.Debug("SpecializationLifestealSystem created")
	}

	return &SpecializationLifestealSystem{
		world:          world,
		logger:         logEntry,
		rng:            rand.New(rand.NewSource(seed)),
		updateInterval: 1.0, // Check once per second (specializations rarely change)
		appliedBonuses: make(map[uint64]float64),
		genreID:        "fantasy",
	}
}

// SetGenre sets the genre for genre-aware bonus calculations.
func (s *SpecializationLifestealSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("Genre set for specialization lifesteal")
	}
}

// Update checks class specializations and applies lifesteal bonuses.
func (s *SpecializationLifestealSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime

	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	for _, entity := range entities {
		s.processEntity(entity)
	}
}

// processEntity checks if an entity has a specialization and applies lifesteal bonus.
func (s *SpecializationLifestealSystem) processEntity(entity *Entity) {
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

	// Apply lifesteal bonus: adds to Lifesteal percentage
	// Base Lifesteal is typically 0.0, bonus adds percentage (e.g., 0.15 = 15%)
	if hasBonus && currentBonus > 0 {
		// Remove previous bonus
		stats.Lifesteal -= currentBonus
	}
	if bonus > 0 {
		// Apply new bonus
		stats.Lifesteal += bonus
	}

	// Cache new bonus
	s.appliedBonuses[entity.ID] = bonus

	if s.logger != nil && bonus > 0 {
		s.logger.WithFields(logrus.Fields{
			"entity_id":       entity.ID,
			"class":           progression.Class.String(),
			"specialization":  progression.Specialization.String(),
			"lifesteal_bonus": bonus * 100,
			"total_lifesteal": stats.Lifesteal * 100,
		}).Debug("Specialization lifesteal bonus applied")
	}
}

// calculateBonus computes the lifesteal bonus for a specialization.
// Returns 0.0-0.30 (0-30% lifesteal) based on specialization and genre.
func (s *SpecializationLifestealSystem) calculateBonus(progression *ClassProgressionComponent) float64 {
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

// calculateBaseClassBonus provides small bonuses for survival-focused base classes.
func (s *SpecializationLifestealSystem) calculateBaseClassBonus(class CharacterClass) float64 {
	switch class {
	case ClassNecromancer:
		return 0.05 // +5% for necromancers (life drain theme)
	case ClassWarrior:
		return 0.02 // +2% for warriors (battle sustain)
	case ClassMonk:
		return 0.03 // +3% for monks (chi balance)
	default:
		return 0.0
	}
}

// getSpecializationBonus returns the base lifesteal bonus for a specialization.
func (s *SpecializationLifestealSystem) getSpecializationBonus(spec SpecializationType) float64 {
	switch spec {
	// Blood magic specializations - highest bonuses
	case SpecializationBloodMage:
		return 0.30 // +30% - blood sacrifice healing
	case SpecializationCrimsonBlade:
		return 0.28 // +28% - bleeding wounds feed
	case SpecializationHemomancer:
		return 0.25 // +25% - blood magic mastery
	case SpecializationSoulweaver:
		return 0.22 // +22% - soul drain healing

	// Death/undead specializations - high bonuses
	case SpecializationDeathKnight:
		return 0.20 // +20% - undead sustain
	case SpecializationUnholy:
		return 0.18 // +18% - disease feeding
	case SpecializationVoidcaller:
		return 0.15 // +15% - void siphon

	// Berserker/feral specializations - moderate-high bonuses
	case SpecializationBerserker:
		return 0.15 // +15% - rage sustain
	case SpecializationFeralRage:
		return 0.18 // +18% - savage feeding
	case SpecializationShapeshifter:
		return 0.12 // +12% - predator instincts

	// Melee combat specializations - moderate bonuses
	case SpecializationSpellsword:
		return 0.10 // +10% - enchanted strikes heal
	case SpecializationDuelist:
		return 0.08 // +8% - precise wounds
	case SpecializationWindwalker:
		return 0.08 // +8% - swift strike sustain
	case SpecializationAssassin:
		return 0.10 // +10% - lethal precision
	case SpecializationShadowdancer:
		return 0.12 // +12% - shadow drain

	// Paladin/holy specializations - low-moderate bonuses
	case SpecializationTemplar:
		return 0.08 // +8% - righteous vigor
	case SpecializationCrusader:
		return 0.10 // +10% - zealous sustain
	case SpecializationPurifier:
		return 0.06 // +6% - purifying strikes
	case SpecializationExorcist:
		return 0.08 // +8% - demon essence drain

	// Warlock specializations - moderate bonuses
	case SpecializationDemonologist:
		return 0.12 // +12% - demon pact sustain
	case SpecializationAffliction:
		return 0.15 // +15% - curse feeding

	// Tank specializations - low bonuses (they have other sustain)
	case SpecializationDefender:
		return 0.05 // +5% - defensive sustain
	case SpecializationGuardian:
		return 0.06 // +6% - protective vigor
	case SpecializationBrewmaster:
		return 0.07 // +7% - drunken recovery

	// Ranged specializations - minimal bonuses
	case SpecializationMarksman:
		return 0.03 // +3% - precision shots
	case SpecializationSpellshot:
		return 0.05 // +5% - enchanted arrows
	case SpecializationSeeker:
		return 0.04 // +4% - guided strikes
	case SpecializationStriker:
		return 0.05 // +5% - thrown precision

	// Pet specializations - low bonuses (pets provide sustain)
	case SpecializationBeastmaster:
		return 0.05 // +5% - beast synergy
	case SpecializationPackLeader:
		return 0.04 // +4% - pack tactics

	// Pure caster specializations - minimal bonuses
	case SpecializationArcanist:
		return 0.02 // +2% - mana efficiency
	case SpecializationElementalist:
		return 0.03 // +3% - elemental attunement
	case SpecializationWarmage:
		return 0.06 // +6% - combat mage sustain
	case SpecializationFrost:
		return 0.04 // +4% - frost siphon
	case SpecializationNaturemage:
		return 0.05 // +5% - natural restoration

	// Healer specializations - no bonus (they heal directly)
	case SpecializationHealer:
		return 0.0 // Pure healer
	case SpecializationOracle:
		return 0.0 // Divination focus
	case SpecializationTheurgist:
		return 0.02 // +2% - divine arcane

	// Ninja specializations
	case SpecializationShinobi:
		return 0.08 // +8% - death strike sustain
	case SpecializationTrickster:
		return 0.06 // +6% - illusory sustain

	// Inquisitor specializations
	case SpecializationJudge:
		return 0.08 // +8% - divine judgment
	case SpecializationInterrogator:
		return 0.10 // +10% - extract essence

	default:
		return 0.03 // Small default bonus for unknown specializations
	}
}

// getGenreMultiplier returns a multiplier based on genre.
func (s *SpecializationLifestealSystem) getGenreMultiplier() float64 {
	switch s.genreID {
	case "fantasy":
		return 1.0 // Standard blood magic
	case "scifi":
		return 0.8 // Nano-repair less effective than magic
	case "horror":
		return 1.3 // Vampiric themes enhanced
	case "cyberpunk":
		return 0.9 // Biotech healing moderate
	case "postapoc":
		return 1.1 // Survivalist sustain
	default:
		return 1.0
	}
}

// removeBonus removes any stored bonus for an entity and restores original lifesteal.
func (s *SpecializationLifestealSystem) removeBonus(entityID uint64) {
	if bonus, exists := s.appliedBonuses[entityID]; exists {
		// Try to restore original lifesteal if entity still exists
		if s.world != nil {
			if entity, found := s.world.GetEntity(entityID); found && entity != nil {
				if statsComp, ok := entity.GetComponent("stats"); ok {
					if stats, ok := statsComp.(*StatsComponent); ok && bonus > 0 {
						stats.Lifesteal -= bonus
						// Ensure lifesteal doesn't go negative
						if stats.Lifesteal < 0 {
							stats.Lifesteal = 0
						}
					}
				}
			}
		}
		delete(s.appliedBonuses, entityID)
	}
}

// GetLifestealBonus returns the lifesteal bonus for an entity.
// Returns the percentage bonus (e.g., 0.15 for +15%).
func (s *SpecializationLifestealSystem) GetLifestealBonus(entityID uint64) float64 {
	if bonus, exists := s.appliedBonuses[entityID]; exists {
		return bonus
	}
	return 0.0
}

// IsSpecializationActive returns whether an entity has a lifesteal bonus applied.
func (s *SpecializationLifestealSystem) IsSpecializationActive(entityID uint64) bool {
	_, exists := s.appliedBonuses[entityID]
	return exists
}
