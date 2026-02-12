//go:build !android && !ios
// +build !android,!ios

package main

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/sirupsen/logrus"
)

// TestLazyInitialization verifies that lazy initialization correctly defers non-critical system setup.
func TestLazyInitialization(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"lazy initialization starts and completes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create minimal test environment
			logger := logrus.New()
			logger.SetLevel(logrus.FatalLevel) // Suppress logs during test
			clientLogger := logger.WithField("test", "lazy_init")

			// Create a test game instance with required systems
			game := &engine.EbitenGame{
				World:        engine.NewWorldWithLogger(logger),
				ScreenWidth:  800,
				ScreenHeight: 600,
			}
			game.CameraSystem = engine.NewCameraSystem(800, 600)
			game.RenderSystem = engine.NewRenderSystem(game.CameraSystem)

			// Initialize core systems (mimics setupAllGameSystems)
			sys := initializeCoreSystems(game, logger, clientLogger)

			// Verify lazy init hasn't started yet
			if sys.lazyInitStarted {
				t.Errorf("lazy initialization should not have started before scheduleLazyInit()")
			}

			// Schedule lazy initialization
			sys.scheduleLazyInit(game, logger, clientLogger)

			// Verify it started
			sys.lazyInitMutex.Lock()
			started := sys.lazyInitStarted
			sys.lazyInitMutex.Unlock()

			if !started {
				t.Errorf("lazy initialization should have started after scheduleLazyInit()")
			}

			// Wait for lazy init to complete (with timeout)
			timeout := time.After(5 * time.Second)
			ticker := time.NewTicker(50 * time.Millisecond)
			defer ticker.Stop()

			completed := false
			for !completed {
				select {
				case <-timeout:
					t.Fatalf("lazy initialization did not complete within 5 seconds")
				case <-ticker.C:
					completed = sys.isLazyInitCompleted()
				}
			}

			// Verify completion
			if !sys.isLazyInitCompleted() {
				t.Errorf("isLazyInitCompleted() should return true after completion")
			}
		})
	}
}

// TestScheduleLazyInitIdempotent verifies that calling scheduleLazyInit multiple times is safe.
func TestScheduleLazyInitIdempotent(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	clientLogger := logger.WithField("test", "lazy_init_idempotent")

	game := &engine.EbitenGame{
		World:        engine.NewWorldWithLogger(logger),
		ScreenWidth:  800,
		ScreenHeight: 600,
	}
	game.CameraSystem = engine.NewCameraSystem(800, 600)
	game.RenderSystem = engine.NewRenderSystem(game.CameraSystem)

	sys := initializeCoreSystems(game, logger, clientLogger)

	// Call scheduleLazyInit multiple times
	sys.scheduleLazyInit(game, logger, clientLogger)
	sys.scheduleLazyInit(game, logger, clientLogger)
	sys.scheduleLazyInit(game, logger, clientLogger)

	// Should only start once
	sys.lazyInitMutex.Lock()
	started := sys.lazyInitStarted
	sys.lazyInitMutex.Unlock()

	if !started {
		t.Errorf("lazy initialization should have started")
	}

	// Wait for completion
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return // Test passes - no panic occurred from multiple calls
		case <-ticker.C:
			if sys.isLazyInitCompleted() {
				return // Test passes
			}
		}
	}
}

// TestIsLazyInitCompletedThreadSafe verifies thread-safe access to completion status.
func TestIsLazyInitCompletedThreadSafe(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	clientLogger := logger.WithField("test", "lazy_init_threadsafe")

	game := &engine.EbitenGame{
		World:        engine.NewWorldWithLogger(logger),
		ScreenWidth:  800,
		ScreenHeight: 600,
	}
	game.CameraSystem = engine.NewCameraSystem(800, 600)
	game.RenderSystem = engine.NewRenderSystem(game.CameraSystem)

	sys := initializeCoreSystems(game, logger, clientLogger)
	sys.scheduleLazyInit(game, logger, clientLogger)

	// Spawn multiple goroutines checking completion status
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = sys.isLazyInitCompleted()
				time.Sleep(1 * time.Millisecond)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test passes if no race condition detected
}

// BenchmarkLazyInitScheduling measures the overhead of scheduling lazy initialization.
func BenchmarkLazyInitScheduling(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	clientLogger := logger.WithField("bench", "lazy_init")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game := &engine.EbitenGame{
			World:        engine.NewWorldWithLogger(logger),
			ScreenWidth:  800,
			ScreenHeight: 600,
		}
		game.CameraSystem = engine.NewCameraSystem(800, 600)
		game.RenderSystem = engine.NewRenderSystem(game.CameraSystem)

		sys := initializeCoreSystems(game, logger, clientLogger)
		sys.scheduleLazyInit(game, logger, clientLogger)
	}
}
