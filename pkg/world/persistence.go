package world

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PersistentWorldState represents saveable world state
type PersistentWorldState struct {
	Version    int
	WorldSeed  int64
	ChunkData  map[string]*Chunk
	Timestamp  int64
}

// Chunk represents a world chunk
type Chunk struct {
	X             int
	Y             int
	Modifications []string // Terrain modifications
}

// WorldPersistence handles world saving/loading
type WorldPersistence struct {
	SavePath         string
	AutoSaveInterval float64
	timeSinceLastSave float64
}

// NewWorldPersistence creates a new world persistence manager
func NewWorldPersistence(savePath string) *WorldPersistence {
	return &WorldPersistence{
		SavePath:         savePath,
		AutoSaveInterval: 300.0, // 5 minutes
	}
}

// Update handles auto-save
func (w *WorldPersistence) Update(deltaTime float64) {
	w.timeSinceLastSave += deltaTime
	
	if w.timeSinceLastSave >= w.AutoSaveInterval {
		w.timeSinceLastSave = 0
		// TODO: Trigger auto-save
	}
}

// SaveWorld saves the world state to disk
func (w *WorldPersistence) SaveWorld(state *PersistentWorldState) error {
	// Ensure save directory exists
	if err := os.MkdirAll(filepath.Dir(w.SavePath), 0755); err != nil {
		return fmt.Errorf("failed to create save directory: %w", err)
	}
	
	// Create temp file for atomic write
	tempPath := w.SavePath + ".tmp"
	f, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer f.Close()
	
	// Compress with gzip
	gz := gzip.NewWriter(f)
	defer gz.Close()
	
	// Encode to JSON
	encoder := json.NewEncoder(gz)
	if err := encoder.Encode(state); err != nil {
		return fmt.Errorf("failed to encode state: %w", err)
	}
	
	// Flush gzip writer
	if err := gz.Close(); err != nil {
		return fmt.Errorf("failed to flush gzip: %w", err)
	}
	
	// Close file
	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	
	// Atomic rename
	if err := os.Rename(tempPath, w.SavePath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}
	
	return nil
}

// LoadWorld loads the world state from disk
func (w *WorldPersistence) LoadWorld(seed int64) (*PersistentWorldState, error) {
	f, err := os.Open(w.SavePath)
	if err != nil {
		if os.IsNotExist(err) {
			// No save file, create new state
			return &PersistentWorldState{
				Version:   1,
				WorldSeed: seed,
				ChunkData: make(map[string]*Chunk),
				Timestamp: 0,
			}, nil
		}
		return nil, fmt.Errorf("failed to open save file: %w", err)
	}
	defer f.Close()
	
	// Decompress with gzip
	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gz.Close()
	
	// Decode JSON
	var state PersistentWorldState
	decoder := json.NewDecoder(gz)
	if err := decoder.Decode(&state); err != nil {
		return nil, fmt.Errorf("failed to decode state: %w", err)
	}
	
	return &state, nil
}
