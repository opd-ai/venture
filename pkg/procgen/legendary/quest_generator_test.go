package legendary

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestLegendaryQuestGenerator_Generate(t *testing.T) {
	tests := []struct {
		name       string
		seed       int64
		difficulty float64
		depth      int
		wantErr    bool
	}{
		{
			name:       "basic legendary quest",
			seed:       12345,
			difficulty: 0.5,
			depth:      10,
			wantErr:    false,
		},
		{
			name:       "high difficulty quest",
			seed:       67890,
			difficulty: 0.9,
			depth:      20,
			wantErr:    false,
		},
		{
			name:       "low difficulty quest",
			seed:       11111,
			difficulty: 0.1,
			depth:      5,
			wantErr:    false,
		},
	}

	gen := NewLegendaryQuestGenerator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: tt.difficulty,
				Depth:      tt.depth,
				GenreID:    "fantasy",
			}

			result, err := gen.Generate(tt.seed, params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			quest, ok := result.(*LegendaryQuest)
			if !ok {
				t.Fatal("result is not a *LegendaryQuest")
			}

			// Check basic properties
			if quest.ID == "" {
				t.Error("quest ID is empty")
			}
			if quest.Name == "" {
				t.Error("quest name is empty")
			}
			if quest.Description == "" {
				t.Error("quest description is empty")
			}
			if quest.Seed != tt.seed {
				t.Errorf("quest seed = %d, want %d", quest.Seed, tt.seed)
			}

			// Check phase count
			if len(quest.Phases) < 5 || len(quest.Phases) > 10 {
				t.Errorf("quest has %d phases, want 5-10", len(quest.Phases))
			}

			// Check rewards
			if quest.Rewards == nil {
				t.Error("quest rewards are nil")
			}
		})
	}
}

func TestLegendaryQuestGenerator_Validate(t *testing.T) {
	gen := NewLegendaryQuestGenerator()

	tests := []struct {
		name    string
		quest   *LegendaryQuest
		wantErr bool
	}{
		{
			name: "valid quest",
			quest: &LegendaryQuest{
				ID:             "test_quest",
				Name:           "Test Quest",
				Description:    "Test description",
				Phases:         createValidPhases(7),
				Rewards:        createValidRewards(),
				EstimatedHours: 15.0,
			},
			wantErr: false,
		},
		{
			name: "too few phases",
			quest: &LegendaryQuest{
				ID:             "test_quest",
				Phases:         createValidPhases(3),
				Rewards:        createValidRewards(),
				EstimatedHours: 15.0,
			},
			wantErr: true,
		},
		{
			name: "too many phases",
			quest: &LegendaryQuest{
				ID:             "test_quest",
				Phases:         createValidPhases(12),
				Rewards:        createValidRewards(),
				EstimatedHours: 15.0,
			},
			wantErr: true,
		},
		{
			name: "no travel phase",
			quest: &LegendaryQuest{
				ID:             "test_quest",
				Phases:         createPhasesWithoutTravel(7),
				Rewards:        createValidRewards(),
				EstimatedHours: 15.0,
			},
			wantErr: true,
		},
		{
			name: "insufficient servers",
			quest: &LegendaryQuest{
				ID:             "test_quest",
				Phases:         createPhasesWithInsufficientServers(7),
				Rewards:        createValidRewards(),
				EstimatedHours: 15.0,
			},
			wantErr: true,
		},
		{
			name: "no rewards",
			quest: &LegendaryQuest{
				ID:             "test_quest",
				Phases:         createValidPhases(7),
				Rewards:        nil,
				EstimatedHours: 15.0,
			},
			wantErr: true,
		},
		{
			name: "time too short",
			quest: &LegendaryQuest{
				ID:             "test_quest",
				Phases:         createValidPhases(7),
				Rewards:        createValidRewards(),
				EstimatedHours: 5.0,
			},
			wantErr: true,
		},
		{
			name: "time too long",
			quest: &LegendaryQuest{
				ID:             "test_quest",
				Phases:         createValidPhases(7),
				Rewards:        createValidRewards(),
				EstimatedHours: 25.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Validate(tt.quest)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLegendaryQuest_Progress(t *testing.T) {
	quest := &LegendaryQuest{
		Phases: []*QuestPhase{
			{PhaseNumber: 1, Completed: true},
			{PhaseNumber: 2, Completed: true},
			{PhaseNumber: 3, Completed: false},
			{PhaseNumber: 4, Completed: false},
		},
	}

	progress := quest.Progress()
	expected := 0.5 // 2 out of 4 phases complete
	if progress != expected {
		t.Errorf("Progress() = %f, want %f", progress, expected)
	}
}

func TestLegendaryQuest_CurrentPhase(t *testing.T) {
	quest := &LegendaryQuest{
		Phases: []*QuestPhase{
			{PhaseNumber: 1, Completed: true},
			{PhaseNumber: 2, Completed: false, Name: "Current Phase"},
			{PhaseNumber: 3, Completed: false},
		},
	}

	current := quest.CurrentPhase()
	if current == nil {
		t.Fatal("CurrentPhase() returned nil")
	}
	if current.PhaseNumber != 2 {
		t.Errorf("CurrentPhase() phase number = %d, want 2", current.PhaseNumber)
	}
}

func TestLegendaryQuest_IsComplete(t *testing.T) {
	tests := []struct {
		name     string
		phases   []*QuestPhase
		expected bool
	}{
		{
			name: "all complete",
			phases: []*QuestPhase{
				{PhaseNumber: 1, Completed: true},
				{PhaseNumber: 2, Completed: true},
			},
			expected: true,
		},
		{
			name: "some incomplete",
			phases: []*QuestPhase{
				{PhaseNumber: 1, Completed: true},
				{PhaseNumber: 2, Completed: false},
			},
			expected: false,
		},
		{
			name:     "no phases",
			phases:   []*QuestPhase{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quest := &LegendaryQuest{Phases: tt.phases}
			result := quest.IsComplete()
			if result != tt.expected {
				t.Errorf("IsComplete() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestQuestPhase_PhaseProgress(t *testing.T) {
	tests := []struct {
		name     string
		phase    *QuestPhase
		expected float64
	}{
		{
			name: "kill targets partial",
			phase: &QuestPhase{
				Requirements: &PhaseRequirements{
					KillTargets: map[string]int{
						"dragon": 10,
						"demon":  5,
					},
					KillCompleted: map[string]int{
						"dragon": 5,
						"demon":  5,
					},
				},
			},
			expected: 0.5, // 1 out of 2 targets complete
		},
		{
			name: "all complete",
			phase: &QuestPhase{
				Requirements: &PhaseRequirements{
					KillTargets: map[string]int{
						"dragon": 10,
					},
					KillCompleted: map[string]int{
						"dragon": 10,
					},
				},
			},
			expected: 1.0,
		},
		{
			name: "no requirements",
			phase: &QuestPhase{
				Requirements: NewPhaseRequirements(),
			},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progress := tt.phase.PhaseProgress()
			if progress != tt.expected {
				t.Errorf("PhaseProgress() = %f, want %f", progress, tt.expected)
			}
		})
	}
}

func TestPhaseType_String(t *testing.T) {
	tests := []struct {
		phaseType PhaseType
		expected  string
	}{
		{PhaseKill, "Kill"},
		{PhaseCollect, "Collect"},
		{PhaseCraft, "Craft"},
		{PhaseRaid, "Raid"},
		{PhaseTravel, "Travel"},
		{PhaseExplore, "Explore"},
		{PhaseTalk, "Talk"},
		{PhaseChallenge, "Challenge"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.phaseType.String()
			if result != tt.expected {
				t.Errorf("String() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestChallengeType_String(t *testing.T) {
	tests := []struct {
		challengeType ChallengeType
		expected      string
	}{
		{ChallengeSurvival, "Survival"},
		{ChallengePuzzle, "Puzzle"},
		{ChallengeCombat, "Combat"},
		{ChallengeSpeed, "Speed"},
		{ChallengePerfection, "Perfection"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.challengeType.String()
			if result != tt.expected {
				t.Errorf("String() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestRarity_String(t *testing.T) {
	tests := []struct {
		rarity   Rarity
		expected string
	}{
		{RarityCommon, "Common"},
		{RarityUncommon, "Uncommon"},
		{RarityRare, "Rare"},
		{RarityEpic, "Epic"},
		{RarityLegendary, "Legendary"},
		{Rarity(99), "Unknown"}, // Test default case
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.rarity.String()
			if result != tt.expected {
				t.Errorf("String() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestAddExploreRequirements(t *testing.T) {
	gen := NewLegendaryQuestGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      10,
		GenreID:    "fantasy",
	}

	tests := []struct {
		name       string
		seed       int64
		minExpect  int
		maxExpect  int
	}{
		{
			name:      "basic exploration",
			seed:      12345,
			minExpect: 3, // 3 + rng.Intn(5) gives 3-7 locations
			maxExpect: 7,
		},
		{
			name:      "different seed",
			seed:      67890,
			minExpect: 3,
			maxExpect: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(tt.seed))
			req := NewPhaseRequirements()

			gen.addExploreRequirements(rng, req, params)

			// Verify locations are set
			if len(req.LocationsToDiscover) < tt.minExpect {
				t.Errorf("expected at least %d locations, got %d", tt.minExpect, len(req.LocationsToDiscover))
			}
			if len(req.LocationsToDiscover) > tt.maxExpect {
				t.Errorf("expected at most %d locations, got %d", tt.maxExpect, len(req.LocationsToDiscover))
			}

			// Verify location names are properly formatted
			for i, loc := range req.LocationsToDiscover {
				expectedLoc := fmt.Sprintf("location_%d", i+1)
				if loc != expectedLoc {
					t.Errorf("location[%d] = %s, want %s", i, loc, expectedLoc)
				}
			}
		})
	}
}

func TestExplorePhaseGeneration(t *testing.T) {
	gen := NewLegendaryQuestGenerator()

	// Generate many quests to try to get an explore phase
	// Since phase types are randomly selected, we may need multiple attempts
	foundExplore := false
	for seed := int64(0); seed < 100; seed++ {
		params := procgen.GenerationParams{
			Difficulty: 0.5,
			Depth:      10,
			GenreID:    "fantasy",
		}

		result, err := gen.Generate(seed, params)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		quest := result.(*LegendaryQuest)
		for _, phase := range quest.Phases {
			if phase.Type == PhaseExplore {
				foundExplore = true

				// Verify explore phase has locations
				if phase.Requirements == nil {
					t.Error("explore phase has nil requirements")
				} else if len(phase.Requirements.LocationsToDiscover) == 0 {
					t.Error("explore phase has no locations to discover")
				}
				break
			}
		}

		if foundExplore {
			break
		}
	}

	// Note: PhaseExplore might not appear if templates don't include it
	// This is informational, not a test failure
	if !foundExplore {
		t.Log("Note: no explore phases found in 100 generated quests (phase selection is random)")
	}
}

func TestDeterministicGeneration(t *testing.T) {
	gen := NewLegendaryQuestGenerator()
	seed := int64(42)
	params := procgen.GenerationParams{
		Difficulty: 0.7,
		Depth:      15,
		GenreID:    "fantasy",
	}

	// Generate twice with same seed
	result1, err1 := gen.Generate(seed, params)
	if err1 != nil {
		t.Fatalf("first Generate() error = %v", err1)
	}

	result2, err2 := gen.Generate(seed, params)
	if err2 != nil {
		t.Fatalf("second Generate() error = %v", err2)
	}

	quest1 := result1.(*LegendaryQuest)
	quest2 := result2.(*LegendaryQuest)

	// Compare quests
	if quest1.ID != quest2.ID {
		t.Errorf("quest IDs differ: %s vs %s", quest1.ID, quest2.ID)
	}
	if quest1.Name != quest2.Name {
		t.Errorf("quest names differ: %s vs %s", quest1.Name, quest2.Name)
	}
	if len(quest1.Phases) != len(quest2.Phases) {
		t.Errorf("phase counts differ: %d vs %d", len(quest1.Phases), len(quest2.Phases))
	}
	if quest1.EstimatedHours != quest2.EstimatedHours {
		t.Errorf("estimated hours differ: %f vs %f", quest1.EstimatedHours, quest2.EstimatedHours)
	}
}

func TestCrossServerRequirement(t *testing.T) {
	gen := NewLegendaryQuestGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      10,
		GenreID:    "fantasy",
	}

	// Generate 10 quests and verify all have cross-server requirements
	for i := 0; i < 10; i++ {
		result, err := gen.Generate(int64(i), params)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		quest := result.(*LegendaryQuest)

		// Verify quest has travel phase with >= 3 servers
		hasTravel := false
		minServers := 0
		for _, phase := range quest.Phases {
			if phase.Type == PhaseTravel && phase.Requirements != nil {
				hasTravel = true
				if phase.Requirements.MinServers > minServers {
					minServers = phase.Requirements.MinServers
				}
			}
		}

		if !hasTravel {
			t.Errorf("quest %d has no travel phase", i)
		}
		if minServers < 3 {
			t.Errorf("quest %d requires only %d servers, need >= 3", i, minServers)
		}
	}
}

func TestRaidRequirement(t *testing.T) {
	gen := NewLegendaryQuestGenerator()

	// High difficulty should generate raids
	params := procgen.GenerationParams{
		Difficulty: 0.9,
		Depth:      20,
		GenreID:    "fantasy",
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	quest := result.(*LegendaryQuest)

	// Check for raid phase
	hasRaid := false
	for _, phase := range quest.Phases {
		if phase.Type == PhaseRaid {
			hasRaid = true

			if phase.Requirements == nil || len(phase.Requirements.RaidEncounters) == 0 {
				t.Error("raid phase has no raid encounters")
			}

			// Verify raid requirements
			for _, raid := range phase.Requirements.RaidEncounters {
				if raid.RaidID == "" {
					t.Error("raid has empty ID")
				}
				if raid.MinPartySize < 5 || raid.MinPartySize > 10 {
					t.Errorf("raid party size %d out of range 5-10", raid.MinPartySize)
				}
			}
			break
		}
	}

	if !hasRaid {
		t.Log("Note: high difficulty quest may not always have raid phase (random)")
	}
}

func TestCraftingRequirement(t *testing.T) {
	gen := NewLegendaryQuestGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.7,
		Depth:      15,
		GenreID:    "fantasy",
	}

	// Generate multiple quests to find one with crafting
	foundCrafting := false
	for i := 0; i < 10; i++ {
		result, err := gen.Generate(int64(i), params)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		quest := result.(*LegendaryQuest)

		for _, phase := range quest.Phases {
			if phase.Type == PhaseCraft {
				foundCrafting = true

				if phase.Requirements == nil || len(phase.Requirements.CraftItems) == 0 {
					t.Error("craft phase has no craft items")
				}

				// Verify craft requirements
				for _, craft := range phase.Requirements.CraftItems {
					if craft.ItemType == "" {
						t.Error("craft item has empty type")
					}
					if craft.Quantity < 1 {
						t.Error("craft item has invalid quantity")
					}
					if craft.StationQuality == "" {
						t.Error("craft item has no station quality")
					}
				}
				break
			}
		}

		if foundCrafting {
			break
		}
	}

	if !foundCrafting {
		t.Log("Note: no crafting phases found in sample (random generation)")
	}
}

// Helper functions

func createValidPhases(count int) []*QuestPhase {
	phases := make([]*QuestPhase, count)
	for i := 0; i < count; i++ {
		phaseType := PhaseKill
		if i == count/2 {
			phaseType = PhaseTravel
		}

		req := NewPhaseRequirements()
		if phaseType == PhaseTravel {
			req.MinServers = 3
			req.ServersToVisit = []string{"server1", "server2", "server3"}
		}

		phases[i] = &QuestPhase{
			PhaseNumber:  i + 1,
			Name:         fmt.Sprintf("Phase %d", i+1),
			Type:         phaseType,
			Requirements: req,
		}
	}
	return phases
}

func createPhasesWithoutTravel(count int) []*QuestPhase {
	phases := make([]*QuestPhase, count)
	for i := 0; i < count; i++ {
		phases[i] = &QuestPhase{
			PhaseNumber:  i + 1,
			Name:         fmt.Sprintf("Phase %d", i+1),
			Type:         PhaseKill,
			Requirements: NewPhaseRequirements(),
		}
	}
	return phases
}

func createPhasesWithInsufficientServers(count int) []*QuestPhase {
	phases := make([]*QuestPhase, count)
	for i := 0; i < count; i++ {
		phaseType := PhaseKill
		req := NewPhaseRequirements()

		if i == count/2 {
			phaseType = PhaseTravel
			req.MinServers = 2 // Insufficient
			req.ServersToVisit = []string{"server1", "server2"}
		}

		phases[i] = &QuestPhase{
			PhaseNumber:  i + 1,
			Name:         fmt.Sprintf("Phase %d", i+1),
			Type:         phaseType,
			Requirements: req,
		}
	}
	return phases
}

func createValidRewards() *LegendaryRewards {
	return &LegendaryRewards{
		Items: []LegendaryItem{
			{Name: "Test Item", Rarity: RarityLegendary},
		},
		Titles:         []string{"Test Title"},
		Gold:           100000,
		Experience:     50000,
		PrestigeLevels: 1,
		Achievements:   []string{"Test Achievement"},
		Cosmetics:      []string{"Test Cosmetic"},
	}
}

// Benchmarks

func BenchmarkLegendaryQuestGeneration(b *testing.B) {
	gen := NewLegendaryQuestGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.7,
		Depth:      15,
		GenreID:    "fantasy",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(int64(i), params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkQuestValidation(b *testing.B) {
	gen := NewLegendaryQuestGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.7,
		Depth:      15,
		GenreID:    "fantasy",
	}

	result, _ := gen.Generate(12345, params)
	quest := result.(*LegendaryQuest)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gen.Validate(quest)
	}
}

func BenchmarkPhaseProgress(b *testing.B) {
	phase := &QuestPhase{
		Requirements: &PhaseRequirements{
			KillTargets: map[string]int{
				"dragon": 50,
				"demon":  30,
				"undead": 20,
			},
			KillCompleted: map[string]int{
				"dragon": 25,
				"demon":  15,
				"undead": 10,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = phase.PhaseProgress()
	}
}
