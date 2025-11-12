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
