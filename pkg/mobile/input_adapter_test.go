package mobile

import (
	"testing"
)

// TestMobileInputAdapter_Type verifies component type identification.
func TestMobileInputAdapter_Type(t *testing.T) {
	adapter := NewMobileInputAdapter(800, 600)
	if adapter.Type() != "input" {
		t.Errorf("Type() = %q, want %q", adapter.Type(), "input")
	}
}

// TestMobileInputAdapter_GetMovement verifies movement input retrieval.
func TestMobileInputAdapter_GetMovement(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*MobileInputAdapter)
		wantX float64
		wantY float64
	}{
		{
			name:  "no movement initially",
			setup: func(a *MobileInputAdapter) {},
			wantX: 0,
			wantY: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewMobileInputAdapter(800, 600)
			tt.setup(adapter)
			adapter.Update()

			gotX, gotY := adapter.GetMovement()
			if gotX != tt.wantX || gotY != tt.wantY {
				t.Errorf("GetMovement() = (%v, %v), want (%v, %v)", gotX, gotY, tt.wantX, tt.wantY)
			}
		})
	}
}

// TestMobileInputAdapter_IsActionPressed verifies action button state.
func TestMobileInputAdapter_IsActionPressed(t *testing.T) {
	adapter := NewMobileInputAdapter(800, 600)
	adapter.Update()

	if adapter.IsActionPressed() {
		t.Error("IsActionPressed() = true, want false initially")
	}
}

// TestMobileInputAdapter_IsActionJustPressed verifies single-frame action detection.
func TestMobileInputAdapter_IsActionJustPressed(t *testing.T) {
	adapter := NewMobileInputAdapter(800, 600)
	adapter.Update()

	// Should return false initially
	if adapter.IsActionJustPressed() {
		t.Error("IsActionJustPressed() = true, want false initially")
	}

	// Track last state after first call
	adapter.lastActionState = false

	// Simulate button press - note: without actual touch events,
	// we test the state tracking logic
	if adapter.IsActionJustPressed() {
		t.Error("IsActionJustPressed() = true without touch input")
	}
}

// TestMobileInputAdapter_IsAnyKeyPressed verifies any-input detection.
func TestMobileInputAdapter_IsAnyKeyPressed(t *testing.T) {
	adapter := NewMobileInputAdapter(800, 600)
	adapter.Update()

	// Should be false with no touch input
	if adapter.IsAnyKeyPressed() {
		t.Error("IsAnyKeyPressed() = true, want false with no input")
	}
}

// TestMobileInputAdapter_IsUseItemPressed verifies use item button state.
func TestMobileInputAdapter_IsUseItemPressed(t *testing.T) {
	adapter := NewMobileInputAdapter(800, 600)
	adapter.Update()

	if adapter.IsUseItemPressed() {
		t.Error("IsUseItemPressed() = true, want false initially")
	}
}

// TestMobileInputAdapter_IsUseItemJustPressed verifies single-frame use item detection.
func TestMobileInputAdapter_IsUseItemJustPressed(t *testing.T) {
	adapter := NewMobileInputAdapter(800, 600)
	adapter.Update()

	if adapter.IsUseItemJustPressed() {
		t.Error("IsUseItemJustPressed() = true, want false initially")
	}

	// Track last state
	adapter.lastUseState = false

	if adapter.IsUseItemJustPressed() {
		t.Error("IsUseItemJustPressed() = true without touch input")
	}
}

// TestMobileInputAdapter_IsSpellPressed verifies spell slot queries.
func TestMobileInputAdapter_IsSpellPressed(t *testing.T) {
	adapter := NewMobileInputAdapter(800, 600)
	adapter.Update()

	tests := []struct {
		slot int
		want bool
	}{
		{1, false},
		{2, false},
		{3, false},
		{4, false},
		{5, false},
		{0, false},
		{6, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := adapter.IsSpellPressed(tt.slot); got != tt.want {
				t.Errorf("IsSpellPressed(%d) = %v, want %v", tt.slot, got, tt.want)
			}
		})
	}
}

// TestMobileInputAdapter_GetMousePosition verifies simulated mouse position.
func TestMobileInputAdapter_GetMousePosition(t *testing.T) {
	adapter := NewMobileInputAdapter(800, 600)
	adapter.Update()

	x, y := adapter.GetMousePosition()

	// Should return center initially
	wantX, wantY := 400, 300
	if x != wantX || y != wantY {
		t.Errorf("GetMousePosition() = (%d, %d), want (%d, %d)", x, y, wantX, wantY)
	}
}

// TestMobileInputAdapter_GetMouseDelta verifies mouse delta is always zero.
func TestMobileInputAdapter_GetMouseDelta(t *testing.T) {
	adapter := NewMobileInputAdapter(800, 600)
	adapter.Update()

	dx, dy := adapter.GetMouseDelta()
	if dx != 0 || dy != 0 {
		t.Errorf("GetMouseDelta() = (%d, %d), want (0, 0)", dx, dy)
	}
}

// TestMobileInputAdapter_IsMousePressed verifies mouse press simulation.
func TestMobileInputAdapter_IsMousePressed(t *testing.T) {
	adapter := NewMobileInputAdapter(800, 600)
	adapter.Update()

	if adapter.IsMousePressed() {
		t.Error("IsMousePressed() = true, want false initially")
	}
}

// TestMobileInputAdapter_MenuNavigation verifies menu navigation methods.
func TestMobileInputAdapter_MenuNavigation(t *testing.T) {
	adapter := NewMobileInputAdapter(800, 600)
	adapter.Update()

	tests := []struct {
		name   string
		method func() bool
		want   bool
	}{
		{"IsMenuUpJustPressed", adapter.IsMenuUpJustPressed, false},
		{"IsMenuDownJustPressed", adapter.IsMenuDownJustPressed, false},
		{"IsMenuTabJustPressed", adapter.IsMenuTabJustPressed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.method(); got != tt.want {
				t.Errorf("%s() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// TestMobileInputAdapter_IsMenuConfirmJustPressed verifies menu confirm mapping.
func TestMobileInputAdapter_IsMenuConfirmJustPressed(t *testing.T) {
	adapter := NewMobileInputAdapter(800, 600)
	adapter.Update()

	// Should map to action button
	if adapter.IsMenuConfirmJustPressed() {
		t.Error("IsMenuConfirmJustPressed() = true, want false initially")
	}
}

// TestMobileInputAdapter_IsMenuBackJustPressed verifies menu back button.
func TestMobileInputAdapter_IsMenuBackJustPressed(t *testing.T) {
	adapter := NewMobileInputAdapter(800, 600)
	adapter.Update()

	if adapter.IsMenuBackJustPressed() {
		t.Error("IsMenuBackJustPressed() = true, want false initially")
	}
}

// TestMobileInputAdapter_SetVisible verifies visibility control.
func TestMobileInputAdapter_SetVisible(t *testing.T) {
	adapter := NewMobileInputAdapter(800, 600)

	// Default should be visible
	if !adapter.layout.Visible {
		t.Error("layout.Visible = false, want true initially")
	}

	adapter.SetVisible(false)
	if adapter.layout.Visible {
		t.Error("layout.Visible = true after SetVisible(false)")
	}

	adapter.SetVisible(true)
	if !adapter.layout.Visible {
		t.Error("layout.Visible = false after SetVisible(true)")
	}
}

// TestMobileInputAdapter_GetLayout verifies layout access.
func TestMobileInputAdapter_GetLayout(t *testing.T) {
	adapter := NewMobileInputAdapter(800, 600)

	layout := adapter.GetLayout()
	if layout == nil {
		t.Fatal("GetLayout() = nil, want non-nil")
	}

	if layout != adapter.layout {
		t.Error("GetLayout() returned different layout instance")
	}
}

// TestMobileInputAdapter_SetMovement verifies SetMovement is no-op.
func TestMobileInputAdapter_SetMovement(t *testing.T) {
	adapter := NewMobileInputAdapter(800, 600)
	adapter.Update()

	// Should not panic
	adapter.SetMovement(1.0, -1.0)

	// Movement should still be zero (touch-controlled)
	x, y := adapter.GetMovement()
	if x != 0 || y != 0 {
		t.Errorf("GetMovement() = (%v, %v), want (0, 0) after SetMovement", x, y)
	}
}

// TestMobileInputAdapter_SetActionPressed verifies SetActionPressed is no-op.
func TestMobileInputAdapter_SetActionPressed(t *testing.T) {
	adapter := NewMobileInputAdapter(800, 600)
	adapter.Update()

	// Should not panic
	adapter.SetActionPressed(true)

	// Action should still be false (touch-controlled)
	if adapter.IsActionPressed() {
		t.Error("IsActionPressed() = true after SetActionPressed, want false (touch-controlled)")
	}
}

// TestMobileInputAdapter_ScreenDimensions verifies screen dimension handling.
func TestMobileInputAdapter_ScreenDimensions(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"landscape", 1280, 720},
		{"portrait", 720, 1280},
		{"square", 800, 800},
		{"small", 320, 240},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewMobileInputAdapter(tt.width, tt.height)

			if adapter.screenWidth != tt.width {
				t.Errorf("screenWidth = %d, want %d", adapter.screenWidth, tt.width)
			}
			if adapter.screenHeight != tt.height {
				t.Errorf("screenHeight = %d, want %d", adapter.screenHeight, tt.height)
			}

			// Verify mouse position is initialized to center
			x, y := adapter.GetMousePosition()
			wantX, wantY := tt.width/2, tt.height/2
			if x != wantX || y != wantY {
				t.Errorf("GetMousePosition() = (%d, %d), want (%d, %d)", x, y, wantX, wantY)
			}
		})
	}
}

// TestMobileInputAdapter_ComponentInterface verifies InputProvider interface compliance.
func TestMobileInputAdapter_ComponentInterface(t *testing.T) {
	adapter := NewMobileInputAdapter(800, 600)

	// Verify all InputProvider methods are callable
	adapter.GetMovement()
	adapter.IsActionPressed()
	adapter.IsActionJustPressed()
	adapter.IsAnyKeyPressed()
	adapter.IsUseItemPressed()
	adapter.IsUseItemJustPressed()
	adapter.IsSpellPressed(1)
	adapter.GetMousePosition()
	adapter.GetMouseDelta()
	adapter.IsMousePressed()
	adapter.SetMovement(0, 0)
	adapter.SetActionPressed(false)
	adapter.IsMenuUpJustPressed()
	adapter.IsMenuDownJustPressed()
	adapter.IsMenuConfirmJustPressed()
	adapter.IsMenuBackJustPressed()
	adapter.IsMenuTabJustPressed()

	// Verify Component interface
	if adapter.Type() != "input" {
		t.Errorf("Type() = %q, want %q", adapter.Type(), "input")
	}
}
