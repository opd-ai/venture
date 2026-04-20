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
// The current implementation is a placeholder; full integration requires
// the ebitenmobile Objective-C bridge and access to the UIViewController.
void showIOSKeyboard() {
    // Placeholder: SDK integration pending.
    // When implemented, this will call dispatch_async(dispatch_get_main_queue(), ^{
    //   UIWindow *window = [UIApplication sharedApplication].keyWindow;
    //   UIViewController *rootVC = window.rootViewController;
    //   UITextField *hiddenField = [[UITextField alloc] initWithFrame:CGRectMake(-1,-1,1,1)];
    //   hiddenField.alpha = 0.01;
    //   [rootVC.view addSubview:hiddenField];
    //   [hiddenField becomeFirstResponder];
    // });
}

// hideIOSKeyboard dismisses the iOS on-screen keyboard by resigning first responder.
void hideIOSKeyboard() {
    // Placeholder: SDK integration pending.
    // When implemented, this will call dispatch_async(dispatch_get_main_queue(), ^{
    //   UIWindow *window = [UIApplication sharedApplication].keyWindow;
    //   UIView *firstResponder = [window performSelector:@selector(firstResponder)];
    //   [firstResponder resignFirstResponder];
    // });
}
*/
import "C"

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	log "github.com/sirupsen/logrus"
)

// ShowKeyboard shows the native iOS on-screen keyboard by calling
// UIResponder.becomeFirstResponder via the UIKit bridge.
//
// This function requires ebitenmobile bind toolchain and Xcode/iOS SDK.
// SDK integration is pending: the underlying C function has no effect
// until the UIKit implementation is complete.
func ShowKeyboard() {
	log.WithField("platform", "ios").Debug("ShowKeyboard called (ebitenmobilebind build — UIKit SDK integration pending)")
	C.showIOSKeyboard()
}

// HideKeyboard dismisses the native iOS on-screen keyboard by calling
// UIResponder.resignFirstResponder via the UIKit bridge.
//
// This function requires ebitenmobile bind toolchain and Xcode/iOS SDK.
// SDK integration is pending: the underlying C function has no effect
// until the UIKit implementation is complete.
func HideKeyboard() {
	log.WithField("platform", "ios").Debug("HideKeyboard called (ebitenmobilebind build — UIKit SDK integration pending)")
	C.hideIOSKeyboard()
}

// IsKeyboardSupported returns false until the iOS UIKit bridge is fully wired.
//
// Although this file is only built for ios+cgo+ebitenmobilebind, the underlying
// C keyboard bridge functions are currently placeholders with no runtime effect.
// This returns true once the UIKit first-responder integration is complete.
func IsKeyboardSupported() bool {
	return false
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
