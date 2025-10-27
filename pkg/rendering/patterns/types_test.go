package patterns

import (
	"image/color"
	"testing"
)

// TestPatternType_String tests the String method for all pattern types
func TestPatternType_String(t *testing.T) {
	tests := []struct {
		name     string
		pattern  PatternType
		expected string
	}{
		{"Stripes pattern", PatternStripes, "stripes"},
		{"Dots pattern", PatternDots, "dots"},
		{"Gradient pattern", PatternGradient, "gradient"},
		{"Noise pattern", PatternNoise, "noise"},
		{"Checkerboard pattern", PatternCheckerboard, "checkerboard"},
		{"Circles pattern", PatternCircles, "circles"},
		{"Unknown pattern", PatternType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pattern.String()
			if result != tt.expected {
				t.Errorf("PatternType.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestDefaultConfig tests the DefaultConfig function
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Verify type
	if config.Type != PatternStripes {
		t.Errorf("DefaultConfig().Type = %v, want %v", config.Type, PatternStripes)
	}

	// Verify dimensions
	if config.Width != 32 {
		t.Errorf("DefaultConfig().Width = %v, want 32", config.Width)
	}
	if config.Height != 32 {
		t.Errorf("DefaultConfig().Height = %v, want 32", config.Height)
	}

	// Verify seed
	if config.Seed != 0 {
		t.Errorf("DefaultConfig().Seed = %v, want 0", config.Seed)
	}

	// Verify pattern parameters
	if config.Frequency != 4.0 {
		t.Errorf("DefaultConfig().Frequency = %v, want 4.0", config.Frequency)
	}
	if config.Amplitude != 0.5 {
		t.Errorf("DefaultConfig().Amplitude = %v, want 0.5", config.Amplitude)
	}
	if config.Angle != 0 {
		t.Errorf("DefaultConfig().Angle = %v, want 0", config.Angle)
	}

	// Verify colors
	expectedColor1 := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	if config.Color1 != expectedColor1 {
		t.Errorf("DefaultConfig().Color1 = %v, want %v", config.Color1, expectedColor1)
	}

	expectedColor2 := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	if config.Color2 != expectedColor2 {
		t.Errorf("DefaultConfig().Color2 = %v, want %v", config.Color2, expectedColor2)
	}

	// Verify blending
	if config.Opacity != 0.5 {
		t.Errorf("DefaultConfig().Opacity = %v, want 0.5", config.Opacity)
	}
	if config.BlendMode != "overlay" {
		t.Errorf("DefaultConfig().BlendMode = %v, want 'overlay'", config.BlendMode)
	}
}

// TestConfig_CustomValues tests creating a Config with custom values
func TestConfig_CustomValues(t *testing.T) {
	customColor1 := color.RGBA{R: 128, G: 64, B: 32, A: 200}
	customColor2 := color.RGBA{R: 32, G: 128, B: 64, A: 150}

	config := Config{
		Type:      PatternNoise,
		Width:     64,
		Height:    128,
		Seed:      12345,
		Frequency: 8.0,
		Amplitude: 0.75,
		Angle:     45.0,
		Color1:    customColor1,
		Color2:    customColor2,
		Opacity:   0.8,
		BlendMode: "multiply",
	}

	// Verify all custom values
	if config.Type != PatternNoise {
		t.Errorf("Config.Type = %v, want %v", config.Type, PatternNoise)
	}
	if config.Width != 64 {
		t.Errorf("Config.Width = %v, want 64", config.Width)
	}
	if config.Height != 128 {
		t.Errorf("Config.Height = %v, want 128", config.Height)
	}
	if config.Seed != 12345 {
		t.Errorf("Config.Seed = %v, want 12345", config.Seed)
	}
	if config.Frequency != 8.0 {
		t.Errorf("Config.Frequency = %v, want 8.0", config.Frequency)
	}
	if config.Amplitude != 0.75 {
		t.Errorf("Config.Amplitude = %v, want 0.75", config.Amplitude)
	}
	if config.Angle != 45.0 {
		t.Errorf("Config.Angle = %v, want 45.0", config.Angle)
	}
	if config.Color1 != customColor1 {
		t.Errorf("Config.Color1 = %v, want %v", config.Color1, customColor1)
	}
	if config.Color2 != customColor2 {
		t.Errorf("Config.Color2 = %v, want %v", config.Color2, customColor2)
	}
	if config.Opacity != 0.8 {
		t.Errorf("Config.Opacity = %v, want 0.8", config.Opacity)
	}
	if config.BlendMode != "multiply" {
		t.Errorf("Config.BlendMode = %v, want 'multiply'", config.BlendMode)
	}
}

// TestPatternType_Constants tests that pattern type constants are correctly defined
func TestPatternType_Constants(t *testing.T) {
	// Verify constants have expected values (iota sequence)
	tests := []struct {
		name     string
		pattern  PatternType
		expected int
	}{
		{"PatternStripes is 0", PatternStripes, 0},
		{"PatternDots is 1", PatternDots, 1},
		{"PatternGradient is 2", PatternGradient, 2},
		{"PatternNoise is 3", PatternNoise, 3},
		{"PatternCheckerboard is 4", PatternCheckerboard, 4},
		{"PatternCircles is 5", PatternCircles, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.pattern) != tt.expected {
				t.Errorf("Pattern constant value = %v, want %v", int(tt.pattern), tt.expected)
			}
		})
	}
}

// TestConfig_ZeroValues tests that Config with zero values behaves correctly
func TestConfig_ZeroValues(t *testing.T) {
	var config Config

	// Verify zero values
	if config.Type != PatternStripes {
		t.Errorf("Zero Config.Type = %v, want %v (default iota 0)", config.Type, PatternStripes)
	}
	if config.Width != 0 {
		t.Errorf("Zero Config.Width = %v, want 0", config.Width)
	}
	if config.Height != 0 {
		t.Errorf("Zero Config.Height = %v, want 0", config.Height)
	}
	if config.Seed != 0 {
		t.Errorf("Zero Config.Seed = %v, want 0", config.Seed)
	}
	if config.Frequency != 0 {
		t.Errorf("Zero Config.Frequency = %v, want 0", config.Frequency)
	}
	if config.Amplitude != 0 {
		t.Errorf("Zero Config.Amplitude = %v, want 0", config.Amplitude)
	}
	if config.Angle != 0 {
		t.Errorf("Zero Config.Angle = %v, want 0", config.Angle)
	}
	if config.Opacity != 0 {
		t.Errorf("Zero Config.Opacity = %v, want 0", config.Opacity)
	}
	if config.BlendMode != "" {
		t.Errorf("Zero Config.BlendMode = %v, want ''", config.BlendMode)
	}
}
