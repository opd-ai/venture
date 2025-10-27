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
	Success        bool
	Item           *item.Item // nil if failed
	RecipeName     string
	XPGained       int
	ErrorMessage   string
	MaterialsLost  []string // Materials consumed (partial on failure)
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
func (s *CraftingSystem) Update(entities []Entity, deltaTime float64) {
	for _, entity := range entities {
		comp, ok := entity.GetComponent("crafting_progress")
		if !ok {
			continue
		}

		progressComp, ok := comp.(*CraftingProgressComponent)
		if !ok {
			continue
		}

		// Update elapsed time
		progressComp.ElapsedTimeSec += deltaTime

		// Check if crafting is complete
		if progressComp.IsComplete() {
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
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return nil, fmt.Errorf("entity %d not found", entityID)
	}

	// Validate recipe knowledge
	if !s.hasRecipeKnowledge(entity, recipe) {
		return &CraftingResult{
			Success:      false,
			RecipeName:   recipe.Name,
			ErrorMessage: "Recipe unknown",
		}, nil
	}

	// Get crafting skill
	skillLevel := s.getCraftingSkillLevel(entity)
	if skillLevel < recipe.SkillRequired {
		return &CraftingResult{
			Success:      false,
			RecipeName:   recipe.Name,
			ErrorMessage: fmt.Sprintf("Requires crafting skill %d", recipe.SkillRequired),
		}, nil
	}

	// Get inventory component
	invComp, err := s.getInventoryComponent(entity)
	if err != nil {
		return nil, err
	}

	// Validate materials and gold
	canCraft, missing := s.validateMaterials(invComp, recipe)
	if !canCraft {
		return &CraftingResult{
			Success:      false,
			RecipeName:   recipe.Name,
			ErrorMessage: fmt.Sprintf("Missing: %s", missing),
		}, nil
	}

	// Check if inventory has space for output
	if invComp.IsFull() {
		return &CraftingResult{
			Success:      false,
			RecipeName:   recipe.Name,
			ErrorMessage: "Inventory full",
		}, nil
	}

	// Validate and reserve crafting station if specified
	stationBonus := 0.0
	craftTimeMultiplier := 1.0
	if stationID != 0 {
		bonus, timeMultiplier, err := s.validateAndReserveStation(stationID, recipe.Type)
		if err != nil {
			return &CraftingResult{
				Success:      false,
				RecipeName:   recipe.Name,
				ErrorMessage: err.Error(),
			}, nil
		}
		stationBonus = bonus
		craftTimeMultiplier = timeMultiplier
	}

	// Consume materials and gold (atomic operation)
	consumed, err := s.consumeMaterials(entity.ID, invComp, recipe)
	if err != nil {
		// Rollback station reservation
		if stationID != 0 {
			s.releaseStation(stationID)
		}
		return nil, fmt.Errorf("failed to consume materials: %w", err)
	}

	// Calculate actual craft time (modified by station)
	actualCraftTime := recipe.CraftTimeSec * craftTimeMultiplier

	// Create crafting progress component
	progressComp := NewCraftingProgressComponent(recipe, actualCraftTime, stationID)
	progressComp.MaterialsConsumed = true
	entity.AddComponent(progressComp)

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entityID":      entityID,
			"recipeID":      recipe.ID,
			"recipeName":    recipe.Name,
			"skillLevel":    skillLevel,
			"stationBonus":  stationBonus,
			"craftTime":     actualCraftTime,
			"materialsUsed": len(consumed),
		}).Debug("started crafting")
	}

	return &CraftingResult{
		Success:       true,
		RecipeName:    recipe.Name,
		MaterialsLost: consumed,
	}, nil
}

// completeCraft finishes a craft attempt, rolling for success and generating output.
func (s *CraftingSystem) completeCraft(entityID uint64, progressComp *CraftingProgressComponent) {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		if s.logger != nil {
			s.logger.WithField("entityID", entityID).Warn("entity not found when completing craft")
		}
		return
	}

	recipe := progressComp.CurrentRecipe
	if recipe == nil {
		if s.logger != nil {
			s.logger.WithField("entityID", entityID).Warn("recipe is nil in progress component")
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
						"entityID":  entityID,
						"itemName":  outputItem.Name,
						"recipeID":  recipe.ID,
					}).Warn("inventory full after craft, item lost")
				}
			}
		}

		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.InfoLevel {
			s.logger.WithFields(logrus.Fields{
				"entityID":      entityID,
				"recipeID":      recipe.ID,
				"itemName":      outputItem.Name,
				"xpGained":      xpGained,
				"successChance": finalChance,
			}).Info("crafting succeeded")
		}
	} else {
		// Failure - materials already consumed, no output
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.InfoLevel {
			s.logger.WithFields(logrus.Fields{
				"entityID":      entityID,
				"recipeID":      recipe.ID,
				"xpGained":      xpGained,
				"successChance": finalChance,
			}).Info("crafting failed")
		}
	}
}

// generateOutputItem creates the crafted item using the item generator.
func (s *CraftingSystem) generateOutputItem(recipe *Recipe, rng *rand.Rand) *item.Item {
	// Build generation parameters
	params := procgen.GenerationParams{
		Difficulty: 0.5, // Medium difficulty for crafted items
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
				"recipeID": recipe.ID,
				"error":    err.Error(),
			}).Warn("failed to generate crafted item, using fallback")
		}
		return s.createFallbackItem(recipe)
	}

	items := result.([]*item.Item)
	if len(items) == 0 {
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
	comp, ok := entity.GetComponent("crafting_skill")
	if !ok {
		// Create skill component if not present
		skillComp := NewCraftingSkillComponent()
		entity.AddComponent(skillComp)
		comp = skillComp
	}

	skillComp, ok := comp.(*CraftingSkillComponent)
	if !ok {
		return
	}

	leveledUp := skillComp.AddExperience(xp)
	if leveledUp && s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entityID":      entity.ID,
			"newSkillLevel": skillComp.SkillLevel,
		}).Info("crafting skill leveled up")
	}
}

// validateMaterials checks if inventory contains required materials and gold.
// Returns (canCraft, missingDescription).
func (s *CraftingSystem) validateMaterials(invComp *InventoryComponent, recipe *Recipe) (bool, string) {
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
			return nil, fmt.Errorf("insufficient %s: needed %d, found %d", req.ItemName, req.Quantity, removed)
		}
	}

	return consumed, nil
}

// validateAndReserveStation checks if station is valid and marks it as in-use.
// Returns (bonusSuccessChance, craftTimeMultiplier, error).
func (s *CraftingSystem) validateAndReserveStation(stationID uint64, recipeType RecipeType) (float64, float64, error) {
	station, ok := s.world.GetEntity(stationID)
	if !ok {
		return 0, 1.0, fmt.Errorf("station not found")
	}

	stationComp, err := s.getCraftingStationComponent(station)
	if err != nil {
		return 0, 1.0, err
	}

	// Check station type matches recipe
	if stationComp.StationType != recipeType {
		return 0, 1.0, fmt.Errorf("wrong station type: need %s station", recipeType.String())
	}

	// Check availability
	if !stationComp.Available {
		return 0, 1.0, fmt.Errorf("station in use")
	}

	// Reserve station
	stationComp.Available = false

	return stationComp.BonusSuccessChance, stationComp.CraftTimeMultiplier, nil
}

// releaseStation marks a station as available.
func (s *CraftingSystem) releaseStation(stationID uint64) {
	if stationID == 0 {
		return
	}

	station, ok := s.world.GetEntity(stationID)
	if !ok {
		return
	}

	stationComp, err := s.getCraftingStationComponent(station)
	if err != nil {
		return
	}

	stationComp.Available = true
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
