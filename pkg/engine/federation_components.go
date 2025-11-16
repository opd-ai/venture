package engine

// PortalComponent represents a cross-server portal
type PortalComponent struct {
	DestinationServer string // Server ID or "local" for same-server
	DestinationX      float64
	DestinationY      float64
	RequiredItem      string // Optional key item
	TrustRequired     string // Trust level required
}

// Type returns the component type
func (p PortalComponent) Type() string {
	return "portal"
}

// MailMessage represents a mail message
type MailMessage struct {
	ID          string
	SenderID    string
	RecipientID string
	Subject     string
	Body        string
	Attachments []uint64 // Item IDs
	Postage     int
	SentAt      int64
	DeliveredAt int64
}

// MailStatus represents the delivery status of a mail message
type MailStatus int

const (
	MailStatusSent MailStatus = iota
	MailStatusInTransit
	MailStatusDelivered
	MailStatusFailed
)

// String returns the human-readable status name
func (s MailStatus) String() string {
	switch s {
	case MailStatusSent:
		return "Sent"
	case MailStatusInTransit:
		return "In Transit"
	case MailStatusDelivered:
		return "Delivered"
	case MailStatusFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// GetStatus returns the current delivery status of the mail message
func (m *MailMessage) GetStatus() MailStatus {
	if m.DeliveredAt > 0 {
		return MailStatusDelivered
	}
	if m.SentAt > 0 {
		return MailStatusInTransit
	}
	return MailStatusSent
}

// MailComponent tracks player mail
type MailComponent struct {
	Inbox    []*MailMessage
	Outbox   []*MailMessage
	MaxInbox int
}

// Type returns the component type
func (m MailComponent) Type() string {
	return "mail"
}

// NewMailComponent creates a new mail component with default settings
func NewMailComponent() *MailComponent {
	return &MailComponent{
		Inbox:    make([]*MailMessage, 0),
		Outbox:   make([]*MailMessage, 0),
		MaxInbox: 50,
	}
}

// AddToInbox adds a message to the inbox if there's space
func (m *MailComponent) AddToInbox(msg *MailMessage) bool {
	if len(m.Inbox) >= m.MaxInbox {
		return false
	}
	m.Inbox = append(m.Inbox, msg)
	return true
}

// AddToOutbox adds a message to the outbox
func (m *MailComponent) AddToOutbox(msg *MailMessage) {
	m.Outbox = append(m.Outbox, msg)
}

// RemoveFromInbox removes a message from the inbox by ID
func (m *MailComponent) RemoveFromInbox(messageID string) bool {
	for i, msg := range m.Inbox {
		if msg.ID == messageID {
			m.Inbox = append(m.Inbox[:i], m.Inbox[i+1:]...)
			return true
		}
	}
	return false
}

// RemoveFromOutbox removes a message from the outbox by ID
func (m *MailComponent) RemoveFromOutbox(messageID string) bool {
	for i, msg := range m.Outbox {
		if msg.ID == messageID {
			m.Outbox = append(m.Outbox[:i], m.Outbox[i+1:]...)
			return true
		}
	}
	return false
}

// GetUnreadCount returns the number of unread messages in the inbox (delivered in last 24h)
func (m *MailComponent) GetUnreadCount() int {
	count := 0
	now := int64(0)
	if len(m.Inbox) > 0 {
		now = m.Inbox[0].DeliveredAt
		if now == 0 {
			return 0
		}
	}
	oneDayAgo := now - 86400
	for _, msg := range m.Inbox {
		if msg.DeliveredAt > 0 && msg.DeliveredAt > oneDayAgo {
			count++
		}
	}
	return count
}

// PostOfficeComponent marks an entity as a post office building
type PostOfficeComponent struct {
	ClerkName   string
	ServiceFee  int
	MaxDistance int
}

// Type returns the component type
func (p PostOfficeComponent) Type() string {
	return "postoffice"
}

// NewPostOfficeComponent creates a new post office component
func NewPostOfficeComponent(clerkName string) *PostOfficeComponent {
	return &PostOfficeComponent{
		ClerkName:   clerkName,
		ServiceFee:  10,
		MaxDistance: 100,
	}
}

// ServerFaction represents server-wide faction
type ServerFaction struct {
	ServerID     string
	FactionName  string
	Alignment    Alignment
	AllyServers  []string
	EnemyServers []string
	Reputation   map[string]float64
}

// PoliticalEvent represents faction events
type PoliticalEvent struct {
	Type         string // "alliance", "war", "treaty", "embargo"
	PartyServers []string
	StartTime    int64
	Duration     int64
	Effects      map[string]interface{}
}

// PoliticsComponent tracks political state
type PoliticsComponent struct {
	Faction *ServerFaction
	Events  []PoliticalEvent
}

// Type returns the component type
func (p PoliticsComponent) Type() string {
	return "politics"
}

// TerritoryComponent tracks territory control
type TerritoryComponent struct {
	ZoneID             string
	ControllingFaction string
	CaptureProgress    float64
}

// Type returns the component type
func (t TerritoryComponent) Type() string {
	return "territory"
}
