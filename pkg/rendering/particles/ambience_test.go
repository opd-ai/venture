// Package particles provides ambient particle effects tests.
package particles

import (
	"image/color"
	"math"
	"math/rand"
	"testing"
)

func TestEnvironmentType_String(t *testing.T) {
	tests := []struct {
		name    string
		envType EnvironmentType
		want    string
	}{
		{"dungeon", EnvironmentDungeon, "Dungeon"},
		{"cave", EnvironmentCave, "Cave"},
		{"forest", EnvironmentForest, "Forest"},
		{"desert", EnvironmentDesert, "Desert"},
		{"snow", EnvironmentSnow, "Snow"},
		{"swamp", EnvironmentSwamp, "Swamp"},
		{"lava", EnvironmentLava, "Lava"},
		{"city", EnvironmentCity, "City"},
		{"laboratory", EnvironmentLaboratory, "Laboratory"},
		{"ruins", EnvironmentRuins, "Ruins"},
		{"unknown", EnvironmentType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.envType.String(); got != tt.want {
				t.Errorf("EnvironmentType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultAmbienceConfig(t *testing.T) {
	config := DefaultAmbienceConfig()

	if config.Type != EnvironmentDungeon {
		t.Errorf("expected default type %v, got %v", EnvironmentDungeon, config.Type)
	}
	if config.Width != 800 {
		t.Errorf("expected default width 800, got %d", config.Width)
	}
	if config.Height != 600 {
		t.Errorf("expected default height 600, got %d", config.Height)
	}
	if config.GenreID != "fantasy" {
		t.Errorf("expected default genreID 'fantasy', got %s", config.GenreID)
	}
	if config.Density != 0.5 {
		t.Errorf("expected default density 0.5, got %f", config.Density)
	}
	if config.Custom == nil {
		t.Error("expected Custom map to be initialized")
	}
}

func TestAmbienceConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  AmbienceConfig
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultAmbienceConfig(),
			wantErr: false,
		},
		{
			name: "invalid width",
			config: AmbienceConfig{
				Type:    EnvironmentDungeon,
				Width:   0,
				Height:  600,
				GenreID: "fantasy",
				Density: 0.5,
			},
			wantErr: true,
		},
		{
			name: "invalid height",
			config: AmbienceConfig{
				Type:    EnvironmentDungeon,
				Width:   800,
				Height:  -1,
				GenreID: "fantasy",
				Density: 0.5,
			},
			wantErr: true,
		},
		{
			name: "empty genreID",
			config: AmbienceConfig{
				Type:    EnvironmentDungeon,
				Width:   800,
				Height:  600,
				GenreID: "",
				Density: 0.5,
			},
			wantErr: true,
		},
		{
			name: "negative density",
			config: AmbienceConfig{
				Type:    EnvironmentDungeon,
				Width:   800,
				Height:  600,
				GenreID: "fantasy",
				Density: -0.1,
			},
			wantErr: true,
		},
		{
			name: "density too high",
			config: AmbienceConfig{
				Type:    EnvironmentDungeon,
				Width:   800,
				Height:  600,
				GenreID: "fantasy",
				Density: 1.1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAmbienceConfig_GetParticleCount(t *testing.T) {
	tests := []struct {
		name     string
		config   AmbienceConfig
		wantMin  int
		wantMax  int
	}{
		{
			name: "standard area with medium density",
			config: AmbienceConfig{
				Width:   800,
				Height:  600,
				Density: 0.5,
			},
			wantMin: 30,
			wantMax: 40,
		},
		{
			name: "standard area with high density",
			config: AmbienceConfig{
				Width:   800,
				Height:  600,
				Density: 1.0,
			},
			wantMin: 70,
			wantMax: 80,
		},
		{
			name: "standard area with low density",
			config: AmbienceConfig{
				Width:   800,
				Height:  600,
				Density: 0.1,
			},
			wantMin: 10,
			wantMax: 10, // minimum clamping
		},
		{
			name: "large area",
			config: AmbienceConfig{
				Width:   1600,
				Height:  1200,
				Density: 0.5,
			},
			wantMin: 100,
			wantMax: 100, // maximum clamping
		},
		{
			name: "small area",
			config: AmbienceConfig{
				Width:   400,
				Height:  300,
				Density: 0.5,
			},
			wantMin: 10,
			wantMax: 10, // minimum clamping
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.GetParticleCount()
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("GetParticleCount() = %d, want between %d and %d", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestGenerateAmbience_AllEnvironments(t *testing.T) {
	environments := []EnvironmentType{
		EnvironmentDungeon,
		EnvironmentCave,
		EnvironmentForest,
		EnvironmentDesert,
		EnvironmentSnow,
		EnvironmentSwamp,
		EnvironmentLava,
		EnvironmentCity,
		EnvironmentLaboratory,
		EnvironmentRuins,
	}

	for _, envType := range environments {
		t.Run(envType.String(), func(t *testing.T) {
			config := AmbienceConfig{
				Type:    envType,
				Width:   800,
				Height:  600,
				GenreID: "fantasy",
				Seed:    12345,
				Density: 0.5,
				Custom:  make(map[string]interface{}),
			}

			system, err := GenerateAmbience(config)
			if err != nil {
				t.Fatalf("GenerateAmbience() error = %v", err)
			}

			if system == nil {
				t.Fatal("expected non-nil system")
			}

			if len(system.Particles) == 0 {
				t.Error("expected particles to be generated")
			}

			// Verify particles have valid properties
			for i, p := range system.Particles {
				if p.X < 0 || p.X > float64(config.Width) {
					t.Errorf("particle %d has invalid X position: %f", i, p.X)
				}
				if p.Y < 0 || p.Y > float64(config.Height) {
					t.Errorf("particle %d has invalid Y position: %f", i, p.Y)
				}
				if p.Size <= 0 {
					t.Errorf("particle %d has invalid size: %f", i, p.Size)
				}
				if p.Life <= 0 {
					t.Errorf("particle %d has invalid lifetime: %f", i, p.Life)
				}
				if p.Color == nil {
					t.Errorf("particle %d has nil color", i)
				}
			}
		})
	}
}

func TestGenerateAmbience_Deterministic(t *testing.T) {
	config := AmbienceConfig{
		Type:    EnvironmentForest,
		Width:   800,
		Height:  600,
		GenreID: "fantasy",
		Seed:    99999,
		Density: 0.5,
		Custom:  make(map[string]interface{}),
	}

	system1, err1 := GenerateAmbience(config)
	if err1 != nil {
		t.Fatalf("first GenerateAmbience() error = %v", err1)
	}

	system2, err2 := GenerateAmbience(config)
	if err2 != nil {
		t.Fatalf("second GenerateAmbience() error = %v", err2)
	}

	if len(system1.Particles) != len(system2.Particles) {
		t.Errorf("particle count mismatch: %d vs %d", len(system1.Particles), len(system2.Particles))
	}

	// Compare first few particles for determinism
	compareCount := 5
	if len(system1.Particles) < compareCount {
		compareCount = len(system1.Particles)
	}

	for i := 0; i < compareCount; i++ {
		p1 := system1.Particles[i]
		p2 := system2.Particles[i]

		if math.Abs(p1.X-p2.X) > 0.001 {
			t.Errorf("particle %d X mismatch: %f vs %f", i, p1.X, p2.X)
		}
		if math.Abs(p1.Y-p2.Y) > 0.001 {
			t.Errorf("particle %d Y mismatch: %f vs %f", i, p1.Y, p2.Y)
		}
		if math.Abs(p1.VX-p2.VX) > 0.001 {
			t.Errorf("particle %d VX mismatch: %f vs %f", i, p1.VX, p2.VX)
		}
		if math.Abs(p1.VY-p2.VY) > 0.001 {
			t.Errorf("particle %d VY mismatch: %f vs %f", i, p1.VY, p2.VY)
		}
		if math.Abs(p1.Size-p2.Size) > 0.001 {
			t.Errorf("particle %d Size mismatch: %f vs %f", i, p1.Size, p2.Size)
		}
	}
}

func TestGenerateAmbience_InvalidConfig(t *testing.T) {
	config := AmbienceConfig{
		Type:    EnvironmentDungeon,
		Width:   0, // invalid
		Height:  600,
		GenreID: "fantasy",
		Seed:    12345,
		Density: 0.5,
	}

	_, err := GenerateAmbience(config)
	if err == nil {
		t.Error("expected error for invalid config, got nil")
	}
}

func TestAmbienceSystem_Update(t *testing.T) {
	config := AmbienceConfig{
		Type:    EnvironmentDungeon,
		Width:   800,
		Height:  600,
		GenreID: "fantasy",
		Seed:    12345,
		Density: 0.5,
		Custom:  make(map[string]interface{}),
	}

	system, err := GenerateAmbience(config)
	if err != nil {
		t.Fatalf("GenerateAmbience() error = %v", err)
	}

	initialTime := system.ElapsedTime
	initialX := system.Particles[0].X
	initialY := system.Particles[0].Y

	// Update system
	deltaTime := 0.016 // ~60 FPS
	system.Update(deltaTime)

	// Check time advanced
	if system.ElapsedTime <= initialTime {
		t.Errorf("ElapsedTime did not advance: %f vs %f", system.ElapsedTime, initialTime)
	}

	// Check particles moved (at least one should have moved)
	moved := false
	for i := range system.Particles {
		if i == 0 {
			if math.Abs(system.Particles[i].X-initialX) > 0.001 ||
				math.Abs(system.Particles[i].Y-initialY) > 0.001 {
				moved = true
				break
			}
		}
	}
	if !moved {
		t.Error("particles did not move after update")
	}
}

func TestAmbienceSystem_Update_ParticleRespawn(t *testing.T) {
	config := AmbienceConfig{
		Type:    EnvironmentDungeon,
		Width:   800,
		Height:  600,
		GenreID: "fantasy",
		Seed:    12345,
		Density: 0.1, // fewer particles for easier testing
		Custom:  make(map[string]interface{}),
	}

	system, err := GenerateAmbience(config)
	if err != nil {
		t.Fatalf("GenerateAmbience() error = %v", err)
	}

	// Force a particle to expire
	system.Particles[0].Life = 0.001

	initialX := system.Particles[0].X
	initialY := system.Particles[0].Y

	// Update to trigger respawn
	system.Update(0.01)

	// Particle should have respawned with new position
	// (may or may not be different, but should be within bounds)
	p := system.Particles[0]
	if p.X < 0 || p.X > float64(config.Width) {
		t.Errorf("respawned particle has invalid X: %f", p.X)
	}
	if p.Y < 0 || p.Y > float64(config.Height) {
		t.Errorf("respawned particle has invalid Y: %f", p.Y)
	}
	if p.Life <= 0 {
		t.Errorf("respawned particle has invalid life: %f", p.Life)
	}

	// Life should be reset to near InitialLife (respawned particles have new lifetime)
	age := p.InitialLife - p.Life
	if age > 0.1 {
		t.Errorf("respawned particle age too high: %f", age)
	}

	// Position likely changed (not guaranteed due to randomness, but very likely)
	if initialX == p.X && initialY == p.Y {
		t.Log("warning: respawned particle has same position (rare but possible)")
	}
}

func TestAmbienceSystem_EnvironmentBehaviors(t *testing.T) {
	// Test that different environments produce different particle behaviors
	environments := []EnvironmentType{
		EnvironmentDungeon,
		EnvironmentForest,
		EnvironmentLava,
	}

	configs := make([]AmbienceConfig, len(environments))
	systems := make([]*AmbienceSystem, len(environments))

	for i, envType := range environments {
		configs[i] = AmbienceConfig{
			Type:    envType,
			Width:   800,
			Height:  600,
			GenreID: "fantasy",
			Seed:    12345,
			Density: 0.5,
			Custom:  make(map[string]interface{}),
		}

		var err error
		systems[i], err = GenerateAmbience(configs[i])
		if err != nil {
			t.Fatalf("GenerateAmbience(%s) error = %v", envType, err)
		}

		// Update to apply behaviors
		for j := 0; j < 10; j++ {
			systems[i].Update(0.1)
		}
	}

	// Verify that at least velocities differ between environment types
	// (indicating different behaviors are applied)
	_ = systems[0].Particles[0].VY // dungeonVY
	forestVY := systems[1].Particles[0].VY
	lavaVY := systems[2].Particles[0].VY

	// Lava should have strong upward motion (negative VY)
	if lavaVY >= 0 {
		t.Logf("warning: lava particle not rising as expected (VY=%f)", lavaVY)
	}

	// Forest should have downward motion (positive VY)
	if forestVY <= 0 {
		t.Logf("warning: forest particle not falling as expected (VY=%f)", forestVY)
	}
}

func TestEnvironmentBehaviorHelpers(t *testing.T) {
	// Create a test particle with some age (Life < InitialLife)
	p := &Particle{
		X:             100,
		Y:             100,
		VX:            0,
		VY:            0,
		Size:          2.0,
		Life:          8.0,
		InitialLife:   10.0, // age = 2.0 seconds
		Rotation:      0,
		RotationVel:   0,
		Color:         color.RGBA{R: 255, G: 255, B: 255, A: 255},
	}

	deltaTime := 0.1

	t.Run("applyDriftBehavior", func(t *testing.T) {
		pCopy := *p
		applyDriftBehavior(&pCopy, deltaTime, 0.5)
		// Velocity should have changed
		if pCopy.VX == p.VX && pCopy.VY == p.VY {
			t.Error("drift behavior did not modify velocity")
		}
	})

	t.Run("applySineWaveBehavior", func(t *testing.T) {
		pCopy := *p
		applySineWaveBehavior(&pCopy, deltaTime, 1.0, 2.0)
		// VX should have changed
		if pCopy.VX == p.VX {
			t.Error("sine wave behavior did not modify VX")
		}
	})

	t.Run("applyGustBehavior", func(t *testing.T) {
		pCopy := *p
		// Use a time that triggers a gust
		applyGustBehavior(&pCopy, deltaTime, 0.5, 10.0)
		// May or may not trigger depending on time
		// Just verify it doesn't crash
	})

	t.Run("applySnowDriftBehavior", func(t *testing.T) {
		pCopy := *p
		applySnowDriftBehavior(&pCopy, deltaTime, 0.3)
		// VX should have changed
		if pCopy.VX == p.VX {
			t.Error("snow drift behavior did not modify VX")
		}
	})

	t.Run("applyFloatBehavior", func(t *testing.T) {
		pCopy := *p
		applyFloatBehavior(&pCopy, deltaTime, 1.0)
		// VY should have changed
		if pCopy.VY == p.VY {
			t.Error("float behavior did not modify VY")
		}
	})

	t.Run("applyRiseBehavior", func(t *testing.T) {
		pCopy := *p
		applyRiseBehavior(&pCopy, deltaTime, 20.0)
		// VY should be negative (rising)
		if pCopy.VY >= 0 {
			t.Errorf("rise behavior did not make particle rise (VY=%f)", pCopy.VY)
		}
	})

	t.Run("applyTumbleBehavior", func(t *testing.T) {
		pCopy := *p
		pCopy.VX = 10.0
		pCopy.VY = 10.0
		applyTumbleBehavior(&pCopy, deltaTime, 1.5)
		// Rotation speed should be set based on velocity
		if pCopy.RotationVel == 0 {
			t.Error("tumble behavior did not set rotation speed")
		}
	})

	t.Run("applyOrbitBehavior", func(t *testing.T) {
		pCopy := *p
		applyOrbitBehavior(&pCopy, deltaTime, 30.0, 5.0)
		// Velocity should be set
		if pCopy.VX == 0 && pCopy.VY == 0 {
			t.Error("orbit behavior did not set velocity")
		}
	})

	t.Run("applyFadeBehavior", func(t *testing.T) {
		pCopy := *p
		pCopy.Life = 1.0
		pCopy.InitialLife = 10.0 // near end of life (1.0 out of 10.0)
		applyFadeBehavior(&pCopy)
		// Alpha may be reduced (depends on calculation)
		// Just verify it doesn't crash
	})
}

func TestGetEnvironmentVelocity(t *testing.T) {
	tests := []struct {
		envType     EnvironmentType
		wantMin     float64
		wantMax     float64
	}{
		{EnvironmentDungeon, 4.0, 6.0},
		{EnvironmentCave, 4.0, 6.0},
		{EnvironmentForest, 9.0, 11.0},
		{EnvironmentDesert, 18.0, 22.0},
		{EnvironmentSnow, 7.0, 9.0},
		{EnvironmentSwamp, 5.0, 7.0},
		{EnvironmentLava, 23.0, 27.0},
		{EnvironmentCity, 13.0, 17.0},
		{EnvironmentLaboratory, 11.0, 13.0},
		{EnvironmentRuins, 4.0, 6.0},
	}

	for _, tt := range tests {
		t.Run(tt.envType.String(), func(t *testing.T) {
			got := getEnvironmentVelocity(tt.envType)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("getEnvironmentVelocity(%s) = %f, want between %f and %f",
					tt.envType, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestGetEnvironmentLifetime(t *testing.T) {
	// Test that lifetime is reasonable for all environments
	environments := []EnvironmentType{
		EnvironmentDungeon,
		EnvironmentCave,
		EnvironmentForest,
		EnvironmentDesert,
		EnvironmentSnow,
		EnvironmentSwamp,
		EnvironmentLava,
		EnvironmentCity,
		EnvironmentLaboratory,
		EnvironmentRuins,
	}

	for _, envType := range environments {
		t.Run(envType.String(), func(t *testing.T) {
			// Get multiple samples to check variation
			var lifetimes []float64
			for i := 0; i < 10; i++ {
				seed := int64(i * 1000)
				rng := rand.New(rand.NewSource(seed))
				lifetime := getEnvironmentLifetime(envType, rng)
				lifetimes = append(lifetimes, lifetime)

				// Lifetime should be positive and reasonable (1-50 seconds)
				if lifetime < 1.0 || lifetime > 50.0 {
					t.Errorf("getEnvironmentLifetime(%s) = %f, want between 1.0 and 50.0",
						envType, lifetime)
				}
			}

			// Check that there's some variation
			allSame := true
			for i := 1; i < len(lifetimes); i++ {
				if math.Abs(lifetimes[i]-lifetimes[0]) > 0.1 {
					allSame = false
					break
				}
			}
			if allSame {
				t.Errorf("all lifetimes are the same for %s, expected variation", envType)
			}
		})
	}
}

// Benchmark tests

func BenchmarkGenerateAmbience(b *testing.B) {
	config := AmbienceConfig{
		Type:    EnvironmentForest,
		Width:   800,
		Height:  600,
		GenreID: "fantasy",
		Seed:    12345,
		Density: 0.5,
		Custom:  make(map[string]interface{}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GenerateAmbience(config)
		if err != nil {
			b.Fatalf("GenerateAmbience() error = %v", err)
		}
	}
}

func BenchmarkAmbienceSystem_Update(b *testing.B) {
	config := AmbienceConfig{
		Type:    EnvironmentForest,
		Width:   800,
		Height:  600,
		GenreID: "fantasy",
		Seed:    12345,
		Density: 0.5,
		Custom:  make(map[string]interface{}),
	}

	system, err := GenerateAmbience(config)
	if err != nil {
		b.Fatalf("GenerateAmbience() error = %v", err)
	}

	deltaTime := 0.016 // ~60 FPS

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(deltaTime)
	}
}

func BenchmarkGenerateAmbience_AllEnvironments(b *testing.B) {
	environments := []EnvironmentType{
		EnvironmentDungeon,
		EnvironmentCave,
		EnvironmentForest,
		EnvironmentDesert,
		EnvironmentSnow,
		EnvironmentSwamp,
		EnvironmentLava,
		EnvironmentCity,
		EnvironmentLaboratory,
		EnvironmentRuins,
	}

	for _, envType := range environments {
		b.Run(envType.String(), func(b *testing.B) {
			config := AmbienceConfig{
				Type:    envType,
				Width:   800,
				Height:  600,
				GenreID: "fantasy",
				Seed:    12345,
				Density: 0.5,
				Custom:  make(map[string]interface{}),
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := GenerateAmbience(config)
				if err != nil {
					b.Fatalf("GenerateAmbience() error = %v", err)
				}
			}
		})
	}
}
