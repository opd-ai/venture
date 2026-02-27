// Package parity provides cross-platform visual parity testing for Venture.
//
// This package validates that visual rendering is consistent across all supported
// platforms: desktop (Linux, macOS, Windows), web (WebAssembly), and mobile
// (iOS, Android). It ensures the zero-asset procedural generation produces
// identical visual output regardless of platform.
//
// # Platform Detection
//
// The package automatically detects the current platform using Go's runtime
// package and GOOS/GOARCH constants. Platform-specific rendering paths are
// tested separately.
//
// # Parity Tests
//
// The package implements 10 core parity tests:
//  1. Sprite rendering: identical appearance across platforms
//  2. Color accuracy: RGB values match ±1 (gamma correction)
//  3. Frame rate: 60 FPS achieved on target hardware
//  4. Resolution scaling: UI scales correctly
//  5. Font rendering: clear, anti-aliased text
//  6. Touch input: mobile controls functional
//  7. Fullscreen: desktop fullscreen working
//  8. WebGL rendering: WASM matches desktop quality
//  9. Pixel-perfect collision: same hitboxes
//  10. Performance: within 20% of best platform
//
// # Usage Example
//
//	validator := parity.NewValidator()
//	result := validator.ValidateSprites(seed, genreID)
//	if !result.Passed {
//	    // Production code should use structured logging:
//	    // logrus.WithFields(logrus.Fields{"errors": result.Errors}).Error("Sprite parity failed")
//	}
//
// # Acceptance Criteria
//
// - Visual consistency: <5% perceived difference across platforms
// - Feature parity: all features work on desktop/web/mobile
// - Performance: 60 FPS on target hardware for each platform
//
// # See Also
//
// Phase 63.3 of ROADMAP_V10.md defines the full acceptance criteria.
// The cmd/paritytest tool provides a CLI interface for manual validation.
package parity
