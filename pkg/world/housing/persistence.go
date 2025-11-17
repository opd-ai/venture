package housing

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SaveData represents the serializable state of all housing data.
type SaveData struct {
	Version string  `json:"version"`
	Plots   []*Plot `json:"plots"`
}

// Save writes housing data to a compressed JSON file.
func (m *Manager) Save(filename string) error {
	// Create directory if needed
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Collect all plots
	plots := make([]*Plot, 0, len(m.plots))
	for _, plot := range m.plots {
		plots = append(plots, plot)
	}

	saveData := SaveData{
		Version: "1.0",
		Plots:   plots,
	}

	// Create file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	// Encode JSON
	encoder := json.NewEncoder(gzWriter)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(saveData); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// Load reads housing data from a compressed JSON file.
func (m *Manager) Load(filename string) error {
	// Open file
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Decode JSON
	var saveData SaveData
	decoder := json.NewDecoder(gzReader)
	if err := decoder.Decode(&saveData); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	// Validate version
	if saveData.Version != "1.0" {
		return fmt.Errorf("unsupported save version: %s", saveData.Version)
	}

	// Clear existing data
	m.Clear()

	// Load plots
	for _, plot := range saveData.Plots {
		// Reconstruct permission set if nil
		if plot.Permissions == nil {
			plot.Permissions = NewPermissionSet()
		}

		m.plots[plot.ID] = plot
		m.playerPlots[plot.OwnerID] = append(m.playerPlots[plot.OwnerID], plot)
		m.spatialGrid.Insert(plot)
	}

	return nil
}

// SavePlayerData saves housing data for a specific player.
func (m *Manager) SavePlayerData(playerID, filename string) error {
	plots := m.GetPlayerPlots(playerID)

	saveData := SaveData{
		Version: "1.0",
		Plots:   plots,
	}

	// Create directory if needed
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	// Encode JSON
	encoder := json.NewEncoder(gzWriter)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(saveData); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}
