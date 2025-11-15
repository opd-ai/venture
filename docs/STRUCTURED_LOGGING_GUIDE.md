# Structured Logging with Logrus

Implementation guide for structured logging in Venture using logrus.

## Quick Start

**CLI Tools:**
```go
import (
    "github.com/opd-ai/venture/pkg/logging"
    "github.com/sirupsen/logrus"
)

logger := logging.TestUtilityLogger("mycommand")
logger.WithFields(logrus.Fields{
    "seed":  12345,
    "genre": "fantasy",
}).Info("starting generation")
```

**Server (JSON format):**
```go
logger := logging.NewLogger(logging.Config{
    Level:      logging.InfoLevel,
    Format:     logging.JSONFormat,
    AddCaller:  true,
    EnableColor: false,
})

serverLogger := logger.WithFields(logrus.Fields{
    "component": "server",
    "port":      8080,
})
serverLogger.Info("server started")
```

**Environment Config:**
```bash
export LOG_LEVEL=debug  # debug, info, warn, error, fatal
export LOG_FORMAT=json  # json, text
./server
```

## Log Levels

**Debug:** System state, component lifecycle, seeds, algorithm steps  
```go
logger.WithFields(logrus.Fields{"seed": seed}).Debug("starting generation")
```

**Info:** Startup/shutdown, state changes, completion events, player connections  
```go
logger.WithFields(logrus.Fields{"roomCount": len(rooms)}).Info("terrain generated")
```

**Warn:** Validation issues, retry attempts, performance degradation, deprecations  
```go
logger.WithFields(logrus.Fields{"latency": latency}).Warn("high latency detected")
```

**Error:** Generation failures, network errors, invalid state, I/O errors  
```go
logger.WithError(err).WithFields(logrus.Fields{"operation": "generate"}).Error("generation failed")
```

**Fatal:** Initialization failures, critical errors (calls `os.Exit(1)`)  
```go
logger.WithError(err).Fatal("failed to start server")
```

## Context Helpers

```go
// System context
sysLogger := logging.SystemLogger(logger, "terrain")

// Entity context
entLogger := logging.EntityLogger(logger, entityID, "monster")

// Network context
netLogger := logging.NetworkLogger(logger, playerID, clientAddr)

// Generation context
genLogger := logging.GenerationLogger(logger, "terrain", seed, params)
```

## Best Practices

**1. Use Structured Fields (not string formatting):**
```go
// Good
logger.WithFields(logrus.Fields{"playerID": id, "x": x, "y": y}).Info("player moved")

// Bad
logger.Infof("Player %d moved to (%f, %f)", id, x, y)
```

**2. Standard Field Names:**
- `seed` - Generation seed
- `genre` - Genre ID (fantasy, sci-fi, etc.)
- `entityID` - Entity identifier
- `playerID` - Player identifier
- `component` - System/component name
- `operation` - Operation being performed

**3. Conditional Debug Logging:**
```go
if logger.GetLevel() >= logrus.DebugLevel {
    logger.WithFields(expensiveFields()).Debug("detailed state")
}
```

**4. Error Context:**
```go
logger.WithError(err).WithFields(logrus.Fields{
    "operation": "connect",
    "server": addr,
}).Error("connection failed")
```

## Package Integration

**Initialize in `init()`:**
```go
var logger *logrus.Entry

func init() {
    logger = logging.SystemLogger(logging.GetLogger(), "mypackage")
}
```

**Use in Functions:**
```go
func Generate(seed int64) error {
    logger.WithFields(logrus.Fields{"seed": seed}).Debug("starting generation")
    // ... generation logic ...
    logger.Info("generation complete")
    return nil
}
```

## Configuration Options

```go
type Config struct {
    Level       LogLevel    // debug, info, warn, error, fatal
    Format      LogFormat   // text, json
    AddCaller   bool        // Add file:line info
    EnableColor bool        // Color output (text format only)
    OutputPath  string      // File path or "stdout"/"stderr"
}

// Common configs
logging.ProductionConfig()  // JSON, Info, caller, no color
logging.DevelopmentConfig() // Text, Debug, caller, color
```

## Testing Logs

**Capture Logs in Tests:**
```go
func TestGeneration(t *testing.T) {
    logger := logging.TestLogger(t)
    generator := NewGenerator(logger)
    // Test will show logs on failure
}
```

**Suppress Logs:**
```go
logger := logging.NewLogger(logging.Config{
    Level: logging.ErrorLevel,  // Only errors
})
```

## Examples

**Terrain Generation:**
```go
logger := logging.GenerationLogger(baseLogger, "terrain", seed, params)
logger.Info("starting generation")
logger.WithFields(logrus.Fields{"roomCount": len(rooms)}).Debug("rooms placed")
if err != nil {
    logger.WithError(err).Error("generation failed")
}
logger.Info("generation complete")
```

**Network Events:**
```go
netLogger := logging.NetworkLogger(baseLogger, playerID, conn.RemoteAddr().String())
netLogger.Info("player connected")
netLogger.WithFields(logrus.Fields{"latency": latency}).Warn("high latency")
netLogger.Info("player disconnected")
```

**System Lifecycle:**
```go
sysLogger := logging.SystemLogger(baseLogger, "combat")
sysLogger.Info("system initialized")
sysLogger.WithFields(logrus.Fields{"entities": count}).Debug("processing entities")
```

---

**Last Updated:** November 14, 2025
