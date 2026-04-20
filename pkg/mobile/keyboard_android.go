//go:build android && cgo && ebitenmobilebind
// +build android,cgo,ebitenmobilebind

// Package mobile provides Android-specific keyboard integration.
// This file implements native soft keyboard control via Android's InputMethodManager.
// This file is only included when building with the ebitenmobile bind tool.
package mobile

/*
#cgo LDFLAGS: -landroid
#include <jni.h>
#include <stdlib.h>

// showAndroidKeyboard requests the Android InputMethodManager to show the soft keyboard.
//
// IMPLEMENTATION NOTE: Full JNI integration requires Android NDK environment,
// access to JNIEnv* and the current Activity from the Ebiten mobile runtime.
// When integrated with the gomobile/ebitenmobile build process, this function would:
//
//  1. Receive JNIEnv* and Activity from the ebitenmobile binding layer.
//  2. Call Activity.getSystemService(Context.INPUT_METHOD_SERVICE) via JNI to
//     obtain an InputMethodManager instance.
//  3. Obtain the focused View's IBinder via View.getWindowToken().
//  4. Call InputMethodManager.showSoftInput(view, InputMethodManager.SHOW_IMPLICIT)
//     to display the on-screen keyboard.
//
// The current implementation provides the C function signature for use with the
// JNI bridging infrastructure once it is available from the gomobile runtime.
void showAndroidKeyboard() {
	// Full JNI implementation:
	// JNIEnv *env and jobject activity provided by ebitenmobile binding.
	//
	// jclass contextClass = (*env)->FindClass(env, "android/content/Context");
	// jstring imService = (*env)->NewStringUTF(env, "input_method");
	// jmethodID getSysService = (*env)->GetMethodID(env, contextClass,
	//     "getSystemService", "(Ljava/lang/String;)Ljava/lang/Object;");
	// jobject imm = (*env)->CallObjectMethod(env, activity, getSysService, imService);
	// jclass immClass = (*env)->GetObjectClass(env, imm);
	// jmethodID showSoftInput = (*env)->GetMethodID(env, immClass,
	//     "showSoftInput", "(Landroid/view/View;I)Z");
	// (*env)->CallBooleanMethod(env, imm, showSoftInput, view, 1 /* SHOW_IMPLICIT */);
	//
	// Requires Android NDK environment and runtime JNIEnv* from Ebiten.
}

// hideAndroidKeyboard requests the Android InputMethodManager to hide the soft keyboard.
//
// IMPLEMENTATION NOTE: Requires the same JNI environment as showAndroidKeyboard.
// When integrated, this calls:
//   InputMethodManager.hideSoftInputFromWindow(view.getWindowToken(), 0)
void hideAndroidKeyboard() {
	// Full JNI implementation:
	// jmethodID hideSoftInput = (*env)->GetMethodID(env, immClass,
	//     "hideSoftInputFromWindow", "(Landroid/os/IBinder;I)Z");
	// jobject windowToken = ... view.getWindowToken() ...;
	// (*env)->CallBooleanMethod(env, imm, hideSoftInput, windowToken, 0);
	//
	// Requires Android NDK environment and runtime JNIEnv* from Ebiten.
}
*/
import "C"

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	log "github.com/sirupsen/logrus"
)

// ShowKeyboard shows the native Android soft keyboard by calling
// InputMethodManager.showSoftInput via JNI.
//
// This function requires ebitenmobile bind toolchain and Android NDK.
// SDK integration is pending: the underlying C function has no effect
// until the JNI implementation is complete.
func ShowKeyboard() {
	log.WithField("platform", "android").Debug("ShowKeyboard called (ebitenmobilebind build — JNI SDK integration pending)")
	C.showAndroidKeyboard()
}

// HideKeyboard hides the native Android soft keyboard by calling
// InputMethodManager.hideSoftInputFromWindow via JNI.
//
// This function requires ebitenmobile bind toolchain and Android NDK.
// SDK integration is pending: the underlying C function has no effect
// until the JNI implementation is complete.
func HideKeyboard() {
	log.WithField("platform", "android").Debug("HideKeyboard called (ebitenmobilebind build — JNI SDK integration pending)")
	C.hideAndroidKeyboard()
}

// IsKeyboardSupported reports whether programmatic native soft keyboard control
// is functionally available on this Android ebitenmobilebind build.
//
// The JNI bridge is not yet implemented, so ShowKeyboard and HideKeyboard are
// currently no-ops. This returns false until the JNI integration is complete.
func IsKeyboardSupported() bool {
	return false
}

// GetBackButtonKey returns the Android system back button key.
// In Ebiten, Android's system back button is mapped to ebiten.KeyEscape.
func GetBackButtonKey() ebiten.Key {
	return ebiten.KeyEscape
}

// IsBackButtonPressed checks if the Android system back button was just pressed.
func IsBackButtonPressed() bool {
	return inpututil.IsKeyJustPressed(GetBackButtonKey())
}

// IsBackButtonDown checks if the Android system back button is currently held.
func IsBackButtonDown() bool {
	return ebiten.IsKeyPressed(GetBackButtonKey())
}

// GetBackButtonName returns the human-readable name for the Android back button.
func GetBackButtonName() string {
	return "Back Button"
}
