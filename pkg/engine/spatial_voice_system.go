// Package engine provides core ECS components and systems for the Venture game.
package engine

import (
	log "github.com/sirupsen/logrus"
)

// SpatialVoiceSystem calculates distance-based volume and stereo panning for voice.
type SpatialVoiceSystem struct {
	world *World

	// listenerEntityID tracks the local player for spatial calculations.
	listenerEntityID uint64

	// defaultMaxRange is the default maximum audible range.
	defaultMaxRange float64

	// defaultMinRange is the default minimum range (full volume).
	defaultMinRange float64
}

// NewSpatialVoiceSystem creates a new spatial voice system.
func NewSpatialVoiceSystem(world *World) *SpatialVoiceSystem {
	log.WithFields(log.Fields{
		"system_name": "spatial_voice",
	}).Debug("Creating spatial voice system")

	return &SpatialVoiceSystem{
		world:            world,
		listenerEntityID: 0,
		defaultMaxRange:  500.0,
		defaultMinRange:  50.0,
	}
}

// SetListener sets the entity that will be used as the listener for spatial calculations.
func (s *SpatialVoiceSystem) SetListener(entity *Entity) {
	if entity == nil {
		s.listenerEntityID = 0
		return
	}
	s.listenerEntityID = entity.ID
	log.WithFields(log.Fields{
		"entity_id": entity.ID,
	}).Debug("Spatial voice listener set")
}

// Update processes spatial audio for all entities with spatial voice components.
// Uses the world query cache to pre-filter entities, avoiding the HasComponent+GetComponent
// double map-lookup per entity per frame.
func (s *SpatialVoiceSystem) Update(entities []*Entity, deltaTime float64) {
	// Get listener position
	listenerX, listenerY, listenerOK := s.getListenerPosition(entities)
	if !listenerOK {
		return
	}

	// Pre-filter using the world query cache so we only iterate spatial_voice entities.
	voiceEntities := entities
	if s.world != nil {
		voiceEntities = s.world.GetEntitiesWith("spatial_voice", "position")
	}

	for _, entity := range voiceEntities {
		// Skip the listener entity
		if entity.ID == s.listenerEntityID {
			continue
		}

		comp, ok := entity.GetComponent("spatial_voice")
		if !ok {
			continue
		}
		spatial, ok := comp.(*SpatialVoiceComponent)
		if !ok {
			continue
		}

		// Get entity position
		sourceX, sourceY, posOK := s.getEntityPosition(entity)
		if !posOK {
			continue
		}

		// Update spatial calculations
		spatial.UpdatePositions(sourceX, sourceY, listenerX, listenerY)
	}
}

// getListenerPosition gets the position of the listener entity.
func (s *SpatialVoiceSystem) getListenerPosition(entities []*Entity) (float64, float64, bool) {
	if s.listenerEntityID == 0 {
		return 0, 0, false
	}

	for _, entity := range entities {
		if entity.ID == s.listenerEntityID {
			return s.getEntityPosition(entity)
		}
	}

	return 0, 0, false
}

// getEntityPosition gets the position of an entity from its PositionComponent.
func (s *SpatialVoiceSystem) getEntityPosition(entity *Entity) (float64, float64, bool) {
	if !entity.HasComponent("position") {
		return 0, 0, false
	}

	comp, ok := entity.GetComponent("position")
	if !ok {
		return 0, 0, false
	}
	pos, ok := comp.(*PositionComponent)
	if !ok {
		return 0, 0, false
	}

	return pos.X, pos.Y, true
}

// GetVolumeForEntity returns the current volume for an entity's voice.
func (s *SpatialVoiceSystem) GetVolumeForEntity(entity *Entity) float64 {
	if entity == nil {
		return 0
	}

	comp, ok := entity.GetComponent("spatial_voice")
	if !ok {
		return 1.0 // Default to full volume if no spatial component
	}
	spatial, ok := comp.(*SpatialVoiceComponent)
	if !ok {
		return 1.0
	}

	return spatial.CurrentVolume
}

// GetPanForEntity returns the current stereo pan for an entity's voice.
func (s *SpatialVoiceSystem) GetPanForEntity(entity *Entity) float64 {
	if entity == nil {
		return 0
	}

	comp, ok := entity.GetComponent("spatial_voice")
	if !ok {
		return 0 // Default to center if no spatial component
	}
	spatial, ok := comp.(*SpatialVoiceComponent)
	if !ok {
		return 0
	}

	return spatial.CurrentPan
}

// GetDistanceToEntity returns the distance from the listener to an entity.
func (s *SpatialVoiceSystem) GetDistanceToEntity(entity *Entity) float64 {
	if entity == nil {
		return 0
	}

	comp, ok := entity.GetComponent("spatial_voice")
	if !ok {
		return 0
	}
	spatial, ok := comp.(*SpatialVoiceComponent)
	if !ok {
		return 0
	}

	return spatial.CurrentDistance
}

// IsEntityAudible returns true if an entity's voice is within hearing range.
func (s *SpatialVoiceSystem) IsEntityAudible(entity *Entity) bool {
	if entity == nil {
		return false
	}

	comp, ok := entity.GetComponent("spatial_voice")
	if !ok {
		return true // Default to audible if no spatial component
	}
	spatial, ok := comp.(*SpatialVoiceComponent)
	if !ok {
		return true
	}

	return spatial.IsWithinRange()
}

// SetEntityRange sets the audible range for an entity's voice.
func (s *SpatialVoiceSystem) SetEntityRange(entity *Entity, minRange, maxRange float64) error {
	if entity == nil {
		return ErrNilEntity
	}

	comp, ok := entity.GetComponent("spatial_voice")
	if !ok {
		return ErrNoSpatialComponent
	}
	spatial, ok := comp.(*SpatialVoiceComponent)
	if !ok {
		return ErrNoSpatialComponent
	}

	spatial.SetRange(minRange, maxRange)
	return nil
}

// SetEntityFalloff sets the falloff curve for an entity's voice.
func (s *SpatialVoiceSystem) SetEntityFalloff(entity *Entity, curve VoiceFalloffCurve) error {
	if entity == nil {
		return ErrNilEntity
	}

	comp, ok := entity.GetComponent("spatial_voice")
	if !ok {
		return ErrNoSpatialComponent
	}
	spatial, ok := comp.(*SpatialVoiceComponent)
	if !ok {
		return ErrNoSpatialComponent
	}

	spatial.SetFalloffCurve(curve)
	return nil
}

// EnableSpatialAudio enables or disables spatial audio for an entity.
func (s *SpatialVoiceSystem) EnableSpatialAudio(entity *Entity, enabled bool) error {
	if entity == nil {
		return ErrNilEntity
	}

	comp, ok := entity.GetComponent("spatial_voice")
	if !ok {
		return ErrNoSpatialComponent
	}
	spatial, ok := comp.(*SpatialVoiceComponent)
	if !ok {
		return ErrNoSpatialComponent
	}

	spatial.SetEnabled(enabled)
	return nil
}

// GetAudibleEntities returns all entities with voice that are within audible range.
func (s *SpatialVoiceSystem) GetAudibleEntities(entities []*Entity) []*Entity {
	audible := make([]*Entity, 0, 16)

	for _, entity := range entities {
		if entity.ID == s.listenerEntityID {
			continue
		}

		if !entity.HasComponent("spatial_voice") {
			continue
		}

		if s.IsEntityAudible(entity) {
			audible = append(audible, entity)
		}
	}

	return audible
}

// Error types for spatial voice operations.
var (
	ErrNoSpatialComponent = voiceError("entity has no spatial_voice component")
)
