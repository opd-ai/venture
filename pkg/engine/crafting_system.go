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
	world         *World
	inventory     *InventorySystem
	itemGenerator *item.ItemGenerator
	logger        *logrus.Entry
}

// NewCraftingSystem creates a new crafting system.
func NewCraftingSystem(world *World, inventorySystem *InventorySystem, itemGen *item.ItemGenerator) *CraftingSystem {
	return NewCraftingSystemWithLogger(world, inventorySystem, itemGen, nil)
}

// NewCraftingSystemWithLogger creates a new crafting system with a logger.
func NewCraftingSystemWithLogger(world *World, inventorySystem *InventorySystem, itemGen *item.ItemGenerator, logger *logrus.Logger) *CraftingSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("system", "crafting")
	}

	return &CraftingSystem{
		world:         world,
		inventory:     inventorySystem,
		itemGenerator: itemGen,
		logger:        logEntry,
	}
}

// Update processes crafting progress for all entities with CraftingProgressComponent.
// Call this each game tick with deltaTime in seconds.
func (s *CraftingSystem) Update(entities []*Entity, deltaTime float64) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_count": len(entities),
			"delta_time":   deltaTime,
		}).Debug("crafting system update started")
	}

	for _, entity := range entities {
		comp, ok := entity.GetComponent("crafting_progress")
		if !ok {
			continue
		}

		progressComp, ok := comp.(*CraftingProgressComponent)
		if !ok {
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entity_id":      entity.ID,
					"component_type": "crafting_progress",
				}).Warn("crafting_progress component has wrong type")
			}
			continue
		}

		// Update elapsed time
		progressComp.ElapsedTimeSec += deltaTime

		// Check if crafting is complete
		if progressComp.IsComplete() {
			if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
				s.logger.WithFields(logrus.Fields{
					"entity_id":  entity.ID,
					"recipe_id":  progressComp.CurrentRecipe.ID,
					"elapsed":    progressComp.ElapsedTimeSec,
					"total_time": progressComp.RequiredTimeSec,
				}).Debug("crafting complete, processing result")
			}

			// Complete the craft
			s.completeCraft(entity.ID, progressComp)
			// Remove crafting progress component
			entity.RemoveComponent("crafting_progress")

			// Release crafting station if used
			if progressComp.UsingStationID != 0 {
				s.releaseStation(progressComp.UsingStationID)
			}
		}
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

	if result := s.validateRecipeKnowledge(entity, recipe); result != nil {
		return result, nil
	}

	skillLevel := s.getCraftingSkillLevel(entity)
	if result := s.validateSkillLevel(entityID, recipe, skillLevel); result != nil {
		return result, nil
	}

	invComp, err := s.getInventoryComponent(entity)
	if err != nil {
		s.logInventoryError(entityID, err)
		return nil, err
	}

	if result := s.validateCraftingMaterials(entityID, recipe, invComp); result != nil {
		return result, nil
	}

	if result := s.validateInventorySpace(entityID, recipe, invComp); result != nil {
		return result, nil
	}

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

	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		if s.logger != nil {
			s.logger.WithField("entity_id", entityID).Warn("entity not found when completing craft")
		}
		return
	}

	recipe := progressComp.CurrentRecipe
	if recipe == nil {
		if s.logger != nil {
			s.logger.WithField("entity_id", entityID).Warn("recipe is nil in progress component")
		}
		return
	}

	// Get skill level
	skillLevel := s.getCraftingSkillLevel(entity)

	// Get station bonus
	stationBonus := 0.0
	if progressComp.UsingStationID != 0 {
		if station, ok := s.world.GetEntity(progressComp.UsingStationID); ok {
			if stationComp, err := s.getCraftingStationComponent(station); err == nil {
				stationBonus = stationComp.BonusSuccessChance
			}
		}
	}

	// Calculate final success chance
	baseChance := recipe.GetEffectiveSuccessChance(skillLevel)
	finalChance := baseChance + stationBonus
	if finalChance > 0.95 {
		finalChance = 0.95 // Cap at 95%
	}

	// Roll for success (use recipe seed + entity ID for determinism)
	rng := rand.New(rand.NewSource(recipe.OutputItemSeed + int64(entityID)))
	success := rng.Float64() < finalChance

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"recipe_id":      recipe.ID,
			"success_chance": finalChance,
			"roll_succeeded": success,
		}).Debug("craft success roll performed")
	}

	// Calculate XP gained (more XP for higher rarity recipes)
	xpGained := 10 * (int(recipe.Rarity) + 1)
	if !success {
		xpGained = xpGained / 2 // Half XP on failure
	}

	// Add XP to crafting skill
	s.addCraftingExperience(entity, xpGained)

	if success {
		// Generate output item
		outputItem := s.generateOutputItem(recipe, rng)

		// Add to inventory
		invComp, err := s.getInventoryComponent(entity)
		if err == nil {
			if !invComp.AddItem(outputItem) {
				// Inventory full (shouldn't happen, we validated), drop item near entity
				if s.logger != nil {
					s.logger.WithFields(logrus.Fields{
						"entity_id": entityID,
						"item_name": outputItem.Name,
						"recipe_id": recipe.ID,
					}).Warn("inventory full after craft, item lost")
				}
			} else {
				if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
					s.logger.WithFields(logrus.Fields{
						"entity_id": entityID,
						"item_name": outputItem.Name,
						"recipe_id": recipe.ID,
					}).Debug("crafted item added to inventory")
				}
			}
		} else {
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entity_id": entityID,
					"error":     err.Error(),
				}).Error("failed to get inventory after craft completion")
			}
		}

		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.InfoLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entityID,
				"recipe_id":      recipe.ID,
				"item_name":      outputItem.Name,
				"xp_gained":      xpGained,
				"success_chance": finalChance,
			}).Info("crafting succeeded")
		}
	} else {
		// Failure - materials already consumed, no output
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.InfoLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entityID,
				"recipe_id":      recipe.ID,
				"xp_gained":      xpGained,
				"success_chance": finalChance,
			}).Info("crafting failed")
		}
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
	return &item.Item{
		Name:   recipe.Name,
		Type:   recipe.OutputItemType,
		Rarity: item.RarityCommon,
		Stats: item.Stats{
			Value: recipe.GoldCost,
		},
	}
}

// hasRecipeKnowledge checks if entity knows a recipe.
func (s *CraftingSystem) hasRecipeKnowledge(entity *Entity, recipe *Recipe) bool {
	comp, ok := entity.GetComponent("recipe_knowledge")
	if !ok {
		return false
	}
	knowledgeComp, ok := comp.(*RecipeKnowledgeComponent)
	if !ok {
		return false
	}
	return knowledgeComp.KnowsRecipe(recipe.ID)
}

// getCraftingSkillLevel gets entity's crafting skill level.
func (s *CraftingSystem) getCraftingSkillLevel(entity *Entity) int {
	comp, ok := entity.GetComponent("crafting_skill")
	if !ok {
		return 0 // No skill component = level 0
	}
	skillComp, ok := comp.(*CraftingSkillComponent)
	if !ok {
		return 0
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

	consumed := []string{}

	// Deduct gold
	invComp.Gold -= recipe.GoldCost
	if recipe.GoldCost > 0 {
		consumed = append(consumed, fmt.Sprintf("%d gold", recipe.GoldCost))
	}

	// Remove materials
	for _, req := range recipe.Materials {
		removed := 0
		for i := 0; i < len(invComp.Items) && removed < req.Quantity; i++ {
			itm := invComp.Items[i]
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

			// Remove this item
			invComp.Items[i] = nil
			removed++
			consumed = append(consumed, itm.Name)
		}

		// Verify we removed enough
		if removed < req.Quantity {
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entity_id": entityID,
					"recipe_id": recipe.ID,
					"material":  req.ItemName,
					"needed":    req.Quantity,
					"found":     removed,
				}).Error("insufficient materials during consumption")
			}
			return nil, fmt.Errorf("insufficient %s: needed %d, found %d", req.ItemName, req.Quantity, removed)
		}
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

// validateAndReserveStation checks if station is valid and marks it as in-use.
// Returns (bonusSuccessChance, craftTimeMultiplier, error).
func (s *CraftingSystem) validateAndReserveStation(stationID uint64, recipeType RecipeType) (float64, float64, error) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"station_id":  stationID,
			"recipe_type": recipeType.String(),
		}).Debug("validating and reserving station")
	}

	station, ok := s.world.GetEntity(stationID)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"station_id": stationID,
			}).Warn("station entity not found")
		}
		return 0, 1.0, fmt.Errorf("station not found")
	}

	stationComp, err := s.getCraftingStationComponent(station)
	if err != nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"station_id": stationID,
				"error":      err.Error(),
			}).Warn("failed to get crafting station component")
		}
		return 0, 1.0, err
	}

	// Check station type matches recipe
	if stationComp.StationType != recipeType {
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"station_id":   stationID,
				"station_type": stationComp.StationType.String(),
				"recipe_type":  recipeType.String(),
			}).Debug("station type mismatch")
		}
		return 0, 1.0, fmt.Errorf("wrong station type: need %s station", recipeType.String())
	}

	// Check availability
	if !stationComp.Available {
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"station_id": stationID,
			}).Debug("station already in use")
		}
		return 0, 1.0, fmt.Errorf("station in use")
	}

	// Reserve station
	stationComp.Available = false

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"station_id":      stationID,
			"bonus_chance":    stationComp.BonusSuccessChance,
			"time_multiplier": stationComp.CraftTimeMultiplier,
		}).Debug("station reserved successfully")
	}

	return stationComp.BonusSuccessChance, stationComp.CraftTimeMultiplier, nil
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
	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return nil, fmt.Errorf("entity has no inventory component")
	}
	invComp, ok := comp.(*InventoryComponent)
	if !ok {
		return nil, fmt.Errorf("inventory component has wrong type")
	}
	return invComp, nil
}

func (s *CraftingSystem) getCraftingStationComponent(entity *Entity) (*CraftingStationComponent, error) {
	comp, ok := entity.GetComponent("crafting_station")
	if !ok {
		return nil, fmt.Errorf("entity has no crafting_station component")
	}
	stationComp, ok := comp.(*CraftingStationComponent)
	if !ok {
		return nil, fmt.Errorf("crafting_station component has wrong type")
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
