package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestSettingsOptionString(t *testing.T) {
	tests := []struct {
		option   SettingsOption
		expected string
	}{
		{SettingsOptionMasterVolume, "Master Volume"},
		{SettingsOptionMusicVolume, "Music Volume"},
		{SettingsOptionSFXVolume, "SFX Volume"},
		{SettingsOptionGraphicsQuality, "Graphics Quality"},
		{SettingsOptionVSync, "VSync"},
		{SettingsOptionShowFPS, "Show FPS"},
		{SettingsOptionFullscreen, "Fullscreen"},
		{SettingsOptionBack, "Back"},
		{SettingsOption(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.option.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestNewSettingsUI(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	ui := NewSettingsUI(1280, 720, sm)

	if ui == nil {
		t.Fatal("Expected non-nil SettingsUI")
	}

	if ui.screenWidth != 1280 {
		t.Errorf("Expected screenWidth 1280, got %d", ui.screenWidth)
	}

	if ui.screenHeight != 720 {
		t.Errorf("Expected screenHeight 720, got %d", ui.screenHeight)
	}

	if ui.selectedIdx != 0 {
		t.Errorf("Expected selectedIdx 0, got %d", ui.selectedIdx)
	}

	if len(ui.options) != 8 {
		t.Errorf("Expected 8 options, got %d", len(ui.options))
	}

	if ui.visible {
		t.Error("Expected UI to be hidden initially")
	}
}

func TestSettingsUI_ShowHide(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	ui := NewSettingsUI(1280, 720, sm)

	// Initially hidden
	if ui.IsVisible() {
		t.Error("Expected UI to be hidden initially")
	}

	// Show UI
	ui.Show()
	if !ui.IsVisible() {
		t.Error("Expected UI to be visible after Show()")
	}

	// Hide UI
	ui.Hide()
	if ui.IsVisible() {
		t.Error("Expected UI to be hidden after Hide()")
	}
}

func TestSettingsUI_SetBackCallback(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	ui := NewSettingsUI(1280, 720, sm)

	callbackCalled := false
	ui.SetBackCallback(func() {
		callbackCalled = true
	})

	// Simulate back selection
	ui.Show()
	ui.selectedIdx = 7 // Back option
	ui.activateOption(SettingsOptionBack)

	if !callbackCalled {
		t.Error("Expected callback to be called")
	}

	if ui.IsVisible() {
		t.Error("Expected UI to be hidden after back")
	}
}

func TestSettingsUI_DecreaseValue(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	ui := NewSettingsUI(1280, 720, sm)
	ui.Show()

	tests := []struct {
		name         string
		option       SettingsOption
		initialValue func() interface{}
		expectedFunc func() interface{}
	}{
		{
			name:   "decrease master volume",
			option: SettingsOptionMasterVolume,
			initialValue: func() interface{} {
				ui.currentSettings.MasterVolume = 0.5
				return 0.5
			},
			expectedFunc: func() interface{} {
				return 0.4 // Decreased by 0.1
			},
		},
		{
			name:   "decrease master volume at minimum",
			option: SettingsOptionMasterVolume,
			initialValue: func() interface{} {
				ui.currentSettings.MasterVolume = 0.0
				return 0.0
			},
			expectedFunc: func() interface{} {
				return 0.0 // Can't go below 0
			},
		},
		{
			name:   "cycle graphics quality down",
			option: SettingsOptionGraphicsQuality,
			initialValue: func() interface{} {
				ui.currentSettings.GraphicsQuality = "high"
				return "high"
			},
			expectedFunc: func() interface{} {
				return "medium"
			},
		},
		{
			name:   "toggle VSync",
			option: SettingsOptionVSync,
			initialValue: func() interface{} {
				ui.currentSettings.VSync = true
				return true
			},
			expectedFunc: func() interface{} {
				return false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.initialValue()
			ui.decreaseValue(tt.option)

			var actual interface{}
			switch tt.option {
			case SettingsOptionMasterVolume:
				actual = ui.currentSettings.MasterVolume
			case SettingsOptionGraphicsQuality:
				actual = ui.currentSettings.GraphicsQuality
			case SettingsOptionVSync:
				actual = ui.currentSettings.VSync
			}

			expected := tt.expectedFunc()

			// Handle float comparison
			if floatActual, ok := actual.(float64); ok {
				floatExpected := expected.(float64)
				if !settingsFloatEqual(floatActual, floatExpected, 0.01) {
					t.Errorf("Expected %v, got %v", expected, actual)
				}
			} else if actual != expected {
				t.Errorf("Expected %v, got %v", expected, actual)
			}
		})
	}
}

func TestSettingsUI_IncreaseValue(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	ui := NewSettingsUI(1280, 720, sm)
	ui.Show()

	tests := []struct {
		name         string
		option       SettingsOption
		initialValue func() interface{}
		expectedFunc func() interface{}
	}{
		{
			name:   "increase music volume",
			option: SettingsOptionMusicVolume,
			initialValue: func() interface{} {
				ui.currentSettings.MusicVolume = 0.5
				return 0.5
			},
			expectedFunc: func() interface{} {
				return 0.6 // Increased by 0.1
			},
		},
		{
			name:   "increase music volume at maximum",
			option: SettingsOptionMusicVolume,
			initialValue: func() interface{} {
				ui.currentSettings.MusicVolume = 1.0
				return 1.0
			},
			expectedFunc: func() interface{} {
				return 1.0 // Can't go above 1.0
			},
		},
		{
			name:   "cycle graphics quality up",
			option: SettingsOptionGraphicsQuality,
			initialValue: func() interface{} {
				ui.currentSettings.GraphicsQuality = "low"
				return "low"
			},
			expectedFunc: func() interface{} {
				return "medium"
			},
		},
		{
			name:   "toggle ShowFPS",
			option: SettingsOptionShowFPS,
			initialValue: func() interface{} {
				ui.currentSettings.ShowFPS = false
				return false
			},
			expectedFunc: func() interface{} {
				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.initialValue()
			ui.increaseValue(tt.option)

			var actual interface{}
			switch tt.option {
			case SettingsOptionMusicVolume:
				actual = ui.currentSettings.MusicVolume
			case SettingsOptionGraphicsQuality:
				actual = ui.currentSettings.GraphicsQuality
			case SettingsOptionShowFPS:
				actual = ui.currentSettings.ShowFPS
			}

			expected := tt.expectedFunc()

			// Handle float comparison
			if floatActual, ok := actual.(float64); ok {
				floatExpected := expected.(float64)
				if !settingsFloatEqual(floatActual, floatExpected, 0.01) {
					t.Errorf("Expected %v, got %v", expected, actual)
				}
			} else if actual != expected {
				t.Errorf("Expected %v, got %v", expected, actual)
			}
		})
	}
}

func TestSettingsUI_ActivateOption(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	ui := NewSettingsUI(1280, 720, sm)
	ui.Show()

	tests := []struct {
		name       string
		option     SettingsOption
		setupFunc  func()
		verifyFunc func(t *testing.T)
	}{
		{
			name:   "activate back",
			option: SettingsOptionBack,
			setupFunc: func() {
				ui.Show()
			},
			verifyFunc: func(t *testing.T) {
				if ui.IsVisible() {
					t.Error("Expected UI to be hidden after activating back")
				}
			},
		},
		{
			name:   "activate VSync toggle",
			option: SettingsOptionVSync,
			setupFunc: func() {
				ui.currentSettings.VSync = false
			},
			verifyFunc: func(t *testing.T) {
				if !ui.currentSettings.VSync {
					t.Error("Expected VSync to be toggled on")
				}
			},
		},
		{
			name:   "activate fullscreen toggle",
			option: SettingsOptionFullscreen,
			setupFunc: func() {
				ui.currentSettings.Fullscreen = false
			},
			verifyFunc: func(t *testing.T) {
				if !ui.currentSettings.Fullscreen {
					t.Error("Expected Fullscreen to be toggled on")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()
			ui.activateOption(tt.option)
			tt.verifyFunc(t)
		})
	}
}

func TestSettingsUI_GetValueString(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	ui := NewSettingsUI(1280, 720, sm)

	tests := []struct {
		name     string
		option   SettingsOption
		setup    func()
		expected string
	}{
		{
			name:   "master volume 50%",
			option: SettingsOptionMasterVolume,
			setup: func() {
				ui.currentSettings.MasterVolume = 0.5
			},
			expected: "50%",
		},
		{
			name:   "graphics quality high",
			option: SettingsOptionGraphicsQuality,
			setup: func() {
				ui.currentSettings.GraphicsQuality = "high"
			},
			expected: "high",
		},
		{
			name:   "VSync on",
			option: SettingsOptionVSync,
			setup: func() {
				ui.currentSettings.VSync = true
			},
			expected: "ON",
		},
		{
			name:   "VSync off",
			option: SettingsOptionVSync,
			setup: func() {
				ui.currentSettings.VSync = false
			},
			expected: "OFF",
		},
		{
			name:     "back option (empty)",
			option:   SettingsOptionBack,
			setup:    func() {},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result := ui.getValueString(tt.option)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestSettingsUI_IsAdjustable(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	ui := NewSettingsUI(1280, 720, sm)

	// Most options are adjustable
	adjustableOptions := []SettingsOption{
		SettingsOptionMasterVolume,
		SettingsOptionMusicVolume,
		SettingsOptionSFXVolume,
		SettingsOptionGraphicsQuality,
		SettingsOptionVSync,
		SettingsOptionShowFPS,
		SettingsOptionFullscreen,
	}

	for _, option := range adjustableOptions {
		if !ui.isAdjustable(option) {
			t.Errorf("Expected %s to be adjustable", option.String())
		}
	}

	// Back is not adjustable
	if ui.isAdjustable(SettingsOptionBack) {
		t.Error("Expected Back to not be adjustable")
	}
}

func TestSettingsUI_SaveOnHide(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	ui := NewSettingsUI(1280, 720, sm)
	ui.Show()

	// Modify a setting
	ui.currentSettings.MasterVolume = 0.9

	// Hide (should save)
	ui.Hide()

	// Verify persisted
	sm2 := &SettingsManager{
		settingsPath: sm.settingsPath,
	}
	sm2.LoadSettings()

	if !settingsFloatEqual(sm2.GetSettings().MasterVolume, 0.9, 0.01) {
		t.Errorf("Expected MasterVolume 0.9, got %f", sm2.GetSettings().MasterVolume)
	}
}

func TestSettingsUI_GetCurrentSettings(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	ui := NewSettingsUI(1280, 720, sm)
	ui.Show()

	// Modify settings
	ui.currentSettings.MasterVolume = 0.8
	ui.currentSettings.GraphicsQuality = "low"

	// Get current settings
	current := ui.GetCurrentSettings()

	if !settingsFloatEqual(current.MasterVolume, 0.8, 0.01) {
		t.Errorf("Expected MasterVolume 0.8, got %f", current.MasterVolume)
	}

	if current.GraphicsQuality != "low" {
		t.Errorf("Expected GraphicsQuality 'low', got %s", current.GraphicsQuality)
	}
}

func TestSettingsUI_Draw_NoEbitenRuntime(t *testing.T) {
	// This test verifies the UI doesn't crash when Draw is called,
	// but we can't test actual rendering without Ebiten runtime.
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	ui := NewSettingsUI(1280, 720, sm)

	// Draw when hidden should be no-op
	ui.Draw(nil)

	// Show and draw (will skip without valid screen)
	ui.Show()
	ui.Draw(nil)

	// Test passes if no panic occurs
}

// Helper function for float comparison with tolerance
func settingsFloatEqual(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < tolerance
}

// Benchmark settings UI operations
func BenchmarkSettingsUI_DecreaseValue(b *testing.B) {
	tempDir := b.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	ui := NewSettingsUI(1280, 720, sm)
	ui.Show()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ui.decreaseValue(SettingsOptionMasterVolume)
	}
}

func BenchmarkSettingsUI_IncreaseValue(b *testing.B) {
	tempDir := b.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	ui := NewSettingsUI(1280, 720, sm)
	ui.Show()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ui.increaseValue(SettingsOptionMusicVolume)
	}
}

func BenchmarkSettingsUI_GetValueString(b *testing.B) {
	tempDir := b.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	ui := NewSettingsUI(1280, 720, sm)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ui.getValueString(SettingsOptionMasterVolume)
	}
}

// Test that Update returns false when hidden
func TestSettingsUI_Update_WhenHidden(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	ui := NewSettingsUI(1280, 720, sm)

	// Update when hidden should return false and do nothing
	result := ui.Update()
	if result {
		t.Error("Expected Update to return false when hidden")
	}
}

// Test Draw when screen is nil (should not crash)
func TestSettingsUI_Draw_NilScreen(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	ui := NewSettingsUI(1280, 720, sm)
	ui.Show()

	// Should not panic with nil screen
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Draw panicked with nil screen: %v", r)
		}
	}()

	ui.Draw(nil)
}

// Test creating settings image (Ebiten-dependent, will skip in CI)
func TestSettingsUI_Draw_WithImage(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping Ebiten-dependent test in CI")
	}

	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	ui := NewSettingsUI(1280, 720, sm)
	ui.Show()

	screen := ebiten.NewImage(1280, 720)
	ui.Draw(screen)

	// Test passes if no panic occurs
}
