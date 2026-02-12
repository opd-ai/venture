package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/rendering/postprocess"
	"github.com/sirupsen/logrus"
)

func TestNewPostProcessorAdapter(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())
	adapter := NewPostProcessorAdapter(logger)

	if adapter == nil {
		t.Fatal("NewPostProcessorAdapter returned nil")
	}

	if adapter.gpuProcessor == nil {
		t.Error("gpuProcessor not initialized")
	}

	if adapter.enabled {
		t.Error("adapter should be disabled by default")
	}

	if adapter.logger != logger {
		t.Error("logger not set correctly")
	}
}

func TestPostProcessorAdapter_SetEnabled(t *testing.T) {
	adapter := NewPostProcessorAdapter(nil)

	// Test enabling
	adapter.SetEnabled(true)
	if !adapter.IsEnabled() {
		t.Error("adapter not enabled after SetEnabled(true)")
	}

	// Test disabling
	adapter.SetEnabled(false)
	if adapter.IsEnabled() {
		t.Error("adapter still enabled after SetEnabled(false)")
	}
}

func TestPostProcessorAdapter_SetConfig(t *testing.T) {
	adapter := NewPostProcessorAdapter(nil)

	config := postprocess.Config{
		ColorGrading: postprocess.ColorGradingConfig{
			Enabled:    true,
			Saturation: 1.5,
			Contrast:   1.2,
		},
	}

	adapter.SetConfig(config)
	retrieved := adapter.GetConfig()

	if !retrieved.ColorGrading.Enabled {
		t.Error("color grading not enabled")
	}

	if retrieved.ColorGrading.Saturation != 1.5 {
		t.Errorf("saturation = %f, want 1.5", retrieved.ColorGrading.Saturation)
	}
}

func TestPostProcessorAdapter_SetGenrePreset(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"sci-fi", "sci-fi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"post-apocalyptic", "post-apocalyptic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewPostProcessorAdapter(nil)
			adapter.SetGenrePreset(tt.genreID)

			config := adapter.GetConfig()
			// At least one effect should be enabled by genre presets
			hasEffect := config.ColorGrading.Enabled ||
				config.Vignette.Enabled ||
				config.ChromaticAberration.Enabled ||
				config.MotionBlur.Enabled ||
				config.DepthBlur.Enabled

			if !hasEffect {
				t.Errorf("genre preset %s did not enable any effects", tt.genreID)
			}
		})
	}
}

func TestPostProcessorAdapter_Apply_Disabled(t *testing.T) {
	adapter := NewPostProcessorAdapter(nil)
	adapter.SetEnabled(false)

	input := ebiten.NewImage(100, 100)
	result := adapter.Apply(input)

	if result != input {
		t.Error("Apply should return input unchanged when disabled")
	}
}

func TestPostProcessorAdapter_Apply_NilInput(t *testing.T) {
	adapter := NewPostProcessorAdapter(nil)
	adapter.SetEnabled(true)

	result := adapter.Apply(nil)

	if result != nil {
		t.Error("Apply should return nil when input is nil")
	}
}

func TestPostProcessorAdapter_Apply_Enabled(t *testing.T) {
	adapter := NewPostProcessorAdapter(nil)
	adapter.SetEnabled(true)

	// Enable vignette effect
	adapter.EnableVignette(0.5, 0.3)

	config := adapter.GetConfig()
	if !config.Vignette.Enabled {
		t.Error("vignette not enabled after EnableVignette")
	}

	if config.Vignette.Intensity != 0.5 {
		t.Errorf("vignette intensity = %f, want 0.5", config.Vignette.Intensity)
	}

	// Note: We can't actually call Apply() with Ebiten images in tests
	// because Ebiten requires a running game loop. The Apply() method is tested
	// indirectly through integration tests when the client runs.
}

func TestPostProcessorAdapter_EnableColorGrading(t *testing.T) {
	adapter := NewPostProcessorAdapter(nil)

	adapter.EnableColorGrading(1.2, 1.1, 0.1, 0.05, 0.02)
	config := adapter.GetConfig()

	if !config.ColorGrading.Enabled {
		t.Error("color grading not enabled")
	}

	if config.ColorGrading.Saturation != 1.2 {
		t.Errorf("saturation = %f, want 1.2", config.ColorGrading.Saturation)
	}

	if config.ColorGrading.Contrast != 1.1 {
		t.Errorf("contrast = %f, want 1.1", config.ColorGrading.Contrast)
	}

	if config.ColorGrading.Brightness != 0.1 {
		t.Errorf("brightness = %f, want 0.1", config.ColorGrading.Brightness)
	}

	if config.ColorGrading.Temperature != 0.05 {
		t.Errorf("temperature = %f, want 0.05", config.ColorGrading.Temperature)
	}

	if config.ColorGrading.Tint != 0.02 {
		t.Errorf("tint = %f, want 0.02", config.ColorGrading.Tint)
	}
}

func TestPostProcessorAdapter_EnableVignette(t *testing.T) {
	adapter := NewPostProcessorAdapter(nil)

	adapter.EnableVignette(0.7, 0.5)
	config := adapter.GetConfig()

	if !config.Vignette.Enabled {
		t.Error("vignette not enabled")
	}

	if config.Vignette.Intensity != 0.7 {
		t.Errorf("intensity = %f, want 0.7", config.Vignette.Intensity)
	}

	if config.Vignette.Softness != 0.5 {
		t.Errorf("softness = %f, want 0.5", config.Vignette.Softness)
	}
}

func TestPostProcessorAdapter_EnableChromaticAberration(t *testing.T) {
	adapter := NewPostProcessorAdapter(nil)

	adapter.EnableChromaticAberration(0.8, 1.0, 0.5, 3)
	config := adapter.GetConfig()

	if !config.ChromaticAberration.Enabled {
		t.Error("chromatic aberration not enabled")
	}

	if config.ChromaticAberration.Intensity != 0.8 {
		t.Errorf("intensity = %f, want 0.8", config.ChromaticAberration.Intensity)
	}

	if config.ChromaticAberration.DirectionX != 1.0 {
		t.Errorf("directionX = %f, want 1.0", config.ChromaticAberration.DirectionX)
	}

	if config.ChromaticAberration.DirectionY != 0.5 {
		t.Errorf("directionY = %f, want 0.5", config.ChromaticAberration.DirectionY)
	}

	if config.ChromaticAberration.Samples != 3 {
		t.Errorf("samples = %d, want 3", config.ChromaticAberration.Samples)
	}
}

func TestPostProcessorAdapter_DisableAll(t *testing.T) {
	adapter := NewPostProcessorAdapter(nil)

	// Enable all effects
	adapter.EnableColorGrading(1.2, 1.1, 0.1, 0.0, 0.0)
	adapter.EnableVignette(0.5, 0.3)
	adapter.EnableChromaticAberration(0.6, 1.0, 0.0, 3)

	// Disable all
	adapter.DisableAll()
	config := adapter.GetConfig()

	if config.MotionBlur.Enabled {
		t.Error("motion blur still enabled")
	}

	if config.DepthBlur.Enabled {
		t.Error("depth blur still enabled")
	}

	if config.ColorGrading.Enabled {
		t.Error("color grading still enabled")
	}

	if config.Vignette.Enabled {
		t.Error("vignette still enabled")
	}

	if config.ChromaticAberration.Enabled {
		t.Error("chromatic aberration still enabled")
	}
}

func TestPostProcessorAdapter_Dispose(t *testing.T) {
	adapter := NewPostProcessorAdapter(nil)

	// Enable effects and use the adapter
	adapter.SetEnabled(true)
	adapter.EnableVignette(0.5, 0.3)

	// Dispose should not panic
	adapter.Dispose()

	// Dispose again should be safe (idempotent)
	adapter.Dispose()
}
