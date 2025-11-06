package engine

import (
	"testing"
	"time"
)

func TestNarrativeEventType_String(t *testing.T) {
	tests := []struct {
		name      string
		eventType NarrativeEventType
		want      string
	}{
		{"discovery", EventDiscovery, "Discovery"},
		{"conflict", EventConflict, "Conflict"},
		{"alliance", EventAlliance, "Alliance"},
		{"betrayal", EventBetrayal, "Betrayal"},
		{"revelation", EventRevelation, "Revelation"},
		{"sacrifice", EventSacrifice, "Sacrifice"},
		{"victory", EventVictory, "Victory"},
		{"defeat", EventDefeat, "Defeat"},
		{"unknown", NarrativeEventType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.eventType.String(); got != tt.want {
				t.Errorf("NarrativeEventType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStoryAct_String(t *testing.T) {
	tests := []struct {
		name string
		act  StoryAct
		want string
	}{
		{"setup", ActSetup, "Setup"},
		{"confrontation", ActConfrontation, "Confrontation"},
		{"resolution", ActResolution, "Resolution"},
		{"unknown", StoryAct(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.act.String(); got != tt.want {
				t.Errorf("StoryAct.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewNarrativeComponent(t *testing.T) {
	nc := NewNarrativeComponent()

	if nc.CurrentAct != ActSetup {
		t.Errorf("NewNarrativeComponent() CurrentAct = %v, want %v", nc.CurrentAct, ActSetup)
	}
	if nc.StoryProgress != 0.0 {
		t.Errorf("NewNarrativeComponent() StoryProgress = %v, want 0.0", nc.StoryProgress)
	}
	if len(nc.EventHistory) != 0 {
		t.Errorf("NewNarrativeComponent() EventHistory length = %v, want 0", len(nc.EventHistory))
	}
	if nc.WorldStateFlags == nil {
		t.Error("NewNarrativeComponent() WorldStateFlags is nil")
	}
	if nc.Relationships == nil {
		t.Error("NewNarrativeComponent() Relationships is nil")
	}
}

func TestNarrativeComponent_Type(t *testing.T) {
	nc := NewNarrativeComponent()
	if got := nc.Type(); got != "narrative" {
		t.Errorf("NarrativeComponent.Type() = %v, want narrative", got)
	}
}

func TestNarrativeComponent_AddEvent(t *testing.T) {
	nc := NewNarrativeComponent()

	event := NarrativeEvent{
		Type:        EventDiscovery,
		Description: "Found ancient ruins",
		LocationX:   100.0,
		LocationY:   200.0,
		Importance:  0.5,
	}

	nc.AddEvent(event)

	if len(nc.EventHistory) != 1 {
		t.Fatalf("AddEvent() EventHistory length = %v, want 1", len(nc.EventHistory))
	}

	stored := nc.EventHistory[0]
	if stored.Type != EventDiscovery {
		t.Errorf("AddEvent() stored event Type = %v, want %v", stored.Type, EventDiscovery)
	}
	if stored.Act != ActSetup {
		t.Errorf("AddEvent() stored event Act = %v, want %v", stored.Act, ActSetup)
	}
	if stored.Timestamp.IsZero() {
		t.Error("AddEvent() stored event Timestamp is zero")
	}

	// Check story progress increased
	if nc.StoryProgress <= 0.0 {
		t.Errorf("AddEvent() StoryProgress = %v, want > 0.0", nc.StoryProgress)
	}
}

func TestNarrativeComponent_AddEvent_ProgressCapped(t *testing.T) {
	nc := NewNarrativeComponent()

	// Add many high-importance events to exceed 1.0
	for i := 0; i < 20; i++ {
		nc.AddEvent(NarrativeEvent{
			Type:       EventVictory,
			Importance: 1.0,
		})
	}

	if nc.StoryProgress > 1.0 {
		t.Errorf("AddEvent() StoryProgress = %v, want <= 1.0", nc.StoryProgress)
	}
}

func TestNarrativeComponent_SetWorldFlag(t *testing.T) {
	nc := NewNarrativeComponent()

	nc.SetWorldFlag("defeated_dragon", true)
	nc.SetWorldFlag("saved_village", false)

	if !nc.WorldStateFlags["defeated_dragon"] {
		t.Error("SetWorldFlag() defeated_dragon should be true")
	}
	if nc.WorldStateFlags["saved_village"] {
		t.Error("SetWorldFlag() saved_village should be false")
	}
}

func TestNarrativeComponent_HasWorldFlag(t *testing.T) {
	nc := NewNarrativeComponent()

	nc.SetWorldFlag("quest_complete", true)

	tests := []struct {
		name string
		flag string
		want bool
	}{
		{"existing_true", "quest_complete", true},
		{"non_existing", "unknown_flag", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nc.HasWorldFlag(tt.flag); got != tt.want {
				t.Errorf("HasWorldFlag(%s) = %v, want %v", tt.flag, got, tt.want)
			}
		})
	}
}

func TestNarrativeComponent_ModifyRelationship(t *testing.T) {
	tests := []struct {
		name     string
		initial  float64
		delta    float64
		expected float64
	}{
		{"increase", 0.0, 0.5, 0.5},
		{"decrease", 0.5, -0.3, 0.2},
		{"clamp_upper", 0.9, 0.5, 1.0},
		{"clamp_lower", -0.9, -0.5, -1.0},
		{"neutral_to_hostile", 0.0, -0.8, -0.8},
		{"neutral_to_friendly", 0.0, 0.8, 0.8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nc := NewNarrativeComponent()
			nc.Relationships["faction_test"] = tt.initial

			nc.ModifyRelationship("faction_test", tt.delta)

			got := nc.GetRelationship("faction_test")
			if got != tt.expected {
				t.Errorf("ModifyRelationship() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNarrativeComponent_GetRelationship(t *testing.T) {
	nc := NewNarrativeComponent()

	// Test non-existent relationship (should return 0.0)
	if got := nc.GetRelationship("unknown"); got != 0.0 {
		t.Errorf("GetRelationship(unknown) = %v, want 0.0", got)
	}

	// Test existing relationship
	nc.Relationships["guild"] = 0.75
	if got := nc.GetRelationship("guild"); got != 0.75 {
		t.Errorf("GetRelationship(guild) = %v, want 0.75", got)
	}
}

func TestNarrativeComponent_AddDecision(t *testing.T) {
	nc := NewNarrativeComponent()

	decision := PlayerDecision{
		Description:         "Spared the enemy leader",
		AlternativeOutcomes: []string{"Execute", "Imprison"},
		ActualOutcome:       "Release",
		Consequences:        []string{"Enemy faction becomes ally"},
	}

	nc.AddDecision(decision)

	if len(nc.PlayerDecisions) != 1 {
		t.Fatalf("AddDecision() PlayerDecisions length = %v, want 1", len(nc.PlayerDecisions))
	}

	stored := nc.PlayerDecisions[0]
	if stored.Description != decision.Description {
		t.Errorf("AddDecision() stored Description = %v, want %v", stored.Description, decision.Description)
	}
	if stored.Timestamp.IsZero() {
		t.Error("AddDecision() stored Timestamp is zero")
	}
}

func TestNarrativeComponent_ProgressToNextAct(t *testing.T) {
	tests := []struct {
		name          string
		initialAct    StoryAct
		storyProgress float64
		wantAct       StoryAct
		wantErr       bool
	}{
		{"setup_to_confrontation_ready", ActSetup, 0.35, ActConfrontation, false},
		{"setup_to_confrontation_not_ready", ActSetup, 0.20, ActSetup, true},
		{"confrontation_to_resolution_ready", ActConfrontation, 0.70, ActResolution, false},
		{"confrontation_to_resolution_not_ready", ActConfrontation, 0.50, ActConfrontation, true},
		{"resolution_cannot_progress", ActResolution, 1.0, ActResolution, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nc := NewNarrativeComponent()
			nc.CurrentAct = tt.initialAct
			nc.StoryProgress = tt.storyProgress

			err := nc.ProgressToNextAct()

			if (err != nil) != tt.wantErr {
				t.Errorf("ProgressToNextAct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if nc.CurrentAct != tt.wantAct {
				t.Errorf("ProgressToNextAct() CurrentAct = %v, want %v", nc.CurrentAct, tt.wantAct)
			}
		})
	}
}

func TestNarrativeComponent_StoryThreads(t *testing.T) {
	nc := NewNarrativeComponent()

	// Add threads
	nc.AddStoryThread("Find the ancient artifact")
	nc.AddStoryThread("Defeat the dark lord")

	if len(nc.ActiveThreads) != 2 {
		t.Fatalf("AddStoryThread() ActiveThreads length = %v, want 2", len(nc.ActiveThreads))
	}

	// Resolve one thread
	nc.ResolveStoryThread("Find the ancient artifact")

	if len(nc.ActiveThreads) != 1 {
		t.Errorf("ResolveStoryThread() ActiveThreads length = %v, want 1", len(nc.ActiveThreads))
	}
	if len(nc.ResolvedThreads) != 1 {
		t.Errorf("ResolveStoryThread() ResolvedThreads length = %v, want 1", len(nc.ResolvedThreads))
	}
	if nc.ResolvedThreads[0] != "Find the ancient artifact" {
		t.Errorf("ResolveStoryThread() ResolvedThreads[0] = %v, want 'Find the ancient artifact'", nc.ResolvedThreads[0])
	}
}

func TestNarrativeComponent_GetEventsByType(t *testing.T) {
	nc := NewNarrativeComponent()

	// Add various events
	nc.AddEvent(NarrativeEvent{Type: EventDiscovery, Description: "Found ruins"})
	nc.AddEvent(NarrativeEvent{Type: EventConflict, Description: "Battle"})
	nc.AddEvent(NarrativeEvent{Type: EventDiscovery, Description: "Found treasure"})

	discoveries := nc.GetEventsByType(EventDiscovery)
	if len(discoveries) != 2 {
		t.Errorf("GetEventsByType(Discovery) length = %v, want 2", len(discoveries))
	}

	conflicts := nc.GetEventsByType(EventConflict)
	if len(conflicts) != 1 {
		t.Errorf("GetEventsByType(Conflict) length = %v, want 1", len(conflicts))
	}

	alliances := nc.GetEventsByType(EventAlliance)
	if len(alliances) != 0 {
		t.Errorf("GetEventsByType(Alliance) length = %v, want 0", len(alliances))
	}
}

func TestNarrativeComponent_GetEventsByAct(t *testing.T) {
	nc := NewNarrativeComponent()

	// Add events in Act 1
	nc.AddEvent(NarrativeEvent{Type: EventDiscovery, Description: "Start event"})

	// Progress to Act 2
	nc.StoryProgress = 0.5
	nc.ProgressToNextAct()

	// Add events in Act 2
	nc.AddEvent(NarrativeEvent{Type: EventConflict, Description: "Middle event"})
	nc.AddEvent(NarrativeEvent{Type: EventVictory, Description: "Another middle event"})

	act1Events := nc.GetEventsByAct(ActSetup)
	if len(act1Events) != 1 {
		t.Errorf("GetEventsByAct(Setup) length = %v, want 1", len(act1Events))
	}

	act2Events := nc.GetEventsByAct(ActConfrontation)
	if len(act2Events) != 2 {
		t.Errorf("GetEventsByAct(Confrontation) length = %v, want 2", len(act2Events))
	}

	act3Events := nc.GetEventsByAct(ActResolution)
	if len(act3Events) != 0 {
		t.Errorf("GetEventsByAct(Resolution) length = %v, want 0", len(act3Events))
	}
}

func TestNarrativeComponent_CheckTriggerConditions(t *testing.T) {
	nc := NewNarrativeComponent()

	// Add trigger conditions
	nc.TriggerConditions = []TriggerCondition{
		{
			TriggerType:    "quest_complete",
			RequiredFlags:  []string{"boss_defeated", "key_obtained"},
			EventToTrigger: NarrativeEvent{Type: EventVictory, Description: "Quest completed"},
		},
		{
			TriggerType:    "secret_found",
			RequiredFlags:  []string{"hidden_door_opened"},
			EventToTrigger: NarrativeEvent{Type: EventDiscovery, Description: "Secret area"},
		},
	}

	// Initially, no conditions met
	triggered := nc.CheckTriggerConditions()
	if len(triggered) != 0 {
		t.Errorf("CheckTriggerConditions() initial length = %v, want 0", len(triggered))
	}

	// Set one flag
	nc.SetWorldFlag("boss_defeated", true)
	triggered = nc.CheckTriggerConditions()
	if len(triggered) != 0 {
		t.Errorf("CheckTriggerConditions() partial flags length = %v, want 0", len(triggered))
	}

	// Set second flag to complete first condition
	nc.SetWorldFlag("key_obtained", true)
	triggered = nc.CheckTriggerConditions()
	if len(triggered) != 1 {
		t.Fatalf("CheckTriggerConditions() first complete length = %v, want 1", len(triggered))
	}
	if triggered[0].Type != EventVictory {
		t.Errorf("CheckTriggerConditions() triggered event type = %v, want Victory", triggered[0].Type)
	}

	// Condition should not trigger again
	triggered = nc.CheckTriggerConditions()
	if len(triggered) != 0 {
		t.Errorf("CheckTriggerConditions() second call length = %v, want 0 (already activated)", len(triggered))
	}

	// Activate second condition
	nc.SetWorldFlag("hidden_door_opened", true)
	triggered = nc.CheckTriggerConditions()
	if len(triggered) != 1 {
		t.Fatalf("CheckTriggerConditions() second condition length = %v, want 1", len(triggered))
	}
	if triggered[0].Type != EventDiscovery {
		t.Errorf("CheckTriggerConditions() triggered event type = %v, want Discovery", triggered[0].Type)
	}
}

func TestNarrativeEvent_Struct(t *testing.T) {
	// Test that all fields can be set
	event := NarrativeEvent{
		Type:             EventRevelation,
		Timestamp:        time.Now(),
		Description:      "Discovered the truth",
		InvolvedEntities: []int{1, 2, 3},
		LocationX:        100.0,
		LocationY:        200.0,
		Act:              ActConfrontation,
		Importance:       0.8,
		PlayerTriggered:  true,
		Consequences:     []string{"NPC turns hostile", "New quest available"},
	}

	if event.Type != EventRevelation {
		t.Errorf("NarrativeEvent Type = %v, want Revelation", event.Type)
	}
	if len(event.InvolvedEntities) != 3 {
		t.Errorf("NarrativeEvent InvolvedEntities length = %v, want 3", len(event.InvolvedEntities))
	}
	if len(event.Consequences) != 2 {
		t.Errorf("NarrativeEvent Consequences length = %v, want 2", len(event.Consequences))
	}
}
