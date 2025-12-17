// Package engine provides the NG+ Difficulty component for ECS.
// This file implements NGPlusDifficultyComponent for tracking difficulty modifiers
// that scale with New Game Plus progression.
//
// Phase 113: Difficulty Scaling System
package engine

import (
	"encoding/json"
	"sync"
)

// NGPlusDifficultyComponent tracks difficulty modifiers for entities.
// These modifiers scale enemy stats, loot quality, and XP gain based on NG+ level.
type NGPlusDifficultyComponent struct {
	mu sync.RWMutex

	// HealthMultiplier scales entity max health (1.0 = base)
	HealthMultiplier float64 `json:"health_multiplier"`

	// DamageMultiplier scales entity damage output (1.0 = base)
	DamageMultiplier float64 `json:"damage_multiplier"`

	// DefenseMultiplier scales entity defense (1.0 = base)
	DefenseMultiplier float64 `json:"defense_multiplier"`

	// LootQualityBonus is added to rare/legendary drop chance (0.0 = no bonus)
	LootQualityBonus float64 `json:"loot_quality_bonus"`

	// XPMultiplier scales XP gained from this entity (1.0 = base)
	XPMultiplier float64 `json:"xp_multiplier"`

	// NewMechanicsLevel indicates which enemy mechanics tier is active
	// 0 = base mechanics, 1+ = additional abilities unlocked
	NewMechanicsLevel int `json:"new_mechanics_level"`

	// NGPlusCycle is the NG+ level this entity was spawned at
	NGPlusCycle int `json:"ngplus_cycle"`

	// IsScaled indicates whether difficulty scaling has been applied
	IsScaled bool `json:"is_scaled"`

	// BossEnhancementLevel indicates boss-specific enhancements (0 = none)
	// Level 1: New attack pattern at NG+2
	// Level 2: Additional phase at NG+5
	// Level 3: Enrage mechanic at NG+10
	BossEnhancementLevel int `json:"boss_enhancement_level"`

	// HasEnragedPhase indicates if boss has unlocked enrage mechanic
	HasEnragedPhase bool `json:"has_enraged_phase"`

	// AdditionalAbilities lists extra abilities unlocked at higher NG+ levels
	AdditionalAbilities []string `json:"additional_abilities,omitempty"`
}

// Type returns the component type identifier.
func (n *NGPlusDifficultyComponent) Type() string {
	return "ngplus_difficulty"
}

// NewNGPlusDifficultyComponent creates a new component with base values (no scaling).
func NewNGPlusDifficultyComponent() *NGPlusDifficultyComponent {
	return &NGPlusDifficultyComponent{
		HealthMultiplier:     1.0,
		DamageMultiplier:     1.0,
		DefenseMultiplier:    1.0,
		LootQualityBonus:     0.0,
		XPMultiplier:         1.0,
		NewMechanicsLevel:    0,
		NGPlusCycle:          0,
		IsScaled:             false,
		BossEnhancementLevel: 0,
		HasEnragedPhase:      false,
		AdditionalAbilities:  []string{},
	}
}

// NewNGPlusDifficultyComponentForCycle creates a component with scaling for the given NG+ cycle.
func NewNGPlusDifficultyComponentForCycle(cycle int) *NGPlusDifficultyComponent {
	comp := NewNGPlusDifficultyComponent()
	comp.ApplyScalingForCycle(cycle)
	return comp
}

// ApplyScalingForCycle calculates and applies difficulty scaling for a given NG+ cycle.
// Uses logarithmic scaling to prevent excessive difficulty at high NG+ levels.
func (n *NGPlusDifficultyComponent) ApplyScalingForCycle(cycle int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.NGPlusCycle = cycle
	n.IsScaled = true

	if cycle <= 0 {
		// First playthrough - no scaling
		return
	}

	// Health: +20% per ln(cycle+1)
	// NG+1: ~1.14x, NG+5: ~1.36x, NG+10: ~1.48x
	n.HealthMultiplier = CalculateNGPlusMultiplier(cycle, 1.0, 0.2)

	// Damage: +15% per ln(cycle+1)
	// NG+1: ~1.10x, NG+5: ~1.27x, NG+10: ~1.36x
	n.DamageMultiplier = CalculateNGPlusMultiplier(cycle, 1.0, 0.15)

	// Defense: +10% per ln(cycle+1)
	n.DefenseMultiplier = CalculateNGPlusMultiplier(cycle, 1.0, 0.10)

	// Loot quality: +5% rare chance per ln(cycle+1)
	n.LootQualityBonus = CalculateNGPlusMultiplier(cycle, 0.0, 0.05)

	// XP: Decreases slightly at high NG+ to maintain challenge (-3% per ln(cycle+1))
	// Minimum 50% XP gain
	xpMult := CalculateNGPlusMultiplier(cycle, 1.0, -0.03)
	if xpMult < 0.5 {
		xpMult = 0.5
	}
	n.XPMultiplier = xpMult

	// New mechanics unlock thresholds
	// NG+1: Level 0 (base)
	// NG+3: Level 1 (minor new abilities)
	// NG+5: Level 2 (intermediate abilities)
	// NG+7: Level 3 (advanced abilities)
	// NG+10+: Level 4 (master abilities)
	switch {
	case cycle >= 10:
		n.NewMechanicsLevel = 4
	case cycle >= 7:
		n.NewMechanicsLevel = 3
	case cycle >= 5:
		n.NewMechanicsLevel = 2
	case cycle >= 3:
		n.NewMechanicsLevel = 1
	default:
		n.NewMechanicsLevel = 0
	}

	// Populate additional abilities based on mechanics level
	n.AdditionalAbilities = getAbilitiesForMechanicsLevel(n.NewMechanicsLevel)
}

// getAbilitiesForMechanicsLevel returns abilities unlocked at a given mechanics level.
func getAbilitiesForMechanicsLevel(level int) []string {
	abilities := []string{}

	if level >= 1 {
		abilities = append(abilities, "counter_attack", "dodge_roll")
	}
	if level >= 2 {
		abilities = append(abilities, "elemental_shield", "group_heal")
	}
	if level >= 3 {
		abilities = append(abilities, "summon_minions", "area_denial")
	}
	if level >= 4 {
		abilities = append(abilities, "phase_shift", "ultimate_attack")
	}

	return abilities
}

// ApplyBossEnhancements applies boss-specific enhancements based on NG+ cycle.
func (n *NGPlusDifficultyComponent) ApplyBossEnhancements(cycle int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Boss enhancement levels:
	// NG+2: Level 1 - New attack pattern
	// NG+5: Level 2 - Additional phase
	// NG+10: Level 3 - Enrage mechanic
	switch {
	case cycle >= 10:
		n.BossEnhancementLevel = 3
		n.HasEnragedPhase = true
	case cycle >= 5:
		n.BossEnhancementLevel = 2
	case cycle >= 2:
		n.BossEnhancementLevel = 1
	default:
		n.BossEnhancementLevel = 0
	}
}

// GetHealthMultiplier returns the health scaling multiplier (thread-safe).
func (n *NGPlusDifficultyComponent) GetHealthMultiplier() float64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.HealthMultiplier
}

// GetDamageMultiplier returns the damage scaling multiplier (thread-safe).
func (n *NGPlusDifficultyComponent) GetDamageMultiplier() float64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.DamageMultiplier
}

// GetDefenseMultiplier returns the defense scaling multiplier (thread-safe).
func (n *NGPlusDifficultyComponent) GetDefenseMultiplier() float64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.DefenseMultiplier
}

// GetLootQualityBonus returns the loot quality bonus (thread-safe).
func (n *NGPlusDifficultyComponent) GetLootQualityBonus() float64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.LootQualityBonus
}

// GetXPMultiplier returns the XP scaling multiplier (thread-safe).
func (n *NGPlusDifficultyComponent) GetXPMultiplier() float64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.XPMultiplier
}

// GetNewMechanicsLevel returns the mechanics tier level (thread-safe).
func (n *NGPlusDifficultyComponent) GetNewMechanicsLevel() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.NewMechanicsLevel
}

// GetNGPlusCycle returns the NG+ cycle this entity was spawned at (thread-safe).
func (n *NGPlusDifficultyComponent) GetNGPlusCycle() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.NGPlusCycle
}

// GetBossEnhancementLevel returns the boss enhancement level (thread-safe).
func (n *NGPlusDifficultyComponent) GetBossEnhancementLevel() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.BossEnhancementLevel
}

// HasAbility checks if an ability is unlocked for this entity.
func (n *NGPlusDifficultyComponent) HasAbility(abilityID string) bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	for _, ability := range n.AdditionalAbilities {
		if ability == abilityID {
			return true
		}
	}
	return false
}

// GetAdditionalAbilities returns a copy of the additional abilities slice.
func (n *NGPlusDifficultyComponent) GetAdditionalAbilities() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	result := make([]string, len(n.AdditionalAbilities))
	copy(result, n.AdditionalAbilities)
	return result
}

// Serialize converts the component to JSON for persistence.
func (n *NGPlusDifficultyComponent) Serialize() ([]byte, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return json.Marshal(n)
}

// Deserialize restores the component from JSON data.
func (n *NGPlusDifficultyComponent) Deserialize(data []byte) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	var temp NGPlusDifficultyComponent
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	n.HealthMultiplier = temp.HealthMultiplier
	n.DamageMultiplier = temp.DamageMultiplier
	n.DefenseMultiplier = temp.DefenseMultiplier
	n.LootQualityBonus = temp.LootQualityBonus
	n.XPMultiplier = temp.XPMultiplier
	n.NewMechanicsLevel = temp.NewMechanicsLevel
	n.NGPlusCycle = temp.NGPlusCycle
	n.IsScaled = temp.IsScaled
	n.BossEnhancementLevel = temp.BossEnhancementLevel
	n.HasEnragedPhase = temp.HasEnragedPhase
	n.AdditionalAbilities = temp.AdditionalAbilities

	if n.AdditionalAbilities == nil {
		n.AdditionalAbilities = []string{}
	}

	return nil
}

// ScaledHealth calculates scaled health value from base health.
func (n *NGPlusDifficultyComponent) ScaledHealth(baseHealth float64) float64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return baseHealth * n.HealthMultiplier
}

// ScaledDamage calculates scaled damage value from base damage.
func (n *NGPlusDifficultyComponent) ScaledDamage(baseDamage float64) float64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return baseDamage * n.DamageMultiplier
}

// ScaledDefense calculates scaled defense value from base defense.
func (n *NGPlusDifficultyComponent) ScaledDefense(baseDefense float64) float64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return baseDefense * n.DefenseMultiplier
}

// ScaledXP calculates scaled XP value from base XP.
func (n *NGPlusDifficultyComponent) ScaledXP(baseXP float64) float64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return baseXP * n.XPMultiplier
}

// GetDifficultyLabel returns a human-readable difficulty label.
func (n *NGPlusDifficultyComponent) GetDifficultyLabel() string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.NGPlusCycle <= 0 {
		return "Normal"
	}

	switch {
	case n.NGPlusCycle >= 10:
		return "Legendary"
	case n.NGPlusCycle >= 7:
		return "Nightmare"
	case n.NGPlusCycle >= 5:
		return "Hell"
	case n.NGPlusCycle >= 3:
		return "Hard"
	default:
		return "Challenging"
	}
}
