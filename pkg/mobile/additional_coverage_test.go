package mobile

import (
	"image/color"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestMobileMenu_UpdateWithVisibleMenu tests menu update when visible.
func TestMobileMenu_UpdateWithVisibleMenu(t *testing.T) {
	menu := NewMobileMenu(100, 100, 300, 400)
	menu.Show()

	called := false
	menu.AddItem("Test Item", true, func() {
		called = true
	})

	// Update should not panic
	menu.Update()

	// Without actual touch input (requires Ebiten), callback won't be called
	// But we can test the code path exists
	_ = called // Use the variable to avoid compiler error
}

// TestMobileMenu_UpdateScrolling tests menu scrolling logic.
func TestMobileMenu_UpdateScrolling(t *testing.T) {
	menu := NewMobileMenu(100, 100, 300, 200) // Small height
	menu.Show()

	// Add many items to make scrolling necessary
	for i := 0; i < 10; i++ {
		menu.AddItem("Item", true, nil)
	}

	// Update should not panic even with many items
	menu.Update()

	// Note: Scroll offset clamping happens only when swipe is detected
	// which requires actual touch input (Ebiten APIs)
	// We can't test scroll clamping without Ebiten touch events
}

// TestMobileHUD_UpdateOrientation tests orientation update.
func TestMobileHUD_UpdateOrientation(t *testing.T) {
	hud := NewMobileHUD(800, 600) // Landscape

	// Initially landscape
	if hud.Orientation != OrientationLandscape {
		t.Errorf("Initial orientation = %v, want %v", hud.Orientation, OrientationLandscape)
	}

	// Change to portrait
	hud.UpdateOrientation(600, 800)

	if hud.ScreenWidth != 600 {
		t.Errorf("ScreenWidth = %d, want 600", hud.ScreenWidth)
	}
	if hud.ScreenHeight != 800 {
		t.Errorf("ScreenHeight = %d, want 800", hud.ScreenHeight)
	}
	if hud.Orientation != OrientationPortrait {
		t.Errorf("Orientation = %v, want %v after update", hud.Orientation, OrientationPortrait)
	}
}

// TestMobileHUD_UpdateOrientationNoChange tests orientation with no change.
func TestMobileHUD_UpdateOrientationNoChange(t *testing.T) {
	hud := NewMobileHUD(800, 600)

	initialOrientation := hud.Orientation

	// Update with same aspect ratio
	hud.UpdateOrientation(1600, 1200)

	if hud.Orientation != initialOrientation {
		t.Error("Orientation changed when aspect ratio stayed the same")
	}
}

// TestVirtualControlsLayout_UpdateActive tests layout update with active controls.
func TestVirtualControlsLayout_UpdateActive(t *testing.T) {
	layout := NewVirtualControlsLayout(800, 600)
	layout.SetVisible(true)

	// Update should process all controls without panic
	layout.Update()

	// Verify touch handler was used
	if layout.touchHandler == nil {
		t.Error("touchHandler should not be nil after Update")
	}
}

// TestMobileMenu_UpdateHiddenNoProcessing tests that hidden menu doesn't process input.
func TestMobileMenu_UpdateHiddenNoProcessing(t *testing.T) {
	menu := NewMobileMenu(100, 100, 300, 400)
	menu.Hide()

	called := false
	menu.AddItem("Test", true, func() {
		called = true
	})

	// Multiple updates should not trigger callback when hidden
	for i := 0; i < 10; i++ {
		menu.Update()
	}

	if called {
		t.Error("Menu item callback was called while menu was hidden")
	}
}

// TestProgressBar_MultipleSetValue tests multiple value updates.
func TestProgressBar_MultipleSetValue(t *testing.T) {
	bar := NewProgressBar(10, 20, 100, 15, color.RGBA{255, 0, 0, 255})

	values := []float64{0.0, 0.25, 0.5, 0.75, 1.0, 0.6, 0.3}

	for _, v := range values {
		bar.SetValue(v)
		if bar.Value != v {
			t.Errorf("After SetValue(%.2f), Value = %.2f", v, bar.Value)
		}
	}
}

// TestNotificationWidget_MultipleShows tests showing multiple notifications.
func TestNotificationWidget_MultipleShows(t *testing.T) {
	notif := NewNotificationWidget(300, 50, 200, 50)

	notif.Show("First", 3.0)
	if notif.Message != "First" {
		t.Errorf("Message = %q, want %q", notif.Message, "First")
	}

	notif.Show("Second", 5.0)
	if notif.Message != "Second" {
		t.Errorf("Message = %q after second show, want %q", notif.Message, "Second")
	}
	if notif.Duration != 5.0 {
		t.Errorf("Duration = %.1f after second show, want 5.0", notif.Duration)
	}
}

// TestVirtualDPad_MultipleActivations tests D-pad with multiple touch cycles.
func TestVirtualDPad_MultipleActivations(t *testing.T) {
	dpad := NewVirtualDPad(100, 200, 50)
	touches := make(map[ebiten.TouchID]*Touch)

	// First activation
	touches[0] = &Touch{ID: 0, X: 110, Y: 210, Active: true, StartTime: time.Now()}
	dpad.Update(touches)
	dpad.Update(touches)

	if !dpad.IsActive() {
		t.Error("D-pad should be active after first touch")
	}

	// Release
	delete(touches, 0)
	dpad.Update(touches)

	if dpad.IsActive() {
		t.Error("D-pad should not be active after release")
	}

	// Second activation with different touch ID
	touches[1] = &Touch{ID: 1, X: 120, Y: 210, Active: true, StartTime: time.Now()}
	dpad.Update(touches)
	dpad.Update(touches)

	if !dpad.IsActive() {
		t.Error("D-pad should be active after second touch")
	}
}

// TestVirtualButton_MultiplePress tests button with multiple press cycles.
func TestVirtualButton_MultiplePress(t *testing.T) {
	button := NewVirtualButton(300, 400, 30, "A")
	touches := make(map[ebiten.TouchID]*Touch)

	// First press cycle
	touches[0] = &Touch{ID: 0, X: 310, Y: 410, Active: true, StartTime: time.Now()}
	button.Update(touches)

	delete(touches, 0)
	button.Update(touches)

	if !button.IsPressed() {
		t.Error("Button should be pressed after first release")
	}

	// Second update clears pressed state
	button.Update(touches)
	if button.IsPressed() {
		t.Error("Button should not be pressed in next frame")
	}

	// Second press cycle
	touches[1] = &Touch{ID: 1, X: 305, Y: 405, Active: true, StartTime: time.Now()}
	button.Update(touches)

	delete(touches, 1)
	button.Update(touches)

	if !button.IsPressed() {
		t.Error("Button should be pressed after second release")
	}
}
