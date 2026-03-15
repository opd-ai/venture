# Config Package

## Overview

The `config` package provides comprehensive configuration validation for Venture's server and client. It validates command-line flags and environment variables before game startup, ensuring all settings meet requirements and providing helpful error messages for invalid configurations.

## Package Structure

```
pkg/config/
├── doc.go           - Package documentation with usage examples
├── types.go         - Configuration data structures
├── validator.go     - Validation logic and methods
└── validator_test.go - Comprehensive test suite (92.4% coverage)
```

## Core Types

### Validator
The main validation engine for configuration parameters:
```go
type Validator struct {
    // Contains valid genre mappings
}
```

### Config
Configuration container for validation:
```go
type Config struct {
    Port               string  // Server port ("1024"-"65535")
    MaxPlayers         int     // Max concurrent players (1-100)
    ValidateMaxPlayers bool    // Enable MaxPlayers validation
    TickRate           int     // Server tick rate in Hz (1-60)
    ValidateTickRate   bool    // Enable TickRate validation
    Genre              string  // Game genre identifier
    SaveDir            string  // Save files directory
    LogDir             string  // Log files directory
    ModsDir            string  // Mods directory
    CreateDirs         bool    // Create missing directories
}
```

## Usage Examples

### Basic Validation

```go
import "github.com/opd-ai/venture/pkg/config"

// Create validator
validator := config.NewValidator()

// Validate individual fields
if err := validator.ValidatePort("8080"); err != nil {
    log.Fatalf("Invalid port: %v", err)
}

if err := validator.ValidateGenre("fantasy"); err != nil {
    log.Fatalf("Invalid genre: %v", err)
}

if err := validator.ValidateMaxPlayers(50); err != nil {
    log.Fatalf("Invalid max players: %v", err)
}
```

### Complete Configuration Validation

```go
// Prepare configuration
cfg := &config.Config{
    Port:               "8080",
    MaxPlayers:         10,
    ValidateMaxPlayers: true,
    TickRate:           30,
    ValidateTickRate:   true,
    Genre:              "fantasy",
    SaveDir:            "./saves",
    LogDir:             "./logs",
    ModsDir:            "./mods",
    CreateDirs:         true, // Create directories if missing
}

// Validate all settings
validator := config.NewValidator()
if err := validator.ValidateAll(cfg); err != nil {
    log.Fatalf("Configuration validation failed: %v", err)
}
```

### Get Available Genres

```go
validator := config.NewValidator()
genres := validator.GetAvailableGenres()
fmt.Printf("Available genres: %s\n", strings.Join(genres, ", "))
// Output: Available genres: cyberpunk, fantasy, horror, postapoc, scifi
```

### Directory Validation

```go
validator := config.NewValidator()

// Check existing directory
if err := validator.ValidateDirectory("/path/to/dir", false); err != nil {
    log.Printf("Directory validation failed: %v", err)
}

// Create directory if it doesn't exist
if err := validator.ValidateDirectory("/path/to/newdir", true); err != nil {
    log.Printf("Failed to create directory: %v", err)
}
```

## Validation Rules

### Port Validation
- **Range**: 1024-65535
- **Reason**: Ports below 1024 require root privileges
- **Format**: String representation of integer

### MaxPlayers Validation
- **Range**: 1-100
- **Reason**: Performance degrades with >100 players
- **Type**: Integer

### TickRate Validation
- **Range**: 1-60 Hz
- **Reason**: Diminishing returns above 60 Hz
- **Type**: Integer

### Genre Validation
- **Valid Values**: Retrieved from procgen/dialog package
- **Common Genres**: fantasy, scifi, horror, cyberpunk, postapoc
- **Error Message**: Lists available genres if invalid

### Directory Validation
- **Checks**: Existence, accessibility, directory type
- **Creation**: Optional via `create` parameter
- **Permissions**: Creates with 0o755 mode

## Error Handling

All validation methods return descriptive errors:

```go
// Port error example
"port must be between 1024 and 65535, got 80 (ports < 1024 require root privileges)"

// Genre error example
"invalid genre 'western', available genres: cyberpunk, fantasy, horror, postapoc, scifi (or 'random')"

// Directory error example
"directory /invalid/path is not accessible: no such file or directory"
```

## Testing

Run package tests:
```bash
go test ./pkg/config/...
```

Current test coverage: **83.9%**

View coverage report:
```bash
go test -coverprofile=coverage.out ./pkg/config/...
go tool cover -html=coverage.out
```

## Implementation Status

✅ **Complete**: All validation methods implemented and tested  
⚠️  **Minor Gaps**: Some error paths untested (see AUDIT.md)

See [AUDIT.md](./AUDIT.md) for detailed analysis.

## Design Decisions

### Separation of Concerns
- **types.go**: Data structures only
- **validator.go**: Validation logic
- **doc.go**: Package-level documentation

### Conditional Validation
The `ValidateMaxPlayers` and `ValidateTickRate` flags allow selective validation, useful when:
- Fields have default values that shouldn't be validated
- Client-only vs server-only configurations
- Optional parameters in configuration

### Directory Creation
The `CreateDirs` flag enables automatic directory creation during validation, simplifying setup:
- **true**: Creates missing directories with 0o755 permissions
- **false**: Returns error if directory doesn't exist

### Genre Dependency
Validator depends on `pkg/procgen/dialog` for genre list to maintain single source of truth. See AUDIT.md for discussion of this design choice.

## Common Patterns

### CLI Flag Validation
```go
func validateFlags() error {
    validator := config.NewValidator()
    
    if *port != "" {
        if err := validator.ValidatePort(*port); err != nil {
            return fmt.Errorf("invalid -port flag: %w", err)
        }
    }
    
    if *genre != "" {
        if err := validator.ValidateGenre(*genre); err != nil {
            return fmt.Errorf("invalid -genre flag: %w", err)
        }
    }
    
    return nil
}
```

### Configuration File Validation
```go
func validateConfigFile(path string) error {
    // Load config from file
    cfg, err := loadConfig(path)
    if err != nil {
        return err
    }
    
    // Validate
    validator := config.NewValidator()
    return validator.ValidateAll(cfg)
}
```

## Related Packages

- `pkg/procgen/dialog` - Provides available genre list (coupling is intentional: genre consistency guaranteed by single source of truth)
- `cmd/server` - Uses validator for server configuration
- `cmd/client` - Uses validator for client configuration

## Benchmarks

The validator package includes benchmarks for performance-critical paths:

| Benchmark | Typical ns/op |
|-----------|--------------|
| `BenchmarkValidatePort` | validates port range |
| `BenchmarkValidateMaxPlayers` | validates player count |
| `BenchmarkValidateTickRate` | validates tick rate |
| `BenchmarkValidateGenre` | validates genre string |
| `BenchmarkValidateAll` | full config validation |
| `BenchmarkNewValidator` | validator construction |

Run with: `go test -bench=. ./pkg/config/`

## Future Enhancements

See AUDIT.md for recommendations including:
- Additional test coverage for error paths
- Validation constants for magic numbers
- Alternative genre dependency approaches
