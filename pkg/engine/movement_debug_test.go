// Package engine provides tests for movement system debug logging optimization.
// This file tests the cached debug flag optimization (R6) which eliminates
// per-frame GetLevel() calls for ~1µs/frame performance improvement.
package engine

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

// TestMovementDebugFlagCaching verifies debug flag is cached at system creation.
func TestMovementDebugFlagCaching(t *testing.T) {
	tests := []struct {
		name          string
		logLevel      log.Level
		expectedDebug bool
	}{
		{"debug level", log.DebugLevel, true},
		{"info level", log.InfoLevel, false},
		{"warn level", log.WarnLevel, false},
		{"error level", log.ErrorLevel, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original level
			originalLevel := log.GetLevel()
			defer log.SetLevel(originalLevel)

			// Set desired level
			log.SetLevel(tt.logLevel)

			// Create movement system (should cache debug flag)
			_ = NewMovementSystem(200.0)

			// Verify cached flag matches expected value
			if movementDebugEnabled != tt.expectedDebug {
				t.Errorf("movementDebugEnabled = %v, want %v", movementDebugEnabled, tt.expectedDebug)
			}
		})
	}
}

// TestSetMovementDebugEnabled verifies manual flag setting.
func TestSetMovementDebugEnabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{"enable debug", true},
		{"disable debug", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetMovementDebugEnabled(tt.enabled)

			if movementDebugEnabled != tt.enabled {
				t.Errorf("movementDebugEnabled = %v, want %v", movementDebugEnabled, tt.enabled)
			}
		})
	}
}

// TestRefreshMovementDebugFlag verifies flag refresh from log level.
func TestRefreshMovementDebugFlag(t *testing.T) {
	// Save original level
	originalLevel := log.GetLevel()
	defer log.SetLevel(originalLevel)

	tests := []struct {
		name          string
		logLevel      log.Level
		expectedDebug bool
	}{
		{"refresh to debug", log.DebugLevel, true},
		{"refresh to info", log.InfoLevel, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.SetLevel(tt.logLevel)
			RefreshMovementDebugFlag()

			if movementDebugEnabled != tt.expectedDebug {
				t.Errorf("movementDebugEnabled = %v, want %v after refresh", movementDebugEnabled, tt.expectedDebug)
			}
		})
	}
}

// TestMovementSystem_LogUpdateStart verifies debug logging uses cached flag.
func TestMovementSystem_LogUpdateStart(t *testing.T) {
	// Save original level
	originalLevel := log.GetLevel()
	defer log.SetLevel(originalLevel)

	tests := []struct {
		name          string
		logLevel      log.Level
		expectedDebug bool
	}{
		{"debug enabled", log.DebugLevel, true},
		{"debug disabled", log.InfoLevel, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.SetLevel(tt.logLevel)
			sys := NewMovementSystem(200.0)

			// Call logUpdateStart and verify returned flag matches cached value
			debugEnabled := sys.logUpdateStart(10, 0.016)

			if debugEnabled != tt.expectedDebug {
				t.Errorf("logUpdateStart returned %v, want %v", debugEnabled, tt.expectedDebug)
			}

			if debugEnabled != movementDebugEnabled {
				t.Errorf("logUpdateStart returned %v, but cached flag is %v", debugEnabled, movementDebugEnabled)
			}
		})
	}
}

// TestMovementSystem_DebugLoggingIntegration verifies debug logging in Update.
func TestMovementSystem_DebugLoggingIntegration(t *testing.T) {
	// Save original level
	originalLevel := log.GetLevel()
	defer log.SetLevel(originalLevel)

	tests := []struct {
		name     string
		logLevel log.Level
	}{
		{"with debug logging", log.DebugLevel},
		{"without debug logging", log.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.SetLevel(tt.logLevel)

			world := NewWorld()
			sys := NewMovementSystem(200.0)

			// Create test entity
			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: 100, Y: 100})
			entity.AddComponent(&VelocityComponent{VX: 10, VY: 0})

			entities := []*Entity{entity}

			// Update should work regardless of debug level
			sys.Update(entities, 0.016)

			// Verify entity moved
			pos := entity.GetPosition()
			if pos.X <= 100 {
				t.Error("Entity should have moved")
			}
		})
	}
}

// BenchmarkMovementDebugFlagCaching benchmarks update with/without debug logging.
// This demonstrates the optimization impact of cached debug flag (R6).
func BenchmarkMovementDebugFlagCaching(b *testing.B) {
	benchmarks := []struct {
		name     string
		logLevel log.Level
	}{
		{"with_debug_logging", log.DebugLevel},
		{"without_debug_logging", log.InfoLevel},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Save original level
			originalLevel := log.GetLevel()
			defer log.SetLevel(originalLevel)

			log.SetLevel(bm.logLevel)

			world := NewWorld()
			sys := NewMovementSystem(200.0)

			// Create 100 entities
			entities := make([]*Entity, 100)
			for i := 0; i < 100; i++ {
				e := world.CreateEntity()
				e.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
				e.AddComponent(&VelocityComponent{VX: 1.0, VY: 0.5})
				entities[i] = e
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				sys.Update(entities, 0.016)
			}
		})
	}
}

// BenchmarkMovementLogUpdateStart benchmarks the logUpdateStart method.
// This isolates the cached flag check overhead.
func BenchmarkMovementLogUpdateStart(b *testing.B) {
	benchmarks := []struct {
		name     string
		logLevel log.Level
	}{
		{"with_debug", log.DebugLevel},
		{"without_debug", log.InfoLevel},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Save original level
			originalLevel := log.GetLevel()
			defer log.SetLevel(originalLevel)

			log.SetLevel(bm.logLevel)
			sys := NewMovementSystem(200.0)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				sys.logUpdateStart(100, 0.016)
			}
		})
	}
}

// BenchmarkRefreshMovementDebugFlag benchmarks flag refresh performance.
func BenchmarkRefreshMovementDebugFlag(b *testing.B) {
	// Save original level
	originalLevel := log.GetLevel()
	defer log.SetLevel(originalLevel)

	log.SetLevel(log.InfoLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RefreshMovementDebugFlag()
	}
}
