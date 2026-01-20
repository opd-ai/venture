// Package errors constants define error type categories.
// This file contains all ErrorType constant definitions, originally consolidated
// from errors.go during package reorganization.
package errors

// Originally defined in: errors.go
// Error type constants define the categories of errors in the Venture system.

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
