package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/venture/pkg/social/persistence"
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
func (g *GalleryUI) Update() bool {
	if !g.Visible {
		return false
	}

	// INTEGRATION FIX [Category B]: V8.0 Gallery UI Input Handling
	// Gap: No input handling for image navigation
	// Fix: Added keyboard controls for browsing gallery images
	// Roadmap: ROADMAP_V8.md Phase 49.4
	// BUG FIX: Phase 2 - GalleryUI ESC key using IsKeyPressed instead of IsKeyJustPressed
	// Resolution: Changed to inpututil.IsKeyJustPressed to prevent repeated close on held key
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.Hide()
		return true
	}

	// Update total images from gallery
	if g.gallery != nil {
		g.TotalImages = g.gallery.GetImageCount()
	}

	// BUG FIX: Phase 2 - GalleryUI navigation keys using IsKeyPressed causing rapid scrolling
	// Resolution: Changed to IsKeyJustPressed for single-step navigation on key press
	// Navigate images
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		g.NextImage()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		g.PreviousImage()
	}

	// Consume input when UI is visible
	return true
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
