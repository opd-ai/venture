// Package lighting provides dynamic lighting effects for rendered scenes.package lighting

// This package implements a lighting system with support for multiple light sources,
// color modulation, and realistic light falloff. Light sources can be attached to
// entities or placed in the environment, and their effects are blended together to
// create the final lighting for each pixel.
//
// Key features:
//   - Multiple light source types (point, directional, ambient)
//   - Light intensity and radius control
//   - Color tinting for atmospheric effects
//   - Light falloff with distance (linear, quadratic, or inverse square)
//   - Efficient lighting calculations for real-time rendering
//
// Light sources are defined by their position, color, intensity, and radius.
// The lighting system calculates the combined effect of all light sources on
// each pixel, applying appropriate falloff and blending to create realistic
// lighting effects.
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
// All lighting calculations are deterministic and can be serialized for
// multiplayer synchronization.
package lighting
