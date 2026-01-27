package engine

import (
	"testing"
)

type StubInventoryUI struct {
	UpdateCount int
	DrawCount   int
	active      bool
}

func NewStubInventoryUI() *StubInventoryUI {
	return &StubInventoryUI{}
}

func (s *StubInventoryUI) Update(entities []*Entity, deltaTime float64) {
	s.UpdateCount++
}

func (s *StubInventoryUI) Draw(screen interface{}) {
	s.DrawCount++
}

func (s *StubInventoryUI) IsActive() bool {
	return s.active
}

func (s *StubInventoryUI) SetActive(active bool) {
	s.active = active
}

var _ UISystem = (*StubInventoryUI)(nil)

// TestEaseInOutCubic tests the easing function behavior.
func TestEaseInOutCubic(t *testing.T) {
	tests := []struct {
		name    string
		input   float64
		wantMin float64
		wantMax float64
	}{
		{"start", 0.0, 0.0, 0.0},
		{"quarter", 0.25, 0.0, 0.2},
		{"half", 0.5, 0.45, 0.55},
		{"three-quarter", 0.75, 0.8, 1.0},
		{"end", 1.0, 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := easeInOutCubic(tt.input)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("easeInOutCubic(%f) = %f, want between %f and %f",
					tt.input, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// TestTransitionStateProgression tests state transitions without Ebiten.
func TestTransitionStateProgression(t *testing.T) {
	// Test data structure that mimics EbitenInventoryUI transition logic
	type transitionTest struct {
		state    TransitionState
		progress float64
		alpha    float64
	}

	tests := []struct {
		name     string
		initial  transitionTest
		deltaT   float64
		duration float64
		want     transitionTest
	}{
		{
			name: "fade-in start",
			initial: transitionTest{
				state:    TransitionFadeIn,
				progress: 0.0,
				alpha:    0.0,
			},
			deltaT:   1.0 / 60.0, // 16.67ms
			duration: 0.2,        // 200ms
			want: transitionTest{
				state:    TransitionFadeIn,
				progress: 0.083, // ~8.3% after 1 frame
				alpha:    0.002, // eased value is very small at start
			},
		},
		{
			name: "fade-in complete",
			initial: transitionTest{
				state:    TransitionFadeIn,
				progress: 0.9,
				alpha:    0.972, // eased value near end
			},
			deltaT:   1.0 / 60.0,
			duration: 0.2,
			want: transitionTest{
				state:    TransitionVisible,
				progress: 1.0,
				alpha:    1.0,
			},
		},
		{
			name: "fade-out start",
			initial: transitionTest{
				state:    TransitionFadeOut,
				progress: 0.0,
				alpha:    1.0,
			},
			deltaT:   1.0 / 60.0,
			duration: 0.2,
			want: transitionTest{
				state:    TransitionFadeOut,
				progress: 0.083,
				alpha:    0.998, // 1.0 - eased(0.083)
			},
		},
		{
			name: "fade-out complete",
			initial: transitionTest{
				state:    TransitionFadeOut,
				progress: 0.9,
				alpha:    0.028, // 1.0 - eased(0.9)
			},
			deltaT:   1.0 / 60.0,
			duration: 0.2,
			want: transitionTest{
				state:    TransitionHidden,
				progress: 0.0,
				alpha:    0.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate updateTransition logic
			state := tt.initial.state
			progress := tt.initial.progress
			var alpha float64

			switch state {
			case TransitionFadeIn:
				progress += tt.deltaT / tt.duration
				if progress >= 1.0 {
					progress = 1.0
					state = TransitionVisible
				}
				alpha = easeInOutCubic(progress)

			case TransitionFadeOut:
				progress += tt.deltaT / tt.duration
				if progress >= 1.0 {
					progress = 0.0
					state = TransitionHidden
				}
				alpha = 1.0 - easeInOutCubic(progress)

			case TransitionVisible:
				alpha = 1.0

			case TransitionHidden:
				alpha = 0.0
			}

			if state != tt.want.state {
				t.Errorf("state = %v, want %v", state, tt.want.state)
			}
			if absFloat(progress-tt.want.progress) > 0.01 {
				t.Errorf("progress = %f, want ~%f", progress, tt.want.progress)
			}
			if absFloat(alpha-tt.want.alpha) > 0.01 {
				t.Errorf("alpha = %f, want ~%f", alpha, tt.want.alpha)
			}
		})
	}
}

// TestTransitionDurationScaling tests different transition durations.
func TestTransitionDurationScaling(t *testing.T) {
	durations := []float64{0.1, 0.2, 0.5}

	for _, duration := range durations {
		t.Run("duration_"+formatFloat(duration), func(t *testing.T) {
			// Simulate fade-in with given duration
			state := TransitionFadeIn
			progress := 0.0
			frameTime := 1.0 / 60.0

			frames := 0
			for state == TransitionFadeIn && frames < 120 { // max 2 seconds
				progress += frameTime / duration
				if progress >= 1.0 {
					progress = 1.0
					state = TransitionVisible
				}
				frames++
			}

			// Expected frames = duration / frameTime
			expectedFrames := int(duration / frameTime)
			if absFloat(float64(frames)-float64(expectedFrames)) > 1 {
				t.Errorf("Expected ~%d frames for %fs transition, got %d",
					expectedFrames, duration, frames)
			}
			if state != TransitionVisible {
				t.Errorf("Expected TransitionVisible after %d frames, got %v",
					frames, state)
			}
		})
	}
}

// Helper functions
func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func formatFloat(f float64) string {
	if f == 0.1 {
		return "100ms"
	} else if f == 0.2 {
		return "200ms"
	} else if f == 0.5 {
		return "500ms"
	}
	return "unknown"
}
