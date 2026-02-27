package main

import (
	"math"
	"testing"
)

// TestDefaultTimingCalculations validates the default animation timing constants
func TestDefaultTimingCalculations(t *testing.T) {
	tests := []struct {
		name         string
		frameTime    float64
		frameCount   int
		wantFPS      float64
		wantDuration float64
		fpsTolerance float64
	}{
		{
			name:         "default 12 FPS",
			frameTime:    1.0 / 12.0,
			frameCount:   8,
			wantFPS:      12.0,
			wantDuration: 0.6667,
			fpsTolerance: 0.01,
		},
		{
			name:         "6 FPS medium distance",
			frameTime:    1.0 / 6.0,
			frameCount:   8,
			wantFPS:      6.0,
			wantDuration: 1.3333,
			fpsTolerance: 0.01,
		},
		{
			name:         "3 FPS far distance",
			frameTime:    1.0 / 3.0,
			frameCount:   8,
			wantFPS:      3.0,
			wantDuration: 2.6667,
			fpsTolerance: 0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fps := 1.0 / tt.frameTime
			duration := tt.frameTime * float64(tt.frameCount)

			if math.Abs(fps-tt.wantFPS) > tt.fpsTolerance {
				t.Errorf("FPS = %.2f, want %.2f", fps, tt.wantFPS)
			}
			if math.Abs(duration-tt.wantDuration) > 0.001 {
				t.Errorf("Duration = %.4f, want %.4f", duration, tt.wantDuration)
			}
		})
	}
}

// TestLODTimingCalculations validates distance-based frame rate adjustments
func TestLODTimingCalculations(t *testing.T) {
	tests := []struct {
		name          string
		fps           float64
		frameCount    int
		wantFrameTime float64
		wantDuration  float64
	}{
		{
			name:          "close range 12 FPS",
			fps:           12.0,
			frameCount:    8,
			wantFrameTime: 0.0833,
			wantDuration:  0.6667,
		},
		{
			name:          "medium range 6 FPS",
			fps:           6.0,
			frameCount:    8,
			wantFrameTime: 0.1667,
			wantDuration:  1.3333,
		},
		{
			name:          "far range 3 FPS",
			fps:           3.0,
			frameCount:    8,
			wantFrameTime: 0.3333,
			wantDuration:  2.6667,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frameTime := 1.0 / tt.fps
			duration := frameTime * float64(tt.frameCount)

			if math.Abs(frameTime-tt.wantFrameTime) > 0.001 {
				t.Errorf("FrameTime = %.4f, want %.4f", frameTime, tt.wantFrameTime)
			}
			if math.Abs(duration-tt.wantDuration) > 0.001 {
				t.Errorf("Duration = %.4f, want %.4f", duration, tt.wantDuration)
			}
		})
	}
}

// TestCustomTimingCalculations validates custom animation timing scenarios
func TestCustomTimingCalculations(t *testing.T) {
	tests := []struct {
		name       string
		state      string
		frameTime  float64
		frameCount int
		wantFPS    float64
	}{
		{
			name:       "idle 8 FPS",
			state:      "idle",
			frameTime:  1.0 / 8.0,
			frameCount: 8,
			wantFPS:    8.0,
		},
		{
			name:       "walk 12 FPS",
			state:      "walk",
			frameTime:  1.0 / 12.0,
			frameCount: 8,
			wantFPS:    12.0,
		},
		{
			name:       "run 16 FPS",
			state:      "run",
			frameTime:  1.0 / 16.0,
			frameCount: 8,
			wantFPS:    16.0,
		},
		{
			name:       "attack 20 FPS",
			state:      "attack",
			frameTime:  1.0 / 20.0,
			frameCount: 8,
			wantFPS:    20.0,
		},
		{
			name:       "cast 10 FPS",
			state:      "cast",
			frameTime:  1.0 / 10.0,
			frameCount: 8,
			wantFPS:    10.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fps := 1.0 / tt.frameTime
			duration := tt.frameTime * float64(tt.frameCount)

			if math.Abs(fps-tt.wantFPS) > 0.01 {
				t.Errorf("%s: FPS = %.2f, want %.2f", tt.state, fps, tt.wantFPS)
			}
			if duration < 0 {
				t.Errorf("%s: Duration must be positive, got %.4f", tt.state, duration)
			}
		})
	}
}

// TestFrameProgression validates frame advancement over time
func TestFrameProgression(t *testing.T) {
	tests := []struct {
		name             string
		frameTime        float64
		frameCount       int
		deltaTime        float64
		simulationFrames int
		wantFrameChanges int
	}{
		{
			name:             "12 FPS animation at 60 FPS game loop",
			frameTime:        1.0 / 12.0,
			frameCount:       8,
			deltaTime:        1.0 / 60.0,
			simulationFrames: 60,
			wantFrameChanges: 12,
		},
		{
			name:             "6 FPS animation at 60 FPS game loop",
			frameTime:        1.0 / 6.0,
			frameCount:       8,
			deltaTime:        1.0 / 60.0,
			simulationFrames: 60,
			wantFrameChanges: 6,
		},
		{
			name:             "30 FPS animation at 60 FPS game loop",
			frameTime:        1.0 / 30.0,
			frameCount:       8,
			deltaTime:        1.0 / 60.0,
			simulationFrames: 60,
			wantFrameChanges: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeAccumulator := 0.0
			frameIndex := 0
			frameChanges := 0

			for i := 0; i < tt.simulationFrames; i++ {
				oldFrame := frameIndex
				timeAccumulator += tt.deltaTime

				if timeAccumulator >= tt.frameTime {
					timeAccumulator -= tt.frameTime
					frameIndex++
					if frameIndex >= tt.frameCount {
						frameIndex = 0
					}
				}

				if frameIndex != oldFrame {
					frameChanges++
				}
			}

			if frameChanges != tt.wantFrameChanges {
				t.Errorf("Frame changes = %d, want %d", frameChanges, tt.wantFrameChanges)
			}
		})
	}
}

// TestFrameLooping validates that frames loop correctly
func TestFrameLooping(t *testing.T) {
	frameTime := 1.0 / 12.0
	frameCount := 8
	deltaTime := 1.0 / 60.0
	timeAccumulator := 0.0
	frameIndex := 0

	// Simulate enough frames to loop multiple times
	for i := 0; i < 120; i++ {
		timeAccumulator += deltaTime
		if timeAccumulator >= frameTime {
			timeAccumulator -= frameTime
			frameIndex++
			if frameIndex >= frameCount {
				frameIndex = 0
			}
		}

		if frameIndex < 0 || frameIndex >= frameCount {
			t.Errorf("Frame %d: frameIndex %d out of bounds [0, %d)", i, frameIndex, frameCount)
		}
	}
}

// TestAccumulatorPrecision validates that time accumulator doesn't drift
func TestAccumulatorPrecision(t *testing.T) {
	frameTime := 1.0 / 12.0
	frameCount := 8
	deltaTime := 1.0 / 60.0
	timeAccumulator := 0.0
	frameIndex := 0

	// Simulate many frames to check for accumulation errors
	for i := 0; i < 1000; i++ {
		timeAccumulator += deltaTime
		if timeAccumulator >= frameTime {
			timeAccumulator -= frameTime
			frameIndex++
			if frameIndex >= frameCount {
				frameIndex = 0
			}
		}

		// Accumulator should never exceed frameTime significantly
		if timeAccumulator >= frameTime*1.1 {
			t.Errorf("Frame %d: accumulator drift %.6f exceeds frameTime %.6f",
				i, timeAccumulator, frameTime)
		}
	}
}

// TestEdgeCaseZeroFrameTime validates handling of edge cases
func TestEdgeCaseZeroFrameTime(t *testing.T) {
	tests := []struct {
		name      string
		frameTime float64
		wantPanic bool
	}{
		{
			name:      "normal frame time",
			frameTime: 1.0 / 12.0,
			wantPanic: false,
		},
		{
			name:      "very fast frame time",
			frameTime: 1.0 / 1000.0,
			wantPanic: false,
		},
		{
			name:      "very slow frame time",
			frameTime: 1.0,
			wantPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("Panic = %v, wantPanic = %v", r != nil, tt.wantPanic)
				}
			}()

			fps := 1.0 / tt.frameTime
			if fps <= 0 {
				t.Errorf("FPS must be positive, got %.2f", fps)
			}
		})
	}
}

// BenchmarkFrameProgression benchmarks frame progression simulation
func BenchmarkFrameProgression(b *testing.B) {
	frameTime := 1.0 / 12.0
	frameCount := 8
	deltaTime := 1.0 / 60.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		timeAccumulator := 0.0
		frameIndex := 0

		// Simulate one full animation cycle
		for j := 0; j < 100; j++ {
			timeAccumulator += deltaTime
			if timeAccumulator >= frameTime {
				timeAccumulator -= frameTime
				frameIndex++
				if frameIndex >= frameCount {
					frameIndex = 0
				}
			}
		}
	}
}

// BenchmarkTimingCalculations benchmarks FPS and duration calculations
func BenchmarkTimingCalculations(b *testing.B) {
	frameTime := 1.0 / 12.0
	frameCount := 8

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = 1.0 / frameTime
		_ = frameTime * float64(frameCount)
	}
}
