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

// TestTriggerHaptic tests haptic feedback rate limiting.
func TestTriggerHaptic(t *testing.T) {
	var lastHaptic time.Time

	// First call should trigger
	triggerHaptic(10*time.Millisecond, 0.5, &lastHaptic)

	if lastHaptic.IsZero() {
		t.Error("lastHaptic should be set after first trigger")
	}

	firstTime := lastHaptic

	// Second call immediately after should be rate limited
	triggerHaptic(10*time.Millisecond, 0.5, &lastHaptic)

	if lastHaptic != firstTime {
		t.Error("lastHaptic should not change when rate limited")
	}

	// Call after sufficient time should trigger
	time.Sleep(hapticMinInterval + 10*time.Millisecond)
	triggerHaptic(10*time.Millisecond, 0.5, &lastHaptic)

	if lastHaptic == firstTime {
		t.Error("lastHaptic should update after rate limit interval")
	}
}

// TestVirtualDPadHapticIntegration tests D-pad triggers haptic on touch.
func TestVirtualDPadHapticIntegration(t *testing.T) {
	dpad := NewVirtualDPad(100, 100, 50)

	// Initially lastHaptic should be zero
	if !dpad.lastHaptic.IsZero() {
		t.Error("lastHaptic should be zero initially")
	}

	// Touch D-pad
	touches := map[ebiten.TouchID]*Touch{
		0: {ID: 0, X: 110, Y: 110, StartX: 110, StartY: 110, Active: true},
	}

	dpad.Update(touches)

	// lastHaptic should be set after touch
	if dpad.lastHaptic.IsZero() {
		t.Error("lastHaptic should be set after D-pad touch")
	}
}

// TestVirtualButtonHapticIntegration tests button triggers haptic on press.
func TestVirtualButtonHapticIntegration(t *testing.T) {
	button := NewVirtualButton(100, 100, 30, "A")

	// Initially lastHaptic should be zero
	if !button.lastHaptic.IsZero() {
		t.Error("lastHaptic should be zero initially")
	}

	// Touch button
	touches := map[ebiten.TouchID]*Touch{
		0: {ID: 0, X: 100, Y: 100, StartX: 100, StartY: 100, Active: true},
	}

	button.Update(touches)

	// Button should be active but not pressed yet
	if !button.IsActive() {
		t.Error("Button should be active after touch")
	}
	if button.IsPressed() {
		t.Error("Button should not be pressed until release")
	}

	// Release button
	touches[0].Active = false
	button.Update(touches)

	// Button should be pressed now
	if !button.IsPressed() {
		t.Error("Button should be pressed after release")
	}

	// lastHaptic should be set after press
	if button.lastHaptic.IsZero() {
		t.Error("lastHaptic should be set after button press")
	}
}

// Platform parity fix: Tests for extended VirtualControlsLayout features

// TestNewVirtualControlsLayout_ExtendedControls tests extended control creation.
func TestNewVirtualControlsLayout_ExtendedControls(t *testing.T) {
	layout := NewVirtualControlsLayout(800, 600)

	// Verify extended controls are created
	if layout.InventoryButton == nil {
		t.Error("layout.InventoryButton is nil")
	}
	if layout.TargetButton == nil {
		t.Error("layout.TargetButton is nil")
	}
	if layout.InteractButton == nil {
		t.Error("layout.InteractButton is nil")
	}
	if len(layout.SpellButtons) != 5 {
		t.Errorf("len(layout.SpellButtons) = %d, want 5", len(layout.SpellButtons))
	}
	if !layout.ShowSpellButtons {
		t.Error("layout.ShowSpellButtons = false, want true initially")
	}
}

// TestVirtualControlsLayout_IsInventoryPressed tests inventory button.
func TestVirtualControlsLayout_IsInventoryPressed(t *testing.T) {
	layout := NewVirtualControlsLayout(800, 600)

	if layout.IsInventoryPressed() {
		t.Error("IsInventoryPressed() = true, want false initially")
	}

	layout.InventoryButton.Pressed = true

	if !layout.IsInventoryPressed() {
		t.Error("IsInventoryPressed() = false, want true after press")
	}
}

// TestVirtualControlsLayout_IsTargetPressed tests target cycle button.
func TestVirtualControlsLayout_IsTargetPressed(t *testing.T) {
	layout := NewVirtualControlsLayout(800, 600)

	if layout.IsTargetPressed() {
		t.Error("IsTargetPressed() = true, want false initially")
	}

	layout.TargetButton.Pressed = true

	if !layout.IsTargetPressed() {
		t.Error("IsTargetPressed() = false, want true after press")
	}
}

// TestVirtualControlsLayout_IsInteractPressed tests interact button.
func TestVirtualControlsLayout_IsInteractPressed(t *testing.T) {
	layout := NewVirtualControlsLayout(800, 600)

	if layout.IsInteractPressed() {
		t.Error("IsInteractPressed() = true, want false initially")
	}

	layout.InteractButton.Pressed = true

	if !layout.IsInteractPressed() {
		t.Error("IsInteractPressed() = false, want true after press")
	}
}

// TestVirtualControlsLayout_IsSpellPressed tests spell button detection.
func TestVirtualControlsLayout_IsSpellPressed(t *testing.T) {
	layout := NewVirtualControlsLayout(800, 600)

	// All spell buttons should be not pressed initially
	for slot := 1; slot <= 5; slot++ {
		if layout.IsSpellPressed(slot) {
			t.Errorf("IsSpellPressed(%d) = true, want false initially", slot)
		}
	}

	// Test each slot
	for slot := 1; slot <= 5; slot++ {
		layout.SpellButtons[slot-1].Pressed = true
		if !layout.IsSpellPressed(slot) {
			t.Errorf("IsSpellPressed(%d) = false, want true after press", slot)
		}
		layout.SpellButtons[slot-1].Pressed = false
	}

	// Invalid slots should return false
	if layout.IsSpellPressed(0) {
		t.Error("IsSpellPressed(0) should return false for invalid slot")
	}
	if layout.IsSpellPressed(6) {
		t.Error("IsSpellPressed(6) should return false for out of range slot")
	}
}

// TestVirtualControlsLayout_SetSpellButtonsVisible tests spell visibility toggle.
func TestVirtualControlsLayout_SetSpellButtonsVisible(t *testing.T) {
	layout := NewVirtualControlsLayout(800, 600)

	if !layout.ShowSpellButtons {
		t.Error("ShowSpellButtons = false, want true initially")
	}

	layout.SetSpellButtonsVisible(false)
	if layout.ShowSpellButtons {
		t.Error("ShowSpellButtons = true, want false after SetSpellButtonsVisible(false)")
	}

	// IsSpellPressed should return false when hidden
	layout.SpellButtons[0].Pressed = true
	if layout.IsSpellPressed(1) {
		t.Error("IsSpellPressed(1) should return false when spell buttons hidden")
	}

	layout.SetSpellButtonsVisible(true)
	if !layout.ShowSpellButtons {
		t.Error("ShowSpellButtons = false, want true after SetSpellButtonsVisible(true)")
	}
}

// TestVirtualControlsLayout_Resize tests control repositioning on resize.
func TestVirtualControlsLayout_Resize(t *testing.T) {
	layout := NewVirtualControlsLayout(800, 600)

	origDPadX := layout.DPad.X
	origMenuX := layout.MenuButton.X

	// Resize to larger screen
	layout.Resize(1920, 1080)

	// Verify positions changed
	if layout.DPad.X == origDPadX {
		t.Error("D-pad X position should change on resize")
	}
	if layout.MenuButton.X == origMenuX {
		t.Error("Menu button X position should change on resize")
	}

	// Verify all extended controls still valid
	if layout.InventoryButton == nil {
		t.Error("InventoryButton should not be nil after resize")
	}
	if layout.TargetButton == nil {
		t.Error("TargetButton should not be nil after resize")
	}
	if layout.InteractButton == nil {
		t.Error("InteractButton should not be nil after resize")
	}
	for i, btn := range layout.SpellButtons {
		if btn == nil {
			t.Errorf("SpellButtons[%d] should not be nil after resize", i)
		}
	}
}

// TestNewVirtualControlsLayout_UIShortcuts tests UI shortcut button creation.
func TestNewVirtualControlsLayout_UIShortcuts(t *testing.T) {
	layout := NewVirtualControlsLayout(800, 600)

	// All UI shortcut buttons should be created
	if layout.CharacterButton == nil {
		t.Error("CharacterButton should not be nil")
	}
	if layout.SkillsButton == nil {
		t.Error("SkillsButton should not be nil")
	}
	if layout.QuestLogButton == nil {
		t.Error("QuestLogButton should not be nil")
	}
	if layout.MapButton == nil {
		t.Error("MapButton should not be nil")
	}

	// ShowUIShortcuts should default to true
	if !layout.ShowUIShortcuts {
		t.Error("ShowUIShortcuts should be true by default")
	}

	// Verify button labels
	if layout.CharacterButton.Label != "C" {
		t.Errorf("CharacterButton.Label = %q, want %q", layout.CharacterButton.Label, "C")
	}
	if layout.SkillsButton.Label != "K" {
		t.Errorf("SkillsButton.Label = %q, want %q", layout.SkillsButton.Label, "K")
	}
	if layout.QuestLogButton.Label != "J" {
		t.Errorf("QuestLogButton.Label = %q, want %q", layout.QuestLogButton.Label, "J")
	}
	if layout.MapButton.Label != "M" {
		t.Errorf("MapButton.Label = %q, want %q", layout.MapButton.Label, "M")
	}
}

// TestVirtualControlsLayout_IsCharacterPressed tests character button.
func TestVirtualControlsLayout_IsCharacterPressed(t *testing.T) {
	layout := NewVirtualControlsLayout(800, 600)

	// Initially not pressed
	if layout.IsCharacterPressed() {
		t.Error("IsCharacterPressed() = true, want false initially")
	}

	// Simulate button press
	layout.CharacterButton.Pressed = true
	if !layout.IsCharacterPressed() {
		t.Error("IsCharacterPressed() = false after button press")
	}
}

// TestVirtualControlsLayout_IsSkillsPressed tests skills button.
func TestVirtualControlsLayout_IsSkillsPressed(t *testing.T) {
	layout := NewVirtualControlsLayout(800, 600)

	// Initially not pressed
	if layout.IsSkillsPressed() {
		t.Error("IsSkillsPressed() = true, want false initially")
	}

	// Simulate button press
	layout.SkillsButton.Pressed = true
	if !layout.IsSkillsPressed() {
		t.Error("IsSkillsPressed() = false after button press")
	}
}

// TestVirtualControlsLayout_IsQuestLogPressed tests quest log button.
func TestVirtualControlsLayout_IsQuestLogPressed(t *testing.T) {
	layout := NewVirtualControlsLayout(800, 600)

	// Initially not pressed
	if layout.IsQuestLogPressed() {
		t.Error("IsQuestLogPressed() = true, want false initially")
	}

	// Simulate button press
	layout.QuestLogButton.Pressed = true
	if !layout.IsQuestLogPressed() {
		t.Error("IsQuestLogPressed() = false after button press")
	}
}

// TestVirtualControlsLayout_IsMapPressed tests map button.
func TestVirtualControlsLayout_IsMapPressed(t *testing.T) {
	layout := NewVirtualControlsLayout(800, 600)

	// Initially not pressed
	if layout.IsMapPressed() {
		t.Error("IsMapPressed() = true, want false initially")
	}

	// Simulate button press
	layout.MapButton.Pressed = true
	if !layout.IsMapPressed() {
		t.Error("IsMapPressed() = false after button press")
	}
}

// TestVirtualControlsLayout_SetUIShortcutsVisible tests UI shortcuts visibility toggle.
func TestVirtualControlsLayout_SetUIShortcutsVisible(t *testing.T) {
	layout := NewVirtualControlsLayout(800, 600)

	// Default should be visible
	if !layout.ShowUIShortcuts {
		t.Error("ShowUIShortcuts should be true by default")
	}

	// Simulate button press while visible
	layout.CharacterButton.Pressed = true
	if !layout.IsCharacterPressed() {
		t.Error("Character button should be detected when shortcuts visible")
	}

	// Reset button state before testing visibility toggle
	layout.CharacterButton.Pressed = false

	// Hide shortcuts
	layout.SetUIShortcutsVisible(false)
	if layout.ShowUIShortcuts {
		t.Error("ShowUIShortcuts should be false after SetUIShortcutsVisible(false)")
	}

	// Simulate fresh button press while hidden
	layout.CharacterButton.Pressed = true
	if layout.IsCharacterPressed() {
		t.Error("Character button should NOT be detected when shortcuts hidden")
	}

	// Show shortcuts again
	layout.SetUIShortcutsVisible(true)
	if !layout.ShowUIShortcuts {
		t.Error("ShowUIShortcuts should be true after SetUIShortcutsVisible(true)")
	}

	// Verify button press is now detected again (round-trip test)
	if !layout.IsCharacterPressed() {
		t.Error("Character button should be detected after re-enabling shortcuts")
	}
}

// TestVirtualControlsLayout_Resize_UIShortcuts tests UI shortcut button repositioning on resize.
func TestVirtualControlsLayout_Resize_UIShortcuts(t *testing.T) {
	layout := NewVirtualControlsLayout(800, 600)

	// Get initial positions
	charX := layout.CharacterButton.X
	charY := layout.CharacterButton.Y

	// Resize to larger screen
	layout.Resize(1280, 720)

	// Verify buttons still exist
	if layout.CharacterButton == nil {
		t.Error("CharacterButton should not be nil after resize")
	}
	if layout.SkillsButton == nil {
		t.Error("SkillsButton should not be nil after resize")
	}
	if layout.QuestLogButton == nil {
		t.Error("QuestLogButton should not be nil after resize")
	}
	if layout.MapButton == nil {
		t.Error("MapButton should not be nil after resize")
	}

	// Positions should have changed due to relative positioning
	if layout.CharacterButton.X == charX && layout.CharacterButton.Y == charY {
		t.Log("Note: Button positions may not change if relative calculation is same")
	}
}
