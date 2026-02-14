// Package engine provides the ElementalCompanionSynergySystem for boosting
// elemental companions when their owner has matching elemental status effects.
// This connects CompanionComponent (elemental type) with StatusEffectComponent
// to create tactical synergy: fire elementals gain power when owner is burning, etc.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// ElementalCompanionSynergySystem boosts elemental companion combat stats
// when their owner has matching elemental status effects active.
// This creates tactical depth where players can enhance their elemental companions
// by taking or applying elemental effects to themselves.
type ElementalCompanionSynergySystem struct {
	world   *World
	genreID string
	seed    int64
	rng     *rand.Rand
	logger  *logrus.Entry

	// Bonus configuration
	attackBonusPct  float64 // Percentage attack bonus (e.g., 0.25 = 25%)
	defenseBonusPct float64 // Percentage defense bonus
	speedBonusPct   float64 // Percentage speed bonus

	// Genre-specific multipliers
	genreMultipliers map[string]float64

	// Element-to-status mappings
	elementMappings map[CompanionType][]string

	// Track active synergy bonuses to avoid duplicate application
	activeSynergies map[uint64]bool // companionID -> has active synergy
}

// NewElementalCompanionSynergySystem creates a new elemental companion synergy system.
func NewElementalCompanionSynergySystem(world *World, seed int64) *ElementalCompanionSynergySystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "elemental_companion_synergy")
		logEntry.Debug("elemental companion synergy system created")
	}

	return &ElementalCompanionSynergySystem{
		world:           world,
		seed:            seed,
		rng:             rand.New(rand.NewSource(seed)),
		logger:          logEntry,
		attackBonusPct:  0.25, // 25% attack bonus
		defenseBonusPct: 0.15, // 15% defense bonus
		speedBonusPct:   0.10, // 10% speed bonus
		genreMultipliers: map[string]float64{
			"fantasy":   1.0, // Standard magic
			"scifi":     0.8, // Less elemental affinity
			"horror":    1.2, // Elemental forces are stronger
			"cyberpunk": 0.7, // Tech over magic
			"postapoc":  0.9, // Diminished magic
		},
		elementMappings: map[CompanionType][]string{
			CompanionTypeElemental: {"burning", "frozen", "shocked", "poisoned", "wet"},
		},
		activeSynergies: make(map[uint64]bool),
	}
}

// SetGenre sets the genre ID for genre-aware synergy multipliers.
func (s *ElementalCompanionSynergySystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes all elemental companions and applies/removes synergy bonuses
// based on their owner's active status effects.
func (s *ElementalCompanionSynergySystem) Update(entities []*Entity, deltaTime float64) {
	if s.world == nil {
		return
	}

	// Track which companions have synergy this frame
	currentSynergies := make(map[uint64]bool)

	for _, entity := range entities {
		// Check if entity is an elemental companion
		compComp, ok := entity.GetComponent("companion")
		if !ok {
			continue
		}
		companionComp, ok := compComp.(*CompanionComponent)
		if !ok {
			continue
		}

		// Only process elemental companions
		if companionComp.CompanionType != CompanionTypeElemental {
			continue
		}

		// Get owner entity
		owner, exists := s.world.GetEntity(companionComp.OwnerID)
		if !exists || owner == nil {
			continue
		}

		// Check if owner has any matching elemental status effects
		hasSynergy := s.checkOwnerElementalEffects(owner)

		if hasSynergy {
			currentSynergies[entity.ID] = true

			// Apply bonus if not already active
			if !s.activeSynergies[entity.ID] {
				s.applySynergyBonus(entity, companionComp)
				s.activeSynergies[entity.ID] = true
			}
		}
	}

	// Remove bonuses from companions that no longer have synergy
	for companionID := range s.activeSynergies {
		if !currentSynergies[companionID] {
			companion, exists := s.world.GetEntity(companionID)
			if exists && companion != nil {
				s.removeSynergyBonus(companion)
			}
			delete(s.activeSynergies, companionID)
		}
	}
}

// checkOwnerElementalEffects checks if the owner has any elemental status effects.
func (s *ElementalCompanionSynergySystem) checkOwnerElementalEffects(owner *Entity) bool {
	validEffects := s.elementMappings[CompanionTypeElemental]
	if len(validEffects) == 0 {
		return false
	}

	for _, comp := range owner.Components {
		effect, ok := comp.(*StatusEffectComponent)
		if !ok {
			continue
		}

		// Check if this status effect is one of the elemental types
		for _, validType := range validEffects {
			if effect.EffectType == validType && effect.Duration > 0 {
				return true
			}
		}
	}
	return false
}

// applySynergyBonus applies stat bonuses to the companion.
func (s *ElementalCompanionSynergySystem) applySynergyBonus(companion *Entity, companionComp *CompanionComponent) {
	// Get genre multiplier
	genreMult := s.genreMultipliers[s.genreID]
	if genreMult == 0 {
		genreMult = 1.0
	}

	// Try to boost companion stats
	statsComp, ok := companion.GetComponent("companion_stats")
	if !ok {
		// Try regular stats component
		statsComp, ok = companion.GetComponent("stats")
	}

	if ok {
		if compStats, ok := statsComp.(*CompanionStatsComponent); ok {
			compStats.Attack *= (1.0 + s.attackBonusPct*genreMult)
			compStats.Defense *= (1.0 + s.defenseBonusPct*genreMult)
			compStats.Speed *= (1.0 + s.speedBonusPct*genreMult)
		} else if stats, ok := statsComp.(*StatsComponent); ok {
			stats.Attack *= (1.0 + s.attackBonusPct*genreMult)
			stats.Defense *= (1.0 + s.defenseBonusPct*genreMult)
		}
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"companion_id":   companion.ID,
			"owner_id":       companionComp.OwnerID,
			"attack_bonus":   s.attackBonusPct * genreMult,
			"defense_bonus":  s.defenseBonusPct * genreMult,
			"genre_modifier": genreMult,
		}).Debug("elemental synergy bonus applied")
	}
}

// removeSynergyBonus removes stat bonuses from the companion.
func (s *ElementalCompanionSynergySystem) removeSynergyBonus(companion *Entity) {
	// Get genre multiplier for accurate reversal
	genreMult := s.genreMultipliers[s.genreID]
	if genreMult == 0 {
		genreMult = 1.0
	}

	statsComp, ok := companion.GetComponent("companion_stats")
	if !ok {
		statsComp, ok = companion.GetComponent("stats")
	}

	if ok {
		if compStats, ok := statsComp.(*CompanionStatsComponent); ok {
			compStats.Attack /= (1.0 + s.attackBonusPct*genreMult)
			compStats.Defense /= (1.0 + s.defenseBonusPct*genreMult)
			compStats.Speed /= (1.0 + s.speedBonusPct*genreMult)
		} else if stats, ok := statsComp.(*StatsComponent); ok {
			stats.Attack /= (1.0 + s.attackBonusPct*genreMult)
			stats.Defense /= (1.0 + s.defenseBonusPct*genreMult)
		}
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"companion_id": companion.ID,
		}).Debug("elemental synergy bonus removed")
	}
}

// HasActiveSynergy returns whether a companion currently has an active synergy bonus.
func (s *ElementalCompanionSynergySystem) HasActiveSynergy(companionID uint64) bool {
	return s.activeSynergies[companionID]
}

// GetActiveSynergyCount returns the number of companions with active synergy bonuses.
func (s *ElementalCompanionSynergySystem) GetActiveSynergyCount() int {
	return len(s.activeSynergies)
}
