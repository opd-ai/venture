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
	// Uses flat map with composite key (x<<32 | y) to eliminate nested map allocations
	flatGrid map[int64][]*Entity

	// Collision callbacks
	onCollision func(e1, e2 *Entity)

	// Terrain collision checker for efficient wall collision
	terrainChecker *TerrainCollisionChecker

	// Pool for collision pair tracking maps to reduce allocations
	// Uses flat map with composite keys to eliminate nested map allocations
	checkedPairPool sync.Pool

	// Reusable buffer for collidable entities to reduce allocations
	collidableBuffer []*Entity
}

// NewCollisionSystem creates a new collision system.
func NewCollisionSystem(cellSize float64) *CollisionSystem {
	return &CollisionSystem{
		CellSize:         cellSize,
		flatGrid:         make(map[int64][]*Entity, 256), // Pre-allocate for typical grid size
		collidableBuffer: make([]*Entity, 0, 256),
		checkedPairPool: sync.Pool{
			New: func() interface{} {
				// Pre-allocate for typical collision pair count
				// 200 entities * ~4 nearby each = ~800 pairs
				return make(map[uint64]bool, 1024)
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

	// Use typed getter for ~94x faster access
	collider := entity.GetCollider()
	if collider == nil {
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
	collider1, collider2, pos2, ok := s.validatePredictiveCollisionComponents(entity, other)
	if !ok {
		return false
	}

	if !s.canCollidePredictive(entity, collider1, other, collider2) {
		return false
	}

	return s.checkPredictiveIntersection(entity, collider1, newX, newY, other, collider2, pos2)
}

// validatePredictiveCollisionComponents retrieves and validates collision components for predictive checks.
// Returns collider1, collider2, pos2, and ok status.
// Uses typed getters for ~94x faster access.
func (s *CollisionSystem) validatePredictiveCollisionComponents(entity, other *Entity) (*ColliderComponent, *ColliderComponent, *PositionComponent, bool) {
	collider1 := entity.GetCollider()
	collider2 := other.GetCollider()
	pos2 := other.GetPosition()

	if collider1 == nil || collider2 == nil || pos2 == nil {
		return nil, nil, nil, false
	}

	return collider1, collider2, pos2, true
}

// canCollidePredictive checks if two entities can collide based on solidity, triggers, and layer compatibility.
func (s *CollisionSystem) canCollidePredictive(entity *Entity, collider1 *ColliderComponent, other *Entity, collider2 *ColliderComponent) bool {
	if collider1.IsTrigger || collider2.IsTrigger {
		return false
	}

	if !collider1.Solid || !collider2.Solid {
		return false
	}

	return s.areLayersCompatible(entity, collider1, other, collider2)
}

// checkPredictiveIntersection performs intersection check at predicted position with rotation support.
func (s *CollisionSystem) checkPredictiveIntersection(entity *Entity, collider1 *ColliderComponent, newX, newY float64, other *Entity, collider2 *ColliderComponent, pos2 *PositionComponent) bool {
	rot1Comp, hasRot1 := entity.GetComponent("rotation")
	rot2Comp, hasRot2 := other.GetComponent("rotation")

	if !hasRot1 && !hasRot2 {
		return collider1.Intersects(newX, newY, collider2, pos2.X, pos2.Y)
	}

	angle1, angle2 := s.extractRotationAngles(rot1Comp, hasRot1, rot2Comp, hasRot2)
	return collider1.IntersectsRotated(newX, newY, angle1, collider2, pos2.X, pos2.Y, angle2)
}

// Update detects and resolves collisions between entities.
func (s *CollisionSystem) Update(entities []*Entity, deltaTime float64) {
	collidableEntities := s.collectAndGridCollidableEntities(entities)
	checked := s.acquireCheckedPairs()
	defer s.releaseCheckedPairs(checked)

	for _, entity := range collidableEntities {
		s.processEntityCollisions(entity, collidableEntities, checked)
	}
}

// collectAndGridCollidableEntities filters entities with colliders and builds the spatial grid.
// Uses cached typed getters for ~93x faster component access vs HasComponent map lookups.
func (s *CollisionSystem) collectAndGridCollidableEntities(entities []*Entity) []*Entity {
	// Clear flat grid - reuse existing slices
	for key := range s.flatGrid {
		s.flatGrid[key] = s.flatGrid[key][:0]
	}

	// Reuse collidableBuffer to reduce allocations
	s.collidableBuffer = s.collidableBuffer[:0]
	if cap(s.collidableBuffer) < len(entities) {
		s.collidableBuffer = make([]*Entity, 0, len(entities))
	}

	for _, entity := range entities {
		// Use cached typed getters for ~93x faster access vs HasComponent map lookups
		if entity.GetCollider() != nil && entity.GetPosition() != nil {
			s.collidableBuffer = append(s.collidableBuffer, entity)
		}
	}

	for _, entity := range s.collidableBuffer {
		s.addToGrid(entity)
	}

	return s.collidableBuffer
}

// makeCollisionGridKey creates a composite key from grid cell coordinates.
// Uses the same pattern as makePairKey for consistency.
func makeCollisionGridKey(x, y int) int64 {
	return (int64(x) << 32) | (int64(y) & 0xFFFFFFFF)
}

// acquireCheckedPairs obtains a cleaned collision pair tracking map from the pool.
// Uses flat map with composite keys to eliminate nested map allocations.
func (s *CollisionSystem) acquireCheckedPairs() map[uint64]bool {
	checked := s.checkedPairPool.Get().(map[uint64]bool)
	// Clear the map for reuse
	// Use clear() builtin (Go 1.21+) for ~21% faster map clearing vs delete loop
	clear(checked)
	return checked
}

// releaseCheckedPairs returns the collision pair map to the pool.
func (s *CollisionSystem) releaseCheckedPairs(checked map[uint64]bool) {
	// Clear if too large to prevent memory bloat
	if len(checked) > 4096 {
		checked = make(map[uint64]bool, 1024)
	}
	s.checkedPairPool.Put(checked)
}

// makePairKey creates a unique key for an entity pair.
// Uses canonical ordering (smaller ID first) to ensure (a,b) == (b,a).
func makePairKey(id1, id2 uint64) uint64 {
	if id1 > id2 {
		id1, id2 = id2, id1
	}
	return (id1 << 32) | (id2 & 0xFFFFFFFF)
}

// processEntityCollisions handles collision detection and resolution for a single entity.
func (s *CollisionSystem) processEntityCollisions(entity *Entity, collidableEntities []*Entity, checked map[uint64]bool) {
	// Use typed getters for ~94x faster access
	pos := entity.GetPosition()
	collider := entity.GetCollider()

	if pos == nil || collider == nil {
		return
	}

	// Use pooled result directly without copying
	nr := s.getNearbyEntitiesPooled(entity)
	defer putNearbyResult(nr)

	for _, other := range nr.result {
		if entity.ID == other.ID {
			continue
		}

		pairKey := makePairKey(entity.ID, other.ID)
		if checked[pairKey] {
			continue
		}

		checked[pairKey] = true

		if s.checkAndResolveEntityPair(entity, pos, collider, other) {
			continue
		}
	}

	s.checkTerrainCollision(entity, collider)
}

// isCollisionPairChecked returns true if the entity pair has already been checked.
// Uses flat map with composite key for O(1) lookup without nested map allocations.
func (s *CollisionSystem) isCollisionPairChecked(id1, id2 uint64, checked map[uint64]bool) bool {
	return checked[makePairKey(id1, id2)]
}

// markCollisionPairChecked marks an entity pair as checked.
// Uses flat map with composite key - no inner map allocations needed.
func (s *CollisionSystem) markCollisionPairChecked(id1, id2 uint64, checked map[uint64]bool) {
	checked[makePairKey(id1, id2)] = true
}

// checkAndResolveEntityPair checks if two entities collide and resolves the collision.
// Returns true if processing should skip this pair (invalid components or incompatible layers).
func (s *CollisionSystem) checkAndResolveEntityPair(entity *Entity, pos *PositionComponent, collider *ColliderComponent, other *Entity) bool {
	// Use typed getters for ~94x faster access
	otherPos := other.GetPosition()
	otherCollider := other.GetCollider()

	if otherPos == nil || otherCollider == nil {
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
	// Use typed getters for ~94x faster access
	pos := entity.GetPosition()
	collider := entity.GetCollider()

	if pos == nil || collider == nil {
		// Components missing - should not happen if Update() filters correctly
		return
	}

	// Get bounding box
	minX, minY, maxX, maxY := collider.GetBounds(pos.X, pos.Y)

	// Calculate grid cells this entity occupies
	minCellX := int(math.Floor(minX / s.CellSize))
	minCellY := int(math.Floor(minY / s.CellSize))
	maxCellX := int(math.Floor(maxX / s.CellSize))
	maxCellY := int(math.Floor(maxY / s.CellSize))

	// Add to all occupied cells using flat grid
	for x := minCellX; x <= maxCellX; x++ {
		for y := minCellY; y <= maxCellY; y++ {
			key := makeCollisionGridKey(x, y)
			s.flatGrid[key] = append(s.flatGrid[key], entity)
		}
	}
}

// getNearbyEntitiesPooled returns pooled nearby entities without copying.
// The result must be returned to the pool after use via putNearbyResult().
// DO NOT store the result - it will be reused by the pool.
func (s *CollisionSystem) getNearbyEntitiesPooled(entity *Entity) *nearbyResult {
	// Use typed getters for ~94x faster access
	pos := entity.GetPosition()
	collider := entity.GetCollider()

	nr := getNearbyResult()

	if pos == nil || collider == nil {
		// Components missing - return empty result
		return nr
	}

	minX, minY, maxX, maxY := collider.GetBounds(pos.X, pos.Y)

	// Calculate grid cells
	minCellX := int(math.Floor(minX / s.CellSize))
	minCellY := int(math.Floor(minY / s.CellSize))
	maxCellX := int(math.Floor(maxX / s.CellSize))
	maxCellY := int(math.Floor(maxY / s.CellSize))

	for x := minCellX; x <= maxCellX; x++ {
		for y := minCellY; y <= maxCellY; y++ {
			key := makeCollisionGridKey(x, y)
			if entities := s.flatGrid[key]; len(entities) > 0 {
				for _, e := range entities {
					if !nr.seen[e.ID] {
						nr.seen[e.ID] = true
						nr.result = append(nr.result, e)
					}
				}
			}
		}
	}

	return nr
}

// getNearbyEntities returns entities in the same or adjacent grid cells.
// Precondition: Entity must have position and collider components.
// Note: This method allocates a new slice. For hot paths, use getNearbyEntitiesPooled.
func (s *CollisionSystem) getNearbyEntities(entity *Entity) []*Entity {
	nr := s.getNearbyEntitiesPooled(entity)
	defer putNearbyResult(nr)

	// Copy result before returning (pool will be reused)
	result := make([]*Entity, len(nr.result))
	copy(result, nr.result)
	return result
}

// resolveCollision separates two colliding entities.
// Precondition: Both entities must have position and collider components.
func (s *CollisionSystem) resolveCollision(e1, e2 *Entity) {
	pos1, pos2, collider1, collider2, ok := s.getCollisionComponents(e1, e2)
	if !ok {
		return
	}

	min1X, min1Y, max1X, max1Y := collider1.GetBounds(pos1.X, pos1.Y)
	min2X, min2Y, max2X, max2Y := collider2.GetBounds(pos2.X, pos2.Y)

	overlapX := math.Min(max1X-min2X, max2X-min1X)
	overlapY := math.Min(max1Y-min2Y, max2Y-min1Y)

	if overlapX < overlapY {
		s.separateEntitiesHorizontally(e1, e2, pos1, pos2, overlapX)
	} else {
		s.separateEntitiesVertically(e1, e2, pos1, pos2, overlapY)
	}
}

// getCollisionComponents extracts and validates position and collider components.
// Uses typed getters for ~94x faster access.
func (s *CollisionSystem) getCollisionComponents(e1, e2 *Entity) (*PositionComponent, *PositionComponent, *ColliderComponent, *ColliderComponent, bool) {
	pos1 := e1.GetPosition()
	pos2 := e2.GetPosition()
	collider1 := e1.GetCollider()
	collider2 := e2.GetCollider()

	if pos1 == nil || pos2 == nil || collider1 == nil || collider2 == nil {
		return nil, nil, nil, nil, false
	}

	return pos1, pos2, collider1, collider2, true
}

// separateEntitiesHorizontally separates entities along X axis and stops horizontal velocity.
func (s *CollisionSystem) separateEntitiesHorizontally(e1, e2 *Entity, pos1, pos2 *PositionComponent, overlapX float64) {
	if pos1.X < pos2.X {
		pos1.X -= overlapX / 2
		pos2.X += overlapX / 2
	} else {
		pos1.X += overlapX / 2
		pos2.X -= overlapX / 2
	}

	s.stopHorizontalVelocity(e1)
	s.stopHorizontalVelocity(e2)
}

// separateEntitiesVertically separates entities along Y axis and stops vertical velocity.
func (s *CollisionSystem) separateEntitiesVertically(e1, e2 *Entity, pos1, pos2 *PositionComponent, overlapY float64) {
	if pos1.Y < pos2.Y {
		pos1.Y -= overlapY / 2
		pos2.Y += overlapY / 2
	} else {
		pos1.Y += overlapY / 2
		pos2.Y -= overlapY / 2
	}

	s.stopVerticalVelocity(e1)
	s.stopVerticalVelocity(e2)
}

// stopHorizontalVelocity sets entity's horizontal velocity to zero.
// Uses typed getter for ~92x faster access vs HasComponent+GetComponent+type assertion.
func (s *CollisionSystem) stopHorizontalVelocity(entity *Entity) {
	if vel := entity.GetVelocity(); vel != nil {
		vel.VX = 0
	}
}

// stopVerticalVelocity sets entity's vertical velocity to zero.
// Uses typed getter for ~92x faster access vs HasComponent+GetComponent+type assertion.
func (s *CollisionSystem) stopVerticalVelocity(entity *Entity) {
	if vel := entity.GetVelocity(); vel != nil {
		vel.VY = 0
	}
}

// resolveTerrainCollision resolves collision between an entity and terrain walls.
func (s *CollisionSystem) resolveTerrainCollision(entity *Entity) {
	// Use typed getters for ~94x faster access
	pos := entity.GetPosition()
	collider := entity.GetCollider()

	if pos == nil || collider == nil {
		return
	}

	if s.findValidPosition(entity, pos, collider) {
		return
	}

	s.stopAllMovement(entity)
}

// findValidPosition attempts to find a valid position by pushing away from walls.
// Returns true if a valid position was found, false otherwise.
func (s *CollisionSystem) findValidPosition(entity *Entity, pos *PositionComponent, collider *ColliderComponent) bool {
	directions := []struct{ dx, dy float64 }{
		{-1, 0}, {1, 0}, {0, -1}, {0, 1}, // Cardinal directions
		{-1, -1}, {1, -1}, {-1, 1}, {1, 1}, // Diagonal directions
	}

	originalX, originalY := pos.X, pos.Y
	pushDistance := 2.0

	for _, dir := range directions {
		testX := originalX + dir.dx*pushDistance
		testY := originalY + dir.dy*pushDistance

		if !s.terrainChecker.CheckCollision(testX, testY, collider.Width, collider.Height) {
			pos.X = testX
			pos.Y = testY
			s.stopBlockedMovement(entity, dir.dx, dir.dy)
			return true
		}
	}

	return false
}

// stopBlockedMovement stops velocity components that are moving into a wall.
// Uses typed getter for ~92x faster access vs HasComponent+GetComponent+type assertion.
func (s *CollisionSystem) stopBlockedMovement(entity *Entity, dx, dy float64) {
	vel := entity.GetVelocity()
	if vel == nil {
		return
	}
	if dx != 0 {
		vel.VX = 0
	}
	if dy != 0 {
		vel.VY = 0
	}
}

// stopAllMovement stops all velocity components of an entity.
// Uses typed getter for ~92x faster access vs HasComponent+GetComponent+type assertion.
func (s *CollisionSystem) stopAllMovement(entity *Entity) {
	vel := entity.GetVelocity()
	if vel == nil {
		return
	}
	vel.VX = 0
	vel.VY = 0
}

// CheckCollision checks if two entities are colliding.
// Uses typed getters for ~94x faster access.
func CheckCollision(e1, e2 *Entity) bool {
	pos1 := e1.GetPosition()
	pos2 := e2.GetPosition()
	collider1 := e1.GetCollider()
	collider2 := e2.GetCollider()

	if pos1 == nil || pos2 == nil || collider1 == nil || collider2 == nil {
		return false
	}

	return collider1.Intersects(pos1.X, pos1.Y, collider2, pos2.X, pos2.Y)
}
