package terrain

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestCityGenerator_Generate(t *testing.T) {
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
			params:     procgen.GenerationParams{Difficulty: 0.5, Depth: 1, GenreID: "scifi"},
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
				GenreID:    "cyberpunk",
				Custom: map[string]interface{}{
					"width":  100,
					"height": 60,
				},
			},
			wantWidth:  100,
			wantHeight: 60,
			wantErr:    false,
		},
		{
			name: "small block size",
			seed: 11111,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "scifi",
				Custom: map[string]interface{}{
					"blockSize": 8,
				},
			},
			wantWidth:  80,
			wantHeight: 50,
			wantErr:    false,
		},
		{
			name: "large block size",
			seed: 22222,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "scifi",
				Custom: map[string]interface{}{
					"blockSize": 20,
				},
			},
			wantWidth:  80,
			wantHeight: 50,
			wantErr:    false,
		},
		{
			name: "high building density",
			seed: 33333,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "scifi",
				Custom: map[string]interface{}{
					"buildingDensity": 0.9,
				},
			},
			wantWidth:  80,
			wantHeight: 50,
			wantErr:    false,
		},
		{
			name: "low building density",
			seed: 44444,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "scifi",
				Custom: map[string]interface{}{
					"buildingDensity": 0.3,
				},
			},
			wantWidth:  80,
			wantHeight: 50,
			wantErr:    false,
		},
		{
			name: "wide streets",
			seed: 55555,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "scifi",
				Custom: map[string]interface{}{
					"streetWidth": 3,
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
				GenreID:    "scifi",
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
				GenreID:    "scifi",
				Custom: map[string]interface{}{
					"width":  1500,
					"height": 1500,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid block size",
			seed: 99997,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "scifi",
				Custom: map[string]interface{}{
					"blockSize": 50,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewCityGenerator()
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

func TestCityGenerator_Determinism(t *testing.T) {
	gen1 := NewCityGenerator()
	gen2 := NewCityGenerator()

	seed := int64(12345)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "scifi",
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

func TestCityGenerator_GridSubdivision(t *testing.T) {
	gen := NewCityGenerator()

	tests := []struct {
		name      string
		blockSize int
		width     int
		height    int
		minBlocks int
	}{
		{
			name:      "default block size",
			blockSize: 12,
			width:     80,
			height:    50,
			minBlocks: 6, // Should fit several blocks
		},
		{
			name:      "small blocks",
			blockSize: 8,
			width:     80,
			height:    50,
			minBlocks: 15, // More blocks with smaller size
		},
		{
			name:      "large blocks",
			blockSize: 20,
			width:     100,
			height:    80,
			minBlocks: 3, // Fewer blocks with larger size
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "scifi",
				Custom: map[string]interface{}{
					"width":     tt.width,
					"height":    tt.height,
					"blockSize": tt.blockSize,
				},
			}

			result, err := gen.Generate(12345, params)
			if err != nil {
				t.Fatalf("Generate() failed: %v", err)
			}

			terrain := result.(*Terrain)

			// Count blocks by checking distinct building/plaza regions
			// For this test, we'll check that there are multiple distinct regions
			// by scanning for structure/plaza tiles

			// Count transitions from street to non-street
			blockRegions := 0
			for y := 1; y < terrain.Height-1; y++ {
				wasStreet := terrain.GetTile(0, y) == TileCorridor
				for x := 1; x < terrain.Width; x++ {
					isStreet := terrain.GetTile(x, y) == TileCorridor
					if wasStreet && !isStreet {
						blockRegions++
					}
					wasStreet = isStreet
				}
			}

			if blockRegions < tt.minBlocks {
				t.Errorf("Expected at least %d block regions, got %d", tt.minBlocks, blockRegions)
			}

			t.Logf("Detected %d block region boundaries", blockRegions)
		})
	}
}

func TestCityGenerator_BuildingDensity(t *testing.T) {
	gen := NewCityGenerator()

	tests := []struct {
		name            string
		buildingDensity float64
		minBuildingPct  float64
		maxBuildingPct  float64
	}{
		{
			name:            "default density (70%)",
			buildingDensity: 0.7,
			minBuildingPct:  0.10, // Buildings often have interior rooms (floor tiles)
			maxBuildingPct:  0.40, // Streets also take significant space
		},
		{
			name:            "low density (30%)",
			buildingDensity: 0.3,
			minBuildingPct:  0.05,
			maxBuildingPct:  0.25,
		},
		{
			name:            "high density (90%)",
			buildingDensity: 0.9,
			minBuildingPct:  0.15,
			maxBuildingPct:  0.50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "scifi",
				Custom: map[string]interface{}{
					"buildingDensity": tt.buildingDensity,
				},
			}

			result, err := gen.Generate(12345, params)
			if err != nil {
				t.Fatalf("Generate() failed: %v", err)
			}

			terrain := result.(*Terrain)

			// Count building tiles (walls and structures)
			buildingTiles := 0
			totalTiles := terrain.Width * terrain.Height

			for y := 0; y < terrain.Height; y++ {
				for x := 0; x < terrain.Width; x++ {
					tile := terrain.GetTile(x, y)
					if tile == TileStructure || tile == TileWall {
						buildingTiles++
					}
				}
			}

			buildingPct := float64(buildingTiles) / float64(totalTiles)

			if buildingPct < tt.minBuildingPct || buildingPct > tt.maxBuildingPct {
				t.Errorf("Building percentage %.2f%% outside expected range [%.2f%%, %.2f%%]",
					buildingPct*100, tt.minBuildingPct*100, tt.maxBuildingPct*100)
			}

			t.Logf("Building tiles: %d/%d (%.1f%%), density parameter: %.1f%%",
				buildingTiles, totalTiles, buildingPct*100, tt.buildingDensity*100)
		})
	}
}

func TestCityGenerator_StreetConnectivity(t *testing.T) {
	gen := NewCityGenerator()

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "scifi",
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	terrain := result.(*Terrain)

	// Find a street tile to start flood fill
	var startX, startY int
	foundStart := false
	for y := 0; y < terrain.Height && !foundStart; y++ {
		for x := 0; x < terrain.Width && !foundStart; x++ {
			if terrain.GetTile(x, y) == TileCorridor {
				startX, startY = x, y
				foundStart = true
			}
		}
	}

	if !foundStart {
		t.Skip("No street tiles found")
	}

	// Flood fill from street tile
	visited := make([][]bool, terrain.Height)
	for i := range visited {
		visited[i] = make([]bool, terrain.Width)
	}

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

	// Count total walkable tiles
	totalWalkable := 0
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			if terrain.IsWalkable(x, y) {
				totalWalkable++
			}
		}
	}

	// At least 90% of walkable tiles should be reachable (some building interiors might be isolated)
	minReachable := int(float64(totalWalkable) * 0.9)
	if reachableCount < minReachable {
		t.Errorf("Insufficient connectivity: %d/%d walkable tiles reachable (%.1f%%, expected >= 90%%)",
			reachableCount, totalWalkable, float64(reachableCount)/float64(totalWalkable)*100)
	}

	t.Logf("Street connectivity: %d/%d walkable tiles reachable (%.1f%%)",
		reachableCount, totalWalkable, float64(reachableCount)/float64(totalWalkable)*100)
}

func TestCityGenerator_BuildingInteriors(t *testing.T) {
	gen := NewCityGenerator()

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "scifi",
		Custom: map[string]interface{}{
			"blockSize": 15, // Large enough for interior subdivision
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	terrain := result.(*Terrain)

	// Count doors (indicates interior rooms)
	doorCount := 0
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			if terrain.GetTile(x, y) == TileDoor {
				doorCount++
			}
		}
	}

	// Should have some doors for building entrances and interior rooms
	if doorCount < 3 {
		t.Errorf("Expected at least 3 doors, got %d", doorCount)
	}

	t.Logf("Found %d doors in city", doorCount)
}

func TestCityGenerator_Plazas(t *testing.T) {
	gen := NewCityGenerator()

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "scifi",
		Custom: map[string]interface{}{
			"plazaDensity": 0.3, // 30% plazas
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	terrain := result.(*Terrain)

	// Plazas are tracked as "rooms"
	if len(terrain.Rooms) == 0 {
		t.Error("Expected at least one plaza (room)")
	}

	t.Logf("Created %d plazas", len(terrain.Rooms))
}

func TestCityGenerator_Parks(t *testing.T) {
	gen := NewCityGenerator()

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "scifi",
		Custom: map[string]interface{}{
			"buildingDensity": 0.5, // Lower building density
			"plazaDensity":    0.1, // Lower plaza density
			// Remaining 40% should be parks
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	terrain := result.(*Terrain)

	// Count trees (parks have trees)
	treeCount := 0
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			if terrain.GetTile(x, y) == TileTree {
				treeCount++
			}
		}
	}

	if treeCount == 0 {
		t.Log("No trees found (parks may not have generated or have ponds instead)")
	} else {
		t.Logf("Found %d trees in parks", treeCount)
	}
}

func TestCityGenerator_Stairs(t *testing.T) {
	gen := NewCityGenerator()

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "scifi",
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	terrain := result.(*Terrain)

	// Should have stairs placed
	if len(terrain.StairsUp) == 0 {
		t.Error("No stairs up found")
	}

	if len(terrain.StairsDown) == 0 {
		t.Error("No stairs down found")
	}

	// Verify stairs are on walkable tiles
	for _, stairPos := range terrain.StairsUp {
		if !terrain.IsWalkable(stairPos.X, stairPos.Y) {
			t.Errorf("Stairs up at (%d,%d) not on walkable tile", stairPos.X, stairPos.Y)
		}
	}

	for _, stairPos := range terrain.StairsDown {
		if !terrain.IsWalkable(stairPos.X, stairPos.Y) {
			t.Errorf("Stairs down at (%d,%d) not on walkable tile", stairPos.X, stairPos.Y)
		}
	}

	t.Logf("Stairs: %d up, %d down", len(terrain.StairsUp), len(terrain.StairsDown))
}

func TestCityGenerator_Validate(t *testing.T) {
	gen := NewCityGenerator()

	tests := []struct {
		name    string
		seed    int64
		params  procgen.GenerationParams
		wantErr bool
	}{
		{
			name:    "valid city",
			seed:    12345,
			params:  procgen.GenerationParams{Difficulty: 0.5, Depth: 1, GenreID: "scifi"},
			wantErr: false,
		},
		{
			name: "high building density still valid",
			seed: 67890,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "scifi",
				Custom: map[string]interface{}{
					"buildingDensity": 0.9,
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

func TestCityGenerator_Validate_Invalid(t *testing.T) {
	gen := NewCityGenerator()

	// Test invalid input type
	err := gen.Validate("not a terrain")
	if err == nil {
		t.Error("Validate() should fail for non-Terrain input")
	}

	// Test terrain with insufficient walkable tiles
	terrain := NewTerrain(10, 10, 12345)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			terrain.SetTile(x, y, TileStructure) // All buildings
		}
	}
	err = gen.Validate(terrain)
	if err == nil {
		t.Error("Validate() should fail for terrain with insufficient walkable tiles")
	}
}

func BenchmarkCityGenerator_Small(b *testing.B) {
	gen := NewCityGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "scifi",
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

func BenchmarkCityGenerator_Medium(b *testing.B) {
	gen := NewCityGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "scifi",
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

func BenchmarkCityGenerator_Large(b *testing.B) {
	gen := NewCityGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "scifi",
		Custom: map[string]interface{}{
			"width":  200,
			"height": 200,
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
