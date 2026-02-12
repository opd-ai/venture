// Package ui provides procedural UI element generation.
// This file implements UI element generators for menus, buttons,
// panels, and other interface components.
package ui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"strings"

	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/sirupsen/logrus"
)

// Generator creates procedural UI elements.
type Generator struct {
	paletteGen *palette.Generator
	logger     *logrus.Entry
}

// NewGenerator creates a new UI element generator.
func NewGenerator() *Generator {
	return NewGeneratorWithLogger(nil)
}

// NewGeneratorWithLogger creates a new UI element generator with a logger.
func NewGeneratorWithLogger(logger *logrus.Logger) *Generator {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"generator": "ui",
		})
	}
	return &Generator{
		paletteGen: palette.NewGenerator(),
		logger:     logEntry,
	}
}

// Generate creates a UI element image from the given configuration.
func (g *Generator) Generate(config Config) (*image.RGBA, error) {
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"type":           config.Type,
			"genreID":        config.GenreID,
			"seed":           config.Seed,
			"width":          config.Width,
			"height":         config.Height,
			"hierarchyLevel": config.HierarchyLevel,
		}).Debug("generating UI element")
	}

	if err := g.validateConfigs(config); err != nil {
		return nil, err
	}

	rng := rand.New(rand.NewSource(config.Seed))
	pal, err := g.generatePalette(config)
	if err != nil {
		return nil, err
	}

	img := image.NewRGBA(image.Rect(0, 0, config.Width, config.Height))
	if err := g.generateElementByType(img, pal, rng, config); err != nil {
		return nil, err
	}

	if config.Transition != nil && config.Transition.Type != TransitionNone {
		img = g.ApplyTransition(img, *config.Transition)
	}

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"type":      config.Type,
			"seed":      config.Seed,
			"hierarchy": config.HierarchyLevel,
		}).Info("UI element generated")
	}

	return img, nil
}

// validateConfigs validates the main config and transition config if present.
func (g *Generator) validateConfigs(config Config) error {
	if err := config.Validate(); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("invalid UI config")
		}
		return fmt.Errorf("invalid config: %w", err)
	}

	if config.Transition != nil {
		if err := config.Transition.Validate(); err != nil {
			if g.logger != nil {
				g.logger.WithError(err).Error("invalid transition config")
			}
			return fmt.Errorf("invalid transition config: %w", err)
		}
	}

	return nil
}

// generatePalette creates a color palette for the given config.
func (g *Generator) generatePalette(config Config) (*palette.Palette, error) {
	pal, err := g.paletteGen.Generate(config.GenreID, config.Seed)
	if err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("palette generation failed")
		}
		return nil, fmt.Errorf("failed to generate palette: %w", err)
	}
	return pal, nil
}

// generateElementByType generates the UI element based on its type.
func (g *Generator) generateElementByType(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config Config) error {
	switch config.Type {
	case ElementButton:
		g.generateButton(img, pal, rng, config)
	case ElementPanel:
		g.generatePanel(img, pal, rng, config)
	case ElementHealthBar:
		g.generateHealthBar(img, pal, rng, config)
	case ElementLabel:
		g.generateLabel(img, pal, rng, config)
	case ElementIcon:
		g.generateIcon(img, pal, rng, config)
	case ElementFrame:
		g.generateFrame(img, pal, rng, config)
	default:
		err := fmt.Errorf("unknown element type: %d", config.Type)
		if g.logger != nil {
			g.logger.WithError(err).WithField("type", config.Type).Error("unknown element type")
		}
		return err
	}
	return nil
}

// generateButton creates a button UI element.
func (g *Generator) generateButton(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config Config) {
	// Get hierarchy style
	style := GetHierarchyStyle(config.HierarchyLevel)

	// Select colors based on state
	var bgColor, borderColor color.Color

	// Select a base color with good contrast potential
	// Try to find a color with mid-range luminance for better state differentiation
	baseColor := g.selectButtonBaseColor(pal, rng)

	switch config.State {
	case StateNormal:
		bgColor = baseColor
		borderColor = g.darkenColor(baseColor, 0.3)
	case StateHover:
		bgColor = g.lightenColor(baseColor, 0.2)
		borderColor = g.darkenColor(baseColor, 0.2)
	case StatePressed:
		bgColor = g.darkenColor(baseColor, 0.2)
		borderColor = g.darkenColor(baseColor, 0.4)
	case StateDisabled:
		bgColor = pal.Background
		borderColor = g.lightenColor(pal.Background, 0.2)
	}

	// Apply hierarchy opacity
	bgColor = ApplyHierarchyOpacity(bgColor, config.HierarchyLevel)

	// Fill background
	g.fillRect(img, 0, 0, config.Width, config.Height, bgColor)

	// Draw border based on genre
	borderStyle := g.selectBorderStyle(config.GenreID)
	borderThickness := style.BorderThickness
	g.drawBorder(img, borderColor, borderStyle, borderThickness)

	// Add highlight if not disabled
	// Phase 45: Scaled 2× for 64×64 UI elements (2 → 4, 3 → 6)
	if config.State != StateDisabled {
		highlightColor := g.lightenColor(bgColor, 0.4)
		g.drawLine(img, 4, 4, config.Width-6, 4, highlightColor)
	}
}

// generatePanel creates a panel UI element.
func (g *Generator) generatePanel(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config Config) {
	// Get alpha from custom config, default to 200
	alpha := 200
	if customAlpha, ok := config.Custom["alpha"].(int); ok {
		// Clamp to valid range [0, 255]
		if customAlpha < 0 {
			alpha = 0
		} else if customAlpha > 255 {
			alpha = 255
		} else {
			alpha = customAlpha
		}
	}

	// Semi-transparent background
	bgColor := pal.Background
	r, gr, b, _ := bgColor.RGBA()
	semiTransparent := color.RGBA{
		R: uint8(r >> 8),
		G: uint8(gr >> 8),
		B: uint8(b >> 8),
		A: uint8(alpha),
	}

	// Fill background
	g.fillRect(img, 0, 0, config.Width, config.Height, semiTransparent)

	// Draw border
	// Phase 45: Scaled 2× for 64×64 UI elements (1 → 2)
	borderColor := g.lightenColor(pal.Background, 0.3)
	g.drawBorder(img, borderColor, BorderSolid, 2)
}

// generateHealthBar creates a health/progress bar.
// Phase 45: Padding and borders scaled 2× for 64×64 UI elements.
func (g *Generator) generateHealthBar(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config Config) {
	// Background
	bgColor := g.darkenColor(pal.Background, 0.2)
	g.fillRect(img, 0, 0, config.Width, config.Height, bgColor)

	// Calculate filled width based on value
	// Phase 45: Padding 4 → 8 for 64×64 scaling
	filledWidth := int(float64(config.Width-8) * config.Value)

	// Select fill color based on value
	var fillColor color.Color
	if config.Value > 0.6 {
		fillColor = pal.Success // Green for high health
	} else if config.Value > 0.3 {
		fillColor = color.RGBA{255, 200, 0, 255} // Yellow for medium
	} else {
		fillColor = pal.Danger // Red for low health
	}

	// Fill bar
	// Phase 45: Positions 2 → 4, padding 4 → 8 for 64×64 scaling
	if filledWidth > 0 {
		g.fillRect(img, 4, 4, filledWidth, config.Height-8, fillColor)

		// Add shine effect (positioned proportionally at 20% from top)
		shineY := 4 + maxInt(2, config.Height/5) // At least 2px from fill start (scaled)
		shineColor := g.lightenColor(fillColor, 0.3)
		g.drawLine(img, 4, shineY, filledWidth, shineY, shineColor)
	}

	// Border
	// Phase 45: Scaled 2× for 64×64 UI elements (1 → 2)
	borderColor := g.lightenColor(pal.Background, 0.4)
	g.drawBorder(img, borderColor, BorderSolid, 2)
}

// generateLabel creates a text label background.
// Phase 45: Padding and borders scaled 2× for 64×64 UI elements.
func (g *Generator) generateLabel(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config Config) {
	// Semi-transparent background
	bgColor := pal.Background
	r, gr, b, _ := bgColor.RGBA()
	labelBg := color.RGBA{
		R: uint8(r >> 8),
		G: uint8(gr >> 8),
		B: uint8(b >> 8),
		A: 180,
	}

	// Fill with slight padding
	// Phase 45: Padding 1 → 2 for 64×64 scaling
	g.fillRect(img, 2, 2, config.Width-4, config.Height-4, labelBg)

	// Optional border for emphasis
	// Phase 45: Border 1 → 2 for 64×64 scaling
	if config.State == StateHover {
		borderColor := pal.Primary
		g.drawBorder(img, borderColor, BorderSolid, 2)
	}
}

// generateIcon creates a small iconic UI element.
// Phase 45: Padding and borders scaled 2× for 64×64 UI elements.
// Default icon size updated from 16×16 to 32×32.
func (g *Generator) generateIcon(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config Config) {
	// Background circle or square based on genre
	bgColor := pal.Primary

	// Determine if this should be a square icon based on genre
	// Tech/futuristic genres use squares, organic/natural genres use circles
	useSquare := g.isTechGenre(config.GenreID)

	// Phase 45: Padding 2 → 4 for 64×64 scaling
	if useSquare {
		// Square icon for tech genres
		g.fillRect(img, 4, 4, config.Width-8, config.Height-8, bgColor)
	} else {
		// Circular icon for organic/natural genres
		centerX := config.Width / 2
		centerY := config.Height / 2
		radius := config.Width/2 - 4 // Phase 45: 2 → 4 for 64×64 scaling
		g.drawCircle(img, centerX, centerY, radius, bgColor, true)
	}

	// Border
	// Phase 45: Border 1 → 2 for 64×64 scaling
	borderColor := g.darkenColor(bgColor, 0.3)
	g.drawBorder(img, borderColor, BorderSolid, 2)
}

// generateFrame creates a decorative frame.
// Phase 45: Border thicknesses and corner sizes scaled 2× for 64×64 UI elements.
func (g *Generator) generateFrame(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config Config) {
	// Frame is mostly transparent with ornate border
	borderColor := pal.Primary
	borderStyle := g.selectBorderStyle(config.GenreID)

	// Draw double border for emphasis
	// Phase 45: Border 3 → 6 for 64×64 scaling
	g.drawBorder(img, borderColor, borderStyle, 6)

	// Inner border
	// Phase 45: Border 1 → 2 for 64×64 scaling
	innerColor := g.lightenColor(borderColor, 0.2)
	g.drawBorder(img, innerColor, BorderSolid, 2)

	// Corner decorations for ornate genres
	// Phase 45: Corner size 4 → 8 for 64×64 scaling
	if config.GenreID == "fantasy" || config.GenreID == "horror" {
		cornerSize := 8
		g.fillRect(img, 0, 0, cornerSize, cornerSize, borderColor)
		g.fillRect(img, config.Width-cornerSize, 0, cornerSize, cornerSize, borderColor)
		g.fillRect(img, 0, config.Height-cornerSize, cornerSize, cornerSize, borderColor)
		g.fillRect(img, config.Width-cornerSize, config.Height-cornerSize, cornerSize, cornerSize, borderColor)
	}
}

// Helper methods

func (g *Generator) fillRect(img *image.RGBA, x, y, w, h int, col color.Color) {
	bounds := img.Bounds()
	for py := y; py < y+h && py < bounds.Max.Y; py++ {
		for px := x; px < x+w && px < bounds.Max.X; px++ {
			if px >= bounds.Min.X && py >= bounds.Min.Y {
				img.Set(px, py, col)
			}
		}
	}
}

func (g *Generator) drawBorder(img *image.RGBA, col color.Color, style BorderStyle, thickness int) {
	switch style {
	case BorderSolid:
		g.drawSolidBorder(img, col, thickness)
	case BorderDouble:
		g.drawDoubleBorder(img, col)
	case BorderOrnate:
		g.drawOrnateBorder(img, col, thickness)
	case BorderGlow:
		g.drawGlowBorder(img, col)
	case BorderDashed:
		g.drawDashedBorder(img, col, thickness)
	case BorderDotted:
		g.drawDottedBorder(img, col)
	case BorderEmbossed:
		g.drawEmbossedBorder(img, col, thickness)
	case BorderEngraved:
		g.drawEngravedBorder(img, col, thickness)
	}
}

// drawSolidBorder draws a simple solid rectangular border.
func (g *Generator) drawSolidBorder(img *image.RGBA, col color.Color, thickness int) {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	for t := 0; t < thickness; t++ {
		// Top and bottom
		for x := 0; x < w; x++ {
			img.Set(x, t, col)
			img.Set(x, h-t-1, col)
		}
		// Left and right
		for y := 0; y < h; y++ {
			img.Set(t, y, col)
			img.Set(w-t-1, y, col)
		}
	}
}

// drawDoubleBorder draws two parallel lines with 2px gap.
func (g *Generator) drawDoubleBorder(img *image.RGBA, col color.Color) {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	for x := 0; x < w; x++ {
		img.Set(x, 0, col)
		img.Set(x, 2, col)
		img.Set(x, h-3, col)
		img.Set(x, h-1, col)
	}
	for y := 0; y < h; y++ {
		img.Set(0, y, col)
		img.Set(2, y, col)
		img.Set(w-3, y, col)
		img.Set(w-1, y, col)
	}
}

// drawOrnateBorder draws a solid border with corner decorations.
func (g *Generator) drawOrnateBorder(img *image.RGBA, col color.Color, thickness int) {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	g.drawSolidBorder(img, col, thickness)
	cornerSize := 4
	for dy := 0; dy < cornerSize; dy++ {
		for dx := 0; dx < cornerSize; dx++ {
			if dx < w && dy < h {
				img.Set(dx, dy, col) // Top-left
			}
			if w-cornerSize+dx < w && dy < h {
				img.Set(w-cornerSize+dx, dy, col) // Top-right
			}
			if dx < w && h-cornerSize+dy < h {
				img.Set(dx, h-cornerSize+dy, col) // Bottom-left
			}
			if w-cornerSize+dx < w && h-cornerSize+dy < h {
				img.Set(w-cornerSize+dx, h-cornerSize+dy, col) // Bottom-right
			}
		}
	}
}

// drawGlowBorder draws a gradient fade from opaque to transparent over 5 pixels.
func (g *Generator) drawGlowBorder(img *image.RGBA, col color.Color) {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	r, gr, b, _ := col.RGBA()
	for t := 0; t < 5; t++ {
		alpha := uint8(255 - t*51) // Fade: 255, 204, 153, 102, 51
		glowCol := color.RGBA{
			R: uint8(r >> 8),
			G: uint8(gr >> 8),
			B: uint8(b >> 8),
			A: alpha,
		}
		for x := 0; x < w; x++ {
			if t < h {
				img.Set(x, t, glowCol)
				img.Set(x, h-t-1, glowCol)
			}
		}
		for y := 0; y < h; y++ {
			if t < w {
				img.Set(t, y, glowCol)
				img.Set(w-t-1, y, glowCol)
			}
		}
	}
}

// drawDashedBorder draws a dashed border pattern.
func (g *Generator) drawDashedBorder(img *image.RGBA, col color.Color, thickness int) {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	dashLength := 6
	gapLength := 4

	g.drawHorizontalDashedBorders(img, col, thickness, w, h, dashLength, gapLength)
	g.drawVerticalDashedBorders(img, col, thickness, w, h, dashLength, gapLength)
}

// drawHorizontalDashedBorders draws dashed top and bottom borders.
func (g *Generator) drawHorizontalDashedBorders(img *image.RGBA, col color.Color, thickness, w, h, dashLength, gapLength int) {
	for x := 0; x < w; x += dashLength + gapLength {
		g.drawHorizontalDashSegment(img, col, thickness, x, w, h, dashLength)
	}
}

// drawHorizontalDashSegment draws a single dash segment on top and bottom borders.
func (g *Generator) drawHorizontalDashSegment(img *image.RGBA, col color.Color, thickness, x, w, h, dashLength int) {
	for dx := 0; dx < dashLength && x+dx < w; dx++ {
		g.drawHorizontalDashPixels(img, col, thickness, x+dx, h)
	}
}

// drawHorizontalDashPixels draws thickness pixels vertically at position x for top and bottom borders.
func (g *Generator) drawHorizontalDashPixels(img *image.RGBA, col color.Color, thickness, x, h int) {
	for t := 0; t < thickness; t++ {
		if t < h {
			img.Set(x, t, col)     // Top
			img.Set(x, h-t-1, col) // Bottom
		}
	}
}

// drawVerticalDashedBorders draws dashed left and right borders.
func (g *Generator) drawVerticalDashedBorders(img *image.RGBA, col color.Color, thickness, w, h, dashLength, gapLength int) {
	for y := 0; y < h; y += dashLength + gapLength {
		g.drawVerticalDashSegment(img, col, thickness, y, w, h, dashLength)
	}
}

// drawVerticalDashSegment draws a single dash segment on left and right borders.
func (g *Generator) drawVerticalDashSegment(img *image.RGBA, col color.Color, thickness, y, w, h, dashLength int) {
	for dy := 0; dy < dashLength && y+dy < h; dy++ {
		g.drawVerticalDashPixels(img, col, thickness, y+dy, w)
	}
}

// drawVerticalDashPixels draws thickness pixels horizontally at position y for left and right borders.
func (g *Generator) drawVerticalDashPixels(img *image.RGBA, col color.Color, thickness, y, w int) {
	for t := 0; t < thickness; t++ {
		if t < w {
			img.Set(t, y, col)     // Left
			img.Set(w-t-1, y, col) // Right
		}
	}
}

// drawDottedBorder draws a dotted border pattern.
func (g *Generator) drawDottedBorder(img *image.RGBA, col color.Color) {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	dotSize := 2
	gapSize := 3
	// Top and bottom borders
	for x := 0; x < w; x += dotSize + gapSize {
		for dx := 0; dx < dotSize && x+dx < w; dx++ {
			for dy := 0; dy < dotSize && dy < h; dy++ {
				img.Set(x+dx, dy, col)     // Top
				img.Set(x+dx, h-dy-1, col) // Bottom
			}
		}
	}
	// Left and right borders
	for y := 0; y < h; y += dotSize + gapSize {
		for dy := 0; dy < dotSize && y+dy < h; dy++ {
			for dx := 0; dx < dotSize && dx < w; dx++ {
				img.Set(dx, y+dy, col)     // Left
				img.Set(w-dx-1, y+dy, col) // Right
			}
		}
	}
}

// drawEmbossedBorder draws a 3D embossed effect with light and dark edges.
func (g *Generator) drawEmbossedBorder(img *image.RGBA, col color.Color, thickness int) {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	lightCol := g.lightenColor(col, 0.3)
	darkCol := g.darkenColor(col, 0.3)

	// Light edges (top and left)
	for t := 0; t < thickness; t++ {
		for x := 0; x < w; x++ {
			img.Set(x, t, lightCol)
		}
		for y := 0; y < h; y++ {
			img.Set(t, y, lightCol)
		}
	}

	// Dark edges (bottom and right)
	for t := 0; t < thickness; t++ {
		for x := 0; x < w; x++ {
			img.Set(x, h-t-1, darkCol)
		}
		for y := 0; y < h; y++ {
			img.Set(w-t-1, y, darkCol)
		}
	}
}

// drawEngravedBorder draws a 3D engraved effect (inverse of embossed).
func (g *Generator) drawEngravedBorder(img *image.RGBA, col color.Color, thickness int) {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	lightCol := g.lightenColor(col, 0.3)
	darkCol := g.darkenColor(col, 0.3)

	// Dark edges (top and left)
	for t := 0; t < thickness; t++ {
		for x := 0; x < w; x++ {
			img.Set(x, t, darkCol)
		}
		for y := 0; y < h; y++ {
			img.Set(t, y, darkCol)
		}
	}

	// Light edges (bottom and right)
	for t := 0; t < thickness; t++ {
		for x := 0; x < w; x++ {
			img.Set(x, h-t-1, lightCol)
		}
		for y := 0; y < h; y++ {
			img.Set(w-t-1, y, lightCol)
		}
	}
}

func (g *Generator) drawLine(img *image.RGBA, x1, y1, x2, y2 int, col color.Color) {
	// Simple horizontal line for now
	if y1 == y2 {
		for x := x1; x <= x2 && x < img.Bounds().Max.X; x++ {
			if x >= 0 && y1 >= 0 && y1 < img.Bounds().Max.Y {
				img.Set(x, y1, col)
			}
		}
	}
}

func (g *Generator) drawCircle(img *image.RGBA, cx, cy, radius int, col color.Color, fill bool) {
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			dx := x - cx
			dy := y - cy
			distSq := dx*dx + dy*dy

			if fill {
				if distSq <= radius*radius {
					bounds := img.Bounds()
					if x >= bounds.Min.X && x < bounds.Max.X && y >= bounds.Min.Y && y < bounds.Max.Y {
						img.Set(x, y, col)
					}
				}
			} else {
				// Draw outline only
				if distSq >= (radius-1)*(radius-1) && distSq <= radius*radius {
					bounds := img.Bounds()
					if x >= bounds.Min.X && x < bounds.Max.X && y >= bounds.Min.Y && y < bounds.Max.Y {
						img.Set(x, y, col)
					}
				}
			}
		}
	}
}

func (g *Generator) lightenColor(col color.Color, amount float64) color.Color {
	r, gr, b, a := col.RGBA()
	factor := 1.0 + amount

	newR := uint8(min(255, float64(r>>8)*factor))
	newG := uint8(min(255, float64(gr>>8)*factor))
	newB := uint8(min(255, float64(b>>8)*factor))

	return color.RGBA{R: newR, G: newG, B: newB, A: uint8(a >> 8)}
}

func (g *Generator) darkenColor(col color.Color, amount float64) color.Color {
	r, gr, b, a := col.RGBA()
	factor := 1.0 - amount

	return color.RGBA{
		R: uint8(float64(r>>8) * factor),
		G: uint8(float64(gr>>8) * factor),
		B: uint8(float64(b>>8) * factor),
		A: uint8(a >> 8),
	}
}

func (g *Generator) selectBorderStyle(genreID string) BorderStyle {
	switch genreID {
	case "fantasy":
		return BorderOrnate
	case "scifi", "cyberpunk":
		return BorderGlow
	default:
		return BorderSolid
	}
}

// selectBorderThickness returns consistent border thickness based on genre and element type.
// This ensures UI consistency across different seeds while maintaining genre-specific styling.
// Phase 45: Border thicknesses scaled 2× for 64×64 UI elements.
func (g *Generator) selectBorderThickness(genreID string, elemType ElementType) int {
	if elemType == ElementFrame {
		return 6 // Phase 45: Frames use thicker borders (3 → 6)
	}
	if genreID == "fantasy" || genreID == "horror" {
		return 6 // Phase 45: Ornate genres use thicker borders (3 → 6)
	}
	return 4 // Phase 45: Default thickness for most elements (2 → 4)
}

// selectButtonBaseColor chooses a base color from the palette that has good contrast potential.
// Colors with mid-range luminance (0.25-0.75) work better for state variations (lighten/darken).
// Also validates WCAG 2.1 AA contrast ratio (minimum 4.5:1) with text color.
func (g *Generator) selectButtonBaseColor(pal *palette.Palette, rng *rand.Rand) color.Color {
	// Try up to 10 attempts to find a good color (increased from 5 for WCAG compliance)
	for attempt := 0; attempt < 10; attempt++ {
		colorIndex := rng.Intn(len(pal.Colors))
		candidate := pal.Colors[colorIndex]

		// Calculate relative luminance
		luminance := g.calculateRelativeLuminance(candidate)

		// Prefer colors with mid-range luminance (not too dark or too light)
		// This ensures effective lightening and darkening for button states
		if luminance > 0.25 && luminance < 0.75 {
			// Validate WCAG 2.1 AA contrast ratio with text color
			contrastRatio := g.calculateContrastRatio(candidate, pal.Text)
			if contrastRatio >= 4.5 {
				return candidate
			}
		}
	}

	// Fallback: validate Primary color, or adjust if needed
	if g.calculateContrastRatio(pal.Primary, pal.Text) >= 4.5 {
		return pal.Primary
	}

	// Last resort: create a color that definitely meets WCAG AA
	// Use a dark background if text is light, or light background if text is dark
	textLuminance := g.calculateRelativeLuminance(pal.Text)
	if textLuminance > 0.5 {
		// Light text: use dark background (0.2 luminance)
		return color.RGBA{R: 51, G: 51, B: 51, A: 255} // ~0.2 luminance, contrast ~10:1 with white
	}
	// Dark text: use light background (0.8 luminance)
	return color.RGBA{R: 204, G: 204, B: 204, A: 255} // ~0.8 luminance, contrast ~10:1 with black
}

// calculateRelativeLuminance calculates the relative luminance of a color per WCAG 2.1 spec.
// Formula: L = 0.2126 * R + 0.7152 * G + 0.0722 * B (using linearized RGB values)
// Returns value between 0.0 (darkest) and 1.0 (lightest)
func (g *Generator) calculateRelativeLuminance(c color.Color) float64 {
	r, gr, b, _ := c.RGBA()
	// Convert from 16-bit (0-65535) to 8-bit (0-255) then to 0.0-1.0
	rNorm := float64(r>>8) / 255.0
	gNorm := float64(gr>>8) / 255.0
	bNorm := float64(b>>8) / 255.0

	// Linearize sRGB values
	rLin := g.linearizeSRGB(rNorm)
	gLin := g.linearizeSRGB(gNorm)
	bLin := g.linearizeSRGB(bNorm)

	// Calculate relative luminance using WCAG formula
	return 0.2126*rLin + 0.7152*gLin + 0.0722*bLin
}

// linearizeSRGB converts sRGB color component to linear RGB per WCAG spec.
func (g *Generator) linearizeSRGB(component float64) float64 {
	if component <= 0.03928 {
		return component / 12.92
	}
	return math.Pow((component+0.055)/1.055, 2.4)
}

// calculateContrastRatio calculates WCAG 2.1 contrast ratio between two colors.
// Formula: (L1 + 0.05) / (L2 + 0.05) where L1 is the lighter color's luminance
// Returns ratio between 1:1 (no contrast) and 21:1 (maximum contrast)
// WCAG 2.1 AA requires minimum 4.5:1 for normal text, 3:1 for large text
func (g *Generator) calculateContrastRatio(c1, c2 color.Color) float64 {
	l1 := g.calculateRelativeLuminance(c1)
	l2 := g.calculateRelativeLuminance(c2)

	// Ensure L1 is the lighter color
	if l1 < l2 {
		l1, l2 = l2, l1
	}

	return (l1 + 0.05) / (l2 + 0.05)
}

// isTechGenre determines if a genre ID represents a technological/futuristic theme.
// Tech genres use square icons, while organic/natural genres use circular icons.
// This includes both pure genres and blended genres containing tech elements.
func (g *Generator) isTechGenre(genreID string) bool {
	// Direct tech genre IDs
	if genreID == "scifi" || genreID == "cyberpunk" {
		return true
	}

	// Check for tech-related keywords in blended genre IDs
	// This handles cases like "sci-fi-horror", "cyber-fantasy", etc.
	techKeywords := []string{"scifi", "sci-fi", "cyber", "tech", "digital", "future"}
	for _, keyword := range techKeywords {
		if strings.Contains(genreID, keyword) {
			return true
		}
	}

	return false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Validate implements the procgen.Generator interface.
func (g *Generator) Validate(result interface{}) error {
	img, ok := result.(*image.RGBA)
	if !ok {
		return fmt.Errorf("result is not an *image.RGBA")
	}

	if img == nil {
		return fmt.Errorf("generated image is nil")
	}

	bounds := img.Bounds()
	if bounds.Dx() == 0 || bounds.Dy() == 0 {
		return fmt.Errorf("generated image has zero dimensions")
	}

	return nil
}
