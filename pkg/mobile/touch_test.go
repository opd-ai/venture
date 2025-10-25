package mobile

import (
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestTouch tests the Touch data structure.
func TestTouch(t *testing.T) {
	now := time.Now()
	touch := &Touch{
		ID:        ebiten.TouchID(1),
		X:         100,
		Y:         200,
		StartX:    50,
		StartY:    150,
		StartTime: now,
		Active:    true,
	}

	if touch.ID != 1 {
		t.Errorf("Touch.ID = %d, want 1", touch.ID)
	}
	if touch.X != 100 {
		t.Errorf("Touch.X = %d, want 100", touch.X)
	}
	if touch.Y != 200 {
		t.Errorf("Touch.Y = %d, want 200", touch.Y)
	}
	if touch.StartX != 50 {
		t.Errorf("Touch.StartX = %d, want 50", touch.StartX)
	}
	if touch.StartY != 150 {
		t.Errorf("Touch.StartY = %d, want 150", touch.StartY)
	}
	if !touch.Active {
		t.Error("Touch.Active = false, want true")
	}
	if !touch.StartTime.Equal(now) {
		t.Errorf("Touch.StartTime mismatch")
	}
}

// TestTouch_Inactive tests inactive touch state.
func TestTouch_Inactive(t *testing.T) {
	touch := &Touch{
		ID:     0,
		Active: false,
	}

	if touch.Active {
		t.Error("Touch.Active = true, want false")
	}
}

// TestNewTouchInputHandler tests TouchInputHandler creation.
func TestNewTouchInputHandler(t *testing.T) {
	handler := NewTouchInputHandler()

	if handler == nil {
		t.Fatal("NewTouchInputHandler returned nil")
	}
	if handler.touches == nil {
		t.Error("handler.touches map not initialized")
	}
	if handler.gestureDetector == nil {
		t.Error("handler.gestureDetector not initialized")
	}
	if handler.tapCount != 0 {
		t.Errorf("handler.tapCount = %d, want 0", handler.tapCount)
	}
}

// TestTouchInputHandler_GetTouchCount tests touch counting logic.
func TestTouchInputHandler_GetTouchCount(t *testing.T) {
	handler := NewTouchInputHandler()

	// Initially should be 0
	if count := handler.GetTouchCount(); count != 0 {
		t.Errorf("GetTouchCount() = %d, want 0", count)
	}

	// Add active touches manually for testing
	handler.touches[0] = &Touch{ID: 0, Active: true}
	handler.touches[1] = &Touch{ID: 1, Active: true}
	handler.touches[2] = &Touch{ID: 2, Active: false}

	// Should count only active touches
	if count := handler.GetTouchCount(); count != 2 {
		t.Errorf("GetTouchCount() = %d, want 2", count)
	}
}

// TestTouchInputHandler_GetActiveTouches tests active touch retrieval.
func TestTouchInputHandler_GetActiveTouches(t *testing.T) {
	handler := NewTouchInputHandler()

	// Add mix of active and inactive touches
	handler.touches[0] = &Touch{ID: 0, X: 10, Y: 20, Active: true}
	handler.touches[1] = &Touch{ID: 1, X: 30, Y: 40, Active: false}
	handler.touches[2] = &Touch{ID: 2, X: 50, Y: 60, Active: true}

	activeTouches := handler.GetActiveTouches()

	if len(activeTouches) != 2 {
		t.Errorf("GetActiveTouches() returned %d touches, want 2", len(activeTouches))
	}

	// Verify only active touches are returned
	for _, touch := range activeTouches {
		if !touch.Active {
			t.Errorf("GetActiveTouches() included inactive touch ID %d", touch.ID)
		}
	}
}

// TestTouchInputHandler_EmptyTouches tests with no touches.
func TestTouchInputHandler_EmptyTouches(t *testing.T) {
	handler := NewTouchInputHandler()

	if handler.GetTouchCount() != 0 {
		t.Error("GetTouchCount() should be 0 for new handler")
	}

	touches := handler.GetActiveTouches()
	if len(touches) != 0 {
		t.Errorf("GetActiveTouches() should return empty slice, got %d touches", len(touches))
	}
}

// TestTouchInputHandler_IsTapping tests tap detection via public API.
func TestTouchInputHandler_IsTapping(t *testing.T) {
	handler := NewTouchInputHandler()

	// Initially no tapping (detector hasn't processed any touches)
	if handler.IsTapping() {
		t.Error("IsTapping() = true, want false initially")
	}
}

// TestTouchInputHandler_GetTapPosition tests tap position retrieval.
func TestTouchInputHandler_GetTapPosition(t *testing.T) {
	handler := NewTouchInputHandler()

	// Get initial position (should be 0, 0 before any taps)
	x, y := handler.GetTapPosition()

	// Just verify it doesn't panic and returns integers
	_ = x
	_ = y
}

// TestTouchInputHandler_IsDoubleTap tests double-tap detection.
func TestTouchInputHandler_IsDoubleTap(t *testing.T) {
	handler := NewTouchInputHandler()

	// Initially no double-tap
	if handler.IsDoubleTap() {
		t.Error("IsDoubleTap() = true, want false initially")
	}
}

// TestTouchInputHandler_IsLongPress tests long-press detection.
func TestTouchInputHandler_IsLongPress(t *testing.T) {
	handler := NewTouchInputHandler()

	// Initially no long-press
	if handler.IsLongPress() {
		t.Error("IsLongPress() = true, want false initially")
	}
}

// TestTouchInputHandler_GetSwipe tests swipe gesture retrieval.
func TestTouchInputHandler_GetSwipe(t *testing.T) {
	handler := NewTouchInputHandler()

	direction, distance, hasSwipe := handler.GetSwipe()

	// Initially no swipe
	if hasSwipe {
		t.Error("GetSwipe() hasSwipe = true, want false initially")
	}

	// Values should be zero initially
	if direction != 0.0 {
		t.Errorf("GetSwipe() direction = %.1f, want 0.0 initially", direction)
	}
	if distance != 0.0 {
		t.Errorf("GetSwipe() distance = %.1f, want 0.0 initially", distance)
	}
}

// TestTouchInputHandler_GetPinch tests pinch gesture retrieval.
func TestTouchInputHandler_GetPinch(t *testing.T) {
	handler := NewTouchInputHandler()

	scale := handler.GetPinch()

	// Initial scale should be 1.0 (no pinch)
	if scale != 1.0 {
		t.Errorf("GetPinch() scale = %.1f, want 1.0 initially", scale)
	}
}

// TestNewGestureDetector tests GestureDetector creation.
func TestNewGestureDetector(t *testing.T) {
	detector := NewGestureDetector()

	if detector == nil {
		t.Fatal("NewGestureDetector returned nil")
	}

	// Test that methods don't panic
	detector.IsTap()
	detector.IsDoubleTap()
	detector.IsLongPress()
	detector.GetTapPosition()
	detector.GetLongPressPosition()
	detector.GetSwipe()
	detector.GetPinchScale()
	detector.IsPinching()
}

// TestGestureDetector_IsTap tests tap state.
func TestGestureDetector_IsTap(t *testing.T) {
	detector := NewGestureDetector()

	// Initially no tap
	if detector.IsTap() {
		t.Error("IsTap() = true, want false initially")
	}
}

// TestGestureDetector_GetTapPosition tests tap position retrieval.
func TestGestureDetector_GetTapPosition(t *testing.T) {
	detector := NewGestureDetector()

	x, y := detector.GetTapPosition()

	// Just verify it returns without panic
	_ = x
	_ = y
}

// TestGestureDetector_IsDoubleTap tests double-tap state.
func TestGestureDetector_IsDoubleTap(t *testing.T) {
	detector := NewGestureDetector()

	// Initially no double-tap
	if detector.IsDoubleTap() {
		t.Error("IsDoubleTap() = true, want false initially")
	}
}

// TestGestureDetector_IsLongPress tests long-press state.
func TestGestureDetector_IsLongPress(t *testing.T) {
	detector := NewGestureDetector()

	// Initially no long-press
	if detector.IsLongPress() {
		t.Error("IsLongPress() = true, want false initially")
	}
}

// TestGestureDetector_GetLongPressPosition tests long-press position.
func TestGestureDetector_GetLongPressPosition(t *testing.T) {
	detector := NewGestureDetector()

	x, y := detector.GetLongPressPosition()

	// Just verify it returns without panic
	_ = x
	_ = y
}

// TestGestureDetector_GetSwipe tests swipe gesture data.
func TestGestureDetector_GetSwipe(t *testing.T) {
	detector := NewGestureDetector()

	direction, distance, hasSwipe := detector.GetSwipe()

	// Initially no swipe
	if hasSwipe {
		t.Error("GetSwipe() hasSwipe = true, want false initially")
	}
	if direction != 0.0 {
		t.Errorf("GetSwipe() direction = %.1f, want 0.0", direction)
	}
	if distance != 0.0 {
		t.Errorf("GetSwipe() distance = %.1f, want 0.0", distance)
	}
}

// TestGestureDetector_GetPinchScale tests pinch scale retrieval.
func TestGestureDetector_GetPinchScale(t *testing.T) {
	detector := NewGestureDetector()

	scale := detector.GetPinchScale()

	// Initial scale should be 1.0
	if scale != 1.0 {
		t.Errorf("GetPinchScale() = %.1f, want 1.0", scale)
	}
}

// TestGestureDetector_IsPinching tests pinch state.
func TestGestureDetector_IsPinching(t *testing.T) {
	detector := NewGestureDetector()

	// Initially not pinching
	if detector.IsPinching() {
		t.Error("IsPinching() = true, want false initially")
	}
}

// TestGestureDetector_Update tests gesture update logic with mock touches.
func TestGestureDetector_Update(t *testing.T) {
	detector := NewGestureDetector()
	touches := make(map[ebiten.TouchID]*Touch)

	// Test with no touches
	detector.Update(touches)

	if detector.IsTap() {
		t.Error("IsTap() = true after update with no touches")
	}

	// Test with single active touch
	now := time.Now()
	touches[0] = &Touch{
		ID:        0,
		X:         100,
		Y:         200,
		StartX:    100,
		StartY:    200,
		StartTime: now,
		Active:    true,
	}

	detector.Update(touches)

	// Detector should have processed the touch without panicking
	// Note: Actual gesture detection requires specific touch patterns
}

// TestGestureDetector_UpdateMultipleTimes tests multiple updates.
func TestGestureDetector_UpdateMultipleTimes(t *testing.T) {
	detector := NewGestureDetector()
	touches := make(map[ebiten.TouchID]*Touch)

	// Multiple updates should not panic
	for i := 0; i < 10; i++ {
		detector.Update(touches)
	}

	// Add and remove touches
	touches[0] = &Touch{ID: 0, X: 50, Y: 50, StartTime: time.Now(), Active: true}
	detector.Update(touches)

	delete(touches, 0)
	detector.Update(touches)

	// Should not panic
}

// TestGestureDetector_UpdateWithMultipleTouches tests multi-touch handling.
func TestGestureDetector_UpdateWithMultipleTouches(t *testing.T) {
	detector := NewGestureDetector()
	touches := make(map[ebiten.TouchID]*Touch)

	now := time.Now()

	// Add two touches (for pinch gesture)
	touches[0] = &Touch{
		ID:        0,
		X:         100,
		Y:         100,
		StartX:    100,
		StartY:    100,
		StartTime: now,
		Active:    true,
	}
	touches[1] = &Touch{
		ID:        1,
		X:         200,
		Y:         200,
		StartX:    200,
		StartY:    200,
		StartTime: now,
		Active:    true,
	}

	detector.Update(touches)

	// Should handle two-finger gestures without panicking
	_ = detector.IsPinching()
	_ = detector.GetPinchScale()
}

// TestGestureDetector_UpdateWithInactiveTouches tests inactive touch filtering.
func TestGestureDetector_UpdateWithInactiveTouches(t *testing.T) {
	detector := NewGestureDetector()
	touches := make(map[ebiten.TouchID]*Touch)

	// Add mix of active and inactive touches
	touches[0] = &Touch{ID: 0, X: 100, Y: 100, StartTime: time.Now(), Active: true}
	touches[1] = &Touch{ID: 1, X: 200, Y: 200, StartTime: time.Now(), Active: false}

	detector.Update(touches)

	// Should only process active touches
}
