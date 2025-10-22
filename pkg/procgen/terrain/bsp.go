// Package terrain provides BSP dungeon generation.
// This file implements Binary Space Partitioning algorithm for
// procedural dungeon layout generation.
package terrain


import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
)

// BSPGenerator generates dungeons using Binary Space Partitioning.
// It recursively splits the map into smaller regions and creates rooms.
type BSPGenerator struct {
	minRoomSize int
	maxRoomSize int
}

// NewBSPGenerator creates a new BSP dungeon generator.
func NewBSPGenerator() *BSPGenerator {
	return &BSPGenerator{
		minRoomSize: 6,
		maxRoomSize: 15,
	}
}

// bspNode represents a node in the BSP tree.
type bspNode struct {
	x, y          int
	width, height int
	left, right   *bspNode
	room          *Room
}

// Generate creates a dungeon using BSP algorithm.
func (g *BSPGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	// Use custom parameters if provided, otherwise use defaults
	width := 80
	height := 50
	if params.Custom != nil {
		if w, ok := params.Custom["width"].(int); ok {
			width = w
		}
		if h, ok := params.Custom["height"].(int); ok {
			height = h
		}
	}

	// Validate dimensions to prevent panic on slice allocation
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid dimensions: width and height must be positive (got width=%d, height=%d)", width, height)
	}

	// Set reasonable maximum to prevent memory exhaustion
	const maxDimension = 10000
	if width > maxDimension || height > maxDimension {
		return nil, fmt.Errorf("dimensions too large: maximum is %d (got width=%d, height=%d)", maxDimension, width, height)
	}

	// Create random source from seed for deterministic generation
	rng := rand.New(rand.NewSource(seed))

	// Create empty terrain
	terrain := NewTerrain(width, height, seed)

	// Create root BSP node
	root := &bspNode{
		x:      0,
		y:      0,
		width:  width,
		height: height,
	}

	// Recursively split the space
	g.splitNode(root, rng)

	// Create rooms in the leaf nodes
	g.createRooms(root, terrain, rng)

	// Connect rooms with corridors
	g.connectRooms(root, terrain)

	return terrain, nil
}

// splitNode recursively splits a BSP node into smaller nodes.
func (g *BSPGenerator) splitNode(node *bspNode, rng *rand.Rand) {
	// Stop splitting if the node is too small
	if node.width < g.minRoomSize*2+3 || node.height < g.minRoomSize*2+3 {
		return
	}

	// Decide whether to split horizontally or vertically
	splitHorizontally := rng.Float64() < 0.5

	// If the node is wide, prefer vertical split
	if float64(node.width) > float64(node.height)*1.25 {
		splitHorizontally = false
	} else if float64(node.height) > float64(node.width)*1.25 {
		splitHorizontally = true
	}

	if splitHorizontally {
		// Split horizontally
		splitPos := g.minRoomSize + rng.Intn(node.height-g.minRoomSize*2)

		node.left = &bspNode{
			x:      node.x,
			y:      node.y,
			width:  node.width,
			height: splitPos,
		}

		node.right = &bspNode{
			x:      node.x,
			y:      node.y + splitPos,
			width:  node.width,
			height: node.height - splitPos,
		}
	} else {
		// Split vertically
		splitPos := g.minRoomSize + rng.Intn(node.width-g.minRoomSize*2)

		node.left = &bspNode{
			x:      node.x,
			y:      node.y,
			width:  splitPos,
			height: node.height,
		}

		node.right = &bspNode{
			x:      node.x + splitPos,
			y:      node.y,
			width:  node.width - splitPos,
			height: node.height,
		}
	}

	// Recursively split child nodes
	g.splitNode(node.left, rng)
	g.splitNode(node.right, rng)
}

// createRooms creates rooms in the leaf nodes of the BSP tree.
func (g *BSPGenerator) createRooms(node *bspNode, terrain *Terrain, rng *rand.Rand) {
	if node.left != nil {
		g.createRooms(node.left, terrain, rng)
	}
	if node.right != nil {
		g.createRooms(node.right, terrain, rng)
	}

	// If this is a leaf node, create a room
	if node.left == nil && node.right == nil {
		// Determine room size (smaller than the node)
		maxWidth := min(g.maxRoomSize, node.width-2)
		maxHeight := min(g.maxRoomSize, node.height-2)

		// Ensure we have valid dimensions
		if maxWidth < g.minRoomSize {
			maxWidth = g.minRoomSize
		}
		if maxHeight < g.minRoomSize {
			maxHeight = g.minRoomSize
		}

		widthRange := maxWidth - g.minRoomSize + 1
		heightRange := maxHeight - g.minRoomSize + 1

		roomWidth := g.minRoomSize
		if widthRange > 0 {
			roomWidth += rng.Intn(widthRange)
		}

		roomHeight := g.minRoomSize
		if heightRange > 0 {
			roomHeight += rng.Intn(heightRange)
		}

		// Position room randomly within the node
		xRange := node.width - roomWidth - 1
		yRange := node.height - roomHeight - 1

		roomX := node.x + 1
		if xRange > 0 {
			roomX += rng.Intn(xRange)
		}

		roomY := node.y + 1
		if yRange > 0 {
			roomY += rng.Intn(yRange)
		}

		room := &Room{
			X:      roomX,
			Y:      roomY,
			Width:  roomWidth,
			Height: roomHeight,
		}

		node.room = room
		terrain.Rooms = append(terrain.Rooms, room)

		// Carve out the room
		for y := roomY; y < roomY+roomHeight; y++ {
			for x := roomX; x < roomX+roomWidth; x++ {
				terrain.SetTile(x, y, TileFloor)
			}
		}
	}
}

// connectRooms creates corridors between rooms in sibling nodes.
func (g *BSPGenerator) connectRooms(node *bspNode, terrain *Terrain) {
	if node.left == nil || node.right == nil {
		return
	}

	// Recursively connect rooms in child nodes first
	g.connectRooms(node.left, terrain)
	g.connectRooms(node.right, terrain)

	// Get representative rooms from left and right subtrees
	leftRoom := g.getRoom(node.left)
	rightRoom := g.getRoom(node.right)

	if leftRoom != nil && rightRoom != nil {
		// Create corridor between the two rooms
		x1, y1 := leftRoom.Center()
		x2, y2 := rightRoom.Center()

		g.createCorridor(terrain, x1, y1, x2, y2)
	}
}

// getRoom returns a room from the node or its descendants.
func (g *BSPGenerator) getRoom(node *bspNode) *Room {
	if node.room != nil {
		return node.room
	}
	if node.left != nil {
		if room := g.getRoom(node.left); room != nil {
			return room
		}
	}
	if node.right != nil {
		return g.getRoom(node.right)
	}
	return nil
}

// createCorridor carves a corridor between two points.
func (g *BSPGenerator) createCorridor(terrain *Terrain, x1, y1, x2, y2 int) {
	// Create L-shaped corridor
	// First horizontal, then vertical
	for x := min(x1, x2); x <= max(x1, x2); x++ {
		terrain.SetTile(x, y1, TileCorridor)
	}
	for y := min(y1, y2); y <= max(y1, y2); y++ {
		terrain.SetTile(x2, y, TileCorridor)
	}
}

// Validate checks if the generated terrain is valid.
func (g *BSPGenerator) Validate(result interface{}) error {
	terrain, ok := result.(*Terrain)
	if !ok {
		return fmt.Errorf("result is not a Terrain")
	}

	if len(terrain.Rooms) == 0 {
		return fmt.Errorf("no rooms generated")
	}

	// Check that all rooms are within bounds
	for _, room := range terrain.Rooms {
		if room.X < 0 || room.Y < 0 ||
			room.X+room.Width > terrain.Width ||
			room.Y+room.Height > terrain.Height {
			return fmt.Errorf("room out of bounds")
		}
	}

	return nil
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
