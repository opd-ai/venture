//go:build !android && !ios
// +build !android,!ios

package main

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/sirupsen/logrus"
)

// TestStartPerformanceMonitoring_VerboseDisabled verifies monitoring is skipped when verbose=false.
func TestStartPerformanceMonitoring_VerboseDisabled(t *testing.T) {
	// Save original verbose flag
	originalVerbose := *verbose
	defer func() { *verbose = originalVerbose }()

	*verbose = false

	game := engine.NewEbitenGame(800, 600)

	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)
	clientLogger := logger.WithField("test", "performance_monitoring")

	// Should not panic or start monitoring when verbose=false
	startPerformanceMonitoring(game, clientLogger)

	// Give goroutines time to start (if any)
	time.Sleep(100 * time.Millisecond)

	// No assertions needed - just verify no panic
}

// TestStartPerformanceMonitoring_VerboseEnabled verifies monitoring starts when verbose=true.
func TestStartPerformanceMonitoring_VerboseEnabled(t *testing.T) {
	// Save original verbose flag
	originalVerbose := *verbose
	defer func() { *verbose = originalVerbose }()

	*verbose = true

	game := engine.NewEbitenGame(800, 600)

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	clientLogger := logger.WithField("test", "performance_monitoring")

	// Should start monitoring goroutines
	startPerformanceMonitoring(game, clientLogger)

	// Give goroutines time to start and perform initial checks
	time.Sleep(200 * time.Millisecond)

	// Verify FPS provider works (should return default when no frames recorded)
	fps := game.CurrentFPS()
	if fps < 0 {
		t.Errorf("Expected non-negative FPS, got %.2f", fps)
	}
}

// TestStartPerformanceMonitoring_FPSProviderIntegration verifies stability monitor gets FPS.
func TestStartPerformanceMonitoring_FPSProviderIntegration(t *testing.T) {
	// Save original verbose flag
	originalVerbose := *verbose
	defer func() { *verbose = originalVerbose }()

	*verbose = true

	game := engine.NewEbitenGame(800, 600)

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	clientLogger := logger.WithField("test", "fps_provider")

	startPerformanceMonitoring(game, clientLogger)

	// Verify the game can provide FPS (implements FPSProvider)
	// Without frames recorded, should return default 60.0 or 0.0
	fps := game.CurrentFPS()
	if fps < 0 {
		t.Errorf("FPS provider returned negative value: %.2f", fps)
	}
}

// TestStartPerformanceMonitoring_MemoryMonitoring verifies memory tracking works.
func TestStartPerformanceMonitoring_MemoryMonitoring(t *testing.T) {
	// Save original verbose flag
	originalVerbose := *verbose
	defer func() { *verbose = originalVerbose }()

	*verbose = true

	game := engine.NewEbitenGame(800, 600)

	// Create some entities to use memory
	for i := 0; i < 100; i++ {
		entity := game.World.CreateEntity()
		entity.AddComponent(&engine.PositionComponent{X: float64(i), Y: float64(i)})
		entity.AddComponent(&engine.VelocityComponent{VX: 1.0, VY: 1.0})
	}

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	clientLogger := logger.WithField("test", "memory_monitoring")

	startPerformanceMonitoring(game, clientLogger)

	// Give monitoring time to perform checks
	time.Sleep(100 * time.Millisecond)

	// Memory monitoring should not panic or fail
	// No specific assertions needed - just verify it runs
}

// TestStartPerformanceMonitoring_DefaultFPS verifies FPS reporting.
func TestStartPerformanceMonitoring_DefaultFPS(t *testing.T) {
	// Save original verbose flag
	originalVerbose := *verbose
	defer func() { *verbose = originalVerbose }()

	*verbose = true

	game := engine.NewEbitenGame(800, 600)

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	clientLogger := logger.WithField("test", "default_fps")

	// Should not panic
	startPerformanceMonitoring(game, clientLogger)

	// Verify CurrentFPS returns a valid value
	fps := game.CurrentFPS()
	if fps < 0 {
		t.Errorf("Expected non-negative FPS, got %.2f", fps)
	}
}

// BenchmarkStartPerformanceMonitoring measures the overhead of starting monitoring.
func BenchmarkStartPerformanceMonitoring(b *testing.B) {
	// Save original verbose flag
	originalVerbose := *verbose
	defer func() { *verbose = originalVerbose }()

	*verbose = true

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // Reduce log noise
	clientLogger := logger.WithField("test", "benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game := engine.NewEbitenGame(800, 600)
		startPerformanceMonitoring(game, clientLogger)
	}
}

// TestBuildPerformanceFields verifies performance field construction.
func TestBuildPerformanceFields(t *testing.T) {
	tests := []struct {
		name      string
		fps       float64
		memBytes  uint64
		wantFPS   string
		wantMemMB string
	}{
		{
			name:      "typical values",
			fps:       60.0,
			memBytes:  100 * 1024 * 1024,
			wantFPS:   "60.0",
			wantMemMB: "100.0",
		},
		{
			name:      "low fps high memory",
			fps:       30.5,
			memBytes:  450 * 1024 * 1024,
			wantFPS:   "30.5",
			wantMemMB: "450.0",
		},
		{
			name:      "zero fps minimal memory",
			fps:       0.0,
			memBytes:  1024 * 1024,
			wantFPS:   "0.0",
			wantMemMB: "1.0",
		},
		{
			name:      "fractional values",
			fps:       59.9,
			memBytes:  256*1024*1024 + 512*1024, // 256.5 MB
			wantFPS:   "59.9",
			wantMemMB: "256.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := buildPerformanceFields(tt.fps, tt.memBytes)

			if fps, ok := fields["fps"].(string); !ok || fps != tt.wantFPS {
				t.Errorf("fps = %v, want %v", fields["fps"], tt.wantFPS)
			}
			if memMB, ok := fields["memory_mb"].(string); !ok || memMB != tt.wantMemMB {
				t.Errorf("memory_mb = %v, want %v", fields["memory_mb"], tt.wantMemMB)
			}
			if _, ok := fields["goroutines"]; !ok {
				t.Error("goroutines field missing")
			}
		})
	}
}

// TestBuildPerformanceFieldsGoroutineCount verifies goroutine counting.
func TestBuildPerformanceFieldsGoroutineCount(t *testing.T) {
	fields := buildPerformanceFields(60.0, 100*1024*1024)

	goroutines, ok := fields["goroutines"].(int)
	if !ok {
		t.Fatalf("goroutines is not int: %T", fields["goroutines"])
	}
	if goroutines < 1 {
		t.Errorf("goroutines = %d, want >= 1", goroutines)
	}
}

// TestLogPerformanceStatusBelowFPSTarget tests logging when FPS is below target.
func TestLogPerformanceStatusBelowFPSTarget(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	clientLogger := logger.WithField("test", "fps_warning")

	fields := buildPerformanceFields(45.0, 100*1024*1024)

	// Should not panic and should log warning for low FPS
	logPerformanceStatus(45.0, 100*1024*1024, fields, clientLogger)
}

// TestLogPerformanceStatusAboveMemoryTarget tests logging when memory exceeds target.
func TestLogPerformanceStatusAboveMemoryTarget(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	clientLogger := logger.WithField("test", "memory_warning")

	highMem := uint64(600 * 1024 * 1024) // 600MB, exceeds 500MB limit
	fields := buildPerformanceFields(60.0, highMem)

	// Should not panic and should log warning for high memory
	logPerformanceStatus(60.0, highMem, fields, clientLogger)
}

// TestLogPerformanceStatusAllTargetsMet tests logging when all targets are met.
func TestLogPerformanceStatusAllTargetsMet(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	clientLogger := logger.WithField("test", "stability_passed")

	normalMem := uint64(200 * 1024 * 1024) // 200MB, within 500MB limit
	fields := buildPerformanceFields(75.0, normalMem)

	// Should not panic and should log debug for passing check
	logPerformanceStatus(75.0, normalMem, fields, clientLogger)
}

// TestLogPerformanceStatusBothTargetsFailed tests logging when both targets fail.
func TestLogPerformanceStatusBothTargetsFailed(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	clientLogger := logger.WithField("test", "both_warnings")

	highMem := uint64(600 * 1024 * 1024)
	fields := buildPerformanceFields(30.0, highMem)

	// Should not panic and should log warnings for both violations
	logPerformanceStatus(30.0, highMem, fields, clientLogger)
}

// BenchmarkBuildPerformanceFields benchmarks field construction.
func BenchmarkBuildPerformanceFields(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildPerformanceFields(60.0, 256*1024*1024)
	}
}
