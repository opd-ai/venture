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
		s.updateCooldown(audio, deltaTime)

		// Process transmission state
		s.processTransmission(entity, audio, deltaTime)

		// Sync with voice channel if present
		s.syncWithVoiceChannel(entity, audio)
	}
}

// processTransmission manages the transmission state based on audio input.
func (s *VoiceAudioSystem) processTransmission(entity *Entity, audio *VoiceAudioComponent, deltaTime float64) {
	entityID := entityIDToString(entity.ID)

	if audio.InputMode == VoiceInputVoiceActivity {
		s.processVoiceActivity(entityID, audio, deltaTime)
	} else {
		s.processPushToTalk(entityID, audio)
	}
}

// processVoiceActivity handles voice activity detection with hold timer.
func (s *VoiceAudioSystem) processVoiceActivity(entityID string, audio *VoiceAudioComponent, deltaTime float64) {
	if audio.VoiceActivityDetected {
		s.handleVoiceDetected(entityID, audio)
	} else if timer, exists := s.holdTimers[entityID]; exists && timer > 0 {
		s.handleHoldPeriod(entityID, audio, deltaTime)
	} else if audio.IsTransmitting {
		s.stopTransmitting(audio, s.defaultTransmitCooldown)
	}
}

// handleVoiceDetected starts transmission and resets the hold timer.
func (s *VoiceAudioSystem) handleVoiceDetected(entityID string, audio *VoiceAudioComponent) {
	s.holdTimers[entityID] = s.voiceActivityHoldTime
	if !audio.IsTransmitting {
		s.startTransmitting(audio)
		log.WithFields(log.Fields{
			"entity_id":   entityID,
			"input_level": audio.NormalizedInputLevel,
		}).Debug("Voice activity detected, starting transmission")
	}
}

// handleHoldPeriod manages the countdown timer after voice drops.
func (s *VoiceAudioSystem) handleHoldPeriod(entityID string, audio *VoiceAudioComponent, deltaTime float64) {
	s.holdTimers[entityID] -= deltaTime
	if s.holdTimers[entityID] <= 0 {
		delete(s.holdTimers, entityID)
		if audio.IsTransmitting {
			s.stopTransmitting(audio, s.defaultTransmitCooldown)
			log.WithFields(log.Fields{
				"entity_id": entityID,
			}).Debug("Voice activity ended, stopping transmission")
		}
	}
}

// processPushToTalk handles push-to-talk mode transmission toggling.
func (s *VoiceAudioSystem) processPushToTalk(entityID string, audio *VoiceAudioComponent) {
	shouldTransmit := s.shouldTransmit(audio)

	if shouldTransmit && !audio.IsTransmitting {
		s.startTransmitting(audio)
		log.WithFields(log.Fields{
			"entity_id": entityID,
		}).Debug("Push-to-talk activated")
	} else if !shouldTransmit && audio.IsTransmitting {
		s.stopTransmitting(audio, s.defaultTransmitCooldown)
		log.WithFields(log.Fields{
			"entity_id": entityID,
		}).Debug("Push-to-talk released")
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

	s.updateInputLevel(audio, level)
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

	audio.PushToTalkActive = active
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

	audio.InputMode = mode

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

	audio.VoiceThreshold = clampVoiceFloat(threshold, 0.0, 1.0)
	return nil
}

// SetNoiseGateLevel sets the noise gate threshold for an entity.
func (s *VoiceAudioSystem) SetNoiseGateLevel(entity *Entity, level float64) error {
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

	audio.NoiseGateLevel = clampVoiceFloat(level, 0.0, 1.0)
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

	audio.OutputVolume = clampVoiceFloat(volume, 0.0, 1.0)
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

	audio.InputGain = clampVoiceFloat(gain, 0.0, 2.0)
	return nil
}

// SetPushToTalkKey sets the push-to-talk key binding for an entity.
func (s *VoiceAudioSystem) SetPushToTalkKey(entity *Entity, key string) error {
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

	audio.PushToTalkKey = key
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

// GetEffectiveOutputLevel calculates the output level for received audio.
func (s *VoiceAudioSystem) GetEffectiveOutputLevel(entity *Entity, inputLevel float64) float64 {
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

	return clampVoiceFloat(inputLevel*audio.OutputVolume, 0.0, 1.0)
}

// IsConfigured returns true if the entity's voice audio component has valid configuration.
func (s *VoiceAudioSystem) IsConfigured(entity *Entity) bool {
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

	return audio.VoiceThreshold >= 0 && audio.VoiceThreshold <= 1.0 &&
		audio.NoiseGateLevel >= 0 && audio.NoiseGateLevel <= 1.0 &&
		audio.OutputVolume >= 0 && audio.OutputVolume <= 1.0 &&
		audio.InputGain >= 0 && audio.InputGain <= 2.0
}

// StartTransmitting begins voice transmission for an entity.
func (s *VoiceAudioSystem) StartTransmitting(entity *Entity) bool {
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

	return s.startTransmitting(audio)
}

// StopTransmitting ends voice transmission for an entity with optional cooldown.
func (s *VoiceAudioSystem) StopTransmitting(entity *Entity, cooldown float64) bool {
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

	return s.stopTransmitting(audio, cooldown)
}

// Internal logic functions that operate directly on component data.

// updateInputLevel updates the current input audio level.
func (s *VoiceAudioSystem) updateInputLevel(audio *VoiceAudioComponent, level float64) {
	audio.CurrentInputLevel = clampVoiceFloat(level, 0.0, 1.0)
	audio.NormalizedInputLevel = clampVoiceFloat(level*audio.InputGain, 0.0, 1.0)

	// Check noise gate
	audio.NoiseGateOpen = audio.NormalizedInputLevel > audio.NoiseGateLevel

	// Check voice activity
	audio.VoiceActivityDetected = audio.NoiseGateOpen && audio.NormalizedInputLevel > audio.VoiceThreshold
}

// shouldTransmit returns true if audio should be transmitted based on mode and state.
func (s *VoiceAudioSystem) shouldTransmit(audio *VoiceAudioComponent) bool {
	if audio.TransmitCooldown > 0 {
		return audio.IsTransmitting // Maintain current state during cooldown
	}

	switch audio.InputMode {
	case VoiceInputPushToTalk:
		return audio.PushToTalkActive && audio.NoiseGateOpen
	case VoiceInputVoiceActivity:
		return audio.VoiceActivityDetected
	default:
		return false
	}
}

// startTransmitting begins voice transmission.
func (s *VoiceAudioSystem) startTransmitting(audio *VoiceAudioComponent) bool {
	if audio.IsTransmitting {
		return false
	}
	audio.IsTransmitting = true
	return true
}

// stopTransmitting ends voice transmission with optional cooldown.
func (s *VoiceAudioSystem) stopTransmitting(audio *VoiceAudioComponent, cooldown float64) bool {
	if !audio.IsTransmitting {
		return false
	}
	audio.IsTransmitting = false
	audio.TransmitCooldown = cooldown
	return true
}

// updateCooldown decreases the transmit cooldown by delta time.
func (s *VoiceAudioSystem) updateCooldown(audio *VoiceAudioComponent, deltaTime float64) {
	if audio.TransmitCooldown > 0 {
		audio.TransmitCooldown -= deltaTime
		if audio.TransmitCooldown < 0 {
			audio.TransmitCooldown = 0
		}
	}
}

// Error types for voice audio operations.
var (
	ErrNoAudioComponent = voiceError("entity has no voice_audio component")
)
