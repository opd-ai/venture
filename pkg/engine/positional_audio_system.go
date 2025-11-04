package engine

import (
	"math"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// PositionalAudioSystem updates positional audio parameters based on entity positions.
// Calculates stereo panning, distance-based volume falloff, and occlusion for all
// entities with PositionalAudioComponent.
// Phase 14.4: Audio System Enhancement
type PositionalAudioSystem struct {
	world           *World
	logger          *logrus.Entry
	listenerX       float64 // Listener position (camera/player)
	listenerY       float64
	terrainWidth    int // For occlusion detection
	terrainHeight   int
	terrainTiles    [][]terrain.TileType // Reference to terrain for occlusion
	performanceMode bool                 // Skip occlusion checks if true
}

// NewPositionalAudioSystem creates a new positional audio system.
func NewPositionalAudioSystem(world *World) *PositionalAudioSystem {
	return NewPositionalAudioSystemWithLogger(world, nil)
}

// NewPositionalAudioSystemWithLogger creates a new positional audio system with a logger.
func NewPositionalAudioSystemWithLogger(world *World, logger *logrus.Logger) *PositionalAudioSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"system": "positional_audio",
		})
	}

	return &PositionalAudioSystem{
		world:           world,
		logger:          logEntry,
		performanceMode: false,
	}
}

// SetListener updates the listener position (typically the player's position).
func (s *PositionalAudioSystem) SetListener(x, y float64) {
	s.listenerX = x
	s.listenerY = y
}

// SetTerrain provides terrain data for occlusion detection.
func (s *PositionalAudioSystem) SetTerrain(width, height int, tiles [][]terrain.TileType) {
	s.terrainWidth = width
	s.terrainHeight = height
	s.terrainTiles = tiles
}

// SetPerformanceMode enables/disables performance mode.
// When enabled, skips expensive occlusion calculations.
func (s *PositionalAudioSystem) SetPerformanceMode(enabled bool) {
	s.performanceMode = enabled
}

// Update calculates positional audio parameters for all entities.
func (s *PositionalAudioSystem) Update(deltaTime float64) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"listenerX": s.listenerX,
			"listenerY": s.listenerY,
		}).Debug("updating positional audio")
	}

	// Get all entities with PositionalAudioComponent
	entities := s.world.GetEntitiesWith("positional_audio", "position")

	for _, entity := range entities {
		s.updateEntity(entity)
	}
}

// updateEntity calculates audio parameters for a single entity.
func (s *PositionalAudioSystem) updateEntity(entity *Entity) {
	audioComp, ok := entity.GetComponent("positional_audio")
	if !ok {
		return
	}
	audioCompTyped, ok := audioComp.(*PositionalAudioComponent)
	if !ok {
		return
	}

	posComp, ok := entity.GetComponent("position")
	if !ok {
		return
	}
	posCompTyped, ok := posComp.(*PositionComponent)
	if !ok {
		return
	}

	// Update audio component position from entity position
	audioCompTyped.X = posCompTyped.X
	audioCompTyped.Y = posCompTyped.Y

	// Calculate distance to listener
	distance := CalculateDistance(audioCompTyped.X, audioCompTyped.Y, s.listenerX, s.listenerY)

	// Calculate volume falloff
	audioCompTyped.VolumeMultiplier = CalculateVolumeFalloff(distance, audioCompTyped.MaxDistance, audioCompTyped.FalloffExponent)

	// Calculate stereo pan
	audioCompTyped.StereoPan = CalculateStereoPan(audioCompTyped.X, audioCompTyped.Y, s.listenerX, s.listenerY)

	// Check occlusion
	if audioCompTyped.OcclusionEnabled && !s.performanceMode {
		audioCompTyped.IsOccluded = s.checkOcclusion(audioCompTyped.X, audioCompTyped.Y, s.listenerX, s.listenerY)
		if audioCompTyped.IsOccluded {
			audioCompTyped.VolumeMultiplier *= audioCompTyped.OcclusionFactor
		}
	} else {
		audioCompTyped.IsOccluded = false
	}
}

// checkOcclusion determines if a sound path is blocked by walls.
// Uses simple raycasting to detect wall intersections.
func (s *PositionalAudioSystem) checkOcclusion(sourceX, sourceY, targetX, targetY float64) bool {
	if s.terrainTiles == nil || s.terrainWidth == 0 || s.terrainHeight == 0 {
		return false // No terrain data, assume no occlusion
	}

	// Calculate direction and distance
	dx := targetX - sourceX
	dy := targetY - sourceY
	distance := math.Sqrt(dx*dx + dy*dy)

	if distance < 1.0 {
		return false // Too close to matter
	}

	// Normalize direction
	dirX := dx / distance
	dirY := dy / distance

	// Sample along the ray (every 16 pixels = half tile)
	const sampleStep = 16.0
	numSamples := int(distance / sampleStep)

	for i := 1; i < numSamples; i++ {
		// Sample point along ray
		sampleX := sourceX + dirX*float64(i)*sampleStep
		sampleY := sourceY + dirY*float64(i)*sampleStep

		// Convert to tile coordinates (assuming 32x32 pixel tiles)
		tileX := int(sampleX / 32)
		tileY := int(sampleY / 32)

		// Check bounds
		if tileX < 0 || tileX >= s.terrainWidth || tileY < 0 || tileY >= s.terrainHeight {
			continue
		}

		// Check if tile is a wall
		if s.terrainTiles[tileY][tileX] == terrain.TileWall ||
			s.terrainTiles[tileY][tileX] == terrain.TileWallNE ||
			s.terrainTiles[tileY][tileX] == terrain.TileWallNW ||
			s.terrainTiles[tileY][tileX] == terrain.TileWallSE ||
			s.terrainTiles[tileY][tileX] == terrain.TileWallSW {
			return true // Occluded by wall
		}
	}

	return false // No walls in the way
}

// GetAudioParameters returns volume and pan for an entity with PositionalAudioComponent.
// Returns (volume, pan, ok) where ok is false if the entity doesn't have the component.
func (s *PositionalAudioSystem) GetAudioParameters(entity *Entity) (volume float64, pan float64, ok bool) {
	audioComp, hasAudio := entity.GetComponent("positional_audio")
	if !hasAudio {
		return 0.0, 0.0, false
	}
	audioCompTyped, ok := audioComp.(*PositionalAudioComponent)
	if !ok {
		return 0.0, 0.0, false
	}

	return audioCompTyped.VolumeMultiplier, audioCompTyped.StereoPan, true
}
