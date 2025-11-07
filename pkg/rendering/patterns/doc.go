package patterns

// Package patterns provides procedural pattern generation for textures and overlays.
//
// This package implements various pattern types and texture generation for tile rendering:
//
// # Pattern Types (Phase 14)
//
// Basic patterns (stripes, dots, gradients, noise, checkerboard, circles) that can be
// applied to existing images to add visual variety.
//
// # Texture Generation (Phase 16.1)
//
// Advanced procedural texture generation for realistic material appearance:
//   - Stone: Multi-octave Perlin noise for natural rock formations
//   - Wood: Radial grain patterns with turbulence for organic wood appearance
//   - Metal: Anisotropic brushed metal with specular highlights
//   - Organic: Cellular noise with multiple octaves for biological textures
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
// # Usage Example
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
//		log.Fatal(err)
//	}
