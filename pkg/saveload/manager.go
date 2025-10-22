package saveload

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SaveManager handles save/load operations for game state.
type SaveManager struct {
	// Directory where save files are stored
	saveDir string
}

// NewSaveManager creates a new save manager with the specified save directory.
// The directory will be created if it doesn't exist.
func NewSaveManager(saveDir string) (*SaveManager, error) {
	// Create save directory if it doesn't exist
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create save directory: %w", err)
	}

	return &SaveManager{
		saveDir: saveDir,
	}, nil
}

// SaveGame saves the game state to a file with the specified name.
// The .sav extension is added automatically if not present.
func (m *SaveManager) SaveGame(name string, save *GameSave) error {
	if save == nil {
		return fmt.Errorf("save cannot be nil")
	}

	// Validate save name
	if err := m.validateSaveName(name); err != nil {
		return err
	}

	// Ensure version and timestamp are set
	save.Version = SaveVersion
	save.Timestamp = time.Now()

	// Marshal to JSON with indentation for readability
	data, err := json.MarshalIndent(save, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal save data: %w", err)
	}

	// Write to file
	filename := m.getFilePath(name)
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write save file: %w", err)
	}

	return nil
}

// LoadGame loads the game state from a file with the specified name.
func (m *SaveManager) LoadGame(name string) (*GameSave, error) {
	// Validate save name
	if err := m.validateSaveName(name); err != nil {
		return nil, err
	}

	// Read file
	filename := m.getFilePath(name)
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("save file not found: %s", name)
		}
		return nil, fmt.Errorf("failed to read save file: %w", err)
	}

	// Unmarshal JSON
	var save GameSave
	if err := json.Unmarshal(data, &save); err != nil {
		return nil, fmt.Errorf("failed to parse save file: %w", err)
	}

	// Validate save version and migrate if necessary
	if err := m.validateAndMigrate(&save); err != nil {
		return nil, fmt.Errorf("failed to validate/migrate save: %w", err)
	}

	return &save, nil
}

// DeleteSave deletes a save file.
func (m *SaveManager) DeleteSave(name string) error {
	// Validate save name
	if err := m.validateSaveName(name); err != nil {
		return err
	}

	filename := m.getFilePath(name)
	if err := os.Remove(filename); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("save file not found: %s", name)
		}
		return fmt.Errorf("failed to delete save file: %w", err)
	}

	return nil
}

// ListSaves returns metadata for all save files in the save directory.
func (m *SaveManager) ListSaves() ([]*SaveMetadata, error) {
	// Read directory
	entries, err := os.ReadDir(m.saveDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read save directory: %w", err)
	}

	// Collect metadata for each save file
	var saves []*SaveMetadata
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process .sav files
		if !strings.HasSuffix(entry.Name(), ".sav") {
			continue
		}

		// Get metadata
		metadata, err := m.GetSaveMetadata(strings.TrimSuffix(entry.Name(), ".sav"))
		if err != nil {
			// Skip files that can't be read
			continue
		}

		saves = append(saves, metadata)
	}

	// Sort by timestamp (newest first)
	sort.Slice(saves, func(i, j int) bool {
		return saves[i].Timestamp.After(saves[j].Timestamp)
	})

	return saves, nil
}

// GetSaveMetadata reads metadata from a save file without loading the entire save.
func (m *SaveManager) GetSaveMetadata(name string) (*SaveMetadata, error) {
	// Validate save name
	if err := m.validateSaveName(name); err != nil {
		return nil, err
	}

	filename := m.getFilePath(name)

	// Get file info
	fileInfo, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("save file not found: %s", name)
		}
		return nil, fmt.Errorf("failed to stat save file: %w", err)
	}

	// Open and read file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open save file: %w", err)
	}
	defer file.Close()

	// Decode just enough to get metadata
	var save GameSave
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&save); err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("empty save file: %s", name)
		}
		return nil, fmt.Errorf("failed to decode save file: %w", err)
	}

	// Build metadata
	metadata := &SaveMetadata{
		Name:      name,
		Version:   save.Version,
		Timestamp: save.Timestamp,
		FileSize:  fileInfo.Size(),
	}

	// Add player and world info if available
	if save.PlayerState != nil {
		metadata.PlayerLevel = save.PlayerState.Level
	}
	if save.WorldState != nil {
		metadata.GenreID = save.WorldState.GenreID
		metadata.GameTime = save.WorldState.GameTime
	}

	return metadata, nil
}

// SaveExists checks if a save file exists.
func (m *SaveManager) SaveExists(name string) bool {
	if err := m.validateSaveName(name); err != nil {
		return false
	}

	filename := m.getFilePath(name)
	_, err := os.Stat(filename)
	return err == nil
}

// getFilePath returns the full path to a save file.
func (m *SaveManager) getFilePath(name string) string {
	// Add .sav extension if not present
	if !strings.HasSuffix(name, ".sav") {
		name += ".sav"
	}
	return filepath.Join(m.saveDir, name)
}

// validateSaveName validates that a save name is acceptable.
func (m *SaveManager) validateSaveName(name string) error {
	if name == "" {
		return fmt.Errorf("save name cannot be empty")
	}

	// Remove extension for validation
	name = strings.TrimSuffix(name, ".sav")

	// Check for path separators (security check)
	if strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("save name cannot contain path separators")
	}

	// Check for special characters
	if strings.ContainsAny(name, "<>:\"|?*") {
		return fmt.Errorf("save name contains invalid characters")
	}

	return nil
}

// validateAndMigrate validates a save file and migrates it if necessary.
func (m *SaveManager) validateAndMigrate(save *GameSave) error {
	if save == nil {
		return fmt.Errorf("save cannot be nil")
	}

	// Check version
	if save.Version == "" {
		return fmt.Errorf("save file has no version")
	}

	// For now, we only support the current version
	// Future versions can add migration logic here
	if save.Version != SaveVersion {
		// Example migration logic (not needed yet):
		// if save.Version == "0.9.0" {
		//     migrateSaveFrom090To100(save)
		// }

		// For now, just warn about version mismatch
		// In production, you might want to reject or migrate
		return fmt.Errorf("save file version %s is not supported (current version: %s)", save.Version, SaveVersion)
	}

	// Validate required fields
	if save.PlayerState == nil {
		return fmt.Errorf("save file missing player state")
	}
	if save.WorldState == nil {
		return fmt.Errorf("save file missing world state")
	}
	if save.Settings == nil {
		return fmt.Errorf("save file missing settings")
	}

	return nil
}
