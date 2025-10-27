// Package engine provides the fire propagation system for spreading fire across tiles.
// This file implements FirePropagationSystem which uses cellular automata to spread
// fire between adjacent tiles based on material flammability and fire intensity.
//
// Design Philosophy:
// - Cellular automata with 4-connected neighbor checks (up, down, left, right)
// - Fire spreads based on: intensity, material flammability, and randomness
// - Fire burns for 10-15 seconds before extinguishing (configurable)
// - Performance target: <2ms per frame for up to 100 active fires
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// FirePropagationSystem handles fire spread across tiles using cellular automata.
type FirePropagationSystem struct {
	world    *World
	terrain  *terrain.Terrain
	tileSize int
	rng      *rand.Rand
	logger   *logrus.Entry

	// Performance optimization: track entities with fire components
	fireEntities map[uint64]*Entity
}

// NewFirePropagationSystem creates a new fire propagation system.
func NewFirePropagationSystem(tileSize int, seed int64) *FirePropagationSystem {
	return NewFirePropagationSystemWithLogger(tileSize, seed, nil)
}

// NewFirePropagationSystemWithLogger creates a system with a logger.
func NewFirePropagationSystemWithLogger(tileSize int, seed int64, logger *logrus.Logger) *FirePropagationSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"system":   "fire_propagation",
			"tileSize": tileSize,
			"seed":     seed,
		})
		logEntry.Debug("fire propagation system created")
	}

	return &FirePropagationSystem{
		tileSize:     tileSize,
		rng:          rand.New(rand.NewSource(seed)),
		logger:       logEntry,
		fireEntities: make(map[uint64]*Entity),
	}
}

// SetWorld sets the ECS world reference.
func (s *FirePropagationSystem) SetWorld(world *World) {
	s.world = world
}

// SetTerrain sets the terrain data reference.
func (s *FirePropagationSystem) SetTerrain(t *terrain.Terrain) {
	s.terrain = t
}

// Update implements the System interface.
// Updates fire components and spreads fire to adjacent tiles.
func (s *FirePropagationSystem) Update(entities []*Entity, deltaTime float64) {
	if s.terrain == nil {
		return
	}

	// Update tracking map
	s.updateFireEntities(entities)

	// Process each fire entity
	for _, entity := range s.fireEntities {
		s.updateFire(entity, deltaTime)
	}
}

// updateFireEntities updates the cached map of fire entities.
func (s *FirePropagationSystem) updateFireEntities(entities []*Entity) {
	// Clear old map
	s.fireEntities = make(map[uint64]*Entity)

	// Rebuild from current entities
	for _, entity := range entities {
		if _, ok := entity.GetComponent("fire"); ok {
			s.fireEntities[entity.ID] = entity
		}
	}
}

// updateFire updates a single fire entity.
func (s *FirePropagationSystem) updateFire(entity *Entity, deltaTime float64) {
	comp, ok := entity.GetComponent("fire")
	if !ok {
		return
	}

	fireComp, ok := comp.(*FireComponent)
	if !ok {
		return
	}

	// Update fire timer
	fireComp.Update(deltaTime)

	// Check if fire should be extinguished
	if fireComp.IsExtinguished {
		s.extinguishFire(entity)
		return
	}

	// Get tile position
	tileX, tileY := s.getTilePosition(entity)
	if tileX < 0 || tileY < 0 {
		return
	}

	// Attempt to spread fire to adjacent tiles
	s.attemptFireSpread(entity, tileX, tileY, deltaTime)
}

// attemptFireSpread tries to spread fire to adjacent tiles.
func (s *FirePropagationSystem) attemptFireSpread(entity *Entity, tileX, tileY int, deltaTime float64) {
	comp, ok := entity.GetComponent("fire")
	if !ok {
		return
	}

	fireComp, ok := comp.(*FireComponent)
	if !ok {
		return
	}

	// Check 4-connected neighbors (up, down, left, right)
	neighbors := []struct{ dx, dy int }{
		{0, -1}, // up
		{0, 1},  // down
		{-1, 0}, // left
		{1, 0},  // right
	}

	for _, neighbor := range neighbors {
		neighborX := tileX + neighbor.dx
		neighborY := tileY + neighbor.dy

		// Calculate spread chance for this frame
		spreadChancePerSecond := fireComp.SpreadChance * fireComp.Intensity
		spreadChance := spreadChancePerSecond * deltaTime

		// Roll for spread
		if s.rng.Float64() < spreadChance {
			s.trySpreadToTile(neighborX, neighborY, fireComp.Intensity)
		}
	}
}

// trySpreadToTile attempts to spread fire to a specific tile.
func (s *FirePropagationSystem) trySpreadToTile(tileX, tileY int, intensity float64) {
	if s.terrain == nil {
		return
	}

	// Check if tile is valid
	if !s.terrain.IsInBounds(tileX, tileY) {
		return
	}

	// Check if tile already has fire
	if s.findFireEntityAt(tileX, tileY) != nil {
		return // Already on fire
	}

	// Check if tile is flammable
	if !s.isTileFlammable(tileX, tileY) {
		return
	}

	// Create fire entity at this tile
	s.createFireEntity(tileX, tileY, intensity)
}

// isTileFlammable checks if a tile can catch fire.
func (s *FirePropagationSystem) isTileFlammable(tileX, tileY int) bool {
	if s.terrain == nil {
		return false
	}

	tileType := s.terrain.GetTile(tileX, tileY)

	// Check if tile has a destructible component to determine material
	if destructibleEntity := s.findDestructibleEntityAt(tileX, tileY); destructibleEntity != nil {
		if comp, ok := destructibleEntity.GetComponent("destructible"); ok {
			if destComp, ok := comp.(*DestructibleComponent); ok {
				return destComp.Material.IsFlammable()
			}
		}
	}

	// Default flammability based on tile type
	// Floors can catch fire (assume wood/organic materials)
	// Walls default to non-flammable (assume stone) unless they have a destructible component
	return tileType == terrain.TileFloor
}

// findFireEntityAt finds an existing fire entity at tile coordinates.
func (s *FirePropagationSystem) findFireEntityAt(tileX, tileY int) *Entity {
	for _, entity := range s.fireEntities {
		entityTileX, entityTileY := s.getTilePosition(entity)
		if entityTileX == tileX && entityTileY == tileY {
			return entity
		}
	}
	return nil
}

// findDestructibleEntityAt finds a destructible entity at tile coordinates.
func (s *FirePropagationSystem) findDestructibleEntityAt(tileX, tileY int) *Entity {
	if s.world == nil {
		return nil
	}

	for _, entity := range s.world.GetEntities() {
		if comp, ok := entity.GetComponent("destructible"); ok {
			if destComp, ok := comp.(*DestructibleComponent); ok {
				if destComp.TileX == tileX && destComp.TileY == tileY {
					return entity
				}
			}
		}
	}
	return nil
}

// getTilePosition gets tile coordinates from an entity's position component.
func (s *FirePropagationSystem) getTilePosition(entity *Entity) (int, int) {
	comp, ok := entity.GetComponent("position")
	if !ok {
		return -1, -1
	}

	posComp, ok := comp.(*PositionComponent)
	if !ok {
		return -1, -1
	}

	tileX := int(posComp.X / float64(s.tileSize))
	tileY := int(posComp.Y / float64(s.tileSize))
	return tileX, tileY
}

// createFireEntity creates a new fire entity at tile coordinates.
func (s *FirePropagationSystem) createFireEntity(tileX, tileY int, intensity float64) {
	if s.world == nil {
		return
	}

	// Create entity
	entity := s.world.CreateEntity()

	// Add fire component (default max duration: 12 seconds)
	fireComp := NewFireComponent(intensity, tileX, tileY, 12.0)
	entity.AddComponent(fireComp)

	// Add position component (for spatial queries)
	posComp := &PositionComponent{
		X: float64(tileX*s.tileSize + s.tileSize/2),
		Y: float64(tileY*s.tileSize + s.tileSize/2),
	}
	entity.AddComponent(posComp)

	// Add to tracking map immediately (entity may not be in world.GetEntities() yet)
	s.fireEntities[entity.ID] = entity

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"tileX":     tileX,
			"tileY":     tileY,
			"intensity": intensity,
		}).Debug("fire spread to tile")
	}
}

// extinguishFire removes a fire entity.
func (s *FirePropagationSystem) extinguishFire(entity *Entity) {
	if s.world == nil {
		return
	}

	tileX, tileY := s.getTilePosition(entity)

	// Remove from tracking map
	delete(s.fireEntities, entity.ID)

	// Remove entity
	s.world.RemoveEntity(entity.ID)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"tileX": tileX,
			"tileY": tileY,
		}).Debug("fire extinguished")
	}
}

// IgniteTile creates a fire at the specified tile coordinates.
// This is a helper method for external systems (e.g., fire spells, explosions).
func (s *FirePropagationSystem) IgniteTile(tileX, tileY int, intensity float64) {
	// Don't ignite if already on fire
	if s.findFireEntityAt(tileX, tileY) != nil {
		return
	}

	// Don't ignite non-flammable tiles
	if !s.isTileFlammable(tileX, tileY) {
		return
	}

	s.createFireEntity(tileX, tileY, intensity)
}

// IgniteTilesInArea creates fires in a circular area.
// Useful for fire explosions or area-of-effect fire spells.
func (s *FirePropagationSystem) IgniteTilesInArea(centerX, centerY, radius, intensity float64) {
	if s.terrain == nil {
		return
	}

	// Convert to tile coordinates
	centerTileX := int(centerX / float64(s.tileSize))
	centerTileY := int(centerY / float64(s.tileSize))
	radiusTiles := int(radius/float64(s.tileSize)) + 1

	// Check tiles in square around center
	for dy := -radiusTiles; dy <= radiusTiles; dy++ {
		for dx := -radiusTiles; dx <= radiusTiles; dx++ {
			tileX := centerTileX + dx
			tileY := centerTileY + dy

			// Check if tile is within circular radius
			tileCenterX := float64(tileX*s.tileSize + s.tileSize/2)
			tileCenterY := float64(tileY*s.tileSize + s.tileSize/2)
			distSq := (tileCenterX-centerX)*(tileCenterX-centerX) +
				(tileCenterY-centerY)*(tileCenterY-centerY)

			if distSq <= radius*radius {
				s.IgniteTile(tileX, tileY, intensity)
			}
		}
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"centerX":   centerX,
			"centerY":   centerY,
			"radius":    radius,
			"intensity": intensity,
		}).Debug("area ignited")
	}
}

// GetActiveFireCount returns the number of active fire entities.
// Useful for performance monitoring and debugging.
func (s *FirePropagationSystem) GetActiveFireCount() int {
	return len(s.fireEntities)
}
