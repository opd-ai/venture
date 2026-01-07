package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestErrorType_String(t *testing.T) {
	tests := []struct {
		name     string
		errType  ErrorType
		expected string
	}{
		{"Unknown", ErrorTypeUnknown, "Unknown"},
		{"Network", ErrorTypeNetwork, "Network"},
		{"Validation", ErrorTypeValidation, "Validation"},
		{"Configuration", ErrorTypeConfiguration, "Configuration"},
		{"Generation", ErrorTypeGeneration, "Generation"},
		{"Serialization", ErrorTypeSerialization, "Serialization"},
		{"FileSystem", ErrorTypeFileSystem, "FileSystem"},
		{"Database", ErrorTypeDatabase, "Database"},
		{"Authentication", ErrorTypeAuthentication, "Authentication"},
		{"RateLimit", ErrorTypeRateLimit, "RateLimit"},
		{"Concurrency", ErrorTypeConcurrency, "Concurrency"},
		{"Resource", ErrorTypeResource, "Resource"},
		{"Timeout", ErrorTypeTimeout, "Timeout"},
		{"Invalid", ErrorType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.errType.String(); got != tt.expected {
				t.Errorf("ErrorType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestVentureError_Error(t *testing.T) {
	tests := []struct {
		name      string
		err       *VentureError
		wantParts []string // Parts that should be in the error string
	}{
		{
			name: "basic error",
			err: &VentureError{
				Type:    ErrorTypeNetwork,
				Message: "connection failed",
			},
			wantParts: []string{"Network", "connection failed"},
		},
		{
			name: "error with correlation ID",
			err: &VentureError{
				Type:          ErrorTypeValidation,
				Message:       "invalid input",
				CorrelationID: "test-123",
			},
			wantParts: []string{"Validation", "test-123", "invalid input"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			for _, part := range tt.wantParts {
				if !strings.Contains(got, part) {
					t.Errorf("VentureError.Error() = %v, missing part %v", got, part)
				}
			}
		})
	}
}

func TestVentureError_Unwrap(t *testing.T) {
	baseErr := fmt.Errorf("base error")
	ventureErr := &VentureError{
		Type:    ErrorTypeNetwork,
		Message: "wrapped",
		Err:     baseErr,
	}

	if unwrapped := ventureErr.Unwrap(); unwrapped != baseErr {
		t.Errorf("VentureError.Unwrap() = %v, want %v", unwrapped, baseErr)
	}
}

func TestVentureError_WithContext(t *testing.T) {
	err := New(ErrorTypeNetwork, "test error")
	err.WithContext("key1", "value1").WithContext("key2", 42)

	if err.Context["key1"] != "value1" {
		t.Errorf("Context[key1] = %v, want value1", err.Context["key1"])
	}
	if err.Context["key2"] != 42 {
		t.Errorf("Context[key2] = %v, want 42", err.Context["key2"])
	}
}

func TestVentureError_WithCorrelationID(t *testing.T) {
	err := New(ErrorTypeNetwork, "test error")
	err.WithCorrelationID("test-correlation-id")

	if err.CorrelationID != "test-correlation-id" {
		t.Errorf("CorrelationID = %v, want test-correlation-id", err.CorrelationID)
	}
}

func TestVentureError_GetUserMessage(t *testing.T) {
	tests := []struct {
		name        string
		err         *VentureError
		wantContain string
	}{
		{
			name: "custom user message",
			err: &VentureError{
				Type:        ErrorTypeNetwork,
				Message:     "technical details",
				UserMessage: "Custom friendly message",
			},
			wantContain: "Custom friendly message",
		},
		{
			name: "network default",
			err: &VentureError{
				Type:    ErrorTypeNetwork,
				Message: "technical details",
			},
			wantContain: "Network error",
		},
		{
			name: "validation default",
			err: &VentureError{
				Type:    ErrorTypeValidation,
				Message: "technical details",
			},
			wantContain: "Invalid input",
		},
		{
			name: "timeout default",
			err: &VentureError{
				Type:    ErrorTypeTimeout,
				Message: "technical details",
			},
			wantContain: "timed out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.GetUserMessage()
			if !strings.Contains(got, tt.wantContain) {
				t.Errorf("GetUserMessage() = %v, want to contain %v", got, tt.wantContain)
			}
		})
	}
}

func TestVentureError_IsRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  *VentureError
		want bool
	}{
		{
			name: "retryable set true",
			err:  &VentureError{Type: ErrorTypeNetwork, Retryable: true},
			want: true,
		},
		{
			name: "retryable set false",
			err:  &VentureError{Type: ErrorTypeValidation, Retryable: false},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.IsRetryable(); got != tt.want {
				t.Errorf("IsRetryable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	err := New(ErrorTypeNetwork, "test message")
	if err.Type != ErrorTypeNetwork {
		t.Errorf("New() Type = %v, want %v", err.Type, ErrorTypeNetwork)
	}
	if err.Message != "test message" {
		t.Errorf("New() Message = %v, want test message", err.Message)
	}
	if err.Context == nil {
		t.Error("New() Context is nil, want initialized map")
	}
}

func TestWrap(t *testing.T) {
	baseErr := fmt.Errorf("base error")
	err := Wrap(baseErr, ErrorTypeNetwork, "wrapped")

	if err.Type != ErrorTypeNetwork {
		t.Errorf("Wrap() Type = %v, want %v", err.Type, ErrorTypeNetwork)
	}
	if err.Message != "wrapped" {
		t.Errorf("Wrap() Message = %v, want wrapped", err.Message)
	}
	if err.Err != baseErr {
		t.Errorf("Wrap() Err = %v, want %v", err.Err, baseErr)
	}

	// Test wrapping nil
	if nilErr := Wrap(nil, ErrorTypeNetwork, "test"); nilErr != nil {
		t.Errorf("Wrap(nil) = %v, want nil", nilErr)
	}
}

func TestWrapf(t *testing.T) {
	baseErr := fmt.Errorf("base error")
	err := Wrapf(baseErr, ErrorTypeNetwork, "error code %d: %s", 500, "server error")

	expected := "error code 500: server error"
	if err.Message != expected {
		t.Errorf("Wrapf() Message = %v, want %v", err.Message, expected)
	}

	// Test wrapping nil
	if nilErr := Wrapf(nil, ErrorTypeNetwork, "test %s", "message"); nilErr != nil {
		t.Errorf("Wrapf(nil) = %v, want nil", nilErr)
	}
}

func TestIs(t *testing.T) {
	networkErr := Network("test")
	validationErr := Validation("test")

	if !Is(networkErr, ErrorTypeNetwork) {
		t.Error("Is() should return true for matching error type")
	}
	if Is(networkErr, ErrorTypeValidation) {
		t.Error("Is() should return false for non-matching error type")
	}
	if Is(validationErr, ErrorTypeNetwork) {
		t.Error("Is() should return false for different error type")
	}

	// Test with non-VentureError
	stdErr := fmt.Errorf("standard error")
	if Is(stdErr, ErrorTypeNetwork) {
		t.Error("Is() should return false for standard error")
	}
}

func TestAsVentureError(t *testing.T) {
	ventureErr := Network("test")

	// Test with VentureError
	if got, ok := AsVentureError(ventureErr); !ok || got != ventureErr {
		t.Errorf("AsVentureError() = %v, %v, want %v, true", got, ok, ventureErr)
	}

	// Test with standard error
	stdErr := fmt.Errorf("standard error")
	if got, ok := AsVentureError(stdErr); ok || got != nil {
		t.Errorf("AsVentureError(stdErr) = %v, %v, want nil, false", got, ok)
	}

	// Test with nil
	if got, ok := AsVentureError(nil); ok || got != nil {
		t.Errorf("AsVentureError(nil) = %v, %v, want nil, false", got, ok)
	}
}

func TestHelperFunctions(t *testing.T) {
	tests := []struct {
		name      string
		createErr func() *VentureError
		wantType  ErrorType
		retryable bool
	}{
		{"Network", func() *VentureError { return Network("test") }, ErrorTypeNetwork, true},
		{"Validation", func() *VentureError { return Validation("test") }, ErrorTypeValidation, false},
		{"Configuration", func() *VentureError { return Configuration("test") }, ErrorTypeConfiguration, false},
		{"Timeout", func() *VentureError { return Timeout("test") }, ErrorTypeTimeout, true},
		{"Serialization", func() *VentureError { return Serialization("test") }, ErrorTypeSerialization, false},
		{"Generation", func() *VentureError { return Generation("test") }, ErrorTypeGeneration, false},
		{"Database", func() *VentureError { return Database("test") }, ErrorTypeDatabase, true},
		{"RateLimit", func() *VentureError { return RateLimit("test") }, ErrorTypeRateLimit, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.createErr()
			if err.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", err.Type, tt.wantType)
			}
			if err.Retryable != tt.retryable {
				t.Errorf("Retryable = %v, want %v", err.Retryable, tt.retryable)
			}
			if err.Message != "test" {
				t.Errorf("Message = %v, want test", err.Message)
			}
		})
	}
}

func TestWrapHelperFunctions(t *testing.T) {
	baseErr := fmt.Errorf("base error")

	tests := []struct {
		name      string
		wrapErr   func() *VentureError
		wantType  ErrorType
		retryable bool
	}{
		{"NetworkWrap", func() *VentureError { return NetworkWrap(baseErr, "wrapped") }, ErrorTypeNetwork, true},
		{"ValidationWrap", func() *VentureError { return ValidationWrap(baseErr, "wrapped") }, ErrorTypeValidation, false},
		{"ConfigurationWrap", func() *VentureError { return ConfigurationWrap(baseErr, "wrapped") }, ErrorTypeConfiguration, false},
		{"TimeoutWrap", func() *VentureError { return TimeoutWrap(baseErr, "wrapped") }, ErrorTypeTimeout, true},
		{"SerializationWrap", func() *VentureError { return SerializationWrap(baseErr, "wrapped") }, ErrorTypeSerialization, false},
		{"GenerationWrap", func() *VentureError { return GenerationWrap(baseErr, "wrapped") }, ErrorTypeGeneration, false},
		{"DatabaseWrap", func() *VentureError { return DatabaseWrap(baseErr, "wrapped") }, ErrorTypeDatabase, true},
		{"RateLimitWrap", func() *VentureError { return RateLimitWrap(baseErr, "wrapped") }, ErrorTypeRateLimit, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.wrapErr()
			if err.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", err.Type, tt.wantType)
			}
			if err.Retryable != tt.retryable {
				t.Errorf("Retryable = %v, want %v", err.Retryable, tt.retryable)
			}
			if err.Message != "wrapped" {
				t.Errorf("Message = %v, want wrapped", err.Message)
			}
			if !errors.Is(err, baseErr) {
				t.Error("Wrapped error should support errors.Is")
			}
		})
	}
}

func TestWrapHelperFunctions_NilError(t *testing.T) {
	// All wrap functions should return nil when wrapping nil
	wrapFuncs := []func(error, string) *VentureError{
		NetworkWrap,
		ValidationWrap,
		ConfigurationWrap,
		TimeoutWrap,
		SerializationWrap,
		GenerationWrap,
		DatabaseWrap,
		RateLimitWrap,
	}

	for i, wrapFunc := range wrapFuncs {
		t.Run(fmt.Sprintf("WrapFunc_%d", i), func(t *testing.T) {
			if err := wrapFunc(nil, "test"); err != nil {
				t.Errorf("Wrap(nil) = %v, want nil", err)
			}
		})
	}
}

func TestErrorChaining(t *testing.T) {
	// Create an error chain
	baseErr := fmt.Errorf("io error")
	networkErr := NetworkWrap(baseErr, "connection failed")
	wrappedErr := Wrap(networkErr, ErrorTypeTimeout, "operation timed out")

	// Test errors.Is
	if !errors.Is(wrappedErr, baseErr) {
		t.Error("Error chain should support errors.Is through multiple levels")
	}

	// Test errors.As
	var ventureErr *VentureError
	if !errors.As(wrappedErr, &ventureErr) {
		t.Error("Error chain should support errors.As")
	}
}
