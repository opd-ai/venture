package engine

// ChatChannel represents different chat channels
type ChatChannel int

const (
	ChatGlobal ChatChannel = iota
	ChatLocal
	ChatParty
	ChatWhisper
)

// ChatMessage represents a single chat message
type ChatMessage struct {
	ID        string
	SenderID  uint64
	Channel   ChatChannel
	Content   string
	Timestamp int64
	Encrypted bool
}

// ChatComponent tracks chat state
type ChatComponent struct {
	Messages       []ChatMessage
	UnreadCount    int
	ActiveChannels []ChatChannel
}

// Type returns the component type
func (c *ChatComponent) Type() string {
	return "chat"
}

// TradeProposal represents an active trade
type TradeProposal struct {
	ProposerID     uint64
	RecipientID    uint64
	OfferedItems   []uint64
	RequestedItems []uint64
	Status         string // "pending", "accepted", "rejected"
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
