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

// TerritoryConfig holds configurable values for territory mechanics.
// These can be modified by mods to customize gameplay.
type TerritoryConfig struct {
	// Capture timing
	BaseCaptureTime   int // Seconds to capture a territory (default: 60)
	DefenderTimeBonus int // Additional seconds per defender (default: 30)

	// Bonuses
	BaseResourceBonus float64 // Resource bonus per territory (default: 0.10 = 10%)
	BaseXPBonus       float64 // XP bonus per territory (default: 0.05 = 5%)

	// War costs
	WarDeclarationCost   int // Gold to declare war (default: 1000)
	PeaceDeclarationCost int // Gold to declare peace (default: 500)
	SurrenderCost        int // Gold to surrender (default: 250)
	WarDurationDays      int // Duration of war in days (default: 7)

	// Structure stats
	WallBaseHP  float64 // Wall HP (default: 1000)
	TowerBaseHP float64 // Tower HP (default: 500)
	GuardBaseHP float64 // Guard HP (default: 500)
	TowerDamage float64 // Tower damage (default: 100)
	GuardLevel  int     // Guard NPC level (default: 30)

	// Siege settings
	GuildHallMaxHP         float64 // Guild hall max HP (default: 10000)
	TotalControlPoints     int     // Control points in a siege (default: 5)
	LootPercentage         float64 // Percentage of treasury looted (default: 0.15 = 15%)
	MaxSiegeParticipants   int     // Max players per siege (default: 100)
	MaxReinforcementGuilds int     // Max allied guilds (default: 5)
}

// DefaultTerritoryConfig returns the default configuration values.
// These match the original constant values for backward compatibility.
func DefaultTerritoryConfig() *TerritoryConfig {
	return &TerritoryConfig{
		BaseCaptureTime:        BaseCaptureTime,
		DefenderTimeBonus:      DefenderTimeBonus,
		BaseResourceBonus:      BaseResourceBonus,
		BaseXPBonus:            BaseXPBonus,
		WarDeclarationCost:     WarDeclarationCost,
		PeaceDeclarationCost:   PeaceDeclarationCost,
		SurrenderCost:          SurrenderCost,
		WarDurationDays:        WarDurationDays,
		WallBaseHP:             WallBaseHP,
		TowerBaseHP:            TowerBaseHP,
		GuardBaseHP:            GuardBaseHP,
		TowerDamage:            TowerDamage,
		GuardLevel:             GuardLevel,
		GuildHallMaxHP:         10000.0,
		TotalControlPoints:     5,
		LootPercentage:         0.15,
		MaxSiegeParticipants:   100,
		MaxReinforcementGuilds: 5,
	}
}

// Constants for territory mechanics (default values, can be overridden via TerritoryConfig)
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
