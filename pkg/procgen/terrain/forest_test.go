package terrain

import (
	"math"
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestForestGenerator_Generate(t *testing.T) {
	tests := []struct {
		name       string
		seed       int64
		params     procgen.GenerationParams
		wantWidth  int
		wantHeight int
		wantErr    bool
	}{
		{
			name:       "default parameters",
			seed:       12345,
			params:     procgen.GenerationParams{Difficulty: 0.5, Depth: 1, GenreID: "fantasy"},
			wantWidth:  80,
			wantHeight: 50,
			wantErr:    false,
		},
		{
			name: "custom dimensions",
			seed: 67890,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"width":  60,
					"height": 40,
				},
			},
			wantWidth:  60,
			wantHeight: 40,
			wantErr:    false,
		},
		{
			name: "high tree density",
			seed: 11111,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"treeDensity": 0.5,
				},
			},
			wantWidth:  80,
			wantHeight: 50,
			wantErr:    false,
		},
		{
			name: "low tree density",
			seed: 22222,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"treeDensity": 0.1,
				},
			},
			wantWidth:  80,
			wantHeight: 50,
			wantErr:    false,
		},
		{
			name: "many clearings",
			seed: 33333,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"clearingCount": 5,
				},
			},
			wantWidth:  80,
			wantHeight: 50,
			wantErr:    false,
		},
		{
			name: "guaranteed water features",
			seed: 44444,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"waterChance": 1.0,
				},
			},
			wantWidth:  80,
			wantHeight: 50,
			wantErr:    false,
		},
		{
			name: "zero width",
			seed: 99999,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"width": 0,
				},
			},
			wantErr: true,
		},
		{
			name: "dimensions too large",
			seed: 99998,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"width":  1500,
					"height": 1500,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewForestGenerator()
			result, err := gen.Generate(tt.seed, tt.params)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Generate() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Generate() unexpected error: %v", err)
			}

			terrain, ok := result.(*Terrain)
			if !ok {
				t.Fatalf("Generate() did not return *Terrain")
			}

			if terrain.Width != tt.wantWidth {
				t.Errorf("Width = %d, want %d", terrain.Width, tt.wantWidth)
			}

			if terrain.Height != tt.wantHeight {
				t.Errorf("Height = %d, want %d", terrain.Height, tt.wantHeight)
			}

			// Validate the terrain
			if err := gen.Validate(terrain); err != nil {
				t.Errorf("Validate() failed: %v", err)
			}
		})
	}
}

func TestForestGenerator_Determinism(t *testing.T) {
	gen1 := NewForestGenerator()
	gen2 := NewForestGenerator()

	seed := int64(12345)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
	}

	result1, err1 := gen1.Generate(seed, params)
	if err1 != nil {
		t.Fatalf("First Generate() failed: %v", err1)
	}

	result2, err2 := gen2.Generate(seed, params)
	if err2 != nil {
		t.Fatalf("Second Generate() failed: %v", err2)
	}

	terrain1 := result1.(*Terrain)
	terrain2 := result2.(*Terrain)

	// Compare all tiles
	if terrain1.Width != terrain2.Width || terrain1.Height != terrain2.Height {
		t.Errorf("Dimensions differ: (%d,%d) vs (%d,%d)",
			terrain1.Width, terrain1.Height, terrain2.Width, terrain2.Height)
	}

	differences := 0
	for y := 0; y < terrain1.Height; y++ {
		for x := 0; x < terrain1.Width; x++ {
			if terrain1.GetTile(x, y) != terrain2.GetTile(x, y) {
				differences++
			}
		}
	}

	if differences > 0 {
		t.Errorf("Found %d tile differences between two generations with same seed", differences)
	}
}

func TestForestGenerator_TreeDistribution(t *testing.T) {
	gen := NewForestGenerator()

	tests := []struct {
		name        string
		treeDensity float64
		minTreePct  float64 // Minimum tree percentage
		maxTreePct  float64 // Maximum tree percentage
	}{
		{
			name:        "default density (30%)",
			treeDensity: 0.3,
			minTreePct:  0.01, // At least 1% trees (clearings reduce this significantly)
			maxTreePct:  0.25, // At most 25% (some space taken by clearings/paths)
		},
		{
			name:        "low density (10%)",
			treeDensity: 0.1,
			minTreePct:  0.005, // At least 0.5% trees
			maxTreePct:  0.15,  // At most 15%
		},
		{
			name:        "high density (50%)",
			treeDensity: 0.5,
			minTreePct:  0.02, // At least 2% trees
			maxTreePct:  0.40, // At most 40%
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"treeDensity": tt.treeDensity,
					"waterChance": 0.0, // Disable water for this test
				},
			}

			result, err := gen.Generate(12345, params)
			if err != nil {
				t.Fatalf("Generate() failed: %v", err)
			}

			terrain := result.(*Terrain)

			// Count trees
			treeCount := 0
			totalTiles := terrain.Width * terrain.Height
			for y := 0; y < terrain.Height; y++ {
				for x := 0; x < terrain.Width; x++ {
					if terrain.GetTile(x, y) == TileTree {
						treeCount++
					}
				}
			}

			treePct := float64(treeCount) / float64(totalTiles)

			// Check if tree percentage is within expected range
			if treePct < tt.minTreePct || treePct > tt.maxTreePct {
				t.Errorf("Tree percentage %.2f%% outside expected range [%.2f%%, %.2f%%]",
					treePct*100, tt.minTreePct*100, tt.maxTreePct*100)
			}

			t.Logf("Tree count: %d/%d (%.1f%%), density parameter: %.1f%%",
				treeCount, totalTiles, treePct*100, tt.treeDensity*100)
		})
	}
}

func TestForestGenerator_Clearings(t *testing.T) {
	gen := NewForestGenerator()

	tests := []struct {
		name          string
		clearingCount int
		minClearings  int
	}{
		{
			name:          "default clearings (3)",
			clearingCount: 3,
			minClearings:  2, // Allow at least 2 due to placement constraints
		},
		{
			name:          "many clearings (5)",
			clearingCount: 5,
			minClearings:  3,
		},
		{
			name:          "single clearing (1)",
			clearingCount: 1,
			minClearings:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"clearingCount": tt.clearingCount,
				},
			}

			result, err := gen.Generate(12345, params)
			if err != nil {
				t.Fatalf("Generate() failed: %v", err)
			}

			terrain := result.(*Terrain)

			if len(terrain.Rooms) < tt.minClearings {
				t.Errorf("Got %d clearings, want at least %d", len(terrain.Rooms), tt.minClearings)
			}

			// Verify clearings don't overlap
			for i := 0; i < len(terrain.Rooms); i++ {
				for j := i + 1; j < len(terrain.Rooms); j++ {
					if terrain.Rooms[i].Overlaps(terrain.Rooms[j]) {
						t.Errorf("Clearings %d and %d overlap", i, j)
					}
				}
			}

			t.Logf("Created %d clearings", len(terrain.Rooms))
		})
	}
}

func TestForestGenerator_WaterFeatures(t *testing.T) {
	gen := NewForestGenerator()

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"waterChance": 1.0, // Guarantee water features
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	terrain := result.(*Terrain)

	// Count water tiles
	shallowWater := 0
	deepWater := 0
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			tile := terrain.GetTile(x, y)
			if tile == TileWaterShallow {
				shallowWater++
			} else if tile == TileWaterDeep {
				deepWater++
			}
		}
	}

	totalWater := shallowWater + deepWater

	if totalWater == 0 {
		t.Error("Expected water features but found none")
	}

	t.Logf("Water tiles: %d total (%d shallow, %d deep)", totalWater, shallowWater, deepWater)

	// Verify water features are not too large (should be < 20% of map)
	totalTiles := terrain.Width * terrain.Height
	waterPct := float64(totalWater) / float64(totalTiles)
	if waterPct > 0.3 {
		t.Errorf("Water coverage %.1f%% exceeds 30%% of map", waterPct*100)
	}
}

func TestForestGenerator_Bridges(t *testing.T) {
	gen := NewForestGenerator()

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"waterChance": 1.0, // Guarantee water features
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	terrain := result.(*Terrain)

	// Count bridges
	bridgeCount := 0
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			if terrain.GetTile(x, y) == TileBridge {
				bridgeCount++
			}
		}
	}

	// Bridges should exist if there are water features and paths
	t.Logf("Found %d bridges", bridgeCount)

	// Verify each bridge has floor tiles adjacent
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			if terrain.GetTile(x, y) == TileBridge {
				hasAdjacentFloor := false
				for _, neighbor := range (Point{X: x, Y: y}).Neighbors() {
					if terrain.IsInBounds(neighbor.X, neighbor.Y) {
						tile := terrain.GetTile(neighbor.X, neighbor.Y)
						if tile == TileFloor || tile == TileBridge {
							hasAdjacentFloor = true
							break
						}
					}
				}

				if !hasAdjacentFloor {
					t.Errorf("Bridge at (%d,%d) has no adjacent floor tiles", x, y)
				}
			}
		}
	}
}

func TestForestGenerator_Stairs(t *testing.T) {
	gen := NewForestGenerator()

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	terrain := result.(*Terrain)

	// Should have stairs placed in clearings
	if len(terrain.StairsUp) == 0 {
		t.Error("No stairs up found")
	}

	if len(terrain.StairsDown) == 0 {
		t.Error("No stairs down found")
	}

	// Verify stairs are in clearings
	for _, stairPos := range terrain.StairsUp {
		inClearing := false
		for _, clearing := range terrain.Rooms {
			if stairPos.X >= clearing.X && stairPos.X < clearing.X+clearing.Width &&
				stairPos.Y >= clearing.Y && stairPos.Y < clearing.Y+clearing.Height {
				inClearing = true
				break
			}
		}
		if !inClearing {
			t.Errorf("Stairs up at (%d,%d) not in a clearing", stairPos.X, stairPos.Y)
		}
	}

	t.Logf("Stairs: %d up, %d down", len(terrain.StairsUp), len(terrain.StairsDown))
}

func TestForestGenerator_Connectivity(t *testing.T) {
	gen := NewForestGenerator()

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"waterChance": 0.0, // Disable water for simpler connectivity test
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	terrain := result.(*Terrain)

	if len(terrain.Rooms) < 2 {
		t.Skip("Need at least 2 clearings for connectivity test")
	}

	// Flood fill from first clearing
	visited := make([][]bool, terrain.Height)
	for i := range visited {
		visited[i] = make([]bool, terrain.Width)
	}

	// Start flood fill from center of first clearing
	startX, startY := terrain.Rooms[0].Center()
	queue := []Point{{X: startX, Y: startY}}
	visited[startY][startX] = true
	reachableCount := 1

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Check all neighbors
		for _, neighbor := range current.Neighbors() {
			if !terrain.IsInBounds(neighbor.X, neighbor.Y) {
				continue
			}

			if visited[neighbor.Y][neighbor.X] {
				continue
			}

			if terrain.IsWalkable(neighbor.X, neighbor.Y) {
				visited[neighbor.Y][neighbor.X] = true
				queue = append(queue, neighbor)
				reachableCount++
			}
		}
	}

	// Check if all clearings are reachable
	for i, clearing := range terrain.Rooms {
		cx, cy := clearing.Center()
		if !visited[cy][cx] {
			t.Errorf("Clearing %d at (%d,%d) is not reachable from clearing 0", i, cx, cy)
		}
	}

	t.Logf("Flood fill reached %d walkable tiles", reachableCount)
}

func TestForestGenerator_PoissonDiscSampling(t *testing.T) {
	gen := NewForestGenerator()
	rng := rand.New(rand.NewSource(12345))

	width, height := 100, 100
	minDist := 5.0

	points := gen.poissonDiscSampling(width, height, minDist, rng)

	if len(points) == 0 {
		t.Fatal("Poisson disc sampling generated no points")
	}

	// Verify minimum distance constraint
	for i := 0; i < len(points); i++ {
		for j := i + 1; j < len(points); j++ {
			dist := points[i].Distance(points[j])
			if dist < minDist {
				t.Errorf("Points %d and %d are too close: distance %.2f < %.2f",
					i, j, dist, minDist)
			}
		}
	}

	// Verify points are within bounds
	for i, point := range points {
		if point.X < 0 || point.X >= width || point.Y < 0 || point.Y >= height {
			t.Errorf("Point %d at (%d,%d) is out of bounds [0,%d)x[0,%d)",
				i, point.X, point.Y, width, height)
		}
	}

	t.Logf("Generated %d points with min distance %.1f", len(points), minDist)

	// Verify density is reasonable (should fill space efficiently)
	expectedPoints := int(float64(width*height) / (math.Pi * minDist * minDist / 4))
	if len(points) < expectedPoints/3 {
		t.Errorf("Point density too low: got %d points, expected at least %d",
			len(points), expectedPoints/3)
	}
}

func TestForestGenerator_Validate(t *testing.T) {
	gen := NewForestGenerator()

	tests := []struct {
		name    string
		seed    int64
		params  procgen.GenerationParams
		wantErr bool
	}{
		{
			name:    "valid forest",
			seed:    12345,
			params:  procgen.GenerationParams{Difficulty: 0.5, Depth: 1, GenreID: "fantasy"},
			wantErr: false,
		},
		{
			name: "high tree density still valid",
			seed: 67890,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"treeDensity": 0.5,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.Generate(tt.seed, tt.params)
			if err != nil {
				t.Fatalf("Generate() failed: %v", err)
			}

			err = gen.Validate(result)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestForestGenerator_Validate_Invalid(t *testing.T) {
	gen := NewForestGenerator()

	// Test invalid input type
	err := gen.Validate("not a terrain")
	if err == nil {
		t.Error("Validate() should fail for non-Terrain input")
	}

	// Test terrain with insufficient walkable tiles
	terrain := NewTerrain(10, 10, 12345)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			terrain.SetTile(x, y, TileWall) // All walls
		}
	}
	err = gen.Validate(terrain)
	if err == nil {
		t.Error("Validate() should fail for terrain with insufficient walkable tiles")
	}

	// Test terrain with no clearings
	terrain = NewTerrain(10, 10, 12345)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			terrain.SetTile(x, y, TileFloor)
		}
	}
	terrain.Rooms = nil
	err = gen.Validate(terrain)
	if err == nil {
		t.Error("Validate() should fail for terrain with no clearings")
	}
}

func BenchmarkForestGenerator_Small(b *testing.B) {
	gen := NewForestGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  40,
			"height": 30,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(12345, params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkForestGenerator_Medium(b *testing.B) {
	gen := NewForestGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
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
			b.Fatal(err)
		}
	}
}

func BenchmarkForestGenerator_Large(b *testing.B) {
	gen := NewForestGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  150,
			"height": 100,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(12345, params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPoissonDiscSampling(b *testing.B) {
	gen := NewForestGenerator()
	rng := rand.New(rand.NewSource(12345))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.poissonDiscSampling(100, 100, 5.0, rng)
	}
}
