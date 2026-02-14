//go:build ignore

package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/palette"
)

func TestNewTimeOfDayHealthRegenSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayHealthRegenSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewTimeOfDayHealthRegenSystem returned nil")
	}
	if sys.world != world {
		t.Error("world not set correctly")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("default genre = %s, want fantasy", sys.genreID)
	}
	if sys.updateInterval != 1.0 {
		t.Errorf("updateInterval = %f, want 1.0", sys.updateInterval)
	}
}

func TestTimeOfDayHealthRegenSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayHealthRegenSystem(world, 12345)

	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("genre = %s, want horror", sys.genreID)
	}
}

func TestTimeOfDayHealthRegenSystem_SetLightingSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayHealthRegenSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 54321)

	sys.SetLightingSystem(lightingSys)
	if sys.lightingSystem != lightingSys {
		t.Error("lighting system not set correctly")
	}
}

func TestTimeOfDayHealthRegenSystem_Update_NoLightingSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayHealthRegenSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 50, Max: 100, Regen: 5.0})

	// Should not panic and should not modify health
	sys.Update([]*Entity{entity}, 1.0)

	healthComp, _ := entity.GetComponent("health")
	health := healthComp.(*HealthComponent)
	if health.Regen != 5.0 {
		t.Errorf("Regen = %f, want 5.0 (unmodified)", health.Regen)
	}
}

func TestTimeOfDayHealthRegenSystem_BaseMultipliers(t *testing.T) {
	tests := []struct {
		name      string
		timeOfDay palette.TimeOfDay
		wantBase  float64
	}{
		{"dawn", palette.TimeOfDayDawn, 1.10},
		{"day", palette.TimeOfDayDay, 1.0},
		{"dusk", palette.TimeOfDayDusk, 1.05},
		{"night", palette.TimeOfDayNight, 1.20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewTimeOfDayHealthRegenSystem(world, 12345)

			base := sys.baseMultipliers[tt.timeOfDay]
			if base != tt.wantBase {
				t.Errorf("baseMultipliers[%s] = %f, want %f", tt.timeOfDay.String(), base, tt.wantBase)
			}
		})
	}
}

func TestTimeOfDayHealthRegenSystem_GenreFantasy(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayHealthRegenSystem(world, 12345)
	sys.SetGenre("fantasy")

	// Night in fantasy should get +10% bonus on top of base 1.20
	nightMult := sys.getHealthRegenMultiplier(palette.TimeOfDayNight)
	expectedNight := 1.30 // 1.20 base + 0.10 genre
	if nightMult != expectedNight {
		t.Errorf("fantasy night multiplier = %f, want %f", nightMult, expectedNight)
	}

	// Dawn in fantasy should get +5% bonus on top of base 1.10
	dawnMult := sys.getHealthRegenMultiplier(palette.TimeOfDayDawn)
	expectedDawn := 1.15 // 1.10 base + 0.05 genre
	if dawnMult != expectedDawn {
		t.Errorf("fantasy dawn multiplier = %f, want %f", dawnMult, expectedDawn)
	}
}

func TestTimeOfDayHealthRegenSystem_GenreHorror(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayHealthRegenSystem(world, 12345)
	sys.SetGenre("horror")

	// Day in horror should get +15% bonus on top of base 1.0
	dayMult := sys.getHealthRegenMultiplier(palette.TimeOfDayDay)
	expectedDay := 1.15 // 1.0 base + 0.15 genre
	if dayMult != expectedDay {
		t.Errorf("horror day multiplier = %f, want %f", dayMult, expectedDay)
	}

	// Night in horror should get -15% penalty on base 1.20
	nightMult := sys.getHealthRegenMultiplier(palette.TimeOfDayNight)
	expectedNight := 1.05 // 1.20 base - 0.15 genre
	if nightMult != expectedNight {
		t.Errorf("horror night multiplier = %f, want %f", nightMult, expectedNight)
	}
}

func TestTimeOfDayHealthRegenSystem_GenreScifi(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayHealthRegenSystem(world, 12345)
	sys.SetGenre("scifi")

	// Day in scifi should get +5% bonus on top of base 1.0
	dayMult := sys.getHealthRegenMultiplier(palette.TimeOfDayDay)
	expectedDay := 1.05 // 1.0 base + 0.05 genre
	if dayMult != expectedDay {
		t.Errorf("scifi day multiplier = %f, want %f", dayMult, expectedDay)
	}

	// Night in scifi should get -10% penalty on base 1.20
	nightMult := sys.getHealthRegenMultiplier(palette.TimeOfDayNight)
	expectedNight := 1.10 // 1.20 base - 0.10 genre
	if nightMult != expectedNight {
		t.Errorf("scifi night multiplier = %f, want %f", nightMult, expectedNight)
	}
}

func TestTimeOfDayHealthRegenSystem_GenreCyberpunk(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayHealthRegenSystem(world, 12345)
	sys.SetGenre("cyberpunk")

	// Night in cyberpunk should get +5% bonus on top of base 1.20
	nightMult := sys.getHealthRegenMultiplier(palette.TimeOfDayNight)
	expectedNight := 1.25 // 1.20 base + 0.05 genre
	if nightMult != expectedNight {
		t.Errorf("cyberpunk night multiplier = %f, want %f", nightMult, expectedNight)
	}

	// Day in cyberpunk should get -5% penalty on base 1.0
	dayMult := sys.getHealthRegenMultiplier(palette.TimeOfDayDay)
	expectedDay := 0.95 // 1.0 base - 0.05 genre
	if dayMult != expectedDay {
		t.Errorf("cyberpunk day multiplier = %f, want %f", dayMult, expectedDay)
	}
}

func TestTimeOfDayHealthRegenSystem_GenrePostapoc(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayHealthRegenSystem(world, 12345)
	sys.SetGenre("postapoc")

	// Dawn in postapoc should get +10% bonus on top of base 1.10
	dawnMult := sys.getHealthRegenMultiplier(palette.TimeOfDayDawn)
	expectedDawn := 1.20 // 1.10 base + 0.10 genre
	if dawnMult != expectedDawn {
		t.Errorf("postapoc dawn multiplier = %f, want %f", dawnMult, expectedDawn)
	}

	// Day in postapoc should get -10% penalty on base 1.0
	dayMult := sys.getHealthRegenMultiplier(palette.TimeOfDayDay)
	expectedDay := 0.90 // 1.0 base - 0.10 genre
	if dayMult != expectedDay {
		t.Errorf("postapoc day multiplier = %f, want %f", dayMult, expectedDay)
	}
}

func TestTimeOfDayHealthRegenSystem_UpdateModifiesRegen(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayHealthRegenSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 54321)
	sys.SetLightingSystem(lightingSys)
	sys.SetGenre("fantasy")

	// Set lighting system to night
	lightingSys.ForceTimeOfDay(palette.TimeOfDayNight)

	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 50, Max: 100, Regen: 10.0})

	// Wait for update interval
	sys.Update([]*Entity{entity}, 1.5)

	healthComp, _ := entity.GetComponent("health")
	health := healthComp.(*HealthComponent)

	// Expected: 10.0 * 1.30 = 13.0 (fantasy night bonus)
	expectedRegen := 13.0
	if health.Regen != expectedRegen {
		t.Errorf("Regen = %f, want %f", health.Regen, expectedRegen)
	}
}

func TestTimeOfDayHealthRegenSystem_GetActiveMultiplier(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayHealthRegenSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 54321)
	sys.SetLightingSystem(lightingSys)
	sys.SetGenre("fantasy")

	lightingSys.ForceTimeOfDay(palette.TimeOfDayNight)

	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 50, Max: 100, Regen: 10.0})

	// Before update, should return default 1.0
	if mult := sys.GetActiveMultiplier(entity.ID); mult != 1.0 {
		t.Errorf("GetActiveMultiplier before update = %f, want 1.0", mult)
	}

	sys.Update([]*Entity{entity}, 1.5)

	// After update, should return calculated multiplier
	expected := 1.30 // fantasy night
	if mult := sys.GetActiveMultiplier(entity.ID); mult != expected {
		t.Errorf("GetActiveMultiplier after update = %f, want %f", mult, expected)
	}
}

func TestTimeOfDayHealthRegenSystem_GetCurrentMultiplier(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayHealthRegenSystem(world, 12345)

	// Without lighting system, should return 1.0
	if mult := sys.GetCurrentMultiplier(); mult != 1.0 {
		t.Errorf("GetCurrentMultiplier without lighting = %f, want 1.0", mult)
	}

	lightingSys := NewTimeOfDayLightingSystem(world, 54321)
	sys.SetLightingSystem(lightingSys)
	sys.SetGenre("horror")

	lightingSys.ForceTimeOfDay(palette.TimeOfDayDay)

	// Horror day should be 1.15
	if mult := sys.GetCurrentMultiplier(); mult != 1.15 {
		t.Errorf("GetCurrentMultiplier horror day = %f, want 1.15", mult)
	}
}

func TestTimeOfDayHealthRegenSystem_GetBonusDescription(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayHealthRegenSystem(world, 12345)

	// Without lighting system, should return empty
	if desc := sys.GetBonusDescription(); desc != "" {
		t.Errorf("GetBonusDescription without lighting = %q, want empty", desc)
	}

	lightingSys := NewTimeOfDayLightingSystem(world, 54321)
	sys.SetLightingSystem(lightingSys)
	sys.SetGenre("fantasy")

	lightingSys.ForceTimeOfDay(palette.TimeOfDayNight)
	desc := sys.GetBonusDescription()
	if desc == "" {
		t.Error("GetBonusDescription returned empty for bonus")
	}
	// Should contain "+" for bonus
	found := false
	for _, c := range desc {
		if c == '+' {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("GetBonusDescription = %q, expected to contain +", desc)
	}
}

func TestTimeOfDayHealthRegenSystem_RestoreOriginalRegen(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayHealthRegenSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 54321)
	sys.SetLightingSystem(lightingSys)
	sys.SetGenre("fantasy")
	lightingSys.ForceTimeOfDay(palette.TimeOfDayNight)

	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 50, Max: 100, Regen: 10.0})

	// Modify regen
	sys.Update([]*Entity{entity}, 1.5)

	healthComp, _ := entity.GetComponent("health")
	health := healthComp.(*HealthComponent)
	if health.Regen == 10.0 {
		t.Error("Regen should be modified after Update")
	}

	// Restore
	sys.RestoreOriginalRegen(entity.ID)

	if health.Regen != 10.0 {
		t.Errorf("Regen after restore = %f, want 10.0", health.Regen)
	}

	// Multiplier should be cleared
	if mult := sys.GetActiveMultiplier(entity.ID); mult != 1.0 {
		t.Errorf("GetActiveMultiplier after restore = %f, want 1.0", mult)
	}
}

func TestTimeOfDayHealthRegenSystem_MultiplierClamping(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayHealthRegenSystem(world, 12345)

	// Manually set extreme modifiers to test clamping
	sys.genreModifiers["extreme"] = map[palette.TimeOfDay]float64{
		palette.TimeOfDayNight: 5.0, // Would push to 6.20, should clamp to 2.0
	}
	sys.SetGenre("extreme")

	mult := sys.getHealthRegenMultiplier(palette.TimeOfDayNight)
	if mult != 2.0 {
		t.Errorf("extreme night multiplier = %f, want 2.0 (clamped)", mult)
	}

	// Test lower bound
	sys.genreModifiers["negative"] = map[palette.TimeOfDay]float64{
		palette.TimeOfDayDay: -2.0, // Would push to -1.0, should clamp to 0.5
	}
	sys.SetGenre("negative")

	mult = sys.getHealthRegenMultiplier(palette.TimeOfDayDay)
	if mult != 0.5 {
		t.Errorf("negative day multiplier = %f, want 0.5 (clamped)", mult)
	}
}

func TestTimeOfDayHealthRegenSystem_NoHealthComponent(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayHealthRegenSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 54321)
	sys.SetLightingSystem(lightingSys)

	// Entity without health component
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})

	// Should not panic
	sys.Update([]*Entity{entity}, 1.5)
}

func TestTimeOfDayHealthRegenSystem_UpdateInterval(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayHealthRegenSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 54321)
	sys.SetLightingSystem(lightingSys)
	lightingSys.ForceTimeOfDay(palette.TimeOfDayNight)

	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 50, Max: 100, Regen: 10.0})

	// First update with small delta time - should not modify yet
	sys.Update([]*Entity{entity}, 0.1)

	healthComp, _ := entity.GetComponent("health")
	health := healthComp.(*HealthComponent)
	if health.Regen != 10.0 {
		t.Errorf("Regen after short update = %f, want 10.0 (unchanged)", health.Regen)
	}

	// Update again to exceed interval
	sys.Update([]*Entity{entity}, 1.0)

	if health.Regen == 10.0 {
		t.Error("Regen should be modified after exceeding interval")
	}
}

func TestHealthRegenItoa(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{10, "10"},
		{123, "123"},
		{-5, "-5"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := healthRegenItoa(tt.input)
			if got != tt.want {
				t.Errorf("healthRegenItoa(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func BenchmarkTimeOfDayHealthRegenSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewTimeOfDayHealthRegenSystem(world, 12345)
	lightingSys := NewTimeOfDayLightingSystem(world, 54321)
	sys.SetLightingSystem(lightingSys)

	entities := make([]*Entity, 100)
	for i := range entities {
		e := world.CreateEntity()
		e.AddComponent(&HealthComponent{Current: 50, Max: 100, Regen: 5.0})
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 1.5)
	}
}

func BenchmarkTimeOfDayHealthRegenSystem_GetMultiplier(b *testing.B) {
	world := NewWorld()
	sys := NewTimeOfDayHealthRegenSystem(world, 12345)
	sys.SetGenre("fantasy")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.getHealthRegenMultiplier(palette.TimeOfDayNight)
	}
}
