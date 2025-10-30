//go:build !js && !android && !ios
// +build !js,!android,!ios

// Package engine provides desktop-specific character creation functionality.
// This file implements the file dialog functionality using zenity for desktop platforms.
package engine

import (
	"fmt"

	"github.com/ncruces/zenity"
)

// OpenPortraitDialog opens a native file picker dialog for selecting a portrait image.
// Returns the selected file path or empty string if cancelled.
// This desktop implementation uses zenity for cross-platform native dialogs.
func OpenPortraitDialog() (string, error) {
	// Start in user's Pictures directory
	defaultDir := GetDefaultPicturesDirectory()

	// Create dialog with PNG filter
	// Zenity uses native dialogs on each platform (Windows, macOS, Linux)
	filename, err := zenity.SelectFile(
		zenity.Title("Select Portrait Image"),
		zenity.Filename(defaultDir),
		zenity.FileFilter{
			Name:     "PNG Images",
			Patterns: []string{"*.png"},
			CaseFold: false,
		},
	)
	if err != nil {
		// User cancelled (zenity.ErrCanceled) or error occurred
		if err == zenity.ErrCanceled {
			return "", nil // Not an error, user cancelled
		}
		return "", fmt.Errorf("file dialog error: %w", err)
	}

	return filename, nil
}
