package engine

import (
	"testing"
	"time"
)

func TestVoiceChannelComponent_Type(t *testing.T) {
	c := NewVoiceChannelComponent("test-channel", VoiceChannelParty)
	if c.Type() != "voice_channel" {
		t.Errorf("expected type 'voice_channel', got '%s'", c.Type())
	}
}

func TestNewVoiceChannelComponent(t *testing.T) {
	tests := []struct {
		name        string
		channelID   string
		channelType VoiceChannelType
	}{
		{"party channel", "party:123", VoiceChannelParty},
		{"guild channel", "guild:456", VoiceChannelGuild},
		{"proximity channel", "proximity", VoiceChannelProximity},
		{"private channel", "private:789", VoiceChannelPrivate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewVoiceChannelComponent(tt.channelID, tt.channelType)

			if c.ChannelID != tt.channelID {
				t.Errorf("expected ChannelID %s, got %s", tt.channelID, c.ChannelID)
			}
			if c.ChannelType != tt.channelType {
				t.Errorf("expected ChannelType %s, got %s", tt.channelType, c.ChannelType)
			}
			if c.IsSelfMuted {
				t.Error("expected IsSelfMuted to be false")
			}
			if c.IsSelfDeafened {
				t.Error("expected IsSelfDeafened to be false")
			}
			if c.IsSpeaking {
				t.Error("expected IsSpeaking to be false")
			}
			if len(c.Participants) != 0 {
				t.Errorf("expected 0 participants, got %d", len(c.Participants))
			}
			if c.Permissions.Flags != VoicePermDefault {
				t.Errorf("expected default permissions, got %d", c.Permissions.Flags)
			}
		})
	}
}

func TestNewPartyVoiceChannelComponent(t *testing.T) {
	c := NewPartyVoiceChannelComponent("party-123")

	if c.ChannelID != "party:party-123" {
		t.Errorf("expected ChannelID 'party:party-123', got '%s'", c.ChannelID)
	}
	if c.ChannelType != VoiceChannelParty {
		t.Errorf("expected ChannelType party, got %s", c.ChannelType)
	}
	if c.LinkedGroupID != "party-123" {
		t.Errorf("expected LinkedGroupID 'party-123', got '%s'", c.LinkedGroupID)
	}
}

func TestNewGuildVoiceChannelComponent(t *testing.T) {
	tests := []struct {
		name      string
		guildID   string
		isOfficer bool
		wantPerm  VoicePermission
	}{
		{"regular member", "guild-1", false, VoicePermDefault},
		{"officer", "guild-2", true, VoicePermModerator},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewGuildVoiceChannelComponent(tt.guildID, tt.isOfficer)

			if c.ChannelType != VoiceChannelGuild {
				t.Errorf("expected ChannelType guild, got %s", c.ChannelType)
			}
			if c.LinkedGroupID != tt.guildID {
				t.Errorf("expected LinkedGroupID '%s', got '%s'", tt.guildID, c.LinkedGroupID)
			}
			if c.Permissions.Flags != tt.wantPerm {
				t.Errorf("expected permissions %d, got %d", tt.wantPerm, c.Permissions.Flags)
			}
		})
	}
}

func TestNewProximityVoiceChannelComponent(t *testing.T) {
	c := NewProximityVoiceChannelComponent()

	if c.ChannelID != "proximity" {
		t.Errorf("expected ChannelID 'proximity', got '%s'", c.ChannelID)
	}
	if c.ChannelType != VoiceChannelProximity {
		t.Errorf("expected ChannelType proximity, got %s", c.ChannelType)
	}
}

func TestVoiceChannelComponent_MuteDeafen(t *testing.T) {
	c := NewVoiceChannelComponent("test", VoiceChannelParty)

	// Initially not muted or deafened
	if c.IsMuted() {
		t.Error("expected IsMuted to be false initially")
	}
	if c.IsDeafened() {
		t.Error("expected IsDeafened to be false initially")
	}

	// Self mute
	c.ToggleSelfMute()
	if !c.IsSelfMuted {
		t.Error("expected IsSelfMuted to be true after toggle")
	}
	if !c.IsMuted() {
		t.Error("expected IsMuted to return true after self mute")
	}

	// Self unmute
	c.ToggleSelfMute()
	if c.IsSelfMuted {
		t.Error("expected IsSelfMuted to be false after second toggle")
	}

	// Server mute
	c.IsServerMuted = true
	if !c.IsMuted() {
		t.Error("expected IsMuted to return true with server mute")
	}

	// Self deafen
	c.ToggleSelfDeafen()
	if !c.IsSelfDeafened {
		t.Error("expected IsSelfDeafened to be true after toggle")
	}
	if !c.IsDeafened() {
		t.Error("expected IsDeafened to return true after self deafen")
	}

	// Server deafen
	c.IsSelfDeafened = false
	c.IsServerDeafened = true
	if !c.IsDeafened() {
		t.Error("expected IsDeafened to return true with server deafen")
	}
}

func TestVoiceChannelComponent_CanTransmitReceive(t *testing.T) {
	c := NewVoiceChannelComponent("test", VoiceChannelParty)

	// Initially can transmit and receive
	if !c.CanTransmit() {
		t.Error("expected CanTransmit to be true initially")
	}
	if !c.CanReceive() {
		t.Error("expected CanReceive to be true initially")
	}

	// Muted user cannot transmit
	c.ToggleSelfMute()
	if c.CanTransmit() {
		t.Error("expected CanTransmit to be false when muted")
	}
	if !c.CanReceive() {
		t.Error("expected CanReceive to still be true when muted")
	}

	// Unmute and deafen
	c.ToggleSelfMute()
	c.ToggleSelfDeafen()
	if !c.CanTransmit() {
		t.Error("expected CanTransmit to be true when deafened")
	}
	if c.CanReceive() {
		t.Error("expected CanReceive to be false when deafened")
	}

	// No speak permission
	c.ToggleSelfDeafen()
	c.Permissions = VoicePermissions{Flags: 0}
	if c.CanTransmit() {
		t.Error("expected CanTransmit to be false without speak permission")
	}
}

func TestVoiceChannelComponent_Speaking(t *testing.T) {
	c := NewVoiceChannelComponent("test", VoiceChannelParty)

	// Set speaking
	c.SetSpeaking(true)
	if !c.IsSpeaking {
		t.Error("expected IsSpeaking to be true")
	}

	// Stop speaking
	c.SetSpeaking(false)
	if c.IsSpeaking {
		t.Error("expected IsSpeaking to be false")
	}

	// Muted user cannot speak
	c.ToggleSelfMute()
	c.SetSpeaking(true)
	if c.IsSpeaking {
		t.Error("expected IsSpeaking to be false when muted")
	}

	// Muting while speaking stops speaking
	c.ToggleSelfMute()
	c.SetSpeaking(true)
	c.ToggleSelfMute() // This should set IsSpeaking to false
	if c.IsSpeaking {
		t.Error("expected IsSpeaking to be false after mute toggle while speaking")
	}
}

func TestVoiceChannelComponent_Participants(t *testing.T) {
	c := NewVoiceChannelComponent("test", VoiceChannelParty)

	// Add participants
	c.AddParticipant("player-1")
	c.AddParticipant("player-2")
	c.AddParticipant("player-3")

	if c.GetParticipantCount() != 3 {
		t.Errorf("expected 3 participants, got %d", c.GetParticipantCount())
	}

	// Add duplicate (should be ignored)
	c.AddParticipant("player-1")
	if c.GetParticipantCount() != 3 {
		t.Errorf("expected still 3 participants after duplicate add, got %d", c.GetParticipantCount())
	}

	// Get participant
	p := c.GetParticipant("player-2")
	if p == nil {
		t.Fatal("expected to find participant player-2")
	}
	if p.EntityID != "player-2" {
		t.Errorf("expected EntityID 'player-2', got '%s'", p.EntityID)
	}

	// Get non-existent participant
	p = c.GetParticipant("player-99")
	if p != nil {
		t.Error("expected nil for non-existent participant")
	}

	// Remove participant
	c.RemoveParticipant("player-2")
	if c.GetParticipantCount() != 2 {
		t.Errorf("expected 2 participants after removal, got %d", c.GetParticipantCount())
	}
	if c.GetParticipant("player-2") != nil {
		t.Error("expected player-2 to be removed")
	}

	// Remove non-existent (should not error)
	c.RemoveParticipant("player-99")
	if c.GetParticipantCount() != 2 {
		t.Errorf("expected still 2 participants, got %d", c.GetParticipantCount())
	}
}

func TestVoiceChannelComponent_SpeakingParticipants(t *testing.T) {
	c := NewVoiceChannelComponent("test", VoiceChannelParty)

	c.AddParticipant("player-1")
	c.AddParticipant("player-2")
	c.AddParticipant("player-3")

	// No one speaking initially
	speaking := c.GetSpeakingParticipants()
	if len(speaking) != 0 {
		t.Errorf("expected 0 speaking participants, got %d", len(speaking))
	}

	// Update speaking states
	c.UpdateParticipantSpeaking("player-1", true)
	c.UpdateParticipantSpeaking("player-3", true)

	speaking = c.GetSpeakingParticipants()
	if len(speaking) != 2 {
		t.Errorf("expected 2 speaking participants, got %d", len(speaking))
	}

	// Muted participant should not appear in speaking list
	p := c.GetParticipant("player-1")
	p.IsMuted = true
	speaking = c.GetSpeakingParticipants()
	if len(speaking) != 1 {
		t.Errorf("expected 1 speaking participant after mute, got %d", len(speaking))
	}

	// Update non-existent participant (should not error)
	c.UpdateParticipantSpeaking("player-99", true)
}

func TestVoiceChannelComponent_IsInChannel(t *testing.T) {
	c := NewVoiceChannelComponent("test-channel", VoiceChannelParty)

	if !c.IsInChannel() {
		t.Error("expected IsInChannel to be true when channel ID is set")
	}

	c.LeaveChannel()
	if c.IsInChannel() {
		t.Error("expected IsInChannel to be false after LeaveChannel")
	}
	if c.ChannelType != "" {
		t.Error("expected ChannelType to be empty after LeaveChannel")
	}
	if len(c.Participants) != 0 {
		t.Error("expected Participants to be empty after LeaveChannel")
	}
}

func TestVoiceChannelComponent_Serialize(t *testing.T) {
	c := NewVoiceChannelComponent("test-channel", VoiceChannelGuild)
	c.AddParticipant("player-1")
	c.IsSelfMuted = true

	data, err := c.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	c2 := &VoiceChannelComponent{}
	err = c2.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if c2.ChannelID != c.ChannelID {
		t.Errorf("expected ChannelID '%s', got '%s'", c.ChannelID, c2.ChannelID)
	}
	if c2.ChannelType != c.ChannelType {
		t.Errorf("expected ChannelType '%s', got '%s'", c.ChannelType, c2.ChannelType)
	}
	if c2.IsSelfMuted != c.IsSelfMuted {
		t.Errorf("expected IsSelfMuted %v, got %v", c.IsSelfMuted, c2.IsSelfMuted)
	}
	if c2.GetParticipantCount() != c.GetParticipantCount() {
		t.Errorf("expected %d participants, got %d", c.GetParticipantCount(), c2.GetParticipantCount())
	}
}

func TestVoicePermissions(t *testing.T) {
	tests := []struct {
		name   string
		flags  VoicePermission
		speak  bool
		mute   bool
		deafen bool
		kick   bool
		manage bool
	}{
		{"default", VoicePermDefault, true, false, false, false, false},
		{"moderator", VoicePermModerator, true, true, false, false, false},
		{"admin", VoicePermAdmin, true, true, true, true, true},
		{"none", 0, false, false, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := VoicePermissions{Flags: tt.flags}

			if p.CanSpeak() != tt.speak {
				t.Errorf("CanSpeak: expected %v, got %v", tt.speak, p.CanSpeak())
			}
			if p.CanMuteOthers() != tt.mute {
				t.Errorf("CanMuteOthers: expected %v, got %v", tt.mute, p.CanMuteOthers())
			}
			if p.CanDeafenOthers() != tt.deafen {
				t.Errorf("CanDeafenOthers: expected %v, got %v", tt.deafen, p.CanDeafenOthers())
			}
			if p.CanKick() != tt.kick {
				t.Errorf("CanKick: expected %v, got %v", tt.kick, p.CanKick())
			}
			if p.CanManage() != tt.manage {
				t.Errorf("CanManage: expected %v, got %v", tt.manage, p.CanManage())
			}
		})
	}
}

func TestVoiceParticipant_JoinedAt(t *testing.T) {
	c := NewVoiceChannelComponent("test", VoiceChannelParty)
	beforeAdd := time.Now()
	c.AddParticipant("player-1")
	afterAdd := time.Now()

	p := c.GetParticipant("player-1")
	if p == nil {
		t.Fatal("expected to find participant")
	}

	if p.JoinedAt.Before(beforeAdd) || p.JoinedAt.After(afterAdd) {
		t.Errorf("JoinedAt %v should be between %v and %v", p.JoinedAt, beforeAdd, afterAdd)
	}
}
