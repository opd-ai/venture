// Package engine provides the terrain modification system for destructible terrain.
// This file implements TerrainModificationSystem which handles weapon and spell-based
// terrain destruction, applying damage to tiles and replacing destroyed tiles with floor.
//
// Design Philosophy:
// - Server-authoritative for multiplayer (network sync required)
// - Weapon types (pickaxe, bombs) and spell types (fire, explosion) determine damage
// - Material type affects durability and destruction behavior
// - Performance target: <1ms per frame for damage checks
package engine

import (
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/world"
	"github.com/sirupsen/logrus"
)

// TerrainModificationSystem handles destructible terrain and tile damage.
type TerrainModificationSystem struct {
	world    *World
	terrain  *terrain.Terrain
	worldMap *world.Map
	tileSize int
	logger   *logrus.Entry
}

// NewTerrainModificationSystem creates a new terrain modification system.
func NewTerrainModificationSystem(tileSize int) *TerrainModificationSystem {
	return NewTerrainModificationSystemWithLogger(tileSize, nil)
}

// NewTerrainModificationSystemWithLogger creates a system with a logger.
func NewTerrainModificationSystemWithLogger(tileSize int, logger *logrus.Logger) *TerrainModificationSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"system":   "terrain_modification",
			"tileSize": tileSize,
		})
		logEntry.Debug("terrain modification system created")
	}

	return &TerrainModificationSystem{
		tileSize: tileSize,
		logger:   logEntry,
	}
}

// SetWorld sets the ECS world reference.
func (s *TerrainModificationSystem) SetWorld(world *World) {
	s.world = world
}

// SetTerrain sets the terrain data reference.
func (s *TerrainModificationSystem) SetTerrain(t *terrain.Terrain) {
	s.terrain = t
}

// SetWorldMap sets the world map reference for tile modification.
func (s *TerrainModificationSystem) SetWorldMap(m *world.Map) {
	s.worldMap = m
}

// Update implements the System interface.
// Processes terrain damage from attacks and spells.
func (s *TerrainModificationSystem) Update(entities []*Entity, deltaTime float64) {
	if s.terrain == nil || s.worldMap == nil {
		return
	}

	// Update destructible tile entities
	for _, entity := range entities {
		if comp, ok := entity.GetComponent("destructible"); ok {
			if destComp, ok := comp.(*DestructibleComponent); ok {
				if destComp.IsDestroyed {
					s.replaceTileWithFloor(destComp.TileX, destComp.TileY)
					// Remove the destructible entity
					if s.world != nil {
						s.world.RemoveEntity(entity.ID)
					}
				}
			}
		}
	}
}

// ProcessWeaponAttack handles terrain damage from a weapon attack.
// Should be called when an entity performs an attack action with a weapon.
func (s *TerrainModificationSystem) ProcessWeaponAttack(entity *Entity, weapon *item.Item) {
	if weapon == nil {
		return
	}

	s.processWeaponDamage(entity, weapon)
}

// processWeaponDamage handles terrain damage from weapon attacks.
func (s *TerrainModificationSystem) processWeaponDamage(entity *Entity, weapon *item.Item) {
	// Get entity position and facing direction
	comp, hasPos := entity.GetComponent("position")
	if !hasPos {
		return
	}
	posComp, ok := comp.(*PositionComponent)
	if !ok {
		return
	}

	// Only certain weapon types can damage terrain
	if !s.canWeaponDamageTerrain(weapon) {
		return
	}

	// Calculate tile in front of entity based on facing direction
	attackDir := s.getAttackDirection(entity)
	targetTileX := int(posComp.X/float64(s.tileSize)) + attackDir.X
	targetTileY := int(posComp.Y/float64(s.tileSize)) + attackDir.Y

	// Apply damage to tile
	damage := s.getWeaponTerrainDamage(weapon)
	s.damageTile(targetTileX, targetTileY, damage)
}

// processSpellDamage handles terrain damage from spells.
func (s *TerrainModificationSystem) processSpellDamage(entity *Entity, spellComp *SpellSlotComponent) {
	// Implementation for spell-based terrain damage
	// Would check for active fire/explosion spells and apply area damage
	// Deferred: Requires spell system integration
}

// canWeaponDamageTerrain checks if a weapon type can damage terrain.
func (s *TerrainModificationSystem) canWeaponDamageTerrain(weapon *item.Item) bool {
	if weapon == nil {
		return false
	}

	// Pickaxes, hammers, and bombs can damage terrain
	// Check weapon name or type (simplified check)
	weaponType := weapon.Type
	return weaponType == item.TypeWeapon // Simplified: all weapons can damage terrain
}

// getWeaponTerrainDamage calculates terrain damage from a weapon.
func (s *TerrainModificationSystem) getWeaponTerrainDamage(weapon *item.Item) float64 {
	if weapon == nil {
		return 0
	}

	// Base damage is weapon's damage stat
	baseDamage := float64(weapon.Stats.Damage)

	// Terrain takes reduced damage (weapons designed for creatures)
	terrainMultiplier := 0.5

	return baseDamage * terrainMultiplier
}

// getAttackDirection gets the direction of attack based on entity facing/movement.
func (s *TerrainModificationSystem) getAttackDirection(entity *Entity) struct{ X, Y int } {
	// Default: attack to the right
	dir := struct{ X, Y int }{X: 1, Y: 0}

	// Check animation component for facing direction
	if comp, ok := entity.GetComponent("animation"); ok {
		if animComp, ok := comp.(*AnimationComponent); ok {
			switch animComp.Facing {
			case DirUp:
				dir = struct{ X, Y int }{X: 0, Y: -1}
			case DirDown:
				dir = struct{ X, Y int }{X: 0, Y: 1}
			case DirLeft:
				dir = struct{ X, Y int }{X: -1, Y: 0}
			case DirRight:
				dir = struct{ X, Y int }{X: 1, Y: 0}
			}
		}
	}

	return dir
}

// damageTile applies damage to a tile at the given coordinates.
func (s *TerrainModificationSystem) damageTile(tileX, tileY int, damage float64) {
	if s.terrain == nil || s.worldMap == nil {
		return
	}

	// Check if tile is a wall
	tileType := s.terrain.GetTile(tileX, tileY)
	if tileType != terrain.TileWall {
		return // Can only damage walls
	}

	// Find or create destructible component for this tile
	destEntity := s.findDestructibleEntityAt(tileX, tileY)
	if destEntity == nil {
		// Create a new destructible entity for this tile
		destEntity = s.createDestructibleEntity(tileX, tileY)
	}

	// Apply damage
	if comp, ok := destEntity.GetComponent("destructible"); ok {
		if destComp, ok := comp.(*DestructibleComponent); ok {
			destroyed := destComp.TakeDamage(damage)

			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"tileX":     tileX,
					"tileY":     tileY,
					"damage":    damage,
					"health":    destComp.Health,
					"destroyed": destroyed,
				}).Debug("tile damaged")
			}
		}
	}
}

// findDestructibleEntityAt finds an existing destructible entity at tile coordinates.
func (s *TerrainModificationSystem) findDestructibleEntityAt(tileX, tileY int) *Entity {
	if s.world == nil {
		return nil
	}

	// Iterate through all entities to find one with matching tile coordinates
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

// createDestructibleEntity creates a new destructible entity for a tile.
func (s *TerrainModificationSystem) createDestructibleEntity(tileX, tileY int) *Entity {
	if s.world == nil {
		return nil
	}

	// Determine material type based on tile properties (default to stone)
	material := MaterialStone

	// Create entity
	entity := s.world.CreateEntity()

	// Add destructible component
	destComp := NewDestructibleComponent(material, tileX, tileY)
	entity.AddComponent(destComp)

	// Add position component (for spatial queries)
	posComp := &PositionComponent{
		X: float64(tileX*s.tileSize + s.tileSize/2),
		Y: float64(tileY*s.tileSize + s.tileSize/2),
	}
	entity.AddComponent(posComp)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"tileX":    tileX,
			"tileY":    tileY,
			"material": material.String(),
			"health":   destComp.MaxHealth,
		}).Debug("destructible entity created")
	}

	return entity
}

// replaceTileWithFloor replaces a destroyed tile with floor.
func (s *TerrainModificationSystem) replaceTileWithFloor(tileX, tileY int) {
	if s.terrain == nil || s.worldMap == nil {
		return
	}

	// Update terrain data
	s.terrain.SetTile(tileX, tileY, terrain.TileFloor)

	// Update world map
	tile := world.Tile{
		Type:     world.TileFloor,
		Walkable: true,
		X:        tileX,
		Y:        tileY,
	}
	s.worldMap.SetTile(tileX, tileY, tile)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"tileX": tileX,
			"tileY": tileY,
		}).Info("tile destroyed and replaced with floor")
	}
}

// DamageTileAtWorldPosition applies damage to a tile at world coordinates.
// This is a helper method for external systems (e.g., spell explosions).
func (s *TerrainModificationSystem) DamageTileAtWorldPosition(worldX, worldY, damage float64) {
	tileX := int(worldX / float64(s.tileSize))
	tileY := int(worldY / float64(s.tileSize))
	s.damageTile(tileX, tileY, damage)
}

// DamageTilesInArea applies damage to all tiles in a circular area.
// Useful for explosion effects.
func (s *TerrainModificationSystem) DamageTilesInArea(centerX, centerY, radius, damage float64) {
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
				s.damageTile(tileX, tileY, damage)
			}
		}
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"centerX": centerX,
			"centerY": centerY,
			"radius":  radius,
			"damage":  damage,
		}).Debug("area damage applied to tiles")
	}
}
