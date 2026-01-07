package logging

import (
	"github.com/opd-ai/venture/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ErrorLogger creates a logger entry with error context.
// Extracts VentureError fields (type, correlation ID, context) if available.
func ErrorLogger(logger *logrus.Logger, err error) *logrus.Entry {
	fields := logrus.Fields{
		"error": err.Error(),
	}

	// Extract VentureError details if available
	if ventureErr, ok := errors.AsVentureError(err); ok {
		fields["error_type"] = ventureErr.Type.String()
		fields["retryable"] = ventureErr.Retryable

		if ventureErr.CorrelationID != "" {
			fields["correlation_id"] = ventureErr.CorrelationID
		}

		// Add error context as individual fields
		for key, value := range ventureErr.Context {
			fields[key] = value
		}
	}

	return logger.WithFields(fields)
}

// LogError logs an error with full context extraction.
// Uses Info level for retryable errors, Error level for non-retryable.
func LogError(logger *logrus.Logger, err error, message string) {
	entry := ErrorLogger(logger, err)

	// Use different log levels based on error properties
	if ventureErr, ok := errors.AsVentureError(err); ok {
		if ventureErr.Retryable {
			entry.Warn(message)
		} else {
			entry.Error(message)
		}
	} else {
		entry.Error(message)
	}
}

// CorrelationLogger creates a logger entry with correlation ID from context.
// This should be used at the start of request handlers to establish tracing.
func CorrelationLogger(logger *logrus.Logger, correlationID string) *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"correlation_id": correlationID,
	})
}
