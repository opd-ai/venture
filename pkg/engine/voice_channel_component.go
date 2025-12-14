// Package engine provides core ECS components and systems for the Venture game.
package engine

import (
	"encoding/json"
	"time"
)

// VoiceChannelType defines the type of voice channel.
type VoiceChannelType string

const (
	// VoiceChannelParty is a voice channel for party members.
	VoiceChannelParty VoiceChannelType = "party"
	// VoiceChannelGuild is a voice channel for guild members.
	VoiceChannelGuild VoiceChannelType = "guild"
	// VoiceChannelProximity is a spatial voice channel based on distance.
	VoiceChannelProximity VoiceChannelType = "proximity"
	// VoiceChannelPrivate is a private voice channel between specific players.
	VoiceChannelPrivate VoiceChannelType = "private"
)

// VoicePermission defines permissions for voice channel actions.
type VoicePermission int

const (
	// VoicePermSpeak allows the user to speak in the channel.
	VoicePermSpeak VoicePermission = 1 << iota
	// VoicePermMuteOthers allows the user to mute other participants.
	VoicePermMuteOthers
	// VoicePermDeafenOthers allows the user to deafen other participants.
	VoicePermDeafenOthers
	// VoicePermKick allows the user to kick participants from the channel.
	VoicePermKick
	// VoicePermManage allows the user to manage channel settings.
	VoicePermManage
)

// VoicePermDefault is the default permission set for regular members.
const VoicePermDefault = VoicePermSpeak

// VoicePermModerator is the permission set for moderators.
const VoicePermModerator = VoicePermSpeak | VoicePermMuteOthers

// VoicePermAdmin is the permission set for administrators.
const VoicePermAdmin = VoicePermSpeak | VoicePermMuteOthers | VoicePermDeafenOthers | VoicePermKick | VoicePermManage

// VoicePermissions holds the permission flags for a voice channel participant.
type VoicePermissions struct {
	Flags VoicePermission `json:"flags"`
}

// CanSpeak returns true if the user has permission to speak.
func (p VoicePermissions) CanSpeak() bool {
	return p.Flags&VoicePermSpeak != 0
}

// CanMuteOthers returns true if the user can mute other participants.
func (p VoicePermissions) CanMuteOthers() bool {
	return p.Flags&VoicePermMuteOthers != 0
}

// CanDeafenOthers returns true if the user can deafen other participants.
func (p VoicePermissions) CanDeafenOthers() bool {
	return p.Flags&VoicePermDeafenOthers != 0
}

// CanKick returns true if the user can kick participants.
func (p VoicePermissions) CanKick() bool {
	return p.Flags&VoicePermKick != 0
}

// CanManage returns true if the user can manage channel settings.
func (p VoicePermissions) CanManage() bool {
	return p.Flags&VoicePermManage != 0
}

// VoiceParticipant represents a participant in a voice channel.
type VoiceParticipant struct {
	EntityID   string    `json:"entity_id"`
	IsMuted    bool      `json:"is_muted"`    // Server-side muted by moderator
	IsDeafened bool      `json:"is_deafened"` // Server-side deafened by moderator
	IsSpeaking bool      `json:"is_speaking"` // Currently transmitting voice
	JoinedAt   time.Time `json:"joined_at"`
}

// VoiceChannelComponent tracks an entity's voice channel membership and state.
type VoiceChannelComponent struct {
	// ChannelID uniquely identifies the voice channel.
	ChannelID string `json:"channel_id"`

	// ChannelType defines the type of voice channel.
	ChannelType VoiceChannelType `json:"channel_type"`

	// IsSelfMuted indicates if the player has muted themselves.
	IsSelfMuted bool `json:"is_self_muted"`

	// IsSelfDeafened indicates if the player has deafened themselves.
	IsSelfDeafened bool `json:"is_self_deafened"`

	// IsServerMuted indicates if the player was muted by a moderator.
	IsServerMuted bool `json:"is_server_muted"`

	// IsServerDeafened indicates if the player was deafened by a moderator.
	IsServerDeafened bool `json:"is_server_deafened"`

	// IsSpeaking indicates if the player is currently transmitting voice.
	IsSpeaking bool `json:"is_speaking"`

	// Participants lists other participants in the same channel.
	Participants []VoiceParticipant `json:"participants"`

	// Permissions defines what actions this player can perform.
	Permissions VoicePermissions `json:"permissions"`

	// JoinedAt is when the player joined the channel.
	JoinedAt time.Time `json:"joined_at"`

	// LinkedGroupID is the party/guild ID this channel is linked to (if any).
	LinkedGroupID string `json:"linked_group_id"`
}

// Type returns the component type identifier.
func (c *VoiceChannelComponent) Type() string {
	return "voice_channel"
}

// NewVoiceChannelComponent creates a new voice channel component with default values.
func NewVoiceChannelComponent(channelID string, channelType VoiceChannelType) *VoiceChannelComponent {
	return &VoiceChannelComponent{
		ChannelID:        channelID,
		ChannelType:      channelType,
		IsSelfMuted:      false,
		IsSelfDeafened:   false,
		IsServerMuted:    false,
		IsServerDeafened: false,
		IsSpeaking:       false,
		Participants:     make([]VoiceParticipant, 0),
		Permissions:      VoicePermissions{Flags: VoicePermDefault},
		JoinedAt:         time.Now(),
		LinkedGroupID:    "",
	}
}

// NewPartyVoiceChannelComponent creates a voice channel linked to a party.
func NewPartyVoiceChannelComponent(partyID string) *VoiceChannelComponent {
	c := NewVoiceChannelComponent("party:"+partyID, VoiceChannelParty)
	c.LinkedGroupID = partyID
	return c
}

// NewGuildVoiceChannelComponent creates a voice channel linked to a guild.
func NewGuildVoiceChannelComponent(guildID string, isOfficer bool) *VoiceChannelComponent {
	c := NewVoiceChannelComponent("guild:"+guildID, VoiceChannelGuild)
	c.LinkedGroupID = guildID
	if isOfficer {
		c.Permissions = VoicePermissions{Flags: VoicePermModerator}
	}
	return c
}

// NewProximityVoiceChannelComponent creates a proximity-based voice channel.
func NewProximityVoiceChannelComponent() *VoiceChannelComponent {
	return NewVoiceChannelComponent("proximity", VoiceChannelProximity)
}

// IsMuted returns true if the player is muted (self or server).
func (c *VoiceChannelComponent) IsMuted() bool {
	return c.IsSelfMuted || c.IsServerMuted
}

// IsDeafened returns true if the player is deafened (self or server).
func (c *VoiceChannelComponent) IsDeafened() bool {
	return c.IsSelfDeafened || c.IsServerDeafened
}

// CanTransmit returns true if the player can transmit voice.
func (c *VoiceChannelComponent) CanTransmit() bool {
	return !c.IsMuted() && c.Permissions.CanSpeak()
}

// CanReceive returns true if the player can receive voice.
func (c *VoiceChannelComponent) CanReceive() bool {
	return !c.IsDeafened()
}

// ToggleSelfMute toggles the player's self-mute state.
func (c *VoiceChannelComponent) ToggleSelfMute() {
	c.IsSelfMuted = !c.IsSelfMuted
	if c.IsSelfMuted {
		c.IsSpeaking = false
	}
}

// ToggleSelfDeafen toggles the player's self-deafen state.
func (c *VoiceChannelComponent) ToggleSelfDeafen() {
	c.IsSelfDeafened = !c.IsSelfDeafened
}

// SetSpeaking sets the speaking state.
func (c *VoiceChannelComponent) SetSpeaking(speaking bool) {
	if c.CanTransmit() {
		c.IsSpeaking = speaking
	} else {
		c.IsSpeaking = false
	}
}

// AddParticipant adds a participant to the channel.
func (c *VoiceChannelComponent) AddParticipant(entityID string) {
	for _, p := range c.Participants {
		if p.EntityID == entityID {
			return // Already exists
		}
	}
	c.Participants = append(c.Participants, VoiceParticipant{
		EntityID:   entityID,
		IsMuted:    false,
		IsDeafened: false,
		IsSpeaking: false,
		JoinedAt:   time.Now(),
	})
}

// RemoveParticipant removes a participant from the channel.
func (c *VoiceChannelComponent) RemoveParticipant(entityID string) {
	for i, p := range c.Participants {
		if p.EntityID == entityID {
			c.Participants = append(c.Participants[:i], c.Participants[i+1:]...)
			return
		}
	}
}

// GetParticipant returns a participant by entity ID, or nil if not found.
func (c *VoiceChannelComponent) GetParticipant(entityID string) *VoiceParticipant {
	for i := range c.Participants {
		if c.Participants[i].EntityID == entityID {
			return &c.Participants[i]
		}
	}
	return nil
}

// UpdateParticipantSpeaking updates a participant's speaking state.
func (c *VoiceChannelComponent) UpdateParticipantSpeaking(entityID string, speaking bool) {
	if p := c.GetParticipant(entityID); p != nil {
		p.IsSpeaking = speaking
	}
}

// GetSpeakingParticipants returns a list of entity IDs currently speaking.
func (c *VoiceChannelComponent) GetSpeakingParticipants() []string {
	speaking := make([]string, 0)
	for _, p := range c.Participants {
		if p.IsSpeaking && !p.IsMuted {
			speaking = append(speaking, p.EntityID)
		}
	}
	return speaking
}

// GetParticipantCount returns the number of participants in the channel.
func (c *VoiceChannelComponent) GetParticipantCount() int {
	return len(c.Participants)
}

// IsInChannel returns true if the component is connected to a voice channel.
func (c *VoiceChannelComponent) IsInChannel() bool {
	return c.ChannelID != ""
}

// LeaveChannel clears the channel connection.
func (c *VoiceChannelComponent) LeaveChannel() {
	c.ChannelID = ""
	c.ChannelType = ""
	c.IsSpeaking = false
	c.Participants = make([]VoiceParticipant, 0)
	c.LinkedGroupID = ""
}

// Serialize converts the component to JSON bytes.
func (c *VoiceChannelComponent) Serialize() ([]byte, error) {
	return json.Marshal(c)
}

// Deserialize loads the component from JSON bytes.
func (c *VoiceChannelComponent) Deserialize(data []byte) error {
	return json.Unmarshal(data, c)
}
