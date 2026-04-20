//go:build !android && !ios
// +build !android,!ios

// Package main provides the dedicated server application.
// Servers are not supported on mobile platforms.
package main

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/opd-ai/venture/pkg/config"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/engine/prestige"
	"github.com/opd-ai/venture/pkg/engine/qol"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/modding"
	"github.com/opd-ai/venture/pkg/network"
	"github.com/opd-ai/venture/pkg/network/resilience"
	"github.com/opd-ai/venture/pkg/observability"
	"github.com/opd-ai/venture/pkg/procgen"
	itemgen "github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/stability"
	"github.com/opd-ai/venture/pkg/version"
	"github.com/sirupsen/logrus"
)

var (
	port          = flag.String("port", "8080", "Server port")
	maxPlayers    = flag.Int("max-players", 8, "Maximum number of players")
	seed          = flag.Int64("seed", 12345, "World generation seed")
	genreID       = flag.String("genre", "fantasy", "Genre ID for world generation")
	terrainType   = flag.String("terrain-type", "bsp", "Terrain generator type: bsp, cellular, city, forest, composite, grammar, maze")
	tickRate      = flag.Int("tick-rate", 30, "Server update rate (updates per second)")
	verbose       = flag.Bool("verbose", true, "Enable verbose logging")
	aerialSprites = flag.Bool("aerial-sprites", true, "Enable aerial-view perspective sprites for top-down gameplay")
	highLatency   = flag.Bool("high-latency", false, "Use high-latency configuration optimized for Tor/onion services (200-5000ms latency)")

	// Phase 2 (PLAN.md): Security and stability integration (enabled by default for production readiness)
	securityAudit    = flag.Bool("security-audit", true, "Run security audit at startup and log results")
	stabilityMonitor = flag.Bool("stability-monitor", true, "Enable stability monitoring for production validation")

	// Phase 4.1 (PLAN.md): Network resilience testing
	simulateNetwork   = flag.String("simulate-network", "", "Simulate network conditions for testing: low, medium, high, very-high, extreme")
	resilienceMetrics = flag.Bool("resilience-metrics", true, "Enable network resilience metrics collection")

	// Phase 6.1 (PLAN.md): Balance validation integration
	balanceValidate = flag.Bool("balance-validate", false, "Run combat and economic balance validation at startup")

	// Phase 6.2 (PLAN.md): Migration validation integration
	migrationValidate = flag.Bool("migration-validate", false, "Run save file migration validation at startup")

	// Phase 6.4 (PLAN.md): UX journey validation integration
	uxValidate = flag.Bool("ux-validate", false, "Run user experience journey validation at startup")

	// Phase 6.3 (PLAN.md): Modding system integration (enabled by default)
	enableMods = flag.Bool("enable-mods", true, "Enable mod system with sandbox security")
	modsDir    = flag.String("mods-dir", "mods", "Directory to load mods from")

	// Phase 3 (PLAN.md): Metrics export for production monitoring (enabled by default)
	metricsPort   = flag.String("metrics-port", "9090", "Port for Prometheus metrics HTTP endpoint")
	enableMetrics = flag.Bool("enable-metrics", true, "Enable Prometheus metrics export at /metrics endpoint")

	// Federation server identity (uses environment variable if set)
	serverName = flag.String("server-name", getEnvOrDefault("SERVER_NAME", "venture-server"), "Server name for federation identity")

	// Version flag
	showVersion = flag.Bool("version", false, "Print version information and exit")
)

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	flag.Parse()

	if *showVersion {
		version.PrintVersion()
		os.Exit(0)
	}

	validateConfiguration()

	logger := initializeLogger()
	serverLogger := logger.WithFields(logrus.Fields{
		"component": "server",
		"seed":      *seed,
		"genre":     *genreID,
	})

	// Create root context with signal handling for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logServerStartup(serverLogger)
	runStartupValidations(serverLogger)

	modManager, stabilityMon, networkSim, metricsCollector := initializeOptionalSystems(ctx, serverLogger)

	world, enhancedChatSystem, courierSystem := createGameWorld(logger)

	// Wire mod system to world for game system access
	// Phase 6.3 (PLAN.md): Modding System Integration
	if modManager != nil {
		adapter := modding.NewProviderAdapter(modManager)
		world.SetModRules(adapter)
		serverLogger.WithFields(logrus.Fields{
			"active_rules": len(modManager.ListMods()),
		}).Info("mod system wired to game world")
	}

	generatedTerrain := generateWorldTerrain(logger, serverLogger)
	spawnV4Entities(world, generatedTerrain, logger)

	// AUDIT.md Task 4: Wire PostOfficeSpawner after CourierSystem
	// PostOfficeSpawner places post office buildings with clerk NPCs in the terrain.
	// Without this, CourierSystem runs but has no post offices for mail delivery.
	spawnPostOffices(world, generatedTerrain, courierSystem, logger)

	server, snapshotManager, lagCompensator := initializeNetworkSystems(logger)

	var metricsExporter *observability.MetricsExporter
	if *enableMetrics {
		var err error
		metricsExporter, err = initializeMetricsExporter(logger, world, server)
		if err != nil {
			logger.WithError(err).Fatal("failed to initialize metrics exporter (set -enable-metrics=false to disable)")
		}
	}

	startNetworkServer(server, serverLogger)
	logServerReady(serverLogger, world)

	// Execute game loop and wait for shutdown signal
	executeGameLoop(ctx, world, enhancedChatSystem, server, snapshotManager, lagCompensator, logger, serverLogger)

	// Graceful shutdown with deadline
	serverLogger.Info("shutdown signal received, initiating graceful shutdown")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	shutdownComplete := make(chan struct{})
	go func() {
		shutdownServer(serverLogger, metricsExporter, stabilityMon, metricsCollector, networkSim, server)
		close(shutdownComplete)
	}()

	select {
	case <-shutdownComplete:
		serverLogger.Info("graceful shutdown completed")
	case <-shutdownCtx.Done():
		serverLogger.Error("shutdown deadline exceeded, forcing exit")
	}
}

// validateConfiguration validates server configuration before startup.
func validateConfiguration() {
	validator := config.NewValidator()
	cfg := &config.Config{
		Port:               *port,
		MaxPlayers:         *maxPlayers,
		ValidateMaxPlayers: true,
		TickRate:           *tickRate,
		ValidateTickRate:   true,
		Genre:              *genreID,
		ModsDir:            *modsDir,
		CreateDirs:         true,
	}

	if err := validator.ValidateAll(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nUsage: %s [options]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
}

// logServerStartup logs the initial server startup information.
func logServerStartup(serverLogger *logrus.Entry) {
	serverLogger.Infof("Starting Venture Game Server %s", version.FullVersion)
	serverLogger.WithFields(logrus.Fields{
		"port":          *port,
		"maxPlayers":    *maxPlayers,
		"tickRate":      *tickRate,
		"seed":          *seed,
		"genre":         *genreID,
		"aerialSprites": *aerialSprites,
	}).Info("server configuration")
}

// runStartupValidations executes all enabled validation checks at startup.
func runStartupValidations(serverLogger *logrus.Entry) {
	if *securityAudit {
		runSecurityAudit(serverLogger)
	}

	if *balanceValidate {
		runBalanceValidation(serverLogger)
	}

	if *migrationValidate {
		runMigrationValidation(serverLogger)
	}

	if *uxValidate {
		runUXValidation(serverLogger)
	}
}

// initializeOptionalSystems initializes optional server systems based on configuration.
func initializeOptionalSystems(ctx context.Context, serverLogger *logrus.Entry) (*modding.Manager, *stability.Monitor, *resilience.NetworkSimulator, *resilience.MetricsCollector) {
	var modManager *modding.Manager
	if *enableMods {
		modManager = initializeModSystem(serverLogger)
	}

	var stabilityMon *stability.Monitor
	if *stabilityMonitor {
		stabilityMon = startStabilityMonitoring(ctx, serverLogger)
	}

	var networkSim *resilience.NetworkSimulator
	var metricsCollector *resilience.MetricsCollector
	if *simulateNetwork != "" || *resilienceMetrics {
		networkSim, metricsCollector = initializeResilienceTesting(serverLogger)
	}

	return modManager, stabilityMon, networkSim, metricsCollector
}

// startNetworkServer starts the network server and handles startup errors.
func startNetworkServer(server *network.TCPServer, serverLogger *logrus.Entry) {
	if err := server.Start(); err != nil {
		serverLogger.WithError(err).Fatal("failed to start network server")
	}
}

// logServerReady logs that the server is ready to accept connections.
func logServerReady(serverLogger *logrus.Entry, world *engine.World) {
	serverLogger.Info("server initialized successfully")
	serverLogger.WithFields(logrus.Fields{
		"port":        *port,
		"maxPlayers":  *maxPlayers,
		"updateRate":  *tickRate,
		"entityCount": len(world.GetEntities()),
	}).Info("server listening")
}

// shutdownServer performs graceful shutdown of all server components.
func shutdownServer(serverLogger *logrus.Entry, metricsExporter *observability.MetricsExporter, stabilityMon *stability.Monitor, metricsCollector *resilience.MetricsCollector, networkSim *resilience.NetworkSimulator, server *network.TCPServer) {
	serverLogger.Info("shutting down server")

	shutdownMetricsExporter(metricsExporter, serverLogger)
	shutdownStabilityMonitor(stabilityMon, serverLogger)
	logResilienceMetrics(metricsCollector, serverLogger)
	logNetworkSimulationStats(networkSim, serverLogger)

	if err := server.Stop(); err != nil {
		serverLogger.WithError(err).Error("error stopping server")
	}
}

// shutdownMetricsExporter stops the metrics exporter if it is running.
func shutdownMetricsExporter(metricsExporter *observability.MetricsExporter, serverLogger *logrus.Entry) {
	if metricsExporter != nil {
		if err := metricsExporter.Stop(); err != nil {
			serverLogger.WithError(err).Error("error stopping metrics exporter")
		} else {
			serverLogger.Info("metrics exporter stopped")
		}
	}
}

// shutdownStabilityMonitor stops the stability monitor if it is running.
func shutdownStabilityMonitor(stabilityMon *stability.Monitor, serverLogger *logrus.Entry) {
	if stabilityMon != nil {
		stabilityMon.Stop()
		serverLogger.Info("stability monitor stopped")
	}
}

// logResilienceMetrics logs network resilience metrics summary if available.
func logResilienceMetrics(metricsCollector *resilience.MetricsCollector, serverLogger *logrus.Entry) {
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
}

// logNetworkSimulationStats logs network simulation statistics if available.
func logNetworkSimulationStats(networkSim *resilience.NetworkSimulator, serverLogger *logrus.Entry) {
	if networkSim != nil {
		sent, dropped, bytes := networkSim.GetStats()
		serverLogger.WithFields(logrus.Fields{
			"packets_sent":    sent,
			"packets_dropped": dropped,
			"bytes_processed": bytes,
		}).Info("network simulation stats")
	}
}

// executeGameLoop sets up and runs the main server game loop.
func executeGameLoop(ctx context.Context, world *engine.World, enhancedChatSystem *engine.EnhancedChatSystem, server *network.TCPServer, snapshotManager *network.SnapshotManager, lagCompensator *network.LagCompensator, logger *logrus.Logger, serverLogger *logrus.Entry) {
	tickDuration := time.Duration(1000000000 / *tickRate)
	ticker := time.NewTicker(tickDuration)
	defer ticker.Stop()

	lastUpdate := time.Now()

	serverLogger.WithField("tickRate", *tickRate).Info("starting authoritative game loop")

	startErrorHandler(server, logger)
	startPlayerManagementHandlers(server, world, nil, enhancedChatSystem, logger)

	runGameLoop(ctx, world, server, snapshotManager, lagCompensator, ticker, logger, serverLogger, &lastUpdate)
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
// Returns the world, the enhanced chat system for player registration, and the courier system for post office spawning.
func createGameWorld(logger *logrus.Logger) (*engine.World, *engine.EnhancedChatSystem, *engine.CourierSystem) {
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
	craftingSystem, narrativeSystem := initializeCoreGameplaySystems(world, *seed, logger, inventorySystem, itemGen)

	companionLoyaltySystem, _ := initializeV4Systems(world, *seed, *genreID, logger, economySystem)
	enhancedChatSystem, courierSystem := initializeV5SystemsServer(world, logger)
	initializeV6SystemsServer(world, *seed, logger, economySystem)

	// AUDIT.md Task 3: Wire EnhancedChatSystem Player Registration
	// EnhancedChatSystem provides persistent chat history per player.
	// RegisterPlayer/UnregisterPlayer are now wired to network connection handlers
	// in handlePlayerJoins/handlePlayerLeaves for cross-session history persistence.

	// INTEGRATION FIX [Category A]: V8.0 Server System Initialization
	// Gap: V8.0 systems implemented but never initialized on server
	// Fix: Added V8.0 system initialization call for housing, fluids, vehicle physics, fleet management, trade routes
	// Roadmap: ROADMAP_V8.md (Phase 49-51), ROADMAP_V9.md (Phase 56.1)
	// Returns guild.Manager, FleetManager, and RouteManager for integration and cleanup
	guildManager, fleetManager, tradeRouteManager := initializeV8SystemsServer(world, *seed, *serverName, logger)

	// INTEGRATION FIX [Category A]: V9.0 Server Integration Manager Initialization
	// Gap: V9.0 integration managers were client-only, allowing XP/loyalty/permission exploits
	// Fix: Added V9.0 manager initialization and V9ValidationService for server-authoritative validation
	// Roadmap: ROADMAP_V9.md (Phase 55.1-55.3, Phase 56.3, Phase 58.1)
	stationMgr, petHomeMgr, guildHousingMgr, narrativeWorldSys, politicalWarfareSys := initializeV9SystemsServer(world, *seed, guildManager, logger)
	SetV9ValidationService(NewV9ValidationService(stationMgr, petHomeMgr, guildHousingMgr, logger))

	// INTEGRATION FIX [AUDIT.md REM-018]: Territory Systems Server Integration
	// Gap: TerritorySystem and TerritorySiegeSystem were client-only, enabling exploits
	// Fix: Initialize territory systems on server for authoritative validation
	// Impact: Prevents fake capture progress, war cost bypass, phantom sieges, multiplayer desync
	territoryMgr, territorySys, siegeMgr, siegeSys := initializeTerritorySystemsServer(world, logger)
	_, _, _, _ = territoryMgr, territorySys, siegeMgr, siegeSys // Systems registered via AddSystem in init function

	// INTEGRATION FIX [AUDIT.md Task #6]: Wire HousingCraftingSystem into CraftingSystem
	// Gap: Station bonuses required manual registration, no auto-discovery
	// Fix: Inject StationManager into CraftingSystem for automatic bonus calculation
	// Impact: Players automatically receive crafting bonuses from owned housing stations
	craftingSystem.SetStationManager(stationMgr)

	// INTEGRATION FIX [AUDIT.md Task #9]: Wire CompanionHousingSystem into CompanionLoyaltySystem
	// Gap: Companion housing bonuses required manual PetHomeManager calls
	// Fix: Inject PetHomeManager into CompanionLoyaltySystem for automatic loyalty bonus calculation
	// Impact: Companions automatically receive loyalty bonuses from owned housing (bedding quality)
	companionLoyaltySystem.SetPetHomeProvider(petHomeMgr)

	// INTEGRATION FIX [AUDIT.md Task #10]: Wire NarrativeWorldSystem into NarrativeSystem
	// Gap: Companion story events (quests, memories, conflicts) not integrated with main narrative
	// Fix: Inject narrativeWorldSystem into NarrativeSystem for automatic companion story tracking
	// Impact: Companions automatically generate personal quests and track memories during narrative events
	narrativeSystem.SetCompanionStoryProvider(narrativeWorldSys)

	// INTEGRATION FIX [AUDIT.md Task #12]: PoliticalWarfareSystem for guild-level warfare
	// Gap: Political warfare system implemented but not initialized on server
	// Fix: Initialize PoliticalWarfareSystem in v9_systems.go and add to world ECS
	// Impact: Guild wars, treaties, embargoes, and diplomatic victories enabled server-side
	// Note: Manager accessible via politicalWarfareSys.GetManager() for direct API calls
	_ = politicalWarfareSys // System runs via world.Update(), manager available for future use

	// INTEGRATION FIX [AUDIT.md Task #2]: FleetManager for guild vehicle coordination
	// Gap: FleetManager defined but never instantiated for guild vehicle coordination
	// Fix: Initialize FleetManager in v8_systems.go for server-authoritative fleet management
	// Impact: Guild fleet formations, siege engines, and vehicle maintenance enabled server-side
	// Note: Available for VehicleSystem integration and network packet handling
	_ = fleetManager // Manager available for future vehicle fleet coordination features

	// INTEGRATION FIX [AUDIT.md]: Trade Route Manager for multiplayer caravan sync
	// Gap: Server did not call Start() on RouteManager, unlike client pattern
	// Fix: RouteManager now initialized and started in v8_systems.go
	// Impact: AI merchant caravans and trade routes sync across multiplayer sessions
	// Note: Returned for cleanup during server shutdown via Stop()
	_ = tradeRouteManager // Manager returned for Stop() call during server shutdown

	// INTEGRATION FIX [AUDIT.md Priority 2]: Prestige System Server Registration
	// Gap: Prestige system was client-only, causing prestige data desync in multiplayer
	// Fix: Initialize and register prestige.System on server for authoritative prestige tracking
	// Impact: Prestige levels, paragon points, and prestige abilities sync across multiplayer sessions
	prestigeSystem := prestige.NewSystemWithLogger(logger)
	world.AddSystem(&prestigeSystemWrapper{system: prestigeSystem})
	worldLogger.Debug("prestige system initialized for multiplayer sync")

	// INTEGRATION FIX [AUDIT.md]: QoL System Server Registration
	// Gap: QoL system was client-only, causing craft queue validation issues in multiplayer
	// Fix: Initialize and register QoL system on server for authoritative craft queue validation
	// Impact: Craft queues, guild invitations, and other QoL features properly validated server-side
	qolManager := qol.NewManager(qol.Config{
		AutoLoot:     true, // Server-side loot validation
		AutoSort:     true, // Storage sort validation
		QuickDeposit: true, // Quick deposit validation
	})
	qolSystem := engine.NewQoLSystem(qolManager)
	world.AddSystem(qolSystem)
	worldLogger.Debug("QoL system initialized for multiplayer validation")

	// INTEGRATION FIX [pkg/engine/performance/AUDIT.md]: Performance Monitoring Server Integration
	// Gap: Performance monitoring was client-only, preventing production observability on servers
	// Fix: Initialize and register PerformanceMonitoringSystem on server for production metrics
	// Impact: Server memory usage, tick rate, and network stats now observable for production monitoring
	performanceSystem := engine.NewPerformanceMonitoringSystem()
	world.AddSystem(performanceSystem)
	worldLogger.Debug("performance monitoring system initialized for production observability")

	if logger.GetLevel() >= logrus.DebugLevel {
		worldLogger.Debug("game systems initialized")
	}

	return world, enhancedChatSystem, courierSystem
}

// generateWorldTerrain creates the initial terrain and spawns V4 entities.
func generateWorldTerrain(logger *logrus.Logger, serverLogger *logrus.Entry) *terrain.Terrain {
	terrainLogger := logging.GeneratorLogger(logger, "terrain", *seed, *genreID)
	if logger.GetLevel() >= logrus.DebugLevel {
		terrainLogger.WithFields(logrus.Fields{
			"width":       100,
			"height":      100,
			"terrainType": *terrainType,
		}).Debug("generating world terrain")
	}

	// Select generator based on terrain type
	var terrainGen procgen.Generator
	switch *terrainType {
	case "grammar":
		// Use GraphGrammarGenerator with genre-specific config
		var config terrain.LSystemConfig
		switch *genreID {
		case "fantasy":
			config = terrain.GetFantasyConfig(*seed)
		case "scifi", "sci-fi":
			config = terrain.GetSciFiConfig(*seed)
		case "horror":
			config = terrain.GetHorrorConfig(*seed)
		case "cyberpunk":
			config = terrain.GetCyberpunkConfig(*seed)
		case "post-apocalyptic", "postapocalyptic":
			config = terrain.GetPostApocalypticConfig(*seed)
		default:
			config = terrain.GetFantasyConfig(*seed)
		}
		terrainGen = terrain.NewGraphGrammarGeneratorWithLogger(config, logger)
	case "cellular":
		terrainGen = terrain.NewCellularGeneratorWithLogger(logger)
	case "city":
		terrainGen = terrain.NewCityGeneratorWithLogger(logger)
	case "forest":
		terrainGen = terrain.NewForestGeneratorWithLogger(logger)
	case "composite":
		terrainGen = terrain.NewCompositeGeneratorWithLogger(logger)
	case "maze":
		terrainGen = terrain.NewMazeGeneratorWithLogger(logger)
	case "bsp":
		fallthrough
	default:
		terrainGen = terrain.NewBSPGeneratorWithLogger(logger)
	}

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

	// Convert result to Terrain (handles both *Terrain and *DungeonGraph)
	var generatedTerrain *terrain.Terrain
	if graph, ok := terrainResult.(*terrain.DungeonGraph); ok {
		// Grammar generator returns DungeonGraph, convert it
		generatedTerrain = terrain.GraphToTerrain(graph)
		terrainLogger.WithFields(logrus.Fields{
			"graphRooms": len(graph.Rooms),
			"converted":  true,
		}).Debug("converted DungeonGraph to Terrain")
	} else {
		generatedTerrain = terrainResult.(*terrain.Terrain)
	}

	terrainLogger.WithFields(logrus.Fields{
		"width":       generatedTerrain.Width,
		"height":      generatedTerrain.Height,
		"roomCount":   len(generatedTerrain.Rooms),
		"terrainType": *terrainType,
	}).Info("world terrain generated")

	return generatedTerrain
}

// spawnPostOffices instantiates PostOfficeSpawner and places a post office
// in the terrain so the courier system can deliver mail (AUDIT.md Task 4).
func spawnPostOffices(world *engine.World, generatedTerrain *terrain.Terrain, courierSystem *engine.CourierSystem, logger *logrus.Logger) {
	spawner := engine.NewPostOfficeSpawner(world, courierSystem)
	result, err := spawner.SpawnInTerrain(generatedTerrain, *genreID, *seed)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"system": "post_office_spawning",
			"seed":   *seed,
			"genre":  *genreID,
			"width":  generatedTerrain.Width,
			"height": generatedTerrain.Height,
			"rooms":  len(generatedTerrain.Rooms),
			"error":  err,
		}).Fatal("post office spawning failed; courier system requires at least one post office")
		return
	}
	logger.WithFields(logrus.Fields{
		"building_id": result.BuildingID,
		"clerk_name":  result.ClerkName,
		"x":           result.X,
		"y":           result.Y,
	}).Info("post office spawned")
}

// spawnV4Entities spawns vehicles, companions, and bookshelves in the terrain.
func spawnV4Entities(world *engine.World, generatedTerrain *terrain.Terrain, logger *logrus.Logger) {
	v4Logger := logging.GeneratorLogger(logger, "v4-spawning", *seed, *genreID)
	params := createGenerationParams()

	spawnEnemies(world, generatedTerrain, params, v4Logger)
	spawnMerchants(world, generatedTerrain, params, v4Logger)
	spawnVehicles(world, generatedTerrain, params, logger, v4Logger)
	spawnCompanions(world, generatedTerrain, params, logger, v4Logger)
	spawnBookshelves(world, generatedTerrain, params, logger, v4Logger)
	spawnFactions(world, params, v4Logger)
}

// createGenerationParams creates standard generation parameters for entity spawning.
func createGenerationParams() procgen.GenerationParams {
	return procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    *genreID,
		Custom: map[string]interface{}{
			"width":  100,
			"height": 100,
		},
	}
}

// spawnEnemies spawns enemy entities in the terrain.
func spawnEnemies(world *engine.World, generatedTerrain *terrain.Terrain, params procgen.GenerationParams, v4Logger *logrus.Entry) {
	enemyCount, err := engine.SpawnEnemiesInTerrain(world, generatedTerrain, *seed, params)
	if err != nil {
		v4Logger.WithError(err).Warn("failed to spawn enemies")
	} else if enemyCount > 0 {
		v4Logger.WithField("count", enemyCount).Info("enemies spawned")
	}
}

// spawnMerchants spawns merchant entities in the terrain.
func spawnMerchants(world *engine.World, generatedTerrain *terrain.Terrain, params procgen.GenerationParams, v4Logger *logrus.Entry) {
	merchantCount := 2
	merchantSpawned, err := engine.SpawnMerchantsInTerrain(world, generatedTerrain, *seed, params, merchantCount)
	if err != nil {
		v4Logger.WithError(err).Warn("failed to spawn merchants")
	} else if merchantSpawned > 0 {
		v4Logger.WithField("count", merchantSpawned).Info("merchants spawned")
	}
}

// spawnVehicles spawns vehicle entities in the terrain.
func spawnVehicles(world *engine.World, generatedTerrain *terrain.Terrain, params procgen.GenerationParams, logger *logrus.Logger, v4Logger *logrus.Entry) {
	vehicleCount, err := spawnVehiclesInTerrain(world, generatedTerrain, *seed, params, logger)
	if err != nil {
		v4Logger.WithError(err).Warn("failed to spawn vehicles")
	} else if vehicleCount > 0 {
		v4Logger.WithField("count", vehicleCount).Info("vehicles spawned")
	}
}

// spawnCompanions spawns companion entities in the terrain.
func spawnCompanions(world *engine.World, generatedTerrain *terrain.Terrain, params procgen.GenerationParams, logger *logrus.Logger, v4Logger *logrus.Entry) {
	companionCount, err := spawnCompanionsInTerrain(world, generatedTerrain, *seed, params, logger)
	if err != nil {
		v4Logger.WithError(err).Warn("failed to spawn companions")
	} else if companionCount > 0 {
		v4Logger.WithField("count", companionCount).Info("companions spawned")
	}
}

// spawnBookshelves spawns bookshelf entities in the terrain.
func spawnBookshelves(world *engine.World, generatedTerrain *terrain.Terrain, params procgen.GenerationParams, logger *logrus.Logger, v4Logger *logrus.Entry) {
	bookshelfCount, err := spawnBookshelvesInTerrain(world, generatedTerrain, *seed, params, logger)
	if err != nil {
		v4Logger.WithError(err).Warn("failed to spawn bookshelves")
	} else if bookshelfCount > 0 {
		v4Logger.WithField("count", bookshelfCount).Info("bookshelves spawned")
	}
}

// spawnFactions generates and registers world factions on the server.
// Factions must be generated server-side to ensure authoritative state
// and prevent desync when clients join mid-game or have mod conflicts.
func spawnFactions(world *engine.World, params procgen.GenerationParams, v4Logger *logrus.Entry) {
	// Use faction-specific seed offset for deterministic generation
	const seedOffsetFaction = 3000
	factionCount, err := generateWorldFactions(world, *seed+seedOffsetFaction, params, v4Logger)
	if err != nil {
		v4Logger.WithError(err).Warn("failed to generate factions")
	} else if factionCount > 0 {
		v4Logger.WithField("count", factionCount).Info("factions registered")
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
func startPlayerManagementHandlers(server *network.TCPServer, world *engine.World, generatedTerrain *terrain.Terrain, enhancedChatSystem *engine.EnhancedChatSystem, logger *logrus.Logger) {
	playerEntities := make(map[uint64]*engine.Entity)
	playerEntitiesMu := &sync.RWMutex{}
	networkLogger := logger.WithFields(logrus.Fields{"system": "network"})

	go handlePlayerJoins(server, world, generatedTerrain, playerEntities, playerEntitiesMu, enhancedChatSystem, logger)
	go handlePlayerLeaves(server, world, playerEntities, playerEntitiesMu, enhancedChatSystem, logger)
	go handleInputCommands(server, playerEntities, playerEntitiesMu, networkLogger, logger)
}

// handlePlayerJoins processes new player connections and creates entities.
func handlePlayerJoins(server *network.TCPServer, world *engine.World, generatedTerrain *terrain.Terrain, playerEntities map[uint64]*engine.Entity, playerEntitiesMu *sync.RWMutex, enhancedChatSystem *engine.EnhancedChatSystem, logger *logrus.Logger) {
	for playerID := range server.ReceivePlayerJoin() {
		playerLogger := logging.NetworkLogger(logger, "", "connected").WithField("playerID", playerID)
		playerLogger.Info("player joined - creating entity")

		entity := createPlayerEntity(world, generatedTerrain, playerID, *seed, *genreID, *aerialSprites, logger)

		playerEntitiesMu.Lock()
		playerEntities[playerID] = entity
		playerEntitiesMu.Unlock()

		// Register player with enhanced chat system for persistent history
		if err := enhancedChatSystem.RegisterPlayer(entity.ID); err != nil {
			playerLogger.WithError(err).Warn("failed to register player with chat system")
		} else {
			playerLogger.Debug("player registered with chat system")
		}

		playerLogger.WithField("entityID", entity.ID).Debug("player entity created")
	}
}

// handlePlayerLeaves processes player disconnections and removes entities.
func handlePlayerLeaves(server *network.TCPServer, world *engine.World, playerEntities map[uint64]*engine.Entity, playerEntitiesMu *sync.RWMutex, enhancedChatSystem *engine.EnhancedChatSystem, logger *logrus.Logger) {
	for playerID := range server.ReceivePlayerLeave() {
		playerLogger := logging.NetworkLogger(logger, "", "disconnected").WithField("playerID", playerID)
		playerLogger.Info("player left - removing entity")

		playerEntitiesMu.Lock()
		if entity, exists := playerEntities[playerID]; exists {
			// Unregister player from enhanced chat system
			enhancedChatSystem.UnregisterPlayer(entity.ID)
			playerLogger.Debug("player unregistered from chat system")

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
func runGameLoop(ctx context.Context, world *engine.World, server *network.TCPServer, snapshotManager *network.SnapshotManager, lagCompensator *network.LagCompensator, ticker *time.Ticker, logger *logrus.Logger, serverLogger *logrus.Entry, lastUpdate *time.Time) {
	var performanceSystem *engine.PerformanceMonitoringSystem
	// Extract performance system from world for metrics logging
	for _, system := range world.GetSystems() {
		if perfSys, ok := system.(*engine.PerformanceMonitoringSystem); ok {
			performanceSystem = perfSys
			break
		}
	}

	for {
		select {
		case <-ctx.Done():
			serverLogger.Info("game loop stopping due to shutdown signal")
			return
		case <-ticker.C:
			now := time.Now()
			deltaTime := now.Sub(*lastUpdate).Seconds()
			*lastUpdate = now

			world.Update(deltaTime)

			snapshot := buildWorldSnapshot(world, now)
			snapshotManager.AddSnapshot(snapshot)
			lagCompensator.RecordSnapshot(snapshot)

			stateUpdates := convertSnapshotToStateUpdates(snapshot)
			for _, update := range stateUpdates {
				server.BroadcastStateUpdate(update)
			}

			// Log server metrics periodically (every 10 seconds)
			if logger.GetLevel() >= logrus.DebugLevel && int(now.Unix())%10 == 0 {
				playerCount := server.GetPlayerCount()
				logFields := logrus.Fields{
					"entityCount": len(world.GetEntities()),
					"playerCount": playerCount,
				}

				// Add performance metrics if available
				if performanceSystem != nil {
					logFields["tick_rate"] = performanceSystem.GetFPS()
					logFields["tick_time_ms"] = performanceSystem.GetFrameTime()
					logFields["memory_mb"] = performanceSystem.GetMemoryUsageMB()
				}

				serverLogger.WithFields(logFields).Debug("server tick metrics")
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

	for _, entity := range world.GetEntities() {
		if entitySnap := buildEntitySnapshot(entity); entitySnap != nil {
			snapshot.Entities[entity.ID] = *entitySnap
		}
	}

	return snapshot
}

// buildEntitySnapshot creates a network snapshot for a single entity.
func buildEntitySnapshot(entity *engine.Entity) *network.EntitySnapshot {
	posComp, ok := entity.GetComponent("position")
	if !ok {
		return nil
	}

	pos := posComp.(*engine.PositionComponent)
	velX, velY := extractVelocity(entity)

	entitySnapshot := network.EntitySnapshot{
		EntityID:   entity.ID,
		Position:   network.Position{X: pos.X, Y: pos.Y},
		Velocity:   network.Velocity{VX: velX, VY: velY},
		Components: make(map[string][]byte),
	}

	serializeEntityComponents(entity, &entitySnapshot)
	return &entitySnapshot
}

// extractVelocity retrieves velocity from entity or returns zero values.
func extractVelocity(entity *engine.Entity) (float64, float64) {
	if velComp, ok := entity.GetComponent("velocity"); ok {
		vel := velComp.(*engine.VelocityComponent)
		return vel.VX, vel.VY
	}
	return 0.0, 0.0
}

// serializeEntityComponents serializes all network-relevant components.
func serializeEntityComponents(entity *engine.Entity, snapshot *network.EntitySnapshot) {
	serializeIfPresent(entity, snapshot, "vehicle", func(c interface{}) []byte {
		return c.(*engine.VehicleComponent).Serialize()
	})
	serializeIfPresent(entity, snapshot, "companion", func(c interface{}) []byte {
		return c.(*engine.CompanionComponent).Serialize()
	})
	serializeIfPresent(entity, snapshot, "mount", func(c interface{}) []byte {
		return c.(*engine.MountComponent).Serialize()
	})
	serializeIfPresent(entity, snapshot, "achievement", func(c interface{}) []byte {
		return c.(*engine.AchievementComponent).Serialize()
	})
	serializeIfPresent(entity, snapshot, "bookshelf", func(c interface{}) []byte {
		return c.(*engine.BookshelfComponent).Serialize()
	})
}

// serializeIfPresent serializes a component if it exists on the entity.
func serializeIfPresent(entity *engine.Entity, snapshot *network.EntitySnapshot, compType string, serializer func(interface{}) []byte) {
	if comp, ok := entity.GetComponent(compType); ok {
		snapshot.Components[compType] = serializer(comp)
	}
}

// convertSnapshotToStateUpdates converts a WorldSnapshot to StateUpdates for broadcasting.
// Returns one StateUpdate per entity with full component data.
func convertSnapshotToStateUpdates(snapshot network.WorldSnapshot) []*network.StateUpdate {
	updates := make([]*network.StateUpdate, 0, len(snapshot.Entities))
	timestampMs := uint64(snapshot.Timestamp.UnixNano() / 1000000)

	for entityID, entitySnap := range snapshot.Entities {
		// Build ComponentData slice from serialized components
		components := make([]network.ComponentData, 0, len(entitySnap.Components)+2)

		// Add position as a component (always present since we build snapshots from entities with position)
		posData := serializePosition(entitySnap.Position)
		components = append(components, network.ComponentData{Type: "position", Data: posData})

		// Add velocity as a component
		velData := serializeVelocity(entitySnap.Velocity)
		components = append(components, network.ComponentData{Type: "velocity", Data: velData})

		// Add all other serialized components
		for compType, compData := range entitySnap.Components {
			components = append(components, network.ComponentData{Type: compType, Data: compData})
		}

		update := &network.StateUpdate{
			Timestamp:  timestampMs,
			EntityID:   entityID,
			Components: components,
			Priority:   network.PriorityNormal,
		}
		updates = append(updates, update)
	}

	return updates
}

// serializePosition serializes a network.Position to bytes for network transmission.
func serializePosition(pos network.Position) []byte {
	// Simple binary encoding: 8 bytes for X + 8 bytes for Y = 16 bytes
	data := make([]byte, 16)
	putFloat64(data[0:8], pos.X)
	putFloat64(data[8:16], pos.Y)
	return data
}

// serializeVelocity serializes a network.Velocity to bytes for network transmission.
func serializeVelocity(vel network.Velocity) []byte {
	// Simple binary encoding: 8 bytes for VX + 8 bytes for VY = 16 bytes
	data := make([]byte, 16)
	putFloat64(data[0:8], vel.VX)
	putFloat64(data[8:16], vel.VY)
	return data
}

// putFloat64 encodes a float64 to bytes using little-endian IEEE 754 format.
func putFloat64(b []byte, v float64) {
	binary.LittleEndian.PutUint64(b, math.Float64bits(v))
}

// startStabilityMonitoring initializes and starts the stability monitor for production validation.
// Phase 2.3 (PLAN.md): Unconditional stability package integration
func startStabilityMonitoring(ctx context.Context, serverLogger *logrus.Entry) *stability.Monitor {
	config := stability.DefaultConfig()
	// Use shorter duration for non-production testing
	config.Duration = 24 * time.Hour
	config.CheckInterval = 60 * time.Second

	monitor := stability.NewMonitor(config)

	// Start the monitor in a background goroutine to perform actual health checks
	go func() {
		report, err := monitor.Run(ctx)
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

// initializeMetricsExporter sets up the Prometheus metrics HTTP endpoint.
// Phase 3 (PLAN.md): Metrics export for production monitoring
// Returns error if metrics server cannot start, allowing caller to decide whether to fail fast.
func initializeMetricsExporter(logger *logrus.Logger, world *engine.World, server *network.TCPServer) (*observability.MetricsExporter, error) {
	metricsLogger := logger.WithFields(logrus.Fields{"component": "metrics"})

	addr := ":" + *metricsPort
	exporter := observability.NewMetricsExporterWithLogger(addr, logger)

	// Find and register performance monitoring system from world
	var perfMon *engine.PerformanceMonitoringSystem
	for _, system := range world.GetSystems() {
		if pms, ok := system.(*engine.PerformanceMonitoringSystem); ok {
			perfMon = pms
			break
		}
	}

	// If no performance monitoring system exists, create one
	if perfMon == nil {
		metricsLogger.Info("creating performance monitoring system for metrics")
		perfMon = engine.NewPerformanceMonitoringSystem()
		world.AddSystem(perfMon)
	}

	// Register metrics sources
	exporter.RegisterPerformanceMonitor(perfMon)
	exporter.RegisterNetworkServer(server)
	exporter.RegisterWorld(world)

	// Start metrics HTTP server
	if err := exporter.Start(); err != nil {
		metricsLogger.WithError(err).Error("failed to start metrics exporter")
		return nil, fmt.Errorf("metrics exporter start failed: %w", err)
	}

	metricsLogger.WithFields(logrus.Fields{
		"port":     *metricsPort,
		"endpoint": "/metrics",
	}).Info("Prometheus metrics exporter started")

	return exporter, nil
}
