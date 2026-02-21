package engine

import (
	"image/color"
	"testing"
)

func TestHumanoidTextureComponent_Type(t *testing.T) {
	c := NewHumanoidTextureComponent()
	if c.Type() != "humanoid_texture" {
		t.Errorf("Type() = %s, want humanoid_texture", c.Type())
	}
}

func TestNewHumanoidTextureComponent_Defaults(t *testing.T) {
	c := NewHumanoidTextureComponent()

	if c.SkinTextureType != 0 {
		t.Errorf("SkinTextureType = %d, want 0", c.SkinTextureType)
	}
	if c.SkinIntensity != 0.2 {
		t.Errorf("SkinIntensity = %f, want 0.2", c.SkinIntensity)
	}
	if c.SkinScale != 1.0 {
		t.Errorf("SkinScale = %f, want 1.0", c.SkinScale)
	}
	if c.ClothingTopTextureType != 5 {
		t.Errorf("ClothingTopTextureType = %d, want 5 (FabricLinen)", c.ClothingTopTextureType)
	}
	if c.HairTextureType != 11 {
		t.Errorf("HairTextureType = %d, want 11 (HairStraight)", c.HairTextureType)
	}
	if !c.Enabled {
		t.Error("Enabled should be true by default")
	}
	if c.Dirty {
		t.Error("Dirty should be false by default")
	}
}

func TestNewHumanoidTextureSystem(t *testing.T) {
	world := NewWorld(nil)
	sys := NewHumanoidTextureSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewHumanoidTextureSystem returned nil")
	}
	if sys.world != world {
		t.Error("world not set correctly")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("genreID = %s, want fantasy", sys.genreID)
	}
	if sys.scanInterval != 2.0 {
		t.Errorf("scanInterval = %f, want 2.0", sys.scanInterval)
	}
}

func TestHumanoidTextureSystem_SetGenre(t *testing.T) {
	world := NewWorld(nil)
	sys := NewHumanoidTextureSystem(world, 12345)

	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("genreID = %s, want horror", sys.genreID)
	}

	sys.SetGenre("cyberpunk")
	if sys.genreID != "cyberpunk" {
		t.Errorf("genreID = %s, want cyberpunk", sys.genreID)
	}
}

func TestHumanoidTextureSystem_Update_NilEntities(t *testing.T) {
	world := NewWorld(nil)
	sys := NewHumanoidTextureSystem(world, 12345)

	// Should not panic with nil entities
	sys.Update(nil, 3.0)
}

func TestHumanoidTextureSystem_Update_EmptyEntities(t *testing.T) {
	world := NewWorld(nil)
	sys := NewHumanoidTextureSystem(world, 12345)

	// Should not panic with empty entities
	sys.Update([]*Entity{}, 3.0)
}

func TestHumanoidTextureSystem_Update_ScanInterval(t *testing.T) {
	world := NewWorld(nil)
	sys := NewHumanoidTextureSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&PlayerComponent{})

	// First call with deltaTime < scanInterval should not process
	sys.Update([]*Entity{entity}, 1.0)
	_, has := entity.GetComponent("humanoid_texture")
	if has {
		t.Error("Component added before scan interval elapsed")
	}

	// Second call should trigger scan (total time >= scanInterval)
	sys.Update([]*Entity{entity}, 1.5)
	_, has = entity.GetComponent("humanoid_texture")
	if !has {
		t.Error("Component not added after scan interval elapsed")
	}
}

func TestHumanoidTextureSystem_Update_PlayerEntity(t *testing.T) {
	world := NewWorld(nil)
	sys := NewHumanoidTextureSystem(world, 42)

	entity := world.CreateEntity()
	entity.AddComponent(&PlayerComponent{})

	// Trigger scan
	sys.Update([]*Entity{entity}, 3.0)

	comp, has := entity.GetComponent("humanoid_texture")
	if !has {
		t.Fatal("HumanoidTextureComponent not added to player entity")
	}

	ht, ok := comp.(*HumanoidTextureComponent)
	if !ok {
		t.Fatal("Component is not HumanoidTextureComponent")
	}

	if !ht.Enabled {
		t.Error("Texture not enabled")
	}
	if ht.GenreID != "fantasy" {
		t.Errorf("GenreID = %s, want fantasy", ht.GenreID)
	}
}

func TestHumanoidTextureSystem_Update_NPCEntity(t *testing.T) {
	world := NewWorld(nil)
	sys := NewHumanoidTextureSystem(world, 42)

	entity := world.CreateEntity()
	entity.AddComponent(&NPCComponent{})

	// Trigger scan
	sys.Update([]*Entity{entity}, 3.0)

	_, has := entity.GetComponent("humanoid_texture")
	if !has {
		t.Error("HumanoidTextureComponent not added to NPC entity")
	}
}

func TestHumanoidTextureSystem_Update_HumanoidCreature(t *testing.T) {
	world := NewWorld(nil)
	sys := NewHumanoidTextureSystem(world, 42)

	entity := world.CreateEntity()
	entity.AddComponent(&CreatureVisualComponent{Form: FormHumanoid})

	// Trigger scan
	sys.Update([]*Entity{entity}, 3.0)

	_, has := entity.GetComponent("humanoid_texture")
	if !has {
		t.Error("HumanoidTextureComponent not added to humanoid creature")
	}
}

func TestHumanoidTextureSystem_Update_NonHumanoidCreature(t *testing.T) {
	world := NewWorld(nil)
	sys := NewHumanoidTextureSystem(world, 42)

	entity := world.CreateEntity()
	entity.AddComponent(&CreatureVisualComponent{Form: FormQuadruped})

	// Trigger scan
	sys.Update([]*Entity{entity}, 3.0)

	_, has := entity.GetComponent("humanoid_texture")
	if has {
		t.Error("HumanoidTextureComponent added to non-humanoid creature (should be skipped)")
	}
}

func TestHumanoidTextureSystem_Update_MerchantEntity(t *testing.T) {
	world := NewWorld(nil)
	sys := NewHumanoidTextureSystem(world, 42)

	entity := world.CreateEntity()
	entity.AddComponent(&MerchantComponent{})

	// Trigger scan
	sys.Update([]*Entity{entity}, 3.0)

	_, has := entity.GetComponent("humanoid_texture")
	if !has {
		t.Error("HumanoidTextureComponent not added to merchant entity")
	}
}

func TestHumanoidTextureSystem_Update_GenreChange(t *testing.T) {
	world := NewWorld(nil)
	sys := NewHumanoidTextureSystem(world, 42)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&PlayerComponent{})

	// First scan
	sys.Update([]*Entity{entity}, 3.0)

	comp, _ := entity.GetComponent("humanoid_texture")
	ht := comp.(*HumanoidTextureComponent)
	ht.Dirty = false

	// Change genre
	sys.SetGenre("horror")

	// Second scan should update component
	sys.Update([]*Entity{entity}, 3.0)

	if ht.GenreID != "horror" {
		t.Errorf("GenreID not updated: %s", ht.GenreID)
	}
	if !ht.Dirty {
		t.Error("Dirty flag not set after genre change")
	}
}

func TestHumanoidTextureSystem_GetActiveTextureCount(t *testing.T) {
	world := NewWorld(nil)
	sys := NewHumanoidTextureSystem(world, 42)

	// No entities yet
	if count := sys.GetActiveTextureCount(); count != 0 {
		t.Errorf("Count = %d, want 0", count)
	}

	// Add entity with texture
	entity := world.CreateEntity()
	entity.AddComponent(&PlayerComponent{})
	sys.Update([]*Entity{entity}, 3.0)

	if count := sys.GetActiveTextureCount(); count != 1 {
		t.Errorf("Count = %d, want 1", count)
	}

	// Add another
	entity2 := world.CreateEntity()
	entity2.AddComponent(&NPCComponent{})
	sys.Update([]*Entity{entity, entity2}, 3.0)

	if count := sys.GetActiveTextureCount(); count != 2 {
		t.Errorf("Count = %d, want 2", count)
	}
}

func TestHumanoidTextureSystem_GetTextureBreakdown(t *testing.T) {
	world := NewWorld(nil)
	sys := NewHumanoidTextureSystem(world, 42)

	entity := world.CreateEntity()
	entity.AddComponent(&PlayerComponent{})
	sys.Update([]*Entity{entity}, 3.0)

	breakdown := sys.GetTextureBreakdown()
	if len(breakdown) == 0 {
		t.Error("Breakdown is empty")
	}

	// Should have at least skin, clothing, and hair entries
	hasSkin := false
	hasClothing := false
	hasHair := false
	for key := range breakdown {
		if len(key) > 5 && key[:5] == "skin:" {
			hasSkin = true
		}
		if len(key) > 9 && key[:9] == "clothing:" {
			hasClothing = true
		}
		if len(key) > 5 && key[:5] == "hair:" {
			hasHair = true
		}
	}

	if !hasSkin {
		t.Error("No skin entry in breakdown")
	}
	if !hasClothing {
		t.Error("No clothing entry in breakdown")
	}
	if !hasHair {
		t.Error("No hair entry in breakdown")
	}
}

func TestResolveTextureColor(t *testing.T) {
	base := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	primary := color.RGBA{R: 200, G: 50, B: 50, A: 200}
	secondary := color.RGBA{R: 50, G: 200, B: 50, A: 200}

	// Zero intensity should return base
	result := ResolveTextureColor(primary, secondary, base, 0)
	if result != base {
		t.Errorf("Zero intensity should return base: got %v, want %v", result, base)
	}

	// 50% intensity should blend
	result = ResolveTextureColor(primary, secondary, base, 0.5)
	if result.R < 100 || result.R > 200 {
		t.Errorf("R value not blended correctly: %d", result.R)
	}

	// Over 1.0 intensity should clamp to 1.0
	result = ResolveTextureColor(primary, secondary, base, 2.0)
	if result.R != primary.R {
		t.Errorf("Clamped intensity should use primary: R=%d, want %d", result.R, primary.R)
	}
}

func TestHumanoidTextureSystem_DeterministicGeneration(t *testing.T) {
	seed := int64(12345)

	world1 := NewWorld(nil)
	sys1 := NewHumanoidTextureSystem(world1, seed)
	entity1 := world1.CreateEntity()
	entity1.AddComponent(&PlayerComponent{})
	sys1.Update([]*Entity{entity1}, 3.0)

	world2 := NewWorld(nil)
	sys2 := NewHumanoidTextureSystem(world2, seed)
	entity2 := world2.CreateEntity()
	entity2.AddComponent(&PlayerComponent{})
	sys2.Update([]*Entity{entity2}, 3.0)

	comp1, _ := entity1.GetComponent("humanoid_texture")
	comp2, _ := entity2.GetComponent("humanoid_texture")

	ht1 := comp1.(*HumanoidTextureComponent)
	ht2 := comp2.(*HumanoidTextureComponent)

	// Same seed should produce same results
	if ht1.SkinTextureType != ht2.SkinTextureType {
		t.Error("Skin texture type not deterministic")
	}
	if ht1.ClothingTopTextureType != ht2.ClothingTopTextureType {
		t.Error("Clothing texture type not deterministic")
	}
	if ht1.HairTextureType != ht2.HairTextureType {
		t.Error("Hair texture type not deterministic")
	}
}

func BenchmarkHumanoidTextureSystem_Update(b *testing.B) {
	world := NewWorld(nil)
	sys := NewHumanoidTextureSystem(world, 42)

	entities := make([]*Entity, 100)
	for i := range entities {
		e := world.CreateEntity()
		e.AddComponent(&PlayerComponent{})
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceScan = 10.0 // Force scan every iteration
		sys.Update(entities, 0.016)
	}
}
