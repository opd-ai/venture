// Package vehicle provides vehicle-specific physics including suspension and weight transfer.
package vehicle

import (
	"math"
	"math/rand"
)

// TerrainDeformationComponent manages tire tracks and terrain deformation.
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

	// Reusable buffer for GetVisibleTracks to avoid per-call allocations
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
		visibleBuffer:   make([]TrackMark, 0, 200), // Pre-allocate buffer for GetVisibleTracks
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

// AddTrack creates a new track mark at the specified position.
// wheelLoad: force on wheel (affects depth), vehicleAngle: direction of travel
func (t *TerrainDeformationComponent) AddTrack(x, y, vehicleAngle, wheelLoad float64, terrainType TerrainType) {
	// Check spacing to avoid too many close tracks
	dx := x - t.LastTrackX
	dy := y - t.LastTrackY
	distance := math.Sqrt(dx*dx + dy*dy)

	if distance < t.MinTrackSpacing {
		return // Too close to last track
	}

	// Get deformation parameters for this terrain
	baseDepth := t.DeformationDepth[terrainType]
	fadeTime := t.FadeTime[terrainType]

	if baseDepth <= 0 || fadeTime <= 0 {
		return // No tracks on this terrain
	}

	// Calculate depth based on wheel load
	// Higher load = deeper tracks
	// Normalize load to reasonable range (assume 0-10000 N)
	loadFactor := math.Min(wheelLoad/5000.0, 2.0)
	depth := baseDepth * loadFactor
	if depth > 1.0 {
		depth = 1.0
	}

	// Add some noise to depth for realism
	depthNoise := t.rng.Float64()*0.1 - 0.05 // ±5%
	depth = math.Max(0.0, math.Min(1.0, depth+depthNoise))

	// Track width varies slightly (8-12 pixels)
	width := 10.0 + t.rng.Float64()*4.0 - 2.0

	// Create track mark
	track := TrackMark{
		X:           x,
		Y:           y,
		Angle:       vehicleAngle,
		Depth:       depth,
		Width:       width,
		Age:         0.0,
		TerrainType: terrainType,
		FadeTime:    fadeTime,
	}

	// Add to collection
	t.Tracks = append(t.Tracks, track)

	// Update last track position
	t.LastTrackX = x
	t.LastTrackY = y

	// Enforce max tracks limit (remove oldest)
	if len(t.Tracks) > t.MaxTracks {
		// Remove oldest 10% to avoid frequent reallocations
		removeCount := t.MaxTracks / 10
		if removeCount < 1 {
			removeCount = 1
		}
		t.Tracks = t.Tracks[removeCount:]
	}
}

// Update ages existing tracks and removes faded ones.
func (t *TerrainDeformationComponent) Update(deltaTime float64) {
	// Age all tracks
	for i := range t.Tracks {
		t.Tracks[i].Age += deltaTime
	}

	// Remove fully faded tracks
	// Work backwards to safely remove elements
	for i := len(t.Tracks) - 1; i >= 0; i-- {
		track := &t.Tracks[i]
		if track.Age >= track.FadeTime {
			// Remove this track
			t.Tracks = append(t.Tracks[:i], t.Tracks[i+1:]...)
		}
	}
}

// GetVisibleTracks returns all tracks that should be rendered.
// minX, minY, maxX, maxY: viewport bounds for culling
// Returns a slice that is reused on subsequent calls - do not modify or store.
func (t *TerrainDeformationComponent) GetVisibleTracks(minX, minY, maxX, maxY float64) []TrackMark {
	// Reset buffer length while preserving capacity
	t.visibleBuffer = t.visibleBuffer[:0]

	for i := range t.Tracks {
		track := &t.Tracks[i]

		// Simple AABB culling
		if track.X >= minX && track.X <= maxX && track.Y >= minY && track.Y <= maxY {
			t.visibleBuffer = append(t.visibleBuffer, *track)
		}
	}

	return t.visibleBuffer
}

// GetTrackAlpha calculates the opacity of a track based on age and fade time.
// Returns value in range [0.0, 1.0] where 0.0 is fully faded, 1.0 is fresh.
func (t *TerrainDeformationComponent) GetTrackAlpha(track *TrackMark) float64 {
	if track.FadeTime <= 0 {
		return 0.0
	}

	// Linear fade
	alpha := 1.0 - (track.Age / track.FadeTime)
	if alpha < 0.0 {
		alpha = 0.0
	}

	return alpha
}

// Clear removes all track marks.
func (t *TerrainDeformationComponent) Clear() {
	t.Tracks = t.Tracks[:0]
}

// GetTrackCount returns the current number of track marks.
func (t *TerrainDeformationComponent) GetTrackCount() int {
	return len(t.Tracks)
}

// GetTerrainTypeFromTile converts a tile type to terrain type for deformation.
// This is a helper function for integration with the terrain system.
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
