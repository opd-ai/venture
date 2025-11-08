// Command qualitytest provides a CLI tool for testing and visualizing the quality system.
// This tool demonstrates quality configurations, performance monitoring, and auto-adjustment.
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/opd-ai/venture/pkg/rendering/quality"
)

func main() {
	// Command line flags
	levelStr := flag.String("level", "medium", "Quality level: low, medium, high")
	showConfig := flag.Bool("show-config", false, "Show detailed configuration")
	simulate := flag.Bool("simulate", false, "Simulate performance monitoring")
	targetFPS := flag.Float64("target-fps", 60.0, "Target FPS for simulation")
	duration := flag.Int("duration", 10, "Simulation duration in seconds")
	autoAdjust := flag.Bool("auto-adjust", false, "Enable automatic quality adjustment during simulation")
	flag.Parse()

	fmt.Println("=== Venture Quality System Test Tool ===")
	fmt.Println()

	// Parse quality level
	level := parseQualityLevel(*levelStr)
	fmt.Printf("Quality Level: %s\n", level)
	fmt.Println()

	// Get configuration for level
	var config quality.Config
	switch level {
	case quality.QualityLow:
		config = quality.LowQualityConfig()
	case quality.QualityMedium:
		config = quality.MediumQualityConfig()
	case quality.QualityHigh:
		config = quality.HighQualityConfig()
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		fmt.Printf("ERROR: Invalid configuration: %v\n", err)
		return
	}
	fmt.Println("✓ Configuration is valid")
	fmt.Println()

	// Show configuration if requested
	if *showConfig {
		printConfiguration(config)
		fmt.Println()
	}

	// Run simulation if requested
	if *simulate {
		if *autoAdjust {
			runAutoAdjustSimulation(*targetFPS, *duration)
		} else {
			runPerformanceSimulation(*targetFPS, *duration)
		}
	}

	// Show summary
	printSummary(config)
}

func parseQualityLevel(s string) quality.QualityLevel {
	switch s {
	case "low":
		return quality.QualityLow
	case "medium":
		return quality.QualityMedium
	case "high":
		return quality.QualityHigh
	default:
		fmt.Printf("WARNING: Unknown quality level %q, using medium\n", s)
		return quality.QualityMedium
	}
}

func printConfiguration(config quality.Config) {
	fmt.Println("=== Quality Configuration ===")
	fmt.Println()

	fmt.Println("Post-Processing:")
	fmt.Printf("  Enabled:              %v\n", config.EnablePostProcessing)
	fmt.Printf("  Bloom:                %v\n", config.EnableBloom)
	fmt.Printf("  Ambient Occlusion:    %v\n", config.EnableAmbientOcclusion)
	fmt.Printf("  Motion Blur:          %v\n", config.EnableMotionBlur)
	fmt.Printf("  Depth Blur:           %v\n", config.EnableDepthBlur)
	fmt.Printf("  Color Grading:        %v\n", config.EnableColorGrading)
	fmt.Printf("  Vignette:             %v\n", config.EnableVignette)
	fmt.Printf("  Chromatic Aberration: %v\n", config.EnableChromaticAb)
	fmt.Println()

	fmt.Println("Lighting:")
	fmt.Printf("  Soft Shadows:         %v\n", config.EnableSoftShadows)
	fmt.Printf("  Colored Lighting:     %v\n", config.EnableColoredLighting)
	fmt.Printf("  Dynamic Lighting:     %v\n", config.EnableDynamicLighting)
	fmt.Printf("  Shadow Samples:       %d\n", config.ShadowSampleCount)
	fmt.Println()

	fmt.Println("Sprites:")
	fmt.Printf("  Detail Level:         %.1f\n", config.SpriteDetailLevel)
	fmt.Printf("  Anti-Aliasing:        %v\n", config.EnableAntiAliasing)
	fmt.Printf("  AA Quality:           %d\n", config.AntiAliasingQuality)
	fmt.Printf("  Sprite Cache:         %v\n", config.EnableSpriteCache)
	fmt.Printf("  Equipment Glow:       %v\n", config.EnableEquipmentGlow)
	fmt.Printf("  Damage States:        %v\n", config.EnableDamageStates)
	fmt.Println()

	fmt.Println("Tiles:")
	fmt.Printf("  Texture Patterns:     %v\n", config.EnableTexturePatterns)
	fmt.Printf("  Transitions:          %v\n", config.EnableTileTransitions)
	fmt.Printf("  Parallax Depth:       %v\n", config.EnableParallaxDepth)
	fmt.Printf("  Layer Count:          %d\n", config.TileLayerCount)
	fmt.Printf("  Ambient Occlusion:    %v\n", config.EnableTileAO)
	fmt.Printf("  Normal Maps:          %v\n", config.EnableTileNormals)
	fmt.Println()

	fmt.Println("Particles:")
	fmt.Printf("  Count Multiplier:     %.2f\n", config.ParticleCountMultiplier)
	fmt.Printf("  Physics:              %v\n", config.EnableParticlePhysics)
	fmt.Printf("  Weather Effects:      %v\n", config.EnableWeatherEffects)
	fmt.Printf("  Ambience Particles:   %v\n", config.EnableAmbienceParticles)
	fmt.Printf("  LOD Distance:         %.0f\n", config.ParticleLODDistance)
	fmt.Printf("  Max Particles:        %d\n", config.MaxParticles)
	fmt.Println()

	fmt.Println("UI:")
	fmt.Printf("  Decorations:          %v\n", config.EnableUIDecor)
	fmt.Printf("  Transitions:          %v\n", config.EnableUITransitions)
	fmt.Printf("  Hierarchy:            %v\n", config.EnableUIHierarchy)
	fmt.Printf("  Patterns:             %v\n", config.EnableUIPatterns)
	fmt.Println()

	fmt.Println("Environment:")
	fmt.Printf("  Decorations:          %v\n", config.EnableDecorations)
	fmt.Printf("  Density:              %.1f\n", config.DecorationDensity)
	fmt.Printf("  Visual Variations:    %v\n", config.EnableVisualVariations)
	fmt.Println()

	fmt.Println("Performance:")
	fmt.Printf("  Cache Size:           %d MB\n", config.CacheSizeMB)
	fmt.Printf("  Viewport Culling:     %v\n", config.ViewportCulling)
	fmt.Printf("  Batch Rendering:      %v\n", config.BatchRendering)
	fmt.Printf("  Object Pooling:       %v\n", config.ObjectPooling)
}

func runPerformanceSimulation(targetFPS float64, durationSec int) {
	fmt.Println("=== Performance Monitoring Simulation ===")
	fmt.Printf("Target FPS: %.0f\n", targetFPS)
	fmt.Printf("Duration: %d seconds\n", durationSec)
	fmt.Println()

	monitor := quality.NewPerformanceMonitor(targetFPS, 60)

	// Simulate frames
	totalFrames := int(targetFPS * float64(durationSec))
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	fmt.Println("Simulating frames with varying performance...")

	for i := 0; i < totalFrames; i++ {
		// Simulate variable frame time (base + noise)
		baseFrameTime := 1000.0 / targetFPS
		noise := (rng.Float64() - 0.5) * 5.0 // ±2.5ms noise
		frameTime := baseFrameTime + noise

		// Add occasional spikes
		if i%100 == 0 && rng.Float64() < 0.3 {
			frameTime += rng.Float64() * 10.0 // Add up to 10ms spike
		}

		monitor.RecordFrame(frameTime)

		// Print stats every second
		if i%int(targetFPS) == 0 && i > 0 {
			stats := monitor.GetStats()
			seconds := i / int(targetFPS)
			fmt.Printf("[%2ds] Avg: %5.1f FPS | Min: %5.1f | Max: %5.1f | Quality: %s\n",
				seconds, stats.AverageFPS, stats.MinFPS, stats.MaxFPS, stats.CurrentQuality)

			recommended, shouldChange := monitor.GetRecommendedQuality()
			if shouldChange {
				fmt.Printf("      ⚠ Performance issue detected! Recommend: %s\n", recommended)
			}
		}
	}

	fmt.Println()
	fmt.Println("Final Statistics:")
	stats := monitor.GetStats()
	fmt.Printf("  Average FPS:     %.1f\n", stats.AverageFPS)
	fmt.Printf("  Minimum FPS:     %.1f\n", stats.MinFPS)
	fmt.Printf("  Maximum FPS:     %.1f\n", stats.MaxFPS)
	fmt.Printf("  Current Quality: %s\n", stats.CurrentQuality)
	fmt.Println()
}

func runAutoAdjustSimulation(targetFPS float64, durationSec int) {
	fmt.Println("=== Auto-Adjustment Simulation ===")
	fmt.Printf("Target FPS: %.0f\n", targetFPS)
	fmt.Printf("Duration: %d seconds\n", durationSec)
	fmt.Println()

	config := quality.HighQualityConfig()
	adjuster := quality.NewAutoAdjuster(&config, targetFPS)

	// Set callback to log quality changes
	adjuster.SetOnChange(func(level quality.QualityLevel) {
		fmt.Printf("🔄 QUALITY ADJUSTED TO: %s\n", level)
	})

	// Simulate frames with performance degradation and recovery
	totalFrames := int(targetFPS * float64(durationSec))
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	fmt.Println("Simulating performance degradation and recovery...")
	fmt.Println()

	// Phases:
	// 1. Good performance (0-30%)
	// 2. Degradation (30-50%)
	// 3. Poor performance (50-70%)
	// 4. Recovery (70-100%)

	for i := 0; i < totalFrames; i++ {
		progress := float64(i) / float64(totalFrames)
		var baseFrameTime float64

		if progress < 0.3 {
			// Good performance
			baseFrameTime = 1000.0 / (targetFPS * 1.2) // 20% better than target
		} else if progress < 0.5 {
			// Gradual degradation
			degradationFactor := (progress - 0.3) / 0.2 // 0 to 1
			baseFrameTime = 1000.0 / (targetFPS * (1.2 - degradationFactor*0.6))
		} else if progress < 0.7 {
			// Poor performance
			baseFrameTime = 1000.0 / (targetFPS * 0.6) // 40% worse than target
		} else {
			// Recovery
			recoveryFactor := (progress - 0.7) / 0.3 // 0 to 1
			baseFrameTime = 1000.0 / (targetFPS * (0.6 + recoveryFactor*0.6))
		}

		// Add noise
		noise := (rng.Float64() - 0.5) * 3.0
		frameTime := baseFrameTime + noise

		// Update adjuster
		changed := adjuster.Update(frameTime)
		if changed {
			// Quality was automatically adjusted
			config := adjuster.GetConfig()
			fmt.Printf("    Configuration updated to %s quality settings\n", config.Level)
		}

		// Print stats every second
		if i%int(targetFPS) == 0 && i > 0 {
			stats := adjuster.GetStats()
			seconds := i / int(targetFPS)
			fmt.Printf("[%2ds] Avg: %5.1f FPS | Phase: %s | Quality: %s\n",
				seconds, stats.AverageFPS, getPhase(progress), stats.CurrentQuality)
		}
	}

	fmt.Println()
	fmt.Println("Final Statistics:")
	stats := adjuster.GetStats()
	finalConfig := adjuster.GetConfig()
	fmt.Printf("  Average FPS:     %.1f\n", stats.AverageFPS)
	fmt.Printf("  Minimum FPS:     %.1f\n", stats.MinFPS)
	fmt.Printf("  Maximum FPS:     %.1f\n", stats.MaxFPS)
	fmt.Printf("  Final Quality:   %s\n", finalConfig.Level)
	fmt.Println()
}

func getPhase(progress float64) string {
	if progress < 0.3 {
		return "Good      "
	} else if progress < 0.5 {
		return "Degrading "
	} else if progress < 0.7 {
		return "Poor      "
	} else {
		return "Recovering"
	}
}

func printSummary(config quality.Config) {
	fmt.Println("=== Feature Summary ===")
	fmt.Println()

	enabledFeatures := 0
	totalFeatures := 0

	// Count enabled features
	features := []struct {
		name    string
		enabled bool
	}{
		{"Post-processing", config.EnablePostProcessing},
		{"Bloom", config.EnableBloom},
		{"Ambient Occlusion", config.EnableAmbientOcclusion},
		{"Soft Shadows", config.EnableSoftShadows},
		{"Anti-Aliasing", config.EnableAntiAliasing},
		{"Texture Patterns", config.EnableTexturePatterns},
		{"Tile Transitions", config.EnableTileTransitions},
		{"Parallax Depth", config.EnableParallaxDepth},
		{"Particle Physics", config.EnableParticlePhysics},
		{"Weather Effects", config.EnableWeatherEffects},
		{"UI Decorations", config.EnableUIDecor},
		{"Visual Variations", config.EnableVisualVariations},
	}

	for _, f := range features {
		totalFeatures++
		if f.enabled {
			enabledFeatures++
		}
	}

	fmt.Printf("Enabled Features: %d / %d (%.0f%%)\n",
		enabledFeatures, totalFeatures, float64(enabledFeatures)/float64(totalFeatures)*100)
	fmt.Println()

	// Performance estimates
	fmt.Println("Performance Impact Estimates:")
	fmt.Printf("  Sprite Detail:        %.0f%%\n", config.SpriteDetailLevel*100)
	fmt.Printf("  Particle Count:       %.0f%%\n", config.ParticleCountMultiplier*100)
	fmt.Printf("  Decoration Density:   %.0f%%\n", config.DecorationDensity*100)
	fmt.Printf("  Shadow Quality:       %dx samples\n", config.ShadowSampleCount)
	fmt.Println()

	// Memory estimates
	estimatedMemoryMB := estimateMemoryUsage(config)
	fmt.Printf("Estimated Memory Usage: ~%d MB\n", estimatedMemoryMB)
	fmt.Println()

	// Expected FPS multiplier
	multiplier := estimateFPSMultiplier(config.Level)
	fmt.Printf("Expected FPS Multiplier: %.1fx (vs High quality)\n", multiplier)
	fmt.Println()
}

func estimateMemoryUsage(config quality.Config) int {
	baseMB := 73 // Current baseline

	// Cache
	cacheMB := config.CacheSizeMB

	// Particles (rough estimate: 100 bytes per particle)
	particleMB := (config.MaxParticles * 100) / (1024 * 1024)

	// Decorations and textures
	decorMB := 10
	if config.EnableTexturePatterns {
		decorMB += 20
	}
	if config.EnableParallaxDepth {
		decorMB += 30
	}

	return baseMB + cacheMB + particleMB + decorMB
}

func estimateFPSMultiplier(level quality.QualityLevel) float64 {
	switch level {
	case quality.QualityLow:
		return 2.0 // 2x faster
	case quality.QualityMedium:
		return 1.3 // 30% faster
	case quality.QualityHigh:
		return 1.0 // Baseline
	default:
		return 1.0
	}
}
