package ui

import (
	"testing"
	"time"
)

func TestNewTradeDialogUI(t *testing.T) {
	ui := NewTradeDialogUI(100, 100, 400, 300)

	if ui.X != 100 || ui.Y != 100 {
		t.Errorf("Expected position (100,100), got (%d,%d)", ui.X, ui.Y)
	}
	if ui.Width != 400 || ui.Height != 300 {
		t.Errorf("Expected size (400,300), got (%d,%d)", ui.Width, ui.Height)
	}
	if ui.Active {
		t.Error("Expected dialog to start inactive")
	}
	if ui.TimeoutDuration != 30*time.Second {
		t.Errorf("Expected 30s timeout, got %v", ui.TimeoutDuration)
	}
}

func TestShowProposal(t *testing.T) {
	ui := NewTradeDialogUI(0, 0, 400, 300)

	offeredItems := []TradeItem{
		{ID: "1", Name: "Iron Sword", Quantity: 1, Rarity: "Common"},
	}
	requestedItems := []TradeItem{
		{ID: "2", Name: "Gold Coin", Quantity: 100, Rarity: "Common"},
	}

	ui.ShowProposal("Alice", "Bob", offeredItems, requestedItems, true)

	if !ui.Active {
		t.Error("Expected dialog to be active after ShowProposal")
	}
	if ui.ProposerName != "Alice" {
		t.Errorf("Expected proposer 'Alice', got '%s'", ui.ProposerName)
	}
	if ui.RecipientName != "Bob" {
		t.Errorf("Expected recipient 'Bob', got '%s'", ui.RecipientName)
	}
	if len(ui.OfferedItems) != 1 {
		t.Errorf("Expected 1 offered item, got %d", len(ui.OfferedItems))
	}
	if len(ui.RequestedItems) != 1 {
		t.Errorf("Expected 1 requested item, got %d", len(ui.RequestedItems))
	}
	if ui.Status != TradePending {
		t.Errorf("Expected status TradePending, got %v", ui.Status)
	}
	if !ui.IsProposer {
		t.Error("Expected IsProposer to be true")
	}
	if !ui.ConfirmSelected {
		t.Error("Expected ConfirmSelected to default to true")
	}
}

func TestTradeDialogHide(t *testing.T) {
	ui := NewTradeDialogUI(0, 0, 400, 300)
	ui.ShowProposal("Alice", "Bob", nil, nil, false)

	if !ui.Active {
		t.Error("Expected dialog to be active before Hide")
	}

	ui.Hide()

	if ui.Active {
		t.Error("Expected dialog to be inactive after Hide")
	}
}

func TestTradeDialogToggleSelection(t *testing.T) {
	ui := NewTradeDialogUI(0, 0, 400, 300)
	ui.ShowProposal("Alice", "Bob", nil, nil, false)

	if !ui.ConfirmSelected {
		t.Error("Expected ConfirmSelected to start as true")
	}

	ui.ToggleSelection()
	if ui.ConfirmSelected {
		t.Error("Expected ConfirmSelected to be false after first toggle")
	}

	ui.ToggleSelection()
	if !ui.ConfirmSelected {
		t.Error("Expected ConfirmSelected to be true after second toggle")
	}
}

func TestTradeDialogGetSelectedAction(t *testing.T) {
	ui := NewTradeDialogUI(0, 0, 400, 300)
	ui.ShowProposal("Alice", "Bob", nil, nil, false)

	action := ui.GetSelectedAction()
	if action != "accept" {
		t.Errorf("Expected 'accept', got '%s'", action)
	}

	ui.ToggleSelection()
	action = ui.GetSelectedAction()
	if action != "reject" {
		t.Errorf("Expected 'reject', got '%s'", action)
	}
}

func TestUpdateStatus(t *testing.T) {
	ui := NewTradeDialogUI(0, 0, 400, 300)
	ui.ShowProposal("Alice", "Bob", nil, nil, false)

	ui.UpdateStatus(TradeAccepted, "")
	if ui.Status != TradeAccepted {
		t.Errorf("Expected status TradeAccepted, got %v", ui.Status)
	}
	if ui.FailureReason != "" {
		t.Errorf("Expected empty failure reason, got '%s'", ui.FailureReason)
	}

	ui.UpdateStatus(TradeFailed, "Insufficient proximity")
	if ui.Status != TradeFailed {
		t.Errorf("Expected status TradeFailed, got %v", ui.Status)
	}
	if ui.FailureReason != "Insufficient proximity" {
		t.Errorf("Expected failure reason 'Insufficient proximity', got '%s'", ui.FailureReason)
	}
}

func TestTradeStatusString(t *testing.T) {
	tests := []struct {
		status   TradeStatus
		expected string
	}{
		{TradePending, "Pending"},
		{TradeAccepted, "Accepted"},
		{TradeRejected, "Rejected"},
		{TradeCommitted, "Completed"},
		{TradeCancelled, "Cancelled"},
		{TradeFailed, "Failed"},
		{TradeStatus(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.status.String()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestUpdateTimeout(t *testing.T) {
	ui := NewTradeDialogUI(0, 0, 400, 300)
	ui.TimeoutDuration = 100 * time.Millisecond
	ui.ShowProposal("Alice", "Bob", nil, nil, false)

	// Initially pending
	if ui.Status != TradePending {
		t.Errorf("Expected status TradePending, got %v", ui.Status)
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)
	ui.Update(0.1)

	// Should now be cancelled
	if ui.Status != TradeCancelled {
		t.Errorf("Expected status TradeCancelled after timeout, got %v", ui.Status)
	}
	if ui.FailureReason != "Timeout: No response received" {
		t.Errorf("Expected timeout failure reason, got '%s'", ui.FailureReason)
	}
}

func TestUpdateNoTimeoutAfterResponse(t *testing.T) {
	ui := NewTradeDialogUI(0, 0, 400, 300)
	ui.TimeoutDuration = 100 * time.Millisecond
	ui.ShowProposal("Alice", "Bob", nil, nil, false)

	// Accept trade before timeout
	ui.UpdateStatus(TradeAccepted, "")

	// Wait for what would be timeout
	time.Sleep(150 * time.Millisecond)
	ui.Update(0.1)

	// Should still be accepted (not cancelled by timeout)
	if ui.Status != TradeAccepted {
		t.Errorf("Expected status TradeAccepted, got %v", ui.Status)
	}
}

func TestGetRarityColor(t *testing.T) {
	ui := NewTradeDialogUI(0, 0, 400, 300)

	tests := []struct {
		rarity string
		// We can't easily test exact color values without exposing the method,
		// but we can test that it doesn't panic
	}{
		{"Common"},
		{"Uncommon"},
		{"Rare"},
		{"Epic"},
		{"Legendary"},
		{"Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.rarity, func(t *testing.T) {
			// Just verify it doesn't panic
			_ = ui.getRarityColor(tt.rarity)
		})
	}
}

func TestTradeItemsEmpty(t *testing.T) {
	ui := NewTradeDialogUI(0, 0, 400, 300)

	// Test with empty item lists
	ui.ShowProposal("Alice", "Bob", []TradeItem{}, []TradeItem{}, false)

	if len(ui.OfferedItems) != 0 {
		t.Errorf("Expected 0 offered items, got %d", len(ui.OfferedItems))
	}
	if len(ui.RequestedItems) != 0 {
		t.Errorf("Expected 0 requested items, got %d", len(ui.RequestedItems))
	}
}

func TestTradeItemsMultiple(t *testing.T) {
	ui := NewTradeDialogUI(0, 0, 400, 300)

	offeredItems := []TradeItem{
		{ID: "1", Name: "Iron Sword", Quantity: 1, Rarity: "Common"},
		{ID: "2", Name: "Wooden Shield", Quantity: 1, Rarity: "Common"},
		{ID: "3", Name: "Health Potion", Quantity: 5, Rarity: "Uncommon"},
	}
	requestedItems := []TradeItem{
		{ID: "4", Name: "Gold Coin", Quantity: 100, Rarity: "Common"},
		{ID: "5", Name: "Rare Gem", Quantity: 1, Rarity: "Rare"},
	}

	ui.ShowProposal("Alice", "Bob", offeredItems, requestedItems, true)

	if len(ui.OfferedItems) != 3 {
		t.Errorf("Expected 3 offered items, got %d", len(ui.OfferedItems))
	}
	if len(ui.RequestedItems) != 2 {
		t.Errorf("Expected 2 requested items, got %d", len(ui.RequestedItems))
	}
}

func TestProposerVsRecipient(t *testing.T) {
	// Test as proposer
	proposerUI := NewTradeDialogUI(0, 0, 400, 300)
	proposerUI.ShowProposal("Alice", "Bob", nil, nil, true)

	if !proposerUI.IsProposer {
		t.Error("Expected IsProposer to be true for proposer")
	}

	// Test as recipient
	recipientUI := NewTradeDialogUI(0, 0, 400, 300)
	recipientUI.ShowProposal("Alice", "Bob", nil, nil, false)

	if recipientUI.IsProposer {
		t.Error("Expected IsProposer to be false for recipient")
	}
}

// Benchmarks

func BenchmarkShowProposal(b *testing.B) {
	ui := NewTradeDialogUI(0, 0, 400, 300)
	items := []TradeItem{
		{ID: "1", Name: "Item", Quantity: 1, Rarity: "Common"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ui.ShowProposal("Alice", "Bob", items, items, false)
	}
}

func BenchmarkTradeDialogUpdate(b *testing.B) {
	ui := NewTradeDialogUI(0, 0, 400, 300)
	ui.ShowProposal("Alice", "Bob", nil, nil, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ui.Update(0.016)
	}
}

func BenchmarkToggleSelection(b *testing.B) {
	ui := NewTradeDialogUI(0, 0, 400, 300)
	ui.ShowProposal("Alice", "Bob", nil, nil, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ui.ToggleSelection()
	}
}
