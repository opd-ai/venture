package rendering

import (
	"image/color"
	"testing"
)

func TestPalette_Creation(t *testing.T) {
	primary := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	secondary := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	background := color.RGBA{R: 0, G: 0, B: 255, A: 255}
	text := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	palette := Palette{
		Primary:    primary,
		Secondary:  secondary,
		Background: background,
		Text:       text,
		Colors:     []color.Color{primary, secondary},
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

	if len(palette.Colors) != 2 {
		t.Errorf("Palette.Colors length = %d, want 2", len(palette.Colors))
	}
}

func TestPalette_EmptyColors(t *testing.T) {
	palette := Palette{
		Primary:    color.RGBA{R: 255, G: 255, B: 255, A: 255},
		Secondary:  color.RGBA{R: 0, G: 0, B: 0, A: 255},
		Background: color.RGBA{R: 128, G: 128, B: 128, A: 255},
		Text:       color.RGBA{R: 255, G: 255, B: 255, A: 255},
		Colors:     []color.Color{},
	}

	if len(palette.Colors) != 0 {
		t.Errorf("Palette.Colors should be empty, got length %d", len(palette.Colors))
	}
}

func TestPalette_NilColors(t *testing.T) {
	palette := Palette{
		Primary:    color.RGBA{R: 255, G: 0, B: 0, A: 255},
		Secondary:  color.RGBA{R: 0, G: 255, B: 0, A: 255},
		Background: color.RGBA{R: 0, G: 0, B: 255, A: 255},
		Text:       color.RGBA{R: 255, G: 255, B: 255, A: 255},
		Colors:     nil,
	}

	if palette.Colors != nil {
		t.Errorf("Palette.Colors should be nil, got %v", palette.Colors)
	}
}

func TestPalette_MultipleColors(t *testing.T) {
	colors := []color.Color{
		color.RGBA{R: 255, G: 0, B: 0, A: 255},
		color.RGBA{R: 0, G: 255, B: 0, A: 255},
		color.RGBA{R: 0, G: 0, B: 255, A: 255},
		color.RGBA{R: 255, G: 255, B: 0, A: 255},
		color.RGBA{R: 255, G: 0, B: 255, A: 255},
	}

	palette := Palette{
		Primary:    colors[0],
		Secondary:  colors[1],
		Background: colors[2],
		Text:       color.White,
		Colors:     colors,
	}

	if len(palette.Colors) != 5 {
		t.Errorf("Palette.Colors length = %d, want 5", len(palette.Colors))
	}

	for i, c := range palette.Colors {
		if c != colors[i] {
			t.Errorf("Palette.Colors[%d] = %v, want %v", i, c, colors[i])
		}
	}
}

func TestSpriteConfig_Creation(t *testing.T) {
	palette := &Palette{
		Primary:   color.RGBA{R: 255, G: 0, B: 0, A: 255},
		Secondary: color.RGBA{R: 0, G: 255, B: 0, A: 255},
	}

	config := SpriteConfig{
		Width:   64,
		Height:  64,
		Seed:    12345,
		Palette: palette,
		Type:    "entity",
		Custom:  map[string]interface{}{"test": "value"},
	}

	if config.Width != 64 {
		t.Errorf("SpriteConfig.Width = %d, want 64", config.Width)
	}

	if config.Height != 64 {
		t.Errorf("SpriteConfig.Height = %d, want 64", config.Height)
	}

	if config.Seed != 12345 {
		t.Errorf("SpriteConfig.Seed = %d, want 12345", config.Seed)
	}

	if config.Type != "entity" {
		t.Errorf("SpriteConfig.Type = %s, want 'entity'", config.Type)
	}

	if config.Palette == nil {
		t.Error("SpriteConfig.Palette should not be nil")
	}

	if config.Custom == nil {
		t.Error("SpriteConfig.Custom should not be nil")
	}

	if val, ok := config.Custom["test"]; !ok || val != "value" {
		t.Errorf("SpriteConfig.Custom['test'] = %v, want 'value'", val)
	}
}

func TestSpriteConfig_ZeroDimensions(t *testing.T) {
	config := SpriteConfig{
		Width:  0,
		Height: 0,
	}

	if config.Width != 0 {
		t.Errorf("SpriteConfig.Width = %d, want 0", config.Width)
	}

	if config.Height != 0 {
		t.Errorf("SpriteConfig.Height = %d, want 0", config.Height)
	}
}

func TestSpriteConfig_LargeDimensions(t *testing.T) {
	config := SpriteConfig{
		Width:  1024,
		Height: 2048,
	}

	if config.Width != 1024 {
		t.Errorf("SpriteConfig.Width = %d, want 1024", config.Width)
	}

	if config.Height != 2048 {
		t.Errorf("SpriteConfig.Height = %d, want 2048", config.Height)
	}
}

func TestSpriteConfig_NegativeSeed(t *testing.T) {
	config := SpriteConfig{
		Seed: -12345,
	}

	if config.Seed != -12345 {
		t.Errorf("SpriteConfig.Seed = %d, want -12345", config.Seed)
	}
}

func TestSpriteConfig_DifferentTypes(t *testing.T) {
	types := []string{"entity", "item", "tile", "particle", "ui", "custom"}

	for _, spriteType := range types {
		t.Run(spriteType, func(t *testing.T) {
			config := SpriteConfig{
				Type: spriteType,
			}

			if config.Type != spriteType {
				t.Errorf("SpriteConfig.Type = %s, want %s", config.Type, spriteType)
			}
		})
	}
}

func TestSpriteConfig_CustomParameters(t *testing.T) {
	config := SpriteConfig{
		Custom: map[string]interface{}{
			"animation":  true,
			"frameCount": 4,
			"duration":   1.5,
			"name":       "test",
		},
	}

	if val, ok := config.Custom["animation"]; !ok || val != true {
		t.Errorf("Custom['animation'] = %v, want true", val)
	}

	if val, ok := config.Custom["frameCount"]; !ok || val != 4 {
		t.Errorf("Custom['frameCount'] = %v, want 4", val)
	}

	if val, ok := config.Custom["duration"]; !ok || val != 1.5 {
		t.Errorf("Custom['duration'] = %v, want 1.5", val)
	}

	if val, ok := config.Custom["name"]; !ok || val != "test" {
		t.Errorf("Custom['name'] = %v, want 'test'", val)
	}
}

func TestSpriteConfig_EmptyCustom(t *testing.T) {
	config := SpriteConfig{
		Custom: map[string]interface{}{},
	}

	if len(config.Custom) != 0 {
		t.Errorf("SpriteConfig.Custom should be empty, got length %d", len(config.Custom))
	}
}

func TestSpriteConfig_NilCustom(t *testing.T) {
	config := SpriteConfig{
		Custom: nil,
	}

	if config.Custom != nil {
		t.Errorf("SpriteConfig.Custom should be nil, got %v", config.Custom)
	}
}

func TestSpriteConfig_NilPalette(t *testing.T) {
	config := SpriteConfig{
		Width:   32,
		Height:  32,
		Palette: nil,
	}

	if config.Palette != nil {
		t.Errorf("SpriteConfig.Palette should be nil, got %v", config.Palette)
	}
}

func TestSpriteConfig_CompleteConfiguration(t *testing.T) {
	palette := &Palette{
		Primary:    color.RGBA{R: 255, G: 0, B: 0, A: 255},
		Secondary:  color.RGBA{R: 0, G: 255, B: 0, A: 255},
		Background: color.RGBA{R: 0, G: 0, B: 255, A: 255},
		Text:       color.RGBA{R: 255, G: 255, B: 255, A: 255},
	}

	config := SpriteConfig{
		Width:   128,
		Height:  128,
		Seed:    99999,
		Palette: palette,
		Type:    "boss",
		Custom: map[string]interface{}{
			"tier":     5,
			"animated": true,
			"rarity":   "legendary",
		},
	}

	if config.Width != 128 {
		t.Errorf("SpriteConfig.Width = %d, want 128", config.Width)
	}

	if config.Height != 128 {
		t.Errorf("SpriteConfig.Height = %d, want 128", config.Height)
	}

	if config.Seed != 99999 {
		t.Errorf("SpriteConfig.Seed = %d, want 99999", config.Seed)
	}

	if config.Type != "boss" {
		t.Errorf("SpriteConfig.Type = %s, want 'boss'", config.Type)
	}

	if config.Palette == nil {
		t.Fatal("SpriteConfig.Palette should not be nil")
	}

	if config.Palette.Primary != palette.Primary {
		t.Errorf("SpriteConfig.Palette.Primary = %v, want %v", config.Palette.Primary, palette.Primary)
	}

	if len(config.Custom) != 3 {
		t.Errorf("SpriteConfig.Custom length = %d, want 3", len(config.Custom))
	}
}

func TestPalette_ColorVariety(t *testing.T) {
	// Test with different color combinations
	tests := []struct {
		name    string
		palette Palette
	}{
		{
			name: "Monochrome",
			palette: Palette{
				Primary:    color.RGBA{R: 255, G: 255, B: 255, A: 255},
				Secondary:  color.RGBA{R: 200, G: 200, B: 200, A: 255},
				Background: color.RGBA{R: 100, G: 100, B: 100, A: 255},
				Text:       color.RGBA{R: 0, G: 0, B: 0, A: 255},
			},
		},
		{
			name: "HighContrast",
			palette: Palette{
				Primary:    color.RGBA{R: 255, G: 255, B: 255, A: 255},
				Secondary:  color.RGBA{R: 0, G: 0, B: 0, A: 255},
				Background: color.RGBA{R: 0, G: 0, B: 0, A: 255},
				Text:       color.RGBA{R: 255, G: 255, B: 255, A: 255},
			},
		},
		{
			name: "Colorful",
			palette: Palette{
				Primary:    color.RGBA{R: 255, G: 0, B: 0, A: 255},
				Secondary:  color.RGBA{R: 0, G: 255, B: 0, A: 255},
				Background: color.RGBA{R: 0, G: 0, B: 255, A: 255},
				Text:       color.RGBA{R: 255, G: 255, B: 0, A: 255},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.palette.Primary == nil {
				t.Error("Palette.Primary should not be nil")
			}
			if tt.palette.Secondary == nil {
				t.Error("Palette.Secondary should not be nil")
			}
			if tt.palette.Background == nil {
				t.Error("Palette.Background should not be nil")
			}
			if tt.palette.Text == nil {
				t.Error("Palette.Text should not be nil")
			}
		})
	}
}
