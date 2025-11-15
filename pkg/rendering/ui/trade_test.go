package ui

import (
	"image/color"
	"testing"
	"time"
)

func TestNewTradeUI(t *testing.T) {
	ui := NewTradeUI(100, 100, 400, 500)

	if ui.X != 100 || ui.Y != 100 {
		t.Errorf("Position mismatch: got (%d, %d), want (100, 100)", ui.X, ui.Y)
	}
	if ui.Width != 400 || ui.Height != 500 {
		t.Errorf("Size mismatch: got (%d, %d), want (400, 500)", ui.Width, ui.Height)
	}
	if ui.Visible {
		t.Error("New TradeUI should not be visible initially")
	}
	if ui.Proposal != nil {
		t.Error("New TradeUI should have nil proposal")
	}
}

func TestTradeUISetProposal(t *testing.T) {
	ui := NewTradeUI(0, 0, 400, 500)

	proposal := &TradeProposal{
		ProposerName:  "Alice",
		RecipientName: "Bob",
		OfferedItems: []TradeItem{
			{ID: "item1", Name: "Sword", Quantity: 1, Rarity: "Rare"},
		},
		RequestedItems: []TradeItem{
			{ID: "item2", Name: "Shield", Quantity: 1, Rarity: "Common"},
		},
		Status:       "pending",
		ProposalTime: time.Now(),
	}

	ui.SetProposal(proposal)

	if ui.Proposal != proposal {
		t.Error("SetProposal did not set proposal correctly")
	}
	if !ui.Visible {
		t.Error("SetProposal should make UI visible for pending proposals")
	}
}

func TestTradeUISetProposalNonPending(t *testing.T) {
	ui := NewTradeUI(0, 0, 400, 500)

	proposal := &TradeProposal{
		ProposerName:   "Alice",
		RecipientName:  "Bob",
		OfferedItems:   []TradeItem{},
		RequestedItems: []TradeItem{},
		Status:         "accepted", // Non-pending
		ProposalTime:   time.Now(),
	}

	ui.SetProposal(proposal)

	if ui.Proposal != proposal {
		t.Error("SetProposal did not set proposal correctly")
	}
	if ui.Visible {
		t.Error("SetProposal should not make UI visible for non-pending proposals")
	}
}

func TestTradeUIClearProposal(t *testing.T) {
	ui := NewTradeUI(0, 0, 400, 500)

	proposal := &TradeProposal{
		ProposerName:  "Alice",
		RecipientName: "Bob",
		Status:        "pending",
		ProposalTime:  time.Now(),
	}

	ui.SetProposal(proposal)
	ui.ClearProposal()

	if ui.Proposal != nil {
		t.Error("ClearProposal did not clear proposal")
	}
	if ui.Visible {
		t.Error("ClearProposal should hide UI")
	}
}

func TestTradeUIIsVisible(t *testing.T) {
	ui := NewTradeUI(0, 0, 400, 500)

	// Initially not visible
	if ui.IsVisible() {
		t.Error("New TradeUI should not be visible")
	}

	// Visible after setting pending proposal
	proposal := &TradeProposal{
		ProposerName:  "Alice",
		RecipientName: "Bob",
		Status:        "pending",
		ProposalTime:  time.Now(),
	}
	ui.SetProposal(proposal)
	if !ui.IsVisible() {
		t.Error("TradeUI should be visible with pending proposal")
	}

	// Not visible after hiding
	ui.Hide()
	if ui.IsVisible() {
		t.Error("TradeUI should not be visible after Hide()")
	}

	// Visible again after Show()
	ui.Show()
	if !ui.IsVisible() {
		t.Error("TradeUI should be visible after Show()")
	}

	// Not visible with no proposal
	ui.ClearProposal()
	ui.Show()
	if ui.IsVisible() {
		t.Error("TradeUI should not be visible with no proposal even after Show()")
	}
}

func TestTradeUIGetRarityColor(t *testing.T) {
	ui := NewTradeUI(0, 0, 400, 500)

	tests := []struct {
		rarity string
		want   color.Color
	}{
		{"Common", color.RGBA{150, 150, 150, 255}},
		{"Uncommon", color.RGBA{80, 200, 80, 255}},
		{"Rare", color.RGBA{80, 80, 220, 255}},
		{"Epic", color.RGBA{180, 80, 220, 255}},
		{"Legendary", color.RGBA{220, 180, 50, 255}},
		{"Unknown", color.RGBA{100, 100, 100, 255}}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.rarity, func(t *testing.T) {
			got := ui.getRarityColor(tt.rarity)
			if got != tt.want {
				t.Errorf("getRarityColor(%q) = %v, want %v", tt.rarity, got, tt.want)
			}
		})
	}
}

func TestTradeUIGetHoveredButton(t *testing.T) {
	ui := NewTradeUI(100, 100, 400, 500)

	// Accept button bounds: x = 100 + 400/2 - 100 - 5 = 95, y = 100 + 500 - 30 - 10 = 560
	acceptX := 100 + 400/2 - 100 - 5
	acceptY := 100 + 500 - 30 - 10

	// Reject button bounds: x = 100 + 400/2 + 5 = 205, y = 560
	rejectX := 100 + 400/2 + 5
	rejectY := acceptY

	tests := []struct {
		name string
		mx   int
		my   int
		want string
	}{
		{"Accept button center", acceptX + 50, acceptY + 15, "accept"},
		{"Reject button center", rejectX + 50, rejectY + 15, "reject"},
		{"Outside buttons", 0, 0, ""},
		{"Between buttons", acceptX + 102, acceptY + 15, ""}, // Between accept (ends at 195) and reject (starts at 205)
		{"Above buttons", acceptX + 50, acceptY - 10, ""},
		{"Below buttons", acceptX + 50, acceptY + 40, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ui.getHoveredButton(tt.mx, tt.my)
			if got != tt.want {
				t.Errorf("getHoveredButton(%d, %d) = %q, want %q", tt.mx, tt.my, got, tt.want)
			}
		})
	}
}

func TestTradeUISetColors(t *testing.T) {
	ui := NewTradeUI(0, 0, 400, 500)

	bg := color.RGBA{10, 10, 10, 255}
	text := color.RGBA{200, 200, 200, 255}
	header := color.RGBA{20, 20, 20, 255}
	border := color.RGBA{50, 50, 50, 255}
	accept := color.RGBA{0, 255, 0, 255}
	reject := color.RGBA{255, 0, 0, 255}

	ui.SetColors(bg, text, header, border, accept, reject)

	if ui.BackgroundColor != bg {
		t.Errorf("BackgroundColor mismatch: got %v, want %v", ui.BackgroundColor, bg)
	}
	if ui.TextColor != text {
		t.Errorf("TextColor mismatch: got %v, want %v", ui.TextColor, text)
	}
	if ui.HeaderColor != header {
		t.Errorf("HeaderColor mismatch: got %v, want %v", ui.HeaderColor, header)
	}
	if ui.BorderColor != border {
		t.Errorf("BorderColor mismatch: got %v, want %v", ui.BorderColor, border)
	}
	if ui.AcceptColor != accept {
		t.Errorf("AcceptColor mismatch: got %v, want %v", ui.AcceptColor, accept)
	}
	if ui.RejectColor != reject {
		t.Errorf("RejectColor mismatch: got %v, want %v", ui.RejectColor, reject)
	}
}

func TestTradeUIProposalStatuses(t *testing.T) {
	ui := NewTradeUI(0, 0, 400, 500)

	tests := []struct {
		status        string
		shouldVisible bool
	}{
		{"pending", true},
		{"accepted", false},
		{"rejected", false},
		{"committed", false},
		{"cancelled", false},
		{"failed", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			proposal := &TradeProposal{
				ProposerName:  "Alice",
				RecipientName: "Bob",
				Status:        tt.status,
				ProposalTime:  time.Now(),
			}

			ui.SetProposal(proposal)

			if ui.Visible != tt.shouldVisible {
				t.Errorf("Status %q: Visible = %v, want %v", tt.status, ui.Visible, tt.shouldVisible)
			}
		})
	}
}

func TestTradeUIMultipleItems(t *testing.T) {
	ui := NewTradeUI(0, 0, 400, 500)

	// Create proposal with more items than can be displayed
	offeredItems := make([]TradeItem, 10)
	for i := 0; i < 10; i++ {
		offeredItems[i] = TradeItem{
			ID:       "item" + string(rune('A'+i)),
			Name:     "Item " + string(rune('A'+i)),
			Quantity: i + 1,
			Rarity:   "Common",
		}
	}

	proposal := &TradeProposal{
		ProposerName:   "Alice",
		RecipientName:  "Bob",
		OfferedItems:   offeredItems,
		RequestedItems: []TradeItem{},
		Status:         "pending",
		ProposalTime:   time.Now(),
	}

	ui.SetProposal(proposal)

	if len(ui.Proposal.OfferedItems) != 10 {
		t.Errorf("Expected 10 offered items, got %d", len(ui.Proposal.OfferedItems))
	}

	// ItemsPerPanel default is 6, so we should show 6 items with "+4 more..." indicator
	if ui.ItemsPerPanel != 6 {
		t.Errorf("Expected ItemsPerPanel = 6, got %d", ui.ItemsPerPanel)
	}
}

func TestTradeUIFailureReason(t *testing.T) {
	ui := NewTradeUI(0, 0, 400, 500)

	proposal := &TradeProposal{
		ProposerName:   "Alice",
		RecipientName:  "Bob",
		Status:         "failed",
		FailureReason:  "Insufficient proximity",
		ProposalTime:   time.Now(),
		OfferedItems:   []TradeItem{},
		RequestedItems: []TradeItem{},
	}

	ui.SetProposal(proposal)

	if ui.Proposal.FailureReason != "Insufficient proximity" {
		t.Errorf("FailureReason mismatch: got %q, want %q",
			ui.Proposal.FailureReason, "Insufficient proximity")
	}
}

func TestTradeUIHideShow(t *testing.T) {
	ui := NewTradeUI(0, 0, 400, 500)

	proposal := &TradeProposal{
		ProposerName:  "Alice",
		RecipientName: "Bob",
		Status:        "pending",
		ProposalTime:  time.Now(),
	}

	// Set proposal (should become visible)
	ui.SetProposal(proposal)
	if !ui.Visible {
		t.Error("UI should be visible after SetProposal with pending status")
	}

	// Hide
	ui.Hide()
	if ui.Visible {
		t.Error("UI should not be visible after Hide()")
	}

	// Show again
	ui.Show()
	if !ui.Visible {
		t.Error("UI should be visible after Show() with proposal present")
	}

	// Clear proposal
	ui.ClearProposal()
	if ui.Visible {
		t.Error("UI should not be visible after ClearProposal()")
	}

	// Show should not work without proposal
	ui.Show()
	if ui.Visible {
		t.Error("UI should not be visible after Show() without proposal")
	}
}

// Benchmark tests
func BenchmarkTradeUISetProposal(b *testing.B) {
	ui := NewTradeUI(0, 0, 400, 500)

	proposal := &TradeProposal{
		ProposerName:  "Alice",
		RecipientName: "Bob",
		OfferedItems: []TradeItem{
			{ID: "item1", Name: "Sword", Quantity: 1, Rarity: "Rare"},
			{ID: "item2", Name: "Shield", Quantity: 1, Rarity: "Common"},
		},
		RequestedItems: []TradeItem{
			{ID: "item3", Name: "Potion", Quantity: 5, Rarity: "Uncommon"},
		},
		Status:       "pending",
		ProposalTime: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ui.SetProposal(proposal)
	}
}

func BenchmarkTradeUIGetHoveredButton(b *testing.B) {
	ui := NewTradeUI(100, 100, 400, 500)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ui.getHoveredButton(250, 560)
	}
}

func BenchmarkTradeUIGetRarityColor(b *testing.B) {
	ui := NewTradeUI(0, 0, 400, 500)

	rarities := []string{"Common", "Uncommon", "Rare", "Epic", "Legendary"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ui.getRarityColor(rarities[i%len(rarities)])
	}
}
