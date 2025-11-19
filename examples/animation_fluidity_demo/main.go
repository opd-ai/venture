// Animation Fluidity Demo - Phase 46
//
// This example demonstrates the new 8-frame animation system with:
//   - 8-directional movement (N, NE, E, SE, S, SW, W, NW)
//   - Body part articulation (arms ±3px, legs ±4px)
//   - Animation caching with LRU eviction
//   - Pre-computation of common sequences
//
// Usage:
//
//	go run ./examples/animation_fluidity_demo
//
// Controls:
//
//	Arrow keys: Move character (8 directions with diagonals)
//	Space: Attack animation
//	C: Cast animation
//	J: Jump animation
//	H: Hit animation
//	D: Death animation
//	1-8: Test individual animation states
//	P: Toggle performance metrics display
//	ESC: Exit
package main

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/venture/pkg/rendering/animation"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

const (
	screenWidth  = 800
	screenHeight = 600
	spriteSize   = 64
)

type Game struct {
	controller *animation.Controller

	// Current animation state
	currentState string
	currentDir   animation.Direction8
	frameIndex   int
	frameTime    float64
	timeAccum    float64

	// Character position
	x, y   float64
	vx, vy float64

	// Sprite config
	spriteConfig sprites.Config

	// Performance tracking
	showMetrics bool
	lastUpdate  time.Time
}

func NewGame() *Game {
	gen := sprites.NewGenerator()
	controller := animation.NewController(gen)

	// Pre-compute common animations for faster runtime
	seeds := []int64{12345} // Player seed
	controller.PrecomputeCommon(seeds, sprites.Config{
		Width:  spriteSize,
		Height: spriteSize,
		Seed:   12345,
	})

	return &Game{
		controller:   controller,
		currentState: "idle",
		currentDir:   animation.Dir8South,
		x:            float64(screenWidth) / 2,
		y:            float64(screenHeight) / 2,
		spriteConfig: sprites.Config{
			Type:       sprites.SpriteEntity,
			Width:      spriteSize,
			Height:     spriteSize,
			Seed:       12345,
			Complexity: 0.7,
			GenreID:    "fantasy",
			Custom:     make(map[string]interface{}),
		},
		showMetrics: true,
		lastUpdate:  time.Now(),
	}
}

// Update processes input and updates game state each frame.
func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.showMetrics = !g.showMetrics
	}

	g.handleAnimationInput()
	g.handleMovementInput()
	g.constrainPositionToScreen()
	g.updateDirectionAndState()
	g.updateAnimationFrame()

	return nil
}

// handleAnimationInput processes key inputs for animation state changes.
func (g *Game) handleAnimationInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.setAnimationState("attack")
	} else if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		g.setAnimationState("cast")
	} else if inpututil.IsKeyJustPressed(ebiten.KeyJ) {
		g.setAnimationState("jump")
	} else if inpututil.IsKeyJustPressed(ebiten.KeyH) {
		g.setAnimationState("hit")
	} else if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		g.setAnimationState("death")
	}

	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		g.currentState = "idle"
	} else if inpututil.IsKeyJustPressed(ebiten.Key2) {
		g.currentState = "walk"
	} else if inpututil.IsKeyJustPressed(ebiten.Key3) {
		g.currentState = "run"
	}
}

// setAnimationState changes the current animation state and resets the frame index.
func (g *Game) setAnimationState(state string) {
	g.currentState = state
	g.frameIndex = 0
}

// handleMovementInput processes arrow key inputs and updates velocity.
func (g *Game) handleMovementInput() {
	g.vx, g.vy = 0, 0
	speed := 2.0

	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.vy = -speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.vy = speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.vx = -speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.vx = speed
	}

	g.x += g.vx
	g.y += g.vy
}

// constrainPositionToScreen ensures the character stays within screen bounds.
func (g *Game) constrainPositionToScreen() {
	if g.x < spriteSize/2 {
		g.x = spriteSize / 2
	}
	if g.x > screenWidth-spriteSize/2 {
		g.x = screenWidth - spriteSize/2
	}
	if g.y < spriteSize/2 {
		g.y = spriteSize / 2
	}
	if g.y > screenHeight-spriteSize/2 {
		g.y = screenHeight - spriteSize/2
	}
}

// updateDirectionAndState updates the character's direction and animation state based on movement.
func (g *Game) updateDirectionAndState() {
	if g.vx != 0 || g.vy != 0 {
		g.currentDir = animation.CalculateDirection8(g.vx, g.vy)
		if g.currentState == "idle" {
			g.currentState = "walk"
		}
	} else {
		if g.currentState == "walk" || g.currentState == "run" {
			g.currentState = "idle"
		}
	}
}

// updateAnimationFrame advances the animation frame based on elapsed time.
func (g *Game) updateAnimationFrame() {
	deltaTime := time.Since(g.lastUpdate).Seconds()
	g.lastUpdate = time.Now()

	g.frameTime = animation.GetFrameTime(g.currentState)
	g.timeAccum += deltaTime

	if g.timeAccum >= g.frameTime {
		g.timeAccum -= g.frameTime
		frameCount := animation.GetFrameCount(g.currentState)
		g.frameIndex = (g.frameIndex + 1) % frameCount
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{40, 40, 50, 255})

	// Generate and draw current animation frame
	frameCount := animation.GetFrameCount(g.currentState)
	frame, err := g.controller.GenerateFrame(
		g.spriteConfig.Seed,
		g.currentState,
		g.frameIndex,
		frameCount,
		g.currentDir,
		g.spriteConfig,
	)

	if err == nil && frame != nil {
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(g.x-float64(spriteSize)/2, g.y-float64(spriteSize)/2)
		screen.DrawImage(frame, opts)
	}

	// Draw UI
	g.drawUI(screen)
}

func (g *Game) drawUI(screen *ebiten.Image) {
	// Instructions
	instructions := []string{
		"Animation Fluidity Demo - Phase 46",
		"",
		"Controls:",
		"  Arrow Keys: Move (8 directions)",
		"  Space: Attack  C: Cast  J: Jump",
		"  H: Hit  D: Death",
		"  1: Idle  2: Walk  3: Run",
		"  P: Toggle metrics",
		"  ESC: Exit",
	}

	y := 10
	for _, line := range instructions {
		ebitenutil.DebugPrintAt(screen, line, 10, y)
		y += 15
	}

	// Current state info
	y = screenHeight - 120
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("State: %s", g.currentState), 10, y)
	y += 15
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Direction: %s", g.currentDir.String()), 10, y)
	y += 15
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Frame: %d/%d", g.frameIndex+1, animation.GetFrameCount(g.currentState)), 10, y)

	// Performance metrics
	if g.showMetrics {
		metrics := g.controller.GetPerformanceMetrics()

		y = 10
		x := screenWidth - 300
		ebitenutil.DebugPrintAt(screen, "Performance Metrics:", x, y)
		y += 15
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Cache Hit Rate: %.1f%%", metrics.CacheHitRate), x, y)
		y += 15
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Cache Size: %d bytes", metrics.CacheSize), x, y)
		y += 15
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Cache Count: %d entries", metrics.CacheCount), x, y)
		y += 15
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Gen Time: %v", metrics.FrameGenerationTime), x, y)
		y += 15
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  FPS: %.1f", ebiten.ActualFPS()), x, y)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Animation Fluidity Demo - Phase 46")

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
