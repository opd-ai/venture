package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/narrative/branching"
)

func TestNewStoryChoiceUI(t *testing.T) {
	world := NewWorld()
	system := NewBranchingNarrativeSystem(world)
	ui := NewStoryChoiceUI(world, system, 800, 600)

	if ui == nil {
		t.Fatal("NewStoryChoiceUI returned nil")
	}
	if ui.world != world {
		t.Error("world not set correctly")
	}
	if ui.narrativeSystem != system {
		t.Error("narrativeSystem not set correctly")
	}
	if ui.screenWidth != 800 {
		t.Errorf("screenWidth = %d, want 800", ui.screenWidth)
	}
	if ui.screenHeight != 600 {
		t.Errorf("screenHeight = %d, want 600", ui.screenHeight)
	}
	if ui.visible {
		t.Error("UI should not be visible initially")
	}
}

func TestStoryChoiceUISetPlayerEntity(t *testing.T) {
	world := NewWorld()
	system := NewBranchingNarrativeSystem(world)
	ui := NewStoryChoiceUI(world, system, 800, 600)

	entity := world.CreateEntity()
	ui.SetPlayerEntity(entity)

	if ui.playerEntity != entity {
		t.Error("playerEntity not set correctly")
	}
}

func TestStoryChoiceUIUpdate(t *testing.T) {
	world := NewWorld()
	system := NewBranchingNarrativeSystem(world)
	ui := NewStoryChoiceUI(world, system, 800, 600)

	entity := world.CreateEntity()
	entity.AddComponent(&BranchingNarrativeComponent{})
	ui.SetPlayerEntity(entity)

	t.Run("no pending choices", func(t *testing.T) {
		ui.Update(1.0)
		if ui.visible {
			t.Error("UI should not be visible without pending choices")
		}
	})

	t.Run("with pending choices", func(t *testing.T) {
		manager := branching.NewManager()
		arc := &branching.StoryArc{
			ID:          "test",
			Title:       "Test",
			StartNodeID: "choice1",
			Nodes: map[string]*branching.StoryNode{
				"choice1": {
					ID:          "choice1",
					Type:        branching.NodeTypeChoice,
					Title:       "First Choice",
					Description: "What do you do?",
					Choices: []branching.Choice{
						{ID: "c1", Text: "Option 1"},
						{ID: "c2", Text: "Option 2"},
					},
				},
			},
		}

		system.StartStoryArc(entity, arc, manager)
		ui.Update(1.0)

		if !ui.visible {
			t.Error("UI should be visible with pending choices")
		}
		if ui.currentNode == nil {
			t.Error("currentNode should be set")
		}
		if len(ui.pendingChoices) != 2 {
			t.Errorf("pendingChoices length = %d, want 2", len(ui.pendingChoices))
		}
	})
}

func TestStoryChoiceUIVisibility(t *testing.T) {
	world := NewWorld()
	system := NewBranchingNarrativeSystem(world)
	ui := NewStoryChoiceUI(world, system, 800, 600)

	if ui.IsVisible() {
		t.Error("UI should not be visible initially")
	}

	ui.visible = true
	if !ui.IsVisible() {
		t.Error("IsVisible should return true when visible")
	}

	ui.Hide()
	if ui.IsVisible() {
		t.Error("UI should be hidden after Hide()")
	}
}

func TestStoryChoiceUIHandleInput(t *testing.T) {
	world := NewWorld()
	system := NewBranchingNarrativeSystem(world)
	ui := NewStoryChoiceUI(world, system, 800, 600)

	t.Run("not visible", func(t *testing.T) {
		handled := ui.HandleInput()
		if handled {
			t.Error("HandleInput should return false when not visible")
		}
	})

	t.Run("no choices", func(t *testing.T) {
		ui.visible = true
		handled := ui.HandleInput()
		if handled {
			t.Error("HandleInput should return false with no choices")
		}
	})
}

func TestStoryChoiceUIShow(t *testing.T) {
	world := NewWorld()
	system := NewBranchingNarrativeSystem(world)
	ui := NewStoryChoiceUI(world, system, 800, 600)

	entity := world.CreateEntity()
	entity.AddComponent(&BranchingNarrativeComponent{})
	ui.SetPlayerEntity(entity)

	// Show without choices should not make visible
	ui.Show()
	if ui.IsVisible() {
		t.Error("Show should not make visible without choices")
	}

	// Add choices
	manager := branching.NewManager()
	arc := &branching.StoryArc{
		ID:          "test",
		Title:       "Test",
		StartNodeID: "choice1",
		Nodes: map[string]*branching.StoryNode{
			"choice1": {
				ID:          "choice1",
				Type:        branching.NodeTypeChoice,
				Title:       "Choice",
				Description: "Choose",
				Choices: []branching.Choice{
					{ID: "c1", Text: "Option 1"},
				},
			},
		},
	}
	system.StartStoryArc(entity, arc, manager)

	ui.Show()
	if !ui.IsVisible() {
		t.Error("Show should make visible with choices")
	}
}

func TestStoryChoiceUIUpdateTiming(t *testing.T) {
	world := NewWorld()
	system := NewBranchingNarrativeSystem(world)
	ui := NewStoryChoiceUI(world, system, 800, 600)

	entity := world.CreateEntity()
	entity.AddComponent(&BranchingNarrativeComponent{})
	ui.SetPlayerEntity(entity)

	// First update with small delta - should not check
	ui.Update(0.1)
	if ui.lastCheckTime != 0.1 {
		t.Errorf("lastCheckTime = %v, want 0.1", ui.lastCheckTime)
	}

	// Second update reaching 0.5 threshold - should check
	ui.Update(0.4)
	if ui.lastCheckTime != 0 { // Should reset to 0 after check
		t.Errorf("lastCheckTime = %v, want 0 (should reset)", ui.lastCheckTime)
	}
}
