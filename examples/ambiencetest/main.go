// ambiencetest is a CLI tool for testing and visualizing ambient particle systems.
//
// Usage:
//
//	ambiencetest [flags]
//
// Flags:
//
//	-environment string
//	    Environment type (dungeon, cave, forest, desert, snow, swamp, lava, city, laboratory, ruins) (default "dungeon")
//	-genre string
//	    Genre ID for color selection (fantasy, scifi, horror, cyberpunk, postapoc) (default "fantasy")
//	-width int
//	    Width of ambient area (default 800)
//	-height int
//	    Height of ambient area (default 600)
//	-seed int
//	    Random seed for deterministic generation (default 0)
//	-density float
//	    Particle density (0.0-1.0, default 0.5)
//	-duration float
//	    Simulation duration in seconds (default 10)
//	-verbose
//	    Enable verbose output
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

func main() {
	// Command-line flags
	environmentType := flag.String("environment", "dungeon", "Environment type (dungeon, cave, forest, desert, snow, swamp, lava, city, laboratory, ruins)")
	genre := flag.String("genre", "fantasy", "Genre ID (fantasy, scifi, horror, cyberpunk, postapoc)")
	width := flag.Int("width", 800, "Width of ambient area")
	height := flag.Int("height", 600, "Height of ambient area")
	seed := flag.Int64("seed", 0, "Random seed for deterministic generation")
	density := flag.Float64("density", 0.5, "Particle density (0.0-1.0)")
	duration := flag.Float64("duration", 10, "Simulation duration in seconds")
	verbose := flag.Bool("verbose", false, "Enable verbose output")

	flag.Parse()

	// Configure logging
	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	// Parse environment type
	envType := parseEnvironmentType(*environmentType)
	if envType == -1 {
		log.Fatalf("Invalid environment type: %s", *environmentType)
	}

	// Validate density
	if *density < 0.0 || *density > 1.0 {
		log.Fatalf("Invalid density: %f (must be 0.0-1.0)", *density)
	}

	// Create ambience configuration
	config := particles.AmbienceConfig{
		Type:    envType,
		Width:   *width,
		Height:  *height,
		GenreID: *genre,
		Seed:    *seed,
		Density: *density,
		Custom:  make(map[string]interface{}),
	}

	// Generate ambient particle system
	logrus.Infof("Generating %s ambience (density %.1f) for %s genre",
		config.Type, config.Density, config.GenreID)
	logrus.Infof("Area: %dx%d, Seed: %d",
		config.Width, config.Height, config.Seed)

	startTime := time.Now()
	ambience, err := particles.GenerateAmbience(config)
	if err != nil {
		log.Fatalf("Failed to generate ambience: %v", err)
	}
	genDuration := time.Since(startTime)

	logrus.Infof("Generated %d particles in %v", len(ambience.Particles), genDuration)

	// Display particle statistics
	displayParticleStats(ambience)

	// Simulate ambient particles
	logrus.Infof("\nSimulating ambience for %.1f seconds...", *duration)

	frames := int(*duration * 60.0) // 60 FPS
	deltaTime := 1.0 / 60.0

	var totalUpdateTime time.Duration
	for i := 0; i < frames; i++ {
		updateStart := time.Now()
		ambience.Update(deltaTime)
		totalUpdateTime += time.Since(updateStart)

		// Log progress every 60 frames (1 second)
		if (i+1)%60 == 0 {
			elapsed := float64(i+1) / 60.0
			logrus.Debugf("%.1fs elapsed, %d particles active",
				elapsed, len(ambience.Particles))
		}
	}

	avgUpdateTime := totalUpdateTime / time.Duration(frames)
	frameTimePercent := (float64(avgUpdateTime.Microseconds()) / 16667.0) * 100.0

	logrus.Infof("\nSimulation complete:")
	logrus.Infof("  Total time: %.1f seconds", *duration)
	logrus.Infof("  Frames: %d", frames)
	logrus.Infof("  Average update time: %v", avgUpdateTime)
	logrus.Infof("  Frame time percentage: %.3f%% (target: <2%%)", frameTimePercent)

	if frameTimePercent > 2.0 {
		logrus.Warnf("WARNING: Frame time exceeds 2%% target")
	} else {
		logrus.Infof("✓ Performance target met!")
	}

	// Display final particle stats
	fmt.Println()
	displayParticleStats(ambience)
}

// parseEnvironmentType parses an environment type string.
func parseEnvironmentType(s string) particles.EnvironmentType {
	switch strings.ToLower(s) {
	case "dungeon":
		return particles.EnvironmentDungeon
	case "cave":
		return particles.EnvironmentCave
	case "forest":
		return particles.EnvironmentForest
	case "desert":
		return particles.EnvironmentDesert
	case "snow":
		return particles.EnvironmentSnow
	case "swamp":
		return particles.EnvironmentSwamp
	case "lava":
		return particles.EnvironmentLava
	case "city":
		return particles.EnvironmentCity
	case "laboratory", "lab":
		return particles.EnvironmentLaboratory
	case "ruins":
		return particles.EnvironmentRuins
	default:
		return -1
	}
}

// displayParticleStats displays statistics about the ambient particles.
func displayParticleStats(ambience *particles.AmbienceSystem) {
	if len(ambience.Particles) == 0 {
		logrus.Info("No particles")
		return
	}

	// Calculate particle statistics
	var totalVelocity, minSize, maxSize, avgSize float64
	minSize = 999999.0
	maxSize = 0.0

	for _, p := range ambience.Particles {
		velocity := p.VX*p.VX + p.VY*p.VY
		totalVelocity += velocity

		if p.Size < minSize {
			minSize = p.Size
		}
		if p.Size > maxSize {
			maxSize = p.Size
		}
		avgSize += p.Size
	}

	count := float64(len(ambience.Particles))
	avgSize /= count
	avgVelocity := totalVelocity / count

	logrus.Infof("Particle Statistics:")
	logrus.Infof("  Count: %d", len(ambience.Particles))
	logrus.Infof("  Size range: %.2f - %.2f pixels (avg: %.2f)",
		minSize, maxSize, avgSize)
	logrus.Infof("  Average velocity magnitude: %.2f px/s", avgVelocity)
	logrus.Infof("  Elapsed time: %.2f seconds", ambience.ElapsedTime)
}
