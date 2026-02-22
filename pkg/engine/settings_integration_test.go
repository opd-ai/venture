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

// TestApplySettings_ShowTutorials tests that ShowTutorials setting is applied to tutorial systems.
// This validates the fix for Task 3.3 from PLAN.md (ShowTutorials Setting Not Wired).
func TestApplySettings_ShowTutorials(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	// Create tutorial systems
	tutorialSystem := NewTutorialSystem()
	charCreationTutorial := NewCharacterCreationTutorial()

	game := &EbitenGame{
		SettingsManager:           sm,
		TutorialSystem:            tutorialSystem,
		CharacterCreationTutorial: charCreationTutorial,
	}

	// Verify initial state (tutorials enabled by default)
	if !tutorialSystem.Enabled {
		t.Error("Expected TutorialSystem to be enabled by default")
	}
	if !charCreationTutorial.Enabled {
		t.Error("Expected CharacterCreationTutorial to be enabled by default")
	}

	// Disable tutorials via settings
	settings := sm.GetSettings()
	settings.ShowTutorials = false
	sm.UpdateSettings(settings)

	// Apply settings
	err := game.ApplySettings()
	if err != nil {
		t.Fatalf("ApplySettings failed: %v", err)
	}

	// Verify tutorials were disabled
	if tutorialSystem.Enabled {
		t.Error("Expected TutorialSystem to be disabled after ShowTutorials=false")
	}
	if tutorialSystem.ShowUI {
		t.Error("Expected TutorialSystem.ShowUI to be disabled after ShowTutorials=false")
	}
	if charCreationTutorial.Enabled {
		t.Error("Expected CharacterCreationTutorial to be disabled after ShowTutorials=false")
	}
	if charCreationTutorial.ShowUI {
		t.Error("Expected CharacterCreationTutorial.ShowUI to be disabled after ShowTutorials=false")
	}

	// Re-enable tutorials via settings
	settings.ShowTutorials = true
	sm.UpdateSettings(settings)
	err = game.ApplySettings()
	if err != nil {
		t.Fatalf("ApplySettings failed on re-enable: %v", err)
	}

	// Verify tutorials were re-enabled
	if !tutorialSystem.Enabled {
		t.Error("Expected TutorialSystem to be re-enabled after ShowTutorials=true")
	}
	if !charCreationTutorial.Enabled {
		t.Error("Expected CharacterCreationTutorial to be re-enabled after ShowTutorials=true")
	}
}

// TestApplySettings_ShowTutorials_NilTutorialSystems tests graceful handling when tutorial systems are nil.
func TestApplySettings_ShowTutorials_NilTutorialSystems(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	game := &EbitenGame{
		SettingsManager:           sm,
		TutorialSystem:            nil, // Explicitly nil
		CharacterCreationTutorial: nil, // Explicitly nil
	}

	// Should not panic when tutorial systems are nil
	settings := sm.GetSettings()
	settings.ShowTutorials = false
	sm.UpdateSettings(settings)

	err := game.ApplySettings()
	if err != nil {
		t.Fatalf("ApplySettings should not fail with nil tutorial systems: %v", err)
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

// stubContextualTutorial is a test stub implementing ContextualTutorialProvider.
type stubContextualTutorial struct {
	enabled bool
}

func (s *stubContextualTutorial) Enable()         { s.enabled = true }
func (s *stubContextualTutorial) Disable()        { s.enabled = false }
func (s *stubContextualTutorial) IsEnabled() bool { return s.enabled }

// TestApplySettings_ContextualTutorial tests that ShowTutorials setting is applied to ContextualTutorial.
// This validates the Phase 3.3 implementation for context-sensitive help (TutorialManager).
func TestApplySettings_ContextualTutorial(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	// Create contextual tutorial stub
	contextualTutorial := &stubContextualTutorial{enabled: true}

	game := &EbitenGame{
		SettingsManager:    sm,
		ContextualTutorial: contextualTutorial,
	}

	// Verify initial state (enabled)
	if !contextualTutorial.IsEnabled() {
		t.Error("Expected ContextualTutorial to be enabled initially")
	}

	// Disable tutorials via settings
	settings := sm.GetSettings()
	settings.ShowTutorials = false
	sm.UpdateSettings(settings)

	err := game.ApplySettings()
	if err != nil {
		t.Fatalf("ApplySettings failed: %v", err)
	}

	// Verify contextual tutorial was disabled
	if contextualTutorial.IsEnabled() {
		t.Error("Expected ContextualTutorial to be disabled after ShowTutorials=false")
	}

	// Re-enable tutorials via settings
	settings.ShowTutorials = true
	sm.UpdateSettings(settings)
	err = game.ApplySettings()
	if err != nil {
		t.Fatalf("ApplySettings failed on re-enable: %v", err)
	}

	// Verify contextual tutorial was re-enabled
	if !contextualTutorial.IsEnabled() {
		t.Error("Expected ContextualTutorial to be re-enabled after ShowTutorials=true")
	}
}

// TestApplySettings_OnboardingManager tests that ShowTutorials setting is applied to OnboardingManager.
// This validates the Phase 3.3 implementation for coordinated tutorial control.
func TestApplySettings_OnboardingManager(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	// Create onboarding manager
	onboardingManager := NewOnboardingManager(nil)

	game := &EbitenGame{
		SettingsManager:   sm,
		OnboardingManager: onboardingManager,
	}

	// Verify initial state (enabled)
	if !onboardingManager.IsEnabled() {
		t.Error("Expected OnboardingManager to be enabled initially")
	}

	// Disable tutorials via settings
	settings := sm.GetSettings()
	settings.ShowTutorials = false
	sm.UpdateSettings(settings)

	err := game.ApplySettings()
	if err != nil {
		t.Fatalf("ApplySettings failed: %v", err)
	}

	// Verify onboarding was disabled
	if onboardingManager.IsEnabled() {
		t.Error("Expected OnboardingManager to be disabled after ShowTutorials=false")
	}

	// Re-enable tutorials via settings
	settings.ShowTutorials = true
	sm.UpdateSettings(settings)
	err = game.ApplySettings()
	if err != nil {
		t.Fatalf("ApplySettings failed on re-enable: %v", err)
	}

	// Verify onboarding was re-enabled
	if !onboardingManager.IsEnabled() {
		t.Error("Expected OnboardingManager to be re-enabled after ShowTutorials=true")
	}
}

// TestSetContextualTutorial tests the SetContextualTutorial method applies settings immediately.
func TestSetContextualTutorial(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	// Disable tutorials before setting contextual tutorial
	settings := sm.GetSettings()
	settings.ShowTutorials = false
	sm.UpdateSettings(settings)

	game := &EbitenGame{
		SettingsManager: sm,
	}

	// Create contextual tutorial stub (enabled by default)
	contextualTutorial := &stubContextualTutorial{enabled: true}

	// Setting contextual tutorial should immediately apply the ShowTutorials setting
	game.SetContextualTutorial(contextualTutorial)

	// Verify contextual tutorial was disabled based on settings
	if contextualTutorial.IsEnabled() {
		t.Error("Expected ContextualTutorial to be disabled after SetContextualTutorial with ShowTutorials=false")
	}
}

// TestApplySettings_AllTutorialSystems tests that all three tutorial layers are controlled together.
// This is the comprehensive test for Phase 3.3 (PLAN.md).
func TestApplySettings_AllTutorialSystems(t *testing.T) {
	tempDir := t.TempDir()
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: filepath.Join(tempDir, "settings.json"),
	}

	// Create all tutorial systems
	tutorialSystem := NewTutorialSystem()
	charCreationTutorial := NewCharacterCreationTutorial()
	contextualTutorial := &stubContextualTutorial{enabled: true}
	onboardingManager := NewOnboardingManager(nil)

	game := &EbitenGame{
		SettingsManager:           sm,
		TutorialSystem:            tutorialSystem,
		CharacterCreationTutorial: charCreationTutorial,
		ContextualTutorial:        contextualTutorial,
		OnboardingManager:         onboardingManager,
	}

	// Verify all systems start enabled
	if !tutorialSystem.Enabled {
		t.Error("Expected TutorialSystem to be enabled initially")
	}
	if !charCreationTutorial.Enabled {
		t.Error("Expected CharacterCreationTutorial to be enabled initially")
	}
	if !contextualTutorial.IsEnabled() {
		t.Error("Expected ContextualTutorial to be enabled initially")
	}
	if !onboardingManager.IsEnabled() {
		t.Error("Expected OnboardingManager to be enabled initially")
	}

	// Disable tutorials
	settings := sm.GetSettings()
	settings.ShowTutorials = false
	sm.UpdateSettings(settings)
	err := game.ApplySettings()
	if err != nil {
		t.Fatalf("ApplySettings failed: %v", err)
	}

	// Verify ALL systems were disabled
	if tutorialSystem.Enabled {
		t.Error("TutorialSystem should be disabled")
	}
	if charCreationTutorial.Enabled {
		t.Error("CharacterCreationTutorial should be disabled")
	}
	if contextualTutorial.IsEnabled() {
		t.Error("ContextualTutorial should be disabled")
	}
	if onboardingManager.IsEnabled() {
		t.Error("OnboardingManager should be disabled")
	}

	// Re-enable tutorials
	settings.ShowTutorials = true
	sm.UpdateSettings(settings)
	err = game.ApplySettings()
	if err != nil {
		t.Fatalf("ApplySettings failed on re-enable: %v", err)
	}

	// Verify ALL systems were re-enabled
	if !tutorialSystem.Enabled {
		t.Error("TutorialSystem should be re-enabled")
	}
	if !charCreationTutorial.Enabled {
		t.Error("CharacterCreationTutorial should be re-enabled")
	}
	if !contextualTutorial.IsEnabled() {
		t.Error("ContextualTutorial should be re-enabled")
	}
	if !onboardingManager.IsEnabled() {
		t.Error("OnboardingManager should be re-enabled")
	}
}
