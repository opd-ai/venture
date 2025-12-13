package engine

import (
	"time"

	"github.com/opd-ai/venture/pkg/world/raids"
)

// RaidInstanceComponent marks an entity as a raid instance portal or entrance.
// Attach to entities that players can interact with to enter raid dungeons.
type RaidInstanceComponent struct {
	InstanceID   string
	RaidDungeon  *raids.RaidDungeon
	Tier         raids.RaidTier
	GroupID      string
	PlayerIDs    []string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	Completed    bool
	ActiveBoss   int // Index of currently active boss (0-based)
	BossesKilled []bool
}

// Type returns the component type identifier.
func (r *RaidInstanceComponent) Type() string {
	return "raid_instance"
}

// RaidBossComponent marks an entity as a raid boss with special mechanics.
type RaidBossComponent struct {
	BossIndex     int
	CurrentPhase  int
	Mechanics     []raids.BossMechanic
	Phases        []raids.BossPhase
	MechanicTimer map[string]float64 // Time until mechanic ready
	PhaseEntered  bool
	InstanceID    string
}

// Type returns the component type identifier.
func (r *RaidBossComponent) Type() string {
	return "raid_boss"
}

// RaidLockoutComponent tracks a player's raid lockouts.
// Attach to player entities to enforce weekly lockouts.
type RaidLockoutComponent struct {
	Lockouts map[raids.RaidTier]*raids.PlayerLockout
}

// Type returns the component type identifier.
func (r *RaidLockoutComponent) Type() string {
	return "raid_lockout"
}

// NewRaidLockoutComponent creates a new lockout component with empty lockouts.
func NewRaidLockoutComponent() *RaidLockoutComponent {
	return &RaidLockoutComponent{
		Lockouts: make(map[raids.RaidTier]*raids.PlayerLockout),
	}
}

// IsLockedOut returns true if the player cannot run this raid tier yet.
func (r *RaidLockoutComponent) IsLockedOut(tier raids.RaidTier) bool {
	lockout, exists := r.Lockouts[tier]
	if !exists {
		return false
	}
	return time.Now().Before(lockout.NextReset)
}

// SetLockout marks the player as having completed this tier.
func (r *RaidLockoutComponent) SetLockout(playerID string, tier raids.RaidTier) {
	now := time.Now()
	nextReset := now.Add(7 * 24 * time.Hour) // Weekly reset

	r.Lockouts[tier] = &raids.PlayerLockout{
		PlayerID:  playerID,
		Tier:      tier,
		LastRun:   now,
		NextReset: nextReset,
	}
}
