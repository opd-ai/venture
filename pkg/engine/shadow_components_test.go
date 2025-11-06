package engine

import (
	"image/color"
	"testing"
)

func TestShadowComponent_Type(t *testing.T) {
	shadow := NewShadowComponent(16)
	if shadow.Type() != "shadow" {
		t.Errorf("Expected type 'shadow', got '%s'", shadow.Type())
	}
}

func TestNewShadowComponent(t *testing.T) {
	tests := []struct {
		name            string
		radius          float64
		expectedRadius  float64
		expectedType    ShadowType
		expectedOpacity float64
	}{
		{"valid radius", 32, 32, ShadowTypeHard, 0.5},
		{"zero radius uses default", 0, 16, ShadowTypeHard, 0.5},
		{"negative radius uses default", -10, 16, ShadowTypeHard, 0.5},
		{"large radius", 100, 100, ShadowTypeHard, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shadow := NewShadowComponent(tt.radius)

			if shadow.Radius != tt.expectedRadius {
				t.Errorf("Expected radius %.2f, got %.2f", tt.expectedRadius, shadow.Radius)
			}
			if shadow.ShadowType != tt.expectedType {
				t.Errorf("Expected shadow type %v, got %v", tt.expectedType, shadow.ShadowType)
			}
			if shadow.Opacity != tt.expectedOpacity {
				t.Errorf("Expected opacity %.2f, got %.2f", tt.expectedOpacity, shadow.Opacity)
			}
			if !shadow.Enabled {
				t.Error("Expected shadow to be enabled by default")
			}
			if !shadow.CastsShadow {
				t.Error("Expected shadow to cast by default")
			}
			if !shadow.ReceivesShadow {
				t.Error("Expected shadow to receive by default")
			}
		})
	}
}

func TestNewSoftShadow(t *testing.T) {
	shadow := NewSoftShadow(32, 8)

	if shadow.ShadowType != ShadowTypeSoft {
		t.Errorf("Expected ShadowTypeSoft, got %v", shadow.ShadowType)
	}
	if shadow.Radius != 32 {
		t.Errorf("Expected radius 32, got %.2f", shadow.Radius)
	}
	if shadow.SoftEdgeRadius != 8 {
		t.Errorf("Expected soft edge radius 8, got %.2f", shadow.SoftEdgeRadius)
	}
}

func TestNewContactShadow(t *testing.T) {
	shadow := NewContactShadow(24, 16)

	if shadow.ShadowType != ShadowTypeContact {
		t.Errorf("Expected ShadowTypeContact, got %v", shadow.ShadowType)
	}
	if shadow.Radius != 24 {
		t.Errorf("Expected radius 24, got %.2f", shadow.Radius)
	}
	if shadow.Height != 16 {
		t.Errorf("Expected height 16, got %.2f", shadow.Height)
	}
	if shadow.Opacity != 0.3 {
		t.Errorf("Expected contact shadow opacity 0.3, got %.2f", shadow.Opacity)
	}
}

func TestShadowType_String(t *testing.T) {
	tests := []struct {
		shadowType ShadowType
		expected   string
	}{
		{ShadowTypeHard, "hard"},
		{ShadowTypeSoft, "soft"},
		{ShadowTypeContact, "contact"},
		{ShadowType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.shadowType.String()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestShadowComponent_Fields(t *testing.T) {
	shadow := NewShadowComponent(32)

	// Test field modifications
	shadow.Enabled = false
	if shadow.Enabled {
		t.Error("Expected Enabled to be false")
	}

	shadow.CastsShadow = false
	if shadow.CastsShadow {
		t.Error("Expected CastsShadow to be false")
	}

	shadow.ReceivesShadow = false
	if shadow.ReceivesShadow {
		t.Error("Expected ReceivesShadow to be false")
	}

	shadow.Opacity = 0.8
	if shadow.Opacity != 0.8 {
		t.Errorf("Expected opacity 0.8, got %.2f", shadow.Opacity)
	}

	shadow.ShadowType = ShadowTypeSoft
	if shadow.ShadowType != ShadowTypeSoft {
		t.Errorf("Expected ShadowTypeSoft, got %v", shadow.ShadowType)
	}

	testColor := color.RGBA{100, 100, 100, 200}
	shadow.Color = testColor
	if shadow.Color != testColor {
		t.Errorf("Expected color %v, got %v", testColor, shadow.Color)
	}
}

func TestAmbientOcclusionComponent_Type(t *testing.T) {
	ao := NewAmbientOcclusionComponent(0.3, 32)
	if ao.Type() != "ambient_occlusion" {
		t.Errorf("Expected type 'ambient_occlusion', got '%s'", ao.Type())
	}
}

func TestNewAmbientOcclusionComponent(t *testing.T) {
	tests := []struct {
		name              string
		intensity         float64
		radius            float64
		expectedIntensity float64
		expectedRadius    float64
	}{
		{"valid values", 0.5, 64, 0.5, 64},
		{"zero intensity uses default", 0, 32, 0.3, 32},
		{"negative intensity uses default", -0.1, 32, 0.3, 32},
		{"zero radius uses default", 0.3, 0, 0.3, 32},
		{"negative radius uses default", 0.3, -10, 0.3, 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ao := NewAmbientOcclusionComponent(tt.intensity, tt.radius)

			if ao.Intensity != tt.expectedIntensity {
				t.Errorf("Expected intensity %.2f, got %.2f", tt.expectedIntensity, ao.Intensity)
			}
			if ao.Radius != tt.expectedRadius {
				t.Errorf("Expected radius %.2f, got %.2f", tt.expectedRadius, ao.Radius)
			}
			if !ao.Enabled {
				t.Error("Expected AO to be enabled by default")
			}
			if !ao.CornerDarkening {
				t.Error("Expected corner darkening to be enabled by default")
			}
			if ao.Samples != 8 {
				t.Errorf("Expected 8 samples, got %d", ao.Samples)
			}
			if ao.CornerAmount != 0.5 {
				t.Errorf("Expected corner amount 0.5, got %.2f", ao.CornerAmount)
			}
		})
	}
}

func TestAmbientOcclusionComponent_Fields(t *testing.T) {
	ao := NewAmbientOcclusionComponent(0.3, 32)

	// Test field modifications
	ao.Enabled = false
	if ao.Enabled {
		t.Error("Expected Enabled to be false")
	}

	ao.Intensity = 0.7
	if ao.Intensity != 0.7 {
		t.Errorf("Expected intensity 0.7, got %.2f", ao.Intensity)
	}

	ao.Radius = 50
	if ao.Radius != 50 {
		t.Errorf("Expected radius 50, got %.2f", ao.Radius)
	}

	ao.Samples = 16
	if ao.Samples != 16 {
		t.Errorf("Expected samples 16, got %d", ao.Samples)
	}

	ao.CornerDarkening = false
	if ao.CornerDarkening {
		t.Error("Expected CornerDarkening to be false")
	}

	ao.CornerAmount = 0.9
	if ao.CornerAmount != 0.9 {
		t.Errorf("Expected corner amount 0.9, got %.2f", ao.CornerAmount)
	}
}

func TestShadowComponent_DefaultColor(t *testing.T) {
	shadow := NewShadowComponent(16)

	// Default color should be semi-transparent black
	if shadow.Color.R != 0 || shadow.Color.G != 0 || shadow.Color.B != 0 {
		t.Errorf("Expected black shadow color, got RGB(%d, %d, %d)",
			shadow.Color.R, shadow.Color.G, shadow.Color.B)
	}
	if shadow.Color.A != 128 {
		t.Errorf("Expected alpha 128, got %d", shadow.Color.A)
	}
}

func TestShadowComponent_HeightForContactShadows(t *testing.T) {
	tests := []struct {
		name   string
		height float64
	}{
		{"ground level", 0},
		{"low height", 8},
		{"medium height", 16},
		{"high height", 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shadow := NewContactShadow(24, tt.height)
			if shadow.Height != tt.height {
				t.Errorf("Expected height %.2f, got %.2f", tt.height, shadow.Height)
			}
		})
	}
}
