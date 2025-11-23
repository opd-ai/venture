// Package migration provides backward compatibility validation for save file migrations.
//
// # Overview
//
// The migration package validates that save files from all previous versions (v1.0 through v9.0)
// can be successfully loaded and migrated to v10.0 format. This ensures players can continue
// their progress across major version updates.
//
// # Features
//
// - Automated migration testing for all version pairs (v1.0→v10.0, v2.0→v10.0, etc.)
// - Save file format validation and compatibility checks
// - Component migration verification (ensures all components preserved)
// - Data integrity validation (checksums, required fields)
// - Performance benchmarking for migration operations
//
// # Example Usage
//
//	// Validate all migrations to v10.0
//	validator := migration.NewValidator(migration.Config{
//	    TargetVersion: "v10.0",
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
// - v1.0: Initial release (Phase 1-14 complete)
// - v2.0: Phase 15-20 visual enhancements
// - v3.0: Phase 21-30 gameplay expansion
// - v4.0: Phase 31-36 social systems
// - v5.0: Phase 37-42 persistent worlds
// - v6.0: Phase 43-48 advanced visuals
// - v7.0: Phase 49-54 performance optimization
// - v8.0: Phase 55-60 content expansion
// - v9.0: Phase 61-66 production readiness
// - v10.0: Current target version
//
// # CLI Tool
//
// See cmd/migrationtest/ for standalone migration validation tool.
package migration
