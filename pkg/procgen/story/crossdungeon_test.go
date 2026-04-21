package story

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestCrossDungeonGenerate(t *testing.T) {
	gen := NewCrossDungeonGenerator()
	seed := int64(67890)

	tests := []struct {
		name         string
		params       procgen.GenerationParams
		wantErr      bool
		minLevelSpan int
		maxLevelSpan int
		minFrags     int
	}{
		{
			name: "basic generation",
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      3,
				GenreID:    "fantasy",
			},
			wantErr:      false,
			minLevelSpan: 2,
			maxLevelSpan: 5,
			minFrags:     4,
		},
		{
			name: "high difficulty",
			params: procgen.GenerationParams{
				Difficulty: 0.9,
				Depth:      5,
				GenreID:    "scifi",
			},
			wantErr:      false,
			minLevelSpan: 2,
			maxLevelSpan: 5,
			minFrags:     4,
		},
		{
			name: "low difficulty",
			params: procgen.GenerationParams{
				Difficulty: 0.1,
				Depth:      2,
				GenreID:    "horror",
			},
			wantErr:      false,
			minLevelSpan: 2,
			maxLevelSpan: 5,
			minFrags:     4,
		},
		{
			name: "invalid difficulty low",
			params: procgen.GenerationParams{
				Difficulty: -0.5,
				Depth:      3,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
		{
			name: "invalid difficulty high",
			params: procgen.GenerationParams{
				Difficulty: 2.0,
				Depth:      3,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
		{
			name: "invalid depth",
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      0,
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

			story, ok := result.(*CrossDungeonStory)
			if !ok {
				t.Errorf("Generate() returned wrong type")
				return
			}

			// Check level span
			if story.LevelSpan < tt.minLevelSpan || story.LevelSpan > tt.maxLevelSpan {
				t.Errorf("LevelSpan = %d, want between %d and %d", story.LevelSpan, tt.minLevelSpan, tt.maxLevelSpan)
			}

			// Check fragment count
			if len(story.Fragments) < tt.minFrags {
				t.Errorf("Fragment count = %d, want at least %d", len(story.Fragments), tt.minFrags)
			}

			// Check depth range
			if story.MinDepth != tt.params.Depth {
				t.Errorf("MinDepth = %d, want %d", story.MinDepth, tt.params.Depth)
			}

			if story.MaxDepth <= story.MinDepth {
				t.Errorf("MaxDepth (%d) should be > MinDepth (%d)", story.MaxDepth, story.MinDepth)
			}

			// Check coherence and continuity
			if story.Coherence < 0.5 || story.Coherence > 1.0 {
				t.Errorf("Coherence = %.2f, want between 0.5 and 1.0", story.Coherence)
			}

			if story.Continuity < 0.5 || story.Continuity > 1.0 {
				t.Errorf("Continuity = %.2f, want between 0.5 and 1.0", story.Continuity)
			}
		})
	}
}

func TestCrossDungeonValidate(t *testing.T) {
	gen := NewCrossDungeonGenerator()

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name: "valid story",
			input: &CrossDungeonStory{
				Title:     "Test Story",
				LevelSpan: 3,
				MinDepth:  1,
				MaxDepth:  3,
				Fragments: []CrossDungeonFragment{
					{StoryFragment: StoryFragment{Content: "frag1"}, Level: DungeonLevel{Depth: 1}},
					{StoryFragment: StoryFragment{Content: "frag2"}, Level: DungeonLevel{Depth: 2}},
					{StoryFragment: StoryFragment{Content: "frag3"}, Level: DungeonLevel{Depth: 2}},
					{StoryFragment: StoryFragment{Content: "frag4"}, Level: DungeonLevel{Depth: 3}},
					{StoryFragment: StoryFragment{Content: "frag5"}, Level: DungeonLevel{Depth: 3}},
					{StoryFragment: StoryFragment{Content: "frag6"}, Level: DungeonLevel{Depth: 3}},
				},
				Coherence:  0.7,
				Continuity: 0.8,
			},
			wantErr: false,
		},
		{
			name:    "wrong type",
			input:   "not a story",
			wantErr: true,
		},
		{
			name: "empty title",
			input: &CrossDungeonStory{
				Title:      "",
				LevelSpan:  2,
				MinDepth:   1,
				MaxDepth:   2,
				Fragments:  []CrossDungeonFragment{{StoryFragment: StoryFragment{Content: "test"}}, {StoryFragment: StoryFragment{Content: "test2"}}},
				Coherence:  0.7,
				Continuity: 0.8,
			},
			wantErr: true,
		},
		{
			name: "level span too small",
			input: &CrossDungeonStory{
				Title:      "Test",
				LevelSpan:  1,
				MinDepth:   1,
				MaxDepth:   1,
				Fragments:  []CrossDungeonFragment{{StoryFragment: StoryFragment{Content: "test"}}, {StoryFragment: StoryFragment{Content: "test2"}}},
				Coherence:  0.7,
				Continuity: 0.8,
			},
			wantErr: true,
		},
		{
			name: "level span too large",
			input: &CrossDungeonStory{
				Title:      "Test",
				LevelSpan:  6,
				MinDepth:   1,
				MaxDepth:   6,
				Fragments:  make([]CrossDungeonFragment, 12),
				Coherence:  0.7,
				Continuity: 0.8,
			},
			wantErr: true,
		},
		{
			name: "too few fragments",
			input: &CrossDungeonStory{
				Title:      "Test",
				LevelSpan:  3,
				MinDepth:   1,
				MaxDepth:   3,
				Fragments:  []CrossDungeonFragment{{StoryFragment: StoryFragment{Content: "test"}}},
				Coherence:  0.7,
				Continuity: 0.8,
			},
			wantErr: true,
		},
		{
			name: "low coherence",
			input: &CrossDungeonStory{
				Title:      "Test",
				LevelSpan:  2,
				MinDepth:   1,
				MaxDepth:   2,
				Fragments:  make([]CrossDungeonFragment, 4),
				Coherence:  0.3,
				Continuity: 0.8,
			},
			wantErr: true,
		},
		{
			name: "low continuity",
			input: &CrossDungeonStory{
				Title:      "Test",
				LevelSpan:  2,
				MinDepth:   1,
				MaxDepth:   2,
				Fragments:  make([]CrossDungeonFragment, 4),
				Coherence:  0.7,
				Continuity: 0.3,
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

func TestCrossDungeonDeterminism(t *testing.T) {
	gen := NewCrossDungeonGenerator()
	seed := int64(11111)
	params := procgen.GenerationParams{
		Difficulty: 0.7,
		Depth:      4,
		GenreID:    "postapoc",
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

	story1 := result1.(*CrossDungeonStory)
	story2 := result2.(*CrossDungeonStory)

	// Check same level span
	if story1.LevelSpan != story2.LevelSpan {
		t.Errorf("Different level spans: %d vs %d", story1.LevelSpan, story2.LevelSpan)
	}

	// Check same fragment count
	if len(story1.Fragments) != len(story2.Fragments) {
		t.Errorf("Different fragment counts: %d vs %d", len(story1.Fragments), len(story2.Fragments))
	}

	// Check same completion bonus
	if story1.CompletionBonus != story2.CompletionBonus {
		t.Errorf("Different completion bonuses: %.2f vs %.2f", story1.CompletionBonus, story2.CompletionBonus)
	}
}

func TestIsFragmentAccessible(t *testing.T) {
	story := &CrossDungeonStory{
		Fragments: []CrossDungeonFragment{
			{Prerequisite: []int{}},     // Fragment 0, no prerequisites
			{Prerequisite: []int{0}},    // Fragment 1, requires 0
			{Prerequisite: []int{0, 1}}, // Fragment 2, requires 0 and 1
			{Prerequisite: []int{1}},    // Fragment 3, requires 1
		},
	}

	tests := []struct {
		name       string
		fragIndex  int
		discovered map[int]bool
		want       bool
	}{
		{
			name:       "no prerequisites",
			fragIndex:  0,
			discovered: map[int]bool{},
			want:       true,
		},
		{
			name:       "prerequisite met",
			fragIndex:  1,
			discovered: map[int]bool{0: true},
			want:       true,
		},
		{
			name:       "prerequisite not met",
			fragIndex:  1,
			discovered: map[int]bool{},
			want:       false,
		},
		{
			name:       "multiple prerequisites all met",
			fragIndex:  2,
			discovered: map[int]bool{0: true, 1: true},
			want:       true,
		},
		{
			name:       "multiple prerequisites partial",
			fragIndex:  2,
			discovered: map[int]bool{0: true},
			want:       false,
		},
		{
			name:       "invalid index",
			fragIndex:  10,
			discovered: map[int]bool{},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := story.IsFragmentAccessible(tt.fragIndex, tt.discovered)
			if got != tt.want {
				t.Errorf("IsFragmentAccessible() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFragmentsForLevel(t *testing.T) {
	story := &CrossDungeonStory{
		Fragments: []CrossDungeonFragment{
			{Level: DungeonLevel{Depth: 1}},
			{Level: DungeonLevel{Depth: 1}},
			{Level: DungeonLevel{Depth: 2}},
			{Level: DungeonLevel{Depth: 3}},
			{Level: DungeonLevel{Depth: 3}},
		},
	}

	tests := []struct {
		name      string
		depth     int
		wantCount int
	}{
		{
			name:      "level 1",
			depth:     1,
			wantCount: 2,
		},
		{
			name:      "level 2",
			depth:     2,
			wantCount: 1,
		},
		{
			name:      "level 3",
			depth:     3,
			wantCount: 2,
		},
		{
			name:      "level 4 (none)",
			depth:     4,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frags := story.GetFragmentsForLevel(tt.depth)
			if len(frags) != tt.wantCount {
				t.Errorf("GetFragmentsForLevel(%d) returned %d fragments, want %d", tt.depth, len(frags), tt.wantCount)
			}
		})
	}
}

func TestGetRequiredLevels(t *testing.T) {
	story := &CrossDungeonStory{
		Fragments: []CrossDungeonFragment{
			{Level: DungeonLevel{Depth: 1, Required: true}},
			{Level: DungeonLevel{Depth: 1, Required: false}},
			{Level: DungeonLevel{Depth: 2, Required: true}},
			{Level: DungeonLevel{Depth: 3, Required: false}},
			{Level: DungeonLevel{Depth: 3, Required: true}},
		},
	}

	levels := story.GetRequiredLevels()

	// Should return depths 1, 2, 3 (all have at least one required fragment)
	if len(levels) != 3 {
		t.Errorf("GetRequiredLevels() returned %d levels, want 3", len(levels))
	}

	// Check all required depths are present
	requiredDepths := map[int]bool{1: false, 2: false, 3: false}
	for _, level := range levels {
		if _, exists := requiredDepths[level]; exists {
			requiredDepths[level] = true
		}
	}

	for depth, found := range requiredDepths {
		if !found {
			t.Errorf("Required depth %d not found in result", depth)
		}
	}
}

func BenchmarkCrossDungeonGenerate(b *testing.B) {
	gen := NewCrossDungeonGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      3,
		GenreID:    "fantasy",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate(int64(i), params)
	}
}

func BenchmarkIsFragmentAccessible(b *testing.B) {
	story := &CrossDungeonStory{
		Fragments: []CrossDungeonFragment{
			{Prerequisite: []int{}},
			{Prerequisite: []int{0}},
			{Prerequisite: []int{0, 1}},
		},
	}
	discovered := map[int]bool{0: true, 1: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = story.IsFragmentAccessible(2, discovered)
	}
}
