// Package engine provides the terrain construction system for building walls.
// This file implements TerrainConstructionSystem which handles placing buildable
// markers, consuming materials from inventory, and creating walls when construction completes.
//
// Design Philosophy:
// - Server-authoritative for multiplayer (network sync required)
// - Validates placement (tile must be walkable, not occupied)
// - Consumes materials from inventory before starting construction
// - BuildableComponent tracks progress (3 seconds default)
// - Performance target: <1ms per frame for validation checks
package engine

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/world"
	"github.com/sirupsen/logrus"
)

// TerrainConstructionSystem handles wall building and construction.
type TerrainConstructionSystem struct {
	world    *World
	terrain  *terrain.Terrain
	worldMap *world.Map
	tileSize int
	logger   *logrus.Entry
}

// NewTerrainConstructionSystem creates a new terrain construction system.
func NewTerrainConstructionSystem(tileSize int) *TerrainConstructionSystem {
	return NewTerrainConstructionSystemWithLogger(tileSize, nil)
}

// NewTerrainConstructionSystemWithLogger creates a system with a logger.
func NewTerrainConstructionSystemWithLogger(tileSize int, logger *logrus.Logger) *TerrainConstructionSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"system":   "terrain_construction",
			"tileSize": tileSize,
		})
		logEntry.Debug("terrain construction system created")
	}

	return &TerrainConstructionSystem{
		tileSize: tileSize,
		logger:   logEntry,
	}
}

// SetWorld sets the ECS world reference.
func (s *TerrainConstructionSystem) SetWorld(world *World) {
	s.world = world
}

// SetTerrain sets the terrain data reference.
func (s *TerrainConstructionSystem) SetTerrain(t *terrain.Terrain) {
	s.terrain = t
}

// SetWorldMap sets the world map reference for tile modification.
func (s *TerrainConstructionSystem) SetWorldMap(m *world.Map) {
	s.worldMap = m
}

// Update implements the System interface.
// Processes buildable entities and completes construction.
func (s *TerrainConstructionSystem) Update(entities []*Entity, deltaTime float64) {
	if s.terrain == nil || s.worldMap == nil {
		return
	}

	// Update buildable entities
	for _, entity := range entities {
		if comp, ok := entity.GetComponent("buildable"); ok {
			if buildComp, ok := comp.(*BuildableComponent); ok {
				buildComp.Update(deltaTime)

				if buildComp.IsComplete {
					s.completeConstruction(entity, buildComp)
				}
			}
		}
	}
}

// StartConstruction begins construction at a tile location.
// Validates placement, checks materials, and consumes them from inventory.
func (s *TerrainConstructionSystem) StartConstruction(builderEntity *Entity, tileX, tileY int, resultType world.TileType) error {
	// Validate placement
	if err := s.validatePlacement(tileX, tileY); err != nil {
		return fmt.Errorf("invalid placement: %w", err)
	}

	// Get builder's inventory
	invComp, err := s.getInventoryComponent(builderEntity)
	if err != nil {
		return fmt.Errorf("builder has no inventory: %w", err)
	}

	// Create buildable component to check material requirements
	buildComp := NewBuildableComponent(tileX, tileY, resultType, 3.0)

	// Check if builder has required materials
	if err := s.checkMaterials(invComp, buildComp.RequiredMaterials); err != nil {
		return fmt.Errorf("insufficient materials: %w", err)
	}

	// Consume materials
	s.consumeMaterials(invComp, buildComp.RequiredMaterials)

	// Create buildable entity
	s.createBuildableEntity(tileX, tileY, buildComp)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"tileX": tileX,
			"tileY": tileY,
			"type":  resultType,
		}).Info("construction started")
	}

	return nil
}

// validatePlacement checks if a tile location is valid for construction.
func (s *TerrainConstructionSystem) validatePlacement(tileX, tileY int) error {
	if s.terrain == nil {
		return fmt.Errorf("no terrain set")
	}

	// Check bounds
	if !s.terrain.IsInBounds(tileX, tileY) {
		return fmt.Errorf("out of bounds")
	}

	// Tile must be walkable (floor)
	tileType := s.terrain.GetTile(tileX, tileY)
	if tileType != terrain.TileFloor {
		return fmt.Errorf("tile must be walkable floor")
	}

	// Check if already occupied by a buildable entity
	if s.findBuildableEntityAt(tileX, tileY) != nil {
		return fmt.Errorf("construction already in progress")
	}

	return nil
}

// getInventoryComponent retrieves inventory from an entity.
func (s *TerrainConstructionSystem) getInventoryComponent(entity *Entity) (*InventoryComponent, error) {
	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return nil, fmt.Errorf("no inventory component")
	}

	invComp, ok := comp.(*InventoryComponent)
	if !ok {
		return nil, fmt.Errorf("invalid inventory component type")
	}

	return invComp, nil
}

// checkMaterials verifies entity has required materials.
func (s *TerrainConstructionSystem) checkMaterials(inv *InventoryComponent, required map[MaterialType]int) error {
	materialCounts := s.countMaterialsInInventory(inv)

	for material, needed := range required {
		if materialCounts[material] < needed {
			return fmt.Errorf("need %d %s, have %d", needed, material.String(), materialCounts[material])
		}
	}

	return nil
}

// countMaterialsInInventory counts construction materials in inventory.
func (s *TerrainConstructionSystem) countMaterialsInInventory(inv *InventoryComponent) map[MaterialType]int {
	counts := make(map[MaterialType]int)

	// For now, check item names for material types
	// Future: add explicit material field to item.Item
	for _, itm := range inv.Items {
		if itm == nil {
			continue
		}

		material := s.getMaterialFromItemName(itm.Name)
		counts[material]++
	}

	// Also check gold as stone equivalent (1 gold = 1 stone for now)
	counts[MaterialStone] += inv.Gold / 10 // 10 gold = 1 stone equivalent

	return counts
}

// getMaterialFromItemName determines material type from item name.
func (s *TerrainConstructionSystem) getMaterialFromItemName(name string) MaterialType {
	// Simple name-based material detection
	// Future: use explicit material field on items
	switch {
	case len(name) > 4 && name[:4] == "Wood":
		return MaterialWood
	case len(name) > 5 && name[:5] == "Stone":
		return MaterialStone
	case len(name) > 5 && name[:5] == "Metal":
		return MaterialMetal
	default:
		return MaterialStone // Default fallback
	}
}

// consumeMaterials removes materials from inventory.
func (s *TerrainConstructionSystem) consumeMaterials(inv *InventoryComponent, required map[MaterialType]int) {
	for material, needed := range required {
		remaining := needed

		// Remove items
		for i := len(inv.Items) - 1; i >= 0 && remaining > 0; i-- {
			if inv.Items[i] == nil {
				continue
			}

			itemMaterial := s.getMaterialFromItemName(inv.Items[i].Name)
			if itemMaterial == material {
				inv.Items = append(inv.Items[:i], inv.Items[i+1:]...)
				remaining--
			}
		}

		// Deduct from gold if using stone equivalent
		if material == MaterialStone && remaining > 0 {
			goldCost := remaining * 10
			if inv.Gold >= goldCost {
				inv.Gold -= goldCost
				remaining = 0
			}
		}
	}
}

// createBuildableEntity creates a buildable entity at a tile.
func (s *TerrainConstructionSystem) createBuildableEntity(tileX, tileY int, buildComp *BuildableComponent) {
	if s.world == nil {
		return
	}

	entity := s.world.CreateEntity()
	entity.AddComponent(buildComp)

	// Add position component
	posComp := &PositionComponent{
		X: float64(tileX*s.tileSize + s.tileSize/2),
		Y: float64(tileY*s.tileSize + s.tileSize/2),
	}
	entity.AddComponent(posComp)
}

// completeConstruction finishes construction and places the tile.
func (s *TerrainConstructionSystem) completeConstruction(entity *Entity, buildComp *BuildableComponent) {
	if s.terrain == nil || s.worldMap == nil {
		return
	}

	// Place the tile
	tileX, tileY := buildComp.TileX, buildComp.TileY

	// Convert world.TileType to terrain.TileType
	terrainType := s.worldTileToTerrainTile(buildComp.ResultTileType)
	s.terrain.SetTile(tileX, tileY, terrainType)

	// Update world map
	tile := world.Tile{
		Type:     buildComp.ResultTileType,
		Walkable: false, // Walls are not walkable
		X:        tileX,
		Y:        tileY,
	}
	s.worldMap.SetTile(tileX, tileY, tile)

	// Remove buildable entity
	if s.world != nil {
		s.world.RemoveEntity(entity.ID)
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"tileX": tileX,
			"tileY": tileY,
			"type":  buildComp.ResultTileType,
		}).Info("construction completed")
	}
}

// worldTileToTerrainTile converts world.TileType to terrain.TileType.
func (s *TerrainConstructionSystem) worldTileToTerrainTile(wt world.TileType) terrain.TileType {
	switch wt {
	case world.TileWall:
		return terrain.TileWall
	case world.TileFloor:
		return terrain.TileFloor
	case world.TileDoor:
		return terrain.TileDoor
	default:
		return terrain.TileWall
	}
}

// findBuildableEntityAt finds a buildable entity at tile coordinates.
func (s *TerrainConstructionSystem) findBuildableEntityAt(tileX, tileY int) *Entity {
	if s.world == nil {
		return nil
	}

	for _, entity := range s.world.GetEntities() {
		if comp, ok := entity.GetComponent("buildable"); ok {
			if buildComp, ok := comp.(*BuildableComponent); ok {
				if buildComp.TileX == tileX && buildComp.TileY == tileY {
					return entity
				}
			}
		}
	}

	return nil
}

// GetConstructionProgress returns the progress percentage (0.0-1.0) for construction at a tile.
func (s *TerrainConstructionSystem) GetConstructionProgress(tileX, tileY int) float64 {
	entity := s.findBuildableEntityAt(tileX, tileY)
	if entity == nil {
		return 0.0
	}

	if comp, ok := entity.GetComponent("buildable"); ok {
		if buildComp, ok := comp.(*BuildableComponent); ok {
			if buildComp.ConstructionTime <= 0 {
				return 1.0
			}
			progress := buildComp.ElapsedTime / buildComp.ConstructionTime
			if progress > 1.0 {
				return 1.0
			}
			return progress
		}
	}

	return 0.0
}
