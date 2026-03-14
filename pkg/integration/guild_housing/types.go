package guild_housing

import (
	"encoding/json"
	"time"

	"github.com/opd-ai/venture/pkg/network/federation/guild"
	"github.com/opd-ai/venture/pkg/world/housing"
)

// types.go defines core domain types for guild housing.
// This includes guild houses, storage, items, meeting halls, and the ECS component.

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

// Serialize encodes the GuildHousingComponent to JSON bytes.
func (c *GuildHousingComponent) Serialize() ([]byte, error) {
	return json.Marshal(c)
}

// Deserialize decodes JSON bytes into the GuildHousingComponent.
func (c *GuildHousingComponent) Deserialize(data []byte) error {
	return json.Unmarshal(data, c)
}
