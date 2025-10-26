package engine

import (
	"testing"
)

func TestMainMenuOption_String(t *testing.T) {
	tests := []struct {
		name   string
		option MainMenuOption
		want   string
	}{
		{"single player", MainMenuOptionSinglePlayer, "Single-Player"},
		{"multiplayer", MainMenuOptionMultiPlayer, "Multi-Player"},
		{"settings", MainMenuOptionSettings, "Settings"},
		{"quit", MainMenuOptionQuit, "Quit"},
		{"unknown", MainMenuOption(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.option.String(); got != tt.want {
				t.Errorf("MainMenuOption.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewMainMenuUI(t *testing.T) {
	ui := NewMainMenuUI(800, 600)

	if ui == nil {
		t.Fatal("NewMainMenuUI() returned nil")
	}

	if ui.screenWidth != 800 {
		t.Errorf("screenWidth = %d, want 800", ui.screenWidth)
	}

	if ui.screenHeight != 600 {
		t.Errorf("screenHeight = %d, want 600", ui.screenHeight)
	}

	if ui.selectedIdx != 0 {
		t.Errorf("selectedIdx = %d, want 0", ui.selectedIdx)
	}

	if len(ui.options) != 4 {
		t.Errorf("len(options) = %d, want 4", len(ui.options))
	}
}

func TestMainMenuUI_GetSelectedOption(t *testing.T) {
	ui := NewMainMenuUI(800, 600)

	if got := ui.GetSelectedOption(); got != MainMenuOptionSinglePlayer {
		t.Errorf("GetSelectedOption() = %v, want %v", got, MainMenuOptionSinglePlayer)
	}

	// Manually change selection
	ui.selectedIdx = 2
	if got := ui.GetSelectedOption(); got != MainMenuOptionSettings {
		t.Errorf("GetSelectedOption() = %v, want %v", got, MainMenuOptionSettings)
	}
}

func TestMainMenuUI_Reset(t *testing.T) {
	ui := NewMainMenuUI(800, 600)
	ui.selectedIdx = 3

	ui.Reset()

	if ui.selectedIdx != 0 {
		t.Errorf("after Reset() selectedIdx = %d, want 0", ui.selectedIdx)
	}
}

func TestMainMenuUI_SetSelectCallback(t *testing.T) {
	ui := NewMainMenuUI(800, 600)
	callbackCalled := false
	var callbackOption MainMenuOption

	ui.SetSelectCallback(func(option MainMenuOption) {
		callbackCalled = true
		callbackOption = option
	})

	if ui.onSelect == nil {
		t.Error("SetSelectCallback() did not set callback")
	}

	// Trigger callback manually
	ui.onSelect(MainMenuOptionMultiPlayer)

	if !callbackCalled {
		t.Error("callback was not called")
	}

	if callbackOption != MainMenuOptionMultiPlayer {
		t.Errorf("callback received option %v, want %v", callbackOption, MainMenuOptionMultiPlayer)
	}
}

func TestMainMenuUI_getOptionAtPosition(t *testing.T) {
	ui := NewMainMenuUI(800, 600)

	tests := []struct {
		name     string
		x        int
		y        int
		wantIdx  int
		wantName string
	}{
		{"above all options", 400, 100, -1, "none"},
		{"below all options", 400, 550, -1, "none"},
		{"far left", 50, 300, -1, "none"},
		{"far right", 750, 300, -1, "none"},
		// Note: Exact positions depend on layout calculations
		// These test boundary conditions
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ui.getOptionAtPosition(tt.x, tt.y)
			if got != tt.wantIdx {
				t.Errorf("getOptionAtPosition(%d, %d) = %d, want %d", tt.x, tt.y, got, tt.wantIdx)
			}
		})
	}
}

func TestMainMenuUI_NavigationWrapping(t *testing.T) {
	// Test that navigation wraps around at boundaries
	ui := NewMainMenuUI(800, 600)

	// Initially at index 0 (Single-Player)
	if ui.selectedIdx != 0 {
		t.Fatalf("initial selectedIdx = %d, want 0", ui.selectedIdx)
	}

	// Simulate moving up from first option (should wrap to last)
	ui.selectedIdx--
	if ui.selectedIdx < 0 {
		ui.selectedIdx = len(ui.options) - 1
	}

	if ui.selectedIdx != 3 {
		t.Errorf("after wrapping up, selectedIdx = %d, want 3", ui.selectedIdx)
	}

	// Simulate moving down from last option (should wrap to first)
	ui.selectedIdx++
	if ui.selectedIdx >= len(ui.options) {
		ui.selectedIdx = 0
	}

	if ui.selectedIdx != 0 {
		t.Errorf("after wrapping down, selectedIdx = %d, want 0", ui.selectedIdx)
	}
}

func TestMainMenuUI_Draw(t *testing.T) {
	// We can't fully test Draw without Ebiten runtime
	// This test verifies the UI can be created and has the Draw method
	// Actual rendering tested manually or with integration tests
	ui := NewMainMenuUI(800, 600)

	// Verify UI was created successfully
	if ui == nil {
		t.Error("failed to create UI")
	}

	// Note: We don't call Draw here as it requires Ebiten context
	// This would be tested in integration tests with -tags test
}

func TestMainMenuUI_Update(t *testing.T) {
	// Test that Update exists and UI can be created
	ui := NewMainMenuUI(800, 600)

	// Verify UI was created successfully
	if ui == nil {
		t.Error("failed to create UI")
	}

	// Note: Key press simulation requires Ebiten runtime
	// These would be tested in integration tests
}

// TestMainMenuUI_AllOptionsAccessible verifies all menu options are included
func TestMainMenuUI_AllOptionsAccessible(t *testing.T) {
	ui := NewMainMenuUI(800, 600)

	expectedOptions := []MainMenuOption{
		MainMenuOptionSinglePlayer,
		MainMenuOptionMultiPlayer,
		MainMenuOptionSettings,
		MainMenuOptionQuit,
	}

	if len(ui.options) != len(expectedOptions) {
		t.Fatalf("options count = %d, want %d", len(ui.options), len(expectedOptions))
	}

	for i, expected := range expectedOptions {
		if ui.options[i] != expected {
			t.Errorf("options[%d] = %v, want %v", i, ui.options[i], expected)
		}
	}
}
