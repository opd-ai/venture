package engine

import (
	"testing"
)

func TestNewVoiceChannelSystem(t *testing.T) {
	world := NewWorld()
	system := NewVoiceChannelSystem(world)

	if system == nil {
		t.Fatal("NewVoiceChannelSystem returned nil")
	}
	if system.world != world {
		t.Error("world not set correctly")
	}
	if system.maxParticipantsPerChannel != 50 {
		t.Errorf("expected maxParticipantsPerChannel 50, got %d", system.maxParticipantsPerChannel)
	}
	if system.GetActiveChannelCount() != 0 {
		t.Errorf("expected 0 active channels, got %d", system.GetActiveChannelCount())
	}
}

func TestVoiceChannelSystem_JoinLeaveChannel(t *testing.T) {
	world := NewWorld()
	system := NewVoiceChannelSystem(world)

	entity := world.CreateEntity()

	// Join channel
	err := system.JoinChannel(entity, "test-channel", VoiceChannelParty, VoicePermDefault)
	if err != nil {
		t.Fatalf("JoinChannel failed: %v", err)
	}

	// Verify entity has voice component
	if !entity.HasComponent("voice_channel") {
		t.Error("expected entity to have voice_channel component")
	}

	comp, _ := entity.GetComponent("voice_channel")
	vc := comp.(*VoiceChannelComponent)

	if vc.ChannelID != "test-channel" {
		t.Errorf("expected ChannelID 'test-channel', got '%s'", vc.ChannelID)
	}
	if vc.ChannelType != VoiceChannelParty {
		t.Errorf("expected ChannelType party, got %s", vc.ChannelType)
	}

	// Verify channel was created
	if system.GetActiveChannelCount() != 1 {
		t.Errorf("expected 1 active channel, got %d", system.GetActiveChannelCount())
	}

	channel, exists := system.GetChannel("test-channel")
	if !exists {
		t.Fatal("expected channel to exist")
	}
	if len(channel.Participants) != 1 {
		t.Errorf("expected 1 participant, got %d", len(channel.Participants))
	}

	// Leave channel
	err = system.LeaveChannel(entity)
	if err != nil {
		t.Fatalf("LeaveChannel failed: %v", err)
	}

	if vc.IsInChannel() {
		t.Error("expected entity to not be in channel after leave")
	}
}

func TestVoiceChannelSystem_JoinNilEntity(t *testing.T) {
	world := NewWorld()
	system := NewVoiceChannelSystem(world)

	err := system.JoinChannel(nil, "test", VoiceChannelParty, VoicePermDefault)
	if err != ErrNilEntity {
		t.Errorf("expected ErrNilEntity, got %v", err)
	}
}

func TestVoiceChannelSystem_LeaveNilEntity(t *testing.T) {
	world := NewWorld()
	system := NewVoiceChannelSystem(world)

	err := system.LeaveChannel(nil)
	if err != ErrNilEntity {
		t.Errorf("expected ErrNilEntity, got %v", err)
	}
}

func TestVoiceChannelSystem_LeaveNotInChannel(t *testing.T) {
	world := NewWorld()
	system := NewVoiceChannelSystem(world)

	entity := world.CreateEntity()

	// Leave without being in a channel should not error
	err := system.LeaveChannel(entity)
	if err != nil {
		t.Errorf("expected no error leaving when not in channel, got %v", err)
	}
}

func TestVoiceChannelSystem_JoinSwitchChannel(t *testing.T) {
	world := NewWorld()
	system := NewVoiceChannelSystem(world)

	entity := world.CreateEntity()

	// Join first channel
	err := system.JoinChannel(entity, "channel-1", VoiceChannelParty, VoicePermDefault)
	if err != nil {
		t.Fatalf("JoinChannel 1 failed: %v", err)
	}

	// Join second channel (should leave first)
	err = system.JoinChannel(entity, "channel-2", VoiceChannelGuild, VoicePermDefault)
	if err != nil {
		t.Fatalf("JoinChannel 2 failed: %v", err)
	}

	comp, _ := entity.GetComponent("voice_channel")
	vc := comp.(*VoiceChannelComponent)

	if vc.ChannelID != "channel-2" {
		t.Errorf("expected ChannelID 'channel-2', got '%s'", vc.ChannelID)
	}

	// First channel should have 0 participants after cleanup
	system.cleanupEmptyChannels()
	if system.GetChannelParticipantCount("channel-1") != 0 {
		t.Error("expected channel-1 to have 0 participants")
	}
}

func TestVoiceChannelSystem_ChannelCapacity(t *testing.T) {
	world := NewWorld()
	system := NewVoiceChannelSystem(world)
	system.maxParticipantsPerChannel = 2

	entity1 := world.CreateEntity()
	entity2 := world.CreateEntity()
	entity3 := world.CreateEntity()

	// Fill channel to capacity
	system.JoinChannel(entity1, "test", VoiceChannelParty, VoicePermDefault)
	system.JoinChannel(entity2, "test", VoiceChannelParty, VoicePermDefault)

	// Third join should fail
	err := system.JoinChannel(entity3, "test", VoiceChannelParty, VoicePermDefault)
	if err != ErrChannelFull {
		t.Errorf("expected ErrChannelFull, got %v", err)
	}
}

func TestVoiceChannelSystem_MuteParticipant(t *testing.T) {
	world := NewWorld()
	system := NewVoiceChannelSystem(world)

	moderator := world.CreateEntity()
	target := world.CreateEntity()

	// Join channel with moderator permissions
	system.JoinChannel(moderator, "test", VoiceChannelParty, VoicePermModerator)
	system.JoinChannel(target, "test", VoiceChannelParty, VoicePermDefault)

	// Mute target
	err := system.MuteParticipant(moderator, target, true)
	if err != nil {
		t.Fatalf("MuteParticipant failed: %v", err)
	}

	targetComp, _ := target.GetComponent("voice_channel")
	targetVC := targetComp.(*VoiceChannelComponent)

	if !targetVC.IsServerMuted {
		t.Error("expected target to be server muted")
	}

	// Unmute
	err = system.MuteParticipant(moderator, target, false)
	if err != nil {
		t.Fatalf("UnmuteParticipant failed: %v", err)
	}

	if targetVC.IsServerMuted {
		t.Error("expected target to not be server muted after unmute")
	}
}

func TestVoiceChannelSystem_MuteNoPermission(t *testing.T) {
	world := NewWorld()
	system := NewVoiceChannelSystem(world)

	user1 := world.CreateEntity()
	user2 := world.CreateEntity()

	// Both join with default permissions
	system.JoinChannel(user1, "test", VoiceChannelParty, VoicePermDefault)
	system.JoinChannel(user2, "test", VoiceChannelParty, VoicePermDefault)

	// user1 tries to mute user2 without permission
	err := system.MuteParticipant(user1, user2, true)
	if err != ErrNoPermission {
		t.Errorf("expected ErrNoPermission, got %v", err)
	}
}

func TestVoiceChannelSystem_MuteDifferentChannel(t *testing.T) {
	world := NewWorld()
	system := NewVoiceChannelSystem(world)

	moderator := world.CreateEntity()
	target := world.CreateEntity()

	// Join different channels
	system.JoinChannel(moderator, "channel-1", VoiceChannelParty, VoicePermModerator)
	system.JoinChannel(target, "channel-2", VoiceChannelParty, VoicePermDefault)

	// Cannot mute across channels
	err := system.MuteParticipant(moderator, target, true)
	if err != ErrNotInChannel {
		t.Errorf("expected ErrNotInChannel, got %v", err)
	}
}

func TestVoiceChannelSystem_DeafenParticipant(t *testing.T) {
	world := NewWorld()
	system := NewVoiceChannelSystem(world)

	admin := world.CreateEntity()
	target := world.CreateEntity()

	// Join channel with admin permissions
	system.JoinChannel(admin, "test", VoiceChannelParty, VoicePermAdmin)
	system.JoinChannel(target, "test", VoiceChannelParty, VoicePermDefault)

	// Deafen target
	err := system.DeafenParticipant(admin, target, true)
	if err != nil {
		t.Fatalf("DeafenParticipant failed: %v", err)
	}

	targetComp, _ := target.GetComponent("voice_channel")
	targetVC := targetComp.(*VoiceChannelComponent)

	if !targetVC.IsServerDeafened {
		t.Error("expected target to be server deafened")
	}
}

func TestVoiceChannelSystem_DeafenNoPermission(t *testing.T) {
	world := NewWorld()
	system := NewVoiceChannelSystem(world)

	// Moderator can mute but not deafen
	moderator := world.CreateEntity()
	target := world.CreateEntity()

	system.JoinChannel(moderator, "test", VoiceChannelParty, VoicePermModerator)
	system.JoinChannel(target, "test", VoiceChannelParty, VoicePermDefault)

	err := system.DeafenParticipant(moderator, target, true)
	if err != ErrNoPermission {
		t.Errorf("expected ErrNoPermission, got %v", err)
	}
}

func TestVoiceChannelSystem_KickParticipant(t *testing.T) {
	world := NewWorld()
	system := NewVoiceChannelSystem(world)

	admin := world.CreateEntity()
	target := world.CreateEntity()

	system.JoinChannel(admin, "test", VoiceChannelParty, VoicePermAdmin)
	system.JoinChannel(target, "test", VoiceChannelParty, VoicePermDefault)

	// Kick target
	err := system.KickParticipant(admin, target)
	if err != nil {
		t.Fatalf("KickParticipant failed: %v", err)
	}

	targetComp, _ := target.GetComponent("voice_channel")
	targetVC := targetComp.(*VoiceChannelComponent)

	if targetVC.IsInChannel() {
		t.Error("expected target to not be in channel after kick")
	}
}

func TestVoiceChannelSystem_KickNoPermission(t *testing.T) {
	world := NewWorld()
	system := NewVoiceChannelSystem(world)

	// Moderator cannot kick
	moderator := world.CreateEntity()
	target := world.CreateEntity()

	system.JoinChannel(moderator, "test", VoiceChannelParty, VoicePermModerator)
	system.JoinChannel(target, "test", VoiceChannelParty, VoicePermDefault)

	err := system.KickParticipant(moderator, target)
	if err != ErrNoPermission {
		t.Errorf("expected ErrNoPermission, got %v", err)
	}
}

func TestVoiceChannelSystem_SetSpeaking(t *testing.T) {
	world := NewWorld()
	system := NewVoiceChannelSystem(world)

	entity := world.CreateEntity()
	system.JoinChannel(entity, "test", VoiceChannelParty, VoicePermDefault)

	// Set speaking
	err := system.SetSpeaking(entity, true)
	if err != nil {
		t.Fatalf("SetSpeaking failed: %v", err)
	}

	comp, _ := entity.GetComponent("voice_channel")
	vc := comp.(*VoiceChannelComponent)

	if !vc.IsSpeaking {
		t.Error("expected IsSpeaking to be true")
	}

	// Stop speaking
	err = system.SetSpeaking(entity, false)
	if err != nil {
		t.Fatalf("SetSpeaking(false) failed: %v", err)
	}

	if vc.IsSpeaking {
		t.Error("expected IsSpeaking to be false")
	}
}

func TestVoiceChannelSystem_SetSpeakingNotInChannel(t *testing.T) {
	world := NewWorld()
	system := NewVoiceChannelSystem(world)

	entity := world.CreateEntity()

	err := system.SetSpeaking(entity, true)
	if err != ErrNotInChannel {
		t.Errorf("expected ErrNotInChannel, got %v", err)
	}
}

func TestVoiceChannelSystem_Update(t *testing.T) {
	world := NewWorld()
	system := NewVoiceChannelSystem(world)

	entity1 := world.CreateEntity()
	entity2 := world.CreateEntity()

	system.JoinChannel(entity1, "test", VoiceChannelParty, VoicePermDefault)
	system.JoinChannel(entity2, "test", VoiceChannelParty, VoicePermDefault)

	entities := []*Entity{entity1, entity2}

	// Run update
	system.Update(entities, 0.016)

	// Check that participant lists are synced
	comp1, _ := entity1.GetComponent("voice_channel")
	vc1 := comp1.(*VoiceChannelComponent)

	comp2, _ := entity2.GetComponent("voice_channel")
	vc2 := comp2.(*VoiceChannelComponent)

	// Entity1 should see Entity2 as participant
	if len(vc1.Participants) != 1 {
		t.Errorf("expected 1 participant for entity1, got %d", len(vc1.Participants))
	}
	if len(vc1.Participants) > 0 && vc1.Participants[0].EntityID != entityIDToString(entity2.ID) {
		t.Errorf("expected participant to be entity2")
	}

	// Entity2 should see Entity1 as participant
	if len(vc2.Participants) != 1 {
		t.Errorf("expected 1 participant for entity2, got %d", len(vc2.Participants))
	}
}

func TestVoiceChannelSystem_CleanupEmptyChannels(t *testing.T) {
	world := NewWorld()
	system := NewVoiceChannelSystem(world)

	entity := world.CreateEntity()
	system.JoinChannel(entity, "test", VoiceChannelParty, VoicePermDefault)

	if system.GetActiveChannelCount() != 1 {
		t.Fatalf("expected 1 active channel, got %d", system.GetActiveChannelCount())
	}

	// Leave and cleanup
	system.LeaveChannel(entity)
	system.cleanupEmptyChannels()

	if system.GetActiveChannelCount() != 0 {
		t.Errorf("expected 0 active channels after cleanup, got %d", system.GetActiveChannelCount())
	}
}

func TestVoiceChannelSystem_JoinPartyChannel(t *testing.T) {
	world := NewWorld()
	system := NewVoiceChannelSystem(world)

	tests := []struct {
		name     string
		isLeader bool
		wantPerm VoicePermission
	}{
		{"regular member", false, VoicePermDefault},
		{"party leader", true, VoicePermModerator},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := world.CreateEntity()
			err := system.JoinPartyChannel(entity, "party-123", tt.isLeader)
			if err != nil {
				t.Fatalf("JoinPartyChannel failed: %v", err)
			}

			comp, _ := entity.GetComponent("voice_channel")
			vc := comp.(*VoiceChannelComponent)

			if vc.ChannelID != "party:party-123" {
				t.Errorf("expected ChannelID 'party:party-123', got '%s'", vc.ChannelID)
			}
			if vc.ChannelType != VoiceChannelParty {
				t.Errorf("expected ChannelType party, got %s", vc.ChannelType)
			}
			if vc.Permissions.Flags != tt.wantPerm {
				t.Errorf("expected permissions %d, got %d", tt.wantPerm, vc.Permissions.Flags)
			}
		})
	}
}

func TestVoiceChannelSystem_JoinGuildChannel(t *testing.T) {
	world := NewWorld()
	system := NewVoiceChannelSystem(world)

	tests := []struct {
		name     string
		rank     string
		wantPerm VoicePermission
	}{
		{"recruit", "Recruit", VoicePermDefault},
		{"member", "Member", VoicePermDefault},
		{"officer", "Officer", VoicePermModerator},
		{"leader", "Leader", VoicePermModerator},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := world.CreateEntity()
			err := system.JoinGuildChannel(entity, "guild-456", tt.rank)
			if err != nil {
				t.Fatalf("JoinGuildChannel failed: %v", err)
			}

			comp, _ := entity.GetComponent("voice_channel")
			vc := comp.(*VoiceChannelComponent)

			if vc.ChannelID != "guild:guild-456" {
				t.Errorf("expected ChannelID 'guild:guild-456', got '%s'", vc.ChannelID)
			}
			if vc.Permissions.Flags != tt.wantPerm {
				t.Errorf("expected permissions %d, got %d", tt.wantPerm, vc.Permissions.Flags)
			}
		})
	}
}

func TestVoiceChannelSystem_SpeakingTimeout(t *testing.T) {
	world := NewWorld()
	system := NewVoiceChannelSystem(world)
	system.speakingTimeout = 0.1 // 100ms for testing

	entity := world.CreateEntity()
	system.JoinChannel(entity, "test", VoiceChannelParty, VoicePermDefault)

	comp, _ := entity.GetComponent("voice_channel")
	vc := comp.(*VoiceChannelComponent)

	// Start speaking
	vc.SetSpeaking(true)
	entities := []*Entity{entity}

	// First update should not clear speaking timer
	system.Update(entities, 0.05)

	// Set speaking to false, timer starts
	vc.SetSpeaking(false)
	system.Update(entities, 0.05)

	// Timer should still exist
	if _, exists := system.speakingTimers[entityIDToString(entity.ID)]; !exists {
		t.Error("expected speaking timer to exist")
	}

	// Wait past timeout
	system.Update(entities, 0.1)

	// Timer should be cleared
	if _, exists := system.speakingTimers[entityIDToString(entity.ID)]; exists {
		t.Error("expected speaking timer to be cleared after timeout")
	}
}

func TestVoiceChannelSystem_NilEntityOperations(t *testing.T) {
	world := NewWorld()
	system := NewVoiceChannelSystem(world)

	if err := system.MuteParticipant(nil, nil, true); err != ErrNilEntity {
		t.Errorf("MuteParticipant nil: expected ErrNilEntity, got %v", err)
	}
	if err := system.DeafenParticipant(nil, nil, true); err != ErrNilEntity {
		t.Errorf("DeafenParticipant nil: expected ErrNilEntity, got %v", err)
	}
	if err := system.KickParticipant(nil, nil); err != ErrNilEntity {
		t.Errorf("KickParticipant nil: expected ErrNilEntity, got %v", err)
	}
	if err := system.SetSpeaking(nil, true); err != ErrNilEntity {
		t.Errorf("SetSpeaking nil: expected ErrNilEntity, got %v", err)
	}
}

func TestVoiceError(t *testing.T) {
	err := ErrChannelFull
	if err.Error() != "voice channel at capacity" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestVoiceChannelSystem_GetChannel(t *testing.T) {
	world := NewWorld()
	system := NewVoiceChannelSystem(world)

	// Non-existent channel
	_, exists := system.GetChannel("nonexistent")
	if exists {
		t.Error("expected channel to not exist")
	}

	// Create channel
	entity := world.CreateEntity()
	system.JoinChannel(entity, "test", VoiceChannelParty, VoicePermDefault)

	channel, exists := system.GetChannel("test")
	if !exists {
		t.Error("expected channel to exist")
	}
	if channel.ID != "test" {
		t.Errorf("expected channel ID 'test', got '%s'", channel.ID)
	}
	if channel.Type != VoiceChannelParty {
		t.Errorf("expected channel type party, got %s", channel.Type)
	}
}

func TestVoiceChannelSystem_GetChannelParticipantCount(t *testing.T) {
	world := NewWorld()
	system := NewVoiceChannelSystem(world)

	// Non-existent channel
	count := system.GetChannelParticipantCount("nonexistent")
	if count != 0 {
		t.Errorf("expected 0 for non-existent channel, got %d", count)
	}

	// Create channel with participants
	entity1 := world.CreateEntity()
	entity2 := world.CreateEntity()
	system.JoinChannel(entity1, "test", VoiceChannelParty, VoicePermDefault)
	system.JoinChannel(entity2, "test", VoiceChannelParty, VoicePermDefault)

	count = system.GetChannelParticipantCount("test")
	if count != 2 {
		t.Errorf("expected 2 participants, got %d", count)
	}
}
