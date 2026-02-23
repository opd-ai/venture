package narrative_world

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/companion/learning"
	"github.com/opd-ai/venture/pkg/engine"
)

func TestNewStoryEventManager(t *testing.T) {
	manager := NewStoryEventManager(12345)

	if manager == nil {
		t.Fatal("expected non-nil manager")
	}

	if manager.seed != 12345 {
		t.Errorf("expected seed 12345, got %d", manager.seed)
	}

	if manager.conflictChance != 0.15 {
		t.Errorf("expected default conflict chance 0.15, got %.2f", manager.conflictChance)
	}

	if manager.maxMemoryEvents != 75 {
		t.Errorf("expected default max memory 75, got %d", manager.maxMemoryEvents)
	}

	// Verify quest templates initialized
	for compType := engine.CompanionTypePet; compType <= engine.CompanionTypeInsect; compType++ {
		templates, exists := manager.questTemplates[compType]
		if !exists {
			t.Errorf("missing quest templates for companion type %d", compType)
			continue
		}

		if len(templates) < 3 {
			t.Errorf("companion type %d has %d templates, expected at least 3", compType, len(templates))
		}
	}
}

func TestNewStoryEventManagerWithTimeProvider(t *testing.T) {
	fixedTime := int64(1000000)
	tp := &FixedTimeProvider{Timestamp: fixedTime}

	manager := NewStoryEventManager(12345, WithTimeProvider(tp))

	if manager == nil {
		t.Fatal("expected non-nil manager")
	}

	if manager.timeProvider != tp {
		t.Error("expected custom TimeProvider to be set")
	}

	// Verify the manager uses the custom time provider
	companionID := uint64(100)
	manager.RecordMemory(companionID, EventTypeCombat, "Test event")

	context := manager.GetDialogueContext(companionID, 1)
	if len(context.RecentEvents) == 0 {
		t.Fatal("expected at least one memory")
	}

	if context.RecentEvents[0].Timestamp != fixedTime {
		t.Errorf("expected timestamp %d, got %d", fixedTime, context.RecentEvents[0].Timestamp)
	}
}

func TestNewStoryEventManagerWithNilTimeProvider(t *testing.T) {
	// Passing nil TimeProvider should not crash and should use default
	manager := NewStoryEventManager(12345, WithTimeProvider(nil))

	if manager == nil {
		t.Fatal("expected non-nil manager")
	}

	if manager.timeProvider != nil {
		t.Error("expected nil TimeProvider (use package default)")
	}
}

func TestGeneratePersonalQuest(t *testing.T) {
	manager := NewStoryEventManager(12345)

	tests := []struct {
		name          string
		companion     *engine.CompanionComponent
		seed          int64
		expectError   bool
		errorContains string
	}{
		{
			name: "valid pet quest at 0.7 loyalty",
			companion: &engine.CompanionComponent{
				CompanionType: engine.CompanionTypePet,
				Loyalty:       0.7,
				Level:         5,
			},
			seed:        54321,
			expectError: false,
		},
		{
			name: "valid hireling quest at 0.9 loyalty",
			companion: &engine.CompanionComponent{
				CompanionType: engine.CompanionTypeHireling,
				Loyalty:       0.9,
				Level:         10,
			},
			seed:        11111,
			expectError: false,
		},
		{
			name: "insufficient loyalty",
			companion: &engine.CompanionComponent{
				CompanionType: engine.CompanionTypePet,
				Loyalty:       0.6,
				Level:         5,
			},
			seed:          22222,
			expectError:   true,
			errorContains: "below minimum 0.7",
		},
		{
			name: "valid elemental quest at minimum loyalty",
			companion: &engine.CompanionComponent{
				CompanionType: engine.CompanionTypeElemental,
				Loyalty:       0.7,
				Level:         5,
			},
			seed:        33333,
			expectError: false,
		},
		{
			name: "valid spirit quest at minimum loyalty",
			companion: &engine.CompanionComponent{
				CompanionType: engine.CompanionTypeSpirit,
				Loyalty:       0.7,
				Level:         5,
			},
			seed:        44444,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quest, err := manager.GeneratePersonalQuest(1, tt.companion, tt.seed)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if quest == nil {
				t.Fatal("expected non-nil quest")
			}

			if quest.CompanionID != 1 {
				t.Errorf("expected companion ID 1, got %d", quest.CompanionID)
			}

			if quest.CompanionType != tt.companion.CompanionType {
				t.Errorf("expected companion type %d, got %d", tt.companion.CompanionType, quest.CompanionType)
			}

			if len(quest.Objectives) < 2 || len(quest.Objectives) > 4 {
				t.Errorf("expected 2-4 objectives, got %d", len(quest.Objectives))
			}

			if len(quest.Consequences) == 0 {
				t.Error("expected at least one consequence")
			}

			if !quest.Consequences[0].Permanent {
				t.Error("expected permanent consequence")
			}

			if quest.StoryBranches == nil {
				t.Error("expected non-nil story branches")
			}

			// Verify quest is tracked
			activeQuests := manager.GetActiveQuests(1)
			if len(activeQuests) == 0 {
				t.Error("expected active quest to be tracked")
			}
		})
	}
}

func TestGeneratePersonalQuestDeterminism(t *testing.T) {
	manager1 := NewStoryEventManager(12345)
	manager2 := NewStoryEventManager(12345)

	companion := &engine.CompanionComponent{
		CompanionType: engine.CompanionTypePet,
		Loyalty:       0.8,
		Level:         5,
	}

	quest1, err1 := manager1.GeneratePersonalQuest(1, companion, 99999)
	quest2, err2 := manager2.GeneratePersonalQuest(1, companion, 99999)

	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected errors: %v, %v", err1, err2)
	}

	if quest1.Title != quest2.Title {
		t.Errorf("titles differ: %q vs %q", quest1.Title, quest2.Title)
	}

	if len(quest1.Objectives) != len(quest2.Objectives) {
		t.Errorf("objective count differs: %d vs %d", len(quest1.Objectives), len(quest2.Objectives))
	}
}

func TestRecordMemory(t *testing.T) {
	manager := NewStoryEventManager(12345)

	companionID := uint64(100)

	// Record single event
	manager.RecordMemory(companionID, EventTypeCombat, "Defeated goblin")

	count := manager.GetMemoryCount(companionID)
	if count != 1 {
		t.Errorf("expected 1 memory, got %d", count)
	}

	totalRecorded := manager.GetTotalEventsRecorded(companionID)
	if totalRecorded != 1 {
		t.Errorf("expected 1 total event, got %d", totalRecorded)
	}

	// Record multiple events
	for i := 0; i < 10; i++ {
		manager.RecordMemory(companionID, EventTypeTreasure, "Found gold")
	}

	count = manager.GetMemoryCount(companionID)
	if count != 11 {
		t.Errorf("expected 11 memories, got %d", count)
	}

	// Verify total count includes all
	totalRecorded = manager.GetTotalEventsRecorded(companionID)
	if totalRecorded != 11 {
		t.Errorf("expected 11 total events, got %d", totalRecorded)
	}
}

func TestMemoryPruning(t *testing.T) {
	// Use incrementing time provider for deterministic timestamps
	SetTimeProvider(&IncrementingTimeProvider{Current: 1000000, Step: 60}) // 60 second increments
	defer ResetTimeProvider()

	manager := NewStoryEventManager(12345)
	manager.SetMaxMemoryEvents(10) // Set low limit for testing

	companionID := uint64(200)

	// Record more events than limit
	for i := 0; i < 20; i++ {
		eventType := EventTypeCombat
		if i%5 == 0 {
			eventType = EventTypeSacrifice // High importance
		}
		manager.RecordMemory(companionID, eventType, "Test event")
		// No time.Sleep needed - IncrementingTimeProvider ensures different timestamps
	}

	count := manager.GetMemoryCount(companionID)
	if count > 10 {
		t.Errorf("expected max 10 memories after pruning, got %d", count)
	}

	totalRecorded := manager.GetTotalEventsRecorded(companionID)
	if totalRecorded != 20 {
		t.Errorf("expected 20 total events recorded, got %d", totalRecorded)
	}

	// Verify high importance events are retained
	memory := manager.memories[companionID]
	importantCount := 0
	for _, event := range memory.Events {
		if event.Type == EventTypeSacrifice {
			importantCount++
		}
	}

	if importantCount == 0 {
		t.Error("expected some high-importance events to be retained")
	}
}

func TestCheckConflict(t *testing.T) {
	manager := NewStoryEventManager(12345)
	manager.SetConflictChance(1.0) // Always generate conflict for testing

	comp1 := &engine.CompanionComponent{
		CompanionType: engine.CompanionTypePet,
	}
	comp2 := &engine.CompanionComponent{
		CompanionType: engine.CompanionTypeHireling,
	}

	// Create opposing personalities
	personality1 := &learning.PersonalityEvolution{
		Traits: map[learning.PersonalityTrait]float64{
			learning.TraitAggressive: 0.9,
			learning.TraitBrave:      0.8,
		},
	}

	personality2 := &learning.PersonalityEvolution{
		Traits: map[learning.PersonalityTrait]float64{
			learning.TraitPacifist: 0.9,
			learning.TraitCautious: 0.8,
		},
	}

	conflict, exists := manager.CheckConflict(comp1, comp2, 1, 2, personality1, personality2)

	if !exists {
		t.Fatal("expected conflict to be detected")
	}

	if conflict.Companion1 != 1 || conflict.Companion2 != 2 {
		t.Errorf("unexpected companion IDs: %d, %d", conflict.Companion1, conflict.Companion2)
	}

	if conflict.Severity < 0.3 || conflict.Severity > 1.0 {
		t.Errorf("severity %.2f out of range [0.3, 1.0]", conflict.Severity)
	}

	if !conflict.Active {
		t.Error("conflict should be active")
	}

	// Verify conflict is tracked
	activeConflicts := manager.GetActiveConflicts()
	if len(activeConflicts) != 1 {
		t.Errorf("expected 1 active conflict, got %d", len(activeConflicts))
	}

	// Check same pair again - should return existing conflict
	conflict2, exists2 := manager.CheckConflict(comp1, comp2, 1, 2, personality1, personality2)
	if !exists2 {
		t.Error("expected existing conflict to be found")
	}

	if conflict2.ConflictType != conflict.ConflictType {
		t.Error("should return same conflict")
	}
}

func TestConflictProbabilityCalculation(t *testing.T) {
	manager := NewStoryEventManager(12345)

	tests := []struct {
		name    string
		p1      *learning.PersonalityEvolution
		p2      *learning.PersonalityEvolution
		minProb float64
		maxProb float64
	}{
		{
			name: "opposing aggressive/pacifist",
			p1: &learning.PersonalityEvolution{
				Traits: map[learning.PersonalityTrait]float64{
					learning.TraitAggressive: 0.9,
				},
			},
			p2: &learning.PersonalityEvolution{
				Traits: map[learning.PersonalityTrait]float64{
					learning.TraitPacifist: 0.9,
				},
			},
			minProb: 0.7,
			maxProb: 1.0,
		},
		{
			name: "no strong traits",
			p1: &learning.PersonalityEvolution{
				Traits: map[learning.PersonalityTrait]float64{
					learning.TraitCautious: 0.3,
				},
			},
			p2: &learning.PersonalityEvolution{
				Traits: map[learning.PersonalityTrait]float64{
					learning.TraitBrave: 0.3,
				},
			},
			minProb: 0.4,
			maxProb: 0.6,
		},
		{
			name:    "nil personalities",
			p1:      nil,
			p2:      nil,
			minProb: 0.5,
			maxProb: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prob := manager.calculateConflictProbability(tt.p1, tt.p2)

			if prob < tt.minProb || prob > tt.maxProb {
				t.Errorf("probability %.2f outside expected range [%.2f, %.2f]", prob, tt.minProb, tt.maxProb)
			}
		})
	}
}

func TestGenerateCrossCompanionStory(t *testing.T) {
	manager := NewStoryEventManager(12345)

	// Record some memories first
	manager.RecordMemory(1, EventTypeBonding, "Companion 1 and 2 fought together")
	manager.RecordMemory(2, EventTypeBonding, "Companion 1 and 2 fought together")
	manager.RecordMemory(3, EventTypeConflict, "Companion 3 disagreed")

	story, err := manager.GenerateCrossCompanionStory([]uint64{1, 2, 3}, 77777)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if story == nil {
		t.Fatal("expected non-nil story")
	}

	if len(story.Participants) != 3 {
		t.Errorf("expected 3 participants, got %d", len(story.Participants))
	}

	if story.Title == "" {
		t.Error("expected non-empty title")
	}

	if story.Description == "" {
		t.Error("expected non-empty description")
	}

	if !story.Active {
		t.Error("story should be active")
	}

	if story.Outcome != OutcomeUnresolved {
		t.Errorf("expected unresolved outcome, got %v", story.Outcome)
	}

	if story.Narrative == nil {
		t.Error("expected non-nil narrative")
	}

	// Verify story is tracked
	activeStories := manager.GetActiveCrossStories()
	if len(activeStories) != 1 {
		t.Errorf("expected 1 active story, got %d", len(activeStories))
	}
}

func TestGenerateCrossCompanionStoryInsufficientCompanions(t *testing.T) {
	manager := NewStoryEventManager(12345)

	_, err := manager.GenerateCrossCompanionStory([]uint64{1}, 12345)

	if err == nil {
		t.Fatal("expected error for insufficient companions")
	}

	if !contains(err.Error(), "at least 2 companions") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestGetDialogueContext(t *testing.T) {
	manager := NewStoryEventManager(12345)

	companionID := uint64(300)

	// Record various events
	manager.RecordMemory(companionID, EventTypeCombat, "Minor combat")   // Low importance
	manager.RecordMemory(companionID, EventTypeSacrifice, "Saved owner") // High importance
	manager.RecordMemory(companionID, EventTypeTreasure, "Found gold")
	manager.RecordMemory(companionID, EventTypeBetray, "Witnessed betrayal") // High importance

	context := manager.GetDialogueContext(companionID, 10)

	if context == nil {
		t.Fatal("expected non-nil context")
	}

	if len(context.RecentEvents) != 4 {
		t.Errorf("expected 4 recent events, got %d", len(context.RecentEvents))
	}

	// Check important events (sacrifice and betray have importance >= 0.7)
	if len(context.ImportantEvents) < 2 {
		t.Errorf("expected at least 2 important events, got %d", len(context.ImportantEvents))
	}

	// Verify important events are actually important
	for _, event := range context.ImportantEvents {
		if event.Importance < 0.7 {
			t.Errorf("event marked important has importance %.2f < 0.7", event.Importance)
		}
	}
}

func TestGetDialogueContextNoMemories(t *testing.T) {
	manager := NewStoryEventManager(12345)

	context := manager.GetDialogueContext(999, 10)

	if context == nil {
		t.Fatal("expected non-nil context even with no memories")
	}

	if len(context.RecentEvents) != 0 {
		t.Errorf("expected 0 recent events, got %d", len(context.RecentEvents))
	}

	if len(context.ImportantEvents) != 0 {
		t.Errorf("expected 0 important events, got %d", len(context.ImportantEvents))
	}
}

func TestUpdateConflict(t *testing.T) {
	manager := NewStoryEventManager(12345)
	manager.SetConflictChance(1.0)

	comp1 := &engine.CompanionComponent{CompanionType: engine.CompanionTypePet}
	comp2 := &engine.CompanionComponent{CompanionType: engine.CompanionTypeHireling}

	p1 := &learning.PersonalityEvolution{
		Traits: map[learning.PersonalityTrait]float64{learning.TraitAggressive: 0.9},
	}
	p2 := &learning.PersonalityEvolution{
		Traits: map[learning.PersonalityTrait]float64{learning.TraitPacifist: 0.9},
	}

	conflict, _ := manager.CheckConflict(comp1, comp2, 1, 2, p1, p2)
	initialSeverity := conflict.Severity

	// Update conflict over simulated time
	manager.UpdateConflict(0, 24*time.Hour) // 1 day

	updatedConflict := manager.conflicts[0]
	if updatedConflict.Severity <= initialSeverity {
		t.Error("severity should increase over time")
	}

	if updatedConflict.TimeSinceStart != 24*time.Hour {
		t.Errorf("expected 24h elapsed, got %v", updatedConflict.TimeSinceStart)
	}
}

func TestResolveConflict(t *testing.T) {
	manager := NewStoryEventManager(12345)
	manager.SetConflictChance(1.0)

	comp1 := &engine.CompanionComponent{CompanionType: engine.CompanionTypePet}
	comp2 := &engine.CompanionComponent{CompanionType: engine.CompanionTypeHireling}

	p1 := &learning.PersonalityEvolution{
		Traits: map[learning.PersonalityTrait]float64{learning.TraitAggressive: 0.9},
	}
	p2 := &learning.PersonalityEvolution{
		Traits: map[learning.PersonalityTrait]float64{learning.TraitPacifist: 0.9},
	}

	manager.CheckConflict(comp1, comp2, 1, 2, p1, p2)

	resolved := manager.ResolveConflict(1, 2)
	if !resolved {
		t.Fatal("expected conflict to be resolved")
	}

	activeConflicts := manager.GetActiveConflicts()
	if len(activeConflicts) != 0 {
		t.Errorf("expected 0 active conflicts after resolution, got %d", len(activeConflicts))
	}
}

func TestCompleteQuest(t *testing.T) {
	manager := NewStoryEventManager(12345)

	companion := &engine.CompanionComponent{
		CompanionType: engine.CompanionTypePet,
		Loyalty:       0.8,
		Level:         5,
	}

	quest, err := manager.GeneratePersonalQuest(1, companion, 12345)
	if err != nil {
		t.Fatalf("failed to generate quest: %v", err)
	}

	// Mark all objectives as completed
	for i := range quest.Objectives {
		quest.Objectives[i].Progress = quest.Objectives[i].Required
		quest.Objectives[i].Completed = true
	}

	completedQuest, err := manager.CompleteQuest(1, quest.QuestID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !completedQuest.Completed {
		t.Error("quest should be marked completed")
	}

	// Verify memory was recorded
	count := manager.GetMemoryCount(1)
	if count == 0 {
		t.Error("expected quest completion to be recorded as memory")
	}
}

func TestCompleteQuestIncompleteObjectives(t *testing.T) {
	manager := NewStoryEventManager(12345)

	companion := &engine.CompanionComponent{
		CompanionType: engine.CompanionTypePet,
		Loyalty:       0.8,
		Level:         5,
	}

	quest, _ := manager.GeneratePersonalQuest(1, companion, 12345)

	// Don't complete objectives
	_, err := manager.CompleteQuest(1, quest.QuestID)
	if err == nil {
		t.Fatal("expected error for incomplete objectives")
	}

	if !contains(err.Error(), "incomplete objectives") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestUpdateQuestObjective(t *testing.T) {
	manager := NewStoryEventManager(12345)

	companion := &engine.CompanionComponent{
		CompanionType: engine.CompanionTypePet,
		Loyalty:       0.8,
		Level:         5,
	}

	quest, _ := manager.GeneratePersonalQuest(1, companion, 12345)

	err := manager.UpdateQuestObjective(1, quest.QuestID, 0, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if quest.Objectives[0].Progress != 3 {
		t.Errorf("expected progress 3, got %d", quest.Objectives[0].Progress)
	}

	// Update to completion
	err = manager.UpdateQuestObjective(1, quest.QuestID, 0, quest.Objectives[0].Required)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !quest.Objectives[0].Completed {
		t.Error("objective should be marked completed")
	}
}

func TestSetters(t *testing.T) {
	manager := NewStoryEventManager(12345)

	// Test SetConflictChance
	manager.SetConflictChance(0.25)
	if manager.conflictChance != 0.25 {
		t.Errorf("expected conflict chance 0.25, got %.2f", manager.conflictChance)
	}

	// Test bounds
	manager.SetConflictChance(-0.1)
	if manager.conflictChance != 0.0 {
		t.Errorf("expected conflict chance clamped to 0.0, got %.2f", manager.conflictChance)
	}

	manager.SetConflictChance(1.5)
	if manager.conflictChance != 1.0 {
		t.Errorf("expected conflict chance clamped to 1.0, got %.2f", manager.conflictChance)
	}

	// Test SetMaxMemoryEvents
	manager.SetMaxMemoryEvents(50)
	if manager.maxMemoryEvents != 50 {
		t.Errorf("expected max memory 50, got %d", manager.maxMemoryEvents)
	}

	// Test bounds
	manager.SetMaxMemoryEvents(5)
	if manager.maxMemoryEvents != 10 {
		t.Errorf("expected max memory clamped to 10, got %d", manager.maxMemoryEvents)
	}

	manager.SetMaxMemoryEvents(300)
	if manager.maxMemoryEvents != 200 {
		t.Errorf("expected max memory clamped to 200, got %d", manager.maxMemoryEvents)
	}
}

// Benchmark tests
func BenchmarkGeneratePersonalQuest(b *testing.B) {
	manager := NewStoryEventManager(12345)
	companion := &engine.CompanionComponent{
		CompanionType: engine.CompanionTypePet,
		Loyalty:       0.8,
		Level:         5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.GeneratePersonalQuest(uint64(i), companion, int64(i))
	}
}

func BenchmarkRecordMemory(b *testing.B) {
	manager := NewStoryEventManager(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.RecordMemory(1, EventTypeCombat, "Test event")
	}
}

func BenchmarkCheckConflict(b *testing.B) {
	manager := NewStoryEventManager(12345)
	comp1 := &engine.CompanionComponent{CompanionType: engine.CompanionTypePet}
	comp2 := &engine.CompanionComponent{CompanionType: engine.CompanionTypeHireling}
	p1 := &learning.PersonalityEvolution{
		Traits: map[learning.PersonalityTrait]float64{learning.TraitAggressive: 0.9},
	}
	p2 := &learning.PersonalityEvolution{
		Traits: map[learning.PersonalityTrait]float64{learning.TraitPacifist: 0.9},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.CheckConflict(comp1, comp2, 1, 2, p1, p2)
	}
}

func BenchmarkGetDialogueContext(b *testing.B) {
	manager := NewStoryEventManager(12345)

	// Populate with memories
	for i := 0; i < 100; i++ {
		manager.RecordMemory(1, EventTypeCombat, "Test event")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.GetDialogueContext(1, 10)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
