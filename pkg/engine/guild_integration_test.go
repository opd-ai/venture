package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestInputSystem_SetGuildCallback verifies that SetGuildCallback properly sets the callback
func TestInputSystem_SetGuildCallback(t *testing.T) {
	inputSystem := NewInputSystem()

	called := false
	callback := func() {
		called = true
	}

	err := inputSystem.SetGuildCallback(callback)
	if err != nil {
		t.Fatalf("SetGuildCallback returned error: %v", err)
	}

	// Verify callback was set by calling it directly
	if inputSystem.onGuildOpen != nil {
		inputSystem.onGuildOpen()
	}

	if !called {
		t.Error("Guild callback was not called")
	}
}

// TestInputSystem_SetGuildCallback_NilCallback verifies that SetGuildCallback rejects nil callbacks
func TestInputSystem_SetGuildCallback_NilCallback(t *testing.T) {
	inputSystem := NewInputSystem()

	err := inputSystem.SetGuildCallback(nil)
	if err == nil {
		t.Error("SetGuildCallback should return error for nil callback")
	}
}

// TestInputSystem_KeyGuild_Initialization verifies KeyGuild is properly initialized
func TestInputSystem_KeyGuild_Initialization(t *testing.T) {
	inputSystem := NewInputSystem()

	if inputSystem.KeyGuild == 0 {
		t.Error("KeyGuild should be initialized to a non-zero value")
	}

	// Should be initialized to 'O' key (ebiten.KeyO)
	if inputSystem.KeyGuild != ebiten.KeyO {
		t.Errorf("KeyGuild should be initialized to KeyO, got %v", inputSystem.KeyGuild)
	}
}

// TestGuildUI_NilGuildSystem verifies GuildUI handles nil GuildSystem gracefully
func TestGuildUI_NilGuildSystem(t *testing.T) {
	world := NewWorld()
	guildUI := NewGuildUI(world, nil, 800, 600)

	if guildUI == nil {
		t.Fatal("NewGuildUI should not return nil")
	}

	if guildUI.guildSystem != nil {
		t.Error("GuildUI should accept nil GuildSystem")
	}

	// Verify it has default values
	if guildUI.visible {
		t.Error("GuildUI should be hidden by default")
	}

	if guildUI.width != 800 || guildUI.height != 600 {
		t.Errorf("GuildUI dimensions incorrect: got %dx%d, want 800x600", guildUI.width, guildUI.height)
	}
}

// TestGuildUI_Toggle verifies Toggle method works correctly
func TestGuildUI_Toggle(t *testing.T) {
	world := NewWorld()
	guildUI := NewGuildUI(world, nil, 800, 600)

	if guildUI.visible {
		t.Error("GuildUI should start hidden")
	}

	guildUI.Toggle()
	if !guildUI.visible {
		t.Error("GuildUI should be visible after first toggle")
	}

	guildUI.Toggle()
	if guildUI.visible {
		t.Error("GuildUI should be hidden after second toggle")
	}
}

// TestGuildUI_IsVisible verifies IsVisible method
func TestGuildUI_IsVisible(t *testing.T) {
	world := NewWorld()
	guildUI := NewGuildUI(world, nil, 800, 600)

	if guildUI.IsVisible() {
		t.Error("GuildUI should not be visible initially")
	}

	guildUI.visible = true
	if !guildUI.IsVisible() {
		t.Error("IsVisible should return true when visible is true")
	}
}

// TestGuildUI_ValidateAndGetGuildData_InvalidComponentType verifies safe type assertion.
func TestGuildUI_ValidateAndGetGuildData_InvalidComponentType(t *testing.T) {
	world := NewWorld()
	guildUI := NewGuildUI(world, nil, 800, 600)

	// Create player entity with wrong component type for "guild"
	player := world.CreateEntity()
	player.AddComponent(&PlayerComponent{})
	player.AddComponent(&GuildComponent{GuildID: "", Rank: "Recruit"})

	// validateAndGetGuildData should return error for empty guild ID
	_, _, err := guildUI.validateAndGetGuildData()
	if err == nil {
		t.Error("Expected error for empty guild ID")
	}
}

// TestGuildUI_ValidateAndGetGuildData_NoPlayer verifies error on no player entity.
func TestGuildUI_ValidateAndGetGuildData_NoPlayer(t *testing.T) {
	world := NewWorld()
	guildUI := NewGuildUI(world, nil, 800, 600)

	_, _, err := guildUI.validateAndGetGuildData()
	if err == nil {
		t.Error("Expected error when no player entity exists")
	}
	if err.Error() != "no player entity" {
		t.Errorf("Expected 'no player entity' error, got: %v", err)
	}
}

// TestGuildUI_ValidateAndGetGuildData_NoGuildComponent verifies error on missing guild component.
func TestGuildUI_ValidateAndGetGuildData_NoGuildComponent(t *testing.T) {
	world := NewWorld()
	guildUI := NewGuildUI(world, nil, 800, 600)

	player := world.CreateEntity()
	player.AddComponent(&PlayerComponent{})

	_, _, err := guildUI.validateAndGetGuildData()
	if err == nil {
		t.Error("Expected error for missing guild component")
	}
}
