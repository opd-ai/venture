// Package engine provides core ECS components and systems for the Venture game.
package engine

import (
	log "github.com/sirupsen/logrus"
)

// VoiceAudioSystem processes audio input/output for voice chat.
type VoiceAudioSystem struct {
	world *World

	// defaultTransmitCooldown is the cooldown after stopping transmission.
	defaultTransmitCooldown float64

	// voiceActivityHoldTime is how long to keep transmitting after voice drops.
	voiceActivityHoldTime float64

	// holdTimers tracks remaining hold time for each entity.
	holdTimers map[string]float64
}

// NewVoiceAudioSystem creates a new voice audio system.
func NewVoiceAudioSystem(world *World) *VoiceAudioSystem {
	log.WithFields(log.Fields{
		"system_name": "voice_audio",
	}).Debug("Creating voice audio system")

	return &VoiceAudioSystem{
		world:                   world,
		defaultTransmitCooldown: 0.15,
		voiceActivityHoldTime:   0.3,
		holdTimers:              make(map[string]float64),
	}
}

// Update processes audio state for all entities with voice audio components.
func (s *VoiceAudioSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		if !entity.HasComponent("voice_audio") {
			continue
		}

		comp, ok := entity.GetComponent("voice_audio")
		if !ok {
			continue
		}
		audio, ok := comp.(*VoiceAudioComponent)
		if !ok {
			continue
		}

		// Update cooldown
		audio.UpdateCooldown(deltaTime)

		// Process transmission state
		s.processTransmission(entity, audio, deltaTime)

		// Sync with voice channel if present
		s.syncWithVoiceChannel(entity, audio)
	}
}

// processTransmission manages the transmission state based on audio input.
func (s *VoiceAudioSystem) processTransmission(entity *Entity, audio *VoiceAudioComponent, deltaTime float64) {
	entityID := entityIDToString(entity.ID)
	shouldTransmit := audio.ShouldTransmit()

	if audio.InputMode == VoiceInputVoiceActivity {
		// Handle voice activity with hold timer
		if audio.VoiceActivityDetected {
			// Voice detected, transmit and reset hold timer
			s.holdTimers[entityID] = s.voiceActivityHoldTime
			if !audio.IsTransmitting {
				audio.StartTransmitting()
				log.WithFields(log.Fields{
					"entity_id":   entityID,
					"input_level": audio.NormalizedInputLevel,
				}).Debug("Voice activity detected, starting transmission")
			}
		} else if timer, exists := s.holdTimers[entityID]; exists && timer > 0 {
			// Voice dropped but in hold period
			s.holdTimers[entityID] = timer - deltaTime
			if s.holdTimers[entityID] <= 0 {
				// Hold period expired, stop transmitting
				delete(s.holdTimers, entityID)
				if audio.IsTransmitting {
					audio.StopTransmitting(s.defaultTransmitCooldown)
					log.WithFields(log.Fields{
						"entity_id": entityID,
					}).Debug("Voice activity ended, stopping transmission")
				}
			}
		} else if audio.IsTransmitting {
			// No voice and no hold timer, stop
			audio.StopTransmitting(s.defaultTransmitCooldown)
		}
	} else {
		// Push-to-talk mode - simpler logic
		if shouldTransmit && !audio.IsTransmitting {
			audio.StartTransmitting()
			log.WithFields(log.Fields{
				"entity_id": entityID,
			}).Debug("Push-to-talk activated")
		} else if !shouldTransmit && audio.IsTransmitting {
			audio.StopTransmitting(s.defaultTransmitCooldown)
			log.WithFields(log.Fields{
				"entity_id": entityID,
			}).Debug("Push-to-talk released")
		}
	}
}

// syncWithVoiceChannel syncs audio state with the voice channel component.
func (s *VoiceAudioSystem) syncWithVoiceChannel(entity *Entity, audio *VoiceAudioComponent) {
	if !entity.HasComponent("voice_channel") {
		return
	}

	comp, ok := entity.GetComponent("voice_channel")
	if !ok {
		return
	}
	vc, ok := comp.(*VoiceChannelComponent)
	if !ok {
		return
	}

	// Update speaking state in voice channel
	vc.SetSpeaking(audio.IsTransmitting)
}

// SimulateInput simulates audio input for testing purposes.
func (s *VoiceAudioSystem) SimulateInput(entity *Entity, level float64) error {
	if entity == nil {
		return ErrNilEntity
	}

	comp, ok := entity.GetComponent("voice_audio")
	if !ok {
		return ErrNoAudioComponent
	}
	audio, ok := comp.(*VoiceAudioComponent)
	if !ok {
		return ErrNoAudioComponent
	}

	audio.UpdateInputLevel(level)
	return nil
}

// SetPushToTalk sets the push-to-talk state for an entity.
func (s *VoiceAudioSystem) SetPushToTalk(entity *Entity, active bool) error {
	if entity == nil {
		return ErrNilEntity
	}

	comp, ok := entity.GetComponent("voice_audio")
	if !ok {
		return ErrNoAudioComponent
	}
	audio, ok := comp.(*VoiceAudioComponent)
	if !ok {
		return ErrNoAudioComponent
	}

	audio.SetPushToTalk(active)
	return nil
}

// SetInputMode sets the voice input mode for an entity.
func (s *VoiceAudioSystem) SetInputMode(entity *Entity, mode VoiceInputMode) error {
	if entity == nil {
		return ErrNilEntity
	}

	comp, ok := entity.GetComponent("voice_audio")
	if !ok {
		return ErrNoAudioComponent
	}
	audio, ok := comp.(*VoiceAudioComponent)
	if !ok {
		return ErrNoAudioComponent
	}

	audio.SetInputMode(mode)

	log.WithFields(log.Fields{
		"entity_id":  entityIDToString(entity.ID),
		"input_mode": mode,
	}).Debug("Voice input mode changed")

	return nil
}

// SetVoiceThreshold sets the voice activity threshold for an entity.
func (s *VoiceAudioSystem) SetVoiceThreshold(entity *Entity, threshold float64) error {
	if entity == nil {
		return ErrNilEntity
	}

	comp, ok := entity.GetComponent("voice_audio")
	if !ok {
		return ErrNoAudioComponent
	}
	audio, ok := comp.(*VoiceAudioComponent)
	if !ok {
		return ErrNoAudioComponent
	}

	audio.SetVoiceThreshold(threshold)
	return nil
}

// SetOutputVolume sets the output volume for an entity.
func (s *VoiceAudioSystem) SetOutputVolume(entity *Entity, volume float64) error {
	if entity == nil {
		return ErrNilEntity
	}

	comp, ok := entity.GetComponent("voice_audio")
	if !ok {
		return ErrNoAudioComponent
	}
	audio, ok := comp.(*VoiceAudioComponent)
	if !ok {
		return ErrNoAudioComponent
	}

	audio.SetOutputVolume(volume)
	return nil
}

// SetInputGain sets the input gain for an entity.
func (s *VoiceAudioSystem) SetInputGain(entity *Entity, gain float64) error {
	if entity == nil {
		return ErrNilEntity
	}

	comp, ok := entity.GetComponent("voice_audio")
	if !ok {
		return ErrNoAudioComponent
	}
	audio, ok := comp.(*VoiceAudioComponent)
	if !ok {
		return ErrNoAudioComponent
	}

	audio.SetInputGain(gain)
	return nil
}

// IsTransmitting returns true if the entity is currently transmitting voice.
func (s *VoiceAudioSystem) IsTransmitting(entity *Entity) bool {
	if entity == nil {
		return false
	}

	comp, ok := entity.GetComponent("voice_audio")
	if !ok {
		return false
	}
	audio, ok := comp.(*VoiceAudioComponent)
	if !ok {
		return false
	}

	return audio.IsTransmitting
}

// GetInputLevel returns the current normalized input level for an entity.
func (s *VoiceAudioSystem) GetInputLevel(entity *Entity) float64 {
	if entity == nil {
		return 0
	}

	comp, ok := entity.GetComponent("voice_audio")
	if !ok {
		return 0
	}
	audio, ok := comp.(*VoiceAudioComponent)
	if !ok {
		return 0
	}

	return audio.NormalizedInputLevel
}

// Error types for voice audio operations.
var (
	ErrNoAudioComponent = voiceError("entity has no voice_audio component")
)
