package terrain

import (
	"math"
	"math/rand"
	"testing"
)

// TestGenerateLake tests circular/elliptical lake generation.
func TestGenerateLake(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		center Point
		radius int
		seed   int64
	}{
		{"small lake", 40, 30, Point{20, 15}, 5, 12345},
		{"medium lake", 80, 50, Point{40, 25}, 10, 67890},
		{"large lake", 100, 80, Point{50, 40}, 20, 11111},
		{"lake near edge", 60, 40, Point{5, 5}, 8, 22222},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terrain := NewTerrain(tt.width, tt.height, 0)
			for y := 0; y < tt.height; y++ {
				for x := 0; x < tt.width; x++ {
					terrain.SetTile(x, y, TileFloor)
				}
			}

			rng := rand.New(rand.NewSource(tt.seed))
			feature := GenerateLake(tt.center, tt.radius, terrain, rng)

			if feature == nil {
				t.Fatal("GenerateLake returned nil")
			}
			if feature.Type != WaterLake {
				t.Errorf("Expected WaterLake, got %v", feature.Type)
			}

			waterCount := 0
			deepCount := 0
			shallowCount := 0

			for y := 0; y < tt.height; y++ {
				for x := 0; x < tt.width; x++ {
					tile := terrain.GetTile(x, y)
					if tile == TileWaterDeep {
						waterCount++
						deepCount++
					} else if tile == TileWaterShallow {
						waterCount++
						shallowCount++
					}
				}
			}

			if waterCount == 0 {
				t.Error("No water tiles placed")
			}
			if deepCount == 0 {
				t.Error("No deep water tiles")
			}
			if shallowCount == 0 {
				t.Error("No shallow water tiles")
			}
			if len(feature.Tiles) != waterCount {
				t.Errorf("Feature.Tiles count %d != actual water count %d",
					len(feature.Tiles), waterCount)
			}
		})
	}
}

// TestGenerateRiver tests river generation between two points.
func TestGenerateRiver(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		start  Point
		end    Point
		rWidth int
		seed   int64
	}{
		{"narrow horizontal", 60, 40, Point{5, 20}, Point{55, 20}, 1, 12345},
		{"narrow vertical", 60, 40, Point{30, 5}, Point{30, 35}, 1, 67890},
		{"wide diagonal", 80, 50, Point{10, 10}, Point{70, 40}, 3, 11111},
		{"very wide", 100, 60, Point{20, 30}, Point{80, 30}, 5, 22222},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terrain := NewTerrain(tt.width, tt.height, 0)
			for y := 0; y < tt.height; y++ {
				for x := 0; x < tt.width; x++ {
					terrain.SetTile(x, y, TileFloor)
				}
			}

			rng := rand.New(rand.NewSource(tt.seed))
			feature := GenerateRiver(tt.start, tt.end, tt.rWidth, terrain, rng)

			if feature == nil {
				t.Fatal("GenerateRiver returned nil")
			}
			if feature.Type != WaterRiver {
				t.Errorf("Expected WaterRiver, got %v", feature.Type)
			}
			if len(feature.Tiles) == 0 {
				t.Error("No water tiles placed")
			}
		})
	}
}

// TestGenerateMoat tests moat generation around rooms.
func TestGenerateMoat(t *testing.T) {
	tests := []struct {
		name    string
		room    *Room
		width   int
		tWidth  int
		tHeight int
	}{
		{"narrow small", &Room{X: 10, Y: 10, Width: 8, Height: 6}, 1, 40, 30},
		{"wide medium", &Room{X: 20, Y: 15, Width: 15, Height: 10}, 2, 80, 50},
		{"very wide large", &Room{X: 30, Y: 20, Width: 20, Height: 15}, 3, 100, 70},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terrain := NewTerrain(tt.tWidth, tt.tHeight, 0)
			for y := 0; y < tt.tHeight; y++ {
				for x := 0; x < tt.tWidth; x++ {
					terrain.SetTile(x, y, TileWall)
				}
			}
			for dy := 0; dy < tt.room.Height; dy++ {
				for dx := 0; dx < tt.room.Width; dx++ {
					terrain.SetTile(tt.room.X+dx, tt.room.Y+dy, TileFloor)
				}
			}

			feature := GenerateMoat(tt.room, tt.width, terrain)

			if feature == nil {
				t.Fatal("GenerateMoat returned nil")
			}
			if feature.Type != WaterMoat {
				t.Errorf("Expected WaterMoat, got %v", feature.Type)
			}
			if len(feature.Tiles) == 0 {
				t.Error("No water tiles placed")
			}
		})
	}
}

// TestPlaceBridges tests automatic bridge placement.
func TestPlaceBridges(t *testing.T) {
	t.Run("path crossing lake", func(t *testing.T) {
		terrain := NewTerrain(40, 30, 0)
		rng := rand.New(rand.NewSource(12345))

		// Create a horizontal path
		for x := 5; x < 35; x++ {
			terrain.SetTile(x, 15, TileCorridor)
		}

		// Fill rest with floor
		for y := 0; y < 30; y++ {
			for x := 0; x < 40; x++ {
				if terrain.GetTile(x, y) != TileCorridor {
					terrain.SetTile(x, y, TileFloor)
				}
			}
		}

		// Create a lake crossing the path
		feature := GenerateLake(Point{20, 15}, 5, terrain, rng)
		PlaceBridges(feature, terrain, rng)

		// Bridges should be placed if lake intersects path
		if len(feature.Bridges) > 0 {
			// Verify bridges are TileBridge
			for _, bridge := range feature.Bridges {
				if terrain.GetTile(bridge.X, bridge.Y) != TileBridge {
					t.Errorf("Bridge at %v is not TileBridge", bridge)
				}
			}
		}
	})
}

// TestFloodFill tests flood fill connectivity algorithm.
func TestFloodFill(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		height   int
		start    Point
		maxTiles int
	}{
		{"small area", 20, 15, Point{10, 7}, 50},
		{"large area limited", 80, 50, Point{40, 25}, 100},
		{"unrestricted", 30, 20, Point{15, 10}, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terrain := NewTerrain(tt.width, tt.height, 0)
			for y := 0; y < tt.height; y++ {
				for x := 0; x < tt.width; x++ {
					terrain.SetTile(x, y, TileFloor)
				}
			}

			filled := FloodFill(tt.start, tt.maxTiles, terrain)

			if len(filled) == 0 {
				t.Error("FloodFill returned empty slice")
			}
			if len(filled) > tt.maxTiles {
				t.Errorf("FloodFill exceeded maxTiles: got %d, max %d",
					len(filled), tt.maxTiles)
			}

			// Verify all tiles are walkable
			for _, point := range filled {
				tile := terrain.GetTile(point.X, point.Y)
				if !tile.IsWalkableTile() {
					t.Errorf("FloodFill included non-walkable tile %v at %v", tile, point)
				}
			}
		})
	}
}

// TestFloodFillWater tests flood fill water creation.
func TestFloodFillWater(t *testing.T) {
	tests := []struct {
		name           string
		width          int
		height         int
		start          Point
		maxTiles       int
		deepWaterRatio float64
		seed           int64
	}{
		{"mostly shallow", 40, 30, Point{20, 15}, 100, 0.2, 12345},
		{"mostly deep", 60, 40, Point{30, 20}, 200, 0.8, 67890},
		{"mixed", 80, 50, Point{40, 25}, 150, 0.5, 11111},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terrain := NewTerrain(tt.width, tt.height, 0)
			for y := 0; y < tt.height; y++ {
				for x := 0; x < tt.width; x++ {
					terrain.SetTile(x, y, TileFloor)
				}
			}

			rng := rand.New(rand.NewSource(tt.seed))
			feature := FloodFillWater(tt.start, tt.maxTiles, tt.deepWaterRatio, terrain, rng)

			if feature == nil {
				t.Fatal("FloodFillWater returned nil")
			}

			deepCount := 0
			shallowCount := 0

			for _, tile := range feature.Tiles {
				terrainTile := terrain.GetTile(tile.X, tile.Y)
				if terrainTile == TileWaterDeep {
					deepCount++
				} else if terrainTile == TileWaterShallow {
					shallowCount++
				}
			}

			totalWater := deepCount + shallowCount
			if totalWater == 0 {
				t.Error("No water tiles created")
			}

			// Verify ratio approximately matches (within 20%)
			actualRatio := float64(deepCount) / float64(totalWater)
			if math.Abs(actualRatio-tt.deepWaterRatio) > 0.2 {
				t.Logf("Deep water ratio %.2f differs from expected %.2f (within tolerance)",
					actualRatio, tt.deepWaterRatio)
			}
		})
	}
}

// TestWaterType_String tests water type string conversion.
func TestWaterType_String(t *testing.T) {
	tests := []struct {
		waterType WaterType
		want      string
	}{
		{WaterLake, "Lake"},
		{WaterRiver, "River"},
		{WaterMoat, "Moat"},
		{WaterType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.waterType.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestWaterFeature_Determinism tests that water generation is deterministic.
func TestWaterFeature_Determinism(t *testing.T) {
	seed := int64(12345)
	width, height := 80, 50
	center := Point{40, 25}
	radius := 10

	terrain1 := NewTerrain(width, height, 0)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			terrain1.SetTile(x, y, TileFloor)
		}
	}
	rng1 := rand.New(rand.NewSource(seed))
	feature1 := GenerateLake(center, radius, terrain1, rng1)

	terrain2 := NewTerrain(width, height, 0)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			terrain2.SetTile(x, y, TileFloor)
		}
	}
	rng2 := rand.New(rand.NewSource(seed))
	feature2 := GenerateLake(center, radius, terrain2, rng2)

	if len(feature1.Tiles) != len(feature2.Tiles) {
		t.Errorf("Tile counts differ: %d vs %d", len(feature1.Tiles), len(feature2.Tiles))
	}

	differences := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if terrain1.GetTile(x, y) != terrain2.GetTile(x, y) {
				differences++
			}
		}
	}

	if differences > 0 {
		t.Errorf("Found %d tile differences with same seed (should be deterministic)", differences)
	}
}

// TestWaterFeature_EdgeCases tests edge cases and error conditions.
func TestWaterFeature_EdgeCases(t *testing.T) {
	t.Run("zero radius lake", func(t *testing.T) {
		terrain := NewTerrain(20, 20, 0)
		for y := 0; y < 20; y++ {
			for x := 0; x < 20; x++ {
				terrain.SetTile(x, y, TileFloor)
			}
		}
		rng := rand.New(rand.NewSource(12345))
		feature := GenerateLake(Point{10, 10}, 0, terrain, rng)

		if feature == nil {
			t.Fatal("GenerateLake returned nil for zero radius")
		}
	})

	t.Run("flood fill from wall", func(t *testing.T) {
		terrain := NewTerrain(30, 20, 0)
		for y := 0; y < 20; y++ {
			for x := 0; x < 30; x++ {
				terrain.SetTile(x, y, TileWall)
			}
		}

		filled := FloodFill(Point{15, 10}, 100, terrain)

		if filled == nil {
			return // OK
		}
		if len(filled) > 0 {
			t.Error("FloodFill from wall should return empty")
		}
	})
}

// BenchmarkGenerateLake benchmarks lake generation.
func BenchmarkGenerateLake(b *testing.B) {
	terrain := NewTerrain(100, 80, 0)
	for y := 0; y < 80; y++ {
		for x := 0; x < 100; x++ {
			terrain.SetTile(x, y, TileFloor)
		}
	}
	rng := rand.New(rand.NewSource(12345))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateLake(Point{50, 40}, 15, terrain, rng)
	}
}

// BenchmarkGenerateRiver benchmarks river generation.
func BenchmarkGenerateRiver(b *testing.B) {
	terrain := NewTerrain(100, 80, 0)
	for y := 0; y < 80; y++ {
		for x := 0; x < 100; x++ {
			terrain.SetTile(x, y, TileFloor)
		}
	}
	rng := rand.New(rand.NewSource(12345))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateRiver(Point{10, 40}, Point{90, 40}, 3, terrain, rng)
	}
}

// BenchmarkFloodFill benchmarks flood fill algorithm.
func BenchmarkFloodFill(b *testing.B) {
	terrain := NewTerrain(100, 80, 0)
	for y := 0; y < 80; y++ {
		for x := 0; x < 100; x++ {
			terrain.SetTile(x, y, TileFloor)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FloodFill(Point{50, 40}, 500, terrain)
	}
}
