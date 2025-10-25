# Logging Package

The `logging` package provides centralized structured logging configuration for Venture using [logrus](https://github.com/sirupsen/logrus).

## Features

- **Structured logging** with contextual fields
- **Environment-based configuration** via `LOG_LEVEL` and `LOG_FORMAT`
- **Multiple formatters**: JSON (production) and Text (development)
- **Performance-aware**: Conditional debug logging for hot paths
- **Helper functions** for common logging contexts

## Usage

### Basic Setup

```go
import "github.com/opd-ai/venture/pkg/logging"

// Create logger from environment variables
logger := logging.NewLoggerFromEnv()

// Or create with explicit config
logger := logging.NewLogger(logging.Config{
    Level:       logging.InfoLevel,
    Format:      logging.JSONFormat,
    AddCaller:   true,
    EnableColor: false,
})
```

### Structured Logging

Use fields to add context to log entries:

```go
logger.WithFields(logrus.Fields{
    "entityID": 12345,
    "componentType": "position",
}).Info("component added to entity")
```

### Context Helpers

The package provides helpers for common contexts:

```go
// System context
sysLogger := logging.SystemLogger(logger, "terrain")
sysLogger.Info("generating terrain")

// Entity context
entLogger := logging.EntityLogger(logger, entityID)
entLogger.Debug("entity created")

// Generator context
genLogger := logging.GeneratorLogger(logger, "terrain", seed, "fantasy")
genLogger.WithField("depth", 5).Info("generation complete")

// Network context
netLogger := logging.NetworkLogger(logger, playerID, "connected")
netLogger.Info("player joined")
```

## Log Levels

Use appropriate levels consistently:

- **Debug**: Detailed system state, component creation, seed values, generation parameters
- **Info**: Startup/shutdown, phase transitions, world generation, player connections
- **Warn**: Validation failures, retries, performance degradation, high latency
- **Error**: Generation failures, network errors, invalid state, component errors
- **Fatal**: Unrecoverable errors (initialization failures, critical resource errors)

## Performance Considerations

Avoid logging in hot paths above Info level. Use conditional debug logging:

```go
if logger.GetLevel() >= logrus.DebugLevel {
    // Expensive operation only executed when debug logging is enabled
    logger.WithFields(expensiveFieldComputation()).Debug("detailed state")
}
```

## Configuration

### Environment Variables

- `LOG_LEVEL`: Set minimum log level (`debug`, `info`, `warn`, `error`, `fatal`)
- `LOG_FORMAT`: Set output format (`json`, `text`)

Example:
```bash
export LOG_LEVEL=debug
export LOG_FORMAT=json
./client
```

### Formatters

**Text Format** (development):
```
2024-10-25 14:30:45.123 level=info msg="server started" port=8080
```

**JSON Format** (production):
```json
{"timestamp":"2024-10-25T14:30:45.123Z","level":"info","message":"server started","port":8080}
```

## Integration Guidelines

### In Packages

Pass logger instances with context:

```go
type TerrainGenerator struct {
    logger *logrus.Entry
}

func NewTerrainGenerator(baseLogger *logrus.Logger) *TerrainGenerator {
    return &TerrainGenerator{
        logger: logging.SystemLogger(baseLogger, "terrain"),
    }
}

func (tg *TerrainGenerator) Generate(seed int64) error {
    tg.logger.WithFields(logrus.Fields{
        "seed": seed,
        "algorithm": "BSP",
    }).Info("starting terrain generation")
    
    // ... generation logic ...
    
    tg.logger.Debug("terrain generation complete")
    return nil
}
```

### Error Logging

Wrap errors with context:

```go
if err != nil {
    logger.WithError(err).WithFields(logrus.Fields{
        "entityID": entityID,
        "operation": "spawn",
    }).Error("failed to spawn entity")
    return fmt.Errorf("spawn entity: %w", err)
}
```

## Testing

The package includes comprehensive unit tests:

```bash
go test -v ./pkg/logging/...
```

Coverage target: 80%+
