package main

import (
	"flag"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/opd-ai/venture/pkg/rendering/shapes"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

var (
	seed    = flag.Int64("seed", 12345, "Random seed for generation")
	genreID = flag.String("genre", "fantasy", "Genre ID (fantasy, scifi, horror, cyberpunk, postapoc)")
)

// Game represents the anatomy test application.
type Game struct {
	spriteGen     *sprites.Generator
	palette       *palette.Palette
	currentEntity int
	sprites       []*ebiten.Image
	labels        []string
	logger        *logrus.Logger
}

// NewGame creates a new anatomy test game.
func NewGame(seed int64, genreID string, logger *logrus.Logger) (*Game, error) {
	spriteGen := sprites.NewGenerator()

	// Generate palette
	paletteGen := spriteGen.GetPaletteGenerator()
	pal, err := paletteGen.Generate(genreID, seed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate palette: %w", err)
	}

	game := &Game{
		spriteGen:     spriteGen,
		palette:       pal,
		currentEntity: 0,
		sprites:       make([]*ebiten.Image, 0),
		labels:        make([]string, 0),
		logger:        logger,
	}

	// Generate test sprites for all entity types
	game.generateTestSprites(seed)

	return game, nil
}

// generateTestSprites generates sprites for all anatomical templates.
func (g *Game) generateTestSprites(seed int64) {
	entityTypes := []struct {
		name string
		size int
	}{
		{"humanoid (player 28x28)", 28},
		{"humanoid (NPC 32x32)", 32},
		{"quadruped (wolf 32x32)", 32},
		{"blob (slime 32x32)", 32},
		{"mechanical (robot 32x32)", 32},
		{"flying (dragon 32x32)", 32},
	}

	entityNames := []string{
		"humanoid",
		"humanoid",
		"quadruped",
		"blob",
		"mechanical",
		"flying",
	}

	// Generate sprites for each entity type
	for i, entityType := range entityTypes {
		config := sprites.Config{
			Type:       sprites.SpriteEntity,
			Width:      entityType.size,
			Height:     entityType.size,
			Seed:       seed + int64(i*100),
			Palette:    g.palette,
			Complexity: 0.5,
			Variation:  i,
			GenreID:    *genreID,
			Custom: map[string]interface{}{
				"entityType": entityNames[i],
			},
		}

		sprite, err := g.spriteGen.Generate(config)
		if err != nil {
			if g.logger != nil {
				g.logger.WithError(err).WithField("entityType", entityType.name).Warn("failed to generate sprite")
			}
			continue
		}

		g.sprites = append(g.sprites, sprite)
		g.labels = append(g.labels, entityType.name)
	}

	// Generate individual shape test sprites
	shapeTypes := []struct {
		name      string
		shapeType shapes.ShapeType
	}{
		{"Ellipse", shapes.ShapeEllipse},
		{"Capsule", shapes.ShapeCapsule},
		{"Bean", shapes.ShapeBean},
		{"Wedge", shapes.ShapeWedge},
		{"Shield", shapes.ShapeShield},
		{"Blade", shapes.ShapeBlade},
		{"Skull", shapes.ShapeSkull},
	}

	shapeGen := shapes.NewGenerator()
	for i, st := range shapeTypes {
		shapeConfig := shapes.Config{
			Type:      st.shapeType,
			Width:     48,
			Height:    48,
			Color:     g.palette.Primary,
			Seed:      seed + int64(i*10),
			Smoothing: 0.2,
		}

		shapeImg, err := shapeGen.Generate(shapeConfig)
		if err != nil {
			if g.logger != nil {
				g.logger.WithError(err).WithField("shapeType", st.name).Warn("failed to generate shape")
			}
			continue
		}

		g.sprites = append(g.sprites, shapeImg)
		g.labels = append(g.labels, fmt.Sprintf("Shape: %s", st.name))
	}
}

// Update updates the game state.
func (g *Game) Update() error {
	// Cycle through sprites with arrow keys
	if ebiten.IsKeyPressed(ebiten.KeyRight) && ebiten.IsKeyPressed(ebiten.KeyArrowRight) == false {
		g.currentEntity = (g.currentEntity + 1) % len(g.sprites)
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) && ebiten.IsKeyPressed(ebiten.KeyArrowLeft) == false {
		g.currentEntity = (g.currentEntity - 1 + len(g.sprites)) % len(g.sprites)
	}

	return nil
}

// Draw draws the game screen.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 40, G: 40, B: 40, A: 255})

	// Draw title
	ebitenutil.DebugPrintAt(screen, "Venture - Anatomical Template Viewer", 10, 10)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Seed: %d | Genre: %s", *seed, *genreID), 10, 25)
	ebitenutil.DebugPrintAt(screen, "Use LEFT/RIGHT arrow keys to navigate", 10, 40)
	ebitenutil.DebugPrintAt(screen, "Press ESC to quit", 10, 55)

	// Draw current sprite (large)
	if g.currentEntity < len(g.sprites) {
		sprite := g.sprites[g.currentEntity]
		label := g.labels[g.currentEntity]

		// Draw sprite centered and scaled up
		opts := &ebiten.DrawImageOptions{}
		scale := 8.0 // 8x scale for visibility
		opts.GeoM.Scale(scale, scale)

		// Center the sprite
		bounds := sprite.Bounds()
		centerX := float64(800/2) - float64(bounds.Dx())*scale/2
		centerY := float64(600/2) - float64(bounds.Dy())*scale/2
		opts.GeoM.Translate(centerX, centerY)

		screen.DrawImage(sprite, opts)

		// Draw label
		labelText := fmt.Sprintf("[%d/%d] %s", g.currentEntity+1, len(g.sprites), label)
		ebitenutil.DebugPrintAt(screen, labelText, 10, 80)

		// Draw sprite dimensions
		dimText := fmt.Sprintf("Dimensions: %dx%d pixels (shown at 8x scale)", bounds.Dx(), bounds.Dy())
		ebitenutil.DebugPrintAt(screen, dimText, 10, 95)
	}

	// Draw thumbnail strip at bottom
	thumbnailY := 500
	thumbnailStartX := 50
	thumbnailSpacing := 70

	for i, sprite := range g.sprites {
		opts := &ebiten.DrawImageOptions{}

		// Scale thumbnail (2x)
		thumbnailScale := 2.0
		opts.GeoM.Scale(thumbnailScale, thumbnailScale)

		// Position thumbnail
		x := float64(thumbnailStartX + i*thumbnailSpacing)
		y := float64(thumbnailY)
		opts.GeoM.Translate(x, y)

		// Highlight current thumbnail
		if i == g.currentEntity {
			// Draw selection box
			bounds := sprite.Bounds()
			boxW := float64(bounds.Dx()) * thumbnailScale
			boxH := float64(bounds.Dy()) * thumbnailScale

			// Draw white rectangle around selected thumbnail
			for offset := 0; offset < 3; offset++ {
				fx := x - float64(offset)
				fy := y - float64(offset)
				fw := boxW + float64(offset*2)
				fh := boxH + float64(offset*2)

				ebitenutil.DrawRect(screen, fx, fy, 2, fh, color.White)      // Left
				ebitenutil.DrawRect(screen, fx, fy, fw, 2, color.White)      // Top
				ebitenutil.DrawRect(screen, fx+fw-2, fy, 2, fh, color.White) // Right
				ebitenutil.DrawRect(screen, fx, fy+fh-2, fw, 2, color.White) // Bottom
			}
		}

		screen.DrawImage(sprite, opts)
	}

	// Draw instructions
	ebitenutil.DebugPrintAt(screen, "Phase 5.1 Implementation: Anatomical Templates & New Shapes", 10, 570)
}

// Layout returns the game screen size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 800, 600
}

func main() {
	flag.Parse()

	logger := logging.TestUtilityLogger("anatomytest")
	logger.WithFields(logrus.Fields{
		"seed":  *seed,
		"genre": *genreID,
	}).Info("starting anatomy test")

	game, err := NewGame(*seed, *genreID, logger)
	if err != nil {
		logger.WithError(err).Fatal("failed to create game")
	}

	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("Venture - Anatomical Template Viewer (Phase 5.1)")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(game); err != nil {
		logger.WithError(err).Fatal("game error")
	}
}
