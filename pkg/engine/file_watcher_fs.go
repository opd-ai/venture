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
// Returns SHA256 hash of mod file contents.
func (w *FileSystemFileWatcher) GetFileHash(modID string) (string, error) {
	w.mu.RLock()
	cached, exists := w.cache[modID]
	w.mu.RUnlock()

	if exists {
		return cached.hash, nil
	}

	// Cache miss - compute hash from filesystem
	data, err := w.readModFile(modID)
	if err != nil {
		return "", err
	}

	hash := ComputeHash(data)

	// Update cache
	version, _ := w.extractVersion(data)
	w.mu.Lock()
	w.cache[modID] = &fileSystemMod{
		id:      modID,
		version: version,
		hash:    hash,
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
// Extracts version field from mod JSON.
func (w *FileSystemFileWatcher) GetModVersion(modID string) (string, error) {
	w.mu.RLock()
	cached, exists := w.cache[modID]
	w.mu.RUnlock()

	if exists {
		return cached.version, nil
	}

	// Cache miss - read from filesystem
	data, err := w.readModFile(modID)
	if err != nil {
		return "", err
	}

	version, err := w.extractVersion(data)
	if err != nil {
		return "", err
	}

	// Update cache
	hash := ComputeHash(data)
	w.mu.Lock()
	w.cache[modID] = &fileSystemMod{
		id:      modID,
		version: version,
		hash:    hash,
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
