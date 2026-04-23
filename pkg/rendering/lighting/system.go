// Package lighting provides dynamic lighting effects for rendered scenes.
package lighting

import (
	"image"
	"image/color"
	"math"

	log "github.com/sirupsen/logrus"
)

// System manages multiple light sources and applies lighting to images.
// This system is integrated via pkg/engine/lighting_adapter.go (LightingAdapter)
// which wraps this system for ECS compatibility.
type System struct {
	lights []Light
	config LightingConfig
	logger *log.Logger
}

// NewSystem creates a new lighting system with default configuration.
func NewSystem() *System {
	return &System{
		lights: make([]Light, 0),
		config: DefaultConfig(),
		logger: log.StandardLogger(),
	}
}

// NewSystemWithConfig creates a new lighting system with custom configuration.
func NewSystemWithConfig(config LightingConfig) *System {
	return &System{
		lights: make([]Light, 0),
		config: config,
		logger: log.StandardLogger(),
	}
}

// NewSystemWithLogger creates a new lighting system with custom configuration and logger.
func NewSystemWithLogger(config LightingConfig, logger *log.Logger) *System {
	if logger == nil {
		logger = log.StandardLogger()
	}
	return &System{
		lights: make([]Light, 0),
		config: config,
		logger: logger,
	}
}

// AddLight adds a light source to the system.
func (s *System) AddLight(light Light) error {
	if err := light.Validate(); err != nil {
		s.logger.WithFields(log.Fields{
			"package":    "lighting",
			"light_type": light.Type,
			"error":      err.Error(),
		}).Debug("failed to validate light")
		return err
	}
	if len(s.lights) >= s.config.MaxLights {
		err := &ValidationError{Field: "lights", Message: "maximum number of lights reached"}
		s.logger.WithFields(log.Fields{
			"package":    "lighting",
			"max_lights": s.config.MaxLights,
			"error":      err.Error(),
		}).Debug("cannot add light: maximum lights reached")
		return err
	}
	s.lights = append(s.lights, light)
	return nil
}

// RemoveLight removes a light at the specified index.
func (s *System) RemoveLight(index int) error {
	if index < 0 || index >= len(s.lights) {
		err := &ValidationError{Field: "index", Message: "out of bounds"}
		s.logger.WithFields(log.Fields{
			"package":     "lighting",
			"index":       index,
			"light_count": len(s.lights),
			"error":       err.Error(),
		}).Debug("cannot remove light: index out of bounds")
		return err
	}
	s.lights = append(s.lights[:index], s.lights[index+1:]...)
	return nil
}

// ClearLights removes all lights from the system.
func (s *System) ClearLights() {
	s.lights = s.lights[:0]
}

// GetLights returns a copy of all lights in the system.
func (s *System) GetLights() []Light {
	result := make([]Light, len(s.lights))
	copy(result, s.lights)
	return result
}

// GetLight returns the light at the specified index.
func (s *System) GetLight(index int) (Light, error) {
	if index < 0 || index >= len(s.lights) {
		err := &ValidationError{Field: "index", Message: "out of bounds"}
		s.logger.WithFields(log.Fields{
			"package":     "lighting",
			"index":       index,
			"light_count": len(s.lights),
			"error":       err.Error(),
		}).Debug("cannot get light: index out of bounds")
		return Light{}, err
	}
	return s.lights[index], nil
}

// UpdateLight updates the light at the specified index.
func (s *System) UpdateLight(index int, light Light) error {
	if index < 0 || index >= len(s.lights) {
		err := &ValidationError{Field: "index", Message: "out of bounds"}
		s.logger.WithFields(log.Fields{
			"package":     "lighting",
			"index":       index,
			"light_count": len(s.lights),
			"error":       err.Error(),
		}).Debug("cannot update light: index out of bounds")
		return err
	}
	if err := light.Validate(); err != nil {
		s.logger.WithFields(log.Fields{
			"package":    "lighting",
			"index":      index,
			"light_type": light.Type,
			"error":      err.Error(),
		}).Debug("cannot update light: validation failed")
		return err
	}
	s.lights[index] = light
	return nil
}

// LightCount returns the number of lights in the system.
func (s *System) LightCount() int {
	return len(s.lights)
}

// ApplyLighting applies lighting to an image and returns a new lit image.
func (s *System) ApplyLighting(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	// Process each pixel
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pos := image.Point{X: x, Y: y}
			baseColor := img.At(x, y)
			litColor := s.CalculateLighting(pos, baseColor)
			result.Set(x, y, litColor)
		}
	}

	return result
}

// CalculateLighting calculates the final lit color for a pixel.
func (s *System) CalculateLighting(pos image.Point, baseColor color.Color) color.Color {
	// Start with ambient light
	totalR := s.config.AmbientIntensity
	totalG := s.config.AmbientIntensity
	totalB := s.config.AmbientIntensity

	// Add contribution from each enabled light
	for _, light := range s.lights {
		if !light.Enabled {
			continue
		}

		intensity := s.calculateLightIntensity(pos, light)
		if intensity <= 0 {
			continue
		}

		// Get light color components
		lr, lg, lb, _ := light.Color.RGBA()
		lrf := float64(lr) / 65535.0
		lgf := float64(lg) / 65535.0
		lbf := float64(lb) / 65535.0

		// Add light contribution
		totalR += intensity * lrf * light.Intensity
		totalG += intensity * lgf * light.Intensity
		totalB += intensity * lbf * light.Intensity
	}

	// Clamp to valid range
	totalR = clamp(totalR, 0, 1)
	totalG = clamp(totalG, 0, 1)
	totalB = clamp(totalB, 0, 1)

	// Apply gamma correction if enabled
	if s.config.GammaCorrection > 0 {
		gamma := s.config.GammaCorrection
		totalR = math.Pow(totalR, 1.0/gamma)
		totalG = math.Pow(totalG, 1.0/gamma)
		totalB = math.Pow(totalB, 1.0/gamma)
	}

	// Modulate base color with lighting
	br, bg, bb, ba := baseColor.RGBA()
	brf := float64(br) / 65535.0
	bgf := float64(bg) / 65535.0
	bbf := float64(bb) / 65535.0

	finalR := uint8(clamp(brf*totalR, 0, 1) * 255)
	finalG := uint8(clamp(bgf*totalG, 0, 1) * 255)
	finalB := uint8(clamp(bbf*totalB, 0, 1) * 255)
	finalA := uint8(ba / 257) // Convert from 16-bit to 8-bit

	return color.RGBA{R: finalR, G: finalG, B: finalB, A: finalA}
}

// calculateLightIntensity calculates the intensity of a light at a given position.
func (s *System) calculateLightIntensity(pos image.Point, light Light) float64 {
	switch light.Type {
	case TypeAmbient:
		return light.Intensity

	case TypePoint:
		return s.calculatePointLightIntensity(pos, light)

	case TypeDirectional:
		return light.Intensity

	default:
		return 0
	}
}

// calculatePointLightIntensity calculates intensity for a point light.
func (s *System) calculatePointLightIntensity(pos image.Point, light Light) float64 {
	// Calculate distance from light to pixel
	dx := float64(pos.X - light.Position.X)
	dy := float64(pos.Y - light.Position.Y)
	distance := math.Sqrt(dx*dx + dy*dy)

	// If outside radius, no contribution
	if distance > light.Radius {
		return 0
	}

	// Calculate falloff based on type
	var attenuation float64
	switch light.Falloff {
	case FalloffNone:
		attenuation = 1.0

	case FalloffLinear:
		attenuation = 1.0 - (distance / light.Radius)

	case FalloffQuadratic:
		t := distance / light.Radius
		attenuation = 1.0 - (t * t)

	case FalloffInverseSquare:
		// Prevent division by zero
		if distance < 1.0 {
			distance = 1.0
		}
		attenuation = 1.0 / (distance * distance)
		// Normalize to radius
		maxIntensity := 1.0 / (light.Radius * light.Radius)
		attenuation = attenuation / maxIntensity
		if attenuation > 1.0 {
			attenuation = 1.0
		}

	default:
		attenuation = 1.0
	}

	return attenuation
}

// ApplyLightingToRegion applies lighting to a specific region of an image.
func (s *System) ApplyLightingToRegion(img *image.RGBA, region image.Rectangle) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	// First, copy all pixels from original
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			result.Set(x, y, img.At(x, y))
		}
	}

	// Then apply lighting only to the region
	for y := region.Min.Y; y < region.Max.Y; y++ {
		for x := region.Min.X; x < region.Max.X; x++ {
			pt := image.Point{X: x, Y: y}
			if !pt.In(bounds) {
				continue
			}
			baseColor := img.At(x, y)
			litColor := s.CalculateLighting(pt, baseColor)
			result.Set(x, y, litColor)
		}
	}

	return result
}

// GetConfig returns a copy of the system configuration.
func (s *System) GetConfig() LightingConfig {
	return s.config
}

// SetConfig updates the system configuration.
func (s *System) SetConfig(config LightingConfig) {
	s.config = config
}

// ApplyBloomToImage applies bloom/glow effects to an already lit image.
// This should be called after ApplyLighting for post-processing.
func (s *System) ApplyBloomToImage(img *image.RGBA) *image.RGBA {
	return ApplyBloom(img, s.config.BloomConfig)
}

// ApplyAOToImage applies ambient occlusion to an image.
// The depthMap can be nil to auto-generate from luminance.
func (s *System) ApplyAOToImage(img, depthMap *image.RGBA) *image.RGBA {
	return ApplyEnhancedAO(img, depthMap, s.config.AOConfig)
}

// ApplyFullPostProcessing applies all post-processing effects (AO, bloom).
// Order: AO first (darkens), then bloom (brightens highlights).
// The depthMap can be nil to auto-generate from luminance.
// Returns a new image; never modifies the input.
func (s *System) ApplyFullPostProcessing(img, depthMap *image.RGBA) *image.RGBA {
	result := img

	// Apply ambient occlusion first (darkening).
	// EnableShadows is a legacy compatibility toggle that forces AO on.
	if s.config.AOConfig.Enabled || s.config.EnableShadows {
		result = s.ApplyAOToImage(result, depthMap)
	}

	// Apply bloom last (brightening highlights)
	if s.config.BloomConfig.Enabled {
		result = s.ApplyBloomToImage(result)
	}

	// If no effects were applied, return a copy to avoid unintended mutations
	if result == img {
		bounds := img.Bounds()
		copy := image.NewRGBA(bounds)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				copy.Set(x, y, img.At(x, y))
			}
		}
		return copy
	}

	return result
}

// Helper functions

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
