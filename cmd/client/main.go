package main

import (
	"flag"
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
	movementSystem := &engine.MovementSystem{}
	collisionSystem := &engine.CollisionSystem{}
	combatSystem := engine.NewCombatSystem(*seed)
	aiSystem := &engine.AISystem{}
	progressionSystem := &engine.ProgressionSystem{}
	inventorySystem := &engine.InventorySystem{}

	game.World.AddSystem(movementSystem)
	game.World.AddSystem(collisionSystem)
	game.World.AddSystem(combatSystem)
	game.World.AddSystem(aiSystem)
	game.World.AddSystem(progressionSystem)
	game.World.AddSystem(inventorySystem)

	if *verbose {
		log.Println("Systems initialized: Movement, Collision, Combat, AI, Progression, Inventory")
	}

	// Generate initial world terrain
	if *verbose {
		log.Println("Generating procedural terrain...")
	}

	terrainGen := terrain.NewTerrainGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    *genreID,
		Custom: map[string]interface{}{
			"width":     80,
			"height":    50,
			"algorithm": "bsp",
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
	player.AddComponent(&engine.VelocityComponent{X: 0, Y: 0})
	player.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
	player.AddComponent(&engine.TeamComponent{TeamID: 1}) // Player team

	// Add player stats
	playerStats := engine.NewStatsComponent()
	playerStats.Level = 1
	playerStats.Health = 100
	playerStats.Attack = 10
	playerStats.Defense = 5
	playerStats.Speed = 5.0
	player.AddComponent(playerStats)

	// Add player progression
	playerProgress := &engine.ProgressionComponent{
		Level:             1,
		ExperiencePoints:  0,
		ExperienceToLevel: 100,
		SkillPoints:       0,
		UnlockedSkills:    make([]string, 0),
	}
	player.AddComponent(playerProgress)

	// Add player inventory
	playerInventory := &engine.InventoryComponent{
		Items:    make([]engine.InventoryItem, 0),
		Capacity: 20,
		Gold:     100,
	}
	player.AddComponent(playerInventory)

	// Add player attack capability
	player.AddComponent(&engine.AttackComponent{
		Damage:     15,
		DamageType: combat.DamagePhysical,
		Range:      50,
		Cooldown:   0.5,
	})

	// Add collision for player
	player.AddComponent(&engine.CollisionComponent{
		Radius:    16,
		Mass:      1.0,
		IsTrigger: false,
		IsStatic:  false,
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
