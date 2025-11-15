// Package ui provides trade UI rendering functionality for item trading between players.
package ui

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

// TradeUI renders the trade interface for proposing and accepting item trades.
type TradeUI struct {
	// Position and size
	X, Y          int
	Width, Height int

	// Trade proposal data
	Proposal *TradeProposal

	// Visual settings
	BackgroundColor color.Color
	TextColor       color.Color
	HeaderColor     color.Color
	BorderColor     color.Color
	AcceptColor     color.Color
	RejectColor     color.Color
	Font            font.Face

	// Layout
	HeaderHeight  int
	ItemHeight    int
	ButtonHeight  int
	ButtonWidth   int
	Padding       int
	ItemsPerPanel int

	// State
	HoveredButton string // "accept", "reject", "cancel", ""
	Visible       bool
}

// TradeProposal represents a trade proposal for UI display.
type TradeProposal struct {
	ProposerName   string
	RecipientName  string
	OfferedItems   []TradeItem
	RequestedItems []TradeItem
	Status         string // "pending", "accepted", "rejected", "committed", "cancelled", "failed"
	ProposalTime   time.Time
	FailureReason  string
}

// TradeItem represents an item in the trade UI.
type TradeItem struct {
	ID       string
	Name     string
	Quantity int
	Rarity   string
	IconX    int // For sprite rendering
	IconY    int
}

// NewTradeUI creates a new trade UI instance.
func NewTradeUI(x, y, width, height int) *TradeUI {
	return &TradeUI{
		X:             x,
		Y:             y,
		Width:         width,
		Height:        height,
		Font:          basicfont.Face7x13,
		HeaderHeight:  30,
		ItemHeight:    40,
		ButtonHeight:  30,
		ButtonWidth:   100,
		Padding:       10,
		ItemsPerPanel: 6,
		// Default colors
		BackgroundColor: color.RGBA{30, 30, 40, 220},
		TextColor:       color.RGBA{220, 220, 220, 255},
		HeaderColor:     color.RGBA{50, 50, 70, 255},
		BorderColor:     color.RGBA{100, 100, 120, 255},
		AcceptColor:     color.RGBA{50, 150, 50, 255},
		RejectColor:     color.RGBA{150, 50, 50, 255},
		Visible:         false,
	}
}

// SetProposal sets the current trade proposal to display.
func (t *TradeUI) SetProposal(proposal *TradeProposal) {
	t.Proposal = proposal
	t.Visible = proposal != nil && proposal.Status == "pending"
}

// ClearProposal clears the current trade proposal.
func (t *TradeUI) ClearProposal() {
	t.Proposal = nil
	t.Visible = false
}

// IsVisible returns whether the trade UI is currently visible.
func (t *TradeUI) IsVisible() bool {
	return t.Visible && t.Proposal != nil
}

// Hide hides the trade UI.
func (t *TradeUI) Hide() {
	t.Visible = false
}

// Show shows the trade UI if a proposal exists.
func (t *TradeUI) Show() {
	if t.Proposal != nil {
		t.Visible = true
	}
}

// Update handles input and updates the trade UI state.
func (t *TradeUI) Update() {
	if !t.IsVisible() {
		return
	}

	// Update hover state based on mouse position
	mx, my := ebiten.CursorPosition()
	t.HoveredButton = t.getHoveredButton(mx, my)
}

// Draw renders the trade UI to the screen.
func (t *TradeUI) Draw(screen *ebiten.Image) {
	if !t.IsVisible() {
		return
	}

	// Draw background
	t.drawBackground(screen)

	// Draw header
	t.drawHeader(screen)

	// Draw offered items panel
	t.drawItemsPanel(screen, "You Offer:", t.Proposal.OfferedItems, t.X+t.Padding, t.Y+t.HeaderHeight+t.Padding)

	// Draw requested items panel
	panelY := t.Y + t.HeaderHeight + t.Padding + (t.ItemHeight*t.ItemsPerPanel + 40)
	t.drawItemsPanel(screen, "They Offer:", t.Proposal.RequestedItems, t.X+t.Padding, panelY)

	// Draw status message if not pending
	if t.Proposal.Status != "pending" {
		t.drawStatusMessage(screen)
	}

	// Draw buttons (only if pending)
	if t.Proposal.Status == "pending" {
		t.drawButtons(screen)
	}
}

// drawBackground draws the background panel.
func (t *TradeUI) drawBackground(screen *ebiten.Image) {
	bounds := image.Rect(t.X, t.Y, t.X+t.Width, t.Y+t.Height)
	vector.DrawFilledRect(screen, float32(bounds.Min.X), float32(bounds.Min.Y),
		float32(bounds.Dx()), float32(bounds.Dy()), t.BackgroundColor, false)

	// Draw border
	vector.StrokeRect(screen, float32(bounds.Min.X), float32(bounds.Min.Y),
		float32(bounds.Dx()), float32(bounds.Dy()), 2, t.BorderColor, false)
}

// drawHeader draws the header section with trade title and partner name.
func (t *TradeUI) drawHeader(screen *ebiten.Image) {
	headerBounds := image.Rect(t.X, t.Y, t.X+t.Width, t.Y+t.HeaderHeight)
	vector.DrawFilledRect(screen, float32(headerBounds.Min.X), float32(headerBounds.Min.Y),
		float32(headerBounds.Dx()), float32(headerBounds.Dy()), t.HeaderColor, false)

	// Draw title
	title := fmt.Sprintf("Trade with %s", t.Proposal.RecipientName)
	titleX := t.X + t.Width/2 - len(title)*7/2
	titleY := t.Y + t.HeaderHeight/2 + 6
	text.Draw(screen, title, t.Font, titleX, titleY, t.TextColor)

	// Draw border
	vector.StrokeLine(screen, float32(t.X), float32(t.Y+t.HeaderHeight),
		float32(t.X+t.Width), float32(t.Y+t.HeaderHeight), 1, t.BorderColor, false)
}

// drawItemsPanel draws a panel of trade items.
func (t *TradeUI) drawItemsPanel(screen *ebiten.Image, label string, items []TradeItem, x, y int) {
	// Draw label
	text.Draw(screen, label, t.Font, x, y+15, t.TextColor)

	// Draw items
	startY := y + 25
	for i, item := range items {
		if i >= t.ItemsPerPanel {
			break
		}
		itemY := startY + i*t.ItemHeight

		// Draw item background
		itemBounds := image.Rect(x, itemY, x+t.Width-2*t.Padding, itemY+t.ItemHeight-5)
		itemBG := color.RGBA{40, 40, 50, 180}
		vector.DrawFilledRect(screen, float32(itemBounds.Min.X), float32(itemBounds.Min.Y),
			float32(itemBounds.Dx()), float32(itemBounds.Dy()), itemBG, false)

		// Draw item name and quantity
		itemText := fmt.Sprintf("%s x%d", item.Name, item.Quantity)
		text.Draw(screen, itemText, t.Font, x+5, itemY+20, t.TextColor)

		// Draw rarity indicator
		rarityColor := t.getRarityColor(item.Rarity)
		vector.DrawFilledRect(screen, float32(itemBounds.Max.X-30), float32(itemY+10),
			20, 20, rarityColor, false)
	}

	// Draw "more items" indicator if needed
	if len(items) > t.ItemsPerPanel {
		moreText := fmt.Sprintf("+%d more...", len(items)-t.ItemsPerPanel)
		text.Draw(screen, moreText, t.Font, x+5, startY+t.ItemsPerPanel*t.ItemHeight+15,
			color.RGBA{150, 150, 150, 255})
	}
}

// drawButtons draws the accept/reject/cancel buttons.
func (t *TradeUI) drawButtons(screen *ebiten.Image) {
	buttonY := t.Y + t.Height - t.ButtonHeight - t.Padding

	// Accept button
	acceptX := t.X + t.Width/2 - t.ButtonWidth - 5
	acceptColor := t.AcceptColor
	if t.HoveredButton == "accept" {
		acceptColor = color.RGBA{70, 180, 70, 255}
	}
	t.drawButton(screen, "Accept", acceptX, buttonY, acceptColor)

	// Reject button
	rejectX := t.X + t.Width/2 + 5
	rejectColor := t.RejectColor
	if t.HoveredButton == "reject" {
		rejectColor = color.RGBA{180, 70, 70, 255}
	}
	t.drawButton(screen, "Reject", rejectX, buttonY, rejectColor)
}

// drawButton draws a single button.
func (t *TradeUI) drawButton(screen *ebiten.Image, label string, x, y int, bgColor color.Color) {
	buttonBounds := image.Rect(x, y, x+t.ButtonWidth, y+t.ButtonHeight)
	vector.DrawFilledRect(screen, float32(buttonBounds.Min.X), float32(buttonBounds.Min.Y),
		float32(buttonBounds.Dx()), float32(buttonBounds.Dy()), bgColor, false)

	// Draw border
	vector.StrokeRect(screen, float32(buttonBounds.Min.X), float32(buttonBounds.Min.Y),
		float32(buttonBounds.Dx()), float32(buttonBounds.Dy()), 2, t.BorderColor, false)

	// Draw text
	textX := x + t.ButtonWidth/2 - len(label)*7/2
	textY := y + t.ButtonHeight/2 + 6
	text.Draw(screen, label, t.Font, textX, textY, color.White)
}

// drawStatusMessage draws the current status message.
func (t *TradeUI) drawStatusMessage(screen *ebiten.Image) {
	var statusMsg string
	var statusColor color.Color

	switch t.Proposal.Status {
	case "accepted":
		statusMsg = "Trade Accepted - Waiting for commit..."
		statusColor = color.RGBA{50, 200, 50, 255}
	case "rejected":
		statusMsg = "Trade Rejected"
		statusColor = color.RGBA{200, 50, 50, 255}
	case "committed":
		statusMsg = "Trade Complete!"
		statusColor = color.RGBA{50, 200, 50, 255}
	case "cancelled":
		statusMsg = "Trade Cancelled"
		statusColor = color.RGBA{200, 100, 50, 255}
	case "failed":
		statusMsg = fmt.Sprintf("Trade Failed: %s", t.Proposal.FailureReason)
		statusColor = color.RGBA{200, 50, 50, 255}
	default:
		return
	}

	msgY := t.Y + t.Height - t.ButtonHeight - t.Padding
	msgX := t.X + t.Width/2 - len(statusMsg)*7/2
	text.Draw(screen, statusMsg, t.Font, msgX, msgY+15, statusColor)
}

// getHoveredButton returns which button is currently hovered.
func (t *TradeUI) getHoveredButton(mx, my int) string {
	buttonY := t.Y + t.Height - t.ButtonHeight - t.Padding

	// Accept button bounds
	acceptX := t.X + t.Width/2 - t.ButtonWidth - 5
	if mx >= acceptX && mx <= acceptX+t.ButtonWidth && my >= buttonY && my <= buttonY+t.ButtonHeight {
		return "accept"
	}

	// Reject button bounds
	rejectX := t.X + t.Width/2 + 5
	if mx >= rejectX && mx <= rejectX+t.ButtonWidth && my >= buttonY && my <= buttonY+t.ButtonHeight {
		return "reject"
	}

	return ""
}

// IsButtonClicked returns true if the specified button is clicked.
func (t *TradeUI) IsButtonClicked(button string) bool {
	if !t.IsVisible() {
		return false
	}
	return t.HoveredButton == button && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
}

// GetClickedButton returns which button was clicked, or empty string.
func (t *TradeUI) GetClickedButton() string {
	if !t.IsVisible() {
		return ""
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		return t.HoveredButton
	}
	return ""
}

// getRarityColor returns a color for the given rarity string.
func (t *TradeUI) getRarityColor(rarity string) color.Color {
	switch rarity {
	case "Common":
		return color.RGBA{150, 150, 150, 255}
	case "Uncommon":
		return color.RGBA{80, 200, 80, 255}
	case "Rare":
		return color.RGBA{80, 80, 220, 255}
	case "Epic":
		return color.RGBA{180, 80, 220, 255}
	case "Legendary":
		return color.RGBA{220, 180, 50, 255}
	default:
		return color.RGBA{100, 100, 100, 255}
	}
}

// SetColors allows customizing UI colors.
func (t *TradeUI) SetColors(bg, text, header, border, accept, reject color.Color) {
	t.BackgroundColor = bg
	t.TextColor = text
	t.HeaderColor = header
	t.BorderColor = border
	t.AcceptColor = accept
	t.RejectColor = reject
}

// SetFont allows customizing the UI font.
func (t *TradeUI) SetFont(f font.Face) {
	t.Font = f
}
