// Package engine provides vehicle spawning utilities for V4.0 integration.
// This file implements functions to spawn procedurally generated vehicles into the game world.
package engine

import (
	"fmt"
	"image/color"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// VehicleSpawnData contains the necessary information to spawn a vehicle entity.
// This avoids import cycles by not depending on pkg/procgen/vehicle directly.
type VehicleSpawnData struct {
	Name         string
	VehicleType  VehicleType
	Components   []Component // Pre-generated components from Vehicle.ToComponents()
	Color        color.RGBA
	Size         int     // Sprite size
	ColliderSize float64 // Collider dimensions
}

// SpawnVehiclesInTerrain spawns procedurally generated vehicles into terrain rooms.
// Places vehicles in random rooms (avoiding first room - player spawn).
// Returns the number of vehicles spawned.
//
// Note: This function expects vehicle generation to be done externally to avoid import cycles.
// Call this from cmd/client with vehicles generated via vehiclegen.NewVehicleGenerator().
func SpawnVehiclesInTerrain(world *World, terr *terrain.Terrain, vehicles []VehicleSpawnData, seed int64) (int, error) {
	if len(terr.Rooms) < 2 {
		return 0, fmt.Errorf("insufficient rooms for vehicle spawning (need at least 2, got %d)", len(terr.Rooms))
	}

	if len(vehicles) == 0 {
		return 0, nil
	}

	// Create RNG for room selection
	rng := rand.New(rand.NewSource(seed + 5000))

	// Select random rooms (avoid first room - player spawn)
	availableRooms := make([]*terrain.Room, 0, len(terr.Rooms)-1)
	for i := 1; i < len(terr.Rooms); i++ {
		availableRooms = append(availableRooms, terr.Rooms[i])
	}

	if len(availableRooms) == 0 {
		return 0, fmt.Errorf("no available rooms for vehicle spawning")
	}

	// Shuffle rooms
	rng.Shuffle(len(availableRooms), func(i, j int) {
		availableRooms[i], availableRooms[j] = availableRooms[j], availableRooms[i]
	})

	spawned := 0
	for i, vehicleData := range vehicles {
		if i >= len(availableRooms) {
			break // No more rooms available
		}

		room := availableRooms[i]
		cx, cy := room.Center()

		// Add some randomness to position within room
		offsetX := rng.Float64()*20 - 10 // -10 to +10
		offsetY := rng.Float64()*20 - 10
		spawnX := float64(cx*32) + offsetX // 32 = tile size
		spawnY := float64(cy*32) + offsetY

		// Create vehicle entity
		entity := createVehicleEntity(world, vehicleData, spawnX, spawnY)
		if entity != nil {
			spawned++
		}
	}

	return spawned, nil
}

// createVehicleEntity creates a vehicle entity with the provided components.
func createVehicleEntity(world *World, vehicleData VehicleSpawnData, x, y float64) *Entity {
	entity := world.CreateEntity()

	// Add position
	entity.AddComponent(&PositionComponent{X: x, Y: y})

	// Add velocity (vehicles start stationary)
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})

	// Add all pre-generated components from the vehicle generator
	for _, comp := range vehicleData.Components {
		entity.AddComponent(comp)
	}

	// Add mount component (empty initially)
	entity.AddComponent(&MountComponent{
		MountedEntityID: 0, // No rider initially
	})

	// Add sprite
	sprite := NewSpriteComponent(float64(vehicleData.Size), float64(vehicleData.Size), vehicleData.Color)
	sprite.Layer = 8 // Below player (layer 10) but above ground
	sprite.Visible = true
	entity.AddComponent(sprite)

	// Add collider
	entity.AddComponent(&ColliderComponent{
		Width:     vehicleData.ColliderSize,
		Height:    vehicleData.ColliderSize,
		Solid:     true,
		IsTrigger: false,
		Layer:     2, // Vehicle layer
		OffsetX:   -vehicleData.ColliderSize / 2,
		OffsetY:   -vehicleData.ColliderSize / 2,
	})

	// Add team component (neutral initially)
	entity.AddComponent(&TeamComponent{TeamID: 0}) // Neutral team

	return entity
}
