package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

func TestNewTerrainStealthSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStealthSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewTerrainStealthSystem returned nil")
	}
	if sys.tileSize != 32 {
		t.Errorf("tileSize = %d, want 32", sys.tileSize)
	}
	if sys.updateInterval != 0.25 {
		t.Errorf("updateInterval = %f, want 0.25", sys.updateInterval)
	}
}

func TestTerrainStealthSystem_SetTerrain(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStealthSystem(world, 12345)

	// Create test terrain
	terr := terrain.NewTerrain(10, 10, 12345)
	sys.SetTerrain(terr)

	if sys.terrain != terr {
		t.Error("SetTerrain did not set terrain")
	}
}

func TestTerrainStealthSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStealthSystem(world, 12345)

	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("genreID = %s, want horror", sys.genreID)
	}
}

func TestTerrainStealthSystem_SetTileSize(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStealthSystem(world, 12345)

	sys.SetTileSize(64)
	if sys.tileSize != 64 {
		t.Errorf("tileSize = %d, want 64", sys.tileSize)
	}

	// Invalid size should not change
	sys.SetTileSize(0)
	if sys.tileSize != 64 {
		t.Errorf("tileSize changed to %d on invalid input", sys.tileSize)
	}
}

func TestTerrainStealthSystem_CalculateStealthMultiplier(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStealthSystem(world, 12345)

	// Create test terrain with various tile types
	terr := terrain.NewTerrain(10, 10, 12345)

	// Set up terrain: mostly floor with specific tiles
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			terr.SetTile(x, y, terrain.TileFloor)
		}
	}
	terr.SetTile(5, 5, terrain.TileWaterShallow) // Splashing
	terr.SetTile(3, 3, terrain.TileCorridor)     // Echoing
	terr.SetTile(7, 7, terrain.TileBridge)       // Elevated

	sys.SetTerrain(terr)

	tests := []struct {
		name      string
		tileX     int
		tileY     int
		wantLower float64
		wantUpper float64
	}{
		{"floor_normal", 1, 1, 0.9, 1.1},
		{"water_louder", 5, 5, 1.2, 1.4},
		{"corridor_echoing", 3, 3, 1.0, 1.2},
		{"bridge_elevated", 7, 7, 0.8, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mult := sys.calculateStealthMultiplier(tt.tileX, tt.tileY)
			if mult < tt.wantLower || mult > tt.wantUpper {
				t.Errorf("calculateStealthMultiplier(%d,%d) = %f, want [%f, %f]",
					tt.tileX, tt.tileY, mult, tt.wantLower, tt.wantUpper)
			}
		})
	}
}

func TestTerrainStealthSystem_CoverBonus(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStealthSystem(world, 12345)

	// Create terrain with trees for cover
	terr := terrain.NewTerrain(10, 10, 12345)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			terr.SetTile(x, y, terrain.TileFloor)
		}
	}

	// Add trees around position (5,5)
	terr.SetTile(4, 4, terrain.TileTree)
	terr.SetTile(4, 5, terrain.TileTree)
	terr.SetTile(5, 4, terrain.TileTree)

	sys.SetTerrain(terr)

	// Position with cover should have lower stealth multiplier
	withCover := sys.calculateStealthMultiplier(5, 5)
	// Position without cover
	noCover := sys.calculateStealthMultiplier(8, 8)

	if withCover >= noCover {
		t.Errorf("Cover should reduce detection: withCover=%f, noCover=%f", withCover, noCover)
	}
}

func TestTerrainStealthSystem_SecretDoorConcealment(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStealthSystem(world, 12345)

	terr := terrain.NewTerrain(10, 10, 12345)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			terr.SetTile(x, y, terrain.TileFloor)
		}
	}

	// Add secret door adjacent to (5,5)
	terr.SetTile(4, 5, terrain.TileSecretDoor)

	sys.SetTerrain(terr)

	// Position near secret door should have good concealment
	nearSecret := sys.calculateStealthMultiplier(5, 5)
	farFromSecret := sys.calculateStealthMultiplier(8, 8)

	if nearSecret >= farFromSecret {
		t.Errorf("Secret door should provide concealment: near=%f, far=%f", nearSecret, farFromSecret)
	}
}

func TestTerrainStealthSystem_GenreModifiers(t *testing.T) {
	tests := []struct {
		name       string
		genre      string
		setupTiles func(*terrain.Terrain)
		testX      int
		testY      int
		expectLess bool // Whether this genre should make stealth better
	}{
		{
			name:  "fantasy_trees",
			genre: "fantasy",
			setupTiles: func(terr *terrain.Terrain) {
				terr.SetTile(4, 5, terrain.TileTree)
			},
			testX:      5,
			testY:      5,
			expectLess: true,
		},
		{
			name:  "scifi_corridor",
			genre: "scifi",
			setupTiles: func(terr *terrain.Terrain) {
				terr.SetTile(5, 5, terrain.TileCorridor)
			},
			testX:      5,
			testY:      5,
			expectLess: false, // Security sensors make it worse
		},
		{
			name:  "horror_everywhere",
			genre: "horror",
			setupTiles: func(terr *terrain.Terrain) {
				// No special setup
			},
			testX:      5,
			testY:      5,
			expectLess: false, // Horror makes stealth harder
		},
		{
			name:  "cyberpunk_walls",
			genre: "cyberpunk",
			setupTiles: func(terr *terrain.Terrain) {
				terr.SetTile(4, 5, terrain.TileWall)
				terr.SetTile(6, 5, terrain.TileWall)
			},
			testX:      5,
			testY:      5,
			expectLess: true, // Camera blind spots
		},
		{
			name:  "postapoc_debris",
			genre: "postapoc",
			setupTiles: func(terr *terrain.Terrain) {
				// No special setup - debris everywhere
			},
			testX:      5,
			testY:      5,
			expectLess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()

			// Calculate baseline with no genre
			sysBase := NewTerrainStealthSystem(world, 12345)
			terrBase := terrain.NewTerrain(10, 10, 12345)
			for y := 0; y < 10; y++ {
				for x := 0; x < 10; x++ {
					terrBase.SetTile(x, y, terrain.TileFloor)
				}
			}
			tt.setupTiles(terrBase)
			sysBase.SetTerrain(terrBase)
			baseline := sysBase.calculateStealthMultiplier(tt.testX, tt.testY)

			// Calculate with genre
			sysGenre := NewTerrainStealthSystem(world, 12345)
			terrGenre := terrain.NewTerrain(10, 10, 12345)
			for y := 0; y < 10; y++ {
				for x := 0; x < 10; x++ {
					terrGenre.SetTile(x, y, terrain.TileFloor)
				}
			}
			tt.setupTiles(terrGenre)
			sysGenre.SetTerrain(terrGenre)
			sysGenre.SetGenre(tt.genre)
			withGenre := sysGenre.calculateStealthMultiplier(tt.testX, tt.testY)

			if tt.expectLess && withGenre >= baseline {
				t.Errorf("Genre %s should improve stealth: baseline=%f, withGenre=%f",
					tt.genre, baseline, withGenre)
			}
			if !tt.expectLess && withGenre <= baseline {
				t.Errorf("Genre %s should worsen stealth: baseline=%f, withGenre=%f",
					tt.genre, baseline, withGenre)
			}
		})
	}
}

func TestTerrainStealthSystem_Update(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStealthSystem(world, 12345)

	terr := terrain.NewTerrain(10, 10, 12345)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			terr.SetTile(x, y, terrain.TileFloor)
		}
	}
	terr.SetTile(3, 3, terrain.TileTree) // Cover nearby
	sys.SetTerrain(terr)

	// Create test entity
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100}) // Near (3,3) in tiles

	entities := []*Entity{entity}

	// First update should trigger calculation (after interval)
	sys.timeSinceCheck = sys.updateInterval
	sys.Update(entities, 0.1)

	// Check that stealth was calculated
	mult := sys.GetStealthMultiplier(entity.ID)
	if mult == 0 {
		t.Error("Stealth multiplier should be calculated after update")
	}
}

func TestTerrainStealthSystem_AIDetectionModification(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStealthSystem(world, 12345)

	terr := terrain.NewTerrain(20, 20, 12345)
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			terr.SetTile(x, y, terrain.TileFloor)
		}
	}
	// Add trees around player position for cover
	terr.SetTile(3, 3, terrain.TileTree)
	terr.SetTile(3, 4, terrain.TileTree)
	sys.SetTerrain(terr)

	// Create AI entity
	aiEntity := world.CreateEntity()
	aiEntity.AddComponent(&PositionComponent{X: 320, Y: 320}) // (10, 10) in tiles
	aiEntity.AddComponent(&AIComponent{DetectionRange: 200})

	// Create target entity with cover
	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 128, Y: 128}) // Near (4, 4) - has tree cover
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})

	entities := []*Entity{aiEntity, target}

	// Force update
	sys.timeSinceCheck = sys.updateInterval
	sys.Update(entities, 0.1)

	// Get modified detection range
	aiComp, _ := aiEntity.GetComponent("ai")
	ai := aiComp.(*AIComponent)

	// AI detection should be reduced due to target's cover
	if ai.DetectionRange >= 200 {
		t.Errorf("AI detection range should be reduced from cover: got %f", ai.DetectionRange)
	}
}

func TestTerrainStealthSystem_GetStealthMultiplier(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStealthSystem(world, 12345)

	// Non-cached entity should return 1.0
	mult := sys.GetStealthMultiplier(999)
	if mult != 1.0 {
		t.Errorf("Uncached entity stealth = %f, want 1.0", mult)
	}

	// Add to cache
	sys.stealthCache[123] = 0.75
	mult = sys.GetStealthMultiplier(123)
	if mult != 0.75 {
		t.Errorf("Cached entity stealth = %f, want 0.75", mult)
	}
}

func TestTerrainStealthSystem_SetUpdateInterval(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStealthSystem(world, 12345)

	sys.SetUpdateInterval(0.5)
	if sys.updateInterval != 0.5 {
		t.Errorf("updateInterval = %f, want 0.5", sys.updateInterval)
	}

	// Invalid interval should not change
	sys.SetUpdateInterval(0)
	if sys.updateInterval != 0.5 {
		t.Errorf("updateInterval changed on invalid input")
	}
}

func TestTerrainStealthSystem_NilTerrain(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStealthSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Should not panic with nil terrain
	sys.Update([]*Entity{entity}, 0.1)
}

func TestTerrainStealthSystem_MultiplierClamping(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStealthSystem(world, 12345)

	// Test that multipliers are clamped
	terr := terrain.NewTerrain(10, 10, 12345)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			terr.SetTile(x, y, terrain.TileFloor)
		}
	}

	// Maximum possible cover setup
	terr.SetTile(4, 4, terrain.TileTree)
	terr.SetTile(4, 5, terrain.TileTree)
	terr.SetTile(4, 6, terrain.TileTree)
	terr.SetTile(5, 4, terrain.TileTree)
	terr.SetTile(6, 4, terrain.TileTree)
	terr.SetTile(6, 5, terrain.TileTree)
	terr.SetTile(6, 6, terrain.TileTree)
	terr.SetTile(5, 6, terrain.TileTree)
	terr.SetTile(3, 5, terrain.TileSecretDoor)

	sys.SetTerrain(terr)
	sys.SetGenre("postapoc") // Additional bonus

	mult := sys.calculateStealthMultiplier(5, 5)

	// Should be clamped to minimum 0.3
	if mult < 0.3 {
		t.Errorf("Multiplier should be clamped to 0.3 minimum, got %f", mult)
	}
}
