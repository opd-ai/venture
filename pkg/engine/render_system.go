// Package engine provides sprite rendering for entities.
// This file implements RenderSystem which handles entity sprite rendering
// with camera transformations and visual effects.
package engine

import (
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// EbitenSprite holds visual representation data for an entity (Ebiten implementation).
// Implements SpriteProvider interface.
type EbitenSprite struct {
	// Sprite image (procedurally generated)
	Image *ebiten.Image

	// Directional sprite images for aerial-view rendering (Phase 2: Aerial Template Integration)
	// Maps Direction (Up/Down/Left/Right) to corresponding sprite image
	DirectionalImages map[int]*ebiten.Image

	// Current facing direction for sprite selection
	CurrentDirection int

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

// Type returns the component type identifier (implements Component).
func (s *EbitenSprite) Type() string {
	return "sprite"
}

// GetImage implements SpriteProvider interface.
func (s *EbitenSprite) GetImage() ImageProvider {
	if s.Image == nil {
		return nil
	}
	return &EbitenImage{image: s.Image}
}

// GetSize implements SpriteProvider interface.
func (s *EbitenSprite) GetSize() (width, height float64) {
	return s.Width, s.Height
}

// GetColor implements SpriteProvider interface.
func (s *EbitenSprite) GetColor() color.Color {
	if s.Color == nil {
		return color.White
	}
	return s.Color
}

// GetRotation implements SpriteProvider interface.
func (s *EbitenSprite) GetRotation() float64 {
	return s.Rotation
}

// GetLayer implements SpriteProvider interface.
func (s *EbitenSprite) GetLayer() int {
	return s.Layer
}

// IsVisible implements SpriteProvider interface.
func (s *EbitenSprite) IsVisible() bool {
	return s.Visible
}

// SetVisible implements SpriteProvider interface.
func (s *EbitenSprite) SetVisible(visible bool) {
	s.Visible = visible
}

// SetColor implements SpriteProvider interface.
func (s *EbitenSprite) SetColor(col color.Color) {
	s.Color = col
}

// SetRotation implements SpriteProvider interface.
func (s *EbitenSprite) SetRotation(rotation float64) {
	s.Rotation = rotation
}

// NewSpriteComponent creates a new Ebiten sprite component.
func NewSpriteComponent(width, height float64, color color.Color) *EbitenSprite {
	return &EbitenSprite{
		Width:             width,
		Height:            height,
		Color:             color,
		Visible:           true,
		Layer:             0,
		DirectionalImages: make(map[int]*ebiten.Image), // Initialize directional sprite map
		CurrentDirection:  1,                            // Default to DirDown (1)
	}
}

// EbitenImage wraps an Ebiten image for the ImageProvider interface.
type EbitenImage struct {
	image *ebiten.Image
}

// GetSize implements ImageProvider interface.
func (e *EbitenImage) GetSize() (width, height int) {
	if e.image == nil {
		return 0, 0
	}
	return e.image.Bounds().Dx(), e.image.Bounds().Dy()
}

// GetPixel implements ImageProvider interface.
func (e *EbitenImage) GetPixel(x, y int) color.Color {
	if e.image == nil {
		return color.Transparent
	}
	return e.image.At(x, y)
}

// Compile-time interface checks
var (
	_ SpriteProvider = (*EbitenSprite)(nil)
	_ ImageProvider  = (*EbitenImage)(nil)
)

// EbitenRenderSystem handles rendering of entities to the screen (Ebiten implementation).
// Implements RenderingSystem interface.
type EbitenRenderSystem struct {
	screen       *ebiten.Image
	cameraSystem *CameraSystem

	// Spatial partitioning for viewport culling
	spatialPartition *SpatialPartitionSystem
	enableCulling    bool

	// Batch rendering optimization
	enableBatching bool
	batches        map[*ebiten.Image][]*Entity // Group entities by sprite image
	batchPool      []map[*ebiten.Image][]*Entity

	// Debug rendering flags
	ShowColliders bool
	ShowGrid      bool

	// Performance statistics
	stats RenderStats
}

// RenderStats tracks rendering performance metrics.
type RenderStats struct {
	TotalEntities    int // Total entities in scene
	RenderedEntities int // Entities actually rendered
	CulledEntities   int // Entities culled by viewport check
	BatchCount       int // Number of batches created
	LastFrameTime    float64
}

// NewRenderSystem creates a new render system.
func NewRenderSystem(cameraSystem *CameraSystem) *EbitenRenderSystem {
	return &EbitenRenderSystem{
		cameraSystem:     cameraSystem,
		spatialPartition: nil, // Will be set when world bounds are known
		enableCulling:    true,
		enableBatching:   true, // Batching enabled by default
		batches:          make(map[*ebiten.Image][]*Entity),
		batchPool:        make([]map[*ebiten.Image][]*Entity, 0, 2),
		ShowColliders:    false,
		ShowGrid:         false,
	}
}

// SetScreen sets the render target.
func (r *EbitenRenderSystem) SetScreen(screen *ebiten.Image) {
	r.screen = screen
}

// SetSpatialPartition sets the spatial partition system for viewport culling.
// This enables efficient culling of off-screen entities.
func (r *EbitenRenderSystem) SetSpatialPartition(partition *SpatialPartitionSystem) {
	r.spatialPartition = partition
}

// EnableCulling enables or disables viewport culling.
// When disabled, all entities are rendered (useful for debugging).
func (r *EbitenRenderSystem) EnableCulling(enable bool) {
	r.enableCulling = enable
}

// EnableBatching enables or disables batch rendering.
// When enabled, entities with the same sprite are grouped to reduce GPU state changes.
func (r *EbitenRenderSystem) EnableBatching(enable bool) {
	r.enableBatching = enable
}

// GetStats returns rendering performance statistics.
func (r *EbitenRenderSystem) GetStats() RenderStats {
	return r.stats
}

// Update is called every frame but doesn't modify entities.
// Actual rendering happens in Draw which is called by ebiten.
func (r *EbitenRenderSystem) Update(entities []*Entity, deltaTime float64) {
	// RenderSystem doesn't need to update entity state
	// Rendering is handled in the Draw call
}

// Draw renders all visible entities to the screen (implements RenderingSystem interface).
// This should be called from the game's Draw method.
// The screen parameter should be *ebiten.Image in production.
func (r *EbitenRenderSystem) Draw(screen interface{}, entities []*Entity) {
	// Type assert to *ebiten.Image
	ebitenScreen, ok := screen.(*ebiten.Image)
	if !ok {
		return // Invalid screen type
	}
	r.screen = ebitenScreen

	// Reset stats for this frame
	r.stats = RenderStats{
		TotalEntities: len(entities),
	}

	// Note: Screen clearing is handled by terrain rendering system

	// Get visible entities using spatial partition (if enabled)
	visibleEntities := entities
	if r.enableCulling && r.spatialPartition != nil && r.cameraSystem != nil {
		visibleEntities = r.getVisibleEntities(entities)
	}

	// Sort entities by layer
	sortedEntities := r.sortEntitiesByLayer(visibleEntities)

	// Render using batching (if enabled) or individual draws
	if r.enableBatching {
		r.drawBatched(sortedEntities)
	} else {
		for _, entity := range sortedEntities {
			r.drawEntity(entity)
			r.stats.RenderedEntities++
		}
	}

	// Calculate culled count
	r.stats.CulledEntities = r.stats.TotalEntities - r.stats.RenderedEntities

	// GAP-016 REPAIR: Draw particle effects
	r.drawParticles(entities)

	// Draw debug overlays
	if r.ShowColliders {
		r.drawColliders(sortedEntities)
	}
}

// drawBatched renders entities using batch optimization to reduce GPU state changes.
// Entities with the same sprite image are grouped together.
func (r *EbitenRenderSystem) drawBatched(entities []*Entity) {
	// Get or create batch map from pool
	batches := r.getBatchMap()
	defer r.returnBatchMap(batches)

	// Group entities by sprite image
	for _, entity := range entities {
		spriteComp, hasSprite := entity.GetComponent("sprite")
		if !hasSprite {
			continue
		}
		sprite := spriteComp.(*EbitenSprite)

		if !sprite.Visible || sprite.Image == nil {
			continue
		}

		// Group by sprite image pointer (entities with same sprite are batched)
		batches[sprite.Image] = append(batches[sprite.Image], entity)
	}

	r.stats.BatchCount = len(batches)

	// Draw each batch
	for _, batch := range batches {
		r.drawBatch(batch)
	}
}

// drawBatch renders a group of entities with the same sprite image.
func (r *EbitenRenderSystem) drawBatch(entities []*Entity) {
	for _, entity := range entities {
		r.drawEntity(entity)
		r.stats.RenderedEntities++
	}
}

// getBatchMap retrieves a batch map from the pool or creates a new one.
func (r *EbitenRenderSystem) getBatchMap() map[*ebiten.Image][]*Entity {
	if len(r.batchPool) > 0 {
		// Pop from pool
		batches := r.batchPool[len(r.batchPool)-1]
		r.batchPool = r.batchPool[:len(r.batchPool)-1]

		// Clear the map
		for k := range batches {
			batches[k] = batches[k][:0] // Reuse slice capacity
		}
		return batches
	}

	// Create new map with initial capacity
	return make(map[*ebiten.Image][]*Entity, 32)
}

// returnBatchMap returns a batch map to the pool for reuse.
func (r *EbitenRenderSystem) returnBatchMap(batches map[*ebiten.Image][]*Entity) {
	if len(r.batchPool) < 2 { // Keep small pool
		r.batchPool = append(r.batchPool, batches)
	}
}

// getVisibleEntities returns only entities visible in the current viewport.
// This uses spatial partitioning for efficient culling.
func (r *EbitenRenderSystem) getVisibleEntities(entities []*Entity) []*Entity {
	// Get camera bounds in world space
	cam := r.cameraSystem.activeCamera
	if cam == nil {
		return entities // No camera, render all
	}

	camComp, ok := cam.GetComponent("camera")
	if !ok {
		return entities
	}
	camera := camComp.(*CameraComponent)

	// Calculate viewport bounds in world space with margin for sprites
	margin := 100.0 // Extra space to render sprites partially off-screen

	// Get camera position
	camPos, ok := cam.GetComponent("position")
	if !ok {
		return entities
	}
	pos := camPos.(*PositionComponent)

	// Calculate world viewport bounds
	viewportWidth := float64(r.cameraSystem.ScreenWidth) / camera.Zoom
	viewportHeight := float64(r.cameraSystem.ScreenHeight) / camera.Zoom

	viewportBounds := Bounds{
		X:      pos.X - viewportWidth/2 - margin,
		Y:      pos.Y - viewportHeight/2 - margin,
		Width:  viewportWidth + margin*2,
		Height: viewportHeight + margin*2,
	}

	// Query spatial partition for entities in viewport
	visible := r.spatialPartition.QueryBounds(viewportBounds)

	return visible
}

// drawEntity renders a single entity.
func (r *EbitenRenderSystem) drawEntity(entity *Entity) {
	// Get required components
	posComp, hasPos := entity.GetComponent("position")
	spriteComp, hasSprite := entity.GetComponent("sprite")

	if !hasPos || !hasSprite {
		return
	}

	pos := posComp.(*PositionComponent)
	sprite := spriteComp.(*EbitenSprite)

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
	// Phase 2: Support directional sprites with fallback to single image
	var spriteImage *ebiten.Image
	if len(sprite.DirectionalImages) > 0 {
		// Use directional sprite if available
		if dirImg, exists := sprite.DirectionalImages[sprite.CurrentDirection]; exists && dirImg != nil {
			spriteImage = dirImg
		} else {
			// Fallback to default direction or single image
			spriteImage = sprite.Image
		}
	} else {
		// Use single image (backward compatibility)
		spriteImage = sprite.Image
	}

	if spriteImage != nil {
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
		r.screen.DrawImage(spriteImage, opts)
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
func (r *EbitenRenderSystem) drawHealthBar(entity *Entity, screenX, screenY, spriteWidth, spriteHeight float64) {
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
func (r *EbitenRenderSystem) drawParticles(entities []*Entity) {
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
func (r *EbitenRenderSystem) drawRect(x, y, width, height float64, col color.Color) {
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
func (r *EbitenRenderSystem) drawColliders(entities []*Entity) {
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
// Optimized: Uses Go's sort.Slice (O(n log n)) and caches sprite components to avoid repeated map lookups.
func (r *EbitenRenderSystem) sortEntitiesByLayer(entities []*Entity) []*Entity {
	// Pre-allocate with capacity
	sorted := make([]*Entity, 0, len(entities))

	// Cache sprite components to avoid repeated GetComponent calls
	type entitySprite struct {
		entity *Entity
		sprite *EbitenSprite
		layer  int
	}

	cache := make([]entitySprite, 0, len(entities))

	// Collect entities with sprites and cache their sprite components
	for _, entity := range entities {
		if sprite, ok := entity.GetComponent("sprite"); ok {
			ebitenSprite := sprite.(*EbitenSprite)
			cache = append(cache, entitySprite{
				entity: entity,
				sprite: ebitenSprite,
				layer:  ebitenSprite.Layer,
			})
		}
	}

	// Sort using Go's optimized sort (O(n log n) instead of O(n²) bubble sort)
	sort.Slice(cache, func(i, j int) bool {
		return cache[i].layer < cache[j].layer
	})

	// Extract sorted entities
	for _, es := range cache {
		sorted = append(sorted, es.entity)
	}

	return sorted
}

// SetShowColliders implements RenderingSystem interface.
func (r *EbitenRenderSystem) SetShowColliders(show bool) {
	r.ShowColliders = show
}

// SetShowGrid implements RenderingSystem interface.
func (r *EbitenRenderSystem) SetShowGrid(show bool) {
	r.ShowGrid = show
}

// Compile-time interface check
var _ RenderingSystem = (*EbitenRenderSystem)(nil)
