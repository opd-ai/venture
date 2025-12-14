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

	// Full inventory item data
	Items []ItemData `json:"items"`

	// Gold currency
	Gold int `json:"gold"`

	// Full equipment data
	EquippedItems EquipmentData `json:"equipped_items"`

	// Mana (for spell casting)
	CurrentMana int `json:"current_mana"`
	MaxMana     int `json:"max_mana"`

	// Spell slots
	Spells []SpellData `json:"spells,omitempty"`

	// INTEGRATION FIX [Category D]: V8/V9 Feature Persistence
	// Gap: Housing plots, trust scores, reputation, guild membership, owned vehicles, and companions not persisted
	// Fix: Added persistence fields for V8.0 housing system and V9.0 integration features
	// Roadmap: ROADMAP_V8.md (Phase 49-53) and ROADMAP_V9.md (Phase 55-56)

	// Phase 49.1: Housing ownership data
	OwnedPlots []HousingPlotData `json:"owned_plots,omitempty"`

	// Phase 49.2: Trust & Reputation persistence
	TrustScores      map[string]float64 `json:"trust_scores,omitempty"`      // PlayerID -> Trust score
	ReputationScores map[string]int     `json:"reputation_scores,omitempty"` // Category -> Score

	// Phase 50.1: Guild membership
	GuildMembership *GuildMembershipData `json:"guild_membership,omitempty"`

	// Phase 50.3: Owned vehicles
	OwnedVehicles []VehicleData `json:"owned_vehicles,omitempty"`

	// Phase 22: Companion data
	ActiveCompanions []CompanionData `json:"active_companions,omitempty"`

	// GAP-003 REPAIR: Tutorial progress persistence
	TutorialState *TutorialStateData `json:"tutorial_state,omitempty"`

	// Phase 7.2: Animation state persistence
	AnimationState *AnimationStateData `json:"animation_state,omitempty"`
}

// TutorialStateData represents saved tutorial progress
// GAP-003 REPAIR: Allows tutorial state to persist across saves/loads
type TutorialStateData struct {
	Enabled        bool            `json:"enabled"`
	ShowUI         bool            `json:"show_ui"`
	CurrentStepIdx int             `json:"current_step_idx"`
	CompletedSteps map[string]bool `json:"completed_steps"` // Step ID -> completed
}

// AnimationStateData represents saved animation state for entities
// Phase 7.2: Allows animation states to persist across saves/loads
type AnimationStateData struct {
	// Current animation state (idle, walk, run, attack, etc.)
	State string `json:"state"`

	// Current frame index in the animation
	FrameIndex uint8 `json:"frame_index"`

	// Whether the animation should loop
	Loop bool `json:"loop"`

	// Timestamp of last frame update (for timing calculations)
	LastUpdateTime float64 `json:"last_update_time,omitempty"`
}

// INTEGRATION FIX [Category D]: V8/V9 Save Data Types
// Gap: Type definitions missing for V8/V9 feature persistence
// Fix: Added HousingPlotData, GuildMembershipData, VehicleData, CompanionData
// Roadmap: ROADMAP_V8.md and ROADMAP_V9.md

// HousingPlotData represents a player-owned housing plot.
type HousingPlotData struct {
	PlotID    string          `json:"plot_id"`
	X         float64         `json:"x"`
	Y         float64         `json:"y"`
	Width     float64         `json:"width"`
	Height    float64         `json:"height"`
	Tier      int             `json:"tier"` // Housing tier level
	Furniture []FurnitureData `json:"furniture,omitempty"`
}

// FurnitureData represents a furniture item in a housing plot.
type FurnitureData struct {
	FurnitureID string  `json:"furniture_id"`
	Type        string  `json:"type"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Rotation    float64 `json:"rotation,omitempty"`
}

// GuildMembershipData represents player's guild membership.
type GuildMembershipData struct {
	GuildID     string    `json:"guild_id"`
	GuildName   string    `json:"guild_name"`
	Rank        string    `json:"rank"`
	JoinedAt    time.Time `json:"joined_at"`
	Permissions []string  `json:"permissions,omitempty"`
}

// VehicleData represents an owned vehicle.
type VehicleData struct {
	VehicleID     string  `json:"vehicle_id"`
	Type          string  `json:"type"`
	Name          string  `json:"name"`
	Health        float64 `json:"health"`
	MaxHealth     float64 `json:"max_health"`
	Fuel          float64 `json:"fuel,omitempty"`
	MaxFuel       float64 `json:"max_fuel,omitempty"`
	Speed         float64 `json:"speed"`
	Durability    int     `json:"durability"`
	MaxDurability int     `json:"max_durability"`
	Seed          int64   `json:"seed"`
}

// CompanionData represents a player companion.
type CompanionData struct {
	EntityID      uint64     `json:"entity_id"`
	Name          string     `json:"name"`
	Type          string     `json:"type"`
	Level         int        `json:"level"`
	Experience    int        `json:"experience"`
	Health        float64    `json:"health"`
	MaxHealth     float64    `json:"max_health"`
	Attack        float64    `json:"attack"`
	Defense       float64    `json:"defense"`
	Loyalty       float64    `json:"loyalty"` // Companion loyalty (0.0-1.0)
	Equipment     []ItemData `json:"equipment,omitempty"`
	LearnedSkills []string   `json:"learned_skills,omitempty"`
	Personality   string     `json:"personality,omitempty"`
	Seed          int64      `json:"seed"`
}

// ItemData represents a serialized item for save files.
type ItemData struct {
	ID             string   `json:"id"`
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

	// INTEGRATION FIX [Category D]: V8/V9 World State Persistence
	// Gap: Guild halls, territory control, global events not persisted in world state
	// Fix: Added GuildHalls, TerritoryControl, ActiveEvents for persistent world features
	// Roadmap: ROADMAP_V8.md (Phase 51.2) and ROADMAP_V6.md (Phase 42)

	// Phase 51.2: Guild hall placements
	GuildHalls []GuildHallData `json:"guild_halls,omitempty"`

	// Phase 42: Territory control state
	TerritoryControl map[string]string `json:"territory_control,omitempty"` // ZoneID -> GuildID

	// Phase 42: Active meta-game events
	ActiveEvents []WorldEventData `json:"active_events,omitempty"`

	// Phase 70: Living World Memory (V11.0)
	// Persists city states, NPC states, world events, and player reputations
	LivingWorldMemory *LivingWorldMemoryData `json:"living_world_memory,omitempty"`
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

	// Phase 7.2: Animation state for entity
	AnimationState *AnimationStateData `json:"animation_state,omitempty"`
}

// INTEGRATION FIX [Category D]: V8/V9 World Data Types
// Gap: Type definitions for guild halls and world events missing
// Fix: Added GuildHallData and WorldEventData for persistent world state
// Roadmap: ROADMAP_V8.md (Phase 51.2) and ROADMAP_V6.md (Phase 42)

// GuildHallData represents a guild hall placement in the world.
type GuildHallData struct {
	GuildID  string     `json:"guild_id"`
	X        float64    `json:"x"`
	Y        float64    `json:"y"`
	Width    float64    `json:"width"`
	Height   float64    `json:"height"`
	Tier     int        `json:"tier"`
	PlacedAt time.Time  `json:"placed_at"`
	Rooms    []RoomData `json:"rooms,omitempty"`
}

// RoomData represents a room within a guild hall.
type RoomData struct {
	RoomID string  `json:"room_id"`
	Type   string  `json:"type"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Level  int     `json:"level,omitempty"`
}

// WorldEventData represents an active meta-game event.
type WorldEventData struct {
	EventID      string                 `json:"event_id"`
	Type         string                 `json:"type"`
	StartedAt    time.Time              `json:"started_at"`
	ExpiresAt    time.Time              `json:"expires_at"`
	Participants []string               `json:"participants,omitempty"`
	State        map[string]interface{} `json:"state,omitempty"`
}

// LivingWorldMemoryData represents the Phase 70 living world state for persistence.
// This tracks city states, NPC states, world events, and player reputations
// across game sessions to create a persistent, evolving world.
type LivingWorldMemoryData struct {
	// WorldSeed is the deterministic seed for this world
	WorldSeed int64 `json:"world_seed"`
	// LastSaveTime is the in-game time when world was last saved
	LastSaveTime float64 `json:"last_save_time"`
	// CityStates stores serialized city state data by city ID
	CityStates map[string]LivingCityStateData `json:"city_states,omitempty"`
	// NPCStates stores serialized NPC state data by entity ID
	NPCStates map[string]LivingNPCStateData `json:"npc_states,omitempty"`
	// EventHistory stores recent significant world events
	EventHistory []LivingWorldEventRecord `json:"event_history,omitempty"`
	// PlayerReputations stores per-city reputation by player ID
	PlayerReputations map[string]map[string]float64 `json:"player_reputations,omitempty"`
	// TimeProgressionEnabled allows world to advance while player is away
	TimeProgressionEnabled bool `json:"time_progression_enabled"`
	// TimeProgressionRate is the multiplier for offline time advancement
	TimeProgressionRate float64 `json:"time_progression_rate"`
}

// LivingCityStateData represents serialized city state for living world persistence.
type LivingCityStateData struct {
	CityID            string  `json:"city_id"`
	CityName          string  `json:"city_name"`
	Prosperity        float64 `json:"prosperity"`
	Population        int     `json:"population"`
	MaxPopulation     int     `json:"max_population"`
	Infrastructure    float64 `json:"infrastructure"`
	Defense           float64 `json:"defense"`
	State             string  `json:"state"`
	TradeVolume       float64 `json:"trade_volume"`
	ResourceStockpile float64 `json:"resource_stockpile"`
	Seed              int64   `json:"seed"`
}

// LivingNPCStateData represents serialized NPC state for living world persistence.
type LivingNPCStateData struct {
	EntityID           string                        `json:"entity_id"`
	Name               string                        `json:"name"`
	X                  float64                       `json:"x"`
	Y                  float64                       `json:"y"`
	HomeX              float64                       `json:"home_x"`
	HomeY              float64                       `json:"home_y"`
	CurrentActivityIdx int                           `json:"current_activity_idx"`
	Schedule           []LivingScheduledActivityData `json:"schedule,omitempty"`
	IsMoving           bool                          `json:"is_moving"`
	LastUpdateHour     int                           `json:"last_update_hour"`
}

// LivingScheduledActivityData represents a scheduled activity for NPC persistence.
type LivingScheduledActivityData struct {
	ActivityType string  `json:"activity_type"`
	StartHour    int     `json:"start_hour"`
	EndHour      int     `json:"end_hour"`
	LocationX    float64 `json:"location_x"`
	LocationY    float64 `json:"location_y"`
	LocationName string  `json:"location_name"`
}

// LivingWorldEventRecord represents a significant world event for history tracking.
type LivingWorldEventRecord struct {
	EventID          string                 `json:"event_id"`
	EventType        string                 `json:"event_type"`
	Description      string                 `json:"description"`
	GameTime         float64                `json:"game_time"`
	AffectedCityID   string                 `json:"affected_city_id,omitempty"`
	AffectedPlayerID string                 `json:"affected_player_id,omitempty"`
	Magnitude        float64                `json:"magnitude"`
	Details          map[string]interface{} `json:"details,omitempty"`
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
			Items: make([]ItemData, 0),
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
