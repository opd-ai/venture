package world

import (
	"fmt"
	"math"
)

// ChunkSize is the size of a chunk in tiles
const ChunkSize = 32

// ChunkLoaderSystem handles loading/unloading chunks based on player proximity
type ChunkLoaderSystem struct {
	loadRadius   int                    // Chunk radius around player (default: 5)
	loadedChunks map[string]*Chunk      // Currently loaded chunks
	worldSeed    int64                  // Seed for generating new chunks
	persistence  *WorldPersistence      // For loading persisted chunks
	generator    ChunkGenerator         // For generating new chunks
	playerPos    map[uint64]ChunkCoords // Track player positions
	onEvict      func(*Chunk)           // Called when a chunk is evicted from memory
}

// ChunkCoords represents chunk coordinates
type ChunkCoords struct {
	X int
	Y int
}

// ChunkGenerator generates terrain for new chunks
type ChunkGenerator interface {
	GenerateChunk(chunkX, chunkY int, seed int64) (*Chunk, error)
}

// NewChunkLoaderSystem creates a new chunk loader system
func NewChunkLoaderSystem(worldSeed int64, persistence *WorldPersistence, generator ChunkGenerator) *ChunkLoaderSystem {
	return &ChunkLoaderSystem{
		loadRadius:   5,
		loadedChunks: make(map[string]*Chunk),
		worldSeed:    worldSeed,
		persistence:  persistence,
		generator:    generator,
		playerPos:    make(map[uint64]ChunkCoords),
	}
}

// SetLoadRadius sets the chunk load radius
func (c *ChunkLoaderSystem) SetLoadRadius(radius int) {
	c.loadRadius = radius
}

// Update processes chunk loading/unloading based on player positions
func (c *ChunkLoaderSystem) Update(playerPositions map[uint64]struct{ X, Y float64 }) error {
	// Update player chunk positions
	for playerID, pos := range playerPositions {
		c.playerPos[playerID] = ChunkCoords{
			X: int(math.Floor(pos.X / ChunkSize)),
			Y: int(math.Floor(pos.Y / ChunkSize)),
		}
	}

	// Determine which chunks should be loaded
	neededChunks := c.calculateNeededChunks()

	// Load missing chunks
	for chunkID, coords := range neededChunks {
		if _, exists := c.loadedChunks[chunkID]; !exists {
			chunk, err := c.loadChunk(coords.X, coords.Y)
			if err != nil {
				return fmt.Errorf("failed to load chunk %s: %w", chunkID, err)
			}
			c.loadedChunks[chunkID] = chunk
		}
	}

	// Unload distant chunks
	for chunkID, chunk := range c.loadedChunks {
		if _, needed := neededChunks[chunkID]; !needed {
			c.unloadChunk(chunk)
			delete(c.loadedChunks, chunkID)
		}
	}

	return nil
}

// calculateNeededChunks returns chunks that should be loaded based on player positions
func (c *ChunkLoaderSystem) calculateNeededChunks() map[string]ChunkCoords {
	needed := make(map[string]ChunkCoords)

	for _, playerChunk := range c.playerPos {
		for dy := -c.loadRadius; dy <= c.loadRadius; dy++ {
			for dx := -c.loadRadius; dx <= c.loadRadius; dx++ {
				coords := ChunkCoords{
					X: playerChunk.X + dx,
					Y: playerChunk.Y + dy,
				}
				chunkID := chunkCoordsToID(coords.X, coords.Y)
				needed[chunkID] = coords
			}
		}
	}

	return needed
}

// loadChunk loads a chunk from persistence or generates it
func (c *ChunkLoaderSystem) loadChunk(chunkX, chunkY int) (*Chunk, error) {
	chunkID := chunkCoordsToID(chunkX, chunkY)

	// Try loading from persistence
	if c.persistence != nil {
		// Fast path: check for a previously compressed chunk file first.
		if data, err := c.persistence.LoadChunk(chunkX, chunkY); err == nil {
			compressor := &ChunkCompressionSystem{}
			if chunk, decErr := compressor.DecompressChunk(data); decErr == nil {
				return chunk, nil
			}
		}

		// Fallback: full world-state save (legacy path).
		state, err := c.persistence.LoadWorld(c.worldSeed)
		if err == nil && state.ChunkData != nil {
			if chunk, exists := state.ChunkData[chunkID]; exists {
				return chunk, nil
			}
		}
	}

	// Generate new chunk if not found
	if c.generator != nil {
		chunk, err := c.generator.GenerateChunk(chunkX, chunkY, c.worldSeed)
		if err != nil {
			return nil, fmt.Errorf("failed to generate chunk: %w", err)
		}
		return chunk, nil
	}

	// Create empty chunk as fallback
	return &Chunk{
		X:             chunkX,
		Y:             chunkY,
		Terrain:       nil, // nil means use seed-based generation
		Modifications: []TerrainMod{},
	}, nil
}

// SetOnEvict registers a callback that is called with the evicted chunk just
// before it is removed from memory. Use this to compress or persist the chunk.
func (c *ChunkLoaderSystem) SetOnEvict(fn func(*Chunk)) {
	c.onEvict = fn
}

// unloadChunk handles cleanup before unloading
func (c *ChunkLoaderSystem) unloadChunk(chunk *Chunk) {
	if c.onEvict != nil {
		c.onEvict(chunk)
	}
}

// GetChunk returns a loaded chunk by coordinates
func (c *ChunkLoaderSystem) GetChunk(chunkX, chunkY int) (*Chunk, bool) {
	chunkID := chunkCoordsToID(chunkX, chunkY)
	chunk, exists := c.loadedChunks[chunkID]
	return chunk, exists
}

// GetLoadedChunks returns all currently loaded chunks
func (c *ChunkLoaderSystem) GetLoadedChunks() map[string]*Chunk {
	return c.loadedChunks
}

// GetLoadedChunkCount returns the number of loaded chunks
func (c *ChunkLoaderSystem) GetLoadedChunkCount() int {
	return len(c.loadedChunks)
}

// chunkCoordsToID converts chunk coordinates to a string ID
func chunkCoordsToID(chunkX, chunkY int) string {
	return fmt.Sprintf("%d,%d", chunkX, chunkY)
}
