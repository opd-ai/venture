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
	// WeatherSandstorm represents intense sandstorm (post-apocalyptic)
	WeatherSandstorm
	// WeatherBloodRain represents horror-themed blood rain
	WeatherBloodRain
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
	case WeatherSandstorm:
		return "Sandstorm"
	case WeatherBloodRain:
		return "BloodRain"
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

// WeatherEffect represents environmental effects from weather.
type WeatherEffect struct {
	// Puddles tracks rain accumulation at tile positions
	// Map key is "x,y" tile position, value is accumulation level (0.0-1.0)
	Puddles map[string]float64

	// SnowLevel tracks snow accumulation at tile positions
	// Map key is "x,y" tile position, value is snow depth (0.0-1.0)
	SnowLevel map[string]float64

	// VisibilityModifier affects how far entities can see (1.0 = normal, 0.0 = blind)
	// Fog and sandstorms reduce visibility
	VisibilityModifier float64

	// WindDrift tracks current wind drift direction and strength
	WindDriftX float64
	WindDriftY float64
}

// NewWeatherEffect creates a new weather effect tracker.
func NewWeatherEffect() *WeatherEffect {
	return &WeatherEffect{
		Puddles:            make(map[string]float64),
		SnowLevel:          make(map[string]float64),
		VisibilityModifier: 1.0,
		WindDriftX:         0.0,
		WindDriftY:         0.0,
	}
}

// WeatherSystem represents a weather particle system.
type WeatherSystem struct {
	// Configuration
	Config WeatherConfig

	// Active particles
	Particles []Particle

	// Environmental effects
	Effects *WeatherEffect

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
	case WeatherSandstorm:
		generateSandstormParticles(particles, config, rng)
	case WeatherBloodRain:
		generateBloodRainParticles(particles, config, rng)
	default:
		generateRainParticles(particles, config, rng)
	}

	ws := &WeatherSystem{
		Config:    config,
		Particles: particles,
		Effects:   NewWeatherEffect(),
		rng:       rng,
	}

	// Initialize visibility modifier based on weather type
	ws.updateVisibility()

	return ws, nil
}

// Update updates the weather system.
func (ws *WeatherSystem) Update(deltaTime float64) {
	ws.ElapsedTime += deltaTime

	// Update wind drift for snow
	if ws.Config.Type == WeatherSnow {
		ws.Effects.WindDriftX = ws.Config.WindX * 0.5
		ws.Effects.WindDriftY = ws.Config.WindY * 0.5
	}

	for i := range ws.Particles {
		p := &ws.Particles[i]

		// Update position
		p.X += (p.VX + ws.Config.WindX) * deltaTime
		p.Y += (p.VY + ws.Config.WindY) * deltaTime

		// Update rotation
		p.Rotation += p.RotationVel * deltaTime

		// Check for ground impact and accumulation
		if p.Y >= float64(ws.Config.Height) {
			ws.handleParticleImpact(p)
			
			// Wrap particles around screen edges
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

// updateVisibility sets the visibility modifier based on weather type and intensity.
func (ws *WeatherSystem) updateVisibility() {
	baseVisibility := 1.0

	switch ws.Config.Type {
	case WeatherFog:
		// Fog reduces visibility 30-50% based on intensity
		switch ws.Config.Intensity {
		case IntensityLight:
			baseVisibility = 0.85
		case IntensityMedium:
			baseVisibility = 0.65
		case IntensityHeavy:
			baseVisibility = 0.50
		case IntensityExtreme:
			baseVisibility = 0.35
		}
	case WeatherSmog:
		// Smog similar to fog but slightly less severe
		switch ws.Config.Intensity {
		case IntensityLight:
			baseVisibility = 0.90
		case IntensityMedium:
			baseVisibility = 0.75
		case IntensityHeavy:
			baseVisibility = 0.60
		case IntensityExtreme:
			baseVisibility = 0.45
		}
	case WeatherSandstorm:
		// Sandstorm severely reduces visibility
		switch ws.Config.Intensity {
		case IntensityLight:
			baseVisibility = 0.75
		case IntensityMedium:
			baseVisibility = 0.55
		case IntensityHeavy:
			baseVisibility = 0.35
		case IntensityExtreme:
			baseVisibility = 0.20
		}
	case WeatherBloodRain:
		// Blood rain slightly obscures vision
		switch ws.Config.Intensity {
		case IntensityLight:
			baseVisibility = 0.95
		case IntensityMedium:
			baseVisibility = 0.90
		case IntensityHeavy:
			baseVisibility = 0.85
		case IntensityExtreme:
			baseVisibility = 0.75
		}
	}

	ws.Effects.VisibilityModifier = baseVisibility
}

// handleParticleImpact processes particle impacts for accumulation effects.
func (ws *WeatherSystem) handleParticleImpact(p *Particle) {
	// Convert world position to tile position (assuming 32x32 tiles)
	tileSize := 32
	tileX := int(p.X) / tileSize
	tileY := int(p.Y) / tileSize
	tileKey := fmt.Sprintf("%d,%d", tileX, tileY)

	switch ws.Config.Type {
	case WeatherRain, WeatherBloodRain:
		// Accumulate puddles
		current := ws.Effects.Puddles[tileKey]
		// Each impact adds a small amount, capped at 1.0
		ws.Effects.Puddles[tileKey] = math.Min(current+0.001, 1.0)

	case WeatherSnow:
		// Accumulate snow
		current := ws.Effects.SnowLevel[tileKey]
		// Snow accumulates more slowly than rain
		ws.Effects.SnowLevel[tileKey] = math.Min(current+0.0005, 1.0)
	}
}

// GetPuddleLevel returns the puddle accumulation level at a tile position.
func (ws *WeatherSystem) GetPuddleLevel(tileX, tileY int) float64 {
	tileKey := fmt.Sprintf("%d,%d", tileX, tileY)
	return ws.Effects.Puddles[tileKey]
}

// GetSnowLevel returns the snow accumulation level at a tile position.
func (ws *WeatherSystem) GetSnowLevel(tileX, tileY int) float64 {
	tileKey := fmt.Sprintf("%d,%d", tileX, tileY)
	return ws.Effects.SnowLevel[tileKey]
}

// GetVisibilityModifier returns the current visibility modifier (1.0 = normal, 0.0 = blind).
func (ws *WeatherSystem) GetVisibilityModifier() float64 {
	return ws.Effects.VisibilityModifier
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

func generateSandstormParticles(particles []Particle, config WeatherConfig, rng *rand.Rand) {
	// Sandy browns and yellows
	sandColors := []color.RGBA{
		{194, 178, 128, 200}, // Sandy brown
		{210, 180, 140, 200}, // Tan
		{222, 184, 135, 200}, // Burlywood
		{237, 201, 175, 200}, // Light sand
	}

	for i := range particles {
		particles[i] = Particle{
			X:           float64(rng.Intn(config.Width)),
			Y:           float64(rng.Intn(config.Height)),
			VX:          rng.Float64()*80 - 40,  // High horizontal velocity
			VY:          rng.Float64()*30 - 15,  // Some vertical movement
			Color:       sandColors[rng.Intn(len(sandColors))],
			Size:        1 + rng.Float64()*4,
			Life:        1.0,
			InitialLife: 1.5 + rng.Float64()*2.5,
			Rotation:    rng.Float64() * 2 * math.Pi,
			RotationVel: (rng.Float64() - 0.5) * 4,
		}
	}
}

func generateBloodRainParticles(particles []Particle, config WeatherConfig, rng *rand.Rand) {
	// Deep red colors for blood
	bloodColors := []color.RGBA{
		{139, 0, 0, 220},    // Dark red
		{178, 34, 34, 220},  // Firebrick
		{165, 42, 42, 220},  // Brown
		{128, 0, 0, 220},    // Maroon
	}

	for i := range particles {
		particles[i] = Particle{
			X:           float64(rng.Intn(config.Width)),
			Y:           float64(rng.Intn(config.Height)),
			VX:          rng.Float64()*20 - 10,
			VY:          180 + rng.Float64()*80,  // Similar to rain
			Color:       bloodColors[rng.Intn(len(bloodColors))],
			Size:        1 + rng.Float64()*3,     // Slightly larger drops
			Life:        1.0,
			InitialLife: 1.2 + rng.Float64()*2.5,
			Rotation:    0,
			RotationVel: 0,
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
		return []WeatherType{WeatherFog, WeatherBloodRain, WeatherAsh}
	case "cyberpunk":
		return []WeatherType{WeatherNeonRain, WeatherSmog, WeatherFog}
	case "postapoc":
		return []WeatherType{WeatherSandstorm, WeatherAsh, WeatherRadiation}
	default:
		return []WeatherType{WeatherRain, WeatherSnow, WeatherFog}
	}
}
