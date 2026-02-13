// Package saveload provides save name validation used by all platform implementations.
package saveload

import (
	"fmt"
	"strings"
)

// ValidateSaveName validates that a save name is acceptable for use.
// This function is shared between desktop and WASM implementations.
//
// Validation rules:
//   - Name cannot be empty
//   - Name cannot contain path separators (/ or \)
//   - Name cannot contain special characters (<>:"|?*)
//
// Returns an error if the name is invalid, nil otherwise.
func ValidateSaveName(name string) error {
	if name == "" {
		return fmt.Errorf("save name cannot be empty")
	}

	// Remove extension for validation (if present)
	name = strings.TrimSuffix(name, ".sav")

	// Check for path separators (security check)
	if strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("save name cannot contain path separators")
	}

	// Check for special characters
	if strings.ContainsAny(name, "<>:\"|?*") {
		return fmt.Errorf("save name contains invalid characters")
	}

	return nil
}
