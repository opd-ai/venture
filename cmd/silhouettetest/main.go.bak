// Package main - silhouettetest - Visual validation tool for Phase 5.5 silhouette analysis.package silhouettetest

// This tool demonstrates silhouette analysis, outline generation, and contrast testing
// to verify sprite readability optimization systems.
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
	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

const (
	screenWidth  = 1024
	screenHeight = 768
	spriteScale  = 8
)

var (
	seed  = flag.Int64("seed", 12345, "Random seed for generation")
	genre = flag.String("genre", "fantasy", "Genre theme")
)

// Game holds the game state for the silhouette test viewer.
type Game struct {
	sprites      []*ebiten.Image
	silhouettes  []*ebiten.Image
	outlined     []*ebiten.Image
	analyses     []sprites.SilhouetteAnalysis
	spriteNames  []string
	currentIndex int
	viewMode     ViewMode
	bgColor      color.Color
	spriteGen    *sprites.Generator
	paletteGen   *palette.Generator
	currentPal   *palette.Palette
}

// ViewMode determines what is displayed.
type ViewMode int

const (
	ViewOriginal ViewMode = iota
	ViewSilhouette
	ViewOutlined
	ViewOnBackground
)

func (v ViewMode) String() string {
	switch v {
	case ViewOriginal:
		return "Original"
	case ViewSilhouette:
		return "Silhouette"
	case ViewOutlined:
		return "Outlined"
	case ViewOnBackground:
		return "On Background"
	default:
		return "Unknown"
	}
}

// NewGame creates a new silhouette test game.
func NewGame(seed int64, genreID string) (*Game, error) {
	spriteGen := sprites.NewGenerator()
	paletteGen := palette.NewGenerator()

	// Generate palette for this genre
	pal, err := paletteGen.Generate(genreID, seed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate palette: %w", err)
	}

	g := &Game{
		sprites:      make([]*ebiten.Image, 0),
		silhouettes:  make([]*ebiten.Image, 0),
		outlined:     make([]*ebiten.Image, 0),
		analyses:     make([]sprites.SilhouetteAnalysis, 0),
		spriteNames:  make([]string, 0),
		currentIndex: 0,
		viewMode:     ViewOriginal,
		bgColor:      color.RGBA{50, 50, 50, 255},
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

// generateAllSprites generates various sprites for testing.
func (g *Game) generateAllSprites() error {
	// Humanoid entities
	g.addEntity("Player (Down)", "player", "down", false, false)
	g.addEntity("Player w/ Equipment", "player", "down", true, true)
	g.addEntity("Knight (Fantasy)", "knight", "down", true, true, "fantasy")
	g.addEntity("Android (Sci-Fi)", "warrior", "down", true, false, "scifi")

	// Monster types
	g.addEntity("Quadruped", "quadruped", "", false, false)
	g.addEntity("Flying", "flying", "", false, false)
	g.addEntity("Blob", "blob", "", false, false)
	g.addEntity("Serpentine", "serpentine", "", false, false)
	g.addEntity("Mechanical", "mechanical", "", false, false)

	// Boss entities
	g.addEntity("Boss (2x)", "boss", "down", false, false)

	// Item sprites
	g.addItem("Sword (Common)", sprites.ItemSword, sprites.RarityCommon)
	g.addItem("Sword (Legendary)", sprites.ItemSword, sprites.RarityLegendary)
	g.addItem("Potion (Rare)", sprites.ItemPotion, sprites.RarityRare)
	g.addItem("Helmet", sprites.ItemHelmet, sprites.RarityUncommon)

	return nil
}

// addEntity adds an entity sprite with analysis.
func (g *Game) addEntity(name, entityType, facing string, hasWeapon, hasShield bool, customGenre ...string) {
	config := sprites.Config{
		Type:       sprites.SpriteEntity,
		Width:      32,
		Height:     32,
		Complexity: 0.7,
		Palette:    g.currentPal,
		Seed:       *seed + int64(len(g.sprites)),
		GenreID:    *genre,
		Custom: map[string]interface{}{
			"entityType": entityType,
			"facing":     facing,
			"hasWeapon":  hasWeapon,
			"hasShield":  hasShield,
		},
	}

	if len(customGenre) > 0 {
		config.Custom["genre"] = customGenre[0]
	}

	sprite, err := g.spriteGen.Generate(config)
	if err != nil {
		log.Printf("Failed to generate entity %s: %v", name, err)
		sprite = ebiten.NewImage(32, 32)
	}

	g.processSprite(name, sprite)
}

// addItem adds an item sprite with analysis.
func (g *Game) addItem(name string, itemType sprites.ItemType, rarity sprites.ItemRarity) {
	config := sprites.Config{
		Type:       sprites.SpriteItem,
		Width:      32,
		Height:     32,
		Complexity: 0.7,
		Palette:    g.currentPal,
		Seed:       *seed + int64(len(g.sprites)),
		GenreID:    *genre,
		Custom: map[string]interface{}{
			"itemType": itemType,
			"rarity":   rarity,
		},
	}

	sprite, err := g.spriteGen.Generate(config)
	if err != nil {
		log.Printf("Failed to generate item %s: %v", name, err)
		sprite = ebiten.NewImage(32, 32)
	}

	g.processSprite(name, sprite)
}

// processSprite analyzes sprite and generates variants.
func (g *Game) processSprite(name string, sprite *ebiten.Image) {
	// Analyze silhouette
	analysis := sprites.AnalyzeSilhouette(sprite)

	// Generate silhouette
	silhouette := sprites.GenerateSilhouette(sprite)

	// Generate outlined version
	outlineConfig := sprites.DefaultOutlineConfig()
	outlined := sprites.AddOutline(sprite, outlineConfig.Color, outlineConfig.Thickness)

	g.sprites = append(g.sprites, sprite)
	g.silhouettes = append(g.silhouettes, silhouette)
	g.outlined = append(g.outlined, outlined)
	g.analyses = append(g.analyses, analysis)
	g.spriteNames = append(g.spriteNames, name)
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

	// Change view mode with number keys
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		g.viewMode = ViewOriginal
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		g.viewMode = ViewSilhouette
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		g.viewMode = ViewOutlined
	}
	if inpututil.IsKeyJustPressed(ebiten.Key4) {
		g.viewMode = ViewOnBackground
	}

	// Change background color with B key
	if inpututil.IsKeyJustPressed(ebiten.KeyB) {
		g.cycleBackground()
	}

	// Quit with ESC
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return fmt.Errorf("quit")
	}

	return nil
}

// cycleBackground cycles through different background colors.
func (g *Game) cycleBackground() {
	r, gr, b, _ := g.bgColor.RGBA()
	r, gr, b = r/257, gr/257, b/257

	// Cycle: dark gray -> light gray -> dark blue -> dark green -> brown -> back to dark gray
	switch {
	case r == 50 && gr == 50 && b == 50:
		g.bgColor = color.RGBA{180, 180, 180, 255} // Light gray
	case r == 180:
		g.bgColor = color.RGBA{20, 20, 80, 255} // Dark blue
	case b == 80:
		g.bgColor = color.RGBA{20, 60, 20, 255} // Dark green
	case gr == 60:
		g.bgColor = color.RGBA{80, 60, 40, 255} // Brown
	default:
		g.bgColor = color.RGBA{50, 50, 50, 255} // Dark gray
	}
}

// Draw renders the current frame.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{30, 30, 40, 255})

	if g.currentIndex >= len(g.sprites) {
		return
	}

	// Get current sprite data
	var displaySprite *ebiten.Image
	switch g.viewMode {
	case ViewOriginal:
		displaySprite = g.sprites[g.currentIndex]
	case ViewSilhouette:
		displaySprite = g.silhouettes[g.currentIndex]
	case ViewOutlined:
		displaySprite = g.outlined[g.currentIndex]
	case ViewOnBackground:
		displaySprite = sprites.TestOnBackground(g.sprites[g.currentIndex], g.bgColor)
	}

	analysis := g.analyses[g.currentIndex]
	name := g.spriteNames[g.currentIndex]

	// Draw large sprite
	displaySize := 32 * spriteScale
	displayX := (screenWidth - displaySize) / 2
	displayY := 100

	// Draw background for sprite
	if g.viewMode != ViewOnBackground {
		// Checkerboard
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
	} else {
		// Solid background
		ebitenutil.DrawRect(screen, float64(displayX), float64(displayY),
			float64(displaySize), float64(displaySize), g.bgColor)
	}

	// Draw sprite
	if displaySprite != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(spriteScale), float64(spriteScale))
		op.GeoM.Translate(float64(displayX), float64(displayY))
		screen.DrawImage(displaySprite, op)
	}

	// Draw border
	borderColor := color.RGBA{255, 255, 255, 255}
	ebitenutil.DrawRect(screen, float64(displayX-2), float64(displayY-2), float64(displaySize+4), 2, borderColor)
	ebitenutil.DrawRect(screen, float64(displayX-2), float64(displayY+displaySize), float64(displaySize+4), 2, borderColor)
	ebitenutil.DrawRect(screen, float64(displayX-2), float64(displayY-2), 2, float64(displaySize+4), borderColor)
	ebitenutil.DrawRect(screen, float64(displayX+displaySize), float64(displayY-2), 2, float64(displaySize+4), borderColor)

	// Draw analysis info
	infoY := displayY + displaySize + 30
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Sprite: %s", name), displayX, infoY)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("View Mode: %s", g.viewMode.String()), displayX, infoY+20)

	// Analysis metrics
	metricsY := infoY + 50
	ebitenutil.DebugPrintAt(screen, "Silhouette Analysis:", displayX, metricsY)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Overall Score: %.3f", analysis.OverallScore), displayX, metricsY+20)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Quality: %s", analysis.GetQuality().String()), displayX, metricsY+40)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Compactness: %.3f", analysis.Compactness), displayX, metricsY+60)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Coverage: %.3f (%.1f%%)", analysis.Coverage, analysis.Coverage*100), displayX, metricsY+80)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Edge Clarity: %.3f", analysis.EdgeClarity), displayX, metricsY+100)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Needs Improvement: %v", analysis.NeedsImprovement()), displayX, metricsY+120)

	// Pixel counts
	pixelY := metricsY + 150
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Opaque Pixels: %d", analysis.OpaquePixels), displayX, pixelY)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Perimeter Pixels: %d", analysis.PerimeterPixels), displayX, pixelY+20)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Total Pixels: %d", analysis.TotalPixels), displayX, pixelY+40)

	// Instructions
	instructions := []string{
		"Arrow Keys: Navigate sprites",
		"1: Original   2: Silhouette",
		"3: Outlined   4: On Background",
		"B: Change background color",
		fmt.Sprintf("Sprite %d/%d", g.currentIndex+1, len(g.sprites)),
		fmt.Sprintf("Seed: %d  Genre: %s", *seed, *genre),
	}
	for i, line := range instructions {
		ebitenutil.DebugPrintAt(screen, line, 10, 10+i*15)
	}

	// Quality color indicator
	qualityColor := getQualityColor(analysis.GetQuality())
	ebitenutil.DrawRect(screen, 10, 120, 30, 30, qualityColor)
	ebitenutil.DebugPrintAt(screen, "Quality", 45, 130)
}

// getQualityColor returns a color representing the quality level.
func getQualityColor(quality sprites.SilhouetteQuality) color.Color {
	switch quality {
	case sprites.QualityPoor:
		return color.RGBA{200, 50, 50, 255} // Red
	case sprites.QualityFair:
		return color.RGBA{200, 150, 50, 255} // Orange
	case sprites.QualityGood:
		return color.RGBA{150, 200, 50, 255} // Yellow-Green
	case sprites.QualityExcellent:
		return color.RGBA{50, 200, 50, 255} // Green
	default:
		return color.RGBA{100, 100, 100, 255} // Gray
	}
}

// Layout defines the game's logical screen size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	flag.Parse()

	game, err := NewGame(*seed, *genre)
	if err != nil {
		log.Fatal(err)
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Silhouette Analysis Test - Phase 5.5")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(game); err != nil && err.Error() != "quit" {
		log.Fatal(err)
	}
}
