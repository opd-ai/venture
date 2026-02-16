package world

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// ChunkCompressionSystem handles RLE encoding for uniform terrain
type ChunkCompressionSystem struct{}

// NewChunkCompressionSystem creates a new chunk compression system
func NewChunkCompressionSystem() *ChunkCompressionSystem {
	return &ChunkCompressionSystem{}
}

// CompressChunk compresses chunk terrain using RLE encoding
// Returns compressed data and compression ratio (original size / compressed size)
func (c *ChunkCompressionSystem) CompressChunk(chunk *Chunk) ([]byte, float64, error) {
	if chunk.Terrain == nil || len(chunk.Terrain) == 0 {
		return nil, 1.0, nil // No terrain to compress
	}

	var buf bytes.Buffer

	// Write chunk dimensions
	if err := binary.Write(&buf, binary.LittleEndian, int32(len(chunk.Terrain))); err != nil {
		return nil, 0, fmt.Errorf("failed to write height: %w", err)
	}
	if len(chunk.Terrain[0]) == 0 {
		return nil, 1.0, nil // No columns to compress
	}
	if err := binary.Write(&buf, binary.LittleEndian, int32(len(chunk.Terrain[0]))); err != nil {
		return nil, 0, fmt.Errorf("failed to write width: %w", err)
	}

	// RLE compression
	for _, row := range chunk.Terrain {
		if err := c.compressRow(row, &buf); err != nil {
			return nil, 0, fmt.Errorf("failed to compress row: %w", err)
		}
	}

	compressed := buf.Bytes()
	originalSize := len(chunk.Terrain) * len(chunk.Terrain[0]) * 4 // int32 per tile
	compressionRatio := float64(originalSize) / float64(len(compressed))

	return compressed, compressionRatio, nil
}

// compressRow compresses a single row using RLE
func (c *ChunkCompressionSystem) compressRow(row []TileType, buf *bytes.Buffer) error {
	if len(row) == 0 {
		return nil
	}

	currentTile := row[0]
	count := int32(1)

	for i := 1; i < len(row); i++ {
		if row[i] == currentTile && count < 255 {
			count++
		} else {
			// Write run
			if err := binary.Write(buf, binary.LittleEndian, int32(currentTile)); err != nil {
				return err
			}
			if err := binary.Write(buf, binary.LittleEndian, count); err != nil {
				return err
			}

			currentTile = row[i]
			count = 1
		}
	}

	// Write final run
	if err := binary.Write(buf, binary.LittleEndian, int32(currentTile)); err != nil {
		return err
	}
	return binary.Write(buf, binary.LittleEndian, count)
}

// DecompressChunk decompresses RLE-encoded chunk terrain
func (c *ChunkCompressionSystem) DecompressChunk(data []byte) (*Chunk, error) {
	if len(data) == 0 {
		return &Chunk{
			Terrain:       nil,
			Modifications: []TerrainMod{},
		}, nil
	}

	buf := bytes.NewReader(data)

	// Read dimensions
	var height, width int32
	if err := binary.Read(buf, binary.LittleEndian, &height); err != nil {
		return nil, fmt.Errorf("failed to read height: %w", err)
	}
	if err := binary.Read(buf, binary.LittleEndian, &width); err != nil {
		return nil, fmt.Errorf("failed to read width: %w", err)
	}

	// Allocate terrain
	terrain := make([][]TileType, height)
	for i := range terrain {
		terrain[i] = make([]TileType, width)
	}

	// Decompress rows
	for y := int32(0); y < height; y++ {
		if err := c.decompressRow(buf, terrain[y]); err != nil {
			return nil, fmt.Errorf("failed to decompress row %d: %w", y, err)
		}
	}

	return &Chunk{
		Terrain:       terrain,
		Modifications: []TerrainMod{},
	}, nil
}

// decompressRow decompresses a single row from RLE
func (c *ChunkCompressionSystem) decompressRow(buf *bytes.Reader, row []TileType) error {
	pos := 0

	for pos < len(row) {
		var tileType int32
		var count int32

		if err := binary.Read(buf, binary.LittleEndian, &tileType); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if err := binary.Read(buf, binary.LittleEndian, &count); err != nil {
			return err
		}

		// Write run
		for i := int32(0); i < count && pos < len(row); i++ {
			row[pos] = TileType(tileType)
			pos++
		}
	}

	return nil
}

// EstimateCompressionRatio estimates compression ratio for a chunk
// Returns ratio (original size / estimated compressed size)
func (c *ChunkCompressionSystem) EstimateCompressionRatio(chunk *Chunk) float64 {
	if chunk.Terrain == nil || len(chunk.Terrain) == 0 {
		return 1.0
	}

	totalTiles := 0
	uniqueRuns := 0
	lastTile := TileType(-1)

	for _, row := range chunk.Terrain {
		for _, tile := range row {
			totalTiles++
			if tile != lastTile {
				uniqueRuns++
				lastTile = tile
			}
		}
	}

	if uniqueRuns == 0 {
		return 1.0
	}

	// Each run takes 8 bytes (4 for tile type, 4 for count)
	// Original takes 4 bytes per tile
	originalSize := totalTiles * 4
	compressedSize := uniqueRuns*8 + 8 // +8 for dimensions

	if compressedSize == 0 {
		return 1.0
	}

	return float64(originalSize) / float64(compressedSize)
}

// GetMemorySize calculates uncompressed memory size of chunk in bytes
func (c *ChunkCompressionSystem) GetMemorySize(chunk *Chunk) int {
	if chunk.Terrain == nil {
		return 0
	}

	size := 0
	for _, row := range chunk.Terrain {
		size += len(row) * 4 // 4 bytes per TileType (int32)
	}
	return size
}
