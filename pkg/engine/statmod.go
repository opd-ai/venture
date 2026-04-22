// Package engine statmod.go provides a reusable cache/apply/remove helper for
// stat-bonus systems. It abstracts the repeated pattern of caching an entity's
// original stat value, detecting whether the bonus changed, applying the new
// value via a caller-supplied formula, and restoring the original on removal.
//
// Usage:
//
//	type MySystem struct {
//	    world  *World
//	    bonus  statBonusApplier
//	}
//
//	func NewMySystem(world *World) *MySystem {
//	    return &MySystem{world: world, bonus: newStatBonusApplier()}
//	}
//
//	func (s *MySystem) processEntity(entity *Entity, amount float64) {
//	    s.bonus.apply(entity, amount,
//	        func(st *StatsComponent) float64  { return st.Speed },
//	        func(st *StatsComponent, v float64) { st.Speed = v },
//	        multiplicativeBonus,
//	    )
//	}
package engine

// statBonusApplier handles the cache/apply/remove lifecycle for a single
// per-entity stat modifier. It is value-type (embed by value, not pointer) and
// intended to replace the `originalXxx map[uint64]float64` + `appliedBonuses
// map[uint64]float64` pair that appears verbatim across 20+ stat-modifier
// systems in pkg/engine.
type statBonusApplier struct {
	original map[uint64]float64
	applied  map[uint64]float64
}

// newStatBonusApplier returns a zero-value applier with initialised maps.
func newStatBonusApplier() statBonusApplier {
	return statBonusApplier{
		original: make(map[uint64]float64),
		applied:  make(map[uint64]float64),
	}
}

// apply updates the stat on entity's StatsComponent using the supplied
// accessors and formula.  It is a no-op (and returns false) when the bonus
// value has not changed since the last call.
//
// Parameters:
//   - entity: target entity (must have a "stats" component)
//   - bonus: the modifier value to apply
//   - getStat: reads the stat from a StatsComponent
//   - setStat: writes the new value back to the StatsComponent
//   - formula: computes newValue from (originalValue, bonus)
//
// Returns true when the stat was actually modified.
func (a *statBonusApplier) apply(
	entity *Entity,
	bonus float64,
	getStat func(*StatsComponent) float64,
	setStat func(*StatsComponent, float64),
	formula func(original, bonus float64) float64,
) bool {
	statsComp, ok := entity.GetComponent("stats")
	if !ok {
		return false
	}
	stats, ok := statsComp.(*StatsComponent)
	if !ok {
		return false
	}

	// Skip if nothing changed.
	if cur, has := a.applied[entity.ID]; has && cur == bonus {
		return false
	}

	// Capture the original value once, before any modification.
	if _, exists := a.original[entity.ID]; !exists {
		a.original[entity.ID] = getStat(stats)
	}

	setStat(stats, formula(a.original[entity.ID], bonus))
	a.applied[entity.ID] = bonus
	return true
}

// remove clears any cached state for entityID without touching the
// StatsComponent.  Call this when an entity loses its components or is
// despawned.
func (a *statBonusApplier) remove(entityID uint64) {
	delete(a.applied, entityID)
	delete(a.original, entityID)
}

// currentBonus returns the most recently applied bonus value for entityID,
// or 0 if none has been applied.
func (a *statBonusApplier) currentBonus(entityID uint64) float64 {
	return a.applied[entityID]
}

// multiplicativeBonus computes original*(1+bonus), the standard formula for
// percentage increases (e.g. "20% more defense").
func multiplicativeBonus(original, bonus float64) float64 {
	return original * (1.0 + bonus)
}

// additiveBonus computes original+bonus, the standard formula for flat
// increases (e.g. "+0.10 evasion").
func additiveBonus(original, bonus float64) float64 {
	return original + bonus
}

// clampedAdditiveBonus returns an additiveBonus formula with an upper cap.
func clampedAdditiveBonus(cap float64) func(float64, float64) float64 {
	return func(original, bonus float64) float64 {
		v := original + bonus
		if v > cap {
			return cap
		}
		if v < 0 {
			return 0
		}
		return v
	}
}
