package errors

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestNewCorrelationID(t *testing.T) {
	// Generate multiple IDs and ensure they're unique
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := NewCorrelationID()
		if id == "" {
			t.Error("NewCorrelationID() returned empty string")
		}
		if ids[id] {
			t.Errorf("NewCorrelationID() generated duplicate ID: %s", id)
		}
		ids[id] = true

		// Check UUID format (basic validation)
		parts := strings.Split(id, "-")
		if len(parts) != 5 {
			t.Errorf("NewCorrelationID() = %s, not in UUID format", id)
		}
	}
}

func TestNewSequentialCorrelationID(t *testing.T) {
	// Generate multiple sequential IDs
	ids := make(map[string]bool)
	for i := 0; i < 10; i++ {
		id := NewSequentialCorrelationID()
		if id == "" {
			t.Error("NewSequentialCorrelationID() returned empty string")
		}
		if ids[id] {
			t.Errorf("NewSequentialCorrelationID() generated duplicate ID: %s", id)
		}
		ids[id] = true
	}
}

func TestWithCorrelationID(t *testing.T) {
	ctx := context.Background()
	testID := "test-correlation-123"

	ctx = WithCorrelationID(ctx, testID)
	retrievedID := GetCorrelationID(ctx)

	if retrievedID != testID {
		t.Errorf("GetCorrelationID() = %s, want %s", retrievedID, testID)
	}
}

func TestGetCorrelationID_NoID(t *testing.T) {
	ctx := context.Background()
	id := GetCorrelationID(ctx)

	if id != "" {
		t.Errorf("GetCorrelationID() = %s, want empty string", id)
	}
}

func TestGetOrCreateCorrelationID(t *testing.T) {
	t.Run("existing ID", func(t *testing.T) {
		ctx := context.Background()
		testID := "test-existing-id"
		ctx = WithCorrelationID(ctx, testID)

		id := GetOrCreateCorrelationID(ctx)
		if id != testID {
			t.Errorf("GetOrCreateCorrelationID() = %s, want %s", id, testID)
		}
	})

	t.Run("create new ID", func(t *testing.T) {
		ctx := context.Background()
		id := GetOrCreateCorrelationID(ctx)

		if id == "" {
			t.Error("GetOrCreateCorrelationID() returned empty string")
		}
	})
}

func TestWrapWithContext(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		ctx := context.Background()
		err := WrapWithContext(ctx, nil, ErrorTypeNetwork, "test")
		if err != nil {
			t.Errorf("WrapWithContext(nil) = %v, want nil", err)
		}
	})

	t.Run("standard error with correlation ID", func(t *testing.T) {
		ctx := context.Background()
		testID := "test-correlation-456"
		ctx = WithCorrelationID(ctx, testID)

		baseErr := fmt.Errorf("base error")
		ventureErr := WrapWithContext(ctx, baseErr, ErrorTypeNetwork, "wrapped error")

		if ventureErr.CorrelationID != testID {
			t.Errorf("CorrelationID = %s, want %s", ventureErr.CorrelationID, testID)
		}
		if ventureErr.Type != ErrorTypeNetwork {
			t.Errorf("Type = %v, want %v", ventureErr.Type, ErrorTypeNetwork)
		}
		if !errors.Is(ventureErr, baseErr) {
			t.Error("WrapWithContext should preserve error chain")
		}
	})

	t.Run("standard error without correlation ID", func(t *testing.T) {
		ctx := context.Background()
		baseErr := fmt.Errorf("base error")
		ventureErr := WrapWithContext(ctx, baseErr, ErrorTypeNetwork, "wrapped error")

		if ventureErr.CorrelationID != "" {
			t.Errorf("CorrelationID = %s, want empty string", ventureErr.CorrelationID)
		}
	})

	t.Run("VentureError with existing correlation ID", func(t *testing.T) {
		ctx := context.Background()
		newID := "new-correlation-id"
		ctx = WithCorrelationID(ctx, newID)

		existingID := "existing-correlation-id"
		baseErr := Network("test error").WithCorrelationID(existingID)

		ventureErr := WrapWithContext(ctx, baseErr, ErrorTypeTimeout, "wrapped")

		// Should preserve existing correlation ID
		if ventureErr.CorrelationID != existingID {
			t.Errorf("CorrelationID = %s, want %s (should preserve existing)", ventureErr.CorrelationID, existingID)
		}

		// Should create a new wrapper with the specified type and message
		if ventureErr.Type != ErrorTypeTimeout {
			t.Errorf("Type = %v, want %v", ventureErr.Type, ErrorTypeTimeout)
		}
		if ventureErr.Message != "wrapped" {
			t.Errorf("Message = %s, want wrapped", ventureErr.Message)
		}

		// Should preserve error chain
		if !errors.Is(ventureErr, baseErr) {
			t.Error("Error chain should be preserved")
		}
	})

	t.Run("VentureError without correlation ID", func(t *testing.T) {
		ctx := context.Background()
		testID := "test-correlation-789"
		ctx = WithCorrelationID(ctx, testID)

		baseErr := Network("test error") // No correlation ID
		ventureErr := WrapWithContext(ctx, baseErr, ErrorTypeTimeout, "wrapped")

		// Should add correlation ID from context
		if ventureErr.CorrelationID != testID {
			t.Errorf("CorrelationID = %s, want %s", ventureErr.CorrelationID, testID)
		}

		// Should create a new wrapper with the specified type and message
		if ventureErr.Type != ErrorTypeTimeout {
			t.Errorf("Type = %v, want %v", ventureErr.Type, ErrorTypeTimeout)
		}
		if ventureErr.Message != "wrapped" {
			t.Errorf("Message = %s, want wrapped", ventureErr.Message)
		}

		// Should preserve error chain
		if !errors.Is(ventureErr, baseErr) {
			t.Error("Error chain should be preserved")
		}
	})
}

func TestNewWithContext(t *testing.T) {
	t.Run("with correlation ID", func(t *testing.T) {
		ctx := context.Background()
		testID := "test-new-with-context"
		ctx = WithCorrelationID(ctx, testID)

		err := NewWithContext(ctx, ErrorTypeValidation, "validation error")

		if err.CorrelationID != testID {
			t.Errorf("CorrelationID = %s, want %s", err.CorrelationID, testID)
		}
		if err.Type != ErrorTypeValidation {
			t.Errorf("Type = %v, want %v", err.Type, ErrorTypeValidation)
		}
		if err.Message != "validation error" {
			t.Errorf("Message = %s, want validation error", err.Message)
		}
	})

	t.Run("without correlation ID", func(t *testing.T) {
		ctx := context.Background()
		err := NewWithContext(ctx, ErrorTypeValidation, "validation error")

		if err.CorrelationID != "" {
			t.Errorf("CorrelationID = %s, want empty string", err.CorrelationID)
		}
	})
}

func TestCorrelationID_Integration(t *testing.T) {
	// Simulate a request flow with correlation ID propagation
	ctx := context.Background()
	correlationID := NewCorrelationID()
	ctx = WithCorrelationID(ctx, correlationID)

	// Simulate multiple layers of error wrapping
	err1 := NewWithContext(ctx, ErrorTypeDatabase, "database connection failed")
	if err1.CorrelationID != correlationID {
		t.Error("First error should have correlation ID")
	}

	err2 := WrapWithContext(ctx, err1, ErrorTypeNetwork, "network timeout during database operation")
	if err2.CorrelationID != correlationID {
		t.Error("Wrapped error should preserve correlation ID")
	}

	// Extract correlation ID from error
	if ventureErr, ok := AsVentureError(err2); ok {
		if ventureErr.CorrelationID != correlationID {
			t.Errorf("Extracted correlation ID = %s, want %s", ventureErr.CorrelationID, correlationID)
		}
	} else {
		t.Error("Should be able to extract VentureError")
	}
}

func TestCorrelationID_Concurrency(t *testing.T) {
	// Test that correlation IDs are unique even when generated concurrently
	const numGoroutines = 100
	const idsPerGoroutine = 100

	ids := make(chan string, numGoroutines*idsPerGoroutine)
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < idsPerGoroutine; j++ {
				ids <- NewCorrelationID()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	close(ids)

	// Check for uniqueness
	idMap := make(map[string]bool)
	for id := range ids {
		if idMap[id] {
			t.Errorf("Duplicate correlation ID generated: %s", id)
		}
		idMap[id] = true
	}

	expectedCount := numGoroutines * idsPerGoroutine
	if len(idMap) != expectedCount {
		t.Errorf("Generated %d unique IDs, want %d", len(idMap), expectedCount)
	}
}
