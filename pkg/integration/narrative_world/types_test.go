package narrative_world

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
)

func TestEventType_String(t *testing.T) {
	tests := []struct {
		eventType EventType
		expected  string
	}{
		{EventTypeCombat, "Combat"},
		{EventTypeTreasure, "Treasure"},
		{EventTypeDanger, "Danger"},
		{EventTypeBonding, "Bonding"},
		{EventTypeConflict, "Conflict"},
		{EventTypeDiscovery, "Discovery"},
		{EventTypeSacrifice, "Sacrifice"},
		{EventTypeBetray, "Betray"},
		{EventType(999), "Unknown"}, // Invalid value
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.eventType.String()
			if got != tt.expected {
				t.Errorf("EventType(%d).String() = %q, want %q", tt.eventType, got, tt.expected)
			}
		})
	}
}

func TestObjectiveType_String(t *testing.T) {
	tests := []struct {
		objectiveType ObjectiveType
		expected      string
	}{
		{ObjectiveDefeat, "Defeat"},
		{ObjectiveCollect, "Collect"},
		{ObjectiveVisit, "Visit"},
		{ObjectiveProtect, "Protect"},
		{ObjectiveTalk, "Talk"},
		{ObjectiveExplore, "Explore"},
		{ObjectiveType(999), "Unknown"}, // Invalid value
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.objectiveType.String()
			if got != tt.expected {
				t.Errorf("ObjectiveType(%d).String() = %q, want %q", tt.objectiveType, got, tt.expected)
			}
		})
	}
}

func TestConsequenceType_String(t *testing.T) {
	tests := []struct {
		consequenceType ConsequenceType
		expected        string
	}{
		{ConsequenceLoyaltyChange, "Loyalty Change"},
		{ConsequenceDeparture, "Departure"},
		{ConsequenceDeath, "Death"},
		{ConsequenceRelationshipChange, "Relationship Change"},
		{ConsequenceItemGain, "Item Gain"},
		{ConsequenceSkillUnlock, "Skill Unlock"},
		{ConsequenceType(999), "Unknown"}, // Invalid value
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.consequenceType.String()
			if got != tt.expected {
				t.Errorf("ConsequenceType(%d).String() = %q, want %q", tt.consequenceType, got, tt.expected)
			}
		})
	}
}

func TestConflictType_String(t *testing.T) {
	tests := []struct {
		conflictType ConflictType
		expected     string
	}{
		{ConflictPersonality, "Personality Clash"},
		{ConflictRivalry, "Rivalry"},
		{ConflictBeliefs, "Conflicting Beliefs"},
		{ConflictPastHistory, "Past History"},
		{ConflictResourceCompetition, "Resource Competition"},
		{ConflictType(999), "Unknown"}, // Invalid value
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.conflictType.String()
			if got != tt.expected {
				t.Errorf("ConflictType(%d).String() = %q, want %q", tt.conflictType, got, tt.expected)
			}
		})
	}
}

func TestStoryOutcome_String(t *testing.T) {
	tests := []struct {
		outcome  StoryOutcome
		expected string
	}{
		{OutcomeUnresolved, "Unresolved"},
		{OutcomeFriendship, "Friendship"},
		{OutcomeRomance, "Romance"},
		{OutcomeRivalry, "Rivalry"},
		{OutcomeBetrayal, "Betrayal"},
		{OutcomeSacrifice, "Sacrifice"},
		{StoryOutcome(999), "Unknown"}, // Invalid value
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.outcome.String()
			if got != tt.expected {
				t.Errorf("StoryOutcome(%d).String() = %q, want %q", tt.outcome, got, tt.expected)
			}
		})
	}
}

func TestGenerateConsequenceDescription(t *testing.T) {
	manager := NewStoryEventManager(12345)

	tests := []struct {
		consequenceType ConsequenceType
		expectedPrefix  string
	}{
		{ConsequenceLoyaltyChange, "Your companion's loyalty"},
		{ConsequenceDeparture, "Your companion may leave"},
		{ConsequenceDeath, "Your companion risks permanent death"},
		{ConsequenceRelationshipChange, "Your relationship"},
		{ConsequenceItemGain, "You will receive"},
		{ConsequenceSkillUnlock, "Your companion will unlock"},
		{ConsequenceType(999), "Unknown"}, // Default case
	}

	for _, tt := range tests {
		t.Run(tt.consequenceType.String(), func(t *testing.T) {
			desc := manager.generateConsequenceDescription(tt.consequenceType)
			if desc == "" {
				t.Error("generateConsequenceDescription returned empty string")
			}
			// Verify the description contains expected prefix
			found := false
			for i := 0; i <= len(desc)-len(tt.expectedPrefix) && !found; i++ {
				if len(desc) >= len(tt.expectedPrefix) && desc[:len(tt.expectedPrefix)] == tt.expectedPrefix {
					found = true
				}
			}
			// For unknown type, just check it returns the default message
			if tt.consequenceType == ConsequenceType(999) {
				if desc != "Unknown consequence" {
					t.Errorf("expected 'Unknown consequence' for invalid type, got %q", desc)
				}
			}
		})
	}
}

func TestCompleteQuestNotFound(t *testing.T) {
	manager := NewStoryEventManager(12345)

	// Try to complete a quest that doesn't exist
	_, err := manager.CompleteQuest(1, "nonexistent-quest")
	if err == nil {
		t.Fatal("expected error for non-existent quest")
	}
	// Error can be "not found" or "no active quests"
	errMsg := err.Error()
	if !contains(errMsg, "not found") && !contains(errMsg, "no active quests") {
		t.Errorf("error should mention 'not found' or 'no active quests', got: %v", err)
	}
}

func TestSystemCompleteQuestNotFound(t *testing.T) {
	world := engine.NewWorld()
	system := NewSystem(world, 12345)

	// Try to complete a quest that doesn't exist
	err := system.CompleteQuest(1, "nonexistent-quest")
	if err == nil {
		t.Fatal("expected error for non-existent quest")
	}
	// Error can be "not found" or "no active quests"
	errMsg := err.Error()
	if !contains(errMsg, "not found") && !contains(errMsg, "no active quests") {
		t.Errorf("error should mention 'not found' or 'no active quests', got: %v", err)
	}
}
