package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/quality"
)

func TestNewQualitySystem(t *testing.T) {
	config := quality.MediumQualityConfig()
	qs := NewQualitySystem(&config, 60.0)

	if qs == nil {
		t.Fatal("NewQualitySystem returned nil")
	}

	if !qs.IsEnabled() {
		t.Error("QualitySystem should be enabled by default")
	}

	if !qs.IsAutoAdjustEnabled() {
		t.Error("Auto-adjust should be enabled by default")
	}

	gotConfig := qs.GetConfig()
	if gotConfig.Level != quality.QualityMedium {
		t.Errorf("Initial quality level = %v, want %v", gotConfig.Level, quality.QualityMedium)
	}
}

func TestNewQualitySystem_NilConfig(t *testing.T) {
	qs := NewQualitySystem(nil, 60.0)

	if qs == nil {
		t.Fatal("NewQualitySystem returned nil")
	}

	// Should use default config
	gotConfig := qs.GetConfig()
	if gotConfig.Level != quality.QualityMedium {
		t.Errorf("Default quality level = %v, want %v", gotConfig.Level, quality.QualityMedium)
	}
}

func TestQualitySystem_Update(t *testing.T) {
	config := quality.HighQualityConfig()
	qs := NewQualitySystem(&config, 60.0)

	// Update with good frame time (60 FPS)
	for i := 0; i < 120; i++ {
		qs.Update(0.0167) // ~16.7ms in seconds
	}

	stats := qs.GetStats()
	if stats.AverageFPS < 55 || stats.AverageFPS > 65 {
		t.Errorf("Average FPS = %.1f, want ~60", stats.AverageFPS)
	}
}

func TestQualitySystem_SetGetConfig(t *testing.T) {
	config := quality.MediumQualityConfig()
	qs := NewQualitySystem(&config, 60.0)

	// Change to high quality
	highConfig := quality.HighQualityConfig()
	err := qs.SetConfig(&highConfig)
	if err != nil {
		t.Errorf("SetConfig failed: %v", err)
	}

	gotConfig := qs.GetConfig()
	if gotConfig.Level != quality.QualityHigh {
		t.Errorf("Quality level = %v, want %v", gotConfig.Level, quality.QualityHigh)
	}
}

func TestQualitySystem_SetConfig_Invalid(t *testing.T) {
	config := quality.MediumQualityConfig()
	qs := NewQualitySystem(&config, 60.0)

	// Try to set invalid config
	invalidConfig := quality.Config{
		SpriteDetailLevel: -0.5, // Invalid
	}

	err := qs.SetConfig(&invalidConfig)
	if err == nil {
		t.Error("SetConfig should fail for invalid config")
	}

	// Config should remain unchanged
	gotConfig := qs.GetConfig()
	if gotConfig.Level != quality.QualityMedium {
		t.Error("Config should remain unchanged after invalid SetConfig")
	}
}

func TestQualitySystem_SetQualityLevel(t *testing.T) {
	config := quality.MediumQualityConfig()
	qs := NewQualitySystem(&config, 60.0)

	// Change quality level
	qs.SetQualityLevel(quality.QualityLow)

	level := qs.GetQualityLevel()
	if level != quality.QualityLow {
		t.Errorf("Quality level = %v, want %v", level, quality.QualityLow)
	}

	// Verify config was updated
	gotConfig := qs.GetConfig()
	if gotConfig.Level != quality.QualityLow {
		t.Errorf("Config level = %v, want %v", gotConfig.Level, quality.QualityLow)
	}
}

func TestQualitySystem_EnableDisable(t *testing.T) {
	config := quality.MediumQualityConfig()
	qs := NewQualitySystem(&config, 60.0)

	// Disable system
	qs.Disable()
	if qs.IsEnabled() {
		t.Error("QualitySystem should be disabled")
	}

	// Enable system
	qs.Enable()
	if !qs.IsEnabled() {
		t.Error("QualitySystem should be enabled")
	}
}

func TestQualitySystem_EnableDisableAutoAdjust(t *testing.T) {
	config := quality.MediumQualityConfig()
	qs := NewQualitySystem(&config, 60.0)

	// Disable auto-adjust
	qs.DisableAutoAdjust()
	if qs.IsAutoAdjustEnabled() {
		t.Error("Auto-adjust should be disabled")
	}

	// Enable auto-adjust
	qs.EnableAutoAdjust()
	if !qs.IsAutoAdjustEnabled() {
		t.Error("Auto-adjust should be enabled")
	}
}

func TestQualitySystem_OnQualityChange(t *testing.T) {
	config := quality.HighQualityConfig()
	qs := NewQualitySystem(&config, 60.0)

	callbackCalled := false
	var callbackLevel quality.QualityLevel

	qs.SetOnQualityChange(func(level quality.QualityLevel) {
		callbackCalled = true
		callbackLevel = level
	})

	// Change quality
	qs.SetQualityLevel(quality.QualityLow)

	if !callbackCalled {
		t.Error("Quality change callback was not called")
	}

	if callbackLevel != quality.QualityLow {
		t.Errorf("Callback level = %v, want %v", callbackLevel, quality.QualityLow)
	}
}

func TestQualitySystem_GetStats(t *testing.T) {
	config := quality.MediumQualityConfig()
	qs := NewQualitySystem(&config, 60.0)

	// Generate some frame data
	for i := 0; i < 60; i++ {
		qs.Update(0.0167)
	}

	stats := qs.GetStats()
	if stats.SampleCount == 0 {
		t.Error("Stats should have sample count > 0")
	}

	if stats.AverageFPS <= 0 {
		t.Error("Stats should have positive average FPS")
	}
}

func TestGetEntityQualityOverride(t *testing.T) {
	world := NewWorld()
	entity := world.CreateEntity()

	// No override initially
	override, hasOverride := GetEntityQualityOverride(entity)
	if hasOverride {
		t.Error("Entity should not have override initially")
	}

	// Add override
	entity.AddComponent(quality.WithSpriteDetail(0.5))

	override, hasOverride = GetEntityQualityOverride(entity)
	if !hasOverride {
		t.Error("Entity should have override after adding component")
	}

	if override.SpriteDetailOverride != 0.5 {
		t.Errorf("Override detail = %f, want 0.5", override.SpriteDetailOverride)
	}
}

func TestApplyQualityToSpriteDetail(t *testing.T) {
	config := quality.MediumQualityConfig()
	world := NewWorld()

	tests := []struct {
		name     string
		setup    func(*Entity)
		expected float64
	}{
		{
			name:     "no override uses global",
			setup:    func(e *Entity) {},
			expected: config.SpriteDetailLevel,
		},
		{
			name: "entity override takes precedence",
			setup: func(e *Entity) {
				e.AddComponent(quality.WithSpriteDetail(0.8))
			},
			expected: 0.8,
		},
		{
			name: "disabled override uses global",
			setup: func(e *Entity) {
				comp := quality.NewQualitySettingsComponent()
				comp.Override = false
				comp.SpriteDetailOverride = 0.3
				e.AddComponent(comp)
			},
			expected: config.SpriteDetailLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := world.CreateEntity()
			tt.setup(entity)

			detail := ApplyQualityToSpriteDetail(&config, entity)
			if detail != tt.expected {
				t.Errorf("Sprite detail = %f, want %f", detail, tt.expected)
			}
		})
	}
}

func TestApplyQualityToParticleCount(t *testing.T) {
	config := quality.MediumQualityConfig()
	world := NewWorld()

	tests := []struct {
		name     string
		setup    func(*Entity)
		expected float64
	}{
		{
			name:     "no override uses global",
			setup:    func(e *Entity) {},
			expected: config.ParticleCountMultiplier,
		},
		{
			name: "entity override takes precedence",
			setup: func(e *Entity) {
				e.AddComponent(quality.WithParticleMultiplier(0.3))
			},
			expected: 0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := world.CreateEntity()
			tt.setup(entity)

			multiplier := ApplyQualityToParticleCount(&config, entity)
			if multiplier != tt.expected {
				t.Errorf("Particle multiplier = %f, want %f", multiplier, tt.expected)
			}
		})
	}
}

func TestShouldRenderEffects(t *testing.T) {
	world := NewWorld()

	tests := []struct {
		name     string
		config   quality.Config
		setup    func(*Entity)
		expected bool
	}{
		{
			name:     "high quality enables effects",
			config:   quality.HighQualityConfig(),
			setup:    func(e *Entity) {},
			expected: true,
		},
		{
			name:     "low quality may disable effects",
			config:   quality.LowQualityConfig(),
			setup:    func(e *Entity) {},
			expected: false, // Low quality disables post-processing and bloom
		},
		{
			name:   "entity override disables effects",
			config: quality.HighQualityConfig(),
			setup: func(e *Entity) {
				e.AddComponent(quality.WithoutEffects())
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := world.CreateEntity()
			tt.setup(entity)

			shouldRender := ShouldRenderEffects(&tt.config, entity)
			if shouldRender != tt.expected {
				t.Errorf("ShouldRenderEffects = %v, want %v", shouldRender, tt.expected)
			}
		})
	}
}

func TestQualitySystem_UpdateWhenDisabled(t *testing.T) {
	config := quality.MediumQualityConfig()
	qs := NewQualitySystem(&config, 60.0)

	// Disable the system
	qs.Disable()

	// Update should be no-op
	initialStats := qs.GetStats()
	qs.Update(0.0167)
	afterStats := qs.GetStats()

	// Stats should be identical (no frames recorded)
	if initialStats.SampleCount != afterStats.SampleCount {
		t.Error("Update should be no-op when system is disabled")
	}
}

func TestQualitySystem_UpdateWhenAutoAdjustDisabled(t *testing.T) {
	config := quality.MediumQualityConfig()
	qs := NewQualitySystem(&config, 60.0)

	// Disable auto-adjust but keep system enabled
	qs.DisableAutoAdjust()

	// Update should still record stats but not adjust
	initialLevel := qs.GetQualityLevel()

	// Record poor performance
	for i := 0; i < 120; i++ {
		qs.Update(0.05) // ~20 FPS
	}

	// Quality should not have changed
	afterLevel := qs.GetQualityLevel()
	if initialLevel != afterLevel {
		t.Error("Quality should not auto-adjust when disabled")
	}
}

// Benchmarks

func BenchmarkQualitySystem_Update(b *testing.B) {
	config := quality.MediumQualityConfig()
	qs := NewQualitySystem(&config, 60.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qs.Update(0.0167)
	}
}

func BenchmarkQualitySystem_GetConfig(b *testing.B) {
	config := quality.MediumQualityConfig()
	qs := NewQualitySystem(&config, 60.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = qs.GetConfig()
	}
}

func BenchmarkApplyQualityToSpriteDetail(b *testing.B) {
	config := quality.MediumQualityConfig()
	world := NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(quality.WithSpriteDetail(0.5))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ApplyQualityToSpriteDetail(&config, entity)
	}
}
