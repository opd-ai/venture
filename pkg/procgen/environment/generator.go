// Package environment provides procedural generation of environmental objects.
package environment

import (
	"fmt"
	"image"
	"image/color"
	"math"
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
	g.logDebug("generating environmental object", logrus.Fields{
		"subType": config.SubType,
		"genreID": config.GenreID,
		"seed":    config.Seed,
		"width":   config.Width,
		"height":  config.Height,
	})

	if err := config.Validate(); err != nil {
		g.logError("invalid config", err)
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	rng := rand.New(rand.NewSource(config.Seed))
	collidable, interactable, harmful, damage := GetProperties(config.SubType)

	sprite, err := g.generateSprite(config, rng)
	if err != nil {
		g.logError("sprite generation failed", err, logrus.Fields{"subType": config.SubType})
		return nil, fmt.Errorf("failed to generate sprite: %w", err)
	}

	name := g.generateName(config.SubType, config.GenreID, rng)
	obj := g.createObject(config, sprite, name, collidable, interactable, harmful, damage)

	g.logInfo("environmental object generated", logrus.Fields{
		"name":    name,
		"subType": config.SubType,
	})

	return obj, nil
}

// createObject assembles an EnvironmentalObject from components.
func (g *Generator) createObject(config Config, sprite *image.RGBA, name string,
	collidable, interactable, harmful bool, damage int,
) *EnvironmentalObject {
	return &EnvironmentalObject{
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
}

// generateSprite creates a sprite for the object.
func (g *Generator) generateSprite(config Config, rng *rand.Rand) (*image.RGBA, error) {
	pal, err := g.paletteGen.Generate(config.GenreID, config.Seed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate palette: %w", err)
	}

	img := image.NewRGBA(image.Rect(0, 0, config.Width, config.Height))
	baseColor, accentColor := g.selectColors(config.SubType.GetObjectType(), pal)
	g.drawObjectSprite(img, config, baseColor, accentColor, rng)

	return img, nil
}

// selectColors chooses base and accent colors based on object type.
func (g *Generator) selectColors(objectType ObjectType, pal *palette.Palette) (base, accent color.Color) {
	switch objectType {
	case ObjectFurniture:
		return pal.Secondary, pal.Accent1
	case ObjectDecoration:
		return pal.Accent1, pal.Primary
	case ObjectObstacle:
		return pal.Neutral, pal.Secondary
	case ObjectHazard:
		return pal.Primary, pal.Accent1
	default:
		return pal.Secondary, pal.Accent1
	}
}

// drawObjectSprite renders the object sprite based on subtype.
func (g *Generator) drawObjectSprite(img *image.RGBA, config Config, baseColor, accentColor color.Color, rng *rand.Rand) {
	w, h := config.Width, config.Height

	switch config.SubType {
	case SubTypeTable, SubTypeDesk:
		g.drawTable(img, w, h, baseColor, accentColor)
	case SubTypeChair, SubTypeBench:
		g.drawChair(img, w, h, baseColor, accentColor)
	case SubTypeBed:
		g.drawBed(img, w, h, baseColor, accentColor)
	case SubTypeShelf, SubTypeCabinet:
		g.drawShelf(img, w, h, baseColor, accentColor)
	case SubTypeChest:
		g.drawChest(img, w, h, baseColor, accentColor)
	case SubTypePlant:
		g.drawPlant(img, w, h, baseColor, accentColor, rng)
	case SubTypeStatue:
		g.drawStatue(img, w, h, baseColor, accentColor)
	case SubTypePainting, SubTypeTapestry:
		g.drawPainting(img, w, h, baseColor, accentColor, rng)
	case SubTypeBanner:
		g.drawBanner(img, w, h, baseColor, accentColor)
	case SubTypeTorch, SubTypeCandlestick:
		g.drawTorch(img, w, h, baseColor, accentColor)
	case SubTypeVase:
		g.drawVase(img, w, h, baseColor, accentColor)
	case SubTypeCrystal:
		g.drawCrystal(img, w, h, baseColor, accentColor, rng)
	case SubTypeBook:
		g.drawBook(img, w, h, baseColor, accentColor)
	// Phase 20.1: New decoration types
	case SubTypeSconce:
		g.drawSconce(img, w, h, baseColor, accentColor)
	case SubTypeWallCrack:
		g.drawWallCrack(img, w, h, baseColor, accentColor, rng)
	case SubTypeBloodstain:
		g.drawBloodstain(img, w, h, baseColor, accentColor, rng)
	case SubTypeGrass:
		g.drawGrass(img, w, h, baseColor, accentColor, rng)
	case SubTypeMushroom:
		g.drawMushroom(img, w, h, baseColor, accentColor, rng)
	case SubTypeSkull:
		g.drawSkull(img, w, h, baseColor, accentColor)
	case SubTypeChain:
		g.drawChain(img, w, h, baseColor, accentColor)
	case SubTypeWeb:
		g.drawWeb(img, w, h, baseColor, accentColor, rng)
	case SubTypeMoss:
		g.drawMoss(img, w, h, baseColor, accentColor, rng)
	case SubTypeGraffiti:
		g.drawGraffiti(img, w, h, baseColor, accentColor, rng)
	case SubTypeBarrel, SubTypeCrate:
		g.drawBarrel(img, w, h, baseColor, accentColor)
	case SubTypeRubble, SubTypeDebris:
		g.drawRubble(img, w, h, baseColor, accentColor, rng)
	case SubTypePillar, SubTypeColumn:
		g.drawPillar(img, w, h, baseColor, accentColor)
	case SubTypeBoulder:
		g.drawBoulder(img, w, h, baseColor, accentColor, rng)
	case SubTypeWreckage:
		g.drawWreckage(img, w, h, baseColor, accentColor, rng)
	case SubTypeSpikes:
		g.drawSpikes(img, w, h, baseColor, accentColor)
	case SubTypeFirePit, SubTypeLavaPit:
		g.drawFirePit(img, w, h, baseColor, accentColor, rng)
	case SubTypeAcidPool:
		g.drawAcidPool(img, w, h, baseColor, accentColor, rng)
	case SubTypeBearTrap:
		g.drawBearTrap(img, w, h, baseColor, accentColor)
	case SubTypePoisonGas:
		g.drawPoisonGas(img, w, h, baseColor, accentColor, rng)
	case SubTypeElectricField:
		g.drawElectricField(img, w, h, baseColor, accentColor, rng)
	case SubTypeIceField:
		g.drawIceField(img, w, h, baseColor, accentColor, rng)
	default:
		g.drawRectangle(img, w, h, baseColor, accentColor)
	}
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

// Phase 20.1: New decoration drawing functions

func (g *Generator) drawSconce(img *image.RGBA, width, height int, base, accent color.Color) {
	// Draw wall mount
	for y := height / 3; y < height*2/3; y++ {
		for x := width / 3; x < width*2/3; x++ {
			img.Set(x, y, accent)
		}
	}
	// Draw flame holder
	for y := height / 6; y < height/3; y++ {
		for x := width * 2 / 5; x < width*3/5; x++ {
			img.Set(x, y, accent)
		}
	}
	// Draw flame
	centerX, centerY := width/2, height/10
	for dy := -height / 12; dy <= height/12; dy++ {
		for dx := -width / 12; dx <= width/12; dx++ {
			if dx*dx+dy*dy <= (width/12)*(height/12) && centerX+dx >= 0 && centerX+dx < width && centerY+dy >= 0 && centerY+dy < height {
				img.Set(centerX+dx, centerY+dy, base)
			}
		}
	}
}

func (g *Generator) drawWallCrack(img *image.RGBA, width, height int, base, accent color.Color, rng *rand.Rand) {
	x := width / 2
	for y := 0; y < height; y++ {
		g.drawCrackSegment(img, x, y, width, base)
		x = g.moveHorizontally(x, width, rng)
		g.addBranch(img, x, y, width, height, accent, rng)
	}
}

// drawCrackSegment draws a crack line segment with thickness.
func (g *Generator) drawCrackSegment(img *image.RGBA, x, y, width int, base color.Color) {
	for dx := -1; dx <= 1; dx++ {
		if x+dx >= 0 && x+dx < width {
			img.Set(x+dx, y, base)
		}
	}
}

// moveHorizontally applies random horizontal movement to crack position.
func (g *Generator) moveHorizontally(x, width int, rng *rand.Rand) int {
	if rng.Float64() < 0.3 {
		x += rng.Intn(5) - 2
		if x < width/4 {
			x = width / 4
		}
		if x >= width*3/4 {
			x = width*3/4 - 1
		}
	}
	return x
}

// addBranch occasionally adds a side branch to the crack.
func (g *Generator) addBranch(img *image.RGBA, x, y, width, height int, accent color.Color, rng *rand.Rand) {
	if rng.Float64() >= 0.1 {
		return
	}

	branchLen := 2 + rng.Intn(4)
	dir := 1
	if rng.Float64() < 0.5 {
		dir = -1
	}

	for i := 0; i < branchLen; i++ {
		bx := x + i*dir
		by := y + i
		if bx >= 0 && bx < width && by < height {
			img.Set(bx, by, accent)
		}
	}
}

func (g *Generator) drawBloodstain(img *image.RGBA, width, height int, base, accent color.Color, rng *rand.Rand) {
	centerX, centerY := width/2, height/2
	g.drawMainBloodStain(img, width, height, centerX, centerY, base, accent, rng)
	g.drawBloodSplatters(img, width, height, centerX, centerY, base, rng)
}

// drawMainBloodStain draws the main irregular blood stain.
func (g *Generator) drawMainBloodStain(img *image.RGBA, width, height, centerX, centerY int, base, accent color.Color, rng *rand.Rand) {
	radius := min(width, height) / 3
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if g.isInsideBloodStain(x, y, centerX, centerY, radius, rng) {
				if rng.Float64() < 0.8 {
					img.Set(x, y, base)
				} else {
					img.Set(x, y, accent)
				}
			}
		}
	}
}

// isInsideBloodStain checks if a point is inside the irregular blood stain.
func (g *Generator) isInsideBloodStain(x, y, centerX, centerY, radius int, rng *rand.Rand) bool {
	dx := x - centerX
	dy := y - centerY
	dist := dx*dx + dy*dy
	threshold := radius * radius
	noise := rng.Intn(radius)
	return dist <= threshold+noise*noise
}

// drawBloodSplatters adds small blood splatters around the main stain.
func (g *Generator) drawBloodSplatters(img *image.RGBA, width, height, centerX, centerY int, base color.Color, rng *rand.Rand) {
	for i := 0; i < 5; i++ {
		sx := centerX + rng.Intn(width/2) - width/4
		sy := centerY + rng.Intn(height/2) - height/4
		size := 2 + rng.Intn(4)
		g.drawSingleSplatter(img, width, height, sx, sy, size, base)
	}
}

// drawSingleSplatter draws a single circular splatter.
func (g *Generator) drawSingleSplatter(img *image.RGBA, width, height, sx, sy, size int, base color.Color) {
	for dy := -size; dy <= size; dy++ {
		for dx := -size; dx <= size; dx++ {
			if sx+dx >= 0 && sx+dx < width && sy+dy >= 0 && sy+dy < height {
				if dx*dx+dy*dy <= size*size {
					img.Set(sx+dx, sy+dy, base)
				}
			}
		}
	}
}

func (g *Generator) drawGrass(img *image.RGBA, width, height int, base, accent color.Color, rng *rand.Rand) {
	// Draw grass blades
	numBlades := 8 + rng.Intn(8)
	for i := 0; i < numBlades; i++ {
		x := rng.Intn(width)
		bladeHeight := height/2 + rng.Intn(height/4)

		// Draw blade (curved line)
		for y := height - 1; y >= height-bladeHeight && y >= 0; y-- {
			offset := int(float64(height-y) * 0.2 * (rng.Float64() - 0.5))
			bx := x + offset
			if bx >= 0 && bx < width {
				if rng.Float64() < 0.7 {
					img.Set(bx, y, base)
				} else {
					img.Set(bx, y, accent)
				}
			}
		}
	}
}

func (g *Generator) drawMushroom(img *image.RGBA, width, height int, base, accent color.Color, rng *rand.Rand) {
	// Draw mushroom cap
	capY := height / 3
	capRadius := width / 3
	for y := capY - capRadius; y <= capY; y++ {
		for x := 0; x < width; x++ {
			dx := x - width/2
			dy := y - capY
			dist := dx*dx + dy*dy
			if dist <= capRadius*capRadius {
				if rng.Float64() < 0.9 {
					img.Set(x, y, base)
				} else {
					img.Set(x, y, accent) // Spots
				}
			}
		}
	}

	// Draw stem
	stemWidth := width / 6
	for y := capY; y < height*5/6; y++ {
		for x := width/2 - stemWidth; x <= width/2+stemWidth; x++ {
			if x >= 0 && x < width {
				img.Set(x, y, accent)
			}
		}
	}
}

func (g *Generator) drawSkull(img *image.RGBA, width, height int, base, accent color.Color) {
	g.drawCranium(img, width, height, base)
	g.drawEyeSockets(img, width, height, accent)
	g.drawNasalCavity(img, width, height, accent)
}

// drawCranium draws the skull cranium as a large oval.
func (g *Generator) drawCranium(img *image.RGBA, width, height int, base color.Color) {
	centerX, centerY := width/2, height/3
	radiusX, radiusY := width/3, height/4
	for y := 0; y < height*2/3; y++ {
		for x := 0; x < width; x++ {
			if g.isInsideEllipse(x, y, centerX, centerY, radiusX, radiusY) {
				img.Set(x, y, base)
			}
		}
	}
}

// isInsideEllipse checks if a point is inside an ellipse.
func (g *Generator) isInsideEllipse(x, y, centerX, centerY, radiusX, radiusY int) bool {
	dx := float64(x - centerX)
	dy := float64(y - centerY)
	dist := (dx*dx)/float64(radiusX*radiusX) + (dy*dy)/float64(radiusY*radiusY)
	return dist <= 1.0
}

// drawEyeSockets draws two circular eye sockets.
func (g *Generator) drawEyeSockets(img *image.RGBA, width, height int, accent color.Color) {
	eyeY := height / 3
	eyeRadius := width / 10
	eyePositions := []int{width / 3, width * 2 / 3}

	for _, eyeX := range eyePositions {
		g.drawCircle(img, eyeX, eyeY, eyeRadius, width, height, accent)
	}
}

// drawCircle draws a filled circle at the specified position.
func (g *Generator) drawCircle(img *image.RGBA, centerX, centerY, radius, width, height int, c color.Color) {
	for y := centerY - radius; y <= centerY+radius; y++ {
		for x := centerX - radius; x <= centerX+radius; x++ {
			if g.isInBounds(x, y, width, height) && g.isInsideCircle(x, y, centerX, centerY, radius) {
				img.Set(x, y, c)
			}
		}
	}
}

// isInsideCircle checks if a point is inside a circle.
func (g *Generator) isInsideCircle(x, y, centerX, centerY, radius int) bool {
	dx := x - centerX
	dy := y - centerY
	return dx*dx+dy*dy <= radius*radius
}

// isInBounds checks if coordinates are within image bounds.
func (g *Generator) isInBounds(x, y, width, height int) bool {
	return x >= 0 && x < width && y >= 0 && y < height
}

// drawNasalCavity draws a triangular nasal cavity.
func (g *Generator) drawNasalCavity(img *image.RGBA, width, height int, accent color.Color) {
	noseY := height / 2
	noseHeight := height / 8

	for y := noseY; y < noseY+noseHeight; y++ {
		yOffset := y - noseY
		xStart := width/2 - yOffset/2
		xEnd := width/2 + yOffset/2

		for x := xStart; x <= xEnd; x++ {
			if g.isInBounds(x, y, width, height) {
				img.Set(x, y, accent)
			}
		}
	}
}

func (g *Generator) drawChain(img *image.RGBA, width, height int, base, accent color.Color) {
	// Draw chain links hanging down
	numLinks := 4
	linkHeight := height / (numLinks + 1)
	linkWidth := width / 4

	for i := 0; i < numLinks; i++ {
		y := (i + 1) * linkHeight

		// Draw link oval outline
		centerX := width / 2
		radiusX := linkWidth / 2
		radiusY := linkHeight / 3

		for dy := -radiusY; dy <= radiusY; dy++ {
			for dx := -radiusX; dx <= radiusX; dx++ {
				px := centerX + dx
				py := y + dy
				if px >= 0 && px < width && py >= 0 && py < height {
					dist := float64(dx*dx)/float64(radiusX*radiusX) + float64(dy*dy)/float64(radiusY*radiusY)
					// Only draw outline
					if dist >= 0.6 && dist <= 1.0 {
						if i%2 == 0 {
							img.Set(px, py, base)
						} else {
							img.Set(px, py, accent)
						}
					}
				}
			}
		}
	}
}

func (g *Generator) drawWeb(img *image.RGBA, width, height int, base, accent color.Color, rng *rand.Rand) {
	centerX, centerY := width/2, height/2

	// Draw radial threads
	numRadial := 8
	for i := 0; i < numRadial; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(numRadial)
		endX := centerX + int(float64(width/2)*math.Cos(angle))
		endY := centerY + int(float64(height/2)*math.Sin(angle))
		g.drawLine(img, centerX, centerY, endX, endY, base)
	}

	// Draw spiral threads
	numSpirals := 4
	for i := 1; i <= numSpirals; i++ {
		radius := int(float64(min(width, height)) * float64(i) / float64(numSpirals*2))
		steps := 32
		for j := 0; j < steps; j++ {
			angle := float64(j) * 2.0 * math.Pi / float64(steps)
			x := centerX + int(float64(radius)*math.Cos(angle))
			y := centerY + int(float64(radius)*math.Sin(angle))
			if x >= 0 && x < width && y >= 0 && y < height {
				img.Set(x, y, accent)
			}
		}
	}
}

func (g *Generator) drawMoss(img *image.RGBA, width, height int, base, accent color.Color, rng *rand.Rand) {
	// Draw patches of moss
	numPatches := 10 + rng.Intn(10)
	for i := 0; i < numPatches; i++ {
		cx := rng.Intn(width)
		cy := rng.Intn(height)
		size := 2 + rng.Intn(4)

		// Draw irregular patch
		for dy := -size; dy <= size; dy++ {
			for dx := -size; dx <= size; dx++ {
				px := cx + dx
				py := cy + dy
				if px >= 0 && px < width && py >= 0 && py < height {
					// Irregular edges
					if rng.Float64() < 0.6 {
						if rng.Float64() < 0.8 {
							img.Set(px, py, base)
						} else {
							img.Set(px, py, accent)
						}
					}
				}
			}
		}
	}
}

func (g *Generator) drawGraffiti(img *image.RGBA, width, height int, base, accent color.Color, rng *rand.Rand) {
	numMarks := 3 + rng.Intn(5)
	for i := 0; i < numMarks; i++ {
		if rng.Float64() < 0.5 {
			g.drawGraffitiLine(img, width, height, base, accent, rng)
		} else {
			g.drawGraffitiBlob(img, width, height, base, accent, rng)
		}
	}
}

// drawGraffitiLine draws a random line mark on the graffiti.
func (g *Generator) drawGraffitiLine(img *image.RGBA, width, height int, base, accent color.Color, rng *rand.Rand) {
	x1 := rng.Intn(width)
	y1 := rng.Intn(height)
	x2 := x1 + rng.Intn(width/2) - width/4
	y2 := y1 + rng.Intn(height/2) - height/4

	lineColor := base
	if rng.Float64() >= 0.5 {
		lineColor = accent
	}

	g.drawLine(img, x1, y1, x2, y2, lineColor)
}

// drawGraffitiBlob draws a circular blob mark on the graffiti.
func (g *Generator) drawGraffitiBlob(img *image.RGBA, width, height int, base, accent color.Color, rng *rand.Rand) {
	cx := rng.Intn(width)
	cy := rng.Intn(height)
	radius := 3 + rng.Intn(6)

	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy > radius*radius {
				continue
			}

			px := cx + dx
			py := cy + dy

			if px < 0 || px >= width || py < 0 || py >= height {
				continue
			}

			pixelColor := base
			if rng.Float64() >= 0.5 {
				pixelColor = accent
			}
			img.Set(px, py, pixelColor)
		}
	}
}

// Helper functions

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// logDebug logs a debug message if logger and level are configured.
func (g *Generator) logDebug(msg string, fields logrus.Fields) {
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(fields).Debug(msg)
	}
}

// logInfo logs an info message if logger is configured.
func (g *Generator) logInfo(msg string, fields logrus.Fields) {
	if g.logger != nil {
		g.logger.WithFields(fields).Info(msg)
	}
}

// logError logs an error message if logger is configured.
func (g *Generator) logError(msg string, err error, fields ...logrus.Fields) {
	if g.logger != nil {
		entry := g.logger.WithError(err)
		if len(fields) > 0 {
			entry = entry.WithFields(fields[0])
		}
		entry.Error(msg)
	}
}
