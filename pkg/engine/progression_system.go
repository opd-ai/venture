package engine

import (
	"fmt"
	"math"
)

// LevelUpCallback is called when an entity levels up.
// It receives the entity and the new level.
type LevelUpCallback func(entity *Entity, newLevel int)

// ProgressionSystem manages character progression, experience gain, and leveling.
// It handles XP distribution, level-ups, and stat scaling.
type ProgressionSystem struct {
	world            *World
	levelUpCallbacks []LevelUpCallback
	xpCurve          XPCurveFunc
}

// XPCurveFunc defines how much XP is required for each level.
// It takes the target level and returns the XP required to reach it.
type XPCurveFunc func(level int) int

// NewProgressionSystem creates a new progression system.
func NewProgressionSystem(world *World) *ProgressionSystem {
	return &ProgressionSystem{
		world:            world,
		levelUpCallbacks: make([]LevelUpCallback, 0),
		xpCurve:          DefaultXPCurve,
	}
}

// DefaultXPCurve provides a standard exponential XP curve.
// Formula: 100 * (level ^ 1.5)
// This creates a curve that gets steeper but remains achievable.
func DefaultXPCurve(level int) int {
	if level < 1 {
		level = 1
	}
	// Base XP * (level^1.5)
	xp := 100.0 * math.Pow(float64(level), 1.5)
	return int(xp)
}

// LinearXPCurve provides a linear XP curve.
// Formula: 100 * level
// Each level requires the same additional XP.
func LinearXPCurve(level int) int {
	if level < 1 {
		level = 1
	}
	return 100 * level
}

// ExponentialXPCurve provides a steep exponential XP curve.
// Formula: 100 * (level ^ 2)
// This creates very steep progression for hardcore games.
func ExponentialXPCurve(level int) int {
	if level < 1 {
		level = 1
	}
	return 100 * level * level
}

// SetXPCurve sets the XP curve function for the system.
func (ps *ProgressionSystem) SetXPCurve(curve XPCurveFunc) {
	if curve != nil {
		ps.xpCurve = curve
	}
}

// AddLevelUpCallback adds a callback that will be called when an entity levels up.
func (ps *ProgressionSystem) AddLevelUpCallback(callback LevelUpCallback) {
	if callback != nil {
		ps.levelUpCallbacks = append(ps.levelUpCallbacks, callback)
	}
}

// AwardXP gives experience points to an entity.
// If the entity levels up, stats are automatically updated.
func (ps *ProgressionSystem) AwardXP(entity *Entity, xp int) error {
	if entity == nil {
		return fmt.Errorf("cannot award XP to nil entity")
	}
	if xp <= 0 {
		return fmt.Errorf("XP amount must be positive")
	}

	expComp, ok := entity.GetComponent("experience")
	if !ok {
		return fmt.Errorf("entity does not have experience component")
	}

	exp := expComp.(*ExperienceComponent)

	// Add the XP
	leveled := exp.AddXP(xp)

	// Process level ups
	if leveled {
		ps.processLevelUps(entity, exp)
	}

	return nil
}

// processLevelUps handles one or more level ups for an entity.
func (ps *ProgressionSystem) processLevelUps(entity *Entity, exp *ExperienceComponent) {
	// Process all level ups (in case of large XP gain)
	for exp.CanLevelUp() {
		// Deduct XP for this level
		exp.CurrentXP -= exp.RequiredXP
		exp.Level++

		// Award skill point
		exp.SkillPoints++

		// Calculate XP for next level
		exp.RequiredXP = ps.xpCurve(exp.Level)

		// Update stats based on level scaling
		ps.updateStatsForLevel(entity, exp.Level)

		// Trigger callbacks
		for _, callback := range ps.levelUpCallbacks {
			callback(entity, exp.Level)
		}
	}
}

// updateStatsForLevel updates an entity's stats based on their new level.
func (ps *ProgressionSystem) updateStatsForLevel(entity *Entity, level int) {
	// Get level scaling component
	scalingComp, ok := entity.GetComponent("level_scaling")
	if !ok {
		return // No scaling defined
	}
	scaling := scalingComp.(*LevelScalingComponent)

	// Update health component
	healthComp, ok := entity.GetComponent("health")
	if ok {
		health := healthComp.(*HealthComponent)
		oldMax := health.Max
		health.Max = scaling.CalculateHealthForLevel(level)
		// Increase current health by the same amount
		health.Current += (health.Max - oldMax)
	}

	// Update stats component
	statsComp, ok := entity.GetComponent("stats")
	if ok {
		stats := statsComp.(*StatsComponent)
		stats.Attack = scaling.CalculateAttackForLevel(level)
		stats.Defense = scaling.CalculateDefenseForLevel(level)
		stats.MagicPower = scaling.CalculateMagicPowerForLevel(level)
		stats.MagicDefense = scaling.CalculateMagicDefenseForLevel(level)
	}
}

// CalculateXPReward calculates how much XP a defeated entity should award.
// This is based on the entity's level and stats.
func (ps *ProgressionSystem) CalculateXPReward(defeatedEntity *Entity) int {
	// Get the entity's level
	expComp, ok := defeatedEntity.GetComponent("experience")
	if !ok {
		// If no experience component, use a base reward
		return 10
	}

	exp := expComp.(*ExperienceComponent)
	level := exp.Level

	// Base XP = 10 * level
	// This means a level 5 enemy gives 50 XP
	baseXP := 10 * level

	// Bonus for elite/boss entities (could check for boss component later)
	// For now, just use the base calculation

	return baseXP
}

// GetLevel returns the current level of an entity.
// Returns 1 if the entity has no experience component.
func (ps *ProgressionSystem) GetLevel(entity *Entity) int {
	if entity == nil {
		return 1
	}

	expComp, ok := entity.GetComponent("experience")
	if !ok {
		return 1
	}

	return expComp.(*ExperienceComponent).Level
}

// GetXPProgress returns the XP progress as a value between 0.0 and 1.0.
func (ps *ProgressionSystem) GetXPProgress(entity *Entity) float64 {
	if entity == nil {
		return 0.0
	}

	expComp, ok := entity.GetComponent("experience")
	if !ok {
		return 0.0
	}

	return expComp.(*ExperienceComponent).ProgressToNextLevel()
}

// SpendSkillPoint spends a skill point for an entity.
// Returns an error if the entity has no skill points to spend.
func (ps *ProgressionSystem) SpendSkillPoint(entity *Entity) error {
	if entity == nil {
		return fmt.Errorf("cannot spend skill point for nil entity")
	}

	expComp, ok := entity.GetComponent("experience")
	if !ok {
		return fmt.Errorf("entity does not have experience component")
	}

	exp := expComp.(*ExperienceComponent)
	if exp.SkillPoints <= 0 {
		return fmt.Errorf("entity has no skill points to spend")
	}

	exp.SkillPoints--
	return nil
}

// GetSkillPoints returns the number of unspent skill points for an entity.
func (ps *ProgressionSystem) GetSkillPoints(entity *Entity) int {
	if entity == nil {
		return 0
	}

	expComp, ok := entity.GetComponent("experience")
	if !ok {
		return 0
	}

	return expComp.(*ExperienceComponent).SkillPoints
}

// InitializeEntityAtLevel sets up an entity at a specific level.
// This is useful for spawning monsters at appropriate levels.
func (ps *ProgressionSystem) InitializeEntityAtLevel(entity *Entity, level int) error {
	if entity == nil {
		return fmt.Errorf("cannot initialize nil entity")
	}
	if level < 1 {
		level = 1
	}

	// Get or create experience component
	expComp, ok := entity.GetComponent("experience")
	var exp *ExperienceComponent
	if !ok {
		exp = NewExperienceComponent()
		entity.AddComponent(exp)
	} else {
		exp = expComp.(*ExperienceComponent)
	}

	// Set level and XP
	exp.Level = level
	exp.CurrentXP = 0
	exp.RequiredXP = ps.xpCurve(level)

	// Calculate total XP for this level
	totalXP := 0
	for i := 1; i < level; i++ {
		totalXP += ps.xpCurve(i)
	}
	exp.TotalXP = totalXP

	// Update stats for this level
	ps.updateStatsForLevel(entity, level)

	return nil
}

// Update implements the System interface.
// ProgressionSystem doesn't need per-frame updates, so this is a no-op.
func (ps *ProgressionSystem) Update(entities []*Entity, deltaTime float64) {
	// ProgressionSystem is event-driven (AwardXP), not frame-driven
	// No per-frame updates needed
}
