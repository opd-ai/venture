package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

func TestNewWeaponMaterialParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponMaterialParticleSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewWeaponMaterialParticleSystem returned nil")
	}
	if sys.world != world {
		t.Error("world not set")
	}
	if sys.seed != 12345 {
		t.Errorf("seed = %d, want 12345", sys.seed)
	}
	if sys.rng == nil {
		t.Error("rng not initialized")
	}
	if sys.cooldowns == nil {
		t.Error("cooldowns map not initialized")
	}
	if len(sys.materialConfigs) != 6 {
		t.Errorf("materialConfigs count = %d, want 6", len(sys.materialConfigs))
	}
	if sys.baseInterval != 2.0 {
		t.Errorf("baseInterval = %f, want 2.0", sys.baseInterval)
	}
}

func TestWeaponMaterialParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponMaterialParticleSystem(world, 42)
	ps := &ParticleSystem{}
	sys.SetParticleSystem(ps)
	if sys.particleSystem != ps {
		t.Error("particle system not set")
	}
}

func TestWeaponMaterialParticleSystem_SetGenre(t *testing.T) {
	tests := []struct {
		genre      string
		wantFactor float64
	}{
		{"fantasy", 1.0},
		{"sci-fi", 0.8},
		{"cyberpunk", 0.7},
		{"horror", 1.4},
		{"post-apocalyptic", 1.3},
		{"unknown", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewWeaponMaterialParticleSystem(world, 42)
			sys.SetGenre(tt.genre)
			if sys.genreID != tt.genre {
				t.Errorf("genreID = %s, want %s", sys.genreID, tt.genre)
			}
			if sys.genreMultiplier != tt.wantFactor {
				t.Errorf("genreMultiplier = %f, want %f", sys.genreMultiplier, tt.wantFactor)
			}
		})
	}
}

func TestWeaponMaterialParticleSystem_UpdateSkipsWithoutParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponMaterialParticleSystem(world, 42)
	entity := world.CreateEntity()
	entity.AddComponent(NewEquipmentComponent())
	// Should not panic with nil particle system
	sys.Update([]*Entity{entity}, 0.016)
}

func TestWeaponMaterialParticleSystem_UpdateSkipsWithoutEquipment(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponMaterialParticleSystem(world, 42)
	sys.SetParticleSystem(&ParticleSystem{})
	entity := world.CreateEntity()
	// Entity has no equipment component; should be skipped
	sys.Update([]*Entity{entity}, 0.016)
	if _, exists := sys.cooldowns[entity.ID]; exists {
		t.Error("cooldown should not be set for entity without equipment")
	}
}

func TestWeaponMaterialParticleSystem_GetWeaponMaterial(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Entity)
		wantMat sprites.MaterialType
		wantHas bool
	}{
		{
			name:    "no_equipment_component",
			setup:   func(e *Entity) {},
			wantMat: sprites.MaterialMetal,
			wantHas: false,
		},
		{
			name: "empty_equipment",
			setup: func(e *Entity) {
				e.AddComponent(NewEquipmentComponent())
			},
			wantMat: sprites.MaterialMetal,
			wantHas: false,
		},
		{
			name: "sword_equipped",
			setup: func(e *Entity) {
				eq := NewEquipmentComponent()
				eq.Slots[SlotMainHand] = &item.Item{
					Name:       "Iron Sword",
					Type:       item.TypeWeapon,
					WeaponType: item.WeaponSword,
				}
				e.AddComponent(eq)
			},
			wantMat: sprites.MaterialMetal,
			wantHas: true,
		},
		{
			name: "staff_equipped",
			setup: func(e *Entity) {
				eq := NewEquipmentComponent()
				eq.Slots[SlotMainHand] = &item.Item{
					Name:       "Oak Staff",
					Type:       item.TypeWeapon,
					WeaponType: item.WeaponStaff,
				}
				e.AddComponent(eq)
			},
			wantMat: sprites.MaterialWood,
			wantHas: true,
		},
		{
			name: "crystal_tag_weapon",
			setup: func(e *Entity) {
				eq := NewEquipmentComponent()
				eq.Slots[SlotMainHand] = &item.Item{
					Name:       "Crystal Blade",
					Type:       item.TypeWeapon,
					WeaponType: item.WeaponSword,
					Tags:       []string{"crystal"},
				}
				e.AddComponent(eq)
			},
			wantMat: sprites.MaterialCrystal,
			wantHas: true,
		},
		{
			name: "energy_tag_weapon",
			setup: func(e *Entity) {
				eq := NewEquipmentComponent()
				eq.Slots[SlotMainHand] = &item.Item{
					Name:       "Arcane Wand",
					Type:       item.TypeWeapon,
					WeaponType: item.WeaponWand,
					Tags:       []string{"energy"},
				}
				e.AddComponent(eq)
			},
			wantMat: sprites.MaterialEnergy,
			wantHas: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewWeaponMaterialParticleSystem(world, 42)
			sys.SetGenre("fantasy")
			entity := world.CreateEntity()
			tt.setup(entity)

			gotMat, gotHas := sys.getWeaponMaterial(entity)
			if gotHas != tt.wantHas {
				t.Errorf("hasMaterial = %v, want %v", gotHas, tt.wantHas)
			}
			if gotHas && gotMat != tt.wantMat {
				t.Errorf("material = %v, want %v", gotMat, tt.wantMat)
			}
		})
	}
}

func TestWeaponMaterialParticleSystem_CooldownTracking(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponMaterialParticleSystem(world, 42)
	sys.SetParticleSystem(NewParticleSystem())
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})
	eq := NewEquipmentComponent()
	eq.Slots[SlotMainHand] = &item.Item{
		Name:       "Iron Sword",
		Type:       item.TypeWeapon,
		WeaponType: item.WeaponSword,
	}
	entity.AddComponent(eq)

	// First update sets cooldown
	sys.Update([]*Entity{entity}, 0.016)
	cd, exists := sys.cooldowns[entity.ID]
	if !exists {
		t.Fatal("cooldown should be set after first update")
	}
	if cd <= 0 {
		t.Errorf("cooldown should be positive, got %f", cd)
	}

	// Second update should tick down cooldown but not respawn
	prevCD := cd
	sys.Update([]*Entity{entity}, 0.1)
	newCD := sys.cooldowns[entity.ID]
	if newCD >= prevCD {
		t.Errorf("cooldown should decrease: was %f, now %f", prevCD, newCD)
	}
}

func TestWeaponMaterialParticleSystem_CooldownCleanup(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponMaterialParticleSystem(world, 42)
	sys.SetParticleSystem(NewParticleSystem())
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})
	eq := NewEquipmentComponent()
	eq.Slots[SlotMainHand] = &item.Item{
		Name:       "Iron Sword",
		Type:       item.TypeWeapon,
		WeaponType: item.WeaponSword,
	}
	entity.AddComponent(eq)

	// First update creates cooldown
	sys.Update([]*Entity{entity}, 0.016)
	if _, exists := sys.cooldowns[entity.ID]; !exists {
		t.Fatal("cooldown should exist")
	}

	// Remove weapon
	eq.Slots = make(map[EquipmentSlot]*item.Item)

	// Update should clean up cooldown
	sys.Update([]*Entity{entity}, 0.016)
	if _, exists := sys.cooldowns[entity.ID]; exists {
		t.Error("cooldown should be cleaned up when weapon is removed")
	}
}

func TestBuildMaterialConfigs(t *testing.T) {
	configs := buildMaterialConfigs()

	materials := []sprites.MaterialType{
		sprites.MaterialMetal,
		sprites.MaterialCrystal,
		sprites.MaterialEnergy,
		sprites.MaterialWood,
		sprites.MaterialLeather,
		sprites.MaterialCloth,
	}

	for _, mat := range materials {
		t.Run(mat.String(), func(t *testing.T) {
			cfg, ok := configs[mat]
			if !ok {
				t.Fatalf("no config for material %s", mat.String())
			}
			if cfg.Count <= 0 {
				t.Errorf("count should be positive, got %d", cfg.Count)
			}
			if cfg.Duration <= 0 {
				t.Errorf("duration should be positive, got %f", cfg.Duration)
			}
			if cfg.ColorHint == "" {
				t.Error("color hint should not be empty")
			}
		})
	}
}

func TestWeaponMaterialParticleSystem_NilWorld(t *testing.T) {
	sys := NewWeaponMaterialParticleSystem(nil, 42)
	if sys == nil {
		t.Fatal("should handle nil world")
	}
	// Should not panic
	sys.Update(nil, 0.016)
}

func TestWeaponMaterialParticleSystem_GenreMultiplierAffectsCooldown(t *testing.T) {
	tests := []struct {
		genre    string
		wantMore bool // true if cooldown should be higher than fantasy
	}{
		{"fantasy", false},
		{"horror", true},
		{"post-apocalyptic", true},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewWeaponMaterialParticleSystem(world, 42)
			sys.SetParticleSystem(NewParticleSystem())
			sys.SetGenre(tt.genre)

			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: 10, Y: 20})
			eq := NewEquipmentComponent()
			eq.Slots[SlotMainHand] = &item.Item{
				Name:       "Iron Sword",
				Type:       item.TypeWeapon,
				WeaponType: item.WeaponSword,
			}
			entity.AddComponent(eq)

			sys.Update([]*Entity{entity}, 0.016)
			cd := sys.cooldowns[entity.ID]

			// Fantasy baseline
			fantasySys := NewWeaponMaterialParticleSystem(world, 42)
			fantasySys.SetParticleSystem(NewParticleSystem())
			fantasySys.SetGenre("fantasy")

			fantasyEntity := world.CreateEntity()
			fantasyEntity.AddComponent(&PositionComponent{X: 10, Y: 20})
			feq := NewEquipmentComponent()
			feq.Slots[SlotMainHand] = &item.Item{
				Name:       "Iron Sword",
				Type:       item.TypeWeapon,
				WeaponType: item.WeaponSword,
			}
			fantasyEntity.AddComponent(feq)

			fantasySys.Update([]*Entity{fantasyEntity}, 0.016)
			fantasyCD := fantasySys.cooldowns[fantasyEntity.ID]

			if tt.wantMore && cd <= fantasyCD {
				t.Errorf("genre %s should have longer cooldown than fantasy: got %f vs %f", tt.genre, cd, fantasyCD)
			}
		})
	}
}
