package logging

import (
	"github.com/opd-ai/venture/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ErrorLogger creates a logger entry with error context.
// Extracts VentureError fields (type, correlation ID, context) if available.
// Returns nil if logger is nil to prevent panics.
func ErrorLogger(logger *logrus.Logger, err error) *logrus.Entry {
	if logger == nil {
		return nil
	}
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

		// Add error context as a nested object to avoid field name collisions
		if len(ventureErr.Context) > 0 {
			fields["error_context"] = ventureErr.Context
		}
	}

	return logger.WithFields(fields)
}

// LogError logs an error with full context extraction.
// Uses Warn level for retryable errors, Error level for non-retryable.
// Does nothing if logger is nil.
func LogError(logger *logrus.Logger, err error, message string) {
	if logger == nil {
		return
	}
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
// Returns nil if logger is nil to prevent panics.
func CorrelationLogger(logger *logrus.Logger, correlationID string) *logrus.Entry {
	if logger == nil {
		return nil
	}
	return logger.WithFields(logrus.Fields{
		"correlation_id": correlationID,
	})
}
