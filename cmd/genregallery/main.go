package main

import (
	"flag"
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

const (
	screenWidth  = 1280
	screenHeight = 720
	spriteSize   = 32
	scale        = 3
	gridColumns  = 12
	gridRows     = 6
)

// Game implements the Ebiten game interface for genre gallery.
type Game struct {
	generator  *sprites.CombinedGenerator
	genres     []string
	genreIndex int
	sprites    []*ebiten.Image
	configs    []sprites.Config
	seedBase   int64
	page       int
	maxPages   int
	showInfo   bool
	logger     *logrus.Logger
}

// NewGame creates a new genre gallery game.
func NewGame(seedBase int64, logger *logrus.Logger) *Game {
	return &Game{
		generator:  sprites.NewCombinedGenerator(200),
		genres:     []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apoc"},
		genreIndex: 0,
		seedBase:   seedBase,
		page:       0,
		maxPages:   5, // 5 pages × 72 sprites = 360 sprites per genre
		showInfo:   true,
		logger:     logger,
	}
}

// generateSprites generates a full page of sprites for the current genre.
func (g *Game) generateSprites() {
	g.sprites = make([]*ebiten.Image, 0, gridColumns*gridRows)
	g.configs = make([]sprites.Config, 0, gridColumns*gridRows)

	rng := rand.New(rand.NewSource(g.seedBase + int64(g.page)*1000 + int64(g.genreIndex)*10000))

	spriteTypes := []sprites.SpriteType{
		sprites.SpriteEntity,
		sprites.SpriteItem,
		sprites.SpriteTile,
		sprites.SpriteParticle,
	}

	for i := 0; i < gridColumns*gridRows; i++ {
		config := sprites.Config{
			Type:       spriteTypes[rng.Intn(len(spriteTypes))],
			Width:      spriteSize,
			Height:     spriteSize,
			Seed:       rng.Int63(),
			GenreID:    g.genres[g.genreIndex],
			Complexity: 0.3 + rng.Float64()*0.4, // 0.3-0.7
			Variation:  rng.Intn(3),
		}

		sprite, err := g.generator.Generate(config)
		if err != nil {
			g.logger.WithError(err).WithFields(logrus.Fields{
				"type":  config.Type,
				"genre": config.GenreID,
				"seed":  config.Seed,
			}).Error("failed to generate sprite")
			continue
		}

		g.sprites = append(g.sprites, sprite)
		g.configs = append(g.configs, config)
	}
}

// Update handles game logic updates.
func (g *Game) Update() error {
	needsRegenerate := false

	// Handle input
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return fmt.Errorf("quit")
	}

	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		if g.genreIndex < len(g.genres)-1 {
			g.genreIndex++
			g.page = 0
			needsRegenerate = true
		}
		time.Sleep(200 * time.Millisecond)
	}

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		if g.genreIndex > 0 {
			g.genreIndex--
			g.page = 0
			needsRegenerate = true
		}
		time.Sleep(200 * time.Millisecond)
	}

	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		if g.page < g.maxPages-1 {
			g.page++
			needsRegenerate = true
		}
		time.Sleep(200 * time.Millisecond)
	}

	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		if g.page > 0 {
			g.page--
			needsRegenerate = true
		}
		time.Sleep(200 * time.Millisecond)
	}

	if ebiten.IsKeyPressed(ebiten.KeyI) {
		g.showInfo = !g.showInfo
		time.Sleep(200 * time.Millisecond)
	}

	if ebiten.IsKeyPressed(ebiten.KeyR) {
		needsRegenerate = true
		time.Sleep(200 * time.Millisecond)
	}

	if ebiten.IsKeyPressed(ebiten.KeyC) {
		g.generator.ClearCache()
		needsRegenerate = true
		time.Sleep(200 * time.Millisecond)
	}

	// Generate sprites on first frame or when needed
	if len(g.sprites) == 0 || needsRegenerate {
		g.generateSprites()
	}

	return nil
}

// Draw renders the game screen.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{15, 15, 20, 255})

	// Draw sprites in grid
	for i, sprite := range g.sprites {
		if sprite == nil {
			continue
		}

		col := i % gridColumns
		row := i / gridColumns

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(scale), float64(scale))
		op.GeoM.Translate(
			float64(col*(spriteSize*scale+10)+20),
			float64(row*(spriteSize*scale+10)+100),
		)
		screen.DrawImage(sprite, op)
	}

	// Draw UI
	genreName := g.genres[g.genreIndex]
	title := fmt.Sprintf("Genre Gallery: %s (Page %d/%d)", genreName, g.page+1, g.maxPages)
	ebitenutil.DebugPrintAt(screen, title, 10, 10)

	// Get cache stats
	stats := g.generator.Stats()
	statsText := fmt.Sprintf("Cache: %d/%d | Hits: %d | Misses: %d | Hit Rate: %.1f%%",
		stats.Size, stats.Capacity, stats.Hits, stats.Misses, stats.HitRate*100)
	ebitenutil.DebugPrintAt(screen, statsText, 10, 30)

	spriteCount := len(g.sprites)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Sprites: %d", spriteCount), 10, 50)

	// Draw info panel if enabled
	if g.showInfo {
		g.drawInfoPanel(screen)
	}

	// Draw controls
	controlsY := screenHeight - 100
	controls := []string{
		"Controls:",
		"LEFT/RIGHT - Change Genre",
		"UP/DOWN - Change Page",
		"I - Toggle Info",
		"R - Regenerate",
		"C - Clear Cache",
		"ESC - Quit",
	}

	for i, line := range controls {
		ebitenutil.DebugPrintAt(screen, line, 10, controlsY+i*15)
	}
}

// drawInfoPanel renders genre-specific information.
func (g *Game) drawInfoPanel(screen *ebiten.Image) {
	panelX := screenWidth - 320
	panelY := 100
	panelWidth := 300
	panelHeight := 400

	// Draw background
	panel := ebiten.NewImage(panelWidth, panelHeight)
	panel.Fill(color.RGBA{30, 30, 40, 200})

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(panelX), float64(panelY))
	screen.DrawImage(panel, op)

	// Draw genre information
	genreName := g.genres[g.genreIndex]
	y := panelY + 10

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Genre: %s", genreName), panelX+10, y)
	y += 25

	// Genre-specific characteristics
	info := g.getGenreInfo(genreName)
	for _, line := range info {
		ebitenutil.DebugPrintAt(screen, line, panelX+10, y)
		y += 15
	}
}

// getGenreInfo returns descriptive information for a genre.
func (g *Game) getGenreInfo(genreID string) []string {
	switch genreID {
	case "fantasy":
		return []string{
			"Characteristics:",
			"• Medieval aesthetic",
			"• Rounded, organic shapes",
			"• Plate/chain armor patterns",
			"• Magical glows",
			"• Earthy color palettes",
			"",
			"Sprites:",
			"• Knights & warriors",
			"• Dragons & beasts",
			"• Enchanted weapons",
			"• Magic potions",
			"• Castle tiles",
		}

	case "sci-fi":
		return []string{
			"Characteristics:",
			"• Angular, geometric",
			"• Metallic sheens",
			"• Neon accents",
			"• Cool colors (blue/cyan)",
			"• Tech elements",
			"",
			"Sprites:",
			"• Androids & mechs",
			"• Laser weapons",
			"• Energy shields",
			"• Nanobots",
			"• Space station tiles",
		}

	case "horror":
		return []string{
			"Characteristics:",
			"• Distorted proportions",
			"• Organic/irregular",
			"• Dark palettes",
			"• Blood red accents",
			"• Translucent effects",
			"",
			"Sprites:",
			"• Zombies & ghosts",
			"• Tentacle monsters",
			"• Cursed weapons",
			"• Dark rituals",
			"• Corrupted tiles",
		}

	case "cyberpunk":
		return []string{
			"Characteristics:",
			"• High contrast",
			"• Neon vs shadow",
			"• Implant glows",
			"• Urban/industrial",
			"• Pink/purple/cyan",
			"",
			"Sprites:",
			"• Augmented humans",
			"• Cyber weapons",
			"• Hacking tools",
			"• Neon signs",
			"• City tiles",
		}

	case "post-apoc":
		return []string{
			"Characteristics:",
			"• Rough, damaged edges",
			"• Makeshift equipment",
			"• Muted, desaturated",
			"• Rust/decay patterns",
			"• Survival theme",
			"",
			"Sprites:",
			"• Wasteland survivors",
			"• Mutant creatures",
			"• Salvaged weapons",
			"• Scrap armor",
			"• Ruined tiles",
		}

	default:
		return []string{"No information available"}
	}
}

// Layout returns the game's screen dimensions.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	// Parse command-line flags
	seed := flag.Int64("seed", 12345, "Random seed for generation")
	flag.Parse()

	// Initialize logger for test utility
	logger := logging.TestUtilityLogger("genregallery")

	logger.WithField("seed", *seed).Info("starting genre gallery")
	logger.Info("controls:")
	logger.Info("  LEFT/RIGHT - Change Genre")
	logger.Info("  UP/DOWN - Change Page")
	logger.Info("  I - Toggle Info Panel")
	logger.Info("  R - Regenerate Current Page")
	logger.Info("  C - Clear Cache")
	logger.Info("  ESC - Quit")

	// Create and run game
	game := NewGame(*seed, logger)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Venture - Genre Gallery")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(game); err != nil && err.Error() != "quit" {
		logger.WithError(err).Fatal("game error")
	}

	// Print final stats
	stats := game.generator.Stats()
	logger.WithFields(logrus.Fields{
		"cacheSize":     stats.Size,
		"cacheCapacity": stats.Capacity,
		"hits":          stats.Hits,
		"misses":        stats.Misses,
		"hitRate":       stats.HitRate * 100,
		"pagesViewed":   game.page + 1,
		"currentGenre":  game.genres[game.genreIndex],
	}).Info("final statistics")

	fmt.Println("\nFinal Statistics:")
	fmt.Printf("  Cache Size: %d / %d\n", stats.Size, stats.Capacity)
	fmt.Printf("  Cache Hits: %d\n", stats.Hits)
	fmt.Printf("  Cache Misses: %d\n", stats.Misses)
	fmt.Printf("  Hit Rate: %.2f%%\n", stats.HitRate*100)
	fmt.Printf("  Pages Viewed: %d\n", game.page+1)
	fmt.Printf("  Current Genre: %s\n", game.genres[game.genreIndex])
}
