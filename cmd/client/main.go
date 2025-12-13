//go:build !android && !ios
// +build !android,!ios

// Package main provides the desktop client application.
// For mobile platforms (Android/iOS), use cmd/mobile with ebitenmobile build tool.
package main

import (
	"flag"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/mobile"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// autoEnabledHostAndPlay tracks whether host-and-play was auto-enabled (vs explicitly requested).
var autoEnabledHostAndPlay bool

func main() {
	flag.Parse()

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
func setupAllGameSystems(game *engine.EbitenGame, logger *logrus.Logger, clientLogger *logrus.Entry) *systemsContainer {
	sys := initializeCoreSystems(game, logger, clientLogger)

	initializeVirtualControls(sys.inputSystem, clientLogger)
	initializeGenerators(sys)
	initializeObjectiveSystem(sys, logger)

	configureDeathCallback(sys, game, logger)

	initializeProgressionSystems(game, sys, logger)
	initializeAudioSystem(game, sys, clientLogger)
	initializeCombatSystems(game, sys)

	tutorialSystem, helpSystem := initializeTutorialAndHelp(sys.inputSystem, game.CameraSystem)

	// CRITICAL: Initialize environmental systems BEFORE registering them
	initializeEnvironmentalSystems(game, sys, clientLogger)

	// CRITICAL: Initialize V4.0 systems (Phase 21-27) BEFORE registering them
	initializeV4Systems(game, sys, clientLogger)

	// CRITICAL: Initialize V5.0 systems (social, chat, mail) BEFORE registering them
	initializeV5Systems(game, sys, clientLogger)

	// CRITICAL: Initialize V6.0 systems (federation, persistent worlds) BEFORE registering them
	initializeV6Systems(game, sys, clientLogger)

	// INTEGRATION FIX [Category A]: Initialize V8.0 systems (housing, physics, social persistence)
	// Gap: V8.0 features fully implemented but never initialized in game client
	// Fix: Added V8.0 system initialization call before registration
	// Roadmap: ROADMAP_V8.md (Phase 49-51)
	initializeV8Systems(game, sys, clientLogger)

	// INTEGRATION FIX [Category A]: Initialize V7.0 systems (display, viewport)
	// Gap: V7.0 features fully implemented but never initialized in game client
	// Fix: Added V7.0 system initialization call for display and viewport optimization
	// Roadmap: ROADMAP_V7.md (Phase 43-44)
	initializeV7Systems(game, sys, clientLogger)

	// INTEGRATION FIX [Category A]: Initialize V9.0 integration managers
	// Gap: V9.0 integration packages fully implemented but never initialized
	// Fix: Added V9.0 manager initialization call for housing/companion/guild integration
	// Roadmap: ROADMAP_V9.md (Phase 55.1-55.3)
	initializeV9Systems(game, sys, clientLogger)

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
		clientLogger.Info("systems initialized")
	}

	return sys
}

// setupWorldTerrain generates terrain, initializes rendering, lighting, and collision.
func setupWorldTerrain(game *engine.EbitenGame, sys *systemsContainer, logger *logrus.Logger, clientLogger *logrus.Entry) *terrain.Terrain {
	generatedTerrain := generateWorldTerrain(logger, clientLogger)
	initializeTerrainRendering(game, generatedTerrain, clientLogger)
	configureLightingSystem(game, clientLogger)
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
