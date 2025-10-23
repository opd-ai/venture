// Package saveload provides type definitions for save data.
// This file defines save game data structures including player state,
// inventory, and world information for persistence.
package saveload

import (
	"time"
)

// SaveVersion represents the save file format version.
const SaveVersion = "1.0.0"

// GameSave represents a complete save file with all game state.
type GameSave struct {
	// Version of the save file format
	Version string `json:"version"`

	// Timestamp when the save was created
	Timestamp time.Time `json:"timestamp"`

	// Player state
	PlayerState *PlayerState `json:"player"`

	// World state
	WorldState *WorldState `json:"world"`

	// Game settings
	Settings *GameSettings `json:"settings"`
}

// PlayerState represents all player-related state that needs to be saved.
type PlayerState struct {
	// Entity ID of the player
	EntityID uint64 `json:"entity_id"`

	// Position
	X float64 `json:"x"`
	Y float64 `json:"y"`

	// Health
	CurrentHealth float64 `json:"current_health"`
	MaxHealth     float64 `json:"max_health"`

	// Stats
	Level      int     `json:"level"`
	Experience int     `json:"experience"`
	Attack     float64 `json:"attack"`
	Defense    float64 `json:"defense"`
	MagicPower float64 `json:"magic_power"`
	Speed      float64 `json:"speed"`

	// Inventory (item IDs - DEPRECATED, use Items instead)
	InventoryItems []uint64 `json:"inventory_items,omitempty"`

	// GAP-007 REPAIR: Full inventory item data
	Items []ItemData `json:"items"`

	// GAP-009 REPAIR: Gold currency
	Gold int `json:"gold"`

	// Equipment (item IDs - DEPRECATED, use EquippedItems instead)
	EquippedWeapon    uint64 `json:"equipped_weapon,omitempty"`
	EquippedArmor     uint64 `json:"equipped_armor,omitempty"`
	EquippedAccessory uint64 `json:"equipped_accessory,omitempty"`

	// GAP-008 REPAIR: Full equipment data
	EquippedItems EquipmentData `json:"equipped_items"`

	// Mana (for spell casting)
	CurrentMana int `json:"current_mana"`
	MaxMana     int `json:"max_mana"`

	// Spell slots
	Spells []SpellData `json:"spells,omitempty"`
}

// ItemData represents a serialized item for save files.
type ItemData struct {
	Name           string   `json:"name"`
	Type           string   `json:"type"` // "weapon", "armor", "consumable", "accessory"
	WeaponType     string   `json:"weapon_type,omitempty"`
	ArmorType      string   `json:"armor_type,omitempty"`
	ConsumableType string   `json:"consumable_type,omitempty"`
	Rarity         string   `json:"rarity"` // "common", "uncommon", "rare", "epic", "legendary"
	Seed           int64    `json:"seed"`
	Tags           []string `json:"tags,omitempty"`
	Description    string   `json:"description,omitempty"`

	// Stats
	Damage        int     `json:"damage,omitempty"`
	Defense       int     `json:"defense,omitempty"`
	AttackSpeed   float64 `json:"attack_speed,omitempty"`
	Value         int     `json:"value"`
	Weight        float64 `json:"weight"`
	RequiredLevel int     `json:"required_level,omitempty"`
	DurabilityMax int     `json:"durability_max,omitempty"`
	Durability    int     `json:"durability,omitempty"`
}

// EquipmentData represents equipped items.
type EquipmentData struct {
	Weapon    *ItemData `json:"weapon,omitempty"`
	Armor     *ItemData `json:"armor,omitempty"`
	Accessory *ItemData `json:"accessory,omitempty"`
}

// SpellData represents a serialized spell for save files.
type SpellData struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`    // "offensive", "defensive", "healing", etc.
	Element     string   `json:"element"` // "fire", "ice", "lightning", etc.
	Target      string   `json:"target"`  // "self", "single", "area", etc.
	Rarity      string   `json:"rarity"`
	Seed        int64    `json:"seed"`
	Tags        []string `json:"tags,omitempty"`
	Description string   `json:"description,omitempty"`

	// Stats
	Damage   int     `json:"damage,omitempty"`
	Healing  int     `json:"healing,omitempty"`
	ManaCost int     `json:"mana_cost"`
	Cooldown float64 `json:"cooldown"`
	CastTime float64 `json:"cast_time"`
	Range    float64 `json:"range,omitempty"`
	AreaSize float64 `json:"area_size,omitempty"`
	Duration float64 `json:"duration,omitempty"`
}

// WorldState represents all world-related state that needs to be saved.
type WorldState struct {
	// World generation seed
	Seed int64 `json:"seed"`

	// Genre ID
	GenreID string `json:"genre_id"`

	// Map dimensions
	Width  int `json:"width"`
	Height int `json:"height"`

	// Game time (in seconds)
	GameTime float64 `json:"game_time"`

	// Current difficulty
	Difficulty float64 `json:"difficulty"`

	// Current depth (dungeon level)
	Depth int `json:"depth"`

	// GAP-005 REPAIR: Fog of war exploration state
	// 2D array where true = explored, false = unexplored
	// Serialized as nested arrays for JSON compatibility
	FogOfWar [][]bool `json:"fog_of_war,omitempty"`

	// Entity states (for NPCs, monsters, items in the world)
	// We store minimal info and rely on seed-based regeneration
	// for most entities, only saving what's been modified
	ModifiedEntities []ModifiedEntity `json:"modified_entities,omitempty"`
}

// ModifiedEntity represents an entity that has been modified from its
// procedurally generated state and needs to be saved.
type ModifiedEntity struct {
	EntityID uint64  `json:"entity_id"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Health   float64 `json:"health,omitempty"`
	IsAlive  bool    `json:"is_alive"`
	IsPicked bool    `json:"is_picked,omitempty"` // For items
}

// GameSettings represents game configuration that should persist.
type GameSettings struct {
	// Graphics settings
	ScreenWidth  int  `json:"screen_width"`
	ScreenHeight int  `json:"screen_height"`
	Fullscreen   bool `json:"fullscreen"`
	VSync        bool `json:"vsync"`

	// Audio settings
	MasterVolume float64 `json:"master_volume"`
	MusicVolume  float64 `json:"music_volume"`
	SFXVolume    float64 `json:"sfx_volume"`

	// Control settings
	KeyBindings map[string]string `json:"key_bindings,omitempty"`
}

// SaveMetadata provides summary information about a save file without loading it completely.
type SaveMetadata struct {
	// Save file name
	Name string `json:"name"`

	// Save file version
	Version string `json:"version"`

	// Creation timestamp
	Timestamp time.Time `json:"timestamp"`

	// Player level (for display in save list)
	PlayerLevel int `json:"player_level"`

	// World genre
	GenreID string `json:"genre_id"`

	// Game time
	GameTime float64 `json:"game_time"`

	// File size in bytes
	FileSize int64 `json:"file_size,omitempty"`
}

// NewGameSave creates a new GameSave with default values.
func NewGameSave() *GameSave {
	return &GameSave{
		Version:   SaveVersion,
		Timestamp: time.Now(),
		PlayerState: &PlayerState{
			InventoryItems: make([]uint64, 0),
		},
		WorldState: &WorldState{
			ModifiedEntities: make([]ModifiedEntity, 0),
		},
		Settings: &GameSettings{
			ScreenWidth:  800,
			ScreenHeight: 600,
			Fullscreen:   false,
			VSync:        true,
			MasterVolume: 1.0,
			MusicVolume:  0.7,
			SFXVolume:    0.8,
			KeyBindings:  make(map[string]string),
		},
	}
}
