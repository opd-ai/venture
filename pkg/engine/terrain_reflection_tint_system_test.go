package engine

import (
	"math"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// TestTerrainReflectionTintComponentType verifies the component type string.
func TestTerrainReflectionTintComponentType(t *testing.T) {
	comp := NewTerrainReflectionTintComponent()
	if comp.Type() != "terrain_reflection_tint" {
		t.Errorf("expected type 'terrain_reflection_tint', got %q", comp.Type())
	}
}

// TestTerrainReflectionTintComponentDefaults verifies neutral defaults.
func TestTerrainReflectionTintComponentDefaults(t *testing.T) {
	comp := NewTerrainReflectionTintComponent()
	if comp.TintR != 1.0 || comp.TintG != 1.0 || comp.TintB != 1.0 {
		t.Errorf("expected neutral tint (1,1,1), got (%f,%f,%f)", comp.TintR, comp.TintG, comp.TintB)
	}
	if comp.LastTileType != -1 {
		t.Errorf("expected LastTileType -1, got %d", comp.LastTileType)
	}
}

// TestNewTerrainReflectionTintSystem verifies system creation.
func TestNewTerrainReflectionTintSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainReflectionTintSystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got %q", sys.genreID)
	}
	if sys.tileSize != 32 {
		t.Errorf("expected tile size 32, got %d", sys.tileSize)
	}
}

// TestTerrainReflectionTintSetGenre verifies genre switching.
func TestTerrainReflectionTintSetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainReflectionTintSystem(world, 42)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, g := range genres {
		sys.SetGenre(g)
		if sys.genreID != g {
			t.Errorf("expected genre %q, got %q", g, sys.genreID)
		}
	}
}

// TestTerrainReflectionTintSetTileSize verifies tile size setting.
func TestTerrainReflectionTintSetTileSize(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainReflectionTintSystem(world, 42)

	sys.SetTileSize(64)
	if sys.tileSize != 64 {
		t.Errorf("expected tile size 64, got %d", sys.tileSize)
	}

	sys.SetTileSize(0)
	if sys.tileSize != 64 {
		t.Errorf("expected tile size to remain 64, got %d", sys.tileSize)
	}
	sys.SetTileSize(-1)
	if sys.tileSize != 64 {
		t.Errorf("expected tile size to remain 64, got %d", sys.tileSize)
	}
}

// TestTerrainReflectionTintUpdateNoTerrain verifies no-op without terrain.
func TestTerrainReflectionTintUpdateNoTerrain(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainReflectionTintSystem(world, 42)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	entity.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})

	sys.Update([]*Entity{entity}, 0.016)
	_, ok := entity.GetComponent("terrain_reflection_tint")
	if ok {
		t.Error("should not add tint component without terrain")
	}
}

// TestTerrainReflectionTintUpdateNilEntity verifies nil entity handling.
func TestTerrainReflectionTintUpdateNilEntity(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainReflectionTintSystem(world, 42)
	terr := terrain.NewTerrain(10, 10, 42)
	sys.SetTerrain(terr)

	sys.Update([]*Entity{nil}, 0.016)
}

// TestTerrainReflectionTintUpdateNoPosition verifies skip without position.
func TestTerrainReflectionTintUpdateNoPosition(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainReflectionTintSystem(world, 42)
	terr := terrain.NewTerrain(10, 10, 42)
	sys.SetTerrain(terr)

	entity := world.CreateEntity()
	entity.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})

	sys.Update([]*Entity{entity}, 0.016)
	_, ok := entity.GetComponent("terrain_reflection_tint")
	if ok {
		t.Error("should not process entity without position")
	}
}

// TestTerrainReflectionTintUpdateNoSprite verifies skip without sprite.
func TestTerrainReflectionTintUpdateNoSprite(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainReflectionTintSystem(world, 42)
	terr := terrain.NewTerrain(10, 10, 42)
	sys.SetTerrain(terr)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})

	sys.Update([]*Entity{entity}, 0.016)
	_, ok := entity.GetComponent("terrain_reflection_tint")
	if ok {
		t.Error("should not process entity without sprite")
	}
}

// TestTerrainReflectionTintTileTypes verifies tint values for different terrain types.
func TestTerrainReflectionTintTileTypes(t *testing.T) {
	tests := []struct {
		name     string
		tileType terrain.TileType
		wantTint bool
	}{
		{"floor_neutral", terrain.TileFloor, false},
		{"corridor_neutral", terrain.TileCorridor, false},
		{"door_neutral", terrain.TileDoor, false},
		{"water_shallow_blue", terrain.TileWaterShallow, true},
		{"water_deep_blue", terrain.TileWaterDeep, true},
		{"lava_orange", terrain.TileLavaFlow, true},
		{"tree_green", terrain.TileTree, true},
		{"bridge_warm", terrain.TileBridge, true},
		{"structure_grey", terrain.TileStructure, true},
		{"platform_bright", terrain.TilePlatform, true},
		{"pit_dark", terrain.TilePit, true},
		{"trap_door_danger", terrain.TileTrapDoor, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewTerrainReflectionTintSystem(world, 42)
			terr := terrain.NewTerrain(10, 10, 42)
			terr.SetTile(1, 1, tt.tileType)
			sys.SetTerrain(terr)

			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: 40, Y: 40})
			entity.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})

			sys.Update([]*Entity{entity}, 0.016)

			comp, ok := entity.GetComponent("terrain_reflection_tint")
			if !ok {
				t.Fatal("expected terrain_reflection_tint component")
			}
			tint := comp.(*TerrainReflectionTintComponent)

			isNeutral := tint.TintR == 1.0 && tint.TintG == 1.0 && tint.TintB == 1.0
			if tt.wantTint && isNeutral {
				t.Errorf("expected non-neutral tint for %s, got (1.0, 1.0, 1.0)", tt.name)
			}
			if !tt.wantTint && !isNeutral {
				t.Errorf("expected neutral tint for %s, got (%f, %f, %f)",
					tt.name, tint.TintR, tint.TintG, tint.TintB)
			}
		})
	}
}

// TestTerrainReflectionTintGenreVariation verifies genre presets produce different intensities.
func TestTerrainReflectionTintGenreVariation(t *testing.T) {
	genres := []string{"fantasy", "horror", "cyberpunk"}
	results := make(map[string]float64)

	for _, genre := range genres {
		world := NewWorld()
		sys := NewTerrainReflectionTintSystem(world, 42)
		sys.SetGenre(genre)
		terr := terrain.NewTerrain(10, 10, 42)
		terr.SetTile(1, 1, terrain.TileLavaFlow)
		sys.SetTerrain(terr)

		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: 40, Y: 40})
		entity.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})

		sys.Update([]*Entity{entity}, 0.016)

		comp, ok := entity.GetComponent("terrain_reflection_tint")
		if !ok {
			t.Fatalf("genre %s: expected tint component", genre)
		}
		tint := comp.(*TerrainReflectionTintComponent)
		results[genre] = tint.TintR
	}

	fantasyDev := math.Abs(results["fantasy"] - 1.0)
	horrorDev := math.Abs(results["horror"] - 1.0)
	if fantasyDev <= horrorDev {
		t.Errorf("expected fantasy (%f) to have stronger tint than horror (%f)",
			fantasyDev, horrorDev)
	}
}

// TestTerrainReflectionTintCacheSkip verifies same tile type skips recomputation.
func TestTerrainReflectionTintCacheSkip(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainReflectionTintSystem(world, 42)
	terr := terrain.NewTerrain(10, 10, 42)
	terr.SetTile(1, 1, terrain.TileWaterShallow)
	terr.SetTile(2, 2, terrain.TileLavaFlow)
	sys.SetTerrain(terr)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 40, Y: 40})
	entity.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})

	// First update creates component
	sys.Update([]*Entity{entity}, 0.016)
	comp, ok := entity.GetComponent("terrain_reflection_tint")
	if !ok {
		t.Fatal("expected tint component after first update")
	}
	tint := comp.(*TerrainReflectionTintComponent)

	// Manually modify to detect if it gets overwritten
	tint.TintR = 0.5

	// Second update on same tile should skip (cache hit)
	sys.Update([]*Entity{entity}, 0.016)
	if tint.TintR != 0.5 {
		t.Errorf("expected cache hit to skip update, TintR changed from 0.5 to %f", tint.TintR)
	}

	// Move to different tile type to trigger cache miss
	posComp, _ := entity.GetComponent("position")
	pos := posComp.(*PositionComponent)
	pos.X = 70
	pos.Y = 70

	sys.Update([]*Entity{entity}, 0.016)
	if tint.TintR == 0.5 {
		t.Error("expected cache miss to update tint after tile change")
	}
}

// TestApplyTerrainTint verifies the tint blending math.
func TestApplyTerrainTint(t *testing.T) {
	tests := []struct {
		name       string
		baseColor  float64
		intensity  float64
		saturation float64
		want       float64
	}{
		{"neutral_base", 1.0, 0.7, 0.9, 1.0},
		{"darken", 0.9, 1.0, 1.0, 0.9},
		{"brighten", 1.1, 1.0, 1.0, 1.1},
		{"half_intensity", 0.8, 0.5, 1.0, 0.9},
		{"zero_intensity", 0.5, 0.0, 1.0, 1.0},
		{"zero_saturation", 0.5, 1.0, 0.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyTerrainTint(tt.baseColor, tt.intensity, tt.saturation)
			if math.Abs(got-tt.want) > 0.001 {
				t.Errorf("applyTerrainTint(%f, %f, %f) = %f, want %f",
					tt.baseColor, tt.intensity, tt.saturation, got, tt.want)
			}
		})
	}
}

// TestTerrainReflectionTintSetTerrainClearsCache verifies cache is cleared on terrain change.
func TestTerrainReflectionTintSetTerrainClearsCache(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainReflectionTintSystem(world, 42)

	terr1 := terrain.NewTerrain(10, 10, 42)
	sys.SetTerrain(terr1)
	sys.lastTileCache[1] = 5

	terr2 := terrain.NewTerrain(10, 10, 42)
	sys.SetTerrain(terr2)

	if len(sys.lastTileCache) != 0 {
		t.Error("expected cache to be cleared after SetTerrain")
	}
}

// BenchmarkTerrainReflectionTintUpdate benchmarks system update performance.
func BenchmarkTerrainReflectionTintUpdate(b *testing.B) {
	world := NewWorld()
	sys := NewTerrainReflectionTintSystem(world, 42)
	terr := terrain.NewTerrain(100, 100, 42)

	types := []terrain.TileType{
		terrain.TileFloor, terrain.TileWaterShallow, terrain.TileLavaFlow,
		terrain.TileTree, terrain.TileStructure, terrain.TilePlatform,
	}
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			terr.SetTile(x, y, types[(x+y)%len(types)])
		}
	}
	sys.SetTerrain(terr)

	entities := make([]*Entity, 200)
	for i := range entities {
		e := world.CreateEntity()
		e.AddComponent(&PositionComponent{
			X: float64((i * 17) % 3200),
			Y: float64((i * 31) % 3200),
		})
		e.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})
		entities[i] = e
	}

	sys.Update(entities, 0.016)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(entities); j += 10 {
			posComp, _ := entities[j].GetComponent("position")
			pos := posComp.(*PositionComponent)
			pos.X += 32
		}
		sys.Update(entities, 0.016)
	}
}
