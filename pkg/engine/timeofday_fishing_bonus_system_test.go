package engine

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/rendering/palette"
)

// mockTimeOfDayClock implements GameClock for testing
type mockTimeOfDayClock struct {
	hour int
}

func (c *mockTimeOfDayClock) Now() time.Time {
	return time.Date(2026, 1, 1, c.hour, 0, 0, 0, time.UTC)
}

func TestNewTimeOfDayFishingBonusSystem(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayFishingBonusSystem(world, 12345)

	if system == nil {
		t.Fatal("NewTimeOfDayFishingBonusSystem returned nil")
	}

	if system.world != world {
		t.Error("World not set correctly")
	}

	if system.updateInterval != 1.0 {
		t.Errorf("updateInterval = %v, want 1.0", system.updateInterval)
	}

	if system.genreID != "fantasy" {
		t.Errorf("genreID = %v, want fantasy", system.genreID)
	}
}

func TestTimeOfDayFishingBonusSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayFishingBonusSystem(world, 12345)

	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"scifi", "scifi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system.SetGenre(tt.genreID)
			if system.genreID != tt.genreID {
				t.Errorf("genreID = %v, want %v", system.genreID, tt.genreID)
			}
		})
	}
}

func TestTimeOfDayFishingBonusSystem_SetLightingSystem(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayFishingBonusSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	system.SetLightingSystem(lighting)

	if system.lightingSystem != lighting {
		t.Error("lightingSystem not set correctly")
	}
}

func TestTimeOfDayFishingBonusSystem_SetFishingSystem(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayFishingBonusSystem(world, 12345)
	fishing := NewFishingSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	system.SetLightingSystem(lighting)
	system.SetFishingSystem(fishing)

	if system.fishingSystem != fishing {
		t.Error("fishingSystem not set correctly")
	}

	// Verify callback was set
	if fishing.CurrentTimeOfDay == nil {
		t.Error("CurrentTimeOfDay callback not set on FishingSystem")
	}
}

func TestTimeOfDayFishingBonusSystem_GetBaseModifier(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayFishingBonusSystem(world, 12345)

	tests := []struct {
		name      string
		timeOfDay palette.TimeOfDay
		waterType WaterType
		wantMin   float64
		wantMax   float64
	}{
		{"dawn freshwater", palette.TimeOfDayDawn, WaterTypeFreshwater, 1.30, 1.40},
		{"dawn saltwater", palette.TimeOfDayDawn, WaterTypeSaltwater, 1.20, 1.30},
		{"dawn magical", palette.TimeOfDayDawn, WaterTypeMagical, 1.20, 1.30},
		{"day freshwater", palette.TimeOfDayDay, WaterTypeFreshwater, 0.99, 1.01},
		{"day saltwater", palette.TimeOfDayDay, WaterTypeSaltwater, 0.99, 1.01},
		{"dusk saltwater", palette.TimeOfDayDusk, WaterTypeSaltwater, 1.35, 1.45},
		{"dusk freshwater", palette.TimeOfDayDusk, WaterTypeFreshwater, 1.25, 1.35},
		{"night magical", palette.TimeOfDayNight, WaterTypeMagical, 1.55, 1.65},
		{"night freshwater", palette.TimeOfDayNight, WaterTypeFreshwater, 1.40, 1.50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := system.getBaseModifier(tt.timeOfDay, tt.waterType)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("getBaseModifier(%v, %v) = %v, want between %v and %v",
					tt.timeOfDay, tt.waterType, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestTimeOfDayFishingBonusSystem_GetGenreBonus(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayFishingBonusSystem(world, 12345)

	tests := []struct {
		name      string
		genreID   string
		timeOfDay palette.TimeOfDay
		wantMin   float64
		wantMax   float64
	}{
		{"fantasy dawn", "fantasy", palette.TimeOfDayDawn, 1.10, 1.20},
		{"fantasy night", "fantasy", palette.TimeOfDayNight, 1.10, 1.20},
		{"fantasy day", "fantasy", palette.TimeOfDayDay, 0.99, 1.01},
		{"scifi any", "scifi", palette.TimeOfDayDay, 0.90, 1.00},
		{"horror night", "horror", palette.TimeOfDayNight, 1.30, 1.40},
		{"horror dusk", "horror", palette.TimeOfDayDusk, 1.15, 1.25},
		{"horror day", "horror", palette.TimeOfDayDay, 0.99, 1.01},
		{"cyberpunk night", "cyberpunk", palette.TimeOfDayNight, 1.20, 1.30},
		{"cyberpunk day", "cyberpunk", palette.TimeOfDayDay, 0.99, 1.01},
		{"postapoc dawn", "postapoc", palette.TimeOfDayDawn, 1.15, 1.25},
		{"postapoc day", "postapoc", palette.TimeOfDayDay, 0.99, 1.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system.SetGenre(tt.genreID)
			got := system.getGenreBonus(tt.timeOfDay)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("getGenreBonus(%v) with genre %v = %v, want between %v and %v",
					tt.timeOfDay, tt.genreID, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestTimeOfDayFishingBonusSystem_CalculateBonusModifier(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayFishingBonusSystem(world, 12345)

	// Test combined modifier (base * genre)
	system.SetGenre("horror")

	// Night + horror = high bonus
	modifier := system.calculateBonusModifier(palette.TimeOfDayNight, WaterTypeMagical)
	// Night magical base = 1.60, horror night genre = 1.35
	// Expected: 1.60 * 1.35 = 2.16
	if modifier < 2.0 || modifier > 2.3 {
		t.Errorf("calculateBonusModifier(night, magical) with horror = %v, want ~2.16", modifier)
	}

	// Day + horror = no bonus
	modifier = system.calculateBonusModifier(palette.TimeOfDayDay, WaterTypeFreshwater)
	if modifier < 0.99 || modifier > 1.01 {
		t.Errorf("calculateBonusModifier(day, freshwater) with horror = %v, want ~1.0", modifier)
	}
}

func TestTimeOfDayFishingBonusSystem_UpdateFishingBonuses(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayFishingBonusSystem(world, 12345)

	// Create fishing spot entity
	spot := NewEntity(1)
	spotComp := NewFishingSpotComponent(WaterTypeFreshwater, DepthMedium, "forest")
	originalBonus := 1.5
	spotComp.RareFishBonus = originalBonus
	spot.AddComponent(spotComp)

	entities := []*Entity{spot}

	// Apply dawn bonus
	system.updateFishingBonuses(entities, palette.TimeOfDayDawn)

	// Dawn freshwater base = 1.35, fantasy genre = 1.15
	// Expected: 1.5 * 1.35 * 1.15 = 2.33
	if spotComp.RareFishBonus < 2.0 || spotComp.RareFishBonus > 2.5 {
		t.Errorf("RareFishBonus after dawn update = %v, want ~2.33", spotComp.RareFishBonus)
	}

	// Verify original was cached
	if stored, exists := system.originalBonuses[spot.ID]; !exists || stored != originalBonus {
		t.Errorf("Original bonus not cached: exists=%v, stored=%v, want %v", exists, stored, originalBonus)
	}

	// Apply day (neutral)
	system.updateFishingBonuses(entities, palette.TimeOfDayDay)

	// Day = 1.0, no genre bonus = 1.0
	// Expected: original * 1.0 = 1.5
	if spotComp.RareFishBonus < 1.4 || spotComp.RareFishBonus > 1.6 {
		t.Errorf("RareFishBonus after day update = %v, want ~1.5", spotComp.RareFishBonus)
	}
}

func TestTimeOfDayFishingBonusSystem_Update(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayFishingBonusSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	// Set up clock for dawn (hour 6)
	clock := &mockTimeOfDayClock{hour: 6}
	lighting.SetClock(clock)

	system.SetLightingSystem(lighting)

	// Create fishing spot
	spot := NewEntity(1)
	spotComp := NewFishingSpotComponent(WaterTypeSaltwater, DepthDeep, "ocean")
	spotComp.RareFishBonus = 1.0
	spot.AddComponent(spotComp)

	entities := []*Entity{spot}

	// Run lighting system to establish time
	lighting.Update(entities, 0.016)

	// Run fishing bonus system (needs enough delta to trigger check)
	system.Update(entities, 2.0)

	// Should have applied dawn bonus
	if spotComp.RareFishBonus < 1.1 {
		t.Errorf("RareFishBonus not updated on dawn: %v", spotComp.RareFishBonus)
	}
}

func TestTimeOfDayFishingBonusSystem_GetCurrentTimeOfDayForFishing(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayFishingBonusSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	system.SetLightingSystem(lighting)

	tests := []struct {
		name     string
		hour     int
		wantTime TimeOfDay
	}{
		{"dawn", 6, TimeDawn},
		{"day", 12, TimeDay},
		{"dusk", 18, TimeDusk},
		{"night", 22, TimeNight},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clock := &mockTimeOfDayClock{hour: tt.hour}
			lighting.SetClock(clock)
			lighting.Update([]*Entity{}, 0.016)

			got := system.getCurrentTimeOfDayForFishing()
			if got != tt.wantTime {
				t.Errorf("getCurrentTimeOfDayForFishing() at hour %d = %v, want %v",
					tt.hour, got, tt.wantTime)
			}
		})
	}
}

func TestTimeOfDayFishingBonusSystem_GetBonusModifier(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayFishingBonusSystem(world, 12345)

	// Default is day
	modifier := system.GetBonusModifier(WaterTypeFreshwater)
	if modifier < 0.99 || modifier > 1.01 {
		t.Errorf("GetBonusModifier(freshwater) default = %v, want ~1.0", modifier)
	}

	// Simulate night
	system.lastTimeOfDay = palette.TimeOfDayNight
	modifier = system.GetBonusModifier(WaterTypeMagical)

	// Night magical = 1.60, fantasy genre night = 1.15
	// Expected: ~1.84
	if modifier < 1.7 || modifier > 2.0 {
		t.Errorf("GetBonusModifier(magical) at night = %v, want ~1.84", modifier)
	}
}

func TestTimeOfDayFishingBonusSystem_NoLightingSystem(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayFishingBonusSystem(world, 12345)

	// Without lighting system, should return day
	got := system.getCurrentTimeOfDayForFishing()
	if got != TimeDay {
		t.Errorf("getCurrentTimeOfDayForFishing() without lighting = %v, want %v", got, TimeDay)
	}
}

func TestTimeOfDayFishingBonusSystem_MultipleSpots(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayFishingBonusSystem(world, 12345)

	// Create multiple fishing spots with different water types
	spots := []*Entity{
		NewEntity(1),
		NewEntity(2),
		NewEntity(3),
	}

	freshwater := NewFishingSpotComponent(WaterTypeFreshwater, DepthShallow, "lake")
	freshwater.RareFishBonus = 1.0
	spots[0].AddComponent(freshwater)

	saltwater := NewFishingSpotComponent(WaterTypeSaltwater, DepthDeep, "ocean")
	saltwater.RareFishBonus = 1.2
	spots[1].AddComponent(saltwater)

	magical := NewFishingSpotComponent(WaterTypeMagical, DepthMedium, "enchanted")
	magical.RareFishBonus = 1.5
	spots[2].AddComponent(magical)

	// Apply night bonus
	system.updateFishingBonuses(spots, palette.TimeOfDayNight)

	// Verify each spot got appropriate bonus
	// Night freshwater = 1.45, night saltwater = 1.45, night magical = 1.60
	// Fantasy night genre = 1.15
	if freshwater.RareFishBonus < 1.5 {
		t.Errorf("Freshwater bonus at night = %v, expected higher", freshwater.RareFishBonus)
	}
	if saltwater.RareFishBonus < 1.8 {
		t.Errorf("Saltwater bonus at night = %v, expected higher", saltwater.RareFishBonus)
	}
	if magical.RareFishBonus < 2.5 {
		t.Errorf("Magical bonus at night = %v, expected higher", magical.RareFishBonus)
	}
}

func BenchmarkTimeOfDayFishingBonusSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewTimeOfDayFishingBonusSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)
	clock := &mockTimeOfDayClock{hour: 6}
	lighting.SetClock(clock)
	system.SetLightingSystem(lighting)

	// Create 100 fishing spots
	entities := make([]*Entity, 100)
	for i := range entities {
		entity := NewEntity(uint64(i))
		spot := NewFishingSpotComponent(WaterTypeFreshwater, DepthMedium, "lake")
		spot.RareFishBonus = 1.0 + float64(i%5)*0.1
		entity.AddComponent(spot)
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}
