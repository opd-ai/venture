package logging

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/opd-ai/venture/pkg/errors"
	"github.com/sirupsen/logrus"
)

func TestErrorLogger(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantFields []string
	}{
		{
			name: "standard error",
			err:  errors.New(errors.ErrorTypeNetwork, "connection failed"),
			wantFields: []string{
				"error_type",
				"retryable",
			},
		},
		{
			name: "error with correlation ID",
			err: errors.Network("connection failed").
				WithCorrelationID("test-123"),
			wantFields: []string{
				"error_type",
				"correlation_id",
				"retryable",
			},
		},
		{
			name: "error with context",
			err: errors.Network("connection failed").
				WithContext("host", "example.com").
				WithContext("port", 8080),
			wantFields: []string{
				"error_type",
				"retryable",
				"error_context",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logrus.New()
			logger.SetOutput(&bytes.Buffer{}) // Discard output

			entry := ErrorLogger(logger, tt.err)

			for _, field := range tt.wantFields {
				if _, exists := entry.Data[field]; !exists {
					t.Errorf("ErrorLogger() missing field %s", field)
				}
			}
		})
	}
}

func TestLogError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		message   string
		wantLevel logrus.Level
	}{
		{
			name:      "retryable error uses warn level",
			err:       errors.Network("temporary failure"),
			message:   "network issue",
			wantLevel: logrus.WarnLevel,
		},
		{
			name:      "non-retryable error uses error level",
			err:       errors.Validation("invalid input"),
			message:   "validation failed",
			wantLevel: logrus.ErrorLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create logger with custom hook to capture log level
			logger := logrus.New()
			var capturedLevel logrus.Level
			logger.AddHook(&testHook{
				onFire: func(entry *logrus.Entry) error {
					capturedLevel = entry.Level
					return nil
				},
			})

			LogError(logger, tt.err, tt.message)

			if capturedLevel != tt.wantLevel {
				t.Errorf("LogError() level = %v, want %v", capturedLevel, tt.wantLevel)
			}
		})
	}
}

func TestErrorLogger_JSONOutput(t *testing.T) {
	// Test that error fields are properly serialized to JSON
	logger := logrus.New()
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.JSONFormatter{})

	err := errors.Network("connection failed").
		WithCorrelationID("test-correlation-123").
		WithContext("host", "example.com").
		WithContext("port", 8080)

	entry := ErrorLogger(logger, err)
	entry.Info("test message")

	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}

	// Verify fields are present
	expectedFields := []string{"error_type", "correlation_id", "error_context", "retryable"}
	for _, field := range expectedFields {
		if _, exists := logEntry[field]; !exists {
			t.Errorf("JSON output missing field %s", field)
		}
	}

	// Verify specific values
	if logEntry["error_type"] != "Network" {
		t.Errorf("error_type = %v, want Network", logEntry["error_type"])
	}
	if logEntry["correlation_id"] != "test-correlation-123" {
		t.Errorf("correlation_id = %v, want test-correlation-123", logEntry["correlation_id"])
	}
	
	// Verify context is nested
	errorContext, ok := logEntry["error_context"].(map[string]interface{})
	if !ok {
		t.Fatalf("error_context is not a map: %T", logEntry["error_context"])
	}
	if errorContext["host"] != "example.com" {
		t.Errorf("error_context.host = %v, want example.com", errorContext["host"])
	}
	// Port is a float64 in JSON
	if port, ok := errorContext["port"].(float64); !ok || port != 8080 {
		t.Errorf("error_context.port = %v, want 8080", errorContext["port"])
	}
}

func TestCorrelationLogger(t *testing.T) {
	logger := logrus.New()
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.JSONFormatter{})

	correlationID := "test-correlation-456"
	entry := CorrelationLogger(logger, correlationID)
	entry.Info("test message")

	output := buf.String()
	if !strings.Contains(output, correlationID) {
		t.Errorf("CorrelationLogger() output missing correlation ID: %s", output)
	}

	// Verify JSON structure
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}

	if logEntry["correlation_id"] != correlationID {
		t.Errorf("correlation_id = %v, want %s", logEntry["correlation_id"], correlationID)
	}
}

// testHook is a test helper for capturing log entries
type testHook struct {
	onFire func(*logrus.Entry) error
}

func (h *testHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *testHook) Fire(entry *logrus.Entry) error {
	if h.onFire != nil {
		return h.onFire(entry)
	}
	return nil
}
