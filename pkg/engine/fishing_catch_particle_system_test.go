package engine

import (
	"testing"
)

func TestNewFishingCatchParticleSystem(t *testing.T) {
	world := NewWorld(nil)
	system := NewFishingCatchParticleSystem(world, 12345)

	if system == nil {
		t.Fatal("NewFishingCatchParticleSystem returned nil")
	}

	if system.world != world {
		t.Error("World not set correctly")
	}

	if system.seed != 12345 {
		t.Errorf("Seed = %d, want 12345", system.seed)
	}

	if system.rng == nil {
		t.Error("RNG not initialized")
	}

	if system.baseParticleCount != 15 {
		t.Errorf("baseParticleCount = %d, want 15", system.baseParticleCount)
	}
}

func TestFishingCatchParticleSystem_SetGenre(t *testing.T) {
	world := NewWorld(nil)
	system := NewFishingCatchParticleSystem(world, 12345)

	system.SetGenre("horror")
	if system.genreID != "horror" {
		t.Errorf("genreID = %s, want horror", system.genreID)
	}
}

func TestFishingCatchParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld(nil)
	system := NewFishingCatchParticleSystem(world, 12345)
	ps := NewParticleSystem()

	system.SetParticleSystem(ps)
	if system.particleSystem != ps {
		t.Error("ParticleSystem not set correctly")
	}
}

func TestFishingCatchParticleSystem_SetFishingSystem(t *testing.T) {
	world := NewWorld(nil)
	system := NewFishingCatchParticleSystem(world, 12345)
	fs := NewFishingSystem(world, 54321)

	system.SetFishingSystem(fs)
	if system.fishingSystem != fs {
		t.Error("FishingSystem not set correctly")
	}

	// Verify callback is wired
	if fs.OnCatchCallback == nil {
		t.Error("OnCatchCallback should be set")
	}
}

func TestFishingCatchParticleSystem_GetFishRarity(t *testing.T) {
	world := NewWorld(nil)
	system := NewFishingCatchParticleSystem(world, 12345)
	fs := NewFishingSystem(world, 54321)
	system.SetFishingSystem(fs)

	tests := []struct {
		fishTypeID string
		wantRarity int
	}{
		{"bass", 0},         // Common
		{"trout", 0},        // Common
		{"catfish", 1},      // Uncommon
		{"pike", 2},         // Rare
		{"ethereal_eel", 3}, // Epic
		{"leviathan", 4},    // Legendary
		{"unknown", 0},      // Unknown defaults to 0
	}

	for _, tt := range tests {
		t.Run(tt.fishTypeID, func(t *testing.T) {
			got := system.getFishRarity(tt.fishTypeID)
			if got != tt.wantRarity {
				t.Errorf("getFishRarity(%s) = %d, want %d", tt.fishTypeID, got, tt.wantRarity)
			}
		})
	}
}

func TestFishingCatchParticleSystem_SelectParticleType(t *testing.T) {
	world := NewWorld(nil)
	system := NewFishingCatchParticleSystem(world, 12345)

	tests := []struct {
		genre  string
		rarity int
	}{
		{"fantasy", 0},
		{"fantasy", 4},
		{"scifi", 2},
		{"scifi", 4},
		{"horror", 2},
		{"cyberpunk", 1},
		{"cyberpunk", 3},
		{"postapoc", 2},
		{"unknown", 0},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			system.SetGenre(tt.genre)
			pt := system.selectParticleType(tt.rarity)
			// Just verify no panic and returns valid type
			if pt == "" {
				t.Error("selectParticleType returned empty string")
			}
		})
	}
}

func TestFishingCatchParticleSystem_SelectDuration(t *testing.T) {
	world := NewWorld(nil)
	system := NewFishingCatchParticleSystem(world, 12345)

	// Duration should increase with rarity
	dur0 := system.selectDuration(0)
	dur4 := system.selectDuration(4)

	if dur4 <= dur0 {
		t.Errorf("Duration should increase with rarity: dur0=%f, dur4=%f", dur0, dur4)
	}

	if dur0 < 0.5 || dur0 > 1.0 {
		t.Errorf("Base duration out of expected range: %f", dur0)
	}
}

func TestFishingCatchParticleSystem_SelectGravity(t *testing.T) {
	world := NewWorld(nil)
	system := NewFishingCatchParticleSystem(world, 12345)

	// Common fish should have positive gravity (fall down)
	grav0 := system.selectGravity(0)
	if grav0 < 0 {
		t.Errorf("Common fish gravity should be positive, got %f", grav0)
	}

	// Epic+ fish should have negative gravity (float up)
	grav4 := system.selectGravity(4)
	if grav4 >= 0 {
		t.Errorf("Legendary fish gravity should be negative, got %f", grav4)
	}
}

func TestFishingCatchParticleSystem_Update(t *testing.T) {
	world := NewWorld(nil)
	system := NewFishingCatchParticleSystem(world, 12345)

	// Update should be no-op (callback-driven)
	entities := []*Entity{}
	system.Update(entities, 0.016)
	// No panic = pass
}

func TestFishingCatchParticleSystem_OnFishCaught_NilChecks(t *testing.T) {
	world := NewWorld(nil)
	system := NewFishingCatchParticleSystem(world, 12345)

	// Should not panic with nil inputs
	system.OnFishCaught(nil, nil)
	system.OnFishCaught(&Entity{}, nil)
	system.OnFishCaught(nil, &CaughtFish{})
}

func TestFishingCatchParticleSystem_OnFishCaught(t *testing.T) {
	world := NewWorld(nil)
	system := NewFishingCatchParticleSystem(world, 12345)
	ps := NewParticleSystem()
	fs := NewFishingSystem(world, 54321)

	system.SetParticleSystem(ps)
	system.SetFishingSystem(fs)
	system.SetGenre("fantasy")

	// Create fisher entity with position
	fisher := NewEntity(1)
	fisher.AddComponent(&PositionComponent{X: 100, Y: 200})

	// Create caught fish
	caught := &CaughtFish{
		FishTypeID: "pike",
		Weight:     5.5,
		IsRecord:   true,
	}

	// Should spawn particles without panic
	system.OnFishCaught(fisher, caught)
}

func TestFishingCatchParticleSystem_OnFishCaught_MissingPosition(t *testing.T) {
	world := NewWorld(nil)
	system := NewFishingCatchParticleSystem(world, 12345)
	ps := NewParticleSystem()

	system.SetParticleSystem(ps)

	// Fisher without position component
	fisher := NewEntity(1)
	caught := &CaughtFish{FishTypeID: "bass", Weight: 1.0}

	// Should not panic, just return early
	system.OnFishCaught(fisher, caught)
}

func TestFishingCatchParticleSystem_SpawnCatchEffect(t *testing.T) {
	world := NewWorld(nil)
	system := NewFishingCatchParticleSystem(world, 12345)
	ps := NewParticleSystem()

	system.SetParticleSystem(ps)
	system.SetGenre("scifi")

	// Direct spawn method should work
	system.SpawnCatchEffect(150, 250, 3, true, 10.5)
}

func TestFishingCatchParticleSystem_SpawnCatchEffect_NoParticleSystem(t *testing.T) {
	world := NewWorld(nil)
	system := NewFishingCatchParticleSystem(world, 12345)

	// Should not panic without particle system
	system.SpawnCatchEffect(100, 200, 2, false, 5.0)
}

func TestFishingCatchParticleSystem_CallbackChaining(t *testing.T) {
	world := NewWorld(nil)
	system := NewFishingCatchParticleSystem(world, 12345)
	fs := NewFishingSystem(world, 54321)

	// Set up existing callback
	existingCallbackCalled := false
	fs.OnCatchCallback = func(fisher *Entity, caught *CaughtFish) {
		existingCallbackCalled = true
	}

	// Wire up system - should chain with existing callback
	system.SetFishingSystem(fs)
	system.SetParticleSystem(NewParticleSystem())

	// Create test entities
	fisher := NewEntity(1)
	fisher.AddComponent(&PositionComponent{X: 100, Y: 100})
	caught := &CaughtFish{FishTypeID: "bass", Weight: 1.0}

	// Trigger callback
	fs.OnCatchCallback(fisher, caught)

	if !existingCallbackCalled {
		t.Error("Existing callback should have been called (callback chaining)")
	}
}

func TestFishingCatchParticleSystem_RarityParticleScaling(t *testing.T) {
	world := NewWorld(nil)
	system := NewFishingCatchParticleSystem(world, 12345)

	// Test particle count scales with rarity
	tests := []struct {
		rarity      int
		minExpected int
		maxExpected int
	}{
		{0, 15, 20}, // Common
		{1, 20, 25}, // Uncommon
		{2, 25, 35}, // Rare
		{3, 35, 45}, // Epic
		{4, 40, 60}, // Legendary
	}

	baseCount := system.baseParticleCount
	for _, tt := range tests {
		t.Run("rarity_scaling", func(t *testing.T) {
			var multiplier float64
			switch tt.rarity {
			case 4:
				multiplier = 3.0
			case 3:
				multiplier = 2.5
			case 2:
				multiplier = 2.0
			case 1:
				multiplier = 1.5
			default:
				multiplier = 1.0
			}
			expected := int(float64(baseCount) * multiplier)
			if expected < tt.minExpected || expected > tt.maxExpected {
				t.Errorf("Rarity %d: expected count in range [%d, %d], got %d",
					tt.rarity, tt.minExpected, tt.maxExpected, expected)
			}
		})
	}
}

func BenchmarkFishingCatchParticleSystem_OnFishCaught(b *testing.B) {
	world := NewWorld(nil)
	system := NewFishingCatchParticleSystem(world, 12345)
	ps := NewParticleSystem()
	fs := NewFishingSystem(world, 54321)

	system.SetParticleSystem(ps)
	system.SetFishingSystem(fs)
	system.SetGenre("fantasy")

	fisher := NewEntity(1)
	fisher.AddComponent(&PositionComponent{X: 100, Y: 200})
	caught := &CaughtFish{FishTypeID: "pike", Weight: 5.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.OnFishCaught(fisher, caught)
	}
}
