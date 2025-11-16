package world

import (
	"testing"
)

// TestChunkCompressionSystem_New tests creation
func TestChunkCompressionSystem_New(t *testing.T) {
	system := NewChunkCompressionSystem()
	if system == nil {
		t.Fatal("expected non-nil system")
	}
}

// TestChunkCompressionSystem_CompressUniform tests compressing uniform terrain
func TestChunkCompressionSystem_CompressUniform(t *testing.T) {
	system := NewChunkCompressionSystem()

	// Create chunk with uniform terrain
	chunk := createUniformChunk(ChunkSize, TileFloor)

	compressed, ratio, err := system.CompressChunk(chunk)
	if err != nil {
		t.Fatalf("CompressChunk failed: %v", err)
	}

	if len(compressed) == 0 {
		t.Error("expected non-empty compressed data")
	}

	// Uniform terrain should compress very well (ratio > 10)
	if ratio < 10.0 {
		t.Errorf("expected compression ratio > 10 for uniform terrain, got %.2f", ratio)
	}
}

// TestChunkCompressionSystem_CompressVaried tests compressing varied terrain
func TestChunkCompressionSystem_CompressVaried(t *testing.T) {
	system := NewChunkCompressionSystem()

	// Create chunk with varied terrain
	chunk := createVariedChunk(ChunkSize)

	compressed, ratio, err := system.CompressChunk(chunk)
	if err != nil {
		t.Fatalf("CompressChunk failed: %v", err)
	}

	if len(compressed) == 0 {
		t.Error("expected non-empty compressed data")
	}

	// Varied terrain may compress poorly (ratio can be < 1 due to overhead)
	if ratio < 0.1 {
		t.Errorf("expected compression ratio >= 0.1, got %.2f", ratio)
	}
}

// TestChunkCompressionSystem_RoundTrip tests compress/decompress cycle
func TestChunkCompressionSystem_RoundTrip(t *testing.T) {
	system := NewChunkCompressionSystem()

	// Create original chunk
	original := createPatternChunk(ChunkSize)

	// Compress
	compressed, _, err := system.CompressChunk(original)
	if err != nil {
		t.Fatalf("CompressChunk failed: %v", err)
	}

	// Decompress
	decompressed, err := system.DecompressChunk(compressed)
	if err != nil {
		t.Fatalf("DecompressChunk failed: %v", err)
	}

	// Compare terrain
	if !terrainEqual(original.Terrain, decompressed.Terrain) {
		t.Error("decompressed terrain does not match original")
	}
}

// TestChunkCompressionSystem_EmptyChunk tests compressing empty chunk
func TestChunkCompressionSystem_EmptyChunk(t *testing.T) {
	system := NewChunkCompressionSystem()

	chunk := &Chunk{
		Terrain:       nil,
		Modifications: []TerrainMod{},
	}

	compressed, ratio, err := system.CompressChunk(chunk)
	if err != nil {
		t.Fatalf("CompressChunk failed: %v", err)
	}

	if compressed != nil {
		t.Error("expected nil compressed data for empty chunk")
	}
	if ratio != 1.0 {
		t.Errorf("expected ratio 1.0 for empty chunk, got %.2f", ratio)
	}
}

// TestChunkCompressionSystem_DecompressEmpty tests decompressing empty data
func TestChunkCompressionSystem_DecompressEmpty(t *testing.T) {
	system := NewChunkCompressionSystem()

	decompressed, err := system.DecompressChunk(nil)
	if err != nil {
		t.Fatalf("DecompressChunk failed: %v", err)
	}

	if decompressed.Terrain != nil {
		t.Error("expected nil terrain for empty compressed data")
	}
}

// TestChunkCompressionSystem_EstimateRatio tests compression ratio estimation
func TestChunkCompressionSystem_EstimateRatio(t *testing.T) {
	system := NewChunkCompressionSystem()

	tests := []struct {
		name     string
		chunk    *Chunk
		minRatio float64
		maxRatio float64
	}{
		{
			name:     "uniform terrain",
			chunk:    createUniformChunk(ChunkSize, TileWall),
			minRatio: 10.0,
			maxRatio: 1000.0,
		},
		{
			name:     "varied terrain",
			chunk:    createVariedChunk(ChunkSize),
			minRatio: 0.1,
			maxRatio: 5.0,
		},
		{
			name:     "pattern terrain",
			chunk:    createPatternChunk(ChunkSize),
			minRatio: 0.1,
			maxRatio: 10.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ratio := system.EstimateCompressionRatio(tt.chunk)
			if ratio < tt.minRatio || ratio > tt.maxRatio {
				t.Errorf("expected ratio between %.2f and %.2f, got %.2f", tt.minRatio, tt.maxRatio, ratio)
			}
		})
	}
}

// TestChunkCompressionSystem_GetMemorySize tests memory size calculation
func TestChunkCompressionSystem_GetMemorySize(t *testing.T) {
	system := NewChunkCompressionSystem()

	chunk := createUniformChunk(ChunkSize, TileFloor)
	size := system.GetMemorySize(chunk)

	// ChunkSize x ChunkSize tiles, 4 bytes each
	expected := ChunkSize * ChunkSize * 4
	if size != expected {
		t.Errorf("expected memory size %d, got %d", expected, size)
	}
}

// TestChunkCompressionSystem_MemorySizeTarget tests <1MB target
func TestChunkCompressionSystem_MemorySizeTarget(t *testing.T) {
	system := NewChunkCompressionSystem()

	chunk := createUniformChunk(ChunkSize, TileFloor)
	size := system.GetMemorySize(chunk)

	// Target: <1MB per chunk
	maxSize := 1024 * 1024 // 1MB
	if size >= maxSize {
		t.Errorf("chunk memory size %d exceeds target %d", size, maxSize)
	}

	// For ChunkSize=32, size should be 32*32*4 = 4096 bytes
	if ChunkSize == 32 {
		expected := 4096
		if size != expected {
			t.Errorf("expected size %d for ChunkSize=32, got %d", expected, size)
		}
	}
}

// TestChunkCompressionSystem_MultipleRuns tests compressing runs
func TestChunkCompressionSystem_MultipleRuns(t *testing.T) {
	system := NewChunkCompressionSystem()

	// Create chunk with multiple runs per row
	terrain := make([][]TileType, 8)
	for i := range terrain {
		terrain[i] = []TileType{
			TileFloor, TileFloor, TileFloor,
			TileWall, TileWall,
			TileFloor, TileFloor, TileFloor,
		}
	}

	chunk := &Chunk{
		Terrain:       terrain,
		Modifications: []TerrainMod{},
	}

	// Compress and decompress
	compressed, _, err := system.CompressChunk(chunk)
	if err != nil {
		t.Fatalf("CompressChunk failed: %v", err)
	}

	decompressed, err := system.DecompressChunk(compressed)
	if err != nil {
		t.Fatalf("DecompressChunk failed: %v", err)
	}

	if !terrainEqual(chunk.Terrain, decompressed.Terrain) {
		t.Error("decompressed terrain does not match original")
	}
}

// Helper functions

func createUniformChunk(size int, tileType TileType) *Chunk {
	terrain := make([][]TileType, size)
	for i := range terrain {
		terrain[i] = make([]TileType, size)
		for j := range terrain[i] {
			terrain[i][j] = tileType
		}
	}
	return &Chunk{
		Terrain:       terrain,
		Modifications: []TerrainMod{},
	}
}

func createVariedChunk(size int) *Chunk {
	terrain := make([][]TileType, size)
	for i := range terrain {
		terrain[i] = make([]TileType, size)
		for j := range terrain[i] {
			// Pseudo-random pattern
			terrain[i][j] = TileType((i*size + j) % 8)
		}
	}
	return &Chunk{
		Terrain:       terrain,
		Modifications: []TerrainMod{},
	}
}

func createPatternChunk(size int) *Chunk {
	terrain := make([][]TileType, size)
	for i := range terrain {
		terrain[i] = make([]TileType, size)
		for j := range terrain[i] {
			// Checkerboard pattern
			if (i+j)%2 == 0 {
				terrain[i][j] = TileFloor
			} else {
				terrain[i][j] = TileWall
			}
		}
	}
	return &Chunk{
		Terrain:       terrain,
		Modifications: []TerrainMod{},
	}
}

func terrainEqual(a, b [][]TileType) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}
	return true
}

// Benchmarks

func BenchmarkCompressUniform(b *testing.B) {
	system := NewChunkCompressionSystem()
	chunk := createUniformChunk(ChunkSize, TileFloor)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.CompressChunk(chunk)
	}
}

func BenchmarkCompressVaried(b *testing.B) {
	system := NewChunkCompressionSystem()
	chunk := createVariedChunk(ChunkSize)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.CompressChunk(chunk)
	}
}

func BenchmarkDecompress(b *testing.B) {
	system := NewChunkCompressionSystem()
	chunk := createUniformChunk(ChunkSize, TileFloor)
	compressed, _, _ := system.CompressChunk(chunk)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.DecompressChunk(compressed)
	}
}

func BenchmarkEstimateRatio(b *testing.B) {
	system := NewChunkCompressionSystem()
	chunk := createPatternChunk(ChunkSize)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.EstimateCompressionRatio(chunk)
	}
}
