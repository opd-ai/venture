package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

func TestNewTerrainManaRegenSystem(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainManaRegenSystem(world, 12345)

	if system == nil {
		t.Fatal("NewTerrainManaRegenSystem returned nil")
	}
	if system.world != world {
		t.Error("World not set correctly")
	}
	if system.tileSize != 32 {
		t.Errorf("Expected tileSize 32, got %d", system.tileSize)
	}
	if system.genreID != "fantasy" {
		t.Errorf("Expected default genre 'fantasy', got '%s'", system.genreID)
	}
}

func TestTerrainManaRegenSystem_SetTerrain(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainManaRegenSystem(world, 12345)

	// Create a small test terrain
	terr := terrain.NewTerrain(10, 10)
	system.SetTerrain(terr)

	if system.terrain != terr {
		t.Error("Terrain not set correctly")
	}
}

func TestTerrainManaRegenSystem_SetTileSize(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainManaRegenSystem(world, 12345)

	system.SetTileSize(64)
	if system.tileSize != 64 {
		t.Errorf("Expected tileSize 64, got %d", system.tileSize)
	}

	// Test invalid size
	system.SetTileSize(0)
	if system.tileSize != 64 {
		t.Error("SetTileSize should ignore zero value")
	}

	system.SetTileSize(-1)
	if system.tileSize != 64 {
		t.Error("SetTileSize should ignore negative value")
	}
}

func TestTerrainManaRegenSystem_SetGenre(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainManaRegenSystem(world, 12345)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		system.SetGenre(genre)
		if system.genreID != genre {
			t.Errorf("Expected genre '%s', got '%s'", genre, system.genreID)
		}
	}
}

func TestTerrainManaRegenComponent_Type(t *testing.T) {
	comp := &TerrainManaRegenComponent{
		RegenMultiplier: 1.25,
		TerrainType:     "shallow_water",
	}
	if comp.Type() != "terrain_mana_regen" {
		t.Errorf("Expected type 'terrain_mana_regen', got '%s'", comp.Type())
	}
}

func TestTerrainManaRegenSystem_UpdateNoTerrain(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainManaRegenSystem(world, 12345)

	entity := NewEntity(world)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})

	// Should not panic without terrain
	system.Update([]*Entity{entity}, 0.016)
}

func TestTerrainManaRegenSystem_WaterBonus(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainManaRegenSystem(world, 12345)

	terr := terrain.NewTerrain(10, 10)
	// Set tile at position (3, 3) to shallow water
	terr.SetTile(3, 3, terrain.TileWaterShallow)
	system.SetTerrain(terr)

	entity := NewEntity(world)
	entity.AddComponent(&PositionComponent{X: 96, Y: 96}) // Tile (3, 3) at tileSize 32
	entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 10.0})

	system.Update([]*Entity{entity}, 0.016)

	manaComp, _ := entity.GetComponent("mana")
	mana := manaComp.(*ManaComponent)

	// Water should give +25% bonus (1.25 * 10.0 = 12.5)
	expectedRegen := 12.5
	if mana.Regen != expectedRegen {
		t.Errorf("Expected regen %.1f for water tile, got %.1f", expectedRegen, mana.Regen)
	}

	// Verify component was added
	if !entity.HasComponent("terrain_mana_regen") {
		t.Error("Expected terrain_mana_regen component to be added")
	}
}

func TestTerrainManaRegenSystem_LavaPenalty(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainManaRegenSystem(world, 12345)

	terr := terrain.NewTerrain(10, 10)
	terr.SetTile(2, 2, terrain.TileLavaFlow)
	system.SetTerrain(terr)

	entity := NewEntity(world)
	entity.AddComponent(&PositionComponent{X: 64, Y: 64}) // Tile (2, 2)
	entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 10.0})

	system.Update([]*Entity{entity}, 0.016)

	manaComp, _ := entity.GetComponent("mana")
	mana := manaComp.(*ManaComponent)

	// Lava should give -30% penalty (0.70 * 10.0 = 7.0)
	expectedRegen := 7.0
	if mana.Regen != expectedRegen {
		t.Errorf("Expected regen %.1f for lava tile, got %.1f", expectedRegen, mana.Regen)
	}
}

func TestTerrainManaRegenSystem_PlatformBonus(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainManaRegenSystem(world, 12345)

	terr := terrain.NewTerrain(10, 10)
	terr.SetTile(4, 4, terrain.TilePlatform)
	system.SetTerrain(terr)

	entity := NewEntity(world)
	entity.AddComponent(&PositionComponent{X: 128, Y: 128}) // Tile (4, 4)
	entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 10.0})

	system.Update([]*Entity{entity}, 0.016)

	manaComp, _ := entity.GetComponent("mana")
	mana := manaComp.(*ManaComponent)

	// Platform should give +15% bonus (1.15 * 10.0 = 11.5)
	expectedRegen := 11.5
	if mana.Regen != expectedRegen {
		t.Errorf("Expected regen %.1f for platform tile, got %.1f", expectedRegen, mana.Regen)
	}
}

func TestTerrainManaRegenSystem_NoManaComponent(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainManaRegenSystem(world, 12345)

	terr := terrain.NewTerrain(10, 10)
	terr.SetTile(3, 3, terrain.TileWaterShallow)
	system.SetTerrain(terr)

	// Entity without mana component
	entity := NewEntity(world)
	entity.AddComponent(&PositionComponent{X: 96, Y: 96})

	system.Update([]*Entity{entity}, 0.016)

	// Should not add terrain_mana_regen component
	if entity.HasComponent("terrain_mana_regen") {
		t.Error("Should not add terrain_mana_regen to entity without mana")
	}
}

func TestTerrainManaRegenSystem_GenreFantasyWaterBonus(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainManaRegenSystem(world, 12345)
	system.SetGenre("fantasy")

	terr := terrain.NewTerrain(10, 10)
	terr.SetTile(3, 3, terrain.TileWaterShallow)
	system.SetTerrain(terr)

	entity := NewEntity(world)
	entity.AddComponent(&PositionComponent{X: 96, Y: 96})
	entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 10.0})

	system.Update([]*Entity{entity}, 0.016)

	manaComp, _ := entity.GetComponent("mana")
	mana := manaComp.(*ManaComponent)

	// Fantasy water bonus: +35% (1.35 * 10.0 = 13.5)
	expectedRegen := 13.5
	if mana.Regen != expectedRegen {
		t.Errorf("Expected fantasy water regen %.1f, got %.1f", expectedRegen, mana.Regen)
	}
}

func TestTerrainManaRegenSystem_GenreHorrorReduction(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainManaRegenSystem(world, 12345)
	system.SetGenre("horror")

	terr := terrain.NewTerrain(10, 10)
	terr.SetTile(3, 3, terrain.TileWaterShallow)
	system.SetTerrain(terr)

	entity := NewEntity(world)
	entity.AddComponent(&PositionComponent{X: 96, Y: 96})
	entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 10.0})

	system.Update([]*Entity{entity}, 0.016)

	manaComp, _ := entity.GetComponent("mana")
	mana := manaComp.(*ManaComponent)

	// Horror halves bonuses: base +25% becomes +12.5% (1.125 * 10.0 = 11.25)
	expectedRegen := 11.25
	if mana.Regen != expectedRegen {
		t.Errorf("Expected horror water regen %.2f, got %.2f", expectedRegen, mana.Regen)
	}
}

func TestTerrainManaRegenSystem_GenrePostapocWaterPenalty(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainManaRegenSystem(world, 12345)
	system.SetGenre("postapoc")

	terr := terrain.NewTerrain(10, 10)
	terr.SetTile(3, 3, terrain.TileWaterShallow)
	system.SetTerrain(terr)

	entity := NewEntity(world)
	entity.AddComponent(&PositionComponent{X: 96, Y: 96})
	entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 10.0})

	system.Update([]*Entity{entity}, 0.016)

	manaComp, _ := entity.GetComponent("mana")
	mana := manaComp.(*ManaComponent)

	// Postapoc: water is contaminated, -10% (0.90 * 10.0 = 9.0)
	expectedRegen := 9.0
	if mana.Regen != expectedRegen {
		t.Errorf("Expected postapoc water regen %.1f, got %.1f", expectedRegen, mana.Regen)
	}
}

func TestTerrainManaRegenSystem_GenreCyberpunkStructure(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainManaRegenSystem(world, 12345)
	system.SetGenre("cyberpunk")

	terr := terrain.NewTerrain(10, 10)
	terr.SetTile(3, 3, terrain.TileStructure)
	system.SetTerrain(terr)

	entity := NewEntity(world)
	entity.AddComponent(&PositionComponent{X: 96, Y: 96})
	entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 10.0})

	system.Update([]*Entity{entity}, 0.016)

	manaComp, _ := entity.GetComponent("mana")
	mana := manaComp.(*ManaComponent)

	// Cyberpunk structure: +25% (1.25 * 10.0 = 12.5)
	expectedRegen := 12.5
	if mana.Regen != expectedRegen {
		t.Errorf("Expected cyberpunk structure regen %.1f, got %.1f", expectedRegen, mana.Regen)
	}
}

func TestTerrainManaRegenSystem_MovingBetweenTiles(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainManaRegenSystem(world, 12345)

	terr := terrain.NewTerrain(10, 10)
	terr.SetTile(3, 3, terrain.TileWaterShallow) // +25%
	terr.SetTile(4, 4, terrain.TileFloor)        // normal
	system.SetTerrain(terr)

	entity := NewEntity(world)
	pos := &PositionComponent{X: 96, Y: 96} // Tile (3, 3)
	entity.AddComponent(pos)
	entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 10.0})

	// First update on water tile
	system.Update([]*Entity{entity}, 0.016)

	manaComp, _ := entity.GetComponent("mana")
	mana := manaComp.(*ManaComponent)

	if mana.Regen != 12.5 {
		t.Errorf("Expected water regen 12.5, got %.1f", mana.Regen)
	}

	// Move to floor tile
	pos.X = 128
	pos.Y = 128

	system.Update([]*Entity{entity}, 0.016)

	// Regen should be restored to original
	if mana.Regen != 10.0 {
		t.Errorf("Expected restored regen 10.0, got %.1f", mana.Regen)
	}

	// Component should be removed
	if entity.HasComponent("terrain_mana_regen") {
		t.Error("terrain_mana_regen should be removed on normal terrain")
	}
}

func TestTerrainManaRegenSystem_CachingBehavior(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainManaRegenSystem(world, 12345)

	terr := terrain.NewTerrain(10, 10)
	terr.SetTile(3, 3, terrain.TileWaterShallow)
	system.SetTerrain(terr)

	entity := NewEntity(world)
	pos := &PositionComponent{X: 96, Y: 96}
	entity.AddComponent(pos)
	entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 10.0})

	// First update
	system.Update([]*Entity{entity}, 0.016)

	// Get cached multiplier
	cachedMultiplier := system.regenCache[entity.ID]
	if cachedMultiplier != 1.25 {
		t.Errorf("Expected cached multiplier 1.25, got %.2f", cachedMultiplier)
	}

	// Update again on same tile - should use cache
	system.Update([]*Entity{entity}, 0.016)

	// Cache should still be the same
	if system.regenCache[entity.ID] != cachedMultiplier {
		t.Error("Cache should not change when entity stays on same tile")
	}
}

func TestTerrainManaRegenSystem_AllTerrainTypes(t *testing.T) {
	tests := []struct {
		name       string
		tileType   terrain.TileType
		wantMult   float64
		wantChange bool
	}{
		{"floor", terrain.TileFloor, 1.0, false},
		{"wall", terrain.TileWall, 1.0, false},
		{"shallow_water", terrain.TileWaterShallow, 1.25, true},
		{"deep_water", terrain.TileWaterDeep, 1.25, true},
		{"platform", terrain.TilePlatform, 1.15, true},
		{"bridge", terrain.TileBridge, 1.10, true},
		{"lava", terrain.TileLavaFlow, 0.70, true},
		{"structure", terrain.TileStructure, 1.10, true},
		{"tree", terrain.TileTree, 1.05, true},
		{"corridor", terrain.TileCorridor, 1.0, false},
		{"door", terrain.TileDoor, 1.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld(nil)
			system := NewTerrainManaRegenSystem(world, 12345)

			terr := terrain.NewTerrain(10, 10)
			terr.SetTile(3, 3, tt.tileType)
			system.SetTerrain(terr)

			entity := NewEntity(world)
			entity.AddComponent(&PositionComponent{X: 96, Y: 96})
			entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 10.0})

			system.Update([]*Entity{entity}, 0.016)

			manaComp, _ := entity.GetComponent("mana")
			mana := manaComp.(*ManaComponent)

			expectedRegen := 10.0 * tt.wantMult
			if mana.Regen != expectedRegen {
				t.Errorf("Expected regen %.2f for %s, got %.2f", expectedRegen, tt.name, mana.Regen)
			}

			hasComp := entity.HasComponent("terrain_mana_regen")
			if tt.wantChange && !hasComp {
				t.Errorf("Expected terrain_mana_regen component for %s", tt.name)
			}
			if !tt.wantChange && hasComp {
				t.Errorf("Did not expect terrain_mana_regen component for %s", tt.name)
			}
		})
	}
}

func BenchmarkTerrainManaRegenSystem_Update(b *testing.B) {
	world := NewWorld(nil)
	system := NewTerrainManaRegenSystem(world, 12345)

	terr := terrain.NewTerrain(100, 100)
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			if (x+y)%3 == 0 {
				terr.SetTile(x, y, terrain.TileWaterShallow)
			}
		}
	}
	system.SetTerrain(terr)

	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := NewEntity(world)
		entity.AddComponent(&PositionComponent{X: float64(i*32 + 16), Y: float64(i*32 + 16)})
		entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 10.0})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}
