//go:build ios && cgo && ebitenmobilebind
// +build ios,cgo,ebitenmobilebind

// Package mobile provides iOS-specific keyboard integration.
// This file implements native soft keyboard control via UIKit's UIResponder API.
// A hidden UITextField is attached to the root UIViewController's view on first
// use; ShowKeyboard calls becomeFirstResponder and HideKeyboard calls
// resignFirstResponder to show/dismiss the on-screen keyboard.
// This file is only included when building with the ebitenmobile bind tool.
package mobile

/*
#cgo LDFLAGS: -framework UIKit
#import <UIKit/UIKit.h>

static UITextField *gHiddenTextField = nil;

// showIOSKeyboard creates a hidden UITextField (if not yet present), attaches it
// to the root view controller's view, and calls becomeFirstResponder to trigger
// the UIKit on-screen keyboard. Runs on the main queue for thread safety.
void showIOSKeyboard(void) {
    dispatch_async(dispatch_get_main_queue(), ^{
        UIWindow *window = [UIApplication sharedApplication].keyWindow;
        if (!window) return;
        UIViewController *rootVC = window.rootViewController;
        if (!rootVC) return;
        if (!gHiddenTextField) {
            gHiddenTextField = [[UITextField alloc] initWithFrame:CGRectMake(-1, -1, 1, 1)];
            gHiddenTextField.alpha = 0.01;
            gHiddenTextField.autocorrectionType = UITextAutocorrectionTypeNo;
            gHiddenTextField.autocapitalizationType = UITextAutocapitalizationTypeNone;
            gHiddenTextField.spellCheckingType = UITextSpellCheckingTypeNo;
            [rootVC.view addSubview:gHiddenTextField];
        }
        [gHiddenTextField becomeFirstResponder];
    });
}

// hideIOSKeyboard dismisses the iOS on-screen keyboard by calling
// resignFirstResponder on the hidden UITextField. Runs on the main queue.
void hideIOSKeyboard(void) {
    dispatch_async(dispatch_get_main_queue(), ^{
        if (gHiddenTextField) {
            [gHiddenTextField resignFirstResponder];
        }
    });
}
*/
import "C"

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	log "github.com/sirupsen/logrus"
)

// ShowKeyboard shows the native iOS on-screen keyboard by calling
// UIResponder.becomeFirstResponder on a hidden UITextField via the UIKit bridge.
func ShowKeyboard() {
	log.WithField("platform", "ios").Debug("ShowKeyboard called")
	C.showIOSKeyboard()
}

// HideKeyboard dismisses the native iOS on-screen keyboard by calling
// UIResponder.resignFirstResponder on the hidden UITextField.
func HideKeyboard() {
	log.WithField("platform", "ios").Debug("HideKeyboard called")
	C.hideIOSKeyboard()
}

// IsKeyboardSupported returns true now that the UIKit hidden-UITextField
// bridge is implemented via showIOSKeyboard/hideIOSKeyboard.
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
