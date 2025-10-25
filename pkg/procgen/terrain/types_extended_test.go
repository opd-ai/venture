package terrain

import (
	"testing"
)

// TestTileType_IsWalkableTile tests the walkability of all tile types.
func TestTileType_IsWalkableTile(t *testing.T) {
	tests := []struct {
		name     string
		tileType TileType
		walkable bool
	}{
		{"Wall is not walkable", TileWall, false},
		{"Floor is walkable", TileFloor, true},
		{"Door is walkable", TileDoor, true},
		{"Corridor is walkable", TileCorridor, true},
		{"Shallow water is walkable", TileWaterShallow, true},
		{"Deep water is not walkable", TileWaterDeep, false},
		{"Tree is not walkable", TileTree, false},
		{"Stairs up is walkable", TileStairsUp, true},
		{"Stairs down is walkable", TileStairsDown, true},
		{"Trap door is walkable", TileTrapDoor, true},
		{"Secret door is walkable", TileSecretDoor, true},
		{"Bridge is walkable", TileBridge, true},
		{"Structure is not walkable", TileStructure, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tileType.IsWalkableTile()
			if result != tt.walkable {
				t.Errorf("TileType.IsWalkableTile() = %v, want %v", result, tt.walkable)
			}
		})
	}
}

// TestTileType_IsTransparent tests the transparency of all tile types.
func TestTileType_IsTransparent(t *testing.T) {
	tests := []struct {
		name        string
		tileType    TileType
		transparent bool
	}{
		{"Wall is not transparent", TileWall, false},
		{"Floor is transparent", TileFloor, true},
		{"Door is transparent", TileDoor, true},
		{"Corridor is transparent", TileCorridor, true},
		{"Shallow water is transparent", TileWaterShallow, true},
		{"Deep water is transparent", TileWaterDeep, true},
		{"Tree is not transparent", TileTree, false},
		{"Stairs up is transparent", TileStairsUp, true},
		{"Stairs down is transparent", TileStairsDown, true},
		{"Trap door is transparent", TileTrapDoor, true},
		{"Secret door is not transparent", TileSecretDoor, false},
		{"Bridge is transparent", TileBridge, true},
		{"Structure is not transparent", TileStructure, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tileType.IsTransparent()
			if result != tt.transparent {
				t.Errorf("TileType.IsTransparent() = %v, want %v", result, tt.transparent)
			}
		})
	}
}

// TestTileType_MovementCost tests the movement cost of all tile types.
func TestTileType_MovementCost(t *testing.T) {
	tests := []struct {
		name     string
		tileType TileType
		cost     float64
	}{
		{"Wall is impassable", TileWall, -1},
		{"Floor has normal cost", TileFloor, 1.0},
		{"Door has normal cost", TileDoor, 1.0},
		{"Corridor has normal cost", TileCorridor, 1.0},
		{"Shallow water slows movement", TileWaterShallow, 2.0},
		{"Deep water is impassable", TileWaterDeep, -1},
		{"Tree is impassable", TileTree, -1},
		{"Stairs up has normal cost", TileStairsUp, 1.0},
		{"Stairs down has normal cost", TileStairsDown, 1.0},
		{"Trap door slows movement", TileTrapDoor, 1.5},
		{"Secret door has normal cost", TileSecretDoor, 1.0},
		{"Bridge has normal cost", TileBridge, 1.0},
		{"Structure is impassable", TileStructure, -1},
		{"Unknown is impassable", TileType(999), -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tileType.MovementCost()
			if result != tt.cost {
				t.Errorf("TileType.MovementCost() = %v, want %v", result, tt.cost)
			}
		})
	}
}

// TestTerrain_MultiLevel tests the multi-level terrain support.
func TestTerrain_MultiLevel(t *testing.T) {
	terr := NewTerrain(10, 10, 12345)

	// Check initial state
	if terr.Level != 0 {
		t.Errorf("New terrain should start at level 0, got %d", terr.Level)
	}

	if len(terr.StairsUp) != 0 {
		t.Errorf("New terrain should have no stairs up, got %d", len(terr.StairsUp))
	}

	if len(terr.StairsDown) != 0 {
		t.Errorf("New terrain should have no stairs down, got %d", len(terr.StairsDown))
	}

	// Set level
	terr.Level = 3
	if terr.Level != 3 {
		t.Errorf("Terrain level should be 3, got %d", terr.Level)
	}
}

// TestTerrain_AddStairs tests adding stairs to terrain.
func TestTerrain_AddStairs(t *testing.T) {
	terr := NewTerrain(10, 10, 12345)

	// Add stairs up
	terr.AddStairs(5, 5, true)
	if len(terr.StairsUp) != 1 {
		t.Errorf("Should have 1 stairs up, got %d", len(terr.StairsUp))
	}
	if terr.GetTile(5, 5) != TileStairsUp {
		t.Errorf("Tile at (5,5) should be stairs up, got %v", terr.GetTile(5, 5))
	}

	// Add duplicate stairs up (should not add)
	terr.AddStairs(5, 5, true)
	if len(terr.StairsUp) != 1 {
		t.Errorf("Should still have 1 stairs up after duplicate, got %d", len(terr.StairsUp))
	}

	// Add stairs down
	terr.AddStairs(2, 2, false)
	if len(terr.StairsDown) != 1 {
		t.Errorf("Should have 1 stairs down, got %d", len(terr.StairsDown))
	}
	if terr.GetTile(2, 2) != TileStairsDown {
		t.Errorf("Tile at (2,2) should be stairs down, got %v", terr.GetTile(2, 2))
	}

	// Try to add stairs out of bounds (should be ignored)
	terr.AddStairs(-1, -1, true)
	terr.AddStairs(100, 100, false)
	if len(terr.StairsUp) != 1 || len(terr.StairsDown) != 1 {
		t.Errorf("Out of bounds stairs should be ignored")
	}
}

// TestTerrain_ValidateStairPlacement tests stair placement validation.
func TestTerrain_ValidateStairPlacement(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*Terrain)
		wantError bool
	}{
		{
			name: "Valid stairs placement",
			setup: func(terr *Terrain) {
				// Create a floor area with stairs
				for x := 4; x <= 6; x++ {
					for y := 4; y <= 6; y++ {
						terr.SetTile(x, y, TileFloor)
					}
				}
				terr.AddStairs(5, 5, true)
				terr.AddStairs(6, 6, false)
			},
			wantError: false,
		},
		{
			name: "Stairs with no accessible neighbors",
			setup: func(terr *Terrain) {
				// Place stairs surrounded by walls
				terr.AddStairs(5, 5, true)
			},
			wantError: true,
		},
		{
			name: "Stairs placed but tile mismatch",
			setup: func(terr *Terrain) {
				// Manually create mismatch
				terr.StairsUp = append(terr.StairsUp, Point{X: 5, Y: 5})
				terr.SetTile(5, 5, TileFloor) // Wrong tile type
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terr := NewTerrain(10, 10, 12345)
			tt.setup(terr)

			err := terr.ValidateStairPlacement()
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateStairPlacement() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestTerrain_IsInBounds tests the bounds checking.
func TestTerrain_IsInBounds(t *testing.T) {
	terr := NewTerrain(10, 10, 12345)

	tests := []struct {
		name     string
		x, y     int
		inBounds bool
	}{
		{"Origin is in bounds", 0, 0, true},
		{"Center is in bounds", 5, 5, true},
		{"Max corner is in bounds", 9, 9, true},
		{"Negative x is out of bounds", -1, 5, false},
		{"Negative y is out of bounds", 5, -1, false},
		{"X too large is out of bounds", 10, 5, false},
		{"Y too large is out of bounds", 5, 10, false},
		{"Both negative is out of bounds", -1, -1, false},
		{"Both too large is out of bounds", 10, 10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := terr.IsInBounds(tt.x, tt.y)
			if result != tt.inBounds {
				t.Errorf("IsInBounds(%d, %d) = %v, want %v", tt.x, tt.y, result, tt.inBounds)
			}
		})
	}
}

// TestTerrain_IsWalkable tests the updated IsWalkable method.
func TestTerrain_IsWalkable(t *testing.T) {
	terr := NewTerrain(10, 10, 12345)

	// Set up various tile types
	terr.SetTile(1, 1, TileFloor)
	terr.SetTile(2, 2, TileWall)
	terr.SetTile(3, 3, TileWaterShallow)
	terr.SetTile(4, 4, TileWaterDeep)
	terr.SetTile(5, 5, TileTree)
	terr.SetTile(6, 6, TileStairsUp)

	tests := []struct {
		name     string
		x, y     int
		walkable bool
	}{
		{"Floor is walkable", 1, 1, true},
		{"Wall is not walkable", 2, 2, false},
		{"Shallow water is walkable", 3, 3, true},
		{"Deep water is not walkable", 4, 4, false},
		{"Tree is not walkable", 5, 5, false},
		{"Stairs up is walkable", 6, 6, true},
		{"Out of bounds is not walkable", -1, -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := terr.IsWalkable(tt.x, tt.y)
			if result != tt.walkable {
				t.Errorf("IsWalkable(%d, %d) = %v, want %v", tt.x, tt.y, result, tt.walkable)
			}
		})
	}
}
