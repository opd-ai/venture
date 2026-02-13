// Package recovery provides panic recovery utilities for production stability.
//
// This package provides functions to safely recover from panics in goroutines,
// log them with structured context, and optionally execute cleanup functions.
// All recovery functions use structured logging with logrus.WithFields for
// consistent error tracking and debugging.
//
// # Basic Usage
//
// The simplest way to add panic recovery to a goroutine is with RecoverPanic:
//
//	go func() {
//	    defer recovery.RecoverPanic(logger, "worker goroutine", nil)()
//	    // goroutine work that may panic...
//	}()
//
// # Recovery with Cleanup
//
// For cases where cleanup is needed after a panic:
//
//	go func() {
//	    defer recovery.RecoverPanic(logger, "client handler", func() {
//	        disconnectClient(clientID)
//	    })()
//	    // handle client...
//	}()
//
// The cleanup function is itself protected against panics, so a panic during
// cleanup will be logged but won't cause an unrecoverable crash.
//
// # Component-Based Recovery
//
// For goroutines without an existing logger, use RecoverPanicWithLogger which
// creates a logger with the component name:
//
//	go func() {
//	    defer recovery.RecoverPanicWithLogger("network_server", "accept loop", nil)()
//	    // accept connections...
//	}()
//
// # Integration
//
// This package is used throughout the codebase:
//   - engine/character_creation.go: UI dialog panic recovery
//   - engine/performance/network_batcher.go: network batch loop safety
//   - engine/performance/cache_and_lod.go: background loader worker goroutines
//   - engine/mod_browser_system.go: mod download operations
//
// # Structured Logging Fields
//
// All panic logs include these standard fields:
//   - panic: the recovered panic value
//   - context: human-readable context string
//   - stack: full stack trace (when available)
//   - error_type: "panic" or "panic_in_cleanup"
package recovery
