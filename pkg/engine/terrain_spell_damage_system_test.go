package engine

import (
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/magic"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

func TestNewTerrainSpellDamageSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainSpellDamageSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewTerrainSpellDamageSystem returned nil")
	}

	if sys.world != world {
		t.Error("System world reference incorrect")
	}

	if sys.tileSize != 32 {
		t.Errorf("Default tile size = %d, want 32", sys.tileSize)
	}

	if len(sys.baseSynergies) == 0 {
		t.Error("Base synergies not initialized")
	}

	if len(sys.genreBonuses) == 0 {
		t.Error("Genre bonuses not initialized")
	}
}

func TestTerrainSpellDamageSystem_SetTileSize(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainSpellDamageSystem(world, 12345)

	sys.SetTileSize(64)
	if sys.tileSize != 64 {
		t.Errorf("TileSize = %d, want 64", sys.tileSize)
	}

	// Invalid size should be ignored
	sys.SetTileSize(0)
	if sys.tileSize != 64 {
		t.Errorf("TileSize changed on invalid input, got %d", sys.tileSize)
	}

	sys.SetTileSize(-10)
	if sys.tileSize != 64 {
		t.Errorf("TileSize changed on negative input, got %d", sys.tileSize)
	}
}

func TestTerrainSpellDamageSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainSpellDamageSystem(world, 12345)

	tests := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range tests {
		sys.SetGenre(genre)
		if sys.genreID != genre {
			t.Errorf("Genre = %s, want %s", sys.genreID, genre)
		}
	}
}

func TestTerrainSpellDamageSystem_SetTerrain(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainSpellDamageSystem(world, 12345)

	// Pre-populate cache
	sys.damageCache[1] = map[magic.ElementType]float64{magic.ElementFire: 1.2}
	sys.lastTileCache[1] = spellDamageTilePos{5, 5}

	terr := &terrain.Terrain{
		Width:  100,
		Height: 100,
		Tiles:  make([][]terrain.TileType, 100),
	}
	for i := range terr.Tiles {
		terr.Tiles[i] = make([]terrain.TileType, 100)
	}

	sys.SetTerrain(terr)

	if sys.terrain != terr {
		t.Error("Terrain not set correctly")
	}

	if len(sys.damageCache) != 0 {
		t.Error("Damage cache not cleared on terrain change")
	}

	if len(sys.lastTileCache) != 0 {
		t.Error("Tile cache not cleared on terrain change")
	}
}

func TestTerrainSpellDamageSystem_Update_NoTerrain(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainSpellDamageSystem(world, 12345)

	entity := NewEntity()
	entity.AddComponent(&ManaComponent{Current: 100, Max: 100})
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Should not panic with nil terrain
	sys.Update([]*Entity{entity}, 0.016)

	// Should not add component without terrain
	for _, c := range entity.Components {
		if _, ok := c.(*TerrainSpellDamageComponent); ok {
			t.Error("Component added without terrain")
		}
	}
}

func TestTerrainSpellDamageSystem_Update_WithTerrain(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainSpellDamageSystem(world, 12345)
	sys.SetGenre("fantasy")

	// Create terrain with lava
	terr := createTestTerrainWithLava()
	sys.SetTerrain(terr)

	entity := NewEntity()
	entity.AddComponent(&ManaComponent{Current: 100, Max: 100})
	entity.AddComponent(&PositionComponent{X: 32, Y: 32}) // Tile (1,1) = lava

	sys.Update([]*Entity{entity}, 0.016)

	// Check component was added
	var comp *TerrainSpellDamageComponent
	for _, c := range entity.Components {
		if tc, ok := c.(*TerrainSpellDamageComponent); ok {
			comp = tc
			break
		}
	}

	if comp == nil {
		t.Fatal("TerrainSpellDamageComponent not added")
	}

	if comp.TerrainType != "lava_flow" {
		t.Errorf("TerrainType = %s, want lava_flow", comp.TerrainType)
	}

	// Fire should have bonus on lava
	if comp.DamageMultipliers[magic.ElementFire] <= 1.0 {
		t.Errorf("Fire damage mult = %f, expected > 1.0", comp.DamageMultipliers[magic.ElementFire])
	}
}

func TestTerrainSpellDamageSystem_Update_SkipsNonManaEntities(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainSpellDamageSystem(world, 12345)
	sys.SetTerrain(createTestTerrainWithLava())

	// Entity without mana
	entity := NewEntity()
	entity.AddComponent(&PositionComponent{X: 32, Y: 32})

	sys.Update([]*Entity{entity}, 0.016)

	// Should not add component
	for _, c := range entity.Components {
		if _, ok := c.(*TerrainSpellDamageComponent); ok {
			t.Error("Component added to non-mana entity")
		}
	}
}

func TestTerrainSpellDamageSystem_CachingOptimization(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainSpellDamageSystem(world, 12345)
	sys.SetTerrain(createTestTerrainWithLava())

	entity := NewEntity()
	entity.AddComponent(&ManaComponent{Current: 100, Max: 100})
	entity.AddComponent(&PositionComponent{X: 32, Y: 32})

	// First update
	sys.Update([]*Entity{entity}, 0.016)

	// Verify cached
	if _, ok := sys.lastTileCache[entity.ID]; !ok {
		t.Error("Position not cached after first update")
	}

	// Second update on same tile should use cache
	sys.Update([]*Entity{entity}, 0.016)

	// Move entity to new tile
	pos := entity.GetPosition()
	pos.X = 96 // Tile (3,3)
	pos.Y = 96

	// Should update cache
	sys.Update([]*Entity{entity}, 0.016)

	cached := sys.lastTileCache[entity.ID]
	if cached.tileX != 3 || cached.tileY != 3 {
		t.Errorf("Cache not updated: got (%d,%d), want (3,3)", cached.tileX, cached.tileY)
	}
}

func TestTerrainSpellDamageSystem_GetDamageModifier(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainSpellDamageSystem(world, 12345)
	sys.SetGenre("fantasy")
	sys.SetTerrain(createTestTerrainWithLava())

	entity := NewEntity()
	entity.AddComponent(&ManaComponent{Current: 100, Max: 100})
	entity.AddComponent(&PositionComponent{X: 32, Y: 32}) // Lava tile

	sys.Update([]*Entity{entity}, 0.016)

	// Fire should have bonus on lava
	fireMod := sys.GetDamageModifier(entity.ID, magic.ElementFire)
	if fireMod <= 1.0 {
		t.Errorf("Fire damage modifier = %f, expected > 1.0", fireMod)
	}

	// Unknown entity returns 1.0
	unknownMod := sys.GetDamageModifier(999, magic.ElementFire)
	if unknownMod != 1.0 {
		t.Errorf("Unknown entity modifier = %f, want 1.0", unknownMod)
	}
}

func TestTerrainSpellDamageSystem_ElementSynergies(t *testing.T) {
	tests := []struct {
		name        string
		tileType    terrain.TileType
		element     magic.ElementType
		expectBonus bool
	}{
		{"fire on lava", terrain.TileLavaFlow, magic.ElementFire, true},
		{"ice on shallow water", terrain.TileWaterShallow, magic.ElementIce, true},
		{"ice on deep water", terrain.TileWaterDeep, magic.ElementIce, true},
		{"lightning on platform", terrain.TilePlatform, magic.ElementLightning, true},
		{"earth on tree", terrain.TileTree, magic.ElementEarth, true},
		{"arcane on structure", terrain.TileStructure, magic.ElementArcane, true},
		{"dark on trap door", terrain.TileTrapDoor, magic.ElementDark, true},
		{"fire on floor", terrain.TileFloor, magic.ElementFire, false},
		{"ice on lava", terrain.TileLavaFlow, magic.ElementIce, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewTerrainSpellDamageSystem(world, 12345)

			mult := sys.calculateMultiplier(tt.tileType, tt.element)

			if tt.expectBonus && mult <= 1.0 {
				t.Errorf("Expected bonus for %s, got %f", tt.name, mult)
			}
			if !tt.expectBonus && mult > 1.0 {
				t.Errorf("Unexpected bonus for %s, got %f", tt.name, mult)
			}
		})
	}
}

func TestTerrainSpellDamageSystem_GenreModifiers(t *testing.T) {
	tests := []struct {
		name    string
		genre   string
		element magic.ElementType
		terrain terrain.TileType
	}{
		{"fantasy fire boost", "fantasy", magic.ElementFire, terrain.TileLavaFlow},
		{"horror dark boost", "horror", magic.ElementDark, terrain.TileTrapDoor},
		{"scifi lightning boost", "scifi", magic.ElementLightning, terrain.TilePlatform},
		{"cyberpunk arcane boost", "cyberpunk", magic.ElementArcane, terrain.TileStructure},
		{"postapoc fire boost", "postapoc", magic.ElementFire, terrain.TileLavaFlow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewTerrainSpellDamageSystem(world, 12345)

			// Get base multiplier
			sys.SetGenre("")
			baseMult := sys.calculateMultiplier(tt.terrain, tt.element)

			// Get genre multiplier
			sys.SetGenre(tt.genre)
			genreMult := sys.calculateMultiplier(tt.terrain, tt.element)

			// Genre should provide additional bonus
			if genreMult <= baseMult {
				t.Errorf("%s: genre mult %f <= base mult %f", tt.name, genreMult, baseMult)
			}
		})
	}
}

func TestTerrainSpellDamageComponent_Type(t *testing.T) {
	comp := &TerrainSpellDamageComponent{
		DamageMultipliers: make(map[magic.ElementType]float64),
		TerrainType:       "lava_flow",
		BonusElement:      magic.ElementFire,
	}

	if comp.Type() != "terrain_spell_damage" {
		t.Errorf("Type() = %s, want terrain_spell_damage", comp.Type())
	}
}

func TestTerrainSpellDamageSystem_GetTerrainType(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainSpellDamageSystem(world, 12345)
	sys.SetTerrain(createTestTerrainWithLava())

	entity := NewEntity()
	entity.AddComponent(&ManaComponent{Current: 100, Max: 100})
	entity.AddComponent(&PositionComponent{X: 32, Y: 32})
	world.AddEntity(entity)

	sys.Update([]*Entity{entity}, 0.016)

	terrainType := sys.GetTerrainType(entity.ID)
	if terrainType != "lava_flow" {
		t.Errorf("GetTerrainType() = %s, want lava_flow", terrainType)
	}

	// Unknown entity
	unknownType := sys.GetTerrainType(999)
	if unknownType != "" {
		t.Errorf("GetTerrainType(999) = %s, want empty", unknownType)
	}
}

func TestTerrainSpellDamageSystem_GetBonusElement(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainSpellDamageSystem(world, 12345)
	sys.SetTerrain(createTestTerrainWithLava())

	entity := NewEntity()
	entity.AddComponent(&ManaComponent{Current: 100, Max: 100})
	entity.AddComponent(&PositionComponent{X: 32, Y: 32})
	world.AddEntity(entity)

	sys.Update([]*Entity{entity}, 0.016)

	bonusElem := sys.GetBonusElement(entity.ID)
	if bonusElem != magic.ElementFire {
		t.Errorf("GetBonusElement() = %v, want Fire", bonusElem)
	}

	// Unknown entity
	unknownElem := sys.GetBonusElement(999)
	if unknownElem != magic.ElementNone {
		t.Errorf("GetBonusElement(999) = %v, want None", unknownElem)
	}
}

func TestTerrainSpellDamageSystem_Determinism(t *testing.T) {
	seed := int64(54321)
	terr := createTestTerrainWithLava()

	// Create two systems with same seed
	world1 := NewWorld()
	sys1 := NewTerrainSpellDamageSystem(world1, seed)
	sys1.SetGenre("fantasy")
	sys1.SetTerrain(terr)

	world2 := NewWorld()
	sys2 := NewTerrainSpellDamageSystem(world2, seed)
	sys2.SetGenre("fantasy")
	sys2.SetTerrain(terr)

	// Compare multipliers for all elements on lava tile
	elements := []magic.ElementType{
		magic.ElementFire, magic.ElementIce, magic.ElementLightning,
		magic.ElementEarth, magic.ElementWind, magic.ElementLight,
		magic.ElementDark, magic.ElementArcane,
	}

	for _, elem := range elements {
		mult1 := sys1.calculateMultiplier(terrain.TileLavaFlow, elem)
		mult2 := sys2.calculateMultiplier(terrain.TileLavaFlow, elem)

		if mult1 != mult2 {
			t.Errorf("Non-deterministic result for %s: %f != %f", elem.String(), mult1, mult2)
		}
	}
}

// createTestTerrainWithLava creates a test terrain with lava at tile (1,1).
func createTestTerrainWithLava() *terrain.Terrain {
	terr := &terrain.Terrain{
		Width:  10,
		Height: 10,
		Tiles:  make([][]terrain.TileType, 10),
	}
	for i := range terr.Tiles {
		terr.Tiles[i] = make([]terrain.TileType, 10)
		for j := range terr.Tiles[i] {
			terr.Tiles[i][j] = terrain.TileFloor
		}
	}
	// Set lava at (1,1)
	terr.Tiles[1][1] = terrain.TileLavaFlow
	return terr
}

func BenchmarkTerrainSpellDamageSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewTerrainSpellDamageSystem(world, 12345)
	sys.SetGenre("fantasy")
	sys.SetTerrain(createTestTerrainWithLava())

	// Create 100 entities
	entities := make([]*Entity, 100)
	rng := rand.New(rand.NewSource(67890))
	for i := 0; i < 100; i++ {
		e := NewEntity()
		e.AddComponent(&ManaComponent{Current: 100, Max: 100})
		e.AddComponent(&PositionComponent{
			X: float64(rng.Intn(320)),
			Y: float64(rng.Intn(320)),
		})
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}

func BenchmarkTerrainSpellDamageSystem_GetDamageModifier(b *testing.B) {
	world := NewWorld()
	sys := NewTerrainSpellDamageSystem(world, 12345)
	sys.SetTerrain(createTestTerrainWithLava())

	entity := NewEntity()
	entity.AddComponent(&ManaComponent{Current: 100, Max: 100})
	entity.AddComponent(&PositionComponent{X: 32, Y: 32})
	sys.Update([]*Entity{entity}, 0.016)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.GetDamageModifier(entity.ID, magic.ElementFire)
	}
}
