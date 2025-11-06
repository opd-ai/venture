package terrain

import (
	"testing"
)

// TestDiagonalWallTypes tests the new diagonal wall tile types.
func TestDiagonalWallTypes(t *testing.T) {
	tests := []struct {
		name         string
		tileType     TileType
		wantString   string
		wantWalkable bool
		wantWall     bool
		wantDiagonal bool
	}{
		{"WallNE", TileWallNE, "wall_ne", false, true, true},
		{"WallNW", TileWallNW, "wall_nw", false, true, true},
		{"WallSE", TileWallSE, "wall_se", false, true, true},
		{"WallSW", TileWallSW, "wall_sw", false, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tileType.String(); got != tt.wantString {
				t.Errorf("String() = %v, want %v", got, tt.wantString)
			}
			if got := tt.tileType.IsWalkableTile(); got != tt.wantWalkable {
				t.Errorf("IsWalkableTile() = %v, want %v", got, tt.wantWalkable)
			}
			if got := tt.tileType.IsWall(); got != tt.wantWall {
				t.Errorf("IsWall() = %v, want %v", got, tt.wantWall)
			}
			if got := tt.tileType.IsDiagonalWall(); got != tt.wantDiagonal {
				t.Errorf("IsDiagonalWall() = %v, want %v", got, tt.wantDiagonal)
			}
			if got := tt.tileType.MovementCost(); got != -1 {
				t.Errorf("MovementCost() = %v, want -1 (impassable)", got)
			}
		})
	}
}

// TestMultiLayerTileTypes tests the new multi-layer tile types.
func TestMultiLayerTileTypes(t *testing.T) {
	tests := []struct {
		name         string
		tileType     TileType
		wantString   string
		wantWalkable bool
		wantLayer    Layer
		wantCost     float64
	}{
		{"Platform", TilePlatform, "platform", true, LayerPlatform, 1.0},
		{"Ramp", TileRamp, "ramp", true, LayerGround, 1.2},
		{"RampUp", TileRampUp, "ramp_up", true, LayerGround, 1.2},
		{"RampDown", TileRampDown, "ramp_down", true, LayerGround, 1.2},
		{"LavaFlow", TileLavaFlow, "lava_flow", false, LayerWater, 3.0},
		{"Pit", TilePit, "pit", false, LayerWater, -1},
		{"WaterShallow", TileWaterShallow, "shallow_water", true, LayerWater, 2.0},
		{"Bridge", TileBridge, "bridge", true, LayerPlatform, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tileType.String(); got != tt.wantString {
				t.Errorf("String() = %v, want %v", got, tt.wantString)
			}
			if got := tt.tileType.IsWalkableTile(); got != tt.wantWalkable {
				t.Errorf("IsWalkableTile() = %v, want %v", got, tt.wantWalkable)
			}
			if got := tt.tileType.GetLayer(); got != tt.wantLayer {
				t.Errorf("GetLayer() = %v, want %v", got, tt.wantLayer)
			}
			if got := tt.tileType.MovementCost(); got != tt.wantCost {
				t.Errorf("MovementCost() = %v, want %v", got, tt.wantCost)
			}
		})
	}
}

// TestLayerTransitions tests layer transition logic.
func TestLayerTransitions(t *testing.T) {
	tests := []struct {
		name        string
		fromTile    TileType
		toLayer     Layer
		canTransit  bool
		description string
	}{
		{
			"Floor to ground",
			TileFloor,
			LayerGround,
			true,
			"Same layer transition always allowed",
		},
		{
			"Floor to platform",
			TileFloor,
			LayerPlatform,
			false,
			"Cannot transition to different layer without ramp/stairs",
		},
		{
			"Ramp to platform",
			TileRamp,
			LayerPlatform,
			true,
			"Ramps allow layer transitions",
		},
		{
			"RampUp to platform",
			TileRampUp,
			LayerPlatform,
			true,
			"RampUp allows upward transition",
		},
		{
			"RampDown to water",
			TileRampDown,
			LayerWater,
			true,
			"RampDown allows downward transition",
		},
		{
			"StairsUp to platform",
			TileStairsUp,
			LayerPlatform,
			true,
			"Stairs allow vertical movement",
		},
		{
			"StairsDown to water",
			TileStairsDown,
			LayerWater,
			true,
			"Stairs allow downward movement",
		},
		{
			"Platform to platform",
			TilePlatform,
			LayerPlatform,
			true,
			"Same layer transition on platforms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fromTile.CanTransitionToLayer(tt.toLayer); got != tt.canTransit {
				t.Errorf("CanTransitionToLayer() = %v, want %v: %s", got, tt.canTransit, tt.description)
			}
		})
	}
}

// TestDiagonalWallTransparency tests that diagonal walls block vision.
func TestDiagonalWallTransparency(t *testing.T) {
	diagonalWalls := []TileType{TileWallNE, TileWallNW, TileWallSE, TileWallSW}

	for _, wall := range diagonalWalls {
		t.Run(wall.String(), func(t *testing.T) {
			if wall.IsTransparent() {
				t.Errorf("%s should block vision (IsTransparent = false)", wall.String())
			}
		})
	}
}

// TestAxisAlignedWallStillWorks tests that regular walls still work correctly.
func TestAxisAlignedWallStillWorks(t *testing.T) {
	if !TileWall.IsWall() {
		t.Error("TileWall.IsWall() should return true")
	}
	if TileWall.IsDiagonalWall() {
		t.Error("TileWall.IsDiagonalWall() should return false")
	}
	if TileWall.IsWalkableTile() {
		t.Error("TileWall should not be walkable")
	}
	if TileWall.MovementCost() != -1 {
		t.Error("TileWall should be impassable (MovementCost = -1)")
	}
}

// TestTileWithLayer tests the Tile struct with layer information.
func TestTileWithLayer(t *testing.T) {
	tile := Tile{
		Type:  TilePlatform,
		X:     10,
		Y:     20,
		Layer: LayerPlatform,
	}

	if tile.Type != TilePlatform {
		t.Errorf("Tile.Type = %v, want TilePlatform", tile.Type)
	}
	if tile.X != 10 {
		t.Errorf("Tile.X = %v, want 10", tile.X)
	}
	if tile.Y != 20 {
		t.Errorf("Tile.Y = %v, want 20", tile.Y)
	}
	if tile.Layer != LayerPlatform {
		t.Errorf("Tile.Layer = %v, want LayerPlatform", tile.Layer)
	}
}

// TestLayerConstants tests that layer constants are defined correctly.
func TestLayerConstants(t *testing.T) {
	tests := []struct {
		name  string
		layer Layer
		want  int
	}{
		{"Ground", LayerGround, 0},
		{"Water", LayerWater, 1},
		{"Platform", LayerPlatform, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.layer) != tt.want {
				t.Errorf("Layer %s = %d, want %d", tt.name, tt.layer, tt.want)
			}
		})
	}
}

// BenchmarkIsDiagonalWall benchmarks the IsDiagonalWall method.
func BenchmarkIsDiagonalWall(b *testing.B) {
	tiles := []TileType{TileWallNE, TileWall, TileFloor, TileWallSW}
	for i := 0; i < b.N; i++ {
		for _, tile := range tiles {
			_ = tile.IsDiagonalWall()
		}
	}
}

// BenchmarkGetLayer benchmarks the GetLayer method.
func BenchmarkGetLayer(b *testing.B) {
	tiles := []TileType{TileFloor, TilePlatform, TileWaterShallow, TileLavaFlow}
	for i := 0; i < b.N; i++ {
		for _, tile := range tiles {
			_ = tile.GetLayer()
		}
	}
}

// BenchmarkCanTransitionToLayer benchmarks the CanTransitionToLayer method.
func BenchmarkCanTransitionToLayer(b *testing.B) {
	tiles := []TileType{TileFloor, TileRamp, TileStairsUp, TilePlatform}
	layers := []Layer{LayerGround, LayerWater, LayerPlatform}
	for i := 0; i < b.N; i++ {
		for _, tile := range tiles {
			for _, layer := range layers {
				_ = tile.CanTransitionToLayer(layer)
			}
		}
	}
}
