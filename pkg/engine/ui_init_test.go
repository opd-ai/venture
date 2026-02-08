package engine

import (
	"testing"
)

// TestInitializeUIComponents_AllComponentsInitialized verifies that all UI components
// are properly initialized after parallel initialization completes.
func TestInitializeUIComponents_AllComponentsInitialized(t *testing.T) {
	world := NewWorld()
	settingsManager, err := NewSettingsManager()
	if err != nil {
		t.Fatalf("Failed to create settings manager: %v", err)
	}
	screenWidth, screenHeight := 800, 600

	ui := initializeUIComponents(world, screenWidth, screenHeight, settingsManager)

	// Verify all UI components are non-nil
	if ui.inventoryUI == nil {
		t.Error("inventoryUI is nil after initialization")
	}
	if ui.questUI == nil {
		t.Error("questUI is nil after initialization")
	}
	if ui.characterUI == nil {
		t.Error("characterUI is nil after initialization")
	}
	if ui.skillsUI == nil {
		t.Error("skillsUI is nil after initialization")
	}
	if ui.mapUI == nil {
		t.Error("mapUI is nil after initialization")
	}
	if ui.settingsUI == nil {
		t.Error("settingsUI is nil after initialization")
	}
	if ui.mainMenuUI == nil {
		t.Error("mainMenuUI is nil after initialization")
	}
	if ui.singlePlayerMenu == nil {
		t.Error("singlePlayerMenu is nil after initialization")
	}
	if ui.genreSelectionMenu == nil {
		t.Error("genreSelectionMenu is nil after initialization")
	}
	if ui.multiplayerMenu == nil {
		t.Error("multiplayerMenu is nil after initialization")
	}
	if ui.serverAddressInput == nil {
		t.Error("serverAddressInput is nil after initialization")
	}
	if ui.characterCreation == nil {
		t.Error("characterCreation is nil after initialization")
	}
	if ui.galleryUI == nil {
		t.Error("galleryUI is nil after initialization")
	}
	if ui.housingUI == nil {
		t.Error("housingUI is nil after initialization")
	}
	if ui.guildUI == nil {
		t.Error("guildUI is nil after initialization")
	}
}

// TestInitializeUIComponents_ScreenDimensions verifies that UI components
// receive correct screen dimensions from the initialization function.
func TestInitializeUIComponents_ScreenDimensions(t *testing.T) {
	world := NewWorld()
	settingsManager, err := NewSettingsManager()
	if err != nil {
		t.Fatalf("Failed to create settings manager: %v", err)
	}
	screenWidth, screenHeight := 1024, 768

	ui := initializeUIComponents(world, screenWidth, screenHeight, settingsManager)

	// Check that screen dimensions are properly set on components that expose them
	if ui.mainMenuUI.screenWidth != screenWidth {
		t.Errorf("mainMenuUI screenWidth: got %d, want %d", ui.mainMenuUI.screenWidth, screenWidth)
	}
	if ui.mainMenuUI.screenHeight != screenHeight {
		t.Errorf("mainMenuUI screenHeight: got %d, want %d", ui.mainMenuUI.screenHeight, screenHeight)
	}

	if ui.settingsUI.screenWidth != screenWidth {
		t.Errorf("settingsUI screenWidth: got %d, want %d", ui.settingsUI.screenWidth, screenWidth)
	}
	if ui.settingsUI.screenHeight != screenHeight {
		t.Errorf("settingsUI screenHeight: got %d, want %d", ui.settingsUI.screenHeight, screenHeight)
	}
}

// TestInitializeUIComponents_WorldReference verifies that UI components
// that depend on the world receive the correct reference.
func TestInitializeUIComponents_WorldReference(t *testing.T) {
	world := NewWorld()
	settingsManager, err := NewSettingsManager()
	if err != nil {
		t.Fatalf("Failed to create settings manager: %v", err)
	}
	screenWidth, screenHeight := 800, 600

	ui := initializeUIComponents(world, screenWidth, screenHeight, settingsManager)

	// Verify world-dependent components have correct world reference
	if ui.inventoryUI.world != world {
		t.Error("inventoryUI has incorrect world reference")
	}
	if ui.questUI.world != world {
		t.Error("questUI has incorrect world reference")
	}
	if ui.characterUI.world != world {
		t.Error("characterUI has incorrect world reference")
	}
	if ui.skillsUI.world != world {
		t.Error("skillsUI has incorrect world reference")
	}
	if ui.mapUI.world != world {
		t.Error("mapUI has incorrect world reference")
	}
	if ui.guildUI.world != world {
		t.Error("guildUI has incorrect world reference")
	}
}

// TestInitializeUIComponents_SettingsManager verifies that settings UI
// receives the correct settings manager reference.
func TestInitializeUIComponents_SettingsManager(t *testing.T) {
	world := NewWorld()
	settingsManager, err := NewSettingsManager()
	if err != nil {
		t.Fatalf("Failed to create settings manager: %v", err)
	}
	screenWidth, screenHeight := 800, 600

	ui := initializeUIComponents(world, screenWidth, screenHeight, settingsManager)

	if ui.settingsUI.settingsManager != settingsManager {
		t.Error("settingsUI has incorrect settingsManager reference")
	}
}

// TestInitializeUIComponents_Deterministic verifies that multiple initializations
// produce consistent results (no race conditions).
func TestInitializeUIComponents_Deterministic(t *testing.T) {
	world := NewWorld()
	settingsManager, err := NewSettingsManager()
	if err != nil {
		t.Fatalf("Failed to create settings manager: %v", err)
	}
	screenWidth, screenHeight := 800, 600

	// Run initialization 10 times to check for race conditions
	for i := 0; i < 10; i++ {
		ui := initializeUIComponents(world, screenWidth, screenHeight, settingsManager)

		// Verify all components are initialized on each iteration
		if ui.inventoryUI == nil || ui.questUI == nil || ui.characterUI == nil ||
			ui.skillsUI == nil || ui.mapUI == nil || ui.settingsUI == nil ||
			ui.mainMenuUI == nil || ui.singlePlayerMenu == nil || ui.genreSelectionMenu == nil ||
			ui.multiplayerMenu == nil || ui.serverAddressInput == nil || ui.characterCreation == nil ||
			ui.galleryUI == nil || ui.housingUI == nil || ui.guildUI == nil {
			t.Errorf("Iteration %d: one or more UI components are nil", i)
		}
	}
}

// TestInitializeUIComponents_ConcurrentCalls verifies thread safety when
// initializeUIComponents is called concurrently (stress test).
func TestInitializeUIComponents_ConcurrentCalls(t *testing.T) {
	world := NewWorld()
	settingsManager, err := NewSettingsManager()
	if err != nil {
		t.Fatalf("Failed to create settings manager: %v", err)
	}
	screenWidth, screenHeight := 800, 600

	// Run 10 concurrent initializations
	results := make(chan *uiComponents, 10)
	for i := 0; i < 10; i++ {
		go func() {
			ui := initializeUIComponents(world, screenWidth, screenHeight, settingsManager)
			results <- ui
		}()
	}

	// Collect and verify all results
	for i := 0; i < 10; i++ {
		ui := <-results
		if ui.inventoryUI == nil || ui.questUI == nil || ui.characterUI == nil ||
			ui.skillsUI == nil || ui.mapUI == nil || ui.settingsUI == nil ||
			ui.mainMenuUI == nil || ui.singlePlayerMenu == nil || ui.genreSelectionMenu == nil ||
			ui.multiplayerMenu == nil || ui.serverAddressInput == nil || ui.characterCreation == nil ||
			ui.galleryUI == nil || ui.housingUI == nil || ui.guildUI == nil {
			t.Errorf("Concurrent iteration %d: one or more UI components are nil", i)
		}
	}
}
