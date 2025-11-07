// weathertest is a CLI tool for testing and visualizing weather particle systems.
//
// Usage:
//
//	weathertest [flags]
//
// Flags:
//
//	-weather string
//	    Weather type (rain, snow, fog, dust, ash, neonrain, smog, radiation, sandstorm, bloodrain) (default "rain")
//	-intensity string
//	    Weather intensity (light, medium, heavy, extreme) (default "medium")
//	-genre string
//	    Genre ID for color selection (fantasy, scifi, horror, cyberpunk, postapoc) (default "fantasy")
//	-width int
//	    Width of weather area (default 800)
//	-height int
//	    Height of weather area (default 600)
//	-seed int
//	    Random seed for deterministic generation (default 0)
//	-windx float
//	    Wind velocity X (default 0)
//	-windy float
//	    Wind velocity Y (default 0)
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

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

func main() {
	// Command-line flags
	weatherType := flag.String("weather", "rain", "Weather type (rain, snow, fog, dust, ash, neonrain, smog, radiation, sandstorm, bloodrain)")
	intensity := flag.String("intensity", "medium", "Weather intensity (light, medium, heavy, extreme)")
	genre := flag.String("genre", "fantasy", "Genre ID (fantasy, scifi, horror, cyberpunk, postapoc)")
	width := flag.Int("width", 800, "Width of weather area")
	height := flag.Int("height", 600, "Height of weather area")
	seed := flag.Int64("seed", 0, "Random seed for deterministic generation")
	windX := flag.Float64("windx", 0, "Wind velocity X")
	windY := flag.Float64("windy", 0, "Wind velocity Y")
	duration := flag.Float64("duration", 10, "Simulation duration in seconds")
	verbose := flag.Bool("verbose", false, "Enable verbose output")

	flag.Parse()

	// Configure logging
	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	// Parse weather type
	wt := parseWeatherType(*weatherType)
	if wt == -1 {
		log.Fatalf("Invalid weather type: %s", *weatherType)
	}

	// Parse intensity
	intens := parseIntensity(*intensity)
	if intens == -1 {
		log.Fatalf("Invalid intensity: %s", *intensity)
	}

	// Create weather configuration
	config := particles.WeatherConfig{
		Type:      wt,
		Intensity: intens,
		Width:     *width,
		Height:    *height,
		GenreID:   *genre,
		Seed:      *seed,
		WindX:     *windX,
		WindY:     *windY,
		Custom:    make(map[string]interface{}),
	}

	// Generate weather system
	logrus.Infof("Generating %s weather (%s intensity) for %s genre",
		config.Type, config.Intensity, config.GenreID)
	logrus.Infof("Area: %dx%d, Seed: %d, Wind: (%.1f, %.1f)",
		config.Width, config.Height, config.Seed, config.WindX, config.WindY)

	ws, err := particles.GenerateWeather(config)
	if err != nil {
		log.Fatalf("Failed to generate weather: %v", err)
	}

	logrus.Infof("Generated %d particles", len(ws.Particles))
	logrus.Infof("Visibility modifier: %.2f (1.0 = normal, 0.0 = blind)", ws.GetVisibilityModifier())

	// Simulate weather for specified duration
	deltaTime := 0.016 // ~60 FPS
	steps := int(*duration / deltaTime)

	logrus.Infof("Simulating for %.1f seconds (%d steps at 60 FPS)", *duration, steps)

	for i := 0; i < steps; i++ {
		ws.Update(deltaTime)

		// Log progress every second
		if i%(60) == 0 && *verbose {
			logrus.Debugf("Step %d: Elapsed time: %.2fs", i, ws.ElapsedTime)
		}
	}

	// Report results
	fmt.Println("\n=== Weather Simulation Results ===")
	fmt.Printf("Weather Type: %s\n", config.Type)
	fmt.Printf("Intensity: %s\n", config.Intensity)
	fmt.Printf("Genre: %s\n", config.GenreID)
	fmt.Printf("Total Particles: %d\n", len(ws.Particles))
	fmt.Printf("Simulation Duration: %.2f seconds\n", ws.ElapsedTime)
	fmt.Printf("Visibility Modifier: %.2f (1.0 = normal, 0.0 = blind)\n", ws.GetVisibilityModifier())

	// Report environmental effects
	if config.Type == particles.WeatherRain || config.Type == particles.WeatherBloodRain {
		puddleCount := len(ws.Effects.Puddles)
		fmt.Printf("Puddles Formed: %d locations\n", puddleCount)

		if puddleCount > 0 && *verbose {
			// Show some sample puddle levels
			count := 0
			fmt.Println("\nSample Puddle Levels:")
			for loc, level := range ws.Effects.Puddles {
				if count >= 5 {
					break
				}
				fmt.Printf("  %s: %.3f\n", loc, level)
				count++
			}
		}
	}

	if config.Type == particles.WeatherSnow {
		snowCount := len(ws.Effects.SnowLevel)
		fmt.Printf("Snow Accumulated: %d locations\n", snowCount)
		fmt.Printf("Wind Drift: (%.2f, %.2f)\n", ws.Effects.WindDriftX, ws.Effects.WindDriftY)

		if snowCount > 0 && *verbose {
			// Show some sample snow levels
			count := 0
			fmt.Println("\nSample Snow Levels:")
			for loc, level := range ws.Effects.SnowLevel {
				if count >= 5 {
					break
				}
				fmt.Printf("  %s: %.3f\n", loc, level)
				count++
			}
		}
	}

	// Show particle statistics if verbose
	if *verbose {
		fmt.Println("\n=== Particle Statistics ===")
		var totalVelocity float64
		var minLife, maxLife float64 = 1.0, 0.0

		for _, p := range ws.Particles {
			vel := p.VX*p.VX + p.VY*p.VY
			totalVelocity += vel

			if p.Life < minLife {
				minLife = p.Life
			}
			if p.Life > maxLife {
				maxLife = p.Life
			}
		}

		avgVelocity := totalVelocity / float64(len(ws.Particles))
		fmt.Printf("Average Velocity: %.2f\n", avgVelocity)
		fmt.Printf("Life Range: [%.3f, %.3f]\n", minLife, maxLife)
	}

	fmt.Println("\n✓ Weather simulation completed successfully")
}

// parseWeatherType converts string to WeatherType.
func parseWeatherType(s string) particles.WeatherType {
	switch strings.ToLower(s) {
	case "rain":
		return particles.WeatherRain
	case "snow":
		return particles.WeatherSnow
	case "fog":
		return particles.WeatherFog
	case "dust":
		return particles.WeatherDust
	case "ash":
		return particles.WeatherAsh
	case "neonrain":
		return particles.WeatherNeonRain
	case "smog":
		return particles.WeatherSmog
	case "radiation":
		return particles.WeatherRadiation
	case "sandstorm":
		return particles.WeatherSandstorm
	case "bloodrain":
		return particles.WeatherBloodRain
	default:
		return -1
	}
}

// parseIntensity converts string to WeatherIntensity.
func parseIntensity(s string) particles.WeatherIntensity {
	switch strings.ToLower(s) {
	case "light":
		return particles.IntensityLight
	case "medium":
		return particles.IntensityMedium
	case "heavy":
		return particles.IntensityHeavy
	case "extreme":
		return particles.IntensityExtreme
	default:
		return -1
	}
}
