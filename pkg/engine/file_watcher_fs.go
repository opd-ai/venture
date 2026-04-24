package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
)

// FileSystemFileWatcher implements FileWatcher interface for filesystem-based mod watching.
// It monitors a directory of mod JSON files for hot-reload support.
type FileSystemFileWatcher struct {
	modsDir string
	cache   map[string]*fileSystemMod
	mu      sync.RWMutex
	logger  *logrus.Logger
}

type fileSystemMod struct {
	id      string
	version string
	hash    string
	modtime int64 // Unix nanosecond timestamp of the file when last hashed
}

// NewFileSystemFileWatcher creates a new filesystem-based file watcher.
// modsDir defaults to "mods" if empty string is provided.
func NewFileSystemFileWatcher(modsDir string) *FileSystemFileWatcher {
	if modsDir == "" {
		modsDir = "mods"
	}
	return &FileSystemFileWatcher{
		modsDir: modsDir,
		cache:   make(map[string]*fileSystemMod),
		logger:  logrus.New(),
	}
}

// SetLogger sets custom logger for debug/info messages.
func (w *FileSystemFileWatcher) SetLogger(logger *logrus.Logger) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.logger = logger
}

// GetFileHash implements FileWatcher interface.
// Returns SHA256 hash of mod file contents.  The cache entry is automatically
// invalidated when the file's modification time has changed since it was last
// hashed, so callers detect on-disk changes without explicit cache management.
func (w *FileSystemFileWatcher) GetFileHash(modID string) (string, error) {
	filename := filepath.Join(w.modsDir, modID+".json")

	// Stat the file first to check modtime before acquiring a write lock.
	info, statErr := os.Stat(filename)

	w.mu.RLock()
	cached, exists := w.cache[modID]
	w.mu.RUnlock()

	if exists && statErr == nil && info.ModTime().UnixNano() == cached.modtime {
		// File is unchanged since last hash — return cached value.
		return cached.hash, nil
	}

	// Cache miss or file modified — (re-)read and hash from filesystem.
	data, err := w.readModFile(modID)
	if err != nil {
		return "", err
	}

	hash := ComputeHash(data)

	// Update cache, recording the current modtime so future calls can detect changes.
	version, _ := w.extractVersion(data)
	var modtime int64
	if statErr == nil {
		modtime = info.ModTime().UnixNano()
	}
	w.mu.Lock()
	w.cache[modID] = &fileSystemMod{
		id:      modID,
		version: version,
		hash:    hash,
		modtime: modtime,
	}
	w.mu.Unlock()

	return hash, nil
}

// GetModData implements FileWatcher interface.
// Returns raw JSON data from mod file.
func (w *FileSystemFileWatcher) GetModData(modID string) ([]byte, error) {
	return w.readModFile(modID)
}

// GetModVersion implements FileWatcher interface.
// Extracts version field from mod JSON.  Like GetFileHash, the cached value is
// automatically invalidated when the file's modtime changes.
func (w *FileSystemFileWatcher) GetModVersion(modID string) (string, error) {
	filename := filepath.Join(w.modsDir, modID+".json")
	info, statErr := os.Stat(filename)

	w.mu.RLock()
	cached, exists := w.cache[modID]
	w.mu.RUnlock()

	if exists && statErr == nil && info.ModTime().UnixNano() == cached.modtime {
		return cached.version, nil
	}

	// Cache miss or stale — read from filesystem.
	data, err := w.readModFile(modID)
	if err != nil {
		return "", err
	}

	version, err := w.extractVersion(data)
	if err != nil {
		return "", err
	}

	hash := ComputeHash(data)
	var modtime int64
	if statErr == nil {
		modtime = info.ModTime().UnixNano()
	}
	w.mu.Lock()
	w.cache[modID] = &fileSystemMod{
		id:      modID,
		version: version,
		hash:    hash,
		modtime: modtime,
	}
	w.mu.Unlock()

	return version, nil
}

// InvalidateCache clears the cache for a specific mod ID.
// This forces the next access to re-read from filesystem.
func (w *FileSystemFileWatcher) InvalidateCache(modID string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.cache, modID)
}

// InvalidateAllCache clears the entire cache.
func (w *FileSystemFileWatcher) InvalidateAllCache() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.cache = make(map[string]*fileSystemMod)
}

// readModFile reads mod JSON file from filesystem.
func (w *FileSystemFileWatcher) readModFile(modID string) ([]byte, error) {
	filename := filepath.Join(w.modsDir, modID+".json")
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("mod %s not found", modID)
		}
		return nil, fmt.Errorf("failed to read mod file %s: %w", filename, err)
	}
	return data, nil
}

// extractVersion parses JSON and extracts version field.
func (w *FileSystemFileWatcher) extractVersion(data []byte) (string, error) {
	var modData struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &modData); err != nil {
		return "", fmt.Errorf("failed to parse mod JSON: %w", err)
	}
	if modData.Version == "" {
		return "1.0.0", nil // Default version if not specified
	}
	return modData.Version, nil
}
