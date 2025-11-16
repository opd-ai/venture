package cache

import (
	"testing"
	"time"
)

func TestNewMemoryMonitor(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	monitor := NewMemoryMonitor(cache)

	if monitor == nil {
		t.Fatal("NewMemoryMonitor returned nil")
	}

	if monitor.softLimit != 250*1024*1024 {
		t.Errorf("Expected softLimit 250MB, got %d", monitor.softLimit)
	}

	if monitor.hardLimit != 300*1024*1024 {
		t.Errorf("Expected hardLimit 300MB, got %d", monitor.hardLimit)
	}

	if monitor.interval != 5*time.Second {
		t.Errorf("Expected interval 5s, got %v", monitor.interval)
	}
}

func TestSetLimits(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	monitor := NewMemoryMonitor(cache)

	softLimit := int64(100 * 1024 * 1024)
	hardLimit := int64(150 * 1024 * 1024)

	monitor.SetLimits(softLimit, hardLimit)

	if monitor.softLimit != softLimit {
		t.Errorf("softLimit: got %d, want %d", monitor.softLimit, softLimit)
	}

	if monitor.hardLimit != hardLimit {
		t.Errorf("hardLimit: got %d, want %d", monitor.hardLimit, hardLimit)
	}
}

func TestSetInterval(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	monitor := NewMemoryMonitor(cache)

	interval := 10 * time.Second
	monitor.SetInterval(interval)

	if monitor.interval != interval {
		t.Errorf("interval: got %v, want %v", monitor.interval, interval)
	}
}

func TestIsHealthy(t *testing.T) {
	tests := []struct {
		name        string
		currentSize int64
		hardLimit   int64
		wantHealthy bool
	}{
		{
			name:        "well below limit",
			currentSize: 50 * 1024 * 1024,
			hardLimit:   300 * 1024 * 1024,
			wantHealthy: true,
		},
		{
			name:        "at limit",
			currentSize: 300 * 1024 * 1024,
			hardLimit:   300 * 1024 * 1024,
			wantHealthy: true,
		},
		{
			name:        "above limit",
			currentSize: 350 * 1024 * 1024,
			hardLimit:   300 * 1024 * 1024,
			wantHealthy: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewSpriteCache(500 * 1024 * 1024)
			monitor := NewMemoryMonitor(cache)
			monitor.SetLimits(tt.hardLimit*8/10, tt.hardLimit)

			// Manually set current usage
			monitor.mu.Lock()
			monitor.stats.CurrentUsage = tt.currentSize
			monitor.mu.Unlock()

			healthy := monitor.IsHealthy()
			if healthy != tt.wantHealthy {
				t.Errorf("IsHealthy() = %v, want %v", healthy, tt.wantHealthy)
			}
		})
	}
}

func TestUsagePercentage(t *testing.T) {
	tests := []struct {
		name        string
		currentSize int64
		hardLimit   int64
		wantPercent float64
	}{
		{
			name:        "0% usage",
			currentSize: 0,
			hardLimit:   300 * 1024 * 1024,
			wantPercent: 0.0,
		},
		{
			name:        "50% usage",
			currentSize: 150 * 1024 * 1024,
			hardLimit:   300 * 1024 * 1024,
			wantPercent: 50.0,
		},
		{
			name:        "100% usage",
			currentSize: 300 * 1024 * 1024,
			hardLimit:   300 * 1024 * 1024,
			wantPercent: 100.0,
		},
		{
			name:        "over 100% usage",
			currentSize: 400 * 1024 * 1024,
			hardLimit:   300 * 1024 * 1024,
			wantPercent: 400.0 / 3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewSpriteCache(500 * 1024 * 1024)
			monitor := NewMemoryMonitor(cache)
			monitor.SetLimits(tt.hardLimit*8/10, tt.hardLimit)

			monitor.mu.Lock()
			monitor.stats.CurrentUsage = tt.currentSize
			monitor.mu.Unlock()

			pct := monitor.UsagePercentage()
			// Use tolerance for floating point comparison
			tolerance := 0.01
			if pct < tt.wantPercent-tolerance || pct > tt.wantPercent+tolerance {
				t.Errorf("UsagePercentage() = %f, want %f (±%f)", pct, tt.wantPercent, tolerance)
			}
		})
	}
}

func TestEstimatedSpriteCapacity(t *testing.T) {
	cache := NewSpriteCache(500 * 1024 * 1024)
	monitor := NewMemoryMonitor(cache)

	// 64x64 RGBA = 16KB per sprite
	// 300MB / 16KB = 19,200 sprites
	capacity := monitor.EstimatedSpriteCapacity()
	expected := 19200

	if capacity != expected {
		t.Errorf("EstimatedSpriteCapacity() = %d, want %d", capacity, expected)
	}
}

func TestEstimatedSpriteCapacity_CustomLimit(t *testing.T) {
	cache := NewSpriteCache(500 * 1024 * 1024)
	monitor := NewMemoryMonitor(cache)

	// Set 100MB hard limit
	monitor.SetLimits(80*1024*1024, 100*1024*1024)

	// 100MB / 16KB = 6,400 sprites
	capacity := monitor.EstimatedSpriteCapacity()
	expected := 6400

	if capacity != expected {
		t.Errorf("EstimatedSpriteCapacity() = %d, want %d", capacity, expected)
	}
}

func TestStats(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	monitor := NewMemoryMonitor(cache)

	// Set some stats
	monitor.mu.Lock()
	monitor.stats.CurrentUsage = 50 * 1024 * 1024
	monitor.stats.PeakUsage = 75 * 1024 * 1024
	monitor.stats.CleanupCount = 5
	monitor.stats.EvictionCount = 2
	monitor.mu.Unlock()

	stats := monitor.Stats()

	if stats.CurrentUsage != 50*1024*1024 {
		t.Errorf("CurrentUsage: got %d, want %d", stats.CurrentUsage, 50*1024*1024)
	}

	if stats.PeakUsage != 75*1024*1024 {
		t.Errorf("PeakUsage: got %d, want %d", stats.PeakUsage, 75*1024*1024)
	}

	if stats.CleanupCount != 5 {
		t.Errorf("CleanupCount: got %d, want 5", stats.CleanupCount)
	}

	if stats.EvictionCount != 2 {
		t.Errorf("EvictionCount: got %d, want 2", stats.EvictionCount)
	}
}

func TestCheckMemory_BelowLimits(t *testing.T) {
	// Create cache with small size
	cache := NewSpriteCache(500 * 1024 * 1024)
	monitor := NewMemoryMonitor(cache)
	monitor.SetLimits(100*1024*1024, 150*1024*1024)

	// Check memory (should not trigger cleanup)
	monitor.checkMemory()

	stats := monitor.Stats()
	if stats.CleanupCount != 0 {
		t.Errorf("Expected no cleanup, got %d cleanups", stats.CleanupCount)
	}

	if stats.EvictionCount != 0 {
		t.Errorf("Expected no eviction, got %d evictions", stats.EvictionCount)
	}
}

func TestStartStop(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	monitor := NewMemoryMonitor(cache)

	// Set very short interval for testing
	monitor.SetInterval(10 * time.Millisecond)

	// Start monitoring
	monitor.Start()

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Stop monitoring
	monitor.Stop()

	// Should complete without hanging
}

func BenchmarkCheckMemory(b *testing.B) {
	cache := NewSpriteCache(300 * 1024 * 1024)
	monitor := NewMemoryMonitor(cache)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.checkMemory()
	}
}

func BenchmarkUsagePercentage(b *testing.B) {
	cache := NewSpriteCache(300 * 1024 * 1024)
	monitor := NewMemoryMonitor(cache)

	monitor.mu.Lock()
	monitor.stats.CurrentUsage = 150 * 1024 * 1024
	monitor.mu.Unlock()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.UsagePercentage()
	}
}
