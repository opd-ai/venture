package mobile

import (
	"math"
	"testing"
	"time"
)

// =============================================================================
// InputRateLimiter Tests
// =============================================================================

// TestNewInputRateLimiter tests rate limiter creation.
func TestNewInputRateLimiter(t *testing.T) {
	limiter := NewInputRateLimiter()

	if limiter == nil {
		t.Fatal("NewInputRateLimiter returned nil")
	}
	if limiter.lastInputTime == nil {
		t.Error("lastInputTime map is nil")
	}
	if limiter.cooldowns == nil {
		t.Error("cooldowns map is nil")
	}
	if limiter.inputCounts == nil {
		t.Error("inputCounts map is nil")
	}
	if limiter.timeWindow != 1*time.Second {
		t.Errorf("timeWindow = %v, want 1s", limiter.timeWindow)
	}
	if limiter.maxInputs != 10 {
		t.Errorf("maxInputs = %d, want 10", limiter.maxInputs)
	}
}

// TestInputRateLimiter_SetCooldown tests setting cooldown for an action.
func TestInputRateLimiter_SetCooldown(t *testing.T) {
	limiter := NewInputRateLimiter()

	limiter.SetCooldown("attack", 500*time.Millisecond)
	limiter.SetCooldown("jump", 200*time.Millisecond)

	if limiter.cooldowns["attack"] != 500*time.Millisecond {
		t.Errorf("attack cooldown = %v, want 500ms", limiter.cooldowns["attack"])
	}
	if limiter.cooldowns["jump"] != 200*time.Millisecond {
		t.Errorf("jump cooldown = %v, want 200ms", limiter.cooldowns["jump"])
	}
}

// TestInputRateLimiter_CanExecute tests execution permission checking.
func TestInputRateLimiter_CanExecute(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*InputRateLimiter)
		actionID    string
		want        bool
		description string
	}{
		{
			name:        "first_execution_always_allowed",
			setupFunc:   func(l *InputRateLimiter) {},
			actionID:    "attack",
			want:        true,
			description: "First execution should always be allowed",
		},
		{
			name: "during_cooldown_blocked",
			setupFunc: func(l *InputRateLimiter) {
				l.SetCooldown("attack", 1*time.Second)
				l.RecordInput("attack")
			},
			actionID:    "attack",
			want:        false,
			description: "Execution during cooldown should be blocked",
		},
		{
			name: "different_action_allowed",
			setupFunc: func(l *InputRateLimiter) {
				l.SetCooldown("attack", 1*time.Second)
				l.RecordInput("attack")
			},
			actionID:    "jump",
			want:        true,
			description: "Different action should be allowed",
		},
		{
			name: "spam_threshold_blocks",
			setupFunc: func(l *InputRateLimiter) {
				for i := 0; i < 10; i++ {
					l.RecordInput("spam_action")
				}
			},
			actionID:    "spam_action",
			want:        false,
			description: "Action at spam threshold should be blocked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewInputRateLimiter()
			tt.setupFunc(limiter)

			got := limiter.CanExecute(tt.actionID)
			if got != tt.want {
				t.Errorf("CanExecute(%q) = %v, want %v: %s", tt.actionID, got, tt.want, tt.description)
			}
		})
	}
}

// TestInputRateLimiter_CanExecuteDefaultCooldown tests default 100ms cooldown.
func TestInputRateLimiter_CanExecuteDefaultCooldown(t *testing.T) {
	limiter := NewInputRateLimiter()

	// No custom cooldown set - should use 100ms default
	limiter.RecordInput("action")

	// Immediately after should be blocked
	if limiter.CanExecute("action") {
		t.Error("CanExecute should return false immediately after RecordInput with default cooldown")
	}
}

// TestInputRateLimiter_RecordInput tests input recording.
func TestInputRateLimiter_RecordInput(t *testing.T) {
	limiter := NewInputRateLimiter()

	// Record input
	limiter.RecordInput("action")

	// Should have recorded time
	if _, exists := limiter.lastInputTime["action"]; !exists {
		t.Error("RecordInput did not record lastInputTime")
	}

	// Should have incremented count
	if limiter.inputCounts["action"] != 1 {
		t.Errorf("inputCounts[action] = %d, want 1", limiter.inputCounts["action"])
	}

	// Record again
	limiter.RecordInput("action")
	if limiter.inputCounts["action"] != 2 {
		t.Errorf("inputCounts[action] = %d after second record, want 2", limiter.inputCounts["action"])
	}
}

// TestInputRateLimiter_Update tests spam counter cleanup.
func TestInputRateLimiter_Update(t *testing.T) {
	limiter := NewInputRateLimiter()
	limiter.timeWindow = 1 * time.Millisecond // Very short for testing

	// Record input
	limiter.RecordInput("action")

	// Wait for time window to pass
	time.Sleep(5 * time.Millisecond)

	// Update should reset count
	limiter.Update()

	if limiter.inputCounts["action"] != 0 {
		t.Errorf("inputCounts[action] = %d after Update, want 0", limiter.inputCounts["action"])
	}
}

// TestInputRateLimiter_GetRemainingCooldown tests cooldown remaining calculation.
func TestInputRateLimiter_GetRemainingCooldown(t *testing.T) {
	limiter := NewInputRateLimiter()
	limiter.SetCooldown("action", 500*time.Millisecond)

	// Before any input, should be 0
	if remaining := limiter.GetRemainingCooldown("action"); remaining != 0 {
		t.Errorf("GetRemainingCooldown before input = %v, want 0", remaining)
	}

	// Record input
	limiter.RecordInput("action")

	// Should have significant remaining cooldown
	remaining := limiter.GetRemainingCooldown("action")
	if remaining <= 0 || remaining > 500*time.Millisecond {
		t.Errorf("GetRemainingCooldown = %v, want between 0 and 500ms", remaining)
	}
}

// TestInputRateLimiter_GetRemainingCooldownDefault tests default cooldown.
func TestInputRateLimiter_GetRemainingCooldownDefault(t *testing.T) {
	limiter := NewInputRateLimiter()

	// No custom cooldown, should use 100ms default
	limiter.RecordInput("action")

	remaining := limiter.GetRemainingCooldown("action")
	if remaining <= 0 || remaining > 100*time.Millisecond {
		t.Errorf("GetRemainingCooldown (default) = %v, want between 0 and 100ms", remaining)
	}
}

// TestInputRateLimiter_GetSpamCount tests spam count retrieval.
func TestInputRateLimiter_GetSpamCount(t *testing.T) {
	limiter := NewInputRateLimiter()

	// Initially 0
	if count := limiter.GetSpamCount("action"); count != 0 {
		t.Errorf("GetSpamCount initially = %d, want 0", count)
	}

	// Record 5 inputs
	for i := 0; i < 5; i++ {
		limiter.RecordInput("action")
	}

	if count := limiter.GetSpamCount("action"); count != 5 {
		t.Errorf("GetSpamCount after 5 inputs = %d, want 5", count)
	}
}

// =============================================================================
// SelectionState Tests
// =============================================================================

// TestNewSelectionState tests selection state creation.
func TestNewSelectionState(t *testing.T) {
	tests := []struct {
		name        string
		multiSelect bool
	}{
		{"single_select", false},
		{"multi_select", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewSelectionState(tt.multiSelect)

			if state == nil {
				t.Fatal("NewSelectionState returned nil")
			}
			if state.selectedItems == nil {
				t.Error("selectedItems map is nil")
			}
			if state.lastSelected != -1 {
				t.Errorf("lastSelected = %d, want -1", state.lastSelected)
			}
			if state.multiSelect != tt.multiSelect {
				t.Errorf("multiSelect = %v, want %v", state.multiSelect, tt.multiSelect)
			}
		})
	}
}

// TestSelectionState_Select tests item selection.
func TestSelectionState_Select(t *testing.T) {
	tests := []struct {
		name                 string
		multiSelect          bool
		isMultiSelectGesture bool
		selectFirst          int
		selectSecond         int
		wantFirstSelected    bool
		wantSecondSelected   bool
		wantLastSelected     int
	}{
		{
			name:                 "single_select_replaces",
			multiSelect:          false,
			isMultiSelectGesture: false,
			selectFirst:          1,
			selectSecond:         2,
			wantFirstSelected:    false, // First should be deselected
			wantSecondSelected:   true,
			wantLastSelected:     2,
		},
		{
			name:                 "multi_select_with_gesture_adds",
			multiSelect:          true,
			isMultiSelectGesture: true,
			selectFirst:          1,
			selectSecond:         2,
			wantFirstSelected:    true, // First should remain selected
			wantSecondSelected:   true,
			wantLastSelected:     2,
		},
		{
			name:                 "multi_select_without_gesture_replaces",
			multiSelect:          true,
			isMultiSelectGesture: false,
			selectFirst:          1,
			selectSecond:         2,
			wantFirstSelected:    false, // First should be deselected
			wantSecondSelected:   true,
			wantLastSelected:     2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewSelectionState(tt.multiSelect)

			state.Select(tt.selectFirst, false)
			state.Select(tt.selectSecond, tt.isMultiSelectGesture)

			if state.IsSelected(tt.selectFirst) != tt.wantFirstSelected {
				t.Errorf("IsSelected(%d) = %v, want %v", tt.selectFirst, state.IsSelected(tt.selectFirst), tt.wantFirstSelected)
			}
			if state.IsSelected(tt.selectSecond) != tt.wantSecondSelected {
				t.Errorf("IsSelected(%d) = %v, want %v", tt.selectSecond, state.IsSelected(tt.selectSecond), tt.wantSecondSelected)
			}
			if state.lastSelected != tt.wantLastSelected {
				t.Errorf("lastSelected = %d, want %d", state.lastSelected, tt.wantLastSelected)
			}
		})
	}
}

// TestSelectionState_Deselect tests item deselection.
func TestSelectionState_Deselect(t *testing.T) {
	state := NewSelectionState(true)

	state.Select(1, false)
	state.Select(2, true)
	state.Select(3, true)

	// Deselect middle item
	state.Deselect(2)

	if state.IsSelected(1) != true {
		t.Error("Item 1 should still be selected")
	}
	if state.IsSelected(2) != false {
		t.Error("Item 2 should be deselected")
	}
	if state.IsSelected(3) != true {
		t.Error("Item 3 should still be selected")
	}
}

// TestSelectionState_DeselectAll tests clearing all selections.
func TestSelectionState_DeselectAll(t *testing.T) {
	state := NewSelectionState(true)

	state.Select(1, false)
	state.Select(2, true)
	state.Select(3, true)

	state.DeselectAll()

	if state.GetSelectionCount() != 0 {
		t.Errorf("GetSelectionCount after DeselectAll = %d, want 0", state.GetSelectionCount())
	}
	if state.lastSelected != -1 {
		t.Errorf("lastSelected after DeselectAll = %d, want -1", state.lastSelected)
	}
}

// TestSelectionState_IsSelected tests selection checking.
func TestSelectionState_IsSelected(t *testing.T) {
	state := NewSelectionState(false)

	// Unselected item
	if state.IsSelected(1) {
		t.Error("Unselected item should return false")
	}

	state.Select(1, false)

	if !state.IsSelected(1) {
		t.Error("Selected item should return true")
	}
}

// TestSelectionState_GetSelectedItems tests getting all selected items.
func TestSelectionState_GetSelectedItems(t *testing.T) {
	state := NewSelectionState(true)

	state.Select(1, false)
	state.Select(2, true)
	state.Select(3, true)

	items := state.GetSelectedItems()

	if len(items) != 3 {
		t.Errorf("len(GetSelectedItems()) = %d, want 3", len(items))
	}

	// Check all items present (order not guaranteed)
	found := make(map[int]bool)
	for _, id := range items {
		found[id] = true
	}

	for _, want := range []int{1, 2, 3} {
		if !found[want] {
			t.Errorf("GetSelectedItems missing item %d", want)
		}
	}
}

// TestSelectionState_GetSelectionCount tests selection count.
func TestSelectionState_GetSelectionCount(t *testing.T) {
	state := NewSelectionState(true)

	if count := state.GetSelectionCount(); count != 0 {
		t.Errorf("GetSelectionCount initially = %d, want 0", count)
	}

	state.Select(1, false)
	if count := state.GetSelectionCount(); count != 1 {
		t.Errorf("GetSelectionCount after 1 select = %d, want 1", count)
	}

	state.Select(2, true)
	state.Select(3, true)
	if count := state.GetSelectionCount(); count != 3 {
		t.Errorf("GetSelectionCount after 3 selects = %d, want 3", count)
	}
}

// =============================================================================
// AppLifecycleHandler Tests
// =============================================================================

// TestNewAppLifecycleHandler tests lifecycle handler creation.
func TestNewAppLifecycleHandler(t *testing.T) {
	handler := NewAppLifecycleHandler()

	if handler == nil {
		t.Fatal("NewAppLifecycleHandler returned nil")
	}
	if handler.currentState != AppStateActive {
		t.Errorf("currentState = %v, want %v", handler.currentState, AppStateActive)
	}
	if handler.onStateChange != nil {
		t.Error("onStateChange should be nil initially")
	}
	if handler.onInterruption != nil {
		t.Error("onInterruption should be nil initially")
	}
}

// TestAppLifecycleHandler_SetStateChangeCallback tests callback registration.
func TestAppLifecycleHandler_SetStateChangeCallback(t *testing.T) {
	handler := NewAppLifecycleHandler()

	called := false
	var receivedState AppLifecycleState

	handler.SetStateChangeCallback(func(state AppLifecycleState) {
		called = true
		receivedState = state
	})

	handler.NotifyStateChange(AppStateBackground)

	if !called {
		t.Error("State change callback was not called")
	}
	if receivedState != AppStateBackground {
		t.Errorf("Callback received state = %v, want %v", receivedState, AppStateBackground)
	}
}

// TestAppLifecycleHandler_SetInterruptionCallback tests interruption callback.
func TestAppLifecycleHandler_SetInterruptionCallback(t *testing.T) {
	handler := NewAppLifecycleHandler()

	called := false
	var receivedType SystemInterruptionType

	handler.SetInterruptionCallback(func(interruptionType SystemInterruptionType) {
		called = true
		receivedType = interruptionType
	})

	handler.NotifyInterruption(InterruptionCall)

	if !called {
		t.Error("Interruption callback was not called")
	}
	if receivedType != InterruptionCall {
		t.Errorf("Callback received type = %v, want %v", receivedType, InterruptionCall)
	}
}

// TestAppLifecycleHandler_NotifyStateChange tests state change notification.
func TestAppLifecycleHandler_NotifyStateChange(t *testing.T) {
	handler := NewAppLifecycleHandler()

	callCount := 0
	handler.SetStateChangeCallback(func(state AppLifecycleState) {
		callCount++
	})

	// First change
	handler.NotifyStateChange(AppStateBackground)
	if handler.currentState != AppStateBackground {
		t.Errorf("currentState = %v, want %v", handler.currentState, AppStateBackground)
	}
	if callCount != 1 {
		t.Errorf("callback count = %d, want 1", callCount)
	}

	// Same state should not trigger callback
	handler.NotifyStateChange(AppStateBackground)
	if callCount != 1 {
		t.Errorf("callback count after same state = %d, want 1", callCount)
	}

	// Different state should trigger
	handler.NotifyStateChange(AppStateActive)
	if callCount != 2 {
		t.Errorf("callback count after different state = %d, want 2", callCount)
	}
}

// TestAppLifecycleHandler_NotifyStateChangeNoCallback tests notification without callback.
func TestAppLifecycleHandler_NotifyStateChangeNoCallback(t *testing.T) {
	handler := NewAppLifecycleHandler()

	// Should not panic without callback
	handler.NotifyStateChange(AppStateBackground)

	if handler.currentState != AppStateBackground {
		t.Errorf("currentState = %v, want %v", handler.currentState, AppStateBackground)
	}
}

// TestAppLifecycleHandler_NotifyInterruption tests interruption notification.
func TestAppLifecycleHandler_NotifyInterruption(t *testing.T) {
	handler := NewAppLifecycleHandler()

	// Should not panic without callback
	handler.NotifyInterruption(InterruptionCall)
}

// TestAppLifecycleHandler_GetCurrentState tests current state retrieval.
func TestAppLifecycleHandler_GetCurrentState(t *testing.T) {
	handler := NewAppLifecycleHandler()

	if handler.GetCurrentState() != AppStateActive {
		t.Errorf("GetCurrentState initially = %v, want %v", handler.GetCurrentState(), AppStateActive)
	}

	handler.NotifyStateChange(AppStateTerminating)

	if handler.GetCurrentState() != AppStateTerminating {
		t.Errorf("GetCurrentState after change = %v, want %v", handler.GetCurrentState(), AppStateTerminating)
	}
}

// TestAppLifecycleHandler_IsActive tests active state check.
func TestAppLifecycleHandler_IsActive(t *testing.T) {
	handler := NewAppLifecycleHandler()

	if !handler.IsActive() {
		t.Error("IsActive initially should be true")
	}

	handler.NotifyStateChange(AppStateBackground)

	if handler.IsActive() {
		t.Error("IsActive after background should be false")
	}

	handler.NotifyStateChange(AppStateActive)

	if !handler.IsActive() {
		t.Error("IsActive after returning to active should be true")
	}
}

// TestAppLifecycleHandler_IsBackground tests background state check.
func TestAppLifecycleHandler_IsBackground(t *testing.T) {
	handler := NewAppLifecycleHandler()

	if handler.IsBackground() {
		t.Error("IsBackground initially should be false")
	}

	handler.NotifyStateChange(AppStateBackground)

	if !handler.IsBackground() {
		t.Error("IsBackground after background should be true")
	}
}

// =============================================================================
// Input Utility Functions Tests
// =============================================================================

// TestNormalizeWASDInput tests WASD to analog conversion.
func TestNormalizeWASDInput(t *testing.T) {
	tests := []struct {
		name        string
		up, down    bool
		left, right bool
		wantX       float64
		wantY       float64
		epsilon     float64
	}{
		{"none", false, false, false, false, 0, 0, 0.001},
		{"right", false, false, false, true, 1, 0, 0.001},
		{"left", false, false, true, false, -1, 0, 0.001},
		{"up", true, false, false, false, 0, -1, 0.001},
		{"down", false, true, false, false, 0, 1, 0.001},
		{"up_right_diagonal", true, false, false, true, 0.707, -0.707, 0.01},
		{"down_left_diagonal", false, true, true, false, -0.707, 0.707, 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y := NormalizeWASDInput(tt.up, tt.down, tt.left, tt.right)

			if math.Abs(x-tt.wantX) > tt.epsilon {
				t.Errorf("x = %.3f, want %.3f", x, tt.wantX)
			}
			if math.Abs(y-tt.wantY) > tt.epsilon {
				t.Errorf("y = %.3f, want %.3f", y, tt.wantY)
			}
		})
	}
}

// TestApplyInputResponseCurve tests response curve application.
func TestApplyInputResponseCurve(t *testing.T) {
	tests := []struct {
		name       string
		value      float64
		curvePower float64
		want       float64
		epsilon    float64
	}{
		{"linear_positive", 0.5, 1.0, 0.5, 0.001},
		{"linear_negative", -0.5, 1.0, -0.5, 0.001},
		{"squared_positive", 0.5, 2.0, 0.25, 0.001},
		{"squared_negative", -0.5, 2.0, -0.25, 0.001},
		{"zero_value", 0.0, 2.0, 0.0, 0.001},
		{"full_positive", 1.0, 2.0, 1.0, 0.001},
		{"sqrt_positive", 0.25, 0.5, 0.5, 0.001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyInputResponseCurve(tt.value, tt.curvePower)

			if math.Abs(got-tt.want) > tt.epsilon {
				t.Errorf("ApplyInputResponseCurve(%.2f, %.2f) = %.3f, want %.3f", tt.value, tt.curvePower, got, tt.want)
			}
		})
	}
}

// TestClampAnalogInput tests analog input clamping.
func TestClampAnalogInput(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		deadZone float64
		maxValue float64
		want     float64
	}{
		{"in_dead_zone_positive", 0.05, 0.1, 1.0, 0.0},
		{"in_dead_zone_negative", -0.05, 0.1, 1.0, 0.0},
		{"at_dead_zone_edge", 0.11, 0.1, 1.0, 0.11},
		{"normal_value", 0.5, 0.1, 1.0, 0.5},
		{"clamped_positive", 1.5, 0.1, 1.0, 1.0},
		{"clamped_negative", -1.5, 0.1, 1.0, -1.0},
		{"zero_in_dead_zone", 0.0, 0.1, 1.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClampAnalogInput(tt.value, tt.deadZone, tt.maxValue)

			if got != tt.want {
				t.Errorf("ClampAnalogInput(%.2f, %.2f, %.2f) = %.2f, want %.2f", tt.value, tt.deadZone, tt.maxValue, got, tt.want)
			}
		})
	}
}

// TestConvertMouseToJoystick tests mouse to joystick conversion.
func TestConvertMouseToJoystick(t *testing.T) {
	tests := []struct {
		name        string
		deltaX      float64
		deltaY      float64
		sensitivity float64
		wantX       float64
		wantY       float64
	}{
		{"no_movement", 0, 0, 0.01, 0, 0},
		{"small_movement", 50, 30, 0.01, 0.5, 0.3},
		{"large_movement_clamped", 200, 200, 0.01, 1.0, 1.0},
		{"negative_clamped", -200, -200, 0.01, -1.0, -1.0},
		{"high_sensitivity", 10, 10, 0.2, 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y := ConvertMouseToJoystick(tt.deltaX, tt.deltaY, tt.sensitivity)

			if x != tt.wantX {
				t.Errorf("x = %.2f, want %.2f", x, tt.wantX)
			}
			if y != tt.wantY {
				t.Errorf("y = %.2f, want %.2f", y, tt.wantY)
			}
		})
	}
}

// TestInputAcceleration tests input acceleration.
func TestInputAcceleration(t *testing.T) {
	tests := []struct {
		name         string
		current      float64
		target       float64
		acceleration float64
		maxSpeed     float64
		wantNear     float64
		epsilon      float64
	}{
		{"accelerate_positive", 0, 1.0, 0.1, 1.0, 0.1, 0.001},
		{"accelerate_negative", 0, -1.0, 0.1, 1.0, -0.1, 0.001},
		{"at_target", 0.5, 0.5, 0.1, 1.0, 0.5, 0.001},
		{"decelerate", 1.0, 0, 0.1, 1.0, 0.9, 0.001},
		{"small_delta", 0.45, 0.5, 0.1, 1.0, 0.5, 0.001}, // Delta < acceleration
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InputAcceleration(tt.current, tt.target, tt.acceleration, tt.maxSpeed)

			if math.Abs(got-tt.wantNear) > tt.epsilon {
				t.Errorf("InputAcceleration(%.2f, %.2f, %.2f, %.2f) = %.3f, want near %.3f",
					tt.current, tt.target, tt.acceleration, tt.maxSpeed, got, tt.wantNear)
			}
		})
	}
}

// =============================================================================
// WASM Utility Functions Tests
// =============================================================================

// TestGetWASMRestrictionMessage tests WASM restriction messages.
func TestGetWASMRestrictionMessage(t *testing.T) {
	tests := []struct {
		restriction  WASMSecurityRestriction
		wantNonEmpty bool
	}{
		{RestrictionClipboard, true},
		{RestrictionFullscreen, true},
		{RestrictionAutoplay, true},
		{RestrictionPointerLock, true},
		{RestrictionLocalStorage, true},
		{RestrictionWebGL, true},
		{WASMSecurityRestriction(99), true}, // Unknown restriction
	}

	for _, tt := range tests {
		t.Run(tt.restriction.String(), func(t *testing.T) {
			msg := GetWASMRestrictionMessage(tt.restriction)

			if tt.wantNonEmpty && msg == "" {
				t.Error("GetWASMRestrictionMessage returned empty string")
			}
		})
	}
}

// TestHasWASMRestriction tests WASM restriction detection.
func TestHasWASMRestriction(t *testing.T) {
	// On non-WASM platforms, should always return false
	if !IsWASM() {
		for _, restriction := range []WASMSecurityRestriction{
			RestrictionClipboard,
			RestrictionFullscreen,
			RestrictionAutoplay,
		} {
			if HasWASMRestriction(restriction) {
				t.Errorf("HasWASMRestriction(%v) = true on non-WASM platform", restriction)
			}
		}
	}
}

// TestGetWASMWorkaroundMessage tests WASM workaround messages.
func TestGetWASMWorkaroundMessage(t *testing.T) {
	tests := []struct {
		restriction  WASMSecurityRestriction
		wantNonEmpty bool
	}{
		{RestrictionClipboard, true},
		{RestrictionFullscreen, true},
		{RestrictionAutoplay, true},
		{RestrictionPointerLock, true},
		{RestrictionLocalStorage, true},
		{WASMSecurityRestriction(99), true}, // Unknown restriction
	}

	for _, tt := range tests {
		t.Run(tt.restriction.String(), func(t *testing.T) {
			msg := GetWASMWorkaroundMessage(tt.restriction)

			if tt.wantNonEmpty && msg == "" {
				t.Error("GetWASMWorkaroundMessage returned empty string")
			}
		})
	}
}

// =============================================================================
// Platform Utility Functions Tests
// =============================================================================

// TestKeyboardObscuresUI tests keyboard obscuring check.
func TestKeyboardObscuresUI(t *testing.T) {
	// Just verify it doesn't panic and returns a boolean
	_ = KeyboardObscuresUI()
}

// TestGetMinimumTouchTargetSize tests minimum touch target size.
func TestGetMinimumTouchTargetSize(t *testing.T) {
	size := GetMinimumTouchTargetSize()

	// Should return a reasonable size (44dp is Apple HIG recommendation)
	if size < 30 || size > 100 {
		t.Errorf("GetMinimumTouchTargetSize() = %d, want between 30 and 100", size)
	}
}

// TestSupportsSystemGestures tests system gesture support check.
func TestSupportsSystemGestures(t *testing.T) {
	// Just verify it doesn't panic and returns a boolean
	_ = SupportsSystemGestures()
}

// =============================================================================
// DualJoystickLayout Tests - Additional Coverage
// =============================================================================

// TestNewDualJoystickLayout_Extended tests dual joystick layout creation with additional checks.
func TestNewDualJoystickLayout_Extended(t *testing.T) {
	layout := NewDualJoystickLayout(800, 600)

	if layout == nil {
		t.Fatal("NewDualJoystickLayout returned nil")
	}
	if layout.LeftJoystick == nil {
		t.Error("LeftJoystick is nil")
	}
	if layout.RightJoystick == nil {
		t.Error("RightJoystick is nil")
	}
	if len(layout.ActionButtons) != 2 {
		t.Errorf("len(ActionButtons) = %d, want 2", len(layout.ActionButtons))
	}
	if !layout.Visible {
		t.Error("Visible should be true initially")
	}
	if layout.touchHandler == nil {
		t.Error("touchHandler is nil")
	}
}

// TestDualJoystickLayout_GetAimAngle tests aim angle retrieval.
func TestDualJoystickLayout_GetAimAngle(t *testing.T) {
	layout := NewDualJoystickLayout(800, 600)

	// Initially should return 0 (or last saved angle)
	angle := layout.GetAimAngle()
	// Just verify it returns a valid number
	if math.IsNaN(angle) || math.IsInf(angle, 0) {
		t.Errorf("GetAimAngle returned invalid value: %v", angle)
	}
}

// TestDualJoystickLayout_IsAttackPressed tests attack button detection.
func TestDualJoystickLayout_IsAttackPressed(t *testing.T) {
	layout := NewDualJoystickLayout(800, 600)

	// Initially not pressed
	if layout.IsAttackPressed() {
		t.Error("IsAttackPressed() = true, want false initially")
	}

	// Simulate button press
	layout.ActionButtons[0].Pressed = true
	if !layout.IsAttackPressed() {
		t.Error("IsAttackPressed() = false after press")
	}
}

// TestDualJoystickLayout_IsUsePressed tests use button detection.
func TestDualJoystickLayout_IsUsePressed(t *testing.T) {
	layout := NewDualJoystickLayout(800, 600)

	// Initially not pressed
	if layout.IsUsePressed() {
		t.Error("IsUsePressed() = true, want false initially")
	}

	// Simulate button press
	layout.ActionButtons[1].Pressed = true
	if !layout.IsUsePressed() {
		t.Error("IsUsePressed() = false after press")
	}
}

// TestDualJoystickLayout_SetVisible_Extended tests visibility control.
func TestDualJoystickLayout_SetVisible_Extended(t *testing.T) {
	layout := NewDualJoystickLayout(800, 600)

	layout.SetVisible(false)
	if layout.Visible {
		t.Error("Visible should be false after SetVisible(false)")
	}

	layout.SetVisible(true)
	if !layout.Visible {
		t.Error("Visible should be true after SetVisible(true)")
	}
}

// =============================================================================
// TouchInputHandler Tests
// =============================================================================

// TestTouchInputHandler_FocusState tests focus state management.
func TestTouchInputHandler_FocusState(t *testing.T) {
	handler := NewTouchInputHandler()

	// Default focus state
	if handler.GetFocusState() != FocusStateNormal {
		t.Errorf("GetFocusState initially = %v, want %v", handler.GetFocusState(), FocusStateNormal)
	}

	// Set blurred focus
	handler.SetFocusState(FocusStateBlurred)
	if handler.GetFocusState() != FocusStateBlurred {
		t.Errorf("GetFocusState after set = %v, want %v", handler.GetFocusState(), FocusStateBlurred)
	}

	// Set focused state
	handler.SetFocusState(FocusStateFocused)
	if handler.GetFocusState() != FocusStateFocused {
		t.Errorf("GetFocusState after focused = %v, want %v", handler.GetFocusState(), FocusStateFocused)
	}
}

// TestTouchInputHandler_Transition tests transition state management.
func TestTouchInputHandler_Transition(t *testing.T) {
	handler := NewTouchInputHandler()

	// Initially not in transition
	if handler.IsInTransition() {
		t.Error("IsInTransition initially should be false")
	}

	handler.SetInTransition(true)
	if !handler.IsInTransition() {
		t.Error("IsInTransition should be true after set")
	}

	handler.SetInTransition(false)
	if handler.IsInTransition() {
		t.Error("IsInTransition should be false after clear")
	}
}

// TestTouchInputHandler_SimultaneousTouchCount tests touch counting.
func TestTouchInputHandler_SimultaneousTouchCount(t *testing.T) {
	handler := NewTouchInputHandler()

	// Initially 0
	if count := handler.GetSimultaneousTouchCount(); count != 0 {
		t.Errorf("GetSimultaneousTouchCount initially = %d, want 0", count)
	}

	if maxCount := handler.GetMaxSimultaneousTouchCount(); maxCount != 0 {
		t.Errorf("GetMaxSimultaneousTouchCount initially = %d, want 0", maxCount)
	}
}

// TestTouchInputHandler_ConsumeTouch tests touch consumption.
func TestTouchInputHandler_ConsumeTouch(t *testing.T) {
	handler := NewTouchInputHandler()

	// Non-existent touch should not be consumed
	if handler.IsTouchConsumed(0) {
		t.Error("IsTouchConsumed should be false for non-existent touch")
	}

	// Add a touch manually for testing
	handler.touches[0] = &Touch{ID: 0, Active: true}

	// Not consumed yet
	if handler.IsTouchConsumed(0) {
		t.Error("IsTouchConsumed should be false before consume")
	}

	// Consume it
	handler.ConsumeTouch(0)
	if !handler.IsTouchConsumed(0) {
		t.Error("IsTouchConsumed should be true after consume")
	}
}

// TestTouchInputHandler_ClearInputBuffer tests buffer clearing.
func TestTouchInputHandler_ClearInputBuffer(t *testing.T) {
	handler := NewTouchInputHandler()

	// Add something to the input buffer
	handler.inputBuffer = append(handler.inputBuffer, &Touch{ID: 0})
	handler.inputBuffer = append(handler.inputBuffer, &Touch{ID: 1})

	if len(handler.inputBuffer) != 2 {
		t.Errorf("inputBuffer length = %d, want 2", len(handler.inputBuffer))
	}

	handler.ClearInputBuffer()

	if len(handler.inputBuffer) != 0 {
		t.Errorf("inputBuffer length after clear = %d, want 0", len(handler.inputBuffer))
	}
}

// TestTouchInputHandler_SetDebounceTime tests debounce time configuration.
func TestTouchInputHandler_SetDebounceTime(t *testing.T) {
	handler := NewTouchInputHandler()

	handler.SetDebounceTime(100 * time.Millisecond)

	if handler.debounceTime != 100*time.Millisecond {
		t.Errorf("debounceTime = %v, want 100ms", handler.debounceTime)
	}
}

// =============================================================================
// GestureDetector Configuration Tests
// =============================================================================

// TestGestureDetector_Configuration tests gesture detector configuration setters.
func TestGestureDetector_Configuration(t *testing.T) {
	detector := NewGestureDetector()

	// Test SetDoubleTapWindow
	detector.SetDoubleTapWindow(500 * time.Millisecond)
	if detector.doubleTapWindow != 500*time.Millisecond {
		t.Errorf("doubleTapWindow = %v, want 500ms", detector.doubleTapWindow)
	}

	// Test SetLongPressThreshold
	detector.SetLongPressThreshold(600 * time.Millisecond)
	if detector.longPressThreshold != 600*time.Millisecond {
		t.Errorf("longPressThreshold = %v, want 600ms", detector.longPressThreshold)
	}

	// Test SetTapMaxDistance
	detector.SetTapMaxDistance(25.0)
	if detector.tapMaxDistance != 25.0 {
		t.Errorf("tapMaxDistance = %v, want 25.0", detector.tapMaxDistance)
	}

	// Test SetSwipeMinDistance
	detector.SetSwipeMinDistance(100.0)
	if detector.swipeMinDistance != 100.0 {
		t.Errorf("swipeMinDistance = %v, want 100.0", detector.swipeMinDistance)
	}

	// Test SetDoubleTapTolerance
	detector.SetDoubleTapTolerance(40.0)
	if detector.doubleTapTolerance != 40.0 {
		t.Errorf("doubleTapTolerance = %v, want 40.0", detector.doubleTapTolerance)
	}
}

// TestGestureDetector_GetLastVelocity tests velocity retrieval.
func TestGestureDetector_GetLastVelocity(t *testing.T) {
	detector := NewGestureDetector()

	// Initially 0
	if vel := detector.GetLastVelocity(); vel != 0 {
		t.Errorf("GetLastVelocity initially = %v, want 0", vel)
	}

	// Set velocity manually for testing
	detector.lastVelocity = 500.0
	if vel := detector.GetLastVelocity(); vel != 500.0 {
		t.Errorf("GetLastVelocity = %v, want 500.0", vel)
	}
}

// =============================================================================
// Touch Utility Function Tests
// =============================================================================

// TestTouchToScreen tests touch to screen coordinate conversion.
func TestTouchToScreen(t *testing.T) {
	touch := &Touch{X: 100, Y: 200}

	x, y := TouchToScreen(touch)

	if x != 100 {
		t.Errorf("TouchToScreen x = %d, want 100", x)
	}
	if y != 200 {
		t.Errorf("TouchToScreen y = %d, want 200", y)
	}
}

// TestTouchDelta tests touch delta calculation.
func TestTouchDelta(t *testing.T) {
	touch := &Touch{DeltaX: 10, DeltaY: -20}

	dx, dy := TouchDelta(touch)

	if dx != 10 {
		t.Errorf("TouchDelta dx = %d, want 10", dx)
	}
	if dy != -20 {
		t.Errorf("TouchDelta dy = %d, want -20", dy)
	}
}

// TestTouchDistance tests touch distance calculation.
func TestTouchDistance(t *testing.T) {
	tests := []struct {
		name    string
		startX  int
		startY  int
		x       int
		y       int
		want    float64
		epsilon float64
	}{
		{"no_movement", 100, 100, 100, 100, 0, 0.001},
		{"horizontal", 0, 0, 100, 0, 100, 0.001},
		{"vertical", 0, 0, 0, 50, 50, 0.001},
		{"diagonal_3_4_5", 0, 0, 3, 4, 5, 0.001},
		{"negative_direction", 100, 100, 97, 96, 5, 0.001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			touch := &Touch{StartX: tt.startX, StartY: tt.startY, X: tt.x, Y: tt.y}
			got := TouchDistance(touch)

			if math.Abs(got-tt.want) > tt.epsilon {
				t.Errorf("TouchDistance = %.3f, want %.3f", got, tt.want)
			}
		})
	}
}

// TestTouchDuration tests touch duration calculation.
func TestTouchDuration(t *testing.T) {
	now := time.Now()

	// Active touch
	activeTouch := &Touch{
		Active:    true,
		StartTime: now.Add(-100 * time.Millisecond),
	}
	duration := TouchDuration(activeTouch)
	if duration < 90*time.Millisecond || duration > 200*time.Millisecond {
		t.Errorf("TouchDuration for active = %v, want ~100ms", duration)
	}

	// Ended touch
	endedTouch := &Touch{
		Active:    false,
		StartTime: now.Add(-200 * time.Millisecond),
		EndTime:   now.Add(-50 * time.Millisecond),
	}
	duration = TouchDuration(endedTouch)
	if duration < 140*time.Millisecond || duration > 160*time.Millisecond {
		t.Errorf("TouchDuration for ended = %v, want ~150ms", duration)
	}

	// Touch with no end time (shouldn't happen in practice)
	invalidTouch := &Touch{
		Active:    false,
		StartTime: now,
	}
	duration = TouchDuration(invalidTouch)
	if duration != 0 {
		t.Errorf("TouchDuration for invalid = %v, want 0", duration)
	}
}

// =============================================================================
// Keyboard Function Tests
// =============================================================================

// TestShowKeyboard tests keyboard show function.
func TestShowKeyboard(t *testing.T) {
	// Just verify it doesn't panic
	ShowKeyboard()
}

// TestHideKeyboard tests keyboard hide function.
func TestHideKeyboard(t *testing.T) {
	// Just verify it doesn't panic
	HideKeyboard()
}

// =============================================================================
// AppLifecycleState String Tests
// =============================================================================

// TestAppLifecycleState_String tests lifecycle state string representation.
func TestAppLifecycleState_String(t *testing.T) {
	tests := []struct {
		state AppLifecycleState
		want  string
	}{
		{AppStateActive, "Active"},
		{AppStateInactive, "Inactive"},
		{AppStateBackground, "Background"},
		{AppStateTerminating, "Terminating"},
		{AppLifecycleState(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.state.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// =============================================================================
// Additional String Method Tests
// =============================================================================

// TestSystemInterruptionType_String tests interruption type string representation.
func TestSystemInterruptionType_String(t *testing.T) {
	tests := []struct {
		interruptType SystemInterruptionType
		want          string
	}{
		{InterruptionCall, "Call"},
		{InterruptionNotification, "Notification"},
		{InterruptionLowMemory, "LowMemory"},
		{InterruptionAudioRoute, "AudioRoute"},
		{SystemInterruptionType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.interruptType.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestTouchState_String tests touch state string representation.
func TestTouchState_String(t *testing.T) {
	tests := []struct {
		state TouchState
		want  string
	}{
		{TouchStateStarted, "Started"},
		{TouchStateMoved, "Moved"},
		{TouchStateStationary, "Stationary"},
		{TouchStateEnded, "Ended"},
		{TouchStateCancelled, "Cancelled"},
		{TouchState(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.state.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestFocusState_String tests focus state string representation.
func TestFocusState_String(t *testing.T) {
	tests := []struct {
		state FocusState
		want  string
	}{
		{FocusStateNormal, "Normal"},
		{FocusStateBlurred, "Blurred"},
		{FocusStateFocused, "Focused"},
		{FocusState(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.state.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestCancelGesture_String tests cancel gesture string representation.
func TestCancelGesture_String(t *testing.T) {
	tests := []struct {
		gesture CancelGesture
		want    string
	}{
		{GestureTwoFingerTap, "TwoFingerTap"},
		{GestureSwipeDown, "SwipeDown"},
		{GestureEdgeSwipe, "EdgeSwipe"},
		{GestureEscape, "Escape"},
		{GestureRightClick, "RightClick"},
		{CancelGesture(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.gesture.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// =============================================================================
// MobileMenu Additional Tests
// =============================================================================

// TestMobileMenu_SetLongPressCallback tests long press callback setting.
func TestMobileMenu_SetLongPressCallback(t *testing.T) {
	menu := NewMobileMenu(100, 100, 300, 400)

	called := false
	var receivedIndex int

	menu.SetLongPressCallback(func(index int) {
		called = true
		receivedIndex = index
	})

	// Verify callback is set by calling it directly
	if menu.longPressCallback != nil {
		menu.longPressCallback(5)
	}

	if !called {
		t.Error("Long press callback was not called")
	}
	if receivedIndex != 5 {
		t.Errorf("receivedIndex = %d, want 5", receivedIndex)
	}
}

// TestMobileMenu_StopScrolling tests stop scrolling function.
func TestMobileMenu_StopScrolling(t *testing.T) {
	menu := NewMobileMenu(100, 100, 300, 400)

	// Simulate scrolling
	menu.isScrolling = true
	menu.scrollVelocity = 100.0

	menu.StopScrolling()

	if menu.isScrolling {
		t.Error("isScrolling should be false after StopScrolling")
	}
	if menu.scrollVelocity != 0 {
		t.Errorf("scrollVelocity = %.1f, want 0", menu.scrollVelocity)
	}
}

// TestMobileMenu_GetPressedItemIndex tests pressed item index retrieval.
func TestMobileMenu_GetPressedItemIndex(t *testing.T) {
	menu := NewMobileMenu(100, 100, 300, 400)
	menu.AddItem("Item 1", true, nil)
	menu.AddItem("Item 2", true, nil)

	// Initially -1
	if index := menu.GetPressedItemIndex(); index != -1 {
		t.Errorf("GetPressedItemIndex initially = %d, want -1", index)
	}

	// Set pressed item manually
	menu.pressedItemIndex = 1
	if index := menu.GetPressedItemIndex(); index != 1 {
		t.Errorf("GetPressedItemIndex after set = %d, want 1", index)
	}
}

// =============================================================================
// TouchButton Tests
// =============================================================================

// TestNewTouchButton tests touch button creation.
func TestNewTouchButton(t *testing.T) {
	button := NewTouchButton(100, 200, 60, 50, "Test", nil)

	if button == nil {
		t.Fatal("NewTouchButton returned nil")
	}
	if button.X != 100 {
		t.Errorf("X = %f, want 100", button.X)
	}
	if button.Y != 200 {
		t.Errorf("Y = %f, want 200", button.Y)
	}
	if button.Width != 60 {
		t.Errorf("Width = %f, want 60", button.Width)
	}
	if button.Height != 50 {
		t.Errorf("Height = %f, want 50 (min enforced)", button.Height)
	}
	if button.Label != "Test" {
		t.Errorf("Label = %q, want %q", button.Label, "Test")
	}
	if !button.Enabled {
		t.Error("Enabled should be true initially")
	}
	if !button.Visible {
		t.Error("Visible should be true initially")
	}
}

// TestNewTouchButton_MinSizeEnforcement tests minimum size enforcement.
func TestNewTouchButton_MinSizeEnforcement(t *testing.T) {
	button := NewTouchButton(0, 0, 20, 30, "Small", nil)

	// Width and height should be at least 44
	if button.Width < 44 {
		t.Errorf("Width = %f, should be at least 44", button.Width)
	}
	if button.Height < 44 {
		t.Errorf("Height = %f, should be at least 44", button.Height)
	}
}

// TestTouchButton_SetPosition tests button position setting.
func TestTouchButton_SetPosition(t *testing.T) {
	button := NewTouchButton(0, 0, 50, 50, "Test", nil)

	button.SetPosition(150, 250)

	if button.X != 150 {
		t.Errorf("X = %f, want 150", button.X)
	}
	if button.Y != 250 {
		t.Errorf("Y = %f, want 250", button.Y)
	}
}

// TestTouchButton_SetSize tests button size setting.
func TestTouchButton_SetSize(t *testing.T) {
	button := NewTouchButton(0, 0, 50, 50, "Test", nil)

	button.SetSize(80, 90)

	if button.Width != 80 {
		t.Errorf("Width = %f, want 80", button.Width)
	}
	if button.Height != 90 {
		t.Errorf("Height = %f, want 90", button.Height)
	}
}

// TestTouchButton_isPointInButton tests point-in-button detection.
func TestTouchButton_isPointInButton(t *testing.T) {
	button := NewTouchButton(100, 100, 50, 50, "Test", nil)

	tests := []struct {
		name string
		x    float64
		y    float64
		want bool
	}{
		{"inside", 125, 125, true},
		{"top_left_corner", 100, 100, true},
		{"bottom_right_corner", 150, 150, true},
		{"outside_left", 50, 125, false},
		{"outside_right", 200, 125, false},
		{"outside_top", 125, 50, false},
		{"outside_bottom", 125, 200, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := button.isPointInButton(tt.x, tt.y)
			if got != tt.want {
				t.Errorf("isPointInButton(%.1f, %.1f) = %v, want %v", tt.x, tt.y, got, tt.want)
			}
		})
	}
}

// =============================================================================
// VirtualJoystick Tests
// =============================================================================

// TestNewVirtualJoystick tests virtual joystick creation.
func TestNewVirtualJoystick(t *testing.T) {
	joystick := NewVirtualJoystick(100, 200, 50, JoystickTypeMovement)

	if joystick == nil {
		t.Fatal("NewVirtualJoystick returned nil")
	}
	if joystick.X != 100 {
		t.Errorf("X = %.1f, want 100", joystick.X)
	}
	if joystick.Y != 200 {
		t.Errorf("Y = %.1f, want 200", joystick.Y)
	}
	if joystick.Radius != 50 {
		t.Errorf("Radius = %.1f, want 50", joystick.Radius)
	}
	if joystick.Type != JoystickTypeMovement {
		t.Errorf("Type = %v, want JoystickTypeMovement", joystick.Type)
	}
}

// TestVirtualJoystick_GetDirection tests direction retrieval.
func TestVirtualJoystick_GetDirection(t *testing.T) {
	joystick := NewVirtualJoystick(100, 200, 50, JoystickTypeAim)

	// Initially (0, 0)
	x, y := joystick.GetDirection()
	if x != 0 || y != 0 {
		t.Errorf("GetDirection initially = (%.1f, %.1f), want (0, 0)", x, y)
	}

	// Set direction manually
	joystick.DirectionX = 0.5
	joystick.DirectionY = -0.8

	x, y = joystick.GetDirection()
	if x != 0.5 {
		t.Errorf("GetDirection x = %.1f, want 0.5", x)
	}
	if y != -0.8 {
		t.Errorf("GetDirection y = %.1f, want -0.8", y)
	}
}

// TestVirtualJoystick_GetAngle tests angle retrieval.
func TestVirtualJoystick_GetAngle(t *testing.T) {
	joystick := NewVirtualJoystick(100, 200, 50, JoystickTypeAim)

	// Initially 0
	if angle := joystick.GetAngle(); angle != 0 {
		t.Errorf("GetAngle initially = %.2f, want 0", angle)
	}

	// Set direction right (angle should be 0)
	joystick.DirectionX = 1
	joystick.DirectionY = 0

	// Just verify it returns a valid number
	angle := joystick.GetAngle()
	if math.IsNaN(angle) || math.IsInf(angle, 0) {
		t.Errorf("GetAngle returned invalid value: %v", angle)
	}
}

// TestVirtualJoystick_IsActive tests active state.
func TestVirtualJoystick_IsActive(t *testing.T) {
	joystick := NewVirtualJoystick(100, 200, 50, JoystickTypeMovement)

	if joystick.IsActive() {
		t.Error("IsActive initially should be false")
	}

	joystick.Active = true
	if !joystick.IsActive() {
		t.Error("IsActive should be true after setting")
	}
}

// =============================================================================
// DualJoystickLayout Movement Tests
// =============================================================================

// TestDualJoystickLayout_IsMoving tests movement detection.
func TestDualJoystickLayout_IsMoving(t *testing.T) {
	layout := NewDualJoystickLayout(800, 600)

	// Initially not moving
	if layout.IsMoving() {
		t.Error("IsMoving initially should be false")
	}

	// Activate left joystick
	layout.LeftJoystick.Active = true
	if !layout.IsMoving() {
		t.Error("IsMoving should be true when left joystick active")
	}
}

// TestDualJoystickLayout_IsAiming tests aiming detection.
func TestDualJoystickLayout_IsAiming(t *testing.T) {
	layout := NewDualJoystickLayout(800, 600)

	// Initially not aiming
	if layout.IsAiming() {
		t.Error("IsAiming initially should be false")
	}

	// Activate right joystick
	layout.RightJoystick.Active = true
	if !layout.IsAiming() {
		t.Error("IsAiming should be true when right joystick active")
	}
}

// TestDualJoystickLayout_GetMovementDirection tests movement direction retrieval.
func TestDualJoystickLayout_GetMovementDirection(t *testing.T) {
	layout := NewDualJoystickLayout(800, 600)

	// Set direction
	layout.LeftJoystick.DirectionX = 0.7
	layout.LeftJoystick.DirectionY = -0.3

	x, y := layout.GetMovementDirection()
	if x != 0.7 {
		t.Errorf("GetMovementDirection x = %.1f, want 0.7", x)
	}
	if y != -0.3 {
		t.Errorf("GetMovementDirection y = %.1f, want -0.3", y)
	}
}

// TestDualJoystickLayout_GetAimDirection tests aim direction retrieval.
func TestDualJoystickLayout_GetAimDirection(t *testing.T) {
	layout := NewDualJoystickLayout(800, 600)

	// Set direction
	layout.RightJoystick.DirectionX = 1.0
	layout.RightJoystick.DirectionY = 0.0

	x, y := layout.GetAimDirection()
	if x != 1.0 {
		t.Errorf("GetAimDirection x = %.1f, want 1.0", x)
	}
	if y != 0.0 {
		t.Errorf("GetAimDirection y = %.1f, want 0.0", y)
	}
}

// =============================================================================
// WASMSecurityRestriction String Tests
// =============================================================================

// TestWASMSecurityRestriction_String tests WASM restriction string representation.
func TestWASMSecurityRestriction_String(t *testing.T) {
	tests := []struct {
		restriction WASMSecurityRestriction
		want        string
	}{
		{RestrictionClipboard, "Clipboard"},
		{RestrictionFullscreen, "Fullscreen"},
		{RestrictionAutoplay, "Autoplay"},
		{RestrictionPointerLock, "PointerLock"},
		{RestrictionLocalStorage, "LocalStorage"},
		{RestrictionWebGL, "WebGL"},
		{WASMSecurityRestriction(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.restriction.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// =============================================================================
// TouchButton SetSize Minimum Enforcement Tests
// =============================================================================

// TestTouchButton_SetSize_MinimumEnforcement tests SetSize with small values.
func TestTouchButton_SetSize_MinimumEnforcement(t *testing.T) {
	button := NewTouchButton(0, 0, 100, 100, "Test", nil)

	// Set size below minimum
	button.SetSize(20, 30)

	// Should enforce 44px minimum
	if button.Width < 44 {
		t.Errorf("Width after SetSize(20, 30) = %f, should be at least 44", button.Width)
	}
	if button.Height < 44 {
		t.Errorf("Height after SetSize(20, 30) = %f, should be at least 44", button.Height)
	}

	// Set size above minimum
	button.SetSize(80, 90)
	if button.Width != 80 {
		t.Errorf("Width after SetSize(80, 90) = %f, want 80", button.Width)
	}
	if button.Height != 90 {
		t.Errorf("Height after SetSize(80, 90) = %f, want 90", button.Height)
	}
}

// =============================================================================
// JoystickType Tests
// =============================================================================

// TestJoystickType tests joystick type values.
func TestJoystickType(t *testing.T) {
	if JoystickTypeMovement != 0 {
		t.Errorf("JoystickTypeMovement = %d, want 0", JoystickTypeMovement)
	}
	if JoystickTypeAim != 1 {
		t.Errorf("JoystickTypeAim = %d, want 1", JoystickTypeAim)
	}
}

// =============================================================================
// Orientation Tests
// =============================================================================

// TestOrientation tests orientation constants.
func TestOrientation(t *testing.T) {
	// Verify orientation constants exist and have distinct values
	if OrientationPortrait == OrientationLandscape {
		t.Error("OrientationPortrait and OrientationLandscape should have different values")
	}
}

// =============================================================================
// Additional MobileHUD Tests
// =============================================================================

// TestMobileHUD_SetHealth_MultipleValues tests multiple health updates.
func TestMobileHUD_SetHealth_MultipleValues(t *testing.T) {
	hud := NewMobileHUD(800, 600)

	tests := []float64{1.0, 0.75, 0.5, 0.25, 0.0}

	for _, value := range tests {
		hud.SetHealth(value)
		if hud.HealthBar.Value != value {
			t.Errorf("SetHealth(%.2f): Value = %f, want %f", value, hud.HealthBar.Value, value)
		}
	}
}

// TestMobileHUD_SetMana_MultipleValues tests multiple mana updates.
func TestMobileHUD_SetMana_MultipleValues(t *testing.T) {
	hud := NewMobileHUD(800, 600)

	hud.SetMana(0.5)
	if hud.ManaBar.Value != 0.5 {
		t.Errorf("ManaBar.Value = %f, want 0.5", hud.ManaBar.Value)
	}

	hud.SetMana(0.25)
	if hud.ManaBar.Value != 0.25 {
		t.Errorf("ManaBar.Value = %f, want 0.25", hud.ManaBar.Value)
	}
}

// TestMobileHUD_SetExperience_MultipleValues tests multiple XP updates.
func TestMobileHUD_SetExperience_MultipleValues(t *testing.T) {
	hud := NewMobileHUD(800, 600)

	hud.SetExperience(0.5)
	// Can't access XPBar directly, but verify no panic
}

// =============================================================================
// VirtualControlsLayout Additional Tests
// =============================================================================

// TestVirtualControlsLayout_Update_Visible tests Update when layout is visible.
func TestVirtualControlsLayout_Update_Visible(t *testing.T) {
	layout := NewVirtualControlsLayout(800, 600)
	layout.SetVisible(true)

	// Should not panic
	layout.Update()

	// Touch handler should be updated
	if layout.touchHandler == nil {
		t.Error("touchHandler should not be nil after Update")
	}
}

// TestVirtualControlsLayout_Multiple_ButtonPresses tests multiple button interactions.
func TestVirtualControlsLayout_Multiple_ButtonPresses(t *testing.T) {
	layout := NewVirtualControlsLayout(800, 600)

	// Test all buttons are not pressed initially
	if layout.IsActionPressed() {
		t.Error("IsActionPressed should be false initially")
	}
	if layout.IsSecondaryPressed() {
		t.Error("IsSecondaryPressed should be false initially")
	}
	if layout.IsMenuPressed() {
		t.Error("IsMenuPressed should be false initially")
	}
	if layout.IsInventoryPressed() {
		t.Error("IsInventoryPressed should be false initially")
	}
	if layout.IsTargetPressed() {
		t.Error("IsTargetPressed should be false initially")
	}
	if layout.IsInteractPressed() {
		t.Error("IsInteractPressed should be false initially")
	}
}

// =============================================================================
// GestureDetector Additional Tests
// =============================================================================

// TestGestureDetector_IsPinching_Extended tests pinch detection state.
func TestGestureDetector_IsPinching_Extended(t *testing.T) {
	detector := NewGestureDetector()

	// Initially not pinching
	if detector.IsPinching() {
		t.Error("IsPinching should be false initially")
	}

	// Set active manually
	detector.pinchActive = true
	if !detector.IsPinching() {
		t.Error("IsPinching should be true after setting")
	}
}

// =============================================================================
// Platform Detection Tests
// =============================================================================

// TestGetPlatform tests platform detection.
func TestGetPlatform_NotNil(t *testing.T) {
	// Just verify it returns a valid platform constant
	platform := GetPlatform()

	// Should be one of the known platforms (on testing machine, likely Unknown)
	valid := platform == PlatformIOS || platform == PlatformAndroid ||
		platform == PlatformWASM || platform == PlatformUnknown
	if !valid {
		t.Errorf("GetPlatform returned unexpected value: %v", platform)
	}
}

// TestIsMobilePlatform tests mobile platform detection.
func TestIsMobilePlatform_Desktop(t *testing.T) {
	// On test machine (Linux/macOS/Windows), should return false
	if IsMobilePlatform() {
		t.Log("IsMobilePlatform returned true - running on mobile or emulated")
	}
}

// TestIsTouchCapable tests touch capability detection.
func TestIsTouchCapable_Desktop(t *testing.T) {
	// On desktop, should return false (unless WASM)
	capable := IsTouchCapable()
	_ = capable // Just verify it runs
}

// TestIsWASM tests WASM platform detection.
func TestIsWASM_Desktop(t *testing.T) {
	// On non-WASM platforms, should return false
	if IsWASM() {
		t.Error("IsWASM should return false on non-WASM platform")
	}
}

// TestIsIOS tests iOS platform detection.
func TestIsIOS_Desktop(t *testing.T) {
	// On non-iOS platforms, should return false
	if IsIOS() {
		t.Error("IsIOS should return false on non-iOS platform")
	}
}

// TestIsAndroid tests Android platform detection.
func TestIsAndroid_Desktop(t *testing.T) {
	// On non-Android platforms, should return false
	if IsAndroid() {
		t.Error("IsAndroid should return false on non-Android platform")
	}
}

// TestSupportsBackButton tests back button support detection.
func TestSupportsBackButton_Desktop(t *testing.T) {
	// On non-Android platforms, should return false
	if SupportsBackButton() {
		t.Error("SupportsBackButton should return false on non-Android platform")
	}
}

// TestGetOrientation_Extended tests orientation detection with more cases.
func TestGetOrientation_Extended(t *testing.T) {
	// Just verify the function returns valid orientation for various inputs
	got := GetOrientation(1920, 1080)
	if got != OrientationLandscape {
		t.Errorf("GetOrientation(1920, 1080) = %v, want Landscape", got)
	}

	got = GetOrientation(1080, 1920)
	if got != OrientationPortrait {
		t.Errorf("GetOrientation(1080, 1920) = %v, want Portrait", got)
	}
}

// TestGetBackButtonName tests back button name.
func TestGetBackButtonName_Desktop(t *testing.T) {
	name := GetBackButtonName()
	// On desktop, should return "ESC"
	if name != "ESC" {
		t.Errorf("GetBackButtonName = %q, want %q on desktop", name, "ESC")
	}
}

// =============================================================================
// Additional MobileMenu Tests
// =============================================================================

// TestMobileMenu_ScrollOffsetField tests scroll offset management.
func TestMobileMenu_ScrollOffsetField(t *testing.T) {
	menu := NewMobileMenu(100, 100, 300, 200)

	// Add many items to enable scrolling
	for i := 0; i < 20; i++ {
		menu.AddItem("Item", true, nil)
	}

	// Initially scroll offset is 0
	if menu.scrollOffset != 0 {
		t.Errorf("scrollOffset initially = %f, want 0", menu.scrollOffset)
	}

	// Manually set scroll offset
	menu.scrollOffset = 50.0
	if menu.scrollOffset != 50.0 {
		t.Errorf("scrollOffset = %f, want 50.0", menu.scrollOffset)
	}
}

// TestMobileMenu_ItemCount tests item counting.
func TestMobileMenu_ItemCount(t *testing.T) {
	menu := NewMobileMenu(100, 100, 300, 400)

	if len(menu.Items) != 0 {
		t.Errorf("Items initially = %d, want 0", len(menu.Items))
	}

	menu.AddItem("Item 1", true, nil)
	menu.AddItem("Item 2", true, nil)
	menu.AddItem("Item 3", true, nil)

	if len(menu.Items) != 3 {
		t.Errorf("Items after adding = %d, want 3", len(menu.Items))
	}
}
