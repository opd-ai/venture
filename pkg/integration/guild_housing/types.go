package guild_housing

import (
	"time"

	"github.com/opd-ai/venture/pkg/network/federation/guild"
	"github.com/opd-ai/venture/pkg/world/housing"
)

// Permission represents access levels for guild houses.
type Permission int

const (
	PermissionNone   Permission = iota // No access
	PermissionView                     // Can enter and view
	PermissionUse                      // Can use crafting stations
	PermissionManage                   // Can place/remove furniture
	PermissionAdmin                    // Full control
)

// String returns human-readable permission name.
func (p Permission) String() string {
	switch p {
	case PermissionNone:
		return "None"
	case PermissionView:
		return "View"
	case PermissionUse:
		return "Use"
	case PermissionManage:
		return "Manage"
	case PermissionAdmin:
		return "Admin"
	default:
		return "Unknown"
	}
}

// UpgradeTier represents guild house upgrade levels.
type UpgradeTier int

const (
	TierBasic    UpgradeTier = iota // Base tier, no bonuses
	TierStandard                    // 10k gold, +20% bonuses
	TierAdvanced                    // 50k gold, +50% bonuses
	TierMaster                      // 100k gold, +100% bonuses
)

// String returns human-readable tier name.
func (t UpgradeTier) String() string {
	switch t {
	case TierBasic:
		return "Basic"
	case TierStandard:
		return "Standard"
	case TierAdvanced:
		return "Advanced"
	case TierMaster:
		return "Master"
	default:
		return "Unknown"
	}
}

// Cost returns upgrade cost in gold.
func (t UpgradeTier) Cost() int {
	switch t {
	case TierBasic:
		return 0
	case TierStandard:
		return 10000
	case TierAdvanced:
		return 50000
	case TierMaster:
		return 100000
	default:
		return 0
	}
}

// BonusMultiplier returns upgrade bonus multiplier.
func (t UpgradeTier) BonusMultiplier() float64 {
	switch t {
	case TierBasic:
		return 1.0
	case TierStandard:
		return 1.2
	case TierAdvanced:
		return 1.5
	case TierMaster:
		return 2.0
	default:
		return 1.0
	}
}

// GuildHouse represents a guild-owned housing structure.
type GuildHouse struct {
	HouseID     string
	GuildID     string
	OwnerID     string
	Size        housing.BuildingSize
	Permissions map[guild.Rank]Permission
	Tier        UpgradeTier
	Stations    []string // Crafting station IDs
	Storage     *GuildStorage
	MeetingHall *MeetingHall
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// GuildStorage represents shared guild storage.
type GuildStorage struct {
	StorageID    string
	GuildID      string
	Capacity     int
	Items        map[string]*StoredItem
	Transactions []*Transaction
	CreatedAt    time.Time
}

// StoredItem represents an item in guild storage.
type StoredItem struct {
	ItemID   string
	Quantity int
	AddedBy  string
	AddedAt  time.Time
}

// Transaction represents a storage deposit/withdraw log entry.
type Transaction struct {
	TransactionID string
	PlayerID      string
	ItemID        string
	Quantity      int
	Action        TransactionType
	Timestamp     time.Time
}

// TransactionType represents deposit or withdrawal.
type TransactionType int

const (
	TransactionDeposit  TransactionType = iota // Item deposited
	TransactionWithdraw                        // Item withdrawn
)

// String returns human-readable transaction type.
func (t TransactionType) String() string {
	switch t {
	case TransactionDeposit:
		return "Deposit"
	case TransactionWithdraw:
		return "Withdraw"
	default:
		return "Unknown"
	}
}

// MeetingHall represents a guild meeting space.
type MeetingHall struct {
	HallID      string
	GuildID     string
	ChatRadius  float64
	MaxCapacity int
	Members     []string // Currently present members
	CreatedAt   time.Time
}

// GuildHousingComponent is an ECS component for guild housing integration.
type GuildHousingComponent struct {
	GuildID    string
	Permission Permission
	Tier       UpgradeTier
}

// Type returns the component type identifier.
func (c GuildHousingComponent) Type() string {
	return "guild_housing"
}
