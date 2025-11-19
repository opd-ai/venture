// Package ui provides procedural UI element generation.
// This file implements visual hierarchy and grouping features.
package ui

import (
	"image"
	"image/color"

	"github.com/opd-ai/venture/pkg/rendering/palette"
)

// HierarchyStyle contains visual properties for different hierarchy levels.
type HierarchyStyle struct {
	// Scale multiplier for the element (1.0 = normal)
	Scale float64
	// Opacity for the element (0.0-1.0)
	Opacity float64
	// Border thickness
	BorderThickness int
	// Background opacity
	BackgroundOpacity float64
	// Font size multiplier (for text elements)
	FontScale float64
}

// GetHierarchyStyle returns the visual style for a given hierarchy level.
func GetHierarchyStyle(level HierarchyLevel) HierarchyStyle {
	switch level {
	case HierarchyPrimary:
		return HierarchyStyle{
			Scale:             1.2,
			Opacity:           1.0,
			BorderThickness:   3,
			BackgroundOpacity: 1.0,
			FontScale:         1.5,
		}
	case HierarchySecondary:
		return HierarchyStyle{
			Scale:             1.0,
			Opacity:           1.0,
			BorderThickness:   2,
			BackgroundOpacity: 0.95,
			FontScale:         1.0,
		}
	case HierarchyTertiary:
		return HierarchyStyle{
			Scale:             0.9,
			Opacity:           0.85,
			BorderThickness:   1,
			BackgroundOpacity: 0.8,
			FontScale:         0.9,
		}
	case HierarchyQuaternary:
		return HierarchyStyle{
			Scale:             0.8,
			Opacity:           0.7,
			BorderThickness:   1,
			BackgroundOpacity: 0.6,
			FontScale:         0.8,
		}
	default:
		return GetHierarchyStyle(HierarchySecondary)
	}
}

// ApplyHierarchyOpacity applies opacity to a color based on hierarchy level.
func ApplyHierarchyOpacity(col color.Color, level HierarchyLevel) color.Color {
	style := GetHierarchyStyle(level)
	r, g, b, a := col.RGBA()

	return color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(float64(a>>8) * style.Opacity),
	}
}

// SeparatorStyle represents different visual separator styles.
type SeparatorStyle int

const (
	// SeparatorLine is a simple horizontal line
	SeparatorLine SeparatorStyle = iota
	// SeparatorDashed is a dashed line
	SeparatorDashed
	// SeparatorDotted is a dotted line
	SeparatorDotted
	// SeparatorGradient is a gradient fade separator
	SeparatorGradient
	// SeparatorOrnamental is a decorative separator with pattern
	SeparatorOrnamental
)

// String returns the string representation of a separator style.
func (s SeparatorStyle) String() string {
	switch s {
	case SeparatorLine:
		return "line"
	case SeparatorDashed:
		return "dashed"
	case SeparatorDotted:
		return "dotted"
	case SeparatorGradient:
		return "gradient"
	case SeparatorOrnamental:
		return "ornamental"
	default:
		return "unknown"
	}
}

// GenerateSeparator creates a visual separator image.
func (g *Generator) GenerateSeparator(width, height int, style SeparatorStyle, genreID string, seed int64) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Generate palette
	pal, err := g.paletteGen.Generate(genreID, seed)
	if err != nil {
		// Fallback to a simple gray separator
		pal = &palette.Palette{
			Primary: color.RGBA{128, 128, 128, 255},
		}
	}

	col := pal.Primary

	switch style {
	case SeparatorLine:
		g.drawHorizontalLine(img, 0, height/2, width, col)
	case SeparatorDashed:
		g.drawDashedSeparator(img, width, height, col)
	case SeparatorDotted:
		g.drawDottedSeparator(img, width, height, col)
	case SeparatorGradient:
		g.drawGradientSeparator(img, width, height, col)
	case SeparatorOrnamental:
		g.drawOrnamentalSeparator(img, width, height, col)
	}

	return img
}

// GroupConfig contains parameters for visual grouping containers.
type GroupConfig struct {
	// Width and height in pixels
	Width  int
	Height int
	// Padding inside the group
	Padding int
	// HierarchyLevel of the group
	Level HierarchyLevel
	// GenreID for styling
	GenreID string
	// Seed for deterministic generation
	Seed int64
	// Whether to show a border
	ShowBorder bool
	// Whether to show a background
	ShowBackground bool
}

// GenerateGroupContainer creates a visual grouping container.
func (g *Generator) GenerateGroupContainer(config GroupConfig) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, config.Width, config.Height))

	// Generate palette
	pal, err := g.paletteGen.Generate(config.GenreID, config.Seed)
	if err != nil {
		return img // Return empty container on error
	}

	style := GetHierarchyStyle(config.Level)

	// Draw background if requested
	if config.ShowBackground {
		bgColor := pal.Background
		r, gr, b, _ := bgColor.RGBA()
		bgWithOpacity := color.RGBA{
			R: uint8(r >> 8),
			G: uint8(gr >> 8),
			B: uint8(b >> 8),
			A: uint8(255 * style.BackgroundOpacity),
		}
		g.fillRect(img, 0, 0, config.Width, config.Height, bgWithOpacity)
	}

	// Draw border if requested
	if config.ShowBorder {
		borderColor := ApplyHierarchyOpacity(pal.Primary, config.Level)
		borderStyle := g.selectBorderStyle(config.GenreID)
		g.drawBorder(img, borderColor, borderStyle, style.BorderThickness)
	}

	return img
}

// Helper functions

func (g *Generator) drawHorizontalLine(img *image.RGBA, x, y, width int, col color.Color) {
	for i := 0; i < width && x+i < img.Bounds().Max.X; i++ {
		if y >= 0 && y < img.Bounds().Max.Y && x+i >= 0 {
			img.Set(x+i, y, col)
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// drawDashedSeparator renders a dashed line separator on the image.
func (g *Generator) drawDashedSeparator(img *image.RGBA, width, height int, col color.Color) {
	dashLength := 8
	gapLength := 4
	y := height / 2
	for x := 0; x < width; x += dashLength + gapLength {
		endX := x + dashLength
		if endX > width {
			endX = width
		}
		g.drawHorizontalLine(img, x, y, endX-x, col)
	}
}

// drawDottedSeparator renders a dotted line separator on the image.
func (g *Generator) drawDottedSeparator(img *image.RGBA, width, height int, col color.Color) {
	dotSize := 2
	gapSize := 4
	y := height / 2
	for x := 0; x < width; x += dotSize + gapSize {
		for dy := 0; dy < dotSize && y+dy < height; dy++ {
			for dx := 0; dx < dotSize && x+dx < width; dx++ {
				img.Set(x+dx, y+dy, col)
			}
		}
	}
}

// drawGradientSeparator renders a gradient fade separator on the image.
func (g *Generator) drawGradientSeparator(img *image.RGBA, width, height int, col color.Color) {
	y := height / 2
	for x := 0; x < width; x++ {
		alpha := calculateGradientAlpha(x, width)
		r, gr, b, _ := col.RGBA()
		gradCol := color.RGBA{
			R: uint8(r >> 8),
			G: uint8(gr >> 8),
			B: uint8(b >> 8),
			A: uint8(255 * alpha),
		}
		img.Set(x, y, gradCol)
		if height > 2 {
			img.Set(x, y+1, gradCol)
		}
	}
}

// calculateGradientAlpha calculates alpha value based on distance from edges.
func calculateGradientAlpha(x, width int) float64 {
	alpha := 1.0
	fadeWidth := width / 4
	if x < fadeWidth {
		alpha = float64(x) / float64(fadeWidth)
	} else if x > width-fadeWidth {
		alpha = float64(width-x) / float64(fadeWidth)
	}
	return alpha
}

// drawOrnamentalSeparator renders a decorative separator with diamond pattern.
func (g *Generator) drawOrnamentalSeparator(img *image.RGBA, width, height int, col color.Color) {
	// Draw base line
	g.drawHorizontalLine(img, 0, height/2, width, col)

	// Add decorative dots at regular intervals
	dotSpacing := width / 5
	y := height / 2
	for i := 1; i < 5; i++ {
		x := i * dotSpacing
		if x < width {
			g.drawDecorativeDiamond(img, x, y, width, height, col)
		}
	}
}

// drawDecorativeDiamond draws a small decorative diamond at the specified position.
func (g *Generator) drawDecorativeDiamond(img *image.RGBA, x, y, width, height int, col color.Color) {
	size := 3
	for dy := -size; dy <= size; dy++ {
		for dx := -size; dx <= size; dx++ {
			if abs(dx)+abs(dy) <= size {
				px := x + dx
				py := y + dy
				if px >= 0 && px < width && py >= 0 && py < height {
					img.Set(px, py, col)
				}
			}
		}
	}
}
