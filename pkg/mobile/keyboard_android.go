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

#define SHOW_IMPLICIT 1
#define HIDE_NO_FLAGS 0

static JavaVM *gJVM = NULL;
static jobject gActivity = NULL;

// JNI_OnLoad caches the JavaVM pointer when the shared library is loaded.
// This provides the JVM reference needed to attach goroutine threads later.
JNIEXPORT jint JNI_OnLoad(JavaVM *vm, void *reserved) {
	gJVM = vm;
	return JNI_VERSION_1_6;
}

// setAndroidActivity stores the Android Activity as a global JNI reference.
// envPtr and activityPtr are validated non-nil before use. Must be called from
// the Java ebitenmobile wrapper (via SetAndroidActivity) before ShowKeyboard or
// HideKeyboard is used.
void setAndroidActivity(void *envPtr, void *activityPtr) {
	if (!envPtr || !activityPtr) return;
	JNIEnv *env = (JNIEnv *)envPtr;
	jobject activity = (jobject)activityPtr;
	if (gActivity) {
		(*env)->DeleteGlobalRef(env, gActivity);
		gActivity = NULL;
	}
	gActivity = (*env)->NewGlobalRef(env, activity);
	if (!gActivity) {
		(*env)->ExceptionClear(env);
	}
}

// hasAndroidActivity returns JNI_TRUE when the Activity global ref is set,
// i.e. SetAndroidActivity has been called successfully. Used by Go-side
// IsKeyboardSupported to report readiness accurately.
jboolean hasAndroidActivity(void) {
	return gActivity != NULL ? JNI_TRUE : JNI_FALSE;
}

// getJNIEnv returns the JNIEnv for the current thread. If the thread was not
// already attached to the JVM it calls AttachCurrentThread and sets *didAttach
// to JNI_TRUE; the caller is then responsible for calling DetachCurrentThread
// when the JNI work is done to prevent unbounded thread-attachment leaks.
static JNIEnv *getJNIEnv(jboolean *didAttach) {
	*didAttach = JNI_FALSE;
	if (!gJVM) return NULL;
	JNIEnv *env = NULL;
	jint res = (*gJVM)->GetEnv(gJVM, (void **)&env, JNI_VERSION_1_6);
	if (res == JNI_OK) return env;
	if (res == JNI_EDETACHED) {
		if ((*gJVM)->AttachCurrentThread(gJVM, &env, NULL) != JNI_OK) return NULL;
		*didAttach = JNI_TRUE;
		return env;
	}
	return NULL;
}

// showAndroidKeyboard calls InputMethodManager.showSoftInput on the Activity's
// DecorView to display the Android on-screen keyboard.
//
// All JNI local references are managed inside a PushLocalFrame/PopLocalFrame
// scope to prevent the local reference table from overflowing on repeated calls.
void showAndroidKeyboard(void) {
	jboolean didAttach = JNI_FALSE;
	JNIEnv *env = getJNIEnv(&didAttach);
	if (!env || !gActivity) goto cleanup_show;
	if ((*env)->PushLocalFrame(env, 16) < 0) {
		(*env)->ExceptionClear(env);
		goto cleanup_show;
	}
	do {
		jclass ctxClass = (*env)->FindClass(env, "android/content/Context");
		if (!ctxClass) { (*env)->ExceptionClear(env); break; }
		jfieldID imsFld = (*env)->GetStaticFieldID(env, ctxClass, "INPUT_METHOD_SERVICE", "Ljava/lang/String;");
		if (!imsFld) { (*env)->ExceptionClear(env); break; }
		jstring imsStr = (jstring)(*env)->GetStaticObjectField(env, ctxClass, imsFld);

		jmethodID getSvc = (*env)->GetMethodID(env, ctxClass, "getSystemService", "(Ljava/lang/String;)Ljava/lang/Object;");
		if (!getSvc) { (*env)->ExceptionClear(env); break; }
		jobject imm = (*env)->CallObjectMethod(env, gActivity, getSvc, imsStr);
		if (!imm) { (*env)->ExceptionClear(env); break; }

		jclass actClass = (*env)->GetObjectClass(env, gActivity);
		jmethodID getWin = (*env)->GetMethodID(env, actClass, "getWindow", "()Landroid/view/Window;");
		if (!getWin) { (*env)->ExceptionClear(env); break; }
		jobject win = (*env)->CallObjectMethod(env, gActivity, getWin);
		if (!win) { (*env)->ExceptionClear(env); break; }

		jclass winClass = (*env)->GetObjectClass(env, win);
		jmethodID getDecor = (*env)->GetMethodID(env, winClass, "getDecorView", "()Landroid/view/View;");
		if (!getDecor) { (*env)->ExceptionClear(env); break; }
		jobject decorView = (*env)->CallObjectMethod(env, win, getDecor);
		if (!decorView) { (*env)->ExceptionClear(env); break; }

		jclass immClass = (*env)->GetObjectClass(env, imm);
		jmethodID showSoftInput = (*env)->GetMethodID(env, immClass, "showSoftInput", "(Landroid/view/View;I)Z");
		if (!showSoftInput) { (*env)->ExceptionClear(env); break; }
		(*env)->CallBooleanMethod(env, imm, showSoftInput, decorView, SHOW_IMPLICIT);
		(*env)->ExceptionClear(env);
	} while (0);
	(*env)->PopLocalFrame(env, NULL);
cleanup_show:
	if (didAttach && gJVM) (*gJVM)->DetachCurrentThread(gJVM);
}

// hideAndroidKeyboard calls InputMethodManager.hideSoftInputFromWindow on the
// Activity's DecorView to dismiss the Android on-screen keyboard.
//
// All JNI local references are managed inside a PushLocalFrame/PopLocalFrame
// scope to prevent the local reference table from overflowing on repeated calls.
void hideAndroidKeyboard(void) {
	jboolean didAttach = JNI_FALSE;
	JNIEnv *env = getJNIEnv(&didAttach);
	if (!env || !gActivity) goto cleanup_hide;
	if ((*env)->PushLocalFrame(env, 16) < 0) {
		(*env)->ExceptionClear(env);
		goto cleanup_hide;
	}
	do {
		jclass ctxClass = (*env)->FindClass(env, "android/content/Context");
		if (!ctxClass) { (*env)->ExceptionClear(env); break; }
		jfieldID imsFld = (*env)->GetStaticFieldID(env, ctxClass, "INPUT_METHOD_SERVICE", "Ljava/lang/String;");
		if (!imsFld) { (*env)->ExceptionClear(env); break; }
		jstring imsStr = (jstring)(*env)->GetStaticObjectField(env, ctxClass, imsFld);

		jmethodID getSvc = (*env)->GetMethodID(env, ctxClass, "getSystemService", "(Ljava/lang/String;)Ljava/lang/Object;");
		if (!getSvc) { (*env)->ExceptionClear(env); break; }
		jobject imm = (*env)->CallObjectMethod(env, gActivity, getSvc, imsStr);
		if (!imm) { (*env)->ExceptionClear(env); break; }

		jclass actClass = (*env)->GetObjectClass(env, gActivity);
		jmethodID getWin = (*env)->GetMethodID(env, actClass, "getWindow", "()Landroid/view/Window;");
		if (!getWin) { (*env)->ExceptionClear(env); break; }
		jobject win = (*env)->CallObjectMethod(env, gActivity, getWin);
		if (!win) { (*env)->ExceptionClear(env); break; }

		jclass winClass = (*env)->GetObjectClass(env, win);
		jmethodID getDecor = (*env)->GetMethodID(env, winClass, "getDecorView", "()Landroid/view/View;");
		if (!getDecor) { (*env)->ExceptionClear(env); break; }
		jobject decorView = (*env)->CallObjectMethod(env, win, getDecor);
		if (!decorView) { (*env)->ExceptionClear(env); break; }

		jclass viewClass = (*env)->GetObjectClass(env, decorView);
		jmethodID getWinTok = (*env)->GetMethodID(env, viewClass, "getWindowToken", "()Landroid/os/IBinder;");
		if (!getWinTok) { (*env)->ExceptionClear(env); break; }
		jobject winTok = (*env)->CallObjectMethod(env, decorView, getWinTok);
		if (!winTok) { (*env)->ExceptionClear(env); break; }

		jclass immClass = (*env)->GetObjectClass(env, imm);
		jmethodID hideSoftInput = (*env)->GetMethodID(env, immClass, "hideSoftInputFromWindow", "(Landroid/os/IBinder;I)Z");
		if (!hideSoftInput) { (*env)->ExceptionClear(env); break; }
		(*env)->CallBooleanMethod(env, imm, hideSoftInput, winTok, HIDE_NO_FLAGS);
		(*env)->ExceptionClear(env);
	} while (0);
	(*env)->PopLocalFrame(env, NULL);
cleanup_hide:
	if (didAttach && gJVM) (*gJVM)->DetachCurrentThread(gJVM);
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
// Zero values are rejected to prevent null-pointer dereferences in C.
func SetAndroidActivity(envPtr, activityPtr uintptr) {
	if envPtr == 0 || activityPtr == 0 {
		log.WithField("platform", "android").Warn("SetAndroidActivity called with nil pointer; ignoring")
		return
	}
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
// is available. Returns true only after SetAndroidActivity has been called
// successfully and the Activity global reference is held.
func IsKeyboardSupported() bool {
	return C.hasAndroidActivity() == C.JNI_TRUE
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
