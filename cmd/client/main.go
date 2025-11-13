//go:build !android && !ios
// +build !android,!ios

// Package main provides the desktop client application.
// For mobile platforms (Android/iOS), use cmd/mobile with ebitenmobile build tool.
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/opd-ai/venture/pkg/combat"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/mobile"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/faction"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/quest"
	"github.com/opd-ai/venture/pkg/procgen/recipe"
	"github.com/opd-ai/venture/pkg/procgen/station"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/opd-ai/venture/pkg/saveload"
	"github.com/opd-ai/venture/pkg/version"
	"github.com/sirupsen/logrus"
)

func main() {
	flag.Parse()

	// Initialize structured logger
	logger, clientLogger := initializeLogger()

	// Handle host-and-play mode: start embedded server before client
	if *hostAndPlay {
		clientLogger.Info("host-and-play mode enabled - starting embedded server")

		// Start embedded server
		serverAddr, cleanup, err := startEmbeddedServer(logger, *seed, *genreID)
		if err != nil {
			clientLogger.WithError(err).Fatal("failed to start embedded server")
		}
		defer cleanup() // Ensure cleanup on exit

		// Override server flag to connect to embedded server
		*server = serverAddr
		*multiplayer = true

		clientLogger.WithField("serverAddr", serverAddr).Info("embedded server started, connecting client")
	}

	// Initialize network client if multiplayer mode is enabled
	networkClient := initializeNetworkClient(logger, clientLogger)

	// Create the game instance
	game := engine.NewEbitenGameWithLogger(*width, *height, logger)

	// Enable performance profiling if requested
	if *profile {
		game.EnableFrameTimeProfiling()
		clientLogger.Info("performance profiling enabled - frame time stats will be logged every 5 seconds")
	}

	// Initialize game systems
	clientLogger.Info("initializing game systems")

	// Add core gameplay systems
	inputSystem := engine.NewInputSystem()

	// WASM TOUCH FIX: Initialize virtual controls immediately for touch-capable platforms
	// This ensures controls are visible on page load rather than waiting for first touch
	if mobile.IsTouchCapable() {
		inputSystem.InitializeVirtualControls(*width, *height)
		clientLogger.WithFields(logrus.Fields{
			"platform": mobile.GetPlatform().String(),
			"width":    *width,
			"height":   *height,
		}).Info("virtual controls initialized for touch-capable platform")
	}

	// GAP-001 & GAP-002 REPAIR: Use proper constructors with required parameters
	movementSystem := engine.NewMovementSystem(200.0)  // 200 units/second max speed
	collisionSystem := engine.NewCollisionSystem(64.0) // 64-unit grid cells for spatial partitioning

	// GAP-001 REPAIR: Connect collision system to movement system for predictive collision
	movementSystem.SetCollisionSystem(collisionSystem)

	combatSystem := engine.NewCombatSystemWithLogger(*seed, logger)

	// Phase 11.2: Initialize interaction system for puzzle element interactions
	interactionSystem := engine.NewInteractionSystem(game.World)

	// GAP-016 REPAIR: Initialize particle system for visual effects
	particleSystem := engine.NewParticleSystem()

	// GAP-017 REPAIR: Initialize animation system for animated sprites
	spriteGenerator := sprites.NewGenerator()
	animationSystem := engine.NewAnimationSystem(spriteGenerator)

	// WASM OPTIMIZATION: Increase animation cache size for better performance
	// Larger cache (300 vs default 100) reduces sprite regeneration in browser environments
	// Each sequence ~100-400KB, total cache ~30-120MB which is acceptable for modern browsers
	animationSystem.SetMaxCacheSize(300)

	// Category 5.2: Initialize equipment visual system for showing equipped items on sprites
	equipmentVisualSystem := engine.NewEquipmentVisualSystem(spriteGenerator)

	// Store player reference for death callback (will be set after player creation)
	var playerEntity *engine.Entity

	// Store audio manager reference for callbacks (will be set after audio system creation)
	var audioManager *engine.AudioManager

	// GAP-004 REPAIR: Add objective tracker system for quest progress
	objectiveTracker := engine.NewObjectiveTrackerSystem()

	// Initialize generators for loot and recipe drops
	itemGen := item.NewItemGenerator()
	recipeGen := recipe.NewRecipeGenerator()

	// Set quest completion callback to award rewards
	objectiveTracker.SetQuestCompleteCallback(func(entity *engine.Entity, qst *quest.Quest) {
		objectiveTracker.AwardQuestRewards(entity, qst)
		if logger.GetLevel() >= logrus.InfoLevel {
			logging.ComponentLogger(logger, "quest").WithFields(logrus.Fields{
				"questName":   qst.Name,
				"xpReward":    qst.Reward.XP,
				"goldReward":  qst.Reward.Gold,
				"skillPoints": qst.Reward.SkillPoints,
			}).Info("quest completed")
		}
	})

	// GAP-001 & GAP-004 REPAIR: Set death callback for loot drops and quest tracking
	combatSystem.SetDeathCallback(createDeathCallback(
		game, &playerEntity, objectiveTracker, &audioManager,
		recipeGen, *seed, *genreID, logger,
	))

	aiSystem := engine.NewAISystem(game.World)
	progressionSystem := engine.NewProgressionSystem(game.World)
	inventorySystem := engine.NewInventorySystem(game.World)

	// GAP-004 REPAIR: Initialize commerce and dialog systems
	commerceSystem := engine.NewCommerceSystemWithLogger(game.World, inventorySystem, logger)
	dialogSystem := engine.NewDialogSystemWithLogger(game.World, logger)

	// Initialize crafting system (itemGen and recipeGen already initialized earlier)
	craftingSystem := engine.NewCraftingSystem(game.World, inventorySystem, itemGen)

	logging.ComponentLogger(logger, "commerce").Info("commerce system initialized")
	logging.ComponentLogger(logger, "dialog").Info("dialog system initialized")
	logging.ComponentLogger(logger, "crafting").Info("crafting system initialized")

	// GAP-010 REPAIR: Initialize audio system
	audioManager = engine.NewAudioManager(44100, *seed) // 44.1kHz sample rate
	audioManagerSystem := engine.NewAudioManagerSystem(audioManager)

	// Wire audio manager to game for settings integration
	game.SetAudioManager(audioManager)

	// Phase 12.3: Enable adaptive music composition with motif system
	audioManager.EnableAdaptiveMusic(true)
	if *verbose {
		clientLogger.Info("adaptive music composition enabled with motif system")
	}

	// Start playing exploration music
	if err := audioManager.PlayMusic(*genreID, "exploration"); err != nil {
		logging.ComponentLogger(logger, "audio").WithError(err).Warn("failed to start background music")
	}

	logging.ComponentLogger(logger, "audio").Info("audio system initialized (music and SFX generators)")

	// Add item pickup system to automatically collect nearby items
	itemPickupSystem := engine.NewItemPickupSystem(game.World)

	// GAP-002 REPAIR: Add spell casting systems
	// Initialize status effect system first (required by spell casting system)
	statusEffectRNG := rand.New(rand.NewSource(*seed + 999)) // Use seed offset for status effects
	statusEffectSystem := engine.NewStatusEffectSystem(game.World, statusEffectRNG)
	spellCastingSystem := engine.NewSpellCastingSystem(game.World, statusEffectSystem)
	playerSpellCastingSystem := engine.NewPlayerSpellCastingSystem(spellCastingSystem, game.World)
	manaRegenSystem := &engine.ManaRegenSystem{} // GAP #2 REPAIR: Add player combat system to connect Space key to combat
	playerCombatSystem := engine.NewPlayerCombatSystem(combatSystem, game.World)

	// GAP #3 REPAIR: Add player item use system to connect E key to inventory
	playerItemUseSystem := engine.NewPlayerItemUseSystem(inventorySystem, game.World)

	// Add tutorial and help systems (Phase 8.6)
	tutorialSystem := engine.NewTutorialSystem()
	// H-004 FIX: Disable tutorial if --no-tutorial flag is set
	if *noTutorial {
		tutorialSystem.Enabled = false
		tutorialSystem.ShowUI = false
	}
	helpSystem := engine.NewHelpSystem()

	// Connect help system to input system for ESC key handling
	inputSystem.SetHelpSystem(helpSystem)
	// Connect tutorial system to input system for ESC key skip handling
	inputSystem.SetTutorialSystem(tutorialSystem)
	// Phase 10.1: Connect camera system to input system for mouse aim (screen-to-world conversion)
	inputSystem.SetCameraSystem(game.CameraSystem)

	// Add systems in correct order:
	// 1. Input - captures player actions
	// 2. Rotation - updates entity facing direction based on aim (Phase 10.1)
	// 3. Player Combat/Item Use/Spell Casting - processes input flags
	// 4. Movement - applies velocity to position
	// 5. Collision - checks and resolves collisions
	// 6. Combat - handles damage/status effects
	// 7. Status Effects - processes DoT, buffs, debuffs, shields
	// 8. AI - enemy decision-making
	// 9. Progression - XP and leveling
	// 10. Skill Progression - applies skill effects to stats
	// 11. Audio Manager - updates music based on game context
	// 12. Objective Tracker - updates quest progress
	// 13. Item Pickup - collects nearby items
	// 14. Spell Casting - executes spell effects
	// 15. Mana Regen - regenerates mana
	// 16. Inventory - item management
	// 17. Animation - updates sprite frames (before rendering)
	// 18. Tutorial/Help - UI overlays
	game.World.AddSystem(inputSystem)

	// Phase 10.3: Add camera system for screen shake and visual feedback
	// Must be in update loop to process ScreenShakeComponent, HitStopComponent, and accessibility settings
	// Processes after input to apply shake effects before rendering
	game.World.AddSystem(game.CameraSystem)

	// Phase 10.1: Add rotation system for 360° rotation and mouse aim
	// Processes after input to update facing direction based on aim component
	rotationSystem := engine.NewRotationSystem(game.World)
	game.World.AddSystem(&rotationSystemWrapper{system: rotationSystem})

	game.World.AddSystem(playerCombatSystem)
	game.World.AddSystem(playerItemUseSystem)
	game.World.AddSystem(playerSpellCastingSystem)
	game.World.AddSystem(movementSystem)
	game.World.AddSystem(collisionSystem)

	// Phase 10.2: Add projectile system for ranged weapon physics
	// Processes after collision to use terrain checker for wall bounces
	projectileSystem := engine.NewProjectileSystem(game.World)
	// Note: terrainChecker will be set after terrain generation
	game.World.AddSystem(projectileSystem)

	game.World.AddSystem(combatSystem)
	game.World.AddSystem(statusEffectSystem) // Process status effects after combat

	// Add revival system for multiplayer death mechanics (Category 1.1)
	// Allows living players to revive dead teammates through proximity interaction
	revivalSystem := engine.NewRevivalSystem(game.World)
	game.World.AddSystem(revivalSystem)

	game.World.AddSystem(aiSystem)

	// Phase 13.1: Add behavior tree system for advanced AI
	// Executes behavior trees for entities with behavior tree components
	behaviorTreeSystem := engine.NewBehaviorTreeSystem(game.World)
	game.World.AddSystem(behaviorTreeSystem)

	// Phase 13.2: Add squad system for coordinated enemy tactics
	// Manages squad formations, coordination, and tactical behaviors
	squadSystem := engine.NewSquadSystem(game.World)
	game.World.AddSystem(&squadSystemWrapper{system: squadSystem})

	game.World.AddSystem(progressionSystem)

	// Phase 13.3: Add faction system for reputation tracking and relationships
	// Manages NPC faction allegiances and player reputation with different groups
	factionSystem := engine.NewFactionSystem(game.World, logger)
	game.World.AddSystem(factionSystem)

	// Phase 28: Add reputation and alignment systems for moral choices
	// Manages player reputation with factions and moral alignment tracking
	reputationSystem := engine.NewReputationSystem(game.World)
	game.World.AddSystem(&reputationSystemWrapper{system: reputationSystem})

	alignmentSystem := engine.NewAlignmentSystem(game.World)
	game.World.AddSystem(&alignmentSystemWrapper{system: alignmentSystem})

	factionReactionSystem := engine.NewFactionReactionSystem(game.World)
	game.World.AddSystem(&factionReactionSystemWrapper{system: factionReactionSystem})

	// Add skill progression system
	skillProgressionSystem := engine.NewSkillProgressionSystem()
	game.World.AddSystem(skillProgressionSystem)

	// GAP-012 REPAIR: Add visual feedback system for hit flashes and tints
	visualFeedbackSystem := engine.NewVisualFeedbackSystem()
	game.World.AddSystem(visualFeedbackSystem)

	// Add audio manager system
	game.World.AddSystem(audioManagerSystem)

	// Add objective tracker system
	game.World.AddSystem(objectiveTracker)

	game.World.AddSystem(itemPickupSystem)
	game.World.AddSystem(spellCastingSystem)
	game.World.AddSystem(manaRegenSystem)
	game.World.AddSystem(inventorySystem)

	// Add commerce, dialog, and crafting systems (Category 1.3 - Commerce & NPC Integration)
	game.World.AddSystem(commerceSystem)
	game.World.AddSystem(dialogSystem)
	game.World.AddSystem(craftingSystem)

	// Phase 11.2: Add interaction system for puzzle element interactions
	game.World.AddSystem(interactionSystem)

	// GAP-017 REPAIR: Add animation system before tutorial/help to update sprites first
	game.World.AddSystem(&animationSystemWrapper{
		system: animationSystem,
		logger: game.World.GetLogger(),
	})

	// Category 5.2: Add equipment visual system after animation to update equipment layers
	game.World.AddSystem(equipmentVisualSystem)

	game.World.AddSystem(tutorialSystem)
	game.World.AddSystem(helpSystem)

	// GAP-016 REPAIR: Add particle system for rendering effects
	game.World.AddSystem(particleSystem)

	// Phase 5.4: Add weather system for atmospheric effects
	weatherSystem := engine.NewWeatherSystem(game.World)
	game.World.AddSystem(weatherSystem)

	// Phase 5.3: Add lifetime system for temporary entities (spell lights, etc.)
	lifetimeSystem := engine.NewLifetimeSystemWithLogger(game.World, clientLogger.Logger)
	game.World.AddSystem(lifetimeSystem)

	// Phase 11.2: Add puzzle system for procedural puzzle management
	puzzleSystem := engine.NewPuzzleSystem(game.World)
	game.World.AddSystem(puzzleSystem)

	// Phase 11.3: Environmental Destruction & Manipulation Systems
	// Initialize fire propagation system for explosive barrel ignition
	const tileSize = 32 // Standard tile size used throughout the engine
	firePropagationSystem := engine.NewFirePropagationSystemWithLogger(tileSize, *seed+1090, clientLogger.Logger)
	firePropagationSystem.SetWorld(game.World)
	game.World.AddSystem(firePropagationSystem)

	// Initialize destructible object system for crates, barrels, furniture
	destructibleObjectSystem := engine.NewDestructibleObjectSystemWithLogger(tileSize, *seed+1100, clientLogger.Logger)
	destructibleObjectSystem.SetWorld(game.World)
	destructibleObjectSystem.SetFireSystem(firePropagationSystem)
	game.World.AddSystem(destructibleObjectSystem)

	// Initialize carry system for pickup and throw mechanics
	carrySystem := engine.NewCarrySystemWithLogger(clientLogger.Logger)
	carrySystem.SetWorld(game.World)
	game.World.AddSystem(carrySystem)

	// Connect carry system to interaction system for F key pickup
	interactionSystem.SetCarrySystem(carrySystem)

	// Initialize hazard system for poison clouds, oil puddles, smoke
	hazardSystem := engine.NewHazardSystemWithLogger(clientLogger.Logger)
	hazardSystem.SetWorld(game.World)
	game.World.AddSystem(hazardSystem)

	// Phase 12.2: Add narrative system for story progression
	// Tracks narrative events and manages story arc advancement
	narrativeSystem := engine.NewNarrativeSystem(game.World)
	game.World.AddSystem(narrativeSystem)

	// Phase 14: Add shadow system for enhanced lighting effects
	// Processes shadow-casting entities and renders shadows
	shadowSystem := engine.NewShadowSystemWithLogger(game.World, clientLogger.Logger)
	game.World.AddSystem(shadowSystem)

	// Store references to tutorial and help systems in game for rendering
	game.TutorialSystem = tutorialSystem
	game.HelpSystem = helpSystem

	// GAP-012 REPAIR: Set camera reference on combat system for screen shake
	combatSystem.SetCamera(game.CameraSystem)

	// GAP-016 REPAIR: Set particle system reference on combat system for hit effects
	combatSystem.SetParticleSystem(particleSystem, game.World, *genreID)

	// Phase 10.2: Set projectile system reference on combat system for ranged weapon spawning
	combatSystem.SetProjectileSystem(projectileSystem)

	// Phase 10.3: Set camera reference on projectile system for impact shake
	projectileSystem.SetCamera(game.CameraSystem)

	// Phase 10.2: Set genre and seed for projectile visual generation
	projectileSystem.SetGenre(*genreID)
	projectileSystem.SetSeed(*seed)

	if *verbose {
		clientLogger.Info("systems initialized")
	} // Gap #3: Initialize performance monitoring (wraps World.Update)
	perfMonitor := engine.NewPerformanceMonitor(game.World)
	if *verbose {
		clientLogger.Info("performance monitoring initialized")
		// Start periodic performance logging in background
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				metrics := perfMonitor.GetMetrics()
				clientLogger.WithField("metrics", metrics.String()).Info("performance metrics")
			}
		}()
	}
	_ = perfMonitor // Suppress unused warning when not verbose

	// Generate initial world terrain
	clientLogger.Info("generating procedural terrain")

	terrainGen := terrain.NewBSPGeneratorWithLogger(logger) // Use BSP algorithm with logging
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    *genreID,
		Custom: map[string]interface{}{
			"width":  80,
			"height": 50,
		},
	}

	terrainResult, err := terrainGen.Generate(*seed, params)
	if err != nil {
		clientLogger.WithError(err).Fatal("failed to generate terrain")
	}

	generatedTerrain := terrainResult.(*terrain.Terrain)
	clientLogger.WithFields(logrus.Fields{
		"width":     generatedTerrain.Width,
		"height":    generatedTerrain.Height,
		"roomCount": len(generatedTerrain.Rooms),
	}).Info("terrain generated")

	// Initialize terrain rendering system
	if *verbose {
		clientLogger.Info("initializing terrain rendering system")
	}

	terrainRenderSystem := engine.NewTerrainRenderSystem(32, 32, *genreID, *seed)
	terrainRenderSystem.SetTerrain(generatedTerrain)
	game.TerrainRenderSystem = terrainRenderSystem

	if *verbose {
		clientLogger.Info("terrain rendering system initialized")
	}

	// Configure lighting system
	if *enableLighting {
		clientLogger.Info("enabling dynamic lighting system")
		game.EnableLighting(true)
		game.SetLightingGenrePreset(*genreID)
		clientLogger.WithFields(logrus.Fields{
			"genre":     *genreID,
			"enabled":   true,
			"maxLights": 16,
		}).Info("lighting system configured")
	}

	// GAP REPAIR: Initialize efficient terrain collision checking
	if *verbose {
		clientLogger.Info("initializing terrain collision system")
	}

	terrainChecker := engine.NewTerrainCollisionChecker(32, 32)
	terrainChecker.SetTerrain(generatedTerrain)

	// Connect terrain checker to collision system and projectile system
	for _, system := range game.World.GetSystems() {
		if collisionSys, ok := system.(*engine.CollisionSystem); ok {
			collisionSys.SetTerrainChecker(terrainChecker)
		}
		if projSys, ok := system.(*engine.ProjectileSystem); ok {
			projSys.SetTerrainChecker(terrainChecker)
		}
	}

	if *verbose {
		clientLogger.Info("terrain collision system initialized (efficient mode)")
	}

	// Phase 13.3: Generate world factions
	clientLogger.Info("generating world factions")
	factionGen := faction.NewGenerator()
	factionResult, err := factionGen.Generate(*seed+1000, params) // Use offset seed for faction variety
	if err != nil {
		clientLogger.WithError(err).Fatal("failed to generate factions")
	}

	worldFactions := factionResult.([]*engine.Faction)
	clientLogger.WithFields(logrus.Fields{
		"count": len(worldFactions),
		"genre": *genreID,
	}).Info("factions generated")

	// Add factions to faction system
	for _, fac := range worldFactions {
		// Get faction system from world
		for _, system := range game.World.GetSystems() {
			if facSys, ok := system.(*engine.FactionSystem); ok {
				facSys.AddFaction(fac)
				if *verbose {
					clientLogger.WithFields(logrus.Fields{
						"factionID":   fac.ID,
						"factionName": fac.Name,
						"factionType": fac.Type,
					}).Debug("faction added to world")
				}
			}
		}
	}

	// CATEGORY 4.3: Initialize spatial partition system for viewport culling
	// Provides significant performance benefits with large entity counts through spatial queries
	// Always enabled as a core optimization (previously optional, now standard)
	if *verbose {
		clientLogger.Info("initializing spatial partition system for viewport culling")
	}

	// Calculate world bounds from terrain dimensions (32 pixels per tile)
	worldWidth := float64(generatedTerrain.Width) * 32.0
	worldHeight := float64(generatedTerrain.Height) * 32.0

	// Create spatial partition system with quadtree-based structure
	spatialSystem := engine.NewSpatialPartitionSystem(worldWidth, worldHeight)

	// Register with ECS World for automatic updates every 60 frames
	game.World.AddSystem(spatialSystem)

	// BUGFIX: Force initial quadtree rebuild with all entities before enabling culling
	// The spatial partition uses lazy rebuild and would be empty on first render frame
	// This caused all entities (including player) to be culled until first rebuild
	spatialSystem.Rebuild(game.World.GetEntities())

	// Connect to render system for viewport culling
	game.RenderSystem.SetSpatialPartition(spatialSystem)
	// Now safe to enable culling since quadtree is populated
	game.RenderSystem.EnableCulling(true)

	// WASM OPTIMIZATION: Enable batch rendering to reduce GPU state changes
	// Groups entities with same sprite image before drawing (1,667x speedup potential)
	// Particularly beneficial for WASM where GPU state changes are expensive
	game.RenderSystem.EnableBatching(true)

	clientLogger.WithFields(logrus.Fields{
		"worldWidth":     worldWidth,
		"worldHeight":    worldHeight,
		"cellSize":       8, // Quadtree capacity per node (8 entities before subdivision)
		"initialCount":   len(game.World.GetEntities()),
		"cullingActive":  true,
		"batchingActive": true,
	}).Info("spatial partition system initialized with initial rebuild, culling and batching enabled")

	if *verbose {
		clientLogger.WithFields(logrus.Fields{
			"worldWidth":  worldWidth,
			"worldHeight": worldHeight,
		}).Info("spatial partition enabled")
	}

	// GAP-001 REPAIR: Connect terrain to MapUI for map functionality
	if *verbose {
		clientLogger.Info("connecting terrain to Map UI")
	}
	game.MapUI.SetTerrain(generatedTerrain)
	if *verbose {
		clientLogger.Info("Map UI configured with terrain data")
	}

	// GAP #1 REPAIR: Spawn enemies in terrain rooms
	if *verbose {
		clientLogger.Info("spawning enemies in dungeon rooms")
	}

	enemyParams := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    *genreID,
	}

	enemyCount, err := engine.SpawnEnemiesInTerrain(game.World, generatedTerrain, *seed, enemyParams)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn enemies")
	} else if *verbose {
		clientLogger.WithFields(logrus.Fields{
			"enemyCount": enemyCount,
			"roomCount":  len(generatedTerrain.Rooms) - 1,
		}).Info("spawned enemies")
	}

	// GAP #4 REPAIR: Spawn merchants in dungeon
	if *verbose {
		clientLogger.Info("spawning merchants in dungeon")
	}

	merchantParams := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    *genreID,
	}

	merchantCount, err := engine.SpawnMerchantsInTerrain(game.World, generatedTerrain, *seed, merchantParams, 2) // Spawn 2 merchants per level
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn merchants")
	} else if *verbose {
		clientLogger.WithField("merchantCount", merchantCount).Info("spawned merchants")
	}

	// Spawn crafting stations in dungeon
	if *verbose {
		clientLogger.Info("spawning crafting stations in dungeon")
	}

	stationGen := station.NewStationGenerator()
	stationCount := engine.SpawnStationsInTerrain(game.World, stationGen, generatedTerrain, 32, *seed+1000, *genreID)
	if *verbose {
		clientLogger.WithField("stationCount", stationCount).Info("spawned crafting stations")
	}

	// Phase 11.2: Spawn procedural puzzles in dungeon
	if *verbose {
		clientLogger.Info("spawning procedural puzzles in dungeon")
	}

	puzzleParams := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    *genreID,
	}

	puzzleCount, err := engine.SpawnPuzzlesInTerrain(game.World, generatedTerrain, *seed+2000, puzzleParams, 5)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn puzzles")
	} else {
		clientLogger.WithFields(logrus.Fields{
			"puzzleCount": puzzleCount,
			"targetCount": 5,
			"roomCount":   len(generatedTerrain.Rooms) - 1,
		}).Info("spawned procedural puzzles")
	}

	// Phase 11.3: Spawn destructible objects in dungeon (crates, barrels, furniture)
	if *verbose {
		clientLogger.Info("spawning destructible objects in dungeon")
	}
	objectCount := spawnDestructibleObjects(game.World, generatedTerrain, *seed+3000, *genreID, clientLogger.Logger)
	clientLogger.WithFields(logrus.Fields{
		"objectCount": objectCount,
		"roomCount":   len(generatedTerrain.Rooms) - 1,
		"genre":       *genreID,
	}).Info("spawned destructible objects")

	// Phase 5.3: Spawn environmental lights in dungeon (if lighting enabled)
	if *enableLighting {
		if *verbose {
			clientLogger.Info("spawning environmental lights in dungeon")
		}
		lightCount := spawnEnvironmentalLights(game.World, generatedTerrain, *seed+2000, *genreID)
		clientLogger.WithFields(logrus.Fields{
			"lightCount": lightCount,
			"genre":      *genreID,
		}).Info("spawned environmental lights")
	}

	// Phase 5.4: Spawn weather effects (if enabled)
	if *enableWeather {
		if *verbose {
			clientLogger.Info("spawning weather effects")
		}
		weatherEntity := spawnWeather(game.World, *width, *height, *seed+3000, *genreID, *weatherType, *weatherIntensity)
		if weatherEntity != nil {
			clientLogger.WithFields(logrus.Fields{
				"type":      *weatherType,
				"intensity": *weatherIntensity,
				"genre":     *genreID,
			}).Info("weather effects spawned")
		}
	}

	// Create player entity
	if *verbose {
		clientLogger.Info("creating player entity")
	}

	player := game.World.CreateEntity()

	// Store player entity reference for death callback
	playerEntity = player

	// GAP #3 REPAIR: Calculate player spawn position from first room
	var playerX, playerY float64
	if len(generatedTerrain.Rooms) > 0 {
		// Spawn in center of first room
		firstRoom := generatedTerrain.Rooms[0]
		cx, cy := firstRoom.Center()
		playerX = float64(cx * 32) // Convert tile coordinates to world coordinates
		playerY = float64(cy * 32)
		if *verbose {
			clientLogger.WithFields(logrus.Fields{
				"tileX":  cx,
				"tileY":  cy,
				"worldX": playerX,
				"worldY": playerY,
			}).Info("player spawning in first room")
		}
	} else {
		// Fallback to default position if no rooms (shouldn't happen with valid terrain)
		playerX, playerY = 400, 300
		clientLogger.Warn("no rooms in terrain, using default spawn position")
	}

	// Add player components
	player.AddComponent(&engine.PositionComponent{X: playerX, Y: playerY})
	player.AddComponent(&engine.VelocityComponent{VX: 0, VY: 0})
	player.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
	player.AddComponent(&engine.TeamComponent{TeamID: 1}) // Player team

	// Add input component for player control
	player.AddComponent(&engine.EbitenInput{})

	// Phase 10.1: Add rotation and aim components for 360° rotation and mouse aim
	// RotationComponent stores facing direction with smooth interpolation (3.0 rad/s = ~172°/s)
	player.AddComponent(engine.NewRotationComponent(0, 3.0)) // Start facing right (0 radians)
	// AimComponent manages independent aim direction (mouse/touch input)
	player.AddComponent(engine.NewAimComponent(0)) // Start aiming right

	// GAP-017 REPAIR: Add animated sprite instead of static sprite
	// WASM FIX: Set Image to nil initially; animation system will create it on first update
	// In WASM, ebiten.NewImage() can only be called after graphics context is ready
	playerSprite := &engine.EbitenSprite{
		Image:   nil, // Will be created by animation system on first update
		Width:   28,
		Height:  28,
		Visible: true,
		Layer:   10, // Draw player on top
	}
	player.AddComponent(playerSprite)

	// Add animation component for multi-frame character animation
	// GAP-019 REPAIR: Use special seed offset for player to ensure distinct color
	playerAnim := engine.NewAnimationComponent(*seed + int64(player.ID*1000))
	playerAnim.CurrentState = engine.AnimationStateIdle
	playerAnim.FrameTime = 0.15 // ~6.7 FPS for smooth animation
	playerAnim.Loop = true
	playerAnim.Playing = true
	playerAnim.FrameCount = 4 // 4 frames per animation
	playerAnim.Dirty = true   // CRITICAL: Mark dirty to trigger initial frame generation
	player.AddComponent(playerAnim)

	// Category 5.2: Add equipment visual component for showing equipped items on sprite
	equipmentVisualComp := engine.NewEquipmentVisualComponent()
	player.AddComponent(equipmentVisualComp)

	// Add camera that follows the player
	camera := engine.NewCameraComponent()
	camera.Smoothing = 0.1
	player.AddComponent(camera)

	// Phase 10.3: Add advanced screen shake component
	screenShake := engine.NewScreenShakeComponent()
	player.AddComponent(screenShake)

	// Phase 10.3: Add hit-stop component
	hitStop := engine.NewHitStopComponent()
	player.AddComponent(hitStop)

	// Phase 11.1: Add layer component for multi-layer collision
	layerComp := engine.NewLayerComponent()
	layerComp.CurrentLayer = 0 // Ground layer
	player.AddComponent(&layerComp)

	// Phase 14: Add shadow component for enhanced lighting
	playerShadow := engine.NewShadowComponent(28) // Player sprite size
	playerShadow.CastsShadow = true
	playerShadow.ShadowType = engine.ShadowTypeSoft // Player gets soft shadow
	player.AddComponent(playerShadow)

	// Phase 5.3: Add player torch for dynamic lighting (if enabled)
	if *enableLighting {
		playerTorch := engine.NewTorchLight(200) // 200-pixel radius torch with flicker
		playerTorch.Enabled = true
		player.AddComponent(playerTorch)

		if *verbose {
			clientLogger.WithFields(logrus.Fields{
				"radius":    200,
				"intensity": playerTorch.Intensity,
			}).Info("player torch added")
		}
	}

	// Set player as the active camera
	game.CameraSystem.SetActiveCamera(player)

	// Phase 14.2: Configure animation system with player and camera for optimizations
	animationSystem.SetCameraSystem(game.CameraSystem)
	animationSystem.SetPlayerEntity(player)
	// Viewport culling and distance LOD are enabled by default
	// Can be disabled for debugging with:
	// animationSystem.EnableViewportCulling(false)
	// animationSystem.EnableDistanceLOD(false)

	if *verbose {
		clientLogger.WithFields(logrus.Fields{
			"viewportCulling": true,
			"distanceLOD":     true,
			"closeThreshold":  200.0,
			"midThreshold":    400.0,
		}).Info("animation system configured with performance optimizations")
	}

	// Set player for HUD display
	game.HUDSystem.SetPlayerEntity(player)

	// Set player for UI systems (inventory, quests, shop)
	game.SetPlayerEntity(player)

	// GAP-004 REPAIR: Initialize and wire up commerce UI
	shopUI := engine.NewShopUI(*width, *height)
	shopUI.SetPlayerEntity(player)
	shopUI.SetCommerceSystem(commerceSystem)
	shopUI.SetDialogSystem(dialogSystem)
	game.ShopUI = shopUI

	if *verbose {
		clientLogger.Info("shop UI initialized and connected to commerce/dialog systems")
	}

	// Initialize and wire up crafting UI
	craftingUI := engine.NewCraftingUI(*width, *height)
	craftingUI.SetPlayerEntity(player)
	craftingUI.SetCraftingSystem(craftingSystem)
	game.CraftingUI = craftingUI

	if *verbose {
		clientLogger.Info("crafting UI initialized and connected to crafting system")
	}

	// Add player stats
	playerStats := engine.NewStatsComponent()
	playerStats.Attack = 10
	playerStats.Defense = 5
	// GAP-003 REPAIR: Initialize derived stats with baseline values
	playerStats.CritChance = 0.05 // 5% crit chance
	playerStats.CritDamage = 1.5  // 1.5x crit damage multiplier
	playerStats.Evasion = 0.05    // 5% evasion chance
	// Resistances default to 0.0 (handled by NewStatsComponent)
	player.AddComponent(playerStats)

	// Add player experience/progression
	playerExp := engine.NewExperienceComponent()
	player.AddComponent(playerExp)

	// Add player inventory
	playerInventory := engine.NewInventoryComponent(20, 100.0) // 20 items, 100 weight max
	playerInventory.Gold = 100
	player.AddComponent(playerInventory)

	// Add player equipment
	playerEquipment := engine.NewEquipmentComponent()
	player.AddComponent(playerEquipment)

	// GAP-002 REPAIR: Add mana and spells
	playerMana := &engine.ManaComponent{
		Current: 100,
		Max:     100,
		Regen:   5.0, // 5 mana per second
	}
	player.AddComponent(playerMana)

	// Load procedurally generated spells
	err = engine.LoadPlayerSpells(player, *seed, *genreID, 1)
	if err != nil {
		clientLogger.WithError(err).Fatal("failed to load player spells")
	}
	if *verbose {
		clientLogger.Info("player spells loaded (keys 1-5)")
	}

	// Load procedurally generated skill tree
	err = engine.LoadPlayerSkillTree(player, *seed, *genreID, 0)
	if err != nil {
		clientLogger.WithError(err).Fatal("failed to load skill tree")
	}
	if *verbose {
		comp, _ := player.GetComponent("skill_tree")
		if comp != nil {
			treeComp := comp.(*engine.SkillTreeComponent)
			clientLogger.WithFields(logrus.Fields{
				"treeName":   treeComp.Tree.Name,
				"skillCount": len(treeComp.Tree.Nodes),
			}).Info("skill tree loaded (press K)")
		}
	}

	// Add quest tracker
	questTracker := engine.NewQuestTrackerComponent(5) // Max 5 active quests
	player.AddComponent(questTracker)

	// Add player attack capability
	player.AddComponent(&engine.AttackComponent{
		Damage:     15,
		DamageType: combat.DamagePhysical,
		Range:      50,
		Cooldown:   0.5,
	})

	// Add collision for player (28x28 to fit through 32px corridors)
	player.AddComponent(&engine.ColliderComponent{
		Width:     28,
		Height:    28,
		Solid:     true,
		IsTrigger: false,
		Layer:     1,
		OffsetX:   -14, // Center the collider (28/2 = 14)
		OffsetY:   -14,
	})

	// GAP-012 REPAIR: Add visual feedback for hit flash
	player.AddComponent(engine.NewVisualFeedbackComponent())

	clientLogger.WithField("entityID", player.ID).Info("player entity created")

	// Apply character class stats if character data is available
	if charData := game.GetPendingCharacterData(); charData != nil {
		clientLogger.WithFields(logrus.Fields{
			"name":  charData.Name,
			"class": charData.Class.String(),
		}).Info("applying character class stats")

		if err := engine.ApplyClassStats(player, charData.Class); err != nil {
			clientLogger.WithError(err).Fatal("failed to apply character class stats")
		}

		// Note: Character name stored in character creation data.
		// Future enhancement: Add NameComponent for displaying names in multiplayer.
	}

	// Add starter items to inventory
	clientLogger.Info("adding starter items to inventory")
	addStarterItems(playerInventory, *seed, *genreID, logger)

	// Add tutorial quest
	clientLogger.Info("creating tutorial quest")
	addTutorialQuest(questTracker, *seed, *genreID, logger)

	// Initialize save/load system (Phase 8.4)
	clientLogger.Info("initializing save/load system")

	saveManager, err := saveload.NewSaveManager("./saves")
	if err != nil {
		clientLogger.WithError(err).Warn("failed to initialize save manager, save/load functionality will be unavailable")
	} else {
		if *verbose {
			clientLogger.Info("save/load system initialized")
		}

		// Setup quick save callback (F5)
		inputSystem.SetQuickSaveCallback(func() error {
			clientLogger.Info("quick save (F5 pressed)")

			// Get player position
			var posX, posY float64
			if posComp, ok := player.GetComponent("position"); ok {
				pos := posComp.(*engine.PositionComponent)
				posX, posY = pos.X, pos.Y
			}

			// Get player health
			var currentHealth, maxHealth float64
			if healthComp, ok := player.GetComponent("health"); ok {
				health := healthComp.(*engine.HealthComponent)
				currentHealth, maxHealth = health.Current, health.Max
			}

			// Get player stats
			var attack, defense, magic float64
			if statsComp, ok := player.GetComponent("stats"); ok {
				stats := statsComp.(*engine.StatsComponent)
				attack, defense, magic = stats.Attack, stats.Defense, stats.MagicPower
			} // Get player level and XP
			var level int
			var currentXP int64
			if expComp, ok := player.GetComponent("experience"); ok {
				exp := expComp.(*engine.ExperienceComponent)
				level, currentXP = exp.Level, int64(exp.CurrentXP)
			}

			// Get inventory data
			var gold int
			itemDataList := make([]saveload.ItemData, 0)
			if invComp, ok := player.GetComponent("inventory"); ok {
				inv := invComp.(*engine.InventoryComponent)
				gold = inv.Gold
				// Serialize full item data
				for _, itm := range inv.Items {
					itemDataList = append(itemDataList, saveload.ItemToData(itm))
				}
			}

			// GAP-008: Serialize equipped items
			var equippedItems saveload.EquipmentData
			if equip, hasEquip := player.GetComponent("equipment"); hasEquip {
				equipment := equip.(*engine.EquipmentComponent)
				// Check main hand for weapon
				if weapon := equipment.Slots[engine.SlotMainHand]; weapon != nil {
					weaponData := saveload.ItemToData(weapon)
					equippedItems.Weapon = &weaponData
				}
				// Check chest for armor (primary armor slot)
				if armor := equipment.Slots[engine.SlotChest]; armor != nil {
					armorData := saveload.ItemToData(armor)
					equippedItems.Armor = &armorData
				}
				// Check accessory slots
				if accessory := equipment.Slots[engine.SlotAccessory1]; accessory != nil {
					accessoryData := saveload.ItemToData(accessory)
					equippedItems.Accessory = &accessoryData
				}
			}

			// Serialize mana
			var currentMana, maxMana int
			if manaComp, hasMana := player.GetComponent("mana"); hasMana {
				mana := manaComp.(*engine.ManaComponent)
				currentMana = mana.Current
				maxMana = mana.Max
			}

			// Serialize spells
			spellDataList := make([]saveload.SpellData, 0)
			if slotsComp, hasSlots := player.GetComponent("spell_slots"); hasSlots {
				slots := slotsComp.(*engine.SpellSlotComponent)
				for i := 0; i < 5; i++ {
					if spell := slots.GetSlot(i); spell != nil {
						spellDataList = append(spellDataList, saveload.SpellToData(spell))
					}
				}
			}

			// GAP-005 REPAIR: Serialize fog of war exploration state
			var fogOfWar [][]bool
			if game.MapUI != nil {
				fogOfWar = game.MapUI.GetFogOfWar()
				if *verbose {
					height := 0
					if len(fogOfWar) > 0 {
						height = len(fogOfWar[0])
					}
					clientLogger.WithFields(logrus.Fields{
						"width":  len(fogOfWar),
						"height": height,
					}).Debug("serializing fog of war")
				}
			}

			// GAP-003 REPAIR: Export tutorial state
			var tutorialStateData *saveload.TutorialStateData
			if game.TutorialSystem != nil {
				enabled, showUI, currentStep, completed := game.TutorialSystem.ExportState()
				tutorialStateData = &saveload.TutorialStateData{
					Enabled:        enabled,
					ShowUI:         showUI,
					CurrentStepIdx: currentStep,
					CompletedSteps: completed,
				}
			}

			// Create game save
			gameSave := &saveload.GameSave{
				Version: saveload.SaveVersion,
				PlayerState: &saveload.PlayerState{
					EntityID:      player.ID,
					X:             posX,
					Y:             posY,
					CurrentHealth: currentHealth,
					MaxHealth:     maxHealth,
					Level:         level,
					Experience:    int(currentXP),
					Attack:        attack,
					Defense:       defense,
					MagicPower:    magic,
					Speed:         1.0,
					Items:         itemDataList,
					Gold:          gold,
					EquippedItems: equippedItems,
					CurrentMana:   currentMana,
					MaxMana:       maxMana,
					Spells:        spellDataList,
					TutorialState: tutorialStateData,
				},
				WorldState: &saveload.WorldState{
					Seed:       *seed,
					GenreID:    *genreID,
					Width:      generatedTerrain.Width,
					Height:     generatedTerrain.Height,
					Difficulty: 0.5,
					Depth:      1,
					FogOfWar:   fogOfWar, // GAP-005: Fog of war persistence
				},
				Settings: &saveload.GameSettings{
					ScreenWidth:  *width,
					ScreenHeight: *height,
					Fullscreen:   false,
					VSync:        true,
					MasterVolume: 1.0,
					MusicVolume:  0.7,
					SFXVolume:    0.8,
					KeyBindings:  make(map[string]string),
				},
			}

			if err := saveManager.SaveGame("quicksave", gameSave); err != nil {
				clientLogger.WithError(err).Error("failed to save game")
				return err
			}

			clientLogger.Info("game saved successfully")
			return nil
		})

		// Setup quick load callback (F9)
		inputSystem.SetQuickLoadCallback(func() error {
			clientLogger.Info("quick load (F9 pressed)")

			gameSave, err := saveManager.LoadGame("quicksave")
			if err != nil {
				clientLogger.WithError(err).Error("failed to load game")
				return err
			}

			// Restore player position
			if posComp, ok := player.GetComponent("position"); ok {
				pos := posComp.(*engine.PositionComponent)
				pos.X = gameSave.PlayerState.X
				pos.Y = gameSave.PlayerState.Y
			}

			// Restore player health
			if healthComp, ok := player.GetComponent("health"); ok {
				health := healthComp.(*engine.HealthComponent)
				health.Current = gameSave.PlayerState.CurrentHealth
				health.Max = gameSave.PlayerState.MaxHealth
			}

			// Restore player stats
			if statsComp, ok := player.GetComponent("stats"); ok {
				stats := statsComp.(*engine.StatsComponent)
				stats.Attack = gameSave.PlayerState.Attack
				stats.Defense = gameSave.PlayerState.Defense
				stats.MagicPower = gameSave.PlayerState.MagicPower
			}

			// Restore player level and XP
			if expComp, ok := player.GetComponent("experience"); ok {
				exp := expComp.(*engine.ExperienceComponent)
				exp.Level = gameSave.PlayerState.Level
				exp.CurrentXP = gameSave.PlayerState.Experience
				// Note: RequiredXP is recalculated by progression system
			}

			// Restore inventory (simplified)
			if invComp, ok := player.GetComponent("inventory"); ok {
				inv := invComp.(*engine.InventoryComponent)

				// GAP-007: Restore full inventory items
				inv.Items = make([]*item.Item, 0, len(gameSave.PlayerState.Items))
				for _, itemData := range gameSave.PlayerState.Items {
					restoredItem := saveload.DataToItem(itemData)
					inv.Items = append(inv.Items, restoredItem)
				}

				// GAP-009: Restore gold
				inv.Gold = gameSave.PlayerState.Gold

				if *verbose {
					clientLogger.WithFields(logrus.Fields{
						"itemCount": len(inv.Items),
						"gold":      inv.Gold,
					}).Debug("restored inventory")
				}
			}

			// GAP-008: Restore equipped items
			if equipComp, ok := player.GetComponent("equipment"); ok {
				equipment := equipComp.(*engine.EquipmentComponent)

				// Clear existing equipment
				equipment.Slots = make(map[engine.EquipmentSlot]*item.Item)

				// Restore weapon
				if gameSave.PlayerState.EquippedItems.Weapon != nil {
					weapon := saveload.DataToItem(*gameSave.PlayerState.EquippedItems.Weapon)
					equipment.Slots[engine.SlotMainHand] = weapon
				}

				// Restore armor
				if gameSave.PlayerState.EquippedItems.Armor != nil {
					armor := saveload.DataToItem(*gameSave.PlayerState.EquippedItems.Armor)
					equipment.Slots[engine.SlotChest] = armor
				}

				// Restore accessory
				if gameSave.PlayerState.EquippedItems.Accessory != nil {
					accessory := saveload.DataToItem(*gameSave.PlayerState.EquippedItems.Accessory)
					equipment.Slots[engine.SlotAccessory1] = accessory
				}

				equipment.StatsDirty = true // Trigger stats recalculation
			}

			// Restore mana
			if manaComp, ok := player.GetComponent("mana"); ok {
				mana := manaComp.(*engine.ManaComponent)
				mana.Current = gameSave.PlayerState.CurrentMana
				mana.Max = gameSave.PlayerState.MaxMana
			}

			// Restore spells
			if slotsComp, ok := player.GetComponent("spell_slots"); ok {
				slots := slotsComp.(*engine.SpellSlotComponent)

				// Clear existing spells
				for i := 0; i < 5; i++ {
					slots.Slots[i] = nil
				}

				// Restore saved spells
				for i, spellData := range gameSave.PlayerState.Spells {
					if i < 5 {
						restoredSpell := saveload.DataToSpell(spellData)
						slots.SetSlot(i, restoredSpell)
					}
				}
			}

			// GAP-005 REPAIR: Restore fog of war exploration state
			if game.MapUI != nil && gameSave.WorldState != nil && gameSave.WorldState.FogOfWar != nil {
				game.MapUI.SetFogOfWar(gameSave.WorldState.FogOfWar)
				if *verbose {
					fogData := gameSave.WorldState.FogOfWar
					height := 0
					if len(fogData) > 0 {
						height = len(fogData[0])
					}
					clientLogger.WithFields(logrus.Fields{
						"width":  len(fogData),
						"height": height,
					}).Debug("restored fog of war")
				}
			}

			// GAP-003 REPAIR: Restore tutorial state
			if game.TutorialSystem != nil && gameSave.PlayerState.TutorialState != nil {
				tutState := gameSave.PlayerState.TutorialState
				game.TutorialSystem.ImportState(
					tutState.Enabled,
					tutState.ShowUI,
					tutState.CurrentStepIdx,
					tutState.CompletedSteps,
				)
				if *verbose {
					clientLogger.WithFields(logrus.Fields{
						"enabled":     tutState.Enabled,
						"currentStep": tutState.CurrentStepIdx,
						"totalSteps":  len(game.TutorialSystem.Steps),
					}).Debug("restored tutorial state")
				}
			}

			clientLogger.Info("game loaded successfully")
			return nil
		})

		if *verbose {
			clientLogger.Info("quick save/load callbacks registered (F5/F9)")
		}
	}

	// Connect inventory system to UI for item actions
	game.SetInventorySystem(inventorySystem)

	// Setup UI input callbacks
	if *verbose {
		clientLogger.Info("setting up UI input callbacks")
	}
	// GAP-014 REPAIR: Pass objective tracker to enable tutorial quest tracking
	// H-008 FIX: Check for callback registration errors
	if err := game.SetupInputCallbacks(inputSystem, objectiveTracker); err != nil {
		clientLogger.WithError(err).Fatal("failed to setup input callbacks")
	}
	if *verbose {
		clientLogger.Info("UI callbacks registered (I: Inventory, J: Quests, ESC: Pause Menu)")
		clientLogger.Info("inventory actions: E to equip/use, D to drop")
	}

	// GAP-004 REPAIR: Setup merchant interaction callback (F key)
	// H-008 FIX: Check for callback registration errors
	if err := inputSystem.SetInteractCallback(func() {
		// Get player position
		if player == nil {
			return
		}
		posComp, ok := player.GetComponent("position")
		if !ok {
			return
		}
		pos := posComp.(*engine.PositionComponent)

		// Find closest merchant within interaction range (64 pixels)
		merchant, dist := engine.FindClosestMerchant(game.World, pos.X, pos.Y, 64.0)
		if merchant == nil {
			// No merchant nearby
			if *verbose {
				clientLogger.Debug("no merchant nearby to interact with")
			}
			return
		}

		// Start dialog with merchant
		success, err := dialogSystem.StartDialog(player.ID, merchant.ID)
		if err != nil {
			clientLogger.WithError(err).Warn("failed to start dialog")
			return
		}

		if !success {
			if *verbose {
				clientLogger.Debug("dialog could not be started")
			}
			return
		}

		// Open shop UI
		shopUI.Open(merchant)

		if *verbose {
			clientLogger.WithField("distance", dist).Debug("opened shop with merchant")
		}
	}); err != nil {
		clientLogger.WithError(err).Fatal("failed to setup interact callback")
	}

	if *verbose {
		clientLogger.Info("merchant interaction registered (F key when near merchant)")
	}

	// Connect save/load callbacks to menu system
	if game.MenuSystem != nil && saveManager != nil {
		if *verbose {
			clientLogger.Info("connecting save/load callbacks to menu system")
		}

		// Create save callback that reuses the quick save logic
		saveCallback := func(saveName string) error {
			if *verbose {
				clientLogger.WithField("saveName", saveName).Info("menu save")
			}

			// Get player position
			var posX, posY float64
			if posComp, ok := player.GetComponent("position"); ok {
				pos := posComp.(*engine.PositionComponent)
				posX, posY = pos.X, pos.Y
			}

			// Get player health
			var currentHealth, maxHealth float64
			if healthComp, ok := player.GetComponent("health"); ok {
				health := healthComp.(*engine.HealthComponent)
				currentHealth, maxHealth = health.Current, health.Max
			}

			// Get player stats
			var attack, defense, magic float64
			if statsComp, ok := player.GetComponent("stats"); ok {
				stats := statsComp.(*engine.StatsComponent)
				attack, defense, magic = stats.Attack, stats.Defense, stats.MagicPower
			}

			// Get player level and XP
			var level int
			var currentXP int64
			if expComp, ok := player.GetComponent("experience"); ok {
				exp := expComp.(*engine.ExperienceComponent)
				level, currentXP = exp.Level, int64(exp.CurrentXP)
			}

			// Get inventory data
			var items []saveload.ItemData
			var gold int
			if invComp, ok := player.GetComponent("inventory"); ok {
				inv := invComp.(*engine.InventoryComponent)
				gold = inv.Gold

				// Convert items to ItemData for persistence
				for _, itm := range inv.Items {
					items = append(items, saveload.ItemData{
						Name:           itm.Name,
						Type:           itm.Type.String(),
						WeaponType:     itm.WeaponType.String(),
						ArmorType:      itm.ArmorType.String(),
						ConsumableType: itm.ConsumableType.String(),
						Rarity:         itm.Rarity.String(),
						Seed:           itm.Seed,
						Tags:           itm.Tags,
						Description:    itm.Description,
						Damage:         itm.Stats.Damage,
						Defense:        itm.Stats.Defense,
						AttackSpeed:    itm.Stats.AttackSpeed,
						Value:          itm.Stats.Value,
						Weight:         itm.Stats.Weight,
						RequiredLevel:  itm.Stats.RequiredLevel,
						DurabilityMax:  itm.Stats.DurabilityMax,
						Durability:     itm.Stats.Durability,
					})
				}
			}

			// Create game save
			gameSave := &saveload.GameSave{
				Version: saveload.SaveVersion,
				PlayerState: &saveload.PlayerState{
					EntityID:      player.ID,
					X:             posX,
					Y:             posY,
					CurrentHealth: currentHealth,
					MaxHealth:     maxHealth,
					Level:         level,
					Experience:    int(currentXP),
					Attack:        attack,
					Defense:       defense,
					MagicPower:    magic,
					Speed:         1.0,
					Items:         items, // Use new Items field instead of InventoryItems
					Gold:          gold,
				},
				WorldState: &saveload.WorldState{
					Seed:       *seed,
					GenreID:    *genreID,
					Width:      generatedTerrain.Width,
					Height:     generatedTerrain.Height,
					Difficulty: 0.5,
					Depth:      1,
				},
				Settings: &saveload.GameSettings{
					ScreenWidth:  *width,
					ScreenHeight: *height,
					Fullscreen:   false,
					VSync:        true,
					MasterVolume: 1.0,
					MusicVolume:  0.7,
					SFXVolume:    0.8,
					KeyBindings:  make(map[string]string),
				},
			}

			if err := saveManager.SaveGame(saveName, gameSave); err != nil {
				clientLogger.WithError(err).WithField("saveName", saveName).Error("failed to save game")
				return err
			}

			clientLogger.WithField("saveName", saveName).Info("game saved successfully")
			return nil
		}

		// Create load callback that reuses the quick load logic
		loadCallback := func(saveName string) error {
			if *verbose {
				clientLogger.WithField("saveName", saveName).Info("menu load")
			}

			gameSave, err := saveManager.LoadGame(saveName)
			if err != nil {
				clientLogger.WithError(err).WithField("saveName", saveName).Error("failed to load game")
				return err
			}

			// Restore player position
			if posComp, ok := player.GetComponent("position"); ok {
				pos := posComp.(*engine.PositionComponent)
				pos.X = gameSave.PlayerState.X
				pos.Y = gameSave.PlayerState.Y
			}

			// Restore player health
			if healthComp, ok := player.GetComponent("health"); ok {
				health := healthComp.(*engine.HealthComponent)
				health.Current = gameSave.PlayerState.CurrentHealth
				health.Max = gameSave.PlayerState.MaxHealth
			}

			// Restore player stats
			if statsComp, ok := player.GetComponent("stats"); ok {
				stats := statsComp.(*engine.StatsComponent)
				stats.Attack = gameSave.PlayerState.Attack
				stats.Defense = gameSave.PlayerState.Defense
				stats.MagicPower = gameSave.PlayerState.MagicPower
			}

			// Restore player level and XP
			if expComp, ok := player.GetComponent("experience"); ok {
				exp := expComp.(*engine.ExperienceComponent)
				exp.Level = gameSave.PlayerState.Level
				exp.CurrentXP = gameSave.PlayerState.Experience
			}

			clientLogger.WithField("saveName", saveName).Info("game loaded successfully")
			return nil
		}

		// Connect callbacks to menu system
		game.MenuSystem.SetSaveCallback(saveCallback)
		game.MenuSystem.SetLoadCallback(loadCallback)

		if *verbose {
			clientLogger.Info("save/load callbacks connected to menu system")
		}
	}

	// Process initial entity additions
	game.World.Update(0)

	clientLogger.Info("game initialized successfully")
	clientLogger.Info("controls: WASD to move, Space to attack, E to use item, I: Inventory, J: Quests")
	clientLogger.WithFields(logrus.Fields{"genre": *genreID, "seed": *seed}).Info("game settings")
	if *multiplayer {
		clientLogger.WithField("server", *server).Info("multiplayer connected")
	}

	// Setup cleanup handler for network client
	defer func() {
		if networkClient != nil {
			clientLogger.Info("disconnecting from server")
			if err := networkClient.Disconnect(); err != nil {
				clientLogger.WithError(err).Warn("error disconnecting")
			}
		}
	}()

	// Run the game loop
	windowTitle := fmt.Sprintf("Venture %s - Procedural Action RPG", version.Version)
	if err := game.Run(windowTitle); err != nil {
		clientLogger.WithError(err).Fatal("error running game")
	}
}
