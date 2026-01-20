// types.go defines migration validation data structures.
// This file contains configuration types and result types used for
// save file migration validation.
//
// Package migration provides backward compatibility validation for save file migrations.
package migration

// Config defines migration validation parameters.
// Originally from: validator.go
type Config struct {
	// TargetVersion is the version to migrate to (e.g., "1.0.0")
	TargetVersion string
	// TestDataPath is the directory containing test save files
	TestDataPath string
	// ValidateData enables deep data integrity checks (slower)
	ValidateData bool
}

// ValidationResults contains results from migration validation.
// Originally from: validator.go
type ValidationResults struct {
	// TotalCount is the number of migrations tested
	TotalCount int
	// PassedCount is the number of successful migrations
	PassedCount int
	// FailedCount is the number of failed migrations
	FailedCount int
	// Migrations contains details for each tested migration
	Migrations []MigrationResult
}

// MigrationResult represents a single version migration test.
// Originally from: validator.go
type MigrationResult struct {
	// SourceVersion is the original save file version
	SourceVersion string
	// TargetVersion is the migrated version
	TargetVersion string
	// Passed indicates whether migration succeeded
	Passed bool
	// Error contains error message if migration failed
	Error string
	// ComponentsPreserved lists components that survived migration
	ComponentsPreserved []string
	// MigrationTime is the duration of migration operation
	MigrationTime float64 // seconds
}
