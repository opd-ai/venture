//go:build !js
// +build !js

package saveload

import (
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestMemorySaveManager_SaveAndLoad(t *testing.T) {
	m := NewMemorySaveManager()

	// Create a test save
	save := &GameSave{
		PlayerState: &PlayerState{
			Level: 10,
			Gold:  500,
		},
		WorldState: &WorldState{
			Seed:     12345,
			GenreID:  "fantasy",
			GameTime: 3600.0,
		},
	}

	// Save the game
	err := m.SaveGame("testsave", save)
	if err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	// Load the game
	loaded, err := m.LoadGame("testsave")
	if err != nil {
		t.Fatalf("LoadGame failed: %v", err)
	}

	// Verify data
	if loaded.PlayerState.Level != 10 {
		t.Errorf("Expected level 10, got %d", loaded.PlayerState.Level)
	}
	if loaded.PlayerState.Gold != 500 {
		t.Errorf("Expected gold 500, got %d", loaded.PlayerState.Gold)
	}
	if loaded.WorldState.Seed != 12345 {
		t.Errorf("Expected seed 12345, got %d", loaded.WorldState.Seed)
	}
	if loaded.Version != SaveVersion {
		t.Errorf("Expected version %s, got %s", SaveVersion, loaded.Version)
	}
}

func TestMemorySaveManager_SaveNil(t *testing.T) {
	m := NewMemorySaveManager()

	err := m.SaveGame("test", nil)
	if err == nil {
		t.Error("Expected error for nil save, got nil")
	}
}

func TestMemorySaveManager_SaveEmptyName(t *testing.T) {
	m := NewMemorySaveManager()

	save := &GameSave{}
	err := m.SaveGame("", save)
	if err == nil {
		t.Error("Expected error for empty name, got nil")
	}
}

func TestMemorySaveManager_LoadNonexistent(t *testing.T) {
	m := NewMemorySaveManager()

	_, err := m.LoadGame("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent save, got nil")
	}
}

func TestMemorySaveManager_LoadEmptyName(t *testing.T) {
	m := NewMemorySaveManager()

	_, err := m.LoadGame("")
	if err == nil {
		t.Error("Expected error for empty name, got nil")
	}
}

func TestMemorySaveManager_DeleteSave(t *testing.T) {
	m := NewMemorySaveManager()

	// Save a game
	save := &GameSave{}
	_ = m.SaveGame("todelete", save)

	// Verify it exists
	if !m.SaveExists("todelete") {
		t.Error("Save should exist after saving")
	}

	// Delete it
	err := m.DeleteSave("todelete")
	if err != nil {
		t.Fatalf("DeleteSave failed: %v", err)
	}

	// Verify it's gone
	if m.SaveExists("todelete") {
		t.Error("Save should not exist after deleting")
	}
}

func TestMemorySaveManager_DeleteNonexistent(t *testing.T) {
	m := NewMemorySaveManager()

	err := m.DeleteSave("nonexistent")
	if err == nil {
		t.Error("Expected error for deleting nonexistent save, got nil")
	}
}

func TestMemorySaveManager_ListSaves(t *testing.T) {
	m := NewMemorySaveManager()

	// Create multiple saves with different timestamps
	saves := []struct {
		name  string
		level int
	}{
		{"save1", 1},
		{"save2", 5},
		{"save3", 10},
	}

	for _, s := range saves {
		save := &GameSave{
			PlayerState: &PlayerState{Level: s.level},
			WorldState:  &WorldState{GenreID: "fantasy"},
		}
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
		_ = m.SaveGame(s.name, save)
	}

	// List saves
	list, err := m.ListSaves()
	if err != nil {
		t.Fatalf("ListSaves failed: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("Expected 3 saves, got %d", len(list))
	}

	// Verify sorted by timestamp descending (newest first)
	if len(list) >= 2 {
		if list[0].Timestamp.Before(list[1].Timestamp) {
			t.Error("ListSaves should return newest first")
		}
	}
}

func TestMemorySaveManager_GetSaveMetadata(t *testing.T) {
	m := NewMemorySaveManager()

	save := &GameSave{
		PlayerState: &PlayerState{Level: 25},
		WorldState: &WorldState{
			GenreID:  "sci-fi",
			GameTime: 7200.0,
		},
	}
	_ = m.SaveGame("metadata_test", save)

	meta, err := m.GetSaveMetadata("metadata_test")
	if err != nil {
		t.Fatalf("GetSaveMetadata failed: %v", err)
	}

	if meta.Name != "metadata_test" {
		t.Errorf("Expected name 'metadata_test', got '%s'", meta.Name)
	}
	if meta.PlayerLevel != 25 {
		t.Errorf("Expected level 25, got %d", meta.PlayerLevel)
	}
	if meta.GenreID != "sci-fi" {
		t.Errorf("Expected genre 'sci-fi', got '%s'", meta.GenreID)
	}
	if meta.GameTime != 7200.0 {
		t.Errorf("Expected game time 7200, got %f", meta.GameTime)
	}
}

func TestMemorySaveManager_GetSaveMetadataNonexistent(t *testing.T) {
	m := NewMemorySaveManager()

	_, err := m.GetSaveMetadata("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent save, got nil")
	}
}

func TestMemorySaveManager_SaveExists(t *testing.T) {
	m := NewMemorySaveManager()

	if m.SaveExists("test") {
		t.Error("Save should not exist before saving")
	}

	_ = m.SaveGame("test", &GameSave{})

	if !m.SaveExists("test") {
		t.Error("Save should exist after saving")
	}
}

func TestMemorySaveManager_SetMigrator(t *testing.T) {
	m := NewMemorySaveManager()

	// SetMigrator should be a no-op and not panic
	m.SetMigrator(nil)
}

func TestMemorySaveManager_ConcurrentAccess(t *testing.T) {
	m := NewMemorySaveManager()

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent saves
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			save := &GameSave{
				PlayerState: &PlayerState{Level: idx},
			}
			_ = m.SaveGame("concurrent", save)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = m.LoadGame("concurrent")
			_ = m.SaveExists("concurrent")
			_, _ = m.ListSaves()
		}()
	}

	wg.Wait()

	// Verify save exists
	if !m.SaveExists("concurrent") {
		t.Error("Save should exist after concurrent operations")
	}
}

func TestMemorySaveManager_WithLogger(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	m := NewMemorySaveManagerWithLogger(logger)

	save := &GameSave{
		PlayerState: &PlayerState{Level: 5},
	}

	// These operations should log messages
	_ = m.SaveGame("logged", save)
	_, _ = m.LoadGame("logged")
	_ = m.DeleteSave("logged")

	// If we get here without panicking, logging works
}

func TestMemorySaveManager_ImplementsInterface(t *testing.T) {
	// Verify MemorySaveManager implements Manager interface
	var _ Manager = (*MemorySaveManager)(nil)
}

func BenchmarkMemorySaveManager_SaveGame(b *testing.B) {
	m := NewMemorySaveManager()
	save := &GameSave{
		PlayerState: &PlayerState{Level: 10, Gold: 500},
		WorldState:  &WorldState{Seed: 12345, GenreID: "fantasy"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.SaveGame("bench", save)
	}
}

func BenchmarkMemorySaveManager_LoadGame(b *testing.B) {
	m := NewMemorySaveManager()
	save := &GameSave{
		PlayerState: &PlayerState{Level: 10, Gold: 500},
		WorldState:  &WorldState{Seed: 12345, GenreID: "fantasy"},
	}
	_ = m.SaveGame("bench", save)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.LoadGame("bench")
	}
}
