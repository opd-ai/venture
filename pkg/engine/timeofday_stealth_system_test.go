package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/palette"
)

func TestNewTimeOfDayStealthSystem(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayStealthSystem(world, 12345)

	if system == nil {
		t.Fatal("NewTimeOfDayStealthSystem returned nil")
	}

	if system.world != world {
		t.Error("world reference not set correctly")
	}

	if system.rng == nil {
		t.Error("RNG not initialized")
	}

	if system.updateInterval != 1.0 {
		t.Errorf("update interval = %f, want 1.0", system.updateInterval)
	}

	if system.lastTimeOfDay != palette.TimeOfDayDay {
		t.Errorf("initial time of day = %v, want Day", system.lastTimeOfDay)
	}
}

func TestTimeOfDayStealthSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayStealthSystem(world, 12345)

	system.SetGenre("horror")

	if system.genreID != "horror" {
		t.Errorf("genre = %s, want horror", system.genreID)
	}
}

func TestTimeOfDayStealthSystem_GetDetectionMultiplier(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayStealthSystem(world, 12345)

	tests := []struct {
		name      string
		timeOfDay palette.TimeOfDay
		genre     string
		wantMin   float64
		wantMax   float64
	}{
		{
			name:      "day baseline",
			timeOfDay: palette.TimeOfDayDay,
			genre:     "",
			wantMin:   0.99,
			wantMax:   1.01,
		},
		{
			name:      "dawn reduced visibility",
			timeOfDay: palette.TimeOfDayDawn,
			genre:     "",
			wantMin:   0.84,
			wantMax:   0.86,
		},
		{
			name:      "dusk reduced visibility",
			timeOfDay: palette.TimeOfDayDusk,
			genre:     "",
			wantMin:   0.79,
			wantMax:   0.81,
		},
		{
			name:      "night darkness cover",
			timeOfDay: palette.TimeOfDayNight,
			genre:     "",
			wantMin:   0.54,
			wantMax:   0.56,
		},
		{
			name:      "night horror bonus",
			timeOfDay: palette.TimeOfDayNight,
			genre:     "horror",
			wantMin:   0.39, // 0.55 - 0.15 = 0.40
			wantMax:   0.41,
		},
		{
			name:      "night scifi penalty",
			timeOfDay: palette.TimeOfDayNight,
			genre:     "scifi",
			wantMin:   0.64, // 0.55 + 0.10 = 0.65
			wantMax:   0.66,
		},
		{
			name:      "night fantasy slight bonus",
			timeOfDay: palette.TimeOfDayNight,
			genre:     "fantasy",
			wantMin:   0.49, // 0.55 - 0.05 = 0.50
			wantMax:   0.51,
		},
		{
			name:      "night cyberpunk slight penalty",
			timeOfDay: palette.TimeOfDayNight,
			genre:     "cyberpunk",
			wantMin:   0.59, // 0.55 + 0.05 = 0.60
			wantMax:   0.61,
		},
		{
			name:      "night postapoc bonus",
			timeOfDay: palette.TimeOfDayNight,
			genre:     "postapoc",
			wantMin:   0.44, // 0.55 - 0.10 = 0.45
			wantMax:   0.46,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system.SetGenre(tt.genre)
			got := system.getDetectionMultiplier(tt.timeOfDay)

			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("getDetectionMultiplier(%v) = %f, want between %f and %f",
					tt.timeOfDay, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestTimeOfDayStealthSystem_UpdateAIDetection(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayStealthSystem(world, 12345)

	// Create an entity with AI component
	entity := NewEntity(1)
	ai := NewAIComponent(100, 100)
	ai.DetectionRange = 200.0 // Base detection range
	entity.AddComponent(ai)

	entities := []*Entity{entity}

	// Update for night time
	system.updateAIDetectionRanges(entities, palette.TimeOfDayNight)

	// Detection should be reduced at night (0.55 multiplier)
	expectedRange := 200.0 * 0.55
	if ai.DetectionRange < expectedRange-1 || ai.DetectionRange > expectedRange+1 {
		t.Errorf("detection range = %f, want ~%f", ai.DetectionRange, expectedRange)
	}

	// Original range should be stored
	original, ok := system.GetOriginalRange(entity.ID)
	if !ok {
		t.Error("original range not stored")
	}
	if original != 200.0 {
		t.Errorf("original range = %f, want 200.0", original)
	}
}

func TestTimeOfDayStealthSystem_UpdateWithLightingSystem(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayStealthSystem(world, 12345)
	lightingSystem := NewTimeOfDayLightingSystem(world, 12345)

	system.SetLightingSystem(lightingSystem)

	// Create AI entity
	entity := NewEntity(1)
	ai := NewAIComponent(100, 100)
	ai.DetectionRange = 200.0
	entity.AddComponent(ai)

	entities := []*Entity{entity}

	// First update should not change anything (no time change yet)
	system.Update(entities, 2.0)

	// Detection should still be 200 (day time is default)
	if ai.DetectionRange != 200.0 {
		t.Errorf("initial detection = %f, want 200.0 (no change during day)", ai.DetectionRange)
	}
}

func TestTimeOfDayStealthSystem_UpdateWithoutLightingSystem(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayStealthSystem(world, 12345)

	// No lighting system set - Update should return early without panic
	entity := NewEntity(1)
	ai := NewAIComponent(100, 100)
	ai.DetectionRange = 200.0
	entity.AddComponent(ai)

	entities := []*Entity{entity}

	// Should not panic and not change detection
	system.Update(entities, 2.0)

	if ai.DetectionRange != 200.0 {
		t.Errorf("detection changed without lighting system: %f", ai.DetectionRange)
	}
}

func TestTimeOfDayStealthSystem_GetCurrentMultiplier(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayStealthSystem(world, 12345)

	// Default time is Day, multiplier should be 1.0
	mult := system.GetCurrentMultiplier()
	if mult < 0.99 || mult > 1.01 {
		t.Errorf("current multiplier = %f, want 1.0", mult)
	}
}

func TestTimeOfDayStealthSystem_MultipleEntities(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayStealthSystem(world, 12345)

	// Create multiple entities with different detection ranges
	entities := make([]*Entity, 3)
	for i := range entities {
		entities[i] = NewEntity(uint64(i + 1))
		ai := NewAIComponent(float64(i*100), float64(i*100))
		ai.DetectionRange = float64((i + 1) * 100) // 100, 200, 300
		entities[i].AddComponent(ai)
	}

	// Update for dusk (0.80 multiplier)
	system.updateAIDetectionRanges(entities, palette.TimeOfDayDusk)

	// Check each entity's detection range
	expectedRanges := []float64{80.0, 160.0, 240.0}
	for i, entity := range entities {
		aiComp, _ := entity.GetComponent("ai")
		ai := aiComp.(*AIComponent)
		if ai.DetectionRange < expectedRanges[i]-1 || ai.DetectionRange > expectedRanges[i]+1 {
			t.Errorf("entity %d detection = %f, want ~%f", i+1, ai.DetectionRange, expectedRanges[i])
		}
	}
}

func TestTimeOfDayStealthSystem_IgnoresNonAIEntities(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayStealthSystem(world, 12345)

	// Create entity without AI component
	entity := NewEntity(1)
	pos := &PositionComponent{X: 100, Y: 100}
	entity.AddComponent(pos)

	entities := []*Entity{entity}

	// Should not panic when processing non-AI entities
	system.updateAIDetectionRanges(entities, palette.TimeOfDayNight)

	// No original range should be stored for this entity
	_, ok := system.GetOriginalRange(entity.ID)
	if ok {
		t.Error("original range stored for non-AI entity")
	}
}

func TestTimeOfDayStealthSystem_ClampMultiplier(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayStealthSystem(world, 12345)

	// Test that extreme genre modifiers are clamped
	// Horror at night: 0.55 - 0.15 = 0.40 (within bounds)
	system.SetGenre("horror")
	mult := system.getDetectionMultiplier(palette.TimeOfDayNight)
	if mult < 0.30 {
		t.Errorf("multiplier below minimum: %f", mult)
	}

	// Scifi at night: 0.55 + 0.10 = 0.65 (within bounds)
	system.SetGenre("scifi")
	mult = system.getDetectionMultiplier(palette.TimeOfDayNight)
	if mult > 1.5 {
		t.Errorf("multiplier above maximum: %f", mult)
	}
}

func TestTimeOfDayStealthSystem_SetLightingSystem(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayStealthSystem(world, 12345)
	lightingSystem := NewTimeOfDayLightingSystem(world, 12345)

	if system.lightingSystem != nil {
		t.Error("lighting system should be nil initially")
	}

	system.SetLightingSystem(lightingSystem)

	if system.lightingSystem != lightingSystem {
		t.Error("lighting system not set correctly")
	}
}

func TestTimeOfDayStealthSystem_UpdateIntervalThrottle(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayStealthSystem(world, 12345)
	lightingSystem := NewTimeOfDayLightingSystem(world, 12345)
	system.SetLightingSystem(lightingSystem)

	entity := NewEntity(1)
	ai := NewAIComponent(100, 100)
	ai.DetectionRange = 200.0
	entity.AddComponent(ai)

	entities := []*Entity{entity}

	// Update with small deltaTime - should not trigger check
	system.Update(entities, 0.5)

	// timeSinceCheck should be 0.5
	if system.timeSinceCheck != 0.5 {
		t.Errorf("timeSinceCheck = %f, want 0.5", system.timeSinceCheck)
	}

	// Another small update
	system.Update(entities, 0.3)

	// timeSinceCheck should be 0.8
	if system.timeSinceCheck != 0.8 {
		t.Errorf("timeSinceCheck = %f, want 0.8", system.timeSinceCheck)
	}

	// Update past interval threshold resets counter
	system.Update(entities, 0.3) // Total 1.1 seconds

	// Should reset after check
	if system.timeSinceCheck >= 1.0 {
		t.Errorf("timeSinceCheck not reset: %f", system.timeSinceCheck)
	}
}
