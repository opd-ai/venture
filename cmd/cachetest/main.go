package main

import (
	"flag"
	"fmt"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/sirupsen/logrus"
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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
}

// NewGame creates a new cache test game.
func NewGame(cacheCapacity int) *Game {
	return &Game{
		cachedGen:      sprites.NewCachedGenerator(cacheCapacity),
		configs:        make([]sprites.Config, 0, gridColumns*gridRows),
		lastUpdate:     time.Now(),
		updateInterval: 500 * time.Millisecond,
		paused:         false,
		genres:         []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apoc"},
		genreIndex:     0,
		showStats:      true,
	}
}

// generateRandomConfig creates a random sprite configuration.
func (g *Game) generateRandomConfig() sprites.Config {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	spriteTypes := []sprites.SpriteType{
		sprites.SpriteEntity,
		sprites.SpriteItem,
		sprites.SpriteTile,
		sprites.SpriteParticle,
	}

	return sprites.Config{
		Type:       spriteTypes[rng.Intn(len(spriteTypes))],
		Width:      spriteSize,
		Height:     spriteSize,
		Seed:       rng.Int63(),
		GenreID:    g.genres[g.genreIndex],
		Complexity: 0.3 + rng.Float64()*0.4, // 0.3-0.7
		Variation:  rng.Intn(3),
	}
}

// Update handles game logic updates.
func (g *Game) Update() error {
	// Handle input
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return fmt.Errorf("quit")
	}

	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		if !g.paused {
			g.paused = true
			time.Sleep(200 * time.Millisecond) // Simple debounce
		} else {
			g.paused = false
			time.Sleep(200 * time.Millisecond)
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyC) {
		g.cachedGen.ClearCache()
		g.configs = g.configs[:0]
		g.generatedCount = 0
		g.totalRequests = 0
		time.Sleep(200 * time.Millisecond)
	}

	if ebiten.IsKeyPressed(ebiten.KeyG) {
		g.genreIndex = (g.genreIndex + 1) % len(g.genres)
		time.Sleep(200 * time.Millisecond)
	}

	if ebiten.IsKeyPressed(ebiten.KeyS) {
		g.showStats = !g.showStats
		time.Sleep(200 * time.Millisecond)
	}

	if ebiten.IsKeyPressed(ebiten.KeyEqual) {
		if g.updateInterval > 100*time.Millisecond {
			g.updateInterval -= 100 * time.Millisecond
		}
		time.Sleep(200 * time.Millisecond)
	}

	if ebiten.IsKeyPressed(ebiten.KeyMinus) {
		if g.updateInterval < 2*time.Second {
			g.updateInterval += 100 * time.Millisecond
		}
		time.Sleep(200 * time.Millisecond)
	}

	// Generate sprites periodically (if not paused)
	if !g.paused && time.Since(g.lastUpdate) >= g.updateInterval {
		g.lastUpdate = time.Now()

		// Generate random config or reuse existing (50/50 chance for cache hits)
		var config sprites.Config
		if len(g.configs) > 0 && rand.Intn(2) == 0 {
			// Reuse existing config (should hit cache)
			config = g.configs[rand.Intn(len(g.configs))]
		} else {
			// Generate new config
			config = g.generateRandomConfig()
			g.configs = append(g.configs, config)
			g.generatedCount++
		}

		// Request sprite (will use cache if available)
		_, err := g.cachedGen.Generate(config)
		if err != nil {
			g.logger.WithError(err).Error("failed to generate sprite")
		}

		g.totalRequests++

		// Keep configs list manageable
		if len(g.configs) > 200 {
			g.configs = g.configs[50:]
		}
	}

	return nil
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
	flag.Parse()

	// Create and run game
	game := NewGame(*cacheCapacity)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Sprite Cache Performance Test")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	logger.WithField("capacity", *cacheCapacity).Info("starting cache test")
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
