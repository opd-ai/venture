package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// LoadingUI displays world generation progress during async terrain loading.
type LoadingUI struct {
	screenWidth  int
	screenHeight int
	progress     float64 // 0.0 to 1.0
	message      string
}

// NewLoadingUI creates a new loading screen UI.
func NewLoadingUI(screenWidth, screenHeight int) *LoadingUI {
	return &LoadingUI{
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		progress:     0.0,
		message:      "Generating world...",
	}
}

// SetProgress updates the loading progress (0.0 to 1.0).
func (l *LoadingUI) SetProgress(progress float64) {
	l.progress = progress
	if progress < 0.0 {
		l.progress = 0.0
	}
	if progress > 1.0 {
		l.progress = 1.0
	}
}

// SetMessage updates the loading message.
func (l *LoadingUI) SetMessage(message string) {
	l.message = message
}

// Draw renders the loading screen with progress bar.
func (l *LoadingUI) Draw(screen *ebiten.Image) {
	// Clear screen with dark background
	screen.Fill(color.RGBA{20, 20, 30, 255})

	// Calculate center positions
	centerX := l.screenWidth / 2
	centerY := l.screenHeight / 2

	// Draw message
	msgBounds := text.BoundString(basicfont.Face7x13, l.message)
	msgX := centerX - msgBounds.Dx()/2
	msgY := centerY - 40
	text.Draw(screen, l.message, basicfont.Face7x13, msgX, msgY, color.White)

	// Draw progress percentage
	percentText := fmt.Sprintf("%d%%", int(l.progress*100))
	pctBounds := text.BoundString(basicfont.Face7x13, percentText)
	pctX := centerX - pctBounds.Dx()/2
	pctY := centerY + 40
	text.Draw(screen, percentText, basicfont.Face7x13, pctX, pctY, color.White)

	// Draw progress bar
	barWidth := 400
	barHeight := 20
	barX := centerX - barWidth/2
	barY := centerY - barHeight/2

	// Background bar (outline)
	l.drawRect(screen, barX-2, barY-2, barWidth+4, barHeight+4, color.RGBA{100, 100, 100, 255})
	l.drawRect(screen, barX, barY, barWidth, barHeight, color.RGBA{40, 40, 50, 255})

	// Filled portion
	fillWidth := int(float64(barWidth) * l.progress)
	if fillWidth > 0 {
		l.drawRect(screen, barX, barY, fillWidth, barHeight, color.RGBA{80, 160, 255, 255})
	}
}

// drawRect is a helper to draw a filled rectangle.
func (l *LoadingUI) drawRect(screen *ebiten.Image, x, y, width, height int, c color.Color) {
	img := ebiten.NewImage(width, height)
	defer img.Dispose()
	img.Fill(c)
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, opts)
}
