package engine

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/rendering/palette"
)

// stubAttackSpeedClock implements GameClock for testing.
type stubAttackSpeedClock struct {
	now time.Time
}

func (c *stubAttackSpeedClock) Now() time.Time {
	return c.now
}

func TestNewTimeOfDayAttackSpeedSystem(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayAttackSpeedSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewTimeOfDayAttackSpeedSystem returned nil")
	}

	if sys.world != world {
		t.Error("world not set correctly")
	}

	if sys.genreID != "fantasy" {
		t.Errorf("default genre = %s, want fantasy", sys.genreID)
	}

	if sys.updateInterval != 1.0 {
		t.Errorf("updateInterval = %v, want 1.0", sys.updateInterval)
	}
}

func TestTimeOfDayAttackSpeedSystem_SetGenre(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayAttackSpeedSystem(world, 12345)

	sys.SetGenre("scifi")
	if sys.genreID != "scifi" {
		t.Errorf("genre = %s, want scifi", sys.genreID)
	}
}

func TestTimeOfDayAttackSpeedSystem_SetLightingSystem(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayAttackSpeedSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	sys.SetLightingSystem(lighting)
	if sys.lightingSystem != lighting {
		t.Error("lighting system not set correctly")
	}
}

func TestTimeOfDayAttackSpeedSystem_NightBonus(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayAttackSpeedSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	// Set clock to night (22:00)
	clock := &stubAttackSpeedClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lighting.SetClock(clock)
	sys.SetLightingSystem(lighting)

	// Force time update
	lighting.Update([]*Entity{}, 0.1)

	// Create entity with attack component
	entity := NewEntity()
	attack := &AttackComponent{
		Damage:        10.0,
		Range:         50.0,
		Cooldown:      1.0, // 1 second base cooldown
		CooldownTimer: 0.0,
	}
	entity.AddComponent(attack)
	world.AddEntity(entity)

	entities := []*Entity{entity}

	// Run update (need to exceed update interval)
	sys.Update(entities, 1.1)

	// Night should reduce cooldown (multiplier 0.90 for fantasy)
	// Fantasy genre has 0.88 night multiplier
	expectedCooldown := 0.88 // 1.0 * 0.88
	if attack.Cooldown != expectedCooldown {
		t.Errorf("cooldown = %v, want %v (night bonus)", attack.Cooldown, expectedCooldown)
	}
}

func TestTimeOfDayAttackSpeedSystem_DayPenalty(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayAttackSpeedSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	// Set clock to noon (12:00)
	clock := &stubAttackSpeedClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	lighting.SetClock(clock)
	sys.SetLightingSystem(lighting)

	// Force time update
	lighting.Update([]*Entity{}, 0.1)

	// Create entity with attack component
	entity := NewEntity()
	attack := &AttackComponent{
		Damage:        10.0,
		Range:         50.0,
		Cooldown:      1.0,
		CooldownTimer: 0.0,
	}
	entity.AddComponent(attack)
	world.AddEntity(entity)

	entities := []*Entity{entity}

	// Run update
	sys.Update(entities, 1.1)

	// Day should increase cooldown slightly
	expectedCooldown := 1.03 // 1.0 * 1.03
	if attack.Cooldown != expectedCooldown {
		t.Errorf("cooldown = %v, want %v (day penalty)", attack.Cooldown, expectedCooldown)
	}
}

func TestTimeOfDayAttackSpeedSystem_NoLightingSystem(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayAttackSpeedSystem(world, 12345)

	// Create entity with attack component
	entity := NewEntity()
	attack := &AttackComponent{Cooldown: 1.0}
	entity.AddComponent(attack)

	entities := []*Entity{entity}

	// Should not panic and not modify cooldown
	sys.Update(entities, 1.1)

	if attack.Cooldown != 1.0 {
		t.Errorf("cooldown modified without lighting system: %v", attack.Cooldown)
	}
}

func TestTimeOfDayAttackSpeedSystem_NoAttackComponent(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayAttackSpeedSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)
	clock := &stubAttackSpeedClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lighting.SetClock(clock)
	sys.SetLightingSystem(lighting)

	// Create entity without attack component
	entity := NewEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.AddEntity(entity)

	entities := []*Entity{entity}

	// Should not panic
	sys.Update(entities, 1.1)
}

func TestTimeOfDayAttackSpeedSystem_GenreModifiers(t *testing.T) {
	tests := []struct {
		name        string
		genre       string
		timeOfDay   palette.TimeOfDay
		hour        int
		wantMult    float64
		description string
	}{
		{
			name:        "fantasy night",
			genre:       "fantasy",
			timeOfDay:   palette.TimeOfDayNight,
			hour:        22,
			wantMult:    0.88,
			description: "fantasy rogues excel at night",
		},
		{
			name:        "scifi night",
			genre:       "scifi",
			timeOfDay:   palette.TimeOfDayNight,
			hour:        22,
			wantMult:    0.95,
			description: "sensors partially compensate",
		},
		{
			name:        "horror day",
			genre:       "horror",
			timeOfDay:   palette.TimeOfDayDay,
			hour:        12,
			wantMult:    1.05,
			description: "enemies more reactive in light",
		},
		{
			name:        "cyberpunk night",
			genre:       "cyberpunk",
			timeOfDay:   palette.TimeOfDayNight,
			hour:        22,
			wantMult:    0.92,
			description: "urban shadows favor attackers",
		},
		{
			name:        "postapoc night",
			genre:       "postapoc",
			timeOfDay:   palette.TimeOfDayNight,
			hour:        22,
			wantMult:    0.89,
			description: "predator instincts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld(nil)
			sys := NewTimeOfDayAttackSpeedSystem(world, 12345)
			lighting := NewTimeOfDayLightingSystem(world, 12345)

			clock := &stubAttackSpeedClock{now: time.Date(2026, 1, 1, tt.hour, 0, 0, 0, time.UTC)}
			lighting.SetClock(clock)
			sys.SetLightingSystem(lighting)
			sys.SetGenre(tt.genre)

			// Force time update
			lighting.Update([]*Entity{}, 0.1)

			mult := sys.getCooldownMultiplier(tt.timeOfDay)
			if mult != tt.wantMult {
				t.Errorf("getCooldownMultiplier(%s) = %v, want %v (%s)",
					tt.timeOfDay, mult, tt.wantMult, tt.description)
			}
		})
	}
}

func TestTimeOfDayAttackSpeedSystem_GetCurrentMultiplier(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayAttackSpeedSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	// Without lighting system, should return 1.0
	mult := sys.GetCurrentMultiplier()
	if mult != 1.0 {
		t.Errorf("GetCurrentMultiplier without lighting = %v, want 1.0", mult)
	}

	// Set clock to night
	clock := &stubAttackSpeedClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lighting.SetClock(clock)
	sys.SetLightingSystem(lighting)
	lighting.Update([]*Entity{}, 0.1)

	mult = sys.GetCurrentMultiplier()
	if mult >= 1.0 {
		t.Errorf("GetCurrentMultiplier at night = %v, want < 1.0", mult)
	}
}

func TestTimeOfDayAttackSpeedSystem_GetOriginalCooldown(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayAttackSpeedSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	clock := &stubAttackSpeedClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lighting.SetClock(clock)
	sys.SetLightingSystem(lighting)
	lighting.Update([]*Entity{}, 0.1)

	entity := NewEntity()
	attack := &AttackComponent{Cooldown: 2.0}
	entity.AddComponent(attack)
	world.AddEntity(entity)

	entities := []*Entity{entity}
	sys.Update(entities, 1.1)

	original, ok := sys.GetOriginalCooldown(entity.ID)
	if !ok {
		t.Fatal("GetOriginalCooldown returned false")
	}
	if original != 2.0 {
		t.Errorf("original cooldown = %v, want 2.0", original)
	}
}

func TestTimeOfDayAttackSpeedSystem_GetBonusDescription(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayAttackSpeedSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	// Without lighting, should return empty
	desc := sys.GetBonusDescription()
	if desc != "" {
		t.Errorf("GetBonusDescription without lighting = %q, want empty", desc)
	}

	// At night, should show bonus
	clock := &stubAttackSpeedClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lighting.SetClock(clock)
	sys.SetLightingSystem(lighting)
	lighting.Update([]*Entity{}, 0.1)

	desc = sys.GetBonusDescription()
	if desc == "" {
		t.Error("GetBonusDescription at night should show bonus")
	}
}

func TestTimeOfDayAttackSpeedSystem_CooldownClamping(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayAttackSpeedSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	clock := &stubAttackSpeedClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lighting.SetClock(clock)
	sys.SetLightingSystem(lighting)
	lighting.Update([]*Entity{}, 0.1)

	// Test very small cooldown doesn't go below 0.1
	entity := NewEntity()
	attack := &AttackComponent{Cooldown: 0.05} // Very fast attack
	entity.AddComponent(attack)
	world.AddEntity(entity)

	entities := []*Entity{entity}
	sys.Update(entities, 1.1)

	if attack.Cooldown < 0.1 {
		t.Errorf("cooldown = %v, should be clamped to at least 0.1", attack.Cooldown)
	}
}

func TestTimeOfDayAttackSpeedSystem_UpdateIntervalRespected(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayAttackSpeedSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	clock := &stubAttackSpeedClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lighting.SetClock(clock)
	sys.SetLightingSystem(lighting)
	lighting.Update([]*Entity{}, 0.1)

	entity := NewEntity()
	attack := &AttackComponent{Cooldown: 1.0}
	entity.AddComponent(attack)
	world.AddEntity(entity)

	entities := []*Entity{entity}

	// Update with less than interval - should not modify
	sys.Update(entities, 0.5)
	if attack.Cooldown != 1.0 {
		t.Errorf("cooldown modified before interval elapsed: %v", attack.Cooldown)
	}

	// Update again to exceed interval
	sys.Update(entities, 0.6)
	if attack.Cooldown == 1.0 {
		t.Error("cooldown not modified after interval elapsed")
	}
}

func TestAttackSpeedItoa(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{5, "5"},
		{10, "10"},
		{12, "12"},
		{100, "100"},
		{-5, "-5"},
	}

	for _, tt := range tests {
		got := attackSpeedItoa(tt.input)
		if got != tt.want {
			t.Errorf("attackSpeedItoa(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// Benchmark for hot path performance
func BenchmarkTimeOfDayAttackSpeedSystem_Update(b *testing.B) {
	world := NewWorld(nil)
	sys := NewTimeOfDayAttackSpeedSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	clock := &stubAttackSpeedClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lighting.SetClock(clock)
	sys.SetLightingSystem(lighting)
	lighting.Update([]*Entity{}, 0.1)

	// Create 100 entities with attack components
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		e := NewEntity()
		e.AddComponent(&AttackComponent{Cooldown: 1.0})
		world.AddEntity(e)
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceCheck = 0 // Force update
		sys.Update(entities, 1.1)
	}
}
