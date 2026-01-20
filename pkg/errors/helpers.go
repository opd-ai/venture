// Package errors helper functions provide convenient constructors for creating
// type-specific VentureError instances. Each error type has two helpers:
// - Type(message) creates a new error
// - TypeWrap(err, message) wraps an existing error
//
// This file contains 24 helper functions (12 types × 2 variants) originally
// consolidated from errors.go during package reorganization.
package errors

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

// FileSystem creates a file system error.
func FileSystem(message string) *VentureError {
	return &VentureError{
		Type:      ErrorTypeFileSystem,
		Message:   message,
		Context:   make(map[string]interface{}),
		Retryable: false, // File errors usually require manual intervention
	}
}

// FileSystemWrap wraps an error as a file system error.
func FileSystemWrap(err error, message string) *VentureError {
	if err == nil {
		return nil
	}
	return &VentureError{
		Type:      ErrorTypeFileSystem,
		Message:   message,
		Err:       err,
		Context:   make(map[string]interface{}),
		Retryable: false,
	}
}

// Authentication creates an authentication error.
func Authentication(message string) *VentureError {
	return &VentureError{
		Type:      ErrorTypeAuthentication,
		Message:   message,
		Context:   make(map[string]interface{}),
		Retryable: false, // Auth errors require new credentials
	}
}

// AuthenticationWrap wraps an error as an authentication error.
func AuthenticationWrap(err error, message string) *VentureError {
	if err == nil {
		return nil
	}
	return &VentureError{
		Type:      ErrorTypeAuthentication,
		Message:   message,
		Err:       err,
		Context:   make(map[string]interface{}),
		Retryable: false,
	}
}

// Concurrency creates a concurrency error.
func Concurrency(message string) *VentureError {
	return &VentureError{
		Type:      ErrorTypeConcurrency,
		Message:   message,
		Context:   make(map[string]interface{}),
		Retryable: false, // Concurrency errors usually indicate bugs
	}
}

// ConcurrencyWrap wraps an error as a concurrency error.
func ConcurrencyWrap(err error, message string) *VentureError {
	if err == nil {
		return nil
	}
	return &VentureError{
		Type:      ErrorTypeConcurrency,
		Message:   message,
		Err:       err,
		Context:   make(map[string]interface{}),
		Retryable: false,
	}
}

// Resource creates a resource exhaustion error.
func Resource(message string) *VentureError {
	return &VentureError{
		Type:      ErrorTypeResource,
		Message:   message,
		Context:   make(map[string]interface{}),
		Retryable: true, // Resource exhaustion may be transient
	}
}

// ResourceWrap wraps an error as a resource exhaustion error.
func ResourceWrap(err error, message string) *VentureError {
	if err == nil {
		return nil
	}
	return &VentureError{
		Type:      ErrorTypeResource,
		Message:   message,
		Err:       err,
		Context:   make(map[string]interface{}),
		Retryable: true,
	}
}
