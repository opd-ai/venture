//go:build ignore

package engine

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/rendering/palette"
)

// mockClockForMovement implements GameClock for testing movement speed system.
type mockClockForMovement struct {
	now time.Time
}

func (c *mockClockForMovement) Now() time.Time {
	return c.now
}

func TestNewTimeOfDayMovementSpeedSystem(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayMovementSpeedSystem(world, 12345)

	if system == nil {
		t.Fatal("NewTimeOfDayMovementSpeedSystem returned nil")
	}

	if system.world != world {
		t.Error("world not set correctly")
	}

	if system.currentMultiplier != 1.0 {
		t.Errorf("initial multiplier = %v, want 1.0", system.currentMultiplier)
	}

	if system.updateInterval != 1.0 {
		t.Errorf("update interval = %v, want 1.0", system.updateInterval)
	}
}

func TestTimeOfDayMovementSpeedSystem_SetLightingSystem(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayMovementSpeedSystem(world, 12345)
	lightingSystem := NewTimeOfDayLightingSystem(world, 12345)

	system.SetLightingSystem(lightingSystem)

	if system.lightingSystem != lightingSystem {
		t.Error("lighting system not set correctly")
	}
}

func TestTimeOfDayMovementSpeedSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayMovementSpeedSystem(world, 12345)

	system.SetGenre("horror")

	if system.genreID != "horror" {
		t.Errorf("genreID = %v, want horror", system.genreID)
	}
}

func TestTimeOfDayMovementSpeedSystem_GetSpeedMultiplier(t *testing.T) {
	tests := []struct {
		name      string
		timeOfDay palette.TimeOfDay
		genre     string
		wantMin   float64
		wantMax   float64
	}{
		{"dawn", palette.TimeOfDayDawn, "", 0.94, 0.96},
		{"day", palette.TimeOfDayDay, "", 0.99, 1.01},
		{"dusk", palette.TimeOfDayDusk, "", 0.89, 0.91},
		{"night", palette.TimeOfDayNight, "", 0.79, 0.81},
		{"night_fantasy", palette.TimeOfDayNight, "fantasy", 0.84, 0.86},
		{"night_horror", palette.TimeOfDayNight, "horror", 0.69, 0.71},
		{"night_scifi", palette.TimeOfDayNight, "scifi", 0.89, 0.91},
		{"night_cyberpunk", palette.TimeOfDayNight, "cyberpunk", 0.87, 0.89},
		{"night_postapoc", palette.TimeOfDayNight, "postapoc", 0.74, 0.76},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewTimeOfDayMovementSpeedSystem(world, 12345)
			if tt.genre != "" {
				system.SetGenre(tt.genre)
			}

			mult := system.getSpeedMultiplier(tt.timeOfDay)

			if mult < tt.wantMin || mult > tt.wantMax {
				t.Errorf("multiplier = %v, want between %v and %v", mult, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestTimeOfDayMovementSpeedSystem_Update_NoLightingSystem(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayMovementSpeedSystem(world, 12345)

	// Create entity with velocity
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&VelocityComponent{VX: 10, VY: 10})

	// Should not panic without lighting system
	system.Update([]*Entity{entity}, 0.016)

	// Velocity should be unchanged
	vel := entity.GetVelocity()
	if vel.VX != 10 || vel.VY != 10 {
		t.Error("velocity should be unchanged without lighting system")
	}
}

func TestTimeOfDayMovementSpeedSystem_Update_Day(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayMovementSpeedSystem(world, 12345)
	lightingSystem := NewTimeOfDayLightingSystem(world, 12345)

	// Set clock to noon (day time)
	clock := &mockClockForMovement{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	lightingSystem.SetClock(clock)

	system.SetLightingSystem(lightingSystem)

	// Update lighting system first to set time of day
	lightingSystem.Update([]*Entity{}, 0.016)

	// Create entity with velocity
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 100})

	// Force time check
	system.timeSinceCheck = 1.0
	system.Update([]*Entity{entity}, 0.016)

	// Day time = 1.0 multiplier, velocity should be unchanged
	vel := entity.GetVelocity()
	if vel.VX < 99 || vel.VX > 101 {
		t.Errorf("velocity X = %v, want ~100 (day time)", vel.VX)
	}
}

func TestTimeOfDayMovementSpeedSystem_Update_Night(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayMovementSpeedSystem(world, 12345)
	lightingSystem := NewTimeOfDayLightingSystem(world, 12345)

	// Set clock to midnight (night time)
	clock := &mockClockForMovement{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	lightingSystem.SetClock(clock)

	system.SetLightingSystem(lightingSystem)

	// Update lighting system first to set time of day
	lightingSystem.Update([]*Entity{}, 0.016)

	// Create entity with velocity
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 100})

	// Force time check by setting enough time since check
	system.timeSinceCheck = 1.0
	system.Update([]*Entity{entity}, 0.016)

	// Night time = 0.8 multiplier, velocity should be reduced
	vel := entity.GetVelocity()
	if vel.VX < 75 || vel.VX > 85 {
		t.Errorf("velocity X = %v, want ~80 (night time)", vel.VX)
	}
}

func TestTimeOfDayMovementSpeedSystem_Update_NightHorror(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayMovementSpeedSystem(world, 12345)
	system.SetGenre("horror")
	lightingSystem := NewTimeOfDayLightingSystem(world, 12345)

	// Set clock to midnight (night time)
	clock := &mockClockForMovement{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	lightingSystem.SetClock(clock)

	system.SetLightingSystem(lightingSystem)

	// Update lighting system first
	lightingSystem.Update([]*Entity{}, 0.016)

	// Create entity with velocity
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 100})

	// Force time check
	system.timeSinceCheck = 1.0
	system.Update([]*Entity{entity}, 0.016)

	// Horror night = 0.8 - 0.10 = 0.70 multiplier
	vel := entity.GetVelocity()
	if vel.VX < 65 || vel.VX > 75 {
		t.Errorf("velocity X = %v, want ~70 (horror night)", vel.VX)
	}
}

func TestTimeOfDayMovementSpeedSystem_Update_NightScifi(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayMovementSpeedSystem(world, 12345)
	system.SetGenre("scifi")
	lightingSystem := NewTimeOfDayLightingSystem(world, 12345)

	// Set clock to midnight (night time)
	clock := &mockClockForMovement{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	lightingSystem.SetClock(clock)

	system.SetLightingSystem(lightingSystem)

	// Update lighting system first
	lightingSystem.Update([]*Entity{}, 0.016)

	// Create entity with velocity
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 100})

	// Force time check
	system.timeSinceCheck = 1.0
	system.Update([]*Entity{entity}, 0.016)

	// Scifi night = 0.8 + 0.10 = 0.90 multiplier
	vel := entity.GetVelocity()
	if vel.VX < 85 || vel.VX > 95 {
		t.Errorf("velocity X = %v, want ~90 (scifi night)", vel.VX)
	}
}

func TestTimeOfDayMovementSpeedSystem_Update_Dusk(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayMovementSpeedSystem(world, 12345)
	lightingSystem := NewTimeOfDayLightingSystem(world, 12345)

	// Set clock to 6 PM (dusk time)
	clock := &mockClockForMovement{now: time.Date(2026, 1, 1, 18, 0, 0, 0, time.UTC)}
	lightingSystem.SetClock(clock)

	system.SetLightingSystem(lightingSystem)

	// Update lighting system first
	lightingSystem.Update([]*Entity{}, 0.016)

	// Create entity with velocity
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 100})

	// Force time check
	system.timeSinceCheck = 1.0
	system.Update([]*Entity{entity}, 0.016)

	// Dusk time = 0.90 multiplier
	vel := entity.GetVelocity()
	if vel.VX < 85 || vel.VX > 95 {
		t.Errorf("velocity X = %v, want ~90 (dusk time)", vel.VX)
	}
}

func TestTimeOfDayMovementSpeedSystem_Update_Dawn(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayMovementSpeedSystem(world, 12345)
	lightingSystem := NewTimeOfDayLightingSystem(world, 12345)

	// Set clock to 6 AM (dawn time)
	clock := &mockClockForMovement{now: time.Date(2026, 1, 1, 6, 0, 0, 0, time.UTC)}
	lightingSystem.SetClock(clock)

	system.SetLightingSystem(lightingSystem)

	// Update lighting system first
	lightingSystem.Update([]*Entity{}, 0.016)

	// Create entity with velocity
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 100})

	// Force time check
	system.timeSinceCheck = 1.0
	system.Update([]*Entity{entity}, 0.016)

	// Dawn time = 0.95 multiplier
	vel := entity.GetVelocity()
	if vel.VX < 90 || vel.VX > 100 {
		t.Errorf("velocity X = %v, want ~95 (dawn time)", vel.VX)
	}
}

func TestTimeOfDayMovementSpeedSystem_GetCurrentMultiplier(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayMovementSpeedSystem(world, 12345)

	// Initial multiplier should be 1.0
	if mult := system.GetCurrentMultiplier(); mult != 1.0 {
		t.Errorf("GetCurrentMultiplier() = %v, want 1.0", mult)
	}

	// Set up lighting system with night time
	lightingSystem := NewTimeOfDayLightingSystem(world, 12345)
	clock := &mockClockForMovement{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	lightingSystem.SetClock(clock)
	system.SetLightingSystem(lightingSystem)

	// Update both systems
	lightingSystem.Update([]*Entity{}, 0.016)
	system.timeSinceCheck = 1.0
	system.Update([]*Entity{}, 0.016)

	// Should now be night multiplier
	mult := system.GetCurrentMultiplier()
	if mult < 0.75 || mult > 0.85 {
		t.Errorf("GetCurrentMultiplier() = %v, want ~0.80", mult)
	}
}

func TestTimeOfDayMovementSpeedSystem_Update_EntityWithoutVelocity(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayMovementSpeedSystem(world, 12345)
	lightingSystem := NewTimeOfDayLightingSystem(world, 12345)

	clock := &mockClockForMovement{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	lightingSystem.SetClock(clock)
	system.SetLightingSystem(lightingSystem)
	lightingSystem.Update([]*Entity{}, 0.016)

	// Create entity without velocity
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Should not panic
	system.timeSinceCheck = 1.0
	system.Update([]*Entity{entity}, 0.016)
}

func TestTimeOfDayMovementSpeedSystem_Update_EntityWithoutPosition(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayMovementSpeedSystem(world, 12345)
	lightingSystem := NewTimeOfDayLightingSystem(world, 12345)

	clock := &mockClockForMovement{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	lightingSystem.SetClock(clock)
	system.SetLightingSystem(lightingSystem)
	lightingSystem.Update([]*Entity{}, 0.016)

	// Create entity with velocity but no position
	entity := world.CreateEntity()
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 100})

	// Should not panic
	system.timeSinceCheck = 1.0
	system.Update([]*Entity{entity}, 0.016)

	// Velocity should be unchanged
	vel := entity.GetVelocity()
	if vel.VX != 100 || vel.VY != 100 {
		t.Error("velocity should be unchanged for entity without position")
	}
}

func TestTimeOfDayMovementSpeedSystem_Update_MultipleEntities(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayMovementSpeedSystem(world, 12345)
	lightingSystem := NewTimeOfDayLightingSystem(world, 12345)

	clock := &mockClockForMovement{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	lightingSystem.SetClock(clock)
	system.SetLightingSystem(lightingSystem)
	lightingSystem.Update([]*Entity{}, 0.016)

	// Create multiple entities
	entities := make([]*Entity, 5)
	for i := 0; i < 5; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 100), Y: 100})
		entity.AddComponent(&VelocityComponent{VX: 100, VY: 100})
		entities[i] = entity
	}

	// Update
	system.timeSinceCheck = 1.0
	system.Update(entities, 0.016)

	// All entities should have reduced velocity (night time)
	for i, entity := range entities {
		vel := entity.GetVelocity()
		if vel.VX < 75 || vel.VX > 85 {
			t.Errorf("entity %d velocity X = %v, want ~80", i, vel.VX)
		}
	}
}

func TestTimeOfDayMovementSpeedSystem_MultiplierClamping(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayMovementSpeedSystem(world, 12345)

	// Test with extreme genre modifier that would push below 0.5
	system.genreNightModifiers["extreme_penalty"] = -0.5
	system.SetGenre("extreme_penalty")

	mult := system.getSpeedMultiplier(palette.TimeOfDayNight)
	if mult < 0.5 {
		t.Errorf("multiplier = %v should be clamped to >= 0.5", mult)
	}

	// Test with extreme genre modifier that would push above 1.1
	system.genreNightModifiers["extreme_bonus"] = 0.5
	system.SetGenre("extreme_bonus")

	mult = system.getSpeedMultiplier(palette.TimeOfDayNight)
	if mult > 1.1 {
		t.Errorf("multiplier = %v should be clamped to <= 1.1", mult)
	}
}

func BenchmarkTimeOfDayMovementSpeedSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewTimeOfDayMovementSpeedSystem(world, 12345)
	lightingSystem := NewTimeOfDayLightingSystem(world, 12345)

	clock := &mockClockForMovement{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	lightingSystem.SetClock(clock)
	system.SetLightingSystem(lightingSystem)
	lightingSystem.Update([]*Entity{}, 0.016)

	// Create 100 entities
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 10), Y: 100})
		entity.AddComponent(&VelocityComponent{VX: 100, VY: 100})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset velocities for consistent benchmark
		for _, e := range entities {
			vel := e.GetVelocity()
			vel.VX = 100
			vel.VY = 100
		}
		system.Update(entities, 0.016)
	}
}
