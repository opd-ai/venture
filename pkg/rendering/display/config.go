package display

import "fmt"

// Resolution represents a standard display resolution.
type Resolution struct {
	Width  int
	Height int
	Name   string
}

// standardResolutions defines supported resolutions per Phase 43.
// Use GetStandardResolutions() to access a safe copy of this slice.
var standardResolutions = []Resolution{
	{Width: 1280, Height: 720, Name: "HD"},
	{Width: 1920, Height: 1080, Name: "Full HD"},
	{Width: 2560, Height: 1440, Name: "QHD"},
	{Width: 3840, Height: 2160, Name: "4K UHD"},
}

// GetStandardResolutions returns a copy of the standard resolutions.
// This prevents external code from mutating the package-level slice.
func GetStandardResolutions() []Resolution {
	result := make([]Resolution, len(standardResolutions))
	copy(result, standardResolutions)
	return result
}

// Config holds display configuration.
type Config struct {
	Width      int
	Height     int
	Fullscreen bool
	VSync      bool
}

// NewConfig creates a display configuration with validation.
func NewConfig(width, height int, fullscreen bool) (*Config, error) {
	if !IsValidResolution(width, height) {
		return nil, fmt.Errorf("unsupported resolution: %dx%d", width, height)
	}

	return &Config{
		Width:      width,
		Height:     height,
		Fullscreen: fullscreen,
		VSync:      true, // Enable by default for smooth rendering
	}, nil
}

// NewConfigDefault creates default 1920x1080 configuration.
// Returns an error if the default resolution is not valid, though this should never happen
// in practice since 1920x1080 is a standard supported resolution.
func NewConfigDefault() (*Config, error) {
	cfg, err := NewConfig(1920, 1080, false)
	if err != nil {
		return nil, fmt.Errorf("NewConfigDefault: failed to create default config: %w", err)
	}
	return cfg, nil
}

// GetNearestValidResolution returns the nearest standard resolution for non-standard
// aspect ratios or mobile screen sizes. Uses Euclidean distance in pixel space.
func GetNearestValidResolution(width, height int) Resolution {
	best := standardResolutions[0]
	bestDist := (width-best.Width)*(width-best.Width) + (height-best.Height)*(height-best.Height)

	for _, res := range standardResolutions[1:] {
		dist := (width-res.Width)*(width-res.Width) + (height-res.Height)*(height-res.Height)
		if dist < bestDist {
			bestDist = dist
			best = res
		}
	}
	return best
}

// IsValidResolution checks if resolution is supported.
func IsValidResolution(width, height int) bool {
	for _, res := range standardResolutions {
		if res.Width == width && res.Height == height {
			return true
		}
	}
	return false
}

// GetResolutionByName returns resolution by name.
func GetResolutionByName(name string) (Resolution, bool) {
	for _, res := range standardResolutions {
		if res.Name == name {
			return res, true
		}
	}
	return Resolution{}, false
}

// GetResolution returns current resolution as Resolution struct.
func (c *Config) GetResolution() Resolution {
	for _, res := range standardResolutions {
		if res.Width == c.Width && res.Height == c.Height {
			return res
		}
	}
	return Resolution{Width: c.Width, Height: c.Height, Name: "Custom"}
}

// AspectRatio calculates aspect ratio.
func (c *Config) AspectRatio() float64 {
	return float64(c.Width) / float64(c.Height)
}

// BaseResolution returns baseline 1920x1080 for scaling calculations.
func BaseResolution() Resolution {
	return standardResolutions[1] // 1920x1080
}
