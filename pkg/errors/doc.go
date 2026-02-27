/*
Package errors provides comprehensive error handling with structured error types,
context enrichment, and correlation ID support for distributed tracing across the
Venture game system.

# Overview

This package implements Phase 2 requirements from PLAN.md for comprehensive error
wrapping with:
- Structured error types for common failure categories
- Context enrichment with arbitrary key-value pairs
- Correlation ID support for distributed request tracing
- User-friendly error messages separate from technical details
- Error wrapping that preserves the error chain (errors.Is/As support)

# Basic Usage

Create a new error:

	err := errors.Network("connection timeout")

Wrap an existing error:

	if err != nil {
		return errors.NetworkWrap(err, "failed to connect to server")
	}

Add context to errors:

	err := errors.Validation("invalid input").
		WithContext("field", "username").
		WithContext("value", username)

# Correlation IDs

Correlation IDs enable distributed tracing across federated servers and components:

	// Create context with correlation ID
	ctx := context.Background()
	correlationID := errors.NewCorrelationID()
	ctx = errors.WithCorrelationID(ctx, correlationID)

	// Errors created from this context inherit the correlation ID
	err := errors.NewWithContext(ctx, errors.ErrorTypeNetwork, "connection failed")
	// err.CorrelationID == correlationID

	// Wrap existing errors with correlation ID from context
	if err := someOperation(); err != nil {
		return errors.WrapWithContext(ctx, err, errors.ErrorTypeNetwork, "operation failed")
	}

# Error Types

The package provides predefined error types for common failure categories:
- ErrorTypeNetwork: Network connectivity and protocol errors
- ErrorTypeValidation: Input validation failures
- ErrorTypeConfiguration: Configuration errors
- ErrorTypeGeneration: Procedural generation failures
- ErrorTypeSerialization: Data encoding/decoding errors
- ErrorTypeFileSystem: File I/O errors
- ErrorTypeDatabase: Database/persistence errors
- ErrorTypeAuthentication: Auth/authorization failures
- ErrorTypeRateLimit: Rate limiting errors
- ErrorTypeTimeout: Operation timeouts
- ErrorTypeConcurrency: Concurrency/locking errors
- ErrorTypeResource: Resource exhaustion (memory, CPU, etc.)

# User-Friendly Messages

Errors can provide both technical and user-friendly messages:

	err := errors.Network("TCP connection refused on 127.0.0.1:8080")
	err.UserMessage = "Cannot connect to game server. Please check your network."

	// For display to users:
	// In production UI code, use err.GetUserMessage() for user-facing error dialogs.
	// Example (simplified): userMsg := err.GetUserMessage()

	// For logs (using logrus structured logging):
	logrus.WithError(err).Error("network connection failed") // "[Network] TCP connection refused..."

# Retryability

Errors indicate whether the operation can be retried:

	err := errors.Network("temporary connection failure")
	if err.IsRetryable() {
		// Retry the operation
	}

Network, timeout, database, and rate limit errors are retryable by default.
Validation, configuration, and generation errors are not retryable.

# Error Chain Support

The package fully supports Go 1.13+ error wrapping:

	baseErr := fmt.Errorf("io error")
	networkErr := errors.NetworkWrap(baseErr, "connection failed")

	// Check error types
	if errors.Is(networkErr, errors.ErrorTypeNetwork) {
		// Handle network error
	}

	// Extract VentureError
	if ventureErr, ok := errors.AsVentureError(networkErr); ok {
		logrus.WithFields(ventureErr.Context).Error(ventureErr.Message)
	}

	// Check wrapped error
	if errors.Is(networkErr, baseErr) {
		// Original error is in the chain
	}

# Integration with Logging

Errors integrate with the logging package (pkg/logging) for structured logging:

	import (
		"github.com/opd-ai/venture/pkg/errors"
		"github.com/sirupsen/logrus"
	)

	err := errors.Network("connection failed").
		WithContext("host", "game-server.example.com").
		WithContext("port", 8080).
		WithCorrelationID(correlationID)

	logger.WithFields(logrus.Fields{
		"error_type":     err.Type.String(),
		"correlation_id": err.CorrelationID,
		"context":        err.Context,
	}).Error(err.Message)

# Best Practices

1. Always wrap errors with context:
  - Use Wrap/Wrapf for existing errors
  - Add relevant context with WithContext
  - Include correlation IDs in request handlers

2. Choose appropriate error types:
  - Use specific types (Network, Validation) over Unknown
  - Match the error type to the root cause

3. Provide user-friendly messages:
  - Set UserMessage for errors shown to end users
  - Keep technical details in Message for logs

4. Preserve error chains:
  - Use Wrap instead of creating new errors
  - Enables errors.Is and errors.As checks

5. Log errors with correlation IDs:
  - Always include correlation ID in logs
  - Enables tracing across distributed system
*/
package errors
