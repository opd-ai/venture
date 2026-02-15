package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

func TestEquipmentEnchantmentGlowParticleSystem_Creation(t *testing.T) {
	tests := []struct {
		name  string
		seed  int64
		world *World
	}{
		{"with_world", 12345, NewWorld()},
		{"nil_world", 99999, nil},
		{"zero_seed", 0, NewWorld()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewEquipmentEnchantmentGlowParticleSystem(tt.world, tt.seed)
			if sys == nil {
				t.Fatal("expected non-nil system")
			}
			if sys.seed != tt.seed {
				t.Errorf("seed = %d, want %d", sys.seed, tt.seed)
			}
			if sys.baseInterval <= 0 {
				t.Errorf("baseInterval should be positive, got %f", sys.baseInterval)
			}
		})
	}
}

func TestEquipmentEnchantmentGlowParticleSystem_SetParticleSystem(t *testing.T) {
	sys := NewEquipmentEnchantmentGlowParticleSystem(NewWorld(), 42)
	if sys.particleSystem != nil {
		t.Error("expected nil particleSystem before Set")
	}

	ps := &ParticleSystem{}
	sys.SetParticleSystem(ps)
	if sys.particleSystem != ps {
		t.Error("particleSystem not set correctly")
	}
}

func TestEquipmentEnchantmentGlowParticleSystem_SetGenre(t *testing.T) {
	sys := NewEquipmentEnchantmentGlowParticleSystem(NewWorld(), 42)

	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	for _, g := range genres {
		sys.SetGenre(g)
		if sys.genreID != g {
			t.Errorf("genreID = %s, want %s", sys.genreID, g)
		}
	}
}

func TestEquipmentEnchantmentGlowParticleSystem_GetHighestEquippedRarity(t *testing.T) {
	tests := []struct {
		name         string
		slots        map[EquipmentSlot]*item.Item
		wantRarity   item.Rarity
		wantHasRare  bool
	}{
		{
			name:        "no_equipment",
			slots:       map[EquipmentSlot]*item.Item{},
			wantRarity:  item.RarityCommon,
			wantHasRare: false,
		},
		{
			name: "common_only",
			slots: map[EquipmentSlot]*item.Item{
				SlotMainHand: {Name: "Stick", Rarity: item.RarityCommon},
			},
			wantRarity:  item.RarityCommon,
			wantHasRare: false,
		},
		{
			name: "uncommon_weapon",
			slots: map[EquipmentSlot]*item.Item{
				SlotMainHand: {Name: "Fine Sword", Rarity: item.RarityUncommon},
			},
			wantRarity:  item.RarityUncommon,
			wantHasRare: true,
		},
		{
			name: "mixed_rarity_highest_epic",
			slots: map[EquipmentSlot]*item.Item{
				SlotMainHand: {Name: "Sword", Rarity: item.RarityCommon},
				SlotChest:    {Name: "Dragon Plate", Rarity: item.RarityEpic},
				SlotBoots:    {Name: "Boots", Rarity: item.RarityRare},
			},
			wantRarity:  item.RarityEpic,
			wantHasRare: true,
		},
		{
			name: "legendary_accessory",
			slots: map[EquipmentSlot]*item.Item{
				SlotAccessory1: {Name: "Ring of Power", Rarity: item.RarityLegendary},
			},
			wantRarity:  item.RarityLegendary,
			wantHasRare: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewEquipmentEnchantmentGlowParticleSystem(world, 42)

			entity := world.CreateEntity()
			equipComp := NewEquipmentComponent()
			for slot, itm := range tt.slots {
				equipComp.Slots[slot] = itm
			}
			entity.AddComponent(equipComp)

			gotRarity, gotHasRare := sys.getHighestEquippedRarity(entity)
			if gotHasRare != tt.wantHasRare {
				t.Errorf("hasRareItem = %v, want %v", gotHasRare, tt.wantHasRare)
			}
			if gotHasRare && gotRarity != tt.wantRarity {
				t.Errorf("rarity = %v, want %v", gotRarity, tt.wantRarity)
			}
		})
	}
}

func TestEquipmentEnchantmentGlowParticleSystem_UpdateSkipsWithoutParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentEnchantmentGlowParticleSystem(world, 42)
	// No particle system set - Update should be a no-op
	entity := world.CreateEntity()
	entity.AddComponent(NewEquipmentComponent())
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})
	sys.Update([]*Entity{entity}, 0.016)
	// Should not panic
}

func TestEquipmentEnchantmentGlowParticleSystem_CooldownTracking(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentEnchantmentGlowParticleSystem(world, 42)

	entity := world.CreateEntity()
	equipComp := NewEquipmentComponent()
	equipComp.Slots[SlotMainHand] = &item.Item{Name: "Rare Sword", Rarity: item.RarityRare}
	entity.AddComponent(equipComp)
	entity.AddComponent(&PositionComponent{X: 5, Y: 5})

	// After first check with no particle system, cooldown should not be set
	sys.Update([]*Entity{entity}, 0.016)
	if _, exists := sys.cooldowns[entity.ID]; exists {
		t.Error("cooldown should not be set when particle system is nil")
	}
}

func TestEquipmentEnchantmentGlowParticleSystem_CleanupCooldownOnDowngrade(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentEnchantmentGlowParticleSystem(world, 42)
	sys.SetParticleSystem(&ParticleSystem{})

	entity := world.CreateEntity()
	equipComp := NewEquipmentComponent()
	entity.AddComponent(equipComp)
	entity.AddComponent(&PositionComponent{X: 5, Y: 5})

	// Manually set a cooldown as if entity previously had rare items
	sys.cooldowns[entity.ID] = 1.0

	// Entity has no rare items -> cooldown should be cleaned up
	sys.Update([]*Entity{entity}, 0.016)
	if _, exists := sys.cooldowns[entity.ID]; exists {
		t.Error("cooldown should be cleaned up when entity has no rare items")
	}
}

func TestEquipmentEnchantmentGlowParticleSystem_EnchantmentGlowMapping(t *testing.T) {
	// Verify the sprites.GetEnchantmentFromRarity mapping matches expected values
	tests := []struct {
		rarity    string
		wantColor string
		wantActive bool
	}{
		{"common", "white", false},
		{"uncommon", "green", true},
		{"rare", "blue", true},
		{"epic", "purple", true},
		{"legendary", "gold", true},
	}

	for _, tt := range tests {
		t.Run(tt.rarity, func(t *testing.T) {
			glow := sprites.GetEnchantmentFromRarity(tt.rarity)
			if glow.Active != tt.wantActive {
				t.Errorf("Active = %v, want %v for rarity %s", glow.Active, tt.wantActive, tt.rarity)
			}
			if glow.Active && glow.Color != tt.wantColor {
				t.Errorf("Color = %s, want %s for rarity %s", glow.Color, tt.wantColor, tt.rarity)
			}
			if glow.Active && glow.ParticleCount <= 0 {
				t.Errorf("ParticleCount should be > 0 for active glow, got %d", glow.ParticleCount)
			}
			if glow.Active && glow.PulseSpeed <= 0 {
				t.Errorf("PulseSpeed should be > 0 for active glow, got %f", glow.PulseSpeed)
			}
		})
	}
}

func TestEquipmentEnchantmentGlowParticleSystem_GenreColorModifier(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentEnchantmentGlowParticleSystem(world, 42)

	tests := []struct {
		genre string
		want  string
	}{
		{"horror", "dark"},
		{"cyberpunk", "neon"},
		{"sci-fi", "electric"},
		{"fantasy", ""},
		{"post-apocalyptic", ""},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys.SetGenre(tt.genre)
			got := sys.getGenreColorModifier()
			if got != tt.want {
				t.Errorf("getGenreColorModifier() = %s, want %s", got, tt.want)
			}
		})
	}
}
