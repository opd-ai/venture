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
	config, duration := parseCommandLineFlags()
	ws := generateWeatherSystem(config)
	simulateWeather(ws, duration)
	printResults(ws, config)
}

// parseCommandLineFlags parses and validates all command-line flags, returning a validated weather configuration.
func parseCommandLineFlags() (particles.WeatherConfig, float64) {
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

	configureLogging(*verbose)

	wt := parseWeatherType(*weatherType)
	if wt == -1 {
		log.Fatalf("Invalid weather type: %s", *weatherType)
	}

	intens := parseIntensity(*intensity)
	if intens == -1 {
		log.Fatalf("Invalid intensity: %s", *intensity)
	}

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

	return config, *duration
}

// configureLogging sets the appropriate logging level based on verbosity flag.
func configureLogging(verbose bool) {
	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}

// generateWeatherSystem creates and initializes a weather system from the provided configuration.
func generateWeatherSystem(config particles.WeatherConfig) *particles.WeatherSystem {
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

	return ws
}

// simulateWeather runs the weather simulation for the specified duration at 60 FPS.
func simulateWeather(ws *particles.WeatherSystem, duration float64) {
	deltaTime := 0.016
	steps := int(duration / deltaTime)

	logrus.Infof("Simulating for %.1f seconds (%d steps at 60 FPS)", duration, steps)

	for i := 0; i < steps; i++ {
		ws.Update(deltaTime)
		if i%60 == 0 {
			logrus.Debugf("Step %d: Elapsed time: %.2fs", i, ws.ElapsedTime)
		}
	}
}

// printResults displays comprehensive weather simulation results including environmental effects and particle statistics.
func printResults(ws *particles.WeatherSystem, config particles.WeatherConfig) {
	printBasicResults(ws, config)
	printEnvironmentalEffects(ws, config)
	printParticleStatistics(ws)
	fmt.Println("\n✓ Weather simulation completed successfully")
}

// printBasicResults displays core simulation results.
func printBasicResults(ws *particles.WeatherSystem, config particles.WeatherConfig) {
	fmt.Println("\n=== Weather Simulation Results ===")
	fmt.Printf("Weather Type: %s\n", config.Type)
	fmt.Printf("Intensity: %s\n", config.Intensity)
	fmt.Printf("Genre: %s\n", config.GenreID)
	fmt.Printf("Total Particles: %d\n", len(ws.Particles))
	fmt.Printf("Simulation Duration: %.2f seconds\n", ws.ElapsedTime)
	fmt.Printf("Visibility Modifier: %.2f (1.0 = normal, 0.0 = blind)\n", ws.GetVisibilityModifier())
}

// printEnvironmentalEffects displays weather-specific environmental effects like puddles and snow accumulation.
func printEnvironmentalEffects(ws *particles.WeatherSystem, config particles.WeatherConfig) {
	if config.Type == particles.WeatherRain || config.Type == particles.WeatherBloodRain {
		printPuddleEffects(ws)
	}

	if config.Type == particles.WeatherSnow {
		printSnowEffects(ws)
	}
}

// printPuddleEffects displays puddle formation data for rain-based weather types.
func printPuddleEffects(ws *particles.WeatherSystem) {
	puddleCount := len(ws.Effects.Puddles)
	fmt.Printf("Puddles Formed: %d locations\n", puddleCount)

	if puddleCount > 0 && logrus.GetLevel() == logrus.DebugLevel {
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

// printSnowEffects displays snow accumulation and wind drift data for snow weather.
func printSnowEffects(ws *particles.WeatherSystem) {
	snowCount := len(ws.Effects.SnowLevel)
	fmt.Printf("Snow Accumulated: %d locations\n", snowCount)
	fmt.Printf("Wind Drift: (%.2f, %.2f)\n", ws.Effects.WindDriftX, ws.Effects.WindDriftY)

	if snowCount > 0 && logrus.GetLevel() == logrus.DebugLevel {
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

// printParticleStatistics calculates and displays detailed particle velocity and lifetime statistics.
func printParticleStatistics(ws *particles.WeatherSystem) {
	if logrus.GetLevel() != logrus.DebugLevel {
		return
	}

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
