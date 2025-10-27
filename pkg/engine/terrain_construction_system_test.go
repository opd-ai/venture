package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/world"
)

// TestNewTerrainConstructionSystem verifies system creation.
func TestNewTerrainConstructionSystem(t *testing.T) {
	system := NewTerrainConstructionSystem(32)

	if system == nil {
		t.Fatal("expected system, got nil")
	}
	if system.tileSize != 32 {
		t.Errorf("expected tileSize=32, got %d", system.tileSize)
	}
}

// TestTerrainConstructionSystem_SetReferences verifies setting references.
func TestTerrainConstructionSystem_SetReferences(t *testing.T) {
	system := NewTerrainConstructionSystem(32)
	w := NewWorld()
	terr := createTestTerrain(10, 10)
	worldMap := &world.Map{}

	system.SetWorld(w)
	system.SetTerrain(terr)
	system.SetWorldMap(worldMap)

	if system.world != w {
		t.Error("expected world reference to be set")
	}
	if system.terrain != terr {
		t.Error("expected terrain reference to be set")
	}
	if system.worldMap != worldMap {
		t.Error("expected worldMap reference to be set")
	}
}

// TestTerrainConstructionSystem_ValidatePlacement verifies placement validation.
func TestTerrainConstructionSystem_ValidatePlacement(t *testing.T) {
	tests := []struct {
		name     string
		tileX    int
		tileY    int
		tileType terrain.TileType
		wantErr  bool
	}{
		{"valid floor", 5, 5, terrain.TileFloor, false},
		{"invalid wall", 5, 5, terrain.TileWall, true},
		{"out of bounds negative", -1, 5, terrain.TileFloor, true},
		{"out of bounds large", 100, 100, terrain.TileFloor, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system := NewTerrainConstructionSystem(32)
			terr := createTestTerrain(10, 10)
			system.SetTerrain(terr)

			if tt.tileX >= 0 && tt.tileX < 10 && tt.tileY >= 0 && tt.tileY < 10 {
				terr.SetTile(tt.tileX, tt.tileY, tt.tileType)
			}

			err := system.validatePlacement(tt.tileX, tt.tileY)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePlacement() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestTerrainConstructionSystem_StartConstruction_Success verifies successful construction.
func TestTerrainConstructionSystem_StartConstruction_Success(t *testing.T) {
	system := NewTerrainConstructionSystem(32)
	w := NewWorld()
	terr := createTestTerrain(10, 10)
	worldMap := world.NewMap(10, 10, 12345)
	system.SetWorld(w)
	system.SetTerrain(terr)
	system.SetWorldMap(worldMap)

	// Setup tile as floor
	terr.SetTile(5, 5, terrain.TileFloor)

	// Create builder with materials
	builder := w.CreateEntity()
	inv := NewInventoryComponent(20, 100.0)
	inv.Gold = 100 // Enough for 10 stone equivalent
	builder.AddComponent(inv)
	w.Update(0)

	// Start construction
	err := system.StartConstruction(builder, 5, 5, world.TileWall)
	if err != nil {
		t.Fatalf("StartConstruction() unexpected error: %v", err)
	}

	// Verify buildable entity created
	w.Update(0)
	entities := w.GetEntities()
	found := false
	for _, e := range entities {
		if _, ok := e.GetComponent("buildable"); ok {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected buildable entity to be created")
	}

	// Verify materials consumed
	invComp, _ := builder.GetComponent("inventory")
	inv2 := invComp.(*InventoryComponent)
	if inv2.Gold >= 100 {
		t.Errorf("expected gold to be consumed, got %d", inv2.Gold)
	}
}

// TestTerrainConstructionSystem_StartConstruction_InvalidPlacement verifies placement errors.
func TestTerrainConstructionSystem_StartConstruction_InvalidPlacement(t *testing.T) {
	system := NewTerrainConstructionSystem(32)
	w := NewWorld()
	terr := createTestTerrain(10, 10)
	worldMap := world.NewMap(10, 10, 12345)
	system.SetWorld(w)
	system.SetTerrain(terr)
	system.SetWorldMap(worldMap)

	// Tile is wall (not walkable)
	terr.SetTile(5, 5, terrain.TileWall)

	builder := w.CreateEntity()
	inv := NewInventoryComponent(20, 100.0)
	inv.Gold = 100
	builder.AddComponent(inv)

	err := system.StartConstruction(builder, 5, 5, world.TileWall)
	if err == nil {
		t.Error("expected error for invalid placement")
	}
}

// TestTerrainConstructionSystem_StartConstruction_InsufficientMaterials verifies material check.
func TestTerrainConstructionSystem_StartConstruction_InsufficientMaterials(t *testing.T) {
	system := NewTerrainConstructionSystem(32)
	w := NewWorld()
	terr := createTestTerrain(10, 10)
	worldMap := world.NewMap(10, 10, 12345)
	system.SetWorld(w)
	system.SetTerrain(terr)
	system.SetWorldMap(worldMap)

	terr.SetTile(5, 5, terrain.TileFloor)

	builder := w.CreateEntity()
	inv := NewInventoryComponent(20, 100.0)
	inv.Gold = 50 // Not enough (need 100 for 10 stone)
	builder.AddComponent(inv)

	err := system.StartConstruction(builder, 5, 5, world.TileWall)
	if err == nil {
		t.Error("expected error for insufficient materials")
	}
}

// TestTerrainConstructionSystem_Update_CompleteConstruction verifies completion.
func TestTerrainConstructionSystem_Update_CompleteConstruction(t *testing.T) {
	system := NewTerrainConstructionSystem(32)
	w := NewWorld()
	terr := createTestTerrain(10, 10)

	// Initialize worldMap with tiles using NewMap
	worldMap := world.NewMap(10, 10, 12345)

	system.SetWorld(w)
	system.SetTerrain(terr)
	system.SetWorldMap(worldMap)

	// Set initial terrain tile
	terr.SetTile(5, 5, terrain.TileFloor)

	// Create buildable entity
	entity := w.CreateEntity()
	buildComp := NewBuildableComponent(5, 5, world.TileWall, 1.0)
	entity.AddComponent(buildComp)
	entity.AddComponent(&PositionComponent{X: 160, Y: 160})
	w.Update(0)

	// Update past completion time
	entities := w.GetEntities()
	system.Update(entities, 1.5)

	// Verify construction completed
	if terr.GetTile(5, 5) != terrain.TileWall {
		t.Errorf("expected tile to be wall, got %v", terr.GetTile(5, 5))
	}

	// Verify buildable entity removed
	w.Update(0)
	entities = w.GetEntities()
	for _, e := range entities {
		if _, ok := e.GetComponent("buildable"); ok {
			t.Error("buildable entity should be removed after completion")
		}
	}
}

// TestTerrainConstructionSystem_GetConstructionProgress verifies progress tracking.
func TestTerrainConstructionSystem_GetConstructionProgress(t *testing.T) {
	system := NewTerrainConstructionSystem(32)
	w := NewWorld()
	system.SetWorld(w)

	// No construction
	progress := system.GetConstructionProgress(5, 5)
	if progress != 0.0 {
		t.Errorf("expected 0.0 progress, got %f", progress)
	}

	// Create buildable entity
	entity := w.CreateEntity()
	buildComp := NewBuildableComponent(5, 5, world.TileWall, 10.0)
	buildComp.ElapsedTime = 5.0 // 50% complete
	entity.AddComponent(buildComp)
	entity.AddComponent(&PositionComponent{X: 160, Y: 160})
	w.Update(0)

	progress = system.GetConstructionProgress(5, 5)
	if progress < 0.49 || progress > 0.51 {
		t.Errorf("expected ~0.5 progress, got %f", progress)
	}
}

// TestTerrainConstructionSystem_CountMaterialsInInventory verifies material counting.
func TestTerrainConstructionSystem_CountMaterialsInInventory(t *testing.T) {
	system := NewTerrainConstructionSystem(32)

	tests := []struct {
		name      string
		items     []*item.Item
		gold      int
		wantStone int
		wantWood  int
	}{
		{
			name:      "gold only",
			items:     []*item.Item{},
			gold:      100,
			wantStone: 10, // 100 gold / 10 = 10 stone
			wantWood:  0,
		},
		{
			name: "stone items",
			items: []*item.Item{
				{Name: "Stone Block"},
				{Name: "Stone Brick"},
			},
			gold:      0,
			wantStone: 2,
			wantWood:  0,
		},
		{
			name: "wood items",
			items: []*item.Item{
				{Name: "Wood Plank"},
				{Name: "Wood Log"},
			},
			gold:      0,
			wantStone: 0,
			wantWood:  2,
		},
		{
			name: "mixed",
			items: []*item.Item{
				{Name: "Stone Block"},
				{Name: "Wood Plank"},
			},
			gold:      50,
			wantStone: 6, // 1 item + 5 from gold
			wantWood:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := NewInventoryComponent(20, 100.0)
			inv.Items = tt.items
			inv.Gold = tt.gold

			counts := system.countMaterialsInInventory(inv)

			if counts[MaterialStone] != tt.wantStone {
				t.Errorf("stone count = %d, want %d", counts[MaterialStone], tt.wantStone)
			}
			if counts[MaterialWood] != tt.wantWood {
				t.Errorf("wood count = %d, want %d", counts[MaterialWood], tt.wantWood)
			}
		})
	}
}

// TestTerrainConstructionSystem_ConsumeMaterials verifies material consumption.
func TestTerrainConstructionSystem_ConsumeMaterials(t *testing.T) {
	system := NewTerrainConstructionSystem(32)

	inv := NewInventoryComponent(20, 100.0)
	inv.Items = []*item.Item{
		{Name: "Stone Block"},
		{Name: "Stone Brick"},
		{Name: "Wood Plank"},
	}
	inv.Gold = 100

	required := map[MaterialType]int{
		MaterialStone: 5, // Will consume 2 items + 30 gold
		MaterialWood:  1,
	}

	system.consumeMaterials(inv, required)

	// Check items consumed
	stoneCount := 0
	woodCount := 0
	for _, itm := range inv.Items {
		if len(itm.Name) > 5 && itm.Name[:5] == "Stone" {
			stoneCount++
		}
		if len(itm.Name) > 4 && itm.Name[:4] == "Wood" {
			woodCount++
		}
	}

	if stoneCount > 0 {
		t.Errorf("expected all stone items consumed, got %d remaining", stoneCount)
	}
	if woodCount > 0 {
		t.Errorf("expected all wood items consumed, got %d remaining", woodCount)
	}

	// Check gold consumed (should consume 30 for remaining 3 stone)
	if inv.Gold > 70 {
		t.Errorf("expected gold consumed, got %d (want <=70)", inv.Gold)
	}
}

// Benchmark construction validation.
func BenchmarkTerrainConstructionSystem_ValidatePlacement(b *testing.B) {
	system := NewTerrainConstructionSystem(32)
	terr := createTestTerrain(50, 50)
	system.SetTerrain(terr)

	// Set tiles as floor
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			terr.SetTile(x, y, terrain.TileFloor)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := i % 50
		y := (i / 50) % 50
		_ = system.validatePlacement(x, y)
	}
}

// Benchmark construction update.
func BenchmarkTerrainConstructionSystem_Update(b *testing.B) {
	system := NewTerrainConstructionSystem(32)
	w := NewWorld()
	terr := createTestTerrain(50, 50)
	worldMap := world.NewMap(50, 50, 12345)
	system.SetWorld(w)
	system.SetTerrain(terr)
	system.SetWorldMap(worldMap)

	// Create 10 buildable entities
	for i := 0; i < 10; i++ {
		entity := w.CreateEntity()
		buildComp := NewBuildableComponent(i, i, world.TileWall, 10.0)
		entity.AddComponent(buildComp)
		entity.AddComponent(&PositionComponent{X: float64(i * 32), Y: float64(i * 32)})
	}
	w.Update(0)

	entities := w.GetEntities()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016) // ~60 FPS
	}
}
