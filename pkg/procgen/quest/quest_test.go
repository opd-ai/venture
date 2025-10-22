package quest

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestQuestTypeString(t *testing.T) {
	tests := []struct {
		name     string
		qType    QuestType
		expected string
	}{
		{"kill quest", TypeKill, "kill"},
		{"collect quest", TypeCollect, "collect"},
		{"escort quest", TypeEscort, "escort"},
		{"explore quest", TypeExplore, "explore"},
		{"talk quest", TypeTalk, "talk"},
		{"boss quest", TypeBoss, "boss"},
		{"unknown quest", QuestType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.qType.String(); got != tt.expected {
				t.Errorf("QuestType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestQuestStatusString(t *testing.T) {
	tests := []struct {
		name     string
		status   QuestStatus
		expected string
	}{
		{"not started", StatusNotStarted, "not_started"},
		{"active", StatusActive, "active"},
		{"complete", StatusComplete, "complete"},
		{"turned in", StatusTurnedIn, "turned_in"},
		{"failed", StatusFailed, "failed"},
		{"unknown", QuestStatus(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("QuestStatus.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDifficultyString(t *testing.T) {
	tests := []struct {
		name       string
		difficulty Difficulty
		expected   string
	}{
		{"trivial", DifficultyTrivial, "trivial"},
		{"easy", DifficultyEasy, "easy"},
		{"normal", DifficultyNormal, "normal"},
		{"hard", DifficultyHard, "hard"},
		{"elite", DifficultyElite, "elite"},
		{"legendary", DifficultyLegendary, "legendary"},
		{"unknown", Difficulty(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.difficulty.String(); got != tt.expected {
				t.Errorf("Difficulty.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestObjectiveIsComplete(t *testing.T) {
	tests := []struct {
		name     string
		current  int
		required int
		want     bool
	}{
		{"not complete", 5, 10, false},
		{"exactly complete", 10, 10, true},
		{"over complete", 15, 10, true},
		{"zero progress", 0, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := Objective{
				Current:  tt.current,
				Required: tt.required,
			}
			if got := obj.IsComplete(); got != tt.want {
				t.Errorf("Objective.IsComplete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestObjectiveProgress(t *testing.T) {
	tests := []struct {
		name     string
		current  int
		required int
		want     float64
	}{
		{"half complete", 5, 10, 0.5},
		{"fully complete", 10, 10, 1.0},
		{"over complete capped", 15, 10, 1.0},
		{"no progress", 0, 10, 0.0},
		{"zero required", 5, 0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := Objective{
				Current:  tt.current,
				Required: tt.required,
			}
			if got := obj.Progress(); got != tt.want {
				t.Errorf("Objective.Progress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQuestIsComplete(t *testing.T) {
	tests := []struct {
		name       string
		objectives []Objective
		want       bool
	}{
		{
			name: "all complete",
			objectives: []Objective{
				{Current: 10, Required: 10},
				{Current: 5, Required: 5},
			},
			want: true,
		},
		{
			name: "one incomplete",
			objectives: []Objective{
				{Current: 10, Required: 10},
				{Current: 3, Required: 5},
			},
			want: false,
		},
		{
			name:       "no objectives",
			objectives: []Objective{},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quest := &Quest{
				Objectives: tt.objectives,
			}
			if got := quest.IsComplete(); got != tt.want {
				t.Errorf("Quest.IsComplete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQuestProgress(t *testing.T) {
	tests := []struct {
		name       string
		objectives []Objective
		want       float64
	}{
		{
			name: "half complete overall",
			objectives: []Objective{
				{Current: 5, Required: 10},
				{Current: 5, Required: 10},
			},
			want: 0.5,
		},
		{
			name: "mixed progress",
			objectives: []Objective{
				{Current: 10, Required: 10}, // 1.0
				{Current: 0, Required: 10},  // 0.0
			},
			want: 0.5,
		},
		{
			name:       "no objectives",
			objectives: []Objective{},
			want:       1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quest := &Quest{
				Objectives: tt.objectives,
			}
			if got := quest.Progress(); got != tt.want {
				t.Errorf("Quest.Progress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQuestGetRewardValue(t *testing.T) {
	quest := &Quest{
		Reward: Reward{
			XP:          100,
			Gold:        50,
			Items:       []string{"item1", "item2"},
			SkillPoints: 2,
		},
	}

	value := quest.GetRewardValue()
	expected := 100 + (50 * 2) + (2 * 100) + (2 * 500)
	
	if value != expected {
		t.Errorf("Quest.GetRewardValue() = %v, want %v", value, expected)
	}
}

func TestQuestGeneratorGenerate(t *testing.T) {
	generator := NewQuestGenerator()

	tests := []struct {
		name    string
		seed    int64
		params  procgen.GenerationParams
		wantErr bool
	}{
		{
			name: "valid fantasy generation",
			seed: 12345,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    "fantasy",
				Custom:     map[string]interface{}{"count": 5},
			},
			wantErr: false,
		},
		{
			name: "valid scifi generation",
			seed: 67890,
			params: procgen.GenerationParams{
				Difficulty: 0.7,
				Depth:      10,
				GenreID:    "scifi",
				Custom:     map[string]interface{}{"count": 3},
			},
			wantErr: false,
		},
		{
			name: "negative depth",
			seed: 11111,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      -1,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
		{
			name: "invalid difficulty low",
			seed: 22222,
			params: procgen.GenerationParams{
				Difficulty: -0.1,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
		{
			name: "invalid difficulty high",
			seed: 33333,
			params: procgen.GenerationParams{
				Difficulty: 1.5,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
		{
			name: "default count",
			seed: 44444,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      3,
				GenreID:    "fantasy",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generator.Generate(tt.seed, tt.params)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("Generate() expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Generate() unexpected error: %v", err)
				return
			}

			quests, ok := result.([]*Quest)
			if !ok {
				t.Errorf("Generate() returned wrong type: %T", result)
				return
			}

			expectedCount := 5 // default
			if c, ok := tt.params.Custom["count"].(int); ok {
				expectedCount = c
			}

			if len(quests) != expectedCount {
				t.Errorf("Generate() returned %d quests, want %d", len(quests), expectedCount)
			}

			// Verify all quests are valid
			for i, quest := range quests {
				if quest == nil {
					t.Errorf("Quest %d is nil", i)
					continue
				}
				if quest.Name == "" {
					t.Errorf("Quest %d has empty name", i)
				}
				if quest.Description == "" {
					t.Errorf("Quest %d has empty description", i)
				}
				if len(quest.Objectives) == 0 {
					t.Errorf("Quest %d has no objectives", i)
				}
				if quest.Reward.XP <= 0 {
					t.Errorf("Quest %d has no XP reward", i)
				}
			}
		})
	}
}

func TestQuestGeneratorDeterminism(t *testing.T) {
	generator := NewQuestGenerator()
	seed := int64(99999)
	params := procgen.GenerationParams{
		Difficulty: 0.6,
		Depth:      7,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 5},
	}

	// Generate twice with same seed
	result1, err1 := generator.Generate(seed, params)
	if err1 != nil {
		t.Fatalf("First generation failed: %v", err1)
	}

	result2, err2 := generator.Generate(seed, params)
	if err2 != nil {
		t.Fatalf("Second generation failed: %v", err2)
	}

	quests1 := result1.([]*Quest)
	quests2 := result2.([]*Quest)

	if len(quests1) != len(quests2) {
		t.Fatalf("Different number of quests: %d vs %d", len(quests1), len(quests2))
	}

	// Verify quests are identical
	for i := range quests1 {
		q1 := quests1[i]
		q2 := quests2[i]

		if q1.Name != q2.Name {
			t.Errorf("Quest %d name differs: %s vs %s", i, q1.Name, q2.Name)
		}
		if q1.Type != q2.Type {
			t.Errorf("Quest %d type differs: %s vs %s", i, q1.Type, q2.Type)
		}
		if q1.Difficulty != q2.Difficulty {
			t.Errorf("Quest %d difficulty differs: %s vs %s", i, q1.Difficulty, q2.Difficulty)
		}
		if q1.Reward.XP != q2.Reward.XP {
			t.Errorf("Quest %d XP differs: %d vs %d", i, q1.Reward.XP, q2.Reward.XP)
		}
		if q1.Reward.Gold != q2.Reward.Gold {
			t.Errorf("Quest %d gold differs: %d vs %d", i, q1.Reward.Gold, q2.Reward.Gold)
		}
	}
}

func TestQuestGeneratorValidate(t *testing.T) {
	generator := NewQuestGenerator()

	tests := []struct {
		name    string
		result  interface{}
		wantErr bool
	}{
		{
			name: "valid quests",
			result: []*Quest{
				{
					Name:        "Test Quest",
					Description: "Test description",
					Objectives: []Objective{
						{Description: "Objective 1", Required: 5},
					},
					Reward:        Reward{XP: 100},
					RequiredLevel: 1,
				},
			},
			wantErr: false,
		},
		{
			name:    "wrong type",
			result:  "not a quest slice",
			wantErr: true,
		},
		{
			name:    "empty slice",
			result:  []*Quest{},
			wantErr: true,
		},
		{
			name: "nil quest",
			result: []*Quest{
				nil,
			},
			wantErr: true,
		},
		{
			name: "empty name",
			result: []*Quest{
				{
					Name:        "",
					Description: "Test",
					Objectives:  []Objective{{Description: "Test", Required: 1}},
					Reward:      Reward{XP: 100},
				},
			},
			wantErr: true,
		},
		{
			name: "empty description",
			result: []*Quest{
				{
					Name:        "Test",
					Description: "",
					Objectives:  []Objective{{Description: "Test", Required: 1}},
					Reward:      Reward{XP: 100},
				},
			},
			wantErr: true,
		},
		{
			name: "no objectives",
			result: []*Quest{
				{
					Name:        "Test",
					Description: "Test",
					Objectives:  []Objective{},
					Reward:      Reward{XP: 100},
				},
			},
			wantErr: true,
		},
		{
			name: "empty objective description",
			result: []*Quest{
				{
					Name:        "Test",
					Description: "Test",
					Objectives:  []Objective{{Description: "", Required: 1}},
					Reward:      Reward{XP: 100},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid required amount",
			result: []*Quest{
				{
					Name:        "Test",
					Description: "Test",
					Objectives:  []Objective{{Description: "Test", Required: 0}},
					Reward:      Reward{XP: 100},
				},
			},
			wantErr: true,
		},
		{
			name: "no XP reward",
			result: []*Quest{
				{
					Name:        "Test",
					Description: "Test",
					Objectives:  []Objective{{Description: "Test", Required: 1}},
					Reward:      Reward{XP: 0},
				},
			},
			wantErr: true,
		},
		{
			name: "negative required level",
			result: []*Quest{
				{
					Name:          "Test",
					Description:   "Test",
					Objectives:    []Objective{{Description: "Test", Required: 1}},
					Reward:        Reward{XP: 100},
					RequiredLevel: -1,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := generator.Validate(tt.result)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQuestGeneratorScaling(t *testing.T) {
	generator := NewQuestGenerator()
	seed := int64(123456)

	tests := []struct {
		name       string
		depth      int
		difficulty float64
	}{
		{"low depth low difficulty", 1, 0.0},
		{"medium depth medium difficulty", 5, 0.5},
		{"high depth high difficulty", 10, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: tt.difficulty,
				Depth:      tt.depth,
				GenreID:    "fantasy",
				Custom:     map[string]interface{}{"count": 3},
			}

			result, err := generator.Generate(seed, params)
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			quests := result.([]*Quest)

			for _, quest := range quests {
				// Verify scaling affects rewards
				if quest.Reward.XP <= 0 {
					t.Errorf("Quest has no XP reward")
				}
				
				// Higher depth/difficulty should generally give better rewards
				// (though randomness can vary individual quests)
				t.Logf("Depth %d, Diff %.1f: Quest '%s' rewards %d XP, %d gold",
					tt.depth, tt.difficulty, quest.Name, quest.Reward.XP, quest.Reward.Gold)
			}
		})
	}
}

func BenchmarkQuestGeneration(b *testing.B) {
	generator := NewQuestGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 10},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generator.Generate(int64(i), params)
		if err != nil {
			b.Fatalf("Generate failed: %v", err)
		}
	}
}

func BenchmarkQuestValidation(b *testing.B) {
	generator := NewQuestGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 10},
	}

	result, _ := generator.Generate(12345, params)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generator.Validate(result)
	}
}
