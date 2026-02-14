package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/palette"
)

func TestNewTimeOfDayXPBonusSystem(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayXPBonusSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewTimeOfDayXPBonusSystem returned nil")
	}
	if sys.world != world {
		t.Error("world reference not set")
	}
	if sys.rng == nil {
		t.Error("RNG not initialized")
	}
	if len(sys.baseMultipliers) != 4 {
		t.Errorf("expected 4 base multipliers, got %d", len(sys.baseMultipliers))
	}
}

func TestTimeOfDayXPBonusSystem_SetGenre(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayXPBonusSystem(world, 12345)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		sys.SetGenre(genre)
		if sys.genreID != genre {
			t.Errorf("SetGenre(%s) failed, got %s", genre, sys.genreID)
		}
	}
}

func TestTimeOfDayXPBonusSystem_SetLightingSystem(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayXPBonusSystem(world, 12345)
	lightingSystem := NewTimeOfDayLightingSystem(world, 12345)

	sys.SetLightingSystem(lightingSystem)

	if sys.lightingSystem != lightingSystem {
		t.Error("SetLightingSystem failed")
	}
}

func TestTimeOfDayXPBonusSystem_SetProgressionSystem(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayXPBonusSystem(world, 12345)
	progressionSystem := NewProgressionSystem(world)

	sys.SetProgressionSystem(progressionSystem)

	if sys.progressionSystem != progressionSystem {
		t.Error("SetProgressionSystem failed")
	}
}

func TestTimeOfDayXPBonusSystem_GetXPMultiplier(t *testing.T) {
	tests := []struct {
		name      string
		genre     string
		timeOfDay palette.TimeOfDay
		wantMin   float64
		wantMax   float64
	}{
		{"fantasy_day", "fantasy", palette.TimeOfDayDay, 0.95, 1.05},
		{"fantasy_night", "fantasy", palette.TimeOfDayNight, 1.15, 1.25},
		{"horror_day", "horror", palette.TimeOfDayDay, 0.90, 1.00},
		{"horror_night", "horror", palette.TimeOfDayNight, 1.25, 1.35},
		{"scifi_day", "scifi", palette.TimeOfDayDay, 1.05, 1.15},
		{"scifi_night", "scifi", palette.TimeOfDayNight, 1.05, 1.15},
		{"cyberpunk_dusk", "cyberpunk", palette.TimeOfDayDusk, 1.10, 1.20},
		{"cyberpunk_night", "cyberpunk", palette.TimeOfDayNight, 1.20, 1.30},
		{"postapoc_dawn", "postapoc", palette.TimeOfDayDawn, 1.05, 1.15},
		{"postapoc_night", "postapoc", palette.TimeOfDayNight, 1.20, 1.30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld(nil)
			sys := NewTimeOfDayXPBonusSystem(world, 12345)
			sys.SetGenre(tt.genre)

			mult := sys.getXPMultiplier(tt.timeOfDay)

			if mult < tt.wantMin || mult > tt.wantMax {
				t.Errorf("getXPMultiplier() = %v, want between %v and %v",
					mult, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestTimeOfDayXPBonusSystem_Update(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayXPBonusSystem(world, 12345)
	lightingSystem := NewTimeOfDayLightingSystem(world, 12345)
	sys.SetLightingSystem(lightingSystem)

	// Create player entity with input component
	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	player.AddComponent(&ExperienceComponent{Level: 1, CurrentXP: 0, RequiredXP: 100})

	// Create NPC without input (should not get multiplier)
	npc := world.CreateEntity()
	npc.AddComponent(&ExperienceComponent{Level: 1, CurrentXP: 0, RequiredXP: 100})

	entities := []*Entity{player, npc}

	// Update with enough delta to trigger check
	sys.Update(entities, 2.0)

	// Player should have multiplier set
	playerMult := sys.GetActiveMultiplier(player.ID)
	if playerMult < 0.5 || playerMult > 2.0 {
		t.Errorf("player multiplier out of range: %v", playerMult)
	}

	// NPC should not have multiplier (returns default 1.0)
	npcMult := sys.GetActiveMultiplier(npc.ID)
	if npcMult != 1.0 {
		t.Errorf("NPC should have default multiplier 1.0, got %v", npcMult)
	}
}

func TestTimeOfDayXPBonusSystem_UpdateWithoutLightingSystem(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayXPBonusSystem(world, 12345)
	// Don't set lighting system

	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	entities := []*Entity{player}

	// Should not panic
	sys.Update(entities, 2.0)

	// Should return default multiplier
	mult := sys.GetActiveMultiplier(player.ID)
	if mult != 1.0 {
		t.Errorf("expected default multiplier 1.0, got %v", mult)
	}
}

func TestTimeOfDayXPBonusSystem_GetCurrentMultiplier(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayXPBonusSystem(world, 12345)

	// Without lighting system, should return 1.0
	mult := sys.GetCurrentMultiplier()
	if mult != 1.0 {
		t.Errorf("expected 1.0 without lighting system, got %v", mult)
	}

	// With lighting system, should return calculated multiplier
	lightingSystem := NewTimeOfDayLightingSystem(world, 12345)
	sys.SetLightingSystem(lightingSystem)

	mult = sys.GetCurrentMultiplier()
	if mult < 0.5 || mult > 2.0 {
		t.Errorf("multiplier out of valid range: %v", mult)
	}
}

func TestTimeOfDayXPBonusSystem_GetBonusDescription(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayXPBonusSystem(world, 12345)
	lightingSystem := NewTimeOfDayLightingSystem(world, 12345)
	sys.SetLightingSystem(lightingSystem)

	// Without lighting system
	sys.lightingSystem = nil
	desc := sys.GetBonusDescription()
	if desc != "" {
		t.Errorf("expected empty description without lighting, got %q", desc)
	}

	// With lighting system
	sys.SetLightingSystem(lightingSystem)
	desc = sys.GetBonusDescription()
	// Description should be a string (may be empty if multiplier is exactly 1.0)
	// Just verify it doesn't panic and returns something reasonable
	if len(desc) > 100 {
		t.Errorf("description too long: %q", desc)
	}
}

func TestTimeOfDayXPBonusSystem_MultiplierClamping(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayXPBonusSystem(world, 12345)

	// Test with extreme genre modifier values
	sys.genreModifiers["test"] = map[palette.TimeOfDay]float64{
		palette.TimeOfDayNight: 10.0, // Extreme value
	}
	sys.SetGenre("test")

	mult := sys.getXPMultiplier(palette.TimeOfDayNight)
	if mult > 2.0 {
		t.Errorf("multiplier should be clamped to 2.0, got %v", mult)
	}

	// Test with negative extreme
	sys.genreModifiers["test"] = map[palette.TimeOfDay]float64{
		palette.TimeOfDayDay: -10.0, // Extreme negative
	}

	mult = sys.getXPMultiplier(palette.TimeOfDayDay)
	if mult < 0.5 {
		t.Errorf("multiplier should be clamped to 0.5, got %v", mult)
	}
}

func TestTimeOfDayXPBonusSystem_Determinism(t *testing.T) {
	world1 := NewWorld(nil)
	world2 := NewWorld(nil)

	sys1 := NewTimeOfDayXPBonusSystem(world1, 99999)
	sys2 := NewTimeOfDayXPBonusSystem(world2, 99999)

	// Same seed should produce same RNG state
	for i := 0; i < 10; i++ {
		v1 := sys1.rng.Float64()
		v2 := sys2.rng.Float64()
		if v1 != v2 {
			t.Errorf("iteration %d: RNG values differ: %v != %v", i, v1, v2)
		}
	}
}

func TestTimeOfDayXPBonusSystem_AllGenresHaveModifiers(t *testing.T) {
	world := NewWorld(nil)
	sys := NewTimeOfDayXPBonusSystem(world, 12345)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		if _, ok := sys.genreModifiers[genre]; !ok {
			t.Errorf("genre %s missing from genreModifiers", genre)
		}
	}
}

func TestTimeOfDayItoa(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{10, "10"},
		{123, "123"},
		{-5, "-5"},
		{-123, "-123"},
	}

	for _, tt := range tests {
		got := timeOfDayItoa(tt.input)
		if got != tt.want {
			t.Errorf("timeOfDayItoa(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func BenchmarkTimeOfDayXPBonusSystem_Update(b *testing.B) {
	world := NewWorld(nil)
	sys := NewTimeOfDayXPBonusSystem(world, 12345)
	lightingSystem := NewTimeOfDayLightingSystem(world, 12345)
	sys.SetLightingSystem(lightingSystem)

	// Create 100 player entities
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		e := world.CreateEntity()
		e.AddComponent(NewStubInput())
		e.AddComponent(&ExperienceComponent{Level: 1})
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016) // ~60fps
	}
}

func BenchmarkTimeOfDayXPBonusSystem_GetXPMultiplier(b *testing.B) {
	world := NewWorld(nil)
	sys := NewTimeOfDayXPBonusSystem(world, 12345)
	sys.SetGenre("horror")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sys.getXPMultiplier(palette.TimeOfDayNight)
	}
}
