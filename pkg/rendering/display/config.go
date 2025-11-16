package display

import "fmt"

// Resolution represents a standard display resolution.
type Resolution struct {
	Width  int
	Height int
	Name   string
}

// StandardResolutions defines supported resolutions per Phase 43.
var StandardResolutions = []Resolution{
	{Width: 1280, Height: 720, Name: "HD"},
	{Width: 1920, Height: 1080, Name: "Full HD"},
	{Width: 2560, Height: 1440, Name: "QHD"},
	{Width: 3840, Height: 2160, Name: "4K UHD"},
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
func NewConfigDefault() *Config {
	cfg, _ := NewConfig(1920, 1080, false)
	return cfg
}

// IsValidResolution checks if resolution is supported.
func IsValidResolution(width, height int) bool {
	for _, res := range StandardResolutions {
		if res.Width == width && res.Height == height {
			return true
		}
	}
	return false
}

// GetResolutionByName returns resolution by name.
func GetResolutionByName(name string) (Resolution, bool) {
	for _, res := range StandardResolutions {
		if res.Name == name {
			return res, true
		}
	}
	return Resolution{}, false
}

// GetResolution returns current resolution as Resolution struct.
func (c *Config) GetResolution() Resolution {
	for _, res := range StandardResolutions {
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
	return StandardResolutions[1] // 1920x1080
}
