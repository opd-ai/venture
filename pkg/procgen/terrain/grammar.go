// Package terrain provides grammar-based dungeon generation.
// This file implements a graph grammar system for structured dungeon layouts
// with narrative flow and genre-appropriate architectural themes.
package terrain

import (
	"fmt"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// RoomNode represents a room in the dungeon graph.
type RoomNode struct {
	ID          int               // Unique identifier
	Type        RoomType          // Functional type (uses existing RoomType enum)
	Position    Point             // Spatial position (grid coordinates)
	Width       int               // Room width in tiles
	Height      int               // Room height in tiles
	Connections []*EdgeConnection // Connections to other rooms
	Depth       int               // Depth in narrative flow (0 = start)
	Theme       string            // Architectural theme (from templates)
	Properties  map[string]interface{} // Custom properties
}

// EdgeConnection represents a connection between two rooms.
type EdgeConnection struct {
	From   *RoomNode      // Source room
	To     *RoomNode      // Target room
	Type   ConnectionType // Type of connection
	Weight float64        // Path weight (affects routing)
}

// ConnectionType defines how rooms are connected.
type ConnectionType int

const (
	// ConnectionCorridor is a normal hallway
	ConnectionCorridor ConnectionType = iota
	// ConnectionDoor is a direct door connection
	ConnectionDoor
	// ConnectionTeleporter is a teleportation link
	ConnectionTeleporter
	// ConnectionSecret is a hidden passage
	ConnectionSecret
	// ConnectionLocked is a locked door/gate
	ConnectionLocked
)

// String returns the string representation of a connection type.
func (c ConnectionType) String() string {
	switch c {
	case ConnectionCorridor:
		return "corridor"
	case ConnectionDoor:
		return "door"
	case ConnectionTeleporter:
		return "teleporter"
	case ConnectionSecret:
		return "secret"
	case ConnectionLocked:
		return "locked"
	default:
		return "unknown"
	}
}

// DungeonGraph represents the complete room graph structure.
type DungeonGraph struct {
	Rooms           []*RoomNode       // All rooms in the dungeon
	StartRoom       *RoomNode         // The starting room
	BossRoom        *RoomNode         // The boss room (if any)
	RoomsByID       map[int]*RoomNode // Quick lookup by ID
	Seed            int64             // Generation seed
	Width           int               // Overall dungeon width
	Height          int               // Overall dungeon height
	NarrativeDepth  int               // Maximum narrative depth
	CriticalPath    []*RoomNode       // Main path from start to boss
	OptionalBranches [][]*RoomNode    // Side paths and branches
}

// NewDungeonGraph creates a new empty dungeon graph.
func NewDungeonGraph(seed int64, width, height int) *DungeonGraph {
	return &DungeonGraph{
		Rooms:            make([]*RoomNode, 0),
		RoomsByID:        make(map[int]*RoomNode),
		Seed:             seed,
		Width:            width,
		Height:           height,
		CriticalPath:     make([]*RoomNode, 0),
		OptionalBranches: make([][]*RoomNode, 0),
	}
}

// AddRoom adds a room node to the graph.
func (g *DungeonGraph) AddRoom(room *RoomNode) {
	g.Rooms = append(g.Rooms, room)
	g.RoomsByID[room.ID] = room
}

// ConnectRooms creates a connection between two rooms.
func (g *DungeonGraph) ConnectRooms(from, to *RoomNode, connType ConnectionType, weight float64) {
	conn := &EdgeConnection{
		From:   from,
		To:     to,
		Type:   connType,
		Weight: weight,
	}
	from.Connections = append(from.Connections, conn)
}

// GetRoom retrieves a room by ID.
func (g *DungeonGraph) GetRoom(id int) *RoomNode {
	return g.RoomsByID[id]
}

// GraphGrammarGenerator creates dungeon layouts using graph grammars and L-systems.
type GraphGrammarGenerator struct {
	lsystemGen *LSystemGenerator
	rng        *rand.Rand
	logger     *logrus.Entry
}

// NewGraphGrammarGenerator creates a new graph grammar generator.
func NewGraphGrammarGenerator(config LSystemConfig) *GraphGrammarGenerator {
	return NewGraphGrammarGeneratorWithLogger(config, nil)
}

// NewGraphGrammarGeneratorWithLogger creates a graph grammar generator with a logger.
func NewGraphGrammarGeneratorWithLogger(config LSystemConfig, logger *logrus.Logger) *GraphGrammarGenerator {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("generator", "graph_grammar")
	}
	return &GraphGrammarGenerator{
		lsystemGen: NewLSystemGenerator(config),
		rng:        rand.New(rand.NewSource(config.Seed)),
		logger:     logEntry,
	}
}

// GenerateGraph creates a dungeon graph from an L-system string.
func (g *GraphGrammarGenerator) GenerateGraph(width, height int) (*DungeonGraph, error) {
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.Debug("generating dungeon graph")
	}

	// Generate L-system string
	lsystemString := g.lsystemGen.Generate()

	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithField("lsystem", lsystemString).Debug("L-system string generated")
	}

	// Create dungeon graph
	graph := NewDungeonGraph(g.lsystemGen.config.Seed, width, height)

	// Parse L-system string into room graph
	err := g.parseIntoGraph(lsystemString, graph)
	if err != nil {
		return nil, fmt.Errorf("failed to parse L-system into graph: %w", err)
	}

	// Assign spatial positions to rooms
	g.assignRoomPositions(graph)

	// Identify critical path (start → boss)
	g.identifyCriticalPath(graph)

	// Validate graph structure
	if err := g.validateGraph(graph); err != nil {
		return nil, fmt.Errorf("graph validation failed: %w", err)
	}

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"roomCount":       len(graph.Rooms),
			"criticalPath":    len(graph.CriticalPath),
			"narrativeDepth":  graph.NarrativeDepth,
		}).Info("dungeon graph generated successfully")
	}

	return graph, nil
}

// parseIntoGraph converts L-system string into a room graph.
func (g *GraphGrammarGenerator) parseIntoGraph(lsystemString string, graph *DungeonGraph) error {
	var currentRoom *RoomNode
	roomID := 0
	depth := 0

	// Stack for tracking branch points
	branchStack := make([]*RoomNode, 0)

	for _, char := range lsystemString {
		symbol := Symbol(char)

		switch symbol {
		case SymbolStart:
			room := &RoomNode{
				ID:         roomID,
				Type:       RoomSpawn, // Use RoomSpawn for start
				Depth:      depth,
				Properties: make(map[string]interface{}),
			}
			roomID++
			graph.AddRoom(room)
			graph.StartRoom = room
			currentRoom = room

		case SymbolEnd:
			room := &RoomNode{
				ID:         roomID,
				Type:       RoomBoss, // Use RoomBoss for end
				Depth:      depth,
				Properties: make(map[string]interface{}),
			}
			roomID++
			graph.AddRoom(room)
			graph.BossRoom = room
			if currentRoom != nil {
				graph.ConnectRooms(currentRoom, room, ConnectionCorridor, 1.0)
			}
			currentRoom = room
			if depth > graph.NarrativeDepth {
				graph.NarrativeDepth = depth
			}

		case SymbolCombat:
			room := &RoomNode{
				ID:         roomID,
				Type:       RoomCombat, // Use new RoomCombat type
				Depth:      depth,
				Properties: make(map[string]interface{}),
			}
			roomID++
			graph.AddRoom(room)
			if currentRoom != nil {
				graph.ConnectRooms(currentRoom, room, ConnectionCorridor, 1.0)
			}
			currentRoom = room

		case SymbolTreasure:
			room := &RoomNode{
				ID:         roomID,
				Type:       RoomTreasure, // Existing type
				Depth:      depth,
				Properties: make(map[string]interface{}),
			}
			roomID++
			graph.AddRoom(room)
			if currentRoom != nil {
				graph.ConnectRooms(currentRoom, room, ConnectionCorridor, 1.0)
			}
			currentRoom = room

		case SymbolPuzzle:
			room := &RoomNode{
				ID:         roomID,
				Type:       RoomPuzzle, // Use new RoomPuzzle type
				Depth:      depth,
				Properties: make(map[string]interface{}),
			}
			roomID++
			graph.AddRoom(room)
			if currentRoom != nil {
				graph.ConnectRooms(currentRoom, room, ConnectionLocked, 1.5)
			}
			currentRoom = room

		case SymbolShop:
			room := &RoomNode{
				ID:         roomID,
				Type:       RoomShop, // Use new RoomShop type
				Depth:      depth,
				Properties: make(map[string]interface{}),
			}
			roomID++
			graph.AddRoom(room)
			if currentRoom != nil {
				graph.ConnectRooms(currentRoom, room, ConnectionDoor, 1.0)
			}
			currentRoom = room

		case SymbolRest:
			room := &RoomNode{
				ID:         roomID,
				Type:       RoomRest, // Use new RoomRest type
				Depth:      depth,
				Properties: make(map[string]interface{}),
			}
			roomID++
			graph.AddRoom(room)
			if currentRoom != nil {
				graph.ConnectRooms(currentRoom, room, ConnectionDoor, 1.0)
			}
			currentRoom = room

		case SymbolSecret:
			room := &RoomNode{
				ID:         roomID,
				Type:       RoomSecret, // Use new RoomSecret type
				Depth:      depth,
				Properties: make(map[string]interface{}),
			}
			roomID++
			graph.AddRoom(room)
			if currentRoom != nil {
				graph.ConnectRooms(currentRoom, room, ConnectionSecret, 2.0)
			}
			currentRoom = room

		case SymbolCorridor:
			// Corridor advances depth
			depth++

		case SymbolBranch:
			// Branch point - create a branch hub room
			room := &RoomNode{
				ID:         roomID,
				Type:       RoomBranch, // Use new RoomBranch type
				Depth:      depth,
				Properties: make(map[string]interface{}),
			}
			roomID++
			graph.AddRoom(room)
			if currentRoom != nil {
				graph.ConnectRooms(currentRoom, room, ConnectionCorridor, 1.0)
			}
			// Push previous room onto stack for later branching
			if currentRoom != nil {
				branchStack = append(branchStack, currentRoom)
			}
			currentRoom = room

		case SymbolEmpty:
			// Skip empty symbols
			continue
		}
	}

	return nil
}

// assignRoomPositions assigns spatial positions to rooms in the graph.
func (g *GraphGrammarGenerator) assignRoomPositions(graph *DungeonGraph) {
	if len(graph.Rooms) == 0 {
		return
	}

	// Calculate grid spacing
	maxDepth := graph.NarrativeDepth
	if maxDepth == 0 {
		maxDepth = 1
	}
	
	horizontalSpacing := graph.Width / (maxDepth + 2)
	verticalSpacing := graph.Height / (len(graph.Rooms) + 2)

	// Track vertical position for each depth level
	depthCounters := make(map[int]int)

	// Assign positions based on depth and order
	for _, room := range graph.Rooms {
		depth := room.Depth
		verticalIndex := depthCounters[depth]
		depthCounters[depth]++

		// Calculate position
		x := horizontalSpacing * (depth + 1)
		y := verticalSpacing * (verticalIndex + 1)

		// Add some randomness for visual variety (±20%)
		jitterX := int(float64(horizontalSpacing) * 0.2 * (g.rng.Float64()*2 - 1))
		jitterY := int(float64(verticalSpacing) * 0.2 * (g.rng.Float64()*2 - 1))

		room.Position = Point{X: x + jitterX, Y: y + jitterY}

		// Assign room dimensions based on type
		room.Width, room.Height = g.getRoomDimensions(room.Type)
	}
}

// getRoomDimensions returns appropriate dimensions for a room type.
func (g *GraphGrammarGenerator) getRoomDimensions(roomType RoomType) (int, int) {
	switch roomType {
	case RoomSpawn, RoomBoss:
		// Large rooms (15-20 tiles)
		w := 15 + g.rng.Intn(6)
		h := 15 + g.rng.Intn(6)
		return w, h
	case RoomCombat:
		// Medium rooms (10-15 tiles)
		w := 10 + g.rng.Intn(6)
		h := 10 + g.rng.Intn(6)
		return w, h
	case RoomTreasure, RoomShop:
		// Medium rooms (8-12 tiles)
		w := 8 + g.rng.Intn(5)
		h := 8 + g.rng.Intn(5)
		return w, h
	case RoomPuzzle:
		// Square rooms (10-14 tiles)
		size := 10 + g.rng.Intn(5)
		return size, size
	case RoomRest:
		// Small cozy rooms (6-10 tiles)
		w := 6 + g.rng.Intn(5)
		h := 6 + g.rng.Intn(5)
		return w, h
	case RoomSecret:
		// Small hidden rooms (5-8 tiles)
		w := 5 + g.rng.Intn(4)
		h := 5 + g.rng.Intn(4)
		return w, h
	case RoomCorridor:
		// Narrow corridors (3-5 wide, random length)
		w := 3 + g.rng.Intn(3)
		h := 5 + g.rng.Intn(8)
		return w, h
	case RoomBranch:
		// Medium hub rooms (8-12 tiles)
		w := 8 + g.rng.Intn(5)
		h := 8 + g.rng.Intn(5)
		return w, h
	default:
		return 8, 8
	}
}

// identifyCriticalPath identifies the main path from start to boss.
func (g *GraphGrammarGenerator) identifyCriticalPath(graph *DungeonGraph) {
	if graph.StartRoom == nil {
		return
	}

	// Build critical path using breadth-first search to boss room
	visited := make(map[int]bool)
	parent := make(map[int]*RoomNode)
	queue := []*RoomNode{graph.StartRoom}
	visited[graph.StartRoom.ID] = true

	var bossRoom *RoomNode

	// BFS to find boss room
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.Type == RoomBoss {
			bossRoom = current
			break
		}

		for _, conn := range current.Connections {
			if !visited[conn.To.ID] {
				visited[conn.To.ID] = true
				parent[conn.To.ID] = current
				queue = append(queue, conn.To)
			}
		}
	}

	// Build path from boss back to start
	if bossRoom != nil {
		path := []*RoomNode{bossRoom}
		current := bossRoom
		for current.ID != graph.StartRoom.ID {
			prev := parent[current.ID]
			if prev == nil {
				break
			}
			path = append([]*RoomNode{prev}, path...)
			current = prev
		}
		graph.CriticalPath = path
	}

	// Identify optional branches (rooms not on critical path)
	criticalSet := make(map[int]bool)
	for _, room := range graph.CriticalPath {
		criticalSet[room.ID] = true
	}

	for _, room := range graph.Rooms {
		if !criticalSet[room.ID] && len(room.Connections) > 0 {
			// This is a branch - trace its path
			branch := []*RoomNode{room}
			graph.OptionalBranches = append(graph.OptionalBranches, branch)
		}
	}
}

// validateGraph ensures the graph meets quality requirements.
func (g *GraphGrammarGenerator) validateGraph(graph *DungeonGraph) error {
	// Check for start room
	if graph.StartRoom == nil {
		return fmt.Errorf("graph has no start room")
	}

	// Check that all rooms are reachable from start
	reachable := g.countReachableRooms(graph)
	if reachable < len(graph.Rooms) {
		return fmt.Errorf("not all rooms are reachable from start (%d/%d)", reachable, len(graph.Rooms))
	}

	// Check for critical path (at least start room)
	// Note: Boss room is optional - not all L-system configurations generate one
	if len(graph.CriticalPath) < 1 {
		// If no boss room, critical path should at least contain start
		if graph.StartRoom != nil {
			graph.CriticalPath = []*RoomNode{graph.StartRoom}
		} else {
			return fmt.Errorf("critical path empty and no start room")
		}
	}

	// Check narrative depth progression
	if graph.NarrativeDepth < 0 {
		return fmt.Errorf("narrative depth negative (%d)", graph.NarrativeDepth)
	}

	return nil
}

// countReachableRooms counts how many rooms are reachable from the start.
func (g *GraphGrammarGenerator) countReachableRooms(graph *DungeonGraph) int {
	if graph.StartRoom == nil {
		return 0
	}

	visited := make(map[int]bool)
	queue := []*RoomNode{graph.StartRoom}
	visited[graph.StartRoom.ID] = true
	count := 1

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, conn := range current.Connections {
			if !visited[conn.To.ID] {
				visited[conn.To.ID] = true
				count++
				queue = append(queue, conn.To)
			}
		}
	}

	return count
}
