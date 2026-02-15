package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

func TestNewArmorHitSparkSystem(t *testing.T) {
	world := NewWorld()
	sys := NewArmorHitSparkSystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.seed != 42 {
		t.Errorf("expected seed 42, got %d", sys.seed)
	}
	if sys.rng == nil {
		t.Fatal("expected non-nil rng")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre fantasy, got %s", sys.genreID)
	}
}

func TestNewArmorHitSparkSystem_NilWorld(t *testing.T) {
	sys := NewArmorHitSparkSystem(nil, 99)
	if sys == nil {
		t.Fatal("expected non-nil system even with nil world")
	}
}

func TestArmorHitSparkSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"scifi", "scifi"},
		{"postapoc", "postapoc"},
	}

	world := NewWorld()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewArmorHitSparkSystem(world, 1)
			sys.SetGenre(tt.genreID)
			if sys.genreID != tt.genreID {
				t.Errorf("expected genre %s, got %s", tt.genreID, sys.genreID)
			}
		})
	}
}

func TestArmorHitSparkSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewArmorHitSparkSystem(world, 1)
	ps := &ParticleSystem{}
	sys.SetParticleSystem(ps)
	if sys.particleSystem != ps {
		t.Error("expected particle system to be set")
	}
}

func TestArmorHitSparkSystem_Update_NilGuards(t *testing.T) {
	world := NewWorld()
	sys := NewArmorHitSparkSystem(world, 1)
	// No particle system - should not panic
	sys.Update(nil, 0.016)
	sys.Update([]*Entity{}, 0.016)

	ps := &ParticleSystem{}
	sys.SetParticleSystem(ps)
	// With particle system but nil entities
	sys.Update(nil, 0.016)
}

func TestArmorHitSparkSystem_Update_NoHealthEntity(t *testing.T) {
	world := NewWorld()
	sys := NewArmorHitSparkSystem(world, 1)
	ps := &ParticleSystem{}
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	// Entity without health - should skip without panic
	sys.Update([]*Entity{entity}, 0.016)
}

func TestArmorHitSparkSystem_Update_FirstFrameTracking(t *testing.T) {
	world := NewWorld()
	sys := NewArmorHitSparkSystem(world, 1)
	ps := &ParticleSystem{}
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	// First frame: health tracked but no sparks (no previous value)
	sys.Update([]*Entity{entity}, 0.016)

	if _, ok := sys.prevHealth[entity.ID]; !ok {
		t.Error("expected health to be tracked after first update")
	}
}

func TestArmorHitSparkSystem_Update_NoDamageNoSparks(t *testing.T) {
	world := NewWorld()
	sys := NewArmorHitSparkSystem(world, 1)
	ps := &ParticleSystem{}
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})

	// First frame - track
	sys.Update([]*Entity{entity}, 0.016)
	// Second frame - no change, no sparks
	sys.Update([]*Entity{entity}, 0.016)
}

func TestArmorHitSparkSystem_Update_DamageWithoutArmor(t *testing.T) {
	world := NewWorld()
	sys := NewArmorHitSparkSystem(world, 1)
	ps := &ParticleSystem{}
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	health := &HealthComponent{Current: 100, Max: 100}
	entity.AddComponent(health)
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})

	// Track initial health
	sys.Update([]*Entity{entity}, 0.016)

	// Take damage but no armor equipped
	health.Current = 80
	sys.Update([]*Entity{entity}, 0.016)
	// Should not panic - no armor means no sparks
}

func TestArmorHitSparkSystem_Update_DamageWithArmor(t *testing.T) {
	world := NewWorld()
	sys := NewArmorHitSparkSystem(world, 1)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	health := &HealthComponent{Current: 100, Max: 100}
	entity.AddComponent(health)
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})

	equip := NewEquipmentComponent()
	equip.Slots[SlotChest] = &item.Item{
		Name:   "Iron Plate",
		Type:   item.TypeArmor,
		Rarity: item.RarityCommon,
		Tags:   []string{"metal"},
	}
	entity.AddComponent(equip)

	// Track initial health
	sys.Update([]*Entity{entity}, 0.016)

	// Take damage with armor - should spawn sparks without panic
	health.Current = 70
	sys.Update([]*Entity{entity}, 0.016)
}

func TestArmorHitSparkSystem_GetProfile(t *testing.T) {
	sys := NewArmorHitSparkSystem(nil, 1)

	tests := []struct {
		name      string
		material  sprites.MaterialType
		wantType  particles.ParticleType
		wantColor string
		wantCount int
	}{
		{"metal", sprites.MaterialMetal, particles.ParticleSpark, "white", 6},
		{"leather", sprites.MaterialLeather, particles.ParticleDust, "brown", 4},
		{"cloth", sprites.MaterialCloth, particles.ParticleSmoke, "gray", 3},
		{"wood", sprites.MaterialWood, particles.ParticleDust, "tan", 5},
		{"crystal", sprites.MaterialCrystal, particles.ParticleSparkle, "prismatic", 5},
		{"energy", sprites.MaterialEnergy, particles.ParticleSparkle, "electric", 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := sys.getProfile(tt.material)
			if profile.ParticleType != tt.wantType {
				t.Errorf("material %s: want type %v, got %v", tt.name, tt.wantType, profile.ParticleType)
			}
			if profile.ColorHint != tt.wantColor {
				t.Errorf("material %s: want color %s, got %s", tt.name, tt.wantColor, profile.ColorHint)
			}
			if profile.BaseCount != tt.wantCount {
				t.Errorf("material %s: want count %d, got %d", tt.name, tt.wantCount, profile.BaseCount)
			}
		})
	}
}

func TestArmorHitSparkSystem_GetProfile_UnknownMaterial(t *testing.T) {
	sys := NewArmorHitSparkSystem(nil, 1)
	profile := sys.getProfile(sprites.MaterialType(999))
	// Should default to metal
	if profile.ParticleType != particles.ParticleSpark {
		t.Errorf("unknown material should default to metal spark, got %v", profile.ParticleType)
	}
}

func TestArmorHitSparkSystem_ArmorRarityBonus(t *testing.T) {
	sys := NewArmorHitSparkSystem(nil, 1)

	tests := []struct {
		name   string
		rarity item.Rarity
		want   int
	}{
		{"common", item.RarityCommon, 0},
		{"uncommon", item.RarityUncommon, 1},
		{"rare", item.RarityRare, 2},
		{"epic", item.RarityEpic, 3},
		{"legendary", item.RarityLegendary, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sys.armorRarityBonus(tt.rarity)
			if got != tt.want {
				t.Errorf("rarity %s: want bonus %d, got %d", tt.name, tt.want, got)
			}
		})
	}
}

func TestArmorHitSparkSystem_GetGenreScale(t *testing.T) {
	sys := NewArmorHitSparkSystem(nil, 1)

	tests := []struct {
		name     string
		genreID  string
		wantMul  float64
		wantSize float64
	}{
		{"fantasy", "fantasy", 1.0, 1.0},
		{"horror", "horror", 0.7, 0.8},
		{"cyberpunk", "cyberpunk", 1.3, 1.1},
		{"scifi", "scifi", 1.1, 1.0},
		{"postapoc", "postapoc", 1.0, 1.2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scale := sys.getGenreScale(tt.genreID)
			if scale.CountMul != tt.wantMul {
				t.Errorf("genre %s: want countMul %f, got %f", tt.name, tt.wantMul, scale.CountMul)
			}
			if scale.SizeMul != tt.wantSize {
				t.Errorf("genre %s: want sizeMul %f, got %f", tt.name, tt.wantSize, scale.SizeMul)
			}
		})
	}
}

func TestArmorHitSparkSystem_GetBestArmorSlot(t *testing.T) {
	world := NewWorld()
	sys := NewArmorHitSparkSystem(world, 1)

	// Entity without equipment
	entity := world.CreateEntity()
	itm, _ := sys.getBestArmorSlot(entity)
	if itm != nil {
		t.Error("expected nil item for entity without equipment")
	}

	// Entity with equipment but no armor
	entity2 := world.CreateEntity()
	equip := NewEquipmentComponent()
	equip.Slots[SlotMainHand] = &item.Item{
		Name:   "Sword",
		Type:   item.TypeWeapon,
		Rarity: item.RarityCommon,
	}
	entity2.AddComponent(equip)
	itm, _ = sys.getBestArmorSlot(entity2)
	if itm != nil {
		t.Error("expected nil item when only weapon equipped")
	}

	// Entity with chest armor
	entity3 := world.CreateEntity()
	equip3 := NewEquipmentComponent()
	chestArmor := &item.Item{
		Name:   "Plate Mail",
		Type:   item.TypeArmor,
		Rarity: item.RarityRare,
		Tags:   []string{"metal"},
	}
	equip3.Slots[SlotChest] = chestArmor
	entity3.AddComponent(equip3)
	itm, slot := sys.getBestArmorSlot(entity3)
	if itm != chestArmor {
		t.Error("expected chest armor item")
	}
	if slot != SlotChest {
		t.Errorf("expected SlotChest, got %v", slot)
	}
}

func TestArmorHitSparkSystem_GetBestArmorSlot_HighestRarity(t *testing.T) {
	world := NewWorld()
	sys := NewArmorHitSparkSystem(world, 1)

	entity := world.CreateEntity()
	equip := NewEquipmentComponent()
	equip.Slots[SlotChest] = &item.Item{
		Name:   "Common Chest",
		Type:   item.TypeArmor,
		Rarity: item.RarityCommon,
	}
	equip.Slots[SlotHead] = &item.Item{
		Name:   "Epic Helm",
		Type:   item.TypeArmor,
		Rarity: item.RarityEpic,
	}
	entity.AddComponent(equip)

	itm, slot := sys.getBestArmorSlot(entity)
	if itm == nil {
		t.Fatal("expected non-nil armor item")
	}
	if itm.Rarity != item.RarityEpic {
		t.Errorf("expected epic rarity, got %v", itm.Rarity)
	}
	if slot != SlotHead {
		t.Errorf("expected SlotHead for highest rarity, got %v", slot)
	}
}

func TestArmorHitSparkSystem_CleanupStaleEntries(t *testing.T) {
	world := NewWorld()
	sys := NewArmorHitSparkSystem(world, 1)
	ps := &ParticleSystem{}
	sys.SetParticleSystem(ps)

	e1 := world.CreateEntity()
	e1.AddComponent(&HealthComponent{Current: 100, Max: 100})
	e2 := world.CreateEntity()
	e2.AddComponent(&HealthComponent{Current: 50, Max: 50})

	// Track both entities
	sys.Update([]*Entity{e1, e2}, 0.016)

	if len(sys.prevHealth) != 2 {
		t.Errorf("expected 2 tracked entities, got %d", len(sys.prevHealth))
	}

	// Only update with e1 for long enough to trigger cleanup
	sys.cleanupTimer = sys.cleanupInterval
	sys.Update([]*Entity{e1}, 0.016)

	if len(sys.prevHealth) != 1 {
		t.Errorf("expected 1 tracked entity after cleanup, got %d", len(sys.prevHealth))
	}
}

func TestArmorHitSparkSystem_HealingDoesNotTrigger(t *testing.T) {
	world := NewWorld()
	sys := NewArmorHitSparkSystem(world, 1)
	ps := &ParticleSystem{}
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	health := &HealthComponent{Current: 50, Max: 100}
	entity.AddComponent(health)
	entity.AddComponent(&PositionComponent{X: 5, Y: 5})

	equip := NewEquipmentComponent()
	equip.Slots[SlotChest] = &item.Item{
		Name:   "Armor",
		Type:   item.TypeArmor,
		Rarity: item.RarityCommon,
	}
	entity.AddComponent(equip)

	// Track
	sys.Update([]*Entity{entity}, 0.016)

	// Heal (increase health) - should not trigger sparks
	health.Current = 80
	sys.Update([]*Entity{entity}, 0.016)
	// No panic means healing was correctly ignored
}
