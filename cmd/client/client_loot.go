//go:build !android && !ios
// +build !android,!ios

package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/magic"
	"github.com/opd-ai/venture/pkg/procgen/quest"
	"github.com/opd-ai/venture/pkg/procgen/recipe"
	"github.com/opd-ai/venture/pkg/procgen/skills"
	"github.com/sirupsen/logrus"
)

// addStarterItems generates and adds starting items to the player's inventory.
// generateStarterWeapon creates a rusty starter weapon for new players.
func generateStarterWeapon(inventory *engine.InventoryComponent, itemGen *item.ItemGenerator, seed int64, genreID string, logger *logrus.Entry) {
	weaponParams := procgen.GenerationParams{
		Difficulty: 0.0, // Easy starter weapon
		Depth:      1,
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"count": 1,
			"type":  "weapon",
			"theme": getGenreTheme(),
		},
	}

	weaponResult, err := itemGen.Generate(seed+1, weaponParams)
	if err != nil {
		logger.WithError(err).Warn("failed to generate starter weapon")
		return
	}

	weapons, ok := weaponResult.([]*item.Item)
	if !ok {
		logger.Warn("unexpected type from weapon generator")
		return
	}
	if len(weapons) > 0 {
		weapon := weapons[0]
		weapon.Name = "Rusty " + weapon.Name // Make it clearly a starter item
		weapon.Stats.Value = 5               // Low value
		inventory.Items = append(inventory.Items, weapon)
		if logger.Logger.GetLevel() >= logrus.InfoLevel {
			logger.WithFields(logrus.Fields{
				"weaponName": weapon.Name,
				"damage":     weapon.Stats.Damage,
			}).Info("added starter weapon")
		}
	}
}

// generateStarterPotions creates minor healing potions for new players.
func generateStarterPotions(inventory *engine.InventoryComponent, itemGen *item.ItemGenerator, seed int64, genreID string, logger *logrus.Entry) {
	potionParams := procgen.GenerationParams{
		Difficulty: 0.0,
		Depth:      1,
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"count": 2,
			"type":  "consumable",
			"theme": getGenreTheme(),
		},
	}

	potionResult, err := itemGen.Generate(seed+2, potionParams)
	if err != nil {
		logger.WithError(err).Warn("failed to generate healing potions")
		return
	}

	potions, ok := potionResult.([]*item.Item)
	if !ok {
		logger.Warn("unexpected type from potion generator")
		return
	}
	for _, potion := range potions {
		potion.Name = "Minor Health Potion"
		potion.Stats.Value = 10
		potion.Stats.Weight = 0.2
		inventory.Items = append(inventory.Items, potion)
	}
	if logger.Logger.GetLevel() >= logrus.InfoLevel && len(potions) > 0 {
		logger.WithField("count", len(potions)).Info("added healing potions")
	}
}

// generateStarterArmor creates worn starter armor for new players.
func generateStarterArmor(inventory *engine.InventoryComponent, itemGen *item.ItemGenerator, seed int64, genreID string, logger *logrus.Entry) {
	armorParams := procgen.GenerationParams{
		Difficulty: 0.0,
		Depth:      1,
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"count": 1,
			"type":  "armor",
			"theme": getGenreTheme(),
		},
	}

	armorResult, err := itemGen.Generate(seed+100, armorParams)
	if err != nil {
		logger.WithError(err).Warn("failed to generate starter armor")
		return
	}

	armors, ok := armorResult.([]*item.Item)
	if !ok {
		logger.Warn("unexpected type from armor generator")
		return
	}
	if len(armors) > 0 {
		armor := armors[0]
		armor.Name = "Worn " + armor.Name
		armor.Stats.Value = 8
		inventory.Items = append(inventory.Items, armor)
		if logger.Logger.GetLevel() >= logrus.InfoLevel {
			logger.WithFields(logrus.Fields{
				"armorName": armor.Name,
				"defense":   armor.Stats.Defense,
			}).Info("added starter armor")
		}
	}
}

func addStarterItems(inventory *engine.InventoryComponent, seed int64, genreID string, logger *logrus.Logger) {
	itemGen := item.NewItemGenerator()
	itemLogger := logging.GeneratorLogger(logger, "item", seed, genreID)

	// Generate starting equipment
	generateStarterWeapon(inventory, itemGen, seed, genreID, itemLogger)
	generateStarterPotions(inventory, itemGen, seed, genreID, itemLogger)
	generateStarterArmor(inventory, itemGen, seed, genreID, itemLogger)

	if logger.GetLevel() >= logrus.InfoLevel {
		itemLogger.WithField("itemCount", len(inventory.Items)).Info("starter items added")
	}
}

// addTutorialQuest creates and adds a tutorial quest to the player's quest tracker.
func addTutorialQuest(tracker *engine.QuestTrackerComponent, seed int64, genreID string, logger *logrus.Logger) {
	// Create a simple tutorial quest manually (more reliable than generation)
	tutorialQuest := &quest.Quest{
		ID:            fmt.Sprintf("tutorial_%d", seed),
		Name:          "Welcome to Venture",
		Type:          quest.TypeExplore,
		Difficulty:    quest.DifficultyTrivial,
		Description:   "Learn the basics of survival in this procedurally generated world. Explore your surroundings, manage your inventory, and prepare for adventure!",
		RequiredLevel: 1,
		Status:        quest.StatusActive,
		Seed:          seed,
		Tags:          []string{"tutorial", "starter"},
		GiverNPC:      "System",
		Objectives: []quest.Objective{
			{
				Description: "Open your inventory (press I)",
				Target:      "inventory",
				Required:    1,
				Current:     0,
			},
			{
				Description: "Check your quest log (press J)",
				Target:      "questlog",
				Required:    1,
				Current:     1, // Auto-complete since they're viewing it now!
			},
			{
				Description: "Explore the dungeon (move with WASD)",
				Target:      "explore",
				Required:    10, // Move 10 tiles
				Current:     0,
			},
		},
		Reward: quest.Reward{
			XP:          50,
			Gold:        25,
			Items:       []string{},
			SkillPoints: 0,
		},
	}

	// Accept the quest
	tracker.AcceptQuest(tutorialQuest, 0)

	if logger.GetLevel() >= logrus.InfoLevel {
		logging.ComponentLogger(logger, "quest").WithFields(logrus.Fields{
			"questName":      tutorialQuest.Name,
			"objectiveCount": len(tutorialQuest.Objectives),
		}).Info("tutorial quest added")
	}
}

// dropInventoryItems drops all items from an entity's inventory with scatter physics.
func dropInventoryItems(
	game *engine.EbitenGame,
	enemy *engine.Entity,
	pos *engine.PositionComponent,
	deadComp *engine.DeadComponent,
) {
	invComp, hasInv := enemy.GetComponent("inventory")
	if !hasInv {
		return
	}
	inventory, ok := invComp.(*engine.InventoryComponent)
	if !ok {
		return
	}

	// Spawn each item in the inventory with scatter physics
	for i, itm := range inventory.Items {
		if itm == nil {
			continue
		}

		// Calculate scatter offset using circular distribution
		angle := float64(i) * 6.28318 / float64(len(inventory.Items)) // 2*PI radians
		scatterDist := 20.0 + float64(i)*5.0                          // Items spread 20-50 pixels out
		offsetX := scatterDist * math.Cos(angle)
		offsetY := scatterDist * math.Sin(angle)

		// Spawn item entity at scattered position
		itemEntity := engine.SpawnItemInWorld(game.World, itm, pos.X+offsetX, pos.Y+offsetY)
		if itemEntity != nil {
			// Add physics velocity for scatter effect (items fly outward then slow down)
			velocityX := offsetX * 3.0 // Initial velocity proportional to offset
			velocityY := offsetY * 3.0
			itemEntity.AddComponent(&engine.VelocityComponent{
				VX: velocityX,
				VY: velocityY,
			})

			// Add friction to slow down items over time
			itemEntity.AddComponent(engine.NewFrictionComponent(0.12)) // 12% friction per frame (at 60 FPS)

			// Track dropped item in DeadComponent
			deadComp.AddDroppedItem(itemEntity.ID)
		}
	}

	// Clear inventory after dropping all items
	inventory.Clear()
}

// dropEquippedItems drops all equipped items from an entity with scatter physics.
func dropEquippedItems(
	game *engine.EbitenGame,
	enemy *engine.Entity,
	pos *engine.PositionComponent,
	deadComp *engine.DeadComponent,
) {
	equipComp, hasEquip := enemy.GetComponent("equipment")
	if !hasEquip {
		return
	}
	equipment, ok := equipComp.(*engine.EquipmentComponent)
	if !ok {
		return
	}
	equippedItems := equipment.UnequipAll()

	// Spawn equipped items with additional scatter
	for i, itm := range equippedItems {
		if itm == nil {
			continue
		}

		// Use different angle range for equipped items (opposite side)
		angle := (float64(i) * 6.28318 / float64(len(equippedItems))) + 3.14159 // Offset by PI
		scatterDist := 30.0 + float64(i)*5.0
		offsetX := scatterDist * math.Cos(angle)
		offsetY := scatterDist * math.Sin(angle)

		itemEntity := engine.SpawnItemInWorld(game.World, itm, pos.X+offsetX, pos.Y+offsetY)
		if itemEntity != nil {
			velocityX := offsetX * 3.0
			velocityY := offsetY * 3.0
			itemEntity.AddComponent(&engine.VelocityComponent{
				VX: velocityX,
				VY: velocityY,
			})

			// Add friction for smooth deceleration
			itemEntity.AddComponent(engine.NewFrictionComponent(0.12))

			deadComp.AddDroppedItem(itemEntity.ID)
		}
	}
}

// generateLegendaryItemDrop generates a legendary item drop for an enemy (1% base, 5% for bosses).
func generateLegendaryItemDrop(world *engine.World, enemy *engine.Entity, x, y float64, seed int64, genreID string) *engine.Entity {
	// Determine drop chance based on enemy stats (bosses have higher stats)
	statsComp, hasStats := enemy.GetComponent("stats")
	baseDropChance := 0.01 // 1% base chance
	if hasStats {
		stats, ok := statsComp.(*engine.StatsComponent)
		if !ok {
			return nil
		}
		// Bosses/elites (high attack/defense) have 5% chance
		if stats.Attack > 20 || stats.Defense > 20 {
			baseDropChance = 0.05
		}
	}

	// Deterministic drop check using enemy ID and seed
	dropRNG := rand.New(rand.NewSource(seed + int64(enemy.ID*7)))
	if dropRNG.Float64() > baseDropChance {
		return nil // No drop
	}

	// Generate legendary item using procgen/item with high quality
	itemGen := item.NewItemGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.9,  // High difficulty for legendary quality
		Depth:      10.0, // High depth for legendary stats
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"quality": "legendary",
		},
	}

	generatedItem, err := itemGen.Generate(seed+int64(enemy.ID*11), params)
	if err != nil {
		return nil
	}

	itm, ok := generatedItem.(*item.Item)
	if !ok {
		return nil
	}

	// Override to ensure legendary rarity
	itm.Rarity = item.RarityLegendary
	itm.Name = "Legendary " + itm.Name // Prefix with "Legendary"

	// Use SpawnItemInWorld to create properly structured item entity
	itemEntity := engine.SpawnItemInWorld(world, itm, x, y)

	// Override sprite color to golden for legendary items
	if spriteComp, ok := itemEntity.GetComponent("sprite"); ok {
		if sprite, ok := spriteComp.(*engine.EbitenSprite); ok {
			sprite.Color = color.RGBA{255, 215, 0, 255} // Golden color
		}
	}

	return itemEntity
}

// spawnProceduralLoot generates and spawns procedural loot drops for non-player entities.
// deps provides game/generator/seed context shared with the calling death callback.
func spawnProceduralLoot(
	deps DeathCallbackDeps,
	playerEntity *engine.Entity,
	enemy *engine.Entity,
	pos *engine.PositionComponent,
	deadComp *engine.DeadComponent,
) {
	// Only for NPCs/enemies, not players
	if enemy.HasComponent("input") {
		return
	}

	// Create deterministic RNG from enemy ID and seed for physics
	physicsRNG := rand.New(rand.NewSource(deps.Seed + int64(enemy.ID)))

	lootEntity := engine.GenerateLootDrop(deps.Game.World, enemy, pos.X, pos.Y, deps.Seed, deps.GenreID)
	if lootEntity != nil {
		// Add physics to procedural loot too
		lootEntity.AddComponent(&engine.VelocityComponent{
			VX: (physicsRNG.Float64()*2.0 - 1.0) * 30.0, // Random velocity -30 to +30
			VY: (physicsRNG.Float64()*2.0 - 1.0) * 30.0,
		})
		// Add friction for smooth deceleration
		lootEntity.AddComponent(engine.NewFrictionComponent(0.12))

		deadComp.AddDroppedItem(lootEntity.ID)
	}

	// Generate and spawn recipe drops (rarer than item drops)
	recipeEntity := engine.GenerateRecipeDrop(deps.RecipeGen, deps.Game.World, enemy, pos.X, pos.Y, deps.Seed, deps.GenreID)
	if recipeEntity != nil {
		// Add physics to recipe drops
		recipeEntity.AddComponent(&engine.VelocityComponent{
			VX: (physicsRNG.Float64()*2.0 - 1.0) * 25.0, // Slightly slower velocity for recipes
			VY: (physicsRNG.Float64()*2.0 - 1.0) * 25.0,
		})
		// Add friction for smooth deceleration
		recipeEntity.AddComponent(engine.NewFrictionComponent(0.12))

		deadComp.AddDroppedItem(recipeEntity.ID)
	}

	// Phase 3.2: Generate and spawn spell scroll drops (10% base, 25% for bosses)
	spellScrollEntity := engine.GenerateSpellScrollDrop(deps.MagicGen, deps.Game.World, enemy, pos.X, pos.Y, deps.Seed, deps.GenreID)
	if spellScrollEntity != nil {
		// Add physics to spell scrolls
		spellScrollEntity.AddComponent(&engine.VelocityComponent{
			VX: (physicsRNG.Float64()*2.0 - 1.0) * 28.0, // Medium velocity for scrolls
			VY: (physicsRNG.Float64()*2.0 - 1.0) * 28.0,
		})
		// Add friction for smooth deceleration
		spellScrollEntity.AddComponent(engine.NewFrictionComponent(0.12))

		deadComp.AddDroppedItem(spellScrollEntity.ID)
	}

	// Phase 3.2: Generate and spawn skill book drops (5% base, 15% for bosses)
	skillBookEntity := engine.GenerateSkillBookDrop(deps.SkillGen, deps.Game.World, enemy, pos.X, pos.Y, deps.Seed, deps.GenreID)
	if skillBookEntity != nil {
		// Add physics to skill books
		skillBookEntity.AddComponent(&engine.VelocityComponent{
			VX: (physicsRNG.Float64()*2.0 - 1.0) * 26.0, // Slower velocity for books (heavier)
			VY: (physicsRNG.Float64()*2.0 - 1.0) * 26.0,
		})
		// Add friction for smooth deceleration
		skillBookEntity.AddComponent(engine.NewFrictionComponent(0.12))

		deadComp.AddDroppedItem(skillBookEntity.ID)
	}

	// Phase 3.4: Generate and spawn legendary item drops (1% base, 5% for bosses)
	legendaryEntity := generateLegendaryItemDrop(deps.Game.World, enemy, pos.X, pos.Y, deps.Seed, deps.GenreID)
	if legendaryEntity != nil {
		// Add physics to legendary items (dramatic velocity)
		legendaryEntity.AddComponent(&engine.VelocityComponent{
			VX: (physicsRNG.Float64()*2.0 - 1.0) * 35.0, // Higher velocity for legendary (dramatic effect)
			VY: (physicsRNG.Float64()*2.0 - 1.0) * 35.0,
		})
		// Add friction for smooth deceleration
		legendaryEntity.AddComponent(engine.NewFrictionComponent(0.12))

		deadComp.AddDroppedItem(legendaryEntity.ID)
	}

	// Track enemy kill for quest objectives
	if playerEntity != nil {
		deps.ObjectiveTracker.OnEnemyKilled(playerEntity, enemy)
	}
}

// createDeathCallback creates the death callback function for the combat system.
// This callback handles entity death by:
// - Dropping inventory and equipped items with scatter physics
// - Generating procedural loot drops for non-player entities
// - Playing death sound effects
// - Spawning death particle effects
// - Tracking quest objectives
// - Awarding XP to the player (AUDIT.md MEDIUM fix)
//
// DeathCallbackDeps bundles the dependencies to avoid a 13-parameter signature
// (AUDIT.md MEDIUM #3 — high parameter-count functions).
type DeathCallbackDeps struct {
	Game                *engine.EbitenGame
	PlayerEntity        **engine.Entity
	ObjectiveTracker    *engine.ObjectiveTrackerSystem
	AudioManager        **engine.AudioManager
	RecipeGen           *recipe.RecipeGenerator
	MagicGen            *magic.SpellGenerator
	SkillGen            *skills.SkillTreeGenerator
	DeathParticleSystem *engine.DeathParticleSystem
	ProgressionSystem   **engine.ProgressionSystem
	Seed                int64
	GenreID             string
	Logger              *logrus.Logger
	TP                  TimeProvider
}

func createDeathCallback(deps DeathCallbackDeps) func(*engine.Entity) {
	game := deps.Game
	playerEntity := deps.PlayerEntity
	audioManager := deps.AudioManager
	deathParticleSystem := deps.DeathParticleSystem
	progressionSystem := deps.ProgressionSystem
	logger := deps.Logger
	tp := deps.TP
	return func(enemy *engine.Entity) {
		// Priority 1.4: Only process death once (callback called every frame while entity is dead)
		if enemy.HasComponent("dead") {
			return
		}

		// Get enemy position
		posComp, hasPos := enemy.GetComponent("position")
		if !hasPos {
			return
		}
		pos := posComp.(*engine.PositionComponent)

		// Priority 1.4: Add DeadComponent to mark entity as dead
		gameTime := float64(tp.Now().Unix())
		deadComp := engine.NewDeadComponent(gameTime)
		enemy.AddComponent(deadComp)

		// AUDIT.md MEDIUM: Award XP to the player for killing the enemy.
		// XP is derived from enemy max HP (1 XP per max-HP point, minimum 10).
		// Guard: skip if the dead entity is a player-controlled entity (has an
		// input component) or if it is the player entity itself.
		if progressionSystem != nil && *progressionSystem != nil &&
			playerEntity != nil && *playerEntity != nil &&
			!enemy.HasComponent("input") && enemy != *playerEntity {
			xpAmount := calculateEnemyXP(enemy)
			if err := (*progressionSystem).AwardXP(*playerEntity, xpAmount); err != nil {
				if logger.GetLevel() >= logrus.DebugLevel {
					logging.ComponentLogger(logger, "progression").WithError(err).Debug("failed to award kill XP")
				}
			}
		}

		// Spawn death particle effects for visual feedback
		if deathParticleSystem != nil {
			deathParticleSystem.OnDeath(enemy)
		}

		// Priority 1.4: Drop all items from entity's inventory
		dropInventoryItems(game, enemy, pos, deadComp)

		// Priority 1.4: Also drop equipped items
		dropEquippedItems(game, enemy, pos, deadComp)

		// Generate and spawn procedural loot drop (in addition to inventory items)
		spawnProceduralLoot(deps, *playerEntity, enemy, pos, deadComp)

		// GAP-010 REPAIR: Play death sound effect
		if *audioManager != nil {
			if err := (*audioManager).PlaySFX("death", tp.Now().UnixNano()); err != nil {
				if logger.GetLevel() >= logrus.WarnLevel {
					logging.ComponentLogger(logger, "audio").WithError(err).Warn("failed to play death SFX")
				}
			}
		}
	}
}

// calculateEnemyXP returns the XP reward for killing an enemy.
// Derives the amount from the enemy's max HP (1 XP per HP point) with a minimum
// of 10 and a maximum of 10 000 to prevent degenerate values.
func calculateEnemyXP(enemy *engine.Entity) int {
	const minXP = 10
	const maxXP = 10_000
	healthComp, ok := enemy.GetComponent("health")
	if !ok {
		return minXP
	}
	hc, ok := healthComp.(*engine.HealthComponent)
	if !ok {
		return minXP
	}
	xp := int(hc.Max)
	if xp < minXP {
		return minXP
	}
	if xp > maxXP {
		return maxXP
	}
	return xp
}
