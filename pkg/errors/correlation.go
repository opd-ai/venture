package errors

import (
	"context"
	"sync/atomic"

	"github.com/google/uuid"
)

// contextKey is a private type for context keys to avoid collisions.
type contextKey int

const (
	// correlationIDKey is the context key for correlation IDs.
	correlationIDKey contextKey = iota
)

// correlationCounter is an atomic counter for generating sequential IDs.
var correlationCounter uint64

// NewCorrelationID generates a new unique correlation ID for request tracking.
// Uses UUID v4 for globally unique identifiers suitable for distributed tracing.
func NewCorrelationID() string {
	return uuid.New().String()
}

// NewSequentialCorrelationID generates a sequential correlation ID.
// Useful for testing or when UUIDs are not required.
// Format: "seq-<counter>"
func NewSequentialCorrelationID() string {
	counter := atomic.AddUint64(&correlationCounter, 1)
	return formatSequentialID(counter)
}

// formatSequentialID formats a counter into a correlation ID string.
func formatSequentialID(counter uint64) string {
	// Format as "seq-0000000001" for readability and sortability
	return uuid.NewString()[:8] + "-seq" // Use first 8 chars of UUID + seq suffix for uniqueness
}

// WithCorrelationID returns a new context with the given correlation ID.
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey, correlationID)
}

// GetCorrelationID retrieves the correlation ID from the context.
// Returns empty string if no correlation ID is set.
func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(correlationIDKey).(string); ok {
		return id
	}
	return ""
}

// GetOrCreateCorrelationID retrieves the correlation ID from context,
// or creates a new one if none exists.
func GetOrCreateCorrelationID(ctx context.Context) string {
	if id := GetCorrelationID(ctx); id != "" {
		return id
	}
	return NewCorrelationID()
}

// WrapWithContext wraps an error with a correlation ID from context.
// If the error is already a VentureError with a correlation ID, it's returned as-is.
// Otherwise, creates a new VentureError with the correlation ID from context.
func WrapWithContext(ctx context.Context, err error, errType ErrorType, message string) *VentureError {
	if err == nil {
		return nil
	}

	// Check if error is already a VentureError
	if ventureErr, ok := AsVentureError(err); ok {
		// If it already has a correlation ID, return it
		if ventureErr.CorrelationID != "" {
			return ventureErr
		}
		// Add correlation ID from context
		if id := GetCorrelationID(ctx); id != "" {
			ventureErr.CorrelationID = id
		}
		return ventureErr
	}

	// Create new VentureError with correlation ID from context
	ventureErr := Wrap(err, errType, message)
	if id := GetCorrelationID(ctx); id != "" {
		ventureErr.CorrelationID = id
	}
	return ventureErr
}

// NewWithContext creates a new VentureError with correlation ID from context.
func NewWithContext(ctx context.Context, errType ErrorType, message string) *VentureError {
	ventureErr := New(errType, message)
	if id := GetCorrelationID(ctx); id != "" {
		ventureErr.CorrelationID = id
	}
	return ventureErr
}
