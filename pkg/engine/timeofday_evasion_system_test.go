//go:build ignore

package engine

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/rendering/palette"
)

// stubEvasionClock implements GameClock for testing.
type stubEvasionClock struct {
	now time.Time
}

func (c *stubEvasionClock) Now() time.Time {
	return c.now
}

func TestNewTimeOfDayEvasionSystem(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayEvasionSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewTimeOfDayEvasionSystem returned nil")
	}
	if sys.world != world {
		t.Error("world not set correctly")
	}
	if len(sys.baseModifiers) != 4 {
		t.Errorf("expected 4 base modifiers, got %d", len(sys.baseModifiers))
	}
	if len(sys.genreModifiers) != 5 {
		t.Errorf("expected 5 genre modifiers, got %d", len(sys.genreModifiers))
	}
}

func TestTimeOfDayEvasionSystem_SetGenre(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayEvasionSystem(world, 12345)

	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("expected genre 'horror', got '%s'", sys.genreID)
	}
}

func TestTimeOfDayEvasionSystem_SetLightingSystem(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayEvasionSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	sys.SetLightingSystem(lighting)
	if sys.lightingSystem != lighting {
		t.Error("lighting system not set correctly")
	}
}

func TestTimeOfDayEvasionSystem_UpdateWithoutLighting(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayEvasionSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(NewStatsComponent())

	// Should not panic without lighting system
	sys.Update([]*Entity{entity}, 1.0)
}

func TestTimeOfDayEvasionSystem_NightEvasionBonus(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayEvasionSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	// Set clock to night (22:00)
	clock := &stubEvasionClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lighting.SetClock(clock)

	sys.SetLightingSystem(lighting)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Evasion = 0.10
	entity.AddComponent(stats)

	// Force time check
	sys.timeSinceCheck = sys.updateInterval

	sys.Update([]*Entity{entity}, 0.1)

	// Fantasy night: base +0.05 + genre +0.03 = +0.08
	expectedEvasion := 0.10 + 0.08
	if stats.Evasion != expectedEvasion {
		t.Errorf("expected evasion %.2f, got %.2f", expectedEvasion, stats.Evasion)
	}
}

func TestTimeOfDayEvasionSystem_DayEvasionPenalty(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayEvasionSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	// Set clock to day (12:00)
	clock := &stubEvasionClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	lighting.SetClock(clock)

	sys.SetLightingSystem(lighting)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Evasion = 0.10
	entity.AddComponent(stats)

	sys.timeSinceCheck = sys.updateInterval
	sys.Update([]*Entity{entity}, 0.1)

	// Fantasy day: base -0.02, no genre modifier = -0.02
	expectedEvasion := 0.10 - 0.02
	if stats.Evasion != expectedEvasion {
		t.Errorf("expected evasion %.2f, got %.2f", expectedEvasion, stats.Evasion)
	}
}

func TestTimeOfDayEvasionSystem_ScifiNightPenalty(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayEvasionSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	// Set clock to night
	clock := &stubEvasionClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lighting.SetClock(clock)

	sys.SetLightingSystem(lighting)
	sys.SetGenre("scifi")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Evasion = 0.10
	entity.AddComponent(stats)

	sys.timeSinceCheck = sys.updateInterval
	sys.Update([]*Entity{entity}, 0.1)

	// Scifi night: base +0.05 + genre -0.03 = +0.02
	expectedEvasion := 0.10 + 0.02
	if stats.Evasion != expectedEvasion {
		t.Errorf("expected evasion %.2f, got %.2f", expectedEvasion, stats.Evasion)
	}
}

func TestTimeOfDayEvasionSystem_HorrorNightBonus(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayEvasionSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	// Set clock to night
	clock := &stubEvasionClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lighting.SetClock(clock)

	sys.SetLightingSystem(lighting)
	sys.SetGenre("horror")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Evasion = 0.10
	entity.AddComponent(stats)

	sys.timeSinceCheck = sys.updateInterval
	sys.Update([]*Entity{entity}, 0.1)

	// Horror night: base +0.05 + genre +0.04 = +0.09
	expectedEvasion := 0.10 + 0.09
	if stats.Evasion != expectedEvasion {
		t.Errorf("expected evasion %.2f, got %.2f", expectedEvasion, stats.Evasion)
	}
}

func TestTimeOfDayEvasionSystem_ClampToValidRange(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayEvasionSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	// Set clock to night for max bonus
	clock := &stubEvasionClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lighting.SetClock(clock)

	sys.SetLightingSystem(lighting)
	sys.SetGenre("horror")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Evasion = 0.95 // Near max
	entity.AddComponent(stats)

	sys.timeSinceCheck = sys.updateInterval
	sys.Update([]*Entity{entity}, 0.1)

	// Should be clamped to 1.0
	if stats.Evasion > 1.0 {
		t.Errorf("evasion should be clamped to 1.0, got %.2f", stats.Evasion)
	}
}

func TestTimeOfDayEvasionSystem_ClampToZero(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayEvasionSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	// Set clock to day for penalty
	clock := &stubEvasionClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	lighting.SetClock(clock)

	sys.SetLightingSystem(lighting)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Evasion = 0.01 // Very low
	entity.AddComponent(stats)

	sys.timeSinceCheck = sys.updateInterval
	sys.Update([]*Entity{entity}, 0.1)

	// Should be clamped to 0.0
	if stats.Evasion < 0.0 {
		t.Errorf("evasion should be clamped to 0.0, got %.2f", stats.Evasion)
	}
}

func TestTimeOfDayEvasionSystem_GetCurrentModifier(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayEvasionSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	// Set clock to night
	clock := &stubEvasionClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lighting.SetClock(clock)

	sys.SetLightingSystem(lighting)
	sys.SetGenre("fantasy")

	mod := sys.GetCurrentModifier()

	// Fantasy night: base +0.05 + genre +0.03 = +0.08
	if mod != 0.08 {
		t.Errorf("expected modifier 0.08, got %.2f", mod)
	}
}

func TestTimeOfDayEvasionSystem_GetCurrentModifierWithoutLighting(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayEvasionSystem(world, 12345)

	mod := sys.GetCurrentModifier()
	if mod != 0.0 {
		t.Errorf("expected modifier 0.0 without lighting, got %.2f", mod)
	}
}

func TestTimeOfDayEvasionSystem_GetOriginalEvasion(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayEvasionSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	clock := &stubEvasionClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lighting.SetClock(clock)

	sys.SetLightingSystem(lighting)

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Evasion = 0.15
	entity.AddComponent(stats)

	sys.timeSinceCheck = sys.updateInterval
	sys.Update([]*Entity{entity}, 0.1)

	orig, ok := sys.GetOriginalEvasion(entity.ID)
	if !ok {
		t.Error("expected to find original evasion")
	}
	if orig != 0.15 {
		t.Errorf("expected original evasion 0.15, got %.2f", orig)
	}
}

func TestTimeOfDayEvasionSystem_GetBonusDescription(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayEvasionSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	// Set clock to night for bonus
	clock := &stubEvasionClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lighting.SetClock(clock)

	sys.SetLightingSystem(lighting)
	sys.SetGenre("fantasy")

	desc := sys.GetBonusDescription()
	if desc == "" {
		t.Error("expected bonus description, got empty string")
	}
	if desc != "Night Agility: +8% Evasion" {
		t.Errorf("unexpected description: %s", desc)
	}
}

func TestTimeOfDayEvasionSystem_GetBonusDescriptionPenalty(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayEvasionSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	// Set clock to day for penalty
	clock := &stubEvasionClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	lighting.SetClock(clock)

	sys.SetLightingSystem(lighting)
	sys.SetGenre("fantasy")

	desc := sys.GetBonusDescription()
	if desc == "" {
		t.Error("expected penalty description, got empty string")
	}
	if desc != "Day Exposed: -2% Evasion" {
		t.Errorf("unexpected description: %s", desc)
	}
}

func TestTimeOfDayEvasionSystem_DuskBonus(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayEvasionSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	// Set clock to dusk (18:00)
	clock := &stubEvasionClock{now: time.Date(2026, 1, 1, 18, 0, 0, 0, time.UTC)}
	lighting.SetClock(clock)

	sys.SetLightingSystem(lighting)
	sys.SetGenre("cyberpunk")

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Evasion = 0.10
	entity.AddComponent(stats)

	sys.timeSinceCheck = sys.updateInterval
	sys.Update([]*Entity{entity}, 0.1)

	// Cyberpunk dusk: base +0.03 + genre +0.02 = +0.05
	expectedEvasion := 0.10 + 0.05
	if stats.Evasion != expectedEvasion {
		t.Errorf("expected evasion %.2f, got %.2f", expectedEvasion, stats.Evasion)
	}
}

func TestTimeOfDayEvasionSystem_UpdateIntervalThrottling(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayEvasionSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	clock := &stubEvasionClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lighting.SetClock(clock)

	sys.SetLightingSystem(lighting)

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Evasion = 0.10
	entity.AddComponent(stats)

	// First update with short delta time - should not process
	sys.Update([]*Entity{entity}, 0.1)

	// Original should still be empty (no processing yet)
	_, ok := sys.GetOriginalEvasion(entity.ID)
	if ok {
		t.Error("should not process until interval reached")
	}

	// Now accumulate enough time
	sys.Update([]*Entity{entity}, 1.0)

	// Now should have processed
	_, ok = sys.GetOriginalEvasion(entity.ID)
	if !ok {
		t.Error("should have processed after interval reached")
	}
}

func TestTimeOfDayEvasionSystem_AllGenres(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			world := NewWorld(nil)
			sys := NewTimeOfDayEvasionSystem(world, 12345)
			lighting := NewTimeOfDayLightingSystem(world, 12345)

			clock := &stubEvasionClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
			lighting.SetClock(clock)

			sys.SetLightingSystem(lighting)
			sys.SetGenre(genre)

			entity := world.CreateEntity()
			stats := NewStatsComponent()
			stats.Evasion = 0.10
			entity.AddComponent(stats)

			sys.timeSinceCheck = sys.updateInterval
			sys.Update([]*Entity{entity}, 0.1)

			// Just verify it doesn't crash and modifies evasion
			if stats.Evasion == 0.10 {
				t.Errorf("genre %s should modify evasion at night", genre)
			}
		})
	}
}

func TestEvasionItoa(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{5, "5"},
		{10, "10"},
		{123, "123"},
	}

	for _, tt := range tests {
		result := evasionItoa(tt.input)
		if result != tt.expected {
			t.Errorf("evasionItoa(%d) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}

func BenchmarkTimeOfDayEvasionSystem_Update(b *testing.B) {
	world := NewWorld(nil)
	sys := NewTimeOfDayEvasionSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	clock := &stubEvasionClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lighting.SetClock(clock)

	sys.SetLightingSystem(lighting)

	// Create 100 entities
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		stats := NewStatsComponent()
		stats.Evasion = 0.10
		entity.AddComponent(stats)
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceCheck = sys.updateInterval
		sys.Update(entities, 0.016)
	}
}
