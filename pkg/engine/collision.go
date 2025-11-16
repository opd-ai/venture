// Package engine provides collision detection and resolution.
// This file implements CollisionSystem with spatial partitioning for efficient
// broad-phase collision detection using a grid-based approach.
package engine

import (
	"math"
	"sync"
)

// CollisionSystem handles collision detection and resolution.
// Uses spatial partitioning (grid-based) for efficient broad-phase detection.
type CollisionSystem struct {
	// Grid cell size for spatial partitioning
	CellSize float64

	// Spatial grid for broad-phase collision detection
	grid map[int]map[int][]*Entity

	// Collision callbacks
	onCollision func(e1, e2 *Entity)

	// Terrain collision checker for efficient wall collision
	terrainChecker *TerrainCollisionChecker

	// Pool for collision pair tracking maps to reduce allocations
	checkedMapPool sync.Pool
}

// NewCollisionSystem creates a new collision system.
func NewCollisionSystem(cellSize float64) *CollisionSystem {
	return &CollisionSystem{
		CellSize: cellSize,
		grid:     make(map[int]map[int][]*Entity),
		checkedMapPool: sync.Pool{
			New: func() interface{} {
				return make(map[uint64]map[uint64]bool)
			},
		},
	}
}

// SetTerrainChecker sets the terrain collision checker for wall collision.
func (s *CollisionSystem) SetTerrainChecker(checker *TerrainCollisionChecker) {
	s.terrainChecker = checker
}

// SetCollisionCallback sets a function to be called when entities collide.
func (s *CollisionSystem) SetCollisionCallback(callback func(e1, e2 *Entity)) {
	s.onCollision = callback
}

// WouldCollideWithTerrain checks if an entity would collide with terrain at the given position.
// This is a predictive check that doesn't modify entity state.
// Returns true if the entity would collide with terrain walls at the specified position.
func (s *CollisionSystem) WouldCollideWithTerrain(entity *Entity, newX, newY float64) bool {
	if s.terrainChecker == nil {
		return false
	}

	if !entity.HasComponent("collider") {
		return false
	}

	colliderComp, _ := entity.GetComponent("collider")
	collider, ok := colliderComp.(*ColliderComponent)
	if !ok {
		return false
	}

	// Only check solid colliders (triggers don't block movement)
	if !collider.Solid || collider.IsTrigger {
		return false
	}

	// Get bounds at the predicted position
	minX, minY, maxX, maxY := collider.GetBounds(newX, newY)

	// Check collision using bounds
	return s.terrainChecker.CheckCollisionBounds(minX, minY, maxX, maxY)
}

// WouldCollideWithEntity checks if an entity would collide with another entity at the given position.
// This is a predictive check that doesn't modify entity state.
// Returns true if the entity would collide with the other entity at the specified position.
func (s *CollisionSystem) WouldCollideWithEntity(entity *Entity, newX, newY float64, other *Entity) bool {
	if !entity.HasComponent("collider") || !other.HasComponent("collider") {
		return false
	}

	if !other.HasComponent("position") {
		return false
	}

	// Get collider components
	collider1Comp, _ := entity.GetComponent("collider")
	collider2Comp, _ := other.GetComponent("collider")
	collider1, ok := collider1Comp.(*ColliderComponent)
	if !ok {
		return false
	}
	collider2, ok := collider2Comp.(*ColliderComponent)
	if !ok {
		return false
	}

	// Skip trigger colliders (they don't block movement)
	if collider1.IsTrigger || collider2.IsTrigger {
		return false
	}

	// Skip if either is not solid
	if !collider1.Solid || !collider2.Solid {
		return false
	}

	// Check layer compatibility (0 = all layers)
	if collider1.Layer != 0 && collider2.Layer != 0 && collider1.Layer != collider2.Layer {
		return false
	}

	// Check terrain layer compatibility for prediction (Phase 11.1 multi-layer support)
	layer1Comp, hasLayer1 := entity.GetComponent("layer")
	layer2Comp, hasLayer2 := other.GetComponent("layer")
	if hasLayer1 && hasLayer2 {
		l1, ok := layer1Comp.(*LayerComponent)
		if !ok {
			return false
		}
		l2, ok := layer2Comp.(*LayerComponent)
		if !ok {
			return false
		}
		// Flying entities collide with all layers
		if !l1.CanFly && !l2.CanFly {
			// Check if entities are on same effective terrain layer
			if !OnSameLayer(l1, l2) {
				return false // No collision across terrain layers
			}
		}
	}

	// Get other entity's current position
	pos2Comp, _ := other.GetComponent("position")
	pos2, ok := pos2Comp.(*PositionComponent)
	if !ok {
		return false
	}

	// Issue #20: Check intersection at predicted position with rotation support
	rot1Comp, hasRot1 := entity.GetComponent("rotation")
	rot2Comp, hasRot2 := other.GetComponent("rotation")

	if hasRot1 || hasRot2 {
		// Use rotation-aware collision for rotated entities
		angle1 := 0.0
		angle2 := 0.0
		if hasRot1 {
			if rot1, ok := rot1Comp.(*RotationComponent); ok {
				angle1 = rot1.Angle
			}
		}
		if hasRot2 {
			if rot2, ok := rot2Comp.(*RotationComponent); ok {
				angle2 = rot2.Angle
			}
		}
		return collider1.IntersectsRotated(newX, newY, angle1, collider2, pos2.X, pos2.Y, angle2)
	}

	// Check intersection at predicted position (no rotation)
	return collider1.Intersects(newX, newY, collider2, pos2.X, pos2.Y)
}

// Update detects and resolves collisions between entities.
func (s *CollisionSystem) Update(entities []*Entity, deltaTime float64) {
	collidableEntities := s.collectAndGridCollidableEntities(entities)
	checked := s.acquireCheckedMap()
	defer s.checkedMapPool.Put(checked)

	for _, entity := range collidableEntities {
		s.processEntityCollisions(entity, collidableEntities, checked)
	}
}

// collectAndGridCollidableEntities filters entities with colliders and builds the spatial grid.
func (s *CollisionSystem) collectAndGridCollidableEntities(entities []*Entity) []*Entity {
	s.grid = make(map[int]map[int][]*Entity)
	collidableEntities := make([]*Entity, 0, len(entities))

	for _, entity := range entities {
		if entity.HasComponent("collider") && entity.HasComponent("position") {
			collidableEntities = append(collidableEntities, entity)
		}
	}

	for _, entity := range collidableEntities {
		s.addToGrid(entity)
	}

	return collidableEntities
}

// acquireCheckedMap obtains a cleaned collision tracking map from the pool.
func (s *CollisionSystem) acquireCheckedMap() map[uint64]map[uint64]bool {
	checked := s.checkedMapPool.Get().(map[uint64]map[uint64]bool)
	for k := range checked {
		delete(checked, k)
	}
	return checked
}

// processEntityCollisions handles collision detection and resolution for a single entity.
func (s *CollisionSystem) processEntityCollisions(entity *Entity, collidableEntities []*Entity, checked map[uint64]map[uint64]bool) {
	posComp, _ := entity.GetComponent("position")
	colliderComp, _ := entity.GetComponent("collider")

	pos, ok1 := posComp.(*PositionComponent)
	collider, ok2 := colliderComp.(*ColliderComponent)

	if !ok1 || !ok2 {
		return
	}

	candidates := s.getNearbyEntities(entity)
	for _, other := range candidates {
		if entity.ID == other.ID {
			continue
		}

		if s.isCollisionPairChecked(entity.ID, other.ID, checked) {
			continue
		}

		s.markCollisionPairChecked(entity.ID, other.ID, checked)

		if s.checkAndResolveEntityPair(entity, pos, collider, other) {
			continue
		}
	}

	s.checkTerrainCollision(entity, collider)
}

// isCollisionPairChecked returns true if the entity pair has already been checked.
func (s *CollisionSystem) isCollisionPairChecked(id1, id2 uint64, checked map[uint64]map[uint64]bool) bool {
	return checked[id1] != nil && checked[id1][id2]
}

// markCollisionPairChecked marks an entity pair as checked in both directions.
func (s *CollisionSystem) markCollisionPairChecked(id1, id2 uint64, checked map[uint64]map[uint64]bool) {
	if checked[id1] == nil {
		checked[id1] = make(map[uint64]bool)
	}
	if checked[id2] == nil {
		checked[id2] = make(map[uint64]bool)
	}
	checked[id1][id2] = true
	checked[id2][id1] = true
}

// checkAndResolveEntityPair checks if two entities collide and resolves the collision.
// Returns true if processing should skip this pair (invalid components or incompatible layers).
func (s *CollisionSystem) checkAndResolveEntityPair(entity *Entity, pos *PositionComponent, collider *ColliderComponent, other *Entity) bool {
	otherPosComp, _ := other.GetComponent("position")
	otherColliderComp, _ := other.GetComponent("collider")

	otherPos, ok1 := otherPosComp.(*PositionComponent)
	otherCollider, ok2 := otherColliderComp.(*ColliderComponent)

	if !ok1 || !ok2 {
		return true
	}

	if !s.areLayersCompatible(entity, collider, other, otherCollider) {
		return true
	}

	if s.detectIntersection(entity, pos, collider, other, otherPos, otherCollider) {
		s.handleCollision(entity, collider, other, otherCollider)
	}

	return false
}

// areLayersCompatible checks if two entities can collide based on their layer settings.
func (s *CollisionSystem) areLayersCompatible(entity *Entity, collider *ColliderComponent, other *Entity, otherCollider *ColliderComponent) bool {
	if collider.Layer != 0 && otherCollider.Layer != 0 && collider.Layer != otherCollider.Layer {
		return false
	}

	layer1Comp, hasLayer1 := entity.GetComponent("layer")
	layer2Comp, hasLayer2 := other.GetComponent("layer")

	if !hasLayer1 || !hasLayer2 {
		return true
	}

	l1, ok := layer1Comp.(*LayerComponent)
	if !ok {
		return false
	}
	l2, ok := layer2Comp.(*LayerComponent)
	if !ok {
		return false
	}

	if l1.CanFly || l2.CanFly {
		return true
	}

	return OnSameLayer(l1, l2)
}

// detectIntersection determines if two entities intersect, accounting for rotation.
func (s *CollisionSystem) detectIntersection(entity *Entity, pos *PositionComponent, collider *ColliderComponent, other *Entity, otherPos *PositionComponent, otherCollider *ColliderComponent) bool {
	rot1Comp, hasRot1 := entity.GetComponent("rotation")
	rot2Comp, hasRot2 := other.GetComponent("rotation")

	if !hasRot1 && !hasRot2 {
		return collider.Intersects(pos.X, pos.Y, otherCollider, otherPos.X, otherPos.Y)
	}

	angle1, angle2 := s.extractRotationAngles(rot1Comp, hasRot1, rot2Comp, hasRot2)
	return collider.IntersectsRotated(pos.X, pos.Y, angle1, otherCollider, otherPos.X, otherPos.Y, angle2)
}

// extractRotationAngles retrieves rotation angles from rotation components.
func (s *CollisionSystem) extractRotationAngles(rot1Comp Component, hasRot1 bool, rot2Comp Component, hasRot2 bool) (float64, float64) {
	angle1 := 0.0
	angle2 := 0.0

	if hasRot1 {
		if rot1, ok := rot1Comp.(*RotationComponent); ok {
			angle1 = rot1.Angle
		}
	}
	if hasRot2 {
		if rot2, ok := rot2Comp.(*RotationComponent); ok {
			angle2 = rot2.Angle
		}
	}

	return angle1, angle2
}

// handleCollision processes a detected collision between two entities.
func (s *CollisionSystem) handleCollision(entity *Entity, collider *ColliderComponent, other *Entity, otherCollider *ColliderComponent) {
	if s.onCollision != nil {
		s.onCollision(entity, other)
	}

	if collider.Solid && otherCollider.Solid && !collider.IsTrigger && !otherCollider.IsTrigger {
		s.resolveCollision(entity, other)
	}
}

// checkTerrainCollision checks and resolves terrain collision for an entity.
func (s *CollisionSystem) checkTerrainCollision(entity *Entity, collider *ColliderComponent) {
	if s.terrainChecker != nil && collider.Solid && !collider.IsTrigger {
		if s.terrainChecker.CheckEntityCollision(entity) {
			s.resolveTerrainCollision(entity)
		}
	}
}

// addToGrid adds an entity to the spatial grid.
// Precondition: Entity must have position and collider components.
func (s *CollisionSystem) addToGrid(entity *Entity) {
	posComp, _ := entity.GetComponent("position")
	colliderComp, _ := entity.GetComponent("collider")

	// Safe type assertions with nil checks
	pos, ok1 := posComp.(*PositionComponent)
	collider, ok2 := colliderComp.(*ColliderComponent)

	if !ok1 || !ok2 {
		// Components missing or wrong type - should not happen if Update() filters correctly
		return
	}

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
// Precondition: Entity must have position and collider components.
func (s *CollisionSystem) getNearbyEntities(entity *Entity) []*Entity {
	posComp, _ := entity.GetComponent("position")
	colliderComp, _ := entity.GetComponent("collider")

	// Safe type assertions with nil checks
	pos, ok1 := posComp.(*PositionComponent)
	collider, ok2 := colliderComp.(*ColliderComponent)

	if !ok1 || !ok2 {
		// Components missing or wrong type - should not happen if Update() filters correctly
		return nil
	}

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
// Precondition: Both entities must have position and collider components.
func (s *CollisionSystem) resolveCollision(e1, e2 *Entity) {
	pos1Comp, _ := e1.GetComponent("position")
	pos2Comp, _ := e2.GetComponent("position")
	collider1Comp, _ := e1.GetComponent("collider")
	collider2Comp, _ := e2.GetComponent("collider")

	// Safe type assertions with nil checks to prevent panics
	pos1, ok1 := pos1Comp.(*PositionComponent)
	pos2, ok2 := pos2Comp.(*PositionComponent)
	collider1, ok3 := collider1Comp.(*ColliderComponent)
	collider2, ok4 := collider2Comp.(*ColliderComponent)

	if !ok1 || !ok2 || !ok3 || !ok4 {
		// Components missing or wrong type - should not happen if Update() filters correctly
		return
	}

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
			if v, ok := vel1.(*VelocityComponent); ok {
				v.VX = 0
			}
		}
		if e2.HasComponent("velocity") {
			vel2, _ := e2.GetComponent("velocity")
			if v, ok := vel2.(*VelocityComponent); ok {
				v.VX = 0
			}
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
			if v, ok := vel1.(*VelocityComponent); ok {
				v.VY = 0
			}
		}
		if e2.HasComponent("velocity") {
			vel2, _ := e2.GetComponent("velocity")
			if v, ok := vel2.(*VelocityComponent); ok {
				v.VY = 0
			}
		}
	}
}

// resolveTerrainCollision resolves collision between an entity and terrain walls.
func (s *CollisionSystem) resolveTerrainCollision(entity *Entity) {
	if !entity.HasComponent("position") || !entity.HasComponent("collider") {
		return
	}

	posComp, _ := entity.GetComponent("position")
	colliderComp, _ := entity.GetComponent("collider")

	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return
	}
	collider, ok := colliderComp.(*ColliderComponent)
	if !ok {
		return
	}

	// Try to find a valid position by moving away from walls
	// Check 8 directions around the entity
	directions := []struct{ dx, dy float64 }{
		{-1, 0}, {1, 0}, {0, -1}, {0, 1}, // Cardinal directions
		{-1, -1}, {1, -1}, {-1, 1}, {1, 1}, // Diagonal directions
	}

	originalX, originalY := pos.X, pos.Y
	pushDistance := 2.0 // Pixels to push away from wall

	for _, dir := range directions {
		testX := originalX + dir.dx*pushDistance
		testY := originalY + dir.dy*pushDistance

		// Check if this position is clear of terrain walls
		if !s.terrainChecker.CheckCollision(testX, testY, collider.Width, collider.Height) {
			pos.X = testX
			pos.Y = testY

			// Stop movement in the blocked direction
			if entity.HasComponent("velocity") {
				vel, _ := entity.GetComponent("velocity")
				velocity, ok := vel.(*VelocityComponent)
				if !ok {
					return
				}

				// Stop velocity component that's moving into the wall
				if dir.dx != 0 {
					velocity.VX = 0
				}
				if dir.dy != 0 {
					velocity.VY = 0
				}
			}
			return
		}
	}

	// If no direction works, stop all movement
	if entity.HasComponent("velocity") {
		vel, _ := entity.GetComponent("velocity")
		velocity, ok := vel.(*VelocityComponent)
		if !ok {
			return
		}
		velocity.VX = 0
		velocity.VY = 0
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

	// Safe type assertions with nil checks
	pos1, ok1 := pos1Comp.(*PositionComponent)
	pos2, ok2 := pos2Comp.(*PositionComponent)
	collider1, ok3 := collider1Comp.(*ColliderComponent)
	collider2, ok4 := collider2Comp.(*ColliderComponent)

	if !ok1 || !ok2 || !ok3 || !ok4 {
		return false
	}

	return collider1.Intersects(pos1.X, pos1.Y, collider2, pos2.X, pos2.Y)
}
