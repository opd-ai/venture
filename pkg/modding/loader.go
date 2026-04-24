package modding

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	logrus "github.com/sirupsen/logrus"
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
	logrus.WithFields(logrus.Fields{
		"path": path,
	}).Debug("loading mod from file")

	if err := l.validateSandboxPath(path); err != nil {
		return nil, &LoadError{ModID: path, Err: err}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"path":  path,
			"error": err,
		}).Error("failed to read mod file")
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

	logrus.WithFields(logrus.Fields{
		"mod_id":   mod.ID,
		"mod_name": mod.Name,
		"version":  mod.Version,
	}).Info("mod loaded successfully")

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

	// INTENTIONAL time.Now() EXCEPTION: LoadedAt is metadata for debugging/audit only.
	// Does not affect procedural content generation. See doc.go:113-120.
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
	logrus.WithFields(logrus.Fields{
		"directory": l.config.ModsDirectory,
	}).Debug("loading all mods from directory")

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
			logrus.WithFields(logrus.Fields{
				"file":  entry.Name(),
				"error": err,
			}).Warn("failed to load mod file")
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
	logrus.WithFields(logrus.Fields{
		"mod_id": mod.ID,
		"path":   path,
	}).Debug("saving mod to file")

	if err := mod.Validate(); err != nil {
		return &LoadError{ModID: mod.ID, Err: err}
	}

	if err := l.validateSavePermissions(mod, path); err != nil {
		return err
	}

	data, err := l.marshalModToJSON(mod)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		logrus.WithFields(logrus.Fields{
			"mod_id": mod.ID,
			"path":   path,
			"error":  err,
		}).Error("failed to write mod file")
		return &LoadError{ModID: mod.ID, Err: err}
	}

	logrus.WithFields(logrus.Fields{
		"mod_id": mod.ID,
		"path":   path,
	}).Info("mod saved successfully")

	return nil
}

// validateSavePermissions validates the save path and mod content against sandbox rules.
func (l *Loader) validateSavePermissions(mod *Mod, path string) error {
	if !l.config.EnableSandbox {
		return nil
	}

	sandbox := NewSandboxWithConfig(SandboxConfig{
		ModsDirectory: l.config.ModsDirectory,
	})

	if err := sandbox.ValidatePath(path); err != nil {
		return &LoadError{ModID: mod.ID, Err: err}
	}

	result := sandbox.ValidateMod(mod)
	if !result.Valid {
		errMsg := buildSandboxErrorMessages(result.Errors)
		return &LoadError{ModID: mod.ID, Err: fmt.Errorf("sandbox validation failed: %s", errMsg)}
	}
	return nil
}

// marshalModToJSON converts a mod to formatted JSON bytes.
func (l *Loader) marshalModToJSON(mod *Mod) ([]byte, error) {
	data, err := json.MarshalIndent(mod, "", "  ")
	if err != nil {
		return nil, &LoadError{ModID: mod.ID, Err: fmt.Errorf("failed to marshal JSON: %w", err)}
	}
	return data, nil
}

// GetModPath returns the file path for a mod by ID.
func (l *Loader) GetModPath(modID string) string {
	return filepath.Join(l.config.ModsDirectory, modID+".json")
}

// ParseModFromBytes unmarshals raw JSON bytes into a Mod, validates it, and
// returns it ready for use with Manager.AddMod.  It is a package-level
// convenience wrapper around Loader.parseModJSON without file-system access,
// intended for use by the ModBrowserSystem install callback.
func ParseModFromBytes(data []byte) (*Mod, error) {
l := NewLoader()
return l.parseModJSON("<bytes>", data)
}
