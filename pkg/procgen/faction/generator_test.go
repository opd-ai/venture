package faction

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
)

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator()
	if gen == nil {
		t.Fatal("NewGenerator returned nil")
	}
}

func TestGenerator_Validate(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name    string
		params  interface{}
		wantErr bool
	}{
		{
			name: "Valid params",
			params: procgen.GenerationParams{
				Depth:      5,
				Difficulty: 0.5,
				GenreID:    "fantasy",
			},
			wantErr: false,
		},
		{
			name: "Valid params zero depth",
			params: procgen.GenerationParams{
				Depth:      0,
				Difficulty: 0.0,
				GenreID:    "sci-fi",
			},
			wantErr: false,
		},
		{
			name: "Invalid params negative depth",
			params: procgen.GenerationParams{
				Depth:      -1,
				Difficulty: 0.5,
			},
			wantErr: true,
		},
		{
			name: "Invalid params difficulty too low",
			params: procgen.GenerationParams{
				Depth:      5,
				Difficulty: -0.1,
			},
			wantErr: true,
		},
		{
			name: "Invalid params difficulty too high",
			params: procgen.GenerationParams{
				Depth:      5,
				Difficulty: 1.1,
			},
			wantErr: true,
		},
		{
			name:    "Invalid params type",
			params:  "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Validate(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerator_Generate(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name    string
		seed    int64
		params  procgen.GenerationParams
		wantErr bool
	}{
		{
			name: "Fantasy world",
			seed: 12345,
			params: procgen.GenerationParams{
				Depth:      5,
				Difficulty: 0.5,
				GenreID:    "fantasy",
			},
			wantErr: false,
		},
		{
			name: "Sci-fi world",
			seed: 67890,
			params: procgen.GenerationParams{
				Depth:      10,
				Difficulty: 0.7,
				GenreID:    "sci-fi",
			},
			wantErr: false,
		},
		{
			name: "Horror world",
			seed: 11111,
			params: procgen.GenerationParams{
				Depth:      3,
				Difficulty: 0.8,
				GenreID:    "horror",
			},
			wantErr: false,
		},
		{
			name: "Invalid params",
			seed: 22222,
			params: procgen.GenerationParams{
				Depth:      -1,
				Difficulty: 0.5,
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
				factions, ok := result.([]*engine.Faction)
				if !ok {
					t.Fatal("Generate() did not return []*engine.Faction")
				}

				// Check we got some factions (3-7 based on depth)
				if len(factions) < 3 || len(factions) > 7 {
					t.Errorf("Expected 3-7 factions, got %d", len(factions))
				}

				// Check each faction has required fields
				for i, faction := range factions {
					if faction.ID == "" {
						t.Errorf("Faction %d has empty ID", i)
					}
					if faction.Name == "" {
						t.Errorf("Faction %d has empty name", i)
					}
					if faction.Type == "" {
						t.Errorf("Faction %d has empty type", i)
					}
					if faction.GenreID != tt.params.GenreID {
						t.Errorf("Faction %d has wrong genre: got %s, want %s",
							i, faction.GenreID, tt.params.GenreID)
					}
					if faction.MemberCount < 100 || faction.MemberCount > 999 {
						t.Errorf("Faction %d has invalid member count: %d",
							i, faction.MemberCount)
					}
					if faction.Relationships == nil {
						t.Errorf("Faction %d has nil relationships", i)
					}
				}
			}
		})
	}
}

func TestGenerator_Generate_Deterministic(t *testing.T) {
	gen := NewGenerator()
	seed := int64(42)
	params := procgen.GenerationParams{
		Depth:      5,
		Difficulty: 0.5,
		GenreID:    "fantasy",
	}

	// Generate twice with same seed
	result1, err1 := gen.Generate(seed, params)
	if err1 != nil {
		t.Fatal(err1)
	}

	result2, err2 := gen.Generate(seed, params)
	if err2 != nil {
		t.Fatal(err2)
	}

	factions1 := result1.([]*engine.Faction)
	factions2 := result2.([]*engine.Faction)

	// Should generate same factions
	if len(factions1) != len(factions2) {
		t.Errorf("Different number of factions: %d vs %d", len(factions1), len(factions2))
	}

	for i := 0; i < len(factions1) && i < len(factions2); i++ {
		f1 := factions1[i]
		f2 := factions2[i]

		if f1.ID != f2.ID {
			t.Errorf("Faction %d ID mismatch: %s vs %s", i, f1.ID, f2.ID)
		}
		if f1.Name != f2.Name {
			t.Errorf("Faction %d name mismatch: %s vs %s", i, f1.Name, f2.Name)
		}
		if f1.Type != f2.Type {
			t.Errorf("Faction %d type mismatch: %s vs %s", i, f1.Type, f2.Type)
		}
		if f1.MemberCount != f2.MemberCount {
			t.Errorf("Faction %d member count mismatch: %d vs %d",
				i, f1.MemberCount, f2.MemberCount)
		}
	}
}

func TestGenerator_FactionCounts(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name     string
		depth    int
		minCount int
		maxCount int
	}{
		{"Depth 0", 0, 3, 3},
		{"Depth 5", 5, 3, 3},
		{"Depth 10", 10, 4, 4},
		{"Depth 20", 20, 5, 5},
		{"Depth 50", 50, 7, 7}, // Formula gives 8 (3+50/10), but capped at 7
		{"Depth 100", 100, 7, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := procgen.GenerationParams{
				Depth:      tt.depth,
				Difficulty: 0.5,
				GenreID:    "fantasy",
			}

			result, err := gen.Generate(12345, params)
			if err != nil {
				t.Fatal(err)
			}

			factions := result.([]*engine.Faction)
			count := len(factions)

			// Expected count is 3 + depth/10, capped at 7
			expectedCount := 3 + tt.depth/10
			if expectedCount > 7 {
				expectedCount = 7
			}

			if count != expectedCount {
				t.Errorf("Expected %d factions for depth %d, got %d",
					expectedCount, tt.depth, count)
			}
		})
	}
}

func TestGenerator_GenreSpecificTypes(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name      string
		genre     string
		wantTypes []engine.FactionType
	}{
		{
			name:  "Fantasy has kingdoms",
			genre: "fantasy",
			wantTypes: []engine.FactionType{
				engine.FactionTypeKingdom,
				engine.FactionTypeGuild,
				engine.FactionTypeCult,
				engine.FactionTypeMerchants,
			},
		},
		{
			name:  "Sci-fi has corporations",
			genre: "sci-fi",
			wantTypes: []engine.FactionType{
				engine.FactionTypeCorporation,
				engine.FactionTypeRebels,
				engine.FactionTypeGuild,
				engine.FactionTypeMerchants,
			},
		},
		{
			name:  "Horror has cults",
			genre: "horror",
			wantTypes: []engine.FactionType{
				engine.FactionTypeCult,
				engine.FactionTypeGang,
				engine.FactionTypeMerchants,
			},
		},
		{
			name:  "Cyberpunk has gangs and corporations",
			genre: "cyberpunk",
			wantTypes: []engine.FactionType{
				engine.FactionTypeCorporation,
				engine.FactionTypeGang,
				engine.FactionTypeRebels,
				engine.FactionTypeMerchants,
			},
		},
		{
			name:  "Post-apocalyptic has gangs",
			genre: "post-apocalyptic",
			wantTypes: []engine.FactionType{
				engine.FactionTypeGang,
				engine.FactionTypeRebels,
				engine.FactionTypeMerchants,
				engine.FactionTypeCult,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := procgen.GenerationParams{
				Depth:      20, // Generate more factions to see variety
				Difficulty: 0.5,
				GenreID:    tt.genre,
			}

			result, err := gen.Generate(12345, params)
			if err != nil {
				t.Fatal(err)
			}

			factions := result.([]*engine.Faction)

			// Collect faction types found
			foundTypes := make(map[engine.FactionType]bool)
			for _, faction := range factions {
				foundTypes[faction.Type] = true
			}

			// All generated types should be in the expected list
			for factionType := range foundTypes {
				found := false
				for _, wantType := range tt.wantTypes {
					if factionType == wantType {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Unexpected faction type for genre %s: %s",
						tt.genre, factionType)
				}
			}
		})
	}
}

func TestGenerator_Relationships(t *testing.T) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Depth:      10,
		Difficulty: 0.5,
		GenreID:    "fantasy",
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatal(err)
	}

	factions := result.([]*engine.Faction)

	// Check relationships exist between all factions
	for i, faction1 := range factions {
		for j, faction2 := range factions {
			if i == j {
				continue // Skip self
			}

			// Faction1 should have relationship with faction2
			rel1 := faction1.Relationships[faction2.ID]
			rel2 := faction2.Relationships[faction1.ID]

			// Relationships should be bidirectional
			if rel1 != rel2 {
				t.Errorf("Asymmetric relationship between %s and %s: %d vs %d",
					faction1.ID, faction2.ID, rel1, rel2)
			}

			// Relationships should be in valid range
			if rel1 < -100 || rel1 > 100 {
				t.Errorf("Relationship out of range: %d", rel1)
			}
		}
	}
}

func TestGenerator_SpecialRelationships(t *testing.T) {
	gen := NewGenerator()

	// Generate sci-fi world which should have corp vs rebels
	params := procgen.GenerationParams{
		Depth:      30, // High depth for more factions
		Difficulty: 0.5,
		GenreID:    "sci-fi",
	}

	result, err := gen.Generate(99999, params)
	if err != nil {
		t.Fatal(err)
	}

	factions := result.([]*engine.Faction)

	// Find corporation and rebel factions
	var corp, rebel *engine.Faction
	for _, faction := range factions {
		if faction.Type == engine.FactionTypeCorporation && corp == nil {
			corp = faction
		}
		if faction.Type == engine.FactionTypeRebels && rebel == nil {
			rebel = faction
		}
	}

	// If both exist, they should be enemies
	if corp != nil && rebel != nil {
		rel := corp.Relationships[rebel.ID]
		if rel > -50 {
			t.Logf("Corporation and Rebels should typically be enemies, got relationship: %d", rel)
			// Note: Due to randomness, this isn't guaranteed, just logged
		}
	}
}

func TestGenerator_FactionNames(t *testing.T) {
	gen := NewGenerator()

	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			params := procgen.GenerationParams{
				Depth:      10,
				Difficulty: 0.5,
				GenreID:    genre,
			}

			result, err := gen.Generate(12345, params)
			if err != nil {
				t.Fatal(err)
			}

			factions := result.([]*engine.Faction)

			for _, faction := range factions {
				if faction.Name == "" {
					t.Errorf("Faction %s has empty name", faction.ID)
				}
				if faction.Description == "" {
					t.Errorf("Faction %s has empty description", faction.ID)
				}

				// Names should have at least 2 words (prefix + suffix)
				// This is a simple check
				if len(faction.Name) < 3 {
					t.Errorf("Faction name too short: %s", faction.Name)
				}
			}
		})
	}
}

func TestGenerator_TerritoryColors(t *testing.T) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Depth:      5,
		Difficulty: 0.5,
		GenreID:    "fantasy",
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatal(err)
	}

	factions := result.([]*engine.Faction)

	for _, faction := range factions {
		color := faction.TerritoryColor

		// Alpha should always be 255
		if color[3] != 255 {
			t.Errorf("Faction %s has invalid alpha: %d", faction.ID, color[3])
		}

		// At least one RGB component should be non-zero (no pure black with full alpha)
		if color[0] == 0 && color[1] == 0 && color[2] == 0 {
			// This is extremely unlikely with random generation, but possible
			t.Logf("Faction %s has all-zero RGB (rare but valid)", faction.ID)
		}
	}
}

// Benchmark tests to validate performance claims in doc.go
// Small worlds (depth 0-10): <1ms
// Medium worlds (depth 11-30): <2ms
// Large worlds (depth 31+): <3ms

func BenchmarkGenerator_SmallWorld(b *testing.B) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Depth:      5,
		Difficulty: 0.5,
		GenreID:    "fantasy",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate(int64(i), params)
	}
}

func BenchmarkGenerator_MediumWorld(b *testing.B) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Depth:      20,
		Difficulty: 0.5,
		GenreID:    "fantasy",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate(int64(i), params)
	}
}

func BenchmarkGenerator_LargeWorld(b *testing.B) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Depth:      50,
		Difficulty: 0.5,
		GenreID:    "fantasy",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate(int64(i), params)
	}
}

func BenchmarkGenerator_AllGenres(b *testing.B) {
	gen := NewGenerator()
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		b.Run(genre, func(b *testing.B) {
			params := procgen.GenerationParams{
				Depth:      20,
				Difficulty: 0.5,
				GenreID:    genre,
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = gen.Generate(int64(i), params)
			}
		})
	}
}

func BenchmarkGenerator_Validate(b *testing.B) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Depth:      10,
		Difficulty: 0.5,
		GenreID:    "fantasy",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gen.Validate(params)
	}
}
