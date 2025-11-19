package world

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewWorldPersistence(t *testing.T) {
	wp := NewWorldPersistence("/tmp/test.world")

	if wp.SavePath != "/tmp/test.world" {
		t.Errorf("SavePath = %s, want /tmp/test.world", wp.SavePath)
	}

	if wp.AutoSaveInterval != 300.0 {
		t.Errorf("AutoSaveInterval = %f, want 300.0", wp.AutoSaveInterval)
	}

	if wp.maxBackups != 3 {
		t.Errorf("maxBackups = %d, want 3", wp.maxBackups)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "test.world")

	wp := NewWorldPersistence(savePath)

	// Create test state
	originalState := &PersistentWorldState{
		Version:   CurrentSchemaVersion,
		WorldSeed: 12345,
		ChunkData: map[string]*Chunk{
			"0,0": {
				X: 0,
				Y: 0,
				Modifications: []TerrainMod{
					{Type: "explosion", X: 10, Y: 20, Radius: 5.0, Timestamp: 1000},
				},
			},
		},
		Entities: []*EntityState{
			{ID: 1, TypeName: "Monster", Components: map[string]interface{}{"health": 100}},
		},
		WorldEvents: []WorldEvent{
			{Type: "war", Timestamp: 2000, Data: map[string]interface{}{"faction": "red"}},
		},
		Timestamp:      time.Now().UnixMilli(),
		ModifiedChunks: make(map[string]bool),
	}

	// Save
	if err := wp.SaveWorld(originalState); err != nil {
		t.Fatalf("SaveWorld failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		t.Fatalf("Save file was not created")
	}

	// Load
	loadedState, err := wp.LoadWorld(12345)
	if err != nil {
		t.Fatalf("LoadWorld failed: %v", err)
	}

	// Verify loaded state matches original
	if loadedState.Version != originalState.Version {
		t.Errorf("Version = %d, want %d", loadedState.Version, originalState.Version)
	}

	if loadedState.WorldSeed != originalState.WorldSeed {
		t.Errorf("WorldSeed = %d, want %d", loadedState.WorldSeed, originalState.WorldSeed)
	}

	if len(loadedState.ChunkData) != len(originalState.ChunkData) {
		t.Errorf("ChunkData length = %d, want %d", len(loadedState.ChunkData), len(originalState.ChunkData))
	}

	if len(loadedState.Entities) != len(originalState.Entities) {
		t.Errorf("Entities length = %d, want %d", len(loadedState.Entities), len(originalState.Entities))
	}

	if len(loadedState.WorldEvents) != len(originalState.WorldEvents) {
		t.Errorf("WorldEvents length = %d, want %d", len(loadedState.WorldEvents), len(originalState.WorldEvents))
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "nonexistent.world")

	wp := NewWorldPersistence(savePath)

	// Load should create new state when file doesn't exist
	state, err := wp.LoadWorld(67890)
	if err != nil {
		t.Fatalf("LoadWorld failed: %v", err)
	}

	if state.Version != CurrentSchemaVersion {
		t.Errorf("Version = %d, want %d", state.Version, CurrentSchemaVersion)
	}

	if state.WorldSeed != 67890 {
		t.Errorf("WorldSeed = %d, want 67890", state.WorldSeed)
	}

	if len(state.ChunkData) != 0 {
		t.Errorf("ChunkData should be empty, got %d entries", len(state.ChunkData))
	}
}

func TestBackupRotation(t *testing.T) {
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "test.world")

	wp := NewWorldPersistence(savePath)

	// Create initial state
	state1 := &PersistentWorldState{
		Version:        CurrentSchemaVersion,
		WorldSeed:      11111,
		ChunkData:      make(map[string]*Chunk),
		Entities:       []*EntityState{},
		WorldEvents:    []WorldEvent{},
		ModifiedChunks: make(map[string]bool),
	}

	// Save first time
	if err := wp.SaveWorld(state1); err != nil {
		t.Fatalf("SaveWorld 1 failed: %v", err)
	}

	// Wait a bit to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	// Save second time (should create .1 backup)
	state2 := &PersistentWorldState{
		Version:        CurrentSchemaVersion,
		WorldSeed:      22222,
		ChunkData:      make(map[string]*Chunk),
		Entities:       []*EntityState{},
		WorldEvents:    []WorldEvent{},
		ModifiedChunks: make(map[string]bool),
	}

	if err := wp.SaveWorld(state2); err != nil {
		t.Fatalf("SaveWorld 2 failed: %v", err)
	}

	// Verify backup exists
	backup1Path := savePath + ".1"
	if _, err := os.Stat(backup1Path); os.IsNotExist(err) {
		t.Fatalf("Backup .1 was not created")
	}

	// Load from backup and verify it's the first save
	loadedBackup, err := wp.loadFromPath(backup1Path, 0)
	if err != nil {
		t.Fatalf("Loading backup failed: %v", err)
	}

	if loadedBackup.WorldSeed != 11111 {
		t.Errorf("Backup WorldSeed = %d, want 11111", loadedBackup.WorldSeed)
	}

	// Verify current save is the second one
	loadedCurrent, err := wp.LoadWorld(0)
	if err != nil {
		t.Fatalf("Loading current failed: %v", err)
	}

	if loadedCurrent.WorldSeed != 22222 {
		t.Errorf("Current WorldSeed = %d, want 22222", loadedCurrent.WorldSeed)
	}
}

func TestBackupRotationMultiple(t *testing.T) {
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "test.world")

	wp := NewWorldPersistence(savePath)

	// Save 5 times (more than maxBackups)
	for i := 1; i <= 5; i++ {
		state := &PersistentWorldState{
			Version:        CurrentSchemaVersion,
			WorldSeed:      int64(i * 1111),
			ChunkData:      make(map[string]*Chunk),
			Entities:       []*EntityState{},
			WorldEvents:    []WorldEvent{},
			ModifiedChunks: make(map[string]bool),
		}

		if err := wp.SaveWorld(state); err != nil {
			t.Fatalf("SaveWorld %d failed: %v", i, err)
		}

		time.Sleep(10 * time.Millisecond)
	}

	// Should only have maxBackups (3) backups
	backups := wp.ListBackups()
	if len(backups) != 3 {
		t.Errorf("Expected 3 backups, got %d", len(backups))
	}

	// Verify backup .4 doesn't exist (oldest should be deleted)
	backup4Path := savePath + ".4"
	if _, err := os.Stat(backup4Path); !os.IsNotExist(err) {
		t.Errorf("Backup .4 should not exist")
	}
}

// OBSOLETE CODE REMOVED: TestMigration
// Replaced by: Pre-1.0 policy - incompatible saves are rejected with error
// Removed: Migration test for upgrading version 0 to version 1
// PRE-1.0: Only current schema version supported, old saves are rejected

func TestIncrementalSave(t *testing.T) {
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "test.world")

	wp := NewWorldPersistence(savePath)

	// Create state with 100 chunks
	state := &PersistentWorldState{
		Version:        CurrentSchemaVersion,
		WorldSeed:      12345,
		ChunkData:      make(map[string]*Chunk),
		Entities:       []*EntityState{},
		WorldEvents:    []WorldEvent{},
		ModifiedChunks: make(map[string]bool),
	}

	for i := 0; i < 100; i++ {
		chunkID := fmt.Sprintf("%d,0", i)
		state.ChunkData[chunkID] = &Chunk{X: i, Y: 0}
	}

	// Do full save
	if err := wp.SaveWorld(state); err != nil {
		t.Fatalf("SaveWorld failed: %v", err)
	}

	// Modify only 10 chunks
	for i := 0; i < 10; i++ {
		chunkID := fmt.Sprintf("%d,0", i)
		state.ModifiedChunks[chunkID] = true
		state.ChunkData[chunkID].Modifications = append(
			state.ChunkData[chunkID].Modifications,
			TerrainMod{Type: "dig", X: i, Y: 0},
		)
	}

	// Incremental save should only save modified chunks
	if err := wp.SaveIncremental(state); err != nil {
		t.Fatalf("SaveIncremental failed: %v", err)
	}

	// Load and verify
	loadedState, err := wp.LoadWorld(12345)
	if err != nil {
		t.Fatalf("LoadWorld failed: %v", err)
	}

	// Should have modifications
	if len(loadedState.ChunkData) < 10 {
		t.Errorf("Expected at least 10 chunks, got %d", len(loadedState.ChunkData))
	}
}

func TestIncrementalSaveFallback(t *testing.T) {
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "test.world")

	wp := NewWorldPersistence(savePath)

	// Create state with 100 chunks
	state := &PersistentWorldState{
		Version:        CurrentSchemaVersion,
		WorldSeed:      12345,
		ChunkData:      make(map[string]*Chunk),
		Entities:       []*EntityState{},
		WorldEvents:    []WorldEvent{},
		ModifiedChunks: make(map[string]bool),
	}

	for i := 0; i < 100; i++ {
		chunkID := fmt.Sprintf("%d,0", i)
		state.ChunkData[chunkID] = &Chunk{X: i, Y: 0}
	}

	// Do full save
	if err := wp.SaveWorld(state); err != nil {
		t.Fatalf("SaveWorld failed: %v", err)
	}

	// Modify more than 50% of chunks (51+)
	for i := 0; i < 60; i++ {
		chunkID := fmt.Sprintf("%d,0", i)
		state.ModifiedChunks[chunkID] = true
	}

	// Should fall back to full save when >50% modified
	if err := wp.SaveIncremental(state); err != nil {
		t.Fatalf("SaveIncremental failed: %v", err)
	}

	// Verify save completed successfully
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		t.Fatalf("Save file was not created")
	}
}

func TestCleanupBackups(t *testing.T) {
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "test.world")

	wp := NewWorldPersistence(savePath)

	// Create multiple saves to generate backups
	for i := 1; i <= 4; i++ {
		state := &PersistentWorldState{
			Version:        CurrentSchemaVersion,
			WorldSeed:      int64(i),
			ChunkData:      make(map[string]*Chunk),
			Entities:       []*EntityState{},
			WorldEvents:    []WorldEvent{},
			ModifiedChunks: make(map[string]bool),
		}

		if err := wp.SaveWorld(state); err != nil {
			t.Fatalf("SaveWorld %d failed: %v", i, err)
		}

		time.Sleep(10 * time.Millisecond)
	}

	// Verify backups exist
	backups := wp.ListBackups()
	if len(backups) == 0 {
		t.Fatalf("No backups found before cleanup")
	}

	// Cleanup
	if err := wp.CleanupBackups(); err != nil {
		t.Fatalf("CleanupBackups failed: %v", err)
	}

	// Verify backups are gone
	backups = wp.ListBackups()
	if len(backups) != 0 {
		t.Errorf("Expected 0 backups after cleanup, got %d", len(backups))
	}
}

func TestListBackups(t *testing.T) {
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "test.world")

	wp := NewWorldPersistence(savePath)

	// Initially no backups
	backups := wp.ListBackups()
	if len(backups) != 0 {
		t.Errorf("Expected 0 backups initially, got %d", len(backups))
	}

	// Create saves
	for i := 1; i <= 3; i++ {
		state := &PersistentWorldState{
			Version:        CurrentSchemaVersion,
			WorldSeed:      int64(i),
			ChunkData:      make(map[string]*Chunk),
			Entities:       []*EntityState{},
			WorldEvents:    []WorldEvent{},
			ModifiedChunks: make(map[string]bool),
		}

		if err := wp.SaveWorld(state); err != nil {
			t.Fatalf("SaveWorld %d failed: %v", i, err)
		}

		time.Sleep(10 * time.Millisecond)
	}

	// Should have 2 backups (3 saves - 1 current)
	backups = wp.ListBackups()
	if len(backups) != 2 {
		t.Errorf("Expected 2 backups, got %d", len(backups))
	}

	// Verify backups are sorted by modification time (newest first)
	for i := 0; i < len(backups)-1; i++ {
		infoI, _ := os.Stat(backups[i])
		infoJ, _ := os.Stat(backups[i+1])

		if infoI.ModTime().Before(infoJ.ModTime()) {
			t.Errorf("Backups not sorted correctly: %s is older than %s", backups[i], backups[i+1])
		}
	}
}

func TestUpdate(t *testing.T) {
	wp := NewWorldPersistence("/tmp/test.world")

	// Set short auto-save interval for testing
	wp.AutoSaveInterval = 1.0 // 1 second

	// Update with less than interval
	wp.Update(0.5)
	if wp.timeSinceLastSave != 0.5 {
		t.Errorf("timeSinceLastSave = %f, want 0.5", wp.timeSinceLastSave)
	}

	// Update again to exceed interval
	wp.Update(0.6)
	if wp.timeSinceLastSave != 0.0 {
		t.Errorf("timeSinceLastSave = %f, want 0.0 (should reset after auto-save)", wp.timeSinceLastSave)
	}
}

// Benchmark tests
func BenchmarkSaveWorld(b *testing.B) {
	tmpDir := b.TempDir()
	savePath := filepath.Join(tmpDir, "bench.world")

	wp := NewWorldPersistence(savePath)

	// Create realistic state (100 modified chunks)
	state := &PersistentWorldState{
		Version:        CurrentSchemaVersion,
		WorldSeed:      12345,
		ChunkData:      make(map[string]*Chunk),
		Entities:       make([]*EntityState, 50),
		WorldEvents:    make([]WorldEvent, 10),
		ModifiedChunks: make(map[string]bool),
	}

	for i := 0; i < 100; i++ {
		chunkID := fmt.Sprintf("%d,%d", i%10, i/10)
		state.ChunkData[chunkID] = &Chunk{
			X:             i % 10,
			Y:             i / 10,
			Modifications: []TerrainMod{{Type: "dig", X: i, Y: i}},
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := wp.SaveWorld(state); err != nil {
			b.Fatalf("SaveWorld failed: %v", err)
		}
	}
}

func BenchmarkLoadWorld(b *testing.B) {
	tmpDir := b.TempDir()
	savePath := filepath.Join(tmpDir, "bench.world")

	wp := NewWorldPersistence(savePath)

	// Create and save test state
	state := &PersistentWorldState{
		Version:        CurrentSchemaVersion,
		WorldSeed:      12345,
		ChunkData:      make(map[string]*Chunk),
		Entities:       make([]*EntityState, 50),
		WorldEvents:    make([]WorldEvent, 10),
		ModifiedChunks: make(map[string]bool),
	}

	for i := 0; i < 100; i++ {
		chunkID := fmt.Sprintf("%d,%d", i%10, i/10)
		state.ChunkData[chunkID] = &Chunk{
			X:             i % 10,
			Y:             i / 10,
			Modifications: []TerrainMod{{Type: "dig", X: i, Y: i}},
		}
	}

	if err := wp.SaveWorld(state); err != nil {
		b.Fatalf("SaveWorld failed: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := wp.LoadWorld(12345); err != nil {
			b.Fatalf("LoadWorld failed: %v", err)
		}
	}
}

func BenchmarkIncrementalSave(b *testing.B) {
	tmpDir := b.TempDir()
	savePath := filepath.Join(tmpDir, "bench.world")

	wp := NewWorldPersistence(savePath)

	// Create state with 100 chunks
	state := &PersistentWorldState{
		Version:        CurrentSchemaVersion,
		WorldSeed:      12345,
		ChunkData:      make(map[string]*Chunk),
		Entities:       make([]*EntityState, 50),
		WorldEvents:    make([]WorldEvent, 10),
		ModifiedChunks: make(map[string]bool),
	}

	for i := 0; i < 100; i++ {
		chunkID := fmt.Sprintf("%d,%d", i%10, i/10)
		state.ChunkData[chunkID] = &Chunk{
			X:             i % 10,
			Y:             i / 10,
			Modifications: []TerrainMod{{Type: "dig", X: i, Y: i}},
		}
	}

	// Initial full save
	if err := wp.SaveWorld(state); err != nil {
		b.Fatalf("SaveWorld failed: %v", err)
	}

	// Mark 10% as modified
	for i := 0; i < 10; i++ {
		chunkID := fmt.Sprintf("%d,0", i)
		state.ModifiedChunks[chunkID] = true
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := wp.SaveIncremental(state); err != nil {
			b.Fatalf("SaveIncremental failed: %v", err)
		}
	}
}
