package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

func TestNewTerrainFishingBonusSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainFishingBonusSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewTerrainFishingBonusSystem returned nil")
	}

	if sys.tileSize != 32 {
		t.Errorf("Expected default tileSize 32, got %d", sys.tileSize)
	}

	if sys.genreID != "fantasy" {
		t.Errorf("Expected default genre 'fantasy', got %s", sys.genreID)
	}
}

func TestTerrainFishingBonusSystem_SetTerrain(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainFishingBonusSystem(world, 12345)

	// Pre-populate cache to verify clearing
	sys.bonusCache[1] = &TerrainFishingBonusData{}
	sys.lastTileCache[1] = fishingTilePos{tileX: 1, tileY: 1}

	terr := &terrain.Terrain{
		Tiles: [][]terrain.TileType{
			{terrain.TileWaterDeep, terrain.TileWaterShallow},
		},
	}
	sys.SetTerrain(terr)

	if sys.terrain != terr {
		t.Error("SetTerrain did not set terrain")
	}

	if len(sys.bonusCache) != 0 {
		t.Error("SetTerrain did not clear bonus cache")
	}

	if len(sys.lastTileCache) != 0 {
		t.Error("SetTerrain did not clear lastTile cache")
	}
}

func TestTerrainFishingBonusSystem_SetTileSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		expected int
	}{
		{"valid size", 64, 64},
		{"zero size", 0, 32},      // Should not change
		{"negative size", -1, 32}, // Should not change
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewTerrainFishingBonusSystem(world, 12345)
			sys.SetTileSize(tt.size)
			if sys.tileSize != tt.expected {
				t.Errorf("SetTileSize(%d) = %d, expected %d", tt.size, sys.tileSize, tt.expected)
			}
		})
	}
}

func TestTerrainFishingBonusSystem_SetGenre(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewTerrainFishingBonusSystem(world, 12345)
			sys.SetGenre(genre)
			if sys.genreID != genre {
				t.Errorf("SetGenre(%s) did not set genre, got %s", genre, sys.genreID)
			}
		})
	}
}

func TestTerrainFishingBonusSystem_UpdateNoTerrain(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainFishingBonusSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	entity.AddComponent(&FishingSpotComponent{RareFishBonus: 1.0})

	// Should not panic with nil terrain
	sys.Update([]*Entity{entity}, 0.016)
}

func TestTerrainFishingBonusSystem_UpdateNonFishingEntity(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainFishingBonusSystem(world, 12345)
	sys.SetTerrain(&terrain.Terrain{
		Tiles: [][]terrain.TileType{
			{terrain.TileWaterDeep, terrain.TileWaterDeep},
			{terrain.TileWaterDeep, terrain.TileWaterDeep},
		},
	})

	// Entity without fishing_spot component
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 16, Y: 16})

	sys.Update([]*Entity{entity}, 0.016)

	// Should not have bonus component
	if _, exists := entity.GetComponent("terrain_fishing_bonus"); exists {
		t.Error("Non-fishing entity should not have terrain_fishing_bonus component")
	}
}

func TestTerrainFishingBonusSystem_DeepWaterBonus(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainFishingBonusSystem(world, 12345)
	sys.SetTileSize(32)

	// Create terrain with deep water surrounding center
	tiles := make([][]terrain.TileType, 5)
	for i := range tiles {
		tiles[i] = make([]terrain.TileType, 5)
		for j := range tiles[i] {
			tiles[i][j] = terrain.TileWaterDeep
		}
	}
	tiles[2][2] = terrain.TileWaterShallow // Fishing spot location

	sys.SetTerrain(&terrain.Terrain{Tiles: tiles})

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 64, Y: 64}) // Tile 2,2
	entity.AddComponent(&FishingSpotComponent{RareFishBonus: 1.0})

	sys.Update([]*Entity{entity}, 0.016)

	comp, exists := entity.GetComponent("terrain_fishing_bonus")
	if !exists {
		t.Fatal("Expected terrain_fishing_bonus component")
	}

	bonus := comp.(*TerrainFishingBonusComponent)
	// 8 adjacent deep water tiles * 0.10 = 0.80, but capped at 0.40
	// So total should be 1.0 + 0.40 = 1.40
	if bonus.RareFishMultiplier < 1.3 {
		t.Errorf("Expected deep water bonus >= 1.3, got %f", bonus.RareFishMultiplier)
	}
}

func TestTerrainFishingBonusSystem_TreeBonus(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainFishingBonusSystem(world, 12345)
	sys.SetTileSize(32)

	// Create terrain with trees nearby
	tiles := make([][]terrain.TileType, 5)
	for i := range tiles {
		tiles[i] = make([]terrain.TileType, 5)
		for j := range tiles[i] {
			tiles[i][j] = terrain.TileWaterShallow
		}
	}
	tiles[0][0] = terrain.TileTree // Tree within 2 tiles of center

	sys.SetTerrain(&terrain.Terrain{Tiles: tiles})

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 64, Y: 64}) // Tile 2,2
	entity.AddComponent(&FishingSpotComponent{RareFishBonus: 1.0})

	sys.Update([]*Entity{entity}, 0.016)

	comp, exists := entity.GetComponent("terrain_fishing_bonus")
	if !exists {
		t.Fatal("Expected terrain_fishing_bonus component")
	}

	bonus := comp.(*TerrainFishingBonusComponent)
	// Tree bonus: +0.15, base 1.0
	if bonus.RareFishMultiplier < 1.1 {
		t.Errorf("Expected tree bonus >= 1.1, got %f", bonus.RareFishMultiplier)
	}

	if bonus.TerrainFeature == "" || bonus.TerrainFeature != "kelp" {
		t.Errorf("Expected terrain feature 'kelp', got '%s'", bonus.TerrainFeature)
	}
}

func TestTerrainFishingBonusSystem_BridgeBonus(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainFishingBonusSystem(world, 12345)
	sys.SetTileSize(32)

	// Create terrain with bridge at center
	tiles := make([][]terrain.TileType, 3)
	for i := range tiles {
		tiles[i] = make([]terrain.TileType, 3)
		for j := range tiles[i] {
			tiles[i][j] = terrain.TileWaterShallow
		}
	}
	tiles[1][1] = terrain.TileBridge

	sys.SetTerrain(&terrain.Terrain{Tiles: tiles})

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 32, Y: 32}) // Tile 1,1
	entity.AddComponent(&FishingSpotComponent{RareFishBonus: 1.0})

	sys.Update([]*Entity{entity}, 0.016)

	comp, exists := entity.GetComponent("terrain_fishing_bonus")
	if !exists {
		t.Fatal("Expected terrain_fishing_bonus component")
	}

	bonus := comp.(*TerrainFishingBonusComponent)
	// Bridge bonus: +0.20 catch speed
	if bonus.CatchSpeedMultiplier < 1.15 {
		t.Errorf("Expected bridge catch speed bonus >= 1.15, got %f", bonus.CatchSpeedMultiplier)
	}
}

func TestTerrainFishingBonusSystem_GenreModifiers(t *testing.T) {
	tests := []struct {
		name            string
		genre           string
		setupTerrain    func() [][]terrain.TileType
		minRareBonus    float64
		maxRareBonus    float64
		checkSpeedBonus bool
		minSpeedBonus   float64
	}{
		{
			name:  "fantasy_kelp",
			genre: "fantasy",
			setupTerrain: func() [][]terrain.TileType {
				tiles := make([][]terrain.TileType, 5)
				for i := range tiles {
					tiles[i] = make([]terrain.TileType, 5)
					for j := range tiles[i] {
						tiles[i][j] = terrain.TileWaterShallow
					}
				}
				tiles[0][0] = terrain.TileTree
				return tiles
			},
			minRareBonus: 1.2, // +0.15 base + 0.10 fantasy bonus
		},
		{
			name:  "horror_deep",
			genre: "horror",
			setupTerrain: func() [][]terrain.TileType {
				tiles := make([][]terrain.TileType, 3)
				for i := range tiles {
					tiles[i] = make([]terrain.TileType, 3)
					for j := range tiles[i] {
						tiles[i][j] = terrain.TileWaterDeep
					}
				}
				tiles[1][1] = terrain.TileWaterShallow
				return tiles
			},
			minRareBonus: 1.5, // Deep water + horror bonus
		},
		{
			name:  "cyberpunk_bridge",
			genre: "cyberpunk",
			setupTerrain: func() [][]terrain.TileType {
				tiles := make([][]terrain.TileType, 3)
				for i := range tiles {
					tiles[i] = make([]terrain.TileType, 3)
					for j := range tiles[i] {
						tiles[i][j] = terrain.TileWaterShallow
					}
				}
				tiles[1][1] = terrain.TileBridge
				return tiles
			},
			checkSpeedBonus: true,
			minSpeedBonus:   1.25, // +0.20 base + 0.10 cyberpunk
		},
		{
			name:  "postapoc_penalty",
			genre: "postapoc",
			setupTerrain: func() [][]terrain.TileType {
				tiles := make([][]terrain.TileType, 5)
				for i := range tiles {
					tiles[i] = make([]terrain.TileType, 5)
					for j := range tiles[i] {
						tiles[i][j] = terrain.TileWaterDeep
					}
				}
				tiles[2][2] = terrain.TileWaterShallow
				return tiles
			},
			minRareBonus: 1.0, // Should be lower due to -10% penalty
			maxRareBonus: 1.4, // Capped deep water bonus * 0.90
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewTerrainFishingBonusSystem(world, 12345)
			sys.SetGenre(tt.genre)
			sys.SetTileSize(32)
			sys.SetTerrain(&terrain.Terrain{Tiles: tt.setupTerrain()})

			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: 64, Y: 64})
			entity.AddComponent(&FishingSpotComponent{RareFishBonus: 1.0})

			sys.Update([]*Entity{entity}, 0.016)

			comp, exists := entity.GetComponent("terrain_fishing_bonus")
			if !exists {
				t.Fatal("Expected terrain_fishing_bonus component")
			}

			bonus := comp.(*TerrainFishingBonusComponent)

			if tt.minRareBonus > 0 && bonus.RareFishMultiplier < tt.minRareBonus {
				t.Errorf("Genre %s: expected rare bonus >= %f, got %f", tt.genre, tt.minRareBonus, bonus.RareFishMultiplier)
			}

			if tt.maxRareBonus > 0 && bonus.RareFishMultiplier > tt.maxRareBonus {
				t.Errorf("Genre %s: expected rare bonus <= %f, got %f", tt.genre, tt.maxRareBonus, bonus.RareFishMultiplier)
			}

			if tt.checkSpeedBonus && bonus.CatchSpeedMultiplier < tt.minSpeedBonus {
				t.Errorf("Genre %s: expected speed bonus >= %f, got %f", tt.genre, tt.minSpeedBonus, bonus.CatchSpeedMultiplier)
			}
		})
	}
}

func TestTerrainFishingBonusSystem_CacheInvalidation(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainFishingBonusSystem(world, 12345)
	sys.SetTileSize(32)

	tiles := make([][]terrain.TileType, 4)
	for i := range tiles {
		tiles[i] = make([]terrain.TileType, 4)
		for j := range tiles[i] {
			tiles[i][j] = terrain.TileWaterShallow
		}
	}
	tiles[0][0] = terrain.TileWaterDeep // Different bonus at 0,0
	tiles[2][2] = terrain.TileTree      // Different bonus at 2,2

	sys.SetTerrain(&terrain.Terrain{Tiles: tiles})

	entity := world.CreateEntity()
	pos := &PositionComponent{X: 16, Y: 16} // Tile 0,0
	entity.AddComponent(pos)
	entity.AddComponent(&FishingSpotComponent{RareFishBonus: 1.0})

	// First update at tile 0,0
	sys.Update([]*Entity{entity}, 0.016)

	bonus1, exists := sys.GetTerrainBonus(entity.ID)
	if !exists {
		t.Fatal("Expected cached bonus after first update")
	}

	// Move entity to different tile
	pos.X = 64
	pos.Y = 64 // Tile 2,2

	// Update should recalculate
	sys.Update([]*Entity{entity}, 0.016)

	bonus2, exists := sys.GetTerrainBonus(entity.ID)
	if !exists {
		t.Fatal("Expected cached bonus after movement")
	}

	// Bonuses should be different (one has deep water, other has tree)
	if bonus1.RareFishBonus == bonus2.RareFishBonus && bonus1.TerrainFeatures == bonus2.TerrainFeatures {
		t.Error("Cache should have been invalidated after tile change")
	}
}

func TestTerrainFishingBonusSystem_RemoveBonus(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainFishingBonusSystem(world, 12345)
	sys.SetTileSize(32)

	tiles := make([][]terrain.TileType, 3)
	for i := range tiles {
		tiles[i] = make([]terrain.TileType, 3)
		for j := range tiles[i] {
			tiles[i][j] = terrain.TileWaterShallow
		}
	}
	sys.SetTerrain(&terrain.Terrain{Tiles: tiles})

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 32, Y: 32})
	entity.AddComponent(&FishingSpotComponent{RareFishBonus: 1.0})

	// Add bonus
	sys.Update([]*Entity{entity}, 0.016)

	if _, exists := entity.GetComponent("terrain_fishing_bonus"); !exists {
		t.Fatal("Expected terrain_fishing_bonus component after update")
	}

	// Remove fishing_spot component
	entity.RemoveComponent("fishing_spot")

	// Update should remove bonus
	sys.Update([]*Entity{entity}, 0.016)

	if _, exists := entity.GetComponent("terrain_fishing_bonus"); exists {
		t.Error("Expected terrain_fishing_bonus to be removed when fishing_spot removed")
	}

	if _, exists := sys.GetTerrainBonus(entity.ID); exists {
		t.Error("Expected bonus cache to be cleared")
	}
}

func TestTerrainFishingBonusSystem_NoPosition(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainFishingBonusSystem(world, 12345)
	sys.SetTerrain(&terrain.Terrain{
		Tiles: [][]terrain.TileType{{terrain.TileWaterShallow}},
	})

	// Entity with fishing spot but no position
	entity := world.CreateEntity()
	entity.AddComponent(&FishingSpotComponent{RareFishBonus: 1.0})

	sys.Update([]*Entity{entity}, 0.016)

	// Should not have bonus without position
	if _, exists := entity.GetComponent("terrain_fishing_bonus"); exists {
		t.Error("Entity without position should not have terrain_fishing_bonus")
	}
}

func TestTerrainFishingBonusComponent_Type(t *testing.T) {
	comp := &TerrainFishingBonusComponent{}
	if comp.Type() != "terrain_fishing_bonus" {
		t.Errorf("Expected type 'terrain_fishing_bonus', got '%s'", comp.Type())
	}
}

func TestTerrainFishingBonusSystem_StructureBonus(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainFishingBonusSystem(world, 12345)
	sys.SetTileSize(32)

	// Create terrain with structure nearby
	tiles := make([][]terrain.TileType, 5)
	for i := range tiles {
		tiles[i] = make([]terrain.TileType, 5)
		for j := range tiles[i] {
			tiles[i][j] = terrain.TileWaterShallow
		}
	}
	tiles[0][0] = terrain.TileStructure // Structure within 2 tiles

	sys.SetTerrain(&terrain.Terrain{Tiles: tiles})

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 64, Y: 64})
	entity.AddComponent(&FishingSpotComponent{RareFishBonus: 1.0})

	sys.Update([]*Entity{entity}, 0.016)

	comp, exists := entity.GetComponent("terrain_fishing_bonus")
	if !exists {
		t.Fatal("Expected terrain_fishing_bonus component")
	}

	bonus := comp.(*TerrainFishingBonusComponent)
	// Structure bonus: +0.10
	if bonus.RareFishMultiplier < 1.05 {
		t.Errorf("Expected structure bonus >= 1.05, got %f", bonus.RareFishMultiplier)
	}
}

func TestTerrainFishingBonusSystem_OutOfBounds(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainFishingBonusSystem(world, 12345)
	sys.SetTileSize(32)

	// Small terrain
	tiles := [][]terrain.TileType{
		{terrain.TileWaterShallow},
	}
	sys.SetTerrain(&terrain.Terrain{Tiles: tiles})

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 1000, Y: 1000}) // Way out of bounds
	entity.AddComponent(&FishingSpotComponent{RareFishBonus: 1.0})

	// Should not panic
	sys.Update([]*Entity{entity}, 0.016)
}

func BenchmarkTerrainFishingBonusSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewTerrainFishingBonusSystem(world, 12345)
	sys.SetTileSize(32)

	// Large terrain
	tiles := make([][]terrain.TileType, 100)
	for i := range tiles {
		tiles[i] = make([]terrain.TileType, 100)
		for j := range tiles[i] {
			if (i+j)%3 == 0 {
				tiles[i][j] = terrain.TileWaterDeep
			} else if (i+j)%5 == 0 {
				tiles[i][j] = terrain.TileTree
			} else {
				tiles[i][j] = terrain.TileWaterShallow
			}
		}
	}
	sys.SetTerrain(&terrain.Terrain{Tiles: tiles})

	// Create 50 fishing spots
	entities := make([]*Entity, 50)
	for i := 0; i < 50; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i*64 + 32), Y: float64(i*32 + 16)})
		entity.AddComponent(&FishingSpotComponent{RareFishBonus: 1.0})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Move entities to force recalculation
		for _, e := range entities {
			pos := e.GetPosition()
			pos.X += 32
			if pos.X > 3000 {
				pos.X = 32
			}
		}
		sys.Update(entities, 0.016)
	}
}
