// Package engine provides weather system components for the ECS.
// This file implements components for managing weather effects on entities.
package engine

import (
	"github.com/opd-ai/venture/pkg/rendering/particles"
)

// WeatherComponent allows entities to display weather effects.
// This component stores an active weather particle system and its configuration.
// Typically attached to world/area entities for environmental atmosphere.
//
// Phase 5.4: Weather Particle System Integration
type WeatherComponent struct {
	// Active weather system
	System *particles.WeatherSystem

	// Weather configuration
	Config particles.WeatherConfig

	// Whether weather is currently active
	Active bool

	// Transition state (for smooth fade in/out)
	Transitioning   bool
	TransitionTime  float64 // Time spent in transition (seconds)
	TransitionTotal float64 // Total transition duration (seconds, default 5.0)
	FadingIn        bool    // true = fading in, false = fading out
}

// Type returns the component type identifier.
func (w *WeatherComponent) Type() string {
	return "weather"
}

// NewWeatherComponent creates a new weather component with the given configuration.
// The weather starts inactive and must be activated with StartWeather().
//
// Parameters:
//   - config: Weather configuration (type, intensity, dimensions, etc.)
//
// Returns: Initialized WeatherComponent
func NewWeatherComponent(config particles.WeatherConfig) *WeatherComponent {
	return &WeatherComponent{
		System:          nil, // Created when activated
		Config:          config,
		Active:          false,
		Transitioning:   false,
		TransitionTime:  0,
		TransitionTotal: 5.0, // Default 5-second transitions
		FadingIn:        false,
	}
}

// StartWeather activates weather with smooth fade-in transition.
// If weather is already active, this is a no-op.
func (w *WeatherComponent) StartWeather() error {
	if w.Active && !w.Transitioning {
		return nil // Already active
	}

	// Generate weather system if not already created
	if w.System == nil {
		system, err := particles.GenerateWeather(w.Config)
		if err != nil {
			return err
		}
		w.System = system
	}

	// Start fade-in transition
	w.Active = true
	w.Transitioning = true
	w.TransitionTime = 0
	w.FadingIn = true

	return nil
}

// StopWeather deactivates weather with smooth fade-out transition.
// If weather is already inactive, this is a no-op.
func (w *WeatherComponent) StopWeather() {
	if !w.Active && !w.Transitioning {
		return // Already inactive
	}

	// Start fade-out transition
	w.Transitioning = true
	w.TransitionTime = 0
	w.FadingIn = false
}

// GetOpacity returns the current opacity for weather rendering based on transition state.
// Returns 0.0 (fully transparent) to 1.0 (fully opaque).
func (w *WeatherComponent) GetOpacity() float64 {
	if !w.Active && !w.Transitioning {
		return 0.0
	}

	if !w.Transitioning {
		return 1.0 // Fully active
	}

	// Calculate transition progress (0.0 to 1.0)
	progress := w.TransitionTime / w.TransitionTotal
	if progress > 1.0 {
		progress = 1.0
	}

	if w.FadingIn {
		return progress // 0.0 → 1.0
	}
	// Fading out
	return 1.0 - progress // 1.0 → 0.0
}

// IsFullyActive returns true if weather is active and not transitioning.
func (w *WeatherComponent) IsFullyActive() bool {
	return w.Active && !w.Transitioning
}

// IsFullyInactive returns true if weather is inactive and not transitioning.
func (w *WeatherComponent) IsFullyInactive() bool {
	return !w.Active && !w.Transitioning
}

// UpdateTransition updates the transition state.
// Should be called by WeatherSystem during Update().
// Returns true if transition completed.
func (w *WeatherComponent) UpdateTransition(deltaTime float64) bool {
	if !w.Transitioning {
		return false
	}

	w.TransitionTime += deltaTime

	// Check if transition completed
	if w.TransitionTime >= w.TransitionTotal {
		w.Transitioning = false
		w.TransitionTime = 0

		// If fading out, mark as inactive
		if !w.FadingIn {
			w.Active = false
		}

		return true
	}

	return false
}

// ChangeWeather switches to a new weather type with crossfade.
// This fades out the current weather and fades in the new weather.
func (w *WeatherComponent) ChangeWeather(newConfig particles.WeatherConfig) error {
	// Validate new config
	if err := newConfig.Validate(); err != nil {
		return err
	}

	// Store new config
	w.Config = newConfig

	// If currently active, fade out first, then fade in new weather
	// If inactive, start new weather immediately
	if w.Active {
		w.StopWeather()
	} else {
		// Clear old system
		w.System = nil
		return w.StartWeather()
	}

	return nil
}
