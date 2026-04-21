//go:build !android && !ios
// +build !android,!ios

// Package main contains the client entry point and game initialization.
// This file contains entity spawning and world population initialization.

package main

import (
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/minigame"
	"github.com/opd-ai/venture/pkg/procgen/station"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// spawnWorldEntities spawns enemies, merchants, stations, puzzles, and objects.
func spawnWorldEntities(game *engine.EbitenGame, generatedTerrain *terrain.Terrain, sys *systemsContainer, clientLogger *logrus.Entry) {
	params := procgen.GenerationParams{
		Difficulty: defaultDifficulty,
		Depth:      defaultDepth,
		GenreID:    *genreID,
		Custom: map[string]interface{}{
			"theme": getGenreTheme(),
		},
	}

	spawnEnemiesWithLogging(game.World, generatedTerrain, params, clientLogger)
	spawnMerchantsWithLogging(game.World, generatedTerrain, params, clientLogger)
	spawnMinigamesWithLogging(game.World, sys.minigameGenerator, params, clientLogger)
	spawnStationsWithLogging(game.World, generatedTerrain, clientLogger)
	spawnPuzzlesWithLogging(game.World, generatedTerrain, params, clientLogger)
	spawnObjectsWithLogging(game.World, generatedTerrain, clientLogger)
	spawnVehiclesWithLogging(game.World, generatedTerrain, params, clientLogger)
	spawnCompanionsWithLogging(game.World, generatedTerrain, params, clientLogger, sys.companionLearningSys)
	spawnBookshelvesWithLogging(game.World, generatedTerrain, params, clientLogger)
	spawnStoryFragmentsWithLogging(game.World, generatedTerrain, params, clientLogger)
	generateNarrativeArcWithLogging(game.World, sys.narrativeGenerator, params, clientLogger, sys.timeProvider)
}

// spawnEnemiesWithLogging spawns enemies in terrain with optional verbose logging.
func spawnEnemiesWithLogging(w *engine.World, generatedTerrain *terrain.Terrain, params procgen.GenerationParams, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("spawning enemies in dungeon rooms")
	}
	enemyCount, err := engine.SpawnEnemiesInTerrain(w, generatedTerrain, *seed, params)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn enemies")
	} else if *verbose {
		clientLogger.WithFields(logrus.Fields{
			"enemyCount": enemyCount,
			"roomCount":  len(generatedTerrain.Rooms) - 1,
		}).Info("spawned enemies")
	}
}

// spawnMerchantsWithLogging spawns merchants in terrain with optional verbose logging.
func spawnMerchantsWithLogging(w *engine.World, generatedTerrain *terrain.Terrain, params procgen.GenerationParams, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("spawning merchants in dungeon")
	}
	merchantCount, err := engine.SpawnMerchantsInTerrain(w, generatedTerrain, *seed, params, defaultMerchantCount)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn merchants")
	} else if *verbose {
		clientLogger.WithField("merchantCount", merchantCount).Info("spawned merchants")
	}
}

// spawnMinigamesWithLogging adds procedural minigames to merchants with optional verbose logging.
// Phase 3.5 (PLAN.md): Minigames in shops/taverns for player entertainment
func spawnMinigamesWithLogging(w *engine.World, minigameGen *minigame.Generator, params procgen.GenerationParams, clientLogger *logrus.Entry) {
	if minigameGen == nil {
		return
	}
	if *verbose {
		clientLogger.Info("adding minigames to merchants")
	}
	minigameCount := addMinigamesToMerchants(w, minigameGen, *seed, params, clientLogger)
	if *verbose {
		clientLogger.WithField("minigameCount", minigameCount).Info("added minigames to merchants")
	}
}

// spawnStationsWithLogging spawns crafting stations in terrain with optional verbose logging.
func spawnStationsWithLogging(w *engine.World, generatedTerrain *terrain.Terrain, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("spawning crafting stations in dungeon")
	}
	stationGen := station.NewStationGenerator()
	stationCount := engine.SpawnStationsInTerrain(w, stationGen, generatedTerrain, tileSize, *seed+seedOffsetStation, *genreID)
	if *verbose {
		clientLogger.WithField("stationCount", stationCount).Info("spawned crafting stations")
	}
}

// spawnPuzzlesWithLogging spawns procedural puzzles in terrain with optional verbose logging.
func spawnPuzzlesWithLogging(w *engine.World, generatedTerrain *terrain.Terrain, params procgen.GenerationParams, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("spawning procedural puzzles in dungeon")
	}
	puzzleCount, err := engine.SpawnPuzzlesInTerrain(w, generatedTerrain, *seed+seedOffsetPuzzle, params, defaultPuzzleCount)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn puzzles")
	} else {
		clientLogger.WithFields(logrus.Fields{
			"puzzleCount": puzzleCount,
			"targetCount": defaultPuzzleCount,
			"roomCount":   len(generatedTerrain.Rooms) - 1,
		}).Info("spawned procedural puzzles")
	}
}

// spawnObjectsWithLogging spawns destructible objects in terrain with logging.
func spawnObjectsWithLogging(w *engine.World, generatedTerrain *terrain.Terrain, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("spawning destructible objects in dungeon")
	}
	objectCount := spawnDestructibleObjects(w, generatedTerrain, *seed+seedOffsetObject, *genreID, clientLogger.Logger)
	clientLogger.WithFields(logrus.Fields{
		"objectCount": objectCount,
		"roomCount":   len(generatedTerrain.Rooms) - 1,
		"genre":       *genreID,
	}).Info("spawned destructible objects")
}

// spawnVehiclesWithLogging spawns vehicles in terrain with optional verbose logging.
func spawnVehiclesWithLogging(w *engine.World, generatedTerrain *terrain.Terrain, params procgen.GenerationParams, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("spawning vehicles in dungeon")
	}
	vehicleCount, err := spawnVehicles(w, generatedTerrain, *seed+seedOffsetVehicle, params, clientLogger)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn vehicles")
	} else if *verbose {
		clientLogger.WithField("vehicleCount", vehicleCount).Info("spawned vehicles")
	}
}

// spawnCompanionsWithLogging spawns companions in terrain with optional verbose logging.
func spawnCompanionsWithLogging(w *engine.World, generatedTerrain *terrain.Terrain, params procgen.GenerationParams, clientLogger *logrus.Entry, companionLearningSys *engine.CompanionLearningSystem) {
	companionCount := spawnCompanionsWithVerboseLogging(w, generatedTerrain, params, clientLogger)
	initializeCompanionLearning(w, companionLearningSys, companionCount, clientLogger)
}

// spawnCompanionsWithVerboseLogging spawns companions and logs the result.
func spawnCompanionsWithVerboseLogging(w *engine.World, generatedTerrain *terrain.Terrain, params procgen.GenerationParams, clientLogger *logrus.Entry) int {
	if *verbose {
		clientLogger.Info("spawning companions in dungeon")
	}
	companionCount, err := spawnCompanions(w, generatedTerrain, *seed+seedOffsetCompanion, params, clientLogger)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn companions")
		return 0
	}
	if *verbose {
		clientLogger.WithField("companionCount", companionCount).Info("spawned companions")
	}
	return companionCount
}

// initializeCompanionLearning initializes learning for all spawned companions.
func initializeCompanionLearning(w *engine.World, companionLearningSys *engine.CompanionLearningSystem, companionCount int, clientLogger *logrus.Entry) {
	if companionLearningSys == nil {
		return
	}

	companionEntities := w.GetEntitiesWith("companion")
	for _, entity := range companionEntities {
		if _, hasLearning := entity.GetComponent("companion_learning"); hasLearning {
			continue
		}

		learningRate := calculateLearningRate(companionCount)
		if err := companionLearningSys.AddCompanionLearning(entity.ID, learningRate); err != nil {
			clientLogger.WithError(err).WithField("companionID", entity.ID).Warn("failed to initialize companion learning")
		} else if *verbose {
			clientLogger.WithFields(logrus.Fields{
				"companionID":  entity.ID,
				"learningRate": learningRate,
			}).Debug("initialized companion learning")
		}
	}
}

// calculateLearningRate calculates learning rate based on companion count.
func calculateLearningRate(companionCount int) float64 {
	learningRate := 1.0 + (float64(companionCount) * 0.1)
	if learningRate > 2.0 {
		learningRate = 2.0
	}
	return learningRate
}

// spawnBookshelvesWithLogging spawns bookshelves with books in terrain with optional verbose logging.
func spawnBookshelvesWithLogging(w *engine.World, generatedTerrain *terrain.Terrain, params procgen.GenerationParams, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("spawning bookshelves in dungeon")
	}
	bookshelfCount, err := spawnBookshelves(w, generatedTerrain, *seed+seedOffsetBook, params, clientLogger)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn bookshelves")
	} else if *verbose {
		clientLogger.WithField("bookshelfCount", bookshelfCount).Info("spawned bookshelves")
	}
}

// spawnStoryFragmentsWithLogging spawns story fragments in terrain with optional verbose logging.
func spawnStoryFragmentsWithLogging(w *engine.World, generatedTerrain *terrain.Terrain, params procgen.GenerationParams, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("spawning story fragments in dungeon")
	}
	fragmentCount, err := spawnStoryFragments(w, generatedTerrain, *seed+seedOffsetStory, params, clientLogger)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn story fragments")
	} else if *verbose {
		clientLogger.WithField("fragmentCount", fragmentCount).Info("spawned story fragments")
	}

	// Phase 30 / AUDIT.md G4: ArchaeologicalSiteComponent, TimelineComponent, and
	// CrossDungeonStoryComponent are defined but no engine system currently queries
	// or updates them. Spawning inert entities would silently create dead content.
	// These spawners will be re-enabled once a runtime consumer (DiscoverySystem
	// integration, excavation progression, or UI hook) is wired.
	clientLogger.WithFields(logrus.Fields{
		"system": "spawn_init",
		"gated": []string{
			"archaeological_sites",
			"timelines",
			"cross_dungeon_stories",
		},
		"reason": "runtime consumers for ArchaeologicalSiteComponent, TimelineComponent, and CrossDungeonStoryComponent are not yet wired",
	}).Debug("skipping inert story-content spawns until engine integration exists")
}

// spawnEnvironmentalEffects spawns lights, weather effects, and environmental hazards (unconditionally enabled).
func spawnEnvironmentalEffects(game *engine.EbitenGame, generatedTerrain *terrain.Terrain, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("spawning environmental lights in dungeon")
	}
	lightCount := spawnEnvironmentalLights(game.World, generatedTerrain, *seed+seedOffsetLight, *genreID)
	clientLogger.WithFields(logrus.Fields{
		"lightCount": lightCount,
		"genre":      *genreID,
	}).Info("spawned environmental lights")

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

	// Phase 3.4: Spawn environmental hazards (fire pits, acid pools, spike traps, etc.)
	if *verbose {
		clientLogger.Info("spawning environmental hazards")
	}
	hazardCount := spawnEnvironmentalHazards(game.World, generatedTerrain, *seed+seedOffsetEnvironment, *genreID, clientLogger)
	clientLogger.WithFields(logrus.Fields{
		"hazardCount": hazardCount,
		"genre":       *genreID,
	}).Info("spawned environmental hazards")
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
