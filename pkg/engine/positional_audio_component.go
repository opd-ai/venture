package engine

import (
	"math"
)

// PositionalAudioComponent adds 3D spatial audio to entities.
// Enables stereo panning, distance-based volume falloff, and occlusion.
// Phase 14.4: Audio System Enhancement
type PositionalAudioComponent struct {
	// Sound source position (updated automatically from PositionComponent)
	X, Y float64

	// Maximum audible distance (in pixels)
	MaxDistance float64

	// Falloff exponent (1.0 = linear, 2.0 = inverse square)
	FalloffExponent float64

	// Whether walls should occlude this sound
	OcclusionEnabled bool

	// Occlusion factor when blocked (0.0 = silent, 1.0 = no effect)
	OcclusionFactor float64

	// Current volume multiplier (0.0 to 1.0)
	// Calculated by PositionalAudioSystem each frame
	VolumeMultiplier float64

	// Current stereo pan (-1.0 = left, 0.0 = center, 1.0 = right)
	// Calculated by PositionalAudioSystem each frame
	StereoPan float64

	// Whether sound is currently occluded
	IsOccluded bool
}

// Type returns the component type identifier.
func (p *PositionalAudioComponent) Type() string {
	return "positional_audio"
}

// NewPositionalAudioComponent creates a positional audio component with default settings.
// maxDistance: maximum audible distance in pixels (typical: 800 for close sounds, 2000 for ambient)
func NewPositionalAudioComponent(maxDistance float64) *PositionalAudioComponent {
	return &PositionalAudioComponent{
		MaxDistance:      maxDistance,
		FalloffExponent:  2.0, // Inverse square law (realistic)
		OcclusionEnabled: true,
		OcclusionFactor:  0.3, // Muffled to 30% volume when occluded
		VolumeMultiplier: 1.0, // Full volume by default
		StereoPan:        0.0, // Center by default
		IsOccluded:       false,
	}
}

// ReverbComponent adds acoustic reverb to entities or the environment.
// Simulates room acoustics based on size and material properties.
// Phase 14.4: Audio System Enhancement
type ReverbComponent struct {
	// Room size category affects reverb time
	RoomSize RoomSize

	// Room material affects absorption
	Material RoomMaterial

	// Reverb amount (0.0 = dry, 1.0 = full reverb)
	Amount float64

	// Pre-delay in seconds (time before reverb starts)
	PreDelay float64

	// Reverb decay time in seconds
	DecayTime float64

	// High-frequency damping factor (0.0 = no damping, 1.0 = full damping)
	Damping float64

	// Whether reverb is currently active
	Enabled bool
}

// Type returns the component type identifier.
func (r *ReverbComponent) Type() string {
	return "reverb"
}

// RoomSize categories for reverb calculation.
type RoomSize int

const (
	RoomSizeSmall  RoomSize = iota // Closet, corridor (5x5 tiles)
	RoomSizeMedium                 // Normal room (10x10 tiles)
	RoomSizeLarge                  // Hall, chamber (20x20 tiles)
	RoomSizeHuge                   // Cathedral, cavern (40x40 tiles)
)

// String returns human-readable room size name.
func (rs RoomSize) String() string {
	switch rs {
	case RoomSizeSmall:
		return "Small"
	case RoomSizeMedium:
		return "Medium"
	case RoomSizeLarge:
		return "Large"
	case RoomSizeHuge:
		return "Huge"
	default:
		return "Unknown"
	}
}

// RoomMaterial types for acoustic absorption.
type RoomMaterial int

const (
	RoomMaterialStone  RoomMaterial = iota // Hard, reflective (dungeons, castles)
	RoomMaterialWood                       // Medium absorption (houses, taverns)
	RoomMaterialCloth                      // High absorption (tents, libraries)
	RoomMaterialMetal                      // Very reflective (sci-fi, labs)
	RoomMaterialCarpet                     // Very high absorption (luxury areas)
)

// String returns human-readable material name.
func (rm RoomMaterial) String() string {
	switch rm {
	case RoomMaterialStone:
		return "Stone"
	case RoomMaterialWood:
		return "Wood"
	case RoomMaterialCloth:
		return "Cloth"
	case RoomMaterialMetal:
		return "Metal"
	case RoomMaterialCarpet:
		return "Carpet"
	default:
		return "Unknown"
	}
}

// NewReverbComponent creates a reverb component with defaults for the given room size and material.
func NewReverbComponent(roomSize RoomSize, material RoomMaterial) *ReverbComponent {
	component := &ReverbComponent{
		RoomSize: roomSize,
		Material: material,
		Amount:   0.5, // 50% wet/dry mix
		Enabled:  true,
	}

	// Calculate reverb parameters based on room size
	switch roomSize {
	case RoomSizeSmall:
		component.PreDelay = 0.01 // 10ms
		component.DecayTime = 0.3 // 300ms
	case RoomSizeMedium:
		component.PreDelay = 0.02 // 20ms
		component.DecayTime = 0.8 // 800ms
	case RoomSizeLarge:
		component.PreDelay = 0.04 // 40ms
		component.DecayTime = 1.5 // 1.5 seconds
	case RoomSizeHuge:
		component.PreDelay = 0.08 // 80ms
		component.DecayTime = 3.0 // 3 seconds
	}

	// Calculate damping based on material
	switch material {
	case RoomMaterialStone:
		component.Damping = 0.1 // Very little damping (bright reverb)
	case RoomMaterialWood:
		component.Damping = 0.3 // Moderate damping
	case RoomMaterialCloth:
		component.Damping = 0.6 // High damping (dark reverb)
	case RoomMaterialMetal:
		component.Damping = 0.05 // Minimal damping (very bright)
	case RoomMaterialCarpet:
		component.Damping = 0.8 // Maximum damping (very dark)
	}

	return component
}

// CalculateDistance computes the distance between two points.
func CalculateDistance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

// CalculateVolumeFalloff computes volume reduction based on distance.
// Returns a multiplier from 0.0 (silent) to 1.0 (full volume).
func CalculateVolumeFalloff(distance, maxDistance, falloffExponent float64) float64 {
	if distance >= maxDistance {
		return 0.0
	}
	if distance <= 0.0 {
		return 1.0
	}

	// Normalized distance (0.0 to 1.0)
	normalizedDist := distance / maxDistance

	// Apply falloff curve
	falloff := 1.0 - math.Pow(normalizedDist, falloffExponent)

	// Clamp to valid range
	if falloff < 0.0 {
		return 0.0
	}
	if falloff > 1.0 {
		return 1.0
	}

	return falloff
}

// CalculateStereoPan computes stereo panning based on relative position.
// Returns a value from -1.0 (left) to 1.0 (right).
func CalculateStereoPan(sourceX, sourceY, listenerX, listenerY float64) float64 {
	// Calculate angle from listener to source
	dx := sourceX - listenerX
	dy := sourceY - listenerY

	if dx == 0.0 && dy == 0.0 {
		return 0.0 // Center pan when at same position
	}

	// Simple panning based on horizontal offset
	// Normalize to screen width (assume 1920 pixels for reference)
	const screenWidth = 1920.0
	pan := dx / (screenWidth * 0.5)

	// Clamp to valid range
	if pan < -1.0 {
		return -1.0
	}
	if pan > 1.0 {
		return 1.0
	}

	return pan
}
