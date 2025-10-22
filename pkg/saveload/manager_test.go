package saveload

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestSaveManager_NewSaveManager tests creating a new save manager.
func TestSaveManager_NewSaveManager(t *testing.T) {
	// Create temporary directory for tests
	tmpDir := t.TempDir()

	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	if manager == nil {
		t.Fatal("NewSaveManager returned nil manager")
	}

	if manager.saveDir != tmpDir {
		t.Errorf("Expected saveDir %s, got %s", tmpDir, manager.saveDir)
	}

	// Verify directory was created
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("Save directory was not created")
	}
}

// TestSaveManager_SaveAndLoad tests basic save/load functionality.
func TestSaveManager_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Create a test save
	save := NewGameSave()
	save.PlayerState.EntityID = 12345
	save.PlayerState.X = 100.5
	save.PlayerState.Y = 200.7
	save.PlayerState.Level = 10
	save.PlayerState.Experience = 5000
	save.WorldState.Seed = 67890
	save.WorldState.GenreID = "fantasy"
	save.WorldState.Width = 100
	save.WorldState.Height = 80

	// Save it
	err = manager.SaveGame("test1", save)
	if err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	// Load it back
	loaded, err := manager.LoadGame("test1")
	if err != nil {
		t.Fatalf("LoadGame failed: %v", err)
	}

	// Verify player state
	if loaded.PlayerState.EntityID != 12345 {
		t.Errorf("Expected EntityID 12345, got %d", loaded.PlayerState.EntityID)
	}
	if loaded.PlayerState.X != 100.5 {
		t.Errorf("Expected X 100.5, got %f", loaded.PlayerState.X)
	}
	if loaded.PlayerState.Y != 200.7 {
		t.Errorf("Expected Y 200.7, got %f", loaded.PlayerState.Y)
	}
	if loaded.PlayerState.Level != 10 {
		t.Errorf("Expected Level 10, got %d", loaded.PlayerState.Level)
	}
	if loaded.PlayerState.Experience != 5000 {
		t.Errorf("Expected Experience 5000, got %d", loaded.PlayerState.Experience)
	}

	// Verify world state
	if loaded.WorldState.Seed != 67890 {
		t.Errorf("Expected Seed 67890, got %d", loaded.WorldState.Seed)
	}
	if loaded.WorldState.GenreID != "fantasy" {
		t.Errorf("Expected GenreID 'fantasy', got %s", loaded.WorldState.GenreID)
	}
	if loaded.WorldState.Width != 100 {
		t.Errorf("Expected Width 100, got %d", loaded.WorldState.Width)
	}
	if loaded.WorldState.Height != 80 {
		t.Errorf("Expected Height 80, got %d", loaded.WorldState.Height)
	}
}

// TestSaveManager_SaveWithExtension tests saving with .sav extension.
func TestSaveManager_SaveWithExtension(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	save := NewGameSave()
	save.PlayerState.Level = 5

	// Save with extension
	err = manager.SaveGame("test.sav", save)
	if err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	// Load without extension
	loaded, err := manager.LoadGame("test")
	if err != nil {
		t.Fatalf("LoadGame failed: %v", err)
	}

	if loaded.PlayerState.Level != 5 {
		t.Errorf("Expected Level 5, got %d", loaded.PlayerState.Level)
	}
}

// TestSaveManager_LoadNonexistent tests loading a nonexistent save.
func TestSaveManager_LoadNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	_, err = manager.LoadGame("nonexistent")
	if err == nil {
		t.Error("Expected error when loading nonexistent save")
	}
}

// TestSaveManager_DeleteSave tests deleting a save file.
func TestSaveManager_DeleteSave(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Create and save
	save := NewGameSave()
	err = manager.SaveGame("test-delete", save)
	if err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	// Verify it exists
	if !manager.SaveExists("test-delete") {
		t.Fatal("Save should exist after saving")
	}

	// Delete it
	err = manager.DeleteSave("test-delete")
	if err != nil {
		t.Fatalf("DeleteSave failed: %v", err)
	}

	// Verify it doesn't exist
	if manager.SaveExists("test-delete") {
		t.Error("Save should not exist after deletion")
	}

	// Try to delete again (should error)
	err = manager.DeleteSave("test-delete")
	if err == nil {
		t.Error("Expected error when deleting nonexistent save")
	}
}

// TestSaveManager_ListSaves tests listing all saves.
func TestSaveManager_ListSaves(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Create multiple saves
	for i := 1; i <= 3; i++ {
		save := NewGameSave()
		save.PlayerState.Level = i * 10
		save.WorldState.GenreID = "fantasy"
		
		// Sleep briefly to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
		
		saveName := "save" + string(rune('0'+i))
		err = manager.SaveGame(saveName, save)
		if err != nil {
			t.Fatalf("SaveGame %d failed: %v", i, err)
		}
	}

	// List saves
	saves, err := manager.ListSaves()
	if err != nil {
		t.Fatalf("ListSaves failed: %v", err)
	}

	if len(saves) != 3 {
		t.Errorf("Expected 3 saves, got %d", len(saves))
	}

	// Verify saves are sorted by timestamp (newest first)
	for i := 0; i < len(saves)-1; i++ {
		if saves[i].Timestamp.Before(saves[i+1].Timestamp) {
			t.Error("Saves should be sorted by timestamp (newest first)")
		}
	}
}

// TestSaveManager_GetSaveMetadata tests getting save metadata.
func TestSaveManager_GetSaveMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Create a save
	save := NewGameSave()
	save.PlayerState.Level = 25
	save.WorldState.GenreID = "scifi"
	save.WorldState.GameTime = 3600.5

	err = manager.SaveGame("metadata-test", save)
	if err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	// Get metadata
	metadata, err := manager.GetSaveMetadata("metadata-test")
	if err != nil {
		t.Fatalf("GetSaveMetadata failed: %v", err)
	}

	if metadata.Name != "metadata-test" {
		t.Errorf("Expected name 'metadata-test', got %s", metadata.Name)
	}
	if metadata.Version != SaveVersion {
		t.Errorf("Expected version %s, got %s", SaveVersion, metadata.Version)
	}
	if metadata.PlayerLevel != 25 {
		t.Errorf("Expected PlayerLevel 25, got %d", metadata.PlayerLevel)
	}
	if metadata.GenreID != "scifi" {
		t.Errorf("Expected GenreID 'scifi', got %s", metadata.GenreID)
	}
	if metadata.GameTime != 3600.5 {
		t.Errorf("Expected GameTime 3600.5, got %f", metadata.GameTime)
	}
	if metadata.FileSize <= 0 {
		t.Error("Expected FileSize > 0")
	}
}

// TestSaveManager_SaveExists tests checking if a save exists.
func TestSaveManager_SaveExists(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Should not exist initially
	if manager.SaveExists("exists-test") {
		t.Error("Save should not exist initially")
	}

	// Create save
	save := NewGameSave()
	err = manager.SaveGame("exists-test", save)
	if err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	// Should exist now
	if !manager.SaveExists("exists-test") {
		t.Error("Save should exist after saving")
	}
}

// TestSaveManager_ValidateSaveName tests save name validation.
func TestSaveManager_ValidateSaveName(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	tests := []struct {
		name      string
		saveName  string
		wantError bool
	}{
		{"valid name", "mysave", false},
		{"valid with numbers", "save123", false},
		{"valid with underscores", "my_save_1", false},
		{"valid with dashes", "my-save-1", false},
		{"empty name", "", true},
		{"path separator slash", "path/to/save", true},
		{"path separator backslash", "path\\to\\save", true},
		{"invalid char colon", "save:1", true},
		{"invalid char pipe", "save|1", true},
		{"invalid char asterisk", "save*", true},
	}

	save := NewGameSave()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.SaveGame(tt.saveName, save)
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// TestSaveManager_SaveNil tests saving nil save.
func TestSaveManager_SaveNil(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	err = manager.SaveGame("test", nil)
	if err == nil {
		t.Error("Expected error when saving nil save")
	}
}

// TestGameSave_NewGameSave tests creating a new game save.
func TestGameSave_NewGameSave(t *testing.T) {
	save := NewGameSave()

	if save == nil {
		t.Fatal("NewGameSave returned nil")
	}

	if save.Version != SaveVersion {
		t.Errorf("Expected version %s, got %s", SaveVersion, save.Version)
	}

	if save.PlayerState == nil {
		t.Error("PlayerState should not be nil")
	}

	if save.WorldState == nil {
		t.Error("WorldState should not be nil")
	}

	if save.Settings == nil {
		t.Error("Settings should not be nil")
	}

	if save.PlayerState.InventoryItems == nil {
		t.Error("InventoryItems should not be nil")
	}

	if save.WorldState.ModifiedEntities == nil {
		t.Error("ModifiedEntities should not be nil")
	}

	if save.Settings.KeyBindings == nil {
		t.Error("KeyBindings should not be nil")
	}
}

// TestSaveManager_ComplexSave tests saving/loading complex data.
func TestSaveManager_ComplexSave(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Create a complex save with all fields populated
	save := NewGameSave()
	
	// Player state
	save.PlayerState.EntityID = 99999
	save.PlayerState.X = 123.45
	save.PlayerState.Y = 678.90
	save.PlayerState.CurrentHealth = 85.5
	save.PlayerState.MaxHealth = 100.0
	save.PlayerState.Level = 42
	save.PlayerState.Experience = 123456
	save.PlayerState.Attack = 75.0
	save.PlayerState.Defense = 50.0
	save.PlayerState.MagicPower = 80.0
	save.PlayerState.Speed = 100.0
	save.PlayerState.InventoryItems = []uint64{1, 2, 3, 4, 5}
	save.PlayerState.EquippedWeapon = 101
	save.PlayerState.EquippedArmor = 202
	save.PlayerState.EquippedAccessory = 303

	// World state
	save.WorldState.Seed = 98765
	save.WorldState.GenreID = "cyberpunk"
	save.WorldState.Width = 200
	save.WorldState.Height = 150
	save.WorldState.GameTime = 7200.5
	save.WorldState.Difficulty = 0.75
	save.WorldState.Depth = 10
	save.WorldState.ModifiedEntities = []ModifiedEntity{
		{EntityID: 1001, X: 10.5, Y: 20.5, Health: 50.0, IsAlive: true, IsPicked: false},
		{EntityID: 1002, X: 30.5, Y: 40.5, Health: 0.0, IsAlive: false, IsPicked: false},
		{EntityID: 2001, X: 50.5, Y: 60.5, Health: 0.0, IsAlive: true, IsPicked: true},
	}

	// Settings
	save.Settings.ScreenWidth = 1920
	save.Settings.ScreenHeight = 1080
	save.Settings.Fullscreen = true
	save.Settings.VSync = false
	save.Settings.MasterVolume = 0.8
	save.Settings.MusicVolume = 0.6
	save.Settings.SFXVolume = 0.9
	save.Settings.KeyBindings = map[string]string{
		"move_up":    "w",
		"move_down":  "s",
		"move_left":  "a",
		"move_right": "d",
		"attack":     "space",
	}

	// Save it
	err = manager.SaveGame("complex", save)
	if err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	// Load it back
	loaded, err := manager.LoadGame("complex")
	if err != nil {
		t.Fatalf("LoadGame failed: %v", err)
	}

	// Verify all fields
	if loaded.PlayerState.EntityID != save.PlayerState.EntityID {
		t.Error("EntityID mismatch")
	}
	if loaded.PlayerState.X != save.PlayerState.X {
		t.Error("X mismatch")
	}
	if loaded.PlayerState.EquippedWeapon != save.PlayerState.EquippedWeapon {
		t.Error("EquippedWeapon mismatch")
	}
	if len(loaded.PlayerState.InventoryItems) != len(save.PlayerState.InventoryItems) {
		t.Error("InventoryItems length mismatch")
	}
	if loaded.WorldState.GenreID != save.WorldState.GenreID {
		t.Error("GenreID mismatch")
	}
	if len(loaded.WorldState.ModifiedEntities) != len(save.WorldState.ModifiedEntities) {
		t.Error("ModifiedEntities length mismatch")
	}
	if loaded.Settings.Fullscreen != save.Settings.Fullscreen {
		t.Error("Fullscreen mismatch")
	}
	if len(loaded.Settings.KeyBindings) != len(save.Settings.KeyBindings) {
		t.Error("KeyBindings length mismatch")
	}
}

// TestSaveManager_LoadCorruptedFile tests loading a corrupted save file.
func TestSaveManager_LoadCorruptedFile(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Write corrupted JSON file
	filename := manager.getFilePath("corrupted")
	err = os.WriteFile(filename, []byte("not valid json {{{"), 0644)
	if err != nil {
		t.Fatalf("Failed to write corrupted file: %v", err)
	}

	// Try to load it
	_, err = manager.LoadGame("corrupted")
	if err == nil {
		t.Error("Expected error when loading corrupted save")
	}
}

// TestSaveManager_LoadMissingFields tests loading a save with missing required fields.
func TestSaveManager_LoadMissingFields(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	tests := []struct {
		name     string
		saveData string
	}{
		{
			"missing_player",
			`{"version":"1.0.0","timestamp":"2025-01-01T00:00:00Z","world":{},"settings":{}}`,
		},
		{
			"missing_world",
			`{"version":"1.0.0","timestamp":"2025-01-01T00:00:00Z","player":{},"settings":{}}`,
		},
		{
			"missing_settings",
			`{"version":"1.0.0","timestamp":"2025-01-01T00:00:00Z","player":{},"world":{}}`,
		},
		{
			"missing_version",
			`{"timestamp":"2025-01-01T00:00:00Z","player":{},"world":{},"settings":{}}`,
		},
		{
			"wrong_version",
			`{"version":"0.5.0","timestamp":"2025-01-01T00:00:00Z","player":{},"world":{},"settings":{}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := manager.getFilePath(tt.name)
			err := os.WriteFile(filename, []byte(tt.saveData), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			_, err = manager.LoadGame(tt.name)
			if err == nil {
				t.Error("Expected error when loading save with missing fields")
			}
		})
	}
}

// TestSaveManager_GetMetadataEmptyFile tests getting metadata from an empty file.
func TestSaveManager_GetMetadataEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Write empty file
	filename := manager.getFilePath("empty")
	err = os.WriteFile(filename, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to write empty file: %v", err)
	}

	// Try to get metadata
	_, err = manager.GetSaveMetadata("empty")
	if err == nil {
		t.Error("Expected error when getting metadata from empty file")
	}
}

// TestSaveManager_ListSavesWithNonSavFiles tests listing saves with non-.sav files present.
func TestSaveManager_ListSavesWithNonSavFiles(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Create a valid save
	save := NewGameSave()
	err = manager.SaveGame("valid", save)
	if err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	// Create some non-.sav files in the directory
	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "config.json"), []byte("{}"), 0644)

	// Create a subdirectory
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	// List saves - should only return the valid save
	saves, err := manager.ListSaves()
	if err != nil {
		t.Fatalf("ListSaves failed: %v", err)
	}

	if len(saves) != 1 {
		t.Errorf("Expected 1 save, got %d", len(saves))
	}

	if len(saves) > 0 && saves[0].Name != "valid" {
		t.Errorf("Expected save name 'valid', got %s", saves[0].Name)
	}
}

// TestSaveManager_NewSaveManagerNonexistentDir tests creating a manager with a directory that needs to be created.
func TestSaveManager_NewSaveManagerNonexistentDir(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "nested", "saves")

	manager, err := NewSaveManager(nestedDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
		t.Error("Nested save directory was not created")
	}

	// Verify we can save to it
	save := NewGameSave()
	err = manager.SaveGame("test", save)
	if err != nil {
		t.Fatalf("SaveGame failed in nested directory: %v", err)
	}
}
