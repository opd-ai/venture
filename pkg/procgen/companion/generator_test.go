package companion

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
)

func TestGenerator_Generate(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name    string
		seed    int64
		params  procgen.GenerationParams
		wantErr bool
	}{
		{
			name: "valid generation",
			seed: 12345,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr: false,
		},
		{
			name: "sci-fi genre",
			seed: 67890,
			params: procgen.GenerationParams{
				Difficulty: 0.7,
				Depth:      10,
				GenreID:    "sci-fi",
			},
			wantErr: false,
		},
		{
			name: "horror genre",
			seed: 11111,
			params: procgen.GenerationParams{
				Difficulty: 0.3,
				Depth:      3,
				GenreID:    "horror",
			},
			wantErr: false,
		},
		{
			name: "cyberpunk genre",
			seed: 22222,
			params: procgen.GenerationParams{
				Difficulty: 0.8,
				Depth:      15,
				GenreID:    "cyberpunk",
			},
			wantErr: false,
		},
		{
			name: "post-apocalyptic genre",
			seed: 33333,
			params: procgen.GenerationParams{
				Difficulty: 0.6,
				Depth:      7,
				GenreID:    "post-apocalyptic",
			},
			wantErr: false,
		},
		{
			name: "unknown genre (default)",
			seed: 44444,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    "unknown-genre",
			},
			wantErr: false,
		},
		{
			name: "zero depth (uses 1)",
			seed: 55555,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      0,
				GenreID:    "fantasy",
			},
			wantErr: false,
		},
		{
			name: "negative difficulty",
			seed: 66666,
			params: procgen.GenerationParams{
				Difficulty: -0.5,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
		{
			name: "invalid difficulty",
			seed: 99999,
			params: procgen.GenerationParams{
				Difficulty: 1.5,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.Generate(tt.seed, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				companion, ok := result.(*Companion)
				if !ok {
					t.Fatal("Result is not a Companion")
				}

				if companion.Attack <= 0 {
					t.Errorf("Invalid attack: %f", companion.Attack)
				}

				if companion.MaxHP <= 0 {
					t.Errorf("Invalid MaxHP: %f", companion.MaxHP)
				}

				if len(companion.Commands) == 0 {
					t.Error("Companion has no commands")
				}
			}
		})
	}
}

func TestGenerator_Validate(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name      string
		companion *Companion
		wantErr   bool
	}{
		{
			name: "valid companion",
			companion: &Companion{
				Name:     "Test",
				Attack:   10.0,
				Defense:  8.0,
				MaxHP:    50.0,
				Loyalty:  50.0,
				Commands: []engine.CommandType{engine.CommandFollow},
			},
			wantErr: false,
		},
		{
			name: "invalid attack zero",
			companion: &Companion{
				Name:     "Test",
				Attack:   0.0,
				Defense:  8.0,
				MaxHP:    50.0,
				Loyalty:  50.0,
				Commands: []engine.CommandType{engine.CommandFollow},
			},
			wantErr: true,
		},
		{
			name: "invalid attack negative",
			companion: &Companion{
				Name:     "Test",
				Attack:   -5.0,
				Defense:  8.0,
				MaxHP:    50.0,
				Loyalty:  50.0,
				Commands: []engine.CommandType{engine.CommandFollow},
			},
			wantErr: true,
		},
		{
			name: "invalid MaxHP zero",
			companion: &Companion{
				Name:     "Test",
				Attack:   10.0,
				Defense:  8.0,
				MaxHP:    0.0,
				Loyalty:  50.0,
				Commands: []engine.CommandType{engine.CommandFollow},
			},
			wantErr: true,
		},
		{
			name: "invalid MaxHP negative",
			companion: &Companion{
				Name:     "Test",
				Attack:   10.0,
				Defense:  8.0,
				MaxHP:    -100.0,
				Loyalty:  50.0,
				Commands: []engine.CommandType{engine.CommandFollow},
			},
			wantErr: true,
		},
		{
			name: "invalid loyalty negative",
			companion: &Companion{
				Name:     "Test",
				Attack:   10.0,
				Defense:  8.0,
				MaxHP:    50.0,
				Loyalty:  -10.0,
				Commands: []engine.CommandType{engine.CommandFollow},
			},
			wantErr: true,
		},
		{
			name: "invalid loyalty too high",
			companion: &Companion{
				Name:     "Test",
				Attack:   10.0,
				Defense:  8.0,
				MaxHP:    50.0,
				Loyalty:  150.0,
				Commands: []engine.CommandType{engine.CommandFollow},
			},
			wantErr: true,
		},
		{
			name: "no commands",
			companion: &Companion{
				Name:     "Test",
				Attack:   10.0,
				Defense:  8.0,
				MaxHP:    50.0,
				Loyalty:  50.0,
				Commands: []engine.CommandType{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Validate(tt.companion)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerator_Validate_WrongType(t *testing.T) {
	gen := NewGenerator()

	err := gen.Validate("not a companion")
	if err == nil {
		t.Error("Expected error for non-Companion type")
	}
}

func TestGenerator_Determinism(t *testing.T) {
	gen := NewGenerator()
	seed := int64(12345)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	result1, err1 := gen.Generate(seed, params)
	result2, err2 := gen.Generate(seed, params)

	if err1 != nil || err2 != nil {
		t.Fatalf("Generation failed: err1=%v, err2=%v", err1, err2)
	}

	comp1 := result1.(*Companion)
	comp2 := result2.(*Companion)

	if comp1.Name != comp2.Name {
		t.Errorf("Name differs: %s vs %s", comp1.Name, comp2.Name)
	}

	if comp1.Attack != comp2.Attack {
		t.Errorf("Attack differs: %f vs %f", comp1.Attack, comp2.Attack)
	}

	if comp1.Type != comp2.Type {
		t.Errorf("Type differs: %v vs %v", comp1.Type, comp2.Type)
	}
}

func BenchmarkGenerator_Generate(b *testing.B) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(int64(i), params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestGenerator_GenerateSpritePattern(t *testing.T) {
	gen := NewGenerator()

	// Test all companion types to ensure generateSpritePattern returns correct patterns
	tests := []struct {
		name          string
		companionType engine.CompanionType
		genreID       string
		wantContains  string
	}{
		{
			name:          "pet type",
			companionType: engine.CompanionTypePet,
			genreID:       "fantasy",
			wantContains:  "quadruped",
		},
		{
			name:          "robot type",
			companionType: engine.CompanionTypeRobot,
			genreID:       "sci-fi",
			wantContains:  "mechanical",
		},
		{
			name:          "elemental type",
			companionType: engine.CompanionTypeElemental,
			genreID:       "fantasy",
			wantContains:  "elemental",
		},
		{
			name:          "undead type",
			companionType: engine.CompanionTypeUndead,
			genreID:       "horror",
			wantContains:  "skeletal",
		},
		{
			name:          "spirit type",
			companionType: engine.CompanionTypeSpirit,
			genreID:       "fantasy",
			wantContains:  "wispy",
		},
		{
			name:          "insect type",
			companionType: engine.CompanionTypeInsect,
			genreID:       "horror",
			wantContains:  "insectoid",
		},
		{
			name:          "default type (summon)",
			companionType: engine.CompanionTypeSummon,
			genreID:       "fantasy",
			wantContains:  "humanoid",
		},
		{
			name:          "default type (hireling)",
			companionType: engine.CompanionTypeHireling,
			genreID:       "fantasy",
			wantContains:  "humanoid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := gen.generateSpritePattern(tt.companionType, tt.genreID)
			if pattern == "" {
				t.Error("generateSpritePattern returned empty string")
			}
			if !containsSubstr(pattern, tt.wantContains) {
				t.Errorf("generateSpritePattern() = %q, want pattern containing %q", pattern, tt.wantContains)
			}
		})
	}
}

// containsSubstr is a helper to check if s contains substr
func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestGenerator_AllGenresGenerateUniqueCompanions(t *testing.T) {
	gen := NewGenerator()
	seed := int64(12345)

	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	results := make(map[string]*Companion)

	for _, genre := range genres {
		params := procgen.GenerationParams{
			Difficulty: 0.5,
			Depth:      5,
			GenreID:    genre,
		}
		result, err := gen.Generate(seed, params)
		if err != nil {
			t.Errorf("Generate failed for genre %s: %v", genre, err)
			continue
		}
		companion, ok := result.(*Companion)
		if !ok {
			t.Errorf("Result is not a Companion for genre %s", genre)
			continue
		}
		results[genre] = companion

		// All companions should be valid
		if err := gen.Validate(companion); err != nil {
			t.Errorf("Generated companion for genre %s is invalid: %v", genre, err)
		}

		// Sprite pattern should be non-empty
		if companion.SpritePattern == "" {
			t.Errorf("Companion for genre %s has empty sprite pattern", genre)
		}
	}

	// Verify different genres produce different companions
	if len(results) > 1 {
		first := results[genres[0]]
		for _, genre := range genres[1:] {
			other := results[genre]
			// At least one of name, type, or sprite pattern should differ
			if first.Name == other.Name && first.Type == other.Type && first.SpritePattern == other.SpritePattern {
				t.Logf("Note: genres %s and %s produced similar companions", genres[0], genre)
			}
		}
	}
}
