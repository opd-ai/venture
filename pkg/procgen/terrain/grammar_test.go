package terrain

import (
	"testing"
)

func TestGrammarRoomType_String(t *testing.T) {
	tests := []struct {
		name     string
		roomType RoomType
		want     string
	}{
		{"spawn", RoomSpawn, "spawn"},
		{"combat", RoomCombat, "combat"},
		{"treasure", RoomTreasure, "treasure"},
		{"puzzle", RoomPuzzle, "puzzle"},
		{"shop", RoomShop, "shop"},
		{"rest", RoomRest, "rest"},
		{"secret", RoomSecret, "secret"},
		{"boss", RoomBoss, "boss"},
		{"corridor", RoomCorridor, "corridor"},
		{"branch", RoomBranch, "branch"},
		{"unknown", RoomType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.roomType.String(); got != tt.want {
				t.Errorf("RoomType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConnectionType_String(t *testing.T) {
	tests := []struct {
		name     string
		connType ConnectionType
		want     string
	}{
		{"corridor", ConnectionCorridor, "corridor"},
		{"door", ConnectionDoor, "door"},
		{"teleporter", ConnectionTeleporter, "teleporter"},
		{"secret", ConnectionSecret, "secret"},
		{"locked", ConnectionLocked, "locked"},
		{"unknown", ConnectionType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.connType.String(); got != tt.want {
				t.Errorf("ConnectionType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewDungeonGraph(t *testing.T) {
	seed := int64(12345)
	width, height := 80, 50

	graph := NewDungeonGraph(seed, width, height)

	if graph == nil {
		t.Fatal("NewDungeonGraph() returned nil")
	}

	if graph.Seed != seed {
		t.Errorf("Seed = %d, want %d", graph.Seed, seed)
	}

	if graph.Width != width {
		t.Errorf("Width = %d, want %d", graph.Width, width)
	}

	if graph.Height != height {
		t.Errorf("Height = %d, want %d", graph.Height, height)
	}

	if graph.Rooms == nil {
		t.Error("Rooms slice not initialized")
	}

	if graph.RoomsByID == nil {
		t.Error("RoomsByID map not initialized")
	}

	if graph.CriticalPath == nil {
		t.Error("CriticalPath slice not initialized")
	}

	if graph.OptionalBranches == nil {
		t.Error("OptionalBranches slice not initialized")
	}
}

func TestDungeonGraph_AddRoom(t *testing.T) {
	graph := NewDungeonGraph(12345, 80, 50)

	room := &RoomNode{
		ID:   0,
		Type: RoomSpawn,
	}

	graph.AddRoom(room)

	if len(graph.Rooms) != 1 {
		t.Errorf("Room count = %d, want 1", len(graph.Rooms))
	}

	if graph.GetRoom(0) != room {
		t.Error("Room not found in RoomsByID map")
	}
}

func TestDungeonGraph_ConnectRooms(t *testing.T) {
	graph := NewDungeonGraph(12345, 80, 50)

	room1 := &RoomNode{ID: 0, Type: RoomSpawn}
	room2 := &RoomNode{ID: 1, Type: RoomCombat}

	graph.AddRoom(room1)
	graph.AddRoom(room2)

	graph.ConnectRooms(room1, room2, ConnectionCorridor, 1.0)

	if len(room1.Connections) != 1 {
		t.Errorf("Connection count = %d, want 1", len(room1.Connections))
	}

	conn := room1.Connections[0]
	if conn.From != room1 {
		t.Error("Connection From is incorrect")
	}
	if conn.To != room2 {
		t.Error("Connection To is incorrect")
	}
	if conn.Type != ConnectionCorridor {
		t.Errorf("Connection Type = %v, want %v", conn.Type, ConnectionCorridor)
	}
	if conn.Weight != 1.0 {
		t.Errorf("Connection Weight = %f, want 1.0", conn.Weight)
	}
}

func TestNewGraphGrammarGenerator(t *testing.T) {
	config := GetFantasyConfig(12345)
	gen := NewGraphGrammarGenerator(config)

	if gen == nil {
		t.Fatal("NewGraphGrammarGenerator() returned nil")
	}

	if gen.lsystemGen == nil {
		t.Error("L-system generator not initialized")
	}

	if gen.rng == nil {
		t.Error("RNG not initialized")
	}
}

func TestGraphGrammarGenerator_GenerateGraph_Fantasy(t *testing.T) {
	config := GetFantasyConfig(12345)
	gen := NewGraphGrammarGenerator(config)

	graph, err := gen.GenerateGraph(80, 50)
	if err != nil {
		t.Fatalf("GenerateGraph() error = %v", err)
	}

	if graph == nil {
		t.Fatal("GenerateGraph() returned nil graph")
	}

	// Verify basic structure
	if graph.StartRoom == nil {
		t.Error("Graph has no start room")
	}

	if len(graph.Rooms) < 3 {
		t.Errorf("Too few rooms: got %d, want at least 3", len(graph.Rooms))
	}

	// Verify all rooms have positions
	for _, room := range graph.Rooms {
		if room.Position.X == 0 && room.Position.Y == 0 && room.ID != 0 {
			t.Errorf("Room %d has no position assigned", room.ID)
		}
	}

	// Verify all rooms have dimensions
	for _, room := range graph.Rooms {
		if room.Width <= 0 || room.Height <= 0 {
			t.Errorf("Room %d has invalid dimensions: %dx%d", room.ID, room.Width, room.Height)
		}
	}

	// Verify critical path exists
	if len(graph.CriticalPath) < 2 {
		t.Errorf("Critical path too short: got %d rooms, want at least 2", len(graph.CriticalPath))
	}
}

func TestGraphGrammarGenerator_GenerateGraph_SciFi(t *testing.T) {
	config := GetSciFiConfig(12345)
	gen := NewGraphGrammarGenerator(config)

	graph, err := gen.GenerateGraph(80, 50)
	if err != nil {
		t.Fatalf("GenerateGraph() error = %v", err)
	}

	if len(graph.Rooms) < 3 {
		t.Errorf("Too few rooms for sci-fi: got %d", len(graph.Rooms))
	}
}

func TestGraphGrammarGenerator_GenerateGraph_Horror(t *testing.T) {
	config := GetHorrorConfig(12345)
	gen := NewGraphGrammarGenerator(config)

	graph, err := gen.GenerateGraph(80, 50)
	if err != nil {
		t.Fatalf("GenerateGraph() error = %v", err)
	}

	if len(graph.Rooms) < 3 {
		t.Errorf("Too few rooms for horror: got %d", len(graph.Rooms))
	}
}

func TestGraphGrammarGenerator_GenerateGraph_Cyberpunk(t *testing.T) {
	config := GetCyberpunkConfig(12345)
	gen := NewGraphGrammarGenerator(config)

	graph, err := gen.GenerateGraph(80, 50)
	if err != nil {
		t.Fatalf("GenerateGraph() error = %v", err)
	}

	if len(graph.Rooms) < 3 {
		t.Errorf("Too few rooms for cyberpunk: got %d", len(graph.Rooms))
	}
}

func TestGraphGrammarGenerator_GenerateGraph_PostApocalyptic(t *testing.T) {
	config := GetPostApocalypticConfig(12345)
	gen := NewGraphGrammarGenerator(config)

	graph, err := gen.GenerateGraph(80, 50)
	if err != nil {
		t.Fatalf("GenerateGraph() error = %v", err)
	}

	if len(graph.Rooms) < 3 {
		t.Errorf("Too few rooms for post-apocalyptic: got %d", len(graph.Rooms))
	}
}

func TestGraphGrammarGenerator_Determinism(t *testing.T) {
	seed := int64(54321)
	config1 := GetFantasyConfig(seed)
	config2 := GetFantasyConfig(seed)

	gen1 := NewGraphGrammarGenerator(config1)
	gen2 := NewGraphGrammarGenerator(config2)

	graph1, err1 := gen1.GenerateGraph(80, 50)
	graph2, err2 := gen2.GenerateGraph(80, 50)

	if err1 != nil || err2 != nil {
		t.Fatalf("Errors: %v, %v", err1, err2)
	}

	// Verify same number of rooms
	if len(graph1.Rooms) != len(graph2.Rooms) {
		t.Errorf("Room count differs: %d vs %d", len(graph1.Rooms), len(graph2.Rooms))
	}

	// Verify same room types in order
	for i := 0; i < len(graph1.Rooms) && i < len(graph2.Rooms); i++ {
		if graph1.Rooms[i].Type != graph2.Rooms[i].Type {
			t.Errorf("Room %d type differs: %v vs %v", i, graph1.Rooms[i].Type, graph2.Rooms[i].Type)
		}
	}

	// Verify same positions (deterministic spatial layout)
	for i := 0; i < len(graph1.Rooms) && i < len(graph2.Rooms); i++ {
		if graph1.Rooms[i].Position != graph2.Rooms[i].Position {
			t.Errorf("Room %d position differs: %v vs %v", i, graph1.Rooms[i].Position, graph2.Rooms[i].Position)
		}
	}
}

func TestGraphGrammarGenerator_GetRoomDimensions(t *testing.T) {
	config := GetFantasyConfig(12345)
	gen := NewGraphGrammarGenerator(config)

	tests := []struct {
		name     string
		roomType RoomType
		minSize  int
		maxSize  int
	}{
		{"start", RoomSpawn, 15, 20},
		{"boss", RoomBoss, 15, 20},
		{"combat", RoomCombat, 10, 15},
		{"treasure", RoomTreasure, 8, 12},
		{"puzzle", RoomPuzzle, 10, 14},
		{"rest", RoomRest, 6, 10},
		{"secret", RoomSecret, 5, 8},
		{"corridor", RoomCorridor, 3, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, h := gen.getRoomDimensions(tt.roomType)

			if w < tt.minSize || w > tt.maxSize {
				t.Errorf("Width %d out of range [%d, %d]", w, tt.minSize, tt.maxSize)
			}

			if h < tt.minSize || h > tt.maxSize {
				t.Errorf("Height %d out of range [%d, %d]", h, tt.minSize, tt.maxSize)
			}
		})
	}
}

func TestGraphGrammarGenerator_ValidateGraph(t *testing.T) {
	config := GetFantasyConfig(12345)
	gen := NewGraphGrammarGenerator(config)

	tests := []struct {
		name    string
		setup   func() *DungeonGraph
		wantErr bool
	}{
		{
			name: "valid_graph",
			setup: func() *DungeonGraph {
				graph := NewDungeonGraph(12345, 80, 50)
				start := &RoomNode{ID: 0, Type: RoomSpawn}
				combat := &RoomNode{ID: 1, Type: RoomCombat, Depth: 1}
				boss := &RoomNode{ID: 2, Type: RoomBoss, Depth: 2}
				graph.AddRoom(start)
				graph.AddRoom(combat)
				graph.AddRoom(boss)
				graph.ConnectRooms(start, combat, ConnectionCorridor, 1.0)
				graph.ConnectRooms(combat, boss, ConnectionCorridor, 1.0)
				graph.StartRoom = start
				graph.BossRoom = boss
				graph.CriticalPath = []*RoomNode{start, combat, boss}
				graph.NarrativeDepth = 2
				return graph
			},
			wantErr: false,
		},
		{
			name: "no_start_room",
			setup: func() *DungeonGraph {
				graph := NewDungeonGraph(12345, 80, 50)
				return graph
			},
			wantErr: true,
		},
		{
			name: "short_critical_path",
			setup: func() *DungeonGraph {
				graph := NewDungeonGraph(12345, 80, 50)
				start := &RoomNode{ID: 0, Type: RoomSpawn}
				graph.AddRoom(start)
				graph.StartRoom = start
				graph.CriticalPath = []*RoomNode{start}
				return graph
			},
			wantErr: true,
		},
		{
			name: "zero_narrative_depth",
			setup: func() *DungeonGraph {
				graph := NewDungeonGraph(12345, 80, 50)
				start := &RoomNode{ID: 0, Type: RoomSpawn}
				boss := &RoomNode{ID: 1, Type: RoomBoss}
				graph.AddRoom(start)
				graph.AddRoom(boss)
				graph.ConnectRooms(start, boss, ConnectionCorridor, 1.0)
				graph.StartRoom = start
				graph.CriticalPath = []*RoomNode{start, boss}
				graph.NarrativeDepth = 0
				return graph
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph := tt.setup()
			err := gen.validateGraph(graph)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateGraph() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGraphGrammarGenerator_CountReachableRooms(t *testing.T) {
	config := GetFantasyConfig(12345)
	gen := NewGraphGrammarGenerator(config)

	// Create a simple graph
	graph := NewDungeonGraph(12345, 80, 50)
	room1 := &RoomNode{ID: 0, Type: RoomSpawn}
	room2 := &RoomNode{ID: 1, Type: RoomCombat}
	room3 := &RoomNode{ID: 2, Type: RoomTreasure}
	unreachable := &RoomNode{ID: 3, Type: RoomSecret}

	graph.AddRoom(room1)
	graph.AddRoom(room2)
	graph.AddRoom(room3)
	graph.AddRoom(unreachable)

	graph.ConnectRooms(room1, room2, ConnectionCorridor, 1.0)
	graph.ConnectRooms(room2, room3, ConnectionCorridor, 1.0)
	// unreachable has no connections

	graph.StartRoom = room1

	count := gen.countReachableRooms(graph)

	if count != 3 {
		t.Errorf("countReachableRooms() = %d, want 3", count)
	}
}

func TestGraphGrammarGenerator_IdentifyCriticalPath(t *testing.T) {
	config := GetFantasyConfig(12345)
	gen := NewGraphGrammarGenerator(config)

	// Create a graph with a clear path to boss
	graph := NewDungeonGraph(12345, 80, 50)
	start := &RoomNode{ID: 0, Type: RoomSpawn}
	combat1 := &RoomNode{ID: 1, Type: RoomCombat}
	combat2 := &RoomNode{ID: 2, Type: RoomCombat}
	boss := &RoomNode{ID: 3, Type: RoomBoss}
	sideRoom := &RoomNode{ID: 4, Type: RoomSecret}

	graph.AddRoom(start)
	graph.AddRoom(combat1)
	graph.AddRoom(combat2)
	graph.AddRoom(boss)
	graph.AddRoom(sideRoom)

	graph.ConnectRooms(start, combat1, ConnectionCorridor, 1.0)
	graph.ConnectRooms(combat1, combat2, ConnectionCorridor, 1.0)
	graph.ConnectRooms(combat2, boss, ConnectionCorridor, 1.0)
	graph.ConnectRooms(combat1, sideRoom, ConnectionSecret, 2.0)

	graph.StartRoom = start
	graph.BossRoom = boss

	gen.identifyCriticalPath(graph)

	// Verify critical path includes start and boss
	if len(graph.CriticalPath) < 2 {
		t.Fatalf("Critical path too short: %d rooms", len(graph.CriticalPath))
	}

	if graph.CriticalPath[0] != start {
		t.Error("Critical path doesn't start with start room")
	}

	lastIdx := len(graph.CriticalPath) - 1
	if graph.CriticalPath[lastIdx] != boss {
		t.Error("Critical path doesn't end with boss room")
	}

	// Verify optional branches identified
	if len(graph.OptionalBranches) == 0 {
		t.Error("Optional branches not identified")
	}
}

func TestParseIntoGraph_SimpleSequence(t *testing.T) {
	config := GetFantasyConfig(12345)
	gen := NewGraphGrammarGenerator(config)

	graph := NewDungeonGraph(12345, 80, 50)
	lsystemString := "S-C-E"

	err := gen.parseIntoGraph(lsystemString, graph)
	if err != nil {
		t.Fatalf("parseIntoGraph() error = %v", err)
	}

	if len(graph.Rooms) != 3 {
		t.Errorf("Room count = %d, want 3", len(graph.Rooms))
	}

	if graph.StartRoom == nil {
		t.Error("Start room not set")
	}

	if graph.BossRoom == nil {
		t.Error("Boss room not set")
	}
}

func TestParseIntoGraph_WithBranch(t *testing.T) {
	config := GetFantasyConfig(12345)
	gen := NewGraphGrammarGenerator(config)

	graph := NewDungeonGraph(12345, 80, 50)
	lsystemString := "S-+-C-T"

	err := gen.parseIntoGraph(lsystemString, graph)
	if err != nil {
		t.Fatalf("parseIntoGraph() error = %v", err)
	}

	// Should have start, branch hub, combat, treasure
	if len(graph.Rooms) < 3 {
		t.Errorf("Room count = %d, want at least 3", len(graph.Rooms))
	}

	// Verify branch room exists
	hasBranch := false
	for _, room := range graph.Rooms {
		if room.Type == RoomBranch {
			hasBranch = true
			break
		}
	}

	if !hasBranch {
		t.Error("Branch room not created")
	}
}

func BenchmarkGraphGrammarGenerator_GenerateGraph(b *testing.B) {
	config := GetFantasyConfig(12345)
	gen := NewGraphGrammarGenerator(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.GenerateGraph(80, 50)
		if err != nil {
			b.Fatalf("GenerateGraph() error = %v", err)
		}
	}
}

func BenchmarkGraphGrammarGenerator_GenerateGraph_Large(b *testing.B) {
	config := GetSciFiConfig(12345)
	config.MaxRoomCount = 50
	gen := NewGraphGrammarGenerator(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.GenerateGraph(160, 100)
		if err != nil {
			b.Fatalf("GenerateGraph() error = %v", err)
		}
	}
}
