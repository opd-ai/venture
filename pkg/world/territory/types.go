package territory

import "time"

// TimeProvider is an interface for obtaining the current time.
// This enables deterministic timestamps for testing and reproducible state.
// In production, use RealTimeProvider; in tests, use a mock implementation.
type TimeProvider interface {
	// Now returns the current time
	Now() time.Time
}

// RealTimeProvider implements TimeProvider using the actual system clock.
//
// INTENTIONAL time.Now() EXCEPTION: This type is specifically designed for
// territory/siege time management (war durations, capture progress timestamps,
// siege phase transitions) and does NOT affect procedural content generation.
// The time.Now() usage here is acceptable because:
//
//  1. Territory mechanics are server-side operations with authoritative time sources
//  2. The TimeProvider interface allows injection of deterministic mocks for testing
//  3. For deterministic server-to-server sync in federated scenarios, servers should
//     use a synchronized time source (NTP, consensus clock) rather than deterministic
//     generation seeds
//
// This differs from procedural generation (terrain, quests, NPCs) which must use
// seed-based RNGs for determinism. Territory state determinism is achieved through
// server authority and network replication, not through seed-based generation.
type RealTimeProvider struct{}

// Now returns the current system time.
//
// See RealTimeProvider godoc for explanation of time.Now() usage.
func (RealTimeProvider) Now() time.Time {
	return time.Now()
}

// DefaultTimeProvider returns the default TimeProvider (real system time).
func DefaultTimeProvider() TimeProvider {
	return RealTimeProvider{}
}

// TerritoryCoords represents the chunk coordinates of a territory (5×5 chunks).
type TerritoryCoords struct {
	ChunkX int
	ChunkZ int
}

// TerritoryStatus represents the ownership status of a territory.
type TerritoryStatus int

const (
	StatusNeutral TerritoryStatus = iota
	StatusOwned
	StatusContested
)

// String returns a human-readable name for the territory status.
func (s TerritoryStatus) String() string {
	switch s {
	case StatusNeutral:
		return "Neutral"
	case StatusOwned:
		return "Owned"
	case StatusContested:
		return "Contested"
	default:
		return "Unknown"
	}
}

// Territory represents a 5×5 chunk zone that can be owned by a guild.
type Territory struct {
	ID              string
	Coords          TerritoryCoords
	OwnerGuildID    string
	Status          TerritoryStatus
	CaptureProgress float64
	CapturingGuild  string
	LastUpdate      time.Time
	Structures      []*DefensiveStructure
	ResourceBonus   float64
	XPBonus         float64
}

// StructureType represents the type of defensive structure.
type StructureType int

const (
	StructureTypeWall StructureType = iota
	StructureTypeTower
	StructureTypeGuard
)

// String returns a human-readable name for the structure type.
func (s StructureType) String() string {
	switch s {
	case StructureTypeWall:
		return "Wall"
	case StructureTypeTower:
		return "Tower"
	case StructureTypeGuard:
		return "Guard"
	default:
		return "Unknown"
	}
}

// DefensiveStructure represents a defensive structure in a territory.
type DefensiveStructure struct {
	ID            string
	Type          StructureType
	X             float64
	Y             float64
	HP            float64
	MaxHP         float64
	Damage        float64
	Level         int
	ConstructedAt time.Time
}

// WarDeclaration represents a formal war between two guilds.
type WarDeclaration struct {
	ID            string
	AttackerGuild string
	DefenderGuild string
	DeclaredAt    time.Time
	EndsAt        time.Time
	Active        bool
	Cost          int
}

// Constants for territory mechanics
const (
	TerritoryChunkSize   = 5
	BaseCaptureTime      = 60
	DefenderTimeBonus    = 30
	BaseResourceBonus    = 0.10
	BaseXPBonus          = 0.05
	WarDeclarationCost   = 1000
	PeaceDeclarationCost = 500
	SurrenderCost        = 250 // Half of peace declaration cost
	WarDurationDays      = 7

	WallBaseHP  = 1000.0
	TowerBaseHP = 500.0
	GuardBaseHP = 500.0
	TowerDamage = 100.0
	GuardLevel  = 30
)
