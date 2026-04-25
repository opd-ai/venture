// Package engine provides the class affinity system for tracking playstyle progression.
// This system awards affinity XP when players use abilities, building mastery in
// combat archetypes that provide passive bonuses and unlock enhancements.
package engine

import (
	log "github.com/sirupsen/logrus"
)

// ClassAffinitySystem manages class affinity progression for entities.
// As players use abilities, they build XP toward combat archetypes that
// provide passive stat bonuses and ability enhancements.
type ClassAffinitySystem struct {
	world *World

	// XP values for different actions
	xpPerAbilityUse int
	xpPerDamage     int     // Per 100 damage
	xpPerHeal       int     // Per 100 healing
	xpPerKill       int     // Bonus for kills
	xpStreakBonus   float64 // Multiplier for consecutive same-affinity actions
	streakDecayTime float64 // Seconds before streak resets

	// elapsedTime is a monotonic counter incremented each frame (G29 fix).
	elapsedTime float64

	// Update interval for performance
	updateInterval int
	frameCounter   int

	// Callbacks
	onAffinityLevelUp func(entity *Entity, affinity AffinityType, newLevel AffinityLevel)
}

// NewClassAffinitySystem creates a new class affinity system.
func NewClassAffinitySystem(world *World) *ClassAffinitySystem {
	log.WithFields(log.Fields{
		"system_name":     "class_affinity",
		"xp_per_ability":  10,
		"xp_per_damage":   5,
		"xp_streak_bonus": 1.5,
	}).Debug("Creating class affinity system")

	return &ClassAffinitySystem{
		world:           world,
		xpPerAbilityUse: 10,
		xpPerDamage:     5,
		xpPerHeal:       8,
		xpPerKill:       25,
		xpStreakBonus:   1.5,
		streakDecayTime: 5.0, // 5 seconds
		updateInterval:  30,  // Check every 30 frames
		frameCounter:    0,
	}
}

// Update processes class affinity bonus recalculations for dirty entities.
func (s *ClassAffinitySystem) Update(entities []*Entity, deltaTime float64) {
	s.elapsedTime += deltaTime // G29 fix: monotonic timer for streak decay
	s.frameCounter++
	if s.frameCounter < s.updateInterval {
		return
	}
	s.frameCounter = 0

	for _, entity := range entities {
		s.processEntity(entity, deltaTime)
	}
}

// processEntity recalculates affinity bonuses if dirty.
func (s *ClassAffinitySystem) processEntity(entity *Entity, deltaTime float64) {
	affinityComp, hasAffinity := entity.GetComponent("class_affinity")
	if !hasAffinity {
		return
	}
	affinity, ok := affinityComp.(*ClassAffinityComponent)
	if !ok {
		return
	}

	// Decay streaks for all affinities based on time since last activity
	s.decayStreaks(affinity, deltaTime)

	if !affinity.Dirty {
		return
	}

	// Recalculate primary/secondary affinities
	affinity.RecalculatePrimaryAffinities()

	// Apply bonuses for primary affinity
	if affinity.PrimaryAffinity != AffinityNone {
		s.applyAffinityBonuses(entity, affinity, affinity.PrimaryAffinity, 1.0)
	}

	// Apply reduced bonuses for secondary affinity (50% effectiveness)
	if affinity.SecondaryAffinity != AffinityNone {
		s.applyAffinityBonuses(entity, affinity, affinity.SecondaryAffinity, 0.5)
	}

	affinity.Dirty = false
}

// decayStreaks reduces streak counters based on time since last activity.
// G29 fix: uses s.elapsedTime (a monotonic counter incremented in Update) instead
// of the hardcoded 0.0 that caused streaks to never decay.
func (s *ClassAffinitySystem) decayStreaks(affinity *ClassAffinityComponent, _ float64) {
	currentTime := s.elapsedTime

	for _, data := range affinity.Affinities {
		if data.CurrentStreak > 0 {
			timeSinceActivity := currentTime - data.LastActivityTime
			if timeSinceActivity > s.streakDecayTime {
				data.CurrentStreak = 0
			}
		}
	}
}

// applyAffinityBonuses applies stat bonuses for an affinity.
func (s *ClassAffinitySystem) applyAffinityBonuses(entity *Entity, comp *ClassAffinityComponent, affinityType AffinityType, effectiveness float64) {
	data := comp.GetAffinity(affinityType)
	level := data.Level

	// Check if we've already applied this level's bonuses
	if applied, exists := comp.BonusesApplied[affinityType]; exists && applied == level {
		return
	}

	bonuses := GetAffinityBonuses(affinityType, level)

	// Apply to stats component
	statsComp, hasStats := entity.GetComponent("stats")
	if hasStats {
		stats, ok := statsComp.(*StatsComponent)
		if ok {
			// Remove old bonuses first
			if oldLevel, exists := comp.BonusesApplied[affinityType]; exists {
				oldBonuses := GetAffinityBonuses(affinityType, oldLevel)
				stats.CritChance -= oldBonuses.CritChanceBonus * effectiveness
				stats.CritDamage -= oldBonuses.CritDamageBonus * effectiveness
				stats.Evasion -= oldBonuses.EvasionBonus * effectiveness
			}
			// Apply new bonuses
			stats.CritChance += bonuses.CritChanceBonus * effectiveness
			stats.CritDamage += bonuses.CritDamageBonus * effectiveness
			stats.Evasion += bonuses.EvasionBonus * effectiveness
		}
	}

	// Apply mana regen bonus
	manaComp, hasMana := entity.GetComponent("mana")
	if hasMana {
		mana, ok := manaComp.(*ManaComponent)
		if ok {
			// G37 fix: remove the stored absolute regen value rather than
			// recomputing it from the (possibly changed) mana.Max, which
			// would remove a different amount than was originally added.
			if comp.AppliedManaRegen == nil {
				comp.AppliedManaRegen = make(map[AffinityType]float64)
			}
			if prev, exists := comp.AppliedManaRegen[affinityType]; exists {
				mana.Regen -= prev
			}
			// Apply new bonus and store the absolute value for future removal.
			newRegen := bonuses.ManaRegenBonus * float64(mana.Max) * effectiveness
			mana.Regen += newRegen
			comp.AppliedManaRegen[affinityType] = newRegen
		}
	}

	// Track that we've applied this level
	comp.BonusesApplied[affinityType] = level

	log.WithFields(log.Fields{
		"system_name":   "class_affinity",
		"entity_id":     entity.ID,
		"affinity":      affinityType.String(),
		"level":         level.String(),
		"effectiveness": effectiveness,
	}).Debug("Applied affinity bonuses")
}

// OnAbilityUsed should be called when an entity uses an ability.
// This awards affinity XP based on the ability's associated archetypes.
func (s *ClassAffinitySystem) OnAbilityUsed(entity *Entity, abilityID string, damage, healing float64) {
	affinityComp, hasAffinity := entity.GetComponent("class_affinity")
	if !hasAffinity {
		return
	}
	affinity, ok := affinityComp.(*ClassAffinityComponent)
	if !ok {
		return
	}

	// Get affinities associated with this ability
	associatedAffinities := GetAbilityAffinities(abilityID)
	if len(associatedAffinities) == 0 {
		return
	}

	for _, affinityType := range associatedAffinities {
		s.awardAffinityXP(entity, affinity, affinityType, abilityID, damage, healing)
	}
}

// awardAffinityXP adds XP to an affinity and checks for level-ups.
func (s *ClassAffinitySystem) awardAffinityXP(entity *Entity, affinity *ClassAffinityComponent, affinityType AffinityType, abilityID string, damage, healing float64) {
	data := affinity.GetAffinity(affinityType)
	oldLevel := data.Level

	// Calculate base XP
	xpAmount := s.xpPerAbilityUse

	// Add XP for damage dealt (per 100 damage)
	if damage > 0 {
		xpAmount += int(damage/100) * s.xpPerDamage
	}

	// Add XP for healing done (per 100 healing)
	if healing > 0 {
		xpAmount += int(healing/100) * s.xpPerHeal
	}

	// Apply streak bonus
	if data.CurrentStreak > 0 {
		streakMultiplier := 1.0 + (float64(data.CurrentStreak) * 0.1) // +10% per streak
		if streakMultiplier > s.xpStreakBonus {
			streakMultiplier = s.xpStreakBonus
		}
		xpAmount = int(float64(xpAmount) * streakMultiplier)
	}

	// Update data
	data.XP += xpAmount
	data.AbilitiesUsed++
	data.DamageDealt += damage
	data.CurrentStreak++
	if data.CurrentStreak > data.PeakStreak {
		data.PeakStreak = data.CurrentStreak
	}
	data.LastActivityTime = s.elapsedTime // G29 fix: record monotonic elapsed time

	affinity.TotalAffinityXP += xpAmount

	// Check for level up
	newLevel := s.calculateAffinityLevel(data.XP)
	if newLevel > oldLevel {
		data.Level = newLevel
		affinity.Dirty = true

		log.WithFields(log.Fields{
			"system_name": "class_affinity",
			"entity_id":   entity.ID,
			"affinity":    affinityType.String(),
			"old_level":   oldLevel.String(),
			"new_level":   newLevel.String(),
			"total_xp":    data.XP,
		}).Info("Affinity level up!")

		// Trigger callback
		if s.onAffinityLevelUp != nil {
			s.onAffinityLevelUp(entity, affinityType, newLevel)
		}
	}
}

// calculateAffinityLevel determines affinity level from XP.
func (s *ClassAffinitySystem) calculateAffinityLevel(xp int) AffinityLevel {
	if xp >= AffinityLevelGrandmaster.XPThreshold() {
		return AffinityLevelGrandmaster
	}
	if xp >= AffinityLevelMaster.XPThreshold() {
		return AffinityLevelMaster
	}
	if xp >= AffinityLevelExpert.XPThreshold() {
		return AffinityLevelExpert
	}
	if xp >= AffinityLevelJourneyman.XPThreshold() {
		return AffinityLevelJourneyman
	}
	if xp >= AffinityLevelApprentice.XPThreshold() {
		return AffinityLevelApprentice
	}
	if xp >= AffinityLevelNovice.XPThreshold() {
		return AffinityLevelNovice
	}
	return AffinityLevelNone
}

// OnKill awards bonus XP for kills to relevant affinities.
func (s *ClassAffinitySystem) OnKill(entity *Entity, lastAbilityUsed string) {
	affinityComp, hasAffinity := entity.GetComponent("class_affinity")
	if !hasAffinity {
		return
	}
	affinity, ok := affinityComp.(*ClassAffinityComponent)
	if !ok {
		return
	}

	// Award bonus XP to affinities associated with the killing blow
	associatedAffinities := GetAbilityAffinities(lastAbilityUsed)
	for _, affinityType := range associatedAffinities {
		data := affinity.GetAffinity(affinityType)
		data.XP += s.xpPerKill
		data.TimesTriggered++
		affinity.TotalAffinityXP += s.xpPerKill

		// Check for level up
		newLevel := s.calculateAffinityLevel(data.XP)
		if newLevel > data.Level {
			data.Level = newLevel
			affinity.Dirty = true

			if s.onAffinityLevelUp != nil {
				s.onAffinityLevelUp(entity, affinityType, newLevel)
			}
		}
	}
}

// GetDamageMultiplier returns the damage multiplier from affinity bonuses.
func (s *ClassAffinitySystem) GetDamageMultiplier(entity *Entity) float64 {
	affinityComp, hasAffinity := entity.GetComponent("class_affinity")
	if !hasAffinity {
		return 1.0
	}
	affinity, ok := affinityComp.(*ClassAffinityComponent)
	if !ok {
		return 1.0
	}

	multiplier := 1.0

	// Primary affinity at full strength
	if affinity.PrimaryAffinity != AffinityNone {
		data := affinity.GetAffinity(affinity.PrimaryAffinity)
		bonuses := GetAffinityBonuses(affinity.PrimaryAffinity, data.Level)
		multiplier *= bonuses.DamageMultiplier
	}

	// Secondary affinity at half strength
	if affinity.SecondaryAffinity != AffinityNone {
		data := affinity.GetAffinity(affinity.SecondaryAffinity)
		bonuses := GetAffinityBonuses(affinity.SecondaryAffinity, data.Level)
		// Apply 50% of the bonus portion (multiplier - 1.0)
		bonusPortion := (bonuses.DamageMultiplier - 1.0) * 0.5
		multiplier += bonusPortion
	}

	return multiplier
}

// GetSpellDamageMultiplier returns the spell damage multiplier from affinity bonuses.
func (s *ClassAffinitySystem) GetSpellDamageMultiplier(entity *Entity) float64 {
	affinityComp, hasAffinity := entity.GetComponent("class_affinity")
	if !hasAffinity {
		return 1.0
	}
	affinity, ok := affinityComp.(*ClassAffinityComponent)
	if !ok {
		return 1.0
	}

	multiplier := 1.0

	// Primary affinity at full strength
	if affinity.PrimaryAffinity != AffinityNone {
		data := affinity.GetAffinity(affinity.PrimaryAffinity)
		bonuses := GetAffinityBonuses(affinity.PrimaryAffinity, data.Level)
		multiplier *= bonuses.SpellDamageMult
	}

	// Secondary affinity at half strength
	if affinity.SecondaryAffinity != AffinityNone {
		data := affinity.GetAffinity(affinity.SecondaryAffinity)
		bonuses := GetAffinityBonuses(affinity.SecondaryAffinity, data.Level)
		bonusPortion := (bonuses.SpellDamageMult - 1.0) * 0.5
		multiplier += bonusPortion
	}

	return multiplier
}

// GetHealingMultiplier returns the healing multiplier from affinity bonuses.
func (s *ClassAffinitySystem) GetHealingMultiplier(entity *Entity) float64 {
	affinityComp, hasAffinity := entity.GetComponent("class_affinity")
	if !hasAffinity {
		return 1.0
	}
	affinity, ok := affinityComp.(*ClassAffinityComponent)
	if !ok {
		return 1.0
	}

	multiplier := 1.0

	if affinity.PrimaryAffinity != AffinityNone {
		data := affinity.GetAffinity(affinity.PrimaryAffinity)
		bonuses := GetAffinityBonuses(affinity.PrimaryAffinity, data.Level)
		multiplier *= bonuses.HealingMultiplier
	}

	if affinity.SecondaryAffinity != AffinityNone {
		data := affinity.GetAffinity(affinity.SecondaryAffinity)
		bonuses := GetAffinityBonuses(affinity.SecondaryAffinity, data.Level)
		bonusPortion := (bonuses.HealingMultiplier - 1.0) * 0.5
		multiplier += bonusPortion
	}

	return multiplier
}

// GetCooldownReduction returns the cooldown reduction from affinity bonuses.
func (s *ClassAffinitySystem) GetCooldownReduction(entity *Entity) float64 {
	affinityComp, hasAffinity := entity.GetComponent("class_affinity")
	if !hasAffinity {
		return 0.0
	}
	affinity, ok := affinityComp.(*ClassAffinityComponent)
	if !ok {
		return 0.0
	}

	reduction := 0.0

	if affinity.PrimaryAffinity != AffinityNone {
		data := affinity.GetAffinity(affinity.PrimaryAffinity)
		bonuses := GetAffinityBonuses(affinity.PrimaryAffinity, data.Level)
		reduction += bonuses.CooldownReduce
	}

	if affinity.SecondaryAffinity != AffinityNone {
		data := affinity.GetAffinity(affinity.SecondaryAffinity)
		bonuses := GetAffinityBonuses(affinity.SecondaryAffinity, data.Level)
		reduction += bonuses.CooldownReduce * 0.5
	}

	// Cap at 40% cooldown reduction
	if reduction > 0.40 {
		reduction = 0.40
	}

	return reduction
}

// SetOnAffinityLevelUp sets a callback for when affinity level increases.
func (s *ClassAffinitySystem) SetOnAffinityLevelUp(callback func(entity *Entity, affinity AffinityType, newLevel AffinityLevel)) {
	s.onAffinityLevelUp = callback
}

// EnsureClassAffinityComponent adds a ClassAffinityComponent to an entity if missing.
func EnsureClassAffinityComponent(entity *Entity) *ClassAffinityComponent {
	if entity == nil {
		return nil
	}

	affinityComp, hasAffinity := entity.GetComponent("class_affinity")
	if hasAffinity {
		if affinity, ok := affinityComp.(*ClassAffinityComponent); ok {
			return affinity
		}
	}

	newComp := NewClassAffinityComponent()
	entity.AddComponent(newComp)
	return newComp
}

// GetAffinityStatusForUI returns display-friendly affinity info for UI.
func (s *ClassAffinitySystem) GetAffinityStatusForUI(entity *Entity, affinityType AffinityType) *AffinityStatus {
	affinityComp, hasAffinity := entity.GetComponent("class_affinity")
	if !hasAffinity {
		return nil
	}
	affinity, ok := affinityComp.(*ClassAffinityComponent)
	if !ok {
		return nil
	}

	data := affinity.GetAffinity(affinityType)
	bonuses := GetAffinityBonuses(affinityType, data.Level)

	return &AffinityStatus{
		AffinityType:      affinityType,
		Name:              affinityType.String(),
		Description:       affinityType.Description(),
		Level:             data.Level,
		LevelName:         data.Level.String(),
		CurrentXP:         data.XP,
		XPToNextLevel:     affinity.GetXPToNextLevel(affinityType),
		Progress:          affinity.GetProgressToNextLevel(affinityType),
		AbilitiesUsed:     data.AbilitiesUsed,
		TotalDamage:       data.DamageDealt,
		CurrentStreak:     data.CurrentStreak,
		PeakStreak:        data.PeakStreak,
		DamageBonus:       (bonuses.DamageMultiplier - 1.0) * 100,
		SpellDamageBonus:  (bonuses.SpellDamageMult - 1.0) * 100,
		HealingBonus:      (bonuses.HealingMultiplier - 1.0) * 100,
		CritChanceBonus:   bonuses.CritChanceBonus * 100,
		CritDamageBonus:   bonuses.CritDamageBonus * 100,
		CooldownReduction: bonuses.CooldownReduce * 100,
		IsPrimary:         affinityType == affinity.PrimaryAffinity,
		IsSecondary:       affinityType == affinity.SecondaryAffinity,
	}
}

// AffinityStatus contains UI-friendly affinity information.
type AffinityStatus struct {
	AffinityType      AffinityType
	Name              string
	Description       string
	Level             AffinityLevel
	LevelName         string
	CurrentXP         int
	XPToNextLevel     int
	Progress          float64
	AbilitiesUsed     int
	TotalDamage       float64
	CurrentStreak     int
	PeakStreak        int
	DamageBonus       float64 // As percentage
	SpellDamageBonus  float64
	HealingBonus      float64
	CritChanceBonus   float64
	CritDamageBonus   float64
	CooldownReduction float64
	IsPrimary         bool
	IsSecondary       bool
}

// GetAllAffinities returns status for all affinities an entity has developed.
func (s *ClassAffinitySystem) GetAllAffinities(entity *Entity) []*AffinityStatus {
	affinityComp, hasAffinity := entity.GetComponent("class_affinity")
	if !hasAffinity {
		return nil
	}
	affinity, ok := affinityComp.(*ClassAffinityComponent)
	if !ok {
		return nil
	}

	result := make([]*AffinityStatus, 0, len(affinity.Affinities))
	for affinityType := range affinity.Affinities {
		status := s.GetAffinityStatusForUI(entity, affinityType)
		if status != nil && status.CurrentXP > 0 {
			result = append(result, status)
		}
	}

	return result
}

// GetPrimaryAffinityName returns the name of the entity's primary affinity.
func (s *ClassAffinitySystem) GetPrimaryAffinityName(entity *Entity) string {
	affinityComp, hasAffinity := entity.GetComponent("class_affinity")
	if !hasAffinity {
		return "None"
	}
	affinity, ok := affinityComp.(*ClassAffinityComponent)
	if !ok {
		return "None"
	}

	return affinity.PrimaryAffinity.String()
}
