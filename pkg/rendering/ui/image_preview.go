// Package ui provides image preview UI rendering functionality.
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

// ImagePreviewUI renders image preview and download confirmation dialogs.
type ImagePreviewUI struct {
	// Position and size
	X, Y          int
	Width, Height int

	// Image state
	Active           bool
	SenderName       string
	Timestamp        time.Time
	ThumbnailImage   *ebiten.Image
	FullImage        *ebiten.Image
	ImageWidth       int
	ImageHeight      int
	ImageSize        int // Size in bytes
	ExpiryTime       time.Time
	DownloadProgress float32 // 0.0 to 1.0
	ShowFullImage    bool    // Whether full image is being displayed
	ConfirmSelected  bool    // Whether "Download" is selected (vs "Decline")

	// Visual settings
	BackgroundColor color.Color
	BorderColor     color.Color
	HeaderColor     color.Color
	TextColor       color.Color
	AccentColor     color.Color
	Font            font.Face

	// Layout
	HeaderHeight int
	ButtonHeight int
	Padding      int
	MaxImageSize int // Maximum size for display (scaled if larger)
}

// NewImagePreviewUI creates a new image preview UI instance.
func NewImagePreviewUI(x, y, width, height int) *ImagePreviewUI {
	return &ImagePreviewUI{
		X:               x,
		Y:               y,
		Width:           width,
		Height:          height,
		Active:          false,
		BackgroundColor: color.RGBA{20, 20, 30, 240},
		BorderColor:     color.RGBA{100, 150, 200, 255},
		HeaderColor:     color.RGBA{40, 50, 70, 255},
		TextColor:       color.RGBA{220, 220, 220, 255},
		AccentColor:     color.RGBA{100, 200, 255, 255},
		Font:            basicfont.Face7x13,
		HeaderHeight:    30,
		ButtonHeight:    35,
		Padding:         10,
		MaxImageSize:    512, // Scale down images larger than 512x512
		ConfirmSelected: true,
	}
}

// ShowThumbnail displays an image thumbnail with download prompt.
func (ui *ImagePreviewUI) ShowThumbnail(senderName string, timestamp time.Time, thumbnail *ebiten.Image, imageWidth, imageHeight, imageSize int, expiryTime time.Time) {
	ui.Active = true
	ui.SenderName = senderName
	ui.Timestamp = timestamp
	ui.ThumbnailImage = thumbnail
	ui.FullImage = nil
	ui.ImageWidth = imageWidth
	ui.ImageHeight = imageHeight
	ui.ImageSize = imageSize
	ui.ExpiryTime = expiryTime
	ui.DownloadProgress = 0.0
	ui.ShowFullImage = false
	ui.ConfirmSelected = true
}

// ShowFullImage displays the downloaded full image.
func (ui *ImagePreviewUI) ShowFullImage(fullImage *ebiten.Image) {
	ui.FullImage = fullImage
	ui.ShowFullImage = true
	ui.DownloadProgress = 1.0
}

// UpdateDownloadProgress updates the download progress indicator.
func (ui *ImagePreviewUI) UpdateDownloadProgress(progress float32) {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	ui.DownloadProgress = progress
}

// Hide closes the image preview dialog.
func (ui *ImagePreviewUI) Hide() {
	ui.Active = false
	ui.ThumbnailImage = nil
	ui.FullImage = nil
}

// ToggleSelection switches between Download and Decline buttons.
func (ui *ImagePreviewUI) ToggleSelection() {
	ui.ConfirmSelected = !ui.ConfirmSelected
}

// GetSelectedAction returns the currently selected action ("download" or "decline").
func (ui *ImagePreviewUI) GetSelectedAction() string {
	if ui.ConfirmSelected {
		return "download"
	}
	return "decline"
}

// IsExpired returns whether the image has expired.
func (ui *ImagePreviewUI) IsExpired() bool {
	return time.Now().After(ui.ExpiryTime)
}

// GetTimeRemaining returns the time remaining until expiry.
func (ui *ImagePreviewUI) GetTimeRemaining() time.Duration {
	remaining := time.Until(ui.ExpiryTime)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Update updates the image preview state.
func (ui *ImagePreviewUI) Update(deltaTime float64) {
	if !ui.Active {
		return
	}

	// Auto-hide if expired
	if ui.IsExpired() && !ui.ShowFullImage {
		ui.Hide()
	}
}

// Render draws the image preview dialog to the screen.
func (ui *ImagePreviewUI) Render(screen *ebiten.Image) {
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

	// Draw image preview
	previewY := ui.Y + ui.HeaderHeight + ui.Padding
	previewHeight := ui.Height - ui.HeaderHeight - ui.ButtonHeight - 3*ui.Padding
	ui.drawImagePreview(screen, previewY, previewHeight)

	// Draw buttons or download progress
	buttonsY := ui.Y + ui.Height - ui.ButtonHeight - ui.Padding
	if ui.ShowFullImage {
		ui.drawCloseButton(screen, buttonsY)
	} else if ui.DownloadProgress > 0 && ui.DownloadProgress < 1 {
		ui.drawDownloadProgress(screen, buttonsY)
	} else {
		ui.drawButtons(screen, buttonsY)
	}
}

// drawPanel draws a filled rectangle with border.
func (ui *ImagePreviewUI) drawPanel(screen *ebiten.Image, x, y, width, height int, bgColor, borderColor color.Color) {
	// Background
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(width), float32(height), bgColor, false)

	// Border (2px thick)
	vector.StrokeRect(screen, float32(x), float32(y), float32(width), float32(height), 2, borderColor, false)
}

// drawHeader draws the dialog header with sender and timestamp.
func (ui *ImagePreviewUI) drawHeader(screen *ebiten.Image) {
	headerY := ui.Y

	// Header background
	vector.DrawFilledRect(screen, float32(ui.X), float32(headerY), float32(ui.Width), float32(ui.HeaderHeight), ui.HeaderColor, false)

	// Title text
	title := fmt.Sprintf("Image from %s - %s", ui.SenderName, ui.Timestamp.Format("15:04"))
	text.Draw(screen, title, ui.Font, ui.X+ui.Padding, headerY+20, ui.TextColor)

	// Expiry indicator
	remaining := ui.GetTimeRemaining()
	expiryText := fmt.Sprintf("Expires in %.0fs", remaining.Seconds())
	expiryX := ui.X + ui.Width - ui.Padding - len(expiryText)*7
	expiryColor := ui.TextColor
	if remaining < 60*time.Second {
		expiryColor = color.RGBA{255, 150, 100, 255} // Warning color
	}
	text.Draw(screen, expiryText, ui.Font, expiryX, headerY+20, expiryColor)
}

// drawImagePreview draws the thumbnail or full image.
func (ui *ImagePreviewUI) drawImagePreview(screen *ebiten.Image, y, height int) {
	// Image area background
	previewBG := color.RGBA{10, 10, 15, 255}
	ui.drawPanel(screen, ui.X+ui.Padding, y, ui.Width-2*ui.Padding, height, previewBG, ui.BorderColor)

	// Draw image (thumbnail or full)
	var img *ebiten.Image
	if ui.ShowFullImage && ui.FullImage != nil {
		img = ui.FullImage
	} else if ui.ThumbnailImage != nil {
		img = ui.ThumbnailImage
	}

	if img != nil {
		// Calculate scaling to fit preview area
		imgBounds := img.Bounds()
		imgW := imgBounds.Dx()
		imgH := imgBounds.Dy()

		maxW := ui.Width - 4*ui.Padding
		maxH := height - 2*ui.Padding

		scale := 1.0
		if imgW > maxW || imgH > maxH {
			scaleW := float64(maxW) / float64(imgW)
			scaleH := float64(maxH) / float64(imgH)
			if scaleW < scaleH {
				scale = scaleW
			} else {
				scale = scaleH
			}
		}

		// Center image in preview area
		scaledW := int(float64(imgW) * scale)
		scaledH := int(float64(imgH) * scale)
		imgX := ui.X + (ui.Width-scaledW)/2
		imgY := y + (height-scaledH)/2

		// Draw scaled image
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Scale(scale, scale)
		opts.GeoM.Translate(float64(imgX), float64(imgY))
		screen.DrawImage(img, opts)

		// Draw image info below
		infoY := y + height - 20
		if !ui.ShowFullImage {
			infoText := fmt.Sprintf("Thumbnail (Full: %dx%d, %.1f KB)", ui.ImageWidth, ui.ImageHeight, float64(ui.ImageSize)/1024)
			text.Draw(screen, infoText, ui.Font, ui.X+ui.Padding*2, infoY, ui.TextColor)
		}
	} else {
		// No image available
		noImageText := "Loading..."
		textX := ui.X + (ui.Width-len(noImageText)*7)/2
		textY := y + height/2
		text.Draw(screen, noImageText, ui.Font, textX, textY, ui.TextColor)
	}
}

// drawButtons draws Download/Decline buttons.
func (ui *ImagePreviewUI) drawButtons(screen *ebiten.Image, y int) {
	buttonWidth := (ui.Width - 3*ui.Padding) / 2
	downloadX := ui.X + ui.Padding
	declineX := ui.X + ui.Width/2 + ui.Padding/2

	// Download button
	downloadColor := ui.AccentColor
	if ui.ConfirmSelected {
		downloadColor = color.RGBA{150, 220, 255, 255} // Highlight
	}
	ui.drawButton(screen, downloadX, y, buttonWidth, ui.ButtonHeight, "Download", downloadColor)

	// Decline button
	declineColor := color.RGBA{180, 80, 80, 255}
	if !ui.ConfirmSelected {
		declineColor = color.RGBA{220, 120, 120, 255} // Highlight
	}
	ui.drawButton(screen, declineX, y, buttonWidth, ui.ButtonHeight, "Decline", declineColor)
}

// drawButton draws a button with text.
func (ui *ImagePreviewUI) drawButton(screen *ebiten.Image, x, y, width, height int, label string, bgColor color.Color) {
	// Button background
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(width), float32(height), bgColor, false)

	// Button border
	vector.StrokeRect(screen, float32(x), float32(y), float32(width), float32(height), 1, ui.BorderColor, false)

	// Button text (centered)
	textWidth := len(label) * 7 // Approximate width with 7x13 font
	textX := x + (width-textWidth)/2
	textY := y + height/2 + 5
	text.Draw(screen, label, ui.Font, textX, textY, color.RGBA{255, 255, 255, 255})
}

// drawDownloadProgress draws a progress bar for image download.
func (ui *ImagePreviewUI) drawDownloadProgress(screen *ebiten.Image, y int) {
	barWidth := ui.Width - 2*ui.Padding
	barHeight := ui.ButtonHeight

	// Background bar
	vector.DrawFilledRect(screen, float32(ui.X+ui.Padding), float32(y), float32(barWidth), float32(barHeight), color.RGBA{50, 50, 60, 255}, false)

	// Progress bar
	progressWidth := float32(barWidth) * ui.DownloadProgress
	vector.DrawFilledRect(screen, float32(ui.X+ui.Padding), float32(y), progressWidth, float32(barHeight), ui.AccentColor, false)

	// Border
	vector.StrokeRect(screen, float32(ui.X+ui.Padding), float32(y), float32(barWidth), float32(barHeight), 1, ui.BorderColor, false)

	// Progress text
	progressText := fmt.Sprintf("Downloading... %.0f%%", ui.DownloadProgress*100)
	textX := ui.X + (ui.Width-len(progressText)*7)/2
	textY := y + barHeight/2 + 5
	text.Draw(screen, progressText, ui.Font, textX, textY, ui.TextColor)
}

// drawCloseButton draws a close button for full image view.
func (ui *ImagePreviewUI) drawCloseButton(screen *ebiten.Image, y int) {
	buttonWidth := ui.Width - 2*ui.Padding
	closeX := ui.X + ui.Padding

	ui.drawButton(screen, closeX, y, buttonWidth, ui.ButtonHeight, "Close", ui.AccentColor)
}
