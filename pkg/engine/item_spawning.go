// Package engine provides item spawning and loot drop functionality.
// This file implements SpawnItemInWorld, SpawnRecipeInWorld, and loot drop mechanics for the combat system.
package engine

import (
	"fmt"
	"image/color"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/magic"
	"github.com/opd-ai/venture/pkg/procgen/skills"
	"github.com/sirupsen/logrus"
)

// ItemEntityComponent marks an entity as representing a collectable item in the world.
// When the player collides with this entity, the item is added to their inventory.
type ItemEntityComponent struct {
	Item *item.Item // The procedurally generated item
}

// Type returns the component type identifier.
func (i *ItemEntityComponent) Type() string {
	return "item_entity"
}

// SpawnItemInWorld creates an item entity at the specified world position.
// The item becomes a physical object that players can walk over to collect.
// Returns the spawned item entity.
func SpawnItemInWorld(world *World, itm *item.Item, x, y float64) *Entity {
	if itm == nil {
		return nil
	}

	// Create item entity
	itemEntity := world.CreateEntity()

	// Position in world
	itemEntity.AddComponent(&PositionComponent{
		X: x,
		Y: y,
	})

	// Visual representation
	itemSize := 24.0
	itemColor := getItemColor(itm)
	sprite := NewSpriteComponent(itemSize, itemSize, itemColor)
	sprite.Layer = 3 // Items drawn below entities but above terrain
	itemEntity.AddComponent(sprite)

	// Collision for pickup detection
	itemEntity.AddComponent(&ColliderComponent{
		Width:     itemSize,
		Height:    itemSize,
		Solid:     false, // Items don't block movement
		IsTrigger: true,  // Trigger collision events for pickup
		Layer:     3,     // Item collision layer
		OffsetX:   -itemSize / 2,
		OffsetY:   -itemSize / 2,
	})

	// Mark as item entity with the item data
	itemEntity.AddComponent(&ItemEntityComponent{
		Item: itm,
	})

	return itemEntity
}

// GenerateLootDrop creates a random item appropriate for the enemy's level and drops it.
// Uses the procedural item generator with scaling based on enemy difficulty.
// Returns nil if no loot should be dropped (based on drop chance).
func GenerateLootDrop(world *World, enemy *Entity, x, y float64, seed int64, genreID string) *Entity {
	dropChance := calculateLootDropChance(enemy)

	rng := rand.New(rand.NewSource(seed + int64(enemy.ID)))
	if rng.Float64() > dropChance {
		return nil // No drop
	}

	depth := extractEnemyLevel(enemy)
	generatedItem := generateLootItem(seed, enemy.ID, depth, genreID)

	if generatedItem == nil {
		return nil
	}

	return SpawnItemInWorld(world, generatedItem, x, y)
}

// calculateLootDropChance determines drop chance based on enemy strength.
func calculateLootDropChance(enemy *Entity) float64 {
	dropChance := 0.3 // 30% base drop chance

	if statsComp, ok := enemy.GetComponent("stats"); ok {
		if stats, ok := statsComp.(*StatsComponent); ok {
			if stats.Attack > 20 || stats.Defense > 20 {
				dropChance = 0.7 // 70% for strong enemies
			}
		}
	}

	return dropChance
}

// extractEnemyLevel determines item generation depth from enemy level.
func extractEnemyLevel(enemy *Entity) int {
	depth := 1
	if expComp, ok := enemy.GetComponent("experience"); ok {
		if exp, ok := expComp.(*ExperienceComponent); ok {
			depth = exp.Level
		}
	}
	return depth
}

// generateLootItem creates a procedurally generated item for the enemy drop.
func generateLootItem(seed int64, enemyID uint64, depth int, genreID string) *item.Item {
	itemGen := item.NewItemGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5 + float64(depth)*0.05, // Scale with depth
		Depth:      depth,
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"count": 1,
		},
	}

	result, err := itemGen.Generate(seed+int64(enemyID)+100, params)
	if err != nil {
		return nil
	}

	items := result.([]*item.Item)
	if len(items) == 0 {
		return nil
	}

	return items[0]
}

// GenerateRecipeDrop creates a random recipe appropriate for the enemy's level and drops it.
// Uses the procedural recipe generator with scaling based on enemy difficulty.
// Returns nil if no recipe should be dropped (based on drop chance).
// Recipe drop chances are lower than item drops to maintain balance.
func GenerateRecipeDrop(recipeGen procgen.Generator, world *World, enemy *Entity, x, y float64, seed int64, genreID string) *Entity {
	dropChance := calculateRecipeDropChance(enemy)

	rng := rand.New(rand.NewSource(seed + int64(enemy.ID) + 500))
	if rng.Float64() > dropChance {
		return nil // No recipe drop
	}

	depth, difficulty := extractRecipeGenerationParams(enemy)
	generatedRecipe := generateRecipe(recipeGen, seed, enemy.ID, depth, difficulty, genreID)

	if generatedRecipe == nil {
		return nil
	}

	return SpawnRecipeInWorld(world, generatedRecipe, x, y)
}

// calculateRecipeDropChance determines recipe drop chance based on enemy strength.
func calculateRecipeDropChance(enemy *Entity) float64 {
	dropChance := 0.05 // 5% base drop chance for recipes

	if statsComp, ok := enemy.GetComponent("stats"); ok {
		if stats, ok := statsComp.(*StatsComponent); ok {
			if stats.Attack > 20 || stats.Defense > 20 {
				dropChance = 0.2 // 20% for strong enemies
			}
		}
	}

	return dropChance
}

// extractRecipeGenerationParams determines recipe depth and difficulty from enemy stats.
func extractRecipeGenerationParams(enemy *Entity) (int, float64) {
	depth := 1
	difficulty := 0.3 // Start lower for common recipes

	if expComp, ok := enemy.GetComponent("experience"); ok {
		if exp, ok := expComp.(*ExperienceComponent); ok {
			depth = exp.Level
			difficulty = 0.3 + float64(depth)*0.05 // Scale with depth
		}
	}

	return depth, difficulty
}

// generateRecipe creates a procedurally generated recipe for the enemy drop.
func generateRecipe(recipeGen procgen.Generator, seed int64, enemyID uint64, depth int, difficulty float64, genreID string) *Recipe {
	params := procgen.GenerationParams{
		Difficulty: difficulty,
		Depth:      depth,
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"count": 1, // Generate 1 recipe
		},
	}

	result, err := recipeGen.Generate(seed+int64(enemyID)+1000, params)
	if err != nil {
		return nil
	}

	recipes := result.([]*Recipe)
	if len(recipes) == 0 {
		return nil
	}

	return recipes[0]
}

// getItemColor determines the sprite color based on item type and rarity.
func getItemColor(itm *item.Item) color.RGBA {
	// Base color by item type
	var baseColor color.RGBA
	switch itm.Type {
	case item.TypeWeapon:
		baseColor = color.RGBA{180, 180, 200, 255} // Silver-ish for weapons
	case item.TypeArmor:
		baseColor = color.RGBA{120, 140, 120, 255} // Green-ish for armor
	case item.TypeConsumable:
		baseColor = color.RGBA{200, 100, 100, 255} // Red-ish for potions
	case item.TypeAccessory:
		baseColor = color.RGBA{200, 200, 100, 255} // Gold-ish for accessories
	default:
		baseColor = color.RGBA{150, 150, 150, 255} // Gray default
	}

	// Modify by rarity
	rarityMultiplier := 1.0
	switch itm.Rarity {
	case item.RarityUncommon:
		rarityMultiplier = 1.1
	case item.RarityRare:
		rarityMultiplier = 1.3
	case item.RarityEpic:
		rarityMultiplier = 1.5
	case item.RarityLegendary:
		rarityMultiplier = 2.0
	}

	// Apply rarity brightness (clamp to 255)
	r := float64(baseColor.R) * rarityMultiplier
	if r > 255 {
		r = 255
	}
	g := float64(baseColor.G) * rarityMultiplier
	if g > 255 {
		g = 255
	}
	b := float64(baseColor.B) * rarityMultiplier
	if b > 255 {
		b = 255
	}

	return color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: 255,
	}
}

// GenerateSpellScrollDrop creates a random spell scroll appropriate for the enemy's level and drops it.
// Uses the procedural magic generator with scaling based on enemy difficulty.
// Returns nil if no spell scroll should be dropped (based on drop chance).
// Spell scrolls are rarer than items: 10% base, 25% for bosses.
func GenerateSpellScrollDrop(spellGen *magic.SpellGenerator, world *World, enemy *Entity, x, y float64, seed int64, genreID string) *Entity {
	dropChance := calculateSpellScrollDropChance(enemy)
	if !rollForScrollDrop(seed, enemy.ID, dropChance) {
		return nil
	}

	spell := generateSpellFromEnemy(spellGen, enemy, seed, genreID)
	if spell == nil {
		return nil
	}

	return createScrollEntity(world, spell, x, y)
}

// calculateSpellScrollDropChance determines drop chance based on enemy strength.
func calculateSpellScrollDropChance(enemy *Entity) float64 {
	dropChance := 0.10
	if statsComp, ok := enemy.GetComponent("stats"); ok {
		if stats, ok := statsComp.(*StatsComponent); ok {
			if stats.Attack > 20 || stats.Defense > 20 {
				dropChance = 0.25
			}
		}
	}
	return dropChance
}

// rollForScrollDrop performs the random drop check.
func rollForScrollDrop(seed int64, enemyID uint64, dropChance float64) bool {
	rng := rand.New(rand.NewSource(seed + int64(enemyID) + 2000))
	return rng.Float64() <= dropChance
}

// generateSpellFromEnemy creates a spell appropriate for the enemy's level.
func generateSpellFromEnemy(spellGen *magic.SpellGenerator, enemy *Entity, seed int64, genreID string) *magic.Spell {
	depth, difficulty := extractSpellParameters(enemy)

	params := procgen.GenerationParams{
		Difficulty: difficulty,
		Depth:      depth,
		GenreID:    genreID,
		Custom:     map[string]interface{}{"count": 1},
	}

	result, err := spellGen.Generate(seed+int64(enemy.ID)+2100, params)
	if err != nil {
		return nil
	}

	spells := result.([]*magic.Spell)
	if len(spells) == 0 {
		return nil
	}
	return spells[0]
}

// extractSpellParameters determines spell depth and difficulty from enemy stats.
func extractSpellParameters(enemy *Entity) (int, float64) {
	depth := 1
	difficulty := 0.4

	if expComp, ok := enemy.GetComponent("experience"); ok {
		if exp, ok := expComp.(*ExperienceComponent); ok {
			depth = exp.Level
			difficulty = 0.4 + float64(depth)*0.05
		}
	}
	return depth, difficulty
}

// createScrollEntity builds the complete scroll entity with components.
func createScrollEntity(world *World, spell *magic.Spell, x, y float64) *Entity {
	scrollEntity := world.CreateEntity()

	scrollEntity.AddComponent(&PositionComponent{X: x, Y: y})

	scrollSize := 24.0
	scrollColor := getSpellScrollColor(spell)
	sprite := NewSpriteComponent(scrollSize, scrollSize, scrollColor)
	sprite.Layer = 3
	scrollEntity.AddComponent(sprite)

	scrollEntity.AddComponent(&ColliderComponent{
		Width:     scrollSize,
		Height:    scrollSize,
		Solid:     false,
		IsTrigger: true,
		Layer:     3,
		OffsetX:   -scrollSize / 2,
		OffsetY:   -scrollSize / 2,
	})

	scrollItem := &item.Item{
		Name:           fmt.Sprintf("Scroll of %s", spell.Name),
		Description:    fmt.Sprintf("A magical scroll containing the spell '%s'. %s (Damage: %d, Mana: %d, Element: %s)", spell.Name, spell.Description, spell.Stats.Damage, spell.Stats.ManaCost, spell.Element.String()),
		Type:           item.TypeConsumable,
		ConsumableType: item.ConsumableScroll,
		Rarity:         item.Rarity(spell.Rarity),
		Seed:           spell.Seed,
	}
	scrollEntity.AddComponent(&ItemEntityComponent{Item: scrollItem})

	return scrollEntity
}

// GenerateSkillBookDrop creates a random skill book for the player to learn new abilities.
// Skill books grant skill points or unlock specific skills when used.
// Returns nil if no skill book should be dropped (based on drop chance).
// Skill books are very rare: 5% base, 15% for bosses.
func GenerateSkillBookDrop(skillGen *skills.SkillTreeGenerator, world *World, enemy *Entity, x, y float64, seed int64, genreID string) *Entity {
	// Calculate skill book drop chance based on enemy type
	dropChance := 0.05 // 5% base drop chance for skill books (very rare)

	// Increase drop chance for bosses/elites
	if statsComp, ok := enemy.GetComponent("stats"); ok {
		if stats, ok := statsComp.(*StatsComponent); ok {
			if stats.Attack > 20 || stats.Defense > 20 {
				dropChance = 0.15 // 15% for strong enemies
			}
		}
	}

	// Roll for drop
	rng := rand.New(rand.NewSource(seed + int64(enemy.ID) + 3000))
	if rng.Float64() > dropChance {
		return nil // No skill book drop
	}

	// Determine skill book depth from enemy stats
	depth := 1

	if expComp, ok := enemy.GetComponent("experience"); ok {
		if exp, ok := expComp.(*ExperienceComponent); ok {
			depth = exp.Level
		}
	}

	// Skill books grant skill points rather than teaching specific skills
	// This allows player choice in how to spend them
	skillPointsGranted := 1 + depth/5 // More points at higher levels

	// Create skill book item entity
	bookEntity := world.CreateEntity()

	// Position in world
	bookEntity.AddComponent(&PositionComponent{
		X: x,
		Y: y,
	})

	// Visual representation - books are brown/tan colored with golden highlights
	bookSize := 24.0
	bookColor := color.RGBA{139, 90, 43, 255} // Brown book with golden tint
	if depth > 10 {
		bookColor = color.RGBA{184, 134, 11, 255} // Golden book for high-level
	}
	sprite := NewSpriteComponent(bookSize, bookSize, bookColor)
	sprite.Layer = 3 // Items drawn below entities but above terrain
	bookEntity.AddComponent(sprite)

	// Collision for pickup detection
	bookEntity.AddComponent(&ColliderComponent{
		Width:     bookSize,
		Height:    bookSize,
		Solid:     false, // Books don't block movement
		IsTrigger: true,  // Trigger collision events for pickup
		Layer:     3,     // Item collision layer
		OffsetX:   -bookSize / 2,
		OffsetY:   -bookSize / 2,
	})

	// Create consumable item representing the skill book
	bookItem := &item.Item{
		Name:           fmt.Sprintf("Tome of Mastery (Tier %d)", depth/5+1),
		Description:    fmt.Sprintf("An ancient tome containing knowledge that grants %d skill point(s) when read.", skillPointsGranted),
		Type:           item.TypeConsumable,
		ConsumableType: item.ConsumableScroll, // Skill books use scroll type (readable consumables)
		Rarity:         item.RarityRare,       // Skill books are always at least rare
		Seed:           seed + int64(enemy.ID) + 3100,
	}

	bookEntity.AddComponent(&ItemEntityComponent{
		Item: bookItem,
	})

	return bookEntity
}

// getSpellScrollColor determines the sprite color based on spell element and rarity.
func getSpellScrollColor(spell *magic.Spell) color.RGBA {
	// Base color by element
	var baseColor color.RGBA
	switch spell.Element {
	case magic.ElementFire:
		baseColor = color.RGBA{220, 80, 40, 255} // Red-orange for fire
	case magic.ElementIce:
		baseColor = color.RGBA{100, 180, 220, 255} // Light blue for ice
	case magic.ElementLightning:
		baseColor = color.RGBA{200, 200, 100, 255} // Yellow for lightning
	case magic.ElementEarth:
		baseColor = color.RGBA{120, 100, 60, 255} // Brown for earth
	case magic.ElementWind:
		baseColor = color.RGBA{180, 220, 200, 255} // Light cyan for wind
	case magic.ElementLight:
		baseColor = color.RGBA{240, 240, 200, 255} // Bright yellow-white for light
	case magic.ElementDark:
		baseColor = color.RGBA{80, 60, 100, 255} // Dark purple for dark magic
	case magic.ElementArcane:
		baseColor = color.RGBA{160, 100, 200, 255} // Purple for arcane
	default:
		baseColor = color.RGBA{150, 120, 180, 255} // Default purple for scrolls
	}

	// Modify by rarity (brighter = rarer)
	rarityMultiplier := 1.0
	switch spell.Rarity {
	case magic.RarityUncommon:
		rarityMultiplier = 1.1
	case magic.RarityRare:
		rarityMultiplier = 1.2
	case magic.RarityEpic:
		rarityMultiplier = 1.4
	case magic.RarityLegendary:
		rarityMultiplier = 1.6
	}

	// Apply rarity brightness (clamp to 255)
	r := float64(baseColor.R) * rarityMultiplier
	if r > 255 {
		r = 255
	}
	g := float64(baseColor.G) * rarityMultiplier
	if g > 255 {
		g = 255
	}
	b := float64(baseColor.B) * rarityMultiplier
	if b > 255 {
		b = 255
	}

	return color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: 255,
	}
}

// ItemPickupSystem handles automatic item pickup when player moves close to items.
type ItemPickupSystem struct {
	world        *World
	pickupRadius float64 // How close player needs to be to auto-pickup
	logger       *logrus.Entry

	// GAP-015 REPAIR: System references for feedback
	audioManager   *AudioManager
	tutorialSystem *EbitenTutorialSystem

	// Callback for particle effects on pickup
	onPickupCallback func(x, y float64, rarity int)
}

// NewItemPickupSystem creates a new item pickup system.
func NewItemPickupSystem(world *World) *ItemPickupSystem {
	return NewItemPickupSystemWithLogger(world, nil)
}

// NewItemPickupSystemWithLogger creates a new item pickup system with a logger.
func NewItemPickupSystemWithLogger(world *World, logger *logrus.Logger) *ItemPickupSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("system", "item_pickup")
	}
	return &ItemPickupSystem{
		world:        world,
		pickupRadius: 32.0, // Default pickup radius (one tile)
		logger:       logEntry,
	}
}

// GAP-015 REPAIR: Helper methods to get system references
func (s *ItemPickupSystem) getAudioManager() *AudioManager {
	if s.audioManager == nil {
		// Lazy lookup from world systems
		for _, sys := range s.world.GetSystems() {
			if audioMgrSys, ok := sys.(*AudioManagerSystem); ok {
				s.audioManager = audioMgrSys.audioManager
				break
			}
		}
	}
	return s.audioManager
}

func (s *ItemPickupSystem) getTutorialSystem() *EbitenTutorialSystem {
	if s.tutorialSystem == nil {
		// Lazy lookup from world systems
		for _, sys := range s.world.GetSystems() {
			if tutSys, ok := sys.(*EbitenTutorialSystem); ok {
				s.tutorialSystem = tutSys
				break
			}
		}
	}
	return s.tutorialSystem
}

// SetPickupCallback sets the callback function for item pickup events.
// The callback receives the pickup position (x, y) and item rarity (0-4).
func (s *ItemPickupSystem) SetPickupCallback(callback func(x, y float64, rarity int)) {
	s.onPickupCallback = callback
}

// Update checks for item-player collisions and handles pickup.
func (s *ItemPickupSystem) Update(entities []*Entity, deltaTime float64) {
	players := s.findPlayers(entities)
	if len(players) == 0 {
		return
	}

	items := s.findItems(entities)
	recipes := s.findRecipes(entities)

	for _, player := range players {
		s.processPlayerPickups(player, items, recipes)
	}
}

// findPlayers returns all entities with input component.
func (s *ItemPickupSystem) findPlayers(entities []*Entity) []*Entity {
	var players []*Entity
	for _, entity := range entities {
		if entity.HasComponent("input") {
			players = append(players, entity)
		}
	}
	return players
}

// findItems returns all entities with item_entity component.
func (s *ItemPickupSystem) findItems(entities []*Entity) []*Entity {
	var items []*Entity
	for _, entity := range entities {
		if entity.HasComponent("item_entity") {
			items = append(items, entity)
		}
	}
	return items
}

// findRecipes returns all entities with recipe_entity component.
func (s *ItemPickupSystem) findRecipes(entities []*Entity) []*Entity {
	var recipes []*Entity
	for _, entity := range entities {
		if entity.HasComponent("recipe_entity") {
			recipes = append(recipes, entity)
		}
	}
	return recipes
}

// processPlayerPickups handles item and recipe pickup for a single player.
func (s *ItemPickupSystem) processPlayerPickups(player *Entity, items, recipes []*Entity) {
	if !s.validatePlayerComponents(player) {
		return
	}

	inventory := s.getPlayerInventory(player)
	if inventory == nil {
		return
	}

	s.processItemPickups(player, inventory, items)
	s.processRecipePickups(player, recipes)
}

// validatePlayerComponents checks if player has required components.
func (s *ItemPickupSystem) validatePlayerComponents(player *Entity) bool {
	_, hasPos := player.GetComponent("position")
	if !hasPos {
		return false
	}

	_, hasInv := player.GetComponent("inventory")
	return hasInv
}

// getPlayerInventory retrieves and validates player's inventory component.
func (s *ItemPickupSystem) getPlayerInventory(player *Entity) *InventoryComponent {
	playerInventory, _ := player.GetComponent("inventory")
	inventory, ok := playerInventory.(*InventoryComponent)
	if !ok {
		return nil
	}
	return inventory
}

// processItemPickups handles item pickup attempts for a player.
func (s *ItemPickupSystem) processItemPickups(player *Entity, inventory *InventoryComponent, items []*Entity) {
	for _, itemEntity := range items {
		if !s.isEntityInRange(player, itemEntity) {
			continue
		}

		itemData := s.getItemEntityData(itemEntity)
		if itemData == nil {
			continue
		}

		s.attemptItemPickup(player, inventory, itemEntity, itemData)
	}
}

// isEntityInRange checks if two entities are within pickup distance.
func (s *ItemPickupSystem) isEntityInRange(player, target *Entity) bool {
	_, hasPos := target.GetComponent("position")
	if !hasPos {
		return false
	}

	distance := GetDistance(player, target)
	return distance <= 32.0
}

// getItemEntityData retrieves and validates item entity component data.
func (s *ItemPickupSystem) getItemEntityData(itemEntity *Entity) *ItemEntityComponent {
	itemEntityComp, hasItemData := itemEntity.GetComponent("item_entity")
	if !hasItemData {
		return nil
	}

	itemData, ok := itemEntityComp.(*ItemEntityComponent)
	if !ok {
		return nil
	}
	return itemData
}

// attemptItemPickup tries to add item to inventory and handles feedback.
func (s *ItemPickupSystem) attemptItemPickup(player *Entity, inventory *InventoryComponent, itemEntity *Entity, itemData *ItemEntityComponent) {
	if inventory.AddItem(itemData.Item) {
		s.world.RemoveEntity(itemEntity.ID)
		s.playItemPickupFeedback(itemEntity, itemData)
	} else {
		s.showInventoryFullMessage()
	}
}

// playItemPickupFeedback provides audio and visual feedback for successful item pickup.
func (s *ItemPickupSystem) playItemPickupFeedback(itemEntity *Entity, itemData *ItemEntityComponent) {
	if audioSys := s.getAudioManager(); audioSys != nil {
		if err := audioSys.PlaySFX("pickup", int64(itemEntity.ID)); err != nil {
			if s.logger != nil {
				s.logger.Debugf("Failed to play item pickup sound: %v", err)
			}
		}
	}

	if tutorialSys := s.getTutorialSystem(); tutorialSys != nil {
		notifText := fmt.Sprintf("Picked up: %s", itemData.Item.Name)
		tutorialSys.ShowNotification(notifText, 2.0)
	}

	// Trigger particle effects callback if registered
	if s.onPickupCallback != nil {
		if pos := itemEntity.GetPosition(); pos != nil {
			rarity := int(itemData.Item.Rarity)
			s.onPickupCallback(pos.X, pos.Y, rarity)
		}
	}
}

// showInventoryFullMessage displays notification when inventory is full.
func (s *ItemPickupSystem) showInventoryFullMessage() {
	if tutorialSys := s.getTutorialSystem(); tutorialSys != nil {
		tutorialSys.ShowNotification("Inventory full!", 2.0)
	}
}

// processRecipePickups handles recipe learning attempts for a player.
func (s *ItemPickupSystem) processRecipePickups(player *Entity, recipes []*Entity) {
	for _, recipeEntity := range recipes {
		if !s.isEntityInRange(player, recipeEntity) {
			continue
		}

		recipeData := s.getRecipeEntityData(recipeEntity)
		if recipeData == nil {
			continue
		}

		s.attemptRecipeLearn(player, recipeEntity, recipeData)
	}
}

// getRecipeEntityData retrieves and validates recipe entity component data.
func (s *ItemPickupSystem) getRecipeEntityData(recipeEntity *Entity) *RecipeEntityComponent {
	recipeEntityComp, hasRecipeData := recipeEntity.GetComponent("recipe_entity")
	if !hasRecipeData {
		return nil
	}

	recipeData, ok := recipeEntityComp.(*RecipeEntityComponent)
	if !ok {
		return nil
	}
	return recipeData
}

// attemptRecipeLearn tries to learn recipe and handles all possible outcomes.
func (s *ItemPickupSystem) attemptRecipeLearn(player, recipeEntity *Entity, recipeData *RecipeEntityComponent) {
	knowledge := s.ensureRecipeKnowledge(player)
	if knowledge == nil {
		return
	}

	if knowledge.KnowsRecipe(recipeData.Recipe.ID) {
		s.showRecipeAlreadyKnownMessage()
		return
	}

	if !knowledge.LearnRecipe(recipeData.Recipe) {
		s.showRecipeLimitReachedMessage()
		return
	}

	s.world.RemoveEntity(recipeEntity.ID)
	s.playRecipeLearnFeedback(recipeEntity, recipeData)
}

// ensureRecipeKnowledge gets or creates recipe knowledge component for player.
func (s *ItemPickupSystem) ensureRecipeKnowledge(player *Entity) *RecipeKnowledgeComponent {
	knowledgeComp, hasKnowledge := player.GetComponent("recipe_knowledge")
	if !hasKnowledge {
		knowledgeComp = NewRecipeKnowledgeComponent(0)
		player.AddComponent(knowledgeComp)
	}

	knowledge, ok := knowledgeComp.(*RecipeKnowledgeComponent)
	if !ok {
		return nil
	}
	return knowledge
}

// showRecipeAlreadyKnownMessage displays notification for duplicate recipe.
func (s *ItemPickupSystem) showRecipeAlreadyKnownMessage() {
	if tutorialSys := s.getTutorialSystem(); tutorialSys != nil {
		tutorialSys.ShowNotification("Recipe already known!", 1.5)
	}
}

// showRecipeLimitReachedMessage displays notification when recipe limit reached.
func (s *ItemPickupSystem) showRecipeLimitReachedMessage() {
	if tutorialSys := s.getTutorialSystem(); tutorialSys != nil {
		tutorialSys.ShowNotification("Cannot learn more recipes!", 2.0)
	}
}

// playRecipeLearnFeedback provides audio and visual feedback for successful recipe learning.
func (s *ItemPickupSystem) playRecipeLearnFeedback(recipeEntity *Entity, recipeData *RecipeEntityComponent) {
	if audioSys := s.getAudioManager(); audioSys != nil {
		if err := audioSys.PlaySFX("spell", int64(recipeEntity.ID)); err != nil {
			if s.logger != nil {
				s.logger.Debugf("Failed to play recipe pickup sound: %v", err)
			}
		}
	}

	if tutorialSys := s.getTutorialSystem(); tutorialSys != nil {
		notifText := fmt.Sprintf("Learned Recipe: %s", recipeData.Recipe.Name)
		tutorialSys.ShowNotification(notifText, 3.0)
	}
}

// RecipeEntityComponent marks an entity as representing a collectable recipe in the world.
// When the player collides with this entity, the recipe is learned.
type RecipeEntityComponent struct {
	Recipe *Recipe // The procedurally generated recipe
}

// Type returns the component type identifier.
func (r *RecipeEntityComponent) Type() string {
	return "recipe_entity"
}

// SpawnRecipeInWorld creates a recipe entity at the specified world position.
// The recipe becomes a physical object that players can walk over to learn.
// Returns the spawned recipe entity.
func SpawnRecipeInWorld(world *World, recipe *Recipe, x, y float64) *Entity {
	if recipe == nil {
		return nil
	}

	// Create recipe entity
	recipeEntity := world.CreateEntity()

	// Position in world
	recipeEntity.AddComponent(&PositionComponent{
		X: x,
		Y: y,
	})

	// Visual representation - recipes look like scrolls/books
	recipeSize := 24.0
	recipeColor := getRecipeColor(recipe)
	sprite := NewSpriteComponent(recipeSize, recipeSize, recipeColor)
	sprite.Layer = 3 // Recipes drawn at same layer as items
	recipeEntity.AddComponent(sprite)

	// Collision for pickup detection
	recipeEntity.AddComponent(&ColliderComponent{
		Width:     recipeSize,
		Height:    recipeSize,
		Solid:     false, // Recipes don't block movement
		IsTrigger: true,  // Trigger collision events for pickup
		Layer:     3,     // Recipe collision layer
		OffsetX:   -recipeSize / 2,
		OffsetY:   -recipeSize / 2,
	})

	// Mark as recipe entity with the recipe data
	recipeEntity.AddComponent(&RecipeEntityComponent{
		Recipe: recipe,
	})

	return recipeEntity
}

// getRecipeColor determines the sprite color based on recipe type and rarity.
// Recipes appear as magical scrolls with colors indicating their properties.
func getRecipeColor(recipe *Recipe) color.RGBA {
	// Base color by recipe type
	var baseColor color.RGBA
	switch recipe.Type {
	case RecipePotion:
		baseColor = color.RGBA{150, 100, 200, 255} // Purple for potions
	case RecipeEnchanting:
		baseColor = color.RGBA{100, 150, 250, 255} // Blue for enchanting
	case RecipeMagicItem:
		baseColor = color.RGBA{200, 150, 100, 255} // Gold for magic items
	default:
		baseColor = color.RGBA{180, 180, 180, 255} // Gray default
	}

	// Modify by rarity
	rarityMultiplier := 1.0
	switch recipe.Rarity {
	case RecipeUncommon:
		rarityMultiplier = 1.15
	case RecipeRare:
		rarityMultiplier = 1.35
	case RecipeEpic:
		rarityMultiplier = 1.6
	case RecipeLegendary:
		rarityMultiplier = 2.0
	}

	// Apply rarity brightness (clamp to 255)
	r := float64(baseColor.R) * rarityMultiplier
	if r > 255 {
		r = 255
	}
	g := float64(baseColor.G) * rarityMultiplier
	if g > 255 {
		g = 255
	}
	b := float64(baseColor.B) * rarityMultiplier
	if b > 255 {
		b = 255
	}

	return color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: 255,
	}
}
