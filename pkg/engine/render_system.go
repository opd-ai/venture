//go:build !test
// +build !test

// Package engine provides sprite rendering for entities.
// This file implements RenderSystem which handles entity sprite rendering
// with camera transformations and visual effects.
package engine

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// SpriteComponent holds visual representation data for an entity.
type SpriteComponent struct {
	// Sprite image (procedurally generated)
	Image *ebiten.Image

	// Color tint
	Color color.Color

	// Size (width, height)
	Width, Height float64

	// Rotation in radians
	Rotation float64

	// Visibility flag
	Visible bool

	// Layer for rendering order (higher = drawn on top)
	Layer int
}

// Type returns the component type identifier.
func (s *SpriteComponent) Type() string {
	return "sprite"
}

// NewSpriteComponent creates a new sprite component.
func NewSpriteComponent(width, height float64, color color.Color) *SpriteComponent {
	return &SpriteComponent{
		Width:   width,
		Height:  height,
		Color:   color,
		Visible: true,
		Layer:   0,
	}
}

// RenderSystem handles rendering of entities to the screen.
type RenderSystem struct {
	screen       *ebiten.Image
	cameraSystem *CameraSystem

	// Debug rendering flags
	ShowColliders bool
	ShowGrid      bool
}

// NewRenderSystem creates a new render system.
func NewRenderSystem(cameraSystem *CameraSystem) *RenderSystem {
	return &RenderSystem{
		cameraSystem:  cameraSystem,
		ShowColliders: false,
		ShowGrid:      false,
	}
}

// SetScreen sets the render target.
func (r *RenderSystem) SetScreen(screen *ebiten.Image) {
	r.screen = screen
}

// Update is called every frame but doesn't modify entities.
// Actual rendering happens in Draw which is called by ebiten.
func (r *RenderSystem) Update(entities []*Entity, deltaTime float64) {
	// RenderSystem doesn't need to update entity state
	// Rendering is handled in the Draw call
}

// Draw renders all visible entities to the screen.
// This should be called from the game's Draw method.
func (r *RenderSystem) Draw(screen *ebiten.Image, entities []*Entity) {
	r.screen = screen

	// Clear screen
	screen.Fill(color.RGBA{20, 20, 30, 255}) // Dark background

	// Sort entities by layer
	sortedEntities := r.sortEntitiesByLayer(entities)

	// Draw each entity
	for _, entity := range sortedEntities {
		r.drawEntity(entity)
	}

	// Draw debug overlays
	if r.ShowColliders {
		r.drawColliders(sortedEntities)
	}
}

// drawEntity renders a single entity.
func (r *RenderSystem) drawEntity(entity *Entity) {
	// Get required components
	posComp, hasPos := entity.GetComponent("position")
	spriteComp, hasSprite := entity.GetComponent("sprite")

	if !hasPos || !hasSprite {
		return
	}

	pos := posComp.(*PositionComponent)
	sprite := spriteComp.(*SpriteComponent)

	if !sprite.Visible {
		return
	}

	// Convert world position to screen position
	screenX, screenY := r.cameraSystem.WorldToScreen(pos.X, pos.Y)

	// Check if entity is visible on screen
	if !r.cameraSystem.IsVisible(pos.X, pos.Y, sprite.Width) {
		return
	}

	// Draw sprite or colored rectangle
	if sprite.Image != nil {
		// Draw procedural sprite
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(-sprite.Width/2, -sprite.Height/2) // Center
		opts.GeoM.Rotate(sprite.Rotation)
		opts.GeoM.Translate(screenX, screenY)
		r.screen.DrawImage(sprite.Image, opts)
	} else {
		// Draw colored rectangle as fallback
		r.drawRect(screenX-sprite.Width/2, screenY-sprite.Height/2,
			sprite.Width, sprite.Height, sprite.Color)
	}
}

// drawRect draws a filled rectangle at the given screen position.
func (r *RenderSystem) drawRect(x, y, width, height float64, col color.Color) {
	// Convert color
	red, green, blue, alpha := col.RGBA()
	clr := color.RGBA{
		R: uint8(red >> 8),
		G: uint8(green >> 8),
		B: uint8(blue >> 8),
		A: uint8(alpha >> 8),
	}

	// Draw filled rectangle using vector
	vector.DrawFilledRect(r.screen, float32(x), float32(y),
		float32(width), float32(height), clr, false)
}

// drawColliders draws collision bounds for debugging.
func (r *RenderSystem) drawColliders(entities []*Entity) {
	debugColor := color.RGBA{0, 255, 0, 128} // Semi-transparent green

	for _, entity := range entities {
		posComp, hasPos := entity.GetComponent("position")
		colliderComp, hasCollider := entity.GetComponent("collider")

		if !hasPos || !hasCollider {
			continue
		}

		pos := posComp.(*PositionComponent)
		collider := colliderComp.(*ColliderComponent)

		// Get collider bounds
		minX, minY, maxX, maxY := collider.GetBounds(pos.X, pos.Y)

		// Convert to screen coordinates
		screenX1, screenY1 := r.cameraSystem.WorldToScreen(minX, minY)
		screenX2, screenY2 := r.cameraSystem.WorldToScreen(maxX, maxY)

		// Draw rectangle outline
		width := float32(screenX2 - screenX1)
		height := float32(screenY2 - screenY1)
		vector.StrokeRect(r.screen, float32(screenX1), float32(screenY1),
			width, height, 1, debugColor, false)
	}
}

// sortEntitiesByLayer sorts entities by their sprite layer for correct draw order.
func (r *RenderSystem) sortEntitiesByLayer(entities []*Entity) []*Entity {
	sorted := make([]*Entity, 0, len(entities))

	// Collect entities with sprites
	for _, entity := range entities {
		if entity.HasComponent("sprite") {
			sorted = append(sorted, entity)
		}
	}

	// Simple bubble sort by layer (good enough for small entity counts)
	n := len(sorted)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			sprite1, _ := sorted[j].GetComponent("sprite")
			sprite2, _ := sorted[j+1].GetComponent("sprite")

			layer1 := sprite1.(*SpriteComponent).Layer
			layer2 := sprite2.(*SpriteComponent).Layer

			if layer1 > layer2 {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}
