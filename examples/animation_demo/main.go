//go:build !testpackage
// +build !testpackage

// Package main demonstrates the animation system with procedural sprite generation.
// This example shows how to integrate animation components with the ECS architecture.
package main

import (
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

// Game represents the game state for animation demo.
type Game struct {
	world         *engine.World
	animSystem    *engine.AnimationSystem
	entities      []*engine.Entity
	currentEntity int
	frameCount    int
}

// NewGame creates a new animation demo game.
func NewGame() (*Game, error) {
	// Initialize world and systems
	world := engine.NewWorld()
	spriteGen := sprites.NewGenerator()
	animSystem := engine.NewAnimationSystem(spriteGen)

	// Add animation system to world
	world.AddSystem(animSystem)

	// Create demo entities with different animation states
	game := &Game{
		world:      world,
		animSystem: animSystem,
		entities:   make([]*engine.Entity, 0),
	}

	// Generate palette
	pal, err := palette.NewGenerator().Generate("fantasy", 12345)
	if err != nil {
		return nil, fmt.Errorf("failed to generate palette: %w", err)
	}

	// Create entities demonstrating different animation states
	states := []engine.AnimationState{
		engine.AnimationStateIdle,
		engine.AnimationStateWalk,
		engine.AnimationStateAttack,
		engine.AnimationStateCast,
		engine.AnimationStateHit,
	}

	for i, state := range states {
		entity := game.createAnimatedEntity(int64(12345+i), state, pal, i)
		game.entities = append(game.entities, entity)
	}

	return game, nil
}

// createAnimatedEntity creates an entity with animation and sprite components.
func (g *Game) createAnimatedEntity(seed int64, state engine.AnimationState, pal *palette.Palette, index int) *engine.Entity {
	entity := g.world.CreateEntity()

	// Add position component
	posComp := &engine.PositionComponent{
		X: float64(100 + index*120),
		Y: 240,
	}
	entity.AddComponent(posComp)

	// Add sprite component (will be updated by animation system)
	spriteComp := &engine.EbitenSprite{
		Image:   ebiten.NewImage(28, 28),
		Width:   28,
		Height:  28,
		Visible: true,
		Layer:   1,
	}
	entity.AddComponent(spriteComp)

	// Add animation component
	animComp := engine.NewAnimationComponent(seed)
	animComp.SetState(state)
	animComp.FrameTime = 0.1 // 10 FPS
	animComp.Loop = true
	entity.AddComponent(animComp)

	return entity
}

// Update updates the game logic.
func (g *Game) Update() error {
	// Update world (including animation system)
	deltaTime := 1.0 / 60.0 // 60 FPS
	g.world.Update(deltaTime)

	g.frameCount++

	// Cycle through states every 3 seconds (180 frames at 60 FPS)
	if g.frameCount%180 == 0 {
		g.currentEntity = (g.currentEntity + 1) % len(g.entities)

		// Transition current entity to next state
		if g.currentEntity < len(g.entities) {
			entity := g.entities[g.currentEntity]
			animComp, ok := entity.GetComponent("animation")
			if ok {
				if anim, ok := animComp.(*engine.AnimationComponent); ok {
					// Cycle through states
					nextState := g.getNextState(anim.CurrentState)
					anim.SetState(nextState)
					fmt.Printf("Entity %d transitioned to state: %s\n", g.currentEntity, nextState)
				}
			}
		}
	}

	return nil
}

// getNextState returns the next animation state in the cycle.
func (g *Game) getNextState(current engine.AnimationState) engine.AnimationState {
	states := []engine.AnimationState{
		engine.AnimationStateIdle,
		engine.AnimationStateWalk,
		engine.AnimationStateAttack,
		engine.AnimationStateCast,
		engine.AnimationStateHit,
	}

	for i, state := range states {
		if state == current {
			return states[(i+1)%len(states)]
		}
	}

	return engine.AnimationStateIdle
}

// Draw renders the game.
func (g *Game) Draw(screen *ebiten.Image) {
	// Clear screen
	screen.Fill(palette.ColorFromRGB(20, 20, 30))

	// Draw all entities
	for _, entity := range g.entities {
		g.drawEntity(screen, entity)
	}

	// Draw labels
	g.drawLabels(screen)
}

// drawEntity renders a single entity.
func (g *Game) drawEntity(screen *ebiten.Image, entity *engine.Entity) {
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

	if sprite.Image == nil || !sprite.Visible {
		return
	}

	// Draw sprite
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(pos.X-sprite.Width/2, pos.Y-sprite.Height/2)
	screen.DrawImage(sprite.Image, opts)
}

// drawLabels draws state labels below entities.
func (g *Game) drawLabels(screen *ebiten.Image) {
	// In a real game, you'd use text rendering here
	// For this example, we skip text to avoid additional dependencies
}

// Layout returns the game screen size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	fmt.Println("Animation System Demo")
	fmt.Println("=====================")
	fmt.Println("This demo showcases the procedural animation system.")
	fmt.Println("Watch as entities cycle through different animation states.")
	fmt.Println()
	fmt.Println("Features demonstrated:")
	fmt.Println("- Multi-frame sprite animation (4-8 frames per state)")
	fmt.Println("- State transitions (Idle → Walk → Attack → Cast → Hit)")
	fmt.Println("- Frame caching for performance")
	fmt.Println("- Deterministic procedural generation")
	fmt.Println()

	game, err := NewGame()
	if err != nil {
		log.Fatalf("Failed to create game: %v", err)
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Venture - Animation System Demo")
	ebiten.SetTPS(60)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
