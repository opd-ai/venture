package engine

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/procgen/magic"
	"github.com/opd-ai/venture/pkg/rendering/palette"
)

// stubSpellDamageClock implements GameClock for testing.
type stubSpellDamageClock struct {
	now time.Time
}

func (c *stubSpellDamageClock) Now() time.Time {
	return c.now
}

func TestNewTimeOfDaySpellDamageSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDaySpellDamageSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewTimeOfDaySpellDamageSystem returned nil")
	}
	if sys.world != world {
		t.Error("world not set correctly")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got '%s'", sys.genreID)
	}
	if sys.modCache == nil {
		t.Error("modCache not initialized")
	}
}

func TestTimeOfDaySpellDamageSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDaySpellDamageSystem(world, 12345)

	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("expected genre 'horror', got '%s'", sys.genreID)
	}
}

func TestTimeOfDaySpellDamageSystem_SetLightingSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDaySpellDamageSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)

	sys.SetLightingSystem(lightingSys)
	if sys.lightingSystem != lightingSys {
		t.Error("lighting system not set correctly")
	}
}

func TestTimeOfDaySpellDamageSystem_GetDamageModifier_NoLighting(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDaySpellDamageSystem(world, 12345)

	// Without lighting system, should return 1.0
	mod := sys.GetDamageModifier(magic.ElementFire)
	if mod != 1.0 {
		t.Errorf("expected modifier 1.0 without lighting, got %f", mod)
	}
}

func TestTimeOfDaySpellDamageSystem_FireDamageDay(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDaySpellDamageSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)

	// Set noon for day time
	clock := &stubSpellDamageClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	lightingSys.SetClock(clock)
	lightingSys.Update(nil, 0.1)
	sys.SetLightingSystem(lightingSys)

	mod := sys.GetDamageModifier(magic.ElementFire)
	if mod < 1.10 || mod > 1.20 {
		t.Errorf("fire damage during day should be ~1.15, got %f", mod)
	}
}

func TestTimeOfDaySpellDamageSystem_DarkDamageNight(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDaySpellDamageSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)

	// Set 22:00 for night time
	clock := &stubSpellDamageClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lightingSys.SetClock(clock)
	lightingSys.Update(nil, 0.1)
	sys.SetLightingSystem(lightingSys)

	mod := sys.GetDamageModifier(magic.ElementDark)
	if mod < 1.15 || mod > 1.25 {
		t.Errorf("dark damage during night should be ~1.20, got %f", mod)
	}
}

func TestTimeOfDaySpellDamageSystem_LightDamageDay(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDaySpellDamageSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)

	clock := &stubSpellDamageClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	lightingSys.SetClock(clock)
	lightingSys.Update(nil, 0.1)
	sys.SetLightingSystem(lightingSys)

	mod := sys.GetDamageModifier(magic.ElementLight)
	if mod < 1.15 || mod > 1.25 {
		t.Errorf("light damage during day should be ~1.20, got %f", mod)
	}
}

func TestTimeOfDaySpellDamageSystem_EarthDamageDawn(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDaySpellDamageSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)

	// Set 6:00 for dawn
	clock := &stubSpellDamageClock{now: time.Date(2026, 1, 1, 6, 0, 0, 0, time.UTC)}
	lightingSys.SetClock(clock)
	lightingSys.Update(nil, 0.1)
	sys.SetLightingSystem(lightingSys)

	mod := sys.GetDamageModifier(magic.ElementEarth)
	if mod < 1.05 || mod > 1.15 {
		t.Errorf("earth damage during dawn should be ~1.10, got %f", mod)
	}
}

func TestTimeOfDaySpellDamageSystem_IceDamageNight(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDaySpellDamageSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)

	clock := &stubSpellDamageClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lightingSys.SetClock(clock)
	lightingSys.Update(nil, 0.1)
	sys.SetLightingSystem(lightingSys)

	mod := sys.GetDamageModifier(magic.ElementIce)
	if mod < 1.05 || mod > 1.15 {
		t.Errorf("ice damage during night should be ~1.10, got %f", mod)
	}
}

func TestTimeOfDaySpellDamageSystem_GenreModifier_Horror(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDaySpellDamageSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)

	// Set night time
	clock := &stubSpellDamageClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lightingSys.SetClock(clock)
	lightingSys.Update(nil, 0.1)
	sys.SetLightingSystem(lightingSys)
	sys.SetGenre("horror")

	// Dark spells get extra 0.10 in horror
	mod := sys.GetDamageModifier(magic.ElementDark)
	if mod < 1.25 || mod > 1.35 {
		t.Errorf("dark damage in horror at night should be ~1.30, got %f", mod)
	}
}

func TestTimeOfDaySpellDamageSystem_Update_CacheInvalidation(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDaySpellDamageSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)

	// Start at noon
	clock := &stubSpellDamageClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	lightingSys.SetClock(clock)
	lightingSys.Update(nil, 0.1)
	sys.SetLightingSystem(lightingSys)

	// Get modifier (should cache)
	mod1 := sys.GetDamageModifier(magic.ElementFire)

	// Change to night
	clock.now = time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)
	lightingSys.Update(nil, 0.1)

	// Update with enough delta time to trigger check
	sys.Update(nil, 1.0)

	// Get modifier again (should be different)
	mod2 := sys.GetDamageModifier(magic.ElementFire)

	if mod1 == mod2 {
		t.Errorf("cache should have invalidated, but got same modifier: %f", mod1)
	}
}

func TestTimeOfDaySpellDamageSystem_ClearCache(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDaySpellDamageSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)

	clock := &stubSpellDamageClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	lightingSys.SetClock(clock)
	lightingSys.Update(nil, 0.1)
	sys.SetLightingSystem(lightingSys)

	// Populate cache
	sys.GetDamageModifier(magic.ElementFire)
	if len(sys.modCache) == 0 {
		t.Error("cache should have been populated")
	}

	// Clear cache
	sys.ClearCache()
	if len(sys.modCache) != 0 {
		t.Error("cache should be empty after ClearCache")
	}
}

func TestTimeOfDaySpellDamageSystem_GetCurrentTimeOfDay(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDaySpellDamageSystem(world, 12345)

	// Without lighting system
	_, ok := sys.GetCurrentTimeOfDay()
	if ok {
		t.Error("GetCurrentTimeOfDay should return false without lighting system")
	}

	// With lighting system at night
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)
	clock := &stubSpellDamageClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lightingSys.SetClock(clock)
	lightingSys.Update(nil, 0.1)
	sys.SetLightingSystem(lightingSys)

	tod, ok := sys.GetCurrentTimeOfDay()
	if !ok {
		t.Error("GetCurrentTimeOfDay should return true with lighting system")
	}
	if tod != palette.TimeOfDayNight {
		t.Errorf("expected Night, got %v", tod)
	}
}

func TestTimeOfDaySpellDamageSystem_GetBonusDescription(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDaySpellDamageSystem(world, 12345)

	// Without lighting system
	desc := sys.GetBonusDescription()
	if desc != "" {
		t.Error("GetBonusDescription should return empty string without lighting system")
	}

	// With lighting system at day (light spells strongest)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)
	clock := &stubSpellDamageClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	lightingSys.SetClock(clock)
	lightingSys.Update(nil, 0.1)
	sys.SetLightingSystem(lightingSys)

	desc = sys.GetBonusDescription()
	if desc == "" {
		t.Error("GetBonusDescription should return bonus info during day")
	}
}

func TestTimeOfDaySpellDamageSystem_AllElements(t *testing.T) {
	tests := []struct {
		name    string
		element magic.ElementType
		hour    int
		minMod  float64
		maxMod  float64
	}{
		{"fire_day", magic.ElementFire, 12, 1.10, 1.20},
		{"fire_night", magic.ElementFire, 22, 0.80, 0.90},
		{"light_day", magic.ElementLight, 12, 1.15, 1.25},
		{"light_night", magic.ElementLight, 22, 0.75, 0.85},
		{"dark_day", magic.ElementDark, 12, 0.75, 0.85},
		{"dark_night", magic.ElementDark, 22, 1.15, 1.25},
		{"arcane_day", magic.ElementArcane, 12, 0.90, 1.00},
		{"arcane_night", magic.ElementArcane, 22, 1.05, 1.15},
		{"earth_dawn", magic.ElementEarth, 6, 1.05, 1.15},
		{"earth_day", magic.ElementEarth, 12, 0.95, 1.05},
		{"wind_dusk", magic.ElementWind, 18, 1.05, 1.15},
		{"ice_night", magic.ElementIce, 22, 1.05, 1.15},
		{"lightning_day", magic.ElementLightning, 12, 1.00, 1.10},
		{"none_any", magic.ElementNone, 12, 0.95, 1.05},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewTimeOfDaySpellDamageSystem(world, 12345)
			lightingSys := NewTimeOfDayLightingSystem(world, 12345)

			clock := &stubSpellDamageClock{now: time.Date(2026, 1, 1, tt.hour, 0, 0, 0, time.UTC)}
			lightingSys.SetClock(clock)
			lightingSys.Update(nil, 0.1)
			sys.SetLightingSystem(lightingSys)

			mod := sys.GetDamageModifier(tt.element)
			if mod < tt.minMod || mod > tt.maxMod {
				t.Errorf("%s: expected modifier in [%f, %f], got %f", tt.name, tt.minMod, tt.maxMod, mod)
			}
		})
	}
}

func BenchmarkTimeOfDaySpellDamageSystem_GetDamageModifier(b *testing.B) {
	world := NewWorld()
	sys := NewTimeOfDaySpellDamageSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)

	clock := &stubSpellDamageClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	lightingSys.SetClock(clock)
	lightingSys.Update(nil, 0.1)
	sys.SetLightingSystem(lightingSys)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.GetDamageModifier(magic.ElementFire)
	}
}
