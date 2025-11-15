package social

import (
	"fmt"
	"testing"
)

func TestErrorTypeString(t *testing.T) {
	tests := []struct {
		errType  ErrorType
		expected string
	}{
		{ErrorTypeRateLimit, "RateLimit"},
		{ErrorTypeMuted, "Muted"},
		{ErrorTypeNotSubscribed, "NotSubscribed"},
		{ErrorTypeProximity, "Proximity"},
		{ErrorTypeTrust, "Trust"},
		{ErrorTypeOwnership, "Ownership"},
		{ErrorTypeInventoryFull, "InventoryFull"},
		{ErrorTypeTimeout, "Timeout"},
		{ErrorTypeDisconnect, "Disconnect"},
		{ErrorTypeInvalidImage, "InvalidImage"},
		{ErrorTypeNetwork, "Network"},
		{ErrorTypeUnknown, "Unknown"},
		{ErrorType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.errType.String()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestNewSocialError(t *testing.T) {
	err := NewSocialError(ErrorTypeRateLimit, "Test message")

	if err.Type != ErrorTypeRateLimit {
		t.Errorf("Expected type RateLimit, got %v", err.Type)
	}
	if err.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", err.Message)
	}
	if err.Context == nil {
		t.Error("Expected Context to be initialized")
	}
}

func TestSocialErrorError(t *testing.T) {
	err := NewSocialError(ErrorTypeProximity, "Too far away")

	expected := "[Proximity] Too far away"
	result := err.Error()

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestWithContext(t *testing.T) {
	err := NewSocialError(ErrorTypeProximity, "Test").
		WithContext("distance", 15.5).
		WithContext("max_distance", 10.0)

	if len(err.Context) != 2 {
		t.Errorf("Expected 2 context entries, got %d", len(err.Context))
	}

	distance, ok := err.Context["distance"].(float64)
	if !ok || distance != 15.5 {
		t.Errorf("Expected distance 15.5, got %v", distance)
	}

	maxDist, ok := err.Context["max_distance"].(float64)
	if !ok || maxDist != 10.0 {
		t.Errorf("Expected max_distance 10.0, got %v", maxDist)
	}
}

func TestGetUserMessage(t *testing.T) {
	tests := []struct {
		errType ErrorType
		message string
	}{
		{ErrorTypeRateLimit, "You're sending messages too quickly. Please slow down."},
		{ErrorTypeMuted, "You are temporarily muted. Please wait before sending messages."},
		{ErrorTypeNotSubscribed, "You are not subscribed to this chat channel."},
		{ErrorTypeProximity, "You are too far away for this action."},
		{ErrorTypeTrust, "Your trust score is too low for this trade. Complete successful trades to increase trust."},
		{ErrorTypeOwnership, "You no longer own this item."},
		{ErrorTypeInventoryFull, "Your inventory is full. Make space before accepting this trade."},
		{ErrorTypeTimeout, "Request timed out. Please try again."},
		{ErrorTypeDisconnect, "The other player has disconnected."},
		{ErrorTypeInvalidImage, "Image validation failed. Check size, type, and dimensions."},
		{ErrorTypeNetwork, "Network error. Please check your connection."},
		{ErrorTypeUnknown, "An unknown error occurred. Please try again."},
	}

	for _, tt := range tests {
		t.Run(tt.errType.String(), func(t *testing.T) {
			err := NewSocialError(tt.errType, "Internal message")
			userMsg := err.GetUserMessage()
			if userMsg != tt.message {
				t.Errorf("Expected '%s', got '%s'", tt.message, userMsg)
			}
		})
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		errType   ErrorType
		retryable bool
	}{
		{ErrorTypeRateLimit, true},
		{ErrorTypeNetwork, true},
		{ErrorTypeTimeout, true},
		{ErrorTypeMuted, false},
		{ErrorTypeProximity, false},
		{ErrorTypeTrust, false},
		{ErrorTypeOwnership, false},
	}

	for _, tt := range tests {
		t.Run(tt.errType.String(), func(t *testing.T) {
			err := NewSocialError(tt.errType, "Test")
			result := err.IsRetryable()
			if result != tt.retryable {
				t.Errorf("Expected retryable=%v, got %v", tt.retryable, result)
			}
		})
	}
}

func TestErrRateLimit(t *testing.T) {
	err := ErrRateLimit("Global")

	if err.Type != ErrorTypeRateLimit {
		t.Errorf("Expected type RateLimit, got %v", err.Type)
	}
	if err.Message != "Rate limit exceeded for channel: Global" {
		t.Errorf("Unexpected message: %s", err.Message)
	}
}

func TestErrMuted(t *testing.T) {
	err := ErrMuted("2024-01-01 12:00:00")

	if err.Type != ErrorTypeMuted {
		t.Errorf("Expected type Muted, got %v", err.Type)
	}
	expectedMsg := "You are muted until: 2024-01-01 12:00:00"
	if err.Message != expectedMsg {
		t.Errorf("Expected message '%s', got '%s'", expectedMsg, err.Message)
	}
}

func TestErrNotSubscribed(t *testing.T) {
	err := ErrNotSubscribed("Party")

	if err.Type != ErrorTypeNotSubscribed {
		t.Errorf("Expected type NotSubscribed, got %v", err.Type)
	}
}

func TestErrProximity(t *testing.T) {
	err := ErrProximity(10.0, 15.5)

	if err.Type != ErrorTypeProximity {
		t.Errorf("Expected type Proximity, got %v", err.Type)
	}

	required, ok := err.Context["required_distance"].(float64)
	if !ok || required != 10.0 {
		t.Errorf("Expected required_distance 10.0, got %v", required)
	}

	actual, ok := err.Context["actual_distance"].(float64)
	if !ok || actual != 15.5 {
		t.Errorf("Expected actual_distance 15.5, got %v", actual)
	}
}

func TestErrTrust(t *testing.T) {
	err := ErrTrust(0.8, 0.3)

	if err.Type != ErrorTypeTrust {
		t.Errorf("Expected type Trust, got %v", err.Type)
	}

	required, ok := err.Context["required_trust"].(float64)
	if !ok || required != 0.8 {
		t.Errorf("Expected required_trust 0.8, got %v", required)
	}

	actual, ok := err.Context["actual_trust"].(float64)
	if !ok || actual != 0.3 {
		t.Errorf("Expected actual_trust 0.3, got %v", actual)
	}
}

func TestErrOwnership(t *testing.T) {
	err := ErrOwnership("item123")

	if err.Type != ErrorTypeOwnership {
		t.Errorf("Expected type Ownership, got %v", err.Type)
	}

	itemID, ok := err.Context["item_id"].(string)
	if !ok || itemID != "item123" {
		t.Errorf("Expected item_id 'item123', got %v", itemID)
	}
}

func TestErrInventoryFull(t *testing.T) {
	err := ErrInventoryFull(5, 2)

	if err.Type != ErrorTypeInventoryFull {
		t.Errorf("Expected type InventoryFull, got %v", err.Type)
	}

	required, ok := err.Context["required_slots"].(int)
	if !ok || required != 5 {
		t.Errorf("Expected required_slots 5, got %v", required)
	}

	available, ok := err.Context["available_slots"].(int)
	if !ok || available != 2 {
		t.Errorf("Expected available_slots 2, got %v", available)
	}
}

func TestErrTimeout(t *testing.T) {
	err := ErrTimeout("trade proposal")

	if err.Type != ErrorTypeTimeout {
		t.Errorf("Expected type Timeout, got %v", err.Type)
	}

	action, ok := err.Context["action"].(string)
	if !ok || action != "trade proposal" {
		t.Errorf("Expected action 'trade proposal', got %v", action)
	}
}

func TestErrDisconnect(t *testing.T) {
	err := ErrDisconnect("Alice")

	if err.Type != ErrorTypeDisconnect {
		t.Errorf("Expected type Disconnect, got %v", err.Type)
	}

	playerName, ok := err.Context["player_name"].(string)
	if !ok || playerName != "Alice" {
		t.Errorf("Expected player_name 'Alice', got %v", playerName)
	}
}

func TestErrInvalidImage(t *testing.T) {
	err := ErrInvalidImage("file too large")

	if err.Type != ErrorTypeInvalidImage {
		t.Errorf("Expected type InvalidImage, got %v", err.Type)
	}

	reason, ok := err.Context["reason"].(string)
	if !ok || reason != "file too large" {
		t.Errorf("Expected reason 'file too large', got %v", reason)
	}
}

func TestErrNetwork(t *testing.T) {
	err := ErrNetwork("connection timeout")

	if err.Type != ErrorTypeNetwork {
		t.Errorf("Expected type Network, got %v", err.Type)
	}

	details, ok := err.Context["details"].(string)
	if !ok || details != "connection timeout" {
		t.Errorf("Expected details 'connection timeout', got %v", details)
	}
}

func TestIsSocialError(t *testing.T) {
	// Test with SocialError
	socialErr := NewSocialError(ErrorTypeProximity, "Test")
	result, ok := IsSocialError(socialErr)
	if !ok {
		t.Error("Expected IsSocialError to return true for SocialError")
	}
	if result != socialErr {
		t.Error("Expected IsSocialError to return the same error")
	}

	// Test with standard error
	stdErr := fmt.Errorf("standard error")
	result, ok = IsSocialError(stdErr)
	if ok {
		t.Error("Expected IsSocialError to return false for standard error")
	}
	if result != nil {
		t.Error("Expected IsSocialError to return nil for standard error")
	}

	// Test with nil
	result, ok = IsSocialError(nil)
	if ok {
		t.Error("Expected IsSocialError to return false for nil")
	}
	if result != nil {
		t.Error("Expected IsSocialError to return nil for nil input")
	}
}

// Benchmarks

func BenchmarkNewSocialError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewSocialError(ErrorTypeProximity, "Test message")
	}
}

func BenchmarkWithContext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewSocialError(ErrorTypeProximity, "Test").
			WithContext("key1", "value1").
			WithContext("key2", 42)
	}
}

func BenchmarkGetUserMessage(b *testing.B) {
	err := NewSocialError(ErrorTypeProximity, "Test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.GetUserMessage()
	}
}

func BenchmarkIsRetryable(b *testing.B) {
	err := NewSocialError(ErrorTypeNetwork, "Test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.IsRetryable()
	}
}

func BenchmarkIsSocialError(b *testing.B) {
	socialErr := NewSocialError(ErrorTypeProximity, "Test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = IsSocialError(socialErr)
	}
}
