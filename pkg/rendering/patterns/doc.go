package patterns

// Package patterns provides procedural pattern generation for textures and overlays.
//
// This package implements various pattern types and texture generation for tile rendering:
//
// # Pattern Types (Phase 14)
//
// Basic patterns (stripes, dots, gradients, noise, checkerboard, circles) that can be
// applied to existing images to add visual variety. Use [Generator.GeneratePattern]
// with [Config] for these primitive patterns.
//
// # Texture Generation (Phase 16.1)
//
// Advanced procedural texture generation for realistic material appearance:
//   - Stone: Multi-octave Perlin noise for natural rock formations
//   - Wood: Radial grain patterns with turbulence for organic wood appearance
//   - Metal: Anisotropic brushed metal with specular highlights
//   - Organic: Cellular noise with multiple octaves for biological textures
//
// Use [Generator.Generate] with [TextureConfig] for material textures.
//
// Textures support:
//   - Genre-specific variations (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic)
//   - Configurable detail levels for fine surface features
//   - Normal map approximation for depth perception
//   - Per-pixel variation to prevent repetitive patterns
//   - Deterministic generation (same seed = identical texture)
//
// # Performance
//
// Generation times for 32x32 textures:
//   - Stone: ~1-2ms
//   - Wood: ~1-2ms
//   - Metal: ~1-2ms
//   - Organic: ~1-2ms
//
// All patterns are generated deterministically using seed-based RNG,
// ensuring reproducible results across game sessions and multiplayer clients.
//
// # Texture Generation Example
//
//	gen := patterns.NewGenerator()
//	config := patterns.TextureConfig{
//		Texture:     patterns.TextureStone,
//		Width:       32,
//		Height:      32,
//		GenreID:     "fantasy",
//		Seed:        12345,
//		Color1:      color.RGBA{R: 80, G: 60, B: 40, A: 255},
//		Color2:      color.RGBA{R: 140, G: 120, B: 100, A: 255},
//		DetailLevel: 0.5,
//		Scale:       0.1,
//	}
//	texture, err := gen.Generate(config)
//	if err != nil {
//		logrus.WithError(err).Fatal("failed to generate texture")
//	}
//
// # Basic Pattern Generation Example
//
//	gen := patterns.NewGenerator()
//	config := patterns.Config{
//		Type:      patterns.PatternStripes,
//		Width:     32,
//		Height:    32,
//		Seed:      12345,
//		Frequency: 4.0,
//		Amplitude: 0.5,
//		Angle:     45,
//		Color1:    color.RGBA{R: 255, G: 255, B: 255, A: 255},
//		Color2:    color.RGBA{R: 0, G: 0, B: 0, A: 255},
//		Opacity:   1.0,
//	}
//	pattern, err := gen.GeneratePattern(config)
//	if err != nil {
//		logrus.WithError(err).Fatal("failed to generate pattern")
//	}
