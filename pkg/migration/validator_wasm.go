//go:build js
// +build js

// Package migration provides backward compatibility validation for save file migrations.
// WASM build: Migration validation is disabled because saveload.NewDefaultMigrator is not available.
package migration

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/saveload"
	"github.com/sirupsen/logrus"
)

// wasmMigrator is a no-op migrator for WASM builds where migration is not supported.
type wasmMigrator struct{}

// CanMigrate always returns false on WASM (migration not supported).
func (m *wasmMigrator) CanMigrate(sourceVersion string) bool {
	return false
}

// Migrate always returns an error on WASM.
func (m *wasmMigrator) Migrate(save *saveload.GameSave, sourceVersion string) (*saveload.GameSave, error) {
	return nil, fmt.Errorf("save migration not supported on WASM builds")
}

// SupportedVersions returns empty slice on WASM.
func (m *wasmMigrator) SupportedVersions() []string {
	return []string{}
}

// Validator performs save file migration validation.
// WASM build: Uses no-op migrator.
type Validator struct {
	config   Config
	migrator saveload.Migrator
	logger   *logrus.Logger
}

// NewValidator creates a migration validator with the given configuration.
// WASM build: Uses wasmMigrator which always returns errors.
func NewValidator(config Config) *Validator {
	if config.TargetVersion == "" {
		config.TargetVersion = "1.0.0"
	}
	if config.TestDataPath == "" {
		config.TestDataPath = "testdata/saves/"
	}

	return &Validator{
		config:   config,
		migrator: &wasmMigrator{},
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

// ValidateAll always returns an error on WASM builds.
// Migration validation is not supported on WASM.
func (v *Validator) ValidateAll() (*ValidationResults, error) {
	v.logger.Warn("Migration validation not supported on WASM builds")
	return &ValidationResults{
		Migrations:  []MigrationResult{},
		TotalCount:  0,
		PassedCount: 0,
		FailedCount: 0,
	}, fmt.Errorf("migration validation not supported on WASM builds")
}

// ValidateMigration always returns a failed result on WASM builds.
// Migration validation is not supported on WASM.
func (v *Validator) ValidateMigration(source, target string) MigrationResult {
	v.logger.WithFields(logrus.Fields{
		"source_version": source,
		"target_version": target,
	}).Warn("Migration validation not supported on WASM builds")

	return MigrationResult{
		SourceVersion: source,
		TargetVersion: target,
		Passed:        false,
		Error:         "migration validation not supported on WASM builds",
	}
}
