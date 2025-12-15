// Package engine provides the VR controller component for VR support.

package engine

import (
	"encoding/json"
	"math"
	"sync"
)

// VRControllerComponent tracks VR controller state for an entity.
// It manages button states, trigger values, thumbstick input, and haptic feedback.
type VRControllerComponent struct {
	mu sync.RWMutex

	// Enabled controls whether controller input is active
	Enabled bool `json:"enabled"`

	// Hand identifies which controller ("left" or "right")
	Hand string `json:"hand"`

	// TriggerValue is the analog trigger value (0.0-1.0)
	TriggerValue float64 `json:"trigger_value"`

	// GripValue is the analog grip button value (0.0-1.0)
	GripValue float64 `json:"grip_value"`

	// ThumbstickX is the horizontal thumbstick axis (-1.0 to 1.0)
	ThumbstickX float64 `json:"thumbstick_x"`

	// ThumbstickY is the vertical thumbstick axis (-1.0 to 1.0)
	ThumbstickY float64 `json:"thumbstick_y"`

	// ButtonA is the primary action button state
	ButtonA bool `json:"button_a"`

	// ButtonB is the secondary action button state
	ButtonB bool `json:"button_b"`

	// MenuButton is the menu/system button state
	MenuButton bool `json:"menu_button"`

	// ThumbstickPressed is true when thumbstick is clicked
	ThumbstickPressed bool `json:"thumbstick_pressed"`

	// HapticIntensity is the current haptic vibration level (0.0-1.0)
	HapticIntensity float64 `json:"haptic_intensity"`

	// HapticDuration is the remaining haptic duration in seconds
	HapticDuration float64 `json:"haptic_duration"`

	// DeadZone is the thumbstick dead zone threshold (0.0-0.5)
	DeadZone float64 `json:"dead_zone"`

	// Previous frame button states for edge detection
	prevButtonA    bool
	prevButtonB    bool
	prevMenuBtn    bool
	prevTrigger    bool
	prevGrip       bool
	prevThumbstick bool
}

const (
	// ControllerLeft identifies the left hand controller
	ControllerLeft = "left"

	// ControllerRight identifies the right hand controller
	ControllerRight = "right"

	// DefaultDeadZone is the default thumbstick dead zone
	DefaultDeadZone = 0.15

	// TriggerPressThreshold is the trigger value that counts as "pressed"
	TriggerPressThreshold = 0.5

	// GripPressThreshold is the grip value that counts as "pressed"
	GripPressThreshold = 0.5
)

// NewVRControllerComponent creates a new VR controller component.
func NewVRControllerComponent(hand string) *VRControllerComponent {
	if hand != ControllerLeft && hand != ControllerRight {
		hand = ControllerRight
	}
	return &VRControllerComponent{
		Enabled:  false,
		Hand:     hand,
		DeadZone: DefaultDeadZone,
	}
}

// Type returns the component type identifier.
func (c *VRControllerComponent) Type() string {
	return "vr_controller"
}

// SetEnabled enables or disables controller input.
func (c *VRControllerComponent) SetEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Enabled = enabled
}

// IsEnabled returns whether controller input is enabled.
func (c *VRControllerComponent) IsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Enabled
}

// GetHand returns which hand this controller is for.
func (c *VRControllerComponent) GetHand() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Hand
}

// SetTrigger sets the trigger value (clamped to 0.0-1.0).
func (c *VRControllerComponent) SetTrigger(value float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.prevTrigger = c.TriggerValue >= TriggerPressThreshold
	c.TriggerValue = clampFloat(value, 0.0, 1.0)
}

// GetTrigger returns the trigger value.
func (c *VRControllerComponent) GetTrigger() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.TriggerValue
}

// IsTriggerPressed returns true if trigger is past the press threshold.
func (c *VRControllerComponent) IsTriggerPressed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.TriggerValue >= TriggerPressThreshold
}

// IsTriggerJustPressed returns true if trigger was just pressed this frame.
func (c *VRControllerComponent) IsTriggerJustPressed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	pressed := c.TriggerValue >= TriggerPressThreshold
	return pressed && !c.prevTrigger
}

// SetGrip sets the grip value (clamped to 0.0-1.0).
func (c *VRControllerComponent) SetGrip(value float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.prevGrip = c.GripValue >= GripPressThreshold
	c.GripValue = clampFloat(value, 0.0, 1.0)
}

// GetGrip returns the grip value.
func (c *VRControllerComponent) GetGrip() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.GripValue
}

// IsGripPressed returns true if grip is past the press threshold.
func (c *VRControllerComponent) IsGripPressed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.GripValue >= GripPressThreshold
}

// SetThumbstick sets the thumbstick position (clamped to -1.0 to 1.0).
func (c *VRControllerComponent) SetThumbstick(x, y float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.prevThumbstick = c.ThumbstickPressed
	c.ThumbstickX = clampFloat(x, -1.0, 1.0)
	c.ThumbstickY = clampFloat(y, -1.0, 1.0)
}

// GetThumbstick returns the thumbstick position with dead zone applied.
func (c *VRControllerComponent) GetThumbstick() (x, y float64) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	x = c.ThumbstickX
	y = c.ThumbstickY

	// Apply dead zone
	if math.Abs(x) < c.DeadZone {
		x = 0
	}
	if math.Abs(y) < c.DeadZone {
		y = 0
	}

	return x, y
}

// GetThumbstickRaw returns the raw thumbstick position without dead zone.
func (c *VRControllerComponent) GetThumbstickRaw() (x, y float64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ThumbstickX, c.ThumbstickY
}

// SetThumbstickPressed sets whether the thumbstick is clicked.
func (c *VRControllerComponent) SetThumbstickPressed(pressed bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.prevThumbstick = c.ThumbstickPressed
	c.ThumbstickPressed = pressed
}

// IsThumbstickPressed returns whether the thumbstick is clicked.
func (c *VRControllerComponent) IsThumbstickPressed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ThumbstickPressed
}

// SetButtonA sets the A button state.
func (c *VRControllerComponent) SetButtonA(pressed bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.prevButtonA = c.ButtonA
	c.ButtonA = pressed
}

// IsButtonAPressed returns whether button A is pressed.
func (c *VRControllerComponent) IsButtonAPressed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ButtonA
}

// IsButtonAJustPressed returns true if button A was just pressed this frame.
func (c *VRControllerComponent) IsButtonAJustPressed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ButtonA && !c.prevButtonA
}

// SetButtonB sets the B button state.
func (c *VRControllerComponent) SetButtonB(pressed bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.prevButtonB = c.ButtonB
	c.ButtonB = pressed
}

// IsButtonBPressed returns whether button B is pressed.
func (c *VRControllerComponent) IsButtonBPressed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ButtonB
}

// IsButtonBJustPressed returns true if button B was just pressed this frame.
func (c *VRControllerComponent) IsButtonBJustPressed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ButtonB && !c.prevButtonB
}

// SetMenuButton sets the menu button state.
func (c *VRControllerComponent) SetMenuButton(pressed bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.prevMenuBtn = c.MenuButton
	c.MenuButton = pressed
}

// IsMenuButtonPressed returns whether the menu button is pressed.
func (c *VRControllerComponent) IsMenuButtonPressed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.MenuButton
}

// IsMenuButtonJustPressed returns true if menu button was just pressed this frame.
func (c *VRControllerComponent) IsMenuButtonJustPressed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.MenuButton && !c.prevMenuBtn
}

// SetHaptic triggers haptic feedback.
func (c *VRControllerComponent) SetHaptic(intensity, duration float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.HapticIntensity = clampFloat(intensity, 0.0, 1.0)
	c.HapticDuration = duration
	if c.HapticDuration < 0 {
		c.HapticDuration = 0
	}
}

// GetHaptic returns the current haptic state.
func (c *VRControllerComponent) GetHaptic() (intensity, duration float64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.HapticIntensity, c.HapticDuration
}

// UpdateHaptic decrements haptic duration by delta time.
func (c *VRControllerComponent) UpdateHaptic(deltaTime float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.HapticDuration -= deltaTime
	if c.HapticDuration <= 0 {
		c.HapticDuration = 0
		c.HapticIntensity = 0
	}
}

// SetDeadZone sets the thumbstick dead zone threshold.
func (c *VRControllerComponent) SetDeadZone(deadZone float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.DeadZone = clampFloat(deadZone, 0.0, 0.5)
}

// GetDeadZone returns the thumbstick dead zone threshold.
func (c *VRControllerComponent) GetDeadZone() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.DeadZone
}

// Serialize converts the component to JSON bytes.
func (c *VRControllerComponent) Serialize() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return json.Marshal(c)
}

// Deserialize loads component state from JSON bytes.
func (c *VRControllerComponent) Deserialize(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return json.Unmarshal(data, c)
}
