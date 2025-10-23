//go:build !test
// +build !test

package mobile

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// Touch represents a single touch point.
type Touch struct {
	ID        ebiten.TouchID
	X, Y      int
	StartX    int
	StartY    int
	StartTime time.Time
	Active    bool
}

// TouchInputHandler manages touch input detection and gesture recognition.
type TouchInputHandler struct {
	touches         map[ebiten.TouchID]*Touch
	lastTapTime     time.Time
	tapCount        int
	gestureDetector *GestureDetector
}

// NewTouchInputHandler creates a new touch input handler.
func NewTouchInputHandler() *TouchInputHandler {
	return &TouchInputHandler{
		touches:         make(map[ebiten.TouchID]*Touch),
		gestureDetector: NewGestureDetector(),
	}
}

// Update processes touch input from Ebiten and updates gesture detection.
// Must be called every frame.
func (h *TouchInputHandler) Update() {
	// Get all active touch IDs
	activeTouchIDs := ebiten.TouchIDs()
	activeSet := make(map[ebiten.TouchID]bool)

	// Update existing touches and add new ones
	for _, id := range activeTouchIDs {
		x, y := ebiten.TouchPosition(id)
		activeSet[id] = true

		if touch, exists := h.touches[id]; exists {
			// Update existing touch
			touch.X = x
			touch.Y = y
		} else {
			// New touch started
			touch := &Touch{
				ID:        id,
				X:         x,
				Y:         y,
				StartX:    x,
				StartY:    y,
				StartTime: time.Now(),
				Active:    true,
			}
			h.touches[id] = touch
		}
	}

	// Remove touches that are no longer active
	for id, touch := range h.touches {
		if !activeSet[id] {
			touch.Active = false
			// Keep touch for one frame for tap detection
			delete(h.touches, id)
		}
	}

	// Update gesture detector with current touches
	h.gestureDetector.Update(h.touches)
}

// GetActiveTouches returns all currently active touches.
func (h *TouchInputHandler) GetActiveTouches() []*Touch {
	touches := make([]*Touch, 0, len(h.touches))
	for _, touch := range h.touches {
		if touch.Active {
			touches = append(touches, touch)
		}
	}
	return touches
}

// GetTouchCount returns the number of active touches.
func (h *TouchInputHandler) GetTouchCount() int {
	count := 0
	for _, touch := range h.touches {
		if touch.Active {
			count++
		}
	}
	return count
}

// IsTapping returns true if a tap gesture was detected this frame.
func (h *TouchInputHandler) IsTapping() bool {
	return h.gestureDetector.IsTap()
}

// GetTapPosition returns the position of the last tap.
func (h *TouchInputHandler) GetTapPosition() (int, int) {
	return h.gestureDetector.GetTapPosition()
}

// IsDoubleTap returns true if a double tap was detected this frame.
func (h *TouchInputHandler) IsDoubleTap() bool {
	return h.gestureDetector.IsDoubleTap()
}

// IsLongPress returns true if a long press is active.
func (h *TouchInputHandler) IsLongPress() bool {
	return h.gestureDetector.IsLongPress()
}

// GetSwipe returns the swipe direction and distance if a swipe was detected.
// Returns (0, 0, 0) if no swipe detected.
func (h *TouchInputHandler) GetSwipe() (direction, distance float64, detected bool) {
	return h.gestureDetector.GetSwipe()
}

// GetPinch returns the pinch scale factor if a pinch gesture is active.
// Returns 1.0 if no pinch detected.
func (h *TouchInputHandler) GetPinch() float64 {
	return h.gestureDetector.GetPinchScale()
}

// GestureDetector recognizes common mobile gestures.
type GestureDetector struct {
	// Tap detection
	lastTapTime      time.Time
	lastTapX         int
	lastTapY         int
	tapCount         int
	currentTap       bool
	currentDoubleTap bool

	// Long press detection
	longPressActive bool
	longPressX      int
	longPressY      int

	// Swipe detection
	swipeDetected  bool
	swipeDirection float64 // Radians
	swipeDistance  float64

	// Pinch detection
	pinchActive     bool
	pinchScale      float64
	initialDistance float64

	// Configuration
	tapMaxDistance     float64       // Max movement for tap
	doubleTapWindow    time.Duration // Time window for double tap
	longPressThreshold time.Duration // Time for long press
	swipeMinDistance   float64       // Min distance for swipe
}

// NewGestureDetector creates a new gesture detector with default thresholds.
func NewGestureDetector() *GestureDetector {
	return &GestureDetector{
		tapMaxDistance:     20.0,
		doubleTapWindow:    300 * time.Millisecond,
		longPressThreshold: 500 * time.Millisecond,
		swipeMinDistance:   50.0,
		pinchScale:         1.0,
	}
}

// Update processes touches and detects gestures.
func (g *GestureDetector) Update(touches map[ebiten.TouchID]*Touch) {
	// Reset frame-specific states
	g.currentTap = false
	g.currentDoubleTap = false
	g.swipeDetected = false

	activeTouches := make([]*Touch, 0, len(touches))
	for _, touch := range touches {
		if touch.Active {
			activeTouches = append(activeTouches, touch)
		}
	}

	touchCount := len(activeTouches)

	if touchCount == 0 {
		g.longPressActive = false
		g.pinchActive = false
		g.pinchScale = 1.0
		return
	}

	if touchCount == 1 {
		// Single touch gestures
		touch := activeTouches[0]
		g.detectSingleTouchGestures(touch)
	} else if touchCount == 2 {
		// Two-finger gestures (pinch/zoom)
		g.detectPinchGesture(activeTouches[0], activeTouches[1])
	}
}

// detectSingleTouchGestures detects tap, double tap, long press, and swipe.
func (g *GestureDetector) detectSingleTouchGestures(touch *Touch) {
	dx := float64(touch.X - touch.StartX)
	dy := float64(touch.Y - touch.StartY)
	distance := math.Sqrt(dx*dx + dy*dy)
	duration := time.Since(touch.StartTime)

	// Tap detection (touch just ended with minimal movement)
	if !touch.Active && distance <= g.tapMaxDistance {
		g.currentTap = true
		g.lastTapX = touch.X
		g.lastTapY = touch.Y

		// Double tap detection
		if time.Since(g.lastTapTime) <= g.doubleTapWindow {
			g.currentDoubleTap = true
			g.tapCount = 0
		} else {
			g.tapCount = 1
		}
		g.lastTapTime = time.Now()
	}

	// Long press detection
	if touch.Active && duration >= g.longPressThreshold && distance <= g.tapMaxDistance {
		g.longPressActive = true
		g.longPressX = touch.X
		g.longPressY = touch.Y
	}

	// Swipe detection (fast movement then release)
	if !touch.Active && distance >= g.swipeMinDistance {
		g.swipeDetected = true
		g.swipeDistance = distance
		g.swipeDirection = math.Atan2(dy, dx)
	}
}

// detectPinchGesture detects pinch/zoom with two fingers.
func (g *GestureDetector) detectPinchGesture(touch1, touch2 *Touch) {
	// Calculate distance between two touches
	dx := float64(touch2.X - touch1.X)
	dy := float64(touch2.Y - touch1.Y)
	currentDistance := math.Sqrt(dx*dx + dy*dy)

	if !g.pinchActive {
		// Initialize pinch
		g.pinchActive = true
		g.initialDistance = currentDistance
		g.pinchScale = 1.0
	} else {
		// Update pinch scale
		if g.initialDistance > 0 {
			g.pinchScale = currentDistance / g.initialDistance
		}
	}
}

// IsTap returns true if a tap was detected this frame.
func (g *GestureDetector) IsTap() bool {
	return g.currentTap
}

// GetTapPosition returns the position of the last tap.
func (g *GestureDetector) GetTapPosition() (int, int) {
	return g.lastTapX, g.lastTapY
}

// IsDoubleTap returns true if a double tap was detected this frame.
func (g *GestureDetector) IsDoubleTap() bool {
	return g.currentDoubleTap
}

// IsLongPress returns true if a long press is currently active.
func (g *GestureDetector) IsLongPress() bool {
	return g.longPressActive
}

// GetLongPressPosition returns the position of the long press.
func (g *GestureDetector) GetLongPressPosition() (int, int) {
	return g.longPressX, g.longPressY
}

// GetSwipe returns swipe information if detected this frame.
func (g *GestureDetector) GetSwipe() (direction, distance float64, detected bool) {
	return g.swipeDirection, g.swipeDistance, g.swipeDetected
}

// GetPinchScale returns the current pinch zoom scale factor.
// 1.0 = no zoom, >1.0 = zoom in, <1.0 = zoom out.
func (g *GestureDetector) GetPinchScale() float64 {
	return g.pinchScale
}

// IsPinching returns true if a pinch gesture is active.
func (g *GestureDetector) IsPinching() bool {
	return g.pinchActive
}
