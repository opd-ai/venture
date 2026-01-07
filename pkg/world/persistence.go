package world

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	bgCtx             context.Context       // Background context for legacy methods
}

// NewWorldPersistence creates a new world persistence manager
func NewWorldPersistence(savePath string) *WorldPersistence {
	return &WorldPersistence{
		SavePath:         savePath,
		AutoSaveInterval: 300.0, // 5 minutes
		maxBackups:       3,
		bgCtx:            context.Background(), // Reusable background context for legacy methods
	}
}

// Update handles auto-save timer tracking
// Returns true if auto-save should be triggered by caller
func (w *WorldPersistence) Update(deltaTime float64) bool {
	w.timeSinceLastSave += deltaTime

	if w.timeSinceLastSave >= w.AutoSaveInterval {
		w.timeSinceLastSave = 0
		return true // Signal that auto-save is due
	}
	return false
}

// SaveWorld saves the world state to disk with backup rotation and incremental saves.
// Supports context-based cancellation for long-running operations.
func (w *WorldPersistence) SaveWorld(state *PersistentWorldState) error {
	return w.SaveWorldWithContext(w.bgCtx, state)
}

// SaveWorldWithContext saves the world state to disk with context support.
// The context can be used to cancel long-running save operations.
func (w *WorldPersistence) SaveWorldWithContext(ctx context.Context, state *PersistentWorldState) error {
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

	// Ensure cleanup on error
	var saveErr error
	defer func() {
		f.Close()
		if saveErr != nil {
			// Remove temp file on error
			os.Remove(tempPath)
		}
	}()

	// Check context before expensive operations
	select {
	case <-ctx.Done():
		saveErr = fmt.Errorf("save cancelled: %w", ctx.Err())
		return saveErr
	default:
	}

	// Compress with gzip
	gz := gzip.NewWriter(f)

	// Encode to JSON
	encoder := json.NewEncoder(gz)
	if err := encoder.Encode(state); err != nil {
		saveErr = fmt.Errorf("failed to encode state: %w", err)
		gz.Close()
		return saveErr
	}

	// Check context before finalizing
	select {
	case <-ctx.Done():
		saveErr = fmt.Errorf("save cancelled before finalize: %w", ctx.Err())
		gz.Close()
		return saveErr
	default:
	}

	// Flush gzip writer
	if err := gz.Close(); err != nil {
		saveErr = fmt.Errorf("failed to flush gzip: %w", err)
		return saveErr
	}

	// Close file
	if err := f.Close(); err != nil {
		saveErr = fmt.Errorf("failed to close temp file: %w", err)
		return saveErr
	}

	// Atomic rename
	if err := os.Rename(tempPath, w.SavePath); err != nil {
		saveErr = fmt.Errorf("failed to rename temp file: %w", err)
		return saveErr
	}

	// Store reference for incremental saves
	w.lastSaveState = state

	return nil
}

// SaveIncremental saves only modified chunks since last save.
// Supports context-based cancellation.
func (w *WorldPersistence) SaveIncremental(state *PersistentWorldState) error {
	return w.SaveIncrementalWithContext(w.bgCtx, state)
}

// SaveIncrementalWithContext saves only modified chunks with context support.
func (w *WorldPersistence) SaveIncrementalWithContext(ctx context.Context, state *PersistentWorldState) error {
	if w.lastSaveState == nil {
		// No previous save, do full save
		return w.SaveWorldWithContext(ctx, state)
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
		return w.SaveWorldWithContext(ctx, state)
	}

	return w.SaveWorldWithContext(ctx, incrementalState)
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

// copyFile copies a file from src to dst with proper error handling.
// Uses io.Copy to handle partial reads/writes and large files efficiently.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}

	// Ensure cleanup on error
	var copyErr error
	defer func() {
		dstFile.Close()
		if copyErr != nil {
			// Remove incomplete destination file on error
			os.Remove(dst)
		}
	}()

	// Copy with proper handling of partial writes
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		copyErr = fmt.Errorf("failed to copy data: %w", err)
		return copyErr
	}

	// Sync to ensure data is written to disk
	if err := dstFile.Sync(); err != nil {
		copyErr = fmt.Errorf("failed to sync destination: %w", err)
		return copyErr
	}

	return nil
}

// LoadWorld loads the world state from disk.
// Attempts to load from main save file, falls back to backups on error.
func (w *WorldPersistence) LoadWorld(seed int64) (*PersistentWorldState, error) {
	return w.LoadWorldWithContext(w.bgCtx, seed)
}

// LoadWorldWithContext loads the world state from disk with context support.
// The context can be used to cancel long-running load operations.
func (w *WorldPersistence) LoadWorldWithContext(ctx context.Context, seed int64) (*PersistentWorldState, error) {
	state, err := w.loadFromPathWithContext(ctx, w.SavePath, seed)
	if err == nil {
		return state, nil
	}

	// Try backups if main save fails
	for i := 1; i <= w.maxBackups; i++ {
		// Check context before trying next backup
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("load cancelled: %w", ctx.Err())
		default:
		}

		backupPath := fmt.Sprintf("%s.%d", w.SavePath, i)
		state, backupErr := w.loadFromPathWithContext(ctx, backupPath, seed)
		if backupErr == nil {
			// Successfully loaded from backup
			return state, nil
		}
	}

	// All loads failed, return original error
	return nil, err
}

// loadFromPath loads state from a specific file path (legacy, no context).
func (w *WorldPersistence) loadFromPath(path string, seed int64) (*PersistentWorldState, error) {
	return w.loadFromPathWithContext(w.bgCtx, path, seed)
}

// loadFromPathWithContext loads state from a specific file path with context support.
func (w *WorldPersistence) loadFromPathWithContext(ctx context.Context, path string, seed int64) (*PersistentWorldState, error) {
	// Check context before opening file
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("load cancelled: %w", ctx.Err())
	default:
	}

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

	// Check context before expensive decode
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("load cancelled before decode: %w", ctx.Err())
	default:
	}

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
