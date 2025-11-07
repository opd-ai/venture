// Package postprocess provides post-processing effects for rendered scenes.
package postprocess

import (
	"image"
	"image/color"
	"testing"
)

func TestEffectType_String(t *testing.T) {
	tests := []struct {
		name string
		et   EffectType
		want string
	}{
		{"motion blur", EffectMotionBlur, "MotionBlur"},
		{"depth blur", EffectDepthBlur, "DepthBlur"},
		{"color grading", EffectColorGrading, "ColorGrading"},
		{"vignette", EffectVignette, "Vignette"},
		{"chromatic aberration", EffectChromaticAberration, "ChromaticAberration"},
		{"unknown", EffectType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.et.String()
			if got != tt.want {
				t.Errorf("EffectType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultConfigs(t *testing.T) {
	t.Run("MotionBlurConfig", func(t *testing.T) {
		config := DefaultMotionBlurConfig()
		if config.Enabled {
			t.Error("DefaultMotionBlurConfig should be disabled")
		}
		if config.Intensity != 0.5 {
			t.Errorf("DefaultMotionBlurConfig.Intensity = %v, want 0.5", config.Intensity)
		}
		if config.Samples != 7 {
			t.Errorf("DefaultMotionBlurConfig.Samples = %v, want 7", config.Samples)
		}
	})

	t.Run("DepthBlurConfig", func(t *testing.T) {
		config := DefaultDepthBlurConfig()
		if config.Enabled {
			t.Error("DefaultDepthBlurConfig should be disabled")
		}
		if config.FocalDistance != 0.5 {
			t.Errorf("DefaultDepthBlurConfig.FocalDistance = %v, want 0.5", config.FocalDistance)
		}
		if config.FocalRange != 0.2 {
			t.Errorf("DefaultDepthBlurConfig.FocalRange = %v, want 0.2", config.FocalRange)
		}
	})

	t.Run("ColorGradingConfig", func(t *testing.T) {
		config := DefaultColorGradingConfig()
		if !config.Enabled {
			t.Error("DefaultColorGradingConfig should be enabled")
		}
		if config.Saturation != 1.0 {
			t.Errorf("DefaultColorGradingConfig.Saturation = %v, want 1.0", config.Saturation)
		}
		if config.Contrast != 1.0 {
			t.Errorf("DefaultColorGradingConfig.Contrast = %v, want 1.0", config.Contrast)
		}
	})

	t.Run("VignetteConfig", func(t *testing.T) {
		config := DefaultVignetteConfig()
		if !config.Enabled {
			t.Error("DefaultVignetteConfig should be enabled")
		}
		if config.Intensity != 0.5 {
			t.Errorf("DefaultVignetteConfig.Intensity = %v, want 0.5", config.Intensity)
		}
	})

	t.Run("ChromaticAberrationConfig", func(t *testing.T) {
		config := DefaultChromaticAberrationConfig()
		if config.Enabled {
			t.Error("DefaultChromaticAberrationConfig should be disabled")
		}
		if config.Intensity != 0.2 {
			t.Errorf("DefaultChromaticAberrationConfig.Intensity = %v, want 0.2", config.Intensity)
		}
	})

	t.Run("Config", func(t *testing.T) {
		config := DefaultConfig()
		if config.MotionBlur.Enabled {
			t.Error("DefaultConfig.MotionBlur should be disabled")
		}
		if !config.ColorGrading.Enabled {
			t.Error("DefaultConfig.ColorGrading should be enabled")
		}
		if !config.Vignette.Enabled {
			t.Error("DefaultConfig.Vignette should be enabled")
		}
	})
}

func TestNewVelocityMap(t *testing.T) {
	tests := []struct {
		name   string
		bounds image.Rectangle
	}{
		{"small", image.Rect(0, 0, 10, 10)},
		{"medium", image.Rect(0, 0, 100, 100)},
		{"offset", image.Rect(50, 50, 150, 150)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			velMap := NewVelocityMap(tt.bounds)

			if velMap.Bounds != tt.bounds {
				t.Errorf("NewVelocityMap bounds = %v, want %v", velMap.Bounds, tt.bounds)
			}

			expectedHeight := tt.bounds.Dy()
			if len(velMap.VelocityX) != expectedHeight {
				t.Errorf("VelocityX height = %d, want %d", len(velMap.VelocityX), expectedHeight)
			}
			if len(velMap.VelocityY) != expectedHeight {
				t.Errorf("VelocityY height = %d, want %d", len(velMap.VelocityY), expectedHeight)
			}

			expectedWidth := tt.bounds.Dx()
			if len(velMap.VelocityX[0]) != expectedWidth {
				t.Errorf("VelocityX width = %d, want %d", len(velMap.VelocityX[0]), expectedWidth)
			}
		})
	}
}

func TestVelocityMap_GetSetVelocity(t *testing.T) {
	bounds := image.Rect(0, 0, 100, 100)
	velMap := NewVelocityMap(bounds)

	tests := []struct {
		name string
		x, y int
		vx   float64
		vy   float64
	}{
		{"center", 50, 50, 5.0, 3.0},
		{"origin", 0, 0, -2.0, 4.0},
		{"corner", 99, 99, 1.5, -1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set velocity
			velMap.SetVelocity(tt.x, tt.y, tt.vx, tt.vy)

			// Get velocity
			gotVx, gotVy := velMap.GetVelocity(tt.x, tt.y)

			if gotVx != tt.vx {
				t.Errorf("GetVelocity vx = %v, want %v", gotVx, tt.vx)
			}
			if gotVy != tt.vy {
				t.Errorf("GetVelocity vy = %v, want %v", gotVy, tt.vy)
			}
		})
	}
}

func TestVelocityMap_OutOfBounds(t *testing.T) {
	bounds := image.Rect(0, 0, 10, 10)
	velMap := NewVelocityMap(bounds)

	// Test get out of bounds
	vx, vy := velMap.GetVelocity(-1, -1)
	if vx != 0 || vy != 0 {
		t.Errorf("GetVelocity out of bounds = (%v, %v), want (0, 0)", vx, vy)
	}

	vx, vy = velMap.GetVelocity(20, 20)
	if vx != 0 || vy != 0 {
		t.Errorf("GetVelocity out of bounds = (%v, %v), want (0, 0)", vx, vy)
	}

	// Test set out of bounds (should not panic)
	velMap.SetVelocity(-1, -1, 5.0, 5.0)
	velMap.SetVelocity(20, 20, 5.0, 5.0)
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "Intensity",
		Message: "must be positive",
	}

	expected := "postprocess: Intensity must be positive"
	if err.Error() != expected {
		t.Errorf("ValidationError.Error() = %q, want %q", err.Error(), expected)
	}
}

func TestPreset(t *testing.T) {
	preset := Preset{
		Name:        "Test",
		Description: "Test preset",
		Config:      DefaultConfig(),
	}

	if preset.Name != "Test" {
		t.Errorf("Preset.Name = %q, want %q", preset.Name, "Test")
	}
	if preset.Description != "Test preset" {
		t.Errorf("Preset.Description = %q, want %q", preset.Description, "Test preset")
	}
}

func TestVignetteConfig_Color(t *testing.T) {
	config := DefaultVignetteConfig()

	// Check color is black
	r, g, b, a := config.Color.RGBA()
	if r != 0 || g != 0 || b != 0 {
		t.Errorf("DefaultVignetteConfig.Color RGB = (%d, %d, %d), want (0, 0, 0)", r>>8, g>>8, b>>8)
	}
	if a != 0xffff {
		t.Errorf("DefaultVignetteConfig.Color alpha = %d, want %d", a, 0xffff)
	}

	// Test custom color
	customColor := color.RGBA{100, 50, 25, 255}
	config.Color = customColor

	r, g, b, a = config.Color.RGBA()
	expectedR := uint32(100 * 257)
	expectedG := uint32(50 * 257)
	expectedB := uint32(25 * 257)

	if r != expectedR || g != expectedG || b != expectedB {
		t.Errorf("Custom color RGB = (%d, %d, %d), want (%d, %d, %d)",
			r, g, b, expectedR, expectedG, expectedB)
	}
}
