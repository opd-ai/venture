// Package engine provides the crafting system for recipe-based item creation.
// This file implements CraftingSystem which handles recipe validation, material
// consumption, crafting progress tracking, and item generation. The system
// integrates with inventory, skills, and item generation systems.
//
// Design Philosophy:
// - Server-authoritative: all crafting must be validated server-side
// - Deterministic: same recipe + materials always produces same result (seed-based)
// - Progressive: crafting takes time, system updates progress each tick
// - Skill-based: success chance and available recipes scale with skill level
// - Risk/reward: failed crafts consume 50% of materials
package engine

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/sirupsen/logrus"
)

// StationBonusProvider provides crafting bonuses from player-owned housing stations.
// This interface allows CraftingSystem to integrate with housing_crafting package
// without creating a circular dependency.
type StationBonusProvider interface {
	// GetCraftingBonus returns the best crafting bonus multiplier for a player's recipe.
	// Returns 1.0 (no bonus) if player has no stations or no stations with the recipe.
	GetCraftingBonus(playerID, recipeID string) float64
}

// CraftingResult contains the outcome of a crafting attempt.
type CraftingResult struct {
	Success       bool
	Item          *item.Item // nil if failed
	RecipeName    string
	XPGained      int
	ErrorMessage  string
	MaterialsLost []string // Materials consumed (partial on failure)
}

// CraftingSystem manages recipe-based item crafting.
type CraftingSystem struct {
	world          *World
	inventory      *InventorySystem
	itemGenerator  *item.ItemGenerator
	logger         *logrus.Entry
	stationManager StationBonusProvider // Optional housing crafting integration
}

// NewCraftingSystem creates a new crafting system.
func NewCraftingSystem(world *World, inventorySystem *InventorySystem, itemGen *item.ItemGenerator) *CraftingSystem {
	return NewCraftingSystemWithLogger(world, inventorySystem, itemGen, nil)
}

// NewCraftingSystemWithLogger creates a new crafting system with a logger.
func NewCraftingSystemWithLogger(world *World, inventorySystem *InventorySystem, itemGen *item.ItemGenerator, logger *logrus.Logger) *CraftingSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logger.SetReportCaller(true)
		logEntry = logger.WithField("system", "crafting")
	}

	return &CraftingSystem{
		world:         world,
		inventory:     inventorySystem,
		itemGenerator: itemGen,
		logger:        logEntry,
	}
}

// SetStationManager configures the housing crafting station manager for bonus calculations.
// This enables auto-discovery of player-owned stations for crafting bonuses.
// Call this after system initialization to wire in the housing_crafting integration.
func (s *CraftingSystem) SetStationManager(manager StationBonusProvider) {
	s.stationManager = manager
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.InfoLevel {
		s.logger.Info("housing crafting station manager configured for auto-discovery")
	}
}

// SetItemGenerator configures the item generator used for crafting output.
// This allows deferred initialization when the generator is not available at construction time.
func (s *CraftingSystem) SetItemGenerator(gen *item.ItemGenerator) {
	s.itemGenerator = gen
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.InfoLevel {
		s.logger.Info("item generator configured for crafting system")
	}
}

// Update processes crafting progress for all entities with CraftingProgressComponent.
// Call this each game tick with deltaTime in seconds.
func (s *CraftingSystem) Update(entities []*Entity, deltaTime float64) {
	s.logUpdateStart(len(entities), deltaTime)

	for _, entity := range entities {
		s.updateCraftingProgress(entity, deltaTime)
	}

	s.logUpdateComplete(len(entities))
}

// logUpdateStart logs the start of crafting system update.
func (s *CraftingSystem) logUpdateStart(entityCount int, deltaTime float64) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_count": entityCount,
			"delta_time":   deltaTime,
		}).Debug("crafting system update started")
	}
}

// updateCraftingProgress updates crafting progress for a single entity.
func (s *CraftingSystem) updateCraftingProgress(entity *Entity, deltaTime float64) {
	progressComp, ok := s.getCraftingProgress(entity)
	if !ok {
		return
	}

	s.advanceProgress(entity, progressComp, deltaTime)

	if progressComp.IsComplete() {
		s.handleCraftingCompletion(entity, progressComp)
	}
}

// getCraftingProgress retrieves and validates the crafting progress component.
func (s *CraftingSystem) getCraftingProgress(entity *Entity) (*CraftingProgressComponent, bool) {
	comp, ok := entity.GetComponent("crafting_progress")
	if !ok {
		return nil, false
	}

	progressComp, ok := comp.(*CraftingProgressComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "crafting_progress",
			}).Warn("crafting_progress component has wrong type")
		}
		return nil, false
	}

	return progressComp, true
}

// advanceProgress updates the elapsed time for crafting progress.
func (s *CraftingSystem) advanceProgress(entity *Entity, progressComp *CraftingProgressComponent, deltaTime float64) {
	oldElapsed := progressComp.ElapsedTimeSec
	progressComp.ElapsedTimeSec += deltaTime

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":    entity.ID,
			"recipe_id":    progressComp.CurrentRecipe.ID,
			"old_elapsed":  oldElapsed,
			"new_elapsed":  progressComp.ElapsedTimeSec,
			"required":     progressComp.RequiredTimeSec,
			"progress_pct": (progressComp.ElapsedTimeSec / progressComp.RequiredTimeSec) * 100,
		}).Debug("updated crafting progress")
	}
}

// handleCraftingCompletion handles the completion of a crafting operation.
func (s *CraftingSystem) handleCraftingCompletion(entity *Entity, progressComp *CraftingProgressComponent) {
	s.logCraftingComplete(entity, progressComp)

	s.completeCraft(entity.ID, progressComp)
	entity.RemoveComponent("crafting_progress")

	s.logComponentRemoved(entity)

	if progressComp.UsingStationID != 0 {
		s.releaseStation(progressComp.UsingStationID)
	}
}

// logCraftingComplete logs when crafting is complete.
func (s *CraftingSystem) logCraftingComplete(entity *Entity, progressComp *CraftingProgressComponent) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  entity.ID,
			"recipe_id":  progressComp.CurrentRecipe.ID,
			"elapsed":    progressComp.ElapsedTimeSec,
			"total_time": progressComp.RequiredTimeSec,
		}).Debug("crafting complete, processing result")
	}
}

// logComponentRemoved logs when crafting progress component is removed.
func (s *CraftingSystem) logComponentRemoved(entity *Entity) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entity.ID,
			"component_type": "crafting_progress",
		}).Debug("removed crafting progress component")
	}
}

// logUpdateComplete logs the completion of crafting system update.
func (s *CraftingSystem) logUpdateComplete(entityCount int) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_count": entityCount,
		}).Debug("crafting system update complete")
	}
}

// StartCraft begins crafting a recipe for an entity.
// Validates recipe knowledge, materials, skill, and station requirements.
// Consumes materials and gold, creates CraftingProgressComponent.
func (s *CraftingSystem) StartCraft(entityID uint64, recipe *Recipe, stationID uint64) (*CraftingResult, error) {
	s.logCraftStart(entityID, recipe, stationID)

	entity, err := s.validateEntity(entityID)
	if err != nil {
		return nil, err
	}

	invComp, err := s.getInventoryComponent(entity)
	if err != nil {
		s.logInventoryError(entityID, err)
		return nil, err
	}

	if result := s.validateAllCraftingRequirements(entity, entityID, recipe, invComp); result != nil {
		return result, nil
	}

	skillLevel := s.getCraftingSkillLevel(entity)
	stationBonus, craftTimeMultiplier, err := s.processStationValidation(entityID, recipe, stationID)
	if err != nil {
		return &CraftingResult{
			Success:      false,
			RecipeName:   recipe.Name,
			ErrorMessage: err.Error(),
		}, nil
	}

	consumed, err := s.consumeMaterialsWithRollback(entity, recipe, stationID, invComp)
	if err != nil {
		return nil, err
	}

	s.createCraftingProgress(entity, recipe, stationID, craftTimeMultiplier, skillLevel, stationBonus, consumed)
	s.logCraftSuccess(entityID, recipe, stationID, skillLevel, consumed)

	return &CraftingResult{
		Success:       true,
		RecipeName:    recipe.Name,
		MaterialsLost: consumed,
	}, nil
}

// completeCraft finishes a craft attempt, rolling for success and generating output.
func (s *CraftingSystem) completeCraft(entityID uint64, progressComp *CraftingProgressComponent) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
		}).Debug("completing craft attempt")
	}

	entity, recipe := s.validateCraftCompletion(entityID, progressComp)
	if entity == nil || recipe == nil {
		return
	}

	skillLevel := s.getCraftingSkillLevel(entity)
	stationBonus := s.extractStationBonus(entityID, recipe.ID, progressComp.UsingStationID)
	finalChance := s.calculateFinalSuccessChance(recipe, skillLevel, stationBonus)

	rng := rand.New(rand.NewSource(recipe.OutputItemSeed + int64(entityID)))
	success := s.rollForCraftSuccess(entityID, recipe.ID, rng, finalChance)

	xpGained := s.calculateXPGained(recipe.Rarity, success)
	s.addCraftingExperience(entity, xpGained)

	if success {
		s.handleCraftSuccess(entityID, entity, recipe, rng, xpGained, finalChance)
	} else {
		s.handleCraftFailure(entityID, recipe.ID, xpGained, finalChance)
	}
}

// validateCraftCompletion validates entity and recipe exist for craft completion.
func (s *CraftingSystem) validateCraftCompletion(entityID uint64, progressComp *CraftingProgressComponent) (*Entity, *Recipe) {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		if s.logger != nil {
			s.logger.WithField("entity_id", entityID).Warn("entity not found when completing craft")
		}
		return nil, nil
	}

	recipe := progressComp.CurrentRecipe
	if recipe == nil {
		if s.logger != nil {
			s.logger.WithField("entity_id", entityID).Warn("recipe is nil in progress component")
		}
		return nil, nil
	}

	return entity, recipe
}

// extractStationBonus retrieves the success bonus from crafting stations.
// First tries auto-discovery via StationManager (player-owned housing stations),
// then falls back to explicit station entity lookup (placed crafting stations).
// This enables automatic bonus calculation without manual station registration.
func (s *CraftingSystem) extractStationBonus(entityID uint64, recipeID string, stationID uint64) float64 {
	// Try auto-discovery via housing crafting stations first
	if bonus := s.tryHousingStationBonus(entityID, recipeID); bonus > 0 {
		return bonus
	}
	// Fallback to explicit station lookup
	return s.tryExplicitStationBonus(stationID)
}

// tryHousingStationBonus attempts to get bonus from housing crafting stations.
func (s *CraftingSystem) tryHousingStationBonus(entityID uint64, recipeID string) float64 {
	if s.stationManager == nil {
		return 0.0
	}
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return 0.0
	}
	playerID := s.extractPlayerID(entity)
	if playerID == "" {
		return 0.0
	}
	bonus := s.stationManager.GetCraftingBonus(playerID, recipeID)
	if bonus > 1.0 {
		s.logHousingBonus(entityID, playerID, recipeID, bonus)
		return bonus - 1.0
	}
	return 0.0
}

// extractPlayerID extracts player ID string from entity's network component.
func (s *CraftingSystem) extractPlayerID(entity *Entity) string {
	networkComp, ok := entity.GetComponent("network")
	if !ok {
		return ""
	}
	nc, ok := networkComp.(*NetworkComponent)
	if !ok || nc.PlayerID == 0 {
		return ""
	}
	return fmt.Sprintf("%d", nc.PlayerID)
}

// logHousingBonus logs housing station bonus discovery at debug level.
func (s *CraftingSystem) logHousingBonus(entityID uint64, playerID, recipeID string, bonus float64) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
			"player_id": playerID,
			"recipe_id": recipeID,
			"bonus":     bonus,
			"source":    "housing_stations",
		}).Debug("housing station bonus auto-discovered")
	}
}

// tryExplicitStationBonus attempts to get bonus from explicit station entity.
func (s *CraftingSystem) tryExplicitStationBonus(stationID uint64) float64 {
	if stationID == 0 {
		s.logDebug("no station specified and no housing bonus, bonus is 0")
		return 0.0
	}
	s.logDebugWithField("station_id", stationID, "extracting station bonus from entity")
	station, ok := s.world.GetEntity(stationID)
	if !ok {
		s.logDebugWithField("station_id", stationID, "station not found, bonus is 0")
		return 0.0
	}
	stationComp, err := s.getCraftingStationComponent(station)
	if err != nil {
		s.logStationError(stationID, err)
		return 0.0
	}
	s.logStationBonus(stationID, stationComp.BonusSuccessChance)
	return stationComp.BonusSuccessChance
}

// logDebug logs a debug message if logger is available and at debug level.
func (s *CraftingSystem) logDebug(msg string) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.Debug(msg)
	}
}

// logDebugWithField logs a debug message with a field if logger is available.
func (s *CraftingSystem) logDebugWithField(key string, value interface{}, msg string) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithField(key, value).Debug(msg)
	}
}

// logStationError logs station component extraction error.
func (s *CraftingSystem) logStationError(stationID uint64, err error) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"station_id": stationID,
			"error":      err.Error(),
		}).Debug("failed to get station component, bonus is 0")
	}
}

// logStationBonus logs successful station bonus extraction.
func (s *CraftingSystem) logStationBonus(stationID uint64, bonus float64) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"station_id": stationID,
			"bonus":      bonus,
			"source":     "station_entity",
		}).Debug("station bonus extracted from entity")
	}
}

// calculateFinalSuccessChance computes the total success chance with all bonuses applied.
func (s *CraftingSystem) calculateFinalSuccessChance(recipe *Recipe, skillLevel int, stationBonus float64) float64 {
	baseChance := recipe.GetEffectiveSuccessChance(skillLevel)
	finalChance := baseChance + stationBonus
	if finalChance > 0.95 {
		finalChance = 0.95
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"recipe_id":       recipe.ID,
			"skill_level":     skillLevel,
			"base_chance":     baseChance,
			"station_bonus":   stationBonus,
			"final_chance":    finalChance,
			"capped_at_95pct": finalChance == 0.95 && (baseChance+stationBonus) > 0.95,
		}).Debug("calculated final success chance")
	}

	return finalChance
}

// rollForCraftSuccess performs the deterministic success roll for crafting.
func (s *CraftingSystem) rollForCraftSuccess(entityID uint64, recipeID string, rng *rand.Rand, successChance float64) bool {
	success := rng.Float64() < successChance

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"recipe_id":      recipeID,
			"success_chance": successChance,
			"roll_succeeded": success,
		}).Debug("craft success roll performed")
	}

	return success
}

// calculateXPGained determines experience points gained from a crafting attempt.
func (s *CraftingSystem) calculateXPGained(rarity RecipeRarity, success bool) int {
	xpGained := 10 * (int(rarity) + 1)
	if !success {
		xpGained = xpGained / 2
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"rarity":    rarity.String(),
			"success":   success,
			"xp_gained": xpGained,
		}).Debug("calculated crafting XP")
	}

	return xpGained
}

// handleCraftSuccess processes a successful crafting attempt.
func (s *CraftingSystem) handleCraftSuccess(entityID uint64, entity *Entity, recipe *Recipe, rng *rand.Rand, xpGained int, successChance float64) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"recipe_id":      recipe.ID,
			"xp_gained":      xpGained,
			"success_chance": successChance,
		}).Debug("handling successful craft")
	}

	outputItem := s.generateOutputItem(recipe, rng)

	if !s.addItemToInventory(entityID, entity, outputItem, recipe.ID) {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entityID,
				"recipe_id": recipe.ID,
			}).Warn("craft succeeded but item could not be added to inventory")
		}
		return
	}

	s.logSuccessfulCraft(entityID, recipe.ID, outputItem.Name, xpGained, successChance)
}

// addItemToInventory attempts to add a crafted item to the entity's inventory.
func (s *CraftingSystem) addItemToInventory(entityID uint64, entity *Entity, outputItem *item.Item, recipeID string) bool {
	invComp, err := s.getInventoryComponent(entity)
	if err != nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entityID,
				"error":     err.Error(),
			}).Error("failed to get inventory after craft completion")
		}
		return false
	}

	if !invComp.AddItem(outputItem) {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entityID,
				"item_name": outputItem.Name,
				"recipe_id": recipeID,
			}).Warn("inventory full after craft, item lost")
		}
		return false
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
			"item_name": outputItem.Name,
			"recipe_id": recipeID,
		}).Debug("crafted item added to inventory")
	}

	return true
}

// logSuccessfulCraft logs information about a successful crafting operation.
func (s *CraftingSystem) logSuccessfulCraft(entityID uint64, recipeID, itemName string, xpGained int, successChance float64) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.InfoLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"recipe_id":      recipeID,
			"item_name":      itemName,
			"xp_gained":      xpGained,
			"success_chance": successChance,
		}).Info("crafting succeeded")
	}
}

// handleCraftFailure processes a failed crafting attempt.
func (s *CraftingSystem) handleCraftFailure(entityID uint64, recipeID string, xpGained int, successChance float64) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.InfoLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"recipe_id":      recipeID,
			"xp_gained":      xpGained,
			"success_chance": successChance,
		}).Info("crafting failed")
	}
}

// generateOutputItem creates the crafted item using the item generator.
func (s *CraftingSystem) generateOutputItem(recipe *Recipe, rng *rand.Rand) *item.Item {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"recipe_id":   recipe.ID,
			"item_type":   recipe.OutputItemType.String(),
			"genre_id":    recipe.GenreID,
			"skill_level": recipe.SkillRequired,
		}).Debug("generating crafted item")
	}

	// Build generation parameters
	params := procgen.GenerationParams{
		Difficulty: 0.5,                  // Medium difficulty for crafted items
		Depth:      recipe.SkillRequired, // Depth scales with skill requirement
		GenreID:    recipe.GenreID,
		Custom: map[string]interface{}{
			"type":  recipe.OutputItemType.String(),
			"count": 1,
		},
	}

	// Generate item using recipe's output seed
	if s.itemGenerator == nil {
		if s.logger != nil {
			s.logger.Warn("item generator not configured, using fallback for crafted item")
		}
		return s.createFallbackItem(recipe)
	}
	result, err := s.itemGenerator.Generate(recipe.OutputItemSeed, params)
	if err != nil {
		// Fallback: create basic item
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"recipe_id": recipe.ID,
				"error":     err.Error(),
			}).Warn("failed to generate crafted item, using fallback")
		}
		return s.createFallbackItem(recipe)
	}

	items := result.([]*item.Item)
	if len(items) == 0 {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"recipe_id": recipe.ID,
			}).Warn("item generation returned empty result, using fallback")
		}
		return s.createFallbackItem(recipe)
	}

	// Return first generated item
	return items[0]
}

// createFallbackItem creates a basic item when generation fails.
func (s *CraftingSystem) createFallbackItem(recipe *Recipe) *item.Item {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"recipe_id":   recipe.ID,
			"recipe_name": recipe.Name,
			"item_type":   recipe.OutputItemType.String(),
		}).Debug("creating fallback item")
	}

	fallback := &item.Item{
		Name:   recipe.Name,
		Type:   recipe.OutputItemType,
		Rarity: item.RarityCommon,
		Stats: item.Stats{
			Value: recipe.GoldCost,
		},
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"recipe_id": recipe.ID,
			"item_name": fallback.Name,
		}).Debug("fallback item created")
	}

	return fallback
}

// hasRecipeKnowledge checks if entity knows a recipe.
func (s *CraftingSystem) hasRecipeKnowledge(entity *Entity, recipe *Recipe) bool {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"recipe_id":   recipe.ID,
			"recipe_name": recipe.Name,
		}).Debug("checking recipe knowledge")
	}

	comp, ok := entity.GetComponent("recipe_knowledge")
	if !ok {
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
			}).Debug("entity has no recipe knowledge component")
		}
		return false
	}
	knowledgeComp, ok := comp.(*RecipeKnowledgeComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "recipe_knowledge",
			}).Warn("recipe_knowledge component has wrong type")
		}
		return false
	}

	knows := knowledgeComp.KnowsRecipe(recipe.ID)
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"recipe_id": recipe.ID,
			"knows":     knows,
		}).Debug("recipe knowledge check complete")
	}

	return knows
}

// getCraftingSkillLevel gets entity's crafting skill level.
func (s *CraftingSystem) getCraftingSkillLevel(entity *Entity) int {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
		}).Debug("getting crafting skill level")
	}

	comp, ok := entity.GetComponent("crafting_skill")
	if !ok {
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
			}).Debug("entity has no crafting skill component, returning level 0")
		}
		return 0 // No skill component = level 0
	}
	skillComp, ok := comp.(*CraftingSkillComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "crafting_skill",
			}).Warn("crafting_skill component has wrong type, returning level 0")
		}
		return 0
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"skill_level": skillComp.SkillLevel,
		}).Debug("crafting skill level retrieved")
	}

	return skillComp.SkillLevel
}

// addCraftingExperience adds XP to entity's crafting skill.
func (s *CraftingSystem) addCraftingExperience(entity *Entity, xp int) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"xp":        xp,
		}).Debug("adding crafting experience")
	}

	comp, ok := entity.GetComponent("crafting_skill")
	if !ok {
		// Create skill component if not present
		skillComp := NewCraftingSkillComponent()
		entity.AddComponent(skillComp)
		comp = skillComp

		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "crafting_skill",
			}).Debug("created crafting skill component")
		}
	}

	skillComp, ok := comp.(*CraftingSkillComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "crafting_skill",
			}).Warn("crafting_skill component has wrong type")
		}
		return
	}

	oldLevel := skillComp.SkillLevel
	leveledUp := skillComp.AddExperience(xp)
	if leveledUp && s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":       entity.ID,
			"old_skill_level": oldLevel,
			"new_skill_level": skillComp.SkillLevel,
		}).Info("crafting skill leveled up")
	}
}

// validateMaterials checks if inventory contains required materials and gold.
// Returns (canCraft, missingDescription).
func (s *CraftingSystem) validateMaterials(invComp *InventoryComponent, recipe *Recipe) (bool, string) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"recipe_id": recipe.ID,
			"gold_cost": recipe.GoldCost,
			"gold_have": invComp.Gold,
			"materials": len(recipe.Materials),
		}).Debug("validating materials for crafting")
	}

	// Check gold
	if invComp.Gold < recipe.GoldCost {
		return false, fmt.Sprintf("%d gold", recipe.GoldCost-invComp.Gold)
	}

	// Check each material
	for _, req := range recipe.Materials {
		count := s.countMaterialInInventory(invComp, req)
		if count < req.Quantity {
			return false, fmt.Sprintf("%dx %s", req.Quantity-count, req.ItemName)
		}
	}

	return true, ""
}

// countMaterialInInventory counts how many of a material are in inventory.
func (s *CraftingSystem) countMaterialInInventory(invComp *InventoryComponent, req MaterialRequirement) int {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"item_name": req.ItemName,
			"quantity":  req.Quantity,
		}).Debug("counting material in inventory")
	}

	count := 0
	for _, itm := range invComp.Items {
		if itm == nil {
			continue
		}
		// Check name match
		if itm.Name != req.ItemName {
			continue
		}
		// Check type match if specified
		if req.ItemType != nil && itm.Type != *req.ItemType {
			continue
		}
		count++
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"item_name": req.ItemName,
			"count":     count,
		}).Debug("material count complete")
	}

	return count
}

// consumeMaterials removes materials and gold from inventory.
// Returns list of consumed material names.
func (s *CraftingSystem) consumeMaterials(entityID uint64, invComp *InventoryComponent, recipe *Recipe) ([]string, error) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
			"recipe_id": recipe.ID,
			"gold_cost": recipe.GoldCost,
			"materials": len(recipe.Materials),
		}).Debug("consuming materials for crafting")
	}

	consumed := s.deductGoldCost(invComp, recipe.GoldCost)

	if err := s.removeMaterialItems(entityID, invComp, recipe, &consumed); err != nil {
		return nil, err
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"recipe_id":      recipe.ID,
			"items_consumed": len(consumed),
		}).Debug("materials consumed successfully")
	}

	return consumed, nil
}

// deductGoldCost removes gold from inventory and returns consumption list.
func (s *CraftingSystem) deductGoldCost(invComp *InventoryComponent, goldCost int) []string {
	invComp.Gold -= goldCost
	consumed := []string{}
	if goldCost > 0 {
		consumed = append(consumed, fmt.Sprintf("%d gold", goldCost))
	}
	return consumed
}

// removeMaterialItems removes required materials from inventory.
func (s *CraftingSystem) removeMaterialItems(entityID uint64, invComp *InventoryComponent, recipe *Recipe, consumed *[]string) error {
	for _, req := range recipe.Materials {
		removed := s.removeMatchingItems(invComp, req, consumed)
		if err := s.validateMaterialRemoval(entityID, recipe, req, removed); err != nil {
			return err
		}
	}
	return nil
}

// removeMatchingItems finds and removes matching items from inventory.
func (s *CraftingSystem) removeMatchingItems(invComp *InventoryComponent, req MaterialRequirement, consumed *[]string) int {
	removed := 0
	for i := 0; i < len(invComp.Items) && removed < req.Quantity; i++ {
		itm := invComp.Items[i]
		if itm == nil || itm.Name != req.ItemName {
			continue
		}
		if req.ItemType != nil && itm.Type != *req.ItemType {
			continue
		}

		invComp.Items[i] = nil
		removed++
		*consumed = append(*consumed, itm.Name)
	}
	return removed
}

// validateMaterialRemoval ensures sufficient materials were removed.
func (s *CraftingSystem) validateMaterialRemoval(entityID uint64, recipe *Recipe, req MaterialRequirement, removed int) error {
	if removed >= req.Quantity {
		return nil
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
			"recipe_id": recipe.ID,
			"material":  req.ItemName,
			"needed":    req.Quantity,
			"found":     removed,
		}).Error("insufficient materials during consumption")
	}
	return fmt.Errorf("insufficient %s: needed %d, found %d", req.ItemName, req.Quantity, removed)
}

// validateAndReserveStation checks if station is valid and marks it as in-use.
// Returns (bonusSuccessChance, craftTimeMultiplier, error).
func (s *CraftingSystem) validateAndReserveStation(stationID uint64, recipeType RecipeType) (float64, float64, error) {
	s.logValidationStart(stationID, recipeType)

	station, err := s.getStationEntity(stationID)
	if err != nil {
		return 0, 1.0, err
	}

	stationComp, err := s.validateStationComponent(station, stationID)
	if err != nil {
		return 0, 1.0, err
	}

	if err := s.validateStationType(stationComp, recipeType, stationID); err != nil {
		return 0, 1.0, err
	}

	if err := s.validateStationAvailability(stationComp, stationID); err != nil {
		return 0, 1.0, err
	}

	s.reserveStation(stationComp, stationID)
	return stationComp.BonusSuccessChance, stationComp.CraftTimeMultiplier, nil
}

// logValidationStart logs the beginning of station validation.
func (s *CraftingSystem) logValidationStart(stationID uint64, recipeType RecipeType) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"station_id":  stationID,
			"recipe_type": recipeType.String(),
		}).Debug("validating and reserving station")
	}
}

// getStationEntity retrieves the station entity by ID.
func (s *CraftingSystem) getStationEntity(stationID uint64) (*Entity, error) {
	station, ok := s.world.GetEntity(stationID)
	if !ok {
		s.logStationNotFound(stationID)
		return nil, fmt.Errorf("station not found")
	}
	return station, nil
}

// logStationNotFound logs when a station entity is not found.
func (s *CraftingSystem) logStationNotFound(stationID uint64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"station_id": stationID,
		}).Warn("station entity not found")
	}
}

// validateStationComponent retrieves and validates the crafting station component.
func (s *CraftingSystem) validateStationComponent(station *Entity, stationID uint64) (*CraftingStationComponent, error) {
	stationComp, err := s.getCraftingStationComponent(station)
	if err != nil {
		s.logComponentError(stationID, err)
		return nil, err
	}
	return stationComp, nil
}

// logComponentError logs failures in getting the crafting station component.
func (s *CraftingSystem) logComponentError(stationID uint64, err error) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"station_id": stationID,
			"error":      err.Error(),
		}).Warn("failed to get crafting station component")
	}
}

// validateStationType checks if the station type matches the recipe type.
func (s *CraftingSystem) validateStationType(stationComp *CraftingStationComponent, recipeType RecipeType, stationID uint64) error {
	if stationComp.StationType != recipeType {
		s.logTypeMismatch(stationID, stationComp.StationType, recipeType)
		return fmt.Errorf("wrong station type: need %s station", recipeType.String())
	}
	return nil
}

// logTypeMismatch logs when station type doesn't match recipe requirements.
func (s *CraftingSystem) logTypeMismatch(stationID uint64, stationType, recipeType RecipeType) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"station_id":   stationID,
			"station_type": stationType.String(),
			"recipe_type":  recipeType.String(),
		}).Debug("station type mismatch")
	}
}

// validateStationAvailability checks if the station is available for use.
func (s *CraftingSystem) validateStationAvailability(stationComp *CraftingStationComponent, stationID uint64) error {
	if !stationComp.Available {
		s.logStationInUse(stationID)
		return fmt.Errorf("station in use")
	}
	return nil
}

// logStationInUse logs when a station is already in use.
func (s *CraftingSystem) logStationInUse(stationID uint64) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"station_id": stationID,
		}).Debug("station already in use")
	}
}

// reserveStation marks the station as reserved and logs the operation.
func (s *CraftingSystem) reserveStation(stationComp *CraftingStationComponent, stationID uint64) {
	stationComp.Available = false

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"station_id":      stationID,
			"bonus_chance":    stationComp.BonusSuccessChance,
			"time_multiplier": stationComp.CraftTimeMultiplier,
		}).Debug("station reserved successfully")
	}
}

// releaseStation marks a station as available.
func (s *CraftingSystem) releaseStation(stationID uint64) {
	if stationID == 0 {
		return
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"station_id": stationID,
		}).Debug("releasing station")
	}

	station, ok := s.world.GetEntity(stationID)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"station_id": stationID,
			}).Warn("station entity not found when releasing")
		}
		return
	}

	stationComp, err := s.getCraftingStationComponent(station)
	if err != nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"station_id": stationID,
				"error":      err.Error(),
			}).Warn("failed to get station component when releasing")
		}
		return
	}

	stationComp.Available = true

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"station_id": stationID,
		}).Debug("station released successfully")
	}
}

// Helper methods for component access

func (s *CraftingSystem) getInventoryComponent(entity *Entity) (*InventoryComponent, error) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
		}).Debug("getting inventory component")
	}

	comp, ok := entity.GetComponent("inventory")
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "inventory",
			}).Warn("entity has no inventory component")
		}
		return nil, fmt.Errorf("entity has no inventory component")
	}
	invComp, ok := comp.(*InventoryComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "inventory",
			}).Error("inventory component has wrong type")
		}
		return nil, fmt.Errorf("inventory component has wrong type")
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"gold":      invComp.Gold,
			"items":     len(invComp.Items),
		}).Debug("inventory component retrieved")
	}

	return invComp, nil
}

func (s *CraftingSystem) getCraftingStationComponent(entity *Entity) (*CraftingStationComponent, error) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
		}).Debug("getting crafting station component")
	}

	comp, ok := entity.GetComponent("crafting_station")
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "crafting_station",
			}).Warn("entity has no crafting_station component")
		}
		return nil, fmt.Errorf("entity has no crafting_station component")
	}
	stationComp, ok := comp.(*CraftingStationComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "crafting_station",
			}).Error("crafting_station component has wrong type")
		}
		return nil, fmt.Errorf("crafting_station component has wrong type")
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":     entity.ID,
			"station_type":  stationComp.StationType.String(),
			"available":     stationComp.Available,
			"bonus_chance":  stationComp.BonusSuccessChance,
			"time_modifier": stationComp.CraftTimeMultiplier,
		}).Debug("crafting station component retrieved")
	}

	return stationComp, nil
}

// logCraftStart logs the start of a crafting request.
func (s *CraftingSystem) logCraftStart(entityID uint64, recipe *Recipe, stationID uint64) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entityID,
			"recipe_id":   recipe.ID,
			"recipe_name": recipe.Name,
			"station_id":  stationID,
		}).Debug("start craft requested")
	}
}

// validateEntity retrieves and validates the entity exists in the world.
func (s *CraftingSystem) validateEntity(entityID uint64) (*Entity, error) {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entityID,
			}).Error("entity not found for crafting")
		}
		return nil, fmt.Errorf("entity %d not found", entityID)
	}
	return entity, nil
}

// validateRecipeKnowledge checks if the entity knows the recipe.
func (s *CraftingSystem) validateRecipeKnowledge(entity *Entity, recipe *Recipe) *CraftingResult {
	if !s.hasRecipeKnowledge(entity, recipe) {
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id":   entity.ID,
				"recipe_id":   recipe.ID,
				"recipe_name": recipe.Name,
			}).Debug("recipe unknown to entity")
		}
		return &CraftingResult{
			Success:      false,
			RecipeName:   recipe.Name,
			ErrorMessage: "Recipe unknown",
		}
	}
	return nil
}

// validateSkillLevel checks if the entity has sufficient crafting skill.
func (s *CraftingSystem) validateSkillLevel(entityID uint64, recipe *Recipe, skillLevel int) *CraftingResult {
	if skillLevel < recipe.SkillRequired {
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entityID,
				"recipe_id":      recipe.ID,
				"skill_level":    skillLevel,
				"required_level": recipe.SkillRequired,
			}).Debug("insufficient crafting skill")
		}
		return &CraftingResult{
			Success:      false,
			RecipeName:   recipe.Name,
			ErrorMessage: fmt.Sprintf("Requires crafting skill %d", recipe.SkillRequired),
		}
	}
	return nil
}

// logInventoryError logs an error when inventory component cannot be retrieved.
func (s *CraftingSystem) logInventoryError(entityID uint64, err error) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
			"error":     err.Error(),
		}).Error("failed to get inventory component")
	}
}

// validateCraftingMaterials checks if the entity has required materials and gold.
func (s *CraftingSystem) validateCraftingMaterials(entityID uint64, recipe *Recipe, invComp *InventoryComponent) *CraftingResult {
	canCraft, missing := s.validateMaterials(invComp, recipe)
	if !canCraft {
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entityID,
				"recipe_id": recipe.ID,
				"missing":   missing,
			}).Debug("insufficient materials for crafting")
		}
		return &CraftingResult{
			Success:      false,
			RecipeName:   recipe.Name,
			ErrorMessage: fmt.Sprintf("Missing: %s", missing),
		}
	}
	return nil
}

// validateInventorySpace checks if the inventory has space for the crafted item.
func (s *CraftingSystem) validateInventorySpace(entityID uint64, recipe *Recipe, invComp *InventoryComponent) *CraftingResult {
	if invComp.IsFull() {
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entityID,
				"recipe_id": recipe.ID,
			}).Debug("inventory full, cannot craft")
		}
		return &CraftingResult{
			Success:      false,
			RecipeName:   recipe.Name,
			ErrorMessage: "Inventory full",
		}
	}
	return nil
}

// processStationValidation validates and reserves the crafting station if specified.
func (s *CraftingSystem) processStationValidation(entityID uint64, recipe *Recipe, stationID uint64) (float64, float64, error) {
	stationBonus := 0.0
	craftTimeMultiplier := 1.0

	if stationID != 0 {
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id":   entityID,
				"station_id":  stationID,
				"recipe_type": recipe.Type.String(),
			}).Debug("validating crafting station")
		}

		bonus, timeMultiplier, err := s.validateAndReserveStation(stationID, recipe.Type)
		if err != nil {
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entity_id":  entityID,
					"station_id": stationID,
					"error":      err.Error(),
				}).Warn("station validation failed")
			}
			return 0, 1.0, err
		}
		stationBonus = bonus
		craftTimeMultiplier = timeMultiplier
	}

	return stationBonus, craftTimeMultiplier, nil
}

// consumeMaterialsWithRollback consumes materials and handles station rollback on failure.
func (s *CraftingSystem) consumeMaterialsWithRollback(entity *Entity, recipe *Recipe, stationID uint64, invComp *InventoryComponent) ([]string, error) {
	consumed, err := s.consumeMaterials(entity.ID, invComp, recipe)
	if err != nil {
		if stationID != 0 {
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entity_id":  entity.ID,
					"station_id": stationID,
				}).Debug("rolling back station reservation due to material consumption failure")
			}
			s.releaseStation(stationID)
		}
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"recipe_id": recipe.ID,
				"error":     err.Error(),
			}).Error("failed to consume materials")
		}
		return nil, fmt.Errorf("failed to consume materials: %w", err)
	}
	return consumed, nil
}

// createCraftingProgress creates and attaches the crafting progress component.
func (s *CraftingSystem) createCraftingProgress(entity *Entity, recipe *Recipe, stationID uint64, craftTimeMultiplier float64, skillLevel int, stationBonus float64, consumed []string) {
	actualCraftTime := recipe.CraftTimeSec * craftTimeMultiplier
	progressComp := NewCraftingProgressComponent(recipe, actualCraftTime, stationID)
	progressComp.MaterialsConsumed = true
	entity.AddComponent(progressComp)

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entity.ID,
			"recipe_id":      recipe.ID,
			"recipe_name":    recipe.Name,
			"skill_level":    skillLevel,
			"station_bonus":  stationBonus,
			"craft_time":     actualCraftTime,
			"materials_used": len(consumed),
		}).Debug("started crafting")
	}
}

// validateAllCraftingRequirements performs all crafting prerequisite validations.
// Returns CraftingResult if validation fails, nil if all validations pass.
func (s *CraftingSystem) validateAllCraftingRequirements(entity *Entity, entityID uint64, recipe *Recipe, invComp *InventoryComponent) *CraftingResult {
	if result := s.validateRecipeKnowledge(entity, recipe); result != nil {
		return result
	}

	skillLevel := s.getCraftingSkillLevel(entity)
	if result := s.validateSkillLevel(entityID, recipe, skillLevel); result != nil {
		return result
	}

	if result := s.validateCraftingMaterials(entityID, recipe, invComp); result != nil {
		return result
	}

	if result := s.validateInventorySpace(entityID, recipe, invComp); result != nil {
		return result
	}

	return nil
}

// logCraftSuccess logs successful craft initiation.
func (s *CraftingSystem) logCraftSuccess(entityID uint64, recipe *Recipe, stationID uint64, skillLevel int, consumed []string) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.InfoLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"recipe_id":      recipe.ID,
			"recipe_name":    recipe.Name,
			"skill_level":    skillLevel,
			"station_id":     stationID,
			"materials_used": len(consumed),
		}).Info("craft started successfully")
	}
}
