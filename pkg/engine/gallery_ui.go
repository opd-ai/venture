package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/venture/pkg/social/persistence"
)

// Gallery UI layout constants for maintainability
const (
	galleryCloseButtonWidth  = 40
	galleryCloseButtonHeight = 30
	galleryCloseButtonMargin = 10
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

	// INTEGRATION FIX [Category B]: V8.0 Gallery UI Integration (Phase 49.4)
	// Gap: UI defined but gallery manager never connected, no rendering implementation
	// Fix: Added gallery reference for image storage and viewing
	// Roadmap: ROADMAP_V8.md Phase 49.4
	gallery *persistence.ImageGallery
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
// Platform parity fix: Added touch/mouse support for navigation.
func (g *GalleryUI) Update() bool {
	if !g.Visible {
		return false
	}

	if g.handleEscapeKey() {
		return true
	}

	if g.handleTouchInput() {
		return true
	}

	g.updateTotalImages()
	g.handleKeyboardNavigation()

	return true
}

// handleEscapeKey processes escape key input to close the gallery.
func (g *GalleryUI) handleEscapeKey() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.Hide()
		return true
	}
	return false
}

// handleTouchInput processes touch and mouse input for navigation and closing.
func (g *GalleryUI) handleTouchInput() bool {
	if !IsTouchOrMouseJustPressed() {
		return false
	}

	mouseX, mouseY, _ := GetTouchOrMousePosition()
	if mouseY < g.Y || mouseY > g.Y+g.Height {
		return false
	}

	if g.isCloseButtonClicked(mouseX, mouseY) {
		g.Hide()
		return true
	}

	g.handleImageNavigation(mouseX)
	return false
}

// isCloseButtonClicked checks if the close button area was clicked.
func (g *GalleryUI) isCloseButtonClicked(mouseX, mouseY int) bool {
	closeX := g.X + g.Width - galleryCloseButtonWidth - galleryCloseButtonMargin
	closeY := g.Y + galleryCloseButtonMargin

	return mouseX >= closeX && mouseX <= closeX+galleryCloseButtonWidth &&
		mouseY >= closeY && mouseY <= closeY+galleryCloseButtonHeight
}

// handleImageNavigation processes left/right screen navigation for images.
func (g *GalleryUI) handleImageNavigation(mouseX int) {
	midX := g.X + g.Width/2

	if mouseX >= g.X && mouseX < midX {
		g.PreviousImage()
	} else if mouseX >= midX && mouseX <= g.X+g.Width {
		g.NextImage()
	}
}

// updateTotalImages refreshes the total image count from the gallery.
func (g *GalleryUI) updateTotalImages() {
	if g.gallery != nil {
		g.TotalImages = g.gallery.GetImageCount()
	}
}

// handleKeyboardNavigation processes arrow key and WASD navigation.
func (g *GalleryUI) handleKeyboardNavigation() {
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		g.NextImage()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		g.PreviousImage()
	}
}

// Draw renders the gallery UI to the screen.
func (g *GalleryUI) Draw(screen *ebiten.Image) {
	if !g.Visible || screen == nil {
		return
	}

	// INTEGRATION FIX [Category B]: V8.0 Gallery UI Rendering
	// Gap: No visual representation of image gallery
	// Fix: Added comprehensive UI rendering with image metadata and navigation controls
	// Roadmap: ROADMAP_V8.md Phase 49.4

	// Import required for rendering
	// Note: Actual image display requires decoding base64 data and converting to ebiten.Image
	// For MVP, we show metadata; full image rendering can be added in enhancement phase

	ebitenutil.DebugPrintAt(screen, "Image Gallery", g.X+10, g.Y+10)

	if g.gallery != nil {
		// Show image count
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Images: %d / %d", g.SelectedIndex+1, g.TotalImages), g.X+10, g.Y+40)

		// Get current image metadata
		images := g.gallery.GetAllImages()
		if g.SelectedIndex >= 0 && g.SelectedIndex < len(images) {
			img := images[g.SelectedIndex]

			// Display image metadata
			metadataY := g.Y + 70
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Title: %s", img.Title), g.X+10, metadataY)
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Format: %s", img.Format), g.X+10, metadataY+20)
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Size: %dx%d (%d bytes)", img.Width, img.Height, img.SizeBytes), g.X+10, metadataY+40)
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Date: %s", img.Timestamp.Format("2006-01-02 15:04")), g.X+10, metadataY+60)

			if len(img.Tags) > 0 {
				ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Tags: %v", img.Tags), g.X+10, metadataY+80)
			}
		}
	} else {
		ebitenutil.DebugPrintAt(screen, "No gallery connected", g.X+10, g.Y+40)
	}

	// Draw controls
	controlsY := g.Y + g.Height - 30
	ebitenutil.DebugPrintAt(screen, "ESC: Close | Left/Right or A/D: Navigate", g.X+10, controlsY)
}

// SetGallery sets the image gallery reference.
func (g *GalleryUI) SetGallery(gallery interface{}) {
	// INTEGRATION FIX [Category B]: V8.0 Gallery UI Manager Wiring
	// Gap: Gallery passed but never stored or used
	// Fix: Type-assert and store gallery reference for image viewing
	// Roadmap: ROADMAP_V8.md Phase 49.4
	if ig, ok := gallery.(*persistence.ImageGallery); ok {
		g.gallery = ig
		g.TotalImages = ig.GetImageCount()
		if g.TotalImages > 0 {
			g.SelectedIndex = 0
		}
	}
}
