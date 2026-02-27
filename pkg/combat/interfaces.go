package combat

// CombatResolver handles combat calculations.
type CombatResolver interface {
	// CalculateDamage computes final damage after resistances and defenses
	CalculateDamage(damage Damage, targetStats *Stats) float64

	// ResolveCombat handles a combat interaction between entities
	ResolveCombat(attackerID, defenderID uint64) []Damage
}
