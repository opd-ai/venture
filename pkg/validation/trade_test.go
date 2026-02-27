package validation

import (
	"fmt"
	"strings"
	"testing"
)

func TestTradeValidator_ValidateItemID(t *testing.T) {
	validator := NewTradeValidator()

	tests := []struct {
		name    string
		itemID  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid alphanumeric ID",
			itemID:  "item123",
			wantErr: false,
		},
		{
			name:    "valid base64 ID",
			itemID:  "dGVzdC1pdGVtLWlk",
			wantErr: false,
		},
		{
			name:    "valid with hyphens",
			itemID:  "item-123-abc",
			wantErr: false,
		},
		{
			name:    "valid with underscores",
			itemID:  "item_123_abc",
			wantErr: false,
		},
		{
			name:    "valid with equals (base64 padding)",
			itemID:  "base64id==",
			wantErr: false,
		},
		{
			name:    "empty ID",
			itemID:  "",
			wantErr: true,
			errMsg:  "too short",
		},
		{
			name:    "too long ID",
			itemID:  strings.Repeat("a", MaxItemIDLength+1),
			wantErr: true,
			errMsg:  "too long",
		},
		{
			name:    "max length ID",
			itemID:  strings.Repeat("a", MaxItemIDLength),
			wantErr: false,
		},
		{
			name:    "invalid characters spaces",
			itemID:  "item 123",
			wantErr: true,
			errMsg:  "invalid characters",
		},
		{
			name:    "invalid characters slashes",
			itemID:  "item/123",
			wantErr: true,
			errMsg:  "invalid characters",
		},
		{
			name:    "invalid characters brackets",
			itemID:  "item[123]",
			wantErr: true,
			errMsg:  "invalid characters",
		},
		{
			name:    "invalid characters special",
			itemID:  "item@123#",
			wantErr: true,
			errMsg:  "invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateItemID(tt.itemID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateItemID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateItemID() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestTradeValidator_ValidateItemIDs(t *testing.T) {
	validator := NewTradeValidator()

	tests := []struct {
		name    string
		itemIDs []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty list",
			itemIDs: []string{},
			wantErr: false,
		},
		{
			name:    "single valid ID",
			itemIDs: []string{"item123"},
			wantErr: false,
		},
		{
			name:    "multiple valid IDs",
			itemIDs: []string{"item1", "item2", "item3"},
			wantErr: false,
		},
		{
			name:    "duplicate IDs",
			itemIDs: []string{"item1", "item2", "item1"},
			wantErr: true,
			errMsg:  "duplicate",
		},
		{
			name:    "invalid ID in list",
			itemIDs: []string{"item1", "invalid id", "item3"},
			wantErr: true,
			errMsg:  "invalid",
		},
		{
			name:    "too many items",
			itemIDs: make([]string, MaxTradeItems+1),
			wantErr: true,
			errMsg:  "too many",
		},
		{
			name:    "max items",
			itemIDs: make([]string, MaxTradeItems),
			wantErr: false,
		},
	}

	// Fill arrays with valid IDs for max items test
	for i := range tests {
		if tests[i].name == "too many items" || tests[i].name == "max items" {
			for j := range tests[i].itemIDs {
				// Create unique IDs using sprintf to avoid duplicates
				tests[i].itemIDs[j] = fmt.Sprintf("item-%d", j)
			}
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateItemIDs(tt.itemIDs)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateItemIDs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateItemIDs() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestTradeValidator_ValidateItemCount(t *testing.T) {
	validator := NewTradeValidator()

	tests := []struct {
		name    string
		count   int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "zero items",
			count:   0,
			wantErr: false,
		},
		{
			name:    "one item",
			count:   1,
			wantErr: false,
		},
		{
			name:    "max items",
			count:   MaxTradeItems,
			wantErr: false,
		},
		{
			name:    "too many items",
			count:   MaxTradeItems + 1,
			wantErr: true,
			errMsg:  "too many",
		},
		{
			name:    "negative count",
			count:   -1,
			wantErr: true,
			errMsg:  "negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateItemCount(tt.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateItemCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateItemCount() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestTradeValidator_ValidateTradeRequest(t *testing.T) {
	validator := NewTradeValidator()

	tests := []struct {
		name           string
		offeredItems   []string
		requestedItems []string
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "valid trade",
			offeredItems:   []string{"item1", "item2"},
			requestedItems: []string{"item3"},
			wantErr:        false,
		},
		{
			name:           "one-sided trade offer",
			offeredItems:   []string{"item1"},
			requestedItems: []string{},
			wantErr:        false,
		},
		{
			name:           "one-sided trade request",
			offeredItems:   []string{},
			requestedItems: []string{"item1"},
			wantErr:        false,
		},
		{
			name:           "empty trade",
			offeredItems:   []string{},
			requestedItems: []string{},
			wantErr:        true,
			errMsg:         "at least one item",
		},
		{
			name:           "invalid offered items",
			offeredItems:   []string{"item 1"},
			requestedItems: []string{"item2"},
			wantErr:        true,
			errMsg:         "offered items",
		},
		{
			name:           "invalid requested items",
			offeredItems:   []string{"item1"},
			requestedItems: []string{"item 2"},
			wantErr:        true,
			errMsg:         "requested items",
		},
		{
			name:           "duplicate in offered",
			offeredItems:   []string{"item1", "item1"},
			requestedItems: []string{"item2"},
			wantErr:        true,
			errMsg:         "offered items",
		},
		{
			name:           "duplicate in requested",
			offeredItems:   []string{"item1"},
			requestedItems: []string{"item2", "item2"},
			wantErr:        true,
			errMsg:         "requested items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTradeRequest(tt.offeredItems, tt.requestedItems)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTradeRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateTradeRequest() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestTradeValidator_SanitizeItemID(t *testing.T) {
	validator := NewTradeValidator()

	tests := []struct {
		name     string
		itemID   string
		expected string
	}{
		{
			name:     "no sanitization needed",
			itemID:   "item123",
			expected: "item123",
		},
		{
			name:     "remove spaces",
			itemID:   "item 123",
			expected: "item123",
		},
		{
			name:     "remove special characters",
			itemID:   "item@123#abc",
			expected: "item123abc",
		},
		{
			name:     "keep hyphens and underscores",
			itemID:   "item-123_abc",
			expected: "item-123_abc",
		},
		{
			name:     "keep equals",
			itemID:   "base64==",
			expected: "base64==",
		},
		{
			name:     "truncate long ID",
			itemID:   strings.Repeat("a", MaxItemIDLength+50),
			expected: strings.Repeat("a", MaxItemIDLength),
		},
		{
			name:     "remove all invalid",
			itemID:   "!@#$%^&*()",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.SanitizeItemID(tt.itemID)
			if result != tt.expected {
				t.Errorf("SanitizeItemID() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func BenchmarkTradeValidator_ValidateItemID(b *testing.B) {
	validator := NewTradeValidator()
	itemID := "valid-item-id-12345"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateItemID(itemID)
	}
}

func BenchmarkTradeValidator_ValidateItemIDs(b *testing.B) {
	validator := NewTradeValidator()
	itemIDs := []string{"item1", "item2", "item3", "item4", "item5"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateItemIDs(itemIDs)
	}
}

func BenchmarkTradeValidator_ValidateTradeRequest(b *testing.B) {
	validator := NewTradeValidator()
	offered := []string{"item1", "item2"}
	requested := []string{"item3", "item4", "item5"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateTradeRequest(offered, requested)
	}
}

func TestTradeValidator_ValidateTradeQuantity(t *testing.T) {
	validator := NewTradeValidator()

	tests := []struct {
		name     string
		quantity int
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid quantity - 1",
			quantity: 1,
			wantErr:  false,
		},
		{
			name:     "valid quantity - large",
			quantity: 999999,
			wantErr:  false,
		},
		{
			name:     "zero quantity",
			quantity: 0,
			wantErr:  true,
			errMsg:   "must be positive",
		},
		{
			name:     "negative quantity",
			quantity: -1,
			wantErr:  true,
			errMsg:   "must be positive",
		},
		{
			name:     "large negative quantity",
			quantity: -999,
			wantErr:  true,
			errMsg:   "must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTradeQuantity(tt.quantity)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTradeQuantity(%d) error = %v, wantErr %v", tt.quantity, err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateTradeQuantity(%d) error = %v, want error containing %q", tt.quantity, err, tt.errMsg)
			}
		})
	}
}
