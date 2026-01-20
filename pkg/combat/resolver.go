// resolver.go provides a default implementation of the CombatResolver interface.
// This file contains the DefaultCombatResolver which handles damage calculation
// considering defense, magic defense, and resistances. It also provides combat
// resolution between entities.
package combat

import "math"

// DefaultCombatResolver is a reference implementation of CombatResolver.
// It applies standard RPG damage formulas:
// - Physical damage reduced by Defense
// - Magical damage reduced by MagicDefense
// - Elemental damage reduced by resistances (0.0 = no resist, 1.0 = immune)
type DefaultCombatResolver struct {
	// MinDamageMultiplier is the minimum damage factor after mitigation (default 0.1).
	// This ensures attacks always deal some damage even against heavily armored targets.
	MinDamageMultiplier float64

	// EntityLookup provides Stats for entities by ID.
	// This is required for ResolveCombat to function.
	EntityLookup EntityStatsProvider
}

// EntityStatsProvider retrieves combat stats for entities.
type EntityStatsProvider interface {
	// GetStats returns the Stats for an entity, or nil if not found.
	GetStats(entityID uint64) *Stats
	// GetAttackDamage returns the base damage for an entity's attack.
	GetAttackDamage(entityID uint64) Damage
}

// NewDefaultCombatResolver creates a new DefaultCombatResolver with sensible defaults.
func NewDefaultCombatResolver(lookup EntityStatsProvider) *DefaultCombatResolver {
	return &DefaultCombatResolver{
		MinDamageMultiplier: 0.1,
		EntityLookup:        lookup,
	}
}

// CalculateDamage computes final damage after applying defenses and resistances.
// The formula is:
//  1. Start with base damage amount
//  2. Apply defense reduction: damage * (100 / (100 + defense))
//  3. Apply resistance: damage * (1 - resistance)
//  4. Ensure minimum damage (MinDamageMultiplier * base damage)
//
// Returns the final damage amount to apply to the target.
func (r *DefaultCombatResolver) CalculateDamage(damage Damage, targetStats *Stats) float64 {
	if targetStats == nil {
		return damage.Amount
	}

	baseDamage := damage.Amount
	if baseDamage <= 0 {
		return 0
	}

	// Step 1: Determine which defense stat to use
	defense := r.getDefenseForDamageType(damage.Type, targetStats)

	// Step 2: Apply defense reduction using diminishing returns formula
	// This formula prevents defense from fully negating damage
	defenseMultiplier := 100.0 / (100.0 + defense)
	mitigatedDamage := baseDamage * defenseMultiplier

	// Step 3: Apply resistance
	resistance := targetStats.Resistances[damage.Type]
	// Clamp resistance to prevent negative damage or over-immunity
	resistance = math.Max(-0.5, math.Min(resistance, 1.0))
	resistedDamage := mitigatedDamage * (1.0 - resistance)

	// Step 4: Ensure minimum damage
	minDamage := baseDamage * r.MinDamageMultiplier
	finalDamage := math.Max(resistedDamage, minDamage)

	return finalDamage
}

// getDefenseForDamageType returns the appropriate defense stat for a damage type.
func (r *DefaultCombatResolver) getDefenseForDamageType(damageType DamageType, stats *Stats) float64 {
	switch damageType {
	case DamagePhysical:
		return stats.Defense
	case DamageMagical, DamageFire, DamageIce, DamageLightning, DamagePoison:
		return stats.MagicDefense
	default:
		return 0
	}
}

// ResolveCombat handles a combat interaction between an attacker and defender.
// It retrieves stats using EntityLookup, calculates damage, and returns all
// damage events from the attack.
//
// Returns nil if EntityLookup is nil or either entity is not found.
func (r *DefaultCombatResolver) ResolveCombat(attackerID, defenderID uint64) []Damage {
	if r.EntityLookup == nil {
		return nil
	}

	attackerStats := r.EntityLookup.GetStats(attackerID)
	defenderStats := r.EntityLookup.GetStats(defenderID)

	if attackerStats == nil || defenderStats == nil {
		return nil
	}

	// Get the base attack damage from the attacker
	baseDamage := r.EntityLookup.GetAttackDamage(attackerID)
	baseDamage.SourceID = attackerID
	baseDamage.TargetID = defenderID

	// Calculate final damage after mitigation
	finalAmount := r.CalculateDamage(baseDamage, defenderStats)

	// Create the result damage event
	resultDamage := Damage{
		Amount:   finalAmount,
		Type:     baseDamage.Type,
		SourceID: attackerID,
		TargetID: defenderID,
	}

	return []Damage{resultDamage}
}
