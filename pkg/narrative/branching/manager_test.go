package branching

import (
	"fmt"
	"math"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestManagerStartArc(t *testing.T) {
	manager := NewManager()
	gen := NewGenerator()

	result, err := gen.Generate(12345, procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	})
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	arc := result.(*StoryArc)
	manager.RegisterArc(arc)

	progress, err := manager.StartArc("player1", arc.ID)
	if err != nil {
		t.Fatalf("StartArc() failed: %v", err)
	}

	if progress.ArcID != arc.ID {
		t.Errorf("Progress arc ID = %s, want %s", progress.ArcID, arc.ID)
	}

	if progress.CurrentNodeID != arc.StartNodeID {
		t.Errorf("Current node = %s, want %s", progress.CurrentNodeID, arc.StartNodeID)
	}

	if len(progress.VisitedNodes) != 1 {
		t.Errorf("Visited nodes count = %d, want 1", len(progress.VisitedNodes))
	}

	if progress.Completed {
		t.Error("Progress should not be completed")
	}

	// Check alignment initialized to neutral
	if progress.Alignment[AlignmentGoodEvil] != 0.0 {
		t.Errorf("Good/Evil alignment = %f, want 0.0", progress.Alignment[AlignmentGoodEvil])
	}
}

func TestManagerStartArcNonExistent(t *testing.T) {
	manager := NewManager()

	_, err := manager.StartArc("player1", "nonexistent")
	if err == nil {
		t.Error("StartArc() should fail for nonexistent arc")
	}
}

func TestManagerMakeChoice(t *testing.T) {
	manager := NewManager()
	gen := NewGenerator()

	result, _ := gen.Generate(12345, procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	})

	arc := result.(*StoryArc)
	manager.RegisterArc(arc)
	manager.StartArc("player1", arc.ID)

	// Find a choice node
	var choiceNode *StoryNode
	for _, node := range arc.Nodes {
		if node.Type == NodeTypeChoice && len(node.Choices) > 0 {
			choiceNode = node
			break
		}
	}

	if choiceNode == nil {
		t.Skip("No choice nodes in generated arc")
	}

	// Manually set current node to choice node for testing
	progress, _ := manager.GetProgress("player1", arc.ID)
	progress.CurrentNodeID = choiceNode.ID

	choice := choiceNode.Choices[0]
	err := manager.MakeChoice("player1", arc.ID, choice.ID)
	if err != nil {
		t.Fatalf("MakeChoice() failed: %v", err)
	}

	// Verify choice was recorded
	updatedProgress, _ := manager.GetProgress("player1", arc.ID)
	if updatedProgress.ChoicesMade[choiceNode.ID] != choice.ID {
		t.Error("Choice was not recorded")
	}

	// Verify moved to next node
	if updatedProgress.CurrentNodeID != choice.NextNodeID {
		t.Errorf("Current node = %s, want %s", updatedProgress.CurrentNodeID, choice.NextNodeID)
	}
}

func TestManagerMakeChoiceInvalidNode(t *testing.T) {
	manager := NewManager()
	gen := NewGenerator()

	result, _ := gen.Generate(12345, procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	})

	arc := result.(*StoryArc)
	manager.RegisterArc(arc)
	manager.StartArc("player1", arc.ID)

	// Try to make a choice on a non-choice node
	err := manager.MakeChoice("player1", arc.ID, "invalid_choice")
	if err == nil {
		t.Error("MakeChoice() should fail on non-choice node")
	}
}

func TestManagerAdvanceStory(t *testing.T) {
	manager := NewManager()
	gen := NewGenerator()

	result, _ := gen.Generate(12345, procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	})

	arc := result.(*StoryArc)
	manager.RegisterArc(arc)
	manager.StartArc("player1", arc.ID)

	// Find an event node with a next node
	var eventNode *StoryNode
	for _, node := range arc.Nodes {
		if node.Type == NodeTypeEvent && node.NextNodeID != "" {
			eventNode = node
			break
		}
	}

	if eventNode == nil {
		t.Skip("No event nodes with next node in generated arc")
	}

	// Set current node to event node
	progress, _ := manager.GetProgress("player1", arc.ID)
	progress.CurrentNodeID = eventNode.ID

	err := manager.AdvanceStory("player1", arc.ID)
	if err != nil {
		t.Fatalf("AdvanceStory() failed: %v", err)
	}

	// Verify moved to next node
	updatedProgress, _ := manager.GetProgress("player1", arc.ID)
	if updatedProgress.CurrentNodeID != eventNode.NextNodeID {
		t.Errorf("Current node = %s, want %s", updatedProgress.CurrentNodeID, eventNode.NextNodeID)
	}
}

func TestManagerGetAlignment(t *testing.T) {
	manager := NewManager()
	gen := NewGenerator()

	result, _ := gen.Generate(12345, procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	})

	arc := result.(*StoryArc)
	manager.RegisterArc(arc)
	manager.StartArc("player1", arc.ID)

	alignment, err := manager.GetAlignment("player1", arc.ID)
	if err != nil {
		t.Fatalf("GetAlignment() failed: %v", err)
	}

	// Should be neutral initially
	if alignment[AlignmentGoodEvil] != 0.0 {
		t.Errorf("Good/Evil = %f, want 0.0", alignment[AlignmentGoodEvil])
	}

	if alignment[AlignmentLawChaos] != 0.0 {
		t.Errorf("Law/Chaos = %f, want 0.0", alignment[AlignmentLawChaos])
	}

	if alignment[AlignmentHonorDishonor] != 0.0 {
		t.Errorf("Honor/Dishonor = %f, want 0.0", alignment[AlignmentHonorDishonor])
	}
}

func TestManagerGetFactionReputation(t *testing.T) {
	manager := NewManager()
	gen := NewGenerator()

	result, _ := gen.Generate(12345, procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	})

	arc := result.(*StoryArc)
	manager.RegisterArc(arc)
	manager.StartArc("player1", arc.ID)

	reputation, err := manager.GetFactionReputation("player1", arc.ID)
	if err != nil {
		t.Fatalf("GetFactionReputation() failed: %v", err)
	}

	// Should be empty initially
	if len(reputation) != 0 {
		t.Errorf("Reputation count = %d, want 0", len(reputation))
	}
}

func TestManagerGetCurrentNode(t *testing.T) {
	manager := NewManager()
	gen := NewGenerator()

	result, _ := gen.Generate(12345, procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	})

	arc := result.(*StoryArc)
	manager.RegisterArc(arc)
	manager.StartArc("player1", arc.ID)

	node, err := manager.GetCurrentNode("player1", arc.ID)
	if err != nil {
		t.Fatalf("GetCurrentNode() failed: %v", err)
	}

	if node.ID != arc.StartNodeID {
		t.Errorf("Node ID = %s, want %s", node.ID, arc.StartNodeID)
	}

	if node.Type != NodeTypeStart {
		t.Errorf("Node type = %s, want Start", node.Type)
	}
}

func TestManagerCheckRequirements(t *testing.T) {
	manager := NewManager()

	progress := &PlayerProgress{
		Variables: map[string]interface{}{
			"gold":   100,
			"level":  5,
			"hasKey": true,
		},
	}

	tests := []struct {
		name         string
		requirements map[string]interface{}
		wantErr      bool
	}{
		{
			name:         "no requirements",
			requirements: map[string]interface{}{},
			wantErr:      false,
		},
		{
			name: "met int requirement",
			requirements: map[string]interface{}{
				"gold": 50,
			},
			wantErr: false,
		},
		{
			name: "unmet int requirement",
			requirements: map[string]interface{}{
				"gold": 200,
			},
			wantErr: true,
		},
		{
			name: "met bool requirement",
			requirements: map[string]interface{}{
				"hasKey": true,
			},
			wantErr: false,
		},
		{
			name: "unmet bool requirement",
			requirements: map[string]interface{}{
				"hasKey": false,
			},
			wantErr: true,
		},
		{
			name: "missing variable",
			requirements: map[string]interface{}{
				"missing": 10,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.checkRequirements(progress, tt.requirements)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkRequirements() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestManagerAlignmentClamping(t *testing.T) {
	tests := []struct {
		name     string
		initial  float64
		shift    float64
		expected float64
	}{
		{"clamp to max", 0.95, 0.2, 1.0},
		{"clamp to min", -0.95, -0.2, -1.0},
		{"no clamping needed", 0.0, 0.1, 0.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progress := &PlayerProgress{
				Alignment: map[AlignmentAxis]float64{
					AlignmentGoodEvil: tt.initial,
				},
			}

			applyAlignmentShifts(progress, map[AlignmentAxis]float64{
				AlignmentGoodEvil: tt.shift,
			})

			if progress.Alignment[AlignmentGoodEvil] != tt.expected {
				t.Errorf("Alignment = %f, want %f", progress.Alignment[AlignmentGoodEvil], tt.expected)
			}
		})
	}
}

func TestManagerArcRemovedAfterStart(t *testing.T) {
	manager := NewManager()
	gen := NewGenerator()

	result, _ := gen.Generate(12345, procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	})

	arc := result.(*StoryArc)
	manager.RegisterArc(arc)
	manager.StartArc("player1", arc.ID)

	// Remove the arc from the graph to simulate a missing arc
	delete(manager.graph.Arcs, arc.ID)

	// GetCurrentNode should return error, not panic
	_, err := manager.GetCurrentNode("player1", arc.ID)
	if err == nil {
		t.Error("GetCurrentNode() should fail when arc is missing")
	}

	// AdvanceStory should return error, not panic
	err = manager.AdvanceStory("player1", arc.ID)
	if err == nil {
		t.Error("AdvanceStory() should fail when arc is missing")
	}

	// MakeChoice should return error, not panic
	err = manager.MakeChoice("player1", arc.ID, "some_choice")
	if err == nil {
		t.Error("MakeChoice() should fail when arc is missing")
	}
}

func TestManagerRegisterConsequence(t *testing.T) {
	manager := NewManager()

	consequence := &Consequence{
		ID:          "test_consequence",
		Description: "Test consequence",
		TriggerConditions: map[string]interface{}{
			"evilChoice": true,
		},
		Effects: map[string]interface{}{
			"cursed": true,
		},
	}

	manager.RegisterConsequence(consequence)

	// Verify registered
	if _, exists := manager.graph.Consequences[consequence.ID]; !exists {
		t.Error("Consequence was not registered")
	}
}

func TestManagerGetArc(t *testing.T) {
	manager := NewManager()
	gen := NewGenerator()

	result, _ := gen.Generate(12345, procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	})

	arc := result.(*StoryArc)
	manager.RegisterArc(arc)

	retrieved, err := manager.GetArc(arc.ID)
	if err != nil {
		t.Fatalf("GetArc() failed: %v", err)
	}

	if retrieved.ID != arc.ID {
		t.Errorf("Retrieved arc ID = %s, want %s", retrieved.ID, arc.ID)
	}
}

func TestManagerGetArcNonExistent(t *testing.T) {
	manager := NewManager()

	_, err := manager.GetArc("nonexistent")
	if err == nil {
		t.Error("GetArc() should fail for nonexistent arc")
	}
}

func TestManagerCheckConsequences(t *testing.T) {
	manager := NewManager()
	gen := NewGenerator()

	result, _ := gen.Generate(12345, procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	})

	arc := result.(*StoryArc)
	manager.RegisterArc(arc)
	manager.StartArc("player1", arc.ID)

	// Register a consequence
	consequence := &Consequence{
		ID:          "test_consequence",
		Description: "Test consequence",
		TriggerConditions: map[string]interface{}{
			"completed_quest": true,
			"gold":            100,
		},
		Effects: map[string]interface{}{
			"reward_received": true,
			"bonus_gold":      50,
		},
	}
	manager.RegisterConsequence(consequence)

	// Set up progress with trigger conditions met (exact match for values, type coercion supported)
	progress, _ := manager.GetProgress("player1", arc.ID)
	progress.Variables["completed_quest"] = true
	progress.Variables["gold"] = 100

	// Check consequences
	triggered := manager.CheckConsequences("player1", arc.ID)

	if len(triggered) != 1 {
		t.Fatalf("Expected 1 triggered consequence, got %d", len(triggered))
	}

	if triggered[0] != "test_consequence" {
		t.Errorf("Expected consequence ID 'test_consequence', got %s", triggered[0])
	}

	// Verify effects were applied
	if progress.Variables["reward_received"] != true {
		t.Error("Effect 'reward_received' was not applied")
	}

	if progress.Variables["bonus_gold"] != 50 {
		t.Errorf("Effect 'bonus_gold' = %v, want 50", progress.Variables["bonus_gold"])
	}

	// Verify consequence is tracked as triggered
	triggeredList := toStringSlice(progress.Variables["triggered_consequences"])
	if len(triggeredList) != 1 || triggeredList[0] != "test_consequence" {
		t.Error("Consequence was not tracked as triggered")
	}

	// Check again - should not trigger twice
	triggered = manager.CheckConsequences("player1", arc.ID)
	if len(triggered) != 0 {
		t.Errorf("Consequence should not trigger twice, got %d triggers", len(triggered))
	}
}

func TestManagerCheckConsequencesNoMatch(t *testing.T) {
	manager := NewManager()
	gen := NewGenerator()

	result, _ := gen.Generate(12345, procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	})

	arc := result.(*StoryArc)
	manager.RegisterArc(arc)
	manager.StartArc("player1", arc.ID)

	// Register a consequence with conditions that won't be met
	consequence := &Consequence{
		ID:          "unmet_consequence",
		Description: "Unmet consequence",
		TriggerConditions: map[string]interface{}{
			"rare_item": true,
		},
		Effects: map[string]interface{}{
			"achievement": true,
		},
	}
	manager.RegisterConsequence(consequence)

	// Check consequences without meeting conditions
	triggered := manager.CheckConsequences("player1", arc.ID)

	if len(triggered) != 0 {
		t.Errorf("Expected 0 triggered consequences, got %d", len(triggered))
	}
}

func TestManagerCheckConsequencesInvalidPlayer(t *testing.T) {
	manager := NewManager()

	// Check consequences for non-existent player
	triggered := manager.CheckConsequences("nonexistent", "arc123")

	if triggered != nil {
		t.Error("CheckConsequences should return nil for non-existent player")
	}
}

func TestEvaluateConditions(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name       string
		variables  map[string]interface{}
		conditions map[string]interface{}
		want       bool
	}{
		{
			name:       "empty conditions",
			variables:  map[string]interface{}{},
			conditions: map[string]interface{}{},
			want:       true,
		},
		{
			name: "int condition met",
			variables: map[string]interface{}{
				"level": 10,
			},
			conditions: map[string]interface{}{
				"level": 10,
			},
			want: true,
		},
		{
			name: "int condition not met",
			variables: map[string]interface{}{
				"level": 5,
			},
			conditions: map[string]interface{}{
				"level": 10,
			},
			want: false,
		},
		{
			name: "float64 condition met",
			variables: map[string]interface{}{
				"score": 99.5,
			},
			conditions: map[string]interface{}{
				"score": 99.5,
			},
			want: true,
		},
		{
			name: "float64 condition not met",
			variables: map[string]interface{}{
				"score": 50.0,
			},
			conditions: map[string]interface{}{
				"score": 99.5,
			},
			want: false,
		},
		{
			name: "bool condition met",
			variables: map[string]interface{}{
				"hasKey": true,
			},
			conditions: map[string]interface{}{
				"hasKey": true,
			},
			want: true,
		},
		{
			name: "bool condition not met",
			variables: map[string]interface{}{
				"hasKey": false,
			},
			conditions: map[string]interface{}{
				"hasKey": true,
			},
			want: false,
		},
		{
			name: "string condition met",
			variables: map[string]interface{}{
				"faction": "alliance",
			},
			conditions: map[string]interface{}{
				"faction": "alliance",
			},
			want: true,
		},
		{
			name: "string condition not met",
			variables: map[string]interface{}{
				"faction": "horde",
			},
			conditions: map[string]interface{}{
				"faction": "alliance",
			},
			want: false,
		},
		{
			name: "missing variable",
			variables: map[string]interface{}{
				"gold": 100,
			},
			conditions: map[string]interface{}{
				"silver": 50,
			},
			want: false,
		},
		{
			name: "multiple conditions all met",
			variables: map[string]interface{}{
				"level":  10,
				"hasKey": true,
				"gold":   100,
			},
			conditions: map[string]interface{}{
				"level":  10,
				"hasKey": true,
				"gold":   100,
			},
			want: true,
		},
		{
			name: "multiple conditions one not met",
			variables: map[string]interface{}{
				"level":  10,
				"hasKey": false,
				"gold":   100,
			},
			conditions: map[string]interface{}{
				"level":  10,
				"hasKey": true,
				"gold":   100,
			},
			want: false,
		},
		{
			name: "int as float64 (JSON unmarshaling)",
			variables: map[string]interface{}{
				"count": 5.0, // JSON unmarshals ints as float64
			},
			conditions: map[string]interface{}{
				"count": 5,
			},
			want: true,
		},
		{
			name: "float64 as int (JSON unmarshaling)",
			variables: map[string]interface{}{
				"count": 5,
			},
			conditions: map[string]interface{}{
				"count": 5.0,
			},
			want: true,
		},
		{
			name: "type mismatch",
			variables: map[string]interface{}{
				"value": "string",
			},
			conditions: map[string]interface{}{
				"value": 123,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progress := &PlayerProgress{
				Variables: tt.variables,
			}

			got := manager.evaluateConditions(progress, tt.conditions)
			if got != tt.want {
				t.Errorf("evaluateConditions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToStringSlice(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  []string
	}{
		{
			name:  "nil input",
			input: nil,
			want:  nil,
		},
		{
			name:  "direct []string",
			input: []string{"a", "b", "c"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "empty []string",
			input: []string{},
			want:  []string{},
		},
		{
			name:  "[]interface{} with strings",
			input: []interface{}{"x", "y", "z"},
			want:  []string{"x", "y", "z"},
		},
		{
			name:  "empty []interface{}",
			input: []interface{}{},
			want:  []string{},
		},
		{
			name:  "[]interface{} filters non-string elements",
			input: []interface{}{"a", 123, "b", true, "c"},
			want:  []string{"a", "b", "c"}, // non-strings are skipped
		},
		{
			name:  "wrong type - int",
			input: 123,
			want:  nil,
		},
		{
			name:  "wrong type - string",
			input: "not a slice",
			want:  nil,
		},
		{
			name:  "wrong type - map",
			input: map[string]string{"key": "value"},
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toStringSlice(tt.input)

			if tt.want == nil {
				if got != nil {
					t.Errorf("toStringSlice() = %v, want nil", got)
				}
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("toStringSlice() length = %d, want %d", len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("toStringSlice()[%d] = %s, want %s", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestApplyFactionChanges(t *testing.T) {
	const epsilon = 1e-9 // For floating-point comparison

	tests := []struct {
		name     string
		initial  map[string]float64
		changes  map[string]float64
		expected map[string]float64
	}{
		{
			name:    "new faction positive change",
			initial: map[string]float64{},
			changes: map[string]float64{
				"alliance": 0.5,
			},
			expected: map[string]float64{
				"alliance": 0.5,
			},
		},
		{
			name: "existing faction increase",
			initial: map[string]float64{
				"alliance": 0.3,
			},
			changes: map[string]float64{
				"alliance": 0.4,
			},
			expected: map[string]float64{
				"alliance": 0.7,
			},
		},
		{
			name: "existing faction decrease",
			initial: map[string]float64{
				"horde": 0.5,
			},
			changes: map[string]float64{
				"horde": -0.3,
			},
			expected: map[string]float64{
				"horde": 0.2,
			},
		},
		{
			name: "clamp to max (1.0)",
			initial: map[string]float64{
				"guild": 0.8,
			},
			changes: map[string]float64{
				"guild": 0.5,
			},
			expected: map[string]float64{
				"guild": 1.0,
			},
		},
		{
			name: "clamp to min (-1.0)",
			initial: map[string]float64{
				"enemy": -0.7,
			},
			changes: map[string]float64{
				"enemy": -0.5,
			},
			expected: map[string]float64{
				"enemy": -1.0,
			},
		},
		{
			name: "multiple factions",
			initial: map[string]float64{
				"alliance": 0.2,
				"horde":    -0.3,
			},
			changes: map[string]float64{
				"alliance": 0.3,
				"horde":    0.1,
				"neutral":  0.5,
			},
			expected: map[string]float64{
				"alliance": 0.5,
				"horde":    -0.2,
				"neutral":  0.5,
			},
		},
		{
			name:     "empty changes",
			initial:  map[string]float64{"test": 0.5},
			changes:  map[string]float64{},
			expected: map[string]float64{"test": 0.5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progress := &PlayerProgress{
				Faction: make(map[string]float64),
			}

			// Set initial values
			for faction, value := range tt.initial {
				progress.Faction[faction] = value
			}

			// Apply changes
			applyFactionChanges(progress, tt.changes)

			// Verify expected values
			for faction, expectedValue := range tt.expected {
				actualValue, exists := progress.Faction[faction]
				if !exists {
					t.Errorf("Faction %s not found in result", faction)
					continue
				}

				// Use epsilon comparison for floating-point values
				diff := math.Abs(actualValue - expectedValue)
				if diff > epsilon {
					t.Errorf("Faction %s = %.10f, want %.10f (diff: %.10e)", faction, actualValue, expectedValue, diff)
				}
			}

			// Verify no extra factions
			if len(progress.Faction) != len(tt.expected) {
				t.Errorf("Faction count = %d, want %d", len(progress.Faction), len(tt.expected))
			}
		})
	}
}

func TestCheckRequirementsEdgeCases(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name         string
		variables    map[string]interface{}
		requirements map[string]interface{}
		wantErr      bool
	}{
		{
			name:         "float64 requirement with int variable",
			variables:    map[string]interface{}{"score": 100},
			requirements: map[string]interface{}{"score": 90.5},
			wantErr:      false,
		},
		{
			name:         "int requirement with float64 variable",
			variables:    map[string]interface{}{"level": 10.0},
			requirements: map[string]interface{}{"level": 5},
			wantErr:      false,
		},
		{
			name:         "string requirement met",
			variables:    map[string]interface{}{"faction": "alliance"},
			requirements: map[string]interface{}{"faction": "alliance"},
			wantErr:      false,
		},
		{
			name:         "string requirement not met",
			variables:    map[string]interface{}{"faction": "horde"},
			requirements: map[string]interface{}{"faction": "alliance"},
			wantErr:      true,
		},
		{
			name:         "float64 requirement not met (int variable)",
			variables:    map[string]interface{}{"score": 50},
			requirements: map[string]interface{}{"score": 90.5},
			wantErr:      true,
		},
		{
			name:         "int requirement not met (float64 variable)",
			variables:    map[string]interface{}{"level": 3.0},
			requirements: map[string]interface{}{"level": 5},
			wantErr:      true,
		},
		{
			name:         "type mismatch - int expected, string provided",
			variables:    map[string]interface{}{"value": "not a number"},
			requirements: map[string]interface{}{"value": 10},
			wantErr:      true,
		},
		{
			name:         "type mismatch - float64 expected, string provided",
			variables:    map[string]interface{}{"value": "not a number"},
			requirements: map[string]interface{}{"value": 10.5},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progress := &PlayerProgress{
				Variables: tt.variables,
			}

			err := manager.checkRequirements(progress, tt.requirements)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkRequirements() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Benchmarks

func BenchmarkStartArc(b *testing.B) {
	manager := NewManager()
	gen := NewGenerator()

	result, _ := gen.Generate(12345, procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	})

	arc := result.(*StoryArc)
	manager.RegisterArc(arc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.StartArc("player_"+string(rune(i)), arc.ID)
	}
}

func BenchmarkGetCurrentNode(b *testing.B) {
	manager := NewManager()
	gen := NewGenerator()

	result, _ := gen.Generate(12345, procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	})

	arc := result.(*StoryArc)
	manager.RegisterArc(arc)
	manager.StartArc("player1", arc.ID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.GetCurrentNode("player1", arc.ID)
	}
}

func BenchmarkGetAlignment(b *testing.B) {
	manager := NewManager()
	gen := NewGenerator()

	result, _ := gen.Generate(12345, procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	})

	arc := result.(*StoryArc)
	manager.RegisterArc(arc)
	manager.StartArc("player1", arc.ID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.GetAlignment("player1", arc.ID)
	}
}

func BenchmarkMakeChoice(b *testing.B) {
	manager := NewManager()
	gen := NewGenerator()

	result, _ := gen.Generate(12345, procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	})

	arc := result.(*StoryArc)
	manager.RegisterArc(arc)
	manager.StartArc("player1", arc.ID)

	// Find a choice node with choices
	var choiceNode *StoryNode
	var choiceID string
	for _, node := range arc.Nodes {
		if node.Type == NodeTypeChoice && len(node.Choices) > 0 {
			choiceNode = node
			choiceID = node.Choices[0].ID
			break
		}
	}

	if choiceNode == nil {
		b.Skip("No choice nodes in generated arc")
	}

	// Set current node to choice node
	progress, _ := manager.GetProgress("player1", arc.ID)
	progress.CurrentNodeID = choiceNode.ID

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.MakeChoice("player1", arc.ID, choiceID)
		// Reset to choice node for next iteration
		progress.CurrentNodeID = choiceNode.ID
	}
}

func BenchmarkAdvanceStory(b *testing.B) {
	manager := NewManager()
	gen := NewGenerator()

	result, _ := gen.Generate(12345, procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	})

	arc := result.(*StoryArc)
	manager.RegisterArc(arc)
	manager.StartArc("player1", arc.ID)

	// Find a non-choice node with next node
	var eventNode *StoryNode
	for _, node := range arc.Nodes {
		if node.Type != NodeTypeChoice && node.NextNodeID != "" {
			eventNode = node
			break
		}
	}

	if eventNode == nil {
		b.Skip("No event nodes with next node in generated arc")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset to event node
		progress, _ := manager.GetProgress("player1", arc.ID)
		progress.CurrentNodeID = eventNode.ID
		progress.Completed = false

		manager.AdvanceStory("player1", arc.ID)
	}
}

func BenchmarkCheckConsequences(b *testing.B) {
	manager := NewManager()
	gen := NewGenerator()

	result, _ := gen.Generate(12345, procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	})

	arc := result.(*StoryArc)
	manager.RegisterArc(arc)
	manager.StartArc("player1", arc.ID)

	// Register test consequences
	for i := 0; i < 10; i++ {
		consequence := &Consequence{
			ID:          fmt.Sprintf("consequence_%d", i),
			Description: fmt.Sprintf("Test consequence %d", i),
			TriggerConditions: map[string]interface{}{
				"alignment_good_evil": float64(i) * 0.1,
			},
			Effects: map[string]interface{}{
				fmt.Sprintf("effect_%d", i): true,
			},
		}
		manager.RegisterConsequence(consequence)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.CheckConsequences("player1", arc.ID)
	}
}

func TestManagerSetLogger(t *testing.T) {
	manager := NewManager()

	// Ensure logger is not nil by default
	if manager.logger == nil {
		t.Error("expected default logger to be non-nil")
	}

	// Test that SetLogger accepts custom logger
	customLogger := manager.logger.WithField("custom", "value")
	manager.SetLogger(customLogger)

	// Test that SetLogger ignores nil
	manager.SetLogger(nil)
	if manager.logger == nil {
		t.Error("SetLogger(nil) should not set logger to nil")
	}
}
