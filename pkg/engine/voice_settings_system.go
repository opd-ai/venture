// Package engine provides core ECS components and systems for the Venture game.
package engine

import (
	log "github.com/sirupsen/logrus"
)

// VoiceSettingsSystem applies user voice settings and handles device changes.
type VoiceSettingsSystem struct {
	world *World

	// speakingTimeout is how long (seconds) before speaking indicator times out.
	speakingTimeout float64

	// speakingTimers tracks when each entity last was marked speaking.
	speakingTimers map[string]float64
}

// NewVoiceSettingsSystem creates a new voice settings system.
func NewVoiceSettingsSystem(world *World) *VoiceSettingsSystem {
	log.WithFields(log.Fields{
		"system_name": "voice_settings",
	}).Debug("Creating voice settings system")

	return &VoiceSettingsSystem{
		world:           world,
		speakingTimeout: 0.5,
		speakingTimers:  make(map[string]float64),
	}
}

// Update processes voice settings for all entities.
func (s *VoiceSettingsSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		if !entity.HasComponent("voice_settings") {
			continue
		}

		comp, ok := entity.GetComponent("voice_settings")
		if !ok {
			continue
		}
		settings, ok := comp.(*VoiceSettingsComponent)
		if !ok {
			continue
		}

		// Update speaking indicators with timeout
		s.updateSpeakingIndicators(settings, deltaTime)
	}
}

// updateSpeakingIndicators updates speaking user timeouts.
func (s *VoiceSettingsSystem) updateSpeakingIndicators(settings *VoiceSettingsComponent, deltaTime float64) {
	for entityID := range settings.SpeakingUsers {
		if timer, exists := s.speakingTimers[entityID]; exists {
			timer += deltaTime
			if timer > s.speakingTimeout {
				settings.UpdateSpeakingUser(entityID, false)
				delete(s.speakingTimers, entityID)
			} else {
				s.speakingTimers[entityID] = timer
			}
		}
	}
}

// MarkSpeaking marks a user as currently speaking.
func (s *VoiceSettingsSystem) MarkSpeaking(settingsEntity *Entity, speakerEntityID string) error {
	if settingsEntity == nil {
		return ErrNilEntity
	}

	comp, ok := settingsEntity.GetComponent("voice_settings")
	if !ok {
		return ErrNoSettingsComponent
	}
	settings, ok := comp.(*VoiceSettingsComponent)
	if !ok {
		return ErrNoSettingsComponent
	}

	settings.UpdateSpeakingUser(speakerEntityID, true)
	s.speakingTimers[speakerEntityID] = 0 // Reset timeout

	return nil
}

// MarkNotSpeaking marks a user as no longer speaking.
func (s *VoiceSettingsSystem) MarkNotSpeaking(settingsEntity *Entity, speakerEntityID string) error {
	if settingsEntity == nil {
		return ErrNilEntity
	}

	comp, ok := settingsEntity.GetComponent("voice_settings")
	if !ok {
		return ErrNoSettingsComponent
	}
	settings, ok := comp.(*VoiceSettingsComponent)
	if !ok {
		return ErrNoSettingsComponent
	}

	settings.UpdateSpeakingUser(speakerEntityID, false)
	delete(s.speakingTimers, speakerEntityID)

	return nil
}

// SetMasterVolume sets the master volume for an entity's voice settings.
func (s *VoiceSettingsSystem) SetMasterVolume(entity *Entity, volume float64) error {
	if entity == nil {
		return ErrNilEntity
	}

	comp, ok := entity.GetComponent("voice_settings")
	if !ok {
		return ErrNoSettingsComponent
	}
	settings, ok := comp.(*VoiceSettingsComponent)
	if !ok {
		return ErrNoSettingsComponent
	}

	settings.SetMasterVolume(volume)

	log.WithFields(log.Fields{
		"entity_id": entityIDToString(entity.ID),
		"volume":    volume,
	}).Debug("Master volume changed")

	return nil
}

// MuteUser mutes a specific user for the settings entity.
func (s *VoiceSettingsSystem) MuteUser(settingsEntity *Entity, targetEntityID string) error {
	if settingsEntity == nil {
		return ErrNilEntity
	}

	comp, ok := settingsEntity.GetComponent("voice_settings")
	if !ok {
		return ErrNoSettingsComponent
	}
	settings, ok := comp.(*VoiceSettingsComponent)
	if !ok {
		return ErrNoSettingsComponent
	}

	settings.MuteUser(targetEntityID)

	log.WithFields(log.Fields{
		"entity_id": entityIDToString(settingsEntity.ID),
		"muted_id":  targetEntityID,
	}).Debug("User muted")

	return nil
}

// UnmuteUser unmutes a specific user for the settings entity.
func (s *VoiceSettingsSystem) UnmuteUser(settingsEntity *Entity, targetEntityID string) error {
	if settingsEntity == nil {
		return ErrNilEntity
	}

	comp, ok := settingsEntity.GetComponent("voice_settings")
	if !ok {
		return ErrNoSettingsComponent
	}
	settings, ok := comp.(*VoiceSettingsComponent)
	if !ok {
		return ErrNoSettingsComponent
	}

	settings.UnmuteUser(targetEntityID)

	log.WithFields(log.Fields{
		"entity_id":  entityIDToString(settingsEntity.ID),
		"unmuted_id": targetEntityID,
	}).Debug("User unmuted")

	return nil
}

// SetUserVolume sets a per-user volume for the settings entity.
func (s *VoiceSettingsSystem) SetUserVolume(settingsEntity *Entity, targetEntityID string, volume float64) error {
	if settingsEntity == nil {
		return ErrNilEntity
	}

	comp, ok := settingsEntity.GetComponent("voice_settings")
	if !ok {
		return ErrNoSettingsComponent
	}
	settings, ok := comp.(*VoiceSettingsComponent)
	if !ok {
		return ErrNoSettingsComponent
	}

	settings.SetUserVolume(targetEntityID, volume)
	return nil
}

// SetInputDevice sets the input device for an entity.
func (s *VoiceSettingsSystem) SetInputDevice(entity *Entity, device string) error {
	if entity == nil {
		return ErrNilEntity
	}

	comp, ok := entity.GetComponent("voice_settings")
	if !ok {
		return ErrNoSettingsComponent
	}
	settings, ok := comp.(*VoiceSettingsComponent)
	if !ok {
		return ErrNoSettingsComponent
	}

	settings.SetInputDevice(device)

	log.WithFields(log.Fields{
		"entity_id": entityIDToString(entity.ID),
		"device":    device,
	}).Debug("Input device changed")

	return nil
}

// SetOutputDevice sets the output device for an entity.
func (s *VoiceSettingsSystem) SetOutputDevice(entity *Entity, device string) error {
	if entity == nil {
		return ErrNilEntity
	}

	comp, ok := entity.GetComponent("voice_settings")
	if !ok {
		return ErrNoSettingsComponent
	}
	settings, ok := comp.(*VoiceSettingsComponent)
	if !ok {
		return ErrNoSettingsComponent
	}

	settings.SetOutputDevice(device)

	log.WithFields(log.Fields{
		"entity_id": entityIDToString(entity.ID),
		"device":    device,
	}).Debug("Output device changed")

	return nil
}

// GetSpeakingUsers returns the list of currently speaking users for an entity.
func (s *VoiceSettingsSystem) GetSpeakingUsers(entity *Entity) []string {
	if entity == nil {
		return nil
	}

	comp, ok := entity.GetComponent("voice_settings")
	if !ok {
		return nil
	}
	settings, ok := comp.(*VoiceSettingsComponent)
	if !ok {
		return nil
	}

	return settings.GetSpeakingUsers()
}

// GetEffectiveVolume calculates the effective volume for a speaker.
func (s *VoiceSettingsSystem) GetEffectiveVolume(settingsEntity *Entity, speakerEntityID string, baseVolume float64) float64 {
	if settingsEntity == nil {
		return baseVolume
	}

	comp, ok := settingsEntity.GetComponent("voice_settings")
	if !ok {
		return baseVolume
	}
	settings, ok := comp.(*VoiceSettingsComponent)
	if !ok {
		return baseVolume
	}

	return settings.GetEffectiveVolume(speakerEntityID, baseVolume)
}

// IsUserMuted checks if a user is muted in the settings entity.
func (s *VoiceSettingsSystem) IsUserMuted(settingsEntity *Entity, targetEntityID string) bool {
	if settingsEntity == nil {
		return false
	}

	comp, ok := settingsEntity.GetComponent("voice_settings")
	if !ok {
		return false
	}
	settings, ok := comp.(*VoiceSettingsComponent)
	if !ok {
		return false
	}

	return settings.IsUserMuted(targetEntityID)
}

// Error types for voice settings operations.
var (
	ErrNoSettingsComponent = voiceError("entity has no voice_settings component")
)
