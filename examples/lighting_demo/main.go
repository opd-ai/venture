// Package main demonstrates the dynamic lighting system.
// This example shows how to create and use different types of lights,
// configure genre-specific lighting, and integrate with the game loop.
//
// Usage:
//
//	go run ./examples/lighting_demo
//	go run ./examples/lighting_demo -genre horror
//	go run ./examples/lighting_demo -no-lighting
package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/sirupsen/logrus"
)

const (
	screenWidth  = 800
	screenHeight = 600
	moveSpeed    = 200.0
)

var (
	genreFlag    = flag.String("genre", "fantasy", "Genre preset (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic)")
	lightingFlag = flag.Bool("no-lighting", false, "Disable lighting system")
)

// Game implements ebiten.Game interface with lighting demonstration.
type Game struct {
	world          *engine.World
	lightingSystem *engine.LightingSystem
	playerID       uint64

	// For rendering
	sceneBuffer *ebiten.Image

	// Demo controls
	paused bool
}

// NewGame creates a new lighting demo game.
func NewGame() *Game {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	world := engine.NewWorld()

	// Configure lighting
	config := engine.NewLightingConfig()
	config.SetGenrePreset(*genreFlag)
	config.Enabled = !*lightingFlag

	lightingSystem := engine.NewLightingSystemWithLogger(world, config, logger)
	lightingSystem.SetViewport(0, 0, screenWidth, screenHeight)

	logger.WithFields(logrus.Fields{
		"genre":   *genreFlag,
		"enabled": config.Enabled,
		"ambient": config.AmbientIntensity,
	}).Info("lighting system initialized")

	game := &Game{
		world:          world,
		lightingSystem: lightingSystem,
		sceneBuffer:    ebiten.NewImage(screenWidth, screenHeight),
	}

	game.setupScene()
	return game
}

// setupScene creates entities with various light types.
func (g *Game) setupScene() {
	// Create player with torch
	player := g.world.CreateEntity()
	player.AddComponent(&engine.PositionComponent{X: 400, Y: 300})
	player.AddComponent(&engine.VelocityComponent{VX: 0, VY: 0})

	// Player torch (follows player)
	playerTorch := engine.NewTorchLight(200)
	player.AddComponent(playerTorch)

	// Visual representation
	playerSprite := &engine.EbitenSprite{
		Image:   ebiten.NewImage(32, 32),
		Width:   32,
		Height:  32,
		Visible: true,
		Layer:   10,
	}
	playerSprite.Image.Fill(color.RGBA{100, 150, 255, 255}) // Blue player
	player.AddComponent(playerSprite)

	g.playerID = player.ID

	// Create stationary torches (flickering)
	positions := [][2]float64{
		{150, 150},
		{650, 150},
		{150, 450},
		{650, 450},
	}

	for i, pos := range positions {
		torch := g.world.CreateEntity()
		torch.AddComponent(&engine.PositionComponent{X: pos[0], Y: pos[1]})

		torchLight := engine.NewTorchLight(180)
		// Vary the flicker for each torch
		torchLight.FlickerSpeed = 2.5 + float64(i)*0.3
		torchLight.FlickerAmount = 0.12 + float64(i)*0.02
		torch.AddComponent(torchLight)

		// Visual representation
		torchSprite := &engine.EbitenSprite{
			Image:   ebiten.NewImage(16, 16),
			Width:   16,
			Height:  16,
			Visible: true,
			Layer:   5,
		}
		torchSprite.Image.Fill(color.RGBA{255, 180, 100, 255}) // Orange
		torch.AddComponent(torchSprite)
	}

	// Create magical crystals (pulsing)
	crystalPositions := [][2]float64{
		{400, 150}, {250, 300}, {550, 300}, {400, 450},
	}

	crystalColors := []color.RGBA{
		{255, 0, 255, 255}, // Magenta
		{0, 255, 255, 255}, // Cyan
		{255, 255, 0, 255}, // Yellow
		{0, 255, 0, 255},   // Green
	}

	for i, pos := range crystalPositions {
		crystal := g.world.CreateEntity()
		crystal.AddComponent(&engine.PositionComponent{X: pos[0], Y: pos[1]})

		crystalLight := engine.NewCrystalLight(120, crystalColors[i])
		// Vary pulse speed
		crystalLight.PulseSpeed = 0.3 + float64(i)*0.1
		crystal.AddComponent(crystalLight)

		// Visual representation
		crystalSprite := &engine.EbitenSprite{
			Image:   ebiten.NewImage(12, 12),
			Width:   12,
			Height:  12,
			Visible: true,
			Layer:   5,
		}
		crystalSprite.Image.Fill(crystalColors[i])
		crystal.AddComponent(crystalSprite)
	}

	// Create moving spell (demonstrates light following entity)
	spell := g.world.CreateEntity()
	spell.AddComponent(&engine.PositionComponent{X: 100, Y: 300})
	spell.AddComponent(&engine.VelocityComponent{VX: 100, VY: 50})

	spellLight := engine.NewSpellLight(90, color.RGBA{255, 100, 200, 255}) // Pink
	spell.AddComponent(spellLight)

	spellSprite := &engine.EbitenSprite{
		Image:   ebiten.NewImage(20, 20),
		Width:   20,
		Height:  20,
		Visible: true,
		Layer:   8,
	}
	spellSprite.Image.Fill(color.RGBA{255, 100, 200, 255})
	spell.AddComponent(spellSprite)

	// Create ambient light entity
	ambientEntity := g.world.CreateEntity()
	ambient := engine.NewAmbientLightComponent(
		g.lightingSystem.GetConfig().AmbientColor,
		g.lightingSystem.GetConfig().AmbientIntensity,
	)
	ambientEntity.AddComponent(ambient)

	log.Printf("Scene setup complete: player=%d, entities=%d, lights=%d",
		g.playerID, len(g.world.GetEntities()), 8) // 1 player + 4 torches + 4 crystals - 1 for ambient
}

// Update updates game logic.
func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return fmt.Errorf("quit")
	}

	// Toggle pause
	if ebiten.IsKeyPressed(ebiten.KeyP) && ebiten.IsKeyPressed(ebiten.KeyControl) {
		g.paused = !g.paused
	}

	// Toggle lighting
	if ebiten.IsKeyPressed(ebiten.KeyL) && ebiten.IsKeyPressed(ebiten.KeyControl) {
		g.lightingSystem.SetEnabled(!g.lightingSystem.IsEnabled())
		log.Printf("Lighting: %v", g.lightingSystem.IsEnabled())
	}

	if g.paused {
		return nil
	}

	deltaTime := 1.0 / 60.0

	// Update player movement
	player, ok := g.world.GetEntity(g.playerID)
	if ok {
		vel, _ := player.GetComponent("velocity")
		velComp := vel.(*engine.VelocityComponent)

		velComp.VX, velComp.VY = 0, 0

		if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
			velComp.VY = -moveSpeed
		}
		if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
			velComp.VY = moveSpeed
		}
		if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
			velComp.VX = -moveSpeed
		}
		if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
			velComp.VX = moveSpeed
		}
	}

	// Update entity positions
	entities := g.world.GetEntities()
	for _, entity := range entities {
		posComp, hasPos := entity.GetComponent("position")
		velComp, hasVel := entity.GetComponent("velocity")

		if hasPos && hasVel {
			pos := posComp.(*engine.PositionComponent)
			vel := velComp.(*engine.VelocityComponent)

			pos.X += vel.VX * deltaTime
			pos.Y += vel.VY * deltaTime

			// Bounce spell off edges
			if entity.ID != g.playerID {
				if pos.X < 0 || pos.X > screenWidth {
					vel.VX = -vel.VX
					pos.X = clamp(pos.X, 0, screenWidth)
				}
				if pos.Y < 0 || pos.Y > screenHeight {
					vel.VY = -vel.VY
					pos.Y = clamp(pos.Y, 0, screenHeight)
				}
			} else {
				// Keep player in bounds
				pos.X = clamp(pos.X, 0, screenWidth)
				pos.Y = clamp(pos.Y, 0, screenHeight)
			}
		}
	}

	// Update lighting system
	g.lightingSystem.Update(entities, deltaTime)

	return nil
}

// Draw renders the game.
func (g *Game) Draw(screen *ebiten.Image) {
	// Clear buffers
	g.sceneBuffer.Clear()
	screen.Fill(color.RGBA{20, 20, 30, 255}) // Dark background

	// Render entities to scene buffer
	entities := g.world.GetEntities()
	for _, entity := range entities {
		spriteComp, hasSprite := entity.GetComponent("sprite")
		posComp, hasPos := entity.GetComponent("position")

		if hasSprite && hasPos {
			sprite := spriteComp.(*engine.EbitenSprite)
			pos := posComp.(*engine.PositionComponent)

			if sprite.Visible {
				opts := &ebiten.DrawImageOptions{}
				opts.GeoM.Translate(pos.X-float64(sprite.Width)/2, pos.Y-float64(sprite.Height)/2)
				g.sceneBuffer.DrawImage(sprite.Image, opts)
			}
		}
	}

	// Apply lighting (post-processing)
	g.lightingSystem.ApplyLighting(screen, g.sceneBuffer, entities)

	// Draw UI
	g.drawUI(screen)
}

// drawUI renders UI elements.
func (g *Game) drawUI(screen *ebiten.Image) {
	player, ok := g.world.GetEntity(g.playerID)
	if !ok {
		return
	}

	posComp, _ := player.GetComponent("position")
	pos := posComp.(*engine.PositionComponent)

	// Calculate light intensity at player position
	entities := g.world.GetEntities()
	intensity := g.lightingSystem.CalculateLightIntensityAt(pos.X, pos.Y, entities)

	// Collect visible lights
	visibleLights := g.lightingSystem.CollectVisibleLights(entities)

	// Display info
	info := fmt.Sprintf(
		"Lighting Demo - Genre: %s\n"+
			"Player: (%.0f, %.0f)\n"+
			"Light Intensity: %.2f\n"+
			"Visible Lights: %d/%d\n"+
			"Lighting: %v\n\n"+
			"Controls:\n"+
			"WASD/Arrows - Move\n"+
			"Ctrl+L - Toggle Lighting\n"+
			"Ctrl+P - Pause\n"+
			"ESC - Quit",
		*genreFlag,
		pos.X, pos.Y,
		intensity,
		len(visibleLights),
		g.lightingSystem.GetConfig().MaxLights,
		g.lightingSystem.IsEnabled(),
	)

	ebitenutil.DebugPrint(screen, info)
}

// Layout returns the screen size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func clamp(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func main() {
	flag.Parse()
	rand.Seed(42) // Fixed seed for consistent demo

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Venture - Dynamic Lighting Demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()

	log.Printf("Starting lighting demo with genre: %s", *genreFlag)
	if err := ebiten.RunGame(game); err != nil {
		if err.Error() != "quit" {
			log.Fatal(err)
		}
	}
}
