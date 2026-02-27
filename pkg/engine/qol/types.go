// Package qol - types.go
// This file contains all type definitions, constants, and helper functions for QoL features.
// Note: time.Now() usage in this package is intentional. QoL features like guild invitations,
// mount summoning, and crafting queues are real-time gameplay mechanics that require actual
// wall-clock time for proper expiry, cooldowns, and UI display. This is distinct from
// procedural generation which must be deterministic.

package qol

import (
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
)

// SortCriteria defines how items should be sorted
type SortCriteria int

const (
	SortByType SortCriteria = iota
	SortByRarity
	SortByName
	SortByValue
	SortByQuantity
)

// String returns human-readable sort criteria name
func (s SortCriteria) String() string {
	switch s {
	case SortByType:
		return "Type"
	case SortByRarity:
		return "Rarity"
	case SortByName:
		return "Name"
	case SortByValue:
		return "Value"
	case SortByQuantity:
		return "Quantity"
	default:
		return "Unknown"
	}
}

// Item represents an inventory item for sorting operations.
// Used by StorageSorter to organize player inventory and storage containers.
type Item struct {
	ID       string
	Name     string
	Type     string
	Rarity   int
	Value    int
	Quantity int
}

// AutoLootConfig defines auto-loot behavior per companion
type AutoLootConfig struct {
	CompanionID    uint64
	Enabled        bool
	Radius         float64 // Tiles (5-10)
	MinRarity      int     // 0=Common, 1=Uncommon, 2=Rare, 3=Epic, 4=Legendary
	FilterTypes    []string
	IgnoreTypes    []string
	MaxPerCycle    int // Max items per collection cycle
	LastCollection time.Time
}

// CraftQueueEntry represents a recipe in the crafting queue
type CraftQueueEntry struct {
	RecipeID       string
	Quantity       int
	MaterialsReady bool
	Position       int // Queue position (0-based)
	AddedAt        time.Time
}

// GuildInvitation represents an offline guild invitation
type GuildInvitation struct {
	InvitationID string
	GuildID      string
	GuildName    string
	InviterID    string
	InviterName  string
	InviteeID    string
	InviteeName  string
	Message      string
	SentAt       time.Time
	ExpiresAt    time.Time
	Accepted     bool
	AcceptedAt   time.Time
}

// MountSummon represents a vehicle summon request
type MountSummon struct {
	PlayerID      uint64
	VehicleID     uint64
	VehicleType   string
	RequestTime   time.Time
	EstimatedTime float64 // Seconds
	CurrentPos    [2]float64
	TargetPos     [2]float64
	Distance      float64
	Completed     bool
}

// RecipeTrackingInfo shows recipe requirements and availability
type RecipeTrackingInfo struct {
	RecipeID        string
	RecipeName      string
	RequiredMats    map[string]int // MaterialID -> Quantity
	AvailableMats   map[string]int
	MissingMats     map[string]int
	CanCraft        bool
	MaxCraftable    int
	MaterialSources map[string][]string // MaterialID -> Source hints
}

// StorageSortPreset defines a reusable sort configuration
type StorageSortPreset struct {
	Name              string
	PrimaryCriteria   SortCriteria
	SecondaryCriteria SortCriteria
	Descending        bool
	GroupByType       bool
}

// QoLComponent is the ECS component for quality of life features.
// Stores player preferences for auto-loot, crafting, sorting, and other QoL settings.
type QoLComponent struct {
	PlayerID        uint64
	AutoLootEnabled bool
	AutoLootRadius  float64
	CraftQueue      []*CraftQueueEntry
	SortPreset      string
	MountWhistle    bool
	RecipeTracking  bool
}

// Type returns the component type identifier.
func (q QoLComponent) Type() string {
	return "qol"
}

// qolComponentData is the serialization format for QoLComponent.
type qolComponentData struct {
	PlayerID        uint64             `json:"player_id"`
	AutoLootEnabled bool               `json:"auto_loot_enabled"`
	AutoLootRadius  float64            `json:"auto_loot_radius"`
	CraftQueue      []*CraftQueueEntry `json:"craft_queue"`
	SortPreset      string             `json:"sort_preset"`
	MountWhistle    bool               `json:"mount_whistle"`
	RecipeTracking  bool               `json:"recipe_tracking"`
}

// Serialize converts the QoLComponent to JSON bytes for persistence.
func (q *QoLComponent) Serialize() ([]byte, error) {
	data := qolComponentData{
		PlayerID:        q.PlayerID,
		AutoLootEnabled: q.AutoLootEnabled,
		AutoLootRadius:  q.AutoLootRadius,
		CraftQueue:      q.CraftQueue,
		SortPreset:      q.SortPreset,
		MountWhistle:    q.MountWhistle,
		RecipeTracking:  q.RecipeTracking,
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"playerID":       q.PlayerID,
			"component_type": "qol",
			"error":          err.Error(),
		}).Error("Failed to serialize QoLComponent")
		return nil, err
	}
	logrus.WithFields(logrus.Fields{
		"playerID":       q.PlayerID,
		"component_type": "qol",
		"size_bytes":     len(bytes),
	}).Debug("Serialized QoLComponent")
	return bytes, nil
}

// Deserialize restores the QoLComponent from JSON bytes.
func (q *QoLComponent) Deserialize(data []byte) error {
	var d qolComponentData
	if err := json.Unmarshal(data, &d); err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "qol",
			"size_bytes":     len(data),
			"error":          err.Error(),
		}).Error("Failed to deserialize QoLComponent")
		return err
	}
	q.PlayerID = d.PlayerID
	q.AutoLootEnabled = d.AutoLootEnabled
	q.AutoLootRadius = d.AutoLootRadius
	q.CraftQueue = d.CraftQueue
	q.SortPreset = d.SortPreset
	q.MountWhistle = d.MountWhistle
	q.RecipeTracking = d.RecipeTracking
	logrus.WithFields(logrus.Fields{
		"playerID":       q.PlayerID,
		"component_type": "qol",
		"size_bytes":     len(data),
	}).Debug("Deserialized QoLComponent")
	return nil
}

// DefaultAutoLootConfig returns default auto-loot settings
func DefaultAutoLootConfig(companionID uint64) *AutoLootConfig {
	return &AutoLootConfig{
		CompanionID:    companionID,
		Enabled:        true,
		Radius:         7.0, // 7 tiles default
		MinRarity:      0,   // Common and above
		FilterTypes:    []string{},
		IgnoreTypes:    []string{"junk"},
		MaxPerCycle:    10,
		LastCollection: time.Time{},
	}
}

// IsExpired checks if a guild invitation has expired
func (g *GuildInvitation) IsExpired() bool {
	return time.Now().After(g.ExpiresAt)
}

// DaysUntilExpiry returns days remaining before expiration
func (g *GuildInvitation) DaysUntilExpiry() float64 {
	if g.IsExpired() {
		return 0
	}
	return time.Until(g.ExpiresAt).Hours() / 24.0
}

// EstimateArrivalTime calculates mount arrival time based on distance
// Formula: 1 second per tile, max 5 seconds
func EstimateArrivalTime(distance float64) float64 {
	if distance < 0 {
		distance = 0
	}
	time := distance * 1.0 // 1 second per tile
	if time > 5.0 {
		time = 5.0 // Max 5 seconds
	}
	return time
}
