package logging

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Level != InfoLevel {
		t.Errorf("expected default level %v, got %v", InfoLevel, config.Level)
	}
	if config.Format != TextFormat {
		t.Errorf("expected default format %v, got %v", TextFormat, config.Format)
	}
	if !config.AddCaller {
		t.Error("expected AddCaller to be true")
	}
	if !config.EnableColor {
		t.Error("expected EnableColor to be true")
	}
}

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		level  logrus.Level
	}{
		{
			name: "debug level",
			config: Config{
				Level:  DebugLevel,
				Format: TextFormat,
			},
			level: logrus.DebugLevel,
		},
		{
			name: "info level",
			config: Config{
				Level:  InfoLevel,
				Format: JSONFormat,
			},
			level: logrus.InfoLevel,
		},
		{
			name: "warn level",
			config: Config{
				Level:  WarnLevel,
				Format: TextFormat,
			},
			level: logrus.WarnLevel,
		},
		{
			name: "error level",
			config: Config{
				Level:  ErrorLevel,
				Format: JSONFormat,
			},
			level: logrus.ErrorLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.config)
			if logger == nil {
				t.Fatal("expected non-nil logger")
			}
			if logger.GetLevel() != tt.level {
				t.Errorf("expected level %v, got %v", tt.level, logger.GetLevel())
			}
		})
	}
}

func TestNewLoggerFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		envLevel string
		envFmt   string
		wantLvl  logrus.Level
	}{
		{
			name:     "debug from env",
			envLevel: "debug",
			envFmt:   "json",
			wantLvl:  logrus.DebugLevel,
		},
		{
			name:     "info from env",
			envLevel: "INFO",
			envFmt:   "text",
			wantLvl:  logrus.InfoLevel,
		},
		{
			name:     "warn from env",
			envLevel: "Warn",
			envFmt:   "json",
			wantLvl:  logrus.WarnLevel,
		},
		{
			name:     "no env vars",
			envLevel: "",
			envFmt:   "",
			wantLvl:  logrus.InfoLevel, // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment
			if tt.envLevel != "" {
				os.Setenv("LOG_LEVEL", tt.envLevel)
				defer os.Unsetenv("LOG_LEVEL")
			}
			if tt.envFmt != "" {
				os.Setenv("LOG_FORMAT", tt.envFmt)
				defer os.Unsetenv("LOG_FORMAT")
			}

			logger := NewLoggerFromEnv()
			if logger == nil {
				t.Fatal("expected non-nil logger")
			}
			if logger.GetLevel() != tt.wantLvl {
				t.Errorf("expected level %v, got %v", tt.wantLvl, logger.GetLevel())
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input LogLevel
		want  logrus.Level
	}{
		{DebugLevel, logrus.DebugLevel},
		{InfoLevel, logrus.InfoLevel},
		{WarnLevel, logrus.WarnLevel},
		{ErrorLevel, logrus.ErrorLevel},
		{FatalLevel, logrus.FatalLevel},
		{"invalid", logrus.InfoLevel}, // default
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			got := parseLogLevel(tt.input)
			if got != tt.want {
				t.Errorf("parseLogLevel(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestWithContext(t *testing.T) {
	logger := NewLogger(DefaultConfig())
	entry := WithContext(logger, logrus.Fields{"key": "value"})

	if entry == nil {
		t.Fatal("expected non-nil entry")
	}
	if entry.Data["key"] != "value" {
		t.Errorf("expected field key=value, got %v", entry.Data["key"])
	}
}

func TestSystemLogger(t *testing.T) {
	logger := NewLogger(DefaultConfig())
	entry := SystemLogger(logger, "terrain")

	if entry == nil {
		t.Fatal("expected non-nil entry")
	}
	if entry.Data["system"] != "terrain" {
		t.Errorf("expected system=terrain, got %v", entry.Data["system"])
	}
}

func TestComponentLogger(t *testing.T) {
	logger := NewLogger(DefaultConfig())
	entry := ComponentLogger(logger, "position")

	if entry == nil {
		t.Fatal("expected non-nil entry")
	}
	if entry.Data["component"] != "position" {
		t.Errorf("expected component=position, got %v", entry.Data["component"])
	}
}

func TestEntityLogger(t *testing.T) {
	logger := NewLogger(DefaultConfig())
	entry := EntityLogger(logger, 12345)

	if entry == nil {
		t.Fatal("expected non-nil entry")
	}
	if entry.Data["entityID"] != 12345 {
		t.Errorf("expected entityID=12345, got %v", entry.Data["entityID"])
	}
}

func TestGeneratorLogger(t *testing.T) {
	logger := NewLogger(DefaultConfig())
	entry := GeneratorLogger(logger, "terrain", 67890, "fantasy")

	if entry == nil {
		t.Fatal("expected non-nil entry")
	}
	if entry.Data["generator"] != "terrain" {
		t.Errorf("expected generator=terrain, got %v", entry.Data["generator"])
	}
	if entry.Data["seed"] != int64(67890) {
		t.Errorf("expected seed=67890, got %v", entry.Data["seed"])
	}
	if entry.Data["genreID"] != "fantasy" {
		t.Errorf("expected genreID=fantasy, got %v", entry.Data["genreID"])
	}
}

func TestNetworkLogger(t *testing.T) {
	logger := NewLogger(DefaultConfig())
	entry := NetworkLogger(logger, "player123", "connected")

	if entry == nil {
		t.Fatal("expected non-nil entry")
	}
	if entry.Data["playerID"] != "player123" {
		t.Errorf("expected playerID=player123, got %v", entry.Data["playerID"])
	}
	if entry.Data["connectionState"] != "connected" {
		t.Errorf("expected connectionState=connected, got %v", entry.Data["connectionState"])
	}
}

func TestLoggerOutput(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:       InfoLevel,
		Format:      TextFormat,
		AddCaller:   false,
		EnableColor: false,
	})
	logger.SetOutput(&buf)

	logger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("expected log output to contain 'test message', got: %s", output)
	}
	if !strings.Contains(output, "info") && !strings.Contains(output, "INFO") {
		t.Errorf("expected log output to contain log level, got: %s", output)
	}
}

func TestJSONFormatter(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:     InfoLevel,
		Format:    JSONFormat,
		AddCaller: false,
	})
	logger.SetOutput(&buf)

	logger.WithFields(logrus.Fields{
		"entityID": 123,
		"system":   "combat",
	}).Info("test message")

	output := buf.String()
	if !strings.Contains(output, "\"message\":\"test message\"") {
		t.Errorf("expected JSON output to contain message field, got: %s", output)
	}
	if !strings.Contains(output, "\"entityID\":123") {
		t.Errorf("expected JSON output to contain entityID field, got: %s", output)
	}
	if !strings.Contains(output, "\"system\":\"combat\"") {
		t.Errorf("expected JSON output to contain system field, got: %s", output)
	}
}
