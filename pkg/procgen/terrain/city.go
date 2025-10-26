// Package terrain provides city generation with urban layouts.
// This file implements a city generation algorithm that creates
// urban environments with buildings, streets, and public spaces.
package terrain

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/sirupsen/logrus"
)

// CityGenerator generates urban environments with buildings and streets.
type CityGenerator struct {
	blockSize       int     // Size of city blocks (8-16 tiles)
	streetWidth     int     // Width of streets (2-3 tiles)
	buildingDensity float64 // Percentage of blocks with buildings (0.7 = 70%)
	plazaDensity    float64 // Percentage of blocks that are plazas (0.2 = 20%)
	logger          *logrus.Entry
}

// NewCityGenerator creates a new city generator with default parameters.
func NewCityGenerator() *CityGenerator {
	return NewCityGeneratorWithLogger(nil)
}

// NewCityGeneratorWithLogger creates a new city generator with a logger.
func NewCityGeneratorWithLogger(logger *logrus.Logger) *CityGenerator {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("generator", "city")
	}
	return &CityGenerator{
		blockSize:       12, // 12x12 tile blocks
		streetWidth:     2,  // 2-tile wide streets
		buildingDensity: 0.7,
		plazaDensity:    0.2,
		logger:          logEntry,
	}
}

// BlockType represents the type of a city block.
type BlockType int

const (
	BlockBuilding BlockType = iota
	BlockPlaza
	BlockPark
)

// CityBlock represents a subdivided block in the city grid.
type CityBlock struct {
	Rect      Rect      // Block boundaries
	BlockType BlockType // Type of block
}

// Rect represents a rectangular area.
type Rect struct {
	X, Y, Width, Height int
}

// Contains checks if a point is inside the rectangle.
func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.Width && y >= r.Y && y < r.Y+r.Height
}

// Center returns the center point of the rectangle.
func (r Rect) Center() (int, int) {
	return r.X + r.Width/2, r.Y + r.Height/2
}

// Generate creates a city environment with buildings, streets, and public spaces.
func (g *CityGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"seed":       seed,
			"genreID":    params.GenreID,
			"depth":      params.Depth,
			"difficulty": params.Difficulty,
		}).Debug("starting city terrain generation")
	}

	// Use custom parameters if provided
	width := 80
	height := 50
	if params.Custom != nil {
		if w, ok := params.Custom["width"].(int); ok {
			width = w
		}
		if h, ok := params.Custom["height"].(int); ok {
			height = h
		}
		if bs, ok := params.Custom["blockSize"].(int); ok {
			g.blockSize = bs
		}
		if sw, ok := params.Custom["streetWidth"].(int); ok {
			g.streetWidth = sw
		}
		if bd, ok := params.Custom["buildingDensity"].(float64); ok {
			g.buildingDensity = bd
		}
		if pd, ok := params.Custom["plazaDensity"].(float64); ok {
			g.plazaDensity = pd
		}
	}

	// Validate dimensions
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid dimensions: width=%d, height=%d (must be positive)", width, height)
	}

	if width > 1000 || height > 1000 {
		return nil, fmt.Errorf("dimensions too large: width=%d, height=%d (max 1000x1000)", width, height)
	}

	// Validate block size and street width
	if g.blockSize < 4 || g.blockSize > 30 {
		return nil, fmt.Errorf("invalid block size: %d (must be 4-30)", g.blockSize)
	}

	if g.streetWidth < 1 || g.streetWidth > 5 {
		return nil, fmt.Errorf("invalid street width: %d (must be 1-5)", g.streetWidth)
	}

	// Create RNG with seed
	rng := rand.New(rand.NewSource(seed))

	// Create terrain (starts with all walls)
	terrain := NewTerrain(width, height, seed)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			terrain.SetTile(x, y, TileWall)
		}
	}

	// Subdivide into city blocks
	blocks := g.subdivideGrid(terrain, rng)

	// Create street network first (before buildings)
	g.createStreetNetwork(blocks, terrain)

	// Determine block types
	g.assignBlockTypes(blocks, rng)

	// Place buildings, plazas, and parks
	g.placeBuildings(blocks, terrain, rng)

	// Add alleys between some buildings
	g.addAlleys(blocks, terrain, rng)

	// Place stairs in central plaza or large building
	g.placeStairs(blocks, terrain, rng)

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"width":  terrain.Width,
			"height": terrain.Height,
			"blocks": len(blocks),
		}).Info("city terrain generation complete")
	}

	return terrain, nil
}

// subdivideGrid divides the map into city blocks separated by streets.
func (g *CityGenerator) subdivideGrid(terrain *Terrain, rng *rand.Rand) []*CityBlock {
	blocks := make([]*CityBlock, 0)

	// Calculate how many blocks fit
	blockWithStreet := g.blockSize + g.streetWidth
	blocksX := (terrain.Width - g.streetWidth) / blockWithStreet
	blocksY := (terrain.Height - g.streetWidth) / blockWithStreet

	if blocksX <= 0 || blocksY <= 0 {
		// Map too small for blocks, create one large block
		blocks = append(blocks, &CityBlock{
			Rect: Rect{
				X:      1,
				Y:      1,
				Width:  terrain.Width - 2,
				Height: terrain.Height - 2,
			},
			BlockType: BlockBuilding,
		})
		return blocks
	}

	// Create grid of blocks
	for by := 0; by < blocksY; by++ {
		for bx := 0; bx < blocksX; bx++ {
			x := bx*blockWithStreet + g.streetWidth
			y := by*blockWithStreet + g.streetWidth
			width := g.blockSize
			height := g.blockSize

			// Adjust last blocks to fill remaining space
			if bx == blocksX-1 {
				width = terrain.Width - x - g.streetWidth
			}
			if by == blocksY-1 {
				height = terrain.Height - y - g.streetWidth
			}

			// Ensure minimum block size
			if width >= 4 && height >= 4 {
				blocks = append(blocks, &CityBlock{
					Rect: Rect{
						X:      x,
						Y:      y,
						Width:  width,
						Height: height,
					},
					BlockType: BlockBuilding, // Default, will be assigned later
				})
			}
		}
	}

	return blocks
}

// createStreetNetwork creates the street grid between blocks.
func (g *CityGenerator) createStreetNetwork(blocks []*CityBlock, terrain *Terrain) {
	// Fill all areas not in blocks with floor tiles (streets)
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			inBlock := false
			for _, block := range blocks {
				if block.Rect.Contains(x, y) {
					inBlock = true
					break
				}
			}

			if !inBlock {
				// This is a street
				terrain.SetTile(x, y, TileCorridor)
			}
		}
	}
}

// assignBlockTypes determines the type of each city block based on density parameters.
func (g *CityGenerator) assignBlockTypes(blocks []*CityBlock, rng *rand.Rand) {
	for _, block := range blocks {
		roll := rng.Float64()

		if roll < g.buildingDensity {
			block.BlockType = BlockBuilding
		} else if roll < g.buildingDensity+g.plazaDensity {
			block.BlockType = BlockPlaza
		} else {
			block.BlockType = BlockPark
		}
	}
}

// placeBuildings places buildings, plazas, and parks in blocks.
func (g *CityGenerator) placeBuildings(blocks []*CityBlock, terrain *Terrain, rng *rand.Rand) {
	for _, block := range blocks {
		switch block.BlockType {
		case BlockBuilding:
			g.createBuilding(block, terrain, rng)
		case BlockPlaza:
			g.createPlaza(block, terrain)
		case BlockPark:
			g.createPark(block, terrain, rng)
		}
	}
}

// createBuilding fills a block with a building structure and optionally adds interior.
func (g *CityGenerator) createBuilding(block *CityBlock, terrain *Terrain, rng *rand.Rand) {
	rect := block.Rect

	// Determine if this is a small or large building
	area := rect.Width * rect.Height
	isLargeBuilding := area >= 100 // 10x10 or larger

	if isLargeBuilding {
		// Large building: create interior with rooms
		g.createBuildingInterior(rect, terrain, rng)
	} else {
		// Small building: solid structure or single room
		if rng.Float64() < 0.7 {
			// 70%: Solid building (no interior access)
			for y := rect.Y; y < rect.Y+rect.Height; y++ {
				for x := rect.X; x < rect.X+rect.Width; x++ {
					terrain.SetTile(x, y, TileStructure)
				}
			}
		} else {
			// 30%: Single room with entrance
			// Walls
			for y := rect.Y; y < rect.Y+rect.Height; y++ {
				for x := rect.X; x < rect.X+rect.Width; x++ {
					if x == rect.X || x == rect.X+rect.Width-1 ||
						y == rect.Y || y == rect.Y+rect.Height-1 {
						terrain.SetTile(x, y, TileWall)
					} else {
						terrain.SetTile(x, y, TileFloor)
					}
				}
			}

			// Add door on random side
			g.addEntranceDoor(rect, terrain, rng)
		}
	}
}

// createBuildingInterior generates interior rooms for large buildings using BSP.
func (g *CityGenerator) createBuildingInterior(rect Rect, terrain *Terrain, rng *rand.Rand) {
	// Create outer walls
	for y := rect.Y; y < rect.Y+rect.Height; y++ {
		for x := rect.X; x < rect.X+rect.Width; x++ {
			if x == rect.X || x == rect.X+rect.Width-1 ||
				y == rect.Y || y == rect.Y+rect.Height-1 {
				terrain.SetTile(x, y, TileWall)
			} else {
				terrain.SetTile(x, y, TileFloor)
			}
		}
	}

	// BSP subdivide interior into rooms (simple version)
	if rect.Width >= 8 && rect.Height >= 8 {
		g.subdivideInterior(rect, terrain, rng, 0)
	}

	// Add entrance door
	g.addEntranceDoor(rect, terrain, rng)
}

// subdivideInterior recursively subdivides building interior into rooms.
func (g *CityGenerator) subdivideInterior(rect Rect, terrain *Terrain, rng *rand.Rand, depth int) {
	// Stop recursion if too small or max depth reached
	if rect.Width < 6 || rect.Height < 6 || depth >= 2 {
		return
	}

	// Decide split direction
	horizontal := rect.Width <= rect.Height
	if rect.Width > 6 && rect.Height > 6 {
		horizontal = rng.Float64() < 0.5
	}

	if horizontal {
		// Split horizontally
		splitY := rect.Y + 3 + rng.Intn(rect.Height-5)

		// Create wall
		for x := rect.X + 1; x < rect.X+rect.Width-1; x++ {
			terrain.SetTile(x, splitY, TileWall)
		}

		// Add door
		doorX := rect.X + 1 + rng.Intn(rect.Width-2)
		terrain.SetTile(doorX, splitY, TileDoor)

		// Recurse
		top := Rect{X: rect.X, Y: rect.Y, Width: rect.Width, Height: splitY - rect.Y}
		bottom := Rect{X: rect.X, Y: splitY + 1, Width: rect.Width, Height: rect.Y + rect.Height - splitY - 1}
		g.subdivideInterior(top, terrain, rng, depth+1)
		g.subdivideInterior(bottom, terrain, rng, depth+1)
	} else {
		// Split vertically
		splitX := rect.X + 3 + rng.Intn(rect.Width-5)

		// Create wall
		for y := rect.Y + 1; y < rect.Y+rect.Height-1; y++ {
			terrain.SetTile(splitX, y, TileWall)
		}

		// Add door
		doorY := rect.Y + 1 + rng.Intn(rect.Height-2)
		terrain.SetTile(splitX, doorY, TileDoor)

		// Recurse
		left := Rect{X: rect.X, Y: rect.Y, Width: splitX - rect.X, Height: rect.Height}
		right := Rect{X: splitX + 1, Y: rect.Y, Width: rect.X + rect.Width - splitX - 1, Height: rect.Height}
		g.subdivideInterior(left, terrain, rng, depth+1)
		g.subdivideInterior(right, terrain, rng, depth+1)
	}
}

// addEntranceDoor adds an entrance door to a building on a random side.
func (g *CityGenerator) addEntranceDoor(rect Rect, terrain *Terrain, rng *rand.Rand) {
	// Choose random side (0=top, 1=right, 2=bottom, 3=left)
	side := rng.Intn(4)

	switch side {
	case 0: // Top
		doorX := rect.X + 1 + rng.Intn(rect.Width-2)
		if terrain.GetTile(doorX, rect.Y-1) == TileCorridor {
			terrain.SetTile(doorX, rect.Y, TileDoor)
		}
	case 1: // Right
		doorY := rect.Y + 1 + rng.Intn(rect.Height-2)
		if terrain.GetTile(rect.X+rect.Width, doorY) == TileCorridor {
			terrain.SetTile(rect.X+rect.Width-1, doorY, TileDoor)
		}
	case 2: // Bottom
		doorX := rect.X + 1 + rng.Intn(rect.Width-2)
		if terrain.GetTile(doorX, rect.Y+rect.Height) == TileCorridor {
			terrain.SetTile(doorX, rect.Y+rect.Height-1, TileDoor)
		}
	case 3: // Left
		doorY := rect.Y + 1 + rng.Intn(rect.Height-2)
		if terrain.GetTile(rect.X-1, doorY) == TileCorridor {
			terrain.SetTile(rect.X, doorY, TileDoor)
		}
	}
}

// createPlaza creates an open public square.
func (g *CityGenerator) createPlaza(block *CityBlock, terrain *Terrain) {
	rect := block.Rect

	// Fill with floor tiles
	for y := rect.Y; y < rect.Y+rect.Height; y++ {
		for x := rect.X; x < rect.X+rect.Width; x++ {
			terrain.SetTile(x, y, TileFloor)
		}
	}

	// Track this as a "room" for stairs placement
	room := &Room{
		X:      rect.X,
		Y:      rect.Y,
		Width:  rect.Width,
		Height: rect.Height,
		Type:   RoomNormal,
	}
	terrain.Rooms = append(terrain.Rooms, room)
}

// createPark creates a park with trees and/or water features.
func (g *CityGenerator) createPark(block *CityBlock, terrain *Terrain, rng *rand.Rand) {
	rect := block.Rect

	// Fill with floor tiles (grass)
	for y := rect.Y; y < rect.Y+rect.Height; y++ {
		for x := rect.X; x < rect.X+rect.Width; x++ {
			terrain.SetTile(x, y, TileFloor)
		}
	}

	// Add trees (30% of park tiles)
	treeCount := int(float64(rect.Width*rect.Height) * 0.3)
	for i := 0; i < treeCount; i++ {
		x := rect.X + rng.Intn(rect.Width)
		y := rect.Y + rng.Intn(rect.Height)
		if terrain.GetTile(x, y) == TileFloor {
			terrain.SetTile(x, y, TileTree)
		}
	}

	// Sometimes add a small pond (20% chance)
	if rng.Float64() < 0.2 && rect.Width >= 6 && rect.Height >= 6 {
		cx := rect.X + rect.Width/2
		cy := rect.Y + rect.Height/2
		radius := 2 + rng.Intn(2) // 2-3 tile radius

		for dy := -radius; dy <= radius; dy++ {
			for dx := -radius; dx <= radius; dx++ {
				if dx*dx+dy*dy <= radius*radius {
					x := cx + dx
					y := cy + dy
					if rect.Contains(x, y) && terrain.GetTile(x, y) != TileTree {
						if dx*dx+dy*dy < (radius-1)*(radius-1) {
							terrain.SetTile(x, y, TileWaterDeep)
						} else {
							terrain.SetTile(x, y, TileWaterShallow)
						}
					}
				}
			}
		}
	}
}

// addAlleys creates narrow passages between some buildings.
func (g *CityGenerator) addAlleys(blocks []*CityBlock, terrain *Terrain, rng *rand.Rand) {
	// Find adjacent building blocks and add alleys (30% chance)
	for i := 0; i < len(blocks); i++ {
		if blocks[i].BlockType != BlockBuilding {
			continue
		}

		for j := i + 1; j < len(blocks); j++ {
			if blocks[j].BlockType != BlockBuilding {
				continue
			}

			// Check if blocks are adjacent
			r1 := blocks[i].Rect
			r2 := blocks[j].Rect

			// Horizontal adjacency (side by side)
			if r1.Y == r2.Y && r1.Height == r2.Height {
				if r1.X+r1.Width == r2.X-g.streetWidth || r2.X+r2.Width == r1.X-g.streetWidth {
					// Already have street between them, skip
					continue
				}
			}

			// Vertical adjacency (top/bottom)
			if r1.X == r2.X && r1.Width == r2.Width {
				if r1.Y+r1.Height == r2.Y-g.streetWidth || r2.Y+r2.Height == r1.Y-g.streetWidth {
					// Already have street between them, skip
					continue
				}
			}
		}
	}

	// Note: Alley generation deferred to keep implementation simple
	// Streets already provide full connectivity
}

// placeStairs places stairs in the largest plaza or a large building.
func (g *CityGenerator) placeStairs(blocks []*CityBlock, terrain *Terrain, rng *rand.Rand) {
	// Find all plazas
	plazas := make([]*CityBlock, 0)
	for _, block := range blocks {
		if block.BlockType == BlockPlaza {
			plazas = append(plazas, block)
		}
	}

	// Place stairs in plazas if available
	if len(plazas) > 0 {
		// Sort by area (largest first)
		for i := 0; i < len(plazas); i++ {
			for j := i + 1; j < len(plazas); j++ {
				area1 := plazas[i].Rect.Width * plazas[i].Rect.Height
				area2 := plazas[j].Rect.Width * plazas[j].Rect.Height
				if area2 > area1 {
					plazas[i], plazas[j] = plazas[j], plazas[i]
				}
			}
		}

		// Place stairs up in largest plaza
		cx, cy := plazas[0].Rect.Center()
		terrain.AddStairs(cx, cy, true)

		// Place stairs down in second plaza or opposite corner of same plaza
		if len(plazas) > 1 {
			cx, cy = plazas[1].Rect.Center()
			terrain.AddStairs(cx, cy, false)
		} else {
			// Use opposite corner of same plaza
			rect := plazas[0].Rect
			x := rect.X + 1
			y := rect.Y + 1
			terrain.AddStairs(x, y, false)
		}
	} else {
		// No plazas, place in first and last block centers
		if len(blocks) > 0 {
			cx, cy := blocks[0].Rect.Center()
			// Find walkable spot near center
			for dy := 0; dy < 3; dy++ {
				for dx := 0; dx < 3; dx++ {
					x := cx + dx - 1
					y := cy + dy - 1
					if terrain.IsWalkable(x, y) {
						terrain.AddStairs(x, y, true)
						goto foundUp
					}
				}
			}
		foundUp:

			if len(blocks) > 1 {
				cx, cy := blocks[len(blocks)-1].Rect.Center()
				for dy := 0; dy < 3; dy++ {
					for dx := 0; dx < 3; dx++ {
						x := cx + dx - 1
						y := cy + dy - 1
						if terrain.IsWalkable(x, y) {
							terrain.AddStairs(x, y, false)
							goto foundDown
						}
					}
				}
			foundDown:
			}
		}
	}
}

// Validate checks if the generated city meets quality requirements.
func (g *CityGenerator) Validate(result interface{}) error {
	terrain, ok := result.(*Terrain)
	if !ok {
		return fmt.Errorf("result is not a Terrain")
	}

	// Count walkable tiles (streets + plazas + building interiors)
	walkable := 0
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			if terrain.IsWalkable(x, y) {
				walkable++
			}
		}
	}

	// Ensure at least 30% of tiles are walkable (cities should have good street coverage)
	totalTiles := terrain.Width * terrain.Height
	if float64(walkable)/float64(totalTiles) < 0.3 {
		return fmt.Errorf("insufficient walkable tiles: %d/%d (%.1f%%, need >= 30%%)",
			walkable, totalTiles, float64(walkable)/float64(totalTiles)*100)
	}

	// Validate stair placement if stairs exist
	if len(terrain.StairsUp) > 0 || len(terrain.StairsDown) > 0 {
		if err := terrain.ValidateStairPlacement(); err != nil {
			return fmt.Errorf("stair validation failed: %w", err)
		}
	}

	return nil
}
