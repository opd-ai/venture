package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultSettings(t *testing.T) {
	settings := DefaultSettings()

	// Test audio defaults
	if settings.MasterVolume != 0.7 {
		t.Errorf("Expected MasterVolume 0.7, got %f", settings.MasterVolume)
	}
	if settings.MusicVolume != 0.6 {
		t.Errorf("Expected MusicVolume 0.6, got %f", settings.MusicVolume)
	}
	if settings.SFXVolume != 0.8 {
		t.Errorf("Expected SFXVolume 0.8, got %f", settings.SFXVolume)
	}

	// Test display defaults
	if settings.WindowWidth != 1280 {
		t.Errorf("Expected WindowWidth 1280, got %d", settings.WindowWidth)
	}
	if settings.WindowHeight != 720 {
		t.Errorf("Expected WindowHeight 720, got %d", settings.WindowHeight)
	}
	if settings.Fullscreen != false {
		t.Error("Expected Fullscreen to be false")
	}

	// Test graphics defaults
	if settings.GraphicsQuality != "medium" {
		t.Errorf("Expected GraphicsQuality 'medium', got %s", settings.GraphicsQuality)
	}
	if settings.VSync != true {
		t.Error("Expected VSync to be true")
	}
	if settings.ShowFPS != false {
		t.Error("Expected ShowFPS to be false")
	}

	// Test gameplay defaults
	if settings.ShowTutorials != true {
		t.Error("Expected ShowTutorials to be true")
	}

	// Test QoL defaults
	if settings.QoLAutoLoot != true {
		t.Error("Expected QoLAutoLoot to be true")
	}
	if settings.QoLAutoLootRadius != 7.0 {
		t.Errorf("Expected QoLAutoLootRadius 7.0, got %f", settings.QoLAutoLootRadius)
	}
	if settings.QoLMountWhistle != true {
		t.Error("Expected QoLMountWhistle to be true")
	}
	if settings.QoLRecipeTracking != true {
		t.Error("Expected QoLRecipeTracking to be true")
	}
	if settings.QoLSortPreset != "type" {
		t.Errorf("Expected QoLSortPreset 'type', got %s", settings.QoLSortPreset)
	}
}

func TestGameSettingsValidate(t *testing.T) {
	tests := []struct {
		name      string
		settings  GameSettings
		shouldFix bool
		checkFunc func(*testing.T, GameSettings)
	}{
		{
			name:      "valid settings unchanged",
			settings:  DefaultSettings(),
			shouldFix: false,
			checkFunc: func(t *testing.T, s GameSettings) {
				if s.MasterVolume != 0.7 {
					t.Errorf("Valid setting was incorrectly changed")
				}
			},
		},
		{
			name: "negative master volume corrected",
			settings: GameSettings{
				MasterVolume:    -0.5,
				MusicVolume:     0.6,
				SFXVolume:       0.8,
				WindowWidth:     1280,
				WindowHeight:    720,
				GraphicsQuality: "medium",
			},
			shouldFix: true,
			checkFunc: func(t *testing.T, s GameSettings) {
				if s.MasterVolume != 0.7 {
					t.Errorf("Expected corrected MasterVolume 0.7, got %f", s.MasterVolume)
				}
			},
		},
		{
			name: "excessive music volume corrected",
			settings: GameSettings{
				MasterVolume:    0.7,
				MusicVolume:     1.5,
				SFXVolume:       0.8,
				WindowWidth:     1280,
				WindowHeight:    720,
				GraphicsQuality: "medium",
			},
			shouldFix: true,
			checkFunc: func(t *testing.T, s GameSettings) {
				if s.MusicVolume != 0.6 {
					t.Errorf("Expected corrected MusicVolume 0.6, got %f", s.MusicVolume)
				}
			},
		},
		{
			name: "invalid window dimensions corrected",
			settings: GameSettings{
				MasterVolume:    0.7,
				MusicVolume:     0.6,
				SFXVolume:       0.8,
				WindowWidth:     100,
				WindowHeight:    50,
				GraphicsQuality: "medium",
			},
			shouldFix: true,
			checkFunc: func(t *testing.T, s GameSettings) {
				if s.WindowWidth != 1280 {
					t.Errorf("Expected corrected WindowWidth 1280, got %d", s.WindowWidth)
				}
				if s.WindowHeight != 720 {
					t.Errorf("Expected corrected WindowHeight 720, got %d", s.WindowHeight)
				}
			},
		},
		{
			name: "excessive window dimensions corrected",
			settings: GameSettings{
				MasterVolume:    0.7,
				MusicVolume:     0.6,
				SFXVolume:       0.8,
				WindowWidth:     10000,
				WindowHeight:    10000,
				GraphicsQuality: "medium",
			},
			shouldFix: true,
			checkFunc: func(t *testing.T, s GameSettings) {
				if s.WindowWidth != 1280 {
					t.Errorf("Expected corrected WindowWidth 1280, got %d", s.WindowWidth)
				}
				if s.WindowHeight != 720 {
					t.Errorf("Expected corrected WindowHeight 720, got %d", s.WindowHeight)
				}
			},
		},
		{
			name: "invalid graphics quality corrected",
			settings: GameSettings{
				MasterVolume:    0.7,
				MusicVolume:     0.6,
				SFXVolume:       0.8,
				WindowWidth:     1280,
				WindowHeight:    720,
				GraphicsQuality: "ultra_mega",
			},
			shouldFix: true,
			checkFunc: func(t *testing.T, s GameSettings) {
				if s.GraphicsQuality != "medium" {
					t.Errorf("Expected corrected GraphicsQuality 'medium', got %s", s.GraphicsQuality)
				}
			},
		},
		{
			name: "multiple invalid settings corrected",
			settings: GameSettings{
				MasterVolume:    2.0,
				MusicVolume:     -1.0,
				SFXVolume:       10.0,
				WindowWidth:     100,
				WindowHeight:    50,
				GraphicsQuality: "invalid",
			},
			shouldFix: true,
			checkFunc: func(t *testing.T, s GameSettings) {
				if s.MasterVolume != 0.7 {
					t.Errorf("Expected corrected MasterVolume 0.7, got %f", s.MasterVolume)
				}
				if s.MusicVolume != 0.6 {
					t.Errorf("Expected corrected MusicVolume 0.6, got %f", s.MusicVolume)
				}
				if s.SFXVolume != 0.8 {
					t.Errorf("Expected corrected SFXVolume 0.8, got %f", s.SFXVolume)
				}
				if s.GraphicsQuality != "medium" {
					t.Errorf("Expected corrected GraphicsQuality 'medium', got %s", s.GraphicsQuality)
				}
			},
		},
		{
			name: "QoL auto-loot radius too low corrected",
			settings: GameSettings{
				MasterVolume:      0.7,
				MusicVolume:       0.6,
				SFXVolume:         0.8,
				WindowWidth:       1280,
				WindowHeight:      720,
				GraphicsQuality:   "medium",
				QoLAutoLootRadius: 3.0, // Below minimum of 5.0
			},
			shouldFix: true,
			checkFunc: func(t *testing.T, s GameSettings) {
				if s.QoLAutoLootRadius != 7.0 {
					t.Errorf("Expected corrected QoLAutoLootRadius 7.0, got %f", s.QoLAutoLootRadius)
				}
			},
		},
		{
			name: "QoL auto-loot radius too high corrected",
			settings: GameSettings{
				MasterVolume:      0.7,
				MusicVolume:       0.6,
				SFXVolume:         0.8,
				WindowWidth:       1280,
				WindowHeight:      720,
				GraphicsQuality:   "medium",
				QoLAutoLootRadius: 15.0, // Above maximum of 10.0
			},
			shouldFix: true,
			checkFunc: func(t *testing.T, s GameSettings) {
				if s.QoLAutoLootRadius != 7.0 {
					t.Errorf("Expected corrected QoLAutoLootRadius 7.0, got %f", s.QoLAutoLootRadius)
				}
			},
		},
		{
			name: "QoL sort preset invalid corrected",
			settings: GameSettings{
				MasterVolume:    0.7,
				MusicVolume:     0.6,
				SFXVolume:       0.8,
				WindowWidth:     1280,
				WindowHeight:    720,
				GraphicsQuality: "medium",
				QoLSortPreset:   "invalid_preset",
			},
			shouldFix: true,
			checkFunc: func(t *testing.T, s GameSettings) {
				if s.QoLSortPreset != "type" {
					t.Errorf("Expected corrected QoLSortPreset 'type', got %s", s.QoLSortPreset)
				}
			},
		},
		{
			name: "QoL settings valid unchanged",
			settings: GameSettings{
				MasterVolume:      0.7,
				MusicVolume:       0.6,
				SFXVolume:         0.8,
				WindowWidth:       1280,
				WindowHeight:      720,
				GraphicsQuality:   "medium",
				QoLAutoLootRadius: 8.0,
				QoLSortPreset:     "rarity",
			},
			shouldFix: false,
			checkFunc: func(t *testing.T, s GameSettings) {
				if s.QoLAutoLootRadius != 8.0 {
					t.Errorf("Valid QoLAutoLootRadius should not change, got %f", s.QoLAutoLootRadius)
				}
				if s.QoLSortPreset != "rarity" {
					t.Errorf("Valid QoLSortPreset should not change, got %s", s.QoLSortPreset)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			corrected := tt.settings.Validate()

			if corrected != tt.shouldFix {
				t.Errorf("Expected Validate to return %v, got %v", tt.shouldFix, corrected)
			}

			tt.checkFunc(t, tt.settings)
		})
	}
}

func TestNewSettingsManager(t *testing.T) {
	sm, err := NewSettingsManager()
	if err != nil {
		t.Fatalf("NewSettingsManager failed: %v", err)
	}

	if sm == nil {
		t.Fatal("Expected non-nil SettingsManager")
	}

	if sm.settingsPath == "" {
		t.Error("Expected non-empty settings path")
	}

	// Verify directory was created
	dir := filepath.Dir(sm.settingsPath)
	info, err := os.Stat(dir)
	if err != nil {
		t.Errorf("Settings directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("Settings path is not a directory")
	}
}

func TestSettingsManager_LoadSettings_FirstTime(t *testing.T) {
	// Create temporary settings directory
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     GameSettings{},
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	// Load settings (file doesn't exist)
	err := sm.LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings failed on first run: %v", err)
	}

	// Should have default settings
	if sm.settings.MasterVolume != 0.7 {
		t.Errorf("Expected default MasterVolume after first load, got %f", sm.settings.MasterVolume)
	}
}

func TestSettingsManager_SaveAndLoad(t *testing.T) {
	// Create temporary settings directory
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     GameSettings{},
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	// Create custom settings including QoL
	customSettings := GameSettings{
		MasterVolume:      0.5,
		MusicVolume:       0.3,
		SFXVolume:         0.9,
		WindowWidth:       1920,
		WindowHeight:      1080,
		Fullscreen:        true,
		GraphicsQuality:   "high",
		VSync:             false,
		ShowFPS:           true,
		ShowTutorials:     false,
		QoLAutoLoot:       false,
		QoLAutoLootRadius: 9.0,
		QoLMountWhistle:   false,
		QoLRecipeTracking: false,
		QoLSortPreset:     "value",
	}
	sm.settings = customSettings

	// Save settings
	err := sm.SaveSettings()
	if err != nil {
		t.Fatalf("SaveSettings failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(sm.settingsPath); err != nil {
		t.Errorf("Settings file not created: %v", err)
	}

	// Create new manager and load
	sm2 := &SettingsManager{
		settingsPath: sm.settingsPath,
	}
	err = sm2.LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings failed: %v", err)
	}

	// Verify loaded settings match saved
	if sm2.settings.MasterVolume != customSettings.MasterVolume {
		t.Errorf("MasterVolume mismatch: expected %f, got %f",
			customSettings.MasterVolume, sm2.settings.MasterVolume)
	}
	if sm2.settings.WindowWidth != customSettings.WindowWidth {
		t.Errorf("WindowWidth mismatch: expected %d, got %d",
			customSettings.WindowWidth, sm2.settings.WindowWidth)
	}
	if sm2.settings.GraphicsQuality != customSettings.GraphicsQuality {
		t.Errorf("GraphicsQuality mismatch: expected %s, got %s",
			customSettings.GraphicsQuality, sm2.settings.GraphicsQuality)
	}

	// Verify QoL settings persisted
	if sm2.settings.QoLAutoLoot != customSettings.QoLAutoLoot {
		t.Errorf("QoLAutoLoot mismatch: expected %v, got %v",
			customSettings.QoLAutoLoot, sm2.settings.QoLAutoLoot)
	}
	if sm2.settings.QoLAutoLootRadius != customSettings.QoLAutoLootRadius {
		t.Errorf("QoLAutoLootRadius mismatch: expected %f, got %f",
			customSettings.QoLAutoLootRadius, sm2.settings.QoLAutoLootRadius)
	}
	if sm2.settings.QoLMountWhistle != customSettings.QoLMountWhistle {
		t.Errorf("QoLMountWhistle mismatch: expected %v, got %v",
			customSettings.QoLMountWhistle, sm2.settings.QoLMountWhistle)
	}
	if sm2.settings.QoLRecipeTracking != customSettings.QoLRecipeTracking {
		t.Errorf("QoLRecipeTracking mismatch: expected %v, got %v",
			customSettings.QoLRecipeTracking, sm2.settings.QoLRecipeTracking)
	}
	if sm2.settings.QoLSortPreset != customSettings.QoLSortPreset {
		t.Errorf("QoLSortPreset mismatch: expected %s, got %s",
			customSettings.QoLSortPreset, sm2.settings.QoLSortPreset)
	}
}

func TestSettingsManager_LoadInvalidJSON(t *testing.T) {
	// Create temporary settings directory
	tempDir := t.TempDir()
	settingsPath := filepath.Join(tempDir, "settings.json")

	// Write invalid JSON
	invalidJSON := []byte("{invalid json content")
	err := os.WriteFile(settingsPath, invalidJSON, 0o644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	sm := &SettingsManager{
		settingsPath: settingsPath,
	}

	// Load should fail with parse error
	err = sm.LoadSettings()
	if err == nil {
		t.Error("Expected error loading invalid JSON, got nil")
	}
}

func TestSettingsManager_LoadCorruptedSettings(t *testing.T) {
	// Create temporary settings directory
	tempDir := t.TempDir()
	settingsPath := filepath.Join(tempDir, "settings.json")

	// Write settings with invalid values
	corruptedSettings := GameSettings{
		MasterVolume:    5.0,  // Invalid
		MusicVolume:     -2.0, // Invalid
		WindowWidth:     10,   // Invalid
		WindowHeight:    10,   // Invalid
		GraphicsQuality: "ultra_invalid",
	}

	data, _ := json.Marshal(corruptedSettings)
	err := os.WriteFile(settingsPath, data, 0o644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	sm := &SettingsManager{
		settingsPath: settingsPath,
	}

	// Load should succeed but validate and correct
	err = sm.LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings failed: %v", err)
	}

	// Verify corrupted values were corrected
	if sm.settings.MasterVolume < 0 || sm.settings.MasterVolume > 1 {
		t.Error("Corrupted MasterVolume not corrected")
	}
	if sm.settings.WindowWidth < 800 {
		t.Error("Corrupted WindowWidth not corrected")
	}
	if sm.settings.GraphicsQuality != "medium" {
		t.Error("Corrupted GraphicsQuality not corrected")
	}
}

func TestSettingsManager_GetSettings(t *testing.T) {
	sm := &SettingsManager{
		settings: DefaultSettings(),
	}

	settings := sm.GetSettings()

	// Verify it's a copy (modification shouldn't affect original)
	settings.MasterVolume = 0.1
	if sm.settings.MasterVolume == 0.1 {
		t.Error("GetSettings should return a copy, not reference")
	}
}

func TestSettingsManager_UpdateSettings(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	newSettings := GameSettings{
		MasterVolume:    0.9,
		MusicVolume:     0.8,
		SFXVolume:       0.7,
		WindowWidth:     1920,
		WindowHeight:    1080,
		Fullscreen:      true,
		GraphicsQuality: "high",
		VSync:           true,
		ShowFPS:         true,
		ShowTutorials:   false,
	}

	// Update settings
	err := sm.UpdateSettings(newSettings)
	if err != nil {
		t.Fatalf("UpdateSettings failed: %v", err)
	}

	// Verify updated
	if sm.settings.MasterVolume != 0.9 {
		t.Error("Settings not updated")
	}

	// Verify persisted
	sm2 := &SettingsManager{
		settingsPath: sm.settingsPath,
	}
	err = sm2.LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings failed: %v", err)
	}

	if sm2.settings.MasterVolume != 0.9 {
		t.Error("Settings not persisted")
	}
}

func TestSettingsManager_UpdateInvalidSettings(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	invalidSettings := GameSettings{
		MasterVolume:    10.0, // Invalid
		WindowWidth:     50,   // Invalid
		GraphicsQuality: "mega_ultra",
	}

	// Update should validate and correct
	err := sm.UpdateSettings(invalidSettings)
	if err != nil {
		t.Fatalf("UpdateSettings failed: %v", err)
	}

	// Verify corrected
	if sm.settings.MasterVolume < 0 || sm.settings.MasterVolume > 1 {
		t.Error("Invalid MasterVolume not corrected during update")
	}
	if sm.settings.WindowWidth < 800 {
		t.Error("Invalid WindowWidth not corrected during update")
	}
}

func TestSettingsManager_GetSettingsPath(t *testing.T) {
	sm := &SettingsManager{
		settingsPath: "/home/user/.venture/settings.json",
	}

	path := sm.GetSettingsPath()
	if path != "/home/user/.venture/settings.json" {
		t.Errorf("Expected /home/user/.venture/settings.json, got %s", path)
	}
}

// Benchmark settings operations
func BenchmarkDefaultSettings(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DefaultSettings()
	}
}

func BenchmarkGameSettings_Validate(b *testing.B) {
	settings := DefaultSettings()
	for i := 0; i < b.N; i++ {
		settings.Validate()
	}
}

func BenchmarkSettingsManager_SaveSettings(b *testing.B) {
	tempDir := b.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.SaveSettings()
	}
}

func BenchmarkSettingsManager_LoadSettings(b *testing.B) {
	tempDir := b.TempDir()
	sm := &SettingsManager{
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	// Create settings file
	sm.settings = DefaultSettings()
	sm.SaveSettings()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.LoadSettings()
	}
}
