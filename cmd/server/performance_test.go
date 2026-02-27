//go:build !android && !ios
// +build !android,!ios

package main

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/sirupsen/logrus"
)

// TestPerformanceSystemIntegration verifies performance monitoring is properly integrated into server
func TestPerformanceSystemIntegration(t *testing.T) {
	tests := []struct {
		name                string
		logLevel            logrus.Level
		expectSystem        bool
		expectMetricsLogged bool
	}{
		{
			name:                "performance system registered",
			logLevel:            logrus.InfoLevel,
			expectSystem:        true,
			expectMetricsLogged: false, // Info level doesn't log debug metrics
		},
		{
			name:                "performance metrics logged at debug level",
			logLevel:            logrus.DebugLevel,
			expectSystem:        true,
			expectMetricsLogged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logging.NewLogger(logging.Config{
				Level:       logging.LogLevel(tt.logLevel.String()),
				Format:      logging.TextFormat,
				AddCaller:   false,
				EnableColor: false,
			})

			world, _ := createGameWorld(logger)

			// Verify performance system is registered
			systems := world.GetSystems()
			var foundPerformanceSystem bool
			for _, system := range systems {
				if _, ok := system.(*engine.PerformanceMonitoringSystem); ok {
					foundPerformanceSystem = true
					break
				}
			}

			if foundPerformanceSystem != tt.expectSystem {
				t.Errorf("expected performance system registered = %v, got %v", tt.expectSystem, foundPerformanceSystem)
			}

			if !foundPerformanceSystem {
				return
			}

			// Run a few update cycles
			for i := 0; i < 5; i++ {
				world.Update(0.016) // ~60 FPS tick
			}

			// Wait 1+ second for memory stats to populate (updateInterval is 1 second)
			time.Sleep(1100 * time.Millisecond)
			world.Update(0.016)

			// Verify performance system is collecting metrics
			for _, system := range systems {
				if perfSys, ok := system.(*engine.PerformanceMonitoringSystem); ok {
					fps := perfSys.GetFPS()
					if fps <= 0 {
						t.Error("expected FPS > 0 after updates")
					}

					frameTime := perfSys.GetFrameTime()
					if frameTime <= 0 {
						t.Error("expected frame time > 0 after updates")
					}

					memUsage := perfSys.GetMemoryUsageMB()
					if memUsage == 0 {
						t.Error("expected memory usage > 0")
					}

					break
				}
			}
		})
	}
}

// TestPerformanceSystemMemoryTracking verifies memory stats are collected
func TestPerformanceSystemMemoryTracking(t *testing.T) {
	logger := logging.NewLogger(logging.Config{
		Level:  logging.InfoLevel,
		Format: logging.TextFormat,
	})

	world, _ := createGameWorld(logger)

	// Extract performance system
	var performanceSystem *engine.PerformanceMonitoringSystem
	for _, system := range world.GetSystems() {
		if perfSys, ok := system.(*engine.PerformanceMonitoringSystem); ok {
			performanceSystem = perfSys
			break
		}
	}

	if performanceSystem == nil {
		t.Fatal("performance system not found in world")
	}

	// Run updates with 1+ second delay to trigger memory stat collection
	world.Update(0.016)
	time.Sleep(1100 * time.Millisecond)
	world.Update(0.016)

	memStats := performanceSystem.GetMemoryStats()
	if memStats == nil {
		t.Fatal("expected memory stats to be non-nil")
	}

	if memStats.TotalMB == 0 {
		t.Error("expected total memory usage > 0")
	}

	if len(memStats.Allocations) == 0 {
		t.Error("expected allocations map to have entries")
	}

	// Verify standard allocation categories exist
	expectedCategories := []string{"heap", "stack", "gc"}
	for _, category := range expectedCategories {
		if _, ok := memStats.Allocations[category]; !ok {
			t.Errorf("expected allocation category %q not found", category)
		}
	}
}

// TestPerformanceSystemTickRateTracking verifies tick rate is calculated correctly
func TestPerformanceSystemTickRateTracking(t *testing.T) {
	logger := logging.NewLogger(logging.Config{
		Level:  logging.InfoLevel,
		Format: logging.TextFormat,
	})

	world, _ := createGameWorld(logger)

	// Extract performance system
	var performanceSystem *engine.PerformanceMonitoringSystem
	for _, system := range world.GetSystems() {
		if perfSys, ok := system.(*engine.PerformanceMonitoringSystem); ok {
			performanceSystem = perfSys
			break
		}
	}

	if performanceSystem == nil {
		t.Fatal("performance system not found in world")
	}

	// Simulate server tick at 30 TPS
	tickDuration := time.Duration(1000000000 / 30) // 30 TPS
	for i := 0; i < 10; i++ {
		world.Update(tickDuration.Seconds())
		time.Sleep(tickDuration)
	}

	fps := performanceSystem.GetFPS()
	if fps < 20 || fps > 40 {
		t.Errorf("expected FPS around 30 for 30 TPS, got %.2f", fps)
	}

	frameTime := performanceSystem.GetFrameTime()
	expectedFrameTime := tickDuration.Seconds() * 1000.0 // Convert to ms
	tolerance := expectedFrameTime * 0.5                 // 50% tolerance for timing variance
	if frameTime < expectedFrameTime-tolerance || frameTime > expectedFrameTime+tolerance {
		t.Errorf("expected frame time around %.2fms for 30 TPS, got %.2fms", expectedFrameTime, frameTime)
	}
}

// TestPerformanceSystemMultipleUpdates verifies system handles many update cycles
func TestPerformanceSystemMultipleUpdates(t *testing.T) {
	logger := logging.NewLogger(logging.Config{
		Level:  logging.InfoLevel,
		Format: logging.TextFormat,
	})

	world, _ := createGameWorld(logger)

	// Run many update cycles to test buffer handling
	for i := 0; i < 200; i++ {
		world.Update(0.016)
	}

	// Extract performance system and verify it's still functional
	for _, system := range world.GetSystems() {
		if perfSys, ok := system.(*engine.PerformanceMonitoringSystem); ok {
			fps := perfSys.GetFPS()
			if fps <= 0 {
				t.Error("expected FPS > 0 after many updates")
			}

			if !perfSys.CheckPerformanceTarget() {
				t.Error("expected performance target to be met")
			}

			return
		}
	}

	t.Fatal("performance system not found after updates")
}

// BenchmarkServerWithPerformanceMonitoring benchmarks server update loop with performance monitoring
func BenchmarkServerWithPerformanceMonitoring(b *testing.B) {
	logger := logging.NewLogger(logging.Config{
		Level:  logging.ErrorLevel, // Minimize logging overhead
		Format: logging.TextFormat,
	})

	world, _ := createGameWorld(logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		world.Update(0.016)
	}
}

// BenchmarkPerformanceSystemExtraction benchmarks extracting performance system from world
func BenchmarkPerformanceSystemExtraction(b *testing.B) {
	logger := logging.NewLogger(logging.Config{
		Level:  logging.ErrorLevel,
		Format: logging.TextFormat,
	})

	world, _ := createGameWorld(logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, system := range world.GetSystems() {
			if _, ok := system.(*engine.PerformanceMonitoringSystem); ok {
				break
			}
		}
	}
}
