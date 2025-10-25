// Package logging provides centralized structured logging configuration and utilities for Venture.package logging

// This package wraps logrus to provide consistent logging across all packages and commands.
// It supports environment-based configuration, multiple formatters, and contextual logging.
//
// # Configuration
//
// The logger can be configured via environment variables:
//   - LOG_LEVEL: Sets the minimum log level (debug, info, warn, error, fatal). Default: info
//   - LOG_FORMAT: Sets the output format (json, text). Default: text for development, json for production
//
// # Usage
//
// Initialize the logger at application startup:
//
//	logger := logging.NewLogger(logging.Config{
//	    Level:      logging.InfoLevel,
//	    Format:     logging.TextFormat,
//	    AddCaller:  true,
//	})
//
// Use structured fields for context:
//
//	logger.WithFields(logrus.Fields{
//	    "entityID": 12345,
//	    "componentType": "position",
//	}).Info("component added")
//
// # Performance
//
// Avoid logging in hot paths (game loop, rendering) above Info level.
// Use conditional debug logging for expensive operations:
//
//	if logger.GetLevel() >= logrus.DebugLevel {
//	    logger.WithFields(expensiveFields()).Debug("detailed state")
//	}
package logging
