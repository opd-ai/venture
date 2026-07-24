// Package engine provides the SpecializationDefenseSystem which bridges class
// specializations with defense bonuses. This system grants passive defense increases
// to tank and melee specializations, creating meaningful gameplay incentives for
// specialization choices that complement the existing specialization systems.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// SpecializationDefenseSystem provides defense bonuses based on character class
// specialization. Tank specializations receive the highest bonuses (40-80%),
// melee specializations receive moderate bonuses (20-40%), and caster
// specializations receive small or no bonus.
type SpecializationDefenseSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// updateInterval controls how often we check specializations (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// defenseMod handles original/applied stat caching for the defense stat.
	defenseMod statBonusApplier

	// Genre affects bonus amounts (scifi has tech-armor, fantasy is standard)
	genreID string
}

// NewSpecializationDefenseSystem creates a new specialization defense system.
func NewSpecializationDefenseSystem(world *World, seed int64) *SpecializationDefenseSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "specialization_defense")
		logEntry.Debug("SpecializationDefenseSystem created")
	}

	return &SpecializationDefenseSystem{
		world:          world,
		logger:         logEntry,
		rng:            rand.New(rand.NewSource(seed)),
		updateInterval: 1.0, // Check once per second (specializations rarely change)
		defenseMod:     newStatBonusApplier(),
		genreID:        "fantasy",
	}
}

// SetGenre sets the genre for genre-aware bonus calculations.
func (s *SpecializationDefenseSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("Genre set for specialization defense")
	}
}

// Update checks class specializations and applies defense bonuses.
func (s *SpecializationDefenseSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime

	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	for _, entity := range entities {
		s.processEntity(entity)
	}
}

// processEntity checks if an entity has a specialization and applies defense bonus.
func (s *SpecializationDefenseSystem) processEntity(entity *Entity) {
	// Need both stats and class_progression components
	if !entity.HasComponent("stats") || !entity.HasComponent("class_progression") {
		// Restore the original defense value (if any) and clear the cache.
		s.defenseMod.restoreAndRemove(entity,
			func(st *StatsComponent, v float64) { st.Defense = v },
		)
		return
	}

	classComp, _ := entity.GetComponent("class_progression")
	progression, ok := classComp.(*ClassProgressionComponent)
	if !ok {
		return
	}

	bonus := s.calculateBonus(progression)
	changed := s.defenseMod.apply(entity, bonus,
		func(st *StatsComponent) float64 { return st.Defense },
		func(st *StatsComponent, v float64) { st.Defense = v },
		multiplicativeBonus,
	)
	if changed && s.logger != nil && bonus > 0 {
		statsComp, _ := entity.GetComponent("stats")
		stats, _ := statsComp.(*StatsComponent)
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entity.ID,
			"class":          progression.Class.String(),
			"specialization": progression.Specialization.String(),
			"bonus_percent":  bonus * 100,
			"new_defense":    stats.Defense,
		}).Debug("Specialization defense bonus applied")
	}
}

// calculateBonus computes the defense bonus multiplier for a specialization.
// Returns 0.0-0.8 (0-80% bonus) based on specialization and genre.
func (s *SpecializationDefenseSystem) calculateBonus(progression *ClassProgressionComponent) float64 {
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

// calculateBaseClassBonus provides small bonuses for defensive classes.
func (s *SpecializationDefenseSystem) calculateBaseClassBonus(class CharacterClass) float64 {
	switch class {
	case ClassWarrior:
		return 0.15 // +15% for warriors
	case ClassPaladin:
		return 0.12 // +12% for paladins
	case ClassDeathKnight, ClassBloodKnight:
		return 0.10 // +10% for death/blood knights
	case ClassMonk:
		return 0.08 // +8% for monks
	case ClassRanger:
		return 0.05 // +5% for rangers
	default:
		return 0.0
	}
}

// getSpecializationBonus returns the base defense bonus for a specialization.
func (s *SpecializationDefenseSystem) getSpecializationBonus(spec SpecializationType) float64 {
	switch spec {
	// Pure tank specializations - highest bonuses
	case SpecializationDefender:
		return 0.80 // +80% - ultimate defensive mastery
	case SpecializationGuardian:
		return 0.70 // +70% - holy shield
	case SpecializationBrewmaster:
		return 0.65 // +65% - stagger mechanics
	case SpecializationPackLeader:
		return 0.55 // +55% - pack protection

	// Hybrid tank/damage specializations - high bonuses
	case SpecializationTemplar:
		return 0.50 // +50% - holy warrior armor
	case SpecializationCrusader:
		return 0.45 // +45% - plate wielder
	case SpecializationDeathKnight:
		return 0.50 // +50% - undead fortitude
	case SpecializationFrost:
		return 0.45 // +45% - frost armor
	case SpecializationUnholy:
		return 0.35 // +35% - bone shield

	// Melee damage specializations - moderate bonuses
	case SpecializationBerserker:
		return 0.25 // +25% - rage armor (despite offense focus)
	case SpecializationDuelist:
		return 0.30 // +30% - parry mastery
	case SpecializationSpellsword:
		return 0.25 // +25% - enchanted armor
	case SpecializationCrimsonBlade:
		return 0.28 // +28% - blood armor
	case SpecializationWindwalker:
		return 0.20 // +20% - evasion stance
	case SpecializationStriker:
		return 0.15 // +15% - combat awareness

	// Evasion-focused specializations - low defense but high evasion
	case SpecializationAssassin:
		return 0.10 // +10% - light armor
	case SpecializationShadowdancer:
		return 0.12 // +12% - shadow defense
	case SpecializationTrickster:
		return 0.15 // +15% - illusion shield
	case SpecializationShinobi:
		return 0.08 // +8% - ninja agility

	// Pet/summon specializations - moderate protection
	case SpecializationBeastmaster:
		return 0.20 // +20% - animal bond
	case SpecializationDemonologist:
		return 0.25 // +25% - demonic pact shield
	case SpecializationFeralRage:
		return 0.30 // +30% - primal toughness
	case SpecializationShapeshifter:
		return 0.35 // +35% - form-dependent armor

	// Healer specializations - low to moderate defense
	case SpecializationHealer:
		return 0.15 // +15% - divine protection
	case SpecializationSoulweaver:
		return 0.18 // +18% - spirit barrier
	case SpecializationOracle:
		return 0.12 // +12% - foresight defense
	case SpecializationTheurgist:
		return 0.20 // +20% - holy arcane barrier
	case SpecializationNaturemage:
		return 0.18 // +18% - bark skin

	// Pure caster specializations - minimal defense
	case SpecializationArcanist:
		return 0.05 // +5% - mana shield minimal
	case SpecializationElementalist:
		return 0.08 // +8% - elemental resistance
	case SpecializationBloodMage:
		return 0.12 // +12% - blood barrier
	case SpecializationWarmage:
		return 0.20 // +20% - battle mage armor
	case SpecializationVoidcaller:
		return 0.10 // +10% - shadow cloak
	case SpecializationAffliction:
		return 0.08 // +8% - curse resistance
	case SpecializationHemomancer:
		return 0.10 // +10% - blood shield

	// Ranged specializations - low defense
	case SpecializationMarksman:
		return 0.12 // +12% - light armor
	case SpecializationSpellshot:
		return 0.15 // +15% - magic-infused armor
	case SpecializationSeeker:
		return 0.10 // +10% - mobile defense

	// Paladin specializations - moderate to high
	case SpecializationExorcist:
		return 0.35 // +35% - holy armor
	case SpecializationPurifier:
		return 0.38 // +38% - cleansing flame shield
	case SpecializationJudge:
		return 0.32 // +32% - righteous defense
	case SpecializationInterrogator:
		return 0.25 // +25% - mental fortitude

	default:
		return 0.10 // Small default bonus for unknown specializations
	}
}

// getGenreMultiplier returns a multiplier based on genre.
func (s *SpecializationDefenseSystem) getGenreMultiplier() float64 {
	switch s.genreID {
	case "fantasy":
		return 1.0 // Standard armor world
	case "scifi":
		return 1.2 // Tech armor and shields
	case "horror":
		return 0.8 // Armor feels less protective
	case "cyberpunk":
		return 1.1 // Cyber-enhancements
	case "postapoc":
		return 0.9 // Degraded armor quality
	default:
		return 1.0
	}
}

// GetBonusForEntity returns the current defense bonus for an entity (for UI).
func (s *SpecializationDefenseSystem) GetBonusForEntity(entityID uint64) float64 {
	return s.defenseMod.currentBonus(entityID)
}
