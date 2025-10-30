//go:build android && cgo && ebitenmobilebind
// +build android,cgo,ebitenmobilebind

// Package mobile provides Android-specific haptic feedback implementation.
// This file contains Android Vibrator service integration.
// This file is only included when building with ebitenmobile bind tool.
package mobile

/*
#cgo LDFLAGS: -landroid
#include <jni.h>
#include <stdlib.h>

// Android haptic feedback using Vibrator service
// Note: Requires android.permission.VIBRATE in AndroidManifest.xml
//
// IMPLEMENTATION NOTE: Full JNI integration requires Android NDK environment,
// access to JNIEnv*, and activity context from the Ebiten mobile runtime.
// This would need to be integrated with the gomobile build process and
// requires testing on an actual Android device. The current implementation
// provides the C function signature and duration calculations.
//
// When integrated with Android NDK, this function would:
// 1. Receive JNIEnv* and activity context from gomobile/Ebiten
// 2. Call Context.getSystemService(Context.VIBRATOR_SERVICE) via JNI
// 3. Call Vibrator.vibrate(duration) via JNI
// 4. Handle API level differences (VibrationEffect for API 26+)
void triggerAndroidHaptic(int intensity) {
	// Calculate duration based on intensity
	// Light: 10ms, Medium: 20ms, Heavy: 50ms
	int duration = 20;
	switch(intensity) {
		case 0: duration = 10; break;
		case 1: duration = 20; break;
		case 2: duration = 50; break;
	}

	// Full JNI implementation requires Android NDK environment
	// The following is example code structure for reference:
	//
	// Get JNIEnv and activity context from gomobile
	// Find Context class and getSystemService method
	// Get Vibrator service
	// Find Vibrator class and vibrate method
	// Call vibrate with calculated duration
	//
	// This cannot be completed without Android NDK build environment
	// and runtime access to JNIEnv from Ebiten's gomobile integration
}
*/
import "C" // triggerHapticImpl implements platform-specific haptic feedback for Android.
func triggerHapticImpl(feedback HapticFeedback) {
	C.triggerAndroidHaptic(C.int(feedback))
}
