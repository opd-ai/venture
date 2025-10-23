package main

import (
	"flag"
	"image/color"
	"log"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/combat"
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

	// Track player entities
	playerEntities := make(map[uint64]*engine.Entity)
	playerEntitiesMu := &sync.RWMutex{}

	// Handle new player connections in background
	go func() {
		for playerID := range server.ReceivePlayerJoin() {
			if *verbose {
				log.Printf("Player %d joined - creating player entity", playerID)
			}

			// Create player entity for new connection
			entity := createPlayerEntity(world, generatedTerrain, playerID, *seed, *genreID, *verbose)

			// Store player entity mapping
			playerEntitiesMu.Lock()
			playerEntities[playerID] = entity
			playerEntitiesMu.Unlock()

			if *verbose {
				log.Printf("Player %d entity created (ID: %d)", playerID, entity.ID)
			}
		}
	}()

	// Handle player disconnections in background
	go func() {
		for playerID := range server.ReceivePlayerLeave() {
			if *verbose {
				log.Printf("Player %d left - removing player entity", playerID)
			}

			// Remove player entity
			playerEntitiesMu.Lock()
			if entity, exists := playerEntities[playerID]; exists {
				world.RemoveEntity(entity.ID)
				delete(playerEntities, playerID)
				if *verbose {
					log.Printf("Player %d entity removed (ID: %d)", playerID, entity.ID)
				}
			}
			playerEntitiesMu.Unlock()
		}
	}()

	// Handle client input commands in background
	go func() {
		for cmd := range server.ReceiveInputCommand() {
			if *verbose {
				log.Printf("Received input from player %d: type=%s, seq=%d",
					cmd.PlayerID, cmd.InputType, cmd.SequenceNumber)
			}

			// Get player entity
			playerEntitiesMu.RLock()
			entity, exists := playerEntities[cmd.PlayerID]
			playerEntitiesMu.RUnlock()

			if !exists {
				if *verbose {
					log.Printf("Warning: No entity for player %d", cmd.PlayerID)
				}
				continue
			}

			// Apply input to entity
			applyInputCommand(entity, cmd, *verbose)
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

// createPlayerEntity creates a player entity for a connected client
func createPlayerEntity(world *engine.World, terrain *terrain.Terrain, playerID uint64, seed int64, genreID string, verbose bool) *engine.Entity {
	// Create player entity
	entity := world.CreateEntity()

	// Find valid spawn position in first room
	spawnX, spawnY := 400.0, 300.0 // Default spawn
	if len(terrain.Rooms) > 0 {
		room := terrain.Rooms[0]
		// Spawn in center of first room
		spawnX = float64(room.X+room.Width/2) * 32  // Convert to pixel coordinates (32px tiles)
		spawnY = float64(room.Y+room.Height/2) * 32
	}

	// Add core components
	entity.AddComponent(&engine.PositionComponent{X: spawnX, Y: spawnY})
	entity.AddComponent(&engine.VelocityComponent{VX: 0, VY: 0})
	entity.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&engine.TeamComponent{TeamID: 1}) // All players on team 1

	// Add network component to mark as networked entity
	entity.AddComponent(&engine.NetworkComponent{
		PlayerID: playerID,
		Synced:   true,
	})

	// Add sprite for rendering
	playerSprite := engine.NewSpriteComponent(32, 32, color.RGBA{100, 150, 255, 255})
	playerSprite.Layer = 10 // Draw players on top
	entity.AddComponent(playerSprite)

	// Add player stats
	playerStats := engine.NewStatsComponent()
	playerStats.Attack = 10
	playerStats.Defense = 5
	entity.AddComponent(playerStats)

	// Add player experience/progression
	playerExp := engine.NewExperienceComponent()
	entity.AddComponent(playerExp)

	// Add player inventory
	playerInventory := engine.NewInventoryComponent(20, 100.0) // 20 items, 100 weight max
	playerInventory.Gold = 100
	entity.AddComponent(playerInventory)

	// Add player equipment
	playerEquipment := engine.NewEquipmentComponent()
	entity.AddComponent(playerEquipment)

	// Add quest tracker
	questTracker := engine.NewQuestTrackerComponent(5) // Max 5 active quests
	entity.AddComponent(questTracker)

	// Add player attack capability
	entity.AddComponent(&engine.AttackComponent{
		Damage:     15,
		DamageType: combat.DamagePhysical,
		Range:      50,
		Cooldown:   0.5,
	})

	// Add collision for player
	entity.AddComponent(&engine.ColliderComponent{
		Width:     32,
		Height:    32,
		Solid:     true,
		IsTrigger: false,
		Layer:     1,
		OffsetX:   -16, // Center the collider
		OffsetY:   -16,
	})

	if verbose {
		log.Printf("Player entity created: ID=%d, PlayerID=%d, Position=(%.1f, %.1f)",
			entity.ID, playerID, spawnX, spawnY)
	}

	return entity
}

// applyInputCommand applies a network input command to a player entity
func applyInputCommand(entity *engine.Entity, cmd *network.InputCommand, verbose bool) {
	// Get velocity component
	velComp, hasVel := entity.GetComponent("velocity")
	if !hasVel {
		return
	}
	velocity := velComp.(*engine.VelocityComponent)

	// Process input based on type
	switch cmd.InputType {
	case "move":
		// Apply movement input to velocity
		if len(cmd.Data) >= 2 {
			moveX := float64(int8(cmd.Data[0])) // Convert byte to signed value (-128 to 127)
			moveY := float64(int8(cmd.Data[1]))

			// Normalize to -1.0 to 1.0 range
			moveX /= 127.0
			moveY /= 127.0

			// Normalize diagonal movement
			if moveX != 0 && moveY != 0 {
				moveX *= 0.707
				moveY *= 0.707
			}

			// Apply movement speed (100 pixels/second)
			velocity.VX = moveX * 100.0
			velocity.VY = moveY * 100.0

			if verbose && (moveX != 0 || moveY != 0) {
				log.Printf("Player %d moving: velocity=(%.1f, %.1f)",
					cmd.PlayerID, velocity.VX, velocity.VY)
			}
		}

	case "attack":
		// Trigger attack (future implementation)
		if verbose {
			log.Printf("Player %d attacking (not yet implemented)", cmd.PlayerID)
		}
		// TODO: Implement attack handling

	case "use_item":
		// Use item (future implementation)
		if verbose {
			log.Printf("Player %d using item (not yet implemented)", cmd.PlayerID)
		}
		// TODO: Implement item use handling

	default:
		if verbose {
			log.Printf("Unknown input type from player %d: %s", cmd.PlayerID, cmd.InputType)
		}
	}
}
