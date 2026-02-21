package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/magic"
	"github.com/opd-ai/venture/pkg/rendering/palette"
)

func TestNewTimeOfDayManaCostSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayManaCostSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewTimeOfDayManaCostSystem returned nil")
	}
	if sys.world != world {
		t.Error("World reference not set correctly")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("Default genre = %s, want fantasy", sys.genreID)
	}
}

func TestTimeOfDayManaCostSystem_SetGenre(t *testing.T) {
	sys := NewTimeOfDayManaCostSystem(nil, 12345)
	sys.SetGenre("horror")

	if sys.genreID != "horror" {
		t.Errorf("Genre = %s, want horror", sys.genreID)
	}
}

func TestTimeOfDayManaCostSystem_SetLightingSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayManaCostSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)

	sys.SetLightingSystem(lightingSys)

	if sys.lightingSystem != lightingSys {
		t.Error("Lighting system reference not set correctly")
	}
}

func TestTimeOfDayManaCostSystem_BaseMultipliers(t *testing.T) {
	tests := []struct {
		name      string
		element   magic.ElementType
		timeOfDay palette.TimeOfDay
		wantLow   float64
		wantHigh  float64
	}{
		// Fire empowered by day (base + fantasy genre modifier -0.02)
		{"fire_day", magic.ElementFire, palette.TimeOfDayDay, 0.82, 0.84},
		{"fire_night", magic.ElementFire, palette.TimeOfDayNight, 1.12, 1.14},
		// Light empowered by day (base + fantasy genre modifier -0.02)
		{"light_day", magic.ElementLight, palette.TimeOfDayDay, 0.82, 0.84},
		{"light_night", magic.ElementLight, palette.TimeOfDayNight, 1.17, 1.19},
		// Dark empowered by night (base + fantasy genre modifier -0.02)
		{"dark_night", magic.ElementDark, palette.TimeOfDayNight, 0.82, 0.84},
		{"dark_day", magic.ElementDark, palette.TimeOfDayDay, 1.12, 1.14},
		// Earth empowered by dawn/dusk
		{"earth_dawn", magic.ElementEarth, palette.TimeOfDayDawn, 0.89, 0.91},
		{"earth_dusk", magic.ElementEarth, palette.TimeOfDayDusk, 0.89, 0.91},
		// None element: always 1.0
		{"none_any", magic.ElementNone, palette.TimeOfDayDay, 0.99, 1.01},
	}

	sys := NewTimeOfDayManaCostSystem(nil, 12345)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mult := sys.getMultiplier(tt.element, tt.timeOfDay)
			if mult < tt.wantLow || mult > tt.wantHigh {
				t.Errorf("getMultiplier(%v, %v) = %v, want [%v, %v]",
					tt.element, tt.timeOfDay, mult, tt.wantLow, tt.wantHigh)
			}
		})
	}
}

func TestTimeOfDayManaCostSystem_GenreModifiers(t *testing.T) {
	tests := []struct {
		name    string
		genre   string
		element magic.ElementType
		time    palette.TimeOfDay
		wantLow float64
		wantHi  float64
	}{
		// Horror boosts dark magic significantly
		{"horror_dark_night", "horror", magic.ElementDark, palette.TimeOfDayNight, 0.74, 0.76},
		// Horror penalizes light magic
		{"horror_light_day", "horror", magic.ElementLight, palette.TimeOfDayDay, 0.94, 0.96},
		// Cyberpunk boosts dark at night
		{"cyberpunk_dark_night", "cyberpunk", magic.ElementDark, palette.TimeOfDayNight, 0.79, 0.81},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewTimeOfDayManaCostSystem(nil, 12345)
			sys.SetGenre(tt.genre)

			mult := sys.getMultiplier(tt.element, tt.time)
			if mult < tt.wantLow || mult > tt.wantHi {
				t.Errorf("genre=%s element=%v time=%v: mult=%v, want [%v, %v]",
					tt.genre, tt.element, tt.time, mult, tt.wantLow, tt.wantHi)
			}
		})
	}
}

func TestTimeOfDayManaCostSystem_GetEffectiveManaCost(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayManaCostSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)
	sys.SetLightingSystem(lightingSys)

	tests := []struct {
		name     string
		element  magic.ElementType
		baseCost int
		wantMin  int
		wantMax  int
	}{
		{"fire_base_30", magic.ElementFire, 30, 20, 40},
		{"dark_base_50", magic.ElementDark, 50, 35, 65},
		{"none_base_20", magic.ElementNone, 20, 19, 21},
	}

	// Create an entity with spell slots
	entity := world.CreateEntity()
	entity.AddComponent(&SpellSlotComponent{})

	// Force update to populate cache
	sys.Update([]*Entity{entity}, 1.0)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spell := &magic.Spell{
				Element: tt.element,
				Stats:   magic.Stats{ManaCost: tt.baseCost},
			}
			cost := sys.GetEffectiveManaCost(entity.ID, spell)
			if cost < tt.wantMin || cost > tt.wantMax {
				t.Errorf("GetEffectiveManaCost() = %d, want [%d, %d]",
					cost, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestTimeOfDayManaCostSystem_NilSpell(t *testing.T) {
	sys := NewTimeOfDayManaCostSystem(nil, 12345)
	cost := sys.GetEffectiveManaCost(1, nil)
	if cost != 0 {
		t.Errorf("GetEffectiveManaCost(nil) = %d, want 0", cost)
	}
}

func TestTimeOfDayManaCostSystem_MinimumManaCost(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayManaCostSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)
	sys.SetLightingSystem(lightingSys)

	entity := world.CreateEntity()
	entity.AddComponent(&SpellSlotComponent{})
	sys.Update([]*Entity{entity}, 1.0)

	// Spell with very low base cost
	spell := &magic.Spell{
		Element: magic.ElementFire,
		Stats:   magic.Stats{ManaCost: 1},
	}

	cost := sys.GetEffectiveManaCost(entity.ID, spell)
	if cost < 1 {
		t.Errorf("Minimum mana cost should be 1, got %d", cost)
	}
}

func TestTimeOfDayManaCostSystem_Update(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayManaCostSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)
	sys.SetLightingSystem(lightingSys)

	// Create entity with spell slots
	entity := world.CreateEntity()
	entity.AddComponent(&SpellSlotComponent{})

	// First update should populate cache
	sys.Update([]*Entity{entity}, 1.0)

	// Check that cache was populated
	if _, ok := sys.elementMultiplierCache[entity.ID]; !ok {
		t.Error("Cache should be populated after Update")
	}
}

func TestTimeOfDayManaCostSystem_NoLightingSystem(t *testing.T) {
	sys := NewTimeOfDayManaCostSystem(nil, 12345)

	// Should not panic without lighting system
	sys.Update([]*Entity{}, 1.0)

	mult := sys.GetElementMultiplier(1, magic.ElementFire)
	if mult != 1.0 {
		t.Errorf("GetElementMultiplier without lighting = %v, want 1.0", mult)
	}

	desc := sys.GetBonusDescription()
	if desc != "" {
		t.Errorf("GetBonusDescription without lighting = %q, want empty", desc)
	}
}

func TestTimeOfDayManaCostSystem_GetCurrentMultiplierForElement(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayManaCostSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)
	sys.SetLightingSystem(lightingSys)

	mult := sys.GetCurrentMultiplierForElement(magic.ElementFire)
	if mult < 0.5 || mult > 1.5 {
		t.Errorf("GetCurrentMultiplierForElement returned %v, want [0.5, 1.5]", mult)
	}
}

func TestTimeOfDayManaCostSystem_GetBonusDescription(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayManaCostSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)
	sys.SetLightingSystem(lightingSys)

	desc := sys.GetBonusDescription()
	// Description should be non-empty when there's a bonus
	// (Depends on current time of day from lighting system)
	if desc != "" {
		// Verify it contains expected keywords
		if len(desc) < 5 {
			t.Errorf("Description too short: %q", desc)
		}
	}
}

func TestTimeOfDayManaCostSystem_ClampMultipliers(t *testing.T) {
	sys := NewTimeOfDayManaCostSystem(nil, 12345)

	// Test that multipliers are clamped
	for elem := magic.ElementNone; elem <= magic.ElementArcane; elem++ {
		for _, time := range []palette.TimeOfDay{
			palette.TimeOfDayDawn,
			palette.TimeOfDayDay,
			palette.TimeOfDayDusk,
			palette.TimeOfDayNight,
		} {
			mult := sys.getMultiplier(elem, time)
			if mult < 0.5 || mult > 1.5 {
				t.Errorf("Multiplier for %v at %v = %v, outside [0.5, 1.5]",
					elem, time, mult)
			}
		}
	}
}

func TestTimeOfDayManaCostSystem_AllElementsCovered(t *testing.T) {
	sys := NewTimeOfDayManaCostSystem(nil, 12345)

	// Verify all elements have entries in baseMultipliers
	elements := []magic.ElementType{
		magic.ElementNone,
		magic.ElementFire,
		magic.ElementIce,
		magic.ElementLightning,
		magic.ElementEarth,
		magic.ElementWind,
		magic.ElementLight,
		magic.ElementDark,
		magic.ElementArcane,
	}

	for _, elem := range elements {
		if _, ok := sys.baseMultipliers[elem]; !ok {
			t.Errorf("Element %v missing from baseMultipliers", elem)
		}
	}
}

func TestTimeOfDayManaCostSystem_EntityWithoutSpellSlots(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayManaCostSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)
	sys.SetLightingSystem(lightingSys)

	// Create entity WITHOUT spell slots
	entity := world.CreateEntity()

	// Update should not crash and should not cache this entity
	sys.Update([]*Entity{entity}, 1.0)

	if _, ok := sys.elementMultiplierCache[entity.ID]; ok {
		t.Error("Entity without spell slots should not be cached")
	}
}

func BenchmarkTimeOfDayManaCostSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewTimeOfDayManaCostSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)
	sys.SetLightingSystem(lightingSys)

	// Create 100 entities with spell slots
	entities := make([]*Entity, 100)
	for i := range entities {
		entities[i] = world.CreateEntity()
		entities[i].AddComponent(&SpellSlotComponent{})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}

func BenchmarkTimeOfDayManaCostSystem_GetEffectiveManaCost(b *testing.B) {
	world := NewWorld()
	sys := NewTimeOfDayManaCostSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)
	sys.SetLightingSystem(lightingSys)

	entity := world.CreateEntity()
	entity.AddComponent(&SpellSlotComponent{})
	sys.Update([]*Entity{entity}, 1.0)

	spell := &magic.Spell{
		Element: magic.ElementFire,
		Stats:   magic.Stats{ManaCost: 30},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.GetEffectiveManaCost(entity.ID, spell)
	}
}
