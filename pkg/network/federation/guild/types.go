// Guild data structures and domain models.
//
// This file defines the core data structures for the guild system:
// - Member: Individual guild member with rank and join timestamps
// - TreasuryTransaction: Audit log for treasury operations
// - Emblem: Visual guild identity (colors, shape, symbol)
// - Guild: Complete guild state including members, treasury, permissions
// - GuildMessage: Federation protocol message envelope
// - Message payload types: MemberJoinData, MemberLeaveData, TerritoryChangeData
//
// All structures are JSON-serializable for network transmission and persistence.
// Constants (Rank, Permission, MessageType) have been moved to constants.go.
//
// INTEGRATION FIX [Category G]: Guild Federation Types
// Gap: Guild federation types missing (ROADMAP_V8.md Phase 50.1)
// Fix: Created complete type system for cross-server guild management
// Roadmap: ROADMAP_V8.md Phase 50.1
package guild

import "time"

// Member represents a guild member
type Member struct {
	PlayerID  string    `json:"player_id"`
	Rank      Rank      `json:"rank"`
	JoinedAt  time.Time `json:"joined_at"`
	LastLogin time.Time `json:"last_login"`
}

// TreasuryTransaction represents a treasury deposit or withdrawal
type TreasuryTransaction struct {
	PlayerID  string    `json:"player_id"`
	Amount    int       `json:"amount"` // Positive for deposit, negative for withdrawal
	Timestamp time.Time `json:"timestamp"`
	Reason    string    `json:"reason,omitempty"`
}

// Emblem represents a guild's visual identity
type Emblem struct {
	Shape      string `json:"shape"`     // "shield", "crest", "banner", "circle", "star"
	PrimaryR   uint8  `json:"primary_r"` // Primary color RGB
	PrimaryG   uint8  `json:"primary_g"`
	PrimaryB   uint8  `json:"primary_b"`
	SecondaryR uint8  `json:"secondary_r"` // Secondary color RGB
	SecondaryG uint8  `json:"secondary_g"`
	SecondaryB uint8  `json:"secondary_b"`
	Symbol     string `json:"symbol"` // "sword", "dragon", "star", "flame", "skull", etc.
}

// Guild represents a multi-server guild
type Guild struct {
	ID           string                `json:"id"`
	Name         string                `json:"name"`
	Emblem       *Emblem               `json:"emblem"`
	LeaderID     string                `json:"leader_id"`
	Members      []Member              `json:"members"`
	Permissions  map[Rank][]Permission `json:"permissions"`
	Treasury     int                   `json:"treasury"` // Shared gold pool
	Transactions []TreasuryTransaction `json:"transactions,omitempty"`
	MOTD         string                `json:"motd,omitempty"` // Message of the day
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
	Reputation   map[string]float64    `json:"reputation,omitempty"` // Server ID -> reputation score
}

// Type implements the guild type method
func (g *Guild) Type() string {
	return "guild"
}

// HasPermission checks if a rank has a specific permission in the given guild.
// It returns true if the given rank is granted the specified permission
// according to the guild's permission configuration. Returns false if
// the rank has no permissions configured or lacks the specific permission.
// This is a standalone function (not a method on Guild) to maintain ECS purity:
// Guild is a pure data component with no behavior methods beyond Type().
func HasPermission(g *Guild, rank Rank, perm Permission) bool {
	perms, exists := g.Permissions[rank]
	if !exists {
		return false
	}
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}

// GetMember returns a pointer to the Member with the given player ID in the guild.
// Returns nil if no member with that ID exists in the guild.
// The returned pointer points directly to the member in the guild's slice,
// so modifications will affect the guild state.
// This is a standalone function (not a method on Guild) to maintain ECS purity:
// Guild is a pure data component with no behavior methods beyond Type().
func GetMember(g *Guild, playerID string) *Member {
	for i := range g.Members {
		if g.Members[i].PlayerID == playerID {
			return &g.Members[i]
		}
	}
	return nil
}

// GuildMessage represents a cross-server guild synchronization message
// Originally located in types.go, federation message types kept together
type GuildMessage struct {
	Type      MessageType `json:"type"`
	GuildID   string      `json:"guild_id"`
	ServerID  string      `json:"server_id"` // Originating server
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"` // Payload varies by message type
}

// MemberJoinData contains data for member join messages
type MemberJoinData struct {
	PlayerID string `json:"player_id"`
	Rank     Rank   `json:"rank"`
}

// MemberLeaveData contains data for member leave messages
type MemberLeaveData struct {
	PlayerID string `json:"player_id"`
}

// TerritoryChangeData contains data for territory change messages
type TerritoryChangeData struct {
	ZoneID   string `json:"zone_id"`
	OldGuild string `json:"old_guild,omitempty"`
	NewGuild string `json:"new_guild"`
}
