//go:build !android && !ios
// +build !android,!ios

package main

import (
	"math/rand"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/faction"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/quest"
	"github.com/opd-ai/venture/pkg/procgen/recipe"
	"github.com/opd-ai/venture/pkg/procgen/station"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// systemsContainer holds all initialized game systems for dependency injection.
type systemsContainer struct {
	inputSystem            *engine.InputSystem
	movementSystem         *engine.MovementSystem
	collisionSystem        *engine.CollisionSystem
	combatSystem           *engine.CombatSystem
	interactionSystem      *engine.InteractionSystem
	particleSystem         *engine.ParticleSystem
	animationSystem        *engine.AnimationSystem
	equipmentVisualSystem  *engine.EquipmentVisualSystem
	objectiveTracker       *engine.ObjectiveTrackerSystem
	aiSystem               *engine.AISystem
	progressionSystem      *engine.ProgressionSystem
	inventorySystem        *engine.InventorySystem
	commerceSystem         *engine.CommerceSystem
	dialogSystem           *engine.DialogSystem
	craftingSystem         *engine.CraftingSystem
	audioManager           *engine.AudioManager
	audioManagerSystem     *engine.AudioManagerSystem
	itemPickupSystem       *engine.ItemPickupSystem
	statusEffectSystem     *engine.StatusEffectSystem
	spellCastingSystem     *engine.SpellCastingSystem
	playerSpellCasting     *engine.PlayerSpellCastingSystem
	manaRegenSystem        *engine.ManaRegenSystem
	playerCombatSystem     *engine.PlayerCombatSystem
	playerItemUseSystem    *engine.PlayerItemUseSystem
	rotationSystem         *engine.RotationSystem
	projectileSystem       *engine.ProjectileSystem
	revivalSystem          *engine.RevivalSystem
	behaviorTreeSystem     *engine.BehaviorTreeSystem
	squadSystem            *engine.SquadSystem
	factionSystem          *engine.FactionSystem
	reputationSystem       *engine.ReputationSystem
	alignmentSystem        *engine.AlignmentSystem
	factionReactionSystem  *engine.FactionReactionSystem
	skillProgressionSystem *engine.SkillProgressionSystem
	visualFeedbackSystem   *engine.VisualFeedbackSystem
	weatherSystem          *engine.WeatherSystem
	lifetimeSystem         *engine.LifetimeSystem
	puzzleSystem           *engine.PuzzleSystem
	firePropagationSystem  *engine.FirePropagationSystem
	destructibleSystem     *engine.DestructibleObjectSystem
	carrySystem            *engine.CarrySystem
	hazardSystem           *engine.HazardSystem
	narrativeSystem        *engine.NarrativeSystem
	shadowSystem           *engine.ShadowSystem
	spriteGenerator        *sprites.Generator
	itemGen                *item.ItemGenerator
	recipeGen              *recipe.RecipeGenerator
	statusEffectRNG        *rand.Rand
}

// initializeCoreSystems creates and initializes all core game systems.
func initializeCoreSystems(game *engine.EbitenGame, logger *logrus.Logger, clientLogger *logrus.Entry) *systemsContainer {
	clientLogger.Info("initializing game systems")

	sys := &systemsContainer{}

	// Core gameplay systems
	sys.inputSystem = engine.NewInputSystem()

	sys.movementSystem = engine.NewMovementSystem(playerMaxSpeed)
	sys.collisionSystem = engine.NewCollisionSystem(collisionGridCellSize)
	sys.movementSystem.SetCollisionSystem(sys.collisionSystem)

	sys.combatSystem = engine.NewCombatSystemWithLogger(*seed, logger)
	sys.interactionSystem = engine.NewInteractionSystem(game.World)
	sys.particleSystem = engine.NewParticleSystem()

	sys.spriteGenerator = sprites.NewGenerator()
	sys.animationSystem = engine.NewAnimationSystem(sys.spriteGenerator)
	sys.animationSystem.SetMaxCacheSize(animationCacheSize)

	sys.equipmentVisualSystem = engine.NewEquipmentVisualSystem(sys.spriteGenerator)

	return sys
}

// initializeGenerators creates item and recipe generators for loot drops.
func initializeGenerators(sys *systemsContainer) {
	sys.itemGen = item.NewItemGenerator()
	sys.recipeGen = recipe.NewRecipeGenerator()
}

// initializeObjectiveSystem creates and configures the objective tracker with callbacks.
func initializeObjectiveSystem(sys *systemsContainer, logger *logrus.Logger) {
	sys.objectiveTracker = engine.NewObjectiveTrackerSystem()

	sys.objectiveTracker.SetQuestCompleteCallback(func(entity *engine.Entity, qst *quest.Quest) {
		sys.objectiveTracker.AwardQuestRewards(entity, qst)
		if logger.GetLevel() >= logrus.InfoLevel {
			logging.ComponentLogger(logger, "quest").WithFields(logrus.Fields{
				"questName":   qst.Name,
				"xpReward":    qst.Reward.XP,
				"goldReward":  qst.Reward.Gold,
				"skillPoints": qst.Reward.SkillPoints,
			}).Info("quest completed")
		}
	})
}

// initializeProgressionSystems creates AI, progression, inventory, and commerce systems.
func initializeProgressionSystems(game *engine.EbitenGame, sys *systemsContainer, logger *logrus.Logger) {
	sys.aiSystem = engine.NewAISystem(game.World)
	sys.progressionSystem = engine.NewProgressionSystem(game.World)
	sys.inventorySystem = engine.NewInventorySystem(game.World)
	sys.commerceSystem = engine.NewCommerceSystemWithLogger(game.World, sys.inventorySystem, logger)
	sys.dialogSystem = engine.NewDialogSystemWithLogger(game.World, logger)
	sys.craftingSystem = engine.NewCraftingSystem(game.World, sys.inventorySystem, sys.itemGen)

	logging.ComponentLogger(logger, "commerce").Info("commerce system initialized")
	logging.ComponentLogger(logger, "dialog").Info("dialog system initialized")
	logging.ComponentLogger(logger, "crafting").Info("crafting system initialized")
}

// initializeAudioSystem creates and configures the audio manager with music playback.
func initializeAudioSystem(game *engine.EbitenGame, sys *systemsContainer, clientLogger *logrus.Entry) {
	sys.audioManager = engine.NewAudioManager(audioSampleRate, *seed)
	sys.audioManagerSystem = engine.NewAudioManagerSystem(sys.audioManager)

	game.SetAudioManager(sys.audioManager)
	sys.audioManager.EnableAdaptiveMusic(true)

	if *verbose {
		clientLogger.Info("adaptive music composition enabled with motif system")
	}

	if err := sys.audioManager.PlayMusic(*genreID, "exploration"); err != nil {
		logging.ComponentLogger(clientLogger.Logger, "audio").WithError(err).Warn("failed to start background music")
	}

	logging.ComponentLogger(clientLogger.Logger, "audio").Info("audio system initialized (music and SFX generators)")
}

// initializeCombatSystems creates spell casting and player combat systems.
func initializeCombatSystems(game *engine.EbitenGame, sys *systemsContainer) {
	sys.itemPickupSystem = engine.NewItemPickupSystem(game.World)

	sys.statusEffectRNG = rand.New(rand.NewSource(*seed + seedOffsetStatusEffect))
	sys.statusEffectSystem = engine.NewStatusEffectSystem(game.World, sys.statusEffectRNG)
	sys.spellCastingSystem = engine.NewSpellCastingSystem(game.World, sys.statusEffectSystem)
	sys.playerSpellCasting = engine.NewPlayerSpellCastingSystem(sys.spellCastingSystem, game.World)
	sys.manaRegenSystem = &engine.ManaRegenSystem{}
	sys.playerCombatSystem = engine.NewPlayerCombatSystem(sys.combatSystem, game.World)
	sys.playerItemUseSystem = engine.NewPlayerItemUseSystem(sys.inventorySystem, game.World)
}

// initializeEnvironmentalSystems creates weather, lifetime, puzzle, and destruction systems.
func initializeEnvironmentalSystems(game *engine.EbitenGame, sys *systemsContainer, clientLogger *logrus.Entry) {
	sys.weatherSystem = engine.NewWeatherSystem(game.World)
	sys.lifetimeSystem = engine.NewLifetimeSystemWithLogger(game.World, clientLogger.Logger)
	sys.puzzleSystem = engine.NewPuzzleSystem(game.World)

	sys.firePropagationSystem = engine.NewFirePropagationSystemWithLogger(tileSize, *seed+seedOffsetFirePropagation, clientLogger.Logger)
	sys.firePropagationSystem.SetWorld(game.World)

	sys.destructibleSystem = engine.NewDestructibleObjectSystemWithLogger(tileSize, *seed+seedOffsetDestructible, clientLogger.Logger)
	sys.destructibleSystem.SetWorld(game.World)
	sys.destructibleSystem.SetFireSystem(sys.firePropagationSystem)

	sys.carrySystem = engine.NewCarrySystemWithLogger(clientLogger.Logger)
	sys.carrySystem.SetWorld(game.World)

	sys.hazardSystem = engine.NewHazardSystemWithLogger(clientLogger.Logger)
	sys.hazardSystem.SetWorld(game.World)

	sys.narrativeSystem = engine.NewNarrativeSystem(game.World)
	sys.shadowSystem = engine.NewShadowSystemWithLogger(game.World, clientLogger.Logger)
}

// registerAllSystems adds all systems to the game world in the correct order.
func registerAllSystems(game *engine.EbitenGame, sys *systemsContainer) {
	game.World.AddSystem(sys.inputSystem)
	game.World.AddSystem(game.CameraSystem)

	sys.rotationSystem = engine.NewRotationSystem(game.World)
	game.World.AddSystem(&rotationSystemWrapper{system: sys.rotationSystem})

	game.World.AddSystem(sys.playerCombatSystem)
	game.World.AddSystem(sys.playerItemUseSystem)
	game.World.AddSystem(sys.playerSpellCasting)
	game.World.AddSystem(sys.movementSystem)
	game.World.AddSystem(sys.collisionSystem)

	sys.projectileSystem = engine.NewProjectileSystem(game.World)
	game.World.AddSystem(sys.projectileSystem)

	game.World.AddSystem(sys.combatSystem)
	game.World.AddSystem(sys.statusEffectSystem)

	sys.revivalSystem = engine.NewRevivalSystem(game.World)
	game.World.AddSystem(sys.revivalSystem)

	game.World.AddSystem(sys.aiSystem)

	sys.behaviorTreeSystem = engine.NewBehaviorTreeSystem(game.World)
	game.World.AddSystem(sys.behaviorTreeSystem)

	sys.squadSystem = engine.NewSquadSystem(game.World)
	game.World.AddSystem(&squadSystemWrapper{system: sys.squadSystem})

	game.World.AddSystem(sys.progressionSystem)

	sys.factionSystem = engine.NewFactionSystem(game.World, game.World.GetLogger().Logger)
	game.World.AddSystem(sys.factionSystem)

	sys.reputationSystem = engine.NewReputationSystem(game.World)
	game.World.AddSystem(&reputationSystemWrapper{system: sys.reputationSystem})

	sys.alignmentSystem = engine.NewAlignmentSystem(game.World)
	game.World.AddSystem(&alignmentSystemWrapper{system: sys.alignmentSystem})

	sys.factionReactionSystem = engine.NewFactionReactionSystem(game.World)
	game.World.AddSystem(&factionReactionSystemWrapper{system: sys.factionReactionSystem})

	sys.skillProgressionSystem = engine.NewSkillProgressionSystem()
	game.World.AddSystem(sys.skillProgressionSystem)

	sys.visualFeedbackSystem = engine.NewVisualFeedbackSystem()
	game.World.AddSystem(sys.visualFeedbackSystem)

	game.World.AddSystem(sys.audioManagerSystem)
	game.World.AddSystem(sys.objectiveTracker)
	game.World.AddSystem(sys.itemPickupSystem)
	game.World.AddSystem(sys.spellCastingSystem)
	game.World.AddSystem(sys.manaRegenSystem)
	game.World.AddSystem(sys.inventorySystem)
	game.World.AddSystem(sys.commerceSystem)
	game.World.AddSystem(sys.dialogSystem)
	game.World.AddSystem(sys.craftingSystem)
	game.World.AddSystem(sys.interactionSystem)

	game.World.AddSystem(&animationSystemWrapper{
		system: sys.animationSystem,
		logger: game.World.GetLogger(),
	})

	game.World.AddSystem(sys.equipmentVisualSystem)
	game.World.AddSystem(sys.particleSystem)
	game.World.AddSystem(sys.weatherSystem)
	game.World.AddSystem(sys.lifetimeSystem)
	game.World.AddSystem(sys.puzzleSystem)
	game.World.AddSystem(sys.firePropagationSystem)
	game.World.AddSystem(sys.destructibleSystem)
	game.World.AddSystem(sys.carrySystem)
	game.World.AddSystem(sys.hazardSystem)
	game.World.AddSystem(sys.narrativeSystem)
	game.World.AddSystem(sys.shadowSystem)
}

// configureSystemConnections wires up interdependent systems.
func configureSystemConnections(game *engine.EbitenGame, sys *systemsContainer) {
	sys.combatSystem.SetCamera(game.CameraSystem)
	sys.combatSystem.SetParticleSystem(sys.particleSystem, game.World, *genreID)
	sys.combatSystem.SetProjectileSystem(sys.projectileSystem)
	sys.projectileSystem.SetCamera(game.CameraSystem)
	sys.projectileSystem.SetGenre(*genreID)
	sys.projectileSystem.SetSeed(*seed)
	sys.interactionSystem.SetCarrySystem(sys.carrySystem)
}

// startPerformanceMonitoring begins background performance metric logging.
func startPerformanceMonitoring(game *engine.EbitenGame, clientLogger *logrus.Entry) {
	if !*verbose {
		return
	}

	perfMonitor := engine.NewPerformanceMonitor(game.World)
	clientLogger.Info("performance monitoring initialized")

	go func() {
		ticker := time.NewTicker(perfMonitorInterval * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			metrics := perfMonitor.GetMetrics()
			clientLogger.WithField("metrics", metrics.String()).Info("performance metrics")
		}
	}()
}

// generateWorldTerrain creates and validates procedural terrain.
func generateWorldTerrain(logger *logrus.Logger, clientLogger *logrus.Entry) *terrain.Terrain {
	clientLogger.Info("generating procedural terrain")

	terrainGen := terrain.NewBSPGeneratorWithLogger(logger)
	params := procgen.GenerationParams{
		Difficulty: defaultDifficulty,
		Depth:      defaultDepth,
		GenreID:    *genreID,
		Custom: map[string]interface{}{
			"width":  defaultTerrainWidth,
			"height": defaultTerrainHeight,
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

	return generatedTerrain
}

// initializeTerrainRendering sets up the terrain rendering system.
func initializeTerrainRendering(game *engine.EbitenGame, generatedTerrain *terrain.Terrain, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("initializing terrain rendering system")
	}

	terrainRenderSystem := engine.NewTerrainRenderSystem(tileSize, tileSize, *genreID, *seed)
	terrainRenderSystem.SetTerrain(generatedTerrain)
	game.TerrainRenderSystem = terrainRenderSystem

	if *verbose {
		clientLogger.Info("terrain rendering system initialized")
	}
}

// configureLightingSystem enables and configures dynamic lighting.
func configureLightingSystem(game *engine.EbitenGame, clientLogger *logrus.Entry) {
	if !*enableLighting {
		return
	}

	clientLogger.Info("enabling dynamic lighting system")
	game.EnableLighting(true)
	game.SetLightingGenrePreset(*genreID)

	clientLogger.WithFields(logrus.Fields{
		"genre":     *genreID,
		"enabled":   true,
		"maxLights": maxLights,
	}).Info("lighting system configured")
}

// initializeTerrainCollision sets up efficient terrain collision checking.
func initializeTerrainCollision(game *engine.EbitenGame, sys *systemsContainer, generatedTerrain *terrain.Terrain, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("initializing terrain collision system")
	}

	terrainChecker := engine.NewTerrainCollisionChecker(tileSize, tileSize)
	terrainChecker.SetTerrain(generatedTerrain)

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
}

// generateWorldFactions creates and registers all world factions.
func generateWorldFactions(game *engine.EbitenGame, params procgen.GenerationParams, clientLogger *logrus.Entry) {
	clientLogger.Info("generating world factions")

	factionGen := faction.NewGenerator()
	factionResult, err := factionGen.Generate(*seed+seedOffsetFaction, params)
	if err != nil {
		clientLogger.WithError(err).Fatal("failed to generate factions")
	}

	worldFactions := factionResult.([]*engine.Faction)
	clientLogger.WithFields(logrus.Fields{
		"count": len(worldFactions),
		"genre": *genreID,
	}).Info("factions generated")

	for _, fac := range worldFactions {
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
}

// initializeSpatialPartitioning creates and configures the spatial partition system.
func initializeSpatialPartitioning(game *engine.EbitenGame, generatedTerrain *terrain.Terrain, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("initializing spatial partition system for viewport culling")
	}

	worldWidth := float64(generatedTerrain.Width) * worldPixelsPerTile
	worldHeight := float64(generatedTerrain.Height) * worldPixelsPerTile

	spatialSystem := engine.NewSpatialPartitionSystem(worldWidth, worldHeight)
	game.World.AddSystem(spatialSystem)

	spatialSystem.Rebuild(game.World.GetEntities())

	game.RenderSystem.SetSpatialPartition(spatialSystem)
	game.RenderSystem.EnableCulling(true)
	game.RenderSystem.EnableBatching(true)

	clientLogger.WithFields(logrus.Fields{
		"worldWidth":     worldWidth,
		"worldHeight":    worldHeight,
		"cellSize":       quadtreeCapacity,
		"initialCount":   len(game.World.GetEntities()),
		"cullingActive":  true,
		"batchingActive": true,
	}).Info("spatial partition system initialized with initial rebuild, culling and batching enabled")
}

// connectMapUIToTerrain wires the map UI to terrain data.
func connectMapUIToTerrain(game *engine.EbitenGame, generatedTerrain *terrain.Terrain, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("connecting terrain to Map UI")
	}

	game.MapUI.SetTerrain(generatedTerrain)

	if *verbose {
		clientLogger.Info("Map UI configured with terrain data")
	}
}

// spawnWorldEntities spawns enemies, merchants, stations, puzzles, and objects.
func spawnWorldEntities(game *engine.EbitenGame, generatedTerrain *terrain.Terrain, clientLogger *logrus.Entry) {
	params := procgen.GenerationParams{
		Difficulty: defaultDifficulty,
		Depth:      defaultDepth,
		GenreID:    *genreID,
	}

	// Spawn enemies
	if *verbose {
		clientLogger.Info("spawning enemies in dungeon rooms")
	}
	enemyCount, err := engine.SpawnEnemiesInTerrain(game.World, generatedTerrain, *seed, params)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn enemies")
	} else if *verbose {
		clientLogger.WithFields(logrus.Fields{
			"enemyCount": enemyCount,
			"roomCount":  len(generatedTerrain.Rooms) - 1,
		}).Info("spawned enemies")
	}

	// Spawn merchants
	if *verbose {
		clientLogger.Info("spawning merchants in dungeon")
	}
	merchantCount, err := engine.SpawnMerchantsInTerrain(game.World, generatedTerrain, *seed, params, defaultMerchantCount)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn merchants")
	} else if *verbose {
		clientLogger.WithField("merchantCount", merchantCount).Info("spawned merchants")
	}

	// Spawn crafting stations
	if *verbose {
		clientLogger.Info("spawning crafting stations in dungeon")
	}
	stationGen := station.NewStationGenerator()
	stationCount := engine.SpawnStationsInTerrain(game.World, stationGen, generatedTerrain, tileSize, *seed+seedOffsetStation, *genreID)
	if *verbose {
		clientLogger.WithField("stationCount", stationCount).Info("spawned crafting stations")
	}

	// Spawn puzzles
	if *verbose {
		clientLogger.Info("spawning procedural puzzles in dungeon")
	}
	puzzleCount, err := engine.SpawnPuzzlesInTerrain(game.World, generatedTerrain, *seed+seedOffsetPuzzle, params, defaultPuzzleCount)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn puzzles")
	} else {
		clientLogger.WithFields(logrus.Fields{
			"puzzleCount": puzzleCount,
			"targetCount": defaultPuzzleCount,
			"roomCount":   len(generatedTerrain.Rooms) - 1,
		}).Info("spawned procedural puzzles")
	}

	// Spawn destructible objects
	if *verbose {
		clientLogger.Info("spawning destructible objects in dungeon")
	}
	objectCount := spawnDestructibleObjects(game.World, generatedTerrain, *seed+seedOffsetObject, *genreID, clientLogger.Logger)
	clientLogger.WithFields(logrus.Fields{
		"objectCount": objectCount,
		"roomCount":   len(generatedTerrain.Rooms) - 1,
		"genre":       *genreID,
	}).Info("spawned destructible objects")
}

// spawnEnvironmentalEffects spawns lights and weather effects.
func spawnEnvironmentalEffects(game *engine.EbitenGame, generatedTerrain *terrain.Terrain, clientLogger *logrus.Entry) {
	if *enableLighting {
		if *verbose {
			clientLogger.Info("spawning environmental lights in dungeon")
		}
		lightCount := spawnEnvironmentalLights(game.World, generatedTerrain, *seed+seedOffsetLight, *genreID)
		clientLogger.WithFields(logrus.Fields{
			"lightCount": lightCount,
			"genre":      *genreID,
		}).Info("spawned environmental lights")
	}

	if *enableWeather {
		if *verbose {
			clientLogger.Info("spawning weather effects")
		}
		weatherEntity := spawnWeather(game.World, *width, *height, *seed+seedOffsetWeather, *genreID, *weatherType, *weatherIntensity)
		if weatherEntity != nil {
			clientLogger.WithFields(logrus.Fields{
				"type":      *weatherType,
				"intensity": *weatherIntensity,
				"genre":     *genreID,
			}).Info("weather effects spawned")
		}
	}
}

// calculatePlayerSpawnPosition determines where the player should spawn.
func calculatePlayerSpawnPosition(generatedTerrain *terrain.Terrain, clientLogger *logrus.Entry) (float64, float64) {
	if len(generatedTerrain.Rooms) > 0 {
		firstRoom := generatedTerrain.Rooms[0]
		cx, cy := firstRoom.Center()
		playerX := float64(cx * tileSize)
		playerY := float64(cy * tileSize)

		if *verbose {
			clientLogger.WithFields(logrus.Fields{
				"tileX":  cx,
				"tileY":  cy,
				"worldX": playerX,
				"worldY": playerY,
			}).Info("player spawning in first room")
		}

		return playerX, playerY
	}

	clientLogger.Warn("no rooms in terrain, using default spawn position")
	return fallbackPlayerX, fallbackPlayerY
}
