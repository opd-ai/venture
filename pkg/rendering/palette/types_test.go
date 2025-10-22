package palette

import (
	"image/color"
	"testing"
)

func TestPalette_Creation(t *testing.T) {
	primary := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	secondary := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	background := color.RGBA{R: 0, G: 0, B: 255, A: 255}
	text := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	accent1 := color.RGBA{R: 128, G: 128, B: 128, A: 255}
	accent2 := color.RGBA{R: 64, G: 64, B: 64, A: 255}
	danger := color.RGBA{R: 200, G: 0, B: 0, A: 255}
	success := color.RGBA{R: 0, G: 200, B: 0, A: 255}

	palette := Palette{
		Primary:    primary,
		Secondary:  secondary,
		Background: background,
		Text:       text,
		Accent1:    accent1,
		Accent2:    accent2,
		Danger:     danger,
		Success:    success,
		Colors:     []color.Color{primary, secondary, background},
	}

	if palette.Primary != primary {
		t.Errorf("Palette.Primary = %v, want %v", palette.Primary, primary)
	}

	if palette.Secondary != secondary {
		t.Errorf("Palette.Secondary = %v, want %v", palette.Secondary, secondary)
	}

	if palette.Background != background {
		t.Errorf("Palette.Background = %v, want %v", palette.Background, background)
	}

	if palette.Text != text {
		t.Errorf("Palette.Text = %v, want %v", palette.Text, text)
	}

	if palette.Accent1 != accent1 {
		t.Errorf("Palette.Accent1 = %v, want %v", palette.Accent1, accent1)
	}

	if palette.Accent2 != accent2 {
		t.Errorf("Palette.Accent2 = %v, want %v", palette.Accent2, accent2)
	}

	if palette.Danger != danger {
		t.Errorf("Palette.Danger = %v, want %v", palette.Danger, danger)
	}

	if palette.Success != success {
		t.Errorf("Palette.Success = %v, want %v", palette.Success, success)
	}

	if len(palette.Colors) != 3 {
		t.Errorf("Palette.Colors length = %d, want 3", len(palette.Colors))
	}
}

func TestPalette_EmptyColors(t *testing.T) {
	palette := Palette{
		Primary:   color.RGBA{R: 255, G: 255, B: 255, A: 255},
		Secondary: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		Colors:    []color.Color{},
	}

	if len(palette.Colors) != 0 {
		t.Errorf("Palette.Colors should be empty, got length %d", len(palette.Colors))
	}
}

func TestPalette_MultipleColors(t *testing.T) {
	colors := []color.Color{
		color.RGBA{R: 255, G: 0, B: 0, A: 255},
		color.RGBA{R: 0, G: 255, B: 0, A: 255},
		color.RGBA{R: 0, G: 0, B: 255, A: 255},
		color.RGBA{R: 255, G: 255, B: 0, A: 255},
		color.RGBA{R: 255, G: 0, B: 255, A: 255},
		color.RGBA{R: 0, G: 255, B: 255, A: 255},
	}

	palette := Palette{
		Colors: colors,
	}

	if len(palette.Colors) != 6 {
		t.Errorf("Palette.Colors length = %d, want 6", len(palette.Colors))
	}

	for i, c := range palette.Colors {
		if c != colors[i] {
			t.Errorf("Palette.Colors[%d] = %v, want %v", i, c, colors[i])
		}
	}
}

func TestColorScheme_Creation(t *testing.T) {
	scheme := ColorScheme{
		BaseHue:             180.0,
		Saturation:          0.7,
		Lightness:           0.5,
		HueVariation:        30.0,
		SaturationVariation: 0.2,
		LightnessVariation:  0.1,
	}

	if scheme.BaseHue != 180.0 {
		t.Errorf("ColorScheme.BaseHue = %f, want 180.0", scheme.BaseHue)
	}

	if scheme.Saturation != 0.7 {
		t.Errorf("ColorScheme.Saturation = %f, want 0.7", scheme.Saturation)
	}

	if scheme.Lightness != 0.5 {
		t.Errorf("ColorScheme.Lightness = %f, want 0.5", scheme.Lightness)
	}

	if scheme.HueVariation != 30.0 {
		t.Errorf("ColorScheme.HueVariation = %f, want 30.0", scheme.HueVariation)
	}

	if scheme.SaturationVariation != 0.2 {
		t.Errorf("ColorScheme.SaturationVariation = %f, want 0.2", scheme.SaturationVariation)
	}

	if scheme.LightnessVariation != 0.1 {
		t.Errorf("ColorScheme.LightnessVariation = %f, want 0.1", scheme.LightnessVariation)
	}
}

func TestColorScheme_HueRange(t *testing.T) {
	tests := []struct {
		name  string
		hue   float64
		valid bool
	}{
		{"MinHue", 0.0, true},
		{"MaxHue", 360.0, true},
		{"MidHue", 180.0, true},
		{"RedHue", 0.0, true},
		{"GreenHue", 120.0, true},
		{"BlueHue", 240.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := ColorScheme{
				BaseHue: tt.hue,
			}

			if scheme.BaseHue != tt.hue {
				t.Errorf("ColorScheme.BaseHue = %f, want %f", scheme.BaseHue, tt.hue)
			}
		})
	}
}

func TestColorScheme_SaturationRange(t *testing.T) {
	tests := []struct {
		name       string
		saturation float64
	}{
		{"NoSaturation", 0.0},
		{"LowSaturation", 0.25},
		{"MediumSaturation", 0.5},
		{"HighSaturation", 0.75},
		{"FullSaturation", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := ColorScheme{
				Saturation: tt.saturation,
			}

			if scheme.Saturation != tt.saturation {
				t.Errorf("ColorScheme.Saturation = %f, want %f", scheme.Saturation, tt.saturation)
			}
		})
	}
}

func TestColorScheme_LightnessRange(t *testing.T) {
	tests := []struct {
		name      string
		lightness float64
	}{
		{"Black", 0.0},
		{"Dark", 0.25},
		{"Medium", 0.5},
		{"Light", 0.75},
		{"White", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := ColorScheme{
				Lightness: tt.lightness,
			}

			if scheme.Lightness != tt.lightness {
				t.Errorf("ColorScheme.Lightness = %f, want %f", scheme.Lightness, tt.lightness)
			}
		})
	}
}

func TestColorScheme_Variations(t *testing.T) {
	scheme := ColorScheme{
		BaseHue:             180.0,
		Saturation:          0.5,
		Lightness:           0.5,
		HueVariation:        45.0,
		SaturationVariation: 0.3,
		LightnessVariation:  0.2,
	}

	if scheme.HueVariation != 45.0 {
		t.Errorf("ColorScheme.HueVariation = %f, want 45.0", scheme.HueVariation)
	}

	if scheme.SaturationVariation != 0.3 {
		t.Errorf("ColorScheme.SaturationVariation = %f, want 0.3", scheme.SaturationVariation)
	}

	if scheme.LightnessVariation != 0.2 {
		t.Errorf("ColorScheme.LightnessVariation = %f, want 0.2", scheme.LightnessVariation)
	}
}

func TestPalette_UIColors(t *testing.T) {
	danger := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	success := color.RGBA{R: 0, G: 255, B: 0, A: 255}

	palette := Palette{
		Danger:  danger,
		Success: success,
	}

	if palette.Danger != danger {
		t.Errorf("Palette.Danger = %v, want %v", palette.Danger, danger)
	}

	if palette.Success != success {
		t.Errorf("Palette.Success = %v, want %v", palette.Success, success)
	}
}

func TestPalette_AccentColors(t *testing.T) {
	accent1 := color.RGBA{R: 255, G: 200, B: 100, A: 255}
	accent2 := color.RGBA{R: 100, G: 200, B: 255, A: 255}

	palette := Palette{
		Accent1: accent1,
		Accent2: accent2,
	}

	if palette.Accent1 != accent1 {
		t.Errorf("Palette.Accent1 = %v, want %v", palette.Accent1, accent1)
	}

	if palette.Accent2 != accent2 {
		t.Errorf("Palette.Accent2 = %v, want %v", palette.Accent2, accent2)
	}
}

func TestPalette_NilColors(t *testing.T) {
	palette := Palette{
		Colors: nil,
	}

	if palette.Colors != nil {
		t.Errorf("Palette.Colors should be nil, got %v", palette.Colors)
	}
}

func TestColorScheme_ZeroValues(t *testing.T) {
	scheme := ColorScheme{}

	if scheme.BaseHue != 0.0 {
		t.Errorf("ColorScheme.BaseHue = %f, want 0.0", scheme.BaseHue)
	}

	if scheme.Saturation != 0.0 {
		t.Errorf("ColorScheme.Saturation = %f, want 0.0", scheme.Saturation)
	}

	if scheme.Lightness != 0.0 {
		t.Errorf("ColorScheme.Lightness = %f, want 0.0", scheme.Lightness)
	}

	if scheme.HueVariation != 0.0 {
		t.Errorf("ColorScheme.HueVariation = %f, want 0.0", scheme.HueVariation)
	}

	if scheme.SaturationVariation != 0.0 {
		t.Errorf("ColorScheme.SaturationVariation = %f, want 0.0", scheme.SaturationVariation)
	}

	if scheme.LightnessVariation != 0.0 {
		t.Errorf("ColorScheme.LightnessVariation = %f, want 0.0", scheme.LightnessVariation)
	}
}

func TestPalette_AllFieldsSet(t *testing.T) {
	primary := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	secondary := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	background := color.RGBA{R: 0, G: 0, B: 255, A: 255}
	text := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	accent1 := color.RGBA{R: 128, G: 0, B: 128, A: 255}
	accent2 := color.RGBA{R: 128, G: 128, B: 0, A: 255}
	danger := color.RGBA{R: 200, G: 0, B: 0, A: 255}
	success := color.RGBA{R: 0, G: 200, B: 0, A: 255}
	colors := []color.Color{primary, secondary}

	palette := Palette{
		Primary:    primary,
		Secondary:  secondary,
		Background: background,
		Text:       text,
		Accent1:    accent1,
		Accent2:    accent2,
		Danger:     danger,
		Success:    success,
		Colors:     colors,
	}

	// Verify all fields are set correctly
	if palette.Primary == nil {
		t.Error("Palette.Primary should not be nil")
	}
	if palette.Secondary == nil {
		t.Error("Palette.Secondary should not be nil")
	}
	if palette.Background == nil {
		t.Error("Palette.Background should not be nil")
	}
	if palette.Text == nil {
		t.Error("Palette.Text should not be nil")
	}
	if palette.Accent1 == nil {
		t.Error("Palette.Accent1 should not be nil")
	}
	if palette.Accent2 == nil {
		t.Error("Palette.Accent2 should not be nil")
	}
	if palette.Danger == nil {
		t.Error("Palette.Danger should not be nil")
	}
	if palette.Success == nil {
		t.Error("Palette.Success should not be nil")
	}
	if palette.Colors == nil {
		t.Error("Palette.Colors should not be nil")
	}
}

func TestColorScheme_FullRange(t *testing.T) {
	scheme := ColorScheme{
		BaseHue:             270.0,
		Saturation:          0.8,
		Lightness:           0.6,
		HueVariation:        90.0,
		SaturationVariation: 0.4,
		LightnessVariation:  0.3,
	}

	// Just verify all fields can be set
	if scheme.BaseHue <= 0 || scheme.BaseHue >= 360 {
		// This is actually valid, but checking it's set
		if scheme.BaseHue != 270.0 {
			t.Errorf("ColorScheme.BaseHue = %f, want 270.0", scheme.BaseHue)
		}
	}
}
