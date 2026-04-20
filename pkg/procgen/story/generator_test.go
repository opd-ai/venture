package story

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestFragmentGenerator(t *testing.T) {
	gen := NewFragmentGenerator()
	seed := int64(12345)

	tests := []struct {
		name    string
		params  procgen.GenerationParams
		wantErr bool
	}{
		{
			name: "valid fantasy story",
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr: false,
		},
		{
			name: "valid scifi story",
			params: procgen.GenerationParams{
				Difficulty: 0.7,
				Depth:      8,
				GenreID:    "scifi",
			},
			wantErr: false,
		},
		{
			name: "invalid difficulty too low",
			params: procgen.GenerationParams{
				Difficulty: -0.1,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
		{
			name: "invalid difficulty too high",
			params: procgen.GenerationParams{
				Difficulty: 1.5,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
		{
			name: "easy difficulty low depth",
			params: procgen.GenerationParams{
				Difficulty: 0.1,
				Depth:      1,
				GenreID:    "horror",
			},
			wantErr: false,
		},
		{
			name: "hard difficulty high depth",
			params: procgen.GenerationParams{
				Difficulty: 0.9,
				Depth:      10,
				GenreID:    "cyberpunk",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.Generate(seed, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			sequence, ok := result.(*StorySequence)
			if !ok {
				t.Errorf("Generate() returned wrong type")
				return
			}

			// Validate basic fields
			if sequence.Title == "" {
				t.Errorf("Title is empty")
			}

			if sequence.Genre != tt.params.GenreID {
				t.Errorf("Genre = %v, want %v", sequence.Genre, tt.params.GenreID)
			}

			if len(sequence.Fragments) < 5 {
				t.Errorf("Too few fragments: %d", len(sequence.Fragments))
			}

			if len(sequence.Fragments) > 15 {
				t.Errorf("Too many fragments: %d", len(sequence.Fragments))
			}

			// Check all fragments have content
			for i, frag := range sequence.Fragments {
				if frag.Content == "" {
					t.Errorf("Fragment %d has empty content", i)
				}
				if frag.SeriesID != sequence.SeriesID {
					t.Errorf("Fragment %d has wrong SeriesID", i)
				}
				if frag.SequenceNum != i {
					t.Errorf("Fragment %d has wrong SequenceNum: got %d, want %d", i, frag.SequenceNum, i)
				}
			}
		})
	}
}

func TestValidate(t *testing.T) {
	gen := NewFragmentGenerator()

	tests := []struct {
		name     string
		sequence *StorySequence
		wantErr  bool
	}{
		{
			name: "valid sequence",
			sequence: &StorySequence{
				SeriesID: "test-123",
				Title:    "Test Story",
				Genre:    "fantasy",
				Fragments: []StoryFragment{
					{Type: FragmentNote, Content: "Fragment 1 content here"},
					{Type: FragmentCarving, Content: "Fragment 2 content here"},
					{Type: FragmentRelic, Content: "Fragment 3 content here"},
					{Type: FragmentCorpse, Content: "Fragment 4 content here"},
					{Type: FragmentBlood, Content: "Fragment 5 content here"},
				},
				Theme:     "mystery",
				Coherence: 0.7,
			},
			wantErr: false,
		},
		{
			name: "empty title",
			sequence: &StorySequence{
				SeriesID: "test-123",
				Title:    "",
				Fragments: []StoryFragment{
					{Content: "Test"},
					{Content: "Test"},
					{Content: "Test"},
					{Content: "Test"},
					{Content: "Test"},
				},
				Coherence: 0.7,
			},
			wantErr: true,
		},
		{
			name: "too few fragments",
			sequence: &StorySequence{
				Title: "Test",
				Fragments: []StoryFragment{
					{Content: "Test"},
					{Content: "Test"},
				},
				Coherence: 0.7,
			},
			wantErr: true,
		},
		{
			name: "low coherence",
			sequence: &StorySequence{
				Title: "Test",
				Fragments: []StoryFragment{
					{Content: "Fragment 1"},
					{Content: "Fragment 2"},
					{Content: "Fragment 3"},
					{Content: "Fragment 4"},
					{Content: "Fragment 5"},
				},
				Coherence: 0.3,
			},
			wantErr: true,
		},
		{
			name: "empty fragment content",
			sequence: &StorySequence{
				Title: "Test",
				Fragments: []StoryFragment{
					{Content: "Fragment 1"},
					{Content: ""},
					{Content: "Fragment 3"},
					{Content: "Fragment 4"},
					{Content: "Fragment 5"},
				},
				Coherence: 0.7,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Validate(tt.sequence)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeterminism(t *testing.T) {
	gen := NewFragmentGenerator()
	seed := int64(99999)
	params := procgen.GenerationParams{
		Difficulty: 0.6,
		Depth:      7,
		GenreID:    "fantasy",
	}

	// Generate twice with same seed
	result1, err1 := gen.Generate(seed, params)
	if err1 != nil {
		t.Fatalf("First generation failed: %v", err1)
	}

	result2, err2 := gen.Generate(seed, params)
	if err2 != nil {
		t.Fatalf("Second generation failed: %v", err2)
	}

	seq1 := result1.(*StorySequence)
	seq2 := result2.(*StorySequence)

	// Verify determinism
	if seq1.Title != seq2.Title {
		t.Errorf("Titles differ: %v vs %v", seq1.Title, seq2.Title)
	}

	if seq1.Theme != seq2.Theme {
		t.Errorf("Themes differ: %v vs %v", seq1.Theme, seq2.Theme)
	}

	if len(seq1.Fragments) != len(seq2.Fragments) {
		t.Errorf("Fragment counts differ: %d vs %d", len(seq1.Fragments), len(seq2.Fragments))
	}

	// Check each fragment
	for i := 0; i < len(seq1.Fragments) && i < len(seq2.Fragments); i++ {
		if seq1.Fragments[i].Content != seq2.Fragments[i].Content {
			t.Errorf("Fragment %d content differs", i)
		}
		if seq1.Fragments[i].Type != seq2.Fragments[i].Type {
			t.Errorf("Fragment %d type differs", i)
		}
	}
}

func TestAllGenres(t *testing.T) {
	gen := NewFragmentGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
	}

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			params.GenreID = genre
			result, err := gen.Generate(12345, params)
			if err != nil {
				t.Fatalf("Generate failed for genre %s: %v", genre, err)
			}

			sequence := result.(*StorySequence)
			if sequence.Genre != genre {
				t.Errorf("Genre = %s, want %s", sequence.Genre, genre)
			}

			// Validate sequence
			if err := gen.Validate(sequence); err != nil {
				t.Errorf("Validation failed for genre %s: %v", genre, err)
			}
		})
	}
}

func TestFragmentTypes(t *testing.T) {
	gen := NewFragmentGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      7,
		GenreID:    "fantasy",
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	sequence := result.(*StorySequence)

	// Count fragment types
	typeCounts := make(map[FragmentType]int)
	for _, frag := range sequence.Fragments {
		typeCounts[frag.Type]++
	}

	// Verify we have variety (at least 2 different types)
	if len(typeCounts) < 2 {
		t.Errorf("Insufficient fragment type variety: %d types", len(typeCounts))
	}

	// Verify all fragment types are valid
	for fragType := range typeCounts {
		if fragType < 0 || fragType > 5 {
			t.Errorf("Invalid fragment type: %d", fragType)
		}
	}
}

func TestFragmentSequencing(t *testing.T) {
	gen := NewFragmentGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      7,
		GenreID:    "fantasy",
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	sequence := result.(*StorySequence)

	// Verify fragments are in sequence order
	for i, frag := range sequence.Fragments {
		if frag.SequenceNum != i {
			t.Errorf("Fragment %d has SequenceNum %d, want %d", i, frag.SequenceNum, i)
		}
	}

	// Verify all fragments share same SeriesID
	for i, frag := range sequence.Fragments {
		if frag.SeriesID != sequence.SeriesID {
			t.Errorf("Fragment %d has wrong SeriesID", i)
		}
	}
}

func TestFragmentLocations(t *testing.T) {
	gen := NewFragmentGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      7,
		GenreID:    "fantasy",
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	sequence := result.(*StorySequence)

	// Verify all fragments have valid locations
	for i, frag := range sequence.Fragments {
		if frag.Location.X < 0 || frag.Location.X > 100 {
			t.Errorf("Fragment %d X location out of bounds: %.2f", i, frag.Location.X)
		}
		if frag.Location.Y < 0 || frag.Location.Y > 100 {
			t.Errorf("Fragment %d Y location out of bounds: %.2f", i, frag.Location.Y)
		}
	}

	// Verify locations are distributed (not all at same spot)
	firstLoc := sequence.Fragments[0].Location
	allSame := true
	for _, frag := range sequence.Fragments[1:] {
		if frag.Location.X != firstLoc.X || frag.Location.Y != firstLoc.Y {
			allSame = false
			break
		}
	}

	if allSame {
		t.Errorf("All fragments at same location")
	}
}

func BenchmarkGenerate(b *testing.B) {
	gen := NewFragmentGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(int64(i), params)
		if err != nil {
			b.Fatalf("Generate failed: %v", err)
		}
	}
}

func BenchmarkValidate(b *testing.B) {
	gen := NewFragmentGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		b.Fatalf("Generate failed: %v", err)
	}

	sequence := result.(*StorySequence)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := gen.Validate(sequence)
		if err != nil {
			b.Fatalf("Validate failed: %v", err)
		}
	}
}

// TestDefaultStoryTemplates validates the default template configuration
func TestDefaultStoryTemplates(t *testing.T) {
	templates := DefaultStoryTemplates()

	// Validate beginning templates
	if len(templates.BeginningTemplates) == 0 {
		t.Error("BeginningTemplates is empty")
	}
	if len(templates.BeginningAdjectives) == 0 {
		t.Error("BeginningAdjectives is empty")
	}
	if len(templates.BeginningCounts) == 0 {
		t.Error("BeginningCounts is empty")
	}
	if len(templates.BeginningDiscoveries) == 0 {
		t.Error("BeginningDiscoveries is empty")
	}

	// Validate middle templates
	if len(templates.MiddleTemplates) == 0 {
		t.Error("MiddleTemplates is empty")
	}
	if len(templates.MiddleLosses) == 0 {
		t.Error("MiddleLosses is empty")
	}
	if len(templates.MiddleThreats) == 0 {
		t.Error("MiddleThreats is empty")
	}
	if len(templates.MiddleRevelations) == 0 {
		t.Error("MiddleRevelations is empty")
	}
	if len(templates.MiddleAttackers) == 0 {
		t.Error("MiddleAttackers is empty")
	}
	if len(templates.MiddleSurvivors) == 0 {
		t.Error("MiddleSurvivors is empty")
	}

	// Validate end templates
	if len(templates.EndTemplates) == 0 {
		t.Error("EndTemplates is empty")
	}
	if len(templates.EndWarnings) == 0 {
		t.Error("EndWarnings is empty")
	}
	if len(templates.EndGoals) == 0 {
		t.Error("EndGoals is empty")
	}
	if len(templates.EndThreats) == 0 {
		t.Error("EndThreats is empty")
	}
	if len(templates.EndFates) == 0 {
		t.Error("EndFates is empty")
	}
	if len(templates.EndMessages) == 0 {
		t.Error("EndMessages is empty")
	}
}

// TestNewFragmentGeneratorWithTemplates validates custom template injection
func TestNewFragmentGeneratorWithTemplates(t *testing.T) {
	tests := []struct {
		name      string
		templates *StoryTemplates
		wantNil   bool
	}{
		{
			name: "custom templates",
			templates: &StoryTemplates{
				BeginningTemplates:   []string{"Custom beginning: %d"},
				BeginningAdjectives:  []string{"custom"},
				BeginningCounts:      []int{1},
				BeginningDiscoveries: []string{"custom discovery"},
				MiddleTemplates:      []string{"Custom middle: %d"},
				MiddleLosses:         []string{"custom loss"},
				MiddleThreats:        []string{"custom threat"},
				MiddleRevelations:    []string{"custom revelation"},
				MiddleAttackers:      []string{"custom attacker"},
				MiddleSurvivors:      []int{1},
				EndTemplates:         []string{"Custom end: %s"},
				EndWarnings:          []string{"custom warning"},
				EndGoals:             []string{"custom goal"},
				EndThreats:           []string{"custom threat"},
				EndFates:             []string{"custom fate"},
				EndMessages:          []string{"custom message"},
			},
			wantNil: false,
		},
		{
			name:      "nil templates should use defaults",
			templates: nil,
			wantNil:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewFragmentGeneratorWithTemplates(tt.templates)
			if gen == nil {
				t.Fatal("NewFragmentGeneratorWithTemplates returned nil")
			}
			if gen.templates == nil {
				t.Fatal("generator templates field is nil")
			}

			// Verify defaults are used when nil is passed
			if tt.templates == nil {
				defaults := DefaultStoryTemplates()
				if len(gen.templates.BeginningTemplates) != len(defaults.BeginningTemplates) {
					t.Error("nil templates did not default to DefaultStoryTemplates")
				}
			}
		})
	}
}

// TestCustomTemplateGeneration validates that custom templates are used
func TestCustomTemplateGeneration(t *testing.T) {
	customTemplates := &StoryTemplates{
		BeginningTemplates:   []string{"CUSTOM_BEGINNING_%d_%s_%d_%s"},
		BeginningAdjectives:  []string{"TEST_ADJ"},
		BeginningCounts:      []int{99},
		BeginningDiscoveries: []string{"TEST_DISC"},
		MiddleTemplates:      []string{"CUSTOM_MIDDLE_%d_%s_%s_%s_%s_%d"},
		MiddleLosses:         []string{"TEST_LOSS"},
		MiddleThreats:        []string{"TEST_THREAT"},
		MiddleRevelations:    []string{"TEST_REV"},
		MiddleAttackers:      []string{"TEST_ATK"},
		MiddleSurvivors:      []int{88},
		EndTemplates:         []string{"CUSTOM_END_%s_%s_%s_%s_%s"},
		EndWarnings:          []string{"TEST_WARN"},
		EndGoals:             []string{"TEST_GOAL"},
		EndThreats:           []string{"TEST_THREAT"},
		EndFates:             []string{"TEST_FATE"},
		EndMessages:          []string{"TEST_MSG"},
	}

	gen := NewFragmentGeneratorWithTemplates(customTemplates)

	seed := int64(12345)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	result, err := gen.Generate(seed, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	sequence := result.(*StorySequence)

	// Verify at least one fragment uses custom templates
	foundCustom := false
	for _, frag := range sequence.Fragments {
		if len(frag.Content) > 6 && frag.Content[:6] == "CUSTOM" {
			foundCustom = true
			break
		}
	}

	if !foundCustom {
		t.Error("custom templates were not used in generation")
	}
}

// TestTemplateDeterminism validates same seed produces same templates
func TestTemplateDeterminism(t *testing.T) {
	seed := int64(99999)
	params := procgen.GenerationParams{
		Difficulty: 0.6,
		Depth:      7,
		GenreID:    "scifi",
	}

	// Generate twice with same seed
	gen1 := NewFragmentGenerator()
	result1, err1 := gen1.Generate(seed, params)
	if err1 != nil {
		t.Fatalf("Generate 1 failed: %v", err1)
	}

	gen2 := NewFragmentGenerator()
	result2, err2 := gen2.Generate(seed, params)
	if err2 != nil {
		t.Fatalf("Generate 2 failed: %v", err2)
	}

	seq1 := result1.(*StorySequence)
	seq2 := result2.(*StorySequence)

	// Verify same number of fragments
	if len(seq1.Fragments) != len(seq2.Fragments) {
		t.Fatalf("fragment count differs: %d vs %d", len(seq1.Fragments), len(seq2.Fragments))
	}

	// Verify each fragment is identical
	for i := range seq1.Fragments {
		if seq1.Fragments[i].Content != seq2.Fragments[i].Content {
			t.Errorf("fragment %d content differs:\n  got: %s\n want: %s",
				i, seq1.Fragments[i].Content, seq2.Fragments[i].Content)
		}
	}
}

// TestSelectTheme_AllGenresCovered verifies that genreThemes and genreTitleSuffixes
// contain entries for every known genre and that each produces non-empty results.
func TestSelectTheme_AllGenresCovered(t *testing.T) {
genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapocalyptic"}
for _, g := range genres {
t.Run(g, func(t *testing.T) {
themes, ok := genreThemes[g]
if !ok || len(themes) == 0 {
t.Errorf("genreThemes[%q] is missing or empty", g)
}
suffixes, ok := genreTitleSuffixes[g]
if !ok || len(suffixes) == 0 {
t.Errorf("genreTitleSuffixes[%q] is missing or empty", g)
}
})
}
}
