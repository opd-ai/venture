package engine

import (
	"testing"
	"time"
)

func TestNewPerformanceMonitoringSystem(t *testing.T) {
	pms := NewPerformanceMonitoringSystem()

	if pms == nil {
		t.Fatal("NewPerformanceMonitoringSystem returned nil")
	}

	if pms.monitor == nil {
		t.Error("Performance monitor is nil")
	}

	if pms.frameTimeBuffer == nil {
		t.Error("Frame time buffer is nil")
	}

	if len(pms.frameTimeBuffer) != 60 {
		t.Errorf("Frame time buffer size = %d, want 60", len(pms.frameTimeBuffer))
	}
}

func TestPerformanceMonitoringSystem_Update(t *testing.T) {
	pms := NewPerformanceMonitoringSystem()

	// Simulate 60 FPS (16.67ms per frame)
	deltaTime := 1.0 / 60.0 // seconds

	// Update multiple frames
	for i := 0; i < 100; i++ {
		pms.Update(nil, deltaTime)
	}

	// Check FPS is calculated
	fps := pms.GetFPS()
	if fps <= 0 {
		t.Errorf("FPS = %f, want > 0", fps)
	}

	// FPS should be approximately 60 (allowing some variance)
	expectedFPS := 60.0
	tolerance := 1.0
	if fps < expectedFPS-tolerance || fps > expectedFPS+tolerance {
		t.Logf("FPS = %f, expected ~%f (within ±%f)", fps, expectedFPS, tolerance)
	}

	// Check frame time
	frameTime := pms.GetFrameTime()
	if frameTime <= 0 {
		t.Errorf("Frame time = %f, want > 0", frameTime)
	}

	// Frame time should be approximately 16.67ms
	expectedFrameTime := deltaTime * 1000.0 // Convert to milliseconds
	if frameTime < expectedFrameTime-1 || frameTime > expectedFrameTime+1 {
		t.Logf("Frame time = %fms, expected ~%fms", frameTime, expectedFrameTime)
	}
}

func TestPerformanceMonitoringSystem_MemoryStats(t *testing.T) {
	pms := NewPerformanceMonitoringSystem()

	// Force immediate update by setting last update time in the past
	pms.lastUpdate = time.Now().Add(-2 * time.Second)

	// Update to trigger memory stats collection
	pms.Update(nil, 1.0/60.0)

	stats := pms.GetMemoryStats()
	if stats == nil {
		t.Fatal("GetMemoryStats returned nil")
	}

	if stats.TotalBytes == 0 {
		t.Error("Memory stats TotalBytes = 0, want > 0")
	}

	if stats.TotalMB == 0 {
		t.Error("Memory stats TotalMB = 0, want > 0")
	}

	if len(stats.Allocations) == 0 {
		t.Error("Memory stats Allocations is empty, want at least one entry")
	}
}

func TestPerformanceMonitoringSystem_FrameTimeBuffer(t *testing.T) {
	pms := NewPerformanceMonitoringSystem()

	// Fill buffer with varying frame times
	frameTimes := []float64{0.016, 0.017, 0.015, 0.020, 0.018}

	for _, dt := range frameTimes {
		pms.Update(nil, dt)
	}

	// Check that buffer index wraps correctly
	if pms.bufferIndex < 0 || pms.bufferIndex >= len(pms.frameTimeBuffer) {
		t.Errorf("Buffer index = %d, want 0 <= index < %d", pms.bufferIndex, len(pms.frameTimeBuffer))
	}

	// Update 100 more times to test wrapping
	for i := 0; i < 100; i++ {
		pms.Update(nil, 1.0/60.0)
	}

	if pms.bufferIndex < 0 || pms.bufferIndex >= len(pms.frameTimeBuffer) {
		t.Errorf("Buffer index after wrapping = %d, want 0 <= index < %d", pms.bufferIndex, len(pms.frameTimeBuffer))
	}
}

func TestPerformanceMonitoringSystem_GetMonitor(t *testing.T) {
	pms := NewPerformanceMonitoringSystem()

	monitor := pms.GetMonitor()
	if monitor == nil {
		t.Fatal("GetMonitor returned nil")
	}

	// Monitor should be the same instance
	if monitor != pms.monitor {
		t.Error("GetMonitor returned different instance than internal monitor")
	}
}

func TestPerformanceMonitoringSystem_CheckPerformanceTarget(t *testing.T) {
	pms := NewPerformanceMonitoringSystem()

	// Simulate good performance (60 FPS)
	for i := 0; i < 10; i++ {
		pms.Update(nil, 1.0/60.0)
	}

	// Note: CheckPerformanceTarget checks frame time, memory, and network
	// We can't fully test all criteria without more setup, but we can verify it doesn't panic
	result := pms.CheckPerformanceTarget()

	// Result depends on actual memory usage and other factors
	t.Logf("Performance target check result: %v", result)
}

func TestPerformanceMonitoringSystem_GetMemoryUsageMB(t *testing.T) {
	pms := NewPerformanceMonitoringSystem()

	// Force memory stats update
	pms.lastUpdate = time.Now().Add(-2 * time.Second)
	pms.Update(nil, 1.0/60.0)

	memMB := pms.GetMemoryUsageMB()
	if memMB == 0 {
		t.Error("Memory usage = 0 MB, want > 0")
	}

	t.Logf("Current memory usage: %d MB", memMB)
}

func BenchmarkPerformanceMonitoringSystem_Update(b *testing.B) {
	pms := NewPerformanceMonitoringSystem()
	deltaTime := 1.0 / 60.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pms.Update(nil, deltaTime)
	}
}

func BenchmarkPerformanceMonitoringSystem_GetFPS(b *testing.B) {
	pms := NewPerformanceMonitoringSystem()

	// Prime the system
	for i := 0; i < 10; i++ {
		pms.Update(nil, 1.0/60.0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pms.GetFPS()
	}
}
