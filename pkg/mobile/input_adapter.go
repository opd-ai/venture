package mobile

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// MobileInputAdapter bridges mobile touch controls to the engine InputProvider interface.
// It wraps DualJoystickLayout to provide a unified input abstraction compatible with
// the desktop EbitenInput implementation while supporting mobile-specific touch controls.
type MobileInputAdapter struct {
	layout *DualJoystickLayout

	// Track "just pressed" states for single-frame actions
	lastActionState bool
	lastUseState    bool
	lastMenuState   bool

	// Screen dimensions for mouse position simulation
	screenWidth  int
	screenHeight int

	// Last aim position for mouse position queries
	lastAimX int
	lastAimY int
}

// NewMobileInputAdapter creates an input adapter wrapping the mobile dual joystick layout.
func NewMobileInputAdapter(screenWidth, screenHeight int) *MobileInputAdapter {
	return &MobileInputAdapter{
		layout:       NewDualJoystickLayout(screenWidth, screenHeight),
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		lastAimX:     screenWidth / 2,
		lastAimY:     screenHeight / 2,
	}
}

// Update processes touch input and updates internal state.
// Must be called once per frame before querying input state.
func (m *MobileInputAdapter) Update() {
	m.layout.Update()
}

// Draw renders the mobile controls on screen.
func (m *MobileInputAdapter) Draw(screen *ebiten.Image) {
	m.layout.Draw(screen)
}

// Type implements Component interface.
func (m *MobileInputAdapter) Type() string {
	return "input"
}

// GetMovement implements InputProvider interface.
// Returns normalized movement input from left joystick.
func (m *MobileInputAdapter) GetMovement() (x, y float64) {
	return m.layout.GetMovementDirection()
}

// IsActionPressed implements InputProvider interface.
// Returns true when attack button is currently pressed.
func (m *MobileInputAdapter) IsActionPressed() bool {
	return m.layout.IsAttackPressed()
}

// IsActionJustPressed implements InputProvider interface.
// Returns true only on the frame when attack button was first pressed.
func (m *MobileInputAdapter) IsActionJustPressed() bool {
	currentState := m.layout.IsAttackPressed()
	justPressed := currentState && !m.lastActionState
	m.lastActionState = currentState
	return justPressed
}

// IsAnyKeyPressed implements InputProvider interface.
// Returns true if any touch input is active.
func (m *MobileInputAdapter) IsAnyKeyPressed() bool {
	return m.layout.IsMoving() || m.layout.IsAiming() || m.layout.IsAttackPressed() || m.layout.IsUsePressed()
}

// IsUseItemPressed implements InputProvider interface.
// Returns true when use item button is currently pressed.
func (m *MobileInputAdapter) IsUseItemPressed() bool {
	return m.layout.IsUsePressed()
}

// IsUseItemJustPressed implements InputProvider interface.
// Returns true only on the frame when use item button was first pressed.
func (m *MobileInputAdapter) IsUseItemJustPressed() bool {
	currentState := m.layout.IsUsePressed()
	justPressed := currentState && !m.lastUseState
	m.lastUseState = currentState
	return justPressed
}

// IsSpellPressed implements InputProvider interface.
// Mobile builds do not support hotkeys 1-5 (no spell buttons in dual joystick layout).
// Returns false for all slots.
func (m *MobileInputAdapter) IsSpellPressed(slot int) bool {
	// Dual joystick layout focuses on movement/aim/attack/use.
	// Spell casting on mobile would require separate UI or gesture controls.
	return false
}

// GetMousePosition implements InputProvider interface.
// Returns simulated mouse position based on right joystick aim direction.
// This allows aiming systems to work without modification on mobile.
func (m *MobileInputAdapter) GetMousePosition() (x, y int) {
	aimX, aimY := m.layout.GetAimDirection()

	// If aiming, convert aim direction to screen position
	if aimX != 0 || aimY != 0 {
		// Place cursor at edge of screen in aim direction
		centerX := m.screenWidth / 2
		centerY := m.screenHeight / 2
		radius := float64(m.screenHeight) / 3.0 // Aim range

		m.lastAimX = centerX + int(aimX*radius)
		m.lastAimY = centerY + int(aimY*radius)
	}

	return m.lastAimX, m.lastAimY
}

// GetMouseDelta implements InputProvider interface.
// Returns (0, 0) as mobile uses joystick aiming, not delta-based.
func (m *MobileInputAdapter) GetMouseDelta() (dx, dy int) {
	return 0, 0
}

// IsMousePressed implements InputProvider interface.
// Returns true when right joystick is actively being used for aiming.
func (m *MobileInputAdapter) IsMousePressed() bool {
	return m.layout.IsAiming()
}

// SetMovement implements InputProvider interface (for testing).
func (m *MobileInputAdapter) SetMovement(x, y float64) {
	// Not applicable for touch input - joystick state is managed by touch events
}

// SetActionPressed implements InputProvider interface (for testing).
func (m *MobileInputAdapter) SetActionPressed(pressed bool) {
	// Not applicable for touch input - button state is managed by touch events
}

// IsMenuUpJustPressed implements InputProvider interface.
// Mobile does not have arrow keys - swipe gestures would be needed for menu navigation.
func (m *MobileInputAdapter) IsMenuUpJustPressed() bool {
	return false
}

// IsMenuDownJustPressed implements InputProvider interface.
// Mobile does not have arrow keys - swipe gestures would be needed for menu navigation.
func (m *MobileInputAdapter) IsMenuDownJustPressed() bool {
	return false
}

// IsMenuConfirmJustPressed implements InputProvider interface.
// Maps to action button press for menu confirmation.
func (m *MobileInputAdapter) IsMenuConfirmJustPressed() bool {
	return m.IsActionJustPressed()
}

// IsMenuBackJustPressed implements InputProvider interface.
// Mobile uses two-finger tap or specific gesture for back/cancel.
// For now, returns false - dedicated back button would be needed in layout.
func (m *MobileInputAdapter) IsMenuBackJustPressed() bool {
	currentState := len(m.layout.ActionButtons) > 1 && m.layout.ActionButtons[1].IsPressed()
	justPressed := currentState && !m.lastMenuState
	m.lastMenuState = currentState
	return justPressed
}

// IsMenuTabJustPressed implements InputProvider interface.
// Mobile does not have tab key - not applicable.
func (m *MobileInputAdapter) IsMenuTabJustPressed() bool {
	return false
}

// SetVisible controls visibility of mobile controls.
func (m *MobileInputAdapter) SetVisible(visible bool) {
	m.layout.SetVisible(visible)
}

// GetLayout returns the underlying dual joystick layout for advanced control.
func (m *MobileInputAdapter) GetLayout() *DualJoystickLayout {
	return m.layout
}
