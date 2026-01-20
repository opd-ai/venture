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

func TestVoiceAudioSystem_SetMethods(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(NewVoiceAudioComponent())

	comp, _ := entity.GetComponent("voice_audio")
	audio := comp.(*VoiceAudioComponent)

	system.SetInputMode(entity, VoiceInputVoiceActivity)
	if audio.InputMode != VoiceInputVoiceActivity {
		t.Errorf("expected InputMode voice_activity, got %s", audio.InputMode)
	}

	system.SetPushToTalkKey(entity, "T")
	if audio.PushToTalkKey != "T" {
		t.Errorf("expected PushToTalkKey 'T', got '%s'", audio.PushToTalkKey)
	}

	system.SetVoiceThreshold(entity, 0.3)
	if audio.VoiceThreshold != 0.3 {
		t.Errorf("expected VoiceThreshold 0.3, got %f", audio.VoiceThreshold)
	}

	system.SetVoiceThreshold(entity, 2.0) // Should be clamped
	if audio.VoiceThreshold != 1.0 {
		t.Errorf("expected VoiceThreshold 1.0 (clamped), got %f", audio.VoiceThreshold)
	}

	system.SetNoiseGateLevel(entity, 0.1)
	if audio.NoiseGateLevel != 0.1 {
		t.Errorf("expected NoiseGateLevel 0.1, got %f", audio.NoiseGateLevel)
	}

	system.SetOutputVolume(entity, 0.5)
	if audio.OutputVolume != 0.5 {
		t.Errorf("expected OutputVolume 0.5, got %f", audio.OutputVolume)
	}

	system.SetInputGain(entity, 1.5)
	if audio.InputGain != 1.5 {
		t.Errorf("expected InputGain 1.5, got %f", audio.InputGain)
	}

	system.SetInputGain(entity, 3.0) // Should be clamped to 2.0
	if audio.InputGain != 2.0 {
		t.Errorf("expected InputGain 2.0 (clamped), got %f", audio.InputGain)
	}
}

func TestVoiceAudioSystem_UpdateInputLevel(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(NewVoiceAudioComponent())

	comp, _ := entity.GetComponent("voice_audio")
	audio := comp.(*VoiceAudioComponent)

	system.SetVoiceThreshold(entity, 0.2)
	system.SetNoiseGateLevel(entity, 0.1)
	system.SetInputGain(entity, 1.0)

	// Level below noise gate
	system.SimulateInput(entity, 0.05)
	if audio.NoiseGateOpen {
		t.Error("expected NoiseGateOpen to be false for level below gate")
	}
	if audio.VoiceActivityDetected {
		t.Error("expected VoiceActivityDetected to be false")
	}

	// Level above noise gate but below voice threshold
	system.SimulateInput(entity, 0.15)
	if !audio.NoiseGateOpen {
		t.Error("expected NoiseGateOpen to be true")
	}
	if audio.VoiceActivityDetected {
		t.Error("expected VoiceActivityDetected to be false (below threshold)")
	}

	// Level above voice threshold
	system.SimulateInput(entity, 0.5)
	if !audio.NoiseGateOpen {
		t.Error("expected NoiseGateOpen to be true")
	}
	if !audio.VoiceActivityDetected {
		t.Error("expected VoiceActivityDetected to be true")
	}

	// Test input gain effect
	system.SetInputGain(entity, 2.0)
	system.SimulateInput(entity, 0.3)
	if audio.NormalizedInputLevel != 0.6 {
		t.Errorf("expected NormalizedInputLevel 0.6, got %f", audio.NormalizedInputLevel)
	}

	// Test clamping on high input
	system.SimulateInput(entity, 0.8) // 0.8 * 2.0 = 1.6, should clamp to 1.0
	if audio.NormalizedInputLevel != 1.0 {
		t.Errorf("expected NormalizedInputLevel 1.0 (clamped), got %f", audio.NormalizedInputLevel)
	}
}

func TestVoiceAudioSystem_ShouldTransmit_PushToTalk(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	entity := world.CreateEntity()
	audio := NewVoiceAudioComponent()
	entity.AddComponent(audio)

	system.SetInputMode(entity, VoiceInputPushToTalk)
	system.SetNoiseGateLevel(entity, 0.1)

	// Not pressing PTT, should not transmit
	system.SimulateInput(entity, 0.5)
	if system.shouldTransmit(audio) {
		t.Error("expected shouldTransmit false when PTT not active")
	}

	// Pressing PTT but below noise gate
	system.SetPushToTalk(entity, true)
	system.SimulateInput(entity, 0.05)
	if system.shouldTransmit(audio) {
		t.Error("expected shouldTransmit false when below noise gate")
	}

	// Pressing PTT and above noise gate
	system.SimulateInput(entity, 0.5)
	if !system.shouldTransmit(audio) {
		t.Error("expected shouldTransmit true when PTT active and above gate")
	}

	// Release PTT
	system.SetPushToTalk(entity, false)
	if system.shouldTransmit(audio) {
		t.Error("expected shouldTransmit false when PTT released")
	}
}

func TestVoiceAudioSystem_ShouldTransmit_VoiceActivity(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	entity := world.CreateEntity()
	audio := NewVoiceActivityComponent(0.2)
	entity.AddComponent(audio)

	system.SetNoiseGateLevel(entity, 0.1)

	// Below threshold
	system.SimulateInput(entity, 0.15)
	if system.shouldTransmit(audio) {
		t.Error("expected shouldTransmit false when below threshold")
	}

	// Above threshold
	system.SimulateInput(entity, 0.5)
	if !system.shouldTransmit(audio) {
		t.Error("expected shouldTransmit true when above threshold")
	}
}

func TestVoiceAudioSystem_ShouldTransmit_Cooldown(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	entity := world.CreateEntity()
	audio := NewVoiceAudioComponent()
	entity.AddComponent(audio)

	system.SetInputMode(entity, VoiceInputPushToTalk)
	system.SetNoiseGateLevel(entity, 0.1)

	// Start transmitting
	system.SetPushToTalk(entity, true)
	system.SimulateInput(entity, 0.5)
	system.StartTransmitting(entity)

	// Stop with cooldown
	system.StopTransmitting(entity, 0.5)

	// During cooldown, should maintain false state (not transmitting)
	system.SetPushToTalk(entity, true)
	system.SimulateInput(entity, 0.5)
	if system.shouldTransmit(audio) {
		t.Error("expected shouldTransmit to return current state during cooldown")
	}

	// After cooldown expires
	system.updateCooldown(audio, 0.6)
	if !system.shouldTransmit(audio) {
		t.Error("expected shouldTransmit true after cooldown expires")
	}
}

func TestVoiceAudioSystem_TransmitState(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	entity := world.CreateEntity()
	audio := NewVoiceAudioComponent()
	entity.AddComponent(audio)

	// Start transmitting
	if !system.StartTransmitting(entity) {
		t.Error("expected StartTransmitting to return true on first call")
	}
	if !audio.IsTransmitting {
		t.Error("expected IsTransmitting to be true")
	}

	// Start again (already transmitting)
	if system.StartTransmitting(entity) {
		t.Error("expected StartTransmitting to return false when already transmitting")
	}

	// Stop transmitting
	if !system.StopTransmitting(entity, 0.2) {
		t.Error("expected StopTransmitting to return true")
	}
	if audio.IsTransmitting {
		t.Error("expected IsTransmitting to be false")
	}
	if audio.TransmitCooldown != 0.2 {
		t.Errorf("expected TransmitCooldown 0.2, got %f", audio.TransmitCooldown)
	}

	// Stop again (not transmitting)
	if system.StopTransmitting(entity, 0.1) {
		t.Error("expected StopTransmitting to return false when not transmitting")
	}
}

func TestVoiceAudioSystem_UpdateCooldown(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	entity := world.CreateEntity()
	audio := NewVoiceAudioComponent()
	entity.AddComponent(audio)

	audio.TransmitCooldown = 0.5

	system.updateCooldown(audio, 0.2)
	if audio.TransmitCooldown != 0.3 {
		t.Errorf("expected TransmitCooldown 0.3, got %f", audio.TransmitCooldown)
	}

	system.updateCooldown(audio, 0.5) // Should clamp to 0
	if audio.TransmitCooldown != 0 {
		t.Errorf("expected TransmitCooldown 0, got %f", audio.TransmitCooldown)
	}

	// No cooldown, should remain 0
	system.updateCooldown(audio, 0.1)
	if audio.TransmitCooldown != 0 {
		t.Errorf("expected TransmitCooldown 0, got %f", audio.TransmitCooldown)
	}
}

func TestVoiceAudioSystem_GetEffectiveOutputLevel(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(NewVoiceAudioComponent())

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
			system.SetOutputVolume(entity, tt.volume)
			result := system.GetEffectiveOutputLevel(entity, tt.inputLevel)
			if result != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestVoiceAudioSystem_IsConfigured(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(NewVoiceAudioComponent())

	comp, _ := entity.GetComponent("voice_audio")
	audio := comp.(*VoiceAudioComponent)

	if !system.IsConfigured(entity) {
		t.Error("expected default component to be configured")
	}

	// Test invalid configurations
	audio.VoiceThreshold = -0.1
	if system.IsConfigured(entity) {
		t.Error("expected IsConfigured false for negative VoiceThreshold")
	}
	audio.VoiceThreshold = 0.1

	audio.VoiceThreshold = 1.5
	if system.IsConfigured(entity) {
		t.Error("expected IsConfigured false for VoiceThreshold > 1.0")
	}
	audio.VoiceThreshold = 0.1

	audio.InputGain = 2.5
	if system.IsConfigured(entity) {
		t.Error("expected IsConfigured false for InputGain > 2.0")
	}
}

func TestVoiceAudioComponent_Serialize(t *testing.T) {
	c := NewVoiceAudioComponent()
	c.InputMode = VoiceInputVoiceActivity
	c.VoiceThreshold = 0.3
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

func TestVoiceAudioSystem_InvalidInputMode(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	entity := world.CreateEntity()
	audio := NewVoiceAudioComponent()
	audio.InputMode = "invalid"
	entity.AddComponent(audio)

	// shouldTransmit should return false for unknown mode
	system.SimulateInput(entity, 0.5)
	system.SetPushToTalk(entity, true)
	if system.shouldTransmit(audio) {
		t.Error("expected shouldTransmit false for invalid input mode")
	}
}

func TestVoiceAudioSystem_SetNoiseGateLevel(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(NewVoiceAudioComponent())

	comp, _ := entity.GetComponent("voice_audio")
	audio := comp.(*VoiceAudioComponent)

	err := system.SetNoiseGateLevel(entity, 0.15)
	if err != nil {
		t.Fatalf("SetNoiseGateLevel failed: %v", err)
	}
	if audio.NoiseGateLevel != 0.15 {
		t.Errorf("expected NoiseGateLevel 0.15, got %f", audio.NoiseGateLevel)
	}

	// Test clamping
	system.SetNoiseGateLevel(entity, 2.0)
	if audio.NoiseGateLevel != 1.0 {
		t.Errorf("expected NoiseGateLevel 1.0 (clamped), got %f", audio.NoiseGateLevel)
	}
}

func TestVoiceAudioSystem_SetPushToTalkKey(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(NewVoiceAudioComponent())

	comp, _ := entity.GetComponent("voice_audio")
	audio := comp.(*VoiceAudioComponent)

	err := system.SetPushToTalkKey(entity, "G")
	if err != nil {
		t.Fatalf("SetPushToTalkKey failed: %v", err)
	}
	if audio.PushToTalkKey != "G" {
		t.Errorf("expected PushToTalkKey 'G', got '%s'", audio.PushToTalkKey)
	}
}

func TestVoiceAudioSystem_NilEntityMethods(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	if err := system.SetNoiseGateLevel(nil, 0.1); err != ErrNilEntity {
		t.Errorf("SetNoiseGateLevel nil: expected ErrNilEntity, got %v", err)
	}
	if err := system.SetPushToTalkKey(nil, "T"); err != ErrNilEntity {
		t.Errorf("SetPushToTalkKey nil: expected ErrNilEntity, got %v", err)
	}
	if system.IsConfigured(nil) {
		t.Error("expected IsConfigured to return false for nil entity")
	}
	if system.GetEffectiveOutputLevel(nil, 0.5) != 0 {
		t.Error("expected GetEffectiveOutputLevel to return 0 for nil entity")
	}
	if system.StartTransmitting(nil) {
		t.Error("expected StartTransmitting to return false for nil entity")
	}
	if system.StopTransmitting(nil, 0.1) {
		t.Error("expected StopTransmitting to return false for nil entity")
	}
}

func TestVoiceAudioSystem_NoComponentMethods(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)
	entity := world.CreateEntity()

	if err := system.SetNoiseGateLevel(entity, 0.1); err != ErrNoAudioComponent {
		t.Errorf("SetNoiseGateLevel no component: expected ErrNoAudioComponent, got %v", err)
	}
	if err := system.SetPushToTalkKey(entity, "T"); err != ErrNoAudioComponent {
		t.Errorf("SetPushToTalkKey no component: expected ErrNoAudioComponent, got %v", err)
	}
	if system.IsConfigured(entity) {
		t.Error("expected IsConfigured to return false for entity without component")
	}
	if system.GetEffectiveOutputLevel(entity, 0.5) != 0 {
		t.Error("expected GetEffectiveOutputLevel to return 0 for entity without component")
	}
	if system.StartTransmitting(entity) {
		t.Error("expected StartTransmitting to return false for entity without component")
	}
	if system.StopTransmitting(entity, 0.1) {
		t.Error("expected StopTransmitting to return false for entity without component")
	}
}
