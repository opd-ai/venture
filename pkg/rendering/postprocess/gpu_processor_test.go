// Package postprocess provides post-processing effects for rendered scenes.
package postprocess

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestGPUProcessorConfig tests configuration get/set.
func TestGPUProcessorConfig(t *testing.T) {
	p := NewGPUProcessor()

	// Test default config
	config := p.GetConfig()
	if config.ColorGrading.Saturation != 1.0 {
		t.Errorf("expected default saturation 1.0, got %v", config.ColorGrading.Saturation)
	}

	// Test setting config
	config.ColorGrading.Saturation = 1.5
	config.Vignette.Intensity = 0.8
	p.SetConfig(config)

	retrieved := p.GetConfig()
	if retrieved.ColorGrading.Saturation != 1.5 {
		t.Errorf("expected saturation 1.5, got %v", retrieved.ColorGrading.Saturation)
	}
	if retrieved.Vignette.Intensity != 0.8 {
		t.Errorf("expected vignette intensity 0.8, got %v", retrieved.Vignette.Intensity)
	}
}

// TestGPUProcessorWithConfig tests creating processor with custom config.
func TestGPUProcessorWithConfig(t *testing.T) {
	config := DefaultConfig()
	config.ColorGrading.Brightness = 0.2
	config.ChromaticAberration.Enabled = true

	p := NewGPUProcessorWithConfig(config)

	retrieved := p.GetConfig()
	if retrieved.ColorGrading.Brightness != 0.2 {
		t.Errorf("expected brightness 0.2, got %v", retrieved.ColorGrading.Brightness)
	}
	if !retrieved.ChromaticAberration.Enabled {
		t.Error("expected chromatic aberration to be enabled")
	}
}

// TestIsNeutralColorGrading tests the neutral check function.
func TestIsNeutralColorGrading(t *testing.T) {
	tests := []struct {
		name     string
		config   ColorGradingConfig
		expected bool
	}{
		{
			name:     "default neutral",
			config:   DefaultColorGradingConfig(),
			expected: true,
		},
		{
			name: "non-neutral saturation",
			config: ColorGradingConfig{
				Saturation:  1.5,
				Contrast:    1.0,
				Brightness:  0.0,
				Temperature: 0.0,
				Tint:        0.0,
			},
			expected: false,
		},
		{
			name: "non-neutral brightness",
			config: ColorGradingConfig{
				Saturation:  1.0,
				Contrast:    1.0,
				Brightness:  0.1,
				Temperature: 0.0,
				Tint:        0.0,
			},
			expected: false,
		},
		{
			name: "non-neutral temperature",
			config: ColorGradingConfig{
				Saturation:  1.0,
				Contrast:    1.0,
				Brightness:  0.0,
				Temperature: 0.3,
				Tint:        0.0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNeutralColorGrading(tt.config)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestGPUProcessorApplyAllNoEffects tests that disabled effects return input unchanged.
func TestGPUProcessorApplyAllNoEffects(t *testing.T) {
	p := NewGPUProcessor()

	// Disable all effects explicitly
	config := p.GetConfig()
	config.ColorGrading.Enabled = false
	config.Vignette.Enabled = false
	config.ChromaticAberration.Enabled = false
	p.SetConfig(config)

	// Create a test image
	input := ebiten.NewImage(64, 64)
	input.Fill(color.RGBA{128, 128, 128, 255})

	// Apply should return input unchanged when no effects enabled
	result := p.ApplyAll(input)
	if result != input {
		t.Error("expected same image when no effects enabled")
	}

	input.Dispose()
}

// TestGPUProcessorApplyAllNilInput tests nil input handling.
func TestGPUProcessorApplyAllNilInput(t *testing.T) {
	p := NewGPUProcessor()

	result := p.ApplyAll(nil)
	if result != nil {
		t.Error("expected nil result for nil input")
	}
}

// TestGPUProcessorDispose tests resource cleanup.
func TestGPUProcessorDispose(t *testing.T) {
	p := NewGPUProcessor()

	// Enable effects and process an image to create resources
	config := p.GetConfig()
	config.ColorGrading.Enabled = true
	config.ColorGrading.Brightness = 0.1
	p.SetConfig(config)

	input := ebiten.NewImage(32, 32)
	input.Fill(color.RGBA{100, 100, 100, 255})
	_ = p.ApplyAll(input)

	// Dispose should not panic
	p.Dispose()

	// Dispose again should be safe
	p.Dispose()

	input.Dispose()
}

// TestGPUProcessorEnsureBuffer tests buffer management.
func TestGPUProcessorEnsureBuffer(t *testing.T) {
	p := NewGPUProcessor()

	// First call creates buffer
	p.ensureBuffer(100, 100)
	if p.outputBuffer == nil {
		t.Error("expected buffer to be created")
	}
	if p.bufferWidth != 100 || p.bufferHeight != 100 {
		t.Errorf("expected 100x100, got %dx%d", p.bufferWidth, p.bufferHeight)
	}

	// Same size should reuse buffer
	oldBuffer := p.outputBuffer
	p.ensureBuffer(100, 100)
	if p.outputBuffer != oldBuffer {
		t.Error("expected buffer to be reused for same size")
	}

	// Different size should create new buffer
	p.ensureBuffer(200, 150)
	if p.bufferWidth != 200 || p.bufferHeight != 150 {
		t.Errorf("expected 200x150, got %dx%d", p.bufferWidth, p.bufferHeight)
	}

	p.Dispose()
}

// BenchmarkGPUProcessorApplyAll benchmarks GPU processing performance.
func BenchmarkGPUProcessorApplyAll(b *testing.B) {
	p := NewGPUProcessor()

	// Enable all effects
	config := p.GetConfig()
	config.ColorGrading.Enabled = true
	config.ColorGrading.Saturation = 1.2
	config.ColorGrading.Contrast = 1.1
	config.Vignette.Enabled = true
	config.Vignette.Intensity = 0.5
	config.ChromaticAberration.Enabled = true
	config.ChromaticAberration.Intensity = 0.3
	p.SetConfig(config)

	// Create test image (smaller for unit test environment)
	input := ebiten.NewImage(640, 480)
	input.Fill(color.RGBA{128, 128, 128, 255})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.ApplyAll(input)
	}

	input.Dispose()
	p.Dispose()
}

// BenchmarkGPUProcessorNoEffects benchmarks bypass path.
func BenchmarkGPUProcessorNoEffects(b *testing.B) {
	p := NewGPUProcessor()

	// Disable all effects
	config := p.GetConfig()
	config.ColorGrading.Enabled = false
	config.Vignette.Enabled = false
	config.ChromaticAberration.Enabled = false
	p.SetConfig(config)

	input := ebiten.NewImage(640, 480)
	input.Fill(color.RGBA{128, 128, 128, 255})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.ApplyAll(input)
	}

	input.Dispose()
	p.Dispose()
}
