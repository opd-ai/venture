package housing

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// SaveData represents the serializable state of all housing data.
type SaveData struct {
	Version string  `json:"version"`
	Plots   []*Plot `json:"plots"`
}

var createSaveFile = func(filename string) (io.WriteCloser, error) {
	return os.Create(filename)
}

// Save writes housing data to a compressed JSON file.
func (m *Manager) Save(filename string) (err error) {
	// Create directory if needed
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.WithFields(log.Fields{
			"filename": filename,
			"error":    err.Error(),
		}).Error("failed to create directory for housing save")
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
	file, err := createSaveFile(filename)
	if err != nil {
		log.WithFields(log.Fields{
			"filename": filename,
			"error":    err.Error(),
		}).Error("failed to create housing save file")
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			if err == nil {
				err = fmt.Errorf("failed to close file: %w", closeErr)
			}
			log.WithFields(log.Fields{
				"filename": filename,
				"error":    closeErr.Error(),
			}).Error("failed to close housing save file")
		}
	}()

	// Create gzip writer
	gzWriter := gzip.NewWriter(file)
	defer func() {
		if closeErr := gzWriter.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close gzip writer: %w", closeErr)
			log.WithFields(log.Fields{
				"filename": filename,
				"error":    closeErr.Error(),
			}).Error("failed to close gzip writer for housing save")
		}
	}()

	// Encode JSON
	encoder := json.NewEncoder(gzWriter)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(saveData); err != nil {
		log.WithFields(log.Fields{
			"filename":   filename,
			"plot_count": len(plots),
			"error":      err.Error(),
		}).Error("failed to encode housing data")
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// Load reads housing data from a compressed JSON file.
func (m *Manager) Load(filename string) error {
	// Open file
	file, err := os.Open(filename)
	if err != nil {
		log.WithFields(log.Fields{
			"filename": filename,
			"error":    err.Error(),
		}).Error("failed to open housing save file")
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		log.WithFields(log.Fields{
			"filename": filename,
			"error":    err.Error(),
		}).Error("failed to create gzip reader for housing data")
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Decode JSON
	var saveData SaveData
	decoder := json.NewDecoder(gzReader)
	if err := decoder.Decode(&saveData); err != nil {
		log.WithFields(log.Fields{
			"filename": filename,
			"error":    err.Error(),
		}).Error("failed to decode housing data")
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	// Validate version
	if saveData.Version != "1.0" {
		log.WithFields(log.Fields{
			"filename": filename,
			"version":  saveData.Version,
		}).Error("unsupported housing save version")
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
func (m *Manager) SavePlayerData(playerID, filename string) (err error) {
	plots := m.GetPlayerPlots(playerID)

	saveData := SaveData{
		Version: "1.0",
		Plots:   plots,
	}

	// Create directory if needed
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.WithFields(log.Fields{
			"playerID": playerID,
			"filename": filename,
			"error":    err.Error(),
		}).Error("failed to create directory for player housing save")
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	file, err := createSaveFile(filename)
	if err != nil {
		log.WithFields(log.Fields{
			"playerID": playerID,
			"filename": filename,
			"error":    err.Error(),
		}).Error("failed to create player housing save file")
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			if err == nil {
				err = fmt.Errorf("failed to close file: %w", closeErr)
			}
			log.WithFields(log.Fields{
				"playerID": playerID,
				"filename": filename,
				"error":    closeErr.Error(),
			}).Error("failed to close player housing save file")
		}
	}()

	// Create gzip writer
	gzWriter := gzip.NewWriter(file)
	defer func() {
		if closeErr := gzWriter.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close gzip writer: %w", closeErr)
			log.WithFields(log.Fields{
				"playerID": playerID,
				"filename": filename,
				"error":    closeErr.Error(),
			}).Error("failed to close gzip writer for player housing save")
		}
	}()

	// Encode JSON
	encoder := json.NewEncoder(gzWriter)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(saveData); err != nil {
		log.WithFields(log.Fields{
			"playerID":   playerID,
			"filename":   filename,
			"plot_count": len(plots),
			"error":      err.Error(),
		}).Error("failed to encode player housing data")
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}
