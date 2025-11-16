package world

import (
	"testing"
)

// TestChunkLoaderSystem_New tests chunk loader creation
func TestChunkLoaderSystem_New(t *testing.T) {
	loader := NewChunkLoaderSystem(12345, nil, nil)
	if loader == nil {
		t.Fatal("expected non-nil loader")
	}
	if loader.loadRadius != 5 {
		t.Errorf("expected default load radius 5, got %d", loader.loadRadius)
	}
	if loader.worldSeed != 12345 {
		t.Errorf("expected world seed 12345, got %d", loader.worldSeed)
	}
}

// TestChunkLoaderSystem_SetLoadRadius tests changing load radius
func TestChunkLoaderSystem_SetLoadRadius(t *testing.T) {
	loader := NewChunkLoaderSystem(12345, nil, nil)
	loader.SetLoadRadius(10)
	if loader.loadRadius != 10 {
		t.Errorf("expected load radius 10, got %d", loader.loadRadius)
	}
}

// MockChunkGenerator generates simple chunks for testing
type MockChunkGenerator struct {
	generated map[string]bool
}

func NewMockChunkGenerator() *MockChunkGenerator {
	return &MockChunkGenerator{
		generated: make(map[string]bool),
	}
}

func (m *MockChunkGenerator) GenerateChunk(chunkX, chunkY int, seed int64) (*Chunk, error) {
	chunkID := chunkCoordsToID(chunkX, chunkY)
	m.generated[chunkID] = true

	// Generate simple uniform terrain
	terrain := make([][]TileType, ChunkSize)
	for i := range terrain {
		terrain[i] = make([]TileType, ChunkSize)
		for j := range terrain[i] {
			terrain[i][j] = TileFloor
		}
	}

	return &Chunk{
		X:             chunkX,
		Y:             chunkY,
		Terrain:       terrain,
		Modifications: []TerrainMod{},
	}, nil
}

// TestChunkLoaderSystem_Update tests chunk loading based on player position
func TestChunkLoaderSystem_Update(t *testing.T) {
	generator := NewMockChunkGenerator()
	loader := NewChunkLoaderSystem(12345, nil, generator)
	loader.SetLoadRadius(2) // Small radius for testing

	// Single player at (0, 0)
	playerPositions := map[uint64]struct{ X, Y float64 }{
		1: {X: 0, Y: 0},
	}

	err := loader.Update(playerPositions)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Should load 5x5 = 25 chunks (radius 2 around chunk 0,0)
	expected := (2*2 + 1) * (2*2 + 1)
	if loader.GetLoadedChunkCount() != expected {
		t.Errorf("expected %d loaded chunks, got %d", expected, loader.GetLoadedChunkCount())
	}

	// Verify center chunk is loaded
	_, exists := loader.GetChunk(0, 0)
	if !exists {
		t.Error("expected chunk (0,0) to be loaded")
	}
}

// TestChunkLoaderSystem_MultiplePlayers tests loading with multiple players
func TestChunkLoaderSystem_MultiplePlayersLoading(t *testing.T) {
	generator := NewMockChunkGenerator()
	loader := NewChunkLoaderSystem(12345, nil, generator)
	loader.SetLoadRadius(1) // Radius 1 = 3x3 chunks per player

	// Two players far apart
	playerPositions := map[uint64]struct{ X, Y float64 }{
		1: {X: 0, Y: 0},                           // Chunk (0, 0)
		2: {X: ChunkSize * 10, Y: ChunkSize * 10}, // Chunk (10, 10)
	}

	err := loader.Update(playerPositions)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Should load chunks around both players
	// Player 1: (-1,-1) to (1,1) = 9 chunks
	// Player 2: (9,9) to (11,11) = 9 chunks
	// Total: 18 chunks (no overlap)
	expected := 18
	if loader.GetLoadedChunkCount() != expected {
		t.Errorf("expected %d loaded chunks, got %d", expected, loader.GetLoadedChunkCount())
	}
}

// TestChunkLoaderSystem_Unloading tests chunk unloading when player moves
func TestChunkLoaderSystem_Unloading(t *testing.T) {
	generator := NewMockChunkGenerator()
	loader := NewChunkLoaderSystem(12345, nil, generator)
	loader.SetLoadRadius(1)

	// Player at (0, 0)
	playerPositions := map[uint64]struct{ X, Y float64 }{
		1: {X: 0, Y: 0},
	}
	loader.Update(playerPositions)
	initialCount := loader.GetLoadedChunkCount()

	// Move player far away
	playerPositions[1] = struct{ X, Y float64 }{X: ChunkSize * 100, Y: ChunkSize * 100}
	loader.Update(playerPositions)

	// Should have unloaded old chunks and loaded new ones
	// Count should be same (radius hasn't changed)
	if loader.GetLoadedChunkCount() != initialCount {
		t.Errorf("expected %d loaded chunks after move, got %d", initialCount, loader.GetLoadedChunkCount())
	}

	// Old chunk should be unloaded
	_, exists := loader.GetChunk(0, 0)
	if exists {
		t.Error("expected chunk (0,0) to be unloaded after player moved")
	}

	// New chunk should be loaded
	_, exists = loader.GetChunk(100, 100)
	if !exists {
		t.Error("expected chunk (100,100) to be loaded at new player position")
	}
}

// TestChunkLoaderSystem_GetChunk tests chunk retrieval
func TestChunkLoaderSystem_GetChunk(t *testing.T) {
	generator := NewMockChunkGenerator()
	loader := NewChunkLoaderSystem(12345, nil, generator)

	playerPositions := map[uint64]struct{ X, Y float64 }{
		1: {X: 0, Y: 0},
	}
	loader.Update(playerPositions)

	// Get existing chunk
	chunk, exists := loader.GetChunk(0, 0)
	if !exists {
		t.Error("expected chunk (0,0) to exist")
	}
	if chunk.X != 0 || chunk.Y != 0 {
		t.Errorf("expected chunk coords (0,0), got (%d,%d)", chunk.X, chunk.Y)
	}

	// Get non-existent chunk
	_, exists = loader.GetChunk(999, 999)
	if exists {
		t.Error("expected chunk (999,999) to not exist")
	}
}

// TestChunkLoaderSystem_PerformanceTarget tests chunk load time
func TestChunkLoaderSystem_LoadTime(t *testing.T) {
	generator := NewMockChunkGenerator()
	loader := NewChunkLoaderSystem(12345, nil, generator)
	loader.SetLoadRadius(5)

	// Load chunks around player
	playerPositions := map[uint64]struct{ X, Y float64 }{
		1: {X: 0, Y: 0},
	}

	// This should load (5*2+1)^2 = 121 chunks
	err := loader.Update(playerPositions)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	expected := 121
	if loader.GetLoadedChunkCount() != expected {
		t.Errorf("expected %d loaded chunks, got %d", expected, loader.GetLoadedChunkCount())
	}
}

// TestChunkCoordsToID tests chunk ID generation
func TestChunkCoordsToID(t *testing.T) {
	tests := []struct {
		x        int
		y        int
		expected string
	}{
		{0, 0, "0,0"},
		{5, 10, "5,10"},
		{-3, 7, "-3,7"},
		{100, -50, "100,-50"},
	}

	for _, tt := range tests {
		result := chunkCoordsToID(tt.x, tt.y)
		if result != tt.expected {
			t.Errorf("chunkCoordsToID(%d, %d) = %s, expected %s", tt.x, tt.y, result, tt.expected)
		}
	}
}

// BenchmarkChunkLoaderUpdate benchmarks chunk loading performance
func BenchmarkChunkLoaderUpdate(b *testing.B) {
	generator := NewMockChunkGenerator()
	loader := NewChunkLoaderSystem(12345, nil, generator)
	loader.SetLoadRadius(5)

	playerPositions := map[uint64]struct{ X, Y float64 }{
		1: {X: 0, Y: 0},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loader.Update(playerPositions)
	}
}

// BenchmarkChunkLoaderMultiPlayer benchmarks multi-player chunk loading
func BenchmarkChunkLoaderMultiPlayer(b *testing.B) {
	generator := NewMockChunkGenerator()
	loader := NewChunkLoaderSystem(12345, nil, generator)
	loader.SetLoadRadius(3)

	// 10 players at different positions
	playerPositions := make(map[uint64]struct{ X, Y float64 })
	for i := uint64(0); i < 10; i++ {
		playerPositions[i] = struct{ X, Y float64 }{
			X: float64(i * ChunkSize * 5),
			Y: float64(i * ChunkSize * 5),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loader.Update(playerPositions)
	}
}
