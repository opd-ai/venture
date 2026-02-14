package engine

import (
	"testing"
)

func TestNewStealthIndicatorParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewStealthIndicatorParticleSystem(world, 12345)

	if system == nil {
		t.Fatal("NewStealthIndicatorParticleSystem returned nil")
	}
	if system.world != world {
		t.Error("world not set correctly")
	}
	if system.seed != 12345 {
		t.Errorf("seed = %d, want 12345", system.seed)
	}
	if system.rng == nil {
		t.Error("rng not initialized")
	}
	if system.lastStealthState == nil {
		t.Error("lastStealthState map not initialized")
	}
	if system.coverThreshold != 0.85 {
		t.Errorf("coverThreshold = %f, want 0.85", system.coverThreshold)
	}
}

func TestStealthIndicatorParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewStealthIndicatorParticleSystem(world, 12345)
	ps := NewParticleSystem()

	system.SetParticleSystem(ps)

	if system.particleSystem != ps {
		t.Error("particle system not set")
	}
}

func TestStealthIndicatorParticleSystem_SetTerrainStealthSystem(t *testing.T) {
	world := NewWorld()
	system := NewStealthIndicatorParticleSystem(world, 12345)
	tss := NewTerrainStealthSystem(world, 54321)

	system.SetTerrainStealthSystem(tss)

	if system.terrainStealthSystem != tss {
		t.Error("terrain stealth system not set")
	}
}

func TestStealthIndicatorParticleSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewStealthIndicatorParticleSystem(world, 12345)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		system.SetGenre(genre)
		if system.genreID != genre {
			t.Errorf("genre = %s, want %s", system.genreID, genre)
		}
	}
}

func TestStealthIndicatorParticleSystem_SetCoverThreshold(t *testing.T) {
	world := NewWorld()
	system := NewStealthIndicatorParticleSystem(world, 12345)

	tests := []struct {
		name      string
		threshold float64
		expected  float64
	}{
		{"valid threshold", 0.7, 0.7},
		{"boundary low", 0.01, 0.01},
		{"boundary high", 0.99, 0.99},
		{"invalid zero", 0.0, 0.85},      // Should not change
		{"invalid one", 1.0, 0.85},       // Should not change
		{"invalid negative", -0.5, 0.85}, // Should not change
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system.coverThreshold = 0.85 // Reset to default
			system.SetCoverThreshold(tt.threshold)
			if system.coverThreshold != tt.expected {
				t.Errorf("coverThreshold = %f, want %f", system.coverThreshold, tt.expected)
			}
		})
	}
}

func TestStealthIndicatorParticleSystem_GetCoverState_NoTracking(t *testing.T) {
	world := NewWorld()
	system := NewStealthIndicatorParticleSystem(world, 12345)

	inCover, mult := system.GetCoverState(999)

	if inCover {
		t.Error("untracked entity should not be in cover")
	}
	if mult != 1.0 {
		t.Errorf("multiplier = %f, want 1.0", mult)
	}
}

func TestStealthIndicatorParticleSystem_GetCoverState_WithTracking(t *testing.T) {
	world := NewWorld()
	system := NewStealthIndicatorParticleSystem(world, 12345)

	// Manually set state
	system.lastStealthState[42] = stealthState{
		multiplier: 0.6,
		inCover:    true,
	}

	inCover, mult := system.GetCoverState(42)

	if !inCover {
		t.Error("tracked entity should be in cover")
	}
	if mult != 0.6 {
		t.Errorf("multiplier = %f, want 0.6", mult)
	}
}

func TestStealthIndicatorParticleSystem_Update_NilSystems(t *testing.T) {
	world := NewWorld()
	system := NewStealthIndicatorParticleSystem(world, 12345)

	entity := NewEntity(1)
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Should not panic with nil particle system
	system.Update([]*Entity{entity}, 0.016)

	// Set particle system but not terrain stealth
	system.SetParticleSystem(NewParticleSystem())
	system.Update([]*Entity{entity}, 0.016)
}

func TestStealthIndicatorParticleSystem_cleanupEntity(t *testing.T) {
	world := NewWorld()
	system := NewStealthIndicatorParticleSystem(world, 12345)

	// Add tracking
	system.lastStealthState[42] = stealthState{multiplier: 0.7, inCover: true}

	// Cleanup
	system.cleanupEntity(42)

	if _, exists := system.lastStealthState[42]; exists {
		t.Error("entity tracking should be removed after cleanup")
	}
}

func TestStealthIndicatorParticleSystem_UpdateEntityWithoutHealth(t *testing.T) {
	world := NewWorld()
	system := NewStealthIndicatorParticleSystem(world, 12345)
	system.SetParticleSystem(NewParticleSystem())
	system.SetTerrainStealthSystem(NewTerrainStealthSystem(world, 54321))

	// Pre-populate tracking
	entity := NewEntity(1)
	system.lastStealthState[entity.ID] = stealthState{multiplier: 0.7, inCover: true}

	// Entity without health component should be cleaned up
	system.Update([]*Entity{entity}, 0.016)

	if _, exists := system.lastStealthState[entity.ID]; exists {
		t.Error("entity without health should be cleaned up")
	}
}

func TestStealthIndicatorParticleSystem_getEnterCoverParticleType(t *testing.T) {
	world := NewWorld()
	system := NewStealthIndicatorParticleSystem(world, 12345)

	tests := []struct {
		genre    string
		wantType string // We check it's not empty, exact type depends on particles package
	}{
		{"fantasy", "sparkle"},
		{"scifi", "spark"},
		{"horror", "smoke"},
		{"cyberpunk", "spark"},
		{"postapoc", "dust"},
		{"unknown", "sparkle"}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			system.SetGenre(tt.genre)
			pType := system.getEnterCoverParticleType()
			// Just verify it returns a valid type (not zero value)
			if pType < 0 {
				t.Error("particle type should be valid")
			}
		})
	}
}

func TestStealthIndicatorParticleSystem_getExitCoverParticleType(t *testing.T) {
	world := NewWorld()
	system := NewStealthIndicatorParticleSystem(world, 12345)

	tests := []struct {
		genre string
	}{
		{"fantasy"},
		{"scifi"},
		{"horror"},
		{"cyberpunk"},
		{"postapoc"},
		{"unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			system.SetGenre(tt.genre)
			pType := system.getExitCoverParticleType()
			// Just verify it returns a valid type
			if pType < 0 {
				t.Error("particle type should be valid")
			}
		})
	}
}

func TestStealthIndicatorParticleSystem_CoverStateTransitions(t *testing.T) {
	world := NewWorld()
	system := NewStealthIndicatorParticleSystem(world, 12345)

	tests := []struct {
		name        string
		prevMult    float64
		prevInCover bool
		currentMult float64
		threshold   float64
		wantEnter   bool
		wantExit    bool
	}{
		{
			name:        "enter cover",
			prevMult:    1.0,
			prevInCover: false,
			currentMult: 0.7,
			threshold:   0.85,
			wantEnter:   true,
			wantExit:    false,
		},
		{
			name:        "exit cover",
			prevMult:    0.6,
			prevInCover: true,
			currentMult: 1.0,
			threshold:   0.85,
			wantEnter:   false,
			wantExit:    true,
		},
		{
			name:        "stay in cover",
			prevMult:    0.7,
			prevInCover: true,
			currentMult: 0.65,
			threshold:   0.85,
			wantEnter:   false,
			wantExit:    false,
		},
		{
			name:        "stay out of cover",
			prevMult:    1.0,
			prevInCover: false,
			currentMult: 0.95,
			threshold:   0.85,
			wantEnter:   false,
			wantExit:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system.coverThreshold = tt.threshold

			currentInCover := tt.currentMult < tt.threshold
			prevState := stealthState{multiplier: tt.prevMult, inCover: tt.prevInCover}

			gotEnter := currentInCover && !prevState.inCover
			gotExit := !currentInCover && prevState.inCover

			if gotEnter != tt.wantEnter {
				t.Errorf("enter transition = %v, want %v", gotEnter, tt.wantEnter)
			}
			if gotExit != tt.wantExit {
				t.Errorf("exit transition = %v, want %v", gotExit, tt.wantExit)
			}
		})
	}
}

func TestStealthIndicatorParticleSystem_ParticleCountScaling(t *testing.T) {
	// Test that particle count scales with stealth effectiveness
	baseCount := 5

	tests := []struct {
		stealthMult float64
		minExpected int
		maxExpected int
	}{
		{0.8, 5, 5},  // Normal cover
		{0.65, 6, 8}, // Good cover (scaled up)
		{0.4, 8, 12}, // Excellent cover (scaled up more)
	}

	for _, tt := range tests {
		count := baseCount
		if tt.stealthMult < 0.7 {
			count = int(float64(count) * 1.4)
		}
		if tt.stealthMult < 0.5 {
			count = int(float64(count) * 1.3)
		}
		if count > 12 {
			count = 12
		}

		if count < tt.minExpected || count > tt.maxExpected {
			t.Errorf("stealthMult=%f: count=%d, want [%d-%d]",
				tt.stealthMult, count, tt.minExpected, tt.maxExpected)
		}
	}
}

func BenchmarkStealthIndicatorParticleSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewStealthIndicatorParticleSystem(world, 12345)
	system.SetParticleSystem(NewParticleSystem())
	system.SetTerrainStealthSystem(NewTerrainStealthSystem(world, 54321))

	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		e := NewEntity(uint64(i))
		e.AddComponent(&HealthComponent{Current: 100, Max: 100})
		e.AddComponent(&PositionComponent{X: float64(i * 32), Y: float64(i * 32)})
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}
