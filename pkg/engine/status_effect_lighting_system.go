// Package engine provides the status effect lighting system.
// This file implements StatusEffectLightingSystem which connects status effects
// to the lighting system, causing entities with certain status effects to emit
// light appropriate to their effect type. For example:
// - Burning entities emit flickering orange light
// - Frozen/chilled entities emit subtle blue light
// - Poisoned entities emit faint green light
// - Regenerating entities emit soft golden light
//
// Design Philosophy:
// - Connects two existing systems (StatusEffect + Lighting) for visual feedback
// - Performance-conscious: only processes entities with both status effects and positions
// - Genre-aware: light colors and intensities can be configured per genre
// - Deterministic: uses seed-based randomness for flicker variations
package engine

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// StatusEffectLightConfig defines the light properties for a status effect type.
type StatusEffectLightConfig struct {
	Color      color.RGBA // Light color
	Radius     float64    // Light radius in pixels
	Intensity  float64    // Base intensity (0.0-1.0)
	Flickering bool       // Whether the light flickers
	FlickerAmt float64    // Flicker intensity variation (0.0-1.0)
	FlickerSpd float64    // Flicker speed (Hz)
	Pulsing    bool       // Whether the light pulses smoothly
	PulseSpeed float64    // Pulse frequency (Hz)
}

// StatusEffectLightingSystem adds dynamic lights to entities with status effects.
// This provides immediate visual feedback for combat and environmental hazards.
type StatusEffectLightingSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// Configuration maps status effect types to light properties
	effectConfigs map[string]StatusEffectLightConfig

	// Track which entities have status effect lights to clean up when effects expire
	// Key: entityID, Value: effect type that created the light
	activeLights map[uint64]string
}

// NewStatusEffectLightingSystem creates a new status effect lighting system.
func NewStatusEffectLightingSystem(world *World, seed int64) *StatusEffectLightingSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "status_effect_lighting",
	})

	sys := &StatusEffectLightingSystem{
		world:         world,
		logger:        logger,
		rng:           rand.New(rand.NewSource(seed)),
		effectConfigs: make(map[string]StatusEffectLightConfig),
		activeLights:  make(map[uint64]string),
	}

	// Initialize default effect configurations
	sys.initDefaultConfigs()

	logger.Debug("Status effect lighting system created")
	return sys
}

// initDefaultConfigs sets up default light configurations for common status effects.
func (s *StatusEffectLightingSystem) initDefaultConfigs() {
	// Burning - flickering orange/red fire light
	s.effectConfigs["burn"] = StatusEffectLightConfig{
		Color:      color.RGBA{R: 255, G: 120, B: 40, A: 255},
		Radius:     80.0,
		Intensity:  0.7,
		Flickering: true,
		FlickerAmt: 0.3,
		FlickerSpd: 8.0,
		Pulsing:    false,
	}

	// Poison - subtle green glow
	s.effectConfigs["poison"] = StatusEffectLightConfig{
		Color:      color.RGBA{R: 80, G: 200, B: 60, A: 255},
		Radius:     50.0,
		Intensity:  0.4,
		Flickering: false,
		Pulsing:    true,
		PulseSpeed: 1.5,
	}

	// Frost/Frozen - cool blue emanation
	s.effectConfigs["frost"] = StatusEffectLightConfig{
		Color:      color.RGBA{R: 100, G: 180, B: 255, A: 255},
		Radius:     60.0,
		Intensity:  0.5,
		Flickering: false,
		Pulsing:    true,
		PulseSpeed: 0.8,
	}
	s.effectConfigs["frozen"] = s.effectConfigs["frost"]

	// Regeneration - warm golden healing light
	s.effectConfigs["regeneration"] = StatusEffectLightConfig{
		Color:      color.RGBA{R: 255, G: 220, B: 100, A: 255},
		Radius:     45.0,
		Intensity:  0.35,
		Flickering: false,
		Pulsing:    true,
		PulseSpeed: 2.0,
	}

	// Shock/Electric - bright flickering white-blue
	s.effectConfigs["shock"] = StatusEffectLightConfig{
		Color:      color.RGBA{R: 200, G: 220, B: 255, A: 255},
		Radius:     70.0,
		Intensity:  0.8,
		Flickering: true,
		FlickerAmt: 0.5,
		FlickerSpd: 15.0,
		Pulsing:    false,
	}
	s.effectConfigs["electrified"] = s.effectConfigs["shock"]

	// Stun - dim pulsing yellow
	s.effectConfigs["stun"] = StatusEffectLightConfig{
		Color:      color.RGBA{R: 255, G: 255, B: 150, A: 255},
		Radius:     35.0,
		Intensity:  0.3,
		Flickering: false,
		Pulsing:    true,
		PulseSpeed: 3.0,
	}
}

// SetEffectConfig allows customizing light for a specific effect type.
// This enables genre-specific variations (e.g., cyberpunk burn = neon orange).
func (s *StatusEffectLightingSystem) SetEffectConfig(effectType string, config StatusEffectLightConfig) {
	s.effectConfigs[effectType] = config
	s.logger.WithFields(logrus.Fields{
		"effect_type": effectType,
		"radius":      config.Radius,
		"intensity":   config.Intensity,
	}).Debug("Effect light config updated")
}

// Update processes all entities, adding/updating/removing lights based on status effects.
func (s *StatusEffectLightingSystem) Update(entities []*Entity, deltaTime float64) {
	// Track which entities still have active status effect lights this frame
	activeThisFrame := make(map[uint64]bool)

	// Build map of entityID -> entity for cleanup
	entityMap := make(map[uint64]*Entity, len(entities))
	for _, entity := range entities {
		entityMap[entity.ID] = entity

		// Skip entities without position (can't place light)
		if entity.GetPosition() == nil {
			continue
		}

		// Find the most impactful status effect on this entity
		dominantEffect := s.findDominantStatusEffect(entity)

		if dominantEffect != "" {
			activeThisFrame[entity.ID] = true
			s.updateOrCreateLight(entity, dominantEffect)
		}
	}

	// Clean up lights for entities that no longer have status effects
	s.cleanupExpiredLights(activeThisFrame, entityMap)
}

// findDominantStatusEffect returns the effect type that should produce light.
// If multiple effects exist, prioritizes by visual impact (burn > shock > frost > poison > regen).
func (s *StatusEffectLightingSystem) findDominantStatusEffect(entity *Entity) string {
	// Priority order for visual effects
	priorityOrder := []string{"burn", "shock", "electrified", "frost", "frozen", "poison", "stun", "regeneration"}

	for _, effectType := range priorityOrder {
		for _, comp := range entity.Components {
			effect, ok := comp.(*StatusEffectComponent)
			if !ok {
				continue
			}

			// Check if this effect type matches and has config
			if effect.EffectType == effectType {
				if _, hasConfig := s.effectConfigs[effectType]; hasConfig {
					return effectType
				}
			}
		}
	}

	// Check for any other configured effect types
	for _, comp := range entity.Components {
		effect, ok := comp.(*StatusEffectComponent)
		if !ok {
			continue
		}
		if _, hasConfig := s.effectConfigs[effect.EffectType]; hasConfig {
			return effect.EffectType
		}
	}

	return ""
}

// updateOrCreateLight adds or updates the light component for a status effect.
func (s *StatusEffectLightingSystem) updateOrCreateLight(entity *Entity, effectType string) {
	config := s.effectConfigs[effectType]

	// Check if entity already has a light from status effects
	existingEffectType, hasExisting := s.activeLights[entity.ID]

	// If effect type changed, we need to update the light properties
	if hasExisting && existingEffectType != effectType {
		s.updateLightForEffect(entity, config, effectType)
		s.activeLights[entity.ID] = effectType
		return
	}

	// If no existing light, create one
	if !hasExisting {
		s.createLightForEffect(entity, config, effectType)
		s.activeLights[entity.ID] = effectType
		return
	}

	// Effect type unchanged, light already exists - no action needed
	// The LightingSystem handles flickering/pulsing animations
}

// createLightForEffect adds a new light component to an entity.
func (s *StatusEffectLightingSystem) createLightForEffect(entity *Entity, config StatusEffectLightConfig, effectType string) {
	light := &LightComponent{
		Color:         config.Color,
		Radius:        config.Radius,
		Intensity:     config.Intensity,
		Falloff:       FalloffQuadratic,
		Enabled:       true,
		Flickering:    config.Flickering,
		FlickerAmount: config.FlickerAmt,
		FlickerSpeed:  config.FlickerSpd,
		Pulsing:       config.Pulsing,
		PulseSpeed:    config.PulseSpeed,
	}

	entity.AddComponent(light)

	s.logger.WithFields(logrus.Fields{
		"entity_id":   entity.ID,
		"effect_type": effectType,
		"radius":      config.Radius,
	}).Debug("Created status effect light")
}

// updateLightForEffect updates an existing light component with new effect properties.
func (s *StatusEffectLightingSystem) updateLightForEffect(entity *Entity, config StatusEffectLightConfig, effectType string) {
	lightComp, hasLight := entity.GetComponent("light")
	if !hasLight {
		// Light was removed externally, recreate it
		s.createLightForEffect(entity, config, effectType)
		return
	}

	light, ok := lightComp.(*LightComponent)
	if !ok {
		return
	}

	// Update light properties for new effect type
	light.Color = config.Color
	light.Radius = config.Radius
	light.Intensity = config.Intensity
	light.Flickering = config.Flickering
	light.FlickerAmount = config.FlickerAmt
	light.FlickerSpeed = config.FlickerSpd
	light.Pulsing = config.Pulsing
	light.PulseSpeed = config.PulseSpeed

	s.logger.WithFields(logrus.Fields{
		"entity_id":   entity.ID,
		"effect_type": effectType,
	}).Debug("Updated status effect light")
}

// cleanupExpiredLights removes lights from entities that no longer have status effects.
func (s *StatusEffectLightingSystem) cleanupExpiredLights(activeThisFrame map[uint64]bool, entityMap map[uint64]*Entity) {
	for entityID := range s.activeLights {
		if !activeThisFrame[entityID] {
			// Entity no longer has a status effect, remove its light
			// First try the entity map (from Update's entity slice)
			if entity, exists := entityMap[entityID]; exists && entity != nil {
				s.removeLightFromEntity(entity)
			} else {
				// Fall back to world lookup for entities not in current Update slice
				if entity, exists := s.world.GetEntity(entityID); exists && entity != nil {
					s.removeLightFromEntity(entity)
				}
			}
			delete(s.activeLights, entityID)
		}
	}
}

// removeLightFromEntity removes the light component added by this system.
func (s *StatusEffectLightingSystem) removeLightFromEntity(entity *Entity) {
	// Use the entity's RemoveComponent method
	entity.RemoveComponent("light")
	s.logger.WithField("entity_id", entity.ID).Debug("Removed status effect light")
}

// GetActiveEffectLightCount returns the number of entities with status effect lights.
// Useful for debugging and performance monitoring.
func (s *StatusEffectLightingSystem) GetActiveEffectLightCount() int {
	return len(s.activeLights)
}

// ScaleIntensityByMagnitude adjusts light intensity based on effect magnitude.
// Higher magnitude effects (more damage, stronger buff) produce brighter lights.
func (s *StatusEffectLightingSystem) ScaleIntensityByMagnitude(baseIntensity, magnitude float64) float64 {
	// Scale factor: magnitude 10 = 1.0x, magnitude 50 = 1.5x, magnitude 100 = 2.0x
	scale := 1.0 + (magnitude-10.0)/90.0
	scale = math.Max(0.5, math.Min(2.0, scale)) // Clamp between 0.5x and 2.0x
	return baseIntensity * scale
}
