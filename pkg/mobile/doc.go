// Package mobile provides touch input handling and mobile UI components for iOS and Android platforms.
//
// This package implements touch-first controls including:
//   - Touch input detection and gesture recognition (tap, swipe, pinch)
//   - Virtual controls (D-pad, action buttons)
//   - Mobile-optimized UI layouts
//   - Haptic feedback support
//   - Orientation handling (portrait/landscape)
//
// The mobile package integrates with the engine package to provide feature parity
// with desktop controls while maintaining native mobile UX patterns.
//
// # Touch Input
//
// Touch input is processed through the TouchInputHandler which tracks all active
// touches via ebiten.TouchIDs(). Each touch is identified by a unique TouchID
// and provides position and pressure data.
//
// Example usage:
//
//	handler := mobile.NewTouchInputHandler()
//	handler.Update()
//	if handler.IsTapping() {
//	    x, y := handler.GetTapPosition()
//	    // Handle tap at (x, y)
//	}
//
// # Virtual Controls
//
// Virtual on-screen controls provide tactile feedback for games requiring
// continuous input. The VirtualDPad and VirtualButton types render on-screen
// controls and detect touch interaction.
//
// Example usage:
//
//	dpad := mobile.NewVirtualDPad(100, 400, 80)
//	dpad.Update()
//	moveX, moveY := dpad.GetDirection()
//
// # Gestures
//
// The GestureDetector recognizes common mobile gestures:
//   - Tap (single touch down/up)
//   - Double tap (two quick taps)
//   - Long press (touch held > threshold)
//   - Swipe (fast directional movement)
//   - Pinch (two-finger zoom)
//
// # Platform Detection
//
// The IsMobilePlatform() function detects iOS and Android at runtime,
// allowing conditional behavior for mobile vs desktop.
package mobile
