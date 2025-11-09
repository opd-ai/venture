package engine

import (
	"sync"

	"github.com/opd-ai/venture/pkg/rendering/quality"
)

// QualitySystem manages quality settings and performance monitoring for the game.
// It handles automatic quality adjustment based on frame rate and provides
// access to quality configuration for rendering systems.
type QualitySystem struct {
	mu sync.RWMutex

	// config is the current quality configuration
	config *quality.Config

	// adjuster handles automatic quality adjustment
	adjuster *quality.AutoAdjuster

	// enabled controls whether the system is active
	enabled bool

	// autoAdjustEnabled controls automatic quality adjustment
	autoAdjustEnabled bool

	// onQualityChange is called when quality level changes
	onQualityChange func(quality.QualityLevel)
}

// NewQualitySystem creates a new quality system with the given initial configuration.
func NewQualitySystem(initialConfig *quality.Config, targetFPS float64) *QualitySystem {
	if initialConfig == nil {
		cfg := quality.DefaultConfig()
		initialConfig = &cfg
	}

	adjuster := quality.NewAutoAdjuster(initialConfig, targetFPS)

	qs := &QualitySystem{
		config:            initialConfig,
		adjuster:          adjuster,
		enabled:           true,
		autoAdjustEnabled: true,
	}

	// Set callback for quality changes
	adjuster.SetOnChange(func(level quality.QualityLevel) {
		qs.mu.RLock()
		callback := qs.onQualityChange
		qs.mu.RUnlock()

		if callback != nil {
			callback(level)
		}
	})

	return qs
}

// Update should be called each frame to record performance and potentially adjust quality.
// deltaTime is the frame time in seconds.
func (qs *QualitySystem) Update(deltaTime float64) {
	qs.mu.RLock()
	enabled := qs.enabled
	autoEnabled := qs.autoAdjustEnabled
	qs.mu.RUnlock()

	if !enabled || !autoEnabled {
		return
	}

	// Convert to milliseconds for the adjuster
	frameTimeMS := deltaTime * 1000.0
	qs.adjuster.Update(frameTimeMS)
}

// GetConfig returns a copy of the current quality configuration.
func (qs *QualitySystem) GetConfig() quality.Config {
	qs.mu.RLock()
	defer qs.mu.RUnlock()
	return *qs.config
}

// SetConfig updates the quality configuration.
func (qs *QualitySystem) SetConfig(config *quality.Config) error {
	if err := config.Validate(); err != nil {
		return err
	}

	qs.mu.Lock()
	qs.config = config
	qs.mu.Unlock()

	// Update adjuster's internal config
	qs.adjuster.SetManualQuality(config.Level)

	return nil
}

// SetQualityLevel changes the quality level and updates the configuration.
func (qs *QualitySystem) SetQualityLevel(level quality.QualityLevel) {
	qs.adjuster.SetManualQuality(level)

	qs.mu.Lock()
	qs.config.ApplyLevel(level)
	qs.mu.Unlock()
}

// GetQualityLevel returns the current quality level.
func (qs *QualitySystem) GetQualityLevel() quality.QualityLevel {
	qs.mu.RLock()
	defer qs.mu.RUnlock()
	return qs.config.Level
}

// Enable enables the quality system.
func (qs *QualitySystem) Enable() {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	qs.enabled = true
}

// Disable disables the quality system.
func (qs *QualitySystem) Disable() {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	qs.enabled = false
}

// IsEnabled returns whether the quality system is enabled.
func (qs *QualitySystem) IsEnabled() bool {
	qs.mu.RLock()
	defer qs.mu.RUnlock()
	return qs.enabled
}

// EnableAutoAdjust enables automatic quality adjustment based on performance.
func (qs *QualitySystem) EnableAutoAdjust() {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	qs.autoAdjustEnabled = true
	qs.adjuster.SetEnabled(true)
}

// DisableAutoAdjust disables automatic quality adjustment.
func (qs *QualitySystem) DisableAutoAdjust() {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	qs.autoAdjustEnabled = false
	qs.adjuster.SetEnabled(false)
}

// IsAutoAdjustEnabled returns whether automatic quality adjustment is enabled.
func (qs *QualitySystem) IsAutoAdjustEnabled() bool {
	qs.mu.RLock()
	defer qs.mu.RUnlock()
	return qs.autoAdjustEnabled
}

// SetOnQualityChange sets a callback to be called when quality level changes.
func (qs *QualitySystem) SetOnQualityChange(callback func(quality.QualityLevel)) {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	qs.onQualityChange = callback
}

// GetStats returns current performance statistics.
func (qs *QualitySystem) GetStats() quality.PerformanceStats {
	return qs.adjuster.GetStats()
}

// GetEntityQualityOverride checks if an entity has quality settings overrides.
// Returns the override component and true if found, otherwise returns default and false.
func GetEntityQualityOverride(entity *Entity) (quality.QualitySettingsComponent, bool) {
	comp, exists := entity.GetComponent("quality_settings")
	if exists && comp != nil {
		if qc, ok := comp.(quality.QualitySettingsComponent); ok {
			return qc, true
		}
	}
	return quality.NewQualitySettingsComponent(), false
}

// ApplyQualityToSpriteDetail returns the effective sprite detail level for an entity,
// considering both global quality settings and entity-specific overrides.
func ApplyQualityToSpriteDetail(config *quality.Config, entity *Entity) float64 {
	if override, hasOverride := GetEntityQualityOverride(entity); hasOverride {
		if override.Override {
			return override.SpriteDetailOverride
		}
	}
	return config.SpriteDetailLevel
}

// ApplyQualityToParticleCount returns the effective particle count multiplier for an entity,
// considering both global quality settings and entity-specific overrides.
func ApplyQualityToParticleCount(config *quality.Config, entity *Entity) float64 {
	if override, hasOverride := GetEntityQualityOverride(entity); hasOverride {
		if override.Override {
			return override.ParticleCountMultiplierOverride
		}
	}
	return config.ParticleCountMultiplier
}

// ShouldRenderEffects returns whether visual effects should be rendered for an entity,
// considering quality settings and entity-specific overrides.
func ShouldRenderEffects(config *quality.Config, entity *Entity) bool {
	if override, hasOverride := GetEntityQualityOverride(entity); hasOverride {
		if override.DisableEffects {
			return false
		}
	}
	// Check global settings
	return config.EnablePostProcessing || config.EnableBloom || config.EnableParticlePhysics
}
