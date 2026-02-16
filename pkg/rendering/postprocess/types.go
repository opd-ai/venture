// Package postprocess provides post-processing effects for rendered scenes.
package postprocess

import (
	"image"
	"image/color"
)

// EffectType defines the type of post-processing effect.
type EffectType int

const (
	// EffectMotionBlur applies velocity-based blur
	EffectMotionBlur EffectType = iota
	// EffectDepthBlur applies depth-based blur (depth of field)
	EffectDepthBlur
	// EffectColorGrading applies color adjustments (saturation, contrast, temperature)
	EffectColorGrading
	// EffectVignette darkens the edges of the screen
	EffectVignette
	// EffectChromaticAberration separates color channels at edges
	EffectChromaticAberration
)

// String returns the string representation of an effect type.
func (e EffectType) String() string {
	switch e {
	case EffectMotionBlur:
		return "MotionBlur"
	case EffectDepthBlur:
		return "DepthBlur"
	case EffectColorGrading:
		return "ColorGrading"
	case EffectVignette:
		return "Vignette"
	case EffectChromaticAberration:
		return "ChromaticAberration"
	default:
		return "Unknown"
	}
}

// MotionBlurConfig contains configuration for motion blur effect.
type MotionBlurConfig struct {
	// Enabled toggles the effect
	Enabled bool

	// Intensity controls blur strength (0.0-1.0, typical 0.3-0.7)
	Intensity float64

	// Samples is the number of blur samples (3-15 typical)
	// Higher values create smoother blur but cost more performance
	Samples int

	// VelocityX is the horizontal velocity in pixels per frame
	VelocityX float64

	// VelocityY is the vertical velocity in pixels per frame
	VelocityY float64
}

// DefaultMotionBlurConfig returns a sensible default motion blur configuration.
func DefaultMotionBlurConfig() MotionBlurConfig {
	return MotionBlurConfig{
		Enabled:   false,
		Intensity: 0.5,
		Samples:   7,
		VelocityX: 0.0,
		VelocityY: 0.0,
	}
}

// DepthBlurConfig contains configuration for depth-of-field blur effect.
type DepthBlurConfig struct {
	// Enabled toggles the effect
	Enabled bool

	// FocalDistance is the distance at which objects are in focus (0.0-1.0)
	FocalDistance float64

	// FocalRange is the range of distances that remain in focus (0.0-1.0)
	FocalRange float64

	// BlurStrength controls maximum blur amount (0.0-1.0, typical 0.3-0.8)
	BlurStrength float64

	// Samples is the number of blur samples (3-15 typical)
	Samples int
}

// DefaultDepthBlurConfig returns a sensible default depth blur configuration.
func DefaultDepthBlurConfig() DepthBlurConfig {
	return DepthBlurConfig{
		Enabled:       false,
		FocalDistance: 0.5,
		FocalRange:    0.2,
		BlurStrength:  0.5,
		Samples:       7,
	}
}

// ColorGradingConfig contains configuration for color grading effect.
type ColorGradingConfig struct {
	// Enabled toggles the effect
	Enabled bool

	// Saturation multiplier (0.0 = grayscale, 1.0 = normal, >1.0 = oversaturated)
	// Typical range: 0.5-1.5
	Saturation float64

	// Contrast multiplier (0.0 = flat gray, 1.0 = normal, >1.0 = high contrast)
	// Typical range: 0.5-2.0
	Contrast float64

	// Brightness adjustment (-1.0 to 1.0, 0.0 = no change)
	// Typical range: -0.3 to 0.3
	Brightness float64

	// Temperature adjustment (-1.0 = cooler/blue, 0.0 = neutral, 1.0 = warmer/orange)
	// Typical range: -0.5 to 0.5
	Temperature float64

	// Tint adjustment (-1.0 = green, 0.0 = neutral, 1.0 = magenta)
	// Typical range: -0.3 to 0.3
	Tint float64
}

// DefaultColorGradingConfig returns a sensible default color grading configuration.
func DefaultColorGradingConfig() ColorGradingConfig {
	return ColorGradingConfig{
		Enabled:     true,
		Saturation:  1.0,
		Contrast:    1.0,
		Brightness:  0.0,
		Temperature: 0.0,
		Tint:        0.0,
	}
}

// VignetteConfig contains configuration for vignette effect.
type VignetteConfig struct {
	// Enabled toggles the effect
	Enabled bool

	// Intensity controls darkness of vignette (0.0-1.0, typical 0.3-0.7)
	Intensity float64

	// Softness controls edge softness (0.0-1.0, typical 0.3-0.8)
	// Higher values create a more gradual falloff
	Softness float64

	// InnerRadius is the radius at which vignette starts (0.0-1.0)
	InnerRadius float64

	// OuterRadius is the radius at which vignette is fully dark (0.0-1.0)
	OuterRadius float64

	// Color is the vignette color (typically black or dark gray)
	Color color.Color
}

// DefaultVignetteConfig returns a sensible default vignette configuration.
func DefaultVignetteConfig() VignetteConfig {
	return VignetteConfig{
		Enabled:     true,
		Intensity:   0.5,
		Softness:    0.6,
		InnerRadius: 0.4,
		OuterRadius: 1.2,
		Color:       color.RGBA{0, 0, 0, 255},
	}
}

// ChromaticAberrationConfig contains configuration for chromatic aberration effect.
type ChromaticAberrationConfig struct {
	// Enabled toggles the effect
	Enabled bool

	// Intensity controls the strength of color separation (0.0-1.0, typical 0.1-0.5)
	Intensity float64

	// Samples is the number of color separation samples (2-5 typical)
	// Higher values create smoother separation but cost more performance
	Samples int

	// DirectionX is the horizontal direction of aberration (-1.0 to 1.0)
	DirectionX float64

	// DirectionY is the vertical direction of aberration (-1.0 to 1.0)
	DirectionY float64
}

// DefaultChromaticAberrationConfig returns a sensible default chromatic aberration configuration.
func DefaultChromaticAberrationConfig() ChromaticAberrationConfig {
	return ChromaticAberrationConfig{
		Enabled:    false,
		Intensity:  0.2,
		Samples:    3,
		DirectionX: 1.0,
		DirectionY: 0.0,
	}
}

// Config contains all post-processing effect configurations.
type Config struct {
	// MotionBlur configuration
	MotionBlur MotionBlurConfig

	// DepthBlur configuration
	DepthBlur DepthBlurConfig

	// ColorGrading configuration
	ColorGrading ColorGradingConfig

	// Vignette configuration
	Vignette VignetteConfig

	// ChromaticAberration configuration
	ChromaticAberration ChromaticAberrationConfig
}

// DefaultConfig returns a default post-processing configuration with all effects disabled or neutral.
func DefaultConfig() Config {
	return Config{
		MotionBlur:          DefaultMotionBlurConfig(),
		DepthBlur:           DefaultDepthBlurConfig(),
		ColorGrading:        DefaultColorGradingConfig(),
		Vignette:            DefaultVignetteConfig(),
		ChromaticAberration: DefaultChromaticAberrationConfig(),
	}
}

// Preset represents a named collection of post-processing settings.
type Preset struct {
	// Name is the human-readable name of the preset
	Name string

	// Description provides context for when to use this preset
	Description string

	// Config is the post-processing configuration
	Config Config
}

// VelocityMap represents per-pixel velocity data for motion blur.
type VelocityMap struct {
	// Bounds is the image bounds
	Bounds image.Rectangle

	// VelocityX contains horizontal velocity for each pixel
	VelocityX [][]float64

	// VelocityY contains vertical velocity for each pixel
	VelocityY [][]float64
}

// NewVelocityMap creates a new velocity map with the given bounds.
func NewVelocityMap(bounds image.Rectangle) *VelocityMap {
	width := bounds.Dx()
	height := bounds.Dy()

	velX := make([][]float64, height)
	velY := make([][]float64, height)
	for y := 0; y < height; y++ {
		velX[y] = make([]float64, width)
		velY[y] = make([]float64, width)
	}

	return &VelocityMap{
		Bounds:    bounds,
		VelocityX: velX,
		VelocityY: velY,
	}
}

// GetVelocity returns the velocity at the given pixel coordinates.
func (v *VelocityMap) GetVelocity(x, y int) (vx, vy float64) {
	if x < v.Bounds.Min.X || x >= v.Bounds.Max.X ||
		y < v.Bounds.Min.Y || y >= v.Bounds.Max.Y {
		return 0, 0
	}

	localX := x - v.Bounds.Min.X
	localY := y - v.Bounds.Min.Y

	return v.VelocityX[localY][localX], v.VelocityY[localY][localX]
}

// SetVelocity sets the velocity at the given pixel coordinates.
func (v *VelocityMap) SetVelocity(x, y int, vx, vy float64) {
	if x < v.Bounds.Min.X || x >= v.Bounds.Max.X ||
		y < v.Bounds.Min.Y || y >= v.Bounds.Max.Y {
		return
	}

	localX := x - v.Bounds.Min.X
	localY := y - v.Bounds.Min.Y

	v.VelocityX[localY][localX] = vx
	v.VelocityY[localY][localX] = vy
}

// Validate checks that all configuration values are within valid ranges.
// Returns nil if the configuration is valid.
func (c *Config) Validate() error {
	if err := c.MotionBlur.Validate(); err != nil {
		return err
	}
	if err := c.DepthBlur.Validate(); err != nil {
		return err
	}
	if err := c.ChromaticAberration.Validate(); err != nil {
		return err
	}
	if c.Vignette.Intensity < 0 || c.Vignette.Intensity > 1 {
		return &ValidationError{Field: "Vignette.Intensity", Message: "must be between 0.0 and 1.0"}
	}
	if c.Vignette.InnerRadius >= c.Vignette.OuterRadius && c.Vignette.Enabled {
		return &ValidationError{Field: "Vignette.InnerRadius", Message: "must be less than OuterRadius"}
	}
	return nil
}

// Validate checks that motion blur configuration values are within valid ranges.
func (c *MotionBlurConfig) Validate() error {
	if c.Intensity < 0 || c.Intensity > 1 {
		return &ValidationError{Field: "MotionBlur.Intensity", Message: "must be between 0.0 and 1.0"}
	}
	if c.Samples < 2 {
		return &ValidationError{Field: "MotionBlur.Samples", Message: "must be at least 2"}
	}
	return nil
}

// Validate checks that depth blur configuration values are within valid ranges.
func (c *DepthBlurConfig) Validate() error {
	if c.BlurStrength < 0 || c.BlurStrength > 1 {
		return &ValidationError{Field: "DepthBlur.BlurStrength", Message: "must be between 0.0 and 1.0"}
	}
	if c.Samples < 1 {
		return &ValidationError{Field: "DepthBlur.Samples", Message: "must be at least 1"}
	}
	if c.FocalDistance < 0 || c.FocalDistance > 1 {
		return &ValidationError{Field: "DepthBlur.FocalDistance", Message: "must be between 0.0 and 1.0"}
	}
	return nil
}

// Validate checks that chromatic aberration configuration values are within valid ranges.
func (c *ChromaticAberrationConfig) Validate() error {
	if c.Intensity < 0 || c.Intensity > 1 {
		return &ValidationError{Field: "ChromaticAberration.Intensity", Message: "must be between 0.0 and 1.0"}
	}
	if c.Samples < 2 {
		return &ValidationError{Field: "ChromaticAberration.Samples", Message: "must be at least 2"}
	}
	return nil
}

// ValidationError represents a post-processing configuration validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return "postprocess: " + e.Field + " " + e.Message
}
