package main

import (
	"flag"
	"image/color"
	"os"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/combat"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/network"
	"github.com/opd-ai/venture/pkg/procgen"
	itemgen "github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

var (
	port          = flag.String("port", "8080", "Server port")
	maxPlayers    = flag.Int("max-players", 4, "Maximum number of players")
	seed          = flag.Int64("seed", 12345, "World generation seed")
	genreID       = flag.String("genre", "fantasy", "Genre ID for world generation")
	tickRate      = flag.Int("tick-rate", 20, "Server update rate (updates per second)")
	verbose       = flag.Bool("verbose", false, "Enable verbose logging")
	aerialSprites = flag.Bool("aerial-sprites", true, "Enable aerial-view perspective sprites for top-down gameplay")
)

func main() {
	flag.Parse()

	// Initialize structured logger with JSON format for server (log aggregation)
	logConfig := logging.Config{
		Level:       logging.InfoLevel,
		Format:      logging.JSONFormat, // Always JSON for server
		AddCaller:   true,
		EnableColor: false,
	}

	// Override log level from environment variable
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		logConfig.Level = logging.LogLevel(logLevel)
	} else if *verbose {
		logConfig.Level = logging.DebugLevel
	}

	logger := logging.NewLogger(logConfig)
	serverLogger := logger.WithFields(logrus.Fields{
		"component": "server",
		"seed":      *seed,
		"genre":     *genreID,
	})

	serverLogger.Info("Starting Venture Game Server")
	serverLogger.WithFields(logrus.Fields{
		"port":          *port,
		"maxPlayers":    *maxPlayers,
		"tickRate":      *tickRate,
		"seed":          *seed,
		"genre":         *genreID,
		"aerialSprites": *aerialSprites,
	}).Info("server configuration")

	// Create game world
	worldLogger := logger.WithFields(logrus.Fields{"system": "world"})
	if logger.GetLevel() >= logrus.DebugLevel {
		worldLogger.Debug("creating game world")
	}

	world := engine.NewWorldWithLogger(logger)

	// Add gameplay systems with proper constructors
	movementSystem := engine.NewMovementSystem(200.0)  // 200 units/second max speed
	collisionSystem := engine.NewCollisionSystem(64.0) // 64-unit grid cells for spatial partitioning
	combatSystem := engine.NewCombatSystemWithLogger(*seed, logger)
	aiSystem := engine.NewAISystem(world)
	progressionSystem := engine.NewProgressionSystem(world)
	inventorySystem := engine.NewInventorySystem(world)

	world.AddSystem(movementSystem)
	world.AddSystem(collisionSystem)
	world.AddSystem(combatSystem)
	world.AddSystem(aiSystem)
	world.AddSystem(progressionSystem)
	world.AddSystem(inventorySystem)

	if logger.GetLevel() >= logrus.DebugLevel {
		worldLogger.Debug("game systems initialized")
	}

	// Generate initial world terrain
	terrainLogger := logging.GeneratorLogger(logger, "terrain", *seed, *genreID)
	if logger.GetLevel() >= logrus.DebugLevel {
		terrainLogger.WithFields(logrus.Fields{
			"width":  100,
			"height": 100,
		}).Debug("generating world terrain")
	}

	terrainGen := terrain.NewBSPGeneratorWithLogger(logger) // Use BSP algorithm with logging
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
		serverLogger.WithError(err).Fatal("failed to generate terrain")
	}

	generatedTerrain := terrainResult.(*terrain.Terrain)
	terrainLogger.WithFields(logrus.Fields{
		"width":     generatedTerrain.Width,
		"height":    generatedTerrain.Height,
		"roomCount": len(generatedTerrain.Rooms),
	}).Info("world terrain generated")

	// Initialize network components
	networkLogger := logger.WithFields(logrus.Fields{"system": "network"})
	if logger.GetLevel() >= logrus.DebugLevel {
		networkLogger.Debug("initializing network systems")
	}

	// Create server with configuration
	serverConfig := network.DefaultServerConfig()
	serverConfig.Address = ":" + *port
	serverConfig.MaxPlayers = *maxPlayers
	serverConfig.UpdateRate = *tickRate

	// Create network server with logging
	server := network.NewServerWithLogger(serverConfig, logger)

	// Create snapshot manager for state synchronization
	snapshotManager := network.NewSnapshotManager(100)

	// Create lag compensator
	lagCompConfig := network.DefaultLagCompensationConfig()
	lagCompensator := network.NewLagCompensator(lagCompConfig)

	networkLogger.WithFields(logrus.Fields{
		"address":    serverConfig.Address,
		"maxPlayers": serverConfig.MaxPlayers,
		"updateRate": serverConfig.UpdateRate,
	}).Info("network systems initialized")

	// Start network server
	if err := server.Start(); err != nil {
		serverLogger.WithError(err).Fatal("failed to start network server")
	}

	serverLogger.Info("server initialized successfully")
	serverLogger.WithFields(logrus.Fields{
		"port":        *port,
		"maxPlayers":  *maxPlayers,
		"updateRate":  *tickRate,
		"entityCount": len(world.GetEntities()),
	}).Info("server listening")

	// Handle server shutdown gracefully
	defer func() {
		serverLogger.Info("shutting down server")
		if err := server.Stop(); err != nil {
			serverLogger.WithError(err).Error("error stopping server")
		}
	}()

	// Run authoritative game loop
	tickDuration := time.Duration(1000000000 / *tickRate) // nanoseconds per tick
	ticker := time.NewTicker(tickDuration)
	defer ticker.Stop()

	lastUpdate := time.Now()

	serverLogger.WithField("tickRate", *tickRate).Info("starting authoritative game loop")

	// Handle server errors in background
	go func() {
		for err := range server.ReceiveError() {
			networkLogger.WithError(err).Error("network error")
		}
	}()

	// Track player entities
	playerEntities := make(map[uint64]*engine.Entity)
	playerEntitiesMu := &sync.RWMutex{}

	// Handle new player connections in background
	go func() {
		for playerID := range server.ReceivePlayerJoin() {
			playerLogger := logging.NetworkLogger(logger, "", "connected").WithField("playerID", playerID)
			playerLogger.Info("player joined - creating entity")

			// Create player entity for new connection
			entity := createPlayerEntity(world, generatedTerrain, playerID, *seed, *genreID, *aerialSprites, logger)

			// Store player entity mapping
			playerEntitiesMu.Lock()
			playerEntities[playerID] = entity
			playerEntitiesMu.Unlock()

			playerLogger.WithField("entityID", entity.ID).Debug("player entity created")
		}
	}()

	// Handle player disconnections in background
	go func() {
		for playerID := range server.ReceivePlayerLeave() {
			playerLogger := logging.NetworkLogger(logger, "", "disconnected").WithField("playerID", playerID)
			playerLogger.Info("player left - removing entity")

			// Remove player entity
			playerEntitiesMu.Lock()
			if entity, exists := playerEntities[playerID]; exists {
				world.RemoveEntity(entity.ID)
				delete(playerEntities, playerID)
				playerLogger.WithField("entityID", entity.ID).Debug("player entity removed")
			}
			playerEntitiesMu.Unlock()
		}
	}()

	// Handle client input commands in background
	go func() {
		for cmd := range server.ReceiveInputCommand() {
			if logger.GetLevel() >= logrus.DebugLevel {
				networkLogger.WithFields(logrus.Fields{
					"playerID":       cmd.PlayerID,
					"inputType":      cmd.InputType,
					"sequenceNumber": cmd.SequenceNumber,
				}).Debug("received input command")
			}

			// Get player entity
			playerEntitiesMu.RLock()
			entity, exists := playerEntities[cmd.PlayerID]
			playerEntitiesMu.RUnlock()

			if !exists {
				if logger.GetLevel() >= logrus.WarnLevel {
					networkLogger.WithField("playerID", cmd.PlayerID).Warn("no entity for player")
				}
				continue
			}

			// Apply input to entity
			applyInputCommand(entity, cmd, logger)
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

			// Periodic server metrics logging
			if logger.GetLevel() >= logrus.DebugLevel && int(now.Unix())%10 == 0 {
				// Log every 10 seconds
				playerCount := server.GetPlayerCount()
				serverLogger.WithFields(logrus.Fields{
					"entityCount": len(world.GetEntities()),
					"playerCount": playerCount,
				}).Debug("server tick metrics")
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
func createPlayerEntity(world *engine.World, terrain *terrain.Terrain, playerID uint64, seed int64, genreID string, useAerialSprites bool, logger *logrus.Logger) *engine.Entity {
	// Create player entity
	entity := world.CreateEntity()

	// Find valid spawn position in first room
	spawnX, spawnY := 400.0, 300.0 // Default spawn
	if len(terrain.Rooms) > 0 {
		room := terrain.Rooms[0]
		// Spawn in center of first room
		spawnX = float64(room.X+room.Width/2) * 32 // Convert to pixel coordinates (32px tiles)
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

	// Add sprite for rendering (28x28 to fit through 32px corridors)
	var playerSprite *engine.EbitenSprite
	if useAerialSprites {
		// Generate procedural directional sprites with aerial-view perspective
		spriteGen := sprites.NewGenerator()
		config := sprites.Config{
			Width:   28,
			Height:  28,
			Seed:    seed + int64(playerID), // Unique seed per player
			GenreID: genreID,
			Type:    sprites.SpriteEntity,
			Custom: map[string]interface{}{
				"entityType": "humanoid",
				"useAerial":  true,
			},
		}

		directionalSprites, err := spriteGen.GenerateDirectionalSprites(config)
		if err != nil {
			logger.WithError(err).Warn("failed to generate directional sprites, using fallback")
			playerSprite = engine.NewSpriteComponent(28, 28, color.RGBA{100, 150, 255, 255})
		} else {
			// Create sprite component with initial down-facing direction
			// directionalSprites is map[int]*ebiten.Image with keys 0-3
			playerSprite = &engine.EbitenSprite{
				Image:             directionalSprites[int(engine.DirDown)],
				Width:             28,
				Height:            28,
				Visible:           true,
				Layer:             10,
				CurrentDirection:  int(engine.DirDown),
				DirectionalImages: directionalSprites, // Already map[int]*ebiten.Image
			}
			// Add animation component to enable automatic facing updates
			entity.AddComponent(&engine.AnimationComponent{
				Seed:         seed + int64(playerID),
				CurrentState: engine.AnimationStateIdle,
				Playing:      true,
			})
		}
	} else {
		// Use simple colored sprite for side-view
		playerSprite = engine.NewSpriteComponent(28, 28, color.RGBA{100, 150, 255, 255})
	}

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

	// Add collision for player (28x28 to fit through 32px corridors)
	entity.AddComponent(&engine.ColliderComponent{
		Width:     28,
		Height:    28,
		Solid:     true,
		IsTrigger: false,
		Layer:     1,
		OffsetX:   -14, // Center the collider (28/2 = 14)
		OffsetY:   -14,
	})

	if logger.GetLevel() >= logrus.DebugLevel {
		logging.EntityLogger(logger, int(entity.ID)).WithFields(logrus.Fields{
			"playerID": playerID,
			"x":        spawnX,
			"y":        spawnY,
		}).Debug("player entity created")
	}

	return entity
}

// applyInputCommand applies a network input command to a player entity
func applyInputCommand(entity *engine.Entity, cmd *network.InputCommand, logger *logrus.Logger) {
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

			if logger.GetLevel() >= logrus.DebugLevel && (moveX != 0 || moveY != 0) {
				logging.NetworkLogger(logger, "", "").WithFields(logrus.Fields{
					"playerID":  cmd.PlayerID,
					"velocityX": velocity.VX,
					"velocityY": velocity.VY,
				}).Debug("player moving")
			}
		}

	case "attack":
		// Trigger attack
		if logger.GetLevel() >= logrus.DebugLevel {
			logging.NetworkLogger(logger, "", "").WithField("playerID", cmd.PlayerID).Debug("player attacking")
		}

		// Get attack component
		attackComp, hasAttack := entity.GetComponent("attack")
		if !hasAttack {
			if logger.GetLevel() >= logrus.WarnLevel {
				logging.NetworkLogger(logger, "", "").WithField("playerID", cmd.PlayerID).Warn("player has no attack component")
			}
			return
		}
		attack := attackComp.(*engine.AttackComponent)

		// Check cooldown using CanAttack method
		if !attack.CanAttack() {
			if logger.GetLevel() >= logrus.DebugLevel {
				logging.NetworkLogger(logger, "", "").WithFields(logrus.Fields{
					"playerID":       cmd.PlayerID,
					"cooldownRemain": attack.CooldownTimer,
				}).Debug("player attack on cooldown")
			}
			return
		}

		// Trigger attack by resetting cooldown
		attack.ResetCooldown()

		if logger.GetLevel() >= logrus.DebugLevel {
			logging.NetworkLogger(logger, "", "").WithFields(logrus.Fields{
				"playerID": cmd.PlayerID,
				"damage":   attack.Damage,
				"range":    attack.Range,
			}).Debug("player attack triggered")
		}

	case "use_item":
		// Use item from inventory
		if logger.GetLevel() >= logrus.DebugLevel {
			logging.NetworkLogger(logger, "", "").WithField("playerID", cmd.PlayerID).Debug("player using item")
		}

		// Get inventory component
		invComp, hasInv := entity.GetComponent("inventory")
		if !hasInv {
			if logger.GetLevel() >= logrus.WarnLevel {
				logging.NetworkLogger(logger, "", "").WithField("playerID", cmd.PlayerID).Warn("player has no inventory component")
			}
			return
		}
		inventory := invComp.(*engine.InventoryComponent)

		// Parse item index from command data
		if len(cmd.Data) < 1 {
			if logger.GetLevel() >= logrus.WarnLevel {
				logging.NetworkLogger(logger, "", "").WithField("playerID", cmd.PlayerID).Warn("use_item command missing item index")
			}
			return
		}
		itemIndex := int(cmd.Data[0])

		// Validate item index
		if itemIndex < 0 || itemIndex >= len(inventory.Items) {
			if logger.GetLevel() >= logrus.WarnLevel {
				logging.NetworkLogger(logger, "", "").WithFields(logrus.Fields{
					"playerID":      cmd.PlayerID,
					"itemIndex":     itemIndex,
					"inventorySize": len(inventory.Items),
				}).Warn("invalid item index")
			}
			return
		}

		// Get item
		item := inventory.Items[itemIndex]

		// Check if item is consumable (using imported item package constant)
		if item.Type != itemgen.TypeConsumable {
			if logger.GetLevel() >= logrus.WarnLevel {
				logging.NetworkLogger(logger, "", "").WithFields(logrus.Fields{
					"playerID": cmd.PlayerID,
					"itemName": item.Name,
				}).Warn("attempted to use non-consumable item")
			}
			return
		}

		// Apply item effect (health restoration for now)
		if healthComp, hasHealth := entity.GetComponent("health"); hasHealth {
			health := healthComp.(*engine.HealthComponent)

			// Restore health based on item power
			healAmount := float64(item.Stats.Defense) // Use defense stat as heal power
			if healAmount > 0 {
				health.Current += healAmount
				if health.Current > health.Max {
					health.Current = health.Max
				}

				if logger.GetLevel() >= logrus.InfoLevel {
					logging.NetworkLogger(logger, "", "").WithFields(logrus.Fields{
						"playerID":      cmd.PlayerID,
						"itemName":      item.Name,
						"healAmount":    healAmount,
						"currentHealth": health.Current,
						"maxHealth":     health.Max,
					}).Info("player used item")
				}

				// Remove consumed item from inventory
				inventory.Items = append(inventory.Items[:itemIndex], inventory.Items[itemIndex+1:]...)
			}
		}

	default:
		if logger.GetLevel() >= logrus.WarnLevel {
			logging.NetworkLogger(logger, "", "").WithFields(logrus.Fields{
				"playerID":  cmd.PlayerID,
				"inputType": cmd.InputType,
			}).Warn("unknown input type")
		}
	}
}
