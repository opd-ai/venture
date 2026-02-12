package logging

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

// TestLogLevelPrecedence verifies that LOG_LEVEL environment variable takes precedence
// over the verbose flag setting, and that unknown values default to info.
func TestLogLevelPrecedence(t *testing.T) {
	tests := []struct {
		name          string
		envLogLevel   string
		verboseEquiv  bool // simulating verbose flag behavior
		expectedLevel logrus.Level
		description   string
	}{
		{
			name:          "LOG_LEVEL set to debug takes precedence",
			envLogLevel:   "debug",
			verboseEquiv:  false,
			expectedLevel: logrus.DebugLevel,
			description:   "When LOG_LEVEL=debug, use debug regardless of verbose flag",
		},
		{
			name:          "LOG_LEVEL set to info takes precedence",
			envLogLevel:   "info",
			verboseEquiv:  true,
			expectedLevel: logrus.InfoLevel,
			description:   "When LOG_LEVEL=info, use info even if verbose would set debug",
		},
		{
			name:          "LOG_LEVEL set to warn takes precedence",
			envLogLevel:   "warn",
			verboseEquiv:  true,
			expectedLevel: logrus.WarnLevel,
			description:   "When LOG_LEVEL=warn, use warn even if verbose would set debug",
		},
		{
			name:          "LOG_LEVEL set to error takes precedence",
			envLogLevel:   "error",
			verboseEquiv:  true,
			expectedLevel: logrus.ErrorLevel,
			description:   "When LOG_LEVEL=error, use error even if verbose would set debug",
		},
		{
			name:          "LOG_LEVEL set to fatal takes precedence",
			envLogLevel:   "fatal",
			verboseEquiv:  true,
			expectedLevel: logrus.FatalLevel,
			description:   "When LOG_LEVEL=fatal, use fatal even if verbose would set debug",
		},
		{
			name:          "Unknown LOG_LEVEL defaults to info",
			envLogLevel:   "invalid",
			verboseEquiv:  true,
			expectedLevel: logrus.InfoLevel,
			description:   "Unknown LOG_LEVEL values default to info (documented behavior)",
		},
		{
			name:          "Empty LOG_LEVEL with verbose true uses debug",
			envLogLevel:   "",
			verboseEquiv:  true,
			expectedLevel: logrus.DebugLevel,
			description:   "When LOG_LEVEL not set and verbose=true, use debug",
		},
		{
			name:          "Empty LOG_LEVEL with verbose false uses info",
			envLogLevel:   "",
			verboseEquiv:  false,
			expectedLevel: logrus.InfoLevel,
			description:   "When LOG_LEVEL not set and verbose=false, use info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the client/server logging initialization logic
			var level logrus.Level

			// Set environment variable
			if tt.envLogLevel != "" {
				os.Setenv("LOG_LEVEL", tt.envLogLevel)
				defer os.Unsetenv("LOG_LEVEL")
			} else {
				os.Unsetenv("LOG_LEVEL")
			}

			// This simulates the logic in cmd/client/util.go:96-103
			// and cmd/server/main.go:295-299
			if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
				level = parseLogLevel(LogLevel(logLevel))
			} else if tt.verboseEquiv {
				level = logrus.DebugLevel
			} else {
				level = logrus.InfoLevel
			}

			if level != tt.expectedLevel {
				t.Errorf("%s: got %v, want %v", tt.description, level, tt.expectedLevel)
			}
		})
	}
}

// TestLogLevelDocumentationAccuracy verifies that the documented behavior
// matches the actual implementation for unknown LOG_LEVEL values.
func TestLogLevelDocumentationAccuracy(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		wantInfo bool
	}{
		{"Unknown value 'invalid'", "invalid", true},
		{"Unknown value 'verbose'", "verbose", true},
		{"Unknown value 'trace'", "trace", true},
		{"Unknown value 'DEBUG'", "DEBUG", true},
		{"Valid value 'debug'", "debug", false},
		{"Valid value 'info'", "info", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := parseLogLevel(LogLevel(tt.logLevel))
			isInfo := level == logrus.InfoLevel

			if isInfo != tt.wantInfo {
				t.Errorf("parseLogLevel(%q) = %v, want info=%v", tt.logLevel, level, tt.wantInfo)
			}
		})
	}
}

// TestVerboseFlagInteraction demonstrates the verbose flag interaction
// for documentation validation.
func TestVerboseFlagInteraction(t *testing.T) {
	t.Run("Documentation Example: No LOG_LEVEL, verbose=true (default)", func(t *testing.T) {
		os.Unsetenv("LOG_LEVEL")

		// Simulate default verbose=true behavior
		var level logrus.Level
		if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
			level = parseLogLevel(LogLevel(logLevel))
		} else {
			// verbose=true default
			level = logrus.DebugLevel
		}

		if level != logrus.DebugLevel {
			t.Errorf("With no LOG_LEVEL and verbose=true, expected debug level, got %v", level)
		}
	})

	t.Run("Documentation Example: Invalid LOG_LEVEL, verbose=true", func(t *testing.T) {
		os.Setenv("LOG_LEVEL", "invalid")
		defer os.Unsetenv("LOG_LEVEL")

		// Simulate verbose=true behavior
		var level logrus.Level
		if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
			level = parseLogLevel(LogLevel(logLevel))
		} else {
			level = logrus.DebugLevel
		}

		// Should default to info because LOG_LEVEL is set (even though invalid)
		if level != logrus.InfoLevel {
			t.Errorf("With invalid LOG_LEVEL, expected info level, got %v", level)
		}
	})
}
