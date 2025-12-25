//go:build !headless
// +build !headless

package engine

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/rendering/postprocess"
	"github.com/sirupsen/logrus"
)

// PostProcessorAdapter wraps pkg/rendering/postprocess.Processor for ECS integration.
// Provides Phase 5.3 post-processing effects: motion blur, depth blur, color grading,
// vignette, and chromatic aberration.
type PostProcessorAdapter struct {
	processor *postprocess.Processor
	enabled   bool
	logger    *logrus.Entry
}

// NewPostProcessorAdapter creates a post-processing adapter with default configuration.
func NewPostProcessorAdapter(logger *logrus.Entry) *PostProcessorAdapter {
	return &PostProcessorAdapter{
		processor: postprocess.NewProcessor(),
		enabled:   false,
		logger:    logger,
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
	p.processor.SetConfig(config)
}

// GetConfig returns the current configuration.
func (p *PostProcessorAdapter) GetConfig() postprocess.Config {
	return p.processor.GetConfig()
}

// SetGenrePreset applies a genre-specific visual style preset.
func (p *PostProcessorAdapter) SetGenrePreset(genreID string) {
	preset := postprocess.GetPresetByGenre(genreID)
	p.processor.SetConfig(preset.Config)
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
func (p *PostProcessorAdapter) Apply(input *ebiten.Image) *ebiten.Image {
	if !p.enabled || input == nil {
		return input
	}

	// Convert Ebiten image to standard image.RGBA
	bounds := input.Bounds()
	rgba := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, input.At(x, y))
		}
	}

	// Apply post-processing effects (no velocity or depth maps for now)
	processed := p.processor.ApplyAll(rgba, nil, nil)

	// Convert back to Ebiten image
	result := ebiten.NewImageFromImage(processed)
	return result
}

// EnableColorGrading enables color grading with specified parameters.
func (p *PostProcessorAdapter) EnableColorGrading(saturation, contrast, brightness, temperature, tint float64) {
	config := p.processor.GetConfig()
	config.ColorGrading.Enabled = true
	config.ColorGrading.Saturation = saturation
	config.ColorGrading.Contrast = contrast
	config.ColorGrading.Brightness = brightness
	config.ColorGrading.Temperature = temperature
	config.ColorGrading.Tint = tint
	p.processor.SetConfig(config)
}

// EnableVignette enables vignette effect with specified intensity and softness.
func (p *PostProcessorAdapter) EnableVignette(intensity, softness float64) {
	config := p.processor.GetConfig()
	config.Vignette.Enabled = true
	config.Vignette.Intensity = intensity
	config.Vignette.Softness = softness
	p.processor.SetConfig(config)
}

// EnableChromaticAberration enables chromatic aberration with specified intensity and direction.
func (p *PostProcessorAdapter) EnableChromaticAberration(intensity, directionX, directionY float64, samples int) {
	config := p.processor.GetConfig()
	config.ChromaticAberration.Enabled = true
	config.ChromaticAberration.Intensity = intensity
	config.ChromaticAberration.DirectionX = directionX
	config.ChromaticAberration.DirectionY = directionY
	config.ChromaticAberration.Samples = samples
	p.processor.SetConfig(config)
}

// DisableAll disables all post-processing effects.
func (p *PostProcessorAdapter) DisableAll() {
	config := p.processor.GetConfig()
	config.MotionBlur.Enabled = false
	config.DepthBlur.Enabled = false
	config.ColorGrading.Enabled = false
	config.Vignette.Enabled = false
	config.ChromaticAberration.Enabled = false
	p.processor.SetConfig(config)
}
