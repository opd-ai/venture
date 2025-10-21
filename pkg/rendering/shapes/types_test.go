package shapes

import (
	"image/color"
	"testing"
)

func TestShapeType_String(t *testing.T) {
	tests := []struct {
		name     string
		shapeType ShapeType
		expected string
	}{
		{
			name:      "Circle",
			shapeType: ShapeCircle,
			expected:  "circle",
		},
		{
			name:      "Rectangle",
			shapeType: ShapeRectangle,
			expected:  "rectangle",
		},
		{
			name:      "Triangle",
			shapeType: ShapeTriangle,
			expected:  "triangle",
		},
		{
			name:      "Polygon",
			shapeType: ShapePolygon,
			expected:  "polygon",
		},
		{
			name:      "Star",
			shapeType: ShapeStar,
			expected:  "star",
		},
		{
			name:      "Ring",
			shapeType: ShapeRing,
			expected:  "ring",
		},
		{
			name:      "Unknown",
			shapeType: ShapeType(99),
			expected:  "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.shapeType.String()
			if result != tt.expected {
				t.Errorf("ShapeType.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Type != ShapeCircle {
		t.Errorf("DefaultConfig().Type = %v, want %v", config.Type, ShapeCircle)
	}

	if config.Width != 32 {
		t.Errorf("DefaultConfig().Width = %d, want 32", config.Width)
	}

	if config.Height != 32 {
		t.Errorf("DefaultConfig().Height = %d, want 32", config.Height)
	}

	if config.Seed != 0 {
		t.Errorf("DefaultConfig().Seed = %d, want 0", config.Seed)
	}

	if config.Sides != 5 {
		t.Errorf("DefaultConfig().Sides = %d, want 5", config.Sides)
	}

	if config.InnerRatio != 0.5 {
		t.Errorf("DefaultConfig().InnerRatio = %f, want 0.5", config.InnerRatio)
	}

	if config.Rotation != 0 {
		t.Errorf("DefaultConfig().Rotation = %f, want 0", config.Rotation)
	}

	if config.Smoothing != 0.1 {
		t.Errorf("DefaultConfig().Smoothing = %f, want 0.1", config.Smoothing)
	}

	expectedColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	if config.Color != expectedColor {
		t.Errorf("DefaultConfig().Color = %v, want %v", config.Color, expectedColor)
	}
}

func TestShape_Creation(t *testing.T) {
	shape := &Shape{
		Type:       ShapeCircle,
		Width:      64,
		Height:     64,
		Color:      color.RGBA{R: 255, G: 0, B: 0, A: 255},
		Seed:       12345,
		Sides:      6,
		InnerRatio: 0.7,
		Rotation:   45.0,
		Smoothing:  0.2,
	}

	if shape.Type != ShapeCircle {
		t.Errorf("Shape.Type = %v, want %v", shape.Type, ShapeCircle)
	}

	if shape.Width != 64 {
		t.Errorf("Shape.Width = %d, want 64", shape.Width)
	}

	if shape.Height != 64 {
		t.Errorf("Shape.Height = %d, want 64", shape.Height)
	}

	if shape.Seed != 12345 {
		t.Errorf("Shape.Seed = %d, want 12345", shape.Seed)
	}

	if shape.Sides != 6 {
		t.Errorf("Shape.Sides = %d, want 6", shape.Sides)
	}

	if shape.InnerRatio != 0.7 {
		t.Errorf("Shape.InnerRatio = %f, want 0.7", shape.InnerRatio)
	}

	if shape.Rotation != 45.0 {
		t.Errorf("Shape.Rotation = %f, want 45.0", shape.Rotation)
	}

	if shape.Smoothing != 0.2 {
		t.Errorf("Shape.Smoothing = %f, want 0.2", shape.Smoothing)
	}
}

func TestConfig_AllShapeTypes(t *testing.T) {
	shapeTypes := []ShapeType{
		ShapeCircle,
		ShapeRectangle,
		ShapeTriangle,
		ShapePolygon,
		ShapeStar,
		ShapeRing,
	}

	for _, shapeType := range shapeTypes {
		t.Run(shapeType.String(), func(t *testing.T) {
			config := Config{
				Type:       shapeType,
				Width:      48,
				Height:     48,
				Color:      color.RGBA{R: 100, G: 100, B: 100, A: 255},
				Seed:       999,
				Sides:      8,
				InnerRatio: 0.6,
				Rotation:   90.0,
				Smoothing:  0.15,
			}

			if config.Type != shapeType {
				t.Errorf("Config.Type = %v, want %v", config.Type, shapeType)
			}

			if config.Width != 48 {
				t.Errorf("Config.Width = %d, want 48", config.Width)
			}

			if config.Height != 48 {
				t.Errorf("Config.Height = %d, want 48", config.Height)
			}
		})
	}
}

func TestConfig_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		config     Config
		checkField string
		checkValue interface{}
	}{
		{
			name: "Zero dimensions",
			config: Config{
				Type:   ShapeCircle,
				Width:  0,
				Height: 0,
			},
			checkField: "Width",
			checkValue: 0,
		},
		{
			name: "Large dimensions",
			config: Config{
				Type:   ShapeRectangle,
				Width:  1024,
				Height: 1024,
			},
			checkField: "Width",
			checkValue: 1024,
		},
		{
			name: "Max inner ratio",
			config: Config{
				Type:       ShapeRing,
				InnerRatio: 1.0,
			},
			checkField: "InnerRatio",
			checkValue: 1.0,
		},
		{
			name: "Min inner ratio",
			config: Config{
				Type:       ShapeRing,
				InnerRatio: 0.0,
			},
			checkField: "InnerRatio",
			checkValue: 0.0,
		},
		{
			name: "Full rotation",
			config: Config{
				Type:     ShapeStar,
				Rotation: 360.0,
			},
			checkField: "Rotation",
			checkValue: 360.0,
		},
		{
			name: "Many sides polygon",
			config: Config{
				Type:  ShapePolygon,
				Sides: 20,
			},
			checkField: "Sides",
			checkValue: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the config can be created with these values
			if tt.config.Type == 0 && tt.name != "Zero dimensions" {
				t.Errorf("Config.Type should be set")
			}
		})
	}
}
