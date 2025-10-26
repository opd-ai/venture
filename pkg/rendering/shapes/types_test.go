package shapes

import (
	"image/color"
	"testing"
)

func TestShapeType_String(t *testing.T) {
	tests := []struct {
		name      string
		shapeType ShapeType
		want      string
	}{
		{"circle", ShapeCircle, "circle"},
		{"rectangle", ShapeRectangle, "rectangle"},
		{"triangle", ShapeTriangle, "triangle"},
		{"polygon", ShapePolygon, "polygon"},
		{"star", ShapeStar, "star"},
		{"ring", ShapeRing, "ring"},
		{"hexagon", ShapeHexagon, "hexagon"},
		{"octagon", ShapeOctagon, "octagon"},
		{"cross", ShapeCross, "cross"},
		{"heart", ShapeHeart, "heart"},
		{"crescent", ShapeCrescent, "crescent"},
		{"gear", ShapeGear, "gear"},
		{"crystal", ShapeCrystal, "crystal"},
		{"lightning", ShapeLightning, "lightning"},
		{"wave", ShapeWave, "wave"},
		{"spiral", ShapeSpiral, "spiral"},
		{"organic", ShapeOrganic, "organic"},
		{"unknown", ShapeType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.shapeType.String()
			if got != tt.want {
				t.Errorf("ShapeType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Type != ShapeCircle {
		t.Errorf("DefaultConfig Type = %v, want %v", config.Type, ShapeCircle)
	}
	if config.Width != 32 {
		t.Errorf("DefaultConfig Width = %v, want 32", config.Width)
	}
	if config.Height != 32 {
		t.Errorf("DefaultConfig Height = %v, want 32", config.Height)
	}
	if config.Color == nil {
		t.Error("DefaultConfig Color is nil")
	}
	if config.Sides != 5 {
		t.Errorf("DefaultConfig Sides = %v, want 5", config.Sides)
	}
	if config.InnerRatio != 0.5 {
		t.Errorf("DefaultConfig InnerRatio = %v, want 0.5", config.InnerRatio)
	}
	if config.Rotation != 0 {
		t.Errorf("DefaultConfig Rotation = %v, want 0", config.Rotation)
	}
	if config.Smoothing != 0.1 {
		t.Errorf("DefaultConfig Smoothing = %v, want 0.1", config.Smoothing)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "circle config",
			config: Config{
				Type:   ShapeCircle,
				Width:  32,
				Height: 32,
				Color:  color.RGBA{R: 255, G: 0, B: 0, A: 255},
			},
		},
		{
			name: "rectangle config",
			config: Config{
				Type:   ShapeRectangle,
				Width:  64,
				Height: 32,
				Color:  color.RGBA{R: 0, G: 255, B: 0, A: 255},
			},
		},
		{
			name: "triangle config",
			config: Config{
				Type:     ShapeTriangle,
				Width:    48,
				Height:   48,
				Color:    color.RGBA{R: 0, G: 0, B: 255, A: 255},
				Rotation: 45,
			},
		},
		{
			name: "polygon config",
			config: Config{
				Type:   ShapePolygon,
				Width:  40,
				Height: 40,
				Color:  color.RGBA{R: 255, G: 255, B: 0, A: 255},
				Sides:  6,
			},
		},
		{
			name: "star config",
			config: Config{
				Type:       ShapeStar,
				Width:      36,
				Height:     36,
				Color:      color.RGBA{R: 255, G: 0, B: 255, A: 255},
				Sides:      5,
				InnerRatio: 0.4,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the config can be created without panicking
			if tt.config.Type < 0 {
				t.Errorf("Invalid shape type: %v", tt.config.Type)
			}
			if tt.config.Width <= 0 || tt.config.Height <= 0 {
				t.Errorf("Invalid dimensions: %dx%d", tt.config.Width, tt.config.Height)
			}
		})
	}
}
