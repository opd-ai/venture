// Package main - itemspritetest - Visual validation tool for Phase 5.4 item templates.
//
// This tool displays item sprites across all rarities and types to verify template-based
// generation, rarity visual indicators, and item recognizability.
package main

import (
	"flag"
	"fmt"
	"image/color"

	"github.com/opd-ai/venture/pkg/logging"
	"github.com/sirupsen/logrus"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

const (
	screenWidth  = 1024
	screenHeight = 768
	spriteScale  = 8 // Display sprites at 8x zoom for inspection
)

var (
	seed   = flag.Int64("seed", 12345, "Random seed for generation")
	genre  = flag.String("genre", "fantasy", "Genre theme")
	width  = flag.Int("width", 32, "Item sprite width")
	height = flag.Int("height", 32, "Item sprite height")
)

// Game holds the game state for the item sprite test viewer.
type Game struct {
	sprites       []*ebiten.Image
	spriteNames   []string
	thumbnails    []*ebiten.Image
	currentIndex  int
	seed          int64
	genreID       string
	spriteGen     *sprites.Generator
	paletteGen    *palette.Generator
	currentPal    *palette.Palette
	itemsPerRow   int
	thumbnailSize int
	zoom          int
	logger        *logrus.Logger
}

// NewGame creates a new item sprite test game.
func NewGame(seed int64, genreID string, logger *logrus.Logger) (*Game, error) {
	spriteGen := sprites.NewGenerator()
	paletteGen := palette.NewGenerator()

	// Generate palette for this genre
	pal, err := paletteGen.Generate(genreID, seed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate palette: %w", err)
	}

	g := &Game{
		sprites:       make([]*ebiten.Image, 0),
		spriteNames:   make([]string, 0),
		thumbnails:    make([]*ebiten.Image, 0),
		currentIndex:  0,
		seed:          seed,
		genreID:       genreID,
		spriteGen:     spriteGen,
		paletteGen:    paletteGen,
		currentPal:    pal,
		itemsPerRow:   10,
		thumbnailSize: 48,
		zoom:          spriteScale,
		logger:        logger,
	}

	// Generate all test sprites
	if err := g.generateAllSprites(); err != nil {
		return nil, err
	}

	// Generate thumbnails
	g.generateThumbnails()

	return g, nil
}

// generateAllSprites generates all item variants for display.
func (g *Game) generateAllSprites() error {
	itemTypes := []sprites.ItemType{
		sprites.ItemSword, sprites.ItemAxe, sprites.ItemBow, sprites.ItemStaff, sprites.ItemGun,
		sprites.ItemHelmet, sprites.ItemPotion, sprites.ItemScroll, sprites.ItemRing, sprites.ItemKey,
	}

	rarities := []sprites.ItemRarity{
		sprites.RarityCommon, sprites.RarityUncommon, sprites.RarityRare,
		sprites.RarityEpic, sprites.RarityLegendary,
	}

	// Generate one item per type at each rarity (10 types × 5 rarities = 50 items)
	for _, rarity := range rarities {
		for _, itemType := range itemTypes {
			sprite := g.generateItem(itemType, rarity)
			name := fmt.Sprintf("%s %s", rarity, itemType)
			g.addSprite(name, sprite)
		}
	}

	return nil
}

// generateItem generates a single item sprite.
func (g *Game) generateItem(itemType sprites.ItemType, rarity sprites.ItemRarity) *ebiten.Image {
	config := sprites.Config{
		Type:       sprites.SpriteItem,
		Width:      *width,
		Height:     *height,
		Complexity: 0.7,
		Palette:    g.currentPal,
		Seed:       g.seed,
		GenreID:    g.genreID,
		Custom: map[string]interface{}{
			"itemType": itemType,
			"rarity":   rarity,
		},
	}

	sprite, err := g.spriteGen.Generate(config)
	if err != nil {
		g.logger.WithError(rarity, itemType, err).WithField("item", "item %s %s").Error("failed to generate")
		return ebiten.NewImage(*width, *height)
	}

	return sprite
}

// addSprite adds a sprite and its name to the display list.
func (g *Game) addSprite(name string, sprite *ebiten.Image) {
	g.spriteNames = append(g.spriteNames, name)
	g.sprites = append(g.sprites, sprite)
}

// generateThumbnails creates thumbnail images for quick navigation.
func (g *Game) generateThumbnails() {
	for _, sprite := range g.sprites {
		thumbnail := ebiten.NewImage(g.thumbnailSize, g.thumbnailSize)

		// Scale to fit
		op := &ebiten.DrawImageOptions{}
		scaleX := float64(g.thumbnailSize) / float64(sprite.Bounds().Dx())
		scaleY := float64(g.thumbnailSize) / float64(sprite.Bounds().Dy())
		scale := scaleX
		if scaleY < scaleX {
			scale = scaleY
		}
		op.GeoM.Scale(scale, scale)

		// Center
		offsetX := (g.thumbnailSize - int(float64(sprite.Bounds().Dx())*scale)) / 2
		offsetY := (g.thumbnailSize - int(float64(sprite.Bounds().Dy())*scale)) / 2
		op.GeoM.Translate(float64(offsetX), float64(offsetY))

		thumbnail.DrawImage(sprite, op)
		g.thumbnails = append(g.thumbnails, thumbnail)
	}
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
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		g.currentIndex = (g.currentIndex + g.itemsPerRow) % len(g.sprites)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		g.currentIndex -= g.itemsPerRow
		if g.currentIndex < 0 {
			g.currentIndex += len(g.sprites)
		}
	}

	// Jump to rarity tier with number keys 1-5
	for i := 1; i <= 5; i++ {
		key := ebiten.Key(int(ebiten.Key0) + i)
		if inpututil.IsKeyJustPressed(key) {
			g.currentIndex = (i - 1) * g.itemsPerRow
		}
	}

	// Quit with ESC
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return fmt.Errorf("quit")
	}

	return nil
}

// Draw renders the current frame.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{40, 40, 50, 255})

	// Draw large version of current item with checkerboard background
	if g.currentIndex < len(g.sprites) {
		currentItem := g.sprites[g.currentIndex]
		displaySize := *width * g.zoom
		displayX := (screenWidth - displaySize) / 2
		displayY := 50

		// Draw checkerboard background
		checkerSize := 16
		for cy := 0; cy < displaySize/checkerSize; cy++ {
			for cx := 0; cx < displaySize/checkerSize; cx++ {
				var col color.Color
				if (cx+cy)%2 == 0 {
					col = color.RGBA{60, 60, 70, 255}
				} else {
					col = color.RGBA{50, 50, 60, 255}
				}
				ebitenutil.DrawRect(screen,
					float64(displayX+cx*checkerSize),
					float64(displayY+cy*checkerSize),
					float64(checkerSize), float64(checkerSize), col)
			}
		}

		// Draw item
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(g.zoom), float64(g.zoom))
		op.GeoM.Translate(float64(displayX), float64(displayY))
		screen.DrawImage(currentItem, op)

		// Draw border
		borderColor := color.RGBA{255, 255, 255, 255}
		ebitenutil.DrawRect(screen, float64(displayX-2), float64(displayY-2), float64(displaySize+4), 2, borderColor)
		ebitenutil.DrawRect(screen, float64(displayX-2), float64(displayY+displaySize), float64(displaySize+4), 2, borderColor)
		ebitenutil.DrawRect(screen, float64(displayX-2), float64(displayY-2), 2, float64(displaySize+4), borderColor)
		ebitenutil.DrawRect(screen, float64(displayX+displaySize), float64(displayY-2), 2, float64(displaySize+4), borderColor)

		// Draw item name
		itemName := g.spriteNames[g.currentIndex]
		ebitenutil.DebugPrintAt(screen, itemName, displayX, displayY+displaySize+10)
	}

	// Draw thumbnail grid at bottom
	thumbnailY := screenHeight - 5*(g.thumbnailSize+24) - 20
	thumbnailStartX := (screenWidth - g.itemsPerRow*(g.thumbnailSize+4)) / 2

	// Draw rarity sections
	rarities := []string{"Common", "Uncommon", "Rare", "Epic", "Legendary"}
	for row := 0; row < 5; row++ {
		// Draw rarity label
		labelY := thumbnailY + row*(g.thumbnailSize+24)
		ebitenutil.DebugPrintAt(screen, rarities[row], 10, labelY+g.thumbnailSize/2)

		for col := 0; col < g.itemsPerRow; col++ {
			idx := row*g.itemsPerRow + col
			if idx >= len(g.thumbnails) {
				break
			}

			x := thumbnailStartX + col*(g.thumbnailSize+4)
			y := labelY

			// Highlight current
			if idx == g.currentIndex {
				ebitenutil.DrawRect(screen, float64(x-2), float64(y-2),
					float64(g.thumbnailSize+4), float64(g.thumbnailSize+4),
					color.RGBA{255, 255, 0, 255})
			}

			// Draw thumbnail with background
			ebitenutil.DrawRect(screen, float64(x), float64(y),
				float64(g.thumbnailSize), float64(g.thumbnailSize),
				color.RGBA{30, 30, 40, 255})

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y))
			screen.DrawImage(g.thumbnails[idx], op)
		}
	}

	// Draw instructions
	instructions := []string{
		"Arrow Keys: Navigate items",
		"1-5: Jump to rarity tier",
		fmt.Sprintf("Item %d/%d", g.currentIndex+1, len(g.sprites)),
		fmt.Sprintf("Seed: %d  Genre: %s", g.seed, g.genreID),
	}
	for i, line := range instructions {
		ebitenutil.DebugPrintAt(screen, line, 10, 10+i*15)
	}

	// Draw category legend
	legend := []string{
		"Weapons: Sword, Axe, Bow, Staff, Gun",
		"Armor: Helmet",
		"Consumables: Potion, Scroll",
		"Accessories: Ring",
		"Quest: Key",
	}
	for i, line := range legend {
		ebitenutil.DebugPrintAt(screen, line, screenWidth-350, 10+i*15)
	}
}

// Layout defines the game's logical screen size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	flag.Parse()

	game, err := NewGame(*seed, *genre, logger)
	if err != nil {
		logger.WithError(err).Fatal("error")
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Item Sprite Template Test - Phase 5.4")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(game); err != nil && err.Error() != "quit" {
		logger.WithError(err).Fatal("error")
	}
}
