package logging

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// LogLevel represents the minimum log level.
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
)

// LogFormat represents the output format for logs.
type LogFormat string

const (
	JSONFormat LogFormat = "json"
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
func WithContext(logger *logrus.Logger, fields logrus.Fields) *logrus.Entry {
	return logger.WithFields(fields)
}

// SystemLogger creates a logger with system context.
func SystemLogger(logger *logrus.Logger, systemName string) *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"system": systemName,
	})
}

// ComponentLogger creates a logger with component context.
func ComponentLogger(logger *logrus.Logger, componentType string) *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"component": componentType,
	})
}

// EntityLogger creates a logger with entity context.
func EntityLogger(logger *logrus.Logger, entityID int) *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"entityID": entityID,
	})
}

// GeneratorLogger creates a logger with procedural generation context.
func GeneratorLogger(logger *logrus.Logger, generatorType string, seed int64, genreID string) *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"generator": generatorType,
		"seed":      seed,
		"genreID":   genreID,
	})
}

// NetworkLogger creates a logger with network context.
func NetworkLogger(logger *logrus.Logger, playerID string, connectionState string) *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"playerID":        playerID,
		"connectionState": connectionState,
	})
}

// PerformanceLogger creates a logger with performance metrics context.
func PerformanceLogger(logger *logrus.Logger, operation string) *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"operation": operation,
	})
}

// CombatLogger creates a logger with combat context.
func CombatLogger(logger *logrus.Logger, attackerID, targetID int) *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"attackerID": attackerID,
		"targetID":   targetID,
	})
}

// SaveLoadLogger creates a logger with save/load context.
func SaveLoadLogger(logger *logrus.Logger, operation string, path string) *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"operation": operation,
		"path":      path,
	})
}

// TestUtilityLogger creates a logger configured for CLI test utilities.
// Uses colored text format for better readability in terminal.
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
	
	// Add utility name as field for all logs
	return logger
}
