// Package engine provides the stereoscopic rendering component for VR support.

package engine

import (
	"encoding/json"
	"sync"
)

// StereoscopicComponent tracks VR stereoscopic rendering state for an entity.
// It manages interpupillary distance, convergence settings, and lens distortion
// parameters for rendering separate left and right eye views.
type StereoscopicComponent struct {
	mu sync.RWMutex

	// Enabled controls whether stereoscopic rendering is active
	Enabled bool `json:"enabled"`

	// IPD is the interpupillary distance in millimeters (55-75mm typical)
	IPD float64 `json:"ipd"`

	// Convergence is the focal distance where both eyes converge (in world units)
	Convergence float64 `json:"convergence"`

	// EyeSeparation is the calculated half-IPD for camera offset (derived from IPD)
	EyeSeparation float64 `json:"eye_separation"`

	// BarrelDistortion enables lens distortion correction
	BarrelDistortion bool `json:"barrel_distortion"`

	// DistortionK1 is the first radial distortion coefficient
	DistortionK1 float64 `json:"distortion_k1"`

	// DistortionK2 is the second radial distortion coefficient
	DistortionK2 float64 `json:"distortion_k2"`

	// RenderWidth is the width of each eye's render target
	RenderWidth int `json:"render_width"`

	// RenderHeight is the height of each eye's render target
	RenderHeight int `json:"render_height"`

	// CurrentEye tracks which eye is being rendered ("left" or "right")
	CurrentEye string `json:"current_eye"`
}

const (
	// EyeLeft represents the left eye view
	EyeLeft = "left"

	// EyeRight represents the right eye view
	EyeRight = "right"

	// DefaultIPD is the average interpupillary distance in mm
	DefaultIPD = 63.0

	// MinIPD is the minimum supported IPD in mm
	MinIPD = 55.0

	// MaxIPD is the maximum supported IPD in mm
	MaxIPD = 75.0

	// DefaultConvergence is the default focal distance in world units
	DefaultConvergence = 10.0

	// DefaultDistortionK1 is a typical barrel distortion coefficient
	DefaultDistortionK1 = 0.22

	// DefaultDistortionK2 is a typical barrel distortion coefficient
	DefaultDistortionK2 = 0.24
)

// NewStereoscopicComponent creates a new stereoscopic component with default settings.
func NewStereoscopicComponent() *StereoscopicComponent {
	c := &StereoscopicComponent{
		Enabled:          false,
		IPD:              DefaultIPD,
		Convergence:      DefaultConvergence,
		BarrelDistortion: true,
		DistortionK1:     DefaultDistortionK1,
		DistortionK2:     DefaultDistortionK2,
		RenderWidth:      960, // Half of 1920 for side-by-side
		RenderHeight:     1080,
		CurrentEye:       EyeLeft,
	}
	c.updateEyeSeparation()
	return c
}

// Type returns the component type identifier.
func (c *StereoscopicComponent) Type() string {
	return "stereoscopic"
}

// SetIPD sets the interpupillary distance and updates eye separation.
// The IPD is clamped to the valid range (55-75mm).
func (c *StereoscopicComponent) SetIPD(ipd float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clamp IPD to valid range
	if ipd < MinIPD {
		ipd = MinIPD
	}
	if ipd > MaxIPD {
		ipd = MaxIPD
	}

	c.IPD = ipd
	c.updateEyeSeparation()
}

// GetIPD returns the current interpupillary distance.
func (c *StereoscopicComponent) GetIPD() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.IPD
}

// updateEyeSeparation calculates eye separation from IPD.
// Call with lock held.
func (c *StereoscopicComponent) updateEyeSeparation() {
	// Convert mm to world units (assuming 1 world unit = 1 meter)
	// IPD is split between left and right eye offsets
	c.EyeSeparation = (c.IPD / 1000.0) / 2.0
}

// GetEyeOffset returns the camera offset for the specified eye.
// Left eye returns negative offset, right eye returns positive.
func (c *StereoscopicComponent) GetEyeOffset(eye string) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if eye == EyeLeft {
		return -c.EyeSeparation
	}
	return c.EyeSeparation
}

// SetEnabled enables or disables stereoscopic rendering.
func (c *StereoscopicComponent) SetEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Enabled = enabled
}

// IsEnabled returns whether stereoscopic rendering is enabled.
func (c *StereoscopicComponent) IsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Enabled
}

// SetConvergence sets the focal convergence distance.
func (c *StereoscopicComponent) SetConvergence(convergence float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if convergence < 0.1 {
		convergence = 0.1
	}
	c.Convergence = convergence
}

// GetConvergence returns the current convergence distance.
func (c *StereoscopicComponent) GetConvergence() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Convergence
}

// SetDistortion sets the barrel distortion coefficients.
func (c *StereoscopicComponent) SetDistortion(k1, k2 float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.DistortionK1 = k1
	c.DistortionK2 = k2
}

// GetDistortion returns the current barrel distortion coefficients.
func (c *StereoscopicComponent) GetDistortion() (k1, k2 float64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.DistortionK1, c.DistortionK2
}

// SetBarrelDistortion enables or disables barrel distortion correction.
func (c *StereoscopicComponent) SetBarrelDistortion(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.BarrelDistortion = enabled
}

// IsBarrelDistortionEnabled returns whether barrel distortion is enabled.
func (c *StereoscopicComponent) IsBarrelDistortionEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.BarrelDistortion
}

// SetRenderSize sets the render target dimensions for each eye.
func (c *StereoscopicComponent) SetRenderSize(width, height int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	c.RenderWidth = width
	c.RenderHeight = height
}

// GetRenderSize returns the render target dimensions.
func (c *StereoscopicComponent) GetRenderSize() (width, height int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.RenderWidth, c.RenderHeight
}

// SetCurrentEye sets which eye is currently being rendered.
func (c *StereoscopicComponent) SetCurrentEye(eye string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if eye != EyeLeft && eye != EyeRight {
		eye = EyeLeft
	}
	c.CurrentEye = eye
}

// GetCurrentEye returns which eye is currently being rendered.
func (c *StereoscopicComponent) GetCurrentEye() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.CurrentEye
}

// ApplyBarrelDistortion applies barrel distortion to normalized coordinates.
// Input coordinates are in range [-1, 1] from center.
// Returns distorted coordinates in the same range.
func (c *StereoscopicComponent) ApplyBarrelDistortion(x, y float64) (float64, float64) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.BarrelDistortion {
		return x, y
	}

	// Calculate radius squared from center
	r2 := x*x + y*y
	r4 := r2 * r2

	// Apply radial distortion
	distortion := 1.0 + c.DistortionK1*r2 + c.DistortionK2*r4

	return x * distortion, y * distortion
}

// Serialize converts the component to JSON bytes.
func (c *StereoscopicComponent) Serialize() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return json.Marshal(c)
}

// Deserialize loads component state from JSON bytes.
func (c *StereoscopicComponent) Deserialize(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return json.Unmarshal(data, c)
}
