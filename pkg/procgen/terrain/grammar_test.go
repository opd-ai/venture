package terrain

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
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

// TestGraphGrammarGenerator_Generate tests the Generator interface implementation
func TestGraphGrammarGenerator_Generate(t *testing.T) {
	config := GetFantasyConfig(12345)
	gen := NewGraphGrammarGenerator(config)

	tests := []struct {
		name       string
		seed       int64
		params     procgen.GenerationParams
		wantErr    bool
		checkGraph func(*testing.T, interface{})
	}{
		{
			name: "valid_fantasy_generation",
			seed: 12345,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"width":  80,
					"height": 50,
				},
			},
			wantErr: false,
			checkGraph: func(t *testing.T, result interface{}) {
				graph := result.(*DungeonGraph)
				if graph == nil {
					t.Error("Generated graph is nil")
				}
				if graph.StartRoom == nil {
					t.Error("Graph has no start room")
				}
				if len(graph.Rooms) < 2 {
					t.Errorf("Too few rooms: got %d", len(graph.Rooms))
				}
			},
		},
		{
			name: "scifi_genre",
			seed: 54321,
			params: procgen.GenerationParams{
				Difficulty: 0.7,
				Depth:      10,
				GenreID:    "scifi",
				Custom: map[string]interface{}{
					"width":  100,
					"height": 60,
				},
			},
			wantErr: false,
			checkGraph: func(t *testing.T, result interface{}) {
				graph := result.(*DungeonGraph)
				if len(graph.Rooms) < 2 {
					t.Errorf("Scifi graph too small: got %d rooms", len(graph.Rooms))
				}
			},
		},
		{
			name: "horror_genre",
			seed: 99999,
			params: procgen.GenerationParams{
				Difficulty: 0.8,
				Depth:      15,
				GenreID:    "horror",
				Custom: map[string]interface{}{
					"width":  120,
					"height": 80,
				},
			},
			wantErr: false,
		},
		{
			name: "cyberpunk_genre",
			seed: 11111,
			params: procgen.GenerationParams{
				Difficulty: 0.3,
				Depth:      3,
				GenreID:    "cyberpunk",
			},
			wantErr: false,
		},
		{
			name: "post_apocalyptic_genre",
			seed: 22222,
			params: procgen.GenerationParams{
				Difficulty: 0.6,
				Depth:      8,
				GenreID:    "post-apocalyptic",
			},
			wantErr: false,
		},
		{
			name: "default_dimensions",
			seed: 33333,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr: false,
			checkGraph: func(t *testing.T, result interface{}) {
				graph := result.(*DungeonGraph)
				// Should use default dimensions (80x50)
				if graph.Width != 80 || graph.Height != 50 {
					t.Errorf("Dimensions = %dx%d, want 80x50", graph.Width, graph.Height)
				}
			},
		},
		{
			name: "high_difficulty",
			seed: 44444,
			params: procgen.GenerationParams{
				Difficulty: 1.0,
				Depth:      20,
				GenreID:    "fantasy",
			},
			wantErr: false,
			checkGraph: func(t *testing.T, result interface{}) {
				graph := result.(*DungeonGraph)
				// High difficulty should produce more rooms
				if len(graph.Rooms) < 5 {
					t.Errorf("High difficulty produced too few rooms: got %d", len(graph.Rooms))
				}
			},
		},
		{
			name: "invalid_difficulty_too_high",
			seed: 12345,
			params: procgen.GenerationParams{
				Difficulty: 1.5,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
		{
			name: "invalid_difficulty_negative",
			seed: 12345,
			params: procgen.GenerationParams{
				Difficulty: -0.5,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
		{
			name: "invalid_depth_negative",
			seed: 12345,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      -1,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
		{
			name: "empty_genre",
			seed: 12345,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    "",
			},
			wantErr: true,
		},
		{
			name: "invalid_dimensions_too_small",
			seed: 12345,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"width":  10,
					"height": 10,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid_dimensions_too_large",
			seed: 12345,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"width":  1000,
					"height": 1000,
				},
			},
			wantErr: true,
		},
		{
			name: "unknown_genre_uses_default",
			seed: 55555,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    "unknown-genre",
			},
			wantErr: false,
			checkGraph: func(t *testing.T, result interface{}) {
				graph := result.(*DungeonGraph)
				if graph == nil {
					t.Error("Unknown genre should fallback to fantasy")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.Generate(tt.seed, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("Generate() returned nil result for valid input")
					return
				}

				// Verify result is a DungeonGraph
				graph, ok := result.(*DungeonGraph)
				if !ok {
					t.Errorf("Generate() returned wrong type: got %T, want *DungeonGraph", result)
					return
				}

				// Run additional checks if provided
				if tt.checkGraph != nil {
					tt.checkGraph(t, graph)
				}
			}
		})
	}
}

// TestGraphGrammarGenerator_Validate tests the Validate method
func TestGraphGrammarGenerator_Validate(t *testing.T) {
	config := GetFantasyConfig(12345)
	gen := NewGraphGrammarGenerator(config)

	tests := []struct {
		name    string
		result  interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid_graph",
			result: func() *DungeonGraph {
				graph := NewDungeonGraph(12345, 80, 50)
				start := &RoomNode{ID: 0, Type: RoomSpawn, Width: 15, Height: 15}
				combat := &RoomNode{ID: 1, Type: RoomCombat, Width: 10, Height: 10, Depth: 1}
				boss := &RoomNode{ID: 2, Type: RoomBoss, Width: 20, Height: 20, Depth: 2}
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
			}(),
			wantErr: false,
		},
		{
			name:    "wrong_type_string",
			result:  "not a graph",
			wantErr: true,
			errMsg:  "not a *DungeonGraph",
		},
		{
			name:    "wrong_type_int",
			result:  42,
			wantErr: true,
			errMsg:  "not a *DungeonGraph",
		},
		{
			name:    "nil_graph",
			result:  (*DungeonGraph)(nil),
			wantErr: true,
			errMsg:  "nil",
		},
		{
			name: "no_start_room",
			result: func() *DungeonGraph {
				graph := NewDungeonGraph(12345, 80, 50)
				room := &RoomNode{ID: 0, Type: RoomCombat, Width: 10, Height: 10}
				graph.AddRoom(room)
				graph.CriticalPath = []*RoomNode{room, room}
				return graph
			}(),
			wantErr: true,
			errMsg:  "no start room",
		},
		{
			name: "too_few_rooms",
			result: func() *DungeonGraph {
				graph := NewDungeonGraph(12345, 80, 50)
				start := &RoomNode{ID: 0, Type: RoomSpawn, Width: 15, Height: 15}
				graph.AddRoom(start)
				graph.StartRoom = start
				graph.CriticalPath = []*RoomNode{start, start}
				return graph
			}(),
			wantErr: true,
			errMsg:  "too few rooms",
		},
		{
			name: "nil_room_in_slice",
			result: func() *DungeonGraph {
				graph := NewDungeonGraph(12345, 80, 50)
				start := &RoomNode{ID: 0, Type: RoomSpawn, Width: 15, Height: 15}
				graph.Rooms = append(graph.Rooms, start)
				graph.Rooms = append(graph.Rooms, nil)
				graph.RoomsByID[0] = start
				graph.StartRoom = start
				graph.CriticalPath = []*RoomNode{start, start}
				return graph
			}(),
			wantErr: true,
			errMsg:  "is nil",
		},
		{
			name: "invalid_room_id",
			result: func() *DungeonGraph {
				graph := NewDungeonGraph(12345, 80, 50)
				start := &RoomNode{ID: -1, Type: RoomSpawn, Width: 15, Height: 15}
				combat := &RoomNode{ID: 1, Type: RoomCombat, Width: 10, Height: 10}
				graph.AddRoom(start)
				graph.AddRoom(combat)
				graph.StartRoom = start
				graph.CriticalPath = []*RoomNode{start, combat}
				return graph
			}(),
			wantErr: true,
			errMsg:  "invalid ID",
		},
		{
			name: "rooms_map_mismatch",
			result: func() *DungeonGraph {
				graph := NewDungeonGraph(12345, 80, 50)
				start := &RoomNode{ID: 0, Type: RoomSpawn, Width: 15, Height: 15}
				combat := &RoomNode{ID: 1, Type: RoomCombat, Width: 10, Height: 10}
				graph.Rooms = append(graph.Rooms, start)
				graph.Rooms = append(graph.Rooms, combat)
				graph.RoomsByID[0] = start
				// Missing room 1 in map
				graph.StartRoom = start
				graph.CriticalPath = []*RoomNode{start, combat}
				return graph
			}(),
			wantErr: true,
			errMsg:  "doesn't match",
		},
		{
			name: "invalid_dimensions_zero",
			result: func() *DungeonGraph {
				graph := NewDungeonGraph(12345, 0, 50)
				start := &RoomNode{ID: 0, Type: RoomSpawn, Width: 15, Height: 15}
				combat := &RoomNode{ID: 1, Type: RoomCombat, Width: 10, Height: 10}
				graph.AddRoom(start)
				graph.AddRoom(combat)
				graph.StartRoom = start
				graph.CriticalPath = []*RoomNode{start, combat}
				return graph
			}(),
			wantErr: true,
			errMsg:  "invalid dimensions",
		},
		{
			name: "invalid_dimensions_negative",
			result: func() *DungeonGraph {
				graph := NewDungeonGraph(12345, 80, -10)
				start := &RoomNode{ID: 0, Type: RoomSpawn, Width: 15, Height: 15}
				combat := &RoomNode{ID: 1, Type: RoomCombat, Width: 10, Height: 10}
				graph.AddRoom(start)
				graph.AddRoom(combat)
				graph.StartRoom = start
				graph.CriticalPath = []*RoomNode{start, combat}
				return graph
			}(),
			wantErr: true,
			errMsg:  "invalid dimensions",
		},
		{
			name: "critical_path_too_short",
			result: func() *DungeonGraph {
				graph := NewDungeonGraph(12345, 80, 50)
				start := &RoomNode{ID: 0, Type: RoomSpawn, Width: 15, Height: 15}
				combat := &RoomNode{ID: 1, Type: RoomCombat, Width: 10, Height: 10}
				graph.AddRoom(start)
				graph.AddRoom(combat)
				graph.StartRoom = start
				graph.CriticalPath = []*RoomNode{start}
				return graph
			}(),
			wantErr: true,
			errMsg:  "too short",
		},
		{
			name: "critical_path_wrong_start",
			result: func() *DungeonGraph {
				graph := NewDungeonGraph(12345, 80, 50)
				start := &RoomNode{ID: 0, Type: RoomSpawn, Width: 15, Height: 15}
				combat := &RoomNode{ID: 1, Type: RoomCombat, Width: 10, Height: 10}
				graph.AddRoom(start)
				graph.AddRoom(combat)
				graph.StartRoom = start
				graph.CriticalPath = []*RoomNode{combat, start}
				return graph
			}(),
			wantErr: true,
			errMsg:  "doesn't start with start room",
		},
		{
			name: "negative_narrative_depth",
			result: func() *DungeonGraph {
				graph := NewDungeonGraph(12345, 80, 50)
				start := &RoomNode{ID: 0, Type: RoomSpawn, Width: 15, Height: 15}
				combat := &RoomNode{ID: 1, Type: RoomCombat, Width: 10, Height: 10}
				graph.AddRoom(start)
				graph.AddRoom(combat)
				graph.StartRoom = start
				graph.CriticalPath = []*RoomNode{start, combat}
				graph.NarrativeDepth = -1
				return graph
			}(),
			wantErr: true,
			errMsg:  "negative",
		},
		{
			name: "boss_room_zero_depth",
			result: func() *DungeonGraph {
				graph := NewDungeonGraph(12345, 80, 50)
				start := &RoomNode{ID: 0, Type: RoomSpawn, Width: 15, Height: 15}
				boss := &RoomNode{ID: 1, Type: RoomBoss, Width: 20, Height: 20}
				graph.AddRoom(start)
				graph.AddRoom(boss)
				graph.StartRoom = start
				graph.BossRoom = boss
				graph.CriticalPath = []*RoomNode{start, boss}
				graph.NarrativeDepth = 0
				return graph
			}(),
			wantErr: true,
			errMsg:  "must be positive when boss room exists",
		},
		{
			name: "room_invalid_dimensions_zero_width",
			result: func() *DungeonGraph {
				graph := NewDungeonGraph(12345, 80, 50)
				start := &RoomNode{ID: 0, Type: RoomSpawn, Width: 0, Height: 15}
				combat := &RoomNode{ID: 1, Type: RoomCombat, Width: 10, Height: 10}
				graph.AddRoom(start)
				graph.AddRoom(combat)
				graph.StartRoom = start
				graph.CriticalPath = []*RoomNode{start, combat}
				return graph
			}(),
			wantErr: true,
			errMsg:  "invalid dimensions",
		},
		{
			name: "room_invalid_dimensions_negative",
			result: func() *DungeonGraph {
				graph := NewDungeonGraph(12345, 80, 50)
				start := &RoomNode{ID: 0, Type: RoomSpawn, Width: 15, Height: -5}
				combat := &RoomNode{ID: 1, Type: RoomCombat, Width: 10, Height: 10}
				graph.AddRoom(start)
				graph.AddRoom(combat)
				graph.StartRoom = start
				graph.CriticalPath = []*RoomNode{start, combat}
				return graph
			}(),
			wantErr: true,
			errMsg:  "invalid dimensions",
		},
		{
			name: "connection_wrong_from",
			result: func() *DungeonGraph {
				graph := NewDungeonGraph(12345, 80, 50)
				start := &RoomNode{ID: 0, Type: RoomSpawn, Width: 15, Height: 15}
				combat := &RoomNode{ID: 1, Type: RoomCombat, Width: 10, Height: 10}
				boss := &RoomNode{ID: 2, Type: RoomBoss, Width: 20, Height: 20}
				graph.AddRoom(start)
				graph.AddRoom(combat)
				graph.AddRoom(boss)
				// Create connection with wrong From pointer
				conn := &EdgeConnection{
					From:   boss, // Wrong - should be start
					To:     combat,
					Type:   ConnectionCorridor,
					Weight: 1.0,
				}
				start.Connections = append(start.Connections, conn)
				graph.StartRoom = start
				graph.CriticalPath = []*RoomNode{start, combat}
				return graph
			}(),
			wantErr: true,
			errMsg:  "mismatched From pointer",
		},
		{
			name: "connection_nil_to",
			result: func() *DungeonGraph {
				graph := NewDungeonGraph(12345, 80, 50)
				start := &RoomNode{ID: 0, Type: RoomSpawn, Width: 15, Height: 15}
				combat := &RoomNode{ID: 1, Type: RoomCombat, Width: 10, Height: 10}
				graph.AddRoom(start)
				graph.AddRoom(combat)
				conn := &EdgeConnection{
					From:   start,
					To:     nil,
					Type:   ConnectionCorridor,
					Weight: 1.0,
				}
				start.Connections = append(start.Connections, conn)
				graph.StartRoom = start
				graph.CriticalPath = []*RoomNode{start, combat}
				return graph
			}(),
			wantErr: true,
			errMsg:  "nil To pointer",
		},
		{
			name: "connection_invalid_weight",
			result: func() *DungeonGraph {
				graph := NewDungeonGraph(12345, 80, 50)
				start := &RoomNode{ID: 0, Type: RoomSpawn, Width: 15, Height: 15}
				combat := &RoomNode{ID: 1, Type: RoomCombat, Width: 10, Height: 10}
				graph.AddRoom(start)
				graph.AddRoom(combat)
				conn := &EdgeConnection{
					From:   start,
					To:     combat,
					Type:   ConnectionCorridor,
					Weight: 0.0, // Invalid - must be > 0
				}
				start.Connections = append(start.Connections, conn)
				graph.StartRoom = start
				graph.CriticalPath = []*RoomNode{start, combat}
				return graph
			}(),
			wantErr: true,
			errMsg:  "invalid weight",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Validate(tt.result)

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

// TestGraphGrammarGenerator_GenerateAndValidate tests that generated graphs pass validation
func TestGraphGrammarGenerator_GenerateAndValidate(t *testing.T) {
	config := GetFantasyConfig(12345)
	gen := NewGraphGrammarGenerator(config)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    genre,
				Custom: map[string]interface{}{
					"width":  80,
					"height": 50,
				},
			}

			result, err := gen.Generate(12345, params)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			err = gen.Validate(result)
			if err != nil {
				t.Errorf("Validate() failed for generated graph: %v", err)
			}
		})
	}
}

// TestGraphGrammarGenerator_DeterminismWithInterface tests that interface methods are deterministic
func TestGraphGrammarGenerator_DeterminismWithInterface(t *testing.T) {
	config := GetFantasyConfig(12345)
	gen := NewGraphGrammarGenerator(config)

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  80,
			"height": 50,
		},
	}

	result1, err1 := gen.Generate(12345, params)
	result2, err2 := gen.Generate(12345, params)

	if err1 != nil || err2 != nil {
		t.Fatalf("Errors: %v, %v", err1, err2)
	}

	graph1 := result1.(*DungeonGraph)
	graph2 := result2.(*DungeonGraph)

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

	// Verify same dimensions
	if graph1.Width != graph2.Width || graph1.Height != graph2.Height {
		t.Errorf("Dimensions differ: %dx%d vs %dx%d",
			graph1.Width, graph1.Height, graph2.Width, graph2.Height)
	}
}

// BenchmarkGraphGrammarGenerator_Generate benchmarks the interface Generate method
func BenchmarkGraphGrammarGenerator_Generate(b *testing.B) {
	config := GetFantasyConfig(12345)
	gen := NewGraphGrammarGenerator(config)

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  80,
			"height": 50,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(12345, params)
		if err != nil {
			b.Fatalf("Generate() error = %v", err)
		}
	}
}

// BenchmarkGraphGrammarGenerator_Validate benchmarks the Validate method
func BenchmarkGraphGrammarGenerator_Validate(b *testing.B) {
	config := GetFantasyConfig(12345)
	gen := NewGraphGrammarGenerator(config)

	// Generate a graph once to validate repeatedly
	graph, err := gen.GenerateGraph(80, 50)
	if err != nil {
		b.Fatalf("GenerateGraph() error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := gen.Validate(graph)
		if err != nil {
			b.Fatalf("Validate() error = %v", err)
		}
	}
}

// contains is a helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			hasSubstring(s, substr)))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestGraphToTerrain(t *testing.T) {
	tests := []struct {
		name            string
		setupGraph      func() *DungeonGraph
		expectedRooms   int
		validateTerrain func(*testing.T, *Terrain)
	}{
		{
			name: "simple graph with two connected rooms",
			setupGraph: func() *DungeonGraph {
				graph := NewDungeonGraph(12345, 100, 80)

				// Create start room
				startRoom := &RoomNode{
					ID:       0,
					Type:     RoomSpawn,
					Position: Point{X: 10, Y: 10},
					Width:    8,
					Height:   6,
				}
				graph.AddRoom(startRoom)
				graph.StartRoom = startRoom

				// Create boss room
				bossRoom := &RoomNode{
					ID:       1,
					Type:     RoomBoss,
					Position: Point{X: 30, Y: 20},
					Width:    10,
					Height:   8,
				}
				graph.AddRoom(bossRoom)
				graph.BossRoom = bossRoom

				// Connect them
				graph.ConnectRooms(startRoom, bossRoom, ConnectionCorridor, 1.0)

				return graph
			},
			expectedRooms: 2,
			validateTerrain: func(t *testing.T, terrain *Terrain) {
				// Check dimensions
				if terrain.Width != 100 || terrain.Height != 80 {
					t.Errorf("unexpected terrain dimensions: %dx%d, want 100x80", terrain.Width, terrain.Height)
				}

				// Check that start room area has floor tiles (avoid center where stairs might be)
				if terrain.GetTile(11, 11) != TileFloor && terrain.GetTile(11, 11) != TileStairsUp {
					t.Errorf("expected floor or stairs tile in start room, got %v", terrain.GetTile(11, 11))
				}

				// Check that boss room area has floor tiles (avoid center where stairs might be)
				if terrain.GetTile(31, 21) != TileFloor && terrain.GetTile(31, 21) != TileStairsDown {
					t.Errorf("expected floor or stairs tile in boss room, got %v", terrain.GetTile(31, 21))
				}

				// Check stairs placement
				if len(terrain.StairsUp) != 1 {
					t.Errorf("expected 1 stairs up, got %d", len(terrain.StairsUp))
				}
				if len(terrain.StairsDown) != 1 {
					t.Errorf("expected 1 stairs down, got %d", len(terrain.StairsDown))
				}
			},
		},
		{
			name: "graph with multiple room types",
			setupGraph: func() *DungeonGraph {
				graph := NewDungeonGraph(54321, 120, 100)

				rooms := []*RoomNode{
					{ID: 0, Type: RoomSpawn, Position: Point{X: 5, Y: 5}, Width: 6, Height: 6},
					{ID: 1, Type: RoomCombat, Position: Point{X: 20, Y: 5}, Width: 8, Height: 6},
					{ID: 2, Type: RoomTreasure, Position: Point{X: 35, Y: 5}, Width: 7, Height: 7},
					{ID: 3, Type: RoomBoss, Position: Point{X: 50, Y: 5}, Width: 10, Height: 10},
				}

				for _, room := range rooms {
					graph.AddRoom(room)
				}
				graph.StartRoom = rooms[0]
				graph.BossRoom = rooms[3]

				// Create a linear path
				for i := 0; i < len(rooms)-1; i++ {
					graph.ConnectRooms(rooms[i], rooms[i+1], ConnectionCorridor, 1.0)
				}

				return graph
			},
			expectedRooms: 4,
			validateTerrain: func(t *testing.T, terrain *Terrain) {
				// Verify all rooms were created
				if len(terrain.Rooms) != 4 {
					t.Errorf("expected 4 rooms, got %d", len(terrain.Rooms))
				}

				// Verify room types are preserved
				roomTypes := make(map[RoomType]int)
				for _, room := range terrain.Rooms {
					roomTypes[room.Type]++
				}

				if roomTypes[RoomSpawn] != 1 {
					t.Errorf("expected 1 spawn room, got %d", roomTypes[RoomSpawn])
				}
				if roomTypes[RoomCombat] != 1 {
					t.Errorf("expected 1 combat room, got %d", roomTypes[RoomCombat])
				}
				if roomTypes[RoomTreasure] != 1 {
					t.Errorf("expected 1 treasure room, got %d", roomTypes[RoomTreasure])
				}
				if roomTypes[RoomBoss] != 1 {
					t.Errorf("expected 1 boss room, got %d", roomTypes[RoomBoss])
				}
			},
		},
		{
			name: "graph with secret connections",
			setupGraph: func() *DungeonGraph {
				graph := NewDungeonGraph(99999, 80, 60)

				room1 := &RoomNode{ID: 0, Type: RoomSpawn, Position: Point{X: 10, Y: 10}, Width: 6, Height: 6}
				room2 := &RoomNode{ID: 1, Type: RoomSecret, Position: Point{X: 30, Y: 10}, Width: 5, Height: 5}

				graph.AddRoom(room1)
				graph.AddRoom(room2)
				graph.StartRoom = room1

				// Connect with secret door
				graph.ConnectRooms(room1, room2, ConnectionSecret, 1.0)

				return graph
			},
			expectedRooms: 2,
			validateTerrain: func(t *testing.T, terrain *Terrain) {
				// Count secret doors in corridor between rooms
				secretDoorCount := 0
				for y := 0; y < terrain.Height; y++ {
					for x := 0; x < terrain.Width; x++ {
						if terrain.GetTile(x, y) == TileSecretDoor {
							secretDoorCount++
						}
					}
				}

				if secretDoorCount == 0 {
					t.Error("expected at least one secret door tile in terrain")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph := tt.setupGraph()
			terrain := GraphToTerrain(graph)

			if terrain == nil {
				t.Fatal("GraphToTerrain returned nil")
			}

			if len(terrain.Rooms) != tt.expectedRooms {
				t.Errorf("expected %d rooms, got %d", tt.expectedRooms, len(terrain.Rooms))
			}

			if terrain.Seed != graph.Seed {
				t.Errorf("terrain seed %d doesn't match graph seed %d", terrain.Seed, graph.Seed)
			}

			if tt.validateTerrain != nil {
				tt.validateTerrain(t, terrain)
			}
		})
	}
}

func TestGraphToTerrain_EmptyGraph(t *testing.T) {
	graph := NewDungeonGraph(12345, 50, 50)
	terrain := GraphToTerrain(graph)

	if terrain == nil {
		t.Fatal("GraphToTerrain returned nil for empty graph")
	}

	if len(terrain.Rooms) != 0 {
		t.Errorf("expected 0 rooms for empty graph, got %d", len(terrain.Rooms))
	}

	// All tiles should be walls
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			if terrain.GetTile(x, y) != TileWall {
				t.Errorf("expected wall at (%d,%d) for empty graph, got %v", x, y, terrain.GetTile(x, y))
				return
			}
		}
	}
}

func TestGraphGrammarGenerator_GenerateAndConvert(t *testing.T) {
	// Integration test: Generate a graph and convert to terrain
	config := GetFantasyConfig(12345)
	gen := NewGraphGrammarGenerator(config)

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  100,
			"height": 80,
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("failed to generate graph: %v", err)
	}

	graph, ok := result.(*DungeonGraph)
	if !ok {
		t.Fatalf("expected *DungeonGraph, got %T", result)
	}

	// Convert to terrain
	terrain := GraphToTerrain(graph)
	if terrain == nil {
		t.Fatal("GraphToTerrain returned nil")
	}

	// Validate terrain structure
	if terrain.Width != 100 || terrain.Height != 80 {
		t.Errorf("unexpected terrain dimensions: %dx%d", terrain.Width, terrain.Height)
	}

	if len(terrain.Rooms) == 0 {
		t.Error("converted terrain has no rooms")
	}

	// Ensure there are walkable tiles
	walkableCount := 0
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			if terrain.IsWalkable(x, y) {
				walkableCount++
			}
		}
	}

	if walkableCount == 0 {
		t.Error("converted terrain has no walkable tiles")
	}
}
