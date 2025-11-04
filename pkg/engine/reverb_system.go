package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// ReverbSystem manages reverb effects based on room acoustics.
// Simulates realistic reverb by analyzing room size and materials.
// Phase 14.4: Audio System Enhancement
type ReverbSystem struct {
	world         *World
	logger        *logrus.Entry
	currentReverb *ReverbComponent // Active reverb for current room
	terrainWidth  int
	terrainHeight int
	terrainTiles  [][]terrain.TileType
	materialMap   map[terrain.TileType]RoomMaterial // Tile to material mapping
	rng           *rand.Rand
}

// NewReverbSystem creates a new reverb system.
func NewReverbSystem(world *World, seed int64) *ReverbSystem {
	return NewReverbSystemWithLogger(world, seed, nil)
}

// NewReverbSystemWithLogger creates a new reverb system with a logger.
func NewReverbSystemWithLogger(world *World, seed int64, logger *logrus.Logger) *ReverbSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"system": "reverb",
		})
	}

	// Create default tile to material mapping
	materialMap := map[terrain.TileType]RoomMaterial{
		terrain.TileWall:         RoomMaterialStone,
		terrain.TileWallNE:       RoomMaterialStone,
		terrain.TileWallNW:       RoomMaterialStone,
		terrain.TileWallSE:       RoomMaterialStone,
		terrain.TileWallSW:       RoomMaterialStone,
		terrain.TileFloor:        RoomMaterialStone,
		terrain.TileDoor:         RoomMaterialWood,
		terrain.TilePlatform:     RoomMaterialWood,
		terrain.TileBridge:       RoomMaterialWood,
		terrain.TileWaterShallow: RoomMaterialStone, // Water reflects sound
		terrain.TileLavaFlow:     RoomMaterialMetal, // Lava/metal has similar acoustics
	}

	return &ReverbSystem{
		world:       world,
		logger:      logEntry,
		materialMap: materialMap,
		rng:         rand.New(rand.NewSource(seed)),
	}
}

// SetTerrain provides terrain data for room analysis.
func (s *ReverbSystem) SetTerrain(width, height int, tiles [][]terrain.TileType) {
	s.terrainWidth = width
	s.terrainHeight = height
	s.terrainTiles = tiles

	// Analyze terrain and create reverb component
	s.analyzeRoomAcoustics()
}

// SetMaterialMapping allows customization of tile to material mapping.
func (s *ReverbSystem) SetMaterialMapping(tileType terrain.TileType, material RoomMaterial) {
	s.materialMap[tileType] = material
}

// Update processes reverb for entities with ReverbComponent.
func (s *ReverbSystem) Update(deltaTime float64) {
	// Currently, reverb is primarily environmental (global room effect)
	// Individual entities with ReverbComponent could have custom reverb
	// but for now we focus on the global room reverb

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		if s.currentReverb != nil {
			s.logger.WithFields(logrus.Fields{
				"roomSize":  s.currentReverb.RoomSize.String(),
				"material":  s.currentReverb.Material.String(),
				"decayTime": s.currentReverb.DecayTime,
			}).Debug("reverb active")
		}
	}
}

// analyzeRoomAcoustics determines reverb parameters from terrain.
func (s *ReverbSystem) analyzeRoomAcoustics() {
	if s.terrainTiles == nil || s.terrainWidth == 0 || s.terrainHeight == 0 {
		s.currentReverb = nil
		return
	}

	// Count floor tiles to estimate room size
	floorCount := 0
	for y := 0; y < s.terrainHeight; y++ {
		for x := 0; x < s.terrainWidth; x++ {
			if s.terrainTiles[y][x] == terrain.TileFloor {
				floorCount++
			}
		}
	}

	// Determine room size based on floor area
	roomSize := s.calculateRoomSize(floorCount)

	// Determine dominant material from walls
	material := s.analyzeMaterials()

	// Create reverb component
	s.currentReverb = NewReverbComponent(roomSize, material)

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.InfoLevel {
		s.logger.WithFields(logrus.Fields{
			"roomSize":   roomSize.String(),
			"material":   material.String(),
			"floorTiles": floorCount,
			"decayTime":  s.currentReverb.DecayTime,
		}).Info("room acoustics analyzed")
	}
}

// calculateRoomSize determines room size category from floor tile count.
func (s *ReverbSystem) calculateRoomSize(floorCount int) RoomSize {
	// Room size thresholds based on typical tile counts
	// Small: <100 tiles (5x5 to 10x10)
	// Medium: 100-400 tiles (10x10 to 20x20)
	// Large: 400-1200 tiles (20x20 to 35x35)
	// Huge: >1200 tiles (35x35+)

	if floorCount < 100 {
		return RoomSizeSmall
	} else if floorCount < 400 {
		return RoomSizeMedium
	} else if floorCount < 1200 {
		return RoomSizeLarge
	} else {
		return RoomSizeHuge
	}
}

// analyzeMaterials determines dominant material from terrain tiles.
func (s *ReverbSystem) analyzeMaterials() RoomMaterial {
	if s.terrainTiles == nil {
		return RoomMaterialStone // Default
	}

	// Count occurrences of each material
	materialCounts := make(map[RoomMaterial]int)

	for y := 0; y < s.terrainHeight; y++ {
		for x := 0; x < s.terrainWidth; x++ {
			tileType := s.terrainTiles[y][x]
			if material, ok := s.materialMap[tileType]; ok {
				materialCounts[material]++
			}
		}
	}

	// Find most common material
	var dominantMaterial RoomMaterial = RoomMaterialStone
	maxCount := 0

	for material, count := range materialCounts {
		if count > maxCount {
			maxCount = count
			dominantMaterial = material
		}
	}

	return dominantMaterial
}

// GetCurrentReverb returns the current environmental reverb settings.
func (s *ReverbSystem) GetCurrentReverb() *ReverbComponent {
	return s.currentReverb
}

// ApplyReverbToSamples applies reverb effect to audio samples (conceptual).
// This is a simplified reverb that would work with the actual audio pipeline.
// Returns modified audio data with reverb applied.
func (s *ReverbSystem) ApplyReverbToSamples(samples []float64, sampleRate int) []float64 {
	if s.currentReverb == nil || !s.currentReverb.Enabled || s.currentReverb.Amount == 0 {
		return samples // No reverb
	}

	// Simplified reverb using comb filter approach
	// In a real implementation, this would use proper convolution reverb

	numSamples := len(samples)
	output := make([]float64, numSamples)

	// Calculate delay buffer size from decay time
	// Use shorter delays for the buffer to create multiple early reflections
	delay1 := int(0.029 * float64(sampleRate)) // ~29ms
	delay2 := int(0.037 * float64(sampleRate)) // ~37ms
	delay3 := int(0.041 * float64(sampleRate)) // ~41ms
	delay4 := int(0.043 * float64(sampleRate)) // ~43ms

	// Clamp delays to buffer size
	if delay1 >= numSamples {
		delay1 = numSamples / 4
	}
	if delay2 >= numSamples {
		delay2 = numSamples / 3
	}
	if delay3 >= numSamples {
		delay3 = numSamples / 2
	}
	if delay4 >= numSamples {
		delay4 = (numSamples * 2) / 3
	}

	// Feedback based on decay time and damping
	feedback := 0.5 * (s.currentReverb.DecayTime / 3.0) * (1.0 - s.currentReverb.Damping)
	if feedback > 0.85 {
		feedback = 0.85 // Prevent instability
	}
	wetLevel := s.currentReverb.Amount
	dryLevel := 1.0 - wetLevel

	// Process samples with multiple comb filters
	for i := 0; i < numSamples; i++ {
		dry := samples[i]

		// Get delayed samples from multiple delay lines (early reflections)
		var wet float64
		if i >= delay1 {
			wet += output[i-delay1] * feedback * 0.37
		}
		if i >= delay2 {
			wet += output[i-delay2] * feedback * 0.33
		}
		if i >= delay3 {
			wet += output[i-delay3] * feedback * 0.19
		}
		if i >= delay4 {
			wet += output[i-delay4] * feedback * 0.11
		}

		// Mix dry and wet signals
		output[i] = dry*dryLevel + dry*wetLevel + wet*wetLevel
	}

	return output
}

// GetReverbParameters returns reverb parameters for audio processing.
// Returns (decayTime, damping, amount, enabled).
func (s *ReverbSystem) GetReverbParameters() (decayTime, damping, amount float64, enabled bool) {
	if s.currentReverb == nil {
		return 0, 0, 0, false
	}

	return s.currentReverb.DecayTime,
		s.currentReverb.Damping,
		s.currentReverb.Amount,
		s.currentReverb.Enabled
}
