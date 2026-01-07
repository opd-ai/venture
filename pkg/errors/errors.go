// Package errors provides structured error types with context and correlation ID support
// for distributed tracing across the Venture game system. This package implements
// comprehensive error wrapping patterns as specified in PLAN.md Phase 2.
package errors

import (
	"errors"
	"fmt"
)

// ErrorType represents different categories of errors in the system.
type ErrorType int

const (
	// ErrorTypeUnknown represents an unclassified error.
	ErrorTypeUnknown ErrorType = iota
	// ErrorTypeNetwork represents network-related errors (connection, timeout, protocol).
	ErrorTypeNetwork
	// ErrorTypeValidation represents input validation errors.
	ErrorTypeValidation
	// ErrorTypeConfiguration represents configuration errors.
	ErrorTypeConfiguration
	// ErrorTypeGeneration represents procedural generation errors.
	ErrorTypeGeneration
	// ErrorTypeSerialization represents data serialization/deserialization errors.
	ErrorTypeSerialization
	// ErrorTypeFileSystem represents file system operation errors.
	ErrorTypeFileSystem
	// ErrorTypeDatabase represents database/persistence errors.
	ErrorTypeDatabase
	// ErrorTypeAuthentication represents authentication/authorization errors.
	ErrorTypeAuthentication
	// ErrorTypeRateLimit represents rate limiting errors.
	ErrorTypeRateLimit
	// ErrorTypeConcurrency represents concurrency/synchronization errors.
	ErrorTypeConcurrency
	// ErrorTypeResource represents resource exhaustion errors (memory, CPU, etc.).
	ErrorTypeResource
	// ErrorTypeTimeout represents operation timeout errors.
	ErrorTypeTimeout
)

// String returns the string representation of an ErrorType.
func (e ErrorType) String() string {
	switch e {
	case ErrorTypeNetwork:
		return "Network"
	case ErrorTypeValidation:
		return "Validation"
	case ErrorTypeConfiguration:
		return "Configuration"
	case ErrorTypeGeneration:
		return "Generation"
	case ErrorTypeSerialization:
		return "Serialization"
	case ErrorTypeFileSystem:
		return "FileSystem"
	case ErrorTypeDatabase:
		return "Database"
	case ErrorTypeAuthentication:
		return "Authentication"
	case ErrorTypeRateLimit:
		return "RateLimit"
	case ErrorTypeConcurrency:
		return "Concurrency"
	case ErrorTypeResource:
		return "Resource"
	case ErrorTypeTimeout:
		return "Timeout"
	default:
		return "Unknown"
	}
}

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

// New creates a new VentureError with the specified type and message.
func New(errType ErrorType, message string) *VentureError {
	return &VentureError{
		Type:      errType,
		Message:   message,
		Context:   make(map[string]interface{}),
		Retryable: false,
	}
}

// Wrap wraps an existing error with VentureError context.
func Wrap(err error, errType ErrorType, message string) *VentureError {
	if err == nil {
		return nil
	}
	return &VentureError{
		Type:      errType,
		Message:   message,
		Err:       err,
		Context:   make(map[string]interface{}),
		Retryable: false,
	}
}

// Wrapf wraps an existing error with formatted message.
func Wrapf(err error, errType ErrorType, format string, args ...interface{}) *VentureError {
	if err == nil {
		return nil
	}
	return &VentureError{
		Type:      errType,
		Message:   fmt.Sprintf(format, args...),
		Err:       err,
		Context:   make(map[string]interface{}),
		Retryable: false,
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

// Helper functions for creating specific error types

// Network creates a network error.
func Network(message string) *VentureError {
	return &VentureError{
		Type:      ErrorTypeNetwork,
		Message:   message,
		Context:   make(map[string]interface{}),
		Retryable: true, // Network errors are usually retryable
	}
}

// NetworkWrap wraps an error as a network error.
func NetworkWrap(err error, message string) *VentureError {
	if err == nil {
		return nil
	}
	return &VentureError{
		Type:      ErrorTypeNetwork,
		Message:   message,
		Err:       err,
		Context:   make(map[string]interface{}),
		Retryable: true,
	}
}

// Validation creates a validation error.
func Validation(message string) *VentureError {
	return &VentureError{
		Type:      ErrorTypeValidation,
		Message:   message,
		Context:   make(map[string]interface{}),
		Retryable: false, // Validation errors require user correction
	}
}

// ValidationWrap wraps an error as a validation error.
func ValidationWrap(err error, message string) *VentureError {
	if err == nil {
		return nil
	}
	return &VentureError{
		Type:      ErrorTypeValidation,
		Message:   message,
		Err:       err,
		Context:   make(map[string]interface{}),
		Retryable: false,
	}
}

// Configuration creates a configuration error.
func Configuration(message string) *VentureError {
	return &VentureError{
		Type:      ErrorTypeConfiguration,
		Message:   message,
		Context:   make(map[string]interface{}),
		Retryable: false,
	}
}

// ConfigurationWrap wraps an error as a configuration error.
func ConfigurationWrap(err error, message string) *VentureError {
	if err == nil {
		return nil
	}
	return &VentureError{
		Type:      ErrorTypeConfiguration,
		Message:   message,
		Err:       err,
		Context:   make(map[string]interface{}),
		Retryable: false,
	}
}

// Timeout creates a timeout error.
func Timeout(message string) *VentureError {
	return &VentureError{
		Type:      ErrorTypeTimeout,
		Message:   message,
		Context:   make(map[string]interface{}),
		Retryable: true, // Timeouts are usually retryable
	}
}

// TimeoutWrap wraps an error as a timeout error.
func TimeoutWrap(err error, message string) *VentureError {
	if err == nil {
		return nil
	}
	return &VentureError{
		Type:      ErrorTypeTimeout,
		Message:   message,
		Err:       err,
		Context:   make(map[string]interface{}),
		Retryable: true,
	}
}

// Serialization creates a serialization error.
func Serialization(message string) *VentureError {
	return &VentureError{
		Type:      ErrorTypeSerialization,
		Message:   message,
		Context:   make(map[string]interface{}),
		Retryable: false,
	}
}

// SerializationWrap wraps an error as a serialization error.
func SerializationWrap(err error, message string) *VentureError {
	if err == nil {
		return nil
	}
	return &VentureError{
		Type:      ErrorTypeSerialization,
		Message:   message,
		Err:       err,
		Context:   make(map[string]interface{}),
		Retryable: false,
	}
}

// Generation creates a procedural generation error.
func Generation(message string) *VentureError {
	return &VentureError{
		Type:      ErrorTypeGeneration,
		Message:   message,
		Context:   make(map[string]interface{}),
		Retryable: false,
	}
}

// GenerationWrap wraps an error as a generation error.
func GenerationWrap(err error, message string) *VentureError {
	if err == nil {
		return nil
	}
	return &VentureError{
		Type:      ErrorTypeGeneration,
		Message:   message,
		Err:       err,
		Context:   make(map[string]interface{}),
		Retryable: false,
	}
}

// Database creates a database error.
func Database(message string) *VentureError {
	return &VentureError{
		Type:      ErrorTypeDatabase,
		Message:   message,
		Context:   make(map[string]interface{}),
		Retryable: true, // Database errors may be transient
	}
}

// DatabaseWrap wraps an error as a database error.
func DatabaseWrap(err error, message string) *VentureError {
	if err == nil {
		return nil
	}
	return &VentureError{
		Type:      ErrorTypeDatabase,
		Message:   message,
		Err:       err,
		Context:   make(map[string]interface{}),
		Retryable: true,
	}
}

// RateLimit creates a rate limit error.
func RateLimit(message string) *VentureError {
	return &VentureError{
		Type:      ErrorTypeRateLimit,
		Message:   message,
		Context:   make(map[string]interface{}),
		Retryable: true, // User can retry after waiting
	}
}

// RateLimitWrap wraps an error as a rate limit error.
func RateLimitWrap(err error, message string) *VentureError {
	if err == nil {
		return nil
	}
	return &VentureError{
		Type:      ErrorTypeRateLimit,
		Message:   message,
		Err:       err,
		Context:   make(map[string]interface{}),
		Retryable: true,
	}
}
