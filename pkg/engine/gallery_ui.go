package engine

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// GalleryUI represents an image gallery viewer interface.
// Phase 49.4
type GalleryUI struct {
	X, Y          int
	Width, Height int
	GenreID       string
	Visible       bool

	// Navigation state
	SelectedIndex int
	TotalImages   int

	// Colors
	BackgroundColor color.Color
	TextColor       color.Color
	HighlightColor  color.Color
}

// NewGalleryUI creates a new gallery UI instance.
func NewGalleryUI(screenWidth, screenHeight int) *GalleryUI {
	width := screenWidth - 40
	height := screenHeight - 100
	x := 20
	y := 50

	return &GalleryUI{
		X:               x,
		Y:               y,
		Width:           width,
		Height:          height,
		GenreID:         "fantasy",
		Visible:         false,
		SelectedIndex:   0,
		TotalImages:     0,
		BackgroundColor: color.RGBA{20, 20, 30, 230},
		TextColor:       color.RGBA{220, 220, 220, 255},
		HighlightColor:  color.RGBA{100, 150, 255, 255},
	}
}

// Toggle toggles the visibility of the gallery UI.
func (g *GalleryUI) Toggle() {
	g.Visible = !g.Visible
}

// Show makes the gallery UI visible.
func (g *GalleryUI) Show() {
	g.Visible = true
}

// Hide makes the gallery UI invisible.
func (g *GalleryUI) Hide() {
	g.Visible = false
}

// NextImage moves to the next image in the gallery.
func (g *GalleryUI) NextImage() {
	if g.TotalImages > 0 {
		g.SelectedIndex = (g.SelectedIndex + 1) % g.TotalImages
	}
}

// PreviousImage moves to the previous image in the gallery.
func (g *GalleryUI) PreviousImage() {
	if g.TotalImages > 0 {
		g.SelectedIndex--
		if g.SelectedIndex < 0 {
			g.SelectedIndex = g.TotalImages - 1
		}
	}
}

// Update updates the gallery UI state.
// Returns true if the UI consumed the input (blocking pass-through).
func (g *GalleryUI) Update() bool {
	if !g.Visible {
		return false
	}
	// Minimal stub implementation - will be enhanced in Phase 49.4
	return true
}

// Draw renders the gallery UI to the screen.
func (g *GalleryUI) Draw(screen *ebiten.Image) {
	if !g.Visible || screen == nil {
		return
	}
	// Minimal stub implementation - will be enhanced in Phase 49.4
}

// SetGallery sets the image gallery reference.
func (g *GalleryUI) SetGallery(gallery interface{}) {
	// Minimal stub implementation - will be enhanced in Phase 49.4
}
