// Package engine provides spatial partitioning for efficient entity queries.
// This file implements SpatialPartition using a grid-based structure for
// fast spatial lookups and collision detection.
package engine

import (
	"math"
	"sync"
)

// Bounds represents a rectangular area in 2D space.
type Bounds struct {
	X, Y          float64 // Top-left corner
	Width, Height float64
}

// Contains checks if a point is within the bounds.
func (b Bounds) Contains(x, y float64) bool {
	return x >= b.X && x < b.X+b.Width &&
		y >= b.Y && y < b.Y+b.Height
}

// Intersects checks if two bounds overlap.
func (b Bounds) Intersects(other Bounds) bool {
	return !(other.X >= b.X+b.Width ||
		other.X+other.Width <= b.X ||
		other.Y >= b.Y+b.Height ||
		other.Y+other.Height <= b.Y)
}

// Quadtree provides spatial partitioning for efficient entity queries.
// It divides 2D space into nested rectangles for O(log n) proximity searches.
type Quadtree struct {
	bounds   Bounds
	capacity int
	entities []*Entity
	divided  bool

	// Child quadrants (NW, NE, SW, SE)
	northwest *Quadtree
	northeast *Quadtree
	southwest *Quadtree
	southeast *Quadtree

	// Result buffer pool for zero-allocation queries
	resultPool sync.Pool
}

// NewQuadtree creates a new quadtree with the given bounds and capacity.
// Capacity determines how many entities can be stored before subdivision.
func NewQuadtree(bounds Bounds, capacity int) *Quadtree {
	qt := &Quadtree{
		bounds:   bounds,
		capacity: capacity,
		entities: make([]*Entity, 0, capacity),
		divided:  false,
	}
	qt.resultPool.New = func() interface{} {
		slice := make([]*Entity, 0, capacity)
		return &slice
	}
	return qt
}

// Insert adds an entity to the quadtree.
// Returns true if successful, false if the entity is outside bounds.
func (q *Quadtree) Insert(entity *Entity) bool {
	// Use cached position accessor for ~96x faster access vs GetComponent
	pos := entity.GetPosition()
	if pos == nil {
		return false
	}

	// Check if point is in bounds
	if !q.bounds.Contains(pos.X, pos.Y) {
		return false
	}

	// If we have capacity, add here
	if len(q.entities) < q.capacity {
		q.entities = append(q.entities, entity)
		return true
	}

	// Otherwise, subdivide and insert into child
	if !q.divided {
		q.subdivide()
	}

	// Try to insert into children
	if q.northwest.Insert(entity) {
		return true
	}
	if q.northeast.Insert(entity) {
		return true
	}
	if q.southwest.Insert(entity) {
		return true
	}
	if q.southeast.Insert(entity) {
		return true
	}

	// Shouldn't happen, but handle gracefully
	return false
}

// subdivide splits this quadrant into four children.
func (q *Quadtree) subdivide() {
	x := q.bounds.X
	y := q.bounds.Y
	w := q.bounds.Width / 2
	h := q.bounds.Height / 2

	q.northwest = NewQuadtree(Bounds{x, y, w, h}, q.capacity)
	q.northeast = NewQuadtree(Bounds{x + w, y, w, h}, q.capacity)
	q.southwest = NewQuadtree(Bounds{x, y + h, w, h}, q.capacity)
	q.southeast = NewQuadtree(Bounds{x + w, y + h, w, h}, q.capacity)

	q.divided = true
}

// Query returns all entities within the given bounds.
// Uses pooled result buffers to eliminate allocations in hot paths.
func (q *Quadtree) Query(queryBounds Bounds) []*Entity {
	resultPtr := q.resultPool.Get().(*[]*Entity)
	result := (*resultPtr)[:0]
	q.queryRecursive(queryBounds, &result)

	// Copy to new slice for caller (they own it)
	output := make([]*Entity, len(result))
	copy(output, result)

	// Return buffer to pool
	*resultPtr = result
	q.resultPool.Put(resultPtr)

	return output
}

// QueryInto appends entities within bounds to the provided buffer.
// This is a zero-allocation query method for hot paths that reuse buffers.
// The caller must pass a slice (typically with len 0 but existing capacity).
// Returns the (possibly reallocated) buffer with results appended.
func (q *Quadtree) QueryInto(queryBounds Bounds, buffer []*Entity) []*Entity {
	q.queryRecursive(queryBounds, &buffer)
	return buffer
}

// queryRecursive performs the actual recursive query.
func (q *Quadtree) queryRecursive(queryBounds Bounds, result *[]*Entity) {
	// If bounds don't intersect, nothing to do
	if !q.bounds.Intersects(queryBounds) {
		return
	}

	// Check entities at this level
	for _, entity := range q.entities {
		// Use cached position accessor for ~96x faster access vs GetComponent
		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		// BUG FIX: Check if entity intersects query bounds
		// Need to consider entity size, not just center point position

		// Get sprite size if available (default to 32x32 if no sprite)
		spriteWidth, spriteHeight := 32.0, 32.0
		if spriteComp, ok := entity.GetComponent("sprite"); ok {
			if sprite, ok := spriteComp.(interface {
				GetSize() (width, height float64)
			}); ok {
				spriteWidth, spriteHeight = sprite.GetSize()
			}
		}

		// Create entity bounds (centered on position)
		entityBounds := Bounds{
			X:      pos.X - spriteWidth/2,
			Y:      pos.Y - spriteHeight/2,
			Width:  spriteWidth,
			Height: spriteHeight,
		}

		// Check if entity bounds intersect query bounds
		if entityBounds.Intersects(queryBounds) {
			*result = append(*result, entity)
		}
	}

	// Recursively check children
	if q.divided {
		q.northwest.queryRecursive(queryBounds, result)
		q.northeast.queryRecursive(queryBounds, result)
		q.southwest.queryRecursive(queryBounds, result)
		q.southeast.queryRecursive(queryBounds, result)
	}
}

// QueryRadius returns all entities within a circular radius of a point.
func (q *Quadtree) QueryRadius(x, y, radius float64) []*Entity {
	// Query a square bounding box first
	queryBounds := Bounds{
		X:      x - radius,
		Y:      y - radius,
		Width:  radius * 2,
		Height: radius * 2,
	}

	// Get pooled buffer for candidates
	candidatesPtr := q.resultPool.Get().(*[]*Entity)
	candidates := (*candidatesPtr)[:0]
	q.queryRecursive(queryBounds, &candidates)

	// Get pooled buffer for results
	resultPtr := q.resultPool.Get().(*[]*Entity)
	result := (*resultPtr)[:0]
	radiusSq := radius * radius

	for _, entity := range candidates {
		// Use cached position accessor for ~96x faster access vs GetComponent
		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		dx := pos.X - x
		dy := pos.Y - y
		distSq := dx*dx + dy*dy

		if distSq <= radiusSq {
			result = append(result, entity)
		}
	}

	// Copy to new slice for caller
	output := make([]*Entity, len(result))
	copy(output, result)

	// Return buffers to pool
	*candidatesPtr = candidates
	q.resultPool.Put(candidatesPtr)
	*resultPtr = result
	q.resultPool.Put(resultPtr)

	return output
}

// Clear removes all entities from the quadtree.
func (q *Quadtree) Clear() {
	q.entities = q.entities[:0]
	q.divided = false
	q.northwest = nil
	q.northeast = nil
	q.southwest = nil
	q.southeast = nil
}

// Rebuild reconstructs the quadtree with current entities.
// This should be called periodically as entities move.
func (q *Quadtree) Rebuild(entities []*Entity) {
	q.Clear()
	for _, entity := range entities {
		q.Insert(entity)
	}
}

// Count returns the total number of entities in the tree.
func (q *Quadtree) Count() int {
	count := len(q.entities)
	if q.divided {
		count += q.northwest.Count()
		count += q.northeast.Count()
		count += q.southwest.Count()
		count += q.southeast.Count()
	}
	return count
}

// SpatialPartitionSystem maintains a quadtree for efficient spatial queries.
type SpatialPartitionSystem struct {
	quadtree     *Quadtree
	worldBounds  Bounds
	rebuildEvery int // Rebuild every N frames
	frameCount   int

	// Dirty tracking for lazy rebuilding
	isDirty          bool
	lastRebuildFrame int
	minRebuildFrames int // Minimum frames between rebuilds (e.g., 3 = 50ms at 60fps)

	// Statistics
	lastRebuildTime float64
	queryCount      int
	skippedRebuilds int
	forcedRebuilds  int
	lazyRebuilds    int
}

// NewSpatialPartitionSystem creates a new spatial partition system.
func NewSpatialPartitionSystem(worldWidth, worldHeight float64) *SpatialPartitionSystem {
	bounds := Bounds{
		X:      0,
		Y:      0,
		Width:  worldWidth,
		Height: worldHeight,
	}

	return &SpatialPartitionSystem{
		quadtree:         NewQuadtree(bounds, 16), // 16 entities per node (tuned for better performance)
		worldBounds:      bounds,
		rebuildEvery:     60, // Check for rebuild every 60 frames (1 second at 60fps)
		frameCount:       0,
		isDirty:          false,
		lastRebuildFrame: 0,
		minRebuildFrames: 3, // Minimum 3 frames (50ms at 60fps) between rebuilds
	}
}

// SetCapacity sets the quadtree capacity (entities per node before subdivision).
// Higher values reduce tree depth but increase query time per node.
// Recommended values: 8-32 depending on entity density.
func (s *SpatialPartitionSystem) SetCapacity(capacity int) {
	// Rebuild quadtree with new capacity
	s.quadtree = NewQuadtree(s.worldBounds, capacity)
	s.isDirty = true
}

// SetRebuildInterval sets how many frames to wait before checking for rebuild.
// Lower values provide more up-to-date spatial data but cost more CPU.
// Higher values reduce CPU but may have stale data.
// Recommended: 30-60 frames (0.5-1 second at 60fps).
func (s *SpatialPartitionSystem) SetRebuildInterval(frames int) {
	s.rebuildEvery = frames
}

// Update rebuilds the quadtree periodically with lazy rebuild optimization.
// Uses dirty tracking to skip rebuilds when entities haven't moved significantly.
func (s *SpatialPartitionSystem) Update(entities []*Entity, deltaTime float64) {
	s.frameCount++

	// Check if enough time has passed since last rebuild
	framesSinceRebuild := s.frameCount - s.lastRebuildFrame

	// Rebuild if:
	// 1. We've reached the rebuild interval, AND
	// 2. Enough frames have passed since last rebuild (rate limiting)
	// 3. OR we're marked as dirty (entities moved)
	shouldRebuild := false

	// CRITICAL FIX: Always rebuild periodically even if not dirty
	// This ensures new entities that spawned are added to the quadtree
	// The original logic only rebuilt if dirty, which meant stationary
	// entities that were newly spawned would never be added
	if framesSinceRebuild >= s.rebuildEvery {
		shouldRebuild = true
		if s.isDirty {
			s.lazyRebuilds++
		} else {
			s.forcedRebuilds++ // Periodic forced rebuild
		}
	}

	if shouldRebuild {
		s.quadtree.Rebuild(entities)
		s.lastRebuildFrame = s.frameCount
		s.isDirty = false // Clear dirty flag after rebuild
	}
} // MarkDirty marks the spatial partition as needing a rebuild.
// Should be called when entities move significantly.
func (s *SpatialPartitionSystem) MarkDirty() {
	s.isDirty = true
}

// IsDirty returns whether the spatial partition needs rebuilding.
func (s *SpatialPartitionSystem) IsDirty() bool {
	return s.isDirty
}

// Rebuild forces an immediate rebuild of the spatial partition with the given entities.
// This bypasses the normal frame-based rebuild logic and is useful for tests or
// when you need to ensure the spatial partition is up-to-date immediately.
func (s *SpatialPartitionSystem) Rebuild(entities []*Entity) {
	s.quadtree.Rebuild(entities)
	s.isDirty = false
	s.lastRebuildFrame = s.frameCount
}

// QueryRadius returns entities within radius of a point.
func (s *SpatialPartitionSystem) QueryRadius(x, y, radius float64) []*Entity {
	s.queryCount++
	return s.quadtree.QueryRadius(x, y, radius)
}

// QueryBounds returns entities within a rectangular area.
func (s *SpatialPartitionSystem) QueryBounds(bounds Bounds) []*Entity {
	s.queryCount++
	return s.quadtree.Query(bounds)
}

// QueryBoundsInto appends entities within bounds to the provided buffer.
// This is a zero-allocation query method for hot paths that reuse buffers.
// Returns the (possibly reallocated) buffer with results appended.
func (s *SpatialPartitionSystem) QueryBoundsInto(bounds Bounds, buffer []*Entity) []*Entity {
	s.queryCount++
	return s.quadtree.QueryInto(bounds, buffer)
}

// GetStatistics returns performance statistics.
func (s *SpatialPartitionSystem) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"entity_count":      s.quadtree.Count(),
		"last_rebuild_time": s.lastRebuildTime,
		"query_count":       s.queryCount,
		"frame_count":       s.frameCount,
		"is_dirty":          s.isDirty,
		"skipped_rebuilds":  s.skippedRebuilds,
		"forced_rebuilds":   s.forcedRebuilds,
		"lazy_rebuilds":     s.lazyRebuilds,
	}
}

// Distance calculates the Euclidean distance between two points.
func Distance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

// DistanceSquared calculates the squared Euclidean distance (faster, no sqrt).
func DistanceSquared(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return dx*dx + dy*dy
}
