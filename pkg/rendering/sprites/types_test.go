package sprites

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/palette"
)

func TestSpriteType_String(t *testing.T) {
	tests := []struct {
		name       string
		spriteType SpriteType
		expected   string
	}{
		{
			name:       "Entity",
			spriteType: SpriteEntity,
			expected:   "entity",
		},
		{
			name:       "Item",
			spriteType: SpriteItem,
			expected:   "item",
		},
		{
			name:       "Tile",
			spriteType: SpriteTile,
			expected:   "tile",
		},
		{
			name:       "Particle",
			spriteType: SpriteParticle,
			expected:   "particle",
		},
		{
			name:       "UI",
			spriteType: SpriteUI,
			expected:   "ui",
		},
		{
			name:       "Unknown",
			spriteType: SpriteType(99),
			expected:   "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.spriteType.String()
			if result != tt.expected {
				t.Errorf("SpriteType.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Type != SpriteEntity {
		t.Errorf("DefaultConfig().Type = %v, want %v", config.Type, SpriteEntity)
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

	if config.GenreID != "fantasy" {
		t.Errorf("DefaultConfig().GenreID = %s, want 'fantasy'", config.GenreID)
	}

	if config.Complexity != 0.5 {
		t.Errorf("DefaultConfig().Complexity = %f, want 0.5", config.Complexity)
	}

	if config.Variation != 0 {
		t.Errorf("DefaultConfig().Variation = %d, want 0", config.Variation)
	}

	if config.Custom == nil {
		t.Error("DefaultConfig().Custom should not be nil")
	}

	if len(config.Custom) != 0 {
		t.Errorf("DefaultConfig().Custom should be empty, got length %d", len(config.Custom))
	}
}

func TestConfig_Creation(t *testing.T) {
	testPalette := &palette.Palette{}
	
	config := Config{
		Type:       SpriteItem,
		Width:      64,
		Height:     64,
		Seed:       12345,
		Palette:    testPalette,
		GenreID:    "scifi",
		Complexity: 0.8,
		Variation:  5,
		Custom:     map[string]interface{}{"test": "value"},
	}

	if config.Type != SpriteItem {
		t.Errorf("Config.Type = %v, want %v", config.Type, SpriteItem)
	}

	if config.Width != 64 {
		t.Errorf("Config.Width = %d, want 64", config.Width)
	}

	if config.Height != 64 {
		t.Errorf("Config.Height = %d, want 64", config.Height)
	}

	if config.Seed != 12345 {
		t.Errorf("Config.Seed = %d, want 12345", config.Seed)
	}

	if config.GenreID != "scifi" {
		t.Errorf("Config.GenreID = %s, want 'scifi'", config.GenreID)
	}

	if config.Complexity != 0.8 {
		t.Errorf("Config.Complexity = %f, want 0.8", config.Complexity)
	}

	if config.Variation != 5 {
		t.Errorf("Config.Variation = %d, want 5", config.Variation)
	}

	if config.Palette == nil {
		t.Error("Config.Palette should not be nil")
	}

	if config.Custom == nil {
		t.Error("Config.Custom should not be nil")
	}

	if val, ok := config.Custom["test"]; !ok || val != "value" {
		t.Errorf("Config.Custom['test'] = %v, want 'value'", val)
	}
}

func TestConfig_AllSpriteTypes(t *testing.T) {
	spriteTypes := []SpriteType{
		SpriteEntity,
		SpriteItem,
		SpriteTile,
		SpriteParticle,
		SpriteUI,
	}

	for _, spriteType := range spriteTypes {
		t.Run(spriteType.String(), func(t *testing.T) {
			config := Config{
				Type:       spriteType,
				Width:      48,
				Height:     48,
				Seed:       999,
				GenreID:    "cyberpunk",
				Complexity: 0.7,
				Variation:  3,
			}

			if config.Type != spriteType {
				t.Errorf("Config.Type = %v, want %v", config.Type, spriteType)
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
		name   string
		config Config
		valid  bool
	}{
		{
			name: "Zero dimensions",
			config: Config{
				Type:   SpriteEntity,
				Width:  0,
				Height: 0,
			},
			valid: true, // Config doesn't validate, just stores values
		},
		{
			name: "Large dimensions",
			config: Config{
				Type:   SpriteEntity,
				Width:  1024,
				Height: 1024,
			},
			valid: true,
		},
		{
			name: "Negative seed",
			config: Config{
				Type: SpriteEntity,
				Seed: -12345,
			},
			valid: true,
		},
		{
			name: "Zero complexity",
			config: Config{
				Type:       SpriteEntity,
				Complexity: 0.0,
			},
			valid: true,
		},
		{
			name: "Max complexity",
			config: Config{
				Type:       SpriteEntity,
				Complexity: 1.0,
			},
			valid: true,
		},
		{
			name: "Negative variation",
			config: Config{
				Type:      SpriteEntity,
				Variation: -1,
			},
			valid: true,
		},
		{
			name: "Large variation",
			config: Config{
				Type:      SpriteEntity,
				Variation: 1000,
			},
			valid: true,
		},
		{
			name: "Empty genre",
			config: Config{
				Type:    SpriteEntity,
				GenreID: "",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the config can be created
			if !tt.valid {
				t.Errorf("Config should be valid but test expects invalid")
			}
		})
	}
}

func TestLayer_Creation(t *testing.T) {
	layer := Layer{
		OffsetX:   10,
		OffsetY:   20,
		ZIndex:    5,
		Opacity:   0.8,
		BlendMode: "normal",
	}

	if layer.OffsetX != 10 {
		t.Errorf("Layer.OffsetX = %d, want 10", layer.OffsetX)
	}

	if layer.OffsetY != 20 {
		t.Errorf("Layer.OffsetY = %d, want 20", layer.OffsetY)
	}

	if layer.ZIndex != 5 {
		t.Errorf("Layer.ZIndex = %d, want 5", layer.ZIndex)
	}

	if layer.Opacity != 0.8 {
		t.Errorf("Layer.Opacity = %f, want 0.8", layer.Opacity)
	}

	if layer.BlendMode != "normal" {
		t.Errorf("Layer.BlendMode = %s, want 'normal'", layer.BlendMode)
	}
}

func TestSprite_Creation(t *testing.T) {
	config := DefaultConfig()
	layers := []Layer{
		{ZIndex: 1, Opacity: 1.0},
		{ZIndex: 2, Opacity: 0.5},
	}

	sprite := Sprite{
		Config: config,
		Layers: layers,
		Width:  64,
		Height: 64,
	}

	if sprite.Width != 64 {
		t.Errorf("Sprite.Width = %d, want 64", sprite.Width)
	}

	if sprite.Height != 64 {
		t.Errorf("Sprite.Height = %d, want 64", sprite.Height)
	}

	if len(sprite.Layers) != 2 {
		t.Errorf("Sprite.Layers length = %d, want 2", len(sprite.Layers))
	}

	if sprite.Config.Type != SpriteEntity {
		t.Errorf("Sprite.Config.Type = %v, want %v", sprite.Config.Type, SpriteEntity)
	}
}

func TestSprite_EmptyLayers(t *testing.T) {
	sprite := Sprite{
		Config: DefaultConfig(),
		Layers: []Layer{},
		Width:  32,
		Height: 32,
	}

	if len(sprite.Layers) != 0 {
		t.Errorf("Sprite.Layers should be empty, got length %d", len(sprite.Layers))
	}
}

func TestSprite_MultipleLayers(t *testing.T) {
	layers := []Layer{
		{ZIndex: 0, Opacity: 1.0, BlendMode: "normal"},
		{ZIndex: 1, Opacity: 0.8, BlendMode: "multiply"},
		{ZIndex: 2, Opacity: 0.6, BlendMode: "add"},
		{ZIndex: 3, Opacity: 0.4, BlendMode: "overlay"},
	}

	sprite := Sprite{
		Config: DefaultConfig(),
		Layers: layers,
		Width:  128,
		Height: 128,
	}

	if len(sprite.Layers) != 4 {
		t.Errorf("Sprite.Layers length = %d, want 4", len(sprite.Layers))
	}

	// Verify layers are stored correctly
	for i, layer := range sprite.Layers {
		if layer.ZIndex != i {
			t.Errorf("Layer %d ZIndex = %d, want %d", i, layer.ZIndex, i)
		}
	}
}

func TestConfig_CustomParameters(t *testing.T) {
	config := DefaultConfig()
	
	// Add custom parameters
	config.Custom["animation"] = true
	config.Custom["frameCount"] = 4
	config.Custom["duration"] = 1.5

	if val, ok := config.Custom["animation"]; !ok || val != true {
		t.Errorf("Custom['animation'] = %v, want true", val)
	}

	if val, ok := config.Custom["frameCount"]; !ok || val != 4 {
		t.Errorf("Custom['frameCount'] = %v, want 4", val)
	}

	if val, ok := config.Custom["duration"]; !ok || val != 1.5 {
		t.Errorf("Custom['duration'] = %v, want 1.5", val)
	}
}

func TestConfig_DifferentGenres(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			config := Config{
				Type:    SpriteEntity,
				GenreID: genre,
			}

			if config.GenreID != genre {
				t.Errorf("Config.GenreID = %s, want %s", config.GenreID, genre)
			}
		})
	}
}

func TestLayer_ZIndexOrdering(t *testing.T) {
	layers := []Layer{
		{ZIndex: 10},
		{ZIndex: 5},
		{ZIndex: 1},
		{ZIndex: 20},
	}

	// Verify layers can be created with any Z-index
	for i, layer := range layers {
		expectedZIndex := []int{10, 5, 1, 20}[i]
		if layer.ZIndex != expectedZIndex {
			t.Errorf("Layer %d ZIndex = %d, want %d", i, layer.ZIndex, expectedZIndex)
		}
	}
}

func TestLayer_OpacityRange(t *testing.T) {
	tests := []struct {
		name    string
		opacity float64
	}{
		{"Transparent", 0.0},
		{"Quarter", 0.25},
		{"Half", 0.5},
		{"ThreeQuarters", 0.75},
		{"Opaque", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layer := Layer{
				Opacity: tt.opacity,
			}

			if layer.Opacity != tt.opacity {
				t.Errorf("Layer.Opacity = %f, want %f", layer.Opacity, tt.opacity)
			}
		})
	}
}

func TestLayer_BlendModes(t *testing.T) {
	blendModes := []string{"normal", "multiply", "add", "overlay", "screen", "darken", "lighten"}

	for _, mode := range blendModes {
		t.Run(mode, func(t *testing.T) {
			layer := Layer{
				BlendMode: mode,
			}

			if layer.BlendMode != mode {
				t.Errorf("Layer.BlendMode = %s, want %s", layer.BlendMode, mode)
			}
		})
	}
}
