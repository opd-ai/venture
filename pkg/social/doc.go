// Package social provides error types and utilities for social interaction systems.
//
// This package defines structured error types that provide:
//   - User-friendly error messages for display in the UI
//   - Error categorization (rate limit, muted, trust, ownership, etc.)
//   - Contextual information for debugging and logging
//   - Retryability indicators for client-side retry logic
//
// # Error Types
//
// The [SocialError] type wraps all social interaction errors with rich metadata:
//
//	err := social.ErrTrust(0.8, 0.3)
//	err.GetUserMessage()  // "Your trust score is too low for this trade..."
//	err.IsRetryable()     // false
//
// # Helper Functions
//
// Convenience functions create typed errors with appropriate context:
//
//	social.ErrRateLimit("global")      // Rate limiting triggered
//	social.ErrMuted("10:30 AM")        // User is muted
//	social.ErrProximity(5.0, 12.5)     // Too far away
//	social.ErrTrust(0.8, 0.3)          // Insufficient trust
//	social.ErrOwnership("item-123")    // Item ownership issue
//	social.ErrInventoryFull(5, 2)      // Not enough inventory space
//	social.ErrTimeout("trade commit")  // Operation timed out
//	social.ErrDisconnect("player1")    // Player disconnected
//	social.ErrInvalidImage("too large")// Image validation failed
//	social.ErrNetwork("connection lost")// Network error
//
// # Integration with Engine Systems
//
// The engine's trade system, chat system, and other social systems use these
// error types to provide consistent error handling and user feedback:
//
//	if err := tradeSystem.ProposeTrade(...); err != nil {
//	    if socialErr, ok := social.IsSocialError(err); ok {
//	        displayMessage(socialErr.GetUserMessage())
//	        if socialErr.IsRetryable() {
//	            scheduleRetry()
//	        }
//	    }
//	}
//
// # Subpackages
//
// The [persistence] subpackage provides persistent storage for social data:
//   - Chat history with search and pagination
//   - Trust and reputation management
//   - Image gallery for shared screenshots
package social
