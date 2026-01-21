//go:build !android && !ios
// +build !android,!ios

// Package main provides the desktop client application.
// For mobile platforms (Android/iOS), use cmd/mobile with ebitenmobile build tool.
package main

import (
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/mobile"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/version"
	"github.com/sirupsen/logrus"
)

// autoEnabledHostAndPlay tracks whether host-and-play was auto-enabled (vs explicitly requested).
var autoEnabledHostAndPlay bool

func main() {
	flag.Parse()

	// Handle --version flag
	if *showVersion {
		version.PrintVersion()
		os.Exit(0)
	}

	// Validate configuration before starting client
	if err := validateClientConfiguration(); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	logger, clientLogger := initializeLogger()

	// BUG FIX: Phase 1 - WASM cannot run embedded server (no network listen in browser)
	// Resolution: Skip host-and-play initialization on WASM platform
	// Platform: WASM (all browsers)
	// WASM multiplayer requires connecting to external WebSocket server (WSS for HTTPS sites)
	if mobile.IsWASM() {
		// On WASM, disable host-and-play and only allow explicit server connection
		*hostAndPlay = false
		if !*multiplayer {
			clientLogger.Info("WASM build - running in single-player mode (embedded server not available in browser)")
			clientLogger.Info("For multiplayer, specify a server with --server flag")
		}
	} else {
		// Auto-enable host-and-play when no explicit server connection is specified
		// This makes localhost server the default behavior instead of single-player mode
		if !*multiplayer && !*hostAndPlay {
			*hostAndPlay = true
			*hostLAN = false // Force localhost binding for implicit mode
			autoEnabledHostAndPlay = true
			clientLogger.Info("no server specified - defaulting to local host-and-play mode on 127.0.0.1")
		}
	}

	serverCleanup := handleHostAndPlay(logger, clientLogger)
	defer serverCleanup() // Cleanup server when application exits

	networkClient := initializeNetworkClient(logger, clientLogger)
	defer cleanupNetworkClient(networkClient, clientLogger)

	game := createGameInstance(logger, clientLogger)
	sys := setupAllGameSystems(game, logger, clientLogger)
	startPerformanceMonitoring(game, clientLogger)

	generatedTerrain := setupWorldTerrain(game, sys, logger, clientLogger)
	params := createGenerationParams()
	generateWorldFactions(game, params, clientLogger)
	initializeSpatialPartitioning(game, generatedTerrain, clientLogger)
	connectMapUIToTerrain(game, generatedTerrain, clientLogger)

	spawnWorldEntities(game, generatedTerrain, sys, clientLogger)
	spawnEnvironmentalEffects(game, generatedTerrain, clientLogger)

	player := setupCompletePlayerEntity(game, generatedTerrain, sys, logger, clientLogger)
	setupGameUI(game, player, generatedTerrain, sys, clientLogger)

	// CRITICAL FIX: Rebuild spatial partition after ALL entities are created
	// The initial rebuild happens before player and enemies are spawned,
	// so we need to rebuild again to include them in the quadtree for culling
	if spatialSystem := game.RenderSystem.GetSpatialPartition(); spatialSystem != nil {
		spatialSystem.Rebuild(game.World.GetEntities())
		if *verbose {
			clientLogger.WithField("entityCount", len(game.World.GetEntities())).Info("spatial partition rebuilt after all entities spawned")
		}
	}

	finalizeGameInitialization(game, player, networkClient, clientLogger)
	runGameLoop(game, clientLogger)
}

// setupAllGameSystems initializes and registers all game systems.
// Uses parallel initialization for independent system groups to reduce startup time.
func setupAllGameSystems(game *engine.EbitenGame, logger *logrus.Logger, clientLogger *logrus.Entry) *systemsContainer {
	sys := initializeCoreSystems(game, logger, clientLogger)

	// Phase 1: Independent early initialization (parallel)
	var wg1 sync.WaitGroup
	wg1.Add(3)
	go func() {
		defer wg1.Done()
		initializeVirtualControls(sys.inputSystem, clientLogger)
	}()
	go func() {
		defer wg1.Done()
		initializeGenerators(sys)
	}()
	go func() {
		defer wg1.Done()
		initializeObjectiveSystem(sys, logger)
	}()
	wg1.Wait()

	configureDeathCallback(sys, game, logger)

	// Phase 2: Progression, audio, and combat systems (sequential - depend on generators)
	initializeProgressionSystems(game, sys, logger)
	initializeAudioSystem(game, sys, clientLogger)
	initializeCombatSystems(game, sys)

	tutorialSystem, helpSystem := initializeTutorialAndHelp(sys.inputSystem, game.CameraSystem)

	// Phase 3: Version systems (parallel - independent of each other)
	// These systems write to separate fields in systemsContainer, making them safe to parallelize
	var wg2 sync.WaitGroup
	wg2.Add(8)
	go func() {
		defer wg2.Done()
		initializeEnvironmentalSystems(game, sys, clientLogger)
	}()
	go func() {
		defer wg2.Done()
		initializeV4Systems(game, sys, clientLogger)
	}()
	go func() {
		defer wg2.Done()
		initializeV5Systems(game, sys, clientLogger)
	}()
	go func() {
		defer wg2.Done()
		initializeV6Systems(game, sys, clientLogger)
	}()
	go func() {
		defer wg2.Done()
		initializeV7Systems(game, sys, clientLogger)
	}()
	go func() {
		defer wg2.Done()
		initializeV8Systems(game, sys, clientLogger)
	}()
	go func() {
		defer wg2.Done()
		initializeV9Systems(game, sys, clientLogger)
	}()
	go func() {
		defer wg2.Done()
		initializeV19Systems(game, sys, clientLogger)
	}()
	wg2.Wait()

	// Phase 3.2: Initialize Guild Federation (PLAN.md)
	// Cross-server guild management with synchronization
	initializePhase3Systems(game, sys, clientLogger)

	// Now register all systems (including environmental systems that are now initialized)
	registerAllSystems(game, sys)

	game.World.AddSystem(tutorialSystem)
	game.World.AddSystem(helpSystem)

	game.TutorialSystem = tutorialSystem
	game.HelpSystem = helpSystem

	configureSystemConnections(game, sys)

	if *verbose {
		clientLogger.Info("systems initialized (parallel mode)")
	}

	return sys
}

// setupWorldTerrain generates terrain, initializes rendering, lighting, and collision.
func setupWorldTerrain(game *engine.EbitenGame, sys *systemsContainer, logger *logrus.Logger, clientLogger *logrus.Entry) *terrain.Terrain {
	generatedTerrain := generateWorldTerrain(logger, clientLogger)
	initializeTerrainRendering(game, generatedTerrain, clientLogger)
	configureLightingSystem(game, clientLogger)

	// Phase 5.3: Initialize and configure post-processing system
	game.PostProcessor = engine.NewPostProcessorAdapter(clientLogger)
	configurePostProcessing(game, clientLogger)

	// Phase 5.4: Configure palette options for sprite generation
	configurePaletteOptions(sys, clientLogger)

	initializeTerrainCollision(game, sys, generatedTerrain, clientLogger)

	return generatedTerrain
}

// setupCompletePlayerEntity creates player, adds components, and applies character class.
func setupCompletePlayerEntity(game *engine.EbitenGame, generatedTerrain *terrain.Terrain, sys *systemsContainer, logger *logrus.Logger, clientLogger *logrus.Entry) *engine.Entity {
	playerX, playerY := calculatePlayerSpawnPosition(generatedTerrain, clientLogger)
	player := createPlayerEntity(game, playerX, playerY, sys.animationSystem, clientLogger)

	addPlayerComponents(player, logger, clientLogger)
	applyCharacterClass(player, game, clientLogger)
	initializePlayerAdvancedClass(player, game, clientLogger)

	return player
}

// setupGameUI initializes all UI systems and input callbacks.
func setupGameUI(game *engine.EbitenGame, player *engine.Entity, generatedTerrain *terrain.Terrain, sys *systemsContainer, clientLogger *logrus.Entry) {
	// INTEGRATION FIX [Category B]: Pass sys to enable V8 UI integration
	// Gap: V8 UI systems need access to managers (housing, gallery, etc.)
	// Fix: Pass systemsContainer to initializeUIIntegration for V8 UI setup
	// Roadmap: ROADMAP_V8.md Phase 49
	shopUI, _ := initializeUIIntegration(game, player, sys.commerceSystem, sys.dialogSystem, sys.craftingSystem, sys.inventorySystem, sys, clientLogger)

	saveManager := configureSaveLoadSystem(player, game, generatedTerrain, sys.inputSystem, clientLogger)

	if err := setupUICallbacks(game, player, generatedTerrain, sys.inputSystem, sys.objectiveTracker, sys.dialogSystem, shopUI, saveManager, clientLogger); err != nil {
		clientLogger.WithError(err).Fatal("failed to setup UI callbacks")
	}
}
