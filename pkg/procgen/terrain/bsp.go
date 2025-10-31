// Package terrain provides BSP dungeon generation.
// This file implements Binary Space Partitioning algorithm for
// procedural dungeon layout generation.
package terrain

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/sirupsen/logrus"
)

// BSPGenerator generates dungeons using Binary Space Partitioning.
// It recursively splits the map into smaller regions and creates rooms.
type BSPGenerator struct {
	minRoomSize int
	maxRoomSize int
	logger      *logrus.Entry
}

// NewBSPGenerator creates a new BSP dungeon generator.
func NewBSPGenerator() *BSPGenerator {
	return NewBSPGeneratorWithLogger(nil)
}

// NewBSPGeneratorWithLogger creates a new BSP dungeon generator with a logger.
func NewBSPGeneratorWithLogger(logger *logrus.Logger) *BSPGenerator {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("generator", "BSP")
	}
	return &BSPGenerator{
		minRoomSize: 6,
		maxRoomSize: 15,
		logger:      logEntry,
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
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"seed":       seed,
			"genreID":    params.GenreID,
			"depth":      params.Depth,
			"difficulty": params.Difficulty,
		}).Debug("starting BSP terrain generation")
	}

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

	// GAP-006 REPAIR: Assign special room types
	g.assignRoomTypes(terrain, rng)

	// Add water features (moats around boss rooms)
	g.addWaterFeatures(terrain, rng)

	// Phase 11.1: Add multi-layer features (platforms, pits, lava flows)
	g.addMultiLayerFeatures(terrain, rng)

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"width":     terrain.Width,
			"height":    terrain.Height,
			"roomCount": len(terrain.Rooms),
			"seed":      seed,
		}).Info("BSP terrain generation complete")
	}

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

		// Phase 11.1: Add diagonal walls by chamfering corners (20-40% chance per room)
		if rng.Float64() < 0.30 { // 30% of rooms get diagonal corners
			g.chamferRoomCorners(terrain, room, rng)
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

// assignRoomTypes assigns special purposes to rooms in the dungeon.
// Ensures dungeons have spawn, exit, boss, treasure, and trap rooms.
func (g *BSPGenerator) assignRoomTypes(terrain *Terrain, rng *rand.Rand) {
	numRooms := len(terrain.Rooms)
	if numRooms == 0 {
		return
	}

	// First room is always the spawn point
	terrain.Rooms[0].Type = RoomSpawn

	// Last room (farthest from spawn) is the exit
	if numRooms > 1 {
		terrain.Rooms[numRooms-1].Type = RoomExit
	}

	// If we have at least 3 rooms, assign a boss room
	// Boss room should be near the end but not the exit
	if numRooms >= 3 {
		bossIdx := numRooms - 1 - rng.Intn(min(3, numRooms-1))
		if bossIdx == numRooms-1 && numRooms > 2 {
			bossIdx = numRooms - 2
		}
		terrain.Rooms[bossIdx].Type = RoomBoss
	}

	// Assign treasure rooms (10-20% of remaining rooms)
	numTreasure := max(1, numRooms/8)
	treasureAssigned := 0
	for i := 1; i < numRooms-1 && treasureAssigned < numTreasure; i++ {
		// Skip if already assigned a special type
		if terrain.Rooms[i].Type != RoomNormal {
			continue
		}
		// 30% chance to be a treasure room
		if rng.Float64() < 0.3 {
			terrain.Rooms[i].Type = RoomTreasure
			treasureAssigned++
		}
	}

	// Assign trap rooms (10-15% of remaining rooms)
	numTraps := max(1, numRooms/10)
	trapsAssigned := 0
	for i := 1; i < numRooms-1 && trapsAssigned < numTraps; i++ {
		// Skip if already assigned a special type
		if terrain.Rooms[i].Type != RoomNormal {
			continue
		}
		// 25% chance to be a trap room
		if rng.Float64() < 0.25 {
			terrain.Rooms[i].Type = RoomTrap
			trapsAssigned++
		}
	}

	// All remaining rooms stay as RoomNormal (already default)
}

// addWaterFeatures adds water features like moats to special rooms.
// Boss rooms get moats for dramatic effect and tactical challenge.
func (g *BSPGenerator) addWaterFeatures(terrain *Terrain, rng *rand.Rand) {
	// Add moats around boss rooms (if they exist and are large enough)
	for _, room := range terrain.Rooms {
		if room.Type == RoomBoss {
			// Only add moat if room is large enough (at least 8x8)
			if room.Width >= 8 && room.Height >= 8 {
				// Moat width: 1-2 tiles based on room size
				moatWidth := 1
				if room.Width >= 12 && room.Height >= 12 {
					moatWidth = 2
				}

				// Generate moat around boss room
				_ = GenerateMoat(room, moatWidth, terrain)
			}
		}
	}
}

// chamferRoomCorners adds diagonal walls to room corners (Phase 11.1).
// This creates 45° angle cuts on corners, making rooms more visually interesting.
func (g *BSPGenerator) chamferRoomCorners(terrain *Terrain, room *Room, rng *rand.Rand) {
	// Determine chamfer size (1-2 tiles)
	chamferSize := 1 + rng.Intn(2)

	// Track which corners to chamfer (at least 1, up to all 4)
	numCorners := 1 + rng.Intn(4) // 1-4 corners
	corners := []bool{false, false, false, false}
	for i := 0; i < numCorners; i++ {
		corners[rng.Intn(4)] = true
	}

	// Top-left corner (NW) - use TileWallSE (\ shape)
	if corners[0] {
		for i := 0; i < chamferSize; i++ {
			x := room.X + i
			y := room.Y + i
			if terrain.IsInBounds(x, y) {
				terrain.SetTile(x, y, TileWallSE)
			}
		}
	}

	// Top-right corner (NE) - use TileWallSW (/ shape)
	if corners[1] {
		for i := 0; i < chamferSize; i++ {
			x := room.X + room.Width - 1 - i
			y := room.Y + i
			if terrain.IsInBounds(x, y) {
				terrain.SetTile(x, y, TileWallSW)
			}
		}
	}

	// Bottom-left corner (SW) - use TileWallNE (/ shape)
	if corners[2] {
		for i := 0; i < chamferSize; i++ {
			x := room.X + i
			y := room.Y + room.Height - 1 - i
			if terrain.IsInBounds(x, y) {
				terrain.SetTile(x, y, TileWallNE)
			}
		}
	}

	// Bottom-right corner (SE) - use TileWallNW (\ shape)
	if corners[3] {
		for i := 0; i < chamferSize; i++ {
			x := room.X + room.Width - 1 - i
			y := room.Y + room.Height - 1 - i
			if terrain.IsInBounds(x, y) {
				terrain.SetTile(x, y, TileWallNW)
			}
		}
	}
}

// addMultiLayerFeatures adds platforms, pits, and ramps to rooms (Phase 11.1).
// This creates vertical variety with elevated platforms and chasms.
func (g *BSPGenerator) addMultiLayerFeatures(terrain *Terrain, rng *rand.Rand) {
	// 30-50% of rooms get multi-layer features
	for _, room := range terrain.Rooms {
		// Skip small rooms
		if room.Width < 8 || room.Height < 8 {
			continue
		}

		featureType := rng.Float64()

		if featureType < 0.15 { // 15% chance: Central platform
			g.addCentralPlatform(terrain, room, rng)
		} else if featureType < 0.25 { // 10% chance: Corner pits
			g.addCornerPits(terrain, room, rng)
		} else if featureType < 0.35 { // 10% chance: Lava flow
			g.addLavaFlow(terrain, room, rng)
		}
	}
}

// addCentralPlatform adds an elevated platform in the center of a room.
func (g *BSPGenerator) addCentralPlatform(terrain *Terrain, room *Room, rng *rand.Rand) {
	// Platform size: 30-60% of room size
	platformWidth := room.Width * (30 + rng.Intn(31)) / 100
	platformHeight := room.Height * (30 + rng.Intn(31)) / 100

	// Ensure minimum size
	if platformWidth < 3 {
		platformWidth = 3
	}
	if platformHeight < 3 {
		platformHeight = 3
	}

	// Center the platform
	platformX := room.X + (room.Width-platformWidth)/2
	platformY := room.Y + (room.Height-platformHeight)/2

	// Create platform
	for y := platformY; y < platformY+platformHeight; y++ {
		for x := platformX; x < platformX+platformWidth; x++ {
			if terrain.IsInBounds(x, y) {
				terrain.SetTile(x, y, TilePlatform)
			}
		}
	}

	// Add ramps on 1-2 sides
	numRamps := 1 + rng.Intn(2)
	sides := []int{0, 1, 2, 3} // North, East, South, West
	// Shuffle sides
	for i := len(sides) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		sides[i], sides[j] = sides[j], sides[i]
	}

	for i := 0; i < numRamps && i < len(sides); i++ {
		side := sides[i]
		switch side {
		case 0: // North
			rampX := platformX + platformWidth/2
			if terrain.IsInBounds(rampX, platformY-1) {
				terrain.SetTile(rampX, platformY-1, TileRampUp)
			}
		case 1: // East
			rampY := platformY + platformHeight/2
			if terrain.IsInBounds(platformX+platformWidth, rampY) {
				terrain.SetTile(platformX+platformWidth, rampY, TileRampUp)
			}
		case 2: // South
			rampX := platformX + platformWidth/2
			if terrain.IsInBounds(rampX, platformY+platformHeight) {
				terrain.SetTile(rampX, platformY+platformHeight, TileRampDown)
			}
		case 3: // West
			rampY := platformY + platformHeight/2
			if terrain.IsInBounds(platformX-1, rampY) {
				terrain.SetTile(platformX-1, rampY, TileRampUp)
			}
		}
	}
}

// addCornerPits adds pits in the corners of a room.
func (g *BSPGenerator) addCornerPits(terrain *Terrain, room *Room, rng *rand.Rand) {
	// Pit size: 2-3 tiles
	pitSize := 2 + rng.Intn(2)

	// Add pits to 1-2 random corners
	numPits := 1 + rng.Intn(2)
	corners := []int{0, 1, 2, 3} // TL, TR, BL, BR
	// Shuffle corners
	for i := len(corners) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		corners[i], corners[j] = corners[j], corners[i]
	}

	for i := 0; i < numPits && i < len(corners); i++ {
		corner := corners[i]
		var startX, startY int

		switch corner {
		case 0: // Top-left
			startX = room.X
			startY = room.Y
		case 1: // Top-right
			startX = room.X + room.Width - pitSize
			startY = room.Y
		case 2: // Bottom-left
			startX = room.X
			startY = room.Y + room.Height - pitSize
		case 3: // Bottom-right
			startX = room.X + room.Width - pitSize
			startY = room.Y + room.Height - pitSize
		}

		// Create pit
		for y := startY; y < startY+pitSize; y++ {
			for x := startX; x < startX+pitSize; x++ {
				if terrain.IsInBounds(x, y) {
					terrain.SetTile(x, y, TilePit)
				}
			}
		}
	}
}

// addLavaFlow adds a lava stream through a room.
func (g *BSPGenerator) addLavaFlow(terrain *Terrain, room *Room, rng *rand.Rand) {
	// Lava flows horizontally or vertically across room
	horizontal := rng.Float64() < 0.5

	if horizontal {
		// Horizontal flow
		lavaY := room.Y + rng.Intn(room.Height)
		for x := room.X; x < room.X+room.Width; x++ {
			if terrain.IsInBounds(x, lavaY) {
				// Skip if it's a wall or door
				currentTile := terrain.GetTile(x, lavaY)
				if currentTile == TileFloor || currentTile == TileCorridor {
					terrain.SetTile(x, lavaY, TileLavaFlow)
				}
			}
		}

		// Add bridges over lava (2-3 crossing points)
		numBridges := 2 + rng.Intn(2)
		for i := 0; i < numBridges; i++ {
			bridgeX := room.X + rng.Intn(room.Width)
			if terrain.IsInBounds(bridgeX, lavaY) && terrain.GetTile(bridgeX, lavaY) == TileLavaFlow {
				terrain.SetTile(bridgeX, lavaY, TileBridge)
			}
		}
	} else {
		// Vertical flow
		lavaX := room.X + rng.Intn(room.Width)
		for y := room.Y; y < room.Y+room.Height; y++ {
			if terrain.IsInBounds(lavaX, y) {
				currentTile := terrain.GetTile(lavaX, y)
				if currentTile == TileFloor || currentTile == TileCorridor {
					terrain.SetTile(lavaX, y, TileLavaFlow)
				}
			}
		}

		// Add bridges over lava
		numBridges := 2 + rng.Intn(2)
		for i := 0; i < numBridges; i++ {
			bridgeY := room.Y + rng.Intn(room.Height)
			if terrain.IsInBounds(lavaX, bridgeY) && terrain.GetTile(lavaX, bridgeY) == TileLavaFlow {
				terrain.SetTile(lavaX, bridgeY, TileBridge)
			}
		}
	}
}
