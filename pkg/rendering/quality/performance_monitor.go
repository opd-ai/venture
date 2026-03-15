// Package quality provides visual quality tier management for Venture.
// This file contains the PerformanceMonitor implementation for tracking
// frame rates and recommending quality adjustments.
//
// Code relocated from: monitor.go during package reorganization.
package quality

import (
	"sync"
	"time"
)

// PerformanceMonitor tracks rendering performance metrics to enable
// automatic quality adjustment based on actual frame rates.
type PerformanceMonitor struct {
	mu sync.RWMutex

	// Frame time tracking (in milliseconds)
	frameTimeSamples []float64
	sampleIndex      int
	sampleCount      int
	maxSamples       int

	// Performance thresholds
	targetFPS       float64
	lowThreshold    float64 // FPS below this triggers quality reduction
	highThreshold   float64 // FPS above this allows quality increase
	adjustmentDelay time.Duration

	// State tracking
	lastAdjustment time.Time
	currentQuality QualityLevel
}

// NewPerformanceMonitor creates a new performance monitor.
// targetFPS is the desired frame rate (typically 60).
// sampleSize is the number of frames to average (typically 60-120).
// Note: time.Now() is used for adjustment timing — this is intentional.
// PerformanceMonitor tracks UI/rendering adjustment delays, not procedural generation.
// Frame timing is a real-time concern that correctly uses the system clock.
func NewPerformanceMonitor(targetFPS float64, sampleSize int) *PerformanceMonitor {
	return &PerformanceMonitor{
		frameTimeSamples: make([]float64, sampleSize),
		sampleIndex:      0,
		sampleCount:      0,
		maxSamples:       sampleSize,
		targetFPS:        targetFPS,
		lowThreshold:     targetFPS * 0.92, // 8% below target (e.g., 55 FPS for 60 target)
		highThreshold:    targetFPS * 1.17, // 17% above target (e.g., 70 FPS for 60 target)
		adjustmentDelay:  5 * time.Second,
		lastAdjustment:   time.Now(),
		currentQuality:   QualityHigh, // Start optimistic
	}
}

// RecordFrame records a frame's rendering time in milliseconds.
func (pm *PerformanceMonitor) RecordFrame(frameTimeMS float64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.frameTimeSamples[pm.sampleIndex] = frameTimeMS
	pm.sampleIndex = (pm.sampleIndex + 1) % pm.maxSamples
	if pm.sampleCount < pm.maxSamples {
		pm.sampleCount++
	}
}

// GetAverageFPS returns the average FPS over the sample window.
func (pm *PerformanceMonitor) GetAverageFPS() float64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.sampleCount == 0 {
		return 0
	}

	// Calculate average frame time
	var totalTime float64
	for i := 0; i < pm.sampleCount; i++ {
		totalTime += pm.frameTimeSamples[i]
	}
	avgFrameTime := totalTime / float64(pm.sampleCount)

	// Convert to FPS (avoid division by zero)
	if avgFrameTime < 0.001 {
		return 1000.0 // Cap at 1000 FPS for very fast frames
	}
	return 1000.0 / avgFrameTime
}

// GetRecommendedQuality returns the recommended quality level based on
// current performance. Returns the current quality and a boolean indicating
// if a change is recommended.
func (pm *PerformanceMonitor) GetRecommendedQuality() (QualityLevel, bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Need enough samples for reliable decision
	if pm.sampleCount < pm.maxSamples/2 {
		return pm.currentQuality, false
	}

	// Don't adjust too frequently
	if time.Since(pm.lastAdjustment) < pm.adjustmentDelay {
		return pm.currentQuality, false
	}

	avgFPS := pm.calculateAverageFPS()

	// Determine if quality adjustment is needed
	var newQuality QualityLevel
	shouldChange := false

	if avgFPS < pm.lowThreshold && pm.currentQuality > QualityLow {
		// Performance is poor, reduce quality
		newQuality = pm.currentQuality - 1
		shouldChange = true
	} else if avgFPS > pm.highThreshold && pm.currentQuality < QualityHigh {
		// Performance is good, increase quality
		// Be more conservative about increasing (require sustained high FPS)
		if avgFPS > pm.highThreshold*1.1 { // Extra margin
			newQuality = pm.currentQuality + 1
			shouldChange = true
		}
	}

	if shouldChange {
		pm.lastAdjustment = time.Now()
		pm.currentQuality = newQuality
		return newQuality, true
	}

	return pm.currentQuality, false
}

// SetCurrentQuality updates the monitor's tracked quality level.
func (pm *PerformanceMonitor) SetCurrentQuality(level QualityLevel) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.currentQuality = level
	pm.lastAdjustment = time.Now()
}

// Reset clears all performance samples.
func (pm *PerformanceMonitor) Reset() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.sampleIndex = 0
	pm.sampleCount = 0
	for i := range pm.frameTimeSamples {
		pm.frameTimeSamples[i] = 0
	}
}

// GetStats returns detailed performance statistics.
func (pm *PerformanceMonitor) GetStats() PerformanceStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	stats := PerformanceStats{
		AverageFPS:     pm.calculateAverageFPS(),
		CurrentQuality: pm.currentQuality,
		SampleCount:    pm.sampleCount,
	}

	if pm.sampleCount > 0 {
		// Calculate min and max frame times
		minFrameTime := pm.frameTimeSamples[0]
		maxFrameTime := pm.frameTimeSamples[0]
		for i := 0; i < pm.sampleCount; i++ {
			ft := pm.frameTimeSamples[i]
			if ft < minFrameTime {
				minFrameTime = ft
			}
			if ft > maxFrameTime {
				maxFrameTime = ft
			}
		}
		stats.MinFPS = 1000.0 / maxFrameTime
		stats.MaxFPS = 1000.0 / minFrameTime
	}

	return stats
}

// calculateAverageFPS calculates average FPS without locking (caller must lock).
func (pm *PerformanceMonitor) calculateAverageFPS() float64 {
	if pm.sampleCount == 0 {
		return 0
	}

	var totalTime float64
	for i := 0; i < pm.sampleCount; i++ {
		totalTime += pm.frameTimeSamples[i]
	}
	avgFrameTime := totalTime / float64(pm.sampleCount)

	if avgFrameTime < 0.001 {
		return 1000.0
	}
	return 1000.0 / avgFrameTime
}
