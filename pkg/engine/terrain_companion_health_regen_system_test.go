package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

func TestNewTerrainCompanionHealthRegenSystem(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainCompanionHealthRegenSystem(world, 12345)

	if system == nil {
		t.Fatal("NewTerrainCompanionHealthRegenSystem returned nil")
	}

	if system.world != world {
		t.Error("world not set correctly")
	}

	if system.tileSize != 32 {
		t.Errorf("expected tileSize 32, got %d", system.tileSize)
	}

	if system.perkBonusMultiplier != 1.5 {
		t.Errorf("expected perkBonusMultiplier 1.5, got %f", system.perkBonusMultiplier)
	}
}

func TestTerrainCompanionHealthRegenSystem_SetGenre(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainCompanionHealthRegenSystem(world, 12345)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		system.SetGenre(genre)
		if system.genreID != genre {
			t.Errorf("expected genre %s, got %s", genre, system.genreID)
		}
	}
}

func TestTerrainCompanionHealthRegenSystem_SetTileSize(t *testing.T) {
	system := NewTerrainCompanionHealthRegenSystem(nil, 12345)

	system.SetTileSize(64)
	if system.tileSize != 64 {
		t.Errorf("expected tileSize 64, got %d", system.tileSize)
	}

	// Invalid sizes should not change
	system.SetTileSize(0)
	if system.tileSize != 64 {
		t.Errorf("expected tileSize to remain 64, got %d", system.tileSize)
	}

	system.SetTileSize(-1)
	if system.tileSize != 64 {
		t.Errorf("expected tileSize to remain 64, got %d", system.tileSize)
	}
}

func TestTerrainCompanionHealthRegenSystem_CalculateTerrainRegen(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainCompanionHealthRegenSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create test terrain
	terr := terrain.NewTerrain(10, 10)
	terr.SetTile(1, 1, terrain.TileWaterShallow)
	terr.SetTile(2, 2, terrain.TileTree)
	terr.SetTile(3, 3, terrain.TilePlatform)
	terr.SetTile(4, 4, terrain.TilePit)
	system.SetTerrain(terr)

	tests := []struct {
		name          string
		tileX, tileY  int
		companionType CompanionType
		hasPerk       bool
		expectRegen   bool
		minRegen      float64
	}{
		{"water_elemental_in_water", 1, 1, CompanionTypeElemental, false, true, 2.0},
		{"spirit_in_water", 1, 1, CompanionTypeSpirit, false, true, 1.0},
		{"robot_in_water_no_regen", 1, 1, CompanionTypeRobot, false, false, 0},
		{"insect_in_tree", 2, 2, CompanionTypeInsect, false, true, 2.5},
		{"pet_in_tree", 2, 2, CompanionTypePet, false, true, 1.5},
		{"spirit_on_platform", 3, 3, CompanionTypeSpirit, false, true, 1.5},
		{"undead_in_pit", 4, 4, CompanionTypeUndead, false, true, 2.5},
		{"elemental_with_perk", 1, 1, CompanionTypeElemental, true, true, 3.5}, // 2.5 * 1.5
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			companion := &CompanionComponent{
				CompanionType: tt.companionType,
				BondingPerks:  []BondingPerk{},
			}
			if tt.hasPerk {
				companion.BondingPerks = []BondingPerk{PerkExtraHealth}
			}

			regen := system.calculateTerrainRegen(tt.tileX, tt.tileY, companion)

			if tt.expectRegen && regen < tt.minRegen {
				t.Errorf("expected regen >= %f, got %f", tt.minRegen, regen)
			}
			if !tt.expectRegen && regen > 0 {
				t.Errorf("expected no regen, got %f", regen)
			}
		})
	}
}

func TestTerrainCompanionHealthRegenSystem_GenreMultipliers(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainCompanionHealthRegenSystem(world, 12345)

	terr := terrain.NewTerrain(10, 10)
	terr.SetTile(1, 1, terrain.TileWaterShallow)
	system.SetTerrain(terr)

	companion := &CompanionComponent{
		CompanionType: CompanionTypeElemental,
	}

	tests := []struct {
		genre      string
		multiplier float64
	}{
		{"fantasy", 1.0},
		{"scifi", 0.8},
		{"horror", 1.3},
		{"cyberpunk", 0.7},
		{"postapoc", 1.1},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			system.SetGenre(tt.genre)
			regen := system.calculateTerrainRegen(1, 1, companion)

			// Base regen for elemental in water is 2.5
			expectedRegen := 2.5 * tt.multiplier
			if regen < expectedRegen*0.99 || regen > expectedRegen*1.01 {
				t.Errorf("expected regen ~%f, got %f", expectedRegen, regen)
			}
		})
	}
}

func TestTerrainCompanionHealthRegenSystem_Update(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainCompanionHealthRegenSystem(world, 12345)
	system.SetGenre("fantasy")

	terr := terrain.NewTerrain(10, 10)
	terr.SetTile(1, 1, terrain.TileTree)
	system.SetTerrain(terr)

	// Create companion entity on tree terrain
	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 40, Y: 40}) // Tile 1,1
	companion.AddComponent(&HealthComponent{Current: 50, Max: 100})
	companion.AddComponent(&CompanionComponent{
		CompanionType: CompanionTypeInsect,
		OwnerID:       1,
	})

	entities := []*Entity{companion}

	// Run update for 1 second (accumulated over multiple calls)
	for i := 0; i < 10; i++ {
		system.Update(entities, 0.1)
	}

	health := companion.GetHealth()
	if health.Current <= 50 {
		t.Error("expected health to increase from terrain regen")
	}
}

func TestTerrainCompanionHealthRegenSystem_NoRegenAtMaxHealth(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainCompanionHealthRegenSystem(world, 12345)
	system.SetGenre("fantasy")

	terr := terrain.NewTerrain(10, 10)
	terr.SetTile(1, 1, terrain.TileTree)
	system.SetTerrain(terr)

	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 40, Y: 40})
	companion.AddComponent(&HealthComponent{Current: 100, Max: 100}) // Already max
	companion.AddComponent(&CompanionComponent{
		CompanionType: CompanionTypeInsect,
	})

	entities := []*Entity{companion}

	// Should not exceed max health
	for i := 0; i < 10; i++ {
		system.Update(entities, 0.1)
	}

	health := companion.GetHealth()
	if health.Current > health.Max {
		t.Errorf("health %f exceeded max %f", health.Current, health.Max)
	}
}

func TestTerrainCompanionHealthRegenSystem_NilTerrain(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainCompanionHealthRegenSystem(world, 12345)
	// No terrain set

	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 40, Y: 40})
	companion.AddComponent(&HealthComponent{Current: 50, Max: 100})
	companion.AddComponent(&CompanionComponent{
		CompanionType: CompanionTypeInsect,
	})

	// Should not panic
	system.Update([]*Entity{companion}, 1.0)
}

func TestTerrainCompanionHealthRegenSystem_PerkExtraHealth(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainCompanionHealthRegenSystem(world, 12345)
	system.SetGenre("fantasy")

	terr := terrain.NewTerrain(10, 10)
	terr.SetTile(1, 1, terrain.TileWaterShallow)
	system.SetTerrain(terr)

	// Companion without perk
	companionNoPerk := &CompanionComponent{
		CompanionType: CompanionTypeElemental,
		BondingPerks:  []BondingPerk{},
	}

	// Companion with PerkExtraHealth
	companionWithPerk := &CompanionComponent{
		CompanionType: CompanionTypeElemental,
		BondingPerks:  []BondingPerk{PerkExtraHealth},
	}

	regenNoPerk := system.calculateTerrainRegen(1, 1, companionNoPerk)
	regenWithPerk := system.calculateTerrainRegen(1, 1, companionWithPerk)

	// With perk should be 50% higher
	expectedRatio := 1.5
	actualRatio := regenWithPerk / regenNoPerk

	if actualRatio < expectedRatio*0.99 || actualRatio > expectedRatio*1.01 {
		t.Errorf("expected perk ratio %f, got %f", expectedRatio, actualRatio)
	}
}

func TestTerrainCompanionHealthRegenSystem_GetRegenRate(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainCompanionHealthRegenSystem(world, 12345)
	system.SetGenre("fantasy")

	terr := terrain.NewTerrain(10, 10)
	terr.SetTile(1, 1, terrain.TileTree)
	system.SetTerrain(terr)

	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 40, Y: 40})
	companion.AddComponent(&CompanionComponent{
		CompanionType: CompanionTypeInsect,
	})

	rate := system.GetRegenRate(companion.ID)
	if rate <= 0 {
		t.Error("expected positive regen rate for insect on tree")
	}

	// Non-existent entity should return 0
	rate = system.GetRegenRate(99999)
	if rate != 0 {
		t.Errorf("expected 0 for non-existent entity, got %f", rate)
	}
}

func TestTerrainCompanionHealthRegenSystem_CompanionTypeName(t *testing.T) {
	system := NewTerrainCompanionHealthRegenSystem(nil, 12345)

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
			name := system.companionTypeName(tt.compType)
			if name != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, name)
			}
		})
	}
}

func TestTerrainCompanionHealthRegenSystem_Determinism(t *testing.T) {
	// Two systems with same seed should produce same results
	world1 := NewWorld(nil)
	system1 := NewTerrainCompanionHealthRegenSystem(world1, 12345)
	system1.SetGenre("fantasy")

	world2 := NewWorld(nil)
	system2 := NewTerrainCompanionHealthRegenSystem(world2, 12345)
	system2.SetGenre("fantasy")

	terr := terrain.NewTerrain(10, 10)
	terr.SetTile(1, 1, terrain.TileWaterShallow)
	system1.SetTerrain(terr)
	system2.SetTerrain(terr)

	companion := &CompanionComponent{
		CompanionType: CompanionTypeElemental,
	}

	regen1 := system1.calculateTerrainRegen(1, 1, companion)
	regen2 := system2.calculateTerrainRegen(1, 1, companion)

	if regen1 != regen2 {
		t.Errorf("determinism failed: %f != %f", regen1, regen2)
	}
}

func BenchmarkTerrainCompanionHealthRegenSystem_Update(b *testing.B) {
	world := NewWorld(nil)
	system := NewTerrainCompanionHealthRegenSystem(world, 12345)
	system.SetGenre("fantasy")

	terr := terrain.NewTerrain(100, 100)
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			terr.SetTile(x, y, terrain.TileTree)
		}
	}
	system.SetTerrain(terr)

	// Create 100 companion entities
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		e := world.CreateEntity()
		e.AddComponent(&PositionComponent{X: float64(i * 32), Y: float64(i * 32)})
		e.AddComponent(&HealthComponent{Current: 50, Max: 100})
		e.AddComponent(&CompanionComponent{
			CompanionType: CompanionTypeInsect,
		})
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}
