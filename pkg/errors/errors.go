// Package errors provides structured error types with context and correlation ID support
// for distributed tracing across the Venture game system. This package implements
// comprehensive error wrapping patterns as specified in PLAN.md Phase 2.
package errors

import (
	"errors"
	"fmt"
)

// VentureError represents a structured error with type, context, and correlation ID.
type VentureError struct {
	Type          ErrorType              // Error category
	Message       string                 // Technical error message
	UserMessage   string                 // User-friendly error message (optional)
	Err           error                  // Wrapped underlying error (optional)
	Context       map[string]interface{} // Additional context data
	CorrelationID string                 // Distributed tracing correlation ID
	Retryable     bool                   // Whether the operation can be retried
}

// Error implements the error interface.
func (e *VentureError) Error() string {
	if e.CorrelationID != "" {
		return fmt.Sprintf("[%s][%s] %s", e.Type.String(), e.CorrelationID, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Type.String(), e.Message)
}

// Unwrap implements error unwrapping for errors.Is and errors.As.
func (e *VentureError) Unwrap() error {
	return e.Err
}

// WithContext adds context information to the error.
func (e *VentureError) WithContext(key string, value interface{}) *VentureError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithCorrelationID adds a correlation ID for distributed tracing.
func (e *VentureError) WithCorrelationID(id string) *VentureError {
	e.CorrelationID = id
	return e
}

// GetUserMessage returns a user-friendly error message.
// Falls back to technical message if no user message is set.
func (e *VentureError) GetUserMessage() string {
	if e.UserMessage != "" {
		return e.UserMessage
	}
	// Provide generic user-friendly messages based on type
	switch e.Type {
	case ErrorTypeNetwork:
		return "Network error. Please check your connection and try again."
	case ErrorTypeValidation:
		return "Invalid input. Please check your data and try again."
	case ErrorTypeConfiguration:
		return "Configuration error. Please check your settings."
	case ErrorTypeGeneration:
		return "Failed to generate content. Please try again with a different seed."
	case ErrorTypeTimeout:
		return "Operation timed out. Please try again."
	case ErrorTypeRateLimit:
		return "Too many requests. Please slow down and try again later."
	case ErrorTypeAuthentication:
		return "Authentication failed. Please check your credentials."
	case ErrorTypeResource:
		return "System resources exhausted. Please try again later."
	default:
		return "An error occurred. Please try again."
	}
}

// IsRetryable returns whether the error represents a retryable condition.
func (e *VentureError) IsRetryable() bool {
	return e.Retryable
}

// isRetryableType determines if an error type is retryable by default.
// Network, timeout, database, rate limit, and resource errors are retryable.
func isRetryableType(errType ErrorType) bool {
	switch errType {
	case ErrorTypeNetwork, ErrorTypeTimeout, ErrorTypeDatabase, ErrorTypeRateLimit, ErrorTypeResource:
		return true
	default:
		return false
	}
}

// New creates a new VentureError with the specified type and message.
// Retryability is set based on the error type.
func New(errType ErrorType, message string) *VentureError {
	return &VentureError{
		Type:      errType,
		Message:   message,
		Context:   make(map[string]interface{}),
		Retryable: isRetryableType(errType),
	}
}

// Wrap wraps an existing error with VentureError context.
// Retryability is set based on the error type.
func Wrap(err error, errType ErrorType, message string) *VentureError {
	if err == nil {
		return nil
	}
	return &VentureError{
		Type:      errType,
		Message:   message,
		Err:       err,
		Context:   make(map[string]interface{}),
		Retryable: isRetryableType(errType),
	}
}

// Wrapf wraps an existing error with formatted message.
// Retryability is set based on the error type.
func Wrapf(err error, errType ErrorType, format string, args ...interface{}) *VentureError {
	if err == nil {
		return nil
	}
	return &VentureError{
		Type:      errType,
		Message:   fmt.Sprintf(format, args...),
		Err:       err,
		Context:   make(map[string]interface{}),
		Retryable: isRetryableType(errType),
	}
}

// Is checks if the error is a VentureError of the specified type.
func Is(err error, errType ErrorType) bool {
	var ventureErr *VentureError
	if errors.As(err, &ventureErr) {
		return ventureErr.Type == errType
	}
	return false
}

// AsVentureError attempts to convert an error to VentureError.
func AsVentureError(err error) (*VentureError, bool) {
	var ventureErr *VentureError
	if errors.As(err, &ventureErr) {
		return ventureErr, true
	}
	return nil, false
}
