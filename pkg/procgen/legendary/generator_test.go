package legendary

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/world/raids"
)

func TestGenerator_Generate(t *testing.T) {
	tests := []struct {
		name    string
		seed    int64
		params  procgen.GenerationParams
		wantErr bool
	}{
		{
			name: "valid fantasy quest",
			seed: 12345,
			params: procgen.GenerationParams{
				Difficulty: 0.9,
				Depth:      50,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"player_level":    50,
					"servers_visited": 3,
				},
			},
			wantErr: false,
		},
		{
			name: "valid sci-fi quest",
			seed: 67890,
			params: procgen.GenerationParams{
				Difficulty: 0.8,
				Depth:      45,
				GenreID:    "scifi",
				Custom: map[string]interface{}{
					"player_level":    45,
					"servers_visited": 4,
				},
			},
			wantErr: false,
		},
		{
			name: "maximum servers",
			seed: 11111,
			params: procgen.GenerationParams{
				Difficulty: 0.95,
				Depth:      50,
				GenreID:    "horror",
				Custom: map[string]interface{}{
					"player_level":    50,
					"servers_visited": 5,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid difficulty low",
			seed: 22222,
			params: procgen.GenerationParams{
				Difficulty: -0.1,
				Depth:      50,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
		{
			name: "invalid difficulty high",
			seed: 33333,
			params: procgen.GenerationParams{
				Difficulty: 1.1,
				Depth:      50,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
		{
			name: "invalid depth",
			seed: 44444,
			params: procgen.GenerationParams{
				Difficulty: 0.9,
				Depth:      0,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
		{
			name: "invalid genre",
			seed: 55555,
			params: procgen.GenerationParams{
				Difficulty: 0.9,
				Depth:      50,
				GenreID:    "invalid",
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

			if err != nil {
				return
			}

			quest, ok := result.(*LegendaryQuest)
			if !ok {
				t.Errorf("Generate() returned %T, want *LegendaryQuest", result)
				return
			}

			// Verify basic structure
			if quest.ID == "" {
				t.Error("quest ID is empty")
			}
			if quest.Name == "" {
				t.Error("quest name is empty")
			}
			if len(quest.Phases) < 5 || len(quest.Phases) > 10 {
				t.Errorf("quest has %d phases, want 5-10", len(quest.Phases))
			}
			if len(quest.Rewards) < 1 || len(quest.Rewards) > 3 {
				t.Errorf("quest has %d rewards, want 1-3", len(quest.Rewards))
			}

			// Verify final phase
			lastPhase := quest.Phases[len(quest.Phases)-1]
			if lastPhase.Type != PhaseFinal {
				t.Errorf("last phase type = %s, want Final", lastPhase.Type)
			}

			// Verify servers required
			if quest.ServersRequired < 3 || quest.ServersRequired > 5 {
				t.Errorf("servers required = %d, want 3-5", quest.ServersRequired)
			}

			// Verify estimated hours
			if quest.EstimatedHours < 10 || quest.EstimatedHours > 20 {
				t.Errorf("estimated hours = %d, want 10-20", quest.EstimatedHours)
			}
		})
	}
}

func TestGenerator_Validate(t *testing.T) {
	tests := []struct {
		name    string
		quest   interface{}
		wantErr bool
	}{
		{
			name: "valid quest",
			quest: &LegendaryQuest{
				ID:              "test_quest",
				Name:            "Test Quest",
				MinLevel:        50,
				EstimatedHours:  15,
				ServersRequired: 3,
				Phases: []*QuestPhase{
					{Type: PhaseExploration},
					{Type: PhaseCombat},
					{Type: PhaseCrafting},
					{Type: PhaseCollection},
					{Type: PhaseRaid},
					{Type: PhaseFinal},
				},
				Rewards: []*LegendaryReward{
					{Type: RewardItem, IsUnique: true},
				},
			},
			wantErr: false,
		},
		{
			name:    "invalid type",
			quest:   "not a quest",
			wantErr: true,
		},
		{
			name: "empty ID",
			quest: &LegendaryQuest{
				ID:              "",
				Name:            "Test",
				MinLevel:        50,
				EstimatedHours:  15,
				ServersRequired: 3,
				Phases: []*QuestPhase{
					{Type: PhaseFinal},
				},
				Rewards: []*LegendaryReward{
					{Type: RewardItem, IsUnique: true},
				},
			},
			wantErr: true,
		},
		{
			name: "too few phases",
			quest: &LegendaryQuest{
				ID:              "test",
				Name:            "Test",
				MinLevel:        50,
				EstimatedHours:  15,
				ServersRequired: 3,
				Phases: []*QuestPhase{
					{Type: PhaseFinal},
				},
				Rewards: []*LegendaryReward{
					{Type: RewardItem, IsUnique: true},
				},
			},
			wantErr: true,
		},
		{
			name: "no final phase",
			quest: &LegendaryQuest{
				ID:              "test",
				Name:            "Test",
				MinLevel:        50,
				EstimatedHours:  15,
				ServersRequired: 3,
				Phases: []*QuestPhase{
					{Type: PhaseExploration},
					{Type: PhaseCombat},
					{Type: PhaseCrafting},
					{Type: PhaseCollection},
					{Type: PhaseRaid},
				},
				Rewards: []*LegendaryReward{
					{Type: RewardItem, IsUnique: true},
				},
			},
			wantErr: true,
		},
		{
			name: "no unique rewards",
			quest: &LegendaryQuest{
				ID:              "test",
				Name:            "Test",
				MinLevel:        50,
				EstimatedHours:  15,
				ServersRequired: 3,
				Phases: []*QuestPhase{
					{Type: PhaseExploration},
					{Type: PhaseCombat},
					{Type: PhaseCrafting},
					{Type: PhaseCollection},
					{Type: PhaseRaid},
					{Type: PhaseFinal},
				},
				Rewards: []*LegendaryReward{
					{Type: RewardItem, IsUnique: false},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator()
			err := gen.Validate(tt.quest)

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerator_Determinism(t *testing.T) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.9,
		Depth:      50,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"player_level":    50,
			"servers_visited": 3,
		},
	}

	seed := int64(12345)

	// Generate twice with same seed
	result1, err1 := gen.Generate(seed, params)
	if err1 != nil {
		t.Fatalf("first generation failed: %v", err1)
	}

	result2, err2 := gen.Generate(seed, params)
	if err2 != nil {
		t.Fatalf("second generation failed: %v", err2)
	}

	quest1 := result1.(*LegendaryQuest)
	quest2 := result2.(*LegendaryQuest)

	// Verify identical generation
	if quest1.Name != quest2.Name {
		t.Errorf("quest names differ: %s vs %s", quest1.Name, quest2.Name)
	}
	if len(quest1.Phases) != len(quest2.Phases) {
		t.Errorf("phase counts differ: %d vs %d", len(quest1.Phases), len(quest2.Phases))
	}
	if len(quest1.Rewards) != len(quest2.Rewards) {
		t.Errorf("reward counts differ: %d vs %d", len(quest1.Rewards), len(quest2.Rewards))
	}

	// Verify phases match
	for i := range quest1.Phases {
		if quest1.Phases[i].Type != quest2.Phases[i].Type {
			t.Errorf("phase %d types differ: %s vs %s", i, quest1.Phases[i].Type, quest2.Phases[i].Type)
		}
		if quest1.Phases[i].Name != quest2.Phases[i].Name {
			t.Errorf("phase %d names differ: %s vs %s", i, quest1.Phases[i].Name, quest2.Phases[i].Name)
		}
	}

	// Verify rewards match
	for i := range quest1.Rewards {
		if quest1.Rewards[i].Type != quest2.Rewards[i].Type {
			t.Errorf("reward %d types differ: %s vs %s", i, quest1.Rewards[i].Type, quest2.Rewards[i].Type)
		}
		if quest1.Rewards[i].Name != quest2.Rewards[i].Name {
			t.Errorf("reward %d names differ: %s vs %s", i, quest1.Rewards[i].Name, quest2.Rewards[i].Name)
		}
	}
}

func TestPhaseType_String(t *testing.T) {
	tests := []struct {
		phaseType PhaseType
		want      string
	}{
		{PhaseExploration, "Exploration"},
		{PhaseCombat, "Combat"},
		{PhaseCrafting, "Crafting"},
		{PhaseCollection, "Collection"},
		{PhaseRaid, "Raid"},
		{PhaseFinal, "Final"},
		{PhaseType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.phaseType.String(); got != tt.want {
				t.Errorf("PhaseType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRewardType_String(t *testing.T) {
	tests := []struct {
		rewardType RewardType
		want       string
	}{
		{RewardItem, "Item"},
		{RewardTitle, "Title"},
		{RewardMount, "Mount"},
		{RewardCompanion, "Companion"},
		{RewardAchievement, "Achievement"},
		{RewardAccountBonus, "AccountBonus"},
		{RewardType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.rewardType.String(); got != tt.want {
				t.Errorf("RewardType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProgressTracker(t *testing.T) {
	tracker := NewProgressTracker()
	questID := "test_quest"
	playerID := "player1"

	// Test initial state
	progress := tracker.GetProgress(questID, playerID)
	if progress != nil {
		t.Error("expected nil progress for new quest/player")
	}

	// Test update phase
	tracker.UpdatePhase(questID, playerID, 0, 50.0)
	progress = tracker.GetProgress(questID, playerID)
	if progress == nil {
		t.Fatal("progress should not be nil after update")
	}
	if progress.CurrentPhase != 0 {
		t.Errorf("current phase = %d, want 0", progress.CurrentPhase)
	}
	if progress.PhaseProgress != 50.0 {
		t.Errorf("phase progress = %f, want 50.0", progress.PhaseProgress)
	}

	// Test add server visited
	tracker.AddServerVisited(questID, playerID, "server1")
	tracker.AddServerVisited(questID, playerID, "server2")
	tracker.AddServerVisited(questID, playerID, "server1") // Duplicate
	progress = tracker.GetProgress(questID, playerID)
	if len(progress.ServersVisited) != 2 {
		t.Errorf("servers visited count = %d, want 2", len(progress.ServersVisited))
	}

	// Test add raid completed
	tracker.AddRaidCompleted(questID, playerID, "raid1")
	tracker.AddRaidCompleted(questID, playerID, "raid2")
	tracker.AddRaidCompleted(questID, playerID, "raid1") // Duplicate
	progress = tracker.GetProgress(questID, playerID)
	if len(progress.RaidsCompleted) != 2 {
		t.Errorf("raids completed count = %d, want 2", len(progress.RaidsCompleted))
	}

	// Test add materials
	tracker.AddMaterial(questID, playerID, "material1", 5)
	tracker.AddMaterial(questID, playerID, "material1", 3)
	tracker.AddMaterial(questID, playerID, "material2", 10)
	progress = tracker.GetProgress(questID, playerID)
	if progress.MaterialsGathered["material1"] != 8 {
		t.Errorf("material1 count = %d, want 8", progress.MaterialsGathered["material1"])
	}
	if progress.MaterialsGathered["material2"] != 10 {
		t.Errorf("material2 count = %d, want 10", progress.MaterialsGathered["material2"])
	}

	// Test complete quest
	tracker.CompleteQuest(questID, playerID)
	progress = tracker.GetProgress(questID, playerID)
	if !progress.IsCompleted {
		t.Error("quest should be marked completed")
	}
	if progress.CompletedAt == nil {
		t.Error("completed at should not be nil")
	}
}

func TestPhaseGeneration(t *testing.T) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.9,
		Depth:      50,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"player_level":    50,
			"servers_visited": 3,
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	quest := result.(*LegendaryQuest)

	// Count phase types
	phaseTypeCounts := make(map[PhaseType]int)
	for _, phase := range quest.Phases {
		phaseTypeCounts[phase.Type]++

		// Verify phase-specific fields
		switch phase.Type {
		case PhaseExploration:
			if phase.ServerID == "" {
				t.Error("exploration phase missing server ID")
			}
		case PhaseCombat:
			if phase.BossName == "" {
				t.Error("combat phase missing boss name")
			}
		case PhaseCrafting:
			if phase.ItemName == "" {
				t.Error("crafting phase missing item name")
			}
			if phase.StationTier != 4 {
				t.Errorf("crafting phase station tier = %d, want 4", phase.StationTier)
			}
		case PhaseCollection:
			if len(phase.MaterialIDs) == 0 {
				t.Error("collection phase has no materials")
			}
			if len(phase.MaterialIDs) != len(phase.Quantities) {
				t.Error("material IDs and quantities length mismatch")
			}
		case PhaseRaid:
			if phase.RaidID == "" {
				t.Error("raid phase missing raid ID")
			}
			if phase.RaidTier < raids.TierNormal || phase.RaidTier > raids.TierNightmare {
				t.Errorf("invalid raid tier: %d", phase.RaidTier)
			}
		case PhaseFinal:
			if phase.XPReward < 10000 {
				t.Errorf("final phase XP reward = %d, want >= 10000", phase.XPReward)
			}
		}
	}

	// Verify final phase exists and is last
	if phaseTypeCounts[PhaseFinal] != 1 {
		t.Errorf("final phase count = %d, want 1", phaseTypeCounts[PhaseFinal])
	}
}

func TestRewardGeneration(t *testing.T) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.9,
		Depth:      50,
		GenreID:    "scifi",
		Custom: map[string]interface{}{
			"player_level":    50,
			"servers_visited": 4,
		},
	}

	result, err := gen.Generate(67890, params)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	quest := result.(*LegendaryQuest)

	// Verify all rewards are unique
	for _, reward := range quest.Rewards {
		if !reward.IsUnique {
			t.Error("legendary reward should be unique")
		}
		if reward.Name == "" {
			t.Error("reward missing name")
		}
		if reward.Description == "" {
			t.Error("reward missing description")
		}

		// Verify type-specific fields
		switch reward.Type {
		case RewardItem:
			if reward.ItemID == "" {
				t.Error("item reward missing item ID")
			}
		case RewardTitle:
			if reward.Title == "" {
				t.Error("title reward missing title")
			}
		case RewardMount:
			if reward.MountID == "" {
				t.Error("mount reward missing mount ID")
			}
		case RewardCompanion:
			if reward.CompanionID == "" {
				t.Error("companion reward missing companion ID")
			}
		case RewardAchievement:
			if reward.AchievementID == "" {
				t.Error("achievement reward missing achievement ID")
			}
		case RewardAccountBonus:
			if reward.AccountBonusID == "" {
				t.Error("account bonus reward missing bonus ID")
			}
			if reward.BonusPercent < 5.0 || reward.BonusPercent > 10.0 {
				t.Errorf("account bonus percent = %f, want 5.0-10.0", reward.BonusPercent)
			}
		}
	}
}

// Benchmarks
func BenchmarkGenerate(b *testing.B) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.9,
		Depth:      50,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"player_level":    50,
			"servers_visited": 3,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate(12345, params)
	}
}

func BenchmarkValidate(b *testing.B) {
	gen := NewGenerator()
	quest := &LegendaryQuest{
		ID:              "test",
		Name:            "Test Quest",
		MinLevel:        50,
		EstimatedHours:  15,
		ServersRequired: 3,
		Phases: []*QuestPhase{
			{Type: PhaseExploration},
			{Type: PhaseCombat},
			{Type: PhaseCrafting},
			{Type: PhaseCollection},
			{Type: PhaseRaid},
			{Type: PhaseFinal},
		},
		Rewards: []*LegendaryReward{
			{Type: RewardItem, IsUnique: true},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gen.Validate(quest)
	}
}

func BenchmarkProgressTracker_UpdatePhase(b *testing.B) {
	tracker := NewProgressTracker()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.UpdatePhase("quest1", "player1", i%10, float64(i%100))
	}
}
