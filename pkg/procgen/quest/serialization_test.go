package quest

import (
	"encoding/json"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestQuestSerializeDeserialize(t *testing.T) {
	original := &Quest{
		ID:                   "quest_5_0",
		Name:                 "Slay the Goblins",
		Type:                 TypeKill,
		Difficulty:           DifficultyHard,
		Description:          "Goblins have been terrorizing the area.",
		RequiredLevel:        6,
		Status:               StatusActive,
		Seed:                 12345,
		Tags:                 []string{"combat", "kill"},
		GiverNPC:             "Elder",
		Location:             "Dark Forest",
		MoralChoiceID:        "choice_1",
		FactionA:             "Alliance",
		FactionB:             "Horde",
		HasMoralConsequences: true,
		Objectives: []Objective{
			{Description: "Defeat 10 Goblins", Target: "Goblin", Required: 10, Current: 5},
			{Description: "Defeat 3 Orc", Target: "Orc", Required: 3, Current: 3},
		},
		Reward: Reward{
			XP:          200,
			Gold:        50,
			Items:       []string{"item_hard_0", "item_hard_1"},
			SkillPoints: 1,
		},
	}

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error: %v", err)
	}

	restored := &Quest{}
	if err := restored.Deserialize(data); err != nil {
		t.Fatalf("Deserialize() error: %v", err)
	}

	// Verify all fields round-trip correctly
	if restored.ID != original.ID {
		t.Errorf("ID = %q, want %q", restored.ID, original.ID)
	}
	if restored.Name != original.Name {
		t.Errorf("Name = %q, want %q", restored.Name, original.Name)
	}
	if restored.Type != original.Type {
		t.Errorf("Type = %v, want %v", restored.Type, original.Type)
	}
	if restored.Difficulty != original.Difficulty {
		t.Errorf("Difficulty = %v, want %v", restored.Difficulty, original.Difficulty)
	}
	if restored.Status != original.Status {
		t.Errorf("Status = %v, want %v", restored.Status, original.Status)
	}
	if restored.Seed != original.Seed {
		t.Errorf("Seed = %v, want %v", restored.Seed, original.Seed)
	}
	if restored.HasMoralConsequences != original.HasMoralConsequences {
		t.Errorf("HasMoralConsequences = %v, want %v", restored.HasMoralConsequences, original.HasMoralConsequences)
	}
	if len(restored.Objectives) != len(original.Objectives) {
		t.Fatalf("Objectives len = %d, want %d", len(restored.Objectives), len(original.Objectives))
	}
	for i, obj := range restored.Objectives {
		if obj.Current != original.Objectives[i].Current {
			t.Errorf("Objective[%d].Current = %d, want %d", i, obj.Current, original.Objectives[i].Current)
		}
		if obj.Required != original.Objectives[i].Required {
			t.Errorf("Objective[%d].Required = %d, want %d", i, obj.Required, original.Objectives[i].Required)
		}
	}
	if restored.Reward.XP != original.Reward.XP {
		t.Errorf("Reward.XP = %d, want %d", restored.Reward.XP, original.Reward.XP)
	}
	if restored.Reward.Gold != original.Reward.Gold {
		t.Errorf("Reward.Gold = %d, want %d", restored.Reward.Gold, original.Reward.Gold)
	}
	if len(restored.Reward.Items) != len(original.Reward.Items) {
		t.Errorf("Reward.Items len = %d, want %d", len(restored.Reward.Items), len(original.Reward.Items))
	}
}

func TestQuestSerializeEmpty(t *testing.T) {
	original := &Quest{}
	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() empty quest error: %v", err)
	}

	restored := &Quest{}
	if err := restored.Deserialize(data); err != nil {
		t.Fatalf("Deserialize() empty quest error: %v", err)
	}
}

func TestQuestDeserializeInvalid(t *testing.T) {
	q := &Quest{}
	if err := q.Deserialize([]byte("invalid json")); err == nil {
		t.Error("Deserialize() expected error for invalid JSON")
	}
}

func TestObjectiveSerializeDeserialize(t *testing.T) {
	original := &Objective{
		Description: "Collect 5 herbs",
		Target:      "Moonflower",
		Required:    5,
		Current:     3,
	}

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Objective.Serialize() error: %v", err)
	}

	restored := &Objective{}
	if err := restored.Deserialize(data); err != nil {
		t.Fatalf("Objective.Deserialize() error: %v", err)
	}

	if restored.Description != original.Description {
		t.Errorf("Description = %q, want %q", restored.Description, original.Description)
	}
	if restored.Current != original.Current {
		t.Errorf("Current = %d, want %d", restored.Current, original.Current)
	}
}

func TestRewardSerializeDeserialize(t *testing.T) {
	original := &Reward{
		XP:          500,
		Gold:        200,
		Items:       []string{"sword", "shield"},
		SkillPoints: 3,
	}

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Reward.Serialize() error: %v", err)
	}

	restored := &Reward{}
	if err := restored.Deserialize(data); err != nil {
		t.Fatalf("Reward.Deserialize() error: %v", err)
	}

	if restored.XP != original.XP {
		t.Errorf("XP = %d, want %d", restored.XP, original.XP)
	}
	if restored.Gold != original.Gold {
		t.Errorf("Gold = %d, want %d", restored.Gold, original.Gold)
	}
	if restored.SkillPoints != original.SkillPoints {
		t.Errorf("SkillPoints = %d, want %d", restored.SkillPoints, original.SkillPoints)
	}
}

func TestSerializeGeneratedQuests(t *testing.T) {
	generator := NewQuestGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 3},
	}

	result, err := generator.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	quests := result.([]*Quest)
	for i, q := range quests {
		data, err := q.Serialize()
		if err != nil {
			t.Fatalf("Quest[%d].Serialize() error: %v", i, err)
		}

		restored := &Quest{}
		if err := restored.Deserialize(data); err != nil {
			t.Fatalf("Quest[%d].Deserialize() error: %v", i, err)
		}

		if restored.Name != q.Name {
			t.Errorf("Quest[%d] Name = %q, want %q", i, restored.Name, q.Name)
		}
		if restored.IsComplete() != q.IsComplete() {
			t.Errorf("Quest[%d] IsComplete = %v, want %v", i, restored.IsComplete(), q.IsComplete())
		}
		if restored.Progress() != q.Progress() {
			t.Errorf("Quest[%d] Progress = %v, want %v", i, restored.Progress(), q.Progress())
		}
		if restored.GetRewardValue() != q.GetRewardValue() {
			t.Errorf("Quest[%d] RewardValue = %v, want %v", i, restored.GetRewardValue(), q.GetRewardValue())
		}
	}
}

func TestSerializeMultipleQuestsAsSlice(t *testing.T) {
	generator := NewQuestGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 5},
	}

	result, err := generator.Generate(99999, params)
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	quests := result.([]*Quest)

	// Serialize as a JSON array
	data, err := json.Marshal(quests)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	var restored []*Quest
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if len(restored) != len(quests) {
		t.Fatalf("Restored len = %d, want %d", len(restored), len(quests))
	}

	for i := range quests {
		if restored[i].ID != quests[i].ID {
			t.Errorf("Quest[%d] ID = %q, want %q", i, restored[i].ID, quests[i].ID)
		}
	}
}

func TestHorrorQuestGeneration(t *testing.T) {
	generator := NewQuestGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "horror",
		Custom:     map[string]interface{}{"count": 5},
	}

	result, err := generator.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate() horror error: %v", err)
	}

	quests := result.([]*Quest)
	if len(quests) != 5 {
		t.Fatalf("Generated %d quests, want 5", len(quests))
	}

	if err := generator.Validate(result); err != nil {
		t.Errorf("Validate() horror quests error: %v", err)
	}

	for i, q := range quests {
		if q.Name == "" {
			t.Errorf("Horror quest %d has empty name", i)
		}
		if q.Description == "" {
			t.Errorf("Horror quest %d has empty description", i)
		}
		if len(q.Objectives) == 0 {
			t.Errorf("Horror quest %d has no objectives", i)
		}
	}
}

func TestCyberpunkQuestGeneration(t *testing.T) {
	generator := NewQuestGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.6,
		Depth:      7,
		GenreID:    "cyberpunk",
		Custom:     map[string]interface{}{"count": 5},
	}

	result, err := generator.Generate(67890, params)
	if err != nil {
		t.Fatalf("Generate() cyberpunk error: %v", err)
	}

	quests := result.([]*Quest)
	if len(quests) != 5 {
		t.Fatalf("Generated %d quests, want 5", len(quests))
	}

	if err := generator.Validate(result); err != nil {
		t.Errorf("Validate() cyberpunk quests error: %v", err)
	}

	for i, q := range quests {
		if q.Name == "" {
			t.Errorf("Cyberpunk quest %d has empty name", i)
		}
		if q.Description == "" {
			t.Errorf("Cyberpunk quest %d has empty description", i)
		}
		if len(q.Objectives) == 0 {
			t.Errorf("Cyberpunk quest %d has no objectives", i)
		}
	}
}

func TestHorrorQuestDeterminism(t *testing.T) {
	generator := NewQuestGenerator()
	seed := int64(55555)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "horror",
		Custom:     map[string]interface{}{"count": 3},
	}

	result1, _ := generator.Generate(seed, params)
	result2, _ := generator.Generate(seed, params)

	quests1 := result1.([]*Quest)
	quests2 := result2.([]*Quest)

	for i := range quests1 {
		if quests1[i].Name != quests2[i].Name {
			t.Errorf("Horror quest %d name differs: %q vs %q", i, quests1[i].Name, quests2[i].Name)
		}
		if quests1[i].Reward.XP != quests2[i].Reward.XP {
			t.Errorf("Horror quest %d XP differs: %d vs %d", i, quests1[i].Reward.XP, quests2[i].Reward.XP)
		}
	}
}

func TestCyberpunkQuestDeterminism(t *testing.T) {
	generator := NewQuestGenerator()
	seed := int64(77777)
	params := procgen.GenerationParams{
		Difficulty: 0.7,
		Depth:      8,
		GenreID:    "cyberpunk",
		Custom:     map[string]interface{}{"count": 3},
	}

	result1, _ := generator.Generate(seed, params)
	result2, _ := generator.Generate(seed, params)

	quests1 := result1.([]*Quest)
	quests2 := result2.([]*Quest)

	for i := range quests1 {
		if quests1[i].Name != quests2[i].Name {
			t.Errorf("Cyberpunk quest %d name differs: %q vs %q", i, quests1[i].Name, quests2[i].Name)
		}
		if quests1[i].Reward.XP != quests2[i].Reward.XP {
			t.Errorf("Cyberpunk quest %d XP differs: %d vs %d", i, quests1[i].Reward.XP, quests2[i].Reward.XP)
		}
	}
}

func TestPackageLevelFunctions(t *testing.T) {
	t.Run("ObjectiveIsComplete", func(t *testing.T) {
		obj := &Objective{Current: 10, Required: 10}
		if !ObjectiveIsComplete(obj) {
			t.Error("ObjectiveIsComplete() = false, want true")
		}

		obj2 := &Objective{Current: 5, Required: 10}
		if ObjectiveIsComplete(obj2) {
			t.Error("ObjectiveIsComplete() = true, want false")
		}
	})

	t.Run("ObjectiveProgress", func(t *testing.T) {
		obj := &Objective{Current: 5, Required: 10}
		if got := ObjectiveProgress(obj); got != 0.5 {
			t.Errorf("ObjectiveProgress() = %v, want 0.5", got)
		}
	})

	t.Run("QuestIsComplete", func(t *testing.T) {
		q := &Quest{
			Objectives: []Objective{
				{Current: 10, Required: 10},
				{Current: 5, Required: 5},
			},
		}
		if !QuestIsComplete(q) {
			t.Error("QuestIsComplete() = false, want true")
		}

		q2 := &Quest{
			Objectives: []Objective{
				{Current: 10, Required: 10},
				{Current: 3, Required: 5},
			},
		}
		if QuestIsComplete(q2) {
			t.Error("QuestIsComplete() = true, want false")
		}
	})

	t.Run("QuestProgress", func(t *testing.T) {
		q := &Quest{
			Objectives: []Objective{
				{Current: 5, Required: 10},
				{Current: 5, Required: 10},
			},
		}
		if got := QuestProgress(q); got != 0.5 {
			t.Errorf("QuestProgress() = %v, want 0.5", got)
		}
	})

	t.Run("QuestRewardValue", func(t *testing.T) {
		q := &Quest{
			Reward: Reward{XP: 100, Gold: 50, Items: []string{"a", "b"}, SkillPoints: 2},
		}
		expected := 100 + (50 * 2) + (2 * 100) + (2 * 500)
		if got := QuestRewardValue(q); got != expected {
			t.Errorf("QuestRewardValue() = %d, want %d", got, expected)
		}
	})
}

func TestValidateNegativeGoldReward(t *testing.T) {
	generator := NewQuestGenerator()
	result := []*Quest{
		{
			Name:        "Test",
			Description: "Test",
			Objectives:  []Objective{{Description: "Test", Required: 1}},
			Reward:      Reward{XP: 100, Gold: -10},
		},
	}

	err := generator.Validate(result)
	if err == nil {
		t.Error("Validate() expected error for negative gold")
	}
}

func TestHorrorTemplates(t *testing.T) {
	templates := []struct {
		name string
		fn   func() []QuestTemplate
	}{
		{"kill", GetHorrorKillTemplates},
		{"collect", GetHorrorCollectTemplates},
		{"boss", GetHorrorBossTemplates},
		{"explore", GetHorrorExploreTemplates},
	}

	for _, tt := range templates {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := tt.fn()
			if len(tmpl) == 0 {
				t.Fatal("expected at least one template")
			}
			for _, tpl := range tmpl {
				if len(tpl.NamePrefixes) == 0 {
					t.Error("template has no name prefixes")
				}
				if len(tpl.NameSuffixes) == 0 {
					t.Error("template has no name suffixes")
				}
				if len(tpl.DescTemplates) == 0 {
					t.Error("template has no description templates")
				}
				if len(tpl.TargetTypes) == 0 {
					t.Error("template has no target types")
				}
			}
		})
	}
}

func TestCyberpunkTemplates(t *testing.T) {
	templates := []struct {
		name string
		fn   func() []QuestTemplate
	}{
		{"kill", GetCyberpunkKillTemplates},
		{"collect", GetCyberpunkCollectTemplates},
		{"boss", GetCyberpunkBossTemplates},
		{"explore", GetCyberpunkExploreTemplates},
	}

	for _, tt := range templates {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := tt.fn()
			if len(tmpl) == 0 {
				t.Fatal("expected at least one template")
			}
			for _, tpl := range tmpl {
				if len(tpl.NamePrefixes) == 0 {
					t.Error("template has no name prefixes")
				}
				if len(tpl.NameSuffixes) == 0 {
					t.Error("template has no name suffixes")
				}
				if len(tpl.DescTemplates) == 0 {
					t.Error("template has no description templates")
				}
				if len(tpl.TargetTypes) == 0 {
					t.Error("template has no target types")
				}
			}
		})
	}
}

func TestSciFiExploreTemplates(t *testing.T) {
	templates := GetSciFiExploreTemplates()
	if len(templates) == 0 {
		t.Fatal("expected at least one sci-fi explore template")
	}
	if templates[0].BaseType != TypeExplore {
		t.Errorf("BaseType = %v, want TypeExplore", templates[0].BaseType)
	}
}

func TestAllGenresProduceValidQuests(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk"}
	generator := NewQuestGenerator()

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    genre,
				Custom:     map[string]interface{}{"count": 10},
			}

			result, err := generator.Generate(42, params)
			if err != nil {
				t.Fatalf("Generate() %s error: %v", genre, err)
			}

			if err := generator.Validate(result); err != nil {
				t.Errorf("Validate() %s error: %v", genre, err)
			}

			quests := result.([]*Quest)
			if len(quests) != 10 {
				t.Errorf("Generated %d quests, want 10", len(quests))
			}
		})
	}
}
