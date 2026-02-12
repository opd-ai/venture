package engine

import (
	"testing"
	"time"
)

// TestEbitenGame_CurrentFPS verifies that EbitenGame implements FPSProvider interface correctly.
func TestEbitenGame_CurrentFPS(t *testing.T) {
	tests := []struct {
		name        string
		setupGame   func() *EbitenGame
		expectFPS   float64
		description string
	}{
		{
			name: "no_frame_tracker",
			setupGame: func() *EbitenGame {
				return &EbitenGame{
					frameTimeTracker: nil,
				}
			},
			expectFPS:   60.0,
			description: "Should return default 60 FPS when frame tracker is nil",
		},
		{
			name: "with_frame_tracker_no_data",
			setupGame: func() *EbitenGame {
				return &EbitenGame{
					frameTimeTracker: NewFrameTimeTracker(100),
				}
			},
			expectFPS:   0.0,
			description: "Should return 0 FPS when no frames have been recorded",
		},
		{
			name: "with_recorded_frames_60fps",
			setupGame: func() *EbitenGame {
				game := &EbitenGame{
					frameTimeTracker: NewFrameTimeTracker(100),
				}
				// Simulate 60 FPS (16.67ms per frame)
				frameDuration := time.Millisecond * 16
				for i := 0; i < 60; i++ {
					game.frameTimeTracker.RecordFrame(frameDuration)
				}
				return game
			},
			expectFPS:   62.5, // 1000ms / 16ms = 62.5 FPS
			description: "Should return ~60 FPS when frames are ~16ms",
		},
		{
			name: "with_recorded_frames_30fps",
			setupGame: func() *EbitenGame {
				game := &EbitenGame{
					frameTimeTracker: NewFrameTimeTracker(100),
				}
				// Simulate 30 FPS (33.33ms per frame)
				frameDuration := time.Millisecond * 33
				for i := 0; i < 30; i++ {
					game.frameTimeTracker.RecordFrame(frameDuration)
				}
				return game
			},
			expectFPS:   30.3, // 1000ms / 33ms = ~30.3 FPS
			description: "Should return ~30 FPS when frames are ~33ms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := tt.setupGame()
			fps := game.CurrentFPS()

			// Allow small tolerance for floating point comparison
			tolerance := 0.5
			if fps < tt.expectFPS-tolerance || fps > tt.expectFPS+tolerance {
				t.Errorf("%s: got FPS=%.2f, want ~%.2f", tt.description, fps, tt.expectFPS)
			}
		})
	}
}

// TestEbitenGame_CurrentFPS_Integration verifies FPS tracking over multiple frames.
func TestEbitenGame_CurrentFPS_Integration(t *testing.T) {
	game := &EbitenGame{
		frameTimeTracker: NewFrameTimeTracker(100),
	}

	// Simulate varying frame times
	frameTimes := []time.Duration{
		15 * time.Millisecond, // Fast frame
		16 * time.Millisecond, // Normal
		17 * time.Millisecond, // Normal
		25 * time.Millisecond, // Slow frame (stutter)
		16 * time.Millisecond, // Back to normal
	}

	for _, ft := range frameTimes {
		game.frameTimeTracker.RecordFrame(ft)
	}

	fps := game.CurrentFPS()

	// Average: (15+16+17+25+16)/5 = 17.8ms per frame
	// Expected FPS: 1000/17.8 = ~56.2 FPS
	expectedFPS := 56.2
	tolerance := 1.0

	if fps < expectedFPS-tolerance || fps > expectedFPS+tolerance {
		t.Errorf("Integration test: got FPS=%.2f, want ~%.2f", fps, expectedFPS)
	}
}

// TestEbitenGame_CurrentFPS_ThreadSafety verifies concurrent access is safe.
func TestEbitenGame_CurrentFPS_ThreadSafety(t *testing.T) {
	game := &EbitenGame{
		frameTimeTracker: NewFrameTimeTracker(1000),
	}

	// Concurrent writes (recording frames)
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				game.frameTimeTracker.RecordFrame(16 * time.Millisecond)
			}
			done <- true
		}()
	}

	// Concurrent reads (getting FPS)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = game.CurrentFPS()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify we got a reasonable FPS value
	fps := game.CurrentFPS()
	if fps < 50 || fps > 70 {
		t.Errorf("Thread safety test: got FPS=%.2f, expected ~60", fps)
	}
}

// BenchmarkEbitenGame_CurrentFPS measures the performance of FPS retrieval.
func BenchmarkEbitenGame_CurrentFPS(b *testing.B) {
	game := &EbitenGame{
		frameTimeTracker: NewFrameTimeTracker(1000),
	}

	// Pre-populate with frame data
	for i := 0; i < 100; i++ {
		game.frameTimeTracker.RecordFrame(16 * time.Millisecond)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = game.CurrentFPS()
	}
}
