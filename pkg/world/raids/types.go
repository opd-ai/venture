package raids

import (
	"time"

	"github.com/opd-ai/venture/pkg/procgen/entity"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// TimeProvider is an interface for obtaining the current time.
// This enables deterministic timestamps for testing and networked multiplayer.
// In production, use RealTimeProvider; in tests, use FixedTimeProvider.
//
// Note: This abstraction is necessary for multiplayer to prevent time drift
// between federated servers, which would cause lockout/instance desync.
type TimeProvider interface {
	// Now returns the current time
	Now() time.Time
}

// RealTimeProvider implements TimeProvider using the actual system clock.
// This is the default implementation for production use.
type RealTimeProvider struct{}

// Now returns the current system time.
func (RealTimeProvider) Now() time.Time {
	return time.Now()
}

// FixedTimeProvider implements TimeProvider with a fixed time for testing.
type FixedTimeProvider struct {
	FixedTime time.Time
}

// Now returns the fixed time.
func (f FixedTimeProvider) Now() time.Time {
	return f.FixedTime
}

// DefaultTimeProvider returns the default TimeProvider (real system time).
func DefaultTimeProvider() TimeProvider {
	return RealTimeProvider{}
}

// RaidTier represents the difficulty level of a raid.
type RaidTier int

const (
	TierNormal RaidTier = iota
	TierHeroic
	TierMythic
	TierLegendary
	TierNightmare
)

// String returns the human-readable name of the raid tier.
func (t RaidTier) String() string {
	switch t {
	case TierNormal:
		return "Normal"
	case TierHeroic:
		return "Heroic"
	case TierMythic:
		return "Mythic"
	case TierLegendary:
		return "Legendary"
	case TierNightmare:
		return "Nightmare"
	default:
		return "Unknown"
	}
}

// DifficultyMultiplier returns the difficulty scaling factor for this tier.
func (t RaidTier) DifficultyMultiplier() float64 {
	switch t {
	case TierNormal:
		return 2.0
	case TierHeroic:
		return 4.0
	case TierMythic:
		return 6.0
	case TierLegendary:
		return 8.0
	case TierNightmare:
		return 10.0
	default:
		return 2.0
	}
}

// MinPlayers returns the minimum group size for this tier.
func (t RaidTier) MinPlayers() int {
	switch t {
	case TierNormal:
		return 5
	case TierHeroic:
		return 6
	case TierMythic:
		return 7
	case TierLegendary:
		return 8
	case TierNightmare:
		return 10
	default:
		return 5
	}
}

// MaxPlayers returns the maximum group size for this tier.
func (t RaidTier) MaxPlayers() int {
	switch t {
	case TierNormal:
		return 8
	case TierHeroic:
		return 9
	case TierMythic:
		return 10
	case TierLegendary:
		return 10
	case TierNightmare:
		return 10
	default:
		return 10
	}
}

// RaidDungeon represents a complete raid instance with bosses and layout.
type RaidDungeon struct {
	ID          string
	Name        string
	Description string
	Tier        RaidTier
	Terrain     *terrain.Terrain
	Bosses      []*RaidBoss
	Rooms       []*RaidRoom
	CreatedAt   time.Time
	Seed        int64
}

// RaidBoss represents a raid boss encounter with mechanics.
type RaidBoss struct {
	Entity    *entity.Entity
	RoomID    string
	Mechanics []BossMechanic
	Phases    []BossPhase
	Position  Position
	LootTable *LootTable
}

// BossMechanic represents a special ability or mechanic the boss uses.
type BossMechanic struct {
	ID          string
	Name        string
	Description string
	Type        MechanicType
	Cooldown    time.Duration
	Damage      int
	AoE         bool
	Radius      float64
}

// MechanicType categorizes boss mechanics.
type MechanicType int

const (
	MechanicSummon MechanicType = iota
	MechanicGroundEffect
	MechanicDebuff
	MechanicBuff
	MechanicChanneled
	MechanicInstant
	MechanicPeriodic
)

// String returns the human-readable name of the mechanic type.
func (m MechanicType) String() string {
	switch m {
	case MechanicSummon:
		return "Summon"
	case MechanicGroundEffect:
		return "GroundEffect"
	case MechanicDebuff:
		return "Debuff"
	case MechanicBuff:
		return "Buff"
	case MechanicChanneled:
		return "Channeled"
	case MechanicInstant:
		return "Instant"
	case MechanicPeriodic:
		return "Periodic"
	default:
		return "Unknown"
	}
}

// BossPhase represents a phase of the boss fight.
type BossPhase struct {
	Number       int
	HealthThresh float64
	Mechanics    []string
	AddSpawns    int
}

// RaidRoom represents a room in the raid dungeon.
type RaidRoom struct {
	ID          string
	Type        RoomType
	X, Y, W, H  int
	Connections []string
	BossID      string
}

// RoomType categorizes raid rooms.
type RoomType int

const (
	RoomEntrance RoomType = iota
	RoomBoss
	RoomTrash
	RoomTreasure
	RoomPuzzle
	RoomRest
)

// String returns the human-readable name of the room type.
func (r RoomType) String() string {
	switch r {
	case RoomEntrance:
		return "Entrance"
	case RoomBoss:
		return "Boss"
	case RoomTrash:
		return "Trash"
	case RoomTreasure:
		return "Treasure"
	case RoomPuzzle:
		return "Puzzle"
	case RoomRest:
		return "Rest"
	default:
		return "Unknown"
	}
}

// LootTable defines potential loot drops from a boss.
type LootTable struct {
	GuaranteedItems int
	PossibleItems   []LootItem
	CurrencyMin     int
	CurrencyMax     int
}

// LootItem represents a potential loot drop.
type LootItem struct {
	ItemID   string
	Rarity   string
	DropRate float64
}

// Position represents a 2D position in the raid.
type Position struct {
	X, Y float64
}

// RaidInstance tracks an active raid session.
type RaidInstance struct {
	InstanceID string
	RaidID     string
	Dungeon    *RaidDungeon
	GroupID    string
	PlayerIDs  []string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	Completed  bool
}

// PlayerLockout tracks when a player can next run a raid tier.
type PlayerLockout struct {
	PlayerID  string
	Tier      RaidTier
	LastRun   time.Time
	NextReset time.Time
}
