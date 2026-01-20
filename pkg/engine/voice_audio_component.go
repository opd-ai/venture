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

// Serialize converts the component to JSON bytes.
func (c *VoiceAudioComponent) Serialize() ([]byte, error) {
	return json.Marshal(c)
}

// Deserialize loads the component from JSON bytes.
func (c *VoiceAudioComponent) Deserialize(data []byte) error {
	return json.Unmarshal(data, c)
}
