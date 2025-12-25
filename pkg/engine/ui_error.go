//go:build !headless
// +build !headless

// Package engine provides shared UI error display functionality.
// H-002 FIX: Standardized error feedback system for all UI components.
package engine

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

// UIErrorState manages temporary error message display for UI systems.
// H-002 FIX: Provides consistent error display across all UI components.
type UIErrorState struct {
	// Message is the current error text to display
	Message string

	// ShowUntil is the time when the error should stop displaying
	ShowUntil time.Time

	// DefaultDuration is how long errors display by default (3 seconds)
	DefaultDuration time.Duration
}

// NewUIErrorState creates a new error state manager with default settings.
func NewUIErrorState() *UIErrorState {
	return &UIErrorState{
		DefaultDuration: 3 * time.Second,
	}
}

// ShowError displays an error message for the default duration.
func (e *UIErrorState) ShowError(message string) {
	e.ShowErrorFor(message, e.DefaultDuration)
}

// ShowErrorFor displays an error message for a specific duration.
func (e *UIErrorState) ShowErrorFor(message string, duration time.Duration) {
	e.Message = message
	e.ShowUntil = time.Now().Add(duration)
}

// HasError returns true if there's an active error message to display.
func (e *UIErrorState) HasError() bool {
	return e.Message != "" && time.Now().Before(e.ShowUntil)
}

// Clear immediately clears any active error message.
func (e *UIErrorState) Clear() {
	e.Message = ""
	e.ShowUntil = time.Time{}
}

// DrawError renders the error message at the bottom center of the screen.
// Should be called from the UI system's Draw method when HasError() returns true.
func (e *UIErrorState) DrawError(screen *ebiten.Image) {
	if !e.HasError() {
		return
	}

	screenWidth := screen.Bounds().Dx()
	screenHeight := screen.Bounds().Dy()

	// Calculate error box dimensions
	messageWidth := len(e.Message) * 7 // Approximate width with 7x13 font
	boxWidth := messageWidth + 40      // Add padding
	boxHeight := 50
	boxX := (screenWidth - boxWidth) / 2
	boxY := screenHeight - boxHeight - 20 // 20px from bottom

	// Draw semi-transparent red background
	vector.DrawFilledRect(screen,
		float32(boxX), float32(boxY),
		float32(boxWidth), float32(boxHeight),
		color.RGBA{150, 30, 30, 220}, false)

	// Draw border
	vector.StrokeRect(screen,
		float32(boxX), float32(boxY),
		float32(boxWidth), float32(boxHeight),
		2, color.RGBA{255, 100, 100, 255}, false)

	// Draw error icon (⚠)
	iconColor := color.RGBA{255, 200, 0, 255}
	text.Draw(screen, "⚠", basicfont.Face7x13, boxX+10, boxY+30, iconColor)

	// Draw error message
	textColor := color.RGBA{255, 255, 255, 255}
	text.Draw(screen, e.Message, basicfont.Face7x13, boxX+30, boxY+30, textColor)

	// Calculate remaining time
	remaining := time.Until(e.ShowUntil).Seconds()
	if remaining > 0 {
		// Draw progress bar showing time remaining
		barWidth := float32(boxWidth - 20)
		barFilled := barWidth * float32(remaining) / float32(e.DefaultDuration.Seconds())

		barY := float32(boxY + boxHeight - 10)
		barX := float32(boxX + 10)

		// Background bar
		vector.DrawFilledRect(screen, barX, barY, barWidth, 4,
			color.RGBA{100, 100, 100, 150}, false)

		// Filled portion
		vector.DrawFilledRect(screen, barX, barY, barFilled, 4,
			color.RGBA{255, 150, 150, 200}, false)
	}
}

// DrawErrorAt renders the error message at a specific position.
// Useful for UI systems that need custom positioning.
func (e *UIErrorState) DrawErrorAt(screen *ebiten.Image, x, y, width int) {
	if !e.HasError() {
		return
	}

	boxHeight := 50

	// Draw semi-transparent red background
	vector.DrawFilledRect(screen,
		float32(x), float32(y),
		float32(width), float32(boxHeight),
		color.RGBA{150, 30, 30, 220}, false)

	// Draw border
	vector.StrokeRect(screen,
		float32(x), float32(y),
		float32(width), float32(boxHeight),
		2, color.RGBA{255, 100, 100, 255}, false)

	// Draw error icon (⚠)
	iconColor := color.RGBA{255, 200, 0, 255}
	text.Draw(screen, "⚠", basicfont.Face7x13, x+10, y+30, iconColor)

	// Draw error message (truncate if too long)
	maxChars := (width - 50) / 7
	message := e.Message
	if len(message) > maxChars {
		message = message[:maxChars-3] + "..."
	}

	textColor := color.RGBA{255, 255, 255, 255}
	text.Draw(screen, message, basicfont.Face7x13, x+30, y+30, textColor)
}
