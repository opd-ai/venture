package mobile

import (
	"math"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestNewDualJoystickLayout verifies dual joystick layout creation.
func TestNewDualJoystickLayout(t *testing.T) {
	tests := []struct {
		name         string
		screenWidth  int
		screenHeight int
		wantVisible  bool
	}{
		{
			name:         "standard mobile screen",
			screenWidth:  1920,
			screenHeight: 1080,
			wantVisible:  true,
		},
		{
			name:         "small mobile screen",
			screenWidth:  1280,
			screenHeight: 720,
			wantVisible:  true,
		},
		{
			name:         "tablet screen",
			screenWidth:  2048,
			screenHeight: 1536,
			wantVisible:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout := NewDualJoystickLayout(tt.screenWidth, tt.screenHeight)

			// Verify layout created
			if layout == nil {
				t.Fatal("NewDualJoystickLayout returned nil")
			}

			// Verify joysticks created
			if layout.LeftJoystick == nil {
				t.Error("LeftJoystick is nil")
			}
			if layout.RightJoystick == nil {
				t.Error("RightJoystick is nil")
			}

			// Verify joystick types
			if layout.LeftJoystick != nil && layout.LeftJoystick.Type != JoystickTypeMovement {
				t.Errorf("LeftJoystick type = %v, want %v", layout.LeftJoystick.Type, JoystickTypeMovement)
			}
			if layout.RightJoystick != nil && layout.RightJoystick.Type != JoystickTypeAim {
				t.Errorf("RightJoystick type = %v, want %v", layout.RightJoystick.Type, JoystickTypeAim)
			}

			// Verify action buttons created
			if len(layout.ActionButtons) < 2 {
				t.Errorf("ActionButtons count = %d, want at least 2", len(layout.ActionButtons))
			}

			// Verify visibility
			if layout.Visible != tt.wantVisible {
				t.Errorf("Visible = %v, want %v", layout.Visible, tt.wantVisible)
			}

			// Verify touch handler created
			if layout.touchHandler == nil {
				t.Error("touchHandler is nil")
			}
		})
	}
}

// TestVirtualJoystickCreation verifies joystick creation and initialization.
func TestVirtualJoystickCreation(t *testing.T) {
	tests := []struct {
		name          string
		x, y, radius  float64
		joystickType  JoystickType
		wantActive    bool
		wantDirection [2]float64
	}{
		{
			name:          "movement joystick",
			x:             100,
			y:             500,
			radius:        80,
			joystickType:  JoystickTypeMovement,
			wantActive:    false,
			wantDirection: [2]float64{0, 0},
		},
		{
			name:          "aim joystick",
			x:             700,
			y:             500,
			radius:        80,
			joystickType:  JoystickTypeAim,
			wantActive:    false,
			wantDirection: [2]float64{0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			joystick := NewVirtualJoystick(tt.x, tt.y, tt.radius, tt.joystickType)

			// Verify creation
			if joystick == nil {
				t.Fatal("NewVirtualJoystick returned nil")
			}

			// Verify position
			if joystick.X != tt.x || joystick.Y != tt.y {
				t.Errorf("Position = (%v, %v), want (%v, %v)", joystick.X, joystick.Y, tt.x, tt.y)
			}

			// Verify radius
			if joystick.Radius != tt.radius {
				t.Errorf("Radius = %v, want %v", joystick.Radius, tt.radius)
			}

			// Verify type
			if joystick.Type != tt.joystickType {
				t.Errorf("Type = %v, want %v", joystick.Type, tt.joystickType)
			}

			// Verify initial state
			if joystick.Active != tt.wantActive {
				t.Errorf("Active = %v, want %v", joystick.Active, tt.wantActive)
			}

			dx, dy := joystick.GetDirection()
			if dx != tt.wantDirection[0] || dy != tt.wantDirection[1] {
				t.Errorf("Direction = (%v, %v), want (%v, %v)", dx, dy, tt.wantDirection[0], tt.wantDirection[1])
			}

			// Verify dead zone configured
			if joystick.DeadZone <= 0 {
				t.Errorf("DeadZone = %v, want > 0", joystick.DeadZone)
			}
			if joystick.DeadZone >= tt.radius {
				t.Errorf("DeadZone = %v, want < Radius (%v)", joystick.DeadZone, tt.radius)
			}
		})
	}
}

// TestVirtualJoystickDirection verifies direction calculation.
func TestVirtualJoystickDirection(t *testing.T) {
	tests := []struct {
		name          string
		centerX       float64
		centerY       float64
		radius        float64
		touchX        float64
		touchY        float64
		wantDirX      float64 // Expected direction X
		wantDirY      float64 // Expected direction Y
		wantMagnitude float64 // Expected magnitude (0-1)
		tolerance     float64 // Acceptable error
	}{
		{
			name:          "right direction",
			centerX:       100,
			centerY:       100,
			radius:        80,
			touchX:        180, // 80 pixels right
			touchY:        100,
			wantDirX:      1.0,
			wantDirY:      0.0,
			wantMagnitude: 1.0,
			tolerance:     0.1,
		},
		{
			name:          "down direction",
			centerX:       100,
			centerY:       100,
			radius:        80,
			touchX:        100,
			touchY:        180, // 80 pixels down
			wantDirX:      0.0,
			wantDirY:      1.0,
			wantMagnitude: 1.0,
			tolerance:     0.1,
		},
		{
			name:          "left direction",
			centerX:       100,
			centerY:       100,
			radius:        80,
			touchX:        20, // 80 pixels left
			touchY:        100,
			wantDirX:      -1.0,
			wantDirY:      0.0,
			wantMagnitude: 1.0,
			tolerance:     0.1,
		},
		{
			name:          "up direction",
			centerX:       100,
			centerY:       100,
			radius:        80,
			touchX:        100,
			touchY:        20, // 80 pixels up
			wantDirX:      0.0,
			wantDirY:      -1.0,
			wantMagnitude: 1.0,
			tolerance:     0.1,
		},
		{
			name:          "diagonal up-right",
			centerX:       100,
			centerY:       100,
			radius:        80,
			touchX:        156, // ~56 pixels right (45°)
			touchY:        44,  // ~56 pixels up
			wantDirX:      0.7,
			wantDirY:      -0.7,
			wantMagnitude: 1.0,
			tolerance:     0.15,
		},
		{
			name:          "half magnitude",
			centerX:       100,
			centerY:       100,
			radius:        80,
			touchX:        140, // 40 pixels right (half radius)
			touchY:        100,
			wantDirX:      0.5,
			wantDirY:      0.0,
			wantMagnitude: 0.5,
			tolerance:     0.15,
		},
		{
			name:          "dead zone - no input",
			centerX:       100,
			centerY:       100,
			radius:        80,
			touchX:        105, // 5 pixels right (within dead zone)
			touchY:        100,
			wantDirX:      0.0,
			wantDirY:      0.0,
			wantMagnitude: 0.0,
			tolerance:     0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			joystick := NewVirtualJoystick(tt.centerX, tt.centerY, tt.radius, JoystickTypeMovement)

			// Simulate touch
			touches := make(map[ebiten.TouchID]*Touch)
			touchID := ebiten.TouchID(1)
			touches[touchID] = &Touch{
				ID:        touchID,
				X:         int(tt.touchX),
				Y:         int(tt.touchY),
				StartX:    int(tt.centerX),
				StartY:    int(tt.centerY),
				StartTime: time.Now(),
				Active:    true,
			}

			// Update joystick
			joystick.Update(touches)

			// Verify direction
			dirX, dirY := joystick.GetDirection()
			if math.Abs(dirX-tt.wantDirX) > tt.tolerance {
				t.Errorf("DirectionX = %v, want %v (±%v)", dirX, tt.wantDirX, tt.tolerance)
			}
			if math.Abs(dirY-tt.wantDirY) > tt.tolerance {
				t.Errorf("DirectionY = %v, want %v (±%v)", dirY, tt.wantDirY, tt.tolerance)
			}

			// Verify magnitude
			magnitude := joystick.GetMagnitude()
			if math.Abs(magnitude-tt.wantMagnitude) > tt.tolerance {
				t.Errorf("Magnitude = %v, want %v (±%v)", magnitude, tt.wantMagnitude, tt.tolerance)
			}

			// Verify active state
			if !joystick.IsActive() && tt.wantMagnitude > 0 {
				t.Error("Joystick should be active with non-zero magnitude")
			}
		})
	}
}

// TestVirtualJoystickAngle verifies angle calculation.
func TestVirtualJoystickAngle(t *testing.T) {
	tests := []struct {
		name      string
		touchX    float64
		touchY    float64
		wantAngle float64 // Radians: 0=right, π/2=down, π=left, 3π/2=up
		tolerance float64
	}{
		{
			name:      "right (0°)",
			touchX:    180,
			touchY:    100,
			wantAngle: 0.0,
			tolerance: 0.1,
		},
		{
			name:      "down (90°)",
			touchX:    100,
			touchY:    180,
			wantAngle: math.Pi / 2,
			tolerance: 0.1,
		},
		{
			name:      "left (180°)",
			touchX:    20,
			touchY:    100,
			wantAngle: math.Pi,
			tolerance: 0.1,
		},
		{
			name:      "up (270°)",
			touchX:    100,
			touchY:    20,
			wantAngle: 3 * math.Pi / 2,
			tolerance: 0.1,
		},
		{
			name:      "diagonal down-right (45°)",
			touchX:    156,
			touchY:    156,
			wantAngle: math.Pi / 4,
			tolerance: 0.1,
		},
	}

	joystick := NewVirtualJoystick(100, 100, 80, JoystickTypeAim)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate touch
			touches := make(map[ebiten.TouchID]*Touch)
			touchID := ebiten.TouchID(1)
			touches[touchID] = &Touch{
				ID:        touchID,
				X:         int(tt.touchX),
				Y:         int(tt.touchY),
				StartX:    100,
				StartY:    100,
				StartTime: time.Now(),
				Active:    true,
			}

			// Update joystick
			joystick.Update(touches)

			// Get angle
			angle := joystick.GetAngle()

			// Verify angle
			if math.Abs(angle-tt.wantAngle) > tt.tolerance {
				t.Errorf("Angle = %v (%.1f°), want %v (%.1f°)",
					angle, angle*180/math.Pi,
					tt.wantAngle, tt.wantAngle*180/math.Pi)
			}
		})
	}
}

// TestDualJoystickIndependence verifies that both joysticks work independently.
func TestDualJoystickIndependence(t *testing.T) {
	layout := NewDualJoystickLayout(1920, 1080)

	// Simulate two touches - one on each joystick
	touches := make(map[ebiten.TouchID]*Touch)

	// Left joystick touch (moving right)
	leftTouch := ebiten.TouchID(1)
	touches[leftTouch] = &Touch{
		ID:        leftTouch,
		X:         int(layout.LeftJoystick.X + 60), // Right of center
		Y:         int(layout.LeftJoystick.Y),
		StartX:    int(layout.LeftJoystick.X),
		StartY:    int(layout.LeftJoystick.Y),
		StartTime: time.Now(),
		Active:    true,
	}

	// Right joystick touch (aiming up)
	rightTouch := ebiten.TouchID(2)
	touches[rightTouch] = &Touch{
		ID:        rightTouch,
		X:         int(layout.RightJoystick.X),
		Y:         int(layout.RightJoystick.Y - 60), // Above center
		StartX:    int(layout.RightJoystick.X),
		StartY:    int(layout.RightJoystick.Y),
		StartTime: time.Now(),
		Active:    true,
	}

	// Update joysticks
	layout.LeftJoystick.Update(touches)
	layout.RightJoystick.Update(touches)

	// Verify left joystick (movement) is right
	moveX, moveY := layout.GetMovementDirection()
	if moveX <= 0 {
		t.Errorf("Movement X = %v, want > 0 (moving right)", moveX)
	}
	if math.Abs(moveY) > 0.1 {
		t.Errorf("Movement Y = %v, want ~0 (no vertical movement)", moveY)
	}

	// Verify right joystick (aim) is up
	aimX, aimY := layout.GetAimDirection()
	if math.Abs(aimX) > 0.1 {
		t.Errorf("Aim X = %v, want ~0 (no horizontal aim)", aimX)
	}
	if aimY >= 0 {
		t.Errorf("Aim Y = %v, want < 0 (aiming up)", aimY)
	}

	// Verify both active
	if !layout.IsMoving() {
		t.Error("IsMoving should be true")
	}
	if !layout.IsAiming() {
		t.Error("IsAiming should be true")
	}
}

// TestDualJoystickLayout_SetVisible verifies visibility control.
func TestDualJoystickLayout_SetVisible(t *testing.T) {
	layout := NewDualJoystickLayout(1920, 1080)

	// Initially visible
	if !layout.Visible {
		t.Error("Layout should be visible by default")
	}

	// Hide
	layout.SetVisible(false)
	if layout.Visible {
		t.Error("SetVisible(false) should hide layout")
	}

	// Show
	layout.SetVisible(true)
	if !layout.Visible {
		t.Error("SetVisible(true) should show layout")
	}
}

// TestDualJoystickLayout_ActionButtons verifies action button functionality.
func TestDualJoystickLayout_ActionButtons(t *testing.T) {
	layout := NewDualJoystickLayout(1920, 1080)

	// Verify buttons exist
	if len(layout.ActionButtons) < 2 {
		t.Fatalf("Expected at least 2 action buttons, got %d", len(layout.ActionButtons))
	}

	// Simulate attack button press
	attackButton := layout.ActionButtons[0]
	touches := make(map[ebiten.TouchID]*Touch)
	touchID := ebiten.TouchID(10)

	// Touch starts
	touches[touchID] = &Touch{
		ID:        touchID,
		X:         int(attackButton.X),
		Y:         int(attackButton.Y),
		StartX:    int(attackButton.X),
		StartY:    int(attackButton.Y),
		StartTime: time.Now(),
		Active:    true,
	}
	attackButton.Update(touches)

	if !attackButton.IsActive() {
		t.Error("Attack button should be active when touched")
	}

	// Touch releases
	touches[touchID].Active = false
	attackButton.Update(touches)

	if !attackButton.IsPressed() {
		t.Error("Attack button should register press on release")
	}

	// Next frame
	delete(touches, touchID)
	attackButton.Update(touches)

	if attackButton.IsPressed() {
		t.Error("Attack button should not be pressed in next frame")
	}
}

// TestVirtualJoystickTouchCapture verifies that joystick captures nearby touches.
func TestVirtualJoystickTouchCapture(t *testing.T) {
	joystick := NewVirtualJoystick(100, 100, 80, JoystickTypeMovement)

	tests := []struct {
		name        string
		touchX      int
		touchY      int
		shouldCapture bool
	}{
		{
			name:          "touch at center",
			touchX:        100,
			touchY:        100,
			shouldCapture: true,
		},
		{
			name:          "touch within radius",
			touchX:        150,
			touchY:        100,
			shouldCapture: true,
		},
		{
			name:          "touch within 1.5x radius",
			touchX:        210, // 110 pixels from center (80 * 1.5 = 120)
			touchY:        100,
			shouldCapture: true,
		},
		{
			name:          "touch beyond capture area",
			touchX:        250, // 150 pixels from center
			touchY:        100,
			shouldCapture: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset joystick
			joystick.TouchID = -1
			joystick.Active = false

			// Simulate touch
			touches := make(map[ebiten.TouchID]*Touch)
			touchID := ebiten.TouchID(1)
			touches[touchID] = &Touch{
				ID:        touchID,
				X:         tt.touchX,
				Y:         tt.touchY,
				StartX:    tt.touchX,
				StartY:    tt.touchY,
				StartTime: time.Now(),
				Active:    true,
			}

			// Update joystick
			joystick.Update(touches)

			// Verify capture
			if joystick.IsActive() != tt.shouldCapture {
				t.Errorf("IsActive = %v, want %v", joystick.IsActive(), tt.shouldCapture)
			}
		})
	}
}

// TestVirtualJoystickMaintainsAimDirection verifies aim joystick maintains last angle when released.
func TestVirtualJoystickMaintainsAimDirection(t *testing.T) {
	joystick := NewVirtualJoystick(100, 100, 80, JoystickTypeAim)

	// Simulate aim to the right
	touches := make(map[ebiten.TouchID]*Touch)
	touchID := ebiten.TouchID(1)
	touches[touchID] = &Touch{
		ID:        touchID,
		X:         180,
		Y:         100,
		StartX:    100,
		StartY:    100,
		StartTime: time.Now(),
		Active:    true,
	}

	joystick.Update(touches)

	// Get initial angle (should be ~0 radians = right)
	angleWhenActive := joystick.GetAngle()
	if math.Abs(angleWhenActive-0.0) > 0.1 {
		t.Fatalf("Initial angle = %v, want ~0.0 (right)", angleWhenActive)
	}

	// Release touch
	touches[touchID].Active = false
	joystick.Update(touches)

	// Angle should be maintained even when inactive
	angleAfterRelease := joystick.GetAngle()
	if math.Abs(angleAfterRelease-angleWhenActive) > 0.01 {
		t.Errorf("Angle after release = %v, want %v (maintained)", angleAfterRelease, angleWhenActive)
	}

	// But direction should be zero (no active input)
	dirX, dirY := joystick.GetDirection()
	if dirX != 0 || dirY != 0 {
		t.Errorf("Direction = (%v, %v), want (0, 0) when inactive", dirX, dirY)
	}
}
