package ui

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/palette"
)

// Generator creates procedural UI elements.
type Generator struct {
	paletteGen *palette.Generator
}

// NewGenerator creates a new UI element generator.
func NewGenerator() *Generator {
	return &Generator{
		paletteGen: palette.NewGenerator(),
	}
}

// Generate creates a UI element image from the given configuration.
func (g *Generator) Generate(config Config) (*image.RGBA, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create RNG from seed
	rng := rand.New(rand.NewSource(config.Seed))

	// Generate color palette for genre
	pal, err := g.paletteGen.Generate(config.GenreID, config.Seed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate palette: %w", err)
	}

	// Create base image
	img := image.NewRGBA(image.Rect(0, 0, config.Width, config.Height))

	// Generate element based on type
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
		return nil, fmt.Errorf("unknown element type: %d", config.Type)
	}

	return img, nil
}

// generateButton creates a button UI element.
func (g *Generator) generateButton(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config Config) {
	// Select colors based on state
	var bgColor, borderColor color.Color
	
	// Use random color from palette to add seed variation
	colorIndex := rng.Intn(len(pal.Colors))
	baseColor := pal.Colors[colorIndex]
	
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

	// Fill background
	g.fillRect(img, 0, 0, config.Width, config.Height, bgColor)

	// Draw border based on genre
	borderStyle := g.selectBorderStyle(config.GenreID)
	borderThickness := 2 + rng.Intn(2) // 2 or 3 pixels
	g.drawBorder(img, borderColor, borderStyle, borderThickness)

	// Add highlight if not disabled
	if config.State != StateDisabled {
		highlightColor := g.lightenColor(bgColor, 0.4)
		g.drawLine(img, 2, 2, config.Width-3, 2, highlightColor)
	}
}

// generatePanel creates a panel UI element.
func (g *Generator) generatePanel(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config Config) {
	// Semi-transparent background
	bgColor := pal.Background
	r, gr, b, _ := bgColor.RGBA()
	semiTransparent := color.RGBA{
		R: uint8(r >> 8),
		G: uint8(gr >> 8),
		B: uint8(b >> 8),
		A: 200,
	}

	// Fill background
	g.fillRect(img, 0, 0, config.Width, config.Height, semiTransparent)

	// Draw border
	borderColor := g.lightenColor(pal.Background, 0.3)
	g.drawBorder(img, borderColor, BorderSolid, 1)
}

// generateHealthBar creates a health/progress bar.
func (g *Generator) generateHealthBar(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config Config) {
	// Background
	bgColor := g.darkenColor(pal.Background, 0.2)
	g.fillRect(img, 0, 0, config.Width, config.Height, bgColor)

	// Calculate filled width based on value
	filledWidth := int(float64(config.Width-4) * config.Value)

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
	if filledWidth > 0 {
		g.fillRect(img, 2, 2, filledWidth, config.Height-4, fillColor)

		// Add shine effect
		shineColor := g.lightenColor(fillColor, 0.3)
		g.drawLine(img, 2, 3, filledWidth, 3, shineColor)
	}

	// Border
	borderColor := g.lightenColor(pal.Background, 0.4)
	g.drawBorder(img, borderColor, BorderSolid, 1)
}

// generateLabel creates a text label background.
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
	g.fillRect(img, 1, 1, config.Width-2, config.Height-2, labelBg)

	// Optional border for emphasis
	if config.State == StateHover {
		borderColor := pal.Primary
		g.drawBorder(img, borderColor, BorderSolid, 1)
	}
}

// generateIcon creates a small iconic UI element.
func (g *Generator) generateIcon(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config Config) {
	// Background circle or square based on genre
	bgColor := pal.Primary

	if config.GenreID == "scifi" || config.GenreID == "cyberpunk" {
		// Square icon for tech genres
		g.fillRect(img, 2, 2, config.Width-4, config.Height-4, bgColor)
	} else {
		// Circular icon for others
		centerX := config.Width / 2
		centerY := config.Height / 2
		radius := config.Width / 2 - 2
		g.drawCircle(img, centerX, centerY, radius, bgColor, true)
	}

	// Border
	borderColor := g.darkenColor(bgColor, 0.3)
	g.drawBorder(img, borderColor, BorderSolid, 1)
}

// generateFrame creates a decorative frame.
func (g *Generator) generateFrame(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config Config) {
	// Frame is mostly transparent with ornate border
	borderColor := pal.Primary
	borderStyle := g.selectBorderStyle(config.GenreID)

	// Draw double border for emphasis
	g.drawBorder(img, borderColor, borderStyle, 3)

	// Inner border
	innerColor := g.lightenColor(borderColor, 0.2)
	g.drawBorder(img, innerColor, BorderSolid, 1)

	// Corner decorations for ornate genres
	if config.GenreID == "fantasy" || config.GenreID == "horror" {
		cornerSize := 4
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
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	switch style {
	case BorderSolid, BorderDouble, BorderOrnate, BorderGlow:
		// All styles use solid for now
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

func min(a, b float64) float64 {
	if a < b {
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
