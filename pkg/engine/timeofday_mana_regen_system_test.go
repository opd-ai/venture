package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/palette"
)

func TestNewTimeOfDayManaRegenSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayManaRegenSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewTimeOfDayManaRegenSystem returned nil")
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
	if len(sys.genreModifiers) != 5 {
		t.Errorf("expected 5 genre modifiers, got %d", len(sys.genreModifiers))
	}
}

func TestTimeOfDayManaRegenSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayManaRegenSystem(world, 12345)

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
			sys.SetGenre(tt.genreID)
			if sys.genreID != tt.genreID {
				t.Errorf("genre not set: got %s, want %s", sys.genreID, tt.genreID)
			}
		})
	}
}

func TestTimeOfDayManaRegenSystem_SetLightingSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayManaRegenSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)

	sys.SetLightingSystem(lighting)
	if sys.lightingSystem != lighting {
		t.Error("lighting system not set")
	}
}

func TestTimeOfDayManaRegenSystem_Update_NoLightingSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayManaRegenSystem(world, 12345)

	entity := NewEntity()
	entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})
	world.AddEntity(entity)

	// Should not panic without lighting system
	sys.Update([]*Entity{entity}, 1.0)

	mana := entity.GetComponentByType(&ManaComponent{}).(*ManaComponent)
	if mana.Regen != 5.0 {
		t.Errorf("regen should not change without lighting system: got %f, want 5.0", mana.Regen)
	}
}

// stubManaRegenClock implements GameClock for testing
type stubManaRegenClock struct {
	hour int
}

func (c *stubManaRegenClock) Hour() int {
	return c.hour
}

func TestTimeOfDayManaRegenSystem_Update_WithLightingSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayManaRegenSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)
	sys.SetLightingSystem(lighting)
	sys.SetGenre("fantasy")

	entity := NewEntity()
	entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 10.0})
	world.AddEntity(entity)

	// Force time interval
	sys.timeSinceCheck = 2.0

	// Update system
	sys.Update([]*Entity{entity}, 0.1)

	mana := entity.GetComponentByType(&ManaComponent{}).(*ManaComponent)
	// Mana regen should be modified based on time of day
	if mana.Regen == 0 {
		t.Error("mana regen should be non-zero after update")
	}
}

func TestTimeOfDayManaRegenSystem_getManaRegenMultiplier(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayManaRegenSystem(world, 12345)

	tests := []struct {
		name      string
		genreID   string
		timeOfDay palette.TimeOfDay
		wantMin   float64
		wantMax   float64
	}{
		{"fantasy night", "fantasy", palette.TimeOfDayNight, 1.3, 1.5},
		{"fantasy day", "fantasy", palette.TimeOfDayDay, 0.9, 1.1},
		{"scifi day", "scifi", palette.TimeOfDayDay, 1.0, 1.2},
		{"scifi night", "scifi", palette.TimeOfDayNight, 1.0, 1.2},
		{"horror night", "horror", palette.TimeOfDayNight, 1.35, 1.55},
		{"cyberpunk night", "cyberpunk", palette.TimeOfDayNight, 1.25, 1.45},
		{"postapoc day", "postapoc", palette.TimeOfDayDay, 0.75, 0.95},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys.SetGenre(tt.genreID)
			mult := sys.getManaRegenMultiplier(tt.timeOfDay)
			if mult < tt.wantMin || mult > tt.wantMax {
				t.Errorf("multiplier %f outside expected range [%f, %f]", mult, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestTimeOfDayManaRegenSystem_GetActiveMultiplier(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayManaRegenSystem(world, 12345)

	// No multiplier set
	mult := sys.GetActiveMultiplier(999)
	if mult != 1.0 {
		t.Errorf("expected 1.0 for unknown entity, got %f", mult)
	}

	// Set multiplier manually
	sys.activeMultipliers[123] = 1.5
	mult = sys.GetActiveMultiplier(123)
	if mult != 1.5 {
		t.Errorf("expected 1.5, got %f", mult)
	}
}

func TestTimeOfDayManaRegenSystem_GetCurrentMultiplier(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayManaRegenSystem(world, 12345)

	// Without lighting system
	mult := sys.GetCurrentMultiplier()
	if mult != 1.0 {
		t.Errorf("expected 1.0 without lighting system, got %f", mult)
	}

	// With lighting system
	lighting := NewTimeOfDayLightingSystem(world, 12345)
	sys.SetLightingSystem(lighting)
	mult = sys.GetCurrentMultiplier()
	if mult < 0.5 || mult > 2.0 {
		t.Errorf("multiplier %f outside valid range [0.5, 2.0]", mult)
	}
}

func TestTimeOfDayManaRegenSystem_GetBonusDescription(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayManaRegenSystem(world, 12345)

	// Without lighting system
	desc := sys.GetBonusDescription()
	if desc != "" {
		t.Errorf("expected empty description without lighting, got %q", desc)
	}

	// With lighting system
	lighting := NewTimeOfDayLightingSystem(world, 12345)
	sys.SetLightingSystem(lighting)

	// Force fantasy night for bonus
	sys.SetGenre("fantasy")
	desc = sys.GetBonusDescription()
	// Description may or may not be empty depending on time
	t.Logf("bonus description: %q", desc)
}

func TestTimeOfDayManaRegenSystem_RestoreOriginalRegen(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayManaRegenSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)
	sys.SetLightingSystem(lighting)

	entity := NewEntity()
	originalRegen := 10.0
	entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: originalRegen})
	world.AddEntity(entity)

	// Store original and modify
	sys.originalRegen[entity.ID] = originalRegen
	sys.activeMultipliers[entity.ID] = 1.5

	// Modify mana regen
	mana := entity.GetComponentByType(&ManaComponent{}).(*ManaComponent)
	mana.Regen = 15.0

	// Restore
	sys.RestoreOriginalRegen(entity.ID)

	if mana.Regen != originalRegen {
		t.Errorf("regen not restored: got %f, want %f", mana.Regen, originalRegen)
	}
	if _, exists := sys.originalRegen[entity.ID]; exists {
		t.Error("original regen not cleaned up")
	}
	if _, exists := sys.activeMultipliers[entity.ID]; exists {
		t.Error("active multiplier not cleaned up")
	}
}

func TestTimeOfDayManaRegenSystem_MultiplierClamping(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayManaRegenSystem(world, 12345)

	// Test that multipliers are clamped to [0.5, 2.0]
	times := []palette.TimeOfDay{
		palette.TimeOfDayDawn,
		palette.TimeOfDayDay,
		palette.TimeOfDayDusk,
		palette.TimeOfDayNight,
	}
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		sys.SetGenre(genre)
		for _, tod := range times {
			mult := sys.getManaRegenMultiplier(tod)
			if mult < 0.5 {
				t.Errorf("genre=%s time=%s: multiplier %f below minimum 0.5", genre, tod, mult)
			}
			if mult > 2.0 {
				t.Errorf("genre=%s time=%s: multiplier %f above maximum 2.0", genre, tod, mult)
			}
		}
	}
}

func TestTimeOfDayManaRegenSystem_UpdateInterval(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayManaRegenSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)
	sys.SetLightingSystem(lighting)

	entity := NewEntity()
	entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 10.0})
	world.AddEntity(entity)

	// Small delta should not trigger update
	sys.Update([]*Entity{entity}, 0.1)
	if sys.timeSinceCheck != 0.1 {
		t.Errorf("timeSinceCheck not accumulated: got %f, want 0.1", sys.timeSinceCheck)
	}

	// Larger delta should trigger update
	sys.Update([]*Entity{entity}, 1.0)
	if sys.timeSinceCheck != 0.0 {
		t.Errorf("timeSinceCheck should reset after update interval: got %f", sys.timeSinceCheck)
	}
}

func TestTimeOfDayManaRegenSystem_EntityWithoutMana(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayManaRegenSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)
	sys.SetLightingSystem(lighting)

	// Entity without mana component
	entity := NewEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.AddEntity(entity)

	// Should not panic
	sys.timeSinceCheck = 2.0
	sys.Update([]*Entity{entity}, 0.1)
}

func TestManaRegenItoa(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{1, "1"},
		{10, "10"},
		{100, "100"},
		{-5, "-5"},
		{25, "25"},
	}

	for _, tt := range tests {
		got := manaRegenItoa(tt.n)
		if got != tt.want {
			t.Errorf("manaRegenItoa(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func BenchmarkTimeOfDayManaRegenSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewTimeOfDayManaRegenSystem(world, 12345)
	lighting := NewTimeOfDayLightingSystem(world, 12345)
	sys.SetLightingSystem(lighting)

	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := NewEntity()
		entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 10.0})
		world.AddEntity(entity)
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceCheck = 2.0 // Force update
		sys.Update(entities, 0.016)
	}
}

func BenchmarkTimeOfDayManaRegenSystem_GetMultiplier(b *testing.B) {
	world := NewWorld()
	sys := NewTimeOfDayManaRegenSystem(world, 12345)
	sys.SetGenre("fantasy")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sys.getManaRegenMultiplier(palette.TimeOfDayNight)
	}
}
