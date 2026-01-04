package modding

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Loader handles loading mods from disk.
type Loader struct {
	config ModConfig
}

// NewLoader creates a new mod loader with default configuration.
func NewLoader() *Loader {
	return &Loader{
		config: DefaultConfig(),
	}
}

// NewLoaderWithConfig creates a new mod loader with custom configuration.
func NewLoaderWithConfig(config ModConfig) *Loader {
	return &Loader{
		config: config,
	}
}

// LoadFromFile loads a single mod from a JSON file.
func (l *Loader) LoadFromFile(path string) (*Mod, error) {
	// Validate path is within mods directory if sandboxing enabled
	if l.config.EnableSandbox {
		sandbox := NewSandboxWithConfig(SandboxConfig{
			ModsDirectory: l.config.ModsDirectory,
		})
		if err := sandbox.ValidatePath(path); err != nil {
			return nil, &LoadError{
				ModID: path,
				Err:   err,
			}
		}
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &LoadError{ModID: path, Err: err}
	}

	// Check file size limit
	if l.config.EnableSandbox && len(data) > 1024*1024 {
		return nil, &LoadError{
			ModID: path,
			Err:   fmt.Errorf("mod file exceeds 1MB size limit"),
		}
	}

	// Parse JSON
	var mod Mod
	if err := json.Unmarshal(data, &mod); err != nil {
		return nil, &LoadError{ModID: path, Err: fmt.Errorf("invalid JSON: %w", err)}
	}

	// Set load timestamp
	mod.LoadedAt = time.Now()

	// Validate mod structure
	if err := mod.Validate(); err != nil {
		return nil, &LoadError{ModID: mod.ID, Err: err}
	}

	// Validate mod content against sandbox rules
	if l.config.EnableSandbox {
		sandbox := NewSandbox()
		result := sandbox.ValidateMod(&mod)
		if !result.Valid {
			errMsgs := make([]string, 0, len(result.Errors))
			for _, e := range result.Errors {
				errMsgs = append(errMsgs, e.Error())
			}
			errMsg := strings.Join(errMsgs, "; ")
			if len(errMsgs) > 0 {
				errMsg += "; "
			}
			return nil, &LoadError{ModID: mod.ID, Err: fmt.Errorf("sandbox validation failed: %s", errMsg)}
		}
	}

	return &mod, nil
}

// LoadAll loads all mods from the mods directory.
func (l *Loader) LoadAll() ([]*Mod, error) {
	// Check if directory exists
	if _, err := os.Stat(l.config.ModsDirectory); os.IsNotExist(err) {
		// Create directory if it doesn't exist
		if err := os.MkdirAll(l.config.ModsDirectory, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create mods directory: %w", err)
		}
		return []*Mod{}, nil
	}

	// Read directory
	entries, err := os.ReadDir(l.config.ModsDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to read mods directory: %w", err)
	}

	var mods []*Mod
	var loadErrors []error

	// Load each JSON file
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		path := filepath.Join(l.config.ModsDirectory, entry.Name())
		mod, err := l.LoadFromFile(path)
		if err != nil {
			loadErrors = append(loadErrors, err)
			continue
		}

		mods = append(mods, mod)

		// Check max mods limit
		if len(mods) >= l.config.MaxMods {
			break
		}
	}

	// If all loads failed, return error
	if len(mods) == 0 && len(loadErrors) > 0 {
		return nil, fmt.Errorf("failed to load any mods: %w", errors.Join(loadErrors...))
	}

	return mods, nil
}

// SaveToFile saves a mod to a JSON file.
func (l *Loader) SaveToFile(mod *Mod, path string) error {
	// Validate mod
	if err := mod.Validate(); err != nil {
		return &LoadError{ModID: mod.ID, Err: err}
	}

	// Validate path is within mods directory if sandboxing enabled
	if l.config.EnableSandbox {
		sandbox := NewSandboxWithConfig(SandboxConfig{
			ModsDirectory: l.config.ModsDirectory,
		})
		if err := sandbox.ValidatePath(path); err != nil {
			return &LoadError{
				ModID: mod.ID,
				Err:   err,
			}
		}

		// Validate mod content against sandbox rules
		result := sandbox.ValidateMod(mod)
		if !result.Valid {
			errMsgs := make([]string, 0, len(result.Errors))
			for _, e := range result.Errors {
				errMsgs = append(errMsgs, e.Error())
			}
			errMsg := strings.Join(errMsgs, "; ")
			if len(errMsgs) > 0 {
				errMsg += "; "
			}
			return &LoadError{ModID: mod.ID, Err: fmt.Errorf("sandbox validation failed: %s", errMsg)}
		}
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(mod, "", "  ")
	if err != nil {
		return &LoadError{ModID: mod.ID, Err: fmt.Errorf("failed to marshal JSON: %w", err)}
	}

	// Write file
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return &LoadError{ModID: mod.ID, Err: err}
	}

	return nil
}

// GetModPath returns the file path for a mod by ID.
func (l *Loader) GetModPath(modID string) string {
	return filepath.Join(l.config.ModsDirectory, modID+".json")
}
