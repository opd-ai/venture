package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

func TestNewMeleeEnchantmentArcParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewMeleeEnchantmentArcParticleSystem(world, 42)

	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.world != world {
		t.Error("world not set")
	}
	if sys.rng == nil {
		t.Error("rng not initialized")
	}
	if sys.emittedArcs == nil {
		t.Error("emittedArcs map not initialized")
	}
}

func TestMeleeEnchantmentArcParticleSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genre   string
		wantMul float64
	}{
		{"fantasy", "fantasy", 1.0},
		{"scifi", "scifi", 0.5},
		{"horror", "horror", 1.5},
		{"cyberpunk", "cyberpunk", 0.3},
		{"postapoc", "postapoc", 1.3},
		{"unknown defaults to fantasy", "unknown", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewMeleeEnchantmentArcParticleSystem(world, 42)
			sys.SetGenre(tt.genre)

			if sys.genreID != tt.genre {
				t.Errorf("genreID = %q, want %q", sys.genreID, tt.genre)
			}
			if sys.genrePreset.GravityMul != tt.wantMul {
				t.Errorf("GravityMul = %v, want %v", sys.genrePreset.GravityMul, tt.wantMul)
			}
		})
	}
}

func TestMeleeEnchantmentArcParticleSystem_UpdateNilGuards(t *testing.T) {
	world := NewWorld()
	sys := NewMeleeEnchantmentArcParticleSystem(world, 42)

	// Should not panic with nil particle system
	sys.Update([]*Entity{}, 0.1)

	// Should not panic with nil world
	sys2 := NewMeleeEnchantmentArcParticleSystem(nil, 42)
	sys2.SetParticleSystem(&ParticleSystem{})
	sys2.Update([]*Entity{}, 0.1)
}

func TestMeleeEnchantmentArcParticleSystem_SkipsNoArc(t *testing.T) {
	world := NewWorld()
	sys := NewMeleeEnchantmentArcParticleSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Entity has no melee_swing_arc component - should be skipped
	sys.checkTimer = sys.checkInterval // force check
	sys.Update([]*Entity{entity}, 0.1)

	if len(sys.emittedArcs) != 0 {
		t.Error("should not have emitted for entity without arc")
	}
}

func TestMeleeEnchantmentArcParticleSystem_SkipsInactiveArc(t *testing.T) {
	world := NewWorld()
	sys := NewMeleeEnchantmentArcParticleSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&MeleeSwingArcComponent{Active: false})

	sys.checkTimer = sys.checkInterval
	sys.Update([]*Entity{entity}, 0.1)

	if sys.emittedArcs[entity.ID] {
		t.Error("should not emit for inactive arc")
	}
}

func TestMeleeEnchantmentArcParticleSystem_SkipsCommonWeapon(t *testing.T) {
	world := NewWorld()
	sys := NewMeleeEnchantmentArcParticleSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&MeleeSwingArcComponent{
		Active:        true,
		ArcAngleStart: 0,
		ArcAngleEnd:   1.57,
		ArcRadius:     20,
	})

	equip := NewEquipmentComponent()
	commonWeapon := &item.Item{
		Name:   "Rusty Sword",
		Rarity: item.RarityCommon,
		Stats:  item.Stats{IsProjectile: false},
	}
	equip.Equip(commonWeapon, SlotMainHand)
	entity.AddComponent(equip)

	sys.checkTimer = sys.checkInterval
	sys.Update([]*Entity{entity}, 0.1)

	if sys.emittedArcs[entity.ID] {
		t.Error("should not emit for common rarity weapon")
	}
}

func TestMeleeEnchantmentArcParticleSystem_SkipsProjectileWeapon(t *testing.T) {
	world := NewWorld()
	sys := NewMeleeEnchantmentArcParticleSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&MeleeSwingArcComponent{
		Active:        true,
		ArcAngleStart: 0,
		ArcAngleEnd:   1.57,
		ArcRadius:     20,
	})

	equip := NewEquipmentComponent()
	bowWeapon := &item.Item{
		Name:   "Epic Bow",
		Rarity: item.RarityEpic,
		Stats:  item.Stats{IsProjectile: true},
	}
	equip.Equip(bowWeapon, SlotMainHand)
	entity.AddComponent(equip)

	sys.checkTimer = sys.checkInterval
	sys.Update([]*Entity{entity}, 0.1)

	if sys.emittedArcs[entity.ID] {
		t.Error("should not emit for projectile weapon")
	}
}

func TestMeleeEnchantmentArcParticleSystem_GetWeaponRarity(t *testing.T) {
	tests := []struct {
		name       string
		rarity     item.Rarity
		projectile bool
		wantOK     bool
	}{
		{"common melee", item.RarityCommon, false, false},
		{"uncommon melee", item.RarityUncommon, false, true},
		{"rare melee", item.RarityRare, false, true},
		{"epic melee", item.RarityEpic, false, true},
		{"legendary melee", item.RarityLegendary, false, true},
		{"epic projectile", item.RarityEpic, true, false},
		{"no equipment", item.RarityCommon, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewMeleeEnchantmentArcParticleSystem(world, 42)

			entity := world.CreateEntity()

			if tt.name != "no equipment" {
				equip := NewEquipmentComponent()
				weapon := &item.Item{
					Name:   "Test Weapon",
					Rarity: tt.rarity,
					Stats:  item.Stats{IsProjectile: tt.projectile},
				}
				equip.Equip(weapon, SlotMainHand)
				entity.AddComponent(equip)
			}

			gotRarity, gotOK := sys.getWeaponRarity(entity)
			if gotOK != tt.wantOK {
				t.Errorf("getWeaponRarity() ok = %v, want %v", gotOK, tt.wantOK)
			}
			if gotOK && gotRarity != tt.rarity {
				t.Errorf("getWeaponRarity() rarity = %v, want %v", gotRarity, tt.rarity)
			}
		})
	}
}

func TestMeleeEnchantmentArcParticleSystem_EmitsOncePerArc(t *testing.T) {
	world := NewWorld()
	sys := NewMeleeEnchantmentArcParticleSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	entity.AddComponent(&MeleeSwingArcComponent{
		Active:        true,
		ArcAngleStart: 0,
		ArcAngleEnd:   1.57,
		ArcRadius:     20,
	})

	equip := NewEquipmentComponent()
	weapon := &item.Item{
		Name:   "Enchanted Blade",
		Rarity: item.RarityRare,
		Stats:  item.Stats{IsProjectile: false},
	}
	equip.Equip(weapon, SlotMainHand)
	entity.AddComponent(equip)

	entities := []*Entity{entity}

	// First update: should emit
	sys.checkTimer = sys.checkInterval
	sys.Update(entities, 0.1)
	if !sys.emittedArcs[entity.ID] {
		t.Error("expected emittedArcs set after first active arc")
	}

	// Second update: should not emit again (already emitted)
	sys.checkTimer = sys.checkInterval
	sys.Update(entities, 0.1)
	// Still marked
	if !sys.emittedArcs[entity.ID] {
		t.Error("emittedArcs should remain set")
	}
}

func TestMeleeEnchantmentArcParticleSystem_ClearsOnArcEnd(t *testing.T) {
	world := NewWorld()
	sys := NewMeleeEnchantmentArcParticleSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	arc := &MeleeSwingArcComponent{
		Active:        true,
		ArcAngleStart: 0,
		ArcAngleEnd:   1.57,
		ArcRadius:     20,
	}
	entity.AddComponent(arc)

	equip := NewEquipmentComponent()
	weapon := &item.Item{
		Name:   "Enchanted Blade",
		Rarity: item.RarityEpic,
		Stats:  item.Stats{IsProjectile: false},
	}
	equip.Equip(weapon, SlotMainHand)
	entity.AddComponent(equip)

	entities := []*Entity{entity}

	// Emit
	sys.checkTimer = sys.checkInterval
	sys.Update(entities, 0.1)
	if !sys.emittedArcs[entity.ID] {
		t.Fatal("expected emittedArcs set")
	}

	// Deactivate arc
	arc.Active = false
	sys.checkTimer = sys.checkInterval
	sys.Update(entities, 0.1)

	if sys.emittedArcs[entity.ID] {
		t.Error("emittedArcs should be cleared when arc is inactive")
	}
}

func TestMeleeEnchantmentArcParticleSystem_ThrottleCheck(t *testing.T) {
	world := NewWorld()
	sys := NewMeleeEnchantmentArcParticleSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	entity.AddComponent(&MeleeSwingArcComponent{
		Active:        true,
		ArcAngleStart: 0,
		ArcAngleEnd:   1.57,
		ArcRadius:     20,
	})

	equip := NewEquipmentComponent()
	weapon := &item.Item{
		Name:   "Enchanted Blade",
		Rarity: item.RarityLegendary,
		Stats:  item.Stats{IsProjectile: false},
	}
	equip.Equip(weapon, SlotMainHand)
	entity.AddComponent(equip)

	entities := []*Entity{entity}

	// Update with very small delta - should not exceed check interval
	sys.Update(entities, 0.001)
	if sys.emittedArcs[entity.ID] {
		t.Error("should not emit before check interval reached")
	}
}

func TestEnchantArcRarityConfigs(t *testing.T) {
	tests := []struct {
		rarity    item.Rarity
		wantColor string
		wantCount int
	}{
		{item.RarityUncommon, "green", 2},
		{item.RarityRare, "blue", 4},
		{item.RarityEpic, "purple", 6},
		{item.RarityLegendary, "gold", 10},
	}

	for _, tt := range tests {
		t.Run(tt.rarity.String(), func(t *testing.T) {
			conf, ok := enchantArcRarityConfigs[tt.rarity]
			if !ok {
				t.Fatalf("missing config for rarity %s", tt.rarity)
			}
			if conf.ColorHint != tt.wantColor {
				t.Errorf("ColorHint = %q, want %q", conf.ColorHint, tt.wantColor)
			}
			if conf.ParticleCount != tt.wantCount {
				t.Errorf("ParticleCount = %d, want %d", conf.ParticleCount, tt.wantCount)
			}
			if conf.Intensity <= 0 || conf.Intensity > 1.0 {
				t.Errorf("Intensity = %v, want (0,1]", conf.Intensity)
			}
			if conf.SpreadFactor <= 0 {
				t.Errorf("SpreadFactor = %v, want > 0", conf.SpreadFactor)
			}
		})
	}
}

func TestEnchantArcGenrePresets(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, g := range genres {
		t.Run(g, func(t *testing.T) {
			preset, ok := enchantArcGenrePresets[g]
			if !ok {
				t.Fatalf("missing preset for genre %s", g)
			}
			if preset.GravityMul <= 0 {
				t.Errorf("GravityMul = %v, want > 0", preset.GravityMul)
			}
			if preset.SizeMul <= 0 {
				t.Errorf("SizeMul = %v, want > 0", preset.SizeMul)
			}
			if preset.DurationMul <= 0 {
				t.Errorf("DurationMul = %v, want > 0", preset.DurationMul)
			}
		})
	}
}

func BenchmarkMeleeEnchantmentArcParticleSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewMeleeEnchantmentArcParticleSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	entities := make([]*Entity, 100)
	for i := range entities {
		e := world.CreateEntity()
		e.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		e.AddComponent(&MeleeSwingArcComponent{Active: false})
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.checkTimer = 0
		sys.Update(entities, 0.016)
	}
}
