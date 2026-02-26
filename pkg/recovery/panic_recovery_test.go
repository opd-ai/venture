package recovery

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
)

// TestRecoverPanic verifies panic recovery with logging
func TestRecoverPanic(t *testing.T) {
	tests := []struct {
		name          string
		panicValue    interface{}
		context       string
		withCleanup   bool
		expectCleanup bool
	}{
		{
			name:          "string panic",
			panicValue:    "test panic",
			context:       "test goroutine",
			withCleanup:   false,
			expectCleanup: false,
		},
		{
			name:          "error panic",
			panicValue:    &testError{msg: "test error"},
			context:       "error handler",
			withCleanup:   true,
			expectCleanup: true,
		},
		{
			name:          "nil panic",
			panicValue:    nil,
			context:       "nil handler",
			withCleanup:   false,
			expectCleanup: false,
		},
		{
			name:          "int panic",
			panicValue:    42,
			context:       "int handler",
			withCleanup:   true,
			expectCleanup: true,
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
				}
			}

			// Create a channel to signal test completion
			done := make(chan bool, 1)

			// Execute panic in goroutine
			go func() {
				defer func() { done <- true }()
				defer RecoverPanic(entry, tt.context, cleanup)()
				if tt.panicValue != nil {
					panic(tt.panicValue)
				}
			}()

			// Wait for goroutine completion
			<-done

			// Verify cleanup was called if expected
			if tt.expectCleanup && !cleanupCalled {
				t.Error("Expected cleanup to be called but it wasn't")
			}
			if !tt.expectCleanup && cleanupCalled {
				t.Error("Expected cleanup not to be called but it was")
			}

			// Verify log output only if panic occurred
			if tt.panicValue != nil {
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
				if !strings.Contains(logOutput, "error_type") {
					t.Error("Expected log to contain 'error_type' field")
				}
			}
		})
	}
}

// TestRecoverPanicWithLogger verifies convenience wrapper
func TestRecoverPanicWithLogger(t *testing.T) {
	// Capture logrus output
	var buf bytes.Buffer
	oldOutput := logrus.StandardLogger().Out
	oldFormatter := logrus.StandardLogger().Formatter
	logrus.SetOutput(&buf)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	defer func() {
		logrus.SetOutput(oldOutput)
		logrus.SetFormatter(oldFormatter)
	}()

	component := "test_component"
	context := "test_context"

	done := make(chan bool, 1)
	go func() {
		defer func() { done <- true }()
		defer RecoverPanicWithLogger(component, context, nil)()
		panic("test panic")
	}()

	<-done

	logOutput := buf.String()
	if logOutput == "" {
		t.Fatal("Expected log output but got none")
	}

	if !strings.Contains(logOutput, component) {
		t.Errorf("Expected log to contain component %q", component)
	}
	if !strings.Contains(logOutput, context) {
		t.Errorf("Expected log to contain context %q", context)
	}
}

// TestRecoverPanicCleanupPanic verifies panic in cleanup is handled
func TestRecoverPanicCleanupPanic(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.JSONFormatter{})
	entry := logger.WithField("test", "cleanup_panic")

	done := make(chan bool, 1)
	go func() {
		defer func() { done <- true }()
		defer RecoverPanic(entry, "test context", func() {
			panic("cleanup panic")
		})()
		panic("original panic")
	}()

	<-done

	logOutput := buf.String()
	if logOutput == "" {
		t.Fatal("Expected log output but got none")
	}

	// Should log both the original panic and the cleanup panic
	lines := strings.Split(strings.TrimSpace(logOutput), "\n")
	if len(lines) < 2 {
		t.Errorf("Expected at least 2 log entries (original + cleanup panic), got %d", len(lines))
	}

	// Verify original panic is logged
	if !strings.Contains(logOutput, "original panic") {
		t.Error("Expected log to contain original panic message")
	}

	// Verify cleanup panic is logged
	if !strings.Contains(logOutput, "cleanup panic") {
		t.Error("Expected log to contain cleanup panic message")
	}

	if !strings.Contains(logOutput, "panic_in_cleanup") {
		t.Error("Expected log to contain panic_in_cleanup error type")
	}
}

// TestRecoverPanicNoPanic verifies no-op when no panic occurs
func TestRecoverPanicNoPanic(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	entry := logger.WithField("test", "no_panic")

	cleanupCalled := false
	done := make(chan bool, 1)

	go func() {
		defer func() { done <- true }()
		defer RecoverPanic(entry, "test context", func() {
			cleanupCalled = true
		})()
		// No panic occurs
	}()

	<-done

	// Cleanup should not be called if no panic
	if cleanupCalled {
		t.Error("Expected cleanup not to be called when no panic occurs")
	}

	// No log output expected
	if buf.Len() > 0 {
		t.Errorf("Expected no log output but got: %s", buf.String())
	}
}

// TestRecoverPanicConcurrent verifies panic recovery under concurrent load
func TestRecoverPanicConcurrent(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.JSONFormatter{})
	entry := logger.WithField("test", "concurrent")

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			defer RecoverPanic(entry, "concurrent test", nil)()
			if id%2 == 0 {
				panic("even panic")
			}
			// Odd goroutines complete normally
		}(i)
	}

	wg.Wait()

	// Verify approximately half the goroutines logged panics
	logOutput := buf.String()
	panicCount := strings.Count(logOutput, "even panic")
	expectedCount := goroutines / 2

	// Allow small variance due to timing
	if panicCount < expectedCount-5 || panicCount > expectedCount+5 {
		t.Errorf("Expected approximately %d panic logs, got %d", expectedCount, panicCount)
	}
}

// TestRecoverPanicNilLogger verifies fallback to default logger
func TestRecoverPanicNilLogger(t *testing.T) {
	// Capture default logrus output
	var buf bytes.Buffer
	oldOutput := logrus.StandardLogger().Out
	oldFormatter := logrus.StandardLogger().Formatter
	logrus.SetOutput(&buf)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	defer func() {
		logrus.SetOutput(oldOutput)
		logrus.SetFormatter(oldFormatter)
	}()

	done := make(chan bool, 1)
	go func() {
		defer func() { done <- true }()
		defer RecoverPanic(nil, "nil logger test", nil)()
		panic("test with nil logger")
	}()

	<-done

	logOutput := buf.String()
	if logOutput == "" {
		t.Fatal("Expected log output from default logger but got none")
	}

	if !strings.Contains(logOutput, "nil logger test") {
		t.Error("Expected log to contain context")
	}
}

// testError is a custom error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// BenchmarkRecoverPanic_NoPanic measures overhead when no panic occurs (hot path)
func BenchmarkRecoverPanic_NoPanic(b *testing.B) {
	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})
	entry := logger.WithField("component", "benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		func() {
			defer RecoverPanic(entry, "benchmark context", nil)()
			// Normal execution path
		}()
	}
}

// BenchmarkRecoverPanic_WithCleanup measures overhead with cleanup function
func BenchmarkRecoverPanic_WithCleanup(b *testing.B) {
	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})
	entry := logger.WithField("component", "benchmark")

	cleanup := func() {
		// Minimal cleanup work
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		func() {
			defer RecoverPanic(entry, "benchmark context", cleanup)()
			// Normal execution path
		}()
	}
}

// BenchmarkRecoverPanic_WithPanic measures recovery overhead when panic occurs
func BenchmarkRecoverPanic_WithPanic(b *testing.B) {
	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})
	entry := logger.WithField("component", "benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		func() {
			defer RecoverPanic(entry, "benchmark context", nil)()
			panic("benchmark panic")
		}()
	}
}

// BenchmarkRecoverPanic_WithPanicAndCleanup measures full recovery path
func BenchmarkRecoverPanic_WithPanicAndCleanup(b *testing.B) {
	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})
	entry := logger.WithField("component", "benchmark")

	cleanup := func() {
		// Minimal cleanup work
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		func() {
			defer RecoverPanic(entry, "benchmark context", cleanup)()
			panic("benchmark panic")
		}()
	}
}

// BenchmarkRecoverPanicWithLogger measures convenience wrapper overhead
func BenchmarkRecoverPanicWithLogger(b *testing.B) {
	oldOutput := logrus.StandardLogger().Out
	logrus.SetOutput(&bytes.Buffer{})
	defer logrus.SetOutput(oldOutput)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		func() {
			defer RecoverPanicWithLogger("component", "context", nil)()
			// Normal execution path
		}()
	}
}

// BenchmarkLogPanicAndCleanup_Direct measures direct logging function
func BenchmarkLogPanicAndCleanup_Direct(b *testing.B) {
	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})
	entry := logger.WithField("component", "benchmark")

	cleanup := func() {
		// Minimal cleanup work
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LogPanicAndCleanup(entry, "benchmark context", "test panic", cleanup)
	}
}
