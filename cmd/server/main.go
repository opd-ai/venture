//go:build !android && !ios
// +build !android,!ios

// Package main provides the dedicated server application.
// Servers are not supported on mobile platforms.
package main

import (
	"context"
	"flag"
	"fmt"
	"image/color"
	"os"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/balance"
	"github.com/opd-ai/venture/pkg/combat"
	"github.com/opd-ai/venture/pkg/config"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/migration"
	"github.com/opd-ai/venture/pkg/modding"
	"github.com/opd-ai/venture/pkg/network"
	"github.com/opd-ai/venture/pkg/network/resilience"
	"github.com/opd-ai/venture/pkg/procgen"
	itemgen "github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/opd-ai/venture/pkg/security"
	"github.com/opd-ai/venture/pkg/stability"
	"github.com/opd-ai/venture/pkg/ux"
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

	// Phase 2 (PLAN.md): Security and stability integration
	securityAudit    = flag.Bool("security-audit", false, "Run security audit at startup and log results")
	stabilityMonitor = flag.Bool("stability-monitor", false, "Enable stability monitoring for production validation")

	// Phase 4.1 (PLAN.md): Network resilience testing
	simulateNetwork   = flag.String("simulate-network", "", "Simulate network conditions for testing: low, medium, high, very-high, extreme")
	resilienceMetrics = flag.Bool("resilience-metrics", false, "Enable network resilience metrics collection")

	// Phase 6.1 (PLAN.md): Balance validation integration
	balanceValidate = flag.Bool("balance-validate", false, "Run combat and economic balance validation at startup")

	// Phase 6.2 (PLAN.md): Migration validation integration
	migrationValidate = flag.Bool("migration-validate", false, "Run save file migration validation at startup")

	// Phase 6.4 (PLAN.md): UX journey validation integration
	uxValidate = flag.Bool("ux-validate", false, "Run user experience journey validation at startup")

	// Phase 6.3 (PLAN.md): Modding system integration
	enableMods = flag.Bool("enable-mods", false, "Enable mod system with sandbox security")
	modsDir    = flag.String("mods-dir", "mods", "Directory to load mods from")

	// V9.0 integration managers for server-authoritative validation
	// These are initialized in createGameWorld() and used by systems for validation
	v9StationManager      interface{}
	v9PetHomeManager      interface{}
	v9GuildHousingManager interface{}
)

func main() {
	flag.Parse()

	// Validate configuration before starting server
	validator := config.NewValidator()
	cfg := &config.Config{
		Port:               *port,
		MaxPlayers:         *maxPlayers,
		ValidateMaxPlayers: true,
		TickRate:           *tickRate,
		ValidateTickRate:   true,
		Genre:              *genreID,
		ModsDir:            *modsDir,
		CreateDirs:         true, // Auto-create mods directory if needed
	}

	if err := validator.ValidateAll(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nUsage: %s [options]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

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

	// Phase 2 (PLAN.md): Run security audit if enabled
	if *securityAudit {
		runSecurityAudit(serverLogger)
	}

	// Phase 6.1 (PLAN.md): Run balance validation if enabled
	if *balanceValidate {
		runBalanceValidation(serverLogger)
	}

	// Phase 6.2 (PLAN.md): Run migration validation if enabled
	if *migrationValidate {
		runMigrationValidation(serverLogger)
	}

	// Phase 6.4 (PLAN.md): Run UX journey validation if enabled
	if *uxValidate {
		runUXValidation(serverLogger)
	}

	// Phase 6.3 (PLAN.md): Initialize mod system if enabled
	var modManager *modding.Manager
	if *enableMods {
		modManager = initializeModSystem(serverLogger)
	}
	_ = modManager // Available for future game rule integration

	// Phase 2 (PLAN.md): Start stability monitoring if enabled
	var stabilityMon *stability.Monitor
	if *stabilityMonitor {
		stabilityMon = startStabilityMonitoring(serverLogger)
	}

	// Phase 4.1 (PLAN.md): Initialize network resilience testing
	var networkSim *resilience.NetworkSimulator
	var metricsCollector *resilience.MetricsCollector
	if *simulateNetwork != "" || *resilienceMetrics {
		networkSim, metricsCollector = initializeResilienceTesting(serverLogger)
	}

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
		if stabilityMon != nil {
			stabilityMon.Stop()
			serverLogger.Info("stability monitor stopped")
		}
		// Phase 4.1 (PLAN.md): Log resilience metrics on shutdown
		if metricsCollector != nil {
			stats := metricsCollector.GetStats()
			serverLogger.WithFields(logrus.Fields{
				"avg_latency":      stats.AvgLatency.String(),
				"packets_sent":     stats.PacketsSent,
				"packets_dropped":  stats.PacketsDropped,
				"packet_loss_rate": stats.PacketLossRate,
				"mispredictions":   stats.MispredictionCount,
				"desyncs":          stats.DesyncCount,
				"duration":         stats.Duration.String(),
			}).Info("network resilience metrics summary")
		}
		if networkSim != nil {
			sent, dropped, bytes := networkSim.GetStats()
			serverLogger.WithFields(logrus.Fields{
				"packets_sent":    sent,
				"packets_dropped": dropped,
				"bytes_processed": bytes,
			}).Info("network simulation stats")
		}
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

	// Phase 1.2 (PLAN.md): Economy system for marketplace and guild banking
	serverID := "server-" + *port // Generate server ID from port for uniqueness
	economySystem := engine.NewEconomySystem(world, serverID)
	world.AddSystem(economySystem)
	worldLogger.WithField("serverID", serverID).Debug("economy system initialized")

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

// runSecurityAudit performs security validation at server startup.
// Phase 2.4 (PLAN.md): Unconditional security package integration
func runSecurityAudit(serverLogger *logrus.Entry) {
	serverLogger.Info("running security audit at startup")
	auditor := security.NewAuditor(nil)
	results := auditor.RunFullAudit()

	serverLogger.WithFields(logrus.Fields{
		"total_checks": results.TotalChecks,
		"passed":       results.PassedChecks,
		"failed":       results.FailedChecks,
		"critical":     results.CriticalCount,
		"high":         results.HighCount,
		"pass_rate":    float64(results.PassedChecks) / float64(results.TotalChecks) * 100.0,
		"duration_ms":  results.EndTime.Sub(results.StartTime).Milliseconds(),
	}).Info("security audit completed")

	if results.HasCritical() {
		serverLogger.Warn("critical security vulnerabilities detected - review security audit results")
	}
}

// startStabilityMonitoring initializes and starts the stability monitor for production validation.
// Phase 2.3 (PLAN.md): Unconditional stability package integration
func startStabilityMonitoring(serverLogger *logrus.Entry) *stability.Monitor {
	config := stability.DefaultConfig()
	// Use shorter duration for non-production testing
	config.Duration = 24 * time.Hour
	config.CheckInterval = 60 * time.Second

	monitor := stability.NewMonitor(config)

	// Start the monitor in a background goroutine to perform actual health checks
	go func() {
		report, err := monitor.Run(context.Background())
		if err != nil {
			serverLogger.WithError(err).Error("stability monitor encountered error")
		}
		if report != nil {
			serverLogger.WithFields(logrus.Fields{
				"passed":       report.Passed,
				"uptime":       report.TotalUptime.String(),
				"avg_fps":      report.AvgFPS,
				"peak_memory":  report.PeakMemory,
				"memory_leaks": report.MemoryLeakCount,
				"checks":       report.Checks,
			}).Info("stability monitoring completed")
			if !report.Passed {
				serverLogger.WithField("reason", report.FailureReason).Warn("stability test failed")
			}
		}
	}()

	serverLogger.WithFields(logrus.Fields{
		"duration":       config.Duration.String(),
		"check_interval": config.CheckInterval.String(),
		"memory_limit":   config.MemoryLimit,
		"min_fps":        config.MinFPS,
	}).Info("stability monitoring started")

	return monitor
}

// initializeResilienceTesting sets up network resilience testing tools.
// Phase 4.1 (PLAN.md): Network resilience testing integration
func initializeResilienceTesting(serverLogger *logrus.Entry) (*resilience.NetworkSimulator, *resilience.MetricsCollector) {
	var networkSim *resilience.NetworkSimulator
	var metricsCollector *resilience.MetricsCollector

	// Initialize metrics collector for network performance tracking
	if *resilienceMetrics {
		metricsCollector = resilience.NewMetricsCollector()
		serverLogger.Info("network resilience metrics collector initialized")
	}

	// Initialize network simulator based on scenario
	if *simulateNetwork != "" {
		var scenario *resilience.TestScenario

		switch *simulateNetwork {
		case "low":
			scenario = &resilience.LowLatencyScenario
		case "medium":
			scenario = &resilience.MediumLatencyScenario
		case "high":
			scenario = &resilience.HighLatencyScenario
		case "very-high":
			scenario = &resilience.VeryHighLatencyScenario
		case "extreme":
			scenario = &resilience.ExtremeLatencyScenario
		default:
			serverLogger.WithField("scenario", *simulateNetwork).Warn("unknown network simulation scenario, using low latency")
			scenario = &resilience.LowLatencyScenario
		}

		var err error
		networkSim, err = resilience.NewNetworkSimulatorWithConfig(scenario.Config)
		if err != nil {
			serverLogger.WithError(err).Error("failed to initialize network simulator")
		} else {
			serverLogger.WithFields(logrus.Fields{
				"scenario":    scenario.Name,
				"latency":     scenario.Config.Latency.String(),
				"packet_loss": scenario.Config.PacketLossRate,
				"jitter":      scenario.Config.Jitter.String(),
				"bandwidth":   scenario.Config.BandwidthLimit,
			}).Info("network simulation enabled for testing")
		}
	}

	return networkSim, metricsCollector
}

// runBalanceValidation performs combat and economic balance validation at startup.
// Phase 6.1 (PLAN.md): Balance package integration for gameplay validation
func runBalanceValidation(serverLogger *logrus.Entry) {
	serverLogger.Info("running balance validation at startup")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	config := balance.NewDefaultConfig()
	config.Seed = *seed

	// Use reduced simulation counts for faster startup validation
	config.SimulationCounts["Combat"] = 1000
	config.SimulationCounts["Economic"] = 500

	// Run combat validation
	combatValidator := balance.NewCombatValidator(config)
	combatResult, err := combatValidator.Validate(ctx)
	if err != nil {
		serverLogger.WithError(err).Error("combat balance validation failed")
	} else {
		logBalanceResult(serverLogger, combatResult)
	}

	// Run economic validation
	economicValidator := balance.NewEconomicValidator(config)
	economicResult, err := economicValidator.Validate(ctx)
	if err != nil {
		serverLogger.WithError(err).Error("economic balance validation failed")
	} else {
		logBalanceResult(serverLogger, economicResult)
	}

	serverLogger.Info("balance validation completed")
}

// logBalanceResult logs the results of a balance validation run.
func logBalanceResult(serverLogger *logrus.Entry, result *balance.ValidationResult) {
	fields := logrus.Fields{
		"domain":           result.Domain,
		"passed":           result.Passed,
		"simulation_count": result.SimulationCount,
		"duration_sec":     result.Duration,
		"issue_count":      len(result.Issues),
	}

	// Add key metrics to log output
	for key, value := range result.Metrics {
		fields[key] = value
	}

	if result.Passed {
		serverLogger.WithFields(fields).Info("balance validation passed")
	} else {
		serverLogger.WithFields(fields).Warn("balance validation detected issues")
		for _, issue := range result.Issues {
			serverLogger.WithField("domain", result.Domain).Warn(issue)
		}
		for _, rec := range result.Recommendations {
			serverLogger.WithField("domain", result.Domain).Info("recommendation: " + rec)
		}
	}
}

// runMigrationValidation validates that all supported save file versions can be migrated.
// Phase 6.2 (PLAN.md): Migration package integration for save file compatibility
func runMigrationValidation(serverLogger *logrus.Entry) {
	serverLogger.Info("running migration validation at startup")

	config := migration.Config{
		TargetVersion: "v10.0",
		ValidateData:  true,
	}

	validator := migration.NewValidator(config)
	results, err := validator.ValidateAll()
	if err != nil {
		serverLogger.WithError(err).Error("migration validation failed")
		return
	}

	serverLogger.WithFields(logrus.Fields{
		"total_migrations": results.TotalCount,
		"passed":           results.PassedCount,
		"failed":           results.FailedCount,
		"pass_rate":        float64(results.PassedCount) / float64(results.TotalCount) * 100.0,
	}).Info("migration validation completed")

	if results.FailedCount > 0 {
		serverLogger.Warn("some migrations failed - save file compatibility may be affected")
		for _, m := range results.Migrations {
			if !m.Passed {
				serverLogger.WithFields(logrus.Fields{
					"source_version": m.SourceVersion,
					"target_version": m.TargetVersion,
					"error":          m.Error,
				}).Warn("migration failed")
			}
		}
	}
}

// runUXValidation performs user experience journey validation at startup.
// Phase 6.4 (PLAN.md): UX package integration for user journey validation
func runUXValidation(serverLogger *logrus.Entry) {
	serverLogger.Info("running UX journey validation at startup")

	config := ux.ValidationConfig{
		Runs:                 5,
		TimeTolerancePercent: 20.0,
		MinCompletionRate:    0.90,
		MinSatisfaction:      0.80,
		MaxErrorRate:         0.05,
	}

	validator := ux.NewJourneyValidatorWithConfig(config)
	results := validator.ValidateAll()
	summary := ux.GetSummary(results)

	serverLogger.WithFields(logrus.Fields{
		"total_journeys":   summary.TotalJourneys,
		"passed":           summary.PassedJourneys,
		"pass_rate":        summary.PassRate * 100.0,
		"avg_completion":   summary.AverageCompletionRate * 100.0,
		"avg_satisfaction": summary.AverageSatisfaction * 100.0,
		"avg_error_rate":   summary.AverageErrorRate * 100.0,
	}).Info("UX journey validation completed")

	if summary.PassedJourneys < summary.TotalJourneys {
		serverLogger.Warn("some user journeys failed validation - UX issues may exist")
		for _, result := range results {
			if !result.Passed {
				serverLogger.WithFields(logrus.Fields{
					"journey":         result.Name,
					"completion_rate": result.CompletionRate * 100.0,
					"satisfaction":    result.Satisfaction * 100.0,
					"error_rate":      result.ErrorRate * 100.0,
				}).Warn("journey failed")
			}
		}
	}
}

// initializeModSystem sets up the sandboxed mod system.
// Phase 6.3 (PLAN.md): Modding package integration with security sandbox
func initializeModSystem(serverLogger *logrus.Entry) *modding.Manager {
	serverLogger.WithFields(logrus.Fields{
		"mods_dir": *modsDir,
	}).Info("initializing mod system with security sandbox")

	// Validate sandbox security first
	sandbox := modding.NewSandboxWithConfig(modding.SandboxConfig{
		ModsDirectory: *modsDir,
	})
	report := sandbox.GenerateSecurityReport()

	if !report.AllChecksPassed() {
		serverLogger.WithFields(logrus.Fields{
			"passed_checks":  report.PassedCount(),
			"file_system":    report.FileSystemIsolation,
			"network":        report.NetworkIsolation,
			"memory":         report.MemoryLimits,
			"cpu":            report.CPULimits,
			"api":            report.APIRestrictions,
			"code_execution": report.CodeExecution,
		}).Error("mod sandbox security checks failed - mods disabled")
		return nil
	}

	serverLogger.WithFields(logrus.Fields{
		"passed_checks": report.PassedCount(),
	}).Debug("mod sandbox security checks passed")

	// Create mod loader and manager
	config := modding.ModConfig{
		ModsDirectory:       *modsDir,
		EnableSandbox:       true,
		MaxMods:             50,
		RuleChangeRateLimit: 10.0,
	}

	loader := modding.NewLoaderWithConfig(config)
	manager := modding.NewManagerWithConfig(config)

	// Load all mods from directory
	mods, err := loader.LoadAll()
	if err != nil {
		serverLogger.WithError(err).Warn("failed to load mods")
		return manager
	}

	// Add and enable loaded mods
	loadedCount := 0
	for _, mod := range mods {
		if err := manager.AddMod(mod); err != nil {
			serverLogger.WithFields(logrus.Fields{
				"mod_id":   mod.ID,
				"mod_name": mod.Name,
				"error":    err.Error(),
			}).Warn("failed to add mod")
			continue
		}

		if err := manager.EnableMod(mod.ID); err != nil {
			serverLogger.WithFields(logrus.Fields{
				"mod_id": mod.ID,
				"error":  err.Error(),
			}).Warn("failed to enable mod")
			continue
		}

		loadedCount++
		serverLogger.WithFields(logrus.Fields{
			"mod_id":      mod.ID,
			"mod_name":    mod.Name,
			"mod_version": mod.Version,
			"mod_type":    mod.Type.String(),
			"rule_count":  len(mod.Rules),
		}).Info("mod loaded and enabled")
	}

	// Apply all enabled mod rules
	if loadedCount > 0 {
		if err := manager.ApplyRules(); err != nil {
			serverLogger.WithError(err).Warn("failed to apply mod rules")
		} else {
			stats := manager.GetStats()
			serverLogger.WithFields(logrus.Fields{
				"total_mods":   stats["total_mods"],
				"enabled_mods": stats["enabled_mods"],
				"active_rules": stats["active_rules"],
			}).Info("mod rules applied")
		}
	}

	serverLogger.WithFields(logrus.Fields{
		"mods_loaded": loadedCount,
		"mods_total":  len(mods),
	}).Info("mod system initialized")

	return manager
}
