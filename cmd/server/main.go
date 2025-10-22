package main

import (
	"flag"
	"log"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/network"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

var (
	port       = flag.String("port", "8080", "Server port")
	maxPlayers = flag.Int("max-players", 4, "Maximum number of players")
	seed       = flag.Int64("seed", 12345, "World generation seed")
	genreID    = flag.String("genre", "fantasy", "Genre ID for world generation")
	tickRate   = flag.Int("tick-rate", 20, "Server update rate (updates per second)")
	verbose    = flag.Bool("verbose", false, "Enable verbose logging")
)

func main() {
	flag.Parse()

	log.Printf("Starting Venture Game Server")
	log.Printf("Port: %s, Max Players: %d, Tick Rate: %d Hz", *port, *maxPlayers, *tickRate)
	log.Printf("World Seed: %d, Genre: %s", *seed, *genreID)

	// Create game world
	if *verbose {
		log.Println("Creating game world...")
	}

	world := engine.NewWorld()

	// Add gameplay systems
	movementSystem := &engine.MovementSystem{}
	collisionSystem := &engine.CollisionSystem{}
	combatSystem := engine.NewCombatSystem(*seed)
	aiSystem := engine.NewAISystem(world)
	progressionSystem := engine.NewProgressionSystem(world)
	inventorySystem := engine.NewInventorySystem(world)

	world.AddSystem(movementSystem)
	world.AddSystem(collisionSystem)
	world.AddSystem(combatSystem)
	world.AddSystem(aiSystem)
	world.AddSystem(progressionSystem)
	world.AddSystem(inventorySystem)

	if *verbose {
		log.Println("Game systems initialized")
	}

	// Generate initial world terrain
	if *verbose {
		log.Println("Generating world terrain...")
	}

	terrainGen := terrain.NewBSPGenerator() // Use BSP algorithm
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    *genreID,
		Custom: map[string]interface{}{
			"width":  100,
			"height": 100,
		},
	}

	terrainResult, err := terrainGen.Generate(*seed, params)
	if err != nil {
		log.Fatalf("Failed to generate terrain: %v", err)
	}

	generatedTerrain := terrainResult.(*terrain.Terrain)
	if *verbose {
		log.Printf("World terrain generated: %dx%d with %d rooms",
			generatedTerrain.Width, generatedTerrain.Height, len(generatedTerrain.Rooms))
	}

	// Initialize network components
	if *verbose {
		log.Println("Initializing network systems...")
	}

	// Create server with configuration
	serverConfig := network.DefaultServerConfig()
	serverConfig.Address = ":" + *port
	serverConfig.MaxPlayers = *maxPlayers
	serverConfig.UpdateRate = *tickRate

	// Create snapshot manager for state synchronization
	snapshotManager := network.NewSnapshotManager(100)

	// Create lag compensator
	lagCompConfig := network.DefaultLagCompensationConfig()
	lagCompensator := network.NewLagCompensator(lagCompConfig)

	if *verbose {
		log.Println("Network systems initialized")
		log.Printf("Server config: Address=%s, MaxPlayers=%d, UpdateRate=%d Hz",
			serverConfig.Address, serverConfig.MaxPlayers, serverConfig.UpdateRate)
	}

	log.Println("Server initialized successfully")
	log.Printf("Server running on port %s (not accepting connections yet - network layer stub)", *port)
	log.Printf("Game world ready with %d entities", len(world.GetEntities()))

	// Run authoritative game loop
	tickDuration := time.Duration(1000000000 / *tickRate) // nanoseconds per tick
	ticker := time.NewTicker(tickDuration)
	defer ticker.Stop()

	lastUpdate := time.Now()

	log.Printf("Starting authoritative game loop at %d Hz...", *tickRate)

	for {
		select {
		case <-ticker.C:
			// Calculate delta time
			now := time.Now()
			deltaTime := now.Sub(lastUpdate).Seconds()
			lastUpdate = now

			// Update game world
			world.Update(deltaTime)

			// Record snapshot for lag compensation and state sync
			snapshot := buildWorldSnapshot(world, now)
			snapshotManager.AddSnapshot(snapshot)
			lagCompensator.RecordSnapshot(snapshot)

			if *verbose && int(now.Unix())%10 == 0 {
				// Log every 10 seconds
				log.Printf("Server tick: %d entities",
					len(world.GetEntities()))
			}

			// TODO: Broadcast state to connected clients (when network server is implemented)
		}
	}
}

// buildWorldSnapshot creates a network snapshot from the current world state
func buildWorldSnapshot(world *engine.World, timestamp time.Time) network.WorldSnapshot {
	snapshot := network.WorldSnapshot{
		Timestamp: timestamp,
		Entities:  make(map[uint64]network.EntitySnapshot),
	}

	// Convert world entities to network entity snapshots
	for _, entity := range world.GetEntities() {
		// Get position component
		if posComp, ok := entity.GetComponent("position"); ok {
			pos := posComp.(*engine.PositionComponent)

			// Get velocity if it exists
			velX, velY := 0.0, 0.0
			if velComp, ok := entity.GetComponent("velocity"); ok {
				vel := velComp.(*engine.VelocityComponent)
				velX = vel.VX
				velY = vel.VY
			}

			snapshot.Entities[entity.ID] = network.EntitySnapshot{
				EntityID: entity.ID,
				Position: network.Position{X: pos.X, Y: pos.Y},
				Velocity: network.Velocity{VX: velX, VY: velY},
			}
		}
	}

	return snapshot
}
