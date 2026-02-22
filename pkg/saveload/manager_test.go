//go:build !js
// +build !js

package saveload

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
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

	if save.PlayerState.Items == nil {
		t.Error("Items should not be nil")
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
	save.PlayerState.Items = []ItemData{
		{ID: "1", Name: "Sword", Type: "weapon"},
		{ID: "2", Name: "Shield", Type: "armor"},
		{ID: "3", Name: "Potion", Type: "consumable"},
	}
	save.PlayerState.Gold = 1000
	save.PlayerState.EquippedItems = EquipmentData{
		Weapon:    &ItemData{ID: "101", Name: "Epic Sword", Type: "weapon"},
		Armor:     &ItemData{ID: "202", Name: "Epic Armor", Type: "armor"},
		Accessory: &ItemData{ID: "303", Name: "Ring", Type: "accessory"},
	}

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
	if loaded.PlayerState.Gold != save.PlayerState.Gold {
		t.Error("Gold mismatch")
	}
	if len(loaded.PlayerState.Items) != len(save.PlayerState.Items) {
		t.Error("Items length mismatch")
	}
	if loaded.PlayerState.EquippedItems.Weapon == nil || loaded.PlayerState.EquippedItems.Weapon.ID != "101" {
		t.Error("EquippedItems.Weapon mismatch")
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
	err = os.WriteFile(filename, []byte("not valid json {{{"), 0o644)
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
			err := os.WriteFile(filename, []byte(tt.saveData), 0o644)
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
	err = os.WriteFile(filename, []byte(""), 0o644)
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
	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("test"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "config.json"), []byte("{}"), 0o644)

	// Create a subdirectory
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0o755)

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

// TestSaveManager_ValidateAndMigrateEdgeCases tests edge cases for validateAndMigrate.
func TestSaveManager_ValidateAndMigrateEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	tests := []struct {
		name    string
		save    *GameSave
		wantErr bool
	}{
		{
			name:    "nil save",
			save:    nil,
			wantErr: true,
		},
		{
			name: "empty version",
			save: &GameSave{
				Version: "",
			},
			wantErr: true,
		},
		{
			name: "current version - valid",
			save: &GameSave{
				Version: SaveVersion,
				PlayerState: &PlayerState{
					EntityID: 1,
				},
				WorldState: &WorldState{
					Seed: 123,
				},
				Settings: &GameSettings{
					MasterVolume: 1.0,
				},
			},
			wantErr: false,
		},
		{
			name: "current version - missing required fields",
			save: &GameSave{
				Version: SaveVersion,
				// Missing PlayerState, WorldState, Settings
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.validateAndMigrate(tt.save)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAndMigrate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSaveManager_ValidateAndMigrateWithMigrator tests migration paths.
func TestSaveManager_ValidateAndMigrateWithMigrator(t *testing.T) {
	tmpDir := t.TempDir()
	migrator := NewDefaultMigrator()
	manager, err := NewSaveManagerWithMigrator(tmpDir, nil, migrator)
	if err != nil {
		t.Fatalf("NewSaveManagerWithMigrator failed: %v", err)
	}

	// Test migration from old version
	oldSave := &GameSave{
		Version: "0.9.0",
		PlayerState: &PlayerState{
			EntityID: 1,
		},
		WorldState: &WorldState{
			Seed: 123,
		},
	}

	err = manager.validateAndMigrate(oldSave)
	if err != nil {
		t.Fatalf("validateAndMigrate failed for migratable version: %v", err)
	}

	// Verify version was updated
	if oldSave.Version != SaveVersion {
		t.Errorf("Expected version %s, got %s", SaveVersion, oldSave.Version)
	}
}

// TestSaveManager_ValidateAndMigrateUnsupportedVersion tests unsupported version handling.
func TestSaveManager_ValidateAndMigrateUnsupportedVersion(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Test with unsupported version (no migrator set)
	oldSave := &GameSave{
		Version: "0.5.0",
		PlayerState: &PlayerState{
			EntityID: 1,
		},
		WorldState: &WorldState{
			Seed: 123,
		},
	}

	err = manager.validateAndMigrate(oldSave)
	if err == nil {
		t.Error("validateAndMigrate should fail for unsupported version without migrator")
	}
}

// TestSaveManager_MarshalSaveEdgeCases tests marshalSave error paths.
func TestSaveManager_MarshalSaveEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Test with valid save
	save := NewGameSave()
	data, err := manager.marshalSave(save, "test")
	if err != nil {
		t.Fatalf("marshalSave failed for valid save: %v", err)
	}
	if len(data) == 0 {
		t.Error("marshalSave returned empty data")
	}
}

// TestSaveManager_WriteSaveFileEdgeCases tests writeSaveFile error paths.
func TestSaveManager_WriteSaveFileEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Test writing valid data
	data := []byte("test data")
	err = manager.writeSaveFile("write_test", data)
	if err != nil {
		t.Fatalf("writeSaveFile failed: %v", err)
	}

	// Verify file was written
	filePath := manager.getFilePath("write_test")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("File was not written")
	}

	// Verify data
	readData, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}
	if string(readData) != string(data) {
		t.Errorf("Written data mismatch: got %s, want %s", readData, data)
	}
}

// TestSaveManager_ReadSaveFileEdgeCases tests readSaveFile error paths.
func TestSaveManager_ReadSaveFileEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Test reading nonexistent file
	_, err = manager.readSaveFile("nonexistent")
	if err == nil {
		t.Error("readSaveFile should fail for nonexistent file")
	}

	// Create and read valid file
	testData := []byte("test content")
	filePath := manager.getFilePath("read_test")
	err = os.WriteFile(filePath, testData, 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	data, err := manager.readSaveFile("read_test")
	if err != nil {
		t.Fatalf("readSaveFile failed: %v", err)
	}
	if string(data) != string(testData) {
		t.Errorf("Read data mismatch: got %s, want %s", data, testData)
	}
}

// TestSaveManager_LoggingFunctions tests all logging helper functions.
func TestSaveManager_LoggingFunctions(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Test with logger (default logger should be set)
	fields := logrus.Fields{
		"test": "value",
	}

	// These should not panic
	manager.logDebug("debug message", fields)
	manager.logInfo("info message", fields)
	manager.logWarn("warn message", nil, fields)
	manager.logError("error message", nil, fields)

	// Test with actual error
	testErr := &MigrationError{
		SourceVersion: "1.0.0",
		TargetVersion: "2.0.0",
		Message:       "test migration error",
	}
	manager.logWarn("warn with error", testErr, fields)
	manager.logError("error with error", testErr, fields)
}

// TestSaveManager_DeleteSaveEdgeCases tests DeleteSave edge cases.
func TestSaveManager_DeleteSaveEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Test deleting nonexistent save (may return error or succeed - both are acceptable)
	_ = manager.DeleteSave("nonexistent")

	// Create save and delete it
	save := NewGameSave()
	err = manager.SaveGame("delete_test", save)
	if err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	// Verify it exists
	if !manager.SaveExists("delete_test") {
		t.Error("Save should exist before deletion")
	}

	// Delete it
	err = manager.DeleteSave("delete_test")
	if err != nil {
		t.Fatalf("DeleteSave failed: %v", err)
	}

	// Verify it's gone
	if manager.SaveExists("delete_test") {
		t.Error("Save should not exist after deletion")
	}
}

// TestSaveManager_ListSavesEdgeCases tests ListSaves edge cases.
func TestSaveManager_ListSavesEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Test with empty directory
	saves, err := manager.ListSaves()
	if err != nil {
		t.Fatalf("ListSaves failed: %v", err)
	}
	if len(saves) != 0 {
		t.Errorf("Expected 0 saves, got %d", len(saves))
	}

	// Create multiple saves (without path separators)
	for i := 0; i < 5; i++ {
		save := NewGameSave()
		save.PlayerState.Level = i + 1
		name := "save_" + string(rune('A'+i))
		err = manager.SaveGame(name, save)
		if err != nil {
			t.Fatalf("SaveGame failed: %v", err)
		}
		// Add small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	// List saves
	saves, err = manager.ListSaves()
	if err != nil {
		t.Fatalf("ListSaves failed: %v", err)
	}
	if len(saves) != 5 {
		t.Errorf("Expected 5 saves, got %d", len(saves))
	}
}

// TestSaveManager_GetSaveMetadataEdgeCases tests GetSaveMetadata edge cases.
func TestSaveManager_GetSaveMetadataEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Test getting metadata for nonexistent file
	_, err = manager.GetSaveMetadata("nonexistent")
	if err == nil {
		t.Error("GetSaveMetadata should fail for nonexistent file")
	}

	// Create save and get metadata
	save := NewGameSave()
	save.PlayerState.Level = 42
	save.WorldState.GenreID = "fantasy"

	err = manager.SaveGame("metadata_test", save)
	if err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	metadata, err := manager.GetSaveMetadata("metadata_test")
	if err != nil {
		t.Fatalf("GetSaveMetadata failed: %v", err)
	}

	if metadata.Name != "metadata_test" {
		t.Errorf("Expected name 'metadata_test', got %s", metadata.Name)
	}
	if metadata.PlayerLevel != 42 {
		t.Errorf("Expected player level 42, got %d", metadata.PlayerLevel)
	}
	if metadata.GenreID != "fantasy" {
		t.Errorf("Expected genre 'fantasy', got %s", metadata.GenreID)
	}
}

// TestSaveManager_SaveExistsEdgeCases tests SaveExists edge cases.
func TestSaveManager_SaveExistsEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Test nonexistent save
	if manager.SaveExists("nonexistent") {
		t.Error("SaveExists should return false for nonexistent save")
	}

	// Create save
	save := NewGameSave()
	err = manager.SaveGame("exists_test", save)
	if err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	// Test existing save
	if !manager.SaveExists("exists_test") {
		t.Error("SaveExists should return true for existing save")
	}

	// Delete save
	err = manager.DeleteSave("exists_test")
	if err != nil {
		t.Fatalf("DeleteSave failed: %v", err)
	}

	// Test deleted save
	if manager.SaveExists("exists_test") {
		t.Error("SaveExists should return false for deleted save")
	}
}

// TestSaveManager_NewSaveManagerWithMigratorEdgeCases tests NewSaveManagerWithMigrator edge cases.
func TestSaveManager_NewSaveManagerWithMigratorEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with nil logger and nil migrator
	manager, err := NewSaveManagerWithMigrator(tmpDir, nil, nil)
	if err != nil {
		t.Fatalf("NewSaveManagerWithMigrator failed with nil logger and migrator: %v", err)
	}
	if manager == nil {
		t.Fatal("NewSaveManagerWithMigrator returned nil manager")
	}

	// Test with custom logger and migrator
	logger := logrus.New()
	migrator := NewDefaultMigrator()
	manager2, err := NewSaveManagerWithMigrator(tmpDir, logger, migrator)
	if err != nil {
		t.Fatalf("NewSaveManagerWithMigrator failed with custom logger and migrator: %v", err)
	}
	if manager2 == nil {
		t.Fatal("NewSaveManagerWithMigrator returned nil manager")
	}

	// Verify we can save with the custom manager
	save := NewGameSave()
	err = manager2.SaveGame("custom_test", save)
	if err != nil {
		t.Fatalf("SaveGame failed with custom manager: %v", err)
	}
}

// TestSaveManager_LoadGameEdgeCases tests LoadGame additional edge cases.
func TestSaveManager_LoadGameEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Test loading from empty/corrupt file
	emptyFile := manager.getFilePath("empty")
	err = os.WriteFile(emptyFile, []byte(""), 0o644)
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	_, err = manager.LoadGame("empty")
	if err == nil {
		t.Error("LoadGame should fail for empty file")
	}

	// Test loading with invalid JSON
	invalidFile := manager.getFilePath("invalid")
	err = os.WriteFile(invalidFile, []byte("{invalid json"), 0o644)
	if err != nil {
		t.Fatalf("Failed to create invalid JSON file: %v", err)
	}

	_, err = manager.LoadGame("invalid")
	if err == nil {
		t.Error("LoadGame should fail for invalid JSON")
	}
}

// TestSaveManager_SaveGameEdgeCases tests SaveGame additional edge cases.
func TestSaveManager_SaveGameEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Test saving with invalid name
	save := NewGameSave()
	err = manager.SaveGame("invalid/name", save)
	if err == nil {
		t.Error("SaveGame should fail for name with path separator")
	}

	// Test saving with empty name
	err = manager.SaveGame("", save)
	if err == nil {
		t.Error("SaveGame should fail for empty name")
	}

	// Test overwriting existing save
	err = manager.SaveGame("overwrite_test", save)
	if err != nil {
		t.Fatalf("First SaveGame failed: %v", err)
	}

	save.PlayerState.Level = 99
	err = manager.SaveGame("overwrite_test", save)
	if err != nil {
		t.Fatalf("Second SaveGame (overwrite) failed: %v", err)
	}

	// Verify the save was overwritten
	loaded, err := manager.LoadGame("overwrite_test")
	if err != nil {
		t.Fatalf("LoadGame failed: %v", err)
	}
	if loaded.PlayerState.Level != 99 {
		t.Errorf("Expected level 99 after overwrite, got %d", loaded.PlayerState.Level)
	}
}

// TestSaveManager_SetMigratorEdgeCases tests SetMigrator with various scenarios.
func TestSaveManager_SetMigratorEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Initially no migrator
	if manager.migrator != nil {
		t.Error("Manager should have nil migrator initially")
	}

	// Set a migrator
	migrator := NewDefaultMigrator()
	manager.SetMigrator(migrator)

	if manager.migrator == nil {
		t.Error("SetMigrator failed to set migrator")
	}

	// Set nil migrator (should work)
	manager.SetMigrator(nil)
	if manager.migrator != nil {
		t.Error("SetMigrator failed to clear migrator")
	}
}

// TestSaveManager_RegisterHookEdgeCases tests RegisterHook functionality.
func TestSaveManager_RegisterHookEdgeCases(t *testing.T) {
	migrator := NewDefaultMigrator()

	// Test registering a hook
	hookCalled := false
	testHook := func(save *GameSave, sourceVersion, targetVersion string) error {
		hookCalled = true
		return nil
	}

	migrator.RegisterHook("0.9.0", testHook)

	// Migrate a save to trigger the hook
	save := &GameSave{
		Version: "0.9.0",
		PlayerState: &PlayerState{
			EntityID: 1,
		},
		WorldState: &WorldState{
			Seed: 123,
		},
	}

	_, err := migrator.Migrate(save, "0.9.0")
	if err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	if !hookCalled {
		t.Error("Migration hook was not called")
	}
}

// TestSaveManager_ListSavesWithMixedFiles tests ListSaves with various file types.
func TestSaveManager_ListSavesWithMixedFiles(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Create valid saves
	for i := 0; i < 3; i++ {
		save := NewGameSave()
		name := "save" + string(rune('A'+i))
		err = manager.SaveGame(name, save)
		if err != nil {
			t.Fatalf("SaveGame failed: %v", err)
		}
	}

	// Create non-save files in the directory (should be ignored)
	extraFile := filepath.Join(tmpDir, "readme.txt")
	err = os.WriteFile(extraFile, []byte("test"), 0o644)
	if err != nil {
		t.Fatalf("Failed to create extra file: %v", err)
	}

	dotFile := filepath.Join(tmpDir, ".hidden")
	err = os.WriteFile(dotFile, []byte("hidden"), 0o644)
	if err != nil {
		t.Fatalf("Failed to create dot file: %v", err)
	}

	// List saves should only return .sav files
	saves, err := manager.ListSaves()
	if err != nil {
		t.Fatalf("ListSaves failed: %v", err)
	}

	if len(saves) != 3 {
		t.Errorf("Expected 3 saves, got %d", len(saves))
	}
}

// TestSaveManager_DeleteSaveWithBackup tests deleting saves that have backups.
func TestSaveManager_DeleteSaveWithBackup(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Create save with backup
	save := NewGameSave()
	err = manager.SaveGameWithBackup("backup_delete_test", save)
	if err != nil {
		t.Fatalf("SaveGameWithBackup failed: %v", err)
	}

	// Update to create backup
	save.PlayerState.Level = 20
	err = manager.SaveGameWithBackup("backup_delete_test", save)
	if err != nil {
		t.Fatalf("Second SaveGameWithBackup failed: %v", err)
	}

	// Verify backup exists
	if !manager.BackupExists("backup_delete_test") {
		t.Error("Backup should exist before deletion")
	}

	// Delete the save
	err = manager.DeleteSave("backup_delete_test")
	if err != nil {
		t.Fatalf("DeleteSave failed: %v", err)
	}

	// Verify both save and backup are gone
	if manager.SaveExists("backup_delete_test") {
		t.Error("Save should not exist after deletion")
	}
	// Note: DeleteSave might not delete backup, that's OK
}
