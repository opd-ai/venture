// Package particles provides weather particle effects.
package particles

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
)

// WeatherType represents different types of weather effects.
type WeatherType int

const (
	// WeatherRain represents falling rain droplets
	WeatherRain WeatherType = iota
	// WeatherSnow represents falling snowflakes
	WeatherSnow
	// WeatherFog represents ambient fog particles
	WeatherFog
	// WeatherDust represents swirling dust particles
	WeatherDust
	// WeatherAsh represents falling ash particles
	WeatherAsh
	// WeatherNeonRain represents cyberpunk-style neon rain
	WeatherNeonRain
	// WeatherSmog represents thick industrial smog
	WeatherSmog
	// WeatherRadiation represents radioactive particles (post-apocalyptic)
	WeatherRadiation
)

// String returns the string representation of a weather type.
func (w WeatherType) String() string {
	switch w {
	case WeatherRain:
		return "Rain"
	case WeatherSnow:
		return "Snow"
	case WeatherFog:
		return "Fog"
	case WeatherDust:
		return "Dust"
	case WeatherAsh:
		return "Ash"
	case WeatherNeonRain:
		return "NeonRain"
	case WeatherSmog:
		return "Smog"
	case WeatherRadiation:
		return "Radiation"
	default:
		return "Unknown"
	}
}

// WeatherIntensity represents the strength of weather effects.
type WeatherIntensity int

const (
	// IntensityLight represents mild weather
	IntensityLight WeatherIntensity = iota
	// IntensityMedium represents moderate weather
	IntensityMedium
	// IntensityHeavy represents strong weather
	IntensityHeavy
	// IntensityExtreme represents extreme weather conditions
	IntensityExtreme
)

// String returns the string representation of weather intensity.
func (i WeatherIntensity) String() string {
	switch i {
	case IntensityLight:
		return "Light"
	case IntensityMedium:
		return "Medium"
	case IntensityHeavy:
		return "Heavy"
	case IntensityExtreme:
		return "Extreme"
	default:
		return "Unknown"
	}
}

// WeatherConfig contains parameters for weather particle generation.
type WeatherConfig struct {
	// Type of weather effect
	Type WeatherType

	// Intensity of the weather
	Intensity WeatherIntensity

	// Width and Height of the weather area
	Width  int
	Height int

	// GenreID for color selection
	GenreID string

	// Seed for deterministic generation
	Seed int64

	// Wind velocity (affects particle drift)
	WindX float64
	WindY float64

	// Custom parameters
	Custom map[string]interface{}
}

// DefaultWeatherConfig returns a default weather configuration.
func DefaultWeatherConfig() WeatherConfig {
	return WeatherConfig{
		Type:      WeatherRain,
		Intensity: IntensityMedium,
		Width:     800,
		Height:    600,
		GenreID:   "fantasy",
		Seed:      0,
		WindX:     0.0,
		WindY:     0.0,
		Custom:    make(map[string]interface{}),
	}
}

// Validate checks if the weather configuration is valid.
func (c WeatherConfig) Validate() error {
	if c.Width <= 0 {
		return fmt.Errorf("width must be positive, got %d", c.Width)
	}
	if c.Height <= 0 {
		return fmt.Errorf("height must be positive, got %d", c.Height)
	}
	if c.GenreID == "" {
		return fmt.Errorf("genreID cannot be empty")
	}
	return nil
}

// GetParticleCount returns the number of particles for the weather intensity.
func (c WeatherConfig) GetParticleCount() int {
	// Base density per 1000 square pixels
	baseCount := float64(c.Width*c.Height) / 1000.0

	switch c.Intensity {
	case IntensityLight:
		return int(baseCount * 2.0)
	case IntensityMedium:
		return int(baseCount * 5.0)
	case IntensityHeavy:
		return int(baseCount * 10.0)
	case IntensityExtreme:
		return int(baseCount * 20.0)
	default:
		return int(baseCount * 5.0)
	}
}

// WeatherSystem represents a weather particle system.
// Future feature: This system is designed for weather effects (rain, snow, etc.) but not yet integrated.
// Planned integration per roadmap category 5.4 for dynamic weather and environmental effects.
type WeatherSystem struct {
	// Configuration
	Config WeatherConfig

	// Active particles
	Particles []Particle

	// Elapsed time
	ElapsedTime float64

	// Random number generator
	rng *rand.Rand
}

// GenerateWeather creates a new weather particle system.
func GenerateWeather(config WeatherConfig) (*WeatherSystem, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	rng := rand.New(rand.NewSource(config.Seed))
	particleCount := config.GetParticleCount()

	// Cap at 10000 particles for performance
	if particleCount > 10000 {
		particleCount = 10000
	}

	particles := make([]Particle, particleCount)

	// Generate particles based on weather type
	switch config.Type {
	case WeatherRain:
		generateRainParticles(particles, config, rng)
	case WeatherSnow:
		generateSnowParticles(particles, config, rng)
	case WeatherFog:
		generateFogParticles(particles, config, rng)
	case WeatherDust:
		generateDustParticles(particles, config, rng)
	case WeatherAsh:
		generateAshParticles(particles, config, rng)
	case WeatherNeonRain:
		generateNeonRainParticles(particles, config, rng)
	case WeatherSmog:
		generateSmogParticles(particles, config, rng)
	case WeatherRadiation:
		generateRadiationParticles(particles, config, rng)
	default:
		generateRainParticles(particles, config, rng)
	}

	return &WeatherSystem{
		Config:    config,
		Particles: particles,
		rng:       rng,
	}, nil
}

// Update updates the weather system.
func (ws *WeatherSystem) Update(deltaTime float64) {
	ws.ElapsedTime += deltaTime

	for i := range ws.Particles {
		p := &ws.Particles[i]

		// Update position
		p.X += (p.VX + ws.Config.WindX) * deltaTime
		p.Y += (p.VY + ws.Config.WindY) * deltaTime

		// Update rotation
		p.Rotation += p.RotationVel * deltaTime

		// Wrap particles around screen edges
		if p.Y > float64(ws.Config.Height) {
			p.Y = 0
			p.X = float64(ws.rng.Intn(ws.Config.Width))
		}
		if p.X < 0 {
			p.X = float64(ws.Config.Width)
		}
		if p.X > float64(ws.Config.Width) {
			p.X = 0
		}

		// Update life (for fading effects)
		if p.InitialLife > 0 {
			p.Life -= deltaTime / p.InitialLife
			if p.Life <= 0 {
				// Respawn particle
				p.Life = 1.0
				p.Y = 0
				p.X = float64(ws.rng.Intn(ws.Config.Width))
			}
		}
	}
}

// Helper functions for generating different weather types

func generateRainParticles(particles []Particle, config WeatherConfig, rng *rand.Rand) {
	rainColor := color.RGBA{150, 180, 255, 200}

	for i := range particles {
		particles[i] = Particle{
			X:           float64(rng.Intn(config.Width)),
			Y:           float64(rng.Intn(config.Height)),
			VX:          rng.Float64()*20 - 10,
			VY:          200 + rng.Float64()*100,
			Color:       rainColor,
			Size:        1 + rng.Float64()*2,
			Life:        1.0,
			InitialLife: 1.0 + rng.Float64()*2,
			Rotation:    0,
			RotationVel: 0,
		}
	}
}

func generateSnowParticles(particles []Particle, config WeatherConfig, rng *rand.Rand) {
	snowColor := color.RGBA{255, 255, 255, 220}

	for i := range particles {
		particles[i] = Particle{
			X:           float64(rng.Intn(config.Width)),
			Y:           float64(rng.Intn(config.Height)),
			VX:          rng.Float64()*20 - 10,
			VY:          20 + rng.Float64()*30,
			Color:       snowColor,
			Size:        2 + rng.Float64()*4,
			Life:        1.0,
			InitialLife: 2.0 + rng.Float64()*3,
			Rotation:    rng.Float64() * 2 * math.Pi,
			RotationVel: (rng.Float64() - 0.5) * 2,
		}
	}
}

func generateFogParticles(particles []Particle, config WeatherConfig, rng *rand.Rand) {
	fogColor := color.RGBA{200, 200, 210, 100}

	for i := range particles {
		particles[i] = Particle{
			X:           float64(rng.Intn(config.Width)),
			Y:           float64(rng.Intn(config.Height)),
			VX:          rng.Float64()*10 - 5,
			VY:          rng.Float64()*5 - 2.5,
			Color:       fogColor,
			Size:        20 + rng.Float64()*40,
			Life:        1.0,
			InitialLife: 10.0 + rng.Float64()*10,
			Rotation:    0,
			RotationVel: 0,
		}
	}
}

func generateDustParticles(particles []Particle, config WeatherConfig, rng *rand.Rand) {
	dustColor := color.RGBA{180, 150, 120, 150}

	for i := range particles {
		particles[i] = Particle{
			X:           float64(rng.Intn(config.Width)),
			Y:           float64(rng.Intn(config.Height)),
			VX:          rng.Float64()*40 - 20,
			VY:          rng.Float64()*20 - 10,
			Color:       dustColor,
			Size:        1 + rng.Float64()*3,
			Life:        1.0,
			InitialLife: 2.0 + rng.Float64()*3,
			Rotation:    0,
			RotationVel: 0,
		}
	}
}

func generateAshParticles(particles []Particle, config WeatherConfig, rng *rand.Rand) {
	ashColor := color.RGBA{120, 120, 120, 180}

	for i := range particles {
		particles[i] = Particle{
			X:           float64(rng.Intn(config.Width)),
			Y:           float64(rng.Intn(config.Height)),
			VX:          rng.Float64()*10 - 5,
			VY:          15 + rng.Float64()*20,
			Color:       ashColor,
			Size:        1 + rng.Float64()*2,
			Life:        1.0,
			InitialLife: 3.0 + rng.Float64()*5,
			Rotation:    rng.Float64() * 2 * math.Pi,
			RotationVel: (rng.Float64() - 0.5),
		}
	}
}

func generateNeonRainParticles(particles []Particle, config WeatherConfig, rng *rand.Rand) {
	// Cyberpunk neon colors
	colors := []color.RGBA{
		{255, 0, 255, 220}, // Magenta
		{0, 255, 255, 220}, // Cyan
		{255, 0, 100, 220}, // Hot pink
		{0, 255, 0, 220},   // Green
	}

	for i := range particles {
		particles[i] = Particle{
			X:           float64(rng.Intn(config.Width)),
			Y:           float64(rng.Intn(config.Height)),
			VX:          rng.Float64()*30 - 15,
			VY:          250 + rng.Float64()*150,
			Color:       colors[rng.Intn(len(colors))],
			Size:        1 + rng.Float64()*3,
			Life:        1.0,
			InitialLife: 0.8 + rng.Float64()*1.2,
			Rotation:    0,
			RotationVel: 0,
		}
	}
}

func generateSmogParticles(particles []Particle, config WeatherConfig, rng *rand.Rand) {
	// Industrial smog colors
	smogColor := color.RGBA{100, 100, 80, 120}

	for i := range particles {
		particles[i] = Particle{
			X:           float64(rng.Intn(config.Width)),
			Y:           float64(rng.Intn(config.Height)),
			VX:          rng.Float64()*15 - 7.5,
			VY:          rng.Float64()*8 - 4,
			Color:       smogColor,
			Size:        15 + rng.Float64()*30,
			Life:        1.0,
			InitialLife: 8.0 + rng.Float64()*12,
			Rotation:    0,
			RotationVel: 0,
		}
	}
}

func generateRadiationParticles(particles []Particle, config WeatherConfig, rng *rand.Rand) {
	// Radioactive glow colors
	colors := []color.RGBA{
		{0, 255, 0, 150},   // Green
		{255, 255, 0, 150}, // Yellow
		{0, 255, 100, 150}, // Greenish
	}

	for i := range particles {
		particles[i] = Particle{
			X:           float64(rng.Intn(config.Width)),
			Y:           float64(rng.Intn(config.Height)),
			VX:          rng.Float64()*10 - 5,
			VY:          rng.Float64()*10 - 5,
			Color:       colors[rng.Intn(len(colors))],
			Size:        2 + rng.Float64()*5,
			Life:        1.0,
			InitialLife: 2.0 + rng.Float64()*4,
			Rotation:    0,
			RotationVel: (rng.Float64() - 0.5) * 3,
		}
	}
}

// GetGenreWeather returns appropriate weather types for a genre.
func GetGenreWeather(genreID string) []WeatherType {
	switch genreID {
	case "fantasy":
		return []WeatherType{WeatherRain, WeatherSnow, WeatherFog}
	case "scifi":
		return []WeatherType{WeatherRain, WeatherDust, WeatherFog}
	case "horror":
		return []WeatherType{WeatherFog, WeatherRain, WeatherAsh}
	case "cyberpunk":
		return []WeatherType{WeatherNeonRain, WeatherSmog, WeatherFog}
	case "postapoc":
		return []WeatherType{WeatherDust, WeatherAsh, WeatherRadiation}
	default:
		return []WeatherType{WeatherRain, WeatherSnow, WeatherFog}
	}
}
