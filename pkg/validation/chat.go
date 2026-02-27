package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	// MaxChatMessageLength is the maximum length of a chat message in characters
	MaxChatMessageLength = 500

	// MinChatMessageLength is the minimum length of a chat message in characters
	MinChatMessageLength = 1

	// MaxChatMessageBytes is the maximum size of a chat message in bytes.
	// This prevents oversized UTF-8 payloads (e.g., 500 4-byte emoji = 2000 bytes).
	// Enforced before rune-length check to avoid allocation overhead.
	MaxChatMessageBytes = 2000
)

var (
	// htmlTagPattern matches HTML-like tags for removal
	// Compiled once at package initialization for performance
	htmlTagPattern = regexp.MustCompile(`<[^>]*>`)

	// controlCharPattern matches control characters for removal
	// Compiled once at package initialization for performance
	controlCharPattern = regexp.MustCompile(`[\x00-\x1F\x7F]`)

	// urlPattern matches URLs for optional filtering
	// Compiled once at package initialization for performance
	urlPattern = regexp.MustCompile(`https?://[^\s]+`)
)

// ChatValidatorConfig provides configuration options for ChatValidator.
//
// This is intended for production deployments where profanity lists should be
// loaded from external configuration rather than hardcoded. The default
// NewChatValidator() uses a minimal stub list suitable for development only.
//
// Example production usage:
//
//	profanityWords, err := loadProfanityFromFile("config/profanity_en.txt")
//	if err != nil {
//	    return err
//	}
//	config := ChatValidatorConfig{
//	    CustomProfanityList: profanityWords,
//	}
//	validator := NewChatValidatorWithConfig(config)
type ChatValidatorConfig struct {
	// CustomProfanityList allows injection of production profanity filtering.
	// If nil or empty, falls back to buildProfanityList() stub.
	// Map keys are lowercase profanity words to filter.
	CustomProfanityList map[string]bool
}

// ChatValidator validates and sanitizes chat messages.
//
// For production use with custom profanity filtering, see NewChatValidatorWithConfig.
type ChatValidator struct {
	// profanityList contains words to filter out
	profanityList map[string]bool
}

// NewChatValidator creates a new chat validator with default settings.
//
// This uses a minimal stub profanity list suitable for development/testing only.
// For production deployments, use NewChatValidatorWithConfig with a custom
// profanity list loaded from configuration.
func NewChatValidator() *ChatValidator {
	return &ChatValidator{
		profanityList: buildProfanityList(),
	}
}

// NewChatValidatorWithConfig creates a new chat validator with custom configuration.
//
// This constructor allows production deployments to inject profanity lists loaded
// from external sources (files, databases, configuration services) rather than
// relying on the hardcoded stub in buildProfanityList().
//
// Example:
//
//	config := ChatValidatorConfig{
//	    CustomProfanityList: loadedProfanityMap,
//	}
//	validator := NewChatValidatorWithConfig(config)
func NewChatValidatorWithConfig(config ChatValidatorConfig) *ChatValidator {
	profanityList := config.CustomProfanityList
	if profanityList == nil || len(profanityList) == 0 {
		// Fall back to stub if no custom list provided
		profanityList = buildProfanityList()
	}
	return &ChatValidator{
		profanityList: profanityList,
	}
}

// ValidateMessage validates a chat message for length and content
// Returns an error if the message violates any validation rules
func (v *ChatValidator) ValidateMessage(message string) error {
	// Check for empty message
	if len(strings.TrimSpace(message)) == 0 {
		return fmt.Errorf("message cannot be empty")
	}

	// Check byte length first (no allocation, prevents oversized UTF-8 payloads)
	byteLen := len(message)
	if byteLen > MaxChatMessageBytes {
		return fmt.Errorf("message too large (maximum %d bytes, got %d)", MaxChatMessageBytes, byteLen)
	}

	// Count characters (Unicode-aware)
	charCount := utf8.RuneCountInString(message)

	// Check minimum length
	if charCount < MinChatMessageLength {
		return fmt.Errorf("message too short (minimum %d characters)", MinChatMessageLength)
	}

	// Check maximum length
	if charCount > MaxChatMessageLength {
		return fmt.Errorf("message too long (maximum %d characters, got %d)", MaxChatMessageLength, charCount)
	}

	// Check for profanity
	if v.containsProfanity(message) {
		return fmt.Errorf("message contains inappropriate content")
	}

	return nil
}

// SanitizeMessage removes dangerous characters and normalizes the message
// This should be called before storing or broadcasting a message
func (v *ChatValidator) SanitizeMessage(message string) string {
	// Remove HTML tags (prevent XSS-like attacks in UI rendering)
	sanitized := htmlTagPattern.ReplaceAllString(message, "")

	// Remove control characters (prevent terminal injection)
	sanitized = controlCharPattern.ReplaceAllString(sanitized, "")

	// Normalize whitespace (collapse multiple spaces)
	sanitized = strings.Join(strings.Fields(sanitized), " ")

	// Trim leading/trailing whitespace
	sanitized = strings.TrimSpace(sanitized)

	return sanitized
}

// SanitizeMessageWithURLFilter removes dangerous characters and optionally filters URLs.
// When filterURLs is true, URLs are replaced with "[link removed]".
// This is useful for channels where URL sharing is not permitted.
func (v *ChatValidator) SanitizeMessageWithURLFilter(message string, filterURLs bool) string {
	// Apply standard sanitization first
	sanitized := v.SanitizeMessage(message)

	// Optionally filter URLs
	if filterURLs {
		sanitized = urlPattern.ReplaceAllString(sanitized, "[link removed]")
	}

	return sanitized
}

// ContainsURL returns true if the message contains any URL patterns
func (v *ChatValidator) ContainsURL(message string) bool {
	return urlPattern.MatchString(message)
}

// ValidateAndSanitize combines validation and sanitization in one call
// Returns the sanitized message and any validation errors
func (v *ChatValidator) ValidateAndSanitize(message string) (string, error) {
	// Sanitize first
	sanitized := v.SanitizeMessage(message)

	// Then validate the sanitized version
	if err := v.ValidateMessage(sanitized); err != nil {
		return "", err
	}

	return sanitized, nil
}

// containsProfanity checks if the message contains any profane words.
// The check is case-insensitive and uses two strategies:
//
//  1. Word-by-word matching: Each word is checked against the profanity list
//     after stripping common punctuation (.,!?;:"').
//
//  2. Substring matching: The entire message is scanned for profane words as
//     substrings to catch l33tspeak bypasses (e.g., "mybadword1here").
//
// NOTE: The substring check intentionally has no word boundary awareness.
// This means words like "password" would be flagged if "ass" were in the list.
// Extending the profanity list should be done carefully to avoid false positives.
// Production systems should consider using word boundary regex patterns instead.
func (v *ChatValidator) containsProfanity(message string) bool {
	// Convert to lowercase for case-insensitive matching
	lower := strings.ToLower(message)

	// Split into words
	words := strings.Fields(lower)

	// Check each word against profanity list
	for _, word := range words {
		// Strip common punctuation
		cleanWord := strings.Trim(word, ".,!?;:\"'")
		if v.profanityList[cleanWord] {
			return true
		}
	}

	// Check for profanity embedded in words (l33tspeak bypass prevention)
	// This is a simple check - production systems might use more sophisticated filtering
	for profaneWord := range v.profanityList {
		if strings.Contains(lower, profaneWord) {
			return true
		}
	}

	return false
}

// buildProfanityList creates a map of words to filter.
//
// NOTE: This implementation is intentionally minimal and is provided only as an
// example. It MUST NOT be used as-is in production. Real deployments are
// expected to:
//   - Load a comprehensive, locale-appropriate profanity list from configuration
//     (for example: a JSON/YAML file, database table, or other config source).
//   - Keep the list outside of the binary so that it can be updated without
//     recompiling.
//   - Potentially maintain multiple lists per language/region and merge them
//     according to server configuration or shard settings.
//
// The returned map is used as a fast lookup structure by containsProfanity; the
// exact mechanism for loading the underlying words (file, DB, etc.) should be
// implemented by caller code that replaces this stub.
func buildProfanityList() map[string]bool {
	// Basic, non-exhaustive profanity list for demonstration purposes only.
	// This stub is intentionally small and generic; production systems should
	// inject a real list via configuration rather than relying on these values.
	words := []string{
		"badword1",
		"badword2",
		"offensive",
		// Example: extend or replace with configured values loaded at startup.
	}

	list := make(map[string]bool, len(words))
	for _, word := range words {
		list[word] = true
	}
	return list
}
