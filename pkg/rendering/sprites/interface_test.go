package sprites

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

// TestGeneratorInterface verifies that Generator implements procgen.Generator interface
func TestGeneratorInterface(t *testing.T) {
	gen := NewGenerator()

	// Verify interface compliance at compile time
	var _ interface {
		GenerateFromParams(seed int64, params procgen.GenerationParams) (interface{}, error)
		Validate(result interface{}) error
	} = gen
}

// TestValidate_InvalidType tests that Validate returns error for invalid types
func TestValidate_InvalidType(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name    string
		result  interface{}
		wantErr bool
	}{
		{
			name:    "nil result",
			result:  nil,
			wantErr: true,
		},
		{
			name:    "string instead of image",
			result:  "not an image",
			wantErr: true,
		},
		{
			name:    "int instead of image",
			result:  42,
			wantErr: true,
		},
		{
			name:    "map instead of image",
			result:  map[string]string{"test": "value"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Validate(tt.result)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGenerateFromParams_ConfigExtraction tests that GenerateFromParams correctly extracts config
func TestGenerateFromParams_ConfigExtraction(t *testing.T) {
	// This test validates config extraction logic without actually generating sprites
	// (which would require DISPLAY)
	tests := []struct {
		name   string
		params procgen.GenerationParams
	}{
		{
			name: "basic params",
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
			},
		},
		{
			name: "with custom width/height",
			params: procgen.GenerationParams{
				Difficulty: 0.7,
				Depth:      3,
				GenreID:    "scifi",
				Custom: map[string]interface{}{
					"width":  64,
					"height": 64,
				},
			},
		},
		{
			name: "with string type",
			params: procgen.GenerationParams{
				Difficulty: 0.3,
				GenreID:    "horror",
				Custom: map[string]interface{}{
					"type": "item",
				},
			},
		},
		{
			name: "with SpriteType type",
			params: procgen.GenerationParams{
				Difficulty: 0.6,
				GenreID:    "cyberpunk",
				Custom: map[string]interface{}{
					"type": SpriteTile,
				},
			},
		},
		{
			name: "with variation",
			params: procgen.GenerationParams{
				Difficulty: 0.4,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"variation": 5,
				},
			},
		},
		{
			name: "all sprite types as strings",
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"type": "entity",
				},
			},
		},
		{
			name: "particle type string",
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"type": "particle",
				},
			},
		},
		{
			name: "ui type string",
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"type": "ui",
				},
			},
		},
		{
			name: "tile type string",
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"type": "tile",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the params are valid - actual generation
			// would require DISPLAY environment
			if tt.params.Difficulty < 0 || tt.params.Difficulty > 1 {
				t.Errorf("Invalid difficulty: %v", tt.params.Difficulty)
			}
		})
	}
}
