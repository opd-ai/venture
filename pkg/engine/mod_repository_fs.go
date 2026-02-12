// Package engine provides filesystem-based mod repository implementation.
// FileSystemModRepository provides production mod repository backed by local directory.
package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/modding"
	"github.com/sirupsen/logrus"
)

// FileSystemModRepository implements ModRepository using local filesystem.
// It scans a directory for JSON mod files and provides them to the mod browser.
type FileSystemModRepository struct {
	modsDir string
	cache   []ModListing
	cacheMu sync.RWMutex
	logger  *logrus.Logger
}

// NewFileSystemModRepository creates a new filesystem-based mod repository.
// modsDir is the directory to scan for mod files (defaults to "mods" if empty).
func NewFileSystemModRepository(modsDir string) *FileSystemModRepository {
	if modsDir == "" {
		modsDir = "mods"
	}

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &FileSystemModRepository{
		modsDir: modsDir,
		cache:   nil,
		logger:  logger,
	}
}

// SetLogger sets a custom logger for the repository.
func (r *FileSystemModRepository) SetLogger(logger *logrus.Logger) {
	r.logger = logger
}

// FetchMods scans the mods directory and returns available mods.
func (r *FileSystemModRepository) FetchMods() ([]ModListing, error) {
	r.cacheMu.Lock()
	defer r.cacheMu.Unlock()

	// Check if directory exists
	if _, err := os.Stat(r.modsDir); os.IsNotExist(err) {
		r.logger.WithFields(logrus.Fields{
			"mods_dir": r.modsDir,
		}).Warn("Mods directory does not exist")
		return []ModListing{}, nil
	}

	// Scan directory for .json files
	entries, err := os.ReadDir(r.modsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read mods directory: %w", err)
	}

	var listings []ModListing
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		modPath := filepath.Join(r.modsDir, entry.Name())
		listing, err := r.loadModListing(modPath)
		if err != nil {
			r.logger.WithFields(logrus.Fields{
				"file":  entry.Name(),
				"error": err.Error(),
			}).Warn("Failed to load mod file")
			continue
		}

		listings = append(listings, listing)
	}

	r.cache = listings
	r.logger.WithFields(logrus.Fields{
		"count":    len(listings),
		"mods_dir": r.modsDir,
	}).Debug("Fetched mods from filesystem")

	return listings, nil
}

// loadModListing loads and parses a mod file into a ModListing.
func (r *FileSystemModRepository) loadModListing(modPath string) (ModListing, error) {
	data, err := os.ReadFile(modPath)
	if err != nil {
		return ModListing{}, fmt.Errorf("failed to read file: %w", err)
	}

	var mod modding.Mod
	if err := json.Unmarshal(data, &mod); err != nil {
		return ModListing{}, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if err := mod.Validate(); err != nil {
		return ModListing{}, fmt.Errorf("invalid mod: %w", err)
	}

	fileInfo, err := os.Stat(modPath)
	if err != nil {
		return ModListing{}, fmt.Errorf("failed to stat file: %w", err)
	}

	listing := ModListing{
		ID:           mod.ID,
		Name:         mod.Name,
		Author:       mod.Author,
		Description:  mod.Description,
		Version:      mod.Version,
		GameVersion:  "1.0.0",
		Categories:   categoriesFromModType(mod.Type),
		Size:         fileInfo.Size(),
		UploadedAt:   fileInfo.ModTime().Unix(),
		UpdatedAt:    fileInfo.ModTime().Unix(),
		Dependencies: mod.Dependencies,
		Screenshots:  []string{},
		Featured:     false,
		Rating:       0.0,
		RatingCount:  0,
		Downloads:    0,
	}

	return listing, nil
}

// categoriesFromModType maps modding.ModType to category tags.
func categoriesFromModType(modType modding.ModType) []string {
	switch modType {
	case modding.ModTypeRule:
		return []string{"gameplay", "balance"}
	case modding.ModTypeGenerator:
		return []string{"content", "generator"}
	case modding.ModTypeEvent:
		return []string{"events", "gameplay"}
	default:
		return []string{"gameplay"}
	}
}

// DownloadMod reads a mod file and returns its contents.
func (r *FileSystemModRepository) DownloadMod(modID string, progressCallback func(downloaded, total int64)) ([]byte, error) {
	// Find mod in cache or rescan
	r.cacheMu.RLock()
	var modFile string
	if r.cache != nil {
		for _, mod := range r.cache {
			if mod.ID == modID {
				modFile = filepath.Join(r.modsDir, modID+".json")
				break
			}
		}
	}
	r.cacheMu.RUnlock()

	if modFile == "" {
		// Try direct file access
		modFile = filepath.Join(r.modsDir, modID+".json")
	}

	data, err := os.ReadFile(modFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read mod file: %w", err)
	}

	if progressCallback != nil {
		total := int64(len(data))
		step := total / 10
		if step < 1 {
			step = 1
		}
		for i := int64(0); i <= total; i += step {
			progressCallback(i, total)
			time.Sleep(time.Millisecond)
		}
		progressCallback(total, total)
	}

	r.logger.WithFields(logrus.Fields{
		"mod_id": modID,
		"size":   len(data),
	}).Debug("Downloaded mod from filesystem")

	return data, nil
}

// GetModDetails returns detailed information for a specific mod.
func (r *FileSystemModRepository) GetModDetails(modID string) (*ModListing, error) {
	r.cacheMu.RLock()
	defer r.cacheMu.RUnlock()

	if r.cache != nil {
		for i := range r.cache {
			if r.cache[i].ID == modID {
				return &r.cache[i], nil
			}
		}
	}

	modFile := filepath.Join(r.modsDir, modID+".json")
	listing, err := r.loadModListing(modFile)
	if err != nil {
		return nil, fmt.Errorf("mod %s not found: %w", modID, err)
	}

	return &listing, nil
}
