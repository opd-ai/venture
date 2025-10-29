//go:build js || android || ios
// +build js android ios

// Package engine provides character creation functionality for onboarding new players.
// This file implements the character creation UI for mobile and WASM platforms.
// File dialog functionality is not available on these platforms.
package engine

import (
	"fmt"
)

// OpenPortraitDialog is not available on mobile/WASM platforms.
// Returns an error indicating the feature is not supported.
func OpenPortraitDialog() (string, error) {
	return "", fmt.Errorf("file dialogs are not supported on mobile/WASM platforms")
}
