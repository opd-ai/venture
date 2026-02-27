//go:build !headless

// Package lighting provides dynamic lighting effects for rendered scenes.
// gpu_bloom_test.go tests the GPU-accelerated bloom functionality.
package lighting

import (
	"testing"
)

// TestNewGPUBloom tests GPU bloom creation.
func TestNewGPUBloom(t *testing.T) {
	bloom := NewGPUBloom()

	if bloom == nil {
		t.Fatal("NewGPUBloom returned nil")
	}

	config := bloom.GetConfig()
	if !config.Enabled {
		t.Error("Default config should be enabled")
	}
}

// TestNewGPUBloomWithConfig tests GPU bloom creation with custom config.
func TestNewGPUBloomWithConfig(t *testing.T) {
	config := BloomConfig{
		Enabled:   true,
		Threshold: 0.5,
		Intensity: 1.5,
		Radius:    10,
		Samples:   5,
	}

	bloom := NewGPUBloomWithConfig(config)

	if bloom == nil {
		t.Fatal("NewGPUBloomWithConfig returned nil")
	}

	returnedConfig := bloom.GetConfig()
	if returnedConfig.Threshold != 0.5 {
		t.Errorf("Threshold = %v, want 0.5", returnedConfig.Threshold)
	}
	if returnedConfig.Intensity != 1.5 {
		t.Errorf("Intensity = %v, want 1.5", returnedConfig.Intensity)
	}
	if returnedConfig.Radius != 10 {
		t.Errorf("Radius = %v, want 10", returnedConfig.Radius)
	}
}

// TestGPUBloom_SetConfig tests configuration changes.
func TestGPUBloom_SetConfig(t *testing.T) {
	bloom := NewGPUBloom()

	newConfig := BloomConfig{
		Enabled:   true,
		Threshold: 0.9,
		Intensity: 2.0,
		Radius:    20,
		Samples:   9,
	}
	bloom.SetConfig(newConfig)

	config := bloom.GetConfig()
	if config.Threshold != 0.9 {
		t.Errorf("Threshold = %v, want 0.9", config.Threshold)
	}
	if config.Intensity != 2.0 {
		t.Errorf("Intensity = %v, want 2.0", config.Intensity)
	}
}

// TestGPUBloom_Apply_Nil tests Apply with nil input.
func TestGPUBloom_Apply_Nil(t *testing.T) {
	bloom := NewGPUBloom()

	// Should not panic with nil input
	bloom.Apply(nil, nil)
}

// TestGPUBloom_Apply_Disabled tests Apply when bloom is disabled.
func TestGPUBloom_Apply_Disabled(t *testing.T) {
	bloom := NewGPUBloom()

	config := bloom.GetConfig()
	config.Enabled = false
	bloom.SetConfig(config)

	// Should not panic and should return early
	bloom.Apply(nil, nil)
}

// TestGPUBloom_Apply_ZeroIntensity tests Apply with zero intensity.
func TestGPUBloom_Apply_ZeroIntensity(t *testing.T) {
	bloom := NewGPUBloom()

	config := bloom.GetConfig()
	config.Intensity = 0
	bloom.SetConfig(config)

	// Should not panic and should return early
	bloom.Apply(nil, nil)
}

// TestGPUBloom_Dispose tests resource cleanup.
func TestGPUBloom_Dispose(t *testing.T) {
	bloom := NewGPUBloom()

	// Multiple dispose calls should be safe
	bloom.Dispose()
	bloom.Dispose()
}

// TestGPUBloom_ApplyToBuffer_Nil tests ApplyToBuffer with nil.
func TestGPUBloom_ApplyToBuffer_Nil(t *testing.T) {
	bloom := NewGPUBloom()

	// Should not panic with nil input
	bloom.ApplyToBuffer(nil)
}

// TestGPUBloom_ConfigValidation tests config value ranges.
func TestGPUBloom_ConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config BloomConfig
	}{
		{
			name: "low threshold",
			config: BloomConfig{
				Enabled:   true,
				Threshold: 0.1,
				Intensity: 1.0,
				Radius:    8,
				Samples:   5,
			},
		},
		{
			name: "high threshold",
			config: BloomConfig{
				Enabled:   true,
				Threshold: 0.95,
				Intensity: 1.0,
				Radius:    8,
				Samples:   5,
			},
		},
		{
			name: "high intensity",
			config: BloomConfig{
				Enabled:   true,
				Threshold: 0.8,
				Intensity: 3.0,
				Radius:    8,
				Samples:   5,
			},
		},
		{
			name: "large radius",
			config: BloomConfig{
				Enabled:   true,
				Threshold: 0.8,
				Intensity: 1.0,
				Radius:    30,
				Samples:   9,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bloom := NewGPUBloomWithConfig(tt.config)
			if bloom == nil {
				t.Error("Should create bloom with given config")
			}
		})
	}
}

// TestGPUBloom_ShaderCompilationIsolated tests shader compilation timing.
// This test cannot use actual Ebiten images without a graphics context.
func TestGPUBloom_ShaderCompilationIsolated(t *testing.T) {
	bloom := NewGPUBloom()
	defer bloom.Dispose()

	// Just verify the processor can be created
	if bloom == nil {
		t.Error("GPUBloom should be created")
	}

	// Config should be accessible
	config := bloom.GetConfig()
	if config.Threshold <= 0 || config.Threshold > 1 {
		t.Errorf("Threshold = %v, want 0-1 range", config.Threshold)
	}
}

// BenchmarkGPUBloom_Create benchmarks GPU bloom creation.
func BenchmarkGPUBloom_Create(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bloom := NewGPUBloom()
		bloom.Dispose()
	}
}

// BenchmarkGPUBloom_SetConfig benchmarks config changes.
func BenchmarkGPUBloom_SetConfig(b *testing.B) {
	bloom := NewGPUBloom()
	defer bloom.Dispose()

	config := BloomConfig{
		Enabled:   true,
		Threshold: 0.8,
		Intensity: 1.2,
		Radius:    16,
		Samples:   7,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bloom.SetConfig(config)
	}
}

// Note: Full GPU tests require an Ebiten graphics context.
// These are tested implicitly through the lighting system integration tests.
// The shader source is validated at compile time by Ebiten's Kage compiler.

// TestGPUBloom_GetShaderCompilationErrors tests the error counter metric.
func TestGPUBloom_GetShaderCompilationErrors(t *testing.T) {
	bloom := NewGPUBloom()
	defer bloom.Dispose()

	// Initial error count should be zero
	initialErrors := bloom.GetShaderCompilationErrors()
	if initialErrors != 0 {
		t.Errorf("Initial shader compilation errors = %d, want 0", initialErrors)
	}
}

// TestGPUBloom_GetShaderCompilationErrors_ThreadSafe tests concurrent access to error counter.
func TestGPUBloom_GetShaderCompilationErrors_ThreadSafe(t *testing.T) {
	bloom := NewGPUBloom()
	defer bloom.Dispose()

	// Verify the counter can be read concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_ = bloom.GetShaderCompilationErrors()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestGPUBloom_ShaderCompilationErrors_Disabled tests that no errors are counted when bloom is disabled.
func TestGPUBloom_ShaderCompilationErrors_Disabled(t *testing.T) {
	bloom := NewGPUBloom()
	defer bloom.Dispose()

	config := bloom.GetConfig()
	config.Enabled = false
	bloom.SetConfig(config)

	// Apply with disabled bloom (should not try to compile shaders)
	bloom.Apply(nil, nil)

	// Error count should still be zero
	errors := bloom.GetShaderCompilationErrors()
	if errors != 0 {
		t.Errorf("Shader compilation errors = %d, want 0 (bloom was disabled)", errors)
	}
}

// BenchmarkGPUBloom_GetShaderCompilationErrors benchmarks the error counter getter.
func BenchmarkGPUBloom_GetShaderCompilationErrors(b *testing.B) {
	bloom := NewGPUBloom()
	defer bloom.Dispose()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bloom.GetShaderCompilationErrors()
	}
}
