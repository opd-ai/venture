// Package recovery provides panic recovery utilities for production stability.
package recovery

import (
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

// LogPanicAndCleanup logs a recovered panic and executes cleanup. This should be
// called from a deferred function after calling recover(). This is a helper that
// does the logging and cleanup, but does NOT call recover() itself.
//
// Example usage:
//
//	go func() {
//	    defer func() {
//	        if r := recover(); r != nil {
//	            LogPanicAndCleanup(logger, "worker goroutine", r, nil)
//	        }
//	    }()
//	    // goroutine work...
//	}()
//
// Example with cleanup:
//
//	go func() {
//	    defer func() {
//	        if r := recover(); r != nil {
//	            LogPanicAndCleanup(logger, "client handler", r, func() {
//	                disconnectClient(clientID)
//	            })
//	        }
//	    }()
//	    // handle client...
//	}()
func LogPanicAndCleanup(logger *logrus.Entry, context string, panicValue interface{}, cleanup func()) {
	// Get stack trace
	stack := debug.Stack()

	// Log the panic with context and stack trace
	if logger != nil {
		logger.WithFields(logrus.Fields{
			"panic":      panicValue,
			"context":    context,
			"stack":      string(stack),
			"error_type": "panic",
		}).Error("Goroutine panic recovered")
	} else {
		// Fallback to standard logrus if no logger provided
		logrus.WithFields(logrus.Fields{
			"panic":      panicValue,
			"context":    context,
			"stack":      string(stack),
			"error_type": "panic",
		}).Error("Goroutine panic recovered")
	}

	// Execute cleanup if provided
	if cleanup != nil {
		// Protect against panics in cleanup function
		defer func() {
			if cleanupPanic := recover(); cleanupPanic != nil {
				if logger != nil {
					logger.WithFields(logrus.Fields{
						"panic":      cleanupPanic,
						"context":    fmt.Sprintf("%s cleanup", context),
						"error_type": "panic_in_cleanup",
					}).Error("Panic during cleanup function")
				} else {
					logrus.WithFields(logrus.Fields{
						"panic":      cleanupPanic,
						"context":    fmt.Sprintf("%s cleanup", context),
						"error_type": "panic_in_cleanup",
					}).Error("Panic during cleanup function")
				}
			}
		}()
		cleanup()
	}
}

// RecoverPanic is a convenience wrapper that combines recover() with logging and cleanup.
// This returns a function that should be deferred. This is the recommended way to add
// panic recovery to goroutines as it properly calls recover() in the deferred context.
//
// Example usage:
//
//	go func() {
//	    defer RecoverPanic(logger, "worker goroutine", nil)
//	    // goroutine work...
//	}()
//
// Example with cleanup:
//
//	go func() {
//	    defer RecoverPanic(logger, "client handler", func() {
//	        disconnectClient(clientID)
//	    })
//	    // handle client...
//	}()
func RecoverPanic(logger *logrus.Entry, context string, cleanup func()) func() {
	return func() {
		if r := recover(); r != nil {
			LogPanicAndCleanup(logger, context, r, cleanup)
		}
	}
}

// RecoverPanicWithLogger is a convenience wrapper that creates a logger entry
// with the given component name and returns a function that should be deferred.
//
// Example usage:
//
//	go func() {
//	    defer RecoverPanicWithLogger("network_server", "accept loop", nil)
//	    // goroutine work...
//	}()
func RecoverPanicWithLogger(component string, context string, cleanup func()) func() {
	logger := logrus.WithField("component", component)
	return RecoverPanic(logger, context, cleanup)
}

