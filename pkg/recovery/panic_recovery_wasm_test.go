//go:build js && wasm
// +build js,wasm

package recovery

import (
	"bytes"
	"strings"
	"syscall/js"
	"testing"

	"github.com/sirupsen/logrus"
)

// TestRecoverPanic_WASMContext verifies panic recovery works in WASM/browser environment.
// This test specifically validates that:
// 1. Panic recovery works with syscall/js operations
// 2. Cleanup functions execute properly in WASM context
// 3. Browser-specific panic scenarios are handled correctly
func TestRecoverPanic_WASMContext(t *testing.T) {
	tests := []struct {
		name          string
		panicFunc     func()
		context       string
		withCleanup   bool
		expectCleanup bool
		description   string
	}{
		{
			name: "js.Global panic",
			panicFunc: func() {
				// Access undefined property (common WASM error)
				js.Global().Get("nonexistentProperty").Call("nonexistentMethod")
			},
			context:       "js.Global access",
			withCleanup:   true,
			expectCleanup: true,
			description:   "Browser API access that doesn't exist",
		},
		{
			name: "localStorage undefined panic",
			panicFunc: func() {
				// Simulate localStorage being undefined (private browsing mode)
				localStorage := js.Global().Get("localStorage")
				if localStorage.IsUndefined() {
					panic("localStorage is undefined")
				}
			},
			context:       "localStorage check",
			withCleanup:   true,
			expectCleanup: true,
			description:   "Storage API unavailable in private mode",
		},
		{
			name: "js.ValueOf panic",
			panicFunc: func() {
				// js.ValueOf with nil can cause issues
				var nilPtr *string
				_ = js.ValueOf(nilPtr)
			},
			context:       "js.ValueOf nil",
			withCleanup:   false,
			expectCleanup: false,
			description:   "Converting nil Go value to JS",
		},
		{
			name: "callback panic",
			panicFunc: func() {
				// Simulate panic in JS callback
				cb := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
					panic("callback panic")
				})
				defer cb.Release()
				// Note: We can't actually invoke this without browser DOM,
				// so we just panic directly to simulate callback panic
				panic("simulated callback panic")
			},
			context:       "JS callback handler",
			withCleanup:   true,
			expectCleanup: true,
			description:   "Panic during browser callback execution",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup logger with buffer to capture output
			var buf bytes.Buffer
			logger := logrus.New()
			logger.SetOutput(&buf)
			logger.SetFormatter(&logrus.JSONFormatter{})
			entry := logger.WithField("test", tt.name)

			// Track cleanup execution
			var cleanupCalled bool
			var cleanup func()
			if tt.withCleanup {
				cleanup = func() {
					cleanupCalled = true
					// Simulate WASM-specific cleanup (e.g., releasing JS funcs)
					// In real code, this might be: jsFuncRef.Release()
				}
			}

			// Create a channel to signal test completion
			done := make(chan bool, 1)

			// Execute panic in goroutine (simulates async WASM operation)
			go func() {
				defer func() { done <- true }()
				defer RecoverPanic(entry, tt.context, cleanup)()
				tt.panicFunc()
			}()

			// Wait for goroutine completion
			<-done

			// Verify cleanup was called if expected
			if tt.expectCleanup && !cleanupCalled {
				t.Errorf("Expected cleanup to be called but it wasn't (scenario: %s)", tt.description)
			}
			if !tt.expectCleanup && cleanupCalled {
				t.Errorf("Expected cleanup not to be called but it was (scenario: %s)", tt.description)
			}

			// Verify log output
			logOutput := buf.String()
			if logOutput == "" {
				t.Error("Expected log output but got none")
			}

			// Verify log contains expected fields
			if !strings.Contains(logOutput, tt.context) {
				t.Errorf("Expected log to contain context %q", tt.context)
			}
			if !strings.Contains(logOutput, "panic") {
				t.Error("Expected log to contain 'panic' field")
			}
			if !strings.Contains(logOutput, "stack") {
				t.Error("Expected log to contain 'stack' field")
			}
		})
	}
}

// TestRecoverPanic_WASMJSFuncCleanup verifies proper cleanup of js.Func references.
// This is critical in WASM to prevent memory leaks of callback handles.
func TestRecoverPanic_WASMJSFuncCleanup(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.JSONFormatter{})
	entry := logger.WithField("test", "jsfunc_cleanup")

	// Track if cleanup released the js.Func
	var funcReleased bool
	var testFunc js.Func

	// Create a js.Func (common pattern in WASM event handlers)
	testFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// This would normally do something with browser APIs
		return nil
	})

	done := make(chan bool, 1)
	go func() {
		defer func() { done <- true }()
		defer RecoverPanic(entry, "event handler", func() {
			testFunc.Release()
			funcReleased = true
		})()
		// Simulate panic during event handling
		panic("event handler panic")
	}()

	<-done

	// Verify js.Func was properly released
	if !funcReleased {
		t.Error("Expected js.Func to be released in cleanup but it wasn't")
	}

	// Verify log output
	logOutput := buf.String()
	if logOutput == "" {
		t.Fatal("Expected log output but got none")
	}

	if !strings.Contains(logOutput, "event handler panic") {
		t.Error("Expected log to contain panic message")
	}
}

// TestRecoverPanic_WASMBrowserAPIPanic verifies recovery from browser API panics.
// Tests common WASM panic scenarios like quota exceeded, permission denied, etc.
func TestRecoverPanic_WASMBrowserAPIPanic(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.JSONFormatter{})
	entry := logger.WithField("test", "browser_api_panic")

	// Simulate localStorage quota exceeded (common WASM issue)
	done := make(chan bool, 1)
	go func() {
		defer func() { done <- true }()
		defer RecoverPanic(entry, "localStorage.setItem", func() {
			// Cleanup might involve clearing cache
			// In real code: clearCache(), fallbackToMemory()
		})()

		// Simulate quota exceeded error (browsers throw DOMException)
		// In real WASM code, this would be: localStorage.setItem(key, largeData)
		panic("QuotaExceededError: Failed to execute 'setItem' on 'Storage'")
	}()

	<-done

	logOutput := buf.String()
	if logOutput == "" {
		t.Fatal("Expected log output but got none")
	}

	if !strings.Contains(logOutput, "QuotaExceededError") {
		t.Error("Expected log to contain quota exceeded error")
	}
	if !strings.Contains(logOutput, "localStorage.setItem") {
		t.Error("Expected log to contain context")
	}
}

// TestRecoverPanic_WASMAsyncOperation verifies panic recovery during async WASM operations.
// WASM often uses promises and async callbacks; this validates recovery in that context.
func TestRecoverPanic_WASMAsyncOperation(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.JSONFormatter{})
	entry := logger.WithField("test", "async_operation")

	// Track cleanup of async operation
	var asyncCleanupCalled bool

	done := make(chan bool, 1)
	go func() {
		defer func() { done <- true }()
		defer RecoverPanic(entry, "async fetch operation", func() {
			asyncCleanupCalled = true
			// In real code: cancel fetch, close connections, etc.
		})()

		// Simulate panic during async operation (e.g., fetch, WebSocket)
		// In real WASM: js.Global().Call("fetch", url).Then(...)
		panic("TypeError: Failed to fetch")
	}()

	<-done

	if !asyncCleanupCalled {
		t.Error("Expected async cleanup to be called")
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "Failed to fetch") {
		t.Error("Expected log to contain fetch error")
	}
}

// TestRecoverPanic_WASMNoPanicAsyncContext verifies no-op behavior in WASM async context.
// Ensures recovery mechanism doesn't interfere with normal async operations.
func TestRecoverPanic_WASMNoPanicAsyncContext(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	entry := logger.WithField("test", "no_panic_async")

	cleanupCalled := false
	done := make(chan bool, 1)

	go func() {
		defer func() { done <- true }()
		defer RecoverPanic(entry, "async operation", func() {
			cleanupCalled = true
		})()

		// Simulate successful async operation
		_ = js.ValueOf("success")
		// No panic occurs
	}()

	<-done

	// Cleanup should not be called when no panic
	if cleanupCalled {
		t.Error("Expected cleanup not to be called when no panic occurs")
	}

	// No log output expected
	if buf.Len() > 0 {
		t.Errorf("Expected no log output but got: %s", buf.String())
	}
}

// BenchmarkRecoverPanic_WASMContext measures overhead in WASM environment.
// WASM may have different performance characteristics than native code.
func BenchmarkRecoverPanic_WASMContext(b *testing.B) {
	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})
	entry := logger.WithField("component", "wasm_benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		func() {
			defer RecoverPanic(entry, "wasm operation", nil)()
			// Normal execution path (common case in WASM event handlers)
			_ = js.ValueOf(i)
		}()
	}
}

// BenchmarkRecoverPanic_WASMWithJSFunc measures overhead with js.Func cleanup.
// This is the most common pattern in WASM: event handlers that need cleanup.
func BenchmarkRecoverPanic_WASMWithJSFunc(b *testing.B) {
	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})
	entry := logger.WithField("component", "wasm_benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		func() {
			testFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				return nil
			})

			defer RecoverPanic(entry, "event handler", func() {
				testFunc.Release()
			})()

			// Normal execution
		}()
	}
}
