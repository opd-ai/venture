package main

import (
	"flag"
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

const (
	screenWidth  = 800
	screenHeight = 600
	spriteSize   = 32
	gridColumns  = 20
	gridRows     = 15
)

// Game implements the Ebiten game interface for cache testing.
type Game struct {
	cachedGen      *sprites.CachedGenerator
	configs        []sprites.Config
	generatedCount int
	totalRequests  int
	lastUpdate     time.Time
	updateInterval time.Duration
	paused         bool
	genreIndex     int
	genres         []string
	showStats      bool
	logger         *logrus.Logger
	rng            *rand.Rand
}

// NewGame creates a new cache test game.
// seed parameter enables deterministic sprite generation for reproducible test runs.
func NewGame(cacheCapacity int, seed int64) *Game {
	return &Game{
		cachedGen:      sprites.NewCachedGenerator(cacheCapacity),
		configs:        make([]sprites.Config, 0, gridColumns*gridRows),
		lastUpdate:     time.Now(),
		updateInterval: 500 * time.Millisecond,
		paused:         false,
		genres:         []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apoc"},
		genreIndex:     0,
		showStats:      true,
		rng:            rand.New(rand.NewSource(seed)),
	}
}

// generateRandomConfig creates a random sprite configuration.
func (g *Game) generateRandomConfig() sprites.Config {
	spriteTypes := []sprites.SpriteType{
		sprites.SpriteEntity,
		sprites.SpriteItem,
		sprites.SpriteTile,
		sprites.SpriteParticle,
	}

	return sprites.Config{
		Type:       spriteTypes[g.rng.Intn(len(spriteTypes))],
		Width:      spriteSize,
		Height:     spriteSize,
		Seed:       g.rng.Int63(),
		GenreID:    g.genres[g.genreIndex],
		Complexity: 0.3 + g.rng.Float64()*0.4, // 0.3-0.7
		Variation:  g.rng.Intn(3),
	}
}

// Update handles game logic updates.
func (g *Game) Update() error {
	if err := g.handleInput(); err != nil {
		return err
	}

	g.generateSprites()

	return nil
}

// handleInput processes keyboard input and updates game state.
func (g *Game) handleInput() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return fmt.Errorf("quit")
	}

	g.handlePauseKey()
	g.handleClearKey()
	g.handleGenreKey()
	g.handleStatsKey()
	g.handleSpeedKeys()

	return nil
}

// handlePauseKey toggles pause state when space is pressed.
func (g *Game) handlePauseKey() {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.paused = !g.paused
	}
}

// handleClearKey clears cache and resets counters when C is pressed.
func (g *Game) handleClearKey() {
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		g.cachedGen.ClearCache()
		g.configs = g.configs[:0]
		g.generatedCount = 0
		g.totalRequests = 0
	}
}

// handleGenreKey cycles through genres when G is pressed.
func (g *Game) handleGenreKey() {
	if inpututil.IsKeyJustPressed(ebiten.KeyG) {
		g.genreIndex = (g.genreIndex + 1) % len(g.genres)
	}
}

// handleStatsKey toggles stats display when S is pressed.
func (g *Game) handleStatsKey() {
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.showStats = !g.showStats
	}
}

// handleSpeedKeys adjusts update interval with +/- keys.
func (g *Game) handleSpeedKeys() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) {
		if g.updateInterval > 100*time.Millisecond {
			g.updateInterval -= 100 * time.Millisecond
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) {
		if g.updateInterval < 2*time.Second {
			g.updateInterval += 100 * time.Millisecond
		}
	}
}

// generateSprites periodically generates sprites when not paused.
func (g *Game) generateSprites() {
	if g.paused || time.Since(g.lastUpdate) < g.updateInterval {
		return
	}

	g.lastUpdate = time.Now()

	config := g.selectSpriteConfig()
	if _, err := g.cachedGen.Generate(config); err != nil {
		g.logger.WithError(err).Error("failed to generate sprite")
	}

	g.totalRequests++

	if len(g.configs) > 200 {
		g.configs = g.configs[50:]
	}
}

// selectSpriteConfig chooses a random or cached sprite config.
func (g *Game) selectSpriteConfig() sprites.Config {
	if len(g.configs) > 0 && g.rng.Intn(2) == 0 {
		return g.configs[g.rng.Intn(len(g.configs))]
	}

	config := g.generateRandomConfig()
	g.configs = append(g.configs, config)
	g.generatedCount++
	return config
}

// Draw renders the game screen.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 30, 255})

	// Draw title
	titleMsg := "Sprite Cache Performance Test"
	ebitenutil.DebugPrintAt(screen, titleMsg, 10, 10)

	// Get cache stats
	stats := g.cachedGen.Stats()

	// Draw cache stats
	if g.showStats {
		y := 40
		statsLines := []string{
			fmt.Sprintf("Cache Size: %d / %d", stats.Size, stats.Capacity),
			fmt.Sprintf("Cache Hits: %d", stats.Hits),
			fmt.Sprintf("Cache Misses: %d", stats.Misses),
			fmt.Sprintf("Hit Rate: %.2f%%", stats.HitRate*100),
			fmt.Sprintf("Total Requests: %d", g.totalRequests),
			fmt.Sprintf("Unique Configs: %d", g.generatedCount),
			fmt.Sprintf("Genre: %s", g.genres[g.genreIndex]),
			fmt.Sprintf("Update Interval: %dms", g.updateInterval.Milliseconds()),
		}

		for _, line := range statsLines {
			ebitenutil.DebugPrintAt(screen, line, 10, y)
			y += 20
		}

		// Draw status
		status := "Running"
		if g.paused {
			status = "PAUSED"
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Status: %s", status), 10, y+10)
	}

	// Draw performance indicator
	hitRate := stats.HitRate
	var perfColor color.RGBA
	var perfText string

	if hitRate >= 0.75 {
		perfColor = color.RGBA{0, 255, 0, 255} // Green - Excellent
		perfText = "EXCELLENT"
	} else if hitRate >= 0.50 {
		perfColor = color.RGBA{200, 255, 0, 255} // Yellow-green - Good
		perfText = "GOOD"
	} else if hitRate >= 0.25 {
		perfColor = color.RGBA{255, 165, 0, 255} // Orange - Fair
		perfText = "FAIR"
	} else {
		perfColor = color.RGBA{255, 0, 0, 255} // Red - Poor
		perfText = "POOR"
	}

	// Draw performance box
	if g.showStats {
		perfBoxX := 10
		perfBoxY := 280
		perfBoxWidth := 150
		perfBoxHeight := 40

		// Draw background
		perfBox := ebiten.NewImage(perfBoxWidth, perfBoxHeight)
		perfBox.Fill(perfColor)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(perfBoxX), float64(perfBoxY))
		screen.DrawImage(perfBox, op)

		// Draw text
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Performance: %s", perfText),
			perfBoxX+10, perfBoxY+12)
	}

	// Draw controls
	controlsY := screenHeight - 140
	controls := []string{
		"Controls:",
		"SPACE - Pause/Resume",
		"C - Clear Cache",
		"G - Change Genre",
		"S - Toggle Stats",
		"+/- - Adjust Speed",
		"ESC - Quit",
	}

	for i, line := range controls {
		ebitenutil.DebugPrintAt(screen, line, 10, controlsY+i*20)
	}
}

// Layout returns the game's screen dimensions.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	// Parse command-line flags
	cacheCapacity := flag.Int("capacity", 100, "Cache capacity (number of sprites)")
	seed := flag.Int64("seed", 12345, "Random seed for deterministic generation (use different values for variety)")
	flag.Parse()

	// Create and run game
	game := NewGame(*cacheCapacity, *seed)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Sprite Cache Performance Test")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	logger.WithFields(logrus.Fields{
		"capacity": *cacheCapacity,
		"seed":     *seed,
	}).Info("starting cache test")
	logger.Info("controls:")
	logger.Info("  SPACE - Pause/Resume")
	logger.Info("  C - Clear Cache")
	logger.Info("  G - Change Genre")
	logger.Info("  S - Toggle Stats")
	logger.Info("  +/- - Adjust Generation Speed")
	logger.Info("  ESC - Quit")

	if err := ebiten.RunGame(game); err != nil && err.Error() != "quit" {
		logger.WithError(err).Fatal("error")
	}

	// Print final stats
	stats := game.cachedGen.Stats()
	fmt.Println("\nFinal Cache Statistics:")
	fmt.Printf("  Cache Size: %d / %d\n", stats.Size, stats.Capacity)
	fmt.Printf("  Cache Hits: %d\n", stats.Hits)
	fmt.Printf("  Cache Misses: %d\n", stats.Misses)
	fmt.Printf("  Hit Rate: %.2f%%\n", stats.HitRate*100)
	fmt.Printf("  Total Requests: %d\n", game.totalRequests)
	fmt.Printf("  Unique Configs: %d\n", game.generatedCount)
}
