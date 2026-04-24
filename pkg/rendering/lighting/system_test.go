// Package lighting provides dynamic lighting effects for rendered scenes.
package lighting

import (
	"image"
	"image/color"
	"testing"

	"github.com/sirupsen/logrus"
)

// TestLightType_String tests light type string conversion.
func TestLightType_String(t *testing.T) {
	tests := []struct {
		name      string
		lightType LightType
		expected  string
	}{
		{"ambient", TypeAmbient, "Ambient"},
		{"point", TypePoint, "Point"},
		{"directional", TypeDirectional, "Directional"},
		{"unknown", LightType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.lightType.String(); got != tt.expected {
				t.Errorf("LightType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestFalloffType_String tests falloff type string conversion.
func TestFalloffType_String(t *testing.T) {
	tests := []struct {
		name     string
		falloff  FalloffType
		expected string
	}{
		{"none", FalloffNone, "None"},
		{"linear", FalloffLinear, "Linear"},
		{"quadratic", FalloffQuadratic, "Quadratic"},
		{"inverse_square", FalloffInverseSquare, "InverseSquare"},
		{"unknown", FalloffType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.falloff.String(); got != tt.expected {
				t.Errorf("FalloffType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestLight_Validate tests light validation.
func TestLight_Validate(t *testing.T) {
	tests := []struct {
		name    string
		light   Light
		wantErr bool
	}{
		{
			name: "valid light",
			light: Light{
				Type:      TypePoint,
				Position:  image.Point{X: 100, Y: 100},
				Color:     color.RGBA{255, 255, 255, 255},
				Intensity: 1.0,
				Radius:    50,
				Falloff:   FalloffLinear,
				Enabled:   true,
			},
			wantErr: false,
		},
		{
			name: "negative intensity",
			light: Light{
				Type:      TypePoint,
				Intensity: -0.5,
				Radius:    50,
			},
			wantErr: true,
		},
		{
			name: "negative radius",
			light: Light{
				Type:      TypePoint,
				Intensity: 1.0,
				Radius:    -10,
			},
			wantErr: true,
		},
		{
			name: "zero values valid",
			light: Light{
				Type:      TypeAmbient,
				Intensity: 0,
				Radius:    0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.light.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Light.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDefaultConfig tests default configuration.
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.AmbientColor == nil {
		t.Error("DefaultConfig AmbientColor is nil")
	}
	if config.AmbientIntensity < 0 || config.AmbientIntensity > 1 {
		t.Errorf("DefaultConfig AmbientIntensity = %v, want 0-1", config.AmbientIntensity)
	}
	if config.MaxLights <= 0 {
		t.Errorf("DefaultConfig MaxLights = %v, want positive", config.MaxLights)
	}
	if config.GammaCorrection <= 0 {
		t.Errorf("DefaultConfig GammaCorrection = %v, want positive", config.GammaCorrection)
	}
}

// TestNewSystem tests system creation.
func TestNewSystem(t *testing.T) {
	system := NewSystem()

	if system == nil {
		t.Fatal("NewSystem returned nil")
	}
	if system.LightCount() != 0 {
		t.Errorf("New system has %d lights, want 0", system.LightCount())
	}
}

// TestNewSystemWithConfig tests system creation with custom config.
func TestNewSystemWithConfig(t *testing.T) {
	config := LightingConfig{
		AmbientColor:     color.RGBA{10, 10, 10, 255},
		AmbientIntensity: 0.1,
		MaxLights:        16,
		GammaCorrection:  2.0,
		EnableShadows:    false,
	}

	system := NewSystemWithConfig(config)

	if system == nil {
		t.Fatal("NewSystemWithConfig returned nil")
	}

	gotConfig := system.GetConfig()
	if gotConfig.AmbientIntensity != config.AmbientIntensity {
		t.Errorf("Config AmbientIntensity = %v, want %v", gotConfig.AmbientIntensity, config.AmbientIntensity)
	}
	if gotConfig.MaxLights != config.MaxLights {
		t.Errorf("Config MaxLights = %v, want %v", gotConfig.MaxLights, config.MaxLights)
	}
}

// TestSystem_AddLight tests adding lights to the system.
func TestSystem_AddLight(t *testing.T) {
	system := NewSystem()

	light := Light{
		Type:      TypePoint,
		Position:  image.Point{X: 100, Y: 100},
		Color:     color.RGBA{255, 255, 255, 255},
		Intensity: 1.0,
		Radius:    50,
		Falloff:   FalloffLinear,
		Enabled:   true,
	}

	err := system.AddLight(light)
	if err != nil {
		t.Errorf("AddLight failed: %v", err)
	}

	if system.LightCount() != 1 {
		t.Errorf("LightCount = %d, want 1", system.LightCount())
	}
}

// TestSystem_AddLight_Invalid tests adding invalid lights.
func TestSystem_AddLight_Invalid(t *testing.T) {
	system := NewSystem()

	invalidLight := Light{
		Type:      TypePoint,
		Intensity: -1.0,
		Radius:    50,
	}

	err := system.AddLight(invalidLight)
	if err == nil {
		t.Error("AddLight accepted invalid light")
	}

	if system.LightCount() != 0 {
		t.Errorf("LightCount = %d, want 0", system.LightCount())
	}
}

// TestSystem_AddLight_MaxLights tests maximum light limit.
func TestSystem_AddLight_MaxLights(t *testing.T) {
	config := LightingConfig{
		AmbientColor:     color.RGBA{10, 10, 10, 255},
		AmbientIntensity: 0.1,
		MaxLights:        2,
		GammaCorrection:  2.2,
	}
	system := NewSystemWithConfig(config)

	light := Light{
		Type:      TypePoint,
		Position:  image.Point{X: 100, Y: 100},
		Color:     color.RGBA{255, 255, 255, 255},
		Intensity: 1.0,
		Radius:    50,
		Falloff:   FalloffLinear,
		Enabled:   true,
	}

	// Add up to max
	for i := 0; i < config.MaxLights; i++ {
		err := system.AddLight(light)
		if err != nil {
			t.Errorf("AddLight %d failed: %v", i, err)
		}
	}

	// Try to add one more
	err := system.AddLight(light)
	if err == nil {
		t.Error("AddLight accepted light beyond max")
	}
}

// TestSystem_RemoveLight tests removing lights.
func TestSystem_RemoveLight(t *testing.T) {
	system := NewSystem()

	light := Light{
		Type:      TypePoint,
		Position:  image.Point{X: 100, Y: 100},
		Color:     color.RGBA{255, 255, 255, 255},
		Intensity: 1.0,
		Radius:    50,
		Falloff:   FalloffLinear,
		Enabled:   true,
	}

	_ = system.AddLight(light)
	_ = system.AddLight(light)

	err := system.RemoveLight(0)
	if err != nil {
		t.Errorf("RemoveLight failed: %v", err)
	}

	if system.LightCount() != 1 {
		t.Errorf("LightCount = %d, want 1", system.LightCount())
	}
}

// TestSystem_RemoveLight_Invalid tests removing with invalid index.
func TestSystem_RemoveLight_Invalid(t *testing.T) {
	system := NewSystem()

	err := system.RemoveLight(0)
	if err == nil {
		t.Error("RemoveLight accepted invalid index")
	}

	err = system.RemoveLight(-1)
	if err == nil {
		t.Error("RemoveLight accepted negative index")
	}
}

// TestSystem_ClearLights tests clearing all lights.
func TestSystem_ClearLights(t *testing.T) {
	system := NewSystem()

	light := Light{
		Type:      TypePoint,
		Position:  image.Point{X: 100, Y: 100},
		Color:     color.RGBA{255, 255, 255, 255},
		Intensity: 1.0,
		Radius:    50,
		Falloff:   FalloffLinear,
		Enabled:   true,
	}

	_ = system.AddLight(light)
	_ = system.AddLight(light)

	// Test GetLights before clearing
	lights := system.GetLights()
	if len(lights) != 2 {
		t.Errorf("GetLights returned %d lights, want 2", len(lights))
	}

	system.ClearLights()

	if system.LightCount() != 0 {
		t.Errorf("LightCount = %d, want 0", system.LightCount())
	}

	// Test GetLights after clearing
	lightsAfter := system.GetLights()
	if len(lightsAfter) != 0 {
		t.Errorf("GetLights returned %d lights after clear, want 0", len(lightsAfter))
	}
}

// TestSystem_GetLight tests getting individual lights.
func TestSystem_GetLight(t *testing.T) {
	system := NewSystem()

	light := Light{
		Type:      TypePoint,
		Position:  image.Point{X: 100, Y: 100},
		Color:     color.RGBA{255, 0, 0, 255},
		Intensity: 1.0,
		Radius:    50,
		Falloff:   FalloffLinear,
		Enabled:   true,
	}

	_ = system.AddLight(light)

	gotLight, err := system.GetLight(0)
	if err != nil {
		t.Errorf("GetLight failed: %v", err)
	}

	if gotLight.Type != light.Type {
		t.Errorf("GetLight Type = %v, want %v", gotLight.Type, light.Type)
	}
	if gotLight.Position != light.Position {
		t.Errorf("GetLight Position = %v, want %v", gotLight.Position, light.Position)
	}
}

// TestSystem_GetLight_Invalid tests getting with invalid index.
func TestSystem_GetLight_Invalid(t *testing.T) {
	system := NewSystem()

	_, err := system.GetLight(0)
	if err == nil {
		t.Error("GetLight accepted invalid index")
	}
}

// TestSystem_UpdateLight tests updating lights.
func TestSystem_UpdateLight(t *testing.T) {
	system := NewSystem()

	light := Light{
		Type:      TypePoint,
		Position:  image.Point{X: 100, Y: 100},
		Color:     color.RGBA{255, 255, 255, 255},
		Intensity: 1.0,
		Radius:    50,
		Falloff:   FalloffLinear,
		Enabled:   true,
	}

	_ = system.AddLight(light)

	updatedLight := light
	updatedLight.Intensity = 0.5

	err := system.UpdateLight(0, updatedLight)
	if err != nil {
		t.Errorf("UpdateLight failed: %v", err)
	}

	gotLight, _ := system.GetLight(0)
	if gotLight.Intensity != 0.5 {
		t.Errorf("Updated Intensity = %v, want 0.5", gotLight.Intensity)
	}
}

// TestSystem_UpdateLight_Invalid tests updating with invalid parameters.
func TestSystem_UpdateLight_Invalid(t *testing.T) {
	system := NewSystem()

	light := Light{
		Type:      TypePoint,
		Intensity: 1.0,
		Radius:    50,
	}

	_ = system.AddLight(light)

	// Invalid index
	err := system.UpdateLight(99, light)
	if err == nil {
		t.Error("UpdateLight accepted invalid index")
	}

	// Invalid light
	invalidLight := Light{
		Type:      TypePoint,
		Intensity: -1.0,
		Radius:    50,
	}
	err = system.UpdateLight(0, invalidLight)
	if err == nil {
		t.Error("UpdateLight accepted invalid light")
	}
}

// TestSystem_ApplyLighting tests basic lighting application.
func TestSystem_ApplyLighting(t *testing.T) {
	system := NewSystem()

	// Create a simple white image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{200, 200, 200, 255})
		}
	}

	// Add a point light
	light := Light{
		Type:      TypePoint,
		Position:  image.Point{X: 50, Y: 50},
		Color:     color.RGBA{255, 255, 255, 255},
		Intensity: 1.0,
		Radius:    30,
		Falloff:   FalloffLinear,
		Enabled:   true,
	}
	_ = system.AddLight(light)

	result := system.ApplyLighting(img)

	if result == nil {
		t.Fatal("ApplyLighting returned nil")
	}

	if result.Bounds() != img.Bounds() {
		t.Errorf("Result bounds = %v, want %v", result.Bounds(), img.Bounds())
	}

	// Check that center is brighter than edges
	centerColor := result.At(50, 50)
	edgeColor := result.At(10, 10)

	cr, _, _, _ := centerColor.RGBA()
	er, _, _, _ := edgeColor.RGBA()

	if cr <= er {
		t.Error("Center should be brighter than edge with point light")
	}
}

// TestSystem_ApplyLighting_AmbientOnly tests ambient lighting.
func TestSystem_ApplyLighting_AmbientOnly(t *testing.T) {
	config := LightingConfig{
		AmbientColor:     color.RGBA{50, 50, 50, 255},
		AmbientIntensity: 0.5,
		MaxLights:        32,
		GammaCorrection:  0, // Disable for simpler testing
	}
	system := NewSystemWithConfig(config)

	// Create a white image
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{200, 200, 200, 255})
		}
	}

	result := system.ApplyLighting(img)

	// All pixels should be dimmed equally
	firstColor := result.At(0, 0)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if result.At(x, y) != firstColor {
				t.Errorf("Pixel (%d, %d) differs with ambient-only lighting", x, y)
			}
		}
	}
}

// TestSystem_ApplyLighting_MultipleLight tests multiple lights.
func TestSystem_ApplyLighting_MultipleLights(t *testing.T) {
	system := NewSystem()

	// Create a simple image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{100, 100, 100, 255})
		}
	}

	// Add two point lights
	light1 := Light{
		Type:      TypePoint,
		Position:  image.Point{X: 25, Y: 50},
		Color:     color.RGBA{255, 100, 100, 255},
		Intensity: 1.0,
		Radius:    30,
		Falloff:   FalloffLinear,
		Enabled:   true,
	}
	light2 := Light{
		Type:      TypePoint,
		Position:  image.Point{X: 75, Y: 50},
		Color:     color.RGBA{100, 100, 255, 255},
		Intensity: 1.0,
		Radius:    30,
		Falloff:   FalloffLinear,
		Enabled:   true,
	}

	_ = system.AddLight(light1)
	_ = system.AddLight(light2)

	result := system.ApplyLighting(img)

	if result == nil {
		t.Fatal("ApplyLighting returned nil")
	}

	// Both light positions should be bright
	color1 := result.At(25, 50)
	color2 := result.At(75, 50)
	darkColor := result.At(50, 10)

	r1, _, _, _ := color1.RGBA()
	r2, _, _, _ := color2.RGBA()
	rd, _, _, _ := darkColor.RGBA()

	if r1 <= rd || r2 <= rd {
		t.Error("Light positions should be brighter than dark areas")
	}
}

// TestSystem_FalloffTypes tests different falloff types.
func TestSystem_FalloffTypes(t *testing.T) {
	falloffs := []FalloffType{FalloffNone, FalloffLinear, FalloffQuadratic, FalloffInverseSquare}

	for _, falloff := range falloffs {
		t.Run(falloff.String(), func(t *testing.T) {
			system := NewSystem()

			img := image.NewRGBA(image.Rect(0, 0, 100, 100))
			for y := 0; y < 100; y++ {
				for x := 0; x < 100; x++ {
					img.Set(x, y, color.RGBA{100, 100, 100, 255})
				}
			}

			light := Light{
				Type:      TypePoint,
				Position:  image.Point{X: 50, Y: 50},
				Color:     color.RGBA{255, 255, 255, 255},
				Intensity: 1.0,
				Radius:    40,
				Falloff:   falloff,
				Enabled:   true,
			}
			_ = system.AddLight(light)

			result := system.ApplyLighting(img)

			if result == nil {
				t.Errorf("ApplyLighting returned nil for falloff %v", falloff)
			}
		})
	}
}

// TestSystem_ApplyLightingToRegion tests regional lighting.
func TestSystem_ApplyLightingToRegion(t *testing.T) {
	system := NewSystem()

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	// Fill with gray
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{128, 128, 128, 255})
		}
	}

	light := Light{
		Type:      TypePoint,
		Position:  image.Point{X: 50, Y: 50},
		Color:     color.RGBA{255, 255, 255, 255},
		Intensity: 2.0,
		Radius:    40,
		Falloff:   FalloffNone, // No falloff for consistent brightness
		Enabled:   true,
	}
	err := system.AddLight(light)
	if err != nil {
		t.Fatalf("Failed to add light: %v", err)
	}

	// Apply lighting only to center region
	region := image.Rect(25, 25, 75, 75)
	result := system.ApplyLightingToRegion(img, region)

	if result == nil {
		t.Fatal("ApplyLightingToRegion returned nil")
	}

	// Verify result has correct bounds
	if result.Bounds() != img.Bounds() {
		t.Errorf("Result bounds = %v, want %v", result.Bounds(), img.Bounds())
	}

	// Just verify the function completes without crashing
	// The actual lighting effect is tested in ApplyLighting tests
	t.Log("ApplyLightingToRegion completed successfully")
}

// BenchmarkSystem_ApplyLighting benchmarks lighting application.
func BenchmarkSystem_ApplyLighting(b *testing.B) {
	system := NewSystem()

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{200, 200, 200, 255})
		}
	}

	light := Light{
		Type:      TypePoint,
		Position:  image.Point{X: 50, Y: 50},
		Color:     color.RGBA{255, 255, 255, 255},
		Intensity: 1.0,
		Radius:    50,
		Falloff:   FalloffQuadratic,
		Enabled:   true,
	}
	_ = system.AddLight(light)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = system.ApplyLighting(img)
	}
}

// BenchmarkSystem_ApplyLighting_MultipleLights benchmarks with multiple lights.
func BenchmarkSystem_ApplyLighting_MultipleLights(b *testing.B) {
	system := NewSystem()

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{200, 200, 200, 255})
		}
	}

	// Add 4 lights
	positions := []image.Point{
		{X: 25, Y: 25},
		{X: 75, Y: 25},
		{X: 25, Y: 75},
		{X: 75, Y: 75},
	}

	for _, pos := range positions {
		light := Light{
			Type:      TypePoint,
			Position:  pos,
			Color:     color.RGBA{255, 200, 150, 255},
			Intensity: 1.0,
			Radius:    40,
			Falloff:   FalloffQuadratic,
			Enabled:   true,
		}
		_ = system.AddLight(light)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = system.ApplyLighting(img)
	}
}

// TestSystem_SetConfig tests setting system configuration.
func TestSystem_SetConfig(t *testing.T) {
	system := NewSystem()

	// Verify default config
	defaultConfig := system.GetConfig()
	if defaultConfig.MaxLights != 32 {
		t.Errorf("Default MaxLights = %d, want 32", defaultConfig.MaxLights)
	}

	// Create custom config
	customConfig := LightingConfig{
		MaxLights:        64,
		AmbientColor:     color.RGBA{R: 50, G: 50, B: 50, A: 255},
		AmbientIntensity: 0.5,
		GammaCorrection:  2.2,
		EnableShadows:    false,
		BloomConfig:      DefaultBloomConfig(),
		AOConfig:         DefaultEnhancedAOConfig(),
	}

	// Set config
	system.SetConfig(customConfig)

	// Verify config was updated
	newConfig := system.GetConfig()
	if newConfig.MaxLights != 64 {
		t.Errorf("Updated MaxLights = %d, want 64", newConfig.MaxLights)
	}
	if newConfig.AmbientIntensity != 0.5 {
		t.Errorf("Updated AmbientIntensity = %v, want 0.5", newConfig.AmbientIntensity)
	}

	// Check ambient color (need to convert to RGBA for comparison)
	ambientRGBA := color.RGBAModel.Convert(newConfig.AmbientColor).(color.RGBA)
	if ambientRGBA.R != 50 {
		t.Errorf("Updated AmbientColor.R = %d, want 50", ambientRGBA.R)
	}
}

// TestValidationError_Error tests ValidationError.Error() method.
func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		expected string
	}{
		{
			name:     "intensity error",
			err:      &ValidationError{Field: "intensity", Message: "must be non-negative"},
			expected: "lighting: intensity must be non-negative",
		},
		{
			name:     "radius error",
			err:      &ValidationError{Field: "radius", Message: "must be positive for point lights"},
			expected: "lighting: radius must be positive for point lights",
		},
		{
			name:     "index error",
			err:      &ValidationError{Field: "index", Message: "out of bounds"},
			expected: "lighting: index out of bounds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("ValidationError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestValidationError_Unwrap tests error unwrapping functionality.
func TestValidationError_Unwrap(t *testing.T) {
	tests := []struct {
		name        string
		err         *ValidationError
		wantWrapped error
	}{
		{
			name:        "no wrapped error",
			err:         &ValidationError{Field: "intensity", Message: "must be non-negative"},
			wantWrapped: nil,
		},
		{
			name: "with wrapped error",
			err: &ValidationError{
				Field:   "radius",
				Message: "validation failed",
				Err:     &ValidationError{Field: "inner", Message: "nested error"},
			},
			wantWrapped: &ValidationError{Field: "inner", Message: "nested error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Unwrap()
			if tt.wantWrapped == nil {
				if got != nil {
					t.Errorf("ValidationError.Unwrap() = %v, want nil", got)
				}
			} else {
				if got == nil {
					t.Errorf("ValidationError.Unwrap() = nil, want non-nil")
				} else if got.Error() != tt.wantWrapped.Error() {
					t.Errorf("ValidationError.Unwrap() = %q, want %q", got.Error(), tt.wantWrapped.Error())
				}
			}
		})
	}
}

// TestValidationError_ErrorWithWrapped tests error message with wrapped errors.
func TestValidationError_ErrorWithWrapped(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		expected string
	}{
		{
			name:     "no wrapped error",
			err:      &ValidationError{Field: "intensity", Message: "must be non-negative"},
			expected: "lighting: intensity must be non-negative",
		},
		{
			name: "with wrapped error",
			err: &ValidationError{
				Field:   "radius",
				Message: "validation failed",
				Err:     &ValidationError{Field: "inner", Message: "nested error"},
			},
			expected: "lighting: radius validation failed: lighting: inner nested error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("ValidationError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestSystem_DirectionalLight tests directional light calculation.
func TestSystem_DirectionalLight(t *testing.T) {
	system := NewSystem()

	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			img.Set(x, y, color.RGBA{100, 100, 100, 255})
		}
	}

	// Add a directional light (illuminates entire scene uniformly)
	directionalLight := Light{
		Type:      TypeDirectional,
		Position:  image.Point{X: 0, Y: 0}, // Position doesn't matter for directional
		Color:     color.RGBA{255, 200, 150, 255},
		Intensity: 1.5,
		Radius:    0, // Radius doesn't matter for directional
		Falloff:   FalloffNone,
		Enabled:   true,
	}

	err := system.AddLight(directionalLight)
	if err != nil {
		t.Fatalf("Failed to add directional light: %v", err)
	}

	// Apply lighting
	result := system.ApplyLighting(img)
	if result == nil {
		t.Fatal("ApplyLighting returned nil")
	}

	// Verify result has correct bounds
	if result.Bounds() != img.Bounds() {
		t.Errorf("Result bounds = %v, want %v", result.Bounds(), img.Bounds())
	}

	// Sample a few pixels to verify directional lighting was applied
	// Directional light should apply uniformly across the image
	centerColor := result.At(25, 25).(color.RGBA)
	cornerColor := result.At(5, 5).(color.RGBA)

	// Both should be brightened equally since directional light is uniform
	if centerColor.R < 100 || cornerColor.R < 100 {
		t.Errorf("Directional light failed to brighten pixels: center=%+v, corner=%+v",
			centerColor, cornerColor)
	}

	// Verify colors are similar (within tolerance due to calculations)
	colorDiff := func(c1, c2 uint8) int {
		diff := int(c1) - int(c2)
		if diff < 0 {
			return -diff
		}
		return diff
	}

	if colorDiff(centerColor.R, cornerColor.R) > 5 {
		t.Errorf("Directional light produced non-uniform results: center=%+v, corner=%+v",
			centerColor, cornerColor)
	}
}

// TestSystem_DirectionalLight_WithPointLights tests combining directional and point lights.
func TestSystem_DirectionalLight_WithPointLights(t *testing.T) {
	system := NewSystem()

	img := image.NewRGBA(image.Rect(0, 0, 60, 60))
	for y := 0; y < 60; y++ {
		for x := 0; x < 60; x++ {
			img.Set(x, y, color.RGBA{50, 50, 50, 255})
		}
	}

	// Add directional light for base illumination
	directional := Light{
		Type:      TypeDirectional,
		Color:     color.RGBA{100, 100, 100, 255},
		Intensity: 0.5,
		Enabled:   true,
	}
	if err := system.AddLight(directional); err != nil {
		t.Fatalf("Failed to add directional light: %v", err)
	}

	// Add point light for localized highlight
	point := Light{
		Type:      TypePoint,
		Position:  image.Point{X: 30, Y: 30},
		Color:     color.RGBA{255, 200, 100, 255},
		Intensity: 1.0,
		Radius:    20,
		Falloff:   FalloffLinear,
		Enabled:   true,
	}
	if err := system.AddLight(point); err != nil {
		t.Fatalf("Failed to add point light: %v", err)
	}

	// Apply lighting
	result := system.ApplyLighting(img)
	if result == nil {
		t.Fatal("ApplyLighting returned nil")
	}

	// Check that center (near point light) is brighter than edge
	centerColor := result.At(30, 30).(color.RGBA)
	edgeColor := result.At(5, 5).(color.RGBA)

	if centerColor.R <= edgeColor.R {
		t.Errorf("Point light did not create localized highlight: center=%+v, edge=%+v",
			centerColor, edgeColor)
	}
}

// TestNewSystemWithLogger_EnableShadowsDeprecationWarning asserts that
// NewSystemWithLogger (and therefore NewSystemWithConfig) emits a logrus warning
// when EnableShadows=true but AOConfig.Enabled=false, satisfying AUDIT.md G9.
func TestNewSystemWithLogger_EnableShadowsDeprecationWarning(t *testing.T) {
var warned bool

// Capture logrus output via a hook.
hook := &captureHook{level: logrus.WarnLevel, onWarn: func(entry *logrus.Entry) {
if entry.Level == logrus.WarnLevel {
if _, ok := entry.Data["field"]; ok {
warned = true
}
}
}}
logrus.AddHook(hook)
defer logrus.StandardLogger().ReplaceHooks(logrus.LevelHooks{})

cfg := LightingConfig{
EnableShadows: true,
AOConfig:      EnhancedAOConfig{AOConfig: AOConfig{Enabled: false}},
}
_ = NewSystemWithLogger(cfg, nil)

if !warned {
t.Error("expected deprecation warning for EnableShadows=true with AOConfig.Enabled=false")
}
}

// TestNewSystemWithLogger_NoWarnWhenAOEnabled asserts no deprecation warning when
// AOConfig.Enabled=true (the preferred migration path).
func TestNewSystemWithLogger_NoWarnWhenAOEnabled(t *testing.T) {
var warned bool

hook := &captureHook{level: logrus.WarnLevel, onWarn: func(entry *logrus.Entry) {
if entry.Level == logrus.WarnLevel {
if _, ok := entry.Data["field"]; ok {
warned = true
}
}
}}
logrus.AddHook(hook)
defer logrus.StandardLogger().ReplaceHooks(logrus.LevelHooks{})

cfg := LightingConfig{
EnableShadows: true,
AOConfig:      EnhancedAOConfig{AOConfig: AOConfig{Enabled: true}},
}
_ = NewSystemWithLogger(cfg, nil)

if warned {
t.Error("expected no deprecation warning when AOConfig.Enabled=true")
}
}

// captureHook is a minimal logrus hook for test-level log capture.
type captureHook struct {
level  logrus.Level
onWarn func(*logrus.Entry)
}

func (h *captureHook) Levels() []logrus.Level { return logrus.AllLevels }
func (h *captureHook) Fire(entry *logrus.Entry) error {
if h.onWarn != nil {
h.onWarn(entry)
}
return nil
}
