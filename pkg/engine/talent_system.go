// Package engine provides the talent system for applying talent bonuses.
// This system reads TalentComponent and applies stat modifications to entities.
package engine

import (
	log "github.com/sirupsen/logrus"
)

// TalentSystem applies talent bonuses to entity stats.
// Recalculates bonuses when the Dirty flag is set and applies them to
// health, mana, stats, and other combat-relevant components.
type TalentSystem struct {
	world *World

	// Update frequency to check dirty flags
	updateInterval int
	frameCounter   int

	// Callbacks for talent events
	onTalentAllocated func(entity *Entity, talentID string, newRank int)
	onTalentReset     func(entity *Entity)
}

// NewTalentSystem creates a new talent system.
func NewTalentSystem(world *World) *TalentSystem {
	log.WithFields(log.Fields{
		"system_name":     "talent",
		"update_interval": 30,
	}).Debug("Creating talent system")

	return &TalentSystem{
		world:          world,
		updateInterval: 30, // Check dirty flags every 30 frames
		frameCounter:   0,
	}
}

// Update processes talent bonus recalculations for dirty entities.
func (s *TalentSystem) Update(entities []*Entity, deltaTime float64) {
	s.frameCounter++
	if s.frameCounter < s.updateInterval {
		return
	}
	s.frameCounter = 0

	processed := 0
	for _, entity := range entities {
		if s.processEntity(entity) {
			processed++
		}
	}

	if processed > 0 && log.GetLevel() >= log.DebugLevel {
		log.WithFields(log.Fields{
			"system_name":        "talent",
			"entities_processed": processed,
		}).Debug("Talent bonuses recalculated")
	}
}

// processEntity recalculates talent bonuses if dirty.
func (s *TalentSystem) processEntity(entity *Entity) bool {
	talentComp, hasTalent := entity.GetComponent("talent")
	if !hasTalent {
		return false
	}
	talent, ok := talentComp.(*TalentComponent)
	if !ok || !talent.Dirty {
		return false
	}
	if talent.AppliedDeltas == nil {
		talent.AppliedDeltas = make(map[string]float64)
	}

	// G23 fix: remove previously applied bonuses before computing new ones so
	// that talent resets and reallocations don't permanently stack old values.
	s.removeTalentBonuses(entity, talent.AppliedDeltas)

	// Calculate total bonuses from all allocated talents
	bonuses := s.calculateTotalBonuses(talent)
	talent.CachedBonuses = bonuses

	// Apply bonuses to entity stats and record absolute deltas for future removal.
	talent.AppliedDeltas = s.applyBonusesTracked(entity, bonuses)
	talent.AppliedBonuses = bonuses

	talent.Dirty = false
	return true
}

// calculateTotalBonuses sums all bonuses from allocated talents.
func (s *TalentSystem) calculateTotalBonuses(talent *TalentComponent) TalentBonus {
	total := TalentBonus{}

	for talentID, ranks := range talent.Allocations {
		if ranks <= 0 {
			continue
		}
		def := GetTalentDefinition(talentID)
		if def == nil {
			continue
		}

		// Multiply bonus per rank by number of ranks
		bonus := def.BonusPerRank
		total.FlatHealth += bonus.FlatHealth * float64(ranks)
		total.FlatMana += bonus.FlatMana * float64(ranks)
		total.FlatDamage += bonus.FlatDamage * float64(ranks)
		total.FlatDefense += bonus.FlatDefense * float64(ranks)
		total.FlatMagicPower += bonus.FlatMagicPower * float64(ranks)
		total.FlatMagicDefense += bonus.FlatMagicDefense * float64(ranks)
		total.FlatSpeed += bonus.FlatSpeed * float64(ranks)

		total.HealthPercent += bonus.HealthPercent * float64(ranks)
		total.ManaPercent += bonus.ManaPercent * float64(ranks)
		total.DamagePercent += bonus.DamagePercent * float64(ranks)
		total.DefensePercent += bonus.DefensePercent * float64(ranks)
		total.MagicPowerPercent += bonus.MagicPowerPercent * float64(ranks)
		total.MagicDefensePercent += bonus.MagicDefensePercent * float64(ranks)
		total.SpeedPercent += bonus.SpeedPercent * float64(ranks)

		total.CritChanceBonus += bonus.CritChanceBonus * float64(ranks)
		total.CritDamageBonus += bonus.CritDamageBonus * float64(ranks)
		total.LifestealPercent += bonus.LifestealPercent * float64(ranks)
		total.CooldownReduction += bonus.CooldownReduction * float64(ranks)
		total.ManaCostReduction += bonus.ManaCostReduction * float64(ranks)
		total.XPBonusPercent += bonus.XPBonusPercent * float64(ranks)
		total.GoldBonusPercent += bonus.GoldBonusPercent * float64(ranks)
		total.DodgeChanceBonus += bonus.DodgeChanceBonus * float64(ranks)
		total.BlockChanceBonus += bonus.BlockChanceBonus * float64(ranks)
		total.HealingReceivedBonus += bonus.HealingReceivedBonus * float64(ranks)
		total.StatusResistBonus += bonus.StatusResistBonus * float64(ranks)
	}

	return total
}

// applyBonuses applies calculated bonuses to entity components.
func (s *TalentSystem) applyBonuses(entity *Entity, bonuses TalentBonus) {
	s.applyHealthBonuses(entity, bonuses)
	s.applyManaBonuses(entity, bonuses)
	s.applyStatsBonuses(entity, bonuses)
	s.applyCombatBonuses(entity, bonuses)
}

// applyBonusesTracked applies bonuses and returns the absolute deltas that were
// added to each stat.  Using absolute deltas (not ratios) ensures that removal
// is exact even when other systems also modified the stats concurrently (G23).
func (s *TalentSystem) applyBonusesTracked(entity *Entity, bonuses TalentBonus) map[string]float64 {
	deltas := make(map[string]float64)

	if statsComp, ok := entity.GetComponent("stats"); ok {
		if stats, ok := statsComp.(*StatsComponent); ok {
			before := *stats
			s.applyStatsBonuses(entity, bonuses)
			deltas["attack"] = stats.Attack - before.Attack
			deltas["defense"] = stats.Defense - before.Defense
			deltas["magicPower"] = stats.MagicPower - before.MagicPower
			deltas["magicDefense"] = stats.MagicDefense - before.MagicDefense
			deltas["critChance"] = stats.CritChance - before.CritChance
			deltas["critDamage"] = stats.CritDamage - before.CritDamage
			deltas["lifesteal"] = stats.Lifesteal - before.Lifesteal
			deltas["blockChance"] = stats.BlockChance - before.BlockChance
			deltas["evasion"] = stats.Evasion - before.Evasion
		}
	}

	if healthComp, ok := entity.GetComponent("health"); ok {
		if health, ok := healthComp.(*HealthComponent); ok {
			beforeMax := health.Max
			s.applyHealthBonuses(entity, bonuses)
			deltas["healthMax"] = health.Max - beforeMax
		}
	}

	if manaComp, ok := entity.GetComponent("mana"); ok {
		if mana, ok := manaComp.(*ManaComponent); ok {
			beforeMax := mana.Max
			s.applyManaBonuses(entity, bonuses)
			deltas["manaMax"] = float64(mana.Max - beforeMax)
		}
	}

	s.applyCombatBonuses(entity, bonuses)
	return deltas
}

// removeTalentBonuses subtracts previously applied talent bonuses from entity stats.
// Uses the absolute deltas recorded at apply time for exact removal (G23 fix).
func (s *TalentSystem) removeTalentBonuses(entity *Entity, deltas map[string]float64) {
	if len(deltas) == 0 {
		return
	}

	if statsComp, ok := entity.GetComponent("stats"); ok {
		if stats, ok := statsComp.(*StatsComponent); ok {
			stats.Attack -= deltas["attack"]
			stats.Defense -= deltas["defense"]
			stats.MagicPower -= deltas["magicPower"]
			stats.MagicDefense -= deltas["magicDefense"]
			stats.CritChance -= deltas["critChance"]
			stats.CritDamage -= deltas["critDamage"]
			stats.Lifesteal -= deltas["lifesteal"]
			stats.BlockChance -= deltas["blockChance"]
			stats.Evasion -= deltas["evasion"]
		}
	}

	if healthComp, ok := entity.GetComponent("health"); ok {
		if health, ok := healthComp.(*HealthComponent); ok {
			health.Max -= deltas["healthMax"]
		}
	}

	if manaComp, ok := entity.GetComponent("mana"); ok {
		if mana, ok := manaComp.(*ManaComponent); ok {
			mana.Max -= int(deltas["manaMax"])
		}
	}
}

// applyHealthBonuses applies health bonuses from talents.
func (s *TalentSystem) applyHealthBonuses(entity *Entity, bonuses TalentBonus) {
	healthComp, ok := entity.GetComponent("health")
	if !ok {
		return
	}
	health, ok := healthComp.(*HealthComponent)
	if !ok {
		return
	}

	// Apply flat and percentage bonuses
	baseMax := health.Max
	health.Max += bonuses.FlatHealth
	health.Max *= (1.0 + bonuses.HealthPercent)

	// If max health increased, increase current health proportionally
	if health.Max > baseMax && health.Current > 0 {
		ratio := health.Current / baseMax
		health.Current = ratio * health.Max
	}
}

// applyManaBonuses applies mana bonuses from talents.
func (s *TalentSystem) applyManaBonuses(entity *Entity, bonuses TalentBonus) {
	manaComp, ok := entity.GetComponent("mana")
	if !ok {
		return
	}
	mana, ok := manaComp.(*ManaComponent)
	if !ok {
		return
	}

	baseMax := float64(mana.Max)
	newMax := baseMax + bonuses.FlatMana
	newMax *= (1.0 + bonuses.ManaPercent)
	mana.Max = int(newMax)

	if float64(mana.Max) > baseMax && mana.Current > 0 {
		ratio := float64(mana.Current) / baseMax
		mana.Current = int(ratio * float64(mana.Max))
	}
}

// applyStatsBonuses applies stat bonuses from talents.
func (s *TalentSystem) applyStatsBonuses(entity *Entity, bonuses TalentBonus) {
	statsComp, ok := entity.GetComponent("stats")
	if !ok {
		return
	}
	stats, ok := statsComp.(*StatsComponent)
	if !ok {
		return
	}

	// Apply flat bonuses
	stats.Attack += bonuses.FlatDamage
	stats.Defense += bonuses.FlatDefense
	stats.MagicPower += bonuses.FlatMagicPower
	stats.MagicDefense += bonuses.FlatMagicDefense

	// Apply percentage bonuses
	stats.Attack *= (1.0 + bonuses.DamagePercent)
	stats.Defense *= (1.0 + bonuses.DefensePercent)
	stats.MagicPower *= (1.0 + bonuses.MagicPowerPercent)
	stats.MagicDefense *= (1.0 + bonuses.MagicDefensePercent)

	// Apply crit and other combat stats
	stats.CritChance += bonuses.CritChanceBonus
	stats.CritDamage += bonuses.CritDamageBonus
	stats.Lifesteal += bonuses.LifestealPercent
	stats.BlockChance += bonuses.BlockChanceBonus
	stats.Evasion += bonuses.DodgeChanceBonus
}

// applyCombatBonuses applies combat-related bonuses from talents.
func (s *TalentSystem) applyCombatBonuses(entity *Entity, bonuses TalentBonus) {
	// Apply to TalentCombatBonusComponent if present
	combatComp, ok := entity.GetComponent("talent_combat_bonus")
	if !ok {
		// Create one if not present and bonuses warrant it
		if bonuses.CritChanceBonus > 0 || bonuses.CritDamageBonus > 0 ||
			bonuses.LifestealPercent > 0 || bonuses.DodgeChanceBonus > 0 ||
			bonuses.BlockChanceBonus > 0 || bonuses.StatusResistBonus > 0 {
			combatComp = &TalentCombatBonusComponent{}
			entity.AddComponent(combatComp)
		} else {
			return
		}
	}
	combat, ok := combatComp.(*TalentCombatBonusComponent)
	if !ok {
		return
	}

	// Update combat bonuses
	combat.CritChanceBonus = bonuses.CritChanceBonus
	combat.CritDamageBonus = bonuses.CritDamageBonus
	combat.LifestealPercent = bonuses.LifestealPercent
	combat.CooldownReduction = bonuses.CooldownReduction
	combat.ManaCostReduction = bonuses.ManaCostReduction
	combat.XPBonusPercent = bonuses.XPBonusPercent
	combat.GoldBonusPercent = bonuses.GoldBonusPercent
	combat.DodgeChanceBonus = bonuses.DodgeChanceBonus
	combat.BlockChanceBonus = bonuses.BlockChanceBonus
	combat.HealingReceivedBonus = bonuses.HealingReceivedBonus
	combat.StatusResistBonus = bonuses.StatusResistBonus
}

// AllocateTalentPoint allocates a talent point on behalf of an entity.
// Returns true if successful.
func (s *TalentSystem) AllocateTalentPoint(entity *Entity, talentID string) bool {
	talentComp, ok := entity.GetComponent("talent")
	if !ok {
		return false
	}
	talent, ok := talentComp.(*TalentComponent)
	if !ok {
		return false
	}

	def := GetTalentDefinition(talentID)
	if def == nil {
		return false
	}

	// Get character level from experience component
	level := 1
	expComp, ok := entity.GetComponent("experience")
	if ok {
		if exp, ok := expComp.(*ExperienceComponent); ok {
			level = exp.Level
		}
	}

	if talent.AllocatePoint(def, level) {
		if s.onTalentAllocated != nil {
			s.onTalentAllocated(entity, talentID, talent.Allocations[talentID])
		}
		log.WithFields(log.Fields{
			"system_name": "talent",
			"entity_id":   entity.ID,
			"talent_id":   talentID,
			"new_rank":    talent.Allocations[talentID],
		}).Debug("Talent point allocated")
		return true
	}
	return false
}

// DeallocateTalentPoint removes a talent point from an entity.
// Returns true if successful.
func (s *TalentSystem) DeallocateTalentPoint(entity *Entity, talentID string) bool {
	talentComp, ok := entity.GetComponent("talent")
	if !ok {
		return false
	}
	talent, ok := talentComp.(*TalentComponent)
	if !ok {
		return false
	}

	def := GetTalentDefinition(talentID)
	if def == nil {
		return false
	}

	return talent.DeallocatePoint(def)
}

// ResetTalents removes all talent allocations for an entity.
func (s *TalentSystem) ResetTalents(entity *Entity) {
	talentComp, ok := entity.GetComponent("talent")
	if !ok {
		return
	}
	talent, ok := talentComp.(*TalentComponent)
	if !ok {
		return
	}

	talent.ResetAll()

	if s.onTalentReset != nil {
		s.onTalentReset(entity)
	}

	log.WithFields(log.Fields{
		"system_name":   "talent",
		"entity_id":     entity.ID,
		"points_refund": talent.UnspentPoints,
	}).Debug("Talents reset")
}

// SetOnTalentAllocated sets the callback for talent allocation events.
func (s *TalentSystem) SetOnTalentAllocated(callback func(entity *Entity, talentID string, newRank int)) {
	s.onTalentAllocated = callback
}

// SetOnTalentReset sets the callback for talent reset events.
func (s *TalentSystem) SetOnTalentReset(callback func(entity *Entity)) {
	s.onTalentReset = callback
}

// GetTalentBonuses returns the cached talent bonuses for an entity.
func (s *TalentSystem) GetTalentBonuses(entity *Entity) TalentBonus {
	talentComp, ok := entity.GetComponent("talent")
	if !ok {
		return TalentBonus{}
	}
	talent, ok := talentComp.(*TalentComponent)
	if !ok {
		return TalentBonus{}
	}
	return talent.CachedBonuses
}

// TalentCombatBonusComponent stores combat bonuses calculated from talents.
// This component is read by the combat system for damage calculation.
type TalentCombatBonusComponent struct {
	CritChanceBonus      float64
	CritDamageBonus      float64
	LifestealPercent     float64
	CooldownReduction    float64
	ManaCostReduction    float64
	XPBonusPercent       float64
	GoldBonusPercent     float64
	DodgeChanceBonus     float64
	BlockChanceBonus     float64
	HealingReceivedBonus float64
	StatusResistBonus    float64
}

// Type returns the component type identifier.
func (t *TalentCombatBonusComponent) Type() string {
	return "talent_combat_bonus"
}
