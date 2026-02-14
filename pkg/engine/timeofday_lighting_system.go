// Package engine provides the TimeOfDayLightingSystem for dynamic ambient lighting.
// This system connects the time-of-day palette modulation (from rendering/palette)
// with AmbientLightComponent, creating day/night cycles with smooth transitions.
// Phase 17.3: Time-of-Day Color Shifts integration with ECS.
package engine

import (
	"image/color"
	"math/rand"
	"time"

	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/sirupsen/logrus"
)

// TimeOfDayLightingSystem modulates ambient lighting based on game clock time.
// It applies genre-aware color and intensity shifts for immersive day/night cycles.
type TimeOfDayLightingSystem struct {
	world   *World
	clock   GameClock
	genreID string
	seed    int64
	rng     *rand.Rand
	logger  *logrus.Entry

	// Current time state
	currentTimeOfDay     palette.TimeOfDay
	transitionProgress   float64
	lastTransitionUpdate time.Time

	// Transition configuration
	transitionDuration float64 // Seconds for smooth transition (default 5.0)
	dayDuration        float64 // Game seconds per full day cycle (default 600.0 = 10 min)

	// Genre-specific intensity multipliers
	genreIntensity map[string]float64

	// Base ambient light settings (cached from first ambient light entity)
	baseAmbientColor     color.RGBA
	baseAmbientIntensity float64
	baseAmbientCached    bool
}

// NewTimeOfDayLightingSystem creates a time-of-day lighting system.
func NewTimeOfDayLightingSystem(world *World, seed int64) *TimeOfDayLightingSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "timeofday_lighting")
		logEntry.Debug("time-of-day lighting system created")
	}

	return &TimeOfDayLightingSystem{
		world:                world,
		seed:                 seed,
		rng:                  rand.New(rand.NewSource(seed)),
		logger:               logEntry,
		currentTimeOfDay:     palette.TimeOfDayDay,
		transitionProgress:   0.0,
		transitionDuration:   5.0,   // 5 second transitions
		dayDuration:          600.0, // 10 minute day cycle
		baseAmbientCached:    false,
		baseAmbientColor:     color.RGBA{200, 200, 210, 255},
		baseAmbientIntensity: 0.7,
		genreIntensity: map[string]float64{
			"fantasy":   1.0, // Full time-of-day effects
			"scifi":     0.7, // Artificial lighting dampens effects
			"horror":    1.2, // Enhanced darkness at night
			"cyberpunk": 0.6, // Neon lighting overrides natural
			"postapoc":  1.1, // Harsh sun, dark nights
		},
	}
}

// SetClock sets the game clock for time tracking.
func (s *TimeOfDayLightingSystem) SetClock(clock GameClock) {
	s.clock = clock
	if s.logger != nil {
		s.logger.Debug("game clock linked")
	}
}

// SetGenre sets the genre ID for genre-aware lighting intensity.
func (s *TimeOfDayLightingSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// SetDayDuration sets the game-time seconds for a full day/night cycle.
func (s *TimeOfDayLightingSystem) SetDayDuration(seconds float64) {
	if seconds > 0 {
		s.dayDuration = seconds
	}
}

// SetTransitionDuration sets the transition time between time periods.
func (s *TimeOfDayLightingSystem) SetTransitionDuration(seconds float64) {
	if seconds > 0 {
		s.transitionDuration = seconds
	}
}

// Update processes ambient light entities and applies time-based modulation.
func (s *TimeOfDayLightingSystem) Update(entities []*Entity, deltaTime float64) {
	if s.world == nil || s.clock == nil {
		return
	}

	// Calculate current time of day from game clock
	gameTime := s.clock.Now()
	newTimeOfDay, transitionProgress := s.calculateTimeOfDay(gameTime)

	// Detect time state transitions
	if newTimeOfDay != s.currentTimeOfDay {
		s.lastTransitionUpdate = gameTime
		s.currentTimeOfDay = newTimeOfDay
		s.transitionProgress = 0.0

		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"new_time_of_day": newTimeOfDay.String(),
			}).Debug("time of day changed")
		}
	}

	// Update transition progress
	s.transitionProgress = transitionProgress

	// Find and update ambient light entities
	for _, entity := range entities {
		ambient := s.getAmbientLight(entity)
		if ambient == nil {
			continue
		}

		// Cache base values on first encounter
		if !s.baseAmbientCached {
			s.baseAmbientColor = ambient.Color
			s.baseAmbientIntensity = ambient.Intensity
			s.baseAmbientCached = true
		}

		// Apply time-of-day modulation
		s.applyTimeModulation(ambient)
	}
}

// calculateTimeOfDay determines the current time period based on game clock.
// Returns the time of day and transition progress to next period.
func (s *TimeOfDayLightingSystem) calculateTimeOfDay(gameTime time.Time) (palette.TimeOfDay, float64) {
	// Use hour-based calculation (0-23)
	hour := gameTime.Hour()
	minute := gameTime.Minute()
	hourFraction := float64(minute) / 60.0

	// Dawn: 5:00-7:59, Day: 8:00-16:59, Dusk: 17:00-19:59, Night: 20:00-4:59
	switch {
	case hour >= 5 && hour < 8:
		// Dawn period
		progress := (float64(hour-5) + hourFraction) / 3.0
		return palette.TimeOfDayDawn, progress
	case hour >= 8 && hour < 17:
		// Day period
		progress := (float64(hour-8) + hourFraction) / 9.0
		return palette.TimeOfDayDay, progress
	case hour >= 17 && hour < 20:
		// Dusk period
		progress := (float64(hour-17) + hourFraction) / 3.0
		return palette.TimeOfDayDusk, progress
	default:
		// Night period (20:00-4:59)
		var progress float64
		if hour >= 20 {
			progress = (float64(hour-20) + hourFraction) / 9.0
		} else {
			progress = (float64(hour+4) + hourFraction) / 9.0
		}
		return palette.TimeOfDayNight, progress
	}
}

// getAmbientLight returns the AmbientLightComponent if present.
func (s *TimeOfDayLightingSystem) getAmbientLight(entity *Entity) *AmbientLightComponent {
	comp, ok := entity.GetComponent("ambient_light")
	if !ok || comp == nil {
		return nil
	}
	ambient, ok := comp.(*AmbientLightComponent)
	if !ok {
		return nil
	}
	return ambient
}

// applyTimeModulation adjusts ambient light based on current time of day.
func (s *TimeOfDayLightingSystem) applyTimeModulation(ambient *AmbientLightComponent) {
	// Build time config for palette modulation
	timeConfig := palette.TimeConfig{
		CurrentTime:         s.currentTimeOfDay,
		TransitionProgress:  s.transitionProgress,
		IntensityMultiplier: s.getGenreIntensity(),
	}

	// Get color modulation from palette system
	modulation := palette.GetModulationWithTransition(timeConfig)

	// Apply modulation to ambient light color
	ambient.Color = s.modulateColor(s.baseAmbientColor, modulation)

	// Apply intensity modulation
	ambient.Intensity = s.modulateIntensity(s.baseAmbientIntensity, modulation)

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"time_of_day":    s.currentTimeOfDay.String(),
			"transition":     s.transitionProgress,
			"intensity":      ambient.Intensity,
			"color_r":        ambient.Color.R,
			"color_g":        ambient.Color.G,
			"color_b":        ambient.Color.B,
			"genre_modifier": s.getGenreIntensity(),
		}).Debug("ambient light modulated")
	}
}

// modulateColor applies color modulation to ambient light color.
func (s *TimeOfDayLightingSystem) modulateColor(base color.RGBA, mod palette.ColorModulation) color.RGBA {
	// Convert to float for manipulation
	r := float64(base.R)
	g := float64(base.G)
	b := float64(base.B)

	// Apply temperature shift (warm = more red/yellow, cool = more blue)
	if mod.TemperatureShift > 0 {
		// Warm: increase red, decrease blue
		r = clampFloat(r*(1.0+mod.TemperatureShift*0.2), 0, 255)
		b = clampFloat(b*(1.0-mod.TemperatureShift*0.15), 0, 255)
	} else if mod.TemperatureShift < 0 {
		// Cool: increase blue, decrease red
		b = clampFloat(b*(1.0-mod.TemperatureShift*0.2), 0, 255)
		r = clampFloat(r*(1.0+mod.TemperatureShift*0.15), 0, 255)
	}

	// Apply saturation multiplier (affects color intensity)
	gray := (r + g + b) / 3.0
	r = clampFloat(gray+(r-gray)*mod.SaturationMultiplier, 0, 255)
	g = clampFloat(gray+(g-gray)*mod.SaturationMultiplier, 0, 255)
	b = clampFloat(gray+(b-gray)*mod.SaturationMultiplier, 0, 255)

	// Apply lightness offset
	offset := mod.LightnessOffset * 255.0
	r = clampFloat(r+offset, 0, 255)
	g = clampFloat(g+offset, 0, 255)
	b = clampFloat(b+offset, 0, 255)

	return color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: base.A,
	}
}

// modulateIntensity applies intensity modulation based on time of day.
func (s *TimeOfDayLightingSystem) modulateIntensity(base float64, mod palette.ColorModulation) float64 {
	// Lightness offset directly affects intensity
	intensity := base + mod.LightnessOffset

	// Clamp to valid range
	if intensity < 0.1 {
		intensity = 0.1 // Minimum visibility
	}
	if intensity > 1.0 {
		intensity = 1.0
	}

	return intensity
}

// getGenreIntensity returns the genre-specific intensity multiplier.
func (s *TimeOfDayLightingSystem) getGenreIntensity() float64 {
	if mult, ok := s.genreIntensity[s.genreID]; ok {
		return mult
	}
	return 1.0
}

// GetCurrentTimeOfDay returns the current time of day for external systems.
func (s *TimeOfDayLightingSystem) GetCurrentTimeOfDay() palette.TimeOfDay {
	return s.currentTimeOfDay
}

// GetTransitionProgress returns the current transition progress (0.0-1.0).
func (s *TimeOfDayLightingSystem) GetTransitionProgress() float64 {
	return s.transitionProgress
}

// GetDayDuration returns the configured day duration in seconds.
func (s *TimeOfDayLightingSystem) GetDayDuration() float64 {
	return s.dayDuration
}

// ForceTimeOfDay forces the current time of day to a specific value (for testing).
// This bypasses the clock-based calculation and sets the time directly.
func (s *TimeOfDayLightingSystem) ForceTimeOfDay(timeOfDay palette.TimeOfDay) {
	s.currentTimeOfDay = timeOfDay
	s.transitionProgress = 0.0
	if s.logger != nil {
		s.logger.WithField("time_of_day", timeOfDay.String()).Debug("forced time of day")
	}
}
