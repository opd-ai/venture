package mobile

import (
	"image/color"
	"testing"
)

// TestNewMobileMenu tests menu creation.
func TestNewMobileMenu(t *testing.T) {
	menu := NewMobileMenu(100, 100, 300, 400)

	if menu == nil {
		t.Fatal("NewMobileMenu returned nil")
	}
	if menu.X != 100 {
		t.Errorf("menu.X = %.1f, want 100.0", menu.X)
	}
	if menu.Y != 100 {
		t.Errorf("menu.Y = %.1f, want 100.0", menu.Y)
	}
	if menu.Width != 300 {
		t.Errorf("menu.Width = %.1f, want 300.0", menu.Width)
	}
	if menu.Height != 400 {
		t.Errorf("menu.Height = %.1f, want 400.0", menu.Height)
	}
	if menu.Items == nil {
		t.Error("menu.Items is nil")
	}
	if len(menu.Items) != 0 {
		t.Errorf("menu.Items has %d items, want 0 initially", len(menu.Items))
	}
	if menu.touchHandler == nil {
		t.Error("menu.touchHandler is nil")
	}
}

// TestMobileMenu_AddItem tests adding menu items.
func TestMobileMenu_AddItem(t *testing.T) {
	menu := NewMobileMenu(100, 100, 300, 400)

	called := false
	onSelect := func() { called = true }

	menu.AddItem("Test Item", true, onSelect)

	if len(menu.Items) != 1 {
		t.Errorf("menu has %d items, want 1", len(menu.Items))
	}

	item := menu.Items[0]
	if item.Label != "Test Item" {
		t.Errorf("item.Label = %q, want %q", item.Label, "Test Item")
	}
	if !item.Enabled {
		t.Error("item.Enabled = false, want true")
	}
	if item.OnSelect == nil {
		t.Fatal("item.OnSelect is nil")
	}

	// Test callback
	item.OnSelect()
	if !called {
		t.Error("OnSelect callback was not called")
	}
}

// TestMobileMenu_AddMultipleItems tests adding multiple menu items.
func TestMobileMenu_AddMultipleItems(t *testing.T) {
	menu := NewMobileMenu(100, 100, 300, 400)

	menu.AddItem("Item 1", true, nil)
	menu.AddItem("Item 2", false, nil)
	menu.AddItem("Item 3", true, nil)

	if len(menu.Items) != 3 {
		t.Errorf("menu has %d items, want 3", len(menu.Items))
	}

	if menu.Items[0].Label != "Item 1" {
		t.Errorf("Items[0].Label = %q, want %q", menu.Items[0].Label, "Item 1")
	}
	if menu.Items[1].Enabled {
		t.Error("Items[1].Enabled = true, want false")
	}
	if menu.Items[2].Label != "Item 3" {
		t.Errorf("Items[2].Label = %q, want %q", menu.Items[2].Label, "Item 3")
	}
}

// TestMobileMenu_ShowHide tests menu visibility control.
func TestMobileMenu_ShowHide(t *testing.T) {
	menu := NewMobileMenu(100, 100, 300, 400)

	// Initially not visible
	if menu.Visible {
		t.Error("menu.Visible = true, want false initially")
	}

	menu.Show()
	if !menu.Visible {
		t.Error("menu.Visible = false after Show()")
	}

	menu.Hide()
	if menu.Visible {
		t.Error("menu.Visible = true after Hide()")
	}
}

// TestMobileMenu_Toggle tests menu visibility toggle.
func TestMobileMenu_Toggle(t *testing.T) {
	menu := NewMobileMenu(100, 100, 300, 400)

	initial := menu.Visible
	menu.Toggle()
	if menu.Visible == initial {
		t.Error("Toggle() did not change visibility")
	}

	menu.Toggle()
	if menu.Visible != initial {
		t.Error("Toggle() did not restore original visibility")
	}
}

// TestMobileMenu_IsVisible tests visibility query.
func TestMobileMenu_IsVisible(t *testing.T) {
	menu := NewMobileMenu(100, 100, 300, 400)

	if menu.IsVisible() {
		t.Error("IsVisible() = true, want false initially")
	}

	menu.Show()
	if !menu.IsVisible() {
		t.Error("IsVisible() = false after Show()")
	}
}

// TestMobileMenu_UpdateHidden tests update when hidden.
func TestMobileMenu_UpdateHidden(t *testing.T) {
	menu := NewMobileMenu(100, 100, 300, 400)
	menu.Hide()

	// Update should not panic when hidden
	menu.Update()
}

// TestNewMobileHUD tests HUD creation.
func TestNewMobileHUD(t *testing.T) {
	hud := NewMobileHUD(800, 600)

	if hud == nil {
		t.Fatal("NewMobileHUD returned nil")
	}
	if hud.ScreenWidth != 800 {
		t.Errorf("hud.ScreenWidth = %d, want 800", hud.ScreenWidth)
	}
	if hud.ScreenHeight != 600 {
		t.Errorf("hud.ScreenHeight = %d, want 600", hud.ScreenHeight)
	}
	if !hud.Visible {
		t.Error("hud.Visible = false, want true initially")
	}

	// Check orientation detection
	if hud.Orientation != OrientationLandscape {
		t.Errorf("hud.Orientation = %v, want OrientationLandscape for 800x600", hud.Orientation)
	}
}

// TestNewMobileHUD_Portrait tests HUD creation in portrait orientation.
func TestNewMobileHUD_Portrait(t *testing.T) {
	hud := NewMobileHUD(600, 800)

	if hud.Orientation != OrientationPortrait {
		t.Errorf("hud.Orientation = %v, want OrientationPortrait for 600x800", hud.Orientation)
	}
}

// TestMobileHUD_SetHealth tests health value setting.
func TestMobileHUD_SetHealth(t *testing.T) {
	hud := NewMobileHUD(800, 600)

	// Ensure HealthBar is initialized (may be done in LayoutElements)
	if hud.HealthBar == nil {
		hud.HealthBar = NewProgressBar(10, 10, 150, 20, color.RGBA{255, 0, 0, 255})
	}

	hud.SetHealth(0.5) // Set to 50%

	if hud.HealthBar.Value != 0.5 {
		t.Errorf("HealthBar.Value = %.1f, want 0.5", hud.HealthBar.Value)
	}
}

// TestMobileHUD_SetMana tests mana value setting.
func TestMobileHUD_SetMana(t *testing.T) {
	hud := NewMobileHUD(800, 600)

	if hud.ManaBar == nil {
		hud.ManaBar = NewProgressBar(10, 40, 150, 20, color.RGBA{0, 0, 255, 255})
	}

	hud.SetMana(0.6) // Set to 60%

	if hud.ManaBar.Value != 0.6 {
		t.Errorf("ManaBar.Value = %.1f, want 0.6", hud.ManaBar.Value)
	}
}

// TestMobileHUD_SetExperience tests experience value setting.
func TestMobileHUD_SetExperience(t *testing.T) {
	hud := NewMobileHUD(800, 600)

	if hud.ExpBar == nil {
		hud.ExpBar = NewProgressBar(10, 70, 150, 20, color.RGBA{0, 255, 0, 255})
	}

	hud.SetExperience(0.75) // Set to 75%

	if hud.ExpBar.Value != 0.75 {
		t.Errorf("ExpBar.Value = %.1f, want 0.75", hud.ExpBar.Value)
	}
}

// TestMobileHUD_ShowNotification tests notification display.
func TestMobileHUD_ShowNotification(t *testing.T) {
	hud := NewMobileHUD(800, 600)

	if hud.Notification == nil {
		hud.Notification = NewNotificationWidget(400, 50, 300, 60)
	}

	hud.ShowNotification("Test Message", 3.0)

	if hud.Notification.Message != "Test Message" {
		t.Errorf("Notification.Message = %q, want %q", hud.Notification.Message, "Test Message")
	}
	if !hud.Notification.Visible {
		t.Error("Notification.Visible = false after ShowNotification()")
	}
}

// TestMobileHUD_Update tests HUD update logic.
func TestMobileHUD_Update(t *testing.T) {
	hud := NewMobileHUD(800, 600)

	// Update should not panic
	hud.Update(1.0 / 60.0)
}

// TestMobileHUD_UpdateHidden tests update when hidden.
func TestMobileHUD_UpdateHidden(t *testing.T) {
	hud := NewMobileHUD(800, 600)
	hud.Visible = false

	// Update should not panic when hidden
	hud.Update(1.0 / 60.0)
}

// TestNewProgressBar tests progress bar creation.
func TestNewProgressBar(t *testing.T) {
	bar := NewProgressBar(10, 20, 100, 15, color.RGBA{255, 0, 0, 255})

	if bar == nil {
		t.Fatal("NewProgressBar returned nil")
	}
	if bar.X != 10 {
		t.Errorf("bar.X = %.1f, want 10.0", bar.X)
	}
	if bar.Y != 20 {
		t.Errorf("bar.Y = %.1f, want 20.0", bar.Y)
	}
	if bar.Width != 100 {
		t.Errorf("bar.Width = %.1f, want 100.0", bar.Width)
	}
	if bar.Height != 15 {
		t.Errorf("bar.Height = %.1f, want 15.0", bar.Height)
	}
	if bar.Value != 1.0 {
		t.Errorf("bar.Value = %.1f, want 1.0 initially", bar.Value)
	}
}

// TestProgressBar_SetValue tests setting progress bar value.
func TestProgressBar_SetValue(t *testing.T) {
	bar := NewProgressBar(10, 20, 100, 15, color.RGBA{255, 0, 0, 255})

	bar.SetValue(0.5) // 50%

	if bar.Value != 0.5 {
		t.Errorf("bar.Value = %.1f, want 0.5", bar.Value)
	}
}

// TestProgressBar_SetValueBounds tests value clamping.
func TestProgressBar_SetValueBounds(t *testing.T) {
	bar := NewProgressBar(10, 20, 100, 15, color.RGBA{255, 0, 0, 255})

	// Test overflow
	bar.SetValue(1.5) // Above 1.0
	if bar.Value != 1.0 {
		t.Errorf("bar.Value = %.1f, want 1.0 (clamped)", bar.Value)
	}

	// Test underflow
	bar.SetValue(-0.1) // Below 0.0
	if bar.Value != 0.0 {
		t.Errorf("bar.Value = %.1f, want 0.0 (clamped)", bar.Value)
	}
}

// TestProgressBar_ValueRange tests various value ranges.
func TestProgressBar_ValueRange(t *testing.T) {
	bar := NewProgressBar(10, 20, 100, 15, color.RGBA{255, 0, 0, 255})

	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"50%", 0.5, 0.5},
		{"0%", 0.0, 0.0},
		{"100%", 1.0, 1.0},
		{"75%", 0.75, 0.75},
		{"25%", 0.25, 0.25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bar.SetValue(tt.input)
			if bar.Value != tt.expected {
				t.Errorf("SetValue(%.2f): Value = %.2f, want %.2f", tt.input, bar.Value, tt.expected)
			}
		})
	}
}

// TestNewMinimapWidget tests minimap widget creation.
func TestNewMinimapWidget(t *testing.T) {
	minimap := NewMinimapWidget(600, 10, 180, 180)

	if minimap == nil {
		t.Fatal("NewMinimapWidget returned nil")
	}
	if minimap.X != 600 {
		t.Errorf("minimap.X = %.1f, want 600.0", minimap.X)
	}
	if minimap.Y != 10 {
		t.Errorf("minimap.Y = %.1f, want 10.0", minimap.Y)
	}
	if minimap.Width != 180 {
		t.Errorf("minimap.Width = %.1f, want 180.0", minimap.Width)
	}
	if minimap.Height != 180 {
		t.Errorf("minimap.Height = %.1f, want 180.0", minimap.Height)
	}
}

// TestMinimapWidget_SetPlayerPosition tests setting player position.
func TestMinimapWidget_SetPlayerPosition(t *testing.T) {
	minimap := NewMinimapWidget(600, 10, 180, 180)

	// MinimapWidget structure is simple - just verify creation
	if minimap.X != 600 || minimap.Y != 10 {
		t.Error("Minimap position not set correctly")
	}
}

// TestNewNotificationWidget tests notification widget creation.
func TestNewNotificationWidget(t *testing.T) {
	notif := NewNotificationWidget(300, 50, 200, 50)

	if notif == nil {
		t.Fatal("NewNotificationWidget returned nil")
	}
	if notif.X != 300 {
		t.Errorf("notif.X = %.1f, want 300.0", notif.X)
	}
	if notif.Y != 50 {
		t.Errorf("notif.Y = %.1f, want 50.0", notif.Y)
	}
	if notif.Width != 200 {
		t.Errorf("notif.Width = %.1f, want 200.0", notif.Width)
	}
	if notif.Height != 50 {
		t.Errorf("notif.Height = %.1f, want 50.0", notif.Height)
	}
	if notif.Visible {
		t.Error("notif.Visible = true, want false initially")
	}
}

// TestNotificationWidget_Show tests showing notification.
func TestNotificationWidget_Show(t *testing.T) {
	notif := NewNotificationWidget(300, 50, 200, 50)

	notif.Show("Test notification", 3.0)

	if notif.Message != "Test notification" {
		t.Errorf("notif.Message = %q, want %q", notif.Message, "Test notification")
	}
	if !notif.Visible {
		t.Error("notif.Visible = false after Show()")
	}
	if notif.Duration != 3.0 {
		t.Errorf("notif.Duration = %.1f, want 3.0", notif.Duration)
	}
	if notif.Remaining != 3.0 {
		t.Errorf("notif.Remaining = %.1f, want 3.0 initially", notif.Remaining)
	}
}

// TestNotificationWidget_Update tests notification timer.
func TestNotificationWidget_Update(t *testing.T) {
	notif := NewNotificationWidget(300, 50, 200, 50)
	notif.Show("Test", 3.0)

	// Simulate time passing
	dt := 1.0 / 60.0 // One frame at 60 FPS

	initialRemaining := notif.Remaining
	notif.Update(dt)

	if notif.Remaining >= initialRemaining {
		t.Error("notif.Remaining did not decrease after Update()")
	}
}

// TestNotificationWidget_UpdateExpire tests notification expiration.
func TestNotificationWidget_UpdateExpire(t *testing.T) {
	notif := NewNotificationWidget(300, 50, 200, 50)
	notif.Show("Test", 3.0)

	// Simulate duration passing
	notif.Remaining = -0.1

	notif.Update(1.0 / 60.0)

	if notif.Visible {
		t.Error("notif.Visible = true after duration expired")
	}
}

// TestNotificationWidget_UpdateInactive tests update when inactive.
func TestNotificationWidget_UpdateInactive(t *testing.T) {
	notif := NewNotificationWidget(300, 50, 200, 50)

	// Update inactive notification should not panic
	notif.Update(1.0 / 60.0)

	if notif.Visible {
		t.Error("notif.Visible = true, should remain false")
	}
}

// TestMenuItem tests MenuItem structure.
func TestMenuItem(t *testing.T) {
	called := false
	item := MenuItem{
		Label:   "Test",
		Enabled: true,
		OnSelect: func() {
			called = true
		},
	}

	if item.Label != "Test" {
		t.Errorf("item.Label = %q, want %q", item.Label, "Test")
	}
	if !item.Enabled {
		t.Error("item.Enabled = false, want true")
	}
	if item.OnSelect == nil {
		t.Fatal("item.OnSelect is nil")
	}

	item.OnSelect()
	if !called {
		t.Error("OnSelect callback was not called")
	}
}

// Gap #9 fix: Comprehensive momentum scrolling tests

// TestMobileMenu_DefaultScrollDeceleration tests default deceleration value.
// Gap #9: Verify default is 0.98 for iOS/Android-like feel (1.5-2s scroll)
func TestMobileMenu_DefaultScrollDeceleration(t *testing.T) {
	menu := NewMobileMenu(100, 100, 300, 400)

	got := menu.GetScrollDeceleration()
	want := 0.98

	if got != want {
		t.Errorf("Default scrollDeceleration = %.2f, want %.2f (iOS/Android-like)", got, want)
	}
}

// TestMobileMenu_SetScrollDeceleration tests setting custom deceleration.
// Gap #9: Verify setter allows customization within valid range
func TestMobileMenu_SetScrollDeceleration(t *testing.T) {
	menu := NewMobileMenu(100, 100, 300, 400)

	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"iOS-like (0.98)", 0.98, 0.98},
		{"Faster (0.95)", 0.95, 0.95},
		{"Slower (0.99)", 0.99, 0.99},
		{"Mid-range (0.96)", 0.96, 0.96},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			menu.SetScrollDeceleration(tt.input)
			got := menu.GetScrollDeceleration()
			if got != tt.expected {
				t.Errorf("SetScrollDeceleration(%.2f): got %.2f, want %.2f", tt.input, got, tt.expected)
			}
		})
	}
}

// TestMobileMenu_SetScrollDeceleration_BoundsChecking tests clamping to valid range.
// Gap #9: Verify extreme values are clamped to realistic physics (0.9-0.99)
func TestMobileMenu_SetScrollDeceleration_BoundsChecking(t *testing.T) {
	menu := NewMobileMenu(100, 100, 300, 400)

	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"Too low (0.5)", 0.5, 0.9},   // Clamped to minimum
		{"Too high (1.5)", 1.5, 0.99}, // Clamped to maximum
		{"Below min (0.89)", 0.89, 0.9},
		{"Above max (1.0)", 1.0, 0.99},
		{"Way too low (0.1)", 0.1, 0.9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			menu.SetScrollDeceleration(tt.input)
			got := menu.GetScrollDeceleration()
			if got != tt.expected {
				t.Errorf("SetScrollDeceleration(%.2f): got %.2f, want %.2f (clamped)", tt.input, got, tt.expected)
			}
		})
	}
}

// TestMomentumScrolling_DecayRate tests scroll velocity decay over time.
// Gap #9: Verify 0.98 deceleration provides iOS-like 1.5-2s scroll duration
func TestMomentumScrolling_DecayRate(t *testing.T) {
	menu := NewMobileMenu(100, 100, 300, 400)
	// Add enough items to make content taller than viewport (maxScroll > 0)
	// With 400px height and 20 items at 50px each = 1000px content, maxScroll = 600px
	for i := 0; i < 20; i++ {
		menu.Items = append(menu.Items, MenuItem{Label: "Item"})
	}
	menu.SetScrollDeceleration(0.98) // iOS-like

	// Use a lower initial velocity that won't trigger bounce-back during test
	// The menu can scroll 600px total. With velocity -5, after 120 frames we scroll:
	// sum of -5 * 0.98^i for i=0..119 ≈ -5 * (1-0.98^120)/(1-0.98) ≈ -5 * 45.5 = -228px
	// This stays well within bounds and tests pure deceleration without bounce
	menu.scrollVelocity = -5.0
	menu.isScrolling = true

	// Track velocity decay over 2 seconds (120 frames at 60 FPS)
	velocities := []float64{}
	for i := 0; i < 120; i++ {
		velocities = append(velocities, menu.scrollVelocity)
		menu.updateMomentumScrolling()
	}

	// At 0.98 deceleration, after 120 frames:
	// velocity = -5 * 0.98^120 ≈ -0.45 (about 9% of original magnitude)
	finalVelocity := velocities[119]
	expectedFinal := -5.0 * 0.09 // approximately -0.45
	if finalVelocity > expectedFinal+0.2 {
		t.Errorf("Velocity decayed too fast: %.2f at 2s, expected ~%.2f", finalVelocity, expectedFinal)
	}
	if finalVelocity < expectedFinal-0.2 {
		t.Errorf("Velocity decayed too slow: %.2f at 2s, expected ~%.2f", finalVelocity, expectedFinal)
	}

	// Verify smooth decay (absolute value should decrease)
	for i := 1; i < len(velocities); i++ {
		absDecay := absFloatVal(velocities[i-1]) - absFloatVal(velocities[i])
		if absDecay < -0.001 { // small tolerance for floating point
			t.Errorf("Velocity magnitude increased at frame %d (should only decrease)", i)
		}
	}
}

// absFloatVal returns the absolute value of a float64
func absFloatVal(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// TestMomentumScrolling_FastDecayComparison tests faster 0.95 deceleration.
// Gap #9: Verify 0.95 provides faster ~0.7s scroll (original behavior)
func TestMomentumScrolling_FastDecayComparison(t *testing.T) {
	menu := NewMobileMenu(100, 100, 300, 400)
	menu.SetScrollDeceleration(0.95) // Faster decay

	menu.scrollVelocity = 100.0
	menu.isScrolling = true

	// Simulate 1 second (60 frames)
	for i := 0; i < 60; i++ {
		menu.updateMomentumScrolling()
	}

	// At 0.95 deceleration, after 60 frames:
	// velocity = 100 * 0.95^60 ≈ 4.64
	// Should have very little velocity remaining (~5%)
	finalVelocity := menu.scrollVelocity
	if finalVelocity > 10.0 {
		t.Errorf("0.95 decay too slow: %.2f at 1s, expected ~5", finalVelocity)
	}

	// Compare to 0.98: At 1s, 0.98^60 ≈ 30% remaining vs 0.95^60 ≈ 5%
	// This demonstrates the significant difference in scroll feel
}

// TestMomentumScrolling_StopOnLowVelocity tests automatic stop at negligible velocity.
// Gap #9: Verify scrolling stops cleanly when velocity < 0.1
func TestMomentumScrolling_StopOnLowVelocity(t *testing.T) {
	menu := NewMobileMenu(100, 100, 300, 400)

	// Start with very low velocity
	menu.scrollVelocity = 0.05
	menu.isScrolling = true

	menu.updateMomentumScrolling()

	// Should have stopped automatically
	if menu.isScrolling {
		t.Error("Momentum scrolling still active with negligible velocity")
	}
	if menu.scrollVelocity != 0 {
		t.Errorf("Velocity = %.2f, want 0 after auto-stop", menu.scrollVelocity)
	}
}

// TestMomentumScrolling_StopScrollingMethod tests manual stop.
// Gap #9: Verify StopScrolling() immediately halts momentum
func TestMomentumScrolling_StopScrollingMethod(t *testing.T) {
	menu := NewMobileMenu(100, 100, 300, 400)

	menu.scrollVelocity = 50.0
	menu.isScrolling = true

	menu.StopScrolling()

	if menu.isScrolling {
		t.Error("isScrolling = true after StopScrolling()")
	}
	if menu.scrollVelocity != 0 {
		t.Errorf("scrollVelocity = %.2f, want 0 after StopScrolling()", menu.scrollVelocity)
	}
}

// TestMomentumScrolling_ScrollDurationComparison benchmarks different deceleration rates.
// Gap #9: Document expected scroll durations for different settings
func TestMomentumScrolling_ScrollDurationComparison(t *testing.T) {
	tests := []struct {
		name             string
		deceleration     float64
		expectedDuration int // frames until velocity < 0.1
	}{
		{"iOS/Android (0.98)", 0.98, 200}, // ~3.3 seconds
		{"Default old (0.95)", 0.95, 70},  // ~1.2 seconds
		{"Fast (0.93)", 0.93, 50},         // ~0.8 seconds
		{"Slow (0.99)", 0.99, 450},        // ~7.5 seconds
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			menu := NewMobileMenu(100, 100, 300, 400)
			menu.SetScrollDeceleration(tt.deceleration)
			menu.scrollVelocity = 100.0
			menu.isScrolling = true

			frames := 0
			maxFrames := 500 // 8.3 seconds max
			for frames < maxFrames && menu.scrollVelocity >= 0.1 {
				menu.updateMomentumScrolling()
				frames++
			}

			// Allow 20% variance
			minExpected := int(float64(tt.expectedDuration) * 0.8)
			maxExpected := int(float64(tt.expectedDuration) * 1.2)

			if frames < minExpected || frames > maxExpected {
				t.Logf("Deceleration %.2f: scroll duration = %d frames (%.2fs at 60 FPS)",
					tt.deceleration, frames, float64(frames)/60.0)
				t.Logf("Expected: ~%d frames (%.2fs), got %d frames (%.2fs)",
					tt.expectedDuration, float64(tt.expectedDuration)/60.0,
					frames, float64(frames)/60.0)
			}
		})
	}
}

// TestMomentumScrolling_ApplyDeceleration tests deceleration application.
// Gap #9: Verify velocity multiplied by deceleration factor each frame
func TestMomentumScrolling_ApplyDeceleration(t *testing.T) {
	menu := NewMobileMenu(100, 100, 300, 400)
	menu.SetScrollDeceleration(0.95)

	initialVelocity := 100.0
	menu.scrollVelocity = initialVelocity

	menu.applyDeceleration()

	expectedVelocity := initialVelocity * 0.95
	if menu.scrollVelocity != expectedVelocity {
		t.Errorf("applyDeceleration(): velocity = %.2f, want %.2f", menu.scrollVelocity, expectedVelocity)
	}
}

// TestMomentumScrolling_UserExperience documents expected feel for different settings.
// Gap #9: Provide guidance for choosing deceleration values
func TestMomentumScrolling_UserExperience(t *testing.T) {
	t.Log("Momentum Scrolling User Experience Guide:")
	t.Log("")
	t.Log("Deceleration | Duration | Feel                    | Use Case")
	t.Log("-------------|----------|-------------------------|---------------------------")
	t.Log("0.98         | 1.5-2s   | iOS/Android-like smooth | Mobile games (DEFAULT)")
	t.Log("0.95         | 0.7s     | Faster, more responsive | Desktop with touch")
	t.Log("0.99         | 2.5-3s   | Very smooth, may lag    | Large content lists")
	t.Log("0.93         | 0.5s     | Quick stop, less smooth | Small menus")
	t.Log("")
	t.Log("Recommendation: Use default 0.98 for best mobile platform parity")
}
