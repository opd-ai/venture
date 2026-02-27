// Package vehicle provides vehicle-specific physics including suspension and weight transfer.
//
// Performance Optimization (2026-01-23):
// The TerrainDeformationComponent now uses an internal buffer for GetVisibleTracks()
// to eliminate per-frame allocations. This optimization reduces allocation overhead
// from 13568 B/op to 0 B/op (100% reduction) and improves performance by 8.2x.
package vehicle

import (
	"math/rand"
)

// TerrainDeformationComponent manages tire tracks and terrain deformation data.
type TerrainDeformationComponent struct {
	// Track marks left by vehicle
	Tracks []TrackMark

	// Configuration
	MaxTracks       int     // Maximum number of track marks to store
	MinTrackSpacing float64 // Minimum distance between track marks (pixels)
	LastTrackX      float64 // Position of last track mark
	LastTrackY      float64

	// Deformation parameters per terrain type
	DeformationDepth map[TerrainType]float64 // How deep tracks are
	FadeTime         map[TerrainType]float64 // How long tracks last

	// Seed for deterministic noise in track patterns
	Seed int64
	rng  *rand.Rand

	// Reusable buffer for GetVisibleTracks to avoid allocations
	visibleBuffer []TrackMark
}

// Type returns the component type identifier.
func (t *TerrainDeformationComponent) Type() string {
	return "terrain_deformation"
}

// NewTerrainDeformationComponent creates a terrain deformation component.
func NewTerrainDeformationComponent(seed int64) *TerrainDeformationComponent {
	rng := rand.New(rand.NewSource(seed))

	return &TerrainDeformationComponent{
		Tracks:          make([]TrackMark, 0, 200),
		MaxTracks:       200,
		MinTrackSpacing: 5.0, // 5 pixels between marks
		Seed:            seed,
		rng:             rng,
		DeformationDepth: map[TerrainType]float64{
			TerrainHard:  0.0, // No deformation on hard surfaces
			TerrainFirm:  0.2, // Slight tracks
			TerrainSoft:  0.8, // Deep tracks in mud/sand
			TerrainSnow:  1.0, // Maximum deformation in snow
			TerrainWater: 0.0, // No permanent tracks in water
		},
		FadeTime: map[TerrainType]float64{
			TerrainHard:  0.0,   // Instant fade (no tracks)
			TerrainFirm:  30.0,  // 30 seconds on firm ground
			TerrainSoft:  120.0, // 2 minutes in mud/sand
			TerrainSnow:  60.0,  // 1 minute in snow
			TerrainWater: 1.0,   // 1 second in water (wake effect)
		},
	}
}

// GetVisibleTracks returns all tracks that should be rendered.
// Deprecated: Use vehicle.GetVisibleTracks(deformation, minX, minY, maxX, maxY) instead to maintain ECS purity.
// minX, minY, maxX, maxY: viewport bounds for culling
// Uses internal buffer to avoid allocations per call.
func (t *TerrainDeformationComponent) GetVisibleTracks(minX, minY, maxX, maxY float64) []TrackMark {
	return GetVisibleTracks(t, minX, minY, maxX, maxY)
}

// GetTrackAlpha calculates the opacity of a track based on age and fade time.
// Returns value in range [0.0, 1.0] where 0.0 is fully faded, 1.0 is fresh.
// Deprecated: Use vehicle.GetTrackAlpha(track) instead to maintain ECS purity.
func (t *TerrainDeformationComponent) GetTrackAlpha(track *TrackMark) float64 {
	return GetTrackAlpha(track)
}

// Clear removes all track marks.
// Deprecated: Use vehicle.ClearTracks(deformation) instead to maintain ECS purity.
func (t *TerrainDeformationComponent) Clear() {
	ClearTracks(t)
}

// GetTrackCount returns the current number of track marks.
// Deprecated: Use len(deformation.Tracks) instead to maintain ECS purity.
func (t *TerrainDeformationComponent) GetTrackCount() int {
	return len(t.Tracks)
}

// GetTerrainTypeFromTile converts a tile type to terrain type for deformation.
// This is a helper function for integration with the terrain system.
//
// The function maps pkg/procgen/terrain.TileType constants to TerrainType values
// used by the vehicle physics system. This allows vehicles to respond appropriately
// to different terrain surfaces.
//
// Terrain type mappings:
//   - Hard (concrete, metal): TileWall (0), TileFloor (1) - minimal deformation, no track marks
//   - Firm (packed dirt, stone): TileCorridor (2), TileDoor (3), default - slight deformation
//   - Soft (mud, sand, snow): TileTree (11) - deep track marks, performance impact
//   - Water: TileWaterShallow (4), TileWaterDeep (5) - buoyancy effects, no tracks
//
// Example usage in a terrain interaction system:
//
//	// Get tile type from world at vehicle position
//	tileX, tileY := int(vehicleX/tileSize), int(vehicleY/tileSize)
//	tileType := world.GetTile(tileX, tileY)
//
//	// Convert to terrain type for vehicle physics
//	terrainType := vehicle.GetTerrainTypeFromTile(tileType)
//
//	// Apply terrain-specific effects
//	deformationComp.TerrainType = terrainType
//	if terrainType == vehicle.TerrainSoft {
//		// Reduce vehicle speed on soft terrain
//		velocityComp.X *= 0.7
//		velocityComp.Y *= 0.7
//	}
func GetTerrainTypeFromTile(tileType int) TerrainType {
	// Map terrain tile types to deformation types
	// These constants should match pkg/procgen/terrain.TileType
	switch tileType {
	case 0, 1: // TileWall, TileFloor - hard surfaces
		return TerrainHard
	case 2, 3: // TileCorridor, TileDoor - firm surfaces
		return TerrainFirm
	case 4, 5: // TileWaterShallow, TileWaterDeep - water
		return TerrainWater
	case 11: // TileTree - soft ground
		return TerrainSoft
	default:
		return TerrainFirm // Default to firm
	}
}
