package display

import (
	"fmt"
)

// ScaleMode defines how the game viewport scales to fit the window.
type ScaleMode int

const (
	// ScaleModeFit scales to fit window while maintaining aspect ratio (letterboxing/pillarboxing).
	ScaleModeFit ScaleMode = iota
	// ScaleModeFill stretches to fill entire window (may distort aspect ratio).
	ScaleModeFill
	// ScaleModeStretch same as Fill, kept for backward compatibility.
	ScaleModeStretch
)

// String returns the string representation of ScaleMode.
func (s ScaleMode) String() string {
	switch s {
	case ScaleModeFit:
		return "Fit"
	case ScaleModeFill:
		return "Fill"
	case ScaleModeStretch:
		return "Stretch"
	default:
		return "Unknown"
	}
}

// Common resolution presets
const (
	// MinWidth is the minimum supported screen width.
	MinWidth = 640
	// MinHeight is the minimum supported screen height.
	MinHeight = 480

	// DefaultWidth is the default screen width (Full HD).
	DefaultWidth = 1920
	// DefaultHeight is the default screen height (Full HD).
	DefaultHeight = 1080

	// LegacyWidth is the v6.0 legacy screen width (HD).
	LegacyWidth = 1280
	// LegacyHeight is the v6.0 legacy screen height (HD).
	LegacyHeight = 720
)

// Config holds display configuration settings.
type Config struct {
	// Width is the window width in pixels.
	Width int
	// Height is the window height in pixels.
	Height int
	// Fullscreen enables fullscreen mode.
	Fullscreen bool
	// ScaleMode determines how content scales to fit the window.
	ScaleMode ScaleMode
	// VSync enables vertical synchronization (default: true).
	VSync bool
}

// NewDefaultConfig creates a Config with default settings (1920x1080, windowed).
func NewDefaultConfig() Config {
	return Config{
		Width:      DefaultWidth,
		Height:     DefaultHeight,
		Fullscreen: false,
		ScaleMode:  ScaleModeFit,
		VSync:      true,
	}
}

// NewLegacyConfig creates a Config with v6.0 legacy settings (1280x720, windowed).
func NewLegacyConfig() Config {
	return Config{
		Width:      LegacyWidth,
		Height:     LegacyHeight,
		Fullscreen: false,
		ScaleMode:  ScaleModeFit,
		VSync:      true,
	}
}

// NewConfig creates a Config with specified dimensions.
func NewConfig(width, height int, fullscreen bool) Config {
	return Config{
		Width:      width,
		Height:     height,
		Fullscreen: fullscreen,
		ScaleMode:  ScaleModeFit,
		VSync:      true,
	}
}

// Validate checks if the configuration is valid.
func (c Config) Validate() error {
	if c.Width < MinWidth {
		return fmt.Errorf("width %d is below minimum %d", c.Width, MinWidth)
	}
	if c.Height < MinHeight {
		return fmt.Errorf("height %d is below minimum %d", c.Height, MinHeight)
	}
	if c.Width > 7680 {
		return fmt.Errorf("width %d exceeds maximum 7680 (8K)", c.Width)
	}
	if c.Height > 4320 {
		return fmt.Errorf("height %d exceeds maximum 4320 (8K)", c.Height)
	}
	return nil
}

// AspectRatio returns the aspect ratio as width/height.
func (c Config) AspectRatio() float64 {
	if c.Height == 0 {
		return 16.0 / 9.0 // Default to 16:9
	}
	return float64(c.Width) / float64(c.Height)
}

// IsCommonResolution returns true if this is a common gaming resolution.
func (c Config) IsCommonResolution() bool {
	switch {
	case c.Width == 1280 && c.Height == 720: // HD
		return true
	case c.Width == 1920 && c.Height == 1080: // Full HD
		return true
	case c.Width == 2560 && c.Height == 1440: // 2K
		return true
	case c.Width == 3840 && c.Height == 2160: // 4K
		return true
	default:
		return false
	}
}

// ResolutionName returns a friendly name for the resolution.
func (c Config) ResolutionName() string {
	switch {
	case c.Width == 1280 && c.Height == 720:
		return "HD (720p)"
	case c.Width == 1920 && c.Height == 1080:
		return "Full HD (1080p)"
	case c.Width == 2560 && c.Height == 1440:
		return "2K (1440p)"
	case c.Width == 3840 && c.Height == 2160:
		return "4K (2160p)"
	default:
		return fmt.Sprintf("%dx%d", c.Width, c.Height)
	}
}
