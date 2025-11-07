// Package lighting provides dynamic lighting effects for rendered scenes.
// This package implements a lighting system with support for multiple light sources,
// color modulation, realistic light falloff, bloom/glow effects, and ambient occlusion.
//
// Light sources can be attached to entities or placed in the environment, and their
// effects are blended together to create the final lighting for each pixel.
//
// Phase 17.1 enhancements (November 2025):
//   - Bloom/glow effects for bright light sources with Gaussian blur
//   - Enhanced ambient occlusion with corner and edge detection
//   - Screen-space post-processing pipeline
//
// Key features:
//   - Multiple light source types (point, directional, ambient)
//   - Light intensity and radius control
//   - Color tinting for atmospheric effects
//   - Light falloff with distance (linear, quadratic, or inverse square)
//   - Bloom/glow effects with configurable threshold, intensity, and radius
//   - Screen-space ambient occlusion (SSAO) with multi-sampling
//   - Corner and edge enhancement for improved depth perception
//   - Efficient lighting calculations for real-time rendering
//
// Light sources are defined by their position, color, intensity, and radius.
// The lighting system calculates the combined effect of all light sources on
// each pixel, applying appropriate falloff and blending to create realistic
// lighting effects.
//
// Post-processing effects (bloom and ambient occlusion) can be applied after
// lighting to enhance visual quality. These effects are configurable and can
// be enabled/disabled independently.
//
// Example usage:
//
//	system := lighting.NewSystem()
//
//	// Add ambient light
//	system.AddLight(lighting.Light{
//	    Type:      lighting.TypeAmbient,
//	    Color:     color.RGBA{50, 50, 60, 255},
//	    Intensity: 0.3,
//	})
//
//	// Add point light (torch)
//	system.AddLight(lighting.Light{
//	    Type:      lighting.TypePoint,
//	    Position:  image.Point{X: 100, Y: 100},
//	    Color:     color.RGBA{255, 180, 100, 255},
//	    Intensity: 1.0,
//	    Radius:    80,
//	    Falloff:   lighting.FalloffQuadratic,
//	})
//
//	// Apply lighting to an image
//	litImage := system.ApplyLighting(baseImage)
//
//	// Apply post-processing effects
//	finalImage := system.ApplyFullPostProcessing(litImage, nil)
//
// Bloom/glow effects:
//
//	// Configure bloom
//	config := lighting.DefaultBloomConfig()
//	config.Threshold = 0.8  // Only very bright pixels bloom
//	config.Intensity = 1.5  // Strong bloom effect
//	config.Radius = 12      // Wide bloom spread
//
//	// Apply bloom to already lit image
//	bloomImage := lighting.ApplyBloom(litImage, config)
//
// Ambient occlusion:
//
//	// Configure enhanced AO
//	aoConfig := lighting.DefaultEnhancedAOConfig()
//	aoConfig.Intensity = 0.6        // Medium darkening
//	aoConfig.Radius = 16            // Sample radius
//	aoConfig.Samples = 16           // Quality vs performance
//	aoConfig.CornerIntensity = 0.4  // Extra corner darkening
//	aoConfig.EdgeIntensity = 0.3    // Extra edge darkening
//
//	// Apply AO (can provide depth map or use auto-generated)
//	aoImage := lighting.ApplyEnhancedAO(litImage, nil, aoConfig)
//
// All lighting calculations are deterministic and can be serialized for
// multiplayer synchronization.
//
// Performance considerations:
//   - Bloom: ~20-50ms for 800x600 image with default settings
//   - Ambient occlusion: ~50-150ms for 800x600 image with 16 samples
//   - Both effects scale with image size and quality settings
//   - Effects can be disabled individually for performance tuning
package lighting
