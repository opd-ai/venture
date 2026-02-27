package engine

import (
	"strings"
	"testing"
)

// TestMenuTypeSettings tests that MenuTypeSettings constant is defined correctly
func TestMenuTypeSettings(t *testing.T) {
	tests := []struct {
		name     string
		menuType MenuType
		want     int
	}{
		{"None", MenuTypeNone, 0},
		{"Main", MenuTypeMain, 1},
		{"Save", MenuTypeSave, 2},
		{"Load", MenuTypeLoad, 3},
		{"Confirm", MenuTypeConfirm, 4},
		{"Settings", MenuTypeSettings, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.menuType) != tt.want {
				t.Errorf("MenuType %s = %d, want %d", tt.name, tt.menuType, tt.want)
			}
		})
	}
}

// TestGetMenuTitleSettings tests that getMenuTitle returns correct title for settings menu
func TestGetMenuTitleSettings(t *testing.T) {
	tests := []struct {
		name     string
		menuType MenuType
		want     string
	}{
		{"Main", MenuTypeMain, "GAME MENU"},
		{"Save", MenuTypeSave, "SAVE GAME"},
		{"Load", MenuTypeLoad, "LOAD GAME"},
		{"Confirm", MenuTypeConfirm, "CONFIRM"},
		{"Settings", MenuTypeSettings, "SETTINGS"},
		{"None", MenuTypeNone, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getMenuTitle(tt.menuType)
			if got != tt.want {
				t.Errorf("getMenuTitle(%v) = %q, want %q", tt.menuType, got, tt.want)
			}
		})
	}
}

// TestSetSettingsManager tests setting the settings manager
func TestSetSettingsManager(t *testing.T) {
	world := NewWorld()
	ms, err := NewEbitenMenuSystem(world, 800, 600, "./test_saves")
	if err != nil {
		t.Fatalf("NewEbitenMenuSystem failed: %v", err)
	}

	if ms.settingsManager != nil {
		t.Error("settingsManager should be nil initially")
	}

	settingsManager := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: "/tmp/test_settings.json",
	}

	ms.SetSettingsManager(settingsManager)

	if ms.settingsManager == nil {
		t.Error("settingsManager should not be nil after SetSettingsManager")
	}

	if ms.settingsManager != settingsManager {
		t.Error("settingsManager should be the same instance")
	}
}

// TestSetApplySettingsCallback tests setting the apply settings callback
func TestSetApplySettingsCallback(t *testing.T) {
	world := NewWorld()
	ms, err := NewEbitenMenuSystem(world, 800, 600, "./test_saves")
	if err != nil {
		t.Fatalf("NewEbitenMenuSystem failed: %v", err)
	}

	if ms.onApplySettings != nil {
		t.Error("onApplySettings should be nil initially")
	}

	callbackCalled := false
	callback := func(GameSettings) error {
		callbackCalled = true
		return nil
	}

	ms.SetApplySettingsCallback(callback)

	if ms.onApplySettings == nil {
		t.Error("onApplySettings should not be nil after SetApplySettingsCallback")
	}

	// Test callback execution
	ms.onApplySettings(DefaultSettings())
	if !callbackCalled {
		t.Error("callback should have been called")
	}
}

// TestBuildSettingsMenu tests building the settings menu
func TestBuildSettingsMenu(t *testing.T) {
	world := NewWorld()
	ms, err := NewEbitenMenuSystem(world, 800, 600, "./test_saves")
	if err != nil {
		t.Fatalf("NewEbitenMenuSystem failed: %v", err)
	}

	// Set up settings manager
	settingsManager := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: "/tmp/test_settings.json",
	}
	ms.SetSettingsManager(settingsManager)

	// Create menu component
	menu := &MenuComponent{
		Active:      true,
		CurrentMenu: MenuTypeSettings,
	}

	ms.buildSettingsMenu(menu)

	// Should have 9 items: 4 resolutions + 3 graphics qualities + VSync + Back
	expectedItemCount := 9
	if len(menu.Items) != expectedItemCount {
		t.Errorf("Expected %d menu items, got %d", expectedItemCount, len(menu.Items))
	}

	// Check that menu items are created
	if menu.SelectedIndex != 0 {
		t.Errorf("Expected SelectedIndex to be 0, got %d", menu.SelectedIndex)
	}

	// Verify resolution items exist
	resolutionLabels := []string{"HD (1280x720)", "Full HD (1920x1080)", "QHD (2560x1440)", "4K UHD (3840x2160)"}
	for i, label := range resolutionLabels {
		if i >= len(menu.Items) {
			t.Errorf("Not enough menu items for resolution %s", label)
			continue
		}
		if !strings.Contains(menu.Items[i].Label, label) {
			t.Errorf("Expected menu item %d to contain %q, got %q", i, label, menu.Items[i].Label)
		}
	}

	// Verify graphics quality items exist
	qualityLabels := []string{"Graphics: low", "Graphics: medium", "Graphics: high"}
	for i, label := range qualityLabels {
		itemIndex := 4 + i // After 4 resolution items
		if itemIndex >= len(menu.Items) {
			t.Errorf("Not enough menu items for quality %s", label)
			continue
		}
		if !strings.Contains(menu.Items[itemIndex].Label, label) {
			t.Errorf("Expected menu item %d to contain %q, got %q", itemIndex, label, menu.Items[itemIndex].Label)
		}
	}

	// Verify VSync item exists
	if !strings.Contains(menu.Items[7].Label, "VSync") {
		t.Errorf("Expected menu item 7 to be VSync, got %q", menu.Items[7].Label)
	}

	// Verify Back button exists
	if menu.Items[8].Label != "< Back" {
		t.Errorf("Expected last menu item to be '< Back', got %q", menu.Items[8].Label)
	}
}

// TestBuildSettingsMenuWithoutManager tests building settings menu without settings manager
func TestBuildSettingsMenuWithoutManager(t *testing.T) {
	world := NewWorld()
	ms, err := NewEbitenMenuSystem(world, 800, 600, "./test_saves")
	if err != nil {
		t.Fatalf("NewEbitenMenuSystem failed: %v", err)
	}

	// Don't set settings manager - should use defaults
	menu := &MenuComponent{
		Active:      true,
		CurrentMenu: MenuTypeSettings,
	}

	ms.buildSettingsMenu(menu)

	// Should still build menu with default settings
	if len(menu.Items) != 9 {
		t.Errorf("Expected 9 menu items even without settings manager, got %d", len(menu.Items))
	}
}

// TestCreateResolutionMenuItem tests resolution menu item creation
func TestCreateResolutionMenuItem(t *testing.T) {
	world := NewWorld()
	ms, err := NewEbitenMenuSystem(world, 800, 600, "./test_saves")
	if err != nil {
		t.Fatalf("NewEbitenMenuSystem failed: %v", err)
	}

	settingsManager := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: "/tmp/test_settings.json",
	}
	ms.SetSettingsManager(settingsManager)

	menu := &MenuComponent{
		Active:      true,
		CurrentMenu: MenuTypeSettings,
	}

	currentSettings := DefaultSettings()

	tests := []struct {
		name        string
		width       int
		height      int
		label       string
		wantMarked  bool
		wantEnabled bool
	}{
		{"Current resolution", 1280, 720, "HD (1280x720)", true, true},
		{"Different resolution", 1920, 1080, "Full HD (1920x1080)", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set current resolution
			currentSettings.WindowWidth = 1280
			currentSettings.WindowHeight = 720

			item := ms.createResolutionMenuItem(&currentSettings, tt.width, tt.height, tt.label, menu)

			if item.Enabled != tt.wantEnabled {
				t.Errorf("Enabled = %v, want %v", item.Enabled, tt.wantEnabled)
			}

			// Check if marked with asterisk
			isMarked := strings.HasPrefix(item.Label, "* ")
			if isMarked != tt.wantMarked {
				t.Errorf("Label marked = %v, want %v (label: %q)", isMarked, tt.wantMarked, item.Label)
			}
		})
	}
}

// TestCreateGraphicsQualityMenuItem tests graphics quality menu item creation
func TestCreateGraphicsQualityMenuItem(t *testing.T) {
	world := NewWorld()
	ms, err := NewEbitenMenuSystem(world, 800, 600, "./test_saves")
	if err != nil {
		t.Fatalf("NewEbitenMenuSystem failed: %v", err)
	}

	settingsManager := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: "/tmp/test_settings.json",
	}
	ms.SetSettingsManager(settingsManager)

	menu := &MenuComponent{
		Active:      true,
		CurrentMenu: MenuTypeSettings,
	}

	currentSettings := DefaultSettings()
	currentSettings.GraphicsQuality = "medium"

	tests := []struct {
		name       string
		quality    string
		wantMarked bool
	}{
		{"Low quality", "low", false},
		{"Medium quality (current)", "medium", true},
		{"High quality", "high", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := ms.createGraphicsQualityMenuItem(&currentSettings, tt.quality, menu)

			if !item.Enabled {
				t.Error("Graphics quality item should always be enabled")
			}

			isMarked := strings.HasPrefix(item.Label, "* ")
			if isMarked != tt.wantMarked {
				t.Errorf("Label marked = %v, want %v (label: %q)", isMarked, tt.wantMarked, item.Label)
			}

			if !strings.Contains(item.Label, tt.quality) {
				t.Errorf("Label should contain quality %q, got %q", tt.quality, item.Label)
			}
		})
	}
}

// TestCreateVSyncToggleMenuItem tests VSync toggle menu item creation
func TestCreateVSyncToggleMenuItem(t *testing.T) {
	world := NewWorld()
	ms, err := NewEbitenMenuSystem(world, 800, 600, "./test_saves")
	if err != nil {
		t.Fatalf("NewEbitenMenuSystem failed: %v", err)
	}

	settingsManager := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: "/tmp/test_settings.json",
	}
	ms.SetSettingsManager(settingsManager)

	menu := &MenuComponent{
		Active:      true,
		CurrentMenu: MenuTypeSettings,
	}

	tests := []struct {
		name        string
		vsyncStatus bool
		wantLabel   string
	}{
		{"VSync enabled", true, "VSync: On"},
		{"VSync disabled", false, "VSync: Off"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentSettings := DefaultSettings()
			currentSettings.VSync = tt.vsyncStatus

			item := ms.createVSyncToggleMenuItem(&currentSettings, menu)

			if !item.Enabled {
				t.Error("VSync toggle should always be enabled")
			}

			if item.Label != tt.wantLabel {
				t.Errorf("Label = %q, want %q", item.Label, tt.wantLabel)
			}
		})
	}
}

// TestRebuildMenuSettings tests that rebuildMenu handles MenuTypeSettings
func TestRebuildMenuSettings(t *testing.T) {
	world := NewWorld()
	ms, err := NewEbitenMenuSystem(world, 800, 600, "./test_saves")
	if err != nil {
		t.Fatalf("NewEbitenMenuSystem failed: %v", err)
	}

	settingsManager := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: "/tmp/test_settings.json",
	}
	ms.SetSettingsManager(settingsManager)

	menu := &MenuComponent{
		Active:      true,
		CurrentMenu: MenuTypeSettings,
	}

	ms.rebuildMenu(menu)

	// Should have built settings menu
	if len(menu.Items) != 9 {
		t.Errorf("Expected 9 menu items after rebuild, got %d", len(menu.Items))
	}

	if menu.CurrentMenu != MenuTypeSettings {
		t.Errorf("CurrentMenu should still be MenuTypeSettings, got %v", menu.CurrentMenu)
	}
}

// TestMainMenuHasSettings tests that main menu includes Settings option
func TestMainMenuHasSettings(t *testing.T) {
	world := NewWorld()
	ms, err := NewEbitenMenuSystem(world, 800, 600, "./test_saves")
	if err != nil {
		t.Fatalf("NewEbitenMenuSystem failed: %v", err)
	}

	menu := &MenuComponent{
		Active:      true,
		CurrentMenu: MenuTypeMain,
	}

	ms.buildMainMenu(menu)

	// Find Settings menu item
	foundSettings := false
	for _, item := range menu.Items {
		if item.Label == "Settings" {
			foundSettings = true
			if !item.Enabled {
				t.Error("Settings menu item should be enabled")
			}
			break
		}
	}

	if !foundSettings {
		t.Error("Main menu should contain Settings item")
	}

	// Verify Settings is second item (after Resume Game)
	if len(menu.Items) < 2 {
		t.Fatal("Main menu should have at least 2 items")
	}
	if menu.Items[1].Label != "Settings" {
		t.Errorf("Second menu item should be Settings, got %q", menu.Items[1].Label)
	}
}
