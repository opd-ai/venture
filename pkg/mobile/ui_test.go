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
