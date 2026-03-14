//go:build !js
// +build !js

// Package saveload provides migration hooks for save file version upgrades.
// This file defines the Migrator interface and default implementation.
package saveload

// MigrationHook is a function that modifies a GameSave during migration.
type MigrationHook func(save *GameSave, sourceVersion, targetVersion string) error

// DefaultMigrator provides basic migration functionality for common version upgrades.
type DefaultMigrator struct {
	// hooks contains version-specific migration functions
	hooks map[string][]MigrationHook
}

// NewDefaultMigrator creates a migrator with built-in migration rules.
func NewDefaultMigrator() *DefaultMigrator {
	m := &DefaultMigrator{
		hooks: make(map[string][]MigrationHook),
	}
	m.registerDefaultHooks()
	return m
}

// CanMigrate returns true if the migrator supports the given source version.
func (m *DefaultMigrator) CanMigrate(sourceVersion string) bool {
	for _, supported := range m.SupportedVersions() {
		if supported == sourceVersion {
			return true
		}
	}
	return false
}

// SupportedVersions returns all versions that can be migrated to current.
func (m *DefaultMigrator) SupportedVersions() []string {
	return []string{"0.9.0", "0.9.1", "0.9.2", "0.9.3"}
}

// Migrate transforms a save from an older version to the current SaveVersion.
func (m *DefaultMigrator) Migrate(save *GameSave, sourceVersion string) (*GameSave, error) {
	if save == nil {
		return nil, ErrNilSave
	}

	if !m.CanMigrate(sourceVersion) {
		return nil, &MigrationError{
			SourceVersion: sourceVersion,
			TargetVersion: SaveVersion,
			Message:       "unsupported source version",
		}
	}

	// Apply version-specific hooks
	if hooks, exists := m.hooks[sourceVersion]; exists {
		for _, hook := range hooks {
			if err := hook(save, sourceVersion, SaveVersion); err != nil {
				return nil, &MigrationError{
					SourceVersion: sourceVersion,
					TargetVersion: SaveVersion,
					Message:       err.Error(),
				}
			}
		}
	}

	// Apply default migrations that apply to all versions
	if err := m.applyDefaultMigrations(save, sourceVersion); err != nil {
		return nil, err
	}

	// Update version to current
	save.Version = SaveVersion

	return save, nil
}

// RegisterHook adds a migration hook for a specific source version.
func (m *DefaultMigrator) RegisterHook(sourceVersion string, hook MigrationHook) {
	if m.hooks == nil {
		m.hooks = make(map[string][]MigrationHook)
	}
	m.hooks[sourceVersion] = append(m.hooks[sourceVersion], hook)
}

// registerDefaultHooks sets up built-in migration transformations.
func (m *DefaultMigrator) registerDefaultHooks() {
	// Example: Migrate from 0.9.0 - ensure new fields have defaults
	m.RegisterHook("0.9.0", func(save *GameSave, source, target string) error {
		// Ensure PlayerState has new fields initialized
		if save.PlayerState != nil {
			if save.PlayerState.TrustScores == nil {
				save.PlayerState.TrustScores = make(map[string]float64)
			}
			if save.PlayerState.ReputationScores == nil {
				save.PlayerState.ReputationScores = make(map[string]int)
			}
		}
		return nil
	})

	// 0.9.1 migration
	m.RegisterHook("0.9.1", func(save *GameSave, source, target string) error {
		if save.PlayerState != nil {
			if save.PlayerState.TrustScores == nil {
				save.PlayerState.TrustScores = make(map[string]float64)
			}
			if save.PlayerState.ReputationScores == nil {
				save.PlayerState.ReputationScores = make(map[string]int)
			}
		}
		return nil
	})

	// 0.9.2 migration
	m.RegisterHook("0.9.2", func(save *GameSave, source, target string) error {
		if save.PlayerState != nil {
			if save.PlayerState.TrustScores == nil {
				save.PlayerState.TrustScores = make(map[string]float64)
			}
		}
		return nil
	})

	// 0.9.3 migration - minimal changes
	m.RegisterHook("0.9.3", func(save *GameSave, source, target string) error {
		return nil
	})
}

// applyDefaultMigrations applies migrations common to all version upgrades.
func (m *DefaultMigrator) applyDefaultMigrations(save *GameSave, sourceVersion string) error {
	// Ensure required fields exist
	if save.PlayerState == nil {
		save.PlayerState = &PlayerState{
			Items: make([]ItemData, 0),
		}
	}
	if save.WorldState == nil {
		save.WorldState = &WorldState{
			ModifiedEntities: make([]ModifiedEntity, 0),
		}
	}
	if save.Settings == nil {
		save.Settings = &GameSettings{
			ScreenWidth:  800,
			ScreenHeight: 600,
			VSync:        true,
			MasterVolume: 1.0,
			MusicVolume:  0.7,
			SFXVolume:    0.8,
			KeyBindings:  make(map[string]string),
		}
	}

	// Ensure slice fields are initialized
	if save.PlayerState.Items == nil {
		save.PlayerState.Items = make([]ItemData, 0)
	}
	if save.WorldState.ModifiedEntities == nil {
		save.WorldState.ModifiedEntities = make([]ModifiedEntity, 0)
	}

	return nil
}

// MigrationError represents a failure during save migration.
type MigrationError struct {
	SourceVersion string
	TargetVersion string
	Message       string
}

func (e *MigrationError) Error() string {
	return "migration failed from " + e.SourceVersion + " to " + e.TargetVersion + ": " + e.Message
}

// ErrNilSave is returned when attempting to migrate a nil save.
var ErrNilSave = &MigrationError{Message: "save cannot be nil"}
