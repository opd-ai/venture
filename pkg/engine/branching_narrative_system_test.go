package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/narrative/branching"
)

func TestNewBranchingNarrativeSystem(t *testing.T) {
	world := NewWorld()
	system := NewBranchingNarrativeSystem(world)

	if system == nil {
		t.Fatal("NewBranchingNarrativeSystem returned nil")
	}
	if system.world != world {
		t.Error("System world not set correctly")
	}
}

func TestBranchingNarrativeSystemUpdate(t *testing.T) {
	world := NewWorld()
	system := NewBranchingNarrativeSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&BranchingNarrativeComponent{
		LastUpdate: 0,
	})

	// First update should not trigger (needs 1 second)
	system.Update([]*Entity{entity}, 0.5)
	comp, _ := entity.GetComponent("branching_narrative")
	narComp := comp.(*BranchingNarrativeComponent)
	if narComp.LastUpdate != 0.5 {
		t.Errorf("LastUpdate = %v, want 0.5", narComp.LastUpdate)
	}

	// Second update should reset to 0 (not remainder)
	system.Update([]*Entity{entity}, 0.6)
	if narComp.LastUpdate != 0 { // Resets to 0 after crossing 1.0 threshold
		t.Errorf("LastUpdate = %v, want 0", narComp.LastUpdate)
	}
}

func TestBranchingNarrativeSystemStartStoryArc(t *testing.T) {
	world := NewWorld()
	system := NewBranchingNarrativeSystem(world)
	manager := branching.NewManager()

	// Create test arc
	arc := &branching.StoryArc{
		ID:          "test_arc",
		Title:       "Test Story",
		StartNodeID: "start",
		Nodes: map[string]*branching.StoryNode{
			"start": {
				ID:          "start",
				Type:        branching.NodeTypeStart,
				Title:       "Beginning",
				Description: "The story begins...",
				NextNodeID:  "choice1",
			},
		},
	}

	t.Run("entity without component", func(t *testing.T) {
		entity := world.CreateEntity()
		err := system.StartStoryArc(entity, arc, manager)
		if err == nil {
			t.Error("Expected error for entity without component")
		}
	})

	t.Run("successful start", func(t *testing.T) {
		entity := world.CreateEntity()
		entity.AddComponent(&BranchingNarrativeComponent{})

		err := system.StartStoryArc(entity, arc, manager)
		if err != nil {
			t.Errorf("StartStoryArc failed: %v", err)
		}

		comp, _ := entity.GetComponent("branching_narrative")
		narComp := comp.(*BranchingNarrativeComponent)
		if narComp.ArcID != "test_arc" {
			t.Errorf("ArcID = %v, want test_arc", narComp.ArcID)
		}
		if narComp.Progress == nil {
			t.Error("Progress not initialized")
		}
		if narComp.ActiveArc != arc {
			t.Error("ActiveArc not set")
		}
		if narComp.Manager != manager {
			t.Error("Manager not set")
		}
	})

	t.Run("start with choice node", func(t *testing.T) {
		choiceArc := &branching.StoryArc{
			ID:          "choice_arc",
			Title:       "Choice Story",
			StartNodeID: "start",
			Nodes: map[string]*branching.StoryNode{
				"start": {
					ID:          "start",
					Type:        branching.NodeTypeChoice,
					Title:       "First Choice",
					Description: "Choose your path",
					Choices: []branching.Choice{
						{ID: "c1", Text: "Go left", NextNodeID: "left"},
						{ID: "c2", Text: "Go right", NextNodeID: "right"},
					},
				},
			},
		}

		entity := world.CreateEntity()
		entity.AddComponent(&BranchingNarrativeComponent{})

		err := system.StartStoryArc(entity, choiceArc, manager)
		if err != nil {
			t.Errorf("StartStoryArc failed: %v", err)
		}

		comp, _ := entity.GetComponent("branching_narrative")
		narComp := comp.(*BranchingNarrativeComponent)
		if len(narComp.PendingChoices) != 2 {
			t.Errorf("PendingChoices length = %d, want 2", len(narComp.PendingChoices))
		}
	})
}

func TestBranchingNarrativeSystemMakeChoice(t *testing.T) {
	world := NewWorld()
	system := NewBranchingNarrativeSystem(world)
	manager := branching.NewManager()

	// Create test arc with choices
	arc := &branching.StoryArc{
		ID:          "test_arc",
		Title:       "Test Story",
		StartNodeID: "choice1",
		Nodes: map[string]*branching.StoryNode{
			"choice1": {
				ID:          "choice1",
				Type:        branching.NodeTypeChoice,
				Title:       "First Choice",
				Description: "What do you do?",
				Choices: []branching.Choice{
					{ID: "attack", Text: "Attack", NextNodeID: "combat"},
					{ID: "flee", Text: "Flee", NextNodeID: "escape"},
				},
			},
			"combat": {
				ID:          "combat",
				Type:        branching.NodeTypeEvent,
				Title:       "Combat",
				Description: "You fight bravely",
				NextNodeID:  "victory",
			},
			"escape": {
				ID:          "escape",
				Type:        branching.NodeTypeEvent,
				Title:       "Escape",
				Description: "You run away",
				NextNodeID:  "safety",
			},
		},
	}

	entity := world.CreateEntity()
	entity.AddComponent(&BranchingNarrativeComponent{})

	t.Run("no active arc", func(t *testing.T) {
		err := system.MakeChoice(entity, "attack")
		if err == nil {
			t.Error("Expected error for no active arc")
		}
	})

	// Start the arc
	if err := system.StartStoryArc(entity, arc, manager); err != nil {
		t.Fatalf("StartStoryArc failed: %v", err)
	}

	t.Run("invalid choice", func(t *testing.T) {
		err := system.MakeChoice(entity, "invalid")
		if err == nil {
			t.Error("Expected error for invalid choice")
		}
	})

	t.Run("valid choice", func(t *testing.T) {
		err := system.MakeChoice(entity, "attack")
		if err != nil {
			t.Errorf("MakeChoice failed: %v", err)
		}

		comp, _ := entity.GetComponent("branching_narrative")
		narComp := comp.(*BranchingNarrativeComponent)
		if len(narComp.PendingChoices) != 0 {
			t.Error("PendingChoices should be cleared after choice")
		}
	})
}

func TestBranchingNarrativeSystemGetCurrentNode(t *testing.T) {
	world := NewWorld()
	system := NewBranchingNarrativeSystem(world)
	manager := branching.NewManager()

	arc := &branching.StoryArc{
		ID:          "test",
		Title:       "Test",
		StartNodeID: "start",
		Nodes: map[string]*branching.StoryNode{
			"start": {
				ID:          "start",
				Type:        branching.NodeTypeStart,
				Title:       "Start",
				Description: "Beginning",
				NextNodeID:  "end",
			},
		},
	}

	entity := world.CreateEntity()
	entity.AddComponent(&BranchingNarrativeComponent{})

	t.Run("no active arc", func(t *testing.T) {
		_, err := system.GetCurrentNode(entity)
		if err == nil {
			t.Error("Expected error for no active arc")
		}
	})

	system.StartStoryArc(entity, arc, manager)

	t.Run("get current node", func(t *testing.T) {
		node, err := system.GetCurrentNode(entity)
		if err != nil {
			t.Errorf("GetCurrentNode failed: %v", err)
		}
		if node == nil {
			t.Fatal("GetCurrentNode returned nil")
		}
		if node.ID != "start" {
			t.Errorf("node.ID = %v, want start", node.ID)
		}
	})
}

func TestBranchingNarrativeSystemGetAlignment(t *testing.T) {
	world := NewWorld()
	system := NewBranchingNarrativeSystem(world)
	manager := branching.NewManager()

	entity := world.CreateEntity()
	entity.AddComponent(&BranchingNarrativeComponent{})

	arc := &branching.StoryArc{
		ID:          "test",
		Title:       "Test",
		StartNodeID: "start",
		Nodes: map[string]*branching.StoryNode{
			"start": {
				ID:          "start",
				Type:        branching.NodeTypeStart,
				Title:       "Start",
				Description: "Beginning",
			},
		},
	}
	system.StartStoryArc(entity, arc, manager)

	alignment, err := system.GetAlignment(entity)
	if err != nil {
		t.Errorf("GetAlignment failed: %v", err)
	}
	if alignment == nil {
		t.Error("GetAlignment returned nil")
	}
}

func TestBranchingNarrativeSystemGetFactionReputation(t *testing.T) {
	world := NewWorld()
	system := NewBranchingNarrativeSystem(world)
	manager := branching.NewManager()

	entity := world.CreateEntity()
	entity.AddComponent(&BranchingNarrativeComponent{})

	arc := &branching.StoryArc{
		ID:          "test",
		Title:       "Test",
		StartNodeID: "start",
		Nodes: map[string]*branching.StoryNode{
			"start": {
				ID:          "start",
				Type:        branching.NodeTypeStart,
				Title:       "Start",
				Description: "Beginning",
			},
		},
	}
	system.StartStoryArc(entity, arc, manager)

	rep, err := system.GetFactionReputation(entity)
	if err != nil {
		t.Errorf("GetFactionReputation failed: %v", err)
	}
	if rep == nil {
		t.Error("GetFactionReputation returned nil")
	}
}
