package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/combat"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/network"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/quest"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/opd-ai/venture/pkg/saveload"
	"github.com/sirupsen/logrus"
)

// animationSystemWrapper adapts AnimationSystem (returns error) to System interface (no return)
type animationSystemWrapper struct {
	system *engine.AnimationSystem
	logger *logrus.Entry
}

func (w *animationSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	if err := w.system.Update(entities, deltaTime); err != nil {
		if w.logger != nil && w.logger.Logger.GetLevel() >= logrus.DebugLevel {
			w.logger.WithError(err).Debug("animation system error")
		}
	}
}

var (
	width       = flag.Int("width", 800, "Screen width")
	height      = flag.Int("height", 600, "Screen height")
	seed        = flag.Int64("seed", seededRandom(), "World generation seed")
	genreID     = flag.String("genre", randomGenre(), "Genre ID (fantasy, scifi, horror, cyberpunk, postapoc)")
	verbose     = flag.Bool("verbose", false, "Enable verbose logging")
	multiplayer = flag.Bool("multiplayer", false, "Enable multiplayer mode (connect to server)")
	server      = flag.String("server", "localhost:8080", "Server address (host:port) for multiplayer")
)

// return a random seed
func seededRandom() int64 {
	time := time.Now().UnixNano()
	rand := rand.New(rand.NewSource(time))
	return rand.Int63()
}

// return a random genre
func randomGenre() string {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	time := time.Now().UnixNano()
	rand := rand.New(rand.NewSource(time))
	return genres[rand.Intn(len(genres))]
}

// addStarterItems generates and adds starting items to the player's inventory.
func addStarterItems(inventory *engine.InventoryComponent, seed int64, genreID string, logger *logrus.Logger) {
	itemGen := item.NewItemGenerator()
	itemLogger := logging.GeneratorLogger(logger, "item", seed, genreID)

	// Generate a starting weapon (1 weapon, common)
	weaponParams := procgen.GenerationParams{
		Difficulty: 0.0, // Easy starter weapon
		Depth:      1,
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"count": 1,
			"type":  "weapon",
		},
	}

	weaponResult, err := itemGen.Generate(seed+1, weaponParams)
	if err != nil {
		itemLogger.WithError(err).Warn("failed to generate starter weapon")
	} else {
		weapons := weaponResult.([]*item.Item)
		if len(weapons) > 0 {
			weapon := weapons[0]
			weapon.Name = "Rusty " + weapon.Name // Make it clearly a starter item
			weapon.Stats.Value = 5               // Low value
			inventory.Items = append(inventory.Items, weapon)
			if logger.GetLevel() >= logrus.InfoLevel {
				itemLogger.WithFields(logrus.Fields{
					"weaponName": weapon.Name,
					"damage":     weapon.Stats.Damage,
				}).Info("added starter weapon")
			}
		}
	}

	// Generate 2 healing potions
	potionParams := procgen.GenerationParams{
		Difficulty: 0.0,
		Depth:      1,
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"count": 2,
			"type":  "consumable",
		},
	}

	potionResult, err := itemGen.Generate(seed+2, potionParams)
	if err != nil {
		itemLogger.WithError(err).Warn("failed to generate healing potions")
	} else {
		potions := potionResult.([]*item.Item)
		for _, potion := range potions {
			potion.Name = "Minor Health Potion"
			potion.Stats.Value = 10
			potion.Stats.Weight = 0.2
			inventory.Items = append(inventory.Items, potion)
		}
		if logger.GetLevel() >= logrus.InfoLevel && len(potions) > 0 {
			itemLogger.WithField("count", len(potions)).Info("added healing potions")
		}
	}

	// Generate a piece of armor (1 armor, common)
	armorParams := procgen.GenerationParams{
		Difficulty: 0.0,
		Depth:      1,
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"count": 1,
			"type":  "armor",
		},
	}

	armorResult, err := itemGen.Generate(seed+100, armorParams)
	if err != nil {
		itemLogger.WithError(err).Warn("failed to generate starter armor")
	} else {
		armors := armorResult.([]*item.Item)
		if len(armors) > 0 {
			armor := armors[0]
			armor.Name = "Worn " + armor.Name
			armor.Stats.Value = 8
			inventory.Items = append(inventory.Items, armor)
			if logger.GetLevel() >= logrus.InfoLevel {
				itemLogger.WithFields(logrus.Fields{
					"armorName": armor.Name,
					"defense":   armor.Stats.Defense,
				}).Info("added starter armor")
			}
		}
	}

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

func main() {
	flag.Parse()

	// Initialize structured logger
	logConfig := logging.DefaultConfig()

	// Check for JSON format from environment (default to text for client)
	if logFormat := os.Getenv("LOG_FORMAT"); logFormat == "json" {
		logConfig.Format = logging.JSONFormat
	} else {
		logConfig.Format = logging.TextFormat
		logConfig.EnableColor = true
	}

	// Set log level from environment or use Info as default
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		logConfig.Level = logging.LogLevel(logLevel)
	} else if *verbose {
		logConfig.Level = logging.DebugLevel
	} else {
		logConfig.Level = logging.InfoLevel
	}

	logger := logging.NewLogger(logConfig)
	clientLogger := logger.WithFields(logrus.Fields{
		"component": "client",
		"genre":     *genreID,
		"seed":      *seed,
	})

	clientLogger.Info("Starting Venture - Procedural Action RPG")
	clientLogger.WithFields(logrus.Fields{
		"width":  *width,
		"height": *height,
		"seed":   *seed,
		"genre":  *genreID,
	}).Info("client configuration")

	// Initialize network client if multiplayer mode is enabled
	var networkClient network.ClientConnection
	if *multiplayer {
		clientLogger.WithField("server", *server).Info("multiplayer mode enabled - connecting to server")

		clientConfig := network.DefaultClientConfig()
		clientConfig.ServerAddress = *server
		networkClient = network.NewClientWithLogger(clientConfig, logger)

		// Connect to server
		if err := networkClient.Connect(); err != nil {
			clientLogger.WithError(err).Fatal("failed to connect to server")
		}

		clientLogger.Info("connected to server successfully")

		// Handle network errors in background
		go func() {
			for err := range networkClient.ReceiveError() {
				clientLogger.WithError(err).Error("network error")
			}
		}()
	} else {
		clientLogger.Info("single-player mode (use -multiplayer flag to connect to server)")
	}

	// Create the game instance
	game := engine.NewEbitenGameWithLogger(*width, *height, logger)

	// Initialize game systems
	clientLogger.Info("initializing game systems")

	// Add core gameplay systems
	inputSystem := engine.NewInputSystem()
	// GAP-001 & GAP-002 REPAIR: Use proper constructors with required parameters
	movementSystem := engine.NewMovementSystem(200.0)  // 200 units/second max speed
	collisionSystem := engine.NewCollisionSystem(64.0) // 64-unit grid cells for spatial partitioning

	// GAP-001 REPAIR: Connect collision system to movement system for predictive collision
	movementSystem.SetCollisionSystem(collisionSystem)

	combatSystem := engine.NewCombatSystemWithLogger(*seed, logger)

	// GAP-016 REPAIR: Initialize particle system for visual effects
	particleSystem := engine.NewParticleSystem()

	// GAP-017 REPAIR: Initialize animation system for animated sprites
	spriteGenerator := sprites.NewGenerator()
	animationSystem := engine.NewAnimationSystem(spriteGenerator)

	// Store player reference for death callback (will be set after player creation)
	var playerEntity *engine.Entity

	// Store audio manager reference for callbacks (will be set after audio system creation)
	var audioManager *engine.AudioManager

	// GAP-004 REPAIR: Add objective tracker system for quest progress
	objectiveTracker := engine.NewObjectiveTrackerSystem()

	// Set quest completion callback to award rewards
	objectiveTracker.SetQuestCompleteCallback(func(entity *engine.Entity, qst *quest.Quest) {
		objectiveTracker.AwardQuestRewards(entity, qst)
		if logger.GetLevel() >= logrus.InfoLevel {
			logging.ComponentLogger(logger, "quest").WithFields(logrus.Fields{
				"questName":   qst.Name,
				"xpReward":    qst.Reward.XP,
				"goldReward":  qst.Reward.Gold,
				"skillPoints": qst.Reward.SkillPoints,
			}).Info("quest completed")
		}
	})

	// GAP-001 & GAP-004 REPAIR: Set death callback for loot drops and quest tracking
	combatSystem.SetDeathCallback(func(enemy *engine.Entity) {
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
		gameTime := float64(time.Now().Unix()) // Use game time if available
		deadComp := engine.NewDeadComponent(gameTime)
		enemy.AddComponent(deadComp)

		// Priority 1.4: Drop all items from entity's inventory
		if invComp, hasInv := enemy.GetComponent("inventory"); hasInv {
			inventory := invComp.(*engine.InventoryComponent)

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

		// Priority 1.4: Also drop equipped items
		if equipComp, hasEquip := enemy.GetComponent("equipment"); hasEquip {
			equipment := equipComp.(*engine.EquipmentComponent)
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

		// Generate and spawn procedural loot drop (in addition to inventory items)
		// This is for enemies that don't have inventory but should drop random loot
		if !enemy.HasComponent("input") { // Only for NPCs/enemies, not players
			lootEntity := engine.GenerateLootDrop(game.World, enemy, pos.X, pos.Y, *seed, *genreID)
			if lootEntity != nil {
				// Add physics to procedural loot too
				lootEntity.AddComponent(&engine.VelocityComponent{
					VX: (rand.Float64()*2.0 - 1.0) * 30.0, // Random velocity -30 to +30
					VY: (rand.Float64()*2.0 - 1.0) * 30.0,
				})
				// Add friction for smooth deceleration
				lootEntity.AddComponent(engine.NewFrictionComponent(0.12))

				deadComp.AddDroppedItem(lootEntity.ID)
			}

			// Track enemy kill for quest objectives
			if playerEntity != nil {
				objectiveTracker.OnEnemyKilled(playerEntity, enemy)
			}
		}

		// GAP-010 REPAIR: Play death sound effect
		if err := audioManager.PlaySFX("death", time.Now().UnixNano()); err != nil {
			if logger.GetLevel() >= logrus.WarnLevel {
				logging.ComponentLogger(logger, "audio").WithError(err).Warn("failed to play death SFX")
			}
		}
	})

	aiSystem := engine.NewAISystem(game.World)
	progressionSystem := engine.NewProgressionSystem(game.World)
	inventorySystem := engine.NewInventorySystem(game.World)

	// GAP-010 REPAIR: Initialize audio system
	audioManager = engine.NewAudioManager(44100, *seed) // 44.1kHz sample rate
	audioManagerSystem := engine.NewAudioManagerSystem(audioManager)

	// Start playing exploration music
	if err := audioManager.PlayMusic(*genreID, "exploration"); err != nil {
		logging.ComponentLogger(logger, "audio").WithError(err).Warn("failed to start background music")
	}

	logging.ComponentLogger(logger, "audio").Info("audio system initialized (music and SFX generators)")

	// Add item pickup system to automatically collect nearby items
	itemPickupSystem := engine.NewItemPickupSystem(game.World)

	// GAP-002 REPAIR: Add spell casting systems
	// Initialize status effect system first (required by spell casting system)
	statusEffectRNG := rand.New(rand.NewSource(*seed + 999)) // Use seed offset for status effects
	statusEffectSystem := engine.NewStatusEffectSystem(game.World, statusEffectRNG)
	spellCastingSystem := engine.NewSpellCastingSystem(game.World, statusEffectSystem)
	playerSpellCastingSystem := engine.NewPlayerSpellCastingSystem(spellCastingSystem, game.World)
	manaRegenSystem := &engine.ManaRegenSystem{} // GAP #2 REPAIR: Add player combat system to connect Space key to combat
	playerCombatSystem := engine.NewPlayerCombatSystem(combatSystem, game.World)

	// GAP #3 REPAIR: Add player item use system to connect E key to inventory
	playerItemUseSystem := engine.NewPlayerItemUseSystem(inventorySystem, game.World)

	// Add tutorial and help systems (Phase 8.6)
	tutorialSystem := engine.NewTutorialSystem()
	helpSystem := engine.NewHelpSystem()

	// Connect help system to input system for ESC key handling
	inputSystem.SetHelpSystem(helpSystem)
	// Connect tutorial system to input system for ESC key skip handling
	inputSystem.SetTutorialSystem(tutorialSystem)

	// Add systems in correct order:
	// 1. Input - captures player actions
	// 2. Player Combat/Item Use/Spell Casting - processes input flags
	// 3. Movement - applies velocity to position
	// 4. Collision - checks and resolves collisions
	// 5. Combat - handles damage/status effects
	// 6. Status Effects - processes DoT, buffs, debuffs, shields
	// 7. AI - enemy decision-making
	// 8. Progression - XP and leveling
	// 9. Skill Progression - applies skill effects to stats
	// 10. Audio Manager - updates music based on game context
	// 11. Objective Tracker - updates quest progress
	// 12. Item Pickup - collects nearby items
	// 13. Spell Casting - executes spell effects
	// 14. Mana Regen - regenerates mana
	// 15. Inventory - item management
	// 16. Animation - updates sprite frames (before rendering)
	// 17. Tutorial/Help - UI overlays
	game.World.AddSystem(inputSystem)
	game.World.AddSystem(playerCombatSystem)
	game.World.AddSystem(playerItemUseSystem)
	game.World.AddSystem(playerSpellCastingSystem)
	game.World.AddSystem(movementSystem)
	game.World.AddSystem(collisionSystem)
	game.World.AddSystem(combatSystem)
	game.World.AddSystem(statusEffectSystem) // Process status effects after combat

	// Add revival system for multiplayer death mechanics (Category 1.1)
	// Allows living players to revive dead teammates through proximity interaction
	revivalSystem := engine.NewRevivalSystem(game.World)
	game.World.AddSystem(revivalSystem)

	game.World.AddSystem(aiSystem)
	game.World.AddSystem(progressionSystem)

	// Add skill progression system
	skillProgressionSystem := engine.NewSkillProgressionSystem()
	game.World.AddSystem(skillProgressionSystem)

	// GAP-012 REPAIR: Add visual feedback system for hit flashes and tints
	visualFeedbackSystem := engine.NewVisualFeedbackSystem()
	game.World.AddSystem(visualFeedbackSystem)

	// Add audio manager system
	game.World.AddSystem(audioManagerSystem)

	// Add objective tracker system
	game.World.AddSystem(objectiveTracker)

	game.World.AddSystem(itemPickupSystem)
	game.World.AddSystem(spellCastingSystem)
	game.World.AddSystem(manaRegenSystem)
	game.World.AddSystem(inventorySystem)

	// GAP-017 REPAIR: Add animation system before tutorial/help to update sprites first
	game.World.AddSystem(&animationSystemWrapper{
		system: animationSystem,
		logger: game.World.GetLogger(),
	})

	game.World.AddSystem(tutorialSystem)
	game.World.AddSystem(helpSystem)

	// GAP-016 REPAIR: Add particle system for rendering effects
	game.World.AddSystem(particleSystem)

	// Store references to tutorial and help systems in game for rendering
	game.TutorialSystem = tutorialSystem
	game.HelpSystem = helpSystem

	// GAP-012 REPAIR: Set camera reference on combat system for screen shake
	combatSystem.SetCamera(game.CameraSystem)

	// GAP-016 REPAIR: Set particle system reference on combat system for hit effects
	combatSystem.SetParticleSystem(particleSystem, game.World, *genreID)

	if *verbose {
		log.Println("Systems initialized: Input, PlayerCombat, PlayerItemUse, PlayerSpellCasting, Movement, Collision, Combat, StatusEffects, AI, Progression, SkillProgression, VisualFeedback, AudioManager, ObjectiveTracker, ItemPickup, SpellCasting, ManaRegen, Inventory, Animation, Tutorial, Help, Particles")
	} // Gap #3: Initialize performance monitoring (wraps World.Update)
	perfMonitor := engine.NewPerformanceMonitor(game.World)
	if *verbose {
		log.Println("Performance monitoring initialized")
		// Start periodic performance logging in background
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				metrics := perfMonitor.GetMetrics()
				log.Printf("Performance: %s", metrics.String())
			}
		}()
	}
	_ = perfMonitor // Suppress unused warning when not verbose

	// Generate initial world terrain
	clientLogger.Info("generating procedural terrain")

	terrainGen := terrain.NewBSPGeneratorWithLogger(logger) // Use BSP algorithm with logging
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    *genreID,
		Custom: map[string]interface{}{
			"width":  80,
			"height": 50,
		},
	}

	terrainResult, err := terrainGen.Generate(*seed, params)
	if err != nil {
		clientLogger.WithError(err).Fatal("failed to generate terrain")
	}

	generatedTerrain := terrainResult.(*terrain.Terrain)
	clientLogger.WithFields(logrus.Fields{
		"width":     generatedTerrain.Width,
		"height":    generatedTerrain.Height,
		"roomCount": len(generatedTerrain.Rooms),
	}).Info("terrain generated")

	// Initialize terrain rendering system
	if *verbose {
		log.Println("Initializing terrain rendering system...")
	}

	terrainRenderSystem := engine.NewTerrainRenderSystem(32, 32, *genreID, *seed)
	terrainRenderSystem.SetTerrain(generatedTerrain)
	game.TerrainRenderSystem = terrainRenderSystem

	if *verbose {
		log.Println("Terrain rendering system initialized")
	}

	// GAP REPAIR: Initialize efficient terrain collision checking
	if *verbose {
		log.Println("Initializing terrain collision system...")
	}

	terrainChecker := engine.NewTerrainCollisionChecker(32, 32)
	terrainChecker.SetTerrain(generatedTerrain)

	// Connect terrain checker to collision system
	for _, system := range game.World.GetSystems() {
		if collisionSys, ok := system.(*engine.CollisionSystem); ok {
			collisionSys.SetTerrainChecker(terrainChecker)
			break
		}
	}

	if *verbose {
		log.Printf("Terrain collision system initialized (efficient mode)")
	}

	// CATEGORY 4.3: Initialize spatial partition system for viewport culling
	// Provides significant performance benefits with large entity counts through spatial queries
	// Always enabled as a core optimization (previously optional, now standard)
	if *verbose {
		log.Println("Initializing spatial partition system for viewport culling...")
	}

	// Calculate world bounds from terrain dimensions (32 pixels per tile)
	worldWidth := float64(generatedTerrain.Width) * 32.0
	worldHeight := float64(generatedTerrain.Height) * 32.0

	// Create spatial partition system with quadtree-based structure
	spatialSystem := engine.NewSpatialPartitionSystem(worldWidth, worldHeight)

	// Register with ECS World for automatic updates every 60 frames
	game.World.AddSystem(spatialSystem)

	// Connect to render system for viewport culling
	game.RenderSystem.SetSpatialPartition(spatialSystem)
	game.RenderSystem.EnableCulling(true)

	clientLogger.WithFields(logrus.Fields{
		"worldWidth":  worldWidth,
		"worldHeight": worldHeight,
		"cellSize":    8, // Quadtree capacity per node (8 entities before subdivision)
	}).Info("spatial partition system initialized with viewport culling enabled")

	if *verbose {
		log.Printf("Spatial partition enabled: world bounds %.0fx%.0f pixels", worldWidth, worldHeight)
	}

	// GAP-001 REPAIR: Connect terrain to MapUI for map functionality
	if *verbose {
		log.Println("Connecting terrain to Map UI...")
	}
	game.MapUI.SetTerrain(generatedTerrain)
	if *verbose {
		log.Println("Map UI configured with terrain data")
	}

	// GAP #1 REPAIR: Spawn enemies in terrain rooms
	if *verbose {
		log.Println("Spawning enemies in dungeon rooms...")
	}

	enemyParams := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    *genreID,
	}

	enemyCount, err := engine.SpawnEnemiesInTerrain(game.World, generatedTerrain, *seed, enemyParams)
	if err != nil {
		log.Printf("Warning: Failed to spawn enemies: %v", err)
	} else if *verbose {
		log.Printf("Spawned %d enemies across %d rooms", enemyCount, len(generatedTerrain.Rooms)-1)
	}

	// Create player entity
	if *verbose {
		log.Println("Creating player entity...")
	}

	player := game.World.CreateEntity()

	// Store player entity reference for death callback
	playerEntity = player

	// GAP #3 REPAIR: Calculate player spawn position from first room
	var playerX, playerY float64
	if len(generatedTerrain.Rooms) > 0 {
		// Spawn in center of first room
		firstRoom := generatedTerrain.Rooms[0]
		cx, cy := firstRoom.Center()
		playerX = float64(cx * 32) // Convert tile coordinates to world coordinates
		playerY = float64(cy * 32)
		if *verbose {
			log.Printf("Player spawning in first room at tile (%d, %d), world (%.0f, %.0f)",
				cx, cy, playerX, playerY)
		}
	} else {
		// Fallback to default position if no rooms (shouldn't happen with valid terrain)
		playerX, playerY = 400, 300
		log.Println("Warning: No rooms in terrain, using default spawn position")
	}

	// Add player components
	player.AddComponent(&engine.PositionComponent{X: playerX, Y: playerY})
	player.AddComponent(&engine.VelocityComponent{VX: 0, VY: 0})
	player.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
	player.AddComponent(&engine.TeamComponent{TeamID: 1}) // Player team

	// Add input component for player control
	player.AddComponent(&engine.EbitenInput{})

	// GAP-017 REPAIR: Add animated sprite instead of static sprite
	playerSprite := &engine.EbitenSprite{
		Image:   ebiten.NewImage(28, 28), // Initial image (will be replaced by animation)
		Width:   28,
		Height:  28,
		Visible: true,
		Layer:   10, // Draw player on top
	}
	player.AddComponent(playerSprite)

	// Add animation component for multi-frame character animation
	// GAP-019 REPAIR: Use special seed offset for player to ensure distinct color
	playerAnim := engine.NewAnimationComponent(*seed + int64(player.ID*1000))
	playerAnim.CurrentState = engine.AnimationStateIdle
	playerAnim.FrameTime = 0.15 // ~6.7 FPS for smooth animation
	playerAnim.Loop = true
	playerAnim.Playing = true
	playerAnim.FrameCount = 4 // 4 frames per animation
	player.AddComponent(playerAnim)

	// Add camera that follows the player
	camera := engine.NewCameraComponent()
	camera.Smoothing = 0.1
	player.AddComponent(camera)

	// Set player as the active camera
	game.CameraSystem.SetActiveCamera(player)

	// Set player for HUD display
	game.HUDSystem.SetPlayerEntity(player)

	// Set player for UI systems (inventory, quests)
	game.SetPlayerEntity(player)

	// Add player stats
	playerStats := engine.NewStatsComponent()
	playerStats.Attack = 10
	playerStats.Defense = 5
	// GAP-003 REPAIR: Initialize derived stats with baseline values
	playerStats.CritChance = 0.05 // 5% crit chance
	playerStats.CritDamage = 1.5  // 1.5x crit damage multiplier
	playerStats.Evasion = 0.05    // 5% evasion chance
	// Resistances default to 0.0 (handled by NewStatsComponent)
	player.AddComponent(playerStats)

	// Add player experience/progression
	playerExp := engine.NewExperienceComponent()
	player.AddComponent(playerExp)

	// Add player inventory
	playerInventory := engine.NewInventoryComponent(20, 100.0) // 20 items, 100 weight max
	playerInventory.Gold = 100
	player.AddComponent(playerInventory)

	// Add player equipment
	playerEquipment := engine.NewEquipmentComponent()
	player.AddComponent(playerEquipment)

	// GAP-002 REPAIR: Add mana and spells
	playerMana := &engine.ManaComponent{
		Current: 100,
		Max:     100,
		Regen:   5.0, // 5 mana per second
	}
	player.AddComponent(playerMana)

	// Load procedurally generated spells
	err = engine.LoadPlayerSpells(player, *seed, *genreID, 1)
	if err != nil {
		log.Fatalf("Failed to load player spells: %v", err)
	}
	if *verbose {
		log.Println("Player spells loaded (keys 1-5)")
	}

	// Load procedurally generated skill tree
	err = engine.LoadPlayerSkillTree(player, *seed, *genreID, 0)
	if err != nil {
		log.Fatalf("Failed to load skill tree: %v", err)
	}
	if *verbose {
		comp, _ := player.GetComponent("skill_tree")
		if comp != nil {
			treeComp := comp.(*engine.SkillTreeComponent)
			log.Printf("Skill tree '%s' loaded with %d skills (press K)",
				treeComp.Tree.Name, len(treeComp.Tree.Nodes))
		}
	}

	// Add quest tracker
	questTracker := engine.NewQuestTrackerComponent(5) // Max 5 active quests
	player.AddComponent(questTracker)

	// Add player attack capability
	player.AddComponent(&engine.AttackComponent{
		Damage:     15,
		DamageType: combat.DamagePhysical,
		Range:      50,
		Cooldown:   0.5,
	})

	// Add collision for player (28x28 to fit through 32px corridors)
	player.AddComponent(&engine.ColliderComponent{
		Width:     28,
		Height:    28,
		Solid:     true,
		IsTrigger: false,
		Layer:     1,
		OffsetX:   -14, // Center the collider (28/2 = 14)
		OffsetY:   -14,
	})

	// GAP-012 REPAIR: Add visual feedback for hit flash
	player.AddComponent(engine.NewVisualFeedbackComponent())

	clientLogger.WithField("entityID", player.ID).Info("player entity created")

	// Add starter items to inventory
	clientLogger.Info("adding starter items to inventory")
	addStarterItems(playerInventory, *seed, *genreID, logger)

	// Add tutorial quest
	clientLogger.Info("creating tutorial quest")
	addTutorialQuest(questTracker, *seed, *genreID, logger)

	// Initialize save/load system (Phase 8.4)
	clientLogger.Info("initializing save/load system")

	saveManager, err := saveload.NewSaveManager("./saves")
	if err != nil {
		log.Printf("Warning: Failed to initialize save manager: %v", err)
		log.Println("Save/load functionality will be unavailable")
	} else {
		if *verbose {
			log.Println("Save/load system initialized")
		}

		// Setup quick save callback (F5)
		inputSystem.SetQuickSaveCallback(func() error {
			log.Println("Quick save (F5 pressed)...")

			// Get player position
			var posX, posY float64
			if posComp, ok := player.GetComponent("position"); ok {
				pos := posComp.(*engine.PositionComponent)
				posX, posY = pos.X, pos.Y
			}

			// Get player health
			var currentHealth, maxHealth float64
			if healthComp, ok := player.GetComponent("health"); ok {
				health := healthComp.(*engine.HealthComponent)
				currentHealth, maxHealth = health.Current, health.Max
			}

			// Get player stats
			var attack, defense, magic float64
			if statsComp, ok := player.GetComponent("stats"); ok {
				stats := statsComp.(*engine.StatsComponent)
				attack, defense, magic = stats.Attack, stats.Defense, stats.MagicPower
			} // Get player level and XP
			var level int
			var currentXP int64
			if expComp, ok := player.GetComponent("experience"); ok {
				exp := expComp.(*engine.ExperienceComponent)
				level, currentXP = exp.Level, int64(exp.CurrentXP)
			}

			// Get inventory data (store only item IDs for now)
			var inventoryItems []uint64
			var gold int
			itemDataList := make([]saveload.ItemData, 0)
			if invComp, ok := player.GetComponent("inventory"); ok {
				inv := invComp.(*engine.InventoryComponent)
				gold = inv.Gold // GAP-009: Save gold
				// GAP-007: Serialize full item data
				for _, itm := range inv.Items {
					itemDataList = append(itemDataList, saveload.ItemToData(itm))
				}
			}

			// GAP-008: Serialize equipped items
			var equippedItems saveload.EquipmentData
			if equip, hasEquip := player.GetComponent("equipment"); hasEquip {
				equipment := equip.(*engine.EquipmentComponent)
				// Check main hand for weapon
				if weapon := equipment.Slots[engine.SlotMainHand]; weapon != nil {
					weaponData := saveload.ItemToData(weapon)
					equippedItems.Weapon = &weaponData
				}
				// Check chest for armor (primary armor slot)
				if armor := equipment.Slots[engine.SlotChest]; armor != nil {
					armorData := saveload.ItemToData(armor)
					equippedItems.Armor = &armorData
				}
				// Check accessory slots
				if accessory := equipment.Slots[engine.SlotAccessory1]; accessory != nil {
					accessoryData := saveload.ItemToData(accessory)
					equippedItems.Accessory = &accessoryData
				}
			}

			// Serialize mana
			var currentMana, maxMana int
			if manaComp, hasMana := player.GetComponent("mana"); hasMana {
				mana := manaComp.(*engine.ManaComponent)
				currentMana = mana.Current
				maxMana = mana.Max
			}

			// Serialize spells
			spellDataList := make([]saveload.SpellData, 0)
			if slotsComp, hasSlots := player.GetComponent("spell_slots"); hasSlots {
				slots := slotsComp.(*engine.SpellSlotComponent)
				for i := 0; i < 5; i++ {
					if spell := slots.GetSlot(i); spell != nil {
						spellDataList = append(spellDataList, saveload.SpellToData(spell))
					}
				}
			}

			// GAP-005 REPAIR: Serialize fog of war exploration state
			var fogOfWar [][]bool
			if game.MapUI != nil {
				fogOfWar = game.MapUI.GetFogOfWar()
				if *verbose {
					log.Printf("Serializing fog of war: %dx%d", len(fogOfWar), func() int {
						if len(fogOfWar) > 0 {
							return len(fogOfWar[0])
						}
						return 0
					}())
				}
			}

			// GAP-003 REPAIR: Export tutorial state
			var tutorialStateData *saveload.TutorialStateData
			if game.TutorialSystem != nil {
				enabled, showUI, currentStep, completed := game.TutorialSystem.ExportState()
				tutorialStateData = &saveload.TutorialStateData{
					Enabled:        enabled,
					ShowUI:         showUI,
					CurrentStepIdx: currentStep,
					CompletedSteps: completed,
				}
			}

			// Create game save
			gameSave := &saveload.GameSave{
				Version: saveload.SaveVersion,
				PlayerState: &saveload.PlayerState{
					EntityID:       player.ID,
					X:              posX,
					Y:              posY,
					CurrentHealth:  currentHealth,
					MaxHealth:      maxHealth,
					Level:          level,
					Experience:     int(currentXP),
					Attack:         attack,
					Defense:        defense,
					MagicPower:     magic,
					Speed:          1.0,
					InventoryItems: inventoryItems, // Keep for backward compatibility
					Items:          itemDataList,   // GAP-007: Full item data
					Gold:           gold,           // GAP-009: Gold persistence
					EquippedItems:  equippedItems,  // GAP-008: Equipment persistence
					CurrentMana:    currentMana,
					MaxMana:        maxMana,
					Spells:         spellDataList,
					TutorialState:  tutorialStateData, // GAP-003 REPAIR: Tutorial persistence
				},
				WorldState: &saveload.WorldState{
					Seed:       *seed,
					GenreID:    *genreID,
					Width:      generatedTerrain.Width,
					Height:     generatedTerrain.Height,
					Difficulty: 0.5,
					Depth:      1,
					FogOfWar:   fogOfWar, // GAP-005: Fog of war persistence
				},
				Settings: &saveload.GameSettings{
					ScreenWidth:  *width,
					ScreenHeight: *height,
					Fullscreen:   false,
					VSync:        true,
					MasterVolume: 1.0,
					MusicVolume:  0.7,
					SFXVolume:    0.8,
					KeyBindings:  make(map[string]string),
				},
			}

			if err := saveManager.SaveGame("quicksave", gameSave); err != nil {
				log.Printf("Failed to save game: %v", err)
				return err
			}

			log.Println("Game saved successfully!")
			return nil
		})

		// Setup quick load callback (F9)
		inputSystem.SetQuickLoadCallback(func() error {
			log.Println("Quick load (F9 pressed)...")

			gameSave, err := saveManager.LoadGame("quicksave")
			if err != nil {
				log.Printf("Failed to load game: %v", err)
				return err
			}

			// Restore player position
			if posComp, ok := player.GetComponent("position"); ok {
				pos := posComp.(*engine.PositionComponent)
				pos.X = gameSave.PlayerState.X
				pos.Y = gameSave.PlayerState.Y
			}

			// Restore player health
			if healthComp, ok := player.GetComponent("health"); ok {
				health := healthComp.(*engine.HealthComponent)
				health.Current = gameSave.PlayerState.CurrentHealth
				health.Max = gameSave.PlayerState.MaxHealth
			}

			// Restore player stats
			if statsComp, ok := player.GetComponent("stats"); ok {
				stats := statsComp.(*engine.StatsComponent)
				stats.Attack = gameSave.PlayerState.Attack
				stats.Defense = gameSave.PlayerState.Defense
				stats.MagicPower = gameSave.PlayerState.MagicPower
			}

			// Restore player level and XP
			if expComp, ok := player.GetComponent("experience"); ok {
				exp := expComp.(*engine.ExperienceComponent)
				exp.Level = gameSave.PlayerState.Level
				exp.CurrentXP = gameSave.PlayerState.Experience
				// Note: RequiredXP is recalculated by progression system
			}

			// Restore inventory (simplified)
			if invComp, ok := player.GetComponent("inventory"); ok {
				inv := invComp.(*engine.InventoryComponent)

				// GAP-007: Restore full inventory items
				inv.Items = make([]*item.Item, 0, len(gameSave.PlayerState.Items))
				for _, itemData := range gameSave.PlayerState.Items {
					restoredItem := saveload.DataToItem(itemData)
					inv.Items = append(inv.Items, restoredItem)
				}

				// GAP-009: Restore gold
				inv.Gold = gameSave.PlayerState.Gold

				if *verbose {
					log.Printf("Restored %d items and %d gold", len(inv.Items), inv.Gold)
				}
			}

			// GAP-008: Restore equipped items
			if equipComp, ok := player.GetComponent("equipment"); ok {
				equipment := equipComp.(*engine.EquipmentComponent)

				// Clear existing equipment
				equipment.Slots = make(map[engine.EquipmentSlot]*item.Item)

				// Restore weapon
				if gameSave.PlayerState.EquippedItems.Weapon != nil {
					weapon := saveload.DataToItem(*gameSave.PlayerState.EquippedItems.Weapon)
					equipment.Slots[engine.SlotMainHand] = weapon
				}

				// Restore armor
				if gameSave.PlayerState.EquippedItems.Armor != nil {
					armor := saveload.DataToItem(*gameSave.PlayerState.EquippedItems.Armor)
					equipment.Slots[engine.SlotChest] = armor
				}

				// Restore accessory
				if gameSave.PlayerState.EquippedItems.Accessory != nil {
					accessory := saveload.DataToItem(*gameSave.PlayerState.EquippedItems.Accessory)
					equipment.Slots[engine.SlotAccessory1] = accessory
				}

				equipment.StatsDirty = true // Trigger stats recalculation
			}

			// Restore mana
			if manaComp, ok := player.GetComponent("mana"); ok {
				mana := manaComp.(*engine.ManaComponent)
				mana.Current = gameSave.PlayerState.CurrentMana
				mana.Max = gameSave.PlayerState.MaxMana
			}

			// Restore spells
			if slotsComp, ok := player.GetComponent("spell_slots"); ok {
				slots := slotsComp.(*engine.SpellSlotComponent)

				// Clear existing spells
				for i := 0; i < 5; i++ {
					slots.Slots[i] = nil
				}

				// Restore saved spells
				for i, spellData := range gameSave.PlayerState.Spells {
					if i < 5 {
						restoredSpell := saveload.DataToSpell(spellData)
						slots.SetSlot(i, restoredSpell)
					}
				}
			}

			// GAP-005 REPAIR: Restore fog of war exploration state
			if game.MapUI != nil && gameSave.WorldState != nil && gameSave.WorldState.FogOfWar != nil {
				game.MapUI.SetFogOfWar(gameSave.WorldState.FogOfWar)
				if *verbose {
					fogData := gameSave.WorldState.FogOfWar
					log.Printf("Restored fog of war: %dx%d", len(fogData), func() int {
						if len(fogData) > 0 {
							return len(fogData[0])
						}
						return 0
					}())
				}
			}

			// GAP-003 REPAIR: Restore tutorial state
			if game.TutorialSystem != nil && gameSave.PlayerState.TutorialState != nil {
				tutState := gameSave.PlayerState.TutorialState
				game.TutorialSystem.ImportState(
					tutState.Enabled,
					tutState.ShowUI,
					tutState.CurrentStepIdx,
					tutState.CompletedSteps,
				)
				if *verbose {
					log.Printf("Restored tutorial state: enabled=%v, step=%d/%d",
						tutState.Enabled, tutState.CurrentStepIdx, len(game.TutorialSystem.Steps))
				}
			}

			log.Println("Game loaded successfully!")
			return nil
		})

		if *verbose {
			log.Println("Quick save/load callbacks registered (F5/F9)")
		}
	}

	// Connect inventory system to UI for item actions
	game.SetInventorySystem(inventorySystem)

	// Setup UI input callbacks
	if *verbose {
		log.Println("Setting up UI input callbacks...")
	}
	// GAP-014 REPAIR: Pass objective tracker to enable tutorial quest tracking
	game.SetupInputCallbacks(inputSystem, objectiveTracker)
	if *verbose {
		log.Println("UI callbacks registered (I: Inventory, J: Quests, ESC: Pause Menu)")
		log.Println("Inventory actions: E to equip/use, D to drop")
	}

	// Connect save/load callbacks to menu system
	if game.MenuSystem != nil && saveManager != nil {
		if *verbose {
			log.Println("Connecting save/load callbacks to menu system...")
		}

		// Create save callback that reuses the quick save logic
		saveCallback := func(saveName string) error {
			if *verbose {
				log.Printf("Menu save to '%s'...", saveName)
			}

			// Get player position
			var posX, posY float64
			if posComp, ok := player.GetComponent("position"); ok {
				pos := posComp.(*engine.PositionComponent)
				posX, posY = pos.X, pos.Y
			}

			// Get player health
			var currentHealth, maxHealth float64
			if healthComp, ok := player.GetComponent("health"); ok {
				health := healthComp.(*engine.HealthComponent)
				currentHealth, maxHealth = health.Current, health.Max
			}

			// Get player stats
			var attack, defense, magic float64
			if statsComp, ok := player.GetComponent("stats"); ok {
				stats := statsComp.(*engine.StatsComponent)
				attack, defense, magic = stats.Attack, stats.Defense, stats.MagicPower
			}

			// Get player level and XP
			var level int
			var currentXP int64
			if expComp, ok := player.GetComponent("experience"); ok {
				exp := expComp.(*engine.ExperienceComponent)
				level, currentXP = exp.Level, int64(exp.CurrentXP)
			}

			// Get inventory data
			var items []saveload.ItemData
			var gold int
			if invComp, ok := player.GetComponent("inventory"); ok {
				inv := invComp.(*engine.InventoryComponent)
				gold = inv.Gold

				// Convert items to ItemData for persistence
				for _, itm := range inv.Items {
					items = append(items, saveload.ItemData{
						Name:           itm.Name,
						Type:           itm.Type.String(),
						WeaponType:     itm.WeaponType.String(),
						ArmorType:      itm.ArmorType.String(),
						ConsumableType: itm.ConsumableType.String(),
						Rarity:         itm.Rarity.String(),
						Seed:           itm.Seed,
						Tags:           itm.Tags,
						Description:    itm.Description,
						Damage:         itm.Stats.Damage,
						Defense:        itm.Stats.Defense,
						AttackSpeed:    itm.Stats.AttackSpeed,
						Value:          itm.Stats.Value,
						Weight:         itm.Stats.Weight,
						RequiredLevel:  itm.Stats.RequiredLevel,
						DurabilityMax:  itm.Stats.DurabilityMax,
						Durability:     itm.Stats.Durability,
					})
				}
			}

			// Create game save
			gameSave := &saveload.GameSave{
				Version: saveload.SaveVersion,
				PlayerState: &saveload.PlayerState{
					EntityID:      player.ID,
					X:             posX,
					Y:             posY,
					CurrentHealth: currentHealth,
					MaxHealth:     maxHealth,
					Level:         level,
					Experience:    int(currentXP),
					Attack:        attack,
					Defense:       defense,
					MagicPower:    magic,
					Speed:         1.0,
					Items:         items, // Use new Items field instead of InventoryItems
					Gold:          gold,
				},
				WorldState: &saveload.WorldState{
					Seed:       *seed,
					GenreID:    *genreID,
					Width:      generatedTerrain.Width,
					Height:     generatedTerrain.Height,
					Difficulty: 0.5,
					Depth:      1,
				},
				Settings: &saveload.GameSettings{
					ScreenWidth:  *width,
					ScreenHeight: *height,
					Fullscreen:   false,
					VSync:        true,
					MasterVolume: 1.0,
					MusicVolume:  0.7,
					SFXVolume:    0.8,
					KeyBindings:  make(map[string]string),
				},
			}

			if err := saveManager.SaveGame(saveName, gameSave); err != nil {
				log.Printf("Failed to save game to '%s': %v", saveName, err)
				return err
			}

			log.Printf("Game saved successfully to '%s'!", saveName)
			return nil
		}

		// Create load callback that reuses the quick load logic
		loadCallback := func(saveName string) error {
			if *verbose {
				log.Printf("Menu load from '%s'...", saveName)
			}

			gameSave, err := saveManager.LoadGame(saveName)
			if err != nil {
				log.Printf("Failed to load game from '%s': %v", saveName, err)
				return err
			}

			// Restore player position
			if posComp, ok := player.GetComponent("position"); ok {
				pos := posComp.(*engine.PositionComponent)
				pos.X = gameSave.PlayerState.X
				pos.Y = gameSave.PlayerState.Y
			}

			// Restore player health
			if healthComp, ok := player.GetComponent("health"); ok {
				health := healthComp.(*engine.HealthComponent)
				health.Current = gameSave.PlayerState.CurrentHealth
				health.Max = gameSave.PlayerState.MaxHealth
			}

			// Restore player stats
			if statsComp, ok := player.GetComponent("stats"); ok {
				stats := statsComp.(*engine.StatsComponent)
				stats.Attack = gameSave.PlayerState.Attack
				stats.Defense = gameSave.PlayerState.Defense
				stats.MagicPower = gameSave.PlayerState.MagicPower
			}

			// Restore player level and XP
			if expComp, ok := player.GetComponent("experience"); ok {
				exp := expComp.(*engine.ExperienceComponent)
				exp.Level = gameSave.PlayerState.Level
				exp.CurrentXP = gameSave.PlayerState.Experience
			}

			log.Printf("Game loaded successfully from '%s'!", saveName)
			return nil
		}

		// Connect callbacks to menu system
		game.MenuSystem.SetSaveCallback(saveCallback)
		game.MenuSystem.SetLoadCallback(loadCallback)

		if *verbose {
			log.Println("Save/load callbacks connected to menu system")
		}
	}

	// Process initial entity additions
	game.World.Update(0)

	log.Println("Game initialized successfully")
	log.Printf("Controls: WASD to move, Space to attack, E to use item, I: Inventory, J: Quests")
	log.Printf("Genre: %s, Seed: %d", *genreID, *seed)
	if *multiplayer {
		log.Printf("Multiplayer: Connected to %s", *server)
	}

	// Setup cleanup handler for network client
	defer func() {
		if networkClient != nil {
			log.Println("Disconnecting from server...")
			if err := networkClient.Disconnect(); err != nil {
				log.Printf("Error disconnecting: %v", err)
			}
		}
	}()

	// Run the game loop
	if err := game.Run("Venture - Procedural Action RPG"); err != nil {
		log.Fatalf("Error running game: %v", err)
	}
}
