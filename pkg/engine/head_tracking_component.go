// Package engine provides the head tracking component for VR support.

package engine

import (
	"encoding/json"
	"math"
	"sync"
)

// HeadTrackingComponent tracks VR headset orientation and position.
// It manages pitch, yaw, roll rotation and positional tracking
// with smoothing and prediction for latency compensation.
type HeadTrackingComponent struct {
	mu sync.RWMutex

	// Enabled controls whether head tracking is active
	Enabled bool `json:"enabled"`

	// Pitch is up/down rotation in radians (-π/2 to π/2)
	Pitch float64 `json:"pitch"`

	// Yaw is left/right rotation in radians (0 to 2π)
	Yaw float64 `json:"yaw"`

	// Roll is tilt rotation in radians (-π to π)
	Roll float64 `json:"roll"`

	// PositionX is the head X offset in world units
	PositionX float64 `json:"position_x"`

	// PositionY is the head Y offset in world units
	PositionY float64 `json:"position_y"`

	// PositionZ is the head Z offset in world units
	PositionZ float64 `json:"position_z"`

	// SmoothingFactor controls jitter reduction (0.0 = no smoothing, 1.0 = max smoothing)
	SmoothingFactor float64 `json:"smoothing_factor"`

	// PredictionMs is the prediction lookahead for latency compensation
	PredictionMs float64 `json:"prediction_ms"`

	// Previous frame values for smoothing
	prevPitch     float64
	prevYaw       float64
	prevRoll      float64
	prevPositionX float64
	prevPositionY float64
	prevPositionZ float64

	// Velocity estimates for prediction
	pitchVelocity float64
	yawVelocity   float64
	rollVelocity  float64
	posXVelocity  float64
	posYVelocity  float64
	posZVelocity  float64

	// RecenterOffset stores the offset to apply when recentering view
	RecenterYaw float64 `json:"recenter_yaw"`
}

const (
	// DefaultSmoothingFactor provides moderate jitter reduction
	DefaultSmoothingFactor = 0.3

	// DefaultPredictionMs provides minimal latency compensation
	DefaultPredictionMs = 10.0

	// MaxPitch limits vertical look angle (85 degrees)
	MaxPitch = 1.4835 // ~85 degrees in radians

	// MinPitch limits vertical look angle (-85 degrees)
	MinPitch = -1.4835
)

// NewHeadTrackingComponent creates a new head tracking component with defaults.
func NewHeadTrackingComponent() *HeadTrackingComponent {
	return &HeadTrackingComponent{
		Enabled:         false,
		Pitch:           0,
		Yaw:             0,
		Roll:            0,
		PositionX:       0,
		PositionY:       0,
		PositionZ:       0,
		SmoothingFactor: DefaultSmoothingFactor,
		PredictionMs:    DefaultPredictionMs,
	}
}

// Type returns the component type identifier.
func (c *HeadTrackingComponent) Type() string {
	return "head_tracking"
}

// SetEnabled enables or disables head tracking.
func (c *HeadTrackingComponent) SetEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Enabled = enabled
}

// IsEnabled returns whether head tracking is enabled.
func (c *HeadTrackingComponent) IsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Enabled
}

// SetOrientation sets the head rotation angles in radians.
// Pitch is clamped to prevent looking straight up/down.
// Yaw is normalized to [0, 2π).
func (c *HeadTrackingComponent) SetOrientation(pitch, yaw, roll float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clamp pitch to prevent gimbal lock
	if pitch > MaxPitch {
		pitch = MaxPitch
	}
	if pitch < MinPitch {
		pitch = MinPitch
	}

	// Normalize yaw to [0, 2π)
	yaw = normalizeAngle(yaw)

	// Normalize roll to [-π, π]
	roll = normalizeAngleSignedVR(roll)

	// Apply smoothing if enabled
	if c.SmoothingFactor > 0 {
		pitch = c.smoothValue(c.prevPitch, pitch)
		yaw = c.smoothAngle(c.prevYaw, yaw)
		roll = c.smoothValue(c.prevRoll, roll)
	}

	// Calculate velocities for prediction
	c.pitchVelocity = pitch - c.prevPitch
	c.yawVelocity = shortestAngularDistance(c.prevYaw, yaw)
	c.rollVelocity = roll - c.prevRoll

	// Store current as previous
	c.prevPitch = pitch
	c.prevYaw = yaw
	c.prevRoll = roll

	c.Pitch = pitch
	c.Yaw = yaw
	c.Roll = roll
}

// GetOrientation returns the current head rotation angles in radians.
// Values are smoothed and optionally predicted.
func (c *HeadTrackingComponent) GetOrientation() (pitch, yaw, roll float64) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	pitch = c.Pitch
	yaw = c.Yaw - c.RecenterYaw
	roll = c.Roll

	// Apply prediction if enabled
	if c.PredictionMs > 0 {
		// Assume ~16.67ms frame time, scale velocity
		predictionFactor := c.PredictionMs / 16.67
		pitch += c.pitchVelocity * predictionFactor
		yaw += c.yawVelocity * predictionFactor
		roll += c.rollVelocity * predictionFactor

		// Re-clamp after prediction
		if pitch > MaxPitch {
			pitch = MaxPitch
		}
		if pitch < MinPitch {
			pitch = MinPitch
		}
	}

	yaw = normalizeAngle(yaw)
	return pitch, yaw, roll
}

// SetPosition sets the head position offset in world units.
func (c *HeadTrackingComponent) SetPosition(x, y, z float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Apply smoothing if enabled
	if c.SmoothingFactor > 0 {
		x = c.smoothValue(c.prevPositionX, x)
		y = c.smoothValue(c.prevPositionY, y)
		z = c.smoothValue(c.prevPositionZ, z)
	}

	// Calculate velocities for prediction
	c.posXVelocity = x - c.prevPositionX
	c.posYVelocity = y - c.prevPositionY
	c.posZVelocity = z - c.prevPositionZ

	// Store current as previous
	c.prevPositionX = x
	c.prevPositionY = y
	c.prevPositionZ = z

	c.PositionX = x
	c.PositionY = y
	c.PositionZ = z
}

// GetPosition returns the current head position offset.
func (c *HeadTrackingComponent) GetPosition() (x, y, z float64) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	x = c.PositionX
	y = c.PositionY
	z = c.PositionZ

	// Apply prediction if enabled
	if c.PredictionMs > 0 {
		predictionFactor := c.PredictionMs / 16.67
		x += c.posXVelocity * predictionFactor
		y += c.posYVelocity * predictionFactor
		z += c.posZVelocity * predictionFactor
	}

	return x, y, z
}

// SetSmoothingFactor sets the smoothing amount (0.0 to 1.0).
func (c *HeadTrackingComponent) SetSmoothingFactor(factor float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if factor < 0 {
		factor = 0
	}
	if factor > 1 {
		factor = 1
	}
	c.SmoothingFactor = factor
}

// GetSmoothingFactor returns the current smoothing factor.
func (c *HeadTrackingComponent) GetSmoothingFactor() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.SmoothingFactor
}

// SetPredictionMs sets the prediction lookahead in milliseconds.
func (c *HeadTrackingComponent) SetPredictionMs(ms float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if ms < 0 {
		ms = 0
	}
	if ms > 100 {
		ms = 100 // Cap at 100ms
	}
	c.PredictionMs = ms
}

// GetPredictionMs returns the current prediction lookahead.
func (c *HeadTrackingComponent) GetPredictionMs() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.PredictionMs
}

// RecenterView resets the yaw to the current forward direction.
// After calling this, the current yaw becomes the new "center".
func (c *HeadTrackingComponent) RecenterView() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.RecenterYaw = c.Yaw
}

// ResetRecenter clears the recenter offset.
func (c *HeadTrackingComponent) ResetRecenter() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.RecenterYaw = 0
}

// smoothValue applies exponential smoothing to a value.
func (c *HeadTrackingComponent) smoothValue(prev, current float64) float64 {
	return prev + (current-prev)*(1.0-c.SmoothingFactor)
}

// smoothAngle applies smoothing to an angle, handling wrap-around.
func (c *HeadTrackingComponent) smoothAngle(prev, current float64) float64 {
	diff := shortestAngularDistance(prev, current)
	return normalizeAngle(prev + diff*(1.0-c.SmoothingFactor))
}

// normalizeAngleSignedVR normalizes an angle to [-π, π].
func normalizeAngleSignedVR(angle float64) float64 {
	for angle > math.Pi {
		angle -= 2 * math.Pi
	}
	for angle < -math.Pi {
		angle += 2 * math.Pi
	}
	return angle
}

// Serialize converts the component to JSON bytes.
func (c *HeadTrackingComponent) Serialize() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return json.Marshal(c)
}

// Deserialize loads component state from JSON bytes.
func (c *HeadTrackingComponent) Deserialize(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return json.Unmarshal(data, c)
}
