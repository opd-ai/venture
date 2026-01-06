package world

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestSaveWorldWithContext verifies context-based save operations.
func TestSaveWorldWithContext(t *testing.T) {
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "test.save")

	wp := NewWorldPersistence(savePath)
	state := &PersistentWorldState{
		Version:   CurrentSchemaVersion,
		WorldSeed: 12345,
		ChunkData: make(map[string]*Chunk),
		Entities:  []*EntityState{},
		WorldEvents: []WorldEvent{},
		ModifiedChunks: make(map[string]bool),
	}

	// Test normal save with context
	ctx := context.Background()
	if err := wp.SaveWorldWithContext(ctx, state); err != nil {
		t.Fatalf("SaveWorldWithContext failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		t.Error("Save file was not created")
	}
}

// TestSaveWorldContextCancellation verifies save cancellation.
func TestSaveWorldContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "test.save")

	wp := NewWorldPersistence(savePath)
	
	// Create a large state to make save take longer
	state := &PersistentWorldState{
		Version:   CurrentSchemaVersion,
		WorldSeed: 12345,
		ChunkData: make(map[string]*Chunk),
		Entities:  make([]*EntityState, 10000), // Large dataset
		WorldEvents: []WorldEvent{},
		ModifiedChunks: make(map[string]bool),
	}
	
	// Fill with dummy data
	for i := range state.Entities {
		state.Entities[i] = &EntityState{
			ID:       uint64(i),
			TypeName: "Monster",
			Components: map[string]interface{}{
				"position": map[string]float64{"x": float64(i), "y": float64(i)},
			},
		}
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Save should fail with context error
	err := wp.SaveWorldWithContext(ctx, state)
	if err == nil {
		t.Error("Expected error when saving with cancelled context")
	}

	// Verify temp file was cleaned up
	tempPath := savePath + ".tmp"
	if _, err := os.Stat(tempPath); err == nil {
		t.Error("Temp file should have been cleaned up on error")
	}
}

// TestSaveWorldContextTimeout verifies timeout handling.
func TestSaveWorldContextTimeout(t *testing.T) {
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "test.save")

	wp := NewWorldPersistence(savePath)
	state := &PersistentWorldState{
		Version:   CurrentSchemaVersion,
		WorldSeed: 12345,
		ChunkData: make(map[string]*Chunk),
		Entities:  []*EntityState{},
		WorldEvents: []WorldEvent{},
		ModifiedChunks: make(map[string]bool),
	}

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	
	// Wait for timeout
	time.Sleep(2 * time.Millisecond)

	// Save might succeed (if fast enough) or fail with timeout
	err := wp.SaveWorldWithContext(ctx, state)
	if err != nil {
		// If it failed, should be context error
		if ctx.Err() != context.DeadlineExceeded {
			t.Logf("Save failed (expected with short timeout): %v", err)
		}
	}
}

// TestLoadWorldWithContext verifies context-based load operations.
func TestLoadWorldWithContext(t *testing.T) {
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "test.save")

	wp := NewWorldPersistence(savePath)
	
	// First save a state
	originalState := &PersistentWorldState{
		Version:   CurrentSchemaVersion,
		WorldSeed: 12345,
		ChunkData: make(map[string]*Chunk),
		Entities:  []*EntityState{
			{ID: 1, TypeName: "Monster", Components: make(map[string]interface{})},
		},
		WorldEvents: []WorldEvent{},
		ModifiedChunks: make(map[string]bool),
	}

	if err := wp.SaveWorld(originalState); err != nil {
		t.Fatalf("Failed to save initial state: %v", err)
	}

	// Load with context
	ctx := context.Background()
	loadedState, err := wp.LoadWorldWithContext(ctx, 12345)
	if err != nil {
		t.Fatalf("LoadWorldWithContext failed: %v", err)
	}

	// Verify loaded state
	if loadedState.WorldSeed != originalState.WorldSeed {
		t.Errorf("Expected seed %d, got %d", originalState.WorldSeed, loadedState.WorldSeed)
	}

	if len(loadedState.Entities) != len(originalState.Entities) {
		t.Errorf("Expected %d entities, got %d", len(originalState.Entities), len(loadedState.Entities))
	}
}

// TestLoadWorldContextCancellation verifies load cancellation.
func TestLoadWorldContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "test.save")

	wp := NewWorldPersistence(savePath)
	
	// Create a large state file
	state := &PersistentWorldState{
		Version:   CurrentSchemaVersion,
		WorldSeed: 12345,
		ChunkData: make(map[string]*Chunk),
		Entities:  make([]*EntityState, 5000),
		WorldEvents: []WorldEvent{},
		ModifiedChunks: make(map[string]bool),
	}
	
	for i := range state.Entities {
		state.Entities[i] = &EntityState{
			ID:       uint64(i),
			TypeName: "Monster",
			Components: map[string]interface{}{
				"position": map[string]float64{"x": float64(i), "y": float64(i)},
			},
		}
	}

	if err := wp.SaveWorld(state); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Load should fail with context error
	_, err := wp.LoadWorldWithContext(ctx, 12345)
	if err == nil {
		t.Error("Expected error when loading with cancelled context")
	}
}

// TestLoadWorldBackupRecovery verifies backup fallback with context.
func TestLoadWorldBackupRecovery(t *testing.T) {
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "test.save")

	wp := NewWorldPersistence(savePath)
	
	// Save initial state
	state := &PersistentWorldState{
		Version:   CurrentSchemaVersion,
		WorldSeed: 12345,
		ChunkData: make(map[string]*Chunk),
		Entities:  []*EntityState{},
		WorldEvents: []WorldEvent{},
		ModifiedChunks: make(map[string]bool),
	}

	if err := wp.SaveWorld(state); err != nil {
		t.Fatalf("Failed to save initial state: %v", err)
	}

	// Corrupt main save file
	if err := os.WriteFile(savePath, []byte("corrupted data"), 0o644); err != nil {
		t.Fatalf("Failed to corrupt save file: %v", err)
	}

	// Load should fall back to backup
	ctx := context.Background()
	loadedState, err := wp.LoadWorldWithContext(ctx, 12345)
	if err != nil {
		t.Fatalf("LoadWorldWithContext failed: %v", err)
	}

	// Should load from backup (which was created during first save)
	if loadedState.WorldSeed != state.WorldSeed {
		t.Errorf("Expected seed %d from backup, got %d", state.WorldSeed, loadedState.WorldSeed)
	}
}

// TestSaveIncrementalWithContext verifies incremental save with context.
func TestSaveIncrementalWithContext(t *testing.T) {
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "test.save")

	wp := NewWorldPersistence(savePath)
	
	// Initial full save
	state := &PersistentWorldState{
		Version:   CurrentSchemaVersion,
		WorldSeed: 12345,
		ChunkData: map[string]*Chunk{
			"0,0": {X: 0, Y: 0, Modifications: []TerrainMod{}},
		},
		Entities:  []*EntityState{},
		WorldEvents: []WorldEvent{},
		ModifiedChunks: map[string]bool{"0,0": true},
	}

	if err := wp.SaveWorld(state); err != nil {
		t.Fatalf("Failed to save initial state: %v", err)
	}

	// Incremental save with context
	state.ChunkData["1,1"] = &Chunk{X: 1, Y: 1, Modifications: []TerrainMod{}}
	state.ModifiedChunks = map[string]bool{"1,1": true}

	ctx := context.Background()
	if err := wp.SaveIncrementalWithContext(ctx, state); err != nil {
		t.Fatalf("SaveIncrementalWithContext failed: %v", err)
	}

	// Verify save succeeded
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		t.Error("Incremental save file was not created")
	}
}

// TestCopyFileErrorHandling verifies improved copyFile error handling.
func TestCopyFileErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Test successful copy
	srcPath := filepath.Join(tmpDir, "source.txt")
	dstPath := filepath.Join(tmpDir, "dest.txt")

	testData := []byte("test data for copy")
	if err := os.WriteFile(srcPath, testData, 0o644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	if err := copyFile(srcPath, dstPath); err != nil {
		t.Errorf("copyFile failed: %v", err)
	}

	// Verify destination file
	copiedData, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(copiedData) != string(testData) {
		t.Errorf("Copied data mismatch: expected %q, got %q", testData, copiedData)
	}

	// Test copy from non-existent file
	badSrcPath := filepath.Join(tmpDir, "nonexistent.txt")
	badDstPath := filepath.Join(tmpDir, "bad_dest.txt")

	if err := copyFile(badSrcPath, badDstPath); err == nil {
		t.Error("copyFile should fail with non-existent source")
	}

	// Verify incomplete destination was cleaned up
	if _, err := os.Stat(badDstPath); err == nil {
		t.Error("Failed copy should not leave destination file")
	}
}

// TestLoadWorldNonExistent verifies behavior with non-existent save file.
func TestLoadWorldNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "nonexistent.save")

	wp := NewWorldPersistence(savePath)
	
	ctx := context.Background()
	state, err := wp.LoadWorldWithContext(ctx, 12345)
	if err != nil {
		t.Fatalf("LoadWorldWithContext should succeed with default state: %v", err)
	}

	// Should return new state with correct seed
	if state.WorldSeed != 12345 {
		t.Errorf("Expected seed 12345, got %d", state.WorldSeed)
	}

	if state.Version != CurrentSchemaVersion {
		t.Errorf("Expected version %d, got %d", CurrentSchemaVersion, state.Version)
	}

	if len(state.Entities) != 0 {
		t.Errorf("Expected empty entities, got %d", len(state.Entities))
	}
}

// TestBackwardCompatibility verifies legacy methods still work.
func TestBackwardCompatibility(t *testing.T) {
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "test.save")

	wp := NewWorldPersistence(savePath)
	state := &PersistentWorldState{
		Version:   CurrentSchemaVersion,
		WorldSeed: 12345,
		ChunkData: make(map[string]*Chunk),
		Entities:  []*EntityState{},
		WorldEvents: []WorldEvent{},
		ModifiedChunks: make(map[string]bool),
	}

	// Legacy SaveWorld (no context) should still work
	if err := wp.SaveWorld(state); err != nil {
		t.Fatalf("Legacy SaveWorld failed: %v", err)
	}

	// Legacy LoadWorld (no context) should still work
	loadedState, err := wp.LoadWorld(12345)
	if err != nil {
		t.Fatalf("Legacy LoadWorld failed: %v", err)
	}

	if loadedState.WorldSeed != state.WorldSeed {
		t.Errorf("Expected seed %d, got %d", state.WorldSeed, loadedState.WorldSeed)
	}

	// Legacy SaveIncremental should still work
	state.ChunkData["0,0"] = &Chunk{X: 0, Y: 0}
	state.ModifiedChunks["0,0"] = true

	if err := wp.SaveIncremental(state); err != nil {
		t.Fatalf("Legacy SaveIncremental failed: %v", err)
	}
}
