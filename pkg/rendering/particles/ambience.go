// Package particles provides ambient particle effects.
package particles

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
)

// EnvironmentType represents different types of environments with ambient effects.
type EnvironmentType int

const (
	// EnvironmentDungeon represents a dungeon environment with dust motes and moisture
	EnvironmentDungeon EnvironmentType = iota
	// EnvironmentCave represents a cave environment with dripping water and mineral dust
	EnvironmentCave
	// EnvironmentForest represents a forest environment with falling leaves and pollen
	EnvironmentForest
	// EnvironmentDesert represents a desert environment with sand and heat shimmer
	EnvironmentDesert
	// EnvironmentSnow represents a snowy environment with drifting snow and ice crystals
	EnvironmentSnow
	// EnvironmentSwamp represents a swamp environment with mist and fireflies
	EnvironmentSwamp
	// EnvironmentLava represents a volcanic environment with ash and embers
	EnvironmentLava
	// EnvironmentCity represents an urban environment with paper debris and smoke
	EnvironmentCity
	// EnvironmentLaboratory represents a lab environment with energy particles and sparks
	EnvironmentLaboratory
	// EnvironmentRuins represents ruined structures with falling dust and debris
	EnvironmentRuins
)

// String returns the string representation of an environment type.
func (e EnvironmentType) String() string {
	switch e {
	case EnvironmentDungeon:
		return "Dungeon"
	case EnvironmentCave:
		return "Cave"
	case EnvironmentForest:
		return "Forest"
	case EnvironmentDesert:
		return "Desert"
	case EnvironmentSnow:
		return "Snow"
	case EnvironmentSwamp:
		return "Swamp"
	case EnvironmentLava:
		return "Lava"
	case EnvironmentCity:
		return "City"
	case EnvironmentLaboratory:
		return "Laboratory"
	case EnvironmentRuins:
		return "Ruins"
	default:
		return "Unknown"
	}
}

// AmbienceConfig contains parameters for ambient particle generation.
type AmbienceConfig struct {
	// Type of environment
	Type EnvironmentType

	// Width and Height of the ambient area
	Width  int
	Height int

	// GenreID for color selection
	GenreID string

	// Seed for deterministic generation
	Seed int64

	// Density controls particle count (0.0-1.0, default 0.5)
	Density float64

	// Custom parameters
	Custom map[string]interface{}
}

// DefaultAmbienceConfig returns a default ambience configuration.
func DefaultAmbienceConfig() AmbienceConfig {
	return AmbienceConfig{
		Type:    EnvironmentDungeon,
		Width:   800,
		Height:  600,
		GenreID: "fantasy",
		Seed:    0,
		Density: 0.5,
		Custom:  make(map[string]interface{}),
	}
}

// Validate checks if the ambience configuration is valid.
func (c AmbienceConfig) Validate() error {
	if c.Width <= 0 {
		return fmt.Errorf("width must be positive, got %d", c.Width)
	}
	if c.Height <= 0 {
		return fmt.Errorf("height must be positive, got %d", c.Height)
	}
	if c.GenreID == "" {
		return fmt.Errorf("genreID cannot be empty")
	}
	if c.Density < 0.0 || c.Density > 1.0 {
		return fmt.Errorf("density must be between 0.0 and 1.0, got %f", c.Density)
	}
	return nil
}

// GetParticleCount returns the number of ambient particles based on density.
func (c AmbienceConfig) GetParticleCount() int {
	// Base density: 50-100 particles for standard 800x600 area
	area := float64(c.Width * c.Height)
	standardArea := 800.0 * 600.0
	
	// Scale particles based on area and density
	baseCount := 75.0 * (area / standardArea) * c.Density
	
	// Clamp between 10 and 100 particles
	count := int(baseCount)
	if count < 10 {
		count = 10
	}
	if count > 100 {
		count = 100
	}
	
	return count
}

// AmbienceSystem represents an ambient particle system.
type AmbienceSystem struct {
	// Configuration
	Config AmbienceConfig

	// Active particles
	Particles []Particle

	// Elapsed time
	ElapsedTime float64

	// Random number generator
	rng *rand.Rand
}

// GenerateAmbience creates a new ambient particle system.
func GenerateAmbience(config AmbienceConfig) (*AmbienceSystem, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	rng := rand.New(rand.NewSource(config.Seed))
	particleCount := config.GetParticleCount()

	particles := make([]Particle, particleCount)

	// Generate particles based on environment type
	switch config.Type {
	case EnvironmentDungeon:
		generateDungeonAmbience(particles, config, rng)
	case EnvironmentCave:
		generateCaveAmbience(particles, config, rng)
	case EnvironmentForest:
		generateForestAmbience(particles, config, rng)
	case EnvironmentDesert:
		generateDesertAmbience(particles, config, rng)
	case EnvironmentSnow:
		generateSnowAmbience(particles, config, rng)
	case EnvironmentSwamp:
		generateSwampAmbience(particles, config, rng)
	case EnvironmentLava:
		generateLavaAmbience(particles, config, rng)
	case EnvironmentCity:
		generateCityAmbience(particles, config, rng)
	case EnvironmentLaboratory:
		generateLaboratoryAmbience(particles, config, rng)
	case EnvironmentRuins:
		generateRuinsAmbience(particles, config, rng)
	default:
		generateDungeonAmbience(particles, config, rng) // fallback
	}

	return &AmbienceSystem{
		Config:      config,
		Particles:   particles,
		ElapsedTime: 0,
		rng:         rng,
	}, nil
}

// Update advances the ambient particle system by deltaTime seconds.
func (s *AmbienceSystem) Update(deltaTime float64) {
	s.ElapsedTime += deltaTime

	for i := range s.Particles {
		p := &s.Particles[i]

		// Update position
		p.X += p.VX * deltaTime
		p.Y += p.VY * deltaTime

		// Update lifetime
		p.Life -= deltaTime

		// Respawn particle if dead
		if p.Life <= 0 {
			s.respawnParticle(p, i)
		}

		// Apply environment-specific behaviors
		s.applyEnvironmentBehavior(p, deltaTime)
	}
}

// respawnParticle resets a particle to initial state with new random properties.
func (s *AmbienceSystem) respawnParticle(p *Particle, index int) {
	// Use a deterministic but varied seed for respawning
	respawnSeed := s.Config.Seed + int64(index) + int64(s.ElapsedTime*1000)
	localRng := rand.New(rand.NewSource(respawnSeed))

	// Reset to random position in area
	p.X = localRng.Float64() * float64(s.Config.Width)
	p.Y = localRng.Float64() * float64(s.Config.Height)

	// Reset velocity based on environment type
	baseVelocity := getEnvironmentVelocity(s.Config.Type)
	p.VX = (localRng.Float64()*2.0 - 1.0) * baseVelocity * 0.3
	p.VY = (localRng.Float64()*2.0 - 1.0) * baseVelocity * 0.3

	// Reset lifetime
	p.Life = getEnvironmentLifetime(s.Config.Type, localRng)
	p.InitialLife = p.Life

	// Reset color (keep original color from generation)
	// Size and rotation stay the same
}

// applyEnvironmentBehavior applies environment-specific particle behaviors.
func (s *AmbienceSystem) applyEnvironmentBehavior(p *Particle, deltaTime float64) {
	switch s.Config.Type {
	case EnvironmentDungeon, EnvironmentCave, EnvironmentRuins:
		// Slow drifting motion
		applyDriftBehavior(p, deltaTime, 0.1)
	case EnvironmentForest:
		// Gentle swaying motion
		applySineWaveBehavior(p, deltaTime, 0.5, 2.0)
	case EnvironmentDesert:
		// Erratic wind gusts
		applyGustBehavior(p, deltaTime, 0.3, s.ElapsedTime)
	case EnvironmentSnow:
		// Gentle falling with drift
		applySnowDriftBehavior(p, deltaTime, 0.2)
	case EnvironmentSwamp:
		// Slow vertical oscillation (fireflies)
		applyFloatBehavior(p, deltaTime, 1.0)
	case EnvironmentLava:
		// Rising motion with turbulence
		applyRiseBehavior(p, deltaTime, 20.0)
	case EnvironmentCity:
		// Tumbling paper-like motion
		applyTumbleBehavior(p, deltaTime, 1.5)
	case EnvironmentLaboratory:
		// Smooth circular motion
		applyOrbitBehavior(p, deltaTime, 30.0, s.ElapsedTime)
	}

	// Fade particles near end of life
	applyFadeBehavior(p)
}

// generateDungeonAmbience creates dust motes and moisture droplets for dungeons.
func generateDungeonAmbience(particles []Particle, config AmbienceConfig, rng *rand.Rand) {
	for i := range particles {
		p := &particles[i]
		
		// Random position
		p.X = rng.Float64() * float64(config.Width)
		p.Y = rng.Float64() * float64(config.Height)
		
		// Very slow drifting motion
		p.VX = (rng.Float64()*2.0 - 1.0) * 5.0
		p.VY = (rng.Float64()*2.0 - 1.0) * 5.0
		
		// Subtle grey/brown dust color
		brightness := uint8(80 + rng.Intn(100))
		p.Color = color.RGBA{R: brightness, G: brightness - 10, B: brightness - 20, A: 80}
		
		// Small size (1-2 pixels)
		p.Size = 1.0 + rng.Float64()
		
		// Long lifetime (10-30 seconds before respawn)
		p.Life = 10.0 + rng.Float64()*20.0
		p.InitialLife = p.Life
		
		p.Rotation = rng.Float64() * 2.0 * math.Pi
		p.RotationVel = (rng.Float64()*2.0 - 1.0) * 0.5
	}
}

// generateCaveAmbience creates mineral dust and water droplets for caves.
func generateCaveAmbience(particles []Particle, config AmbienceConfig, rng *rand.Rand) {
	for i := range particles {
		p := &particles[i]
		
		p.X = rng.Float64() * float64(config.Width)
		p.Y = rng.Float64() * float64(config.Height)
		
		// Mix of floating dust and occasional drips
		isDrip := rng.Float64() < 0.2
		if isDrip {
			p.VX = (rng.Float64()*2.0 - 1.0) * 2.0
			p.VY = 30.0 + rng.Float64()*20.0 // falling
		} else {
			p.VX = (rng.Float64()*2.0 - 1.0) * 3.0
			p.VY = (rng.Float64()*2.0 - 1.0) * 3.0
		}
		
		// Blue-grey cave colors
		if isDrip {
			// Water droplets - light blue
			p.Color = color.RGBA{R: 100, G: 150, B: 200, A: 150}
			p.Size = 1.5 + rng.Float64()*0.5
		} else {
			// Mineral dust - grey
			brightness := uint8(70 + rng.Intn(80))
			p.Color = color.RGBA{R: brightness, G: brightness + 10, B: brightness + 20, A: 60}
			p.Size = 1.0 + rng.Float64()
		}
		
		p.Life = 8.0 + rng.Float64()*12.0
		p.InitialLife = p.Life
		p.Rotation = rng.Float64() * 2.0 * math.Pi
		p.RotationVel = (rng.Float64()*2.0 - 1.0) * 0.3
	}
}

// generateForestAmbience creates falling leaves and pollen for forests.
func generateForestAmbience(particles []Particle, config AmbienceConfig, rng *rand.Rand) {
	for i := range particles {
		p := &particles[i]
		
		p.X = rng.Float64() * float64(config.Width)
		p.Y = rng.Float64() * float64(config.Height)
		
		// Gentle falling motion with horizontal drift
		p.VX = (rng.Float64()*2.0 - 1.0) * 10.0
		p.VY = 8.0 + rng.Float64()*12.0
		
		// Green/brown/yellow leaf colors
		colorType := rng.Intn(3)
		switch colorType {
		case 0: // Green leaves
			p.Color = color.RGBA{R: 60, G: 120 + uint8(rng.Intn(60)), B: 40, A: 120}
		case 1: // Brown leaves
			p.Color = color.RGBA{R: 120 + uint8(rng.Intn(60)), G: 80 + uint8(rng.Intn(40)), B: 40, A: 120}
		case 2: // Yellow pollen
			p.Color = color.RGBA{R: 220, G: 200 + uint8(rng.Intn(35)), B: 100, A: 80}
		}
		
		// Varied sizes for leaves vs pollen
		if colorType == 2 {
			p.Size = 0.5 + rng.Float64()*0.5 // tiny pollen
		} else {
			p.Size = 2.0 + rng.Float64()*2.0 // larger leaves
		}
		
		p.Life = 15.0 + rng.Float64()*10.0
		p.InitialLife = p.Life
		p.Rotation = rng.Float64() * 2.0 * math.Pi
		p.RotationVel = (rng.Float64()*2.0 - 1.0) * 2.0 // leaves spin
	}
}

// generateDesertAmbience creates sand particles and heat shimmer for deserts.
func generateDesertAmbience(particles []Particle, config AmbienceConfig, rng *rand.Rand) {
	for i := range particles {
		p := &particles[i]
		
		p.X = rng.Float64() * float64(config.Width)
		p.Y = rng.Float64() * float64(config.Height)
		
		// Horizontal wind-blown motion
		p.VX = 15.0 + rng.Float64()*25.0
		p.VY = (rng.Float64()*2.0 - 1.0) * 5.0
		
		// Sandy yellow/tan colors
		r := uint8(200 + rng.Intn(55))
		g := uint8(160 + rng.Intn(60))
		b := uint8(100 + rng.Intn(50))
		p.Color = color.RGBA{R: r, G: g, B: b, A: 100}
		
		p.Size = 0.8 + rng.Float64()*1.5
		p.Life = 5.0 + rng.Float64()*8.0
		p.InitialLife = p.Life
		p.Rotation = rng.Float64() * 2.0 * math.Pi
		p.RotationVel = (rng.Float64()*2.0 - 1.0) * 3.0
	}
}

// generateSnowAmbience creates gentle drifting snow and ice crystals.
func generateSnowAmbience(particles []Particle, config AmbienceConfig, rng *rand.Rand) {
	for i := range particles {
		p := &particles[i]
		
		p.X = rng.Float64() * float64(config.Width)
		p.Y = rng.Float64() * float64(config.Height)
		
		// Slow falling with drift
		p.VX = (rng.Float64()*2.0 - 1.0) * 8.0
		p.VY = 5.0 + rng.Float64()*10.0
		
		// White/light blue colors
		brightness := uint8(220 + rng.Intn(35))
		p.Color = color.RGBA{R: brightness, G: brightness, B: 255, A: 150}
		
		p.Size = 1.0 + rng.Float64()*2.0
		p.Life = 12.0 + rng.Float64()*15.0
		p.InitialLife = p.Life
		p.Rotation = rng.Float64() * 2.0 * math.Pi
		p.RotationVel = (rng.Float64()*2.0 - 1.0) * 1.0
	}
}

// generateSwampAmbience creates mist and fireflies for swamps.
func generateSwampAmbience(particles []Particle, config AmbienceConfig, rng *rand.Rand) {
	for i := range particles {
		p := &particles[i]
		
		p.X = rng.Float64() * float64(config.Width)
		p.Y = rng.Float64() * float64(config.Height)
		
		// Mix of mist and fireflies
		isFirefly := rng.Float64() < 0.3
		if isFirefly {
			// Fireflies: slow floating motion
			p.VX = (rng.Float64()*2.0 - 1.0) * 8.0
			p.VY = -3.0 - rng.Float64()*5.0 // slight upward drift
			// Yellow-green glow
			p.Color = color.RGBA{R: 180, G: 220, B: 80, A: 180}
			p.Size = 1.5 + rng.Float64()*1.0
		} else {
			// Mist: very slow drift
			p.VX = (rng.Float64()*2.0 - 1.0) * 3.0
			p.VY = (rng.Float64()*2.0 - 1.0) * 2.0
			// Grey-green mist
			p.Color = color.RGBA{R: 120, G: 140, B: 120, A: 40}
			p.Size = 3.0 + rng.Float64()*4.0
		}
		
		p.Life = 10.0 + rng.Float64()*15.0
		p.InitialLife = p.Life
		p.Rotation = rng.Float64() * 2.0 * math.Pi
		p.RotationVel = (rng.Float64()*2.0 - 1.0) * 0.2
	}
}

// generateLavaAmbience creates ash and embers for volcanic environments.
func generateLavaAmbience(particles []Particle, config AmbienceConfig, rng *rand.Rand) {
	for i := range particles {
		p := &particles[i]
		
		p.X = rng.Float64() * float64(config.Width)
		p.Y = rng.Float64() * float64(config.Height)
		
		// Rising motion with turbulence
		p.VX = (rng.Float64()*2.0 - 1.0) * 15.0
		p.VY = -20.0 - rng.Float64()*30.0 // upward
		
		// Mix of ash and embers
		isEmber := rng.Float64() < 0.2
		if isEmber {
			// Glowing embers - red/orange
			p.Color = color.RGBA{R: 255, G: 100 + uint8(rng.Intn(100)), B: 20, A: 200}
			p.Size = 1.5 + rng.Float64()*1.5
		} else {
			// Dark ash
			brightness := uint8(40 + rng.Intn(60))
			p.Color = color.RGBA{R: brightness, G: brightness - 10, B: brightness - 20, A: 100}
			p.Size = 1.0 + rng.Float64()*1.0
		}
		
		p.Life = 6.0 + rng.Float64()*10.0
		p.InitialLife = p.Life
		p.Rotation = rng.Float64() * 2.0 * math.Pi
		p.RotationVel = (rng.Float64()*2.0 - 1.0) * 2.0
	}
}

// generateCityAmbience creates paper debris and smoke for urban environments.
func generateCityAmbience(particles []Particle, config AmbienceConfig, rng *rand.Rand) {
	for i := range particles {
		p := &particles[i]
		
		p.X = rng.Float64() * float64(config.Width)
		p.Y = rng.Float64() * float64(config.Height)
		
		// Mix of floating paper and smoke
		isPaper := rng.Float64() < 0.6
		if isPaper {
			// Paper debris - tumbling motion
			p.VX = (rng.Float64()*2.0 - 1.0) * 20.0
			p.VY = (rng.Float64()*2.0 - 1.0) * 10.0
			// White/grey paper colors
			brightness := uint8(180 + rng.Intn(75))
			p.Color = color.RGBA{R: brightness, G: brightness, B: brightness, A: 120}
			p.Size = 2.0 + rng.Float64()*3.0
			p.RotationVel = (rng.Float64()*2.0 - 1.0) * 4.0 // rapid tumbling
		} else {
			// Smoke - rising slowly
			p.VX = (rng.Float64()*2.0 - 1.0) * 5.0
			p.VY = -5.0 - rng.Float64()*10.0
			// Dark grey smoke
			brightness := uint8(60 + rng.Intn(80))
			p.Color = color.RGBA{R: brightness, G: brightness, B: brightness, A: 60}
			p.Size = 2.5 + rng.Float64()*3.5
			p.RotationVel = (rng.Float64()*2.0 - 1.0) * 0.5
		}
		
		p.Life = 8.0 + rng.Float64()*12.0
		p.InitialLife = p.Life
		p.Rotation = rng.Float64() * 2.0 * math.Pi
	}
}

// generateLaboratoryAmbience creates energy particles and sparks for labs.
func generateLaboratoryAmbience(particles []Particle, config AmbienceConfig, rng *rand.Rand) {
	for i := range particles {
		p := &particles[i]
		
		p.X = rng.Float64() * float64(config.Width)
		p.Y = rng.Float64() * float64(config.Height)
		
		// Smooth orbital or linear motion
		p.VX = (rng.Float64()*2.0 - 1.0) * 12.0
		p.VY = (rng.Float64()*2.0 - 1.0) * 12.0
		
		// Cyan/blue energy colors with occasional sparks
		isSpark := rng.Float64() < 0.15
		if isSpark {
			// Bright sparks
			p.Color = color.RGBA{R: 255, G: 255, B: 200, A: 220}
			p.Size = 1.0 + rng.Float64()*0.5
		} else {
			// Energy particles - cyan/blue
			p.Color = color.RGBA{R: 80, G: 180 + uint8(rng.Intn(75)), B: 255, A: 150}
			p.Size = 1.5 + rng.Float64()*1.5
		}
		
		p.Life = 7.0 + rng.Float64()*10.0
		p.InitialLife = p.Life
		p.Rotation = rng.Float64() * 2.0 * math.Pi
		p.RotationVel = (rng.Float64()*2.0 - 1.0) * 1.5
	}
}

// generateRuinsAmbience creates dust and falling debris for ruined structures.
func generateRuinsAmbience(particles []Particle, config AmbienceConfig, rng *rand.Rand) {
	for i := range particles {
		p := &particles[i]
		
		p.X = rng.Float64() * float64(config.Width)
		p.Y = rng.Float64() * float64(config.Height)
		
		// Mix of floating dust and occasionally falling debris
		isFalling := rng.Float64() < 0.15
		if isFalling {
			// Falling debris
			p.VX = (rng.Float64()*2.0 - 1.0) * 5.0
			p.VY = 15.0 + rng.Float64()*25.0
			// Greyish debris
			brightness := uint8(100 + rng.Intn(80))
			p.Color = color.RGBA{R: brightness, G: brightness - 10, B: brightness - 15, A: 140}
			p.Size = 2.0 + rng.Float64()*2.0
		} else {
			// Floating dust
			p.VX = (rng.Float64()*2.0 - 1.0) * 4.0
			p.VY = (rng.Float64()*2.0 - 1.0) * 4.0
			// Light dust
			brightness := uint8(90 + rng.Intn(100))
			p.Color = color.RGBA{R: brightness, G: brightness - 5, B: brightness - 10, A: 70}
			p.Size = 1.0 + rng.Float64()*1.5
		}
		
		p.Life = 12.0 + rng.Float64()*18.0
		p.InitialLife = p.Life
		p.Rotation = rng.Float64() * 2.0 * math.Pi
		p.RotationVel = (rng.Float64()*2.0 - 1.0) * 1.0
	}
}

// Helper functions for environment behaviors

// applyDriftBehavior adds slow random drift to particles.
func applyDriftBehavior(p *Particle, deltaTime, strength float64) {
	// Very subtle drift using elapsed time as noise source
	// Calculate age from InitialLife - Life
	age := p.InitialLife - p.Life
	driftX := math.Sin(age*0.5) * strength
	driftY := math.Cos(age*0.3) * strength
	p.VX += driftX * deltaTime
	p.VY += driftY * deltaTime
}

// applySineWaveBehavior adds sine wave motion to particles.
func applySineWaveBehavior(p *Particle, deltaTime, amplitude, frequency float64) {
	// Horizontal sine wave motion (like leaves swaying)
	age := p.InitialLife - p.Life
	p.VX += math.Sin(age*frequency) * amplitude * deltaTime
}

// applyGustBehavior adds sudden wind gusts to particles.
func applyGustBehavior(p *Particle, deltaTime, strength, time float64) {
	// Periodic gusts using global time
	gustPhase := math.Sin(time * 0.3)
	if gustPhase > 0.7 {
		// Gust active
		gustStrength := (gustPhase - 0.7) * 10.0 * strength
		p.VX += gustStrength * deltaTime * 100.0
	}
}

// applySnowDriftBehavior adds gentle drift to falling snow.
func applySnowDriftBehavior(p *Particle, deltaTime, strength float64) {
	// Gentle horizontal drift
	age := p.InitialLife - p.Life
	driftX := math.Sin(age*1.5 + p.X*0.01) * strength
	p.VX += driftX * deltaTime * 10.0
}

// applyFloatBehavior adds slow vertical oscillation to particles.
func applyFloatBehavior(p *Particle, deltaTime, amplitude float64) {
	// Vertical oscillation (like fireflies bobbing)
	age := p.InitialLife - p.Life
	p.VY += math.Sin(age*2.0) * amplitude * deltaTime
}

// applyRiseBehavior adds rising motion with turbulence.
func applyRiseBehavior(p *Particle, deltaTime, strength float64) {
	// Constant upward force with turbulence
	age := p.InitialLife - p.Life
	turbulenceX := math.Sin(age*3.0+p.X*0.05) * strength * 0.3
	turbulenceY := math.Cos(age*2.5+p.Y*0.05) * strength * 0.2
	p.VX += turbulenceX * deltaTime
	p.VY -= strength * deltaTime // upward
	p.VY += turbulenceY * deltaTime
}

// applyTumbleBehavior adds tumbling rotation to particles.
func applyTumbleBehavior(p *Particle, deltaTime, strength float64) {
	// Increase rotation speed based on velocity
	velocityMag := math.Sqrt(p.VX*p.VX + p.VY*p.VY)
	p.RotationVel = velocityMag * strength * 0.1
}

// applyOrbitBehavior adds circular orbital motion to particles.
func applyOrbitBehavior(p *Particle, deltaTime, radius, time float64) {
	// Circular motion around spawn point (stored in Custom if needed)
	// For simplicity, orbit around current position with gentle drift
	angle := time * 0.5
	orbitX := math.Cos(angle) * radius * 0.01
	orbitY := math.Sin(angle) * radius * 0.01
	p.VX = orbitX
	p.VY = orbitY
}

// applyFadeBehavior fades particles near end of life.
func applyFadeBehavior(p *Particle) {
	// Fade out in last 20% of lifetime
	threshold := p.InitialLife * 0.2
	if p.Life < threshold {
		fadeRatio := p.Life / threshold
		if fadeRatio < 0 {
			fadeRatio = 0
		}
		// Reduce alpha
		if c, ok := p.Color.(color.RGBA); ok {
			c.A = uint8(float64(c.A) * fadeRatio)
			p.Color = c
		}
	}
}

// Helper functions for environment properties

// getEnvironmentVelocity returns the base velocity for an environment type.
func getEnvironmentVelocity(envType EnvironmentType) float64 {
	switch envType {
	case EnvironmentDungeon, EnvironmentCave, EnvironmentRuins:
		return 5.0 // slow drift
	case EnvironmentForest:
		return 10.0 // gentle falling
	case EnvironmentDesert:
		return 20.0 // wind-blown
	case EnvironmentSnow:
		return 8.0 // gentle falling
	case EnvironmentSwamp:
		return 6.0 // slow float
	case EnvironmentLava:
		return 25.0 // rising fast
	case EnvironmentCity:
		return 15.0 // moderate motion
	case EnvironmentLaboratory:
		return 12.0 // smooth gliding
	default:
		return 5.0
	}
}

// getEnvironmentLifetime returns a random lifetime for an environment type.
func getEnvironmentLifetime(envType EnvironmentType, rng *rand.Rand) float64 {
	base := 0.0
	variation := 0.0
	
	switch envType {
	case EnvironmentDungeon, EnvironmentCave, EnvironmentRuins:
		base = 15.0
		variation = 15.0
	case EnvironmentForest:
		base = 15.0
		variation = 10.0
	case EnvironmentDesert:
		base = 5.0
		variation = 8.0
	case EnvironmentSnow:
		base = 12.0
		variation = 15.0
	case EnvironmentSwamp:
		base = 10.0
		variation = 15.0
	case EnvironmentLava:
		base = 6.0
		variation = 10.0
	case EnvironmentCity:
		base = 8.0
		variation = 12.0
	case EnvironmentLaboratory:
		base = 7.0
		variation = 10.0
	default:
		base = 10.0
		variation = 10.0
	}
	
	return base + rng.Float64()*variation
}
