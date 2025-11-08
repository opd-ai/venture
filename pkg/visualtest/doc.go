// Package visualtest provides utilities for visual testing and validation of procedurally generated content.
//
// This package enables developers to visually inspect and validate generated content such as sprites,
// tiles, UI elements, and other visual assets. It provides tools to capture snapshots of generated
// content with genre-specific styling and compare visual output for consistency.
//
// Key Features:
//   - Genre-based visual snapshot generation
//   - Side-by-side comparison of generated content
//   - Validation of visual consistency across genres
//   - Support for testing sprite generation, tile rendering, and UI components
//
// Usage Example:
//
//	// Create a snapshot generator for fantasy genre
//	gen := visualtest.NewSnapshotGenerator()
//	snapshot, err := gen.GenerateSnapshot(seed, "fantasy")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Compare two snapshots
//	diff := visualtest.CompareSnapshots(snapshot1, snapshot2)
//	if diff > 0.1 {
//	    log.Printf("Snapshots differ by %.2f%%", diff*100)
//	}
//
// The package is particularly useful for:
//   - Regression testing of visual generation algorithms
//   - Validating deterministic generation (same seed produces same output)
//   - Manual inspection during development
//   - Generating reference images for documentation
//
// Integration with Testing:
//
// Visual tests use stub implementations (StubSprite, StubInput) to avoid Ebiten dependencies
// in CI environments, allowing visual validation without requiring a display.
package visualtest
