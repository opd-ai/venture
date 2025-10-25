//go:build ios
// +build ios

// Package mobile provides iOS-specific haptic feedback implementation.
// This file contains iOS Core Haptics integration.
package mobile

/*
#cgo LDFLAGS: -framework CoreHaptics -framework UIKit
#import <UIKit/UIKit.h>

// iOS haptic feedback using UIImpactFeedbackGenerator
void triggerIOSHaptic(int intensity) {
	dispatch_async(dispatch_get_main_queue(), ^{
		UIImpactFeedbackStyle style;
		switch(intensity) {
			case 0: // Light
				style = UIImpactFeedbackStyleLight;
				break;
			case 1: // Medium
				style = UIImpactFeedbackStyleMedium;
				break;
			case 2: // Heavy
				style = UIImpactFeedbackStyleHeavy;
				break;
			default:
				style = UIImpactFeedbackStyleMedium;
		}
		
		UIImpactFeedbackGenerator *generator = [[UIImpactFeedbackGenerator alloc] initWithStyle:style];
		[generator prepare];
		[generator impactOccurred];
	});
}
*/
import "C"

// triggerHapticImpl implements platform-specific haptic feedback for iOS.
func triggerHapticImpl(feedback HapticFeedback) {
	C.triggerIOSHaptic(C.int(feedback))
}
