package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/rendering/postprocess"
	"github.com/sirupsen/logrus"
)

// PostProcessorAdapter wraps pkg/rendering/postprocess.GPUProcessor for ECS integration.
// Provides Phase 5.3 post-processing effects: color grading, vignette, and chromatic aberration.
// Uses GPU-accelerated shaders instead of CPU-side pixel iteration (V1 fix: eliminates 40-120ms overhead).
type PostProcessorAdapter struct {
	gpuProcessor *postprocess.GPUProcessor
	enabled      bool
	logger       *logrus.Entry
}

// NewPostProcessorAdapter creates a post-processing adapter with default configuration.
// Uses GPU-accelerated processing for optimal performance.
func NewPostProcessorAdapter(logger *logrus.Entry) *PostProcessorAdapter {
	return &PostProcessorAdapter{
		gpuProcessor: postprocess.NewGPUProcessor(),
		enabled:      false,
		logger:       logger,
	}
}

// SetEnabled enables or disables all post-processing effects.
func (p *PostProcessorAdapter) SetEnabled(enabled bool) {
	p.enabled = enabled
	if p.logger != nil {
		p.logger.WithField("enabled", enabled).Debug("post-processing enabled state changed")
	}
}

// IsEnabled returns whether post-processing is enabled.
func (p *PostProcessorAdapter) IsEnabled() bool {
	return p.enabled
}

// SetConfig replaces the current configuration with a new one.
func (p *PostProcessorAdapter) SetConfig(config postprocess.Config) {
	p.gpuProcessor.SetConfig(config)
}

// GetConfig returns the current configuration.
func (p *PostProcessorAdapter) GetConfig() postprocess.Config {
	return p.gpuProcessor.GetConfig()
}

// SetGenrePreset applies a genre-specific visual style preset.
func (p *PostProcessorAdapter) SetGenrePreset(genreID string) {
	preset := postprocess.GetPresetByGenre(genreID)
	p.gpuProcessor.SetConfig(preset.Config)
	if p.logger != nil {
		p.logger.WithFields(logrus.Fields{
			"genre":      genreID,
			"preset":     preset.Name,
			"vignette":   preset.Config.Vignette.Enabled,
			"color_grad": preset.Config.ColorGrading.Enabled,
			"chromatic":  preset.Config.ChromaticAberration.Enabled,
		}).Info("post-processing genre preset applied")
	}
}

// Apply applies all enabled post-processing effects to the input image.
// Returns the processed image. If post-processing is disabled, returns input unchanged.
// Uses GPU-accelerated shaders (<1ms) instead of CPU pixel iteration (40-120ms).
func (p *PostProcessorAdapter) Apply(input *ebiten.Image) *ebiten.Image {
	if !p.enabled || input == nil {
		return input
	}

	// Apply GPU-accelerated post-processing effects
	return p.gpuProcessor.ApplyAll(input)
}

// EnableColorGrading enables color grading with specified parameters.
func (p *PostProcessorAdapter) EnableColorGrading(saturation, contrast, brightness, temperature, tint float64) {
	config := p.gpuProcessor.GetConfig()
	config.ColorGrading.Enabled = true
	config.ColorGrading.Saturation = saturation
	config.ColorGrading.Contrast = contrast
	config.ColorGrading.Brightness = brightness
	config.ColorGrading.Temperature = temperature
	config.ColorGrading.Tint = tint
	p.gpuProcessor.SetConfig(config)
}

// EnableVignette enables vignette effect with specified intensity and softness.
func (p *PostProcessorAdapter) EnableVignette(intensity, softness float64) {
	config := p.gpuProcessor.GetConfig()
	config.Vignette.Enabled = true
	config.Vignette.Intensity = intensity
	config.Vignette.Softness = softness
	p.gpuProcessor.SetConfig(config)
}

// EnableChromaticAberration enables chromatic aberration with specified intensity and direction.
// Note: The samples parameter is preserved for API compatibility but GPU shader uses fixed sampling.
func (p *PostProcessorAdapter) EnableChromaticAberration(intensity, directionX, directionY float64, samples int) {
	config := p.gpuProcessor.GetConfig()
	config.ChromaticAberration.Enabled = true
	config.ChromaticAberration.Intensity = intensity
	config.ChromaticAberration.DirectionX = directionX
	config.ChromaticAberration.DirectionY = directionY
	config.ChromaticAberration.Samples = samples
	p.gpuProcessor.SetConfig(config)
}

// DisableAll disables all post-processing effects.
func (p *PostProcessorAdapter) DisableAll() {
	config := p.gpuProcessor.GetConfig()
	config.MotionBlur.Enabled = false
	config.DepthBlur.Enabled = false
	config.ColorGrading.Enabled = false
	config.Vignette.Enabled = false
	config.ChromaticAberration.Enabled = false
	p.gpuProcessor.SetConfig(config)
}

// PrecompileShaders eagerly compiles all post-processing GPU shaders.
// Call this during initialization (e.g., loading screen) to eliminate the
// 50-200ms stutter that occurs when shaders are first used during gameplay.
// Returns an error if shader compilation fails; post-processing will still
// function but with a delayed compilation on first use.
func (p *PostProcessorAdapter) PrecompileShaders() error {
	if p.logger != nil {
		p.logger.Debug("pre-compiling post-processing shaders")
	}
	err := p.gpuProcessor.PrecompileShaders()
	if err != nil {
		if p.logger != nil {
			p.logger.WithField("error", err.Error()).Warn("shader pre-compilation failed")
		}
		return err
	}
	if p.logger != nil {
		p.logger.Info("post-processing shaders pre-compiled successfully")
	}
	return nil
}

// Dispose releases GPU resources held by the adapter.
// Call this when the adapter is no longer needed.
func (p *PostProcessorAdapter) Dispose() {
	if p.gpuProcessor != nil {
		p.gpuProcessor.Dispose()
	}
}
