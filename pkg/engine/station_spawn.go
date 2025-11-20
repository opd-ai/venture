// Package engine provides helper functions for spawning crafting stations in the game world.
// This file bridges procedural generation (pkg/procgen/station) with the ECS runtime,
// converting StationData into engine entities with proper components.
package engine

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/station"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// StationData represents procedurally generated crafting station data.
// This avoids import cycle between pkg/procgen/station and pkg/engine.
type StationData struct {
	StationType int // 0=alchemy, 1=forge, 2=workbench
	Name        string
	SpawnX      float64
	SpawnY      float64
	Seed        int64
}

// SpawnStationFromData converts procedural StationData into an engine entity.
// This function creates the entity, adds all required components (position, sprite,
// collider, crafting station), and registers it with the world.
//
// Returns the spawned station entity or nil if spawning fails.
func SpawnStationFromData(world *World, stationData *StationData, x, y float64) *Entity {
	if stationData == nil {
		return nil
	}

	// Create station entity
	stationEntity := world.CreateEntity()

	// Add position
	stationEntity.AddComponent(&PositionComponent{X: x, Y: y})

	// Add sprite (distinct visual for stations)
	stationSprite := &EbitenSprite{
		Image:   ebiten.NewImage(32, 32),
		Width:   32,
		Height:  32,
		Visible: true,
		Layer:   9, // Below player/NPCs, above terrain
	}
	stationEntity.AddComponent(stationSprite)

	// Add animation component for visual distinction
	stationAnim := NewAnimationComponent(stationData.Seed)
	stationAnim.CurrentState = AnimationStateIdle
	stationAnim.FrameTime = 0.5 // Slow animation for ambient effect
	stationAnim.Loop = true
	stationAnim.Playing = true
	stationAnim.FrameCount = 2 // Simple pulsing effect
	stationEntity.AddComponent(stationAnim)

	// Add collider (stations are solid, non-walkable)
	stationEntity.AddComponent(&ColliderComponent{
		Width:     32,
		Height:    32,
		Solid:     true,
		IsTrigger: false,
		Layer:     1,
		OffsetX:   -16,
		OffsetY:   -16,
	})

	// Map StationType to RecipeType
	var recipeType RecipeType
	switch stationData.StationType {
	case 0: // StationAlchemyTable
		recipeType = RecipePotion
	case 1: // StationForge
		recipeType = RecipeEnchanting
	case 2: // StationWorkbench
		recipeType = RecipeMagicItem
	default:
		recipeType = RecipePotion
	}

	// Add crafting station component with bonuses
	stationComp := NewCraftingStationComponent(recipeType)
	stationEntity.AddComponent(stationComp)

	return stationEntity
}

// SpawnStationsInTerrain spawns crafting stations at deterministic locations in the terrain.
// Uses station generator to create 3 stations (alchemy, forge, workbench) and places them
// on walkable tiles at safe distances from each other.
//
// Parameters:
//   - world: The ECS world to add stations to
//   - stationGen: The station generator (procgen.Generator)
//   - terrain: The terrain data for tile validation
//   - tileSize: Size of a single tile in pixels
//   - seed: Seed for deterministic generation
//   - genreID: Genre for themed station names
//
// Returns the number of stations spawned.
func SpawnStationsInTerrain(world *World, stationGen procgen.Generator, terrainData *terrain.Terrain, tileSize int, seed int64, genreID string) int {
	if world == nil || stationGen == nil || terrainData == nil {
		return 0
	}

	stations := generateStations(stationGen, seed, genreID)
	if stations == nil {
		return 0
	}

	spawnPoints := generateStationSpawnPoints(seed, terrainData.Width, terrainData.Height, tileSize, 200.0)

	return spawnStationsAtPoints(world, stations, spawnPoints, terrainData, tileSize)
}

// generateStations creates station data using the generator.
func generateStations(stationGen procgen.Generator, seed int64, genreID string) []*station.StationData {
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    genreID,
	}

	result, err := stationGen.Generate(seed, params)
	if err != nil {
		return nil
	}

	stations, ok := result.([]*station.StationData)
	if !ok {
		return nil
	}
	return stations
}

// spawnStationsAtPoints spawns each station at its designated point with validation.
func spawnStationsAtPoints(world *World, stations []*station.StationData, spawnPoints []struct{ X, Y float64 }, terrainData *terrain.Terrain, tileSize int) int {
	spawnedCount := 0
	for i, stationData := range stations {
		if i >= len(spawnPoints) {
			break
		}

		point := spawnPoints[i]
		validPoint := validateAndAdjustSpawnPoint(&point, terrainData, tileSize)
		if !validPoint {
			continue
		}

		if spawnSingleStation(world, stationData, point) {
			spawnedCount++
		}
	}
	return spawnedCount
}

// validateAndAdjustSpawnPoint validates spawn location and adjusts to nearest walkable tile if needed.
func validateAndAdjustSpawnPoint(point *struct{ X, Y float64 }, terrainData *terrain.Terrain, tileSize int) bool {
	tileX := int(point.X) / tileSize
	tileY := int(point.Y) / tileSize

	if tileX < 0 || tileX >= terrainData.Width || tileY < 0 || tileY >= terrainData.Height {
		return false
	}

	tile := terrainData.GetTile(tileX, tileY)
	if tile != terrain.TileFloor {
		return findNearbyWalkableTile(&tileX, &tileY, terrainData, point, tileSize)
	}
	return true
}

// findNearbyWalkableTile searches for a walkable tile near the current position.
func findNearbyWalkableTile(tileX, tileY *int, terrainData *terrain.Terrain, point *struct{ X, Y float64 }, tileSize int) bool {
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			newTileX := *tileX + dx
			newTileY := *tileY + dy
			if newTileX >= 0 && newTileX < terrainData.Width &&
				newTileY >= 0 && newTileY < terrainData.Height {
				if terrainData.GetTile(newTileX, newTileY) == terrain.TileFloor {
					*tileX = newTileX
					*tileY = newTileY
					point.X = float64(newTileX*tileSize + tileSize/2)
					point.Y = float64(newTileY*tileSize + tileSize/2)
					return true
				}
			}
		}
	}
	return false
}

// spawnSingleStation converts station data and spawns it at the specified point.
func spawnSingleStation(world *World, stationData *station.StationData, point struct{ X, Y float64 }) bool {
	localStationData := &StationData{
		StationType: int(stationData.StationType),
		Name:        stationData.Name,
		SpawnX:      point.X,
		SpawnY:      point.Y,
		Seed:        stationData.Seed,
	}

	stationEntity := SpawnStationFromData(world, localStationData, point.X, point.Y)
	return stationEntity != nil
}

// generateStationSpawnPoints returns deterministic spawn positions for stations.
// Uses seed-based RNG to place stations in safe locations (walkable tiles, not too close to each other).
// Returns slice of (x, y) coordinate pairs.
func generateStationSpawnPoints(seed int64, terrainWidth, terrainHeight, tileSize int, minDistanceBetweenStations float64) []struct{ X, Y float64 } {
	rng := rand.New(rand.NewSource(seed))

	// Calculate usable spawn area (avoid edges)
	margin := 3 * tileSize // 3 tiles from edge
	minX := float64(margin)
	maxX := float64(terrainWidth*tileSize - margin)
	minY := float64(margin)
	maxY := float64(terrainHeight*tileSize - margin)

	// Generate 3 spawn points with minimum distance
	spawnPoints := make([]struct{ X, Y float64 }, 0, 3)

	for i := 0; i < 3; i++ {
		maxAttempts := 50
		for attempt := 0; attempt < maxAttempts; attempt++ {
			x := minX + rng.Float64()*(maxX-minX)
			y := minY + rng.Float64()*(maxY-minY)

			// Check minimum distance to existing stations
			validPosition := true
			for _, existing := range spawnPoints {
				dx := x - existing.X
				dy := y - existing.Y
				distance := dx*dx + dy*dy // squared distance (no sqrt for performance)
				if distance < minDistanceBetweenStations*minDistanceBetweenStations {
					validPosition = false
					break
				}
			}

			if validPosition {
				spawnPoints = append(spawnPoints, struct{ X, Y float64 }{X: x, Y: y})
				break
			}
		}

		// If we couldn't find a valid position after max attempts, place anyway
		if len(spawnPoints) < i+1 {
			x := minX + rng.Float64()*(maxX-minX)
			y := minY + rng.Float64()*(maxY-minY)
			spawnPoints = append(spawnPoints, struct{ X, Y float64 }{X: x, Y: y})
		}
	}

	return spawnPoints
}

// GetNearbyStations returns all crafting station entities within the specified radius.
// Useful for finding stations a player can interact with.
//
// Parameters:
//   - entities: Slice of all entities to search
//   - centerX, centerY: Center point to search from
//   - radius: Maximum distance from center
//
// Returns slice of station entities within radius.
func GetNearbyStations(entities []Entity, centerX, centerY, radius float64) []*Entity {
	var nearbyStations []*Entity
	radiusSquared := radius * radius

	for i := range entities {
		entity := &entities[i]

		// Check if entity has crafting station component
		if _, hasStation := entity.GetComponent("crafting_station"); !hasStation {
			continue
		}

		// Check position component
		posComp, hasPos := entity.GetComponent("position")
		if !hasPos {
			continue
		}

		pos, ok := posComp.(*PositionComponent)
		if !ok {
			continue
		}

		// Calculate squared distance (avoid sqrt for performance)
		dx := pos.X - centerX
		dy := pos.Y - centerY
		distSquared := dx*dx + dy*dy

		if distSquared <= radiusSquared {
			nearbyStations = append(nearbyStations, entity)
		}
	}

	return nearbyStations
}

// FindClosestStation finds the crafting station entity nearest to the specified position.
// Returns nil if no stations exist or none are within maxDistance.
//
// Parameters:
//   - entities: Slice of all entities to search
//   - centerX, centerY: Center point to search from
//   - maxDistance: Maximum distance to consider (or -1 for infinite)
//
// Returns the closest station entity and its distance, or nil and 0 if none found.
func FindClosestStation(entities []Entity, centerX, centerY, maxDistance float64) (*Entity, float64) {
	var closestStation *Entity
	closestDist := math.MaxFloat64

	if maxDistance > 0 {
		closestDist = maxDistance
	}

	for i := range entities {
		entity := &entities[i]

		// Check if entity has crafting station component
		if _, hasStation := entity.GetComponent("crafting_station"); !hasStation {
			continue
		}

		// Check position component
		posComp, hasPos := entity.GetComponent("position")
		if !hasPos {
			continue
		}

		pos, ok := posComp.(*PositionComponent)
		if !ok {
			continue
		}

		// Calculate distance
		dx := pos.X - centerX
		dy := pos.Y - centerY
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist < closestDist {
			closestDist = dist
			closestStation = entity
		}
	}

	if closestStation == nil {
		return nil, 0
	}

	return closestStation, closestDist
}

// GetStationInteractionPrompt returns a UI prompt string for station interaction.
// Shows station type and recipe category it supports.
//
// Returns empty string if entity is not a valid station.
func GetStationInteractionPrompt(stationEntity *Entity) string {
	if stationEntity == nil {
		return ""
	}

	stationComp, hasStation := stationEntity.GetComponent("crafting_station")
	if !hasStation {
		return ""
	}

	station, ok := stationComp.(*CraftingStationComponent)
	if !ok {
		return ""
	}

	recipeTypeName := ""
	switch station.StationType {
	case RecipePotion:
		recipeTypeName = "Potions"
	case RecipeEnchanting:
		recipeTypeName = "Enchanting"
	case RecipeMagicItem:
		recipeTypeName = "Magic Items"
	}

	stationName := "Crafting Station"
	switch station.StationType {
	case RecipePotion:
		stationName = "Alchemy Table"
	case RecipeEnchanting:
		stationName = "Forge"
	case RecipeMagicItem:
		stationName = "Workbench"
	}

	return fmt.Sprintf("[U] Use %s (%s)", stationName, recipeTypeName)
}
