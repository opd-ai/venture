// Package quality provides visual quality tier management for Venture.
// This file contains the AutoAdjuster implementation for automatic
// quality adjustment based on real-time performance monitoring.
//
// Code relocated from: monitor.go during package reorganization.
package quality

import (
	"sync"
)

// AutoAdjuster automatically adjusts quality based on performance monitoring.
type AutoAdjuster struct {
	monitor  *PerformanceMonitor
	config   *Config
	enabled  bool
	onChange func(QualityLevel) // Callback when quality changes
	mu       sync.RWMutex
}

// NewAutoAdjuster creates a new automatic quality adjuster.
func NewAutoAdjuster(initialConfig *Config, targetFPS float64) *AutoAdjuster {
	monitor := NewPerformanceMonitor(targetFPS, 60) // 1 second at 60 FPS
	monitor.SetCurrentQuality(initialConfig.Level)

	return &AutoAdjuster{
		monitor: monitor,
		config:  initialConfig,
		enabled: true,
	}
}

// SetEnabled enables or disables automatic quality adjustment.
func (aa *AutoAdjuster) SetEnabled(enabled bool) {
	aa.mu.Lock()
	defer aa.mu.Unlock()
	aa.enabled = enabled
}

// IsEnabled returns whether auto-adjustment is enabled.
func (aa *AutoAdjuster) IsEnabled() bool {
	aa.mu.RLock()
	defer aa.mu.RUnlock()
	return aa.enabled
}

// SetOnChange sets a callback function to be called when quality changes.
func (aa *AutoAdjuster) SetOnChange(callback func(QualityLevel)) {
	aa.mu.Lock()
	defer aa.mu.Unlock()
	aa.onChange = callback
}

// Update should be called each frame with the frame time in milliseconds.
// Returns true if quality was adjusted.
func (aa *AutoAdjuster) Update(frameTimeMS float64) bool {
	aa.mu.Lock()

	// Record performance
	aa.monitor.RecordFrame(frameTimeMS)

	// Check if adjustment is needed
	if !aa.enabled {
		aa.mu.Unlock()
		return false
	}

	newQuality, shouldChange := aa.monitor.GetRecommendedQuality()
	if shouldChange {
		// Update config to match new quality level
		aa.config.ApplyLevel(newQuality)
		cb := aa.onChange
		aa.mu.Unlock()

		// Invoke callback outside the lock to avoid blocking updates.
		if cb != nil {
			cb(newQuality)
		}

		return true
	}

	aa.mu.Unlock()
	return false
}

// GetConfig returns a copy of the current quality configuration.
func (aa *AutoAdjuster) GetConfig() Config {
	aa.mu.RLock()
	defer aa.mu.RUnlock()
	return *aa.config
}

// GetStats returns current performance statistics.
func (aa *AutoAdjuster) GetStats() PerformanceStats {
	return aa.monitor.GetStats()
}

// SetManualQuality manually sets the quality level (disables auto-adjustment temporarily).
func (aa *AutoAdjuster) SetManualQuality(level QualityLevel) {
	aa.mu.Lock()
	defer aa.mu.Unlock()

	aa.config.ApplyLevel(level)
	aa.monitor.SetCurrentQuality(level)

	// Call callback if set
	if aa.onChange != nil {
		aa.onChange(level)
	}
}
