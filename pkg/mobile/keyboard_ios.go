//go:build ios && cgo && ebitenmobilebind
// +build ios,cgo,ebitenmobilebind

// Package mobile provides iOS-specific keyboard integration.
// This file implements native soft keyboard control via UIKit's UIResponder API.
// A hidden UITextField is maintained and re-attached to the current root
// UIViewController's view whenever the view hierarchy changes; ShowKeyboard
// calls becomeFirstResponder and HideKeyboard calls resignFirstResponder.
// Key-window lookup uses the iOS 13+ connectedScenes API with a fallback to the
// deprecated UIApplication.keyWindow for older OS versions.
// This file is only included when building with the ebitenmobile bind tool.
package mobile

/*
#cgo LDFLAGS: -framework UIKit
#import <UIKit/UIKit.h>

static UITextField *gHiddenTextField = nil;

// getKeyWindow returns the application's key window, compatible with iOS 13+.
// On iOS 13+ UIApplication.keyWindow is deprecated; the modern approach
// enumerates connectedScenes to find the foreground-active UIWindowScene's
// key window. Falls back to the deprecated API on iOS < 13.
static UIWindow *getKeyWindow(void) {
    if (@available(iOS 13, *)) {
        for (UIScene *scene in [UIApplication sharedApplication].connectedScenes) {
            if (scene.activationState != UISceneActivationStateForegroundActive) continue;
            UIWindowScene *ws = (UIWindowScene *)scene;
            for (UIWindow *window in ws.windows) {
                if (window.isKeyWindow) return window;
            }
            if (ws.windows.count > 0) return ws.windows[0];
        }
    }
    // Fallback for iOS < 13.
#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wdeprecated-declarations"
    return [UIApplication sharedApplication].keyWindow;
#pragma clang diagnostic pop
}

// ensureTextField creates the hidden UITextField if it doesn't exist yet,
// or re-attaches it to rootVC.view if the view hierarchy changed (e.g. after
// a root-view-controller swap). Alpha is 0.01, not 0.0: UIKit skips
// firstResponder handling for fully-transparent views, which would prevent
// the keyboard from appearing.
static void ensureTextField(UIViewController *rootVC) {
    if (gHiddenTextField && gHiddenTextField.superview == rootVC.view) return;
    if (gHiddenTextField) {
        [gHiddenTextField removeFromSuperview];
        gHiddenTextField = nil;
    }
    // Position far off-screen (-10000,-10000) to avoid layout/hit-testing effects.
    gHiddenTextField = [[UITextField alloc] initWithFrame:CGRectMake(-10000, -10000, 1, 1)];
    gHiddenTextField.alpha = 0.01;
    gHiddenTextField.autocorrectionType = UITextAutocorrectionTypeNo;
    gHiddenTextField.autocapitalizationType = UITextAutocapitalizationTypeNone;
    gHiddenTextField.spellCheckingType = UITextSpellCheckingTypeNo;
    [rootVC.view addSubview:gHiddenTextField];
}

// showIOSKeyboard obtains the key window (iOS 13+ compatible), ensures the
// hidden UITextField is attached to the current root-VC view, and calls
// becomeFirstResponder to trigger the UIKit on-screen keyboard.
// Runs on the main queue for UIKit thread safety.
void showIOSKeyboard(void) {
    dispatch_async(dispatch_get_main_queue(), ^{
        UIWindow *window = getKeyWindow();
        if (!window) return;
        UIViewController *rootVC = window.rootViewController;
        if (!rootVC) return;
        ensureTextField(rootVC);
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
