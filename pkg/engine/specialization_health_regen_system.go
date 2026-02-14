// Package engine provides the SpecializationHealthRegenSystem which bridges class
// specializations with health regeneration. This system grants passive health regen
// bonuses to tank and healer specializations, creating meaningful gameplay incentives
// for specialization choices that complement the SpecializationManaBoostSystem.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// SpecializationHealthRegenSystem provides health regeneration bonuses based on
// character class specialization. Tank specializations receive the highest bonuses
// (20-40%), healer specializations receive moderate bonuses (15-30%), and damage
// specializations receive small or no bonus.
type SpecializationHealthRegenSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// updateInterval controls how often we check specializations (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Cache original health values per entity to compute regen correctly
	lastHealthValues map[uint64]float64

	// Track active regen rates per entity
	regenRates map[uint64]float64

	// Genre affects bonus amounts (horror has slower regen, fantasy is standard)
	genreID string
}

// NewSpecializationHealthRegenSystem creates a new specialization health regen system.
func NewSpecializationHealthRegenSystem(world *World, seed int64) *SpecializationHealthRegenSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "specialization_health_regen")
		logEntry.Debug("SpecializationHealthRegenSystem created")
	}

	return &SpecializationHealthRegenSystem{
		world:            world,
		logger:           logEntry,
		rng:              rand.New(rand.NewSource(seed)),
		updateInterval:   0.5, // Regen every 0.5 seconds for smooth health recovery
		timeSinceCheck:   0,
		lastHealthValues: make(map[uint64]float64),
		regenRates:       make(map[uint64]float64),
		genreID:          "fantasy",
	}
}

// SetGenre sets the genre for genre-aware bonus calculations.
func (s *SpecializationHealthRegenSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("Genre set for specialization health regen")
	}
}

// Update checks class specializations and applies health regeneration.
func (s *SpecializationHealthRegenSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime

	if s.timeSinceCheck < s.updateInterval {
		return
	}
	elapsed := s.timeSinceCheck
	s.timeSinceCheck = 0

	for _, entity := range entities {
		s.processEntity(entity, elapsed)
	}
}

// processEntity applies health regen based on specialization.
func (s *SpecializationHealthRegenSystem) processEntity(entity *Entity, elapsed float64) {
	// Need both health and class_progression components
	if !entity.HasComponent("health") || !entity.HasComponent("class_progression") {
		// If entity lost components, clean up tracking
		s.removeTracking(entity.ID)
		return
	}

	healthComp, _ := entity.GetComponent("health")
	health, ok := healthComp.(*HealthComponent)
	if !ok || health.IsDead() {
		return
	}

	// Don't regen if at full health
	if health.Current >= health.Max {
		return
	}

	classComp, _ := entity.GetComponent("class_progression")
	progression, ok := classComp.(*ClassProgressionComponent)
	if !ok {
		return
	}

	// Calculate regen rate based on specialization
	regenRate := s.calculateRegenRate(progression, health.Max)

	// Store the regen rate for UI display
	s.regenRates[entity.ID] = regenRate

	// Apply regen: rate is per second, multiply by elapsed time
	regenAmount := regenRate * elapsed
	if regenAmount > 0 {
		oldHealth := health.Current
		health.Heal(regenAmount)

		if s.logger != nil && health.Current != oldHealth {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"class":          progression.Class.String(),
				"specialization": progression.Specialization.String(),
				"regen_rate":     regenRate,
				"regen_amount":   regenAmount,
				"old_health":     oldHealth,
				"new_health":     health.Current,
				"max_health":     health.Max,
			}).Debug("Specialization health regen applied")
		}
	}
}

// calculateRegenRate computes the health regen rate (HP per second) for a specialization.
// Base regen is 1% of max HP per second, modified by specialization and genre.
func (s *SpecializationHealthRegenSystem) calculateRegenRate(progression *ClassProgressionComponent, maxHealth float64) float64 {
	// Base regen: 0.5% of max health per second (very slow without specialization)
	baseRate := maxHealth * 0.005

	// Get specialization multiplier
	specMultiplier := s.getSpecializationMultiplier(progression.Specialization)

	// Add class bonus if no specialization
	if progression.Specialization == SpecializationNone {
		specMultiplier = s.getBaseClassMultiplier(progression.Class)
	}

	// Genre modifier
	genreMultiplier := s.getGenreMultiplier()

	return baseRate * specMultiplier * genreMultiplier
}

// getBaseClassMultiplier provides base regen for classes without specialization.
func (s *SpecializationHealthRegenSystem) getBaseClassMultiplier(class CharacterClass) float64 {
	switch class {
	case ClassWarrior:
		return 1.5 // +50% - hardy constitution
	case ClassCleric:
		return 1.8 // +80% - divine healing
	case ClassPaladin:
		return 1.7 // +70% - holy resilience
	case ClassMonk:
		return 1.6 // +60% - inner balance
	case ClassDeathKnight:
		return 1.4 // +40% - undead vitality
	case ClassBloodKnight:
		return 1.5 // +50% - blood absorption
	default:
		return 1.0
	}
}

// getSpecializationMultiplier returns the health regen multiplier for a specialization.
func (s *SpecializationHealthRegenSystem) getSpecializationMultiplier(spec SpecializationType) float64 {
	switch spec {
	// Tank specializations - highest regen (survival focus)
	case SpecializationDefender:
		return 3.0 // +200% - defensive mastery
	case SpecializationGuardian:
		return 2.8 // +180% - holy protection
	case SpecializationBrewmaster:
		return 2.5 // +150% - fortifying brews
	case SpecializationPackLeader:
		return 2.0 // +100% - pack vitality

	// Healer specializations - high regen (self-sustain)
	case SpecializationHealer:
		return 2.5 // +150% - divine restoration
	case SpecializationSoulweaver:
		return 2.3 // +130% - spirit link healing
	case SpecializationOracle:
		return 2.0 // +100% - divine foresight
	case SpecializationTheurgist:
		return 2.2 // +120% - holy arcana

	// Life drain specializations - moderate regen
	case SpecializationBloodMage:
		return 1.8 // +80% - vampiric absorption
	case SpecializationHemomancer:
		return 1.7 // +70% - blood magic
	case SpecializationCrimsonBlade:
		return 1.6 // +60% - blood edge
	case SpecializationAffliction:
		return 1.5 // +50% - life drain curses

	// Hybrid tank/damage specializations - moderate regen
	case SpecializationTemplar:
		return 1.8 // +80% - holy warrior sustain
	case SpecializationCrusader:
		return 1.6 // +60% - divine fury
	case SpecializationDeathKnight:
		return 1.5 // +50% - undead resilience
	case SpecializationUnholy:
		return 1.4 // +40% - festering wounds
	case SpecializationFrost:
		return 1.3 // +30% - frost armor

	// Evasion specializations - low regen (avoid damage instead)
	case SpecializationShadowdancer:
		return 1.2 // +20% - shadow recovery
	case SpecializationWindwalker:
		return 1.3 // +30% - inner peace
	case SpecializationAssassin:
		return 1.1 // +10% - minimal healing
	case SpecializationTrickster:
		return 1.1 // +10% - illusion recovery

	// Pure damage specializations - minimal regen
	case SpecializationBerserker:
		return 0.8 // -20% - rage penalty
	case SpecializationElementalist:
		return 1.0 // +0% - pure damage
	case SpecializationArcanist:
		return 1.0 // +0% - arcane focus
	case SpecializationMarksman:
		return 1.0 // +0% - precision focus

	// Other specializations
	case SpecializationBeastmaster:
		return 1.2 // +20% - pet bond healing
	case SpecializationExorcist:
		return 1.4 // +40% - holy cleansing
	case SpecializationPurifier:
		return 1.5 // +50% - purifying flame
	case SpecializationNaturemage:
		return 1.4 // +40% - nature regeneration
	case SpecializationShapeshifter:
		return 1.6 // +60% - natural healing
	case SpecializationSpellsword:
		return 1.2 // +20% - enchanted recovery
	case SpecializationWarmage:
		return 1.1 // +10% - magic barrier
	case SpecializationDuelist:
		return 1.2 // +20% - combat recovery
	case SpecializationFeralRage:
		return 1.3 // +30% - primal vitality
	case SpecializationSpellshot:
		return 1.0 // +0% - pure damage
	case SpecializationSeeker:
		return 1.0 // +0% - pursuit focus
	case SpecializationVoidcaller:
		return 1.2 // +20% - shadow absorption
	case SpecializationJudge:
		return 1.4 // +40% - righteous recovery
	case SpecializationInterrogator:
		return 1.1 // +10% - mental fortitude
	case SpecializationDemonologist:
		return 1.3 // +30% - demonic pact
	case SpecializationShinobi:
		return 1.0 // +0% - assassin focus
	case SpecializationStriker:
		return 1.0 // +0% - precision focus

	default:
		return 1.0 // Default: no bonus
	}
}

// getGenreMultiplier returns a multiplier based on genre.
func (s *SpecializationHealthRegenSystem) getGenreMultiplier() float64 {
	switch s.genreID {
	case "fantasy":
		return 1.0 // Standard healing world
	case "scifi":
		return 1.2 // Advanced medical tech
	case "horror":
		return 0.6 // Slow, punishing regen
	case "cyberpunk":
		return 1.1 // Cybernetic healing
	case "postapoc":
		return 0.8 // Scarce medical resources
	default:
		return 1.0
	}
}

// removeTracking removes tracking data for an entity.
func (s *SpecializationHealthRegenSystem) removeTracking(entityID uint64) {
	delete(s.lastHealthValues, entityID)
	delete(s.regenRates, entityID)
}

// GetRegenRateForEntity returns the current health regen rate for an entity (for UI).
func (s *SpecializationHealthRegenSystem) GetRegenRateForEntity(entityID uint64) float64 {
	if rate, exists := s.regenRates[entityID]; exists {
		return rate
	}
	return 0.0
}
