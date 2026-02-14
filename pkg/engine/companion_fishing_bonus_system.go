// Package engine provides the CompanionFishingBonusSystem which bridges companion types
// with fishing mechanics. This system modifies fishing bonuses based on nearby companions,
// making water elementals improve catch rates, spirits reveal rare fish, and pets reduce
// line tension through their calming presence.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// CompanionFishingBonusSystem modifies fishing spot bonuses based on nearby companions.
// Water/Spirit elementals boost rare fish chance, pets reduce tension, robots scan for
// rare fish, insects act as living bait. This connects CompanionComponent with FishingSystem
// for immersive fishing gameplay.
//
// Companion type modifiers:
//   - Pet: +15% catch speed (comfort reduces fumbling)
//   - Elemental: +25% rare fish if water-type, +10% otherwise
//   - Spirit: +20% rare fish (senses fish presence)
//   - Robot: +30% legendary fish detection
//   - Insect: +20% bite speed (natural bait attraction)
//   - Undead: +15% rare fish in deep water (darkness affinity)
//   - Summon: +10% all bonuses (magical assistance)
//   - Hireling: +10% catch speed (extra hands)
//
// Genre-specific modifiers:
//   - Fantasy: spirit bonus +35% instead of +20%
//   - Scifi: robot bonus +45% instead of +30%
//   - Horror: undead bonus +30% instead of +15%
//   - Cyberpunk: robot bonus +40%, pet penalty -10%
//   - Postapoc: insect bonus +35% (mutation attraction)
type CompanionFishingBonusSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// updateInterval controls how often we check companions (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Cache applied bonuses per entity to track changes
	appliedBonuses map[uint64]*CompanionFishingBonusData

	// Genre affects which companion types boost fishing
	genreID string

	// fishingSystem reference for bonus integration
	fishingSystem *FishingSystem
}

// CompanionFishingBonusData stores calculated companion bonuses for a fisher.
type CompanionFishingBonusData struct {
	// RareFishBonus is the multiplier for rare fish chance (1.0 = normal)
	RareFishBonus float64

	// CatchSpeedBonus is the multiplier for catch speed (1.0 = normal)
	CatchSpeedBonus float64

	// TensionReductionBonus reduces line tension buildup (1.0 = normal, <1 = reduced)
	TensionReductionBonus float64

	// CompanionType is the type providing the primary bonus
	CompanionType CompanionType

	// Loyalty is the companion's loyalty level affecting bonus strength
	Loyalty float64
}

// CompanionFishingBonusComponent is a transient component storing companion-based
// fishing modifiers. It is not persisted (recalculated each session).
type CompanionFishingBonusComponent struct {
	// RareFishMultiplier applied to base rare fish chance (1.0 = normal)
	RareFishMultiplier float64

	// CatchSpeedMultiplier applied to catch timing (1.0 = normal)
	CatchSpeedMultiplier float64

	// TensionReduction applied to line tension buildup (1.0 = normal, <1 = reduced)
	TensionReduction float64

	// CompanionTypeID is the type of companion providing the bonus
	CompanionTypeID int
}

// Type returns the component type identifier.
func (c *CompanionFishingBonusComponent) Type() string {
	return "companion_fishing_bonus"
}

// NewCompanionFishingBonusSystem creates a new companion fishing bonus system.
func NewCompanionFishingBonusSystem(world *World, seed int64) *CompanionFishingBonusSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "companion_fishing_bonus")
		logEntry.Debug("CompanionFishingBonusSystem created")
	}

	return &CompanionFishingBonusSystem{
		world:          world,
		logger:         logEntry,
		rng:            rand.New(rand.NewSource(seed)),
		updateInterval: 0.5, // Check twice per second
		appliedBonuses: make(map[uint64]*CompanionFishingBonusData, 32),
		genreID:        "fantasy",
	}
}

// SetGenre sets the genre for genre-aware fishing bonuses.
func (s *CompanionFishingBonusSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("Genre set for companion fishing bonus")
	}
}

// SetFishingSystem sets the reference to FishingSystem for integration.
func (s *CompanionFishingBonusSystem) SetFishingSystem(fs *FishingSystem) {
	s.fishingSystem = fs
}

// Update checks nearby companions and modifies fishing bonuses for active fishers.
func (s *CompanionFishingBonusSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime

	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	if s.world == nil {
		return
	}

	// Build companion lookup by owner
	companionsByOwner := s.buildCompanionMap(entities)

	// Process each fishing entity
	for _, entity := range entities {
		s.updateFisherBonus(entity, companionsByOwner)
	}
}

// buildCompanionMap creates a map of owner ID to their companion entities.
func (s *CompanionFishingBonusSystem) buildCompanionMap(entities []*Entity) map[uint64][]*Entity {
	result := make(map[uint64][]*Entity, 16)

	for _, entity := range entities {
		comp, ok := entity.GetComponent("companion")
		if !ok {
			continue
		}

		companion, ok := comp.(*CompanionComponent)
		if !ok {
			continue
		}

		ownerID := companion.OwnerID
		result[ownerID] = append(result[ownerID], entity)
	}

	return result
}

// updateFisherBonus updates fishing bonuses for a single entity.
func (s *CompanionFishingBonusSystem) updateFisherBonus(entity *Entity, companionsByOwner map[uint64][]*Entity) {
	// Only process entities that are fishing
	fishComp, hasFishing := entity.GetComponent("fishing")
	if !hasFishing {
		s.removeBonus(entity)
		return
	}

	fisher, ok := fishComp.(*FishingComponent)
	if !ok || fisher.GetState() == FishingStateIdle {
		s.removeBonus(entity)
		return
	}

	// Get companions for this entity
	companions := companionsByOwner[entity.ID]
	if len(companions) == 0 {
		s.removeBonus(entity)
		return
	}

	// Find the best companion bonus
	bestBonus := s.calculateBestBonus(companions)
	if bestBonus == nil {
		s.removeBonus(entity)
		return
	}

	// Apply bonus to entity
	s.applyBonus(entity, bestBonus)
}

// calculateBestBonus finds the best fishing bonus from available companions.
func (s *CompanionFishingBonusSystem) calculateBestBonus(companions []*Entity) *CompanionFishingBonusData {
	var best *CompanionFishingBonusData

	for _, compEntity := range companions {
		comp, ok := compEntity.GetComponent("companion")
		if !ok {
			continue
		}

		companion, ok := comp.(*CompanionComponent)
		if !ok {
			continue
		}

		bonus := s.calculateCompanionBonus(companion)

		// Keep the bonus with highest combined value
		if best == nil || bonus.combinedValue() > best.combinedValue() {
			best = bonus
		}
	}

	return best
}

// combinedValue returns a single value representing total bonus strength.
func (b *CompanionFishingBonusData) combinedValue() float64 {
	// Weight rare fish bonus highest, then speed, then tension
	return (b.RareFishBonus-1.0)*2.0 + (b.CatchSpeedBonus-1.0)*1.5 + (1.0-b.TensionReductionBonus)*1.0
}

// calculateCompanionBonus computes fishing bonuses from a companion.
func (s *CompanionFishingBonusSystem) calculateCompanionBonus(companion *CompanionComponent) *CompanionFishingBonusData {
	bonus := &CompanionFishingBonusData{
		RareFishBonus:         1.0,
		CatchSpeedBonus:       1.0,
		TensionReductionBonus: 1.0,
		CompanionType:         companion.CompanionType,
		Loyalty:               companion.Loyalty,
	}

	// Loyalty multiplier: 50-100 loyalty = 0.75-1.25x bonus strength
	loyaltyMult := 0.5 + (companion.Loyalty/100.0)*0.75

	// Apply base bonuses by companion type
	switch companion.CompanionType {
	case CompanionTypePet:
		// Pets provide comfort, reducing tension and speeding catches
		bonus.CatchSpeedBonus = 1.0 + 0.15*loyaltyMult
		bonus.TensionReductionBonus = 1.0 - 0.10*loyaltyMult

	case CompanionTypeElemental:
		// Water-affinity elementals excel at fishing
		bonus.RareFishBonus = 1.0 + 0.25*loyaltyMult

	case CompanionTypeSpirit:
		// Spirits sense fish presence
		bonus.RareFishBonus = 1.0 + 0.20*loyaltyMult

	case CompanionTypeRobot:
		// Robots scan for legendary fish
		bonus.RareFishBonus = 1.0 + 0.30*loyaltyMult

	case CompanionTypeInsect:
		// Insects act as natural bait
		bonus.CatchSpeedBonus = 1.0 + 0.20*loyaltyMult

	case CompanionTypeUndead:
		// Undead thrive in deep, dark waters
		bonus.RareFishBonus = 1.0 + 0.15*loyaltyMult

	case CompanionTypeSummon:
		// Magical summons provide general assistance
		bonus.RareFishBonus = 1.0 + 0.10*loyaltyMult
		bonus.CatchSpeedBonus = 1.0 + 0.10*loyaltyMult

	case CompanionTypeHireling:
		// Hirelings provide extra hands
		bonus.CatchSpeedBonus = 1.0 + 0.10*loyaltyMult
	}

	// Apply genre modifiers
	s.applyGenreModifiers(bonus, companion.CompanionType)

	return bonus
}

// applyGenreModifiers adjusts bonuses based on genre.
func (s *CompanionFishingBonusSystem) applyGenreModifiers(bonus *CompanionFishingBonusData, compType CompanionType) {
	switch s.genreID {
	case "fantasy":
		// Spirits are stronger in fantasy
		if compType == CompanionTypeSpirit {
			bonus.RareFishBonus += 0.15 // +35% total instead of +20%
		}

	case "scifi":
		// Robots are stronger in sci-fi
		if compType == CompanionTypeRobot {
			bonus.RareFishBonus += 0.15 // +45% total instead of +30%
		}

	case "horror":
		// Undead are stronger in horror
		if compType == CompanionTypeUndead {
			bonus.RareFishBonus += 0.15 // +30% total instead of +15%
		}

	case "cyberpunk":
		// Robots are stronger, pets are weaker
		if compType == CompanionTypeRobot {
			bonus.RareFishBonus += 0.10 // +40% total
		}
		if compType == CompanionTypePet {
			bonus.CatchSpeedBonus -= 0.10 // Only +5% instead of +15%
		}

	case "postapoc":
		// Mutant insects are more effective
		if compType == CompanionTypeInsect {
			bonus.CatchSpeedBonus += 0.15 // +35% total instead of +20%
		}
	}
}

// applyBonus applies the calculated companion bonus to a fishing entity.
func (s *CompanionFishingBonusSystem) applyBonus(entity *Entity, bonus *CompanionFishingBonusData) {
	// Create or update companion fishing bonus component
	comp, exists := entity.GetComponent("companion_fishing_bonus")
	if !exists {
		comp = &CompanionFishingBonusComponent{}
		entity.AddComponent(comp)
	}

	bonusComp := comp.(*CompanionFishingBonusComponent)
	bonusComp.RareFishMultiplier = bonus.RareFishBonus
	bonusComp.CatchSpeedMultiplier = bonus.CatchSpeedBonus
	bonusComp.TensionReduction = bonus.TensionReductionBonus
	bonusComp.CompanionTypeID = int(bonus.CompanionType)

	// Cache applied bonus
	s.appliedBonuses[entity.ID] = bonus

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":        entity.ID,
			"rare_multiplier":  bonus.RareFishBonus,
			"speed_multiplier": bonus.CatchSpeedBonus,
			"tension_mult":     bonus.TensionReductionBonus,
			"companion_type":   bonus.CompanionType,
			"loyalty":          bonus.Loyalty,
		}).Debug("Applied companion fishing bonus")
	}
}

// removeBonus removes companion fishing bonus from an entity.
func (s *CompanionFishingBonusSystem) removeBonus(entity *Entity) {
	if _, exists := entity.GetComponent("companion_fishing_bonus"); exists {
		entity.RemoveComponent("companion_fishing_bonus")
	}
	delete(s.appliedBonuses, entity.ID)
}

// GetCompanionBonus returns the companion bonus data for an entity.
func (s *CompanionFishingBonusSystem) GetCompanionBonus(entityID uint64) (*CompanionFishingBonusData, bool) {
	bonus, ok := s.appliedBonuses[entityID]
	return bonus, ok
}

// GetBonusForCompanionType returns the base bonus for a companion type.
// Useful for UI display or tooltips.
func (s *CompanionFishingBonusSystem) GetBonusForCompanionType(compType CompanionType, loyalty float64) *CompanionFishingBonusData {
	return s.calculateCompanionBonus(&CompanionComponent{
		CompanionType: compType,
		Loyalty:       loyalty,
	})
}
