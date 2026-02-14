//go:build ignore

package engine

import (
	"math/rand"
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/sirupsen/logrus"
)

func TestNewTimeOfDayBlockChanceSystem(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	world.SetLogger(logger)

	sys := NewTimeOfDayBlockChanceSystem(world, 12345)
	if sys == nil {
		t.Fatal("NewTimeOfDayBlockChanceSystem returned nil")
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

func TestTimeOfDayBlockChanceSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayBlockChanceSystem(world, 12345)

	tests := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range tests {
		sys.SetGenre(genre)
		if sys.genreID != genre {
			t.Errorf("SetGenre(%s): genreID = %s", genre, sys.genreID)
		}
	}
}

func TestTimeOfDayBlockChanceSystem_SetLightingSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayBlockChanceSystem(world, 12345)

	lightingSys := NewTimeOfDayLightingSystem(world, 12345)
	sys.SetLightingSystem(lightingSys)

	if sys.lightingSystem != lightingSys {
		t.Error("lighting system not set")
	}
}

func TestTimeOfDayBlockChanceSystem_getBlockModifier(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayBlockChanceSystem(world, 12345)
	sys.SetGenre("fantasy")

	tests := []struct {
		name      string
		timeOfDay palette.TimeOfDay
		wantMin   float64
		wantMax   float64
	}{
		{"dawn", palette.TimeOfDayDawn, 0.02, 0.06},
		{"day", palette.TimeOfDayDay, 0.06, 0.10}, // Day is best for blocking
		{"dusk", palette.TimeOfDayDusk, 0.00, 0.03},
		{"night", palette.TimeOfDayNight, -0.05, 0.00},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod := sys.getBlockModifier(tt.timeOfDay)
			if mod < tt.wantMin || mod > tt.wantMax {
				t.Errorf("getBlockModifier(%v) = %f, want [%f, %f]", tt.timeOfDay, mod, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestTimeOfDayBlockChanceSystem_GenreModifiers(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayBlockChanceSystem(world, 12345)

	tests := []struct {
		name      string
		genre     string
		timeOfDay palette.TimeOfDay
		wantBase  float64
	}{
		{"scifi night bonus", "scifi", palette.TimeOfDayNight, 0.01},          // base -0.03 + genre +0.04 = 0.01
		{"horror night penalty", "horror", palette.TimeOfDayNight, -0.06},     // base -0.03 + genre -0.03 = -0.06
		{"cyberpunk night bonus", "cyberpunk", palette.TimeOfDayNight, -0.01}, // base -0.03 + genre +0.02 = -0.01
		{"postapoc dawn bonus", "postapoc", palette.TimeOfDayDawn, 0.05},      // base +0.03 + genre +0.02 = 0.05
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys.SetGenre(tt.genre)
			mod := sys.getBlockModifier(tt.timeOfDay)
			if mod < tt.wantBase-0.02 || mod > tt.wantBase+0.02 {
				t.Errorf("%s: getBlockModifier = %f, want ~%f", tt.name, mod, tt.wantBase)
			}
		})
	}
}

// stubBlockClock implements GameClock for testing
type stubBlockClock struct {
	now time.Time
}

func (c *stubBlockClock) Now() time.Time {
	return c.now
}

func TestTimeOfDayBlockChanceSystem_Update(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	world.SetLogger(logger)

	sys := NewTimeOfDayBlockChanceSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)

	// Create a clock at day time (12:00 noon)
	clock := &stubBlockClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	lightingSys.SetClock(clock)
	sys.SetLightingSystem(lightingSys)
	sys.SetGenre("fantasy")

	// Create entity with stats
	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.BlockChance = 0.10 // 10% base block
	entity.AddComponent(stats)

	// Update lighting system first to set time state
	lightingSys.Update([]*Entity{entity}, 0.016)

	// Update block system with enough deltaTime to trigger check
	sys.Update([]*Entity{entity}, 2.0)

	// At day with fantasy genre, should have positive modifier (+0.05 base + 0.03 fantasy)
	expectedMin := 0.10 + 0.05 // base + day bonus
	if stats.BlockChance < expectedMin-0.02 {
		t.Errorf("BlockChance = %f, want >= %f (day bonus)", stats.BlockChance, expectedMin-0.02)
	}
}

func TestTimeOfDayBlockChanceSystem_UpdateNoLightingSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayBlockChanceSystem(world, 12345)
	// Don't set lighting system

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.BlockChance = 0.10
	entity.AddComponent(stats)

	// Should not panic and block should remain unchanged
	sys.Update([]*Entity{entity}, 2.0)

	if stats.BlockChance != 0.10 {
		t.Errorf("BlockChance changed without lighting system: %f", stats.BlockChance)
	}
}

func TestTimeOfDayBlockChanceSystem_GetCurrentModifier(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayBlockChanceSystem(world, 12345)

	// Without lighting system, should return 0
	if mod := sys.GetCurrentModifier(); mod != 0.0 {
		t.Errorf("GetCurrentModifier without lighting = %f, want 0.0", mod)
	}

	// With lighting system at day time
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)
	clock := &stubBlockClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)} // Noon
	lightingSys.SetClock(clock)
	lightingSys.Update([]*Entity{}, 0.016)
	sys.SetLightingSystem(lightingSys)
	sys.SetGenre("fantasy")

	// Day time should have positive modifier
	mod := sys.GetCurrentModifier()
	if mod < 0.03 || mod > 0.12 {
		t.Errorf("GetCurrentModifier at noon = %f, want [0.03, 0.12]", mod)
	}
}

func TestTimeOfDayBlockChanceSystem_GetOriginalBlockChance(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayBlockChanceSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)
	clock := &stubBlockClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	lightingSys.SetClock(clock)
	lightingSys.Update([]*Entity{}, 0.016)
	sys.SetLightingSystem(lightingSys)

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.BlockChance = 0.15
	entity.AddComponent(stats)

	// Before update, no original stored
	_, ok := sys.GetOriginalBlockChance(entity.ID)
	if ok {
		t.Error("GetOriginalBlockChance returned true before update")
	}

	// After update, original should be stored
	sys.Update([]*Entity{entity}, 2.0)
	orig, ok := sys.GetOriginalBlockChance(entity.ID)
	if !ok {
		t.Error("GetOriginalBlockChance returned false after update")
	}
	if orig != 0.15 {
		t.Errorf("original block = %f, want 0.15", orig)
	}
}

func TestTimeOfDayBlockChanceSystem_GetBonusDescription(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayBlockChanceSystem(world, 12345)

	// Without lighting system
	if desc := sys.GetBonusDescription(); desc != "" {
		t.Errorf("GetBonusDescription without lighting = %q, want empty", desc)
	}

	// With lighting at day time
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)
	clock := &stubBlockClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	lightingSys.SetClock(clock)
	lightingSys.Update([]*Entity{}, 0.016)
	sys.SetLightingSystem(lightingSys)
	sys.SetGenre("fantasy")

	desc := sys.GetBonusDescription()
	if desc == "" {
		t.Error("GetBonusDescription at day returned empty string")
	}
}

func TestTimeOfDayBlockChanceSystem_BlockClamp(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayBlockChanceSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)
	clock := &stubBlockClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)} // Noon for max bonus
	lightingSys.SetClock(clock)
	lightingSys.Update([]*Entity{}, 0.016)
	sys.SetLightingSystem(lightingSys)
	sys.SetGenre("fantasy") // Day bonus

	// Entity with very high base block
	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.BlockChance = 0.98 // Very high base
	entity.AddComponent(stats)

	sys.Update([]*Entity{entity}, 2.0)

	// Should be clamped to 1.0
	if stats.BlockChance > 1.0 {
		t.Errorf("BlockChance = %f, should be clamped to 1.0", stats.BlockChance)
	}
}

func TestTimeOfDayBlockChanceSystem_NegativeClamp(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayBlockChanceSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)
	clock := &stubBlockClock{now: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)} // Night for penalty
	lightingSys.SetClock(clock)
	lightingSys.Update([]*Entity{}, 0.016)
	sys.SetLightingSystem(lightingSys)
	sys.SetGenre("horror") // Max night penalty

	// Entity with very low base block
	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.BlockChance = 0.01 // Very low base
	entity.AddComponent(stats)

	sys.Update([]*Entity{entity}, 2.0)

	// Should not go below 0
	if stats.BlockChance < 0.0 {
		t.Errorf("BlockChance = %f, should not be negative", stats.BlockChance)
	}
}

func TestBlockItoa(t *testing.T) {
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
		got := blockItoa(tt.input)
		if got != tt.want {
			t.Errorf("blockItoa(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func BenchmarkTimeOfDayBlockChanceSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewTimeOfDayBlockChanceSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 12345)
	clock := &stubBlockClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	lightingSys.SetClock(clock)
	lightingSys.Update([]*Entity{}, 0.016)
	sys.SetLightingSystem(lightingSys)

	// Create 100 entities with stats
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		stats := NewStatsComponent()
		stats.BlockChance = 0.05 + float64(i)*0.001
		entity.AddComponent(stats)
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceCheck = 0 // Reset to force update
		sys.Update(entities, 2.0)
	}
}

func BenchmarkTimeOfDayBlockChanceSystem_getBlockModifier(b *testing.B) {
	world := NewWorld()
	sys := NewTimeOfDayBlockChanceSystem(world, 12345)
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
		_ = sys.getBlockModifier(timeOfDay)
	}
}
