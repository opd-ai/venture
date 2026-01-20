// Package guild constants and type aliases.
//
// This file contains all constant definitions used throughout the guild system:
// - Rank types (Recruit, Member, Officer, Leader)
// - Permission types (Invite, Kick, Promote, Withdraw, etc.)
// - MessageType for federation protocol
// - Configuration constants (MaxGuildDataSize)
//
// Code relocated from: types.go, manager.go
package guild

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

// MessageType represents the type of guild federation message
type MessageType string

const (
	// MsgTypeGuildSync synchronizes full guild state
	MsgTypeGuildSync MessageType = "guild_sync"
	// MsgTypeMemberJoin notifies of a new member joining
	MsgTypeMemberJoin MessageType = "member_join"
	// MsgTypeMemberLeave notifies of a member leaving
	MsgTypeMemberLeave MessageType = "member_leave"
	// MsgTypeTerritoryChange notifies of territory control change
	MsgTypeTerritoryChange MessageType = "territory_change"
)

// MaxGuildDataSize is the maximum allowed size for decompressed guild data.
// This protects against decompression bomb attacks. Set to 50MB to accommodate
// large guild databases while preventing memory exhaustion.
// Originally defined in: manager.go
const MaxGuildDataSize = 50 * 1024 * 1024 // 50 MB
