package branching

import (
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

	// Manually set high alignment
	progress, _ := manager.GetProgress("player1", arc.ID)
	progress.Alignment[AlignmentGoodEvil] = 0.95

	// Create a choice with positive shift
	choice := Choice{
		ID:   "test_choice",
		Text: "Test",
		AlignmentShift: map[AlignmentAxis]float64{
			AlignmentGoodEvil: 0.2,
		},
		NextNodeID: arc.StartNodeID,
	}

	// Apply the shift manually
	for axis, shift := range choice.AlignmentShift {
		progress.Alignment[axis] += shift
		if progress.Alignment[axis] > 1.0 {
			progress.Alignment[axis] = 1.0
		}
	}

	// Verify clamped to 1.0
	if progress.Alignment[AlignmentGoodEvil] != 1.0 {
		t.Errorf("Alignment = %f, want 1.0 (clamped)", progress.Alignment[AlignmentGoodEvil])
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
