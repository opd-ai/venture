package engine

import (
	"testing"
	"time"
)

func TestPerformanceMetricsCreation(t *testing.T) {
	pm := NewPerformanceMetrics()
	
	if pm == nil {
		t.Fatal("NewPerformanceMetrics() returned nil")
	}
	
	if pm.SystemTimes == nil {
		t.Error("SystemTimes map should be initialized")
	}
	
	if pm.MinFrameTime != time.Hour {
		t.Error("MinFrameTime should start at max value")
	}
}

func TestRecordFrame(t *testing.T) {
	pm := NewPerformanceMetrics()
	
	frameTime := 16 * time.Millisecond
	pm.RecordFrame(frameTime)
	
	if pm.FrameTime != frameTime {
		t.Errorf("FrameTime = %v, want %v", pm.FrameTime, frameTime)
	}
	
	if pm.FrameCount != 1 {
		t.Errorf("FrameCount = %d, want 1", pm.FrameCount)
	}
	
	if pm.MinFrameTime != frameTime {
		t.Errorf("MinFrameTime = %v, want %v", pm.MinFrameTime, frameTime)
	}
	
	if pm.MaxFrameTime != frameTime {
		t.Errorf("MaxFrameTime = %v, want %v", pm.MaxFrameTime, frameTime)
	}
}

func TestRecordFrameAverage(t *testing.T) {
	pm := NewPerformanceMetrics()
	
	// Record multiple frames
	for i := 0; i < 10; i++ {
		pm.RecordFrame(16 * time.Millisecond)
	}
	
	if pm.AverageFrameTime != 16*time.Millisecond {
		t.Errorf("AverageFrameTime = %v, want %v", pm.AverageFrameTime, 16*time.Millisecond)
	}
	
	// FPS should be approximately 62.5 (1000ms / 16ms)
	expectedFPS := 62.5
	if pm.FPS < expectedFPS-1 || pm.FPS > expectedFPS+1 {
		t.Errorf("FPS = %.2f, want approximately %.2f", pm.FPS, expectedFPS)
	}
}

func TestRecordFrameMinMax(t *testing.T) {
	pm := NewPerformanceMetrics()
	
	pm.RecordFrame(10 * time.Millisecond)
	pm.RecordFrame(20 * time.Millisecond)
	pm.RecordFrame(15 * time.Millisecond)
	
	if pm.MinFrameTime != 10*time.Millisecond {
		t.Errorf("MinFrameTime = %v, want %v", pm.MinFrameTime, 10*time.Millisecond)
	}
	
	if pm.MaxFrameTime != 20*time.Millisecond {
		t.Errorf("MaxFrameTime = %v, want %v", pm.MaxFrameTime, 20*time.Millisecond)
	}
}

func TestRecordUpdate(t *testing.T) {
	pm := NewPerformanceMetrics()
	
	updateTime := 5 * time.Millisecond
	pm.RecordUpdate(updateTime)
	
	if pm.UpdateTime != updateTime {
		t.Errorf("UpdateTime = %v, want %v", pm.UpdateTime, updateTime)
	}
	
	// Record multiple updates
	for i := 0; i < 10; i++ {
		pm.RecordUpdate(5 * time.Millisecond)
	}
	
	if pm.AverageUpdateTime != 5*time.Millisecond {
		t.Errorf("AverageUpdateTime = %v, want %v", pm.AverageUpdateTime, 5*time.Millisecond)
	}
}

func TestRecordSystemTime(t *testing.T) {
	pm := NewPerformanceMetrics()
	
	pm.RecordSystemTime("movement", 2*time.Millisecond)
	pm.RecordSystemTime("collision", 3*time.Millisecond)
	
	if pm.SystemTimes["movement"] != 2*time.Millisecond {
		t.Errorf("SystemTimes[movement] = %v, want %v", pm.SystemTimes["movement"], 2*time.Millisecond)
	}
	
	if pm.SystemTimes["collision"] != 3*time.Millisecond {
		t.Errorf("SystemTimes[collision] = %v, want %v", pm.SystemTimes["collision"], 3*time.Millisecond)
	}
}

func TestUpdateEntityCount(t *testing.T) {
	pm := NewPerformanceMetrics()
	
	pm.UpdateEntityCount(100, 75)
	
	if pm.EntityCount != 100 {
		t.Errorf("EntityCount = %d, want 100", pm.EntityCount)
	}
	
	if pm.ActiveEntityCount != 75 {
		t.Errorf("ActiveEntityCount = %d, want 75", pm.ActiveEntityCount)
	}
}

func TestUpdateMemoryStats(t *testing.T) {
	pm := NewPerformanceMetrics()
	
	pm.UpdateMemoryStats(1024*1024, 512*1024)
	
	if pm.MemoryAllocated != 1024*1024 {
		t.Errorf("MemoryAllocated = %d, want %d", pm.MemoryAllocated, 1024*1024)
	}
	
	if pm.MemoryInUse != 512*1024 {
		t.Errorf("MemoryInUse = %d, want %d", pm.MemoryInUse, 512*1024)
	}
}

func TestGetSnapshot(t *testing.T) {
	pm := NewPerformanceMetrics()
	
	pm.RecordFrame(16 * time.Millisecond)
	pm.RecordUpdate(5 * time.Millisecond)
	pm.RecordSystemTime("test", 2*time.Millisecond)
	pm.UpdateEntityCount(50, 30)
	
	snapshot := pm.GetSnapshot()
	
	if snapshot.FrameTime != 16*time.Millisecond {
		t.Error("Snapshot should contain correct frame time")
	}
	
	if snapshot.EntityCount != 50 {
		t.Error("Snapshot should contain correct entity count")
	}
	
	if snapshot.SystemTimes["test"] != 2*time.Millisecond {
		t.Error("Snapshot should contain correct system times")
	}
	
	// Modify original, snapshot should be unchanged
	pm.UpdateEntityCount(100, 80)
	if snapshot.EntityCount != 50 {
		t.Error("Snapshot should be independent of original")
	}
}

func TestReset(t *testing.T) {
	pm := NewPerformanceMetrics()
	
	pm.RecordFrame(16 * time.Millisecond)
	pm.RecordUpdate(5 * time.Millisecond)
	pm.RecordSystemTime("test", 2*time.Millisecond)
	pm.UpdateEntityCount(50, 30)
	
	pm.Reset()
	
	if pm.FrameTime != 0 {
		t.Error("FrameTime should be reset to 0")
	}
	
	if pm.EntityCount != 0 {
		t.Error("EntityCount should be reset to 0")
	}
	
	if pm.FrameCount != 0 {
		t.Error("FrameCount should be reset to 0")
	}
	
	if len(pm.SystemTimes) != 0 {
		t.Error("SystemTimes should be cleared")
	}
}

func TestString(t *testing.T) {
	pm := NewPerformanceMetrics()
	
	pm.RecordFrame(16 * time.Millisecond)
	pm.UpdateEntityCount(100, 75)
	
	str := pm.String()
	if str == "" {
		t.Error("String() should return non-empty string")
	}
}

func TestDetailedString(t *testing.T) {
	pm := NewPerformanceMetrics()
	
	pm.RecordFrame(16 * time.Millisecond)
	pm.RecordSystemTime("movement", 2*time.Millisecond)
	pm.RecordSystemTime("collision", 3*time.Millisecond)
	pm.UpdateMemoryStats(1024*1024, 512*1024)
	
	str := pm.DetailedString()
	if str == "" {
		t.Error("DetailedString() should return non-empty string")
	}
}

func TestIsPerformanceTarget(t *testing.T) {
	pm := NewPerformanceMetrics()
	
	// Simulate 60+ FPS
	for i := 0; i < 10; i++ {
		pm.RecordFrame(16 * time.Millisecond)
	}
	
	if !pm.IsPerformanceTarget() {
		t.Error("Should meet performance target at 62.5 FPS")
	}
	
	// Simulate low FPS
	pm.Reset()
	for i := 0; i < 10; i++ {
		pm.RecordFrame(30 * time.Millisecond) // ~33 FPS
	}
	
	if pm.IsPerformanceTarget() {
		t.Error("Should not meet performance target at 33 FPS")
	}
}

func TestGetFrameTimePercent(t *testing.T) {
	pm := NewPerformanceMetrics()
	
	pm.RecordFrame(10 * time.Millisecond)
	pm.RecordSystemTime("movement", 2*time.Millisecond)
	pm.RecordSystemTime("collision", 3*time.Millisecond)
	
	percentages := pm.GetFrameTimePercent()
	
	if percentages["movement"] != 20.0 {
		t.Errorf("movement percentage = %.2f, want 20.0", percentages["movement"])
	}
	
	if percentages["collision"] != 30.0 {
		t.Errorf("collision percentage = %.2f, want 30.0", percentages["collision"])
	}
}

func TestGetFrameTimePercentZeroFrame(t *testing.T) {
	pm := NewPerformanceMetrics()
	
	pm.RecordSystemTime("test", 5*time.Millisecond)
	
	percentages := pm.GetFrameTimePercent()
	
	// Should return empty map when frame time is 0
	if len(percentages) != 0 {
		t.Error("Should return empty map when frame time is 0")
	}
}

func TestPerformanceMonitorCreation(t *testing.T) {
	world := NewWorld()
	monitor := NewPerformanceMonitor(world)
	
	if monitor == nil {
		t.Fatal("NewPerformanceMonitor() returned nil")
	}
	
	if monitor.metrics == nil {
		t.Error("Metrics should be initialized")
	}
	
	if !monitor.enabled {
		t.Error("Monitor should be enabled by default")
	}
}

func TestPerformanceMonitorEnableDisable(t *testing.T) {
	world := NewWorld()
	monitor := NewPerformanceMonitor(world)
	
	monitor.Disable()
	if monitor.enabled {
		t.Error("Monitor should be disabled")
	}
	
	monitor.Enable()
	if !monitor.enabled {
		t.Error("Monitor should be enabled")
	}
}

func TestPerformanceMonitorUpdate(t *testing.T) {
	world := NewWorld()
	monitor := NewPerformanceMonitor(world)
	
	// Add some entities
	for i := 0; i < 10; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i), Y: float64(i)})
		entity.AddComponent(&VelocityComponent{VX: 1, VY: 1})
	}
	
	// Process additions
	world.Update(0)
	
	// Update with monitoring
	monitor.Update(0.016)
	
	metrics := monitor.GetMetrics()
	
	if metrics.FrameCount != 1 {
		t.Errorf("FrameCount = %d, want 1", metrics.FrameCount)
	}
	
	if metrics.EntityCount != 10 {
		t.Errorf("EntityCount = %d, want 10", metrics.EntityCount)
	}
	
	if metrics.ActiveEntityCount != 10 {
		t.Errorf("ActiveEntityCount = %d, want 10 (entities with velocity)", metrics.ActiveEntityCount)
	}
}

func TestPerformanceMonitorDisabled(t *testing.T) {
	world := NewWorld()
	monitor := NewPerformanceMonitor(world)
	monitor.Disable()
	
	// Update should work but not record metrics
	monitor.Update(0.016)
	
	metrics := monitor.GetMetrics()
	
	if metrics.FrameCount != 0 {
		t.Error("Metrics should not be recorded when disabled")
	}
}

func TestTimer(t *testing.T) {
	timer := NewTimer("test")
	
	if timer == nil {
		t.Fatal("NewTimer() returned nil")
	}
	
	if timer.name != "test" {
		t.Errorf("Timer name = %s, want 'test'", timer.name)
	}
	
	// Simulate some work
	time.Sleep(10 * time.Millisecond)
	
	elapsed := timer.Stop()
	
	if elapsed < 10*time.Millisecond {
		t.Errorf("Elapsed time = %v, want at least 10ms", elapsed)
	}
}

func TestTimerStopAndLog(t *testing.T) {
	timer := NewTimer("test")
	
	// Simulate some work
	time.Sleep(10 * time.Millisecond)
	
	elapsed := timer.StopAndLog()
	
	if elapsed < 10*time.Millisecond {
		t.Errorf("Elapsed time = %v, want at least 10ms", elapsed)
	}
}

// Benchmarks

func BenchmarkRecordFrame(b *testing.B) {
	pm := NewPerformanceMetrics()
	frameTime := 16 * time.Millisecond
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pm.RecordFrame(frameTime)
	}
}

func BenchmarkGetSnapshot(b *testing.B) {
	pm := NewPerformanceMetrics()
	pm.RecordFrame(16 * time.Millisecond)
	pm.RecordSystemTime("test1", 2*time.Millisecond)
	pm.RecordSystemTime("test2", 3*time.Millisecond)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pm.GetSnapshot()
	}
}

func BenchmarkPerformanceMonitorUpdate(b *testing.B) {
	world := NewWorld()
	monitor := NewPerformanceMonitor(world)
	
	// Add some entities
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i), Y: float64(i)})
	}
	world.Update(0)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.Update(0.016)
	}
}
