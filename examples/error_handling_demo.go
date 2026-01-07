// Package main demonstrates comprehensive error handling with correlation IDs
// as specified in PLAN.md Phase 2.
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/opd-ai/venture/pkg/errors"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logging.NewLogger(logging.Config{
		Level:       logging.InfoLevel,
		Format:      logging.JSONFormat,
		AddCaller:   true,
		EnableColor: false,
	})

	logger.Info("Error handling demonstration started")

	// Example 1: Basic error creation
	demonstrateBasicErrors(logger)

	// Example 2: Error wrapping with context
	demonstrateErrorWrapping(logger)

	// Example 3: Correlation IDs for distributed tracing
	demonstrateCorrelationIDs(logger)

	// Example 4: User-friendly messages
	demonstrateUserMessages(logger)

	// Example 5: Retryability
	demonstrateRetryability(logger)

	logger.Info("Error handling demonstration completed")
}

func demonstrateBasicErrors(logger *logrus.Logger) {
	logger.Info("=== Example 1: Basic Error Creation ===")

	// Network error
	networkErr := errors.Network("connection timeout to game server")
	logging.LogError(logger, networkErr, "network operation failed")

	// Validation error
	validationErr := errors.Validation("username must be 3-20 characters")
	logging.LogError(logger, validationErr, "validation failed")

	// Configuration error
	configErr := errors.Configuration("invalid port: must be 1024-65535")
	logging.LogError(logger, configErr, "configuration error")

	fmt.Println()
}

func demonstrateErrorWrapping(logger *logrus.Logger) {
	logger.Info("=== Example 2: Error Wrapping with Context ===")

	// Simulate a function that returns a standard error
	baseErr := fmt.Errorf("connection refused")

	// Wrap with VentureError and add context
	enrichedErr := errors.NetworkWrap(baseErr, "failed to connect to database").
		WithContext("host", "db.example.com").
		WithContext("port", 5432).
		WithContext("retry_count", 3)

	logging.LogError(logger, enrichedErr, "database connection failed")

	fmt.Println()
}

func demonstrateCorrelationIDs(logger *logrus.Logger) {
	logger.Info("=== Example 3: Correlation IDs for Distributed Tracing ===")

	// Simulate a request handler
	ctx := context.Background()
	correlationID := errors.NewCorrelationID()
	ctx = errors.WithCorrelationID(ctx, correlationID)

	logger.WithField("correlation_id", correlationID).Info("processing request")

	// Simulate operations that create errors
	if err := simulateNetworkOperation(ctx); err != nil {
		logging.ErrorLogger(logger, err).Error("network operation failed in request handler")
	}

	if err := simulateValidation(ctx); err != nil {
		logging.ErrorLogger(logger, err).Error("validation failed in request handler")
	}

	fmt.Println()
}

func simulateNetworkOperation(ctx context.Context) error {
	// Simulate a network error with correlation ID from context
	return errors.NewWithContext(ctx, errors.ErrorTypeNetwork, "connection timeout").
		WithContext("host", "game-server.example.com").
		WithContext("timeout", "5s")
}

func simulateValidation(ctx context.Context) error {
	// Simulate validation error with correlation ID from context
	return errors.NewWithContext(ctx, errors.ErrorTypeValidation, "invalid player name").
		WithContext("field", "player_name").
		WithContext("value", "a").
		WithContext("min_length", 3)
}

func demonstrateUserMessages(logger *logrus.Logger) {
	logger.Info("=== Example 4: User-Friendly Messages ===")

	// Create error with both technical and user-friendly messages
	err := errors.Network("TCP connection refused on 127.0.0.1:8080: dial tcp: connection refused")
	err.UserMessage = "Cannot connect to the game server. Please check your network connection and try again."

	// Technical message for logs
	logger.WithField("technical_error", err.Error()).Info("logging technical details")

	// User-friendly message for UI
	fmt.Printf("Display to user: %s\n", err.GetUserMessage())

	// Default user-friendly messages based on error type
	timeoutErr := errors.Timeout("operation exceeded 30s deadline")
	fmt.Printf("Timeout user message: %s\n", timeoutErr.GetUserMessage())

	rateLimitErr := errors.RateLimit("exceeded 100 requests per minute")
	fmt.Printf("Rate limit user message: %s\n", rateLimitErr.GetUserMessage())

	fmt.Println()
}

func demonstrateRetryability(logger *logrus.Logger) {
	logger.Info("=== Example 5: Retryability ===")

	// Retryable errors
	networkErr := errors.Network("temporary connection failure")
	fmt.Printf("Network error retryable: %v\n", networkErr.IsRetryable())

	timeoutErr := errors.Timeout("operation timed out")
	fmt.Printf("Timeout error retryable: %v\n", timeoutErr.IsRetryable())

	// Non-retryable errors
	validationErr := errors.Validation("invalid input format")
	fmt.Printf("Validation error retryable: %v\n", validationErr.IsRetryable())

	configErr := errors.Configuration("missing required field")
	fmt.Printf("Configuration error retryable: %v\n", configErr.IsRetryable())

	// Simulate retry logic
	err := simulateRetryableOperation()
	if err != nil {
		if ventureErr, ok := errors.AsVentureError(err); ok && ventureErr.IsRetryable() {
			logger.Warn("operation failed but is retryable, scheduling retry")
			time.Sleep(100 * time.Millisecond) // Simulate backoff
			// Retry operation...
		} else {
			logger.Error("operation failed with non-retryable error")
		}
	}

	fmt.Println()
}

func simulateRetryableOperation() error {
	// Simulate a retryable network error
	return errors.Network("connection pool exhausted").
		WithContext("available_connections", 0).
		WithContext("max_connections", 10)
}

// Helper to simulate output redirection for demonstration
func init() {
	// Keep stdout for demonstration messages
	// Logrus will write to stdout with JSON format
}
