package ui

import (
	"os"
	"testing"
)

func TestSettingsManager_Defaults(t *testing.T) {
	sm := NewSettingsManager()

	// Test graphics settings
	if val := sm.GetString("graphics.resolution"); val != "1920x1080" {
		t.Errorf("expected resolution 1920x1080, got %s", val)
	}

	if val := sm.GetBool("graphics.fullscreen"); val != false {
		t.Errorf("expected fullscreen false, got %v", val)
	}

	// Test audio settings
	if val := sm.GetInt("audio.master_volume"); val != 100 {
		t.Errorf("expected master volume 100, got %d", val)
	}

	// Test gameplay settings
	if val := sm.GetFloat("gameplay.difficulty"); val != 0.5 {
		t.Errorf("expected difficulty 0.5, got %f", val)
	}
}

func TestSettingsManager_SetValue(t *testing.T) {
	sm := NewSettingsManager()

	// Test bool setting
	if err := sm.SetValue("graphics.fullscreen", true); err != nil {
		t.Errorf("failed to set fullscreen: %v", err)
	}
	if val := sm.GetBool("graphics.fullscreen"); val != true {
		t.Errorf("expected fullscreen true after set")
	}

	// Test int setting with bounds
	if err := sm.SetValue("audio.master_volume", 80); err != nil {
		t.Errorf("failed to set volume: %v", err)
	}
	if val := sm.GetInt("audio.master_volume"); val != 80 {
		t.Errorf("expected volume 80, got %d", val)
	}

	// Test int bounds validation
	if err := sm.SetValue("audio.master_volume", 150); err == nil {
		t.Error("expected error for volume > 100")
	}
	if err := sm.SetValue("audio.master_volume", -10); err == nil {
		t.Error("expected error for volume < 0")
	}

	// Test float setting
	if err := sm.SetValue("gameplay.difficulty", 0.8); err != nil {
		t.Errorf("failed to set difficulty: %v", err)
	}
	if val := sm.GetFloat("gameplay.difficulty"); val != 0.8 {
		t.Errorf("expected difficulty 0.8, got %f", val)
	}

	// Test enum setting
	if err := sm.SetValue("graphics.particles", "Ultra"); err != nil {
		t.Errorf("failed to set particles: %v", err)
	}
	if val := sm.GetString("graphics.particles"); val != "Ultra" {
		t.Errorf("expected particles Ultra, got %s", val)
	}

	// Test invalid enum value
	if err := sm.SetValue("graphics.particles", "Invalid"); err == nil {
		t.Error("expected error for invalid enum value")
	}
}

func TestSettingsManager_ListByCategory(t *testing.T) {
	sm := NewSettingsManager()

	graphicsSettings := sm.ListByCategory(CategoryGraphics)
	if len(graphicsSettings) < 3 {
		t.Errorf("expected at least 3 graphics settings, got %d", len(graphicsSettings))
	}

	audioSettings := sm.ListByCategory(CategoryAudio)
	if len(audioSettings) < 3 {
		t.Errorf("expected at least 3 audio settings, got %d", len(audioSettings))
	}
}

func TestSettingsManager_ResetToDefaults(t *testing.T) {
	sm := NewSettingsManager()

	// Modify settings
	sm.SetValue("graphics.fullscreen", true)
	sm.SetValue("audio.master_volume", 50)

	// Reset
	sm.ResetToDefaults()

	// Verify defaults restored
	if val := sm.GetBool("graphics.fullscreen"); val != false {
		t.Error("expected fullscreen reset to false")
	}
	if val := sm.GetInt("audio.master_volume"); val != 100 {
		t.Error("expected volume reset to 100")
	}
}

func TestSettingsManager_SaveLoad(t *testing.T) {
	sm := NewSettingsManager()
	filename := "test_settings.json"
	defer os.Remove(filename)

	// Modify settings
	sm.SetValue("graphics.fullscreen", true)
	sm.SetValue("audio.master_volume", 75)
	sm.SetValue("gameplay.difficulty", 0.7)

	// Save
	if err := sm.Save(filename); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Create new manager and load
	sm2 := NewSettingsManager()
	if err := sm2.Load(filename); err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	// Verify loaded values
	if val := sm2.GetBool("graphics.fullscreen"); val != true {
		t.Error("expected fullscreen true after load")
	}
	if val := sm2.GetInt("audio.master_volume"); val != 75 {
		t.Error("expected volume 75 after load")
	}
	if val := sm2.GetFloat("gameplay.difficulty"); val != 0.7 {
		t.Error("expected difficulty 0.7 after load")
	}
}

func TestSettingsManager_IsModified(t *testing.T) {
	sm := NewSettingsManager()

	if sm.IsModified() {
		t.Error("expected modified=false on new manager")
	}

	sm.SetValue("graphics.fullscreen", true)
	if !sm.IsModified() {
		t.Error("expected modified=true after setting value")
	}
}

func TestSettingType_String(t *testing.T) {
	tests := []struct {
		typ  SettingType
		want string
	}{
		{TypeBool, "Boolean"},
		{TypeInt, "Integer"},
		{TypeFloat, "Float"},
		{TypeString, "String"},
		{TypeEnum, "Enum"},
	}

	for _, tt := range tests {
		if got := tt.typ.String(); got != tt.want {
			t.Errorf("SettingType.String() = %s, want %s", got, tt.want)
		}
	}
}

func TestSettingsCategory_String(t *testing.T) {
	tests := []struct {
		cat  SettingsCategory
		want string
	}{
		{CategoryGraphics, "Graphics"},
		{CategoryAudio, "Audio"},
		{CategoryControls, "Controls"},
		{CategoryGameplay, "Gameplay"},
		{CategoryNetwork, "Network"},
		{CategoryAccessibility, "Accessibility"},
	}

	for _, tt := range tests {
		if got := tt.cat.String(); got != tt.want {
			t.Errorf("SettingsCategory.String() = %s, want %s", got, tt.want)
		}
	}
}

func TestSettingsManager_TypeValidation(t *testing.T) {
	sm := NewSettingsManager()

	// Test type mismatch errors
	if err := sm.SetValue("graphics.fullscreen", "not a bool"); err == nil {
		t.Error("expected error for string value on bool setting")
	}

	if err := sm.SetValue("audio.master_volume", "not an int"); err == nil {
		t.Error("expected error for string value on int setting")
	}

	if err := sm.SetValue("gameplay.difficulty", "not a float"); err == nil {
		t.Error("expected error for string value on float setting")
	}
}

func TestSettingsManager_AccessibilitySettings(t *testing.T) {
	sm := NewSettingsManager()

	// Test colorblind mode enum
	if err := sm.SetValue("accessibility.colorblind_mode", "Protanopia"); err != nil {
		t.Errorf("failed to set colorblind mode: %v", err)
	}

	// Test font scale bounds
	if err := sm.SetValue("accessibility.font_scale", 1.5); err != nil {
		t.Errorf("failed to set font scale: %v", err)
	}
	if err := sm.SetValue("accessibility.font_scale", 3.0); err == nil {
		t.Error("expected error for font scale > 2.0")
	}
	if err := sm.SetValue("accessibility.font_scale", 0.3); err == nil {
		t.Error("expected error for font scale < 0.5")
	}
}

func TestSettingsManager_SaveClearsModified(t *testing.T) {
	sm := NewSettingsManager()
	filename := "test_settings_modified.json"
	defer os.Remove(filename)

	sm.SetValue("graphics.fullscreen", true)
	if !sm.IsModified() {
		t.Fatal("expected modified=true after setting value")
	}

	if err := sm.Save(filename); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	if sm.IsModified() {
		t.Error("expected modified=false after save")
	}
}

func TestSettingsManager_NonexistentSetting(t *testing.T) {
	sm := NewSettingsManager()

	_, err := sm.GetSetting("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent setting")
	}

	err = sm.SetValue("nonexistent", 42)
	if err == nil {
		t.Error("expected error for setting nonexistent value")
	}
}

func BenchmarkSettingsManager_GetValue(b *testing.B) {
	sm := NewSettingsManager()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.GetValue("graphics.resolution")
	}
}

func BenchmarkSettingsManager_SetValue(b *testing.B) {
	sm := NewSettingsManager()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.SetValue("audio.master_volume", 80)
	}
}

func BenchmarkSettingsManager_ListByCategory(b *testing.B) {
	sm := NewSettingsManager()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.ListByCategory(CategoryGraphics)
	}
}
