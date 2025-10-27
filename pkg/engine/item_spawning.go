// Package engine provides item spawning and loot drop functionality.
// This file implements SpawnItemInWorld, SpawnRecipeInWorld, and loot drop mechanics for the combat system.
package engine

import (
	"fmt"
	"image/color"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/item"
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
	// Calculate drop chance based on enemy type
	dropChance := 0.3 // 30% base drop chance

	// Increase drop chance for bosses/elites
	if statsComp, ok := enemy.GetComponent("stats"); ok {
		stats := statsComp.(*StatsComponent)
		if stats.Attack > 20 || stats.Defense > 20 {
			dropChance = 0.7 // 70% for strong enemies
		}
	}

	// Roll for drop
	rng := rand.New(rand.NewSource(seed + int64(enemy.ID)))
	if rng.Float64() > dropChance {
		return nil // No drop
	}

	// Determine item depth from enemy stats
	depth := 1
	if expComp, ok := enemy.GetComponent("experience"); ok {
		exp := expComp.(*ExperienceComponent)
		depth = exp.Level
	}

	// Generate item
	itemGen := item.NewItemGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5 + float64(depth)*0.05, // Scale with depth
		Depth:      depth,
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"count": 1,
		},
	}

	result, err := itemGen.Generate(seed+int64(enemy.ID)+100, params)
	if err != nil {
		return nil
	}

	items := result.([]*item.Item)
	if len(items) == 0 {
		return nil
	}

	// Spawn the item in the world
	return SpawnItemInWorld(world, items[0], x, y)
}

// GenerateRecipeDrop creates a random recipe appropriate for the enemy's level and drops it.
// Uses the procedural recipe generator with scaling based on enemy difficulty.
// Returns nil if no recipe should be dropped (based on drop chance).
// Recipe drop chances are lower than item drops to maintain balance.
func GenerateRecipeDrop(recipeGen procgen.Generator, world *World, enemy *Entity, x, y float64, seed int64, genreID string) *Entity {
	// Calculate recipe drop chance based on enemy type
	// Recipes are rarer than items: 5% base, 20% for bosses
	dropChance := 0.05 // 5% base drop chance for recipes

	// Increase drop chance for bosses/elites
	if statsComp, ok := enemy.GetComponent("stats"); ok {
		stats := statsComp.(*StatsComponent)
		if stats.Attack > 20 || stats.Defense > 20 {
			dropChance = 0.2 // 20% for strong enemies
		}
	}

	// Roll for drop
	rng := rand.New(rand.NewSource(seed + int64(enemy.ID) + 500))
	if rng.Float64() > dropChance {
		return nil // No recipe drop
	}

	// Determine recipe depth/difficulty from enemy stats
	depth := 1
	difficulty := 0.3 // Start lower for common recipes

	if expComp, ok := enemy.GetComponent("experience"); ok {
		exp := expComp.(*ExperienceComponent)
		depth = exp.Level
		difficulty = 0.3 + float64(depth)*0.05 // Scale with depth
	}

	// Generate recipe
	params := procgen.GenerationParams{
		Difficulty: difficulty,
		Depth:      depth,
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"count": 1, // Generate 1 recipe
		},
	}

	result, err := recipeGen.Generate(seed+int64(enemy.ID)+1000, params)
	if err != nil {
		return nil
	}

	recipes := result.([]*Recipe)
	if len(recipes) == 0 {
		return nil
	}

	// Spawn the recipe in the world
	return SpawnRecipeInWorld(world, recipes[0], x, y)
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

// ItemPickupSystem handles automatic item pickup when player moves close to items.
type ItemPickupSystem struct {
	world        *World
	pickupRadius float64 // How close player needs to be to auto-pickup

	// GAP-015 REPAIR: System references for feedback
	audioManager   *AudioManager
	tutorialSystem *EbitenTutorialSystem
}

// NewItemPickupSystem creates a new item pickup system.
func NewItemPickupSystem(world *World) *ItemPickupSystem {
	return &ItemPickupSystem{
		world:        world,
		pickupRadius: 32.0, // Default pickup radius (one tile)
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

// Update checks for item-player collisions and handles pickup.
func (s *ItemPickupSystem) Update(entities []*Entity, deltaTime float64) {
	// Find player entities (those with input component)
	var players []*Entity
	for _, entity := range entities {
		if entity.HasComponent("input") {
			players = append(players, entity)
		}
	}

	if len(players) == 0 {
		return
	}

	// Find item entities
	var items []*Entity
	for _, entity := range entities {
		if entity.HasComponent("item_entity") {
			items = append(items, entity)
		}
	}

	// Check each player against each item
	for _, player := range players {
		_, hasPos := player.GetComponent("position")
		if !hasPos {
			continue
		}

		playerInventory, hasInv := player.GetComponent("inventory")
		if !hasInv {
			continue
		}

		inventory := playerInventory.(*InventoryComponent)

		for _, itemEntity := range items {
			_, hasItemPos := itemEntity.GetComponent("position")
			if !hasItemPos {
				continue
			}

			itemEntityComp, hasItemData := itemEntity.GetComponent("item_entity")
			if !hasItemData {
				continue
			}

			itemData := itemEntityComp.(*ItemEntityComponent)

			// Check distance for pickup (32 pixels = 1 tile)
			distance := GetDistance(player, itemEntity)
			if distance <= 32.0 {
				// Attempt to add item to inventory
				if inventory.CanAddItem(itemData.Item) {
					inventory.Items = append(inventory.Items, itemData.Item)

					// Remove item entity from world
					s.world.RemoveEntity(itemEntity.ID)

					// GAP-015 REPAIR: Play pickup sound effect
					if audioSys := s.getAudioManager(); audioSys != nil {
						if err := audioSys.PlaySFX("pickup", int64(itemEntity.ID)); err != nil {
							// Audio failure is non-critical, log and continue
							_ = err
						}
					}

					// GAP-015 REPAIR: Show pickup notification
					if tutorialSys := s.getTutorialSystem(); tutorialSys != nil {
						notifText := fmt.Sprintf("Picked up: %s", itemData.Item.Name)
						tutorialSys.ShowNotification(notifText, 2.0)
					}
				} else {
					// GAP-015 REPAIR: Show "inventory full" message
					if tutorialSys := s.getTutorialSystem(); tutorialSys != nil {
						tutorialSys.ShowNotification("Inventory full!", 2.0)
					}
				}
			}
		}

		// Check for recipe entities
		var recipes []*Entity
		for _, entity := range entities {
			if entity.HasComponent("recipe_entity") {
				recipes = append(recipes, entity)
			}
		}

		// Check each player against each recipe
		for _, recipeEntity := range recipes {
			_, hasRecipePos := recipeEntity.GetComponent("position")
			if !hasRecipePos {
				continue
			}

			recipeEntityComp, hasRecipeData := recipeEntity.GetComponent("recipe_entity")
			if !hasRecipeData {
				continue
			}

			recipeData := recipeEntityComp.(*RecipeEntityComponent)

			// Check distance for pickup (32 pixels = 1 tile)
			distance := GetDistance(player, recipeEntity)
			if distance <= 32.0 {
				// Get player's recipe knowledge component
				knowledgeComp, hasKnowledge := player.GetComponent("recipe_knowledge")
				if !hasKnowledge {
					// Player doesn't have recipe knowledge component, create one
					knowledgeComp = NewRecipeKnowledgeComponent(0) // Unlimited recipes
					player.AddComponent(knowledgeComp)
				}

				knowledge := knowledgeComp.(*RecipeKnowledgeComponent)

				// Check if player already knows this recipe
				if knowledge.KnowsRecipe(recipeData.Recipe.ID) {
					// Already known, show message
					if tutorialSys := s.getTutorialSystem(); tutorialSys != nil {
						tutorialSys.ShowNotification("Recipe already known!", 1.5)
					}
					continue
				}

				// Learn the recipe
				if !knowledge.LearnRecipe(recipeData.Recipe) {
					// Failed to learn (recipe limit reached?)
					if tutorialSys := s.getTutorialSystem(); tutorialSys != nil {
						tutorialSys.ShowNotification("Cannot learn more recipes!", 2.0)
					}
					continue
				}

				// Successfully learned
				// Remove recipe entity from world
				s.world.RemoveEntity(recipeEntity.ID)

				// GAP-015 REPAIR: Play pickup sound effect (different from item pickup)
				if audioSys := s.getAudioManager(); audioSys != nil {
					if err := audioSys.PlaySFX("spell", int64(recipeEntity.ID)); err != nil {
						// Audio failure is non-critical, log and continue
						_ = err
					}
				}

				// GAP-015 REPAIR: Show recipe learned notification
				if tutorialSys := s.getTutorialSystem(); tutorialSys != nil {
					notifText := fmt.Sprintf("Learned Recipe: %s", recipeData.Recipe.Name)
					tutorialSys.ShowNotification(notifText, 3.0)
				}
			}
		}
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
