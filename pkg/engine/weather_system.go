// Package engine provides weather system management.
// This file implements the WeatherSystem that updates weather effects
// attached to entities.
package engine

import (
	"image/color"
)

// WeatherParticleData represents a single weather particle for rendering.
// This is a simplified view of particles.Particle optimized for rendering.
type WeatherParticleData struct {
	X        float64
	Y        float64
	Size     float64
	Color    color.RGBA
	Rotation float64
}

// WeatherSystem manages weather effects on entities.
// Phase 5.4: Weather Particle System Integration
//
// This system:
//   - Updates weather particle positions and states
//   - Manages weather transitions (fade in/out)
//   - Handles weather changes and crossfades
//   - Coordinates with rendering system for visual output
type WeatherSystem struct {
	// World reference for entity queries
	world *World

	// Viewport bounds for particle culling (optional)
	viewportX      float64
	viewportY      float64
	viewportWidth  float64
	viewportHeight float64
}

// NewWeatherSystem creates a new weather system.
func NewWeatherSystem(w *World) *WeatherSystem {
	return &WeatherSystem{
		world:          w,
		viewportX:      0,
		viewportY:      0,
		viewportWidth:  800,
		viewportHeight: 600,
	}
}

// SetViewport updates the viewport bounds for particle culling.
// Weather particles outside viewport are not rendered (performance optimization).
func (ws *WeatherSystem) SetViewport(x, y, width, height float64) {
	ws.viewportX = x
	ws.viewportY = y
	ws.viewportWidth = width
	ws.viewportHeight = height
}

// Update processes all weather effects.
// This method:
//   - Updates weather particle systems
//   - Manages transition states (fade in/out)
//   - Handles weather changes and crossfades
//   - Updates particle positions and lifetimes
func (ws *WeatherSystem) Update(entities []*Entity, deltaTime float64) {
	if ws.world == nil {
		return
	}

	// Get all entities with weather component
	weatherEntities := ws.world.GetEntitiesWith("weather")

	for _, entity := range weatherEntities {
		ws.updateWeather(entity, deltaTime)
	}
}

// updateWeather handles a single entity's weather effect.
func (ws *WeatherSystem) updateWeather(entity *Entity, deltaTime float64) {
	comp, ok := entity.GetComponent("weather")
	if !ok {
		return
	}

	weather, ok := comp.(*WeatherComponent)
	if !ok {
		return
	}

	// Update transition state
	transitionCompleted := weather.UpdateTransition(deltaTime)

	// Handle transition completion
	if transitionCompleted {
		// If we just finished fading out and need to start new weather
		if !weather.FadingIn && weather.System == nil {
			// This happens during weather change - old weather faded out
			// Start new weather (fade in)
			weather.StartWeather()
		}

		// If fading out completed and no new weather pending, clear system
		if !weather.Active && weather.System != nil {
			weather.System = nil
		}
	}

	// Update weather particle system if active
	if weather.System != nil && (weather.Active || weather.Transitioning) {
		weather.System.Update(deltaTime)
	}
}

// GetWeatherParticles returns all active weather particles for rendering.
// Returns particles with opacity applied for transitions.
// The rendering system should use this to draw weather effects.
func (ws *WeatherSystem) GetWeatherParticles() []WeatherParticleData {
	if ws.world == nil {
		return nil
	}

	var allParticles []WeatherParticleData

	weatherEntities := ws.world.GetEntitiesWith("weather")
	for _, entity := range weatherEntities {
		comp, ok := entity.GetComponent("weather")
		if !ok {
			continue
		}

		weather, ok := comp.(*WeatherComponent)
		if !ok || weather.System == nil {
			continue
		}

		// Only return particles if weather is visible
		opacity := weather.GetOpacity()
		if opacity <= 0 {
			continue
		}

		// Get particles from weather system
		for i := range weather.System.Particles {
			p := &weather.System.Particles[i]

			// Viewport culling for performance
			if !ws.isInViewport(p.X, p.Y) {
				continue
			}

			// Apply transition opacity to particle color
			color := p.Color
			// Adjust alpha based on opacity
			originalAlpha := float64(color.A)
			color.A = uint8(originalAlpha * opacity)

			allParticles = append(allParticles, WeatherParticleData{
				X:        p.X,
				Y:        p.Y,
				Size:     p.Size,
				Color:    color,
				Rotation: p.Rotation,
			})
		}
	}

	return allParticles
}

// isInViewport checks if a particle is within viewport bounds.
func (ws *WeatherSystem) isInViewport(x, y float64) bool {
	// Add padding for particles partially in viewport
	const padding = 50.0

	return x >= ws.viewportX-padding &&
		x <= ws.viewportX+ws.viewportWidth+padding &&
		y >= ws.viewportY-padding &&
		y <= ws.viewportY+ws.viewportHeight+padding
}

// GetActiveWeatherType returns the weather type currently active, if any.
// Returns empty string if no weather is active.
func (ws *WeatherSystem) GetActiveWeatherType() string {
	if ws.world == nil {
		return ""
	}

	weatherEntities := ws.world.GetEntitiesWith("weather")
	for _, entity := range weatherEntities {
		comp, ok := entity.GetComponent("weather")
		if !ok {
			continue
		}

		weather, ok := comp.(*WeatherComponent)
		if !ok {
			continue
		}

		if weather.IsFullyActive() {
			return weather.Config.Type.String()
		}
	}

	return ""
}

// GetWeatherCount returns the number of active weather entities.
func (ws *WeatherSystem) GetWeatherCount() int {
	if ws.world == nil {
		return 0
	}

	weatherEntities := ws.world.GetEntitiesWith("weather")
	count := 0
	for _, entity := range weatherEntities {
		comp, ok := entity.GetComponent("weather")
		if !ok {
			continue
		}

		weather, ok := comp.(*WeatherComponent)
		if !ok {
			continue
		}

		if weather.Active || weather.Transitioning {
			count++
		}
	}

	return count
}
