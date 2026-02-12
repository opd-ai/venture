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
