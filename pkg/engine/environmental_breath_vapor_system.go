// Package engine provides the EnvironmentalBreathVaporSystem which spawns
// visible breath vapor particles near entity faces during cold weather.
// Cold weather types (Snow, Ash, Fog) trigger periodic puffs that drift
// upward and fade. Genre-aware styling controls color and opacity:
// fantasy=frost-white puffs, scifi=cool condensation, horror=ghostly mist,
// cyberpunk=neon-tinted steam. Connects WeatherComponent with entity
// position/velocity to add environmental immersion.
package engine

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// BreathVaporComponent stores per-entity breath vapor state.
type BreathVaporComponent struct {
	// Current cooldown until next puff (seconds)
	Cooldown float64
	// Base interval between puffs (seconds)
	Interval float64
	// Active vapor puffs being rendered
	Puffs []BreathPuff
	// Whether vapor is currently enabled (driven by weather)
	Active bool
}

// Type returns the component type identifier.
func (b *BreathVaporComponent) Type() string {
	return "breath_vapor"
}

// BreathPuff represents a single vapor puff particle.
type BreathPuff struct {
	X, Y    float64 // Position offset from entity
	VX, VY  float64 // Drift velocity
	Age     float64 // Time alive (seconds)
	MaxAge  float64 // Lifetime (seconds)
	Opacity float64 // Starting opacity
	Size    float64 // Puff radius in pixels
	R, G, B float64 // Color channels [0,1]
}

// breathVaporGenrePreset holds genre-specific vapor appearance settings.
type breathVaporGenrePreset struct {
	R, G, B     float64 // Base color
	Opacity     float64 // Base opacity [0,1]
	SizeMin     float64 // Min puff radius
	SizeMax     float64 // Max puff radius
	DriftSpeedY float64 // Upward drift speed (px/s)
	Lifetime    float64 // Puff lifetime (seconds)
}

// EnvironmentalBreathVaporSystem spawns breath vapor puffs on entities
// when cold weather (Snow, Ash, Fog) is active.
type EnvironmentalBreathVaporSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	genreID string
	preset  breathVaporGenrePreset

	// Throttle weather checks
	weatherCheckInterval float64
	timeSinceCheck       float64
	coldWeatherActive    bool

	// Cold weather types that trigger breath vapor
	coldTypes map[particles.WeatherType]bool
}

// NewEnvironmentalBreathVaporSystem creates a new breath vapor system.
func NewEnvironmentalBreathVaporSystem(world *World, seed int64) *EnvironmentalBreathVaporSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "breath_vapor")
	}

	sys := &EnvironmentalBreathVaporSystem{
		world:                world,
		logger:               logEntry,
		rng:                  rand.New(rand.NewSource(seed)),
		genreID:              "fantasy",
		weatherCheckInterval: 0.5,
		coldTypes: map[particles.WeatherType]bool{
			particles.WeatherSnow: true,
			particles.WeatherFog:  true,
			particles.WeatherAsh:  true,
		},
	}
	sys.preset = sys.buildPreset()

	if logEntry != nil {
		logEntry.Debug("EnvironmentalBreathVaporSystem created")
	}
	return sys
}

// SetGenre configures genre-aware vapor styling.
func (s *EnvironmentalBreathVaporSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.buildPreset()
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for breath vapor")
	}
}

// buildPreset returns genre-specific vapor visual parameters.
func (s *EnvironmentalBreathVaporSystem) buildPreset() breathVaporGenrePreset {
	switch s.genreID {
	case "horror":
		return breathVaporGenrePreset{
			R: 0.7, G: 0.75, B: 0.8,
			Opacity: 0.55, SizeMin: 2.0, SizeMax: 4.5,
			DriftSpeedY: -6.0, Lifetime: 1.2,
		}
	case "cyberpunk":
		return breathVaporGenrePreset{
			R: 0.6, G: 0.85, B: 1.0,
			Opacity: 0.45, SizeMin: 1.5, SizeMax: 3.5,
			DriftSpeedY: -8.0, Lifetime: 0.8,
		}
	case "sci-fi", "scifi":
		return breathVaporGenrePreset{
			R: 0.8, G: 0.9, B: 1.0,
			Opacity: 0.35, SizeMin: 1.5, SizeMax: 3.0,
			DriftSpeedY: -7.0, Lifetime: 0.9,
		}
	case "post-apocalyptic", "postapoc":
		return breathVaporGenrePreset{
			R: 0.75, G: 0.72, B: 0.68,
			Opacity: 0.5, SizeMin: 2.0, SizeMax: 4.0,
			DriftSpeedY: -5.0, Lifetime: 1.1,
		}
	default: // fantasy
		return breathVaporGenrePreset{
			R: 0.9, G: 0.93, B: 1.0,
			Opacity: 0.4, SizeMin: 1.5, SizeMax: 3.5,
			DriftSpeedY: -7.0, Lifetime: 1.0,
		}
	}
}

// Update checks weather state and manages breath vapor puffs on all entities.
func (s *EnvironmentalBreathVaporSystem) Update(entities []*Entity, deltaTime float64) {
	if s.world == nil {
		return
	}

	// Throttle weather detection
	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck >= s.weatherCheckInterval {
		s.timeSinceCheck = 0
		s.coldWeatherActive = s.detectColdWeather(entities)
	}

	for _, entity := range entities {
		s.processEntity(entity, deltaTime)
	}
}

// detectColdWeather scans for an active cold weather entity.
func (s *EnvironmentalBreathVaporSystem) detectColdWeather(entities []*Entity) bool {
	for _, entity := range entities {
		comp, ok := entity.GetComponent("weather")
		if !ok {
			continue
		}
		weather, ok := comp.(*WeatherComponent)
		if !ok || !weather.Active {
			continue
		}
		if s.coldTypes[weather.Config.Type] {
			return true
		}
	}
	return false
}

// processEntity updates or creates breath vapor for a single entity.
func (s *EnvironmentalBreathVaporSystem) processEntity(entity *Entity, deltaTime float64) {
	// Only process entities with position and sprite (visible characters)
	pos := entity.GetPosition()
	if pos == nil {
		return
	}
	if !entity.HasComponent("sprite") {
		return
	}

	vapor := s.getOrCreateVapor(entity)

	// Activate/deactivate based on weather
	vapor.Active = s.coldWeatherActive

	// Age and remove expired puffs
	s.agePuffs(vapor, deltaTime)

	if !vapor.Active {
		return
	}

	// Spawn new puffs on cooldown
	vapor.Cooldown -= deltaTime
	if vapor.Cooldown <= 0 {
		vapor.Cooldown = vapor.Interval + s.rng.Float64()*0.3
		s.spawnPuff(vapor, pos)
	}
}

// getOrCreateVapor retrieves or lazily creates the breath vapor component.
func (s *EnvironmentalBreathVaporSystem) getOrCreateVapor(entity *Entity) *BreathVaporComponent {
	comp, ok := entity.GetComponent("breath_vapor")
	if ok {
		if v, ok := comp.(*BreathVaporComponent); ok {
			return v
		}
	}
	v := &BreathVaporComponent{
		Interval: 0.8 + s.rng.Float64()*0.4,
		Cooldown: s.rng.Float64() * 0.5, // stagger initial puffs
		Puffs:    make([]BreathPuff, 0, 4),
	}
	entity.AddComponent(v)
	return v
}

// spawnPuff adds a new vapor puff near the entity's head position.
func (s *EnvironmentalBreathVaporSystem) spawnPuff(vapor *BreathVaporComponent, pos *PositionComponent) {
	// Cap active puffs to avoid memory growth
	if len(vapor.Puffs) >= 4 {
		return
	}

	p := s.preset
	size := p.SizeMin + s.rng.Float64()*(p.SizeMax-p.SizeMin)

	// Offset from head area: slightly forward and up from center
	offsetX := -3.0 + s.rng.Float64()*6.0
	offsetY := -14.0 + s.rng.Float64()*2.0 // near top of 32px sprite

	puff := BreathPuff{
		X:       pos.X + offsetX,
		Y:       pos.Y + offsetY,
		VX:      -1.0 + s.rng.Float64()*2.0, // slight horizontal drift
		VY:      p.DriftSpeedY,
		Age:     0,
		MaxAge:  p.Lifetime * (0.8 + s.rng.Float64()*0.4),
		Opacity: p.Opacity,
		Size:    size,
		R:       p.R,
		G:       p.G,
		B:       p.B,
	}
	vapor.Puffs = append(vapor.Puffs, puff)
}

// agePuffs advances all puffs and removes expired ones.
func (s *EnvironmentalBreathVaporSystem) agePuffs(vapor *BreathVaporComponent, deltaTime float64) {
	n := 0
	for i := range vapor.Puffs {
		puff := &vapor.Puffs[i]
		puff.Age += deltaTime
		if puff.Age >= puff.MaxAge {
			continue
		}
		// Drift
		puff.X += puff.VX * deltaTime
		puff.Y += puff.VY * deltaTime
		// Expand slightly as it ages
		progress := puff.Age / puff.MaxAge
		puff.Size += 1.5 * deltaTime
		// Fade opacity using smooth falloff
		puff.Opacity = s.preset.Opacity * (1.0 - math.Pow(progress, 1.5))
		vapor.Puffs[n] = vapor.Puffs[i]
		n++
	}
	vapor.Puffs = vapor.Puffs[:n]
}

// GetActiveVaporPuffs returns all current vapor puffs for an entity (used by renderer).
func GetActiveVaporPuffs(entity *Entity) []BreathPuff {
	comp, ok := entity.GetComponent("breath_vapor")
	if !ok {
		return nil
	}
	v, ok := comp.(*BreathVaporComponent)
	if !ok || !v.Active {
		return nil
	}
	return v.Puffs
}
