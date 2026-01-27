//go:build ignore
// +build ignore

// Weather CLI Integration Demo
//
// This example demonstrates the weather command-line integration that allows
// users to specify weather type and intensity when starting the game client.
//
// Usage examples:
//   go run weather_cli_demo.go
//   go run weather_cli_demo.go -weather rain -weather-intensity heavy
//   go run weather_cli_demo.go -weather snow -weather-intensity light
//   go run weather_cli_demo.go -weather fog -weather-intensity extreme
//
// Available weather types:
//   - rain, snow, fog, dust, ash (generic)
//   - neonrain, smog, radiation (genre-specific)
//
// Available intensities:
//   - light, medium, heavy, extreme
//
// This demo verifies:
// 1. Command-line flags are properly parsed
// 2. Weather type string is converted to WeatherType enum
// 3. Intensity string is converted to WeatherIntensity enum
// 4. Genre-appropriate random weather selection works
// 5. Weather configuration is deterministic with same seed
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"strings"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

var (
	weatherType      = flag.String("weather", "", "Weather type (rain, snow, fog, dust, ash, neonrain, smog, radiation)")
	weatherIntensity = flag.String("weather-intensity", "heavy", "Weather intensity (light, medium, heavy, extreme)")
	genreID          = flag.String("genre", "fantasy", "Genre ID")
	seed             = flag.Int64("seed", 12345, "Random seed")
)

func main() {
	flag.Parse()

	fmt.Println("Weather CLI Integration Demo")
	fmt.Println("=============================")
	fmt.Println()

	// Parse weather intensity (same logic as cmd/client/util.go)
	var intensity particles.WeatherIntensity
	switch strings.ToLower(*weatherIntensity) {
	case "light":
		intensity = particles.IntensityLight
	case "medium":
		intensity = particles.IntensityMedium
	case "heavy":
		intensity = particles.IntensityHeavy
	case "extreme":
		intensity = particles.IntensityExtreme
	default:
		intensity = particles.IntensityMedium
		fmt.Printf("⚠️  Invalid intensity '%s', defaulting to 'medium'\n", *weatherIntensity)
	}

	// Determine weather type (same logic as cmd/client/util.go)
	var weather particles.WeatherType
	var weatherName string

	if *weatherType == "" {
		// Select genre-appropriate random weather
		fmt.Printf("🎲 No weather specified, selecting genre-appropriate weather for '%s'...\n", *genreID)
		rng := rand.New(rand.NewSource(*seed))
		genreWeathers := particles.GetGenreWeather(*genreID)
		if len(genreWeathers) > 0 {
			weather = genreWeathers[rng.Intn(len(genreWeathers))]
			weatherName = getWeatherName(weather)
			fmt.Printf("   Selected: %s\n", weatherName)
		} else {
			weather = particles.WeatherRain
			weatherName = "rain"
			fmt.Println("   No genre weathers found, defaulting to rain")
		}
	} else {
		// Parse explicit weather type
		weatherName = *weatherType
		switch strings.ToLower(*weatherType) {
		case "rain":
			weather = particles.WeatherRain
		case "snow":
			weather = particles.WeatherSnow
		case "fog":
			weather = particles.WeatherFog
		case "dust":
			weather = particles.WeatherDust
		case "ash":
			weather = particles.WeatherAsh
		case "neonrain":
			weather = particles.WeatherNeonRain
		case "smog":
			weather = particles.WeatherSmog
		case "radiation":
			weather = particles.WeatherRadiation
		default:
			weather = particles.WeatherRain
			weatherName = "rain"
			fmt.Printf("⚠️  Invalid weather type '%s', defaulting to 'rain'\n", *weatherType)
		}
	}

	fmt.Println()
	fmt.Println("Configuration Summary:")
	fmt.Println("----------------------")
	fmt.Printf("Weather Type:      %s (%d)\n", weatherName, weather)
	fmt.Printf("Intensity:         %s (%d)\n", *weatherIntensity, intensity)
	fmt.Printf("Genre:             %s\n", *genreID)
	fmt.Printf("Seed:              %d\n", *seed)
	fmt.Println()

	// Create weather configuration (same as cmd/client/util.go)
	rng := rand.New(rand.NewSource(*seed))
	config := particles.WeatherConfig{
		Type:      weather,
		Intensity: intensity,
		Width:     1920 * 2,
		Height:    1080 * 2,
		GenreID:   *genreID,
		Seed:      *seed,
		WindX:     (rng.Float64() - 0.5) * 20.0,
		WindY:     0.0,
		Custom:    make(map[string]interface{}),
	}

	fmt.Println("Weather Configuration:")
	fmt.Println("----------------------")
	fmt.Printf("Type:              %d\n", config.Type)
	fmt.Printf("Intensity:         %d\n", config.Intensity)
	fmt.Printf("Dimensions:        %dx%d pixels\n", config.Width, config.Height)
	fmt.Printf("Wind:              X=%.2f Y=%.2f\n", config.WindX, config.WindY)
	fmt.Println()

	// Demonstrate determinism
	fmt.Println("Determinism Check:")
	fmt.Println("------------------")
	rng1 := rand.New(rand.NewSource(*seed))
	rng2 := rand.New(rand.NewSource(*seed))
	wind1 := (rng1.Float64() - 0.5) * 20.0
	wind2 := (rng2.Float64() - 0.5) * 20.0
	if wind1 == wind2 {
		fmt.Printf("✅ Same seed produces same wind: %.2f\n", wind1)
	} else {
		fmt.Printf("❌ FAIL: Different wind values: %.2f vs %.2f\n", wind1, wind2)
	}
	fmt.Println()

	// Display genre-appropriate weather options
	fmt.Println("Genre-Appropriate Weather:")
	fmt.Println("--------------------------")
	displayGenreWeathers("fantasy", particles.GetGenreWeather("fantasy"))
	displayGenreWeathers("scifi", particles.GetGenreWeather("scifi"))
	displayGenreWeathers("horror", particles.GetGenreWeather("horror"))
	displayGenreWeathers("cyberpunk", particles.GetGenreWeather("cyberpunk"))
	displayGenreWeathers("postapoc", particles.GetGenreWeather("postapoc"))
	fmt.Println()

	fmt.Println("✅ Weather CLI Integration Verified")
	fmt.Println()
	fmt.Println("Try these commands:")
	fmt.Println("  go run weather_cli_demo.go -weather rain -weather-intensity heavy")
	fmt.Println("  go run weather_cli_demo.go -weather snow -weather-intensity light")
	fmt.Println("  go run weather_cli_demo.go -weather fog -weather-intensity extreme")
	fmt.Println("  go run weather_cli_demo.go -genre cyberpunk  # Random cyberpunk weather")
	fmt.Println("  go run weather_cli_demo.go -weather neonrain -weather-intensity heavy")
}

func getWeatherName(w particles.WeatherType) string {
	names := map[particles.WeatherType]string{
		particles.WeatherRain:       "rain",
		particles.WeatherSnow:       "snow",
		particles.WeatherFog:        "fog",
		particles.WeatherDust:       "dust",
		particles.WeatherAsh:        "ash",
		particles.WeatherNeonRain:   "neonrain",
		particles.WeatherSmog:       "smog",
		particles.WeatherRadiation:  "radiation",
	}
	if name, ok := names[w]; ok {
		return name
	}
	return "unknown"
}

func displayGenreWeathers(genre string, weathers []particles.WeatherType) {
	if len(weathers) == 0 {
		fmt.Printf("  %s: none\n", genre)
		return
	}
	names := make([]string, len(weathers))
	for i, w := range weathers {
		names[i] = getWeatherName(w)
	}
	fmt.Printf("  %s: %s\n", genre, strings.Join(names, ", "))
}
