package world

import "testing"

// TestNewMap verifies that a new map is created with correct dimensions and initialization.
func TestNewMap(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		seed   int64
	}{
		{"small_map", 10, 10, 12345},
		{"rectangular_map", 20, 15, 67890},
		{"large_map", 100, 80, 11111},
		{"single_tile", 1, 1, 99999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMap(tt.width, tt.height, tt.seed)

			// Verify dimensions
			if m.Width != tt.width {
				t.Errorf("Expected width %d, got %d", tt.width, m.Width)
			}
			if m.Height != tt.height {
				t.Errorf("Expected height %d, got %d", tt.height, m.Height)
			}

			// Verify seed
			if m.Seed != tt.seed {
				t.Errorf("Expected seed %d, got %d", tt.seed, m.Seed)
			}

			// Verify tiles array size
			expectedSize := tt.width * tt.height
			if len(m.Tiles) != expectedSize {
				t.Errorf("Expected %d tiles, got %d", expectedSize, len(m.Tiles))
			}

			// Verify all tiles are initialized properly
			for y := 0; y < tt.height; y++ {
				for x := 0; x < tt.width; x++ {
					idx := y*tt.width + x
					tile := m.Tiles[idx]

					if tile.Type != TileEmpty {
						t.Errorf("Expected tile at (%d,%d) to be TileEmpty, got %v", x, y, tile.Type)
					}
					if !tile.Walkable {
						t.Errorf("Expected tile at (%d,%d) to be walkable", x, y)
					}
					if tile.X != x || tile.Y != y {
						t.Errorf("Expected tile coordinates (%d,%d), got (%d,%d)", x, y, tile.X, tile.Y)
					}
				}
			}
		})
	}
}

// TestGetTile verifies tile retrieval with valid and invalid coordinates.
func TestGetTile(t *testing.T) {
	m := NewMap(10, 10, 12345)

	tests := []struct {
		name      string
		x         int
		y         int
		expectNil bool
	}{
		{"top_left_corner", 0, 0, false},
		{"bottom_right_corner", 9, 9, false},
		{"middle_tile", 5, 5, false},
		{"negative_x", -1, 5, true},
		{"negative_y", 5, -1, true},
		{"out_of_bounds_x", 10, 5, true},
		{"out_of_bounds_y", 5, 10, true},
		{"both_negative", -1, -1, true},
		{"both_out_of_bounds", 10, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tile := m.GetTile(tt.x, tt.y)

			if tt.expectNil {
				if tile != nil {
					t.Errorf("Expected nil tile for coordinates (%d,%d), got %v", tt.x, tt.y, tile)
				}
			} else {
				if tile == nil {
					t.Errorf("Expected valid tile for coordinates (%d,%d), got nil", tt.x, tt.y)
				} else {
					if tile.X != tt.x || tile.Y != tt.y {
						t.Errorf("Expected tile coordinates (%d,%d), got (%d,%d)", tt.x, tt.y, tile.X, tile.Y)
					}
				}
			}
		})
	}
}

// TestSetTile verifies tile modification with valid and invalid coordinates.
func TestSetTile(t *testing.T) {
	m := NewMap(10, 10, 12345)

	tests := []struct {
		name       string
		x          int
		y          int
		tileType   TileType
		walkable   bool
		shouldSet  bool
	}{
		{"set_wall", 5, 5, TileWall, false, true},
		{"set_floor", 3, 3, TileFloor, true, true},
		{"set_door", 7, 7, TileDoor, true, true},
		{"set_water", 2, 2, TileWater, false, true},
		{"negative_x", -1, 5, TileWall, false, false},
		{"negative_y", 5, -1, TileWall, false, false},
		{"out_of_bounds_x", 10, 5, TileWall, false, false},
		{"out_of_bounds_y", 5, 10, TileWall, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newTile := Tile{
				Type:     tt.tileType,
				Walkable: tt.walkable,
			}

			m.SetTile(tt.x, tt.y, newTile)

			if tt.shouldSet {
				// Verify tile was set correctly
				tile := m.GetTile(tt.x, tt.y)
				if tile == nil {
					t.Fatalf("Expected tile to be set at (%d,%d)", tt.x, tt.y)
				}
				if tile.Type != tt.tileType {
					t.Errorf("Expected tile type %v at (%d,%d), got %v", tt.tileType, tt.x, tt.y, tile.Type)
				}
				if tile.Walkable != tt.walkable {
					t.Errorf("Expected walkable=%v at (%d,%d), got %v", tt.walkable, tt.x, tt.y, tile.Walkable)
				}
				// Coordinates should be updated by SetTile
				if tile.X != tt.x || tile.Y != tt.y {
					t.Errorf("Expected coordinates (%d,%d), got (%d,%d)", tt.x, tt.y, tile.X, tile.Y)
				}
			}
		})
	}
}

// TestIsWalkable verifies walkability checks for valid and invalid coordinates.
func TestIsWalkable(t *testing.T) {
	m := NewMap(10, 10, 12345)

	// Set up various tile types
	m.SetTile(1, 1, Tile{Type: TileFloor, Walkable: true})
	m.SetTile(2, 2, Tile{Type: TileWall, Walkable: false})
	m.SetTile(3, 3, Tile{Type: TileDoor, Walkable: true})
	m.SetTile(4, 4, Tile{Type: TileWater, Walkable: false})

	tests := []struct {
		name            string
		x               int
		y               int
		expectedWalkable bool
	}{
		{"empty_tile_default", 0, 0, true},
		{"floor_tile", 1, 1, true},
		{"wall_tile", 2, 2, false},
		{"door_tile", 3, 3, true},
		{"water_tile", 4, 4, false},
		{"out_of_bounds_negative", -1, -1, false},
		{"out_of_bounds_positive", 10, 10, false},
		{"out_of_bounds_x", 15, 5, false},
		{"out_of_bounds_y", 5, 15, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			walkable := m.IsWalkable(tt.x, tt.y)
			if walkable != tt.expectedWalkable {
				t.Errorf("Expected IsWalkable(%d,%d)=%v, got %v", tt.x, tt.y, tt.expectedWalkable, walkable)
			}
		})
	}
}

// TestNewWorldState verifies world state initialization.
func TestNewWorldState(t *testing.T) {
	ws := NewWorldState()

	if ws == nil {
		t.Fatal("Expected non-nil WorldState")
	}

	if ws.PlayerIDs == nil {
		t.Error("Expected PlayerIDs slice to be initialized")
	}

	if len(ws.PlayerIDs) != 0 {
		t.Errorf("Expected empty PlayerIDs slice, got length %d", len(ws.PlayerIDs))
	}

	if ws.State == nil {
		t.Error("Expected State map to be initialized")
	}

	if len(ws.State) != 0 {
		t.Errorf("Expected empty State map, got length %d", len(ws.State))
	}

	if ws.Time != 0 {
		t.Errorf("Expected Time to be 0, got %f", ws.Time)
	}

	if ws.CurrentMap != nil {
		t.Error("Expected CurrentMap to be nil initially")
	}
}

// TestWorldState_MapAssignment verifies that maps can be assigned to world state.
func TestWorldState_MapAssignment(t *testing.T) {
	ws := NewWorldState()
	m := NewMap(20, 20, 54321)

	ws.CurrentMap = m

	if ws.CurrentMap != m {
		t.Error("Expected CurrentMap to be assigned")
	}

	if ws.CurrentMap.Width != 20 {
		t.Errorf("Expected map width 20, got %d", ws.CurrentMap.Width)
	}
}

// TestWorldState_PlayerIDs verifies player ID management.
func TestWorldState_PlayerIDs(t *testing.T) {
	ws := NewWorldState()

	// Add player IDs
	ws.PlayerIDs = append(ws.PlayerIDs, 100)
	ws.PlayerIDs = append(ws.PlayerIDs, 200)
	ws.PlayerIDs = append(ws.PlayerIDs, 300)

	if len(ws.PlayerIDs) != 3 {
		t.Errorf("Expected 3 player IDs, got %d", len(ws.PlayerIDs))
	}

	expectedIDs := []uint64{100, 200, 300}
	for i, id := range ws.PlayerIDs {
		if id != expectedIDs[i] {
			t.Errorf("Expected player ID %d at index %d, got %d", expectedIDs[i], i, id)
		}
	}
}

// TestWorldState_CustomState verifies custom state data storage.
func TestWorldState_CustomState(t *testing.T) {
	ws := NewWorldState()

	// Add custom state data
	ws.State["level"] = 5
	ws.State["name"] = "TestWorld"
	ws.State["active"] = true
	ws.State["score"] = 1000.5

	tests := []struct {
		key           string
		expectedValue interface{}
	}{
		{"level", 5},
		{"name", "TestWorld"},
		{"active", true},
		{"score", 1000.5},
	}

	for _, tt := range tests {
		t.Run("key_"+tt.key, func(t *testing.T) {
			value, ok := ws.State[tt.key]
			if !ok {
				t.Errorf("Expected key '%s' to exist in State", tt.key)
			}
			if value != tt.expectedValue {
				t.Errorf("Expected value %v for key '%s', got %v", tt.expectedValue, tt.key, value)
			}
		})
	}
}

// TestWorldState_TimeProgression verifies time can be updated.
func TestWorldState_TimeProgression(t *testing.T) {
	ws := NewWorldState()

	if ws.Time != 0 {
		t.Errorf("Expected initial time 0, got %f", ws.Time)
	}

	ws.Time = 10.5
	if ws.Time != 10.5 {
		t.Errorf("Expected time 10.5, got %f", ws.Time)
	}

	ws.Time += 5.5
	if ws.Time != 16.0 {
		t.Errorf("Expected time 16.0, got %f", ws.Time)
	}
}

// TestTileType_Constants verifies tile type constants are distinct.
func TestTileType_Constants(t *testing.T) {
	types := []TileType{
		TileEmpty,
		TileFloor,
		TileWall,
		TileDoor,
		TileWater,
		TileLava,
		TileGrass,
		TileStone,
	}

	// Verify all constants are unique
	seen := make(map[TileType]bool)
	for _, tileType := range types {
		if seen[tileType] {
			t.Errorf("Duplicate tile type value: %v", tileType)
		}
		seen[tileType] = true
	}

	// Verify expected number of constants
	if len(types) != 8 {
		t.Errorf("Expected 8 tile type constants, got %d", len(types))
	}
}

// TestMap_Genre verifies genre field can be set and retrieved.
func TestMap_Genre(t *testing.T) {
	m := NewMap(10, 10, 12345)

	if m.Genre != "" {
		t.Errorf("Expected empty genre initially, got '%s'", m.Genre)
	}

	m.Genre = "fantasy"
	if m.Genre != "fantasy" {
		t.Errorf("Expected genre 'fantasy', got '%s'", m.Genre)
	}
}

// TestMap_EdgeCases verifies edge cases in map operations.
func TestMap_EdgeCases(t *testing.T) {
	t.Run("zero_dimensions", func(t *testing.T) {
		// While not typical, ensure it doesn't crash
		m := NewMap(0, 0, 12345)
		if m.Width != 0 || m.Height != 0 {
			t.Error("Expected zero dimensions")
		}
		if len(m.Tiles) != 0 {
			t.Errorf("Expected empty tiles slice, got length %d", len(m.Tiles))
		}
	})

	t.Run("large_map", func(t *testing.T) {
		// Test with larger dimensions
		m := NewMap(1000, 1000, 99999)
		if len(m.Tiles) != 1000000 {
			t.Errorf("Expected 1000000 tiles, got %d", len(m.Tiles))
		}
		// Verify we can access corners
		tile := m.GetTile(999, 999)
		if tile == nil {
			t.Error("Expected to access bottom-right corner")
		}
	})
}
