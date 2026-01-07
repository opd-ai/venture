// Package validation provides input sanitization and validation utilities
// for securing user inputs in chat, trade, and other networked systems.
//
// # Security Features
//
// The validation package implements multiple layers of defense:
//  1. Input Sanitization: Remove/escape dangerous characters
//  2. Length Validation: Enforce reasonable limits on all inputs
//  3. Content Filtering: Block profanity and inappropriate content
//  4. Rate Limiting: Prevent spam and DoS attacks
//  5. Format Validation: Ensure IDs and structured data match expected patterns
//
// # Chat Message Validation
//
// Chat messages undergo multiple validation steps:
//
//	validator := validation.NewChatValidator()
//	if err := validator.ValidateMessage("Hello world!"); err != nil {
//	    // Handle invalid message
//	}
//	sanitized := validator.SanitizeMessage("Hello <script>alert(1)</script>")
//	// Returns: "Hello alert1"
//
// # Trade Validation
//
// Trade requests validate item IDs and counts:
//
//	validator := validation.NewTradeValidator()
//	if err := validator.ValidateItemIDs(itemIDs); err != nil {
//	    // Handle invalid item IDs
//	}
//	if err := validator.ValidateItemCount(len(items)); err != nil {
//	    // Handle too many items
//	}
//
// # Rate Limiting
//
// Per-client rate limiting prevents spam and DoS attacks:
//
//	limiter := validation.NewRateLimiter(10, time.Second) // 10 requests/second
//	if !limiter.Allow(clientID) {
//	    // Reject request - rate limit exceeded
//	}
//
// # Integration
//
// Validation should be applied at system boundaries before processing:
//  - Chat system: Validate all SendMessage calls
//  - Trade system: Validate ProposeTrade parameters
//  - Network layer: Rate limit all incoming requests
//
// # Performance
//
// All validation operations are designed for low latency:
//  - Chat validation: <1ms per message (regex-based)
//  - Item ID validation: <0.1ms per ID (format checking)
//  - Rate limiting: <0.01ms per check (map lookup)
//
// # Testing
//
// Comprehensive test coverage validates all security properties:
//  - Injection attack prevention
//  - Content filtering accuracy
//  - Rate limiter effectiveness
//  - Edge case handling (empty inputs, Unicode, etc.)
package validation
