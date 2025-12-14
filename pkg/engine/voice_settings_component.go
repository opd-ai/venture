// Package engine provides core ECS components and systems for the Venture game.
package engine

import (
	"encoding/json"
)

// VoiceSettingsComponent stores user-configurable voice preferences.
type VoiceSettingsComponent struct {
	// MasterVolume is the overall voice chat volume (0.0-1.0).
	MasterVolume float64 `json:"master_volume"`

	// MutedUsers maps entity IDs to their mute status (true = muted).
	MutedUsers map[string]bool `json:"muted_users"`

	// UserVolumes maps entity IDs to per-user volume adjustments (0.0-2.0).
	UserVolumes map[string]float64 `json:"user_volumes"`

	// InputDevice is the selected input device name (empty = default).
	InputDevice string `json:"input_device"`

	// OutputDevice is the selected output device name (empty = default).
	OutputDevice string `json:"output_device"`

	// InputSensitivity is the microphone sensitivity (0.0-1.0).
	InputSensitivity float64 `json:"input_sensitivity"`

	// SpeakingIndicatorEnabled shows when others are speaking.
	SpeakingIndicatorEnabled bool `json:"speaking_indicator_enabled"`

	// SpeakingUsers tracks entity IDs currently speaking (for UI).
	SpeakingUsers map[string]bool `json:"speaking_users"`

	// ProximityEnabled enables spatial/proximity voice chat.
	ProximityEnabled bool `json:"proximity_enabled"`

	// EchoCancellation enables echo cancellation.
	EchoCancellation bool `json:"echo_cancellation"`

	// NoiseSuppression enables noise suppression.
	NoiseSuppression bool `json:"noise_suppression"`

	// AutoGainControl enables automatic gain control.
	AutoGainControl bool `json:"auto_gain_control"`
}

// Type returns the component type identifier.
func (c *VoiceSettingsComponent) Type() string {
	return "voice_settings"
}

// NewVoiceSettingsComponent creates a new voice settings component with defaults.
func NewVoiceSettingsComponent() *VoiceSettingsComponent {
	return &VoiceSettingsComponent{
		MasterVolume:             1.0,
		MutedUsers:               make(map[string]bool),
		UserVolumes:              make(map[string]float64),
		InputDevice:              "",
		OutputDevice:             "",
		InputSensitivity:         0.5,
		SpeakingIndicatorEnabled: true,
		SpeakingUsers:            make(map[string]bool),
		ProximityEnabled:         true,
		EchoCancellation:         true,
		NoiseSuppression:         true,
		AutoGainControl:          true,
	}
}

// SetMasterVolume sets the master voice volume.
func (c *VoiceSettingsComponent) SetMasterVolume(volume float64) {
	c.MasterVolume = clampVoiceFloat(volume, 0.0, 1.0)
}

// MuteUser mutes a specific user.
func (c *VoiceSettingsComponent) MuteUser(entityID string) {
	c.MutedUsers[entityID] = true
}

// UnmuteUser unmutes a specific user.
func (c *VoiceSettingsComponent) UnmuteUser(entityID string) {
	delete(c.MutedUsers, entityID)
}

// IsUserMuted returns true if the user is muted.
func (c *VoiceSettingsComponent) IsUserMuted(entityID string) bool {
	return c.MutedUsers[entityID]
}

// ToggleUserMute toggles the mute status for a user.
func (c *VoiceSettingsComponent) ToggleUserMute(entityID string) bool {
	if c.MutedUsers[entityID] {
		c.UnmuteUser(entityID)
		return false
	}
	c.MuteUser(entityID)
	return true
}

// SetUserVolume sets a per-user volume adjustment.
func (c *VoiceSettingsComponent) SetUserVolume(entityID string, volume float64) {
	c.UserVolumes[entityID] = clampVoiceFloat(volume, 0.0, 2.0)
}

// GetUserVolume returns the volume adjustment for a user (default 1.0).
func (c *VoiceSettingsComponent) GetUserVolume(entityID string) float64 {
	if vol, exists := c.UserVolumes[entityID]; exists {
		return vol
	}
	return 1.0
}

// ResetUserVolume resets a user's volume to default.
func (c *VoiceSettingsComponent) ResetUserVolume(entityID string) {
	delete(c.UserVolumes, entityID)
}

// GetEffectiveVolume calculates the effective volume for a user.
func (c *VoiceSettingsComponent) GetEffectiveVolume(entityID string, baseVolume float64) float64 {
	if c.IsUserMuted(entityID) {
		return 0
	}
	userVol := c.GetUserVolume(entityID)
	return clampVoiceFloat(baseVolume*c.MasterVolume*userVol, 0.0, 1.0)
}

// SetInputDevice sets the input device name.
func (c *VoiceSettingsComponent) SetInputDevice(device string) {
	c.InputDevice = device
}

// SetOutputDevice sets the output device name.
func (c *VoiceSettingsComponent) SetOutputDevice(device string) {
	c.OutputDevice = device
}

// SetInputSensitivity sets the microphone sensitivity.
func (c *VoiceSettingsComponent) SetInputSensitivity(sensitivity float64) {
	c.InputSensitivity = clampVoiceFloat(sensitivity, 0.0, 1.0)
}

// UpdateSpeakingUser updates whether a user is currently speaking.
func (c *VoiceSettingsComponent) UpdateSpeakingUser(entityID string, speaking bool) {
	if speaking {
		c.SpeakingUsers[entityID] = true
	} else {
		delete(c.SpeakingUsers, entityID)
	}
}

// GetSpeakingUsers returns a list of entity IDs currently speaking.
func (c *VoiceSettingsComponent) GetSpeakingUsers() []string {
	speaking := make([]string, 0, len(c.SpeakingUsers))
	for id := range c.SpeakingUsers {
		speaking = append(speaking, id)
	}
	return speaking
}

// GetSpeakingUserCount returns the number of users currently speaking.
func (c *VoiceSettingsComponent) GetSpeakingUserCount() int {
	return len(c.SpeakingUsers)
}

// ClearSpeakingUsers clears all speaking indicators.
func (c *VoiceSettingsComponent) ClearSpeakingUsers() {
	c.SpeakingUsers = make(map[string]bool)
}

// SetProximityEnabled enables or disables proximity voice.
func (c *VoiceSettingsComponent) SetProximityEnabled(enabled bool) {
	c.ProximityEnabled = enabled
}

// SetEchoCancellation enables or disables echo cancellation.
func (c *VoiceSettingsComponent) SetEchoCancellation(enabled bool) {
	c.EchoCancellation = enabled
}

// SetNoiseSuppression enables or disables noise suppression.
func (c *VoiceSettingsComponent) SetNoiseSuppression(enabled bool) {
	c.NoiseSuppression = enabled
}

// SetAutoGainControl enables or disables automatic gain control.
func (c *VoiceSettingsComponent) SetAutoGainControl(enabled bool) {
	c.AutoGainControl = enabled
}

// GetMutedUserCount returns the number of muted users.
func (c *VoiceSettingsComponent) GetMutedUserCount() int {
	return len(c.MutedUsers)
}

// GetMutedUsers returns a list of muted user IDs.
func (c *VoiceSettingsComponent) GetMutedUsers() []string {
	muted := make([]string, 0, len(c.MutedUsers))
	for id := range c.MutedUsers {
		muted = append(muted, id)
	}
	return muted
}

// Serialize converts the component to JSON bytes.
func (c *VoiceSettingsComponent) Serialize() ([]byte, error) {
	return json.Marshal(c)
}

// Deserialize loads the component from JSON bytes.
func (c *VoiceSettingsComponent) Deserialize(data []byte) error {
	return json.Unmarshal(data, c)
}
