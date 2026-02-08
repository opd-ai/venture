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

	// Reusable buffer for particle collection (performance optimization)
	particleBuffer []WeatherParticleData
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
	// Process all entities passed in (filtered by World if needed)
	for _, entity := range entities {
		// Check if entity has weather component
		if !entity.HasComponent("weather") {
			continue
		}
		ws.updateWeather(entity, deltaTime)
	}
}

// updateWeather handles a single entity's weather effect.
func (ws *WeatherSystem) updateWeather(entity *Entity, deltaTime float64) {
	weather, ok := ws.getWeatherComponent(entity)
	if !ok {
		return
	}

	transitionCompleted := weather.UpdateTransition(deltaTime)

	if transitionCompleted {
		ws.handleTransitionCompletion(weather)
	}

	ws.updateWeatherSystem(weather, deltaTime)
}

// getWeatherComponent retrieves and validates the weather component from an entity.
func (ws *WeatherSystem) getWeatherComponent(entity *Entity) (*WeatherComponent, bool) {
	comp, ok := entity.GetComponent("weather")
	if !ok {
		return nil, false
	}

	weather, ok := comp.(*WeatherComponent)
	if !ok {
		return nil, false
	}

	return weather, true
}

// handleTransitionCompletion handles weather transition completion logic.
func (ws *WeatherSystem) handleTransitionCompletion(weather *WeatherComponent) {
	if !weather.Active && weather.System != nil {
		isWeatherChange := ws.isWeatherConfigChanged(weather)
		weather.System = nil

		if isWeatherChange {
			weather.StartWeather()
		}
	}
}

// isWeatherConfigChanged checks if the weather configuration has changed.
func (ws *WeatherSystem) isWeatherConfigChanged(weather *WeatherComponent) bool {
	if weather.System == nil {
		return false
	}
	return weather.System.Config.Type != weather.Config.Type
}

// updateWeatherSystem updates the weather particle system if active.
func (ws *WeatherSystem) updateWeatherSystem(weather *WeatherComponent, deltaTime float64) {
	if weather.System != nil && (weather.Active || weather.Transitioning) {
		weather.System.Update(deltaTime)
	}
}

// GetWeatherParticles returns all active weather particles for rendering.
// Returns particles with opacity applied for transitions.
// The rendering system should use this to draw weather effects.
// Performance: Reuses internal buffer to avoid per-frame allocations.
func (ws *WeatherSystem) GetWeatherParticles() []WeatherParticleData {
	if ws.world == nil {
		return nil
	}

	// Reuse buffer by resetting length (capacity preserved)
	ws.particleBuffer = ws.particleBuffer[:0]

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
			// Performance: p.Color is already color.RGBA, no conversion needed
			rgba := p.Color
			originalAlpha := float64(rgba.A)
			rgba.A = uint8(originalAlpha * opacity)

			ws.particleBuffer = append(ws.particleBuffer, WeatherParticleData{
				X:        p.X,
				Y:        p.Y,
				Size:     p.Size,
				Color:    rgba,
				Rotation: p.Rotation,
			})
		}
	}

	return ws.particleBuffer
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
