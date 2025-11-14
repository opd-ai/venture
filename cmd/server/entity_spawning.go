//go:build !android && !ios
// +build !android,!ios

// Package main provides server-side entity spawning for V4.0 entities.
// This file ports spawning logic from cmd/client/util.go for server use.
package main

import (
	"fmt"
	"image/color"

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
	// Import vehicle generator
	vehicleGen := vehicle.NewVehicleGenerator()

	// Determine number of vehicles based on room count (2-5 vehicles per level)
	roomCount := len(terrainMap.Rooms)
	if roomCount < 2 {
		return 0, nil // Need at least 2 rooms
	}

	vehicleCount := 2 + (roomCount-2)/4
	if vehicleCount > 5 {
		vehicleCount = 5
	}

	// Generate vehicles
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

	// Convert vehicles to spawn data
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
			Components:   v.ToComponents(),
			Color:        colorRGBA,
			Size:         size,
			ColliderSize: colliderSize,
		}
	}

	// Spawn vehicles in terrain
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

	companionCount := 1 + (roomCount-2)/5 // 1 companion per 5 rooms
	if companionCount > 3 {
		companionCount = 3 // Cap at 3
	}

	// Generate companions
	companionResult, err := companionGen.Generate(seed, params)
	if err != nil {
		return 0, fmt.Errorf("failed to generate companions: %w", err)
	}

	companions, ok := companionResult.([]*companion.Companion)
	if !ok || len(companions) == 0 {
		return 0, fmt.Errorf("companion generator returned invalid result")
	}

	if len(companions) > companionCount {
		companions = companions[:companionCount]
	}

	// Convert to spawn data
	companionSpawnData := make([]engine.CompanionSpawnData, len(companions))
	for i, c := range companions {
		var engineType engine.CompanionType
		switch c.Type {
		case companion.Pet:
			engineType = engine.CompanionTypePet
		case companion.Summon:
			engineType = engine.CompanionTypeSummon
		case companion.Hireling:
			engineType = engine.CompanionTypeHireling
		case companion.Elemental:
			engineType = engine.CompanionTypeElemental
		case companion.Undead:
			engineType = engine.CompanionTypeUndead
		case companion.Robot:
			engineType = engine.CompanionTypeRobot
		case companion.Spirit:
			engineType = engine.CompanionTypeSpirit
		case companion.Insect:
			engineType = engine.CompanionTypeInsect
		}

		r := uint8((c.Color >> 16) & 0xFF)
		g := uint8((c.Color >> 8) & 0xFF)
		b := uint8(c.Color & 0xFF)
		colorRGBA := color.RGBA{R: r, G: g, B: b, A: 255}

		size := 24
		if c.Type == companion.Elemental || c.Type == companion.Robot {
			size = 28
		}

		companionSpawnData[i] = engine.CompanionSpawnData{
			Name:          c.Name,
			CompanionType: engineType,
			Stats:         c.Stats,
			Color:         colorRGBA,
			Size:          size,
			ColliderSize:  float64(size - 4),
		}
	}

	// Spawn companions
	spawned, err := engine.SpawnCompanionsInTerrain(world, terrainMap, companionSpawnData, seed)
	if err != nil {
		return 0, fmt.Errorf("failed to spawn companions in terrain: %w", err)
	}

	if logger.GetLevel() >= logrus.DebugLevel {
		logger.WithFields(logrus.Fields{
			"requested": companionCount,
			"generated": len(companions),
			"spawned":   spawned,
		}).Debug("companion spawning complete")
	}

	return spawned, nil
}

// spawnBookshelvesInTerrain generates and spawns bookshelves for the server.
// Returns the number of bookshelves spawned.
func spawnBookshelvesInTerrain(world *engine.World, terrainMap *terrain.Terrain, seed int64, genreID string, logger *logrus.Logger) (int, error) {
	roomCount := len(terrainMap.Rooms)
	if roomCount < 3 {
		return 0, nil // Need at least 3 rooms
	}

	bookshelfCount := 1 + (roomCount-3)/6 // 1 bookshelf per 6 rooms
	if bookshelfCount > 4 {
		bookshelfCount = 4 // Cap at 4
	}

	spawned := 0

	// Spawn bookshelves using engine utility
	spawned, err := engine.SpawnBookshelvesInTerrain(world, terrainMap, seed, genreID, bookshelfCount)
	if err != nil {
		return 0, fmt.Errorf("failed to spawn bookshelves: %w", err)
	}

	if logger.GetLevel() >= logrus.DebugLevel {
		logger.WithFields(logrus.Fields{
			"requested": bookshelfCount,
			"spawned":   spawned,
		}).Debug("bookshelf spawning complete")
	}

	return spawned, nil
}
