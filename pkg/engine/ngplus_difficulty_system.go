// Package engine provides the NG+ Difficulty system for ECS.
// This file implements NGPlusDifficultySystem which applies difficulty scaling
// to entities based on the current NG+ cycle.
//
// Phase 113: Difficulty Scaling System
package engine

import (
	"github.com/sirupsen/logrus"
)

// NGPlusDifficultySystem applies NG+ difficulty scaling to entities.
// It monitors the player's NG+ level and applies appropriate scaling to
// enemies, loot drops, and experience gains.
type NGPlusDifficultySystem struct {
	world  *World
	logger *logrus.Entry

	// currentNGPlusCycle caches the player's NG+ level for efficiency
	currentNGPlusCycle int

	// callbacks for difficulty events
	onEnemyScaled   func(entityID uint64, cycle int)
	onLootGenerated func(entityID uint64, bonusQuality float64)

	// scalingEnabled allows disabling scaling for testing
	scalingEnabled bool
}

// NewNGPlusDifficultySystem creates a new difficulty scaling system.
func NewNGPlusDifficultySystem(world *World) *NGPlusDifficultySystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system", "ngplus_difficulty")
		logEntry.Debug("NG+ Difficulty system created")
	}
	return &NGPlusDifficultySystem{
		world:              world,
		logger:             logEntry,
		currentNGPlusCycle: 0,
		scalingEnabled:     true,
	}
}

// Update processes all entities, applying difficulty scaling where needed.
func (s *NGPlusDifficultySystem) Update(entities []*Entity, deltaTime float64) {
	if !s.scalingEnabled {
		return
	}

	// First, find the player's NG+ level
	s.updateCurrentNGPlusCycle(entities)

	// Then apply scaling to entities that need it
	for _, entity := range entities {
		s.processEntity(entity)
	}
}

// updateCurrentNGPlusCycle finds and caches the player's NG+ level.
func (s *NGPlusDifficultySystem) updateCurrentNGPlusCycle(entities []*Entity) {
	for _, entity := range entities {
		ngpComp, ok := entity.GetComponent("newgameplus")
		if !ok {
			continue
		}
		ngp, ok := ngpComp.(*NewGamePlusComponent)
		if !ok {
			continue
		}
		// Found a player with NG+ component
		s.currentNGPlusCycle = ngp.GetCycle()
		return
	}
}

// processEntity applies difficulty scaling to a single entity if needed.
func (s *NGPlusDifficultySystem) processEntity(entity *Entity) {
	// Check if entity already has difficulty component
	diffComp, hasDiff := entity.GetComponent("ngplus_difficulty")
	if hasDiff {
		diff, ok := diffComp.(*NGPlusDifficultyComponent)
		if ok && diff.IsScaled {
			// Already scaled, skip
			return
		}
	}

	// Check if this is an enemy (has team component and is not player team)
	teamComp, hasTeam := entity.GetComponent("team")
	if !hasTeam {
		return
	}
	team, ok := teamComp.(*TeamComponent)
	if !ok || team.TeamID == 1 { // Team 1 is player
		return
	}

	// Check if entity has health (only scale combat entities)
	_, hasHealth := entity.GetComponent("health")
	if !hasHealth {
		return
	}

	// Apply scaling
	s.applyDifficultyScaling(entity)
}

// applyDifficultyScaling applies NG+ difficulty modifiers to an entity.
func (s *NGPlusDifficultySystem) applyDifficultyScaling(entity *Entity) {
	if s.currentNGPlusCycle <= 0 {
		return // No scaling needed for first playthrough
	}

	// Check if already scaled
	if existingComp, ok := entity.GetComponent("ngplus_difficulty"); ok {
		if dc, ok := existingComp.(*NGPlusDifficultyComponent); ok {
			if dc.IsScaled {
				return // Already scaled, skip
			}
		}
	}

	// Create or get difficulty component
	var diffComp *NGPlusDifficultyComponent
	if existingComp, ok := entity.GetComponent("ngplus_difficulty"); ok {
		if dc, ok := existingComp.(*NGPlusDifficultyComponent); ok {
			diffComp = dc
		}
	}
	if diffComp == nil {
		diffComp = NewNGPlusDifficultyComponent()
		entity.AddComponent(diffComp)
	}

	// Apply scaling based on current NG+ cycle
	diffComp.ApplyScalingForCycle(s.currentNGPlusCycle)

	// Check if this is a boss (has boss_tag component or has BossBehavior)
	if s.isBossEntity(entity) {
		diffComp.ApplyBossEnhancements(s.currentNGPlusCycle)
	}

	// Scale the actual stats
	s.scaleEntityStats(entity, diffComp)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":         entity.ID,
			"ngplus_cycle":      s.currentNGPlusCycle,
			"health_multiplier": diffComp.GetHealthMultiplier(),
			"damage_multiplier": diffComp.GetDamageMultiplier(),
		}).Debug("Applied NG+ difficulty scaling")
	}

	// Notify callback
	if s.onEnemyScaled != nil {
		s.onEnemyScaled(entity.ID, s.currentNGPlusCycle)
	}
}

// isBossEntity checks if an entity is a boss.
func (s *NGPlusDifficultySystem) isBossEntity(entity *Entity) bool {
	// Check for boss_tag component
	if _, hasBossTag := entity.GetComponent("boss_tag"); hasBossTag {
		return true
	}

	// Check behavior tree for boss tree name
	if btComp, hasBT := entity.GetComponent("behavior_tree"); hasBT {
		if bt, ok := btComp.(*BehaviorTreeComponent); ok {
			// Check if the behavior tree has boss-like name
			if bt.TreeName == "boss" || bt.TreeName == "boss_behavior" {
				return true
			}
		}
	}

	// Check for high base health as a heuristic
	if healthComp, hasHealth := entity.GetComponent("health"); hasHealth {
		if health, ok := healthComp.(*HealthComponent); ok {
			// Bosses typically have >500 base health
			if health.Max >= 500 {
				return true
			}
		}
	}

	return false
}

// scaleEntityStats modifies the entity's actual stat values based on difficulty.
func (s *NGPlusDifficultySystem) scaleEntityStats(entity *Entity, diff *NGPlusDifficultyComponent) {
	// Scale health
	if healthComp, ok := entity.GetComponent("health"); ok {
		if health, ok := healthComp.(*HealthComponent); ok {
			scaledMax := diff.ScaledHealth(health.Max)
			// Scale current health proportionally
			ratio := health.Current / health.Max
			health.Max = scaledMax
			health.Current = scaledMax * ratio
		}
	}

	// Scale attack/damage
	if statsComp, ok := entity.GetComponent("stats"); ok {
		if stats, ok := statsComp.(*StatsComponent); ok {
			stats.Attack = diff.ScaledDamage(stats.Attack)
			stats.Defense = diff.ScaledDefense(stats.Defense)
		}
	}

	// Scale attack component damage
	if attackComp, ok := entity.GetComponent("attack"); ok {
		if attack, ok := attackComp.(*AttackComponent); ok {
			attack.Damage = diff.ScaledDamage(attack.Damage)
		}
	}
}

// SetScalingEnabled enables or disables difficulty scaling.
func (s *NGPlusDifficultySystem) SetScalingEnabled(enabled bool) {
	s.scalingEnabled = enabled
}

// IsScalingEnabled returns whether difficulty scaling is enabled.
func (s *NGPlusDifficultySystem) IsScalingEnabled() bool {
	return s.scalingEnabled
}

// SetOnEnemyScaled sets a callback for when an enemy is scaled.
func (s *NGPlusDifficultySystem) SetOnEnemyScaled(callback func(entityID uint64, cycle int)) {
	s.onEnemyScaled = callback
}

// SetOnLootGenerated sets a callback for loot generation with quality bonus.
func (s *NGPlusDifficultySystem) SetOnLootGenerated(callback func(entityID uint64, bonusQuality float64)) {
	s.onLootGenerated = callback
}

// GetCurrentNGPlusCycle returns the cached NG+ cycle level.
func (s *NGPlusDifficultySystem) GetCurrentNGPlusCycle() int {
	return s.currentNGPlusCycle
}

// SetCurrentNGPlusCycle manually sets the NG+ cycle (for testing).
func (s *NGPlusDifficultySystem) SetCurrentNGPlusCycle(cycle int) {
	s.currentNGPlusCycle = cycle
}

// GetLootQualityBonus returns the loot quality bonus for the current NG+ level.
func (s *NGPlusDifficultySystem) GetLootQualityBonus() float64 {
	if s.currentNGPlusCycle <= 0 {
		return 0.0
	}
	// +5% rare chance per ln(cycle+1)
	return CalculateNGPlusMultiplier(s.currentNGPlusCycle, 0.0, 0.05)
}

// GetXPMultiplier returns the XP multiplier for the current NG+ level.
func (s *NGPlusDifficultySystem) GetXPMultiplier() float64 {
	if s.currentNGPlusCycle <= 0 {
		return 1.0
	}
	mult := CalculateNGPlusMultiplier(s.currentNGPlusCycle, 1.0, -0.03)
	if mult < 0.5 {
		return 0.5
	}
	return mult
}

// ScaleEnemyForNGPlus applies NG+ scaling to a newly spawned enemy.
// This is a convenience method for spawning systems.
func (s *NGPlusDifficultySystem) ScaleEnemyForNGPlus(entity *Entity) {
	if s.currentNGPlusCycle <= 0 || !s.scalingEnabled {
		return
	}
	s.applyDifficultyScaling(entity)
}

// GetDifficultyInfo returns a summary of current difficulty settings.
func (s *NGPlusDifficultySystem) GetDifficultyInfo() DifficultyInfo {
	if s.currentNGPlusCycle <= 0 {
		return DifficultyInfo{
			NGPlusCycle:       0,
			Label:             "Normal",
			HealthMultiplier:  1.0,
			DamageMultiplier:  1.0,
			LootQualityBonus:  0.0,
			XPMultiplier:      1.0,
			NewMechanicsLevel: 0,
		}
	}

	comp := NewNGPlusDifficultyComponentForCycle(s.currentNGPlusCycle)
	return DifficultyInfo{
		NGPlusCycle:       s.currentNGPlusCycle,
		Label:             comp.GetDifficultyLabel(),
		HealthMultiplier:  comp.GetHealthMultiplier(),
		DamageMultiplier:  comp.GetDamageMultiplier(),
		LootQualityBonus:  comp.GetLootQualityBonus(),
		XPMultiplier:      comp.GetXPMultiplier(),
		NewMechanicsLevel: comp.GetNewMechanicsLevel(),
	}
}

// DifficultyInfo contains current difficulty parameters for display.
type DifficultyInfo struct {
	NGPlusCycle       int
	Label             string
	HealthMultiplier  float64
	DamageMultiplier  float64
	LootQualityBonus  float64
	XPMultiplier      float64
	NewMechanicsLevel int
}

// ApplyXPScaling adjusts XP gained based on current NG+ level.
// Returns the scaled XP value.
func (s *NGPlusDifficultySystem) ApplyXPScaling(baseXP float64) float64 {
	return baseXP * s.GetXPMultiplier()
}

// ApplyLootQualityScaling adjusts loot drop chance based on current NG+ level.
// Returns the adjusted rare/legendary drop chance bonus.
func (s *NGPlusDifficultySystem) ApplyLootQualityScaling(baseChance float64) float64 {
	return baseChance + s.GetLootQualityBonus()
}

// ShouldUnlockMechanic checks if a specific mechanic level is unlocked.
func (s *NGPlusDifficultySystem) ShouldUnlockMechanic(requiredLevel int) bool {
	if s.currentNGPlusCycle <= 0 {
		return requiredLevel == 0
	}

	comp := NewNGPlusDifficultyComponentForCycle(s.currentNGPlusCycle)
	return comp.GetNewMechanicsLevel() >= requiredLevel
}

// GetUnlockedAbilities returns the list of additional abilities for current NG+.
func (s *NGPlusDifficultySystem) GetUnlockedAbilities() []string {
	if s.currentNGPlusCycle <= 0 {
		return []string{}
	}

	comp := NewNGPlusDifficultyComponentForCycle(s.currentNGPlusCycle)
	return comp.GetAdditionalAbilities()
}

// ResetScaling removes difficulty scaling from all entities.
// Used when starting a new playthrough or for testing.
func (s *NGPlusDifficultySystem) ResetScaling(entities []*Entity) {
	for _, entity := range entities {
		if _, hasDiff := entity.GetComponent("ngplus_difficulty"); hasDiff {
			entity.RemoveComponent("ngplus_difficulty")
		}
	}
	s.currentNGPlusCycle = 0

	if s.logger != nil {
		s.logger.Info("Reset NG+ difficulty scaling")
	}
}
