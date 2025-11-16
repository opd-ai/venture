package engine

import (
	"sync"
)

// ViewportOptimizer provides enhanced viewport culling for larger display resolutions.
// Phase 44: Optimizes culling for 1920x1080 (2.25x larger than 800x600).
type ViewportOptimizer struct {
	mu sync.RWMutex

	// Tile size for margin calculation (default 32px)
	tileSize float64

	// Margin in tiles (default 1 tile = 32px at 1920x1080)
	marginTiles int

	// Statistics
	stats ViewportStats
}

// ViewportStats tracks viewport optimization metrics.
type ViewportStats struct {
	TotalEntities     int
	VisibleEntities   int
	CulledEntities    int
	OffScreenRendered int // Entities rendered slightly off-screen (within margin)
	QueryTimeMs       float64
}

// NewViewportOptimizer creates viewport optimizer with default settings.
func NewViewportOptimizer() *ViewportOptimizer {
	return &ViewportOptimizer{
		tileSize:    32.0,
		marginTiles: 1, // 1-tile margin as per Phase 44
	}
}

// SetTileSize configures tile size for margin calculation.
func (v *ViewportOptimizer) SetTileSize(size float64) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.tileSize = size
}

// SetMarginTiles configures margin in tiles.
func (v *ViewportOptimizer) SetMarginTiles(tiles int) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.marginTiles = tiles
}

// CalculateViewportBounds computes viewport bounds in world space with margin.
// Returns Bounds struct suitable for quadtree queries.
func (v *ViewportOptimizer) CalculateViewportBounds(cameraX, cameraY, viewportWidth, viewportHeight, zoom float64) Bounds {
	v.mu.RLock()
	margin := v.tileSize * float64(v.marginTiles)
	v.mu.RUnlock()

	// Scale viewport by zoom (higher zoom = smaller world area visible)
	worldWidth := viewportWidth / zoom
	worldHeight := viewportHeight / zoom

	return Bounds{
		X:      cameraX - worldWidth/2 - margin,
		Y:      cameraY - worldHeight/2 - margin,
		Width:  worldWidth + margin*2,
		Height: worldHeight + margin*2,
	}
}

// FrustumCull performs frustum culling check on entity.
// Returns true if entity is within frustum (should be rendered).
func (v *ViewportOptimizer) FrustumCull(entityX, entityY, entityWidth, entityHeight float64, viewportBounds Bounds) bool {
	// Create entity bounds (centered on position)
	entityBounds := Bounds{
		X:      entityX - entityWidth/2,
		Y:      entityY - entityHeight/2,
		Width:  entityWidth,
		Height: entityHeight,
	}

	// Check intersection with viewport
	return entityBounds.Intersects(viewportBounds)
}

// OptimizeVisibleSet filters entities to visible set using spatial partition.
// This is the main optimization entry point.
func (v *ViewportOptimizer) OptimizeVisibleSet(
	camera *CameraComponent,
	screenWidth, screenHeight int,
	spatialPartition *SpatialPartitionSystem,
	allEntities []*Entity,
) []*Entity {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Reset stats
	v.stats = ViewportStats{
		TotalEntities: len(allEntities),
	}

	if camera == nil || spatialPartition == nil {
		v.stats.VisibleEntities = len(allEntities)
		return allEntities
	}

	// Calculate viewport bounds
	viewportBounds := v.CalculateViewportBounds(
		camera.X, camera.Y,
		float64(screenWidth), float64(screenHeight),
		camera.Zoom,
	)

	// Query spatial partition
	visible := spatialPartition.QueryBounds(viewportBounds)

	// Always include player entities (input component)
	playerEntities := v.extractPlayerEntities(allEntities, visible)
	if len(playerEntities) > 0 {
		visible = append(visible, playerEntities...)
	}

	// Update stats
	v.stats.VisibleEntities = len(visible)
	v.stats.CulledEntities = v.stats.TotalEntities - v.stats.VisibleEntities

	// Count off-screen entities (within margin but not in core viewport)
	coreViewportBounds := Bounds{
		X:      camera.X - float64(screenWidth)/(2*camera.Zoom),
		Y:      camera.Y - float64(screenHeight)/(2*camera.Zoom),
		Width:  float64(screenWidth) / camera.Zoom,
		Height: float64(screenHeight) / camera.Zoom,
	}

	for _, entity := range visible {
		posComp, ok := entity.GetComponent("position")
		if !ok {
			continue
		}
		pos, ok := posComp.(*PositionComponent)
		if !ok {
			continue
		}

		// Get sprite size
		width, height := 32.0, 32.0
		if spriteComp, ok := entity.GetComponent("sprite"); ok {
			if sprite, ok := spriteComp.(interface{ GetSize() (float64, float64) }); ok {
				width, height = sprite.GetSize()
			}
		}

		// Check if entity is in margin (not in core viewport)
		entityBounds := Bounds{
			X:      pos.X - width/2,
			Y:      pos.Y - height/2,
			Width:  width,
			Height: height,
		}

		if !entityBounds.Intersects(coreViewportBounds) {
			v.stats.OffScreenRendered++
		}
	}

	return visible
}

// extractPlayerEntities finds player entities not already in visible set.
func (v *ViewportOptimizer) extractPlayerEntities(allEntities, visible []*Entity) []*Entity {
	playerEntities := make([]*Entity, 0, 4)

	for _, entity := range allEntities {
		if !entity.HasComponent("input") {
			continue
		}

		// Check if already visible
		alreadyVisible := false
		for _, visEntity := range visible {
			if visEntity.ID == entity.ID {
				alreadyVisible = true
				break
			}
		}

		if !alreadyVisible {
			playerEntities = append(playerEntities, entity)
		}
	}

	return playerEntities
}

// Stats returns current viewport optimization statistics.
func (v *ViewportOptimizer) Stats() ViewportStats {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.stats
}

// OffScreenPercentage returns percentage of entities rendered off-screen.
func (v *ViewportOptimizer) OffScreenPercentage() float64 {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.stats.VisibleEntities == 0 {
		return 0.0
	}
	return float64(v.stats.OffScreenRendered) / float64(v.stats.VisibleEntities) * 100.0
}

// CullingEfficiency returns culling efficiency (0.0 to 1.0).
// Higher is better. 1.0 = all off-screen entities culled.
func (v *ViewportOptimizer) CullingEfficiency() float64 {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.stats.TotalEntities == 0 {
		return 1.0
	}
	return float64(v.stats.CulledEntities) / float64(v.stats.TotalEntities)
}

// ValidateMetrics checks if optimization meets Phase 44 targets.
// Target: <5% entities rendered off-screen.
func (v *ViewportOptimizer) ValidateMetrics() bool {
	return v.OffScreenPercentage() < 5.0
}
