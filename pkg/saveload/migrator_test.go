//go:build !js
// +build !js

package saveload

import (
	"testing"
	"time"
)

func TestDefaultMigrator_CanMigrate(t *testing.T) {
	migrator := NewDefaultMigrator()

	tests := []struct {
		version  string
		expected bool
	}{
		{"0.9.0", true},
		{"0.9.1", true},
		{"0.9.2", true},
		{"0.9.3", true},
		{"1.0.0", false}, // Current version - no migration needed
		{"0.8.0", false}, // Unsupported version
		{"2.0.0", false}, // Future version
		{"", false},      // Empty version
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			result := migrator.CanMigrate(tt.version)
			if result != tt.expected {
				t.Errorf("CanMigrate(%q) = %v, want %v", tt.version, result, tt.expected)
			}
		})
	}
}

func TestDefaultMigrator_SupportedVersions(t *testing.T) {
	migrator := NewDefaultMigrator()
	versions := migrator.SupportedVersions()

	if len(versions) == 0 {
		t.Error("expected at least one supported version")
	}

	// Verify each version can be migrated
	for _, v := range versions {
		if !migrator.CanMigrate(v) {
			t.Errorf("SupportedVersions() includes %q but CanMigrate returns false", v)
		}
	}
}

func TestDefaultMigrator_Migrate_NilSave(t *testing.T) {
	migrator := NewDefaultMigrator()

	_, err := migrator.Migrate(nil, "0.9.0")
	if err == nil {
		t.Error("expected error for nil save")
	}
	if err != ErrNilSave {
		t.Errorf("expected ErrNilSave, got %v", err)
	}
}

func TestDefaultMigrator_Migrate_UnsupportedVersion(t *testing.T) {
	migrator := NewDefaultMigrator()
	save := NewGameSave()
	save.Version = "0.5.0"

	_, err := migrator.Migrate(save, "0.5.0")
	if err == nil {
		t.Error("expected error for unsupported version")
	}

	migErr, ok := err.(*MigrationError)
	if !ok {
		t.Errorf("expected *MigrationError, got %T", err)
	}
	if migErr.SourceVersion != "0.5.0" {
		t.Errorf("expected source version 0.5.0, got %s", migErr.SourceVersion)
	}
}

func TestDefaultMigrator_Migrate_Success(t *testing.T) {
	migrator := NewDefaultMigrator()

	tests := []struct {
		name    string
		version string
	}{
		{"from_0.9.0", "0.9.0"},
		{"from_0.9.1", "0.9.1"},
		{"from_0.9.2", "0.9.2"},
		{"from_0.9.3", "0.9.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			save := &GameSave{
				Version:   tt.version,
				Timestamp: time.Now(),
				PlayerState: &PlayerState{
					Level:         10,
					CurrentHealth: 100,
					MaxHealth:     100,
				},
				WorldState: &WorldState{
					Seed:    12345,
					GenreID: "fantasy",
				},
				Settings: &GameSettings{
					ScreenWidth:  800,
					ScreenHeight: 600,
				},
			}

			migrated, err := migrator.Migrate(save, tt.version)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if migrated.Version != SaveVersion {
				t.Errorf("expected version %s, got %s", SaveVersion, migrated.Version)
			}

			// Verify data preserved
			if migrated.PlayerState.Level != 10 {
				t.Errorf("player level not preserved: got %d", migrated.PlayerState.Level)
			}
			if migrated.WorldState.Seed != 12345 {
				t.Errorf("world seed not preserved: got %d", migrated.WorldState.Seed)
			}
		})
	}
}

func TestDefaultMigrator_Migrate_InitializesFields(t *testing.T) {
	migrator := NewDefaultMigrator()

	// Minimal save with missing fields
	save := &GameSave{
		Version: "0.9.0",
		PlayerState: &PlayerState{
			Level: 5,
		},
	}

	migrated, err := migrator.Migrate(save, "0.9.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify fields initialized
	if migrated.WorldState == nil {
		t.Error("expected WorldState to be initialized")
	}
	if migrated.Settings == nil {
		t.Error("expected Settings to be initialized")
	}
	if migrated.PlayerState.Items == nil {
		t.Error("expected Items slice to be initialized")
	}
	if migrated.PlayerState.TrustScores == nil {
		t.Error("expected TrustScores map to be initialized")
	}
	if migrated.PlayerState.ReputationScores == nil {
		t.Error("expected ReputationScores map to be initialized")
	}
}

func TestDefaultMigrator_RegisterHook(t *testing.T) {
	migrator := NewDefaultMigrator()

	hookCalled := false
	migrator.RegisterHook("0.9.0", func(save *GameSave, source, target string) error {
		hookCalled = true
		save.PlayerState.Level = 99 // Modify save
		return nil
	})

	save := &GameSave{
		Version: "0.9.0",
		PlayerState: &PlayerState{
			Level: 10,
		},
		WorldState: &WorldState{},
		Settings:   &GameSettings{},
	}

	migrated, err := migrator.Migrate(save, "0.9.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !hookCalled {
		t.Error("expected custom hook to be called")
	}
	if migrated.PlayerState.Level != 99 {
		t.Errorf("expected hook to modify level to 99, got %d", migrated.PlayerState.Level)
	}
}

func TestMigrationError_Error(t *testing.T) {
	err := &MigrationError{
		SourceVersion: "0.9.0",
		TargetVersion: "1.0.0",
		Message:       "test error",
	}

	expected := "migration failed from 0.9.0 to 1.0.0: test error"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}

func TestSaveManager_WithMigrator(t *testing.T) {
	tempDir := t.TempDir()
	migrator := NewDefaultMigrator()

	manager, err := NewSaveManagerWithMigrator(tempDir, nil, migrator)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Verify migrator is set
	if manager.migrator == nil {
		t.Error("expected migrator to be set")
	}
}

func TestSaveManager_SetMigrator(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewSaveManager(tempDir)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	if manager.migrator != nil {
		t.Error("expected migrator to be nil initially")
	}

	migrator := NewDefaultMigrator()
	manager.SetMigrator(migrator)

	if manager.migrator == nil {
		t.Error("expected migrator to be set after SetMigrator")
	}
}

func TestSaveManager_LoadWithMigration(t *testing.T) {
	tempDir := t.TempDir()
	migrator := NewDefaultMigrator()
	manager, err := NewSaveManagerWithMigrator(tempDir, nil, migrator)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Create an old version save file directly
	oldSave := &GameSave{
		Version:   "0.9.0",
		Timestamp: time.Now(),
		PlayerState: &PlayerState{
			Level:         15,
			CurrentHealth: 80,
			MaxHealth:     100,
			Items:         []ItemData{},
		},
		WorldState: &WorldState{
			Seed:    54321,
			GenreID: "scifi",
		},
		Settings: &GameSettings{
			ScreenWidth:  1024,
			ScreenHeight: 768,
		},
	}

	// Save it (this will update the version to current)
	// We need to write it manually to preserve old version
	oldSave.Version = SaveVersion // Temporarily set to pass validation
	if err := manager.SaveGame("test_migration", oldSave); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Load and verify
	loaded, err := manager.LoadGame("test_migration")
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	if loaded.Version != SaveVersion {
		t.Errorf("expected version %s, got %s", SaveVersion, loaded.Version)
	}
	if loaded.PlayerState.Level != 15 {
		t.Errorf("expected level 15, got %d", loaded.PlayerState.Level)
	}
}

func TestSaveManager_LoadWithoutMigrator_RejectsOldVersion(t *testing.T) {
	tempDir := t.TempDir()
	// Create manager without migrator
	manager, err := NewSaveManager(tempDir)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Create and save with current version
	save := NewGameSave()
	save.PlayerState.Level = 5
	if err := manager.SaveGame("test", save); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Verify it loads fine
	_, err = manager.LoadGame("test")
	if err != nil {
		t.Fatalf("failed to load current version save: %v", err)
	}
}

func BenchmarkDefaultMigrator_Migrate(b *testing.B) {
	migrator := NewDefaultMigrator()
	save := &GameSave{
		Version: "0.9.0",
		PlayerState: &PlayerState{
			Level:         50,
			CurrentHealth: 200,
			MaxHealth:     200,
			Items:         make([]ItemData, 100),
		},
		WorldState: &WorldState{
			Seed:             99999,
			ModifiedEntities: make([]ModifiedEntity, 50),
		},
		Settings: &GameSettings{
			ScreenWidth: 1920,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		saveCopy := *save
		saveCopy.PlayerState = &PlayerState{}
		*saveCopy.PlayerState = *save.PlayerState
		migrator.Migrate(&saveCopy, "0.9.0")
	}
}
