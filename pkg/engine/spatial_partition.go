package engine

import (
	"math"
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
}

// NewQuadtree creates a new quadtree with the given bounds and capacity.
// Capacity determines how many entities can be stored before subdivision.
func NewQuadtree(bounds Bounds, capacity int) *Quadtree {
	return &Quadtree{
		bounds:   bounds,
		capacity: capacity,
		entities: make([]*Entity, 0, capacity),
		divided:  false,
	}
}

// Insert adds an entity to the quadtree.
// Returns true if successful, false if the entity is outside bounds.
func (q *Quadtree) Insert(entity *Entity) bool {
	// Get entity position
	posComp, ok := entity.GetComponent("position")
	if !ok {
		return false
	}
	pos := posComp.(*PositionComponent)

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
func (q *Quadtree) Query(queryBounds Bounds) []*Entity {
	result := make([]*Entity, 0)
	q.queryRecursive(queryBounds, &result)
	return result
}

// queryRecursive performs the actual recursive query.
func (q *Quadtree) queryRecursive(queryBounds Bounds, result *[]*Entity) {
	// If bounds don't intersect, nothing to do
	if !q.bounds.Intersects(queryBounds) {
		return
	}

	// Check entities at this level
	for _, entity := range q.entities {
		posComp, ok := entity.GetComponent("position")
		if !ok {
			continue
		}
		pos := posComp.(*PositionComponent)

		if queryBounds.Contains(pos.X, pos.Y) {
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

	candidates := q.Query(queryBounds)

	// Filter by actual circular distance
	result := make([]*Entity, 0, len(candidates))
	radiusSq := radius * radius

	for _, entity := range candidates {
		posComp, ok := entity.GetComponent("position")
		if !ok {
			continue
		}
		pos := posComp.(*PositionComponent)

		dx := pos.X - x
		dy := pos.Y - y
		distSq := dx*dx + dy*dy

		if distSq <= radiusSq {
			result = append(result, entity)
		}
	}

	return result
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

	// Statistics
	lastRebuildTime float64
	queryCount      int
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
		quadtree:     NewQuadtree(bounds, 8), // 8 entities per node
		worldBounds:  bounds,
		rebuildEvery: 60, // Rebuild every 60 frames (1 second at 60fps)
		frameCount:   0,
	}
}

// Update rebuilds the quadtree periodically.
func (s *SpatialPartitionSystem) Update(entities []*Entity, deltaTime float64) {
	s.frameCount++

	// Rebuild periodically to account for entity movement
	if s.frameCount >= s.rebuildEvery {
		start := s.lastRebuildTime
		s.quadtree.Rebuild(entities)
		s.lastRebuildTime = deltaTime - start
		s.frameCount = 0
	}
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

// GetStatistics returns performance statistics.
func (s *SpatialPartitionSystem) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"entity_count":      s.quadtree.Count(),
		"last_rebuild_time": s.lastRebuildTime,
		"query_count":       s.queryCount,
		"frame_count":       s.frameCount,
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
