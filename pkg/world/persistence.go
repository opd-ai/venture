package world

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// CurrentSchemaVersion is the current version of the persistent world state schema
const CurrentSchemaVersion = 1

// PersistentWorldState represents saveable world state
// Matches ROADMAP_V6.md Phase 37 specification
type PersistentWorldState struct {
	Version        int               `json:"version"`      // Schema version for migrations
	WorldSeed      int64             `json:"world_seed"`   // Deterministic generation seed
	ChunkData      map[string]*Chunk `json:"chunk_data"`   // Sparse storage, only modified chunks
	Entities       []*EntityState    `json:"entities"`     // Living entities (NPCs, monsters, items)
	WorldEvents    []WorldEvent      `json:"world_events"` // Global events (wars, disasters)
	Timestamp      int64             `json:"timestamp"`    // Last save time (Unix milliseconds)
	ModifiedChunks map[string]bool   `json:"-"`            // Track dirty chunks (not serialized)
}

// Chunk represents a world chunk with terrain modifications
type Chunk struct {
	X             int          `json:"x"`
	Y             int          `json:"y"`
	Terrain       [][]TileType `json:"terrain,omitempty"` // Modified terrain (nil = use seed generation)
	Modifications []TerrainMod `json:"modifications"`     // Explosions, dug tunnels, built structures
}

// TerrainMod represents a modification to terrain
type TerrainMod struct {
	Type      string  `json:"type"` // "explosion", "dig", "build"
	X         int     `json:"x"`
	Y         int     `json:"y"`
	Radius    float64 `json:"radius"`
	Timestamp int64   `json:"timestamp"`
}

// EntityState represents serializable entity data
type EntityState struct {
	ID         uint64                 `json:"id"`
	TypeName   string                 `json:"type_name"`  // "Monster", "NPC", "Item"
	Components map[string]interface{} `json:"components"` // Serialized component data
}

// WorldEvent represents a global world event
type WorldEvent struct {
	Type      string                 `json:"type"` // "war", "disaster", "festival"
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data"` // Event-specific data
}

// WorldPersistence handles world saving/loading with backup rotation
type WorldPersistence struct {
	SavePath          string
	AutoSaveInterval  float64
	timeSinceLastSave float64
	maxBackups        int                   // Number of backups to keep (default: 3)
	lastSaveState     *PersistentWorldState // For incremental saves
}

// NewWorldPersistence creates a new world persistence manager
func NewWorldPersistence(savePath string) *WorldPersistence {
	return &WorldPersistence{
		SavePath:         savePath,
		AutoSaveInterval: 300.0, // 5 minutes
		maxBackups:       3,
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

// SaveWorld saves the world state to disk with backup rotation and incremental saves
func (w *WorldPersistence) SaveWorld(state *PersistentWorldState) error {
	// Update timestamp
	state.Timestamp = time.Now().UnixMilli()

	// Rotate backups before saving
	if err := w.rotateBackups(); err != nil {
		return fmt.Errorf("failed to rotate backups: %w", err)
	}

	// Ensure save directory exists
	if err := os.MkdirAll(filepath.Dir(w.SavePath), 0o755); err != nil {
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

	// Store reference for incremental saves
	w.lastSaveState = state

	return nil
}

// SaveIncremental saves only modified chunks since last save
func (w *WorldPersistence) SaveIncremental(state *PersistentWorldState) error {
	if w.lastSaveState == nil {
		// No previous save, do full save
		return w.SaveWorld(state)
	}

	// Create incremental state with only modified chunks
	incrementalState := &PersistentWorldState{
		Version:     state.Version,
		WorldSeed:   state.WorldSeed,
		ChunkData:   make(map[string]*Chunk),
		Entities:    state.Entities, // Always save all entities (size is manageable)
		WorldEvents: state.WorldEvents,
		Timestamp:   time.Now().UnixMilli(),
	}

	// Copy only modified chunks
	for chunkID, chunk := range state.ChunkData {
		if state.ModifiedChunks[chunkID] {
			incrementalState.ChunkData[chunkID] = chunk
		}
	}

	// If too many chunks modified (>50%), do full save instead
	modifiedCount := len(incrementalState.ChunkData)
	totalCount := len(state.ChunkData)
	if totalCount > 0 && float64(modifiedCount)/float64(totalCount) > 0.5 {
		return w.SaveWorld(state)
	}

	return w.SaveWorld(incrementalState)
}

// rotateBackups keeps the last maxBackups save files
func (w *WorldPersistence) rotateBackups() error {
	// Check if current save exists
	if _, err := os.Stat(w.SavePath); os.IsNotExist(err) {
		return nil // No current save, nothing to backup
	}

	// Shift existing backups: .2 -> .3, .1 -> .2, current -> .1
	for i := w.maxBackups - 1; i > 0; i-- {
		oldPath := fmt.Sprintf("%s.%d", w.SavePath, i)
		newPath := fmt.Sprintf("%s.%d", w.SavePath, i+1)

		// Delete oldest backup if exists
		if i == w.maxBackups-1 {
			os.Remove(newPath)
		}

		// Rename if exists
		if _, err := os.Stat(oldPath); err == nil {
			if err := os.Rename(oldPath, newPath); err != nil {
				return fmt.Errorf("failed to rotate backup %d: %w", i, err)
			}
		}
	}

	// Copy current save to .1
	backupPath := w.SavePath + ".1"
	if err := copyFile(w.SavePath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

// LoadWorld loads the world state from disk
func (w *WorldPersistence) LoadWorld(seed int64) (*PersistentWorldState, error) {
	state, err := w.loadFromPath(w.SavePath, seed)
	if err == nil {
		return state, nil
	}

	// Try backups if main save fails
	for i := 1; i <= w.maxBackups; i++ {
		backupPath := fmt.Sprintf("%s.%d", w.SavePath, i)
		state, backupErr := w.loadFromPath(backupPath, seed)
		if backupErr == nil {
			// Successfully loaded from backup
			return state, nil
		}
	}

	// All loads failed, return original error
	return nil, err
}

// loadFromPath loads state from a specific file path
func (w *WorldPersistence) loadFromPath(path string, seed int64) (*PersistentWorldState, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No save file, create new state
			return &PersistentWorldState{
				Version:        CurrentSchemaVersion,
				WorldSeed:      seed,
				ChunkData:      make(map[string]*Chunk),
				Entities:       []*EntityState{},
				WorldEvents:    []WorldEvent{},
				Timestamp:      0,
				ModifiedChunks: make(map[string]bool),
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

	// Initialize transient fields
	if state.ModifiedChunks == nil {
		state.ModifiedChunks = make(map[string]bool)
	}

	// OBSOLETE CODE REMOVED: Save format migration (migrateState function)
	// Replaced by: Pre-1.0 policy - only latest save format supported, no backward compatibility
	// Removed: migrateState() function and all migration logic
	// Roadmap: ROADMAP_V6.md Phase 37 - persistence without legacy support
	// PRE-1.0: Only CurrentSchemaVersion is supported
	if state.Version != CurrentSchemaVersion {
		return nil, fmt.Errorf("incompatible save version %d (expected %d) - no migration support before v1.0", state.Version, CurrentSchemaVersion)
	}

	return &state, nil
}

// CleanupBackups removes backup files
func (w *WorldPersistence) CleanupBackups() error {
	var errors []error

	for i := 1; i <= w.maxBackups; i++ {
		backupPath := fmt.Sprintf("%s.%d", w.SavePath, i)
		if err := os.Remove(backupPath); err != nil && !os.IsNotExist(err) {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to cleanup %d backups", len(errors))
	}

	return nil
}

// ListBackups returns paths to all available backup files
func (w *WorldPersistence) ListBackups() []string {
	var backups []string

	for i := 1; i <= w.maxBackups; i++ {
		backupPath := fmt.Sprintf("%s.%d", w.SavePath, i)
		if _, err := os.Stat(backupPath); err == nil {
			backups = append(backups, backupPath)
		}
	}

	// Sort by modification time (newest first)
	sort.Slice(backups, func(i, j int) bool {
		infoI, _ := os.Stat(backups[i])
		infoJ, _ := os.Stat(backups[j])
		return infoI.ModTime().After(infoJ.ModTime())
	})

	return backups
}
