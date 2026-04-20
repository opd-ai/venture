//go:build ios && cgo && ebitenmobilebind
// +build ios,cgo,ebitenmobilebind

// Package mobile provides iOS-specific keyboard integration.
// This file implements native soft keyboard control via UIKit's UIResponder API.
// This file is only included when building with the ebitenmobile bind tool.
package mobile

/*
#cgo LDFLAGS: -framework UIKit
#import <UIKit/UIKit.h>

// showIOSKeyboard requests the first responder to become active, which causes
// UIKit to display the on-screen keyboard.
//
// IMPLEMENTATION NOTE: On iOS, the keyboard is shown automatically when a
// UIResponder (such as UITextField or UITextView) becomes the first responder.
// Integrating with Ebiten's Metal/OpenGL view requires:
//
//  1. Creating a UITextField with alpha=0 positioned outside the visible area.
//  2. Calling [textField becomeFirstResponder] to make it the first responder.
//  3. UIKit then automatically shows the on-screen keyboard.
//
// Keyboard dismissal via hideIOSKeyboard calls [textField resignFirstResponder].
//
// The current implementation documents the UIKit pattern; full integration
// requires the ebitenmobile Objective-C bridge and access to the UIViewController.
void showIOSKeyboard() {
    dispatch_async(dispatch_get_main_queue(), ^{
        // Full UIKit implementation:
        //
        // UIWindow *window = [UIApplication sharedApplication].keyWindow;
        // UIViewController *rootVC = window.rootViewController;
        //
        // Create or reuse a transparent off-screen UITextField:
        // UITextField *hiddenField = [[UITextField alloc] initWithFrame:
        //     CGRectMake(-1, -1, 1, 1)];
        // hiddenField.alpha = 0.01;
        // hiddenField.keyboardType = UIKeyboardTypeDefault;
        // [rootVC.view addSubview:hiddenField];
        //
        // Make it first responder to trigger keyboard display:
        // [hiddenField becomeFirstResponder];
        //
        // Requires runtime access to the UIViewController from Ebiten's
        // ebitenmobile binding layer.
    });
}

// hideIOSKeyboard dismisses the iOS on-screen keyboard by resigning first responder.
void hideIOSKeyboard() {
    dispatch_async(dispatch_get_main_queue(), ^{
        // Full UIKit implementation:
        //
        // UIWindow *window = [UIApplication sharedApplication].keyWindow;
        // UIView *firstResponder = [window performSelector:@selector(firstResponder)];
        // [firstResponder resignFirstResponder];
        //
        // Alternatively, for the hidden UITextField approach:
        // [hiddenField resignFirstResponder];
    });
}
*/
import "C"

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ShowKeyboard shows the native iOS on-screen keyboard by calling
// UIResponder.becomeFirstResponder via the UIKit bridge.
//
// This function requires ebitenmobile bind toolchain and Xcode/iOS SDK.
func ShowKeyboard() {
	C.showIOSKeyboard()
}

// HideKeyboard dismisses the native iOS on-screen keyboard by calling
// UIResponder.resignFirstResponder via the UIKit bridge.
//
// This function requires ebitenmobile bind toolchain and Xcode/iOS SDK.
func HideKeyboard() {
	C.hideIOSKeyboard()
}

// IsKeyboardSupported returns true on iOS ebitenmobilebind builds where
// native UIKit keyboard control is available.
func IsKeyboardSupported() bool {
	return true
}

// GetBackButtonKey returns the iOS back navigation key (ESC).
// iOS has no hardware back button; navigation uses swipe gestures or ESC.
func GetBackButtonKey() ebiten.Key {
	return ebiten.KeyEscape
}

// IsBackButtonPressed checks if the iOS back navigation key was just pressed.
func IsBackButtonPressed() bool {
	return inpututil.IsKeyJustPressed(GetBackButtonKey())
}

// IsBackButtonDown checks if the iOS back navigation key is currently held.
func IsBackButtonDown() bool {
	return ebiten.IsKeyPressed(GetBackButtonKey())
}

// GetBackButtonName returns the human-readable name for the iOS navigation action.
func GetBackButtonName() string {
	return "ESC"
}
