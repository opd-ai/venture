// Package ui provides trade dialog UI rendering functionality.
package ui

import (
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

// TradeDialogUI renders trade proposal and confirmation dialogs.
type TradeDialogUI struct {
	// Position and size
	X, Y          int
	Width, Height int

	// Trade state
	Active          bool
	ProposerName    string
	RecipientName   string
	OfferedItems    []TradeItem
	RequestedItems  []TradeItem
	Status          TradeStatus
	FailureReason   string
	ProposalTime    time.Time
	TimeoutDuration time.Duration
	IsProposer      bool // Whether local player is the proposer
	ConfirmSelected bool // Whether "Accept" is selected (vs "Reject")

	// Visual settings
	BackgroundColor color.Color
	BorderColor     color.Color
	HeaderColor     color.Color
	TextColor       color.Color
	AccentColor     color.Color
	ErrorColor      color.Color
	Font            font.Face

	// Layout
	HeaderHeight int
	ItemHeight   int
	ButtonHeight int
	Padding      int
}

// TradeItem represents an item in a trade.
type TradeItem struct {
	ID       string
	Name     string
	Quantity int
	Rarity   string
}

// TradeStatus represents the current state of a trade.
type TradeStatus int

const (
	// TradePending - Trade proposed, awaiting response
	TradePending TradeStatus = iota
	// TradeAccepted - Trade accepted, awaiting server commit
	TradeAccepted
	// TradeRejected - Trade rejected by recipient
	TradeRejected
	// TradeCommitted - Trade successfully completed
	TradeCommitted
	// TradeCancelled - Trade cancelled by proposer
	TradeCancelled
	// TradeFailed - Trade failed (proximity, trust, ownership issue)
	TradeFailed
)

// String returns the string representation of TradeStatus.
func (s TradeStatus) String() string {
	switch s {
	case TradePending:
		return "Pending"
	case TradeAccepted:
		return "Accepted"
	case TradeRejected:
		return "Rejected"
	case TradeCommitted:
		return "Completed"
	case TradeCancelled:
		return "Cancelled"
	case TradeFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// NewTradeDialogUI creates a new trade dialog UI instance.
func NewTradeDialogUI(x, y, width, height int) *TradeDialogUI {
	return &TradeDialogUI{
		X:               x,
		Y:               y,
		Width:           width,
		Height:          height,
		Active:          false,
		TimeoutDuration: 30 * time.Second,
		BackgroundColor: color.RGBA{30, 30, 40, 240},
		BorderColor:     color.RGBA{100, 150, 200, 255},
		HeaderColor:     color.RGBA{50, 70, 90, 255},
		TextColor:       color.RGBA{220, 220, 220, 255},
		AccentColor:     color.RGBA{100, 200, 100, 255},
		ErrorColor:      color.RGBA{255, 100, 100, 255},
		Font:            basicfont.Face7x13,
		HeaderHeight:    30,
		ItemHeight:      20,
		ButtonHeight:    35,
		Padding:         10,
	}
}

// ShowProposal displays a trade proposal dialog.
func (ui *TradeDialogUI) ShowProposal(proposerName, recipientName string, offeredItems, requestedItems []TradeItem, isProposer bool) {
	ui.Active = true
	ui.ProposerName = proposerName
	ui.RecipientName = recipientName
	ui.OfferedItems = offeredItems
	ui.RequestedItems = requestedItems
	ui.Status = TradePending
	ui.FailureReason = ""
	ui.ProposalTime = time.Now()
	ui.IsProposer = isProposer
	ui.ConfirmSelected = true // Default to "Accept"
}

// UpdateStatus updates the trade status and failure reason.
func (ui *TradeDialogUI) UpdateStatus(status TradeStatus, failureReason string) {
	ui.Status = status
	ui.FailureReason = failureReason
}

// Hide closes the trade dialog.
func (ui *TradeDialogUI) Hide() {
	ui.Active = false
}

// ToggleSelection switches between Accept and Reject buttons.
func (ui *TradeDialogUI) ToggleSelection() {
	ui.ConfirmSelected = !ui.ConfirmSelected
}

// GetSelectedAction returns the currently selected action ("accept" or "reject").
func (ui *TradeDialogUI) GetSelectedAction() string {
	if ui.ConfirmSelected {
		return "accept"
	}
	return "reject"
}

// Update updates the trade dialog state (timeout checking).
func (ui *TradeDialogUI) Update(deltaTime float64) {
	if !ui.Active {
		return
	}

	// Check for timeout (only for pending trades)
	if ui.Status == TradePending {
		elapsed := time.Since(ui.ProposalTime)
		if elapsed >= ui.TimeoutDuration {
			ui.UpdateStatus(TradeCancelled, "Timeout: No response received")
		}
	}
}

// Render draws the trade dialog to the screen.
func (ui *TradeDialogUI) Render(screen *ebiten.Image) {
	if !ui.Active {
		return
	}

	// Draw semi-transparent overlay
	overlay := ebiten.NewImage(screen.Bounds().Dx(), screen.Bounds().Dy())
	overlay.Fill(color.RGBA{0, 0, 0, 150})
	screen.DrawImage(overlay, nil)

	// Draw dialog panel
	ui.drawPanel(screen, ui.X, ui.Y, ui.Width, ui.Height, ui.BackgroundColor, ui.BorderColor)

	// Draw header
	ui.drawHeader(screen)

	// Draw trade details
	detailsY := ui.Y + ui.HeaderHeight + ui.Padding
	ui.drawTradeDetails(screen, detailsY)

	// Draw status/buttons
	buttonsY := ui.Y + ui.Height - ui.ButtonHeight - ui.Padding
	if ui.Status == TradePending && !ui.IsProposer {
		ui.drawButtons(screen, buttonsY)
	} else {
		ui.drawStatus(screen, buttonsY)
	}

	// Draw timeout indicator for pending trades
	if ui.Status == TradePending {
		ui.drawTimeoutIndicator(screen)
	}
}

// drawPanel draws a filled rectangle with border.
func (ui *TradeDialogUI) drawPanel(screen *ebiten.Image, x, y, width, height int, bgColor, borderColor color.Color) {
	// Background
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(width), float32(height), bgColor, false)

	// Border (2px thick)
	vector.StrokeRect(screen, float32(x), float32(y), float32(width), float32(height), 2, borderColor, false)
}

// drawHeader draws the dialog header with title.
func (ui *TradeDialogUI) drawHeader(screen *ebiten.Image) {
	headerY := ui.Y

	// Header background
	vector.DrawFilledRect(screen, float32(ui.X), float32(headerY), float32(ui.Width), float32(ui.HeaderHeight), ui.HeaderColor, false)

	// Title text
	title := fmt.Sprintf("Trade Proposal: %s ⇄ %s", ui.ProposerName, ui.RecipientName)
	text.Draw(screen, title, ui.Font, ui.X+ui.Padding, headerY+20, ui.TextColor)
}

// drawTradeDetails draws the offered and requested items.
func (ui *TradeDialogUI) drawTradeDetails(screen *ebiten.Image, y int) {
	// Calculate column layout
	columnWidth := (ui.Width - 3*ui.Padding) / 2
	leftX := ui.X + ui.Padding
	rightX := ui.X + ui.Width/2 + ui.Padding/2

	// Left column: Offered items
	ui.drawItemList(screen, leftX, y, columnWidth, "Offered by "+ui.ProposerName+":", ui.OfferedItems)

	// Right column: Requested items
	ui.drawItemList(screen, rightX, y, columnWidth, "Requested from "+ui.RecipientName+":", ui.RequestedItems)
}

// drawItemList draws a list of items with header.
func (ui *TradeDialogUI) drawItemList(screen *ebiten.Image, x, y, width int, header string, items []TradeItem) {
	// Header
	text.Draw(screen, header, ui.Font, x, y+15, ui.AccentColor)
	drawY := y + 30

	// Items
	if len(items) == 0 {
		text.Draw(screen, "  (Nothing)", ui.Font, x, drawY+15, color.RGBA{150, 150, 150, 255})
	} else {
		for _, item := range items {
			itemText := fmt.Sprintf("  • %s x%d", item.Name, item.Quantity)
			itemColor := ui.getRarityColor(item.Rarity)
			text.Draw(screen, itemText, ui.Font, x, drawY+15, itemColor)
			drawY += ui.ItemHeight
		}
	}
}

// getRarityColor returns the color for an item rarity.
func (ui *TradeDialogUI) getRarityColor(rarity string) color.Color {
	switch rarity {
	case "Common":
		return color.RGBA{200, 200, 200, 255}
	case "Uncommon":
		return color.RGBA{100, 255, 100, 255}
	case "Rare":
		return color.RGBA{100, 150, 255, 255}
	case "Epic":
		return color.RGBA{200, 100, 255, 255}
	case "Legendary":
		return color.RGBA{255, 180, 50, 255}
	default:
		return ui.TextColor
	}
}

// drawButtons draws Accept/Reject buttons (for recipient only).
func (ui *TradeDialogUI) drawButtons(screen *ebiten.Image, y int) {
	buttonWidth := (ui.Width - 3*ui.Padding) / 2
	acceptX := ui.X + ui.Padding
	rejectX := ui.X + ui.Width/2 + ui.Padding/2

	// Accept button
	acceptColor := ui.AccentColor
	if ui.ConfirmSelected {
		acceptColor = color.RGBA{150, 255, 150, 255} // Highlight
	}
	ui.drawButton(screen, acceptX, y, buttonWidth, ui.ButtonHeight, "Accept", acceptColor)

	// Reject button
	rejectColor := ui.ErrorColor
	if !ui.ConfirmSelected {
		rejectColor = color.RGBA{255, 150, 150, 255} // Highlight
	}
	ui.drawButton(screen, rejectX, y, buttonWidth, ui.ButtonHeight, "Reject", rejectColor)
}

// drawButton draws a button with text.
func (ui *TradeDialogUI) drawButton(screen *ebiten.Image, x, y, width, height int, label string, bgColor color.Color) {
	// Button background
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(width), float32(height), bgColor, false)

	// Button border
	vector.StrokeRect(screen, float32(x), float32(y), float32(width), float32(height), 1, ui.BorderColor, false)

	// Button text (centered)
	textWidth := len(label) * 7 // Approximate width with 7x13 font
	textX := x + (width-textWidth)/2
	textY := y + height/2 + 5
	text.Draw(screen, label, ui.Font, textX, textY, color.RGBA{0, 0, 0, 255})
}

// drawStatus draws the current trade status (for proposer or after response).
func (ui *TradeDialogUI) drawStatus(screen *ebiten.Image, y int) {
	statusText := "Status: " + ui.Status.String()
	statusColor := ui.TextColor

	switch ui.Status {
	case TradePending:
		statusText = "Status: Waiting for response..."
		statusColor = ui.AccentColor
	case TradeAccepted:
		statusText = "Status: Accepted! Completing trade..."
		statusColor = ui.AccentColor
	case TradeCommitted:
		statusText = "Status: Trade completed successfully!"
		statusColor = ui.AccentColor
	case TradeRejected:
		statusText = "Status: Trade rejected"
		statusColor = ui.ErrorColor
	case TradeCancelled:
		statusText = "Status: Trade cancelled"
		statusColor = ui.ErrorColor
	case TradeFailed:
		statusText = fmt.Sprintf("Status: Trade failed - %s", ui.FailureReason)
		statusColor = ui.ErrorColor
	}

	// Draw status text
	textX := ui.X + ui.Padding
	text.Draw(screen, statusText, ui.Font, textX, y+20, statusColor)
}

// drawTimeoutIndicator draws a progress bar showing time remaining.
func (ui *TradeDialogUI) drawTimeoutIndicator(screen *ebiten.Image) {
	elapsed := time.Since(ui.ProposalTime)
	remaining := ui.TimeoutDuration - elapsed
	if remaining < 0 {
		remaining = 0
	}

	progress := float32(remaining.Seconds()) / float32(ui.TimeoutDuration.Seconds())

	// Draw progress bar
	barY := ui.Y + ui.HeaderHeight + 2
	barWidth := ui.Width - 2*ui.Padding
	barHeight := float32(4)

	// Background
	vector.DrawFilledRect(screen, float32(ui.X+ui.Padding), float32(barY), float32(barWidth), barHeight, color.RGBA{50, 50, 50, 255}, false)

	// Progress
	progressColor := ui.AccentColor
	if progress < 0.3 {
		progressColor = ui.ErrorColor // Warning color when time running out
	}
	vector.DrawFilledRect(screen, float32(ui.X+ui.Padding), float32(barY), float32(barWidth)*progress, barHeight, progressColor, false)

	// Time remaining text
	timeText := fmt.Sprintf("%.0fs", remaining.Seconds())
	textX := ui.X + ui.Width - ui.Padding - len(timeText)*7
	text.Draw(screen, timeText, ui.Font, textX, barY+15, ui.TextColor)
}
