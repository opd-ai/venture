# Migration Package

## Overview

The `migration` package provides backward compatibility validation for save file migrations across Venture game versions. It ensures players can continue their progress when upgrading from version 0.9.x to 1.0.0.

## Package Structure

```
pkg/migration/
├── doc.go            - Comprehensive package documentation with examples
├── types.go          - Configuration and result types
├── validator.go      - Migration validation logic
└── validator_test.go - Test suite (91.3% coverage)
```

## Core Types

### Config
Configuration for migration validation:
```go
type Config struct {
    TargetVersion string  // Version to migrate to (e.g., "1.0.0")
    TestDataPath  string  // Directory with test save files
    ValidateData  bool    // Enable deep data integrity checks
}
```

### Validator
Main validation engine:
```go
type Validator struct {
    // Configuration for validation process
}
```

### ValidationResults
Aggregated results from testing multiple migrations:
```go
type ValidationResults struct {
    TotalCount   int
    PassedCount  int
    FailedCount  int
    Migrations   []MigrationResult
}
```

### MigrationResult
Result from a single version-to-version migration test:
```go
type MigrationResult struct {
    SourceVersion       string
    TargetVersion       string
    Passed              bool
    Error               string
    ComponentsPreserved []string
    MigrationTime       float64  // seconds
}
```

## Usage

### Basic Migration Validation

```go
import "github.com/opd-ai/venture/pkg/migration"

// Create validator
validator := migration.NewValidator(migration.Config{
    TargetVersion: "1.0.0",
    TestDataPath:  "testdata/saves/",
    ValidateData:  true,
})

// Validate all supported migrations
results, err := validator.ValidateAll()
if err != nil {
    log.Fatalf("Validation failed: %v", err)
}

// Check results
fmt.Printf("Migrations tested: %d\n", results.TotalCount)
fmt.Printf("Passed: %d, Failed: %d\n", 
    results.PassedCount, results.FailedCount)

for _, migration := range results.Migrations {
    if migration.Passed {
        fmt.Printf("✅ %s → %s (%.3fs)\n", 
            migration.SourceVersion, 
            migration.TargetVersion,
            migration.MigrationTime)
    } else {
        fmt.Printf("❌ %s → %s: %s\n",
            migration.SourceVersion,
            migration.TargetVersion, 
            migration.Error)
    }
}
```

### Single Version Migration Test

```go
validator := migration.NewValidator(migration.Config{
    TargetVersion: "1.0.0",
    ValidateData:  true,
})

// Test specific version migration
result := validator.ValidateMigration("0.9.2", "1.0.0")

if result.Passed {
    fmt.Printf("Migration successful!\n")
    fmt.Printf("Components preserved: %v\n", result.ComponentsPreserved)
} else {
    fmt.Printf("Migration failed: %s\n", result.Error)
}
```

### With Custom Test Data

```go
validator := migration.NewValidator(migration.Config{
    TargetVersion: "1.0.0",
    TestDataPath:  "/custom/path/to/saves/",
    ValidateData:  true,
})

results, _ := validator.ValidateAll()
```

## Migration Rules

The package validates migrations using rules that mirror `pkg/saveload.DefaultMigrator`:

### 0.9.0 → 1.0.0
- Adds `trust_scores` map to player
- Adds `reputation_scores` map to player

### 0.9.1 → 1.0.0
- Adds `trust_scores` map to player
- Adds `reputation_scores` map to player

### 0.9.2 → 1.0.0
- Adds `trust_scores` map to player

### 0.9.3 → 1.0.0
- Minimal changes (close to 1.0.0 format)

### All Versions (Common Migrations)
- Ensures `player` exists with `items` array
- Ensures `world` exists with `modified_entities` array
- Ensures `settings` exists with default values

## Validation Process

1. **Load Save File**: Reads save file or generates synthetic test data
2. **Apply Migrations**: Applies version-specific transformations
3. **Validate Data**: Checks required fields and data integrity (if enabled)
4. **Record Results**: Captures pass/fail status, timing, preserved components

## Synthetic Save Generation

When real save files don't exist, the validator generates minimal test data:

```go
{
  "version": "0.9.0",
  "player": {
    "name": "TestPlayer",
    "level": 10,
    "health": 100
  },
  "world": {
    "seed": 12345,
    "depth": 5
  },
  "inventory": [
    {"id": 1, "count": 10},
    {"id": 2, "count": 5}
  ]
}
```

## Testing

Run migration tests:
```bash
go test ./pkg/migration/...
```

Test with coverage:
```bash
go test -coverprofile=coverage.out ./pkg/migration/...
go tool cover -html=coverage.out
```

**Current coverage: 91.3%** ✅

## Implementation Status

✅ **Complete**:
- Migration rule validation using real `pkg/saveload.Migrator`
- Synthetic save generation for testing without real save files
- Data integrity checks (required fields, nested types, version matching)
- Component preservation tracking
- Real migration timing measurement via `time.Since()`
- Fallback migration for unsupported versions

## Supported Versions

Validates migrations from these source versions to 1.0.0:
- **0.9.0** - Early beta
- **0.9.1** - Beta update
- **0.9.2** - Pre-release candidate  
- **0.9.3** - Release candidate

Matches `pkg/saveload.DefaultMigrator.SupportedVersions()`.

## Testing

Run migration validation tests:
```bash
go test -v ./pkg/migration/...
```

This package provides programmatic migration validation used by the game's
save/load system. No standalone CLI tool is provided.

## Design Philosophy

### Test Isolation
The package intentionally simulates migration logic rather than importing `pkg/saveload` to:
- Avoid circular dependencies
- Enable independent testing
- Provide fast validation without full game initialization

**Trade-off**: Migration rules must be manually kept in sync with `pkg/saveload`.

### Progressive Validation
- Basic: Version field updates
- Standard: Component preservation
- Deep: Data integrity checks (enabled via Config.ValidateData)

## Common Patterns

### Pre-release Validation
```go
func validateRelease() error {
    validator := migration.NewValidator(migration.Config{
        TargetVersion: "1.0.0",
        ValidateData:  true,
    })
    
    results, err := validator.ValidateAll()
    if err != nil {
        return err
    }
    
    if results.FailedCount > 0 {
        return fmt.Errorf("%d migrations failed", results.FailedCount)
    }
    
    return nil
}
```

### CI/CD Integration
```go
func TestBackwardCompatibility(t *testing.T) {
    validator := migration.NewValidator(migration.Config{
        TargetVersion: currentVersion,
        ValidateData:  true,
    })
    
    results, err := validator.ValidateAll()
    require.NoError(t, err)
    require.Equal(t, 0, results.FailedCount, 
        "All migrations must pass")
}
```

## Error Messages

Validation provides detailed error context:

```
// Missing save file
"failed to load save file: failed to read file: no such file or directory"

// Migration failure
"migration failed: failed to apply migration rules: ..."

// Data validation failure
"validation failed: missing required field: player"
"validation failed: version mismatch: expected 1.0.0, got 0.9.0"
```

## Related Packages

- `pkg/saveload` - Actual save/load and migration implementation
- `cmd/migrationtest` - CLI tool for migration validation

## Future Enhancements

See AUDIT.md for recommendations:
- Integration with real saveload migrations
- Expanded synthetic save generation
- Actual migration timing measurement
- Automated sync verification with pkg/saveload
