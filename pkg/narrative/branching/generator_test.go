package branching

import (
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestGeneratorGenerate(t *testing.T) {
	tests := []struct {
		name    string
		seed    int64
		params  procgen.GenerationParams
		wantErr bool
	}{
		{
			name: "valid fantasy arc",
			seed: 12345,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr: false,
		},
		{
			name: "valid scifi arc",
			seed: 67890,
			params: procgen.GenerationParams{
				Difficulty: 0.7,
				Depth:      10,
				GenreID:    "scifi",
			},
			wantErr: false,
		},
		{
			name: "minimum depth",
			seed: 11111,
			params: procgen.GenerationParams{
				Difficulty: 0.3,
				Depth:      1,
				GenreID:    "horror",
			},
			wantErr: false,
		},
		{
			name: "invalid depth",
			seed: 22222,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      0,
				GenreID:    "cyberpunk",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator()
			result, err := gen.Generate(tt.seed, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				arc, ok := result.(*StoryArc)
				if !ok {
					t.Errorf("Generate() returned type %T, want *StoryArc", result)
					return
				}

				if arc.ID == "" {
					t.Error("Generated arc has empty ID")
				}

				if arc.GenreID != tt.params.GenreID {
					t.Errorf("Arc genre = %s, want %s", arc.GenreID, tt.params.GenreID)
				}

				if len(arc.Nodes) < 10 {
					t.Errorf("Arc has %d nodes, want at least 10", len(arc.Nodes))
				}

				if len(arc.Endings) < 1 {
					t.Errorf("Arc has %d endings, want at least 1", len(arc.Endings))
				}
			}
		})
	}
}

func TestGeneratorValidate(t *testing.T) {
	gen := NewGenerator()

	// Generate a valid arc
	result, err := gen.Generate(12345, procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	})
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	arc := result.(*StoryArc)

	tests := []struct {
		name    string
		modify  func(*StoryArc)
		wantErr bool
	}{
		{
			name:    "valid arc",
			modify:  func(a *StoryArc) {},
			wantErr: false,
		},
		{
			name: "empty ID",
			modify: func(a *StoryArc) {
				a.ID = ""
			},
			wantErr: true,
		},
		{
			name: "empty start node",
			modify: func(a *StoryArc) {
				a.StartNodeID = ""
			},
			wantErr: true,
		},
		{
			name: "too few nodes",
			modify: func(a *StoryArc) {
				a.Nodes = make(map[string]*StoryNode)
				for i := 0; i < 5; i++ {
					id := "node_" + string(rune(i))
					a.Nodes[id] = &StoryNode{ID: id}
				}
			},
			wantErr: true,
		},
		{
			name: "no endings",
			modify: func(a *StoryArc) {
				a.Endings = make(map[string]EndingType)
			},
			wantErr: true,
		},
		{
			name: "start node not found",
			modify: func(a *StoryArc) {
				a.StartNodeID = "nonexistent"
			},
			wantErr: true,
		},
		{
			name: "ending node not found",
			modify: func(a *StoryArc) {
				a.Endings["nonexistent"] = EndingTypeHeroic
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the arc
			testArc := &StoryArc{
				ID:          arc.ID,
				Title:       arc.Title,
				Description: arc.Description,
				GenreID:     arc.GenreID,
				StartNodeID: arc.StartNodeID,
				Nodes:       make(map[string]*StoryNode),
				Endings:     make(map[string]EndingType),
				Seed:        arc.Seed,
			}

			for k, v := range arc.Nodes {
				testArc.Nodes[k] = v
			}
			for k, v := range arc.Endings {
				testArc.Endings[k] = v
			}

			tt.modify(testArc)

			err := gen.Validate(testArc)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGeneratorDeterminism(t *testing.T) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	// Generate twice with same seed
	result1, err1 := gen.Generate(12345, params)
	result2, err2 := gen.Generate(12345, params)

	if err1 != nil || err2 != nil {
		t.Fatalf("Generate() failed: %v, %v", err1, err2)
	}

	arc1 := result1.(*StoryArc)
	arc2 := result2.(*StoryArc)

	if arc1.Title != arc2.Title {
		t.Errorf("Titles differ: %s vs %s", arc1.Title, arc2.Title)
	}

	if len(arc1.Nodes) != len(arc2.Nodes) {
		t.Errorf("Node count differs: %d vs %d", len(arc1.Nodes), len(arc2.Nodes))
	}

	if len(arc1.Endings) != len(arc2.Endings) {
		t.Errorf("Ending count differs: %d vs %d", len(arc1.Endings), len(arc2.Endings))
	}
}

func TestNodeTypeString(t *testing.T) {
	tests := []struct {
		nodeType NodeType
		want     string
	}{
		{NodeTypeStart, "Start"},
		{NodeTypeChoice, "Choice"},
		{NodeTypeEvent, "Event"},
		{NodeTypeConsequence, "Consequence"},
		{NodeTypeEnding, "Ending"},
		{NodeType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.nodeType.String()
			if got != tt.want {
				t.Errorf("NodeType.String() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestEndingTypeString(t *testing.T) {
	tests := []struct {
		endingType EndingType
		want       string
	}{
		{EndingTypeHeroic, "Heroic"},
		{EndingTypeTragic, "Tragic"},
		{EndingTypeNeutral, "Neutral"},
		{EndingTypeMystery, "Mystery"},
		{EndingTypeTriumph, "Triumph"},
		{EndingTypeBetrayal, "Betrayal"},
		{EndingType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.endingType.String()
			if got != tt.want {
				t.Errorf("EndingType.String() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestGenerateTitle(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			// Same seed should produce same title
			title1 := generateTitle(newRNG(12345), genre)
			title2 := generateTitle(newRNG(12345), genre)

			if title1 != title2 {
				t.Errorf("Titles differ for same seed: %s vs %s", title1, title2)
			}

			if title1 == "" {
				t.Error("Generated empty title")
			}
		})
	}
}

func TestGenerateAlignmentShift(t *testing.T) {
	rng := newRNG(12345)

	for i := 0; i < 100; i++ {
		shift := generateAlignmentShift(rng)

		for axis, value := range shift {
			if value < -0.2 || value > 0.2 {
				t.Errorf("Alignment shift for %s out of range: %f", axis, value)
			}
		}
	}
}

func TestGenerateFactionChange(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			rng := newRNG(12345)
			change := generateFactionChange(rng, genre)

			for _, value := range change {
				if value < -0.2 || value > 0.2 {
					t.Errorf("Faction change out of range: %f", value)
				}
			}
		})
	}
}

func TestNarrativeComponentType(t *testing.T) {
	component := NarrativeComponent{}
	if component.Type() != "narrative" {
		t.Errorf("Component type = %s, want narrative", component.Type())
	}
}

// Benchmarks

func BenchmarkGenerate(b *testing.B) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Generate(int64(i), params)
	}
}

func BenchmarkValidate(b *testing.B) {
	gen := NewGenerator()
	result, _ := gen.Generate(12345, procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	})
	arc := result.(*StoryArc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Validate(arc)
	}
}

// Helper for tests
func newRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}
