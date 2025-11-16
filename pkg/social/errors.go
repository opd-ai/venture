// Package social provides social interaction error types and utilities.
package social

import "fmt"

// ErrorType represents different types of social interaction errors.
type ErrorType int

const (
	// ErrorTypeRateLimit indicates rate limiting was triggered
	ErrorTypeRateLimit ErrorType = iota
	// ErrorTypeMuted indicates the user is temporarily muted
	ErrorTypeMuted
	// ErrorTypeNotSubscribed indicates the user is not subscribed to a channel
	ErrorTypeNotSubscribed
	// ErrorTypeProximity indicates the user is too far away
	ErrorTypeProximity
	// ErrorTypeTrust indicates insufficient trust score
	ErrorTypeTrust
	// ErrorTypeOwnership indicates item ownership issues
	ErrorTypeOwnership
	// ErrorTypeInventoryFull indicates inventory is full
	ErrorTypeInventoryFull
	// ErrorTypeTimeout indicates a timeout occurred
	ErrorTypeTimeout
	// ErrorTypeDisconnect indicates player disconnection
	ErrorTypeDisconnect
	// ErrorTypeInvalidImage indicates image validation failed
	ErrorTypeInvalidImage
	// ErrorTypeNetwork indicates a network error
	ErrorTypeNetwork
	// ErrorTypeUnknown indicates an unknown error
	ErrorTypeUnknown
)

// String returns the string representation of an ErrorType.
func (e ErrorType) String() string {
	switch e {
	case ErrorTypeRateLimit:
		return "RateLimit"
	case ErrorTypeMuted:
		return "Muted"
	case ErrorTypeNotSubscribed:
		return "NotSubscribed"
	case ErrorTypeProximity:
		return "Proximity"
	case ErrorTypeTrust:
		return "Trust"
	case ErrorTypeOwnership:
		return "Ownership"
	case ErrorTypeInventoryFull:
		return "InventoryFull"
	case ErrorTypeTimeout:
		return "Timeout"
	case ErrorTypeDisconnect:
		return "Disconnect"
	case ErrorTypeInvalidImage:
		return "InvalidImage"
	case ErrorTypeNetwork:
		return "Network"
	case ErrorTypeUnknown:
		return "Unknown"
	default:
		return "Unknown"
	}
}

// SocialError represents an error that occurred during social interaction.
type SocialError struct {
	Type    ErrorType
	Message string
	Context map[string]interface{}
}

// NewSocialError creates a new SocialError with the given type and message.
func NewSocialError(errType ErrorType, message string) *SocialError {
	return &SocialError{
		Type:    errType,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// Error implements the error interface.
func (e *SocialError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Type.String(), e.Message)
}

// WithContext adds context information to the error.
func (e *SocialError) WithContext(key string, value interface{}) *SocialError {
	e.Context[key] = value
	return e
}

// GetUserMessage returns a user-friendly error message.
func (e *SocialError) GetUserMessage() string {
	switch e.Type {
	case ErrorTypeRateLimit:
		return "You're sending messages too quickly. Please slow down."
	case ErrorTypeMuted:
		return "You are temporarily muted. Please wait before sending messages."
	case ErrorTypeNotSubscribed:
		return "You are not subscribed to this chat channel."
	case ErrorTypeProximity:
		return "You are too far away for this action."
	case ErrorTypeTrust:
		return "Your trust score is too low for this trade. Complete successful trades to increase trust."
	case ErrorTypeOwnership:
		return "You no longer own this item."
	case ErrorTypeInventoryFull:
		return "Your inventory is full. Make space before accepting this trade."
	case ErrorTypeTimeout:
		return "Request timed out. Please try again."
	case ErrorTypeDisconnect:
		return "The other player has disconnected."
	case ErrorTypeInvalidImage:
		return "Image validation failed. Check size, type, and dimensions."
	case ErrorTypeNetwork:
		return "Network error. Please check your connection."
	case ErrorTypeUnknown:
		return "An unknown error occurred. Please try again."
	default:
		return "An unknown error occurred. Please try again."
	}
}

// IsRetryable returns whether the error is retryable.
func (e *SocialError) IsRetryable() bool {
	switch e.Type {
	case ErrorTypeRateLimit, ErrorTypeNetwork, ErrorTypeTimeout:
		return true
	default:
		return false
	}
}

// Helper functions for creating specific error types

// ErrRateLimit creates a rate limit error.
func ErrRateLimit(channel string) *SocialError {
	return NewSocialError(ErrorTypeRateLimit, fmt.Sprintf("Rate limit exceeded for channel: %s", channel))
}

// ErrMuted creates a muted error.
func ErrMuted(until string) *SocialError {
	return NewSocialError(ErrorTypeMuted, fmt.Sprintf("You are muted until: %s", until))
}

// ErrNotSubscribed creates a not subscribed error.
func ErrNotSubscribed(channel string) *SocialError {
	return NewSocialError(ErrorTypeNotSubscribed, fmt.Sprintf("Not subscribed to channel: %s", channel))
}

// ErrProximity creates a proximity error.
func ErrProximity(required, actual float64) *SocialError {
	return NewSocialError(ErrorTypeProximity, fmt.Sprintf("Distance %.1f exceeds required %.1f", actual, required)).
		WithContext("required_distance", required).
		WithContext("actual_distance", actual)
}

// ErrTrust creates a trust error.
func ErrTrust(required, actual float64) *SocialError {
	return NewSocialError(ErrorTypeTrust, fmt.Sprintf("Trust %.2f below required %.2f", actual, required)).
		WithContext("required_trust", required).
		WithContext("actual_trust", actual)
}

// ErrOwnership creates an ownership error.
func ErrOwnership(itemID string) *SocialError {
	return NewSocialError(ErrorTypeOwnership, fmt.Sprintf("Item %s ownership invalid", itemID)).
		WithContext("item_id", itemID)
}

// ErrInventoryFull creates an inventory full error.
func ErrInventoryFull(required, available int) *SocialError {
	return NewSocialError(ErrorTypeInventoryFull, fmt.Sprintf("Need %d slots, only %d available", required, available)).
		WithContext("required_slots", required).
		WithContext("available_slots", available)
}

// ErrTimeout creates a timeout error.
func ErrTimeout(action string) *SocialError {
	return NewSocialError(ErrorTypeTimeout, fmt.Sprintf("Timeout during: %s", action)).
		WithContext("action", action)
}

// ErrDisconnect creates a disconnect error.
func ErrDisconnect(playerName string) *SocialError {
	return NewSocialError(ErrorTypeDisconnect, fmt.Sprintf("Player %s disconnected", playerName)).
		WithContext("player_name", playerName)
}

// ErrInvalidImage creates an invalid image error.
func ErrInvalidImage(reason string) *SocialError {
	return NewSocialError(ErrorTypeInvalidImage, fmt.Sprintf("Invalid image: %s", reason)).
		WithContext("reason", reason)
}

// ErrNetwork creates a network error.
func ErrNetwork(details string) *SocialError {
	return NewSocialError(ErrorTypeNetwork, fmt.Sprintf("Network error: %s", details)).
		WithContext("details", details)
}

// IsSocialError checks if an error is a SocialError and returns it if so.
func IsSocialError(err error) (*SocialError, bool) {
	if err == nil {
		return nil, false
	}
	if socialErr, ok := err.(*SocialError); ok {
		return socialErr, true
	}
	return nil, false
}
