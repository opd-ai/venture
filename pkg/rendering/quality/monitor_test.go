package quality

import (
	"testing"
	"time"
)

func TestNewPerformanceMonitor(t *testing.T) {
	targetFPS := 60.0
	sampleSize := 120

	pm := NewPerformanceMonitor(targetFPS, sampleSize)

	if pm == nil {
		t.Fatal("NewPerformanceMonitor returned nil")
	}

	if pm.targetFPS != targetFPS {
		t.Errorf("targetFPS = %f, want %f", pm.targetFPS, targetFPS)
	}

	if pm.maxSamples != sampleSize {
		t.Errorf("maxSamples = %d, want %d", pm.maxSamples, sampleSize)
	}

	if pm.currentQuality != QualityHigh {
		t.Errorf("currentQuality = %v, want %v", pm.currentQuality, QualityHigh)
	}

	// Should start with empty samples
	if pm.sampleCount != 0 {
		t.Errorf("sampleCount = %d, want 0", pm.sampleCount)
	}
}

func TestPerformanceMonitor_RecordFrame(t *testing.T) {
	pm := NewPerformanceMonitor(60.0, 10)

	// Record some frames
	frameTimes := []float64{16.67, 16.5, 17.0, 16.8, 16.9}

	for i, ft := range frameTimes {
		pm.RecordFrame(ft)
		if pm.sampleCount != i+1 {
			t.Errorf("After %d frames, sampleCount = %d, want %d", i+1, pm.sampleCount, i+1)
		}
	}
}

func TestPerformanceMonitor_GetAverageFPS(t *testing.T) {
	pm := NewPerformanceMonitor(60.0, 10)

	// No samples yet
	if fps := pm.GetAverageFPS(); fps != 0 {
		t.Errorf("GetAverageFPS() with no samples = %f, want 0", fps)
	}

	// Record frames at 60 FPS (16.67ms per frame)
	for i := 0; i < 10; i++ {
		pm.RecordFrame(16.67)
	}

	fps := pm.GetAverageFPS()
	// Should be approximately 60 FPS
	if fps < 59.9 || fps > 60.1 {
		t.Errorf("GetAverageFPS() = %f, want ~60.0", fps)
	}

	// Record frames at 30 FPS (33.33ms per frame)
	pm.Reset()
	for i := 0; i < 10; i++ {
		pm.RecordFrame(33.33)
	}

	fps = pm.GetAverageFPS()
	// Should be approximately 30 FPS
	if fps < 29.9 || fps > 30.1 {
		t.Errorf("GetAverageFPS() = %f, want ~30.0", fps)
	}
}

func TestPerformanceMonitor_GetRecommendedQuality(t *testing.T) {
	pm := NewPerformanceMonitor(60.0, 10)
	pm.adjustmentDelay = 0 // Remove delay for testing

	// Not enough samples yet
	_, shouldChange := pm.GetRecommendedQuality()
	if shouldChange {
		t.Error("Should not recommend change with insufficient samples")
	}

	// Record poor performance (30 FPS - well below threshold)
	for i := 0; i < 10; i++ {
		pm.RecordFrame(33.33)
	}

	quality, shouldChange := pm.GetRecommendedQuality()
	if !shouldChange {
		t.Error("Should recommend quality reduction for poor performance")
	}
	if quality != QualityMedium {
		t.Errorf("Recommended quality = %v, want %v", quality, QualityMedium)
	}

	// Apply the recommendation
	pm.SetCurrentQuality(quality)

	// Record good performance (90 FPS)
	pm.Reset()
	for i := 0; i < 10; i++ {
		pm.RecordFrame(11.11)
	}

	// Need to wait for adjustment delay in real scenario, but we set it to 0
	time.Sleep(1 * time.Millisecond)
	pm.adjustmentDelay = 0

	quality, shouldChange = pm.GetRecommendedQuality()
	if !shouldChange {
		t.Error("Should recommend quality increase for good performance")
	}
	if quality != QualityHigh {
		t.Errorf("Recommended quality = %v, want %v", quality, QualityHigh)
	}
}

func TestPerformanceMonitor_Reset(t *testing.T) {
	pm := NewPerformanceMonitor(60.0, 10)

	// Record some frames
	for i := 0; i < 5; i++ {
		pm.RecordFrame(16.67)
	}

	if pm.sampleCount == 0 {
		t.Error("Should have samples before reset")
	}

	pm.Reset()

	if pm.sampleCount != 0 {
		t.Errorf("After reset, sampleCount = %d, want 0", pm.sampleCount)
	}

	if pm.sampleIndex != 0 {
		t.Errorf("After reset, sampleIndex = %d, want 0", pm.sampleIndex)
	}

	fps := pm.GetAverageFPS()
	if fps != 0 {
		t.Errorf("After reset, GetAverageFPS() = %f, want 0", fps)
	}
}

func TestPerformanceMonitor_GetStats(t *testing.T) {
	pm := NewPerformanceMonitor(60.0, 10)

	// Record varying frame times
	frameTimes := []float64{16.67, 20.0, 15.0, 18.0, 16.0}
	for _, ft := range frameTimes {
		pm.RecordFrame(ft)
	}

	stats := pm.GetStats()

	if stats.SampleCount != len(frameTimes) {
		t.Errorf("SampleCount = %d, want %d", stats.SampleCount, len(frameTimes))
	}

	if stats.AverageFPS <= 0 {
		t.Error("AverageFPS should be positive")
	}

	// Min FPS corresponds to max frame time (20.0ms)
	expectedMinFPS := 1000.0 / 20.0
	if stats.MinFPS < expectedMinFPS-0.1 || stats.MinFPS > expectedMinFPS+0.1 {
		t.Errorf("MinFPS = %f, want ~%f", stats.MinFPS, expectedMinFPS)
	}

	// Max FPS corresponds to min frame time (15.0ms)
	expectedMaxFPS := 1000.0 / 15.0
	if stats.MaxFPS < expectedMaxFPS-0.1 || stats.MaxFPS > expectedMaxFPS+0.1 {
		t.Errorf("MaxFPS = %f, want ~%f", stats.MaxFPS, expectedMaxFPS)
	}

	if stats.CurrentQuality != QualityHigh {
		t.Errorf("CurrentQuality = %v, want %v", stats.CurrentQuality, QualityHigh)
	}
}

func TestNewAutoAdjuster(t *testing.T) {
	config := MediumQualityConfig()
	aa := NewAutoAdjuster(&config, 60.0)

	if aa == nil {
		t.Fatal("NewAutoAdjuster returned nil")
	}

	if !aa.IsEnabled() {
		t.Error("AutoAdjuster should be enabled by default")
	}

	gotConfig := aa.GetConfig()
	if gotConfig.Level != QualityMedium {
		t.Errorf("Initial config level = %v, want %v", gotConfig.Level, QualityMedium)
	}
}

func TestAutoAdjuster_SetEnabled(t *testing.T) {
	config := MediumQualityConfig()
	aa := NewAutoAdjuster(&config, 60.0)

	aa.SetEnabled(false)
	if aa.IsEnabled() {
		t.Error("AutoAdjuster should be disabled after SetEnabled(false)")
	}

	aa.SetEnabled(true)
	if !aa.IsEnabled() {
		t.Error("AutoAdjuster should be enabled after SetEnabled(true)")
	}
}

func TestAutoAdjuster_Update(t *testing.T) {
	config := HighQualityConfig()
	aa := NewAutoAdjuster(&config, 60.0)
	aa.monitor.adjustmentDelay = 0 // Remove delay for testing

	// Record poor performance
	for i := 0; i < 60; i++ {
		changed := aa.Update(33.33) // 30 FPS
		// Should eventually trigger quality reduction
		if changed {
			break
		}
	}

	stats := aa.GetStats()
	if stats.AverageFPS > 35 {
		t.Errorf("Average FPS = %f, expected ~30 FPS", stats.AverageFPS)
	}

	// Quality should have been reduced
	gotConfig := aa.GetConfig()
	if gotConfig.Level == QualityHigh {
		t.Error("Quality should have been reduced from High")
	}
}

func TestAutoAdjuster_SetOnChange(t *testing.T) {
	config := HighQualityConfig()
	aa := NewAutoAdjuster(&config, 60.0)

	callbackCalled := false
	var callbackLevel QualityLevel

	aa.SetOnChange(func(level QualityLevel) {
		callbackCalled = true
		callbackLevel = level
	})

	// Manually trigger quality change
	aa.SetManualQuality(QualityMedium)

	if !callbackCalled {
		t.Error("OnChange callback was not called")
	}

	if callbackLevel != QualityMedium {
		t.Errorf("Callback level = %v, want %v", callbackLevel, QualityMedium)
	}
}

func TestAutoAdjuster_SetManualQuality(t *testing.T) {
	config := HighQualityConfig()
	aa := NewAutoAdjuster(&config, 60.0)

	aa.SetManualQuality(QualityLow)

	gotConfig := aa.GetConfig()
	if gotConfig.Level != QualityLow {
		t.Errorf("Manual quality level = %v, want %v", gotConfig.Level, QualityLow)
	}

	stats := aa.GetStats()
	if stats.CurrentQuality != QualityLow {
		t.Errorf("Monitor quality level = %v, want %v", stats.CurrentQuality, QualityLow)
	}
}

func TestAutoAdjuster_DisabledNoChange(t *testing.T) {
	config := HighQualityConfig()
	aa := NewAutoAdjuster(&config, 60.0)
	aa.monitor.adjustmentDelay = 0

	// Disable auto-adjustment
	aa.SetEnabled(false)

	// Record poor performance
	for i := 0; i < 60; i++ {
		changed := aa.Update(50.0) // ~20 FPS, very poor
		if changed {
			t.Error("Quality should not change when auto-adjustment is disabled")
		}
	}

	// Quality should remain High
	gotConfig := aa.GetConfig()
	if gotConfig.Level != QualityHigh {
		t.Errorf("Quality level = %v, want %v (should not change when disabled)", gotConfig.Level, QualityHigh)
	}
}

func TestPerformanceMonitor_CircularBuffer(t *testing.T) {
	pm := NewPerformanceMonitor(60.0, 5) // Small buffer

	// Record more frames than buffer size
	for i := 0; i < 10; i++ {
		pm.RecordFrame(16.67)
	}

	// Should only keep last 5 samples
	if pm.sampleCount != 5 {
		t.Errorf("sampleCount = %d, want 5", pm.sampleCount)
	}

	// Average should still be based on 5 samples at 60 FPS
	fps := pm.GetAverageFPS()
	if fps < 59.9 || fps > 60.1 {
		t.Errorf("GetAverageFPS() = %f, want ~60.0", fps)
	}
}

func TestPerformanceMonitor_VeryFastFrames(t *testing.T) {
	pm := NewPerformanceMonitor(60.0, 10)

	// Record extremely fast frames (500 FPS)
	for i := 0; i < 10; i++ {
		pm.RecordFrame(2.0)
	}

	fps := pm.GetAverageFPS()
	// Should calculate correctly even for very fast frames
	if fps < 499 || fps > 501 {
		t.Errorf("GetAverageFPS() = %f, want ~500", fps)
	}
}

func TestPerformanceMonitor_AdjustmentDelay(t *testing.T) {
	pm := NewPerformanceMonitor(60.0, 10)
	pm.adjustmentDelay = 100 * time.Millisecond

	// Record poor performance to fill buffer
	for i := 0; i < 10; i++ {
		pm.RecordFrame(33.33)
	}

	// Wait a bit to ensure we're past initialization time
	time.Sleep(10 * time.Millisecond)
	
	// Reset lastAdjustment to now for testing
	pm.mu.Lock()
	pm.lastAdjustment = time.Now().Add(-200 * time.Millisecond) // Set to past
	pm.mu.Unlock()

	// First check should recommend change (delay already elapsed due to backdating)
	_, shouldChange := pm.GetRecommendedQuality()
	if !shouldChange {
		t.Error("First check should recommend change")
	}

	// Immediate second check should not recommend change (delay not elapsed)
	_, shouldChange = pm.GetRecommendedQuality()
	if shouldChange {
		t.Error("Second immediate check should not recommend change due to delay")
	}

	// Wait for delay
	time.Sleep(150 * time.Millisecond)

	// Reset quality back to high to test another adjustment
	pm.mu.Lock()
	pm.currentQuality = QualityHigh
	pm.mu.Unlock()

	// Should allow change now
	_, shouldChange = pm.GetRecommendedQuality()
	if !shouldChange {
		t.Error("Check after delay should recommend change")
	}
}

// Benchmarks

func BenchmarkPerformanceMonitor_RecordFrame(b *testing.B) {
	pm := NewPerformanceMonitor(60.0, 120)
	for i := 0; i < b.N; i++ {
		pm.RecordFrame(16.67)
	}
}

func BenchmarkPerformanceMonitor_GetAverageFPS(b *testing.B) {
	pm := NewPerformanceMonitor(60.0, 120)
	// Fill with samples
	for i := 0; i < 120; i++ {
		pm.RecordFrame(16.67)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pm.GetAverageFPS()
	}
}

func BenchmarkPerformanceMonitor_GetRecommendedQuality(b *testing.B) {
	pm := NewPerformanceMonitor(60.0, 120)
	pm.adjustmentDelay = 0
	// Fill with samples
	for i := 0; i < 120; i++ {
		pm.RecordFrame(16.67)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pm.GetRecommendedQuality()
	}
}

func BenchmarkAutoAdjuster_Update(b *testing.B) {
	config := MediumQualityConfig()
	aa := NewAutoAdjuster(&config, 60.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = aa.Update(16.67)
	}
}
