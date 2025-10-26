// Package environment provides procedural generation of environmental objects.
package environment

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/sirupsen/logrus"
)

// Generator generates environmental objects.
type Generator struct {
	paletteGen *palette.Generator
	logger     *logrus.Entry
}

// NewGenerator creates a new environmental object generator.
func NewGenerator() *Generator {
	return NewGeneratorWithLogger(nil)
}

// NewGeneratorWithLogger creates a new environmental object generator with a logger.
func NewGeneratorWithLogger(logger *logrus.Logger) *Generator {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"generator": "environment",
		})
	}
	return &Generator{
		paletteGen: palette.NewGenerator(),
		logger:     logEntry,
	}
}

// Generate creates a single environmental object.
func (g *Generator) Generate(config Config) (*EnvironmentalObject, error) {
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"subType": config.SubType,
			"genreID": config.GenreID,
			"seed":    config.Seed,
			"width":   config.Width,
			"height":  config.Height,
		}).Debug("generating environmental object")
	}

	if err := config.Validate(); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("invalid config")
		}
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create RNG with seed
	rng := rand.New(rand.NewSource(config.Seed))

	// Get object properties
	collidable, interactable, harmful, damage := GetProperties(config.SubType)

	// Generate sprite
	sprite, err := g.generateSprite(config, rng)
	if err != nil {
		if g.logger != nil {
			g.logger.WithError(err).WithField("subType", config.SubType).Error("sprite generation failed")
		}
		return nil, fmt.Errorf("failed to generate sprite: %w", err)
	}

	// Generate name
	name := g.generateName(config.SubType, config.GenreID, rng)

	obj := &EnvironmentalObject{
		Type:         config.SubType.GetObjectType(),
		SubType:      config.SubType,
		Sprite:       sprite,
		Width:        config.Width,
		Height:       config.Height,
		Collidable:   collidable,
		Interactable: interactable,
		Harmful:      harmful,
		Damage:       damage,
		GenreID:      config.GenreID,
		Seed:         config.Seed,
		Name:         name,
	}

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"name":    name,
			"subType": config.SubType,
		}).Info("environmental object generated")
	}

	return obj, nil
}

// generateSprite creates a sprite for the object.
func (g *Generator) generateSprite(config Config, rng *rand.Rand) (*image.RGBA, error) {
	// Generate color palette for genre
	pal, err := g.paletteGen.Generate(config.GenreID, config.Seed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate palette: %w", err)
	}

	// Create base image
	img := image.NewRGBA(image.Rect(0, 0, config.Width, config.Height))

	// Select colors based on object type
	var baseColor, accentColor color.Color
	objectType := config.SubType.GetObjectType()

	switch objectType {
	case ObjectFurniture:
		baseColor = pal.Secondary
		accentColor = pal.Accent1
	case ObjectDecoration:
		baseColor = pal.Accent1
		accentColor = pal.Primary
	case ObjectObstacle:
		baseColor = pal.Neutral
		accentColor = pal.Secondary
	case ObjectHazard:
		baseColor = pal.Primary
		accentColor = pal.Accent1
	}

	// Draw object based on subtype
	switch config.SubType {
	case SubTypeTable, SubTypeDesk:
		g.drawTable(img, config.Width, config.Height, baseColor, accentColor)
	case SubTypeChair, SubTypeBench:
		g.drawChair(img, config.Width, config.Height, baseColor, accentColor)
	case SubTypeBed:
		g.drawBed(img, config.Width, config.Height, baseColor, accentColor)
	case SubTypeShelf, SubTypeCabinet:
		g.drawShelf(img, config.Width, config.Height, baseColor, accentColor)
	case SubTypeChest:
		g.drawChest(img, config.Width, config.Height, baseColor, accentColor)
	case SubTypePlant:
		g.drawPlant(img, config.Width, config.Height, baseColor, accentColor, rng)
	case SubTypeStatue:
		g.drawStatue(img, config.Width, config.Height, baseColor, accentColor)
	case SubTypePainting, SubTypeTapestry:
		g.drawPainting(img, config.Width, config.Height, baseColor, accentColor, rng)
	case SubTypeBanner:
		g.drawBanner(img, config.Width, config.Height, baseColor, accentColor)
	case SubTypeTorch, SubTypeCandlestick:
		g.drawTorch(img, config.Width, config.Height, baseColor, accentColor)
	case SubTypeVase:
		g.drawVase(img, config.Width, config.Height, baseColor, accentColor)
	case SubTypeCrystal:
		g.drawCrystal(img, config.Width, config.Height, baseColor, accentColor, rng)
	case SubTypeBook:
		g.drawBook(img, config.Width, config.Height, baseColor, accentColor)
	case SubTypeBarrel, SubTypeCrate:
		g.drawBarrel(img, config.Width, config.Height, baseColor, accentColor)
	case SubTypeRubble, SubTypeDebris:
		g.drawRubble(img, config.Width, config.Height, baseColor, accentColor, rng)
	case SubTypePillar, SubTypeColumn:
		g.drawPillar(img, config.Width, config.Height, baseColor, accentColor)
	case SubTypeBoulder:
		g.drawBoulder(img, config.Width, config.Height, baseColor, accentColor, rng)
	case SubTypeWreckage:
		g.drawWreckage(img, config.Width, config.Height, baseColor, accentColor, rng)
	case SubTypeSpikes:
		g.drawSpikes(img, config.Width, config.Height, baseColor, accentColor)
	case SubTypeFirePit, SubTypeLavaPit:
		g.drawFirePit(img, config.Width, config.Height, baseColor, accentColor, rng)
	case SubTypeAcidPool:
		g.drawAcidPool(img, config.Width, config.Height, baseColor, accentColor, rng)
	case SubTypeBearTrap:
		g.drawBearTrap(img, config.Width, config.Height, baseColor, accentColor)
	case SubTypePoisonGas:
		g.drawPoisonGas(img, config.Width, config.Height, baseColor, accentColor, rng)
	case SubTypeElectricField:
		g.drawElectricField(img, config.Width, config.Height, baseColor, accentColor, rng)
	case SubTypeIceField:
		g.drawIceField(img, config.Width, config.Height, baseColor, accentColor, rng)
	default:
		// Default: draw rectangle
		g.drawRectangle(img, config.Width, config.Height, baseColor, accentColor)
	}

	return img, nil
}

// Drawing helper functions (simple implementations)

func (g *Generator) drawTable(img *image.RGBA, width, height int, base, accent color.Color) {
	// Draw table top
	for y := height / 3; y < height/3+height/6; y++ {
		for x := width / 6; x < width*5/6; x++ {
			img.Set(x, y, base)
		}
	}
	// Draw legs
	for y := height / 3; y < height*5/6; y++ {
		img.Set(width/6, y, accent)
		img.Set(width*5/6-1, y, accent)
	}
}

func (g *Generator) drawChair(img *image.RGBA, width, height int, base, accent color.Color) {
	// Draw seat
	for y := height / 2; y < height/2+height/6; y++ {
		for x := width / 4; x < width*3/4; x++ {
			img.Set(x, y, base)
		}
	}
	// Draw back
	for y := height / 6; y < height/2; y++ {
		for x := width * 2 / 5; x < width*3/5; x++ {
			img.Set(x, y, accent)
		}
	}
	// Draw legs
	for y := height / 2; y < height*5/6; y++ {
		img.Set(width/4, y, accent)
		img.Set(width*3/4-1, y, accent)
	}
}

func (g *Generator) drawBed(img *image.RGBA, width, height int, base, accent color.Color) {
	// Draw mattress
	for y := height / 2; y < height*3/4; y++ {
		for x := width / 8; x < width*7/8; x++ {
			img.Set(x, y, base)
		}
	}
	// Draw pillow
	for y := height / 3; y < height/2; y++ {
		for x := width / 4; x < width/2; x++ {
			img.Set(x, y, accent)
		}
	}
}

func (g *Generator) drawShelf(img *image.RGBA, width, height int, base, accent color.Color) {
	// Draw shelves (3 horizontal lines)
	for i := 1; i <= 3; i++ {
		y := height * i / 4
		for x := width / 6; x < width*5/6; x++ {
			img.Set(x, y, base)
			img.Set(x, y+1, base)
		}
	}
	// Draw side supports
	for y := height / 6; y < height*5/6; y++ {
		img.Set(width/6, y, accent)
		img.Set(width*5/6-1, y, accent)
	}
}

func (g *Generator) drawChest(img *image.RGBA, width, height int, base, accent color.Color) {
	// Draw chest body
	for y := height / 3; y < height*5/6; y++ {
		for x := width / 6; x < width*5/6; x++ {
			img.Set(x, y, base)
		}
	}
	// Draw lid
	for y := height / 6; y < height/3; y++ {
		for x := width / 6; x < width*5/6; x++ {
			img.Set(x, y, accent)
		}
	}
	// Draw lock
	centerX, centerY := width/2, height*2/3
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			if dx*dx+dy*dy <= 4 {
				img.Set(centerX+dx, centerY+dy, accent)
			}
		}
	}
}

func (g *Generator) drawPlant(img *image.RGBA, width, height int, base, accent color.Color, rng *rand.Rand) {
	// Draw pot
	for y := height * 2 / 3; y < height*5/6; y++ {
		for x := width / 3; x < width*2/3; x++ {
			img.Set(x, y, accent)
		}
	}
	// Draw leaves (random positions)
	for i := 0; i < 5; i++ {
		x := width/2 + rng.Intn(width/4) - width/8
		y := height/3 + rng.Intn(height/4)
		for dy := -2; dy <= 2; dy++ {
			for dx := -2; dx <= 2; dx++ {
				if x+dx >= 0 && x+dx < width && y+dy >= 0 && y+dy < height {
					img.Set(x+dx, y+dy, base)
				}
			}
		}
	}
}

func (g *Generator) drawStatue(img *image.RGBA, width, height int, base, accent color.Color) {
	// Draw pedestal
	for y := height * 2 / 3; y < height*5/6; y++ {
		for x := width / 3; x < width*2/3; x++ {
			img.Set(x, y, accent)
		}
	}
	// Draw figure (simple column)
	for y := height / 4; y < height*2/3; y++ {
		for x := width * 2 / 5; x < width*3/5; x++ {
			img.Set(x, y, base)
		}
	}
	// Draw head
	centerX, centerY := width/2, height/6
	for dy := -height / 10; dy <= height/10; dy++ {
		for dx := -width / 10; dx <= width/10; dx++ {
			if dx*dx+dy*dy <= (width/10)*(height/10) && centerX+dx >= 0 && centerX+dx < width && centerY+dy >= 0 && centerY+dy < height {
				img.Set(centerX+dx, centerY+dy, base)
			}
		}
	}
}

func (g *Generator) drawPainting(img *image.RGBA, width, height int, base, accent color.Color, rng *rand.Rand) {
	// Draw frame
	for y := height / 6; y < height*5/6; y++ {
		for x := width / 6; x < width*5/6; x++ {
			if y < height/6+2 || y >= height*5/6-2 || x < width/6+2 || x >= width*5/6-2 {
				img.Set(x, y, accent)
			} else if rng.Float64() < 0.3 {
				img.Set(x, y, base)
			}
		}
	}
}

func (g *Generator) drawBanner(img *image.RGBA, width, height int, base, accent color.Color) {
	// Draw pole
	for y := 0; y < height; y++ {
		img.Set(width/4, y, accent)
	}
	// Draw fabric
	for y := height / 6; y < height*2/3; y++ {
		for x := width / 4; x < width*5/6; x++ {
			img.Set(x, y, base)
		}
	}
	// Draw wave pattern at bottom
	for x := width / 4; x < width*5/6; x++ {
		y := height*2/3 + ((x-width/4)%4)/2
		if y < height {
			img.Set(x, y, base)
		}
	}
}

func (g *Generator) drawTorch(img *image.RGBA, width, height int, base, accent color.Color) {
	// Draw handle
	for y := height / 3; y < height*5/6; y++ {
		for x := width * 2 / 5; x < width*3/5; x++ {
			img.Set(x, y, accent)
		}
	}
	// Draw flame
	centerX, centerY := width/2, height/6
	for dy := -height / 8; dy <= height/8; dy++ {
		for dx := -width / 8; dx <= width/8; dx++ {
			if dx*dx+dy*dy <= (width/8)*(height/8) && centerX+dx >= 0 && centerX+dx < width && centerY+dy >= 0 && centerY+dy < height {
				img.Set(centerX+dx, centerY+dy, base)
			}
		}
	}
}

func (g *Generator) drawVase(img *image.RGBA, width, height int, base, accent color.Color) {
	// Draw vase body (tapered)
	for y := height / 4; y < height*5/6; y++ {
		widthAtY := width/3 + (width/6)*(height*5/6-y)/(height*5/6-height/4)
		for x := width/2 - widthAtY; x <= width/2+widthAtY; x++ {
			if x >= 0 && x < width {
				img.Set(x, y, base)
			}
		}
	}
	// Draw rim
	for x := width / 3; x < width*2/3; x++ {
		img.Set(x, height/4, accent)
		img.Set(x, height/4+1, accent)
	}
}

func (g *Generator) drawCrystal(img *image.RGBA, width, height int, base, accent color.Color, rng *rand.Rand) {
	// Draw crystal points (simple triangular shapes)
	centerX, centerY := width/2, height/2
	points := []struct{ x, y int }{
		{centerX, height / 6},
		{centerX + width/4, centerY},
		{centerX, height * 5 / 6},
		{centerX - width/4, centerY},
	}
	for _, p := range points {
		for dy := -height / 8; dy <= height/8; dy++ {
			for dx := -width / 8; dx <= width/8; dx++ {
				if abs(dx)+abs(dy) <= width/8 && p.x+dx >= 0 && p.x+dx < width && p.y+dy >= 0 && p.y+dy < height {
					if rng.Float64() < 0.7 {
						img.Set(p.x+dx, p.y+dy, base)
					} else {
						img.Set(p.x+dx, p.y+dy, accent)
					}
				}
			}
		}
	}
}

func (g *Generator) drawBook(img *image.RGBA, width, height int, base, accent color.Color) {
	// Draw book cover
	for y := height / 4; y < height*3/4; y++ {
		for x := width / 4; x < width*3/4; x++ {
			img.Set(x, y, base)
		}
	}
	// Draw spine
	for y := height / 4; y < height*3/4; y++ {
		img.Set(width/4, y, accent)
		img.Set(width/4+1, y, accent)
	}
	// Draw pages
	for y := height/4 + 2; y < height*3/4-2; y++ {
		img.Set(width*3/4-1, y, accent)
	}
}

func (g *Generator) drawBarrel(img *image.RGBA, width, height int, base, accent color.Color) {
	// Draw barrel body (cylindrical)
	for y := height / 6; y < height*5/6; y++ {
		for x := width / 4; x < width*3/4; x++ {
			img.Set(x, y, base)
		}
	}
	// Draw hoops
	for _, y := range []int{height / 3, height / 2, height * 2 / 3} {
		for x := width / 4; x < width*3/4; x++ {
			img.Set(x, y, accent)
		}
	}
}

func (g *Generator) drawRubble(img *image.RGBA, width, height int, base, accent color.Color, rng *rand.Rand) {
	// Draw random chunks
	for i := 0; i < 10; i++ {
		x := rng.Intn(width)
		y := rng.Intn(height)
		size := 2 + rng.Intn(4)
		for dy := 0; dy < size; dy++ {
			for dx := 0; dx < size; dx++ {
				if x+dx < width && y+dy < height {
					if rng.Float64() < 0.7 {
						img.Set(x+dx, y+dy, base)
					} else {
						img.Set(x+dx, y+dy, accent)
					}
				}
			}
		}
	}
}

func (g *Generator) drawPillar(img *image.RGBA, width, height int, base, accent color.Color) {
	// Draw pillar column
	for y := height / 8; y < height*7/8; y++ {
		for x := width / 3; x < width*2/3; x++ {
			img.Set(x, y, base)
		}
	}
	// Draw capital (top)
	for y := height / 8; y < height/8+height/12; y++ {
		for x := width / 4; x < width*3/4; x++ {
			img.Set(x, y, accent)
		}
	}
	// Draw base
	for y := height*7/8 - height/12; y < height*7/8; y++ {
		for x := width / 4; x < width*3/4; x++ {
			img.Set(x, y, accent)
		}
	}
}

func (g *Generator) drawBoulder(img *image.RGBA, width, height int, base, accent color.Color, rng *rand.Rand) {
	// Draw irregular circular shape
	centerX, centerY := width/2, height/2
	radius := min(width, height) / 3
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			dx := x - centerX
			dy := y - centerY
			dist := dx*dx + dy*dy
			if dist <= radius*radius+rng.Intn(radius) {
				if rng.Float64() < 0.8 {
					img.Set(x, y, base)
				} else {
					img.Set(x, y, accent)
				}
			}
		}
	}
}

func (g *Generator) drawWreckage(img *image.RGBA, width, height int, base, accent color.Color, rng *rand.Rand) {
	// Draw scattered debris
	for i := 0; i < 8; i++ {
		x1 := rng.Intn(width)
		y1 := rng.Intn(height)
		x2 := x1 + rng.Intn(width/4) - width/8
		y2 := y1 + rng.Intn(height/4) - height/8
		g.drawLine(img, x1, y1, x2, y2, base)
	}
}

func (g *Generator) drawSpikes(img *image.RGBA, width, height int, base, accent color.Color) {
	// Draw multiple spike points
	numSpikes := width / 8
	for i := 0; i < numSpikes; i++ {
		x := width*i/numSpikes + width/(2*numSpikes)
		// Draw triangle spike
		for y := height * 2 / 3; y < height*5/6; y++ {
			yOffset := y - height*2/3
			leftX := x - yOffset/2
			rightX := x + yOffset/2
			for sx := leftX; sx <= rightX; sx++ {
				if sx >= 0 && sx < width {
					img.Set(sx, y, base)
				}
			}
		}
		// Draw tip
		img.Set(x, height*2/3-1, accent)
	}
}

func (g *Generator) drawFirePit(img *image.RGBA, width, height int, base, accent color.Color, rng *rand.Rand) {
	// Draw pit
	for y := height / 2; y < height*5/6; y++ {
		for x := width / 4; x < width*3/4; x++ {
			img.Set(x, y, accent)
		}
	}
	// Draw flames (random flicker)
	for i := 0; i < 10; i++ {
		x := width/4 + rng.Intn(width/2)
		y := height/3 + rng.Intn(height/6)
		size := 2 + rng.Intn(3)
		for dy := 0; dy < size; dy++ {
			for dx := -size / 2; dx <= size/2; dx++ {
				if x+dx >= 0 && x+dx < width && y+dy >= 0 && y+dy < height {
					img.Set(x+dx, y+dy, base)
				}
			}
		}
	}
}

func (g *Generator) drawAcidPool(img *image.RGBA, width, height int, base, accent color.Color, rng *rand.Rand) {
	// Draw pool (irregular oval)
	centerX, centerY := width/2, height/2
	radiusX, radiusY := width/3, height/3
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			dx := float64(x - centerX)
			dy := float64(y - centerY)
			dist := (dx*dx)/(float64(radiusX*radiusX)) + (dy*dy)/(float64(radiusY*radiusY))
			if dist <= 1.0+rng.Float64()*0.2 {
				if rng.Float64() < 0.7 {
					img.Set(x, y, base)
				} else {
					img.Set(x, y, accent)
				}
			}
		}
	}
}

func (g *Generator) drawBearTrap(img *image.RGBA, width, height int, base, accent color.Color) {
	// Draw trap jaws (open)
	for y := height / 3; y < height*2/3; y++ {
		// Left jaw
		for x := width / 6; x < width/3; x++ {
			img.Set(x, y, base)
		}
		// Right jaw
		for x := width * 2 / 3; x < width*5/6; x++ {
			img.Set(x, y, base)
		}
	}
	// Draw teeth
	for i := 0; i < 5; i++ {
		y := height/3 + i*height/15
		img.Set(width/3, y, accent)
		img.Set(width*2/3-1, y, accent)
	}
}

func (g *Generator) drawPoisonGas(img *image.RGBA, width, height int, base, accent color.Color, rng *rand.Rand) {
	// Draw wispy gas clouds
	for i := 0; i < 15; i++ {
		x := rng.Intn(width)
		y := rng.Intn(height)
		size := 3 + rng.Intn(4)
		for dy := -size; dy <= size; dy++ {
			for dx := -size; dx <= size; dx++ {
				if x+dx >= 0 && x+dx < width && y+dy >= 0 && y+dy < height {
					if rng.Float64() < 0.4 {
						img.Set(x+dx, y+dy, base)
					}
				}
			}
		}
	}
}

func (g *Generator) drawElectricField(img *image.RGBA, width, height int, base, accent color.Color, rng *rand.Rand) {
	// Draw lightning bolts
	for i := 0; i < 5; i++ {
		x := rng.Intn(width)
		y := 0
		for y < height {
			nextY := y + 5 + rng.Intn(10)
			nextX := x + rng.Intn(11) - 5
			if nextX < 0 {
				nextX = 0
			}
			if nextX >= width {
				nextX = width - 1
			}
			g.drawLine(img, x, y, nextX, nextY, base)
			x = nextX
			y = nextY
		}
	}
}

func (g *Generator) drawIceField(img *image.RGBA, width, height int, base, accent color.Color, rng *rand.Rand) {
	// Draw icy crystals
	for i := 0; i < 10; i++ {
		x := rng.Intn(width)
		y := rng.Intn(height)
		size := 2 + rng.Intn(3)
		// Draw cross pattern
		for d := -size; d <= size; d++ {
			if x+d >= 0 && x+d < width {
				img.Set(x+d, y, base)
			}
			if y+d >= 0 && y+d < height {
				img.Set(x, y+d, base)
			}
			// Diagonals
			if x+d >= 0 && x+d < width && y+d >= 0 && y+d < height {
				img.Set(x+d, y+d, accent)
				if x-d >= 0 && x-d < width {
					img.Set(x-d, y+d, accent)
				}
			}
		}
	}
}

func (g *Generator) drawRectangle(img *image.RGBA, width, height int, base, accent color.Color) {
	// Simple rectangle with border
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if y < 2 || y >= height-2 || x < 2 || x >= width-2 {
				img.Set(x, y, accent)
			} else {
				img.Set(x, y, base)
			}
		}
	}
}

func (g *Generator) drawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.Color) {
	// Bresenham's line algorithm
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := -1
	if x1 < x2 {
		sx = 1
	}
	sy := -1
	if y1 < y2 {
		sy = 1
	}
	err := dx - dy

	for {
		if x1 >= 0 && x1 < img.Bounds().Dx() && y1 >= 0 && y1 < img.Bounds().Dy() {
			img.Set(x1, y1, c)
		}
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

// generateName creates a name for the object based on genre and subtype.
func (g *Generator) generateName(subType SubType, genreID string, rng *rand.Rand) string {
	baseName := subType.String()

	// Add genre-specific prefixes
	var prefixes []string
	switch genreID {
	case "fantasy":
		prefixes = []string{"Ancient", "Mystical", "Enchanted", "Regal", "Sacred"}
	case "scifi":
		prefixes = []string{"Advanced", "Quantum", "Cyber", "Plasma", "Nano"}
	case "horror":
		prefixes = []string{"Cursed", "Haunted", "Decayed", "Bloody", "Corrupted"}
	case "cyberpunk":
		prefixes = []string{"Neon", "Chrome", "Data", "Neural", "Digital"}
	case "postapoc":
		prefixes = []string{"Rusted", "Salvaged", "Makeshift", "Broken", "Scavenged"}
	default:
		return baseName
	}

	if len(prefixes) > 0 && rng.Float64() < 0.7 {
		prefix := prefixes[rng.Intn(len(prefixes))]
		return prefix + " " + baseName
	}

	return baseName
}

// Helper functions

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
