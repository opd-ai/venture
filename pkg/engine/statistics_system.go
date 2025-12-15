// Package engine provides the statistics system for tracking player metrics.
// This file implements the StatisticsSystem which updates player statistics
// based on game events across all systems.
//
// Phase 84: Player Statistics System (V15.0)
package engine

import (
	"sync"

	"github.com/sirupsen/logrus"
)

// StatisticsSystem tracks and updates player statistics based on game events.
type StatisticsSystem struct {
	world  *World
	logger *logrus.Entry
	mu     sync.RWMutex

	// lastUpdateTime tracks when stats were last updated for playtime.
	lastUpdateTime float64
}

// NewStatisticsSystem creates a new statistics system.
func NewStatisticsSystem(world *World) *StatisticsSystem {
	var logger *logrus.Entry
	if world != nil && world.logger != nil {
		logger = world.logger.Logger.WithField("system", "statistics")
	}

	return &StatisticsSystem{
		world:  world,
		logger: logger,
	}
}

// Update processes statistics updates, primarily for time-based stats.
func (s *StatisticsSystem) Update(entities []*Entity, deltaTime float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Update playtime for all entities with statistics components.
	for _, entity := range entities {
		if entity == nil {
			continue
		}

		compRaw, exists := entity.GetComponent("player_statistics")
		if !exists {
			continue
		}

		comp, ok := compRaw.(*PlayerStatisticsComponent)
		if !ok {
			continue
		}

		// Convert deltaTime (in seconds) to int64 (accumulate fractional seconds).
		// We track time in whole seconds for simplicity.
		if deltaTime >= 1.0 {
			comp.UpdatePlaytime(int64(deltaTime))
		}
	}
}

// getOrCreateComponent gets or creates a PlayerStatisticsComponent for an entity.
func (s *StatisticsSystem) getOrCreateComponent(entity *Entity) *PlayerStatisticsComponent {
	if entity == nil {
		return nil
	}

	compRaw, exists := entity.GetComponent("player_statistics")
	if exists {
		if comp, ok := compRaw.(*PlayerStatisticsComponent); ok {
			return comp
		}
	}

	comp := NewPlayerStatisticsComponent()
	entity.AddComponent(comp)
	return comp
}

// StartSession initializes a new session for an entity.
func (s *StatisticsSystem) StartSession(entityID uint64, timestamp int64) {
	if s.world == nil {
		return
	}

	entity, exists := s.world.GetEntity(entityID)
	if !exists || entity == nil {
		return
	}

	comp := s.getOrCreateComponent(entity)
	if comp != nil {
		comp.StartSession(timestamp)

		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entityID,
				"timestamp": timestamp,
			}).Debug("Session started")
		}
	}
}

// EndSession finalizes a session for an entity.
func (s *StatisticsSystem) EndSession(entityID uint64, timestamp int64) {
	if s.world == nil {
		return
	}

	entity, exists := s.world.GetEntity(entityID)
	if !exists || entity == nil {
		return
	}

	compRaw, exists := entity.GetComponent("player_statistics")
	if !exists {
		return
	}

	comp, ok := compRaw.(*PlayerStatisticsComponent)
	if !ok {
		return
	}

	comp.EndSession(timestamp)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":     entityID,
			"session_time":  comp.GetSessionPlayTime(),
			"lifetime_time": comp.GetTotalPlayTime(),
		}).Debug("Session ended")
	}
}

// RecordStat records a stat update for an entity.
// For counter stats, use amount to increment.
// For max stats, the value is compared against current max.
func (s *StatisticsSystem) RecordStat(entityID uint64, statID string, amount int64) {
	if s.world == nil {
		return
	}

	entity, exists := s.world.GetEntity(entityID)
	if !exists || entity == nil {
		return
	}

	def, defExists := GetStatDefinition(statID)
	if !defExists {
		if s.logger != nil {
			s.logger.WithField("stat_id", statID).Warn("Unknown stat ID")
		}
		return
	}

	comp := s.getOrCreateComponent(entity)
	if comp == nil {
		return
	}

	if def.IsMax {
		comp.SetMaxStat(statID, amount)
	} else {
		comp.IncrementStat(statID, amount)
	}
}

// IncrementStat increments a counter stat by the given amount.
func (s *StatisticsSystem) IncrementStat(entityID uint64, statID string, amount int64) {
	s.RecordStat(entityID, statID, amount)
}

// SetMaxStat sets a max stat if the new value is higher.
func (s *StatisticsSystem) SetMaxStat(entityID uint64, statID string, value int64) {
	if s.world == nil {
		return
	}

	entity, exists := s.world.GetEntity(entityID)
	if !exists || entity == nil {
		return
	}

	comp := s.getOrCreateComponent(entity)
	if comp != nil {
		comp.SetMaxStat(statID, value)
	}
}

// GetLifetimeStat returns a lifetime stat value for an entity.
func (s *StatisticsSystem) GetLifetimeStat(entityID uint64, statID string) int64 {
	if s.world == nil {
		return 0
	}

	entity, exists := s.world.GetEntity(entityID)
	if !exists || entity == nil {
		return 0
	}

	compRaw, exists := entity.GetComponent("player_statistics")
	if !exists {
		return 0
	}

	comp, ok := compRaw.(*PlayerStatisticsComponent)
	if !ok {
		return 0
	}

	return comp.GetLifetimeStat(statID)
}

// GetSessionStat returns a session stat value for an entity.
func (s *StatisticsSystem) GetSessionStat(entityID uint64, statID string) int64 {
	if s.world == nil {
		return 0
	}

	entity, exists := s.world.GetEntity(entityID)
	if !exists || entity == nil {
		return 0
	}

	compRaw, exists := entity.GetComponent("player_statistics")
	if !exists {
		return 0
	}

	comp, ok := compRaw.(*PlayerStatisticsComponent)
	if !ok {
		return 0
	}

	return comp.GetSessionStat(statID)
}

// GetTotalPlayTime returns total playtime for an entity in seconds.
func (s *StatisticsSystem) GetTotalPlayTime(entityID uint64) int64 {
	if s.world == nil {
		return 0
	}

	entity, exists := s.world.GetEntity(entityID)
	if !exists || entity == nil {
		return 0
	}

	compRaw, exists := entity.GetComponent("player_statistics")
	if !exists {
		return 0
	}

	comp, ok := compRaw.(*PlayerStatisticsComponent)
	if !ok {
		return 0
	}

	return comp.GetTotalPlayTime()
}

// GetSessionPlayTime returns current session playtime for an entity.
func (s *StatisticsSystem) GetSessionPlayTime(entityID uint64) int64 {
	if s.world == nil {
		return 0
	}

	entity, exists := s.world.GetEntity(entityID)
	if !exists || entity == nil {
		return 0
	}

	compRaw, exists := entity.GetComponent("player_statistics")
	if !exists {
		return 0
	}

	comp, ok := compRaw.(*PlayerStatisticsComponent)
	if !ok {
		return 0
	}

	return comp.GetSessionPlayTime()
}

// Event handler methods for game events.

// OnEnemyKilled is called when an enemy is killed.
func (s *StatisticsSystem) OnEnemyKilled(entityID uint64, isBoss bool, damageDealt int64) {
	s.IncrementStat(entityID, "combat_enemies_killed", 1)
	if isBoss {
		s.IncrementStat(entityID, "combat_bosses_killed", 1)
	}
}

// OnDamageDealt is called when damage is dealt.
func (s *StatisticsSystem) OnDamageDealt(entityID uint64, damage int64, isCritical bool) {
	s.IncrementStat(entityID, "combat_damage_dealt", damage)
	if isCritical {
		s.IncrementStat(entityID, "combat_critical_hits", 1)
	}
	s.SetMaxStat(entityID, "combat_highest_hit", damage)
}

// OnDamageTaken is called when damage is received.
func (s *StatisticsSystem) OnDamageTaken(entityID uint64, damage int64) {
	s.IncrementStat(entityID, "combat_damage_taken", damage)
}

// OnDodge is called when an attack is dodged.
func (s *StatisticsSystem) OnDodge(entityID uint64) {
	s.IncrementStat(entityID, "combat_dodges", 1)
}

// OnDeath is called when the player dies.
func (s *StatisticsSystem) OnDeath(entityID uint64) {
	s.IncrementStat(entityID, "combat_deaths", 1)
}

// OnHealing is called when healing is done.
func (s *StatisticsSystem) OnHealing(entityID uint64, amount int64) {
	s.IncrementStat(entityID, "combat_healing_done", amount)
}

// OnSpellCast is called when a spell is cast.
func (s *StatisticsSystem) OnSpellCast(entityID uint64) {
	s.IncrementStat(entityID, "combat_spells_cast", 1)
}

// OnQuestCompleted is called when a quest is completed.
func (s *StatisticsSystem) OnQuestCompleted(entityID uint64, isMain, isSide, isDaily, isEvent bool) {
	s.IncrementStat(entityID, "quest_completed", 1)
	if isMain {
		s.IncrementStat(entityID, "quest_main_completed", 1)
	}
	if isSide {
		s.IncrementStat(entityID, "quest_side_completed", 1)
	}
	if isDaily {
		s.IncrementStat(entityID, "quest_daily_completed", 1)
	}
	if isEvent {
		s.IncrementStat(entityID, "quest_event_completed", 1)
	}
}

// OnQuestFailed is called when a quest fails.
func (s *StatisticsSystem) OnQuestFailed(entityID uint64) {
	s.IncrementStat(entityID, "quest_failed", 1)
}

// OnItemCrafted is called when an item is crafted.
func (s *StatisticsSystem) OnItemCrafted(entityID uint64, isRare, isPerfect, isPotion, isEnchantment bool) {
	s.IncrementStat(entityID, "craft_items_created", 1)
	if isRare {
		s.IncrementStat(entityID, "craft_rare_items", 1)
	}
	if isPerfect {
		s.IncrementStat(entityID, "craft_perfect_items", 1)
	}
	if isPotion {
		s.IncrementStat(entityID, "craft_potions_brewed", 1)
	}
	if isEnchantment {
		s.IncrementStat(entityID, "craft_enchantments_applied", 1)
	}
}

// OnRecipeLearned is called when a recipe is learned.
func (s *StatisticsSystem) OnRecipeLearned(entityID uint64) {
	s.IncrementStat(entityID, "craft_recipes_learned", 1)
}

// OnAreaVisited is called when a new area is visited.
func (s *StatisticsSystem) OnAreaVisited(entityID uint64) {
	s.IncrementStat(entityID, "explore_areas_visited", 1)
}

// OnDungeonCleared is called when a dungeon is cleared.
func (s *StatisticsSystem) OnDungeonCleared(entityID uint64) {
	s.IncrementStat(entityID, "explore_dungeons_cleared", 1)
}

// OnSecretFound is called when a secret is discovered.
func (s *StatisticsSystem) OnSecretFound(entityID uint64) {
	s.IncrementStat(entityID, "explore_secrets_found", 1)
}

// OnChestOpened is called when a chest is opened.
func (s *StatisticsSystem) OnChestOpened(entityID uint64) {
	s.IncrementStat(entityID, "explore_chests_opened", 1)
}

// OnDistanceTraveled is called when distance is traveled.
func (s *StatisticsSystem) OnDistanceTraveled(entityID uint64, distance int64) {
	s.IncrementStat(entityID, "explore_distance_traveled", distance)
}

// OnBiomeVisited is called when a new biome is visited.
func (s *StatisticsSystem) OnBiomeVisited(entityID uint64) {
	s.IncrementStat(entityID, "explore_biomes_visited", 1)
}

// OnMapRevealed is called to update map revealed percentage.
func (s *StatisticsSystem) OnMapRevealed(entityID uint64, percentage int64) {
	s.SetMaxStat(entityID, "explore_map_revealed", percentage)
}

// OnFriendAdded is called when a friend is added.
func (s *StatisticsSystem) OnFriendAdded(entityID uint64) {
	s.IncrementStat(entityID, "social_friends_added", 1)
}

// OnGuildJoined is called when a guild is joined.
func (s *StatisticsSystem) OnGuildJoined(entityID uint64) {
	s.IncrementStat(entityID, "social_guilds_joined", 1)
}

// OnTradeCompleted is called when a trade is completed.
func (s *StatisticsSystem) OnTradeCompleted(entityID uint64) {
	s.IncrementStat(entityID, "social_trades_completed", 1)
}

// OnChatMessage is called when a chat message is sent.
func (s *StatisticsSystem) OnChatMessage(entityID uint64) {
	s.IncrementStat(entityID, "social_chat_messages", 1)
}

// OnPartyActivity is called when a party activity is completed.
func (s *StatisticsSystem) OnPartyActivity(entityID uint64) {
	s.IncrementStat(entityID, "social_party_activities", 1)
}

// OnEmoteUsed is called when an emote is used.
func (s *StatisticsSystem) OnEmoteUsed(entityID uint64) {
	s.IncrementStat(entityID, "social_emotes_used", 1)
}

// OnPvPMatchPlayed is called when a PvP match is played.
func (s *StatisticsSystem) OnPvPMatchPlayed(entityID uint64, won bool) {
	s.IncrementStat(entityID, "pvp_matches_played", 1)
	if won {
		s.IncrementStat(entityID, "pvp_matches_won", 1)
	}
}

// OnTournamentEntered is called when a tournament is entered.
func (s *StatisticsSystem) OnTournamentEntered(entityID uint64) {
	s.IncrementStat(entityID, "pvp_tournaments_entered", 1)
}

// OnTournamentWon is called when a tournament is won.
func (s *StatisticsSystem) OnTournamentWon(entityID uint64) {
	s.IncrementStat(entityID, "pvp_tournaments_won", 1)
}

// OnHonorEarned is called when honor is earned.
func (s *StatisticsSystem) OnHonorEarned(entityID uint64, amount int64) {
	s.IncrementStat(entityID, "pvp_honor_earned", amount)
}

// OnRatingChanged is called when PvP rating changes.
func (s *StatisticsSystem) OnRatingChanged(entityID uint64, newRating int64) {
	s.SetMaxStat(entityID, "pvp_highest_rating", newRating)
}

// OnGoldEarned is called when gold is earned.
func (s *StatisticsSystem) OnGoldEarned(entityID uint64, amount int64) {
	s.IncrementStat(entityID, "economy_gold_earned", amount)
}

// OnGoldSpent is called when gold is spent.
func (s *StatisticsSystem) OnGoldSpent(entityID uint64, amount int64) {
	s.IncrementStat(entityID, "economy_gold_spent", amount)
}

// OnItemSold is called when an item is sold.
func (s *StatisticsSystem) OnItemSold(entityID uint64) {
	s.IncrementStat(entityID, "economy_items_sold", 1)
}

// OnItemBought is called when an item is bought.
func (s *StatisticsSystem) OnItemBought(entityID uint64) {
	s.IncrementStat(entityID, "economy_items_bought", 1)
}

// OnGoldHeld is called to update highest gold held.
func (s *StatisticsSystem) OnGoldHeld(entityID uint64, currentGold int64) {
	s.SetMaxStat(entityID, "economy_highest_gold", currentGold)
}

// OnCombatTime is called to add combat time.
func (s *StatisticsSystem) OnCombatTime(entityID uint64, seconds int64) {
	s.IncrementStat(entityID, "general_combat_time", seconds)
}

// OnCraftingTime is called to add crafting time.
func (s *StatisticsSystem) OnCraftingTime(entityID uint64, seconds int64) {
	s.IncrementStat(entityID, "general_crafting_time", seconds)
}

// OnLevelUp is called when the player levels up.
func (s *StatisticsSystem) OnLevelUp(entityID uint64, newLevel int64) {
	s.SetMaxStat(entityID, "general_level_reached", newLevel)
}

// OnAchievementUnlocked is called when an achievement is unlocked.
func (s *StatisticsSystem) OnAchievementUnlocked(entityID uint64) {
	s.IncrementStat(entityID, "general_achievements_unlocked", 1)
}
