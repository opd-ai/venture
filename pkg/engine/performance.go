package engine

import (
	"fmt"
	"sync"
	"time"
)

// PerformanceMetrics tracks performance statistics for the game.
type PerformanceMetrics struct {
	mu sync.RWMutex

	// Frame timing
	FPS              float64
	FrameTime        time.Duration // Current frame time
	AverageFrameTime time.Duration
	MinFrameTime     time.Duration
	MaxFrameTime     time.Duration

	// Update timing
	UpdateTime        time.Duration
	AverageUpdateTime time.Duration

	// System timing (per system)
	SystemTimes map[string]time.Duration

	// Entity stats
	EntityCount       int
	ActiveEntityCount int

	// Memory stats (sampled)
	MemoryAllocated uint64
	MemoryInUse     uint64

	// Frame counter
	FrameCount uint64

	// Timing history for averaging
	frameTimeHistory  []time.Duration
	updateTimeHistory []time.Duration
	historySize       int
}

// NewPerformanceMetrics creates a new performance metrics tracker.
func NewPerformanceMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		SystemTimes:       make(map[string]time.Duration),
		frameTimeHistory:  make([]time.Duration, 0, 60),
		updateTimeHistory: make([]time.Duration, 0, 60),
		historySize:       60,        // Track last 60 frames
		MinFrameTime:      time.Hour, // Start with max value
	}
}

// RecordFrame records timing for a complete frame.
func (pm *PerformanceMetrics) RecordFrame(frameTime time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.FrameCount++
	pm.FrameTime = frameTime

	// Update min/max
	if frameTime < pm.MinFrameTime {
		pm.MinFrameTime = frameTime
	}
	if frameTime > pm.MaxFrameTime {
		pm.MaxFrameTime = frameTime
	}

	// Add to history
	pm.frameTimeHistory = append(pm.frameTimeHistory, frameTime)
	if len(pm.frameTimeHistory) > pm.historySize {
		pm.frameTimeHistory = pm.frameTimeHistory[1:]
	}

	// Calculate average
	var sum time.Duration
	for _, t := range pm.frameTimeHistory {
		sum += t
	}
	if len(pm.frameTimeHistory) > 0 {
		pm.AverageFrameTime = sum / time.Duration(len(pm.frameTimeHistory))
	}

	// Calculate FPS
	if pm.AverageFrameTime > 0 {
		pm.FPS = float64(time.Second) / float64(pm.AverageFrameTime)
	}
}

// RecordUpdate records timing for the update phase.
func (pm *PerformanceMetrics) RecordUpdate(updateTime time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.UpdateTime = updateTime

	// Add to history
	pm.updateTimeHistory = append(pm.updateTimeHistory, updateTime)
	if len(pm.updateTimeHistory) > pm.historySize {
		pm.updateTimeHistory = pm.updateTimeHistory[1:]
	}

	// Calculate average
	var sum time.Duration
	for _, t := range pm.updateTimeHistory {
		sum += t
	}
	if len(pm.updateTimeHistory) > 0 {
		pm.AverageUpdateTime = sum / time.Duration(len(pm.updateTimeHistory))
	}
}

// RecordSystemTime records timing for a specific system.
func (pm *PerformanceMetrics) RecordSystemTime(systemName string, duration time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.SystemTimes[systemName] = duration
}

// UpdateEntityCount updates the entity count statistics.
func (pm *PerformanceMetrics) UpdateEntityCount(total, active int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.EntityCount = total
	pm.ActiveEntityCount = active
}

// UpdateMemoryStats updates memory statistics.
func (pm *PerformanceMetrics) UpdateMemoryStats(allocated, inUse uint64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.MemoryAllocated = allocated
	pm.MemoryInUse = inUse
}

// GetSnapshot returns a snapshot of current metrics (thread-safe).
func (pm *PerformanceMetrics) GetSnapshot() PerformanceMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Create a copy
	snapshot := PerformanceMetrics{
		FPS:               pm.FPS,
		FrameTime:         pm.FrameTime,
		AverageFrameTime:  pm.AverageFrameTime,
		MinFrameTime:      pm.MinFrameTime,
		MaxFrameTime:      pm.MaxFrameTime,
		UpdateTime:        pm.UpdateTime,
		AverageUpdateTime: pm.AverageUpdateTime,
		EntityCount:       pm.EntityCount,
		ActiveEntityCount: pm.ActiveEntityCount,
		MemoryAllocated:   pm.MemoryAllocated,
		MemoryInUse:       pm.MemoryInUse,
		FrameCount:        pm.FrameCount,
		SystemTimes:       make(map[string]time.Duration),
	}

	// Copy system times
	for k, v := range pm.SystemTimes {
		snapshot.SystemTimes[k] = v
	}

	return snapshot
}

// Reset resets all metrics.
func (pm *PerformanceMetrics) Reset() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.FPS = 0
	pm.FrameTime = 0
	pm.AverageFrameTime = 0
	pm.MinFrameTime = time.Hour
	pm.MaxFrameTime = 0
	pm.UpdateTime = 0
	pm.AverageUpdateTime = 0
	pm.SystemTimes = make(map[string]time.Duration)
	pm.EntityCount = 0
	pm.ActiveEntityCount = 0
	pm.MemoryAllocated = 0
	pm.MemoryInUse = 0
	pm.FrameCount = 0
	pm.frameTimeHistory = pm.frameTimeHistory[:0]
	pm.updateTimeHistory = pm.updateTimeHistory[:0]
}

// String returns a formatted string representation of the metrics.
func (pm *PerformanceMetrics) String() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return fmt.Sprintf(
		"FPS: %.1f | Frame: %.2fms (avg: %.2fms, min: %.2fms, max: %.2fms) | Update: %.2fms | Entities: %d/%d",
		pm.FPS,
		float64(pm.FrameTime.Microseconds())/1000.0,
		float64(pm.AverageFrameTime.Microseconds())/1000.0,
		float64(pm.MinFrameTime.Microseconds())/1000.0,
		float64(pm.MaxFrameTime.Microseconds())/1000.0,
		float64(pm.UpdateTime.Microseconds())/1000.0,
		pm.ActiveEntityCount,
		pm.EntityCount,
	)
}

// DetailedString returns a detailed formatted string with system breakdowns.
func (pm *PerformanceMetrics) DetailedString() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := pm.String() + "\n"
	result += "System Times:\n"

	for name, duration := range pm.SystemTimes {
		result += fmt.Sprintf("  %s: %.2fms\n", name, float64(duration.Microseconds())/1000.0)
	}

	if pm.MemoryAllocated > 0 {
		result += fmt.Sprintf("Memory: %.2f MB allocated, %.2f MB in use\n",
			float64(pm.MemoryAllocated)/(1024*1024),
			float64(pm.MemoryInUse)/(1024*1024))
	}

	return result
}

// IsPerformanceTarget checks if current performance meets target (60 FPS).
func (pm *PerformanceMetrics) IsPerformanceTarget() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return pm.FPS >= 60.0
}

// GetFrameTimePercent returns what percentage of frame time each system uses.
func (pm *PerformanceMetrics) GetFrameTimePercent() map[string]float64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make(map[string]float64)
	if pm.FrameTime == 0 {
		return result
	}

	totalFrameNanos := float64(pm.FrameTime.Nanoseconds())

	for name, duration := range pm.SystemTimes {
		percent := (float64(duration.Nanoseconds()) / totalFrameNanos) * 100.0
		result[name] = percent
	}

	return result
}

// PerformanceMonitor wraps World with performance tracking.
type PerformanceMonitor struct {
	world   *World
	metrics *PerformanceMetrics
	enabled bool
}

// NewPerformanceMonitor creates a new performance monitor for a world.
func NewPerformanceMonitor(world *World) *PerformanceMonitor {
	return &PerformanceMonitor{
		world:   world,
		metrics: NewPerformanceMetrics(),
		enabled: true,
	}
}

// Enable enables performance monitoring.
func (pm *PerformanceMonitor) Enable() {
	pm.enabled = true
}

// Disable disables performance monitoring.
func (pm *PerformanceMonitor) Disable() {
	pm.enabled = false
}

// GetMetrics returns the performance metrics.
func (pm *PerformanceMonitor) GetMetrics() *PerformanceMetrics {
	return pm.metrics
}

// Update updates the world with performance tracking.
func (pm *PerformanceMonitor) Update(deltaTime float64) {
	if !pm.enabled {
		pm.world.Update(deltaTime)
		return
	}

	startFrame := time.Now()

	// Track update time
	startUpdate := time.Now()
	pm.world.Update(deltaTime)
	updateDuration := time.Since(startUpdate)

	pm.metrics.RecordUpdate(updateDuration)

	// Update entity counts
	entities := pm.world.GetEntities()
	activeCount := 0
	for _, entity := range entities {
		// Count entities with velocity as "active"
		if entity.HasComponent("velocity") {
			activeCount++
		}
	}
	pm.metrics.UpdateEntityCount(len(entities), activeCount)

	// Record total frame time
	frameDuration := time.Since(startFrame)
	pm.metrics.RecordFrame(frameDuration)
}

// Timer is a helper for timing code sections.
type Timer struct {
	start time.Time
	name  string
}

// NewTimer creates a new timer with the given name.
func NewTimer(name string) *Timer {
	return &Timer{
		start: time.Now(),
		name:  name,
	}
}

// Stop stops the timer and returns the elapsed duration.
func (t *Timer) Stop() time.Duration {
	return time.Since(t.start)
}

// StopAndLog stops the timer and logs the result.
func (t *Timer) StopAndLog() time.Duration {
	elapsed := time.Since(t.start)
	fmt.Printf("[PERF] %s: %.2fms\n", t.name, float64(elapsed.Microseconds())/1000.0)
	return elapsed
}
