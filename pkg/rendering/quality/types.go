// Package quality provides visual quality tier management for Venture.
// This package enables dynamic adjustment of rendering features based on
// performance requirements and hardware capabilities.
package quality

import (
	"fmt"
)

// QualityLevel represents the overall visual quality tier.
type QualityLevel int

const (
	// QualityLow provides minimum visual features for maximum performance.
	// Target: 2x FPS improvement over High quality.
	// Features: Disabled post-processing, 50% particle count, simplified sprites.
	QualityLow QualityLevel = iota

	// QualityMedium provides balanced visuals and performance.
	// Target: Standard performance suitable for most hardware.
	// Features: Key effects enabled, standard particles, enhanced sprites.
	QualityMedium

	// QualityHigh provides maximum visual fidelity.
	// Target: 60 FPS on capable hardware (current baseline: 106 FPS).
	// Features: All effects enabled, maximum particles, highest detail.
	QualityHigh
)

// String returns the string representation of a quality level.
func (q QualityLevel) String() string {
	switch q {
	case QualityLow:
		return "Low"
	case QualityMedium:
		return "Medium"
	case QualityHigh:
		return "High"
	default:
		return "Unknown"
	}
}

// PerformanceStats contains detailed performance metrics.
type PerformanceStats struct {
	AverageFPS     float64
	MinFPS         float64
	MaxFPS         float64
	CurrentQuality QualityLevel
	SampleCount    int
}

// Config contains granular quality settings for all rendering features.
// Each feature can be independently toggled to create custom quality profiles.
type Config struct {
	// Overall quality level (Low, Medium, High)
	Level QualityLevel

	// Post-processing effects
	EnablePostProcessing   bool
	EnableBloom            bool
	EnableAmbientOcclusion bool
	EnableMotionBlur       bool
	EnableDepthBlur        bool
	EnableColorGrading     bool
	EnableVignette         bool
	EnableChromaticAb      bool

	// Lighting effects
	EnableSoftShadows     bool
	EnableColoredLighting bool
	EnableDynamicLighting bool
	ShadowSampleCount     int // 1-5 samples for shadow softness

	// Sprite rendering
	SpriteDetailLevel   float64 // 0.0-1.0 (low to high detail)
	EnableAntiAliasing  bool
	AntiAliasingQuality int // 0=off, 1=2x2, 2=4x4, 3=8x8
	EnableSpriteCache   bool
	EnableEquipmentGlow bool
	EnableDamageStates  bool

	// Tile rendering
	EnableTexturePatterns bool
	EnableTileTransitions bool
	EnableParallaxDepth   bool
	TileLayerCount        int // 1-3 layers
	EnableTileAO          bool
	EnableTileNormals     bool

	// Particle effects
	ParticleCountMultiplier float64 // 0.0-1.0 (low to high particle count)
	EnableParticlePhysics   bool
	EnableWeatherEffects    bool
	EnableAmbienceParticles bool
	ParticleLODDistance     float64 // Distance at which LOD reduces particles

	// UI rendering
	EnableUIDecor       bool
	EnableUITransitions bool
	EnableUIHierarchy   bool
	EnableUIPatterns    bool

	// Environmental effects
	EnableDecorations      bool
	DecorationDensity      float64 // 0.0-1.0
	EnableVisualVariations bool

	// Performance settings
	MaxParticles    int  // Hard cap on particle count
	CacheSizeMB     int  // Sprite cache size limit
	ViewportCulling bool // Enable spatial culling
	BatchRendering  bool // Enable render batching
	ObjectPooling   bool // Enable object pooling
}

// DefaultConfig returns a sensible default quality configuration (Medium quality).
func DefaultConfig() Config {
	return MediumQualityConfig()
}

// LowQualityConfig returns a configuration optimized for performance.
// Disables most visual enhancements for 2x FPS improvement.
func LowQualityConfig() Config {
	return Config{
		Level: QualityLow,

		// Post-processing: All disabled
		EnablePostProcessing:   false,
		EnableBloom:            false,
		EnableAmbientOcclusion: false,
		EnableMotionBlur:       false,
		EnableDepthBlur:        false,
		EnableColorGrading:     true, // Keep for genre consistency
		EnableVignette:         false,
		EnableChromaticAb:      false,

		// Lighting: Simplified
		EnableSoftShadows:     false,
		EnableColoredLighting: true, // Keep for genre mood
		EnableDynamicLighting: true,
		ShadowSampleCount:     1, // Hard shadows only

		// Sprites: Simplified
		SpriteDetailLevel:   0.3,
		EnableAntiAliasing:  false,
		AntiAliasingQuality: 0,
		EnableSpriteCache:   true, // Keep for performance
		EnableEquipmentGlow: false,
		EnableDamageStates:  true, // Keep for gameplay clarity

		// Tiles: Basic
		EnableTexturePatterns: false,
		EnableTileTransitions: true, // Keep for visual continuity
		EnableParallaxDepth:   false,
		TileLayerCount:        1, // Background only
		EnableTileAO:          false,
		EnableTileNormals:     false,

		// Particles: Reduced
		ParticleCountMultiplier: 0.25, // 75% reduction
		EnableParticlePhysics:   false,
		EnableWeatherEffects:    false,
		EnableAmbienceParticles: false,
		ParticleLODDistance:     100.0,

		// UI: Minimal
		EnableUIDecor:       false,
		EnableUITransitions: false,
		EnableUIHierarchy:   true, // Keep for usability
		EnableUIPatterns:    false,

		// Environment: Sparse
		EnableDecorations:      true,
		DecorationDensity:      0.3, // 70% reduction
		EnableVisualVariations: false,

		// Performance: Aggressive optimization
		MaxParticles:    500,
		CacheSizeMB:     50,
		ViewportCulling: true,
		BatchRendering:  true,
		ObjectPooling:   true,
	}
}

// MediumQualityConfig returns a balanced configuration for most hardware.
// Enables key visual features while maintaining good performance.
func MediumQualityConfig() Config {
	return Config{
		Level: QualityMedium,

		// Post-processing: Selective
		EnablePostProcessing:   true,
		EnableBloom:            false, // Expensive
		EnableAmbientOcclusion: true,
		EnableMotionBlur:       false,
		EnableDepthBlur:        false,
		EnableColorGrading:     true,
		EnableVignette:         true,
		EnableChromaticAb:      false,

		// Lighting: Standard
		EnableSoftShadows:     true,
		EnableColoredLighting: true,
		EnableDynamicLighting: true,
		ShadowSampleCount:     3, // Moderate softness

		// Sprites: Enhanced
		SpriteDetailLevel:   0.7,
		EnableAntiAliasing:  true,
		AntiAliasingQuality: 1, // 2x2 sampling
		EnableSpriteCache:   true,
		EnableEquipmentGlow: true,
		EnableDamageStates:  true,

		// Tiles: Standard
		EnableTexturePatterns: true,
		EnableTileTransitions: true,
		EnableParallaxDepth:   true,
		TileLayerCount:        2, // Background + base
		EnableTileAO:          true,
		EnableTileNormals:     false, // Expensive

		// Particles: Standard
		ParticleCountMultiplier: 0.6, // 40% reduction
		EnableParticlePhysics:   true,
		EnableWeatherEffects:    true,
		EnableAmbienceParticles: true,
		ParticleLODDistance:     200.0,

		// UI: Standard
		EnableUIDecor:       true,
		EnableUITransitions: true,
		EnableUIHierarchy:   true,
		EnableUIPatterns:    false, // Expensive

		// Environment: Moderate
		EnableDecorations:      true,
		DecorationDensity:      0.6, // 40% reduction
		EnableVisualVariations: true,

		// Performance: Balanced
		MaxParticles:    2000,
		CacheSizeMB:     100,
		ViewportCulling: true,
		BatchRendering:  true,
		ObjectPooling:   true,
	}
}

// HighQualityConfig returns a configuration for maximum visual fidelity.
// Enables all visual features for capable hardware.
func HighQualityConfig() Config {
	return Config{
		Level: QualityHigh,

		// Post-processing: All enabled
		EnablePostProcessing:   true,
		EnableBloom:            true,
		EnableAmbientOcclusion: true,
		EnableMotionBlur:       false, // Optional, can cause nausea
		EnableDepthBlur:        false, // Optional, gameplay preference
		EnableColorGrading:     true,
		EnableVignette:         true,
		EnableChromaticAb:      true,

		// Lighting: Maximum quality
		EnableSoftShadows:     true,
		EnableColoredLighting: true,
		EnableDynamicLighting: true,
		ShadowSampleCount:     5, // Smoothest shadows

		// Sprites: Maximum detail
		SpriteDetailLevel:   1.0,
		EnableAntiAliasing:  true,
		AntiAliasingQuality: 2, // 4x4 sampling
		EnableSpriteCache:   true,
		EnableEquipmentGlow: true,
		EnableDamageStates:  true,

		// Tiles: Full features
		EnableTexturePatterns: true,
		EnableTileTransitions: true,
		EnableParallaxDepth:   true,
		TileLayerCount:        3, // Background + base + foreground
		EnableTileAO:          true,
		EnableTileNormals:     true,

		// Particles: Maximum
		ParticleCountMultiplier: 1.0, // Full particle count
		EnableParticlePhysics:   true,
		EnableWeatherEffects:    true,
		EnableAmbienceParticles: true,
		ParticleLODDistance:     500.0,

		// UI: Full features
		EnableUIDecor:       true,
		EnableUITransitions: true,
		EnableUIHierarchy:   true,
		EnableUIPatterns:    true,

		// Environment: Rich
		EnableDecorations:      true,
		DecorationDensity:      1.0, // Full density
		EnableVisualVariations: true,

		// Performance: Maximum quality
		MaxParticles:    10000,
		CacheSizeMB:     150,
		ViewportCulling: true,
		BatchRendering:  true,
		ObjectPooling:   true,
	}
}

// Validate checks if the quality configuration is valid.
func (c *Config) Validate() error {
	validators := []func() error{
		c.validateSpriteDetailLevel,
		c.validateAntiAliasingQuality,
		c.validateTileLayerCount,
		c.validateParticleCountMultiplier,
		c.validateDecorationDensity,
		c.validateShadowSampleCount,
		c.validateMaxParticles,
		c.validateCacheSizeMB,
		c.validateParticleLODDistance,
	}

	for _, validator := range validators {
		if err := validator(); err != nil {
			return err
		}
	}

	return nil
}

// validateSpriteDetailLevel checks SpriteDetailLevel is in valid range.
func (c *Config) validateSpriteDetailLevel() error {
	if c.SpriteDetailLevel < 0.0 || c.SpriteDetailLevel > 1.0 {
		return fmt.Errorf("quality: SpriteDetailLevel must be in range [0.0, 1.0], got %f", c.SpriteDetailLevel)
	}
	return nil
}

// validateAntiAliasingQuality checks AntiAliasingQuality is in valid range.
func (c *Config) validateAntiAliasingQuality() error {
	if c.AntiAliasingQuality < 0 || c.AntiAliasingQuality > 3 {
		return fmt.Errorf("quality: AntiAliasingQuality must be in range [0, 3], got %d", c.AntiAliasingQuality)
	}
	return nil
}

// validateTileLayerCount checks TileLayerCount is in valid range.
func (c *Config) validateTileLayerCount() error {
	if c.TileLayerCount < 1 || c.TileLayerCount > 3 {
		return fmt.Errorf("quality: TileLayerCount must be in range [1, 3], got %d", c.TileLayerCount)
	}
	return nil
}

// validateParticleCountMultiplier checks ParticleCountMultiplier is in valid range.
func (c *Config) validateParticleCountMultiplier() error {
	if c.ParticleCountMultiplier < 0.0 || c.ParticleCountMultiplier > 1.0 {
		return fmt.Errorf("quality: ParticleCountMultiplier must be in range [0.0, 1.0], got %f", c.ParticleCountMultiplier)
	}
	return nil
}

// validateDecorationDensity checks DecorationDensity is in valid range.
func (c *Config) validateDecorationDensity() error {
	if c.DecorationDensity < 0.0 || c.DecorationDensity > 1.0 {
		return fmt.Errorf("quality: DecorationDensity must be in range [0.0, 1.0], got %f", c.DecorationDensity)
	}
	return nil
}

// validateShadowSampleCount checks ShadowSampleCount is in valid range.
func (c *Config) validateShadowSampleCount() error {
	if c.ShadowSampleCount < 1 || c.ShadowSampleCount > 5 {
		return fmt.Errorf("quality: ShadowSampleCount must be in range [1, 5], got %d", c.ShadowSampleCount)
	}
	return nil
}

// validateMaxParticles checks MaxParticles is non-negative.
func (c *Config) validateMaxParticles() error {
	if c.MaxParticles < 0 {
		return fmt.Errorf("quality: MaxParticles must be non-negative, got %d", c.MaxParticles)
	}
	return nil
}

// validateCacheSizeMB checks CacheSizeMB is non-negative.
func (c *Config) validateCacheSizeMB() error {
	if c.CacheSizeMB < 0 {
		return fmt.Errorf("quality: CacheSizeMB must be non-negative, got %d", c.CacheSizeMB)
	}
	return nil
}

// validateParticleLODDistance checks ParticleLODDistance is non-negative.
func (c *Config) validateParticleLODDistance() error {
	if c.ParticleLODDistance < 0.0 {
		return fmt.Errorf("quality: ParticleLODDistance must be non-negative, got %f", c.ParticleLODDistance)
	}
	return nil
}

// ApplyLevel updates the config to match a specific quality level.
func (c *Config) ApplyLevel(level QualityLevel) {
	switch level {
	case QualityLow:
		*c = LowQualityConfig()
	case QualityMedium:
		*c = MediumQualityConfig()
	case QualityHigh:
		*c = HighQualityConfig()
	default:
		// Invalid level, apply safe default (Medium)
		*c = MediumQualityConfig()
	}
}
