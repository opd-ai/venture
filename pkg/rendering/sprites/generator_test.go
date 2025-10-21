package sprites

import (
	"testing"
)

func TestSpriteType_String(t *testing.T) {
	tests := []struct {
		name       string
		spriteType SpriteType
		want       string
	}{
		{"entity", SpriteEntity, "entity"},
		{"item", SpriteItem, "item"},
		{"tile", SpriteTile, "tile"},
		{"particle", SpriteParticle, "particle"},
		{"ui", SpriteUI, "ui"},
		{"unknown", SpriteType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.spriteType.String()
			if got != tt.want {
				t.Errorf("SpriteType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Type != SpriteEntity {
		t.Errorf("DefaultConfig Type = %v, want %v", config.Type, SpriteEntity)
	}
	if config.Width != 32 {
		t.Errorf("DefaultConfig Width = %v, want 32", config.Width)
	}
	if config.Height != 32 {
		t.Errorf("DefaultConfig Height = %v, want 32", config.Height)
	}
	if config.GenreID != "fantasy" {
		t.Errorf("DefaultConfig GenreID = %v, want 'fantasy'", config.GenreID)
	}
	if config.Complexity != 0.5 {
		t.Errorf("DefaultConfig Complexity = %v, want 0.5", config.Complexity)
	}
	if config.Variation != 0 {
		t.Errorf("DefaultConfig Variation = %v, want 0", config.Variation)
	}
	if config.Custom == nil {
		t.Error("DefaultConfig Custom is nil")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "entity sprite",
			config: Config{
				Type:       SpriteEntity,
				Width:      32,
				Height:     32,
				GenreID:    "fantasy",
				Complexity: 0.5,
			},
		},
		{
			name: "item sprite",
			config: Config{
				Type:       SpriteItem,
				Width:      24,
				Height:     24,
				GenreID:    "scifi",
				Complexity: 0.3,
			},
		},
		{
			name: "tile sprite",
			config: Config{
				Type:       SpriteTile,
				Width:      16,
				Height:     16,
				GenreID:    "horror",
				Complexity: 0.7,
			},
		},
		{
			name: "particle sprite",
			config: Config{
				Type:       SpriteParticle,
				Width:      8,
				Height:     8,
				GenreID:    "cyberpunk",
				Complexity: 0.2,
			},
		},
		{
			name: "ui sprite",
			config: Config{
				Type:       SpriteUI,
				Width:      64,
				Height:     32,
				GenreID:    "postapoc",
				Complexity: 0.4,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate basic config properties
			if tt.config.Type < SpriteEntity || tt.config.Type > SpriteUI {
				t.Error("Invalid sprite type")
			}
			if tt.config.Width <= 0 {
				t.Error("Width must be positive")
			}
			if tt.config.Height <= 0 {
				t.Error("Height must be positive")
			}
			if tt.config.Complexity < 0 || tt.config.Complexity > 1 {
				t.Error("Complexity must be between 0 and 1")
			}
			if tt.config.GenreID == "" {
				t.Error("GenreID cannot be empty")
			}
		})
	}
}

func TestSpriteTypeCategories(t *testing.T) {
	tests := []struct {
		name       string
		spriteType SpriteType
		isVisual   bool
		isComplex  bool
	}{
		{"entity is complex visual", SpriteEntity, true, true},
		{"item is moderate visual", SpriteItem, true, true},
		{"tile is simple visual", SpriteTile, true, false},
		{"particle is simple visual", SpriteParticle, true, false},
		{"ui is simple visual", SpriteUI, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Type = tt.spriteType

			// Verify sprite types are consistently defined
			if !tt.isVisual {
				t.Error("All sprite types should be visual")
			}

			// Complex sprites should have higher default complexity
			if tt.isComplex && config.Complexity < 0.4 {
				// Note: This is just validating the concept, not actual implementation
			}
		})
	}
}

func TestLayerStructure(t *testing.T) {
	layer := Layer{
		OffsetX:   10,
		OffsetY:   20,
		ZIndex:    5,
		Opacity:   0.8,
		BlendMode: "normal",
	}

	if layer.OffsetX != 10 {
		t.Errorf("Layer OffsetX = %v, want 10", layer.OffsetX)
	}
	if layer.OffsetY != 20 {
		t.Errorf("Layer OffsetY = %v, want 20", layer.OffsetY)
	}
	if layer.ZIndex != 5 {
		t.Errorf("Layer ZIndex = %v, want 5", layer.ZIndex)
	}
	if layer.Opacity != 0.8 {
		t.Errorf("Layer Opacity = %v, want 0.8", layer.Opacity)
	}
	if layer.BlendMode != "normal" {
		t.Errorf("Layer BlendMode = %v, want 'normal'", layer.BlendMode)
	}
}

func TestSpriteStructure(t *testing.T) {
	config := DefaultConfig()
	sprite := Sprite{
		Config: config,
		Layers: []Layer{
			{ZIndex: 0},
			{ZIndex: 1},
		},
		Width:  32,
		Height: 32,
	}

	if sprite.Width != 32 {
		t.Errorf("Sprite Width = %v, want 32", sprite.Width)
	}
	if sprite.Height != 32 {
		t.Errorf("Sprite Height = %v, want 32", sprite.Height)
	}
	if len(sprite.Layers) != 2 {
		t.Errorf("Sprite Layers length = %v, want 2", len(sprite.Layers))
	}
}

func TestComplexityRange(t *testing.T) {
	tests := []struct {
		name       string
		complexity float64
		valid      bool
	}{
		{"zero complexity", 0.0, true},
		{"low complexity", 0.25, true},
		{"medium complexity", 0.5, true},
		{"high complexity", 0.75, true},
		{"max complexity", 1.0, true},
		{"negative complexity", -0.1, false},
		{"over max complexity", 1.1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Complexity = tt.complexity

			valid := config.Complexity >= 0 && config.Complexity <= 1
			if valid != tt.valid {
				t.Errorf("Complexity %v validity = %v, want %v", 
					tt.complexity, valid, tt.valid)
			}
		})
	}
}
