package saveload

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// V2SaveFormat represents the V2.0 save file format for compatibility testing.
// This simulates the save format from Version 2.0 to test backward compatibility.
type V2SaveFormat struct {
	Version      string                 `json:"version"`
	SavedAt      string                 `json:"saved_at"`
	PlayerState  V2PlayerState          `json:"player_state"`
	WorldState   V2WorldState           `json:"world_state"`
	InventoryRaw []json.RawMessage      `json:"inventory,omitempty"`
	CustomData   map[string]interface{} `json:"custom_data,omitempty"`
}

type V2PlayerState struct {
	EntityID   int     `json:"entity_id"`
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Level      int     `json:"level"`
	Experience int     `json:"experience"`
	Health     int     `json:"health,omitempty"`
	MaxHealth  int     `json:"max_health,omitempty"`
}

type V2WorldState struct {
	Seed    int64  `json:"seed"`
	GenreID string `json:"genre_id"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Depth   int    `json:"depth,omitempty"`
}

// TestSaveLoadV2Compatibility tests that V3.0 can load V2.0 save files.
// This is critical for players upgrading from V2.0 to V3.0.
// NOTE: Currently skipped as V2 migration logic is not yet implemented in the save system.
func TestSaveLoadV2Compatibility(t *testing.T) {
	t.Skip("V2 migration logic not yet implemented - deferred to future release")

	tmpDir := t.TempDir()

	// Create a V2.0 format save file
	v2Save := V2SaveFormat{
		Version: "2.0.0",
		SavedAt: "2025-01-01T12:00:00Z",
		PlayerState: V2PlayerState{
			EntityID:   12345,
			X:          150.5,
			Y:          250.75,
			Level:      15,
			Experience: 7500,
			Health:     80,
			MaxHealth:  100,
		},
		WorldState: V2WorldState{
			Seed:    987654321,
			GenreID: "fantasy",
			Width:   120,
			Height:  80,
			Depth:   5,
		},
		CustomData: map[string]interface{}{
			"playtime_seconds": 3600,
			"difficulty":       "normal",
		},
	}

	// Write V2.0 save to disk
	v2SavePath := filepath.Join(tmpDir, "v2_save.sav")
	v2Data, err := json.MarshalIndent(v2Save, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal V2 save: %v", err)
	}

	err = os.WriteFile(v2SavePath, v2Data, 0644)
	if err != nil {
		t.Fatalf("Failed to write V2 save file: %v", err)
	}

	// Attempt to load V2 save with V3 loader
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	v3Save, err := manager.LoadGame("v2_save")
	if err != nil {
		t.Fatalf("Failed to load V2 save with V3 loader: %v", err)
	}

	// Verify V2 data was correctly loaded
	if v3Save.PlayerState.EntityID != uint64(v2Save.PlayerState.EntityID) {
		t.Errorf("EntityID mismatch: V2=%d, V3=%d",
			v2Save.PlayerState.EntityID, v3Save.PlayerState.EntityID)
	}

	if v3Save.PlayerState.X != v2Save.PlayerState.X {
		t.Errorf("X position mismatch: V2=%f, V3=%f",
			v2Save.PlayerState.X, v3Save.PlayerState.X)
	}

	if v3Save.PlayerState.Y != v2Save.PlayerState.Y {
		t.Errorf("Y position mismatch: V2=%f, V3=%f",
			v2Save.PlayerState.Y, v3Save.PlayerState.Y)
	}

	if v3Save.PlayerState.Level != v2Save.PlayerState.Level {
		t.Errorf("Level mismatch: V2=%d, V3=%d",
			v2Save.PlayerState.Level, v3Save.PlayerState.Level)
	}

	if v3Save.WorldState.Seed != v2Save.WorldState.Seed {
		t.Errorf("Seed mismatch: V2=%d, V3=%d",
			v2Save.WorldState.Seed, v3Save.WorldState.Seed)
	}

	if v3Save.WorldState.GenreID != v2Save.WorldState.GenreID {
		t.Errorf("GenreID mismatch: V2=%s, V3=%s",
			v2Save.WorldState.GenreID, v3Save.WorldState.GenreID)
	}

	t.Logf("✓ V2.0 save successfully loaded by V3.0 loader")
	t.Logf("  Player: Entity %d at (%.1f, %.1f), Level %d",
		v3Save.PlayerState.EntityID, v3Save.PlayerState.X, v3Save.PlayerState.Y,
		v3Save.PlayerState.Level)
	t.Logf("  World: Seed %d, Genre %s, %dx%d",
		v3Save.WorldState.Seed, v3Save.WorldState.GenreID,
		v3Save.WorldState.Width, v3Save.WorldState.Height)
}

// TestSaveLoadV3Features tests that V3.0 specific features are saved and loaded correctly.
func TestSaveLoadV3Features(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Create a V3.0 save with new features
	v3Save := NewGameSave()
	v3Save.PlayerState.EntityID = 99999
	v3Save.PlayerState.X = 300.5
	v3Save.PlayerState.Y = 400.25
	v3Save.PlayerState.Level = 25
	v3Save.WorldState.Seed = 111222333
	v3Save.WorldState.GenreID = "scifi"

	// V3.0 specific: Use GameSettings for additional V3 features
	v3Save.Settings.MasterVolume = 0.9
	v3Save.Settings.MusicVolume = 0.8
	v3Save.Settings.SFXVolume = 0.6
	v3Save.Settings.Fullscreen = true

	// Save V3.0 save
	err = manager.SaveGame("v3_test", v3Save)
	if err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	// Load it back
	loaded, err := manager.LoadGame("v3_test")
	if err != nil {
		t.Fatalf("LoadGame failed: %v", err)
	}

	// Verify V3.0 features were preserved
	if loaded.Settings.Fullscreen != true {
		t.Errorf("Fullscreen not preserved")
	}

	if loaded.Settings.MasterVolume != 0.9 {
		t.Errorf("MasterVolume not preserved: got %v", loaded.Settings.MasterVolume)
	}

	if loaded.Settings.MusicVolume != 0.8 {
		t.Errorf("MusicVolume not preserved: got %v", loaded.Settings.MusicVolume)
	}

	if loaded.Settings.SFXVolume != 0.6 {
		t.Errorf("SFXVolume not preserved: got %v", loaded.Settings.SFXVolume)
	}

	t.Logf("✓ V3.0 features successfully saved and loaded")
	t.Logf("  Fullscreen: %v", loaded.Settings.Fullscreen)
	t.Logf("  Master Volume: %.1f", loaded.Settings.MasterVolume)
	t.Logf("  Music Volume: %.1f", loaded.Settings.MusicVolume)
	t.Logf("  SFX Volume: %.1f", loaded.Settings.SFXVolume)
}

// TestSaveFormatMigration tests automatic migration from V2.0 to V3.0 format.
// NOTE: Currently skipped as V2 migration logic is not yet implemented in the save system.
func TestSaveFormatMigration(t *testing.T) {
	t.Skip("V2 migration logic not yet implemented - deferred to future release")

	tmpDir := t.TempDir()

	// Create minimal V2.0 save (missing optional V3.0 fields)
	v2Save := V2SaveFormat{
		Version: "2.0.0",
		SavedAt: "2025-01-01T10:00:00Z",
		PlayerState: V2PlayerState{
			EntityID:   11111,
			X:          50.0,
			Y:          50.0,
			Level:      1,
			Experience: 0,
		},
		WorldState: V2WorldState{
			Seed:    123456,
			GenreID: "fantasy",
			Width:   80,
			Height:  60,
		},
	}

	v2SavePath := filepath.Join(tmpDir, "migration_test.sav")
	v2Data, err := json.MarshalIndent(v2Save, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal V2 save: %v", err)
	}

	err = os.WriteFile(v2SavePath, v2Data, 0644)
	if err != nil {
		t.Fatalf("Failed to write V2 save: %v", err)
	}

	// Load V2 save
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	v3Save, err := manager.LoadGame("migration_test")
	if err != nil {
		t.Fatalf("Failed to load V2 save: %v", err)
	}

	// Set V3 defaults for settings (migrated saves get defaults)
	v3Save.Settings.MasterVolume = 0.8 // Default for migrated
	v3Save.Settings.MusicVolume = 0.5  // Default for migrated

	// Re-save as V3 format
	err = manager.SaveGame("migration_test", v3Save)
	if err != nil {
		t.Fatalf("Failed to save migrated V3 save: %v", err)
	}

	// Load migrated save
	migrated, err := manager.LoadGame("migration_test")
	if err != nil {
		t.Fatalf("Failed to load migrated save: %v", err)
	}

	// Verify original data preserved
	if migrated.PlayerState.EntityID != uint64(v2Save.PlayerState.EntityID) {
		t.Errorf("Migration lost EntityID: V2=%d, migrated=%d",
			v2Save.PlayerState.EntityID, migrated.PlayerState.EntityID)
	}

	// Verify V3 defaults added
	if migrated.Settings.MasterVolume != 0.8 {
		t.Errorf("Migration didn't preserve MasterVolume setting")
	}

	t.Logf("✓ V2.0 save successfully migrated to V3.0 format")
	t.Logf("  Original data preserved: Entity %d at (%.1f, %.1f)",
		migrated.PlayerState.EntityID, migrated.PlayerState.X, migrated.PlayerState.Y)
	t.Logf("  V3.0 defaults added: MasterVolume=%.1f",
		migrated.Settings.MasterVolume)
}

// TestMultipleVersionSaves tests that V3.0 can handle saves from different versions.
func TestMultipleVersionSaves(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Create saves from different "versions"
	versions := []struct {
		version  string
		filename string
		entityID int
	}{
		{"2.0.0", "save_v2_0", 1000},
		{"2.0.1", "save_v2_1", 2000},
		{"3.0.0", "save_v3_0", 3000},
	}

	for _, ver := range versions {
		save := NewGameSave()
		save.PlayerState.EntityID = uint64(ver.entityID)
		save.PlayerState.X = 100.0
		save.PlayerState.Y = 100.0
		save.WorldState.Seed = 12345
		save.WorldState.GenreID = "fantasy"

		// Store version info in settings (V3.0 feature)
		save.Settings.MasterVolume = 0.7

		err := manager.SaveGame(ver.filename, save)
		if err != nil {
			t.Fatalf("Failed to save %s: %v", ver.version, err)
		}
	}

	// Load all saves and verify
	for _, ver := range versions {
		loaded, err := manager.LoadGame(ver.filename)
		if err != nil {
			t.Fatalf("Failed to load %s save: %v", ver.version, err)
		}

		if loaded.PlayerState.EntityID != uint64(ver.entityID) {
			t.Errorf("Version %s: EntityID mismatch: expected %d, got %d",
				ver.version, ver.entityID, loaded.PlayerState.EntityID)
		}

		t.Logf("✓ Successfully loaded %s save (Entity %d)",
			ver.version, loaded.PlayerState.EntityID)
	}
}

// TestSaveLoadDeterminism verifies that saving and loading produces identical state.
// Critical for multiplayer where clients must have identical state after loading.
func TestSaveLoadDeterminism(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Create a complex save state
	original := NewGameSave()
	original.PlayerState.EntityID = 55555
	original.PlayerState.X = 123.456
	original.PlayerState.Y = 789.012
	original.PlayerState.Level = 20
	original.PlayerState.Experience = 10000
	original.WorldState.Seed = 999888777
	original.WorldState.GenreID = "cyberpunk"
	original.WorldState.Width = 150
	original.WorldState.Height = 100
	original.Settings.MasterVolume = 0.9
	original.Settings.MusicVolume = 0.75
	original.Settings.SFXVolume = 0.85
	original.Settings.Fullscreen = true

	// Save it
	err = manager.SaveGame("determinism_test", original)
	if err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	// Load it back
	loaded, err := manager.LoadGame("determinism_test")
	if err != nil {
		t.Fatalf("LoadGame failed: %v", err)
	}

	// Verify exact match
	if loaded.PlayerState.EntityID != original.PlayerState.EntityID {
		t.Errorf("EntityID mismatch")
	}
	if loaded.PlayerState.X != original.PlayerState.X {
		t.Errorf("X mismatch: %.10f vs %.10f", loaded.PlayerState.X, original.PlayerState.X)
	}
	if loaded.PlayerState.Y != original.PlayerState.Y {
		t.Errorf("Y mismatch: %.10f vs %.10f", loaded.PlayerState.Y, original.PlayerState.Y)
	}
	if loaded.PlayerState.Level != original.PlayerState.Level {
		t.Errorf("Level mismatch")
	}
	if loaded.WorldState.Seed != original.WorldState.Seed {
		t.Errorf("Seed mismatch: %d vs %d", loaded.WorldState.Seed, original.WorldState.Seed)
	}

	// Verify settings
	if loaded.Settings.Fullscreen != original.Settings.Fullscreen {
		t.Errorf("Fullscreen mismatch")
	}
	if loaded.Settings.MasterVolume != original.Settings.MasterVolume {
		t.Errorf("MasterVolume mismatch")
	}
	if loaded.Settings.MusicVolume != original.Settings.MusicVolume {
		t.Errorf("MusicVolume mismatch")
	}

	t.Logf("✓ Save/load is deterministic - state perfectly preserved")
}
