package engine

import (
	"os"
	"path/filepath"
	"testing"
)

// TestApplySettings_AudioVolumes tests that settings are correctly applied to AudioManager.
func TestApplySettings_AudioVolumes(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	game := &EbitenGame{
		SettingsManager: sm,
		AudioManager:    NewAudioManager(44100, 12345),
	}

	// Set custom volumes
	settings := sm.GetSettings()
	settings.MasterVolume = 0.5
	settings.MusicVolume = 0.8
	settings.SFXVolume = 0.6
	sm.UpdateSettings(settings)

	// Apply settings
	err := game.ApplySettings()
	if err != nil {
		t.Fatalf("ApplySettings failed: %v", err)
	}

	// Verify volumes were applied (MasterVolume * specific volume)
	expectedMusicVolume := 0.5 * 0.8 // 0.4
	expectedSFXVolume := 0.5 * 0.6   // 0.3

	// Check music volume (within tolerance for float comparison)
	if !floatEqual(game.AudioManager.musicVolume, expectedMusicVolume, 0.01) {
		t.Errorf("Expected music volume %f, got %f", expectedMusicVolume, game.AudioManager.musicVolume)
	}

	// Check SFX volume
	if !floatEqual(game.AudioManager.sfxVolume, expectedSFXVolume, 0.01) {
		t.Errorf("Expected SFX volume %f, got %f", expectedSFXVolume, game.AudioManager.sfxVolume)
	}
}

// TestApplySettings_NoAudioManager tests graceful handling when AudioManager is nil.
func TestApplySettings_NoAudioManager(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	game := &EbitenGame{
		SettingsManager: sm,
		AudioManager:    nil, // No audio manager
	}

	// Should not crash
	err := game.ApplySettings()
	if err != nil {
		t.Errorf("ApplySettings should not error when AudioManager is nil, got: %v", err)
	}
}

// TestApplySettings_NoSettingsManager tests graceful handling when SettingsManager is nil.
func TestApplySettings_NoSettingsManager(t *testing.T) {
	game := &EbitenGame{
		SettingsManager: nil,
		AudioManager:    NewAudioManager(44100, 12345),
	}

	// Should not crash
	err := game.ApplySettings()
	if err != nil {
		t.Errorf("ApplySettings should not error when SettingsManager is nil, got: %v", err)
	}
}

// TestSetAudioManager tests that SetAudioManager correctly sets and applies settings.
func TestSetAudioManager(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	// Set custom volumes before setting audio manager
	settings := sm.GetSettings()
	settings.MasterVolume = 0.6
	settings.MusicVolume = 0.7
	sm.UpdateSettings(settings)

	game := &EbitenGame{
		SettingsManager: sm,
	}

	audioManager := NewAudioManager(44100, 12345)

	// Set audio manager (should auto-apply settings)
	game.SetAudioManager(audioManager)

	// Verify audio manager was set
	if game.AudioManager != audioManager {
		t.Error("AudioManager was not set correctly")
	}

	// Verify settings were applied
	expectedMusicVolume := 0.6 * 0.7 // 0.42
	if !floatEqual(game.AudioManager.musicVolume, expectedMusicVolume, 0.01) {
		t.Errorf("Expected music volume %f after SetAudioManager, got %f", expectedMusicVolume, game.AudioManager.musicVolume)
	}
}

// TestSettingsUI_ApplyCallback tests that the apply callback is called when settings are saved.
func TestSettingsUI_ApplyCallback(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	ui := NewSettingsUI(1280, 720, sm)

	callbackCalled := false
	ui.SetApplyCallback(func() {
		callbackCalled = true
	})

	// Show and then hide (which saves settings)
	ui.Show()
	ui.Hide()

	if !callbackCalled {
		t.Error("Expected apply callback to be called after Hide()")
	}
}

// TestSettingsUI_ApplyCallback_NotCalled tests that callback is not called when UI is just shown.
func TestSettingsUI_ApplyCallback_NotCalledOnShow(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	ui := NewSettingsUI(1280, 720, sm)

	callbackCalled := false
	ui.SetApplyCallback(func() {
		callbackCalled = true
	})

	// Only show, don't hide
	ui.Show()

	if callbackCalled {
		t.Error("Expected apply callback to NOT be called after Show() alone")
	}
}

// TestApplySettings_MasterVolumeZero tests that zero master volume mutes everything.
func TestApplySettings_MasterVolumeZero(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	game := &EbitenGame{
		SettingsManager: sm,
		AudioManager:    NewAudioManager(44100, 12345),
	}

	// Set master volume to zero
	settings := sm.GetSettings()
	settings.MasterVolume = 0.0
	settings.MusicVolume = 1.0 // Max individual volume
	settings.SFXVolume = 1.0   // Max individual volume
	sm.UpdateSettings(settings)

	// Apply settings
	err := game.ApplySettings()
	if err != nil {
		t.Fatalf("ApplySettings failed: %v", err)
	}

	// Verify both volumes are zero (master volume mutes everything)
	if game.AudioManager.musicVolume != 0.0 {
		t.Errorf("Expected music volume 0.0 with zero master volume, got %f", game.AudioManager.musicVolume)
	}

	if game.AudioManager.sfxVolume != 0.0 {
		t.Errorf("Expected SFX volume 0.0 with zero master volume, got %f", game.AudioManager.sfxVolume)
	}
}

// TestApplySettings_MaxVolumes tests that max settings work correctly.
func TestApplySettings_MaxVolumes(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	game := &EbitenGame{
		SettingsManager: sm,
		AudioManager:    NewAudioManager(44100, 12345),
	}

	// Set all volumes to max
	settings := sm.GetSettings()
	settings.MasterVolume = 1.0
	settings.MusicVolume = 1.0
	settings.SFXVolume = 1.0
	sm.UpdateSettings(settings)

	// Apply settings
	err := game.ApplySettings()
	if err != nil {
		t.Fatalf("ApplySettings failed: %v", err)
	}

	// Verify both volumes are 1.0
	if game.AudioManager.musicVolume != 1.0 {
		t.Errorf("Expected music volume 1.0, got %f", game.AudioManager.musicVolume)
	}

	if game.AudioManager.sfxVolume != 1.0 {
		t.Errorf("Expected SFX volume 1.0, got %f", game.AudioManager.sfxVolume)
	}
}

// TestSettingsUI_IntegrationWithGame tests full integration flow.
func TestSettingsUI_IntegrationWithGame(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping Ebiten-dependent test in CI")
	}

	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	game := &EbitenGame{
		SettingsManager: sm,
		AudioManager:    NewAudioManager(44100, 12345),
		SettingsUI:      NewSettingsUI(1280, 720, sm),
	}

	// Wire the apply callback
	applyCalled := false
	game.SettingsUI.SetApplyCallback(func() {
		applyCalled = true
		_ = game.ApplySettings() // Ignore error for test
	})

	// Simulate user changing settings
	game.SettingsUI.Show()
	game.SettingsUI.currentSettings.MasterVolume = 0.3
	game.SettingsUI.currentSettings.MusicVolume = 0.5
	game.SettingsUI.Hide() // This should trigger save and apply

	// Verify callback was called
	if !applyCalled {
		t.Error("Expected apply callback to be called")
	}

	// Verify settings were applied to audio
	expectedMusicVolume := 0.3 * 0.5 // 0.15
	if !floatEqual(game.AudioManager.musicVolume, expectedMusicVolume, 0.01) {
		t.Errorf("Expected music volume %f after settings change, got %f", expectedMusicVolume, game.AudioManager.musicVolume)
	}
}

// Benchmark applying settings
func BenchmarkApplySettings(b *testing.B) {
	tempDir := b.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	game := &EbitenGame{
		SettingsManager: sm,
		AudioManager:    NewAudioManager(44100, 12345),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = game.ApplySettings()
	}
}

// Benchmark SetAudioManager
func BenchmarkSetAudioManager(b *testing.B) {
	tempDir := b.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	game := &EbitenGame{
		SettingsManager: sm,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		audioManager := NewAudioManager(44100, 12345)
		game.SetAudioManager(audioManager)
	}
}
