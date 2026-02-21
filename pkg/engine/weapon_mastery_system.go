// Package engine provides the weapon mastery system for tracking weapon proficiency.
// This system awards mastery XP on attacks and applies damage/speed bonuses.
package engine

import (
	"github.com/opd-ai/venture/pkg/procgen/item"
	log "github.com/sirupsen/logrus"
)

// WeaponMasteryBonuses defines the stat bonuses granted at each mastery level.
type WeaponMasteryBonuses struct {
	// DamageMultiplier is the damage bonus (1.0 = no bonus, 1.1 = +10%)
	DamageMultiplier float64
	// AttackSpeedBonus is the attack speed bonus (added to base speed)
	AttackSpeedBonus float64
	// CritChanceBonus is the critical hit chance bonus (0.05 = +5%)
	CritChanceBonus float64
	// CritDamageBonus is the critical damage multiplier bonus
	CritDamageBonus float64
}

// masteryBonusTable maps mastery levels to their bonuses.
var masteryBonusTable = map[WeaponMasteryLevel]WeaponMasteryBonuses{
	MasteryNovice: {
		DamageMultiplier: 1.00,
		AttackSpeedBonus: 0.00,
		CritChanceBonus:  0.00,
		CritDamageBonus:  0.00,
	},
	MasteryApprentice: {
		DamageMultiplier: 1.05,
		AttackSpeedBonus: 0.05,
		CritChanceBonus:  0.02,
		CritDamageBonus:  0.05,
	},
	MasteryJourneyman: {
		DamageMultiplier: 1.10,
		AttackSpeedBonus: 0.10,
		CritChanceBonus:  0.04,
		CritDamageBonus:  0.10,
	},
	MasteryExpert: {
		DamageMultiplier: 1.18,
		AttackSpeedBonus: 0.15,
		CritChanceBonus:  0.06,
		CritDamageBonus:  0.18,
	},
	MasteryMaster: {
		DamageMultiplier: 1.28,
		AttackSpeedBonus: 0.22,
		CritChanceBonus:  0.08,
		CritDamageBonus:  0.28,
	},
	MasteryGrandmaster: {
		DamageMultiplier: 1.40,
		AttackSpeedBonus: 0.30,
		CritChanceBonus:  0.10,
		CritDamageBonus:  0.40,
	},
}

// GetMasteryBonuses returns the bonuses for a mastery level.
func GetMasteryBonuses(level WeaponMasteryLevel) WeaponMasteryBonuses {
	if bonuses, exists := masteryBonusTable[level]; exists {
		return bonuses
	}
	return masteryBonusTable[MasteryNovice]
}

// WeaponMasterySystem manages weapon proficiency progression.
// Awards XP on successful attacks and applies mastery bonuses to combat.
type WeaponMasterySystem struct {
	world *World

	// XP values for different actions
	xpPerHit      int
	xpPerKill     int
	xpPerCritical int

	// Cache for performance
	updateInterval int
	frameCounter   int

	// Callbacks for mastery level-ups
	onMasteryLevelUp func(entity *Entity, weaponType string, newLevel WeaponMasteryLevel)
}

// NewWeaponMasterySystem creates a new weapon mastery system.
func NewWeaponMasterySystem(world *World) *WeaponMasterySystem {
	log.WithFields(log.Fields{
		"system_name": "weapon_mastery",
		"xp_per_hit":  5,
		"xp_per_kill": 25,
		"xp_per_crit": 10,
	}).Debug("Creating weapon mastery system")

	return &WeaponMasterySystem{
		world:          world,
		xpPerHit:       5,
		xpPerKill:      25,
		xpPerCritical:  10,
		updateInterval: 30, // Check dirty flags every 30 frames
		frameCounter:   0,
	}
}

// Update processes mastery bonus recalculations for dirty entities.
func (s *WeaponMasterySystem) Update(entities []*Entity, deltaTime float64) {
	s.frameCounter++
	if s.frameCounter < s.updateInterval {
		return
	}
	s.frameCounter = 0

	for _, entity := range entities {
		s.processEntity(entity)
	}
}

// processEntity recalculates mastery bonuses if dirty.
func (s *WeaponMasterySystem) processEntity(entity *Entity) {
	masteryComp, hasMastery := entity.GetComponent("weapon_mastery")
	if !hasMastery {
		return
	}
	mastery, ok := masteryComp.(*WeaponMasteryComponent)
	if !ok || !mastery.Dirty {
		return
	}

	// Get currently equipped weapon
	weaponType := s.getEquippedWeaponType(entity)
	if weaponType == "" {
		mastery.Dirty = false
		return
	}

	// Apply bonuses for current weapon's mastery level
	level := mastery.GetMasteryLevel(weaponType)
	s.applyMasteryBonuses(entity, weaponType, level)

	mastery.Dirty = false
}

// getEquippedWeaponType returns the weapon type string of the equipped main hand weapon.
func (s *WeaponMasterySystem) getEquippedWeaponType(entity *Entity) string {
	return s.GetEquippedWeaponType(entity)
}

// GetEquippedWeaponType returns the weapon type string of the equipped main hand weapon.
// Exported for callback integration from other systems.
func (s *WeaponMasterySystem) GetEquippedWeaponType(entity *Entity) string {
	equipComp, hasEquip := entity.GetComponent("equipment")
	if !hasEquip {
		return ""
	}
	equip, ok := equipComp.(*EquipmentComponent)
	if !ok {
		return ""
	}

	mainHand := equip.GetEquipped(SlotMainHand)
	if mainHand == nil || mainHand.Type != item.TypeWeapon {
		return ""
	}

	return mainHand.WeaponType.String()
}

// applyMasteryBonuses applies stat bonuses based on mastery level to equipped weapon.
func (s *WeaponMasterySystem) applyMasteryBonuses(entity *Entity, weaponType string, level WeaponMasteryLevel) {
	// Get the mastery component to track applied level
	masteryComp, _ := entity.GetComponent("weapon_mastery")
	mastery := masteryComp.(*WeaponMasteryComponent)

	// Check if we've already applied this level's bonuses
	if applied, exists := mastery.BonusesApplied[weaponType]; exists && applied == level {
		return // Already applied
	}

	bonuses := GetMasteryBonuses(level)

	// Apply to stats component
	if statsComp, hasStats := entity.GetComponent("stats"); hasStats {
		if stats, ok := statsComp.(*StatsComponent); ok {
			// Remove old bonuses first (if any)
			if oldLevel, exists := mastery.BonusesApplied[weaponType]; exists {
				oldBonuses := GetMasteryBonuses(oldLevel)
				stats.CritChance -= oldBonuses.CritChanceBonus
				stats.CritDamage -= oldBonuses.CritDamageBonus
			}
			// Apply new bonuses
			stats.CritChance += bonuses.CritChanceBonus
			stats.CritDamage += bonuses.CritDamageBonus
		}
	}

	// Apply attack speed bonus
	if attackComp, hasAttack := entity.GetComponent("attack"); hasAttack {
		if attack, ok := attackComp.(*AttackComponent); ok {
			// Remove old attack speed bonus
			if oldLevel, exists := mastery.BonusesApplied[weaponType]; exists {
				oldBonuses := GetMasteryBonuses(oldLevel)
				attack.Cooldown /= (1.0 - oldBonuses.AttackSpeedBonus)
			}
			// Apply new attack speed bonus (lower cooldown = faster attacks)
			attack.Cooldown *= (1.0 - bonuses.AttackSpeedBonus)
		}
	}

	// Track that we've applied this level
	mastery.BonusesApplied[weaponType] = level

	log.WithFields(log.Fields{
		"system_name":   "weapon_mastery",
		"entity_id":     entity.ID,
		"weapon_type":   weaponType,
		"mastery_level": level.String(),
		"damage_mult":   bonuses.DamageMultiplier,
		"attack_speed":  bonuses.AttackSpeedBonus,
		"crit_chance":   bonuses.CritChanceBonus,
	}).Debug("Applied mastery bonuses")
}

// AwardMasteryXP adds XP to a weapon type's mastery and checks for level-ups.
// This should be called by the combat system on successful hits.
func (s *WeaponMasterySystem) AwardMasteryXP(entity *Entity, weaponType string, xpAmount int) {
	masteryComp, hasMastery := entity.GetComponent("weapon_mastery")
	if !hasMastery {
		return
	}
	mastery, ok := masteryComp.(*WeaponMasteryComponent)
	if !ok {
		return
	}

	data := mastery.GetMastery(weaponType)
	oldLevel := data.Level
	data.XP += xpAmount
	mastery.TotalMasteryXP += xpAmount

	// Check for level up
	newLevel := s.calculateMasteryLevel(data.XP)
	if newLevel > oldLevel {
		data.Level = newLevel
		mastery.Dirty = true // Trigger bonus recalculation

		log.WithFields(log.Fields{
			"system_name": "weapon_mastery",
			"entity_id":   entity.ID,
			"weapon_type": weaponType,
			"old_level":   oldLevel.String(),
			"new_level":   newLevel.String(),
			"total_xp":    data.XP,
		}).Info("Weapon mastery level up!")

		// Trigger callback if set
		if s.onMasteryLevelUp != nil {
			s.onMasteryLevelUp(entity, weaponType, newLevel)
		}
	}
}

// calculateMasteryLevel determines mastery level from XP.
func (s *WeaponMasterySystem) calculateMasteryLevel(xp int) WeaponMasteryLevel {
	if xp >= MasteryGrandmaster.XPThreshold() {
		return MasteryGrandmaster
	}
	if xp >= MasteryMaster.XPThreshold() {
		return MasteryMaster
	}
	if xp >= MasteryExpert.XPThreshold() {
		return MasteryExpert
	}
	if xp >= MasteryJourneyman.XPThreshold() {
		return MasteryJourneyman
	}
	if xp >= MasteryApprentice.XPThreshold() {
		return MasteryApprentice
	}
	return MasteryNovice
}

// OnHit should be called when an entity lands a hit with a weapon.
func (s *WeaponMasterySystem) OnHit(entity *Entity, weaponType string, damage float64) {
	masteryComp, hasMastery := entity.GetComponent("weapon_mastery")
	if !hasMastery {
		return
	}
	mastery, ok := masteryComp.(*WeaponMasteryComponent)
	if !ok {
		return
	}

	data := mastery.GetMastery(weaponType)
	data.TotalDamage += damage

	s.AwardMasteryXP(entity, weaponType, s.xpPerHit)
}

// OnKill should be called when an entity kills a target with a weapon.
func (s *WeaponMasterySystem) OnKill(entity *Entity, weaponType string) {
	masteryComp, hasMastery := entity.GetComponent("weapon_mastery")
	if !hasMastery {
		return
	}
	mastery, ok := masteryComp.(*WeaponMasteryComponent)
	if !ok {
		return
	}

	data := mastery.GetMastery(weaponType)
	data.TotalKills++

	s.AwardMasteryXP(entity, weaponType, s.xpPerKill)
}

// OnCriticalHit should be called when an entity lands a critical hit.
func (s *WeaponMasterySystem) OnCriticalHit(entity *Entity, weaponType string) {
	masteryComp, hasMastery := entity.GetComponent("weapon_mastery")
	if !hasMastery {
		return
	}
	mastery, ok := masteryComp.(*WeaponMasteryComponent)
	if !ok {
		return
	}

	data := mastery.GetMastery(weaponType)
	data.CriticalHits++

	s.AwardMasteryXP(entity, weaponType, s.xpPerCritical)
}

// GetDamageMultiplier returns the damage multiplier for the equipped weapon's mastery.
// Used by combat system to scale damage based on weapon proficiency.
func (s *WeaponMasterySystem) GetDamageMultiplier(entity *Entity) float64 {
	weaponType := s.getEquippedWeaponType(entity)
	if weaponType == "" {
		return 1.0
	}

	masteryComp, hasMastery := entity.GetComponent("weapon_mastery")
	if !hasMastery {
		return 1.0
	}
	mastery, ok := masteryComp.(*WeaponMasteryComponent)
	if !ok {
		return 1.0
	}

	level := mastery.GetMasteryLevel(weaponType)
	return GetMasteryBonuses(level).DamageMultiplier
}

// SetOnMasteryLevelUp sets a callback for when mastery level increases.
// Used for visual/audio feedback on level-ups.
func (s *WeaponMasterySystem) SetOnMasteryLevelUp(callback func(entity *Entity, weaponType string, newLevel WeaponMasteryLevel)) {
	s.onMasteryLevelUp = callback
}

// EnsureWeaponMasteryComponent adds a WeaponMasteryComponent to an entity if missing.
func EnsureWeaponMasteryComponent(entity *Entity) *WeaponMasteryComponent {
	if entity == nil {
		return nil
	}

	masteryComp, hasMastery := entity.GetComponent("weapon_mastery")
	if hasMastery {
		if mastery, ok := masteryComp.(*WeaponMasteryComponent); ok {
			return mastery
		}
	}

	newComp := NewWeaponMasteryComponent()
	entity.AddComponent(newComp)
	return newComp
}

// GetMasteryStatusForUI returns display-friendly mastery info for UI.
func (s *WeaponMasterySystem) GetMasteryStatusForUI(entity *Entity) *MasteryStatus {
	weaponType := s.getEquippedWeaponType(entity)
	if weaponType == "" {
		return nil
	}

	masteryComp, hasMastery := entity.GetComponent("weapon_mastery")
	if !hasMastery {
		return nil
	}
	mastery, ok := masteryComp.(*WeaponMasteryComponent)
	if !ok {
		return nil
	}

	data := mastery.GetMastery(weaponType)
	bonuses := GetMasteryBonuses(data.Level)

	return &MasteryStatus{
		WeaponType:       weaponType,
		Level:            data.Level,
		LevelName:        data.Level.String(),
		CurrentXP:        data.XP,
		XPToNextLevel:    mastery.GetXPToNextLevel(weaponType),
		Progress:         mastery.GetProgressToNextLevel(weaponType),
		DamageBonus:      (bonuses.DamageMultiplier - 1.0) * 100, // As percentage
		AttackSpeedBonus: bonuses.AttackSpeedBonus * 100,
		CritChanceBonus:  bonuses.CritChanceBonus * 100,
		TotalKills:       data.TotalKills,
		TotalDamage:      data.TotalDamage,
		CriticalHits:     data.CriticalHits,
		TotalMasteryXP:   mastery.TotalMasteryXP,
	}
}

// MasteryStatus contains UI-friendly mastery information.
type MasteryStatus struct {
	WeaponType       string
	Level            WeaponMasteryLevel
	LevelName        string
	CurrentXP        int
	XPToNextLevel    int
	Progress         float64
	DamageBonus      float64 // As percentage (e.g., 10.0 = +10%)
	AttackSpeedBonus float64 // As percentage
	CritChanceBonus  float64 // As percentage
	TotalKills       int
	TotalDamage      float64
	CriticalHits     int
	TotalMasteryXP   int
}

// GetAllMasteryLevels returns mastery data for all weapon types an entity has used.
func (s *WeaponMasterySystem) GetAllMasteryLevels(entity *Entity) map[string]*MasteryStatus {
	masteryComp, hasMastery := entity.GetComponent("weapon_mastery")
	if !hasMastery {
		return nil
	}
	mastery, ok := masteryComp.(*WeaponMasteryComponent)
	if !ok {
		return nil
	}

	result := make(map[string]*MasteryStatus)
	for weaponType, data := range mastery.Mastery {
		bonuses := GetMasteryBonuses(data.Level)
		result[weaponType] = &MasteryStatus{
			WeaponType:       weaponType,
			Level:            data.Level,
			LevelName:        data.Level.String(),
			CurrentXP:        data.XP,
			XPToNextLevel:    mastery.GetXPToNextLevel(weaponType),
			Progress:         mastery.GetProgressToNextLevel(weaponType),
			DamageBonus:      (bonuses.DamageMultiplier - 1.0) * 100,
			AttackSpeedBonus: bonuses.AttackSpeedBonus * 100,
			CritChanceBonus:  bonuses.CritChanceBonus * 100,
			TotalKills:       data.TotalKills,
			TotalDamage:      data.TotalDamage,
			CriticalHits:     data.CriticalHits,
		}
	}
	return result
}
