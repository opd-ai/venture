package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/opd-ai/venture/pkg/combat"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/network"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/quest"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/saveload"
)

var (
	width       = flag.Int("width", 800, "Screen width")
	height      = flag.Int("height", 600, "Screen height")
	seed        = flag.Int64("seed", 12345, "World generation seed")
	genreID     = flag.String("genre", "fantasy", "Genre ID (fantasy, scifi, horror, cyberpunk, postapoc)")
	verbose     = flag.Bool("verbose", false, "Enable verbose logging")
	multiplayer = flag.Bool("multiplayer", false, "Enable multiplayer mode (connect to server)")
	server      = flag.String("server", "localhost:8080", "Server address (host:port) for multiplayer")
)

// addStarterItems generates and adds starting items to the player's inventory.
func addStarterItems(inventory *engine.InventoryComponent, seed int64, genreID string, verbose bool) {
	itemGen := item.NewItemGenerator()

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
		log.Printf("Warning: Failed to generate starter weapon: %v", err)
	} else {
		weapons := weaponResult.([]*item.Item)
		if len(weapons) > 0 {
			weapon := weapons[0]
			weapon.Name = "Rusty " + weapon.Name // Make it clearly a starter item
			weapon.Stats.Value = 5               // Low value
			inventory.Items = append(inventory.Items, weapon)
			if verbose {
				log.Printf("Added starter weapon: %s (Damage: %d)", weapon.Name, weapon.Stats.Damage)
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
		log.Printf("Warning: Failed to generate healing potions: %v", err)
	} else {
		potions := potionResult.([]*item.Item)
		for i, potion := range potions {
			potion.Name = "Minor Health Potion"
			potion.Stats.Value = 10
			potion.Stats.Weight = 0.2
			inventory.Items = append(inventory.Items, potion)
			if verbose && i == 0 {
				log.Printf("Added %d healing potions", len(potions))
			}
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
		log.Printf("Warning: Failed to generate starter armor: %v", err)
	} else {
		armors := armorResult.([]*item.Item)
		if len(armors) > 0 {
			armor := armors[0]
			armor.Name = "Worn " + armor.Name
			armor.Stats.Value = 8
			inventory.Items = append(inventory.Items, armor)
			if verbose {
				log.Printf("Added starter armor: %s (Defense: %d)", armor.Name, armor.Stats.Defense)
			}
		}
	}

	if verbose {
		log.Printf("Starter items added: %d items in inventory", len(inventory.Items))
	}
}

// addTutorialQuest creates and adds a tutorial quest to the player's quest tracker.
func addTutorialQuest(tracker *engine.QuestTrackerComponent, seed int64, genreID string, verbose bool) {
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

	if verbose {
		log.Printf("Tutorial quest added: '%s' with %d objectives", tutorialQuest.Name, len(tutorialQuest.Objectives))
	}
}

func main() {
	flag.Parse()

	log.Printf("Starting Venture - Procedural Action RPG")
	log.Printf("Screen: %dx%d, Seed: %d, Genre: %s", *width, *height, *seed, *genreID)

	// Initialize network client if multiplayer mode is enabled
	var networkClient *network.Client
	if *multiplayer {
		log.Printf("Multiplayer mode enabled - connecting to server at %s", *server)

		clientConfig := network.DefaultClientConfig()
		clientConfig.ServerAddress = *server
		networkClient = network.NewClient(clientConfig)

		// Connect to server
		if err := networkClient.Connect(); err != nil {
			log.Fatalf("Failed to connect to server: %v", err)
		}

		log.Printf("Connected to server successfully")

		// Handle network errors in background
		go func() {
			for err := range networkClient.ReceiveError() {
				log.Printf("Network error: %v", err)
			}
		}()

		if *verbose {
			log.Println("Network client initialized and connected")
		}
	} else {
		log.Println("Single-player mode (use -multiplayer flag to connect to server)")
	}

	// Create the game instance
	game := engine.NewGame(*width, *height)

	// Initialize game systems
	if *verbose {
		log.Println("Initializing game systems...")
	}

	// Add core gameplay systems
	inputSystem := engine.NewInputSystem()
	movementSystem := &engine.MovementSystem{}
	collisionSystem := &engine.CollisionSystem{}
	combatSystem := engine.NewCombatSystem(*seed)
	aiSystem := engine.NewAISystem(game.World)
	progressionSystem := engine.NewProgressionSystem(game.World)
	inventorySystem := engine.NewInventorySystem(game.World)

	// GAP #2 REPAIR: Add player combat system to connect Space key to combat
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
	// 2. Player Combat/Item Use - processes input flags
	// 3. Movement - applies velocity to position
	// 4. Collision - checks and resolves collisions
	// 5. Combat - handles damage/status effects
	// 6. AI - enemy decision-making
	// 7. Progression - XP and leveling
	// 8. Inventory - item management
	// 9. Tutorial/Help - UI overlays
	game.World.AddSystem(inputSystem)
	game.World.AddSystem(playerCombatSystem)
	game.World.AddSystem(playerItemUseSystem)
	game.World.AddSystem(movementSystem)
	game.World.AddSystem(collisionSystem)
	game.World.AddSystem(combatSystem)
	game.World.AddSystem(aiSystem)
	game.World.AddSystem(progressionSystem)
	game.World.AddSystem(inventorySystem)
	game.World.AddSystem(tutorialSystem)
	game.World.AddSystem(helpSystem)

	// Store references to tutorial and help systems in game for rendering
	game.TutorialSystem = tutorialSystem
	game.HelpSystem = helpSystem

	if *verbose {
		log.Println("Systems initialized: Input, PlayerCombat, PlayerItemUse, Movement, Collision, Combat, AI, Progression, Inventory, Tutorial, Help")
	}

	// Gap #3: Initialize performance monitoring (wraps World.Update)
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
	if *verbose {
		log.Println("Generating procedural terrain...")
	}

	terrainGen := terrain.NewBSPGenerator() // Use BSP algorithm
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
		log.Fatalf("Failed to generate terrain: %v", err)
	}

	generatedTerrain := terrainResult.(*terrain.Terrain)
	if *verbose {
		log.Printf("Terrain generated: %dx%d with %d rooms",
			generatedTerrain.Width, generatedTerrain.Height, len(generatedTerrain.Rooms))
	}

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

	// Add player components
	player.AddComponent(&engine.PositionComponent{X: 400, Y: 300})
	player.AddComponent(&engine.VelocityComponent{VX: 0, VY: 0})
	player.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
	player.AddComponent(&engine.TeamComponent{TeamID: 1}) // Player team

	// Add input component for player control
	player.AddComponent(&engine.InputComponent{})

	// Add sprite for rendering
	playerSprite := engine.NewSpriteComponent(32, 32, color.RGBA{100, 150, 255, 255})
	playerSprite.Layer = 10 // Draw player on top
	player.AddComponent(playerSprite)

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

	// Add collision for player
	player.AddComponent(&engine.ColliderComponent{
		Width:     32,
		Height:    32,
		Solid:     true,
		IsTrigger: false,
		Layer:     1,
		OffsetX:   -16, // Center the collider
		OffsetY:   -16,
	})

	if *verbose {
		log.Printf("Player entity created (ID: %d) at position (400, 300)", player.ID)
	}

	// Add starter items to inventory
	if *verbose {
		log.Println("Adding starter items to inventory...")
	}
	addStarterItems(playerInventory, *seed, *genreID, *verbose)

	// Add tutorial quest
	if *verbose {
		log.Println("Creating tutorial quest...")
	}
	addTutorialQuest(questTracker, *seed, *genreID, *verbose)

	// Initialize save/load system (Phase 8.4)
	if *verbose {
		log.Println("Initializing save/load system...")
	}

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
			if invComp, ok := player.GetComponent("inventory"); ok {
				inv := invComp.(*engine.InventoryComponent)
				_ = inv.Gold // We have gold but don't store it separately in PlayerState yet
				// Store item IDs (simplified - full serialization would need entity ID mapping)
				for range inv.Items {
					// TODO: Map items to entity IDs for proper persistence
					// For now, we'll skip this as it requires additional entity-item mapping
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
					Speed:          1.0, // Default speed
					InventoryItems: inventoryItems,
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
				// Note: Full item restoration would require recreating item objects
				// from stored inventory item IDs
				_ = invComp
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
	game.SetupInputCallbacks(inputSystem)
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

			// Get inventory data (store only item IDs for now)
			var inventoryItems []uint64
			if invComp, ok := player.GetComponent("inventory"); ok {
				inv := invComp.(*engine.InventoryComponent)
				_ = inv.Gold
				for range inv.Items {
					// TODO: Map items to entity IDs for proper persistence
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
					InventoryItems: inventoryItems,
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
