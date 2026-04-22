package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

func TestNewWeaponMaterialImpactParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponMaterialImpactParticleSystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.seed != 42 {
		t.Errorf("expected seed 42, got %d", sys.seed)
	}
	if sys.rng == nil {
		t.Fatal("expected non-nil rng")
	}
}

func TestNewWeaponMaterialImpactParticleSystem_NilWorld(t *testing.T) {
	sys := NewWeaponMaterialImpactParticleSystem(nil, 99)
	if sys == nil {
		t.Fatal("expected non-nil system even with nil world")
	}
	if sys.logger != nil {
		t.Error("expected nil logger with nil world")
	}
}

func TestWeaponMaterialImpactParticleSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponMaterialImpactParticleSystem(world, 1)
	sys.SetGenre("fantasy")
	if sys.genreID != "fantasy" {
		t.Errorf("expected genre fantasy, got %s", sys.genreID)
	}
}

func TestWeaponMaterialImpactParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponMaterialImpactParticleSystem(world, 1)
	ps := &ParticleSystem{}
	sys.SetParticleSystem(ps)
	if sys.particleSystem != ps {
		t.Error("expected particle system to be set")
	}
}

func TestWeaponMaterialImpactParticleSystem_Update(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponMaterialImpactParticleSystem(world, 1)
	// Update should be a no-op and not panic
	sys.Update(nil, 0.016)
	sys.Update([]*Entity{}, 0.016)
}

func TestWeaponMaterialImpactParticleSystem_OnMeleeImpact_NilGuards(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponMaterialImpactParticleSystem(world, 1)
	// All nil guards should prevent panics
	sys.OnMeleeImpact(nil, nil, 10)
	entity := world.CreateEntity()
	sys.OnMeleeImpact(entity, nil, 10)
	sys.OnMeleeImpact(nil, entity, 10)
}

func TestWeaponMaterialImpactParticleSystem_OnMeleeImpact_NoWeapon(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponMaterialImpactParticleSystem(world, 1)
	ps := &ParticleSystem{}
	sys.SetParticleSystem(ps)

	attacker := world.CreateEntity()
	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 10, Y: 20})

	// No equipment component - should not panic
	sys.OnMeleeImpact(attacker, target, 25)
}

func TestWeaponMaterialImpactParticleSystem_OnMeleeImpact_ProjectileSkipped(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponMaterialImpactParticleSystem(world, 1)
	ps := &ParticleSystem{}
	sys.SetParticleSystem(ps)

	attacker := world.CreateEntity()
	equip := &EquipmentComponent{Slots: make(map[EquipmentSlot]*item.Item)}
	bow := &item.Item{
		Name:   "Bow",
		Rarity: item.RarityCommon,
		Stats:  item.Stats{IsProjectile: true},
		Tags:   []string{"wood"},
	}
	equip.Slots[SlotMainHand] = bow
	attacker.AddComponent(equip)

	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 10, Y: 20})

	// Projectile weapon should be skipped
	sys.OnMeleeImpact(attacker, target, 15)
}

func TestWeaponMaterialImpactParticleSystem_OnMeleeImpact_SpawnsParticles(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponMaterialImpactParticleSystem(world, 1)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	attacker := world.CreateEntity()
	equip := &EquipmentComponent{Slots: make(map[EquipmentSlot]*item.Item)}
	equip.Slots[SlotMainHand] = &item.Item{
		ID:     "impact_sword",
		Name:   "Impact Sword",
		Type:   item.TypeWeapon,
		Rarity: item.RarityRare,
		Stats:  item.Stats{IsProjectile: false},
		Tags:   []string{"metal"},
	}
	attacker.AddComponent(equip)

	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 10, Y: 20})
	world.Update(0)

	before := ps.GetActiveParticleCount()
	sys.OnMeleeImpact(attacker, target, 25)
	after := ps.GetActiveParticleCount()

	if after <= before {
		t.Fatalf("expected melee impact to spawn particles (before=%d after=%d)", before, after)
	}
}

func TestWeaponMaterialImpactParticleSystem_GetProfile(t *testing.T) {
	sys := NewWeaponMaterialImpactParticleSystem(nil, 1)

	tests := []struct {
		name         string
		material     sprites.MaterialType
		wantType     particles.ParticleType
		wantColor    string
		wantMinCount int
	}{
		{"metal", sprites.MaterialMetal, particles.ParticleSpark, "orange", 8},
		{"crystal", sprites.MaterialCrystal, particles.ParticleSparkle, "cyan", 6},
		{"energy", sprites.MaterialEnergy, particles.ParticleSparkle, "purple", 7},
		{"wood", sprites.MaterialWood, particles.ParticleDust, "brown", 5},
		{"leather", sprites.MaterialLeather, particles.ParticleDust, "tan", 4},
		{"cloth", sprites.MaterialCloth, particles.ParticleSmoke, "white", 3},
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
			if profile.BaseCount != tt.wantMinCount {
				t.Errorf("material %s: want count %d, got %d", tt.name, tt.wantMinCount, profile.BaseCount)
			}
		})
	}
}

func TestWeaponMaterialImpactParticleSystem_GetProfile_Unknown(t *testing.T) {
	sys := NewWeaponMaterialImpactParticleSystem(nil, 1)
	// Unknown material should fallback to metal
	profile := sys.getProfile(sprites.MaterialType(99))
	metalProfile := sys.getProfile(sprites.MaterialMetal)
	if profile.ParticleType != metalProfile.ParticleType {
		t.Error("unknown material should fallback to metal profile")
	}
}

func TestWeaponMaterialImpactParticleSystem_RarityBonus(t *testing.T) {
	sys := NewWeaponMaterialImpactParticleSystem(nil, 1)

	tests := []struct {
		name   string
		rarity item.Rarity
		want   int
	}{
		{"common", item.RarityCommon, 0},
		{"uncommon", item.RarityUncommon, 1},
		{"rare", item.RarityRare, 2},
		{"epic", item.RarityEpic, 4},
		{"legendary", item.RarityLegendary, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sys.rarityBonus(tt.rarity)
			if got != tt.want {
				t.Errorf("rarity %s: want bonus %d, got %d", tt.name, tt.want, got)
			}
		})
	}
}

func TestWeaponMaterialImpactParticleSystem_SpawnImpactEffect_NilGuards(t *testing.T) {
	sys := NewWeaponMaterialImpactParticleSystem(nil, 1)
	// Should not panic with nil particle system or world
	sys.SpawnImpactEffect(10, 20, sprites.MaterialMetal, item.RarityCommon, 10)
}

func TestWeaponMaterialImpactParticleSystem_MaterialProfileCoverage(t *testing.T) {
	// Verify all defined materials have valid profiles
	allMaterials := []sprites.MaterialType{
		sprites.MaterialMetal, sprites.MaterialLeather, sprites.MaterialCloth,
		sprites.MaterialWood, sprites.MaterialCrystal, sprites.MaterialEnergy,
	}
	for _, mat := range allMaterials {
		profile, ok := materialImpactProfiles[mat]
		if !ok {
			t.Errorf("material %v missing from materialImpactProfiles", mat)
			continue
		}
		if profile.Duration <= 0 {
			t.Errorf("material %v has non-positive duration", mat)
		}
		if profile.BaseCount <= 0 {
			t.Errorf("material %v has non-positive base count", mat)
		}
		if profile.MaxSize <= profile.MinSize {
			t.Errorf("material %v: maxSize %f <= minSize %f", mat, profile.MaxSize, profile.MinSize)
		}
	}
}

func TestWeaponMaterialImpactParticleSystem_ParticleCountCap(t *testing.T) {
	sys := NewWeaponMaterialImpactParticleSystem(nil, 1)
	// Legendary rarity (+6) on Metal (base 8) with high damage (+2) = 16, under cap of 24
	bonus := sys.rarityBonus(item.RarityLegendary)
	metalProfile := sys.getProfile(sprites.MaterialMetal)
	total := metalProfile.BaseCount + bonus + 2 // +2 for high damage
	if total > 24 {
		t.Errorf("particle count %d exceeds cap 24", total)
	}
}
