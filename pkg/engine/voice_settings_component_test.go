package engine

import (
	"testing"
)

func TestVoiceSettingsComponent_Type(t *testing.T) {
	c := NewVoiceSettingsComponent()
	if c.Type() != "voice_settings" {
		t.Errorf("expected type 'voice_settings', got '%s'", c.Type())
	}
}

func TestNewVoiceSettingsComponent(t *testing.T) {
	c := NewVoiceSettingsComponent()

	if c.MasterVolume != 1.0 {
		t.Errorf("expected MasterVolume 1.0, got %f", c.MasterVolume)
	}
	if c.MutedUsers == nil {
		t.Error("expected MutedUsers to be initialized")
	}
	if c.UserVolumes == nil {
		t.Error("expected UserVolumes to be initialized")
	}
	if !c.SpeakingIndicatorEnabled {
		t.Error("expected SpeakingIndicatorEnabled to be true")
	}
	if !c.ProximityEnabled {
		t.Error("expected ProximityEnabled to be true")
	}
	if !c.EchoCancellation {
		t.Error("expected EchoCancellation to be true")
	}
	if !c.NoiseSuppression {
		t.Error("expected NoiseSuppression to be true")
	}
	if !c.AutoGainControl {
		t.Error("expected AutoGainControl to be true")
	}
}

func TestVoiceSettingsComponent_SetMasterVolume(t *testing.T) {
	c := NewVoiceSettingsComponent()

	c.SetMasterVolume(0.5)
	if c.MasterVolume != 0.5 {
		t.Errorf("expected MasterVolume 0.5, got %f", c.MasterVolume)
	}

	// Test clamping
	c.SetMasterVolume(1.5)
	if c.MasterVolume != 1.0 {
		t.Errorf("expected MasterVolume 1.0 (clamped), got %f", c.MasterVolume)
	}

	c.SetMasterVolume(-0.5)
	if c.MasterVolume != 0.0 {
		t.Errorf("expected MasterVolume 0.0 (clamped), got %f", c.MasterVolume)
	}
}

func TestVoiceSettingsComponent_MuteUser(t *testing.T) {
	c := NewVoiceSettingsComponent()

	c.MuteUser("player-1")
	if !c.IsUserMuted("player-1") {
		t.Error("expected player-1 to be muted")
	}

	c.UnmuteUser("player-1")
	if c.IsUserMuted("player-1") {
		t.Error("expected player-1 to be unmuted")
	}
}

func TestVoiceSettingsComponent_ToggleUserMute(t *testing.T) {
	c := NewVoiceSettingsComponent()

	// Toggle on
	muted := c.ToggleUserMute("player-1")
	if !muted {
		t.Error("expected ToggleUserMute to return true (muted)")
	}
	if !c.IsUserMuted("player-1") {
		t.Error("expected player-1 to be muted")
	}

	// Toggle off
	muted = c.ToggleUserMute("player-1")
	if muted {
		t.Error("expected ToggleUserMute to return false (unmuted)")
	}
	if c.IsUserMuted("player-1") {
		t.Error("expected player-1 to be unmuted")
	}
}

func TestVoiceSettingsComponent_UserVolume(t *testing.T) {
	c := NewVoiceSettingsComponent()

	// Default volume
	vol := c.GetUserVolume("player-1")
	if vol != 1.0 {
		t.Errorf("expected default volume 1.0, got %f", vol)
	}

	// Set volume
	c.SetUserVolume("player-1", 0.5)
	vol = c.GetUserVolume("player-1")
	if vol != 0.5 {
		t.Errorf("expected volume 0.5, got %f", vol)
	}

	// Volume clamping
	c.SetUserVolume("player-1", 3.0)
	vol = c.GetUserVolume("player-1")
	if vol != 2.0 {
		t.Errorf("expected volume 2.0 (clamped), got %f", vol)
	}

	// Reset volume
	c.ResetUserVolume("player-1")
	vol = c.GetUserVolume("player-1")
	if vol != 1.0 {
		t.Errorf("expected volume 1.0 after reset, got %f", vol)
	}
}

func TestVoiceSettingsComponent_GetEffectiveVolume(t *testing.T) {
	c := NewVoiceSettingsComponent()

	tests := []struct {
		name         string
		entityID     string
		baseVolume   float64
		masterVolume float64
		userVolume   float64
		muted        bool
		want         float64
	}{
		{"full volume", "p1", 1.0, 1.0, 1.0, false, 1.0},
		{"half master", "p2", 1.0, 0.5, 1.0, false, 0.5},
		{"half user", "p3", 1.0, 1.0, 0.5, false, 0.5},
		{"combined", "p4", 0.8, 0.5, 0.5, false, 0.2},
		{"muted", "p5", 1.0, 1.0, 1.0, true, 0.0},
		{"boosted", "p6", 0.5, 1.0, 2.0, false, 1.0}, // Clamped to 1.0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.SetMasterVolume(tt.masterVolume)
			if tt.userVolume != 1.0 {
				c.SetUserVolume(tt.entityID, tt.userVolume)
			}
			if tt.muted {
				c.MuteUser(tt.entityID)
			}

			vol := c.GetEffectiveVolume(tt.entityID, tt.baseVolume)
			if vol != tt.want {
				t.Errorf("expected %f, got %f", tt.want, vol)
			}

			// Clean up
			c.UnmuteUser(tt.entityID)
			c.ResetUserVolume(tt.entityID)
		})
	}
}

func TestVoiceSettingsComponent_Devices(t *testing.T) {
	c := NewVoiceSettingsComponent()

	c.SetInputDevice("Microphone 1")
	if c.InputDevice != "Microphone 1" {
		t.Errorf("expected InputDevice 'Microphone 1', got '%s'", c.InputDevice)
	}

	c.SetOutputDevice("Speakers")
	if c.OutputDevice != "Speakers" {
		t.Errorf("expected OutputDevice 'Speakers', got '%s'", c.OutputDevice)
	}
}

func TestVoiceSettingsComponent_SetInputSensitivity(t *testing.T) {
	c := NewVoiceSettingsComponent()

	c.SetInputSensitivity(0.3)
	if c.InputSensitivity != 0.3 {
		t.Errorf("expected InputSensitivity 0.3, got %f", c.InputSensitivity)
	}

	// Clamping
	c.SetInputSensitivity(1.5)
	if c.InputSensitivity != 1.0 {
		t.Errorf("expected InputSensitivity 1.0 (clamped), got %f", c.InputSensitivity)
	}
}

func TestVoiceSettingsComponent_SpeakingUsers(t *testing.T) {
	c := NewVoiceSettingsComponent()

	// Initially no one speaking
	if c.GetSpeakingUserCount() != 0 {
		t.Errorf("expected 0 speaking users, got %d", c.GetSpeakingUserCount())
	}

	// Add speaking users
	c.UpdateSpeakingUser("player-1", true)
	c.UpdateSpeakingUser("player-2", true)

	if c.GetSpeakingUserCount() != 2 {
		t.Errorf("expected 2 speaking users, got %d", c.GetSpeakingUserCount())
	}

	speaking := c.GetSpeakingUsers()
	if len(speaking) != 2 {
		t.Errorf("expected 2 speaking users in list, got %d", len(speaking))
	}

	// Stop speaking
	c.UpdateSpeakingUser("player-1", false)
	if c.GetSpeakingUserCount() != 1 {
		t.Errorf("expected 1 speaking user, got %d", c.GetSpeakingUserCount())
	}

	// Clear all
	c.ClearSpeakingUsers()
	if c.GetSpeakingUserCount() != 0 {
		t.Errorf("expected 0 speaking users after clear, got %d", c.GetSpeakingUserCount())
	}
}

func TestVoiceSettingsComponent_ToggleSettings(t *testing.T) {
	c := NewVoiceSettingsComponent()

	c.SetProximityEnabled(false)
	if c.ProximityEnabled {
		t.Error("expected ProximityEnabled to be false")
	}

	c.SetEchoCancellation(false)
	if c.EchoCancellation {
		t.Error("expected EchoCancellation to be false")
	}

	c.SetNoiseSuppression(false)
	if c.NoiseSuppression {
		t.Error("expected NoiseSuppression to be false")
	}

	c.SetAutoGainControl(false)
	if c.AutoGainControl {
		t.Error("expected AutoGainControl to be false")
	}
}

func TestVoiceSettingsComponent_MutedUserCount(t *testing.T) {
	c := NewVoiceSettingsComponent()

	c.MuteUser("player-1")
	c.MuteUser("player-2")
	c.MuteUser("player-3")

	if c.GetMutedUserCount() != 3 {
		t.Errorf("expected 3 muted users, got %d", c.GetMutedUserCount())
	}

	muted := c.GetMutedUsers()
	if len(muted) != 3 {
		t.Errorf("expected 3 muted users in list, got %d", len(muted))
	}
}

func TestVoiceSettingsComponent_Serialize(t *testing.T) {
	c := NewVoiceSettingsComponent()
	c.SetMasterVolume(0.7)
	c.MuteUser("player-1")
	c.SetUserVolume("player-2", 1.5)
	c.SetInputDevice("Test Mic")
	c.SetProximityEnabled(false)

	data, err := c.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	c2 := &VoiceSettingsComponent{}
	err = c2.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if c2.MasterVolume != c.MasterVolume {
		t.Errorf("expected MasterVolume %f, got %f", c.MasterVolume, c2.MasterVolume)
	}
	if !c2.IsUserMuted("player-1") {
		t.Error("expected player-1 to be muted")
	}
	if c2.GetUserVolume("player-2") != 1.5 {
		t.Errorf("expected player-2 volume 1.5, got %f", c2.GetUserVolume("player-2"))
	}
	if c2.InputDevice != c.InputDevice {
		t.Errorf("expected InputDevice '%s', got '%s'", c.InputDevice, c2.InputDevice)
	}
	if c2.ProximityEnabled != c.ProximityEnabled {
		t.Errorf("expected ProximityEnabled %v, got %v", c.ProximityEnabled, c2.ProximityEnabled)
	}
}
