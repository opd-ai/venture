package engine

import (
	"image/color"
	"testing"
)

func TestTrailComponent_Type(t *testing.T) {
	trail := NewTrailComponent()
	if trail.Type() != "trail" {
		t.Errorf("expected type 'trail', got '%s'", trail.Type())
	}
}

func TestNewTrailComponent(t *testing.T) {
	trail := NewTrailComponent()

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"enabled", trail.Enabled, true},
		{"spawn rate", trail.SpawnRate, 30.0},
		{"time since last spawn", trail.TimeSinceLastSpawn, 0.0},
		{"particle lifetime", trail.ParticleLifetime, 0.5},
		{"particle size", trail.ParticleSize, 2.0},
		{"fade rate", trail.FadeRate, 0.8},
		{"spread x", trail.SpreadX, 2.0},
		{"spread y", trail.SpreadY, 2.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: got %v, expected %v", tt.name, tt.got, tt.expected)
			}
		})
	}

	if trail.Color != nil {
		t.Errorf("expected nil color, got %v", trail.Color)
	}
}

func TestNewMagicTrailComponent(t *testing.T) {
	magicColor := &color.RGBA{R: 255, G: 0, B: 255, A: 255}
	trail := NewMagicTrailComponent(magicColor)

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"enabled", trail.Enabled, true},
		{"spawn rate", trail.SpawnRate, 40.0},
		{"time since last spawn", trail.TimeSinceLastSpawn, 0.0},
		{"particle lifetime", trail.ParticleLifetime, 0.8},
		{"particle size", trail.ParticleSize, 3.0},
		{"fade rate", trail.FadeRate, 0.6},
		{"spread x", trail.SpreadX, 4.0},
		{"spread y", trail.SpreadY, 4.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: got %v, expected %v", tt.name, tt.got, tt.expected)
			}
		})
	}

	if trail.Color == nil {
		t.Error("expected non-nil color for magic trail")
	} else if *trail.Color != *magicColor {
		t.Errorf("expected color %v, got %v", magicColor, trail.Color)
	}
}

func TestNewPhysicalTrailComponent(t *testing.T) {
	trail := NewPhysicalTrailComponent()

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"enabled", trail.Enabled, true},
		{"spawn rate", trail.SpawnRate, 20.0},
		{"time since last spawn", trail.TimeSinceLastSpawn, 0.0},
		{"particle lifetime", trail.ParticleLifetime, 0.3},
		{"particle size", trail.ParticleSize, 1.5},
		{"fade rate", trail.FadeRate, 0.9},
		{"spread x", trail.SpreadX, 1.0},
		{"spread y", trail.SpreadY, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: got %v, expected %v", tt.name, tt.got, tt.expected)
			}
		})
	}

	if trail.Color != nil {
		t.Errorf("expected nil color, got %v", trail.Color)
	}
}

func TestTrailComponent_Enabled(t *testing.T) {
	trail := NewTrailComponent()

	// Test enabling/disabling
	trail.Enabled = false
	if trail.Enabled {
		t.Error("expected trail to be disabled")
	}

	trail.Enabled = true
	if !trail.Enabled {
		t.Error("expected trail to be enabled")
	}
}

func TestTrailComponent_CustomColor(t *testing.T) {
	trail := NewTrailComponent()

	// Test with custom color
	customColor := &color.RGBA{R: 128, G: 64, B: 200, A: 255}
	trail.Color = customColor

	if trail.Color == nil {
		t.Fatal("expected non-nil color")
	}

	if *trail.Color != *customColor {
		t.Errorf("expected color %v, got %v", customColor, trail.Color)
	}
}

func TestTrailComponent_SpawnTiming(t *testing.T) {
	trail := NewTrailComponent()

	// Simulate time passage
	deltaTime := 1.0 / 30.0 // ~33ms, standard frame time
	trail.TimeSinceLastSpawn += deltaTime

	expectedTime := 0.033333333333333333
	tolerance := 0.0001

	if diff := trail.TimeSinceLastSpawn - expectedTime; diff < -tolerance || diff > tolerance {
		t.Errorf("expected time since last spawn ~%f, got %f", expectedTime, trail.TimeSinceLastSpawn)
	}
}

func TestTrailComponent_SpawnRateCalculation(t *testing.T) {
	tests := []struct {
		name                string
		spawnRate           float64
		deltaTime           float64
		expectedShouldSpawn bool
	}{
		{
			name:                "high spawn rate, normal frame",
			spawnRate:           60.0,
			deltaTime:           1.0 / 60.0,
			expectedShouldSpawn: true,
		},
		{
			name:                "low spawn rate, normal frame",
			spawnRate:           10.0,
			deltaTime:           1.0 / 60.0,
			expectedShouldSpawn: false,
		},
		{
			name:                "normal spawn rate, long frame",
			spawnRate:           30.0,
			deltaTime:           1.0 / 15.0, // Slow frame
			expectedShouldSpawn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trail := NewTrailComponent()
			trail.SpawnRate = tt.spawnRate

			trail.TimeSinceLastSpawn += tt.deltaTime
			spawnInterval := 1.0 / trail.SpawnRate
			shouldSpawn := trail.TimeSinceLastSpawn >= spawnInterval

			if shouldSpawn != tt.expectedShouldSpawn {
				t.Errorf("expected shouldSpawn=%v, got %v (interval=%f, time=%f)",
					tt.expectedShouldSpawn, shouldSpawn, spawnInterval, trail.TimeSinceLastSpawn)
			}
		})
	}
}

func TestTrailComponent_AllPresets(t *testing.T) {
	presets := []struct {
		name      string
		component *TrailComponent
	}{
		{"default", NewTrailComponent()},
		{"magic", NewMagicTrailComponent(&color.RGBA{R: 255, G: 0, B: 255, A: 255})},
		{"physical", NewPhysicalTrailComponent()},
	}

	for _, preset := range presets {
		t.Run(preset.name, func(t *testing.T) {
			trail := preset.component

			// All presets should be enabled by default
			if !trail.Enabled {
				t.Errorf("%s preset: expected enabled=true", preset.name)
			}

			// All presets should have positive spawn rate
			if trail.SpawnRate <= 0 {
				t.Errorf("%s preset: spawn rate must be positive, got %f", preset.name, trail.SpawnRate)
			}

			// All presets should have positive lifetime
			if trail.ParticleLifetime <= 0 {
				t.Errorf("%s preset: particle lifetime must be positive, got %f", preset.name, trail.ParticleLifetime)
			}

			// All presets should have positive size
			if trail.ParticleSize <= 0 {
				t.Errorf("%s preset: particle size must be positive, got %f", preset.name, trail.ParticleSize)
			}

			// All presets should have valid fade rate
			if trail.FadeRate < 0 || trail.FadeRate > 1 {
				t.Errorf("%s preset: fade rate must be 0-1, got %f", preset.name, trail.FadeRate)
			}

			// All presets should have non-negative spread
			if trail.SpreadX < 0 || trail.SpreadY < 0 {
				t.Errorf("%s preset: spread must be non-negative, got X=%f Y=%f",
					preset.name, trail.SpreadX, trail.SpreadY)
			}
		})
	}
}
