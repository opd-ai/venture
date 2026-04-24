//go:build js || android || ios
// +build js android ios

// Package engine provides character creation functionality for onboarding new players.
// This file implements the character creation UI for mobile and WASM platforms.
//
// # Portrait Selection on Mobile / WASM
//
// Native file-picker dialogs (zenity) are not available in browser sandboxes or on
// iOS/Android without a native bridge. On these platforms portrait import is
// deliberately unavailable and the character-creation UI hides the "Browse" button.
//
// G12 (AUDIT.md): a future native bridge (pkg/mobile.OpenImagePicker) could route
// iOS photo-library / Android media-picker access here.  Until that bridge exists,
// players select a preset name (procedurally generated portrait) or skip the step.
package engine

// ErrPortraitDialogUnsupported is returned by OpenPortraitDialog on mobile/WASM.
// Callers should surface this as a UI message rather than a hard error, and the
// character-creation UI already hides the Browse button on these platforms.
var ErrPortraitDialogUnsupported = errPortraitDialogUnsupported("portrait import is not available on mobile/WASM — use preset names or skip")

type errPortraitDialogUnsupported string

func (e errPortraitDialogUnsupported) Error() string { return string(e) }

// OpenPortraitDialog is not available on mobile/WASM platforms.
// Returns ErrPortraitDialogUnsupported so callers can distinguish platform
// limitations from generic I/O failures.
func OpenPortraitDialog() (string, error) {
	return "", ErrPortraitDialogUnsupported
}
