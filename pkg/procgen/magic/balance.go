package magic

import (
	"fmt"
	"math"
)

// BalanceConfig defines balance constraints and formulas for spell generation.
// This ensures consistent power levels across all spells regardless of type or rarity.
type BalanceConfig struct {
	// Mana cost factors
	BaseManaPerDamage     float64 // Base mana cost per point of damage
	BaseManaPerHealing    float64 // Base mana cost per point of healing
	AreaMultiplier        float64 // Mana multiplier for area spells
	BuffDebuffMultiplier  float64 // Mana multiplier for buffs/debuffs per second
	
	// Cooldown factors
	BaseCooldownPerDamage float64 // Base cooldown per point of damage
	MinCooldown           float64 // Minimum cooldown in seconds
	MaxCooldown           float64 // Maximum cooldown in seconds
	CooldownCastTimeRatio float64 // Ratio between cooldown and cast time
	
	// Level scaling
	PowerPerLevel         float64 // Power increase per player level
	ManaCostPerLevel      float64 // Mana cost increase per level
	CooldownPerLevel      float64 // Cooldown reduction per level
	
	// PvE balance targets
	DPSTarget             float64 // Target damage per second for offensive spells
	HPSTarget             float64 // Target healing per second for healing spells
	MaxDPSVariance        float64 // Maximum variance from target DPS (e.g., 0.3 = ±30%)
}

// DefaultBalanceConfig returns the standard balance configuration.
func DefaultBalanceConfig() BalanceConfig {
	return BalanceConfig{
		// Mana costs: balanced to make all spell types viable
		BaseManaPerDamage:     0.4,  // 50 damage = 20 mana base cost
		BaseManaPerHealing:    0.35, // 50 healing = 17.5 mana base cost
		AreaMultiplier:        1.3,  // Area spells cost 30% more mana
		BuffDebuffMultiplier:  0.6,  // Buffs/debuffs cost less per second
		
		// Cooldown constraints: ensure regular ability usage
		BaseCooldownPerDamage: 0.03, // 50 damage = 1.5s base cooldown
		MinCooldown:           1.0,  // 1 second minimum
		MaxCooldown:           60.0, // 1 minute maximum
		CooldownCastTimeRatio: 2.0,  // Cooldown should be 2x cast time minimum
		
		// Level scaling: smooth power progression
		PowerPerLevel:         0.05,  // 5% power increase per level
		ManaCostPerLevel:      0.02,  // 2% mana cost increase per level
		CooldownPerLevel:      0.01,  // 1% cooldown reduction per level
		
		// PvE targets: balanced combat pacing
		DPSTarget:             15.0,  // Target 15 DPS at level 1
		HPSTarget:             12.0,  // Target 12 HPS at level 1
		MaxDPSVariance:        0.4,   // ±40% variance allowed
	}
}

// BalanceStats applies balance formulas to generated spell stats.
// This ensures mana costs, cooldowns, and power levels are consistent.
func (c *BalanceConfig) BalanceStats(stats *Stats, spellType SpellType, target TargetType, level int) {
	// Calculate base power of the spell
	power := c.calculatePower(stats, spellType)
	
	// Balance mana cost based on power and type
	stats.ManaCost = c.balanceManaCost(power, spellType, target, level)
	
	// Balance cooldown based on power and cast time
	stats.Cooldown = c.balanceCooldown(power, stats.CastTime, level)
	
	// Scale power with level for progression
	c.scalePowerWithLevel(stats, spellType, level)
}

// calculatePower computes the overall power value of a spell.
// This combines damage, healing, duration, and area into a single metric.
func (c *BalanceConfig) calculatePower(stats *Stats, spellType SpellType) float64 {
	power := 0.0
	
	switch spellType {
	case TypeOffensive:
		// Offensive spells: damage is primary power metric
		power = float64(stats.Damage)
		
	case TypeHealing:
		// Healing spells: healing is primary power metric
		power = float64(stats.Healing)
		
	case TypeDefensive, TypeBuff, TypeDebuff:
		// Support spells: duration-based power
		// Estimate power as 10 points per second of duration
		power = stats.Duration * 10.0
		
	case TypeUtility:
		// Utility spells: range and area-based power
		power = stats.Range * 0.5
		if stats.AreaSize > 0 {
			power += stats.AreaSize * 2.0
		}
		
	case TypeSummon:
		// Summon spells: duration-based power
		power = stats.Duration * 15.0
	}
	
	// Area spells have multiplied power
	if stats.AreaSize > 0 {
		power *= 1.0 + (stats.AreaSize * 0.1)
	}
	
	return power
}

// balanceManaCost calculates balanced mana cost based on spell power.
// Higher power spells cost more mana, with modifiers for spell type and target.
func (c *BalanceConfig) balanceManaCost(power float64, spellType SpellType, target TargetType, level int) int {
	baseCost := 0.0
	
	switch spellType {
	case TypeOffensive:
		baseCost = power * c.BaseManaPerDamage
	case TypeHealing:
		baseCost = power * c.BaseManaPerHealing
	case TypeDefensive, TypeBuff, TypeDebuff:
		// Duration-based cost
		baseCost = power * c.BuffDebuffMultiplier
	case TypeUtility:
		baseCost = power * 0.5
	case TypeSummon:
		baseCost = power * 1.0
	}
	
	// Area effect multiplier
	if target == TargetArea || target == TargetCone || target == TargetLine {
		baseCost *= c.AreaMultiplier
	} else if target == TargetAllAllies || target == TargetAllEnemies {
		baseCost *= c.AreaMultiplier * 1.5
	}
	
	// Level scaling
	levelMultiplier := 1.0 + (float64(level) * c.ManaCostPerLevel)
	finalCost := baseCost * levelMultiplier
	
	// Ensure minimum cost of 5 mana
	if finalCost < 5.0 {
		finalCost = 5.0
	}
	
	return int(math.Round(finalCost))
}

// balanceCooldown calculates balanced cooldown based on spell power and cast time.
// Ensures cooldown is proportional to power and allows regular spell usage.
func (c *BalanceConfig) balanceCooldown(power, castTime float64, level int) float64 {
	// Base cooldown from power
	baseCooldown := power * c.BaseCooldownPerDamage
	
	// Ensure cooldown is at least 3x cast time
	minFromCastTime := castTime * c.CooldownCastTimeRatio
	if baseCooldown < minFromCastTime {
		baseCooldown = minFromCastTime
	}
	
	// Apply level-based cooldown reduction
	levelReduction := 1.0 - (float64(level) * c.CooldownPerLevel)
	if levelReduction < 0.7 {
		levelReduction = 0.7 // Cap at 30% reduction
	}
	finalCooldown := baseCooldown * levelReduction
	
	// Enforce min/max bounds
	if finalCooldown < c.MinCooldown {
		finalCooldown = c.MinCooldown
	}
	if finalCooldown > c.MaxCooldown {
		finalCooldown = c.MaxCooldown
	}
	
	return math.Round(finalCooldown*10) / 10 // Round to 1 decimal place
}

// scalePowerWithLevel applies level-based scaling to spell power.
// This ensures spells remain relevant as the player progresses.
func (c *BalanceConfig) scalePowerWithLevel(stats *Stats, spellType SpellType, level int) {
	if level <= 1 {
		return
	}
	
	levelScale := 1.0 + (float64(level-1) * c.PowerPerLevel)
	
	switch spellType {
	case TypeOffensive:
		stats.Damage = int(math.Round(float64(stats.Damage) * levelScale))
	case TypeHealing:
		stats.Healing = int(math.Round(float64(stats.Healing) * levelScale))
	case TypeDefensive, TypeBuff, TypeDebuff, TypeSummon:
		stats.Duration = math.Round(stats.Duration * levelScale * 10) / 10
	}
	
	// Scale area size slightly with level
	if stats.AreaSize > 0 {
		areaScale := 1.0 + (float64(level-1) * c.PowerPerLevel * 0.5)
		stats.AreaSize = math.Round(stats.AreaSize * areaScale * 10) / 10
	}
}

// ValidateDPS checks if a spell's DPS is within acceptable balance range.
// Returns nil if balanced, error with details if imbalanced.
func (c *BalanceConfig) ValidateDPS(spell *Spell) error {
	if spell.Type != TypeOffensive || spell.Stats.Damage == 0 {
		return nil // Non-offensive spells don't need DPS validation
	}
	
	// Calculate DPS: damage / (castTime + cooldown)
	totalTime := spell.Stats.CastTime + spell.Stats.Cooldown
	if totalTime <= 0 {
		return fmt.Errorf("invalid spell timing: castTime=%.2f, cooldown=%.2f", 
			spell.Stats.CastTime, spell.Stats.Cooldown)
	}
	
	dps := float64(spell.Stats.Damage) / totalTime
	
	// Calculate target DPS for this level
	targetDPS := c.DPSTarget * (1.0 + float64(spell.Stats.RequiredLevel-1)*c.PowerPerLevel)
	
	// Check if within acceptable variance
	minDPS := targetDPS * (1.0 - c.MaxDPSVariance)
	maxDPS := targetDPS * (1.0 + c.MaxDPSVariance)
	
	if dps < minDPS {
		return fmt.Errorf("spell %q DPS too low: %.2f (target: %.2f-%.2f)", 
			spell.Name, dps, minDPS, maxDPS)
	}
	if dps > maxDPS {
		return fmt.Errorf("spell %q DPS too high: %.2f (target: %.2f-%.2f)", 
			spell.Name, dps, minDPS, maxDPS)
	}
	
	return nil
}

// ValidateHPS checks if a healing spell's HPS is within acceptable balance range.
// Returns nil if balanced, error with details if imbalanced.
func (c *BalanceConfig) ValidateHPS(spell *Spell) error {
	if spell.Type != TypeHealing || spell.Stats.Healing == 0 {
		return nil // Non-healing spells don't need HPS validation
	}
	
	// Calculate HPS: healing / (castTime + cooldown)
	totalTime := spell.Stats.CastTime + spell.Stats.Cooldown
	if totalTime <= 0 {
		return fmt.Errorf("invalid spell timing: castTime=%.2f, cooldown=%.2f", 
			spell.Stats.CastTime, spell.Stats.Cooldown)
	}
	
	hps := float64(spell.Stats.Healing) / totalTime
	
	// Calculate target HPS for this level
	targetHPS := c.HPSTarget * (1.0 + float64(spell.Stats.RequiredLevel-1)*c.PowerPerLevel)
	
	// Check if within acceptable variance
	minHPS := targetHPS * (1.0 - c.MaxDPSVariance)
	maxHPS := targetHPS * (1.0 + c.MaxDPSVariance)
	
	if hps < minHPS {
		return fmt.Errorf("spell %q HPS too low: %.2f (target: %.2f-%.2f)", 
			spell.Name, hps, minHPS, maxHPS)
	}
	if hps > maxHPS {
		return fmt.Errorf("spell %q HPS too high: %.2f (target: %.2f-%.2f)", 
			spell.Name, hps, minHPS, maxHPS)
	}
	
	return nil
}

// ValidateManaCostEfficiency checks if mana cost is reasonable for spell power.
// Returns nil if efficient, error if too cheap or expensive.
func (c *BalanceConfig) ValidateManaCostEfficiency(spell *Spell) error {
	power := c.calculatePower(&spell.Stats, spell.Type)
	if power <= 0 {
		return nil // No power, no validation needed
	}
	
	// Calculate efficiency: power per mana point
	efficiency := power / float64(spell.Stats.ManaCost)
	
	// Expected efficiency range (1.0-4.5 power per mana)
	// Relaxed range to accommodate different spell types
	minEfficiency := 1.0
	maxEfficiency := 4.5
	
	if efficiency < minEfficiency {
		return fmt.Errorf("spell %q too expensive: %.2f power per mana (min: %.2f)", 
			spell.Name, efficiency, minEfficiency)
	}
	if efficiency > maxEfficiency {
		return fmt.Errorf("spell %q too cheap: %.2f power per mana (max: %.2f)", 
			spell.Name, efficiency, maxEfficiency)
	}
	
	return nil
}

// CalculatePowerRating returns a numerical power rating for the spell (0-100).
// This provides a standardized way to compare spell power across types.
func (c *BalanceConfig) CalculatePowerRating(spell *Spell) int {
	power := c.calculatePower(&spell.Stats, spell.Type)
	
	// Calculate DPS/HPS for rating
	totalTime := spell.Stats.CastTime + spell.Stats.Cooldown
	if totalTime <= 0 {
		totalTime = 1.0
	}
	
	var throughput float64
	if spell.Type == TypeOffensive {
		throughput = float64(spell.Stats.Damage) / totalTime
	} else if spell.Type == TypeHealing {
		throughput = float64(spell.Stats.Healing) / totalTime
	} else {
		throughput = power / 5.0 // Other spell types use power/5 as rating
	}
	
	// Scale by level and rarity
	levelScale := 1.0 + float64(spell.Stats.RequiredLevel-1)*0.1
	rarityScale := 1.0 + float64(spell.Rarity)*0.2
	
	rating := throughput * levelScale * rarityScale
	
	// Clamp to 0-100 range
	if rating < 0 {
		rating = 0
	}
	if rating > 100 {
		rating = 100
	}
	
	return int(math.Round(rating))
}
