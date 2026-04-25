// Package engine provides sprite rendering for entities.
// This file implements RenderSystem which handles entity sprite rendering
// with camera transformations and visual effects.
package engine

import (
	"cmp"
	"image/color"
	"math"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/venture/pkg/rendering/particles"
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

	// Finalized indicates whether post-processing (outline, rim lighting) has been applied.
	// Set to true by SpriteFinalizerSystem after processing; reset to false on sprite regeneration.
	Finalized bool

	// DepthProcessed indicates whether volumetric depth enhancement has been applied.
	// Set to true by SpriteDepthEnhanceSystem; reset to false on sprite regeneration.
	DepthProcessed bool

	// ColorTempProcessed indicates whether color temperature grading has been applied.
	// Set to true by SpriteColorTemperatureSystem; reset to false on sprite regeneration.
	ColorTempProcessed bool
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
		CurrentDirection:  1,                           // Default to DirDown (1)
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

// entitySprite pairs an entity with its sprite for efficient sorting.
// Caches layer and Y position to eliminate map lookups during sort comparisons.
type entitySprite struct {
	entity *Entity
	sprite *EbitenSprite
	layer  int
	yPos   float64 // Cached Y position for depth sorting (avoids O(n log n) map lookups)
}

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

	// Reusable buffer for non-sprite entities to reduce allocations
	nonSpriteBuffer []*Entity

	// Reusable buffers for sorting to reduce allocations
	sortBuffer      []*Entity
	sortCacheBuffer []entitySprite

	// Reusable buffers for batch geometry to reduce allocations
	vertexBuffer []ebiten.Vertex
	indexBuffer  []uint16

	// Reusable buffer for player entities in viewport culling to reduce allocations
	playerBuffer []*Entity

	// Reusable buffer for viewport query results to reduce allocations
	viewportQueryBuffer []*Entity

	// Reusable map for O(1) visible entity lookup during player inclusion check
	// Eliminates O(n × m) nested loop in getVisibleEntities
	visibleEntityIDs map[uint64]struct{}

	// Debug rendering flags
	ShowColliders bool
	ShowGrid      bool

	// Performance statistics
	stats RenderStats

	// Track whether spatial partition culling was used this frame
	// to avoid redundant per-entity culling
	spatialCullingUsed bool

	// Phase 2.4: Rendering optimization (PLAN.md)
	imagePool        ImagePoolProvider        // Image pool for memory efficiency (optional)
	parallelRenderer ParallelRendererProvider // Parallel renderer for performance (optional)

	// Pre-allocated DrawTrianglesOptions to avoid per-batch allocations
	drawTrianglesOptions ebiten.DrawTrianglesOptions

	// Pre-allocated DrawImageOptions to avoid per-sprite allocations in drawSpriteImage
	drawImageOptions ebiten.DrawImageOptions

	// Drop shadow rendering state
	shadowCache    *dropShadowCache
	shadowDrawOpts ebiten.DrawImageOptions

	// Player entity reference for aim indicator rendering (drawn below player sprite)
	aimPlayerEntity *Entity

	// Render interpolation alpha (0.0 to 1.0) for smooth position interpolation
	// between previous tick (PrevX/PrevY) and current tick (X/Y) positions.
	// Set by the game loop each Draw() frame based on elapsed time since last Update().
	renderAlpha float64

	// Pre-allocated buffer for particle emitter entities to avoid iterating all entities
	particleEntityBuffer []*Entity
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
		cameraSystem:         cameraSystem,
		spatialPartition:     nil,  // Will be set when world bounds are known
		enableCulling:        true, // Culling enabled by default (spatial partition bug fixed)
		enableBatching:       true, // Batching enabled by default
		batches:              make(map[*ebiten.Image][]*Entity),
		batchPool:            make([]map[*ebiten.Image][]*Entity, 0, 2),
		nonSpriteBuffer:      make([]*Entity, 0, 64),         // Pre-allocate for typical non-sprite entity count
		sortBuffer:           make([]*Entity, 0, 2000),       // Pre-allocate for typical entity count
		sortCacheBuffer:      make([]entitySprite, 0, 2000),  // Pre-allocate for typical entity count
		vertexBuffer:         make([]ebiten.Vertex, 0, 8000), // Pre-allocate for 2000 entities * 4 vertices
		indexBuffer:          make([]uint16, 0, 12000),       // Pre-allocate for 2000 entities * 6 indices
		playerBuffer:         make([]*Entity, 0, 4),          // Pre-allocate for typical player count (1-4)
		viewportQueryBuffer:  make([]*Entity, 0, 256),        // Pre-allocate for typical visible entity count
		visibleEntityIDs:     make(map[uint64]struct{}, 256), // Pre-allocate for O(1) visible lookup
		ShowColliders:        false,
		ShowGrid:             false,
		particleEntityBuffer: make([]*Entity, 0, 64), // Pre-allocate for typical particle emitter count
		drawTrianglesOptions: ebiten.DrawTrianglesOptions{
			Filter: ebiten.FilterLinear,
		},
		shadowCache: newDropShadowCache(64),
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

// GetSpatialPartition returns the spatial partition system.
func (r *EbitenRenderSystem) GetSpatialPartition() *SpatialPartitionSystem {
	return r.spatialPartition
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

// SetPool sets the image pool for memory efficiency.
// The pool is used to reuse allocated images, reducing GC pressure.
func (r *EbitenRenderSystem) SetPool(pool ImagePoolProvider) {
	r.imagePool = pool
}

// SetParallelRenderer sets the parallel renderer for performance.
// The parallel renderer distributes rendering tasks across multiple workers.
func (r *EbitenRenderSystem) SetParallelRenderer(renderer ParallelRendererProvider) {
	r.parallelRenderer = renderer
}

// GetStats returns rendering performance statistics.
func (r *EbitenRenderSystem) GetStats() RenderStats {
	return r.stats
}

// SetRenderAlpha sets the interpolation alpha for smooth rendering between ticks.
// alpha ranges from 0.0 (render at previous tick position) to 1.0 (render at current tick position).
// Called by the game loop each Draw() frame.
func (r *EbitenRenderSystem) SetRenderAlpha(alpha float64) {
	r.renderAlpha = alpha
}

// interpolatePosition returns the interpolated screen position for an entity,
// blending between the previous tick position (PrevX/PrevY) and current position (X/Y)
// using the render alpha. Both entity and camera positions are interpolated to
// eliminate visual snapping on high-refresh-rate monitors.
// G38 fix: use pos.Initialized instead of the (PrevX==0 && PrevY==0) heuristic,
// which incorrectly skipped interpolation for entities legitimately at (0,0).
func (r *EbitenRenderSystem) interpolatePosition(pos *PositionComponent) (float64, float64) {
	if r.renderAlpha >= 1.0 || !pos.Initialized {
		return r.cameraSystem.WorldToScreen(pos.X, pos.Y)
	}
	interpX := pos.PrevX + (pos.X-pos.PrevX)*r.renderAlpha
	interpY := pos.PrevY + (pos.Y-pos.PrevY)*r.renderAlpha
	return r.cameraSystem.WorldToScreenInterpolated(interpX, interpY, r.renderAlpha)
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
	ebitenScreen, ok := screen.(*ebiten.Image)
	if !ok {
		return
	}
	r.screen = ebitenScreen

	r.resetFrameStats(len(entities))
	visibleEntities := r.applyCulling(entities)
	sortedEntities := r.sortEntitiesByLayer(visibleEntities)

	// Draw aim indicator below entities (behind player sprite)
	r.drawAimIndicator()

	r.renderEntities(sortedEntities)
	r.drawParticles(r.filterParticleEntities(entities))

	if r.ShowColliders {
		r.drawColliders(sortedEntities)
	}
}

// resetFrameStats resets rendering statistics for the current frame.
func (r *EbitenRenderSystem) resetFrameStats(totalEntities int) {
	r.stats = RenderStats{
		TotalEntities: totalEntities,
	}
}

// applyCulling applies spatial culling if enabled, otherwise returns all entities.
func (r *EbitenRenderSystem) applyCulling(entities []*Entity) []*Entity {
	r.spatialCullingUsed = false
	if r.enableCulling && r.spatialPartition != nil && r.cameraSystem != nil {
		r.spatialCullingUsed = true
		return r.getVisibleEntities(entities)
	}
	return entities
}

// renderEntities renders entities using batching or individual drawing.
func (r *EbitenRenderSystem) renderEntities(entities []*Entity) {
	if r.enableBatching {
		r.drawBatched(entities)
	} else {
		for _, entity := range entities {
			r.drawEntity(entity)
			r.stats.RenderedEntities++
		}
	}
	r.stats.CulledEntities = r.stats.TotalEntities - r.stats.RenderedEntities
}

// drawBatched renders entities using batch optimization to reduce GPU state changes.
// Entities with the same sprite image are grouped together.
func (r *EbitenRenderSystem) drawBatched(entities []*Entity) {
	// Get or create batch map from pool
	batches := r.getBatchMap()
	defer r.returnBatchMap(batches)

	// Prepare non-sprite buffer
	r.prepareNonSpriteBuffer(entities)

	// Group entities by sprite image
	r.groupEntitiesBySprite(entities, batches)

	r.stats.BatchCount = len(batches)

	// Draw batched sprites
	for _, batch := range batches {
		r.drawBatch(batch)
	}

	// Draw non-sprite entities individually (colored rectangles)
	for _, entity := range r.nonSpriteBuffer {
		r.drawEntity(entity)
		r.stats.RenderedEntities++
	}
}

// prepareNonSpriteBuffer resets and prepares the buffer for non-sprite entities.
func (r *EbitenRenderSystem) prepareNonSpriteBuffer(entities []*Entity) {
	r.nonSpriteBuffer = r.nonSpriteBuffer[:0]
	if cap(r.nonSpriteBuffer) < len(entities) {
		r.nonSpriteBuffer = make([]*Entity, 0, len(entities))
	}
}

// groupEntitiesBySprite groups entities by their sprite image for batching.
// Uses cached GetSprite() getter for ~93x faster component access.
func (r *EbitenRenderSystem) groupEntitiesBySprite(entities []*Entity, batches map[*ebiten.Image][]*Entity) {
	for _, entity := range entities {
		sprite := entity.GetSprite()
		if sprite == nil || !sprite.Visible {
			continue
		}

		// Entities without sprite images need individual rendering
		if sprite.Image == nil {
			r.nonSpriteBuffer = append(r.nonSpriteBuffer, entity)
			continue
		}

		// Group by sprite image pointer (entities with same sprite are batched)
		batches[sprite.Image] = append(batches[sprite.Image], entity)
	}
}

// drawBatch renders a group of entities with the same sprite image using vertex batching.
// This combines multiple sprites into a single DrawTriangles call for better performance.
func (r *EbitenRenderSystem) drawBatch(entities []*Entity) {
	if len(entities) == 0 {
		return
	}

	batchSpriteImage := r.extractBatchSpriteImage(entities)
	if batchSpriteImage == nil {
		r.drawEntitiesIndividually(entities)
		return
	}

	vertices, indices := r.buildBatchGeometry(entities, batchSpriteImage)
	r.renderBatchGeometry(vertices, indices, batchSpriteImage)
}

// extractBatchSpriteImage retrieves the shared sprite image from the first entity in the batch.
// Uses cached GetSprite() getter for ~93x faster component access.
func (r *EbitenRenderSystem) extractBatchSpriteImage(entities []*Entity) *ebiten.Image {
	sprite := entities[0].GetSprite()
	if sprite == nil {
		return nil
	}
	return sprite.Image
}

// drawEntitiesIndividually renders each entity separately when batch rendering is not possible.
func (r *EbitenRenderSystem) drawEntitiesIndividually(entities []*Entity) {
	for _, entity := range entities {
		r.drawEntity(entity)
		r.stats.RenderedEntities++
	}
}

// buildBatchGeometry constructs vertex and index buffers for all entities in the batch.
func (r *EbitenRenderSystem) buildBatchGeometry(entities []*Entity, batchSpriteImage *ebiten.Image) ([]ebiten.Vertex, []uint16) {
	// Reuse vertex and index buffers to reduce allocations
	r.vertexBuffer = r.vertexBuffer[:0]
	r.indexBuffer = r.indexBuffer[:0]

	// Ensure capacity
	requiredVertices := len(entities) * 4
	requiredIndices := len(entities) * 6
	if cap(r.vertexBuffer) < requiredVertices {
		r.vertexBuffer = make([]ebiten.Vertex, 0, requiredVertices)
	}
	if cap(r.indexBuffer) < requiredIndices {
		r.indexBuffer = make([]uint16, 0, requiredIndices)
	}

	vertexIndex := uint16(0)

	for _, entity := range entities {
		r.stats.RenderedEntities++

		pos, sprite := r.validateBatchEntity(entity)
		if pos == nil || sprite == nil {
			continue
		}

		r.syncBatchSpriteState(entity, sprite)
		actualSpriteImage := r.selectSpriteImage(sprite)

		if !r.shouldRenderInBatch(entity, pos, sprite, actualSpriteImage, batchSpriteImage) {
			continue
		}

		// Use interpolated position for smooth rendering between simulation ticks
		screenX, screenY := r.interpolatePosition(pos)
		flashAlpha, tintR, tintG, tintB, tintA := r.extractVisualFeedback(entity)

		r.appendSpriteVertices(&r.vertexBuffer, &r.indexBuffer, sprite, screenX, screenY, flashAlpha, tintR, tintG, tintB, tintA, batchSpriteImage, &vertexIndex)
	}

	return r.vertexBuffer, r.indexBuffer
}

// validateBatchEntity checks if an entity has the required components for batch rendering.
// Uses cached GetPosition() and GetSprite() getters for ~93x faster access vs map lookup + type assertion.
func (r *EbitenRenderSystem) validateBatchEntity(entity *Entity) (*PositionComponent, *EbitenSprite) {
	// Use cached getter for position (~93x faster than GetComponent + type assertion)
	pos := entity.GetPosition()
	if pos == nil {
		return nil, nil
	}

	// Use cached getter for sprite (~93x faster than GetComponent + type assertion)
	sprite := entity.GetSprite()
	if sprite == nil || !sprite.Visible {
		return nil, nil
	}

	return pos, sprite
}

// syncBatchSpriteState synchronizes sprite state from animation and rotation components.
// Uses cached GetAnimation() getter for faster access.
func (r *EbitenRenderSystem) syncBatchSpriteState(entity *Entity, sprite *EbitenSprite) {
	// Use cached getter for animation (~90x faster than GetComponent + type assertion)
	if anim := entity.GetAnimation(); anim != nil {
		sprite.CurrentDirection = int(anim.GetFacing())
	}

	// Use cached getter for rotation (avoids map lookup + type assertion in hot path)
	if rotation := entity.GetRotation(); rotation != nil {
		sprite.Rotation = rotation.Angle
	}
}

// shouldRenderInBatch determines if an entity should be included in the current batch.
func (r *EbitenRenderSystem) shouldRenderInBatch(entity *Entity, pos *PositionComponent, sprite *EbitenSprite, actualSpriteImage, batchSpriteImage *ebiten.Image) bool {
	if actualSpriteImage == nil || actualSpriteImage != batchSpriteImage {
		r.drawEntity(entity)
		return false
	}

	if r.enableCulling && !r.spatialCullingUsed && !r.cameraSystem.IsVisible(pos.X, pos.Y, sprite.Width) {
		return false
	}

	return true
}

// appendSpriteVertices adds vertex and index data for a single sprite to the batch buffers.
func (r *EbitenRenderSystem) appendSpriteVertices(vertices *[]ebiten.Vertex, indices *[]uint16, sprite *EbitenSprite, screenX, screenY, flashAlpha, tintR, tintG, tintB, tintA float64, batchSpriteImage *ebiten.Image, vertexIndex *uint16) {
	cos, sin := r.calculateRotation(sprite.Rotation)
	corners := r.calculateSpriteCorners(sprite.Width, sprite.Height)
	texW, texH := r.extractTextureSize(batchSpriteImage)
	colorR, colorG, colorB, colorA := r.calculateVertexColors(flashAlpha, tintR, tintG, tintB, tintA)

	for i, corner := range corners {
		rotatedX := corner[0]*cos - corner[1]*sin
		rotatedY := corner[0]*sin + corner[1]*cos
		u, v := r.calculateTextureCoords(i, texW, texH)

		*vertices = append(*vertices, ebiten.Vertex{
			DstX:   float32(screenX) + rotatedX,
			DstY:   float32(screenY) + rotatedY,
			SrcX:   u,
			SrcY:   v,
			ColorR: colorR,
			ColorG: colorG,
			ColorB: colorB,
			ColorA: colorA,
		})
	}

	*indices = append(*indices,
		*vertexIndex+0, *vertexIndex+1, *vertexIndex+2,
		*vertexIndex+1, *vertexIndex+3, *vertexIndex+2,
	)
	*vertexIndex += 4
}

// calculateRotation computes sine and cosine values for sprite rotation.
func (r *EbitenRenderSystem) calculateRotation(rotation float64) (cos, sin float32) {
	if rotation != 0 {
		return float32(math.Cos(rotation)), float32(math.Sin(rotation))
	}
	return 1.0, 0.0
}

// calculateSpriteCorners returns the four corner positions for a sprite quad.
func (r *EbitenRenderSystem) calculateSpriteCorners(width, height float64) [4][2]float32 {
	halfW := float32(width / 2)
	halfH := float32(height / 2)
	return [4][2]float32{
		{-halfW, -halfH},
		{halfW, -halfH},
		{-halfW, halfH},
		{halfW, halfH},
	}
}

// extractTextureSize retrieves the dimensions of a texture image.
func (r *EbitenRenderSystem) extractTextureSize(image *ebiten.Image) (width, height float32) {
	bounds := image.Bounds()
	return float32(bounds.Dx()), float32(bounds.Dy())
}

// calculateVertexColors computes the final color values including visual feedback effects.
func (r *EbitenRenderSystem) calculateVertexColors(flashAlpha, tintR, tintG, tintB, tintA float64) (colorR, colorG, colorB, colorA float32) {
	return float32(tintR + flashAlpha), float32(tintG + flashAlpha), float32(tintB + flashAlpha), float32(tintA)
}

// calculateTextureCoords determines UV coordinates for a vertex based on its corner index.
func (r *EbitenRenderSystem) calculateTextureCoords(cornerIndex int, texWidth, texHeight float32) (u, v float32) {
	if cornerIndex == 1 || cornerIndex == 3 {
		u = texWidth
	}
	if cornerIndex == 2 || cornerIndex == 3 {
		v = texHeight
	}
	return u, v
}

// renderBatchGeometry submits the batch geometry to the GPU for rendering.
func (r *EbitenRenderSystem) renderBatchGeometry(vertices []ebiten.Vertex, indices []uint16, batchSpriteImage *ebiten.Image) {
	if len(vertices) > 0 && len(indices) > 0 {
		r.screen.DrawTriangles(vertices, indices, batchSpriteImage, &r.drawTrianglesOptions)
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
	camera := r.getValidCamera()
	if camera == nil {
		return entities // No camera, render all
	}

	viewportBounds := r.calculateViewportBounds(camera)
	visible := r.queryVisibleEntities(viewportBounds)
	visible = r.ensurePlayersIncluded(entities, visible)

	return visible
}

// getValidCamera retrieves and validates the active camera component.
func (r *EbitenRenderSystem) getValidCamera() *CameraComponent {
	cam := r.cameraSystem.activeCamera
	if cam == nil {
		return nil
	}

	camComp, ok := cam.GetComponent("camera")
	if !ok {
		return nil
	}

	camera, ok := camComp.(*CameraComponent)
	if !ok {
		return nil
	}

	return camera
}

// calculateViewportBounds computes the viewport bounds in world space with margin.
func (r *EbitenRenderSystem) calculateViewportBounds(camera *CameraComponent) Bounds {
	margin := 100.0 // Extra space to render sprites partially off-screen

	// BUG FIX: Use camera's actual position (camera.X, camera.Y) which includes
	// smoothing and bounds clamping, NOT the entity's position component.
	viewportWidth := float64(r.cameraSystem.ScreenWidth) / camera.Zoom
	viewportHeight := float64(r.cameraSystem.ScreenHeight) / camera.Zoom

	return Bounds{
		X:      camera.X - viewportWidth/2 - margin,
		Y:      camera.Y - viewportHeight/2 - margin,
		Width:  viewportWidth + margin*2,
		Height: viewportHeight + margin*2,
	}
}

// queryVisibleEntities retrieves entities within the viewport bounds.
func (r *EbitenRenderSystem) queryVisibleEntities(viewportBounds Bounds) []*Entity {
	r.viewportQueryBuffer = r.viewportQueryBuffer[:0]
	visible := r.spatialPartition.QueryBoundsInto(viewportBounds, r.viewportQueryBuffer)
	r.viewportQueryBuffer = visible // Update buffer reference in case it was reallocated
	return visible
}

// ensurePlayersIncluded adds player entities to visible list if not already included.
func (r *EbitenRenderSystem) ensurePlayersIncluded(allEntities, visible []*Entity) []*Entity {
	// Build O(1) lookup set from visible entities
	clear(r.visibleEntityIDs)
	for _, visibleEntity := range visible {
		r.visibleEntityIDs[visibleEntity.ID] = struct{}{}
	}

	r.playerBuffer = r.collectNonVisiblePlayers(allEntities)

	// Append player entities to visible list
	if len(r.playerBuffer) > 0 {
		visible = append(visible, r.playerBuffer...)
		r.viewportQueryBuffer = visible // Update in case append reallocated
	}

	return visible
}

// collectNonVisiblePlayers collects player entities not already in the visible set.
func (r *EbitenRenderSystem) collectNonVisiblePlayers(entities []*Entity) []*Entity {
	r.playerBuffer = r.playerBuffer[:0]

	for _, entity := range entities {
		if entity.HasComponent("input") {
			// O(1) map lookup instead of O(m) linear scan
			if _, alreadyVisible := r.visibleEntityIDs[entity.ID]; !alreadyVisible {
				r.playerBuffer = append(r.playerBuffer, entity)
			}
		}
	}

	return r.playerBuffer
}

// drawEntity renders a single entity.
func (r *EbitenRenderSystem) drawEntity(entity *Entity) {
	pos, sprite := r.validateEntityComponents(entity)
	if pos == nil || sprite == nil {
		return
	}

	r.syncSpriteState(entity, sprite)

	// Use interpolated position for smooth rendering between simulation ticks
	screenX, screenY := r.interpolatePosition(pos)

	if r.enableCulling && !r.spatialCullingUsed && !r.cameraSystem.IsVisible(pos.X, pos.Y, sprite.Width) {
		return
	}

	layerYOffset, layerAlpha := r.calculateLayerTransition(entity)
	flashAlpha, tintR, tintG, tintB, tintA := r.extractVisualFeedback(entity)
	tintA *= layerAlpha

	// Apply movement bob vertical offset if present
	if comp, ok := entity.GetComponent("movement_bob"); ok {
		if bob, ok := comp.(*MovementBobComponent); ok {
			layerYOffset += bob.OffsetY
		}
	}

	// Apply movement lean horizontal offset if present
	var layerXOffset float64
	if comp, ok := entity.GetComponent("movement_lean"); ok {
		if lean, ok := comp.(*MovementLeanComponent); ok {
			layerXOffset = lean.OffsetX
		}
	}

	spriteImage := r.selectSpriteImage(sprite)

	// Draw drop shadow beneath the entity sprite for top-down depth grounding
	r.drawDropShadow(entity, screenX+layerXOffset, screenY+layerYOffset)

	if spriteImage != nil {
		r.drawSpriteImage(spriteImage, sprite, screenX+layerXOffset, screenY, layerYOffset, flashAlpha, tintR, tintG, tintB, tintA)
	} else {
		r.drawFallbackRect(sprite, screenX+layerXOffset, screenY, layerYOffset, layerAlpha, flashAlpha)
	}

	r.drawHealthBar(entity, screenX, screenY, sprite.Width, sprite.Height)
}

// validateEntityComponents retrieves and validates position and sprite components.
// Uses cached GetPosition() and GetSprite() getters for ~93x faster access vs map lookup + type assertion.
func (r *EbitenRenderSystem) validateEntityComponents(entity *Entity) (*PositionComponent, *EbitenSprite) {
	// Use cached getter for position (~93x faster than GetComponent + type assertion)
	pos := entity.GetPosition()
	if pos == nil {
		return nil, nil
	}

	// Use cached getter for sprite (~93x faster than GetComponent + type assertion)
	sprite := entity.GetSprite()
	if sprite == nil || !sprite.Visible {
		return nil, nil
	}

	return pos, sprite
}

// syncSpriteState synchronizes sprite direction and rotation from entity components.
// Uses cached GetAnimation() and GetRotation() getters for faster access.
func (r *EbitenRenderSystem) syncSpriteState(entity *Entity, sprite *EbitenSprite) {
	// Use cached getter for animation (~90x faster than GetComponent + type assertion)
	if anim := entity.GetAnimation(); anim != nil {
		sprite.CurrentDirection = int(anim.GetFacing())
	}

	// Use cached getter for rotation (avoids map lookup + type assertion in hot path)
	if rotation := entity.GetRotation(); rotation != nil {
		sprite.Rotation = rotation.Angle
	}
}

// calculateLayerTransition computes depth offset and transparency for layer transitions.
// Uses cached GetLayer() getter for ~93x faster access vs generic GetComponent.
func (r *EbitenRenderSystem) calculateLayerTransition(entity *Entity) (yOffset, alpha float64) {
	alpha = 1.0
	layer := entity.GetLayer()
	if layer == nil || !layer.IsTransitioning() {
		return yOffset, alpha
	}

	const maxDepthOffset = 16.0
	depthOffset := layer.TransitionProgress * maxDepthOffset

	if layer.TargetLayer > layer.CurrentLayer {
		yOffset = -depthOffset
	} else {
		yOffset = depthOffset
	}

	if layer.TransitionProgress < 0.3 {
		alpha = 0.7 + (layer.TransitionProgress / 0.3 * 0.3)
	} else if layer.TransitionProgress > 0.7 {
		alpha = 1.0 - ((layer.TransitionProgress - 0.7) / 0.3 * 0.3)
	}

	return yOffset, alpha
}

// extractVisualFeedback retrieves flash and tint values from visual feedback component.
// Uses cached GetVisualFeedback() getter for ~93x faster access in render hot path.
func (r *EbitenRenderSystem) extractVisualFeedback(entity *Entity) (flashAlpha, tintR, tintG, tintB, tintA float64) {
	tintR, tintG, tintB, tintA = 1.0, 1.0, 1.0, 1.0

	feedback := entity.GetVisualFeedback()
	if feedback != nil {
		flashAlpha = feedback.GetFlashAlpha()
		tintR, tintG, tintB, tintA = feedback.TintR, feedback.TintG, feedback.TintB, feedback.TintA
	}

	// Multiply weather-driven tint (composes with status effect tints)
	// Uses cached getter for zero-overhead access (~93x faster than map lookup)
	if wt := entity.GetWeatherSpriteTint(); wt != nil {
		tintR *= wt.TintR
		tintG *= wt.TintG
		tintB *= wt.TintB
	}

	// Multiply creature genre-driven tint (composes with weather and status tints)
	// Uses cached getter for zero-overhead access (~93x faster than map lookup)
	if ct := entity.GetCreatureGenreTint(); ct != nil {
		tintR *= ct.TintR
		tintG *= ct.TintG
		tintB *= ct.TintB
	}

	return flashAlpha, tintR, tintG, tintB, tintA
}

// selectSpriteImage chooses the appropriate sprite image based on direction.
func (r *EbitenRenderSystem) selectSpriteImage(sprite *EbitenSprite) *ebiten.Image {
	if len(sprite.DirectionalImages) > 0 {
		if dirImg, exists := sprite.DirectionalImages[sprite.CurrentDirection]; exists && dirImg != nil {
			return dirImg
		}
		return sprite.Image
	}
	return sprite.Image
}

// drawSpriteImage renders a sprite image with visual effects applied.
// Uses pre-allocated DrawImageOptions to avoid per-call heap allocations.
func (r *EbitenRenderSystem) drawSpriteImage(img *ebiten.Image, sprite *EbitenSprite, screenX, screenY, layerYOffset, flashAlpha, tintR, tintG, tintB, tintA float64) {
	// Reset pre-allocated options to identity state
	r.drawImageOptions.GeoM.Reset()
	r.drawImageOptions.ColorScale.Reset()

	if flashAlpha > 0 || tintR != 1.0 || tintG != 1.0 || tintB != 1.0 || tintA != 1.0 {
		r.drawImageOptions.ColorScale.ScaleWithColor(color.RGBA{
			R: uint8((tintR + flashAlpha) * 255),
			G: uint8((tintG + flashAlpha) * 255),
			B: uint8((tintB + flashAlpha) * 255),
			A: uint8(tintA * 255),
		})
	}

	r.drawImageOptions.GeoM.Translate(-sprite.Width/2, -sprite.Height/2)
	r.drawImageOptions.GeoM.Rotate(sprite.Rotation)
	r.drawImageOptions.GeoM.Translate(screenX, screenY+layerYOffset)
	r.screen.DrawImage(img, &r.drawImageOptions)
}

// drawFallbackRect renders a colored rectangle when no sprite image exists.
// Pre-extracts RGBA bytes once via type assertion to avoid repeated col.RGBA() overhead.
func (r *EbitenRenderSystem) drawFallbackRect(sprite *EbitenSprite, screenX, screenY, layerYOffset, layerAlpha, flashAlpha float64) {
	col := sprite.Color
	// Safety check: default to opaque magenta if no color is set (makes missing colors obvious)
	if col == nil {
		col = color.RGBA{R: 255, G: 0, B: 255, A: 255}
	}

	// Extract RGBA bytes once to avoid repeated col.RGBA() → uint32 → uint8 roundtrips.
	var cr, cg, cb, ca uint8
	if rgba, ok := col.(color.RGBA); ok {
		cr, cg, cb, ca = rgba.R, rgba.G, rgba.B, rgba.A
	} else {
		rr, gg, bb, aa := col.RGBA()
		cr, cg, cb, ca = uint8(rr>>8), uint8(gg>>8), uint8(bb>>8), uint8(aa>>8)
	}

	if flashAlpha > 0 {
		cr = uint8((float64(cr) + flashAlpha*255) / 2)
		cg = uint8((float64(cg) + flashAlpha*255) / 2)
		cb = uint8((float64(cb) + flashAlpha*255) / 2)
	}

	if layerAlpha < 1.0 {
		ca = uint8(float64(ca) * layerAlpha)
	}

	r.drawRect(screenX-sprite.Width/2, screenY-sprite.Height/2+layerYOffset,
		sprite.Width, sprite.Height, color.RGBA{R: cr, G: cg, B: cb, A: ca})
}

// drawHealthBar renders a health bar above an entity if appropriate.
// GAP-013 REPAIR: Shows health status for enemies (when damaged) and bosses (always).
func (r *EbitenRenderSystem) drawHealthBar(entity *Entity, screenX, screenY, spriteWidth, spriteHeight float64) {
	// Safety check: ensure screen is available
	if r.screen == nil {
		return
	}

	// Validate entity has required components
	health, isBoss, shouldDraw := r.validateHealthBarEntity(entity)
	if !shouldDraw {
		return
	}

	// Only show health bar if: (1) damaged, or (2) is boss
	if health.Current >= health.Max && !isBoss {
		return
	}

	// Calculate health bar dimensions
	barX, barY, barWidth, barHeight := r.calculateHealthBarDimensions(screenX, screenY, spriteWidth, spriteHeight)

	// Draw background
	r.drawHealthBarBackground(barX, barY, barWidth, barHeight)

	// Calculate and draw health bar
	healthPercent := r.calculateHealthPercent(health)
	healthColor := r.getHealthBarColor(healthPercent)
	r.drawHealthBarForeground(barX, barY, barWidth, barHeight, healthPercent, healthColor)

	// Draw border
	r.drawHealthBarBorder(barX, barY, barWidth, barHeight)
}

// validateHealthBarEntity validates if entity should have health bar drawn.
func (r *EbitenRenderSystem) validateHealthBarEntity(entity *Entity) (*HealthComponent, bool, bool) {
	// Only draw health bars for entities with health component
	healthComp, hasHealth := entity.GetComponent("health")
	if !hasHealth {
		return nil, false, false
	}

	health, ok := healthComp.(*HealthComponent)
	if !ok {
		return nil, false, false
	}

	// Don't draw health bar for player (has HUD display)
	if entity.HasComponent("input") {
		return nil, false, false
	}

	// Check if entity is a boss (high attack indicates boss)
	isBoss := false
	if attackComp, ok := entity.GetComponent("attack"); ok {
		if attack, ok := attackComp.(*AttackComponent); ok {
			isBoss = attack.Damage > 20 // Boss threshold
		}
	}

	return health, isBoss, true
}

// calculateHealthBarDimensions calculates position and size of health bar.
func (r *EbitenRenderSystem) calculateHealthBarDimensions(screenX, screenY, spriteWidth, spriteHeight float64) (float64, float64, float64, float64) {
	barWidth := spriteWidth
	barHeight := 4.0
	barX := screenX - barWidth/2
	barY := screenY - spriteHeight/2 - barHeight - 5 // 5px above sprite
	return barX, barY, barWidth, barHeight
}

// drawHealthBarBackground draws the dark background of the health bar.
func (r *EbitenRenderSystem) drawHealthBarBackground(barX, barY, barWidth, barHeight float64) {
	bgColor := color.RGBA{40, 40, 40, 200}
	vector.DrawFilledRect(r.screen, float32(barX), float32(barY),
		float32(barWidth), float32(barHeight), bgColor, false)
}

// calculateHealthPercent calculates and clamps health percentage.
func (r *EbitenRenderSystem) calculateHealthPercent(health *HealthComponent) float64 {
	healthPercent := health.Current / health.Max
	if healthPercent < 0 {
		healthPercent = 0
	}
	if healthPercent > 1 {
		healthPercent = 1
	}
	return healthPercent
}

// getHealthBarColor determines health bar color based on percentage.
func (r *EbitenRenderSystem) getHealthBarColor(healthPercent float64) color.RGBA {
	if healthPercent > 0.6 {
		return color.RGBA{50, 200, 50, 255} // Green (healthy)
	} else if healthPercent > 0.3 {
		return color.RGBA{220, 220, 50, 255} // Yellow (wounded)
	}
	return color.RGBA{220, 50, 50, 255} // Red (critical)
}

// drawHealthBarForeground draws the colored health bar foreground.
func (r *EbitenRenderSystem) drawHealthBarForeground(barX, barY, barWidth, barHeight, healthPercent float64, healthColor color.RGBA) {
	healthBarWidth := barWidth * healthPercent
	vector.DrawFilledRect(r.screen, float32(barX), float32(barY),
		float32(healthBarWidth), float32(barHeight), healthColor, false)
}

// drawHealthBarBorder draws the border around the health bar.
func (r *EbitenRenderSystem) drawHealthBarBorder(barX, barY, barWidth, barHeight float64) {
	borderColor := color.RGBA{200, 200, 200, 255}
	vector.StrokeRect(r.screen, float32(barX), float32(barY),
		float32(barWidth), float32(barHeight), 1, borderColor, false)
}

// filterParticleEntities returns only entities with particle emitters,
// using a pre-allocated buffer to avoid allocations in the render hot path.
func (r *EbitenRenderSystem) filterParticleEntities(entities []*Entity) []*Entity {
	r.particleEntityBuffer = r.particleEntityBuffer[:0]
	for _, entity := range entities {
		if entity.GetParticleEmitter() != nil {
			r.particleEntityBuffer = append(r.particleEntityBuffer, entity)
		}
	}
	return r.particleEntityBuffer
}

// GAP-016 REPAIR: drawParticles renders all particle effects to the screen.
// Receives pre-filtered entities with particle emitters for efficient iteration.
func (r *EbitenRenderSystem) drawParticles(entities []*Entity) {
	// Safety check: ensure screen is available
	if r.screen == nil {
		return
	}

	for _, entity := range entities {
		emitter := entity.GetParticleEmitter()
		if emitter == nil {
			continue
		}

		// Render each particle system using zero-allocation visitor
		for _, system := range emitter.Systems {
			r.drawParticleSystem(system)
		}
	}
}

// drawParticleSystem renders a single particle system using a zero-allocation visitor pattern.
func (r *EbitenRenderSystem) drawParticleSystem(system *particles.ParticleSystem) {
	system.VisitAliveParticles(func(particle *particles.Particle) {
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
	})
}

// drawRect draws a filled rectangle at the given screen position.
// PERF V5: Removed defer recover() from hot path - nil check provides adequate safety.
func (r *EbitenRenderSystem) drawRect(x, y, width, height float64, col color.Color) {
	// Safety check: ensure screen is available
	if r.screen == nil {
		return
	}

	// Safety check: ensure color is not nil
	if col == nil {
		return
	}

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
	// Safety check: ensure screen is available
	if r.screen == nil {
		return
	}

	debugColor := color.RGBA{0, 255, 0, 128} // Semi-transparent green

	for _, entity := range entities {
		posComp, hasPos := entity.GetComponent("position")
		colliderComp, hasCollider := entity.GetComponent("collider")

		if !hasPos || !hasCollider {
			continue
		}

		pos, ok := posComp.(*PositionComponent)
		if !ok {
			continue
		}
		collider, ok := colliderComp.(*ColliderComponent)
		if !ok {
			continue
		}

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
// Optimized: Uses reusable buffers to eliminate per-frame allocations.
// Optimized: Caches Y positions during collection to avoid O(n log n) map lookups during sort.
func (r *EbitenRenderSystem) sortEntitiesByLayer(entities []*Entity) []*Entity {
	r.prepareSortBuffers(entities)
	r.collectEntitiesWithSprites(entities)
	r.sortCollectedEntities()
	return r.extractSortedEntities()
}

// prepareSortBuffers reuses and grows buffers as needed for sorting.
func (r *EbitenRenderSystem) prepareSortBuffers(entities []*Entity) {
	r.sortBuffer = r.sortBuffer[:0]
	r.sortCacheBuffer = r.sortCacheBuffer[:0]

	if cap(r.sortBuffer) < len(entities) {
		r.sortBuffer = make([]*Entity, 0, len(entities))
	}
	if cap(r.sortCacheBuffer) < len(entities) {
		r.sortCacheBuffer = make([]entitySprite, 0, len(entities))
	}
}

// collectEntitiesWithSprites gathers entities with sprites and caches their Y positions.
func (r *EbitenRenderSystem) collectEntitiesWithSprites(entities []*Entity) {
	for _, entity := range entities {
		if sprite := entity.GetSprite(); sprite != nil {
			yPos := 0.0
			if pos := entity.GetPosition(); pos != nil {
				yPos = pos.Y
			}
			r.sortCacheBuffer = append(r.sortCacheBuffer, entitySprite{
				entity: entity,
				sprite: sprite,
				layer:  sprite.Layer,
				yPos:   yPos,
			})
		}
	}
}

// sortCollectedEntities sorts cached entities by layer, Y position, and ID.
func (r *EbitenRenderSystem) sortCollectedEntities() {
	slices.SortFunc(r.sortCacheBuffer, func(a, b entitySprite) int {
		// Primary sort: by sprite layer
		if a.layer != b.layer {
			return cmp.Compare(a.layer, b.layer)
		}
		// Secondary sort: by cached Y position for depth sorting
		if a.yPos != b.yPos {
			return cmp.Compare(a.yPos, b.yPos)
		}
		// Tertiary sort: by entity ID for complete determinism
		return cmp.Compare(a.entity.ID, b.entity.ID)
	})
}

// extractSortedEntities extracts the sorted entity list from the cache buffer.
func (r *EbitenRenderSystem) extractSortedEntities() []*Entity {
	for _, es := range r.sortCacheBuffer {
		r.sortBuffer = append(r.sortBuffer, es.entity)
	}
	return r.sortBuffer
}

// SetShowColliders implements RenderingSystem interface.
func (r *EbitenRenderSystem) SetShowColliders(show bool) {
	r.ShowColliders = show
}

// SetShowGrid implements RenderingSystem interface.
func (r *EbitenRenderSystem) SetShowGrid(show bool) {
	r.ShowGrid = show
}

// SetAimPlayerEntity sets the player entity used to draw the aim direction indicator
// below the player sprite layer. Called from EbitenGame.SetPlayerEntity().
func (r *EbitenRenderSystem) SetAimPlayerEntity(entity *Entity) {
	r.aimPlayerEntity = entity
}

// drawAimIndicator draws the aim direction arrow originating from the player entity's
// screen-space center, rendered before entities so it appears behind (below) the player sprite.
// The arrow and arrowhead are drawn in black.
func (r *EbitenRenderSystem) drawAimIndicator() {
	if r.screen == nil || r.aimPlayerEntity == nil || r.cameraSystem == nil {
		return
	}

	// Get player aim component
	aimComp, ok := r.aimPlayerEntity.GetComponent("aim")
	if !ok {
		return
	}
	aim, ok := aimComp.(*AimComponent)
	if !ok {
		return
	}

	// Get player position in world coordinates
	pos := r.aimPlayerEntity.GetPosition()
	if pos == nil {
		return
	}

	// Convert player world position to screen coordinates (accounting for camera offset).
	// interpolatePosition returns the sprite center (drawSpriteImage translates by -Width/2, -Height/2
	// then +screenX, +screenY), so no additional offset is needed.
	centerX, centerY := r.interpolatePosition(pos)

	// Calculate endpoint 60 pixels away in aim direction
	dirX, dirY := aim.GetAimDirection()
	arrowLength := float32(60.0)

	cx := float32(centerX)
	cy := float32(centerY)
	endX := cx + float32(dirX)*arrowLength
	endY := cy + float32(dirY)*arrowLength

	// Draw aim line (semi-transparent black)
	vector.StrokeLine(r.screen, cx, cy, endX, endY, 2,
		color.RGBA{0, 0, 0, 128}, false)

	// Draw arrowhead
	arrowSize := float32(8.0)
	perpX := -float32(dirY) // Perpendicular vector
	perpY := float32(dirX)

	// Three points of the arrowhead triangle
	tipX := endX
	tipY := endY
	left1X := tipX - float32(dirX)*arrowSize + perpX*arrowSize*0.5
	left1Y := tipY - float32(dirY)*arrowSize + perpY*arrowSize*0.5
	left2X := tipX - float32(dirX)*arrowSize - perpX*arrowSize*0.5
	left2Y := tipY - float32(dirY)*arrowSize - perpY*arrowSize*0.5

	// Draw filled circle at tip and arrowhead lines (black)
	vector.DrawFilledCircle(r.screen, tipX, tipY, 3, color.RGBA{0, 0, 0, 180}, false)
	vector.StrokeLine(r.screen, left1X, left1Y, tipX, tipY, 2,
		color.RGBA{0, 0, 0, 180}, false)
	vector.StrokeLine(r.screen, left2X, left2Y, tipX, tipY, 2,
		color.RGBA{0, 0, 0, 180}, false)
}

// Compile-time interface check
var _ RenderingSystem = (*EbitenRenderSystem)(nil)
