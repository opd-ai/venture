package main

import (
	"flag"
	"image/color"
	"log"

	"github.com/opd-ai/venture/pkg/combat"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
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

	game.World.AddSystem(inputSystem)
	game.World.AddSystem(movementSystem)
	game.World.AddSystem(collisionSystem)
	game.World.AddSystem(combatSystem)
	game.World.AddSystem(aiSystem)
	game.World.AddSystem(progressionSystem)
	game.World.AddSystem(inventorySystem)

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
