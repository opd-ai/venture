package validation

import (
	"fmt"
	"regexp"
)

const (
	// MaxTradeItems is the maximum number of items allowed in a single trade
	MaxTradeItems = 100

	// MinItemIDLength is the minimum length of a valid item ID
	MinItemIDLength = 1

	// MaxItemIDLength is the maximum length of a valid item ID
	MaxItemIDLength = 128
)

// itemIDPattern validates item ID format (alphanumeric + hyphens/underscores/equals)
// Compiled once at package initialization for performance
// Hyphen placed at end of character class for clarity
var itemIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_=-]+$`)

// TradeValidator validates trade-related inputs
type TradeValidator struct{}

// NewTradeValidator creates a new trade validator
func NewTradeValidator() *TradeValidator {
	return &TradeValidator{}
}

// ValidateItemIDs validates a list of item IDs
// Checks format, length, and duplicate detection
func (v *TradeValidator) ValidateItemIDs(itemIDs []string) error {
	if len(itemIDs) == 0 {
		return nil // Empty list is valid (e.g., trading gold only)
	}

	// Check for too many items
	if err := v.ValidateItemCount(len(itemIDs)); err != nil {
		return err
	}

	// Track seen IDs to detect duplicates
	seen := make(map[string]bool, len(itemIDs))

	for i, id := range itemIDs {
		// Check for duplicate
		if seen[id] {
			return fmt.Errorf("duplicate item ID at index %d: %s", i, id)
		}
		seen[id] = true

		// Validate individual ID
		if err := v.ValidateItemID(id); err != nil {
			return fmt.Errorf("invalid item ID at index %d: %w", i, err)
		}
	}

	return nil
}

// ValidateItemID validates a single item ID
func (v *TradeValidator) ValidateItemID(id string) error {
	// Check length
	if len(id) < MinItemIDLength {
		return fmt.Errorf("item ID too short (minimum %d characters)", MinItemIDLength)
	}

	if len(id) > MaxItemIDLength {
		return fmt.Errorf("item ID too long (maximum %d characters, got %d)", MaxItemIDLength, len(id))
	}

	// Check format (alphanumeric + allowed special chars)
	if !itemIDPattern.MatchString(id) {
		return fmt.Errorf("item ID contains invalid characters (allowed: a-z, A-Z, 0-9, -, _, =)")
	}

	return nil
}

// ValidateItemCount validates the number of items in a trade
func (v *TradeValidator) ValidateItemCount(count int) error {
	if count < 0 {
		return fmt.Errorf("item count cannot be negative")
	}

	if count > MaxTradeItems {
		return fmt.Errorf("too many items (maximum %d items, got %d)", MaxTradeItems, count)
	}

	return nil
}

// ValidateTradeQuantity validates the quantity of a single item in a trade.
// Rejects zero-quantity and negative-quantity trades.
func (v *TradeValidator) ValidateTradeQuantity(quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("trade quantity must be positive (got %d)", quantity)
	}
	return nil
}

// ValidateTradeRequest validates a complete trade request
// Checks both offered and requested item lists
func (v *TradeValidator) ValidateTradeRequest(offeredItems, requestedItems []string) error {
	// Validate offered items
	if err := v.ValidateItemIDs(offeredItems); err != nil {
		return fmt.Errorf("offered items validation failed: %w", err)
	}

	// Validate requested items
	if err := v.ValidateItemIDs(requestedItems); err != nil {
		return fmt.Errorf("requested items validation failed: %w", err)
	}

	// Check that at least one side has items (prevent empty trades)
	if len(offeredItems) == 0 && len(requestedItems) == 0 {
		return fmt.Errorf("trade must include at least one item")
	}

	return nil
}

// SanitizeItemID sanitizes an item ID by removing invalid characters
// Returns the sanitized ID - use ValidateItemID to check if sanitization changed the ID
func (v *TradeValidator) SanitizeItemID(id string) string {
	// Keep only valid characters
	result := make([]rune, 0, len(id))
	for _, r := range id {
		// Keep alphanumeric, hyphens, underscores, equals
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' || r == '=' {
			result = append(result, r)
		}
	}

	sanitized := string(result)

	// Enforce max length
	if len(sanitized) > MaxItemIDLength {
		sanitized = sanitized[:MaxItemIDLength]
	}

	return sanitized
}
