package display

import "errors"

// Errors for display package.
var (
	// ErrUnsupportedResolution indicates resolution is not in StandardResolutions.
	ErrUnsupportedResolution = errors.New("unsupported resolution")

	// ErrInvalidConfig indicates configuration has invalid values.
	ErrInvalidConfig = errors.New("invalid configuration")
)
