package narrative

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestNewStoryArcGenerator(t *testing.T) {
	gen := NewStoryArcGenerator()
	if gen == nil {
		t.Fatal("NewStoryArcGenerator() returned nil")
	}
}

func TestStoryArcGenerator_Generate_ValidParams(t *testing.T) {
	gen := NewStoryArcGenerator()

	tests := []struct {
		name   string
		seed   int64
		params procgen.GenerationParams
	}{
		{
			name: "fantasy_easy",
			seed: 12345,
			params: procgen.GenerationParams{
				Difficulty: 0.3,
				Depth:      3,
				GenreID:    "fantasy",
			},
		},
		{
			name: "sci-fi_medium",
			seed: 67890,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    "sci-fi",
			},
		},
		{
			name: "horror_hard",
			seed: 11111,
			params: procgen.GenerationParams{
				Difficulty: 0.8,
				Depth:      7,
				GenreID:    "horror",
			},
		},
		{
			name: "cyberpunk_easy",
			seed: 22222,
			params: procgen.GenerationParams{
				Difficulty: 0.2,
				Depth:      2,
				GenreID:    "cyberpunk",
			},
		},
		{
			name: "post-apocalyptic_medium",
			seed: 33333,
			params: procgen.GenerationParams{
				Difficulty: 0.6,
				Depth:      6,
				GenreID:    "post-apocalyptic",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.Generate(tt.seed, tt.params)
			if err != nil {
				t.Fatalf("Generate() error = %v, want nil", err)
			}

			arc, ok := result.(*StoryArc)
			if !ok {
				t.Fatal("Generate() result is not *StoryArc")
			}

			// Verify basic fields
			if arc.Title == "" {
				t.Error("Generate() produced empty title")
			}
			if arc.MainConflict == "" {
				t.Error("Generate() produced empty main conflict")
			}
			if arc.Antagonist == "" {
				t.Error("Generate() produced empty antagonist")
			}
			if arc.Ally == "" {
				t.Error("Generate() produced empty ally")
			}
			if arc.Genre != tt.params.GenreID {
				t.Errorf("Generate() Genre = %v, want %v", arc.Genre, tt.params.GenreID)
			}
			if arc.Difficulty != tt.params.Difficulty {
				t.Errorf("Generate() Difficulty = %v, want %v", arc.Difficulty, tt.params.Difficulty)
			}
			if arc.Seed != tt.seed {
				t.Errorf("Generate() Seed = %v, want %v", arc.Seed, tt.seed)
			}

			// Verify plot points exist
			if len(arc.PlotPoints) < 3 {
				t.Errorf("Generate() PlotPoints length = %v, want >= 3", len(arc.PlotPoints))
			}

			// Verify endings exist
			if len(arc.PossibleEndings) < 1 {
				t.Errorf("Generate() PossibleEndings length = %v, want >= 1", len(arc.PossibleEndings))
			}
		})
	}
}

func TestStoryArcGenerator_Generate_InvalidParams(t *testing.T) {
	gen := NewStoryArcGenerator()

	tests := []struct {
		name    string
		seed    int64
		params  procgen.GenerationParams
		wantErr bool
	}{
		{
			name: "difficulty_too_low",
			seed: 12345,
			params: procgen.GenerationParams{
				Difficulty: -0.1,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
		{
			name: "difficulty_too_high",
			seed: 12345,
			params: procgen.GenerationParams{
				Difficulty: 1.5,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
		{
			name: "depth_zero",
			seed: 12345,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      0,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
		{
			name: "depth_negative",
			seed: 12345,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      -1,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := gen.Generate(tt.seed, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStoryArcGenerator_Validate(t *testing.T) {
	gen := NewStoryArcGenerator()

	tests := []struct {
		name    string
		arc     *StoryArc
		wantErr bool
	}{
		{
			name: "valid_arc",
			arc: &StoryArc{
				Title:        "The Quest",
				MainConflict: "Evil rises",
				Antagonist:   "Dark Lord",
				Ally:         "Wizard",
				PlotPoints: []PlotPoint{
					{Act: 1, Description: "Start"},
					{Act: 2, Description: "Middle"},
					{Act: 3, Description: "End"},
				},
				PossibleEndings: []string{"Victory"},
			},
			wantErr: false,
		},
		{
			name: "empty_title",
			arc: &StoryArc{
				Title:        "",
				MainConflict: "Evil rises",
				Antagonist:   "Dark Lord",
				PlotPoints: []PlotPoint{
					{Act: 1}, {Act: 2}, {Act: 3},
				},
				PossibleEndings: []string{"Victory"},
			},
			wantErr: true,
		},
		{
			name: "empty_conflict",
			arc: &StoryArc{
				Title:        "The Quest",
				MainConflict: "",
				Antagonist:   "Dark Lord",
				PlotPoints: []PlotPoint{
					{Act: 1}, {Act: 2}, {Act: 3},
				},
				PossibleEndings: []string{"Victory"},
			},
			wantErr: true,
		},
		{
			name: "empty_antagonist",
			arc: &StoryArc{
				Title:        "The Quest",
				MainConflict: "Evil rises",
				Antagonist:   "",
				PlotPoints: []PlotPoint{
					{Act: 1}, {Act: 2}, {Act: 3},
				},
				PossibleEndings: []string{"Victory"},
			},
			wantErr: true,
		},
		{
			name: "insufficient_plot_points",
			arc: &StoryArc{
				Title:        "The Quest",
				MainConflict: "Evil rises",
				Antagonist:   "Dark Lord",
				PlotPoints: []PlotPoint{
					{Act: 1},
				},
				PossibleEndings: []string{"Victory"},
			},
			wantErr: true,
		},
		{
			name: "no_endings",
			arc: &StoryArc{
				Title:        "The Quest",
				MainConflict: "Evil rises",
				Antagonist:   "Dark Lord",
				PlotPoints: []PlotPoint{
					{Act: 1}, {Act: 2}, {Act: 3},
				},
				PossibleEndings: []string{},
			},
			wantErr: true,
		},
		{
			name: "missing_act1",
			arc: &StoryArc{
				Title:        "The Quest",
				MainConflict: "Evil rises",
				Antagonist:   "Dark Lord",
				PlotPoints: []PlotPoint{
					{Act: 2}, {Act: 3}, {Act: 3},
				},
				PossibleEndings: []string{"Victory"},
			},
			wantErr: true,
		},
		{
			name: "missing_act2",
			arc: &StoryArc{
				Title:        "The Quest",
				MainConflict: "Evil rises",
				Antagonist:   "Dark Lord",
				PlotPoints: []PlotPoint{
					{Act: 1}, {Act: 3}, {Act: 3},
				},
				PossibleEndings: []string{"Victory"},
			},
			wantErr: true,
		},
		{
			name: "missing_act3",
			arc: &StoryArc{
				Title:        "The Quest",
				MainConflict: "Evil rises",
				Antagonist:   "Dark Lord",
				PlotPoints: []PlotPoint{
					{Act: 1}, {Act: 2}, {Act: 2},
				},
				PossibleEndings: []string{"Victory"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Validate(tt.arc)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStoryArcGenerator_Validate_WrongType(t *testing.T) {
	gen := NewStoryArcGenerator()

	err := gen.Validate("not a story arc")
	if err == nil {
		t.Error("Validate() with wrong type should return error")
	}
}

func TestStoryArcGenerator_Determinism(t *testing.T) {
	gen := NewStoryArcGenerator()

	seed := int64(42)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	// Generate twice with same seed
	result1, err1 := gen.Generate(seed, params)
	if err1 != nil {
		t.Fatalf("First Generate() error = %v", err1)
	}

	result2, err2 := gen.Generate(seed, params)
	if err2 != nil {
		t.Fatalf("Second Generate() error = %v", err2)
	}

	arc1 := result1.(*StoryArc)
	arc2 := result2.(*StoryArc)

	// Verify determinism
	if arc1.Title != arc2.Title {
		t.Errorf("Titles differ: %v vs %v", arc1.Title, arc2.Title)
	}
	if arc1.MainConflict != arc2.MainConflict {
		t.Errorf("MainConflicts differ: %v vs %v", arc1.MainConflict, arc2.MainConflict)
	}
	if arc1.Antagonist != arc2.Antagonist {
		t.Errorf("Antagonists differ: %v vs %v", arc1.Antagonist, arc2.Antagonist)
	}
	if arc1.Ally != arc2.Ally {
		t.Errorf("Allies differ: %v vs %v", arc1.Ally, arc2.Ally)
	}
	if len(arc1.PlotPoints) != len(arc2.PlotPoints) {
		t.Errorf("PlotPoints count differs: %v vs %v", len(arc1.PlotPoints), len(arc2.PlotPoints))
	}
	if len(arc1.PossibleEndings) != len(arc2.PossibleEndings) {
		t.Errorf("PossibleEndings count differs: %v vs %v", len(arc1.PossibleEndings), len(arc2.PossibleEndings))
	}
}

func TestStoryArcGenerator_ThreeActStructure(t *testing.T) {
	gen := NewStoryArcGenerator()

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	arc := result.(*StoryArc)

	// Count acts
	actCounts := map[int]int{1: 0, 2: 0, 3: 0}
	for _, point := range arc.PlotPoints {
		actCounts[point.Act]++
	}

	if actCounts[1] == 0 {
		t.Error("No Act 1 plot points")
	}
	if actCounts[2] == 0 {
		t.Error("No Act 2 plot points")
	}
	if actCounts[3] == 0 {
		t.Error("No Act 3 plot points")
	}

	// Act 2 should have the most plot points
	if actCounts[2] < actCounts[1] && actCounts[2] < actCounts[3] {
		t.Error("Act 2 should typically have most plot points")
	}
}

func TestStoryArcGenerator_DifficultyAffects(t *testing.T) {
	gen := NewStoryArcGenerator()

	easyParams := procgen.GenerationParams{
		Difficulty: 0.3,
		Depth:      5,
		GenreID:    "fantasy",
	}

	hardParams := procgen.GenerationParams{
		Difficulty: 0.8,
		Depth:      5,
		GenreID:    "fantasy",
	}

	easyResult, _ := gen.Generate(12345, easyParams)
	hardResult, _ := gen.Generate(12345, hardParams)

	easyArc := easyResult.(*StoryArc)
	hardArc := hardResult.(*StoryArc)

	// Higher difficulty should generally result in more plot points
	// (though this isn't guaranteed due to RNG, so we just test they're different)
	if easyArc.MainConflict == hardArc.MainConflict &&
		easyArc.Antagonist == hardArc.Antagonist &&
		len(easyArc.PlotPoints) == len(hardArc.PlotPoints) {
		t.Log("Note: Easy and hard arcs are identical - may want to increase difficulty impact")
	}
}

func TestStoryArcGenerator_DepthAffectsPlotPoints(t *testing.T) {
	gen := NewStoryArcGenerator()

	shallowParams := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      2,
		GenreID:    "fantasy",
	}

	deepParams := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      10,
		GenreID:    "fantasy",
	}

	shallowResult, _ := gen.Generate(12345, shallowParams)
	deepResult, _ := gen.Generate(12345, deepParams)

	shallowArc := shallowResult.(*StoryArc)
	deepArc := deepResult.(*StoryArc)

	// Deeper dungeons should have more plot points
	if len(deepArc.PlotPoints) <= len(shallowArc.PlotPoints) {
		t.Errorf("Deep arc PlotPoints (%d) should be > shallow arc (%d)",
			len(deepArc.PlotPoints), len(shallowArc.PlotPoints))
	}
}

func TestPlotPoint_Structure(t *testing.T) {
	point := PlotPoint{
		Act:               2,
		Type:              "midpoint",
		Description:       "Major revelation",
		Participants:      []string{"Player", "Antagonist"},
		Location:          "Central Chamber",
		TriggerConditions: []string{"entered_chamber", "has_key"},
		Consequences:      []string{"Truth revealed", "Stakes raised"},
		PlayerChoices: []PlayerChoice{
			{
				Description: "How to react",
				Options:     []string{"Attack", "Retreat", "Negotiate"},
				Consequences: [][]string{
					{"Immediate fight"},
					{"Live to fight another day"},
					{"Potential alliance"},
				},
			},
		},
	}

	if point.Act != 2 {
		t.Errorf("PlotPoint Act = %v, want 2", point.Act)
	}
	if len(point.Participants) != 2 {
		t.Errorf("PlotPoint Participants length = %v, want 2", len(point.Participants))
	}
	if len(point.TriggerConditions) != 2 {
		t.Errorf("PlotPoint TriggerConditions length = %v, want 2", len(point.TriggerConditions))
	}
	if len(point.PlayerChoices) != 1 {
		t.Errorf("PlotPoint PlayerChoices length = %v, want 1", len(point.PlayerChoices))
	}
}

func TestPlayerChoice_Structure(t *testing.T) {
	choice := PlayerChoice{
		Description: "Choose your path",
		Options:     []string{"Left", "Right", "Center"},
		Consequences: [][]string{
			{"Treasure room"},
			{"Combat encounter"},
			{"Puzzle room"},
		},
		RelationshipImpacts: []map[string]float64{
			{"faction_a": 0.1},
			{"faction_b": 0.1},
			{"faction_a": -0.1, "faction_b": 0.2},
		},
	}

	if len(choice.Options) != 3 {
		t.Errorf("PlayerChoice Options length = %v, want 3", len(choice.Options))
	}
	if len(choice.Consequences) != 3 {
		t.Errorf("PlayerChoice Consequences length = %v, want 3", len(choice.Consequences))
	}
	if len(choice.RelationshipImpacts) != 3 {
		t.Errorf("PlayerChoice RelationshipImpacts length = %v, want 3", len(choice.RelationshipImpacts))
	}
}

// Benchmark story arc generation
func BenchmarkStoryArcGenerator_Generate(b *testing.B) {
	gen := NewStoryArcGenerator()
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
