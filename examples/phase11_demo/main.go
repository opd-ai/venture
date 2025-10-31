// Phase 11.1 Demonstration: Diagonal Walls & Multi-Layer Terrain
//
// This demo showcases the new tile rendering capabilities:
// - Diagonal walls at 45° angles (NE, NW, SE, SW)
// - Multi-layer terrain (platforms, ramps, pits)
// - 3D visual effects (shadows, gradients, depth)
//
// Usage: go run ./examples/phase11_demo
package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/opd-ai/venture/pkg/rendering/tiles"
)

const (
	screenWidth  = 800
	screenHeight = 600
	tileSize     = 64
	tilesPerRow  = 4
)

// Game represents the demo application state.
type Game struct {
	tileImages       map[tiles.TileType]*ebiten.Image
	currentIdx       int
	tileTypes        []tiles.TileType
	prevRightPressed bool
	prevLeftPressed  bool
}

// NewGame creates a new demo game.
func NewGame() (*Game, error) {
	// Generate all Phase 11.1 tile types
	tileTypes := []tiles.TileType{
		tiles.TileWallNE,
		tiles.TileWallNW,
		tiles.TileWallSE,
		tiles.TileWallSW,
		tiles.TilePlatform,
		tiles.TileRamp,
		tiles.TilePit,
	}

	generator := tiles.NewGenerator()
	tileImages := make(map[tiles.TileType]*ebiten.Image)

	// Generate images for each tile type
	for _, tileType := range tileTypes {
		config := tiles.Config{
			Type:    tileType,
			Width:   tileSize,
			Height:  tileSize,
			GenreID: "fantasy",
			Seed:    12345,
			Variant: 0.5,
		}

		img, err := generator.Generate(config)
		if err != nil {
			return nil, fmt.Errorf("failed to generate tile %s: %w", tileType, err)
		}

		ebitenImg := ebiten.NewImageFromImage(img)
		tileImages[tileType] = ebitenImg
	}

	return &Game{
		tileImages:       tileImages,
		tileTypes:        tileTypes,
		currentIdx:       0,
		prevRightPressed: false,
		prevLeftPressed:  false,
	}, nil
}

// Update updates the game logic.
func (g *Game) Update() error {
	// Cycle through tiles with arrow keys (single-step with debouncing)
	rightPressed := ebiten.IsKeyPressed(ebiten.KeyRight) && ebiten.IsKeyPressed(ebiten.KeyControl)
	leftPressed := ebiten.IsKeyPressed(ebiten.KeyLeft) && ebiten.IsKeyPressed(ebiten.KeyControl)

	// Advance on key press (not while held)
	if rightPressed && !g.prevRightPressed {
		g.currentIdx = (g.currentIdx + 1) % len(g.tileTypes)
	}
	if leftPressed && !g.prevLeftPressed {
		g.currentIdx = (g.currentIdx - 1 + len(g.tileTypes)) % len(g.tileTypes)
	}

	// Track previous frame state for debouncing
	g.prevRightPressed = rightPressed
	g.prevLeftPressed = leftPressed

	return nil
}

// Draw draws the game screen.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{40, 40, 60, 255})

	// Draw title
	msg := "Phase 11.1: Diagonal Walls & Multi-Layer Terrain"
	ebitenutil.DebugPrintAt(screen, msg, 10, 10)

	// Draw instructions
	instructions := "Use CTRL+LEFT/RIGHT arrows to cycle tiles"
	ebitenutil.DebugPrintAt(screen, instructions, 10, 30)

	// Draw all tiles in a grid
	for i, tileType := range g.tileTypes {
		row := i / tilesPerRow
		col := i % tilesPerRow

		x := 50 + col*(tileSize+20)
		y := 80 + row*(tileSize+60)

		// Draw tile
		if img, ok := g.tileImages[tileType]; ok {
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(float64(x), float64(y))

			// Highlight current tile
			if i == g.currentIdx {
				// Draw yellow border
				for t := 0; t < 3; t++ {
					ebitenutil.DrawRect(screen, float64(x-t-1), float64(y-t-1), float64(tileSize+2*(t+1)), 1, color.RGBA{255, 255, 0, 255})
					ebitenutil.DrawRect(screen, float64(x-t-1), float64(y+tileSize+t), float64(tileSize+2*(t+1)), 1, color.RGBA{255, 255, 0, 255})
					ebitenutil.DrawRect(screen, float64(x-t-1), float64(y-t-1), 1, float64(tileSize+2*(t+1)), color.RGBA{255, 255, 0, 255})
					ebitenutil.DrawRect(screen, float64(x+tileSize+t), float64(y-t-1), 1, float64(tileSize+2*(t+1)), color.RGBA{255, 255, 0, 255})
				}
			}

			screen.DrawImage(img, opts)
		}

		// Draw label
		label := tileType.String()
		ebitenutil.DebugPrintAt(screen, label, x, y+tileSize+5)
	}

	// Draw current tile enlarged
	if g.currentIdx >= 0 && g.currentIdx < len(g.tileTypes) {
		currentType := g.tileTypes[g.currentIdx]
		if img, ok := g.tileImages[currentType]; ok {
			opts := &ebiten.DrawImageOptions{}
			// Scale 2.5x and center
			scale := 2.5
			opts.GeoM.Scale(scale, scale)
			centerX := screenWidth/2 - int(float64(tileSize)*scale/2)
			centerY := screenHeight - int(float64(tileSize)*scale) - 80
			opts.GeoM.Translate(float64(centerX), float64(centerY))

			screen.DrawImage(img, opts)

			// Draw description box
			descY := screenHeight - 60
			ebitenutil.DrawRect(screen, 10, float64(descY-5), screenWidth-20, 50, color.RGBA{20, 20, 30, 200})

			desc := getTileDescription(currentType)
			ebitenutil.DebugPrintAt(screen, desc, 20, descY)

			// Draw feature details
			features := getTileFeatures(currentType)
			ebitenutil.DebugPrintAt(screen, features, 20, descY+15)
		}
	}
}

// Layout sets the screen layout.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// getTileDescription returns a description of the tile type.
func getTileDescription(t tiles.TileType) string {
	switch t {
	case tiles.TileWallNE:
		return "Diagonal Wall (NE): 45° diagonal from bottom-left to top-right (/)"
	case tiles.TileWallNW:
		return "Diagonal Wall (NW): 45° diagonal from bottom-right to top-left (\\)"
	case tiles.TileWallSE:
		return "Diagonal Wall (SE): 45° diagonal from top-left to bottom-right (\\)"
	case tiles.TileWallSW:
		return "Diagonal Wall (SW): 45° diagonal from top-right to bottom-left (/)"
	case tiles.TilePlatform:
		return "Platform: Elevated surface with 3D edge effects (highlight/shadow)"
	case tiles.TileRamp:
		return "Ramp: Gradient transition between layers with step lines"
	case tiles.TilePit:
		return "Pit: Dark void with vignette depth effect and edge highlights"
	default:
		return "Unknown tile type"
	}
}

// getTileFeatures returns technical features of the tile type.
func getTileFeatures(t tiles.TileType) string {
	switch t {
	case tiles.TileWallNE, tiles.TileWallNW, tiles.TileWallSE, tiles.TileWallSW:
		return "Features: Triangle fill algorithm, shadow gradients, procedural texture"
	case tiles.TilePlatform:
		return "Features: Raised edges (top/left light, bottom/right dark), solid fill"
	case tiles.TileRamp:
		return "Features: Vertical gradient (dark→light), 4 step lines for depth"
	case tiles.TilePit:
		return "Features: Radial vignette (darker center), edge highlights for depth"
	default:
		return ""
	}
}

func main() {
	game, err := NewGame()
	if err != nil {
		log.Fatalf("Failed to create game: %v", err)
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Phase 11.1 Demo - Diagonal Walls & Multi-Layer Terrain")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
