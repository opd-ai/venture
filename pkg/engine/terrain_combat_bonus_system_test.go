package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

func TestNewTerrainCombatBonusSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewTerrainCombatBonusSystem returned nil")
	}
	if sys.world != world {
		t.Error("world not set correctly")
	}
	if sys.tileSize != 32 {
		t.Errorf("tileSize = %d, want 32", sys.tileSize)
	}
	if sys.bonusCache == nil {
		t.Error("bonusCache is nil")
	}
	if sys.lastTileCache == nil {
		t.Error("lastTileCache is nil")
	}
}

func TestTerrainCombatBonusSystem_SetTerrain(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusSystem(world, 12345)

	terr := terrain.NewTerrain(100, 100, 12345)
	sys.SetTerrain(terr)

	if sys.terrain != terr {
		t.Error("terrain not set correctly")
	}
}

func TestTerrainCombatBonusSystem_SetTileSize(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusSystem(world, 12345)

	sys.SetTileSize(64)
	if sys.tileSize != 64 {
		t.Errorf("tileSize = %d, want 64", sys.tileSize)
	}

	// Invalid size should not change
	sys.SetTileSize(0)
	if sys.tileSize != 64 {
		t.Errorf("tileSize changed on invalid input: %d", sys.tileSize)
	}
	sys.SetTileSize(-1)
	if sys.tileSize != 64 {
		t.Errorf("tileSize changed on negative input: %d", sys.tileSize)
	}
}

func TestTerrainCombatBonusSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusSystem(world, 12345)

	sys.SetGenre("cyberpunk")
	if sys.genreID != "cyberpunk" {
		t.Errorf("genreID = %s, want cyberpunk", sys.genreID)
	}
}

func TestTerrainCombatBonusSystem_PlatformBonus(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusSystem(world, 12345)

	terr := terrain.NewTerrain(100, 100, 12345)
	terr.SetTile(5, 5, terrain.TilePlatform)
	sys.SetTerrain(terr)
	sys.SetTileSize(32)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 5 * 32, Y: 5 * 32})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&StatsComponent{})

	sys.Update([]*Entity{entity}, 0.016)

	bonus := sys.GetTerrainBonus(entity.ID)
	if bonus == nil {
		t.Fatal("expected bonus on platform tile")
	}
	if bonus.DamageBonus != 1.10 {
		t.Errorf("DamageBonus = %f, want 1.10", bonus.DamageBonus)
	}
	if bonus.TerrainType != "platform" {
		t.Errorf("TerrainType = %s, want platform", bonus.TerrainType)
	}
}

func TestTerrainCombatBonusSystem_BridgeBonus(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusSystem(world, 12345)

	terr := terrain.NewTerrain(100, 100, 12345)
	terr.SetTile(3, 3, terrain.TileBridge)
	sys.SetTerrain(terr)
	sys.SetTileSize(32)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 3 * 32, Y: 3 * 32})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&StatsComponent{})

	sys.Update([]*Entity{entity}, 0.016)

	bonus := sys.GetTerrainBonus(entity.ID)
	if bonus == nil {
		t.Fatal("expected bonus on bridge tile")
	}
	if bonus.DamageBonus != 1.10 {
		t.Errorf("DamageBonus = %f, want 1.10", bonus.DamageBonus)
	}
}

func TestTerrainCombatBonusSystem_WaterPenalty(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusSystem(world, 12345)

	terr := terrain.NewTerrain(100, 100, 12345)
	terr.SetTile(4, 4, terrain.TileWaterShallow)
	sys.SetTerrain(terr)
	sys.SetTileSize(32)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 4 * 32, Y: 4 * 32})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&StatsComponent{})

	sys.Update([]*Entity{entity}, 0.016)

	bonus := sys.GetTerrainBonus(entity.ID)
	if bonus == nil {
		t.Fatal("expected bonus on water tile")
	}
	if bonus.DefenseBonus != 0.85 {
		t.Errorf("DefenseBonus = %f, want 0.85", bonus.DefenseBonus)
	}
}

func TestTerrainCombatBonusSystem_CoverBonus(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusSystem(world, 12345)

	terr := terrain.NewTerrain(100, 100, 12345)
	// Create a corner with 2 adjacent walls
	terr.SetTile(5, 5, terrain.TileFloor)
	terr.SetTile(4, 5, terrain.TileWall) // West wall
	terr.SetTile(5, 4, terrain.TileWall) // North wall
	sys.SetTerrain(terr)
	sys.SetTileSize(32)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 5 * 32, Y: 5 * 32})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&StatsComponent{})

	sys.Update([]*Entity{entity}, 0.016)

	bonus := sys.GetTerrainBonus(entity.ID)
	if bonus == nil {
		t.Fatal("expected bonus with cover")
	}
	if bonus.EvasionBonus != 0.10 {
		t.Errorf("EvasionBonus = %f, want 0.10", bonus.EvasionBonus)
	}
}

func TestTerrainCombatBonusSystem_NoCoverWithOneWall(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusSystem(world, 12345)

	terr := terrain.NewTerrain(100, 100, 12345)
	// Only one adjacent wall - no cover
	terr.SetTile(5, 5, terrain.TileFloor)
	terr.SetTile(4, 5, terrain.TileWall)
	terr.SetTile(6, 5, terrain.TileFloor)
	terr.SetTile(5, 4, terrain.TileFloor)
	terr.SetTile(5, 6, terrain.TileFloor)
	sys.SetTerrain(terr)
	sys.SetTileSize(32)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 5 * 32, Y: 5 * 32})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&StatsComponent{})

	sys.Update([]*Entity{entity}, 0.016)

	bonus := sys.GetTerrainBonus(entity.ID)
	if bonus != nil {
		t.Error("expected no bonus with only one wall")
	}
}

func TestTerrainCombatBonusSystem_FantasyGenreBonus(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusSystem(world, 12345)
	sys.SetGenre("fantasy")

	terr := terrain.NewTerrain(100, 100, 12345)
	terr.SetTile(5, 5, terrain.TilePlatform)
	sys.SetTerrain(terr)
	sys.SetTileSize(32)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 5 * 32, Y: 5 * 32})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&StatsComponent{})

	sys.Update([]*Entity{entity}, 0.016)

	bonus := sys.GetTerrainBonus(entity.ID)
	if bonus == nil {
		t.Fatal("expected bonus")
	}
	if bonus.SpellDamageBonus != 1.05 {
		t.Errorf("SpellDamageBonus = %f, want 1.05", bonus.SpellDamageBonus)
	}
}

func TestTerrainCombatBonusSystem_ScifiGenreBonus(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusSystem(world, 12345)
	sys.SetGenre("scifi")

	terr := terrain.NewTerrain(100, 100, 12345)
	terr.SetTile(5, 5, terrain.TileBridge)
	sys.SetTerrain(terr)
	sys.SetTileSize(32)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 5 * 32, Y: 5 * 32})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&StatsComponent{})

	sys.Update([]*Entity{entity}, 0.016)

	bonus := sys.GetTerrainBonus(entity.ID)
	if bonus == nil {
		t.Fatal("expected bonus")
	}
	if bonus.AccuracyBonus != 0.10 {
		t.Errorf("AccuracyBonus = %f, want 0.10", bonus.AccuracyBonus)
	}
}

func TestTerrainCombatBonusSystem_HorrorGenreBonus(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusSystem(world, 12345)
	sys.SetGenre("horror")

	terr := terrain.NewTerrain(100, 100, 12345)
	terr.SetTile(5, 5, terrain.TileWaterShallow)
	sys.SetTerrain(terr)
	sys.SetTileSize(32)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 5 * 32, Y: 5 * 32})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&StatsComponent{})

	sys.Update([]*Entity{entity}, 0.016)

	bonus := sys.GetTerrainBonus(entity.ID)
	if bonus == nil {
		t.Fatal("expected bonus")
	}
	if bonus.DefenseBonus != 0.80 {
		t.Errorf("DefenseBonus = %f, want 0.80 (horror penalty)", bonus.DefenseBonus)
	}
}

func TestTerrainCombatBonusSystem_CyberpunkGenreBonus(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusSystem(world, 12345)
	sys.SetGenre("cyberpunk")

	terr := terrain.NewTerrain(100, 100, 12345)
	// Create corner for cover
	terr.SetTile(5, 5, terrain.TileFloor)
	terr.SetTile(4, 5, terrain.TileWall)
	terr.SetTile(5, 4, terrain.TileWall)
	sys.SetTerrain(terr)
	sys.SetTileSize(32)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 5 * 32, Y: 5 * 32})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&StatsComponent{})

	sys.Update([]*Entity{entity}, 0.016)

	bonus := sys.GetTerrainBonus(entity.ID)
	if bonus == nil {
		t.Fatal("expected bonus")
	}
	// Base cover (+10%) + cyberpunk bonus (+5%) = +15%
	// Use tolerance for floating point comparison
	expected := 0.15
	if bonus.EvasionBonus < expected-0.001 || bonus.EvasionBonus > expected+0.001 {
		t.Errorf("EvasionBonus = %f, want ~%f (cyberpunk cover)", bonus.EvasionBonus, expected)
	}
}

func TestTerrainCombatBonusSystem_PostapocGenreReduction(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusSystem(world, 12345)
	sys.SetGenre("postapoc")

	terr := terrain.NewTerrain(100, 100, 12345)
	terr.SetTile(5, 5, terrain.TilePlatform)
	sys.SetTerrain(terr)
	sys.SetTileSize(32)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 5 * 32, Y: 5 * 32})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&StatsComponent{})

	sys.Update([]*Entity{entity}, 0.016)

	bonus := sys.GetTerrainBonus(entity.ID)
	if bonus == nil {
		t.Fatal("expected bonus")
	}
	// Base platform (+10%) - postapoc penalty (5%) = +5%
	if bonus.DamageBonus != 1.05 {
		t.Errorf("DamageBonus = %f, want 1.05 (postapoc reduction)", bonus.DamageBonus)
	}
}

func TestTerrainCombatBonusSystem_NoBonusOnFloor(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusSystem(world, 12345)

	terr := terrain.NewTerrain(100, 100, 12345)
	terr.SetTile(5, 5, terrain.TileFloor)
	// No adjacent walls
	terr.SetTile(4, 5, terrain.TileFloor)
	terr.SetTile(6, 5, terrain.TileFloor)
	terr.SetTile(5, 4, terrain.TileFloor)
	terr.SetTile(5, 6, terrain.TileFloor)
	sys.SetTerrain(terr)
	sys.SetTileSize(32)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 5 * 32, Y: 5 * 32})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&StatsComponent{})

	sys.Update([]*Entity{entity}, 0.016)

	bonus := sys.GetTerrainBonus(entity.ID)
	if bonus != nil {
		t.Error("expected no bonus on floor tile without cover")
	}
}

func TestTerrainCombatBonusSystem_NoTerrainSet(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusSystem(world, 12345)
	// Don't set terrain

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&StatsComponent{})

	// Should not panic
	sys.Update([]*Entity{entity}, 0.016)

	bonus := sys.GetTerrainBonus(entity.ID)
	if bonus != nil {
		t.Error("expected no bonus when terrain not set")
	}
}

func TestTerrainCombatBonusSystem_EntityWithoutHealth(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusSystem(world, 12345)

	terr := terrain.NewTerrain(100, 100, 12345)
	terr.SetTile(5, 5, terrain.TilePlatform)
	sys.SetTerrain(terr)
	sys.SetTileSize(32)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 5 * 32, Y: 5 * 32})
	// No health component - not a combatant

	sys.Update([]*Entity{entity}, 0.016)

	bonus := sys.GetTerrainBonus(entity.ID)
	if bonus != nil {
		t.Error("expected no bonus for entity without health")
	}
}

func TestTerrainCombatBonusSystem_EntityWithoutPosition(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusSystem(world, 12345)

	terr := terrain.NewTerrain(100, 100, 12345)
	sys.SetTerrain(terr)

	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&StatsComponent{})
	// No position component

	sys.Update([]*Entity{entity}, 0.016)

	bonus := sys.GetTerrainBonus(entity.ID)
	if bonus != nil {
		t.Error("expected no bonus for entity without position")
	}
}

func TestTerrainCombatBonusSystem_CacheInvalidation(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusSystem(world, 12345)

	terr := terrain.NewTerrain(100, 100, 12345)
	terr.SetTile(5, 5, terrain.TilePlatform)
	terr.SetTile(6, 5, terrain.TileFloor)
	// No walls adjacent to floor tile
	terr.SetTile(5, 4, terrain.TileFloor)
	terr.SetTile(7, 5, terrain.TileFloor)
	terr.SetTile(6, 4, terrain.TileFloor)
	terr.SetTile(6, 6, terrain.TileFloor)
	sys.SetTerrain(terr)
	sys.SetTileSize(32)

	entity := world.CreateEntity()
	pos := &PositionComponent{X: 5 * 32, Y: 5 * 32}
	entity.AddComponent(pos)
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&StatsComponent{})

	// First update - on platform
	sys.Update([]*Entity{entity}, 0.016)
	bonus := sys.GetTerrainBonus(entity.ID)
	if bonus == nil || bonus.DamageBonus != 1.10 {
		t.Fatal("expected platform bonus")
	}

	// Move to floor tile
	pos.X = 6 * 32
	sys.Update([]*Entity{entity}, 0.016)
	bonus = sys.GetTerrainBonus(entity.ID)
	if bonus != nil {
		t.Error("expected no bonus after moving to floor")
	}
}

func TestTerrainCombatBonusSystem_RampPenalty(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusSystem(world, 12345)

	terr := terrain.NewTerrain(100, 100, 12345)
	terr.SetTile(5, 5, terrain.TileRamp)
	// Clear adjacent tiles so no cover bonus interferes
	terr.SetTile(4, 5, terrain.TileFloor)
	terr.SetTile(6, 5, terrain.TileFloor)
	terr.SetTile(5, 4, terrain.TileFloor)
	terr.SetTile(5, 6, terrain.TileFloor)
	sys.SetTerrain(terr)
	sys.SetTileSize(32)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 5 * 32, Y: 5 * 32})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&StatsComponent{})

	sys.Update([]*Entity{entity}, 0.016)

	bonus := sys.GetTerrainBonus(entity.ID)
	if bonus == nil {
		t.Fatal("expected bonus on ramp")
	}
	if bonus.EvasionBonus != -0.05 {
		t.Errorf("EvasionBonus = %f, want -0.05 (ramp penalty)", bonus.EvasionBonus)
	}
}

func TestTerrainCombatBonusComponent_Type(t *testing.T) {
	comp := &TerrainCombatBonusComponent{}
	if comp.Type() != "terrain_combat_bonus" {
		t.Errorf("Type() = %s, want terrain_combat_bonus", comp.Type())
	}
}

func BenchmarkTerrainCombatBonusSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewTerrainCombatBonusSystem(world, 12345)

	terr := terrain.NewTerrain(100, 100, 12345)
	// Set up various terrain
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			switch (i + j) % 5 {
			case 0:
				terr.SetTile(i, j, terrain.TilePlatform)
			case 1:
				terr.SetTile(i, j, terrain.TileWaterShallow)
			case 2:
				terr.SetTile(i, j, terrain.TileWall)
			default:
				terr.SetTile(i, j, terrain.TileFloor)
			}
		}
	}
	sys.SetTerrain(terr)
	sys.SetTileSize(32)

	entities := make([]*Entity, 100)
	for i := range entities {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 32), Y: float64(i * 32)})
		entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
		entity.AddComponent(&StatsComponent{})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
