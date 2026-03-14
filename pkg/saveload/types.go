// Package saveload provides type definitions for save data.
// This file defines save game data structures including player state,
// inventory, and world information for persistence.
package saveload

import (
	"time"

	"github.com/opd-ai/venture/pkg/version"
)

// SaveVersion represents the save file format version.
// It mirrors version.Version to ensure a single source of truth.
var SaveVersion = version.Version

// Manager defines the interface for save/load operations.
// Both SaveManager (file-based) and MemorySaveManager (in-memory fallback)
// implement this interface.
type Manager interface {
	// SaveGame saves the game state with the given name.
	SaveGame(name string, save *GameSave) error
	// LoadGame loads the game state with the given name.
	LoadGame(name string) (*GameSave, error)
	// DeleteSave deletes a save with the given name.
	DeleteSave(name string) error
	// ListSaves returns metadata for all saves.
	ListSaves() ([]*SaveMetadata, error)
	// GetSaveMetadata returns metadata for a specific save.
	GetSaveMetadata(name string) (*SaveMetadata, error)
	// SaveExists checks if a save exists.
	SaveExists(name string) bool
	// SetMigrator sets the migrator for version upgrades.
	SetMigrator(migrator Migrator)
}

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

	// Phase 4.5: Onboarding flow persistence
	OnboardingState *OnboardingStateData `json:"onboarding_state,omitempty"`

	// Phase 4.5: Context-sensitive tutorial persistence
	ContextTutorialState *ContextTutorialStateData `json:"context_tutorial_state,omitempty"`

	// Phase 7.2: Animation state persistence
	AnimationState *AnimationStateData `json:"animation_state,omitempty"`

	// Phase 74: Event reward persistence (V12.0 Seasonal Events)
	EventRewardData *EventRewardStateData `json:"event_reward_data,omitempty"`

	// Phase 84: Player statistics persistence (V15.0)
	PlayerStatistics *PlayerStatisticsData `json:"player_statistics,omitempty"`

	// Phase 98: Daily/weekly challenge persistence (V18.0)
	ChallengeData *ChallengeStateData `json:"challenge_data,omitempty"`

	// Phase 111: New Game Plus persistence (V22.0)
	NewGamePlusData *NewGamePlusStateData `json:"newgameplus_data,omitempty"`

	// Phase 112: Carry-Over persistence (V22.0)
	CarryOverData *CarryOverStateData `json:"carryover_data,omitempty"`

	// Phase 114: NG+ Reward persistence (V22.0)
	NGPlusRewardData *NGPlusRewardStateData `json:"ngplus_reward_data,omitempty"`

	// QoL preferences persistence (AUDIT.md fix)
	// Allows player QoL settings (auto-loot, craft queue, sort preset, etc.) to persist across saves
	QoLData *QoLStateData `json:"qol_data,omitempty"`
}

// TutorialStateData represents saved tutorial progress
// GAP-003 REPAIR: Allows tutorial state to persist across saves/loads
type TutorialStateData struct {
	Enabled        bool            `json:"enabled"`
	ShowUI         bool            `json:"show_ui"`
	CurrentStepIdx int             `json:"current_step_idx"`
	CompletedSteps map[string]bool `json:"completed_steps"` // Step ID -> completed
}

// OnboardingStateData represents saved onboarding flow progress.
// Phase 4.5: Allows onboarding state to persist across saves/loads.
type OnboardingStateData struct {
	CurrentState int  `json:"current_state"` // OnboardingState enum value
	Enabled      bool `json:"enabled"`
	Skipped      bool `json:"skipped"`
	PlayerClass  int  `json:"player_class"` // CharacterClass enum value
}

// ContextTutorialStateData represents saved context-sensitive tutorial progress.
// Phase 4.5: Allows TutorialManager viewed topics to persist across saves/loads.
type ContextTutorialStateData struct {
	Enabled      bool             `json:"enabled"`
	ViewedTopics map[string]int64 `json:"viewed_topics"` // Topic string -> Unix timestamp when viewed
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

// EventRewardStateData represents saved event reward progress for Phase 74 (V12.0).
// This allows event currency, rewards, achievements, and cosmetics to persist across saves.
type EventRewardStateData struct {
	// EventCurrency tracks currency per event (eventID -> amount)
	EventCurrency map[string]int `json:"event_currency,omitempty"`
	// EarnedRewards contains all rewards earned from events (serialized as JSON)
	EarnedRewardsJSON []byte `json:"earned_rewards_json,omitempty"`
	// EarnedTitles contains unlocked cosmetic titles (serialized as JSON)
	EarnedTitlesJSON []byte `json:"earned_titles_json,omitempty"`
	// ActiveTitle is the currently displayed title ID
	ActiveTitle string `json:"active_title,omitempty"`
	// EarnedEffects contains unlocked visual effects (serialized as JSON)
	EarnedEffectsJSON []byte `json:"earned_effects_json,omitempty"`
	// ActiveEffect is the currently displayed effect ID
	ActiveEffect string `json:"active_effect,omitempty"`
	// CompletedAchievements contains finished achievement IDs
	CompletedAchievements []string `json:"completed_achievements,omitempty"`
	// TotalEventsParticipated is lifetime event count
	TotalEventsParticipated int `json:"total_events_participated"`
	// TotalQuestsCompleted is lifetime event quests completed
	TotalQuestsCompleted int `json:"total_quests_completed"`
	// TotalCurrencyEarned is lifetime currency earned across all events
	TotalCurrencyEarned int `json:"total_currency_earned"`
}

// PlayerStatisticsData represents saved player statistics for Phase 84 (V15.0).
// This allows lifetime and session statistics to persist across saves.
type PlayerStatisticsData struct {
	// Lifetime contains all lifetime statistics (stat ID -> value).
	Lifetime map[string]int64 `json:"lifetime,omitempty"`
	// FirstPlayTime is the Unix timestamp of the first play session.
	FirstPlayTime int64 `json:"first_play_time"`
	// TotalPlayTime is the total playtime in seconds (lifetime).
	TotalPlayTime int64 `json:"total_play_time"`
}

// ChallengeStateData represents saved challenge progress for Phase 98 (V18.0).
// This allows daily/weekly challenge progress to persist across saves.
type ChallengeStateData struct {
	// ActiveDailyChallengesJSON is serialized daily challenge data.
	ActiveDailyChallengesJSON []byte `json:"active_daily_challenges_json,omitempty"`
	// ActiveWeeklyChallengesJSON is serialized weekly challenge data.
	ActiveWeeklyChallengesJSON []byte `json:"active_weekly_challenges_json,omitempty"`
	// DailyStreak is consecutive days with all dailies completed.
	DailyStreak int `json:"daily_streak"`
	// LongestDailyStreak is the all-time highest daily streak.
	LongestDailyStreak int `json:"longest_daily_streak"`
	// WeeklyStreak is consecutive weeks with all weeklies completed.
	WeeklyStreak int `json:"weekly_streak"`
	// LongestWeeklyStreak is the all-time highest weekly streak.
	LongestWeeklyStreak int `json:"longest_weekly_streak"`
	// TotalChallengesCompleted is lifetime completed challenges.
	TotalChallengesCompleted int `json:"total_challenges_completed"`
	// TotalXPEarned is lifetime XP from challenges.
	TotalXPEarned int `json:"total_xp_earned"`
	// TotalGoldEarned is lifetime gold from challenges.
	TotalGoldEarned int `json:"total_gold_earned"`
	// LastDailyReset is unix timestamp of last daily reset.
	LastDailyReset int64 `json:"last_daily_reset"`
	// LastWeeklyReset is unix timestamp of last weekly reset.
	LastWeeklyReset int64 `json:"last_weekly_reset"`
	// BaseSeed is the base seed for deterministic generation.
	BaseSeed int64 `json:"base_seed"`
}

// NewGamePlusStateData represents saved NG+ progression for Phase 111 (V22.0).
// This allows NG+ state to persist across saves and carry over between cycles.
type NewGamePlusStateData struct {
	// Cycle is the current NG+ cycle (0 = first playthrough)
	Cycle int `json:"cycle"`
	// MaxCycleReached is the highest NG+ cycle ever achieved
	MaxCycleReached int `json:"max_cycle_reached"`
	// LegacyStats accumulates statistics across all playthroughs
	LegacyStats map[string]int64 `json:"legacy_stats,omitempty"`
	// TotalPlaytime is cumulative playtime in seconds across all cycles
	TotalPlaytime int64 `json:"total_playtime"`
	// CycleStartTime is Unix timestamp when current cycle started
	CycleStartTime int64 `json:"cycle_start_time"`
	// CurrentCyclePlaytime is playtime in current cycle (seconds)
	CurrentCyclePlaytime int64 `json:"current_cycle_playtime"`
	// CarryOverSlots is equipment carry-over slots unlocked
	CarryOverSlots int `json:"carry_over_slots"`
	// UnlockedBonuses lists permanent bonuses earned
	UnlockedBonuses []string `json:"unlocked_bonuses,omitempty"`
	// CurrencyCarryOverPercent is currency carry-over percentage
	CurrencyCarryOverPercent float64 `json:"currency_carry_over_percent"`
	// CompletedCyclesJSON is serialized completed cycle records
	CompletedCyclesJSON []byte `json:"completed_cycles_json,omitempty"`
}

// CarryOverStateData represents saved carry-over selections for Phase 112 (V22.0).
// This allows carry-over state to persist and transfer between NG+ cycles.
type CarryOverStateData struct {
	// SelectedEquipment contains item IDs selected for carry-over
	SelectedEquipment []string `json:"selected_equipment,omitempty"`
	// CurrencyCarryOver tracks currency type to amount
	CurrencyCarryOver map[string]int64 `json:"currency_carry_over,omitempty"`
	// SkillsToKeep contains skill IDs that will be preserved
	SkillsToKeep []string `json:"skills_to_keep,omitempty"`
	// CosmeticsUnlocked contains all unlocked cosmetics (always carries over)
	CosmeticsUnlocked []string `json:"cosmetics_unlocked,omitempty"`
	// AchievementsUnlocked contains all unlocked achievements (always carries over)
	AchievementsUnlocked []string `json:"achievements_unlocked,omitempty"`
	// SelectionLocked indicates selections cannot be changed
	SelectionLocked bool `json:"selection_locked"`
	// SelectionConfirmed indicates player confirmed selections
	SelectionConfirmed bool `json:"selection_confirmed"`
	// EquipmentSlotLimit is the maximum equipment carry-over slots
	EquipmentSlotLimit int `json:"equipment_slot_limit"`
	// SkillSlotLimit is the maximum skill carry-over slots
	SkillSlotLimit int `json:"skill_slot_limit"`
	// CurrencyPercentLimit is the maximum currency carry-over percentage
	CurrencyPercentLimit float64 `json:"currency_percent_limit"`
	// TransferComplete indicates the carry-over has been applied
	TransferComplete bool `json:"transfer_complete"`
}

// NGPlusRewardStateData represents saved NG+ exclusive rewards for Phase 114 (V22.0).
// This allows NG+ achievements, items, titles, and challenges to persist.
type NGPlusRewardStateData struct {
	// ExclusiveAchievements lists NG+ achievements earned
	ExclusiveAchievements []string `json:"exclusive_achievements,omitempty"`
	// ExclusiveItems lists NG+ exclusive item IDs acquired
	ExclusiveItems []string `json:"exclusive_items,omitempty"`
	// TitlesUnlocked lists NG+ exclusive titles earned
	TitlesUnlocked []string `json:"titles_unlocked,omitempty"`
	// ChallengesCompleted tracks challenge ID -> completion status (as JSON)
	ChallengesCompletedJSON []byte `json:"challenges_completed_json,omitempty"`
	// ChallengesActive tracks currently active challenge IDs
	ChallengesActive []string `json:"challenges_active,omitempty"`
	// HighestTierReached is the highest NG+ tier for tiered rewards
	HighestTierReached int `json:"highest_tier_reached"`
	// CurrentTitle is the currently equipped title
	CurrentTitle string `json:"current_title,omitempty"`
	// TimeAttackBestTimesJSON is serialized best times for time attacks
	TimeAttackBestTimesJSON []byte `json:"time_attack_best_times_json,omitempty"`
	// NoDeathRunProgressJSON is serialized no-death run progress
	NoDeathRunProgressJSON []byte `json:"no_death_run_progress_json,omitempty"`
	// NPCDialogVariationsUnlocked tracks unlocked dialog variations
	NPCDialogVariationsUnlocked []string `json:"npc_dialog_variations_unlocked,omitempty"`
}

// QoLStateData represents saved QoL (Quality of Life) preferences.
// AUDIT.md fix: Allows player QoL settings to persist across saves/loads.
type QoLStateData struct {
	// PlayerID for the player entity
	PlayerID uint64 `json:"player_id"`
	// AutoLootEnabled toggles automatic item collection
	AutoLootEnabled bool `json:"auto_loot_enabled"`
	// AutoLootRadius is the collection radius in tiles
	AutoLootRadius float64 `json:"auto_loot_radius"`
	// CraftQueue is the list of recipes in the crafting queue (serialized as JSON)
	CraftQueueJSON []byte `json:"craft_queue_json,omitempty"`
	// SortPreset is the selected inventory sort preset name
	SortPreset string `json:"sort_preset,omitempty"`
	// MountWhistle enables one-button mount summoning
	MountWhistle bool `json:"mount_whistle"`
	// RecipeTracking enables recipe ingredient tracking UI
	RecipeTracking bool `json:"recipe_tracking"`
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
