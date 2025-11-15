package engine

import "time"

// ChatChannel represents different chat channels
type ChatChannel int

const (
	// ChatGlobal is visible to all players on the server.
	// Rate limit: 1 message per 3 seconds.
	ChatGlobal ChatChannel = iota

	// ChatLocal is visible to players within 10-tile radius.
	// Rate limit: 1 message per 1 second.
	ChatLocal

	// ChatParty is visible to party members only.
	// Rate limit: 1 message per 0.5 seconds.
	ChatParty

	// ChatWhisper is a direct message to a specific player.
	// Rate limit: 1 message per 0.5 seconds.
	ChatWhisper
)

// String returns the string representation of the ChatChannel.
func (c ChatChannel) String() string {
	switch c {
	case ChatGlobal:
		return "Global"
	case ChatLocal:
		return "Local"
	case ChatParty:
		return "Party"
	case ChatWhisper:
		return "Whisper"
	default:
		return "Unknown"
	}
}

// ChatMessage represents a single chat message
type ChatMessage struct {
	ID          string      // Unique message ID (UUID)
	SenderID    uint64      // Entity ID of sender
	SenderName  string      // Display name of sender
	RecipientID uint64      // For whispers, 0 for other channels
	Channel     ChatChannel // Channel type
	Content     string      // Message content (plaintext for display)
	Encrypted   []byte      // Encrypted content for network transmission
	Timestamp   time.Time   // When message was sent
	Delivered   bool        // Whether message has been acknowledged
	Failed      bool        // Whether message delivery failed
}

// ChatComponent tracks chat state for an entity.
type ChatComponent struct {
	Messages        []ChatMessage             // Message history (last 100 messages)
	UnreadCount     int                       // Number of unread messages
	ActiveChannels  []ChatChannel             // Subscribed channels
	MuteExpiry      time.Time                 // When current mute expires
	ViolationCount  int                       // Number of rate limit violations
	LastMessageTime map[ChatChannel]time.Time // Last message time per channel for rate limiting
	MaxHistorySize  int                       // Maximum messages to store (default 100)

	// Range extension effects
	MegaphoneActive    bool    // Whether megaphone is active (30-tile radius)
	WalkieTalkieActive bool    // Whether walkie-talkie is active (unlimited range)
	MegaphoneUses      int     // Remaining megaphone uses
	LocalRadius        float64 // Current local chat radius (default 10.0)
}

// Type returns the component type
func (c *ChatComponent) Type() string {
	return "chat"
}

// NewChatComponent creates a new chat component with default settings.
func NewChatComponent() *ChatComponent {
	return &ChatComponent{
		Messages:        make([]ChatMessage, 0, 100),
		UnreadCount:     0,
		ActiveChannels:  []ChatChannel{ChatGlobal, ChatLocal},
		LastMessageTime: make(map[ChatChannel]time.Time),
		MaxHistorySize:  100,
		LocalRadius:     10.0, // Default 10-tile radius for local chat
	}
}

// AddMessage adds a message to the history and increments unread count.
func (c *ChatComponent) AddMessage(msg ChatMessage) {
	// Add message
	c.Messages = append(c.Messages, msg)
	c.UnreadCount++

	// Trim history if exceeds max size
	if len(c.Messages) > c.MaxHistorySize {
		c.Messages = c.Messages[len(c.Messages)-c.MaxHistorySize:]
	}
}

// MarkAllRead marks all messages as read (resets unread count).
func (c *ChatComponent) MarkAllRead() {
	c.UnreadCount = 0
}

// GetMessageByID retrieves a message by ID, or nil if not found.
func (c *ChatComponent) GetMessageByID(id string) *ChatMessage {
	for i := range c.Messages {
		if c.Messages[i].ID == id {
			return &c.Messages[i]
		}
	}
	return nil
}

// GetMessagesForChannel returns all messages in a specific channel.
func (c *ChatComponent) GetMessagesForChannel(channel ChatChannel) []ChatMessage {
	result := make([]ChatMessage, 0)
	for _, msg := range c.Messages {
		if msg.Channel == channel {
			result = append(result, msg)
		}
	}
	return result
}

// IsChannelActive returns whether an entity is subscribed to a channel.
func (c *ChatComponent) IsChannelActive(channel ChatChannel) bool {
	for _, active := range c.ActiveChannels {
		if active == channel {
			return true
		}
	}
	return false
}

// SubscribeChannel adds a channel to the active channels list.
func (c *ChatComponent) SubscribeChannel(channel ChatChannel) {
	if !c.IsChannelActive(channel) {
		c.ActiveChannels = append(c.ActiveChannels, channel)
	}
}

// UnsubscribeChannel removes a channel from the active channels list.
func (c *ChatComponent) UnsubscribeChannel(channel ChatChannel) {
	for i, active := range c.ActiveChannels {
		if active == channel {
			c.ActiveChannels = append(c.ActiveChannels[:i], c.ActiveChannels[i+1:]...)
			return
		}
	}
}

// IsMuted returns whether the entity is currently muted.
func (c *ChatComponent) IsMuted() bool {
	return time.Now().Before(c.MuteExpiry)
}

// ApplyMute applies a mute for the specified duration.
// Violations double the duration: 30s → 60s → 120s (max 10 minutes).
func (c *ChatComponent) ApplyMute(duration time.Duration) {
	c.ViolationCount++

	// Double duration for each violation (max 10 minutes)
	muteDuration := duration
	for i := 1; i < c.ViolationCount; i++ {
		muteDuration *= 2
		if muteDuration > 10*time.Minute {
			muteDuration = 10 * time.Minute
			break
		}
	}

	c.MuteExpiry = time.Now().Add(muteDuration)
}

// CanSendMessage checks if entity can send to a channel (not muted, rate limit OK).
func (c *ChatComponent) CanSendMessage(channel ChatChannel) bool {
	if c.IsMuted() {
		return false
	}

	// Check rate limit for channel
	lastTime, exists := c.LastMessageTime[channel]
	if !exists {
		return true
	}

	cooldown := c.GetChannelCooldown(channel)
	return time.Since(lastTime) >= cooldown
}

// GetChannelCooldown returns the cooldown duration for a channel.
func (c *ChatComponent) GetChannelCooldown(channel ChatChannel) time.Duration {
	switch channel {
	case ChatGlobal:
		return 3 * time.Second
	case ChatLocal:
		return 1 * time.Second
	case ChatParty:
		return 500 * time.Millisecond
	case ChatWhisper:
		return 500 * time.Millisecond
	default:
		return 1 * time.Second
	}
}

// RecordMessageSent updates the last message time for a channel.
func (c *ChatComponent) RecordMessageSent(channel ChatChannel) {
	c.LastMessageTime[channel] = time.Now()
}

// ActivateMegaphone activates megaphone effect (30-tile radius, 10 uses).
func (c *ChatComponent) ActivateMegaphone() bool {
	if c.MegaphoneUses > 0 {
		c.MegaphoneActive = true
		c.LocalRadius = 30.0
		c.MegaphoneUses--
		return true
	}
	return false
}

// DeactivateMegaphone deactivates megaphone effect (returns to default radius).
func (c *ChatComponent) DeactivateMegaphone() {
	c.MegaphoneActive = false
	if !c.WalkieTalkieActive {
		c.LocalRadius = 10.0
	}
}

// ActivateWalkieTalkie activates walkie-talkie effect (unlimited range).
func (c *ChatComponent) ActivateWalkieTalkie() {
	c.WalkieTalkieActive = true
	c.LocalRadius = -1.0 // Unlimited range marker
}

// DeactivateWalkieTalkie deactivates walkie-talkie effect.
func (c *ChatComponent) DeactivateWalkieTalkie() {
	c.WalkieTalkieActive = false
	if c.MegaphoneActive {
		c.LocalRadius = 30.0
	} else {
		c.LocalRadius = 10.0
	}
}

// GetEffectiveRadius returns the current effective local chat radius.
// Returns -1.0 for unlimited range (walkie-talkie).
func (c *ChatComponent) GetEffectiveRadius() float64 {
	return c.LocalRadius
}

// TradeProposal represents an active trade
type TradeProposal struct {
	ProposerID     uint64
	RecipientID    uint64
	OfferedItems   []string // Item IDs (strings, not uint64)
	RequestedItems []string // Item IDs (strings, not uint64)
	Status         string   // "pending", "accepted", "rejected", "committed", "cancelled", "failed"
	ProposalTime   int64    // Unix timestamp when proposed
	FailureReason  string   // Reason for failure/cancellation
}

// TradeRecord represents a completed trade
type TradeRecord struct {
	Timestamp int64
	PartnerID uint64
	Success   bool
}

// TradeComponent tracks trading state
type TradeComponent struct {
	ActiveTrade  *TradeProposal
	TradeHistory []TradeRecord
	TrustScore   float64 // 0.0-1.0, affects trade limits
}

// Type returns the component type
func (t *TradeComponent) Type() string {
	return "trade"
}
