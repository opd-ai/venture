package engine

import "math"

// CollisionSystem handles collision detection and resolution.
// Uses spatial partitioning (grid-based) for efficient broad-phase detection.
type CollisionSystem struct {
	// Grid cell size for spatial partitioning
	CellSize float64

	// Spatial grid for broad-phase collision detection
	grid map[int]map[int][]*Entity

	// Collision callbacks
	onCollision func(e1, e2 *Entity)
}

// NewCollisionSystem creates a new collision system.
func NewCollisionSystem(cellSize float64) *CollisionSystem {
	return &CollisionSystem{
		CellSize: cellSize,
		grid:     make(map[int]map[int][]*Entity),
	}
}

// SetCollisionCallback sets a function to be called when entities collide.
func (s *CollisionSystem) SetCollisionCallback(callback func(e1, e2 *Entity)) {
	s.onCollision = callback
}

// Update detects and resolves collisions between entities.
func (s *CollisionSystem) Update(entities []*Entity, deltaTime float64) {
	// Clear the grid
	s.grid = make(map[int]map[int][]*Entity)

	// Collect entities with colliders
	collidableEntities := make([]*Entity, 0)
	for _, entity := range entities {
		if entity.HasComponent("collider") && entity.HasComponent("position") {
			collidableEntities = append(collidableEntities, entity)
		}
	}

	// Build spatial grid (broad phase)
	for _, entity := range collidableEntities {
		s.addToGrid(entity)
	}

	// Check collisions (narrow phase)
	checked := make(map[uint64]map[uint64]bool)

	for _, entity := range collidableEntities {
		posComp, _ := entity.GetComponent("position")
		colliderComp, _ := entity.GetComponent("collider")

		pos := posComp.(*PositionComponent)
		collider := colliderComp.(*ColliderComponent)

		// Get potential collision candidates from nearby cells
		candidates := s.getNearbyEntities(entity)

		for _, other := range candidates {
			// Skip self
			if entity.ID == other.ID {
				continue
			}

			// Skip if already checked this pair
			if checked[entity.ID] != nil && checked[entity.ID][other.ID] {
				continue
			}

			// Mark as checked
			if checked[entity.ID] == nil {
				checked[entity.ID] = make(map[uint64]bool)
			}
			if checked[other.ID] == nil {
				checked[other.ID] = make(map[uint64]bool)
			}
			checked[entity.ID][other.ID] = true
			checked[other.ID][entity.ID] = true

			// Get other entity components
			otherPosComp, _ := other.GetComponent("position")
			otherColliderComp, _ := other.GetComponent("collider")

			otherPos := otherPosComp.(*PositionComponent)
			otherCollider := otherColliderComp.(*ColliderComponent)

			// Check layer compatibility (0 = all layers)
			if collider.Layer != 0 && otherCollider.Layer != 0 && collider.Layer != otherCollider.Layer {
				continue
			}

			// Check intersection
			if collider.Intersects(pos.X, pos.Y, otherCollider, otherPos.X, otherPos.Y) {
				// Call collision callback if set
				if s.onCollision != nil {
					s.onCollision(entity, other)
				}

				// Resolve collision if both are solid
				if collider.Solid && otherCollider.Solid && !collider.IsTrigger && !otherCollider.IsTrigger {
					s.resolveCollision(entity, other)
				}
			}
		}
	}
}

// addToGrid adds an entity to the spatial grid.
func (s *CollisionSystem) addToGrid(entity *Entity) {
	posComp, _ := entity.GetComponent("position")
	colliderComp, _ := entity.GetComponent("collider")

	pos := posComp.(*PositionComponent)
	collider := colliderComp.(*ColliderComponent)

	// Get bounding box
	minX, minY, maxX, maxY := collider.GetBounds(pos.X, pos.Y)

	// Calculate grid cells this entity occupies
	minCellX := int(math.Floor(minX / s.CellSize))
	minCellY := int(math.Floor(minY / s.CellSize))
	maxCellX := int(math.Floor(maxX / s.CellSize))
	maxCellY := int(math.Floor(maxY / s.CellSize))

	// Add to all occupied cells
	for x := minCellX; x <= maxCellX; x++ {
		for y := minCellY; y <= maxCellY; y++ {
			if s.grid[x] == nil {
				s.grid[x] = make(map[int][]*Entity)
			}
			s.grid[x][y] = append(s.grid[x][y], entity)
		}
	}
}

// getNearbyEntities returns entities in the same or adjacent grid cells.
func (s *CollisionSystem) getNearbyEntities(entity *Entity) []*Entity {
	posComp, _ := entity.GetComponent("position")
	colliderComp, _ := entity.GetComponent("collider")

	pos := posComp.(*PositionComponent)
	collider := colliderComp.(*ColliderComponent)

	minX, minY, maxX, maxY := collider.GetBounds(pos.X, pos.Y)

	// Calculate grid cells
	minCellX := int(math.Floor(minX / s.CellSize))
	minCellY := int(math.Floor(minY / s.CellSize))
	maxCellX := int(math.Floor(maxX / s.CellSize))
	maxCellY := int(math.Floor(maxY / s.CellSize))

	// Collect unique entities from cells
	seen := make(map[uint64]bool)
	result := make([]*Entity, 0)

	for x := minCellX; x <= maxCellX; x++ {
		for y := minCellY; y <= maxCellY; y++ {
			if s.grid[x] != nil && s.grid[x][y] != nil {
				for _, e := range s.grid[x][y] {
					if !seen[e.ID] {
						seen[e.ID] = true
						result = append(result, e)
					}
				}
			}
		}
	}

	return result
}

// resolveCollision separates two colliding entities.
func (s *CollisionSystem) resolveCollision(e1, e2 *Entity) {
	pos1Comp, _ := e1.GetComponent("position")
	pos2Comp, _ := e2.GetComponent("position")
	collider1Comp, _ := e1.GetComponent("collider")
	collider2Comp, _ := e2.GetComponent("collider")

	pos1 := pos1Comp.(*PositionComponent)
	pos2 := pos2Comp.(*PositionComponent)
	collider1 := collider1Comp.(*ColliderComponent)
	collider2 := collider2Comp.(*ColliderComponent)

	// Get bounding boxes
	min1X, min1Y, max1X, max1Y := collider1.GetBounds(pos1.X, pos1.Y)
	min2X, min2Y, max2X, max2Y := collider2.GetBounds(pos2.X, pos2.Y)

	// Calculate overlap in each axis
	overlapX := math.Min(max1X-min2X, max2X-min1X)
	overlapY := math.Min(max1Y-min2Y, max2Y-min1Y)

	// Separate along the axis with minimum overlap
	if overlapX < overlapY {
		// Separate horizontally
		if pos1.X < pos2.X {
			pos1.X -= overlapX / 2
			pos2.X += overlapX / 2
		} else {
			pos1.X += overlapX / 2
			pos2.X -= overlapX / 2
		}

		// Stop horizontal velocity
		if e1.HasComponent("velocity") {
			vel1, _ := e1.GetComponent("velocity")
			vel1.(*VelocityComponent).VX = 0
		}
		if e2.HasComponent("velocity") {
			vel2, _ := e2.GetComponent("velocity")
			vel2.(*VelocityComponent).VX = 0
		}
	} else {
		// Separate vertically
		if pos1.Y < pos2.Y {
			pos1.Y -= overlapY / 2
			pos2.Y += overlapY / 2
		} else {
			pos1.Y += overlapY / 2
			pos2.Y -= overlapY / 2
		}

		// Stop vertical velocity
		if e1.HasComponent("velocity") {
			vel1, _ := e1.GetComponent("velocity")
			vel1.(*VelocityComponent).VY = 0
		}
		if e2.HasComponent("velocity") {
			vel2, _ := e2.GetComponent("velocity")
			vel2.(*VelocityComponent).VY = 0
		}
	}
}

// CheckCollision checks if two entities are colliding.
func CheckCollision(e1, e2 *Entity) bool {
	if !e1.HasComponent("position") || !e1.HasComponent("collider") ||
		!e2.HasComponent("position") || !e2.HasComponent("collider") {
		return false
	}

	pos1Comp, _ := e1.GetComponent("position")
	pos2Comp, _ := e2.GetComponent("position")
	collider1Comp, _ := e1.GetComponent("collider")
	collider2Comp, _ := e2.GetComponent("collider")

	pos1 := pos1Comp.(*PositionComponent)
	pos2 := pos2Comp.(*PositionComponent)
	collider1 := collider1Comp.(*ColliderComponent)
	collider2 := collider2Comp.(*ColliderComponent)

	return collider1.Intersects(pos1.X, pos1.Y, collider2, pos2.X, pos2.Y)
}
