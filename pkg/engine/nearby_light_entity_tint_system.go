// Package engine provides the NearbyLightEntityTintSystem which tints entity
// sprites based on nearby light sources. Entities within a light's radius receive
// a color tint proportional to the light's color, intensity, and distance-based
// falloff. Multiple lights blend additively. Genre presets control base ambient
// tint so horror worlds feel darker and fantasy worlds feel warmer. Tints compose
// multiplicatively with existing weather and status effect tints.
package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// NearbyLightTintComponent stores light-driven tint multipliers for an entity.
// These multiply with other tint sources (weather, status effects) in the render pipeline.
type NearbyLightTintComponent struct {
	// RGB multipliers (1.0 = no change). Values > 1.0 brighten, < 1.0 darken.
	TintR float64
	TintG float64
	TintB float64
}

// Type returns the component type identifier.
func (n *NearbyLightTintComponent) Type() string {
	return "nearby_light_tint"
}

// NewNearbyLightTintComponent creates a component with neutral (no-op) tint.
func NewNearbyLightTintComponent() *NearbyLightTintComponent {
	return &NearbyLightTintComponent{
		TintR: 1.0,
		TintG: 1.0,
		TintB: 1.0,
	}
}

// lightAmbientPreset holds genre-specific base ambient RGB multipliers.
type lightAmbientPreset struct {
	R, G, B float64
}

// NearbyLightEntityTintSystem reads light sources in the world and adjusts
// NearbyLightTintComponent tints on entities with sprites based on proximity.
type NearbyLightEntityTintSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// Throttle: recompute every updateInterval seconds
	updateInterval float64
	timeSinceCheck float64

	// Genre for ambient base tint
	genreID string

	// Max lights to consider per entity (performance bound)
	maxLightsPerEntity int

	// Pre-computed genre ambient presets
	ambientPresets map[string]lightAmbientPreset
}

// NewNearbyLightEntityTintSystem creates a new nearby light entity tint system.
func NewNearbyLightEntityTintSystem(world *World, seed int64) *NearbyLightEntityTintSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "nearby_light_entity_tint")
	}

	sys := &NearbyLightEntityTintSystem{
		world:              world,
		logger:             logEntry,
		rng:                rand.New(rand.NewSource(seed)),
		updateInterval:     0.25, // 4 updates per second
		genreID:            "fantasy",
		maxLightsPerEntity: 4,
		ambientPresets:     buildLightAmbientPresets(),
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("nearby light entity tint system created")
	}
	return sys
}

// SetGenre configures the active genre for ambient tint computation.
func (s *NearbyLightEntityTintSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for nearby light entity tint")
	}
}

// Update checks for nearby lights and applies tints to sprite entities.
func (s *NearbyLightEntityTintSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	// Collect light sources
	lights := s.collectLightSources(entities)
	if len(lights) == 0 {
		s.applyAmbientOnly(entities)
		return
	}

	s.applyLightTints(entities, lights)
}

// lightSourceInfo caches light data for tinting calculations.
type lightSourceInfo struct {
	x, y      float64
	r, g, b   float64 // Normalized color [0,1]
	radius    float64
	intensity float64
	falloff   LightFalloffType
}

// collectLightSources gathers enabled light entities with positions.
func (s *NearbyLightEntityTintSystem) collectLightSources(entities []*Entity) []lightSourceInfo {
	var lights []lightSourceInfo
	for _, entity := range entities {
		lightComp, hasLight := entity.GetComponent("light")
		if !hasLight {
			continue
		}
		light, ok := lightComp.(*LightComponent)
		if !ok || !light.Enabled || light.Radius <= 0 || light.Intensity <= 0 {
			continue
		}

		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		lights = append(lights, lightSourceInfo{
			x:         pos.X,
			y:         pos.Y,
			r:         float64(light.Color.R) / 255.0,
			g:         float64(light.Color.G) / 255.0,
			b:         float64(light.Color.B) / 255.0,
			radius:    light.Radius,
			intensity: light.Intensity,
			falloff:   light.Falloff,
		})
	}
	return lights
}

// applyLightTints computes and sets tints on entities near light sources.
func (s *NearbyLightEntityTintSystem) applyLightTints(entities []*Entity, lights []lightSourceInfo) {
	ambient := s.getAmbientPreset()
	count := 0

	for _, entity := range entities {
		if !entity.HasComponent("sprite") {
			continue
		}
		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		// Accumulate light contribution
		accR, accG, accB := 0.0, 0.0, 0.0
		lightsApplied := 0

		for i := range lights {
			if lightsApplied >= s.maxLightsPerEntity {
				break
			}
			light := &lights[i]
			dx := pos.X - light.x
			dy := pos.Y - light.y
			distSq := dx*dx + dy*dy
			radiusSq := light.radius * light.radius

			if distSq >= radiusSq {
				continue
			}

			dist := math.Sqrt(distSq)
			normalizedDist := dist / light.radius
			intensity := calculateLightFalloff(normalizedDist, light.falloff) * light.intensity

			accR += light.r * intensity
			accG += light.g * intensity
			accB += light.b * intensity
			lightsApplied++
		}

		tintComp := s.getOrCreateTintComponent(entity)
		if lightsApplied == 0 {
			tintComp.TintR = ambient.R
			tintComp.TintG = ambient.G
			tintComp.TintB = ambient.B
		} else {
			// Blend accumulated light color with ambient base
			tintComp.TintR = clampLightTint(ambient.R + accR*0.3)
			tintComp.TintG = clampLightTint(ambient.G + accG*0.3)
			tintComp.TintB = clampLightTint(ambient.B + accB*0.3)
		}
		count++
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"genre":           s.genreID,
			"entities_tinted": count,
			"light_sources":   len(lights),
		}).Debug("nearby light tints applied")
	}
}

// applyAmbientOnly sets ambient-only tints when no light sources exist.
func (s *NearbyLightEntityTintSystem) applyAmbientOnly(entities []*Entity) {
	ambient := s.getAmbientPreset()
	for _, entity := range entities {
		if !entity.HasComponent("sprite") {
			continue
		}
		tintComp := s.getOrCreateTintComponent(entity)
		tintComp.TintR = ambient.R
		tintComp.TintG = ambient.G
		tintComp.TintB = ambient.B
	}
}

// getOrCreateTintComponent retrieves or lazily creates the tint component.
func (s *NearbyLightEntityTintSystem) getOrCreateTintComponent(entity *Entity) *NearbyLightTintComponent {
	comp, ok := entity.GetComponent("nearby_light_tint")
	if ok {
		if tintComp, ok := comp.(*NearbyLightTintComponent); ok {
			return tintComp
		}
	}
	tintComp := NewNearbyLightTintComponent()
	entity.AddComponent(tintComp)
	return tintComp
}

// getAmbientPreset returns the genre-specific ambient tint preset.
func (s *NearbyLightEntityTintSystem) getAmbientPreset() lightAmbientPreset {
	preset, ok := s.ambientPresets[s.genreID]
	if !ok {
		return lightAmbientPreset{R: 1.0, G: 1.0, B: 1.0}
	}
	return preset
}

// calculateLightFalloff computes light intensity at normalized distance [0,1].
func calculateLightFalloff(normalizedDist float64, falloff LightFalloffType) float64 {
	if normalizedDist <= 0 {
		return 1.0
	}
	if normalizedDist >= 1.0 {
		return 0.0
	}
	switch falloff {
	case FalloffLinear:
		return 1.0 - normalizedDist
	case FalloffQuadratic:
		remaining := 1.0 - normalizedDist
		return remaining * remaining
	case FalloffInverseSquare:
		return 1.0 / (1.0 + normalizedDist*normalizedDist*4.0)
	case FalloffConstant:
		return 1.0
	default:
		return 1.0 - normalizedDist
	}
}

// buildLightAmbientPresets returns genre-specific base ambient tint values.
// These define the "unlit" base tint for entities in each genre.
func buildLightAmbientPresets() map[string]lightAmbientPreset {
	return map[string]lightAmbientPreset{
		"fantasy":   {R: 1.00, G: 0.98, B: 0.92}, // Warm golden ambient
		"horror":    {R: 0.78, G: 0.75, B: 0.82}, // Cool desaturated darkness
		"scifi":     {R: 0.90, G: 0.95, B: 1.00}, // Sterile blue-white ambient
		"cyberpunk": {R: 0.85, G: 0.80, B: 0.95}, // Dim purple-tinted ambient
		"postapoc":  {R: 0.92, G: 0.85, B: 0.72}, // Dusty sepia ambient
	}
}

// clampLightTint ensures a tint multiplier stays in [0.5, 1.2].
func clampLightTint(v float64) float64 {
	if v < 0.5 {
		return 0.5
	}
	if v > 1.2 {
		return 1.2
	}
	return v
}
