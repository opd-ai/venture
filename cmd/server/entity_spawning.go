//go:build !android && !ios
// +build !android,!ios

// Package main provides server-side entity spawning for V4.0 entities.
// This file ports spawning logic from cmd/client/util.go for server use.
package main

import (
	"fmt"
	"image/color"
	"math/rand"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/companion"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/procgen/vehicle"
	"github.com/sirupsen/logrus"
)

// spawnVehiclesInTerrain generates and spawns vehicles for the server.
// Returns the number of vehicles spawned.
func spawnVehiclesInTerrain(world *engine.World, terrainMap *terrain.Terrain, seed int64, params procgen.GenerationParams, logger *logrus.Logger) (int, error) {
	vehicleGen := vehicle.NewVehicleGenerator()

	roomCount := len(terrainMap.Rooms)
	if roomCount < 2 {
		return 0, nil
	}

	vehicleCount := 2 + (roomCount-2)/4
	if vehicleCount > 5 {
		vehicleCount = 5
	}

	vehicleResult, err := vehicleGen.Generate(seed, params)
	if err != nil {
		return 0, fmt.Errorf("failed to generate vehicles: %w", err)
	}

	vehicles, ok := vehicleResult.([]*vehicle.Vehicle)
	if !ok || len(vehicles) == 0 {
		return 0, fmt.Errorf("vehicle generator returned invalid result")
	}

	if len(vehicles) > vehicleCount {
		vehicles = vehicles[:vehicleCount]
	}

	vehicleSpawnData := make([]engine.VehicleSpawnData, len(vehicles))
	for i, v := range vehicles {
		var engineType engine.VehicleType
		switch v.VehicleType {
		case vehicle.TypeMount:
			engineType = engine.VehicleMount
		case vehicle.TypeCart:
			engineType = engine.VehicleCart
		case vehicle.TypeBoat:
			engineType = engine.VehicleBoat
		case vehicle.TypeGlider:
			engineType = engine.VehicleGlider
		case vehicle.TypeMech:
			engineType = engine.VehicleMech
		}

		r := uint8((v.Color >> 16) & 0xFF)
		g := uint8((v.Color >> 8) & 0xFF)
		b := uint8(v.Color & 0xFF)
		colorRGBA := color.RGBA{R: r, G: g, B: b, A: 255}

		var size int
		var colliderSize float64
		switch v.VehicleType {
		case vehicle.TypeMount:
			size, colliderSize = 32, 28.0
		case vehicle.TypeCart:
			size, colliderSize = 40, 36.0
		case vehicle.TypeBoat:
			size, colliderSize = 48, 44.0
		case vehicle.TypeGlider:
			size, colliderSize = 36, 32.0
		case vehicle.TypeMech:
			size, colliderSize = 44, 40.0
		default:
			size, colliderSize = 32, 28.0
		}

		vehicleSpawnData[i] = engine.VehicleSpawnData{
			Name:         v.Name,
			VehicleType:  engineType,
			Components:   v.ToComponents(), // Generate components from vehicle
			Color:        colorRGBA,
			Size:         size,
			ColliderSize: colliderSize,
		}
	}

	spawned, err := engine.SpawnVehiclesInTerrain(world, terrainMap, vehicleSpawnData, seed)
	if err != nil {
		return 0, fmt.Errorf("failed to spawn vehicles in terrain: %w", err)
	}

	if logger.GetLevel() >= logrus.DebugLevel {
		logger.WithFields(logrus.Fields{
			"requested": vehicleCount,
			"generated": len(vehicles),
			"spawned":   spawned,
		}).Debug("vehicle spawning complete")
	}

	return spawned, nil
}

// spawnCompanionsInTerrain generates and spawns companions for the server.
// Returns the number of companions spawned.
func spawnCompanionsInTerrain(world *engine.World, terrainMap *terrain.Terrain, seed int64, params procgen.GenerationParams, logger *logrus.Logger) (int, error) {
	companionGen := companion.NewGenerator()
	roomCount := len(terrainMap.Rooms)
	if roomCount < 2 {
		return 0, nil
	}

	companionCount := calculateCompanionCount(roomCount)
	companionSpawnData := generateCompanionData(companionGen, companionCount, seed, params, logger)

	if len(companionSpawnData) == 0 {
		return 0, nil
	}

	spawned, err := engine.SpawnCompanionsInTerrain(world, terrainMap, companionSpawnData, seed)
	if err != nil {
		return 0, fmt.Errorf("failed to spawn companions in terrain: %w", err)
	}

	logCompanionSpawning(logger, companionCount, len(companionSpawnData), spawned)
	return spawned, nil
}

// calculateCompanionCount determines number of companions based on room count.
func calculateCompanionCount(roomCount int) int {
	companionCount := 1 + (roomCount-2)/5
	if companionCount > 3 {
		companionCount = 3
	}
	return companionCount
}

// generateCompanionData creates spawn data for all companions.
func generateCompanionData(companionGen procgen.Generator, companionCount int, seed int64, params procgen.GenerationParams, logger *logrus.Logger) []engine.CompanionSpawnData {
	companionSpawnData := make([]engine.CompanionSpawnData, 0, companionCount)
	for i := 0; i < companionCount; i++ {
		companionSeed := seed + int64(i*1000)
		companionResult, err := companionGen.Generate(companionSeed, params)
		if err != nil {
			if logger.GetLevel() >= logrus.WarnLevel {
				logger.WithError(err).Warn("failed to generate companion")
			}
			continue
		}

		comp, ok := companionResult.(*companion.Companion)
		if !ok {
			if logger.GetLevel() >= logrus.WarnLevel {
				logger.Warn("companion generator returned invalid result")
			}
			continue
		}

		companionSpawnData = append(companionSpawnData, createCompanionSpawnData(comp))
	}
	return companionSpawnData
}

// createCompanionSpawnData creates spawn data from a companion.
func createCompanionSpawnData(comp *companion.Companion) engine.CompanionSpawnData {
	companionColor := color.RGBA{R: 100, G: 150, B: 200, A: 255}
	size, colliderSize := getCompanionSizeForType(comp.Type)

	return engine.CompanionSpawnData{
		Name:          comp.Name,
		CompanionType: comp.Type,
		Level:         comp.Level,
		Attack:        comp.Attack,
		Defense:       comp.Defense,
		Speed:         comp.Speed,
		HP:            comp.HP,
		MaxHP:         comp.MaxHP,
		Loyalty:       comp.Loyalty,
		Commands:      comp.Commands,
		Color:         companionColor,
		Size:          size,
		ColliderSize:  colliderSize,
	}
}

// getCompanionSizeForType returns size and collider size for companion type.
func getCompanionSizeForType(companionType engine.CompanionType) (int, float64) {
	switch companionType {
	case engine.CompanionTypePet:
		return 24, 20.0
	case engine.CompanionTypeSummon:
		return 28, 24.0
	case engine.CompanionTypeHireling:
		return 28, 24.0
	case engine.CompanionTypeElemental:
		return 32, 28.0
	case engine.CompanionTypeUndead:
		return 30, 26.0
	case engine.CompanionTypeRobot:
		return 30, 26.0
	case engine.CompanionTypeSpirit:
		return 26, 22.0
	case engine.CompanionTypeInsect:
		return 22, 18.0
	default:
		return 28, 24.0
	}
}

// logCompanionSpawning logs companion spawning results.
func logCompanionSpawning(logger *logrus.Logger, requested, generated, spawned int) {
	if logger.GetLevel() >= logrus.DebugLevel {
		logger.WithFields(logrus.Fields{
			"requested": requested,
			"generated": generated,
			"spawned":   spawned,
		}).Debug("companion spawning complete")
	}
}

// spawnBookshelvesInTerrain generates and spawns bookshelves for the server.
// Returns the number of bookshelves spawned.
func spawnBookshelvesInTerrain(world *engine.World, terrainMap *terrain.Terrain, seed int64, params procgen.GenerationParams, logger *logrus.Logger) (int, error) {
	roomCount := len(terrainMap.Rooms)
	if roomCount < 3 {
		return 0, nil
	}

	bookshelfCount := 1 + (roomCount-3)/6
	if bookshelfCount > 4 {
		bookshelfCount = 4
	}

	bookshelfSpawnData := make([]engine.BookshelfSpawnData, 0, bookshelfCount)
	for i := 0; i < bookshelfCount; i++ {
		bookshelfSeed := seed + int64(i*5000)
		booksPerShelf := 3 + (i % 6) // 3-8 books

		books := make([]*engine.BookComponent, 0, booksPerShelf)
		for j := 0; j < booksPerShelf; j++ {
			bookSeed := bookshelfSeed + int64(j*100)

			// Simple book generation for server (no dependencies on book generator)
			bookTypes := []engine.BookType{
				engine.BookTypeSkill,
				engine.BookTypeLore,
				engine.BookTypeRecipe,
				engine.BookTypeHistory,
			}
			rng := rand.New(rand.NewSource(bookSeed))
			bookType := bookTypes[rng.Intn(len(bookTypes))]

			// Create simple book component
			book := &engine.BookComponent{
				Title:    fmt.Sprintf("Book %d-%d", i, j),
				BookType: bookType,
				Content:  []string{"Server-generated book content."},
			}
			books = append(books, book)
		}

		if len(books) == 0 {
			continue
		}

		// Simple brown wood color for server
		shelfColor := color.RGBA{R: 139, G: 69, B: 19, A: 255}

		bookshelfSpawnData = append(bookshelfSpawnData, engine.BookshelfSpawnData{
			Books:        books,
			ShelfColor:   shelfColor,
			ShelfSize:    32,
			ColliderSize: 28.0,
		})
	}

	if len(bookshelfSpawnData) == 0 {
		return 0, nil
	}

	spawned, err := engine.SpawnBookshelvesInTerrain(world, terrainMap, bookshelfSpawnData, seed)
	if err != nil {
		return 0, fmt.Errorf("failed to spawn bookshelves: %w", err)
	}

	if logger.GetLevel() >= logrus.DebugLevel {
		logger.WithFields(logrus.Fields{
			"requested": bookshelfCount,
			"generated": len(bookshelfSpawnData),
			"spawned":   spawned,
		}).Debug("bookshelf spawning complete")
	}

	return spawned, nil
}
