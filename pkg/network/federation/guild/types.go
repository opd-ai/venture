package guild

import "time"

// INTEGRATION FIX [Category G]: Guild Federation Types
// Gap: Guild federation types missing (ROADMAP_V8.md Phase 50.1)
// Fix: Created complete type system for cross-server guild management
// Roadmap: ROADMAP_V8.md Phase 50.1

// Rank represents a guild member's rank
type Rank string

const (
	// RankRecruit is the entry-level rank
	RankRecruit Rank = "Recruit"
	// RankMember is the standard member rank
	RankMember Rank = "Member"
	// RankOfficer is the officer rank with elevated permissions
	RankOfficer Rank = "Officer"
	// RankLeader is the guild leader rank with full permissions
	RankLeader Rank = "Leader"
)

// Permission represents a guild permission type
type Permission string

const (
	// PermissionInvite allows inviting new members
	PermissionInvite Permission = "invite"
	// PermissionKick allows removing members
	PermissionKick Permission = "kick"
	// PermissionPromote allows promoting members
	PermissionPromote Permission = "promote"
	// PermissionDemote allows demoting members
	PermissionDemote Permission = "demote"
	// PermissionWithdraw allows withdrawing from guild treasury
	PermissionWithdraw Permission = "withdraw"
	// PermissionDeposit allows depositing to guild treasury
	PermissionDeposit Permission = "deposit"
	// PermissionEditMOTD allows editing message of the day
	PermissionEditMOTD Permission = "edit_motd"
	// PermissionManageBank allows managing guild bank items
	PermissionManageBank Permission = "manage_bank"
	// PermissionDeclareWar allows declaring wars on other guilds
	PermissionDeclareWar Permission = "declare_war"
)

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

// HasPermission checks if a rank has a specific permission
func (g *Guild) HasPermission(rank Rank, perm Permission) bool {
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

// GetMember returns a member by player ID
func (g *Guild) GetMember(playerID string) *Member {
	for i := range g.Members {
		if g.Members[i].PlayerID == playerID {
			return &g.Members[i]
		}
	}
	return nil
}
