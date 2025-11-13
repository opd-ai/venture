package engine

import (
	"testing"
	"time"
)

func TestMoralChoiceComponent_Type(t *testing.T) {
	comp := NewMoralChoiceComponent()
	if comp.Type() != "moral_choice" {
		t.Errorf("Expected type 'moral_choice', got '%s'", comp.Type())
	}
}

func TestNewMoralChoiceComponent(t *testing.T) {
	comp := NewMoralChoiceComponent()

	if comp == nil {
		t.Fatal("NewMoralChoiceComponent returned nil")
	}

	if comp.PendingChoices == nil {
		t.Error("PendingChoices not initialized")
	}

	if comp.ChoiceHistory == nil {
		t.Error("ChoiceHistory not initialized")
	}

	if comp.ActiveRedemptions == nil {
		t.Error("ActiveRedemptions not initialized")
	}

	if len(comp.PendingChoices) != 0 {
		t.Errorf("Expected 0 pending choices, got %d", len(comp.PendingChoices))
	}

	if len(comp.ChoiceHistory) != 0 {
		t.Errorf("Expected 0 history entries, got %d", len(comp.ChoiceHistory))
	}

	if len(comp.ActiveRedemptions) != 0 {
		t.Errorf("Expected 0 redemptions, got %d", len(comp.ActiveRedemptions))
	}
}

func TestMoralChoiceComponent_AddChoice(t *testing.T) {
	comp := NewMoralChoiceComponent()

	choice := MoralChoice{
		ID:          "choice1",
		Description: "Help or ignore?",
		Options: []ChoiceOption{
			{Label: "Help", Description: "Aid the villagers"},
			{Label: "Ignore", Description: "Walk away"},
		},
	}

	comp.AddChoice(choice)

	if len(comp.PendingChoices) != 1 {
		t.Fatalf("Expected 1 pending choice, got %d", len(comp.PendingChoices))
	}

	stored := comp.PendingChoices[0]
	if stored.ID != "choice1" {
		t.Errorf("Expected ID 'choice1', got '%s'", stored.ID)
	}

	// Check that TimeOffered was set
	if stored.TimeOffered.IsZero() {
		t.Error("TimeOffered should be set automatically")
	}
}

func TestMoralChoiceComponent_AddChoice_PreservesTimeOffered(t *testing.T) {
	comp := NewMoralChoiceComponent()

	customTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	choice := MoralChoice{
		ID:          "choice1",
		Description: "Test choice",
		TimeOffered: customTime,
	}

	comp.AddChoice(choice)

	if comp.PendingChoices[0].TimeOffered != customTime {
		t.Error("TimeOffered should be preserved if already set")
	}
}

func TestMoralChoiceComponent_GetPendingChoice(t *testing.T) {
	comp := NewMoralChoiceComponent()

	choice1 := MoralChoice{ID: "choice1", Description: "First"}
	choice2 := MoralChoice{ID: "choice2", Description: "Second"}

	comp.AddChoice(choice1)
	comp.AddChoice(choice2)

	// Test finding existing choice
	found := comp.GetPendingChoice("choice2")
	if found == nil {
		t.Fatal("Should find choice2")
	}
	if found.Description != "Second" {
		t.Errorf("Expected 'Second', got '%s'", found.Description)
	}

	// Test not finding non-existent choice
	notFound := comp.GetPendingChoice("choice99")
	if notFound != nil {
		t.Error("Should not find non-existent choice")
	}
}

func TestMoralChoiceComponent_RemovePendingChoice(t *testing.T) {
	comp := NewMoralChoiceComponent()

	comp.AddChoice(MoralChoice{ID: "choice1"})
	comp.AddChoice(MoralChoice{ID: "choice2"})
	comp.AddChoice(MoralChoice{ID: "choice3"})

	// Remove middle element
	removed := comp.RemovePendingChoice("choice2")
	if !removed {
		t.Error("Should successfully remove choice2")
	}

	if len(comp.PendingChoices) != 2 {
		t.Errorf("Expected 2 pending choices, got %d", len(comp.PendingChoices))
	}

	// Verify it's actually gone
	if comp.GetPendingChoice("choice2") != nil {
		t.Error("choice2 should be removed")
	}

	// Try removing non-existent
	notRemoved := comp.RemovePendingChoice("choice99")
	if notRemoved {
		t.Error("Should not successfully remove non-existent choice")
	}
}

func TestMoralChoiceComponent_RecordChoice(t *testing.T) {
	comp := NewMoralChoiceComponent()

	completed := CompletedChoice{
		ChoiceID:       "choice1",
		Description:    "Helped villagers",
		SelectedOption: 0,
		OptionLabel:    "Help",
		AlignmentChange: AlignmentDelta{
			LawDelta:  0.05,
			GoodDelta: 0.1,
		},
		ReputationChanges: map[string]float64{
			"Villagers": 10.0,
		},
	}

	comp.RecordChoice(completed)

	if len(comp.ChoiceHistory) != 1 {
		t.Fatalf("Expected 1 history entry, got %d", len(comp.ChoiceHistory))
	}

	stored := comp.ChoiceHistory[0]
	if stored.ChoiceID != "choice1" {
		t.Errorf("Expected ChoiceID 'choice1', got '%s'", stored.ChoiceID)
	}

	// Check that Timestamp was set
	if stored.Timestamp.IsZero() {
		t.Error("Timestamp should be set automatically")
	}
}

func TestMoralChoiceComponent_GetRecentChoices(t *testing.T) {
	comp := NewMoralChoiceComponent()

	// Test with empty history
	recent := comp.GetRecentChoices(5)
	if len(recent) != 0 {
		t.Errorf("Expected 0 recent choices, got %d", len(recent))
	}

	// Add some choices
	for i := 1; i <= 10; i++ {
		comp.RecordChoice(CompletedChoice{
			ChoiceID: string(rune('a' + i - 1)),
		})
		time.Sleep(time.Millisecond) // Ensure different timestamps
	}

	// Test getting last 3
	recent = comp.GetRecentChoices(3)
	if len(recent) != 3 {
		t.Fatalf("Expected 3 recent choices, got %d", len(recent))
	}

	// Should be last 3 in order (h, i, j)
	expectedIDs := []string{"h", "i", "j"}
	for i, choice := range recent {
		if choice.ChoiceID != expectedIDs[i] {
			t.Errorf("Choice %d: expected ID '%s', got '%s'", i, expectedIDs[i], choice.ChoiceID)
		}
	}

	// Test getting more than available
	recent = comp.GetRecentChoices(20)
	if len(recent) != 10 {
		t.Errorf("Expected 10 recent choices (all), got %d", len(recent))
	}

	// Test with zero/negative count
	recent = comp.GetRecentChoices(0)
	if len(recent) != 0 {
		t.Error("GetRecentChoices(0) should return empty slice")
	}

	recent = comp.GetRecentChoices(-5)
	if len(recent) != 0 {
		t.Error("GetRecentChoices(-5) should return empty slice")
	}
}

func TestMoralChoiceComponent_StartRedemption(t *testing.T) {
	comp := NewMoralChoiceComponent()

	arc := RedemptionArc{
		FactionName:        "Merchants",
		StartingReputation: -30.0,
		TargetReputation:   10.0,
		RequiredActions: []RedemptionAction{
			{Type: "Deliver", Description: "Deliver goods", Quantity: 5},
		},
	}

	comp.StartRedemption(arc)

	if len(comp.ActiveRedemptions) != 1 {
		t.Fatalf("Expected 1 redemption, got %d", len(comp.ActiveRedemptions))
	}

	stored := comp.ActiveRedemptions[0]
	if stored.FactionName != "Merchants" {
		t.Errorf("Expected faction 'Merchants', got '%s'", stored.FactionName)
	}

	// Check that StartTime was set
	if stored.StartTime.IsZero() {
		t.Error("StartTime should be set automatically")
	}
}

func TestMoralChoiceComponent_GetRedemptionArc(t *testing.T) {
	comp := NewMoralChoiceComponent()

	arc1 := RedemptionArc{FactionName: "Merchants"}
	arc2 := RedemptionArc{FactionName: "Guards"}

	comp.StartRedemption(arc1)
	comp.StartRedemption(arc2)

	// Test finding existing arc
	found := comp.GetRedemptionArc("Guards")
	if found == nil {
		t.Fatal("Should find Guards redemption")
	}
	if found.FactionName != "Guards" {
		t.Errorf("Expected 'Guards', got '%s'", found.FactionName)
	}

	// Test not finding non-existent arc
	notFound := comp.GetRedemptionArc("Bandits")
	if notFound != nil {
		t.Error("Should not find non-existent redemption")
	}
}

func TestMoralChoiceComponent_RemoveRedemptionArc(t *testing.T) {
	comp := NewMoralChoiceComponent()

	comp.StartRedemption(RedemptionArc{FactionName: "Merchants"})
	comp.StartRedemption(RedemptionArc{FactionName: "Guards"})
	comp.StartRedemption(RedemptionArc{FactionName: "Nobles"})

	// Remove middle element
	removed := comp.RemoveRedemptionArc("Guards")
	if !removed {
		t.Error("Should successfully remove Guards redemption")
	}

	if len(comp.ActiveRedemptions) != 2 {
		t.Errorf("Expected 2 redemptions, got %d", len(comp.ActiveRedemptions))
	}

	// Verify it's actually gone
	if comp.GetRedemptionArc("Guards") != nil {
		t.Error("Guards redemption should be removed")
	}

	// Try removing non-existent
	notRemoved := comp.RemoveRedemptionArc("Bandits")
	if notRemoved {
		t.Error("Should not successfully remove non-existent redemption")
	}
}

func TestMoralChoiceComponent_HasActiveMoralChoices(t *testing.T) {
	comp := NewMoralChoiceComponent()

	if comp.HasActiveMoralChoices() {
		t.Error("New component should have no active choices")
	}

	comp.AddChoice(MoralChoice{ID: "choice1"})

	if !comp.HasActiveMoralChoices() {
		t.Error("Component should have active choices")
	}
}

func TestMoralChoiceComponent_HasActiveRedemptions(t *testing.T) {
	comp := NewMoralChoiceComponent()

	if comp.HasActiveRedemptions() {
		t.Error("New component should have no active redemptions")
	}

	comp.StartRedemption(RedemptionArc{FactionName: "Test"})

	if !comp.HasActiveRedemptions() {
		t.Error("Component should have active redemptions")
	}
}

func TestRedemptionAction_IsComplete(t *testing.T) {
	tests := []struct {
		name     string
		action   RedemptionAction
		expected bool
	}{
		{
			name:     "Not complete",
			action:   RedemptionAction{Quantity: 5, Progress: 3},
			expected: false,
		},
		{
			name:     "Just complete",
			action:   RedemptionAction{Quantity: 5, Progress: 5},
			expected: true,
		},
		{
			name:     "Over complete",
			action:   RedemptionAction{Quantity: 5, Progress: 7},
			expected: true,
		},
		{
			name:     "Zero quantity",
			action:   RedemptionAction{Quantity: 0, Progress: 0},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.action.IsComplete()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRedemptionAction_GetProgress(t *testing.T) {
	tests := []struct {
		name     string
		action   RedemptionAction
		expected float64
	}{
		{
			name:     "Half complete",
			action:   RedemptionAction{Quantity: 10, Progress: 5},
			expected: 0.5,
		},
		{
			name:     "Fully complete",
			action:   RedemptionAction{Quantity: 10, Progress: 10},
			expected: 1.0,
		},
		{
			name:     "Over complete (clamped)",
			action:   RedemptionAction{Quantity: 10, Progress: 15},
			expected: 1.0,
		},
		{
			name:     "Zero quantity",
			action:   RedemptionAction{Quantity: 0, Progress: 0},
			expected: 1.0,
		},
		{
			name:     "Not started",
			action:   RedemptionAction{Quantity: 10, Progress: 0},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.action.GetProgress()
			if result != tt.expected {
				t.Errorf("Expected %.2f, got %.2f", tt.expected, result)
			}
		})
	}
}

func TestRedemptionArc_IsComplete(t *testing.T) {
	tests := []struct {
		name     string
		arc      RedemptionArc
		expected bool
	}{
		{
			name: "Not complete",
			arc: RedemptionArc{
				RequiredActions:  []RedemptionAction{{}, {}, {}},
				CompletedActions: 2,
			},
			expected: false,
		},
		{
			name: "Just complete",
			arc: RedemptionArc{
				RequiredActions:  []RedemptionAction{{}, {}, {}},
				CompletedActions: 3,
			},
			expected: true,
		},
		{
			name: "Over complete",
			arc: RedemptionArc{
				RequiredActions:  []RedemptionAction{{}, {}, {}},
				CompletedActions: 5,
			},
			expected: true,
		},
		{
			name: "No actions required",
			arc: RedemptionArc{
				RequiredActions:  []RedemptionAction{},
				CompletedActions: 0,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.arc.IsComplete()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRedemptionArc_GetProgress(t *testing.T) {
	tests := []struct {
		name     string
		arc      RedemptionArc
		expected float64
		delta    float64 // Allow small floating point differences
	}{
		{
			name: "Partial progress",
			arc: RedemptionArc{
				RequiredActions: []RedemptionAction{
					{Quantity: 10, Progress: 5},  // 0.5
					{Quantity: 10, Progress: 10}, // 1.0
					{Quantity: 10, Progress: 0},  // 0.0
				},
			},
			expected: 0.5, // (0.5 + 1.0 + 0.0) / 3
			delta:    0.01,
		},
		{
			name: "All complete",
			arc: RedemptionArc{
				RequiredActions: []RedemptionAction{
					{Quantity: 5, Progress: 5},
					{Quantity: 3, Progress: 3},
				},
			},
			expected: 1.0,
			delta:    0.01,
		},
		{
			name: "No actions",
			arc: RedemptionArc{
				RequiredActions: []RedemptionAction{},
			},
			expected: 1.0,
			delta:    0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.arc.GetProgress()
			diff := result - tt.expected
			if diff < 0 {
				diff = -diff
			}
			if diff > tt.delta {
				t.Errorf("Expected %.2f (±%.2f), got %.2f", tt.expected, tt.delta, result)
			}
		})
	}
}

func TestRedemptionArc_IsExpired(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		timeLimit time.Time
		expected  bool
	}{
		{
			name:      "No time limit",
			timeLimit: time.Time{},
			expected:  false,
		},
		{
			name:      "Future deadline",
			timeLimit: now.Add(1 * time.Hour),
			expected:  false,
		},
		{
			name:      "Past deadline",
			timeLimit: now.Add(-1 * time.Hour),
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arc := RedemptionArc{TimeLimit: tt.timeLimit}
			result := arc.IsExpired()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMoralChoice_IsExpired(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		expiresAt time.Time
		expected  bool
	}{
		{
			name:      "No expiry",
			expiresAt: time.Time{},
			expected:  false,
		},
		{
			name:      "Future expiry",
			expiresAt: now.Add(1 * time.Hour),
			expected:  false,
		},
		{
			name:      "Past expiry",
			expiresAt: now.Add(-1 * time.Hour),
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			choice := MoralChoice{ExpiresAt: tt.expiresAt}
			result := choice.IsExpired()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestChoiceOption_ComplexStructure(t *testing.T) {
	// Test that complex option structures work correctly
	option := ChoiceOption{
		Label:       "Heroic Sacrifice",
		Description: "Sacrifice yourself to save the village",
		AlignmentImpact: AlignmentDelta{
			LawDelta:  0.2,
			GoodDelta: 0.5,
		},
		ReputationImpact: map[string]float64{
			"Villagers": 50.0,
			"Bandits":   -30.0,
		},
		Rewards: &ChoiceRewards{
			XP:          500,
			Gold:        0,
			Items:       []string{"Memorial_Statue"},
			UnlockQuest: "quest_legendary_hero",
		},
		Consequences: &ChoiceConsequences{
			HostileFactions: []string{"Bandits", "Thieves_Guild"},
			SpawnEnemies:    10,
		},
	}

	if option.Label != "Heroic Sacrifice" {
		t.Errorf("Label not preserved")
	}

	if option.AlignmentImpact.GoodDelta != 0.5 {
		t.Errorf("Expected GoodDelta 0.5, got %.2f", option.AlignmentImpact.GoodDelta)
	}

	if option.ReputationImpact["Villagers"] != 50.0 {
		t.Errorf("Expected Villagers reputation 50.0, got %.2f", option.ReputationImpact["Villagers"])
	}

	if option.Rewards.XP != 500 {
		t.Errorf("Expected XP 500, got %d", option.Rewards.XP)
	}

	if len(option.Consequences.HostileFactions) != 2 {
		t.Errorf("Expected 2 hostile factions, got %d", len(option.Consequences.HostileFactions))
	}
}
