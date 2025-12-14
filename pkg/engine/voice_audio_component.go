// Package engine provides core ECS components and systems for the Venture game.
package engine

import (
	"encoding/json"
)

// VoiceInputMode defines how voice input is activated.
type VoiceInputMode string

const (
	// VoiceInputPushToTalk requires holding a key to transmit.
	VoiceInputPushToTalk VoiceInputMode = "push_to_talk"
	// VoiceInputVoiceActivity automatically transmits when speaking.
	VoiceInputVoiceActivity VoiceInputMode = "voice_activity"
)

// VoiceAudioComponent tracks audio processing state for voice chat.
type VoiceAudioComponent struct {
	// InputMode determines how voice input is activated.
	InputMode VoiceInputMode `json:"input_mode"`

	// PushToTalkKey is the key binding for push-to-talk mode.
	PushToTalkKey string `json:"push_to_talk_key"`

	// PushToTalkActive indicates if push-to-talk key is currently held.
	PushToTalkActive bool `json:"push_to_talk_active"`

	// VoiceThreshold is the sensitivity for voice activity detection (0.0-1.0).
	VoiceThreshold float64 `json:"voice_threshold"`

	// NoiseGateLevel is the noise gate threshold (0.0-1.0).
	NoiseGateLevel float64 `json:"noise_gate_level"`

	// OutputVolume is the master output volume for received voice (0.0-1.0).
	OutputVolume float64 `json:"output_volume"`

	// InputGain is the microphone input gain (0.0-2.0, 1.0 = normal).
	InputGain float64 `json:"input_gain"`

	// IsTransmitting indicates if currently sending voice data.
	IsTransmitting bool `json:"is_transmitting"`

	// CurrentInputLevel is the current input audio level (0.0-1.0).
	CurrentInputLevel float64 `json:"current_input_level"`

	// NormalizedInputLevel is the input level after gain and normalization.
	NormalizedInputLevel float64 `json:"normalized_input_level"`

	// VoiceActivityDetected indicates if voice was detected above threshold.
	VoiceActivityDetected bool `json:"voice_activity_detected"`

	// NoiseGateOpen indicates if audio is passing through the noise gate.
	NoiseGateOpen bool `json:"noise_gate_open"`

	// TransmitCooldown prevents rapid on/off cycling (seconds remaining).
	TransmitCooldown float64 `json:"transmit_cooldown"`
}

// Type returns the component type identifier.
func (c *VoiceAudioComponent) Type() string {
	return "voice_audio"
}

// NewVoiceAudioComponent creates a new voice audio component with default values.
func NewVoiceAudioComponent() *VoiceAudioComponent {
	return &VoiceAudioComponent{
		InputMode:             VoiceInputPushToTalk,
		PushToTalkKey:         "V",
		PushToTalkActive:      false,
		VoiceThreshold:        0.1,
		NoiseGateLevel:        0.05,
		OutputVolume:          1.0,
		InputGain:             1.0,
		IsTransmitting:        false,
		CurrentInputLevel:     0.0,
		NormalizedInputLevel:  0.0,
		VoiceActivityDetected: false,
		NoiseGateOpen:         false,
		TransmitCooldown:      0.0,
	}
}

// NewVoiceActivityComponent creates a voice audio component in voice activity mode.
func NewVoiceActivityComponent(threshold float64) *VoiceAudioComponent {
	c := NewVoiceAudioComponent()
	c.InputMode = VoiceInputVoiceActivity
	c.VoiceThreshold = clampVoiceFloat(threshold, 0.0, 1.0)
	return c
}

// clampVoiceFloat clamps a float64 value between min and max.
func clampVoiceFloat(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// SetInputMode sets the voice input mode.
func (c *VoiceAudioComponent) SetInputMode(mode VoiceInputMode) {
	c.InputMode = mode
}

// SetPushToTalkKey sets the push-to-talk key binding.
func (c *VoiceAudioComponent) SetPushToTalkKey(key string) {
	c.PushToTalkKey = key
}

// SetVoiceThreshold sets the voice activity detection threshold.
func (c *VoiceAudioComponent) SetVoiceThreshold(threshold float64) {
	c.VoiceThreshold = clampVoiceFloat(threshold, 0.0, 1.0)
}

// SetNoiseGateLevel sets the noise gate threshold.
func (c *VoiceAudioComponent) SetNoiseGateLevel(level float64) {
	c.NoiseGateLevel = clampVoiceFloat(level, 0.0, 1.0)
}

// SetOutputVolume sets the output volume.
func (c *VoiceAudioComponent) SetOutputVolume(volume float64) {
	c.OutputVolume = clampVoiceFloat(volume, 0.0, 1.0)
}

// SetInputGain sets the input gain.
func (c *VoiceAudioComponent) SetInputGain(gain float64) {
	c.InputGain = clampVoiceFloat(gain, 0.0, 2.0)
}

// UpdateInputLevel updates the current input audio level.
func (c *VoiceAudioComponent) UpdateInputLevel(level float64) {
	c.CurrentInputLevel = clampVoiceFloat(level, 0.0, 1.0)
	c.NormalizedInputLevel = clampVoiceFloat(level*c.InputGain, 0.0, 1.0)

	// Check noise gate
	c.NoiseGateOpen = c.NormalizedInputLevel > c.NoiseGateLevel

	// Check voice activity
	c.VoiceActivityDetected = c.NoiseGateOpen && c.NormalizedInputLevel > c.VoiceThreshold
}

// ShouldTransmit returns true if audio should be transmitted based on mode and state.
func (c *VoiceAudioComponent) ShouldTransmit() bool {
	if c.TransmitCooldown > 0 {
		return c.IsTransmitting // Maintain current state during cooldown
	}

	switch c.InputMode {
	case VoiceInputPushToTalk:
		return c.PushToTalkActive && c.NoiseGateOpen
	case VoiceInputVoiceActivity:
		return c.VoiceActivityDetected
	default:
		return false
	}
}

// SetPushToTalk sets the push-to-talk state (key pressed/released).
func (c *VoiceAudioComponent) SetPushToTalk(active bool) {
	c.PushToTalkActive = active
}

// StartTransmitting begins voice transmission.
func (c *VoiceAudioComponent) StartTransmitting() bool {
	if c.IsTransmitting {
		return false
	}
	c.IsTransmitting = true
	return true
}

// StopTransmitting ends voice transmission with optional cooldown.
func (c *VoiceAudioComponent) StopTransmitting(cooldown float64) bool {
	if !c.IsTransmitting {
		return false
	}
	c.IsTransmitting = false
	c.TransmitCooldown = cooldown
	return true
}

// UpdateCooldown decreases the transmit cooldown by delta time.
func (c *VoiceAudioComponent) UpdateCooldown(deltaTime float64) {
	if c.TransmitCooldown > 0 {
		c.TransmitCooldown -= deltaTime
		if c.TransmitCooldown < 0 {
			c.TransmitCooldown = 0
		}
	}
}

// GetEffectiveOutputLevel calculates the output level for received audio.
func (c *VoiceAudioComponent) GetEffectiveOutputLevel(inputLevel float64) float64 {
	return clampVoiceFloat(inputLevel*c.OutputVolume, 0.0, 1.0)
}

// IsConfigured returns true if the component has valid configuration.
func (c *VoiceAudioComponent) IsConfigured() bool {
	return c.VoiceThreshold >= 0 && c.VoiceThreshold <= 1.0 &&
		c.NoiseGateLevel >= 0 && c.NoiseGateLevel <= 1.0 &&
		c.OutputVolume >= 0 && c.OutputVolume <= 1.0 &&
		c.InputGain >= 0 && c.InputGain <= 2.0
}

// Serialize converts the component to JSON bytes.
func (c *VoiceAudioComponent) Serialize() ([]byte, error) {
	return json.Marshal(c)
}

// Deserialize loads the component from JSON bytes.
func (c *VoiceAudioComponent) Deserialize(data []byte) error {
	return json.Unmarshal(data, c)
}
