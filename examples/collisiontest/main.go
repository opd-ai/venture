// collisiontest demonstrates Phase 48 pixel-perfect collision detection.
// Run with: go run ./cmd/collisiontest [-debug-collision]
//
// Features demonstrated:
// - 0.1-pixel precision collision detection
// - Smooth wall sliding with edge normals
// - Collision shapes (AABB, Circle, RoundedRect)
// - Visual/collision alignment verification
package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/venture/pkg/engine"
)

const (
	screenWidth  = 800
	screenHeight = 600
	tileSize     = 32
)

var debugCollision = flag.Bool("debug-collision", true, "Enable collision debug visualization")

// Game implements the Ebiten game interface for collision testing.
type Game struct {
	player     *engine.Entity
	walls      []*engine.Entity
	testObject *engine.Entity

	// Input state
	moveLeft  bool
	moveRight bool
	moveUp    bool
	moveDown  bool

	// Test configuration
	shapeMode int // 0=AABB, 1=Circle, 2=RoundedRect

	// Entity ID counter
	nextEntityID uint64
}

// NewGame creates a new collision test game.
func NewGame() *Game {
	g := &Game{
		shapeMode:    0,
		nextEntityID: 1,
	}

	// Create player entity
	g.player = engine.NewEntity(g.nextEntityID)
	g.nextEntityID++
	g.player.AddComponent(&engine.PositionComponent{X: 400, Y: 300})
	g.player.AddComponent(&engine.VelocityComponent{})
	g.player.AddComponent(&engine.PreciseColliderComponent{
		Width: 24, Height: 24,
		OffsetX: -12, OffsetY: -12,
		Shape: engine.ShapeAABB,
		Solid: true,
	})

	// Create test walls
	g.walls = []*engine.Entity{
		// Top wall
		g.createWall(0, 0, screenWidth, tileSize),
		// Bottom wall
		g.createWall(0, screenHeight-tileSize, screenWidth, tileSize),
		// Left wall
		g.createWall(0, 0, tileSize, screenHeight),
		// Right wall
		g.createWall(screenWidth-tileSize, 0, tileSize, screenHeight),

		// Center obstacles
		g.createWall(200, 200, 64, 64),
		g.createWall(500, 300, 96, 32),
		g.createWall(300, 450, 128, 48),
	}

	// Create rounded test object
	g.testObject = engine.NewEntity(100)
	g.testObject.AddComponent(&engine.PositionComponent{X: 600, Y: 150})
	g.testObject.AddComponent(&engine.PreciseColliderComponent{
		Width: 48, Height: 48,
		OffsetX: -24, OffsetY: -24,
		Shape:        engine.ShapeRoundedRect,
		CornerRadius: 8,
		Solid:        true,
	})

	return g
}

// createWall creates a wall entity at the specified position and size.
func (g *Game) createWall(x, y, width, height float64) *engine.Entity {
	id := g.nextEntityID
	g.nextEntityID++
	wall := engine.NewEntity(id)
	wall.AddComponent(&engine.PositionComponent{X: x + width/2, Y: y + height/2})
	wall.AddComponent(&engine.PreciseColliderComponent{
		Width: width, Height: height,
		OffsetX: -width / 2, OffsetY: -height / 2,
		Shape: engine.ShapeAABB,
		Solid: true,
	})
	return wall
}

// Update updates the game state.
func (g *Game) Update() error {
	// Handle input
	g.handleInput()

	// Update player velocity based on input
	posComp, _ := g.player.GetComponent("position")
	pos := posComp.(*engine.PositionComponent)

	velComp, _ := g.player.GetComponent("velocity")
	vel := velComp.(*engine.VelocityComponent)

	playerSpeed := 3.0
	vel.VX, vel.VY = 0, 0

	if g.moveLeft {
		vel.VX = -playerSpeed
	}
	if g.moveRight {
		vel.VX = playerSpeed
	}
	if g.moveUp {
		vel.VY = -playerSpeed
	}
	if g.moveDown {
		vel.VY = playerSpeed
	}

	// Normalize diagonal movement
	if vel.VX != 0 && vel.VY != 0 {
		factor := 1.0 / math.Sqrt(2)
		vel.VX *= factor
		vel.VY *= factor
	}

	// Calculate new position
	newX := pos.X + vel.VX
	newY := pos.Y + vel.VY

	// Quantize to collision precision
	newX, newY = engine.QuantizePosition(newX, newY)

	// Check wall collisions
	collided := false
	for _, wall := range g.walls {
		if g.checkWallCollision(newX, newY, wall) {
			collided = true

			// Compute wall normal and apply sliding
			wallPosComp, _ := wall.GetComponent("position")
			wallPos := wallPosComp.(*engine.PositionComponent)

			normal := engine.ComputeWallNormal(pos.X, pos.Y, wallPos.X, wallPos.Y)
			vel.VX, vel.VY = engine.ApplyWallSlide(vel.VX, vel.VY, normal)

			// Try sliding along wall
			newX = pos.X + vel.VX
			newY = pos.Y + vel.VY
			newX, newY = engine.QuantizePosition(newX, newY)

			// If still colliding, stop movement
			if g.checkWallCollision(newX, newY, wall) {
				newX, newY = engine.ResolveWallCollision(g.player, normal)
				vel.VX, vel.VY = 0, 0
			}
			break
		}
	}

	// Update position if no collision
	if !collided {
		pos.X, pos.Y = newX, newY
	}

	return nil
}

// checkWallCollision checks if player would collide with wall at new position.
func (g *Game) checkWallCollision(newX, newY float64, wall *engine.Entity) bool {
	playerColliderComp, _ := g.player.GetComponent("precise_collider")
	playerCollider := playerColliderComp.(*engine.PreciseColliderComponent)

	wallColliderComp, _ := wall.GetComponent("precise_collider")
	wallCollider := wallColliderComp.(*engine.PreciseColliderComponent)

	wallPosComp, _ := wall.GetComponent("position")
	wallPos := wallPosComp.(*engine.PositionComponent)

	return playerCollider.Intersects(newX, newY, wallCollider, wallPos.X, wallPos.Y)
}

// handleInput processes keyboard input.
func (g *Game) handleInput() {
	g.moveLeft = ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft)
	g.moveRight = ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight)
	g.moveUp = ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyArrowUp)
	g.moveDown = ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyArrowDown)

	// Cycle shape modes
	if ebiten.IsKeyPressed(ebiten.Key1) {
		g.shapeMode = 0 // AABB
		g.updatePlayerShape()
	}
	if ebiten.IsKeyPressed(ebiten.Key2) {
		g.shapeMode = 1 // Circle
		g.updatePlayerShape()
	}
	if ebiten.IsKeyPressed(ebiten.Key3) {
		g.shapeMode = 2 // RoundedRect
		g.updatePlayerShape()
	}
}

// updatePlayerShape updates the player's collision shape.
func (g *Game) updatePlayerShape() {
	colliderComp, _ := g.player.GetComponent("precise_collider")
	collider := colliderComp.(*engine.PreciseColliderComponent)

	switch g.shapeMode {
	case 0:
		collider.Shape = engine.ShapeAABB
		collider.CornerRadius = 0
	case 1:
		collider.Shape = engine.ShapeCircle
		collider.CornerRadius = 0
	case 2:
		collider.Shape = engine.ShapeRoundedRect
		collider.CornerRadius = 6
	}
}

// Draw renders the game.
func (g *Game) Draw(screen *ebiten.Image) {
	// Clear screen
	screen.Fill(color.RGBA{20, 20, 30, 255})

	// Draw walls
	for _, wall := range g.walls {
		g.drawEntity(screen, wall, color.RGBA{100, 100, 120, 255})
	}

	// Draw test object
	g.drawEntity(screen, g.testObject, color.RGBA{150, 100, 50, 255})

	// Draw player
	g.drawEntity(screen, g.player, color.RGBA{100, 200, 100, 255})

	// Debug visualization
	if *debugCollision {
		g.drawDebugInfo(screen)
	}

	// Instructions
	g.drawInstructions(screen)
}

// drawEntity renders an entity's collision bounds.
func (g *Game) drawEntity(screen *ebiten.Image, entity *engine.Entity, clr color.Color) {
	posComp, hasPos := entity.GetComponent("position")
	if !hasPos {
		return
	}
	pos := posComp.(*engine.PositionComponent)

	colliderComp, hasCollider := entity.GetComponent("precise_collider")
	if !hasCollider {
		return
	}
	collider := colliderComp.(*engine.PreciseColliderComponent)

	minX, minY, maxX, maxY := collider.GetBounds(pos.X, pos.Y)

	// Draw based on shape
	switch collider.Shape {
	case engine.ShapeCircle:
		centerX := (minX + maxX) / 2
		centerY := (minY + maxY) / 2
		radius := (maxX - minX) / 2
		vector.DrawFilledCircle(screen, float32(centerX), float32(centerY), float32(radius), clr, false)

	case engine.ShapeRoundedRect:
		// Draw rounded rectangle (simplified as regular rect for now)
		vector.DrawFilledRect(screen, float32(minX), float32(minY),
			float32(maxX-minX), float32(maxY-minY), clr, false)
		// Draw corner circles
		if collider.CornerRadius > 0 {
			corners := [4][2]float64{
				{minX + collider.CornerRadius, minY + collider.CornerRadius},
				{maxX - collider.CornerRadius, minY + collider.CornerRadius},
				{minX + collider.CornerRadius, maxY - collider.CornerRadius},
				{maxX - collider.CornerRadius, maxY - collider.CornerRadius},
			}
			highlightClr := color.RGBA{
				uint8(clr.(color.RGBA).R + 50),
				uint8(clr.(color.RGBA).G + 50),
				uint8(clr.(color.RGBA).B + 50), 200,
			}
			for _, corner := range corners {
				vector.DrawFilledCircle(screen, float32(corner[0]), float32(corner[1]),
					float32(collider.CornerRadius), highlightClr, false)
			}
		}

	default: // ShapeAABB
		vector.DrawFilledRect(screen, float32(minX), float32(minY),
			float32(maxX-minX), float32(maxY-minY), clr, false)
	}
}

// drawDebugInfo renders collision debug information.
func (g *Game) drawDebugInfo(screen *ebiten.Image) {
	posComp, _ := g.player.GetComponent("position")
	pos := posComp.(*engine.PositionComponent)

	// Draw position grid
	qx, qy := engine.QuantizePosition(pos.X, pos.Y)

	// Draw crosshair at quantized position
	crosshairClr := color.RGBA{255, 255, 0, 200}
	vector.StrokeLine(screen, float32(qx-5), float32(qy), float32(qx+5), float32(qy),
		1, crosshairClr, false)
	vector.StrokeLine(screen, float32(qx), float32(qy-5), float32(qx), float32(qy+5),
		1, crosshairClr, false)

	// Draw collision bounds outline
	colliderComp, _ := g.player.GetComponent("precise_collider")
	collider := colliderComp.(*engine.PreciseColliderComponent)
	minX, minY, maxX, maxY := collider.GetBounds(pos.X, pos.Y)

	outlineClr := color.RGBA{255, 0, 0, 255}
	vector.StrokeRect(screen, float32(minX), float32(minY),
		float32(maxX-minX), float32(maxY-minY), 2, outlineClr, false)

	// Calculate and display alignment error
	alignmentError := engine.GetCollisionAlignment(g.player)

	// Display debug info
	debugY := 10
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Position: (%.1f, %.1f)", pos.X, pos.Y), 10, debugY)
	debugY += 15
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Quantized: (%.1f, %.1f)", qx, qy), 10, debugY)
	debugY += 15
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Alignment Error: %.3f px", alignmentError), 10, debugY)
	debugY += 15

	// Verify Phase 48 requirement
	if alignmentError < 0.5 {
		ebitenutil.DebugPrintAt(screen, "✓ Alignment <0.5px (PASS)", 10, debugY)
	} else {
		ebitenutil.DebugPrintAt(screen, "✗ Alignment ≥0.5px (FAIL)", 10, debugY)
	}
	debugY += 15

	shapeName := "AABB"
	if collider.Shape == engine.ShapeCircle {
		shapeName = "Circle"
	} else if collider.Shape == engine.ShapeRoundedRect {
		shapeName = "RoundedRect"
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Shape: %s", shapeName), 10, debugY)
}

// drawInstructions renders control instructions.
func (g *Game) drawInstructions(screen *ebiten.Image) {
	y := screenHeight - 60
	ebitenutil.DebugPrintAt(screen, "Controls:", 10, y)
	y += 15
	ebitenutil.DebugPrintAt(screen, "WASD/Arrows: Move  |  1: AABB  |  2: Circle  |  3: RoundedRect", 10, y)
	y += 15
	ebitenutil.DebugPrintAt(screen, "Test smooth wall sliding at 0.1px precision", 10, y)
}

// Layout returns the screen dimensions.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	flag.Parse()

	fmt.Println("Phase 48: Pixel-Perfect Collision Test")
	fmt.Println("Features:")
	fmt.Println("  - 0.1-pixel precision collision detection")
	fmt.Println("  - Smooth wall sliding with edge normals")
	fmt.Println("  - Multiple collision shapes (AABB, Circle, RoundedRect)")
	fmt.Println("  - Visual/collision alignment verification")
	fmt.Println()

	game := NewGame()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Phase 48: Pixel-Perfect Collision Demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
