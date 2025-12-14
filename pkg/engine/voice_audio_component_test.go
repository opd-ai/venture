package engine

import (
	"testing"
)

func TestVoiceAudioComponent_Type(t *testing.T) {
	c := NewVoiceAudioComponent()
	if c.Type() != "voice_audio" {
		t.Errorf("expected type 'voice_audio', got '%s'", c.Type())
	}
}

func TestNewVoiceAudioComponent(t *testing.T) {
	c := NewVoiceAudioComponent()

	if c.InputMode != VoiceInputPushToTalk {
		t.Errorf("expected InputMode push_to_talk, got %s", c.InputMode)
	}
	if c.PushToTalkKey != "V" {
		t.Errorf("expected PushToTalkKey 'V', got '%s'", c.PushToTalkKey)
	}
	if c.VoiceThreshold != 0.1 {
		t.Errorf("expected VoiceThreshold 0.1, got %f", c.VoiceThreshold)
	}
	if c.NoiseGateLevel != 0.05 {
		t.Errorf("expected NoiseGateLevel 0.05, got %f", c.NoiseGateLevel)
	}
	if c.OutputVolume != 1.0 {
		t.Errorf("expected OutputVolume 1.0, got %f", c.OutputVolume)
	}
	if c.InputGain != 1.0 {
		t.Errorf("expected InputGain 1.0, got %f", c.InputGain)
	}
	if c.IsTransmitting {
		t.Error("expected IsTransmitting to be false")
	}
}

func TestNewVoiceActivityComponent(t *testing.T) {
	tests := []struct {
		name          string
		threshold     float64
		wantThreshold float64
	}{
		{"normal", 0.2, 0.2},
		{"clamped low", -0.5, 0.0},
		{"clamped high", 1.5, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewVoiceActivityComponent(tt.threshold)

			if c.InputMode != VoiceInputVoiceActivity {
				t.Errorf("expected InputMode voice_activity, got %s", c.InputMode)
			}
			if c.VoiceThreshold != tt.wantThreshold {
				t.Errorf("expected VoiceThreshold %f, got %f", tt.wantThreshold, c.VoiceThreshold)
			}
		})
	}
}

func TestClampVoiceFloat(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		min      float64
		max      float64
		expected float64
	}{
		{"within range", 0.5, 0.0, 1.0, 0.5},
		{"at min", 0.0, 0.0, 1.0, 0.0},
		{"at max", 1.0, 0.0, 1.0, 1.0},
		{"below min", -0.5, 0.0, 1.0, 0.0},
		{"above max", 1.5, 0.0, 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := clampVoiceFloat(tt.value, tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestVoiceAudioComponent_SetMethods(t *testing.T) {
	c := NewVoiceAudioComponent()

	c.SetInputMode(VoiceInputVoiceActivity)
	if c.InputMode != VoiceInputVoiceActivity {
		t.Errorf("expected InputMode voice_activity, got %s", c.InputMode)
	}

	c.SetPushToTalkKey("T")
	if c.PushToTalkKey != "T" {
		t.Errorf("expected PushToTalkKey 'T', got '%s'", c.PushToTalkKey)
	}

	c.SetVoiceThreshold(0.3)
	if c.VoiceThreshold != 0.3 {
		t.Errorf("expected VoiceThreshold 0.3, got %f", c.VoiceThreshold)
	}

	c.SetVoiceThreshold(2.0) // Should be clamped
	if c.VoiceThreshold != 1.0 {
		t.Errorf("expected VoiceThreshold 1.0 (clamped), got %f", c.VoiceThreshold)
	}

	c.SetNoiseGateLevel(0.1)
	if c.NoiseGateLevel != 0.1 {
		t.Errorf("expected NoiseGateLevel 0.1, got %f", c.NoiseGateLevel)
	}

	c.SetOutputVolume(0.5)
	if c.OutputVolume != 0.5 {
		t.Errorf("expected OutputVolume 0.5, got %f", c.OutputVolume)
	}

	c.SetInputGain(1.5)
	if c.InputGain != 1.5 {
		t.Errorf("expected InputGain 1.5, got %f", c.InputGain)
	}

	c.SetInputGain(3.0) // Should be clamped to 2.0
	if c.InputGain != 2.0 {
		t.Errorf("expected InputGain 2.0 (clamped), got %f", c.InputGain)
	}
}

func TestVoiceAudioComponent_UpdateInputLevel(t *testing.T) {
	c := NewVoiceAudioComponent()
	c.SetVoiceThreshold(0.2)
	c.SetNoiseGateLevel(0.1)
	c.SetInputGain(1.0)

	// Level below noise gate
	c.UpdateInputLevel(0.05)
	if c.NoiseGateOpen {
		t.Error("expected NoiseGateOpen to be false for level below gate")
	}
	if c.VoiceActivityDetected {
		t.Error("expected VoiceActivityDetected to be false")
	}

	// Level above noise gate but below voice threshold
	c.UpdateInputLevel(0.15)
	if !c.NoiseGateOpen {
		t.Error("expected NoiseGateOpen to be true")
	}
	if c.VoiceActivityDetected {
		t.Error("expected VoiceActivityDetected to be false (below threshold)")
	}

	// Level above voice threshold
	c.UpdateInputLevel(0.5)
	if !c.NoiseGateOpen {
		t.Error("expected NoiseGateOpen to be true")
	}
	if !c.VoiceActivityDetected {
		t.Error("expected VoiceActivityDetected to be true")
	}

	// Test input gain effect
	c.SetInputGain(2.0)
	c.UpdateInputLevel(0.3)
	if c.NormalizedInputLevel != 0.6 {
		t.Errorf("expected NormalizedInputLevel 0.6, got %f", c.NormalizedInputLevel)
	}

	// Test clamping on high input
	c.UpdateInputLevel(0.8) // 0.8 * 2.0 = 1.6, should clamp to 1.0
	if c.NormalizedInputLevel != 1.0 {
		t.Errorf("expected NormalizedInputLevel 1.0 (clamped), got %f", c.NormalizedInputLevel)
	}
}

func TestVoiceAudioComponent_ShouldTransmit_PushToTalk(t *testing.T) {
	c := NewVoiceAudioComponent()
	c.SetInputMode(VoiceInputPushToTalk)
	c.SetNoiseGateLevel(0.1)

	// Not pressing PTT, should not transmit
	c.UpdateInputLevel(0.5)
	if c.ShouldTransmit() {
		t.Error("expected ShouldTransmit false when PTT not active")
	}

	// Pressing PTT but below noise gate
	c.SetPushToTalk(true)
	c.UpdateInputLevel(0.05)
	if c.ShouldTransmit() {
		t.Error("expected ShouldTransmit false when below noise gate")
	}

	// Pressing PTT and above noise gate
	c.UpdateInputLevel(0.5)
	if !c.ShouldTransmit() {
		t.Error("expected ShouldTransmit true when PTT active and above gate")
	}

	// Release PTT
	c.SetPushToTalk(false)
	if c.ShouldTransmit() {
		t.Error("expected ShouldTransmit false when PTT released")
	}
}

func TestVoiceAudioComponent_ShouldTransmit_VoiceActivity(t *testing.T) {
	c := NewVoiceActivityComponent(0.2)
	c.SetNoiseGateLevel(0.1)

	// Below threshold
	c.UpdateInputLevel(0.15)
	if c.ShouldTransmit() {
		t.Error("expected ShouldTransmit false when below threshold")
	}

	// Above threshold
	c.UpdateInputLevel(0.5)
	if !c.ShouldTransmit() {
		t.Error("expected ShouldTransmit true when above threshold")
	}
}

func TestVoiceAudioComponent_ShouldTransmit_Cooldown(t *testing.T) {
	c := NewVoiceAudioComponent()
	c.SetInputMode(VoiceInputPushToTalk)
	c.SetNoiseGateLevel(0.1)

	// Start transmitting
	c.SetPushToTalk(true)
	c.UpdateInputLevel(0.5)
	c.StartTransmitting()

	// Stop with cooldown
	c.StopTransmitting(0.5)

	// During cooldown, should maintain false state (not transmitting)
	c.SetPushToTalk(true)
	c.UpdateInputLevel(0.5)
	if c.ShouldTransmit() {
		t.Error("expected ShouldTransmit to return current state during cooldown")
	}

	// After cooldown expires
	c.UpdateCooldown(0.6)
	if !c.ShouldTransmit() {
		t.Error("expected ShouldTransmit true after cooldown expires")
	}
}

func TestVoiceAudioComponent_TransmitState(t *testing.T) {
	c := NewVoiceAudioComponent()

	// Start transmitting
	if !c.StartTransmitting() {
		t.Error("expected StartTransmitting to return true on first call")
	}
	if !c.IsTransmitting {
		t.Error("expected IsTransmitting to be true")
	}

	// Start again (already transmitting)
	if c.StartTransmitting() {
		t.Error("expected StartTransmitting to return false when already transmitting")
	}

	// Stop transmitting
	if !c.StopTransmitting(0.2) {
		t.Error("expected StopTransmitting to return true")
	}
	if c.IsTransmitting {
		t.Error("expected IsTransmitting to be false")
	}
	if c.TransmitCooldown != 0.2 {
		t.Errorf("expected TransmitCooldown 0.2, got %f", c.TransmitCooldown)
	}

	// Stop again (not transmitting)
	if c.StopTransmitting(0.1) {
		t.Error("expected StopTransmitting to return false when not transmitting")
	}
}

func TestVoiceAudioComponent_UpdateCooldown(t *testing.T) {
	c := NewVoiceAudioComponent()
	c.TransmitCooldown = 0.5

	c.UpdateCooldown(0.2)
	if c.TransmitCooldown != 0.3 {
		t.Errorf("expected TransmitCooldown 0.3, got %f", c.TransmitCooldown)
	}

	c.UpdateCooldown(0.5) // Should clamp to 0
	if c.TransmitCooldown != 0 {
		t.Errorf("expected TransmitCooldown 0, got %f", c.TransmitCooldown)
	}

	// No cooldown, should remain 0
	c.UpdateCooldown(0.1)
	if c.TransmitCooldown != 0 {
		t.Errorf("expected TransmitCooldown 0, got %f", c.TransmitCooldown)
	}
}

func TestVoiceAudioComponent_GetEffectiveOutputLevel(t *testing.T) {
	c := NewVoiceAudioComponent()

	tests := []struct {
		name       string
		inputLevel float64
		volume     float64
		expected   float64
	}{
		{"full volume", 0.5, 1.0, 0.5},
		{"half volume", 0.5, 0.5, 0.25},
		{"muted", 0.5, 0.0, 0.0},
		{"high input clamped", 1.2, 1.0, 1.0}, // Input 1.2 * 1.0 = 1.2, clamped to 1.0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.SetOutputVolume(tt.volume)
			result := c.GetEffectiveOutputLevel(tt.inputLevel)
			if result != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestVoiceAudioComponent_IsConfigured(t *testing.T) {
	c := NewVoiceAudioComponent()

	if !c.IsConfigured() {
		t.Error("expected default component to be configured")
	}

	// Test invalid configurations
	c.VoiceThreshold = -0.1
	if c.IsConfigured() {
		t.Error("expected IsConfigured false for negative VoiceThreshold")
	}
	c.VoiceThreshold = 0.1

	c.VoiceThreshold = 1.5
	if c.IsConfigured() {
		t.Error("expected IsConfigured false for VoiceThreshold > 1.0")
	}
	c.VoiceThreshold = 0.1

	c.InputGain = 2.5
	if c.IsConfigured() {
		t.Error("expected IsConfigured false for InputGain > 2.0")
	}
}

func TestVoiceAudioComponent_Serialize(t *testing.T) {
	c := NewVoiceAudioComponent()
	c.SetInputMode(VoiceInputVoiceActivity)
	c.SetVoiceThreshold(0.3)
	c.IsTransmitting = true

	data, err := c.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	c2 := &VoiceAudioComponent{}
	err = c2.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if c2.InputMode != c.InputMode {
		t.Errorf("expected InputMode %s, got %s", c.InputMode, c2.InputMode)
	}
	if c2.VoiceThreshold != c.VoiceThreshold {
		t.Errorf("expected VoiceThreshold %f, got %f", c.VoiceThreshold, c2.VoiceThreshold)
	}
	if c2.IsTransmitting != c.IsTransmitting {
		t.Errorf("expected IsTransmitting %v, got %v", c.IsTransmitting, c2.IsTransmitting)
	}
}

func TestVoiceAudioComponent_InvalidInputMode(t *testing.T) {
	c := NewVoiceAudioComponent()
	c.InputMode = "invalid"

	// ShouldTransmit should return false for unknown mode
	c.UpdateInputLevel(0.5)
	c.SetPushToTalk(true)
	if c.ShouldTransmit() {
		t.Error("expected ShouldTransmit false for invalid input mode")
	}
}
