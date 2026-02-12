// Package main demonstrates mouse delta tracking API (Gap #8 fix).
// This example shows how to use GetMouseDelta() for camera control and aiming.
package main

import (
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/opd-ai/venture/pkg/engine"
)

// Game demonstrates mouse delta usage for camera control.
type Game struct {
	world  *engine.World
	player *engine.Entity
	system *engine.InputSystem

	// Camera state
	cameraX        float64
	cameraY        float64
	cameraRotation float64

	// Settings
	mouseSensitivity float64
}

// NewGame creates a new demo game.
func NewGame() *Game {
	world := engine.NewWorld()
	system := engine.NewInputSystem()

	// Create player entity with input component
	player := world.CreateEntity()
	input := &engine.EbitenInput{}
	player.AddComponent(input)

	return &Game{
		world:            world,
		player:           player,
		system:           system,
		mouseSensitivity: 0.3, // Camera sensitivity
	}
}

// Update updates the game state.
func (g *Game) Update() error {
	entities := []*engine.Entity{g.player}

	// Update input system to track mouse delta
	g.system.Update(entities, 1.0/60.0)

	// Get input component
	inputComp, ok := g.player.GetComponent("input")
	if !ok {
		return nil
	}

	input, ok := inputComp.(*engine.EbitenInput)
	if !ok {
		return nil
	}

	// Use mouse delta for camera control
	dx, dy := input.GetMouseDelta()

	// Update camera rotation based on mouse movement
	g.cameraRotation += float64(dx) * g.mouseSensitivity
	cameraTilt := float64(dy) * g.mouseSensitivity

	// Use WASD for camera position
	moveX, moveY := input.GetMovement()
	g.cameraX += moveX * 5.0
	g.cameraY += moveY * 5.0

	// Print debug info every 60 frames (1 second)
	if ebiten.TPS() > 0 && ebiten.CurrentTPS() > 0 {
		frameCount := int(ebiten.CurrentTPS())
		if frameCount%60 == 0 {
			fmt.Printf("Mouse Delta: (%3d, %3d) | Camera Rotation: %.1f° | Tilt: %.1f | Position: (%.0f, %.0f)\n",
				dx, dy, g.cameraRotation, cameraTilt, g.cameraX, g.cameraY)
		}
	}

	return nil
}

// Draw renders the game.
func (g *Game) Draw(screen *ebiten.Image) {
	// Get current mouse delta for display
	inputComp, ok := g.player.GetComponent("input")
	if !ok {
		return
	}

	input, ok := inputComp.(*engine.EbitenInput)
	if !ok {
		return
	}

	dx, dy := input.GetMouseDelta()
	mouseX, mouseY := input.GetMousePosition()

	// Display instructions and current state
	msg := fmt.Sprintf(
		"Mouse Delta Demo (Gap #8 Fix)\n\n"+
			"Move Mouse: Control camera rotation\n"+
			"WASD: Move camera position\n"+
			"ESC: Exit\n\n"+
			"Mouse Position: (%d, %d)\n"+
			"Mouse Delta: (%d, %d)\n"+
			"Camera Rotation: %.1f°\n"+
			"Camera Position: (%.0f, %.0f)\n"+
			"Mouse Sensitivity: %.1f\n\n"+
			"Usage Examples:\n"+
			"- Camera Control: rotation += dx * sensitivity\n"+
			"- Aiming: targetAngle = atan2(dy, dx)\n"+
			"- Drag & Drop: itemX += dx, itemY += dy",
		mouseX, mouseY,
		dx, dy,
		g.cameraRotation,
		g.cameraX, g.cameraY,
		g.mouseSensitivity,
	)

	ebitenutil.DebugPrint(screen, msg)

	// Visual indicator of mouse delta (arrow showing direction)
	if dx != 0 || dy != 0 {
		arrowMsg := fmt.Sprintf("\n\n\n\n\n\n\n\n\n\n\nMouse Delta Arrow: ")
		if dx > 0 {
			arrowMsg += "→"
		} else if dx < 0 {
			arrowMsg += "←"
		}
		if dy > 0 {
			arrowMsg += "↓"
		} else if dy < 0 {
			arrowMsg += "↑"
		}
		ebitenutil.DebugPrint(screen, arrowMsg)
	}
}

// Layout returns the game's screen dimensions.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 800, 600
}

func main() {
	game := NewGame()

	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("Mouse Delta Demo - Gap #8 Fix")

	fmt.Println("=== Mouse Delta API Demo ===")
	fmt.Println("This example demonstrates the GetMouseDelta() API for camera control.")
	fmt.Println()
	fmt.Println("Gap #8 Fix: Expose mouse delta tracking in InputProvider interface")
	fmt.Println()
	fmt.Println("API Usage:")
	fmt.Println("  input := entity.GetComponent(\"input\").(*engine.EbitenInput)")
	fmt.Println("  dx, dy := input.GetMouseDelta()")
	fmt.Println()
	fmt.Println("Common Use Cases:")
	fmt.Println("  1. First-Person Camera Control")
	fmt.Println("     cameraYaw += float64(dx) * sensitivity")
	fmt.Println("     cameraPitch += float64(dy) * sensitivity")
	fmt.Println()
	fmt.Println("  2. Aiming Assist")
	fmt.Println("     aimDelta := math.Sqrt(float64(dx*dx + dy*dy))")
	fmt.Println("     if aimDelta > threshold { /* apply smoothing */ }")
	fmt.Println()
	fmt.Println("  3. Drag and Drop")
	fmt.Println("     if mousePressed {")
	fmt.Println("       itemX += float64(dx)")
	fmt.Println("       itemY += float64(dy)")
	fmt.Println("     }")
	fmt.Println()
	fmt.Println("Move your mouse to see delta values in action!")
	fmt.Println()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
