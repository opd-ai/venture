// Package engine provides the extended achievement system for tracking
// player accomplishments across all game systems.
//
// # Relationship to AchievementSystem
//
// The engine contains two complementary achievement trackers:
//
//   - AchievementSystem (achievement.go) — tracks expression/social combos.
//   - ExtendedAchievementSystem (this file) — tracks gameplay events (kills,
//     quests, crafting, exploration, PvP) with multi-tier progression.
//
// Both systems use distinct component types (AchievementComponent vs
// ExtendedAchievementComponent) so registering both in the same World is safe
// and does not cause double-firing of rewards.  InitializeGameSystems registers
// both to provide full achievement coverage.
//
// Phase 83: Extended Achievement Categories (V15.0)
package engine

import (
	"sync"

	"github.com/sirupsen/logrus"
)

// achievementDefinitions contains all achievement definitions.
// Achievement IDs follow the pattern: category_name (e.g., combat_first_blood)
var achievementDefinitions = map[string]AchievementDefinition{
	// Combat achievements (10)
	"combat_first_blood": {
		ID:          "combat_first_blood",
		Name:        "First Blood",
		Description: "Defeat enemies in combat",
		Category:    AchievementCategoryCombat,
		Thresholds:  [4]int64{1, 10, 100, 1000},
	},
	"combat_boss_slayer": {
		ID:          "combat_boss_slayer",
		Name:        "Boss Slayer",
		Description: "Defeat powerful bosses",
		Category:    AchievementCategoryCombat,
		Thresholds:  [4]int64{1, 5, 25, 100},
	},
	"combat_damage_dealer": {
		ID:          "combat_damage_dealer",
		Name:        "Damage Dealer",
		Description: "Deal damage to enemies",
		Category:    AchievementCategoryCombat,
		Thresholds:  [4]int64{1000, 100000, 1000000, 10000000},
	},
	"combat_survivor": {
		ID:          "combat_survivor",
		Name:        "Survivor",
		Description: "Survive battles at low health",
		Category:    AchievementCategoryCombat,
		Thresholds:  [4]int64{1, 10, 100, 1000},
	},
	"combat_critical_master": {
		ID:          "combat_critical_master",
		Name:        "Critical Master",
		Description: "Land critical hits",
		Category:    AchievementCategoryCombat,
		Thresholds:  [4]int64{10, 100, 1000, 10000},
	},
	"combat_combo_king": {
		ID:          "combat_combo_king",
		Name:        "Combo King",
		Description: "Execute combat combos",
		Category:    AchievementCategoryCombat,
		Thresholds:  [4]int64{5, 50, 500, 5000},
	},
	"combat_untouchable": {
		ID:          "combat_untouchable",
		Name:        "Untouchable",
		Description: "Dodge enemy attacks",
		Category:    AchievementCategoryCombat,
		Thresholds:  [4]int64{10, 100, 1000, 10000},
	},
	"combat_berserker": {
		ID:          "combat_berserker",
		Name:        "Berserker",
		Description: "Defeat multiple enemies quickly",
		Category:    AchievementCategoryCombat,
		Thresholds:  [4]int64{3, 30, 300, 3000},
	},
	"combat_spell_slinger": {
		ID:          "combat_spell_slinger",
		Name:        "Spell Slinger",
		Description: "Cast combat spells",
		Category:    AchievementCategoryCombat,
		Thresholds:  [4]int64{10, 100, 1000, 10000},
	},
	"combat_healer": {
		ID:          "combat_healer",
		Name:        "Combat Medic",
		Description: "Heal during combat",
		Category:    AchievementCategoryCombat,
		Thresholds:  [4]int64{100, 10000, 100000, 1000000},
	},

	// Quest achievements (10)
	"quest_adventurer": {
		ID:          "quest_adventurer",
		Name:        "Adventurer",
		Description: "Complete quests",
		Category:    AchievementCategoryQuest,
		Thresholds:  [4]int64{1, 10, 50, 200},
	},
	"quest_main_story": {
		ID:          "quest_main_story",
		Name:        "Story Seeker",
		Description: "Complete main story chapters",
		Category:    AchievementCategoryQuest,
		Thresholds:  [4]int64{1, 3, 7, 15},
	},
	"quest_side_tracker": {
		ID:          "quest_side_tracker",
		Name:        "Side Tracker",
		Description: "Complete side quests",
		Category:    AchievementCategoryQuest,
		Thresholds:  [4]int64{5, 25, 100, 500},
	},
	"quest_speed_runner": {
		ID:          "quest_speed_runner",
		Name:        "Speed Runner",
		Description: "Complete quests under time bonus",
		Category:    AchievementCategoryQuest,
		Thresholds:  [4]int64{1, 10, 50, 200},
	},
	"quest_completionist": {
		ID:          "quest_completionist",
		Name:        "Completionist",
		Description: "Achieve 100% quest completion in zones",
		Category:    AchievementCategoryQuest,
		Thresholds:  [4]int64{1, 5, 15, 50},
	},
	"quest_helper": {
		ID:          "quest_helper",
		Name:        "Helping Hand",
		Description: "Help others complete quests",
		Category:    AchievementCategoryQuest,
		Thresholds:  [4]int64{1, 10, 50, 200},
	},
	"quest_collector": {
		ID:          "quest_collector",
		Name:        "Collector",
		Description: "Complete collection quests",
		Category:    AchievementCategoryQuest,
		Thresholds:  [4]int64{5, 25, 100, 500},
	},
	"quest_escort": {
		ID:          "quest_escort",
		Name:        "Escort Expert",
		Description: "Complete escort quests without failure",
		Category:    AchievementCategoryQuest,
		Thresholds:  [4]int64{1, 10, 50, 200},
	},
	"quest_legendary": {
		ID:          "quest_legendary",
		Name:        "Legend Seeker",
		Description: "Complete legendary quests",
		Category:    AchievementCategoryQuest,
		Thresholds:  [4]int64{1, 5, 15, 50},
	},
	"quest_daily": {
		ID:          "quest_daily",
		Name:        "Daily Dedication",
		Description: "Complete daily quests",
		Category:    AchievementCategoryQuest,
		Thresholds:  [4]int64{7, 30, 100, 365},
	},

	// Crafting achievements (10)
	"craft_apprentice": {
		ID:          "craft_apprentice",
		Name:        "Apprentice Crafter",
		Description: "Craft items",
		Category:    AchievementCategoryCrafting,
		Thresholds:  [4]int64{1, 10, 100, 1000},
	},
	"craft_master": {
		ID:          "craft_master",
		Name:        "Master Craftsman",
		Description: "Craft items of each type",
		Category:    AchievementCategoryCrafting,
		Thresholds:  [4]int64{3, 6, 10, 15},
	},
	"craft_quality": {
		ID:          "craft_quality",
		Name:        "Quality Crafter",
		Description: "Craft rare or better items",
		Category:    AchievementCategoryCrafting,
		Thresholds:  [4]int64{1, 10, 50, 200},
	},
	"craft_recipe_collector": {
		ID:          "craft_recipe_collector",
		Name:        "Recipe Collector",
		Description: "Learn crafting recipes",
		Category:    AchievementCategoryCrafting,
		Thresholds:  [4]int64{10, 50, 200, 500},
	},
	"craft_specialized": {
		ID:          "craft_specialized",
		Name:        "Specialized",
		Description: "Max level in a crafting skill",
		Category:    AchievementCategoryCrafting,
		Thresholds:  [4]int64{1, 2, 4, 8},
	},
	"craft_perfect": {
		ID:          "craft_perfect",
		Name:        "Perfect Craft",
		Description: "Craft with perfect quality bonus",
		Category:    AchievementCategoryCrafting,
		Thresholds:  [4]int64{1, 10, 100, 1000},
	},
	"craft_experimenter": {
		ID:          "craft_experimenter",
		Name:        "Experimenter",
		Description: "Discover recipes through experimentation",
		Category:    AchievementCategoryCrafting,
		Thresholds:  [4]int64{1, 10, 50, 200},
	},
	"craft_enchanter": {
		ID:          "craft_enchanter",
		Name:        "Enchanter",
		Description: "Enchant items",
		Category:    AchievementCategoryCrafting,
		Thresholds:  [4]int64{1, 10, 100, 1000},
	},
	"craft_alchemist": {
		ID:          "craft_alchemist",
		Name:        "Alchemist",
		Description: "Brew potions",
		Category:    AchievementCategoryCrafting,
		Thresholds:  [4]int64{10, 100, 500, 2000},
	},
	"craft_blacksmith": {
		ID:          "craft_blacksmith",
		Name:        "Blacksmith",
		Description: "Forge weapons and armor",
		Category:    AchievementCategoryCrafting,
		Thresholds:  [4]int64{10, 100, 500, 2000},
	},

	// Exploration achievements (10)
	"explore_wanderer": {
		ID:          "explore_wanderer",
		Name:        "Wanderer",
		Description: "Visit unique areas",
		Category:    AchievementCategoryExploration,
		Thresholds:  [4]int64{1, 10, 50, 200},
	},
	"explore_dungeon_delver": {
		ID:          "explore_dungeon_delver",
		Name:        "Dungeon Delver",
		Description: "Clear dungeons",
		Category:    AchievementCategoryExploration,
		Thresholds:  [4]int64{1, 10, 50, 200},
	},
	"explore_secret_finder": {
		ID:          "explore_secret_finder",
		Name:        "Secret Finder",
		Description: "Discover hidden secrets",
		Category:    AchievementCategoryExploration,
		Thresholds:  [4]int64{1, 10, 50, 200},
	},
	"explore_world_traveler": {
		ID:          "explore_world_traveler",
		Name:        "World Traveler",
		Description: "Visit all biome types",
		Category:    AchievementCategoryExploration,
		Thresholds:  [4]int64{3, 6, 10, 15},
	},
	"explore_cartographer": {
		ID:          "explore_cartographer",
		Name:        "Cartographer",
		Description: "Reveal map percentage",
		Category:    AchievementCategoryExploration,
		Thresholds:  [4]int64{25, 50, 75, 100},
	},
	"explore_treasure_hunter": {
		ID:          "explore_treasure_hunter",
		Name:        "Treasure Hunter",
		Description: "Find treasure chests",
		Category:    AchievementCategoryExploration,
		Thresholds:  [4]int64{5, 50, 200, 1000},
	},
	"explore_mountain_climber": {
		ID:          "explore_mountain_climber",
		Name:        "Mountain Climber",
		Description: "Reach high altitude locations",
		Category:    AchievementCategoryExploration,
		Thresholds:  [4]int64{1, 10, 50, 200},
	},
	"explore_deep_diver": {
		ID:          "explore_deep_diver",
		Name:        "Deep Diver",
		Description: "Explore underwater areas",
		Category:    AchievementCategoryExploration,
		Thresholds:  [4]int64{1, 10, 50, 200},
	},
	"explore_night_owl": {
		ID:          "explore_night_owl",
		Name:        "Night Owl",
		Description: "Explore during night time",
		Category:    AchievementCategoryExploration,
		Thresholds:  [4]int64{10, 100, 500, 2000},
	},
	"explore_distance": {
		ID:          "explore_distance",
		Name:        "Long Distance",
		Description: "Travel total distance (units)",
		Category:    AchievementCategoryExploration,
		Thresholds:  [4]int64{10000, 100000, 1000000, 10000000},
	},

	// Social achievements (10)
	"social_friend_maker": {
		ID:          "social_friend_maker",
		Name:        "Friend Maker",
		Description: "Add friends",
		Category:    AchievementCategorySocial,
		Thresholds:  [4]int64{1, 5, 25, 100},
	},
	"social_guild_member": {
		ID:          "social_guild_member",
		Name:        "Guild Member",
		Description: "Join guilds",
		Category:    AchievementCategorySocial,
		Thresholds:  [4]int64{1, 2, 5, 10},
	},
	"social_guild_leader": {
		ID:          "social_guild_leader",
		Name:        "Guild Leader",
		Description: "Lead a guild",
		Category:    AchievementCategorySocial,
		Thresholds:  [4]int64{1, 2, 5, 10},
	},
	"social_trade_master": {
		ID:          "social_trade_master",
		Name:        "Trade Master",
		Description: "Complete trades with players",
		Category:    AchievementCategorySocial,
		Thresholds:  [4]int64{10, 100, 1000, 10000},
	},
	"social_chat_active": {
		ID:          "social_chat_active",
		Name:        "Chatterbox",
		Description: "Send chat messages",
		Category:    AchievementCategorySocial,
		Thresholds:  [4]int64{100, 1000, 10000, 100000},
	},
	"social_party_player": {
		ID:          "social_party_player",
		Name:        "Party Player",
		Description: "Complete activities in a party",
		Category:    AchievementCategorySocial,
		Thresholds:  [4]int64{10, 100, 500, 2000},
	},
	"social_helper": {
		ID:          "social_helper",
		Name:        "Community Helper",
		Description: "Help other players",
		Category:    AchievementCategorySocial,
		Thresholds:  [4]int64{5, 25, 100, 500},
	},
	"social_mentor": {
		ID:          "social_mentor",
		Name:        "Mentor",
		Description: "Guide new players",
		Category:    AchievementCategorySocial,
		Thresholds:  [4]int64{1, 10, 50, 200},
	},
	"social_reputation": {
		ID:          "social_reputation",
		Name:        "Renowned",
		Description: "Build reputation in cities",
		Category:    AchievementCategorySocial,
		Thresholds:  [4]int64{1, 5, 15, 50},
	},
	"social_emote": {
		ID:          "social_emote",
		Name:        "Expressive",
		Description: "Use emotes and expressions",
		Category:    AchievementCategorySocial,
		Thresholds:  [4]int64{10, 100, 1000, 10000},
	},

	// PvP achievements (10)
	"pvp_gladiator": {
		ID:          "pvp_gladiator",
		Name:        "Gladiator",
		Description: "Win PvP matches",
		Category:    AchievementCategoryPvP,
		Thresholds:  [4]int64{1, 10, 100, 1000},
	},
	"pvp_tournament_victor": {
		ID:          "pvp_tournament_victor",
		Name:        "Tournament Victor",
		Description: "Win tournaments",
		Category:    AchievementCategoryPvP,
		Thresholds:  [4]int64{1, 5, 25, 100},
	},
	"pvp_rank_silver": {
		ID:          "pvp_rank_silver",
		Name:        "Rising Star",
		Description: "Reach Silver rank or higher",
		Category:    AchievementCategoryPvP,
		Thresholds:  [4]int64{1, 1, 1, 1},
	},
	"pvp_rank_gold": {
		ID:          "pvp_rank_gold",
		Name:        "Competitor",
		Description: "Reach Gold rank or higher",
		Category:    AchievementCategoryPvP,
		Thresholds:  [4]int64{1, 1, 1, 1},
	},
	"pvp_rank_diamond": {
		ID:          "pvp_rank_diamond",
		Name:        "Elite",
		Description: "Reach Diamond rank or higher",
		Category:    AchievementCategoryPvP,
		Thresholds:  [4]int64{1, 1, 1, 1},
	},
	"pvp_rank_legend": {
		ID:          "pvp_rank_legend",
		Name:        "Legend",
		Description: "Reach Legend rank",
		Category:    AchievementCategoryPvP,
		Thresholds:  [4]int64{1, 1, 1, 1},
	},
	"pvp_winning_streak": {
		ID:          "pvp_winning_streak",
		Name:        "Winning Streak",
		Description: "Win consecutive matches",
		Category:    AchievementCategoryPvP,
		Thresholds:  [4]int64{5, 10, 25, 50},
	},
	"pvp_honor_bound": {
		ID:          "pvp_honor_bound",
		Name:        "Honor Bound",
		Description: "Earn honor points",
		Category:    AchievementCategoryPvP,
		Thresholds:  [4]int64{1000, 10000, 100000, 1000000},
	},
	"pvp_duelist": {
		ID:          "pvp_duelist",
		Name:        "Duelist",
		Description: "Win 1v1 matches",
		Category:    AchievementCategoryPvP,
		Thresholds:  [4]int64{10, 100, 500, 2000},
	},
	"pvp_team_player": {
		ID:          "pvp_team_player",
		Name:        "Team Player",
		Description: "Win team matches",
		Category:    AchievementCategoryPvP,
		Thresholds:  [4]int64{10, 100, 500, 2000},
	},
}

// GetAchievementDefinition returns an achievement definition by ID.
func GetAchievementDefinition(id string) (AchievementDefinition, bool) {
	def, exists := achievementDefinitions[id]
	return def, exists
}

// GetAllAchievementDefinitions returns all achievement definitions.
func GetAllAchievementDefinitions() map[string]AchievementDefinition {
	return achievementDefinitions
}

// GetAchievementDefinitionsByCategory returns definitions for a category.
func GetAchievementDefinitionsByCategory(category AchievementCategory) []AchievementDefinition {
	var result []AchievementDefinition
	for _, def := range achievementDefinitions {
		if def.Category == category {
			result = append(result, def)
		}
	}
	return result
}

// ExtendedAchievementSystem tracks and unlocks achievements based on game events.
type ExtendedAchievementSystem struct {
	world  *World
	logger *logrus.Entry
	mu     sync.RWMutex

	// Callbacks for achievement unlocks
	onUnlock func(entityID uint64, achievementID string, tier AchievementTier)
}

// NewExtendedAchievementSystem creates a new extended achievement system.
func NewExtendedAchievementSystem(world *World) *ExtendedAchievementSystem {
	var logger *logrus.Entry
	if world != nil && world.logger != nil {
		logger = world.logger.Logger.WithField("system", "extended_achievement")
	}

	return &ExtendedAchievementSystem{
		world:  world,
		logger: logger,
	}
}

// SetOnUnlockCallback sets a callback for when achievements are unlocked.
func (s *ExtendedAchievementSystem) SetOnUnlockCallback(callback func(entityID uint64, achievementID string, tier AchievementTier)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onUnlock = callback
}

// Update processes achievement updates. Most achievements are event-driven.
func (s *ExtendedAchievementSystem) Update(entities []*Entity, deltaTime float64) {
	// Achievement updates are event-driven, not time-based
	// This method exists for interface compliance
}

// RecordProgress records progress for an achievement.
// Returns true if a new tier was unlocked.
func (s *ExtendedAchievementSystem) RecordProgress(entityID uint64, achievementID string, progress int64) bool {
	if s.world == nil {
		return false
	}

	entity, exists := s.world.GetEntity(entityID)
	if !exists || entity == nil {
		return false
	}

	def, defExists := GetAchievementDefinition(achievementID)
	if !defExists {
		if s.logger != nil {
			s.logger.WithField("achievement_id", achievementID).Warn("Unknown achievement ID")
		}
		return false
	}

	comp := s.getOrCreateComponent(entity)
	oldTier := comp.GetTier(achievementID)
	timestamp := s.world.Clock.Now().Unix()

	unlocked := comp.SetProgress(achievementID, def.Category, progress, def.Thresholds, timestamp)

	if unlocked {
		newTier := comp.GetTier(achievementID)
		s.notifyUnlock(entityID, achievementID, newTier, oldTier)
	}

	return unlocked
}

// IncrementProgress adds to progress for an achievement.
// Returns true if a new tier was unlocked.
func (s *ExtendedAchievementSystem) IncrementProgress(entityID uint64, achievementID string, amount int64) bool {
	if s.world == nil {
		return false
	}

	entity, exists := s.world.GetEntity(entityID)
	if !exists || entity == nil {
		return false
	}

	def, defExists := GetAchievementDefinition(achievementID)
	if !defExists {
		if s.logger != nil {
			s.logger.WithField("achievement_id", achievementID).Warn("Unknown achievement ID")
		}
		return false
	}

	comp := s.getOrCreateComponent(entity)
	oldTier := comp.GetTier(achievementID)
	timestamp := s.world.Clock.Now().Unix()

	unlocked := comp.IncrementProgress(achievementID, def.Category, amount, def.Thresholds, timestamp)

	if unlocked {
		newTier := comp.GetTier(achievementID)
		s.notifyUnlock(entityID, achievementID, newTier, oldTier)
	}

	return unlocked
}

// getOrCreateComponent gets or creates an ExtendedAchievementComponent.
func (s *ExtendedAchievementSystem) getOrCreateComponent(entity *Entity) *ExtendedAchievementComponent {
	compRaw, exists := entity.GetComponent("extended_achievement")
	if exists {
		if comp, ok := compRaw.(*ExtendedAchievementComponent); ok {
			return comp
		}
	}

	comp := NewExtendedAchievementComponent()
	entity.AddComponent(comp)
	return comp
}

// notifyUnlock handles achievement unlock notifications.
func (s *ExtendedAchievementSystem) notifyUnlock(entityID uint64, achievementID string, newTier, oldTier AchievementTier) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"achievement_id": achievementID,
			"old_tier":       oldTier.String(),
			"new_tier":       newTier.String(),
		}).Info("Achievement tier unlocked")
	}

	s.mu.RLock()
	callback := s.onUnlock
	s.mu.RUnlock()

	if callback != nil {
		callback(entityID, achievementID, newTier)
	}
}

// GetProgress returns current progress for an achievement.
func (s *ExtendedAchievementSystem) GetProgress(entityID uint64, achievementID string) int64 {
	if s.world == nil {
		return 0
	}

	entity, exists := s.world.GetEntity(entityID)
	if !exists || entity == nil {
		return 0
	}

	compRaw, exists := entity.GetComponent("extended_achievement")
	if !exists {
		return 0
	}

	comp, ok := compRaw.(*ExtendedAchievementComponent)
	if !ok {
		return 0
	}

	return comp.GetProgress(achievementID)
}

// GetTier returns current tier for an achievement.
func (s *ExtendedAchievementSystem) GetTier(entityID uint64, achievementID string) AchievementTier {
	if s.world == nil {
		return AchievementTierNone
	}

	entity, exists := s.world.GetEntity(entityID)
	if !exists || entity == nil {
		return AchievementTierNone
	}

	compRaw, exists := entity.GetComponent("extended_achievement")
	if !exists {
		return AchievementTierNone
	}

	comp, ok := compRaw.(*ExtendedAchievementComponent)
	if !ok {
		return AchievementTierNone
	}

	return comp.GetTier(achievementID)
}

// GetTotalPoints returns total achievement points for an entity.
func (s *ExtendedAchievementSystem) GetTotalPoints(entityID uint64) int {
	if s.world == nil {
		return 0
	}

	entity, exists := s.world.GetEntity(entityID)
	if !exists || entity == nil {
		return 0
	}

	compRaw, exists := entity.GetComponent("extended_achievement")
	if !exists {
		return 0
	}

	comp, ok := compRaw.(*ExtendedAchievementComponent)
	if !ok {
		return 0
	}

	return comp.GetTotalPoints()
}

// GetCategoryPoints returns achievement points for a category.
func (s *ExtendedAchievementSystem) GetCategoryPoints(entityID uint64, category AchievementCategory) int {
	if s.world == nil {
		return 0
	}

	entity, exists := s.world.GetEntity(entityID)
	if !exists || entity == nil {
		return 0
	}

	compRaw, exists := entity.GetComponent("extended_achievement")
	if !exists {
		return 0
	}

	comp, ok := compRaw.(*ExtendedAchievementComponent)
	if !ok {
		return 0
	}

	return comp.GetCategoryPoints(category)
}

// OnEnemyKilled is called when an enemy is killed.
func (s *ExtendedAchievementSystem) OnEnemyKilled(entityID uint64, isBoss bool) {
	s.IncrementProgress(entityID, "combat_first_blood", 1)
	if isBoss {
		s.IncrementProgress(entityID, "combat_boss_slayer", 1)
	}
}

// OnDamageDealt is called when damage is dealt.
func (s *ExtendedAchievementSystem) OnDamageDealt(entityID uint64, damage int64, isCritical bool) {
	s.IncrementProgress(entityID, "combat_damage_dealer", damage)
	if isCritical {
		s.IncrementProgress(entityID, "combat_critical_master", 1)
	}
}

// OnSurvived is called when surviving at low health.
func (s *ExtendedAchievementSystem) OnSurvived(entityID uint64) {
	s.IncrementProgress(entityID, "combat_survivor", 1)
}

// OnQuestCompleted is called when a quest is completed.
func (s *ExtendedAchievementSystem) OnQuestCompleted(entityID uint64, isMainStory, isSideQuest, isLegendary bool) {
	s.IncrementProgress(entityID, "quest_adventurer", 1)
	if isMainStory {
		s.IncrementProgress(entityID, "quest_main_story", 1)
	}
	if isSideQuest {
		s.IncrementProgress(entityID, "quest_side_tracker", 1)
	}
	if isLegendary {
		s.IncrementProgress(entityID, "quest_legendary", 1)
	}
}

// OnItemCrafted is called when an item is crafted.
func (s *ExtendedAchievementSystem) OnItemCrafted(entityID uint64, isRare, isPerfect bool) {
	s.IncrementProgress(entityID, "craft_apprentice", 1)
	if isRare {
		s.IncrementProgress(entityID, "craft_quality", 1)
	}
	if isPerfect {
		s.IncrementProgress(entityID, "craft_perfect", 1)
	}
}

// OnAreaVisited is called when visiting a new area.
func (s *ExtendedAchievementSystem) OnAreaVisited(entityID uint64) {
	s.IncrementProgress(entityID, "explore_wanderer", 1)
}

// OnDungeonCleared is called when clearing a dungeon.
func (s *ExtendedAchievementSystem) OnDungeonCleared(entityID uint64) {
	s.IncrementProgress(entityID, "explore_dungeon_delver", 1)
}

// OnSecretFound is called when finding a secret.
func (s *ExtendedAchievementSystem) OnSecretFound(entityID uint64) {
	s.IncrementProgress(entityID, "explore_secret_finder", 1)
}

// OnPvPMatchWon is called when winning a PvP match.
func (s *ExtendedAchievementSystem) OnPvPMatchWon(entityID uint64, is1v1, isTeam bool) {
	s.IncrementProgress(entityID, "pvp_gladiator", 1)
	if is1v1 {
		s.IncrementProgress(entityID, "pvp_duelist", 1)
	}
	if isTeam {
		s.IncrementProgress(entityID, "pvp_team_player", 1)
	}
}

// OnTournamentWon is called when winning a tournament.
func (s *ExtendedAchievementSystem) OnTournamentWon(entityID uint64) {
	s.IncrementProgress(entityID, "pvp_tournament_victor", 1)
}

// OnHonorEarned is called when earning honor points.
func (s *ExtendedAchievementSystem) OnHonorEarned(entityID uint64, amount int64) {
	s.IncrementProgress(entityID, "pvp_honor_bound", amount)
}

// OnFriendAdded is called when adding a friend.
func (s *ExtendedAchievementSystem) OnFriendAdded(entityID uint64) {
	s.IncrementProgress(entityID, "social_friend_maker", 1)
}

// OnGuildJoined is called when joining a guild.
func (s *ExtendedAchievementSystem) OnGuildJoined(entityID uint64) {
	s.IncrementProgress(entityID, "social_guild_member", 1)
}

// OnTradeCompleted is called when completing a trade.
func (s *ExtendedAchievementSystem) OnTradeCompleted(entityID uint64) {
	s.IncrementProgress(entityID, "social_trade_master", 1)
}

// OnChatMessage is called when sending a chat message.
func (s *ExtendedAchievementSystem) OnChatMessage(entityID uint64) {
	s.IncrementProgress(entityID, "social_chat_active", 1)
}

// OnDistanceTraveled is called when traveling distance.
func (s *ExtendedAchievementSystem) OnDistanceTraveled(entityID uint64, distance int64) {
	s.IncrementProgress(entityID, "explore_distance", distance)
}

// GetAchievementCount returns the count of definitions.
func GetAchievementCount() int {
	return len(achievementDefinitions)
}
