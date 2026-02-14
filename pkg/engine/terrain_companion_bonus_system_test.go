package engine

import (
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// TestNewTerrainCompanionBonusSystem verifies system creation.
func TestNewTerrainCompanionBonusSystem(t *testing.T) {
	world := NewWorld()
	system := NewTerrainCompanionBonusSystem(world, 12345)

	if system == nil {
		t.Fatal("expected non-nil system")
	}
	if system.world != world {
		t.Error("expected world reference to be set")
	}
	if system.tileSize != 32 {
		t.Errorf("expected default tile size 32, got %d", system.tileSize)
	}
	if len(system.terrainBonuses) == 0 {
		t.Error("expected terrain bonuses to be initialized")
	}
}

// TestTerrainCompanionBonusSystem_SetTerrain tests terrain setter.
func TestTerrainCompanionBonusSystem_SetTerrain(t *testing.T) {
	world := NewWorld()
	system := NewTerrainCompanionBonusSystem(world, 12345)

	rng := rand.New(rand.NewSource(42))
	terr := terrain.NewTerrain(100, 100, rng)

	system.SetTerrain(terr)

	if system.terrain != terr {
		t.Error("expected terrain to be set")
	}
}

// TestTerrainCompanionBonusSystem_SetTileSize tests tile size setter.
func TestTerrainCompanionBonusSystem_SetTileSize(t *testing.T) {
	world := NewWorld()
	system := NewTerrainCompanionBonusSystem(world, 12345)

	system.SetTileSize(64)
	if system.tileSize != 64 {
		t.Errorf("expected tile size 64, got %d", system.tileSize)
	}

	// Invalid size should be ignored
	system.SetTileSize(0)
	if system.tileSize != 64 {
		t.Error("expected invalid tile size to be ignored")
	}

	system.SetTileSize(-1)
	if system.tileSize != 64 {
		t.Error("expected negative tile size to be ignored")
	}
}

// TestTerrainCompanionBonusSystem_SetGenre tests genre setter.
func TestTerrainCompanionBonusSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewTerrainCompanionBonusSystem(world, 12345)

	system.SetGenre("horror")
	if system.genreID != "horror" {
		t.Errorf("expected genre 'horror', got '%s'", system.genreID)
	}
}

// TestTerrainCompanionBonusSystem_Update_NoTerrain tests update with nil terrain.
func TestTerrainCompanionBonusSystem_Update_NoTerrain(t *testing.T) {
	world := NewWorld()
	system := NewTerrainCompanionBonusSystem(world, 12345)

	companion := createTestCompanion(world, CompanionTypeElemental, 100, 100)
	entities := []*Entity{companion}

	// Should not panic with nil terrain
	system.Update(entities, 0.016)

	// No bonus should be applied
	if companion.HasComponent("terrain_companion_bonus") {
		t.Error("expected no terrain bonus without terrain data")
	}
}

// TestTerrainCompanionBonusSystem_Update_NonCompanion tests update with non-companion entity.
func TestTerrainCompanionBonusSystem_Update_NonCompanion(t *testing.T) {
	world := NewWorld()
	system := NewTerrainCompanionBonusSystem(world, 12345)

	rng := rand.New(rand.NewSource(42))
	terr := terrain.NewTerrain(100, 100, rng)
	system.SetTerrain(terr)

	// Create regular entity (not a companion)
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	entity.AddComponent(&StatsComponent{Attack: 10})

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// No bonus should be applied to non-companions
	if entity.HasComponent("terrain_companion_bonus") {
		t.Error("expected no terrain bonus for non-companion")
	}
}

// TestTerrainCompanionBonusSystem_Update_WaterElemental tests water elemental bonus in water.
func TestTerrainCompanionBonusSystem_Update_WaterElemental(t *testing.T) {
	world := NewWorld()
	system := NewTerrainCompanionBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create terrain with water at position 2,2
	rng := rand.New(rand.NewSource(42))
	terr := terrain.NewTerrain(10, 10, rng)
	terr.SetTile(2, 2, terrain.TileWaterShallow)
	system.SetTerrain(terr)

	// Create elemental companion at water tile (2*32 + 16 = 80 center)
	companion := createTestCompanion(world, CompanionTypeElemental, 80, 80)
	initialAttack := getCompanionAttack(companion)

	entities := []*Entity{companion}
	system.Update(entities, 0.016)

	// Should have terrain bonus
	if !companion.HasComponent("terrain_companion_bonus") {
		t.Fatal("expected terrain bonus for elemental in water")
	}

	bonusComp, _ := companion.GetComponent("terrain_companion_bonus")
	bonus := bonusComp.(*TerrainCompanionBonusComponent)

	if bonus.AttackBonus <= 1.0 {
		t.Errorf("expected positive attack bonus for elemental in water, got %f", bonus.AttackBonus)
	}

	// Stats should be increased
	newAttack := getCompanionAttack(companion)
	if newAttack <= initialAttack {
		t.Errorf("expected attack to increase: initial=%f, new=%f", initialAttack, newAttack)
	}
}

// TestTerrainCompanionBonusSystem_Update_RobotInWater tests robot penalty in water.
func TestTerrainCompanionBonusSystem_Update_RobotInWater(t *testing.T) {
	world := NewWorld()
	system := NewTerrainCompanionBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create terrain with water
	rng := rand.New(rand.NewSource(42))
	terr := terrain.NewTerrain(10, 10, rng)
	terr.SetTile(2, 2, terrain.TileWaterShallow)
	system.SetTerrain(terr)

	// Create robot companion at water tile
	companion := createTestCompanion(world, CompanionTypeRobot, 80, 80)
	initialAttack := getCompanionAttack(companion)

	entities := []*Entity{companion}
	system.Update(entities, 0.016)

	// Should have terrain bonus (penalty)
	if !companion.HasComponent("terrain_companion_bonus") {
		t.Fatal("expected terrain bonus component for robot in water")
	}

	bonusComp, _ := companion.GetComponent("terrain_companion_bonus")
	bonus := bonusComp.(*TerrainCompanionBonusComponent)

	if bonus.AttackBonus >= 1.0 {
		t.Errorf("expected negative attack bonus (penalty) for robot in water, got %f", bonus.AttackBonus)
	}

	// Stats should be decreased
	newAttack := getCompanionAttack(companion)
	if newAttack >= initialAttack {
		t.Errorf("expected attack to decrease: initial=%f, new=%f", initialAttack, newAttack)
	}
}

// TestTerrainCompanionBonusSystem_Update_InsectInForest tests insect bonus in trees.
func TestTerrainCompanionBonusSystem_Update_InsectInForest(t *testing.T) {
	world := NewWorld()
	system := NewTerrainCompanionBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create terrain with trees
	rng := rand.New(rand.NewSource(42))
	terr := terrain.NewTerrain(10, 10, rng)
	terr.SetTile(3, 3, terrain.TileTree)
	system.SetTerrain(terr)

	// Create insect companion at tree tile
	companion := createTestCompanion(world, CompanionTypeInsect, 112, 112) // 3*32+16
	initialAttack := getCompanionAttack(companion)

	entities := []*Entity{companion}
	system.Update(entities, 0.016)

	// Should have significant bonus
	if !companion.HasComponent("terrain_companion_bonus") {
		t.Fatal("expected terrain bonus for insect in forest")
	}

	bonusComp, _ := companion.GetComponent("terrain_companion_bonus")
	bonus := bonusComp.(*TerrainCompanionBonusComponent)

	if bonus.AttackBonus < 1.2 {
		t.Errorf("expected strong attack bonus for insect in trees, got %f", bonus.AttackBonus)
	}

	newAttack := getCompanionAttack(companion)
	if newAttack <= initialAttack {
		t.Errorf("expected attack increase: initial=%f, new=%f", initialAttack, newAttack)
	}
}

// TestTerrainCompanionBonusSystem_GenreMultipliers tests genre-specific scaling.
func TestTerrainCompanionBonusSystem_GenreMultipliers(t *testing.T) {
	tests := []struct {
		name           string
		genre          string
		expectedHigher bool // horror should give higher bonuses
	}{
		{"fantasy_baseline", "fantasy", false},
		{"horror_amplified", "horror", true},
		{"cyberpunk_reduced", "cyberpunk", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewTerrainCompanionBonusSystem(world, 12345)
			system.SetGenre(tt.genre)

			rng := rand.New(rand.NewSource(42))
			terr := terrain.NewTerrain(10, 10, rng)
			terr.SetTile(2, 2, terrain.TileWaterShallow)
			system.SetTerrain(terr)

			companion := createTestCompanion(world, CompanionTypeElemental, 80, 80)
			entities := []*Entity{companion}
			system.Update(entities, 0.016)

			bonusComp, _ := companion.GetComponent("terrain_companion_bonus")
			if bonusComp == nil {
				t.Fatal("expected terrain bonus")
			}
			bonus := bonusComp.(*TerrainCompanionBonusComponent)

			if tt.expectedHigher && bonus.AttackBonus <= 1.25 {
				t.Errorf("expected horror genre to amplify bonus: got %f", bonus.AttackBonus)
			}
		})
	}
}

// TestTerrainCompanionBonusSystem_MovementCacheUpdate tests bonus recalculation on tile change.
func TestTerrainCompanionBonusSystem_MovementCacheUpdate(t *testing.T) {
	world := NewWorld()
	system := NewTerrainCompanionBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	rng := rand.New(rand.NewSource(42))
	terr := terrain.NewTerrain(10, 10, rng)
	terr.SetTile(2, 2, terrain.TileWaterShallow)
	terr.SetTile(3, 3, terrain.TileTree)
	system.SetTerrain(terr)

	companion := createTestCompanion(world, CompanionTypeElemental, 80, 80) // Water tile
	entities := []*Entity{companion}

	system.Update(entities, 0.016)

	// Should have water bonus
	bonusComp, _ := companion.GetComponent("terrain_companion_bonus")
	waterBonus := bonusComp.(*TerrainCompanionBonusComponent)
	if waterBonus.TerrainType != terrain.TileWaterShallow.String() {
		t.Errorf("expected water terrain type, got %s", waterBonus.TerrainType)
	}

	// Move to tree tile
	pos := companion.GetPosition()
	pos.X = 112 // 3*32+16
	pos.Y = 112

	system.Update(entities, 0.016)

	// Bonus should update for tree terrain
	bonusComp2, _ := companion.GetComponent("terrain_companion_bonus")
	if bonusComp2 != nil {
		treeBonus := bonusComp2.(*TerrainCompanionBonusComponent)
		if treeBonus.TerrainType != terrain.TileTree.String() {
			t.Errorf("expected tree terrain type after movement, got %s", treeBonus.TerrainType)
		}
	}
}

// TestTerrainCompanionBonusSystem_SameTileNoRecompute tests caching efficiency.
func TestTerrainCompanionBonusSystem_SameTileNoRecompute(t *testing.T) {
	world := NewWorld()
	system := NewTerrainCompanionBonusSystem(world, 12345)

	rng := rand.New(rand.NewSource(42))
	terr := terrain.NewTerrain(10, 10, rng)
	terr.SetTile(2, 2, terrain.TileWater)
	system.SetTerrain(terr)

	companion := createTestCompanion(world, CompanionTypeElemental, 80, 80)
	entities := []*Entity{companion}

	// First update
	system.Update(entities, 0.016)
	initialBonus, _ := companion.GetComponent("terrain_companion_bonus")

	// Second update on same tile
	system.Update(entities, 0.016)
	secondBonus, _ := companion.GetComponent("terrain_companion_bonus")

	// Should be same component instance (not recreated)
	if initialBonus != secondBonus {
		t.Error("expected cached bonus to be reused on same tile")
	}
}

// TestTerrainCompanionBonusSystem_RemoveBonus tests bonus removal.
func TestTerrainCompanionBonusSystem_RemoveBonus(t *testing.T) {
	world := NewWorld()
	system := NewTerrainCompanionBonusSystem(world, 12345)

	rng := rand.New(rand.NewSource(42))
	terr := terrain.NewTerrain(10, 10, rng)
	terr.SetTile(2, 2, terrain.TileWater)
	system.SetTerrain(terr)

	companion := createTestCompanion(world, CompanionTypeElemental, 80, 80)
	initialAttack := getCompanionAttack(companion)
	entities := []*Entity{companion}

	// Apply bonus
	system.Update(entities, 0.016)
	if !system.HasActiveBonus(companion.ID) {
		t.Fatal("expected active bonus after update")
	}

	// Remove companion component to trigger bonus removal
	companion.RemoveComponent("companion")
	system.Update(entities, 0.016)

	if system.HasActiveBonus(companion.ID) {
		t.Error("expected bonus to be removed after companion component removed")
	}

	// Stats should be reverted (approximately)
	finalAttack := getCompanionAttack(companion)
	diff := finalAttack - initialAttack
	if diff > 0.1 || diff < -0.1 {
		t.Errorf("expected stats to revert: initial=%f, final=%f, diff=%f", initialAttack, finalAttack, diff)
	}
}

// TestTerrainCompanionBonusSystem_GetBonusCount tests bonus count tracking.
func TestTerrainCompanionBonusSystem_GetBonusCount(t *testing.T) {
	world := NewWorld()
	system := NewTerrainCompanionBonusSystem(world, 12345)

	rng := rand.New(rand.NewSource(42))
	terr := terrain.NewTerrain(10, 10, rng)
	terr.SetTile(2, 2, terrain.TileWater)
	terr.SetTile(3, 2, terrain.TileWater)
	system.SetTerrain(terr)

	comp1 := createTestCompanion(world, CompanionTypeElemental, 80, 80)
	comp2 := createTestCompanion(world, CompanionTypeElemental, 112, 80)
	entities := []*Entity{comp1, comp2}

	system.Update(entities, 0.016)

	count := system.GetBonusCount()
	if count != 2 {
		t.Errorf("expected 2 active bonuses, got %d", count)
	}
}

// TestTerrainCompanionBonusComponent_Type tests component type identifier.
func TestTerrainCompanionBonusComponent_Type(t *testing.T) {
	comp := &TerrainCompanionBonusComponent{}
	if comp.Type() != "terrain_companion_bonus" {
		t.Errorf("expected 'terrain_companion_bonus', got '%s'", comp.Type())
	}
}

// TestTerrainCompanionBonusSystem_CompanionTypeName tests type name mapping.
func TestTerrainCompanionBonusSystem_CompanionTypeName(t *testing.T) {
	world := NewWorld()
	system := NewTerrainCompanionBonusSystem(world, 12345)

	tests := []struct {
		compType CompanionType
		expected string
	}{
		{CompanionTypePet, "Pet"},
		{CompanionTypeSummon, "Summon"},
		{CompanionTypeHireling, "Hireling"},
		{CompanionTypeElemental, "Elemental"},
		{CompanionTypeUndead, "Undead"},
		{CompanionTypeRobot, "Robot"},
		{CompanionTypeSpirit, "Spirit"},
		{CompanionTypeInsect, "Insect"},
		{CompanionType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := system.companionTypeName(tt.compType)
			if got != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, got)
			}
		})
	}
}

// Helper function to create a test companion entity.
func createTestCompanion(world *World, compType CompanionType, x, y float64) *Entity {
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: x, Y: y})
	entity.AddComponent(&CompanionComponent{
		CompanionType: compType,
		OwnerID:       1,
		Loyalty:       50,
		Level:         1,
	})
	entity.AddComponent(&CompanionStatsComponent{
		Attack:  10.0,
		Defense: 8.0,
		Speed:   5.0,
		HP:      100.0,
		MaxHP:   100.0,
	})
	return entity
}

// Helper function to get companion attack stat.
func getCompanionAttack(entity *Entity) float64 {
	statsComp, ok := entity.GetComponent("companionstats")
	if !ok {
		return 0
	}
	stats, ok := statsComp.(*CompanionStatsComponent)
	if !ok {
		return 0
	}
	return stats.Attack
}

// BenchmarkTerrainCompanionBonusSystem_Update benchmarks the Update method.
func BenchmarkTerrainCompanionBonusSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewTerrainCompanionBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	rng := rand.New(rand.NewSource(42))
	terr := terrain.NewTerrain(100, 100, rng)
	// Fill with various terrain types
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			switch (x + y) % 5 {
			case 0:
				terr.SetTile(x, y, terrain.TileWater)
			case 1:
				terr.SetTile(x, y, terrain.TileTree)
			case 2:
				terr.SetTile(x, y, terrain.TilePlatform)
			case 3:
				terr.SetTile(x, y, terrain.TilePit)
			default:
				terr.SetTile(x, y, terrain.TileFloor)
			}
		}
	}
	system.SetTerrain(terr)

	// Create 50 companions at various positions
	entities := make([]*Entity, 50)
	compTypes := []CompanionType{
		CompanionTypeElemental, CompanionTypeUndead, CompanionTypeRobot,
		CompanionTypeSpirit, CompanionTypeInsect, CompanionTypePet,
	}
	for i := 0; i < 50; i++ {
		entities[i] = createTestCompanion(world, compTypes[i%len(compTypes)],
			float64((i%10)*32+16), float64((i/10)*32+16))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}
