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
	if err := l.validateSandboxPath(path); err != nil {
		return nil, &LoadError{ModID: path, Err: err}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &LoadError{ModID: path, Err: err}
	}

	if err := l.validateFileSize(path, data); err != nil {
		return nil, err
	}

	mod, err := l.parseModJSON(path, data)
	if err != nil {
		return nil, err
	}

	if err := l.validateModContent(mod); err != nil {
		return nil, err
	}

	return mod, nil
}

// validateSandboxPath checks if the path is within the allowed mods directory.
func (l *Loader) validateSandboxPath(path string) error {
	if !l.config.EnableSandbox {
		return nil
	}
	sandbox := NewSandboxWithConfig(SandboxConfig{
		ModsDirectory: l.config.ModsDirectory,
	})
	return sandbox.ValidatePath(path)
}

// validateFileSize checks if the mod file exceeds the size limit.
func (l *Loader) validateFileSize(path string, data []byte) error {
	if l.config.EnableSandbox && len(data) > 1024*1024 {
		return &LoadError{
			ModID: path,
			Err:   fmt.Errorf("mod file exceeds 1MB size limit"),
		}
	}
	return nil
}

// parseModJSON unmarshals JSON data into a Mod structure.
func (l *Loader) parseModJSON(path string, data []byte) (*Mod, error) {
	var mod Mod
	if err := json.Unmarshal(data, &mod); err != nil {
		return nil, &LoadError{ModID: path, Err: fmt.Errorf("invalid JSON: %w", err)}
	}

	mod.LoadedAt = time.Now()

	if err := mod.Validate(); err != nil {
		return nil, &LoadError{ModID: mod.ID, Err: err}
	}

	return &mod, nil
}

// validateModContent validates the mod against sandbox security rules.
func (l *Loader) validateModContent(mod *Mod) error {
	if !l.config.EnableSandbox {
		return nil
	}

	sandbox := NewSandbox()
	result := sandbox.ValidateMod(mod)
	if !result.Valid {
		errMsgs := buildSandboxErrorMessages(result.Errors)
		return &LoadError{ModID: mod.ID, Err: fmt.Errorf("sandbox validation failed: %s", errMsgs)}
	}
	return nil
}

// buildSandboxErrorMessages concatenates SandboxError messages.
func buildSandboxErrorMessages(errors []SandboxError) string {
	errMsgs := make([]string, 0, len(errors))
	for _, e := range errors {
		errMsgs = append(errMsgs, e.Error())
	}
	errMsg := strings.Join(errMsgs, "; ")
	if len(errMsgs) > 0 {
		errMsg += "; "
	}
	return errMsg
}

// LoadAll loads all mods from the mods directory.
func (l *Loader) LoadAll() ([]*Mod, error) {
	if err := l.ensureModsDirectoryExists(); err != nil {
		return nil, err
	}

	entries, err := l.readModsDirectory()
	if err != nil {
		return nil, err
	}

	return l.loadModsFromEntries(entries)
}

// ensureModsDirectoryExists creates the mods directory if it doesn't exist.
func (l *Loader) ensureModsDirectoryExists() error {
	if _, err := os.Stat(l.config.ModsDirectory); os.IsNotExist(err) {
		if err := os.MkdirAll(l.config.ModsDirectory, 0o755); err != nil {
			return fmt.Errorf("failed to create mods directory: %w", err)
		}
	}
	return nil
}

// readModsDirectory reads entries from the mods directory.
func (l *Loader) readModsDirectory() ([]os.DirEntry, error) {
	entries, err := os.ReadDir(l.config.ModsDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to read mods directory: %w", err)
	}
	return entries, nil
}

// loadModsFromEntries loads mods from directory entries up to the max mods limit.
func (l *Loader) loadModsFromEntries(entries []os.DirEntry) ([]*Mod, error) {
	var mods []*Mod
	var loadErrors []error

	for _, entry := range entries {
		if !l.isValidModFile(entry) {
			continue
		}

		mod, err := l.loadModFromEntry(entry)
		if err != nil {
			loadErrors = append(loadErrors, err)
			continue
		}

		mods = append(mods, mod)

		if l.reachedMaxModsLimit(mods) {
			break
		}
	}

	return l.validateLoadResults(mods, loadErrors)
}

// isValidModFile checks if the directory entry is a valid mod JSON file.
func (l *Loader) isValidModFile(entry os.DirEntry) bool {
	return !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json")
}

// loadModFromEntry loads a mod from a directory entry.
func (l *Loader) loadModFromEntry(entry os.DirEntry) (*Mod, error) {
	path := filepath.Join(l.config.ModsDirectory, entry.Name())
	return l.LoadFromFile(path)
}

// reachedMaxModsLimit checks if the max mods limit has been reached.
func (l *Loader) reachedMaxModsLimit(mods []*Mod) bool {
	return len(mods) >= l.config.MaxMods
}

// validateLoadResults validates the load results and returns appropriate error.
func (l *Loader) validateLoadResults(mods []*Mod, loadErrors []error) ([]*Mod, error) {
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
