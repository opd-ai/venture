// Package migration provides backward compatibility validation for save file migrations.
//
// # Overview
//
// The migration package validates that save files from previous versions (0.9.0 through 0.9.3)
// can be successfully loaded and migrated to 1.0.0 format. This ensures players can continue
// their progress across version updates.
//
// # Features
//
// - Automated migration testing for all version pairs (0.9.0→1.0.0, 0.9.1→1.0.0, etc.)
// - Save file format validation and compatibility checks
// - Component migration verification (ensures all components preserved)
// - Data integrity validation (required fields, nested types)
// - Performance benchmarking for migration operations
//
// # Example Usage
//
//	// Validate all migrations to 1.0.0
//	validator := migration.NewValidator(migration.Config{
//	    TargetVersion: "1.0.0",
//	    TestDataPath:  "testdata/saves/",
//	})
//
//	results, err := validator.ValidateAll()
//	if err != nil {
//	    log.Fatalf("Migration validation failed: %v", err)
//	}
//
//	fmt.Printf("Migrations passed: %d/%d\n", results.PassedCount, results.TotalCount)
//
// # Version Support
//
// This package validates migrations from the following source versions to 1.0.0:
//
//   - 0.9.0: Requires TrustScores and ReputationScores initialization
//   - 0.9.1: Requires TrustScores and ReputationScores initialization
//   - 0.9.2: Requires TrustScores initialization
//   - 0.9.3: Minimal changes, close to 1.0.0 format
//
// The migration rules mirror pkg/saveload.DefaultMigrator.registerDefaultHooks().
//
// # Testing
//
// Run migration validation tests:
//
//	go test -v ./pkg/migration/...
//
// This package provides programmatic migration validation used by the game's
// save/load system. No standalone CLI tool is provided.
package migration
