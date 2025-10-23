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

	// GAP-016 REPAIR: Draw particle effects
	r.drawParticles(entities)

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

	// GAP-012 REPAIR: Apply visual feedback effects (hit flash, tints)
	var flashAlpha float64
	var tintR, tintG, tintB, tintA float64 = 1.0, 1.0, 1.0, 1.0
	if feedbackComp, ok := entity.GetComponent("visual_feedback"); ok {
		feedback := feedbackComp.(*VisualFeedbackComponent)
		flashAlpha = feedback.GetFlashAlpha()
		tintR, tintG, tintB, tintA = feedback.TintR, feedback.TintG, feedback.TintB, feedback.TintA
	}

	// Draw sprite or colored rectangle
	if sprite.Image != nil {
		// Draw procedural sprite
		opts := &ebiten.DrawImageOptions{}

		// GAP-012 REPAIR: Apply color effects
		if flashAlpha > 0 || tintR != 1.0 || tintG != 1.0 || tintB != 1.0 || tintA != 1.0 {
			// Apply flash (additive white) and tint (multiplicative color)
			opts.ColorScale.ScaleWithColor(color.RGBA{
				R: uint8((tintR + flashAlpha) * 255),
				G: uint8((tintG + flashAlpha) * 255),
				B: uint8((tintB + flashAlpha) * 255),
				A: uint8(tintA * 255),
			})
		}

		opts.GeoM.Translate(-sprite.Width/2, -sprite.Height/2) // Center
		opts.GeoM.Rotate(sprite.Rotation)
		opts.GeoM.Translate(screenX, screenY)
		r.screen.DrawImage(sprite.Image, opts)
	} else {
		// Draw colored rectangle as fallback
		col := sprite.Color

		// GAP-012 REPAIR: Apply flash to fallback rect
		if flashAlpha > 0 {
			red, green, blue, alpha := col.RGBA()
			col = color.RGBA{
				R: uint8((float64(red>>8) + flashAlpha*255) / 2),
				G: uint8((float64(green>>8) + flashAlpha*255) / 2),
				B: uint8((float64(blue>>8) + flashAlpha*255) / 2),
				A: uint8(alpha >> 8),
			}
		}

		r.drawRect(screenX-sprite.Width/2, screenY-sprite.Height/2,
			sprite.Width, sprite.Height, col)
	}

	// GAP-013 REPAIR: Draw health bar for damaged enemies and bosses
	r.drawHealthBar(entity, screenX, screenY, sprite.Width, sprite.Height)
}

// drawHealthBar renders a health bar above an entity if appropriate.
// GAP-013 REPAIR: Shows health status for enemies (when damaged) and bosses (always).
func (r *RenderSystem) drawHealthBar(entity *Entity, screenX, screenY, spriteWidth, spriteHeight float64) {
	// Only draw health bars for entities with health component
	healthComp, hasHealth := entity.GetComponent("health")
	if !hasHealth {
		return
	}

	health := healthComp.(*HealthComponent)

	// Don't draw health bar for player (has HUD display)
	if entity.HasComponent("input") {
		return
	}

	// Check if entity is a boss (high attack indicates boss)
	isBoss := false
	if attackComp, ok := entity.GetComponent("attack"); ok {
		attack := attackComp.(*AttackComponent)
		isBoss = attack.Damage > 20 // Boss threshold
	}

	// Only show health bar if: (1) damaged, or (2) is boss
	if health.Current >= health.Max && !isBoss {
		return
	}

	// Calculate health bar dimensions
	barWidth := spriteWidth
	barHeight := 4.0
	barX := screenX - barWidth/2
	barY := screenY - spriteHeight/2 - barHeight - 5 // 5px above sprite

	// Draw background (dark gray)
	bgColor := color.RGBA{40, 40, 40, 200}
	vector.DrawFilledRect(r.screen, float32(barX), float32(barY),
		float32(barWidth), float32(barHeight), bgColor, false)

	// Calculate health percentage
	healthPercent := health.Current / health.Max
	if healthPercent < 0 {
		healthPercent = 0
	}
	if healthPercent > 1 {
		healthPercent = 1
	}

	// Determine health bar color (green → yellow → red)
	var healthColor color.RGBA
	if healthPercent > 0.6 {
		// Green (healthy)
		healthColor = color.RGBA{50, 200, 50, 255}
	} else if healthPercent > 0.3 {
		// Yellow (wounded)
		healthColor = color.RGBA{220, 220, 50, 255}
	} else {
		// Red (critical)
		healthColor = color.RGBA{220, 50, 50, 255}
	}

	// Draw health bar (scaled by percentage)
	healthBarWidth := barWidth * healthPercent
	vector.DrawFilledRect(r.screen, float32(barX), float32(barY),
		float32(healthBarWidth), float32(barHeight), healthColor, false)

	// Draw border around health bar (makes it more visible)
	borderColor := color.RGBA{200, 200, 200, 255}
	vector.StrokeRect(r.screen, float32(barX), float32(barY),
		float32(barWidth), float32(barHeight), 1, borderColor, false)
}

// GAP-016 REPAIR: drawParticles renders all particle effects to the screen.
func (r *RenderSystem) drawParticles(entities []*Entity) {
	for _, entity := range entities {
		comp, ok := entity.GetComponent("particle_emitter")
		if !ok {
			continue
		}

		emitter := comp.(*ParticleEmitterComponent)

		// Render each particle system
		for _, system := range emitter.Systems {
			for _, particle := range system.GetAliveParticles() {
				// Convert world coordinates to screen coordinates
				screenX, screenY := r.cameraSystem.WorldToScreen(particle.X, particle.Y)

				// Calculate alpha based on particle life (fade out)
				alpha := particle.Life
				if alpha < 0 {
					alpha = 0
				}
				if alpha > 1 {
					alpha = 1
				}

				// Extract color with alpha applied
				pr, pg, pb, _ := particle.Color.RGBA()
				particleColor := color.RGBA{
					R: uint8(pr >> 8),
					G: uint8(pg >> 8),
					B: uint8(pb >> 8),
					A: uint8(float64(255) * alpha),
				}

				// Draw particle as a small filled circle
				vector.DrawFilledCircle(r.screen,
					float32(screenX), float32(screenY),
					float32(particle.Size),
					particleColor, false)
			}
		}
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
