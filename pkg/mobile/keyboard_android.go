//go:build android && cgo && ebitenmobilebind
// +build android,cgo,ebitenmobilebind

// Package mobile provides Android-specific keyboard integration.
// This file implements native soft keyboard control via Android's InputMethodManager
// using JNI. The JNI bridge is activated by calling SetAndroidActivity from the
// ebitenmobile Java wrapper, which provides the Activity reference required to
// access InputMethodManager. This file is only included when building with the
// ebitenmobile bind tool.
package mobile

/*
#cgo LDFLAGS: -landroid
#include <jni.h>
#include <stdlib.h>
#include <string.h>

static JavaVM *gJVM = NULL;
static jobject gActivity = NULL;

// JNI_OnLoad caches the JavaVM pointer when the shared library is loaded.
// This provides the JVM reference needed to attach goroutine threads later.
JNIEXPORT jint JNI_OnLoad(JavaVM *vm, void *reserved) {
	gJVM = vm;
	return JNI_VERSION_1_6;
}

// setAndroidActivity stores the Android Activity as a global JNI reference.
// Must be called from the Java ebitenmobile wrapper (via SetAndroidActivity)
// before ShowKeyboard or HideKeyboard is used.
void setAndroidActivity(void *envPtr, void *activityPtr) {
	JNIEnv *env = (JNIEnv *)envPtr;
	jobject activity = (jobject)activityPtr;
	if (gActivity) {
		(*env)->DeleteGlobalRef(env, gActivity);
	}
	gActivity = (*env)->NewGlobalRef(env, activity);
}

// getJNIEnv attaches the calling thread to the JVM and returns its JNIEnv.
static JNIEnv *getJNIEnv(void) {
	if (!gJVM) return NULL;
	JNIEnv *env = NULL;
	if ((*gJVM)->AttachCurrentThread(gJVM, &env, NULL) != JNI_OK) return NULL;
	return env;
}

// showAndroidKeyboard calls InputMethodManager.showSoftInput on the Activity's
// DecorView to display the Android on-screen keyboard.
void showAndroidKeyboard(void) {
	JNIEnv *env = getJNIEnv();
	if (!env || !gActivity) return;

	jclass ctxClass = (*env)->FindClass(env, "android/content/Context");
	if (!ctxClass) return;
	jfieldID imsFld = (*env)->GetStaticFieldID(env, ctxClass, "INPUT_METHOD_SERVICE", "Ljava/lang/String;");
	if (!imsFld) return;
	jstring imsStr = (jstring)(*env)->GetStaticObjectField(env, ctxClass, imsFld);

	jmethodID getSvc = (*env)->GetMethodID(env, ctxClass, "getSystemService", "(Ljava/lang/String;)Ljava/lang/Object;");
	if (!getSvc) return;
	jobject imm = (*env)->CallObjectMethod(env, gActivity, getSvc, imsStr);
	if (!imm) return;

	jclass actClass = (*env)->GetObjectClass(env, gActivity);
	jmethodID getWin = (*env)->GetMethodID(env, actClass, "getWindow", "()Landroid/view/Window;");
	if (!getWin) return;
	jobject win = (*env)->CallObjectMethod(env, gActivity, getWin);
	if (!win) return;

	jclass winClass = (*env)->GetObjectClass(env, win);
	jmethodID getDecor = (*env)->GetMethodID(env, winClass, "getDecorView", "()Landroid/view/View;");
	if (!getDecor) return;
	jobject decorView = (*env)->CallObjectMethod(env, win, getDecor);
	if (!decorView) return;

	jclass immClass = (*env)->GetObjectClass(env, imm);
	jmethodID showSoftInput = (*env)->GetMethodID(env, immClass, "showSoftInput", "(Landroid/view/View;I)Z");
	if (!showSoftInput) return;
	(*env)->CallBooleanMethod(env, imm, showSoftInput, decorView, 1 /* SHOW_IMPLICIT */);
}

// hideAndroidKeyboard calls InputMethodManager.hideSoftInputFromWindow on the
// Activity's DecorView to dismiss the Android on-screen keyboard.
void hideAndroidKeyboard(void) {
	JNIEnv *env = getJNIEnv();
	if (!env || !gActivity) return;

	jclass ctxClass = (*env)->FindClass(env, "android/content/Context");
	if (!ctxClass) return;
	jfieldID imsFld = (*env)->GetStaticFieldID(env, ctxClass, "INPUT_METHOD_SERVICE", "Ljava/lang/String;");
	if (!imsFld) return;
	jstring imsStr = (jstring)(*env)->GetStaticObjectField(env, ctxClass, imsFld);

	jmethodID getSvc = (*env)->GetMethodID(env, ctxClass, "getSystemService", "(Ljava/lang/String;)Ljava/lang/Object;");
	if (!getSvc) return;
	jobject imm = (*env)->CallObjectMethod(env, gActivity, getSvc, imsStr);
	if (!imm) return;

	jclass actClass = (*env)->GetObjectClass(env, gActivity);
	jmethodID getWin = (*env)->GetMethodID(env, actClass, "getWindow", "()Landroid/view/Window;");
	if (!getWin) return;
	jobject win = (*env)->CallObjectMethod(env, gActivity, getWin);
	if (!win) return;

	jclass winClass = (*env)->GetObjectClass(env, win);
	jmethodID getDecor = (*env)->GetMethodID(env, winClass, "getDecorView", "()Landroid/view/View;");
	if (!getDecor) return;
	jobject decorView = (*env)->CallObjectMethod(env, win, getDecor);
	if (!decorView) return;

	jclass viewClass = (*env)->GetObjectClass(env, decorView);
	jmethodID getWinTok = (*env)->GetMethodID(env, viewClass, "getWindowToken", "()Landroid/os/IBinder;");
	if (!getWinTok) return;
	jobject winTok = (*env)->CallObjectMethod(env, decorView, getWinTok);
	if (!winTok) return;

	jclass immClass = (*env)->GetObjectClass(env, imm);
	jmethodID hideSoftInput = (*env)->GetMethodID(env, immClass, "hideSoftInputFromWindow", "(Landroid/os/IBinder;I)Z");
	if (!hideSoftInput) return;
	(*env)->CallBooleanMethod(env, imm, hideSoftInput, winTok, 0);
}
*/
import "C"

import (
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	log "github.com/sirupsen/logrus"
)

// SetAndroidActivity stores the Android Activity for keyboard control.
// This must be called from the Java ebitenmobile wrapper before ShowKeyboard
// or HideKeyboard is used. envPtr is a JNIEnv* and activityPtr is a jobject,
// both cast to uintptr by the calling Java-side JNI glue.
func SetAndroidActivity(envPtr, activityPtr uintptr) {
	C.setAndroidActivity(unsafe.Pointer(envPtr), unsafe.Pointer(activityPtr))
}

// ShowKeyboard shows the native Android soft keyboard via
// InputMethodManager.showSoftInput on the Activity's DecorView.
// Requires SetAndroidActivity to have been called by the Java wrapper first.
func ShowKeyboard() {
	log.WithField("platform", "android").Debug("ShowKeyboard called")
	C.showAndroidKeyboard()
}

// HideKeyboard hides the native Android soft keyboard via
// InputMethodManager.hideSoftInputFromWindow on the Activity's DecorView.
// Requires SetAndroidActivity to have been called by the Java wrapper first.
func HideKeyboard() {
	log.WithField("platform", "android").Debug("HideKeyboard called")
	C.hideAndroidKeyboard()
}

// IsKeyboardSupported reports whether programmatic native soft keyboard control
// is available on this Android ebitenmobilebind build.
// Returns true once the JNI bridge is wired via SetAndroidActivity.
func IsKeyboardSupported() bool {
	return true
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
