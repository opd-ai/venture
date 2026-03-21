// Package engine provides shared helpers for time-of-day stat modification systems.
// These helpers reduce duplication across TimeOfDayBlockChanceSystem,
// TimeOfDayCriticalChanceSystem, TimeOfDayEvasionSystem, and similar systems.
package engine

// StatAccessor defines how to read and write a stat field from StatsComponent.
type StatAccessor struct {
	// Get returns the current value of the stat from the component.
	Get func(*StatsComponent) float64
	// Set updates the stat value on the component.
	Set func(*StatsComponent, float64)
}

// Common stat accessors for time-of-day systems.
var (
	// BlockChanceAccessor accesses StatsComponent.BlockChance.
	BlockChanceAccessor = StatAccessor{
		Get: func(s *StatsComponent) float64 { return s.BlockChance },
		Set: func(s *StatsComponent, v float64) { s.BlockChance = v },
	}
	// CritChanceAccessor accesses StatsComponent.CritChance.
	CritChanceAccessor = StatAccessor{
		Get: func(s *StatsComponent) float64 { return s.CritChance },
		Set: func(s *StatsComponent, v float64) { s.CritChance = v },
	}
	// EvasionAccessor accesses StatsComponent.Evasion.
	EvasionAccessor = StatAccessor{
		Get: func(s *StatsComponent) float64 { return s.Evasion },
		Set: func(s *StatsComponent, v float64) { s.Evasion = v },
	}
)

// ApplyTimeOfDayStatModifier applies a modifier to a stat for all entities,
// caching original values and clamping results to [0.0, 1.0].
// Returns the number of entities modified.
func ApplyTimeOfDayStatModifier(
	entities []*Entity,
	modifier float64,
	originalCache map[uint64]float64,
	accessor StatAccessor,
) int {
	count := 0
	for _, entity := range entities {
		if !entity.HasComponent("stats") {
			continue
		}

		statsComp, ok := entity.GetComponent("stats")
		if !ok {
			continue
		}

		stats, ok := statsComp.(*StatsComponent)
		if !ok {
			continue
		}

		// Store original value if not already cached
		if _, exists := originalCache[entity.ID]; !exists {
			originalCache[entity.ID] = accessor.Get(stats)
		}

		// Apply modified value
		original := originalCache[entity.ID]
		newValue := original + modifier

		// Clamp to valid range (0.0 to 1.0)
		if newValue < 0.0 {
			newValue = 0.0
		}
		if newValue > 1.0 {
			newValue = 1.0
		}

		accessor.Set(stats, newValue)
		count++
	}
	return count
}
