// Package quality provides visual quality tier management for Venture's rendering system.
//
// This package enables dynamic adjustment of rendering features based on performance
// requirements and hardware capabilities. It supports three quality levels (Low, Medium, High)
// with granular per-feature control and automatic performance-based adjustment.
//
// # Quality Levels
//
// The package defines three quality levels:
//
//   - Low: Optimized for maximum performance (2x FPS improvement target)
//     Disables most post-processing, reduces particles by 75%, simplifies sprites
//
//   - Medium: Balanced quality and performance for most hardware
//     Enables key effects, standard particles, enhanced sprites
//
//   - High: Maximum visual fidelity (60 FPS target on capable hardware)
//     All effects enabled, maximum particles, highest detail
//
// # Basic Usage
//
// Create a quality configuration and apply it:
//
//	// Start with medium quality
//	config := quality.MediumQualityConfig()
//
//	// Or create custom configuration
//	config := quality.Config{
//	    Level: quality.QualityMedium,
//	    EnablePostProcessing: true,
//	    EnableBloom: false, // Disable expensive bloom
//	    ParticleCountMultiplier: 0.6,
//	}
//
//	// Validate configuration
//	if err := config.Validate(); err != nil {
//	    log.Fatal(err)
//	}
//
// # Automatic Quality Adjustment
//
// Use the auto-adjuster for dynamic quality based on performance:
//
//	// Create auto-adjuster targeting 60 FPS
//	config := quality.HighQualityConfig()
//	adjuster := quality.NewAutoAdjuster(&config, 60.0)
//
//	// Set callback for quality changes
//	adjuster.SetOnChange(func(level quality.QualityLevel) {
//	    log.Printf("Quality adjusted to: %s", level)
//	})
//
//	// In game loop, update with frame time
//	frameStart := time.Now()
//	// ... render frame ...
//	frameTimeMS := float64(time.Since(frameStart).Microseconds()) / 1000.0
//	if adjuster.Update(frameTimeMS) {
//	    log.Println("Quality auto-adjusted")
//	}
//
//	// Get current stats
//	stats := adjuster.GetStats()
//	fmt.Printf("FPS: %.1f (min: %.1f, max: %.1f)\n",
//	    stats.AverageFPS, stats.MinFPS, stats.MaxFPS)
//
// # Per-Entity Quality Overrides
//
// Apply custom quality settings to specific entities:
//
//	// Create entity with quality override
//	entity := world.CreateEntity()
//	entity.AddComponent(quality.WithSpriteDetail(0.5)) // Half detail
//
//	// Or disable effects for background entities
//	bgEntity := world.CreateEntity()
//	bgEntity.AddComponent(quality.WithoutEffects())
//
// # Integration with Rendering Systems
//
// The quality configuration controls various rendering features:
//
//   - Post-processing: Bloom, AO, motion blur, color grading, vignette
//   - Lighting: Soft shadows, colored lights, shadow sample count
//   - Sprites: Detail level, anti-aliasing quality, cache settings
//   - Tiles: Texture patterns, transitions, parallax depth, layer count
//   - Particles: Count multiplier, physics, weather, ambience
//   - UI: Decorations, transitions, hierarchy, patterns
//   - Environment: Decoration density, visual variations
//
// # Performance Monitoring
//
// Monitor performance without auto-adjustment:
//
//	monitor := quality.NewPerformanceMonitor(60.0, 120) // 60 FPS target, 120 sample size
//
//	// Record frames
//	monitor.RecordFrame(frameTimeMS)
//
//	// Get recommended quality
//	recommended, shouldChange := monitor.GetRecommendedQuality()
//	if shouldChange {
//	    config.ApplyLevel(recommended)
//	}
//
// # Design Principles
//
//   - Deterministic: Quality settings don't affect procedural generation
//   - Granular: Each feature can be independently toggled
//   - Performance-aware: Auto-adjustment prevents frame rate drops
//   - Flexible: Support both global and per-entity quality settings
//   - Conservative: Quality reductions are quicker than increases
//
// # Performance Targets
//
//   - Low quality: 2x FPS improvement over High
//   - Medium quality: 60 FPS on mid-range hardware
//   - High quality: 60 FPS on capable hardware (baseline: 106 FPS with 2000 entities)
//
// # Example: Platform-Specific Defaults
//
//	func GetPlatformDefaultQuality() quality.QualityLevel {
//	    switch runtime.GOOS {
//	    case "js": // WebAssembly
//	        return quality.QualityMedium
//	    case "android", "ios":
//	        return quality.QualityLow
//	    default:
//	        return quality.QualityHigh
//	    }
//	}
//
// # See Also
//
//   - pkg/rendering/postprocess - Post-processing effects configuration
//   - pkg/rendering/lighting - Lighting system configuration
//   - pkg/rendering/sprites - Sprite generation and caching
//   - pkg/rendering/particles - Particle system configuration
package quality
