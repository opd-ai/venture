package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

func TestTerrainRangedAccuracyComponent_Type(t *testing.T) {
	comp := &TerrainRangedAccuracyComponent{}
	if comp.Type() != "terrain_ranged_accuracy" {
		t.Errorf("expected 'terrain_ranged_accuracy', got %q", comp.Type())
	}
}

func TestNewTerrainRangedAccuracySystem(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainRangedAccuracySystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.tileSize != 32 {
		t.Errorf("expected tileSize 32, got %d", sys.tileSize)
	}
	if sys.terrain != nil {
		t.Error("expected nil terrain before SetTerrain")
	}
}

func TestTerrainRangedAccuracySystem_SetTileSize(t *testing.T) {
	sys := NewTerrainRangedAccuracySystem(nil, 1)
	sys.SetTileSize(64)
	if sys.tileSize != 64 {
		t.Errorf("expected 64, got %d", sys.tileSize)
	}
	sys.SetTileSize(0) // Should not change
	if sys.tileSize != 64 {
		t.Errorf("expected 64 after invalid set, got %d", sys.tileSize)
	}
}

func TestTerrainRangedAccuracySystem_SetGenre(t *testing.T) {
	sys := NewTerrainRangedAccuracySystem(nil, 1)
	sys.SetGenre("scifi")
	if sys.genreID != "scifi" {
		t.Errorf("expected 'scifi', got %q", sys.genreID)
	}
}

func TestTerrainRangedAccuracySystem_UpdateNoTerrain(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainRangedAccuracySystem(world, 42)
	entity := NewEntity(1)
	// Should not panic with nil terrain
	sys.Update([]*Entity{entity}, 0.016)
}

func TestTerrainRangedAccuracySystem_CalculateModifier(t *testing.T) {
	tests := []struct {
		name       string
		tileType   terrain.TileType
		genre      string
		wantMin    float64
		wantMax    float64
		wantNil    bool
	}{
		{"corridor_bonus", terrain.TileCorridor, "fantasy", 1.10, 1.20, false},
		{"platform_bonus", terrain.TilePlatform, "fantasy", 1.05, 1.15, false},
		{"bridge_bonus", terrain.TileBridge, "fantasy", 1.03, 1.12, false},
		{"water_penalty", terrain.TileWaterShallow, "fantasy", 0.85, 0.95, false},
		{"lava_penalty", terrain.TileLavaFlow, "fantasy", 0.80, 0.90, false},
		{"floor_neutral", terrain.TileFloor, "fantasy", 0.99, 1.01, true},
		{"corridor_horror", terrain.TileCorridor, "horror", 1.00, 1.10, false},
		{"water_scifi", terrain.TileWaterShallow, "scifi", 0.95, 1.05, true},
		{"platform_scifi", terrain.TilePlatform, "scifi", 1.10, 1.20, false},
		{"corridor_postapoc", terrain.TileCorridor, "postapoc", 1.05, 1.15, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewTerrainRangedAccuracySystem(nil, 42)
			sys.SetGenre(tt.genre)

			// Create a small terrain with the target tile surrounded by floor
			terr := terrain.NewTerrain(5, 5, 42)
			for x := 0; x < 5; x++ {
				for y := 0; y < 5; y++ {
					terr.SetTile(x, y, terrain.TileFloor)
				}
			}
			terr.SetTile(2, 2, tt.tileType)
			sys.SetTerrain(terr)

			comp := sys.calculateModifier(2, 2)

			if tt.wantNil {
				if comp != nil {
					t.Errorf("expected nil component, got modifier=%f", comp.AccuracyModifier)
				}
				return
			}

			if comp == nil {
				t.Fatal("expected non-nil component")
			}
			if comp.AccuracyModifier < tt.wantMin || comp.AccuracyModifier > tt.wantMax {
				t.Errorf("modifier %f outside range [%f, %f]", comp.AccuracyModifier, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestTerrainRangedAccuracySystem_AdjacentTrees(t *testing.T) {
	sys := NewTerrainRangedAccuracySystem(nil, 42)
	sys.SetGenre("fantasy")

	terr := terrain.NewTerrain(5, 5, 42)
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			terr.SetTile(x, y, terrain.TileFloor)
		}
	}
	// Place trees around center
	terr.SetTile(1, 2, terrain.TileTree)
	terr.SetTile(3, 2, terrain.TileTree)
	terr.SetTile(2, 1, terrain.TileTree)
	sys.SetTerrain(terr)

	count := sys.countAdjacentTrees(2, 2)
	if count != 3 {
		t.Errorf("expected 3 adjacent trees, got %d", count)
	}

	comp := sys.calculateModifier(2, 2)
	if comp == nil {
		t.Fatal("expected modifier with adjacent trees")
	}
	if comp.AccuracyModifier >= 1.0 {
		t.Errorf("expected accuracy penalty from trees, got %f", comp.AccuracyModifier)
	}
	if comp.AdjacentTreeCount != 3 {
		t.Errorf("expected 3 adjacent trees in component, got %d", comp.AdjacentTreeCount)
	}
}

func TestTerrainRangedAccuracySystem_GetAccuracyModifier(t *testing.T) {
	sys := NewTerrainRangedAccuracySystem(nil, 42)

	// Default should return 1.0
	if mod := sys.GetAccuracyModifier(999); mod != 1.0 {
		t.Errorf("expected 1.0 default, got %f", mod)
	}

	// Set a cached value
	sys.accuracyCache[42] = 0.85
	if mod := sys.GetAccuracyModifier(42); mod != 0.85 {
		t.Errorf("expected 0.85, got %f", mod)
	}
}

func TestTerrainRangedAccuracySystem_GenreModifiers(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewTerrainRangedAccuracySystem(nil, 42)
			sys.SetGenre(genre)

			// Test modifier application doesn't panic
			mod := sys.applyGenreModifier(0.80, terrain.TileWaterShallow, 2)
			if mod < 0.5 || mod > 1.25 {
				t.Errorf("genre %s: modifier %f out of expected range", genre, mod)
			}
		})
	}
}

func TestTerrainRangedAccuracySystem_CyberpunkCompensation(t *testing.T) {
	sys := NewTerrainRangedAccuracySystem(nil, 42)
	sys.SetGenre("cyberpunk")

	// Cyberpunk should recover 25% of penalty
	mod := sys.applyGenreModifier(0.80, terrain.TileFloor, 0)
	// 0.80 penalty = 0.20 deficit, recover 25% = 0.05, so 0.85
	if mod < 0.84 || mod > 0.86 {
		t.Errorf("expected ~0.85 cyberpunk compensation, got %f", mod)
	}
}

func TestTerrainRangedAccuracySystem_UpdateWithEntities(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainRangedAccuracySystem(world, 42)

	terr := terrain.NewTerrain(10, 10, 42)
	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			terr.SetTile(x, y, terrain.TileFloor)
		}
	}
	terr.SetTile(3, 3, terrain.TileCorridor)
	sys.SetTerrain(terr)
	sys.SetTileSize(32)

	// Entity with ranged capability (has mana)
	ranged := NewEntity(1)
	ranged.AddComponent(&PositionComponent{X: 96, Y: 96}) // tile 3,3
	ranged.AddComponent(&ManaComponent{Current: 50, Max: 100})

	// Entity without ranged capability
	melee := NewEntity(2)
	melee.AddComponent(&PositionComponent{X: 96, Y: 96})
	melee.AddComponent(&HealthComponent{Current: 100, Max: 100})

	sys.Update([]*Entity{ranged, melee}, 0.016)

	// Ranged entity should have accuracy component
	if _, ok := ranged.GetComponent("terrain_ranged_accuracy"); !ok {
		t.Error("expected ranged entity to have terrain_ranged_accuracy component")
	}

	// Melee entity should NOT have accuracy component
	if _, ok := melee.GetComponent("terrain_ranged_accuracy"); ok {
		t.Error("melee entity should not have terrain_ranged_accuracy component")
	}

	// Check cached modifier
	mod := sys.GetAccuracyModifier(1)
	if mod <= 1.0 {
		t.Errorf("expected corridor bonus > 1.0, got %f", mod)
	}
}

func TestTerrainRangedAccuracySystem_TileCacheInvalidation(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainRangedAccuracySystem(world, 42)

	terr := terrain.NewTerrain(10, 10, 42)
	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			terr.SetTile(x, y, terrain.TileFloor)
		}
	}
	terr.SetTile(3, 3, terrain.TileCorridor)
	terr.SetTile(4, 4, terrain.TileWaterShallow)
	sys.SetTerrain(terr)
	sys.SetTileSize(32)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 96, Y: 96}) // tile 3,3
	entity.AddComponent(&ManaComponent{Current: 50, Max: 100})

	sys.Update([]*Entity{entity}, 0.016)
	mod1 := sys.GetAccuracyModifier(1)

	// Move entity to water tile
	pos := entity.GetPosition()
	pos.X = 128.0
	pos.Y = 128.0

	sys.Update([]*Entity{entity}, 0.016)
	mod2 := sys.GetAccuracyModifier(1)

	if mod1 == mod2 {
		t.Errorf("expected different modifiers on different tiles, both were %f", mod1)
	}
	if mod2 >= 1.0 {
		t.Errorf("expected water penalty < 1.0, got %f", mod2)
	}
}

func BenchmarkTerrainRangedAccuracySystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewTerrainRangedAccuracySystem(world, 42)

	terr := terrain.NewTerrain(32, 32, 42)
	for x := 0; x < 32; x++ {
		for y := 0; y < 32; y++ {
			terr.SetTile(x, y, terrain.TileFloor)
		}
	}
	terr.SetTile(5, 5, terrain.TileCorridor)
	sys.SetTerrain(terr)
	sys.SetTileSize(32)

	entities := make([]*Entity, 100)
	for i := range entities {
		e := NewEntity(uint64(i + 1))
		e.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		e.AddComponent(&ManaComponent{Current: 50, Max: 100})
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
