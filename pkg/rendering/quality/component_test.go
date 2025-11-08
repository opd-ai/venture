package quality

import (
	"testing"
)

func TestQualitySettingsComponent_Type(t *testing.T) {
	comp := NewQualitySettingsComponent()
	if comp.Type() != "quality_settings" {
		t.Errorf("Type() = %q, want %q", comp.Type(), "quality_settings")
	}
}

func TestNewQualitySettingsComponent(t *testing.T) {
	comp := NewQualitySettingsComponent()

	if comp.Override {
		t.Error("Override should be false by default")
	}

	if comp.DisableEffects {
		t.Error("DisableEffects should be false by default")
	}

	if comp.SpriteDetailOverride != 1.0 {
		t.Errorf("SpriteDetailOverride = %f, want 1.0", comp.SpriteDetailOverride)
	}

	if !comp.EnableAntiAliasingOverride {
		t.Error("EnableAntiAliasingOverride should be true by default")
	}

	if comp.ParticleCountMultiplierOverride != 1.0 {
		t.Errorf("ParticleCountMultiplierOverride = %f, want 1.0", comp.ParticleCountMultiplierOverride)
	}
}

func TestWithSpriteDetail(t *testing.T) {
	tests := []struct {
		detail float64
	}{
		{0.0},
		{0.3},
		{0.5},
		{0.8},
		{1.0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			comp := WithSpriteDetail(tt.detail)

			if !comp.Override {
				t.Error("Override should be true")
			}

			if comp.SpriteDetailOverride != tt.detail {
				t.Errorf("SpriteDetailOverride = %f, want %f", comp.SpriteDetailOverride, tt.detail)
			}
		})
	}
}

func TestWithParticleMultiplier(t *testing.T) {
	tests := []struct {
		multiplier float64
	}{
		{0.0},
		{0.25},
		{0.5},
		{0.75},
		{1.0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			comp := WithParticleMultiplier(tt.multiplier)

			if !comp.Override {
				t.Error("Override should be true")
			}

			if comp.ParticleCountMultiplierOverride != tt.multiplier {
				t.Errorf("ParticleCountMultiplierOverride = %f, want %f",
					comp.ParticleCountMultiplierOverride, tt.multiplier)
			}
		})
	}
}

func TestWithoutEffects(t *testing.T) {
	comp := WithoutEffects()

	if !comp.DisableEffects {
		t.Error("DisableEffects should be true")
	}
}

func TestQualitySettingsComponent_UseCases(t *testing.T) {
	t.Run("background entity with reduced detail", func(t *testing.T) {
		comp := WithSpriteDetail(0.3)
		if comp.SpriteDetailOverride != 0.3 {
			t.Error("Should allow low detail for background entities")
		}
	})

	t.Run("important entity with no effects disabled", func(t *testing.T) {
		comp := NewQualitySettingsComponent()
		if comp.DisableEffects {
			t.Error("Default component should not disable effects")
		}
	})

	t.Run("distant entity with no effects", func(t *testing.T) {
		comp := WithoutEffects()
		if !comp.DisableEffects {
			t.Error("Should disable effects for distant entities")
		}
	})

	t.Run("particle-heavy entity with reduced particles", func(t *testing.T) {
		comp := WithParticleMultiplier(0.5)
		if comp.ParticleCountMultiplierOverride != 0.5 {
			t.Error("Should allow reducing particles for specific entities")
		}
	})
}

func TestQualitySettingsComponent_Integration(t *testing.T) {
	// Simulate entity with quality override
	type Entity struct {
		components map[string]interface{}
	}

	entity := &Entity{
		components: make(map[string]interface{}),
	}

	comp := WithSpriteDetail(0.5)
	entity.components[comp.Type()] = comp

	// Retrieve component
	retrieved, ok := entity.components["quality_settings"]
	if !ok {
		t.Fatal("Component not found in entity")
	}

	qualComp, ok := retrieved.(QualitySettingsComponent)
	if !ok {
		t.Fatal("Component is not QualitySettingsComponent")
	}

	if qualComp.SpriteDetailOverride != 0.5 {
		t.Errorf("Retrieved component detail = %f, want 0.5", qualComp.SpriteDetailOverride)
	}
}

// Benchmarks

func BenchmarkNewQualitySettingsComponent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewQualitySettingsComponent()
	}
}

func BenchmarkWithSpriteDetail(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = WithSpriteDetail(0.5)
	}
}

func BenchmarkWithParticleMultiplier(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = WithParticleMultiplier(0.5)
	}
}

func BenchmarkWithoutEffects(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = WithoutEffects()
	}
}

func BenchmarkComponent_Type(b *testing.B) {
	comp := NewQualitySettingsComponent()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = comp.Type()
	}
}
