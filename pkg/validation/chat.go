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
)

// ChatValidator validates and sanitizes chat messages
type ChatValidator struct {
	// profanityList contains words to filter out
	profanityList map[string]bool

	// htmlTagPattern matches HTML-like tags for removal
	htmlTagPattern *regexp.Regexp

	// controlCharPattern matches control characters for removal
	controlCharPattern *regexp.Regexp

	// urlPattern matches URLs for optional filtering
	urlPattern *regexp.Regexp
}

// NewChatValidator creates a new chat validator with default settings
func NewChatValidator() *ChatValidator {
	return &ChatValidator{
		profanityList:      buildProfanityList(),
		htmlTagPattern:     regexp.MustCompile(`<[^>]*>`),
		controlCharPattern: regexp.MustCompile(`[\x00-\x1F\x7F]`),
		urlPattern:         regexp.MustCompile(`https?://[^\s]+`),
	}
}

// ValidateMessage validates a chat message for length and content
// Returns an error if the message violates any validation rules
func (v *ChatValidator) ValidateMessage(message string) error {
	// Check for empty message
	if len(strings.TrimSpace(message)) == 0 {
		return fmt.Errorf("message cannot be empty")
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
	sanitized := v.htmlTagPattern.ReplaceAllString(message, "")

	// Remove control characters (prevent terminal injection)
	sanitized = v.controlCharPattern.ReplaceAllString(sanitized, "")

	// Normalize whitespace (collapse multiple spaces)
	sanitized = strings.Join(strings.Fields(sanitized), " ")

	// Trim leading/trailing whitespace
	sanitized = strings.TrimSpace(sanitized)

	return sanitized
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

// containsProfanity checks if the message contains any profane words
// Uses case-insensitive matching with word boundaries
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

// buildProfanityList creates a map of words to filter
// In production, this would be loaded from a configuration file or database
func buildProfanityList() map[string]bool {
	// Basic profanity list for demonstration
	// Production systems should use a comprehensive, configurable list
	words := []string{
		"badword1",
		"badword2",
		"offensive",
		// Add more words as needed
	}

	list := make(map[string]bool, len(words))
	for _, word := range words {
		list[word] = true
	}
	return list
}

// AllowURLs configures whether URLs are allowed in messages
// By default, URLs are allowed. Call with false to strip them.
func (v *ChatValidator) AllowURLs(allow bool) {
	// This is a placeholder for URL filtering configuration
	// Implementation would modify SanitizeMessage behavior
}
