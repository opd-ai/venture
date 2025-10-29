package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// ===== KEYBOARD INPUT METHOD TESTS =====

// TestInputSystem_IsKeyReleased tests key release detection (BUG-001 fix).
func TestInputSystem_IsKeyReleased(t *testing.T) {
	inputSys := NewInputSystem()

	// Note: Cannot actually test Ebiten input without game loop running
	// This test verifies the method exists and is callable

	// The method should be defined and callable
	result := inputSys.IsKeyReleased(27) // Escape key code
	if result {
		t.Log("Key released (unexpected in test env)")
	}
}

// TestInputSystem_IsKeyJustReleased tests key release alias (BUG-002 fix).
func TestInputSystem_IsKeyJustReleased(t *testing.T) {
	inputSys := NewInputSystem()

	// Should be callable
	result := inputSys.IsKeyJustReleased(27)
	if result {
		t.Log("Key just released (unexpected in test env)")
	}
}

// TestInputSystem_GetPressedKeys tests bulk key query (BUG-003 fix).
func TestInputSystem_GetPressedKeys(t *testing.T) {
	inputSys := NewInputSystem()

	// Should return empty slice when no keys pressed
	keys := inputSys.GetPressedKeys()
	if keys == nil {
		t.Error("GetPressedKeys should return non-nil slice")
	}

	// Note: In test environment with no Ebiten game loop, will always be empty
	if len(keys) != 0 {
		t.Logf("Got %d pressed keys (expected 0 in test env): %v", len(keys), keys)
	}
}

// TestInputSystem_IsAnyKeyPressed tests any-key detection (BUG-021 fix).
func TestInputSystem_IsAnyKeyPressed(t *testing.T) {
	inputSys := NewInputSystem()

	// In test environment, should return false
	if inputSys.IsAnyKeyPressed() {
		t.Error("IsAnyKeyPressed should return false in test environment")
	}
}

// TestInputSystem_GetAnyPressedKey tests single key query (BUG-021 fix).
func TestInputSystem_GetAnyPressedKey(t *testing.T) {
	inputSys := NewInputSystem()

	// Should return (0, false) when no keys pressed
	key, pressed := inputSys.GetAnyPressedKey()
	if pressed {
		t.Error("GetAnyPressedKey should return false in test environment")
	}
	if key != 0 {
		t.Errorf("Expected key 0, got %d", key)
	}
}

// TestInputSystem_ModifierKeys tests modifier key detection (BUG-022 fix).
func TestInputSystem_ModifierKeys(t *testing.T) {
	inputSys := NewInputSystem()

	tests := []struct {
		name   string
		method func() bool
	}{
		{"IsShiftPressed", inputSys.IsShiftPressed},
		{"IsControlPressed", inputSys.IsControlPressed},
		{"IsAltPressed", inputSys.IsAltPressed},
		{"IsSuperPressed", inputSys.IsSuperPressed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should return false in test environment
			if tt.method() {
				t.Errorf("%s should return false in test environment", tt.name)
			}
		})
	}
}

// ===== MOUSE INPUT METHOD TESTS =====

// TestInputSystem_IsMouseButtonJustPressed tests edge-triggered click (BUG-004 fix).
func TestInputSystem_IsMouseButtonJustPressed(t *testing.T) {
	inputSys := NewInputSystem()

	// Test with left mouse button (value 0)
	result := inputSys.IsMouseButtonJustPressed(0)
	if result {
		t.Log("Mouse button just pressed (unexpected in test env)")
	}
}

// TestInputSystem_IsMouseButtonReleased tests button release detection (BUG-005 fix).
func TestInputSystem_IsMouseButtonReleased(t *testing.T) {
	inputSys := NewInputSystem()

	result := inputSys.IsMouseButtonReleased(0)
	if result {
		t.Log("Mouse button released (unexpected in test env)")
	}
}

// TestInputSystem_IsMouseButtonJustReleased tests release alias (BUG-006 fix).
func TestInputSystem_IsMouseButtonJustReleased(t *testing.T) {
	inputSys := NewInputSystem()

	result := inputSys.IsMouseButtonJustReleased(0)
	if result {
		t.Log("Mouse button just released (unexpected in test env)")
	}
}

// TestInputSystem_GetMousePosition tests position query.
func TestInputSystem_GetMousePosition(t *testing.T) {
	inputSys := NewInputSystem()

	// Should be callable
	x, y := inputSys.GetMousePosition()
	if x != 0 || y != 0 {
		t.Logf("Mouse position: (%d, %d)", x, y)
	}
}

// TestInputSystem_GetCursorPosition tests cursor alias.
func TestInputSystem_GetCursorPosition(t *testing.T) {
	inputSys := NewInputSystem()

	x, y := inputSys.GetCursorPosition()
	if x != 0 || y != 0 {
		t.Logf("Cursor position: (%d, %d)", x, y)
	}
}

// TestInputSystem_GetMouseDelta tests delta tracking (BUG-008 fix).
func TestInputSystem_GetMouseDelta(t *testing.T) {
	inputSys := NewInputSystem()

	// Initial delta should be zero
	dx, dy := inputSys.GetMouseDelta()
	if dx != 0 || dy != 0 {
		t.Errorf("Initial mouse delta should be (0, 0), got (%d, %d)", dx, dy)
	}

	// After Update(), delta should reflect movement
	inputSys.Update([]*Entity{}, 0.016)
	dx, dy = inputSys.GetMouseDelta()

	// In test environment, will likely be 0, but method should work
	t.Logf("Mouse delta after update: (%d, %d)", dx, dy)
}

// TestInputSystem_GetMouseWheel tests scroll wheel (BUG-007 fix).
func TestInputSystem_GetMouseWheel(t *testing.T) {
	inputSys := NewInputSystem()

	// Should be callable
	wx, wy := inputSys.GetMouseWheel()
	if wx != 0 || wy != 0 {
		t.Logf("Mouse wheel: (%f, %f)", wx, wy)
	}
}

// ===== KEY BINDING MANAGEMENT TESTS =====

// TestInputSystem_SetKeyBinding tests comprehensive binding API (BUG-019 fix).
func TestInputSystem_SetKeyBinding(t *testing.T) {
	inputSys := NewInputSystem()

	tests := []struct {
		action   string
		key      ebiten.Key
		expected bool
	}{
		// Valid actions
		{"up", ebiten.KeyW, true},
		{"down", ebiten.KeyS, true},
		{"left", ebiten.KeyA, true},
		{"right", ebiten.KeyD, true},
		{"action", ebiten.KeySpace, true},
		{"useitem", ebiten.KeyE, true},
		{"inventory", ebiten.KeyI, true},
		{"character", ebiten.KeyC, true},
		{"skills", ebiten.KeyK, true},
		{"quests", ebiten.KeyJ, true},
		{"map", ebiten.KeyM, true},
		{"help", ebiten.KeyEscape, true},
		{"quicksave", ebiten.KeyF5, true},
		{"quickload", ebiten.KeyF9, true},
		{"cycletargets", ebiten.KeyTab, true},
		// Invalid action
		{"invalid", ebiten.Key0, false},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			result := inputSys.SetKeyBinding(tt.action, tt.key)
			if result != tt.expected {
				t.Errorf("SetKeyBinding(%q, %d) = %v, want %v",
					tt.action, tt.key, result, tt.expected)
			}

			// Verify binding was set for valid actions
			if tt.expected {
				key, ok := inputSys.GetKeyBinding(tt.action)
				if !ok {
					t.Errorf("GetKeyBinding(%q) should return true after SetKeyBinding", tt.action)
				}
				if key != tt.key {
					t.Errorf("GetKeyBinding(%q) = %d, want %d", tt.action, key, tt.key)
				}
			}
		})
	}
}

// TestInputSystem_GetKeyBinding tests binding query (BUG-020 fix).
func TestInputSystem_GetKeyBinding(t *testing.T) {
	inputSys := NewInputSystem()

	// Test default bindings exist
	tests := []struct {
		action string
		valid  bool
	}{
		{"up", true},
		{"down", true},
		{"left", true},
		{"right", true},
		{"action", true},
		{"useitem", true},
		{"inventory", true},
		{"character", true},
		{"skills", true},
		{"quests", true},
		{"map", true},
		{"help", true},
		{"quicksave", true},
		{"quickload", true},
		{"cycletargets", true},
		{"invalid_action", false},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			_, ok := inputSys.GetKeyBinding(tt.action)
			if ok != tt.valid {
				t.Errorf("GetKeyBinding(%q) returned ok=%v, want %v", tt.action, ok, tt.valid)
			}
			// Note: Don't check if key == 0, as ebiten.KeyA has value 0
			// If ok is true, the binding exists regardless of key value
		})
	}
}

// TestInputSystem_GetAllKeyBindings tests full binding query (BUG-020 fix).
func TestInputSystem_GetAllKeyBindings(t *testing.T) {
	inputSys := NewInputSystem()

	bindings := inputSys.GetAllKeyBindings()

	// Should have all 15 actions
	expectedActions := []string{
		"up", "down", "left", "right",
		"action", "useitem",
		"inventory", "character", "skills", "quests", "map",
		"help", "quicksave", "quickload", "cycletargets",
	}

	if len(bindings) != len(expectedActions) {
		t.Errorf("GetAllKeyBindings returned %d bindings, want %d",
			len(bindings), len(expectedActions))
	}

	for _, action := range expectedActions {
		if _, ok := bindings[action]; !ok {
			t.Errorf("GetAllKeyBindings missing action: %q", action)
		}
	}
}

// TestInputSystem_SetKeyBinding_UpdatesGetAll tests consistency (BUG-019/020).
func TestInputSystem_SetKeyBinding_UpdatesGetAll(t *testing.T) {
	inputSys := NewInputSystem()

	// Set custom binding
	customKey := ebiten.KeyF1
	if !inputSys.SetKeyBinding("action", customKey) {
		t.Fatal("SetKeyBinding failed")
	}

	// Verify it appears in GetAllKeyBindings
	bindings := inputSys.GetAllKeyBindings()
	if bindings["action"] != customKey {
		t.Errorf("GetAllKeyBindings[\"action\"] = %v, want %v",
			bindings["action"], customKey)
	}

	// Verify it appears in GetKeyBinding
	key, ok := inputSys.GetKeyBinding("action")
	if !ok {
		t.Error("GetKeyBinding(\"action\") should return true")
	}
	if key != customKey {
		t.Errorf("GetKeyBinding(\"action\") = %v, want %v", key, customKey)
	}
}

// ===== MOBILE INPUT TESTS =====

// TestInputSystem_MobileInitialization tests auto-init (BUG-023 fix).
func TestInputSystem_MobileInitialization(t *testing.T) {
	inputSys := NewInputSystem()

	// Simulate mobile enabled but not initialized
	inputSys.SetMobileEnabled(true)

	// virtualControls should be nil initially
	if inputSys.virtualControls != nil {
		t.Error("virtualControls should be nil before initialization")
	}

	// Call Update - should auto-initialize (BUG-023 fix)
	inputSys.Update([]*Entity{}, 0.016)

	// virtualControls should now be initialized
	if inputSys.virtualControls == nil {
		t.Error("virtualControls should be auto-initialized after Update()")
	}
}

// TestInputSystem_MouseDeltaTracking tests delta state management (BUG-010 fix).
func TestInputSystem_MouseDeltaTracking(t *testing.T) {
	inputSys := NewInputSystem()

	// Initial state
	if inputSys.lastMouseX != 0 || inputSys.lastMouseY != 0 {
		t.Errorf("Initial mouse position should be (0, 0), got (%d, %d)",
			inputSys.lastMouseX, inputSys.lastMouseY)
	}

	// Update should track position
	inputSys.Update([]*Entity{}, 0.016)

	// lastMouse values should be set to current cursor position
	// (which is likely 0,0 in test env, but that's fine)
	t.Logf("After update, lastMouse = (%d, %d)", inputSys.lastMouseX, inputSys.lastMouseY)
}

// ===== INTEGRATION TESTS =====

// TestInputSystem_FullAPIAvailability tests all methods are accessible.
func TestInputSystem_FullAPIAvailability(t *testing.T) {
	inputSys := NewInputSystem()

	// Keyboard methods
	_ = inputSys.IsKeyPressed(27)
	_ = inputSys.IsKeyJustPressed(27)
	_ = inputSys.IsKeyReleased(27)
	_ = inputSys.IsKeyJustReleased(27)
	_ = inputSys.GetPressedKeys()
	_ = inputSys.IsAnyKeyPressed()
	_, _ = inputSys.GetAnyPressedKey()
	_ = inputSys.IsShiftPressed()
	_ = inputSys.IsControlPressed()
	_ = inputSys.IsAltPressed()
	_ = inputSys.IsSuperPressed()

	// Mouse methods
	_ = inputSys.IsMouseButtonPressed(0)
	_ = inputSys.IsMouseButtonJustPressed(0)
	_ = inputSys.IsMouseButtonReleased(0)
	_ = inputSys.IsMouseButtonJustReleased(0)
	_, _ = inputSys.GetMousePosition()
	_, _ = inputSys.GetCursorPosition()
	_, _ = inputSys.GetMouseDelta()
	_, _ = inputSys.GetMouseWheel()

	// Key binding methods
	_ = inputSys.SetKeyBinding("up", 100)
	_, _ = inputSys.GetKeyBinding("up")
	_ = inputSys.GetAllKeyBindings()

	// If we get here, all methods are accessible
	t.Log("All input methods are accessible")
}

// TestInputSystem_OriginalFunctionalityIntact tests no regression.
func TestInputSystem_OriginalFunctionalityIntact(t *testing.T) {
	inputSys := NewInputSystem()

	// Test original methods still work
	// Note: HelpSystem test commented out - requires EbitenHelpSystem which needs Ebiten
	// inputSys.SetHelpSystem(&EbitenHelpSystem{})
	inputSys.SetQuickSaveCallback(func() error { return nil })
	inputSys.SetQuickLoadCallback(func() error { return nil })
	inputSys.SetInventoryCallback(func() {})
	inputSys.SetCharacterCallback(func() {})
	inputSys.SetSkillsCallback(func() {})
	inputSys.SetQuestsCallback(func() {})
	inputSys.SetMapCallback(func() {})
	inputSys.SetCycleTargetsCallback(func() {})
	inputSys.SetMenuToggleCallback(func() {})

	// Test Update still works
	entities := []*Entity{}
	inputSys.Update(entities, 0.016)

	t.Log("Original functionality intact")
}

// Benchmark tests for performance validation

// BenchmarkInputSystem_IsKeyPressed benchmarks continuous state check.
func BenchmarkInputSystem_IsKeyPressed(b *testing.B) {
	inputSys := NewInputSystem()
	for i := 0; i < b.N; i++ {
		_ = inputSys.IsKeyPressed(27)
	}
}

// BenchmarkInputSystem_GetMouseDelta benchmarks delta calculation.
func BenchmarkInputSystem_GetMouseDelta(b *testing.B) {
	inputSys := NewInputSystem()
	inputSys.Update([]*Entity{}, 0.016) // Initialize
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = inputSys.GetMouseDelta()
	}
}

// BenchmarkInputSystem_GetAllKeyBindings benchmarks map allocation.
func BenchmarkInputSystem_GetAllKeyBindings(b *testing.B) {
	inputSys := NewInputSystem()
	for i := 0; i < b.N; i++ {
		_ = inputSys.GetAllKeyBindings()
	}
}

// BenchmarkInputSystem_Update benchmarks full update cycle.
func BenchmarkInputSystem_Update(b *testing.B) {
	inputSys := NewInputSystem()
	entities := []*Entity{}
	for i := 0; i < b.N; i++ {
		inputSys.Update(entities, 0.016)
	}
}

// ===== TOUCH INPUT TESTS =====

// TestInputSystem_TouchInputInitialization tests that touch input is properly initialized
// for touch-capable platforms including WASM/browser builds.
func TestInputSystem_TouchInputInitialization(t *testing.T) {
	inputSys := NewInputSystem()

	// useTouchInput should be set based on platform capability at initialization
	// For testing purposes, we can verify the flag exists and can be set
	if inputSys.touchHandler == nil {
		t.Error("TouchHandler should be initialized for all platforms")
	}

	// Test that virtual controls can be initialized explicitly
	inputSys.InitializeVirtualControls(800, 600)

	// If useTouchInput is true, virtual controls should be created
	if inputSys.useTouchInput && inputSys.virtualControls == nil {
		t.Error("Virtual controls should be initialized when useTouchInput is true")
	}

	t.Log("Touch input initialization verified")
}

// TestInputSystem_VirtualControlsAutoInit tests that virtual controls are
// automatically initialized when needed (WASM support fix).
func TestInputSystem_VirtualControlsAutoInit(t *testing.T) {
	inputSys := NewInputSystem()

	// Manually set useTouchInput to simulate WASM platform
	inputSys.useTouchInput = true
	inputSys.virtualControls = nil // Clear any existing controls

	// Create a test entity with input component
	entity := NewEntity()
	inputComp := &EbitenInput{}
	entity.AddComponent(inputComp)
	entities := []*Entity{entity}

	// Update should auto-initialize virtual controls
	inputSys.Update(entities, 0.016)

	// Virtual controls should now be initialized
	if inputSys.virtualControls == nil {
		t.Error("Virtual controls should be auto-initialized during Update when useTouchInput is true")
	}

	t.Log("Virtual controls auto-initialization verified")
}

// TestInputSystem_DrawVirtualControls tests that virtual controls can be drawn
// when touch input is enabled (WASM support).
func TestInputSystem_DrawVirtualControls(t *testing.T) {
	inputSys := NewInputSystem()
	inputSys.useTouchInput = true
	inputSys.InitializeVirtualControls(800, 600)

	// DrawVirtualControls should be callable without panicking
	// Note: We can't actually test drawing without Ebiten context,
	// but we can verify the method exists and handles nil screen gracefully
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("DrawVirtualControls should not panic: %v", r)
		}
	}()

	// This will likely do nothing or fail gracefully in test environment
	// The important part is that it doesn't crash
	var screen *ebiten.Image = nil
	inputSys.DrawVirtualControls(screen)

	t.Log("DrawVirtualControls method verified")
}
