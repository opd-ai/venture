// Package shapes provides procedural geometric shape generation for sprites and visual elements.
// Shapes are created using mathematical functions and can be combined to create complex visuals.
//
// # Anti-Aliasing (Phase 15.1)
//
// The package supports sub-pixel anti-aliasing for smooth edge rendering, particularly beneficial
// for diagonal edges and curved shapes. Anti-aliasing uses super-sampling to calculate pixel
// coverage, producing partial alpha values at shape boundaries.
//
// Quality Levels:
//   - AntiAliasOff: Hard edges (fastest, legacy behavior)
//   - AntiAliasLow: 2x2 super-sampling (4 samples per pixel)
//   - AntiAliasMedium: 4x4 super-sampling (16 samples per pixel)
//   - AntiAliasHigh: 8x8 super-sampling (64 samples per pixel)
//
// Usage Example:
//
//	gen := shapes.NewGenerator()
//	config := shapes.Config{
//	    Type:      shapes.ShapeCircle,
//	    Width:     32,
//	    Height:    32,
//	    Color:     color.RGBA{R: 255, G: 0, B: 0, A: 255},
//	    AntiAlias: shapes.AntiAliasMedium,  // Enable anti-aliasing
//	}
//	img, err := gen.Generate(config)
//
// Performance:
//   - Off: ~0.02ms per 32x32 shape
//   - Low: ~0.07ms per 32x32 shape
//   - Medium: ~0.22ms per 32x32 shape
//   - High: ~0.79ms per 32x32 shape
//
// All quality levels are well below the 5ms generation target for Phase 15.1.
package shapes
