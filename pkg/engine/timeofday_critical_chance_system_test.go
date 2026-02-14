//go:build ignore

package engine

import (
	"math/rand"
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/sirupsen/logrus"
)

func TestNewTimeOfDayCriticalChanceSystem(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	world.SetLogger(logger)

	sys := NewTimeOfDayCriticalChanceSystem(world, 12345)
	if sys == nil {
		t.Fatal("NewTimeOfDayCriticalChanceSystem returned nil")
	}

	if sys.world != world {
		t.Error("world reference not set")
	}

	if sys.genreID != "fantasy" {
		t.Errorf("default genre = %s, want fantasy", sys.genreID)
	}

	if sys.updateInterval != 1.0 {
		t.Errorf("updateInterval = %f, want 1.0", sys.updateInterval)
	}
}

func TestTimeOfDayCriticalChanceSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayCriticalChanceSystem(world, 12345)

	tests := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range tests {
		sys.SetGenre(genre)
		if sys.genreID != genre {
			t.Errorf("SetGenre(%s): genreID = %s", genre, sys.genreID)
		}
	}
}

func TestTimeOfDayCriticalChanceSystem_SetLightingSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayCriticalChanceSystem(world, 12345)

	lightingSys := NewTimeOfDayLightingSystem(world, 12345)
	sys.SetLightingSystem(lightingSys)

	if sys.lightingSystem != lightingSys {
		t.Error("lighting system not set")
	}
}

func TestTimeOfDayCriticalChanceSystem_getCritModifier(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayCriticalChanceSystem(world, 12345)
	sys.SetGenre("fantasy")

	tests := []struct {
		name      string
		timeOfDay palette.TimeOfDay
		wantMin   float64
		wantMax   float64
	}{
		{"dawn", palette.TimeOfDayDawn, 0.01, 0.05},
		{"day", palette.TimeOfDayDay, -0.01, 0.01},
		{"dusk", palette.TimeOfDayDusk, 0.02, 0.06},
		{"night", palette.TimeOfDayNight, 0.05, 0.10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod := sys.getCritModifier(tt.timeOfDay)
			if mod < tt.wantMin || mod > tt.wantMax {
				t.Errorf("getCritModifier(%v) = %f, want [%f, %f]", tt.timeOfDay, mod, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestTimeOfDayCriticalChanceSystem_GenreModifiers(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayCriticalChanceSystem(world, 12345)

	tests := []struct {
		name      string
		genre     string
		timeOfDay palette.TimeOfDay
		wantBase  float64
	}{
		{"scifi day bonus", "scifi", palette.TimeOfDayDay, 0.03},
		{"scifi night penalty", "scifi", palette.TimeOfDayNight, 0.03}, // 0.05 - 0.02 = 0.03
		{"horror night bonus", "horror", palette.TimeOfDayNight, 0.10},
		{"cyberpunk dusk bonus", "cyberpunk", palette.TimeOfDayDusk, 0.06},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys.SetGenre(tt.genre)
			mod := sys.getCritModifier(tt.timeOfDay)
			if mod < tt.wantBase-0.02 || mod > tt.wantBase+0.02 {
				t.Errorf("%s: getCritModifier = %f, want ~%f", tt.name, mod, tt.wantBase)
			}
		})
	}
}

// stubClock implements GameClock for testing
type stubCritClock struct {
	now time.Time
}

func (c *stubCritClock) Now() time.Time {
	return c.now
}

func TestTimeOfDayCriticalChanceSystem_Update(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	world.SetLogger(logger)

	sys := NewTimeOfDayCriticalChanceSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)

	// Create a clock at night time (22:00)
	clock := &stubCritClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lightingSys.SetClock(clock)
	sys.SetLightingSystem(lightingSys)
	sys.SetGenre("fantasy")

	// Create entity with stats
	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.CritChance = 0.10 // 10% base crit
	entity.AddComponent(stats)

	// Update lighting system first to set time state
	lightingSys.Update([]*Entity{entity}, 0.016)

	// Update crit system with enough deltaTime to trigger check
	sys.Update([]*Entity{entity}, 2.0)

	// At night with fantasy genre, should have positive modifier
	expectedMin := 0.10 + 0.05 // base + night bonus
	if stats.CritChance < expectedMin-0.02 {
		t.Errorf("CritChance = %f, want >= %f (night bonus)", stats.CritChance, expectedMin-0.02)
	}
}

func TestTimeOfDayCriticalChanceSystem_UpdateNoLightingSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayCriticalChanceSystem(world, 12345)
	// Don't set lighting system

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.CritChance = 0.10
	entity.AddComponent(stats)

	// Should not panic and crit should remain unchanged
	sys.Update([]*Entity{entity}, 2.0)

	if stats.CritChance != 0.10 {
		t.Errorf("CritChance changed without lighting system: %f", stats.CritChance)
	}
}

func TestTimeOfDayCriticalChanceSystem_GetCurrentModifier(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayCriticalChanceSystem(world, 12345)

	// Without lighting system, should return 0
	if mod := sys.GetCurrentModifier(); mod != 0.0 {
		t.Errorf("GetCurrentModifier without lighting = %f, want 0.0", mod)
	}

	// With lighting system
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)
	clock := &stubCritClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)} // Noon
	lightingSys.SetClock(clock)
	lightingSys.Update([]*Entity{}, 0.016)
	sys.SetLightingSystem(lightingSys)

	// Day time should have ~0.0 modifier
	mod := sys.GetCurrentModifier()
	if mod < -0.05 || mod > 0.05 {
		t.Errorf("GetCurrentModifier at noon = %f, want ~0.0", mod)
	}
}

func TestTimeOfDayCriticalChanceSystem_GetOriginalCritChance(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayCriticalChanceSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)
	clock := &stubCritClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lightingSys.SetClock(clock)
	lightingSys.Update([]*Entity{}, 0.016)
	sys.SetLightingSystem(lightingSys)

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.CritChance = 0.15
	entity.AddComponent(stats)

	// Before update, no original stored
	_, ok := sys.GetOriginalCritChance(entity.ID)
	if ok {
		t.Error("GetOriginalCritChance returned true before update")
	}

	// After update, original should be stored
	sys.Update([]*Entity{entity}, 2.0)
	orig, ok := sys.GetOriginalCritChance(entity.ID)
	if !ok {
		t.Error("GetOriginalCritChance returned false after update")
	}
	if orig != 0.15 {
		t.Errorf("original crit = %f, want 0.15", orig)
	}
}

func TestTimeOfDayCriticalChanceSystem_GetBonusDescription(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayCriticalChanceSystem(world, 12345)

	// Without lighting system
	if desc := sys.GetBonusDescription(); desc != "" {
		t.Errorf("GetBonusDescription without lighting = %q, want empty", desc)
	}

	// With lighting at night
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)
	clock := &stubCritClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lightingSys.SetClock(clock)
	lightingSys.Update([]*Entity{}, 0.016)
	sys.SetLightingSystem(lightingSys)
	sys.SetGenre("fantasy")

	desc := sys.GetBonusDescription()
	if desc == "" {
		t.Error("GetBonusDescription at night returned empty string")
	}
}

func TestTimeOfDayCriticalChanceSystem_CritClamp(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayCriticalChanceSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)
	clock := &stubCritClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lightingSys.SetClock(clock)
	lightingSys.Update([]*Entity{}, 0.016)
	sys.SetLightingSystem(lightingSys)
	sys.SetGenre("horror") // Max night bonus

	// Entity with very high base crit
	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.CritChance = 0.98 // Very high base
	entity.AddComponent(stats)

	sys.Update([]*Entity{entity}, 2.0)

	// Should be clamped to 1.0
	if stats.CritChance > 1.0 {
		t.Errorf("CritChance = %f, should be clamped to 1.0", stats.CritChance)
	}
}

func TestTimeOfDayCriticalChanceSystem_NegativeClamp(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayCriticalChanceSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)
	clock := &stubCritClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lightingSys.SetClock(clock)
	lightingSys.Update([]*Entity{}, 0.016)
	sys.SetLightingSystem(lightingSys)
	sys.SetGenre("scifi") // Night penalty

	// Entity with very low base crit
	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.CritChance = 0.01 // Very low base
	entity.AddComponent(stats)

	sys.Update([]*Entity{entity}, 2.0)

	// Should not go below 0
	if stats.CritChance < 0.0 {
		t.Errorf("CritChance = %f, should not be negative", stats.CritChance)
	}
}

func TestCritItoa(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{5, "5"},
		{10, "10"},
		{123, "123"},
		{-5, "-5"},
	}

	for _, tt := range tests {
		got := critItoa(tt.input)
		if got != tt.want {
			t.Errorf("critItoa(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func BenchmarkTimeOfDayCriticalChanceSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewTimeOfDayCriticalChanceSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)
	clock := &stubCritClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)}
	lightingSys.SetClock(clock)
	lightingSys.Update([]*Entity{}, 0.016)
	sys.SetLightingSystem(lightingSys)

	// Create 100 entities with stats
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		stats := NewStatsComponent()
		stats.CritChance = 0.05 + float64(i)*0.001
		entity.AddComponent(stats)
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceCheck = 0 // Reset to force update
		sys.Update(entities, 2.0)
	}
}

func BenchmarkTimeOfDayCriticalChanceSystem_getCritModifier(b *testing.B) {
	world := NewWorld()
	sys := NewTimeOfDayCriticalChanceSystem(world, 12345)
	sys.SetGenre("fantasy")

	rng := rand.New(rand.NewSource(12345))
	times := []palette.TimeOfDay{
		palette.TimeOfDayDawn,
		palette.TimeOfDayDay,
		palette.TimeOfDayDusk,
		palette.TimeOfDayNight,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		timeOfDay := times[rng.Intn(4)]
		_ = sys.getCritModifier(timeOfDay)
	}
}
