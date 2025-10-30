package engine

import (
	"image/color"
	"testing"
)

func TestLightFalloffType_String(t *testing.T) {
	tests := []struct {
		name     string
		falloff  LightFalloffType
		expected string
	}{
		{"linear", FalloffLinear, "linear"},
		{"quadratic", FalloffQuadratic, "quadratic"},
		{"inverse square", FalloffInverseSquare, "inverse_square"},
		{"constant", FalloffConstant, "constant"},
		{"unknown", LightFalloffType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.falloff.String()
			if got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewLightComponent(t *testing.T) {
	tests := []struct {
		name      string
		radius    float64
		color     color.RGBA
		intensity float64
		wantRad   float64
		wantInt   float64
	}{
		{
			name:      "valid params",
			radius:    150,
			color:     color.RGBA{255, 255, 255, 255},
			intensity: 0.8,
			wantRad:   150,
			wantInt:   0.8,
		},
		{
			name:      "zero radius defaults to 200",
			radius:    0,
			color:     color.RGBA{255, 255, 255, 255},
			intensity: 1.0,
			wantRad:   200,
			wantInt:   1.0,
		},
		{
			name:      "negative radius defaults to 200",
			radius:    -50,
			color:     color.RGBA{255, 255, 255, 255},
			intensity: 1.0,
			wantRad:   200,
			wantInt:   1.0,
		},
		{
			name:      "zero intensity defaults to 1.0",
			radius:    100,
			color:     color.RGBA{255, 255, 255, 255},
			intensity: 0,
			wantRad:   100,
			wantInt:   1.0,
		},
		{
			name:      "negative intensity defaults to 1.0",
			radius:    100,
			color:     color.RGBA{255, 255, 255, 255},
			intensity: -0.5,
			wantRad:   100,
			wantInt:   1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			light := NewLightComponent(tt.radius, tt.color, tt.intensity)

			if light.Radius != tt.wantRad {
				t.Errorf("Radius = %v, want %v", light.Radius, tt.wantRad)
			}
			if light.Intensity != tt.wantInt {
				t.Errorf("Intensity = %v, want %v", light.Intensity, tt.wantInt)
			}
			if light.Color != tt.color {
				t.Errorf("Color = %v, want %v", light.Color, tt.color)
			}
			if !light.Enabled {
				t.Error("Light should be enabled by default")
			}
			if light.Falloff != FalloffQuadratic {
				t.Errorf("Falloff = %v, want FalloffQuadratic", light.Falloff)
			}
		})
	}
}

func TestLightComponent_Type(t *testing.T) {
	light := NewLightComponent(100, color.RGBA{255, 255, 255, 255}, 1.0)
	if light.Type() != "light" {
		t.Errorf("Type() = %v, want 'light'", light.Type())
	}
}

func TestNewTorchLight(t *testing.T) {
	light := NewTorchLight(150)

	if light.Radius != 150 {
		t.Errorf("Radius = %v, want 150", light.Radius)
	}
	if !light.Flickering {
		t.Error("Torch light should flicker")
	}
	if light.FlickerSpeed <= 0 {
		t.Error("Flicker speed should be positive")
	}
	if light.FlickerAmount <= 0 {
		t.Error("Flicker amount should be positive")
	}
	// Check warm orange color (approximate)
	if light.Color.R < 200 || light.Color.G < 100 || light.Color.B > 150 {
		t.Errorf("Torch color %v doesn't look warm orange", light.Color)
	}
}

func TestNewSpellLight(t *testing.T) {
	spellColor := color.RGBA{0, 100, 255, 255}
	light := NewSpellLight(80, spellColor)

	if light.Radius != 80 {
		t.Errorf("Radius = %v, want 80", light.Radius)
	}
	if light.Color != spellColor {
		t.Errorf("Color = %v, want %v", light.Color, spellColor)
	}
	if !light.Pulsing {
		t.Error("Spell light should pulse")
	}
	if light.Falloff != FalloffLinear {
		t.Errorf("Spell light should use linear falloff, got %v", light.Falloff)
	}
}

func TestNewCrystalLight(t *testing.T) {
	crystalColor := color.RGBA{255, 0, 255, 255}
	light := NewCrystalLight(120, crystalColor)

	if light.Radius != 120 {
		t.Errorf("Radius = %v, want 120", light.Radius)
	}
	if light.Color != crystalColor {
		t.Errorf("Color = %v, want %v", light.Color, crystalColor)
	}
	if !light.Pulsing {
		t.Error("Crystal light should pulse")
	}
}

func TestLightComponent_GetCurrentIntensity(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*LightComponent)
		checkRange bool
		minVal     float64
		maxVal     float64
	}{
		{
			name: "disabled light returns 0",
			setup: func(l *LightComponent) {
				l.Enabled = false
				l.Intensity = 1.0
			},
			checkRange: false,
			minVal:     0,
			maxVal:     0,
		},
		{
			name: "enabled light without effects returns base intensity",
			setup: func(l *LightComponent) {
				l.Enabled = true
				l.Intensity = 0.8
				l.Flickering = false
				l.Pulsing = false
			},
			checkRange: false,
			minVal:     0.8,
			maxVal:     0.8,
		},
		{
			name: "flickering light varies intensity",
			setup: func(l *LightComponent) {
				l.Enabled = true
				l.Intensity = 1.0
				l.Flickering = true
				l.FlickerAmount = 0.2
				l.internalTime = 0.5
			},
			checkRange: true,
			minVal:     0.8,
			maxVal:     1.0,
		},
		{
			name: "pulsing light varies intensity",
			setup: func(l *LightComponent) {
				l.Enabled = true
				l.Intensity = 1.0
				l.Pulsing = true
				l.PulseAmount = 0.3
				l.internalTime = 0.25
			},
			checkRange: true,
			minVal:     0.7,
			maxVal:     1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			light := NewLightComponent(100, color.RGBA{255, 255, 255, 255}, 1.0)
			tt.setup(light)

			intensity := light.GetCurrentIntensity()

			if tt.checkRange {
				if intensity < tt.minVal || intensity > tt.maxVal {
					t.Errorf("GetCurrentIntensity() = %v, want range [%v, %v]", intensity, tt.minVal, tt.maxVal)
				}
			} else {
				if intensity != tt.minVal {
					t.Errorf("GetCurrentIntensity() = %v, want %v", intensity, tt.minVal)
				}
			}
		})
	}
}

func TestAmbientLightComponent_Type(t *testing.T) {
	ambient := NewAmbientLightComponent(color.RGBA{100, 100, 100, 255}, 0.5)
	if ambient.Type() != "ambient_light" {
		t.Errorf("Type() = %v, want 'ambient_light'", ambient.Type())
	}
}

func TestNewAmbientLightComponent(t *testing.T) {
	tests := []struct {
		name      string
		color     color.RGBA
		intensity float64
		wantInt   float64
	}{
		{
			name:      "valid intensity",
			color:     color.RGBA{100, 100, 100, 255},
			intensity: 0.5,
			wantInt:   0.5,
		},
		{
			name:      "intensity clamped to 0",
			color:     color.RGBA{100, 100, 100, 255},
			intensity: -0.5,
			wantInt:   0,
		},
		{
			name:      "intensity clamped to 1",
			color:     color.RGBA{100, 100, 100, 255},
			intensity: 1.5,
			wantInt:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ambient := NewAmbientLightComponent(tt.color, tt.intensity)

			if ambient.Intensity != tt.wantInt {
				t.Errorf("Intensity = %v, want %v", ambient.Intensity, tt.wantInt)
			}
			if ambient.Color != tt.color {
				t.Errorf("Color = %v, want %v", ambient.Color, tt.color)
			}
		})
	}
}

func TestNewLightingConfig(t *testing.T) {
	config := NewLightingConfig()

	if !config.Enabled {
		t.Error("Lighting should be enabled by default")
	}
	if config.MaxLights != 16 {
		t.Errorf("MaxLights = %v, want 16", config.MaxLights)
	}
	if !config.GammaCorrection {
		t.Error("Gamma correction should be enabled by default")
	}
	if config.Gamma != 2.2 {
		t.Errorf("Gamma = %v, want 2.2", config.Gamma)
	}
	if config.AmbientIntensity <= 0 || config.AmbientIntensity > 1 {
		t.Errorf("AmbientIntensity = %v, should be in (0, 1]", config.AmbientIntensity)
	}
}

func TestLightingConfig_SetGenrePreset(t *testing.T) {
	tests := []struct {
		name         string
		genreID      string
		wantMinAmb   float64
		wantMaxAmb   float64
		checkColor   bool
		wantRedRange [2]uint8
	}{
		{
			name:         "fantasy has warm tone",
			genreID:      "fantasy",
			wantMinAmb:   0.35,
			wantMaxAmb:   0.45,
			checkColor:   true,
			wantRedRange: [2]uint8{100, 140},
		},
		{
			name:         "sci-fi has cool tone",
			genreID:      "sci-fi",
			wantMinAmb:   0.3,
			wantMaxAmb:   0.4,
			checkColor:   true,
			wantRedRange: [2]uint8{80, 100},
		},
		{
			name:         "horror is very dark",
			genreID:      "horror",
			wantMinAmb:   0.1,
			wantMaxAmb:   0.2,
			checkColor:   true,
			wantRedRange: [2]uint8{70, 90},
		},
		{
			name:         "cyberpunk has purple tint",
			genreID:      "cyberpunk",
			wantMinAmb:   0.2,
			wantMaxAmb:   0.3,
			checkColor:   true,
			wantRedRange: [2]uint8{90, 110},
		},
		{
			name:         "post-apocalyptic is dusty",
			genreID:      "post-apocalyptic",
			wantMinAmb:   0.25,
			wantMaxAmb:   0.35,
			checkColor:   true,
			wantRedRange: [2]uint8{120, 140},
		},
		{
			name:       "unknown genre uses default",
			genreID:    "unknown",
			wantMinAmb: 0.25,
			wantMaxAmb: 0.35,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewLightingConfig()
			config.SetGenrePreset(tt.genreID)

			if config.AmbientIntensity < tt.wantMinAmb || config.AmbientIntensity > tt.wantMaxAmb {
				t.Errorf("AmbientIntensity = %v, want range [%v, %v]", config.AmbientIntensity, tt.wantMinAmb, tt.wantMaxAmb)
			}

			if tt.checkColor {
				if config.AmbientColor.R < tt.wantRedRange[0] || config.AmbientColor.R > tt.wantRedRange[1] {
					t.Errorf("AmbientColor.R = %v, want range [%v, %v]", config.AmbientColor.R, tt.wantRedRange[0], tt.wantRedRange[1])
				}
			}
		})
	}
}

func TestLightComponent_fastSin(t *testing.T) {
	light := NewLightComponent(100, color.RGBA{255, 255, 255, 255}, 1.0)

	tests := []struct {
		name      string
		input     float64
		wantRange [2]float64
	}{
		{"zero", 0, [2]float64{-0.1, 0.1}},
		{"pi/2 ≈ 1", 1.57, [2]float64{0.9, 1.1}},
		{"pi ≈ 0", 3.14, [2]float64{-0.1, 0.1}},
		{"3pi/2 ≈ -1", 4.71, [2]float64{-1.1, -0.9}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := light.fastSin(tt.input)
			if result < tt.wantRange[0] || result > tt.wantRange[1] {
				t.Errorf("fastSin(%v) = %v, want range [%v, %v]", tt.input, result, tt.wantRange[0], tt.wantRange[1])
			}
		})
	}
}
