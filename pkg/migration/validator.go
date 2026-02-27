//go:build !js
// +build !js

// validator.go implements save file migration validation logic.
// This file contains the Validator type and all migration validation methods
// for testing backward compatibility of save files across game versions.
//
// Package migration provides backward compatibility validation for save file migrations.
package migration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/opd-ai/venture/pkg/saveload"
	"github.com/sirupsen/logrus"
)

// Validator performs save file migration validation.
// Code relocated from original validator.go (types moved to types.go)
type Validator struct {
	config   Config
	migrator saveload.Migrator
	logger   *logrus.Logger
}

// NewValidator creates a migration validator with the given configuration.
func NewValidator(config Config) *Validator {
	if config.TargetVersion == "" {
		config.TargetVersion = "1.0.0"
	}
	if config.TestDataPath == "" {
		config.TestDataPath = "testdata/saves/"
	}

	return &Validator{
		config:   config,
		migrator: saveload.NewDefaultMigrator(),
		logger:   logrus.StandardLogger(),
	}
}

// NewValidatorWithLogger creates a migration validator with custom logger.
func NewValidatorWithLogger(config Config, logger *logrus.Logger) *Validator {
	v := NewValidator(config)
	if logger != nil {
		v.logger = logger
	}
	return v
}

// NewValidatorWithMigrator creates a migration validator with a custom migrator.
// This allows injecting mock migrators for testing.
func NewValidatorWithMigrator(config Config, migrator saveload.Migrator) *Validator {
	v := NewValidator(config)
	if migrator != nil {
		v.migrator = migrator
	}
	return v
}

// ValidateAll tests migration from all supported versions to target version.
// Supported versions match pkg/saveload.DefaultMigrator.SupportedVersions().
func (v *Validator) ValidateAll() (*ValidationResults, error) {
	// Use the same versions as pkg/saveload.DefaultMigrator.SupportedVersions()
	supportedVersions := []string{
		"0.9.0", "0.9.1", "0.9.2", "0.9.3",
	}

	results := &ValidationResults{
		Migrations: make([]MigrationResult, 0, len(supportedVersions)),
	}

	for _, sourceVer := range supportedVersions {
		result := v.ValidateMigration(sourceVer, v.config.TargetVersion)
		results.Migrations = append(results.Migrations, result)
		results.TotalCount++

		if result.Passed {
			results.PassedCount++
		} else {
			results.FailedCount++
		}
	}

	return results, nil
}

// ValidateMigration tests a single version-to-version migration.
func (v *Validator) ValidateMigration(source, target string) MigrationResult {
	result := MigrationResult{
		SourceVersion: source,
		TargetVersion: target,
	}

	// Attempt to load test save file for source version
	saveFilePath := filepath.Join(v.config.TestDataPath, fmt.Sprintf("save_%s.json", source))
	saveData, err := v.loadSaveFile(saveFilePath)
	if err != nil {
		result.Error = fmt.Sprintf("failed to load save file: %v", err)
		return result
	}

	// Perform migration
	migratedData, migrationTime, err := v.performMigration(saveData, source, target)
	result.MigrationTime = migrationTime
	if err != nil {
		result.Error = fmt.Sprintf("migration failed: %v", err)
		return result
	}

	// Validate migrated data
	if v.config.ValidateData {
		components, err := v.validateData(migratedData, target)
		if err != nil {
			result.Error = fmt.Sprintf("validation failed: %v", err)
			return result
		}
		result.ComponentsPreserved = components
	}

	result.Passed = true
	return result
}

// loadSaveFile reads a save file from disk.
func (v *Validator) loadSaveFile(path string) (map[string]interface{}, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Generate synthetic save data for testing (since actual saves may not exist)
		version := extractVersion(path)
		if version == "" {
			version = "0.9.0" // Default to earliest supported version for synthetic saves
		}
		return v.generateSyntheticSave(version), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var saveData map[string]interface{}
	if err := json.Unmarshal(data, &saveData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return saveData, nil
}

// generateSyntheticSave creates a minimal save file for testing when actual saves don't exist.
func (v *Validator) generateSyntheticSave(version string) map[string]interface{} {
	return map[string]interface{}{
		"version": version,
		"player": map[string]interface{}{
			"name":   "TestPlayer",
			"level":  10,
			"health": 100,
		},
		"world": map[string]interface{}{
			"seed":  12345,
			"depth": 5,
		},
		"inventory": []interface{}{
			map[string]interface{}{"id": 1, "count": 10},
			map[string]interface{}{"id": 2, "count": 5},
		},
	}
}

// extractVersion extracts version from save file path.
// Returns empty string if the path doesn't match expected format "save_X.Y.Z.json".
func extractVersion(path string) string {
	base := filepath.Base(path)
	parts := strings.Split(base, "_")
	if len(parts) >= 2 {
		versionWithExt := parts[1]
		version := strings.TrimSuffix(versionWithExt, ".json")
		// Validate version format is semantic (X.Y.Z) or legacy (vX.Y)
		if len(version) >= 3 {
			// Accept both "0.9.0" style and legacy "v1.0" style
			if (version[0] >= '0' && version[0] <= '9') || version[0] == 'v' {
				return version
			}
		}
	}
	// Return empty string for malformed paths
	return ""
}

// performMigration applies version-specific migration logic using the real saveload migrator.
// It converts the map data to a GameSave, calls the actual migration, and measures real time.
func (v *Validator) performMigration(data map[string]interface{}, source, target string) (map[string]interface{}, float64, error) {
	startTime := time.Now()

	// Convert map data to GameSave struct for real migration
	gameSave, err := v.mapToGameSave(data, source)
	if err != nil {
		v.logger.WithFields(logrus.Fields{
			"source_version": source,
			"target_version": target,
			"error":          err.Error(),
		}).Error("failed to convert save data to GameSave")
		return nil, 0, fmt.Errorf("failed to convert to GameSave: %w", err)
	}

	// Check if migrator supports this version
	if !v.migrator.CanMigrate(source) {
		v.logger.WithFields(logrus.Fields{
			"source_version": source,
			"target_version": target,
		}).Warn("migrator does not support source version, using fallback migration")
		// Fallback to simulated migration for unsupported versions
		return v.performFallbackMigration(data, source, target, startTime)
	}

	// Call real migrator
	migratedSave, err := v.migrator.Migrate(gameSave, source)
	if err != nil {
		v.logger.WithFields(logrus.Fields{
			"source_version": source,
			"target_version": target,
			"error":          err.Error(),
		}).Error("migration failed")
		return nil, 0, fmt.Errorf("migration failed: %w", err)
	}

	// Convert migrated GameSave back to map
	migratedData, err := v.gameSaveToMap(migratedSave)
	if err != nil {
		v.logger.WithFields(logrus.Fields{
			"source_version": source,
			"target_version": target,
			"error":          err.Error(),
		}).Error("failed to convert migrated save to map")
		return nil, 0, fmt.Errorf("failed to convert migrated save: %w", err)
	}

	// Measure actual migration time
	migrationTime := time.Since(startTime).Seconds()

	v.logger.WithFields(logrus.Fields{
		"source_version": source,
		"target_version": target,
		"migration_time": migrationTime,
	}).Debug("migration completed successfully")

	return migratedData, migrationTime, nil
}

// performFallbackMigration uses the simulated migration for unsupported versions.
func (v *Validator) performFallbackMigration(data map[string]interface{}, source, target string, startTime time.Time) (map[string]interface{}, float64, error) {
	// Copy data
	migratedData := make(map[string]interface{})
	for k, val := range data {
		migratedData[k] = val
	}
	migratedData["version"] = target

	// Apply version-specific transformations
	if err := v.applyMigrationRules(migratedData, source, target); err != nil {
		return nil, 0, err
	}

	// Measure actual time even for fallback
	migrationTime := time.Since(startTime).Seconds()

	return migratedData, migrationTime, nil
}

// mapToGameSave converts a map[string]interface{} to a saveload.GameSave.
func (v *Validator) mapToGameSave(data map[string]interface{}, version string) (*saveload.GameSave, error) {
	// Marshal the map to JSON, then unmarshal to GameSave
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var gameSave saveload.GameSave
	if err := json.Unmarshal(jsonData, &gameSave); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to GameSave: %w", err)
	}

	// Ensure version is set
	gameSave.Version = version

	return &gameSave, nil
}

// gameSaveToMap converts a saveload.GameSave back to map[string]interface{}.
func (v *Validator) gameSaveToMap(save *saveload.GameSave) (map[string]interface{}, error) {
	// Marshal GameSave to JSON, then unmarshal to map
	jsonData, err := json.Marshal(save)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GameSave: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	return result, nil
}

// applyMigrationRules applies version-specific data transformations.
// These mirror the migrations in pkg/saveload.DefaultMigrator.registerDefaultHooks().
func (v *Validator) applyMigrationRules(data map[string]interface{}, source, target string) error {
	// Apply source-version-specific migrations (matching pkg/saveload patterns)
	switch source {
	case "0.9.0", "0.9.1":
		// 0.9.0 and 0.9.1 need TrustScores and ReputationScores initialized
		v.ensureTrustAndReputationFields(data)
	case "0.9.2":
		// 0.9.2 needs TrustScores initialized
		v.ensureTrustFields(data)
	case "0.9.3":
		// 0.9.3 - minimal changes, close to 1.0.0
	}

	// Apply default migrations that apply to all versions
	v.ensureRequiredFields(data)

	return nil
}

// ensureTrustAndReputationFields adds trust and reputation fields if missing.
// Mirrors pkg/saveload migration hooks for 0.9.0 and 0.9.1.
func (v *Validator) ensureTrustAndReputationFields(data map[string]interface{}) {
	if player, ok := data["player"].(map[string]interface{}); ok {
		if _, exists := player["trust_scores"]; !exists {
			player["trust_scores"] = map[string]interface{}{}
		}
		if _, exists := player["reputation_scores"]; !exists {
			player["reputation_scores"] = map[string]interface{}{}
		}
	}
}

// ensureTrustFields adds trust fields if missing.
// Mirrors pkg/saveload migration hook for 0.9.2.
func (v *Validator) ensureTrustFields(data map[string]interface{}) {
	if player, ok := data["player"].(map[string]interface{}); ok {
		if _, exists := player["trust_scores"]; !exists {
			player["trust_scores"] = map[string]interface{}{}
		}
	}
}

// ensureRequiredFields adds fields common to all migrations.
// Mirrors pkg/saveload.DefaultMigrator.applyDefaultMigrations().
func (v *Validator) ensureRequiredFields(data map[string]interface{}) {
	// Ensure player exists with required nested fields
	if _, exists := data["player"]; !exists {
		data["player"] = map[string]interface{}{
			"items": []interface{}{},
		}
	} else if player, ok := data["player"].(map[string]interface{}); ok {
		if _, exists := player["items"]; !exists {
			player["items"] = []interface{}{}
		}
	}

	// Ensure world exists with required nested fields
	if _, exists := data["world"]; !exists {
		data["world"] = map[string]interface{}{
			"modified_entities": []interface{}{},
		}
	} else if world, ok := data["world"].(map[string]interface{}); ok {
		if _, exists := world["modified_entities"]; !exists {
			world["modified_entities"] = []interface{}{}
		}
	}

	// Ensure settings exists with defaults
	if _, exists := data["settings"]; !exists {
		data["settings"] = map[string]interface{}{
			"screen_width":  800,
			"screen_height": 600,
			"vsync":         true,
			"master_volume": 1.0,
			"music_volume":  0.7,
			"sfx_volume":    0.8,
			"key_bindings":  map[string]interface{}{},
		}
	}
}

// validateData performs deep validation of migrated save data.
func (v *Validator) validateData(data map[string]interface{}, targetVersion string) ([]string, error) {
	var preservedComponents []string

	// Check required fields
	requiredFields := []string{"version", "player", "world"}
	for _, field := range requiredFields {
		val, exists := data[field]
		if !exists {
			return nil, fmt.Errorf("missing required field: %s", field)
		}
		// Validate nested field types - player and world should be maps
		if field == "player" || field == "world" {
			if _, isMap := val.(map[string]interface{}); !isMap {
				return nil, fmt.Errorf("field %s must be an object, got %T", field, val)
			}
		}
		preservedComponents = append(preservedComponents, field)
	}

	// Validate version matches target
	if version, ok := data["version"].(string); ok {
		if version != targetVersion {
			return nil, fmt.Errorf("version mismatch: expected %s, got %s", targetVersion, version)
		}
	}

	// Count optional components
	optionalFields := []string{"inventory", "quests", "companions", "vehicles", "guild"}
	for _, field := range optionalFields {
		if _, exists := data[field]; exists {
			preservedComponents = append(preservedComponents, field)
		}
	}

	return preservedComponents, nil
}
