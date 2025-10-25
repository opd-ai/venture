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
	return &ObjectiveTrackerSystem{
		exploredTiles: make(map[uint64]map[string]bool),
		itemGenerator: item.NewItemGenerator(),
	}
}

// SetQuestCompleteCallback sets the function called when a quest completes.
func (s *ObjectiveTrackerSystem) SetQuestCompleteCallback(callback func(entity *Entity, qst *quest.Quest)) {
	s.onQuestComplete = callback
}

// Update processes quest objectives based on game state.
func (s *ObjectiveTrackerSystem) Update(entities []*Entity, deltaTime float64) {
	// Update exploration objectives for entities with position
	for _, entity := range entities {
		if !entity.HasComponent("questtracker") {
			continue
		}

		// Track exploration
		s.updateExplorationObjectives(entity)

		// Check for newly completed quests
		s.checkQuestCompletion(entity)
	}
}

// OnEnemyKilled should be called by combat system when an enemy dies.
func (s *ObjectiveTrackerSystem) OnEnemyKilled(killer, enemy *Entity) {
	if !killer.HasComponent("questtracker") {
		return
	}

	comp, ok := killer.GetComponent("questtracker")
	if !ok {
		return
	}
	tracker := comp.(*QuestTrackerComponent)

	// For now, all enemies count as "enemy" or "monster"
	// In future, could extract type from entity components
	enemyName := "enemy"

	// Update kill objectives
	for _, tracked := range tracker.ActiveQuests {
		if tracked.Quest.Type != quest.TypeKill && tracked.Quest.Type != quest.TypeBoss {
			continue
		}

		for i, obj := range tracked.Quest.Objectives {
			// Check if objective targets this enemy type
			if s.matchesTarget(obj.Target, enemyName, "kill") {
				tracker.IncrementProgress(tracked.Quest.ID, i, 1)
			}
		}
	}
}

// OnItemCollected should be called by inventory system when player picks up item.
func (s *ObjectiveTrackerSystem) OnItemCollected(collector *Entity, itemName string) {
	if !collector.HasComponent("questtracker") {
		return
	}

	comp, ok := collector.GetComponent("questtracker")
	if !ok {
		return
	}
	tracker := comp.(*QuestTrackerComponent)

	// Update collect objectives
	for _, tracked := range tracker.ActiveQuests {
		if tracked.Quest.Type != quest.TypeCollect {
			continue
		}

		for i, obj := range tracked.Quest.Objectives {
			// Check if objective targets this item
			if s.matchesTarget(obj.Target, itemName, "collect") {
				tracker.IncrementProgress(tracked.Quest.ID, i, 1)
			}
		}
	}
}

// GAP-014 REPAIR: OnUIOpened should be called when player opens a UI screen.
// Parameters:
//
//	entity - Entity opening the UI (usually player)
//	uiName - Name of UI screen: "inventory", "quest_log", "character", "skills", "map"
//
// Called by: InputSystem callbacks when UI toggle keys are pressed
func (s *ObjectiveTrackerSystem) OnUIOpened(entity *Entity, uiName string) {
	if !entity.HasComponent("questtracker") {
		return
	}

	comp, ok := entity.GetComponent("questtracker")
	if !ok {
		return
	}
	tracker := comp.(*QuestTrackerComponent)

	// Update UI interaction objectives (used in tutorial quests)
	for _, tracked := range tracker.ActiveQuests {
		for i, obj := range tracked.Quest.Objectives {
			// Check if objective targets this UI
			if s.matchesTarget(obj.Target, uiName, "ui") {
				tracker.IncrementProgress(tracked.Quest.ID, i, 1)
			}
		}
	}
}

// OnTileExplored should be called by movement system when player enters new tile.
func (s *ObjectiveTrackerSystem) OnTileExplored(explorer *Entity, x, y int) {
	if !explorer.HasComponent("questtracker") {
		return
	}

	// Track unique tiles
	if s.exploredTiles[explorer.ID] == nil {
		s.exploredTiles[explorer.ID] = make(map[string]bool)
	}

	tileKey := tileKeyFromCoords(x, y)
	if s.exploredTiles[explorer.ID][tileKey] {
		return // Already explored
	}
	s.exploredTiles[explorer.ID][tileKey] = true

	comp, ok := explorer.GetComponent("questtracker")
	if !ok {
		return
	}
	tracker := comp.(*QuestTrackerComponent)

	// Update explore objectives
	for _, tracked := range tracker.ActiveQuests {
		if tracked.Quest.Type != quest.TypeExplore {
			continue
		}

		for i, obj := range tracked.Quest.Objectives {
			// Exploration objectives count unique tiles
			if strings.Contains(strings.ToLower(obj.Target), "tile") ||
				strings.Contains(strings.ToLower(obj.Target), "dungeon") ||
				strings.Contains(strings.ToLower(obj.Target), "explore") {
				tracker.UpdateProgress(tracked.Quest.ID, i, len(s.exploredTiles[explorer.ID]))
			}
		}
	}
}

// updateExplorationObjectives updates exploration progress based on current position.
func (s *ObjectiveTrackerSystem) updateExplorationObjectives(entity *Entity) {
	posComp, ok := entity.GetComponent("position")
	if !ok {
		return
	}
	pos := posComp.(*PositionComponent)

	// Convert world coordinates to tile coordinates (assuming 32-pixel tiles)
	tileX := int(pos.X / 32)
	tileY := int(pos.Y / 32)

	s.OnTileExplored(entity, tileX, tileY)
}

// checkQuestCompletion checks if any active quests have been completed.
func (s *ObjectiveTrackerSystem) checkQuestCompletion(entity *Entity) {
	comp, ok := entity.GetComponent("questtracker")
	if !ok {
		return
	}
	tracker := comp.(*QuestTrackerComponent)

	// Check each active quest
	for _, tracked := range tracker.ActiveQuests {
		if tracked.Status != QuestStatusActive {
			continue
		}

		// Check if all objectives are complete
		if tracked.Quest.IsComplete() {
			// Mark quest as complete with current timestamp
			timestamp := time.Now().Unix()
			tracker.CompleteQuest(tracked.Quest.ID, timestamp)

			// Trigger completion callback for rewards
			if s.onQuestComplete != nil {
				s.onQuestComplete(entity, tracked.Quest)
			}
		}
	}
}

// matchesTarget checks if an item/enemy name matches a quest objective target.
func (s *ObjectiveTrackerSystem) matchesTarget(target, name, context string) bool {
	targetLower := strings.ToLower(target)
	nameLower := strings.ToLower(name)

	// Exact match
	if targetLower == nameLower {
		return true
	}

	// Partial match (target contains name or vice versa)
	if strings.Contains(targetLower, nameLower) || strings.Contains(nameLower, targetLower) {
		return true
	}

	// Context-specific matching
	switch context {
	case "kill":
		// Generic kill objectives match any enemy
		if targetLower == "enemy" || targetLower == "enemies" || targetLower == "monster" {
			return true
		}
	case "collect":
		// Generic collect objectives match any item
		if targetLower == "item" || targetLower == "items" {
			return true
		}
	case "ui":
		// GAP-014 REPAIR: UI objective matching (for tutorial)
		// Handle variations in objective naming
		if targetLower == "inventory" && nameLower == "inventory" {
			return true
		}
		if targetLower == "quest_log" && nameLower == "quest_log" {
			return true
		}
		if strings.Contains(targetLower, "character") && nameLower == "character" {
			return true
		}
		if strings.Contains(targetLower, "skill") && nameLower == "skills" {
			return true
		}
		if strings.Contains(targetLower, "map") && nameLower == "map" {
			return true
		}
	}

	return false
}

// tileKeyFromCoords creates a unique key for a tile position.
func tileKeyFromCoords(x, y int) string {
	// Use a simple string format for tile coordinates
	return fmt.Sprintf("%d,%d", x, y)
}

// AwardQuestRewards distributes rewards from a completed quest.
func (s *ObjectiveTrackerSystem) AwardQuestRewards(entity *Entity, qst *quest.Quest) {
	// Award XP
	if qst.Reward.XP > 0 {
		if expComp, ok := entity.GetComponent("experience"); ok {
			if exp, ok := expComp.(*ExperienceComponent); ok {
				exp.AddXP(qst.Reward.XP)
			}
		}
	}

	// Award gold
	if qst.Reward.Gold > 0 {
		if invComp, ok := entity.GetComponent("inventory"); ok {
			if inv, ok := invComp.(*InventoryComponent); ok {
				inv.Gold += qst.Reward.Gold
			}
		}
	}

	// Award skill points
	if qst.Reward.SkillPoints > 0 {
		if expComp, ok := entity.GetComponent("experience"); ok {
			if exp, ok := expComp.(*ExperienceComponent); ok {
				exp.SkillPoints += qst.Reward.SkillPoints
			}
		}
	}

	// Award items
	if len(qst.Reward.Items) > 0 {
		s.awardQuestItems(entity, qst)
	}
}

// awardQuestItems generates and awards items from quest rewards.
func (s *ObjectiveTrackerSystem) awardQuestItems(entity *Entity, qst *quest.Quest) {
	invComp, ok := entity.GetComponent("inventory")
	if !ok {
		return // No inventory to receive items
	}
	inv := invComp.(*InventoryComponent)

	// Get genre for item generation
	genreID := "fantasy" // Default genre
	if genreComp, hasGenre := entity.GetComponent("genre"); hasGenre {
		if genre, ok := genreComp.(*GenreComponent); ok {
			genreID = genre.GetPrimaryGenre()
		}
	}

	// Generate items based on quest reward item names
	for i, itemName := range qst.Reward.Items {
		// Create generation seed from quest seed and item index
		itemSeed := qst.Seed + int64(i)*100

		// Determine item type from name
		itemType := s.inferItemTypeFromName(itemName)

		// Set up generation parameters
		params := procgen.GenerationParams{
			Difficulty: float64(qst.Difficulty) / 5.0, // Convert difficulty to 0-1 range
			Depth:      qst.RequiredLevel,
			GenreID:    genreID,
			Custom: map[string]interface{}{
				"count": 1,
				"type":  itemType,
			},
		}

		// Generate item
		result, err := s.itemGenerator.Generate(itemSeed, params)
		if err != nil {
			continue // Skip items that fail to generate
		}

		items, ok := result.([]*item.Item)
		if !ok || len(items) == 0 {
			continue
		}

		// Add item to inventory
		generatedItem := items[0]
		if inv.CanAddItem(generatedItem) {
			inv.AddItem(generatedItem)
		}
	}
}

// inferItemTypeFromName attempts to determine item type from the item name string.
// Quest item names are descriptive (e.g., "healing potion", "iron sword", "leather armor").
func (s *ObjectiveTrackerSystem) inferItemTypeFromName(itemName string) string {
	nameLower := strings.ToLower(itemName)

	// Check for weapon keywords
	weaponKeywords := []string{"sword", "axe", "bow", "staff", "dagger", "mace", "spear", "hammer"}
	for _, keyword := range weaponKeywords {
		if strings.Contains(nameLower, keyword) {
			return "weapon"
		}
	}

	// Check for armor keywords
	armorKeywords := []string{"armor", "helm", "helmet", "boots", "gloves", "shield", "chest", "plate", "mail"}
	for _, keyword := range armorKeywords {
		if strings.Contains(nameLower, keyword) {
			return "armor"
		}
	}

	// Check for consumable keywords
	consumableKeywords := []string{"potion", "elixir", "scroll", "food", "drink", "medicine"}
	for _, keyword := range consumableKeywords {
		if strings.Contains(nameLower, keyword) {
			return "consumable"
		}
	}

	// Default to random item type
	itemTypes := []string{"weapon", "armor", "consumable"}
	return itemTypes[rand.Intn(len(itemTypes))]
}
