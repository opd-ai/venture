// Package engine provides game settings management with persistent storage.
// Settings are stored in JSON format in the user's home directory (~/.venture/settings.json).
package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// GameSettings holds all configurable game settings.
// Uses sensible defaults for first-time users.
type GameSettings struct {
	// Audio settings
	MasterVolume float64 `json:"master_volume"` // 0.0-1.0
	MusicVolume  float64 `json:"music_volume"`  // 0.0-1.0
	SFXVolume    float64 `json:"sfx_volume"`    // 0.0-1.0

	// Display settings
	WindowWidth  int  `json:"window_width"`
	WindowHeight int  `json:"window_height"`
	Fullscreen   bool `json:"fullscreen"`

	// Graphics settings
	GraphicsQuality string `json:"graphics_quality"` // "low", "medium", "high"
	VSync           bool   `json:"vsync"`
	ShowFPS         bool   `json:"show_fps"`

	// Gameplay settings
	ShowTutorials bool `json:"show_tutorials"`
}

// DefaultSettings returns game settings with default values.
// These are sensible defaults for a good first-time experience.
func DefaultSettings() GameSettings {
	return GameSettings{
		// Audio defaults - moderate volume
		MasterVolume: 0.7,
		MusicVolume:  0.6,
		SFXVolume:    0.8,

		// Display defaults - 1280x720 windowed
		WindowWidth:  1280,
		WindowHeight: 720,
		Fullscreen:   false,

		// Graphics defaults - medium quality
		GraphicsQuality: "medium",
		VSync:           true,
		ShowFPS:         false,

		// Gameplay defaults
		ShowTutorials: true,
	}
}

// Validate checks if settings have valid values and corrects them if not.
// Returns true if any corrections were made.
func (s *GameSettings) Validate() bool {
	corrected := false

	// Validate volumes (0.0-1.0)
	if s.MasterVolume < 0.0 || s.MasterVolume > 1.0 {
		s.MasterVolume = 0.7
		corrected = true
	}
	if s.MusicVolume < 0.0 || s.MusicVolume > 1.0 {
		s.MusicVolume = 0.6
		corrected = true
	}
	if s.SFXVolume < 0.0 || s.SFXVolume > 1.0 {
		s.SFXVolume = 0.8
		corrected = true
	}

	// Validate window dimensions
	if s.WindowWidth < 800 || s.WindowWidth > 3840 {
		s.WindowWidth = 1280
		corrected = true
	}
	if s.WindowHeight < 600 || s.WindowHeight > 2160 {
		s.WindowHeight = 720
		corrected = true
	}

	// Validate graphics quality
	validQualities := map[string]bool{"low": true, "medium": true, "high": true}
	if !validQualities[s.GraphicsQuality] {
		s.GraphicsQuality = "medium"
		corrected = true
	}

	return corrected
}

// SettingsManager handles loading and saving game settings.
type SettingsManager struct {
	settings     GameSettings
	settingsPath string
}

// NewSettingsManager creates a settings manager.
// Automatically creates settings directory if it doesn't exist.
func NewSettingsManager() (*SettingsManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	settingsDir := filepath.Join(homeDir, ".venture")
	settingsPath := filepath.Join(settingsDir, "settings.json")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create settings directory: %w", err)
	}

	return &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: settingsPath,
	}, nil
}

// LoadSettings loads settings from disk or uses defaults if file doesn't exist.
// Returns error only on file read/parse failures, not if file doesn't exist.
func (sm *SettingsManager) LoadSettings() error {
	data, err := os.ReadFile(sm.settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, use defaults (first time running)
			sm.settings = DefaultSettings()
			return nil
		}
		return fmt.Errorf("failed to read settings file: %w", err)
	}

	var loaded GameSettings
	if err := json.Unmarshal(data, &loaded); err != nil {
		return fmt.Errorf("failed to parse settings file: %w", err)
	}

	// Validate loaded settings
	loaded.Validate()
	sm.settings = loaded

	return nil
}

// SaveSettings writes current settings to disk.
func (sm *SettingsManager) SaveSettings() error {
	// Validate before saving
	sm.settings.Validate()

	data, err := json.MarshalIndent(sm.settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(sm.settingsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

// GetSettings returns a copy of current settings.
func (sm *SettingsManager) GetSettings() GameSettings {
	return sm.settings
}

// UpdateSettings updates settings and saves to disk.
// Validates settings before applying.
func (sm *SettingsManager) UpdateSettings(newSettings GameSettings) error {
	newSettings.Validate()
	sm.settings = newSettings
	return sm.SaveSettings()
}

// GetSettingsPath returns the full path to the settings file.
// Useful for debugging or manual editing.
func (sm *SettingsManager) GetSettingsPath() string {
	return sm.settingsPath
}
