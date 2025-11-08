package ui

import (
	"image"
	"image/color"
	"math"
	"testing"
)

func TestTransitionType_String(t *testing.T) {
	tests := []struct {
		name     string
		tType    TransitionType
		expected string
	}{
		{"None", TransitionNone, "none"},
		{"Fade", TransitionFade, "fade"},
		{"SlideLeft", TransitionSlideLeft, "slide-left"},
		{"SlideRight", TransitionSlideRight, "slide-right"},
		{"SlideUp", TransitionSlideUp, "slide-up"},
		{"SlideDown", TransitionSlideDown, "slide-down"},
		{"Zoom", TransitionZoom, "zoom"},
		{"Unknown", TransitionType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tType.String()
			if got != tt.expected {
				t.Errorf("TransitionType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEasingFunction_String(t *testing.T) {
	tests := []struct {
		name     string
		easing   EasingFunction
		expected string
	}{
		{"Linear", EaseLinear, "linear"},
		{"InQuad", EaseInQuad, "ease-in-quad"},
		{"OutQuad", EaseOutQuad, "ease-out-quad"},
		{"InOutQuad", EaseInOutQuad, "ease-in-out-quad"},
		{"InCubic", EaseInCubic, "ease-in-cubic"},
		{"OutCubic", EaseOutCubic, "ease-out-cubic"},
		{"InOutCubic", EaseInOutCubic, "ease-in-out-cubic"},
		{"Unknown", EasingFunction(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.easing.String()
			if got != tt.expected {
				t.Errorf("EasingFunction.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDefaultTransitionConfig(t *testing.T) {
	config := DefaultTransitionConfig()

	if config.Type != TransitionFade {
		t.Errorf("Default Type = %v, want TransitionFade", config.Type)
	}
	if config.Duration != 300.0 {
		t.Errorf("Default Duration = %v, want 300.0", config.Duration)
	}
	if config.Easing != EaseInOutQuad {
		t.Errorf("Default Easing = %v, want EaseInOutQuad", config.Easing)
	}
	if config.Progress != 0.0 {
		t.Errorf("Default Progress = %v, want 0.0", config.Progress)
	}
}

func TestTransitionConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  TransitionConfig
		wantErr bool
	}{
		{
			name: "Valid",
			config: TransitionConfig{
				Type:     TransitionFade,
				Duration: 300.0,
				Easing:   EaseLinear,
				Progress: 0.5,
			},
			wantErr: false,
		},
		{
			name: "NegativeDuration",
			config: TransitionConfig{
				Type:     TransitionFade,
				Duration: -100.0,
				Easing:   EaseLinear,
				Progress: 0.5,
			},
			wantErr: true,
		},
		{
			name: "ProgressTooLow",
			config: TransitionConfig{
				Type:     TransitionFade,
				Duration: 300.0,
				Easing:   EaseLinear,
				Progress: -0.1,
			},
			wantErr: true,
		},
		{
			name: "ProgressTooHigh",
			config: TransitionConfig{
				Type:     TransitionFade,
				Duration: 300.0,
				Easing:   EaseLinear,
				Progress: 1.5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApplyEasing(t *testing.T) {
	tests := []struct {
		name     string
		progress float64
		easing   EasingFunction
	}{
		{"Linear_0", 0.0, EaseLinear},
		{"Linear_0.5", 0.5, EaseLinear},
		{"Linear_1", 1.0, EaseLinear},
		{"InQuad_0.5", 0.5, EaseInQuad},
		{"OutQuad_0.5", 0.5, EaseOutQuad},
		{"InOutQuad_0.5", 0.5, EaseInOutQuad},
		{"InCubic_0.5", 0.5, EaseInCubic},
		{"OutCubic_0.5", 0.5, EaseOutCubic},
		{"InOutCubic_0.5", 0.5, EaseInOutCubic},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyEasing(tt.progress, tt.easing)

			// Result should be in [0, 1]
			if result < 0 || result > 1 {
				t.Errorf("ApplyEasing() = %v, should be in [0, 1]", result)
			}

			// Linear should return input
			if tt.easing == EaseLinear && result != tt.progress {
				t.Errorf("Linear easing should return input, got %v, want %v", result, tt.progress)
			}
		})
	}
}

func TestApplyEasing_Bounds(t *testing.T) {
	// Test clamping of out-of-bounds values
	tests := []struct {
		name     string
		progress float64
		expected float64
	}{
		{"BelowZero", -0.5, 0.0},
		{"AboveOne", 1.5, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyEasing(tt.progress, EaseLinear)
			if result != tt.expected {
				t.Errorf("ApplyEasing(%v) = %v, want %v", tt.progress, result, tt.expected)
			}
		})
	}
}

func TestApplyEasing_Monotonic(t *testing.T) {
	// All easing functions should be monotonic (never decrease)
	easings := []EasingFunction{
		EaseLinear, EaseInQuad, EaseOutQuad, EaseInOutQuad,
		EaseInCubic, EaseOutCubic, EaseInOutCubic,
	}

	for _, easing := range easings {
		t.Run(easing.String(), func(t *testing.T) {
			prev := 0.0
			for progress := 0.0; progress <= 1.0; progress += 0.1 {
				result := ApplyEasing(progress, easing)
				if result < prev {
					t.Errorf("Easing function not monotonic at progress=%v: prev=%v, current=%v",
						progress, prev, result)
				}
				prev = result
			}
		})
	}
}

func TestApplyTransition(t *testing.T) {
	gen := NewGenerator()

	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	tests := []struct {
		name   string
		config TransitionConfig
	}{
		{
			name: "Fade",
			config: TransitionConfig{
				Type:     TransitionFade,
				Duration: 300,
				Easing:   EaseLinear,
				Progress: 0.5,
			},
		},
		{
			name: "SlideLeft",
			config: TransitionConfig{
				Type:     TransitionSlideLeft,
				Duration: 300,
				Easing:   EaseLinear,
				Progress: 0.5,
			},
		},
		{
			name: "SlideRight",
			config: TransitionConfig{
				Type:     TransitionSlideRight,
				Duration: 300,
				Easing:   EaseLinear,
				Progress: 0.5,
			},
		},
		{
			name: "SlideUp",
			config: TransitionConfig{
				Type:     TransitionSlideUp,
				Duration: 300,
				Easing:   EaseLinear,
				Progress: 0.5,
			},
		},
		{
			name: "SlideDown",
			config: TransitionConfig{
				Type:     TransitionSlideDown,
				Duration: 300,
				Easing:   EaseLinear,
				Progress: 0.5,
			},
		},
		{
			name: "Zoom",
			config: TransitionConfig{
				Type:     TransitionZoom,
				Duration: 300,
				Easing:   EaseLinear,
				Progress: 0.5,
			},
		},
		{
			name: "None",
			config: TransitionConfig{
				Type:     TransitionNone,
				Duration: 300,
				Easing:   EaseLinear,
				Progress: 0.5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.ApplyTransition(img, tt.config)

			if result == nil {
				t.Fatal("ApplyTransition returned nil")
			}

			// Verify dimensions
			if result.Bounds() != img.Bounds() {
				t.Errorf("Result bounds changed: got %v, want %v", result.Bounds(), img.Bounds())
			}

			// TransitionNone should return original image
			if tt.config.Type == TransitionNone && result != img {
				t.Error("TransitionNone should return original image")
			}
		})
	}
}

func TestInterpolateTransition(t *testing.T) {
	gen := NewGenerator()

	// Create two test images
	img1 := image.NewRGBA(image.Rect(0, 0, 50, 50))
	img2 := image.NewRGBA(image.Rect(0, 0, 50, 50))

	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			img1.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
			img2.Set(x, y, color.RGBA{R: 0, G: 0, B: 255, A: 255})
		}
	}

	tests := []struct {
		name     string
		progress float64
	}{
		{"Start", 0.0},
		{"Quarter", 0.25},
		{"Half", 0.5},
		{"ThreeQuarters", 0.75},
		{"End", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.InterpolateTransition(img1, img2, tt.progress)

			if result == nil {
				t.Fatal("InterpolateTransition returned nil")
			}

			// Check a sample pixel
			c := result.RGBAAt(25, 25)

			// At progress 0, should be fully img1 (red)
			if tt.progress == 0.0 {
				if c.R != 255 || c.B != 0 {
					t.Errorf("At progress 0, expected red, got %v", c)
				}
			}

			// At progress 1, should be fully img2 (blue)
			if tt.progress == 1.0 {
				if c.R != 0 || c.B != 255 {
					t.Errorf("At progress 1, expected blue, got %v", c)
				}
			}

			// At progress 0.5, should be purple (mid-point)
			if tt.progress == 0.5 {
				expectedR := uint8(127)
				expectedB := uint8(127)
				tolerance := uint8(1)

				if !withinTolerance(c.R, expectedR, tolerance) ||
					!withinTolerance(c.B, expectedB, tolerance) {
					t.Errorf("At progress 0.5, expected ~purple (%d, 0, %d), got %v",
						expectedR, expectedB, c)
				}
			}
		})
	}
}

func TestInterpolateTransition_NilImages(t *testing.T) {
	gen := NewGenerator()

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	tests := []struct {
		name     string
		img1     *image.RGBA
		img2     *image.RGBA
		expected *image.RGBA
	}{
		{"BothNil", nil, nil, nil},
		{"FirstNil", nil, img, img},
		{"SecondNil", img, nil, img},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.InterpolateTransition(tt.img1, tt.img2, 0.5)
			if result != tt.expected {
				t.Errorf("InterpolateTransition() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTransitionConfig_UpdateProgress(t *testing.T) {
	tests := []struct {
		name           string
		initialProgress float64
		duration       float64
		deltaTime      float64
		expectedProgress float64
	}{
		{"NormalUpdate", 0.0, 300.0, 150.0, 0.5},
		{"Complete", 0.9, 300.0, 50.0, 1.0},
		{"OverComplete", 0.5, 300.0, 200.0, 1.0},
		{"ZeroDuration", 0.0, 0.0, 100.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := TransitionConfig{
				Progress: tt.initialProgress,
				Duration: tt.duration,
			}

			result := config.UpdateProgress(tt.deltaTime)

			if math.Abs(result-tt.expectedProgress) > 0.001 {
				t.Errorf("UpdateProgress() = %v, want %v", result, tt.expectedProgress)
			}
			if math.Abs(config.Progress-tt.expectedProgress) > 0.001 {
				t.Errorf("config.Progress = %v, want %v", config.Progress, tt.expectedProgress)
			}
		})
	}
}

func TestTransitionConfig_IsComplete(t *testing.T) {
	tests := []struct {
		name     string
		progress float64
		expected bool
	}{
		{"NotComplete", 0.5, false},
		{"JustComplete", 1.0, true},
		{"OverComplete", 1.1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := TransitionConfig{Progress: tt.progress}
			result := config.IsComplete()

			if result != tt.expected {
				t.Errorf("IsComplete() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTransitionConfig_GetTransitionDuration(t *testing.T) {
	config := TransitionConfig{Duration: 500.0}
	if config.GetTransitionDuration() != 500.0 {
		t.Errorf("GetTransitionDuration() = %v, want 500.0", config.GetTransitionDuration())
	}
}

// Helper function to check if two uint8 values are within tolerance
func withinTolerance(a, b, tolerance uint8) bool {
	diff := int(a) - int(b)
	if diff < 0 {
		diff = -diff
	}
	return diff <= int(tolerance)
}
