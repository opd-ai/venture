package engine

import (
	"sync"

	"github.com/sirupsen/logrus"
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

	// Logger for structured logging
	logger *logrus.Entry
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
	logger := logrus.New()
	logger.SetReportCaller(true)

	vo := &ViewportOptimizer{
		tileSize:    32.0,
		marginTiles: 1, // 1-tile margin as per Phase 44
		logger:      logger.WithField("system_name", "viewport_optimizer"),
	}

	vo.logger.WithFields(logrus.Fields{
		"tile_size":    vo.tileSize,
		"margin_tiles": vo.marginTiles,
	}).Debug("ViewportOptimizer created with default settings")

	return vo
}

// SetTileSize configures tile size for margin calculation.
func (v *ViewportOptimizer) SetTileSize(size float64) {
	v.mu.Lock()
	defer v.mu.Unlock()

	oldSize := v.tileSize
	v.tileSize = size

	v.logger.WithFields(logrus.Fields{
		"old_tile_size": oldSize,
		"new_tile_size": size,
	}).Debug("Tile size updated")
}

// SetMarginTiles configures margin in tiles.
func (v *ViewportOptimizer) SetMarginTiles(tiles int) {
	v.mu.Lock()
	defer v.mu.Unlock()

	oldMargin := v.marginTiles
	v.marginTiles = tiles

	v.logger.WithFields(logrus.Fields{
		"old_margin_tiles": oldMargin,
		"new_margin_tiles": tiles,
	}).Debug("Margin tiles updated")
}

// CalculateViewportBounds computes viewport bounds in world space with margin.
// Returns Bounds struct suitable for quadtree queries.
func (v *ViewportOptimizer) CalculateViewportBounds(cameraX, cameraY, viewportWidth, viewportHeight, zoom float64) Bounds {
	v.mu.RLock()
	defer v.mu.RUnlock()

	v.logger.WithFields(logrus.Fields{
		"camera_x":        cameraX,
		"camera_y":        cameraY,
		"viewport_width":  viewportWidth,
		"viewport_height": viewportHeight,
		"zoom":            zoom,
	}).Debug("Calculating viewport bounds")

	bounds := v.calculateViewportBoundsUnlocked(cameraX, cameraY, viewportWidth, viewportHeight, zoom)

	v.logger.WithFields(logrus.Fields{
		"bounds_x":      bounds.X,
		"bounds_y":      bounds.Y,
		"bounds_width":  bounds.Width,
		"bounds_height": bounds.Height,
	}).Debug("Viewport bounds calculated")

	return bounds
}

// calculateViewportBoundsUnlocked is the unlocked version for internal use.
func (v *ViewportOptimizer) calculateViewportBoundsUnlocked(cameraX, cameraY, viewportWidth, viewportHeight, zoom float64) Bounds {
	margin := v.tileSize * float64(v.marginTiles)

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
	v.logger.WithFields(logrus.Fields{
		"entity_x":      entityX,
		"entity_y":      entityY,
		"entity_width":  entityWidth,
		"entity_height": entityHeight,
	}).Debug("Performing frustum cull check")

	// Create entity bounds (centered on position)
	entityBounds := Bounds{
		X:      entityX - entityWidth/2,
		Y:      entityY - entityHeight/2,
		Width:  entityWidth,
		Height: entityHeight,
	}

	// Check intersection with viewport
	result := entityBounds.Intersects(viewportBounds)

	v.logger.WithFields(logrus.Fields{
		"entity_bounds_x":      entityBounds.X,
		"entity_bounds_y":      entityBounds.Y,
		"entity_bounds_width":  entityBounds.Width,
		"entity_bounds_height": entityBounds.Height,
		"intersects":           result,
	}).Debug("Frustum cull check completed")

	return result
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

	v.logger.WithFields(logrus.Fields{
		"total_entities":  len(allEntities),
		"screen_width":    screenWidth,
		"screen_height":   screenHeight,
		"camera_provided": camera != nil,
		"spatial_enabled": spatialPartition != nil,
	}).Debug("Starting visible set optimization")

	// Reset stats
	v.stats = ViewportStats{
		TotalEntities: len(allEntities),
	}

	if camera == nil || spatialPartition == nil {
		v.logger.WithFields(logrus.Fields{
			"camera_nil":            camera == nil,
			"spatial_partition_nil": spatialPartition == nil,
		}).Warn("Camera or spatial partition not available, returning all entities")

		v.stats.VisibleEntities = len(allEntities)
		return allEntities
	}

	// Calculate viewport bounds
	viewportBounds := v.calculateViewportBoundsUnlocked(
		camera.X, camera.Y,
		float64(screenWidth), float64(screenHeight),
		camera.Zoom,
	)

	v.logger.WithFields(logrus.Fields{
		"camera_x":        camera.X,
		"camera_y":        camera.Y,
		"camera_zoom":     camera.Zoom,
		"viewport_bounds": viewportBounds,
	}).Debug("Viewport bounds calculated for optimization")

	// Query spatial partition
	visible := spatialPartition.QueryBounds(viewportBounds)

	v.logger.WithFields(logrus.Fields{
		"spatial_query_results": len(visible),
	}).Debug("Spatial partition query completed")

	// Always include player entities (input component)
	playerEntities := v.extractPlayerEntities(allEntities, visible)
	if len(playerEntities) > 0 {
		visible = append(visible, playerEntities...)
		v.logger.WithFields(logrus.Fields{
			"player_entities_added": len(playerEntities),
		}).Debug("Added player entities to visible set")
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

	v.logger.WithFields(logrus.Fields{
		"core_viewport_bounds": coreViewportBounds,
	}).Debug("Calculating off-screen rendered count")

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

	v.logger.WithFields(logrus.Fields{
		"total_entities":       v.stats.TotalEntities,
		"visible_entities":     v.stats.VisibleEntities,
		"culled_entities":      v.stats.CulledEntities,
		"offscreen_rendered":   v.stats.OffScreenRendered,
		"culling_efficiency":   v.CullingEfficiency(),
		"offscreen_percentage": v.OffScreenPercentage(),
	}).Info("Visible set optimization completed")

	return visible
}

// extractPlayerEntities finds player entities not already in visible set.
func (v *ViewportOptimizer) extractPlayerEntities(allEntities, visible []*Entity) []*Entity {
	v.logger.WithFields(logrus.Fields{
		"total_entities":   len(allEntities),
		"visible_entities": len(visible),
	}).Debug("Extracting player entities")

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
			v.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
			}).Debug("Added player entity not in visible set")
		}
	}

	v.logger.WithFields(logrus.Fields{
		"player_entities_count": len(playerEntities),
	}).Debug("Player entity extraction completed")

	return playerEntities
}

// Stats returns current viewport optimization statistics.
func (v *ViewportOptimizer) Stats() ViewportStats {
	v.mu.RLock()
	defer v.mu.RUnlock()

	v.logger.WithFields(logrus.Fields{
		"total_entities":     v.stats.TotalEntities,
		"visible_entities":   v.stats.VisibleEntities,
		"culled_entities":    v.stats.CulledEntities,
		"offscreen_rendered": v.stats.OffScreenRendered,
	}).Debug("Returning viewport statistics")

	return v.stats
}

// OffScreenPercentage returns percentage of entities rendered off-screen.
func (v *ViewportOptimizer) OffScreenPercentage() float64 {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.stats.VisibleEntities == 0 {
		v.logger.Debug("No visible entities, returning 0% off-screen percentage")
		return 0.0
	}

	percentage := float64(v.stats.OffScreenRendered) / float64(v.stats.VisibleEntities) * 100.0

	v.logger.WithFields(logrus.Fields{
		"offscreen_rendered": v.stats.OffScreenRendered,
		"visible_entities":   v.stats.VisibleEntities,
		"percentage":         percentage,
	}).Debug("Calculated off-screen percentage")

	return percentage
}

// CullingEfficiency returns culling efficiency (0.0 to 1.0).
// Higher is better. 1.0 = all off-screen entities culled.
func (v *ViewportOptimizer) CullingEfficiency() float64 {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.stats.TotalEntities == 0 {
		v.logger.Debug("No total entities, returning 1.0 culling efficiency")
		return 1.0
	}

	efficiency := float64(v.stats.CulledEntities) / float64(v.stats.TotalEntities)

	v.logger.WithFields(logrus.Fields{
		"culled_entities": v.stats.CulledEntities,
		"total_entities":  v.stats.TotalEntities,
		"efficiency":      efficiency,
	}).Debug("Calculated culling efficiency")

	return efficiency
}

// ValidateMetrics checks if optimization meets Phase 44 targets.
// Target: <5% entities rendered off-screen.
func (v *ViewportOptimizer) ValidateMetrics() bool {
	offscreenPct := v.OffScreenPercentage()
	valid := offscreenPct < 5.0

	v.logger.WithFields(logrus.Fields{
		"offscreen_percentage": offscreenPct,
		"target_threshold":     5.0,
		"metrics_valid":        valid,
	}).Info("Validated optimization metrics against Phase 44 targets")

	if !valid {
		v.logger.WithFields(logrus.Fields{
			"offscreen_percentage": offscreenPct,
			"threshold_exceeded":   offscreenPct - 5.0,
		}).Warn("Off-screen percentage exceeds Phase 44 target of 5%")
	}

	return valid
}
