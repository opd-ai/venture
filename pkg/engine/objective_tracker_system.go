// Package engine provides quest objective tracking system.
package engine

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/quest"
	log "github.com/sirupsen/logrus"
)

// ObjectiveTrackerSystem monitors game events and updates quest objectives.
// This system tracks:
// - Enemy kills (TypeKill quests)
// - Item collection (TypeCollect quests)
// - Tile exploration (TypeExplore quests)
// - Boss defeats (TypeBoss quests)
type ObjectiveTrackerSystem struct {
	// Callbacks for reward distribution
	onQuestComplete func(entity *Entity, qst *quest.Quest)

	// Tracking state
	exploredTiles map[uint64]map[string]bool // entityID -> tileKey -> explored

	// Item generator for quest rewards
	itemGenerator *item.ItemGenerator
}

// NewObjectiveTrackerSystem creates a new objective tracker.
func NewObjectiveTrackerSystem() *ObjectiveTrackerSystem {
	log.WithFields(log.Fields{
		"system_name": "objective_tracker",
	}).Debug("Creating new objective tracker system")

	s := &ObjectiveTrackerSystem{
		exploredTiles: make(map[uint64]map[string]bool),
		itemGenerator: item.NewItemGenerator(),
	}

	log.WithFields(log.Fields{
		"system_name": "objective_tracker",
	}).Info("Objective tracker system initialized")

	return s
}

// SetQuestCompleteCallback sets the function called when a quest completes.
func (s *ObjectiveTrackerSystem) SetQuestCompleteCallback(callback func(entity *Entity, qst *quest.Quest)) {
	log.WithFields(log.Fields{
		"system_name":  "objective_tracker",
		"callback_set": callback != nil,
	}).Debug("Setting quest complete callback")

	s.onQuestComplete = callback
}

// Update processes quest objectives based on game state.
func (s *ObjectiveTrackerSystem) Update(entities []*Entity, deltaTime float64) {
	debugLog := log.GetLevel() >= log.DebugLevel
	if debugLog {
		log.WithFields(log.Fields{
			"system_name":   "objective_tracker",
			"entity_count":  len(entities),
			"delta_time_ms": deltaTime * 1000,
		}).Debug("Updating objective tracker system")
	}

	entitiesProcessed := 0

	// Update exploration objectives for entities with position
	for _, entity := range entities {
		if !entity.HasComponent("questtracker") {
			continue
		}

		entitiesProcessed++

		if debugLog {
			log.WithFields(log.Fields{
				"system_name": "objective_tracker",
				"entity_id":   entity.ID,
			}).Debug("Processing entity objectives")
		}

		// Track exploration
		s.updateExplorationObjectives(entity)

		// Check for newly completed quests
		s.checkQuestCompletion(entity)
	}

	if debugLog {
		log.WithFields(log.Fields{
			"system_name":        "objective_tracker",
			"entities_processed": entitiesProcessed,
		}).Debug("Objective tracker system update complete")
	}
}

// OnEnemyKilled should be called by combat system when an enemy dies.
func (s *ObjectiveTrackerSystem) OnEnemyKilled(killer, enemy *Entity) {
	log.WithFields(log.Fields{
		"system_name": "objective_tracker",
		"killer_id":   killer.ID,
		"enemy_id":    enemy.ID,
	}).Debug("Processing enemy kill event")

	if !killer.HasComponent("questtracker") {
		log.WithFields(log.Fields{
			"system_name": "objective_tracker",
			"killer_id":   killer.ID,
		}).Debug("Killer has no quest tracker component")
		return
	}

	comp, ok := killer.GetComponent("questtracker")
	if !ok {
		log.WithFields(log.Fields{
			"system_name": "objective_tracker",
			"killer_id":   killer.ID,
		}).Warn("Failed to retrieve quest tracker component")
		return
	}
	tracker, ok := comp.(*QuestTrackerComponent)
	if !ok {
		log.WithFields(log.Fields{
			"system_name":    "objective_tracker",
			"killer_id":      killer.ID,
			"component_type": fmt.Sprintf("%T", comp),
		}).Error("Quest tracker component has invalid type")
		return
	}

	// For now, all enemies count as "enemy" or "monster"
	// In future, could extract type from entity components
	enemyName := "enemy"

	objectivesUpdated := 0

	// Update kill objectives
	for _, tracked := range tracker.ActiveQuests {
		if tracked.Quest.Type != quest.TypeKill && tracked.Quest.Type != quest.TypeBoss {
			continue
		}

		for i, obj := range tracked.Quest.Objectives {
			// Check if objective targets this enemy type
			if s.matchesTarget(obj.Target, enemyName, "kill") {
				log.WithFields(log.Fields{
					"system_name":     "objective_tracker",
					"entity_id":       killer.ID,
					"quest_id":        tracked.Quest.ID,
					"objective_index": i,
					"target":          obj.Target,
					"enemy_name":      enemyName,
				}).Debug("Incrementing kill objective progress")

				tracker.IncrementProgress(tracked.Quest.ID, i, 1)
				objectivesUpdated++
			}
		}
	}

	log.WithFields(log.Fields{
		"system_name":        "objective_tracker",
		"entity_id":          killer.ID,
		"objectives_updated": objectivesUpdated,
	}).Debug("Enemy kill event processing complete")
}

// OnItemCollected should be called by inventory system when player picks up item.
func (s *ObjectiveTrackerSystem) OnItemCollected(collector *Entity, itemName string) {
	log.WithFields(log.Fields{
		"system_name": "objective_tracker",
		"entity_id":   collector.ID,
		"item_name":   itemName,
	}).Debug("Processing item collection event")

	if !collector.HasComponent("questtracker") {
		log.WithFields(log.Fields{
			"system_name": "objective_tracker",
			"entity_id":   collector.ID,
		}).Debug("Collector has no quest tracker component")
		return
	}

	comp, ok := collector.GetComponent("questtracker")
	if !ok {
		log.WithFields(log.Fields{
			"system_name": "objective_tracker",
			"entity_id":   collector.ID,
		}).Warn("Failed to retrieve quest tracker component")
		return
	}
	tracker, ok := comp.(*QuestTrackerComponent)
	if !ok {
		log.WithFields(log.Fields{
			"system_name":    "objective_tracker",
			"entity_id":      collector.ID,
			"component_type": fmt.Sprintf("%T", comp),
		}).Error("Quest tracker component has invalid type")
		return
	}

	objectivesUpdated := 0

	// Update collect objectives
	for _, tracked := range tracker.ActiveQuests {
		if tracked.Quest.Type != quest.TypeCollect {
			continue
		}

		for i, obj := range tracked.Quest.Objectives {
			// Check if objective targets this item
			if s.matchesTarget(obj.Target, itemName, "collect") {
				log.WithFields(log.Fields{
					"system_name":     "objective_tracker",
					"entity_id":       collector.ID,
					"quest_id":        tracked.Quest.ID,
					"objective_index": i,
					"target":          obj.Target,
					"item_name":       itemName,
				}).Debug("Incrementing collect objective progress")

				tracker.IncrementProgress(tracked.Quest.ID, i, 1)
				objectivesUpdated++
			}
		}
	}

	log.WithFields(log.Fields{
		"system_name":        "objective_tracker",
		"entity_id":          collector.ID,
		"objectives_updated": objectivesUpdated,
	}).Debug("Item collection event processing complete")
}

// GAP-014 REPAIR: OnUIOpened should be called when player opens a UI screen.
// Parameters:
//
//	entity - Entity opening the UI (usually player)
//	uiName - Name of UI screen: "inventory", "quest_log", "character", "skills", "map"
//
// Called by: InputSystem callbacks when UI toggle keys are pressed
func (s *ObjectiveTrackerSystem) OnUIOpened(entity *Entity, uiName string) {
	log.WithFields(log.Fields{
		"system_name": "objective_tracker",
		"entity_id":   entity.ID,
		"ui_name":     uiName,
	}).Debug("Processing UI opened event")

	if !entity.HasComponent("questtracker") {
		log.WithFields(log.Fields{
			"system_name": "objective_tracker",
			"entity_id":   entity.ID,
		}).Debug("Entity has no quest tracker component")
		return
	}

	comp, ok := entity.GetComponent("questtracker")
	if !ok {
		log.WithFields(log.Fields{
			"system_name": "objective_tracker",
			"entity_id":   entity.ID,
		}).Warn("Failed to retrieve quest tracker component")
		return
	}
	tracker, ok := comp.(*QuestTrackerComponent)
	if !ok {
		log.WithFields(log.Fields{
			"system_name":    "objective_tracker",
			"entity_id":      entity.ID,
			"component_type": fmt.Sprintf("%T", comp),
		}).Error("Quest tracker component has invalid type")
		return
	}

	objectivesUpdated := 0

	// Update UI interaction objectives (used in tutorial quests)
	for _, tracked := range tracker.ActiveQuests {
		for i, obj := range tracked.Quest.Objectives {
			// Check if objective targets this UI
			if s.matchesTarget(obj.Target, uiName, "ui") {
				log.WithFields(log.Fields{
					"system_name":     "objective_tracker",
					"entity_id":       entity.ID,
					"quest_id":        tracked.Quest.ID,
					"objective_index": i,
					"target":          obj.Target,
					"ui_name":         uiName,
				}).Debug("Incrementing UI interaction objective progress")

				tracker.IncrementProgress(tracked.Quest.ID, i, 1)
				objectivesUpdated++
			}
		}
	}

	log.WithFields(log.Fields{
		"system_name":        "objective_tracker",
		"entity_id":          entity.ID,
		"objectives_updated": objectivesUpdated,
	}).Debug("UI opened event processing complete")
}

// OnTileExplored should be called by movement system when player enters new tile.
// initializeTileTracking initializes exploration tracking for an entity if not already present
func (s *ObjectiveTrackerSystem) initializeTileTracking(explorer *Entity) {
	if s.exploredTiles[explorer.ID] == nil {
		s.exploredTiles[explorer.ID] = make(map[string]bool)
		log.WithFields(log.Fields{
			"system_name": "objective_tracker",
			"entity_id":   explorer.ID,
		}).Debug("Initialized exploration tracking for entity")
	}
}

// recordTileExploration marks a tile as explored and returns whether it was newly explored
func (s *ObjectiveTrackerSystem) recordTileExploration(explorer *Entity, tileKey string) bool {
	if s.exploredTiles[explorer.ID][tileKey] {
		log.WithFields(log.Fields{
			"system_name": "objective_tracker",
			"entity_id":   explorer.ID,
			"tile_key":    tileKey,
		}).Debug("Tile already explored")
		return false
	}
	s.exploredTiles[explorer.ID][tileKey] = true
	return true
}

// getQuestTracker retrieves and validates the quest tracker component from an entity
func (s *ObjectiveTrackerSystem) getQuestTracker(explorer *Entity) (*QuestTrackerComponent, bool) {
	comp, ok := explorer.GetComponent("questtracker")
	if !ok {
		log.WithFields(log.Fields{
			"system_name": "objective_tracker",
			"entity_id":   explorer.ID,
		}).Warn("Failed to retrieve quest tracker component")
		return nil, false
	}
	tracker, ok := comp.(*QuestTrackerComponent)
	if !ok {
		log.WithFields(log.Fields{
			"system_name":    "objective_tracker",
			"entity_id":      explorer.ID,
			"component_type": fmt.Sprintf("%T", comp),
		}).Error("Quest tracker component has invalid type")
		return nil, false
	}
	return tracker, true
}

// updateExploreObjectives updates all exploration-related quest objectives for an entity
func (s *ObjectiveTrackerSystem) updateExploreObjectives(explorer *Entity, tracker *QuestTrackerComponent, totalExplored int) int {
	objectivesUpdated := 0

	for _, tracked := range tracker.ActiveQuests {
		if tracked.Quest.Type != quest.TypeExplore {
			continue
		}

		for i, obj := range tracked.Quest.Objectives {
			if strings.Contains(strings.ToLower(obj.Target), "tile") ||
				strings.Contains(strings.ToLower(obj.Target), "dungeon") ||
				strings.Contains(strings.ToLower(obj.Target), "explore") {
				log.WithFields(log.Fields{
					"system_name":     "objective_tracker",
					"entity_id":       explorer.ID,
					"quest_id":        tracked.Quest.ID,
					"objective_index": i,
					"target":          obj.Target,
					"explored_count":  totalExplored,
				}).Debug("Updating explore objective progress")

				tracker.UpdateProgress(tracked.Quest.ID, i, totalExplored)
				objectivesUpdated++
			}
		}
	}

	return objectivesUpdated
}

func (s *ObjectiveTrackerSystem) OnTileExplored(explorer *Entity, x, y int) {
	log.WithFields(log.Fields{
		"system_name": "objective_tracker",
		"entity_id":   explorer.ID,
		"tile_x":      x,
		"tile_y":      y,
	}).Debug("Processing tile exploration event")

	if !explorer.HasComponent("questtracker") {
		log.WithFields(log.Fields{
			"system_name": "objective_tracker",
			"entity_id":   explorer.ID,
		}).Debug("Explorer has no quest tracker component")
		return
	}

	s.initializeTileTracking(explorer)

	tileKey := tileKeyFromCoords(x, y)
	if !s.recordTileExploration(explorer, tileKey) {
		return
	}

	totalExplored := len(s.exploredTiles[explorer.ID])

	log.WithFields(log.Fields{
		"system_name":    "objective_tracker",
		"entity_id":      explorer.ID,
		"tile_key":       tileKey,
		"total_explored": totalExplored,
	}).Debug("New tile explored")

	tracker, ok := s.getQuestTracker(explorer)
	if !ok {
		return
	}

	objectivesUpdated := s.updateExploreObjectives(explorer, tracker, totalExplored)

	log.WithFields(log.Fields{
		"system_name":        "objective_tracker",
		"entity_id":          explorer.ID,
		"objectives_updated": objectivesUpdated,
	}).Debug("Tile exploration event processing complete")
}

// updateExplorationObjectives updates exploration progress based on current position.
func (s *ObjectiveTrackerSystem) updateExplorationObjectives(entity *Entity) {
	log.WithFields(log.Fields{
		"system_name": "objective_tracker",
		"entity_id":   entity.ID,
	}).Debug("Updating exploration objectives")

	posComp, ok := entity.GetComponent("position")
	if !ok {
		log.WithFields(log.Fields{
			"system_name": "objective_tracker",
			"entity_id":   entity.ID,
		}).Debug("Entity has no position component")
		return
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		log.WithFields(log.Fields{
			"system_name":    "objective_tracker",
			"entity_id":      entity.ID,
			"component_type": fmt.Sprintf("%T", posComp),
		}).Error("Position component has invalid type")
		return
	}

	// Convert world coordinates to tile coordinates (assuming 32-pixel tiles)
	tileX := int(pos.X / 32)
	tileY := int(pos.Y / 32)

	s.OnTileExplored(entity, tileX, tileY)
}

// checkQuestCompletion checks if any active quests have been completed.
func (s *ObjectiveTrackerSystem) checkQuestCompletion(entity *Entity) {
	log.WithFields(log.Fields{
		"system_name": "objective_tracker",
		"entity_id":   entity.ID,
	}).Debug("Checking quest completion")

	comp, ok := entity.GetComponent("questtracker")
	if !ok {
		log.WithFields(log.Fields{
			"system_name": "objective_tracker",
			"entity_id":   entity.ID,
		}).Debug("Entity has no quest tracker component")
		return
	}
	tracker, ok := comp.(*QuestTrackerComponent)
	if !ok {
		log.WithFields(log.Fields{
			"system_name":    "objective_tracker",
			"entity_id":      entity.ID,
			"component_type": fmt.Sprintf("%T", comp),
		}).Error("Quest tracker component has invalid type")
		return
	}

	completedQuests := 0

	// Check each active quest
	for _, tracked := range tracker.ActiveQuests {
		if tracked.Status != QuestStatusActive {
			continue
		}

		// Check if all objectives are complete
		if tracked.Quest.IsComplete() {
			// Mark quest as complete with current timestamp
			timestamp := time.Now().Unix()

			log.WithFields(log.Fields{
				"system_name": "objective_tracker",
				"entity_id":   entity.ID,
				"quest_id":    tracked.Quest.ID,
				"quest_name":  tracked.Quest.Name,
				"timestamp":   timestamp,
			}).Info("Quest completed")

			tracker.CompleteQuest(tracked.Quest.ID, timestamp)
			completedQuests++

			// Trigger completion callback for rewards
			if s.onQuestComplete != nil {
				log.WithFields(log.Fields{
					"system_name": "objective_tracker",
					"entity_id":   entity.ID,
					"quest_id":    tracked.Quest.ID,
				}).Debug("Triggering quest completion callback")

				s.onQuestComplete(entity, tracked.Quest)
			} else {
				log.WithFields(log.Fields{
					"system_name": "objective_tracker",
					"entity_id":   entity.ID,
					"quest_id":    tracked.Quest.ID,
				}).Warn("No quest completion callback set")
			}
		}
	}

	if completedQuests > 0 {
		log.WithFields(log.Fields{
			"system_name":      "objective_tracker",
			"entity_id":        entity.ID,
			"quests_completed": completedQuests,
		}).Info("Quest completion check complete")
	}
}

// matchesTarget checks if an item/enemy name matches a quest objective target.
func (s *ObjectiveTrackerSystem) matchesTarget(target, name, context string) bool {
	log.WithFields(log.Fields{
		"system_name": "objective_tracker",
		"target":      target,
		"name":        name,
		"context":     context,
	}).Debug("Checking target match")

	targetLower := strings.ToLower(target)
	nameLower := strings.ToLower(name)

	// Check exact and partial matches
	if s.matchesExactOrPartial(targetLower, nameLower, target, name) {
		return true
	}

	// Check context-specific matching
	matched, matchType := s.matchesContext(context, targetLower, nameLower)
	if matched {
		s.logTargetMatch(target, name, matchType)
	}

	return matched
}

// matchesExactOrPartial checks for exact or partial string matches.
func (s *ObjectiveTrackerSystem) matchesExactOrPartial(targetLower, nameLower, target, name string) bool {
	// Exact match
	if targetLower == nameLower {
		s.logTargetMatch(target, name, "exact")
		return true
	}

	// Partial match (target contains name or vice versa)
	if strings.Contains(targetLower, nameLower) || strings.Contains(nameLower, targetLower) {
		s.logTargetMatch(target, name, "partial")
		return true
	}

	return false
}

// matchesContext checks context-specific target matching rules.
func (s *ObjectiveTrackerSystem) matchesContext(context, targetLower, nameLower string) (bool, string) {
	switch context {
	case "kill":
		return s.matchesKillContext(targetLower)
	case "collect":
		return s.matchesCollectContext(targetLower)
	case "ui":
		return s.matchesUIContext(targetLower, nameLower)
	}
	return false, ""
}

// matchesKillContext checks if target matches generic kill objectives.
func (s *ObjectiveTrackerSystem) matchesKillContext(targetLower string) (bool, string) {
	if targetLower == "enemy" || targetLower == "enemies" || targetLower == "monster" {
		return true, "generic_kill"
	}
	return false, ""
}

// matchesCollectContext checks if target matches generic collect objectives.
func (s *ObjectiveTrackerSystem) matchesCollectContext(targetLower string) (bool, string) {
	if targetLower == "item" || targetLower == "items" {
		return true, "generic_collect"
	}
	return false, ""
}

// matchesUIContext checks if target matches UI objectives for tutorial.
func (s *ObjectiveTrackerSystem) matchesUIContext(targetLower, nameLower string) (bool, string) {
	// GAP-014 REPAIR: UI objective matching (for tutorial)
	if targetLower == "inventory" && nameLower == "inventory" {
		return true, "ui_inventory"
	}
	if targetLower == "quest_log" && nameLower == "quest_log" {
		return true, "ui_quest_log"
	}
	if strings.Contains(targetLower, "character") && nameLower == "character" {
		return true, "ui_character"
	}
	if strings.Contains(targetLower, "skill") && nameLower == "skills" {
		return true, "ui_skills"
	}
	if strings.Contains(targetLower, "map") && nameLower == "map" {
		return true, "ui_map"
	}
	return false, ""
}

// logTargetMatch logs successful target matches.
func (s *ObjectiveTrackerSystem) logTargetMatch(target, name, matchType string) {
	log.WithFields(log.Fields{
		"system_name": "objective_tracker",
		"target":      target,
		"name":        name,
		"match_type":  matchType,
	}).Debug("Target matched")
}

// tileKeyFromCoords creates a unique key for a tile position.
func tileKeyFromCoords(x, y int) string {
	// Use a simple string format for tile coordinates
	return fmt.Sprintf("%d,%d", x, y)
}

// AwardQuestRewards distributes rewards from a completed quest.
func (s *ObjectiveTrackerSystem) AwardQuestRewards(entity *Entity, qst *quest.Quest) {
	log.WithFields(log.Fields{
		"system_name":         "objective_tracker",
		"entity_id":           entity.ID,
		"quest_id":            qst.ID,
		"quest_name":          qst.Name,
		"xp_reward":           qst.Reward.XP,
		"gold_reward":         qst.Reward.Gold,
		"skill_points_reward": qst.Reward.SkillPoints,
		"item_count":          len(qst.Reward.Items),
	}).Info("Awarding quest rewards")

	s.awardXPReward(entity, qst)
	s.awardGoldReward(entity, qst)
	s.awardSkillPointsReward(entity, qst)

	if len(qst.Reward.Items) > 0 {
		log.WithFields(log.Fields{
			"system_name": "objective_tracker",
			"entity_id":   entity.ID,
			"quest_id":    qst.ID,
			"item_count":  len(qst.Reward.Items),
		}).Debug("Awarding item rewards")
		s.awardQuestItems(entity, qst)
	}

	log.WithFields(log.Fields{
		"system_name": "objective_tracker",
		"entity_id":   entity.ID,
		"quest_id":    qst.ID,
	}).Info("Quest rewards awarded")
}

// awardXPReward awards experience points to an entity from quest rewards.
func (s *ObjectiveTrackerSystem) awardXPReward(entity *Entity, qst *quest.Quest) {
	if qst.Reward.XP <= 0 {
		return
	}

	expComp, ok := entity.GetComponent("experience")
	if !ok {
		return
	}

	exp, ok := expComp.(*ExperienceComponent)
	if !ok {
		return
	}

	log.WithFields(log.Fields{
		"system_name": "objective_tracker",
		"entity_id":   entity.ID,
		"quest_id":    qst.ID,
		"xp_awarded":  qst.Reward.XP,
	}).Debug("Awarded XP reward")
	exp.AddXP(qst.Reward.XP)
}

// awardGoldReward awards gold to an entity from quest rewards.
func (s *ObjectiveTrackerSystem) awardGoldReward(entity *Entity, qst *quest.Quest) {
	if qst.Reward.Gold <= 0 {
		return
	}

	invComp, ok := entity.GetComponent("inventory")
	if !ok {
		return
	}

	inv, ok := invComp.(*InventoryComponent)
	if !ok {
		return
	}

	log.WithFields(log.Fields{
		"system_name":  "objective_tracker",
		"entity_id":    entity.ID,
		"quest_id":     qst.ID,
		"gold_awarded": qst.Reward.Gold,
	}).Debug("Awarded gold reward")
	inv.Gold += qst.Reward.Gold
}

// awardSkillPointsReward awards skill points to an entity from quest rewards.
func (s *ObjectiveTrackerSystem) awardSkillPointsReward(entity *Entity, qst *quest.Quest) {
	if qst.Reward.SkillPoints <= 0 {
		return
	}

	expComp, ok := entity.GetComponent("experience")
	if !ok {
		return
	}

	exp, ok := expComp.(*ExperienceComponent)
	if !ok {
		return
	}

	log.WithFields(log.Fields{
		"system_name":          "objective_tracker",
		"entity_id":            entity.ID,
		"quest_id":             qst.ID,
		"skill_points_awarded": qst.Reward.SkillPoints,
	}).Debug("Awarded skill points reward")
	exp.SkillPoints += qst.Reward.SkillPoints
}

// awardQuestItems generates and awards items from quest rewards.
// extractEntityGenre retrieves the genre ID from entity's genre component.
func extractEntityGenre(entity *Entity) string {
	genreID := "fantasy"
	if genreComp, hasGenre := entity.GetComponent("genre"); hasGenre {
		if genre, ok := genreComp.(*GenreComponent); ok {
			genreID = genre.GetPrimaryGenre()
		}
	}
	return genreID
}

// createItemGenerationParams creates generation parameters for quest item rewards.
func createItemGenerationParams(qst *quest.Quest, genreID, itemType string) procgen.GenerationParams {
	return procgen.GenerationParams{
		Difficulty: float64(qst.Difficulty) / 5.0,
		Depth:      qst.RequiredLevel,
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"count": 1,
			"type":  itemType,
		},
	}
}

// generateQuestItem generates a single quest item using the item generator.
func (s *ObjectiveTrackerSystem) generateQuestItem(itemSeed int64, params procgen.GenerationParams, itemName, questID string, entityID uint64, itemIndex int) *item.Item {
	result, err := s.itemGenerator.Generate(itemSeed, params)
	if err != nil {
		log.WithFields(log.Fields{
			"system_name": "objective_tracker",
			"entity_id":   entityID,
			"quest_id":    questID,
			"item_index":  itemIndex,
			"item_name":   itemName,
			"error":       err.Error(),
		}).Error("Failed to generate quest item reward")
		return nil
	}

	items, ok := result.([]*item.Item)
	if !ok || len(items) == 0 {
		log.WithFields(log.Fields{
			"system_name": "objective_tracker",
			"entity_id":   entityID,
			"quest_id":    questID,
			"item_index":  itemIndex,
		}).Warn("Item generation returned invalid or empty result")
		return nil
	}

	return items[0]
}

func (s *ObjectiveTrackerSystem) awardQuestItems(entity *Entity, qst *quest.Quest) {
	log.WithFields(log.Fields{
		"system_name": "objective_tracker",
		"entity_id":   entity.ID,
		"quest_id":    qst.ID,
		"item_count":  len(qst.Reward.Items),
	}).Debug("Generating quest item rewards")

	invComp, ok := entity.GetComponent("inventory")
	if !ok {
		log.WithFields(log.Fields{
			"system_name": "objective_tracker",
			"entity_id":   entity.ID,
			"quest_id":    qst.ID,
		}).Warn("Entity has no inventory component for item rewards")
		return
	}
	inv, ok := invComp.(*InventoryComponent)
	if !ok {
		log.WithFields(log.Fields{
			"system_name":    "objective_tracker",
			"entity_id":      entity.ID,
			"quest_id":       qst.ID,
			"component_type": fmt.Sprintf("%T", invComp),
		}).Error("Inventory component has invalid type")
		return
	}

	genreID := extractEntityGenre(entity)

	log.WithFields(log.Fields{
		"system_name": "objective_tracker",
		"entity_id":   entity.ID,
		"quest_id":    qst.ID,
		"genre_id":    genreID,
	}).Debug("Using genre for item generation")

	itemsAwarded := 0

	for i, itemName := range qst.Reward.Items {
		itemSeed := qst.Seed + int64(i)*100
		itemType := s.inferItemTypeFromName(itemName, itemSeed)

		log.WithFields(log.Fields{
			"system_name": "objective_tracker",
			"entity_id":   entity.ID,
			"quest_id":    qst.ID,
			"item_index":  i,
			"item_name":   itemName,
			"item_type":   itemType,
			"seed":        itemSeed,
		}).Debug("Generating quest item reward")

		params := createItemGenerationParams(qst, genreID, itemType)
		generatedItem := s.generateQuestItem(itemSeed, params, itemName, qst.ID, entity.ID, i)
		if generatedItem == nil {
			continue
		}

		if inv.CanAddItem(generatedItem) {
			log.WithFields(log.Fields{
				"system_name": "objective_tracker",
				"entity_id":   entity.ID,
				"quest_id":    qst.ID,
				"item_index":  i,
				"item_name":   generatedItem.Name,
				"item_rarity": generatedItem.Rarity,
			}).Debug("Added quest item to inventory")
			inv.AddItem(generatedItem)
			itemsAwarded++
		} else {
			log.WithFields(log.Fields{
				"system_name": "objective_tracker",
				"entity_id":   entity.ID,
				"quest_id":    qst.ID,
				"item_name":   generatedItem.Name,
			}).Warn("Inventory full, cannot add quest item reward")
		}
	}

	log.WithFields(log.Fields{
		"system_name":   "objective_tracker",
		"entity_id":     entity.ID,
		"quest_id":      qst.ID,
		"items_awarded": itemsAwarded,
		"items_total":   len(qst.Reward.Items),
	}).Info("Quest item rewards processing complete")
}

// inferItemTypeFromName attempts to determine item type from the item name string.
// Quest item names are descriptive (e.g., "healing potion", "iron sword", "leather armor").
// The seed parameter ensures deterministic fallback selection when no keyword matches.
func (s *ObjectiveTrackerSystem) inferItemTypeFromName(itemName string, seed int64) string {
	log.WithFields(log.Fields{
		"system_name": "objective_tracker",
		"item_name":   itemName,
	}).Debug("Inferring item type from name")

	nameLower := strings.ToLower(itemName)

	// Check for weapon keywords
	weaponKeywords := []string{"sword", "axe", "bow", "staff", "dagger", "mace", "spear", "hammer"}
	for _, keyword := range weaponKeywords {
		if strings.Contains(nameLower, keyword) {
			log.WithFields(log.Fields{
				"system_name": "objective_tracker",
				"item_name":   itemName,
				"item_type":   "weapon",
				"keyword":     keyword,
			}).Debug("Inferred item type")
			return "weapon"
		}
	}

	// Check for armor keywords
	armorKeywords := []string{"armor", "helm", "helmet", "boots", "gloves", "shield", "chest", "plate", "mail"}
	for _, keyword := range armorKeywords {
		if strings.Contains(nameLower, keyword) {
			log.WithFields(log.Fields{
				"system_name": "objective_tracker",
				"item_name":   itemName,
				"item_type":   "armor",
				"keyword":     keyword,
			}).Debug("Inferred item type")
			return "armor"
		}
	}

	// Check for consumable keywords
	consumableKeywords := []string{"potion", "elixir", "scroll", "food", "drink", "medicine"}
	for _, keyword := range consumableKeywords {
		if strings.Contains(nameLower, keyword) {
			log.WithFields(log.Fields{
				"system_name": "objective_tracker",
				"item_name":   itemName,
				"item_type":   "consumable",
				"keyword":     keyword,
			}).Debug("Inferred item type")
			return "consumable"
		}
	}

	// Default to deterministic item type selection based on seed
	itemTypes := []string{"weapon", "armor", "consumable"}
	rng := rand.New(rand.NewSource(seed))
	selectedType := itemTypes[rng.Intn(len(itemTypes))]

	log.WithFields(log.Fields{
		"system_name": "objective_tracker",
		"item_name":   itemName,
		"item_type":   selectedType,
		"method":      "random_fallback",
	}).Debug("Inferred item type using fallback")

	return selectedType
}
