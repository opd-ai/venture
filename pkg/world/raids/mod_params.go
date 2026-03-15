package raids

import "time"

// ModParams holds data-driven overrides for raid balance parameters.
// These are applied at runtime and allow JSON-based mods (via pkg/modding)
// to tune raid difficulty, lockout periods, and boss mechanics without
// recompiling the game binary.
//
// Usage:
//
//	mgr := raids.NewManager(seed, "fantasy")
//	mgr.SetModParams(raids.ModParams{
//	    DifficultyMultipliers: map[string]float64{"normal": 0.8},
//	    LockoutPeriod:         3 * 24 * time.Hour,
//	})
type ModParams struct {
	// DifficultyMultipliers overrides per-tier difficulty scaling.
	// Keys are tier names (e.g., "normal", "heroic", "mythic").
	// Values are multipliers applied on top of the base tier scaling.
	// Zero or negative values are ignored; use the base value instead.
	DifficultyMultipliers map[string]float64

	// LockoutPeriod overrides the default 7-day lockout duration.
	// Zero value leaves the current lockout period unchanged.
	LockoutPeriod time.Duration

	// BossHealthMultiplier scales boss HP for all generated raids.
	// Values ≤ 0 are ignored.
	BossHealthMultiplier float64

	// BossAttackMultiplier scales boss attack for all generated raids.
	// Values ≤ 0 are ignored.
	BossAttackMultiplier float64
}

// SetModParams applies data-driven balance overrides to the raid manager.
// This is safe to call at any time; changes take effect on the next
// GenerateRaid or CreateInstance call. Thread-safe.
func (m *Manager) SetModParams(params ModParams) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.modParams = params

	if params.LockoutPeriod > 0 {
		m.lockoutManager.SetLockoutPeriod(params.LockoutPeriod)
	}
}

// applyModDifficulty applies DifficultyMultipliers to a base difficulty value.
// Must be called with at least a read lock held.
func (m *Manager) applyModDifficulty(tier RaidTier, base float64) float64 {
	if len(m.modParams.DifficultyMultipliers) == 0 {
		return base
	}
	if mult, ok := m.modParams.DifficultyMultipliers[tier.String()]; ok && mult > 0 {
		return base * mult
	}
	return base
}
