// Package errors type definitions for error categorization.
// This file defines ErrorType and its String() method for converting error types
// to human-readable strings. Originally consolidated from errors.go during
// package reorganization.
package errors

// ErrorType represents different categories of errors in the system.
// Constants are defined in constants.go
type ErrorType int

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
