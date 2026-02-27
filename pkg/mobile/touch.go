package mobile

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// TouchState represents the lifecycle state of a touch event.
// Platform parity fix: Explicit state tracking for consistent event ordering across platforms
// TouchState and FocusState types moved to types.go

// Touch represents a single touch point.
// Platform parity fix: Enhanced touch state tracking for event ordering consistency
// across mobile (iOS/Android) and web (WASM) platforms - touchstart/touchmove/touchend/touchcancel
type Touch struct {
	ID        ebiten.TouchID
	X, Y      int
	StartX    int
	StartY    int
	StartTime time.Time
	Active    bool

	// Platform parity fix: Touch state lifecycle tracking for consistent event ordering
	// Addresses differences between native touch events and WASM touch event timing
	State    TouchState // Current lifecycle state (started/moved/ended/cancelled)
	LastX    int        // Previous X position for delta calculation
	LastY    int        // Previous Y position for delta calculation
	DeltaX   int        // Frame-to-frame X movement delta
	DeltaY   int        // Frame-to-frame Y movement delta
	Consumed bool       // Platform parity fix: Prevent duplicate processing across UI systems
	EndTime  time.Time  // Platform parity fix: Track end time for gesture timing analysis
}

// TouchInputHandler manages touch input detection and gesture recognition.
// Platform parity fix: Added input buffering and state management for consistent behavior
// across desktop (mouse simulation), mobile (native touch), and web (WASM touch events)
type TouchInputHandler struct {
	touches         map[ebiten.TouchID]*Touch
	lastTapTime     time.Time
	tapCount        int
	gestureDetector *GestureDetector

	// Platform parity fix: Input buffering for consistent event processing across platforms
	// Addresses timing differences between native events and Ebiten's polling model
	inputBuffer     []*Touch      // Buffered touch events for debouncing
	bufferMaxSize   int           // Maximum buffered events (prevents memory issues on lag)
	debounceTime    time.Duration // Minimum time between processed touch changes
	lastProcessTime time.Time     // Last input processing time for debouncing

	// Platform parity fix: Multi-touch simultaneity tracking for gesture parity
	// Enables shift+click → two-finger tap equivalents
	simultaneousTouches int        // Current count of simultaneous active touches
	maxSimultaneous     int        // Maximum simultaneous touches seen this gesture
	focusState          FocusState // Platform parity fix: Track focus/blur for input filtering
	inTransition        bool       // Platform parity fix: Block input during UI transitions
}

// NewTouchInputHandler creates a new touch input handler.
// Platform parity fix: Configurable debouncing and buffering for consistent cross-platform behavior
func NewTouchInputHandler() *TouchInputHandler {
	return &TouchInputHandler{
		touches:         make(map[ebiten.TouchID]*Touch),
		gestureDetector: NewGestureDetector(),
		// Platform parity fix: Default input buffering configuration
		// 16ms debounce (~60 FPS frame time) prevents duplicate processing
		inputBuffer:   make([]*Touch, 0, 32),
		bufferMaxSize: 32,
		debounceTime:  16 * time.Millisecond,
		// INTENTIONAL time.Now() usage: Input debouncing timestamp (non-procgen operational timing).
		// This is NOT part of procedural content generation and does not affect determinism.
		lastProcessTime: time.Now(),
		focusState:      FocusStateNormal,
		inTransition:    false,
	}
}

// Update processes touch input from Ebiten and updates gesture detection.
// Must be called every frame.
// Platform parity fix: Enhanced with state tracking, debouncing, and focus awareness
func (h *TouchInputHandler) Update() {
	if h.shouldSkipProcessing() {
		return
	}

	// INTENTIONAL time.Now() usage: Input debouncing timestamp (non-procgen operational timing).
	// This is NOT part of procedural content generation and does not affect determinism.
	now := time.Now()
	if !h.shouldProcessNow(now) {
		return
	}
	h.lastProcessTime = now

	activeTouchIDs := ebiten.TouchIDs()
	activeSet := make(map[ebiten.TouchID]bool)

	h.updateSimultaneityTracking(activeTouchIDs)
	h.processActiveTouches(activeTouchIDs, activeSet, now)
	h.processEndedTouches(activeSet, now)
	h.resetSimultaneityIfNeeded()
	h.gestureDetector.Update(h.touches)
}

// shouldSkipProcessing checks if input should be skipped due to transitions or blur.
func (h *TouchInputHandler) shouldSkipProcessing() bool {
	return h.inTransition || h.focusState == FocusStateBlurred
}

// shouldProcessNow checks if enough time has passed for debouncing.
func (h *TouchInputHandler) shouldProcessNow(now time.Time) bool {
	return now.Sub(h.lastProcessTime) >= h.debounceTime
}

// updateSimultaneityTracking tracks simultaneous touches for multi-input gestures.
func (h *TouchInputHandler) updateSimultaneityTracking(activeTouchIDs []ebiten.TouchID) {
	h.simultaneousTouches = len(activeTouchIDs)
	if h.simultaneousTouches > h.maxSimultaneous {
		h.maxSimultaneous = h.simultaneousTouches
	}
}

// processActiveTouches updates existing touches and creates new ones.
func (h *TouchInputHandler) processActiveTouches(activeTouchIDs []ebiten.TouchID, activeSet map[ebiten.TouchID]bool, now time.Time) {
	for _, id := range activeTouchIDs {
		x, y := ebiten.TouchPosition(id)
		activeSet[id] = true

		if touch, exists := h.touches[id]; exists {
			h.updateExistingTouch(touch, x, y)
		} else {
			h.createNewTouch(id, x, y, now)
		}
	}
}

// updateExistingTouch updates an existing touch with delta calculation and state tracking.
func (h *TouchInputHandler) updateExistingTouch(touch *Touch, x, y int) {
	touch.LastX = touch.X
	touch.LastY = touch.Y
	touch.X = x
	touch.Y = y
	touch.DeltaX = touch.X - touch.LastX
	touch.DeltaY = touch.Y - touch.LastY

	if touch.DeltaX != 0 || touch.DeltaY != 0 {
		touch.State = TouchStateMoved
	} else {
		touch.State = TouchStateStationary
	}
}

// createNewTouch creates a new touch with full state tracking.
func (h *TouchInputHandler) createNewTouch(id ebiten.TouchID, x, y int, now time.Time) {
	touch := &Touch{
		ID:        id,
		X:         x,
		Y:         y,
		StartX:    x,
		StartY:    y,
		LastX:     x,
		LastY:     y,
		DeltaX:    0,
		DeltaY:    0,
		StartTime: now,
		Active:    true,
		State:     TouchStateStarted,
		Consumed:  false,
	}
	h.touches[id] = touch
	h.bufferTouch(touch)
}

// processEndedTouches handles touch end/cancel with state tracking.
func (h *TouchInputHandler) processEndedTouches(activeSet map[ebiten.TouchID]bool, now time.Time) {
	for id, touch := range h.touches {
		if !activeSet[id] {
			h.endTouch(touch, now)
			delete(h.touches, id)
		}
	}
}

// endTouch marks a touch as ended or cancelled based on duration.
func (h *TouchInputHandler) endTouch(touch *Touch, now time.Time) {
	touch.Active = false
	touch.EndTime = now

	duration := now.Sub(touch.StartTime)
	if duration < 50*time.Millisecond {
		touch.State = TouchStateCancelled
	} else {
		touch.State = TouchStateEnded
	}
}

// resetSimultaneityIfNeeded resets simultaneity counter when all touches end.
func (h *TouchInputHandler) resetSimultaneityIfNeeded() {
	if h.simultaneousTouches == 0 {
		h.maxSimultaneous = 0
	}
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

// Platform parity fix: Additional methods for enhanced input state management

// bufferTouch adds a touch to the input buffer for debouncing.
// Platform parity fix: Prevents duplicate processing of rapid touch events
func (h *TouchInputHandler) bufferTouch(touch *Touch) {
	if len(h.inputBuffer) >= h.bufferMaxSize {
		// Remove oldest buffered touch
		h.inputBuffer = h.inputBuffer[1:]
	}
	// Create a copy to avoid reference issues
	bufferedTouch := &Touch{
		ID:        touch.ID,
		X:         touch.X,
		Y:         touch.Y,
		StartX:    touch.StartX,
		StartY:    touch.StartY,
		StartTime: touch.StartTime,
		Active:    touch.Active,
		State:     touch.State,
	}
	h.inputBuffer = append(h.inputBuffer, bufferedTouch)
}

// SetFocusState updates the focus state for input filtering.
// Platform parity fix: Allows UI systems to block input during focus changes
// Mobile/Web: Called when keyboard appears, tab backgrounds, etc.
func (h *TouchInputHandler) SetFocusState(state FocusState) {
	h.focusState = state
}

// GetFocusState returns the current focus state.
func (h *TouchInputHandler) GetFocusState() FocusState {
	return h.focusState
}

// SetInTransition sets whether UI is transitioning.
// Platform parity fix: Prevents input during UI state changes (menu opening/closing)
func (h *TouchInputHandler) SetInTransition(inTransition bool) {
	h.inTransition = inTransition
}

// IsInTransition returns whether UI is currently transitioning.
func (h *TouchInputHandler) IsInTransition() bool {
	return h.inTransition
}

// GetSimultaneousTouchCount returns the current number of simultaneous touches.
// Platform parity fix: Enables multi-input simultaneity detection (shift+click → two-finger tap)
func (h *TouchInputHandler) GetSimultaneousTouchCount() int {
	return h.simultaneousTouches
}

// GetMaxSimultaneousTouchCount returns the maximum simultaneous touches in current gesture.
// Platform parity fix: Useful for detecting complex multi-finger gestures
func (h *TouchInputHandler) GetMaxSimultaneousTouchCount() int {
	return h.maxSimultaneous
}

// ConsumeTouch marks a touch as consumed to prevent duplicate processing.
// Platform parity fix: Prevents multiple UI systems from processing the same touch
func (h *TouchInputHandler) ConsumeTouch(id ebiten.TouchID) {
	if touch, exists := h.touches[id]; exists {
		touch.Consumed = true
	}
}

// IsTouchConsumed returns whether a touch has been consumed.
func (h *TouchInputHandler) IsTouchConsumed(id ebiten.TouchID) bool {
	if touch, exists := h.touches[id]; exists {
		return touch.Consumed
	}
	return false
}

// ClearInputBuffer clears the input buffer.
// Platform parity fix: Called during major state transitions (scene changes, etc.)
func (h *TouchInputHandler) ClearInputBuffer() {
	h.inputBuffer = h.inputBuffer[:0]
}

// SetDebounceTime configures the debounce delay.
// Platform parity fix: Allows tuning for different platform performance characteristics
func (h *TouchInputHandler) SetDebounceTime(duration time.Duration) {
	h.debounceTime = duration
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

	// Platform parity fix: Enhanced gesture configuration for cross-platform consistency
	// Addresses timing/threshold differences between mouse double-click and touch double-tap
	doubleTapTolerance float64 // Max distance between taps for double-tap (pixels)
	dragThreshold      float64 // Min movement to distinguish drag from tap (pixels)
	velocityThreshold  float64 // Min velocity for fling/swipe detection (pixels/ms)
	lastVelocity       float64 // Last calculated velocity for fling detection
}

// NewGestureDetector creates a new gesture detector with default thresholds.
// Platform parity fix: Normalized timing thresholds across desktop/mobile/web
func NewGestureDetector() *GestureDetector {
	return &GestureDetector{
		// Platform parity fix: Tap distance normalized for both mouse and touch
		// 20px works well for desktop mouse precision and mobile fat-finger touches
		tapMaxDistance: 20.0,

		// Platform parity fix: Double-tap/click timing normalized
		// 300ms matches typical double-click timing on desktop
		doubleTapWindow: 300 * time.Millisecond,

		// Platform parity fix: Long press threshold
		// 500ms is iOS standard, also works well for right-click alternative
		longPressThreshold: 500 * time.Millisecond,

		// Platform parity fix: Swipe detection threshold
		// 50px minimum prevents accidental swipes during tap/drag
		swipeMinDistance: 50.0,

		// Platform parity fix: Additional thresholds for gesture consistency
		doubleTapTolerance: 50.0, // Taps within 50px considered same location
		dragThreshold:      10.0, // 10px minimum to start drag (prevents jitter)
		velocityThreshold:  0.5,  // 0.5 px/ms minimum for fling (300 px/s)

		pinchScale: 1.0,
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
// Platform parity fix: Enhanced double-tap detection with position tolerance
func (g *GestureDetector) detectSingleTouchGestures(touch *Touch) {
	dx := float64(touch.X - touch.StartX)
	dy := float64(touch.Y - touch.StartY)
	distance := math.Sqrt(dx*dx + dy*dy)
	duration := time.Since(touch.StartTime)

	// Tap detection (touch just ended with minimal movement)
	if !touch.Active && distance <= g.tapMaxDistance {
		g.currentTap = true

		// Platform parity fix: Enhanced double tap detection with position tolerance
		// Matches desktop double-click behavior where clicks must be near each other
		tapDx := float64(touch.X - g.lastTapX)
		tapDy := float64(touch.Y - g.lastTapY)
		tapDistance := math.Sqrt(tapDx*tapDx + tapDy*tapDy)

		g.lastTapX = touch.X
		g.lastTapY = touch.Y

		// Double tap requires both timing AND position proximity
		if time.Since(g.lastTapTime) <= g.doubleTapWindow && tapDistance <= g.doubleTapTolerance {
			g.currentDoubleTap = true
			g.tapCount = 0
		} else {
			g.tapCount = 1
		}
		// INTENTIONAL time.Now() usage: Gesture timing detection (non-procgen operational timing).
		// This is NOT part of procedural content generation and does not affect determinism.
		g.lastTapTime = time.Now()
	}

	// Long press detection
	if touch.Active && duration >= g.longPressThreshold && distance <= g.tapMaxDistance {
		g.longPressActive = true
		g.longPressX = touch.X
		g.longPressY = touch.Y
	}

	// Platform parity fix: Enhanced swipe detection with velocity calculation
	if !touch.Active && distance >= g.swipeMinDistance {
		// Calculate velocity (pixels per millisecond)
		durationMs := float64(duration.Milliseconds())
		if durationMs > 0 {
			velocity := distance / durationMs
			g.lastVelocity = velocity

			// Only register as swipe if velocity exceeds threshold (prevents slow drags)
			if velocity >= g.velocityThreshold {
				g.swipeDetected = true
				g.swipeDistance = distance
				g.swipeDirection = math.Atan2(dy, dx)
			}
		}
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

// Platform parity fix: Gesture configuration setters for platform-specific tuning

// SetDoubleTapWindow configures the time window for double-tap detection.
// Platform parity fix: Allows matching platform-specific double-click timing
func (g *GestureDetector) SetDoubleTapWindow(duration time.Duration) {
	g.doubleTapWindow = duration
}

// SetLongPressThreshold configures the duration for long press detection.
// Platform parity fix: iOS uses 500ms, Android often 600ms - allow customization
func (g *GestureDetector) SetLongPressThreshold(duration time.Duration) {
	g.longPressThreshold = duration
}

// SetTapMaxDistance configures maximum movement allowed for tap detection.
// Platform parity fix: Touch needs larger tolerance than mouse due to finger size
func (g *GestureDetector) SetTapMaxDistance(distance float64) {
	g.tapMaxDistance = distance
}

// SetSwipeMinDistance configures minimum distance for swipe detection.
// Platform parity fix: Prevents accidental swipes, tunable per platform DPI
func (g *GestureDetector) SetSwipeMinDistance(distance float64) {
	g.swipeMinDistance = distance
}

// SetDoubleTapTolerance configures maximum distance between taps for double-tap.
// Platform parity fix: Larger tolerance for touch vs mouse precision
func (g *GestureDetector) SetDoubleTapTolerance(distance float64) {
	g.doubleTapTolerance = distance
}

// GetLastVelocity returns the velocity of the last swipe/fling gesture.
// Platform parity fix: Enables velocity-based fling scrolling like mobile OS
func (g *GestureDetector) GetLastVelocity() float64 {
	return g.lastVelocity
}

// Platform parity fix: Coordinate transformation utilities for consistent world-space positioning

// TouchToScreen converts a Touch position to screen coordinates.
// Platform parity fix: No-op helper for API consistency - touches are already in screen space
func TouchToScreen(touch *Touch) (int, int) {
	return touch.X, touch.Y
}

// TouchDelta returns the movement delta for a touch.
// Platform parity fix: Consistent delta calculation across all platforms
func TouchDelta(touch *Touch) (int, int) {
	return touch.DeltaX, touch.DeltaY
}

// TouchDistance returns the total distance traveled by a touch from start.
// Platform parity fix: Used for swipe/drag distance calculations
func TouchDistance(touch *Touch) float64 {
	dx := float64(touch.X - touch.StartX)
	dy := float64(touch.Y - touch.StartY)
	return math.Sqrt(dx*dx + dy*dy)
}

// TouchDuration returns the duration of a touch.
// Platform parity fix: Consistent timing across platforms for gesture detection
func TouchDuration(touch *Touch) time.Duration {
	if touch.Active {
		return time.Since(touch.StartTime)
	}
	if !touch.EndTime.IsZero() {
		return touch.EndTime.Sub(touch.StartTime)
	}
	return 0
}
