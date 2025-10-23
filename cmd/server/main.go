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

	// Create network server
	server := network.NewServer(serverConfig)

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

	// Start network server
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start network server: %v", err)
	}

	log.Println("Server initialized successfully")
	log.Printf("Server listening on port %s", *port)
	log.Printf("Max players: %d, Update rate: %d Hz", *maxPlayers, *tickRate)
	log.Printf("Game world ready with %d entities", len(world.GetEntities()))

	// Handle server shutdown gracefully
	defer func() {
		if err := server.Stop(); err != nil {
			log.Printf("Error stopping server: %v", err)
		}
	}()

	// Run authoritative game loop
	tickDuration := time.Duration(1000000000 / *tickRate) // nanoseconds per tick
	ticker := time.NewTicker(tickDuration)
	defer ticker.Stop()

	lastUpdate := time.Now()

	log.Printf("Starting authoritative game loop at %d Hz...", *tickRate)

	// Handle server errors in background
	go func() {
		for err := range server.ReceiveError() {
			log.Printf("Network error: %v", err)
		}
	}()

	// Handle client input commands in background
	go func() {
		for cmd := range server.ReceiveInputCommand() {
			// TODO: Process player input commands
			// For now, just log them in verbose mode
			if *verbose {
				log.Printf("Received input from player %d: type=%s, seq=%d",
					cmd.PlayerID, cmd.InputType, cmd.SequenceNumber)
			}
		}
	}()

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

			// Broadcast state to connected clients
			stateUpdate := convertSnapshotToStateUpdate(snapshot)
			server.BroadcastStateUpdate(stateUpdate)

			if *verbose && int(now.Unix())%10 == 0 {
				// Log every 10 seconds
				playerCount := server.GetPlayerCount()
				log.Printf("Server tick: %d entities, %d players connected",
					len(world.GetEntities()), playerCount)
			}
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

// convertSnapshotToStateUpdate converts a WorldSnapshot to a StateUpdate for broadcasting
func convertSnapshotToStateUpdate(snapshot network.WorldSnapshot) *network.StateUpdate {
	// For now, create a simple state update
	// In a full implementation, this would serialize component data efficiently
	update := &network.StateUpdate{
		Timestamp: uint64(snapshot.Timestamp.UnixNano() / 1000000), // milliseconds
		Priority:  128,                                             // Normal priority
	}
	return update
}
