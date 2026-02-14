package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

func TestNewTerrainStatusEffectSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStatusEffectSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewTerrainStatusEffectSystem returned nil")
	}
	if sys.world != world {
		t.Error("world not set correctly")
	}
	if sys.tileSize != 32 {
		t.Errorf("tileSize = %d, want 32", sys.tileSize)
	}
	if sys.baseDuration != 3.0 {
		t.Errorf("baseDuration = %f, want 3.0", sys.baseDuration)
	}
}

func TestTerrainStatusEffectSystem_SetTerrain(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStatusEffectSystem(world, 12345)

	terr := terrain.NewTerrain(10, 10, 12345)
	sys.SetTerrain(terr)

	if sys.terrain != terr {
		t.Error("terrain not set correctly")
	}
}

func TestTerrainStatusEffectSystem_SetTileSize(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStatusEffectSystem(world, 12345)

	sys.SetTileSize(64)
	if sys.tileSize != 64 {
		t.Errorf("tileSize = %d, want 64", sys.tileSize)
	}

	// Invalid size should be ignored
	sys.SetTileSize(0)
	if sys.tileSize != 64 {
		t.Errorf("tileSize changed to %d, want 64 (unchanged)", sys.tileSize)
	}
}

func TestTerrainStatusEffectSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStatusEffectSystem(world, 12345)

	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("genreID = %s, want horror", sys.genreID)
	}
}

func TestTerrainStatusEffectSystem_Update_NoTerrain(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStatusEffectSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Should not panic without terrain
	sys.Update([]*Entity{entity}, 0.016)
}

func TestTerrainStatusEffectSystem_WaterAppliesWet(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStatusEffectSystem(world, 12345)
	sys.SetTileSize(32)

	// Create terrain with water tile
	terr := terrain.NewTerrain(10, 10, 12345)
	terr.SetTile(3, 3, terrain.TileWaterShallow)
	sys.SetTerrain(terr)

	// Create entity on water
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100}) // Tile 3,3
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	sys.Update([]*Entity{entity}, 0.016)

	// Check for wet status
	hasWet := false
	for _, comp := range entity.Components {
		if effect, ok := comp.(*StatusEffectComponent); ok {
			if effect.EffectType == "wet" {
				hasWet = true
				break
			}
		}
	}

	if !hasWet {
		t.Error("entity on water should have 'wet' status effect")
	}
}

func TestTerrainStatusEffectSystem_LavaAppliesBurning(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStatusEffectSystem(world, 12345)
	sys.SetTileSize(32)

	// Create terrain with lava tile
	terr := terrain.NewTerrain(10, 10, 12345)
	terr.SetTile(3, 3, terrain.TileLavaFlow)
	sys.SetTerrain(terr)

	// Create entity on lava
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100}) // Tile 3,3
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	sys.Update([]*Entity{entity}, 0.016)

	// Check for burning status
	hasBurning := false
	for _, comp := range entity.Components {
		if effect, ok := comp.(*StatusEffectComponent); ok {
			if effect.EffectType == "burning" {
				hasBurning = true
				if effect.TickInterval == 0 {
					t.Error("burning should have tick interval for DoT")
				}
				break
			}
		}
	}

	if !hasBurning {
		t.Error("entity on lava should have 'burning' status effect")
	}
}

func TestTerrainStatusEffectSystem_CooldownPreventsStacking(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStatusEffectSystem(world, 12345)
	sys.SetTileSize(32)

	// Create terrain with water
	terr := terrain.NewTerrain(10, 10, 12345)
	terr.SetTile(3, 3, terrain.TileWaterShallow)
	sys.SetTerrain(terr)

	// Create entity on water
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	// First update
	sys.Update([]*Entity{entity}, 0.016)

	// Count wet effects
	countWet := 0
	for _, comp := range entity.Components {
		if effect, ok := comp.(*StatusEffectComponent); ok {
			if effect.EffectType == "wet" {
				countWet++
			}
		}
	}

	// Second update should not add another wet effect due to cooldown
	sys.Update([]*Entity{entity}, 0.016)

	countWet2 := 0
	for _, comp := range entity.Components {
		if effect, ok := comp.(*StatusEffectComponent); ok {
			if effect.EffectType == "wet" {
				countWet2++
			}
		}
	}

	if countWet2 > countWet {
		t.Errorf("wet effects stacked: had %d, now have %d", countWet, countWet2)
	}
}

func TestTerrainStatusEffectSystem_SkipsDeadEntities(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStatusEffectSystem(world, 12345)
	sys.SetTileSize(32)

	terr := terrain.NewTerrain(10, 10, 12345)
	terr.SetTile(3, 3, terrain.TileWaterShallow)
	sys.SetTerrain(terr)

	// Create dead entity on water
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&HealthComponent{Current: 0, Max: 100})

	sys.Update([]*Entity{entity}, 0.016)

	// Should not have wet effect
	for _, comp := range entity.Components {
		if effect, ok := comp.(*StatusEffectComponent); ok {
			if effect.EffectType == "wet" {
				t.Error("dead entity should not receive terrain effects")
			}
		}
	}
}

func TestTerrainStatusEffectSystem_GenreMultipliers(t *testing.T) {
	tests := []struct {
		genre    string
		expected float64
	}{
		{"fantasy", 1.0},
		{"horror", 1.3},
		{"scifi", 0.8},
		{"cyberpunk", 0.9},
		{"postapoc", 1.2},
		{"unknown", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewTerrainStatusEffectSystem(world, 12345)
			sys.SetGenre(tt.genre)

			mult := sys.getGenreMultiplier()
			if mult != tt.expected {
				t.Errorf("genre %s: multiplier = %f, want %f", tt.genre, mult, tt.expected)
			}
		})
	}
}

func TestTerrainStatusEffectSystem_CooldownExpires(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStatusEffectSystem(world, 12345)
	sys.SetTileSize(32)
	sys.SetCooldownTime(0.5) // Short cooldown for testing

	terr := terrain.NewTerrain(10, 10, 12345)
	terr.SetTile(3, 3, terrain.TileWaterShallow)
	sys.SetTerrain(terr)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	// First update applies effect
	sys.Update([]*Entity{entity}, 0.016)

	// Remove effect to test reapplication
	var toRemove Component
	for _, comp := range entity.Components {
		if effect, ok := comp.(*StatusEffectComponent); ok {
			if effect.EffectType == "wet" {
				toRemove = comp
				break
			}
		}
	}
	if toRemove != nil {
		entity.RemoveComponent(toRemove.Type())
	}

	// Wait for cooldown to expire
	sys.Update([]*Entity{entity}, 0.6) // Exceeds cooldown

	// Now entity should be able to get wet again
	sys.Update([]*Entity{entity}, 0.016)

	hasWet := false
	for _, comp := range entity.Components {
		if effect, ok := comp.(*StatusEffectComponent); ok {
			if effect.EffectType == "wet" {
				hasWet = true
				break
			}
		}
	}

	if !hasWet {
		t.Error("effect should reapply after cooldown expires")
	}
}

func TestTerrainStatusEffectSystem_HorrorTrapDoorChilled(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStatusEffectSystem(world, 12345)
	sys.SetTileSize(32)
	sys.SetGenre("horror")

	terr := terrain.NewTerrain(10, 10, 12345)
	terr.SetTile(3, 3, terrain.TileTrapDoor)
	sys.SetTerrain(terr)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	sys.Update([]*Entity{entity}, 0.016)

	hasChilled := false
	for _, comp := range entity.Components {
		if effect, ok := comp.(*StatusEffectComponent); ok {
			if effect.EffectType == "chilled" {
				hasChilled = true
				break
			}
		}
	}

	if !hasChilled {
		t.Error("horror genre trap door should apply 'chilled' effect")
	}
}

func TestTerrainStatusEffectSystem_NoEffectOnNormalFloor(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStatusEffectSystem(world, 12345)
	sys.SetTileSize(32)

	terr := terrain.NewTerrain(10, 10, 12345)
	terr.SetTile(3, 3, terrain.TileFloor)
	sys.SetTerrain(terr)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	sys.Update([]*Entity{entity}, 0.016)

	// Should have no terrain-based effects
	for _, comp := range entity.Components {
		if _, ok := comp.(*StatusEffectComponent); ok {
			t.Error("normal floor should not apply any status effects")
		}
	}
}

func TestTerrainStatusEffectSystem_GetSetBaseDuration(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStatusEffectSystem(world, 12345)

	if sys.GetBaseDuration() != 3.0 {
		t.Errorf("GetBaseDuration = %f, want 3.0", sys.GetBaseDuration())
	}

	sys.SetBaseDuration(5.0)
	if sys.GetBaseDuration() != 5.0 {
		t.Errorf("GetBaseDuration = %f, want 5.0", sys.GetBaseDuration())
	}

	// Invalid duration ignored
	sys.SetBaseDuration(0)
	if sys.GetBaseDuration() != 5.0 {
		t.Errorf("GetBaseDuration = %f, want 5.0 (unchanged)", sys.GetBaseDuration())
	}
}

func TestTerrainStatusEffectSystem_GetSetCooldownTime(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainStatusEffectSystem(world, 12345)

	if sys.GetCooldownTime() != 1.5 {
		t.Errorf("GetCooldownTime = %f, want 1.5", sys.GetCooldownTime())
	}

	sys.SetCooldownTime(2.0)
	if sys.GetCooldownTime() != 2.0 {
		t.Errorf("GetCooldownTime = %f, want 2.0", sys.GetCooldownTime())
	}
}

func BenchmarkTerrainStatusEffectSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewTerrainStatusEffectSystem(world, 12345)
	sys.SetTileSize(32)

	terr := terrain.NewTerrain(100, 100, 12345)
	// Add some water tiles
	for i := 0; i < 20; i++ {
		terr.SetTile(i*5, i*5, terrain.TileWaterShallow)
	}
	sys.SetTerrain(terr)

	entities := make([]*Entity, 100)
	for i := range entities {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 32), Y: float64(i * 32)})
		entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
