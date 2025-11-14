package story

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestBranchingNarrativeGenerate(t *testing.T) {
	gen := NewBranchingNarrativeGenerator()
	seed := int64(12345)

	tests := []struct {
		name       string
		params     procgen.GenerationParams
		wantErr    bool
		minChoices int
		maxChoices int
		minPaths   int
		maxPaths   int
	}{
		{
			name: "basic generation",
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr:    false,
			minChoices: 1,
			maxChoices: 3,
			minPaths:   2,
			maxPaths:   8,
		},
		{
			name: "low depth",
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "scifi",
			},
			wantErr:    false,
			minChoices: 1,
			maxChoices: 3,
			minPaths:   2,
			maxPaths:   8,
		},
		{
			name: "high depth",
			params: procgen.GenerationParams{
				Difficulty: 0.8,
				Depth:      20,
				GenreID:    "horror",
			},
			wantErr:    false,
			minChoices: 1,
			maxChoices: 3,
			minPaths:   2,
			maxPaths:   8,
		},
		{
			name: "invalid difficulty low",
			params: procgen.GenerationParams{
				Difficulty: -0.1,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
		{
			name: "invalid difficulty high",
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
			result, err := gen.Generate(seed, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			narrative, ok := result.(*BranchingNarrative)
			if !ok {
				t.Errorf("Generate() returned wrong type")
				return
			}

			// Check choice count
			if len(narrative.ChoicePoints) < tt.minChoices || len(narrative.ChoicePoints) > tt.maxChoices {
				t.Errorf("ChoicePoints count = %d, want between %d and %d", len(narrative.ChoicePoints), tt.minChoices, tt.maxChoices)
			}

			// Check path count
			if len(narrative.Paths) < tt.minPaths || len(narrative.Paths) > tt.maxPaths {
				t.Errorf("Paths count = %d, want between %d and %d", len(narrative.Paths), tt.minPaths, tt.maxPaths)
			}

			// Check common fragments exist
			if len(narrative.CommonFrags) == 0 {
				t.Error("No common fragments generated")
			}

			// Check coherence
			if narrative.Coherence < 0.5 || narrative.Coherence > 1.0 {
				t.Errorf("Coherence = %.2f, want between 0.5 and 1.0", narrative.Coherence)
			}
		})
	}
}

func TestBranchingNarrativeValidate(t *testing.T) {
	gen := NewBranchingNarrativeGenerator()

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name: "valid narrative",
			input: &BranchingNarrative{
				ChoicePoints: []ChoicePoint{{FragmentIndex: 0, Options: []string{"A", "B"}}},
				Paths: []NarrativePath{
					{PathID: "path-0", Outcome: "Success", Fragments: []StoryFragment{{Content: "test"}, {Content: "test2"}}},
					{PathID: "path-1", Outcome: "Failure", Fragments: []StoryFragment{{Content: "test3"}, {Content: "test4"}}},
				},
				CommonFrags: []StoryFragment{{Content: "intro"}},
				Coherence:   0.7,
			},
			wantErr: false,
		},
		{
			name:    "wrong type",
			input:   "not a narrative",
			wantErr: true,
		},
		{
			name: "no choice points",
			input: &BranchingNarrative{
				ChoicePoints: []ChoicePoint{},
				Paths:        []NarrativePath{{PathID: "path-0", Outcome: "Success", Fragments: []StoryFragment{{Content: "test"}}}},
				CommonFrags:  []StoryFragment{{Content: "intro"}},
				Coherence:    0.7,
			},
			wantErr: true,
		},
		{
			name: "too many choice points",
			input: &BranchingNarrative{
				ChoicePoints: []ChoicePoint{{}, {}, {}, {}},
				Paths:        []NarrativePath{{PathID: "path-0", Outcome: "Success", Fragments: []StoryFragment{{Content: "test"}, {Content: "test2"}}}},
				CommonFrags:  []StoryFragment{{Content: "intro"}},
				Coherence:    0.7,
			},
			wantErr: true,
		},
		{
			name: "too few paths",
			input: &BranchingNarrative{
				ChoicePoints: []ChoicePoint{{FragmentIndex: 0, Options: []string{"A", "B"}}},
				Paths:        []NarrativePath{{PathID: "path-0", Outcome: "Success", Fragments: []StoryFragment{{Content: "test"}, {Content: "test2"}}}},
				CommonFrags:  []StoryFragment{{Content: "intro"}},
				Coherence:    0.7,
			},
			wantErr: true,
		},
		{
			name: "low coherence",
			input: &BranchingNarrative{
				ChoicePoints: []ChoicePoint{{FragmentIndex: 0, Options: []string{"A", "B"}}},
				Paths: []NarrativePath{
					{PathID: "path-0", Outcome: "Success", Fragments: []StoryFragment{{Content: "test"}, {Content: "test2"}}},
					{PathID: "path-1", Outcome: "Failure", Fragments: []StoryFragment{{Content: "test3"}, {Content: "test4"}}},
				},
				CommonFrags: []StoryFragment{{Content: "intro"}},
				Coherence:   0.3,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBranchingNarrativeDeterminism(t *testing.T) {
	gen := NewBranchingNarrativeGenerator()
	seed := int64(54321)
	params := procgen.GenerationParams{
		Difficulty: 0.6,
		Depth:      7,
		GenreID:    "cyberpunk",
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

	narrative1 := result1.(*BranchingNarrative)
	narrative2 := result2.(*BranchingNarrative)

	// Check same number of choices
	if len(narrative1.ChoicePoints) != len(narrative2.ChoicePoints) {
		t.Errorf("Different choice counts: %d vs %d", len(narrative1.ChoicePoints), len(narrative2.ChoicePoints))
	}

	// Check same number of paths
	if len(narrative1.Paths) != len(narrative2.Paths) {
		t.Errorf("Different path counts: %d vs %d", len(narrative1.Paths), len(narrative2.Paths))
	}

	// Check coherence matches
	if narrative1.Coherence != narrative2.Coherence {
		t.Errorf("Different coherence: %.4f vs %.4f", narrative1.Coherence, narrative2.Coherence)
	}
}

func TestMakeChoice(t *testing.T) {
	narrative := &BranchingNarrative{
		ChoicePoints: []ChoicePoint{
			{FragmentIndex: 0, Options: []string{"Left", "Right"}, Chosen: -1},
		},
		Paths: []NarrativePath{
			{PathID: "path-0", ChoicesMade: []int{0}},
			{PathID: "path-1", ChoicesMade: []int{1}},
		},
	}

	tests := []struct {
		name         string
		choiceIndex  int
		optionIndex  int
		wantErr      bool
		wantActiveID string
	}{
		{
			name:         "valid choice 0",
			choiceIndex:  0,
			optionIndex:  0,
			wantErr:      false,
			wantActiveID: "path-0",
		},
		{
			name:        "invalid choice index",
			choiceIndex: 5,
			optionIndex: 0,
			wantErr:     true,
		},
		{
			name:        "invalid option index",
			choiceIndex: 0,
			optionIndex: 5,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset choices
			for i := range narrative.ChoicePoints {
				narrative.ChoicePoints[i].Chosen = -1
			}
			narrative.ActivePathID = ""

			err := narrative.MakeChoice(tt.choiceIndex, tt.optionIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeChoice() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && narrative.ActivePathID != tt.wantActiveID {
				t.Errorf("ActivePathID = %s, want %s", narrative.ActivePathID, tt.wantActiveID)
			}
		})
	}
}

func TestGetActiveFragments(t *testing.T) {
	narrative := &BranchingNarrative{
		CommonFrags: []StoryFragment{
			{Content: "Common 1"},
			{Content: "Common 2"},
		},
		Paths: []NarrativePath{
			{
				PathID: "path-0",
				Fragments: []StoryFragment{
					{Content: "Path 0 Fragment"},
				},
			},
		},
		ActivePathID: "",
	}

	// Test with no active path
	frags := narrative.GetActiveFragments()
	if len(frags) != 2 {
		t.Errorf("Expected 2 fragments (common only), got %d", len(frags))
	}

	// Activate a path
	narrative.ActivePathID = "path-0"
	frags = narrative.GetActiveFragments()
	if len(frags) != 3 {
		t.Errorf("Expected 3 fragments (common + path), got %d", len(frags))
	}
}

func TestGetActivePath(t *testing.T) {
	narrative := &BranchingNarrative{
		Paths: []NarrativePath{
			{PathID: "path-0", Outcome: "Success"},
			{PathID: "path-1", Outcome: "Failure"},
		},
		ActivePathID: "",
	}

	// No active path
	path := narrative.GetActivePath()
	if path != nil {
		t.Error("Expected nil for no active path")
	}

	// Active path
	narrative.ActivePathID = "path-1"
	path = narrative.GetActivePath()
	if path == nil {
		t.Fatal("Expected path, got nil")
	}
	if path.PathID != "path-1" {
		t.Errorf("PathID = %s, want path-1", path.PathID)
	}
	if path.Outcome != "Failure" {
		t.Errorf("Outcome = %s, want Failure", path.Outcome)
	}
}

func BenchmarkBranchingNarrativeGenerate(b *testing.B) {
	gen := NewBranchingNarrativeGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate(int64(i), params)
	}
}

func BenchmarkMakeChoice(b *testing.B) {
	narrative := &BranchingNarrative{
		ChoicePoints: []ChoicePoint{
			{FragmentIndex: 0, Options: []string{"A", "B"}, Chosen: -1},
		},
		Paths: []NarrativePath{
			{PathID: "path-0", ChoicesMade: []int{0}},
			{PathID: "path-1", ChoicesMade: []int{1}},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		narrative.ChoicePoints[0].Chosen = -1
		narrative.ActivePathID = ""
		_ = narrative.MakeChoice(0, i%2)
	}
}
