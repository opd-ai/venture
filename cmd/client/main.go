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
	"github.com/opd-ai/venture/pkg/procgen/genre"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/version"
	"github.com/sirupsen/logrus"
)

// autoEnabledHostAndPlay tracks whether host-and-play was auto-enabled (vs explicitly requested).
var autoEnabledHostAndPlay bool

func main() {
	flag.Parse()

	// Resolve "random" genre to concrete genre using seed for determinism
	if *genreID == "random" {
		selectedGenre := genre.GetRandomTheme(*seed)
		*genreID = selectedGenre.ID
	}

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

	// BUG FIX: Skip network client initialization in WASM single-player mode.
	// In WASM, host-and-play is disabled (no network listen in browser sandbox),
	// so if *multiplayer is not set, there is no server to connect to.
	// Attempting to connect would fail with a Fatal error, crashing the game.
	var networkClient interface{}
	if *multiplayer || *hostAndPlay {
		nc := initializeNetworkClient(logger, clientLogger)
		defer cleanupNetworkClient(nc, clientLogger)
		networkClient = nc
	} else {
		clientLogger.Info("running in offline single-player mode - no network connection")
	}

	game := createGameInstance(logger, clientLogger)
	sys := setupAllGameSystems(game, logger, clientLogger)
	startPerformanceMonitoring(game, clientLogger)

	// Wire voice transport to network client (AUDIT.md Gap 1: VoiceTransport never wired)
	initializeVoiceTransport(sys, networkClient, clientLogger)

	// Performance Audit Fix: Start async terrain generation and show loading screen
	// instead of blocking main thread. This prevents 2-8s freeze for large terrains.
	startAsyncTerrainGeneration(game, logger, clientLogger)

	// Set up callbacks for when terrain loading completes
	setupAsyncTerrainCallbacks(game, sys, networkClient, logger, clientLogger)

	// Transition to loading state and start game loop immediately
	// The loading screen will be shown while terrain generates in background
	if err := game.StateManager.TransitionTo(engine.AppStateLoading); err != nil {
		clientLogger.WithError(err).Fatal("failed to transition to loading state")
	}

	runGameLoop(game, clientLogger)
}

// setupAllGameSystems initializes and registers all game systems.
// Uses parallel initialization for independent system groups to reduce startup time.
// Non-critical systems are deferred for lazy initialization after first frame.
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

	// Phase 2: Core gameplay systems (critical for first frame)
	initializeProgressionSystems(game, sys, logger)
	initializeCombatSystems(game, sys)

	// Read ShowTutorials setting to wire it into tutorial initialization
	showTutorials := true
	if game.SettingsManager != nil {
		showTutorials = game.SettingsManager.GetSettings().ShowTutorials
	}
	tutorialSystem, helpSystem, contextualTutorial := initializeTutorialAndHelp(sys.inputSystem, game.CameraSystem, showTutorials, *width, *height)

	// Register critical systems immediately for first frame
	registerCriticalSystems(game, sys)

	game.World.AddSystem(tutorialSystem)
	game.World.AddSystem(helpSystem)

	// Phase 3.2: Wire tutorial system to onboarding manager for seamless transitions
	game.SetTutorialSystemForOnboarding(tutorialSystem)
	game.HelpSystem = helpSystem

	// Phase 3.3: Wire context-sensitive tutorial manager to game for settings propagation
	game.SetContextualTutorial(contextualTutorial)

	// Apply --no-tutorial flag: skip the entire onboarding flow so the state
	// machine reaches StateComplete and ApplySettings cannot re-enable tutorials.
	// The in-game tutorial and contextual tutorial are already handled in initializeTutorialAndHelp.
	if *noTutorial {
		if game.OnboardingManager != nil {
			game.OnboardingManager.SkipAll()
		} else {
			// Fallback: disable systems directly when no onboarding manager
			if game.CharacterCreationTutorial != nil {
				game.CharacterCreationTutorial.SkipTutorial()
			}
		}
	}

	configureSystemConnections(game, sys)

	if *verbose {
		clientLogger.Info("critical systems initialized - deferring non-critical systems")
	}

	// Schedule lazy initialization for non-critical systems (deferred until after first frame)
	sys.scheduleLazyInit(game, logger, clientLogger)

	return sys
}

// startAsyncTerrainGeneration begins terrain generation in background and returns the loader.
// Performance Audit: This prevents blocking the main thread during 2-8s terrain generation.
func startAsyncTerrainGeneration(game *engine.EbitenGame, logger *logrus.Logger, clientLogger *logrus.Entry) *terrain.AsyncLoader {
	clientLogger.Info("initiating async terrain generation")

	// Start terrain generation in background
	loader := generateWorldTerrain(logger, clientLogger)

	// Store loader in game for progress tracking
	game.SetTerrainLoader(loader)

	return loader
}

// setupAsyncTerrainCallbacks configures the callback that runs when terrain loading completes.
// This callback performs all world initialization that depends on terrain being ready.
func setupAsyncTerrainCallbacks(game *engine.EbitenGame, sys *systemsContainer, networkClient interface{}, logger *logrus.Logger, clientLogger *logrus.Entry) {
	game.SetTerrainLoadCompleteCallback(func(player *engine.Entity) error {
		clientLogger.Info("terrain loaded, completing world initialization")

		// Get terrain loader reference (stored as interface{})
		loaderInterface := game.GetTerrainLoader()
		if loaderInterface == nil {
			return fmt.Errorf("terrain loader is nil")
		}

		loader, ok := loaderInterface.(*terrain.AsyncLoader)
		if !ok {
			return fmt.Errorf("invalid terrain loader type")
		}

		generatedTerrain, err := loader.Wait()
		if err != nil {
			return fmt.Errorf("failed to get terrain result: %w", err)
		}

		clientLogger.WithFields(logrus.Fields{
			"width":     generatedTerrain.Width,
			"height":    generatedTerrain.Height,
			"roomCount": len(generatedTerrain.Rooms),
		}).Info("terrain generated successfully")

		// Now complete all terrain-dependent initialization
		completeWorldInitialization(game, sys, generatedTerrain, logger, clientLogger)

		// Create player entity
		playerEntity := setupCompletePlayerEntity(game, generatedTerrain, sys, logger, clientLogger)
		game.PlayerEntity = playerEntity

		// Setup UI with player and terrain
		setupGameUI(game, playerEntity, generatedTerrain, sys, clientLogger)

		// Rebuild spatial partition after all entities created
		if spatialSystem := game.RenderSystem.GetSpatialPartition(); spatialSystem != nil {
			spatialSystem.Rebuild(game.World.GetEntities())
			if *verbose {
				clientLogger.WithField("entityCount", len(game.World.GetEntities())).Info("spatial partition rebuilt")
			}
		}

		// Finalize initialization
		finalizeGameInitialization(game, playerEntity, networkClient, clientLogger)

		return nil
	})
}

// completeWorldInitialization performs all terrain-dependent setup.
func completeWorldInitialization(game *engine.EbitenGame, sys *systemsContainer, generatedTerrain *terrain.Terrain, logger *logrus.Logger, clientLogger *logrus.Entry) {
	// Initialize terrain rendering and lighting
	initializeTerrainRendering(game, generatedTerrain, clientLogger)
	configureLightingSystem(game, clientLogger)

	// Phase 5.3: Initialize and configure post-processing system
	game.PostProcessor = engine.NewPostProcessorAdapter(clientLogger)
	if err := game.PostProcessor.PrecompileShaders(); err != nil {
		clientLogger.WithField("error", err.Error()).Warn("shader pre-compilation failed, will compile on first use")
	}
	configurePostProcessing(game, clientLogger)

	// Phase 5.4: Configure palette options for sprite generation
	configurePaletteOptions(sys, clientLogger)

	// Initialize collision and spatial systems
	initializeTerrainCollision(game, sys, generatedTerrain, clientLogger)

	// Generate factions
	params := createGenerationParams()
	generateWorldFactions(game, params, clientLogger)

	// Setup spatial partitioning and map UI
	initializeSpatialPartitioning(game, sys, generatedTerrain, clientLogger)
	connectMapUIToTerrain(game, generatedTerrain, clientLogger)

	// Spawn entities and effects
	spawnWorldEntities(game, generatedTerrain, sys, clientLogger)
	spawnEnvironmentalEffects(game, generatedTerrain, clientLogger)
}

// setupCompletePlayerEntity creates player, adds components, and applies character class.
func setupCompletePlayerEntity(game *engine.EbitenGame, generatedTerrain *terrain.Terrain, sys *systemsContainer, logger *logrus.Logger, clientLogger *logrus.Entry) *engine.Entity {
	playerX, playerY := calculatePlayerSpawnPosition(generatedTerrain, clientLogger)
	player := createPlayerEntity(game, playerX, playerY, sys.animationSystem, clientLogger)

	// Set camera bounds to terrain pixel dimensions so the viewport never
	// scrolls past the terrain edges into void.
	terrainWidthPx := float64(generatedTerrain.Width * tileSize)
	terrainHeightPx := float64(generatedTerrain.Height * tileSize)

	if cameraComp, ok := player.GetComponent("camera"); ok {
		if cam, ok := cameraComp.(*engine.CameraComponent); ok {
			engine.SetCameraBoundsFromTerrain(cam, terrainWidthPx, terrainHeightPx, game.ScreenWidth, game.ScreenHeight)
		}
	}

	// Add player movement bounds so the player cannot walk off the terrain.
	player.AddComponent(&engine.BoundsComponent{
		MinX: 0,
		MinY: 0,
		MaxX: terrainWidthPx,
		MaxY: terrainHeightPx,
	})

	addPlayerComponents(player, logger, clientLogger)
	applyCharacterClass(player, game, clientLogger)
	initializePlayerAdvancedClass(player, game, clientLogger)

	// Mark character creation tutorial as complete on the player entity
	// so returning players (load game) will skip the tutorial
	if game.CharacterCreationTutorial != nil {
		skipped := game.CharacterCreationTutorial.Skipped
		engine.MarkCreationTutorialComplete(player, skipped)
	}

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
