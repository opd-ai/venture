package engine

import (
	"fmt"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/dialog"
)

func TestNewDialogUI(t *testing.T) {
	world := NewWorld()
	dialogSys := NewDialogSystem(world)
	npcDialogSys := NewNPCDialogSystem(world, 12345)

	ui := NewDialogUI(world, dialogSys, npcDialogSys, 800, 600)

	if ui == nil {
		t.Fatal("NewDialogUI returned nil")
	}
	if ui.visible {
		t.Error("DialogUI should not be visible on creation")
	}
	if ui.state != DialogUIIdle {
		t.Errorf("DialogUI state should be Idle, got %v", ui.state)
	}
	if ui.screenWidth != 800 || ui.screenHeight != 600 {
		t.Errorf("DialogUI dimensions incorrect, got %dx%d", ui.screenWidth, ui.screenHeight)
	}
}

func TestDialogUISetPlayerEntity(t *testing.T) {
	world := NewWorld()
	dialogSys := NewDialogSystem(world)
	npcDialogSys := NewNPCDialogSystem(world, 12345)
	ui := NewDialogUI(world, dialogSys, npcDialogSys, 800, 600)

	player := world.CreateEntity()
	ui.SetPlayerEntity(player)

	if ui.player == nil {
		t.Error("Player entity not set")
	}
	if ui.player.ID != player.ID {
		t.Errorf("Player ID mismatch, expected %d, got %d", player.ID, ui.player.ID)
	}
}

func TestDialogUIIsVisible(t *testing.T) {
	world := NewWorld()
	dialogSys := NewDialogSystem(world)
	npcDialogSys := NewNPCDialogSystem(world, 12345)
	ui := NewDialogUI(world, dialogSys, npcDialogSys, 800, 600)

	if ui.IsVisible() {
		t.Error("DialogUI should not be visible initially")
	}

	ui.visible = true
	if !ui.IsVisible() {
		t.Error("DialogUI should be visible after setting visible=true")
	}
}

func TestDialogUIShowNoPlayer(t *testing.T) {
	world := NewWorld()
	dialogSys := NewDialogSystem(world)
	npcDialogSys := NewNPCDialogSystem(world, 12345)
	ui := NewDialogUI(world, dialogSys, npcDialogSys, 800, 600)

	// Create NPC
	npc := world.CreateEntity()

	// Try to show without player
	err := ui.Show(npc.ID)
	if err == nil {
		t.Error("Show should fail without player entity set")
	}
}

func TestDialogUIShowNPCNotFound(t *testing.T) {
	world := NewWorld()
	dialogSys := NewDialogSystem(world)
	npcDialogSys := NewNPCDialogSystem(world, 12345)
	ui := NewDialogUI(world, dialogSys, npcDialogSys, 800, 600)

	player := world.CreateEntity()
	ui.SetPlayerEntity(player)

	// Try to show with non-existent NPC
	err := ui.Show(99999)
	if err == nil {
		t.Error("Show should fail for non-existent NPC")
	}
}

func TestDialogUIShowSuccess(t *testing.T) {
	world := NewWorld()
	dialogSys := NewDialogSystem(world)
	npcDialogSys := NewNPCDialogSystem(world, 12345)
	ui := NewDialogUI(world, dialogSys, npcDialogSys, 800, 600)

	player := world.CreateEntity()
	ui.SetPlayerEntity(player)

	// Create NPC with required components
	npc := world.CreateEntity()
	npc.AddComponent(&DialogComponent{})
	npc.AddComponent(NewNPCDialogComponent("fantasy", dialog.NewPersonality(dialog.PersonalityHelpful), 12345))

	// Initialize NPC dialog
	personality := dialog.NewPersonality(dialog.PersonalityMerchant)
	err := npcDialogSys.InitializeNPCDialog(npc, "fantasy", personality, 12345)
	if err != nil {
		t.Fatalf("Failed to initialize NPC dialog: %v", err)
	}

	// Update world to process entity additions
	world.Update(0)

	// Show dialog
	err = ui.Show(npc.ID)
	if err != nil {
		t.Errorf("Show failed: %v", err)
	}

	if !ui.visible {
		t.Error("DialogUI should be visible after Show")
	}
	if ui.state != DialogUIGreeting && ui.state != DialogUIOptions {
		t.Errorf("DialogUI state should be Greeting or Options, got %v", ui.state)
	}
	if ui.currentNPCID != npc.ID {
		t.Errorf("Current NPC ID mismatch, expected %d, got %d", npc.ID, ui.currentNPCID)
	}
	expectedName := fmt.Sprintf("NPC #%d", npc.ID)
	if ui.currentNPCName != expectedName {
		t.Errorf("NPC name incorrect, expected '%s', got '%s'", expectedName, ui.currentNPCName)
	}
	if len(ui.playerOptions) == 0 {
		t.Error("Player options should be set up")
	}
}

func TestDialogUIHide(t *testing.T) {
	world := NewWorld()
	dialogSys := NewDialogSystem(world)
	npcDialogSys := NewNPCDialogSystem(world, 12345)
	ui := NewDialogUI(world, dialogSys, npcDialogSys, 800, 600)

	player := world.CreateEntity()
	ui.SetPlayerEntity(player)

	npc := world.CreateEntity()
	npc.AddComponent(&DialogComponent{})
	npc.AddComponent(&NPCDialogComponent{NPCPersonality: dialog.NewPersonality(dialog.PersonalityHelpful)})

	npcDialogSys.InitializeNPCDialog(npc, "fantasy", dialog.NewPersonality(dialog.PersonalityHelpful), 12345)
	ui.Show(npc.ID)

	// Hide dialog
	ui.Hide()

	if ui.visible {
		t.Error("DialogUI should not be visible after Hide")
	}
	if ui.state != DialogUIIdle {
		t.Errorf("DialogUI state should be Idle after Hide, got %v", ui.state)
	}
	if ui.currentNPCID != 0 {
		t.Error("Current NPC ID should be cleared after Hide")
	}
	if len(ui.history) != 0 {
		t.Error("History should be cleared after Hide")
	}
}

func TestDialogUIInitialOptions(t *testing.T) {
	world := NewWorld()
	dialogSys := NewDialogSystem(world)
	npcDialogSys := NewNPCDialogSystem(world, 12345)
	ui := NewDialogUI(world, dialogSys, npcDialogSys, 800, 600)

	ui.setupInitialOptions()

	if len(ui.playerOptions) == 0 {
		t.Fatal("Initial options should be set up")
	}
	if ui.selectedOption != 0 {
		t.Error("Selected option should be 0 initially")
	}
	if ui.state != DialogUIOptions {
		t.Errorf("State should be Options, got %v", ui.state)
	}

	// Check for expected options
	hasGoodbye := false
	for _, opt := range ui.playerOptions {
		if opt == "Goodbye" {
			hasGoodbye = true
		}
	}
	if !hasGoodbye {
		t.Error("Initial options should include 'Goodbye'")
	}
}

func TestDialogUIAddToHistory(t *testing.T) {
	world := NewWorld()
	dialogSys := NewDialogSystem(world)
	npcDialogSys := NewNPCDialogSystem(world, 12345)
	ui := NewDialogUI(world, dialogSys, npcDialogSys, 800, 600)
	ui.maxHistory = 3

	ui.addToHistory("Line 1")
	ui.addToHistory("Line 2")
	ui.addToHistory("Line 3")

	if len(ui.history) != 3 {
		t.Errorf("History length should be 3, got %d", len(ui.history))
	}

	// Add more to trigger limit
	ui.addToHistory("Line 4")
	ui.addToHistory("Line 5")

	if len(ui.history) != 3 {
		t.Errorf("History should be limited to maxHistory (3), got %d", len(ui.history))
	}

	// Check oldest entries were removed
	if ui.history[0] != "Line 3" {
		t.Errorf("Oldest entry should be 'Line 3', got '%s'", ui.history[0])
	}
	if ui.history[2] != "Line 5" {
		t.Errorf("Newest entry should be 'Line 5', got '%s'", ui.history[2])
	}
}

func TestDialogUIUpdate(t *testing.T) {
	world := NewWorld()
	dialogSys := NewDialogSystem(world)
	npcDialogSys := NewNPCDialogSystem(world, 12345)
	ui := NewDialogUI(world, dialogSys, npcDialogSys, 800, 600)

	// Update when not visible should be no-op
	err := ui.Update()
	if err != nil {
		t.Errorf("Update should not error when not visible: %v", err)
	}

	// Make visible and update
	ui.visible = true
	ui.playerOptions = []string{"Option 1", "Option 2", "Option 3"}
	ui.selectedOption = 0

	err = ui.Update()
	if err != nil {
		t.Errorf("Update should not error: %v", err)
	}
}

func TestDialogUIGenerateGreeting(t *testing.T) {
	world := NewWorld()
	dialogSys := NewDialogSystem(world)
	npcDialogSys := NewNPCDialogSystem(world, 12345)
	ui := NewDialogUI(world, dialogSys, npcDialogSys, 800, 600)

	// Test with NPC that has dialog component
	npc := world.CreateEntity()
	npc.AddComponent(NewNPCDialogComponent("fantasy", dialog.NewPersonality(dialog.PersonalityHelpful), 12345))

	err := ui.generateGreeting(npc)
	if err != nil {
		t.Errorf("generateGreeting failed: %v", err)
	}
	if ui.npcText == "" {
		t.Error("NPC text should be set after greeting generation")
	}

	// Test with NPC without dialog component (fallback)
	npc2 := world.CreateEntity()
	err = ui.generateGreeting(npc2)
	if err != nil {
		t.Errorf("generateGreeting should not fail without dialog component: %v", err)
	}
	if ui.npcText != "Hello, traveler." {
		t.Errorf("Expected fallback greeting, got '%s'", ui.npcText)
	}
}

func TestDialogUIHandleOptionSelectGoodbye(t *testing.T) {
	world := NewWorld()
	dialogSys := NewDialogSystem(world)
	npcDialogSys := NewNPCDialogSystem(world, 12345)
	ui := NewDialogUI(world, dialogSys, npcDialogSys, 800, 600)

	player := world.CreateEntity()
	ui.SetPlayerEntity(player)

	npc := world.CreateEntity()
	npc.AddComponent(&DialogComponent{})
	npc.AddComponent(&NPCDialogComponent{NPCPersonality: dialog.NewPersonality(dialog.PersonalityHelpful)})
	npcDialogSys.InitializeNPCDialog(npc, "fantasy", dialog.NewPersonality(dialog.PersonalityHelpful), 12345)
	ui.currentNPCID = npc.ID
	ui.currentNPCName = "NPC"
	ui.visible = true
	ui.state = DialogUIOptions

	ui.playerOptions = []string{"Hello", "Goodbye"}
	ui.selectedOption = 1 // Select Goodbye

	err := ui.handleOptionSelect()
	if err != nil {
		t.Errorf("handleOptionSelect failed: %v", err)
	}

	if ui.state != DialogUIEnding {
		t.Errorf("State should be Ending after goodbye, got %v", ui.state)
	}
	if len(ui.playerOptions) != 1 {
		t.Errorf("Should have one option ([Close]), got %d", len(ui.playerOptions))
	}
}
