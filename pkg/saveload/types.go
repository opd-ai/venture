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

	// Inventory (item IDs)
	InventoryItems []uint64 `json:"inventory_items"`

	// Equipment (item IDs)
	EquippedWeapon    uint64 `json:"equipped_weapon,omitempty"`
	EquippedArmor     uint64 `json:"equipped_armor,omitempty"`
	EquippedAccessory uint64 `json:"equipped_accessory,omitempty"`
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
