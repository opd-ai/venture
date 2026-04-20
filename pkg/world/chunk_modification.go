package world

import (
	"fmt"
	"time"
)

// ChunkModificationSystem tracks terrain changes and marks chunks dirty
type ChunkModificationSystem struct {
	dirtyChunks map[string]bool // Track modified chunks
	state       *PersistentWorldState
}

// NewChunkModificationSystem creates a new chunk modification tracker
func NewChunkModificationSystem(state *PersistentWorldState) *ChunkModificationSystem {
	if state.ModifiedChunks == nil {
		state.ModifiedChunks = make(map[string]bool)
	}
	return &ChunkModificationSystem{
		dirtyChunks: state.ModifiedChunks,
		state:       state,
	}
}

// ModifyTerrain modifies terrain at a specific position
func (c *ChunkModificationSystem) ModifyTerrain(x, y int, tileType TileType) error {
	chunkX := x / ChunkSize
	chunkY := y / ChunkSize
	chunkID := chunkCoordsToID(chunkX, chunkY)

	// Get or create chunk
	chunk, exists := c.state.ChunkData[chunkID]
	if !exists {
		chunk = &Chunk{
			X:             chunkX,
			Y:             chunkY,
			Terrain:       make([][]TileType, ChunkSize),
			Modifications: []TerrainMod{},
		}
		for i := range chunk.Terrain {
			chunk.Terrain[i] = make([]TileType, ChunkSize)
		}
		c.state.ChunkData[chunkID] = chunk
	}

	// Ensure terrain array is initialized
	if chunk.Terrain == nil {
		chunk.Terrain = make([][]TileType, ChunkSize)
		for i := range chunk.Terrain {
			chunk.Terrain[i] = make([]TileType, ChunkSize)
		}
	}

	// Calculate local coordinates within chunk
	localX := x % ChunkSize
	localY := y % ChunkSize
	if localX < 0 {
		localX += ChunkSize
	}
	if localY < 0 {
		localY += ChunkSize
	}

	// Validate bounds
	if localX < 0 || localX >= ChunkSize || localY < 0 || localY >= ChunkSize {
		return fmt.Errorf("invalid local coordinates: (%d, %d)", localX, localY)
	}

	// Modify terrain
	chunk.Terrain[localY][localX] = tileType

	// Mark chunk as dirty
	c.dirtyChunks[chunkID] = true

	return nil
}

// AddModification adds a terrain modification (explosion, dig, build)
func (c *ChunkModificationSystem) AddModification(modType string, x, y int, radius float64) error {
	chunkX := x / ChunkSize
	chunkY := y / ChunkSize
	chunkID := chunkCoordsToID(chunkX, chunkY)

	// Get or create chunk
	chunk, exists := c.state.ChunkData[chunkID]
	if !exists {
		chunk = &Chunk{
			X:             chunkX,
			Y:             chunkY,
			Terrain:       nil, // nil = use seed generation
			Modifications: []TerrainMod{},
		}
		c.state.ChunkData[chunkID] = chunk
	}

	// Add modification
	mod := TerrainMod{
		Type:      modType,
		X:         x,
		Y:         y,
		Radius:    radius,
		Timestamp: time.Now().UnixMilli(),
	}
	chunk.Modifications = append(chunk.Modifications, mod)

	// Mark chunk as dirty
	c.dirtyChunks[chunkID] = true

	return nil
}

// GetModifiedChunks returns a list of chunk IDs that have been modified
func (c *ChunkModificationSystem) GetModifiedChunks() []string {
	modified := make([]string, 0, len(c.dirtyChunks))
	for chunkID, isDirty := range c.dirtyChunks {
		if isDirty {
			modified = append(modified, chunkID)
		}
	}
	return modified
}

// ClearDirtyFlags clears all dirty chunk flags (called after save)
func (c *ChunkModificationSystem) ClearDirtyFlags() {
	c.dirtyChunks = make(map[string]bool)
	c.state.ModifiedChunks = c.dirtyChunks
}

// GetModificationCount returns the number of modifications in a chunk
func (c *ChunkModificationSystem) GetModificationCount(chunkX, chunkY int) int {
	chunkID := chunkCoordsToID(chunkX, chunkY)
	chunk, exists := c.state.ChunkData[chunkID]
	if !exists {
		return 0
	}
	return len(chunk.Modifications)
}

// HasModifications returns true if a chunk has been modified
func (c *ChunkModificationSystem) HasModifications(chunkX, chunkY int) bool {
	chunkID := chunkCoordsToID(chunkX, chunkY)
	return c.dirtyChunks[chunkID]
}

// MarkDirty flags the chunk at (chunkX, chunkY) as modified without applying
// any terrain change. Use this when a chunk is being evicted to ensure it is
// included in the next incremental save pass.
func (c *ChunkModificationSystem) MarkDirty(chunkX, chunkY int) {
	chunkID := chunkCoordsToID(chunkX, chunkY)
	c.dirtyChunks[chunkID] = true
	c.state.ModifiedChunks[chunkID] = true
}
