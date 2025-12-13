package engine

import (
	"runtime"
	"time"

	"github.com/opd-ai/venture/pkg/engine/performance"
)

// PerformanceMonitoringSystem wraps the performance monitor for ECS integration.
type PerformanceMonitoringSystem struct {
	monitor         *performance.PerformanceMonitor
	lastUpdate      time.Time
	updateInterval  time.Duration
	lastMemStats    runtime.MemStats
	frameTimeBuffer []float64
	bufferIndex     int
}

// NewPerformanceMonitoringSystem creates a new performance monitoring system.
func NewPerformanceMonitoringSystem() *PerformanceMonitoringSystem {
	return &PerformanceMonitoringSystem{
		monitor:         performance.NewPerformanceMonitor(),
		lastUpdate:      time.Now(),
		updateInterval:  time.Second,         // Update stats every second
		frameTimeBuffer: make([]float64, 60), // Track last 60 frames
	}
}

// Update processes performance monitoring.
func (pms *PerformanceMonitoringSystem) Update(entities []*Entity, deltaTime float64) {
	// Track frame time in milliseconds
	frameTimeMs := deltaTime * 1000.0
	pms.frameTimeBuffer[pms.bufferIndex] = frameTimeMs
	pms.bufferIndex = (pms.bufferIndex + 1) % len(pms.frameTimeBuffer)

	// Update monitor immediately with current frame time
	pms.monitor.UpdateFrameTime(frameTimeMs)

	// Periodically update memory and other stats
	now := time.Now()
	if now.Sub(pms.lastUpdate) < pms.updateInterval {
		return
	}
	pms.lastUpdate = now

	// Update memory statistics
	runtime.ReadMemStats(&pms.lastMemStats)
	memStats := &performance.MemoryStats{
		TotalBytes:     pms.lastMemStats.Alloc,
		TotalMB:        pms.lastMemStats.Alloc / (1024 * 1024),
		Allocations:    make(map[string]uint64),
		LargestAlloc:   "heap",
		LargestAllocMB: pms.lastMemStats.HeapAlloc / (1024 * 1024),
	}
	memStats.Allocations["heap"] = pms.lastMemStats.HeapAlloc
	memStats.Allocations["stack"] = pms.lastMemStats.StackInuse
	memStats.Allocations["gc"] = pms.lastMemStats.NextGC

	pms.monitor.UpdateMemoryStats(memStats)
}

// GetMonitor returns the underlying performance monitor.
func (pms *PerformanceMonitoringSystem) GetMonitor() *performance.PerformanceMonitor {
	return pms.monitor
}

// GetFPS returns current frames per second.
func (pms *PerformanceMonitoringSystem) GetFPS() float64 {
	return pms.monitor.GetFPS()
}

// GetFrameTime returns current frame time in milliseconds.
func (pms *PerformanceMonitoringSystem) GetFrameTime() float64 {
	return pms.monitor.GetFrameTime()
}

// GetMemoryUsageMB returns current memory usage in megabytes.
func (pms *PerformanceMonitoringSystem) GetMemoryUsageMB() uint64 {
	stats := pms.monitor.GetMemoryStats()
	return stats.TotalMB
}

// CheckPerformanceTarget returns true if all performance targets are met.
func (pms *PerformanceMonitoringSystem) CheckPerformanceTarget() bool {
	return pms.monitor.CheckPerformanceTarget()
}

// GetMemoryStats returns detailed memory statistics.
func (pms *PerformanceMonitoringSystem) GetMemoryStats() *performance.MemoryStats {
	return pms.monitor.GetMemoryStats()
}
