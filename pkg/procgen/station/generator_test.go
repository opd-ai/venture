package station

import (
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

// TestNewStationGenerator tests station generator creation.
func TestNewStationGenerator(t *testing.T) {
	gen := NewStationGenerator()
	if gen == nil {
		t.Fatal("NewStationGenerator returned nil")
	}

	if gen.nameTemplates == nil {
		t.Error("nameTemplates not initialized")
	}

	// Check that all genres are registered
	expectedGenres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc", ""}
	for _, genre := range expectedGenres {
		if _, exists := gen.nameTemplates[genre]; !exists {
			t.Errorf("genre %q not registered", genre)
		}
	}
}

// TestGenerate tests basic station generation.
func TestGenerate(t *testing.T) {
	gen := NewStationGenerator()

	tests := []struct {
		name    string
		seed    int64
		genreID string
	}{
		{"fantasy genre", 12345, "fantasy"},
		{"scifi genre", 54321, "scifi"},
		{"horror genre", 11111, "horror"},
		{"cyberpunk genre", 22222, "cyberpunk"},
		{"postapoc genre", 33333, "postapoc"},
		{"empty genre defaults to fantasy", 99999, ""},
		{"unknown genre defaults to fantasy", 88888, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    tt.genreID,
			}

			result, err := gen.Generate(tt.seed, params)
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			stations, ok := result.([]*StationData)
			if !ok {
				t.Fatal("result is not []*StationData")
			}

			if len(stations) != 3 {
				t.Errorf("expected 3 stations, got %d", len(stations))
			}

			// Check station properties
			for i, station := range stations {
				if station == nil {
					t.Errorf("station %d is nil", i)
					continue
				}

				if station.Name == "" {
					t.Errorf("station %d has empty name", i)
				}

				if station.GenreID == "" {
					expectedGenre := tt.genreID
					if expectedGenre == "" || expectedGenre == "unknown" {
						expectedGenre = "fantasy"
					}
					if station.GenreID != expectedGenre {
						t.Errorf("station %d has wrong genre: got %q, want %q", i, station.GenreID, expectedGenre)
					}
				}

				if station.Seed == 0 {
					t.Errorf("station %d has zero seed", i)
				}
			}

			// Check that we have one of each type
			typeCount := make(map[StationType]int)
			for _, station := range stations {
				typeCount[station.StationType]++
			}

			if typeCount[StationAlchemyTable] != 1 {
				t.Errorf("expected 1 alchemy table, got %d", typeCount[StationAlchemyTable])
			}
			if typeCount[StationForge] != 1 {
				t.Errorf("expected 1 forge, got %d", typeCount[StationForge])
			}
			if typeCount[StationWorkbench] != 1 {
				t.Errorf("expected 1 workbench, got %d", typeCount[StationWorkbench])
			}
		})
	}
}

// TestGenerateDeterminism tests that generation is deterministic.
func TestGenerateDeterminism(t *testing.T) {
	gen := NewStationGenerator()
	seed := int64(42)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
	}

	// Generate twice with same seed
	result1, err1 := gen.Generate(seed, params)
	result2, err2 := gen.Generate(seed, params)

	if err1 != nil || err2 != nil {
		t.Fatalf("Generate failed: %v, %v", err1, err2)
	}

	stations1 := result1.([]*StationData)
	stations2 := result2.([]*StationData)

	if len(stations1) != len(stations2) {
		t.Fatalf("station counts differ: %d vs %d", len(stations1), len(stations2))
	}

	// Check that all properties match
	for i := range stations1 {
		s1 := stations1[i]
		s2 := stations2[i]

		if s1.StationType != s2.StationType {
			t.Errorf("station %d type mismatch: %v vs %v", i, s1.StationType, s2.StationType)
		}

		if s1.Name != s2.Name {
			t.Errorf("station %d name mismatch: %q vs %q", i, s1.Name, s2.Name)
		}

		if s1.GenreID != s2.GenreID {
			t.Errorf("station %d genre mismatch: %q vs %q", i, s1.GenreID, s2.GenreID)
		}

		if s1.Seed != s2.Seed {
			t.Errorf("station %d seed mismatch: %d vs %d", i, s1.Seed, s2.Seed)
		}
	}
}

// TestGenerateDifferentSeeds tests that different seeds produce different names.
func TestGenerateDifferentSeeds(t *testing.T) {
	gen := NewStationGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
	}

	result1, _ := gen.Generate(100, params)
	result2, _ := gen.Generate(200, params)

	stations1 := result1.([]*StationData)
	stations2 := result2.([]*StationData)

	// At least one station should have a different name
	allSame := true
	for i := range stations1 {
		if stations1[i].Name != stations2[i].Name {
			allSame = false
			break
		}
	}

	if allSame {
		t.Error("different seeds produced identical station names")
	}
}

// TestValidate tests the validation function.
func TestValidate(t *testing.T) {
	gen := NewStationGenerator()

	tests := []struct {
		name      string
		input     interface{}
		wantError bool
	}{
		{
			name: "valid stations",
			input: []*StationData{
				{StationType: StationAlchemyTable, Name: "Test Table", GenreID: "fantasy", Seed: 1},
				{StationType: StationForge, Name: "Test Forge", GenreID: "fantasy", Seed: 2},
				{StationType: StationWorkbench, Name: "Test Bench", GenreID: "fantasy", Seed: 3},
			},
			wantError: false,
		},
		{
			name:      "wrong type",
			input:     "not a station slice",
			wantError: true,
		},
		{
			name: "wrong count",
			input: []*StationData{
				{StationType: StationAlchemyTable, Name: "Test", GenreID: "fantasy", Seed: 1},
			},
			wantError: true,
		},
		{
			name: "nil station",
			input: []*StationData{
				{StationType: StationAlchemyTable, Name: "Test", GenreID: "fantasy", Seed: 1},
				nil,
				{StationType: StationWorkbench, Name: "Test", GenreID: "fantasy", Seed: 3},
			},
			wantError: true,
		},
		{
			name: "empty name",
			input: []*StationData{
				{StationType: StationAlchemyTable, Name: "", GenreID: "fantasy", Seed: 1},
				{StationType: StationForge, Name: "Test", GenreID: "fantasy", Seed: 2},
				{StationType: StationWorkbench, Name: "Test", GenreID: "fantasy", Seed: 3},
			},
			wantError: true,
		},
		{
			name: "duplicate types",
			input: []*StationData{
				{StationType: StationAlchemyTable, Name: "Test 1", GenreID: "fantasy", Seed: 1},
				{StationType: StationAlchemyTable, Name: "Test 2", GenreID: "fantasy", Seed: 2},
				{StationType: StationWorkbench, Name: "Test 3", GenreID: "fantasy", Seed: 3},
			},
			wantError: true,
		},
		{
			name: "invalid station type",
			input: []*StationData{
				{StationType: StationType(99), Name: "Test", GenreID: "fantasy", Seed: 1},
				{StationType: StationForge, Name: "Test", GenreID: "fantasy", Seed: 2},
				{StationType: StationWorkbench, Name: "Test", GenreID: "fantasy", Seed: 3},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Validate(tt.input)
			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestStationTypeString tests the String method.
func TestStationTypeString(t *testing.T) {
	tests := []struct {
		stationType StationType
		want        string
	}{
		{StationAlchemyTable, "Alchemy Table"},
		{StationForge, "Forge"},
		{StationWorkbench, "Workbench"},
		{StationType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.stationType.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestGenreSpecificNames tests that different genres produce different names.
func TestGenreSpecificNames(t *testing.T) {
	gen := NewStationGenerator()
	seed := int64(12345)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	namesByGenre := make(map[string][]string)

	for _, genre := range genres {
		params := procgen.GenerationParams{
			Difficulty: 0.5,
			Depth:      1,
			GenreID:    genre,
		}

		result, err := gen.Generate(seed, params)
		if err != nil {
			t.Fatalf("Generate failed for genre %q: %v", genre, err)
		}

		stations := result.([]*StationData)
		for _, station := range stations {
			namesByGenre[genre] = append(namesByGenre[genre], station.Name)
		}
	}

	// Check that at least some names differ between genres
	// (with same seed, we should still get genre-appropriate names)
	fantasyNames := namesByGenre["fantasy"]
	scifiNames := namesByGenre["scifi"]

	allSame := true
	for i := range fantasyNames {
		if fantasyNames[i] != scifiNames[i] {
			allSame = false
			break
		}
	}

	if allSame {
		t.Error("fantasy and scifi genres produced identical names")
	}
}

// TestGenerateStationName tests the name generation function.
func TestGenerateStationName(t *testing.T) {
	gen := NewStationGenerator()

	// Test with various templates
	templates := []StationNameTemplate{
		{
			Prefix:    []string{"Ancient", "Mystical"},
			Adjective: []string{"Magical", "Enchanted"},
			Noun:      []string{"Table", "Altar"},
		},
		{
			Prefix:    []string{}, // No prefix
			Adjective: []string{"Flaming"},
			Noun:      []string{"Forge"},
		},
		{
			Prefix:    []string{"Tech"},
			Adjective: []string{}, // No adjective
			Noun:      []string{"Station"},
		},
		{
			Prefix:    []string{},
			Adjective: []string{},
			Noun:      []string{"Workbench"}, // Only noun
		},
	}

	for i, template := range templates {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			// Generate multiple names to test variability
			names := make(map[string]bool)
			for seed := int64(0); seed < 50; seed++ {
				rng := newRNG(seed)
				name := gen.generateStationName(rng, template)

				if name == "" {
					t.Error("generated empty name")
				}

				names[name] = true
			}

			// Should generate at least 2 different names with 50 iterations
			// (unless template is very constrained)
			if len(template.Prefix)+len(template.Adjective)+len(template.Noun) > 3 {
				if len(names) < 2 {
					t.Errorf("only generated %d unique names from 50 iterations", len(names))
				}
			}
		})
	}
}

// TestAllGenresHaveTemplates tests that all genres have complete templates.
func TestAllGenresHaveTemplates(t *testing.T) {
	gen := NewStationGenerator()

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	stationTypes := []StationType{StationAlchemyTable, StationForge, StationWorkbench}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			templates, exists := gen.nameTemplates[genre]
			if !exists {
				t.Fatalf("genre %q not registered", genre)
			}

			for _, stationType := range stationTypes {
				template, exists := templates[stationType]
				if !exists {
					t.Errorf("genre %q missing template for %v", genre, stationType)
					continue
				}

				// Every template must have at least a noun
				if len(template.Noun) == 0 {
					t.Errorf("genre %q, station %v has no nouns", genre, stationType)
				}

				// Check that lists are non-empty where expected
				if len(template.Adjective) == 0 {
					t.Logf("genre %q, station %v has no adjectives (may be intentional)", genre, stationType)
				}
			}
		})
	}
}

// Helper function to create a seeded RNG.
func newRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

// BenchmarkGenerate benchmarks station generation.
func BenchmarkGenerate(b *testing.B) {
	gen := NewStationGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate(int64(i), params)
	}
}

// BenchmarkValidate benchmarks validation.
func BenchmarkValidate(b *testing.B) {
	gen := NewStationGenerator()
	stations := []*StationData{
		{StationType: StationAlchemyTable, Name: "Test Table", GenreID: "fantasy", Seed: 1},
		{StationType: StationForge, Name: "Test Forge", GenreID: "fantasy", Seed: 2},
		{StationType: StationWorkbench, Name: "Test Bench", GenreID: "fantasy", Seed: 3},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gen.Validate(stations)
	}
}
