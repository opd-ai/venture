// Package rendering provides namespace organization for the rendering subsystem.
// All graphics are generated at runtime without external asset files.
//
// This package serves as a namespace; actual rendering implementations live in subdirectories.
// For color palettes, use [github.com/opd-ai/venture/pkg/rendering/palette.Palette].
// For sprite configuration, use [github.com/opd-ai/venture/pkg/rendering/sprites.Config].
//
// Subdirectories:
//   - sprites: Procedural sprite generation with equipment overlays
//   - animation: Animation system with articulation and caching
//   - tiles: Tile generation with parallax and transitions
//   - lighting: Dynamic lighting with bloom and ambient occlusion
//   - postprocess: Post-processing effects (blur, color grading, vignette)
//   - particles: Particle system with physics and weather effects
//   - ui: UI generation for chat, notifications, tutorials
//   - palette: Color palette generation with time-of-day support
//   - patterns: Texture pattern generation
//   - cache: Sprite caching and predictive warming
//   - pool: Resource pooling for sprites and images
//   - parallel: Parallel rendering utilities
//   - quality: Quality settings and LOD management
//   - display: Display configuration
//   - shapes: Geometric shape primitives
package rendering
