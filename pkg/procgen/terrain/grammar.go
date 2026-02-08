// Package terrain provides grammar-based dungeon generation.
// This file implements a graph grammar system for structured dungeon layouts
// with narrative flow and genre-appropriate architectural themes.
//
// GraphGrammarGenerator implements the procgen.Generator interface, providing both:
//   - Generate(seed, params) for standard interface usage with parameter-based configuration
//   - GenerateGraph(width, height) for internal direct usage with pre-configured settings
//
// The interface implementation supports difficulty and depth-based scaling of dungeon complexity,
// allowing procedural dungeons to adapt to player progression and challenge level.
package terrain

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/sirupsen/logrus"
)

// RoomNode represents a room in the dungeon graph.
type RoomNode struct {
	ID          int                    // Unique identifier
	Type        RoomType               // Functional type (uses existing RoomType enum)
	Position    Point                  // Spatial position (grid coordinates)
	Width       int                    // Room width in tiles
	Height      int                    // Room height in tiles
	Connections []*EdgeConnection      // Connections to other rooms
	Depth       int                    // Depth in narrative flow (0 = start)
	Theme       string                 // Architectural theme (from templates)
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
	Rooms            []*RoomNode       // All rooms in the dungeon
	StartRoom        *RoomNode         // The starting room
	BossRoom         *RoomNode         // The boss room (if any)
	RoomsByID        map[int]*RoomNode // Quick lookup by ID
	Seed             int64             // Generation seed
	Width            int               // Overall dungeon width
	Height           int               // Overall dungeon height
	NarrativeDepth   int               // Maximum narrative depth
	CriticalPath     []*RoomNode       // Main path from start to boss
	OptionalBranches [][]*RoomNode     // Side paths and branches
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
	lsystemString := g.lsystemGen.GenerateString()

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
			"roomCount":      len(graph.Rooms),
			"criticalPath":   len(graph.CriticalPath),
			"narrativeDepth": graph.NarrativeDepth,
		}).Info("dungeon graph generated successfully")
	}

	return graph, nil
}

// Generate implements the procgen.Generator interface.
// It generates a dungeon graph using graph grammars based on the seed and parameters.
// The genreID in params is used to select genre-specific L-system rules.
// Width and height are extracted from params.Custom["width"] and params.Custom["height"].
func (g *GraphGrammarGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	// Validate parameters
	if err := procgen.ValidateParams(params); err != nil {
		return nil, fmt.Errorf("invalid generation parameters: %w", err)
	}

	// Extract width and height from custom parameters
	width, height := 80, 50 // Default dimensions
	if params.Custom != nil {
		if w, ok := params.Custom["width"].(int); ok && w > 0 {
			width = w
		}
		if h, ok := params.Custom["height"].(int); ok && h > 0 {
			height = h
		}
	}

	// Validate dimensions
	if err := procgen.ValidateDimensions(width, height, 20, 20, 500, 500); err != nil {
		return nil, fmt.Errorf("invalid dungeon dimensions: %w", err)
	}

	// Get genre-specific L-system configuration
	var config LSystemConfig
	switch params.GenreID {
	case "fantasy":
		config = GetFantasyConfig(seed)
	case "scifi", "sci-fi":
		config = GetSciFiConfig(seed)
	case "horror":
		config = GetHorrorConfig(seed)
	case "cyberpunk":
		config = GetCyberpunkConfig(seed)
	case "post-apocalyptic", "postapocalyptic":
		config = GetPostApocalypticConfig(seed)
	default:
		// Default to fantasy for unknown genres
		config = GetFantasyConfig(seed)
	}

	// Adjust room counts based on difficulty and depth
	// Higher difficulty = more rooms (more challenging navigation)
	// Higher depth = more complex layouts
	roomMultiplier := 1.0 + (params.Difficulty * 0.5) + (float64(params.Depth) * 0.1)
	config.MinRoomCount = int(float64(config.MinRoomCount) * roomMultiplier)
	config.MaxRoomCount = int(float64(config.MaxRoomCount) * roomMultiplier)

	// Prevent excessive room counts
	if config.MinRoomCount > 50 {
		config.MinRoomCount = 50
	}
	if config.MaxRoomCount > 100 {
		config.MaxRoomCount = 100
	}

	// Ensure MinRoomCount doesn't exceed MaxRoomCount
	if config.MinRoomCount > config.MaxRoomCount {
		config.MinRoomCount = config.MaxRoomCount
	}

	// Create a temporary generator with the updated config
	tempGen := NewGraphGrammarGenerator(config)

	// Generate the dungeon graph
	graph, err := tempGen.GenerateGraph(width, height)
	if err != nil {
		return nil, fmt.Errorf("failed to generate dungeon graph: %w", err)
	}

	return graph, nil
}

// Validate implements the procgen.Generator interface.
// It checks if the generated dungeon graph is valid.
func (g *GraphGrammarGenerator) Validate(result interface{}) error {
	// Type check
	graph, ok := result.(*DungeonGraph)
	if !ok {
		return fmt.Errorf("result is not a *DungeonGraph, got type %T", result)
	}

	// Check for nil graph
	if graph == nil {
		return fmt.Errorf("dungeon graph is nil")
	}

	// Check for start room
	if graph.StartRoom == nil {
		return fmt.Errorf("dungeon graph has no start room")
	}

	// Check minimum room count (must have at least start + one other room)
	if len(graph.Rooms) < 2 {
		return fmt.Errorf("dungeon graph has too few rooms (%d), need at least 2", len(graph.Rooms))
	}

	// Check that all rooms have valid IDs
	for i, room := range graph.Rooms {
		if room == nil {
			return fmt.Errorf("room at index %d is nil", i)
		}
		if room.ID < 0 {
			return fmt.Errorf("room %d has invalid ID: %d", i, room.ID)
		}
	}

	// Check that all rooms are in the RoomsByID map
	if len(graph.RoomsByID) != len(graph.Rooms) {
		return fmt.Errorf("RoomsByID map size (%d) doesn't match Rooms slice size (%d)",
			len(graph.RoomsByID), len(graph.Rooms))
	}

	// Check for valid dimensions
	if graph.Width <= 0 || graph.Height <= 0 {
		return fmt.Errorf("dungeon graph has invalid dimensions: %dx%d", graph.Width, graph.Height)
	}

	// Check critical path exists and is valid
	if len(graph.CriticalPath) < 2 {
		return fmt.Errorf("critical path too short (%d rooms), need at least 2", len(graph.CriticalPath))
	}

	// Check that critical path starts with start room
	if graph.CriticalPath[0] != graph.StartRoom {
		return fmt.Errorf("critical path doesn't start with start room")
	}

	// Check narrative depth is valid
	if graph.NarrativeDepth < 0 {
		return fmt.Errorf("narrative depth is negative: %d", graph.NarrativeDepth)
	}

	// If boss room exists, narrative depth must be positive
	if graph.BossRoom != nil && graph.NarrativeDepth <= 0 {
		return fmt.Errorf("narrative depth must be positive when boss room exists (%d)", graph.NarrativeDepth)
	}

	// Check that all rooms have valid dimensions
	for _, room := range graph.Rooms {
		if room.Width <= 0 || room.Height <= 0 {
			return fmt.Errorf("room %d has invalid dimensions: %dx%d", room.ID, room.Width, room.Height)
		}
	}

	// Check that all connections are valid
	for _, room := range graph.Rooms {
		for _, conn := range room.Connections {
			if conn.From != room {
				return fmt.Errorf("room %d has connection with mismatched From pointer", room.ID)
			}
			if conn.To == nil {
				return fmt.Errorf("room %d has connection with nil To pointer", room.ID)
			}
			if conn.Weight <= 0 {
				return fmt.Errorf("room %d has connection with invalid weight: %f", room.ID, conn.Weight)
			}
		}
	}

	return nil
}

// parseIntoGraph converts L-system string into a room graph.
func (g *GraphGrammarGenerator) parseIntoGraph(lsystemString string, graph *DungeonGraph) error {
	var currentRoom *RoomNode
	roomID := 0
	depth := 0
	branchStack := make([]*RoomNode, 0)

	for _, char := range lsystemString {
		symbol := Symbol(char)

		switch symbol {
		case SymbolStart:
			currentRoom = g.createStartRoom(&roomID, depth, graph)
		case SymbolEnd:
			currentRoom, depth = g.createEndRoom(&roomID, depth, graph, currentRoom)
		case SymbolCombat:
			currentRoom = g.createCombatRoom(&roomID, depth, graph, currentRoom)
		case SymbolTreasure:
			currentRoom = g.createTreasureRoom(&roomID, depth, graph, currentRoom)
		case SymbolPuzzle:
			currentRoom = g.createPuzzleRoom(&roomID, depth, graph, currentRoom)
		case SymbolShop:
			currentRoom = g.createShopRoom(&roomID, depth, graph, currentRoom)
		case SymbolRest:
			currentRoom = g.createRestRoom(&roomID, depth, graph, currentRoom)
		case SymbolSecret:
			currentRoom = g.createSecretRoom(&roomID, depth, graph, currentRoom)
		case SymbolCorridor:
			depth++
		case SymbolBranch:
			currentRoom = g.createBranchRoom(&roomID, depth, graph, currentRoom, &branchStack)
		case SymbolEmpty:
			continue
		}
	}

	return nil
}

// createStartRoom creates the starting spawn room.
func (g *GraphGrammarGenerator) createStartRoom(roomID *int, depth int, graph *DungeonGraph) *RoomNode {
	room := &RoomNode{
		ID:         *roomID,
		Type:       RoomSpawn,
		Depth:      depth,
		Properties: make(map[string]interface{}),
	}
	*roomID++
	graph.AddRoom(room)
	graph.StartRoom = room
	return room
}

// createEndRoom creates the boss room and updates narrative depth.
func (g *GraphGrammarGenerator) createEndRoom(roomID *int, depth int, graph *DungeonGraph, currentRoom *RoomNode) (*RoomNode, int) {
	room := &RoomNode{
		ID:         *roomID,
		Type:       RoomBoss,
		Depth:      depth,
		Properties: make(map[string]interface{}),
	}
	*roomID++
	graph.AddRoom(room)
	graph.BossRoom = room
	if currentRoom != nil {
		graph.ConnectRooms(currentRoom, room, ConnectionCorridor, 1.0)
	}
	if depth > graph.NarrativeDepth {
		graph.NarrativeDepth = depth
	}
	return room, depth
}

// createCombatRoom creates a combat encounter room.
func (g *GraphGrammarGenerator) createCombatRoom(roomID *int, depth int, graph *DungeonGraph, currentRoom *RoomNode) *RoomNode {
	room := &RoomNode{
		ID:         *roomID,
		Type:       RoomCombat,
		Depth:      depth,
		Properties: make(map[string]interface{}),
	}
	*roomID++
	graph.AddRoom(room)
	if currentRoom != nil {
		graph.ConnectRooms(currentRoom, room, ConnectionCorridor, 1.0)
	}
	return room
}

// createTreasureRoom creates a treasure room.
func (g *GraphGrammarGenerator) createTreasureRoom(roomID *int, depth int, graph *DungeonGraph, currentRoom *RoomNode) *RoomNode {
	room := &RoomNode{
		ID:         *roomID,
		Type:       RoomTreasure,
		Depth:      depth,
		Properties: make(map[string]interface{}),
	}
	*roomID++
	graph.AddRoom(room)
	if currentRoom != nil {
		graph.ConnectRooms(currentRoom, room, ConnectionCorridor, 1.0)
	}
	return room
}

// createPuzzleRoom creates a puzzle room with locked connection.
func (g *GraphGrammarGenerator) createPuzzleRoom(roomID *int, depth int, graph *DungeonGraph, currentRoom *RoomNode) *RoomNode {
	room := &RoomNode{
		ID:         *roomID,
		Type:       RoomPuzzle,
		Depth:      depth,
		Properties: make(map[string]interface{}),
	}
	*roomID++
	graph.AddRoom(room)
	if currentRoom != nil {
		graph.ConnectRooms(currentRoom, room, ConnectionLocked, 1.5)
	}
	return room
}

// createShopRoom creates a merchant shop room.
func (g *GraphGrammarGenerator) createShopRoom(roomID *int, depth int, graph *DungeonGraph, currentRoom *RoomNode) *RoomNode {
	room := &RoomNode{
		ID:         *roomID,
		Type:       RoomShop,
		Depth:      depth,
		Properties: make(map[string]interface{}),
	}
	*roomID++
	graph.AddRoom(room)
	if currentRoom != nil {
		graph.ConnectRooms(currentRoom, room, ConnectionDoor, 1.0)
	}
	return room
}

// createRestRoom creates a rest/safe room.
func (g *GraphGrammarGenerator) createRestRoom(roomID *int, depth int, graph *DungeonGraph, currentRoom *RoomNode) *RoomNode {
	room := &RoomNode{
		ID:         *roomID,
		Type:       RoomRest,
		Depth:      depth,
		Properties: make(map[string]interface{}),
	}
	*roomID++
	graph.AddRoom(room)
	if currentRoom != nil {
		graph.ConnectRooms(currentRoom, room, ConnectionDoor, 1.0)
	}
	return room
}

// createSecretRoom creates a hidden secret room.
func (g *GraphGrammarGenerator) createSecretRoom(roomID *int, depth int, graph *DungeonGraph, currentRoom *RoomNode) *RoomNode {
	room := &RoomNode{
		ID:         *roomID,
		Type:       RoomSecret,
		Depth:      depth,
		Properties: make(map[string]interface{}),
	}
	*roomID++
	graph.AddRoom(room)
	if currentRoom != nil {
		graph.ConnectRooms(currentRoom, room, ConnectionSecret, 2.0)
	}
	return room
}

// createBranchRoom creates a branching hub room.
func (g *GraphGrammarGenerator) createBranchRoom(roomID *int, depth int, graph *DungeonGraph, currentRoom *RoomNode, branchStack *[]*RoomNode) *RoomNode {
	room := &RoomNode{
		ID:         *roomID,
		Type:       RoomBranch,
		Depth:      depth,
		Properties: make(map[string]interface{}),
	}
	*roomID++
	graph.AddRoom(room)
	if currentRoom != nil {
		graph.ConnectRooms(currentRoom, room, ConnectionCorridor, 1.0)
		*branchStack = append(*branchStack, currentRoom)
	}
	return room
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
		// Narrow corridors (3-5 wide, 3-5 tall)
		w := 3 + g.rng.Intn(3)
		h := 3 + g.rng.Intn(3)
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
// findBossOrDeepestRoom performs BFS to locate the boss room or deepest room in the dungeon graph.
// Returns the boss room, deepest room, and parent mapping for path reconstruction.
func (g *GraphGrammarGenerator) findBossOrDeepestRoom(startRoom *RoomNode) (*RoomNode, *RoomNode, map[int]*RoomNode) {
	visited := make(map[int]bool)
	parent := make(map[int]*RoomNode)
	queue := []*RoomNode{startRoom}
	visited[startRoom.ID] = true

	var bossRoom *RoomNode
	var deepestRoom *RoomNode
	maxDepth := 0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.Type == RoomBoss {
			bossRoom = current
			break
		}

		if current.Depth > maxDepth {
			maxDepth = current.Depth
			deepestRoom = current
		}

		for _, conn := range current.Connections {
			if !visited[conn.To.ID] {
				visited[conn.To.ID] = true
				parent[conn.To.ID] = current
				queue = append(queue, conn.To)
			}
		}
	}

	return bossRoom, deepestRoom, parent
}

// buildPathToStart constructs the path from the target room back to the start room.
// Returns the complete path from start to target.
func (g *GraphGrammarGenerator) buildPathToStart(targetRoom, startRoom *RoomNode, parent map[int]*RoomNode) []*RoomNode {
	if targetRoom == nil || targetRoom == startRoom {
		return []*RoomNode{startRoom}
	}

	path := []*RoomNode{targetRoom}
	current := targetRoom
	for current.ID != startRoom.ID {
		prev := parent[current.ID]
		if prev == nil {
			break
		}
		path = append([]*RoomNode{prev}, path...)
		current = prev
	}
	return path
}

// identifyOptionalBranches finds rooms that branch off from the critical path.
// Returns a list of branch paths.
func (g *GraphGrammarGenerator) identifyOptionalBranches(graph *DungeonGraph, criticalPath []*RoomNode) [][]*RoomNode {
	criticalSet := make(map[int]bool)
	for _, room := range criticalPath {
		criticalSet[room.ID] = true
	}

	var branches [][]*RoomNode
	for _, room := range graph.Rooms {
		if criticalSet[room.ID] {
			continue
		}

		if g.connectsToCriticalPath(room, criticalPath, criticalSet) {
			branch := []*RoomNode{room}
			branches = append(branches, branch)
		}
	}
	return branches
}

// connectsToCriticalPath checks if a room connects to any room on the critical path.
func (g *GraphGrammarGenerator) connectsToCriticalPath(room *RoomNode, criticalPath []*RoomNode, criticalSet map[int]bool) bool {
	for _, conn := range room.Connections {
		if criticalSet[conn.To.ID] {
			return true
		}
	}

	for _, critRoom := range criticalPath {
		for _, conn := range critRoom.Connections {
			if conn.To.ID == room.ID {
				return true
			}
		}
	}
	return false
}

func (g *GraphGrammarGenerator) identifyCriticalPath(graph *DungeonGraph) {
	if graph.StartRoom == nil {
		return
	}

	// Find boss room or deepest room using BFS
	bossRoom, deepestRoom, parent := g.findBossOrDeepestRoom(graph.StartRoom)

	// Select target room (boss if available, otherwise deepest)
	targetRoom := bossRoom
	if targetRoom == nil {
		targetRoom = deepestRoom
	}

	// Build critical path from start to target
	graph.CriticalPath = g.buildPathToStart(targetRoom, graph.StartRoom, parent)
	graph.BossRoom = bossRoom

	// Identify optional branches
	graph.OptionalBranches = g.identifyOptionalBranches(graph, graph.CriticalPath)
}

// validateGraph ensures the graph meets quality requirements.
func (g *GraphGrammarGenerator) validateGraph(graph *DungeonGraph) error {
	if graph.StartRoom == nil {
		return fmt.Errorf("graph has no start room")
	}

	if err := g.validateRoomReachability(graph); err != nil {
		return err
	}

	if err := g.validateCriticalPath(graph); err != nil {
		return err
	}

	return g.validateNarrativeDepth(graph)
}

// validateRoomReachability checks that all rooms are reachable from start.
func (g *GraphGrammarGenerator) validateRoomReachability(graph *DungeonGraph) error {
	reachable := g.countReachableRooms(graph)
	if reachable < len(graph.Rooms) {
		return fmt.Errorf("not all rooms are reachable from start (%d/%d)", reachable, len(graph.Rooms))
	}
	return nil
}

// validateCriticalPath checks that the critical path has sufficient length.
func (g *GraphGrammarGenerator) validateCriticalPath(graph *DungeonGraph) error {
	if len(graph.CriticalPath) < 2 {
		return fmt.Errorf("critical path too short (%d rooms, need at least 2)", len(graph.CriticalPath))
	}
	return nil
}

// validateNarrativeDepth ensures narrative depth is valid when boss room exists.
func (g *GraphGrammarGenerator) validateNarrativeDepth(graph *DungeonGraph) error {
	if graph.NarrativeDepth < 0 {
		return fmt.Errorf("narrative depth negative (%d)", graph.NarrativeDepth)
	}

	if hasBossRoom(graph) && graph.NarrativeDepth <= 0 {
		return fmt.Errorf("narrative depth must be positive when boss room exists (%d)", graph.NarrativeDepth)
	}

	return nil
}

// hasBossRoom checks if the graph contains a boss room.
func hasBossRoom(graph *DungeonGraph) bool {
	if graph.BossRoom != nil {
		return true
	}
	for _, room := range graph.CriticalPath {
		if room.Type == RoomBoss {
			return true
		}
	}
	return false
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

// GraphToTerrain converts a DungeonGraph into a Terrain map suitable for gameplay.
// It places rooms according to their graph positions and connects them with corridors.
func GraphToTerrain(graph *DungeonGraph) *Terrain {
	// Create terrain filled with walls
	terrain := NewTerrain(graph.Width, graph.Height, graph.Seed)

	// Convert each room node to a terrain room and carve it out
	for _, roomNode := range graph.Rooms {
		// Create a Room struct from RoomNode
		room := &Room{
			X:      roomNode.Position.X,
			Y:      roomNode.Position.Y,
			Width:  roomNode.Width,
			Height: roomNode.Height,
			Type:   roomNode.Type,
		}
		terrain.Rooms = append(terrain.Rooms, room)

		// Carve out the room floor
		for y := roomNode.Position.Y; y < roomNode.Position.Y+roomNode.Height; y++ {
			for x := roomNode.Position.X; x < roomNode.Position.X+roomNode.Width; x++ {
				if x >= 0 && x < graph.Width && y >= 0 && y < graph.Height {
					terrain.SetTile(x, y, TileFloor)
				}
			}
		}
	}

	// Create corridors between connected rooms
	for _, roomNode := range graph.Rooms {
		for _, conn := range roomNode.Connections {
			// Get room centers for corridor placement
			x1, y1 := roomNode.Position.X+roomNode.Width/2, roomNode.Position.Y+roomNode.Height/2
			x2, y2 := conn.To.Position.X+conn.To.Width/2, conn.To.Position.Y+conn.To.Height/2

			// Determine tile type based on connection type
			tileType := TileCorridor
			if conn.Type == ConnectionDoor {
				tileType = TileDoor
			} else if conn.Type == ConnectionSecret {
				tileType = TileSecretDoor
			}

			// Create L-shaped corridor: horizontal then vertical
			// Horizontal segment
			minX, maxX := x1, x2
			if minX > maxX {
				minX, maxX = maxX, minX
			}
			for x := minX; x <= maxX; x++ {
				if y1 >= 0 && y1 < graph.Height && x >= 0 && x < graph.Width {
					if terrain.GetTile(x, y1) == TileWall {
						terrain.SetTile(x, y1, tileType)
					}
				}
			}

			// Vertical segment
			minY, maxY := y1, y2
			if minY > maxY {
				minY, maxY = maxY, minY
			}
			for y := minY; y <= maxY; y++ {
				if y >= 0 && y < graph.Height && x2 >= 0 && x2 < graph.Width {
					if terrain.GetTile(x2, y) == TileWall {
						terrain.SetTile(x2, y, tileType)
					}
				}
			}
		}
	}

	// Add stairs if we have a start and boss room
	if graph.StartRoom != nil {
		startX, startY := graph.StartRoom.Position.X+graph.StartRoom.Width/2, graph.StartRoom.Position.Y+graph.StartRoom.Height/2
		terrain.AddStairs(startX, startY, true) // Stairs up at start
	}
	if graph.BossRoom != nil {
		bossX, bossY := graph.BossRoom.Position.X+graph.BossRoom.Width/2, graph.BossRoom.Position.Y+graph.BossRoom.Height/2
		terrain.AddStairs(bossX, bossY, false) // Stairs down at boss
	}

	return terrain
}
