// Package main - monstertest - Visual validation tool for Phase 5.3 entity variety & monster templates.package monstertest

// This tool displays various monster archetypes to verify distinct silhouettes and boss scaling.
package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"

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
	spriteScale  = 6 // Display sprites at 6x zoom
)

// Game holds the game state for the monster test viewer.
type Game struct {
	sprites       []*ebiten.Image
	spriteNames   []string
	currentIndex  int
	seed          int64
	genreID       string
	shapeGen      *shapes.Generator
	spriteGen     *sprites.Generator
	paletteGen    *palette.Generator
	currentPal    *palette.Palette
}

// NewGame creates a new monster test game.
func NewGame(seed int64, genreID string) (*Game, error) {
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
	}

	// Generate all test sprites
	if err := g.generateAllSprites(); err != nil {
		return nil, err
	}

	return g, nil
}

// generateAllSprites generates all monster variants for display.
func (g *Game) generateAllSprites() error {
	// Monster archetypes (32x32 - standard enemy size)
	g.addSprite("Quadruped (Wolf)", g.generateMonster(32, 32, "wolf", false, 1.0))
	g.addSprite("Quadruped (Bear)", g.generateMonster(32, 32, "bear", false, 1.0))
	g.addSprite("Flying (Bird)", g.generateMonster(32, 32, "bird", false, 1.0))
	g.addSprite("Flying (Dragon)", g.generateMonster(32, 32, "dragon", false, 1.0))
	g.addSprite("Serpentine (Snake)", g.generateMonster(32, 32, "snake", false, 1.0))
	g.addSprite("Serpentine (Worm)", g.generateMonster(32, 32, "worm", false, 1.0))
	g.addSprite("Blob (Slime)", g.generateMonster(32, 32, "slime", false, 1.0))
	g.addSprite("Blob (Amoeba)", g.generateMonster(32, 32, "amoeba", false, 1.0))
	g.addSprite("Arachnid (Spider)", g.generateMonster(32, 32, "spider", false, 1.0))
	g.addSprite("Arachnid (Beetle)", g.generateMonster(32, 32, "beetle", false, 1.0))
	g.addSprite("Mechanical (Robot)", g.generateMonster(32, 32, "robot", false, 1.0))
	g.addSprite("Mechanical (Golem)", g.generateMonster(32, 32, "golem", false, 1.0))
	g.addSprite("Undead (Skeleton)", g.generateMonster(32, 32, "skeleton", false, 1.0))
	g.addSprite("Undead (Ghost)", g.generateMonster(32, 32, "ghost", false, 1.0))

	// Boss variants (64x64 - 2x scale)
	g.addSprite("Boss Wolf (2x)", g.generateMonster(64, 64, "wolf", true, 2.0))
	g.addSprite("Boss Dragon (2x)", g.generateMonster(64, 64, "dragon", true, 2.0))
	g.addSprite("Boss Spider (2x)", g.generateMonster(64, 64, "spider", true, 2.0))
	g.addSprite("Boss Golem (2x)", g.generateMonster(64, 64, "golem", true, 2.0))

	// Large bosses (96x96 - 3x scale)
	g.addSprite("Mega Boss Bear (3x)", g.generateMonster(96, 96, "bear", true, 3.0))
	g.addSprite("Mega Boss Worm (3x)", g.generateMonster(96, 96, "worm", true, 3.0))

	// Colossal bosses (128x128 - 4x scale)
	g.addSprite("Colossal Dragon (4x)", g.generateMonster(128, 128, "dragon", true, 4.0))
	g.addSprite("Colossal Slime (4x)", g.generateMonster(128, 128, "slime", true, 4.0))

	return nil
}

// generateMonster generates a monster sprite with specified parameters.
func (g *Game) generateMonster(width, height int, entityType string, isBoss bool, bossScale float64) *ebiten.Image {
	config := sprites.Config{
		Type:       sprites.SpriteEntity,
		Width:      width,
		Height:     height,
		Seed:       g.seed,
		Palette:    g.currentPal,
		Complexity: 0.7,
		Custom: map[string]interface{}{
			"entityType": entityType,
			"isBoss":     isBoss,
			"bossScale":  bossScale,
		},
	}

	sprite, err := g.spriteGen.Generate(config)
	if err != nil {
		log.Printf("Failed to generate sprite: %v", err)
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

	// Draw large centered sprite (scaled)
	bounds := currentSprite.Bounds()
	scaledWidth := bounds.Dx() * spriteScale
	scaledHeight := bounds.Dy() * spriteScale

	// Draw checkerboard background for transparency visualization
	checkerSize := 8
	startX := screenWidth/2 - scaledWidth/2
	startY := screenHeight/2 - scaledHeight/2
	for y := startY; y < startY+scaledHeight; y += checkerSize {
		for x := startX; x < startX+scaledWidth; x += checkerSize {
			if ((x-startX)/checkerSize+(y-startY)/checkerSize)%2 == 0 {
				ebitenutil.DrawRect(screen, float64(x), float64(y), float64(checkerSize), float64(checkerSize),
					color.RGBA{R: 50, G: 50, B: 60, A: 255})
			}
		}
	}

	// Draw sprite
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(float64(spriteScale), float64(spriteScale))
	opts.GeoM.Translate(float64(startX), float64(startY))
	screen.DrawImage(currentSprite, opts)

	// Draw thumbnail strip at bottom
	thumbnailY := screenHeight - 80
	thumbnailSpacing := 45
	thumbnailsPerRow := screenWidth / thumbnailSpacing
	startThumbX := screenWidth/2 - (thumbnailsPerRow*thumbnailSpacing)/2

	// Show thumbnails around current index
	visibleStart := (g.currentIndex / thumbnailsPerRow) * thumbnailsPerRow
	visibleEnd := visibleStart + thumbnailsPerRow
	if visibleEnd > len(g.sprites) {
		visibleEnd = len(g.sprites)
	}

	for i := visibleStart; i < visibleEnd; i++ {
		sprite := g.sprites[i]
		thumbIdx := i - visibleStart

		thumbOpts := &ebiten.DrawImageOptions{}
		// Scale to fit in thumbnail (max 40px)
		thumbScale := 40.0 / float64(sprite.Bounds().Dx())
		if thumbScale > 2.0 {
			thumbScale = 2.0 // Max 2x for small sprites
		}
		thumbOpts.GeoM.Scale(thumbScale, thumbScale)
		thumbOpts.GeoM.Translate(float64(startThumbX+thumbIdx*thumbnailSpacing), float64(thumbnailY))

		// Highlight current sprite
		if i == g.currentIndex {
			ebitenutil.DrawRect(screen,
				float64(startThumbX+thumbIdx*thumbnailSpacing-2),
				float64(thumbnailY-2),
				float64(sprite.Bounds().Dx())*thumbScale+4,
				float64(sprite.Bounds().Dy())*thumbScale+4,
				color.RGBA{R: 255, G: 255, B: 100, A: 255})
		}

		screen.DrawImage(sprite, thumbOpts)
	}

	// Draw UI text
	ebitenutil.DebugPrintAt(screen, "Monster Test - Phase 5.3", 10, 10)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Seed: %d | Genre: %s", g.seed, g.genreID), 10, 30)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Monster %d/%d: %s", g.currentIndex+1, len(g.sprites), spriteName), 10, 50)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Dimensions: %dx%d pixels", bounds.Dx(), bounds.Dy()), 10, 70)
	ebitenutil.DebugPrintAt(screen, "Controls: LEFT/RIGHT arrows, Numbers 1-9, ESC to quit", 10, screenHeight-30)

	// Draw archetype legend
	legendY := 100
	ebitenutil.DebugPrintAt(screen, "Archetypes:", 10, legendY)
	ebitenutil.DebugPrintAt(screen, "• Quadruped: 4 legs, horizontal", 10, legendY+20)
	ebitenutil.DebugPrintAt(screen, "• Flying: Wings, aerial", 10, legendY+40)
	ebitenutil.DebugPrintAt(screen, "• Serpentine: Elongated, snake", 10, legendY+60)
	ebitenutil.DebugPrintAt(screen, "• Blob: Amorphous mass", 10, legendY+80)
	ebitenutil.DebugPrintAt(screen, "• Arachnid: Multi-leg spread", 10, legendY+100)
	ebitenutil.DebugPrintAt(screen, "• Mechanical: Angular, geometric", 10, legendY+120)
	ebitenutil.DebugPrintAt(screen, "• Undead: Skeletal, translucent", 10, legendY+140)
}

// Layout returns the game's screen dimensions.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	// Parse command-line flags
	seed := flag.Int64("seed", 54321, "Random seed for generation")
	genreID := flag.String("genre", "fantasy", "Genre ID (fantasy, scifi, horror, cyberpunk, postapoc)")
	flag.Parse()

	// Create game
	game, err := NewGame(*seed, *genreID)
	if err != nil {
		log.Fatalf("Failed to create game: %v", err)
	}

	// Set window properties
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Monster Test - Phase 5.3 - Venture")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Run game
	if err := ebiten.RunGame(game); err != nil {
		if err.Error() != "quit" {
			log.Fatal(err)
		}
	}
}
