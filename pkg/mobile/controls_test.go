package mobile

import (
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestNewVirtualDPad tests D-pad creation.
func TestNewVirtualDPad(t *testing.T) {
	dpad := NewVirtualDPad(100, 200, 50)

	if dpad == nil {
		t.Fatal("NewVirtualDPad returned nil")
	}
	if dpad.X != 100 {
		t.Errorf("dpad.X = %.1f, want 100.0", dpad.X)
	}
	if dpad.Y != 200 {
		t.Errorf("dpad.Y = %.1f, want 200.0", dpad.Y)
	}
	if dpad.Radius != 50 {
		t.Errorf("dpad.Radius = %.1f, want 50.0", dpad.Radius)
	}
	if dpad.InnerRadius != 15 { // 50 * 0.3
		t.Errorf("dpad.InnerRadius = %.1f, want 15.0", dpad.InnerRadius)
	}
	if dpad.TouchID != -1 {
		t.Errorf("dpad.TouchID = %d, want -1", dpad.TouchID)
	}
	if dpad.Active {
		t.Error("dpad.Active = true, want false initially")
	}
}

// TestVirtualDPad_GetDirection tests direction retrieval.
func TestVirtualDPad_GetDirection(t *testing.T) {
	dpad := NewVirtualDPad(100, 200, 50)

	// Initially should be (0, 0)
	x, y := dpad.GetDirection()
	if x != 0 || y != 0 {
		t.Errorf("GetDirection() = (%.1f, %.1f), want (0.0, 0.0)", x, y)
	}

	// Set direction manually for testing
	dpad.DirectionX = 0.5
	dpad.DirectionY = -0.8

	x, y = dpad.GetDirection()
	if x != 0.5 {
		t.Errorf("GetDirection() x = %.1f, want 0.5", x)
	}
	if y != -0.8 {
		t.Errorf("GetDirection() y = %.1f, want -0.8", y)
	}
}

// TestVirtualDPad_IsActive tests active state.
func TestVirtualDPad_IsActive(t *testing.T) {
	dpad := NewVirtualDPad(100, 200, 50)

	// Initially not active
	if dpad.IsActive() {
		t.Error("IsActive() = true, want false initially")
	}

	// Set active manually
	dpad.Active = true
	if !dpad.IsActive() {
		t.Error("IsActive() = false, want true after setting")
	}
}

// TestVirtualDPad_Update tests D-pad touch processing.
func TestVirtualDPad_Update(t *testing.T) {
	dpad := NewVirtualDPad(100, 200, 50)
	touches := make(map[ebiten.TouchID]*Touch)

	// Update with no touches
	dpad.Update(touches)
	if dpad.IsActive() {
		t.Error("D-pad should not be active with no touches")
	}

	// Add touch within D-pad area
	touches[0] = &Touch{
		ID:        0,
		X:         110, // Within 50 pixel radius
		Y:         210,
		StartTime: time.Now(),
		Active:    true,
	}

	dpad.Update(touches)
	if !dpad.IsActive() {
		t.Error("D-pad should be active with touch in area")
	}
	if dpad.TouchID != 0 {
		t.Errorf("TouchID = %d, want 0", dpad.TouchID)
	}
}

// TestVirtualDPad_UpdateTouchRelease tests touch release handling.
func TestVirtualDPad_UpdateTouchRelease(t *testing.T) {
	dpad := NewVirtualDPad(100, 200, 50)
	touches := make(map[ebiten.TouchID]*Touch)

	// Activate D-pad
	touches[0] = &Touch{ID: 0, X: 110, Y: 210, Active: true, StartTime: time.Now()}
	dpad.Update(touches)

	if !dpad.IsActive() {
		t.Error("D-pad should be active")
	}

	// Release touch
	delete(touches, 0)
	dpad.Update(touches)

	if dpad.IsActive() {
		t.Error("D-pad should not be active after touch release")
	}
	if dpad.TouchID != -1 {
		t.Errorf("TouchID = %d, want -1 after release", dpad.TouchID)
	}
}

// TestVirtualDPad_UpdateOutsideArea tests touches outside D-pad area.
func TestVirtualDPad_UpdateOutsideArea(t *testing.T) {
	dpad := NewVirtualDPad(100, 200, 50)
	touches := make(map[ebiten.TouchID]*Touch)

	// Touch far from D-pad
	touches[0] = &Touch{ID: 0, X: 500, Y: 500, Active: true, StartTime: time.Now()}
	dpad.Update(touches)

	if dpad.IsActive() {
		t.Error("D-pad should not activate for touch outside area")
	}
}

// TestNewVirtualButton tests button creation.
func TestNewVirtualButton(t *testing.T) {
	button := NewVirtualButton(300, 400, 30, "A")

	if button == nil {
		t.Fatal("NewVirtualButton returned nil")
	}
	if button.X != 300 {
		t.Errorf("button.X = %.1f, want 300.0", button.X)
	}
	if button.Y != 400 {
		t.Errorf("button.Y = %.1f, want 400.0", button.Y)
	}
	if button.Radius != 30 {
		t.Errorf("button.Radius = %.1f, want 30.0", button.Radius)
	}
	if button.Label != "A" {
		t.Errorf("button.Label = %q, want %q", button.Label, "A")
	}
	if button.TouchID != -1 {
		t.Errorf("button.TouchID = %d, want -1", button.TouchID)
	}
	if button.Active {
		t.Error("button.Active = true, want false initially")
	}
	if button.Pressed {
		t.Error("button.Pressed = true, want false initially")
	}
}

// TestVirtualButton_IsPressed tests pressed state.
func TestVirtualButton_IsPressed(t *testing.T) {
	button := NewVirtualButton(300, 400, 30, "A")

	if button.IsPressed() {
		t.Error("IsPressed() = true, want false initially")
	}

	// Set pressed manually
	button.Pressed = true
	if !button.IsPressed() {
		t.Error("IsPressed() = false, want true after setting")
	}
}

// TestVirtualButton_IsActive tests active state.
func TestVirtualButton_IsActive(t *testing.T) {
	button := NewVirtualButton(300, 400, 30, "A")

	if button.IsActive() {
		t.Error("IsActive() = true, want false initially")
	}

	button.Active = true
	if !button.IsActive() {
		t.Error("IsActive() = false, want true after setting")
	}
}

// TestVirtualButton_Update tests button touch processing.
func TestVirtualButton_Update(t *testing.T) {
	button := NewVirtualButton(300, 400, 30, "A")
	touches := make(map[ebiten.TouchID]*Touch)

	// Update with no touches
	button.Update(touches)
	if button.IsActive() {
		t.Error("Button should not be active with no touches")
	}
	if button.IsPressed() {
		t.Error("Button should not be pressed with no touches")
	}

	// Add touch within button area
	touches[0] = &Touch{
		ID:        0,
		X:         310, // Within 30 pixel radius
		Y:         410,
		StartTime: time.Now(),
		Active:    true,
	}

	button.Update(touches)
	if !button.IsActive() {
		t.Error("Button should be active with touch in area")
	}
	if button.TouchID != 0 {
		t.Errorf("TouchID = %d, want 0", button.TouchID)
	}
}

// TestVirtualButton_UpdatePress tests button press detection.
func TestVirtualButton_UpdatePress(t *testing.T) {
	button := NewVirtualButton(300, 400, 30, "A")
	touches := make(map[ebiten.TouchID]*Touch)

	// Activate button
	touches[0] = &Touch{ID: 0, X: 310, Y: 410, Active: true, StartTime: time.Now()}
	button.Update(touches)

	if !button.IsActive() {
		t.Error("Button should be active")
	}
	if button.IsPressed() {
		t.Error("Button should not be pressed yet (touch not released)")
	}

	// Release touch - should trigger press
	delete(touches, 0)
	button.Update(touches)

	if button.IsActive() {
		t.Error("Button should not be active after release")
	}
	if !button.IsPressed() {
		t.Error("Button should be pressed on touch release")
	}

	// Next update should clear pressed state
	button.Update(touches)
	if button.IsPressed() {
		t.Error("Button pressed state should clear after one frame")
	}
}

// TestVirtualButton_UpdateOutsideArea tests touches outside button area.
func TestVirtualButton_UpdateOutsideArea(t *testing.T) {
	button := NewVirtualButton(300, 400, 30, "A")
	touches := make(map[ebiten.TouchID]*Touch)

	// Touch far from button
	touches[0] = &Touch{ID: 0, X: 500, Y: 500, Active: true, StartTime: time.Now()}
	button.Update(touches)

	if button.IsActive() {
		t.Error("Button should not activate for touch outside area")
	}
}

// TestNewVirtualControlsLayout tests layout creation.
func TestNewVirtualControlsLayout(t *testing.T) {
	layout := NewVirtualControlsLayout(800, 600)

	if layout == nil {
		t.Fatal("NewVirtualControlsLayout returned nil")
	}
	if layout.DPad == nil {
		t.Error("layout.DPad is nil")
	}
	if layout.ActionButton == nil {
		t.Error("layout.ActionButton is nil")
	}
	if layout.SecondaryButton == nil {
		t.Error("layout.SecondaryButton is nil")
	}
	if layout.MenuButton == nil {
		t.Error("layout.MenuButton is nil")
	}
	if !layout.Visible {
		t.Error("layout.Visible = false, want true initially")
	}
	if layout.touchHandler == nil {
		t.Error("layout.touchHandler is nil")
	}
}

// TestVirtualControlsLayout_GetMovementInput tests movement input retrieval.
func TestVirtualControlsLayout_GetMovementInput(t *testing.T) {
	layout := NewVirtualControlsLayout(800, 600)

	// Initially should be (0, 0)
	x, y := layout.GetMovementInput()
	if x != 0 || y != 0 {
		t.Errorf("GetMovementInput() = (%.1f, %.1f), want (0.0, 0.0)", x, y)
	}

	// Set D-pad direction manually
	layout.DPad.DirectionX = 1.0
	layout.DPad.DirectionY = -0.5

	x, y = layout.GetMovementInput()
	if x != 1.0 {
		t.Errorf("GetMovementInput() x = %.1f, want 1.0", x)
	}
	if y != -0.5 {
		t.Errorf("GetMovementInput() y = %.1f, want -0.5", y)
	}
}

// TestVirtualControlsLayout_IsActionPressed tests action button press detection.
func TestVirtualControlsLayout_IsActionPressed(t *testing.T) {
	layout := NewVirtualControlsLayout(800, 600)

	if layout.IsActionPressed() {
		t.Error("IsActionPressed() = true, want false initially")
	}

	// Simulate button press
	layout.ActionButton.Pressed = true

	if !layout.IsActionPressed() {
		t.Error("IsActionPressed() = false, want true after press")
	}
}

// TestVirtualControlsLayout_IsSecondaryPressed tests secondary button press detection.
func TestVirtualControlsLayout_IsSecondaryPressed(t *testing.T) {
	layout := NewVirtualControlsLayout(800, 600)

	if layout.IsSecondaryPressed() {
		t.Error("IsSecondaryPressed() = true, want false initially")
	}

	layout.SecondaryButton.Pressed = true

	if !layout.IsSecondaryPressed() {
		t.Error("IsSecondaryPressed() = false, want true after press")
	}
}

// TestVirtualControlsLayout_IsMenuPressed tests menu button press detection.
func TestVirtualControlsLayout_IsMenuPressed(t *testing.T) {
	layout := NewVirtualControlsLayout(800, 600)

	if layout.IsMenuPressed() {
		t.Error("IsMenuPressed() = true, want false initially")
	}

	layout.MenuButton.Pressed = true

	if !layout.IsMenuPressed() {
		t.Error("IsMenuPressed() = false, want true after press")
	}
}

// TestVirtualControlsLayout_SetVisible tests visibility control.
func TestVirtualControlsLayout_SetVisible(t *testing.T) {
	layout := NewVirtualControlsLayout(800, 600)

	if !layout.Visible {
		t.Error("layout.Visible = false, want true initially")
	}

	layout.SetVisible(false)
	if layout.Visible {
		t.Error("layout.Visible = true, want false after SetVisible(false)")
	}

	layout.SetVisible(true)
	if !layout.Visible {
		t.Error("layout.Visible = false, want true after SetVisible(true)")
	}
}

// TestVirtualControlsLayout_UpdateHidden tests update when hidden.
func TestVirtualControlsLayout_UpdateHidden(t *testing.T) {
	layout := NewVirtualControlsLayout(800, 600)
	layout.SetVisible(false)

	// Update should not panic when hidden
	layout.Update()

	// Controls should not be active
	if layout.DPad.IsActive() {
		t.Error("D-pad should not be active when layout is hidden")
	}
}

// TestVirtualDPad_DirectionNormalization tests direction vector normalization.
func TestVirtualDPad_DirectionNormalization(t *testing.T) {
	dpad := NewVirtualDPad(100, 200, 50)
	touches := make(map[ebiten.TouchID]*Touch)

	// Touch at edge of D-pad (distance = radius)
	touches[0] = &Touch{
		ID:        0,
		X:         150, // 50 pixels right
		Y:         200, // Same Y
		StartTime: time.Now(),
		Active:    true,
	}

	// First update detects the touch
	dpad.Update(touches)

	// Second update calculates direction
	dpad.Update(touches)

	x, y := dpad.GetDirection()

	// Direction should be normalized (max 1.0)
	if x < 0.9 || x > 1.0 {
		t.Errorf("DirectionX = %.2f, expected ~1.0", x)
	}
	if y < -0.1 || y > 0.1 {
		t.Errorf("DirectionY = %.2f, expected ~0.0", y)
	}
}

// TestVirtualDPad_DeadZone tests dead zone behavior.
func TestVirtualDPad_DeadZone(t *testing.T) {
	dpad := NewVirtualDPad(100, 200, 50)
	touches := make(map[ebiten.TouchID]*Touch)

	// Touch very close to center (within inner radius = 15)
	touches[0] = &Touch{
		ID:        0,
		X:         105, // 5 pixels right (< 15)
		Y:         202, // 2 pixels down
		StartTime: time.Now(),
		Active:    true,
	}

	// Initial touch activates D-pad
	dpad.Update(touches)

	// Second update should apply dead zone
	dpad.Update(touches)

	x, y := dpad.GetDirection()

	// Direction should be zero in dead zone
	if x != 0.0 || y != 0.0 {
		t.Errorf("Direction in dead zone = (%.2f, %.2f), want (0.0, 0.0)", x, y)
	}
}
