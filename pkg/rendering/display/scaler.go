package display

import "math"

// Scaler provides UI scaling calculations based on resolution.
// Uses 1920x1080 as baseline (scale factor 1.0).
type Scaler struct {
	config      *Config
	scaleFactor float64
}

// NewScaler creates a UI scaler.
func NewScaler(cfg *Config) *Scaler {
	base := BaseResolution()
	scaleFactor := float64(cfg.Width) / float64(base.Width)

	return &Scaler{
		config:      cfg,
		scaleFactor: scaleFactor,
	}
}

// GetScaleFactor returns current scale factor relative to 1920x1080.
func (s *Scaler) GetScaleFactor() float64 {
	return s.scaleFactor
}

// ScaleWidth scales a width value.
func (s *Scaler) ScaleWidth(width int) int {
	return int(math.Round(float64(width) * s.scaleFactor))
}

// ScaleHeight scales a height value.
func (s *Scaler) ScaleHeight(height int) int {
	return int(math.Round(float64(height) * s.scaleFactor))
}

// ScaleFloat scales a float64 value.
func (s *Scaler) ScaleFloat(value float64) float64 {
	return value * s.scaleFactor
}

// ScaleFontSize scales font size (minimum 8px).
func (s *Scaler) ScaleFontSize(size int) int {
	scaled := int(math.Round(float64(size) * s.scaleFactor))
	if scaled < 8 {
		scaled = 8
	}
	return scaled
}

// ScalePosition scales x,y coordinates.
func (s *Scaler) ScalePosition(x, y int) (int, int) {
	return s.ScaleWidth(x), s.ScaleHeight(y)
}

// ScaleSize scales width and height together.
func (s *Scaler) ScaleSize(width, height int) (int, int) {
	return s.ScaleWidth(width), s.ScaleHeight(height)
}

// UnscaleWidth converts scaled width back to base resolution.
func (s *Scaler) UnscaleWidth(width int) int {
	return int(math.Round(float64(width) / s.scaleFactor))
}

// UnscaleHeight converts scaled height back to base resolution.
func (s *Scaler) UnscaleHeight(height int) int {
	return int(math.Round(float64(height) / s.scaleFactor))
}

// UnscalePosition converts scaled coordinates back to base resolution.
func (s *Scaler) UnscalePosition(x, y int) (int, int) {
	return s.UnscaleWidth(x), s.UnscaleHeight(y)
}
