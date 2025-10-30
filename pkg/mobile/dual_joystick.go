package mobile

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// DualJoystickLayout implements dual virtual joysticks for dual-stick shooter mechanics.
// Left joystick controls movement (WASD equivalent), right joystick controls aim direction.
// This enables Phase 10.1's 360° rotation and independent movement/aim control on mobile.
type DualJoystickLayout struct {
	LeftJoystick  *VirtualJoystick // Movement control
	RightJoystick *VirtualJoystick // Aim control

	// Action buttons (attack, use item, etc.)
	ActionButtons []*VirtualButton

	// Configuration
	Visible      bool
	touchHandler *TouchInputHandler
	screenWidth  int
	screenHeight int
}

// NewDualJoystickLayout creates a dual joystick layout optimized for Phase 10.1 controls.
// Left joystick: bottom-left for movement (thumb position)
// Right joystick: bottom-right for aiming (thumb position)
// Action buttons: positioned for easy thumb reach
func NewDualJoystickLayout(screenWidth, screenHeight int) *DualJoystickLayout {
	// Calculate responsive sizes based on screen dimensions
	joystickRadius := float64(screenHeight) * 0.12  // 12% of screen height
	buttonRadius := float64(screenHeight) * 0.06    // 6% of screen height
	margin := float64(screenHeight) * 0.04          // 4% margin

	// Left joystick (movement) - bottom-left corner
	leftX := margin + joystickRadius
	leftY := float64(screenHeight) - margin - joystickRadius

	// Right joystick (aim) - bottom-right corner
	rightX := float64(screenWidth) - margin - joystickRadius
	rightY := float64(screenHeight) - margin - joystickRadius

	layout := &DualJoystickLayout{
		LeftJoystick:  NewVirtualJoystick(leftX, leftY, joystickRadius, JoystickTypeMovement),
		RightJoystick: NewVirtualJoystick(rightX, rightY, joystickRadius, JoystickTypeAim),
		Visible:       true,
		touchHandler:  NewTouchInputHandler(),
		screenWidth:   screenWidth,
		screenHeight:  screenHeight,
	}

	// Create action buttons positioned above right joystick for easy reach
	buttonYOffset := rightY - joystickRadius - margin - buttonRadius
	layout.ActionButtons = []*VirtualButton{
		NewVirtualButton(rightX, buttonYOffset, buttonRadius, "⚔"), // Attack
		NewVirtualButton(rightX-buttonRadius*2.5, buttonYOffset, buttonRadius*0.8, "E"), // Use item
	}

	return layout
}

// Update processes touch input for both joysticks and buttons.
// Automatically handles multi-touch - each joystick captures one touch.
func (l *DualJoystickLayout) Update() {
	if !l.Visible {
		return
	}

	// Update touch handler to get all active touches
	l.touchHandler.Update()
	touches := make(map[ebiten.TouchID]*Touch)
	for _, touch := range l.touchHandler.GetActiveTouches() {
		touches[touch.ID] = touch
	}

	// Update joysticks (they handle their own touch capture)
	l.LeftJoystick.Update(touches)
	l.RightJoystick.Update(touches)

	// Update action buttons
	for _, button := range l.ActionButtons {
		button.Update(touches)
	}
}

// Draw renders both joysticks and action buttons.
func (l *DualJoystickLayout) Draw(screen *ebiten.Image) {
	if !l.Visible {
		return
	}

	l.LeftJoystick.Draw(screen)
	l.RightJoystick.Draw(screen)

	for _, button := range l.ActionButtons {
		button.Draw(screen)
	}
}

// GetMovementDirection returns normalized movement direction from left joystick.
// Returns (0, 0) when joystick is centered or inactive.
func (l *DualJoystickLayout) GetMovementDirection() (float64, float64) {
	return l.LeftJoystick.GetDirection()
}

// GetAimDirection returns normalized aim direction from right joystick.
// Returns (0, 0) when joystick is centered or inactive.
func (l *DualJoystickLayout) GetAimDirection() (float64, float64) {
	return l.RightJoystick.GetDirection()
}

// GetAimAngle returns the aim angle in radians (0=right, π/2=down, π=left, 3π/2=up).
// Returns current angle even if joystick is inactive (maintains last aim direction).
func (l *DualJoystickLayout) GetAimAngle() float64 {
	return l.RightJoystick.GetAngle()
}

// IsMoving returns true if the left joystick is actively being touched.
func (l *DualJoystickLayout) IsMoving() bool {
	return l.LeftJoystick.IsActive()
}

// IsAiming returns true if the right joystick is actively being touched.
func (l *DualJoystickLayout) IsAiming() bool {
	return l.RightJoystick.IsActive()
}

// IsAttackPressed returns true when attack button is pressed (one frame).
func (l *DualJoystickLayout) IsAttackPressed() bool {
	return len(l.ActionButtons) > 0 && l.ActionButtons[0].IsPressed()
}

// IsUsePressed returns true when use item button is pressed (one frame).
func (l *DualJoystickLayout) IsUsePressed() bool {
	return len(l.ActionButtons) > 1 && l.ActionButtons[1].IsPressed()
}

// SetVisible controls whether the dual joystick layout is shown and active.
func (l *DualJoystickLayout) SetVisible(visible bool) {
	l.Visible = visible
}

// JoystickType defines the purpose of a virtual joystick.
type JoystickType int

const (
	JoystickTypeMovement JoystickType = iota // Left joystick for WASD movement
	JoystickTypeAim                           // Right joystick for mouse aim
)

// VirtualJoystick represents a single virtual joystick with analog input.
// Supports both floating and fixed joystick modes:
// - Floating: joystick appears where user touches
// - Fixed: joystick stays in one position
type VirtualJoystick struct {
	// Configuration
	Type          JoystickType
	X, Y          float64 // Center position (base position for floating mode)
	Radius        float64 // Outer boundary radius
	DeadZone      float64 // Inner dead zone radius (no input)
	FloatingMode  bool    // If true, joystick appears at touch position

	// Current state
	TouchID      ebiten.TouchID
	Active       bool
	CurrentX     float64 // Current center (for floating mode)
	CurrentY     float64 // Current center (for floating mode)
	DirectionX   float64 // -1.0 to 1.0
	DirectionY   float64 // -1.0 to 1.0
	Angle        float64 // Current angle in radians (0=right, π/2=down)
	Magnitude    float64 // 0.0 to 1.0 (distance from center)

	// Visual settings
	BaseColor    color.Color // Base circle color
	StickColor   color.Color // Stick/thumb color
	ActiveColor  color.Color // Color when active
	DeadZoneColor color.Color // Dead zone indicator
	Opacity      float64
}

// NewVirtualJoystick creates a virtual joystick at the specified position.
func NewVirtualJoystick(x, y, radius float64, joystickType JoystickType) *VirtualJoystick {
	// Color coding by type for visual distinction
	var baseColor, activeColor color.Color
	if joystickType == JoystickTypeMovement {
		// Blue tint for movement joystick
		baseColor = color.RGBA{80, 80, 120, 160}
		activeColor = color.RGBA{100, 100, 200, 220}
	} else {
		// Red tint for aim joystick
		baseColor = color.RGBA{120, 80, 80, 160}
		activeColor = color.RGBA{200, 100, 100, 220}
	}

	return &VirtualJoystick{
		Type:          joystickType,
		X:             x,
		Y:             y,
		Radius:        radius,
		DeadZone:      radius * 0.2,  // 20% dead zone
		FloatingMode:  false,         // Fixed by default
		TouchID:       -1,
		CurrentX:      x,
		CurrentY:      y,
		BaseColor:     baseColor,
		StickColor:    color.RGBA{180, 180, 180, 240},
		ActiveColor:   activeColor,
		DeadZoneColor: color.RGBA{40, 40, 40, 100},
		Opacity:       0.6,
	}
}

// Update processes touch input for the joystick.
// Handles touch capture, direction calculation, and angle/magnitude updates.
func (j *VirtualJoystick) Update(touches map[ebiten.TouchID]*Touch) {
	// Check if we have an active touch
	if j.TouchID >= 0 {
		if touch, exists := touches[j.TouchID]; exists && touch.Active {
			// Update joystick direction based on touch position
			j.updateDirection(float64(touch.X), float64(touch.Y))
			j.Active = true
			return
		} else {
			// Touch released - reset joystick
			j.TouchID = -1
			j.Active = false
			j.DirectionX = 0
			j.DirectionY = 0
			j.Magnitude = 0
			// Keep last angle for aim joystick (maintains aim direction)
			if j.FloatingMode {
				j.CurrentX = j.X
				j.CurrentY = j.Y
			}
			return
		}
	}

	// Look for new touch in joystick area
	for id, touch := range touches {
		if !touch.Active {
			continue
		}

		// Check if touch is within joystick capture area
		dx := float64(touch.StartX) - j.X
		dy := float64(touch.StartY) - j.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		// Capture area is 1.5x radius for easier touch
		if distance <= j.Radius*1.5 {
			j.TouchID = id
			j.Active = true

			// For floating mode, move joystick center to touch start
			if j.FloatingMode {
				j.CurrentX = float64(touch.StartX)
				j.CurrentY = float64(touch.StartY)
			}

			// Initial direction from current position
			j.updateDirection(float64(touch.X), float64(touch.Y))
			break
		}
	}
}

// updateDirection calculates direction, angle, and magnitude from touch position.
func (j *VirtualJoystick) updateDirection(touchX, touchY float64) {
	// Calculate offset from joystick center
	dx := touchX - j.CurrentX
	dy := touchY - j.CurrentY
	distance := math.Sqrt(dx*dx + dy*dy)

	// Apply dead zone
	if distance < j.DeadZone {
		j.DirectionX = 0
		j.DirectionY = 0
		j.Magnitude = 0
		return
	}

	// Calculate angle (atan2 returns -π to π, convert to 0 to 2π)
	angle := math.Atan2(dy, dx)
	if angle < 0 {
		angle += 2 * math.Pi
	}
	j.Angle = angle

	// Calculate magnitude (0.0 to 1.0)
	// Clamp distance to radius
	if distance > j.Radius {
		distance = j.Radius
	}
	j.Magnitude = (distance - j.DeadZone) / (j.Radius - j.DeadZone)

	// Calculate normalized direction (-1.0 to 1.0)
	j.DirectionX = (dx / j.Radius) * math.Max(1.0, j.Magnitude)
	j.DirectionY = (dy / j.Radius) * math.Max(1.0, j.Magnitude)

	// Clamp to [-1.0, 1.0] range
	j.DirectionX = math.Max(-1.0, math.Min(1.0, j.DirectionX))
	j.DirectionY = math.Max(-1.0, math.Min(1.0, j.DirectionY))
}

// GetDirection returns the normalized direction vector.
func (j *VirtualJoystick) GetDirection() (float64, float64) {
	return j.DirectionX, j.DirectionY
}

// GetAngle returns the current angle in radians.
// 0 = right, π/2 = down, π = left, 3π/2 = up
func (j *VirtualJoystick) GetAngle() float64 {
	return j.Angle
}

// GetMagnitude returns the current magnitude (0.0 to 1.0).
func (j *VirtualJoystick) GetMagnitude() float64 {
	return j.Magnitude
}

// IsActive returns true if the joystick is currently being touched.
func (j *VirtualJoystick) IsActive() bool {
	return j.Active
}

// Draw renders the joystick on screen.
func (j *VirtualJoystick) Draw(screen *ebiten.Image) {
	// Draw base circle (outer boundary)
	baseColor := j.BaseColor
	if j.Active {
		baseColor = j.ActiveColor
	}
	vector.DrawFilledCircle(screen, float32(j.CurrentX), float32(j.CurrentY), float32(j.Radius), baseColor, true)

	// Draw dead zone indicator
	vector.DrawFilledCircle(screen, float32(j.CurrentX), float32(j.CurrentY), float32(j.DeadZone), j.DeadZoneColor, true)

	// Draw stick position (shows current input)
	stickX := j.CurrentX + j.DirectionX*j.Radius*0.6
	stickY := j.CurrentY + j.DirectionY*j.Radius*0.6
	stickRadius := j.Radius * 0.4
	vector.DrawFilledCircle(screen, float32(stickX), float32(stickY), float32(stickRadius), j.StickColor, true)

	// Draw directional indicator line (from center to stick)
	if j.Active && j.Magnitude > 0.1 {
		indicatorColor := color.RGBA{255, 255, 255, 180}
		vector.StrokeLine(screen,
			float32(j.CurrentX), float32(j.CurrentY),
			float32(stickX), float32(stickY),
			3, indicatorColor, true)
	}

	// Draw border
	borderColor := color.RGBA{200, 200, 200, 200}
	vector.StrokeCircle(screen, float32(j.CurrentX), float32(j.CurrentY), float32(j.Radius), 2, borderColor, true)
}
