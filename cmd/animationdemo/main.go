package main

import (
	"flag"
	"fmt"
	"image/color"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

const (
	screenWidth  = 800
	screenHeight = 600
)

// Game implements ebiten.Game for animation demonstration.
type Game struct {
	world            *engine.World
	player           *engine.Entity
	enemy            *engine.Entity
	spriteGen        *sprites.Generator
	animationSys     *engine.AnimationSystem
	combatSys        *engine.CombatSystem
	movementSys      *engine.MovementSystem
	camera           *engine.CameraSystem
	lastUpdate       time.Time
	mode             int // 0=idle, 1=walk, 2=attack demo
	demoTimer        float64
	showInstructions bool
	logger           *logrus.Logger
}

// NewGame creates a new animation demo game.
func NewGame(seed int64, genreID string) *Game {
	world := engine.NewWorld()
	spriteGen := sprites.NewGenerator()
	animationSys := engine.NewAnimationSystem(spriteGen)
	combatSys := engine.NewCombatSystem(seed)
	movementSys := engine.NewMovementSystem(200.0) // Max speed 200 pixels/sec
	camera := engine.NewCameraSystem(screenWidth, screenHeight)

	// Create player entity
	player := world.CreateEntity()

	// Add components
	player.AddComponent(&engine.PositionComponent{X: 200, Y: 300})
	player.AddComponent(&engine.VelocityComponent{VX: 0, VY: 0})
	player.AddComponent(&engine.EbitenSprite{Width: 32, Height: 32})
	player.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
	player.AddComponent(&engine.AttackComponent{
		Damage:   20,
		Range:    50,
		Cooldown: 0.5,
	})
	player.AddComponent(&engine.FrictionComponent{Coefficient: 0.15})

	// Add animation component
	playerAnim := &engine.AnimationComponent{
		Seed:         seed,
		CurrentState: engine.AnimationStateIdle,
		FrameTime:    0.1, // 100ms per frame
		Loop:         true,
		Playing:      true,
		Dirty:        true, // Generate frames on first update
	}
	player.AddComponent(playerAnim)

	// Create enemy entity
	enemy := world.CreateEntity()
	enemy.AddComponent(&engine.PositionComponent{X: 500, Y: 300})
	enemy.AddComponent(&engine.VelocityComponent{VX: 0, VY: 0})
	enemy.AddComponent(&engine.EbitenSprite{Width: 32, Height: 32})
	enemy.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
	enemy.AddComponent(&engine.AttackComponent{
		Damage:   15,
		Range:    50,
		Cooldown: 0.5,
	})

	// Add enemy animation
	enemyAnim := &engine.AnimationComponent{
		Seed:         seed + 1000,
		CurrentState: engine.AnimationStateIdle,
		FrameTime:    0.1,
		Loop:         true,
		Playing:      true,
		Dirty:        true,
	}
	enemy.AddComponent(enemyAnim)

	return &Game{
		world:            world,
		player:           player,
		enemy:            enemy,
		spriteGen:        spriteGen,
		animationSys:     animationSys,
		combatSys:        combatSys,
		movementSys:      movementSys,
		camera:           camera,
		lastUpdate:       time.Now(),
		mode:             0,
		demoTimer:        0,
		showInstructions: true,
	}
}

// Update updates the game state.
func (g *Game) Update() error {
	// Calculate delta time
	now := time.Now()
	dt := now.Sub(g.lastUpdate).Seconds()
	g.lastUpdate = now

	// Limit delta time to prevent large jumps
	if dt > 0.1 {
		dt = 0.1
	}

	// Handle input
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return fmt.Errorf("quit")
	}

	if ebiten.IsKeyPressed(ebiten.Key1) {
		g.mode = 0 // Idle demo
		g.demoTimer = 0
		time.Sleep(100 * time.Millisecond)
	}

	if ebiten.IsKeyPressed(ebiten.Key2) {
		g.mode = 1 // Walk demo
		g.demoTimer = 0
		time.Sleep(100 * time.Millisecond)
	}

	if ebiten.IsKeyPressed(ebiten.Key3) {
		g.mode = 2 // Attack demo
		g.demoTimer = 0
		time.Sleep(100 * time.Millisecond)
	}

	if ebiten.IsKeyPressed(ebiten.KeyH) {
		g.showInstructions = !g.showInstructions
		time.Sleep(100 * time.Millisecond)
	}

	// Update demo timer
	g.demoTimer += dt

	// Run demo based on mode
	switch g.mode {
	case 0: // Idle
		g.runIdleDemo()
	case 1: // Walk
		g.runWalkDemo(dt)
	case 2: // Attack
		g.runAttackDemo(dt)
	}

	// Update systems
	entities := g.world.GetEntities()

	g.movementSys.Update(entities, dt)
	if err := g.animationSys.Update(entities, dt); err != nil {
		g.logger.WithError(err).Error("animation system error")
	}
	g.combatSys.Update(entities, dt)

	// Update camera to follow player
	if posComp, ok := g.player.GetComponent("position"); ok {
		pos := posComp.(*engine.PositionComponent)
		// Note: Camera system requires an active camera entity with CameraComponent
		// For this demo, we'll handle camera transform directly in WorldToScreen calls
		_ = pos
	}

	return nil
}

// runIdleDemo keeps entities idle.
func (g *Game) runIdleDemo() {
	// Stop movement
	engine.SetVelocity(g.player, 0, 0)
	engine.SetVelocity(g.enemy, 0, 0)

	// Ensure idle animation
	if animComp, ok := g.player.GetComponent("animation"); ok {
		anim := animComp.(*engine.AnimationComponent)
		if anim.CurrentState != engine.AnimationStateIdle {
			anim.SetState(engine.AnimationStateIdle)
		}
	}
}

// runWalkDemo makes entities walk in patterns.
func (g *Game) runWalkDemo(dt float64) {
	// Move player in a circle pattern
	speed := 50.0
	vx := speed * 2.0 * 3.14159 / 4.0
	vy := speed * 2.0 * 3.14159 / 4.0

	if int(g.demoTimer*2)%4 == 0 {
		engine.SetVelocity(g.player, vx, 0)
	} else if int(g.demoTimer*2)%4 == 1 {
		engine.SetVelocity(g.player, 0, vy)
	} else if int(g.demoTimer*2)%4 == 2 {
		engine.SetVelocity(g.player, -vx, 0)
	} else {
		engine.SetVelocity(g.player, 0, -vy)
	}

	// Move enemy in opposite pattern
	if int(g.demoTimer*2)%4 == 2 {
		engine.SetVelocity(g.enemy, vx, 0)
	} else if int(g.demoTimer*2)%4 == 3 {
		engine.SetVelocity(g.enemy, 0, vy)
	} else if int(g.demoTimer*2)%4 == 0 {
		engine.SetVelocity(g.enemy, -vx, 0)
	} else {
		engine.SetVelocity(g.enemy, 0, -vy)
	}
}

// runAttackDemo triggers attacks periodically.
func (g *Game) runAttackDemo(dt float64) {
	// Stop movement
	engine.SetVelocity(g.player, 0, 0)
	engine.SetVelocity(g.enemy, 0, 0)

	// Trigger attack every 2 seconds
	if int(g.demoTimer)%2 == 0 && g.demoTimer-float64(int(g.demoTimer)) < dt*2 {
		// Player attacks enemy
		if attackComp, ok := g.player.GetComponent("attack"); ok {
			attack := attackComp.(*engine.AttackComponent)
			if attack.CanAttack() {
				g.combatSys.Attack(g.player, g.enemy)
			}
		}
	} else if (int(g.demoTimer)+1)%2 == 0 && g.demoTimer-float64(int(g.demoTimer)) < dt*2 {
		// Enemy attacks player (offset by 1 second)
		if attackComp, ok := g.enemy.GetComponent("attack"); ok {
			attack := attackComp.(*engine.AttackComponent)
			if attack.CanAttack() {
				g.combatSys.Attack(g.enemy, g.player)
			}
		}
	}
}

// Draw renders the game screen.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 30, 255})

	// Draw entities
	g.drawEntity(screen, g.player, color.RGBA{100, 150, 255, 255})
	g.drawEntity(screen, g.enemy, color.RGBA{255, 100, 100, 255})

	// Draw UI
	g.drawUI(screen)
}

// drawEntity draws an entity with its current animation frame.
func (g *Game) drawEntity(screen *ebiten.Image, entity *engine.Entity, tintColor color.RGBA) {
	posComp, ok := entity.GetComponent("position")
	if !ok {
		return
	}
	pos := posComp.(*engine.PositionComponent)

	spriteComp, ok := entity.GetComponent("sprite")
	if !ok {
		return
	}
	sprite := spriteComp.(*engine.EbitenSprite)

	// Get current animation frame
	if sprite.Image != nil {
		op := &ebiten.DrawImageOptions{}

		// Apply camera transform
		worldX, worldY := pos.X, pos.Y
		screenX, screenY := g.camera.WorldToScreen(worldX, worldY)

		// Center sprite on position
		op.GeoM.Translate(-float64(sprite.Width)/2, -float64(sprite.Height)/2)
		op.GeoM.Translate(screenX, screenY)

		// Scale up for visibility
		scale := 3.0
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(float64(sprite.Width)*scale/2, float64(sprite.Height)*scale/2)

		screen.DrawImage(sprite.Image, op)
	} else {
		// Fallback: draw colored rectangle if no sprite
		screenX, screenY := g.camera.WorldToScreen(pos.X, pos.Y)
		ebitenutil.DrawRect(screen, screenX-16, screenY-16, 32, 32, tintColor)
	}

	// Draw health bar
	g.drawHealthBar(screen, entity, pos.X, pos.Y)

	// Draw animation state label
	g.drawAnimationState(screen, entity, pos.X, pos.Y)
}

// drawHealthBar draws a health bar above the entity.
func (g *Game) drawHealthBar(screen *ebiten.Image, entity *engine.Entity, worldX, worldY float64) {
	healthComp, ok := entity.GetComponent("health")
	if !ok {
		return
	}
	health := healthComp.(*engine.HealthComponent)

	screenX, screenY := g.camera.WorldToScreen(worldX, worldY)

	barWidth := 60.0
	barHeight := 6.0
	barX := screenX - barWidth/2
	barY := screenY - 70

	// Background (red)
	ebitenutil.DrawRect(screen, barX, barY, barWidth, barHeight, color.RGBA{100, 0, 0, 200})

	// Foreground (green based on health %)
	healthPercent := health.Current / health.Max
	ebitenutil.DrawRect(screen, barX, barY, barWidth*healthPercent, barHeight, color.RGBA{0, 200, 0, 200})

	// Border
	ebitenutil.DrawRect(screen, barX, barY, barWidth, 1, color.RGBA{255, 255, 255, 255})
	ebitenutil.DrawRect(screen, barX, barY+barHeight, barWidth, 1, color.RGBA{255, 255, 255, 255})
	ebitenutil.DrawRect(screen, barX, barY, 1, barHeight, color.RGBA{255, 255, 255, 255})
	ebitenutil.DrawRect(screen, barX+barWidth, barY, 1, barHeight, color.RGBA{255, 255, 255, 255})
}

// drawAnimationState draws the current animation state label.
func (g *Game) drawAnimationState(screen *ebiten.Image, entity *engine.Entity, worldX, worldY float64) {
	animComp, ok := entity.GetComponent("animation")
	if !ok {
		return
	}
	anim := animComp.(*engine.AnimationComponent)

	screenX, screenY := g.camera.WorldToScreen(worldX, worldY)

	stateText := anim.CurrentState.String()
	frameText := fmt.Sprintf("Frame: %d/%d", anim.FrameIndex+1, len(anim.Frames))

	ebitenutil.DebugPrintAt(screen, stateText, int(screenX)-30, int(screenY)+60)
	ebitenutil.DebugPrintAt(screen, frameText, int(screenX)-30, int(screenY)+75)
}

// drawUI draws the user interface.
func (g *Game) drawUI(screen *ebiten.Image) {
	// Title
	ebitenutil.DebugPrintAt(screen, "Animation System Demo", 10, 10)

	// Mode indicator
	modeText := ""
	switch g.mode {
	case 0:
		modeText = "Mode: IDLE (Press 1)"
	case 1:
		modeText = "Mode: WALK (Press 2)"
	case 2:
		modeText = "Mode: ATTACK (Press 3)"
	}
	ebitenutil.DebugPrintAt(screen, modeText, 10, 30)

	// Timer
	timerText := fmt.Sprintf("Time: %.1fs", g.demoTimer)
	ebitenutil.DebugPrintAt(screen, timerText, 10, 50)

	// Instructions
	if g.showInstructions {
		instructions := []string{
			"Controls:",
			"1 - Idle Animation Demo",
			"2 - Walk Animation Demo",
			"3 - Attack Animation Demo",
			"H - Toggle Help",
			"ESC - Quit",
			"",
			"Blue = Player",
			"Red = Enemy",
		}

		y := screenHeight - len(instructions)*15 - 10
		for i, line := range instructions {
			ebitenutil.DebugPrintAt(screen, line, 10, y+i*15)
		}
	} else {
		ebitenutil.DebugPrintAt(screen, "Press H for help", 10, screenHeight-25)
	}

	// Animation system stats
	cacheSize := g.animationSys.GetCacheSize()
	statsText := fmt.Sprintf("Animation Cache: %d sequences", cacheSize)
	ebitenutil.DebugPrintAt(screen, statsText, screenWidth-250, 10)
}

// Layout returns the game's logical screen size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	seed := flag.Int64("seed", 12345, "Random seed")
	genre := flag.String("genre", "fantasy", "Genre (fantasy, scifi, horror, cyberpunk, postapoc)")
	flag.Parse()

	game := NewGame(*seed, *genre)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Venture - Animation System Demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	logger.Info("Starting animation demo...")
	logger.Info("Press 1, 2, or 3 to switch demo modes")
	logger.Info("Press H to toggle help")
	logger.Info("Press ESC to quit")

	if err := ebiten.RunGame(game); err != nil && err.Error() != "quit" {
		logger.WithError(err).Fatal("error")
	}

	logger.Info("Demo finished")
}
