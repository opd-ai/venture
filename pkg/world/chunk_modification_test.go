package world

import (
	"testing"
)

// TestChunkModificationSystem_New tests creation
func TestChunkModificationSystem_New(t *testing.T) {
	state := &PersistentWorldState{
		ChunkData:      make(map[string]*Chunk),
		ModifiedChunks: make(map[string]bool),
	}

	system := NewChunkModificationSystem(state)
	if system == nil {
		t.Fatal("expected non-nil system")
	}
	if system.state != state {
		t.Error("expected system to store state reference")
	}
}

// TestChunkModificationSystem_ModifyTerrain tests terrain modification
func TestChunkModificationSystem_ModifyTerrain(t *testing.T) {
	state := &PersistentWorldState{
		ChunkData:      make(map[string]*Chunk),
		ModifiedChunks: make(map[string]bool),
	}
	system := NewChunkModificationSystem(state)

	// Modify terrain at (5, 5)
	err := system.ModifyTerrain(5, 5, TileWall)
	if err != nil {
		t.Fatalf("ModifyTerrain failed: %v", err)
	}

	// Check chunk was created and marked dirty
	chunkID := chunkCoordsToID(0, 0) // (5,5) is in chunk (0,0)
	if !system.dirtyChunks[chunkID] {
		t.Error("expected chunk to be marked dirty")
	}

	// Check terrain was modified
	chunk := state.ChunkData[chunkID]
	if chunk == nil {
		t.Fatal("expected chunk to exist")
	}
	if chunk.Terrain[5][5] != TileWall {
		t.Errorf("expected TileWall at (5,5), got %v", chunk.Terrain[5][5])
	}
}

// TestChunkModificationSystem_ModifyMultipleTiles tests multiple modifications
func TestChunkModificationSystem_ModifyMultipleTiles(t *testing.T) {
	state := &PersistentWorldState{
		ChunkData:      make(map[string]*Chunk),
		ModifiedChunks: make(map[string]bool),
	}
	system := NewChunkModificationSystem(state)

	// Modify multiple tiles
	positions := []struct{ x, y int }{
		{0, 0},
		{10, 10},
		{31, 31},
	}

	for _, pos := range positions {
		err := system.ModifyTerrain(pos.x, pos.y, TileWater)
		if err != nil {
			t.Fatalf("ModifyTerrain(%d, %d) failed: %v", pos.x, pos.y, err)
		}
	}

	// All tiles are in chunk (0,0)
	chunkID := chunkCoordsToID(0, 0)
	chunk := state.ChunkData[chunkID]
	if chunk == nil {
		t.Fatal("expected chunk to exist")
	}

	for _, pos := range positions {
		if chunk.Terrain[pos.y][pos.x] != TileWater {
			t.Errorf("expected TileWater at (%d,%d), got %v", pos.x, pos.y, chunk.Terrain[pos.y][pos.x])
		}
	}
}

// TestChunkModificationSystem_AddModification tests adding modifications
func TestChunkModificationSystem_AddModification(t *testing.T) {
	state := &PersistentWorldState{
		ChunkData:      make(map[string]*Chunk),
		ModifiedChunks: make(map[string]bool),
	}
	system := NewChunkModificationSystem(state)

	// Add explosion modification
	err := system.AddModification("explosion", 10, 10, 5.0)
	if err != nil {
		t.Fatalf("AddModification failed: %v", err)
	}

	chunkID := chunkCoordsToID(0, 0)
	chunk := state.ChunkData[chunkID]
	if chunk == nil {
		t.Fatal("expected chunk to exist")
	}

	if len(chunk.Modifications) != 1 {
		t.Fatalf("expected 1 modification, got %d", len(chunk.Modifications))
	}

	mod := chunk.Modifications[0]
	if mod.Type != "explosion" {
		t.Errorf("expected type 'explosion', got '%s'", mod.Type)
	}
	if mod.X != 10 || mod.Y != 10 {
		t.Errorf("expected position (10,10), got (%d,%d)", mod.X, mod.Y)
	}
	if mod.Radius != 5.0 {
		t.Errorf("expected radius 5.0, got %f", mod.Radius)
	}
	if mod.Timestamp == 0 {
		t.Error("expected non-zero timestamp")
	}
}

// TestChunkModificationSystem_MultipleModifications tests multiple modifications
func TestChunkModificationSystem_MultipleModifications(t *testing.T) {
	state := &PersistentWorldState{
		ChunkData:      make(map[string]*Chunk),
		ModifiedChunks: make(map[string]bool),
	}
	system := NewChunkModificationSystem(state)

	// Add multiple modifications
	mods := []struct {
		modType string
		x, y    int
		radius  float64
	}{
		{"explosion", 5, 5, 3.0},
		{"dig", 10, 10, 2.0},
		{"build", 15, 15, 1.0},
	}

	for _, m := range mods {
		err := system.AddModification(m.modType, m.x, m.y, m.radius)
		if err != nil {
			t.Fatalf("AddModification failed: %v", err)
		}
	}

	chunkID := chunkCoordsToID(0, 0)
	chunk := state.ChunkData[chunkID]
	if len(chunk.Modifications) != len(mods) {
		t.Fatalf("expected %d modifications, got %d", len(mods), len(chunk.Modifications))
	}

	for i, expected := range mods {
		actual := chunk.Modifications[i]
		if actual.Type != expected.modType {
			t.Errorf("mod %d: expected type '%s', got '%s'", i, expected.modType, actual.Type)
		}
	}
}

// TestChunkModificationSystem_GetModifiedChunks tests retrieving modified chunks
func TestChunkModificationSystem_GetModifiedChunks(t *testing.T) {
	state := &PersistentWorldState{
		ChunkData:      make(map[string]*Chunk),
		ModifiedChunks: make(map[string]bool),
	}
	system := NewChunkModificationSystem(state)

	// Modify chunks at different positions
	system.ModifyTerrain(0, 0, TileWall)          // Chunk (0,0)
	system.ModifyTerrain(ChunkSize, 0, TileWater) // Chunk (1,0)
	system.ModifyTerrain(0, ChunkSize, TileLava)  // Chunk (0,1)

	modified := system.GetModifiedChunks()
	if len(modified) != 3 {
		t.Fatalf("expected 3 modified chunks, got %d", len(modified))
	}
}

// TestChunkModificationSystem_ClearDirtyFlags tests clearing dirty flags
func TestChunkModificationSystem_ClearDirtyFlags(t *testing.T) {
	state := &PersistentWorldState{
		ChunkData:      make(map[string]*Chunk),
		ModifiedChunks: make(map[string]bool),
	}
	system := NewChunkModificationSystem(state)

	// Modify terrain
	system.ModifyTerrain(5, 5, TileWall)

	// Verify dirty
	modified := system.GetModifiedChunks()
	if len(modified) == 0 {
		t.Fatal("expected modified chunks before clear")
	}

	// Clear flags
	system.ClearDirtyFlags()

	// Verify clean
	modified = system.GetModifiedChunks()
	if len(modified) != 0 {
		t.Errorf("expected 0 modified chunks after clear, got %d", len(modified))
	}
}

// TestChunkModificationSystem_GetModificationCount tests counting modifications
func TestChunkModificationSystem_GetModificationCount(t *testing.T) {
	state := &PersistentWorldState{
		ChunkData:      make(map[string]*Chunk),
		ModifiedChunks: make(map[string]bool),
	}
	system := NewChunkModificationSystem(state)

	// No modifications initially
	count := system.GetModificationCount(0, 0)
	if count != 0 {
		t.Errorf("expected 0 modifications, got %d", count)
	}

	// Add modifications
	system.AddModification("explosion", 5, 5, 3.0)
	system.AddModification("dig", 10, 10, 2.0)

	count = system.GetModificationCount(0, 0)
	if count != 2 {
		t.Errorf("expected 2 modifications, got %d", count)
	}
}

// TestChunkModificationSystem_HasModifications tests checking if chunk is modified
func TestChunkModificationSystem_HasModifications(t *testing.T) {
	state := &PersistentWorldState{
		ChunkData:      make(map[string]*Chunk),
		ModifiedChunks: make(map[string]bool),
	}
	system := NewChunkModificationSystem(state)

	// Initially no modifications
	if system.HasModifications(0, 0) {
		t.Error("expected chunk to not have modifications")
	}

	// Modify terrain
	system.ModifyTerrain(5, 5, TileWall)

	// Now should have modifications
	if !system.HasModifications(0, 0) {
		t.Error("expected chunk to have modifications")
	}
}

// TestChunkModificationSystem_NegativeCoordinates tests negative coordinates
func TestChunkModificationSystem_NegativeCoordinates(t *testing.T) {
	state := &PersistentWorldState{
		ChunkData:      make(map[string]*Chunk),
		ModifiedChunks: make(map[string]bool),
	}
	system := NewChunkModificationSystem(state)

	// Modify at negative coordinates
	// (-5, -5) should be in chunk (-1, -1) due to floored division
	err := system.ModifyTerrain(-5, -5, TileWall)
	if err != nil {
		t.Fatalf("ModifyTerrain failed for negative coords: %v", err)
	}

	// Check all possible chunks that might be dirty
	foundDirty := false
	for chunkID, isDirty := range system.dirtyChunks {
		if isDirty {
			t.Logf("Dirty chunk: %s", chunkID)
			foundDirty = true
		}
	}

	if !foundDirty {
		t.Error("expected at least one chunk to be marked dirty")
	}
}

// BenchmarkModifyTerrain benchmarks terrain modification
func BenchmarkModifyTerrain(b *testing.B) {
	state := &PersistentWorldState{
		ChunkData:      make(map[string]*Chunk),
		ModifiedChunks: make(map[string]bool),
	}
	system := NewChunkModificationSystem(state)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := i % 100
		y := (i / 100) % 100
		system.ModifyTerrain(x, y, TileFloor)
	}
}

// BenchmarkAddModification benchmarks adding modifications
func BenchmarkAddModification(b *testing.B) {
	state := &PersistentWorldState{
		ChunkData:      make(map[string]*Chunk),
		ModifiedChunks: make(map[string]bool),
	}
	system := NewChunkModificationSystem(state)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := i % 100
		y := (i / 100) % 100
		system.AddModification("explosion", x, y, 3.0)
	}
}
