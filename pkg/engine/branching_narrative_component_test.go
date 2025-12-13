package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/narrative/branching"
)

func TestBranchingNarrativeComponent(t *testing.T) {
	tests := []struct {
		name string
		comp *BranchingNarrativeComponent
	}{
		{
			name: "empty component",
			comp: &BranchingNarrativeComponent{},
		},
		{
			name: "component with progress",
			comp: &BranchingNarrativeComponent{
				ArcID: "test_arc",
				Progress: &branching.PlayerProgress{
					ArcID:         "test_arc",
					CurrentNodeID: "start",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.comp.Type() != "branching_narrative" {
				t.Errorf("Type() = %v, want branching_narrative", tt.comp.Type())
			}
		})
	}
}

func TestBranchingNarrativeComponentFields(t *testing.T) {
	manager := branching.NewManager()
	progress := &branching.PlayerProgress{
		ArcID:         "test",
		CurrentNodeID: "node1",
	}
	arc := &branching.StoryArc{
		ID:    "test",
		Title: "Test Arc",
	}
	choices := []branching.Choice{
		{ID: "choice1", Text: "Option 1"},
	}

	comp := &BranchingNarrativeComponent{
		ArcID:          "test",
		Progress:       progress,
		ActiveArc:      arc,
		Manager:        manager,
		PendingChoices: choices,
		LastUpdate:     1.5,
	}

	if comp.ArcID != "test" {
		t.Errorf("ArcID = %v, want test", comp.ArcID)
	}
	if comp.Progress != progress {
		t.Error("Progress not set correctly")
	}
	if comp.ActiveArc != arc {
		t.Error("ActiveArc not set correctly")
	}
	if comp.Manager != manager {
		t.Error("Manager not set correctly")
	}
	if len(comp.PendingChoices) != 1 {
		t.Errorf("PendingChoices length = %d, want 1", len(comp.PendingChoices))
	}
	if comp.LastUpdate != 1.5 {
		t.Errorf("LastUpdate = %v, want 1.5", comp.LastUpdate)
	}
}
