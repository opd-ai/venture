// Package main - humanoidtest - Visual validation tool for Phase 5.2 humanoid enhancements.package humanoidtest

// This tool displays humanoid sprites with directional variants, genre-specific styling,
// and equipment overlays to verify anatomical template improvements.
package main

import (
	"flag"
	"fmt"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/sirupsen/logrus"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/opd-ai/venture/pkg/rendering/shapes"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

const (
	screenWidth  = 1024
	screenHeight = 768
	spriteScale  = 8 // Display sprites at 8x zoom for inspection
)

// Game holds the game state for the humanoid test viewer.
type Game struct {
	sprites      []*ebiten.Image
	spriteNames  []string
	currentIndex int
	seed         int64
	genreID      string
	shapeGen     *shapes.Generator
	spriteGen    *sprites.Generator
	paletteGen   *palette.Generator
	currentPal   *palette.Palette
	logger       *logrus.Logger
}

// NewGame creates a new humanoid test game.
func NewGame(seed int64, genreID string, logger *logrus.Logger) (*Game, error) {
	shapeGen := shapes.NewGenerator()
	spriteGen := sprites.NewGenerator()
	paletteGen := palette.NewGenerator()

	// Generate palette for this genre
	pal, err := paletteGen.Generate(genreID, seed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate palette: %w", err)
	}

	g := &Game{
		sprites:      make([]*ebiten.Image, 0),
		spriteNames:  make([]string, 0),
		currentIndex: 0,
		seed:         seed,
		genreID:      genreID,
		shapeGen:     shapeGen,
		spriteGen:    spriteGen,
		paletteGen:   paletteGen,
		currentPal:   pal,
		logger:       logger,
	}

	// Generate all test sprites
	if err := g.generateAllSprites(); err != nil {
		return nil, err
	}

	return g, nil
}

// generateAllSprites generates all humanoid variants for display.
func (g *Game) generateAllSprites() error {
	// Standard humanoid (28x28 - player size)
	g.addSprite("Player Down", g.generateHumanoid(28, 28, "player", "down", "", false, false))
	g.addSprite("Player Up", g.generateHumanoid(28, 28, "player", "up", "", false, false))
	g.addSprite("Player Left", g.generateHumanoid(28, 28, "player", "left", "", false, false))
	g.addSprite("Player Right", g.generateHumanoid(28, 28, "player", "right", "", false, false))

	// With equipment (28x28)
	g.addSprite("Player w/ Weapon", g.generateHumanoid(28, 28, "player", "down", "", true, false))
	g.addSprite("Player w/ Shield", g.generateHumanoid(28, 28, "player", "down", "", false, true))
	g.addSprite("Player w/ Both", g.generateHumanoid(28, 28, "player", "down", "", true, true))

	// Genre-specific humanoids (32x32)
	g.addSprite("Fantasy Knight", g.generateHumanoid(32, 32, "knight", "down", "fantasy", true, true))
	g.addSprite("Sci-Fi Android", g.generateHumanoid(32, 32, "warrior", "down", "scifi", true, false))
	g.addSprite("Horror Creature", g.generateHumanoid(32, 32, "humanoid", "down", "horror", false, false))
	g.addSprite("Cyberpunk Merc", g.generateHumanoid(32, 32, "warrior", "down", "cyberpunk", true, false))
	g.addSprite("Postapoc Survivor", g.generateHumanoid(32, 32, "warrior", "down", "postapoc", true, true))

	// Direction comparison (32x32)
	g.addSprite("NPC Down", g.generateHumanoid(32, 32, "npc", "down", g.genreID, false, false))
	g.addSprite("NPC Up", g.generateHumanoid(32, 32, "npc", "up", g.genreID, false, false))
	g.addSprite("NPC Left", g.generateHumanoid(32, 32, "npc", "left", g.genreID, false, false))
	g.addSprite("NPC Right", g.generateHumanoid(32, 32, "npc", "right", g.genreID, false, false))

	// All genres with equipment (32x32)
	g.addSprite("Fantasy Equipped", g.generateHumanoid(32, 32, "knight", "right", "fantasy", true, true))
	g.addSprite("SciFi Equipped", g.generateHumanoid(32, 32, "warrior", "right", "scifi", true, false))
	g.addSprite("Horror Equipped", g.generateHumanoid(32, 32, "warrior", "right", "horror", true, false))
	g.addSprite("Cyberpunk Equipped", g.generateHumanoid(32, 32, "warrior", "right", "cyberpunk", true, true))
	g.addSprite("Postapoc Equipped", g.generateHumanoid(32, 32, "warrior", "right", "postapoc", true, true))

	return nil
}

// generateHumanoid generates a humanoid sprite with specified parameters.
func (g *Game) generateHumanoid(width, height int, entityType, facing, genre string, hasWeapon, hasShield bool) *ebiten.Image {
	config := sprites.Config{
		Type:       sprites.SpriteEntity,
		Width:      width,
		Height:     height,
		Seed:       g.seed,
		Palette:    g.currentPal,
		Complexity: 0.7,
		Custom: map[string]interface{}{
			"entityType": entityType,
			"facing":     facing,
			"hasWeapon":  hasWeapon,
			"hasShield":  hasShield,
		},
	}

	// Add genre if specified
	if genre != "" {
		config.Custom["genre"] = genre
	}

	sprite, err := g.spriteGen.Generate(config)
	if err != nil {
		g.logger.WithError(err).WithFields(logrus.Fields{
			"variant": variant,
			"width":   width,
			"height":  height,
		}).Error("failed to generate sprite")
		return ebiten.NewImage(width, height)
	}

	return sprite
}

// addSprite adds a sprite and its name to the display list.
func (g *Game) addSprite(name string, sprite *ebiten.Image) {
	g.spriteNames = append(g.spriteNames, name)
	g.sprites = append(g.sprites, sprite)
}

// Update handles input and state updates.
func (g *Game) Update() error {
	// Navigate with arrow keys
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		g.currentIndex = (g.currentIndex + 1) % len(g.sprites)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		g.currentIndex--
		if g.currentIndex < 0 {
			g.currentIndex = len(g.sprites) - 1
		}
	}

	// Jump with number keys
	for i := 0; i < 9; i++ {
		key := ebiten.Key(int(ebiten.Key0) + i + 1)
		if inpututil.IsKeyJustPressed(key) && i < len(g.sprites) {
			g.currentIndex = i
		}
	}

	// Quit with ESC
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return fmt.Errorf("quit")
	}

	return nil
}

// Draw renders the current state to the screen.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 40, G: 40, B: 50, A: 255})

	if len(g.sprites) == 0 {
		return
	}

	currentSprite := g.sprites[g.currentIndex]
	spriteName := g.spriteNames[g.currentIndex]

	// Draw large centered sprite (8x scale)
	bounds := currentSprite.Bounds()
	scaledWidth := bounds.Dx() * spriteScale
	scaledHeight := bounds.Dy() * spriteScale

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(spriteScale, spriteScale)
	opts.GeoM.Translate(
		float64(screenWidth/2-scaledWidth/2),
		float64(screenHeight/2-scaledHeight/2),
	)
	screen.DrawImage(currentSprite, opts)

	// Draw checkerboard background behind sprite for transparency visualization
	checkerSize := 4 * spriteScale
	for y := screenHeight/2 - scaledHeight/2; y < screenHeight/2+scaledHeight/2; y += checkerSize {
		for x := screenWidth/2 - scaledWidth/2; x < screenWidth/2+scaledWidth/2; x += checkerSize {
			if ((x/checkerSize)+(y/checkerSize))%2 == 0 {
				ebitenutil.DrawRect(screen, float64(x), float64(y), float64(checkerSize), float64(checkerSize),
					color.RGBA{R: 50, G: 50, B: 60, A: 255})
			}
		}
	}

	// Re-draw sprite on top of checkerboard
	screen.DrawImage(currentSprite, opts)

	// Draw thumbnail strip at bottom
	thumbnailY := screenHeight - 60
	thumbnailSpacing := 50
	startX := screenWidth/2 - (len(g.sprites)*thumbnailSpacing)/2

	for i, sprite := range g.sprites {
		thumbOpts := &ebiten.DrawImageOptions{}
		thumbOpts.GeoM.Scale(2, 2) // 2x scale for thumbnails
		thumbOpts.GeoM.Translate(float64(startX+i*thumbnailSpacing), float64(thumbnailY))

		// Highlight current sprite
		if i == g.currentIndex {
			ebitenutil.DrawRect(screen,
				float64(startX+i*thumbnailSpacing-2),
				float64(thumbnailY-2),
				float64(sprite.Bounds().Dx()*2+4),
				float64(sprite.Bounds().Dy()*2+4),
				color.RGBA{R: 255, G: 255, B: 100, A: 255})
		}

		screen.DrawImage(sprite, thumbOpts)
	}

	// Draw UI text
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Humanoid Test - Phase 5.2"), 10, 10)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Seed: %d | Genre: %s", g.seed, g.genreID), 10, 30)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Sprite %d/%d: %s", g.currentIndex+1, len(g.sprites), spriteName), 10, 50)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Dimensions: %dx%d pixels", bounds.Dx(), bounds.Dy()), 10, 70)
	ebitenutil.DebugPrintAt(screen, "Controls: LEFT/RIGHT arrows to navigate, ESC to quit", 10, screenHeight-30)
	ebitenutil.DebugPrintAt(screen, "Numbers 1-9 to jump to sprite", 10, screenHeight-10)
}

// Layout returns the game's screen dimensions.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	// Parse command-line flags
	seed := flag.Int64("seed", 12345, "Random seed for generation")
	genreID := flag.String("genre", "fantasy", "Genre ID (fantasy, scifi, horror, cyberpunk, postapoc)")
	flag.Parse()

	// Initialize logger for test utility
	logger := logging.TestUtilityLogger("humanoidtest")
	logger.WithFields(logrus.Fields{
		"seed":  *seed,
		"genre": *genreID,
	}).Info("starting humanoid test")

	// Create game
	game, err := NewGame(*seed, *genreID, logger)
	if err != nil {
		logger.WithError(err).Fatal("failed to create game")
	}

	// Set window properties
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Humanoid Test - Phase 5.2 - Venture")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Run game
	if err := ebiten.RunGame(game); err != nil {
		if err.Error() != "quit" {
			logger.WithError(err).Fatal("game error")
		}
	}

	logger.Info("humanoid test complete")
}
