// Package engine provides core ECS components and systems for the Venture game.
package engine

import (
	"strconv"

	log "github.com/sirupsen/logrus"
)

// entityIDToString converts an entity ID to a string for map keys.
func entityIDToString(id uint64) string {
	return strconv.FormatUint(id, 10)
}

// VoiceChannelSystem manages voice channel lifecycle and participant synchronization.
type VoiceChannelSystem struct {
	world *World

	// activeChannels tracks all active voice channels by ID.
	activeChannels map[string]*VoiceChannel

	// maxParticipantsPerChannel limits channel capacity.
	maxParticipantsPerChannel int

	// speakingTimeout is how long (in seconds) before speaking indicator times out.
	speakingTimeout float64

	// speakingTimers tracks when each entity last transmitted voice.
	speakingTimers map[string]float64
}

// VoiceChannel represents an active voice channel.
type VoiceChannel struct {
	// ID uniquely identifies the channel.
	ID string

	// Type is the channel type (party, guild, proximity, private).
	Type VoiceChannelType

	// Participants maps entity IDs to their voice state.
	Participants map[string]*VoiceParticipantState

	// LinkedGroupID is the party/guild ID if linked.
	LinkedGroupID string

	// CreatedAt is when the channel was created.
	CreatedAt float64
}

// VoiceParticipantState tracks server-side voice state for a participant.
type VoiceParticipantState struct {
	EntityID    string
	IsMuted     bool    // Server-muted by moderator
	IsDeafened  bool    // Server-deafened by moderator
	IsSpeaking  bool    // Currently transmitting
	LastSpoke   float64 // Time of last voice transmission
	JoinedAt    float64 // Time when joined
	Permissions VoicePermission
}

// NewVoiceChannelSystem creates a new voice channel system.
func NewVoiceChannelSystem(world *World) *VoiceChannelSystem {
	log.WithFields(log.Fields{
		"system_name": "voice_channel",
	}).Debug("Creating voice channel system")

	return &VoiceChannelSystem{
		world:                     world,
		activeChannels:            make(map[string]*VoiceChannel),
		maxParticipantsPerChannel: 50,
		speakingTimeout:           0.5, // 500ms timeout
		speakingTimers:            make(map[string]float64),
	}
}

// Update processes voice channel state for all entities.
func (s *VoiceChannelSystem) Update(entities []*Entity, deltaTime float64) {
	// Update speaking timeouts
	s.updateSpeakingTimeouts(entities, deltaTime)

	// Sync participant lists between entities and channels
	s.syncParticipants(entities)

	// Clean up empty channels
	s.cleanupEmptyChannels()
}

// updateSpeakingTimeouts clears speaking state after timeout.
func (s *VoiceChannelSystem) updateSpeakingTimeouts(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		if !entity.HasComponent("voice_channel") {
			continue
		}

		comp, ok := entity.GetComponent("voice_channel")
		if !ok {
			continue
		}
		vc, ok := comp.(*VoiceChannelComponent)
		if !ok {
			continue
		}

		// Update the entity's speaking timer
		if vc.IsSpeaking {
			s.speakingTimers[entityIDToString(entity.ID)] = 0
		} else if timer, exists := s.speakingTimers[entityIDToString(entity.ID)]; exists {
			timer += deltaTime
			if timer > s.speakingTimeout {
				delete(s.speakingTimers, entityIDToString(entity.ID))
			} else {
				s.speakingTimers[entityIDToString(entity.ID)] = timer
			}
		}
	}
}

// syncParticipants synchronizes participant lists between entities and channels.
func (s *VoiceChannelSystem) syncParticipants(entities []*Entity) {
	// Build map of entity ID to voice component for quick lookup
	entityVoice := make(map[string]*VoiceChannelComponent)
	for _, entity := range entities {
		if !entity.HasComponent("voice_channel") {
			continue
		}
		comp, ok := entity.GetComponent("voice_channel")
		if !ok {
			continue
		}
		vc, ok := comp.(*VoiceChannelComponent)
		if !ok {
			continue
		}
		if vc.IsInChannel() {
			entityVoice[entityIDToString(entity.ID)] = vc
		}
	}

	// Group entities by channel
	channelEntities := make(map[string][]string)
	for entityID, vc := range entityVoice {
		channelEntities[vc.ChannelID] = append(channelEntities[vc.ChannelID], entityID)
	}

	// Update each entity's participant list
	for entityID, vc := range entityVoice {
		participants := channelEntities[vc.ChannelID]

		// Clear current participants and rebuild
		vc.Participants = make([]VoiceParticipant, 0, len(participants)-1)

		for _, otherID := range participants {
			if otherID == entityID {
				continue // Don't include self in participants list
			}

			otherVC := entityVoice[otherID]
			vc.Participants = append(vc.Participants, VoiceParticipant{
				EntityID:   otherID,
				IsMuted:    otherVC.IsServerMuted,
				IsDeafened: otherVC.IsServerDeafened,
				IsSpeaking: otherVC.IsSpeaking,
				JoinedAt:   otherVC.JoinedAt,
			})
		}
	}
}

// cleanupEmptyChannels removes channels with no participants.
func (s *VoiceChannelSystem) cleanupEmptyChannels() {
	for id, channel := range s.activeChannels {
		if len(channel.Participants) == 0 {
			log.WithFields(log.Fields{
				"channel_id": id,
				"type":       channel.Type,
			}).Debug("Removing empty voice channel")
			delete(s.activeChannels, id)
		}
	}
}

// JoinChannel adds an entity to a voice channel.
func (s *VoiceChannelSystem) JoinChannel(entity *Entity, channelID string, channelType VoiceChannelType, permissions VoicePermission) error {
	if entity == nil {
		return ErrNilEntity
	}

	// Get or create the channel
	channel, exists := s.activeChannels[channelID]
	if !exists {
		channel = &VoiceChannel{
			ID:           channelID,
			Type:         channelType,
			Participants: make(map[string]*VoiceParticipantState),
			CreatedAt:    0, // Would use game time
		}
		s.activeChannels[channelID] = channel
	}

	// Check capacity
	if len(channel.Participants) >= s.maxParticipantsPerChannel {
		log.WithFields(log.Fields{
			"channel_id": channelID,
			"entity_id":  entityIDToString(entity.ID),
			"capacity":   s.maxParticipantsPerChannel,
		}).Warn("Voice channel at capacity")
		return ErrChannelFull
	}

	// Add participant to channel
	channel.Participants[entityIDToString(entity.ID)] = &VoiceParticipantState{
		EntityID:    entityIDToString(entity.ID),
		IsMuted:     false,
		IsDeafened:  false,
		IsSpeaking:  false,
		Permissions: permissions,
	}

	// Create or update entity's voice component
	var vc *VoiceChannelComponent
	if entity.HasComponent("voice_channel") {
		comp, _ := entity.GetComponent("voice_channel")
		vc = comp.(*VoiceChannelComponent)
		// Leave previous channel if in one
		if vc.IsInChannel() && vc.ChannelID != channelID {
			s.LeaveChannel(entity)
		}
	} else {
		vc = NewVoiceChannelComponent(channelID, channelType)
		entity.AddComponent(vc)
	}

	vc.ChannelID = channelID
	vc.ChannelType = channelType
	vc.Permissions = VoicePermissions{Flags: permissions}

	log.WithFields(log.Fields{
		"channel_id":   channelID,
		"channel_type": channelType,
		"entity_id":    entityIDToString(entity.ID),
	}).Debug("Entity joined voice channel")

	return nil
}

// LeaveChannel removes an entity from their current voice channel.
func (s *VoiceChannelSystem) LeaveChannel(entity *Entity) error {
	if entity == nil {
		return ErrNilEntity
	}

	if !entity.HasComponent("voice_channel") {
		return nil
	}

	comp, ok := entity.GetComponent("voice_channel")
	if !ok {
		return nil
	}
	vc, ok := comp.(*VoiceChannelComponent)
	if !ok {
		return nil
	}

	channelID := vc.ChannelID
	if channelID == "" {
		return nil
	}

	// Remove from active channel
	if channel, exists := s.activeChannels[channelID]; exists {
		delete(channel.Participants, entityIDToString(entity.ID))
	}

	// Clear speaking timer
	delete(s.speakingTimers, entityIDToString(entity.ID))

	// Clear entity's voice component
	vc.LeaveChannel()

	log.WithFields(log.Fields{
		"channel_id": channelID,
		"entity_id":  entityIDToString(entity.ID),
	}).Debug("Entity left voice channel")

	return nil
}

// MuteParticipant server-mutes a participant in a channel.
func (s *VoiceChannelSystem) MuteParticipant(moderator, target *Entity, muted bool) error {
	if moderator == nil || target == nil {
		return ErrNilEntity
	}

	modComp, ok := moderator.GetComponent("voice_channel")
	if !ok {
		return ErrNotInChannel
	}
	modVC := modComp.(*VoiceChannelComponent)

	if !modVC.Permissions.CanMuteOthers() {
		return ErrNoPermission
	}

	targetComp, ok := target.GetComponent("voice_channel")
	if !ok {
		return ErrNotInChannel
	}
	targetVC := targetComp.(*VoiceChannelComponent)

	if modVC.ChannelID != targetVC.ChannelID {
		return ErrNotInChannel
	}

	targetVC.IsServerMuted = muted
	if muted {
		targetVC.IsSpeaking = false
	}

	// Update in active channel
	if channel, exists := s.activeChannels[modVC.ChannelID]; exists {
		if state, exists := channel.Participants[entityIDToString(target.ID)]; exists {
			state.IsMuted = muted
		}
	}

	log.WithFields(log.Fields{
		"channel_id":   modVC.ChannelID,
		"moderator_id": entityIDToString(moderator.ID),
		"target_id":    entityIDToString(target.ID),
		"muted":        muted,
	}).Debug("Participant mute state changed")

	return nil
}

// DeafenParticipant server-deafens a participant in a channel.
func (s *VoiceChannelSystem) DeafenParticipant(moderator, target *Entity, deafened bool) error {
	if moderator == nil || target == nil {
		return ErrNilEntity
	}

	modComp, ok := moderator.GetComponent("voice_channel")
	if !ok {
		return ErrNotInChannel
	}
	modVC := modComp.(*VoiceChannelComponent)

	if !modVC.Permissions.CanDeafenOthers() {
		return ErrNoPermission
	}

	targetComp, ok := target.GetComponent("voice_channel")
	if !ok {
		return ErrNotInChannel
	}
	targetVC := targetComp.(*VoiceChannelComponent)

	if modVC.ChannelID != targetVC.ChannelID {
		return ErrNotInChannel
	}

	targetVC.IsServerDeafened = deafened

	// Update in active channel
	if channel, exists := s.activeChannels[modVC.ChannelID]; exists {
		if state, exists := channel.Participants[entityIDToString(target.ID)]; exists {
			state.IsDeafened = deafened
		}
	}

	log.WithFields(log.Fields{
		"channel_id":   modVC.ChannelID,
		"moderator_id": entityIDToString(moderator.ID),
		"target_id":    entityIDToString(target.ID),
		"deafened":     deafened,
	}).Debug("Participant deafen state changed")

	return nil
}

// KickParticipant removes a participant from the channel.
func (s *VoiceChannelSystem) KickParticipant(moderator, target *Entity) error {
	if moderator == nil || target == nil {
		return ErrNilEntity
	}

	modComp, ok := moderator.GetComponent("voice_channel")
	if !ok {
		return ErrNotInChannel
	}
	modVC := modComp.(*VoiceChannelComponent)

	if !modVC.Permissions.CanKick() {
		return ErrNoPermission
	}

	targetComp, ok := target.GetComponent("voice_channel")
	if !ok {
		return ErrNotInChannel
	}
	targetVC := targetComp.(*VoiceChannelComponent)

	if modVC.ChannelID != targetVC.ChannelID {
		return ErrNotInChannel
	}

	log.WithFields(log.Fields{
		"channel_id":   modVC.ChannelID,
		"moderator_id": entityIDToString(moderator.ID),
		"target_id":    entityIDToString(target.ID),
	}).Debug("Participant kicked from voice channel")

	return s.LeaveChannel(target)
}

// SetSpeaking updates the speaking state for an entity.
func (s *VoiceChannelSystem) SetSpeaking(entity *Entity, speaking bool) error {
	if entity == nil {
		return ErrNilEntity
	}

	comp, ok := entity.GetComponent("voice_channel")
	if !ok {
		return ErrNotInChannel
	}
	vc := comp.(*VoiceChannelComponent)

	if !vc.IsInChannel() {
		return ErrNotInChannel
	}

	vc.SetSpeaking(speaking)

	// Update in active channel
	if channel, exists := s.activeChannels[vc.ChannelID]; exists {
		if state, exists := channel.Participants[entityIDToString(entity.ID)]; exists {
			state.IsSpeaking = speaking
		}
	}

	return nil
}

// GetChannel returns information about a voice channel.
func (s *VoiceChannelSystem) GetChannel(channelID string) (*VoiceChannel, bool) {
	channel, exists := s.activeChannels[channelID]
	return channel, exists
}

// GetChannelParticipantCount returns the number of participants in a channel.
func (s *VoiceChannelSystem) GetChannelParticipantCount(channelID string) int {
	if channel, exists := s.activeChannels[channelID]; exists {
		return len(channel.Participants)
	}
	return 0
}

// GetActiveChannelCount returns the number of active voice channels.
func (s *VoiceChannelSystem) GetActiveChannelCount() int {
	return len(s.activeChannels)
}

// JoinPartyChannel is a convenience method to join a party's voice channel.
func (s *VoiceChannelSystem) JoinPartyChannel(entity *Entity, partyID string, isLeader bool) error {
	channelID := "party:" + partyID
	permissions := VoicePermDefault
	if isLeader {
		permissions = VoicePermModerator
	}
	return s.JoinChannel(entity, channelID, VoiceChannelParty, permissions)
}

// JoinGuildChannel is a convenience method to join a guild's voice channel.
func (s *VoiceChannelSystem) JoinGuildChannel(entity *Entity, guildID, rank string) error {
	channelID := "guild:" + guildID
	permissions := VoicePermDefault
	if rank == "Officer" || rank == "Leader" {
		permissions = VoicePermModerator
	}
	return s.JoinChannel(entity, channelID, VoiceChannelGuild, permissions)
}

// Error types for voice channel operations.
var (
	ErrNilEntity    = voiceError("nil entity")
	ErrChannelFull  = voiceError("voice channel at capacity")
	ErrNotInChannel = voiceError("entity not in voice channel")
	ErrNoPermission = voiceError("insufficient permissions")
)

type voiceError string

func (e voiceError) Error() string {
	return string(e)
}
