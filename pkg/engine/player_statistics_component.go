// Package engine provides comprehensive player statistics tracking.
// This file implements the PlayerStatisticsComponent for tracking
// player accomplishments and metrics across all game systems.
//
// Phase 84: Player Statistics System (V15.0)
package engine

import (
	"encoding/json"
	"sync"
)

// StatCategory represents the category of a statistic.
type StatCategory int

const (
	// StatCategoryCombat tracks combat-related statistics.
	StatCategoryCombat StatCategory = iota
	// StatCategoryQuest tracks quest-related statistics.
	StatCategoryQuest
	// StatCategoryCrafting tracks crafting-related statistics.
	StatCategoryCrafting
	// StatCategoryExploration tracks exploration-related statistics.
	StatCategoryExploration
	// StatCategorySocial tracks social-related statistics.
	StatCategorySocial
	// StatCategoryPvP tracks PvP-related statistics.
	StatCategoryPvP
	// StatCategoryEconomy tracks economy-related statistics.
	StatCategoryEconomy
	// StatCategoryGeneral tracks general game statistics.
	StatCategoryGeneral
)

// String returns the string representation of the stat category.
func (c StatCategory) String() string {
	names := []string{"Combat", "Quest", "Crafting", "Exploration", "Social", "PvP", "Economy", "General"}
	if int(c) < len(names) {
		return names[c]
	}
	return "Unknown"
}

// StatDefinition defines a statistic's properties.
type StatDefinition struct {
	ID          string
	Name        string
	Description string
	Category    StatCategory
	// IsTime indicates if this stat tracks time (in seconds).
	IsTime bool
	// IsCounter indicates if this stat is a simple counter.
	IsCounter bool
	// IsMax indicates if this stat should track max value (for "best" stats).
	IsMax bool
}

// statDefinitions contains all statistic definitions.
// Over 40 statistics covering all game systems as required by Phase 84.
var statDefinitions = map[string]StatDefinition{
	// Combat statistics (10)
	"combat_enemies_killed": {
		ID:          "combat_enemies_killed",
		Name:        "Enemies Killed",
		Description: "Total enemies defeated in combat",
		Category:    StatCategoryCombat,
		IsCounter:   true,
	},
	"combat_bosses_killed": {
		ID:          "combat_bosses_killed",
		Name:        "Bosses Killed",
		Description: "Boss enemies defeated",
		Category:    StatCategoryCombat,
		IsCounter:   true,
	},
	"combat_damage_dealt": {
		ID:          "combat_damage_dealt",
		Name:        "Damage Dealt",
		Description: "Total damage dealt to enemies",
		Category:    StatCategoryCombat,
		IsCounter:   true,
	},
	"combat_damage_taken": {
		ID:          "combat_damage_taken",
		Name:        "Damage Taken",
		Description: "Total damage received from enemies",
		Category:    StatCategoryCombat,
		IsCounter:   true,
	},
	"combat_critical_hits": {
		ID:          "combat_critical_hits",
		Name:        "Critical Hits",
		Description: "Critical hits landed",
		Category:    StatCategoryCombat,
		IsCounter:   true,
	},
	"combat_dodges": {
		ID:          "combat_dodges",
		Name:        "Attacks Dodged",
		Description: "Enemy attacks successfully dodged",
		Category:    StatCategoryCombat,
		IsCounter:   true,
	},
	"combat_deaths": {
		ID:          "combat_deaths",
		Name:        "Deaths",
		Description: "Times fallen in combat",
		Category:    StatCategoryCombat,
		IsCounter:   true,
	},
	"combat_healing_done": {
		ID:          "combat_healing_done",
		Name:        "Healing Done",
		Description: "Total health restored",
		Category:    StatCategoryCombat,
		IsCounter:   true,
	},
	"combat_spells_cast": {
		ID:          "combat_spells_cast",
		Name:        "Spells Cast",
		Description: "Combat spells used",
		Category:    StatCategoryCombat,
		IsCounter:   true,
	},
	"combat_highest_hit": {
		ID:          "combat_highest_hit",
		Name:        "Highest Single Hit",
		Description: "Maximum damage dealt in one hit",
		Category:    StatCategoryCombat,
		IsMax:       true,
	},

	// Quest statistics (6)
	"quest_completed": {
		ID:          "quest_completed",
		Name:        "Quests Completed",
		Description: "Total quests finished",
		Category:    StatCategoryQuest,
		IsCounter:   true,
	},
	"quest_main_completed": {
		ID:          "quest_main_completed",
		Name:        "Main Quests",
		Description: "Main story quests completed",
		Category:    StatCategoryQuest,
		IsCounter:   true,
	},
	"quest_side_completed": {
		ID:          "quest_side_completed",
		Name:        "Side Quests",
		Description: "Side quests completed",
		Category:    StatCategoryQuest,
		IsCounter:   true,
	},
	"quest_daily_completed": {
		ID:          "quest_daily_completed",
		Name:        "Daily Quests",
		Description: "Daily quests completed",
		Category:    StatCategoryQuest,
		IsCounter:   true,
	},
	"quest_failed": {
		ID:          "quest_failed",
		Name:        "Quests Failed",
		Description: "Quests that were abandoned or failed",
		Category:    StatCategoryQuest,
		IsCounter:   true,
	},
	"quest_event_completed": {
		ID:          "quest_event_completed",
		Name:        "Event Quests",
		Description: "Seasonal event quests completed",
		Category:    StatCategoryQuest,
		IsCounter:   true,
	},

	// Crafting statistics (6)
	"craft_items_created": {
		ID:          "craft_items_created",
		Name:        "Items Crafted",
		Description: "Total items crafted",
		Category:    StatCategoryCrafting,
		IsCounter:   true,
	},
	"craft_rare_items": {
		ID:          "craft_rare_items",
		Name:        "Rare Items Crafted",
		Description: "Rare or better items crafted",
		Category:    StatCategoryCrafting,
		IsCounter:   true,
	},
	"craft_perfect_items": {
		ID:          "craft_perfect_items",
		Name:        "Perfect Crafts",
		Description: "Items crafted with perfect quality",
		Category:    StatCategoryCrafting,
		IsCounter:   true,
	},
	"craft_recipes_learned": {
		ID:          "craft_recipes_learned",
		Name:        "Recipes Learned",
		Description: "Crafting recipes discovered",
		Category:    StatCategoryCrafting,
		IsCounter:   true,
	},
	"craft_potions_brewed": {
		ID:          "craft_potions_brewed",
		Name:        "Potions Brewed",
		Description: "Potions and elixirs created",
		Category:    StatCategoryCrafting,
		IsCounter:   true,
	},
	"craft_enchantments_applied": {
		ID:          "craft_enchantments_applied",
		Name:        "Enchantments Applied",
		Description: "Items enchanted",
		Category:    StatCategoryCrafting,
		IsCounter:   true,
	},

	// Exploration statistics (7)
	"explore_areas_visited": {
		ID:          "explore_areas_visited",
		Name:        "Areas Visited",
		Description: "Unique locations discovered",
		Category:    StatCategoryExploration,
		IsCounter:   true,
	},
	"explore_dungeons_cleared": {
		ID:          "explore_dungeons_cleared",
		Name:        "Dungeons Cleared",
		Description: "Dungeons fully explored",
		Category:    StatCategoryExploration,
		IsCounter:   true,
	},
	"explore_secrets_found": {
		ID:          "explore_secrets_found",
		Name:        "Secrets Found",
		Description: "Hidden areas and items discovered",
		Category:    StatCategoryExploration,
		IsCounter:   true,
	},
	"explore_chests_opened": {
		ID:          "explore_chests_opened",
		Name:        "Chests Opened",
		Description: "Treasure chests looted",
		Category:    StatCategoryExploration,
		IsCounter:   true,
	},
	"explore_distance_traveled": {
		ID:          "explore_distance_traveled",
		Name:        "Distance Traveled",
		Description: "Total distance walked",
		Category:    StatCategoryExploration,
		IsCounter:   true,
	},
	"explore_biomes_visited": {
		ID:          "explore_biomes_visited",
		Name:        "Biomes Visited",
		Description: "Different biome types explored",
		Category:    StatCategoryExploration,
		IsCounter:   true,
	},
	"explore_map_revealed": {
		ID:          "explore_map_revealed",
		Name:        "Map Revealed",
		Description: "Percentage of map uncovered",
		Category:    StatCategoryExploration,
		IsMax:       true,
	},

	// Social statistics (6)
	"social_friends_added": {
		ID:          "social_friends_added",
		Name:        "Friends Added",
		Description: "Players added as friends",
		Category:    StatCategorySocial,
		IsCounter:   true,
	},
	"social_guilds_joined": {
		ID:          "social_guilds_joined",
		Name:        "Guilds Joined",
		Description: "Guilds joined over time",
		Category:    StatCategorySocial,
		IsCounter:   true,
	},
	"social_trades_completed": {
		ID:          "social_trades_completed",
		Name:        "Trades Completed",
		Description: "Player-to-player trades finished",
		Category:    StatCategorySocial,
		IsCounter:   true,
	},
	"social_chat_messages": {
		ID:          "social_chat_messages",
		Name:        "Chat Messages",
		Description: "Messages sent in chat",
		Category:    StatCategorySocial,
		IsCounter:   true,
	},
	"social_party_activities": {
		ID:          "social_party_activities",
		Name:        "Party Activities",
		Description: "Activities completed with a party",
		Category:    StatCategorySocial,
		IsCounter:   true,
	},
	"social_emotes_used": {
		ID:          "social_emotes_used",
		Name:        "Emotes Used",
		Description: "Expressions and emotes performed",
		Category:    StatCategorySocial,
		IsCounter:   true,
	},

	// PvP statistics (6)
	"pvp_matches_played": {
		ID:          "pvp_matches_played",
		Name:        "PvP Matches Played",
		Description: "Total PvP matches participated in",
		Category:    StatCategoryPvP,
		IsCounter:   true,
	},
	"pvp_matches_won": {
		ID:          "pvp_matches_won",
		Name:        "PvP Matches Won",
		Description: "PvP matches won",
		Category:    StatCategoryPvP,
		IsCounter:   true,
	},
	"pvp_tournaments_entered": {
		ID:          "pvp_tournaments_entered",
		Name:        "Tournaments Entered",
		Description: "Tournament competitions joined",
		Category:    StatCategoryPvP,
		IsCounter:   true,
	},
	"pvp_tournaments_won": {
		ID:          "pvp_tournaments_won",
		Name:        "Tournaments Won",
		Description: "Tournament victories",
		Category:    StatCategoryPvP,
		IsCounter:   true,
	},
	"pvp_honor_earned": {
		ID:          "pvp_honor_earned",
		Name:        "Honor Earned",
		Description: "Total honor points gained",
		Category:    StatCategoryPvP,
		IsCounter:   true,
	},
	"pvp_highest_rating": {
		ID:          "pvp_highest_rating",
		Name:        "Highest Rating",
		Description: "Peak ELO rating achieved",
		Category:    StatCategoryPvP,
		IsMax:       true,
	},

	// Economy statistics (5)
	"economy_gold_earned": {
		ID:          "economy_gold_earned",
		Name:        "Gold Earned",
		Description: "Total gold acquired",
		Category:    StatCategoryEconomy,
		IsCounter:   true,
	},
	"economy_gold_spent": {
		ID:          "economy_gold_spent",
		Name:        "Gold Spent",
		Description: "Total gold spent on purchases",
		Category:    StatCategoryEconomy,
		IsCounter:   true,
	},
	"economy_items_sold": {
		ID:          "economy_items_sold",
		Name:        "Items Sold",
		Description: "Items sold to vendors or players",
		Category:    StatCategoryEconomy,
		IsCounter:   true,
	},
	"economy_items_bought": {
		ID:          "economy_items_bought",
		Name:        "Items Bought",
		Description: "Items purchased from vendors or players",
		Category:    StatCategoryEconomy,
		IsCounter:   true,
	},
	"economy_highest_gold": {
		ID:          "economy_highest_gold",
		Name:        "Highest Gold Held",
		Description: "Maximum gold held at one time",
		Category:    StatCategoryEconomy,
		IsMax:       true,
	},

	// General/Time statistics (6)
	"general_playtime": {
		ID:          "general_playtime",
		Name:        "Total Playtime",
		Description: "Total time spent playing",
		Category:    StatCategoryGeneral,
		IsTime:      true,
		IsCounter:   true,
	},
	"general_combat_time": {
		ID:          "general_combat_time",
		Name:        "Combat Time",
		Description: "Time spent in combat",
		Category:    StatCategoryGeneral,
		IsTime:      true,
		IsCounter:   true,
	},
	"general_crafting_time": {
		ID:          "general_crafting_time",
		Name:        "Crafting Time",
		Description: "Time spent crafting",
		Category:    StatCategoryGeneral,
		IsTime:      true,
		IsCounter:   true,
	},
	"general_level_reached": {
		ID:          "general_level_reached",
		Name:        "Highest Level",
		Description: "Maximum character level reached",
		Category:    StatCategoryGeneral,
		IsMax:       true,
	},
	"general_sessions_played": {
		ID:          "general_sessions_played",
		Name:        "Sessions Played",
		Description: "Number of game sessions",
		Category:    StatCategoryGeneral,
		IsCounter:   true,
	},
	"general_achievements_unlocked": {
		ID:          "general_achievements_unlocked",
		Name:        "Achievements Unlocked",
		Description: "Total achievements earned",
		Category:    StatCategoryGeneral,
		IsCounter:   true,
	},
}

// GetStatDefinition returns a stat definition by ID.
func GetStatDefinition(id string) (StatDefinition, bool) {
	def, exists := statDefinitions[id]
	return def, exists
}

// GetAllStatDefinitions returns all statistic definitions.
func GetAllStatDefinitions() map[string]StatDefinition {
	return statDefinitions
}

// GetStatDefinitionsByCategory returns definitions for a category.
func GetStatDefinitionsByCategory(category StatCategory) []StatDefinition {
	var result []StatDefinition
	for _, def := range statDefinitions {
		if def.Category == category {
			result = append(result, def)
		}
	}
	return result
}

// GetStatCount returns the total number of tracked statistics.
func GetStatCount() int {
	return len(statDefinitions)
}

// PlayerStatisticsComponent tracks all player statistics for lifetime and session.
type PlayerStatisticsComponent struct {
	mu sync.RWMutex
	// Lifetime statistics that persist across sessions.
	Lifetime map[string]int64 `json:"lifetime"`
	// Session statistics that reset each session.
	Session map[string]int64 `json:"session"`
	// FirstPlayTime is the Unix timestamp of the first play session.
	FirstPlayTime int64 `json:"first_play_time"`
	// TotalPlayTime is the total playtime in seconds (lifetime).
	TotalPlayTime int64 `json:"total_play_time"`
	// SessionStartTime is the Unix timestamp when current session started.
	SessionStartTime int64 `json:"session_start_time"`
	// SessionPlayTime is the current session playtime in seconds.
	SessionPlayTime int64 `json:"session_play_time"`
}

// Type returns the component type identifier.
func (c *PlayerStatisticsComponent) Type() string {
	return "player_statistics"
}

// NewPlayerStatisticsComponent creates a new player statistics component.
func NewPlayerStatisticsComponent() *PlayerStatisticsComponent {
	return &PlayerStatisticsComponent{
		Lifetime: make(map[string]int64),
		Session:  make(map[string]int64),
	}
}

// StartSession initializes a new session with the given timestamp.
func (c *PlayerStatisticsComponent) StartSession(timestamp int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clear session stats.
	c.Session = make(map[string]int64)
	c.SessionStartTime = timestamp
	c.SessionPlayTime = 0

	// Set first play time if not set.
	if c.FirstPlayTime == 0 {
		c.FirstPlayTime = timestamp
	}

	// Increment session count.
	c.Lifetime["general_sessions_played"]++
}

// EndSession finalizes the current session.
func (c *PlayerStatisticsComponent) EndSession(timestamp int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.SessionStartTime > 0 {
		sessionDuration := timestamp - c.SessionStartTime
		c.SessionPlayTime = sessionDuration
		c.TotalPlayTime += sessionDuration
		c.Lifetime["general_playtime"] += sessionDuration
	}
}

// UpdatePlaytime updates the playtime stats with elapsed seconds.
func (c *PlayerStatisticsComponent) UpdatePlaytime(elapsedSeconds int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.SessionPlayTime += elapsedSeconds
	c.TotalPlayTime += elapsedSeconds
	c.Lifetime["general_playtime"] += elapsedSeconds
	c.Session["general_playtime"] += elapsedSeconds
}

// IncrementStat increments a counter stat by the given amount.
func (c *PlayerStatisticsComponent) IncrementStat(statID string, amount int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Lifetime[statID] += amount
	c.Session[statID] += amount
}

// SetMaxStat sets a max stat if the new value is higher.
func (c *PlayerStatisticsComponent) SetMaxStat(statID string, value int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if value > c.Lifetime[statID] {
		c.Lifetime[statID] = value
	}
	if value > c.Session[statID] {
		c.Session[statID] = value
	}
}

// GetLifetimeStat returns a lifetime stat value.
func (c *PlayerStatisticsComponent) GetLifetimeStat(statID string) int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Lifetime[statID]
}

// GetSessionStat returns a session stat value.
func (c *PlayerStatisticsComponent) GetSessionStat(statID string) int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Session[statID]
}

// GetAllLifetimeStats returns a copy of all lifetime stats.
func (c *PlayerStatisticsComponent) GetAllLifetimeStats() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]int64, len(c.Lifetime))
	for k, v := range c.Lifetime {
		result[k] = v
	}
	return result
}

// GetAllSessionStats returns a copy of all session stats.
func (c *PlayerStatisticsComponent) GetAllSessionStats() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]int64, len(c.Session))
	for k, v := range c.Session {
		result[k] = v
	}
	return result
}

// GetStatsByCategory returns lifetime stats for a specific category.
func (c *PlayerStatisticsComponent) GetStatsByCategory(category StatCategory) map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]int64)
	for id, def := range statDefinitions {
		if def.Category == category {
			result[id] = c.Lifetime[id]
		}
	}
	return result
}

// GetTotalPlayTime returns the total playtime in seconds.
func (c *PlayerStatisticsComponent) GetTotalPlayTime() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.TotalPlayTime
}

// GetSessionPlayTime returns the current session playtime in seconds.
func (c *PlayerStatisticsComponent) GetSessionPlayTime() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.SessionPlayTime
}

// GetFirstPlayTime returns the first play Unix timestamp.
func (c *PlayerStatisticsComponent) GetFirstPlayTime() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.FirstPlayTime
}

// Serialize converts the component to JSON bytes for persistence.
func (c *PlayerStatisticsComponent) Serialize() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return json.Marshal(struct {
		Lifetime         map[string]int64 `json:"lifetime"`
		Session          map[string]int64 `json:"session"`
		FirstPlayTime    int64            `json:"first_play_time"`
		TotalPlayTime    int64            `json:"total_play_time"`
		SessionStartTime int64            `json:"session_start_time"`
		SessionPlayTime  int64            `json:"session_play_time"`
	}{
		Lifetime:         c.Lifetime,
		Session:          c.Session,
		FirstPlayTime:    c.FirstPlayTime,
		TotalPlayTime:    c.TotalPlayTime,
		SessionStartTime: c.SessionStartTime,
		SessionPlayTime:  c.SessionPlayTime,
	})
}

// Deserialize restores the component from JSON bytes.
func (c *PlayerStatisticsComponent) Deserialize(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var temp struct {
		Lifetime         map[string]int64 `json:"lifetime"`
		Session          map[string]int64 `json:"session"`
		FirstPlayTime    int64            `json:"first_play_time"`
		TotalPlayTime    int64            `json:"total_play_time"`
		SessionStartTime int64            `json:"session_start_time"`
		SessionPlayTime  int64            `json:"session_play_time"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	c.Lifetime = temp.Lifetime
	c.Session = temp.Session
	c.FirstPlayTime = temp.FirstPlayTime
	c.TotalPlayTime = temp.TotalPlayTime
	c.SessionStartTime = temp.SessionStartTime
	c.SessionPlayTime = temp.SessionPlayTime

	// Ensure maps are initialized.
	if c.Lifetime == nil {
		c.Lifetime = make(map[string]int64)
	}
	if c.Session == nil {
		c.Session = make(map[string]int64)
	}

	return nil
}
