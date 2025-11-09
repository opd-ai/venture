// Package ui provides procedural UI element generation.
// This file implements Phase 19.3: Procedural UI Decorations including
// genre-themed frames, procedural icon symbols, and enhanced visual effects.
package ui

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/opd-ai/venture/pkg/rendering/patterns"
)

// GenerateDecorativeFrame generates a genre-themed decorative frame.
// This implements Phase 19.3: genre-themed UI frames and borders.
func (g *Generator) GenerateDecorativeFrame(config Config, style FrameStyle) (*image.RGBA, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	pal, err := g.paletteGen.Generate(config.GenreID, config.Seed)
	if err != nil {
		return nil, err
	}

	img := image.NewRGBA(image.Rect(0, 0, config.Width, config.Height))

	// Apply genre-specific frame decoration based on style
	switch style {
	case FrameOrnateCorners:
		g.generateOrnateCornerFrame(img, pal, config)
	case FrameTechAngular:
		g.generateTechAngularFrame(img, pal, config)
	case FrameWeathered:
		g.generateWeatheredFrame(img, pal, config)
	default:
		// Auto-select or default to ornate
		g.generateOrnateCornerFrame(img, pal, config)
	}

	// Apply state-based visual enhancements (hover/focus effects)
	g.applyStateEffectsToFrame(img, pal, config)

	return img, nil
}

// generateOrnateCornerFrame creates decorative corner elements (fantasy).
func (g *Generator) generateOrnateCornerFrame(img *image.RGBA, pal *palette.Palette, config Config) {
	borderColor := pal.Primary
	accentColor := pal.Accent1
	w, h := config.Width, config.Height

	// Draw double border
	g.drawBorder(img, borderColor, BorderDouble, 1)

	// Add corner decorations (diamond shapes)
	cornerSize := 8
	for i := 0; i < cornerSize && i < w && i < h; i++ {
		// Top-left
		if i < w && cornerSize-i-1 < h {
			img.Set(i, cornerSize-i-1, accentColor)
		}
		if cornerSize-i-1 < w && i < h {
			img.Set(cornerSize-i-1, i, accentColor)
		}
		// Top-right
		if w-i-1 >= 0 && cornerSize-i-1 < h {
			img.Set(w-i-1, cornerSize-i-1, accentColor)
		}
		if w-cornerSize+i < w && i < h {
			img.Set(w-cornerSize+i, i, accentColor)
		}
		// Bottom-left
		if i < w && h-cornerSize+i < h && h-cornerSize+i >= 0 {
			img.Set(i, h-cornerSize+i, accentColor)
		}
		if cornerSize-i-1 < w && h-i-1 >= 0 {
			img.Set(cornerSize-i-1, h-i-1, accentColor)
		}
		// Bottom-right
		if w-i-1 >= 0 && h-cornerSize+i < h && h-cornerSize+i >= 0 {
			img.Set(w-i-1, h-cornerSize+i, accentColor)
		}
		if w-cornerSize+i < w && h-i-1 >= 0 {
			img.Set(w-cornerSize+i, h-i-1, accentColor)
		}
	}

	// Add center decorations on each side
	midX, midY := w/2, h/2
	decorSize := 4
	for i := -decorSize; i <= decorSize; i++ {
		if midX+i >= 0 && midX+i < w {
			// Top
			img.Set(midX+i, 0, accentColor)
			if h > 1 {
				img.Set(midX+i, 1, accentColor)
			}
			// Bottom
			img.Set(midX+i, h-1, accentColor)
			if h > 1 {
				img.Set(midX+i, h-2, accentColor)
			}
		}
		if midY+i >= 0 && midY+i < h {
			// Left
			img.Set(0, midY+i, accentColor)
			if w > 1 {
				img.Set(1, midY+i, accentColor)
			}
			// Right
			img.Set(w-1, midY+i, accentColor)
			if w > 1 {
				img.Set(w-2, midY+i, accentColor)
			}
		}
	}
}

// generateTechAngularFrame creates angular geometric frame (scifi/cyberpunk).
func (g *Generator) generateTechAngularFrame(img *image.RGBA, pal *palette.Palette, config Config) {
	borderColor := pal.Primary
	accentColor := pal.Accent1
	w, h := config.Width, config.Height

	// Draw main border
	g.drawBorder(img, borderColor, BorderSolid, 1)

	// Add angular corner cuts
	cutSize := 6
	for i := 0; i < cutSize && i < w && i < h; i++ {
		// Top-left cut
		if i < w && cutSize-i-1 < h {
			img.Set(i, cutSize-i-1, accentColor)
		}
		// Top-right cut
		if w-i-1 >= 0 && cutSize-i-1 < h {
			img.Set(w-i-1, cutSize-i-1, accentColor)
		}
		// Bottom-left cut
		if i < w && h-cutSize+i < h && h-cutSize+i >= 0 {
			img.Set(i, h-cutSize+i, accentColor)
		}
		// Bottom-right cut
		if w-i-1 >= 0 && h-cutSize+i < h && h-cutSize+i >= 0 {
			img.Set(w-i-1, h-cutSize+i, accentColor)
		}
	}
}

// generateWeatheredFrame creates damaged, worn frame (postapoc).
func (g *Generator) generateWeatheredFrame(img *image.RGBA, pal *palette.Palette, config Config) {
	borderColor := pal.Primary

	// Draw thick border
	g.drawBorder(img, borderColor, BorderSolid, 3)

	// Add weathering effect by drawing some darker pixels
	darkColor := g.darkenColor(borderColor, 0.4)
	w, h := config.Width, config.Height

	// Add some weathered spots along the border
	for i := 0; i < 20; i++ {
		x := (i * 13) % w // Deterministic "random" positions
		y := (i * 17) % h
		if x < 5 || x >= w-5 || y < 5 || y >= h-5 {
			img.Set(x, y, darkColor)
		}
	}
}

// applyStateEffectsToFrame applies visual effects based on element state.
// This implements Phase 19.3: hover/focus visual effects.
func (g *Generator) applyStateEffectsToFrame(img *image.RGBA, pal *palette.Palette, config Config) {
	w, h := config.Width, config.Height
	borderColor := pal.Primary

	switch config.State {
	case StateHover:
		// Add subtle glow for hover
		glowColor := g.lightenColor(borderColor, 0.3)
		r, gr, b, _ := glowColor.RGBA()
		glowRGBA := color.RGBA{
			R: uint8(r >> 8),
			G: uint8(gr >> 8),
			B: uint8(b >> 8),
			A: 128,
		}
		// Draw glow (simple implementation - outer pixels)
		for x := 0; x < w; x++ {
			if x > 0 && x < w-1 {
				img.Set(x, 0, glowRGBA)
				img.Set(x, h-1, glowRGBA)
			}
		}
		for y := 0; y < h; y++ {
			if y > 0 && y < h-1 {
				img.Set(0, y, glowRGBA)
				img.Set(w-1, y, glowRGBA)
			}
		}

	case StatePressed:
		// Darken border for pressed state
		pressedColor := g.darkenColor(borderColor, 0.2)
		g.drawBorder(img, pressedColor, BorderSolid, 1)

	case StateDisabled:
		// Desaturate for disabled state
		disabledColor := color.RGBA{R: 128, G: 128, B: 128, A: 255}
		g.drawBorder(img, disabledColor, BorderDashed, 1)
	}
}

// GenerateSymbol generates a procedural icon symbol.
// This implements Phase 19.3: procedural icons and symbols.
func (g *Generator) GenerateSymbol(config Config, symbol IconSymbol) (*image.RGBA, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	pal, err := g.paletteGen.Generate(config.GenreID, config.Seed)
	if err != nil {
		return nil, err
	}

	img := image.NewRGBA(image.Rect(0, 0, config.Width, config.Height))
	symbolColor := pal.Primary

	cx := config.Width / 2
	cy := config.Height / 2
	size := minInt(config.Width, config.Height) / 3

	switch symbol {
	case IconSword:
		g.generateSwordSymbol(img, symbolColor, cx, cy, size)
	case IconShield:
		g.generateShieldSymbol(img, symbolColor, cx, cy, size)
	case IconPotion:
		g.generatePotionSymbol(img, symbolColor, cx, cy, size)
	case IconCoin:
		g.generateCoinSymbol(img, symbolColor, cx, cy, size)
	case IconHeart:
		g.generateHeartSymbol(img, symbolColor, cx, cy, size)
	case IconStar:
		g.generateStarSymbol(img, symbolColor, cx, cy, size)
	case IconGear:
		g.generateGearSymbol(img, symbolColor, cx, cy, size)
	case IconCheckmark:
		g.generateCheckmarkSymbol(img, symbolColor, cx, cy, size)
	case IconX:
		g.generateXSymbol(img, symbolColor, cx, cy, size)
	case IconArrowUp:
		g.generateArrowSymbol(img, symbolColor, cx, cy, size, 0)
	case IconArrowDown:
		g.generateArrowSymbol(img, symbolColor, cx, cy, size, 2)
	case IconArrowLeft:
		g.generateArrowSymbol(img, symbolColor, cx, cy, size, 3)
	case IconArrowRight:
		g.generateArrowSymbol(img, symbolColor, cx, cy, size, 1)
	default:
		return nil, fmt.Errorf("unknown symbol: %v", symbol)
	}

	return img, nil
}

// Symbol generation helper methods (simple implementations)

func (g *Generator) generateSwordSymbol(img *image.RGBA, col color.Color, cx, cy, size int) {
	// Blade (vertical line)
	for y := cy - size; y < cy && y >= 0 && y < img.Bounds().Dy(); y++ {
		img.Set(cx, y, col)
	}
	// Crossguard
	for x := cx - size/2; x < cx+size/2 && x >= 0 && x < img.Bounds().Dx(); x++ {
		if cy >= 0 && cy < img.Bounds().Dy() {
			img.Set(x, cy, col)
		}
	}
	// Handle
	for y := cy + 1; y < cy+size/2 && y >= 0 && y < img.Bounds().Dy(); y++ {
		img.Set(cx, y, col)
	}
}

func (g *Generator) generateShieldSymbol(img *image.RGBA, col color.Color, cx, cy, size int) {
	// Shield outline
	for y := cy - size; y <= cy+size && y >= 0 && y < img.Bounds().Dy(); y++ {
		for x := cx - size/2; x <= cx+size/2 && x >= 0 && x < img.Bounds().Dx(); x++ {
			if x == cx-size/2 || x == cx+size/2 || y == cy-size || y == cy+size {
				img.Set(x, y, col)
			}
		}
	}
	// Center circle
	g.drawCircle(img, cx, cy, size/3, col, false)
}

func (g *Generator) generatePotionSymbol(img *image.RGBA, col color.Color, cx, cy, size int) {
	// Bottle neck
	for y := cy - size; y < cy-size/2 && y >= 0 && y < img.Bounds().Dy(); y++ {
		for dx := -1; dx <= 1; dx++ {
			if cx+dx >= 0 && cx+dx < img.Bounds().Dx() {
				img.Set(cx+dx, y, col)
			}
		}
	}
	// Bottle body (simple rectangle)
	for y := cy - size/2; y < cy+size && y >= 0 && y < img.Bounds().Dy(); y++ {
		width := size / 2
		for x := cx - width; x <= cx+width && x >= 0 && x < img.Bounds().Dx(); x++ {
			if x == cx-width || x == cx+width || y == cy+size-1 || y == cy-size/2 {
				img.Set(x, y, col)
			}
		}
	}
}

func (g *Generator) generateCoinSymbol(img *image.RGBA, col color.Color, cx, cy, size int) {
	// Outer circle
	g.drawCircle(img, cx, cy, size, col, false)
	// Inner circle
	if size > 2 {
		g.drawCircle(img, cx, cy, size-2, col, false)
	}
	// Center line (simplified currency symbol)
	for y := cy - size/2; y <= cy+size/2 && y >= 0 && y < img.Bounds().Dy(); y++ {
		if cx >= 0 && cx < img.Bounds().Dx() {
			img.Set(cx, y, col)
		}
	}
}

func (g *Generator) generateGearSymbol(img *image.RGBA, col color.Color, cx, cy, size int) {
	// Center circle
	g.drawCircle(img, cx, cy, size/2, col, false)

	// Gear teeth (8 teeth)
	teeth := 8
	for i := 0; i < teeth; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(teeth)
		// Outer point of tooth
		x := cx + int(float64(size)*math.Cos(angle))
		y := cy + int(float64(size)*math.Sin(angle))
		// Draw tooth as small square
		bounds := img.Bounds()
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				px, py := x+dx, y+dy
				if px >= 0 && px < bounds.Dx() && py >= 0 && py < bounds.Dy() {
					img.Set(px, py, col)
				}
			}
		}
	}
}

func (g *Generator) generateHeartSymbol(img *image.RGBA, col color.Color, cx, cy, size int) {
	// Simple heart using circles and triangle
	g.drawCircle(img, cx-size/2, cy-size/3, size/2, col, true)
	g.drawCircle(img, cx+size/2, cy-size/3, size/2, col, true)
	// Triangle bottom
	for y := cy; y <= cy+size && y >= 0 && y < img.Bounds().Dy(); y++ {
		width := size - (y - cy)
		for x := cx - width; x <= cx+width && x >= 0 && x < img.Bounds().Dx(); x++ {
			img.Set(x, y, col)
		}
	}
}

func (g *Generator) generateStarSymbol(img *image.RGBA, col color.Color, cx, cy, size int) {
	// 5-pointed star - simpler version that definitely draws pixels
	points := 5

	// Draw from center to each outer point
	for i := 0; i < points; i++ {
		angle := float64(i)*2.0*math.Pi/float64(points) - math.Pi/2
		x := cx + int(float64(size)*math.Cos(angle))
		y := cy + int(float64(size)*math.Sin(angle))
		g.drawLine(img, cx, cy, x, y, col)

		// Also draw to inner points (between outer points at half radius)
		innerAngle := angle + math.Pi/float64(points)
		ix := cx + int(float64(size/2)*math.Cos(innerAngle))
		iy := cy + int(float64(size/2)*math.Sin(innerAngle))
		g.drawLine(img, cx, cy, ix, iy, col)
	}

	// Draw center point to ensure something is visible
	if cx >= 0 && cx < img.Bounds().Dx() && cy >= 0 && cy < img.Bounds().Dy() {
		img.Set(cx, cy, col)
	}
}

func (g *Generator) generateCheckmarkSymbol(img *image.RGBA, col color.Color, cx, cy, size int) {
	// Checkmark as two lines
	// Short part
	for i := 0; i <= size/3; i++ {
		x := cx - size/3
		y := cy + i
		if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
			img.Set(x, y, col)
		}
	}
	// Long part
	for i := 0; i <= size; i++ {
		x := cx - size/3 + i
		y := cy + size/3 - i
		if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
			img.Set(x, y, col)
		}
	}
}

func (g *Generator) generateXSymbol(img *image.RGBA, col color.Color, cx, cy, size int) {
	// X as two diagonals
	for i := -size; i <= size; i++ {
		x1, y1 := cx+i, cy+i
		x2, y2 := cx+i, cy-i
		if x1 >= 0 && x1 < img.Bounds().Dx() && y1 >= 0 && y1 < img.Bounds().Dy() {
			img.Set(x1, y1, col)
		}
		if x2 >= 0 && x2 < img.Bounds().Dx() && y2 >= 0 && y2 < img.Bounds().Dy() {
			img.Set(x2, y2, col)
		}
	}
}

func (g *Generator) generateArrowSymbol(img *image.RGBA, col color.Color, cx, cy, size, direction int) {
	// direction: 0=up, 1=right, 2=down, 3=left
	switch direction {
	case 0: // Up
		for y := cy - size; y <= cy; y++ {
			if y >= 0 && y < img.Bounds().Dy() {
				img.Set(cx, y, col)
			}
		}
		for i := 1; i <= size/2; i++ {
			y := cy - size + i
			if y >= 0 && y < img.Bounds().Dy() {
				if cx-i >= 0 {
					img.Set(cx-i, y, col)
				}
				if cx+i < img.Bounds().Dx() {
					img.Set(cx+i, y, col)
				}
			}
		}
	case 1: // Right
		for x := cx; x <= cx+size && x >= 0 && x < img.Bounds().Dx(); x++ {
			img.Set(x, cy, col)
		}
		for i := 1; i <= size/2; i++ {
			x := cx + size - i
			if x >= 0 && x < img.Bounds().Dx() {
				if cy-i >= 0 {
					img.Set(x, cy-i, col)
				}
				if cy+i < img.Bounds().Dy() {
					img.Set(x, cy+i, col)
				}
			}
		}
	case 2: // Down
		for y := cy; y <= cy+size && y >= 0 && y < img.Bounds().Dy(); y++ {
			img.Set(cx, y, col)
		}
		for i := 1; i <= size/2; i++ {
			y := cy + size - i
			if y >= 0 && y < img.Bounds().Dy() {
				if cx-i >= 0 {
					img.Set(cx-i, y, col)
				}
				if cx+i < img.Bounds().Dx() {
					img.Set(cx+i, y, col)
				}
			}
		}
	case 3: // Left
		for x := cx - size; x <= cx && x >= 0 && x < img.Bounds().Dx(); x++ {
			img.Set(x, cy, col)
		}
		for i := 1; i <= size/2; i++ {
			x := cx - size + i
			if x >= 0 && x < img.Bounds().Dx() {
				if cy-i >= 0 {
					img.Set(x, cy-i, col)
				}
				if cy+i < img.Bounds().Dy() {
					img.Set(x, cy+i, col)
				}
			}
		}
	}
}

// GeneratePanelWithPattern generates a panel with a background pattern.
// This implements Phase 19.3: background patterns for panels.
func (g *Generator) GeneratePanelWithPattern(config Config) (*image.RGBA, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	pal, err := g.paletteGen.Generate(config.GenreID, config.Seed)
	if err != nil {
		return nil, err
	}

	// Generate base panel
	img := image.NewRGBA(image.Rect(0, 0, config.Width, config.Height))

	// Fill with background color
	bgColor := pal.Background
	for y := 0; y < config.Height; y++ {
		for x := 0; x < config.Width; x++ {
			img.Set(x, y, bgColor)
		}
	}

	// Generate pattern texture
	patGen := patterns.NewGenerator()
	patternImg, err := patGen.Generate(patterns.TextureConfig{
		Texture:     patterns.TextureStone,
		Width:       config.Width,
		Height:      config.Height,
		Seed:        config.Seed,
		GenreID:     config.GenreID,
		Color1:      pal.Primary,
		Color2:      pal.Secondary,
		DetailLevel: 0.5,
		Scale:       1.0,
	})
	if err != nil {
		return nil, err
	}

	// Blend pattern with panel (30% opacity)
	for y := 0; y < config.Height; y++ {
		for x := 0; x < config.Width; x++ {
			bgR, bgG, bgB, _ := img.At(x, y).RGBA()
			patR, patG, patB, _ := patternImg.At(x, y).RGBA()

			blendR := uint8((bgR*7 + patR*3) / 10 >> 8)
			blendG := uint8((bgG*7 + patG*3) / 10 >> 8)
			blendB := uint8((bgB*7 + patB*3) / 10 >> 8)

			img.Set(x, y, color.RGBA{R: blendR, G: blendG, B: blendB, A: 255})
		}
	}

	// Add border
	borderColor := pal.Primary
	borderStyle := g.selectBorderStyle(config.GenreID)
	g.drawBorder(img, borderColor, borderStyle, 2)

	return img, nil
}

// minInt returns the minimum of two integers.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
