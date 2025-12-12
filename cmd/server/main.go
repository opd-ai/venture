//go:build !android && !ios
// +build !android,!ios

// Package main provides the dedicated server application.
// Servers are not supported on mobile platforms.
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
	"github.com/opd-ai/venture/pkg/version"
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
	highLatency   = flag.Bool("high-latency", false, "Use high-latency configuration optimized for Tor/onion services (200-5000ms latency)")

	// V9.0 integration managers for server-authoritative validation
	// These are initialized in createGameWorld() and used by systems for validation
	v9StationManager      interface{}
	v9PetHomeManager      interface{}
	v9GuildHousingManager interface{}
)

func main() {
	flag.Parse()

	logger := initializeLogger()
	serverLogger := logger.WithFields(logrus.Fields{
		"component": "server",
		"seed":      *seed,
		"genre":     *genreID,
	})

	serverLogger.Infof("Starting Venture Game Server %s", version.FullVersion)
	serverLogger.WithFields(logrus.Fields{
		"port":          *port,
		"maxPlayers":    *maxPlayers,
		"tickRate":      *tickRate,
		"seed":          *seed,
		"genre":         *genreID,
		"aerialSprites": *aerialSprites,
	}).Info("server configuration")

	world := createGameWorld(logger)
	generatedTerrain := generateWorldTerrain(logger, serverLogger)
	spawnV4Entities(world, generatedTerrain, logger)

	server, snapshotManager, lagCompensator := initializeNetworkSystems(logger)

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

	startErrorHandler(server, logger)
	startPlayerManagementHandlers(server, world, generatedTerrain, logger)

	runGameLoop(world, server, snapshotManager, lagCompensator, ticker, logger, serverLogger, &lastUpdate)
}

// initializeLogger creates and configures the server logger with appropriate settings.
func initializeLogger() *logrus.Logger {
	logConfig := logging.Config{
		Level:       logging.InfoLevel,
		Format:      logging.JSONFormat,
		AddCaller:   true,
		EnableColor: false,
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		logConfig.Level = logging.LogLevel(logLevel)
	} else if *verbose {
		logConfig.Level = logging.DebugLevel
	}

	return logging.NewLogger(logConfig)
}

// createGameWorld initializes the game world with all required systems.
func createGameWorld(logger *logrus.Logger) *engine.World {
	worldLogger := logger.WithFields(logrus.Fields{"system": "world"})
	if logger.GetLevel() >= logrus.DebugLevel {
		worldLogger.Debug("creating game world")
	}

	world := engine.NewWorldWithLogger(logger)

	movementSystem := engine.NewMovementSystem(200.0)
	collisionSystem := engine.NewCollisionSystem(64.0)
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

	// Create item generator for crafting system
	itemGen := itemgen.NewItemGenerator()

	// INTEGRATION FIX [Category A]: Core Gameplay Systems Added to Server
	// Gap: 29 server-critical systems were only on client, causing multiplayer desync
	// Fix: Added all missing gameplay systems for server-authoritative state
	// Roadmap: Multiple phases (V3-V6) - complete multiplayer parity
	initializeCoreGameplaySystems(world, *seed, logger, inventorySystem, itemGen)

	initializeV4Systems(world, *seed, logger)
	initializeV5SystemsServer(world, logger)
	initializeV6SystemsServer(world, *seed, logger)

	// INTEGRATION FIX [Category A]: V8.0 Server System Initialization
	// Gap: V8.0 systems implemented but never initialized on server
	// Fix: Added V8.0 system initialization call for housing, fluids, vehicle physics
	// Roadmap: ROADMAP_V8.md (Phase 49-51)
	initializeV8SystemsServer(world, *seed, logger)

	// INTEGRATION FIX [Category A]: V9.0 Server Integration Manager Initialization
	// Gap: V9.0 integration managers were client-only, allowing XP/loyalty/permission exploits
	// Fix: Added V9.0 manager initialization for server-authoritative validation
	// Roadmap: ROADMAP_V9.md (Phase 55.1-55.3)
	v9StationManager, v9PetHomeManager, v9GuildHousingManager = initializeV9SystemsServer(logger)

	if logger.GetLevel() >= logrus.DebugLevel {
		worldLogger.Debug("game systems initialized")
	}

	return world
}

// generateWorldTerrain creates the initial terrain and spawns V4 entities.
func generateWorldTerrain(logger *logrus.Logger, serverLogger *logrus.Entry) *terrain.Terrain {
	terrainLogger := logging.GeneratorLogger(logger, "terrain", *seed, *genreID)
	if logger.GetLevel() >= logrus.DebugLevel {
		terrainLogger.WithFields(logrus.Fields{
			"width":  100,
			"height": 100,
		}).Debug("generating world terrain")
	}

	terrainGen := terrain.NewBSPGeneratorWithLogger(logger)
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

	return generatedTerrain
}

// spawnV4Entities spawns vehicles, companions, and bookshelves in the terrain.
func spawnV4Entities(world *engine.World, generatedTerrain *terrain.Terrain, logger *logrus.Logger) {
	v4Logger := logging.GeneratorLogger(logger, "v4-spawning", *seed, *genreID)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    *genreID,
		Custom: map[string]interface{}{
			"width":  100,
			"height": 100,
		},
	}

	// Spawn enemies using entity generator
	enemyCount, err := engine.SpawnEnemiesInTerrain(world, generatedTerrain, *seed, params)
	if err != nil {
		v4Logger.WithError(err).Warn("failed to spawn enemies")
	} else if enemyCount > 0 {
		v4Logger.WithField("count", enemyCount).Info("enemies spawned")
	}

	// Spawn merchants using entity generator
	merchantCount := 2
	merchantSpawned, err := engine.SpawnMerchantsInTerrain(world, generatedTerrain, *seed, params, merchantCount)
	if err != nil {
		v4Logger.WithError(err).Warn("failed to spawn merchants")
	} else if merchantSpawned > 0 {
		v4Logger.WithField("count", merchantSpawned).Info("merchants spawned")
	}

	vehicleCount, err := spawnVehiclesInTerrain(world, generatedTerrain, *seed, params, logger)
	if err != nil {
		v4Logger.WithError(err).Warn("failed to spawn vehicles")
	} else if vehicleCount > 0 {
		v4Logger.WithField("count", vehicleCount).Info("vehicles spawned")
	}

	companionCount, err := spawnCompanionsInTerrain(world, generatedTerrain, *seed, params, logger)
	if err != nil {
		v4Logger.WithError(err).Warn("failed to spawn companions")
	} else if companionCount > 0 {
		v4Logger.WithField("count", companionCount).Info("companions spawned")
	}

	bookshelfCount, err := spawnBookshelvesInTerrain(world, generatedTerrain, *seed, params, logger)
	if err != nil {
		v4Logger.WithError(err).Warn("failed to spawn bookshelves")
	} else if bookshelfCount > 0 {
		v4Logger.WithField("count", bookshelfCount).Info("bookshelves spawned")
	}
}

// initializeNetworkSystems configures network server, snapshot manager, and lag compensator.
func initializeNetworkSystems(logger *logrus.Logger) (*network.TCPServer, *network.SnapshotManager, *network.LagCompensator) {
	networkLogger := logger.WithFields(logrus.Fields{"system": "network"})
	if logger.GetLevel() >= logrus.DebugLevel {
		networkLogger.Debug("initializing network systems")
	}

	var serverConfig network.ServerConfig
	if *highLatency {
		serverConfig = network.HighLatencyServerConfig()
		networkLogger.Info("using high-latency server configuration (Tor/onion service optimized)")
	} else {
		serverConfig = network.DefaultServerConfig()
	}
	serverConfig.Address = ":" + *port
	serverConfig.MaxPlayers = *maxPlayers
	serverConfig.UpdateRate = *tickRate

	server := network.NewServerWithLogger(serverConfig, logger)
	snapshotManager := network.NewSnapshotManager(100)

	var lagCompConfig network.LagCompensationConfig
	if *highLatency {
		lagCompConfig = network.HighLatencyLagCompensationConfig()
		networkLogger.Info("using high-latency lag compensation configuration")
	} else {
		lagCompConfig = network.DefaultLagCompensationConfig()
	}
	lagCompensator := network.NewLagCompensator(lagCompConfig)

	networkLogger.WithFields(logrus.Fields{
		"address":    serverConfig.Address,
		"maxPlayers": serverConfig.MaxPlayers,
		"updateRate": serverConfig.UpdateRate,
	}).Info("network systems initialized")

	return server, snapshotManager, lagCompensator
}

// startErrorHandler starts a goroutine to handle network errors.
func startErrorHandler(server *network.TCPServer, logger *logrus.Logger) {
	networkLogger := logger.WithFields(logrus.Fields{"system": "network"})
	go func() {
		for err := range server.ReceiveError() {
			networkLogger.WithError(err).Error("network error")
		}
	}()
}

// startPlayerManagementHandlers starts goroutines for player join, leave, and input handling.
func startPlayerManagementHandlers(server *network.TCPServer, world *engine.World, generatedTerrain *terrain.Terrain, logger *logrus.Logger) {
	playerEntities := make(map[uint64]*engine.Entity)
	playerEntitiesMu := &sync.RWMutex{}
	networkLogger := logger.WithFields(logrus.Fields{"system": "network"})

	go handlePlayerJoins(server, world, generatedTerrain, playerEntities, playerEntitiesMu, logger)
	go handlePlayerLeaves(server, world, playerEntities, playerEntitiesMu, logger)
	go handleInputCommands(server, playerEntities, playerEntitiesMu, networkLogger, logger)
}

// handlePlayerJoins processes new player connections and creates entities.
func handlePlayerJoins(server *network.TCPServer, world *engine.World, generatedTerrain *terrain.Terrain, playerEntities map[uint64]*engine.Entity, playerEntitiesMu *sync.RWMutex, logger *logrus.Logger) {
	for playerID := range server.ReceivePlayerJoin() {
		playerLogger := logging.NetworkLogger(logger, "", "connected").WithField("playerID", playerID)
		playerLogger.Info("player joined - creating entity")

		entity := createPlayerEntity(world, generatedTerrain, playerID, *seed, *genreID, *aerialSprites, logger)

		playerEntitiesMu.Lock()
		playerEntities[playerID] = entity
		playerEntitiesMu.Unlock()

		playerLogger.WithField("entityID", entity.ID).Debug("player entity created")
	}
}

// handlePlayerLeaves processes player disconnections and removes entities.
func handlePlayerLeaves(server *network.TCPServer, world *engine.World, playerEntities map[uint64]*engine.Entity, playerEntitiesMu *sync.RWMutex, logger *logrus.Logger) {
	for playerID := range server.ReceivePlayerLeave() {
		playerLogger := logging.NetworkLogger(logger, "", "disconnected").WithField("playerID", playerID)
		playerLogger.Info("player left - removing entity")

		playerEntitiesMu.Lock()
		if entity, exists := playerEntities[playerID]; exists {
			world.RemoveEntity(entity.ID)
			delete(playerEntities, playerID)
			playerLogger.WithField("entityID", entity.ID).Debug("player entity removed")
		}
		playerEntitiesMu.Unlock()
	}
}

// handleInputCommands processes client input and applies it to player entities.
func handleInputCommands(server *network.TCPServer, playerEntities map[uint64]*engine.Entity, playerEntitiesMu *sync.RWMutex, networkLogger *logrus.Entry, logger *logrus.Logger) {
	for cmd := range server.ReceiveInputCommand() {
		if logger.GetLevel() >= logrus.DebugLevel {
			networkLogger.WithFields(logrus.Fields{
				"playerID":       cmd.PlayerID,
				"inputType":      cmd.InputType,
				"sequenceNumber": cmd.SequenceNumber,
			}).Debug("received input command")
		}

		playerEntitiesMu.RLock()
		entity, exists := playerEntities[cmd.PlayerID]
		playerEntitiesMu.RUnlock()

		if !exists {
			if logger.GetLevel() >= logrus.WarnLevel {
				networkLogger.WithField("playerID", cmd.PlayerID).Warn("no entity for player")
			}
			continue
		}

		applyInputCommand(entity, cmd, logger)
	}
}

// runGameLoop executes the authoritative server game loop.
func runGameLoop(world *engine.World, server *network.TCPServer, snapshotManager *network.SnapshotManager, lagCompensator *network.LagCompensator, ticker *time.Ticker, logger *logrus.Logger, serverLogger *logrus.Entry, lastUpdate *time.Time) {
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			deltaTime := now.Sub(*lastUpdate).Seconds()
			*lastUpdate = now

			world.Update(deltaTime)

			snapshot := buildWorldSnapshot(world, now)
			snapshotManager.AddSnapshot(snapshot)
			lagCompensator.RecordSnapshot(snapshot)

			stateUpdate := convertSnapshotToStateUpdate(snapshot)
			server.BroadcastStateUpdate(stateUpdate)

			if logger.GetLevel() >= logrus.DebugLevel && int(now.Unix())%10 == 0 {
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

			entitySnapshot := network.EntitySnapshot{
				EntityID:   entity.ID,
				Position:   network.Position{X: pos.X, Y: pos.Y},
				Velocity:   network.Velocity{VX: velX, VY: velY},
				Components: make(map[string][]byte), // Initialize component data map
			}

			// Serialize V4.0 component data for network transmission
			// Vehicle component
			if vehicleComp, ok := entity.GetComponent("vehicle"); ok {
				vehicle := vehicleComp.(*engine.VehicleComponent)
				entitySnapshot.Components["vehicle"] = vehicle.Serialize()
			}

			// Companion component
			if companionComp, ok := entity.GetComponent("companion"); ok {
				companion := companionComp.(*engine.CompanionComponent)
				entitySnapshot.Components["companion"] = companion.Serialize()
			}

			// Mount component (rider-vehicle relationship)
			if mountComp, ok := entity.GetComponent("mount"); ok {
				mount := mountComp.(*engine.MountComponent)
				entitySnapshot.Components["mount"] = mount.Serialize()
			}

			// Achievement component
			if achievementComp, ok := entity.GetComponent("achievement"); ok {
				achievement := achievementComp.(*engine.AchievementComponent)
				entitySnapshot.Components["achievement"] = achievement.Serialize()
			}

			// Bookshelf component
			if bookshelfComp, ok := entity.GetComponent("bookshelf"); ok {
				bookshelf := bookshelfComp.(*engine.BookshelfComponent)
				entitySnapshot.Components["bookshelf"] = bookshelf.Serialize()
			}

			snapshot.Entities[entity.ID] = entitySnapshot
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
		Priority:  network.PriorityNormal,                          // Normal priority for regular updates
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
	velComp, hasVel := entity.GetComponent("velocity")
	if !hasVel {
		return
	}
	velocity := velComp.(*engine.VelocityComponent)

	switch cmd.InputType {
	case "move":
		applyMovement(velocity, cmd, logger)
	case "attack":
		applyAttack(entity, cmd, logger)
	case "use_item":
		applyItemUse(entity, cmd, logger)
	default:
		logUnknownInput(cmd, logger)
	}
}

// applyMovement processes movement input and updates entity velocity.
func applyMovement(velocity *engine.VelocityComponent, cmd *network.InputCommand, logger *logrus.Logger) {
	if len(cmd.Data) < 2 {
		return
	}

	moveX := float64(int8(cmd.Data[0])) / 127.0
	moveY := float64(int8(cmd.Data[1])) / 127.0

	if moveX != 0 && moveY != 0 {
		moveX *= 0.707
		moveY *= 0.707
	}

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

// applyAttack processes attack input and triggers entity attack if cooldown permits.
func applyAttack(entity *engine.Entity, cmd *network.InputCommand, logger *logrus.Logger) {
	if logger.GetLevel() >= logrus.DebugLevel {
		logging.NetworkLogger(logger, "", "").WithField("playerID", cmd.PlayerID).Debug("player attacking")
	}

	attackComp, hasAttack := entity.GetComponent("attack")
	if !hasAttack {
		if logger.GetLevel() >= logrus.WarnLevel {
			logging.NetworkLogger(logger, "", "").WithField("playerID", cmd.PlayerID).Warn("player has no attack component")
		}
		return
	}
	attack := attackComp.(*engine.AttackComponent)

	if !attack.CanAttack() {
		if logger.GetLevel() >= logrus.DebugLevel {
			logging.NetworkLogger(logger, "", "").WithFields(logrus.Fields{
				"playerID":       cmd.PlayerID,
				"cooldownRemain": attack.CooldownTimer,
			}).Debug("player attack on cooldown")
		}
		return
	}

	attack.ResetCooldown()

	if logger.GetLevel() >= logrus.DebugLevel {
		logging.NetworkLogger(logger, "", "").WithFields(logrus.Fields{
			"playerID": cmd.PlayerID,
			"damage":   attack.Damage,
			"range":    attack.Range,
		}).Debug("player attack triggered")
	}
}

// applyItemUse processes item usage from inventory and applies effects.
func applyItemUse(entity *engine.Entity, cmd *network.InputCommand, logger *logrus.Logger) {
	if logger.GetLevel() >= logrus.DebugLevel {
		logging.NetworkLogger(logger, "", "").WithField("playerID", cmd.PlayerID).Debug("player using item")
	}

	invComp, hasInv := entity.GetComponent("inventory")
	if !hasInv {
		if logger.GetLevel() >= logrus.WarnLevel {
			logging.NetworkLogger(logger, "", "").WithField("playerID", cmd.PlayerID).Warn("player has no inventory component")
		}
		return
	}
	inventory := invComp.(*engine.InventoryComponent)

	item, itemIndex, ok := validateItemUse(inventory, cmd, logger)
	if !ok {
		return
	}

	consumeItem(entity, inventory, item, itemIndex, cmd.PlayerID, logger)
}

// validateItemUse validates item index and type for usage.
func validateItemUse(inventory *engine.InventoryComponent, cmd *network.InputCommand, logger *logrus.Logger) (*itemgen.Item, int, bool) {
	if len(cmd.Data) < 1 {
		if logger.GetLevel() >= logrus.WarnLevel {
			logging.NetworkLogger(logger, "", "").WithField("playerID", cmd.PlayerID).Warn("use_item command missing item index")
		}
		return nil, 0, false
	}
	itemIndex := int(cmd.Data[0])

	if itemIndex < 0 || itemIndex >= len(inventory.Items) {
		if logger.GetLevel() >= logrus.WarnLevel {
			logging.NetworkLogger(logger, "", "").WithFields(logrus.Fields{
				"playerID":      cmd.PlayerID,
				"itemIndex":     itemIndex,
				"inventorySize": len(inventory.Items),
			}).Warn("invalid item index")
		}
		return nil, 0, false
	}

	item := inventory.Items[itemIndex]
	if item.Type != itemgen.TypeConsumable {
		if logger.GetLevel() >= logrus.WarnLevel {
			logging.NetworkLogger(logger, "", "").WithFields(logrus.Fields{
				"playerID": cmd.PlayerID,
				"itemName": item.Name,
			}).Warn("attempted to use non-consumable item")
		}
		return nil, 0, false
	}

	return item, itemIndex, true
}

// consumeItem applies consumable item effects and removes it from inventory.
func consumeItem(entity *engine.Entity, inventory *engine.InventoryComponent, item *itemgen.Item, itemIndex int, playerID uint64, logger *logrus.Logger) {
	healthComp, hasHealth := entity.GetComponent("health")
	if !hasHealth {
		return
	}
	health := healthComp.(*engine.HealthComponent)

	healAmount := float64(item.Stats.Defense)
	if healAmount <= 0 {
		return
	}

	health.Current += healAmount
	if health.Current > health.Max {
		health.Current = health.Max
	}

	if logger.GetLevel() >= logrus.InfoLevel {
		logging.NetworkLogger(logger, "", "").WithFields(logrus.Fields{
			"playerID":      playerID,
			"itemName":      item.Name,
			"healAmount":    healAmount,
			"currentHealth": health.Current,
			"maxHealth":     health.Max,
		}).Info("player used item")
	}

	inventory.Items = append(inventory.Items[:itemIndex], inventory.Items[itemIndex+1:]...)
}

// logUnknownInput logs warning for unknown input types.
func logUnknownInput(cmd *network.InputCommand, logger *logrus.Logger) {
	if logger.GetLevel() >= logrus.WarnLevel {
		logging.NetworkLogger(logger, "", "").WithFields(logrus.Fields{
			"playerID":  cmd.PlayerID,
			"inputType": cmd.InputType,
		}).Warn("unknown input type")
	}
}
