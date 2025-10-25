//go:build android
// +build android

// Package mobile provides Android-specific haptic feedback implementation.
// This file contains Android Vibrator service integration.
package mobile

/*
#cgo LDFLAGS: -landroid
#include <jni.h>
#include <stdlib.h>

// Android haptic feedback using Vibrator service
// Note: Requires android.permission.VIBRATE in AndroidManifest.xml
void triggerAndroidHaptic(int intensity) {
	// This is a placeholder for JNI integration
	// Actual implementation would require:
	// 1. Get JNIEnv and activity context
	// 2. Get Vibrator service via Context.getSystemService
	// 3. Call vibrate() with appropriate duration

	// Example durations for different intensities:
	// Light: 10ms, Medium: 20ms, Heavy: 50ms
	int duration = 20; // Default medium
	switch(intensity) {
		case 0: duration = 10; break;  // Light
		case 1: duration = 20; break;  // Medium
		case 2: duration = 50; break;  // Heavy
	}

	// TODO: Implement actual JNI calls to Android Vibrator API
	// This requires access to JNIEnv and activity context from Ebiten/mobile
}
*/
import "C"

// triggerHapticImpl implements platform-specific haptic feedback for Android.
func triggerHapticImpl(feedback HapticFeedback) {
	C.triggerAndroidHaptic(C.int(feedback))
}
