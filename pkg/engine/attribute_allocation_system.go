// Package engine provides the attribute allocation system for character customization.
// This system handles allocating attribute points and applying their effects to stats.
package engine

import (
	"fmt"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// AttributeAllocationSystem manages attribute point allocation and stat derivation.
// It handles level-up point grants, allocation, respec, and applying attribute bonuses.
type AttributeAllocationSystem struct {
	world           *World
	rng             *rand.Rand
	effects         *AttributeEffects
	genreID         string
	logEntry        *logrus.Entry
	updateInterval  float64 // How often to check for dirty components (seconds)
	timeSinceUpdate float64
	genreMultiplier float64 // Genre-based scaling modifier
}

// NewAttributeAllocationSystem creates a new attribute allocation system.
func NewAttributeAllocationSystem(world *World, seed int64) *AttributeAllocationSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil && world.logger.Logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "attribute_allocation")
	} else {
		logEntry = logrus.NewEntry(logrus.StandardLogger()).WithField("system_name", "attribute_allocation")
	}

	sys := &AttributeAllocationSystem{
		world:           world,
		rng:             rand.New(rand.NewSource(seed)),
		effects:         DefaultAttributeEffects(),
		logEntry:        logEntry,
		updateInterval:  0.5, // Check twice per second
		timeSinceUpdate: 0,
		genreMultiplier: 1.0,
	}

	logEntry.Debug("attribute allocation system created")
	return sys
}

// SetGenre configures genre-specific attribute scaling.
func (s *AttributeAllocationSystem) SetGenre(genreID string) {
	s.genreID = genreID
	switch genreID {
	case "fantasy":
		s.genreMultiplier = 1.0 // Balanced
	case "scifi":
		s.genreMultiplier = 0.9          // Tech relies less on raw stats
		s.effects.MagicBonusPerInt = 3.0 // Tech amplifies INT more
	case "horror":
		s.genreMultiplier = 1.1          // Stats matter more for survival
		s.effects.MaxHealthPerVit = 20.0 // Vitality more valuable
	case "cyberpunk":
		s.genreMultiplier = 0.95
		s.effects.AttackSpeedPerAgi = 0.7 // Augments boost agility effects
	case "postapoc":
		s.genreMultiplier = 1.05
		s.effects.StatusResistPerVit = 0.7 // Vitality helps vs radiation
	default:
		s.genreMultiplier = 1.0
	}

	if s.logEntry != nil {
		s.logEntry.WithFields(logrus.Fields{
			"genre":      genreID,
			"multiplier": s.genreMultiplier,
		}).Debug("genre set for attribute allocation")
	}
}

// Update processes all entities with attribute allocation components.
func (s *AttributeAllocationSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceUpdate += deltaTime
	if s.timeSinceUpdate < s.updateInterval {
		return
	}
	s.timeSinceUpdate = 0

	for _, entity := range entities {
		s.processEntity(entity)
	}
}

// processEntity checks and applies attribute bonuses for a single entity.
func (s *AttributeAllocationSystem) processEntity(entity *Entity) {
	comp, ok := entity.GetComponent("attribute_allocation")
	if !ok {
		return
	}

	attrComp, ok := comp.(*AttributeAllocationComponent)
	if !ok {
		return
	}

	// Only recalculate if dirty
	if !attrComp.Dirty {
		return
	}

	s.applyAttributeBonuses(entity, attrComp)
	attrComp.Dirty = false
}

// applyAttributeBonuses calculates and applies all attribute effects to entity stats.
func (s *AttributeAllocationSystem) applyAttributeBonuses(entity *Entity, attrComp *AttributeAllocationComponent) {
	statsComp, hasStats := entity.GetComponent("stats")
	healthComp, hasHealth := entity.GetComponent("health")
	manaComp, hasMana := entity.GetComponent("mana")

	var stats *StatsComponent
	if hasStats {
		stats, _ = statsComp.(*StatsComponent)
	}

	// Remove previously applied bonuses
	s.removeAppliedBonuses(entity, stats, attrComp, hasHealth, hasMana)

	effects := s.effects

	// Strength bonuses (requires StatsComponent)
	if stats != nil {
		strTotal := float64(attrComp.GetTotal(AttrStrength))
		strAttackBonus := strTotal * effects.AttackBonusPerStr * s.genreMultiplier
		stats.Attack += strAttackBonus
		attrComp.AppliedBonuses[AttrStrength] = strAttackBonus
	}

	// G26 fix: STR → carry capacity (InventoryComponent.MaxWeight)
	invComp, hasInv := entity.GetComponent("inventory")
	if hasInv {
		if inv, ok := invComp.(*InventoryComponent); ok {
			strTotal := float64(attrComp.GetTotal(AttrStrength))
			strCarryBonus := strTotal * effects.CarryCapPerStr * s.genreMultiplier
			inv.MaxWeight += strCarryBonus
			attrComp.AppliedBonuses[attrTertiaryOffset+AttrStrength] = strCarryBonus
		}
	}

	// Agility bonuses (evasion on StatsComponent)
	if stats != nil {
		agiTotal := float64(attrComp.GetTotal(AttrAgility))
		agiEvasionBonus := agiTotal * effects.EvasionBonusPerAgi / 100.0 * s.genreMultiplier
		stats.Evasion += agiEvasionBonus
		attrComp.AppliedBonuses[AttrAgility] = agiEvasionBonus
	}

	// G26 fix: AGI → movement speed bonus (StatsComponent.SpeedBonus)
	if stats != nil {
		agiTotal := float64(attrComp.GetTotal(AttrAgility))
		agiSpeedBonus := agiTotal * effects.SpeedBonusPerAgi / 100.0 * s.genreMultiplier
		stats.SpeedBonus += agiSpeedBonus
		attrComp.AppliedBonuses[attrTertiaryOffset+AttrAgility] = agiSpeedBonus
	}

	// Intelligence bonuses (MagicPower on StatsComponent, mana on ManaComponent)
	if stats != nil {
		intTotal := float64(attrComp.GetTotal(AttrIntelligence))
		intMagicBonus := intTotal * effects.MagicBonusPerInt * s.genreMultiplier
		stats.MagicPower += intMagicBonus
		attrComp.AppliedBonuses[AttrIntelligence] = intMagicBonus
	}
	if hasHealth && hasMana {
		mana, ok := manaComp.(*ManaComponent)
		if ok {
			intTotal := float64(attrComp.GetTotal(AttrIntelligence))
			intManaBonus := int(intTotal * effects.MaxManaPerInt * s.genreMultiplier)
			mana.Max += intManaBonus
			attrComp.AppliedBonuses[attrSecondaryOffset+AttrIntelligence] = float64(intManaBonus) // Store mana bonus separately
		}
	}

	// G26 fix: INT → mana regen (ManaComponent.Regen)
	if hasMana {
		if mana, ok := manaComp.(*ManaComponent); ok {
			intTotal := float64(attrComp.GetTotal(AttrIntelligence))
			intManaRegenBonus := intTotal * effects.ManaRegenPerInt * s.genreMultiplier
			mana.Regen += intManaRegenBonus
			attrComp.AppliedBonuses[attrTertiaryOffset+AttrIntelligence] = intManaRegenBonus
		}
	}

	// Vitality bonuses (HealthComponent)
	if hasHealth {
		health, ok := healthComp.(*HealthComponent)
		if ok {
			vitTotal := float64(attrComp.GetTotal(AttrVitality))
			vitHealthBonus := vitTotal * effects.MaxHealthPerVit * s.genreMultiplier
			health.Max += vitHealthBonus
			attrComp.AppliedBonuses[AttrVitality] = vitHealthBonus
		}
	}

	// G26 fix: VIT → health regen (HealthComponent.RegenRate)
	if hasHealth {
		if health, ok := healthComp.(*HealthComponent); ok {
			vitTotal := float64(attrComp.GetTotal(AttrVitality))
			vitRegenBonus := vitTotal * effects.HealthRegenPerVit * s.genreMultiplier
			health.RegenRate += vitRegenBonus
			attrComp.AppliedBonuses[attrTertiaryOffset+AttrVitality] = vitRegenBonus
		}
	}

	// Endurance bonuses (Defense and BlockChance on StatsComponent)
	if stats != nil {
		endTotal := float64(attrComp.GetTotal(AttrEndurance))
		endDefenseBonus := endTotal * effects.DefenseBonusPerEnd * s.genreMultiplier
		endBlockBonus := endTotal * effects.BlockChancePerEnd / 100.0 * s.genreMultiplier
		stats.Defense += endDefenseBonus
		stats.BlockChance += endBlockBonus
		attrComp.AppliedBonuses[AttrEndurance] = endDefenseBonus
	}

	// Luck bonuses (CritChance on StatsComponent)
	if stats != nil {
		lckTotal := float64(attrComp.GetTotal(AttrLuck))
		lckCritBonus := lckTotal * effects.CritChancePerLuck / 100.0 * s.genreMultiplier
		stats.CritChance += lckCritBonus
		attrComp.AppliedBonuses[AttrLuck] = lckCritBonus
	}

	if s.logEntry != nil && s.logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		s.logEntry.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"str_bonus":   attrComp.AppliedBonuses[AttrStrength],
			"int_bonus":   attrComp.AppliedBonuses[AttrIntelligence],
			"vit_bonus":   attrComp.AppliedBonuses[AttrVitality],
			"end_bonus":   attrComp.AppliedBonuses[AttrEndurance],
			"agi_evasion": attrComp.AppliedBonuses[AttrAgility],
			"lck_crit":    attrComp.AppliedBonuses[AttrLuck],
		}).Debug("attribute bonuses applied")
	}
}

// removeAppliedBonuses removes previously applied attribute bonuses from stats.
func (s *AttributeAllocationSystem) removeAppliedBonuses(entity *Entity, stats *StatsComponent, attrComp *AttributeAllocationComponent, hasHealth, hasMana bool) {
	s.removeStrengthBonus(stats, attrComp)
	s.removeStrengthCarryBonus(entity, attrComp)
	s.removeAgilityBonus(stats, attrComp)
	s.removeAgilitySpeedBonus(stats, attrComp)
	s.removeIntelligenceBonus(stats, attrComp)
	s.removeManaBonus(entity, attrComp, hasMana)
	s.removeManaRegenBonus(entity, attrComp, hasMana)
	s.removeVitalityBonus(entity, attrComp, hasHealth)
	s.removeVitalityRegenBonus(entity, attrComp, hasHealth)
	s.removeEnduranceBonus(stats, attrComp)
	s.removeLuckBonus(stats, attrComp)
}

// removeStrengthBonus removes the attack bonus from strength.
func (s *AttributeAllocationSystem) removeStrengthBonus(stats *StatsComponent, attrComp *AttributeAllocationComponent) {
	if stats == nil {
		return
	}
	if bonus, ok := attrComp.AppliedBonuses[AttrStrength]; ok {
		stats.Attack -= bonus
	}
}

// removeStrengthCarryBonus removes the carry capacity bonus from strength (G26 fix).
func (s *AttributeAllocationSystem) removeStrengthCarryBonus(entity *Entity, attrComp *AttributeAllocationComponent) {
	if bonus, ok := attrComp.AppliedBonuses[attrTertiaryOffset+AttrStrength]; ok {
		if invComp, hasInv := entity.GetComponent("inventory"); hasInv {
			if inv, ok2 := invComp.(*InventoryComponent); ok2 {
				inv.MaxWeight -= bonus
			}
		}
	}
}

// removeAgilityBonus removes the evasion bonus from agility.
func (s *AttributeAllocationSystem) removeAgilityBonus(stats *StatsComponent, attrComp *AttributeAllocationComponent) {
	if stats == nil {
		return
	}
	if bonus, ok := attrComp.AppliedBonuses[AttrAgility]; ok {
		stats.Evasion -= bonus
	}
}

// removeAgilitySpeedBonus removes the speed bonus from agility (G26 fix).
func (s *AttributeAllocationSystem) removeAgilitySpeedBonus(stats *StatsComponent, attrComp *AttributeAllocationComponent) {
	if stats == nil {
		return
	}
	if bonus, ok := attrComp.AppliedBonuses[attrTertiaryOffset+AttrAgility]; ok {
		stats.SpeedBonus -= bonus
	}
}

// removeIntelligenceBonus removes the magic power bonus from intelligence.
func (s *AttributeAllocationSystem) removeIntelligenceBonus(stats *StatsComponent, attrComp *AttributeAllocationComponent) {
	if stats == nil {
		return
	}
	if bonus, ok := attrComp.AppliedBonuses[AttrIntelligence]; ok {
		stats.MagicPower -= bonus
	}
}

// removeManaBonus removes the mana bonus from intelligence.
func (s *AttributeAllocationSystem) removeManaBonus(entity *Entity, attrComp *AttributeAllocationComponent, hasMana bool) {
	if !hasMana {
		return
	}
	manaComp, ok := entity.GetComponent("mana")
	if !ok {
		return
	}
	mana, ok := manaComp.(*ManaComponent)
	if !ok {
		return
	}
	if manaBonus, ok := attrComp.AppliedBonuses[attrSecondaryOffset+AttrIntelligence]; ok {
		mana.Max -= int(manaBonus)
	}
}

// removeManaRegenBonus removes the mana regen bonus from intelligence (G26 fix).
func (s *AttributeAllocationSystem) removeManaRegenBonus(entity *Entity, attrComp *AttributeAllocationComponent, hasMana bool) {
	if !hasMana {
		return
	}
	manaComp, ok := entity.GetComponent("mana")
	if !ok {
		return
	}
	mana, ok := manaComp.(*ManaComponent)
	if !ok {
		return
	}
	if bonus, ok := attrComp.AppliedBonuses[attrTertiaryOffset+AttrIntelligence]; ok {
		mana.Regen -= bonus
	}
}

// removeVitalityBonus removes the health bonus from vitality.
func (s *AttributeAllocationSystem) removeVitalityBonus(entity *Entity, attrComp *AttributeAllocationComponent, hasHealth bool) {
	if !hasHealth {
		return
	}
	healthComp, ok := entity.GetComponent("health")
	if !ok {
		return
	}
	health, ok := healthComp.(*HealthComponent)
	if !ok {
		return
	}
	if bonus, ok := attrComp.AppliedBonuses[AttrVitality]; ok {
		health.Max -= bonus
	}
}

// removeVitalityRegenBonus removes the health regen bonus from vitality (G26 fix).
func (s *AttributeAllocationSystem) removeVitalityRegenBonus(entity *Entity, attrComp *AttributeAllocationComponent, hasHealth bool) {
	if !hasHealth {
		return
	}
	healthComp, ok := entity.GetComponent("health")
	if !ok {
		return
	}
	health, ok := healthComp.(*HealthComponent)
	if !ok {
		return
	}
	if bonus, ok := attrComp.AppliedBonuses[attrTertiaryOffset+AttrVitality]; ok {
		health.RegenRate -= bonus
	}
}

// removeEnduranceBonus removes the defense and block bonuses from endurance.
func (s *AttributeAllocationSystem) removeEnduranceBonus(stats *StatsComponent, attrComp *AttributeAllocationComponent) {
	if stats == nil {
		return
	}
	if bonus, ok := attrComp.AppliedBonuses[AttrEndurance]; ok {
		stats.Defense -= bonus
		endTotal := float64(attrComp.GetTotal(AttrEndurance))
		blockBonus := endTotal * s.effects.BlockChancePerEnd / 100.0 * s.genreMultiplier
		stats.BlockChance -= blockBonus
	}
}

// removeLuckBonus removes the crit chance bonus from luck.
func (s *AttributeAllocationSystem) removeLuckBonus(stats *StatsComponent, attrComp *AttributeAllocationComponent) {
	if stats == nil {
		return
	}
	if bonus, ok := attrComp.AppliedBonuses[AttrLuck]; ok {
		stats.CritChance -= bonus
	}
}

// AllocatePoint allocates a single attribute point.
func (s *AttributeAllocationSystem) AllocatePoint(entity *Entity, attr CoreAttribute) error {
	return s.AllocatePoints(entity, attr, 1)
}

// AllocatePoints allocates multiple points to an attribute.
func (s *AttributeAllocationSystem) AllocatePoints(entity *Entity, attr CoreAttribute, points int) error {
	if points <= 0 {
		return fmt.Errorf("points must be positive")
	}

	comp, ok := entity.GetComponent("attribute_allocation")
	if !ok {
		return fmt.Errorf("entity %d has no attribute_allocation component", entity.ID)
	}

	attrComp, ok := comp.(*AttributeAllocationComponent)
	if !ok {
		return fmt.Errorf("invalid attribute_allocation component type")
	}

	if !attrComp.CanAllocate(attr, points) {
		return fmt.Errorf("cannot allocate %d points to %s: only %d unspent",
			points, attr.String(), attrComp.UnspentPoints)
	}

	// Allocate the points
	attrComp.AllocatedPoints[attr] += points
	attrComp.UnspentPoints -= points
	attrComp.Dirty = true

	if s.logEntry != nil {
		s.logEntry.WithFields(logrus.Fields{
			"entity_id":    entity.ID,
			"attribute":    attr.String(),
			"points":       points,
			"new_total":    attrComp.GetTotal(attr),
			"unspent_left": attrComp.UnspentPoints,
		}).Info("attribute points allocated")
	}

	return nil
}

// AwardPoints grants attribute points to an entity (e.g., on level up).
func (s *AttributeAllocationSystem) AwardPoints(entity *Entity, points int) error {
	if points <= 0 {
		return fmt.Errorf("points must be positive")
	}

	comp, ok := entity.GetComponent("attribute_allocation")
	if !ok {
		return fmt.Errorf("entity %d has no attribute_allocation component", entity.ID)
	}

	attrComp, ok := comp.(*AttributeAllocationComponent)
	if !ok {
		return fmt.Errorf("invalid attribute_allocation component type")
	}

	attrComp.UnspentPoints += points
	attrComp.TotalPointsEarned += points

	if s.logEntry != nil {
		s.logEntry.WithFields(logrus.Fields{
			"entity_id":      entity.ID,
			"points_awarded": points,
			"unspent_total":  attrComp.UnspentPoints,
			"lifetime_total": attrComp.TotalPointsEarned,
		}).Info("attribute points awarded")
	}

	return nil
}

// SetBonusPoints sets temporary bonus points for an attribute (e.g., from equipment).
func (s *AttributeAllocationSystem) SetBonusPoints(entity *Entity, attr CoreAttribute, points int) error {
	comp, ok := entity.GetComponent("attribute_allocation")
	if !ok {
		return fmt.Errorf("entity %d has no attribute_allocation component", entity.ID)
	}

	attrComp, ok := comp.(*AttributeAllocationComponent)
	if !ok {
		return fmt.Errorf("invalid attribute_allocation component type")
	}

	attrComp.BonusPoints[attr] = points
	attrComp.Dirty = true

	return nil
}

// Respec resets all allocated points, returning them to the unspent pool.
// Optionally costs gold (0 = free respec).
func (s *AttributeAllocationSystem) Respec(entity *Entity, goldCost int) error {
	comp, ok := entity.GetComponent("attribute_allocation")
	if !ok {
		return fmt.Errorf("entity %d has no attribute_allocation component", entity.ID)
	}

	attrComp, ok := comp.(*AttributeAllocationComponent)
	if !ok {
		return fmt.Errorf("invalid attribute_allocation component type")
	}

	// Check gold if cost is required
	if goldCost > 0 {
		invComp, ok := entity.GetComponent("inventory")
		if !ok {
			return fmt.Errorf("entity has no inventory for respec cost")
		}
		inv, ok := invComp.(*InventoryComponent)
		if !ok {
			return fmt.Errorf("invalid inventory component type")
		}
		if inv.Gold < goldCost {
			return fmt.Errorf("not enough gold: need %d, have %d", goldCost, inv.Gold)
		}
		inv.Gold -= goldCost
	}

	// Calculate total allocated points
	totalAllocated := attrComp.TotalAllocatedPoints()

	// Reset all allocations
	for i := 0; i < int(NumCoreAttributes); i++ {
		attrComp.AllocatedPoints[i] = 0
	}

	// Return points to unspent pool
	attrComp.UnspentPoints += totalAllocated
	attrComp.RespecCount++
	attrComp.Dirty = true

	if s.logEntry != nil {
		s.logEntry.WithFields(logrus.Fields{
			"entity_id":       entity.ID,
			"points_refunded": totalAllocated,
			"gold_cost":       goldCost,
			"respec_count":    attrComp.RespecCount,
		}).Info("attribute respec completed")
	}

	return nil
}

// GetRespecCost calculates the gold cost for a respec (scales with respec count).
func (s *AttributeAllocationSystem) GetRespecCost(entity *Entity) int {
	comp, ok := entity.GetComponent("attribute_allocation")
	if !ok {
		return 0
	}

	attrComp, ok := comp.(*AttributeAllocationComponent)
	if !ok {
		return 0
	}

	// Base cost 500 gold, doubles with each respec (up to cap)
	baseCost := 500
	multiplier := 1 << attrComp.RespecCount // 2^respecCount
	if multiplier > 16 {
		multiplier = 16 // Cap at 8000 gold
	}

	return baseCost * multiplier
}

// GetAttributeTotal returns the total value for an attribute.
func (s *AttributeAllocationSystem) GetAttributeTotal(entity *Entity, attr CoreAttribute) int {
	comp, ok := entity.GetComponent("attribute_allocation")
	if !ok {
		return 0
	}

	attrComp, ok := comp.(*AttributeAllocationComponent)
	if !ok {
		return 0
	}

	return attrComp.GetTotal(attr)
}

// GetUnspentPoints returns the number of unspent attribute points.
func (s *AttributeAllocationSystem) GetUnspentPoints(entity *Entity) int {
	comp, ok := entity.GetComponent("attribute_allocation")
	if !ok {
		return 0
	}

	attrComp, ok := comp.(*AttributeAllocationComponent)
	if !ok {
		return 0
	}

	return attrComp.UnspentPoints
}

// GetEffects returns the current attribute effects configuration.
func (s *AttributeAllocationSystem) GetEffects() *AttributeEffects {
	return s.effects
}

// SetEffects allows customizing attribute effect values.
func (s *AttributeAllocationSystem) SetEffects(effects *AttributeEffects) {
	if effects != nil {
		s.effects = effects
	}
}
