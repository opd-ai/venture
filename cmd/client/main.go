package main

import (
	"flag"
	"image/color"
	"log"

	"github.com/opd-ai/venture/pkg/combat"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/saveload"
)

var (
	width   = flag.Int("width", 800, "Screen width")
	height  = flag.Int("height", 600, "Screen height")
	seed    = flag.Int64("seed", 12345, "World generation seed")
	genreID = flag.String("genre", "fantasy", "Genre ID (fantasy, scifi, horror, cyberpunk, postapoc)")
	verbose = flag.Bool("verbose", false, "Enable verbose logging")
)

func main() {
	flag.Parse()

	log.Printf("Starting Venture - Procedural Action RPG")
	log.Printf("Screen: %dx%d, Seed: %d, Genre: %s", *width, *height, *seed, *genreID)

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

	// Add tutorial and help systems (Phase 8.6)
	tutorialSystem := engine.NewTutorialSystem()
	helpSystem := engine.NewHelpSystem()

	// Connect help system to input system for ESC key handling
	inputSystem.SetHelpSystem(helpSystem)

	game.World.AddSystem(inputSystem)
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
		log.Println("Systems initialized: Input, Movement, Collision, Combat, AI, Progression, Inventory")
	}

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
				attack, defense, magic = stats.Attack, stats.Defense, stats.Magic
			}

			// Get player level and XP
			var level int
			var currentXP, xpToNext int64
			if expComp, ok := player.GetComponent("experience"); ok {
				exp := expComp.(*engine.ExperienceComponent)
				level, currentXP, xpToNext = exp.Level, exp.CurrentXP, exp.XPToNext
			}

			// Get inventory data
			var gold int
			var items []saveload.ItemData
			if invComp, ok := player.GetComponent("inventory"); ok {
				inv := invComp.(*engine.InventoryComponent)
				gold = inv.Gold
				// Convert inventory items to ItemData (simplified - would need full serialization)
				for _, item := range inv.Items {
					items = append(items, saveload.ItemData{
						Name:   item.Name,
						Type:   string(item.Type),
						Weight: item.Weight,
					})
				}
			}

			// Create game save
			gameSave := &saveload.GameSave{
				Player: saveload.PlayerState{
					Position: saveload.Position{X: posX, Y: posY},
					Health:   saveload.Health{Current: currentHealth, Max: maxHealth},
					Stats: saveload.Stats{
						Attack:  attack,
						Defense: defense,
						Magic:   magic,
					},
					Level:     level,
					CurrentXP: currentXP,
					XPToNext:  xpToNext,
					Inventory: saveload.Inventory{
						Items: items,
						Gold:  gold,
					},
				},
				World: saveload.WorldState{
					Seed:       *seed,
					Genre:      *genreID,
					Width:      generatedTerrain.Width,
					Height:     generatedTerrain.Height,
					Difficulty: 0.5,
				},
				Settings: saveload.GameSettings{
					ScreenWidth:  *width,
					ScreenHeight: *height,
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
				pos.X = gameSave.Player.Position.X
				pos.Y = gameSave.Player.Position.Y
			}

			// Restore player health
			if healthComp, ok := player.GetComponent("health"); ok {
				health := healthComp.(*engine.HealthComponent)
				health.Current = gameSave.Player.Health.Current
				health.Max = gameSave.Player.Health.Max
			}

			// Restore player stats
			if statsComp, ok := player.GetComponent("stats"); ok {
				stats := statsComp.(*engine.StatsComponent)
				stats.Attack = gameSave.Player.Stats.Attack
				stats.Defense = gameSave.Player.Stats.Defense
				stats.Magic = gameSave.Player.Stats.Magic
			}

			// Restore player level and XP
			if expComp, ok := player.GetComponent("experience"); ok {
				exp := expComp.(*engine.ExperienceComponent)
				exp.Level = gameSave.Player.Level
				exp.CurrentXP = gameSave.Player.CurrentXP
				exp.XPToNext = gameSave.Player.XPToNext
			}

			// Restore inventory (simplified)
			if invComp, ok := player.GetComponent("inventory"); ok {
				inv := invComp.(*engine.InventoryComponent)
				inv.Gold = gameSave.Player.Inventory.Gold
				// Note: Full item restoration would require recreating item objects
			}

			log.Println("Game loaded successfully!")
			return nil
		})

		if *verbose {
			log.Println("Quick save/load callbacks registered (F5/F9)")
		}
	}

	// Process initial entity additions
	game.World.Update(0)

	log.Println("Game initialized successfully")
	log.Printf("Controls: Arrow keys to move, Space to attack")
	log.Printf("Genre: %s, Seed: %d", *genreID, *seed)

	// Run the game loop
	if err := game.Run("Venture - Procedural Action RPG"); err != nil {
		log.Fatalf("Error running game: %v", err)
	}
}
