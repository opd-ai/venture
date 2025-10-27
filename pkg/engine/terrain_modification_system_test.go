package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/world"
)

// TestNewTerrainModificationSystem verifies system creation.
func TestNewTerrainModificationSystem(t *testing.T) {
	system := NewTerrainModificationSystem(32)

	if system == nil {
		t.Fatal("expected system, got nil")
	}
	if system.tileSize != 32 {
		t.Errorf("expected tileSize=32, got %d", system.tileSize)
	}
}

// TestTerrainModificationSystem_SetReferences verifies setting references.
func TestTerrainModificationSystem_SetReferences(t *testing.T) {
	system := NewTerrainModificationSystem(32)
	w := NewWorld()
	terr := createTestTerrain(10, 10)
	worldMap := world.NewMap(10, 10, 12345)

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

// TestTerrainModificationSystem_CanWeaponDamageTerrain verifies weapon type checks.
func TestTerrainModificationSystem_CanWeaponDamageTerrain(t *testing.T) {
	tests := []struct {
		name   string
		weapon *item.Item
		want   bool
	}{
		{"nil weapon", nil, false},
		{"weapon type", &item.Item{Type: item.TypeWeapon}, true},
		{"armor type", &item.Item{Type: item.TypeArmor}, false},
		{"consumable type", &item.Item{Type: item.TypeConsumable}, false},
	}

	system := NewTerrainModificationSystem(32)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := system.canWeaponDamageTerrain(tt.weapon)
			if got != tt.want {
				t.Errorf("canWeaponDamageTerrain() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTerrainModificationSystem_GetWeaponTerrainDamage verifies damage calculation.
func TestTerrainModificationSystem_GetWeaponTerrainDamage(t *testing.T) {
	tests := []struct {
		name    string
		weapon  *item.Item
		wantMin float64
		wantMax float64
	}{
		{"nil weapon", nil, 0, 0},
		{"10 damage weapon", &item.Item{Stats: item.Stats{Damage: 10}}, 4.9, 5.1},
		{"20 damage weapon", &item.Item{Stats: item.Stats{Damage: 20}}, 9.9, 10.1},
		{"100 damage weapon", &item.Item{Stats: item.Stats{Damage: 100}}, 49.9, 50.1},
	}

	system := NewTerrainModificationSystem(32)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := system.getWeaponTerrainDamage(tt.weapon)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("getWeaponTerrainDamage() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// TestTerrainModificationSystem_GetAttackDirection verifies direction calculation.
func TestTerrainModificationSystem_GetAttackDirection(t *testing.T) {
	tests := []struct {
		name   string
		facing Direction
		wantX  int
		wantY  int
	}{
		{"no animation component", DirRight, 1, 0}, // default
		{"facing up", DirUp, 0, -1},
		{"facing down", DirDown, 0, 1},
		{"facing left", DirLeft, -1, 0},
		{"facing right", DirRight, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWorld()
			entity := w.CreateEntity()

			if tt.name != "no animation component" {
				animComp := &AnimationComponent{Facing: tt.facing}
				entity.AddComponent(animComp)
			}

			system := NewTerrainModificationSystem(32)
			dir := system.getAttackDirection(entity)

			if dir.X != tt.wantX || dir.Y != tt.wantY {
				t.Errorf("getAttackDirection() = {%d, %d}, want {%d, %d}", dir.X, dir.Y, tt.wantX, tt.wantY)
			}
		})
	}
}

// TestTerrainModificationSystem_DamageTile verifies tile damage.
func TestTerrainModificationSystem_DamageTile(t *testing.T) {
	system := NewTerrainModificationSystem(32)
	w := NewWorld()
	terr := createTestTerrain(10, 10)
	worldMap := world.NewMap(10, 10, 12345)

	system.SetWorld(w)
	system.SetTerrain(terr)
	system.SetWorldMap(worldMap)

	// Set tile as wall
	terr.SetTile(5, 5, terrain.TileWall)

	// Apply damage
	system.damageTile(5, 5, 50.0)

	// Verify destructible entity created
	w.Update(0)
	entities := w.GetEntities()
	found := false
	for _, e := range entities {
		if comp, ok := e.GetComponent("destructible"); ok {
			if destComp, ok := comp.(*DestructibleComponent); ok {
				if destComp.TileX == 5 && destComp.TileY == 5 {
					found = true
					if destComp.Health >= destComp.MaxHealth {
						t.Error("expected tile to have taken damage")
					}
					break
				}
			}
		}
	}
	if !found {
		t.Error("expected destructible entity to be created")
	}
}

// TestTerrainModificationSystem_DamageTile_NonWall verifies only walls take damage.
func TestTerrainModificationSystem_DamageTile_NonWall(t *testing.T) {
	system := NewTerrainModificationSystem(32)
	w := NewWorld()
	terr := createTestTerrain(10, 10)
	worldMap := world.NewMap(10, 10, 12345)

	system.SetWorld(w)
	system.SetTerrain(terr)
	system.SetWorldMap(worldMap)

	// Set tile as floor (not wall)
	terr.SetTile(5, 5, terrain.TileFloor)

	// Apply damage
	system.damageTile(5, 5, 50.0)

	// Verify no destructible entity created
	w.Update(0)
	entities := w.GetEntities()
	for _, e := range entities {
		if comp, ok := e.GetComponent("destructible"); ok {
			if destComp, ok := comp.(*DestructibleComponent); ok {
				if destComp.TileX == 5 && destComp.TileY == 5 {
					t.Error("should not create destructible entity for non-wall")
				}
			}
		}
	}
}

// TestTerrainModificationSystem_Update_DestroysTile verifies tile destruction.
func TestTerrainModificationSystem_Update_DestroysTile(t *testing.T) {
	system := NewTerrainModificationSystem(32)
	w := NewWorld()
	terr := createTestTerrain(10, 10)
	worldMap := world.NewMap(10, 10, 12345)

	system.SetWorld(w)
	system.SetTerrain(terr)
	system.SetWorldMap(worldMap)

	// Set tile as wall
	terr.SetTile(5, 5, terrain.TileWall)

	// Create destructible entity at low health
	entity := w.CreateEntity()
	destComp := NewDestructibleComponent(MaterialStone, 5, 5)
	destComp.Health = 10
	destComp.TakeDamage(15) // Destroy it
	entity.AddComponent(destComp)
	w.Update(0)

	// Verify tile is still wall before update
	if terr.GetTile(5, 5) != terrain.TileWall {
		t.Error("tile should still be wall before update")
	}

	// Update system
	entities := w.GetEntities()
	system.Update(entities, 0.016)

	// Verify tile replaced with floor
	if terr.GetTile(5, 5) != terrain.TileFloor {
		t.Errorf("expected tile to be floor, got %v", terr.GetTile(5, 5))
	}

	// Verify entity removed
	w.Update(0)
	entities = w.GetEntities()
	for _, e := range entities {
		if e.ID == entity.ID {
			t.Error("destroyed tile entity should be removed")
		}
	}
}

// TestTerrainModificationSystem_ProcessWeaponAttack verifies weapon attack processing.
func TestTerrainModificationSystem_ProcessWeaponAttack(t *testing.T) {
	system := NewTerrainModificationSystem(32)
	w := NewWorld()
	terr := createTestTerrain(10, 10)
	worldMap := world.NewMap(10, 10, 12345)

	system.SetWorld(w)
	system.SetTerrain(terr)
	system.SetWorldMap(worldMap)

	// Create player entity at tile (5, 5), facing right
	player := w.CreateEntity()
	player.AddComponent(&PositionComponent{X: 5*32 + 16, Y: 5*32 + 16})
	player.AddComponent(&AnimationComponent{Facing: DirRight})

	// Set wall in front of player (6, 5)
	terr.SetTile(6, 5, terrain.TileWall)

	// Create weapon
	weapon := &item.Item{
		Type:  item.TypeWeapon,
		Stats: item.Stats{Damage: 20},
	}

	// Process attack
	system.ProcessWeaponAttack(player, weapon)

	// Verify destructible entity created at (6, 5)
	w.Update(0)
	entities := w.GetEntities()
	found := false
	for _, e := range entities {
		if comp, ok := e.GetComponent("destructible"); ok {
			if destComp, ok := comp.(*DestructibleComponent); ok {
				if destComp.TileX == 6 && destComp.TileY == 5 {
					found = true
					if destComp.Health >= destComp.MaxHealth {
						t.Error("tile should have taken damage")
					}
					break
				}
			}
		}
	}
	if !found {
		t.Error("expected destructible entity at (6, 5)")
	}
}

// TestTerrainModificationSystem_ProcessWeaponAttack_NoPosition verifies no crash without position.
func TestTerrainModificationSystem_ProcessWeaponAttack_NoPosition(t *testing.T) {
	system := NewTerrainModificationSystem(32)
	w := NewWorld()
	terr := createTestTerrain(10, 10)
	worldMap := world.NewMap(10, 10, 12345)

	system.SetWorld(w)
	system.SetTerrain(terr)
	system.SetWorldMap(worldMap)

	// Create entity without position
	entity := w.CreateEntity()

	weapon := &item.Item{Type: item.TypeWeapon}

	// Should not crash
	system.ProcessWeaponAttack(entity, weapon)

	// No entities should be created
	w.Update(0)
	entities := w.GetEntities()
	count := 0
	for _, e := range entities {
		if _, ok := e.GetComponent("destructible"); ok {
			count++
		}
	}
	if count > 0 {
		t.Error("should not create destructible entities without valid position")
	}
}

// TestTerrainModificationSystem_ProcessWeaponAttack_NonWeapon verifies non-weapons don't damage terrain.
func TestTerrainModificationSystem_ProcessWeaponAttack_NonWeapon(t *testing.T) {
	system := NewTerrainModificationSystem(32)
	w := NewWorld()
	terr := createTestTerrain(10, 10)
	worldMap := world.NewMap(10, 10, 12345)

	system.SetWorld(w)
	system.SetTerrain(terr)
	system.SetWorldMap(worldMap)

	// Create entity
	entity := w.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 5*32 + 16, Y: 5*32 + 16})
	entity.AddComponent(&AnimationComponent{Facing: DirRight})

	// Set wall in front
	terr.SetTile(6, 5, terrain.TileWall)

	// Use armor (not a weapon)
	armor := &item.Item{Type: item.TypeArmor}

	// Process attack
	system.ProcessWeaponAttack(entity, armor)

	// Verify no destructible entity created
	w.Update(0)
	entities := w.GetEntities()
	for _, e := range entities {
		if comp, ok := e.GetComponent("destructible"); ok {
			if destComp, ok := comp.(*DestructibleComponent); ok {
				if destComp.TileX == 6 && destComp.TileY == 5 {
					t.Error("non-weapon should not damage terrain")
				}
			}
		}
	}
}

// TestTerrainModificationSystem_DamageTileAtWorldPosition verifies world coordinate damage.
func TestTerrainModificationSystem_DamageTileAtWorldPosition(t *testing.T) {
	system := NewTerrainModificationSystem(32)
	w := NewWorld()
	terr := createTestTerrain(10, 10)
	worldMap := world.NewMap(10, 10, 12345)

	system.SetWorld(w)
	system.SetTerrain(terr)
	system.SetWorldMap(worldMap)

	// Set wall at tile (5, 5)
	terr.SetTile(5, 5, terrain.TileWall)

	// Damage using world coordinates (middle of tile)
	worldX := float64(5*32 + 16)
	worldY := float64(5*32 + 16)
	system.DamageTileAtWorldPosition(worldX, worldY, 50.0)

	// Verify destructible entity created
	w.Update(0)
	entities := w.GetEntities()
	found := false
	for _, e := range entities {
		if comp, ok := e.GetComponent("destructible"); ok {
			if destComp, ok := comp.(*DestructibleComponent); ok {
				if destComp.TileX == 5 && destComp.TileY == 5 {
					found = true
					break
				}
			}
		}
	}
	if !found {
		t.Error("expected destructible entity at (5, 5)")
	}
}

// TestTerrainModificationSystem_DamageTilesInArea verifies area damage.
func TestTerrainModificationSystem_DamageTilesInArea(t *testing.T) {
	system := NewTerrainModificationSystem(32)
	w := NewWorld()
	terr := createTestTerrain(20, 20)
	worldMap := world.NewMap(20, 20, 12345)

	system.SetWorld(w)
	system.SetTerrain(terr)
	system.SetWorldMap(worldMap)

	// Create a 3x3 grid of walls centered at (10, 10)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			terr.SetTile(10+dx, 10+dy, terrain.TileWall)
		}
	}

	// Apply area damage (radius of ~1.5 tiles = 48 units)
	centerX := float64(10*32 + 16)
	centerY := float64(10*32 + 16)
	system.DamageTilesInArea(centerX, centerY, 48.0, 30.0)

	// Count damaged tiles
	w.Update(0)
	entities := w.GetEntities()
	damagedCount := 0
	for _, e := range entities {
		if comp, ok := e.GetComponent("destructible"); ok {
			if destComp, ok := comp.(*DestructibleComponent); ok {
				// Check if tile is in the 3x3 grid
				if destComp.TileX >= 9 && destComp.TileX <= 11 &&
					destComp.TileY >= 9 && destComp.TileY <= 11 {
					damagedCount++
				}
			}
		}
	}

	// Should damage at least the center tile, possibly adjacent ones
	if damagedCount == 0 {
		t.Error("expected at least one tile to be damaged in area")
	}
}

// TestTerrainModificationSystem_DamageTilesInArea_CircularRadius verifies circular area damage.
func TestTerrainModificationSystem_DamageTilesInArea_CircularRadius(t *testing.T) {
	system := NewTerrainModificationSystem(32)
	w := NewWorld()
	terr := createTestTerrain(20, 20)
	worldMap := world.NewMap(20, 20, 12345)

	system.SetWorld(w)
	system.SetTerrain(terr)
	system.SetWorldMap(worldMap)

	// Create 5x5 grid of walls
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			terr.SetTile(10+dx, 10+dy, terrain.TileWall)
		}
	}

	// Apply small radius (should only hit center tile)
	centerX := float64(10*32 + 16)
	centerY := float64(10*32 + 16)
	system.DamageTilesInArea(centerX, centerY, 16.0, 30.0)

	// Count damaged tiles
	w.Update(0)
	entities := w.GetEntities()
	damagedCount := 0
	for _, e := range entities {
		if _, ok := e.GetComponent("destructible"); ok {
			damagedCount++
		}
	}

	// Small radius should only hit center tile
	if damagedCount != 1 {
		t.Errorf("expected 1 damaged tile with small radius, got %d", damagedCount)
	}
}

// TestTerrainModificationSystem_FindDestructibleEntityAt verifies entity lookup.
func TestTerrainModificationSystem_FindDestructibleEntityAt(t *testing.T) {
	system := NewTerrainModificationSystem(32)
	w := NewWorld()

	system.SetWorld(w)

	// Create destructible entity at (5, 5)
	entity := w.CreateEntity()
	destComp := NewDestructibleComponent(MaterialStone, 5, 5)
	entity.AddComponent(destComp)
	w.Update(0)

	// Find it
	found := system.findDestructibleEntityAt(5, 5)
	if found == nil {
		t.Error("expected to find destructible entity at (5, 5)")
	}
	if found.ID != entity.ID {
		t.Error("found wrong entity")
	}

	// Try wrong coordinates
	notFound := system.findDestructibleEntityAt(10, 10)
	if notFound != nil {
		t.Error("should not find entity at wrong coordinates")
	}
}

// TestTerrainModificationSystem_CreateDestructibleEntity verifies entity creation.
func TestTerrainModificationSystem_CreateDestructibleEntity(t *testing.T) {
	system := NewTerrainModificationSystem(32)
	w := NewWorld()

	system.SetWorld(w)

	// Create entity
	entity := system.createDestructibleEntity(5, 5)
	if entity == nil {
		t.Fatal("expected entity to be created")
	}

	// Verify destructible component
	comp, ok := entity.GetComponent("destructible")
	if !ok {
		t.Fatal("expected destructible component")
	}
	destComp, ok := comp.(*DestructibleComponent)
	if !ok {
		t.Fatal("component is not DestructibleComponent")
	}
	if destComp.TileX != 5 || destComp.TileY != 5 {
		t.Errorf("expected tile coords (5, 5), got (%d, %d)", destComp.TileX, destComp.TileY)
	}

	// Verify position component
	posComp, ok := entity.GetComponent("position")
	if !ok {
		t.Fatal("expected position component")
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		t.Fatal("component is not PositionComponent")
	}

	// Position should be at tile center
	expectedX := float64(5*32 + 16)
	expectedY := float64(5*32 + 16)
	if pos.X != expectedX || pos.Y != expectedY {
		t.Errorf("expected position (%f, %f), got (%f, %f)", expectedX, expectedY, pos.X, pos.Y)
	}
}

// TestTerrainModificationSystem_ReplaceTileWithFloor verifies tile replacement.
func TestTerrainModificationSystem_ReplaceTileWithFloor(t *testing.T) {
	system := NewTerrainModificationSystem(32)
	terr := createTestTerrain(10, 10)
	worldMap := world.NewMap(10, 10, 12345)

	system.SetTerrain(terr)
	system.SetWorldMap(worldMap)

	// Set tile as wall
	terr.SetTile(5, 5, terrain.TileWall)

	// Replace with floor
	system.replaceTileWithFloor(5, 5)

	// Verify terrain updated
	if terr.GetTile(5, 5) != terrain.TileFloor {
		t.Errorf("expected terrain tile to be floor, got %v", terr.GetTile(5, 5))
	}

	// Verify world map updated
	tile := worldMap.GetTile(5, 5)
	if tile == nil {
		t.Fatal("expected world map tile")
	}
	if tile.Type != world.TileFloor {
		t.Errorf("expected world map tile to be floor, got %v", tile.Type)
	}
	if !tile.Walkable {
		t.Error("floor tile should be walkable")
	}
}

// Benchmark tile damage processing.
func BenchmarkTerrainModificationSystem_DamageTile(b *testing.B) {
	system := NewTerrainModificationSystem(32)
	w := NewWorld()
	terr := createTestTerrain(50, 50)
	worldMap := world.NewMap(50, 50, 12345)

	system.SetWorld(w)
	system.SetTerrain(terr)
	system.SetWorldMap(worldMap)

	// Set all tiles as walls
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			terr.SetTile(x, y, terrain.TileWall)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := i % 50
		y := (i / 50) % 50
		system.damageTile(x, y, 10.0)
	}
}

// Benchmark area damage processing.
func BenchmarkTerrainModificationSystem_DamageTilesInArea(b *testing.B) {
	system := NewTerrainModificationSystem(32)
	w := NewWorld()
	terr := createTestTerrain(50, 50)
	worldMap := world.NewMap(50, 50, 12345)

	system.SetWorld(w)
	system.SetTerrain(terr)
	system.SetWorldMap(worldMap)

	// Set all tiles as walls
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			terr.SetTile(x, y, terrain.TileWall)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := float64((i%40 + 5) * 32)
		y := float64(((i/40)%40 + 5) * 32)
		system.DamageTilesInArea(x, y, 64.0, 20.0)
	}
}

// Benchmark weapon attack processing.
func BenchmarkTerrainModificationSystem_ProcessWeaponAttack(b *testing.B) {
	system := NewTerrainModificationSystem(32)
	w := NewWorld()
	terr := createTestTerrain(50, 50)
	worldMap := world.NewMap(50, 50, 12345)

	system.SetWorld(w)
	system.SetTerrain(terr)
	system.SetWorldMap(worldMap)

	// Set all tiles as walls
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			terr.SetTile(x, y, terrain.TileWall)
		}
	}

	// Create player entity
	player := w.CreateEntity()
	player.AddComponent(&PositionComponent{X: 25 * 32, Y: 25 * 32})
	player.AddComponent(&AnimationComponent{Facing: DirRight})

	weapon := &item.Item{
		Type:  item.TypeWeapon,
		Stats: item.Stats{Damage: 20},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.ProcessWeaponAttack(player, weapon)
	}
}
