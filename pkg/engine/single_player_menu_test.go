package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestSinglePlayerMenuOption_String tests the String method for menu options.
func TestSinglePlayerMenuOption_String(t *testing.T) {
	tests := []struct {
		name     string
		option   SinglePlayerMenuOption
		expected string
	}{
		{"new game", SinglePlayerMenuOptionNewGame, "New Game"},
		{"load game", SinglePlayerMenuOptionLoadGame, "Load Game"},
		{"back", SinglePlayerMenuOptionBack, "Back"},
		{"unknown", SinglePlayerMenuOption(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.option.String()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestNewSinglePlayerMenu tests menu initialization.
func TestNewSinglePlayerMenu(t *testing.T) {
	width, height := 1280, 720
	menu := NewSinglePlayerMenu(width, height)

	if menu == nil {
		t.Fatal("expected non-nil menu")
	}

	if menu.screenWidth != width {
		t.Errorf("expected width %d, got %d", width, menu.screenWidth)
	}

	if menu.screenHeight != height {
		t.Errorf("expected height %d, got %d", height, menu.screenHeight)
	}

	if menu.selectedIdx != 0 {
		t.Errorf("expected initial selection 0, got %d", menu.selectedIdx)
	}

	if menu.visible {
		t.Error("expected menu to be initially invisible")
	}

	expectedOptions := 3
	if len(menu.options) != expectedOptions {
		t.Errorf("expected %d options, got %d", expectedOptions, len(menu.options))
	}

	// Verify option order
	if menu.options[0] != SinglePlayerMenuOptionNewGame {
		t.Error("expected first option to be New Game")
	}
	if menu.options[1] != SinglePlayerMenuOptionLoadGame {
		t.Error("expected second option to be Load Game")
	}
	if menu.options[2] != SinglePlayerMenuOptionBack {
		t.Error("expected third option to be Back")
	}
}

// TestSinglePlayerMenu_ShowHide tests visibility toggling.
func TestSinglePlayerMenu_ShowHide(t *testing.T) {
	menu := NewSinglePlayerMenu(1280, 720)

	// Initially invisible
	if menu.IsVisible() {
		t.Error("expected menu to be initially invisible")
	}

	// Show menu
	menu.Show()
	if !menu.IsVisible() {
		t.Error("expected menu to be visible after Show()")
	}

	// Show resets selection to first option
	menu.selectedIdx = 2
	menu.Show()
	if menu.selectedIdx != 0 {
		t.Errorf("expected selection reset to 0 after Show(), got %d", menu.selectedIdx)
	}

	// Hide menu
	menu.Hide()
	if menu.IsVisible() {
		t.Error("expected menu to be invisible after Hide()")
	}
}

// TestSinglePlayerMenu_SetCallbacks tests callback registration.
func TestSinglePlayerMenu_SetCallbacks(t *testing.T) {
	menu := NewSinglePlayerMenu(1280, 720)

	newGameCalled := false
	loadGameCalled := false
	backCalled := false

	menu.SetNewGameCallback(func() {
		newGameCalled = true
	})

	menu.SetLoadGameCallback(func() {
		loadGameCalled = true
	})

	menu.SetBackCallback(func() {
		backCalled = true
	})

	// Verify callbacks are set
	if menu.onNewGame == nil {
		t.Error("expected onNewGame callback to be set")
	}
	if menu.onLoadGame == nil {
		t.Error("expected onLoadGame callback to be set")
	}
	if menu.onBack == nil {
		t.Error("expected onBack callback to be set")
	}

	// Test callbacks are callable
	menu.onNewGame()
	if !newGameCalled {
		t.Error("expected onNewGame to be called")
	}

	menu.onLoadGame()
	if !loadGameCalled {
		t.Error("expected onLoadGame to be called")
	}

	menu.onBack()
	if !backCalled {
		t.Error("expected onBack to be called")
	}
}

// TestSinglePlayerMenu_Update_NotVisible tests that update does nothing when invisible.
func TestSinglePlayerMenu_Update_NotVisible(t *testing.T) {
	menu := NewSinglePlayerMenu(1280, 720)
	menu.Hide() // Ensure invisible

	callbackCalled := false
	menu.SetNewGameCallback(func() {
		callbackCalled = true
	})

	// Update should return false and not process input
	selected := menu.Update()
	if selected {
		t.Error("expected Update to return false when menu is invisible")
	}

	if callbackCalled {
		t.Error("expected no callback to be called when menu is invisible")
	}
}

// TestSinglePlayerMenu_SelectCurrentOption_NewGame tests New Game selection.
func TestSinglePlayerMenu_SelectCurrentOption_NewGame(t *testing.T) {
	menu := NewSinglePlayerMenu(1280, 720)
	menu.Show()
	menu.selectedIdx = 0 // Select New Game

	newGameCalled := false
	menu.SetNewGameCallback(func() {
		newGameCalled = true
	})

	// Select current option
	selected := menu.selectCurrentOption()
	if !selected {
		t.Error("expected selectCurrentOption to return true for New Game")
	}

	if !newGameCalled {
		t.Error("expected New Game callback to be called")
	}
}

// TestSinglePlayerMenu_SelectCurrentOption_LoadGame tests Load Game selection (disabled).
func TestSinglePlayerMenu_SelectCurrentOption_LoadGame(t *testing.T) {
	menu := NewSinglePlayerMenu(1280, 720)
	menu.Show()
	menu.selectedIdx = 1 // Select Load Game (disabled)

	loadGameCalled := false
	menu.SetLoadGameCallback(func() {
		loadGameCalled = true
	})

	// Select current option (should be disabled)
	selected := menu.selectCurrentOption()
	if selected {
		t.Error("expected selectCurrentOption to return false for disabled Load Game")
	}

	if loadGameCalled {
		t.Error("expected Load Game callback NOT to be called (disabled)")
	}
}

// TestSinglePlayerMenu_SelectCurrentOption_Back tests Back selection.
func TestSinglePlayerMenu_SelectCurrentOption_Back(t *testing.T) {
	menu := NewSinglePlayerMenu(1280, 720)
	menu.Show()
	menu.selectedIdx = 2 // Select Back

	backCalled := false
	menu.SetBackCallback(func() {
		backCalled = true
	})

	// Select current option
	selected := menu.selectCurrentOption()
	if !selected {
		t.Error("expected selectCurrentOption to return true for Back")
	}

	if !backCalled {
		t.Error("expected Back callback to be called")
	}
}

// TestSinglePlayerMenu_Navigation tests keyboard navigation (simulated).
func TestSinglePlayerMenu_Navigation(t *testing.T) {
	menu := NewSinglePlayerMenu(1280, 720)
	menu.Show()

	// Test navigation wrapping
	tests := []struct {
		name        string
		startIdx    int
		navigation  func()
		expectedIdx int
	}{
		{"down from first", 0, func() {
			menu.selectedIdx++
			if menu.selectedIdx >= len(menu.options) {
				menu.selectedIdx = 0
			}
		}, 1},
		{"down from second", 1, func() {
			menu.selectedIdx++
			if menu.selectedIdx >= len(menu.options) {
				menu.selectedIdx = 0
			}
		}, 2},
		{"down wrap", 2, func() {
			menu.selectedIdx++
			if menu.selectedIdx >= len(menu.options) {
				menu.selectedIdx = 0
			}
		}, 0},
		{"up from first wrap", 0, func() {
			menu.selectedIdx--
			if menu.selectedIdx < 0 {
				menu.selectedIdx = len(menu.options) - 1
			}
		}, 2},
		{"up from third", 2, func() {
			menu.selectedIdx--
			if menu.selectedIdx < 0 {
				menu.selectedIdx = len(menu.options) - 1
			}
		}, 1},
		{"up from second", 1, func() {
			menu.selectedIdx--
			if menu.selectedIdx < 0 {
				menu.selectedIdx = len(menu.options) - 1
			}
		}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			menu.selectedIdx = tt.startIdx
			tt.navigation()
			if menu.selectedIdx != tt.expectedIdx {
				t.Errorf("expected index %d, got %d", tt.expectedIdx, menu.selectedIdx)
			}
		})
	}
}

// TestSinglePlayerMenu_GetOptionAtPosition tests mouse position detection.
func TestSinglePlayerMenu_GetOptionAtPosition(t *testing.T) {
	menu := NewSinglePlayerMenu(1280, 720)

	tests := []struct {
		name     string
		x        int
		y        int
		expected int
	}{
		{"outside menu top", 10, 10, -1},
		{"far right", 1200, 300, -1},
		{"far left", 50, 300, -1},
		{"near first option", 640, 300, 0},  // Approximate position of "New Game"
		{"near second option", 640, 340, 1}, // Approximate position of "Load Game"
		{"near third option", 640, 380, 2},  // Approximate position of "Back"
		{"far below", 640, 600, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := menu.getOptionAtPosition(tt.x, tt.y)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// TestSinglePlayerMenu_Draw tests drawing (no-op when invisible).
func TestSinglePlayerMenu_Draw(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Ebiten-dependent test in short mode")
	}

	menu := NewSinglePlayerMenu(1280, 720)

	// Create a minimal test image
	img := ebiten.NewImage(1280, 720)

	// Draw when invisible (should be no-op)
	menu.Hide()
	menu.Draw(img) // Should not panic

	// Draw when visible (should render)
	menu.Show()
	menu.Draw(img) // Should not panic

	// No assertions - just verify it doesn't crash
}

// TestSinglePlayerMenu_DisabledOption tests that disabled options don't trigger callbacks.
func TestSinglePlayerMenu_DisabledOption(t *testing.T) {
	menu := NewSinglePlayerMenu(1280, 720)
	menu.Show()

	// Set Load Game as selected (disabled option)
	menu.selectedIdx = 1

	callbackCalled := false
	menu.SetLoadGameCallback(func() {
		callbackCalled = true
	})

	// Attempt to select disabled option
	selected := menu.selectCurrentOption()

	// Should return false (not selected)
	if selected {
		t.Error("expected disabled option to return false")
	}

	// Callback should NOT be called
	if callbackCalled {
		t.Error("expected callback NOT to be called for disabled option")
	}
}

// TestSinglePlayerMenu_IntegrationWithGame tests menu integration with game state.
func TestSinglePlayerMenu_IntegrationWithGame(t *testing.T) {
	game := NewEbitenGame(1280, 720)

	// Verify menu is initialized
	if game.SinglePlayerMenu == nil {
		t.Fatal("expected SinglePlayerMenu to be initialized")
	}

	// Verify menu is initially invisible
	if game.SinglePlayerMenu.IsVisible() {
		t.Error("expected menu to be initially invisible")
	}

	// Verify callbacks are wired
	if game.SinglePlayerMenu.onNewGame == nil {
		t.Error("expected onNewGame callback to be wired")
	}
	if game.SinglePlayerMenu.onLoadGame == nil {
		t.Error("expected onLoadGame callback to be wired")
	}
	if game.SinglePlayerMenu.onBack == nil {
		t.Error("expected onBack callback to be wired")
	}

	// Test state transitions via handlers
	initialState := game.StateManager.CurrentState()
	if initialState != AppStateMainMenu {
		t.Errorf("expected initial state MainMenu, got %s", initialState.String())
	}

	// Simulate selecting Single-Player from main menu
	game.handleMainMenuSelection(MainMenuOptionSinglePlayer)

	// Should transition to single-player menu state
	if game.StateManager.CurrentState() != AppStateSinglePlayerMenu {
		t.Errorf("expected state SinglePlayerMenu, got %s", game.StateManager.CurrentState().String())
	}

	// Menu should be visible
	if !game.SinglePlayerMenu.IsVisible() {
		t.Error("expected menu to be visible after transition")
	}

	// Test back navigation
	game.handleSinglePlayerMenuBack()

	// Should return to main menu
	if game.StateManager.CurrentState() != AppStateMainMenu {
		t.Errorf("expected state MainMenu after back, got %s", game.StateManager.CurrentState().String())
	}

	// Menu should be hidden
	if game.SinglePlayerMenu.IsVisible() {
		t.Error("expected menu to be hidden after back")
	}
}

// Benchmark single-player menu operations
func BenchmarkSinglePlayerMenu_SelectCurrentOption(b *testing.B) {
	menu := NewSinglePlayerMenu(1280, 720)
	menu.Show()
	menu.SetNewGameCallback(func() {})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		menu.selectCurrentOption()
	}
}

func BenchmarkSinglePlayerMenu_GetOptionAtPosition(b *testing.B) {
	menu := NewSinglePlayerMenu(1280, 720)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		menu.getOptionAtPosition(640, 300)
	}
}
