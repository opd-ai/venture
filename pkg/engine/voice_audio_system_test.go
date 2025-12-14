package engine

import (
	"testing"
)

func TestNewVoiceAudioSystem(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	if system == nil {
		t.Fatal("NewVoiceAudioSystem returned nil")
	}
	if system.world != world {
		t.Error("world not set correctly")
	}
	if system.defaultTransmitCooldown != 0.15 {
		t.Errorf("expected defaultTransmitCooldown 0.15, got %f", system.defaultTransmitCooldown)
	}
	if system.voiceActivityHoldTime != 0.3 {
		t.Errorf("expected voiceActivityHoldTime 0.3, got %f", system.voiceActivityHoldTime)
	}
}

func TestVoiceAudioSystem_SimulateInput(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(NewVoiceAudioComponent())

	err := system.SimulateInput(entity, 0.5)
	if err != nil {
		t.Fatalf("SimulateInput failed: %v", err)
	}

	comp, _ := entity.GetComponent("voice_audio")
	audio := comp.(*VoiceAudioComponent)

	if audio.CurrentInputLevel != 0.5 {
		t.Errorf("expected CurrentInputLevel 0.5, got %f", audio.CurrentInputLevel)
	}
}

func TestVoiceAudioSystem_SimulateInputNilEntity(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	err := system.SimulateInput(nil, 0.5)
	if err != ErrNilEntity {
		t.Errorf("expected ErrNilEntity, got %v", err)
	}
}

func TestVoiceAudioSystem_SimulateInputNoComponent(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	entity := world.CreateEntity()

	err := system.SimulateInput(entity, 0.5)
	if err != ErrNoAudioComponent {
		t.Errorf("expected ErrNoAudioComponent, got %v", err)
	}
}

func TestVoiceAudioSystem_SetPushToTalk(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(NewVoiceAudioComponent())

	err := system.SetPushToTalk(entity, true)
	if err != nil {
		t.Fatalf("SetPushToTalk failed: %v", err)
	}

	comp, _ := entity.GetComponent("voice_audio")
	audio := comp.(*VoiceAudioComponent)

	if !audio.PushToTalkActive {
		t.Error("expected PushToTalkActive to be true")
	}

	err = system.SetPushToTalk(entity, false)
	if err != nil {
		t.Fatalf("SetPushToTalk failed: %v", err)
	}

	if audio.PushToTalkActive {
		t.Error("expected PushToTalkActive to be false")
	}
}

func TestVoiceAudioSystem_SetInputMode(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(NewVoiceAudioComponent())

	err := system.SetInputMode(entity, VoiceInputVoiceActivity)
	if err != nil {
		t.Fatalf("SetInputMode failed: %v", err)
	}

	comp, _ := entity.GetComponent("voice_audio")
	audio := comp.(*VoiceAudioComponent)

	if audio.InputMode != VoiceInputVoiceActivity {
		t.Errorf("expected InputMode voice_activity, got %s", audio.InputMode)
	}
}

func TestVoiceAudioSystem_SetVoiceThreshold(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(NewVoiceAudioComponent())

	err := system.SetVoiceThreshold(entity, 0.25)
	if err != nil {
		t.Fatalf("SetVoiceThreshold failed: %v", err)
	}

	comp, _ := entity.GetComponent("voice_audio")
	audio := comp.(*VoiceAudioComponent)

	if audio.VoiceThreshold != 0.25 {
		t.Errorf("expected VoiceThreshold 0.25, got %f", audio.VoiceThreshold)
	}
}

func TestVoiceAudioSystem_SetOutputVolume(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(NewVoiceAudioComponent())

	err := system.SetOutputVolume(entity, 0.7)
	if err != nil {
		t.Fatalf("SetOutputVolume failed: %v", err)
	}

	comp, _ := entity.GetComponent("voice_audio")
	audio := comp.(*VoiceAudioComponent)

	if audio.OutputVolume != 0.7 {
		t.Errorf("expected OutputVolume 0.7, got %f", audio.OutputVolume)
	}
}

func TestVoiceAudioSystem_SetInputGain(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(NewVoiceAudioComponent())

	err := system.SetInputGain(entity, 1.5)
	if err != nil {
		t.Fatalf("SetInputGain failed: %v", err)
	}

	comp, _ := entity.GetComponent("voice_audio")
	audio := comp.(*VoiceAudioComponent)

	if audio.InputGain != 1.5 {
		t.Errorf("expected InputGain 1.5, got %f", audio.InputGain)
	}
}

func TestVoiceAudioSystem_IsTransmitting(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(NewVoiceAudioComponent())

	if system.IsTransmitting(entity) {
		t.Error("expected IsTransmitting to be false initially")
	}

	comp, _ := entity.GetComponent("voice_audio")
	audio := comp.(*VoiceAudioComponent)
	audio.StartTransmitting()

	if !system.IsTransmitting(entity) {
		t.Error("expected IsTransmitting to be true after starting")
	}

	// Test nil entity
	if system.IsTransmitting(nil) {
		t.Error("expected IsTransmitting to return false for nil entity")
	}
}

func TestVoiceAudioSystem_GetInputLevel(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(NewVoiceAudioComponent())

	system.SimulateInput(entity, 0.6)

	level := system.GetInputLevel(entity)
	if level != 0.6 {
		t.Errorf("expected GetInputLevel 0.6, got %f", level)
	}

	// Test nil entity
	if system.GetInputLevel(nil) != 0 {
		t.Error("expected GetInputLevel to return 0 for nil entity")
	}
}

func TestVoiceAudioSystem_Update_PushToTalk(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	entity := world.CreateEntity()
	audio := NewVoiceAudioComponent()
	audio.SetInputMode(VoiceInputPushToTalk)
	audio.SetNoiseGateLevel(0.1)
	entity.AddComponent(audio)

	entities := []*Entity{entity}

	// Not pressing PTT
	audio.UpdateInputLevel(0.5)
	system.Update(entities, 0.016)
	if audio.IsTransmitting {
		t.Error("expected not transmitting without PTT")
	}

	// Press PTT
	audio.SetPushToTalk(true)
	audio.UpdateInputLevel(0.5)
	system.Update(entities, 0.016)
	if !audio.IsTransmitting {
		t.Error("expected transmitting with PTT active")
	}

	// Release PTT
	audio.SetPushToTalk(false)
	system.Update(entities, 0.016)
	if audio.IsTransmitting {
		t.Error("expected not transmitting after PTT release")
	}
}

func TestVoiceAudioSystem_Update_VoiceActivity(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)
	system.voiceActivityHoldTime = 0.1 // Short for testing

	entity := world.CreateEntity()
	audio := NewVoiceActivityComponent(0.2)
	audio.SetNoiseGateLevel(0.1)
	entity.AddComponent(audio)

	entities := []*Entity{entity}

	// Below threshold
	audio.UpdateInputLevel(0.15)
	system.Update(entities, 0.016)
	if audio.IsTransmitting {
		t.Error("expected not transmitting below threshold")
	}

	// Above threshold
	audio.UpdateInputLevel(0.5)
	system.Update(entities, 0.016)
	if !audio.IsTransmitting {
		t.Error("expected transmitting above threshold")
	}

	// Drop voice but still in hold period
	audio.UpdateInputLevel(0.05)
	system.Update(entities, 0.05)
	if !audio.IsTransmitting {
		t.Error("expected still transmitting during hold period")
	}

	// Wait for hold period to expire
	system.Update(entities, 0.1)
	if audio.IsTransmitting {
		t.Error("expected not transmitting after hold period")
	}
}

func TestVoiceAudioSystem_Update_SyncWithVoiceChannel(t *testing.T) {
	world := NewWorld()
	audioSystem := NewVoiceAudioSystem(world)

	entity := world.CreateEntity()
	audio := NewVoiceAudioComponent()
	audio.SetInputMode(VoiceInputPushToTalk)
	audio.SetNoiseGateLevel(0.1)
	entity.AddComponent(audio)

	vc := NewVoiceChannelComponent("test", VoiceChannelParty)
	entity.AddComponent(vc)

	entities := []*Entity{entity}

	// Start transmitting
	audio.SetPushToTalk(true)
	audio.UpdateInputLevel(0.5)
	audioSystem.Update(entities, 0.016)

	// Voice channel should be updated
	if !vc.IsSpeaking {
		t.Error("expected VoiceChannelComponent.IsSpeaking to be true")
	}

	// Stop transmitting
	audio.SetPushToTalk(false)
	audioSystem.Update(entities, 0.016)

	if vc.IsSpeaking {
		t.Error("expected VoiceChannelComponent.IsSpeaking to be false")
	}
}

func TestVoiceAudioSystem_Update_NoComponent(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	entity := world.CreateEntity()
	entities := []*Entity{entity}

	// Should not panic
	system.Update(entities, 0.016)
}

func TestVoiceAudioSystem_NilEntityErrors(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)

	if err := system.SetPushToTalk(nil, true); err != ErrNilEntity {
		t.Errorf("SetPushToTalk nil: expected ErrNilEntity, got %v", err)
	}
	if err := system.SetInputMode(nil, VoiceInputPushToTalk); err != ErrNilEntity {
		t.Errorf("SetInputMode nil: expected ErrNilEntity, got %v", err)
	}
	if err := system.SetVoiceThreshold(nil, 0.5); err != ErrNilEntity {
		t.Errorf("SetVoiceThreshold nil: expected ErrNilEntity, got %v", err)
	}
	if err := system.SetOutputVolume(nil, 0.5); err != ErrNilEntity {
		t.Errorf("SetOutputVolume nil: expected ErrNilEntity, got %v", err)
	}
	if err := system.SetInputGain(nil, 1.0); err != ErrNilEntity {
		t.Errorf("SetInputGain nil: expected ErrNilEntity, got %v", err)
	}
}

func TestVoiceAudioSystem_NoComponentErrors(t *testing.T) {
	world := NewWorld()
	system := NewVoiceAudioSystem(world)
	entity := world.CreateEntity()

	if err := system.SetPushToTalk(entity, true); err != ErrNoAudioComponent {
		t.Errorf("SetPushToTalk no component: expected ErrNoAudioComponent, got %v", err)
	}
	if err := system.SetInputMode(entity, VoiceInputPushToTalk); err != ErrNoAudioComponent {
		t.Errorf("SetInputMode no component: expected ErrNoAudioComponent, got %v", err)
	}
	if err := system.SetVoiceThreshold(entity, 0.5); err != ErrNoAudioComponent {
		t.Errorf("SetVoiceThreshold no component: expected ErrNoAudioComponent, got %v", err)
	}
	if err := system.SetOutputVolume(entity, 0.5); err != ErrNoAudioComponent {
		t.Errorf("SetOutputVolume no component: expected ErrNoAudioComponent, got %v", err)
	}
	if err := system.SetInputGain(entity, 1.0); err != ErrNoAudioComponent {
		t.Errorf("SetInputGain no component: expected ErrNoAudioComponent, got %v", err)
	}
}
