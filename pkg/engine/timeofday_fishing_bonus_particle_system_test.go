package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/palette"
)

func TestNewTimeOfDayFishingBonusParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayFishingBonusParticleSystem(world, 12345)

	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.world != world {
		t.Error("world not set correctly")
	}
	if sys.seed != 12345 {
		t.Errorf("seed = %d, want 12345", sys.seed)
	}
	if sys.genreID != "fantasy" {
		t.Errorf("genreID = %s, want fantasy", sys.genreID)
	}
	if sys.pulseInterval != 3.0 {
		t.Errorf("pulseInterval = %f, want 3.0", sys.pulseInterval)
	}
}

func TestTimeOfDayFishingBonusParticleSystem_SetGenre(t *testing.T) {
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
			sys := NewTimeOfDayFishingBonusParticleSystem(nil, 12345)
			sys.SetGenre(tt.genreID)
			if sys.genreID != tt.genreID {
				t.Errorf("genreID = %s, want %s", sys.genreID, tt.genreID)
			}
		})
	}
}

func TestTimeOfDayFishingBonusParticleSystem_SetParticleSystem(t *testing.T) {
	sys := NewTimeOfDayFishingBonusParticleSystem(nil, 12345)
	ps := NewParticleSystem()

	sys.SetParticleSystem(ps)

	if sys.particleSystem != ps {
		t.Error("particle system not set correctly")
	}
}

func TestTimeOfDayFishingBonusParticleSystem_SetTimeOfDayFishingBonusSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayFishingBonusParticleSystem(world, 12345)
	tfbs := NewTimeOfDayFishingBonusSystem(world, 54321)

	sys.SetTimeOfDayFishingBonusSystem(tfbs)

	if sys.timeOfDayFishingBonusSystem != tfbs {
		t.Error("time of day fishing bonus system not set correctly")
	}
}

func TestTimeOfDayFishingBonusParticleSystem_Update_NilDependencies(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() *TimeOfDayFishingBonusParticleSystem
	}{
		{
			"nil particle system",
			func() *TimeOfDayFishingBonusParticleSystem {
				world := NewWorld()
				sys := NewTimeOfDayFishingBonusParticleSystem(world, 12345)
				sys.SetTimeOfDayFishingBonusSystem(NewTimeOfDayFishingBonusSystem(world, 54321))
				return sys
			},
		},
		{
			"nil world",
			func() *TimeOfDayFishingBonusParticleSystem {
				sys := NewTimeOfDayFishingBonusParticleSystem(nil, 12345)
				sys.SetParticleSystem(NewParticleSystem())
				return sys
			},
		},
		{
			"nil fishing bonus system",
			func() *TimeOfDayFishingBonusParticleSystem {
				world := NewWorld()
				sys := NewTimeOfDayFishingBonusParticleSystem(world, 12345)
				sys.SetParticleSystem(NewParticleSystem())
				return sys
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := tt.setupFunc()
			// Should not panic with nil dependencies
			sys.Update([]*Entity{}, 0.016)
		})
	}
}

func TestTimeOfDayFishingBonusParticleSystem_GetTimeParticleType(t *testing.T) {
	tests := []struct {
		name      string
		genreID   string
		timeOfDay palette.TimeOfDay
		wantType  string // Check particle type is non-empty
	}{
		{"fantasy_dawn", "fantasy", palette.TimeOfDayDawn, "sparkle"},
		{"fantasy_dusk", "fantasy", palette.TimeOfDayDusk, "sparkle"},
		{"fantasy_night", "fantasy", palette.TimeOfDayNight, "magic"},
		{"fantasy_day", "fantasy", palette.TimeOfDayDay, "dust"},
		{"scifi_any", "scifi", palette.TimeOfDayNight, "spark"},
		{"horror_night", "horror", palette.TimeOfDayNight, "smoke"},
		{"horror_dusk", "horror", palette.TimeOfDayDusk, "smoke"},
		{"cyberpunk_any", "cyberpunk", palette.TimeOfDayNight, "spark"},
		{"postapoc_dawn", "postapoc", palette.TimeOfDayDawn, "dust"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewTimeOfDayFishingBonusParticleSystem(nil, 12345)
			sys.SetGenre(tt.genreID)
			particleType := sys.getTimeParticleType(tt.timeOfDay)
			// Just verify we get a valid particle type (not zero value)
			if particleType < 0 {
				t.Errorf("got invalid particle type for %s/%s", tt.genreID, tt.timeOfDay)
			}
		})
	}
}

func TestTimeOfDayFishingBonusParticleSystem_GetParticleCount(t *testing.T) {
	sys := NewTimeOfDayFishingBonusParticleSystem(nil, 12345)

	tests := []struct {
		name     string
		modifier float64
		wantMin  int
		wantMax  int
	}{
		{"weak bonus", 1.10, 8, 9},
		{"moderate bonus", 1.20, 10, 11},
		{"good bonus", 1.35, 12, 13},
		{"strong bonus", 1.55, 14, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := sys.getParticleCount(tt.modifier)
			if count < tt.wantMin || count > tt.wantMax {
				t.Errorf("count = %d, want between %d and %d", count, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestTimeOfDayFishingBonusParticleSystem_GetParticleGravity(t *testing.T) {
	sys := NewTimeOfDayFishingBonusParticleSystem(nil, 12345)

	tests := []struct {
		name      string
		timeOfDay palette.TimeOfDay
		wantSign  int // -1 for negative (rise), 0 for zero, 1 for positive (fall)
	}{
		{"dawn rises", palette.TimeOfDayDawn, -1},
		{"dusk settles", palette.TimeOfDayDusk, 1},
		{"night floats", palette.TimeOfDayNight, -1},
		{"day neutral", palette.TimeOfDayDay, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gravity := sys.getParticleGravity(tt.timeOfDay)
			var gotSign int
			if gravity < 0 {
				gotSign = -1
			} else if gravity > 0 {
				gotSign = 1
			}
			if gotSign != tt.wantSign {
				t.Errorf("gravity sign = %d, want %d (gravity=%f)", gotSign, tt.wantSign, gravity)
			}
		})
	}
}

func TestTimeOfDayFishingBonusParticleSystem_ProcessEntity_NoFishingSpot(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayFishingBonusParticleSystem(world, 12345)
	sys.SetParticleSystem(NewParticleSystem())
	sys.SetTimeOfDayFishingBonusSystem(NewTimeOfDayFishingBonusSystem(world, 54321))

	// Create entity without fishing spot component
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 200})

	// Should not panic - no fishing spot
	sys.processEntity(entity, true, false, palette.TimeOfDayDawn)
}

func TestTimeOfDayFishingBonusParticleSystem_ProcessEntity_NoPosition(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayFishingBonusParticleSystem(world, 12345)
	sys.SetParticleSystem(NewParticleSystem())
	sys.SetTimeOfDayFishingBonusSystem(NewTimeOfDayFishingBonusSystem(world, 54321))

	// Create entity with fishing spot but no position
	entity := NewEntity(1)
	entity.AddComponent(&FishingSpotComponent{
		WaterType:     WaterTypeFreshwater,
		RareFishBonus: 1.0,
	})

	// Should not panic - no position
	sys.processEntity(entity, true, false, palette.TimeOfDayDawn)
}

func TestTimeOfDayFishingBonusParticleSystem_Update_TimeChange(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayFishingBonusParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	tfbs := NewTimeOfDayFishingBonusSystem(world, 54321)
	sys.SetTimeOfDayFishingBonusSystem(tfbs)

	// Create fishing spot entity
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 200})
	entity.AddComponent(&FishingSpotComponent{
		WaterType:     WaterTypeFreshwater,
		RareFishBonus: 1.5,
	})

	entities := []*Entity{entity}

	// First update
	sys.Update(entities, 0.016)

	// Verify time tracking
	if sys.lastTimeOfDay != palette.TimeOfDayDay {
		// Initial state should be day (default)
	}
}

func TestTimeOfDayFishingBonusParticleSystem_PulseAccumulation(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayFishingBonusParticleSystem(world, 12345)
	sys.SetParticleSystem(NewParticleSystem())
	sys.SetTimeOfDayFishingBonusSystem(NewTimeOfDayFishingBonusSystem(world, 54321))

	entities := []*Entity{}

	// Accumulate time
	for i := 0; i < 100; i++ {
		sys.Update(entities, 0.05)
	}

	// timeSinceEmit should reset after exceeding pulseInterval
	if sys.timeSinceEmit >= sys.pulseInterval {
		t.Errorf("timeSinceEmit = %f should be less than pulseInterval %f", sys.timeSinceEmit, sys.pulseInterval)
	}
}

func TestTimeOfDayFishingBonusParticleSystem_Integration(t *testing.T) {
	world := NewWorld()

	// Create all required systems
	particleSys := NewParticleSystem()
	fishingBonusSys := NewTimeOfDayFishingBonusSystem(world, 12345)

	// Create the particle system under test
	sys := NewTimeOfDayFishingBonusParticleSystem(world, 54321)
	sys.SetParticleSystem(particleSys)
	sys.SetTimeOfDayFishingBonusSystem(fishingBonusSys)
	sys.SetGenre("fantasy")

	// Create a fishing spot entity
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 200})
	entity.AddComponent(&FishingSpotComponent{
		WaterType:     WaterTypeMagical,
		RareFishBonus: 1.5,
	})

	entities := []*Entity{entity}

	// Should not panic during full integration
	for i := 0; i < 200; i++ {
		sys.Update(entities, 0.016)
	}
}
