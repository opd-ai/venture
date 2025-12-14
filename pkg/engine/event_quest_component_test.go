package engine

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/procgen/quest"
)

func TestEventQuestComponent_Type(t *testing.T) {
	comp := NewEventQuestComponent(3)
	if got := comp.Type(); got != "event_quest" {
		t.Errorf("Type() = %v, want %v", got, "event_quest")
	}
}

func TestNewEventQuestComponent(t *testing.T) {
	tests := []struct {
		name      string
		maxActive int
		wantMax   int
	}{
		{"default max", 3, 3},
		{"custom max", 5, 5},
		{"zero defaults to 3", 0, 3},
		{"negative defaults to 3", -1, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewEventQuestComponent(tt.maxActive)
			if comp.MaxActiveEventQuests != tt.wantMax {
				t.Errorf("MaxActiveEventQuests = %v, want %v", comp.MaxActiveEventQuests, tt.wantMax)
			}
			if comp.ActiveQuests == nil {
				t.Error("ActiveQuests should not be nil")
			}
			if comp.AvailableQuests == nil {
				t.Error("AvailableQuests should not be nil")
			}
			if comp.CompletedQuests == nil {
				t.Error("CompletedQuests should not be nil")
			}
			if comp.ExpiredQuests == nil {
				t.Error("ExpiredQuests should not be nil")
			}
		})
	}
}

func TestEventQuestComponent_CanAcceptQuest(t *testing.T) {
	comp := NewEventQuestComponent(2)

	if !comp.CanAcceptQuest() {
		t.Error("Should be able to accept quest when no active quests")
	}

	// Add available quest and accept it
	comp.AvailableQuests = append(comp.AvailableQuests, EventQuestDefinition{
		ID:      "test_quest_1",
		EventID: "test_event",
	})
	comp.AcceptQuest("test_quest_1", time.Now().Add(24*time.Hour))

	if !comp.CanAcceptQuest() {
		t.Error("Should be able to accept quest when under max")
	}

	// Add and accept another
	comp.AvailableQuests = append(comp.AvailableQuests, EventQuestDefinition{
		ID:      "test_quest_2",
		EventID: "test_event",
	})
	comp.AcceptQuest("test_quest_2", time.Now().Add(24*time.Hour))

	if comp.CanAcceptQuest() {
		t.Error("Should not be able to accept quest when at max")
	}
}

func TestEventQuestComponent_AcceptQuest(t *testing.T) {
	comp := NewEventQuestComponent(2)

	// Add available quest
	comp.AvailableQuests = append(comp.AvailableQuests, EventQuestDefinition{
		ID:        "test_quest",
		EventID:   "test_event",
		Name:      "Test Quest",
		QuestType: EventQuestCollection,
		Objectives: []quest.Objective{
			{Description: "Collect items", Target: "item", Required: 5},
		},
	})

	expires := time.Now().Add(24 * time.Hour)
	if !comp.AcceptQuest("test_quest", expires) {
		t.Error("AcceptQuest should succeed")
	}

	if len(comp.ActiveQuests) != 1 {
		t.Errorf("Should have 1 active quest, got %d", len(comp.ActiveQuests))
	}

	if len(comp.AvailableQuests) != 0 {
		t.Errorf("Should have 0 available quests, got %d", len(comp.AvailableQuests))
	}

	activeQuest := comp.ActiveQuests[0]
	if activeQuest.Status != quest.StatusActive {
		t.Errorf("Quest status = %v, want %v", activeQuest.Status, quest.StatusActive)
	}

	if activeQuest.Progress[0] != 0 {
		t.Errorf("Progress[0] = %d, want 0", activeQuest.Progress[0])
	}
}

func TestEventQuestComponent_AcceptQuest_NotFound(t *testing.T) {
	comp := NewEventQuestComponent(2)

	if comp.AcceptQuest("nonexistent", time.Now().Add(24*time.Hour)) {
		t.Error("AcceptQuest should fail for nonexistent quest")
	}
}

func TestEventQuestComponent_UpdateProgress(t *testing.T) {
	comp := createTestEventQuestComponent(t)

	comp.UpdateProgress("test_quest", 0, 3)

	active := comp.GetActiveQuest("test_quest")
	if active == nil {
		t.Fatal("Expected active quest")
	}

	if active.Progress[0] != 3 {
		t.Errorf("Progress[0] = %d, want 3", active.Progress[0])
	}
}

func TestEventQuestComponent_IncrementProgress(t *testing.T) {
	comp := createTestEventQuestComponent(t)

	comp.IncrementProgress("test_quest", 0, 2)
	comp.IncrementProgress("test_quest", 0, 2)

	active := comp.GetActiveQuest("test_quest")
	if active.Progress[0] != 4 {
		t.Errorf("Progress[0] = %d, want 4", active.Progress[0])
	}

	// Test capping at required
	comp.IncrementProgress("test_quest", 0, 10)
	if active.Progress[0] != 5 { // Required is 5
		t.Errorf("Progress[0] = %d, want 5 (capped)", active.Progress[0])
	}
}

func TestEventQuestComponent_IsQuestComplete(t *testing.T) {
	comp := createTestEventQuestComponent(t)

	if comp.IsQuestComplete("test_quest") {
		t.Error("Quest should not be complete initially")
	}

	comp.UpdateProgress("test_quest", 0, 5)

	if !comp.IsQuestComplete("test_quest") {
		t.Error("Quest should be complete after reaching required")
	}
}

func TestEventQuestComponent_CompleteQuest(t *testing.T) {
	comp := createTestEventQuestComponent(t)

	if !comp.CompleteQuest("test_quest") {
		t.Error("CompleteQuest should succeed")
	}

	if len(comp.ActiveQuests) != 0 {
		t.Errorf("Should have 0 active quests, got %d", len(comp.ActiveQuests))
	}

	if len(comp.CompletedQuests) != 1 {
		t.Errorf("Should have 1 completed quest, got %d", len(comp.CompletedQuests))
	}

	completed := comp.CompletedQuests[0]
	if completed.Status != quest.StatusComplete {
		t.Errorf("Status = %v, want %v", completed.Status, quest.StatusComplete)
	}
}

func TestEventQuestComponent_ExpireQuest(t *testing.T) {
	comp := createTestEventQuestComponent(t)

	if !comp.ExpireQuest("test_quest") {
		t.Error("ExpireQuest should succeed")
	}

	if len(comp.ActiveQuests) != 0 {
		t.Errorf("Should have 0 active quests, got %d", len(comp.ActiveQuests))
	}

	if len(comp.ExpiredQuests) != 1 {
		t.Errorf("Should have 1 expired quest, got %d", len(comp.ExpiredQuests))
	}

	expired := comp.ExpiredQuests[0]
	if expired.Status != quest.StatusFailed {
		t.Errorf("Status = %v, want %v", expired.Status, quest.StatusFailed)
	}
}

func TestEventQuestComponent_GetActiveQuestsForEvent(t *testing.T) {
	comp := NewEventQuestComponent(5)

	comp.AvailableQuests = append(comp.AvailableQuests,
		EventQuestDefinition{ID: "q1", EventID: "event_a"},
		EventQuestDefinition{ID: "q2", EventID: "event_b"},
		EventQuestDefinition{ID: "q3", EventID: "event_a"},
	)

	comp.AcceptQuest("q1", time.Now().Add(time.Hour))
	comp.AcceptQuest("q2", time.Now().Add(time.Hour))
	comp.AcceptQuest("q3", time.Now().Add(time.Hour))

	eventAQuests := comp.GetActiveQuestsForEvent("event_a")
	if len(eventAQuests) != 2 {
		t.Errorf("Expected 2 quests for event_a, got %d", len(eventAQuests))
	}

	eventBQuests := comp.GetActiveQuestsForEvent("event_b")
	if len(eventBQuests) != 1 {
		t.Errorf("Expected 1 quest for event_b, got %d", len(eventBQuests))
	}
}

func TestEventQuestComponent_ClearEventQuests(t *testing.T) {
	comp := NewEventQuestComponent(5)

	comp.AvailableQuests = append(comp.AvailableQuests,
		EventQuestDefinition{ID: "q1", EventID: "event_a"},
		EventQuestDefinition{ID: "q2", EventID: "event_b"},
	)
	comp.AcceptQuest("q1", time.Now().Add(time.Hour))

	// Add more available quests
	comp.AvailableQuests = append(comp.AvailableQuests,
		EventQuestDefinition{ID: "q3", EventID: "event_a"},
	)

	comp.ClearEventQuests("event_a")

	if len(comp.ActiveQuests) != 0 {
		t.Errorf("Should have 0 active quests for event_a, got %d", len(comp.ActiveQuests))
	}

	// q3 should be removed from available
	for _, q := range comp.AvailableQuests {
		if q.EventID == "event_a" {
			t.Error("Should not have available quests for event_a")
		}
	}

	// event_b should still have its quest
	if len(comp.GetAvailableQuestsForEvent("event_b")) != 1 {
		t.Error("event_b quest should remain")
	}

	// Expired quests should contain the cleared active quest
	if len(comp.ExpiredQuests) != 1 {
		t.Errorf("Should have 1 expired quest, got %d", len(comp.ExpiredQuests))
	}
}

func TestEventQuestComponent_Serialization(t *testing.T) {
	comp := createTestEventQuestComponent(t)
	comp.UpdateProgress("test_quest", 0, 3)

	data, err := comp.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	comp2 := NewEventQuestComponent(3)
	if err := comp2.Deserialize(data); err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if len(comp2.ActiveQuests) != 1 {
		t.Errorf("Deserialized ActiveQuests count = %d, want 1", len(comp2.ActiveQuests))
	}

	active := comp2.GetActiveQuest("test_quest")
	if active == nil {
		t.Fatal("Expected active quest after deserialize")
	}

	if active.Progress[0] != 3 {
		t.Errorf("Deserialized Progress[0] = %d, want 3", active.Progress[0])
	}
}

func TestGenerateEventQuests(t *testing.T) {
	tests := []struct {
		name  string
		theme EventTheme
		seed  int64
	}{
		{"spring event", EventThemeSpring, 12345},
		{"summer event", EventThemeSummer, 67890},
		{"autumn event", EventThemeAutumn, 11111},
		{"winter event", EventThemeWinter, 22222},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := EventInstance{
				Definition: EventDefinition{
					ID:    "test_" + string(tt.theme),
					Name:  "Test " + string(tt.theme) + " Festival",
					Theme: tt.theme,
					Seed:  tt.seed,
				},
			}

			quests := GenerateEventQuests(event, tt.seed)

			if len(quests) != 3 {
				t.Errorf("Expected 3 quests, got %d", len(quests))
			}

			// Check quest types
			hasCollection := false
			hasExploration := false
			hasBoss := false

			for _, q := range quests {
				switch q.QuestType {
				case EventQuestCollection:
					hasCollection = true
				case EventQuestExploration:
					hasExploration = true
				case EventQuestBoss:
					hasBoss = true
				}

				// Verify event ID is set
				if q.EventID != event.Definition.ID {
					t.Errorf("Quest EventID = %v, want %v", q.EventID, event.Definition.ID)
				}

				// Verify quest has objectives
				if len(q.Objectives) == 0 {
					t.Errorf("Quest %s has no objectives", q.ID)
				}
			}

			if !hasCollection {
				t.Error("Missing collection quest")
			}
			if !hasExploration {
				t.Error("Missing exploration quest")
			}
			if !hasBoss {
				t.Error("Missing boss quest")
			}
		})
	}
}

func TestGenerateEventQuests_Determinism(t *testing.T) {
	event := EventInstance{
		Definition: EventDefinition{
			ID:    "spring_festival",
			Name:  "Spring Festival",
			Theme: EventThemeSpring,
			Seed:  42,
		},
	}

	// Generate quests multiple times with same seed
	quests1 := GenerateEventQuests(event, 42)
	quests2 := GenerateEventQuests(event, 42)

	if len(quests1) != len(quests2) {
		t.Fatal("Quest counts differ between generations")
	}

	for i := range quests1 {
		if quests1[i].ID != quests2[i].ID {
			t.Errorf("Quest ID mismatch at %d: %s != %s", i, quests1[i].ID, quests2[i].ID)
		}
		if quests1[i].Name != quests2[i].Name {
			t.Errorf("Quest Name mismatch at %d: %s != %s", i, quests1[i].Name, quests2[i].Name)
		}
		if len(quests1[i].Objectives) != len(quests2[i].Objectives) {
			t.Errorf("Objectives count mismatch at %d", i)
		}
		if len(quests1[i].Objectives) > 0 && quests1[i].Objectives[0].Required != quests2[i].Objectives[0].Required {
			t.Errorf("Objective required mismatch at %d", i)
		}
	}
}

func TestGetThemeItems(t *testing.T) {
	themes := []EventTheme{EventThemeSpring, EventThemeSummer, EventThemeAutumn, EventThemeWinter}

	for _, theme := range themes {
		items := getThemeItems(theme)
		if len(items) < 2 {
			t.Errorf("Theme %s should have at least 2 items, got %d", theme, len(items))
		}
	}

	// Test unknown theme
	items := getThemeItems("unknown")
	if len(items) < 2 {
		t.Error("Unknown theme should have fallback items")
	}
}

func TestGetThemeLocations(t *testing.T) {
	themes := []EventTheme{EventThemeSpring, EventThemeSummer, EventThemeAutumn, EventThemeWinter}

	for _, theme := range themes {
		locations := getThemeLocations(theme)
		if len(locations) < 2 {
			t.Errorf("Theme %s should have at least 2 locations, got %d", theme, len(locations))
		}
	}
}

func TestGetThemeBosses(t *testing.T) {
	themes := []EventTheme{EventThemeSpring, EventThemeSummer, EventThemeAutumn, EventThemeWinter}

	for _, theme := range themes {
		bosses := getThemeBosses(theme)
		if len(bosses) < 2 {
			t.Errorf("Theme %s should have at least 2 bosses, got %d", theme, len(bosses))
		}
	}
}

// createTestEventQuestComponent creates a component with one active quest for testing.
func createTestEventQuestComponent(t *testing.T) *EventQuestComponent {
	t.Helper()

	comp := NewEventQuestComponent(3)
	comp.AvailableQuests = append(comp.AvailableQuests, EventQuestDefinition{
		ID:        "test_quest",
		EventID:   "test_event",
		Name:      "Test Quest",
		QuestType: EventQuestCollection,
		Objectives: []quest.Objective{
			{Description: "Collect items", Target: "item", Required: 5},
		},
		Reward: quest.Reward{XP: 100, Gold: 50},
	})

	comp.AcceptQuest("test_quest", time.Now().Add(24*time.Hour))
	return comp
}
