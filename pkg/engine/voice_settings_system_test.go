package engine

import (
	"testing"
)

func TestNewVoiceSettingsSystem(t *testing.T) {
	world := NewWorld()
	system := NewVoiceSettingsSystem(world)

	if system == nil {
		t.Fatal("NewVoiceSettingsSystem returned nil")
	}
	if system.world != world {
		t.Error("world not set correctly")
	}
}

func TestVoiceSettingsSystem_MarkSpeaking(t *testing.T) {
	world := NewWorld()
	system := NewVoiceSettingsSystem(world)

	entity := world.CreateEntity()
	settings := NewVoiceSettingsComponent()
	entity.AddComponent(settings)

	err := system.MarkSpeaking(entity, "speaker-1")
	if err != nil {
		t.Fatalf("MarkSpeaking failed: %v", err)
	}

	if settings.GetSpeakingUserCount() != 1 {
		t.Errorf("expected 1 speaking user, got %d", settings.GetSpeakingUserCount())
	}

	// Check timer was set
	if _, exists := system.speakingTimers["speaker-1"]; !exists {
		t.Error("expected speaking timer to be set")
	}
}

func TestVoiceSettingsSystem_MarkNotSpeaking(t *testing.T) {
	world := NewWorld()
	system := NewVoiceSettingsSystem(world)

	entity := world.CreateEntity()
	settings := NewVoiceSettingsComponent()
	entity.AddComponent(settings)

	system.MarkSpeaking(entity, "speaker-1")
	err := system.MarkNotSpeaking(entity, "speaker-1")
	if err != nil {
		t.Fatalf("MarkNotSpeaking failed: %v", err)
	}

	if settings.GetSpeakingUserCount() != 0 {
		t.Errorf("expected 0 speaking users, got %d", settings.GetSpeakingUserCount())
	}

	// Check timer was removed
	if _, exists := system.speakingTimers["speaker-1"]; exists {
		t.Error("expected speaking timer to be removed")
	}
}

func TestVoiceSettingsSystem_SetMasterVolume(t *testing.T) {
	world := NewWorld()
	system := NewVoiceSettingsSystem(world)

	entity := world.CreateEntity()
	settings := NewVoiceSettingsComponent()
	entity.AddComponent(settings)

	err := system.SetMasterVolume(entity, 0.7)
	if err != nil {
		t.Fatalf("SetMasterVolume failed: %v", err)
	}

	if settings.MasterVolume != 0.7 {
		t.Errorf("expected MasterVolume 0.7, got %f", settings.MasterVolume)
	}
}

func TestVoiceSettingsSystem_MuteUnmuteUser(t *testing.T) {
	world := NewWorld()
	system := NewVoiceSettingsSystem(world)

	entity := world.CreateEntity()
	settings := NewVoiceSettingsComponent()
	entity.AddComponent(settings)

	err := system.MuteUser(entity, "player-1")
	if err != nil {
		t.Fatalf("MuteUser failed: %v", err)
	}

	if !settings.IsUserMuted("player-1") {
		t.Error("expected player-1 to be muted")
	}

	err = system.UnmuteUser(entity, "player-1")
	if err != nil {
		t.Fatalf("UnmuteUser failed: %v", err)
	}

	if settings.IsUserMuted("player-1") {
		t.Error("expected player-1 to be unmuted")
	}
}

func TestVoiceSettingsSystem_SetUserVolume(t *testing.T) {
	world := NewWorld()
	system := NewVoiceSettingsSystem(world)

	entity := world.CreateEntity()
	settings := NewVoiceSettingsComponent()
	entity.AddComponent(settings)

	err := system.SetUserVolume(entity, "player-1", 0.5)
	if err != nil {
		t.Fatalf("SetUserVolume failed: %v", err)
	}

	if settings.GetUserVolume("player-1") != 0.5 {
		t.Errorf("expected user volume 0.5, got %f", settings.GetUserVolume("player-1"))
	}
}

func TestVoiceSettingsSystem_SetDevices(t *testing.T) {
	world := NewWorld()
	system := NewVoiceSettingsSystem(world)

	entity := world.CreateEntity()
	settings := NewVoiceSettingsComponent()
	entity.AddComponent(settings)

	err := system.SetInputDevice(entity, "Microphone 1")
	if err != nil {
		t.Fatalf("SetInputDevice failed: %v", err)
	}
	if settings.InputDevice != "Microphone 1" {
		t.Errorf("expected InputDevice 'Microphone 1', got '%s'", settings.InputDevice)
	}

	err = system.SetOutputDevice(entity, "Speakers")
	if err != nil {
		t.Fatalf("SetOutputDevice failed: %v", err)
	}
	if settings.OutputDevice != "Speakers" {
		t.Errorf("expected OutputDevice 'Speakers', got '%s'", settings.OutputDevice)
	}
}

func TestVoiceSettingsSystem_GetSpeakingUsers(t *testing.T) {
	world := NewWorld()
	system := NewVoiceSettingsSystem(world)

	entity := world.CreateEntity()
	settings := NewVoiceSettingsComponent()
	entity.AddComponent(settings)

	system.MarkSpeaking(entity, "speaker-1")
	system.MarkSpeaking(entity, "speaker-2")

	speaking := system.GetSpeakingUsers(entity)
	if len(speaking) != 2 {
		t.Errorf("expected 2 speaking users, got %d", len(speaking))
	}

	// Nil entity
	speaking = system.GetSpeakingUsers(nil)
	if speaking != nil {
		t.Error("expected nil for nil entity")
	}

	// No component
	entity2 := world.CreateEntity()
	speaking = system.GetSpeakingUsers(entity2)
	if speaking != nil {
		t.Error("expected nil for entity without component")
	}
}

func TestVoiceSettingsSystem_GetEffectiveVolume(t *testing.T) {
	world := NewWorld()
	system := NewVoiceSettingsSystem(world)

	entity := world.CreateEntity()
	settings := NewVoiceSettingsComponent()
	entity.AddComponent(settings)

	settings.SetMasterVolume(0.5)
	settings.SetUserVolume("player-1", 0.5)

	vol := system.GetEffectiveVolume(entity, "player-1", 1.0)
	// 1.0 * 0.5 (master) * 0.5 (user) = 0.25
	if vol != 0.25 {
		t.Errorf("expected effective volume 0.25, got %f", vol)
	}

	// Nil entity returns base volume
	vol = system.GetEffectiveVolume(nil, "player-1", 0.8)
	if vol != 0.8 {
		t.Errorf("expected base volume 0.8, got %f", vol)
	}
}

func TestVoiceSettingsSystem_IsUserMuted(t *testing.T) {
	world := NewWorld()
	system := NewVoiceSettingsSystem(world)

	entity := world.CreateEntity()
	settings := NewVoiceSettingsComponent()
	entity.AddComponent(settings)

	if system.IsUserMuted(entity, "player-1") {
		t.Error("expected player-1 to not be muted initially")
	}

	settings.MuteUser("player-1")
	if !system.IsUserMuted(entity, "player-1") {
		t.Error("expected player-1 to be muted")
	}

	// Nil entity
	if system.IsUserMuted(nil, "player-1") {
		t.Error("expected false for nil entity")
	}
}

func TestVoiceSettingsSystem_Update(t *testing.T) {
	world := NewWorld()
	system := NewVoiceSettingsSystem(world)
	system.speakingTimeout = 0.1 // Short for testing

	entity := world.CreateEntity()
	settings := NewVoiceSettingsComponent()
	entity.AddComponent(settings)

	// Mark as speaking
	system.MarkSpeaking(entity, "speaker-1")

	entities := []*Entity{entity}

	// Update with small delta - should still be speaking
	system.Update(entities, 0.05)
	if settings.GetSpeakingUserCount() != 1 {
		t.Error("expected speaker to still be speaking")
	}

	// Update past timeout - should stop speaking
	system.Update(entities, 0.1)
	if settings.GetSpeakingUserCount() != 0 {
		t.Errorf("expected speaker to stop speaking, got %d", settings.GetSpeakingUserCount())
	}
}

func TestVoiceSettingsSystem_UpdateNoComponent(t *testing.T) {
	world := NewWorld()
	system := NewVoiceSettingsSystem(world)

	entity := world.CreateEntity()
	entities := []*Entity{entity}

	// Should not panic
	system.Update(entities, 0.016)
}

func TestVoiceSettingsSystem_NilEntityErrors(t *testing.T) {
	world := NewWorld()
	system := NewVoiceSettingsSystem(world)

	if err := system.MarkSpeaking(nil, "x"); err != ErrNilEntity {
		t.Errorf("MarkSpeaking nil: expected ErrNilEntity, got %v", err)
	}
	if err := system.MarkNotSpeaking(nil, "x"); err != ErrNilEntity {
		t.Errorf("MarkNotSpeaking nil: expected ErrNilEntity, got %v", err)
	}
	if err := system.SetMasterVolume(nil, 0.5); err != ErrNilEntity {
		t.Errorf("SetMasterVolume nil: expected ErrNilEntity, got %v", err)
	}
	if err := system.MuteUser(nil, "x"); err != ErrNilEntity {
		t.Errorf("MuteUser nil: expected ErrNilEntity, got %v", err)
	}
	if err := system.UnmuteUser(nil, "x"); err != ErrNilEntity {
		t.Errorf("UnmuteUser nil: expected ErrNilEntity, got %v", err)
	}
	if err := system.SetUserVolume(nil, "x", 0.5); err != ErrNilEntity {
		t.Errorf("SetUserVolume nil: expected ErrNilEntity, got %v", err)
	}
	if err := system.SetInputDevice(nil, "x"); err != ErrNilEntity {
		t.Errorf("SetInputDevice nil: expected ErrNilEntity, got %v", err)
	}
	if err := system.SetOutputDevice(nil, "x"); err != ErrNilEntity {
		t.Errorf("SetOutputDevice nil: expected ErrNilEntity, got %v", err)
	}
}

func TestVoiceSettingsSystem_NoComponentErrors(t *testing.T) {
	world := NewWorld()
	system := NewVoiceSettingsSystem(world)
	entity := world.CreateEntity()

	if err := system.MarkSpeaking(entity, "x"); err != ErrNoSettingsComponent {
		t.Errorf("MarkSpeaking no component: expected ErrNoSettingsComponent, got %v", err)
	}
	if err := system.MarkNotSpeaking(entity, "x"); err != ErrNoSettingsComponent {
		t.Errorf("MarkNotSpeaking no component: expected ErrNoSettingsComponent, got %v", err)
	}
	if err := system.SetMasterVolume(entity, 0.5); err != ErrNoSettingsComponent {
		t.Errorf("SetMasterVolume no component: expected ErrNoSettingsComponent, got %v", err)
	}
	if err := system.MuteUser(entity, "x"); err != ErrNoSettingsComponent {
		t.Errorf("MuteUser no component: expected ErrNoSettingsComponent, got %v", err)
	}
	if err := system.UnmuteUser(entity, "x"); err != ErrNoSettingsComponent {
		t.Errorf("UnmuteUser no component: expected ErrNoSettingsComponent, got %v", err)
	}
	if err := system.SetUserVolume(entity, "x", 0.5); err != ErrNoSettingsComponent {
		t.Errorf("SetUserVolume no component: expected ErrNoSettingsComponent, got %v", err)
	}
	if err := system.SetInputDevice(entity, "x"); err != ErrNoSettingsComponent {
		t.Errorf("SetInputDevice no component: expected ErrNoSettingsComponent, got %v", err)
	}
	if err := system.SetOutputDevice(entity, "x"); err != ErrNoSettingsComponent {
		t.Errorf("SetOutputDevice no component: expected ErrNoSettingsComponent, got %v", err)
	}
}
