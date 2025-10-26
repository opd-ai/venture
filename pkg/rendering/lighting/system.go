// Package lighting provides dynamic lighting effects for rendered scenes.
package lighting

import (
	"image"
	"image/color"
	"math"
)

// System manages multiple light sources and applies lighting to images.
// Future feature: This system is designed for dynamic lighting but not yet integrated into the main game loop.
// Planned integration in Phase 9+ for enhanced visual effects.
type System struct {
	lights []Light
	config LightingConfig
}

// NewSystem creates a new lighting system with default configuration.
func NewSystem() *System {
	return &System{
		lights: make([]Light, 0),
		config: DefaultConfig(),
	}
}

// NewSystemWithConfig creates a new lighting system with custom configuration.
func NewSystemWithConfig(config LightingConfig) *System {
	return &System{
		lights: make([]Light, 0),
		config: config,
	}
}

// AddLight adds a light source to the system.
func (s *System) AddLight(light Light) error {
	if err := light.Validate(); err != nil {
		return err
	}
	if len(s.lights) >= s.config.MaxLights {
		return &ValidationError{Field: "lights", Message: "maximum number of lights reached"}
	}
	s.lights = append(s.lights, light)
	return nil
}

// RemoveLight removes a light at the specified index.
func (s *System) RemoveLight(index int) error {
	if index < 0 || index >= len(s.lights) {
		return &ValidationError{Field: "index", Message: "out of bounds"}
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
		return Light{}, &ValidationError{Field: "index", Message: "out of bounds"}
	}
	return s.lights[index], nil
}

// UpdateLight updates the light at the specified index.
func (s *System) UpdateLight(index int, light Light) error {
	if index < 0 || index >= len(s.lights) {
		return &ValidationError{Field: "index", Message: "out of bounds"}
	}
	if err := light.Validate(); err != nil {
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
