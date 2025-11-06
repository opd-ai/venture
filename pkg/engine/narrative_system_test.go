package engine

import (
	"testing"
)

// TestNarrativeSystem_Creation tests system initialization
func TestNarrativeSystem_Creation(t *testing.T) {
	// Create without world (for unit testing)
	ns := NewNarrativeSystem(nil)
	if ns == nil {
		t.Fatal("NewNarrativeSystem() returned nil")
	}
}

// TestNarrativeSystem_TriggerEvent tests manual event triggering
func TestNarrativeSystem_TriggerEvent(t *testing.T) {
	ns := NewNarrativeSystem(nil)
	narrative := NewNarrativeComponent()

	initialCount := len(narrative.EventHistory)

	ns.TriggerEvent(narrative, EventDiscovery, "Test discovery", 0.5, []int{1, 2}, 100.0, 200.0)

	if len(narrative.EventHistory) != initialCount+1 {
		t.Errorf("TriggerEvent() EventHistory length = %v, want %v", len(narrative.EventHistory), initialCount+1)
	}

	event := narrative.EventHistory[len(narrative.EventHistory)-1]
	if event.Type != EventDiscovery {
		t.Errorf("TriggerEvent() event Type = %v, want Discovery", event.Type)
	}
	if event.Description != "Test discovery" {
		t.Errorf("TriggerEvent() event Description = %v, want 'Test discovery'", event.Description)
	}
	if event.Importance != 0.5 {
		t.Errorf("TriggerEvent() event Importance = %v, want 0.5", event.Importance)
	}
}

// TestNarrativeSystem_OnCombatVictory tests combat victory events
func TestNarrativeSystem_OnCombatVictory(t *testing.T) {
	ns := NewNarrativeSystem(nil)
	narrative := NewNarrativeComponent()

	ns.OnCombatVictory(narrative, nil, 150.0, 250.0)

	if len(narrative.EventHistory) != 1 {
		t.Fatalf("OnCombatVictory() EventHistory length = %v, want 1", len(narrative.EventHistory))
	}

	event := narrative.EventHistory[0]
	if event.Type != EventVictory {
		t.Errorf("OnCombatVictory() event Type = %v, want Victory", event.Type)
	}
	if event.LocationX != 150.0 {
		t.Errorf("OnCombatVictory() event LocationX = %v, want 150.0", event.LocationX)
	}
	if event.LocationY != 250.0 {
		t.Errorf("OnCombatVictory() event LocationY = %v, want 250.0", event.LocationY)
	}
}

// TestNarrativeSystem_OnDiscovery tests discovery events
func TestNarrativeSystem_OnDiscovery(t *testing.T) {
	ns := NewNarrativeSystem(nil)
	narrative := NewNarrativeComponent()

	ns.OnDiscovery(narrative, "ancient ruins", 100.0, 100.0)

	if len(narrative.EventHistory) != 1 {
		t.Fatalf("OnDiscovery() EventHistory length = %v, want 1", len(narrative.EventHistory))
	}

	event := narrative.EventHistory[0]
	if event.Type != EventDiscovery {
		t.Errorf("OnDiscovery() event Type = %v, want Discovery", event.Type)
	}
	if event.Description != "Discovered ancient ruins" {
		t.Errorf("OnDiscovery() event Description = %v, want 'Discovered ancient ruins'", event.Description)
	}
}

// TestNarrativeSystem_OnDialogueComplete tests dialogue completion events
func TestNarrativeSystem_OnDialogueComplete(t *testing.T) {
	ns := NewNarrativeSystem(nil)
	narrative := NewNarrativeComponent()

	ns.OnDialogueComplete(narrative, nil, 0)

	if len(narrative.EventHistory) != 1 {
		t.Fatalf("OnDialogueComplete() EventHistory length = %v, want 1", len(narrative.EventHistory))
	}

	event := narrative.EventHistory[0]
	if event.Type != EventAlliance {
		t.Errorf("OnDialogueComplete() event Type = %v, want Alliance", event.Type)
	}
}

// TestNarrativeSystem_OnPuzzleSolved tests puzzle solved events
func TestNarrativeSystem_OnPuzzleSolved(t *testing.T) {
	ns := NewNarrativeSystem(nil)
	narrative := NewNarrativeComponent()

	ns.OnPuzzleSolved(narrative, "lever", 200.0, 300.0)

	if len(narrative.EventHistory) != 1 {
		t.Fatalf("OnPuzzleSolved() EventHistory length = %v, want 1", len(narrative.EventHistory))
	}

	event := narrative.EventHistory[0]
	if event.Type != EventDiscovery {
		t.Errorf("OnPuzzleSolved() event Type = %v, want Discovery", event.Type)
	}
	if event.Description != "Solved lever puzzle" {
		t.Errorf("OnPuzzleSolved() event Description = %v, want 'Solved lever puzzle'", event.Description)
	}
}

// TestNarrativeSystem_OnPlayerDeath tests player death events
func TestNarrativeSystem_OnPlayerDeath(t *testing.T) {
	ns := NewNarrativeSystem(nil)
	narrative := NewNarrativeComponent()

	ns.OnPlayerDeath(narrative, 50.0, 75.0)

	if len(narrative.EventHistory) != 1 {
		t.Fatalf("OnPlayerDeath() EventHistory length = %v, want 1", len(narrative.EventHistory))
	}

	event := narrative.EventHistory[0]
	if event.Type != EventDefeat {
		t.Errorf("OnPlayerDeath() event Type = %v, want Defeat", event.Type)
	}
	if event.Importance != 0.6 {
		t.Errorf("OnPlayerDeath() event Importance = %v, want 0.6", event.Importance)
	}
}

// TestNarrativeSystem_OnQuestComplete tests quest completion events
func TestNarrativeSystem_OnQuestComplete(t *testing.T) {
	ns := NewNarrativeSystem(nil)
	narrative := NewNarrativeComponent()

	ns.OnQuestComplete(narrative, "Save the Village")

	if len(narrative.EventHistory) != 1 {
		t.Fatalf("OnQuestComplete() EventHistory length = %v, want 1", len(narrative.EventHistory))
	}

	event := narrative.EventHistory[0]
	if event.Type != EventVictory {
		t.Errorf("OnQuestComplete() event Type = %v, want Victory", event.Type)
	}
	if event.Description != "Completed quest: Save the Village" {
		t.Errorf("OnQuestComplete() event Description = %v, want 'Completed quest: Save the Village'", event.Description)
	}
}

// TestNarrativeSystem_GetStoryStatus tests story status retrieval
func TestNarrativeSystem_GetStoryStatus(t *testing.T) {
	ns := NewNarrativeSystem(nil)
	narrative := NewNarrativeComponent()
	narrative.MainObjective = "Test Objective"
	narrative.AddStoryThread("Thread 1")
	narrative.AddStoryThread("Thread 2")

	ns.TriggerEvent(narrative, EventDiscovery, "Event 1", 0.3, []int{}, 0, 0)
	ns.TriggerEvent(narrative, EventConflict, "Event 2", 0.4, []int{}, 0, 0)

	status := ns.GetStoryStatus(narrative)

	if status.CurrentAct != ActSetup {
		t.Errorf("GetStoryStatus() CurrentAct = %v, want Setup", status.CurrentAct)
	}
	if status.MainObjective != "Test Objective" {
		t.Errorf("GetStoryStatus() MainObjective = %v, want 'Test Objective'", status.MainObjective)
	}
	if status.ActiveThreads != 2 {
		t.Errorf("GetStoryStatus() ActiveThreads = %v, want 2", status.ActiveThreads)
	}
	if status.EventCount != 2 {
		t.Errorf("GetStoryStatus() EventCount = %v, want 2", status.EventCount)
	}
}

// TestNarrativeSystem_SetMainObjective tests objective setting
func TestNarrativeSystem_SetMainObjective(t *testing.T) {
	ns := NewNarrativeSystem(nil)
	narrative := NewNarrativeComponent()

	ns.SetMainObjective(narrative, "Find the ancient artifact")

	if narrative.MainObjective != "Find the ancient artifact" {
		t.Errorf("SetMainObjective() MainObjective = %v, want 'Find the ancient artifact'", narrative.MainObjective)
	}
}

// TestNarrativeSystem_AddStoryTrigger tests trigger addition
func TestNarrativeSystem_AddStoryTrigger(t *testing.T) {
	ns := NewNarrativeSystem(nil)
	narrative := NewNarrativeComponent()

	event := NarrativeEvent{
		Type:        EventRevelation,
		Description: "Secret revealed",
		Importance:  0.8,
	}

	ns.AddStoryTrigger(narrative, "exploration", []string{"found_key", "opened_door"}, event)

	if len(narrative.TriggerConditions) != 1 {
		t.Fatalf("AddStoryTrigger() TriggerConditions length = %v, want 1", len(narrative.TriggerConditions))
	}

	trigger := narrative.TriggerConditions[0]
	if trigger.TriggerType != "exploration" {
		t.Errorf("AddStoryTrigger() TriggerType = %v, want 'exploration'", trigger.TriggerType)
	}
	if len(trigger.RequiredFlags) != 2 {
		t.Errorf("AddStoryTrigger() RequiredFlags length = %v, want 2", len(trigger.RequiredFlags))
	}
}

// TestNarrativeSystem_WorldFlags tests world flag operations
func TestNarrativeSystem_WorldFlags(t *testing.T) {
	ns := NewNarrativeSystem(nil)
	narrative := NewNarrativeComponent()

	// Initially false
	if ns.CheckWorldFlag(narrative, "test_flag") {
		t.Error("CheckWorldFlag() should return false for unset flag")
	}

	// Set to true
	ns.SetWorldFlag(narrative, "test_flag", true)
	if !ns.CheckWorldFlag(narrative, "test_flag") {
		t.Error("CheckWorldFlag() should return true after setting flag")
	}

	// Set to false
	ns.SetWorldFlag(narrative, "test_flag", false)
	if ns.CheckWorldFlag(narrative, "test_flag") {
		t.Error("CheckWorldFlag() should return false after unsetting flag")
	}
}

// TestNarrativeSystem_GetRelationshipStatus tests relationship status text
func TestNarrativeSystem_GetRelationshipStatus(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		expected string
	}{
		{"allied", 0.9, "Allied"},
		{"friendly", 0.5, "Friendly"},
		{"neutral", 0.0, "Neutral"},
		{"hostile", -0.5, "Hostile"},
		{"enemy", -0.9, "Enemy"},
		{"edge_allied", 0.75, "Allied"},
		{"edge_friendly", 0.25, "Friendly"},
		{"edge_hostile", -0.75, "Enemy"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := NewNarrativeSystem(nil)
			narrative := NewNarrativeComponent()
			narrative.Relationships["faction"] = tt.value

			status := ns.GetRelationshipStatus(narrative, "faction")
			if status != tt.expected {
				t.Errorf("GetRelationshipStatus() = %v, want %v", status, tt.expected)
			}
		})
	}
}

// TestNarrativeSystem_RecordPlayerDecision tests decision recording
func TestNarrativeSystem_RecordPlayerDecision(t *testing.T) {
	ns := NewNarrativeSystem(nil)
	narrative := NewNarrativeComponent()

	alternatives := []string{"Option A", "Option B", "Option C"}
	consequences := []string{"Result 1", "Result 2"}

	ns.RecordPlayerDecision(narrative, "Major choice", "Option B", alternatives, consequences)

	if len(narrative.PlayerDecisions) != 1 {
		t.Fatalf("RecordPlayerDecision() PlayerDecisions length = %v, want 1", len(narrative.PlayerDecisions))
	}

	decision := narrative.PlayerDecisions[0]
	if decision.Description != "Major choice" {
		t.Errorf("RecordPlayerDecision() Description = %v, want 'Major choice'", decision.Description)
	}
	if decision.ActualOutcome != "Option B" {
		t.Errorf("RecordPlayerDecision() ActualOutcome = %v, want 'Option B'", decision.ActualOutcome)
	}
	if len(decision.AlternativeOutcomes) != 3 {
		t.Errorf("RecordPlayerDecision() AlternativeOutcomes length = %v, want 3", len(decision.AlternativeOutcomes))
	}
	if len(decision.Consequences) != 2 {
		t.Errorf("RecordPlayerDecision() Consequences length = %v, want 2", len(decision.Consequences))
	}
}

// TestNarrativeSystem_shouldProgressAct tests act progression logic
func TestNarrativeSystem_shouldProgressAct(t *testing.T) {
	tests := []struct {
		name     string
		act      StoryAct
		progress float64
		expected bool
	}{
		{"setup_ready", ActSetup, 0.35, true},
		{"setup_not_ready", ActSetup, 0.25, false},
		{"confrontation_ready", ActConfrontation, 0.70, true},
		{"confrontation_not_ready", ActConfrontation, 0.50, false},
		{"resolution_cannot_progress", ActResolution, 1.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := NewNarrativeSystem(nil)
			narrative := NewNarrativeComponent()
			narrative.CurrentAct = tt.act
			narrative.StoryProgress = tt.progress

			result := ns.shouldProgressAct(narrative)
			if result != tt.expected {
				t.Errorf("shouldProgressAct() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestNarrativeSystem_getRecentEvents tests recent events retrieval
func TestNarrativeSystem_getRecentEvents(t *testing.T) {
	ns := NewNarrativeSystem(nil)
	narrative := NewNarrativeComponent()

	// Add 10 events
	for i := 0; i < 10; i++ {
		narrative.AddEvent(NarrativeEvent{
			Type:        EventDiscovery,
			Description: "",
		})
	}

	// Get last 5 events
	recent := ns.getRecentEvents(narrative, 5)
	if len(recent) != 5 {
		t.Errorf("getRecentEvents() length = %v, want 5", len(recent))
	}

	// Get more events than exist
	all := ns.getRecentEvents(narrative, 20)
	if len(all) != 10 {
		t.Errorf("getRecentEvents() length = %v, want 10", len(all))
	}

	// Empty history
	emptyNarrative := NewNarrativeComponent()
	empty := ns.getRecentEvents(emptyNarrative, 5)
	if len(empty) != 0 {
		t.Errorf("getRecentEvents() on empty history length = %v, want 0", len(empty))
	}
}
