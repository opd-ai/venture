package logging

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// LogLevel represents the minimum log level.
type LogLevel string

// Log level constants define the minimum severity level for log messages.
const (
	// DebugLevel is for detailed debugging information.
	DebugLevel LogLevel = "debug"
	// InfoLevel is for general informational messages.
	InfoLevel LogLevel = "info"
	// WarnLevel is for warning messages.
	WarnLevel LogLevel = "warn"
	// ErrorLevel is for error messages.
	ErrorLevel LogLevel = "error"
	// FatalLevel is for fatal errors that cause the program to exit.
	FatalLevel LogLevel = "fatal"
)

// LogFormat represents the output format for logs.
type LogFormat string

// Log format constants define the output format for logs.
const (
	// JSONFormat outputs logs in JSON format for machine parsing.
	JSONFormat LogFormat = "json"
	// TextFormat outputs logs in human-readable text format.
	TextFormat LogFormat = "text"
)

// Config holds logger configuration.
type Config struct {
	// Level sets the minimum log level
	Level LogLevel

	// Format sets the output format (json or text)
	Format LogFormat

	// AddCaller adds file and line number to log entries
	AddCaller bool

	// EnableColor enables colored output for text format
	EnableColor bool
}

// DefaultConfig returns a default logger configuration.
func DefaultConfig() Config {
	return Config{
		Level:       InfoLevel,
		Format:      TextFormat,
		AddCaller:   true,
		EnableColor: true,
	}
}

// NewLogger creates a new configured logger instance.
func NewLogger(config Config) *logrus.Logger {
	logger := logrus.New()

	// Set log level
	logger.SetLevel(parseLogLevel(config.Level))

	// Set formatter
	switch config.Format {
	case JSONFormat:
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "caller",
			},
		})
	default:
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05.000",
			FullTimestamp:   true,
			ForceColors:     config.EnableColor,
			DisableColors:   !config.EnableColor,
		})
	}

	// Enable caller reporting if requested
	logger.SetReportCaller(config.AddCaller)

	// Set output to stdout
	logger.SetOutput(os.Stdout)

	return logger
}

// NewLoggerFromEnv creates a logger configured from environment variables.
// Reads LOG_LEVEL and LOG_FORMAT environment variables.
func NewLoggerFromEnv() *logrus.Logger {
	config := DefaultConfig()

	// Override from environment
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Level = LogLevel(strings.ToLower(level))
	}

	if format := os.Getenv("LOG_FORMAT"); format != "" {
		config.Format = LogFormat(strings.ToLower(format))
	}

	return NewLogger(config)
}

// parseLogLevel converts LogLevel to logrus.Level.
func parseLogLevel(level LogLevel) logrus.Level {
	switch level {
	case DebugLevel:
		return logrus.DebugLevel
	case InfoLevel:
		return logrus.InfoLevel
	case WarnLevel:
		return logrus.WarnLevel
	case ErrorLevel:
		return logrus.ErrorLevel
	case FatalLevel:
		return logrus.FatalLevel
	default:
		return logrus.InfoLevel
	}
}

// WithContext creates a logger with standard context fields.
// This is useful for adding common fields that should appear in all logs from a component.
// Returns nil if logger is nil to prevent panics.
func WithContext(logger *logrus.Logger, fields logrus.Fields) *logrus.Entry {
	if logger == nil {
		return nil
	}
	return logger.WithFields(fields)
}

// SystemLogger creates a logger with system context.
// Returns nil if logger is nil to prevent panics.
func SystemLogger(logger *logrus.Logger, systemName string) *logrus.Entry {
	if logger == nil {
		return nil
	}
	return logger.WithFields(logrus.Fields{
		"system": systemName,
	})
}

// ComponentLogger creates a logger with component context.
// Returns nil if logger is nil to prevent panics.
func ComponentLogger(logger *logrus.Logger, componentType string) *logrus.Entry {
	if logger == nil {
		return nil
	}
	return logger.WithFields(logrus.Fields{
		"component": componentType,
	})
}

// EntityLogger creates a logger with entity context.
// Returns nil if logger is nil to prevent panics.
func EntityLogger(logger *logrus.Logger, entityID int) *logrus.Entry {
	if logger == nil {
		return nil
	}
	return logger.WithFields(logrus.Fields{
		"entityID": entityID,
	})
}

// GeneratorLogger creates a logger with procedural generation context.
// Returns nil if logger is nil to prevent panics.
//
// Example with conditional debug logging for expensive operations:
//
//	log := GeneratorLogger(logger, "terrain", seed, "fantasy")
//	if log.Logger.GetLevel() >= logrus.DebugLevel {
//	    log.WithFields(logrus.Fields{
//	        "noiseParams": expensiveNoiseParamString(),
//	        "chunkCount": len(chunks),
//	    }).Debug("detailed generation state")
//	}
func GeneratorLogger(logger *logrus.Logger, generatorType string, seed int64, genreID string) *logrus.Entry {
	if logger == nil {
		return nil
	}
	return logger.WithFields(logrus.Fields{
		"generator": generatorType,
		"seed":      seed,
		"genreID":   genreID,
	})
}

// NetworkLogger creates a logger with network context.
// Returns nil if logger is nil to prevent panics.
func NetworkLogger(logger *logrus.Logger, playerID, connectionState string) *logrus.Entry {
	if logger == nil {
		return nil
	}
	return logger.WithFields(logrus.Fields{
		"playerID":        playerID,
		"connectionState": connectionState,
	})
}

// PerformanceLogger creates a logger with performance metrics context.
// Returns nil if logger is nil to prevent panics.
//
// Example with conditional debug logging for expensive operations:
//
//	log := PerformanceLogger(logger, "renderFrame")
//	if log.Logger.GetLevel() >= logrus.DebugLevel {
//	    log.WithFields(logrus.Fields{
//	        "vertexCount": countVertices(),
//	        "drawCalls": countDrawCalls(),
//	    }).Debug("detailed performance metrics")
//	}
func PerformanceLogger(logger *logrus.Logger, operation string) *logrus.Entry {
	if logger == nil {
		return nil
	}
	return logger.WithFields(logrus.Fields{
		"operation": operation,
	})
}

// CombatLogger creates a logger with combat context.
// Returns nil if logger is nil to prevent panics.
func CombatLogger(logger *logrus.Logger, attackerID, targetID int) *logrus.Entry {
	if logger == nil {
		return nil
	}
	return logger.WithFields(logrus.Fields{
		"attackerID": attackerID,
		"targetID":   targetID,
	})
}

// SaveLoadLogger creates a logger with save/load context.
// Returns nil if logger is nil to prevent panics.
func SaveLoadLogger(logger *logrus.Logger, operation, path string) *logrus.Entry {
	if logger == nil {
		return nil
	}
	return logger.WithFields(logrus.Fields{
		"operation": operation,
		"path":      path,
	})
}

// TestUtilityLogger creates a logger configured for CLI test utilities.
// Uses colored text format for better readability in terminal.
// The utilityName is added as a "utility" field to all log entries via a hook.
func TestUtilityLogger(utilityName string) *logrus.Logger {
	config := Config{
		Level:       InfoLevel,
		Format:      TextFormat,
		AddCaller:   false, // Cleaner output for CLI tools
		EnableColor: true,  // Color for terminal readability
	}

	// Override from environment if set
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Level = LogLevel(strings.ToLower(level))
	}

	logger := NewLogger(config)

	// Add utility name as field for all logs via hook
	logger.AddHook(&utilityNameHook{utilityName: utilityName})

	return logger
}

// utilityNameHook is a logrus hook that adds the utility name field to all log entries.
type utilityNameHook struct {
	utilityName string
}

// Levels returns all log levels for which the hook should fire.
func (h *utilityNameHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire adds the utility name field to the log entry.
func (h *utilityNameHook) Fire(entry *logrus.Entry) error {
	entry.Data["utility"] = h.utilityName
	return nil
}
