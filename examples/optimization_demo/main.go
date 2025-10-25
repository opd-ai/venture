package main

import (
	"fmt"
	"image/color"
	"log"
	"runtime"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/venture/pkg/engine"
)

const (
	screenWidth  = 1024
	screenHeight = 768
	worldWidth   = 5000
	worldHeight  = 5000
)

// Game implements ebiten.Game interface
type Game struct {
	// Systems
	cameraSystem     *engine.CameraSystem
	cameraEntity     *engine.Entity
	renderSystem     *engine.EbitenRenderSystem
	spatialPartition *engine.SpatialPartitionSystem

	// Entities
	entities []*engine.Entity

	// Shared sprites for batching efficiency
	sprites []*ebiten.Image

	// Performance tracking
	frameCount    int
	lastFPSUpdate time.Time
	currentFPS    float64
	frameTimeMS   float64

	// Memory tracking
	lastMemStats runtime.MemStats

	// Optimization toggles
	cullingEnabled  bool
	batchingEnabled bool

	// UI state
	showHelp bool
}

// NewGame creates a new optimization demo game
func NewGame() *Game {
	g := &Game{
		cameraSystem:    engine.NewCameraSystem(screenWidth, screenHeight),
		cullingEnabled:  true,
		batchingEnabled: true,
		showHelp:        true,
		lastFPSUpdate:   time.Now(),
		currentFPS:      60,
	}

	// Create camera entity
	g.cameraEntity = engine.NewEntity(9999999)
	cameraComp := engine.NewCameraComponent()
	cameraComp.X = worldWidth/2 - screenWidth/2
	cameraComp.Y = worldHeight/2 - screenHeight/2
	g.cameraEntity.AddComponent(cameraComp)
	g.cameraSystem.SetActiveCamera(g.cameraEntity)

	// Initialize render system
	g.renderSystem = engine.NewRenderSystem(g.cameraSystem)

	// Initialize spatial partition
	g.spatialPartition = engine.NewSpatialPartitionSystem(worldWidth, worldHeight)
	g.renderSystem.SetSpatialPartition(g.spatialPartition)
	g.renderSystem.EnableCulling(g.cullingEnabled)
	g.renderSystem.EnableBatching(g.batchingEnabled)

	// Create shared sprites (10 types for good batching)
	g.sprites = make([]*ebiten.Image, 10)
	for i := 0; i < 10; i++ {
		sprite := ebiten.NewImage(32, 32)
		// Fill with different colors
		hue := float64(i) / 10.0
		r := uint8(255 * hue)
		b := uint8(255 * (1 - hue))
		sprite.Fill(color.RGBA{R: r, G: 128, B: b, A: 255})
		g.sprites[i] = sprite
	}

	// Create 2000 entities spread across the world
	g.createEntities(2000)

	return g
}

// createEntities generates entities spread across the world
func (g *Game) createEntities(count int) {
	g.entities = make([]*engine.Entity, count)

	for i := 0; i < count; i++ {
		entity := engine.NewEntity(uint64(i))

		// Spread entities across world
		pos := &engine.PositionComponent{
			X: float64((i % 100) * 50),
			Y: float64((i / 100) * 50),
		}
		entity.AddComponent(pos)

		// Assign sprite (reuse for batching)
		sprite := &engine.EbitenSprite{
			Image:   g.sprites[i%10],
			Width:   32,
			Height:  32,
			Visible: true,
		}
		entity.AddComponent(sprite)

		g.entities[i] = entity
	}

	// Rebuild spatial partition
	g.spatialPartition.Update(g.entities, 0)
}

// Update handles game logic
func (g *Game) Update() error {
	// Handle input
	g.handleInput()

	// Update FPS counter
	g.frameCount++
	now := time.Now()
	if now.Sub(g.lastFPSUpdate) >= time.Second {
		g.currentFPS = float64(g.frameCount) / now.Sub(g.lastFPSUpdate).Seconds()
		g.frameCount = 0
		g.lastFPSUpdate = now

		// Update memory stats
		runtime.ReadMemStats(&g.lastMemStats)
	}

	// Periodically rebuild spatial partition
	if g.frameCount%60 == 0 {
		g.spatialPartition.Update(g.entities, 0)
	}

	return nil
}

// getCameraComp returns the camera component for manipulation
func (g *Game) getCameraComp() *engine.CameraComponent {
	if comp, ok := g.cameraEntity.GetComponent("camera"); ok {
		return comp.(*engine.CameraComponent)
	}
	return nil
}

// handleInput processes keyboard input
func (g *Game) handleInput() {
	cam := g.getCameraComp()
	if cam == nil {
		return
	}

	// Camera movement (WASD)
	const moveSpeed = 10.0
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		cam.Y -= moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		cam.Y += moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		cam.X -= moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		cam.X += moveSpeed
	}

	// Toggle optimizations
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		g.cullingEnabled = !g.cullingEnabled
		g.renderSystem.EnableCulling(g.cullingEnabled)
		log.Printf("Viewport Culling: %v", g.cullingEnabled)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyB) {
		g.batchingEnabled = !g.batchingEnabled
		g.renderSystem.EnableBatching(g.batchingEnabled)
		log.Printf("Batch Rendering: %v", g.batchingEnabled)
	}

	// Entity count adjustment
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		g.createEntities(500)
		log.Println("Entity count: 500")
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		g.createEntities(1000)
		log.Println("Entity count: 1000")
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		g.createEntities(2000)
		log.Println("Entity count: 2000")
	}
	if inpututil.IsKeyJustPressed(ebiten.Key4) {
		g.createEntities(5000)
		log.Println("Entity count: 5000")
	}

	// Toggle help
	if inpututil.IsKeyJustPressed(ebiten.KeyH) {
		g.showHelp = !g.showHelp
	}

	// Reset camera position
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		cam.X = worldWidth/2 - screenWidth/2
		cam.Y = worldHeight/2 - screenHeight/2
		log.Println("Camera reset to center")
	}
}

// Draw renders the game
func (g *Game) Draw(screen *ebiten.Image) {
	// Clear screen
	screen.Fill(color.RGBA{R: 20, G: 20, B: 30, A: 255})

	// Measure render time
	startTime := time.Now()

	// Render entities
	g.renderSystem.Draw(screen, g.entities)

	// Calculate frame time
	g.frameTimeMS = float64(time.Since(startTime).Microseconds()) / 1000.0

	// Draw UI
	g.drawUI(screen)
}

// drawUI renders the statistics overlay
func (g *Game) drawUI(screen *ebiten.Image) {
	stats := g.renderSystem.GetStats()

	// Calculate percentages
	culledPercent := 0.0
	if stats.TotalEntities > 0 {
		culledPercent = float64(stats.CulledEntities) / float64(stats.TotalEntities) * 100
	}

	// Performance stats
	memMB := float64(g.lastMemStats.HeapAlloc) / (1024 * 1024)

	// Status colors
	cullingColor := "OFF"
	if g.cullingEnabled {
		cullingColor = "ON"
	}
	batchingColor := "OFF"
	if g.batchingEnabled {
		batchingColor = "ON"
	}

	// Build stats text
	statsText := fmt.Sprintf(
		"FPS: %.1f (%.2f ms/frame)\n"+
			"Memory: %.2f MB\n"+
			"Entities: %d total\n"+
			"Rendered: %d (%.1f%%)\n"+
			"Culled: %d (%.1f%%)\n"+
			"Batches: %d\n"+
			"\n"+
			"Optimizations:\n"+
			"  Culling: %s [C]\n"+
			"  Batching: %s [B]\n",
		g.currentFPS,
		g.frameTimeMS,
		memMB,
		stats.TotalEntities,
		stats.RenderedEntities,
		float64(stats.RenderedEntities)/float64(stats.TotalEntities)*100,
		stats.CulledEntities,
		culledPercent,
		stats.BatchCount,
		cullingColor,
		batchingColor,
	)

	// Draw stats with shadow for readability
	ebitenutil.DebugPrintAt(screen, statsText, 11, 11) // Shadow
	ebitenutil.DebugPrintAt(screen, statsText, 10, 10) // Text

	// Draw help panel if enabled
	if g.showHelp {
		g.drawHelp(screen)
	} else {
		helpText := "Press [H] for help"
		ebitenutil.DebugPrintAt(screen, helpText, 11, screenHeight-29)
		ebitenutil.DebugPrintAt(screen, helpText, 10, screenHeight-30)
	}

	// Draw camera position
	cam := g.getCameraComp()
	if cam != nil {
		camText := fmt.Sprintf("Camera: (%.0f, %.0f)", cam.X, cam.Y)
		ebitenutil.DebugPrintAt(screen, camText, screenWidth-191, 11)
		ebitenutil.DebugPrintAt(screen, camText, screenWidth-190, 10)
	}
}

// drawHelp renders the help panel
func (g *Game) drawHelp(screen *ebiten.Image) {
	helpText := "" +
		"=== OPTIMIZATION DEMO CONTROLS ===\n" +
		"\n" +
		"Camera Movement:\n" +
		"  [W][A][S][D] - Move camera\n" +
		"  [R] - Reset to center\n" +
		"\n" +
		"Optimizations:\n" +
		"  [C] - Toggle Viewport Culling\n" +
		"  [B] - Toggle Batch Rendering\n" +
		"\n" +
		"Entity Count:\n" +
		"  [1] - 500 entities\n" +
		"  [2] - 1000 entities\n" +
		"  [3] - 2000 entities\n" +
		"  [4] - 5000 entities\n" +
		"\n" +
		"UI:\n" +
		"  [H] - Toggle this help\n" +
		"\n" +
		"Experiment with toggling culling\n" +
		"on/off while moving the camera to\n" +
		"see the performance difference!"

	x := screenWidth - 340
	y := screenHeight - 380

	// Draw background
	bg := ebiten.NewImage(330, 370)
	bg.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 200})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x-5), float64(y-5))
	screen.DrawImage(bg, op)

	// Draw text
	ebitenutil.DebugPrintAt(screen, helpText, x, y)
}

// Layout returns the game screen size
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	game := NewGame()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Venture - Optimization Demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	log.Println("=== Venture Optimization Demo ===")
	log.Println("Press [H] for controls")
	log.Println("Move camera with [W][A][S][D]")
	log.Println("Toggle culling with [C]")
	log.Println("Toggle batching with [B]")
	log.Println("")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
